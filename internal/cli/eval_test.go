package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func testdataPath(name string) string {
	return filepath.Join("..", "..", "testdata", "events", name)
}

// runCLI executes the CLI with args and returns everything it printed.
func runCLI(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

type evalCase struct {
	name     string
	file     string
	want     []string
	mustSkip []string
}

var evalCases = []evalCase{
	{
		name: "card in request is detected and masked",
		file: "pre_card.json",
		want: []string{
			"Event:      request mail/send_email",
			"Session:    test-session-001",
			"Detections: 1",
			"args.body  credit_card/visa 4111********1111 [22:41]",
		},
		mustSkip: []string{"4111 1111 1111 1111"},
	},
	{
		name: "clean request has no detections",
		file: "pre_clean.json",
		want: []string{"No detections."},
	},
	{
		name: "luhn-invalid number is not flagged",
		file: "pre_near_miss.json",
		want: []string{"No detections."},
	},
	{
		name: "card in response result is detected",
		file: "post_card.json",
		want: []string{
			"Event:      response crm/get_customer",
			"Detections: 1",
			"result.content[0].text  credit_card/mastercard 5555********4444 [24:43]",
		},
		mustSkip: []string{"5555-5555-5555-4444"},
	},
}

func (tc evalCase) run(t *testing.T) {
	out, err := runCLI(t, "", "eval", testdataPath(tc.file))
	if err != nil {
		t.Fatalf("eval %s: %v\n%s", tc.file, err, out)
	}
	for _, w := range tc.want {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q\n--- output ---\n%s", w, out)
		}
	}
	for _, s := range tc.mustSkip {
		if strings.Contains(out, s) {
			t.Errorf("output leaks raw value %q\n--- output ---\n%s", s, out)
		}
	}
}

func TestEval(t *testing.T) {
	for _, tc := range evalCases {
		t.Run(tc.name, tc.run)
	}
}

func TestEvalFromStdin(t *testing.T) {
	payload := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"echo 4111111111111111"}}`
	out, err := runCLI(t, payload, "eval", "-")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "args.command  credit_card/visa 4111********1111 [5:21]") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestEvalMissingFile(t *testing.T) {
	if _, err := runCLI(t, "", "eval", testdataPath("does_not_exist.json")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

func TestEvalInvalidPayload(t *testing.T) {
	if _, err := runCLI(t, `{"tool_name":"Bash"}`, "eval", "-"); err == nil {
		t.Error("expected an error for a payload with no hook event")
	}
}

func TestEvalRequiresOneArgument(t *testing.T) {
	if _, err := runCLI(t, "", "eval"); err == nil {
		t.Error("expected an error with no arguments")
	}
}

func TestVersion(t *testing.T) {
	out, err := runCLI(t, "", "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "toolguard dev") {
		t.Errorf("version output = %q", out)
	}
}
