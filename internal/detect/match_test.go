package detect_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/detect"
)

func TestMatchStringWithSubtype(t *testing.T) {
	m := detect.Match{Detector: "credit_card", Subtype: "visa", Start: 6, End: 25, Masked: "4111********1111"}
	want := "credit_card/visa 4111********1111 [6:25]"
	if got := m.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestMatchStringWithoutSubtype(t *testing.T) {
	m := detect.Match{Detector: "ssn", Start: 4, End: 15, Masked: "*******1120"}
	want := "ssn *******1120 [4:15]"
	if got := m.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// TestMatchFormattingIsSafe checks that the common ways of printing a
// Match all go through String, so none of them can expose more than
// the masked value.
func TestMatchFormattingIsSafe(t *testing.T) {
	m := detect.Match{Detector: "credit_card", Subtype: "visa", Masked: "4111********1111"}
	formats := []string{"%v", "%s", "%+v"}
	for _, f := range formats {
		out := fmt.Sprintf(f, m)
		if !strings.Contains(out, "4111********1111") {
			t.Errorf("format %s produced %q, expected the masked value", f, out)
		}
	}
}
