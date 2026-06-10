# SPEC: 核销模块 — 资金流水与债权债务匹配

## 背景

慧智财务 Go 版已实现收款单/付款单、应收单/应付单的基础管理功能，但缺少核心的核销能力：
1. **资金流与债权债务无关联**：收款单/付款单与应收单/应付单之间没有核销关系
2. **手工核销缺失**：无法手动配对单据进行核销
3. **自动核销规则缺失**：无智能匹配机制，需人工逐单处理
4. **反核销能力缺失**：核销错误后无法撤销恢复

## 设计目标

实现完整的核销模块，支持：
1. 自动核销：按配置规则智能匹配收/付款单与应收/应付单
2. 手工核销：提供界面支持灵活配对和金额指定
3. 反核销：支持撤销已核销记录，恢复单据状态
4. 容差处理：支持金额容差范围内的自动核销
5. 完整追溯：所有操作记录日志，支持审计

---

## 一、核心数据模型

### 1.1 核销记录表 (write_off_record)

```sql
CREATE TABLE write_off_records (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    write_off_no varchar(64) UNIQUE NOT NULL,
    type varchar(32) NOT NULL,  -- payment_ar / payment_ap
    receipt_payment_id uuid NOT NULL REFERENCES payment_entries(id),
    receivable_payable_id uuid NOT NULL REFERENCES ar_invoices(id),
    amount decimal(18,4) NOT NULL,
    write_off_date date NOT NULL,
    operator uuid REFERENCES users(id),
    status smallint NOT NULL DEFAULT 1,  -- 1=正常, 0=已反核销
    remark text,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
```

### 1.2 单据余额字段扩展

在 `payment_entries`、`ar_invoices`、`ap_invoices` 表中增加：

```sql
ALTER TABLE payment_entries ADD COLUMN total_amount decimal(18,4);
ALTER TABLE payment_entries ADD COLUMN write_off_amount decimal(18,4) DEFAULT 0;
ALTER TABLE payment_entries ADD COLUMN remaining_amount decimal(18,4);
ALTER TABLE payment_entries ADD COLUMN write_off_status varchar(32) DEFAULT 'unwritten';  -- unwritten / partial / written

ALTER TABLE ar_invoices ADD COLUMN write_off_amount decimal(18,4) DEFAULT 0;
ALTER TABLE ar_invoices ADD COLUMN remaining_amount decimal(18,4);
ALTER TABLE ar_invoices ADD COLUMN write_off_status varchar(32) DEFAULT 'unwritten';

ALTER TABLE ap_invoices ADD COLUMN write_off_amount decimal(18,4) DEFAULT 0;
ALTER TABLE ap_invoices ADD COLUMN remaining_amount decimal(18,4);
ALTER TABLE ap_invoices ADD COLUMN write_off_status varchar(32) DEFAULT 'unwritten';
```

### 1.3 核销规则配置表 (write_off_rules)

```sql
CREATE TABLE write_off_rules (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tenant_id uuid NOT NULL REFERENCES tenants(id),
    name varchar(64) NOT NULL,
    enabled boolean DEFAULT true,
    match_priority jsonb,  -- ["counterparty", "amount", "document_no", "date"]
    tolerance_type varchar(16),  -- absolute / relative
    tolerance_value decimal(18,4),
    date_threshold_days int DEFAULT 30,
    cross_period_confirm boolean DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
```

---

## 二、自动核销规则引擎

### 2.1 匹配条件优先级

| 优先级 | 匹配维度 | 规则说明 |
|-------|---------|---------|
| 1 | 往来单位 | 收款单付款方 = 应收单客户；付款单收款方 = 应付单供应商 |
| 2 | 金额相等 | 单据金额在容差范围内相等 |
| 3 | 单据编号关联 | 收/付款单摘要包含应收/应付单号，或反之 |
| 4 | 日期接近 | 单据日期之差 ≤ 设定天数（默认30天） |

### 2.2 容差处理

```go
type ToleranceConfig struct {
    Type  string  // "absolute" | "relative"
    Value float64 // 绝对值或百分比
}

func (t *ToleranceConfig) IsWithinTolerance(actual, expected decimal.Decimal) bool {
    diff := actual.Sub(expected).Abs()
    if t.Type == "absolute" {
        return diff.LessThanOrEqual(decimal.NewFromFloat(t.Value))
    }
    // relative: diff / expected <= tolerance
    ratio := diff.Div(expected).Mul(decimal.NewFromFloat(100))
    return ratio.LessThanOrEqual(decimal.NewFromFloat(t.Value))
}
```

### 2.3 多笔匹配逻辑

```go
type WriteOffEngine struct {
    ruleRepo        WriteOffRuleRepository
    paymentRepo     PaymentEntryRepository
    arInvoiceRepo   ArInvoiceRepository
    apInvoiceRepo   ApInvoiceRepository
}

// 一对多核销：一笔收款匹配多张应收
func (e *WriteOffEngine) MatchOneToMany(
    ctx context.Context,
    payment *model.PaymentEntry,
    invoices []*model.ArInvoice,
) ([]*model.WriteOffRecord, error) {
    // 按金额排序，优先匹配金额大的
    sort.Slice(invoices, func(i, j int) bool {
        return invoices[i].RemainingAmount.GreaterThan(invoices[j].RemainingAmount)
    })
    
    remaining := payment.RemainingAmount
    var records []*model.WriteOffRecord
    
    for _, inv := range invoices {
        if remaining.IsZero() {
            break
        }
        amount := decimal.Min(remaining, inv.RemainingAmount)
        records = append(records, &model.WriteOffRecord{
            Type:                 "payment_ar",
            ReceiptPaymentID:     payment.ID,
            ReceivablePayableID:  inv.ID,
            Amount:               amount,
        })
        remaining = remaining.Sub(amount)
    }
    
    // 尾差转为预收
    if !remaining.IsZero() {
        // 创建预收记录...
    }
    
    return records, nil
}
```

---

## 三、核心服务接口

### 3.1 WriteOffService 接口

```go
type WriteOffService interface {
    // 自动核销
    AutoWriteOff(ctx context.Context, tenantID uuid.UUID, opts AutoWriteOffOptions) (*WriteOffResult, error)
    
    // 手工核销
    ManualWriteOff(ctx context.Context, tenantID, operatorID uuid.UUID, req ManualWriteOffRequest) (*model.WriteOffRecord, error)
    
    // 反核销
    ReverseWriteOff(ctx context.Context, tenantID, operatorID, recordID uuid.UUID) error
    
    // 查询核销记录
    GetWriteOffRecords(ctx context.Context, tenantID uuid.UUID, params QueryParams) ([]*model.WriteOffRecord, error)
    
    // 查询未核销汇总
    GetUnmatchedSummary(ctx context.Context, tenantID uuid.UUID) (*UnmatchedSummary, error)
}

type AutoWriteOffOptions struct {
    DocumentType string    // "payment_ar" | "payment_ap"
    StartDate    time.Time
    EndDate      time.Time
    CounterpartyID uuid.UUID
}

type ManualWriteOffRequest struct {
    ReceiptPaymentID    uuid.UUID
    ReceivablePayableID uuid.UUID
    Amount              decimal.Decimal
    Remark              string
}

type WriteOffResult struct {
    TotalMatched      int
    TotalAmount       decimal.Decimal
    FailedCount       int
    UnmatchedDocuments []UnmatchedDocument
}
```

### 3.2 WriteOffRepository 接口

```go
type WriteOffRepository interface {
    Create(ctx context.Context, record *model.WriteOffRecord) error
    Update(ctx context.Context, record *model.WriteOffRecord) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetByID(ctx context.Context, tenantID, id uuid.UUID) (*model.WriteOffRecord, error)
    List(ctx context.Context, tenantID uuid.UUID, params QueryParams) ([]*model.WriteOffRecord, error)
    BatchCreate(ctx context.Context, records []*model.WriteOffRecord) error
}
```

---

## 四、API 接口设计

### 4.1 触发自动核销

- **方法**：POST
- **路径**：`/api/writeoff/auto`
- **Body**：
```json
{
  "document_type": "payment_ar",
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "counterparty_id": "uuid"
}
```
- **返回**：
```json
{
  "total_matched": 10,
  "total_amount": 10000.00,
  "failed_count": 2,
  "unmatched_documents": []
}
```

### 4.2 手工核销

- **方法**：POST
- **路径**：`/api/writeoff/manual`
- **Body**：
```json
{
  "receipt_payment_id": "uuid",
  "receivable_payable_id": "uuid",
  "amount": 500.00,
  "remark": "手工核销"
}
```

### 4.3 反核销

- **方法**：POST
- **路径**：`/api/writeoff/reverse/{record_id}`

### 4.4 查询核销记录

- **方法**：GET
- **路径**：`/api/writeoff/records`
- **参数**：tenant_id, start_date, end_date, counterparty_id, page, page_size

### 4.5 查询未核销汇总

- **方法**：GET
- **路径**：`/api/writeoff/unmatched-summary`
- **参数**：tenant_id, counterparty_id (可选)
- **返回**：
```json
{
  "total_unmatched_amount": 100000.00,
  "overdue_amount": 20000.00,
  "by_counterparty": [
    {"counterparty_id": "uuid", "name": "客户A", "amount": 50000.00}
  ]
}
```

---

## 五、异常处理机制

### 5.1 无法匹配
- **触发**：自动核销后仍未匹配的单据
- **处理**：归入待处理队列，生成待处理任务

### 5.2 金额冲突
- **触发**：收款单金额远大于所有未核销应收单总额
- **处理**：提示"超出金额将转为预收"，生成预收记录

### 5.3 跨期太长
- **触发**：收/付款单与应收/应付单日期相差超过180天
- **处理**：自动核销需经二次确认

---

## 六、涉及的文件变更

| 文件 | 变更类型 | 内容 |
|------|---------|------|
| `internal/model/write_off_record.go` | 新增 | 核销记录模型 |
| `internal/model/payment_entry.go` | 修改 | 新增核销相关字段 |
| `internal/model/ar_invoice.go` | 修改 | 新增核销相关字段 |
| `internal/model/ap_invoice.go` | 修改 | 新增核销相关字段 |
| `internal/repository/write_off_repo.go` | 新增 | 核销记录仓储 |
| `internal/service/write_off_service.go` | 新增 | 核销服务 |
| `internal/handler/write_off_handler.go` | 新增 | 核销 HTTP 接口 |
| `internal/service/write_off_engine.go` | 新增 | 自动核销引擎 |
| `db/migrations/` | 新增 | 表结构变更 Migration |

---

## 七、验证标准

| 序号 | 验证项 | 方式 |
|-----|-------|------|
| 1 | 自动核销正确匹配"单位相同+金额相等"的单据 | 测试用例 |
| 2 | 容差范围内的单据可自动核销 | 测试用例 |
| 3 | 手工核销可自由组合单据并指定金额 | 功能测试 |
| 4 | 反核销后单据金额恢复可重新核销 | 测试用例 |
| 5 | 支持一对多、多对一、部分核销 | 测试用例 |
| 6 | 核销历史可导出，日志可查 | 功能测试 |
| 7 | 1万张单据核销耗时 ≤ 3秒 | 性能测试 |
| 8 | 并发核销无冲突 | 并发测试 |