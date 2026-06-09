package model

// PreCheckStatus represents the overall precheck result status.
type PreCheckStatus string

const (
	PreCheckStatusPassed  PreCheckStatus = "passed"
	PreCheckStatusWarning PreCheckStatus = "warning"
	PreCheckStatusBlocked PreCheckStatus = "blocked"
)

// PreCheckItem represents a single precheck validation result.
type PreCheckItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"` // "passed", "warning", "blocked"
	Message  string `json:"message"`
	Severity string `json:"severity"` // "info", "warning", "error"
}

// PreCheckResult represents the complete precheck result for a payment-invoice pair.
type PreCheckResult struct {
	InvoiceID       string        `json:"invoice_id"`
	PaymentID       string        `json:"payment_id"`
	Passed          bool          `json:"passed"`
	OverallMessage  string        `json:"overall_message"`
	Checks          []PreCheckItem `json:"checks"`
	CanForcePass    bool          `json:"can_force_pass"`
}
