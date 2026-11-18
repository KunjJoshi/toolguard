package event

import "strings"

const (
	mcpPrefix    = "mcp__"
	mcpSeparator = "__"
)

// SplitToolName splits a Claude Code MCP tool name of the form
// "mcp__<server>__<tool>" into server and tool. Any other name, including a
// malformed MCP name, is returned unchanged as tool with an empty server.
func SplitToolName(name string) (server, tool string) {
	rest, ok := strings.CutPrefix(name, mcpPrefix)
	if !ok {
		return "", name
	}

	server, tool, ok = strings.Cut(rest, mcpSeparator)
	if !ok || server == "" || tool == "" {
		return "", name
	}

	return server, tool
}
