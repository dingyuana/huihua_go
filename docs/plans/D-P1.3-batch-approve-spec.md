# SPEC: D-P1.3 — 批量审核接口

## 基本信息
- **任务 ID**: D-P1.3
- **类型**: feature
- **优先级**: P1
- **依赖**: D-P0.5
- **负责 Profile**: dev

## 背景
财务人员需要批量过账凭证，不能逐张操作。批量接口需要处理"跳过"逻辑（上游草稿未审核时跳过并提示）。

## 目标
在 `internal/handler/voucher_handler.go` 和 `internal/handler/invoice_handler.go` 新增/完善批量接口：

### 接口1：批量确认发票
```
POST /api/v1/invoices/batch-confirm
Body: { "invoice_ids": ["uuid1", "uuid2", ...] }
```

### 接口2：批量过账凭证（已有部分 stub，需完善）
```
POST /api/v1/vouchers/batch-approve
Body: { "voucher_ids": ["uuid1", "uuid2", ...] }
```

**批量过账逻辑**：
```
for each voucherID:
  1. 查询凭证，检查 docstatus=0（草稿）
  2. 查询关联的 ArInvoice（通过 source_id）：
     - ArInvoice status != confirmed → 跳过（记录 skip_reason="ar_invoice_not_confirmed"）
  3. docstatus → 1（posted），记录 posted_at/by
return { success_count, skipped_count, skipped_list[{id, reason}] }
```

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 批量确认发票：已确认的发票不再重复生成 ArInvoice
- [ ] 批量过账凭证：上游 ArInvoice 未确认时跳过，不报错
- [ ] 返回结构包含 `success_count`、`skipped_count`、`skipped_list`

## 技术约束
- 批量接口不做事务（各单据独立处理，部分成功部分失败均可接受）
- 凭证关联 ArInvoice：通过 `source_type='invoice'` AND `source_id` 查询

## OpenCode 指令模板
**目标**：实现批量审核接口

**约束**：
- 修改 `voucher_handler.go`（批量过账）
- 修改 `invoice_handler.go`（批量确认发票）
- 需要注入 `arInvoiceRepo` 到 VoucherHandler（若尚未注入，先检查 main.go）

**上下文**：
- 项目：`/root/data/disk/huihua-finance`

**验收**：
- `go build ./...` 无报错
- 批量过账时，上游草稿的凭证被正确跳过