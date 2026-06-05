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
	journalRepo     *repository.JournalRepository
	auditRepo       *repository.AuditRepository
	glRepo          *repository.GLEntryRepository
	bankJournalRepo *repository.BankJournalRepository
	bankRepo        *repository.BankRepository
	arInvoiceRepo   *repository.ArInvoiceRepository
	paymentRepo     *repository.PaymentEntryRepository
}

// NewVoucherStateMachine creates a new VoucherStateMachine.
func NewVoucherStateMachine(journalRepo *repository.JournalRepository, auditRepo *repository.AuditRepository, glRepo *repository.GLEntryRepository) *VoucherStateMachine {
	return &VoucherStateMachine{
		journalRepo: journalRepo,
		auditRepo:   auditRepo,
		glRepo:      glRepo,
	}
}

// NewVoucherStateMachineWithBankJournal creates a VoucherStateMachine with bank journal auto-registration.
func NewVoucherStateMachineWithBankJournal(journalRepo *repository.JournalRepository, auditRepo *repository.AuditRepository, glRepo *repository.GLEntryRepository, bankJournalRepo *repository.BankJournalRepository, bankRepo *repository.BankRepository) *VoucherStateMachine {
	return &VoucherStateMachine{
		journalRepo:     journalRepo,
		auditRepo:       auditRepo,
		glRepo:          glRepo,
		bankJournalRepo: bankJournalRepo,
		bankRepo:        bankRepo,
	}
}

// InjectLockRepos injects the repositories needed for source-document locking (ArInvoice, PaymentEntry).
func (s *VoucherStateMachine) InjectLockRepos(arInvoiceRepo *repository.ArInvoiceRepository, paymentRepo *repository.PaymentEntryRepository) {
	s.arInvoiceRepo = arInvoiceRepo
	s.paymentRepo = paymentRepo
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
		case model.VoucherActionApprove, model.VoucherActionReject, model.VoucherActionReverse, model.VoucherActionCancel:
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

// ExecuteTransition performs a state transition on a voucher within a single database transaction.
// This ensures atomicity: status update, GL entries, and audit log either all succeed or all fail.
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

	// Begin a single database transaction for all operations
	tx, err := s.journalRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update status within the transaction
	if err := s.journalRepo.UpdateStatusTx(ctx, tx, tenantID, journalID, newStatus, userID, userNamePtr, action, reasonPtr); err != nil {
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
		if err := s.glRepo.WriteGLEntriesTx(ctx, tx, tenantID, journalID, lines, je.PostingDate, voucherType, je.CompanyID); err != nil {
			return fmt.Errorf("write GL entries: %w", err)
		}
	}

	// Cancel GL entries on reverse/cancel
	if (action == model.VoucherActionReverse || action == model.VoucherActionCancel) && s.glRepo != nil {
		if err := s.glRepo.CancelGLEntriesByVoucherTx(ctx, tx, tenantID, journalID); err != nil {
			return fmt.Errorf("cancel GL entries: %w", err)
		}
	}

	// Auto-register bank journal entry on approve (posted -> verified)
	if action == model.VoucherActionApprove && s.bankJournalRepo != nil {
		if err := s.registerBankJournalEntry(ctx, tx, tenantID, journalID); err != nil {
			return fmt.Errorf("register bank journal entry: %w", err)
		}
	}

	// Record audit log within the transaction
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
	if err := s.auditRepo.CreateTx(ctx, tx, tenantID, auditLog); err != nil {
		return fmt.Errorf("record audit log: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	// Post-commit: lock/unlock source document based on action
	s.postTransitionAction(ctx, tenantID, journalID, action, userID)

	return nil
}

// postTransitionAction handles source-document lock/unlock after the transaction commits.
// Must be called outside the transaction to avoid holding a transaction across I/O.
func (s *VoucherStateMachine) postTransitionAction(ctx context.Context, tenantID, journalID uuid.UUID, action model.VoucherAction, userID uuid.UUID) {
	if s.arInvoiceRepo == nil && s.paymentRepo == nil {
		return
	}

	je, err := s.journalRepo.GetByID(ctx, tenantID, journalID)
	if err != nil {
		return
	}

	switch action {
	case model.VoucherActionSubmit:
		// Lock source doc and set voucher_id
		s.lockSourceDoc(ctx, tenantID, je, journalID, userID)
	case model.VoucherActionCancel:
		// Unlock source doc (voucher was in draft, no voucher_id set yet)
		s.unlockSourceDoc(ctx, tenantID, je)
	case model.VoucherActionReverse:
		// Unlock original source doc
		s.unlockSourceDoc(ctx, tenantID, je)
	}
}

// lockSourceDoc locks the source document (ArInvoice or PaymentEntry) and writes voucher_id.
func (s *VoucherStateMachine) lockSourceDoc(ctx context.Context, tenantID uuid.UUID, je *model.JournalEntry, voucherID, userID uuid.UUID) {
	if je.SourceDocType == nil || je.SourceDocID == nil {
		return
	}
	switch *je.SourceDocType {
	case "ar_invoice":
		if s.arInvoiceRepo != nil {
			_ = s.arInvoiceRepo.Lock(ctx, tenantID, *je.SourceDocID, userID)
			_ = s.arInvoiceRepo.SetVoucherID(ctx, tenantID, *je.SourceDocID, voucherID)
		}
	case "payment_entry":
		if s.paymentRepo != nil {
			_ = s.paymentRepo.UpdateStatusAndClearVoucherTx(ctx, nil, tenantID, *je.SourceDocID, 2) // docstatus=2 (approved/posted)
			// SetVoucherID on payment entry via generic update
			_ = s.paymentRepo.SetVoucherID(ctx, tenantID, *je.SourceDocID, voucherID)
		}
	}
}

// unlockSourceDoc unlocks the source document (ArInvoice or PaymentEntry).
func (s *VoucherStateMachine) unlockSourceDoc(ctx context.Context, tenantID uuid.UUID, je *model.JournalEntry) {
	if je.SourceDocType == nil || je.SourceDocID == nil {
		return
	}
	switch *je.SourceDocType {
	case "ar_invoice":
		if s.arInvoiceRepo != nil {
			_ = s.arInvoiceRepo.Unlock(ctx, tenantID, *je.SourceDocID)
		}
	case "payment_entry":
		if s.paymentRepo != nil {
			_ = s.paymentRepo.UpdateStatusAndClearVoucherTx(ctx, nil, tenantID, *je.SourceDocID, 1) // revert to docstatus=1 (submitted)
		}
	}
}

// ReverseVoucher creates a reversal voucher (red letter) for the given voucher within a single transaction.
// It swaps debit and credit amounts for each line, sets reversal_id on the original,
// and marks the new voucher with action='reversal' and original_voucher_id.
func (s *VoucherStateMachine) ReverseVoucher(ctx context.Context, tenantID uuid.UUID, journalID uuid.UUID, userID uuid.UUID, userName string) (*model.JournalEntry, error) {
	// Get original voucher (read-only, no transaction needed yet)
	original, err := s.journalRepo.GetByID(ctx, tenantID, journalID)
	if err != nil {
		return nil, fmt.Errorf("get original voucher: %w", err)
	}

	// Check if already reversed
	if original.ReversedID != nil {
		return nil, errors.New("voucher is already reversed")
	}

	// Get lines (read-only, needed to build reversal lines)
	lines, err := s.journalRepo.GetLines(ctx, tenantID, journalID)
	if err != nil {
		return nil, fmt.Errorf("get lines: %w", err)
	}

	// Create reversal voucher within a single transaction
	tx, err := s.journalRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	reversal := &model.JournalEntry{
		ID:          uuid.New(),
		VoucherNo:   original.VoucherNo + "-R",
		VoucherType: original.VoucherType,
		PostingDate: now,
		CompanyID:   original.CompanyID,
		TenantID:    tenantID,
		DocStatus:   1, // posted by default for reversal
		ReversalID:  &journalID,
		SubmittedBy: &userID,
		SubmittedAt: &now,
		CreatedBy:   userID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Insert reversal voucher within transaction
	reversal, err = s.journalRepo.CreateTx(ctx, tx, tenantID, reversal)
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

	// Insert reversal lines within transaction
	_, err = s.journalRepo.AddLinesTx(ctx, tx, tenantID, reversal.ID, reversalLines)
	if err != nil {
		return nil, fmt.Errorf("add reversal lines: %w", err)
	}

	// Write GL entries for reversal voucher
	if s.glRepo != nil {
		if err := s.glRepo.WriteGLEntriesTx(ctx, tx, tenantID, reversal.ID, reversalLines, now, original.VoucherType, original.CompanyID); err != nil {
			return nil, fmt.Errorf("write GL entries for reversal: %w", err)
		}
	}

	// Mark original as reversed within transaction
	if err := s.journalRepo.UpdateStatusTx(ctx, tx, tenantID, journalID, 3, userID, &userName, model.VoucherActionReverse, nil); err != nil {
		return nil, fmt.Errorf("mark original as reversed: %w", err)
	}

	// Record audit log for reversal within transaction
	userNamePtr := &userName
	metadata, _ := json.Marshal(map[string]interface{}{
		"original_voucher_id": journalID.String(),
		"reversal_voucher_id": reversal.ID.String(),
	})
	auditLog := &model.AuditLog{
		ID:         uuid.New(),
		Action:     "voucher_status_change",
		ObjectType: "journal_entry",
		ObjectID:   journalID,
		TenantID:   tenantID,
		ActorID:    userID,
		ActorName:  userNamePtr,
		Metadata:   metadata,
	}
	if err := s.auditRepo.CreateTx(ctx, tx, tenantID, auditLog); err != nil {
		return nil, fmt.Errorf("record audit log: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Post-commit: lock new reversal voucher, unlock original
	s.postTransitionAction(ctx, tenantID, reversal.ID, model.VoucherActionSubmit, userID)
	s.postTransitionAction(ctx, tenantID, original.ID, model.VoucherActionReverse, userID)

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

// registerBankJournalEntry registers a bank journal entry when a voucher is approved.
// It finds the bank account (1002) line in the voucher and creates a corresponding entry.
func (s *VoucherStateMachine) registerBankJournalEntry(ctx context.Context, tx interface{}, tenantID uuid.UUID, journalID uuid.UUID) error {
	// Get voucher lines
	lines, err := s.journalRepo.GetLines(ctx, tenantID, journalID)
	if err != nil {
		return fmt.Errorf("get lines: %w", err)
	}

	// Get the journal entry for voucher number and posting date
	je, err := s.journalRepo.GetByID(ctx, tenantID, journalID)
	if err != nil {
		return fmt.Errorf("get journal entry: %w", err)
	}

	// Find the bank account (1002) line
	var bankLine *model.JournalEntryLine
	for i := range lines {
		if lines[i].AccountCode == "1002" {
			bankLine = &lines[i]
			break
		}
	}
	if bankLine == nil {
		// No bank account line found, skip registration
		return nil
	}

	// Determine bank account ID
	var bankAccountID uuid.UUID
	// Try to find a bank account linked to 1002
	bankAccounts, err := s.bankRepo.ListByTenant(ctx, tenantID)
	if err == nil && len(bankAccounts) > 0 {
		// Use the first bank account as default
		bankAccountID = bankAccounts[0].ID
	}

	// If voucher has source_doc_type=payment and source_doc_id, use that payment's bank_account_id
	if je.SourceDocType != nil && je.SourceDocID != nil {
		// We could look up the payment entry to get the bank_account_id
		// For now, use the first bank account or skip
	}

	// Build description
	var desc string
	if je.Remark != nil {
		desc = *je.Remark
	} else {
		desc = fmt.Sprintf("凭证 %s", je.VoucherNo)
	}

	// Determine debit/credit direction based on the bank line
	var debit, credit decimal.Decimal
	if bankLine.Debit.GreaterThan(decimal.Zero) {
		debit = bankLine.Debit
	} else {
		credit = bankLine.Credit
	}

	// Create bank journal entry
	entry := &model.BankJournalEntry{
		ID:            uuid.New(),
		BankAccountID: bankAccountID,
		TxnDate:       je.PostingDate,
		Description:  &desc,
		Debit:         debit,
		Credit:        credit,
		VoucherID:     &journalID,
		VoucherNo:     &je.VoucherNo,
		TenantID:      tenantID,
	}

	// Use tx if it's a pgx.Tx (bank journal create doesn't need tx, but we pass it anyway)
	if s.bankJournalRepo != nil {
		_, err := s.bankJournalRepo.Create(ctx, tenantID, entry)
		if err != nil {
			return fmt.Errorf("create bank journal entry: %w", err)
		}
	}

	return nil
}
