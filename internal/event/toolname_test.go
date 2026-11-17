package event_test

import (
	"testing"

	"github.com/KunjJoshi/toolguard/internal/event"
)

type toolNameCase struct {
	input      string
	wantServer string
	wantTool   string
}

var toolNameCases = []toolNameCase{
	{input: "mcp__mail__send_email", wantServer: "mail", wantTool: "send_email"},
	{input: "mcp__github__create_issue", wantServer: "github", wantTool: "create_issue"},
	{input: "mcp__srv__a__b", wantServer: "srv", wantTool: "a__b"},
	{input: "mcp__my-server__do_thing", wantServer: "my-server", wantTool: "do_thing"},
	{input: "Bash", wantServer: "", wantTool: "Bash"},
	{input: "WebFetch", wantServer: "", wantTool: "WebFetch"},
	{input: "mcp__onlyserver", wantServer: "", wantTool: "mcp__onlyserver"},
	{input: "mcp____tool", wantServer: "", wantTool: "mcp____tool"},
	{input: "mcp__server__", wantServer: "", wantTool: "mcp__server__"},
	{input: "", wantServer: "", wantTool: ""},
}

func (tc toolNameCase) run(t *testing.T) {
	server, tool := event.SplitToolName(tc.input)
	if server != tc.wantServer || tool != tc.wantTool {
		t.Errorf("SplitToolName(%q) = (%q, %q), want (%q, %q)", tc.input, server, tool, tc.wantServer, tc.wantTool)
	}
}

func TestSplitToolName(t *testing.T) {
	for _, tc := range toolNameCases {
		t.Run(tc.input, tc.run)
	}
}
