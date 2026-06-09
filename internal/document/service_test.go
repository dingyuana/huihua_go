package document

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// fakeLocker captures the calls made against a Locker so we can assert
// behavior without a real database.
type fakeLocker struct {
	id                 uuid.UUID
	tenantID           uuid.UUID
	kind               Kind
	total, outstanding decimal.Decimal
	status             string
	locked             bool
	updatedOut         *decimal.Decimal
	updatedStatus      *string
}

func (f *fakeLocker) Kind() Kind                { return f.kind }
func (f *fakeLocker) GetID() uuid.UUID          { return f.id }
func (f *fakeLocker) GetTenantID() uuid.UUID    { return f.tenantID }
func (f *fakeLocker) GetTotalAmount() decimal.Decimal { return f.total }
func (f *fakeLocker) GetOutstandingAmount() decimal.Decimal { return f.outstanding }
func (f *fakeLocker) GetStatus() string         { return f.status }
func (f *fakeLocker) LockForUpdate(_ context.Context, _ pgx.Tx) error {
	f.locked = true
	return nil
}
func (f *fakeLocker) UpdateOutstanding(_ context.Context, _ pgx.Tx, v decimal.Decimal) error {
	cp := v
	f.updatedOut = &cp
	f.outstanding = v
	return nil
}
func (f *fakeLocker) UpdateStatus(_ context.Context, _ pgx.Tx, s string) error {
	cp := s
	f.updatedStatus = &cp
	f.status = s
	return nil
}
func (f *fakeLocker) AsLocker() Locker { return f }

func TestSettleReducesOutstandingAndUpdatesStatus(t *testing.T) {
	f := &fakeLocker{
		id: uuid.New(), tenantID: uuid.New(),
		kind: KindSalesInvoice, total: decimal.NewFromInt(1000),
		outstanding: decimal.NewFromInt(1000), status: "unpaid",
	}
	svc := NewService()
	newOut, err := svc.Settle(context.Background(), nil, f, decimal.NewFromInt(300))
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}
	if !newOut.Equal(decimal.NewFromInt(700)) {
		t.Errorf("new outstanding = %s, want 700", newOut.String())
	}
	if f.outstanding.String() != "700" {
		t.Errorf("doc outstanding = %s, want 700", f.outstanding.String())
	}
	if f.status != "partially_paid" {
		t.Errorf("doc status = %s, want partially_paid", f.status)
	}
	if !f.locked {
		t.Error("expected LockForUpdate to be called")
	}
}

func TestSettleToZeroMarksPaid(t *testing.T) {
	f := &fakeLocker{
		id: uuid.New(), tenantID: uuid.New(),
		kind: KindArInvoice, total: decimal.NewFromInt(500),
		outstanding: decimal.NewFromInt(500), status: "unpaid",
	}
	svc := NewService()
	_, err := svc.Settle(context.Background(), nil, f, decimal.NewFromInt(500))
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}
	if f.status != "paid" {
		t.Errorf("status = %s, want paid", f.status)
	}
}

func TestSettleRejectsExcessDelta(t *testing.T) {
	f := &fakeLocker{
		id: uuid.New(), tenantID: uuid.New(),
		kind: KindApInvoice, total: decimal.NewFromInt(100),
		outstanding: decimal.NewFromInt(50), status: "partially_paid",
	}
	svc := NewService()
	_, err := svc.Settle(context.Background(), nil, f, decimal.NewFromInt(80))
	if err == nil {
		t.Fatal("expected error when delta > outstanding")
	}
}

func TestReverseAddsBackOutstanding(t *testing.T) {
	f := &fakeLocker{
		id: uuid.New(), tenantID: uuid.New(),
		kind: KindSalesInvoice, total: decimal.NewFromInt(1000),
		outstanding: decimal.NewFromInt(700), status: "partially_paid",
	}
	svc := NewService()
	newOut, err := svc.Reverse(context.Background(), nil, f, decimal.NewFromInt(300))
	if err != nil {
		t.Fatalf("Reverse: %v", err)
	}
	if !newOut.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("new outstanding = %s, want 1000", newOut.String())
	}
	// After full reverse (new outstanding = 1000, total = 1000), should
	// transition back to "unpaid".
	if f.status != "unpaid" {
		t.Errorf("status = %s, want unpaid", f.status)
	}
}

func TestReverseFromPaidGoesPartiallyPaid(t *testing.T) {
	f := &fakeLocker{
		id: uuid.New(), tenantID: uuid.New(),
		kind: KindArInvoice, total: decimal.NewFromInt(1000),
		outstanding: decimal.Zero, status: "paid",
	}
	svc := NewService()
	_, err := svc.Reverse(context.Background(), nil, f, decimal.NewFromInt(200))
	if err != nil {
		t.Fatalf("Reverse: %v", err)
	}
	if f.status != "partially_paid" {
		t.Errorf("status = %s, want partially_paid", f.status)
	}
}
