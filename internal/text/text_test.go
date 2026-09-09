package text

import "testing"

func TestCountNonSpaceChars(t *testing.T) {
	if got := CountNonSpaceChars(" A\tB\u00a0C "); got != 3 {
		t.Fatalf("CountNonSpaceChars() = %d, want 3", got)
	}
}
