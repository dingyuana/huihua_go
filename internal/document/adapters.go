package document

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// salesInvoiceAdapter adapts *model.SalesInvoice to ProcessDocument / Locker.
type salesInvoiceAdapter struct {
	inv   *model.SalesInvoice
	repo  InvoiceRepoForLock
}

func (a *salesInvoiceAdapter) Kind() Kind             { return KindSalesInvoice }
func (a *salesInvoiceAdapter) GetID() uuid.UUID       { return a.inv.ID }
func (a *salesInvoiceAdapter) GetTenantID() uuid.UUID { return a.inv.TenantID }
func (a *salesInvoiceAdapter) GetTotalAmount() decimal.Decimal {
	return a.inv.TotalAmount
}
func (a *salesInvoiceAdapter) GetOutstandingAmount() decimal.Decimal {
	return a.inv.OutstandingAmount
}
func (a *salesInvoiceAdapter) GetStatus() string { return a.inv.Status }

func (a *salesInvoiceAdapter) AsLocker() Locker { return a }

func (a *salesInvoiceAdapter) LockForUpdate(ctx context.Context, tx pgx.Tx) error {
	return a.repo.LockInvoiceForUpdate(ctx, tx, a.inv.TenantID, a.inv.ID)
}

func (a *salesInvoiceAdapter) UpdateOutstanding(ctx context.Context, tx pgx.Tx, newOutstanding decimal.Decimal) error {
	return a.repo.UpdateOutstandingAmountTx(ctx, tx, a.inv.TenantID, a.inv.ID, newOutstanding.String())
}

func (a *salesInvoiceAdapter) UpdateStatus(ctx context.Context, tx pgx.Tx, newStatus string) error {
	return a.repo.UpdateStatusTx(ctx, tx, a.inv.TenantID, a.inv.ID, newStatus)
}

// arInvoiceAdapter adapts *model.ArInvoice to ProcessDocument / Locker.
type arInvoiceAdapter struct {
	inv  *model.ArInvoice
	repo ArRepoForLock
}

func (a *arInvoiceAdapter) Kind() Kind             { return KindArInvoice }
func (a *arInvoiceAdapter) GetID() uuid.UUID       { return a.inv.ID }
func (a *arInvoiceAdapter) GetTenantID() uuid.UUID { return a.inv.TenantID }
func (a *arInvoiceAdapter) GetTotalAmount() decimal.Decimal {
	return a.inv.Amount
}
func (a *arInvoiceAdapter) GetOutstandingAmount() decimal.Decimal {
	return a.inv.OutstandingAmount
}
func (a *arInvoiceAdapter) GetStatus() string { return a.inv.Status }

func (a *arInvoiceAdapter) AsLocker() Locker { return a }

func (a *arInvoiceAdapter) LockForUpdate(ctx context.Context, tx pgx.Tx) error {
	return a.repo.LockForUpdate(ctx, tx, a.inv.TenantID, a.inv.ID)
}

func (a *arInvoiceAdapter) UpdateOutstanding(ctx context.Context, tx pgx.Tx, newOutstanding decimal.Decimal) error {
	return a.repo.UpdateOutstandingAmountTx(ctx, tx, a.inv.TenantID, a.inv.ID, newOutstanding)
}

func (a *arInvoiceAdapter) UpdateStatus(ctx context.Context, tx pgx.Tx, newStatus string) error {
	return a.repo.UpdateStatusTx(ctx, tx, a.inv.TenantID, a.inv.ID, newStatus)
}

// apInvoiceAdapter adapts *model.ApInvoice to ProcessDocument / Locker.
type apInvoiceAdapter struct {
	inv  *model.ApInvoice
	repo ApRepoForLock
}

func (a *apInvoiceAdapter) Kind() Kind             { return KindApInvoice }
func (a *apInvoiceAdapter) GetID() uuid.UUID       { return a.inv.ID }
func (a *apInvoiceAdapter) GetTenantID() uuid.UUID { return a.inv.TenantID }
func (a *apInvoiceAdapter) GetTotalAmount() decimal.Decimal {
	return a.inv.Amount
}
func (a *apInvoiceAdapter) GetOutstandingAmount() decimal.Decimal {
	return a.inv.OutstandingAmount
}
func (a *apInvoiceAdapter) GetStatus() string { return a.inv.Status }

func (a *apInvoiceAdapter) AsLocker() Locker { return a }

func (a *apInvoiceAdapter) LockForUpdate(ctx context.Context, tx pgx.Tx) error {
	return a.repo.LockForUpdate(ctx, tx, a.inv.TenantID, a.inv.ID)
}

func (a *apInvoiceAdapter) UpdateOutstanding(ctx context.Context, tx pgx.Tx, newOutstanding decimal.Decimal) error {
	return a.repo.UpdateOutstandingAmountTx(ctx, tx, a.inv.TenantID, a.inv.ID, newOutstanding)
}

func (a *apInvoiceAdapter) UpdateStatus(ctx context.Context, tx pgx.Tx, newStatus string) error {
	return a.repo.UpdateStatusTx(ctx, tx, a.inv.TenantID, a.inv.ID, newStatus)
}

// Public adapter constructors. Each accepts a repo satisfying the minimal
// lock-and-update interface so callers don't have to depend on the concrete
// *repository.InvoiceRepository / *repository.ArInvoiceRepository / etc.

func AsSalesInvoice(inv *model.SalesInvoice, repo InvoiceRepoForLock) Adapter {
	return &salesInvoiceAdapter{inv: inv, repo: repo}
}
func AsArInvoice(inv *model.ArInvoice, repo ArRepoForLock) Adapter {
	return &arInvoiceAdapter{inv: inv, repo: repo}
}
func AsApInvoice(inv *model.ApInvoice, repo ApRepoForLock) Adapter {
	return &apInvoiceAdapter{inv: inv, repo: repo}
}

// Repo interfaces. Implementations live in internal/repository/.

type InvoiceRepoForLock interface {
	LockInvoiceForUpdate(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID) error
	UpdateOutstandingAmountTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, amount string) error
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, status string) error
}

type ArRepoForLock interface {
	LockForUpdate(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID) error
	UpdateOutstandingAmountTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, outstanding decimal.Decimal) error
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, status string) error
}

type ApRepoForLock interface {
	LockForUpdate(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID) error
	UpdateOutstandingAmountTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, outstanding decimal.Decimal) error
	UpdateStatusTx(ctx context.Context, tx pgx.Tx, tenantID, id uuid.UUID, status string) error
}

// Compile-time interface compliance check.
var (
	_ Adapter = (*salesInvoiceAdapter)(nil)
	_ Adapter = (*arInvoiceAdapter)(nil)
	_ Adapter = (*apInvoiceAdapter)(nil)
	_ Locker  = (*salesInvoiceAdapter)(nil)
	_ Locker  = (*arInvoiceAdapter)(nil)
	_ Locker  = (*apInvoiceAdapter)(nil)
)
