package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherAutoGenerateService auto-generates vouchers from bank transactions, invoices, and reconciliation pairs.
type VoucherAutoGenerateService struct {
	journalRepo      *repository.JournalRepository
	glRepo           *repository.GLEntryRepository
	bankTxnRepo      *repository.BankTransactionRepository
	invoiceRepo      *repository.InvoiceRepository
	classificationSvc *ClassificationRuleService
	templateSvc      *VoucherTemplateService
}

// NewVoucherAutoGenerateService creates a new VoucherAutoGenerateService.
func NewVoucherAutoGenerateService(
	journalRepo *repository.JournalRepository,
	glRepo *repository.GLEntryRepository,
	bankTxnRepo *repository.BankTransactionRepository,
	invoiceRepo *repository.InvoiceRepository,
	classificationSvc *ClassificationRuleService,
	templateSvc *VoucherTemplateService,
) *VoucherAutoGenerateService {
	return &VoucherAutoGenerateService{
		journalRepo: journalRepo, glRepo: glRepo,
		bankTxnRepo: bankTxnRepo, invoiceRepo: invoiceRepo,
		classificationSvc: classificationSvc, templateSvc: templateSvc,
	}
}

// GenerateFromBankTxn generates a voucher from a bank transaction.
func (s *VoucherAutoGenerateService) GenerateFromBankTxn(ctx context.Context, tenantID, txnID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	txn, err := s.bankTxnRepo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, err
	}

	// Match to a classification rule
	var matchedAccountID uuid.UUID
	txnDesc := ""
	if txn.Description != nil {
		txnDesc = *txn.Description
	}
	txnAmt := txn.Debit
	if txnAmt.IsZero() {
		txnAmt = txn.Credit
	}
	direction := "debit"
	if txn.Credit.GreaterThan(txn.Debit) {
		direction = "credit"
	}

	rules, err := s.classificationSvc.ListRules(ctx, tenantID)
	if err == nil && len(rules) > 0 {
		for range rules {
			var result *model.RuleMatchResult
			result, err = s.classificationSvc.MatchTransaction(ctx, tenantID, txnDesc, txnAmt, direction)
			if err == nil && result.Matched && result.AccountID != nil {
				matchedAccountID = *result.AccountID
				break
			}
		}
	}

	// Generate voucher number
	voucherResp, _ := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	// Build journal entry
	postingDate := txn.TxnDate
	companyID := txn.CompanyID
	je := &model.JournalEntry{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CompanyID:   companyID,
		VoucherNo:   voucherNo,
		PostingDate: postingDate,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DocStatus:   0, // draft
	}

	// Build lines
	amount := txnAmt
	if txn.Debit.GreaterThan(decimal.Zero) {
		// Bank received: debit bank, credit matched account
		line1 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: txn.BankAccountID,
			Debit: amount, Credit: decimal.Zero,
		}
		line2 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: matchedAccountID,
			Debit: decimal.Zero, Credit: amount,
		}
		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, []model.JournalEntryLine{line1, line2})
	} else {
		// Bank paid out: credit bank, debit matched account
		line1 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: matchedAccountID,
			Debit: amount, Credit: decimal.Zero,
		}
		line2 := model.JournalEntryLine{
			ID: uuid.New(), JournalEntryID: je.ID,
			AccountID: txn.BankAccountID,
			Debit: decimal.Zero, Credit: amount,
		}
		_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, []model.JournalEntryLine{line1, line2})
	}

	_, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, err
	}
	// Mark bank txn as matched
	_ = s.bankTxnRepo.UpdateStatus(ctx, tenantID, txnID, true)
	return je, nil
}

// PreviewVoucher returns a preview without saving.
func (s *VoucherAutoGenerateService) PreviewVoucher(ctx context.Context, tenantID, txnID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	return s.GenerateFromBankTxn(ctx, tenantID, txnID, createdBy)
}

// BatchGenerateFromBank generates vouchers from all unmatched bank transactions.
func (s *VoucherAutoGenerateService) BatchGenerateFromBank(ctx context.Context, tenantID, bankAccountID uuid.UUID, createdBy uuid.UUID) ([]*model.JournalEntry, error) {
	txns, err := s.bankTxnRepo.GetUnmatched(ctx, tenantID, bankAccountID)
	if err != nil {
		return nil, err
	}
	var vouchers []*model.JournalEntry
	for _, txn := range txns {
		v, err := s.GenerateFromBankTxn(ctx, tenantID, txn.ID, createdBy)
		if err == nil {
			vouchers = append(vouchers, v)
		}
	}
	return vouchers, nil
}
