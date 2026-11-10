package validators_test

import (
	"testing"

	"github.com/KunjJoshi/toolguard/internal/detect/validators"
)

type luhnCase struct {
	name  string
	input string
	want  bool
}

var luhnCases = []luhnCase{
	// Valid: published test card numbers and the classic textbook example.
	{name: "visa test number", input: "4111111111111111", want: true},
	{name: "visa test number 2", input: "4012888888881881", want: true},
	{name: "mastercard test number", input: "5555555555554444", want: true},
	{name: "amex test number", input: "378282246310005", want: true},
	{name: "discover test number", input: "6011111111111117", want: true},
	{name: "textbook example", input: "79927398713", want: true},

	// Invalid: one digit off from a valid number.
	{name: "visa last digit changed", input: "4111111111111112", want: false},
	{name: "textbook last digit changed", input: "79927398710", want: false},
	{name: "mastercard middle digit changed", input: "5555555555564444", want: false},

	// Rejected input shapes.
	{name: "empty", input: "", want: false},
	{name: "single digit", input: "0", want: false},
	{name: "contains spaces", input: "4111 1111 1111 1111", want: false},
	{name: "contains dashes", input: "4111-1111-1111-1111", want: false},
	{name: "contains letter", input: "4111a11111111111", want: false},
	{name: "fullwidth unicode digits", input: "４１１１１１１１１１１１１１１１", want: false},

	// Documents a known property: the checksum alone accepts all zeros.
	// The issuer-prefix check in card.go is what rejects this.
	{name: "all zeros passes checksum", input: "0000000000000000", want: true},
}

func (tc luhnCase) run(t *testing.T) {
	got := validators.Luhn(tc.input)
	if got != tc.want {
		t.Errorf("Luhn(%q) = %v, want %v", tc.input, got, tc.want)
	}
}

func TestLuhn(t *testing.T) {
	for _, tc := range luhnCases {
		t.Run(tc.name, tc.run)
	}
}

func FuzzLuhn(f *testing.F) {
	for _, tc := range luhnCases {
		f.Add(tc.input)
	}
	f.Fuzz(checkLuhnProperties)
}

// checkLuhnProperties asserts invariants that must hold for any input:
// Luhn never panics, and it only ever accepts ASCII-digit strings of at
// least two characters.
func checkLuhnProperties(t *testing.T, s string) {
	if !validators.Luhn(s) {
		return
	}
	if len(s) < 2 {
		t.Errorf("Luhn accepted %q, shorter than 2 characters", s)
	}
	if !isASCIIDigits(s) {
		t.Errorf("Luhn accepted %q, which contains non-digit bytes", s)
	}
}

func isASCIIDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func BenchmarkLuhn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validators.Luhn("4111111111111111")
	}
}
