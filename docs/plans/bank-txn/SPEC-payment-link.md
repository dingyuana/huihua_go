# SPEC: PaymentEntry 审核后自动核销发票

## 背景

PaymentEntry 由银行流水的 B 类分支生成（`SubmitReview` case "B"），生成后状态为 docstatus=0（草稿）。

当前问题：
1. `ReconciliationService.Reconcile()` 只匹配 bank_txn ↔ invoice，不消费 PaymentEntry
2. PaymentEntry approve 后不触发核销，发票 OutstandingAmount 不更新
3. PaymentEntry 与 invoice 的核销链路断路

## 目标

在 PaymentEntry 审核（approve）通过后，自动调用核销引擎：
1. 根据 PaymentEntry 的 counterparty/金额匹配未核销发票
2. 创建 `payment_allocation` 记录
3. 更新发票 `outstanding_amount`

## 改动范围

### 1. `ReconciliationService` 新增 `ReconcilePaymentEntry` 方法

签名：
```go
func (s *ReconciliationService) ReconcilePaymentEntry(
    ctx context.Context,
    tenantID uuid.UUID,
    paymentEntryID uuid.UUID,
) (*model.ReconciliationPair, error)
```

逻辑：
1. 根据 PaymentEntryID 加载 PaymentEntry（含 counterparty/amount）
2. 查同租户、同公司的未核销发票列表（outstanding_amount > 0，status=verified）
3. 按金额 + counterparty 精确匹配（L1：reference_no 含 invoice ID；L2：invoice_no 在 PaymentEntry description 中；L3：counterparty + amount + date(±3天)）
4. 匹配成功后：
   - 创建 `payment_allocations` 记录（payment_entry_id, invoice_id, allocated_amount）
   - 更新发票 `outstanding_amount = outstanding_amount - allocated_amount`
   - 返回 ReconciliationPair (status="pending"，需人确认)
5. 若无匹配，返回 nil, nil（不报错，容许纯付款无发票场景）

### 2. PaymentEntry Service 新增 `Approve` 方法

```go
func (s *PaymentEntryService) Approve(ctx context.Context, tenantID, paymentID, userID uuid.UUID) error
```

逻辑：
1. 验证 DocStatus == 1（已提交），非法则报错
2. 更新 DocStatus = 2（已审核）
3. 调用 `s.reconSvc.ReconcilePaymentEntry(ctx, tenantID, paymentID)` 触发核销（失败不阻断 approve）
4. 若有核销结果，人工确认后再更新状态

### 3. `PaymentEntryHandler` 新增 `POST /payment-entries/:id/approve`

流程：
1. 人点击"审核通过"
2. 后端调用 `PaymentEntryService.Approve()`
3. 返回核销结果（含建议配对，人可调整）
4. 人确认后调用 `POST /reconciliation/pairs/:id/confirm`

### 4. 现有 `ReconciliationService.Reconcile()` 保持不变

全量银行流水核销（定时任务/手动触发）与 PaymentEntry 核销（实时触发）并行，互不干扰。

## 核心原则

人是审核唯一主体。系统只做建议和预核销，所有核销配对需人确认后才落 `payment_allocations`（status=pending → confirmed）。

## 验证

1. `go build ./...` 通过
2. 手动测试：approve 一笔收款单 → 检查发票 outstanding_amount 是否减少