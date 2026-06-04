# SPEC: D-P0.1 — ArInvoice 模型 + Migration

## 基本信息
- **任务 ID**: D-P0.1
- **类型**: infra
- **优先级**: P0
- **依赖**: 无
- **负责 Profile**: dev

## 背景
三级草稿链路缺失，根源是无 ArInvoice（应收单）模型。`ConfirmSalesInvoice` 仅改发票 status，无法生成应收单和凭证。本任务创建 ArInvoice 模型及 DB 层。

## 目标
1. 创建 `internal/model/ar_invoice.go` — ArInvoice 模型定义
2. 创建 `migrations/043_ar_invoices.sql` — 建表语句
3. 创建 `migrations/044_ar_invoice_indexes.sql` — 索引

## 验收标准
- [ ] `go build ./...` 编译通过（model 新增文件）
- [ ] migration SQL 可在 DB 执行（语法正确）
- [ ] ArInvoice 字段覆盖：ID, TenantID, CompanyID, CustomerID, InvoiceID, InvoiceNo, Amount, DueDate, Status, SourceType, CreatedBy, CreatedAt, ConfirmedAt, ConfirmedBy

## 技术约束
- 模型文件放 `internal/model/ar_invoice.go`
- migration 序号接续（上一条是 042）
- Status 枚举：`draft` / `confirmed` / `reversed`
- `Amount` 用 `decimal.Decimal`，`DueDate` 用 `*time.Time`

## OpenCode 指令模板
**目标**：在 `/root/data/disk/huihua-finance` 创建 ArInvoice 模型 + 2个 migration 文件

**约束**：
- 只创建新文件，不修改现有文件
- migration SQL 语法正确（PostgreSQL）
- model 使用 shopspring/decimal 和 google/uuid

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- spec：本文件

**验收**：
- `go build ./...` 无报错
- migration 文件存在且语法正确