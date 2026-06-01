package repository

import "testing"

func TestPaymentTypePrefix(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"receive", "REC"},
		{"pay", "PAY"},
		{"expense", "EXP"},
		{"interest", "INT"},
		{"transfer", "TRF"},
		{"", "DOC"},
		{"unknown", "DOC"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := paymentTypePrefix(c.in)
			if got != c.want {
				t.Errorf("paymentTypePrefix(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestIsValidPaymentType(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"receive", true},
		{"pay", true},
		{"expense", true},
		{"interest", true},
		{"transfer", true},
		{"", false},
		{"invalid", false},
		{"RECEIVE", false},
		{"refund", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := IsValidPaymentType(c.in)
			if got != c.want {
				t.Errorf("IsValidPaymentType(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
