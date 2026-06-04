# SPEC: 银行流水 A/B/C → 第一类/第二类框架对齐

## 背景

两套框架是正交的：
- **A/B/C**：处理状态（"这条流水现在到哪一步"）
- **第一类/第二类**：输入单据信息完整度（"这条流水含多少管理信息"）

当前 `bank_txn_review_service.go` 的 `classifyType()` 和 `SubmitReview()` 使用 A/B/C 描述，但注释和文档未用第一类/第二类框架解释判断逻辑。两套框架同时存在但缺乏显式对齐说明。

## 核心原则

**人是审核的唯一主体。** 系统生成的所有凭证（`DocStatus=0`）和所有 PaymentEntry（`DocStatus=0`）均为草稿状态，等待人审核。系统不执行任何最终核准动作。

 txn status 标签（如 `voucher_generated` / `payment_created`）的含义是"草稿已生成、等待审核"，不是"已完成"。

## 目标

1. **代码注释对齐**：用第一类/第二类语言重写 `classifyType` 和 `SubmitReview` 的注释，不改业务逻辑
2. **设计文档更新**：`bank-txn-input-classification-gap.md` 补充新框架的完整说明

## 代码改动

### 文件：`internal/service/bank_txn_review_service.go`

**classifyType() 注释升级**（不改函数体）：

```go
// classifyType returns the high-level type ("A", "B", or "C") for a given
// classification string.
//
// 第一类（可直接制证）：信息完整、无歧义，直接生成记账凭证
//   - bank_fee, interest_income, tax_payment, social_security,
//     insurance_fee → A类 → 直接 GenerateFromBankTxn
//
// 第二类（需中转）：
//   - internal_transfer → B类 → 生成 PaymentEntry（内部转账单）→ 直接制证
//   - business_receipt, business_payment, pay, receive, expense → B类
//     → 生成 PaymentEntry → 核销发票（如有）→ 生成凭证
//
// C类：无法自动分类，status=manual_pending，待人工处理工作台
```

**SubmitReview() case 分支注释**（不改逻辑）：

```go
// 第一类直接制证：银行费用/税费/社保/利息/保险费
case "A":
// 第二类需中转（生成 PaymentEntry）
case "B":
```

**文件头注释**：

```go
// BankTxnReviewService handles the review workflow for bank transactions.
// Transactions are first classified by classifyType(), then routed:
//   - 第一类（A）: 直接制证（银行费用/税费/社保/利息/保险）
//   - 第二类（B）: 生成 PaymentEntry 后制证（收付款往来/内部转账）
//   - C类: status=manual_pending，待处理工作台
```

## 设计文档改动

### 文件：`references/bank-txn-input-classification-gap.md`

在"两套分类体系对比"章节之后新增**框架对齐说明**：

```markdown
## 框架对齐说明：A/B/C 与 第一类/第二类

两套框架是正交的，同时生效：

| 维度 | 框架 | 含义 |
|------|------|------|
| 处理状态 | A/B/C | 这条流水"现在到哪一步" |
| 信息完整度 | 第一类/第二类 | 这条流水"含多少管理信息" |

**A/B/C 判断路径不变**，注释改为用第一类/第二类描述：

| classifyType 结果 | 第一类/第二类 | 含义 | 处理路径 |
|-------------------|---------------|------|---------|
| A | 第一类 | 信息完整，直接制证 | bank_fee/tax/social/interest/insurance → GenerateFromBankTxn |
| B | 第二类 | 需中转单据 | business_receipt/payment/internal_transfer → PaymentEntry → 凭证 |
| C | — | 无法分类 | manual_pending → 待处理工作台 |

**关键细节**：
- `internal_transfer`（内部转账）归入 B 类第二类，但因无外部往来，单据确认后直接制证，无核销环节
- `business_receipt/payment`（收付款）归入 B 类第二类，需生成 PaymentEntry 后匹配发票核销，再生成凭证
```

## 验证标准

- [ ] `go build ./...` 通过
- [ ] `classifyType` 逻辑不变（现有 A/B/C 判断结果不变）
- [ ] `SubmitReview` 逻辑不变（A→直接制证，B→PaymentEntry，C→跳过）
- [ ] `bank-txn-input-classification-gap.md` 文档包含框架对齐表格