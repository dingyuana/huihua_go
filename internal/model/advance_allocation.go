package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AdvanceAllocation struct {
	ID              uuid.UUID       `db:"id"`
	TenantID        uuid.UUID       `db:"tenant_id"`
	AdvanceID       uuid.UUID       `db:"advance_id"`
	AdvanceType     string          `db:"advance_type"`
	TargetID        uuid.UUID       `db:"target_id"`
	TargetType      string          `db:"target_type"`
	AllocatedAmount decimal.Decimal `db:"allocated_amount"`
	AllocationDate  time.Time       `db:"allocation_date"`
	VoucherID       *uuid.UUID      `db:"voucher_id"`
	VoucherNo       *string         `db:"voucher_no"`
	Remark          *string         `db:"remark"`
	CreatedBy       *uuid.UUID      `db:"created_by"`
	CreatedAt       time.Time       `db:"created_at"`
}
