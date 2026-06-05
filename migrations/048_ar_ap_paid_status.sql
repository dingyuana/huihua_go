-- Phase 3 Day 8: AR/AP 状态机扩充 + 回填
-- 1. 回填：confirmed + outstanding > 0 → partially_paid；confirmed + outstanding = 0 → paid
UPDATE ar_invoices
SET status = CASE
    WHEN outstanding_amount <= 0 THEN 'paid'
    WHEN outstanding_amount < amount THEN 'partially_paid'
    ELSE status
END
WHERE status = 'confirmed' AND outstanding_amount IS NOT NULL;

UPDATE ap_invoices
SET status = CASE
    WHEN outstanding_amount <= 0 THEN 'paid'
    WHEN outstanding_amount < amount THEN 'partially_paid'
    ELSE status
END
WHERE status = 'confirmed' AND outstanding_amount IS NOT NULL;

-- 2. 添加 CHECK 约束（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ar_invoices_status_check'
    ) THEN
        ALTER TABLE ar_invoices ADD CONSTRAINT ar_invoices_status_check
            CHECK (status IN ('draft','confirmed','partially_paid','paid','reversed'));
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'ap_invoices_status_check'
    ) THEN
        ALTER TABLE ap_invoices ADD CONSTRAINT ap_invoices_status_check
            CHECK (status IN ('draft','confirmed','partially_paid','paid','reversed'));
    END IF;
END$$;

-- 3. 索引：按状态过滤未结清 AR/AP
CREATE INDEX IF NOT EXISTS idx_ar_invoices_status_outstanding ON ar_invoices(tenant_id, status, outstanding_amount);
CREATE INDEX IF NOT EXISTS idx_ap_invoices_status_outstanding ON ap_invoices(tenant_id, status, outstanding_amount);
