// Package frame provides data structures and operations for database frames.
package frame

// Op represents a database operation type.
type Op uint8

const (
	// OpDefault represents the default operation.
	OpDefault Op = 0
	// OpGet represents a get operation.
	OpGet Op = 1
	// OpDel represents a delete operation.
	OpDel Op = 2
	// OpSet represents a set operation.
	OpSet Op = 3
	// OpAuth represents an authentication operation.
	OpAuth Op = 4
	// OpPing represents a ping operation.
	OpPing Op = 5
	// OpPong represents a pong operation.
	OpPong Op = 6
	// OpClosing represents a closing operation.
	OpClosing Op = 7
)

// opNames maps operations to their string representations.
var opNames = map[Op]string{
	OpGet:     "GET",
	OpDel:     "DELETE",
	OpSet:     "SET",
	OpAuth:    "AUTH",
	OpPing:    "PING",
	OpPong:    "PONG",
	OpClosing: "CLOSING",
}

func (op Op) String() string {
	if name, ok := opNames[op]; ok {
		return name
	}
	return "[INVALID]"
}

// Has checks if the operation matches the given operation.
func (op Op) Has(o Op) bool {
	return op == o
}
