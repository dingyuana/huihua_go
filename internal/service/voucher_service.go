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

// VoucherService handles manual voucher CRUD operations.
type VoucherService struct {
	journalRepo   *repository.JournalRepository
	templateSvc    *VoucherTemplateService
}

// NewVoucherService creates a new VoucherService.
func NewVoucherService(journalRepo *repository.JournalRepository, templateSvc *VoucherTemplateService) *VoucherService {
	return &VoucherService{
		journalRepo:   journalRepo,
		templateSvc:    templateSvc,
	}
}

// CreateVoucherRequest is the request body for creating a voucher.
type CreateVoucherRequest struct {
	VoucherType  *string                   `json:"voucher_type"`
	PostingDate  time.Time                  `json:"posting_date"`
	CompanyID    uuid.UUID                  `json:"company_id"`
	Remark       *string                    `json:"remark,omitempty"`
	Lines        []VoucherLineRequest       `json:"lines"`
}

// VoucherLineRequest represents a single line in a voucher creation request.
type VoucherLineRequest struct {
	AccountID    uuid.UUID  `json:"account_id"`
	Debit        string     `json:"debit"`        // decimal as string
	Credit       string     `json:"credit"`       // decimal as string
	PartyType    *string    `json:"party_type,omitempty"`
	PartyID      *uuid.UUID `json:"party_id,omitempty"`
	CostCenterID *uuid.UUID `json:"cost_center_id,omitempty"`
	ProjectID    *uuid.UUID `json:"project_id,omitempty"`
	UserRemark   *string    `json:"user_remark,omitempty"`
}

// CreateVoucher creates a new manual voucher with lines (draft status).
func (s *VoucherService) CreateVoucher(ctx context.Context, tenantID, createdBy uuid.UUID, req *CreateVoucherRequest) (*model.JournalEntry, error) {
	if len(req.Lines) == 0 {
		return nil, errors.New("at least one line is required")
	}

	// Generate voucher number
	voucherResp, err := s.templateSvc.GenerateVoucherNumber(ctx, tenantID)
	voucherNo := ""
	if err == nil && voucherResp != nil {
		voucherNo = voucherResp.VoucherNumber
	}
	if voucherNo == "" {
		voucherNo = fmt.Sprintf("MANUAL-%s-%d", time.Now().Format("20060102"), time.Now().UnixNano()%1000000)
	}

	je := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   voucherNo,
		VoucherType: req.VoucherType,
		PostingDate: req.PostingDate,
		CompanyID:   req.CompanyID,
		TenantID:    tenantID,
		Remark:      req.Remark,
		DocStatus:   0, // draft
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	created, err := s.journalRepo.Create(ctx, tenantID, je)
	if err != nil {
		return nil, fmt.Errorf("create journal entry: %w", err)
	}

	// Build lines
	lines := make([]model.JournalEntryLine, 0, len(req.Lines))
	for _, lr := range req.Lines {
		debit, _ := decimal.NewFromString(lr.Debit)
		credit, _ := decimal.NewFromString(lr.Credit)
		line := model.JournalEntryLine{
			ID:             uuid.New(),
			JournalEntryID: created.ID,
			AccountID:      lr.AccountID,
			Debit:          debit,
			Credit:         credit,
			DebitCcy:       debit,
			CreditCcy:      credit,
			ExchangeRate:   decimal.NewFromInt(1),
			PartyType:      lr.PartyType,
			PartyID:        lr.PartyID,
			CostCenterID:   lr.CostCenterID,
			ProjectID:      lr.ProjectID,
			UserRemark:     lr.UserRemark,
			Reconciled:     false,
			TenantID:       tenantID,
		}
		lines = append(lines, line)
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, created.ID, lines)
	if err != nil {
		return nil, fmt.Errorf("add lines: %w", err)
	}

	return created, nil
}

// ListVouchersRequest holds filter params for listing vouchers.
type ListVouchersRequest struct {
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	VoucherType *string    `json:"voucher_type,omitempty"`
	DocStatus   *int16     `json:"doc_status,omitempty"`
	AccountID   *uuid.UUID `json:"account_id,omitempty"` // filter by line account
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// VoucherDetail is a journal entry with its lines.
type VoucherDetail struct {
	JournalEntry          model.JournalEntry       `json:"journal_entry"`
	JournalEntryLines     []model.JournalEntryLine `json:"journal_entry_lines"`
	VoucherStateTransitions []model.VoucherStateTransition `json:"voucher_state_transitions,omitempty"`
}

// ListVouchers lists vouchers with optional filters.
func (s *VoucherService) ListVouchers(ctx context.Context, tenantID uuid.UUID, req *ListVouchersRequest) ([]model.JournalEntry, error) {
	limit := 50
	if req.Limit > 0 && req.Limit <= 200 {
		limit = req.Limit
	}
	return s.journalRepo.ListVouchers(ctx, tenantID, req.StartDate, req.EndDate, req.VoucherType, req.DocStatus, req.AccountID, limit, req.Offset)
}

// GetVoucher retrieves a voucher with its lines.
func (s *VoucherService) GetVoucher(ctx context.Context, tenantID, voucherID uuid.UUID) (*VoucherDetail, error) {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get voucher: %w", err)
	}

	lines, err := s.journalRepo.GetLines(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get lines: %w", err)
	}

	transitions, err := s.journalRepo.GetTransitions(ctx, tenantID, voucherID)
	if err != nil {
		return nil, fmt.Errorf("get transitions: %w", err)
	}

	return &VoucherDetail{
		JournalEntry:             *je,
		JournalEntryLines:       lines,
		VoucherStateTransitions:  transitions,
	}, nil
}

// UpdateVoucherRequest is the request body for updating a voucher.
type UpdateVoucherRequest struct {
	VoucherType *string               `json:"voucher_type,omitempty"`
	PostingDate *time.Time            `json:"posting_date,omitempty"`
	Remark      *string               `json:"remark,omitempty"`
	Lines       []VoucherLineRequest `json:"lines"`
}

// UpdateVoucher updates a draft voucher and replaces its lines.
func (s *VoucherService) UpdateVoucher(ctx context.Context, tenantID, voucherID, updatedBy uuid.UUID, req *UpdateVoucherRequest) error {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return fmt.Errorf("get voucher: %w", err)
	}

	if je.DocStatus != 0 {
		return errors.New("only draft vouchers can be updated")
	}

	// Update header fields
	if req.VoucherType != nil {
		je.VoucherType = req.VoucherType
	}
	if req.PostingDate != nil {
		je.PostingDate = *req.PostingDate
	}
	if req.Remark != nil {
		je.Remark = req.Remark
	}
	je.UpdatedAt = time.Now()

	if err := s.journalRepo.Update(ctx, tenantID, je); err != nil {
		return fmt.Errorf("update journal entry: %w", err)
	}

	if len(req.Lines) > 0 {
		// Delete existing lines
		if err := s.journalRepo.DeleteLines(ctx, tenantID, voucherID); err != nil {
			return fmt.Errorf("delete existing lines: %w", err)
		}

		// Insert new lines
		lines := make([]model.JournalEntryLine, 0, len(req.Lines))
		for _, lr := range req.Lines {
			debit, _ := decimal.NewFromString(lr.Debit)
			credit, _ := decimal.NewFromString(lr.Credit)
			line := model.JournalEntryLine{
				ID:             uuid.New(),
				JournalEntryID: voucherID,
				AccountID:      lr.AccountID,
				Debit:          debit,
				Credit:         credit,
				DebitCcy:       debit,
				CreditCcy:      credit,
				ExchangeRate:   decimal.NewFromInt(1),
				PartyType:      lr.PartyType,
				PartyID:        lr.PartyID,
				CostCenterID:   lr.CostCenterID,
				ProjectID:      lr.ProjectID,
				UserRemark:     lr.UserRemark,
				Reconciled:     false,
				TenantID:       tenantID,
			}
			lines = append(lines, line)
		}

		if _, err := s.journalRepo.AddLines(ctx, tenantID, voucherID, lines); err != nil {
			return fmt.Errorf("add lines: %w", err)
		}
	}

	return nil
}

// DeleteVoucher deletes a draft voucher and its lines.
func (s *VoucherService) DeleteVoucher(ctx context.Context, tenantID, voucherID, updatedBy uuid.UUID) error {
	je, err := s.journalRepo.GetByID(ctx, tenantID, voucherID)
	if err != nil {
		return fmt.Errorf("get voucher: %w", err)
	}

	if je.DocStatus != 0 {
		return errors.New("only draft vouchers can be deleted")
	}

	return s.journalRepo.DeleteVoucher(ctx, tenantID, voucherID)
}
