# API 契约 — 凭证管理 F5（模板 + 状态机 + 自动生成 + 审核）

---

## 凭证模板

### GET /api/v1/voucher-templates

获取凭证模板列表。

### POST /api/v1/voucher-templates

创建凭证模板。

**Request Body:**
```json
{
  "title": "收款-银行存款",
  "doc_type": "business_receipt",
  "lines": [
    {
      "account_id": "uuid(银行存款)",
      "debit_or_credit": "debit",
      "is_fixed": true
    },
    {
      "account_id": "uuid(应收账款)",
      "debit_or_credit": "credit",
      "is_fixed": false
    }
  ],
  "is_active": true
}
```

### PUT /api/v1/voucher-templates/:id

更新模板。

---

## 状态机操作

### POST /api/v1/vouchers

创建凭证草稿。

**Request Body:**
```json
{
  "voucher_type": "记",
  "posting_date": "2026-05-27",
  "remark": "收款-上海XX贸易公司",
  "lines": [
    {
      "account_id": "uuid(1001-01)",
      "debit": "12000.00",
      "credit": "0.00",
      "party_type": "customer",
      "party_id": "uuid",
      "user_remark": "工行收款"
    },
    {
      "account_id": "uuid(1122)",
      "debit": "0.00",
      "credit": "12000.00",
      "party_type": "customer",
      "party_id": "uuid"
    }
  ]
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "voucher_no": null,
    "docstatus": 0,
    "total_debit": "12000.00",
    "total_credit": "12000.00",
    "balanced": true,
    "warnings": []
  }
}
```

### POST /api/v1/vouchers/:id/submit

提交审核（docstatus: 0 → 1）。

- 校验借贷平衡
- 生成凭证编号 `记-YYYY-MM-NNNN`
- 生成 GL Entry（双写）
- 更新 docstatus=1
- 记录 audit_log

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "voucher_no": "记-2026-05-0001",
    "docstatus": 1,
    "submitted_by": "uuid",
    "submitted_at": "2026-05-27T10:30:00Z",
    "gl_entries_generated": 2
  }
}
```

**Error 400:**
```json
{
  "code": 20001,
  "message": "借贷不平衡: 借方 12000.00 != 贷方 10000.00，差额 2000.00"
}
```

### POST /api/v1/vouchers/:id/cancel

作废凭证（docstatus: 1 → 2）。

- 生成反向 GL Entry
- 更新 docstatus=2
- 记录 audit_log

**Request Body:**
```json
{
  "reason": "凭证金额错误，需重新编制"
}
```

### POST /api/v1/vouchers/:id/reverse

红字冲销（生成新的反向凭证）。

**Request Body:**
```json
{
  "reverse_date": "2026-05-28",
  "reason": "科目用错，红字冲销后重新编制"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "original_voucher_no": "记-2026-05-0001",
    "reverse_voucher_no": "记-2026-05-0002",
    "reverse_voucher_id": "uuid",
    "docstatus": 0
  }
}
```

---

## 凭证查询

### GET /api/v1/vouchers

凭证列表查询。

**Query Params:** `?voucher_type=记&docstatus=1&start_date=2026-05-01&end_date=2026-05-31&keyword=上海&page=1&pageSize=20&sort=posting_date:desc`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "voucher_no": "记-2026-05-0001",
        "voucher_type": "记",
        "posting_date": "2026-05-27",
        "remark": "收款-上海XX贸易公司",
        "total_debit": "12000.00",
        "total_credit": "12000.00",
        "docstatus": 1,
        "submitted_by_name": "张三",
        "created_at": "2026-05-27T10:00:00Z"
      }
    ],
    "total": 350,
    "page": 1,
    "pageSize": 20
  }
}
```

### GET /api/v1/vouchers/:id

凭证详情（含分录行）。

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "voucher_no": "记-2026-05-0001",
    "voucher_type": "记",
    "posting_date": "2026-05-27",
    "remark": "收款-上海XX贸易公司",
    "docstatus": 1,
    "submitted_by_name": "张三",
    "submitted_at": "2026-05-27T10:30:00Z",
    "lines": [
      {
        "account_code": "1001-01",
        "account_name": "银行存款-工行",
        "debit": "12000.00",
        "credit": "0.00",
        "party_name": "上海XX贸易公司",
        "user_remark": "工行收款"
      },
      {
        "account_code": "1122",
        "account_name": "应收账款",
        "debit": "0.00",
        "credit": "12000.00",
        "party_name": "上海XX贸易公司"
      }
    ],
    "gl_entries": [
      {
        "account_code": "1001-01",
        "debit": "12000.00",
        "credit": "0.00",
        "posting_date": "2026-05-27"
      }
    ]
  }
}
```

---

## 审核工作台

### GET /api/v1/vouchers/pending-review

获取待审核凭证列表。

**Query Params:** `?page=1&pageSize=20&sort=created_at:asc`

### POST /api/v1/vouchers/batch-review

批量审核/驳回。

**Request Body:**
```json
{
  "ids": ["uuid1", "uuid2"],
  "action": "approve",
  "remark": "审核通过"
}
```

### POST /api/v1/vouchers/batch-reject

批量驳回。

**Request Body:**
```json
{
  "ids": ["uuid1"],
  "reason": "摘要不清晰，需补充对方单位信息"
}
```

### POST /api/v1/vouchers/batch-generate

批量生成凭证（从已完成核销的单据生成凭证）。

**Request Body:**
```json
{
  "payment_ids": ["uuid1", "uuid2"],
  "doc_type": "business_receipt"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "total": 10,
    "succeeded": 8,
    "failed": 2,
    "failed_details": [
      { "payment_id": "uuid", "error": "科目映射规则未配置" }
    ],
    "voucher_ids": ["uuid1", "uuid2", ...]
  }
}
```

---

## 凭证编号

### GET /api/v1/vouchers/next-no?date=2026-05-27

预览下一个凭证编号。

**Response:**
```json
{
  "code": 0,
  "data": {
    "next_voucher_no": "记-2026-05-0002"
  }
}
```

### GET /api/v1/vouchers/gaps?year=2026&month=5

检查凭证编号断号。

**Response:**
```json
{
  "code": 0,
  "data": {
    "gaps": [
      { "expected": "记-2026-05-0003", "status": "cancelled" },
      { "expected": "记-2026-05-0007", "status": "missing" }
    ],
    "total_gaps": 2
  }
}
```
