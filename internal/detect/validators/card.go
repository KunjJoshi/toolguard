package validators

type CardBrand string

const (
	CardBrandVisa       CardBrand = "visa"
	CardBrandAmex       CardBrand = "amex"
	CardBrandMasterCard CardBrand = "mastercard"
	CardBrandDiscover   CardBrand = "discover"
	CardBrandJCB        CardBrand = "jcb"
	CardBrandUnionPay   CardBrand = "union-pay"
	CardBrandDiners     CardBrand = "diners"
)

const (
	minCardLength = 13
	maxCardLength = 19
)

type prefixRange struct {
	low   int
	high  int
	width int
}

type brandSpec struct {
	brand   CardBrand
	ranges  []prefixRange
	lengths []int
}

var cardBrands = []brandSpec{
	{
		brand:   CardBrandVisa,
		ranges:  []prefixRange{prefix(4, 4)},
		lengths: []int{13, 16, 19},
	},

	{
		brand:   CardBrandMasterCard,
		ranges:  []prefixRange{prefix(51, 55), prefix(2221, 2720)},
		lengths: []int{16},
	},
	{
		brand:   CardBrandAmex,
		ranges:  []prefixRange{prefix(34, 34), prefix(37, 37)},
		lengths: []int{15},
	},
	{
		brand:   CardBrandDiscover,
		ranges:  []prefixRange{prefix(6011, 6011), prefix(644, 649), prefix(65, 65)},
		lengths: lengthsBetween(16, 19),
	},
	{
		brand:   CardBrandJCB,
		ranges:  []prefixRange{prefix(3528, 3589)},
		lengths: lengthsBetween(16, 19),
	},
	{
		brand:   CardBrandDiners,
		ranges:  []prefixRange{prefix(300, 305), prefix(36, 36), prefix(38, 39)},
		lengths: lengthsBetween(14, 19),
	},
	{
		brand:   CardBrandUnionPay,
		ranges:  []prefixRange{prefix(62, 62)},
		lengths: lengthsBetween(16, 19),
	},
}

func IsPlausibleCard(digits string) (CardBrand, bool) {
	if len(digits) < minCardLength || len(digits) > maxCardLength {
		return "", false
	}

	if !allDigits(digits) {
		return "", false
	}

	spec, ok := lookUpBrand(digits)
	if !ok {
		return "", false
	}

	if !spec.allowsLength(len(digits)) {
		return "", false
	}

	return spec.brand, true
}

func lookUpBrand(digits string) (brandSpec, bool) {
	for _, spec := range cardBrands {
		if spec.matchesPrefix(digits) {
			return spec, true
		}
	}
	return brandSpec{}, false
}

func (s brandSpec) matchesPrefix(digits string) bool {
	for _, r := range s.ranges {
		if r.contains(digits) {
			return true
		}
	}
	return false
}

func (s brandSpec) allowsLength(n int) bool {
	for _, l := range s.lengths {
		if l == n {
			return true
		}
	}
	return false
}

func (r prefixRange) contains(digits string) bool {
	if len(digits) < r.width {
		return false
	}
	n := leadingNumber(digits, r.width)
	return n >= r.low && n <= r.high
}

func leadingNumber(digits string, width int) int {
	n := 0
	for i := 0; i < width; i++ {
		n = n*10 + int(digits[i]-'0')
	}
	return n
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if _, ok := digitsValue(s[i]); !ok {
			return false
		}
	}
	return true
}

func prefix(low, high int) prefixRange {
	return prefixRange{low: low, high: high, width: digitCount(low)}
}

func digitCount(n int) int {
	count := 1
	for n >= 10 {
		n /= 10
		count++
	}
	return count
}

func lengthsBetween(lo, hi int) []int {
	lengths := make([]int, 0, hi-lo+1)
	for l := lo; l <= hi; l++ {
		lengths = append(lengths, l)
	}
	return lengths
}
