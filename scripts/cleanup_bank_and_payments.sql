-- ============================================================================
-- 清理脚本：清空银行流水和收款付款单
-- 执行前请确保已备份数据！
-- ============================================================================

BEGIN;

-- 1. 清空银行对账明细表（依赖于 bank_transactions）
DELETE FROM bank_reconciliation_details;

-- 2. 清空银行对账报表表
DELETE FROM bank_reconciliation_statements;

-- 3. 清空支付分配表（依赖于 payment_entries）
DELETE FROM payment_allocations;

-- 4. 清空收款付款单
DELETE FROM payment_entries;

-- 5. 清空银行流水
DELETE FROM bank_transactions;

COMMIT;

-- 验证清理结果
SELECT 'bank_transactions' AS table_name, COUNT(*) AS row_count FROM bank_transactions
UNION ALL
SELECT 'payment_entries', COUNT(*) FROM payment_entries
UNION ALL
SELECT 'payment_allocations', COUNT(*) FROM payment_allocations
UNION ALL
SELECT 'bank_reconciliation_details', COUNT(*) FROM bank_reconciliation_details
UNION ALL
SELECT 'bank_reconciliation_statements', COUNT(*) FROM bank_reconciliation_statements;
