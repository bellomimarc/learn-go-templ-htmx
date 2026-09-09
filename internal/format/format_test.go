package format

import "testing"

func TestFormatCents(t *testing.T) {
	if got := FormatCents(123456); got != "1234.56" {
		t.Fatalf("FormatCents(123456) = %q, want %q", got, "1234.56")
	}
}
