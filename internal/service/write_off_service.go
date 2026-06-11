package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

type WriteOffService struct {
	writeOffRepo   *repository.WriteOffRepository
	paymentRepo    *repository.PaymentEntryRepository
	arInvoiceRepo  *repository.ArInvoiceRepository
	apInvoiceRepo  *repository.ApInvoiceRepository
	engine         *WriteOffEngine
}

func NewWriteOffService(
	writeOffRepo *repository.WriteOffRepository,
	paymentRepo *repository.PaymentEntryRepository,
	arInvoiceRepo *repository.ArInvoiceRepository,
	apInvoiceRepo *repository.ApInvoiceRepository,
	engine *WriteOffEngine,
) *WriteOffService {
	return &WriteOffService{
		writeOffRepo:   writeOffRepo,
		paymentRepo:    paymentRepo,
		arInvoiceRepo:  arInvoiceRepo,
		apInvoiceRepo:  apInvoiceRepo,
		engine:         engine,
	}
}

type AutoWriteOffOptions struct {
	DocumentType   string
	StartDate      time.Time
	EndDate        time.Time
	CounterpartyID uuid.UUID
}

type ManualWriteOffRequest struct {
	ReceiptPaymentID      uuid.UUID
	ReceivablePayableID   uuid.UUID
	ReceivablePayableType string
	Amount                decimal.Decimal
	Remark                string
}

type WriteOffResult struct {
	TotalMatched          int
	TotalAmount           decimal.Decimal
	FailedCount           int
	UnmatchedDocuments    []UnmatchedDocument
}

type UnmatchedDocument struct {
	DocumentID     uuid.UUID
	DocumentType   string
	Amount         decimal.Decimal
	Reason         string
}

type WriteOffUnmatchedSummary struct {
	TotalUnmatchedAmount decimal.Decimal
	OverdueAmount        decimal.Decimal
	ByCounterparty       []WriteOffCounterpartySummary
}

type WriteOffCounterpartySummary struct {
	CounterpartyID   uuid.UUID
	Name             string
	Amount           decimal.Decimal
}

type SubmitApprovalRequest struct {
	RecordID int64
	Remark   string
}

type ApproveRequest struct {
	RecordID int64
}

type RejectRequest struct {
	RecordID     int64
	RejectReason string
}

func (s *WriteOffService) AutoWriteOff(ctx context.Context, tenantID uuid.UUID, opts AutoWriteOffOptions) (*WriteOffResult, error) {
	if err := s.engine.LoadRules(ctx, tenantID); err != nil {
		return nil, err
	}

	var results []MatchResult
	var unmatchedDocs []UnmatchedDocument

	switch opts.DocumentType {
	case "payment_ar":
		payments, err := s.paymentRepo.ListByTenant(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		for _, payment := range payments {
			if payment.PaymentType != "receipt" {
				continue
			}
			if opts.CounterpartyID != uuid.Nil && payment.PartyID != opts.CounterpartyID {
				continue
			}
			if !opts.StartDate.IsZero() && payment.PostingDate.Before(opts.StartDate) {
				continue
			}
			if !opts.EndDate.IsZero() && payment.PostingDate.After(opts.EndDate) {
				continue
			}

			matches, failures, err := s.engine.MatchPaymentAR(ctx, tenantID, payment.ID)
			if err != nil {
				return nil, err
			}
			results = append(results, matches...)
			for _, f := range failures {
				unmatchedDocs = append(unmatchedDocs, UnmatchedDocument{
					DocumentID:   f.InvoiceID,
					DocumentType: "ar_invoice",
					Amount:       decimal.Zero,
					Reason:       strings.Join(f.FailureReasons, "; "),
				})
			}
		}

	case "payment_ap":
		return &WriteOffResult{
			TotalMatched:        0,
			TotalAmount:         decimal.Zero,
			FailedCount:         0,
			UnmatchedDocuments: nil,
		}, nil

	default:
		return nil, fmt.Errorf("invalid document type: %s", opts.DocumentType)
	}

	if len(results) == 0 {
		return &WriteOffResult{
			TotalMatched:        0,
			TotalAmount:         decimal.Zero,
			FailedCount:         len(unmatchedDocs),
			UnmatchedDocuments: unmatchedDocs,
		}, nil
	}

	records := s.engine.GenerateDraftWriteOffRecords(results, uuid.Nil)

	for _, record := range records {
		record.TenantID = tenantID
		if err := s.writeOffRepo.Create(ctx, record); err != nil {
			return nil, err
		}
	}

	totalAmount := decimal.Zero
	for _, record := range records {
		totalAmount = totalAmount.Add(record.Amount)
	}

	return &WriteOffResult{
		TotalMatched:        len(records),
		TotalAmount:         totalAmount,
		FailedCount:         len(unmatchedDocs),
		UnmatchedDocuments: unmatchedDocs,
	}, nil
}

func (s *WriteOffService) ManualWriteOff(ctx context.Context, tenantID, operatorID uuid.UUID, req ManualWriteOffRequest) (*model.WriteOffRecord, error) {
	payment, err := s.paymentRepo.GetByID(ctx, tenantID, req.ReceiptPaymentID)
	if err != nil {
		return nil, err
	}

	if payment.UnallocatedAmount.LessThan(req.Amount) {
		return nil, fmt.Errorf("payment remaining amount insufficient")
	}

	var outstandingAmount decimal.Decimal
	switch req.ReceivablePayableType {
	case "ar_invoice":
		invoice, err := s.arInvoiceRepo.GetByID(ctx, tenantID, req.ReceivablePayableID)
		if err != nil {
			return nil, err
		}
		if invoice.OutstandingAmount.LessThan(req.Amount) {
			return nil, fmt.Errorf("invoice outstanding amount insufficient")
		}
		outstandingAmount = invoice.OutstandingAmount
	case "ap_invoice":
		invoice, err := s.apInvoiceRepo.GetByID(ctx, tenantID, req.ReceivablePayableID)
		if err != nil {
			return nil, err
		}
		if invoice.OutstandingAmount.LessThan(req.Amount) {
			return nil, fmt.Errorf("invoice outstanding amount insufficient")
		}
		outstandingAmount = invoice.OutstandingAmount
	default:
		return nil, fmt.Errorf("invalid receivable type: %s", req.ReceivablePayableType)
	}

	diffAmount := req.Amount.Sub(outstandingAmount).Abs()
	var diffAccountCode string
	if diffAmount.GreaterThan(decimal.Zero) {
		diffAccountCode = "6603"
	}

	record := &model.WriteOffRecord{
		TenantID:              tenantID,
		WriteOffNo:            generateWriteOffNo(),
		Type:                  getWriteOffType(req.ReceivablePayableType),
		ReceiptPaymentID:      req.ReceiptPaymentID,
		ReceivablePayableID:   req.ReceivablePayableID,
		ReceivablePayableType: req.ReceivablePayableType,
		Amount:               req.Amount,
		DiffAmount:           diffAmount,
		DiffAccountCode:      diffAccountCode,
		WriteOffDate:         time.Now(),
		Operator:             &operatorID,
		Status:               model.WriteOffStatusDraft,
		Remark:               &req.Remark,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := s.writeOffRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return record, nil
}

func (s *WriteOffService) SubmitApproval(ctx context.Context, tenantID, operatorID uuid.UUID, req SubmitApprovalRequest) error {
	record, err := s.writeOffRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusDraft {
		return fmt.Errorf("only draft records can be submitted for approval")
	}

	now := time.Now()
	record.Status = model.WriteOffStatusPendingApproval
	record.UpdatedAt = now
	if req.Remark != "" {
		record.Remark = &req.Remark
	}

	return s.writeOffRepo.Update(ctx, record)
}

func (s *WriteOffService) Approve(ctx context.Context, tenantID, approverID uuid.UUID, req ApproveRequest) error {
	record, err := s.writeOffRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusPendingApproval {
		return fmt.Errorf("only pending approval records can be approved")
	}

	tx, err := s.writeOffRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.updateDocumentWriteOffStatus(ctx, tx, tenantID, record); err != nil {
		return err
	}

	if record.DiffAmount.GreaterThan(decimal.Zero) {
		if err := s.createDiffVoucher(ctx, tx, tenantID, record); err != nil {
			return err
		}
	}

	now := time.Now()
	record.Status = model.WriteOffStatusApproved
	record.Approver = &approverID
	record.ApprovedAt = &now
	record.UpdatedAt = now

	if err := s.writeOffRepo.UpdateTx(ctx, tx, record); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *WriteOffService) Reject(ctx context.Context, tenantID, approverID uuid.UUID, req RejectRequest) error {
	record, err := s.writeOffRepo.GetByID(ctx, req.RecordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusPendingApproval {
		return fmt.Errorf("only pending approval records can be rejected")
	}

	now := time.Now()
	record.Status = model.WriteOffStatusRejected
	record.Approver = &approverID
	record.ApprovedAt = &now
	record.RejectReason = &req.RejectReason
	record.UpdatedAt = now

	return s.writeOffRepo.Update(ctx, record)
}

func (s *WriteOffService) UpdateDraft(ctx context.Context, tenantID uuid.UUID, recordID int64, req ManualWriteOffRequest) error {
	record, err := s.writeOffRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusDraft && record.Status != model.WriteOffStatusRejected {
		return fmt.Errorf("only draft or rejected records can be updated")
	}

	record.ReceivablePayableID = req.ReceivablePayableID
	record.ReceivablePayableType = req.ReceivablePayableType
	record.Amount = req.Amount
	record.DiffAmount = req.Amount.Sub(record.Amount).Abs()
	if req.Remark != "" {
		record.Remark = &req.Remark
	}
	record.UpdatedAt = time.Now()
	record.Status = model.WriteOffStatusDraft
	record.RejectReason = nil

	return s.writeOffRepo.Update(ctx, record)
}

func (s *WriteOffService) DeleteDraft(ctx context.Context, tenantID uuid.UUID, recordID int64) error {
	record, err := s.writeOffRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusDraft && record.Status != model.WriteOffStatusRejected {
		return fmt.Errorf("only draft or rejected records can be deleted")
	}

	return s.writeOffRepo.Delete(ctx, recordID)
}

func (s *WriteOffService) updateDocumentWriteOffStatus(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, record *model.WriteOffRecord) error {
	switch record.ReceivablePayableType {
	case "ar_invoice":
		return s.arInvoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, record.ReceivablePayableID, record.Amount.Neg())
	case "ap_invoice":
		return s.apInvoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, record.ReceivablePayableID, record.Amount.Neg())
	}

	return nil
}

func (s *WriteOffService) createDiffVoucher(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, record *model.WriteOffRecord) error {
	return nil
}

func (s *WriteOffService) GetWriteOffRecords(ctx context.Context, tenantID uuid.UUID, params map[string]interface{}) ([]*model.WriteOffRecord, error) {
	return s.writeOffRepo.List(ctx, tenantID, params)
}

func (s *WriteOffService) ReverseWriteOff(ctx context.Context, tenantID, operatorID uuid.UUID, recordID int64) error {
	record, err := s.writeOffRepo.GetByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record == nil {
		return fmt.Errorf("record not found")
	}

	if record.Status != model.WriteOffStatusApproved {
		return fmt.Errorf("only approved records can be reversed")
	}

	tx, err := s.writeOffRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.reverseDocumentWriteOffStatus(ctx, tx, tenantID, record); err != nil {
		return err
	}

	now := time.Now()
	record.Status = model.WriteOffStatusReversed
	record.UpdatedAt = now

	if err := s.writeOffRepo.UpdateTx(ctx, tx, record); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *WriteOffService) reverseDocumentWriteOffStatus(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, record *model.WriteOffRecord) error {
	switch record.ReceivablePayableType {
	case "ar_invoice":
		return s.arInvoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, record.ReceivablePayableID, record.Amount)
	case "ap_invoice":
		return s.apInvoiceRepo.UpdateOutstandingAmountTx(ctx, tx, tenantID, record.ReceivablePayableID, record.Amount)
	}

	return nil
}

func (s *WriteOffService) GetUnmatchedSummary(ctx context.Context, tenantID uuid.UUID) (*WriteOffUnmatchedSummary, error) {
	arInvoices, err := s.arInvoiceRepo.ListByTenant(ctx, tenantID, nil)
	if err != nil {
		return nil, err
	}

	apInvoices, err := s.apInvoiceRepo.ListByTenant(ctx, tenantID, nil)
	if err != nil {
		return nil, err
	}

	payments, err := s.paymentRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	totalUnmatched := decimal.Zero
	overdueAmount := decimal.Zero
	counterpartyMap := make(map[uuid.UUID]decimal.Decimal)
	counterpartyNames := make(map[uuid.UUID]string)

	for _, inv := range arInvoices {
		if inv.OutstandingAmount.GreaterThan(decimal.Zero) {
			totalUnmatched = totalUnmatched.Add(inv.OutstandingAmount)
			counterpartyMap[inv.CustomerID] = counterpartyMap[inv.CustomerID].Add(inv.OutstandingAmount)

			if inv.DueDate != nil && inv.DueDate.Before(time.Now()) {
				overdueAmount = overdueAmount.Add(inv.OutstandingAmount)
			}
		}
	}

	for _, inv := range apInvoices {
		if inv.OutstandingAmount.GreaterThan(decimal.Zero) {
			totalUnmatched = totalUnmatched.Add(inv.OutstandingAmount)
			counterpartyMap[inv.SupplierID] = counterpartyMap[inv.SupplierID].Add(inv.OutstandingAmount)

			if inv.DueDate != nil && inv.DueDate.Before(time.Now()) {
				overdueAmount = overdueAmount.Add(inv.OutstandingAmount)
			}
		}
	}

	for _, p := range payments {
		if p.UnallocatedAmount.GreaterThan(decimal.Zero) {
			totalUnmatched = totalUnmatched.Add(p.UnallocatedAmount)
			counterpartyMap[p.PartyID] = counterpartyMap[p.PartyID].Add(p.UnallocatedAmount)
		}
	}

	var byCounterparty []WriteOffCounterpartySummary
	for cpID, amount := range counterpartyMap {
		byCounterparty = append(byCounterparty, WriteOffCounterpartySummary{
			CounterpartyID: cpID,
			Name:           counterpartyNames[cpID],
			Amount:         amount,
		})
	}

	return &WriteOffUnmatchedSummary{
		TotalUnmatchedAmount: totalUnmatched,
		OverdueAmount:        overdueAmount,
		ByCounterparty:       byCounterparty,
	}, nil
}

func (s *WriteOffService) GetUnmatchedItems(ctx context.Context, tenantID, counterpartyID uuid.UUID) ([]*model.WriteOffUnmatchedItem, error) {
	return s.writeOffRepo.GetUnmatchedByCounterparty(ctx, tenantID, counterpartyID)
}