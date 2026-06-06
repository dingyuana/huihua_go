package service

import "testing"

// TestExtractBlueInvoiceNo covers the 6 regex patterns in extractBlueInvoiceNo()
// which is invoked during sales-invoice import to lift the linked blue-invoice
// number out of the 备注 (remark) field for red-letter (红字) credit notes.
//
// The function lives on internal/service/invoice_service.go:761 and is called
// from the import path at line 731 when SourceRedInvoiceNo is empty.
//
// Patterns (in priority order):
//   1. 被红冲蓝字数电票号码[：:]\s*(\d{8,20})
//   2. 被红冲蓝字数电发票号码[：:]\s*(\d{8,20})
//   3. 对应蓝字发票号[：:]\s*(\d{8,20})
//   4. 对应正数发票号码[：:]\s*(\d{8,20})
//   5. 红冲发票[：:号]?\s*(\d{8,20})
//   6. 蓝字发票[：:号]?\s*(\d{8,20})

func TestExtractBlueInvoiceNo_Pattern1_BeiHongChongLanZiShuDianPiaoHaoMa(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "half-width colon with spaces",
			remark: "对应红字发票：被红冲蓝字数电票号码: 25113300000012345678 备注结束",
			want:   "25113300000012345678",
		},
		{
			name:   "full-width colon with spaces",
			remark: "被红冲蓝字数电票号码：25113300000098765432",
			want:   "25113300000098765432",
		},
		{
			name:   "no spaces after colon",
			remark: "被红冲蓝字数电票号码:25113300000011111111",
			want:   "25113300000011111111",
		},
		{
			name:   "minimal 8-digit number",
			remark: "被红冲蓝字数电票号码: 12345678",
			want:   "12345678",
		},
		{
			name:   "maximal 20-digit number",
			remark: "被红冲蓝字数电票号码: 12345678901234567890",
			want:   "12345678901234567890",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

func TestExtractBlueInvoiceNo_Pattern2_BeiHongChongLanZiShuDianFaPiaoHaoMa(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "full-width colon",
			remark: "被红冲蓝字数电发票号码：25113300000022222222",
			want:   "25113300000022222222",
		},
		{
			name:   "half-width colon",
			remark: "被红冲蓝字数电发票号码:25113300000033333333",
			want:   "25113300000033333333",
		},
		{
			name:   "with spaces",
			remark: "被红冲蓝字数电发票号码: 25113300000044444444 末尾",
			want:   "25113300000044444444",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

func TestExtractBlueInvoiceNo_Pattern3_DuiYingLanZiFaPiaoHao(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "full-width colon",
			remark: "对应蓝字发票号：25113300000055555555",
			want:   "25113300000055555555",
		},
		{
			name:   "half-width colon",
			remark: "对应蓝字发票号:25113300000066666666",
			want:   "25113300000066666666",
		},
		{
			name:   "with spaces",
			remark: "对应蓝字发票号: 25113300000077777777 结束",
			want:   "25113300000077777777",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

func TestExtractBlueInvoiceNo_Pattern4_DuiYingZhengShuFaPiaoHaoMa(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "full-width colon",
			remark: "对应正数发票号码：25113300000088888888",
			want:   "25113300000088888888",
		},
		{
			name:   "half-width colon",
			remark: "对应正数发票号码: 25113300000099999999",
			want:   "25113300000099999999",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

func TestExtractBlueInvoiceNo_Pattern5_HongChongFaPiao(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "full-width colon",
			remark: "红冲发票：25113300000100000001",
			want:   "25113300000100000001",
		},
		{
			name:   "full-width hao suffix",
			remark: "红冲发票号: 25113300000100000002",
			want:   "25113300000100000002",
		},
		{
			name:   "no separator (optional)",
			remark: "红冲发票25113300000100000003",
			want:   "25113300000100000003",
		},
		{
			name:   "half-width colon",
			remark: "红冲发票:25113300000100000004",
			want:   "25113300000100000004",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

func TestExtractBlueInvoiceNo_Pattern6_LanZiFaPiao(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "full-width colon",
			remark: "蓝字发票：25113300000100000005",
			want:   "25113300000100000005",
		},
		{
			name:   "full-width hao suffix",
			remark: "蓝字发票号: 25113300000100000006",
			want:   "25113300000100000006",
		},
		{
			name:   "no separator (optional)",
			remark: "蓝字发票25113300000100000007",
			want:   "25113300000100000007",
		},
		{
			name:   "half-width colon",
			remark: "蓝字发票:25113300000100000008",
			want:   "25113300000100000008",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != tc.want {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want %q", tc.remark, got, tc.want)
			}
		})
	}
}

// TestExtractBlueInvoiceNo_PatternPriority verifies that when multiple patterns
// could match in the same remark, the higher-priority pattern (lower index in
// the patterns slice) wins — because the implementation iterates in order and
// returns on the first match.
func TestExtractBlueInvoiceNo_PatternPriority(t *testing.T) {
	// Remark contains both pattern 1 (最具体) and pattern 6 (最通用).
	// Pattern 1 must win.
	remark := "被红冲蓝字数电票号码: 25113300000111111111 蓝字发票: 25113300000222222222"
	got := extractBlueInvoiceNo(remark)
	want := "25113300000111111111"
	if got != want {
		t.Errorf("priority test: got %q; want %q (pattern 1 should win over pattern 6)", got, want)
	}
}

func TestExtractBlueInvoiceNo_NoMatch(t *testing.T) {
	cases := []struct {
		name   string
		remark string
		want   string
	}{
		{
			name:   "empty remark",
			remark: "",
		},
		{
			name:   "remark without any keyword",
			remark: "普通备注：客户已付款",
		},
		{
			name:   "number too short (7 digits)",
			remark: "被红冲蓝字数电票号码: 1234567",
		},
		{
			name:   "non-numeric value",
			remark: "被红冲蓝字数电票号码: abc",
		},
		{
			name:   "keyword without number",
			remark: "红冲发票已开具",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBlueInvoiceNo(tc.remark)
			if got != "" {
				t.Errorf("extractBlueInvoiceNo(%q) = %q; want empty string", tc.remark, got)
			}
		})
	}
}

// TestExtractBlueInvoiceNo_RealisticRemark simulates a realistic full
// 备注 string as it would appear in a 数电发票 (e-invoice) export.
func TestExtractBlueInvoiceNo_RealisticRemark(t *testing.T) {
	remark := "红字发票 对应蓝字发票号: 25113300000987654321 开具日期2025-12-01 税率13%"
	got := extractBlueInvoiceNo(remark)
	want := "25113300000987654321"
	if got != want {
		t.Errorf("realistic remark: got %q; want %q", got, want)
	}
}

// TestExtractBlueInvoiceNo_FirstMatchWinsWhenMultipleKeywordsPresent checks
// that within a remark, the first-occurring match is the one returned, even
// when a later, less-specific pattern is also present.
func TestExtractBlueInvoiceNo_FirstMatchWinsWhenMultipleKeywordsPresent(t *testing.T) {
	// The most-specific (pattern 1) appears AFTER the generic (pattern 6)
	// in the remark text. Pattern 1 must still be returned because the
	// implementation iterates the pattern list, not the remark position.
	remark := "蓝字发票 99999999 被红冲蓝字数电票号码: 25113300000333333333"
	got := extractBlueInvoiceNo(remark)
	want := "25113300000333333333"
	if got != want {
		t.Errorf("first-match-wins by pattern order: got %q; want %q", got, want)
	}
}

// TestExtractBlueInvoiceNo_LongNumberTruncatesAt20Digits documents the
// current behavior of the regex \d{8,20} when the number exceeds 20 digits.
// \d{8,20} is greedy and matches up to 20 digits from the start, so a 21+
// digit run is silently truncated to its first 20. This is an implementation
// quirk worth surfacing — real invoice numbers are 8-20 digits, but the
// function does not validate max length.
func TestExtractBlueInvoiceNo_LongNumberTruncatesAt20Digits(t *testing.T) {
	remark := "被红冲蓝字数电票号码: 123456789012345678901"
	got := extractBlueInvoiceNo(remark)
	want := "12345678901234567890"
	if got != want {
		t.Errorf("long number: got %q; want %q (current greedy behavior)", got, want)
	}
}
