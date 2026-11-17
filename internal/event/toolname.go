package event

import "strings"

const (
	mcpPrefix    = "mcp__"
	mcpSeparator = "__"
)

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
