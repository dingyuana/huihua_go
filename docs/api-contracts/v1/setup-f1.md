# API 契约 — 基础设置 F1（账套、科目、资金账户、客商、规则）

---

## 账套

### POST /api/v1/companies

创建账套。

**Request Body:**
```json
{
  "name": "北京XX科技有限公司",
  "fiscal_year_start": "2026-01-01",
  "period_type": "monthly",
  "period_count": 12,
  "enable_date": "2026-01-01",
  "currency": "CNY",
  "init_chart_of_accounts": true
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "id": "uuid",
    "name": "北京XX科技有限公司",
    "fiscal_year_start": "2026-01-01",
    "status": "active",
    "created_at": "2026-05-27T10:00:00Z"
  }
}
```

> 账套创建时若 `init_chart_of_accounts=true`，自动导入内置《小企业会计准则》科目表。

### GET /api/v1/companies/:id

获取账套详情。

### PUT /api/v1/companies/:id

更新账套信息（期初数据变更需要主管审批，通过 `audit_logs` 记录）。

### GET /api/v1/companies

获取当前租户的账套列表。

**Query Params:** `?page=1&pageSize=20`

---

## 会计期间

### GET /api/v1/periods

获取当前账套的会计期间列表。

**Query Params:** `?year=2026&status=open`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "period_name": "2026-05",
        "start_date": "2026-05-01",
        "end_date": "2026-05-31",
        "status": "open",
        "is_locked": false
      }
    ]
  }
}
```

### POST /api/v1/periods/generate

生成会计期间（首次创建账套时调用）。

**Request Body:**
```json
{
  "company_id": "uuid",
  "fiscal_year": "2026",
  "period_type": "monthly",
  "period_count": 12
}
```

---

## 科目表

### GET /api/v1/accounts/tree

获取科目树（一次性返回所有节点，前端按 parent_id 构建树）。

**Response 200:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "uuid",
      "code": "1001",
      "name": "银行存款",
      "account_type": "asset",
      "root_type": "debit",
      "parent_id": null,
      "lft": 1,
      "rgt": 10,
      "is_group": true,
      "children": [
        {
          "id": "uuid",
          "code": "1001-01",
          "name": "银行存款-工行",
          "parent_id": "uuid",
          "lft": 2,
          "rgt": 3,
          "is_group": false,
          "currency": "CNY"
        }
      ]
    }
  ]
}
```

### GET /api/v1/accounts

分页查询科目列表。**Query Params:** `?page=1&pageSize=20&keyword=银行&account_type=asset&is_group=false`

### GET /api/v1/accounts/:id

获取单个科目详情。

### POST /api/v1/accounts

创建科目。

**Request Body:**
```json
{
  "name": "银行存款-招行",
  "parent_id": "uuid(1001)",        // 父科目 ID
  "account_type": "asset",
  "currency": "CNY",
  "is_group": false
}
```

**后端自动处理：** 生成编码（`1001-02`）、设置 `lft/rgt`、设置 `root_type`。

**Error 400:**
```json
{
  "code": 20001,
  "message": "父科目不存在或不属于当前租户"
}
```

### PUT /api/v1/accounts/:id

更新科目（`is_group` 变更时检查是否有子科目）。

### DELETE /api/v1/accounts/:id

删除科目。**规则：** 仅允许删除无子科目、无业务引用的科目。

### GET /api/v1/accounts/ledger-only

获取仅可记账的科目列表（`is_group=false`），供凭证分录选择。

**Response 200:**
```json
{
  "code": 0,
  "data": [
    { "id": "uuid", "code": "1001-01", "name": "银行存款-工行", "account_type": "asset" }
  ]
}
```

### POST /api/v1/accounts/auto-code

预览自动生成的科目编码。

**Request Body:**
```json
{ "parent_id": "uuid(1001)" }
```

**Response:**
```json
{ "code": 0, "data": { "suggested_code": "1001-02" } }
```

---

## 资金账户

### GET /api/v1/bank-accounts

获取银行账户列表。

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "bank_name": "中国工商银行",
        "account_number": "110202121900...",
        "clearing_account": { "id": "uuid", "code": "1001-01", "name": "银行存款-工行" },
        "currency": "CNY",
        "bank_account_type": "current",
        "is_active": true
      }
    ]
  }
}
```

### POST /api/v1/bank-accounts

创建银行账户。

**Request Body:**
```json
{
  "bank_name": "中国工商银行",
  "account_number": "110202121900...",
  "clearing_account_id": "uuid",    // 关联的 GL 科目 ID
  "company_id": "uuid",
  "currency": "CNY",
  "bank_account_type": "current"
}
```

> **校验：** `clearing_account_id` 必须是 `account_type=asset` 且 `is_group=false`，创建后不可修改。

### PUT /api/v1/bank-accounts/:id

更新银行账户（`clearing_account_id` 不可更新）。

### DELETE /api/v1/bank-accounts/:id

停用银行账户（软删除，设置 `is_active=false`）。

### GET /api/v1/bank-accounts/:id/balance

获取银行账户当前余额。

**Response:**
```json
{
  "code": 0,
  "data": {
    "bank_account_id": "uuid",
    "account_number": "11020212...",
    "current_balance": "1250000.00",
    "balance_date": "2026-05-27"
  }
}
```

---

## 客商档案

### GET /api/v1/parties

查询客商列表。

**Query Params:** `?page=1&pageSize=20&keyword=&party_type=customer&is_active=true`

### POST /api/v1/parties

创建客商。

**Request Body:**
```json
{
  "name": "北京XX科技有限公司",
  "tax_id": "91110108MA...",
  "party_type": "customer",
  "bank_name": "中国银行",
  "bank_account": "345678901234567",
  "credit_limit": "500000.00",
  "payment_terms": "net30",
  "phone": "010-88886666",
  "address": "北京市海淀区..."
}
```

### PUT /api/v1/parties/:id

更新客商。

### DELETE /api/v1/parties/:id

删除客商（有业务单据引用时拒绝删除）。

### POST /api/v1/parties/import

Excel 批量导入客商。

**Request:** `multipart/form-data`，文件字段 `file`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "total": 100,
    "imported": 98,
    "failed": 2,
    "errors": [
      { "row": 5, "message": "税号格式不正确" },
      { "row": 23, "message": "已存在的客商（税号重复）" }
    ]
  }
}
```

### GET /api/v1/parties/search

客商快速搜索（供其他模块的选择器使用）。

**Query Params:** `?keyword=北京&party_type=customer&limit=10`

**Response 200:**
```json
{
  "code": 0,
  "data": [
    { "id": "uuid", "name": "北京XX科技", "tax_id": "91110...", "party_type": "customer" }
  ]
}
```

---

## 智能分类规则库

### GET /api/v1/classification-rules

获取规则列表。

**Query Params:** `?page=1&pageSize=20&keyword=手续费&rule_type=keyword`

### POST /api/v1/classification-rules

创建分类规则。

**Request Body:**
```json
{
  "name": "银行手续费匹配",
  "rule_type": "keyword_regex",
  "pattern": "手续费|管理费|年费",
  "match_field": "description",
  "direction": "out",
  "classification": "bank_fee",
  "priority": 1,
  "is_active": true
}
```

**优先级说明：** 数字越小优先级越高，命中后不再匹配后续规则。

### PUT /api/v1/classification-rules/:id

更新规则。

### DELETE /api/v1/classification-rules/:id

删除规则。

### POST /api/v1/classification-rules/reorder

调整规则优先级。

**Request Body:**
```json
{
  "rule_ids": ["uuid1", "uuid2", "uuid3"]
}
```

---

## 科目映射规则

### GET /api/v1/mapping-rules

获取单据类型→科目的映射规则列表。

### PUT /api/v1/mapping-rules/:id

更新映射规则。

**Request Body:**
```json
{
  "doc_type": "business_receipt",
  "debit_account_id": "uuid(银行存款)",
  "credit_account_id": "uuid(应收账款)",
  "debit_account_alt": "uuid(现金)",
  "credit_account_alt": null,
  "condition_expr": "amount > 5000"
}
```
