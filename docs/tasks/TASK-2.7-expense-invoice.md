# TASK-2.7 | 进项发票模块（从零开发）

**版本**：V1.0
**优先级**：P0
**工时**：24-32h
**前置**：TASK-2.5（销售发票增强）
**状态**：待开发

---

## 任务目标

从零构建进项发票（ExpenseInvoice/InboundInvoice）模块，包含独立数据模型、Excel导入、OCR识别、验真集成，对应需求分析书V6.1第十四章。

---

## 1. 数据模型

### 1.1 新建表 expense_invoices

```sql
CREATE TABLE expense_invoices (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  company_id VARCHAR(36) NOT NULL,
  invoice_no VARCHAR(50) NOT NULL UNIQUE,  -- 发票号码（唯一）
  invoice_code VARCHAR(20) DEFAULT NULL,   -- 发票代码
  invoice_date DATE NOT NULL,              -- 开票日期
  invoice_kind VARCHAR(20) NOT NULL,       -- paper_special/paper_normal/electronic_special/electronic_normal
  tax_amount DECIMAL(18,2) NOT NULL,       -- 税额
  total_amount DECIMAL(18,2) NOT NULL,     -- 价税合计
  vendor_id VARCHAR(36) DEFAULT NULL,      -- 供应商ID
  vendor_name VARCHAR(200) DEFAULT NULL,   -- 供应商名称
  tax_id VARCHAR(50) DEFAULT NULL,         -- 供应商税号
  verify_status VARCHAR(20) DEFAULT 'unverified',  -- unverified/verified/invalid
  verified_at TIMESTAMP NULL,
  verify_result VARCHAR(200) DEFAULT NULL,
  deduction_status VARCHAR(20) DEFAULT 'undeducted', -- undeducted/deducted
  deducted_at TIMESTAMP NULL,
  source_file VARCHAR(500) DEFAULT NULL,   -- 来源文件路径
  ocr_data JSON DEFAULT NULL,               -- OCR原始数据
  status VARCHAR(20) DEFAULT 'pending',    -- pending/verified/deducted/invalid
  doc_status INT DEFAULT 0,                -- 0草稿/1提交/2审核/3过账
  remark VARCHAR(500) DEFAULT NULL,
  created_by VARCHAR(36),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant_id (tenant_id),
  INDEX idx_vendor_id (vendor_id),
  INDEX idx_invoice_date (invoice_date),
  INDEX idx_verify_status (verify_status),
  INDEX idx_deduction_status (deduction_status)
);
```

### 1.2 Model对象

`internal/model/expense_invoice.go`

```go
type ExpenseInvoice struct {
    ID              uuid.UUID
    TenantID        uuid.UUID
    CompanyID       uuid.UUID
    InvoiceNo       string
    InvoiceCode     *string
    InvoiceDate     time.Time
    InvoiceKind     string  // paper_special/paper_normal/electronic_special/electronic_normal
    TaxAmount       decimal.Decimal
    TotalAmount     decimal.Decimal
    VendorID        *uuid.UUID
    VendorName      *string
    TaxID           *string
    VerifyStatus    string  // unverified/verified/invalid
    VerifiedAt      *time.Time
    VerifyResult    *string
    DeductionStatus string  // undeducted/deducted
    DeductedAt      *time.Time
    SourceFile      *string
    OcrData         *string  // JSON
    Status          string
    DocStatus       int16
    Remark          *string
    CreatedBy       *uuid.UUID
    CreatedAt       time.Time
    UpdatedAt       *time.Time
}

// 常量
const (
    VerifyStatusUnverified = "unverified"
    VerifyStatusVerified   = "verified"
    VerifyStatusInvalid    = "invalid"
    DeductionStatusUndeducted = "undeducted"
    DeductionStatusDeducted   = "deducted"
    InvoiceKindPaperSpecial    = "paper_special"
    InvoiceKindPaperNormal     = "paper_normal"
    InvoiceKindElectronicSpecial = "electronic_special"
    InvoiceKindElectronicNormal = "electronic_normal"
)
```

### 1.3 Repository

`internal/repository/expense_invoice_repo.go`

- Create / GetByID / GetByInvoiceNo / List / Update / Delete
- FindByVendor / FindUnverified / FindUndeducted

### 1.4 Service

`internal/service/expense_invoice_service.go`

---

## 2. Excel导入流程

### 2.1 三步导入向导

```
Step1: 上传 → 解析 → 预校验
Step2: 预览（展示解析结果，错误行标红）
Step3: 确认 → 写入DB
```

### 2.2 列映射

| Excel列 | 字段 | 说明 |
|---------|------|------|
| 发票代码 | invoice_code | 可选 |
| 发票号码 | invoice_no | 必填，唯一 |
| 开票日期 | invoice_date | 必填，YYYY-MM-DD |
| 价税合计 | total_amount | 必填 |
| 税额 | tax_amount | 必填 |
| 供应商名称 | vendor_name | 可选 |

### 2.3 预校验规则

- invoice_no 判重（已存在→标记duplicate，跳过）
- 日期格式校验（非YYYY-MM-DD→标记error）
- 金额格式校验（非数字→标记error）
- 必填字段检查

### 2.4 API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/expense-invoices/import/upload` | 上传Excel文件，返回batch_id |
| GET | `/expense-invoices/import/{batch_id}/preview` | 获取预览数据 |
| POST | `/expense-invoices/import/{batch_id}/confirm` | 确认导入 |

---

## 3. OCR识别

### 3.1 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/expense-invoices/ocr` | 上传发票图片，调用OCR返回结构化数据 |

### 3.2 识别内容

- 发票号码（invoice_no）
- 金额（total_amount / tax_amount）
- 开票日期（invoice_date）
- 供应商名称（vendor_name）
- 发票类型（invoice_kind）

### 3.3 实现

复用现有的 OCR Service（参考 Python版 `services/ocr.py`），支持：
- 百度OCR
- 腾讯OCR
- 阿里云OCR

---

## 4. 验真集成

### 4.1 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/expense-invoices/{id}/verify` | 单张验真 |
| POST | `/expense-invoices/verify/batch` | 批量验真 |

### 4.2 验真逻辑

```
1. 调用国税查验平台API（或第三方聚合）
2. 传入：invoice_no + invoice_code + total_amount + invoice_date
3. 接收：verified / inconsistent / not_found / error
4. 更新：verify_status + verified_at + verify_result
```

### 4.3 异常处理

| 结果 | 处理 |
|------|------|
| verified | verify_status=verified |
| inconsistent | verify_status=invalid + verify_result="金额不一致" |
| not_found | verify_status=invalid + verify_result="查无此票" |
| error | verify_status=invalid + verify_result="系统异常" |

---

## 5. 勾选抵扣（V2预留）

### 5.1 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/expense-invoices/{id}/deduct` | 确认抵扣 |

### 5.2 抵扣逻辑

- 状态必须为 verified 才能抵扣
- 更新 deduction_status = deducted + deducted_at
- 生成抵扣凭证

---

## 6. 前端页面

### 6.1 页面清单

| 页面 | 路由 | 功能 |
|------|------|------|
| 进项发票列表 | `/expense-invoices` | 列表+导入+OCR+验真 |
| 进项发票详情 | `/expense-invoices/:id` | 详情+验真+抵扣 |
| 导入向导弹窗 | — | 上传→预览→确认 |
| OCR识别弹窗 | — | 拍照→识别→确认 |

---

## 验收标准

- [ ] expense_invoices 表创建成功
- [ ] GET/POST /expense-invoices 基础CRUD
- [ ] POST /expense-invoices/import/upload → preview → confirm 完整流程
- [ ] 预校验：invoice_no重复/日期格式错误/金额格式错误均能识别并标红
- [ ] POST /expense-invoices/ocr 返回结构化数据
- [ ] POST /expense-invoices/{id}/verify 更新verify_status
- [ ] 前端列表页支持：导入/OCR/验真/抵扣按钮
- [ ] 前端详情页展示验真结果