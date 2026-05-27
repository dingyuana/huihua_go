package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExchangeRate represents a currency exchange rate.
type ExchangeRate struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	CurrencyCode   string          `json:"currency_code" db:"currency_code"`     // 源货币（如 USD）
	TargetCurrency string          `json:"target_currency" db:"target_currency"` // 目标货币（如 CNY）
	Rate           decimal.Decimal `json:"rate" db:"rate"`                       // 汇率（1 源货币 = Rate 目标货币）
	EffectiveDate  time.Time      `json:"effective_date" db:"effective_date"`   // 生效日期
	RateType       string         `json:"rate_type" db:"rate_type"`             // spot/月均/年均
	Source         string          `json:"source" db:"source"`                   // 数据来源（manual/bank API）
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
}

// ExchangeRateRequest is the request to add an exchange rate.
type ExchangeRateRequest struct {
	CurrencyCode   string `json:"currency_code"`
	TargetCurrency string `json:"target_currency"`
	Rate           string `json:"rate"`
	EffectiveDate string `json:"effective_date"`
	RateType       string `json:"rate_type"`
	Source         string `json:"source"`
}

// ConvertRequest is a currency conversion request.
type ConvertRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Date   string `json:"date"`
}

// ConvertResult is the result of a currency conversion.
type ConvertResult struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	Rate     string `json:"rate"`
	Result   string `json:"result"`
	Date     string `json:"date"`
}
