-- 063_journal_bank_link.sql: 给 journal_entries 加 bank_transaction_id 列
-- 修复缺陷 #9: reconciliation_items (实际是 ListUnmatched journals) 查询
-- bank_transaction_id 列缺失导致 reconciliation 500 错误

ALTER TABLE journal_entries
    ADD COLUMN IF NOT EXISTS bank_transaction_id UUID REFERENCES bank_transactions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_journal_entries_bank_txn ON journal_entries(bank_transaction_id)
    WHERE bank_transaction_id IS NOT NULL;

COMMENT ON COLUMN journal_entries.bank_transaction_id IS '关联的银行流水 ID:用于银企对账的反向查找';
