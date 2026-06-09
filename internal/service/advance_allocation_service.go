package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

type AdvanceAllocationService struct {
	receiptRepo      *repository.AdvanceReceiptRepository
	paymentRepo      *repository.AdvancePaymentRepository
	allocRepo        *repository.AdvanceAllocationRepository
	arRepo           *repository.ArInvoiceRepository
	apRepo           *repository.ApInvoiceRepository
	voucherSvc       *VoucherAutoGenerateService
	settlementLogRepo *repository.SettlementLogRepository
}

func NewAdvanceAllocationService(
	receiptRepo *repository.AdvanceReceiptRepository,
	paymentRepo *repository.AdvancePaymentRepository,
	allocRepo *repository.AdvanceAllocationRepository,
	arRepo *repository.ArInvoiceRepository,
	apRepo *repository.ApInvoiceRepository,
	voucherSvc *VoucherAutoGenerateService,
) *AdvanceAllocationService {
	return &AdvanceAllocationService{
		receiptRepo: receiptRepo, paymentRepo: paymentRepo, allocRepo: allocRepo,
		arRepo: arRepo, apRepo: apRepo, voucherSvc: voucherSvc,
	}
}

// InjectSettlementLogRepo injects the settlement log repository (used by main.go after initialization).
func (s *AdvanceAllocationService) InjectSettlementLogRepo(repo *repository.SettlementLogRepository) {
	s.settlementLogRepo = repo
}

type AllocateRequest struct {
	AdvanceID      uuid.UUID       `json:"advance_id"`
	AdvanceType    string          `json:"advance_type"`
	TargetID       uuid.UUID       `json:"target_id"`
	TargetType     string          `json:"target_type"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	AllocationDate time.Time       `json:"allocation_date"`
	Remark         *string         `json:"remark,omitempty"`
}

// Allocate performs an advance-to-invoice allocation with pessimistic locking.
// The full mutation sequence runs inside a single DB transaction:
//   1) Lock the advance row, then lock the target AR/AP row.
//   2) Re-read both inside the transaction to get fresh balances.
//   3) Insert advance_allocation row.
//   4) Update balances + statuses.
//   5) Write settlement_log for audit trail.
//   6) Commit.
//
// Voucher generation runs after commit as a best-effort side effect.
// A voucher failure does NOT roll back the allocation.
func (s *AdvanceAllocationService) Allocate(ctx context.Context, tenantID, userID uuid.UUID, req *AllocateRequest) (*model.AdvanceAllocation, error) {
	if err := s.validateAllocateRequest(req); err != nil {
		return nil, err
	}
	if req.AllocationDate.IsZero() {
		req.AllocationDate = time.Now()
	}

	var advanceAvail decimal.Decimal
	if req.AdvanceType == "receipt" {
		adv, err := s.receiptRepo.GetByID(ctx, tenantID, req.AdvanceID)
		if err != nil {
			return nil, err
		}
		if adv == nil {
			return nil, errors.New("advance receipt not found")
		}
		if adv.Status != string(model.AdvanceReceiptStatusConfirmed) &&
			adv.Status != string(model.AdvanceReceiptStatusPartiallyAllocated) {
			return nil, fmt.Errorf("advance receipt not in allocatable state: %s", adv.Status)
		}
		advanceAvail = adv.OutstandingAmount
	} else {
		adv, err := s.paymentRepo.GetByID(ctx, tenantID, req.AdvanceID)
		if err != nil {
			return nil, err
		}
		if adv == nil {
			return nil, errors.New("advance payment not found")
		}
		if adv.Status != string(model.AdvancePaymentStatusConfirmed) &&
			adv.Status != string(model.AdvancePaymentStatusPartiallyAllocated) {
			return nil, fmt.Errorf("advance payment not in allocatable state: %s", adv.Status)
		}
		advanceAvail = adv.OutstandingAmount
	}
	if req.AllocatedAmount.GreaterThan(advanceAvail) {
		return nil, fmt.Errorf("allocated amount %s exceeds advance outstanding %s",
			req.AllocatedAmount.String(), advanceAvail.String())
	}

	var targetAvail decimal.Decimal
	if req.TargetType == "ar" {
		ar, err := s.arRepo.GetByID(ctx, tenantID, req.TargetID)
		if err != nil {
			return nil, err
		}
		if ar == nil {
			return nil, errors.New("ar invoice not found")
		}
		if ar.Status != "confirmed" && ar.Status != "partially_paid" {
			return nil, fmt.Errorf("ar invoice not allocatable: %s", ar.Status)
		}
		targetAvail = ar.OutstandingAmount
	} else {
		ap, err := s.apRepo.GetByID(ctx, tenantID, req.TargetID)
		if err != nil {
			return nil, err
		}
		if ap == nil {
			return nil, errors.New("ap invoice not found")
		}
		if ap.Status != "confirmed" && ap.Status != "partially_paid" {
			return nil, fmt.Errorf("ap invoice not allocatable: %s", ap.Status)
		}
		targetAvail = ap.OutstandingAmount
	}
	if req.AllocatedAmount.GreaterThan(targetAvail) {
		return nil, fmt.Errorf("allocated amount %s exceeds target outstanding %s",
			req.AllocatedAmount.String(), targetAvail.String())
	}

	// Open transaction on the receipt or payment pool.
	var tx pgx.Tx
	if req.AdvanceType == "receipt" {
		t, err := s.receiptRepo.BeginTx(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		tx = t
	} else {
		t, err := s.paymentRepo.BeginTx(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		tx = t
	}
	defer tx.Rollback(ctx)

	if req.AdvanceType == "receipt" {
		if err := s.receiptRepo.LockForUpdate(ctx, tx, tenantID, req.AdvanceID); err != nil {
			return nil, fmt.Errorf("lock advance receipt: %w", err)
		}
	} else {
		if err := s.paymentRepo.LockForUpdate(ctx, tx, tenantID, req.AdvanceID); err != nil {
			return nil, fmt.Errorf("lock advance payment: %w", err)
		}
	}
	if req.TargetType == "ar" {
		if err := s.arRepo.LockForUpdate(ctx, tx, tenantID, req.TargetID); err != nil {
			return nil, fmt.Errorf("lock ar invoice: %w", err)
		}
	} else {
		if err := s.apRepo.LockForUpdate(ctx, tx, tenantID, req.TargetID); err != nil {
			return nil, fmt.Errorf("lock ap invoice: %w", err)
		}
	}

	// Re-read inside the transaction to get fresh outstanding balances.
	var advanceOutAfter decimal.Decimal
	if req.AdvanceType == "receipt" {
		adv, err := s.receiptRepo.GetByID(ctx, tenantID, req.AdvanceID)
		if err != nil {
			return nil, fmt.Errorf("re-read advance receipt: %w", err)
		}
		if adv == nil || adv.OutstandingAmount.LessThan(req.AllocatedAmount) {
			return nil, fmt.Errorf("allocated amount %s exceeds advance outstanding (concurrent update)",
				req.AllocatedAmount.String())
		}
		advanceOutAfter = adv.OutstandingAmount.Sub(req.AllocatedAmount)
	} else {
		adv, err := s.paymentRepo.GetByID(ctx, tenantID, req.AdvanceID)
		if err != nil {
			return nil, fmt.Errorf("re-read advance payment: %w", err)
		}
		if adv == nil || adv.OutstandingAmount.LessThan(req.AllocatedAmount) {
			return nil, fmt.Errorf("allocated amount %s exceeds advance outstanding (concurrent update)",
				req.AllocatedAmount.String())
		}
		advanceOutAfter = adv.OutstandingAmount.Sub(req.AllocatedAmount)
	}

	var targetOutAfter decimal.Decimal
	if req.TargetType == "ar" {
		ar, err := s.arRepo.GetByID(ctx, tenantID, req.TargetID)
		if err != nil {
			return nil, fmt.Errorf("re-read ar invoice: %w", err)
		}
		if ar == nil || ar.OutstandingAmount.LessThan(req.AllocatedAmount) {
			return nil, fmt.Errorf("allocated amount %s exceeds ar outstanding (concurrent update)",
				req.AllocatedAmount.String())
		}
		targetOutAfter = ar.OutstandingAmount.Sub(req.AllocatedAmount)
	} else {
		ap, err := s.apRepo.GetByID(ctx, tenantID, req.TargetID)
		if err != nil {
			return nil, fmt.Errorf("re-read ap invoice: %w", err)
		}
		if ap == nil || ap.OutstandingAmount.LessThan(req.AllocatedAmount) {
			return nil, fmt.Errorf("allocated amount %s exceeds ap outstanding (concurrent update)",
				req.AllocatedAmount.String())
		}
		targetOutAfter = ap.OutstandingAmount.Sub(req.AllocatedAmount)
	}

	alloc := &model.AdvanceAllocation{
		ID:              uuid.New(),
		TenantID:        tenantID,
		AdvanceID:       req.AdvanceID,
		AdvanceType:     req.AdvanceType,
		TargetID:        req.TargetID,
		TargetType:      req.TargetType,
		AllocatedAmount: req.AllocatedAmount,
		AllocationDate:  req.AllocationDate,
		Remark:          req.Remark,
		CreatedBy:       &userID,
		CreatedAt:       time.Now(),
	}
	if err := s.allocRepo.CreateTx(ctx, tx, alloc); err != nil {
		return nil, fmt.Errorf("create allocation: %w", err)
	}

	delta := req.AllocatedAmount.String()
	var newAdvanceStatus string
	if req.AdvanceType == "receipt" {
		if err := s.receiptRepo.IncrementAllocatedTx(ctx, tx, tenantID, req.AdvanceID, delta); err != nil {
			return nil, fmt.Errorf("update advance receipt (tx): %w", err)
		}
		newAdvanceStatus = string(model.AdvanceReceiptStatusFullyAllocated)
		if advanceOutAfter.IsPositive() {
			newAdvanceStatus = string(model.AdvanceReceiptStatusPartiallyAllocated)
		}
		if _, err := tx.Exec(ctx, `UPDATE advance_receipts SET status = $3 WHERE tenant_id = $1 AND id = $2`,
			tenantID, req.AdvanceID, newAdvanceStatus); err != nil {
			return nil, fmt.Errorf("update advance receipt status (tx): %w", err)
		}
	} else {
		if err := s.paymentRepo.IncrementAllocatedTx(ctx, tx, tenantID, req.AdvanceID, delta); err != nil {
			return nil, fmt.Errorf("update advance payment (tx): %w", err)
		}
		newAdvanceStatus = string(model.AdvancePaymentStatusFullyAllocated)
		if advanceOutAfter.IsPositive() {
			newAdvanceStatus = string(model.AdvancePaymentStatusPartiallyAllocated)
		}
		if _, err := tx.Exec(ctx, `UPDATE advance_payments SET status = $3 WHERE tenant_id = $1 AND id = $2`,
			tenantID, req.AdvanceID, newAdvanceStatus); err != nil {
			return nil, fmt.Errorf("update advance payment status (tx): %w", err)
		}
	}

	if req.TargetType == "ar" {
		newArStatus := string(model.ArInvoiceStatusPaid)
		if targetOutAfter.IsPositive() {
			newArStatus = string(model.ArInvoiceStatusPartiallyPaid)
		}
		if err := s.arRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, req.TargetID, targetOutAfter); err != nil {
			return nil, fmt.Errorf("update ar outstanding (tx): %w", err)
		}
		if err := s.arRepo.UpdateStatusTx(ctx, tx, tenantID, req.TargetID, newArStatus); err != nil {
			return nil, fmt.Errorf("update ar status (tx): %w", err)
		}
	} else {
		newApStatus := string(model.ApInvoiceStatusPaid)
		if targetOutAfter.IsPositive() {
			newApStatus = string(model.ApInvoiceStatusPartiallyPaid)
		}
		if err := s.apRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, req.TargetID, targetOutAfter); err != nil {
			return nil, fmt.Errorf("update ap outstanding (tx): %w", err)
		}
		if err := s.apRepo.UpdateStatusTx(ctx, tx, tenantID, req.TargetID, newApStatus); err != nil {
			return nil, fmt.Errorf("update ap status (tx): %w", err)
		}
	}

	if s.settlementLogRepo != nil {
		var docType model.SettlementLogDocType
		var direction model.SettlementLogDirection
		if req.TargetType == "ar" {
			docType = model.SettlementLogDocArInvoice
			direction = model.SettlementLogDirectionDebit
		} else {
			docType = model.SettlementLogDocApInvoice
			direction = model.SettlementLogDirectionCredit
		}
		if err := repository.LogWriteOff(
			ctx, tx, s.settlementLogRepo,
			tenantID, alloc.ID, req.TargetID,
			model.SettlementLogSourceAdvanceAllocation,
			docType, direction,
			req.AllocatedAmount,
			targetAvail, targetOutAfter,
			&userID,
		); err != nil {
			return nil, fmt.Errorf("write settlement log (target): %w", err)
		}
		var advanceDocType model.SettlementLogDocType
		var advDirection model.SettlementLogDirection
		if req.AdvanceType == "receipt" {
			advanceDocType = model.SettlementLogDocAdvanceReceipt
			advDirection = model.SettlementLogDirectionCredit
		} else {
			advanceDocType = model.SettlementLogDocAdvancePayment
			advDirection = model.SettlementLogDirectionDebit
		}
		if err := repository.LogWriteOff(
			ctx, tx, s.settlementLogRepo,
			tenantID, alloc.ID, req.AdvanceID,
			model.SettlementLogSourceAdvanceAllocation,
			advanceDocType, advDirection,
			req.AllocatedAmount,
			advanceAvail, advanceOutAfter,
			&userID,
		); err != nil {
			return nil, fmt.Errorf("write settlement log (advance): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Voucher generation as a best-effort post-commit side effect.
	if s.voucherSvc != nil {
		if voucherNo, err := s.voucherSvc.GenerateFromAdvanceOffset(ctx, tenantID, req.AdvanceID, req.TargetID, req.AllocatedAmount, userID); err == nil {
			_ = s.allocRepo.SetVoucher(ctx, alloc.ID, voucherNo)
			alloc.VoucherNo = &voucherNo
		}
	}

	return alloc, nil
}

func (s *AdvanceAllocationService) validateAllocateRequest(req *AllocateRequest) error {
	if req.AllocatedAmount.LessThanOrEqual(decimal.Zero) {
		return errors.New("allocated_amount must be positive")
	}
	if req.AdvanceID == uuid.Nil || req.TargetID == uuid.Nil {
		return errors.New("advance_id and target_id required")
	}
	if req.AdvanceType != "receipt" && req.AdvanceType != "payment" {
		return errors.New("advance_type must be 'receipt' or 'payment'")
	}
	if req.TargetType != "ar" && req.TargetType != "ap" {
		return errors.New("target_type must be 'ar' or 'ap'")
	}
	if req.AdvanceType == "receipt" && req.TargetType != "ar" {
		return errors.New("receipt advance can only offset ar target")
	}
	if req.AdvanceType == "payment" && req.TargetType != "ap" {
		return errors.New("payment advance can only offset ap target")
	}
	return nil
}

func (s *AdvanceAllocationService) ListByAdvance(ctx context.Context, tenantID, advanceID uuid.UUID) ([]*model.AdvanceAllocation, error) {
	return s.allocRepo.ListByAdvance(ctx, tenantID, advanceID)
}

func (s *AdvanceAllocationService) ListByTarget(ctx context.Context, tenantID, targetID uuid.UUID) ([]*model.AdvanceAllocation, error) {
	return s.allocRepo.ListByTarget(ctx, tenantID, targetID)
}

func (s *AdvanceAllocationService) AutoMatch(ctx context.Context, tenantID, userID, advanceID uuid.UUID) ([]*model.AdvanceAllocation, error) {
	adv, err := s.receiptRepo.GetByID(ctx, tenantID, advanceID)
	var partyID uuid.UUID
	var targetType string
	if adv != nil && (adv.Status == string(model.AdvanceReceiptStatusConfirmed) ||
		adv.Status == string(model.AdvanceReceiptStatusPartiallyAllocated)) {
		partyID = adv.CustomerID
		targetType = "ar"
	} else {
		ap, err2 := s.paymentRepo.GetByID(ctx, tenantID, advanceID)
		if err2 != nil || ap == nil {
			return nil, errors.New("advance not found or not allocatable")
		}
		if ap.Status != string(model.AdvancePaymentStatusConfirmed) &&
			ap.Status != string(model.AdvancePaymentStatusPartiallyAllocated) {
			return nil, errors.New("advance payment not in allocatable state")
		}
		partyID = ap.SupplierID
		targetType = "ap"
	}
	if err != nil && adv == nil {
		return nil, err
	}

	var targets []struct {
		ID            uuid.UUID
		Outstanding   decimal.Decimal
	}
	if targetType == "ar" {
		ars, err := s.arRepo.ListUnpaidByCustomer(ctx, tenantID, partyID)
		if err != nil {
			return nil, err
		}
		for _, ar := range ars {
			targets = append(targets, struct {
				ID          uuid.UUID
				Outstanding decimal.Decimal
			}{ar.ID, ar.OutstandingAmount})
		}
	} else {
		aps, err := s.apRepo.ListUnpaidBySupplier(ctx, tenantID, partyID)
		if err != nil {
			return nil, err
		}
		for _, ap := range aps {
			targets = append(targets, struct {
				ID          uuid.UUID
				Outstanding decimal.Decimal
			}{ap.ID, ap.OutstandingAmount})
		}
	}

	var advanceAvail decimal.Decimal
	if targetType == "ar" {
		advanceAvail = adv.OutstandingAmount
	} else {
		var ap2 *model.AdvancePayment
		ap2, _ = s.paymentRepo.GetByID(ctx, tenantID, advanceID)
		advanceAvail = ap2.OutstandingAmount
	}

	var results []*model.AdvanceAllocation
	for _, t := range targets {
		if !advanceAvail.IsPositive() {
			break
		}
		allocAmt := advanceAvail
		if t.Outstanding.LessThan(advanceAvail) {
			allocAmt = t.Outstanding
		}
		r, err := s.Allocate(ctx, tenantID, userID, &AllocateRequest{
			AdvanceID:      advanceID,
			AdvanceType:    targetType,
			TargetID:       t.ID,
			TargetType:     targetType,
			AllocatedAmount: allocAmt,
			AllocationDate: time.Now(),
		})
		if err != nil {
			break
		}
		results = append(results, r)
		advanceAvail = advanceAvail.Sub(allocAmt)
	}
	return results, nil
}
