package detect

import (
	"strings"
	"unicode"
)

const maxDigitRunScan = 100

func DigitsOnly(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if isAsciiDigit(s[i]) {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func isAsciiDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func NormalizeDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		if unicode.Is(unicode.Nd, r) {
			b.WriteByte(byte('0' + unicodeDigitValue(r)))
		}
	}
	return b.String()
}

func unicodeDigitValue(r rune) int {
	n := 0
	for p := r - 1; n < maxDigitRunScan && unicode.Is(unicode.Nd, p); p-- {
		n++
	}

	return n % 10
}
