-- 059_settlement_logs.sql: 不可变核销审计日志
-- 记录每一笔核销/反核销操作，只追加不删除，审计可追溯
-- Depends on: 003_invoice_payment.sql (payment_allocations), 046_advance_receipt_payment.sql (advance_allocations)

-- 核销方向枚举
-- direction: 'debit' = 债权减少（如发票被核销）, 'credit' = 债务减少（如付款被核销）

CREATE TABLE IF NOT EXISTS settlement_logs (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    -- 来源类型：payment_allocation | advance_allocation | manual_reversal
    source_type VARCHAR(30) NOT NULL,
    -- 来源记录 ID（指向 payment_allocations 或 advance_allocations）
    source_id UUID NOT NULL,
    -- 被核销单据类型：sales_invoice | ar_invoice | ap_invoice | advance_receipt | advance_payment
    doc_type VARCHAR(30) NOT NULL,
    -- 被核销单据 ID
    doc_id UUID NOT NULL,
    -- 核销方向：debit = 债权减少（outstanding 减少）, credit = 债务减少
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('debit', 'credit')),
    -- 核销金额（反核销时 amount 为正数，但 is_reversal = true）
    amount DECIMAL(18,2) NOT NULL CHECK (amount > 0),
    -- 核销前 outstanding 余额快照
    outstanding_before DECIMAL(18,2) NOT NULL,
    -- 核销后 outstanding 余额快照
    outstanding_after DECIMAL(18,2) NOT NULL,
    -- 是否为反核销（回滚）记录
    is_reversal BOOLEAN NOT NULL DEFAULT FALSE,
    -- 如果是反核销，指向被撤销的原始 settlement_log
    reversed_log_id UUID REFERENCES settlement_logs(id),
    -- 操作人
    created_by UUID,
    -- 创建时间（同时也是核销时间）
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引：按单据查询核销历史
CREATE INDEX IF NOT EXISTS idx_settlement_logs_doc ON settlement_logs(tenant_id, doc_type, doc_id);
-- 索引：按来源记录查询
CREATE INDEX IF NOT EXISTS idx_settlement_logs_source ON settlement_logs(tenant_id, source_type, source_id);
-- 索引：按时间排序审计
CREATE INDEX IF NOT EXISTS idx_settlement_logs_tenant_time ON settlement_logs(tenant_id, created_at DESC);

-- RLS
ALTER TABLE settlement_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON settlement_logs
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

COMMENT ON TABLE settlement_logs IS '不可变核销审计日志：记录每一笔核销/反核销操作，只追加不删除';
COMMENT ON COLUMN settlement_logs.direction IS '核销方向：debit=债权减少, credit=债务减少';
COMMENT ON COLUMN settlement_logs.is_reversal IS '是否为反核销回滚操作';
COMMENT ON COLUMN settlement_logs.reversed_log_id IS '如果是反核销，指向被撤销的原始日志';
