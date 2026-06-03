package service

import (
	"testing"

	"github.com/shopspring/decimal"
	"huihua/finance/internal/model"
)

// ─── containsSubstring ───

func TestContainsSubstring_ExactMatch(t *testing.T) {
	if !containsSubstring("hello world", "hello") {
		t.Error("expected 'hello' to be found in 'hello world'")
	}
}

func TestContainsSubstring_CaseInsensitive(t *testing.T) {
	if !containsSubstring("Hello World", "hello") {
		t.Error("expected case-insensitive match")
	}
	if !containsSubstring("hello world", "WORLD") {
		t.Error("expected case-insensitive match")
	}
}

func TestContainsSubstring_NoMatch(t *testing.T) {
	if containsSubstring("hello world", "xyz") {
		t.Error("expected no match")
	}
}

func TestContainsSubstring_EmptyKeyword(t *testing.T) {
	if containsSubstring("hello", "") {
		t.Error("empty keyword should return false")
	}
}

func TestContainsSubstring_Chinese(t *testing.T) {
	if !containsSubstring("支付货款给上海公司", "货款") {
		t.Error("expected chinese substring match")
	}
	if !containsSubstring("支付货款给上海公司", "上海") {
		t.Error("expected chinese substring match")
	}
	if containsSubstring("支付货款给上海公司", "北京") {
		t.Error("expected no match for 北京")
	}
}

func TestContainsSubstring_EmptyText(t *testing.T) {
	if containsSubstring("", "hello") {
		t.Error("empty text should not match")
	}
}

// ─── ValidateDebitCredit ───

func TestValidateDebitCredit_Balanced(t *testing.T) {
	lines := []model.JournalEntryLine{
		{Debit: decimal.NewFromInt(1000), Credit: decimal.Zero},
		{Debit: decimal.Zero, Credit: decimal.NewFromInt(1000)},
	}
	if err := ValidateDebitCredit(lines); err != nil {
		t.Errorf("expected balanced to pass, got: %v", err)
	}
}

func TestValidateDebitCredit_MultipleLinesBalanced(t *testing.T) {
	lines := []model.JournalEntryLine{
		{Debit: decimal.NewFromInt(500), Credit: decimal.Zero},
		{Debit: decimal.NewFromInt(300), Credit: decimal.Zero},
		{Debit: decimal.Zero, Credit: decimal.NewFromInt(800)},
	}
	if err := ValidateDebitCredit(lines); err != nil {
		t.Errorf("expected balanced to pass, got: %v", err)
	}
}

func TestValidateDebitCredit_Unbalanced(t *testing.T) {
	lines := []model.JournalEntryLine{
		{Debit: decimal.NewFromInt(1000), Credit: decimal.Zero},
		{Debit: decimal.Zero, Credit: decimal.NewFromInt(900)},
	}
	if err := ValidateDebitCredit(lines); err == nil {
		t.Error("expected unbalanced to fail")
	}
}

func TestValidateDebitCredit_Zero(t *testing.T) {
	lines := []model.JournalEntryLine{}
	if err := ValidateDebitCredit(lines); err != nil {
		t.Errorf("empty lines should be balanced, got: %v", err)
	}
}

func TestValidateDebitCredit_DecimalPrecision(t *testing.T) {
	lines := []model.JournalEntryLine{
		{Debit: decimal.NewFromFloat(1000.01), Credit: decimal.Zero},
		{Debit: decimal.Zero, Credit: decimal.NewFromFloat(1000.01)},
	}
	if err := ValidateDebitCredit(lines); err != nil {
		t.Errorf("expected decimal precision balanced: %v", err)
	}
}

// ─── DocStatusToVoucherStatus ───

func TestDocStatusToVoucherStatus(t *testing.T) {
	tests := []struct {
		docstatus int16
		expected  model.VoucherStatus
	}{
		{0, model.VoucherStatusDraft},
		{1, model.VoucherStatusPosted},
		{2, model.VoucherStatusVerified},
		{3, model.VoucherStatusCancelled},
		{99, model.VoucherStatusDraft}, // fallback
	}

	for _, tt := range tests {
		got := DocStatusToVoucherStatus(tt.docstatus)
		if got != tt.expected {
			t.Errorf("DocStatusToVoucherStatus(%d) = %s, want %s", tt.docstatus, got, tt.expected)
		}
	}
}

// ─── VoucherStatusToDocStatus ───

func TestVoucherStatusToDocStatus(t *testing.T) {
	tests := []struct {
		status   model.VoucherStatus
		expected int16
	}{
		{model.VoucherStatusDraft, 0},
		{model.VoucherStatusPosted, 1},
		{model.VoucherStatusVerified, 2},
		{model.VoucherStatusCancelled, 3},
		{"unknown", 0}, // fallback
	}

	for _, tt := range tests {
		got := VoucherStatusToDocStatus(tt.status)
		if got != tt.expected {
			t.Errorf("VoucherStatusToDocStatus(%s) = %d, want %d", tt.status, got, tt.expected)
		}
	}
}
