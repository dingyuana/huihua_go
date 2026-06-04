-- Migration: 048_audit_approval_fields
-- Purpose: P1 - Post-approval audit trail
-- Adds approved_by / approved_at fields to ar_invoices and journal_entries
-- to record who approved/posted each document and when.

BEGIN;

-- ar_invoices: record approver and approval time
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;

-- journal_entries: record approver and approval time (used when docstatus transitions 0→1)
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP WITH TIME ZONE;

-- Index for querying by approver
CREATE INDEX IF NOT EXISTS idx_ar_invoices_approved_by ON ar_invoices(approved_by) WHERE approved_by IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_journal_entries_approved_by ON journal_entries(approved_by) WHERE approved_by IS NOT NULL;

COMMIT;