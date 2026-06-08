package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ApprovalStatus represents the status of an approval task.
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

// ApprovalFlow represents a configured approval workflow.
type ApprovalFlow struct {
	ID                    uuid.UUID       `json:"id" db:"id"`
	TenantID              uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	FlowName              string          `json:"flow_name" db:"flow_name"`
	Description           *string         `json:"description,omitempty" db:"description"`
	Approvers             json.RawMessage `json:"approvers" db:"approvers"`                             // JSON array of approver IDs and levels
	ThresholdAmountLevel2 decimal.Decimal `json:"threshold_amount_level2" db:"threshold_amount_level2"` // default 1000000
	ThresholdAmountLevel3 decimal.Decimal `json:"threshold_amount_level3" db:"threshold_amount_level3"` // default 5000000
	Currency              string          `json:"currency" db:"currency"`                               // default CNY
	CreatedBy             *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt             time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at" db:"updated_at"`
}

// ApproverInfo represents a single approver level in an approval flow.
type ApproverInfo struct {
	Level      int       `json:"level"`
	ApproverID uuid.UUID `json:"approver_id"`
	Role       string    `json:"role"` // e.g., "financial_manager", "general_manager"
}

// ApprovalTask represents a single approval task instance.
type ApprovalTask struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	FlowID         uuid.UUID       `json:"flow_id" db:"flow_id"`
	JournalEntryID uuid.UUID       `json:"journal_entry_id" db:"journal_entry_id"`
	ApproverID     uuid.UUID       `json:"approver_id" db:"approver_id"`
	ApproverName   *string         `json:"approver_name,omitempty" db:"approver_name"`
	Level          int             `json:"level" db:"level"`
	Status         ApprovalStatus  `json:"status" db:"status"`
	Comment        *string         `json:"comment,omitempty" db:"comment"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedBy      *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
}

// ApprovalTaskWithVoucher combines approval task info with voucher details.
type ApprovalTaskWithVoucher struct {
	ApprovalTask
	VoucherNo        string  `json:"voucher_no" db:"voucher_no"`
	PostingDate      string  `json:"posting_date" db:"posting_date"`
	VoucherType      string  `json:"voucher_type" db:"voucher_type"`
	Remark           string  `json:"remark,omitempty" db:"remark"`
	CompanyName      string  `json:"company_name,omitempty" db:"company_name"`
	CurrentLevel     int     `json:"current_level" db:"current_level"`
	TotalLevels      int     `json:"total_levels" db:"total_levels"`
	SubmittedBy      string  `json:"submitted_by,omitempty" db:"submitted_by"`
	SubmittedName    string  `json:"submitted_name,omitempty" db:"submitted_name"`
	CounterpartyName *string `json:"counterparty_name,omitempty" db:"counterparty_name"`
	SourceDocNo      *string `json:"source_doc_no,omitempty" db:"source_doc_no"`
	DocStatus        int16   `json:"docstatus" db:"docstatus"`
	DebitTotal       string  `json:"debit_total,omitempty" db:"debit_total"`
	CreditTotal      string  `json:"credit_total,omitempty" db:"credit_total"`
	FirstAccountCode *string `json:"first_account_code,omitempty" db:"first_account_code"`
	FirstAccountName *string `json:"first_account_name,omitempty" db:"first_account_name"`
}

// ApprovalHistory represents the history of all approval actions.
type ApprovalHistory struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	JournalEntryID uuid.UUID       `json:"journal_entry_id" db:"journal_entry_id"`
	VoucherNo      string          `json:"voucher_no" db:"voucher_no"`
	Action         string          `json:"action" db:"action"` // submitted, approved, rejected
	Status         ApprovalStatus  `json:"status" db:"status"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	ActorID        uuid.UUID       `json:"actor_id" db:"actor_id"`
	ActorName      *string         `json:"actor_name,omitempty" db:"actor_name"`
	Comment        *string         `json:"comment,omitempty" db:"comment"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
}
