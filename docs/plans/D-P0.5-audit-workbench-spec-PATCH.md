# SPEC Patch: D-P0.5 — 审核工作台（补充阻断统计 + 过账动作）

## 基本信息
- **任务 ID**: D-P0.5-v2
- **类型**: spec-patch
- **优先级**: P1
- **依赖**: D-P0.5（原有逻辑）
- **补丁目标**: `docs/plans/D-P0.5-audit-workbench-spec.md`

---

## 变更 1：BlockedCount 分层统计（新增）

### 背景
原 SPEC 的 `BlockedCount` 只统计"凭证因 ArInvoice draft 被阻断"，但新文档 §5 明确了两类阻断：
1. **应收单草稿阻断**：凭证草稿已生成，但关联的 ArInvoice 仍为 draft 状态
2. **客户草稿阻断**：若有客户草稿状态，客户草稿会阻断其下级单据（我们当前不做客户草稿，此条仅供参考）

### 修正后的汇总统计
```go
type AuditTaskSummary struct {
    InvoiceDraftCount   int `json:"invoice_draft_count"`
    ArInvoiceDraftCount int `json:"ar_invoice_draft_count"`
    VoucherDraftCount   int `json:"voucher_draft_count"`

    // 两层阻断统计（新增）
    ArInvoiceBlockedCount int `json:"ar_invoice_blocked_count"` // 凭证因 ArInvoice=draft 无法过账
    InvoiceBlockedCount    int `json:"invoice_blocked_count"`    // 应收单因发票草稿无法确认（本系统已设为自动审核发票草稿，实际应为 0）
}
```

### 阻断逻辑说明
| 层级 | 阻断关系 | 说明 |
|------|---------|------|
| 发票草稿 → 应收单草稿 | 发票 `status=draft` → `ar_invoices` 不可 confirm | 当前设计为导入即自动审核发票草稿，此阻断实际为 0 |
| 应收单草稿 → 凭证草稿 | ArInvoice `status=draft` → `journal_entries` 不可过账 | 需 JOIN `journal_entries.source_id = ar_invoices.id` |

### SQL 查询（BlockedCount）
```sql
-- ArInvoiceBlockedCount: 凭证草稿(docstatus=0) 且 source_type='invoice' 且 source_id 指向的 ArInvoice 仍为 draft
SELECT COUNT(DISTINCT je.id)
FROM journal_entries je
LEFT JOIN ar_invoices ai ON ai.id = je.source_id::uuid
WHERE je.docstatus = 0
  AND je.source_type = 'invoice'
  AND ai.status = 'draft'
  AND je.tenant_id = $1;
```

### 验收标准（补充）
- [ ] `ar_invoice_blocked_count` 正确反映因 ArInvoice draft 无法过账的凭证数量
- [ ] `invoice_blocked_count` 应为 0（发票草稿在当前设计中被视为自动审核）

---

## 变更 2：过账后 Audit Trail（新增）

### 背景
新文档 §3.5.4 要求：过账后系统自动记录审核人、审核时间，在应收单和凭证上留下审计痕迹。

### 设计方案
**方案 A（轻量）**：在 `ar_invoices` 和 `journal_entries` 表上加 `approved_by` / `approved_at` 字段
**方案 B（通用）**：新建 `audit_logs` 表，记录所有单据的状态变更

**推荐方案 A（当前阶段优先）**：
- `ar_invoices` 表新增 `approved_by` (UUID) + `approved_at` (TIMESTAMP)
- `journal_entries` 表新增 `approved_by` (UUID) + `approved_at` (TIMESTAMP)
- 过账 API 在执行时写入这两个字段

### 影响文件

| 表 | 新增字段 |
|----|---------|
| `ar_invoices` | `approved_by UUID` + `approved_at TIMESTAMP` |
| `journal_entries` | `approved_by UUID` + `approved_at TIMESTAMP` |

### 验收标准（补充）
- [ ] 过账后查询 `ar_invoices.approved_by` 和 `approved_at` 有值
- [ ] 过账后查询 `journal_entries.approved_by` 和 `approved_at` 有值
- [ ] 审核人 ID 为执行过账操作的用户 ID（从 JWT claim 获取）

---

## 影响文件

| 文件 | 变更 |
|------|------|
| `docs/plans/D-P0.5-audit-workbench-spec.md` | 本补丁合并到原 SPEC |
| 新建 `migrations/047_audit_fields.sql` | `ar_invoices` + `journal_entries` 新增 approved_by/approved_at |
| `internal/model/ar_invoice.go` | 新增 ApprovedBy / ApprovedAt 字段 |
| `internal/model/journal.go` | 新增 ApprovedBy / ApprovedAt 字段 |
| `internal/handler/audit_handler.go` | `GetAuditTasks` 返回结构体补充两层阻断计数 |
| `internal/service/ar_invoice_service.go`（若存在）| 过账时写入 approved_by/approved_at |
| `internal/service/voucher_service.go` | 过账时写入 approved_by/approved_at |