# 银行流水智能处理与对账模块

## 模块概述

银行流水是业财一体化的核心入口，通过"第一类直接制证 / 第二类生成付款单 / C类人工处理"三条路径，将银行交易数据转化为记账凭证，实现业务单据管"事"、记账凭证管"账"、银行日记账管"钱"的三层架构。

## 核心设计原则

1. **人本机制**：所有系统生成单据均为 docstatus=0 草稿，人工审核后才生效
2. **AI辅助**：分类引擎（规则+AI双轨）识别流水类型，第一类/第二类自动分流
3. **原子性提交**：SubmitReview 一次性完成分类+制证/生成付款单，保证数据一致性
4. **五维打分勾对**：金额50+日期20+户名15+摘要10+凭证号5，满分100，≥85自动勾兑

## 状态机

```
银行流水状态流转：
  pending → classified → approved → [第一类:生成凭证草稿 | 第二类:生成付款单草稿]
  → draft_voucher_generated / draft_payment_generated → approved → 过账/确认

凭证状态机（独立）：
  draft(0) → posted(1) → verified(2)
                   ↘ cancelled(3)
                   ↘ reversed(红字冲销)
```

## 子文档索引

| 文档 | 标题 | 用途 |
|------|------|------|
| `SPEC-classification.md` | 分类框架 | 第一类/第二类/C类定义及 classifyType 逻辑 |
| `SPEC-review-workflow.md` | 审核流程 | SubmitReview → 草稿生成 → 人工审核 → 过账完整链路 |
| `SPEC-c-workbench.md` | C类工作台 | C类流水的 ProcessManual 接口（人工选择路径） |
| `SPEC-reconciliation.md` | 余额调节表 | P2-3 五维打分自动勾对 + 四类未达账项 |
| `SPEC-payment-link.md` | 付款单联动 | PaymentEntry 审核后自动核销发票 |
| `SPEC-roadmap.md` | 路线图 | 银行流水全链路改进计划 Phase1-4 |

## 依赖模块

- `ar-invoice/`：第二类流水生成的付款单需与 ArInvoice/Invoice 核销
- `voucher/`：第一类流水和第二类链路最终都生成 JournalEntry（记账凭证）
