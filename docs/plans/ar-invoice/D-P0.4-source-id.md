# SPEC: D-P0.4 — 凭证 source_id 关联字段

## 基本信息
- **任务 ID**: D-P0.4
- **类型**: infra
- **优先级**: P0
- **依赖**: 无（可与 D-P0.1 并行）
- **负责 Profile**: dev

## 背景
凭证（JournalEntry）目前无追溯链，无法从凭证追溯到原始发票。三级草稿链路要求：凭证 → 应收单 → 发票 全链路可追溯。

## 目标
1. 创建 `migrations/045_journal_entries_source.sql` — 新增列：`source_type`, `source_id`, `source_invoice_id`
2. 修改 `internal/model/journal_entry.go` — 新增三个字段
3. 修改 `internal/service/voucher_auto_generate_service.go` 的 `GenerateFromInvoice` — 填充 source 字段

## 新增字段（JournalEntry）
```go
SourceType     string     `db:"source_type"`      // "invoice" | "bank_txn" | "payment"
SourceID       uuid.UUID  `db:"source_id"`        // 关联源单据ID（如 invoiceID）
SourceInvoiceID uuid.UUID `db:"source_invoice_id"` // 追溯到原始发票（跨级）
```

## GenerateFromInvoice 填充逻辑
```go
// 在 GenerateFromInvoice 方法中，创建 JournalEntry 后：
entry.SourceType = "invoice"
entry.SourceID = invoiceID
entry.SourceInvoiceID = invoiceID  // 直接关联发票，不需要跨级追溯
```

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] migration SQL 语法正确，可在 DB 执行
- [ ] 调用 `GenerateFromInvoice` 生成的凭证记录，source_type='invoice'，source_id=invoiceID

## 技术约束
- `JournalEntry` 已有 `SourceType` 等字段（可能已有类似字段），先读文件确认再决定是否新增
- migration 用 `ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS` 风格

## OpenCode 指令模板
**目标**：为 JournalEntry 添加 source 追溯字段，并在 GenerateFromInvoice 中填充

**约束**：
- 先 `read_file` 确认 `journal_entry.go` 现有字段，避免重复定义
- migration 使用 `ADD COLUMN IF NOT EXISTS` 防止重复执行报错

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- 修改文件：`journal_entry.go`、`voucher_auto_generate_service.go`、新 migration 文件

**验收**：
- `go build ./...` 无报错
- 数据库 migration 执行后，凭证记录包含 source 字段