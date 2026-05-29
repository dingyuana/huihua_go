# API接口对比分析报告

## 概述

| 项目 | 技术栈 | 端口 | 前缀 |
|------|--------|------|------|
| Go项目 (huihua-finance) | Go Fiber + PostgreSQL | 8080 | /api/v1 |
| Python项目 (huihua-financial-master) | Python FastAPI + MySQL | 8000 | /api/v1 |

---

## 1. API路由清单对比

### 1.1 Go项目路由 (Go Fiber)

| 模块 | 路径 | 方法 | 功能 |
|------|------|------|------|
| **健康检查** | /health | GET | 健康检查 |
| **审计日志** | /api/v1/audit-logs | GET | 获取审计日志列表 |
| **审计日志** | /api/v1/audit-logs/:object_type/:object_id | GET | 按对象获取审计日志 |
| **科目** | /api/v1/accounts/tree | GET | 获取科目树 |
| **科目** | /api/v1/accounts/init-seed | POST | 初始化科目数据 |
| **汇率** | /api/v1/exchange-rates | GET | 获取汇率列表 |
| **汇率** | /api/v1/exchange-rates | POST | 创建汇率 |
| **汇率** | /api/v1/exchange-rates/:id | GET | 获取汇率详情 |
| **汇率** | /api/v1/exchange-rates/:id | DELETE | 删除汇率 |
| **汇率** | /api/v1/exchange-rates/convert | GET | 汇率转换 |
| **银行账户** | /api/v1/bank-accounts | GET | 获取银行账户列表 |
| **银行账户** | /api/v1/bank-accounts | POST | 创建银行账户 |
| **银行账户** | /api/v1/bank-accounts/:id | PUT | 更新银行账户 |
| **银行账户** | /api/v1/bank-accounts/:id | DELETE | 删除银行账户 |
| **往来单位** | /api/v1/parties | GET | 获取往来单位列表 |
| **往来单位** | /api/v1/parties/import | POST | 导入往来单位Excel |
| **往来单位** | /api/v1/parties/:id | GET | 获取往来单位详情 |
| **往来单位** | /api/v1/parties | POST | 创建往来单位 |
| **往来单位** | /api/v1/parties/:id | PUT | 更新往来单位 |
| **往来单位** | /api/v1/parties/:id | DELETE | 删除往来单位 |
| **账套设置** | /api/v1/account-setup/status | GET | 获取账套状态 |
| **账套设置** | /api/v1/account-setup/wizard | POST | 创建公司/账套 |
| **资产折旧** | /api/v1/assets/:id/depreciation/schedule | POST | 创建折旧计划 |
| **资产折旧** | /api/v1/assets/:id/depreciation/schedule | GET | 获取折旧计划 |
| **资产折旧** | /api/v1/depreciation/run | POST | 执行折旧 |
| **资产折旧** | /api/v1/depreciation/run | GET | 获取折旧执行列表 |
| **发票** | /api/v1/invoices | GET | 获取发票列表 |
| **发票** | /api/v1/invoices | POST | 创建发票 |
| **发票** | /api/v1/invoices/import | POST | 从Excel导入发票 |
| **发票** | /api/v1/invoices/parse | POST | 解析发票 |
| **发票** | /api/v1/invoices/:id | GET | 获取发票详情 |
| **发票** | /api/v1/invoices/:id/status | PUT | 更新发票状态 |
| **分类规则** | /api/v1/classification-rules | GET | 获取分类规则列表 |
| **分类规则** | /api/v1/classification-rules | POST | 创建分类规则 |
| **分类规则** | /api/v1/classification-rules/:id | PUT | 更新分类规则 |
| **分类规则** | /api/v1/classification-rules/:id | DELETE | 删除分类规则 |
| **分类规则** | /api/v1/classification-rules/reorder | POST | 重新排序规则 |
| **分类规则** | /api/v1/classification-rules/match | POST | 匹配规则 |
| **银行流水** | /api/v1/bank-transactions | GET | 获取银行流水列表 |
| **银行流水** | /api/v1/bank-transactions/import | POST | 导入银行流水 |
| **银行流水** | /api/v1/bank-transactions/:id/classify | POST | 分类银行流水 |
| **银行流水** | /api/v1/bank-transactions/:id/mark-matched | POST | 标记为已匹配 |
| **银行流水** | /api/v1/bank-transactions/unmatched | GET | 获取未匹配流水 |
| **银行流水** | /api/v1/bank-transactions/:id | GET | 获取银行流水详情 |
| **银行流水** | /api/v1/bank-transactions/:id | DELETE | 删除银行流水 |
| **凭证模板** | /api/v1/voucher-templates | GET | 获取凭证模板列表 |
| **凭证模板** | /api/v1/voucher-templates | POST | 创建凭证模板 |
| **凭证模板** | /api/v1/voucher-templates/:id | GET | 获取凭证模板详情 |
| **凭证模板** | /api/v1/voucher-templates/:id | PUT | 更新凭证模板 |
| **凭证模板** | /api/v1/voucher-templates/:id | DELETE | 删除凭证模板 |
| **凭证模板** | /api/v1/voucher-templates/numbering-rule | GET | 获取编号规则 |
| **凭证模板** | /api/v1/voucher-templates/numbering-rule | POST | 更新编号规则 |
| **凭证模板** | /api/v1/voucher-templates/numbering-rule/next | POST | 生成下一个编号 |
| **凭证** | /api/v1/vouchers | GET | 获取凭证列表 |
| **凭证** | /api/v1/vouchers | POST | 创建凭证 |
| **凭证** | /api/v1/vouchers/:id | GET | 获取凭证详情 |
| **凭证** | /api/v1/vouchers/:id | PUT | 更新凭证 |
| **凭证** | /api/v1/vouchers/:id | DELETE | 删除凭证 |
| **凭证** | /api/v1/vouchers/:id/submit | POST | 提交凭证 |
| **凭证** | /api/v1/vouchers/:id/approve | POST | 审核凭证 |
| **凭证** | /api/v1/vouchers/:id/reject | POST | 驳回凭证 |
| **凭证** | /api/v1/vouchers/:id/cancel | POST | 作废凭证 |
| **凭证** | /api/v1/vouchers/:id/reverse | POST | 红字冲销 |
| **凭证** | /api/v1/vouchers/:id/status | GET | 获取凭证状态 |
| **凭证** | /api/v1/vouchers/:id/transitions | GET | 获取状态转换 |
| **期初余额** | /api/v1/opening-balances/import | POST | 导入期初余额 |
| **期初余额** | /api/v1/opening-balances | GET | 获取期初余额列表 |
| **期初余额** | /api/v1/opening-balances/trial-balance | GET | 获取试算平衡表 |
| **期初余额** | /api/v1/opening-balances/validate | POST | 验证期初余额 |
| **期初余额** | /api/v1/opening-balances/:account_id | GET | 按科目获取期初余额 |
| **会计期间** | /api/v1/periods | GET | 获取会计期间列表 |
| **会计期间** | /api/v1/periods/current | GET | 获取当前会计期间 |
| **会计期间** | /api/v1/periods/:period_no/close | POST | 结账 |
| **核销** | /api/v1/reconciliation/run | POST | 执行核销 |
| **核销** | /api/v1/reconciliation/pairs | GET | 获取核销对列表 |
| **核销** | /api/v1/reconciliation/pairs/:id/confirm | POST | 确认核销对 |
| **核销** | /api/v1/reconciliation/pairs/:id/unconfirm | POST | 取消核销对 |
| **核销** | /api/v1/reconciliation/unmatched | GET | 获取未核销列表 |
| **银行对账** | /api/v1/bank-reconciliation/reconcile | POST | 执行银行对账 |
| **银行对账** | /api/v1/bank-reconciliation/report | GET | 获取银行对账报告 |
| **银行对账** | /api/v1/bank-reconciliation/mark-done | POST | 标记对账完成 |
| **银行对账** | /api/v1/bank-reconciliation/status | GET | 获取对账状态 |
| **财务报表** | /api/v1/reports/trial-balance | GET | 获取试算平衡表 |
| **财务报表** | /api/v1/reports/income-statement | GET | 获取利润表 |
| **财务报表** | /api/v1/reports/balance-sheet | GET | 获取资产负债表 |
| **审批** | /api/v1/approvals/submit | POST | 提交审批 |
| **审批** | /api/v1/approvals/:id/approve | POST | 审批通过 |
| **审批** | /api/v1/approvals/:id/reject | POST | 审批驳回 |
| **审批** | /api/v1/approvals/pending | GET | 获取待审批任务 |
| **审批** | /api/v1/approvals/history | GET | 获取审批历史 |
| **审批** | /api/v1/approvals/voucher/:id/status | GET | 获取凭证审批状态 |
| **审批流程** | /api/v1/approval-flows | POST | 创建审批流程 |
| **审批流程** | /api/v1/approval-flows | GET | 获取审批流程列表 |
| **审批流程** | /api/v1/approval-flows/:id | PUT | 更新审批流程 |
| **审批流程** | /api/v1/approval-flows/:id | DELETE | 删除审批流程 |
| **自动生成凭证** | /api/v1/bank-transactions/:id/generate-voucher | POST | 从银行流水生成凭证 |
| **自动生成凭证** | /api/v1/bank-transactions/batch-generate | POST | 批量生成凭证 |

### 1.2 Python项目路由 (FastAPI)

| 模块 | 路径 | 方法 | 功能 |
|------|------|------|------|
| **认证** | /api/v1/auth/* | * | 认证相关路由 |
| **用户** | /api/v1/users/* | * | 用户管理路由 |
| **部门** | /api/v1/depts/* | * | 部门管理路由 |
| **角色权限** | /api/v1/roles/* | * | 角色权限路由 |
| **会计核算** | /api/v1/accounting/vouchers | GET | 获取凭证列表 |
| **会计核算** | /api/v1/accounting/vouchers/generate-no | GET | 生成凭证号 |
| **会计核算** | /api/v1/accounting/vouchers | POST | 创建凭证 |
| **会计核算** | /api/v1/accounting/vouchers/:voucher_id | GET | 获取凭证详情 |
| **会计核算** | /api/v1/accounting/vouchers/:voucher_id/audit | GET | 获取凭证操作日志 |
| **会计核算** | /api/v1/accounting/vouchers/:voucher_id | PUT | 更新凭证 |
| **会计核算** | /api/v1/accounting/vouchers/:voucher_id | DELETE | 删除凭证 |
| **报表中心** | /api/v1/reports/* | * | 财务报表路由 |
| **现金管理** | /api/v1/cash/banks | GET | 获取银行账户列表 |
| **现金管理** | /api/v1/cash/banks | POST | 创建银行账户 |
| **现金管理** | /api/v1/cash/banks/:bank_id | GET | 获取银行账户详情 |
| **现金管理** | /api/v1/cash/banks/:bank_id | PUT | 更新银行账户 |
| **现金管理** | /api/v1/cash/banks/:bank_id | DELETE | 删除银行账户 |
| **现金管理** | /api/v1/cash/transactions/summary | GET | 获取资金流水汇总 |
| **现金管理** | /api/v1/cash/transactions | GET | 获取资金流水列表 |
| **现金管理** | /api/v1/cash/transactions | POST | 创建资金流水 |
| **现金管理** | /api/v1/cash/transactions/reconcile | GET | 获取待对账流水 |
| **现金管理** | /api/v1/cash/transactions/reconcile | POST | 执行对账 |
| **现金管理** | /api/v1/cash/reconciliation/:bank_id | GET | 获取银行对账报告 |
| **现金管理** | /api/v1/cash/transactions/pending-pool | GET | 获取待处理流水池 |
| **现金管理** | /api/v1/cash/transactions/:txn_id | GET | 获取资金流水详情 |
| **现金管理** | /api/v1/cash/transactions/:txn_id | PUT | 更新资金流水 |
| **现金管理** | /api/v1/cash/transactions/:txn_id | DELETE | 删除资金流水 |
| **银行流水分类** | /api/v1/cash-category/* | * | 银行流水分类路由 |
| **税务管理** | /api/v1/tax/* | * | 税务管理路由 |
| **总账管理** | /api/v1/ledger/* | * | 总账管理路由 |
| **应收管理** | /api/v1/receivable/invoices | GET | 获取应收发票列表 |
| **应收管理** | /api/v1/receivable/summary | GET | 应收余额汇总 |
| **应收管理** | /api/v1/receivable/invoices/aging | GET | 应收账龄分析 |
| **应付管理** | /api/v1/payable/invoices | GET | 获取应付发票列表 |
| **应付管理** | /api/v1/payable/invoices/aging | GET | 应付账龄分析 |
| **应付管理** | /api/v1/payable/invoices/warnings | GET | 应付发票预警 |
| **固定资产** | /api/v1/fixed-assets/* | * | 固定资产路由 |
| **工资核算** | /api/v1/wage/* | * | 工资核算路由 |
| **发票管理** | /api/v1/invoice/purchases | GET | 获取发票领购列表 |
| **发票管理** | /api/v1/invoice/purchases/:purchase_id | GET | 获取领购详情 |
| **发票管理** | /api/v1/invoice/purchases | POST | 创建领购记录 |
| **发票管理** | /api/v1/invoice/purchases/:purchase_id | PUT | 更新领购记录 |
| **发票管理** | /api/v1/invoice/purchases/:purchase_id | DELETE | 删除领购记录 |
| **发票管理** | /api/v1/invoice/issues | GET | 获取发票开具列表 |
| **发票管理** | /api/v1/invoice/issues/:issue_id | GET | 获取发票详情 |
| **发票管理** | /api/v1/invoice/issues | POST | 开具发票 |
| **发票管理** | /api/v1/invoice/issues/:issue_id | PUT | 更新发票 |
| **发票管理** | /api/v1/invoice/issues/:issue_id | DELETE | 删除发票 |
| **发票管理** | /api/v1/invoice/issues/:issue_id/void | POST | 作废/红冲发票 |
| **发票管理** | /api/v1/invoice/issues/:issue_id/redflush | POST | 红冲发票 |
| **发票管理** | /api/v1/invoice/statistics | GET | 发票统计 |
| **发票管理** | /api/v1/invoice/available-numbers | GET | 可用发票号码 |
| **发票管理** | /api/v1/invoice/ocr | POST | 发票OCR识别 |
| **预算管理** | /api/v1/budget/* | * | 预算管理路由 |
| **成本管理** | /api/v1/cost/* | * | 成本管理路由 |
| **档案管理** | /api/v1/archive/* | * | 档案管理路由 |
| **系统设置** | /api/v1/company/* | * | 公司设置路由 |
| **期间管理** | /api/v1/period/* | * | 期间管理路由 |
| **审计日志** | /api/v1/audit/* | * | 审计日志路由 |
| **智能查询** | /api/v1/query/* | * | 智能查询路由 |
| **辅助核算** | /api/v1/auxiliary/* | * | 辅助核算路由 |
| **预警** | /api/v1/warnings/* | * | 预警路由 |
| **业务单据** | /api/v1/business/* | * | 业务单据路由 |
| **备份** | /api/v1/system/backup/* | * | 系统备份路由 |
| **核销管理** | /api/v1/reconciliation/* | * | 核销管理路由 |

---

## 2. 核心数据模型对比

### 2.1 凭证 (Voucher/Journal Entry)

#### Go项目 (JournalEntry)

```go
type JournalEntry struct {
    ID           uuid.UUID  `json:"id"`
    VoucherNo    string     `json:"voucher_no"`
    VoucherType  *string    `json:"voucher_type,omitempty"`
    PostingDate  time.Time  `json:"posting_date"`
    CompanyID    uuid.UUID  `json:"company_id"`
    TenantID     uuid.UUID  `json:"tenant_id"`
    Remark       *string    `json:"remark,omitempty"`
    DocStatus    int16      `json:"docstatus"`
    ReversedID   *uuid.UUID `json:"reversed_id,omitempty"`
    ReversalID   *uuid.UUID `json:"reversal_id,omitempty"`
    SubmittedBy  *uuid.UUID `json:"submitted_by,omitempty"`
    SubmittedAt  *time.Time `json:"submitted_at,omitempty"`
    CreatedBy    uuid.UUID  `json:"created_by"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type JournalEntryLine struct {
    ID             uuid.UUID       `json:"id"`
    JournalEntryID uuid.UUID       `json:"journal_entry_id"`
    AccountID      uuid.UUID       `json:"account_id"`
    Debit          decimal.Decimal `json:"debit"`
    Credit         decimal.Decimal `json:"credit"`
    DebitCcy       decimal.Decimal `json:"debit_ccy"`
    CreditCcy      decimal.Decimal `json:"credit_ccy"`
    AccountCcy     *string         `json:"account_ccy,omitempty"`
    ExchangeRate   decimal.Decimal `json:"exchange_rate"`
    PartyType      *string         `json:"party_type,omitempty"`
    PartyID        *uuid.UUID      `json:"party_id,omitempty"`
    CostCenterID   *uuid.UUID      `json:"cost_center_id,omitempty"`
    ProjectID      *uuid.UUID      `json:"project_id,omitempty"`
    UserRemark     *string         `json:"user_remark,omitempty"`
    Reconciled     bool            `json:"reconciled"`
    TenantID       uuid.UUID       `json:"tenant_id"`
}
```

#### Python项目 (AccVoucher)

```python
class AccVoucher(Base):
    id = Column(String(36), primary_key=True)
    voucher_no = Column(String(20), nullable=False)
    voucher_date = Column(Date, nullable=False)
    attachment_count = Column(Integer, default=0)
    status = Column(String(20), default="draft")  # draft/approved/posted/closed
    source_type = Column(String(50), default="manual")
    source_id = Column(String(36))
    voucher_word = Column(String(10), default="记")
    summary = Column(String(500))
    created_by = Column(String(36))
    created_by_name = Column(String(100))
    created_at = Column(DateTime)
    submitted_by = Column(String(36))
    submitted_by_name = Column(String(100))
    submitted_at = Column(DateTime)
    approved_by = Column(String(36))
    approved_by_name = Column(String(100))
    approved_at = Column(DateTime)
    posted_by = Column(String(36))
    posted_by_name = Column(String(100))
    posted_at = Column(DateTime)
    closed_by = Column(String(36))
    closed_by_name = Column(String(100))
    closed_at = Column(DateTime)
    period_year = Column(Integer, nullable=False)
    period_month = Column(Integer, nullable=False)
    remark = Column(String(500))
    reversed_from = Column(String(36), ForeignKey("acc_voucher.id"))
    reversed_by = Column(String(36))
    reversed_by_name = Column(String(100))
    reversed_at = Column(DateTime)
    reversed_reason = Column(String(500))
    is_reversed = Column(Boolean, default=False)
    reject_reason = Column(String(500))
    rejected_by = Column(String(36))
    rejected_by_name = Column(String(100))
    rejected_at = Column(DateTime)

class AccVoucherDetail(Base):
    id = Column(String(36), primary_key=True)
    voucher_id = Column(String(36), ForeignKey("acc_voucher.id"))
    line_no = Column(Integer, nullable=False)
    summary = Column(String(500))
    subject_code = Column(String(50), nullable=False)
    subject_name = Column(String(200), nullable=False)
    auxiliary_info = Column(String(500))  # JSON
    debit_amount = Column(Numeric(18, 2), default=0)
    credit_amount = Column(Numeric(18, 2), default=0)
    quantity = Column(Numeric(18, 4))
    unit_price = Column(Numeric(18, 4))
    exchange_rate = Column(Numeric(10, 6))
    foreign_amount = Column(Numeric(18, 2))
```

#### 凭证字段对比

| 字段 | Go项目 | Python项目 | 映射说明 |
|------|--------|------------|----------|
| 主键 | uuid | String(36) | 类型不同但功能等价 |
| 凭证号 | voucher_no | voucher_no | 一致 |
| 凭证日期 | posting_date | voucher_date | 名称不同 |
| 附件数 | (无) | attachment_count | Python独有 |
| 状态 | docstatus (int16) | status (String) | 类型不同，值也不同 |
| 凭证字 | voucher_type | voucher_word | 名称不同 |
| 摘要 | remark | summary | 名称不同 |
| 期间年 | (无) | period_year | Python独有 |
| 期间月 | (无) | period_month | Python独有 |
| 冲红ID | reversed_id | reversed_from | 名称不同 |
| 冲红原因 | (无) | reversed_reason | Python独有 |
| 驳回原因 | (无) | reject_reason | Python独有 |
| 借方金额 | debit | debit_amount | 名称不同 |
| 贷方金额 | credit | credit_amount | 名称不同 |
| 科目编码 | AccountID (uuid) | subject_code (String) | 类型不同 |

**状态值对比：**

| Go项目 (docstatus) | Python项目 (status) |
|-------------------|-------------------|
| 0 | draft |
| 1 | approved |
| 2 | posted |
| 3 | closed |
| (无) | rejected |

---

### 2.2 科目 (Account)

#### Go项目
- BankAccount: bank_name, account_number, clearing_account_id, currency, iban, swift_code, bank_account_type, is_active

#### Python项目
- CaBank: bank_name, account_name, account_number, account_type, currency, balance, is_active, is_default, remark
- AccSubject/AccAccount: 会计科目

**银行账户字段对比：**

| 字段 | Go项目 | Python项目 | 映射说明 |
|------|--------|------------|----------|
| 银行名称 | bank_name | bank_name | 一致 |
| 账号 | account_number | account_number | 一致 |
| 账户名称 | (无) | account_name | Python独有 |
| 结算科目ID | clearing_account_id | (无) | Go独有 |
| 币种 | currency | currency | 一致 |
| IBAN | iban | (无) | Go独有 |
| SwiftCode | swift_code | (无) | Go独有 |
| 账户类型 | bank_account_type | account_type | 名称不同 |
| 是否启用 | is_active | is_active | 一致 |
| 余额 | (无) | balance | Python独有 |

---

### 2.3 发票 (Invoice)

#### Go项目 (SalesInvoice)

```go
type SalesInvoice struct {
    ID                uuid.UUID       `json:"id"`
    InvoiceNo         string          `json:"invoice_no"`
    InvoiceType       string          `json:"invoice_type"`
    CustomerID        uuid.UUID       `json:"customer_id"`
    TaxID             *string         `json:"tax_id,omitempty"`
    CompanyID         uuid.UUID       `json:"company_id"`
    TenantID          uuid.UUID       `json:"tenant_id"`
    PostingDate       time.Time       `json:"posting_date"`
    DueDate           *time.Time      `json:"due_date,omitempty"`
    TotalAmount       decimal.Decimal `json:"total_amount"`
    TaxAmount         decimal.Decimal `json:"tax_amount"`
    NetAmount         decimal.Decimal `json:"net_amount"`
    OutstandingAmount decimal.Decimal `json:"outstanding_amount"`
    Status            string          `json:"status"`
    TaxTemplateID     *uuid.UUID      `json:"tax_template_id,omitempty"`
    ReturnAgainst     *uuid.UUID      `json:"return_against,omitempty"`
    IsReturn          bool            `json:"is_return"`
    DocStatus         int16           `json:"docstatus"`
    CreatedBy         *uuid.UUID      `json:"created_by,omitempty"`
    CreatedAt         time.Time       `json:"created_at"`
}
```

#### Python项目

- InvoiceIssue: invoice_code, invoice_number, invoice_date, invoice_type, buyer_name, buyer_tax_no, amount, tax_rate, tax_amount, total_amount, seller_name, issuer, status
- InvoicePurchase: purchase_date, invoice_type, start_number, end_number, quantity, purchaser, status
- ArInvoice: invoice_no, invoice_date, total_amount, tax_rate, tax_amount, net_amount, writeoff_amount, balance_amount, status, customer_id
- ApInvoice: 类似ArInvoice

**发票字段对比：**

| 字段 | Go项目 | Python项目 | 映射说明 |
|------|--------|------------|----------|
| 发票号 | invoice_no | invoice_number | 名称不同 |
| 发票类型 | invoice_type | invoice_type | 一致 |
| 客户ID | customer_id | customer_id | 一致 |
| 税率 | (无) | tax_rate | Python独有 |
| 税额 | tax_amount | tax_amount | 一致 |
| 总金额 | total_amount | total_amount | 一致 |
| 不含税金额 | net_amount | net_amount | 一致 |
| 余额 | outstanding_amount | balance_amount | 名称不同 |
| 状态 | status | status | 一致 |
| 开票日期 | posting_date | invoice_date | 名称不同 |

---

### 2.4 银行流水 (Bank Transaction)

#### Go项目

```go
type BankTransaction struct {
    ID                     uuid.UUID       `json:"id"`
    BankAccountID          uuid.UUID       `json:"bank_account_id"`
    TxnDate                time.Time       `json:"txn_date"`
    Description            *string         `json:"description,omitempty"`
    Debit                  decimal.Decimal `json:"debit"`
    Credit                 decimal.Decimal `json:"credit"`
    Direction              *string         `json:"direction,omitempty"`
    ReferenceNo            *string         `json:"reference_no,omitempty"`
    CounterpartyName       *string         `json:"counterparty_name,omitempty"`
    Matched                bool            `json:"matched"`
    MatchedPaymentEntryID  *uuid.UUID      `json:"matched_payment_entry_id,omitempty"`
    MatchedGLEntryID       *uuid.UUID      `json:"matched_gl_entry_id,omitempty"`
    ImportedFrom           *string         `json:"imported_from,omitempty"`
    RawData                json.RawMessage `json:"raw_data,omitempty"`
    CompanyID              uuid.UUID       `json:"company_id"`
    TenantID               uuid.UUID       `json:"tenant_id"`
    CreatedAt              time.Time       `json:"created_at"`
}
```

#### Python项目

```python
class CaTransaction(Base):
    id = Column(String(36), primary_key=True)
    transaction_no = Column(String(50), unique=True)
    bank_id = Column(String(36), ForeignKey("ca_bank.id"))
    transaction_date = Column(DateTime, nullable=False)
    transaction_type = Column(String(20))  # inflow/outflow
    amount = Column(Numeric(18, 2), nullable=False)
    balance_after = Column(Numeric(18, 2))
    counterparty_name = Column(String(200))
    counterparty_account = Column(String(50))
    counterparty_bank = Column(String(100))
    summary = Column(String(500))
    category = Column(String(50))
    business_type = Column(String(50))
    ref_no = Column(String(100))
    status = Column(String(20))  # pending/document_generated/reconciled
    is_reconciled = Column(Boolean, default=False)
    reconciled_at = Column(DateTime)
    reconciled_by = Column(String(36))
    reconciliation_remark = Column(Text)
```

**银行流水字段对比：**

| 字段 | Go项目 | Python项目 | 映射说明 |
|------|--------|------------|----------|
| 主键 | uuid | String(36) | 类型不同 |
| 银行账户ID | bank_account_id | bank_id | 名称不同 |
| 交易日期 | txn_date | transaction_date | 名称不同 |
| 描述 | description | summary | 名称不同 |
| 借方金额 | debit | (通过transaction_type区分) | 需转换 |
| 贷方金额 | credit | (通过transaction_type区分) | 需转换 |
| 交易方向 | direction | transaction_type | 名称不同 |
| 参考号 | reference_no | transaction_no | 名称不同 |
| 对方名称 | counterparty_name | counterparty_name | 一致 |
| 对方账号 | (无) | counterparty_account | Python独有 |
| 对方银行 | (无) | counterparty_bank | Python独有 |
| 是否匹配 | matched | is_reconciled | 名称不同/布尔值类型不同 |
| 对账时间 | (无) | reconciled_at | Python独有 |
| 对账人 | (无) | reconciled_by | Python独有 |
| 对账备注 | (无) | reconciliation_remark | Python独有 |
| 分类 | (无) | category | Python独有 |
| 业务类型 | (无) | business_type | Python独有 |
| 关联单据号 | (无) | ref_no | Python独有 |

---

### 2.5 往来单位 (Party/Customer/Supplier)

#### Go项目
- Party: 包含在往来单位中

#### Python项目
- CrmCustomer: 客户
- Supplier: 供应商

**说明：** Python项目将客户和供应商分开管理，Go项目统一为Party

---

### 2.6 审批流 (Approval Flow)

#### Go项目

| 路由 | 方法 | 功能 |
|------|------|------|
| /approvals/submit | POST | 提交审批 |
| /approvals/:id/approve | POST | 审批通过 |
| /approvals/:id/reject | POST | 审批驳回 |
| /approvals/pending | GET | 获取待审批任务 |
| /approvals/history | GET | 获取审批历史 |
| /approvals/voucher/:id/status | GET | 获取凭证审批状态 |
| /approval-flows | POST/GET | 创建/获取审批流程 |
| /approval-flows/:id | PUT/DELETE | 更新/删除审批流程 |

#### Python项目
- 审批流程集成在业务模块中，通过status状态机管理

---

## 3. 认证机制对比

### 3.1 JWT Token格式对比

| 项目 | 算法 | Secret配置 | 过期时间 |
|------|------|------------|----------|
| Go项目 | (从代码推断HS256) | cfg.JWT.Secret | (未明确) |
| Python项目 | HS256 | settings.SECRET_KEY | ACCESS_TOKEN_EXPIRE_MINUTES=30 |

### 3.2 Token Payload结构

#### Go项目 (jwt.ParseToken)
```go
claims.UserID    // uuid.UUID
claims.TenantID  // uuid.UUID
claims.Role      // string
```

#### Python项目 (decode_token)
```python
payload = {
    "sub": user_id,      # 用户ID (String)
    "exp": expire,       # 过期时间
    "type": "refresh"    # (可选) 刷新令牌类型
}
```

### 3.3 认证方式

| 项目 | Header格式 | 获取用户信息 |
|------|-----------|-------------|
| Go项目 | Authorization: Bearer <token> | c.Locals("user_id"), c.Locals("tenant_id"), c.Locals("role") |
| Python项目 | Authorization: Bearer <token> | Depends(get_current_user) |

### 3.4 认证差异总结

1. **Go项目**的用户信息存储在`c.Locals`中，包含user_id、tenant_id、role
2. **Python项目**使用FastAPI的依赖注入系统，通过`get_current_user`获取完整用户对象
3. **Python项目**支持刷新令牌(refresh_token)，Go项目需要确认是否支持
4. **Payload字段名不同**：
   - Go: user_id, tenant_id, role
   - Python: sub (对应user_id)

---

## 4. 通用响应格式对比

### 4.1 Go项目响应格式

```json
// 成功响应（直接返回数据）
{"data": ..., "error": null}
或者直接返回对象

// 错误响应
{"error": "error message"}
```

### 4.2 Python项目响应格式

```json
// 成功响应
{"code": 200, "msg": "success", "data": ...}

// 错误响应
{"detail": "error message"}
```

### 4.3 响应格式差异总结

| 项目 | 成功格式 | 错误格式 | 分页格式 |
|------|---------|---------|---------|
| Go项目 | {data: x, error: null} 或直接对象 | {error: msg} | (无统一格式) |
| Python项目 | {code: 200, msg: "success", data: x} | {detail: msg} | {total: n, list: [...]} |

**关键差异：**
1. **Python项目**有统一的成功响应包装`{code, msg, data}`
2. **Go项目**直接返回数据，无统一包装
3. **Python项目**使用`detail`字段表示错误，Go使用`error`
4. **Python项目**分页响应使用`{total, list}`结构

---

## 5. API兼容性问题列表

### 5.1 路径不一致问题

| 功能 | Go路径 | Python路径 | 建议 |
|------|--------|-----------|------|
| 凭证列表 | /api/v1/vouchers | /api/v1/accounting/vouchers | 前端需根据后端切换路径 |
| 银行账户 | /api/v1/bank-accounts | /api/v1/cash/banks | 路径不同 |
| 银行流水 | /api/v1/bank-transactions | /api/v1/cash/transactions | 路径不同 |
| 发票 | /api/v1/invoices | /api/v1/invoice/issues | 路径不同 |
| 科目 | /api/v1/accounts/tree | /api/v1/ledger/accounts | 路径不同 |

### 5.2 字段不一致问题

| 实体 | Go字段 | Python字段 | 映射建议 |
|------|--------|-----------|---------|
| 凭证日期 | posting_date | voucher_date | 需转换 |
| 凭证状态 | docstatus (int) | status (string) | 0=draft, 1=approved, 2=posted |
| 借方金额 | debit | debit_amount | 需转换 |
| 贷方金额 | credit | credit_amount | 需转换 |
| 对方名称 | counterparty_name | counterparty_name | 一致 |
| 银行流水金额 | debit/credit分开的两个字段 | amount一个字段+transaction_type区分 | 需转换 |
| 附件数 | (无) | attachment_count | 需忽略或默认0 |
| 期间信息 | (无直接字段) | period_year, period_month | Python独有 |

### 5.3 响应格式兼容

| 问题 | Go项目 | Python项目 | 前端处理 |
|------|--------|-----------|---------|
| 成功响应 | 直接数据或{data, error} | {code, msg, data} | 前端需适配两种格式 |
| 错误响应 | {error: msg} | {detail: msg} | 前端需适配两种格式 |
| 分页响应 | (无统一) | {total, list} | 前端需适配 |

### 5.4 数据类型差异

| 问题 | Go项目 | Python项目 | 影响 |
|------|--------|-----------|------|
| 主键类型 | uuid.UUID | String(36) | 前端统一作为字符串处理 |
| 金额类型 | decimal.Decimal | Numeric(18,2) | 都转为字符串或数字 |
| 日期类型 | time.Time | Date/Datetime | 前端统一转为ISO格式 |

### 5.5 认证兼容

| 问题 | Go项目 | Python项目 | 建议 |
|------|--------|-----------|------|
| Token字段名 | user_id, tenant_id | sub | 登录时获取并统一存储 |
| 过期时间 | 未明确 | 30分钟 | 前端需处理token刷新 |

---

## 6. 前端适配建议

### 6.1 API适配层设计

建议前端实现API适配层，统一两个后端的接口：

```javascript
// 统一响应格式
const unifiedResponse = (goResponse) => {
  if (goResponse.error) {
    return { success: false, error: goResponse.error }
  }
  return { success: true, data: goResponse.data || goResponse }
}

// 凭证状态映射
const voucherStatusMap = {
  0: 'draft',
  1: 'approved',
  2: 'posted',
  3: 'closed'
}

// Python到Go状态转换
const toGoStatus = (pyStatus) => {
  const reverseMap = { draft: 0, approved: 1, posted: 2, closed: 3 }
  return reverseMap[pyStatus] ?? 0
}
```

### 6.2 路径适配

| 功能模块 | Go路径前缀 | Python路径前缀 |
|---------|-----------|--------------|
| 凭证 | /api/v1/vouchers | /api/v1/accounting/vouchers |
| 银行账户 | /api/v1/bank-accounts | /api/v1/cash/banks |
| 银行流水 | /api/v1/bank-transactions | /api/v1/cash/transactions |
| 发票 | /api/v1/invoices | /api/v1/invoice |

### 6.3 字段映射表

建议在前端维护统一的字段映射配置，确保数据在各模块间流转时的一致性。

---

## 7. 总结

两个项目的API存在以下主要差异：

1. **路由结构**：Python项目按功能模块组织（accounting、cash、invoice等），Go项目相对扁平
2. **数据模型**：字段名、类型、状态值均有差异
3. **响应格式**：Python统一使用{code, msg, data}包装，Go直接返回
4. **认证机制**：Token payload字段名不同，需统一处理
5. **分页格式**：Python使用{total, list}，Go无统一格式

**建议**：前端应实现API适配层，隔离两个后端的差异，提供统一的接口给前端业务代码。
