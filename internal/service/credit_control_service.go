package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/repository"
)

type CreditControlService struct {
	partyRepo *repository.PartyRepository
}

func NewCreditControlService(partyRepo *repository.PartyRepository) *CreditControlService {
	return &CreditControlService{partyRepo: partyRepo}
}

type CreditStatus struct {
	PartyID         uuid.UUID       `json:"party_id"`
	PartyName       string          `json:"party_name"`
	CreditLimit     decimal.Decimal `json:"credit_limit"`
	CreditUsed      decimal.Decimal `json:"credit_used"`
	Available       decimal.Decimal `json:"available"`
	Utilization     decimal.Decimal `json:"utilization_pct"`
	OverdraftAllowed bool            `json:"overdraft_allowed"`
	OverLimit       bool            `json:"over_limit"`
}

func (s *CreditControlService) GetStatus(ctx context.Context, tenantID, partyID uuid.UUID) (*CreditStatus, error) {
	p, err := s.partyRepo.GetByID(ctx, tenantID, partyID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("party not found")
	}
	available := p.CreditLimit.Sub(p.CreditUsed)
	if available.IsNegative() {
		available = decimal.Zero
	}
	utilization := decimal.Zero
	if !p.CreditLimit.IsZero() {
		utilization = p.CreditUsed.Div(p.CreditLimit).Mul(decimal.NewFromInt(100)).Round(2)
	}
	return &CreditStatus{
		PartyID: p.ID, PartyName: p.Name,
		CreditLimit: p.CreditLimit, CreditUsed: p.CreditUsed, Available: available,
		Utilization: utilization, OverdraftAllowed: p.CreditOverdraftDays > 0,
		OverLimit: p.CreditUsed.GreaterThan(p.CreditLimit),
	}, nil
}

func (s *CreditControlService) CheckOnConfirm(ctx context.Context, tenantID, partyID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	p, err := s.partyRepo.GetByID(ctx, tenantID, partyID)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("party not found")
	}
	available := p.CreditLimit.Sub(p.CreditUsed)
	if amount.LessThanOrEqual(available) {
		return nil
	}
	if p.CreditOverdraftDays <= 0 {
		return fmt.Errorf("客户 %s 信用额度不足：可用 %.2f，需要 %.2f", p.Name, available, amount)
	}
	return nil
}

func (s *CreditControlService) OccupyOnAR(ctx context.Context, tenantID, partyID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	return s.partyRepo.UpdateCreditUsed(ctx, tenantID, partyID, amount.InexactFloat64())
}

func (s *CreditControlService) ReleaseOnReceipt(ctx context.Context, tenantID, partyID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	return s.partyRepo.UpdateCreditUsed(ctx, tenantID, partyID, -amount.InexactFloat64())
}

func (s *CreditControlService) ListOverLimit(ctx context.Context, tenantID uuid.UUID) ([]CreditStatus, error) {
	parties, err := s.partyRepo.ListOverLimit(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	result := make([]CreditStatus, len(parties))
	for i, p := range parties {
		available := p.CreditLimit.Sub(p.CreditUsed)
		if available.IsNegative() {
			available = decimal.Zero
		}
		utilization := decimal.Zero
		if !p.CreditLimit.IsZero() {
			utilization = p.CreditUsed.Div(p.CreditLimit).Mul(decimal.NewFromInt(100)).Round(2)
		}
		result[i] = CreditStatus{
			PartyID: p.ID, PartyName: p.Name,
			CreditLimit: p.CreditLimit, CreditUsed: p.CreditUsed, Available: available,
			Utilization: utilization, OverdraftAllowed: p.CreditOverdraftDays > 0,
			OverLimit: true,
		}
	}
	return result, nil
}
