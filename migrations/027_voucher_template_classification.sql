-- 027_voucher_template_classification.sql
-- Bind voucher templates to classification rules so auto-generation can pick the right template.

BEGIN;

ALTER TABLE voucher_templates
    ADD COLUMN IF NOT EXISTS classification VARCHAR(50);

CREATE UNIQUE INDEX IF NOT EXISTS uq_voucher_template_classification
    ON voucher_templates (tenant_id, classification)
    WHERE classification IS NOT NULL AND is_active = TRUE;

COMMIT;
