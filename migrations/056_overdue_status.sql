-- Migration 056: 应收单 (ar_invoices) 和应付单 (ap_invoices) 新增 "已逾期" (overdue) 状态
--
-- 说明:
--   ar_invoices.status 和 ap_invoices.status 字段是 TEXT 类型（或 VARCHAR），
--   不需要 ALTER TABLE 即可接受新枚举值 'overdue'。
--   本迁移仅用于文档化新增的枚举值。
--
-- 涉及常量（internal/model）:
--   ArInvoiceStatusOverdue ArInvoiceStatus = "overdue"
--   ApInvoiceStatusOverdue ApInvoiceStatus = "overdue"
--
-- 状态机说明:
--   overdue 是一个由系统自动派生（automatic transition）的状态。
--   触发条件: due_date < CURRENT_DATE AND outstanding_amount > 0
--             AND status IN ('confirmed', 'partially_paid')
--   退出条件: 收/付款完成 → status 转为 'paid'
--             部分收/付款 → status 保持或回到 'partially_paid'
--             反向冲销  → status 转为 'reversed'
--
--   overdue 不接受手动 confirm/approve —— 它是 derived state，不在手动状态机中。
--   手动转换只能从 'confirmed' / 'partially_paid' 起始，结束于 'paid' / 'reversed'。
--
-- 建议的 cron 任务（暂未实现，框架待定）:
--   -- 每日 00:30 跑一次
--   UPDATE ar_invoices
--      SET status = 'overdue', updated_at = NOW()
--    WHERE due_date < CURRENT_DATE
--      AND outstanding_amount > 0
--      AND status IN ('confirmed', 'partially_paid');
--
--   UPDATE ap_invoices
--      SET status = 'overdue', updated_at = NOW()
--    WHERE due_date < CURRENT_DATE
--      AND outstanding_amount > 0
--      AND status IN ('confirmed', 'partially_paid');
--
--   对应的应用层代码在 internal/cron 或 internal/jobs 包下新增
--   MarkOverdueInvoices(ctx, db) 之类的实现，目前不实现。
--
-- 已存在的枚举值（按 status 字段顺序）:
--   ar_invoices.status:  draft, confirmed, partially_paid, paid, reversed, overdue
--   ap_invoices.status:  draft, confirmed, partially_paid, paid, reversed, overdue
--
-- 此迁移为 NOOP（不修改 schema），由应用层在 enum 常量里支持新值。

-- 1) 文档化 ar_invoices.status 允许的取值（注释形式，不影响行为）
COMMENT ON COLUMN ar_invoices.status IS
  '应收单状态: draft | confirmed | partially_paid | paid | reversed | overdue';

-- 2) 文档化 ap_invoices.status 允许的取值（注释形式，不影响行为）
COMMENT ON COLUMN ap_invoices.status IS
  '应付单状态: draft | confirmed | partially_paid | paid | reversed | overdue';
