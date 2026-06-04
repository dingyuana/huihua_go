# SPEC: D-P1.2 — 导入预览增强

## 基本信息
- **任务 ID**: D-P1.2
- **类型**: feature
- **优先级**: P1
- **依赖**: D-P0.1
- **负责 Profile**: dev

## 背景
Preview 当前返回字段校验+重复检查结果，缺少"客户匹配详情"和"将生成单据汇总"信息。

## 目标
修改 `BatchImportPreviewResult`（在 `internal/model/invoice.go` 中），新增字段：

```go
type BatchImportPreviewResult struct {
    // 已有（保留）
    Total   int
    Valid   int
    Errors  int
    Details []PreviewRowDetail
    // 新增
    CustomerMatches []CustomerMatchInfo `json:"customer_matches"`
    WillGenerateAs  WillGenerateSummary `json:"will_generate_as"`
}

type CustomerMatchInfo struct {
    RowIndex       int    `json:"row_index"`
    Status         string `json:"status"`          // "matched" | "fuzzy" | "auto_created"
    CustomerID     *string `json:"customer_id,omitempty"`
    CustomerName   string `json:"customer_name"`
    WarningMessage string `json:"warning_message,omitempty"`
}

type WillGenerateSummary struct {
    InvoicesWillCreate  int `json:"invoices_will_create"`
    ARWillCreate       int `json:"ar_will_create"`
    VouchersWillCreate int `json:"vouchers_will_create"`
}
```

**填充逻辑**（在 `BatchImportPreview` 方法中）：
- `CustomerMatches`：每行调用 `resolveCustomer`，记录匹配状态
- `WillGenerateAs.InvoicesWillCreate = Valid`（通过校验的行都会生成发票草稿）
- `WillGenerateAs.ARWillCreate = 0`（预览阶段不生成 ArInvoice，确认时才生成）
- `WillGenerateAs.VouchersWillCreate = 0`（同上）

## 验收标准
- [ ] `go build ./...` 编译通过
- [ ] Preview 响应包含 `customer_matches` 和 `will_generate_as` 字段
- [ ] `customer_matches` 每行一条记录

## 技术约束
- 需在 model 文件新增结构体 `CustomerMatchInfo` 和 `WillGenerateSummary`
- 不要修改现有的 `PreviewRowDetail` 结构
- Preview 是只读操作，不实际写入 DB

## OpenCode 指令模板
**目标**：增强 BatchImportPreview 响应

**约束**：
- 只修改 `internal/model/invoice.go`（新增结构体）和 `internal/service/invoice_service.go`（填充逻辑）

**上下文**：
- 项目：`/root/data/disk/huihua-finance`

**验收**：
- `go build ./...` 无报错
- Preview 接口返回新字段（可用 curl 验证）