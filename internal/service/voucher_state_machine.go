package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// VoucherStateMachine handles voucher status transitions.
type VoucherStateMachine struct {
	journalRepo *repository.JournalRepository
	auditRepo   *repository.AuditRepository
	glRepo      *repository.GLEntryRepository
}

// NewVoucherStateMachine creates a new VoucherStateMachine.
func NewVoucherStateMachine(journalRepo *repository.JournalRepository, auditRepo *repository.AuditRepository, glRepo *repository.GLEntryRepository) *VoucherStateMachine {
	return &VoucherStateMachine{
		journalRepo: journalRepo,
		auditRepo:   auditRepo,
		glRepo:      glRepo,
	}
}

// ValidateTransition checks if a state transition is legal.
func (s *VoucherStateMachine) ValidateTransition(currentStatus int16, action model.VoucherAction) error {
	// docstatus: 0=draft, 1=posted, 2=verified, 3=cancelled
	switch currentStatus {
	case 0: // draft
		switch action {
		case model.VoucherActionSubmit, model.VoucherActionCancel:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case 1: // posted
		switch action {
		case model.VoucherActionApprove, model.VoucherActionReject, model.VoucherActionReverse:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case 2: // verified
		switch action {
		case model.VoucherActionReverse:
			return nil
		default:
			return fmt.Errorf("invalid action %q for status %d", action, currentStatus)
		}
	case 3: // cancelled
		return errors.New("voucher is already cancelled")
	default:
		return fmt.Errorf("unknown status: %d", currentStatus)
	}
}

// ExecuteTransition performs a state transition on a voucher.
func (s *VoucherStateMachine) ExecuteTransition(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID, action model.VoucherAction, userID uuid.UUID, userName string, reason string) error {
	// Get current status
	currentStatus, err := s.journalRepo.GetStatus(ctx, tenantID, journalID)
	if err != nil {
		return fmt.Errorf("get current status: %w", err)
	}

	// Validate transition
	if err := s.ValidateTransition(currentStatus, action); err != nil {
		return fmt.Errorf("invalid transition: %w", err)
	}

	// Determine new status
	var newStatus int16
	switch action {
	case model.VoucherActionSubmit:
		newStatus = 1 // draft -> posted
	case model.VoucherActionApprove:
		newStatus = 2 // posted -> verified
	case model.VoucherActionReject:
		newStatus = 0 // posted -> draft
	case model.VoucherActionCancel:
		newStatus = 3 // draft -> cancelled
	case model.VoucherActionReverse:
		newStatus = 3 // posted/verified -> cancelled
	}

	// Prepare reason
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}
	var userNamePtr *string
	if userName != "" {
		userNamePtr = &userName
	}

	// Update status in repository
	if err := s.journalRepo.UpdateStatus(ctx, tenantID, journalID, newStatus, userID, userNamePtr, action, reasonPtr); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Write GL entries on submit (draft -> posted)
	if action == model.VoucherActionSubmit && s.glRepo != nil {
		lines, err := s.journalRepo.GetLines(ctx, tenantID, journalID)
		if err != nil {
			return fmt.Errorf("get journal lines for GL: %w", err)
		}
		je, err := s.journalRepo.GetByID(ctx, tenantID, journalID)
		if err != nil {
			return fmt.Errorf("get journal entry for GL: %w", err)
		}
		voucherType := je.VoucherType
		if err := s.glRepo.WriteGLEntries(ctx, tenantID, journalID, lines, je.PostingDate, voucherType, je.CompanyID); err != nil {
			return fmt.Errorf("write GL entries: %w", err)
		}
	}

	// Cancel GL entries on reverse/cancel
	if (action == model.VoucherActionReverse || action == model.VoucherActionCancel) && s.glRepo != nil {
		if err := s.glRepo.CancelGLEntriesByVoucher(ctx, tenantID, journalID); err != nil {
			return fmt.Errorf("cancel GL entries: %w", err)
		}
	}

	// Record audit log
	changedFields, _ := json.Marshal(map[string][]interface{}{
		"docstatus": {currentStatus, newStatus},
	})
	auditLog := &model.AuditLog{
		ID:            uuid.New(),
		Action:        "voucher_status_change",
		ObjectType:    "journal_entry",
		ObjectID:      journalID,
		TenantID:      tenantID,
		ActorID:       userID,
		ActorName:     userNamePtr,
		ChangedFields: changedFields,
	}
	if reasonPtr != nil {
		metadata, _ := json.Marshal(map[string]string{"reason": *reasonPtr})
		auditLog.Metadata = metadata
	}
	if err := s.auditRepo.Create(ctx, tenantID, auditLog); err != nil {
		return fmt.Errorf("record audit log: %w", err)
	}

	return nil
}

// ReverseVoucher creates a reversal voucher (red letter) for the given voucher.
// It swaps debit and credit amounts for each line, sets reversal_id on the original,
// and marks the new voucher with action='reversal' and original_voucher_id.
func (s *VoucherStateMachine) ReverseVoucher(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID, userID uuid.UUID, userName string) (*model.JournalEntry, error) {
	// Get original voucher
	original, err := s.journalRepo.GetByID(ctx, tenantID, journalID)
	if err != nil {
		return nil, fmt.Errorf("get original voucher: %w", err)
	}

	// Check if already reversed
	if original.ReversedID != nil {
		return nil, errors.New("voucher is already reversed")
	}

	// Get lines
	lines, err := s.journalRepo.GetLines(ctx, tenantID, journalID)
	if err != nil {
		return nil, fmt.Errorf("get lines: %w", err)
	}

	// Create reversal voucher
	now := time.Now()
	reversal := &model.JournalEntry{
		ID:            uuid.New(),
		VoucherNo:     original.VoucherNo + "-R",
		VoucherType:   original.VoucherType,
		PostingDate:   now,
		CompanyID:     original.CompanyID,
		TenantID:      tenantID,
		DocStatus:     1, // posted by default for reversal
		ReversalID:    &journalID,
		SubmittedBy:   &userID,
		SubmittedAt:   &now,
		CreatedBy:     userID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Insert reversal voucher
	reversal, err = s.journalRepo.Create(ctx, tenantID, reversal)
	if err != nil {
		return nil, fmt.Errorf("create reversal voucher: %w", err)
	}

	// Swap debit/credit for each line
	reversalLines := make([]model.JournalEntryLine, len(lines))
	for i, line := range lines {
		reversalLines[i] = model.JournalEntryLine{
			ID:             uuid.New(),
			JournalEntryID: reversal.ID,
			AccountID:      line.AccountID,
			Debit:          line.Credit,
			Credit:         line.Debit,
			DebitCcy:       line.CreditCcy,
			CreditCcy:      line.DebitCcy,
			AccountCcy:     line.AccountCcy,
			ExchangeRate:   line.ExchangeRate,
			PartyType:      line.PartyType,
			PartyID:        line.PartyID,
			CostCenterID:   line.CostCenterID,
			ProjectID:      line.ProjectID,
			UserRemark:     line.UserRemark,
			Reconciled:     false,
			TenantID:       tenantID,
		}
	}

	_, err = s.journalRepo.AddLines(ctx, tenantID, reversal.ID, reversalLines)
	if err != nil {
		return nil, fmt.Errorf("add reversal lines: %w", err)
	}

	// Mark original as reversed
	if err := s.journalRepo.UpdateStatus(ctx, tenantID, journalID, 3, userID, &userName, model.VoucherActionReverse, nil); err != nil {
		return nil, fmt.Errorf("mark original as reversed: %w", err)
	}

	// Record audit log for reversal
	userNamePtr := &userName
	metadata, _ := json.Marshal(map[string]interface{}{
		"original_voucher_id": journalID.String(),
		"reversal_voucher_id": reversal.ID.String(),
	})
	auditLog := &model.AuditLog{
		ID:            uuid.New(),
		Action:        "voucher_status_change",
		ObjectType:    "journal_entry",
		ObjectID:      journalID,
		TenantID:      tenantID,
		ActorID:       userID,
		ActorName:     userNamePtr,
		Metadata:      metadata,
	}
	if err := s.auditRepo.Create(ctx, tenantID, auditLog); err != nil {
		return nil, fmt.Errorf("record audit log: %w", err)
	}

	return reversal, nil
}

// ValidateTransitionStatusString validates transition using string status.
func (s *VoucherStateMachine) ValidateTransitionStatusString(currentStatus string, action model.VoucherAction) error {
	var status int16
	switch currentStatus {
	case "draft":
		status = 0
	case "posted":
		status = 1
	case "verified":
		status = 2
	case "cancelled":
		status = 3
	default:
		return fmt.Errorf("unknown status: %s", currentStatus)
	}
	return s.ValidateTransition(status, action)
}

// DocStatusToVoucherStatus converts a docstatus int to VoucherStatus string.
func DocStatusToVoucherStatus(docstatus int16) model.VoucherStatus {
	switch docstatus {
	case 0:
		return model.VoucherStatusDraft
	case 1:
		return model.VoucherStatusPosted
	case 2:
		return model.VoucherStatusVerified
	case 3:
		return model.VoucherStatusCancelled
	default:
		return model.VoucherStatusDraft
	}
}

// VoucherStatusToDocStatus converts a VoucherStatus string to docstatus int.
func VoucherStatusToDocStatus(status model.VoucherStatus) int16 {
	switch status {
	case model.VoucherStatusDraft:
		return 0
	case model.VoucherStatusPosted:
		return 1
	case model.VoucherStatusVerified:
		return 2
	case model.VoucherStatusCancelled:
		return 3
	default:
		return 0
	}
}

// ValidateDebitCredit ensures total debits equals total credits.
func ValidateDebitCredit(lines []model.JournalEntryLine) error {
	var totalDebit, totalCredit decimal.Decimal
	for _, line := range lines {
		totalDebit = totalDebit.Add(line.Debit)
		totalCredit = totalCredit.Add(line.Credit)
	}
	if !totalDebit.Equal(totalCredit) {
		return fmt.Errorf("debit (%s) != credit (%s)", totalDebit.String(), totalCredit.String())
	}
	return nil
}