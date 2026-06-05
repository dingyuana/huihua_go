# 应收模块 (ArInvoice)

## 模块概述

ArInvoice（应收单）是销售发票确认后的中间状态单据，承载"销售发票 → 应收单 → 应收记账凭证 → 收款确认 → 资金记账凭证"完整应收闭环的核心链路。

## 核心链路

```
销售发票确认
  ↓
ArInvoice（草稿） + 应收记账凭证（草稿）
  ↓ 人工审核
ArInvoice（已确认）+ 凭证（已过账 posted）
  ↓ 收款银行流水到账
PaymentEntry（核准） → 发票核销 → 资金记账凭证
  ↓
ArInvoice（已核销）+ 银行日记账（已记账）
```

## 子文档索引

| 文档 | 标题 | 用途 |
|------|------|------|
| `D-P0.1-model.md` | ArInvoice 数据模型 | 模型定义 + DB建表 Migration |
| `D-P0.2-repo.md` | Repository 接口 | Create/GetByID/ListByTenant 等 7 个方法 |
| `D-P0.3-confirm.md` | 发票确认逻辑 | ConfirmSalesInvoice 重写，确认时生成 ArInvoice 草稿 |
| `D-P0.4-source-id.md` | 凭证关联字段 | JournalEntry 新增 source_type/source_id/source_invoice_id 追溯链 |
| `D-P0.5-audit-workbench.md` | 审核工作台 | 待审任务池视图 + 阻断统计（ArInvoice/Invoice 分层） |
| `D-P1.1-auto-customer.md` | 自动创建客户 | 税号查不到时名称相似度模糊匹配 + 并发 Upsert 保护 |
| `D-P1.2-preview.md` | 导入预览增强 | BatchImportPreviewResult 新增 CustomerMatches + WillGenerateSummary |
| `D-P1.3-batch-approve.md` | 批量审核 | 批量确认发票 + 批量过账凭证，跳过上游草稿未审核项 |
| `D-P1.4-exception.md` | 异常工作台 | 硬拒绝/阻断条目 + 凭证→发票全链路追溯 |

## 当前状态

⚠️ ArInvoice model 存在（`internal/model/ar_invoice.go`），但无完整的 Service 层实现。
