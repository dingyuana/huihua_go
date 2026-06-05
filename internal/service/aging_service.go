package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type AgingService struct {
	pool *pgxpool.Pool
}

func NewAgingService(pool *pgxpool.Pool) *AgingService {
	return &AgingService{pool: pool}
}

type AgingBucket struct {
	PartyID    uuid.UUID       `json:"party_id"`
	PartyName  string          `json:"party_name"`
	Current    decimal.Decimal `json:"current"`
	B0to30     decimal.Decimal `json:"b_0_30"`
	B30to60    decimal.Decimal `json:"b_30_60"`
	B60to90    decimal.Decimal `json:"b_60_90"`
	B90Plus    decimal.Decimal `json:"b_90_plus"`
	Total      decimal.Decimal `json:"total"`
	Overdue    decimal.Decimal `json:"overdue"`
	InvoiceCount int           `json:"invoice_count"`
}

func (s *AgingService) ComputeAR(ctx context.Context, tenantID uuid.UUID, asOfDate time.Time) ([]AgingBucket, error) {
	return s.compute(ctx, tenantID, asOfDate, "ar", "ar_invoices", "customer_id", "c")
}

func (s *AgingService) ComputeAP(ctx context.Context, tenantID uuid.UUID, asOfDate time.Time) ([]AgingBucket, error) {
	return s.compute(ctx, tenantID, asOfDate, "ap", "ap_invoices", "supplier_id", "s")
}

func (s *AgingService) compute(ctx context.Context, tenantID uuid.UUID, asOfDate time.Time, _, table, partyCol, partyAlias string) ([]AgingBucket, error) {
	asOf := asOfDate
	if asOf.IsZero() {
		asOf = time.Now()
	}

	query := `
		SELECT p.id, p.name,
			COALESCE(SUM(` + table + `.outstanding_amount), 0) AS total,
			COALESCE(SUM(CASE WHEN ` + table + `.due_date >= $2 THEN ` + table + `.outstanding_amount ELSE 0 END), 0) AS current,
			COALESCE(SUM(CASE WHEN ($2 - ` + table + `.due_date) BETWEEN 0 AND 30 THEN ` + table + `.outstanding_amount ELSE 0 END), 0) AS b_0_30,
			COALESCE(SUM(CASE WHEN ($2 - ` + table + `.due_date) BETWEEN 31 AND 60 THEN ` + table + `.outstanding_amount ELSE 0 END), 0) AS b_30_60,
			COALESCE(SUM(CASE WHEN ($2 - ` + table + `.due_date) BETWEEN 61 AND 90 THEN ` + table + `.outstanding_amount ELSE 0 END), 0) AS b_60_90,
			COALESCE(SUM(CASE WHEN ($2 - ` + table + `.due_date) > 90 THEN ` + table + `.outstanding_amount ELSE 0 END), 0) AS b_90_plus,
			COUNT(` + table + `.id) AS invoice_count
		FROM parties p
		LEFT JOIN ` + table + ` ` + partyAlias + ` ON ` + partyAlias + `.` + partyCol + ` = p.id
			AND ` + table + `.tenant_id = $1
			AND ` + table + `.outstanding_amount > 0
			AND ` + table + `.status IN ('confirmed','partially_paid')
		WHERE p.tenant_id = $1 AND p.is_active = TRUE
		GROUP BY p.id, p.name
		HAVING SUM(COALESCE(` + table + `.outstanding_amount, 0)) > 0
		ORDER BY total DESC`

	rows, err := s.pool.Query(ctx, query, tenantID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var buckets []AgingBucket
	for rows.Next() {
		var b AgingBucket
		if err := rows.Scan(&b.PartyID, &b.PartyName, &b.Total, &b.Current,
			&b.B0to30, &b.B30to60, &b.B60to90, &b.B90Plus, &b.InvoiceCount); err != nil {
			return nil, err
		}
		b.Overdue = b.B0to30.Add(b.B30to60).Add(b.B60to90).Add(b.B90Plus)
		buckets = append(buckets, b)
	}
	return buckets, rows.Err()
}

type AgingSummary struct {
	AsOfDate     time.Time       `json:"as_of_date"`
	TotalCurrent decimal.Decimal `json:"total_current"`
	Total0to30   decimal.Decimal `json:"total_0_30"`
	Total30to60  decimal.Decimal `json:"total_30_60"`
	Total60to90  decimal.Decimal `json:"total_60_90"`
	Total90Plus  decimal.Decimal `json:"total_90_plus"`
	GrandTotal   decimal.Decimal `json:"grand_total"`
	OverdueCount int             `json:"overdue_count"`
}

func (s *AgingService) Summarize(buckets []AgingBucket, asOf time.Time) AgingSummary {
	sum := AgingSummary{AsOfDate: asOf}
	for _, b := range buckets {
		sum.TotalCurrent = sum.TotalCurrent.Add(b.Current)
		sum.Total0to30 = sum.Total0to30.Add(b.B0to30)
		sum.Total30to60 = sum.Total30to60.Add(b.B30to60)
		sum.Total60to90 = sum.Total60to90.Add(b.B60to90)
		sum.Total90Plus = sum.Total90Plus.Add(b.B90Plus)
		if b.B90Plus.IsPositive() {
			sum.OverdueCount += b.InvoiceCount
		}
	}
	sum.GrandTotal = sum.TotalCurrent.Add(sum.Total0to30).Add(sum.Total30to60).Add(sum.Total60to90).Add(sum.Total90Plus)
	return sum
}
