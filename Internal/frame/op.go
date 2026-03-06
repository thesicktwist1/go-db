// Package frame provides data structures and operations for database frames.
package frame

import "strings"

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

func (op Op) String() string {
	b := new(strings.Builder)
	if op.Has(OpGet) {
		b.WriteString("GET")
	} else if op.Has(OpDel) {
		b.WriteString("DELETE")
	} else if op.Has(OpSet) {
		b.WriteString("SET")
	} else if op.Has(OpAuth) {
		b.WriteString("AUTH")
	} else if op.Has(OpPing) {
		b.WriteString("PING")
	} else if op.Has(OpPong) {
		b.WriteString("PONG")
	} else if op.Has(OpClosing) {
		b.WriteString("CLOSING")
	} else {
		b.WriteString("[INVALID]")
	}
	return b.String()
}

// Has checks if the operation matches the given operation.
func (op Op) Has(o Op) bool {
	return op == o
}
