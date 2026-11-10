package validators_test

import (
	"strings"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/detect/validators"
)

// cardDigits builds a test number: the given prefix padded with zeros to
// the given total length. These are for prefix and length checks only;
// most of them do not pass Luhn, and that is fine here.
func cardDigits(prefix string, length int) string {
	return prefix + strings.Repeat("0", length-len(prefix))
}

type cardCase struct {
	name          string
	input         string
	wantCardBrand validators.CardBrand
	wantOK        bool
}

var cardCases = []cardCase{
	// Published test numbers.
	{name: "visa 16", input: "4111111111111111", wantCardBrand: validators.CardBrandVisa, wantOK: true},
	{name: "visa 13", input: "4222222222222", wantCardBrand: validators.CardBrandVisa, wantOK: true},
	{name: "mastercard 5-series", input: "5555555555554444", wantCardBrand: validators.CardBrandMasterCard, wantOK: true},
	{name: "mastercard 2-series", input: "2221000000000009", wantCardBrand: validators.CardBrandMasterCard, wantOK: true},
	{name: "amex", input: "378282246310005", wantCardBrand: validators.CardBrandAmex, wantOK: true},
	{name: "discover", input: "6011111111111117", wantCardBrand: validators.CardBrandDiscover, wantOK: true},
	{name: "jcb", input: "3530111333300000", wantCardBrand: validators.CardBrandJCB, wantOK: true},
	{name: "diners 14", input: "30569309025904", wantCardBrand: validators.CardBrandDiners, wantOK: true},
	{name: "unionpay", input: "6200000000000005", wantCardBrand: validators.CardBrandUnionPay, wantOK: true},

	// Shortest and longest lengths per CardBrand.
	{name: "visa 19", input: cardDigits("4", 19), wantCardBrand: validators.CardBrandVisa, wantOK: true},
	{name: "discover 19", input: cardDigits("6011", 19), wantCardBrand: validators.CardBrandDiscover, wantOK: true},
	{name: "jcb 19", input: cardDigits("3528", 19), wantCardBrand: validators.CardBrandJCB, wantOK: true},
	{name: "diners 19", input: cardDigits("36", 19), wantCardBrand: validators.CardBrandDiners, wantOK: true},
	{name: "unionpay 19", input: cardDigits("62", 19), wantCardBrand: validators.CardBrandUnionPay, wantOK: true},

	// Mastercard boundaries.
	{name: "mastercard 2220 below range", input: cardDigits("2220", 16), wantOK: false},
	{name: "mastercard 2720 top of range", input: cardDigits("2720", 16), wantCardBrand: validators.CardBrandMasterCard, wantOK: true},
	{name: "mastercard 2721 above range", input: cardDigits("2721", 16), wantOK: false},
	{name: "mastercard 51 bottom", input: cardDigits("51", 16), wantCardBrand: validators.CardBrandMasterCard, wantOK: true},
	{name: "prefix 50 not mastercard", input: cardDigits("50", 16), wantOK: false},
	{name: "prefix 56 not mastercard", input: cardDigits("56", 16), wantOK: false},

	// Discover boundaries.
	{name: "discover 643 below range", input: cardDigits("643", 16), wantOK: false},
	{name: "discover 644 bottom", input: cardDigits("644", 16), wantCardBrand: validators.CardBrandDiscover, wantOK: true},
	{name: "discover 649 top", input: cardDigits("649", 16), wantCardBrand: validators.CardBrandDiscover, wantOK: true},
	{name: "discover 65", input: cardDigits("65", 16), wantCardBrand: validators.CardBrandDiscover, wantOK: true},
	{name: "prefix 6010 not discover", input: cardDigits("6010", 16), wantOK: false},

	// JCB boundaries.
	{name: "jcb 3527 below range", input: cardDigits("3527", 16), wantOK: false},
	{name: "jcb 3589 top", input: cardDigits("3589", 16), wantCardBrand: validators.CardBrandJCB, wantOK: true},
	{name: "jcb 3590 above range", input: cardDigits("3590", 16), wantOK: false},

	// Diners boundaries.
	{name: "diners 300 bottom", input: cardDigits("300", 14), wantCardBrand: validators.CardBrandDiners, wantOK: true},
	{name: "diners 305 top", input: cardDigits("305", 14), wantCardBrand: validators.CardBrandDiners, wantOK: true},
	{name: "prefix 306 not diners", input: cardDigits("306", 14), wantOK: false},
	{name: "diners 39", input: cardDigits("39", 16), wantCardBrand: validators.CardBrandDiners, wantOK: true},

	// Real prefix, wrong length for that CardBrand.
	{name: "amex prefix with 16 digits", input: cardDigits("34", 16), wantOK: false},
	{name: "mastercard with 15 digits", input: cardDigits("55", 15), wantOK: false},
	{name: "visa with 14 digits", input: cardDigits("4", 14), wantOK: false},
	{name: "unionpay with 15 digits", input: cardDigits("62", 15), wantOK: false},

	// Common false positives.
	{name: "all zeros", input: "0000000000000000", wantOK: false},
	{name: "all nines", input: "9999999999999999", wantOK: false},
	{name: "unused prefix 1", input: cardDigits("1", 16), wantOK: false},
	{name: "unused prefix 7", input: cardDigits("7", 16), wantOK: false},
	{name: "unused prefix 8", input: cardDigits("8", 16), wantOK: false},

	// Bad input shapes.
	{name: "empty", input: "", wantOK: false},
	{name: "too short", input: cardDigits("4", 12), wantOK: false},
	{name: "too long", input: cardDigits("4", 20), wantOK: false},
	{name: "contains spaces", input: "4111 1111 1111 1111", wantOK: false},
	{name: "contains letter", input: "41111111111111a1", wantOK: false},
}

func (tc cardCase) run(t *testing.T) {
	CardBrand, ok := validators.IsPlausibleCard(tc.input)
	if ok != tc.wantOK {
		t.Fatalf("IsPlausibleCard(%q) ok = %v, want %v", tc.input, ok, tc.wantOK)
	}
	if CardBrand != tc.wantCardBrand {
		t.Errorf("IsPlausibleCard(%q) CardBrand = %q, want %q", tc.input, CardBrand, tc.wantCardBrand)
	}
}

func TestIsPlausibleCard(t *testing.T) {
	for _, tc := range cardCases {
		t.Run(tc.name, tc.run)
	}
}

func FuzzIsPlausibleCard(f *testing.F) {
	for _, tc := range cardCases {
		f.Add(tc.input)
	}
	f.Fuzz(checkCardProperties)
}

// checkCardProperties asserts invariants for any input: no panic, and an
// accepted value is 13–19 ASCII digits with a non-empty CardBrand.
func checkCardProperties(t *testing.T, s string) {
	CardBrand, ok := validators.IsPlausibleCard(s)
	if !ok {
		if CardBrand != "" {
			t.Errorf("IsPlausibleCard(%q) rejected but returned CardBrand %q", s, CardBrand)
		}
		return
	}
	if len(s) < 13 || len(s) > 19 {
		t.Errorf("IsPlausibleCard accepted %q with length %d", s, len(s))
	}
	if !isASCIIDigits(s) {
		t.Errorf("IsPlausibleCard accepted %q, which contains non-digit bytes", s)
	}
	if CardBrand == "" {
		t.Errorf("IsPlausibleCard accepted %q with an empty CardBrand", s)
	}
}

func BenchmarkIsPlausibleCard(b *testing.B) {
	for i := 0; i < b.N; i++ {
		validators.IsPlausibleCard("4111111111111111")
	}
}
