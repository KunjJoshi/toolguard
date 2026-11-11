package detect

import "strings"

const (
	minLenFirst4Last4 = 12
	minLenLast4       = 8
	visibleCharacters = 4
)

func MaskFirst4Last4(s string) string {
	n := len(s)
	if n < minLenFirst4Last4 {
		return maskAll(n)
	}

	hidden := n - 2*visibleCharacters
	return string(s[:visibleCharacters]) + maskAll(hidden) + string(s[n-visibleCharacters:])
}

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
