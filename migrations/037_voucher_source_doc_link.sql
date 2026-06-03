-- 凭证与单据双向关联：journal_entries 加来源单据字段，payment_entries 加 voucher_id
ALTER TABLE journal_entries
  ADD COLUMN IF NOT EXISTS source_doc_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS source_doc_id UUID,
  ADD COLUMN IF NOT EXISTS source_doc_no VARCHAR(50);

ALTER TABLE payment_entries
  ADD COLUMN IF NOT EXISTS voucher_id UUID,
  ADD COLUMN IF NOT EXISTS voucher_no VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_journal_entries_source_doc
  ON journal_entries(tenant_id, source_doc_type, source_doc_id);

CREATE INDEX IF NOT EXISTS idx_payment_entries_voucher
  ON payment_entries(tenant_id, voucher_id);

-- 回填历史数据：从 bank_transactions.matched_gl_entry_id 反推
UPDATE journal_entries je
SET source_doc_type = 'bank_txn',
    source_doc_id = bt.id,
    source_doc_no = bt.reference_no
FROM bank_transactions bt
WHERE bt.matched_gl_entry_id = je.id
  AND je.source_doc_type IS NULL;

-- 回填 payment_entries 的 voucher_id
UPDATE payment_entries pe
SET voucher_id = bt.matched_gl_entry_id
FROM bank_transactions bt
WHERE bt.matched_payment_entry_id = pe.id
  AND bt.matched_gl_entry_id IS NOT NULL
  AND pe.voucher_id IS NULL;
