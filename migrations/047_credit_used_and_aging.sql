-- Phase 2 Day 6: 信用管控 + 账龄分析
ALTER TABLE parties ADD COLUMN IF NOT EXISTS credit_used DECIMAL(18,2) DEFAULT 0;
ALTER TABLE parties ADD COLUMN IF NOT EXISTS credit_overdraft_days INT DEFAULT 0;
ALTER TABLE parties ADD COLUMN IF NOT EXISTS last_credit_check_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN parties.credit_used IS '已占用信用额度（AR 确认时累加，核销/预收时释放）';
COMMENT ON COLUMN parties.credit_overdraft_days IS '允许的信用超期天数，0=严格不允许超期';

CREATE INDEX IF NOT EXISTS idx_parties_credit_overdue ON parties(tenant_id) WHERE credit_used > 0;

UPDATE parties SET credit_used = 0 WHERE credit_used IS NULL;
