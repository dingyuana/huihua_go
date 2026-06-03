-- Add counterparty_name column to journal_entries for voucher counterparty display
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS counterparty_name VARCHAR(255);

-- Update existing journal entries that have bank transaction matches
UPDATE journal_entries je
SET counterparty_name = bt.counterparty_name
FROM bank_transactions bt
WHERE bt.matched_gl_entry_id = je.id
  AND je.counterparty_name IS NULL;
