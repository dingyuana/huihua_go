package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	WriteOffTypeReceiptAR = "receipt_ar"
	WriteOffTypePaymentAP = "payment_ap"
	WriteOffTypePrepaidAR = "prepaid_ar"

	WriteOffStatusDraft          = "draft"
	WriteOffStatusPendingApproval = "pending_approval"
	WriteOffStatusApproved       = "approved"
	WriteOffStatusRejected       = "rejected"
	WriteOffStatusReversed       = "reversed"
)

type WriteOffRecord struct {
	ID                   int64           `db:"id"`
	TenantID             uuid.UUID       `db:"tenant_id"`
	WriteOffNo           string          `db:"write_off_no"`
	Type                 string          `db:"type"`
	ReceiptPaymentID     uuid.UUID       `db:"receipt_payment_id"`
	ReceivablePayableID  uuid.UUID       `db:"receivable_payable_id"`
	ReceivablePayableType string         `db:"receivable_payable_type"`
	Amount               decimal.Decimal `db:"amount"`
	DiffAmount           decimal.Decimal `db:"diff_amount"`
	DiffAccountCode      string          `db:"diff_account_code"`
	WriteOffDate         time.Time       `db:"write_off_date"`
	Operator             *uuid.UUID      `db:"operator"`
	Status               string          `db:"status"`
	Remark               *string         `db:"remark"`
	MatchRule            *string         `db:"match_rule"`
	Approver             *uuid.UUID      `db:"approver"`
	ApprovedAt           *time.Time      `db:"approved_at"`
	RejectReason         *string         `db:"reject_reason"`
	CreatedAt            time.Time       `db:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at"`
}

type WriteOffRule struct {
	ID                int64           `db:"id"`
	TenantID          uuid.UUID       `db:"tenant_id"`
	RuleName          string          `db:"rule_name"`
	RuleType          string          `db:"rule_type"`
	Priority          int             `db:"priority"`
	Enabled           bool            `db:"enabled"`
	ToleranceAmount   string          `db:"tolerance_amount"`
	TolerancePercent  string          `db:"tolerance_percent"`
	DateWindow        int             `db:"date_window"`
	DiffAccountCode   string          `db:"diff_account_code"`
	Description       *string         `db:"description"`
	CreatedAt         time.Time       `db:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at"`
}

type WriteOffUnmatchedItem struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	DocumentType    string          `json:"document_type" db:"document_type"`
	DocumentNo      string          `json:"document_no" db:"document_no"`
	CounterpartyID  uuid.UUID       `json:"counterparty_id" db:"counterparty_id"`
	CounterpartyName string         `json:"counterparty_name" db:"counterparty_name"`
	Amount          decimal.Decimal `json:"amount" db:"amount"`
	RemainingAmount decimal.Decimal `json:"remaining_amount" db:"remaining_amount"`
	DocumentDate    *time.Time      `json:"document_date" db:"document_date"`
	DueDate         *time.Time      `json:"due_date" db:"due_date"`
	Description     *string         `json:"description" db:"description"`
	UnmatchedReason *string         `json:"unmatched_reason" db:"unmatched_reason"`
}

