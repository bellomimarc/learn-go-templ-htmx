package loans

import "testing"

func TestParseCents(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int64
		ok   bool
	}{
		{name: "decimal comma", raw: "1234,56", want: 123456, ok: true},
		{name: "whole amount", raw: "100", want: 10000, ok: true},
		{name: "one decimal", raw: "12.3", want: 1230, ok: true},
		{name: "too many decimals", raw: "12.345", ok: false},
		{name: "bare decimal point", raw: ".", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseCents(test.raw)
			if got != test.want || ok != test.ok {
				t.Fatalf("parseCents(%q) = (%d, %t), want (%d, %t)", test.raw, got, ok, test.want, test.ok)
			}
		})
	}
}

func TestCalculateMonthlyInstallmentUsesBasisPoints(t *testing.T) {
	got := calculateMonthlyInstallment(100000, 1200, 12)
	if got != 8885 {
		t.Fatalf("calculateMonthlyInstallment() = %d cents, want 8885 cents", got)
	}
}
