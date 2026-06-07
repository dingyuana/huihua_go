package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// ReimbursementService handles reimbursement business logic.
type ReimbursementService struct {
	reimbRepo    *repository.ReimbursementRepository
	journalRepo  *repository.JournalRepository
	accountRepo  *repository.AccountRepository
	templateSvc  *VoucherTemplateService
}

// NewReimbursementService creates a new ReimbursementService.
func NewReimbursementService(
	reimbRepo *repository.ReimbursementRepository,
	journalRepo *repository.JournalRepository,
	accountRepo *repository.AccountRepository,
	templateSvc *VoucherTemplateService,
) *ReimbursementService {
	return &ReimbursementService{
		reimbRepo:   reimbRepo,
		journalRepo: journalRepo,
		accountRepo: accountRepo,
		templateSvc: templateSvc,
	}
}

// CreateReimbursementRequest is the request body for creating a reimbursement.
type CreateReimbursementRequest struct {
	EmployeeName   string  `json:"employee_name"`
	Department     *string `json:"department,omitempty"`
	ExpenseType    string  `json:"expense_type"`
	SubExpenseType *string `json:"sub_expense_type,omitempty"`
	Amount         string  `json:"amount"`
	PostingDate    string  `json:"posting_date"`
	Description *string `json:"description,omitempty"`
	BankAccount   *string `json:"bank_account,omitempty"`
	CompanyID     string  `json:"company_id"`
}

// CreateReimbursement creates a new reimbursement (docstatus=0 draft).
func (s *ReimbursementService) CreateReimbursement(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID, req *CreateReimbursementRequest) (*model.BusReimbursement, error) {
	reimbNo, err := s.reimbRepo.GetNextReimbursementNo(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get next reimbursement no: %w", err)
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		return nil, fmt.Errorf("invalid posting_date: %w", err)
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return nil, fmt.Errorf("invalid company_id: %w", err)
	}

	reimbursement := &model.BusReimbursement{
		ID:              uuid.New(),
		ReimbursementNo: reimbNo,
		EmployeeName:    req.EmployeeName,
		Department:      req.Department,
		ExpenseType:     req.ExpenseType,
		SubExpenseType:  req.SubExpenseType,
		Amount:          amount,
		PostingDate:     postingDate,
		Description:     req.Description,
		BankAccount:     req.BankAccount,
		DocStatus:       0,
		CreatedBy:       &userID,
		TenantID:        tenantID,
		CompanyID:       companyID,
	}

	return s.reimbRepo.Create(ctx, tenantID, reimbursement)
}

// GetReimbursement retrieves a reimbursement by ID.
func (s *ReimbursementService) GetReimbursement(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) (*model.BusReimbursement, error) {
	return s.reimbRepo.GetByID(ctx, tenantID, id)
}

// ListReimbursements lists reimbursements with optional status filter.
func (s *ReimbursementService) ListReimbursements(ctx context.Context, tenantID uuid.UUID, status *int16) ([]model.BusReimbursement, error) {
	return s.reimbRepo.ListByTenant(ctx, tenantID, status)
}

// UpdateReimbursement updates a reimbursement (only allowed for docstatus=0 draft).
func (s *ReimbursementService) UpdateReimbursement(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, req *CreateReimbursementRequest) error {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if reimb.DocStatus != 0 {
		return errors.New("only draft reimbursements can be updated")
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	postingDate, err := time.Parse("2006-01-02", req.PostingDate)
	if err != nil {
		return fmt.Errorf("invalid posting_date: %w", err)
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return fmt.Errorf("invalid company_id: %w", err)
	}

	reimb.EmployeeName = req.EmployeeName
	reimb.Department = req.Department
	reimb.ExpenseType = req.ExpenseType
	reimb.Amount = amount
	reimb.PostingDate = postingDate
	reimb.Description = req.Description
	reimb.BankAccount = req.BankAccount
	reimb.CompanyID = companyID

	return s.reimbRepo.Update(ctx, tenantID, reimb)
}

// DeleteReimbursement deletes a reimbursement (only allowed for docstatus=0 draft).
func (s *ReimbursementService) DeleteReimbursement(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if reimb.DocStatus != 0 {
		return errors.New("only draft reimbursements can be deleted")
	}
	return s.reimbRepo.Delete(ctx, tenantID, id)
}

// Submit submits a reimbursement: changes docstatus from 0 to 1.
func (s *ReimbursementService) Submit(ctx context.Context, tenantID uuid.UUID, id, userID uuid.UUID) error {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("get reimbursement: %w", err)
	}
	if reimb.DocStatus != 0 {
		return errors.New("only draft reimbursements can be submitted")
	}
	return s.reimbRepo.UpdateStatus(ctx, tenantID, id, 1)
}

// Approve approves a reimbursement: changes docstatus from 1 to 2 and generates a voucher.
func (s *ReimbursementService) Approve(ctx context.Context, tenantID uuid.UUID, id, userID uuid.UUID) (*model.JournalEntry, error) {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("get reimbursement: %w", err)
	}
	if reimb.DocStatus != 1 {
		return nil, errors.New("only submitted reimbursements can be approved")
	}

	voucher, err := s.GenerateVoucher(ctx, tenantID, id, userID)
	if err != nil {
		return nil, fmt.Errorf("generate voucher: %w", err)
	}

	if err := s.reimbRepo.UpdateStatus(ctx, tenantID, id, 2); err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	if err := s.reimbRepo.UpdateVoucher(ctx, tenantID, id, voucher.ID, voucher.VoucherNo); err != nil {
		return nil, fmt.Errorf("update voucher ref: %w", err)
	}

	return voucher, nil
}

// Reject rejects a submitted reimbursement: changes docstatus from 1 to 3 and stores reject reason.
func (s *ReimbursementService) Reject(ctx context.Context, tenantID uuid.UUID, id uuid.UUID, reason string) error {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("get reimbursement: %w", err)
	}
	if reimb.DocStatus != 1 {
		return errors.New("only submitted reimbursements can be rejected")
	}

	now := time.Now()
	fields := map[string]interface{}{
		"docstatus":     int16(3),
		"reject_reason": reason,
		"updated_at":    now,
	}
	return s.reimbRepo.UpdateFields(ctx, tenantID, id, fields)
}

// GenerateVoucher generates a journal voucher for a reimbursement.
// Debit: expense account based on ExpenseType (6602.03/6602.04/6602.01/6602.05/6602.99)
// Credit: 1002 银行存款
func (s *ReimbursementService) GenerateVoucher(ctx context.Context, tenantID uuid.UUID, id, userID uuid.UUID) (*model.JournalEntry, error) {
	reimb, err := s.reimbRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return nil, fmt.Errorf("get reimbursement: %w", err)
	}

	// Map expense type to account code
	expenseAccountCode := s.getExpenseAccountCode(reimb.ExpenseType)
	bankAccountCode := "1002"

	// Lookup accounts
	expenseAccount, err := s.accountRepo.GetByCode(ctx, tenantID, expenseAccountCode)
	if err != nil {
		return nil, fmt.Errorf("lookup expense account %s: %w", expenseAccountCode, err)
	}
	bankAccount, err := s.accountRepo.GetByCode(ctx, tenantID, bankAccountCode)
	if err != nil {
		return nil, fmt.Errorf("lookup bank account %s: %w", bankAccountCode, err)
	}

	// Generate voucher number
	voucherResp, err := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if err == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}
	if voucherNo == "" {
		voucherNo = fmt.Sprintf("REIMB-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	}

	// Build journal entry
	je := &model.JournalEntry{
		ID:            uuid.New(),
		VoucherNo:     voucherNo,
		PostingDate:   reimb.PostingDate,
		CompanyID:     reimb.CompanyID,
		TenantID:      tenantID,
		DocStatus:     1,
		CreatedBy:     userID,
		SourceDocType: ptr("reimbursement"),
		SourceDocID:   &id,
		SourceDocNo:   &reimb.ReimbursementNo,
	}

	je, err = s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// Build lines: debit expense, credit bank
	lines := []model.JournalEntryLine{
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      expenseAccount.ID,
			Debit:          reimb.Amount,
			Credit:         decimal.Zero,
			AccountCode:    expenseAccount.Code,
			AccountName:    expenseAccount.Name,
			TenantID:       tenantID,
		},
		{
			ID:             uuid.New(),
			JournalEntryID: je.ID,
			AccountID:      bankAccount.ID,
			Debit:          decimal.Zero,
			Credit:         reimb.Amount,
			AccountCode:    bankAccount.Code,
			AccountName:    bankAccount.Name,
			TenantID:       tenantID,
		},
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, je.ID, lines)
	if err != nil {
		return nil, fmt.Errorf("add journal entry lines: %w", err)
	}

	return je, nil
}

// getExpenseAccountCode maps expense type to account code.
func (s *ReimbursementService) getExpenseAccountCode(expenseType string) string {
	switch expenseType {
	case model.ExpenseTypeTravel:
		return "6602.03"
	case model.ExpenseTypeEntertain:
		return "6602.04"
	case model.ExpenseTypeOffice:
		return "6602.01"
	case model.ExpenseTypeTransport:
		return "6602.05"
	case model.ExpenseTypeCommunication:
		return "6602.06"
	case model.ExpenseTypeTraining:
		return "6602.07"
	case model.ExpenseTypeWelfare:
		return "6602.08"
	default:
		return "6602.99"
	}
}

func ptr(s string) *string {
	return &s
}