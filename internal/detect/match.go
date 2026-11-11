package detect

import "fmt"

type Match struct {
	Detector string
	Subtype  string
	Start    int
	End      int
	Masked   string
}

func (m Match) String() string {
	if m.Subtype == "" {
		return fmt.Sprintf("%s %s [%d:%d]", m.Detector, m.Masked, m.Start, m.End)
	} else {
		return fmt.Sprintf("%s/%s %s [%d:%d]", m.Detector, m.Subtype, m.Masked, m.Start, m.End)
	}
}
