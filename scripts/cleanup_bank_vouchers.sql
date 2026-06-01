-- ============================================================================
-- 清理脚本：清空银行流水和凭证数据
-- 执行前请确保已备份数据！
-- ============================================================================

-- 开始事务
BEGIN;

-- ============================================================================
-- 1. 清空凭证相关数据（注意外键依赖顺序）
-- ============================================================================

-- 先清空凭证分录表（依赖于 journal_entries）
DELETE FROM journal_entry_lines;

-- 清空凭证主表
DELETE FROM journal_entries;

-- 清空GL分录表
DELETE FROM gl_entries;

-- ============================================================================
-- 2. 清空银行流水相关数据
-- ============================================================================

-- 清空银行对账明细表（依赖于 bank_transactions）
DELETE FROM bank_reconciliation_details;

-- 清空银行对账报表表
DELETE FROM bank_reconciliation_statements;

-- 清空银行流水表
DELETE FROM bank_transactions;

-- ============================================================================
-- 3. 清空核销记录（如果存在）
-- ============================================================================

-- 检查并清空核销记录表（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'reconciliation_records') THEN
        DELETE FROM reconciliation_records;
    END IF;
END $$;

-- 检查并清空发票核销表（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_payments') THEN
        DELETE FROM invoice_payments;
    END IF;
END $$;

-- ============================================================================
-- 4. 重置序列（确保新插入的数据ID从1开始）
-- ============================================================================

-- 注意：UUID类型不需要重置序列
-- 如果使用SERIAL/BIGSERIAL类型的ID，可以使用以下语句重置：
-- SELECT setval('table_id_seq', 1, false);

-- ============================================================================
-- 提交事务
-- ============================================================================

COMMIT;

-- ============================================================================
-- 验证清理结果
-- ============================================================================

SELECT 'journal_entry_lines' as table_name, COUNT(*) as row_count FROM journal_entry_lines
UNION ALL
SELECT 'journal_entries', COUNT(*) FROM journal_entries
UNION ALL
SELECT 'gl_entries', COUNT(*) FROM gl_entries
UNION ALL
SELECT 'bank_transactions', COUNT(*) FROM bank_transactions
UNION ALL
SELECT 'bank_reconciliation_details', COUNT(*) FROM bank_reconciliation_details
UNION ALL
SELECT 'bank_reconciliation_statements', COUNT(*) FROM bank_reconciliation_statements;
