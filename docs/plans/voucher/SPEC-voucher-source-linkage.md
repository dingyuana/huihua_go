# SPEC: 业财一体化 — 凭证与业务单据双向联动

## 背景

慧智财务 Go 版已实现银行流水分类（第一类/第二类）、PaymentEntry 核销引擎、凭证状态机的基本功能。但凭证与业务单据之间的联动存在断路：

1. **PaymentEntry.Approve 后不生成资金记账凭证**（`GenerateFromPaymentEntry` 已存在但未被调用）
2. **凭证过账后不锁定上游业务单据**（上游可被随意修改，账务数据被篡改风险）
3. **凭证草稿作废后不回退上游状态**（上游无法重新下推凭证）
4. **ArInvoice.voucher_id 字段缺失**（无法追溯"某应收单生了哪张凭证"）
5. **凭证 docstatus 变更后不更新 bank_txn matched 标记**

## 设计目标

在凭证状态变更（posted / cancelled / reversed）时，**主动**影响上游业务单据状态，实现"业务单据生成凭证、凭证过账锁定业务单据、凭证作废释放业务单据"的完整闭环。

---

## 一、凭证状态变更联动机制

### 1.1 联动触发时机

凭证状态机每种状态变更，都可能触发对上游业务单据的联动：

| 凭证状态变更 | 触发联动 | 影响 |
|-------------|---------|------|
| 0→1（submit/posted，过账） | `LockSourceDoc` | 上游 ArInvoice/PaymentEntry 被锁定（禁止修改/作废） |
| 1→3（cancel，作废） | `UnlockSourceDoc` | 上游业务单据解除锁定，可重新下推 |
| 1→4（reverse，红字冲销） | `LockSourceDoc`（新凭证） | 新凭证锁定上游，新+原凭证均标记 reversed |

### 1.2 LockSourceDoc 实现

凭证 posted 时，锁定上游业务单据：

```go
func (s *VoucherService) LockSourceDoc(ctx context.Context, voucher *model.JournalEntry) error {
    if voucher.SourceType == "" || voucher.SourceID == uuid.Nil {
        return nil // 无上游，忽略
    }
    switch voucher.SourceType {
    case "ar_invoice":
        return s.arInvoiceRepo.Lock(ctx, voucher.SourceID)
    case "payment_entry":
        return s.paymentRepo.Lock(ctx, voucher.SourceID)
    case "reimbursement":
        return s.reimbursementRepo.Lock(ctx, voucher.SourceID)
    case "bank_txn":
        return s.bankTxnRepo.SetMatchedJournalEntry(ctx, voucher.SourceID, voucher.VoucherID)
    }
    return nil
}
```

### 1.3 UnlockSourceDoc 实现

凭证 cancelled 时，解除上游业务单据锁定：

```go
func (s *VoucherService) UnlockSourceDoc(ctx context.Context, voucher *model.JournalEntry) error {
    if voucher.SourceType == "" || voucher.SourceID == uuid.Nil {
        return nil
    }
    switch voucher.SourceType {
    case "ar_invoice":
        return s.arInvoiceRepo.Unlock(ctx, voucher.SourceID)
    case "payment_entry":
        return s.paymentRepo.Unlock(ctx, voucher.SourceID)
    case "reimbursement":
        return s.reimbursementRepo.Unlock(ctx, voucher.SourceID)
    case "bank_txn":
        return s.bankTxnRepo.ClearMatchedJournalEntry(ctx, voucher.SourceID)
    }
    return nil
}
```

### 1.4 VoucherService 结构变更

新增 `VoucherService` 或在现有 `VoucherStateMachineService` 中增加联动方法：

```
新增依赖注入：
- arInvoiceRepo ArInvoiceRepository
- paymentRepo PaymentEntryRepository
- reimbursementRepo ReimbursementRepository
- bankTxnRepo BankTransactionRepository

新增方法：
- LockSourceDoc(ctx, voucher) error
- UnlockSourceDoc(ctx, voucher) error
- UpdateSourceDocVoucherID(ctx, voucher) error  // posted 时回写 voucher_id 到上游

现有方法变更：
- Submit(ctx, voucherID) → 调用 LockSourceDoc
- Cancel(ctx, voucherID) → 调用 UnlockSourceDoc
- Reverse(ctx, voucherID) → 调用 LockSourceDoc（新凭证）
```

---

## 二、PaymentEntry.Approve → 生成资金记账凭证

### 2.1 现状

`PaymentEntryService.Approve()` 目前只做两件事：
1. 更新 DocStatus 0→2
2. 调用 `ReconcilePaymentEntry`（发票核销）

**缺失**：核准后不生成资金记账凭证。

### 2.2 改动

`GenerateFromPaymentEntry`（voucher_auto_generate_service.go:399）已存在，直接在 `Approve` 中调用：

```go
func (s *PaymentEntryService) Approve(ctx context.Context, tenantID, paymentID, userID uuid.UUID) (*ApproveResult, error) {
    pe, err := s.repo.GetByID(ctx, tenantID, paymentID)
    if err != nil {
        return nil, fmt.Errorf("load payment entry: %w", err)
    }
    if pe.DocStatus != 1 {
        return nil, fmt.Errorf("payment entry must be in submitted status (docstatus=1)")
    }

    // 1. 生成资金记账凭证（草稿，docstatus=0）
    voucher, err := s.voucherGenSvc.GenerateFromPaymentEntry(ctx, tenantID, paymentID, userID)
    if err != nil {
        return nil, fmt.Errorf("generate voucher from payment entry: %w", err)
    }

    // 2. 更新 DocStatus 为 2（已核准）
    if err := s.repo.UpdateStatus(ctx, tenantID, paymentID, 2); err != nil {
        return nil, fmt.Errorf("update docstatus: %w", err)
    }

    // 3. 同步 bank_txn matched 标记
    if pe.BankTxnID != nil {
        s.bankTxnRepo.SetMatchedPaymentEntry(ctx, *pe.BankTxnID, paymentID)
    }

    // 4. 调用核销引擎（发票核销）
    pair, err := s.reconSvc.ReconcilePaymentEntry(ctx, tenantID, paymentID)
    if err != nil {
        // 核销失败不阻断核准
        pair = nil
    }

    return &ApproveResult{Voucher: voucher, ReconciliationPair: pair}, nil
}

type ApproveResult struct {
    Voucher            *model.JournalEntry
    ReconciliationPair *model.ReconciliationPair
}
```

### 2.3 internal_transfer 特殊处理

`GenerateFromPaymentEntry` 中已区分普通类型 vs `internal_transfer`：
- **普通类型**：借：银行存款 / 贷：应收账款（核销 ArInvoice）
- **internal_transfer**：借：银行存款A / 贷：银行存款B（直接制证，无 ArInvoice 核销）

此逻辑已在 `GenerateFromPaymentEntry` 中实现，无需额外修改。

---

## 三、ArInvoice.voucher_id 与双向绑定

### 3.1 模型变更

`ar_invoice` 表新增字段：

```sql
ALTER TABLE ar_invoice ADD COLUMN voucher_id uuid REFERENCES journal_entries(id);
ALTER TABLE ar_invoice ADD COLUMN locked_at timestamptz;  -- 被凭证锁定的时间
ALTER TABLE ar_invoice ADD COLUMN locked_by uuid REFERENCES users(id);
```

### 3.2 ArInvoiceRepository 新增方法

```go
type ArInvoiceRepository interface {
    // 现有方法...

    // 新增
    Lock(ctx context.Context, id uuid.UUID) error           // 凭证过账时调用
    Unlock(ctx context.Context, id uuid.UUID) error          // 凭证作废时调用
    SetVoucherID(ctx context.Context, id, voucherID uuid.UUID) error  // 回写 voucher_id
}
```

### 3.3 凭证 posted 时回写 ArInvoice.voucher_id

凭证状态变为 posted 时（`LockSourceDoc` 同一时机），回写：

```go
func (s *VoucherService) UpdateSourceDocVoucherID(ctx context.Context, voucher *model.JournalEntry) error {
    if voucher.SourceType != "ar_invoice" || voucher.SourceID == uuid.Nil {
        return nil
    }
    return s.arInvoiceRepo.SetVoucherID(ctx, voucher.SourceID, voucher.VoucherID)
}
```

---

## 四、凭证状态机现有方法变更

### 4.1 VoucherStateMachineService

现有 `VoucherStateMachineService` 是独立的状态机服务，需要增加联动：

```go
type VoucherStateMachineService struct {
    voucherRepo       JournalEntryRepository
    arInvoiceRepo     ArInvoiceRepository      // 新增
    paymentRepo      PaymentEntryRepository   // 新增
    reimbursementRepo ReimbursementRepository  // 新增
    bankTxnRepo      BankTransactionRepository // 新增
}

func (s *VoucherStateMachineService) ExecuteTransition(ctx context.Context, req ExecuteTransitionRequest) (*model.JournalEntry, error) {
    // ... 原有状态流转逻辑 ...

    // 状态变更后触发联动
    if newStatus == model.StatusPosted {
        // 锁定上游 + 回写 voucher_id
        go func() {
            s.lockSourceDoc(context.Background(), voucher)
            s.updateSourceDocVoucherID(context.Background(), voucher)
        }()
    } else if newStatus == model.StatusCancelled {
        go func() {
            s.unlockSourceDoc(context.Background(), voucher)
        }()
    }

    return voucher, nil
}
```

### 4.2 事务边界说明

联动操作（Lock/Unlock/SetVoucherID）在**异步 goroutine** 中执行，不影响凭证状态变更的主体事务。这是**最终一致性**而非强一致性：

- 凭证 posted 是主体事务（必须成功）
- 上游业务单据锁定是配套操作（可异步，最大保证最终一致）

> 若业务对强一致性有要求，可将联动操作纳入同一事务（在同一 db transaction 中执行）。当前设计优先考虑性能和解耦。

---

## 五、验证标准

### 功能验证

1. PaymentEntry.Approve → 生成凭证草稿（`source_type='payment_entry'`，`source_id=payment_id`）
2. 凭证 Submit（posted）→ ArInvoice.voucher_id 被回写，ArInvoice 被锁定（修改接口返回错误）
3. 凭证 Cancel → ArInvoice.voucher_id 清空，ArInvoice 解除锁定
4. 凭证 Reverse → 新凭证锁定上游，原凭证不变
5. bank_txn matched 标记在凭证 posted 时正确更新

### 字段追溯链验证

```
凭证 V1 --source_type/source_id--> PaymentEntry P1
    ↑                                              |
    |--- ArInvoice A1（通过 payment_allocations）  |
                                                     |
凭证 V2（资金凭证）--source_type/source_id------→ PaymentEntry P1
```

---

## 六、涉及的文件变更

| 文件 | 变更类型 | 内容 |
|------|---------|------|
| `internal/model/ar_invoice.go` | 修改 | 新增 voucher_id / locked_at / locked_by 字段 |
| `internal/repository/ar_invoice_repo.go` | 修改 | 新增 Lock/Unlock/SetVoucherID 方法 |
| `internal/repository/payment_entry_repo.go` | 修改 | 新增 Lock/Unlock 方法 |
| `internal/repository/bank_txn_repo.go` | 修改 | SetMatchedJournalEntry/ClearMatchedJournalEntry |
| `internal/service/voucher_state_machine_service.go` | 修改 | 注入新依赖，ExecuteTransition 增加联动 |
| `internal/service/payment_service.go` | 修改 | Approve 调用 GenerateFromPaymentEntry |
| `db/migrations/` | 新增 | ar_invoice 表字段变更 Migration |

---

## 七、优先级

| 优先级 | 内容 | 理由 |
|--------|------|------|
| **P0** | PaymentEntry.Approve → 生成资金记账凭证 | 第二类链路核心缺口，业务无法闭环 |
| **P0** | PaymentEntry.Approve → 同步 bank_txn matched | 银行流水状态联动 |
| **P1** | VoucherStateMachine → LockSourceDoc / UnlockSourceDoc | 凭证锁定上游 |
| **P1** | ArInvoice.voucher_id 回写 | 双向追溯 |
| **P2** | ArInvoice 模型完整 Service 层 | 以独立 ArInvoice 审核功能为基础 |
