-- 026_cash_management.sql
-- Extend bank_accounts to also represent cash accounts (库存现金)
-- A "cash account" is a bank_account with is_cash=true. The clearing_account
-- is conventionally 1001 库存现金.

BEGIN;

ALTER TABLE bank_accounts
    ADD COLUMN IF NOT EXISTS is_cash     BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS custodian   VARCHAR(100),
    ADD COLUMN IF NOT EXISTS location    VARCHAR(200);

CREATE INDEX IF NOT EXISTS idx_bank_accounts_cash
    ON bank_accounts (tenant_id, company_id) WHERE is_cash = TRUE;

-- Seed a default cash account for each existing company that doesn't have one yet.
-- Uses 1001 库存现金 as the GL clearing account.
DO $$
DECLARE
    comp RECORD;
    cash_gl_id UUID;
BEGIN
    FOR comp IN
        SELECT DISTINCT company_id, tenant_id
        FROM bank_accounts
        WHERE is_cash = FALSE
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM bank_accounts
            WHERE company_id = comp.company_id AND is_cash = TRUE
        ) THEN
            SELECT id INTO cash_gl_id FROM accounts
            WHERE tenant_id = comp.tenant_id AND code = '1001' LIMIT 1;
            INSERT INTO bank_accounts (
                id, bank_name, account_number, clearing_account_id,
                company_id, tenant_id, currency, bank_account_type,
                is_cash, custodian, location, is_active
            ) VALUES (
                gen_random_uuid(),
                '库存现金',
                'CASH-' || substr(md5(random()::text), 1, 8),
                cash_gl_id,
                comp.company_id,
                comp.tenant_id,
                'CNY',
                'cash',
                TRUE,
                '出纳',
                '财务部保险柜',
                TRUE
            );
        END IF;
    END LOOP;
END $$;

COMMIT;
