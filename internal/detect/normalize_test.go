package detect_test

import (
	"testing"

	"github.com/KunjJoshi/toolguard/internal/detect"
)

type digitsOnlyCase struct {
	name  string
	input string
	want  string
}

var digitsOnlyCases = []digitsOnlyCase{
	{name: "already digits", input: "4111111111111111", want: "4111111111111111"},
	{name: "spaces", input: "4111 1111 1111 1111", want: "4111111111111111"},
	{name: "dashes", input: "4111-1111-1111-1111", want: "4111111111111111"},
	{name: "mixed separators", input: "4111-1111 1111-1111", want: "4111111111111111"},
	{name: "ssn format", input: "078-05-1120", want: "078051120"},
	{name: "letters dropped", input: "a1b2c3", want: "123"},
	{name: "fullwidth digits dropped", input: "４１１１", want: ""},
	{name: "empty", input: "", want: ""},
	{name: "no digits", input: "card number", want: ""},
}

func (tc digitsOnlyCase) run(t *testing.T) {
	if got := detect.DigitsOnly(tc.input); got != tc.want {
		t.Errorf("DigitsOnly(%q) = %q, want %q", tc.input, got, tc.want)
	}
}

func TestDigitsOnly(t *testing.T) {
	for _, tc := range digitsOnlyCases {
		t.Run(tc.name, tc.run)
	}
}

func FuzzDigitsOnly(f *testing.F) {
	for _, tc := range digitsOnlyCases {
		f.Add(tc.input)
	}
	f.Fuzz(checkDigitsOnlyProperties)
}

// checkDigitsOnlyProperties asserts that the output is never longer than
// the input and contains only ASCII digits.
func checkDigitsOnlyProperties(t *testing.T, s string) {
	out := detect.DigitsOnly(s)
	if len(out) > len(s) {
		t.Errorf("DigitsOnly(%q) = %q, longer than input", s, out)
	}
	for i := 0; i < len(out); i++ {
		if out[i] < '0' || out[i] > '9' {
			t.Errorf("DigitsOnly(%q) = %q, contains non-digit byte", s, out)
			return
		}
	}
}

var normalizeDigitsCases = []digitsOnlyCase{
	{name: "ascii with separators", input: "4111-1111 1111", want: "411111111111"},
	{name: "fullwidth", input: "４１１１", want: "4111"},
	{name: "arabic-indic", input: "٤١٢٣", want: "4123"},
	{name: "devanagari", input: "०१२३४५६७८९", want: "0123456789"},
	{name: "mathematical bold", input: "𝟎𝟗", want: "09"},
	{name: "zero-width space dropped", input: "41\u200b11", want: "4111"},
	{name: "letters dropped", input: "a1b2", want: "12"},
	{name: "empty", input: "", want: ""},
}

func (tc digitsOnlyCase) runNormalize(t *testing.T) {
	if got := detect.NormalizeDigits(tc.input); got != tc.want {
		t.Errorf("NormalizeDigits(%q) = %q, want %q", tc.input, got, tc.want)
	}
}

func TestNormalizeDigits(t *testing.T) {
	for _, tc := range normalizeDigitsCases {
		t.Run(tc.name, tc.runNormalize)
	}
}
