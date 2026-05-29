# TASK-F0.2 | F0 | 核心数据模型

**版本**：V1.0
**日期**：2026-05-27
**优先级**：P0（基础支撑）
**状态**：待开发

---

## 任务描述

按照 requirements-v4.0-full.md 第六章，创建全部核心表结构：

### 4.1 科目表（树形嵌套集合）

```sql
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,        -- 4-2-2-2 结构，e.g. "1001-01-00-00"
    name VARCHAR(100) NOT NULL,
    account_type VARCHAR(20),               -- asset/liability/expense/income/equity
    root_type VARCHAR(10),                  -- debit/credit，余额方向
    parent_id UUID REFERENCES accounts(id),  -- NULL 表示根科目
    lft INT NOT NULL,                        -- 嵌套集合左值
    rgt INT NOT NULL,                        -- 嵌套集合右值
    is_group BOOLEAN DEFAULT FALSE,         -- TRUE=汇总科目不可记账，FALSE=明细科目可记账
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,                 -- 多租户隔离
    currency VARCHAR(3) DEFAULT 'CNY',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_accounts_tree ON accounts(lft, rgt);
CREATE INDEX idx_accounts_tenant ON accounts(tenant_id);
```

### 4.2 凭证主表 + 分录

```sql
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voucher_no VARCHAR(50) UNIQUE NOT NULL,   -- 格式：记-YYYY-MM-NNNN
    voucher_type VARCHAR(30),                 -- 记/银/现/转/折旧/结转
    posting_date DATE NOT NULL,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    remark TEXT,
    docstatus SMALLINT DEFAULT 0,             -- 0=draft, 1=submitted, 2=cancelled
    reversed_id UUID,                           -- 被哪张凭证冲销
    reversal_id UUID,                           -- 冲销了哪张凭证
    submitted_by UUID,
    submitted_at TIMESTAMP,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE journal_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id),
    debit DECIMAL(18,2) DEFAULT 0 CHECK (debit >= 0),
    credit DECIMAL(18,2) DEFAULT 0 CHECK (credit >= 0),
    debit_ccy DECIMAL(18,2) DEFAULT 0,
    credit_ccy DECIMAL(18,2) DEFAULT 0,
    account_ccy VARCHAR(3),
    exchange_rate DECIMAL(18,6) DEFAULT 1.0,
    party_type VARCHAR(20),                   -- customer/supplier/employee
    party_id UUID,
    cost_center_id UUID,
    project_id UUID,
    user_remark VARCHAR(200),
    reconciled BOOLEAN DEFAULT FALSE,
    tenant_id UUID NOT NULL,
    UNIQUE(journal_entry_id, account_id, COALESCE(party_id, '00000000-0000-0000-0000-000000000000'))
);

-- 凭证借贷平衡 CHECK（通过触发器实现，应用层在提交前检查借贷合计）
```

### 4.3 GL Entry（总账条目）

```sql
CREATE TABLE gl_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id),
    posting_date DATE NOT NULL,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    debit_ccy DECIMAL(18,2) DEFAULT 0,
    credit_ccy DECIMAL(18,2) DEFAULT 0,
    account_ccy VARCHAR(3),
    voucher_type VARCHAR(30),                 -- journal_entry/invoice/payment
    voucher_id UUID,                          -- 关联原始单据 ID
    against_voucher_type VARCHAR(30),
    against_voucher_id UUID,
    party_type VARCHAR(20),
    party_id UUID,
    cost_center_id UUID,
    project_id UUID,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    is_cancelled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_gl_voucher ON gl_entries(voucher_type, voucher_id);
CREATE INDEX idx_gl_posting ON gl_entries(posting_date, account_id);
CREATE INDEX idx_gl_tenant ON gl_entries(tenant_id);
```

### 4.4 发票 + 付款 + 核销

```sql
CREATE TABLE sales_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_no VARCHAR(50) UNIQUE NOT NULL,
    invoice_type VARCHAR(20) DEFAULT 'sale',   -- sale/purchase/credit_note
    customer_id UUID NOT NULL,
    tax_id VARCHAR(20),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    posting_date DATE NOT NULL,
    due_date DATE,
    total_amount DECIMAL(18,2) NOT NULL,      -- 价税合计
    tax_amount DECIMAL(18,2) DEFAULT 0,
    net_amount DECIMAL(18,2) DEFAULT 0,       -- 不含税金额
    outstanding_amount DECIMAL(18,2) NOT NULL, -- 未结清金额
    status VARCHAR(20) DEFAULT 'unpaid',     -- unpaid/partially_paid/paid/credit_note/written_off
    tax_template_id UUID,
    return_against UUID,
    is_return BOOLEAN DEFAULT FALSE,
    docstatus SMALLINT DEFAULT 0,
    created_by UUID, created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE payment_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_no VARCHAR(50) UNIQUE NOT NULL,
    payment_type VARCHAR(10) NOT NULL,        -- receive/pay
    party_type VARCHAR(20) NOT NULL,
    party_id UUID NOT NULL,
    paid_from_id UUID REFERENCES accounts(id),
    paid_to_id UUID REFERENCES accounts(id),
    paid_amount DECIMAL(18,2) NOT NULL,
    received_amount DECIMAL(18,2),
    reference_no VARCHAR(50),
    reference_date DATE,
    posting_date DATE NOT NULL,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    bank_account_id UUID,
    docstatus SMALLINT DEFAULT 0,
    created_by UUID, created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE payment_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_entry_id UUID REFERENCES payment_entries(id),
    invoice_id UUID NOT NULL,
    invoice_type VARCHAR(30),
    allocated_amount DECIMAL(18,2) NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 4.5 银行流水 + 对账

```sql
CREATE TABLE bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    clearing_account_id UUID REFERENCES accounts(id),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    currency VARCHAR(3) DEFAULT 'CNY',
    iban VARCHAR(50),
    swift_code VARCHAR(20),
    bank_account_type VARCHAR(20), -- savings/current
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE bank_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID REFERENCES bank_accounts(id),
    txn_date DATE NOT NULL,
    description TEXT,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    direction VARCHAR(4),                    -- in/out
    reference_no VARCHAR(100),
    counterparty_name VARCHAR(100),
    matched BOOLEAN DEFAULT FALSE,
    matched_payment_entry_id UUID,
    matched_gl_entry_id UUID,
    imported_from VARCHAR(20),               -- csv/excel/camt053/mt940
    raw_data JSONB,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE bank_reconciliation_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_transaction_id UUID REFERENCES bank_transactions(id),
    payment_entry_id UUID REFERENCES payment_entries(id),
    gl_entry_id UUID REFERENCES gl_entries(id),
    match_score DECIMAL(5,2),
    difference_account_id UUID REFERENCES accounts(id),
    reconciled_at TIMESTAMP,
    reconciled_by UUID,
    tenant_id UUID NOT NULL
);

CREATE TABLE bank_reconciliation_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID REFERENCES bank_accounts(id),
    statement_date DATE NOT NULL,
    bank_statement_balance DECIMAL(18,2) NOT NULL,
    gl_balance DECIMAL(18,2) NOT NULL,
    difference DECIMAL(18,2) DEFAULT 0,
    bank_only_total DECIMAL(18,2) DEFAULT 0,
    gl_only_total DECIMAL(18,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft',
    locked BOOLEAN DEFAULT FALSE,
    locked_by UUID,
    tenant_id UUID NOT NULL,
    created_by UUID, created_at TIMESTAMP DEFAULT NOW()
);
```

### 4.6 固定资产

```sql
CREATE TABLE asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    depreciation_method VARCHAR(20), -- straight_line/wdv/ddb/manual
    total_number_depreciations INT,
    frequency_of_depreciation INT,
    rate DECIMAL(6,4),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    fixed_asset_account_id UUID REFERENCES accounts(id),
    accumulated_depreciation_account_id UUID REFERENCES accounts(id),
    depreciation_expense_account_id UUID REFERENCES accounts(id),
    cwip_account_id UUID REFERENCES accounts(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_name VARCHAR(100) NOT NULL,
    asset_category_id UUID REFERENCES asset_categories(id),
    item_id UUID,
    purchase_date DATE,
    gross_purchase_amount DECIMAL(18,2) NOT NULL,
    available_for_use_date DATE,
    calculate_depreciation BOOLEAN DEFAULT FALSE,
    depreciation_method VARCHAR(20),
    total_number_depreciations INT,
    frequency_of_depreciation INT,
    expected_value_after_useful_life DECIMAL(18,2) DEFAULT 0,
    current_value DECIMAL(18,2),
    accumulated_depreciation DECIMAL(18,2) DEFAULT 0,
    status VARCHAR(30) DEFAULT 'draft',
    fixed_asset_account_id UUID REFERENCES accounts(id),
    depreciation_expense_account_id UUID REFERENCES accounts(id),
    accumulated_depreciation_account_id UUID REFERENCES accounts(id),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    custodian_id UUID,
    location VARCHAR(100),
    created_by UUID, created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE depreciation_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID REFERENCES assets(id),
    schedule_date DATE NOT NULL,
    depreciation_amount DECIMAL(18,2) NOT NULL,
    posted BOOLEAN DEFAULT FALSE,
    journal_entry_id UUID REFERENCES journal_entries(id),
    tenant_id UUID NOT NULL,
    UNIQUE(asset_id, schedule_date, posted)
);
```

### 4.7 预算

```sql
CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    budget_against VARCHAR(20) NOT NULL,  -- cost_center/project
    fiscal_year VARCHAR(10) NOT NULL,
    monthly_distribution VARCHAR(20),     -- monthly/quarterly/half_yearly/yearly
    status VARCHAR(20) DEFAULT 'draft',
    created_by UUID, created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(company_id, budget_against, fiscal_year, tenant_id)
);

CREATE TABLE budget_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID REFERENCES budgets(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id),
    annual_budget DECIMAL(18,2) NOT NULL,
    tenant_id UUID NOT NULL
);

CREATE TABLE budget_distributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_account_id UUID REFERENCES budget_accounts(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    percent DECIMAL(5,2),
    tenant_id UUID NOT NULL
);

CREATE TABLE budget_control_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id),
    cost_center_id UUID,
    fiscal_year VARCHAR(10),
    action_annual VARCHAR(20) DEFAULT 'warn',
    action_monthly VARCHAR(20) DEFAULT 'warn',
    applicable_on_mr BOOLEAN DEFAULT TRUE,
    applicable_on_po BOOLEAN DEFAULT TRUE,
    applicable_on_actual BOOLEAN DEFAULT TRUE,
    exception_approver_role VARCHAR(50),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL
);
```

### 4.8 审计日志

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(50) NOT NULL,            -- create/update/delete/submit/cancel/reverse
    object_type VARCHAR(50) NOT NULL,         -- journal_entry/payment_entry/invoice
    object_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    actor_id UUID NOT NULL,                  -- 操作人 ID
    actor_name VARCHAR(100),
    changed_fields JSONB,                    -- 变更字段（字段名→[旧值,新值]）
    metadata JSONB,                           -- 额外信息（如 IP 地址）
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_audit_object ON audit_logs(object_type, object_id);
CREATE INDEX idx_audit_actor ON audit_logs(actor_id, created_at);
CREATE INDEX idx_audit_tenant ON audit_logs(tenant_id);
-- 注意：audit_logs 表本身禁用 UPDATE/DELETE 接口，只有 INSERT
```

---

## 验收标准

- [ ] 所有表含 `tenant_id` 字段，RLS 策略对所有业务表生效
- [ ] `accounts` 支持树形查询（lft/rgt），Group 科目不可直接记账（INSERT 时检查 is_group）
- [ ] `journal_entry_lines` 提交前借贷合计差值为 0，不为 0 时返回错误
- [ ] `gl_entries` 通过触发器自动生成（Journal Entry Submit 时），不可绕过
- [ ] `audit_logs` 无 update/delete 接口，只有 insert
- [ ] 所有外键约束含 `ON DELETE RESTRICT`，禁止级联删除
- [ ] 迁移脚本可在空数据库上执行成功（ idempotent ）

---

## 前置依赖

TASK-F0.1（项目脚手架），因为需要先有 PostgreSQL 连接配置

---

## 预计工时

- 最小：32h
- 最大：64h

---

## 技术提示

### 嵌套集合树形模型

- 插入/删除节点时需要重新计算 lft/rgt，建议使用触发器或在 Service 层封装事务
- 树形查询：`SELECT * FROM accounts WHERE lft > $parent_lft AND rgt < $parent_rgt ORDER BY lft`
- 根节点查找：`SELECT * FROM accounts WHERE parent_id IS NULL`

### 凭证平衡强制

```go
// journal_entry_lines 提交前检查
func (s *JournalEntryService) Submit(ctx context.Context, id uuid.UUID) error {
    lines, _ := s.repo.GetLines(ctx, id)
    var totalDebit, totalCredit decimal.Decimal
    for _, l := range lines {
        totalDebit = totalDebit.Add(l.Debit)
        totalCredit = totalCredit.Add(l.Credit)
    }
    if !totalDebit.Equal(totalCredit) {
        return fmt.Errorf("借贷不平衡: 借方 %.2f != 贷方 %.2f", totalDebit, totalCredit)
    }
    // 更新 docstatus → 1
    // 生成 gl_entries（双写）
}
```

### 参考资料

- ERPNext journal_entry.py：参考 `docstatus` 状态机实现
- ERPNext GL Entry 双写模式：`erpnext.accounts.utils.make_gl_entries()`
- PostgreSQL CHECK 约束：https://www.postgresql.org/docs/current/ddl-constraints.html

---

## 上下文信息（架构师决策记录）

- **决策**：accounts 表使用嵌套集合（lft/rgt）而非邻接表，因为财务科目查询需要一次性加载整棵树，性能最优
- **决策**：不使用 GORM AutoMigrate，全部表通过手写 SQL Migration 管理，确保 RLS 策略和约束与表结构同步
- **决策**：GL Entry 通过数据库触发器或 Service 层双写，不允许应用层绕过直接操作 GL 表
- **风险**：嵌套集合 lft/rgt 在并发插入时需要锁表，建议所有科目操作封装在事务内