package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// ApprovalRepository provides data access for approval flows and tasks.
type ApprovalRepository struct {
	pool *pgxpool.Pool
}

// NewApprovalRepository creates a new ApprovalRepository.
func NewApprovalRepository(pool *pgxpool.Pool) *ApprovalRepository {
	return &ApprovalRepository{pool: pool}
}

// --- Flow Operations ---

// CreateFlow creates a new approval flow.
func (r *ApprovalRepository) CreateFlow(ctx context.Context, tenantID uuid.UUID, flow *model.ApprovalFlow) (*model.ApprovalFlow, error) {
	query := `
		INSERT INTO approval_flows (id, tenant_id, flow_name, description, approvers, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	if flow.ID == uuid.Nil {
		flow.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		flow.ID, tenantID, flow.FlowName, flow.Description, flow.Approvers, flow.CreatedBy,
	).Scan(&flow.CreatedAt, &flow.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create approval flow: %w", err)
	}

	flow.TenantID = tenantID
	return flow, nil
}

// GetFlow retrieves an approval flow by ID.
func (r *ApprovalRepository) GetFlow(ctx context.Context, tenantID uuid.UUID, flowID uuid.UUID) (*model.ApprovalFlow, error) {
	query := `
		SELECT id, tenant_id, flow_name, description, approvers,
		       threshold_amount_level2, threshold_amount_level3, currency,
		       created_by, created_at, updated_at
		FROM approval_flows
		WHERE id = $1 AND tenant_id = $2`

	flow := &model.ApprovalFlow{}
	err := r.pool.QueryRow(ctx, query, flowID, tenantID).Scan(
		&flow.ID, &flow.TenantID, &flow.FlowName, &flow.Description,
		&flow.Approvers,
		&flow.ThresholdAmountLevel2, &flow.ThresholdAmountLevel3, &flow.Currency,
		&flow.CreatedBy, &flow.CreatedAt, &flow.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get approval flow: %w", err)
	}

	return flow, nil
}

// ListFlows retrieves all approval flows for a tenant.
func (r *ApprovalRepository) ListFlows(ctx context.Context, tenantID uuid.UUID) ([]model.ApprovalFlow, error) {
	query := `
		SELECT id, tenant_id, flow_name, description, approvers,
		       threshold_amount_level2, threshold_amount_level3, currency,
		       created_by, created_at, updated_at
		FROM approval_flows
		WHERE tenant_id = $1
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list approval flows: %w", err)
	}
	defer rows.Close()

	var flows []model.ApprovalFlow
	for rows.Next() {
		var flow model.ApprovalFlow
		if err := rows.Scan(
			&flow.ID, &flow.TenantID, &flow.FlowName, &flow.Description,
			&flow.Approvers,
			&flow.ThresholdAmountLevel2, &flow.ThresholdAmountLevel3, &flow.Currency,
			&flow.CreatedBy, &flow.CreatedAt, &flow.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan approval flow: %w", err)
		}
		flows = append(flows, flow)
	}

	return flows, rows.Err()
}

// UpdateFlow updates an existing approval flow.
func (r *ApprovalRepository) UpdateFlow(ctx context.Context, tenantID uuid.UUID, flow *model.ApprovalFlow) error {
	query := `
		UPDATE approval_flows
		SET flow_name = $3, description = $4, approvers = $5, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query,
		flow.ID, tenantID, flow.FlowName, flow.Description, flow.Approvers,
	)
	if err != nil {
		return fmt.Errorf("update approval flow: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("approval flow not found")
	}

	return nil
}

// DeleteFlow deletes an approval flow by ID.
func (r *ApprovalRepository) DeleteFlow(ctx context.Context, tenantID, flowID uuid.UUID) error {
	query := `DELETE FROM approval_flows WHERE id = $1 AND tenant_id = $2`
	result, err := r.pool.Exec(ctx, query, flowID, tenantID)
	if err != nil {
		return fmt.Errorf("delete approval flow: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("approval flow not found")
	}
	return nil
}

// GetDefaultFlow retrieves the default approval flow for a tenant.
func (r *ApprovalRepository) GetDefaultFlow(ctx context.Context, tenantID uuid.UUID) (*model.ApprovalFlow, error) {
	query := `
		SELECT id, tenant_id, flow_name, description, approvers,
		       threshold_amount_level2, threshold_amount_level3, currency,
		       created_by, created_at, updated_at
		FROM approval_flows
		WHERE tenant_id = $1
		ORDER BY created_at ASC
		LIMIT 1`

	flow := &model.ApprovalFlow{}
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&flow.ID, &flow.TenantID, &flow.FlowName, &flow.Description,
		&flow.Approvers,
		&flow.ThresholdAmountLevel2, &flow.ThresholdAmountLevel3, &flow.Currency,
		&flow.CreatedBy, &flow.CreatedAt, &flow.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get default flow: %w", err)
	}

	return flow, nil
}

// --- Task Operations ---

// CreateTask creates a new approval task.
func (r *ApprovalRepository) CreateTask(ctx context.Context, tenantID uuid.UUID, task *model.ApprovalTask) (*model.ApprovalTask, error) {
	query := `
		INSERT INTO approval_tasks (id, flow_id, journal_entry_id, approver_id, approver_name, level, status, comment, amount, tenant_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		task.ID, task.FlowID, task.JournalEntryID, task.ApproverID, task.ApproverName,
		task.Level, task.Status, task.Comment, task.Amount, tenantID, task.CreatedBy,
	).Scan(&task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create approval task: %w", err)
	}

	task.TenantID = tenantID
	return task, nil
}

// CompleteTask marks a task as approved.
func (r *ApprovalRepository) CompleteTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, comment *string) error {
	query := `
		UPDATE approval_tasks
		SET status = $3, comment = $4, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query, taskID, tenantID, model.ApprovalStatusApproved, comment)
	if err != nil {
		return fmt.Errorf("complete task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// RejectTask marks a task as rejected.
func (r *ApprovalRepository) RejectTask(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID, comment *string) error {
	query := `
		UPDATE approval_tasks
		SET status = $3, comment = $4, completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.pool.Exec(ctx, query, taskID, tenantID, model.ApprovalStatusRejected, comment)
	if err != nil {
		return fmt.Errorf("reject task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// ListPending retrieves pending approval tasks for an approver.
func (r *ApprovalRepository) ListPending(ctx context.Context, tenantID uuid.UUID, approverID uuid.UUID) ([]model.ApprovalTaskWithVoucher, error) {
	query := `
		SELECT at.id, at.flow_id, at.journal_entry_id, at.approver_id, at.approver_name, at.level,
		       at.status, at.comment, at.amount, at.tenant_id, at.created_by, at.created_at, at.updated_at, at.completed_at,
		       je.voucher_no, TO_CHAR(je.posting_date, 'YYYY-MM-DD') as posting_date,
		       COALESCE(je.voucher_type, '') as voucher_type,
		       (SELECT COUNT(*) FROM approval_tasks WHERE journal_entry_id = at.journal_entry_id AND tenant_id = $1) as total_levels
		FROM approval_tasks at
		JOIN journal_entries je ON je.id = at.journal_entry_id
		WHERE at.approver_id = $2 AND at.status = $3 AND at.tenant_id = $1
		ORDER BY at.created_at DESC`

	rows, err := r.pool.Query(ctx, query, tenantID, approverID, model.ApprovalStatusPending)
	if err != nil {
		return nil, fmt.Errorf("list pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.ApprovalTaskWithVoucher
	for rows.Next() {
		var t model.ApprovalTaskWithVoucher
		if err := rows.Scan(
			&t.ID, &t.FlowID, &t.JournalEntryID, &t.ApproverID, &t.ApproverName, &t.Level,
			&t.Status, &t.Comment, &t.Amount, &t.TenantID, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
			&t.VoucherNo, &t.PostingDate, &t.VoucherType, &t.TotalLevels,
		); err != nil {
			return nil, fmt.Errorf("scan pending task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// GetTaskByID retrieves an approval task by ID.
func (r *ApprovalRepository) GetTaskByID(ctx context.Context, tenantID uuid.UUID, taskID uuid.UUID) (*model.ApprovalTask, error) {
	query := `
		SELECT id, flow_id, journal_entry_id, approver_id, approver_name, level, status, comment, amount, tenant_id, created_by, created_at, updated_at, completed_at
		FROM approval_tasks
		WHERE id = $1 AND tenant_id = $2`

	task := &model.ApprovalTask{}
	err := r.pool.QueryRow(ctx, query, taskID, tenantID).Scan(
		&task.ID, &task.FlowID, &task.JournalEntryID, &task.ApproverID, &task.ApproverName,
		&task.Level, &task.Status, &task.Comment, &task.Amount, &task.TenantID,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return task, nil
}

// GetPendingTasksByJournalEntry retrieves pending tasks for a specific journal entry.
func (r *ApprovalRepository) GetPendingTasksByJournalEntry(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) ([]model.ApprovalTask, error) {
	query := `
		SELECT id, flow_id, journal_entry_id, approver_id, approver_name, level, status, comment, amount, tenant_id, created_by, created_at, updated_at, completed_at
		FROM approval_tasks
		WHERE journal_entry_id = $1 AND tenant_id = $2 AND status = $3
		ORDER BY level ASC`

	rows, err := r.pool.Query(ctx, query, journalEntryID, tenantID, model.ApprovalStatusPending)
	if err != nil {
		return nil, fmt.Errorf("get pending tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.ApprovalTask
	for rows.Next() {
		var t model.ApprovalTask
		if err := rows.Scan(
			&t.ID, &t.FlowID, &t.JournalEntryID, &t.ApproverID, &t.ApproverName,
			&t.Level, &t.Status, &t.Comment, &t.Amount, &t.TenantID,
			&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// GetTaskHistory retrieves approval history for a journal entry.
func (r *ApprovalRepository) GetTaskHistory(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) ([]model.ApprovalHistory, error) {
	query := `
		SELECT at.id, at.journal_entry_id, je.voucher_no, 
		       CASE WHEN at.status = 'approved' THEN 'approved' ELSE 'rejected' END as action,
		       at.status, at.amount, at.tenant_id, at.approver_id as actor_id, 
		       at.approver_name as actor_name, at.comment, at.completed_at as created_at
		FROM approval_tasks at
		JOIN journal_entries je ON je.id = at.journal_entry_id
		WHERE at.journal_entry_id = $1 AND at.tenant_id = $2 AND at.status != 'pending'
		ORDER BY at.completed_at ASC`

	rows, err := r.pool.Query(ctx, query, journalEntryID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get task history: %w", err)
	}
	defer rows.Close()

	var history []model.ApprovalHistory
	for rows.Next() {
		var h model.ApprovalHistory
		if err := rows.Scan(
			&h.ID, &h.JournalEntryID, &h.VoucherNo, &h.Action,
			&h.Status, &h.Amount, &h.TenantID, &h.ActorID,
			&h.ActorName, &h.Comment, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// GetApprovalHistoryByTenant retrieves all approval history for a tenant with pagination.
func (r *ApprovalRepository) GetApprovalHistoryByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]model.ApprovalHistory, error) {
	query := `
		SELECT at.id, at.journal_entry_id, je.voucher_no,
		       CASE WHEN at.status = 'approved' THEN 'approved' ELSE 'rejected' END as action,
		       at.status, at.amount, at.tenant_id, at.approver_id as actor_id,
		       at.approver_name as actor_name, at.comment, at.completed_at as created_at
		FROM approval_tasks at
		JOIN journal_entries je ON je.id = at.journal_entry_id
		WHERE at.tenant_id = $1 AND at.status != 'pending'
		ORDER BY at.completed_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get approval history: %w", err)
	}
	defer rows.Close()

	var history []model.ApprovalHistory
	for rows.Next() {
		var h model.ApprovalHistory
		if err := rows.Scan(
			&h.ID, &h.JournalEntryID, &h.VoucherNo, &h.Action,
			&h.Status, &h.Amount, &h.TenantID, &h.ActorID,
			&h.ActorName, &h.Comment, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// GetApproversForFlow parses and returns approvers for a given flow.
func (r *ApprovalRepository) GetApproversForFlow(flow *model.ApprovalFlow) ([]model.ApproverInfo, error) {
	var approvers []model.ApproverInfo
	if err := json.Unmarshal(flow.Approvers, &approvers); err != nil {
		return nil, fmt.Errorf("parse approvers: %w", err)
	}
	return approvers, nil
}

// HasApprovedTasks checks if a journal entry has any approved tasks.
func (r *ApprovalRepository) HasApprovedTasks(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM approval_tasks 
			WHERE journal_entry_id = $1 AND tenant_id = $2 AND status = 'approved'
		)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, journalEntryID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check approved tasks: %w", err)
	}

	return exists, nil
}

// RecordApprovalAction records an approval action to history.
func (r *ApprovalRepository) RecordApprovalAction(ctx context.Context, tenantID uuid.UUID, journalEntryID uuid.UUID, action string, status model.ApprovalStatus, approverID uuid.UUID, approverName *string, comment *string, amount decimal.Decimal) error {
	// For simplicity, we store the action in the approval_tasks table with the status
	// The completed_at and comment fields already capture this
	return nil
}