package event_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/event"
)

// decode parses JSON the way ToolGuard must: with UseNumber.
func decode(t *testing.T, raw string) any {
	t.Helper()
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return v
}

type flattenCase struct {
	name string
	raw  string
	want []event.Field
}

var flattenCases = []flattenCase{
	{
		name: "flat object, keys sorted",
		raw:  `{"to":"a@b.com","body":"hello"}`,
		want: []event.Field{{Path: "body", Value: "hello"}, {Path: "to", Value: "a@b.com"}},
	},
	{
		name: "nested object",
		raw:  `{"payment":{"card":"4111111111111111","name":"Jane"}}`,
		want: []event.Field{{Path: "payment.card", Value: "4111111111111111"}, {Path: "payment.name", Value: "Jane"}},
	},
	{
		name: "array keeps order",
		raw:  `{"recipients":["z@x.com","a@x.com"]}`,
		want: []event.Field{{Path: "recipients[0]", Value: "z@x.com"}, {Path: "recipients[1]", Value: "a@x.com"}},
	},
	{
		name: "objects inside arrays",
		raw:  `{"items":[{"note":"one"},{"note":"two"}]}`,
		want: []event.Field{{Path: "items[0].note", Value: "one"}, {Path: "items[1].note", Value: "two"}},
	},
	{
		name: "19-digit number survives exactly",
		raw:  `{"card":4000000000000000002}`,
		want: []event.Field{{Path: "card", Value: "4000000000000000002"}},
	},
	{
		name: "booleans and nulls skipped",
		raw:  `{"a":true,"b":null,"c":"x"}`,
		want: []event.Field{{Path: "c", Value: "x"}},
	},
	{name: "empty object", raw: `{}`, want: nil},
	{name: "top-level string", raw: `"just text"`, want: []event.Field{{Path: "", Value: "just text"}}},
}

func (tc flattenCase) run(t *testing.T) {
	got := event.Flatten(decode(t, tc.raw))
	if len(got) != len(tc.want) {
		t.Fatalf("Flatten(%s) = %v, want %v", tc.raw, got, tc.want)
	}
	for i := range got {
		if got[i] != tc.want[i] {
			t.Errorf("Flatten(%s)[%d] = %+v, want %+v", tc.raw, i, got[i], tc.want[i])
		}
	}
}

func TestFlatten(t *testing.T) {
	for _, tc := range flattenCases {
		t.Run(tc.name, tc.run)
	}
}

// TestFlattenIsDeterministic guards against Go's random map order leaking
// into results: the same input must always flatten the same way.
func TestFlattenIsDeterministic(t *testing.T) {
	v := decode(t, `{"e":"5","d":"4","c":"3","b":"2","a":"1","f":{"z":"9","y":"8"}}`)
	first := event.Flatten(v)
	for i := 0; i < 50; i++ {
		again := event.Flatten(v)
		for j := range first {
			if again[j] != first[j] {
				t.Fatalf("run %d differs at %d: %+v vs %+v", i, j, again[j], first[j])
			}
		}
	}
}

func TestSortedKeys(t *testing.T) {
	got := event.SortedKeys(map[string]any{"to": 1, "body": 2, "cc": 3})
	want := []string{"body", "cc", "to"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("SortedKeys = %v, want %v", got, want)
	}
}
