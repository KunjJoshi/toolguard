package validators

const minLuhnLength = 2

func Luhn(digits string) bool {
	if len(digits) < minLuhnLength {
		return false
	}

	sum := 0
	double := false

	for i := len(digits) - 1; i >= 0; i-- {
		d, ok := digitsValue(digits[i])
		if !ok {
			return false
		}

		if double {
			d = doubleLuhnDigit(d)
		}
		sum += d
		double = !double
	}

	return sum%10 == 0
}

func digitsValue(c byte) (int, bool) {
	if c < '0' || c > '9' {
		return 0, false
	}
	return int(c - '0'), true
}

func doubleLuhnDigit(d int) int {
	d *= 2
	if d > 9 {
		d -= 9
	}
	return d
}
