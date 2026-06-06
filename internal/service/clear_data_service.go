package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClearDataService erases tenant-scoped data. Destructive; gated by
// app.mode != "production" at the handler layer.
//
// Two scopes:
//   - ClearBusinessData  wipes transactional data (vouchers, invoices, bank
//     transactions, payments, allocations, advances, opening balances, AI
//     feedback). Master data (accounts, parties, company_settings, bank
//     accounts, accounting periods, voucher templates, classification rules,
//     approval flows, assets, budgets, exchange rates) is preserved.
//   - ClearBasicInfo    wipes master data (company info, bank info, chart of
//     accounts, parties, periods, templates, rules, approvals, assets,
//     budgets, exchange rates). REQUIRES ClearBusinessData to have run
//     first, otherwise RESTRICT FKs from journal_entry_lines, payment_entries,
//     bank_transactions, etc. will abort the transaction.
//
// NEVER clears: tenants, users, audit_logs, standard_accounts_seed (global
// reference data, not tenant-scoped).
type ClearDataService struct {
	pool *pgxpool.Pool
}

func NewClearDataService(pool *pgxpool.Pool) *ClearDataService {
	return &ClearDataService{pool: pool}
}

// ClearResult maps table name → rows deleted.
type ClearResult map[string]int64

// clearStmt is one DELETE step. $1 binds tenantID; child tables without
// tenant_id use a subquery on their parent's tenant_id.
type clearStmt struct {
	name string
	sql  string
}

// businessDataStmts: children before parents. All scoped to the tenant.
var businessDataStmts = []clearStmt{
	// ── invoice subtree ────────────────────────────────────────────
	{"invoice_line_items", `DELETE FROM invoice_line_items WHERE invoice_id IN (SELECT id FROM sales_invoices WHERE tenant_id = $1)`},
	{"ar_invoices", `DELETE FROM ar_invoices WHERE tenant_id = $1`},
	{"ap_invoices", `DELETE FROM ap_invoices WHERE tenant_id = $1`},
	{"sales_invoices", `DELETE FROM sales_invoices WHERE tenant_id = $1`},
	// ── payment subtree ────────────────────────────────────────────
	{"payment_allocations", `DELETE FROM payment_allocations WHERE tenant_id = $1`},
	{"payment_entries", `DELETE FROM payment_entries WHERE tenant_id = $1`},
	// ── advance subtree ────────────────────────────────────────────
	{"advance_allocations", `DELETE FROM advance_allocations WHERE tenant_id = $1`},
	{"advance_payments", `DELETE FROM advance_payments WHERE tenant_id = $1`},
	{"advance_receipts", `DELETE FROM advance_receipts WHERE tenant_id = $1`},
	// ── bank reconciliation & transactions ────────────────────────
	{"bank_reconciliation_details", `DELETE FROM bank_reconciliation_details WHERE tenant_id = $1`},
	{"bank_reconciliation_statements", `DELETE FROM bank_reconciliation_statements WHERE tenant_id = $1`},
	{"bank_transactions", `DELETE FROM bank_transactions WHERE tenant_id = $1`},
	{"bank_balance_adjustments", `DELETE FROM bank_balance_adjustments WHERE tenant_id = $1`},
	// ── gl + voucher subtree ───────────────────────────────────────
	{"gl_entries", `DELETE FROM gl_entries WHERE tenant_id = $1`},
	{"journal_entry_lines", `DELETE FROM journal_entry_lines WHERE tenant_id = $1`},
	{"journal_entries", `DELETE FROM journal_entries WHERE tenant_id = $1`},
	{"voucher_state_transitions", `DELETE FROM voucher_state_transitions WHERE tenant_id = $1`},
	// ── reconciliation pairs (referenced by bank_recon_details) ───
	{"unreconciled_items", `DELETE FROM unreconciled_items WHERE tenant_id = $1`},
	{"reconciliation_records", `DELETE FROM reconciliation_records WHERE tenant_id = $1`},
	{"reconciliation_pairs", `DELETE FROM reconciliation_pairs WHERE tenant_id = $1`},
	// ── depreciation schedules (reference journal_entries) ────────
	{"depreciation_schedules", `DELETE FROM depreciation_schedules WHERE tenant_id = $1`},
	// ── misc transactional ─────────────────────────────────────────
	{"opening_balances", `DELETE FROM opening_balances WHERE tenant_id = $1`},
	{"ai_feedback_logs", `DELETE FROM ai_feedback_logs WHERE tenant_id = $1`},
}

// basicInfoStmts: master/setting data. Requires business data cleared first.
var basicInfoStmts = []clearStmt{
	// ── voucher templates (children first via subquery) ───────────
	{"voucher_template_lines", `DELETE FROM voucher_template_lines WHERE template_id IN (SELECT id FROM voucher_templates WHERE tenant_id = $1)`},
	{"voucher_templates", `DELETE FROM voucher_templates WHERE tenant_id = $1`},
	{"voucher_numbering_rules", `DELETE FROM voucher_numbering_rules WHERE tenant_id = $1`},
	// ── approval flows (approval_tasks auto-cleared via CASCADE) ──
	{"approval_flows", `DELETE FROM approval_flows WHERE tenant_id = $1`},
	// ── budgets (children first; budget_distributions via subquery) ─
	{"budget_distributions", `DELETE FROM budget_distributions WHERE budget_account_id IN (SELECT id FROM budget_accounts WHERE tenant_id = $1)`},
	{"budget_accounts", `DELETE FROM budget_accounts WHERE tenant_id = $1`},
	{"budget_control_configs", `DELETE FROM budget_control_configs WHERE tenant_id = $1`},
	{"budgets", `DELETE FROM budgets WHERE tenant_id = $1`},
	// ── assets & categories ────────────────────────────────────────
	{"assets", `DELETE FROM assets WHERE tenant_id = $1`},
	{"asset_categories", `DELETE FROM asset_categories WHERE tenant_id = $1`},
	{"depreciation_runs", `DELETE FROM depreciation_runs WHERE tenant_id = $1`},
	// ── classification & mapping rules ─────────────────────────────
	{"classification_rules", `DELETE FROM classification_rules WHERE tenant_id = $1`},
	{"classification_rules_old", `DELETE FROM classification_rules_old WHERE tenant_id = $1`},
	{"bus_doc_mapping", `DELETE FROM bus_doc_mapping WHERE tenant_id = $1`},
	// ── parties, bank accounts (parents of business-data FKs) ──────
	{"parties", `DELETE FROM parties WHERE tenant_id = $1`},
	{"bank_accounts", `DELETE FROM bank_accounts WHERE tenant_id = $1`},
	// bank_balance_adjustments auto-cleared via ON DELETE CASCADE on bank_accounts
	// ── chart of accounts (parent of most FKs) ────────────────────
	{"accounts", `DELETE FROM accounts WHERE tenant_id = $1`},
	// ── period / settings / rates ──────────────────────────────────
	{"accounting_periods", `DELETE FROM accounting_periods WHERE tenant_id = $1`},
	{"exchange_rates", `DELETE FROM exchange_rates WHERE tenant_id = $1`},
	{"company_settings", `DELETE FROM company_settings WHERE tenant_id = $1`},
	// standard_accounts_seed is intentionally NOT cleared: global seed, no tenant_id.
}

// ClearBusinessData deletes all transactional data for the tenant.
func (s *ClearDataService) ClearBusinessData(ctx context.Context, tenantID uuid.UUID) (ClearResult, error) {
	return s.runDelete(ctx, tenantID, businessDataStmts)
}

// ClearBasicInfo deletes all master/setting data for the tenant.
// Caller must run ClearBusinessData first.
func (s *ClearDataService) ClearBasicInfo(ctx context.Context, tenantID uuid.UUID) (ClearResult, error) {
	return s.runDelete(ctx, tenantID, basicInfoStmts)
}

func (s *ClearDataService) runDelete(ctx context.Context, tenantID uuid.UUID, stmts []clearStmt) (ClearResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	result := make(ClearResult, len(stmts))
	for _, s := range stmts {
		ct, err := tx.Exec(ctx, s.sql, tenantID)
		if err != nil {
			return nil, fmt.Errorf("delete %s: %w", s.name, err)
		}
		result[s.name] = ct.RowsAffected()
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return result, nil
}
