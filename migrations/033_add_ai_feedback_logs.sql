-- 033_add_ai_feedback_logs.sql: Create AI feedback logs table for AC10
-- Part of: TASK-BANK-01 §3.1
-- Depends on: 014_bank_transactions.sql, 032_add_bank_txn_status.sql

-- ai_feedback_logs: records every human action on a bank transaction
-- so the AI system can learn from corrections.
CREATE TABLE IF NOT EXISTS ai_feedback_logs (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    bank_txn_id     uuid NOT NULL,
    ai_suggested_action  text,
    ai_confidence        int,
    ai_business_scene    text,
    human_action        text NOT NULL,
    human_modified_fields jsonb,
    created_by      uuid,
    created_at      timestamp NOT NULL DEFAULT NOW()
);

-- Index for querying logs by bank transaction (used by ListByTxnID)
CREATE INDEX IF NOT EXISTS idx_ai_feedback_logs_bank_txn_id ON ai_feedback_logs(bank_txn_id);

-- Index for querying logs by tenant (common access pattern)
CREATE INDEX IF NOT EXISTS idx_ai_feedback_logs_tenant_id ON ai_feedback_logs(tenant_id);

-- Index for time-based queries (AI learning pipelines)
CREATE INDEX IF NOT EXISTS idx_ai_feedback_logs_created_at ON ai_feedback_logs(created_at);