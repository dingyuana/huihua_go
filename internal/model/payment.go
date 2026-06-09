package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentStatus represents the status of a payment entry.
type PaymentStatus int16

const (
	PaymentStatusDraft     PaymentStatus = 0 // 草稿
	PaymentStatusSubmitted PaymentStatus = 1 // 已提交
	PaymentStatusApproved  PaymentStatus = 2 // 已审核
	PaymentStatusPosted    PaymentStatus = 3 // 已过账（生成凭证）
	PaymentStatusCancelled PaymentStatus = 4 // 已作废
)

// PaymentAction represents an action that triggers a status transition.
type PaymentAction string

const (
	PaymentActionSubmit          PaymentAction = "submit"
	PaymentActionApprove         PaymentAction = "approve"
	PaymentActionReject          PaymentAction = "reject"
	PaymentActionCancel          PaymentAction = "cancel"
	PaymentActionGenerateVoucher PaymentAction = "generate_voucher"
	PaymentActionReverse         PaymentAction = "reverse"
)

// PaymentEntry represents the payment_entries table.
type PaymentEntry struct {
	ID               uuid.UUID        `json:"id" db:"id"`
	PaymentNo        string           `json:"payment_no" db:"payment_no"`
	PaymentType      string           `json:"payment_type" db:"payment_type"`
	PartyType        string           `json:"party_type" db:"party_type"`
	PartyID          uuid.UUID        `json:"party_id" db:"party_id"`
	CounterpartyName *string          `json:"counterparty_name,omitempty" db:"counterparty_name"`
	PaidFromID       *uuid.UUID       `json:"paid_from_id,omitempty" db:"paid_from_id"`
	PaidToID         *uuid.UUID       `json:"paid_to_id,omitempty" db:"paid_to_id"`
	PaidAmount       decimal.Decimal  `json:"paid_amount" db:"paid_amount"`
	ReceivedAmount   *decimal.Decimal `json:"received_amount,omitempty" db:"received_amount"`
	// UnallocatedAmount 收款/付款单未核销金额（独立字段，P1 拆分）
	// 等于 paid_amount - SUM(payment_allocations.allocated_amount)
	UnallocatedAmount decimal.Decimal `json:"unallocated_amount" db:"unallocated_amount"`
	// Currency 币种代码（ISO 4217，V1.1 默认 CNY，V2 远期多币种扩展）
	Currency      *string         `json:"currency,omitempty" db:"currency"`
	ExchangeRate  decimal.Decimal `json:"exchange_rate" db:"exchange_rate"`
	ReferenceNo   *string         `json:"reference_no,omitempty" db:"reference_no"`
	ReferenceDate *time.Time      `json:"reference_date,omitempty" db:"reference_date"`
	PostingDate   time.Time       `json:"posting_date" db:"posting_date"`
	CompanyID     uuid.UUID       `json:"company_id" db:"company_id"`
	TenantID      uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	BankAccountID *uuid.UUID      `json:"bank_account_id,omitempty" db:"bank_account_id"`
	DocStatus     int16           `json:"docstatus" db:"docstatus"`
	VoucherID     *uuid.UUID      `json:"voucher_id,omitempty" db:"voucher_id"`
	VoucherNo     *string         `json:"voucher_no,omitempty" db:"voucher_no"`
	Description   *string         `json:"description,omitempty" db:"description"`
	PaymentMethod *string         `json:"payment_method,omitempty" db:"payment_method"`
	CreatedBy     *uuid.UUID      `json:"created_by,omitempty" db:"created_by"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// PaymentAllocation represents the payment_allocations table (linking payments to invoices).
type PaymentAllocation struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	PaymentEntryID  uuid.UUID       `json:"payment_entry_id" db:"payment_entry_id"`
	InvoiceID       uuid.UUID       `json:"invoice_id" db:"invoice_id"`
	InvoiceType     *string         `json:"invoice_type,omitempty" db:"invoice_type"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount" db:"allocated_amount"`
	// DiscountAmount 现金折扣金额（采购 V1.0 §3.7）：折扣期内付款可享受的现金折扣
	// 实际付款 = allocated_amount - discount_amount
	DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	ReversedAt     *time.Time      `json:"reversed_at,omitempty" db:"reversed_at"`
}
