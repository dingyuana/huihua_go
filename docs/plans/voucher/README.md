# 凭证模块

## 模块概述

记账凭证（JournalEntry）是业财一体化的财务数据桥梁，所有业务单据（银行流水、付款单、应收单、费用报销单等）最终都通过凭证接入总账。

## 凭证状态机

| 状态 | docstatus | 允许动作 | 说明 |
|------|-----------|----------|------|
| 草稿 | 0 | submit / cancel | 系统自动生成，人审后生效 |
| 已过账 | 1 | approve / reject / reverse / cancel | 借：银行 贷：应收/应付 等资金凭证 |
| 已审核 | 2 | reverse | 凭证已核准入账 |
| 已作废 | 3 | — | 作废后凭证不可用 |

### 特殊动作
- **reverse（红字冲销）**：原凭证标记 `reversed_id`，生成借贷对调的新凭证
- **cancel（作废）**：草稿状态下作废，上游业务单据应回退状态（待实现）

## 与上游业务单据的双向联动

### 当前已实现
- `voucher_auto_generate_service.go`：SubmitReview 后生成凭证草稿（docstatus=0）
- `voucher_state_machine.go`：完整状态流转（draft→posted→verified/cancelled/reversed）

### 待实现（P1优先级）
- 凭证过账后反向锁定上游业务单据（ArInvoice/PaymentEntry）
- 凭证草稿作废后回退上游业务单据状态
- 凭证审核后更新 bank_txn matched 标记

## 子文档索引

| 文档 | 标题 | 用途 |
|------|------|------|
| `SPEC-depreciation.md` | 折旧/摊销凭证 | 固定资产折旧+无形资产摊销自动生成凭证 |

## 依赖模块

- `bank-txn/`：第一类流水直接生成凭证
- `ar-invoice/`：应收单审核生成应收记账凭证
