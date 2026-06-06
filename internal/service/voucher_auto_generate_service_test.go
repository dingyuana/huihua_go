package service

import "testing"

// TestInvoiceTypeSaleConstantLocksCanonicalValue guards against a regression
// of the bug where the voucher auto-generate checked for "sales" (plural)
// instead of the canonical "sale" (singular) used by InvoiceImport and the
// frontend enum. A mismatch routes sales invoices through the purchase
// branch, producing Dr 5401 / Cr 2202 with no tax line.
//
// If you intentionally rename the canonical value, update BOTH:
//   - this constant
//   - invoice_service.go ImportFromExcel (invType := "sale")
//   - frontend/src/types/enums.ts (InvoiceType.Sale = 'sale')
func TestInvoiceTypeSaleConstantLocksCanonicalValue(t *testing.T) {
	if invoiceTypeSale != "sale" {
		t.Errorf("invoiceTypeSale = %q; want \"sale\" (must match InvoiceImport and frontend enum)", invoiceTypeSale)
	}
}

// TestInvoiceTypeSaleConstantRejectsLegacyPluralForm documents that the
// legacy "sales" (plural) form is wrong. If a refactor reintroduces the
// plural form, the assertion below would also need to change — and the
// reviewer should know that means re-introducing the route-to-purchase bug.
func TestInvoiceTypeSaleConstantRejectsLegacyPluralForm(t *testing.T) {
	if invoiceTypeSale == "sales" {
		t.Error("invoiceTypeSale must NOT be \"sales\" (plural); it causes sales invoices to be routed to the purchase branch in GenerateFromInvoice")
	}
}
