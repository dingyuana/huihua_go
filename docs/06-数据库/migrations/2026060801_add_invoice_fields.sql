-- TASK-2.5: 销售发票增强 - 数电票字段 + 部分红冲
-- 执行时间: 2026-06-08
-- 执行前请备份数据库

ALTER TABLE sales_invoices
  ADD COLUMN invoice_kind VARCHAR(20) DEFAULT NULL COMMENT '纸票/数电普票/数电专票: paper_special/paper_normal/electronic_special/electronic_normal',
  ADD COLUMN electronic_url VARCHAR(500) DEFAULT NULL COMMENT '数电发票版式文件URL',
  ADD COLUMN red_letter_info_id VARCHAR(50) DEFAULT NULL COMMENT '红字信息表编号（关联开具）',
  ADD COLUMN red_letter_reason VARCHAR(200) DEFAULT NULL COMMENT '开具红字发票原因',
  ADD COLUMN original_invoice_id VARCHAR(36) DEFAULT NULL COMMENT '原蓝字发票ID（红冲时填）',
  ADD COLUMN is_part_red BOOLEAN DEFAULT FALSE COMMENT '是否部分红冲',
  ADD COLUMN red_amount DECIMAL(18,2) DEFAULT NULL COMMENT '红冲金额（部分红冲时填写）',
  ADD COLUMN tax_authority_code VARCHAR(20) DEFAULT NULL COMMENT '主管税务机关代码',
  ADD COLUMN confirm_status VARCHAR(20) DEFAULT 'unconfirmed' COMMENT '确认状态: unconfirmed/confirmed/invalid',
  ADD COLUMN confirm_date DATE DEFAULT NULL COMMENT '确认日期';

-- 回滚（如需撤销）
-- ALTER TABLE sales_invoices
--   DROP COLUMN IF EXISTS invoice_kind,
--   DROP COLUMN IF EXISTS electronic_url,
--   DROP COLUMN IF EXISTS red_letter_info_id,
--   DROP COLUMN IF EXISTS red_letter_reason,
--   DROP COLUMN IF EXISTS original_invoice_id,
--   DROP COLUMN IF EXISTS is_part_red,
--   DROP COLUMN IF EXISTS red_amount,
--   DROP COLUMN IF EXISTS tax_authority_code,
--   DROP COLUMN IF EXISTS confirm_status,
--   DROP COLUMN IF EXISTS confirm_date;