# SPEC: D-P0.5 — 审核工作台 API

## 基本信息
- **任务 ID**: D-P0.5
- **类型**: feature
- **优先级**: P0
- **依赖**: D-P0.3（需要 ArInvoice 已可查询）
- **负责 Profile**: dev

## 背景
人工审核需要"待审任务池"视图，汇总展示所有草稿单据，供财务人员逐级审核。

## 目标
在 `internal/handler/audit_handler.go` 新增审核工作台 handler，注册路由 `GET /api/v1/audit/tasks`

## 响应结构
```go
type AuditTasksResult struct {
    InvoiceDrafts []InvoiceDraftSummary  `json:"invoice_drafts"`
    ArInvoices    []ArInvoiceSummary     `json:"ar_invoices"`
    Vouchers      []VoucherDraftSummary  `json:"vouchers"`
    Summary       AuditTaskSummary       `json:"summary"`
}

type AuditTaskSummary struct {
    InvoiceDraftCount int `json:"invoice_draft_count"`
    ArInvoiceDraftCount int `json:"ar_invoice_draft_count"`
    VoucherDraftCount int `json:"voucher_draft_count"`
    BlockedCount      int `json:"blocked_count"` // 凭证因上游草稿被阻断的数量
}
```

**汇总逻辑**：
- `InvoiceDrafts`：查询 `sales_invoices WHERE status = 'draft'`，取前50条
- `ArInvoices`：查询 `ar_invoices WHERE status = 'draft'`，取前50条
- `Vouchers`：查询 `journal_entries WHERE docstatus = 0`（草稿），取前50条
- `BlockedCount`：草稿凭证中，关联的 ArInvoice 仍为 draft 的数量（需 JOIN）

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] 路由 `GET /api/v1/audit/tasks` 返回上述结构
- [ ] 各分类列表条数正确（过滤条件正确）
- [ ] BlockedCount 逻辑正确（凭证草稿 且 关联 ArInvoice status=draft）

## 技术约束
- 新建 `internal/handler/audit_handler.go`（参照 invoice_handler.go 风格）
- 在 `main.go` 注册路由：`api.Get("/audit/tasks", handler.GetAuditTasks)`
- handler 中注入：`arInvoiceRepo`、`invoiceRepo`、`journalRepo`（或通过 service 封装）

## OpenCode 指令模板
**目标**：创建审核工作台 API

**约束**：
- 新建文件：`internal/handler/audit_handler.go`
- 注册路由到 main.go
- 只查不改数据（只读方法）

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- 参照：`internal/handler/invoice_handler.go`

**验收**：
- `go build ./...` 无报错
- `curl` 调用返回正确 JSON 结构