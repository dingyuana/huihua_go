# API 契约 — 银企对账 F4 + 期末处理 F6 + 经营分析 F7

---

## 银企对账 (F4)

### POST /api/v1/bank-reconciliation/run

执行对账匹配（按多维打分算法）。

**Request Body:**
```json
{
  "bank_account_id": "uuid",
  "period_start": "2026-05-01",
  "period_end": "2026-05-31"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "bank_account_id": "uuid",
    "period": "2026-05",
    "total_bank_txns": 150,
    "total_gl_entries": 145,
    "auto_matched": 130,
    "need_confirm": 12,
    "unmatched": 8,
    "auto_match_rate": 89.66,
    "matches": [
      {
        "bank_txn_id": "uuid",
        "gl_entry_id": "uuid",
        "score": 92.5,
        "is_auto_matched": true,
        "dimensions": {
          "amount_score": 50,
          "date_score": 20,
          "name_score": 15,
          "desc_score": 7.5,
          "ref_no_score": 0
        }
      }
    ]
  }
}
```

### GET /api/v1/bank-reconciliation/matches

查看对账匹配结果。

**Query Params:** `?bank_account_id=uuid&period=2026-05&match_status=auto_matched&page=1&pageSize=20`

### POST /api/v1/bank-reconciliation/confirm

确认待确认的匹配项。

**Request Body:**
```json
{
  "match_ids": ["uuid1", "uuid2"],
  "confirmed": true
}
```

### GET /api/v1/bank-reconciliation/balance-sheet

生成余额调节表。

**Query Params:** `?bank_account_id=uuid&period=2026-05`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "bank_account": "工行-基本户(11020212...)",
    "statement_date": "2026-05-31",
    "bank_statement_balance": "1250000.00",
    "gl_balance": "1245000.00",
    "difference": "5000.00",
    "bank_only_items": {
      "receipts": [
        { "txn_id": "uuid", "date": "2026-05-30", "amount": "5000.00", "description": "银行利息", "type": "bank_receipt_not_in_gl" }
      ],
      "payments": []
    },
    "gl_only_items": {
      "receipts": [],
      "payments": [
        { "voucher_no": "记-2026-05-0012", "date": "2026-05-28", "amount": "2000.00", "description": "在途付款", "type": "gl_payment_not_in_bank" }
      ]
    },
    "adjusted_balances": {
      "bank_balance_adjusted": "1247000.00",
      "gl_balance_adjusted": "1247000.00",
      "balanced": true
    }
  }
}
```

### POST /api/v1/bank-reconciliation/lock

锁定对账结果。

**Request Body:**
```json
{
  "bank_account_id": "uuid",
  "period": "2026-05"
}
```

### POST /api/v1/bank-reconciliation/unlock

解锁对账结果（需主管审批）。

**Request Body:**
```json
{
  "bank_account_id": "uuid",
  "period": "2026-05",
  "reason": "发现一笔流水遗漏，需重新对账"
}
```

---

## 资金日记账

### GET /api/v1/cash-journal

查询现金日记账。

**Query Params:** `?bank_account_id=uuid&start_date=2026-05-01&end_date=2026-05-31&page=1&pageSize=20`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "account_name": "库存现金",
    "list": [
      {
        "date": "2026-05-20",
        "voucher_no": "记-2026-05-0020",
        "description": "提现",
        "debit": "5000.00",
        "credit": "0.00",
        "balance": "15000.00"
      },
      {
        "date": "2026-05-21",
        "voucher_no": "记-2026-05-0025",
        "description": "差旅费报销",
        "debit": "0.00",
        "credit": "3000.00",
        "balance": "12000.00"
      }
    ]
  }
}
```

### POST /api/v1/cash-journal/manual-entry

手工补录现金日记账条目。

**Request Body:**
```json
{
  "entry_date": "2026-05-27",
  "description": "现金盘点差异调整",
  "debit": "100.00",
  "credit": "0.00",
  "remark": "实盘比账面多100元，原因待查"
}
```

### POST /api/v1/cash-journal/count

录入现金盘点。

**Request Body:**
```json
{
  "count_date": "2026-05-31",
  "actual_cash": "12000.00"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "book_balance": "12100.00",
    "actual_cash": "12000.00",
    "difference": "-100.00",
    "within_tolerance": true,
    "needs_approval": false
  }
}
```

---

## 结账体检 (F6)

### POST /api/v1/periods/health-check

运行结账前 10 项体检。

**Request Body:**
```json
{
  "period": "2026-05",
  "company_id": "uuid"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "period": "2026-05",
    "company": "北京XX科技有限公司",
    "overall_status": "red",
    "checks": [
      { "id": 1, "name": "凭证借贷平衡", "status": "passed", "message": "全部凭证借贷平衡" },
      { "id": 2, "name": "凭证完整性", "status": "warning", "message": "3 张凭证待审核", "redirect": "/vouchers/review" },
      { "id": 3, "name": "凭证编号连续性", "status": "passed", "message": "编号连续" },
      { "id": 4, "name": "固定资产折旧", "status": "blocked", "message": "1 笔折旧计划未过账", "redirect": "/period/depreciation" },
      { "id": 5, "name": "银行日记账一致性", "status": "passed", "message": "全部一致" },
      { "id": 6, "name": "现金账实一致", "status": "warning", "message": "本月尚未盘点" },
      { "id": 7, "name": "往来核销完成度", "status": "blocked", "message": "2 笔超 30 天未核销，金额 ¥35,000" },
      { "id": 8, "name": "进项发票到期", "status": "warning", "message": "3 张发票即将过期" },
      { "id": 9, "name": "损益结转", "status": "blocked", "message": "损益类科目尚未结转" },
      { "id": 10, "name": "期间锁定", "status": "passed", "message": "当前期间未锁定" }
    ],
    "blockers": 3,
    "warnings": 3
  }
}
```

### POST /api/v1/periods/close

执行结账。所有阻断项必须已修复。

**Request Body:**
```json
{
  "period": "2026-05",
  "company_id": "uuid"
}
```

### POST /api/v1/periods/reopen

反结账（需主管审批）。

---

## 固定资产折旧 (F6)

### GET /api/v1/assets

固定资产列表。

### POST /api/v1/assets

新增资产。

### GET /api/v1/depreciation-schedules?asset_id=uuid&posted=false

获取待过账折旧计划。

### POST /api/v1/depreciation-schedules/run

执行折旧过账（生成折旧凭证）。

---

## 财务报表 (F6)

### GET /api/v1/reports/balance-sheet

资产负债表。

**Query Params:** `?period=2026-05&company_id=uuid&show_zero=true`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "title": "资产负债表",
    "company": "北京XX科技有限公司",
    "period": "2026-05",
    "generated_at": "2026-05-27T12:00:00Z",
    "columns": ["科目编码", "科目名称", "期初余额", "期末余额"],
    "rows": [
      { "account_code": "1001", "account_name": "银行存款", "opening": "1000000.00", "closing": "1250000.00", "level": 0, "is_group": true },
      { "account_code": "1001-01", "account_name": "工行", "opening": "800000.00", "closing": "1000000.00", "level": 1, "is_group": false },
      { "account_code": "1001-02", "account_name": "建行", "opening": "200000.00", "closing": "250000.00", "level": 1, "is_group": false }
    ],
    "totals": {
      "total_assets": "3500000.00",
      "total_liabilities": "1200000.00",
      "total_equity": "2300000.00"
    }
  }
}
```

### GET /api/v1/reports/profit-loss

利润表。

**Query Params:** `?period=2026-05&company_id=uuid`

### GET /api/v1/reports/cash-flow

现金流量表。

**Query Params:** `?period=2026-05&company_id=uuid&method=indirect`

### POST /api/v1/reports/export

导出报表。

**Request Body:**
```json
{
  "report_type": "balance_sheet",
  "period": "2026-05",
  "format": "pdf",
  "company_id": "uuid"
}
```

---

## 经营分析 (F7)

### POST /api/v1/analytics/query

自然语言查询。

**Request Body:**
```json
{
  "question": "5月净利润多少？",
  "company_id": "uuid",
  "period": "2026-05"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "answer": "5月净利润 ¥123,000.00，环比上月（¥87,000）增长 41.4%。主要增长来源：主营业务收入增加 ¥150,000，但营业成本同步增加 ¥100,000。",
    "sources": [
      { "label": "主营业务收入(6001)", "value": "850,000.00", "voucher_ref": "/vouchers?id=xxx" },
      { "label": "营业成本(6401)", "value": "480,000.00", "voucher_ref": "/vouchers?id=yyy" }
    ],
    "alerts": [
      { "type": "warning", "message": "管理费用-差旅费超支 20%，本月 ¥25,000 vs 预算 ¥20,000" }
    ],
    "data_timestamp": "2026-05-27T12:00:00Z"
  }
}
```

### GET /api/v1/analytics/dashboard

经营看板数据。

**Query Params:** `?company_id=uuid&period=2026-05`

### GET /api/v1/analytics/cash-flow-forecast

现金流预测。

**Query Params:** `?company_id=uuid&horizon=6`

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "horizon": 6,
    "monthly": [
      { "month": "2026-06", "opening": "1250000.00", "receivable": "800000.00", "payable": "600000.00", "balance": "1450000.00" },
      { "month": "2026-07", "opening": "1450000.00", "receivable": "750000.00", "payable": "650000.00", "balance": "1550000.00" }
    ],
    "alerts": [
      { "month": "2026-09", "type": "warning", "message": "预计资金缺口 ¥200,000" }
    ],
    "confidence": "medium",
    "note": "基于近 12 个月历史数据预测，仅供参考"
  }
}
```

### POST /api/v1/analytics/monthly-report/generate

生成经营月报长图。

**Request Body:**
```json
{
  "company_id": "uuid",
  "period": "2026-05"
}
```

**Response 200:**
```json
{
  "code": 0,
  "data": {
    "report_url": "/api/v1/files/monthly-report-2026-05.png",
    "generated_at": "2026-05-27T12:05:00Z"
  }
}
```
