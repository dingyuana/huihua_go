package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

// Threshold amounts — defaults only; real values come from approval_flows.threshold_amount_level2/3
var (
	ThresholdLevel2 int64 = 1000000 // 100万 — fallback if DB has no value
	ThresholdLevel3 int64 = 5000000 // 500万 — fallback if DB has no value
)

// ApprovalService handles approval workflow operations.
type ApprovalService struct {
	approvalRepo *repository.ApprovalRepository
	journalRepo  *repository.JournalRepository
}

// NewApprovalService creates a new ApprovalService.
func NewApprovalService(approvalRepo *repository.ApprovalRepository, journalRepo *repository.JournalRepository) *ApprovalService {
	return &ApprovalService{
		approvalRepo: approvalRepo,
		journalRepo:  journalRepo,
	}
}

// SubmitForApproval submits a voucher for approval.
// If flowID is provided (not nil), that specific approval flow is used; otherwise the tenant's default flow is used.
func (s *ApprovalService) SubmitForApproval(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID, userID uuid.UUID, flowID *uuid.UUID) error {
	// Get the journal entry
	journal, err := s.journalRepo.GetByID(ctx, tenantID, journalEntryID)
	if err != nil {
		return fmt.Errorf("get journal entry: %w", err)
	}

	// Check if already pending approval
	pendingTasks, err := s.approvalRepo.GetPendingTasksByJournalEntry(ctx, tenantID, journalEntryID)
	if err != nil {
		return fmt.Errorf("check pending tasks: %w", err)
	}
	if len(pendingTasks) > 0 {
		return errors.New("voucher already pending approval")
	}

	// Check if already approved
	hasApproved, err := s.approvalRepo.HasApprovedTasks(ctx, tenantID, journalEntryID)
	if err != nil {
		return fmt.Errorf("check approved tasks: %w", err)
	}
	if hasApproved {
		return errors.New("voucher already approved")
	}

	// Get voucher lines to calculate total amount
	lines, err := s.journalRepo.GetLines(ctx, tenantID, journalEntryID)
	if err != nil {
		return fmt.Errorf("get journal lines: %w", err)
	}

	// Calculate total amount (use credit as proxy for amount for AR/AP style vouchers)
	var totalAmount decimal.Decimal
	for _, line := range lines {
		if line.Credit.GreaterThan(decimal.Zero) {
			totalAmount = totalAmount.Add(line.Credit)
		} else {
			totalAmount = totalAmount.Add(line.Debit)
		}
	}

	// Determine required approval levels based on amount (DB thresholds, fallback to package defaults)
	// Get the approval flow (specific flow if flowID provided, otherwise tenant's default)
	var flow *model.ApprovalFlow
	if flowID != nil {
		flow, err = s.approvalRepo.GetFlow(ctx, tenantID, *flowID)
		if err != nil {
			return fmt.Errorf("get approval flow: %w", err)
		}
		if flow == nil {
			return fmt.Errorf("approval flow %s not found", *flowID)
		}
	} else {
		flow, err = s.approvalRepo.GetDefaultFlow(ctx, tenantID)
		if err != nil {
			// Create a default flow if none exists
			flow, err = s.createDefaultFlow(ctx, tenantID)
			if err != nil {
				return fmt.Errorf("create default flow: %w", err)
			}
		}
	}

	// Extract thresholds from flow record, falling back to package defaults
	level2 := ThresholdLevel2
	level3 := ThresholdLevel3
	if flow != nil {
		if !flow.ThresholdAmountLevel2.IsZero() {
			level2 = flow.ThresholdAmountLevel2.CoefficientInt64()
		}
		if !flow.ThresholdAmountLevel3.IsZero() {
			level3 = flow.ThresholdAmountLevel3.CoefficientInt64()
		}
	}
	levels := s.determineApprovalLevels(totalAmount, level2, level3)

	// Parse approvers from flow
	approvers, err := s.approvalRepo.GetApproversForFlow(flow)
	if err != nil {
		return fmt.Errorf("parse approvers: %w", err)
	}

	if len(approvers) == 0 {
		return errors.New("no approvers configured in flow")
	}

	// Create approval tasks based on required levels
	for _, level := range levels {
		if level <= len(approvers) {
			approver := approvers[level-1]
			task := &model.ApprovalTask{
				ID:             uuid.New(),
				FlowID:         flow.ID,
				JournalEntryID:  journalEntryID,
				ApproverID:     approver.ApproverID,
				Level:          level,
				Status:         model.ApprovalStatusPending,
				Amount:         totalAmount,
				CreatedBy:      &userID,
			}
			if _, err := s.approvalRepo.CreateTask(ctx, tenantID, task); err != nil {
				return fmt.Errorf("create approval task: %w", err)
			}

			// Send notification (placeholder - log for now)
			s.sendApprovalNotification(ctx, tenantID, approver.ApproverID, journal.VoucherNo, level)
		}
	}

	// Update journal entry status to posted (docstatus=1)
	if err := s.journalRepo.UpdateStatus(ctx, tenantID, journalEntryID, 1, userID, nil, model.VoucherActionSubmit, nil); err != nil {
		log.Printf("warning: could not update voucher status: %v", err)
	}

	return nil
}

// Approve approves an approval task.
func (s *ApprovalService) Approve(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, userID uuid.UUID, comment string) error {
	// Get the task
	task, err := s.approvalRepo.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	// Verify approver
	if task.ApproverID != userID {
		return errors.New("unauthorized: not the assigned approver")
	}

	// Check if already processed
	if task.Status != model.ApprovalStatusPending {
		return errors.New("task already processed")
	}

	// Complete the task
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}
	if err := s.approvalRepo.CompleteTask(ctx, tenantID, taskID, commentPtr); err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	// Update journal entry status to verified (docstatus=2)
	if err := s.journalRepo.UpdateStatus(ctx, tenantID, task.JournalEntryID, 2, userID, nil, model.VoucherActionApprove, commentPtr); err != nil {
		return fmt.Errorf("update voucher status: %w", err)
	}

	return nil
}

// Reject rejects an approval task.
func (s *ApprovalService) Reject(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, userID uuid.UUID, comment string) error {
	// Get the task
	task, err := s.approvalRepo.GetTaskByID(ctx, tenantID, taskID)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	// Verify approver
	if task.ApproverID != userID {
		return errors.New("unauthorized: not the assigned approver")
	}

	// Check if already processed
	if task.Status != model.ApprovalStatusPending {
		return errors.New("task already processed")
	}

	// Reject the task
	var commentPtr *string
	if comment != "" {
		commentPtr = &comment
	}
	if err := s.approvalRepo.RejectTask(ctx, tenantID, taskID, commentPtr); err != nil {
		return fmt.Errorf("reject task: %w", err)
	}

	// Update journal entry status back to draft (docstatus=0)
	if err := s.journalRepo.UpdateStatus(ctx, tenantID, task.JournalEntryID, 0, userID, nil, model.VoucherActionReject, commentPtr); err != nil {
		return fmt.Errorf("update voucher status: %w", err)
	}

	return nil
}

// GetPendingTasks retrieves pending approval tasks for a user.
func (s *ApprovalService) GetPendingTasks(ctx context.Context, tenantID uuid.UUID, userID uuid.UUID) ([]model.ApprovalTaskWithVoucher, error) {
	return s.approvalRepo.ListPending(ctx, tenantID, userID)
}

// GetApprovalHistory retrieves approval history for a tenant.
func (s *ApprovalService) GetApprovalHistory(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]model.ApprovalHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.approvalRepo.GetApprovalHistoryByTenant(ctx, tenantID, limit, offset)
}

// GetJournalEntryApprovalStatus gets the approval status for a specific journal entry.
func (s *ApprovalService) GetJournalEntryApprovalStatus(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) (string, []model.ApprovalTask, error) {
	pendingTasks, err := s.approvalRepo.GetPendingTasksByJournalEntry(ctx, tenantID, journalEntryID)
	if err != nil {
		return "", nil, fmt.Errorf("get pending tasks: %w", err)
	}

	if len(pendingTasks) == 0 {
		// Check if there are completed tasks
		hasApproved, err := s.approvalRepo.HasApprovedTasks(ctx, tenantID, journalEntryID)
		if err != nil {
			return "", nil, err
		}
		if hasApproved {
			return "approved", nil, nil
		}
		return "no_approval_required", nil, nil
	}

	return "pending", pendingTasks, nil
}

// determineApprovalLevels reads thresholds from the flow record (DB-first),
// falling back to package-level defaults if not set in the DB.
func (s *ApprovalService) determineApprovalLevels(amount decimal.Decimal, level2, level3 int64) []int {
	levels := []int{1} // Level 1 always required

	if level2 > 0 && amount.GreaterThan(decimal.NewFromInt(level2)) {
		levels = append(levels, 2)
	}

	if level3 > 0 && amount.GreaterThan(decimal.NewFromInt(level3)) {
		levels = append(levels, 3)
	}

	return levels
}

// createDefaultFlow creates a default approval flow if none exists.
func (s *ApprovalService) createDefaultFlow(ctx context.Context, tenantID uuid.UUID) (*model.ApprovalFlow, error) {
	// Default approvers: level 1 = general approver, level 2 = finance manager, level 3 = general manager
	approvers := []model.ApproverInfo{
		{Level: 1, ApproverID: uuid.Nil, Role: "general"},
		{Level: 2, ApproverID: uuid.Nil, Role: "financial_manager"},
		{Level: 3, ApproverID: uuid.Nil, Role: "general_manager"},
	}

	approversJSON, err := json.Marshal(approvers)
	if err != nil {
		return nil, fmt.Errorf("marshal approvers: %w", err)
	}

	desc := "Default approval flow"
	flow := &model.ApprovalFlow{
		ID:          uuid.New(),
		FlowName:    "Default",
		Description: &desc,
		Approvers:   approversJSON,
	}

	return s.approvalRepo.CreateFlow(ctx, tenantID, flow)
}

// sendApprovalNotification sends a notification to an approver.
// This is a placeholder - in production, this would integrate with a notification service.
func (s *ApprovalService) sendApprovalNotification(ctx context.Context, tenantID uuid.UUID, approverID uuid.UUID, voucherNo string, level int) {
	log.Printf("[APPROVAL] Notification: User %s has pending approval for voucher %s at level %d", approverID, voucherNo, level)
	// In production: integrate with email/SMS/websocket notification service
}

// CreateApprovalFlow creates a new approval flow.
func (s *ApprovalService) CreateApprovalFlow(ctx context.Context, tenantID uuid.UUID, flowName string, description *string, approvers []ApproverInput, createdBy *uuid.UUID) (*model.ApprovalFlow, error) {
	approversModel := make([]model.ApproverInfo, len(approvers))
	for i, a := range approvers {
		approversModel[i] = model.ApproverInfo{
			Level:      a.Level,
			ApproverID: a.ApproverID,
			Role:       a.Role,
		}
	}
	approversJSON, err := json.Marshal(approversModel)
	if err != nil {
		return nil, fmt.Errorf("marshal approvers: %w", err)
	}

	flow := &model.ApprovalFlow{
		ID:          uuid.New(),
		FlowName:    flowName,
		Description: description,
		Approvers:   approversJSON,
		CreatedBy:   createdBy,
	}

	return s.approvalRepo.CreateFlow(ctx, tenantID, flow)
}

// ApproverInput represents an approver input for creating flows.
type ApproverInput struct {
	Level      int       `json:"level"`
	ApproverID uuid.UUID `json:"approver_id"`
	Role       string    `json:"role"`
}

// ListApprovalFlows lists all approval flows for a tenant.
func (s *ApprovalService) ListApprovalFlows(ctx context.Context, tenantID uuid.UUID) ([]model.ApprovalFlow, error) {
	return s.approvalRepo.ListFlows(ctx, tenantID)
}

// GetJournalEntryTotalAmount calculates the total amount for a journal entry.
func (s *ApprovalService) GetJournalEntryTotalAmount(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) (decimal.Decimal, error) {
	lines, err := s.journalRepo.GetLines(ctx, tenantID, journalEntryID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get lines: %w", err)
	}

	var total decimal.Decimal
	for _, line := range lines {
		total = total.Add(line.Credit)
	}
	return total, nil
}