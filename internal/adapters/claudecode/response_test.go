package claudecode_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/KunjJoshi/toolguard/internal/adapters/claudecode"
)

func TestWriteBlockPreToolUse(t *testing.T) {
	var buf bytes.Buffer
	if err := claudecode.WriteBlock(&buf, claudecode.HookPreToolUse, "card found"); err != nil {
		t.Fatal(err)
	}

	var got struct {
		HookSpecificOutput struct {
			HookEventName            string `json:"hookEventName"`
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	h := got.HookSpecificOutput
	if h.HookEventName != "PreToolUse" || h.PermissionDecision != "deny" || h.PermissionDecisionReason != "card found" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestWriteBlockPostToolUse(t *testing.T) {
	var buf bytes.Buffer
	if err := claudecode.WriteBlock(&buf, claudecode.HookPostToolUse, "card in result"); err != nil {
		t.Fatal(err)
	}

	var got struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if got.Decision != "block" || got.Reason != "card in result" {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestWriteBlockUnsupportedEvent(t *testing.T) {
	var buf bytes.Buffer
	err := claudecode.WriteBlock(&buf, "UserPromptSubmit", "x")
	if !errors.Is(err, claudecode.ErrUnsupportedHookEvent) {
		t.Errorf("error = %v, want %v", err, claudecode.ErrUnsupportedHookEvent)
	}
	if buf.Len() != 0 {
		t.Errorf("wrote output for an unsupported event: %q", buf.String())
	}
}
