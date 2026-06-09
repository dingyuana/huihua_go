package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SettlementLogSourceType identifies what generated this settlement log.
type SettlementLogSourceType string

const (
	SettlementLogSourcePaymentAllocation SettlementLogSourceType = "payment_allocation"
	SettlementLogSourceAdvanceAllocation SettlementLogSourceType = "advance_allocation"
	SettlementLogSourceManualReversal    SettlementLogSourceType = "manual_reversal"
)

// SettlementLogDocType identifies the document type being settled.
type SettlementLogDocType string

const (
	SettlementLogDocSalesInvoice   SettlementLogDocType = "sales_invoice"
	SettlementLogDocArInvoice      SettlementLogDocType = "ar_invoice"
	SettlementLogDocApInvoice      SettlementLogDocType = "ap_invoice"
	SettlementLogDocAdvanceReceipt SettlementLogDocType = "advance_receipt"
	SettlementLogDocAdvancePayment SettlementLogDocType = "advance_payment"
)

// SettlementLogDirection indicates whether the settlement reduces a debit or credit.
type SettlementLogDirection string

const (
	SettlementLogDirectionDebit  SettlementLogDirection = "debit"  // 债权减少
	SettlementLogDirectionCredit SettlementLogDirection = "credit" // 债务减少
)

// SettlementLog represents an immutable settlement audit log entry.
// Each record records one atomic settlement or reversal operation.
type SettlementLog struct {
	ID                uuid.UUID                `json:"id" db:"id"`
	TenantID          uuid.UUID                `json:"tenant_id" db:"tenant_id"`
	SourceType        SettlementLogSourceType  `json:"source_type" db:"source_type"`
	SourceID          uuid.UUID                `json:"source_id" db:"source_id"`
	DocType           SettlementLogDocType     `json:"doc_type" db:"doc_type"`
	DocID             uuid.UUID                `json:"doc_id" db:"doc_id"`
	Direction         SettlementLogDirection   `json:"direction" db:"direction"`
	Amount            decimal.Decimal          `json:"amount" db:"amount"`
	OutstandingBefore decimal.Decimal          `json:"outstanding_before" db:"outstanding_before"`
	OutstandingAfter  decimal.Decimal          `json:"outstanding_after" db:"outstanding_after"`
	IsReversal        bool                     `json:"is_reversal" db:"is_reversal"`
	ReversedLogID     *uuid.UUID               `json:"reversed_log_id,omitempty" db:"reversed_log_id"`
	CreatedBy         *uuid.UUID               `json:"created_by,omitempty" db:"created_by"`
	CreatedAt         time.Time                `json:"created_at" db:"created_at"`
}
