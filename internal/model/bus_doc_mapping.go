package model

import "time"

// BusDocMapping represents the bus_doc_mapping table — configurable
// document-type-to-account mappings used by voucher auto-generation.
// Each (doc_type, condition_key) pair maps to a debit/credit account pair.
// condition_key is "default" unless a sub-scenario (e.g. expense_type, tax)
// is detected from the document content.
type BusDocMapping struct {
	ID                string    `json:"id" db:"id"`
	TenantID          string    `json:"tenant_id" db:"tenant_id"`
	DocType           string    `json:"doc_type" db:"doc_type"`
	ConditionKey      string    `json:"condition_key" db:"condition_key"`
	ConditionLabel    *string   `json:"condition_label,omitempty" db:"condition_label"`
	DebitAccountID    *string   `json:"debit_account_id,omitempty" db:"debit_account_id"`
	DebitSubjectCode  string    `json:"debit_subject_code" db:"debit_subject_code"`
	DebitSubjectName  *string   `json:"debit_subject_name,omitempty" db:"debit_subject_name"`
	CreditAccountID   *string   `json:"credit_account_id,omitempty" db:"credit_account_id"`
	CreditSubjectCode string    `json:"credit_subject_code" db:"credit_subject_code"`
	CreditSubjectName *string   `json:"credit_subject_name,omitempty" db:"credit_subject_name"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	SortOrder         int       `json:"sort_order" db:"sort_order"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
