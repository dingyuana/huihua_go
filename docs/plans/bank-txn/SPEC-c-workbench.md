# SPEC: C 类流水待处理工作台

## 背景

银行流水中无法自动分类的交易（`classifyType` 返回 "C"）进入 `manual_pending` 状态，等待出纳处理。

当前状态：
- AI 分类后，C 类流水状态 = `manual_pending` ✅
- `RejectManual` 可将第一类/第二类流水打回 C 类 ✅
- **缺少正向流程**：人看到 C 类流水后，如何分类+执行？ ❌

## 目标

提供 C 类流水工作台：
1. 列表 API：C 类状态（`manual_pending`）的流水，按日期/金额筛选
2. 处理 API：人对某条 C 类流水选择处理方式（**第一类**直接制证 / **第二类**生成付款单），系统执行对应操作

## 改动范围

### 1. BankTxnReviewService 新增 `ProcessManual` 方法

```go
func (s *BankTxnReviewService) ProcessManual(
    ctx context.Context,
    tenantID uuid.UUID,
    txnID string,        // 流水ID
    action string,       // "第一类" 或 "第二类"，人选择的处理方式
    userID uuid.UUID,
) (*TxnResult, error)
```

逻辑：
1. 加载流水，验证状态 = `manual_pending`
2. 根据 action 执行：
   - "第一类"：调用 `s.voucherAutoSvc.GenerateFromBankTxn()` 生成凭证草稿（docstatus=0）
   - "第二类"：调用 `s.paymentSvc.CreateFromBankTransaction()` 生成付款单草稿（status=draft）
3. 更新流水状态：
   - "第一类" → `BankTxnReviewStatusVoucherGenerated`
   - "第二类" → `BankTxnReviewStatusPaymentCreated`
4. 返回 TxnResult（与人处理第一类/第二类结果格式一致）

### 2. BankTxnReviewHandler 新增两个端点

**列表 C 类流水：**
```
GET /api/v1/bank-transactions/manual-pending?start_date=&end_date=&bank_account_id=&page=&page_size=
```

响应：
```json
{
  "data": {
    "items": [...],  // BankTransaction 列表
    "total": N,
    "page": 1,
    "page_size": 50
  }
}
```

过滤条件：`status = 'manual_pending'`

**处理 C 类流水：**
```
POST /api/v1/bank-transactions/:id/process-manual
Body: { "action": "第一类" | "第二类", "payment_type": "pay" | "receive" }  // action 必填，payment_type 仅第二类时需要
```

响应：`TxnResult` JSON

### 3. 数据库索引（如不存在）

`bank_transactions` 表上 `idx_txn_status_tenant` 应支持 `WHERE status = 'manual_pending' AND tenant_id = ?` 查询。

## 核心原则

人是审核唯一主体。系统不猜测 C 类流水的分类，由人判断并指定处理方式（**第一类**直接制证 / **第二类**需中转付款单）。所有生成的单据均为草稿，人工审核后正式生效。

## 验证

1. `go build ./...` 通过
2. 手动测试：C 类流水 `POST /:id/process-manual` action="第一类" → 凭证生成
3. 手动测试：C 类流水 `POST /:id/process-manual` action="第二类" → 付款单生成