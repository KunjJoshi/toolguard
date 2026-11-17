package claudecode_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/adapters/claudecode"
	"github.com/KunjJoshi/toolguard/internal/event"
)

// testdataPath points at the shared testdata/events folder at the repo root.
func testdataPath(name string) string {
	return filepath.Join("..", "..", "..", "testdata", "events", name)
}

func loadPayload(t *testing.T, name string) claudecode.Payload {
	t.Helper()
	f, err := os.Open(testdataPath(name))
	if err != nil {
		t.Fatalf("open %s: %v", name, err)
	}
	defer f.Close()

	p, err := claudecode.ParsePayload(f)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return p
}

func TestPreToolUseToEvent(t *testing.T) {
	ev := loadPayload(t, "pre_card.json").ToEvent()

	if ev.Direction != event.DirRequest {
		t.Errorf("Direction = %q, want %q", ev.Direction, event.DirRequest)
	}
	if ev.SessionID != "test-session-001" {
		t.Errorf("SessionID = %q", ev.SessionID)
	}
	if ev.Server != "mail" || ev.Tool != "send_email" {
		t.Errorf("Server, Tool = %q, %q, want mail, send_email", ev.Server, ev.Tool)
	}
	if got := strings.Join(ev.ArgNames, ","); got != "body,subject,to" {
		t.Errorf("ArgNames = %q, want body,subject,to", got)
	}
	if len(ev.ArgFields) != 3 || ev.ArgFields[0].Path != "body" {
		t.Errorf("ArgFields = %+v, want 3 fields starting with body", ev.ArgFields)
	}
	if ev.ResultFields != nil {
		t.Errorf("ResultFields = %+v, want none for PreToolUse", ev.ResultFields)
	}
}

func TestPostToolUseToEvent(t *testing.T) {
	ev := loadPayload(t, "post_card.json").ToEvent()

	if ev.Direction != event.DirResponse {
		t.Errorf("Direction = %q, want %q", ev.Direction, event.DirResponse)
	}
	if ev.Server != "crm" || ev.Tool != "get_customer" {
		t.Errorf("Server, Tool = %q, %q, want crm, get_customer", ev.Server, ev.Tool)
	}
	if len(ev.ResultFields) != 2 {
		t.Fatalf("ResultFields = %+v, want 2", ev.ResultFields)
	}
	if ev.ResultFields[0].Path != "content[0].text" {
		t.Errorf("first result path = %q, want content[0].text", ev.ResultFields[0].Path)
	}
}

func TestBuiltInToolHasNoServer(t *testing.T) {
	raw := `{"hook_event_name":"PreToolUse","tool_name":"Bash","tool_input":{"command":"ls"}}`
	p, err := claudecode.ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	ev := p.ToEvent()
	if ev.Server != "" || ev.Tool != "Bash" {
		t.Errorf("Server, Tool = %q, %q, want \"\", Bash", ev.Server, ev.Tool)
	}
}

func TestLongNumbersKeepEveryDigit(t *testing.T) {
	raw := `{"hook_event_name":"PreToolUse","tool_name":"mcp__pay__charge","tool_input":{"card":4000000000000000002}}`
	p, err := claudecode.ParsePayload(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	fields := p.ToEvent().ArgFields
	if len(fields) != 1 || fields[0].Value != "4000000000000000002" {
		t.Errorf("ArgFields = %+v, want the exact 19-digit number", fields)
	}
}

func TestUnknownFieldsAreIgnored(t *testing.T) {
	raw := `{"hook_event_name":"PreToolUse","tool_name":"Read","tool_input":{},"some_future_field":123}`
	if _, err := claudecode.ParsePayload(strings.NewReader(raw)); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

type parseErrorCase struct {
	name    string
	raw     string
	wantErr error
}

var parseErrorCases = []parseErrorCase{
	{name: "invalid json", raw: `{"hook_event_name":`, wantErr: claudecode.ErrInvalidJSON},
	{name: "empty input", raw: ``, wantErr: claudecode.ErrInvalidJSON},
	{name: "unsupported hook", raw: `{"hook_event_name":"UserPromptSubmit","tool_name":"x"}`, wantErr: claudecode.ErrUnsupportedHookEvent},
	{name: "missing hook name", raw: `{"tool_name":"Bash"}`, wantErr: claudecode.ErrUnsupportedHookEvent},
	{name: "missing tool name", raw: `{"hook_event_name":"PreToolUse"}`, wantErr: claudecode.ErrMissingToolName},
}

func (tc parseErrorCase) run(t *testing.T) {
	_, err := claudecode.ParsePayload(strings.NewReader(tc.raw))
	if !errors.Is(err, tc.wantErr) {
		t.Errorf("ParsePayload(%q) error = %v, want %v", tc.raw, err, tc.wantErr)
	}
}

func TestParsePayloadErrors(t *testing.T) {
	for _, tc := range parseErrorCases {
		t.Run(tc.name, tc.run)
	}
}

func TestPayloadTooLarge(t *testing.T) {
	big := strings.NewReader(strings.Repeat("a", claudecode.MaxPayloadBytes+1))
	if _, err := claudecode.ParsePayload(big); !errors.Is(err, claudecode.ErrPayloadTooLarge) {
		t.Errorf("error = %v, want %v", err, claudecode.ErrPayloadTooLarge)
	}
}
