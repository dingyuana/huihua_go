# SPEC: D-P1.4 — 异常工作台与全链路追溯

## 基本信息
- **任务 ID**: D-P1.4
- **类型**: feature
- **优先级**: P1
- **依赖**: D-P0.3（ArInvoice 草稿链路）+ D-P0.4（凭证 source_id 追溯链）
- **负责 Profile**: dev

---

## 1. 背景与定位

### 业务定位
**异常工作台**是"正常单据流"之外的辅助处理界面，专门处理：
- 校验引擎硬拒绝的条目（如税额严重不平、重复发票号、购方名称和税号同时为空）
- 被阻断但需要人工干预才能继续的单据

**全链路追溯**是审计和追责工具，支持从任意正式单据逆向追溯到原始导入文件。

### 与审核工作台（D-P0.5）的区别

| 维度 | 审核工作台（D-P0.5） | 异常工作台（D-P1.4） |
|------|--------------------|--------------------|
| 处理对象 | 草稿单据（待人工确认） | 异常/被阻断的单据 |
| 操作 | 审核、过账 | 查看、修复、重试、放弃 |
| 触发条件 | 草稿状态 | 校验失败 OR 关联实体缺失 |
| 目标 | 将草稿推进为正式 | 清除异常、恢复流程 |

---

## 2. 异常工作台 API

### 路由
```
GET /api/v1/audit/exceptions
```

### 响应结构
```go
type ExceptionWorkbenchResult struct {
    HardRejections  []HardRejectionItem  `json:"hard_rejections"`   // 硬拒绝条目
    BlockedItems    []BlockedItem         `json:"blocked_items"`      // 被阻断条目
    Summary         ExceptionSummary      `json:"summary"`
}

type HardRejectionItem struct {
    RowIndex        int     `json:"row_index"`         // Excel 行号
    InvoiceNo       string  `json:"invoice_no"`
    CustomerName    string  `json:"customer_name"`
    ErrorCode       string  `json:"error_code"`        // "DUPLICATE_INVOICE_NO" | "TAX_AMOUNT_MISMATCH" | "MISSING_CUSTOMER_INFO" | ...
    ErrorMessage    string  `json:"error_message"`
    SourceFile      string  `json:"source_file"`       // 原始文件名
    ImportedAt      string  `json:"imported_at"`       // ISO 时间
}

type BlockedItem struct {
    ArInvoiceID     string  `json:"ar_invoice_id"`
    ArInvoiceNo     string  `json:"ar_invoice_no"`
    CustomerName    string  `json:"customer_name"`
    BlockReason     string  `json:"block_reason"`      // "CUSTOMER_DRAFT" | "MISSING_TAX_CLASSIFICATION" | ...
    BlockedAt       string  `json:"blocked_at"`
}

type ExceptionSummary struct {
    HardRejectionCount int `json:"hard_rejection_count"`
    BlockedCount        int `json:"blocked_count"`
}
```

### 异常类型编码

| error_code | 说明 | 可处理动作 |
|-----------|------|-----------|
| `DUPLICATE_INVOICE_NO` | 与系统中已存在的发票号冲突 | 修改发票号（需新建） |
| `TAX_AMOUNT_MISMATCH` | 税额与价税合计-金额不符，偏差>0.1元 | 手动调整税额或金额 |
| `MISSING_CUSTOMER_INFO` | 购方名称和税号同时为空 | 提供购方信息 |
| `RED_INVOICE_NO_ORPHANED` | 红字发票无法关联到蓝字 | 补填蓝字号或标记放弃 |
| `IMPORT_FILE_CORRUPTED` | Excel/CSV 解析失败 | 重新上传 |
| `CUSTOMER_DRAFT_BLOCKED` | 客户草稿阻断（仅当我们有客户草稿时） | 审核客户草稿 |

---

## 3. 全链路追溯 API

### 路由
```
GET /api/v1/audit/trace/:voucher_id
```

### 响应结构
```go
type TraceResult struct {
    Voucher          VoucherTrace         `json:"voucher"`
    ArInvoice        *ArInvoiceTrace      `json:"ar_invoice,omitempty"`
    Invoice          *InvoiceTrace        `json:"invoice,omitempty"`
    ImportBatch      *ImportBatchTrace    `json:"import_batch,omitempty"`
    SourceFile       string               `json:"source_file,omitempty"`    // 原始文件名（含路径）
    AuditLog         []AuditLogEntry      `json:"audit_log"`                  // 状态变更历史
}

type VoucherTrace struct {
    ID           string  `json:"id"`
    VoucherNo    string  `json:"voucher_no"`
    DocStatus    string  `json:"doc_status"`
    TotalAmount  float64 `json:"total_amount"`
    CreatedAt    string  `json:"created_at"`
    CreatedBy    string  `json:"created_by"`
    SourceType   string  `json:"source_type"`   // "invoice"
    SourceID     string  `json:"source_id"`     // ArInvoice ID
}

type ArInvoiceTrace struct {
    ID           string  `json:"id"`
    InvoiceNo    string  `json:"invoice_no"`
    CustomerName string  `json:"customer_name"`
    TotalAmount  float64 `json:"total_amount"`
    Status       string  `json:"status"`
    CreatedAt    string  `json:"created_at"`
    SourceType   string  `json:"source_type"`   // "invoice"
    SourceID     string  `json:"source_id"`     // 源头发票 ID
}

type InvoiceTrace struct {
    ID           string  `json:"id"`
    InvoiceNo    string  `json:"invoice_no"`
    CustomerName string  `json:"customer_name"`
    TotalAmount  float64 `json:"total_amount"`
    Status       string  `json:"status"`
    CreatedAt    string  `json:"created_at"`
    SourceFile   string  `json:"source_file,omitempty"`
    ImportBatchID string `json:"import_batch_id,omitempty"`
}

type ImportBatchTrace struct {
    ID           string  `json:"id"`
    FileName     string  `json:"file_name"`
    ImportedBy   string  `json:"imported_by"`
    ImportedAt   string  `json:"imported_at"`
    TotalCount   int     `json:"total_count"`
    SuccessCount int     `json:"success_count"`
    FailedCount  int     `json:"failed_count"`
}

type AuditLogEntry struct {
    EntityType   string  `json:"entity_type"`   // "voucher" | "ar_invoice" | "invoice"
    EntityID     string  `json:"entity_id"`
    Action       string  `json:"action"`         // "created" | "confirmed" | "posted" | "reversed"
    OperatorID   string  `json:"operator_id"`
    OperatorName string  `json:"operator_name"`
    OperatedAt   string  `json:"operated_at"`
    Remark       string  `json:"remark,omitempty"`
}
```

### 追溯链
```
凭证 → source_id → ArInvoice → source_id → 发票 → import_batch_id → 原始文件
                            ↓
                      audit_log（每级状态变更历史）
```

---

## 4. 数据模型变更

### 4.1 新建 import_batches 表（记录导入批次）
```sql
CREATE TABLE import_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    file_name VARCHAR(255) NOT NULL,
    file_hash VARCHAR(64),           -- SHA256 用于去重
    imported_by UUID NOT NULL REFERENCES users(id),
    imported_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total_count INT DEFAULT 0,
    success_count INT DEFAULT 0,
    failed_count INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'completed', -- 'processing' | 'completed' | 'failed'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_import_batches_tenant ON import_batches(tenant_id);
CREATE INDEX idx_import_batches_imported_at ON import_batches(imported_at);
```

### 4.2 新建 audit_logs 表（审计日志）
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    entity_type VARCHAR(20) NOT NULL,   -- 'voucher' | 'ar_invoice' | 'invoice' | 'party'
    entity_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL,
    operator_id UUID REFERENCES users(id),
    operator_name VARCHAR(100),
    operated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    extra JSONB,                        -- 存储变更前后的快照
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id);
```

### 4.3 sales_invoices 表新增字段
```sql
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS import_batch_id UUID REFERENCES import_batches(id);
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS source_file VARCHAR(255);
```

### 4.4 journal_entries 表新增字段
```sql
-- 用于全链路追溯
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS import_batch_id UUID REFERENCES import_batches(id);
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS source_file VARCHAR(255);
```

---

## 5. API 详细设计

### 5.1 GET /api/v1/audit/exceptions

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `tenant_id` | UUID | 必填（从 JWT 获取） |
| `error_code` | string | 过滤特定异常类型 |
| `date_from` | date | 导入起始日期 |
| `date_to` | date | 导入截止日期 |
| `page` | int | 页码（默认 1） |
| `page_size` | int | 每页条数（默认 20，上限 100） |

**权限**：仅 `accounting` 和 `admin` 角色可访问

### 5.2 GET /api/v1/audit/trace/:voucher_id

**路径参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `voucher_id` | UUID | 凭证 ID |

**查询参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `tenant_id` | UUID | 必填（从 JWT 获取） |

**逻辑**：
1. 查询 `journal_entries WHERE id = $voucher_id`，获取 `source_type` / `source_id` / `source_invoice_id`
2. 若 `source_type = 'invoice'`，查 `ar_invoices WHERE id = source_id`
3. 继续向上追溯到 `sales_invoices` → `import_batch_id`
4. 查 `audit_logs` 按 entity_type + entity_id 聚合所有状态变更

**权限**：仅 `accounting` 和 `admin` 角色可访问

---

## 6. 验收标准

- [ ] `GET /api/v1/audit/exceptions` 返回硬拒绝列表和被阻断条目
- [ ] 硬拒绝条目包含行号、发票号、错误码、错误信息
- [ ] `GET /api/v1/audit/trace/:voucher_id` 可追溯到源头发票和导入批次
- [ ] 追溯链路中每级都有 `audit_log` 记录
- [ ] `import_batches` 表记录每次导入的文件名、时间、成功率
- [ ] `audit_logs` 表记录所有单据的状态变更历史

---

## 7. 技术约束

- 追溯查询避免 N+1，用 JOIN 在单次查询中完成
- `audit_logs` 表写入使用异步（goroutine + channel）避免影响主事务性能
- 异常工作台分页限制 100 条/页
- `import_batch_id` 在发票导入 Confirm 阶段写入（需要修改 `BatchImportConfirm`）

---

## 8. OpenCode 指令模板

**目标**：实现异常工作台 + 全链路追溯功能

**约束**：
- 新建 `internal/handler/audit_handler.go` 的 `GetExceptions` 方法
- 新建 `internal/handler/audit_handler.go` 的 `TraceVoucher` 方法
- 新建 `internal/repository/import_batch_repo.go`
- 新建 `internal/repository/audit_log_repo.go`
- 新建 3 个 migration 文件

**上下文**：
- 项目：`/root/data/disk/huihua-finance`
- 参照：`internal/handler/invoice_handler.go` / `audit_handler.go`

**验收**：
- `go build ./...` 无报错
- `GET /api/v1/audit/exceptions` 返回正确 JSON 结构
- `GET /api/v1/audit/trace/:voucher_id` 能正确追溯到源头发票