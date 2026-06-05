package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

func TestSubmitReview_AC5(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	txnID := seedBankTxn(ctx, pool, t, &bankTxnSeed{
		tenantID:       tenantID,
		status:         string(model.BankTxnReviewStatusClassified),
		classification: ptrString("bank_fee"),
		debit:          decimal.NewFromFloat(50.00),
		credit:         decimal.Zero,
		direction:      ptrString("out"),
	})

	svc := newBankTxnReviewService(pool)

	results, err := svc.SubmitReview(ctx, tenantID, []string{txnID.String()}, nil)
	if err != nil {
		t.Fatalf("SubmitReview failed: %v", err)
	}
	if len(results.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results.Results))
	}
	result := results.Results[0]
	if result.Outcome != "voucher_generated" {
		t.Errorf("expected outcome voucher_generated, got %q", result.Outcome)
	}
	if result.VoucherID == nil {
		t.Error("expected VoucherID to be set")
	}

	txn := getBankTxnByID(ctx, pool, t, txnID)
	if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusVoucherGenerated) {
		t.Errorf("expected status voucher_generated, got %v", txn.Status)
	}
	if !txn.Matched {
		t.Error("expected matched=true")
	}
}

func TestSubmitReview_AC6(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Classification that will fail because the account lookup returns nil
	txnID := seedBankTxn(ctx, pool, t, &bankTxnSeed{
		tenantID:       tenantID,
		status:         string(model.BankTxnReviewStatusClassified),
		classification: ptrString("__invalid_class__"),
		debit:          decimal.NewFromFloat(999.00),
		credit:         decimal.Zero,
		direction:      ptrString("out"),
	})

	svc := newBankTxnReviewService(pool)

	_, err := svc.SubmitReview(ctx, tenantID, []string{txnID.String()}, nil)
	if err == nil {
		t.Fatal("expected SubmitReview to fail for missing account, but it succeeded")
	}

	txn := getBankTxnByID(ctx, pool, t, txnID)
	if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusClassified) {
		t.Errorf("expected status to remain classified after rollback, got %v", txn.Status)
	}
}

func TestRejectManual_AC7(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	txnID := seedBankTxn(ctx, pool, t, &bankTxnSeed{
		tenantID:       tenantID,
		status:         string(model.BankTxnReviewStatusClassified),
		classification: ptrString("bank_fee"),
		debit:          decimal.NewFromFloat(50.00),
		credit:         decimal.Zero,
		direction:      ptrString("out"),
	})

	svc := newBankTxnReviewService(pool)

	err := svc.RejectManual(ctx, []string{txnID.String()})
	if err != nil {
		t.Fatalf("RejectManual failed: %v", err)
	}

	txn := getBankTxnByID(ctx, pool, t, txnID)
	if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusManualPending) {
		t.Errorf("expected status manual_pending, got %v", txn.Status)
	}
}

func TestPreviewDraft_AC3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	ctx := context.Background()
	pool := testPool(t)
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	txnID := seedBankTxn(ctx, pool, t, &bankTxnSeed{
		tenantID:       tenantID,
		status:         string(model.BankTxnReviewStatusClassified),
		classification: ptrString("bank_fee"),
		debit:          decimal.NewFromFloat(50.00),
		credit:         decimal.Zero,
		direction:      ptrString("out"),
	})

	svc := newBankTxnReviewService(pool)

	voucher, err := svc.PreviewDraft(ctx, tenantID, txnID)
	if err != nil {
		t.Fatalf("PreviewDraft failed: %v", err)
	}
	if voucher == nil {
		t.Fatal("expected draft voucher, got nil")
	}
	if voucher.DocStatus != 0 {
		t.Errorf("expected docstatus=0 (draft), got %d", voucher.DocStatus)
	}

	txn := getBankTxnByID(ctx, pool, t, txnID)
	if txn.Status == nil || *txn.Status != string(model.BankTxnReviewStatusClassified) {
		t.Errorf("expected status to remain classified after preview, got %v", txn.Status)
	}
}

// ---------------------------------------------------------------------------
// Test helpers

type bankTxnSeed struct {
	tenantID       uuid.UUID
	status         string
	classification *string
	debit          decimal.Decimal
	credit         decimal.Decimal
	direction      *string
}

func seedBankTxn(ctx context.Context, pool *pgxpool.Pool, t *testing.T, s *bankTxnSeed) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now()
	bankAccountID := uuid.MustParse("7cd2c7fd-68c5-419f-be60-579c0f1a7b4b")
	companyID := uuid.MustParse("3586c914-5eb4-426a-9d20-f297a10b147c")
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Use a random description suffix to avoid unique constraint violations
	// when reference_no is null/empty (idx_bank_txn_unique_no_ref)
	desc := fmt.Sprintf("test description %s", id.String())

	_, err := pool.Exec(ctx, `
		INSERT INTO bank_transactions
			(id, tenant_id, bank_account_id, company_id, txn_date, description,
			 debit, credit, direction, classification, status, matched, confirmed,
			 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		id, tenantID, bankAccountID, companyID, now, desc,
		s.debit, s.credit, s.direction, s.classification, s.status, false, false,
		now, now)
	if err != nil {
		t.Fatalf("seed bank_txn: %v", err)
	}
	return id
}

func getBankTxnByID(ctx context.Context, pool *pgxpool.Pool, t *testing.T, id uuid.UUID) *model.BankTransaction {
	t.Helper()
	var txn model.BankTransaction
	err := pool.QueryRow(ctx, `
		SELECT id, tenant_id, bank_account_id, txn_date, description,
			   debit, credit, direction, classification, matched, confirmed,
			   status, matched_payment_entry_id, matched_gl_entry_id,
			   imported_from, raw_data, company_id, created_at
		FROM bank_transactions WHERE id = $1`,
		id).Scan(
		&txn.ID, &txn.TenantID, &txn.BankAccountID, &txn.TxnDate, &txn.Description,
		&txn.Debit, &txn.Credit, &txn.Direction, &txn.Classification,
		&txn.Matched, &txn.Confirmed, &txn.Status,
		&txn.MatchedPaymentEntryID, &txn.MatchedGLEntryID,
		&txn.ImportedFrom, &txn.RawData, &txn.CompanyID, &txn.CreatedAt,
	)
	if err != nil {
		t.Fatalf("getBankTxnByID: %v", err)
	}
	return &txn
}

func ptrString(s string) *string { return &s }

func newBankTxnReviewService(pool *pgxpool.Pool) *BankTxnReviewService {
	repo := repository.NewBankTransactionRepository(pool)
	bankRepo := repository.NewBankRepository(pool)
	journalRepo := repository.NewJournalRepository(pool)
	glRepo := repository.NewGLEntryRepository(pool)
	invoiceRepo := repository.NewInvoiceRepository(pool)
	paymentRepo := repository.NewPaymentEntryRepository(pool)
	partyRepo := repository.NewPartyRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	busDocMappingRepo := repository.NewBusDocMappingRepository(pool)
	voucherTemplateRepo := repository.NewVoucherTemplateRepository(pool)
	arRepo := repository.NewArInvoiceRepository(pool)
	apRepo := repository.NewApInvoiceRepository(pool)
	advanceReceiptRepo := repository.NewAdvanceReceiptRepository(pool)
	advancePaymentRepo := repository.NewAdvancePaymentRepository(pool)
	approvalRepo := repository.NewApprovalRepository(pool)

	classificationSvc := NewClassificationRuleService(repository.NewClassificationRuleRepository(pool))
	templateSvc := NewVoucherTemplateService(voucherTemplateRepo, accountRepo)
	approvalSvc := NewApprovalService(approvalRepo, journalRepo)

	voucherSvc := NewVoucherAutoGenerateService(
		journalRepo, glRepo, repo, bankRepo,
		invoiceRepo, arRepo, apRepo, paymentRepo, partyRepo, accountRepo,
		busDocMappingRepo, advanceReceiptRepo, advancePaymentRepo,
		classificationSvc, templateSvc, approvalSvc,
	)

	paymentSvc := NewPaymentEntryService(paymentRepo, partyRepo, bankRepo, accountRepo, repo)

	return NewBankTxnReviewService(pool, repo, voucherSvc, paymentSvc)
}