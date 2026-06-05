# PLAN: 业财一体化凭证联动 — 开发执行计划

> 对应 SPEC: `SPEC-voucher-source-linkage.md`
> 优先级: P0/P1
> 目标: 实现凭证与业务单据（PaymentEntry / ArInvoice / bank_txn）的双向联动

---

## 执行顺序

按依赖关系排序，先完成 P0 再做 P1。

```
P0-1  PaymentEntry.Approve → 生成资金记账凭证
         ↓
P0-2  bank_txn.matched 同步（PaymentEntry.Approve 内）
         ↓
P1-1  VoucherStateMachine → LockSourceDoc / UnlockSourceDoc
         ↓
P1-2  ArInvoice 模型完善（voucher_id + Lock/Unlock + Repository）
         ↓
P1-3  ArInvoice.voucher_id 回写
```

---

## P0-1: PaymentEntry.Approve → 生成资金记账凭证

### 任务范围

修改 `internal/service/payment_service.go` 的 `Approve` 方法，在核准时调用 `GenerateFromPaymentEntry`。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/payment_service.go` | Approve 方法增加 GenerateFromPaymentEntry 调用；返回结构改为 ApproveResult |
| `internal/service/payment_service.go` | ApproveResult 类型定义 |
| `internal/handler/payment_entry_handler.go` | 适配新的返回结构（返回 voucher_id 给前端） |

### 验证条件

- [ ] 核准一笔付款单（第二类银行流水生成的 PaymentEntry）后，检查 `journal_entries` 表有新记录
- [ ] 新凭证 `source_type='payment_entry'`，`source_id` 等于该 PaymentEntry 的 ID
- [ ] 新凭证 docstatus=0（草稿）
- [ ] `go build ./...` 通过
- [ ] `curl` 或前端触发 approve，响应包含新生成的凭证 ID

### 风险点

- `GenerateFromPaymentEntry` 依赖 `voucherGenSvc`（VoucherAutoGenerateService）注入：需确认 `PaymentEntryService` 构造时已注入
- 若 PaymentEntry 无关联 ArInvoice，`GenerateFromPaymentEntry` 会走 hardcoded fallback 科目，验证时注意检查分录科目

---

## P0-2: bank_txn.matched 同步

### 任务范围

PaymentEntry.Approve 成功后，同步更新 `bank_txn.matched_payment_entry_id` 和 `bank_txn.status='matched'`。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/payment_service.go` | Approve 方法中，核准后调用 `s.bankTxnRepo.SetMatchedPaymentEntry` |
| `internal/repository/bank_txn_repo.go` | 确认 `SetMatchedPaymentEntry` 方法存在且实现正确 |

### 验证条件

- [ ] 核准 PaymentEntry 后，对应的 bank_txn 行 `matched_payment_entry_id` = 该 PaymentEntry ID
- [ ] bank_txn.status = 'matched'
- [ ] 已通过 commit `88acd9d3` 修复过一次，此处验证行为正确

---

## P1-1: VoucherStateMachine → LockSourceDoc / UnlockSourceDoc

### 任务范围

凭证状态变为 posted 时锁定上游，变为 cancelled 时解除。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/voucher_state_machine_service.go` | 新增依赖注入（arInvoiceRepo / paymentRepo / reimbursementRepo） |
| `internal/service/voucher_state_machine_service.go` | 新增 LockSourceDoc / UnlockSourceDoc / UpdateSourceDocVoucherID 方法 |
| `internal/service/voucher_state_machine_service.go` | ExecuteTransition 方法在状态变更后触发联动（异步 goroutine） |
| `internal/repository/ar_invoice_repo.go` | 新增 Lock(ctx, id) / Unlock(ctx, id) 方法 |
| `internal/repository/payment_entry_repo.go` | 新增 Lock(ctx, id) / Unlock(ctx, id) 方法 |
| `db/migrations/` | 新增 Migration（ar_invoice 表 + payment_entry 表增加 locked 字段） |

### 改动说明

```sql
-- ar_invoice 表
ALTER TABLE ar_invoice ADD COLUMN locked_at timestamptz;
ALTER TABLE ar_invoice ADD COLUMN locked_by uuid REFERENCES users(id);
ALTER TABLE ar_invoice ADD COLUMN voucher_id uuid REFERENCES journal_entries(id);

-- payment_entry 表
ALTER TABLE payment_entry ADD COLUMN locked_at timestamptz;
ALTER TABLE payment_entry ADD COLUMN locked_by uuid REFERENCES users(id);
```

### 验证条件

- [ ] 凭证 submit（posted）后，查询对应 ArInvoice/PaymentEntry，`locked_at` 有值
- [ ] 凭证 cancelled 后，`locked_at` 被清空
- [ ] 锁定状态下尝试修改 ArInvoice（PUT 接口）返回 400/403 错误
- [ ] `go build ./...` 通过

---

## P1-2: ArInvoice 模型完善

### 任务范围

补充 ArInvoice 的完整模型 + Repository + 基本 Service。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/model/ar_invoice.go` | 新增字段（voucher_id / locked_at / locked_by） |
| `internal/repository/ar_invoice_repo.go` | 新增 Lock / Unlock / SetVoucherID / IsLocked 方法 |
| `db/migrations/` | 新增 Migration |

### 验证条件

- [ ] `go build ./...` 通过
- [ ] Migration 可独立运行（`migrate up`）

---

## P1-3: ArInvoice.voucher_id 回写

### 任务范围

凭证 posted 时，将凭证 ID 回写到 ArInvoice.voucher_id 字段。

### 改动文件

| 文件 | 改动 |
|------|------|
| `internal/service/voucher_state_machine_service.go` | ExecuteTransition 中，posted 状态变更后调用 UpdateSourceDocVoucherID |

### 验证条件

- [ ] 触发一张应收记账凭证过账，查询对应 ArInvoice 行，`voucher_id` 等于该凭证 ID
- [ ] `go build ./...` 通过

---

## 依赖关系

```
P0-1 (Approve → 凭证)
    └── 依赖: GenerateFromPaymentEntry 已存在（无需修改 voucher_auto_generate_service.go）
    └── 风险: voucherGenSvc 是否已注入 PaymentEntryService（需确认构造函数）

P0-2 (bank_txn matched 同步)
    └── 依赖: P0-1 同一次 Approve 调用中完成
    └── 风险: SetMatchedPaymentEntry 实现正确性（已修复过，需验证）

P1-1 (LockSourceDoc)
    └── 依赖: P1-2（ArInvoice 有 Lock 方法）
    └── 风险: 异步 goroutine 中的错误无法返回给调用方（可接受，Eventually Consistent）

P1-2 (ArInvoice 模型完善)
    └── 依赖: 无
    └── 风险: Migration 需要在已有数据的环境下执行（加 IF NOT EXISTS / Default 值保护）

P1-3 (voucher_id 回写)
    └── 依赖: P1-1（P1-2 的 Lock 方法是同一个 Migration）
    └── 风险: 无
```

---

## 预计工作量

| 任务 | 估计 |
|------|------|
| P0-1 | 1-2 小时（主要改 Approve 方法 + Handler 适配） |
| P0-2 | 15 分钟（一行代码，但需验证） |
| P1-1 | 3-4 小时（Repository 新增方法 + Service 联动逻辑 + Migration） |
| P1-2 | 2 小时（模型 + Repository 完善） |
| P1-3 | 30 分钟（调用链已通，加一行调用） |
