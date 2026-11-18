// Package event defines the normalized tool-call event that detectors and
// rules inspect, independent of the agent that produced it.
package event

// Direction says which part of a tool call an Event describes.
type Direction string

const (
	// DirRequest is a tool call about to run (its arguments).
	DirRequest Direction = "request"

	// DirResponse is a tool call that has run (its arguments and result).
	DirResponse Direction = "response"

	// DirToolsList is a server's list of available tools.
	DirToolsList Direction = "tools_list"
)

// Event is one normalized tool call. Server is empty for built-in tools.
// ArgFields and ResultFields hold every string and number in the arguments
// and result, flattened with their paths; ArgNames lists the top-level
// argument names.
type Event struct {
	Direction Direction
	SessionID string

	Server string
	Tool   string

	ArgFields    []Field
	ArgNames     []string
	ResultFields []Field
}
