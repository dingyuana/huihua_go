ALTER TABLE write_off_records 
    ALTER COLUMN status TYPE VARCHAR(30),
    ALTER COLUMN status SET DEFAULT 'draft',
    ADD COLUMN approver UUID,
    ADD COLUMN approved_at TIMESTAMP,
    ADD COLUMN reject_reason TEXT;

UPDATE write_off_records SET status = 'approved' WHERE status = '1';
UPDATE write_off_records SET status = 'reversed' WHERE status = '0';

CREATE INDEX idx_write_off_records_approver ON write_off_records(approver);