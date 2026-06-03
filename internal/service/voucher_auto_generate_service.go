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
	paymentRepo       *repository.PaymentEntryRepository
	partyRepo         *repository.PartyRepository
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
	paymentRepo *repository.PaymentEntryRepository,
	partyRepo *repository.PartyRepository,
	accountRepo *repository.AccountRepository,
	classificationSvc *ClassificationRuleService,
	templateSvc *VoucherTemplateService,
	approvalSvc *ApprovalService,
) *VoucherAutoGenerateService {
	return &VoucherAutoGenerateService{
		journalRepo: journalRepo, glRepo: glRepo,
		bankTxnRepo: bankTxnRepo, bankRepo: bankRepo,
		invoiceRepo:       invoiceRepo,
		paymentRepo:       paymentRepo,
		partyRepo:         partyRepo,
		accountRepo:       accountRepo,
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

	if txn.Matched {
		return nil, fmt.Errorf("bank transaction %s already has a journal entry", txnID)
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

// GenerateFromPaymentEntry generates a voucher from a payment entry (收款单/付款单).
// Account determination priority:
//  1. Party's configured default account (ar_account_id / ap_account_id)
//  2. Classification rule matching (matched on payment counterparty / description)
//  3. Hardcoded fallback: 应收账款(1122) for receive, 应付账款(2202) for pay
//
// When payment is allocated to invoices, additional detail lines are added
// using classification rule matching on each invoice.
func (s *VoucherAutoGenerateService) GenerateFromPaymentEntry(ctx context.Context, tenantID, paymentID, userID uuid.UUID) (*model.JournalEntry, error) {
	pe, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get payment entry: %w", err)
	}

	var bankClearingAcctID *uuid.UUID
	if pe.BankAccountID != nil {
		bankAcct, repoErr := s.bankRepo.GetByID(ctx, tenantID, *pe.BankAccountID)
		if repoErr == nil && bankAcct != nil {
			bankClearingAcctID = bankAcct.ClearingAccountID
		}
	}
	if bankClearingAcctID == nil {
		bankClearingAcctID = s.findAccountByCode(ctx, tenantID, "1002")
	}
	if bankClearingAcctID == nil {
		return nil, fmt.Errorf("cannot determine bank clearing account for payment entry %s", paymentID)
	}

	amount := pe.PaidAmount
	if pe.PaymentType == "receive" && pe.ReceivedAmount != nil {
		amount = *pe.ReceivedAmount
	}

	party, _ := s.partyRepo.GetByID(ctx, tenantID, pe.PartyID)
	partyName := ""
	if party != nil {
		partyName = party.Name
	} else if pe.CounterpartyName != nil {
		partyName = *pe.CounterpartyName
	}

	direction := "out"
	if pe.PaymentType == "receive" {
		direction = "in"
	}

	counterAcctID, err := s.resolveCounterAccount(ctx, tenantID, pe.PaymentType, party, pe.ReferenceNo, partyName, direction)
	if err != nil {
		return nil, err
	}

	var debitAccountID, creditAccountID uuid.UUID
	switch pe.PaymentType {
	case "receive":
		debitAccountID = *bankClearingAcctID
		creditAccountID = counterAcctID
	case "pay":
		debitAccountID = counterAcctID
		creditAccountID = *bankClearingAcctID
	default:
		return nil, fmt.Errorf("payment type %q cannot generate voucher directly", pe.PaymentType)
	}

	voucherResp, tmplErr := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if tmplErr == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	je := &model.JournalEntry{
		ID:          uuid.New(),
		TenantID:    tenantID,
		CompanyID:   pe.CompanyID,
		VoucherNo:   voucherNo,
		PostingDate: pe.PostingDate,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		DocStatus:   0,
	}

	partyTypeStr := pe.PartyType
	partyIDCopy := pe.PartyID

	lines := []model.JournalEntryLine{
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      debitAccountID,
			Debit:          amount,
			Credit:         decimal.Zero,
			PartyType:      &partyTypeStr,
			PartyID:        &partyIDCopy,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      creditAccountID,
			Debit:          decimal.Zero,
			Credit:         amount,
			PartyType:      &partyTypeStr,
			PartyID:        &partyIDCopy,
		},
	}

	allocs, allocErr := s.invoiceRepo.GetAllocationsByPaymentEntry(ctx, tenantID, paymentID)
	if allocErr == nil && len(allocs) > 0 {
		for _, alloc := range allocs {
			invoice, invErr := s.invoiceRepo.GetByID(ctx, tenantID, alloc.InvoiceID)
			if invErr != nil || invoice == nil {
				continue
			}
			invResult, matchErr := s.classificationSvc.MatchTransaction(ctx, tenantID,
				invoice.InvoiceNo+" "+invoice.InvoiceType, partyName, direction)
			if matchErr != nil || !invResult.Matched || invResult.RuleID == nil {
				continue
			}
			invRule, ruleErr := s.classificationSvc.GetRuleByID(ctx, tenantID, *invResult.RuleID)
			if ruleErr != nil || invRule == nil {
				continue
			}
			if invRule.DebitAccountID != nil && invRule.CreditAccountID != nil {
				lines = append(lines, model.JournalEntryLine{
					ID:             uuid.New(),
					JournalEntryID: je.ID,
					AccountID:      *invRule.DebitAccountID,
					Debit:          alloc.AllocatedAmount,
					Credit:         decimal.Zero,
					PartyType:      &partyTypeStr,
					PartyID:        &partyIDCopy,
				}, model.JournalEntryLine{
					ID:             uuid.New(),
					JournalEntryID: je.ID,
					AccountID:      *invRule.CreditAccountID,
					Debit:          decimal.Zero,
					Credit:         alloc.AllocatedAmount,
					PartyType:      &partyTypeStr,
					PartyID:        &partyIDCopy,
				})
			}
		}
	}

	if _, err = s.journalRepo.Create(ctx, tenantID, je); err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}
	if _, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines); err != nil {
		return nil, fmt.Errorf("add journal entry lines: %w", err)
	}

	return je, nil
}

// resolveCounterAccount determines the non-bank account for a payment voucher.
// Priority: 1) Party default AR/AP account  2) Classification rule match  3) Hardcoded 1122/2202.
func (s *VoucherAutoGenerateService) resolveCounterAccount(ctx context.Context, tenantID uuid.UUID, paymentType string, party *model.Party, refNo *string, partyName string, direction string) (uuid.UUID, error) {
	if party != nil {
		if paymentType == "receive" && party.ArAccountID != nil {
			return *party.ArAccountID, nil
		}
		if paymentType == "pay" && party.ApAccountID != nil {
			return *party.ApAccountID, nil
		}
	}

	desc := ""
	if refNo != nil {
		desc = *refNo
	}
	result, matchErr := s.classificationSvc.MatchTransaction(ctx, tenantID, desc, partyName, direction)
	if matchErr == nil && result.Matched && result.RuleID != nil {
		rule, ruleErr := s.classificationSvc.GetRuleByID(ctx, tenantID, *result.RuleID)
		if ruleErr == nil && rule != nil {
			if paymentType == "receive" && rule.CreditAccountID != nil {
				return *rule.CreditAccountID, nil
			}
			if paymentType == "pay" && rule.DebitAccountID != nil {
				return *rule.DebitAccountID, nil
			}
		}
	}

	var code string
	if paymentType == "receive" {
		code = "1122"
	} else {
		code = "2202"
	}
	acct := s.findAccountByCode(ctx, tenantID, code)
	if acct == nil {
		return uuid.Nil, fmt.Errorf("fallback account (code %s) not found; no party default or classification rule matched", code)
	}
	return *acct, nil
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
