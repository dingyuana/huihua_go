# 分析：PZ-20260603-331 为什么"主动完成科目"看起来没实现

## 基本信息

- **凭证**: PZ-20260603-331
- **日期**: 2026-06-03
- **来源**: REC-000004 (payment_type=receive, party_type=customer)
- **对方**: 北京数字认证股份有限公司
- **现有数据**:
  - 借：1002 银行存款 0.01
  - 贷：1122 应收账款 0.01
  - party_type='customer' ✓
  - user_remark='' ✗

---

## 现状诊断（数据库直查 + 代码追踪）

### 数据层面：是有的，但用户说"没有"

```
journal_entry_lines:
  1002 银行存款  借 0.01   ← 自动填入
  1122 应收账款  贷 0.01   ← 自动填入
```

**结论：后端是自动填了科目的**。用户说"没实现"指的是**几个具体场景下没主动生成**，不是完全没生成。

---

## "主动完成科目"在哪些场景下没工作

### 场景 1：凭证行 `user_remark` 是空的

**问题**：分录行的 `user_remark` 字段没有任何业务说明，比如应该填"收到北京数字认证股份有限公司货款 0.01"。

**当前代码**：`voucher_auto_generate_service.go::GenerateFromPaymentEntry` 创建分录行时**完全没填 user_remark**。

```go
lines := []model.JournalEntryLine{
    {
        ID:             uuid.New(),
        JournalEntryID: je.ID,
        AccountID:      debitAccountID,
        Debit:          amount,
        Credit:         decimal.Zero,
        PartyType:      &partyTypeStr,
        PartyID:        &partyIDCopy,
        // ❌ 没填 user_remark
    },
    { ... }, // ❌ 同样没填
}
```

**应该填什么**：
- 借方（1002 银行存款）："收款" 或 "收到北京数字认证股份有限公司货款"
- 贷方（1122 应收账款）："应收北京数字认证股份有限公司"

### 场景 2：用户手动新增凭证时，没有"智能推荐科目"

**问题**：在 `VoucherEdit.vue` 手动新增凭证时，用户输入摘要"收到北京数字认证股份有限公司货款"，没有任何智能提示告诉用户应该用什么科目。

**当前代码**：`VoucherEdit.vue` 第 163-167 行只初始化了空行，**没有任何 watch/onChange 触发"输入摘要→推荐科目"**。

```typescript
const form = reactive({ date: '', type: '记', remark: '' })
const lines = ref<LineItem[]>([
  { account: null, debit: '', credit: '' },  // ❌ 默认空
  { account: null, debit: '', credit: '' },
])
```

**应该补什么**：
- 用户在 `form.remark` 输入内容时
- 调后端 API `POST /v1/vouchers/suggest-accounts?remark=...`
- 后端用分类规则匹配 + 启发式推断返回最可能的借/贷科目组合
- 前端自动填入 lines[0] 和 lines[1]

### 场景 3：`party_id` 是 zero-uuid，没有用上 `parties` 表的真实数据

**问题**：payment_entries.party_id = '00000000-0000-0000-0000-000000000000'，而 `parties` 表是空的（0 rows）。

**根因**：
- `CreateFromBankTxn` 在 `bank_txn_review_service.go` 里硬编码了 `PartyID: uuid.Nil`
- `parties` 表没有数据，也**没有** `ar_account_id`/`ap_account_id` 列
- 所以 `resolveCounterAccount` 的 Priority 1（"Party 默认 AR/AP 账户"）**永远走不到**

**结果**：科目只能落到 Priority 3（智能映射）→ 1122 by default。

### 场景 4：多行分录场景没处理

**问题**：如果一笔记账涉及多个科目（比如部分核销、部分挂账），现在只生成 2 行固定借贷。

**当前代码**：硬编码两行 debit + credit。

**应该补什么**：根据 `payment_allocations` 拆分多行（部分核销到具体发票）。

### 场景 5：摘要格式固定，不能体现业务含义

**当前**：`remark = "[REC-000004] 北京数字认证股份有限公司"`（仅单据号+对方户名）

**应该**：`remark = "收到[北京数字认证股份有限公司]货款 [REC-000004]"` （更符合会计实务）

---

## 根因总结

| 场景 | 现状 | 期望 |
|------|------|------|
| 后端自动填借贷科目 | ✅ 已实现 | 保持 |
| 分录行 `user_remark` | ❌ 永远为空 | 应自动填业务说明 |
| 手动新增凭证时推荐科目 | ❌ 完全没做 | 输入摘要后自动推荐 |
| Party 真实数据 | ❌ 永远 zero-uuid | 应支持真实客商挂接 |
| 多行分录 | ❌ 固定 2 行 | 应支持按核销拆分 |
| 摘要格式 | ⚠️ 仅单据号+对方 | 应业务化 |

---

## 解决路径（按优先级）

### 优先级 P0：基础体验修复

1. **分录行 `user_remark` 自动填**
   - 后端在 `GenerateFromPaymentEntry` 加 `UserRemark` 赋值
   - 借方行："收到{对方户名} {payment_type_zh} {金额}"
   - 贷方行："应收/应付{对方户名}"

2. **前端编辑页"输入摘要→推荐科目"按钮**
   - 加一个"智能推荐"按钮
   - 后端新增 `POST /api/v1/vouchers/suggest-accounts` API
   - 用 `classification_rules` 匹配 + 关键词启发式（货款/服务费/差旅费/工资 等）
   - 至少能让"收到XX货款"自动推出 1002/1122

### 优先级 P1：数据完善

3. **真实客商挂接**
   - `parties` 表加 `default_ar_account_id`、`default_ap_account_id` 列
   - bank_transactions 导入时尝试按 `counterparty_name` 匹配/自动创建 party

4. **多行分录支持**
   - 检查 `payment_allocations`，按发票拆分多行

### 优先级 P2：智能增强（V2 远期）

5. **AI 推荐科目**（已在 SPEC §十二列出）
   - 接入 LLM，对摘要做 NER
   - 返回 top-3 候选科目
   - 会计确认后写入

---

## 给用户的结论

**"主动完成科目"在 PZ-20260603-331 这个例子里部分工作了**：
- ✅ 借贷两行科目都自动填了
- ❌ 分录行的 `user_remark` 是空的（最显眼的问题）
- ❌ 手动新增凭证时没推荐机制

**最影响体验的两件事**（按工时从小到大）：

1. **分录行 `user_remark` 主动填**（0.5 小时）
   - 后端 1 行赋值即可
   - 立刻改善"看到凭证但看不出是干什么的"问题

2. **VoucherEdit 摘要输入→推荐科目按钮**（2-3 小时）
   - 需要后端新 API + 前端 watch
   - 让"打一行字就生成凭证"成为可能

要我先做 1（user_remark），还是 1+2 一起做？
