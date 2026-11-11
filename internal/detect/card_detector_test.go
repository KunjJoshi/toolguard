package detect_test

import (
	"strings"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/detect"
	"github.com/KunjJoshi/toolguard/internal/detect/validators"
)

type cardDetectorCase struct {
	name  string
	input string
	want  []detect.Match
}

func visaMatch(start, end int) detect.Match {
	return detect.Match{Detector: "credit_card", Subtype: "visa", Start: start, End: end, Masked: "4111********1111"}
}

func mastercardMatch(start, end int) detect.Match {
	return detect.Match{Detector: "credit_card", Subtype: "mastercard", Start: start, End: end, Masked: "5555********4444"}
}

var cardDetectorCases = []cardDetectorCase{
	// Ordinary formats.
	{name: "plain card in a sentence", input: "Charge card 4111111111111111 today", want: []detect.Match{visaMatch(12, 28)}},
	{name: "card is the whole text", input: "4111111111111111", want: []detect.Match{visaMatch(0, 16)}},
	{
		name:  "two cards with spaces and dashes",
		input: "Cards: 4111 1111 1111 1111 and 5555-5555-5555-4444.",
		want:  []detect.Match{visaMatch(7, 26), mastercardMatch(31, 50)},
	},
	{
		name:  "amex in its 4-6-5 grouping",
		input: "Amex 3782 822463 10005 on file",
		want:  []detect.Match{{Detector: "credit_card", Subtype: "amex", Start: 5, End: 22, Masked: "3782*******0005"}},
	},
	{name: "card followed by an extra number group", input: "Pay 4111 1111 1111 1111 22 now", want: []detect.Match{visaMatch(4, 23)}},
	{name: "card preceded by an extra number group", input: "ref 12 4111 1111 1111 1111", want: []detect.Match{visaMatch(7, 26)}},
	{name: "card inside JSON", input: `{"note":"4111111111111111"}`, want: []detect.Match{visaMatch(9, 25)}},

	// Evasion attempts that must still be caught.
	{name: "glued to letters before", input: "abc4111111111111111", want: []detect.Match{visaMatch(3, 19)}},
	{name: "glued to letters on both sides", input: "x4111111111111111y", want: []detect.Match{visaMatch(1, 17)}},
	{name: "after an underscore", input: "card_4111111111111111", want: []detect.Match{visaMatch(5, 21)}},
	{
		name:  "two cards separated by a comma",
		input: "4111111111111111,5555555555554444",
		want:  []detect.Match{visaMatch(0, 16), mastercardMatch(17, 33)},
	},
	{name: "dot separators", input: "4111.1111.1111.1111", want: []detect.Match{visaMatch(0, 19)}},
	{name: "double spaces", input: "4111  1111  1111  1111", want: []detect.Match{visaMatch(0, 22)}},
	{name: "line breaks", input: "4111\n1111\n1111\n1111", want: []detect.Match{visaMatch(0, 19)}},
	{name: "zero-width spaces", input: "4111\u200b1111\u200b1111\u200b1111", want: []detect.Match{visaMatch(0, 25)}},
	{name: "fullwidth digits", input: "４１１１１１１１１１１１１１１１", want: []detect.Match{visaMatch(0, 48)}},
	{name: "arabic-indic digits", input: "٤١١١١١١١١١١١١١١١", want: []detect.Match{visaMatch(0, 32)}},
	{name: "invalid utf-8 around card", input: "\xff4111111111111111\xff", want: []detect.Match{visaMatch(1, 17)}},

	// Not cards.
	{name: "luhn invalid", input: "Card 4111111111111112", want: nil},
	{name: "luhn valid but unknown prefix", input: "Ref 9999999999999995", want: nil},
	{name: "20-digit run", input: "id 41111111111111111111", want: nil},
	{name: "13-digit order number", input: "Order 1234567890123", want: nil},
	{name: "phone number", input: "Call +1 617-555-0123", want: nil},
	{name: "date and time", input: "2026-11-10 14:05:42", want: nil},
	{name: "ip address and version", input: "Server 10.0.0.1 v1.2.3", want: nil},
	{name: "empty", input: "", want: nil},

	// Known limits of the default mode, kept as tests so changes are deliberate.
	{name: "limit: four spaces split the number", input: "4111    1111 1111 1111", want: nil},
	{name: "limit: letters used as separators", input: "4111a1111b1111c1111", want: nil},
	{name: "limit: extra digit attached", input: "41111111111111111", want: nil},
}

var strictCardCases = []cardDetectorCase{
	{name: "extra digit attached after", input: "41111111111111111", want: []detect.Match{visaMatch(0, 16)}},
	{name: "extra digit attached before", input: "94111111111111111", want: []detect.Match{visaMatch(1, 17)}},
	{name: "card at start of 20-digit run", input: "id 41111111111111111111", want: []detect.Match{visaMatch(3, 19)}},
	{name: "normal card unchanged", input: "Charge card 4111111111111111 today", want: []detect.Match{visaMatch(12, 28)}},
	{
		// Documents the cost of strict mode: a window inside a long
		// sequential number happens to pass every check.
		name:  "false positive inside long number",
		input: "Hash 123456789012345678901234567890",
		want:  []detect.Match{{Detector: "credit_card", Subtype: "visa", Start: 8, End: 27, Masked: "4567***********9012"}},
	},
}

func (tc cardDetectorCase) run(t *testing.T) {
	tc.check(t, detect.FindCards(tc.input))
}

func (tc cardDetectorCase) runStrict(t *testing.T) {
	tc.check(t, detect.FindCardsWithOptions(tc.input, detect.CardOptions{ScanInsideLongRuns: true}))
}

func (tc cardDetectorCase) check(t *testing.T, got []detect.Match) {
	if len(got) != len(tc.want) {
		t.Fatalf("input %q: found %d matches %v, want %d %v", tc.input, len(got), got, len(tc.want), tc.want)
	}
	for i := range got {
		if got[i] != tc.want[i] {
			t.Errorf("input %q: match %d = %+v, want %+v", tc.input, i, got[i], tc.want[i])
		}
	}
}

func TestFindCards(t *testing.T) {
	for _, tc := range cardDetectorCases {
		t.Run(tc.name, tc.run)
	}
}

func TestFindCardsStrict(t *testing.T) {
	for _, tc := range strictCardCases {
		t.Run(tc.name, tc.runStrict)
	}
}

func FuzzFindCards(f *testing.F) {
	for _, tc := range cardDetectorCases {
		f.Add(tc.input)
	}
	f.Fuzz(checkFindCardsProperties)
}

func FuzzFindCardsStrict(f *testing.F) {
	for _, tc := range strictCardCases {
		f.Add(tc.input)
	}
	f.Fuzz(checkStrictFindCardsProperties)
}

func checkFindCardsProperties(t *testing.T, s string) {
	checkMatchProperties(t, s, detect.FindCards(s))
}

func checkStrictFindCardsProperties(t *testing.T, s string) {
	checkMatchProperties(t, s, detect.FindCardsWithOptions(s, detect.CardOptions{ScanInsideLongRuns: true}))
}

// checkMatchProperties asserts, for any input: no panic; matches are in
// bounds, in order and non-overlapping; and each masked value is a well
// formed card mask. For plain ASCII input it also re-validates the digits.
func checkMatchProperties(t *testing.T, s string, matches []detect.Match) {
	prevEnd := 0
	for _, m := range matches {
		if m.Start < prevEnd || m.End <= m.Start || m.End > len(s) {
			t.Fatalf("input %q: bad span %+v (previous end %d)", s, m, prevEnd)
		}
		prevEnd = m.End

		if !isWellFormedCardMask(m.Masked) {
			t.Errorf("input %q: malformed mask %q", s, m.Masked)
		}
		if isPlainASCII(s) {
			checkASCIIMatch(t, s, m)
		}
	}
}

func checkASCIIMatch(t *testing.T, s string, m detect.Match) {
	digits := detect.DigitsOnly(s[m.Start:m.End])
	if _, ok := validators.IsPlausibleCard(digits); !ok {
		t.Errorf("input %q: implausible card at %+v", s, m)
	}
	if !validators.Luhn(digits) {
		t.Errorf("input %q: Luhn-invalid card at %+v", s, m)
	}
}

// isWellFormedCardMask checks the shape "dddd****dddd": 13–19 characters,
// first and last 4 ASCII digits, '*' in between.
func isWellFormedCardMask(m string) bool {
	if len(m) < 13 || len(m) > 19 {
		return false
	}
	for i := 0; i < len(m); i++ {
		visible := i < 4 || i >= len(m)-4
		if visible && (m[i] < '0' || m[i] > '9') {
			return false
		}
		if !visible && m[i] != '*' {
			return false
		}
	}
	return true
}

func isPlainASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

func BenchmarkFindCards(b *testing.B) {
	text := benchmarkText()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detect.FindCards(text)
	}
}

func BenchmarkFindCardsStrict(b *testing.B) {
	text := benchmarkText()
	opts := detect.CardOptions{ScanInsideLongRuns: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detect.FindCardsWithOptions(text, opts)
	}
}

func benchmarkText() string {
	return strings.Repeat("Some ordinary tool output with numbers 12345 and words. ", 20) +
		"Card 4111 1111 1111 1111 here."
}
