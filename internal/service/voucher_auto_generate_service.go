package service

import (
	"context"
	"fmt"
	"strings"
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
	busDocMappingRepo *repository.BusDocMappingRepository
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
	busDocMappingRepo *repository.BusDocMappingRepository,
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
		busDocMappingRepo: busDocMappingRepo,
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
		bankClearingAcctID = s.findAccountByCode(ctx, tenantID, "1002")
	}
	if bankClearingAcctID == nil {
		return nil, fmt.Errorf("bank account %s has no clearing_account_id and fallback account 1002 not found", txn.BankAccountID)
	}

	var debitAccountID, creditAccountID uuid.UUID
	var classification string
	if txn.Classification != nil {
		classification = *txn.Classification
	}

	// Determine accounts based on classification
	switch classification {
	case "bank_fee":
		acct := s.findAccountByCode(ctx, tenantID, "5602")
		if acct == nil {
			return nil, fmt.Errorf("account 5602 (财务费用) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	case "interest_income":
		acct := s.findAccountByCode(ctx, tenantID, "5602")
		if acct == nil {
			return nil, fmt.Errorf("account 5602 (财务费用) not found")
		}
		creditAccountID = *acct
		debitAccountID = *bankClearingAcctID
	case "tax_payment":
		acct := s.findAccountByCode(ctx, tenantID, "2221")
		if acct == nil {
			return nil, fmt.Errorf("account 2221 (应交税费) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	case "social_security":
		acct := s.findAccountByCode(ctx, tenantID, "2211")
		if acct == nil {
			return nil, fmt.Errorf("account 2211 (应付职工薪酬) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	case "insurance_fee":
		acct := s.findAccountByCode(ctx, tenantID, "5601")
		if acct == nil {
			return nil, fmt.Errorf("account 5601 (管理费用) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	case "business_receipt":
		acct := s.findAccountByCode(ctx, tenantID, "1122")
		if acct == nil {
			return nil, fmt.Errorf("account 1122 (应收账款) not found")
		}
		creditAccountID = *acct
		debitAccountID = *bankClearingAcctID
	case "business_payment":
		acct := s.findAccountByCode(ctx, tenantID, "2202")
		if acct == nil {
			return nil, fmt.Errorf("account 2202 (应付账款) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	default:
		acct := s.findAccountByCode(ctx, tenantID, "5601")
		if acct == nil {
			return nil, fmt.Errorf("account 5601 (管理费用) not found")
		}
		debitAccountID = *acct
		creditAccountID = *bankClearingAcctID
	}

	// Generate voucher number
	voucherResp, _ := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	// Build journal entry
	// Posting date is the voucher creation date (会计转凭证当日),
	// not the bank transaction date — accounting practice for 跨期调整.
	postingDate := time.Now()
	companyID := txn.CompanyID
	je := &model.JournalEntry{
		ID:               uuid.New(),
		TenantID:         tenantID,
		CompanyID:        companyID,
		VoucherNo:        voucherNo,
		PostingDate:      postingDate,
		CreatedBy:        createdBy,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DocStatus:        0,
		CounterpartyName: txn.CounterpartyName,
		Remark:           txn.Description,
	}

	// Build lines
	txnAmt := txn.Debit
	if txnAmt.IsZero() {
		txnAmt = txn.Credit
	}
	amount := txnAmt
	var lines []model.JournalEntryLine
	if txn.Debit.GreaterThan(decimal.Zero) {
		lines = []model.JournalEntryLine{
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: *bankClearingAcctID, Debit: amount, Credit: decimal.Zero},
			{ID: uuid.New(), JournalEntryID: je.ID, AccountID: creditAccountID, Debit: decimal.Zero, Credit: amount},
		}
	} else {
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
		ID:              uuid.New(),
		TenantID:        tenantID,
		CompanyID:       invoice.CompanyID,
		VoucherNo:       voucherNo,
		PostingDate:     time.Now(),
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		DocStatus:       0, // draft
		SourceType:      "invoice",
		SourceID:        invoiceID,
		SourceInvoiceID: invoiceID,
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

// GenerateFromPaymentEntry generates a voucher from a payment entry (收款单/付款单/费用/利息/转账).
// Account determination:
//  1. bus_doc_mapping by (doc_type, condition_key) — supports receipt/payment/expense/interest/transfer
//  2. Bank side overridden by actual bank clearing account or payment_method (cash→1001, bank→1002)
//  3. Fallback to hardcoded account codes if mapping not found
func (s *VoucherAutoGenerateService) GenerateFromPaymentEntry(ctx context.Context, tenantID, paymentID, userID uuid.UUID) (*model.JournalEntry, error) {
	pe, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("get payment entry: %w", err)
	}

	// Prevent duplicate voucher generation
	if pe.DocStatus >= 1 {
		return nil, fmt.Errorf("payment entry %s already has a voucher (docstatus=%d), regenerate rejected", paymentID, pe.DocStatus)
	}

	// Resolve the bank/cash side account based on payment_method.
	// payment_method: "bank"→1002, "cash"→1001, "wechat"/"alipay"→1002, default→1002
	paymentMethod := "bank"
	if pe.PaymentMethod != nil && *pe.PaymentMethod != "" {
		paymentMethod = *pe.PaymentMethod
	}

	var bankSideAcctID *uuid.UUID
	switch paymentMethod {
	case "cash":
		bankSideAcctID = s.findAccountByCode(ctx, tenantID, "1001")
		if bankSideAcctID == nil {
			bankSideAcctID = s.findAccountByCode(ctx, tenantID, "1002")
		}
	default:
		// bank, wechat, alipay, other — use bank clearing account
		if pe.BankAccountID != nil {
			bankAcct, repoErr := s.bankRepo.GetByID(ctx, tenantID, *pe.BankAccountID)
			if repoErr == nil && bankAcct != nil && bankAcct.ClearingAccountID != nil {
				bankSideAcctID = bankAcct.ClearingAccountID
			}
		}
		if bankSideAcctID == nil {
			bankSideAcctID = s.findAccountByCode(ctx, tenantID, "1002")
		}
		if bankSideAcctID == nil {
			// Final fallback: try 1001 (库存现金)
			bankSideAcctID = s.findAccountByCode(ctx, tenantID, "1001")
		}
	}
	if bankSideAcctID == nil {
		return nil, fmt.Errorf("cannot determine bank/cash account for payment entry %s (method=%s)", paymentID, paymentMethod)
	}

	amount := pe.PaidAmount
	if pe.PaymentType == "receive" && pe.ReceivedAmount != nil {
		amount = *pe.ReceivedAmount
	}

	// Try to resolve party from counterparty_name if PartyID is zero
	party, _ := s.partyRepo.GetByID(ctx, tenantID, pe.PartyID)
	if party == nil && pe.CounterpartyName != nil && *pe.CounterpartyName != "" {
		party, _ = s.partyRepo.GetByName(ctx, tenantID, *pe.CounterpartyName)
	}

	var partyName string
	if party != nil && party.Name != "" {
		partyName = party.Name
	} else if pe.CounterpartyName != nil && *pe.CounterpartyName != "" {
		partyName = *pe.CounterpartyName
	} else if pe.ReferenceNo != nil {
		partyName = *pe.ReferenceNo
	}

	direction := "out"
	if pe.PaymentType == "receive" {
		direction = "in"
	}

	// Resolve debit/credit accounts from bus_doc_mapping
	docType, conditionKey := s.detectDocTypeAndCondition(pe, partyName)
	debitAccountID, creditAccountID, err := s.resolveAccountsFromMapping(ctx, tenantID, docType, conditionKey, *bankSideAcctID)
	if err != nil {
		// Mapping lookup failed — use hardcoded fallback
		debitAccountID, creditAccountID, err = s.hardcodedFallback(ctx, tenantID, pe.PaymentType, pe.PartyType, *bankSideAcctID)
		if err != nil {
			return nil, err
		}
	}

	// Override the bank/cash side with the resolved account
	switch pe.PaymentType {
	case "receive":
		debitAccountID = *bankSideAcctID
	case "pay":
		creditAccountID = *bankSideAcctID
	case "expense":
		// expense: debit from mapping (费用科目), credit = bank/cash
		creditAccountID = *bankSideAcctID
	case "interest":
		// interest: debit = bank/cash, credit from mapping (财务费用)
		debitAccountID = *bankSideAcctID
	case "transfer":
		// transfer: both sides are bank accounts
		debitAccountID = *bankSideAcctID
		// credit stays from mapping (also 1002)
	}

	voucherResp, tmplErr := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if tmplErr == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}

	sourceDocType := "payment_entry"
	peID := pe.ID
	peNo := pe.PaymentNo
	remark := fmt.Sprintf("[%s] %s", pe.PaymentNo, partyName)

	je := &model.JournalEntry{
		ID:               uuid.New(),
		TenantID:         tenantID,
		CompanyID:        pe.CompanyID,
		VoucherNo:        voucherNo,
		// PostingDate = voucher creation date (会计转凭证当日), not the payment entry date.
		PostingDate:      time.Now(),
		CreatedBy:        userID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		DocStatus:        0,
		CounterpartyName: &partyName,
		SourceDocType:    &sourceDocType,
		SourceDocID:      &peID,
		SourceDocNo:      &peNo,
		Remark:           &remark,
	}

	// Use resolved party if available, otherwise keep original
	partyTypeStr := pe.PartyType
	partyIDCopy := pe.PartyID
	if party != nil && partyIDCopy == uuid.Nil {
		partyIDCopy = party.ID
		if party.PartyType != "" {
			partyTypeStr = party.PartyType
		}
	}

	verbZH := "收款"
	creditNounZH := "应收"
	switch pe.PaymentType {
	case "pay":
		verbZH = "付款"
		creditNounZH = "应付"
	case "expense":
		verbZH = "费用"
		creditNounZH = "支付"
	case "interest":
		verbZH = "利息"
		creditNounZH = "收入"
	case "transfer":
		verbZH = "转账"
		creditNounZH = "转出"
	}
	debitUserRemark := fmt.Sprintf("%s %s %s", verbZH, partyName, amount.String())
	creditUserRemark := fmt.Sprintf("%s%s", creditNounZH, partyName)

	lines := []model.JournalEntryLine{
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      debitAccountID,
			Debit:          amount,
			Credit:         decimal.Zero,
			PartyType:      &partyTypeStr,
			PartyID:        &partyIDCopy,
			UserRemark:     &debitUserRemark,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      creditAccountID,
			Debit:          decimal.Zero,
			Credit:         amount,
			PartyType:      &partyTypeStr,
			PartyID:        &partyIDCopy,
			UserRemark:     &creditUserRemark,
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
				allocDrRemark := fmt.Sprintf("核销 %s", invoice.InvoiceNo)
				allocCrRemark := fmt.Sprintf("应收/应付 %s", invoice.InvoiceNo)
				lines = append(lines, model.JournalEntryLine{
					ID:             uuid.New(),
					JournalEntryID: je.ID,
					AccountID:      *invRule.DebitAccountID,
					Debit:          alloc.AllocatedAmount,
					Credit:         decimal.Zero,
					PartyType:      &partyTypeStr,
					PartyID:        &partyIDCopy,
					UserRemark:     &allocDrRemark,
				}, model.JournalEntryLine{
					ID:             uuid.New(),
					JournalEntryID: je.ID,
					AccountID:      *invRule.CreditAccountID,
					Debit:          decimal.Zero,
					Credit:         alloc.AllocatedAmount,
					PartyType:      &partyTypeStr,
					PartyID:        &partyIDCopy,
					UserRemark:     &allocCrRemark,
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

	// Mark payment entry as posted to prevent duplicate voucher generation
	pe.DocStatus = 1
	pe.VoucherID = &je.ID
	pe.VoucherNo = &voucherNo
	if err = s.paymentRepo.Update(ctx, tenantID, pe); err != nil {
		return nil, fmt.Errorf("update payment entry docstatus: %w", err)
	}

	return je, nil
}

// hardcodedFallback provides hardcoded account resolution when bus_doc_mapping
// lookup fails. Uses payment_type + party_type to determine the correct accounts.
func (s *VoucherAutoGenerateService) hardcodedFallback(ctx context.Context, tenantID uuid.UUID, paymentType, partyType string, bankSideAcctID uuid.UUID) (debitAccountID, creditAccountID uuid.UUID, err error) {
	switch paymentType {
	case "receive":
		code := "1122" // 应收账款
		if partyType == "employee" {
			code = "1221" // 其他应收款
		}
		acct := s.findAccountByCode(ctx, tenantID, code)
		if acct == nil {
			acct = s.findAccountByCode(ctx, tenantID, "1122")
		}
		if acct == nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("hardcoded fallback: account %s not found", code)
		}
		return bankSideAcctID, *acct, nil

	case "pay":
		code := "2202" // 应付账款
		if partyType == "employee" {
			code = "2211" // 应付职工薪酬
		}
		acct := s.findAccountByCode(ctx, tenantID, code)
		if acct == nil {
			acct = s.findAccountByCode(ctx, tenantID, "2202")
		}
		if acct == nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("hardcoded fallback: account %s not found", code)
		}
		return *acct, bankSideAcctID, nil

	case "expense":
		acct := s.findAccountByCode(ctx, tenantID, "5601") // 管理费用
		if acct == nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("hardcoded fallback: expense account 5601 not found")
		}
		return *acct, bankSideAcctID, nil

	case "interest":
		acct := s.findAccountByCode(ctx, tenantID, "5602") // 财务费用
		if acct == nil {
			acct = s.findAccountByCode(ctx, tenantID, "5601")
		}
		if acct == nil {
			return uuid.Nil, uuid.Nil, fmt.Errorf("hardcoded fallback: interest account not found")
		}
		return bankSideAcctID, *acct, nil

	case "transfer":
		return bankSideAcctID, bankSideAcctID, nil

	default:
		return uuid.Nil, uuid.Nil, fmt.Errorf("unsupported payment_type %q for voucher generation", paymentType)
	}
}

// resolveCounterAccount determines the non-bank account for a payment voucher.
// Priority: 1) Party default AR/AP account  2) Classification rule match  3) party type + payment type mapping  4) Hardcoded 1122/2202.
func (s *VoucherAutoGenerateService) resolveCounterAccount(ctx context.Context, tenantID uuid.UUID, paymentType, partyType string, party *model.Party, refNo *string, partyName string, direction string) (uuid.UUID, error) {
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

	// Smart mapping: use party type + payment type to determine the right account code
	code := s.accountCodeByPartyType(paymentType, partyType)
	acct := s.findAccountByCode(ctx, tenantID, code)
	if acct != nil {
		return *acct, nil
	}

	// Final hardcoded fallback
	if paymentType == "receive" {
		code = "1122"
	} else {
		code = "2202"
	}
	acct = s.findAccountByCode(ctx, tenantID, code)
	if acct == nil {
		return uuid.Nil, fmt.Errorf("fallback account (code %s) not found; no party default or classification rule matched", code)
	}
	return *acct, nil
}

// accountCodeByPartyType returns the account code based on payment type and party type.
func (s *VoucherAutoGenerateService) accountCodeByPartyType(paymentType, partyType string) string {
	if paymentType == "receive" {
		switch partyType {
		case "customer":
			return "1122" // 应收账款
		case "employee":
			return "1221" // 其他应收款
		default:
			return "1122" // 应收账款（默认）
		}
	}
	// paymentType == "pay"
	switch partyType {
	case "supplier":
		return "2202" // 应付账款
	case "employee":
		return "2211" // 应付职工薪酬
	default:
		return "2202" // 应付账款（默认）
	}
}

// findAccountByCode safely looks up an account UUID by code. Returns nil if not found.
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

// detectDocTypeAndCondition returns (doc_type, condition_key) for looking up
// bus_doc_mapping. It maps pe.PaymentType to doc_type, and inspects the
// document description and counterparty for sub-scenarios like "tax" or "cash".
func (s *VoucherAutoGenerateService) detectDocTypeAndCondition(pe *model.PaymentEntry, partyName string) (string, string) {
	docType := normalizePaymentType(pe.PaymentType)

	desc := ""
	if pe.Description != nil {
		desc = *pe.Description
	}
	combined := strings.ToLower(desc + " " + partyName)

	// Detect tax scenario
	taxKeywords := []string{"国库", "金库", "财政", "税务局", "税款", "税务", "缴税", "增值税", "所得税", "附加税"}
	for _, kw := range taxKeywords {
		if strings.Contains(combined, strings.ToLower(kw)) {
			return docType, "tax"
		}
	}

	// Detect cash scenario
	if pe.PaymentMethod != nil && *pe.PaymentMethod == "cash" {
		return docType, "cash"
	}

	return docType, "default"
}

// normalizePaymentType maps internal payment_type values (receive/pay/expense/interest/transfer)
// to bus_doc_mapping doc_type values (receipt/payment/expense/interest/transfer).
func normalizePaymentType(pt string) string {
	switch pt {
	case "receive":
		return "receipt"
	case "pay":
		return "payment"
	default:
		return pt // expense, interest, transfer already match
	}
}

// resolveAccountsFromMapping looks up bus_doc_mapping and returns the
// resolved debit/credit account UUIDs, with the bankClearingID as the
// fallback for the bank-clearing side.
func (s *VoucherAutoGenerateService) resolveAccountsFromMapping(
	ctx context.Context, tenantID uuid.UUID, docType, conditionKey string, bankClearingID uuid.UUID,
) (debitAccountID, creditAccountID uuid.UUID, err error) {

	mapping, err := s.busDocMappingRepo.FindMapping(ctx, tenantID, docType, conditionKey)
	if err != nil || mapping == nil {
		mapping, err = s.busDocMappingRepo.FindDefaultMapping(ctx, tenantID, docType)
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("lookup bus_doc_mapping: %w", err)
	}
	if mapping == nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("no bus_doc_mapping for doc_type=%s, condition=%s", docType, conditionKey)
	}

	if mapping.DebitAccountID != nil {
		uid, perr := uuid.Parse(*mapping.DebitAccountID)
		if perr == nil {
			debitAccountID = uid
		}
	}
	if debitAccountID == uuid.Nil {
		if found := s.findAccountByCode(ctx, tenantID, mapping.DebitSubjectCode); found != nil {
			debitAccountID = *found
		}
	}
	if debitAccountID == uuid.Nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("bus_doc_mapping debit account %s not found", mapping.DebitSubjectCode)
	}

	if mapping.CreditAccountID != nil {
		uid, perr := uuid.Parse(*mapping.CreditAccountID)
		if perr == nil {
			creditAccountID = uid
		}
	}
	if creditAccountID == uuid.Nil {
		if found := s.findAccountByCode(ctx, tenantID, mapping.CreditSubjectCode); found != nil {
			creditAccountID = *found
		}
	}
	if creditAccountID == uuid.Nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("bus_doc_mapping credit account %s not found", mapping.CreditSubjectCode)
	}
	return debitAccountID, creditAccountID, nil
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
