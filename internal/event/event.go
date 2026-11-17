package event

type Direction string

const (
	DirRequest Direction = "request"

	DirResponse Direction = "response"

	DirToolsList Direction = "tools_list"
)

type Event struct {
	Direction Direction
	SessionID string

	Server string
	Tool   string

	ArgFields    []Field
	ArgNames     []string
	ResultFields []Field
}
