package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
	"huihua/finance/internal/repository"
)

var ErrRateNotFound = errors.New("exchange rate not found")
var ErrInvalidRate = errors.New("rate must be positive")

// ExchangeRateService handles currency exchange rate operations.
type ExchangeRateService struct {
	repo *repository.ExchangeRateRepository
}

// NewExchangeRateService creates a new ExchangeRateService.
func NewExchangeRateService(repo *repository.ExchangeRateRepository) *ExchangeRateService {
	return &ExchangeRateService{repo: repo}
}

// AddExchangeRate adds a new exchange rate.
func (s *ExchangeRateService) AddExchangeRate(ctx context.Context, tenantID uuid.UUID, req model.ExchangeRateRequest) (*model.ExchangeRate, error) {
	rate, err := decimal.NewFromString(req.Rate)
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return nil, ErrInvalidRate
	}
	date, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, errors.New("effective_date must be YYYY-MM-DD")
	}
	if req.RateType == "" {
		req.RateType = "spot"
	}
	if req.Source == "" {
		req.Source = "manual"
	}
	er := &model.ExchangeRate{
		ID:             uuid.New(),
		CurrencyCode:   req.CurrencyCode,
		TargetCurrency: req.TargetCurrency,
		Rate:           rate,
		EffectiveDate:  date,
		RateType:       req.RateType,
		Source:         req.Source,
	}
	if err := s.repo.Create(ctx, tenantID, er); err != nil {
		return nil, err
	}
	return er, nil
}

// GetExchangeRate retrieves a rate, falling back to latest if exact date not found.
func (s *ExchangeRateService) GetExchangeRate(ctx context.Context, tenantID uuid.UUID, from, to string, date time.Time) (*model.ExchangeRate, error) {
	// Direct rate
	rate, err := s.repo.GetRate(ctx, tenantID, from, to, date)
	if err == nil {
		return rate, nil
	}
	// Try latest before date
	rate, err = s.repo.GetLatestRate(ctx, tenantID, from, to, date)
	if err == nil {
		return rate, nil
	}
	// Try inverse
	invRate, err := s.repo.GetLatestRate(ctx, tenantID, to, from, date)
	if err == nil {
		rate = &model.ExchangeRate{
			ID:             invRate.ID,
			CurrencyCode:   from,
			TargetCurrency: to,
			Rate:           decimal.NewFromInt(1).Div(invRate.Rate),
			EffectiveDate:  invRate.EffectiveDate,
			RateType:       invRate.RateType,
			Source:         invRate.Source,
		}
		return rate, nil
	}
	return nil, ErrRateNotFound
}

// ConvertAmount converts an amount from one currency to another.
func (s *ExchangeRateService) ConvertAmount(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, from, to string, date time.Time) (*model.ConvertResult, error) {
	if from == to {
		return &model.ConvertResult{
			From: from, To: to, Amount: amount.String(),
			Rate: "1", Result: amount.String(), Date: date.Format("2006-01-02"),
		}, nil
	}
	exRate, err := s.GetExchangeRate(ctx, tenantID, from, to, date)
	if err != nil {
		return nil, err
	}
	result := amount.Mul(exRate.Rate)
	return &model.ConvertResult{
		From:   from,
		To:     to,
		Amount: amount.String(),
		Rate:   exRate.Rate.String(),
		Result: result.String(),
		Date:   date.Format("2006-01-02"),
	}, nil
}

// ListAll lists all rates for the tenant.
func (s *ExchangeRateService) ListAll(ctx context.Context, tenantID uuid.UUID) ([]model.ExchangeRate, error) {
	return s.repo.ListAll(ctx, tenantID)
}

// Delete deletes an exchange rate.
func (s *ExchangeRateService) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	return s.repo.DeleteRate(ctx, tenantID, id)
}
