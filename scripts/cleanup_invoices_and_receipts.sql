-- ============================================================================
-- 清理脚本：清空发票和应收收款单相关数据
-- 执行前请确保已备份数据！
-- ============================================================================

-- 开始事务
BEGIN;

-- ============================================================================
-- 1. 清空发票行项目表（依赖于 sales_invoices）
-- ============================================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_line_items') THEN
        DELETE FROM invoice_line_items;
    END IF;
END $$;

-- ============================================================================
-- 2. 清空支付分配表（依赖于 sales_invoices 和 payment_entries）
-- ============================================================================
DELETE FROM payment_allocations;

-- ============================================================================
-- 3. 清空应收发票表（依赖于 sales_invoices）
-- ============================================================================
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'ar_invoices') THEN
        DELETE FROM ar_invoices;
    END IF;
END $$;

-- ============================================================================
-- 4. 清空销售发票主表
-- ============================================================================
DELETE FROM sales_invoices;

-- ============================================================================
-- 5. 清空收款单（payment_entries）
-- ============================================================================
DELETE FROM payment_entries;

-- ============================================================================
-- 提交事务
-- ============================================================================
COMMIT;

-- ============================================================================
-- 验证清理结果
-- ============================================================================
SELECT 'sales_invoices' AS table_name, COUNT(*) AS row_count FROM sales_invoices
UNION ALL
SELECT 'payment_entries', COUNT(*) FROM payment_entries
UNION ALL
SELECT 'payment_allocations', COUNT(*) FROM payment_allocations
UNION ALL
SELECT 'invoice_line_items', COUNT(*) FROM invoice_line_items
UNION ALL
SELECT 'ar_invoices', COUNT(*) FROM ar_invoices;
