-- ============================================================================
-- 清理脚本：清空凭证相关数据
-- 执行前请确保已备份数据！
-- ============================================================================

BEGIN;

-- 1. 清空凭证分录表
DELETE FROM journal_entry_lines;

-- 2. 清空凭证主表
DELETE FROM journal_entries;

-- 3. 清空GL分录表
DELETE FROM gl_entries;

COMMIT;

-- 验证清理结果
SELECT 'journal_entry_lines' AS table_name, COUNT(*) AS row_count FROM journal_entry_lines
UNION ALL
SELECT 'journal_entries', COUNT(*) FROM journal_entries
UNION ALL
SELECT 'gl_entries', COUNT(*) FROM gl_entries;
