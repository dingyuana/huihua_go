# SPEC: D-P0.3 — 重写 ConfirmSalesInvoice：生成 ArInvoice + 触发凭证生成

## 基本信息
- **任务 ID**: D-P0.3
- **类型**: feature
- **优先级**: P0
- **依赖**: D-P0.2
- **负责 Profile**: dev

## 背景
当前 `ConfirmSalesInvoice` 仅改发票 status="verified"，不生成 ArInvoice 和凭证。三级草稿链路的第一环是：发票确认 → 生成 ArInvoice 草稿 + 凭证草稿。

## 目标
重写 `internal/service/invoice_service.go` 中的 `ConfirmSalesInvoice` 方法，实现：

```
ConfirmSalesInvoice(ctx, tenantID, invoiceID, userID uuid.UUID) error

逻辑：
  1. GetByID 查发票，验证 status in [draft, submitted, verified]
  2. 调用 arInvoiceRepo.ListByInvoiceID 查是否已有 ArInvoice → 有则报错"已存在"
  3. 生成 ArInvoice draft（amount=发票.totalAmount，customerID/invoiceNo 从发票取）
  4. 调用 s.voucherAutoSvc.GenerateFromInvoice(tenantID, invoiceID, userID) 生成凭证草稿
  5. arInvoiceRepo.Create 写入 ArInvoice draft
  6. repo.UpdateStatus 发票 status → verified
  7. 返回 nil
```

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 发票确认后：ArInvoice 记录存在（status=draft）、凭证记录存在（status=draft）
- [ ] 防重复：同一发票调用两次 `ConfirmSalesInvoice`，第二次返回错误

## 技术约束
- 不修改 `InvoiceService` 的构造签名（保持兼容）
- `voucherAutoSvc` 在 `InvoiceService` 中已存在（检查 `s.voucherAutoSvc` 是否为 nil，若 nil 则跳过凭证生成并记录 log）
- 发票 `status` 检查：仅 `draft`/`submitted`/`verified` 可确认，其他报错
- ArInvoice.Create 在凭证生成**之后**执行（确保凭证先生成，草稿状态正确）

## OpenCode 指令模板
**目标**：重写 `ConfirmSalesInvoice` 方法，生成 ArInvoice 草稿 + 触发凭证生成

**约束**：
- 只修改 `internal/service/invoice_service.go` 中的 `ConfirmSalesInvoice` 方法
- 不改其他方法的逻辑
- `s.voucherAutoSvc.GenerateFromInvoice` 调用需传正确参数（参考 `voucher_auto_generate_service.go` 的签名）

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- arInvoiceRepo：`internal/repository/ar_invoice_repo.go`
- voucherAutoSvc：`VoucherAutoGenerateService`（已在 InvoiceService 中通过依赖注入获得）

**验收**：
- `go build ./...` 无报错
- 单元测试或集成测试验证：确认发票后 ArInvoice 存在 + 凭证存在