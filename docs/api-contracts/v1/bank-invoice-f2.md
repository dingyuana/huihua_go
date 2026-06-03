# API 契约 — 票据采集 F2（银行流水导入 + 发票 OCR）

---

## 银行流水导入

### POST /api/v1/bank-transactions/import

上传银行对账单文件并解析。

**Request:** `multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|:---|:---|:---:|:---|
| file | File | ✅ | CSV/Excel/XML |
| bank_account_id | UUID | ✅ | 目标银行账户 |
| format | String | ❌ | 自动检测，可选覆盖: csv/excel/camt053/mt940 |
| template_id | UUID | ❌ | 字段映射模板 ID |

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "batch_id": "uuid",
    "bank_account_id": "uuid",
    "total": 150,
    "imported": 145,
    "duplicated": 3,
    "failed": 2,
    "errors": [
      { "row": 12, "message": "金额格式无法解析" },
      { "row": 88, "message": "日期字段为空" }
    ],
    "preview": [
      {
        "txn_date": "2026-05-20",
        "description": "网银转账-张-B2B收款",
        "amount": "12000.00",
        "direction": "in",
        "counterparty_name": "上海XX贸易公司",
        "reference_no": "B20260520001"
      }
    ]
  }
}
```

### POST /api/v1/bank-transactions/classify

对导入后的流水执行智能分类。

**Request Body:**
```json
{
  "batch_id": "uuid"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "batch_id": "uuid",
    "classifications": {
      "business_receipt": 45,
      "business_payment": 60,
      "bank_fee": 12,
      "interest_income": 3,
      "internal_transfer": 15,
      "pending": 10
    },
    "confidence_avg": 0.87,
    "pending_count": 10
  }
}
```

### GET /api/v1/bank-transactions

查询银行流水列表。

**Query Params:** `?batch_id=uuid&bank_account_id=uuid&direction=in&classification=pending&matched=false&start_date=2026-05-01&end_date=2026-05-31&page=1&pageSize=20`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "txn_date": "2026-05-20",
        "description": "网银转账-张-B2B收款",
        "debit": "12000.00",
        "credit": "0.00",
        "direction": "in",
        "counterparty_name": "上海XX贸易公司",
        "reference_no": "B20260520001",
        "classification": "business_receipt",
        "matched": false,
        "is_duplicate": false,
        "imported_from": "excel"
      }
    ],
    "total": 145,
    "page": 1,
    "pageSize": 20
  }
}
```

### PUT /api/v1/bank-transactions/:id/classify

手动修正单条流水的分类。

**Request Body:**
```json
{
  "classification": "business_payment",
  "counterparty_id": "uuid",
  "remark": "修正为付款-供应商XX"
}
```

### POST /api/v1/bank-transactions/batch-classify

批量修正分类。

**Request Body:**
```json
{
  "ids": ["uuid1", "uuid2"],
  "classification": "business_payment"
}
```

### POST /api/v1/bank-transactions/:id/confirm

确认单条流水（分类确认后自动生成对应单据草稿）。

### POST /api/v1/bank-transactions/batch-confirm

批量确认流水。

**Request Body:**
```json
{
  "ids": ["uuid1", "uuid2", "uuid3"]
}
```

### GET /api/v1/bank-transactions/import-logs

查看导入历史。

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "batch_id": "uuid",
        "bank_account_name": "工行-基本户",
        "file_name": "202605_工行对账单.xlsx",
        "imported_at": "2026-05-27T10:00:00Z",
        "total": 150,
        "imported": 145,
        "duplicated": 3,
        "failed": 2,
        "imported_by": "张三"
      }
    ]
  }
}
```

---

## 字段映射模板

### GET /api/v1/field-mapping-templates

获取预设的银行字段映射模板列表。

### POST /api/v1/field-mapping-templates

保存自定义字段映射。

**Request Body:**
```json
{
  "bank_name": "中国银行",
  "format": "excel",
  "column_mapping": {
    "date_col": "交易日期",
    "debit_col": "收入金额",
    "credit_col": "支出金额",
    "counterparty_col": "对方户名",
    "description_col": "摘要",
    "reference_col": "流水号",
    "balance_col": "余额"
  },
  "header_rows": 1,
  "encoding": "UTF-8"
}
```

---

## 发票

### POST /api/v1/invoices/upload

上传发票文件（支持图片、PDF、OFD）。

**Request:** `multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|:---|:---|:---:|:---|
| file | File | ✅ | 发票文件 |
| party_id | UUID | ❌ | 预先指定对方单位 |

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "invoice_no": "12345678",
    "invoice_type": "sale",
    "posting_date": "2026-05-20",
    "total_amount": "113000.00",
    "tax_amount": "13000.00",
    "net_amount": "100000.00",
    "customer_name": "上海XX贸易公司",
    "customer_tax_id": "91310000MA...",
    "confidence": 0.96,
    "status": "pending_review",
    "field_errors": []
  }
}
```

### GET /api/v1/invoices

发票列表查询。

**Query Params:** `?status=unpaid&invoice_type=purchase&start_date=2026-01-01&end_date=2026-05-31&customer_id=uuid&keyword=发票号&page=1&pageSize=20`

### GET /api/v1/invoices/:id

发票详情（含 OCR 识别结果原始数据）。

### PUT /api/v1/invoices/:id

手动修正发票字段（置信度低时的补充）。

### DELETE /api/v1/invoices/:id

删除发票（仅限 docstatus=0）。

### POST /api/v1/invoices/batch-import

批量导入发票（电子税务局一键取票）。

**Request Body:**
```json
{
  "source": "e-tax",
  "invoices": [
    {
      "invoice_no": "12345678",
      "invoice_type": "purchase",
      "posting_date": "2026-05-20",
      "total_amount": "113000.00",
      "tax_amount": "13000.00",
      "net_amount": "100000.00",
      "supplier_name": "上海XX贸易公司",
      "supplier_tax_id": "91310000MA..."
    }
  ]
}
```

### GET /api/v1/invoices/expiring

获取即将过期的进项发票（默认提前 30 天提醒）。

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "invoice_no": "12345678",
        "supplier_name": "上海XX",
        "total_amount": "50000.00",
        "due_date": "2026-06-15",
        "days_remaining": 19
      }
    ],
    "total_expiring": 5
  }
}
```
