package detect

import (
	"regexp"
	"sort"
	"unicode/utf8"

	"github.com/KunjJoshi/toolguard/internal/detect/validators"
)

const CardDetectorName = "credit_card"
const (
	minDigitsForCard = 13
	maxDigitsForCard = 19
)

var cardCandidate = regexp.MustCompile(`\p{Nd}(?:[\s\-._,/\p{Zs}\p{Pd}\p{Cf}\p{Mn}\p{Me}]{0,3}\p{Nd}){12,}`)
var digitRun = regexp.MustCompile(`\p{Nd}+`)

type CardOptions struct {
	ScanInsideLongRuns bool
}

type span struct {
	start int
	end   int
}

func FindCards(text string) []Match {
	return FindCardsWithOptions(text, CardOptions{})
}

func FindCardsWithOptions(text string, opts CardOptions) []Match {
	matches := make([]Match, 0)
	for _, loc := range cardCandidate.FindAllStringIndex(text, -1) {
		matches = append(matches, cardsInCandidate(text, loc[0], loc[1], opts)...)
	}
	return matches
}

func cardsInCandidate(text string, locStart, locEnd int, opts CardOptions) []Match {
	groups := digitGroups(text, locStart, locEnd)
	cards := cardsInGroups(groups, text)

	if opts.ScanInsideLongRuns {
		cards = append(cards, windowCards(text, groups, cards)...)
		sort.Sort(byStart(cards))
	}

	return cards
}

func windowCards(text string, groups []span, existing []Match) []Match {

	cards := make([]Match, 0)
	for _, g := range groups {
		if overlapsAny(g, existing) {
			continue
		}

		cards = append(cards, windowCardsInGroup(text, g)...)
	}
	return cards
}

func overlapsAny(g span, existing []Match) bool {
	g_start := g.start
	g_end := g.end
	for _, m := range existing {
		m_start := m.Start
		m_end := m.End

		if g_start > m_end || g_end < m_start {
			continue
		} else {
			return true
		}
	}
	return false
}

func windowCardsInGroup(text string, g span) []Match {
	offsets := digitOffsets(text, g)
	count := len(offsets) - 1
	cards := make([]Match, 0)
	pos := 0

	for pos+minDigitsForCard <= count {
		m, length, ok := longestWindowAt(text, offsets, pos)
		if ok {
			cards = append(cards, m)
			pos += length
			continue
		}
		pos++
	}
	return cards
}

func longestWindowAt(text string, offsets []int, pos int) (Match, int, bool) {
	count := len(offsets) - 1
	for length := maxDigitsForCard; length >= minDigitsForCard; length-- {
		if pos+length > count {
			continue
		}

		s := span{start: offsets[pos], end: offsets[pos+length]}
		if m, ok := cardAt(text, s); ok {
			return m, length, true
		}
	}

	return Match{}, 0, false
}

func digitOffsets(text string, g span) []int {
	offsets := make([]int, 0, g.end-g.start+1)
	for i := g.start; i < g.end; {
		offsets = append(offsets, i)
		_, size := utf8.DecodeRuneInString(text[i:g.end])
		i += size
	}

	return append(offsets, g.end)
}

func digitGroups(text string, start, end int) []span {
	testText := text[start:end]
	locs := digitRun.FindAllStringIndex(testText, -1)
	groups := make([]span, 0, len(locs))
	for _, l := range locs {
		groups = append(groups, span{start: start + l[0], end: start + l[1]})
	}
	return groups
}

func cardsInGroups(groups []span, text string) []Match {
	cards := make([]Match, 0)
	i := 0
	for i < len(groups) {
		m, next, ok := longestCardFrom(text, groups, i)
		if ok {
			cards = append(cards, m)
			i = next
			continue
		}

		i++
	}
	return cards
}

func cardAt(text string, s span) (Match, bool) {
	digits := NormalizeDigits(text[s.start:s.end])

	brand, ok := validators.IsPlausibleCard(digits)
	if !ok {
		return Match{}, false
	}

	if !validators.Luhn(digits) {
		return Match{}, false
	}

	return Match{
		Detector: CardDetectorName,
		Subtype:  string(brand),
		Start:    s.start,
		End:      s.end,
		Masked:   MaskFirst4Last4(digits),
	}, true
}

func longestCardFrom(text string, groups []span, first int) (Match, int, bool) {
	for last := len(groups) - 1; last >= first; last-- {
		s := span{start: groups[first].start, end: groups[last].end}
		if m, ok := cardAt(text, s); ok {
			return m, last + 1, true
		}
	}
	return Match{}, 0, false
}

type byStart []Match

func (b byStart) Len() int           { return len(b) }
func (b byStart) Less(i, j int) bool { return b[i].Start < b[j].Start }
func (b byStart) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
