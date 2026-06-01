package service

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestParseDate(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		wantOk bool
		wantY  int
		wantM  time.Month
		wantD  int
	}{
		{"ymd_no_sep", "20240710", true, 2024, time.July, 10},
		{"iso_dash", "2024-07-10", true, 2024, time.July, 10},
		{"iso_slash", "2024/07/10", true, 2024, time.July, 10},
		{"us_style", "07/10/2024", true, 2024, time.July, 10},
		{"datetime", "2024-07-10 12:34:56", true, 2024, time.July, 10},
		{"chinese", "2024年07月10日", true, 2024, time.July, 10},
		{"with_whitespace", "  2024-07-10  ", true, 2024, time.July, 10},
		{"invalid_garbage", "not a date", false, 0, 0, 0},
		{"empty", "", false, 0, 0, 0},
		{"partial_year", "2024-07", false, 0, 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseDate(c.input)
			if c.wantOk {
				if err != nil {
					t.Fatalf("parseDate(%q) returned error: %v", c.input, err)
				}
				if got.Year() != c.wantY || got.Month() != c.wantM || got.Day() != c.wantD {
					t.Errorf("parseDate(%q) = %s, want %d-%d-%d", c.input, got.Format("2006-01-02"), c.wantY, c.wantM, c.wantD)
				}
			} else {
				if err == nil {
					t.Errorf("parseDate(%q) should fail but returned %s", c.input, got.Format("2006-01-02"))
				}
			}
		})
	}
}

func TestFallbackClassify(t *testing.T) {
	zero := decimal.Zero
	small := decimal.NewFromFloat(9.5)
	mid := decimal.NewFromFloat(200)
	big := decimal.NewFromFloat(1000)

	cases := []struct {
		name        string
		desc        string
		counterparty string
		direction   string
		debit       decimal.Decimal
		credit      decimal.Decimal
		want        string
	}{
		{"fee_in_desc_outgoing", "对公跨行转账汇款手续费", "银行", "out", zero, small, "bank_fee"},
		{"fee_in_desc_small_amount", "工本费", "银行", "out", zero, small, "bank_fee"},
		{"interest_desc", "存款利息", "银行", "in", mid, zero, "interest_income"},
		{"interest_keyword", "结息", "银行", "in", small, zero, "interest_income"},
		{"tax_keyword_desc", "实时缴税 123", "税务局", "out", zero, mid, "business_payment"},
		{"tax_keyword_counterparty", "XX", "国家税务总局", "out", zero, big, "business_payment"},
		{"social_security", "社保缴费 7月", "社保局", "out", zero, mid, "business_payment"},
		{"transfer_keyword", "内部转账", "本公司", "in", big, zero, "internal_transfer"},
		{"small_outgoing_no_keyword", "普通扣款", "商户", "out", zero, small, "bank_fee"},
		{"inbound_default", "来账", "客户A", "in", big, zero, "business_receipt"},
		{"outbound_default_large", "出账", "供应商B", "out", zero, big, "business_payment"},
		{"empty", "", "", "in", mid, zero, "business_receipt"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := fallbackClassify(c.desc, c.counterparty, c.direction, c.debit, c.credit)
			if got != c.want {
				t.Errorf("fallbackClassify(%q, %q, %q) = %q, want %q", c.desc, c.counterparty, c.direction, got, c.want)
			}
		})
	}
}
