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

type PaymentEntryService struct {
	repo        *repository.PaymentEntryRepository
	partyRepo   *repository.PartyRepository
	bankRepo    *repository.BankRepository
	accountRepo *repository.AccountRepository
	bankTxnRepo *repository.BankTransactionRepository
	reconSvc    *ReconciliationService
}

func NewPaymentEntryService(
	repo *repository.PaymentEntryRepository,
	partyRepo *repository.PartyRepository,
	bankRepo *repository.BankRepository,
	accountRepo *repository.AccountRepository,
	bankTxnRepo *repository.BankTransactionRepository,
) *PaymentEntryService {
	return &PaymentEntryService{
		repo:        repo,
		partyRepo:   partyRepo,
		bankRepo:    bankRepo,
		accountRepo: accountRepo,
		bankTxnRepo: bankTxnRepo,
	}
}

// InjectReconciliationService injects the reconciliation service.
func (s *PaymentEntryService) InjectReconciliationService(reconSvc *ReconciliationService) {
	s.reconSvc = reconSvc
}

type CreatePaymentFromBankTxnRequest struct {
	BankTransactionID uuid.UUID
	PaymentType       string
	PartyType         string
	PartyID           uuid.UUID
	CounterpartyName  *string
	PostingDate       time.Time
	ReferenceNo       string
}

func (s *PaymentEntryService) CreateFromBankTransaction(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	req *CreatePaymentFromBankTxnRequest,
	bankTxn *model.BankTransaction,
	companyID uuid.UUID,
) (*model.PaymentEntry, error) {

	if !repository.IsValidPaymentType(req.PaymentType) {
		return nil, fmt.Errorf("invalid payment_type %q (allowed: receive, pay, expense, interest, transfer)", req.PaymentType)
	}

	paymentNo, err := s.repo.GetNextPaymentNo(ctx, tenantID, req.PaymentType)
	if err != nil {
		return nil, fmt.Errorf("generate payment no: %w", err)
	}

	var amount decimal.Decimal
	var receivedAmt *decimal.Decimal
	switch req.PaymentType {
	case "receive", "interest":
		amount = bankTxn.Debit
		receivedAmt = &amount
	default:
		amount = bankTxn.Credit
	}

	var paidFromID, paidToID *uuid.UUID
	switch req.PaymentType {
	case "receive", "interest":
		paidToID = s.getClearingAccount(ctx, tenantID, bankTxn.BankAccountID)
	default:
		paidFromID = s.getClearingAccount(ctx, tenantID, bankTxn.BankAccountID)
	}

	paymentMethod := "bank"
	pe := &model.PaymentEntry{
		PaymentNo:        paymentNo,
		PaymentType:      req.PaymentType,
		PartyType:        req.PartyType,
		PartyID:          req.PartyID,
		CounterpartyName: req.CounterpartyName,
		PaidFromID:       paidFromID,
		PaidToID:         paidToID,
		PaidAmount:       amount,
		ReceivedAmount:   receivedAmt,
		ReferenceNo:      &req.ReferenceNo,
		ReferenceDate:    &bankTxn.TxnDate,
		PostingDate:      req.PostingDate,
		CompanyID:        companyID,
		BankAccountID:    &bankTxn.BankAccountID,
		DocStatus:        0,
		Description:      bankTxn.Description,
		PaymentMethod:    &paymentMethod,
		CreatedBy:        &userID,
	}

	entry, err := s.repo.Create(ctx, tenantID, pe)
	if err != nil {
		return nil, fmt.Errorf("create payment entry: %w", err)
	}

	_ = s.bankTxnRepo.SetMatchedPaymentEntry(ctx, tenantID, bankTxn.ID, entry.ID)
	return entry, nil
}

func (s *PaymentEntryService) getClearingAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID) *uuid.UUID {
	bank, err := s.bankRepo.GetByID(ctx, tenantID, bankAccountID)
	if err != nil || bank == nil {
		return nil
	}
	return bank.ClearingAccountID
}

func (s *PaymentEntryService) ListPaymentEntries(ctx context.Context, tenantID uuid.UUID) ([]model.PaymentEntry, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}

func (s *PaymentEntryService) ListByBankAccount(ctx context.Context, tenantID, bankAccountID uuid.UUID) ([]model.PaymentEntry, error) {
	return s.repo.ListByBankAccount(ctx, tenantID, bankAccountID)
}

func (s *PaymentEntryService) GetPaymentEntry(ctx context.Context, tenantID, id uuid.UUID) (*model.PaymentEntry, error) {
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *PaymentEntryService) UpdatePaymentEntry(ctx context.Context, tenantID uuid.UUID, pe *model.PaymentEntry) error {
	return s.repo.Update(ctx, tenantID, pe)
}

func (s *PaymentEntryService) DeletePaymentEntry(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.repo.Delete(ctx, tenantID, id)
}

func (s *PaymentEntryService) CreateFromBankTxn(ctx context.Context, tenantID, bankTxnID, userID uuid.UUID) (*model.PaymentEntry, error) {
	bankTxn, err := s.bankTxnRepo.GetByID(ctx, tenantID, bankTxnID)
	if err != nil {
		return nil, fmt.Errorf("get bank transaction: %w", err)
	}

	if bankTxn.Classification == nil {
		return nil, fmt.Errorf("bank transaction has no classification")
	}

	var paymentType, partyType string
	classification := *bankTxn.Classification
	switch classification {
	case "business_receipt":
		paymentType = "receive"
		partyType = "customer"
	case "business_payment":
		paymentType = "pay"
		partyType = "supplier"
	case "internal_transfer":
		paymentType = "transfer"
		partyType = "internal"
	default:
		return nil, fmt.Errorf("invalid classification %q for payment entry", classification)
	}

	var referenceNo string
	if bankTxn.ReferenceNo != nil {
		referenceNo = *bankTxn.ReferenceNo
	}

	req := &CreatePaymentFromBankTxnRequest{
		BankTransactionID: bankTxnID,
		PaymentType:       paymentType,
		PartyType:         partyType,
		PartyID:           uuid.Nil,
		PostingDate:       time.Now(),
		ReferenceNo:       referenceNo,
		CounterpartyName:  bankTxn.CounterpartyName,
	}

	return s.CreateFromBankTransaction(ctx, tenantID, userID, req, bankTxn, bankTxn.CompanyID)
}

// Approve submits and approves a payment entry, triggering automatic invoice reconciliation.
func (s *PaymentEntryService) Approve(ctx context.Context, tenantID, paymentID, userID uuid.UUID) (*model.ReconciliationPair, error) {
	// 1. Load PaymentEntry, verify DocStatus == 1 (submitted)
	pe, err := s.repo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("load payment entry: %w", err)
	}
	if pe.DocStatus != 1 {
		return nil, fmt.Errorf("payment entry must be in submitted status (docstatus=1), got %d", pe.DocStatus)
	}

	// 2. Update DocStatus to 2 (approved)
	if err := s.repo.UpdateStatus(ctx, tenantID, paymentID, 2); err != nil {
		return nil, fmt.Errorf("update docstatus: %w", err)
	}

	// 3. Call reconciliation service
	if s.reconSvc == nil {
		return nil, nil // reconciliation service not injected — skip
	}
	pair, err := s.reconSvc.ReconcilePaymentEntry(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("reconcile payment entry: %w", err)
	}
	return pair, nil
}
