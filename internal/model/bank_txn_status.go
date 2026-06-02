package model

// BankTxnReviewStatus 银行流水审核流程状态枚举
type BankTxnReviewStatus string

const (
	BankTxnReviewStatusPending          BankTxnReviewStatus = "pending"
	BankTxnReviewStatusClassified       BankTxnReviewStatus = "classified"
	BankTxnReviewStatusApproved         BankTxnReviewStatus = "approved"
	BankTxnReviewStatusVoucherGenerated BankTxnReviewStatus = "voucher_generated"
	BankTxnReviewStatusPaymentCreated   BankTxnReviewStatus = "payment_created"
	BankTxnReviewStatusManualPending    BankTxnReviewStatus = "manual_pending"
)

func (s BankTxnReviewStatus) String() string { return string(s) }

func (s BankTxnReviewStatus) IsValid() bool {
	switch s {
	case BankTxnReviewStatusPending, BankTxnReviewStatusClassified,
		BankTxnReviewStatusApproved, BankTxnReviewStatusVoucherGenerated,
		BankTxnReviewStatusPaymentCreated, BankTxnReviewStatusManualPending:
		return true
	}
	return false
}