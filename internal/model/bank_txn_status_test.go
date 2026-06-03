package model

import "testing"

// 验证 BankTxnReviewStatus 枚举值
func TestBankTxnReviewStatus_Values(t *testing.T) {
	statuses := []BankTxnReviewStatus{
		BankTxnReviewStatusPending,
		BankTxnReviewStatusClassified,
		BankTxnReviewStatusApproved,
		BankTxnReviewStatusVoucherGenerated,
		BankTxnReviewStatusPaymentCreated,
		BankTxnReviewStatusManualPending,
	}
	for _, s := range statuses {
		if s == "" {
			t.Errorf("BankTxnReviewStatus should not be empty")
		}
	}
}

// 验证 String() 方法
func TestBankTxnReviewStatus_String(t *testing.T) {
	tests := []struct {
		status   BankTxnReviewStatus
		expected string
	}{
		{BankTxnReviewStatusPending, "pending"},
		{BankTxnReviewStatusClassified, "classified"},
		{BankTxnReviewStatusApproved, "approved"},
		{BankTxnReviewStatusVoucherGenerated, "voucher_generated"},
		{BankTxnReviewStatusPaymentCreated, "payment_created"},
		{BankTxnReviewStatusManualPending, "manual_pending"},
	}
	for _, tt := range tests {
		if got := tt.status.String(); got != tt.expected {
			t.Errorf("String() = %q, want %q", got, tt.expected)
		}
	}
}

// 验证 IsValid() 方法
func TestBankTxnReviewStatus_IsValid(t *testing.T) {
	valid := []BankTxnReviewStatus{BankTxnReviewStatusPending, BankTxnReviewStatusClassified}
	invalid := []BankTxnReviewStatus{"", "unknown", "PENDING"}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("IsValid(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("IsValid(%q) = true, want false", s)
		}
	}
}