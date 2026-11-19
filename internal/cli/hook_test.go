package cli

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(testdataPath(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestHookDeniesCardInRequest(t *testing.T) {
	out, err := runCLI(t, readTestdata(t, "pre_card.json"), "hook")
	if err != nil {
		t.Fatalf("hook: %v", err)
	}

	var got struct {
		HookSpecificOutput struct {
			PermissionDecision       string `json:"permissionDecision"`
			PermissionDecisionReason string `json:"permissionDecisionReason"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	h := got.HookSpecificOutput
	if h.PermissionDecision != "deny" {
		t.Errorf("permissionDecision = %q, want deny", h.PermissionDecision)
	}
	if !strings.Contains(h.PermissionDecisionReason, "4111********1111 in args.body") {
		t.Errorf("reason = %q", h.PermissionDecisionReason)
	}
	if strings.Contains(out, "4111 1111 1111 1111") {
		t.Errorf("output leaks the raw card number: %s", out)
	}
}

func TestHookAllowsCleanRequestSilently(t *testing.T) {
	for _, name := range []string{"pre_clean.json", "pre_near_miss.json"} {
		out, err := runCLI(t, readTestdata(t, name), "hook")
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if out != "" {
			t.Errorf("%s: expected no output (allow), got %q", name, out)
		}
	}
}

func TestHookFlagsCardInResult(t *testing.T) {
	out, err := runCLI(t, readTestdata(t, "post_card.json"), "hook")
	if err != nil {
		t.Fatalf("hook: %v", err)
	}

	var got struct {
		Decision string `json:"decision"`
		Reason   string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out)
	}
	if got.Decision != "block" {
		t.Errorf("decision = %q, want block", got.Decision)
	}
	if !strings.Contains(got.Reason, "5555********4444 in result.content[0].text") {
		t.Errorf("reason = %q", got.Reason)
	}
	if strings.Contains(out, "5555-5555-5555-4444") {
		t.Errorf("output leaks the raw card number: %s", out)
	}
}

func TestHookFailsClosedOnBadPayload(t *testing.T) {
	inputs := []string{``, `not json`, `{"tool_name":"Bash"}`, `{"hook_event_name":"PreToolUse"}`}
	for _, in := range inputs {
		_, err := runCLI(t, in, "hook")
		var e *exitError
		if !errors.As(err, &e) || e.code != exitBlock {
			t.Errorf("input %q: error = %v, want exit code %d", in, err, exitBlock)
		}
	}
}

func TestExitCodeFor(t *testing.T) {
	if got := exitCodeFor(errors.New("plain")); got != 1 {
		t.Errorf("plain error exit code = %d, want 1", got)
	}
	if got := exitCodeFor(&exitError{code: 2, err: errors.New("x")}); got != 2 {
		t.Errorf("exitError exit code = %d, want 2", got)
	}
}
