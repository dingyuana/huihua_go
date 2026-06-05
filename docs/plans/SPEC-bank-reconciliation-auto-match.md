# SPEC: P2-3 余额调节表 + 银企自动勾对

## 背景

P2-3 是银行对账的最后一个子模块，包含两大功能：
1. **多维度打分自动勾对** — 加权打分制，≥85 分自动勾兑，60-85 分待确认，<60 分进余额调节表
2. **余额调节表** — 四类未达账项：银行已收企业未达 / 银行已付企业未达 / 企业已收银行未达 / 企业已付银行未达

当前状态：BankReconciliationService 仅用 date+amount 精确匹配（无打分），四类未达账项 DB 枚举不完整，SaveReconciliationResult 未写明细，BalanceSheet.vue 有 UI 但数据 fallback 到 hardcoded，对账锁定无 Service 方法。

## 目标

完整实现 TASK-F4.1 的全部子功能。

---

## 一、打分勾对算法

### MatchScore 结构体

```go
type MatchScore struct {
    TotalScore    float64 // 总分 0-100
    AmountScore  float64 // 金额 0-50（完全一致得 50，差异 1% 内按比例衰减）
    DateScore    float64 // 日期 0-20（同一天得 20，±1 天按比例衰减，±3 天以上 0）
    NameScore    float64 // 户名 0-15（Levenshtein ≥90% 得 15，80%-90% 得 10，70%-80% 得 5）
    DescScore    float64 // 摘要 0-10（命中预设关键词得 10，部分命中得 5）
    RefNoScore   float64 // 流水号 0-5（精确匹配得 5）
    IsAutoMatched bool   // ≥85 分
    NeedConfirm  bool    // 60-85 分
    Candidate    *uuid.UUID // 匹配的 GL Entry ID
}
```

### 五维打分规则

| 维度 | 权重 | 满分 | 评分规则 |
|------|------|------|---------|
| 金额匹配 | 50% | 50 | 完全一致 = 50；差异 1% 内按比例衰减 |
| 日期匹配 | 20% | 20 | 同一天 = 20；±1 天按比例衰减；±3 天以上 = 0 |
| 对方户名相似度 | 15% | 15 | Levenshtein ≥90% = 15；80%-90% = 10；70%-80% = 5 |
| 摘要关键词匹配 | 10% | 10 | 命中预设关键词 = 10；部分命中 = 5 |
| 银行流水号精确匹配 | 5% | 5 | 流水号完全匹配 = 5 |

### 分档阈值

- **≥85 分**：自动勾兑，无需人工确认
- **60-85 分**：推送至待确认列表，会计确认后勾兑
- **<60 分**：进入余额调节表

### 一对多/多对一支持

- POS 汇总入账：同一天同一客户多笔小额 sum ≈ 大额发票，按比例分配
- 分批付款：一笔发票分多笔银行转账支付，聚合匹配

---

## 二、四类未达账项

### DB Schema 变更

`migrations/050_reconciliation_items_extend.sql`：

```sql
-- 扩展 item_type 枚举，支持方向
ALTER TABLE unreconciled_items DROP CONSTRAINT IF EXISTS unreconciled_items_item_type_check;
ALTER TABLE unreconciled_items ADD CONSTRAINT unreconciled_items_item_type_check
    CHECK (item_type IN (
        'bank_receipt_not_in_gl',   -- 银行已收企业未达
        'bank_payment_not_in_gl',   -- 银行已付企业未达
        'gl_receipt_not_in_bank',   -- 企业已收银行未达
        'gl_payment_not_in_bank',    -- 企业已付银行未达
        'bank_only',                 -- 兼容旧数据
        'book_only'                  -- 兼容旧数据
    ));
```

### Service 层变更

**ReconcileBankAccount** 重写为打分制：
1. 对每条未匹配 bank_txn，按 date±3天 + amount±1% 筛选 GL 候选
2. 对每个候选计算 MatchScore 五维得分
3. 最高分 ≥85 → 自动勾兑（更新 matched=true + matched_gl_entry_id）
4. 最高分 60-85 → 进入待确认列表（pending_confirm）
5. 无候选或最高分 <60 → 进入余额调节表四类

**SaveReconciliationResult** 补写：
- 对每条 bankOnlyItems，判定是 receipt 还是 payment（按 direction）
- 对每条 bookOnlyItems，判定是 receipt 还是 payment
- INSERT 到 unreconciled_items 表，item_type 填对应四类枚举值

---

## 三、余额调节表

### API 端点

```
GET /api/v1/bank-reconciliation/items?bank_account_id=&period_no=&item_type=
```

返回结构：
```json
{
  "bank_receipt_not_in_gl": [...UnreconciledItem],
  "bank_payment_not_in_gl": [...UnreconciledItem],
  "gl_receipt_not_in_bank": [...UnreconciledItem],
  "gl_payment_not_in_bank": [...UnreconciledItem]
}
```

### BalanceSheet.vue 重构

- 移除 hardcoded fallback 数据
- API 异常时显示错误提示，不回退到示例数据
- 四个卡片分别展示四类未达账项

### 待确认列表

```
GET /api/v1/bank-reconciliation/pending-confirm?bank_account_id=&period_no=
```

前端 `PendingConfirmView.vue` — 展示 60-85 分候选对，供会计逐笔确认。

---

## 四、对账锁定

### 新增 Service 方法

```go
func (s *BankReconciliationService) LockReconciliation(ctx, tenantID, bankAccountID, periodNo, userID) error
func (s *BankReconciliationService) UnlockReconciliation(ctx, tenantID, bankAccountID, periodNo, userID) error // 需主管审批
```

### DB 更新

```sql
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked BOOLEAN DEFAULT FALSE;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked_by UUID;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS unlock_approved_by UUID;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS unlock_approved_at TIMESTAMPTZ;
```

### 锁定规则

- 对账结果确认后 → locked=true + locked_by + locked_at
- 锁定后修改匹配状态须主管审批（ApprovalFlow）
- 解锁创建审批任务 → 主管审批 → UnlockReconciliation

### 前端

`LockButton` — 锁定按钮，点击后调用 POST `/bank-reconciliation/lock`，成功后灰化按钮。

---

## 五、验收标准

- [ ] 50 条银行流水 + 50 条 GL，自动勾兑率 ≥ 90%
- [ ] 四类未达账项正确归类写入 DB
- [ ] 余额调节表四类卡片展示真实数据（无 hardcoded）
- [ ] 对账锁定后修改匹配状态须主管审批，审计日志记录
- [ ] 一对多/多对一匹配，金额按比例分配
- [ ] go build ./... 通过

---

## 文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `migrations/050_reconciliation_items_extend.sql` | 新建 | 扩展 item_type 枚举 |
| `migrations/051_reconciliation_lock.sql` | 新建 | locked 相关字段 |
| `internal/service/bank_reconciliation_service.go` | 重写 | 打分引擎 + 四类写入 + Lock/Unlock |
| `internal/handler/bank_reconciliation_handler.go` | 修改 | 挂载 Lock/Unlock/PendingConfirm 端点 |
| `cmd/api/main.go` | 修改 | 注册新路由 |
| `internal/model/bank.go` | 修改 | ReconciliationRecord 增加 Locked 等字段 |
| `frontend/src/views/reconciliation-bank/PendingConfirmView.vue` | 新建 | 待确认列表页 |
| `frontend/src/views/reconciliation-bank/BalanceSheet.vue` | 修改 | 移除 hardcoded fallback |
| `frontend/src/views/reconciliation-bank/MatchingView.vue` | 修改 | 展示锁定状态 + 锁定按钮 |
| `frontend/src/api/modules/reconciliation.ts` | 修改 | 增加 lock/unlock/pending-confirm API |
| `frontend/src/router/routes/base.ts` | 修改 | 注册 pending-confirm 路由 |