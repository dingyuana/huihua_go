BEGIN;

-- 清空凭证分录
DELETE FROM journal_entry_lines;

-- 清空凭证主表
DELETE FROM journal_entries;

-- 清空GL分录
DELETE FROM gl_entries;

-- 清空银行流水
DELETE FROM bank_transactions;

-- 清空对账明细
DELETE FROM bank_reconciliation_details;

-- 清空对账报表
DELETE FROM bank_reconciliation_statements;

-- 清空收款/付款单
DELETE FROM payment_entries;

-- 重置序列（可选）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.sequences WHERE sequence_name = 'journal_entries_id_seq') THEN
        PERFORM setval('journal_entries_id_seq', 1, false);
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.sequences WHERE sequence_name = 'payment_entries_id_seq') THEN
        PERFORM setval('payment_entries_id_seq', 1, false);
    END IF;
END $$;

COMMIT;

-- 验证删除结果
SELECT 'journal_entry_lines' AS table_name, COUNT(*) AS row_count FROM journal_entry_lines
UNION ALL
SELECT 'journal_entries' AS table_name, COUNT(*) AS row_count FROM journal_entries
UNION ALL
SELECT 'gl_entries' AS table_name, COUNT(*) AS row_count FROM gl_entries
UNION ALL
SELECT 'bank_transactions' AS table_name, COUNT(*) AS row_count FROM bank_transactions
UNION ALL
SELECT 'bank_reconciliation_details' AS table_name, COUNT(*) AS row_count FROM bank_reconciliation_details
UNION ALL
SELECT 'bank_reconciliation_statements' AS table_name, COUNT(*) AS row_count FROM bank_reconciliation_statements
UNION ALL
SELECT 'payment_entries' AS table_name, COUNT(*) AS row_count FROM payment_entries;
