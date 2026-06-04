# SPEC Patch: D-P1.3 — 批量审核（补充 audit trail + 性能指标）

## 基本信息
- **任务 ID**: D-P1.3-v2
- **类型**: spec-patch
- **优先级**: P1
- **依赖**: D-P1.3（原有逻辑）
- **补丁目标**: `docs/plans/D-P1.3-batch-approve-spec.md`

---

## 变更 1：批量过账跳过逻辑（补充）

### 背景
新文档 §3.5.3："批量过账时，若某凭证关联的客户仍为草稿，系统自动跳过并提示"。

由于我们不做客户草稿，此条改为：**若某凭证关联的 ArInvoice 仍为 draft 状态，系统自动跳过并记录跳过原因**。

### 修正后的批量过账语义
```go
type BatchConfirmResult struct {
    SuccessCount int              `json:"success_count"`
    FailedCount  int              `json:"failed_count"`
    SkippedCount int              `json:"skipped_count"`         // 新增：跳过数量
    FailedList   []FailedItem     `json:"failed_list,omitempty"`
    SkippedList  []SkippedItem    `json:"skipped_list,omitempty"` // 新增：跳过列表
}

type SkippedItem struct {
    VoucherID   string `json:"voucher_id"`
    VoucherNo   string `json:"voucher_no"`
    Reason      string `json:"reason"` // "ar_invoice_draft" | "voucher_already_posted" | "permission_denied"
}
```

### 跳过逻辑
| 场景 | 跳过原因 | 返回码 |
|------|---------|--------|
| ArInvoice `status = 'draft'` | `ar_invoice_draft` | 不计入 success/failed，单独计入 skipped |
| 凭证 `docstatus != 0`（已过账） | `voucher_already_posted` | 同上 |
| 当前用户无过账权限 | `permission_denied` | 同上 |

### 验收标准（补充）
- [ ] 批量过账时，被跳过的凭证计入 `skipped_count`，不计入 `failed_count`
- [ ] `skipped_list` 包含每条跳过的 `voucher_id` 和 `reason`
- [ ] 正常过账的凭证计入 `success_count`

---

## 变更 2：过账后状态变更规则（显式文档化）

### 背景
新文档 §3.5.4 明确：过账后凭证转为"已记账"状态，不可修改。

当前 `docstatus` 枚举：
- `0` = 草稿（draft）
- `1` = 已记账（posted）

### 过账后状态变更
```
过账前：journal_entries.docstatus = 0
过账后：journal_entries.docstatus = 1， approved_by = $userID， approved_at = NOW()
```

**不可修改约束**（业务规则）：
- `docstatus = 1` 时，任何 `UPDATE` 操作均返回 400 错误
- 若需更正，必须通过**红字冲销**（新建 `docstatus=1` 的红字凭证，金额取负）

### 红字冲销规则（新文档 §4）
```
已过账凭证若需更正：
  1. 新建凭证，金额取负（原分录 × -1）
  2. 凭证摘要注明"红冲原凭证 PZ-XXXXXX"
  3. 原凭证状态不变，标记为"已被红冲"（新增 is_reversed = true）
  4. 两张凭证共同影响余额 = 0
```

### 影响文件

| 文件 | 变更 |
|------|------|
| `internal/service/voucher_service.go` | 过账方法增加 `docstatus=1` + audit 字段写入 |
| `internal/model/journal.go` | 新增 `IsReversed bool` 字段（可选，本次不实现） |
| `internal/repository/journal_repo.go` | 过账 UPDATE 增加条件 `WHERE docstatus = 0`，防止重复过账 |

### 验收标准（补充）
- [ ] 过账后凭证 `docstatus = 1`，`approved_by` / `approved_at` 有值
- [ ] 对 `docstatus = 1` 的凭证再次调用过账接口，返回 400（防止重复过账）
- [ ] 批量过账时，已过账的凭证自动进入 `skipped_list`

---

## 变更 3：性能指标（量化）

### 背景
新文档 §5 非功能需求：批量过账 100 张凭证草稿的响应时间不超过 30 秒。

### 性能要求
| 指标 | 要求 |
|------|------|
| 单张凭证过账 | ≤ 200ms |
| 100 张批量过账 | ≤ 30s |
| 批量过账事务 | 每张凭证独立事务（不要长事务） |

### 性能设计要点
- **独立事务**：每张凭证的 `UPDATE docstatus` + `INSERT audit` 为独立小事务，避免长事务锁
- **批量限制**：单次批量请求上限 100 张（超出拆分为多次）
- **无锁设计**：凭证按 ID 分片独立处理，避免全局锁串行化

### 验收标准（补充）
- [ ] 性能测试：批量过账 100 张 `docstatus=0` 凭证 ≤ 30 秒
- [ ] 单张凭证过账响应时间 ≤ 200ms
- [ ] 超过 100 张时返回 400 并提示"单次批量上限 100 张"

---

## 影响文件汇总

| 文件 | 变更 |
|------|------|
| `docs/plans/D-P1.3-batch-approve-spec.md` | 本补丁合并到原 SPEC |
| `internal/model/invoice.go` | `BatchConfirmResult` 新增 SkippedCount / SkippedList |
| `internal/service/voucher_service.go` | 过账增加 audit 字段写入 + 重复过账拦截 + 独立事务 |
| `internal/repository/journal_repo.go` | 过账 UPDATE 增加 `WHERE docstatus = 0` 乐观锁条件 |
| `migrations/047_audit_fields.sql` | `approved_by` / `approved_at`（与 D-P0.5 共用同一 migration） |