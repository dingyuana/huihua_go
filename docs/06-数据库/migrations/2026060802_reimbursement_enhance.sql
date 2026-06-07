-- 报销单字段扩展
ALTER TABLE bus_reimbursements
  ADD COLUMN reject_reason VARCHAR(500) DEFAULT NULL,
  ADD COLUMN sub_expense_type VARCHAR(50) DEFAULT NULL,
  ADD COLUMN updated_at TIMESTAMP DEFAULT NULL;

-- 附件表
CREATE TABLE IF NOT EXISTS reimbursement_attachments (
  id VARCHAR(36) PRIMARY KEY,
  reimbursement_id VARCHAR(36) NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_path VARCHAR(500) NOT NULL,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(100),
  uploaded_by VARCHAR(36),
  uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_reimbursement_id (reimbursement_id)
);

-- 报销单-发票关联表
CREATE TABLE IF NOT EXISTS reimbursement_invoice_links (
  id VARCHAR(36) PRIMARY KEY,
  reimbursement_id VARCHAR(36) NOT NULL,
  invoice_id VARCHAR(36) NOT NULL,
  invoice_type VARCHAR(20) NOT NULL DEFAULT 'expense_invoice',
  linked_amount DECIMAL(18,2) NOT NULL,
  linked_by VARCHAR(36),
  linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_reimbursement_id (reimbursement_id),
  UNIQUE KEY uk_reim_invoice (reimbursement_id, invoice_id)
);
