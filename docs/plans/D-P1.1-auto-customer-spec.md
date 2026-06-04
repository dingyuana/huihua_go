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

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 税号不存在时，导入流程继续（不报错）
- [ ] 自动创建的客户，`source='auto_import'`，`code` 以 `AUTO` 开头

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