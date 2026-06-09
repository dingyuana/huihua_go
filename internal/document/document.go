// Package document defines a unified ProcessDocument abstraction over the
// three invoice-like tables (sales_invoices, ar_invoices, ap_invoices) so
// that code which needs to "operate on a document regardless of type" can
// share logic instead of branching on invoice type.
//
// Each document type carries: an id, a tenant, an outstanding balance, a
// total/amount, and a status. Settlement operations (allocate, rollback,
// confirm) repeatedly touch these fields. Today each invoice service
// implements the same operations against its own table; this package
// centralizes the read/write semantics behind a single interface so future
// settlement flows (e.g. cross-document netting, auto-reconciliation)
// can target a uniform API.
//
// Backwards compatibility: existing invoice_service / ar_invoice_service /
// ap_invoice_service continue to use their own repos. New code can depend
// on this interface and the Adapter helpers to bridge model -> interface.
package document

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// Kind identifies which concrete table a ProcessDocument represents.
type Kind string

const (
	KindSalesInvoice Kind = "sales_invoice"
	KindArInvoice    Kind = "ar_invoice"
	KindApInvoice    Kind = "ap_invoice"
)

// ProcessDocument is the read-only view used by cross-cutting settlement logic.
// Implementations adapt *model.SalesInvoice / *model.ArInvoice / *model.ApInvoice.
type ProcessDocument interface {
	Kind() Kind
	GetID() uuid.UUID
	GetTenantID() uuid.UUID
	GetTotalAmount() decimal.Decimal
	GetOutstandingAmount() decimal.Decimal
	GetStatus() string
}

// Locker is the write-side companion to ProcessDocument: it knows how to
// acquire a row lock and update outstanding + status atomically within a tx.
// Implemented per document kind by Adapter + the underlying repo.
type Locker interface {
	ProcessDocument
	LockForUpdate(ctx context.Context, tx pgx.Tx) error
	UpdateOutstanding(ctx context.Context, tx pgx.Tx, newOutstanding decimal.Decimal) error
	UpdateStatus(ctx context.Context, tx pgx.Tx, newStatus string) error
}

// Adapter wraps a ProcessDocument with its concrete repo so that the locker
// methods can dispatch to the right SQL.
type Adapter interface {
	ProcessDocument
	AsLocker() Locker
}
