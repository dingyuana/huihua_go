package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// VoucherStatus represents the status of a voucher (journal entry).
type VoucherStatus string

const (
	VoucherStatusDraft     VoucherStatus = "draft"     // 草稿
	VoucherStatusPosted    VoucherStatus = "posted"    // 已过账/已提交审核
	VoucherStatusVerified VoucherStatus = "verified"  // 已审核/已核准
	VoucherStatusCancelled VoucherStatus = "cancelled" // 已作废
)

// VoucherAction represents an action that triggers a status transition.
type VoucherAction string

const (
	VoucherActionSubmit  VoucherAction = "submit"  // 提交审核
	VoucherActionApprove VoucherAction = "approve" // 核准过账
	VoucherActionReject  VoucherAction = "reject"  // 驳回
	VoucherActionReverse VoucherAction = "reverse" // 红字冲销
	VoucherActionCancel  VoucherAction = "cancel"  // 作废
)

// VoucherStateTransition records a state transition for audit purposes.
type VoucherStateTransition struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	VoucherID    uuid.UUID     `json:"voucher_id" db:"voucher_id"`
	TenantID     uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	FromStatus   VoucherStatus `json:"from_status" db:"from_status"`
	ToStatus     VoucherStatus `json:"to_status" db:"to_status"`
	Action       VoucherAction `json:"action" db:"action"`
	ChangedBy    uuid.UUID     `json:"changed_by" db:"changed_by"`
	ChangedByName *string      `json:"changed_by_name,omitempty" db:"changed_by_name"`
	Reason       *string       `json:"reason,omitempty" db:"reason"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

// JournalEntry represents the journal_entries table (voucher master).
type JournalEntry struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	VoucherNo    string          `json:"voucher_no" db:"voucher_no"`
	VoucherType  *string         `json:"voucher_type,omitempty" db:"voucher_type"`
	PostingDate  time.Time       `json:"posting_date" db:"posting_date"`
	CompanyID    uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Remark       *string         `json:"remark,omitempty" db:"remark"`
	DocStatus    int16           `json:"docstatus" db:"docstatus"`
	ReversedID   *uuid.UUID      `json:"reversed_id,omitempty" db:"reversed_id"`
	ReversalID   *uuid.UUID      `json:"reversal_id,omitempty" db:"reversal_id"`
	SubmittedBy  *uuid.UUID      `json:"submitted_by,omitempty" db:"submitted_by"`
	SubmittedAt  *time.Time      `json:"submitted_at,omitempty" db:"submitted_at"`
	CreatedBy    uuid.UUID       `json:"created_by" db:"created_by"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
	DebitTotal   decimal.Decimal `json:"debit_total,omitempty" db:"debit_total"`
	CreditTotal  decimal.Decimal `json:"credit_total,omitempty" db:"credit_total"`
}

// JournalEntryLine represents the journal_entry_lines table.
type JournalEntryLine struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	JournalEntryID uuid.UUID       `json:"journal_entry_id" db:"journal_entry_id"`
	AccountID      uuid.UUID       `json:"account_id" db:"account_id"`
	Debit          decimal.Decimal `json:"debit" db:"debit"`
	Credit         decimal.Decimal `json:"credit" db:"credit"`
	DebitCcy       decimal.Decimal `json:"debit_ccy" db:"debit_ccy"`
	CreditCcy      decimal.Decimal `json:"credit_ccy" db:"credit_ccy"`
	AccountCcy     *string         `json:"account_ccy,omitempty" db:"account_ccy"`
	ExchangeRate   decimal.Decimal `json:"exchange_rate" db:"exchange_rate"`
	PartyType      *string         `json:"party_type,omitempty" db:"party_type"`
	PartyID        *uuid.UUID      `json:"party_id,omitempty" db:"party_id"`
	CostCenterID   *uuid.UUID      `json:"cost_center_id,omitempty" db:"cost_center_id"`
	ProjectID      *uuid.UUID      `json:"project_id,omitempty" db:"project_id"`
	UserRemark     *string         `json:"user_remark,omitempty" db:"user_remark"`
	Reconciled     bool            `json:"reconciled" db:"reconciled"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	AccountCode    string          `json:"account_code,omitempty" db:"account_code"`
	AccountName    string          `json:"account_name,omitempty" db:"account_name"`
}
