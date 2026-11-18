package detect

import "fmt"

// Match is one sensitive value found by a detector. It never holds the raw
// value, only its masked form and its [Start, End) byte offsets in the text.
type Match struct {
	Detector string
	Subtype  string
	Start    int
	End      int
	Masked   string
}

// String formats m as "detector/subtype masked [start:end]", omitting
// "/subtype" when Subtype is empty.
func (m Match) String() string {
	if m.Subtype == "" {
		return fmt.Sprintf("%s %s [%d:%d]", m.Detector, m.Masked, m.Start, m.End)
	} else {
		return fmt.Sprintf("%s/%s %s [%d:%d]", m.Detector, m.Subtype, m.Masked, m.Start, m.End)
	}
}
