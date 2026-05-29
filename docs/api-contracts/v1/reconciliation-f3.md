# API 契约 — 核销引擎 F3

---

## 核销预检

### POST /api/v1/reconciliation/precheck

执行核销前的 6 项预检。

**Request Body:**
```json
{
  "invoice_id": "uuid",
  "payment_id": "uuid"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "invoice_id": "uuid",
    "payment_id": "uuid",
    "passed": false,
    "overall_message": "6 项检查中 1 项未通过",
    "checks": [
      {
        "id": 1,
        "name": "对方单位匹配",
        "status": "passed",
        "message": "上海XX贸易公司 ↔ 发票购方税号一致",
        "severity": "info"
      },
      {
        "id": 2,
        "name": "金额超限检查",
        "status": "passed",
        "message": "核销金额 ¥12,000 ≤ 未结清金额 ¥15,000",
        "severity": "info"
      },
      {
        "id": 3,
        "name": "重复核销检查",
        "status": "passed",
        "message": "该发票尚未被此收款单核销",
        "severity": "info"
      },
      {
        "id": 4,
        "name": "跨账套检查",
        "status": "passed",
        "message": "核销双方属于同一租户",
        "severity": "info"
      },
      {
        "id": 5,
        "name": "业务类型一致性",
        "status": "passed",
        "message": "收款单匹配销项发票",
        "severity": "info"
      },
      {
        "id": 6,
        "name": "到期日检查",
        "status": "blocked",
        "message": "进项发票已过期（到期日 2026-04-20），请确认是否参与核销",
        "severity": "error"
      }
    ],
    "can_force_pass": true
  }
}
```

> 预检不通过时，会计可选择 `force_pass=true` 并备注原因强制通过。

### POST /api/v1/reconciliation/precheck/force-pass

强制通过预检并执行核销。

**Request Body:**
```json
{
  "invoice_id": "uuid",
  "payment_id": "uuid",
  "allocated_amount": "12000.00",
  "reason": "客户急需回款，已电话确认发票有效",
  "check_ids_to_override": [6]
}
```

---

## 智能匹配

### POST /api/v1/reconciliation/match

执行五级匹配，返回推荐结果。

**Request Body:**
```json
{
  "payment_id": "uuid",
  "strategy": "auto"         // auto / manual
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "recommendations": [
      {
        "level": "L1_exact",
        "invoice_id": "uuid",
        "invoice_no": "12345678",
        "invoice_date": "2026-05-10",
        "allocated_amount": "12000.00",
        "outstanding_before": "12000.00",
        "outstanding_after": "0.00",
        "score": 100,
        "auto_execute": true,
        "need_confirm": false
      },
      {
        "level": "L2_fifo",
        "invoice_id": "uuid2",
        "invoice_no": "87654321",
        "allocated_amount": "5000.00",
        "outstanding_before": "8000.00",
        "outstanding_after": "3000.00",
        "auto_execute": true,
        "need_confirm": false
      },
      {
        "level": "L3_fuzzy",
        "invoice_id": "uuid3",
        "invoice_no": "11223344",
        "score": 82.5,
        "customer_name_similarity": "上海XX(88%)",
        "auto_execute": false,
        "need_confirm": true
      }
    ]
  }
}
```

### POST /api/v1/reconciliation/match/confirm

确认匹配推荐（L3 及以上需要用户确认）。

**Request Body:**
```json
{
  "payment_id": "uuid",
  "allocations": [
    { "invoice_id": "uuid3", "allocated_amount": "3000.00" }
  ]
}
```

---

## 手工核销

### GET /api/v1/reconciliation/candidates

获取可核销的候选列表（左侧发票、右侧收付款）。

**Query Params:** `?payment_id=uuid&party_id=uuid&page=1&pageSize=20`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "payment": {
      "id": "uuid",
      "payment_no": "SK-2026-05-0001",
      "party_name": "上海XX贸易公司",
      "paid_amount": "12000.00",
      "unallocated_amount": "5000.00"
    },
    "candidate_invoices": [
      {
        "id": "uuid",
        "invoice_no": "12345678",
        "posting_date": "2026-05-10",
        "total_amount": "15000.00",
        "outstanding_amount": "3000.00",
        "allocatable": true
      }
    ]
  }
}
```

### POST /api/v1/reconciliation/manual-allocate

执行手工核销。

**Request Body:**
```json
{
  "payment_id": "uuid",
  "allocations": [
    { "invoice_id": "uuid", "allocated_amount": "3000.00" }
  ]
}
```

---

## 核销执行 & 回退

### POST /api/v1/reconciliation/execute

执行核销（更新发票状态 + 登记 payment_allocations）。

**Request Body:**
```json
{
  "payment_id": "uuid",
  "allocations": [
    { "invoice_id": "uuid", "allocated_amount": "12000.00" }
  ]
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "payment_id": "uuid",
    "allocations": [
      {
        "id": "uuid",
        "invoice_id": "uuid",
        "allocated_amount": "12000.00",
        "invoice_outstanding_before": "12000.00",
        "invoice_outstanding_after": "0.00",
        "invoice_status": "paid"
      }
    ],
    "audit_log_id": "uuid"
  }
}
```

### POST /api/v1/reconciliation/:id/reverse

核销回退（核销后 30 天内可操作）。

**Request Body:**
```json
{
  "reason": "发票退回重开"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "reversed_allocation_id": "uuid",
    "invoice_new_outstanding": "12000.00",
    "invoice_new_status": "unpaid"
  }
}
```

---

## 核销记录查询

### GET /api/v1/reconciliation/allocations

**Query Params:** `?invoice_id=uuid&payment_id=uuid&page=1&pageSize=20`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": "uuid",
        "payment_no": "SK-2026-05-0001",
        "invoice_no": "12345678",
        "allocated_amount": "12000.00",
        "created_at": "2026-05-27T10:30:00Z",
        "cancelled": false
      }
    ]
  }
}
```
