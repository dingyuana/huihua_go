# SPEC: D-P1.1 — 导入时自动创建正式客户

## 基本信息
- **任务 ID**: D-P1.1
- **类型**: feature
- **优先级**: P1
- **依赖**: 无（独立于 P0 链路）
- **负责 Profile**: dev

## 背景
当前 `resolveCustomer` 在税号查不到时返回 error，中断导入流程。新需求：查不到时直接创建正式客户档案（不做草稿）。

## 目标
修改 `internal/service/invoice_service.go` 的 `resolveCustomer` 相关逻辑：

```
resolveCustomer(taxID, customerName):
  1. 精确查 Party WHERE tax_id = $taxID → 找到 → return
  2. 未找到 → 自动创建 Party：
     - name = customerName
     - tax_id = taxID
     - party_type = "customer"
     - code = "AUTO" + 年月日流水号（如 AUTO20260604001）
     - source = "auto_import"  ← 新增字段
  3. return newCustomer
```

**Party 新增字段**（migration 046）：
```sql
ALTER TABLE parties ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'manual';
```

---

## 变更 1：模糊匹配分支（新增）

### 背景
税号查不到时，如果"购方名称"与现有客户高度相似，系统应弹出提示让用户选择"关联已有客户"或"仍自动创建"。

### 补充逻辑（新增 Step 1.5）
```
Step 1: 精确查 Party WHERE tax_id = $taxID
  → 找到 → return party

Step 1.5（新增）: 税号不存在时，检查名称相似度
  → 从 Party 表查出所有 customer 类型客户的名称
  → 计算与 $customerName 的相似度（如编辑距离 / Jaro-Winkler）
  → 相似度 ≥ 阈值（如 0.85）且唯一命中：
      → 记录为"模糊匹配候选"，在 Preview 响应中标记
      → 用户可在预览界面选择"关联已有"或"仍自动创建"
  → 无相似匹配或多个模糊匹配 → 自动创建（行为同原 Step 2）

Step 2: 未找到任何匹配 → 自动创建 Party：
  - name = customerName
  - tax_id = taxID
  - party_type = "customer"
  - code = "AUTO" + 年月日流水号（如 AUTO20260604001）
  - source = "auto_import"
```

### 数据结构变更
**`BatchImportPreviewResult`** 新增字段：
```go
type CustomerMatchInfo struct {
    TaxID           string  `json:"tax_id"`
    CustomerName    string  `json:"customer_name"`
    MatchType      string  `json:"match_type"`   // "exact" | "fuzzy" | "auto_create"
    MatchedPartyID *string `json:"matched_party_id,omitempty"` // exact/fuzzy 时有值
    FuzzyScore     float64 `json:"fuzzy_score,omitempty"`       // fuzzy 时有值
    FuzzyCandidate *string `json:"fuzzy_candidate,omitempty"`   // fuzzy 模糊匹配到的高分客户名
}

type BatchImportPreviewResult struct {
    // ... 现有字段 ...
    CustomerMatches []CustomerMatchInfo `json:"customer_matches,omitempty"` // 每行购方的匹配情况
}
```

**用户选择传递**（Confirm 阶段新增参数）：
```go
type BatchImportConfirmRequest struct {
    InvoiceIDs        []string `json:"invoice_ids"`
    // 新增：用户对模糊匹配的选择
    CustomerMappings map[string]CustomerMappingOption `json:"customer_mappings,omitempty"`
}

type CustomerMappingOption struct {
    Action     string  `json:"action"` // "use_existing" | "create_new"
    PartyID    string  `json:"party_id,omitempty"` // use_existing 时传此 ID
}
```

### 验收标准（新增）
- [ ] 税号精确查到 → `match_type = "exact"`
- [ ] 税号不存在但名称相似度 ≥ 0.85 → `match_type = "fuzzy"`，返回相似客户名和得分
- [ ] 用户在预览阶段可指定"关联已有"或"仍创建"
- [ ] 不指定时默认自动创建

---

## 变更 2：并发去重保护（修正 Step 2 语义）

### 背景
原 SPEC 说"DB 报错后重查"——但 `ON CONFLICT DO NOTHING` 会静默跳过，导致同一税号发票关联到不同的自动创建客户（数据不一致）。

### 修正为 Upsert 语义
```sql
-- 改用 ON CONFLICT DO UPDATE（唯一索引在 tax_number 上已存在）
INSERT INTO parties (id, tenant_id, party_type, name, tax_number, source, code, ...)
VALUES ($1, $2, $3, $4, $5, $6, $7, ...)
ON CONFLICT (tax_number) WHERE tenant_id = $2
DO UPDATE SET
    name = EXCLUDED.name,
    source = CASE WHEN parties.source = 'manual' THEN 'auto_import' ELSE parties.source END
RETURNING id
```

**关键点**：
- `ON CONFLICT DO UPDATE` 确保并发时以第一个成功写入的为准
- `tax_number` 唯一索引已在 `010_account_setup.sql` 中建立（检查确认）
- 不允许从 `auto_import` 覆盖 `manual`（防止误将手动客户改为系统自建）
- 返回 `RETURNING id` 让调用方获取实际写入的客户 ID

### 验收标准（补充）
- [ ] 并发导入两条同税号发票，只创建一个客户（不重复）
- [ ] 先写成功的客户 ID 被两条发票共用

---

## 变更 3：客户编码流水号需 DB 层保护

### 问题
`code = "AUTO" + 年月日流水号` 的序号用代码计算（`count + 1`），并发时会产生重复 code。

### 解决方案
改由 DB `UNIQUE` 约束 + `ON CONFLICT DO UPDATE` 兜底，或使用 `SELECT FOR UPDATE` 锁。

---

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 税号不存在时，导入流程继续（不报错）
- [ ] 自动创建的客户，`source='auto_import'`，`code` 以 `AUTO` 开头
- [ ] 模糊匹配功能正常（相似度 ≥ 0.85 时标记）
- [ ] 并发导入同税号发票，只创建一个客户

## 技术约束
- 客户编码规则：`AUTO` + `YYYYMMDD` + 4位序号（如 `AUTO202606040001`）
- 序号在当天的客户数基础上 +1（需要 count 查询）
- 并发保护：`tax_id` 有 unique index，若并发创建同税号客户，DB 报错后重查即可

## OpenCode 指令模板
**目标**：实现导入时自动创建客户

**约束**：
- 修改 `resolveCustomer` 方法（找同名方法确认当前实现）
- 新增 migration `046_parties_source.sql`

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- 参照：`party_repo.go` 的 `Create` / `BatchCreate` 方法

**验收**：
- `go build ./...` 无报错
- 测试：tax_id 不存在时，自动创建客户并继续

---

## 影响文件汇总

| 文件 | 变更 |
|------|------|
| `docs/plans/D-P1.1-auto-customer-spec.md` | 本补丁合并到原 SPEC |
| `internal/service/invoice_service.go` | `resolveCustomer` 新增 Step 1.5 模糊匹配 |
| `internal/model/invoice.go` | `CustomerMatchInfo` 结构体 |
| `internal/repository/party_repo.go` | Create 改为 upsert 语义 |