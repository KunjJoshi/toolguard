package detect

import "strings"

const (
	minLenFirst4Last4 = 12
	minLenLast4       = 8
	visibleCharacters = 4
)

// MaskFirst4Last4 replaces all but the first 4 and last 4 bytes of s with '*'.
// Strings shorter than 12 bytes are masked completely.
func MaskFirst4Last4(s string) string {
	n := len(s)
	if n < minLenFirst4Last4 {
		return maskAll(n)
	}

	hidden := n - 2*visibleCharacters
	return string(s[:visibleCharacters]) + maskAll(hidden) + string(s[n-visibleCharacters:])
}

// MaskLast4 replaces all but the last 4 bytes of s with '*'.
// Strings shorter than 8 bytes are masked completely.
func MaskLast4(s string) string {
	n := len(s)
	if n < minLenLast4 {
		return maskAll(n)
	}

	hidden := n - visibleCharacters
	return maskAll(hidden) + string(s[hidden:])
}

func maskAll(n int) string {
	return strings.Repeat("*", n)
}
