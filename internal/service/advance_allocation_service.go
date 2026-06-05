package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

type AdvanceAllocationService struct {
	receiptRepo  *repository.AdvanceReceiptRepository
	paymentRepo  *repository.AdvancePaymentRepository
	allocRepo    *repository.AdvanceAllocationRepository
	arRepo       *repository.ArInvoiceRepository
	apRepo       *repository.ApInvoiceRepository
	voucherSvc   *VoucherAutoGenerateService
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

type AllocateRequest struct {
	AdvanceID      uuid.UUID       `json:"advance_id"`
	AdvanceType    string          `json:"advance_type"`
	TargetID       uuid.UUID       `json:"target_id"`
	TargetType     string          `json:"target_type"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount"`
	AllocationDate time.Time       `json:"allocation_date"`
	Remark         *string         `json:"remark,omitempty"`
}

func (s *AdvanceAllocationService) Allocate(ctx context.Context, tenantID, userID uuid.UUID, req *AllocateRequest) (*model.AdvanceAllocation, error) {
	if req.AllocatedAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("allocated_amount must be positive")
	}
	if req.AdvanceID == uuid.Nil || req.TargetID == uuid.Nil {
		return nil, errors.New("advance_id and target_id required")
	}
	if req.AdvanceType != "receipt" && req.AdvanceType != "payment" {
		return nil, errors.New("advance_type must be 'receipt' or 'payment'")
	}
	if req.TargetType != "ar" && req.TargetType != "ap" {
		return nil, errors.New("target_type must be 'ar' or 'ap'")
	}
	if req.AdvanceType == "receipt" && req.TargetType != "ar" {
		return nil, errors.New("receipt advance can only offset ar target")
	}
	if req.AdvanceType == "payment" && req.TargetType != "ap" {
		return nil, errors.New("payment advance can only offset ap target")
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
	if err := s.allocRepo.Create(ctx, alloc); err != nil {
		return nil, err
	}

	delta := req.AllocatedAmount.InexactFloat64()

	if req.AdvanceType == "receipt" {
		if err := s.receiptRepo.IncrementAllocated(ctx, tenantID, req.AdvanceID, delta); err != nil {
			return nil, fmt.Errorf("update advance receipt: %w", err)
		}
		newStatus := string(model.AdvanceReceiptStatusFullyAllocated)
		if advanceAvail.Sub(req.AllocatedAmount).IsPositive() {
			newStatus = string(model.AdvanceReceiptStatusPartiallyAllocated)
		}
		if err := s.receiptRepo.UpdateStatus(ctx, tenantID, req.AdvanceID, newStatus); err != nil {
			return nil, err
		}
	} else {
		if err := s.paymentRepo.IncrementAllocated(ctx, tenantID, req.AdvanceID, delta); err != nil {
			return nil, fmt.Errorf("update advance payment: %w", err)
		}
		newStatus := string(model.AdvancePaymentStatusFullyAllocated)
		if advanceAvail.Sub(req.AllocatedAmount).IsPositive() {
			newStatus = string(model.AdvancePaymentStatusPartiallyAllocated)
		}
		if err := s.paymentRepo.UpdateStatus(ctx, tenantID, req.AdvanceID, newStatus); err != nil {
			return nil, err
		}
	}

	if req.TargetType == "ar" {
		newArStatus := "paid"
		if targetAvail.Sub(req.AllocatedAmount).IsPositive() {
			newArStatus = "partially_paid"
		}
		if err := s.arRepo.IncrementPaid(ctx, tenantID, req.TargetID, delta, newArStatus); err != nil {
			return nil, fmt.Errorf("update ar paid: %w", err)
		}
	} else {
		newApStatus := "paid"
		if targetAvail.Sub(req.AllocatedAmount).IsPositive() {
			newApStatus = "partially_paid"
		}
		if err := s.apRepo.IncrementPaid(ctx, tenantID, req.TargetID, delta, newApStatus); err != nil {
			return nil, fmt.Errorf("update ap paid: %w", err)
		}
	}

	if s.voucherSvc != nil {
		if voucherNo, err := s.voucherSvc.GenerateFromAdvanceOffset(ctx, tenantID, req.AdvanceID, req.TargetID, req.AllocatedAmount, userID); err == nil {
			_ = s.allocRepo.SetVoucher(ctx, alloc.ID, voucherNo)
			alloc.VoucherNo = &voucherNo
		}
	}

	return alloc, nil
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
