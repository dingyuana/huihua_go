package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherAutoGenerateService auto-generates vouchers from bank transactions, invoices, and reconciliation pairs.
type VoucherAutoGenerateService struct {
	journalRepo       *repository.JournalRepository
	glRepo            *repository.GLEntryRepository
	bankTxnRepo       *repository.BankTransactionRepository
	bankRepo          *repository.BankRepository
	invoiceRepo       *repository.InvoiceRepository
	accountRepo       *repository.AccountRepository
	classificationSvc *ClassificationRuleService
	templateSvc       *VoucherTemplateService
	approvalSvc       *ApprovalService
}

func NewVoucherAutoGenerateService(
	journalRepo *repository.JournalRepository,
	glRepo *repository.GLEntryRepository,
	bankTxnRepo *repository.BankTransactionRepository,
	bankRepo *repository.BankRepository,
	invoiceRepo *repository.InvoiceRepository,
	accountRepo *repository.AccountRepository,
	classificationSvc *ClassificationRuleService,
	templateSvc *VoucherTemplateService,
	approvalSvc *ApprovalService,
) *VoucherAutoGenerateService {
	return &VoucherAutoGenerateService{
		journalRepo: journalRepo, glRepo: glRepo,
		bankTxnRepo: bankTxnRepo, bankRepo: bankRepo,
		invoiceRepo: invoiceRepo,
		accountRepo: accountRepo,
		classificationSvc: classificationSvc, templateSvc: templateSvc,
		approvalSvc: approvalSvc,
	}
}

// GenerateFromBankTxn generates a voucher from a bank transaction.
func (s *VoucherAutoGenerateService) GenerateFromBankTxn(ctx context.Context, tenantID, txnID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	txn, err := s.bankTxnRepo.GetByID(ctx, tenantID, txnID)
	if err != nil {
		return nil, err
	}

	bankAcct, err := s.bankRepo.GetByID(ctx, tenantID, txn.BankAccountID)
	if err != nil {
		return nil, fmt.Errorf("lookup bank account: %w", err)
	}
	bankClearingAcctID := bankAcct.ClearingAccountID
	if bankClearingAcctID == nil {
		acct, lookupErr := s.accountRepo.GetByCode(ctx, tenantID, "1002")
		if lookupErr != nil || acct == nil {
			return nil, fmt.Errorf("bank account %s has no clearing_account_id and fallback account 1002 not found", txn.BankAccountID)
		}
		bankClearingAcctID = &acct.ID
	}

	var debitAccountID, creditAccountID uuid.UUID
	var classification string
	if txn.Classification != nil {
		classification = *txn.Classification
	}

	// Determine accounts based on classification
	switch classification {
	case "bank_fee":
		// 银行手续费 → 财务费用 (5602)
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "5602")
		creditAccountID = *bankClearingAcctID
	case "interest_income":
		// 利息收入 → 财务费用 (5602) 贷方，或者投资收益
		creditAccountID = *s.findAccountByCode(ctx, tenantID, "5602")
		debitAccountID = *bankClearingAcctID
	case "tax_payment":
		// 税务缴费 → 应交税费 (2221)
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "2221")
		creditAccountID = *bankClearingAcctID
	case "social_security":
		// 社保缴费 → 应付职工薪酬 (2211)
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "2211")
		creditAccountID = *bankClearingAcctID
	case "insurance_fee":
		// 保险费用 → 管理费用 (5601)
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "5601")
		creditAccountID = *bankClearingAcctID
	case "business_receipt":
		// 业务收款 → 应收账款 (1122) 贷方，或者主营业务收入
		creditAccountID = *s.findAccountByCode(ctx, tenantID, "1122")
		debitAccountID = *bankClearingAcctID
	case "business_payment":
		// 业务付款 → 应付账款 (2202)
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "2202")
		creditAccountID = *bankClearingAcctID
	default:
		// 默认用管理费用
		debitAccountID = *s.findAccountByCode(ctx, tenantID, "5601")
		creditAccountID = *bankClearingAcctID
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
		DocStatus:   0, // 始终为草稿状态，需要人工审核
	}

	// Build lines
	txnAmt := txn.Debit
	if txnAmt.IsZero() {
		txnAmt = txn.Credit
	}
	amount := txnAmt
	var lines []model.JournalEntryLine
	if txn.Debit.GreaterThan(decimal.Zero) {
		// 收入：借银行，贷对方科目
		lines = []model.JournalEntryLine{
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: *bankClearingAcctID, Debit: amount, Credit: decimal.Zero},
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: creditAccountID, Debit: decimal.Zero, Credit: amount},
		}
	} else {
		// 支出：借对方科目，贷银行
		lines = []model.JournalEntryLine{
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: debitAccountID, Debit: amount, Credit: decimal.Zero},
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: *bankClearingAcctID, Debit: decimal.Zero, Credit: amount},
		}
	}
	_, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, err
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
	if err != nil {
		return nil, err
	}

	_ = s.bankTxnRepo.MarkAsMatched(ctx, tenantID, []uuid.UUID{txnID}, je.ID)
	return je, nil
}

// GenerateFromInvoice generates a voucher from a sales or purchase invoice.
func (s *VoucherAutoGenerateService) GenerateFromInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID, createdBy uuid.UUID) (*model.JournalEntry, error) {
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, err
	}

	voucherResp, _ := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	je := &model.JournalEntry{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CompanyID:   invoice.CompanyID,
		VoucherNo:   voucherNo,
		PostingDate: invoice.PostingDate,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DocStatus:   0, // draft
	}

	var lines []model.JournalEntryLine
	isSales := invoice.InvoiceType == "sales" || invoice.InvoiceType == "output"
	invoiceType := "sales"
	if !isSales {
		invoiceType = "purchase"
	}

	var counterDr, counterCr uuid.UUID
	txnDesc := invoice.InvoiceNo + " " + invoiceType
	counterpartyStr := ""
	direction := "in"
	if !isSales {
		direction = "out"
	}
	result, err := s.classificationSvc.MatchTransaction(ctx, tenantID, txnDesc, counterpartyStr, direction)
	if err == nil && result.Matched && result.RuleID != nil {
		if r, lookupErr := s.classificationSvc.GetRuleByID(ctx, tenantID, *result.RuleID); lookupErr == nil {
			if r.DebitAccountID != nil {
				counterDr = *r.DebitAccountID
			}
			if r.CreditAccountID != nil {
				counterCr = *r.CreditAccountID
			}
		}
	}

	if isSales {
		arAccountID := s.findAccountByCode(ctx, tenantID, "1122")
		if arAccountID != nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: *arAccountID,
				Debit:     invoice.TotalAmount, Credit: decimal.Zero,
			})
		}
		if counterCr != uuid.Nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: counterCr,
				Debit:     decimal.Zero, Credit: invoice.TotalAmount,
			})
		}
	} else {
		if counterDr != uuid.Nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: counterDr,
				Debit:     invoice.TotalAmount, Credit: decimal.Zero,
			})
		}
		apAccountID := s.findAccountByCode(ctx, tenantID, "2202")
		if apAccountID != nil {
			lines = append(lines, model.JournalEntryLine{
				ID: uuid.New(), JournalEntryID: je.ID,
				AccountID: *apAccountID,
				Debit:     decimal.Zero, Credit: invoice.TotalAmount,
			})
		}
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid accounts found for invoice %s", invoice.InvoiceNo)
	}

	if _, err := s.journalRepo.Create(ctx, tenantID, je); err != nil {
		return nil, err
	}
	if _, err := s.journalRepo.AddLines(ctx, tenantID, je.ID, lines); err != nil {
		return nil, err
	}

	return je, nil
}

func (s *VoucherAutoGenerateService) findAccountByCode(ctx context.Context, tenantID uuid.UUID, code string) *uuid.UUID {
	acct, err := s.accountRepo.GetByCode(ctx, tenantID, code)
	if err != nil || acct == nil {
		return nil
	}
	return &acct.ID
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
