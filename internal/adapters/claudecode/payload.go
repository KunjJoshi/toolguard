// Package claudecode converts Claude Code hook payloads into ToolGuard events.
package claudecode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/KunjJoshi/toolguard/internal/event"
)

// Hook event names that ToolGuard accepts in a payload's hook_event_name.
const (
	HookPreToolUse  = "PreToolUse"
	HookPostToolUse = "PostToolUse"
)

// MaxPayloadBytes is the largest hook payload ParsePayload will read (16 MiB).
const MaxPayloadBytes = 16 << 20

// Errors returned by ParsePayload and Payload.Validate.
var (
	ErrPayloadTooLarge      = errors.New("hook payload exceeds size limit")
	ErrInvalidJSON          = errors.New("hook payload is not valid JSON")
	ErrUnsupportedHookEvent = errors.New("unsupported hook event")
	ErrMissingToolName      = errors.New("hook payload has no tool_name")
)

// Payload is the JSON object Claude Code sends to a PreToolUse or PostToolUse
// hook on stdin. ToolResponse is only set for PostToolUse.
type Payload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	ToolName       string `json:"tool_name"`
	ToolInput      any    `json:"tool_input"`
	ToolResponse   any    `json:"tool_response"`
}

// ParsePayload reads one hook payload from r, decodes it (keeping numbers as
// json.Number) and validates it. It fails with ErrPayloadTooLarge if r holds
// more than MaxPayloadBytes, with ErrInvalidJSON if the JSON is malformed, and
// with any error from Validate.
func ParsePayload(r io.Reader) (Payload, error) {
	raw, err := io.ReadAll(io.LimitReader(r, MaxPayloadBytes+1))
	if err != nil {
		return Payload{}, fmt.Errorf("read hook payload: %w", err)
	}

	if len(raw) > MaxPayloadBytes {
		return Payload{}, ErrPayloadTooLarge
	}

	var p Payload

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()

	if err := dec.Decode(&p); err != nil {
		return Payload{}, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if err := p.Validate(); err != nil {
		return Payload{}, err
	}

	return p, nil
}

// Validate reports whether p is a supported PreToolUse or PostToolUse payload
// with a tool name. It returns ErrUnsupportedHookEvent or ErrMissingToolName.
func (p Payload) Validate() error {
	if p.HookEventName != HookPostToolUse && p.HookEventName != HookPreToolUse {
		return fmt.Errorf("%w: %q", ErrUnsupportedHookEvent, p.HookEventName)
	}

	if p.ToolName == "" {
		return ErrMissingToolName
	}

	return nil
}

// ToEvent converts p into a normalized event: the tool name is split into MCP
// server and tool, the tool input is flattened into ArgFields, and for
// PostToolUse the tool response is flattened into ResultFields.
func (p Payload) ToEvent() event.Event {
	server, tool := event.SplitToolName(p.ToolName)

	ev := event.Event{
		Direction: directionFor(p.HookEventName),
		SessionID: p.SessionID,
		Server:    server,
		Tool:      tool,
		ArgFields: event.Flatten(p.ToolInput),
		ArgNames:  topLevelNames(p.ToolInput),
	}

	if ev.Direction == event.DirResponse {
		ev.ResultFields = event.Flatten(p.ToolResponse)
	}

	return ev
}

func directionFor(hookevent string) event.Direction {
	if hookevent == HookPostToolUse {
		return event.DirResponse
	}
	return event.DirRequest
}

func topLevelNames(input any) []string {
	obj, ok := input.(map[string]any)
	if !ok {
		return nil
	}
	return event.SortedKeys(obj)
}
