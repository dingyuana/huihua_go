-- ApInvoice 并发控制字段
-- TASK-3.2: 给 ap_invoices 添加 locked_at / locked_by 字段，与 ArInvoice 对齐
-- 用于并发核销时锁定/解锁 ApInvoice，防止多个用户同时操作同一张应付单
-- 注意：PostgreSQL 不支持 ON UPDATE CURRENT_TIMESTAMP（之前踩过坑），应用层维护时间戳

-- 幂等迁移：使用 IF NOT EXISTS 防止重复执行报错
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ;
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS locked_by UUID;

-- 索引：便于按锁定状态查询/清理
CREATE INDEX IF NOT EXISTS idx_ap_invoices_locked
  ON ap_invoices(tenant_id, locked_at) WHERE locked_at IS NOT NULL;
