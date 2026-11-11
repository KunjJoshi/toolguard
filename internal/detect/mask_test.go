package detect_test

import (
	"testing"
	"unicode/utf8"

	"github.com/KunjJoshi/toolguard/internal/detect"
)

type maskCase struct {
	name  string
	input string
	want  string
}

var maskFirst4Last4Cases = []maskCase{
	{name: "16-digit card", input: "4111111111111111", want: "4111********1111"},
	{name: "15-digit amex", input: "378282246310005", want: "3782*******0005"},
	{name: "19-digit card", input: "4000000000000000002", want: "4000***********0002"},
	{name: "12 is the minimum", input: "123456789012", want: "1234****9012"},
	{name: "11 masked completely", input: "12345678901", want: "***********"},
	{name: "empty", input: "", want: ""},
}

var maskLast4Cases = []maskCase{
	{name: "ssn", input: "078051120", want: "*****1120"},
	{name: "routing number", input: "021000021", want: "*****0021"},
	{name: "8 is the minimum", input: "12345678", want: "****5678"},
	{name: "7 masked completely", input: "1234567", want: "*******"},
	{name: "empty", input: "", want: ""},
}

func (tc maskCase) runFirst4Last4(t *testing.T) {
	if got := detect.MaskFirst4Last4(tc.input); got != tc.want {
		t.Errorf("MaskFirst4Last4(%q) = %q, want %q", tc.input, got, tc.want)
	}
}

func (tc maskCase) runLast4(t *testing.T) {
	if got := detect.MaskLast4(tc.input); got != tc.want {
		t.Errorf("MaskLast4(%q) = %q, want %q", tc.input, got, tc.want)
	}
}

func TestMaskFirst4Last4(t *testing.T) {
	for _, tc := range maskFirst4Last4Cases {
		t.Run(tc.name, tc.runFirst4Last4)
	}
}

func TestMaskLast4(t *testing.T) {
	for _, tc := range maskLast4Cases {
		t.Run(tc.name, tc.runLast4)
	}
}

func FuzzMasks(f *testing.F) {
	for _, tc := range maskFirst4Last4Cases {
		f.Add(tc.input)
	}
	f.Fuzz(checkMaskProperties)
}

// checkMaskProperties asserts that masking never panics and never
// changes the length of the value (in characters).
func checkMaskProperties(t *testing.T, s string) {
	if !utf8.ValidString(s) {
		return
	}
	want := utf8.RuneCountInString(s)
	if got := utf8.RuneCountInString(detect.MaskFirst4Last4(s)); got != want {
		t.Errorf("MaskFirst4Last4(%q) has %d characters, want %d", s, got, want)
	}
	if got := utf8.RuneCountInString(detect.MaskLast4(s)); got != want {
		t.Errorf("MaskLast4(%q) has %d characters, want %d", s, got, want)
	}
}
