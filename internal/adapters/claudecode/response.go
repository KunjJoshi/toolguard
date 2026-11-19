package claudecode

import (
	"encoding/json"
	"fmt"
	"io"
)

type preToolUseOutput struct {
	HookSpecificOutput preToolUseDecision `json:"hookSpecificOutput"`
}

type preToolUseDecision struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

type postToolUseOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func WriteBlock(w io.Writer, hookEvent, reason string) error {
	out, err := blockOutput(hookEvent, reason)
	if err != nil {
		return err
	}

	data, err := json.Marshal(out)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(w, string(data))
	return err
}

func blockOutput(hookEvent, reason string) (any, error) {
	switch hookEvent {
	case HookPreToolUse:
		return preToolUseOutput{HookSpecificOutput: preToolUseDecision{
			HookEventName:            HookPreToolUse,
			PermissionDecision:       "deny",
			PermissionDecisionReason: reason,
		}}, nil

	case HookPostToolUse:
		return postToolUseOutput{Decision: "block", Reason: reason}, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedHookEvent, hookEvent)
	}

}
