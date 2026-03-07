package frame

import (
	"bytes"
	"encoding/binary"
)

// Frame represents a generic protocol frame.
// Different frame types (Query, Payload, Control, Error) implement this interface
// to provide a unified way to build and serialize protocol messages.
type Frame interface {
	SetOp(Op)                  // Sets the operation code of the frame
	SetID(uint32)              // Sets the request/response identifier
	SetKeyLength(uint32)       // Sets the key length (used by Query frames)
	Write([]byte) (int, error) // Writes payload bytes into the frame buffer
	Bytes() []byte             // Serializes the frame into its wire format
}

// Query represents a client query frame.
// It contains an operation, request ID, key length, and a payload buffer.
type Query struct {
	Op     Op
	ID     uint32
	KeyLen uint32
	Buffer []byte
	offset int
}

// SetID assigns the request identifier.
func (q *Query) SetID(Id uint32) {
	q.ID = Id
}

// Write copies data into the internal buffer starting at the current offset.
// The offset advances by the number of bytes written.
func (q *Query) Write(b []byte) (int, error) {
	n := copy(q.Buffer[q.offset:], b)
	q.offset += n
	return n, nil
}

// Bytes serializes the query into the protocol wire format.
//
// Layout:
// [type][op][id][keyLen][payloadLen][payload]
func (q *Query) Bytes() []byte {
	buf := bytes.NewBuffer([]byte{byte(TypeQuery), byte(q.Op)})
	binary.Write(buf, binary.LittleEndian, q.ID)
	binary.Write(buf, binary.LittleEndian, q.KeyLen)
	binary.Write(buf, binary.LittleEndian, uint32(len(q.Buffer)))
	buf.Write(q.Buffer)
	return buf.Bytes()
}

// SetOp sets the query operation.
func (q *Query) SetOp(op Op) {
	q.Op = op
}

// SetKeyLength sets the key length field of the query.
func (q *Query) SetKeyLength(n uint32) {
	q.KeyLen = n
}

// Payload represents a response payload frame.
// It carries a request ID and raw payload bytes.
type Payload struct {
	Id     uint32
	Buffer []byte
	offset int
}

// SetID assigns the request identifier associated with the payload.
func (p *Payload) SetID(Id uint32) {
	p.Id = Id
}

// Write copies data into the payload buffer starting at the current offset.
func (p *Payload) Write(b []byte) (int, error) {
	n := copy(p.Buffer[p.offset:], b)
	p.offset += n
	return n, nil
}

// Bytes serializes the payload into the protocol wire format.
//
// Layout:
// [type][id][payloadLen][payload]
func (p *Payload) Bytes() []byte {
	buf := bytes.NewBuffer([]byte{byte(TypePayload)})
	binary.Write(buf, binary.LittleEndian, p.Id)
	binary.Write(buf, binary.LittleEndian, uint32(len(p.Buffer)))
	buf.Write(p.Buffer)
	return buf.Bytes()
}

// SetOp is a no-op for Payload frames (required to satisfy Frame interface).
func (p *Payload) SetOp(op Op) {}

// SetKeyLength is a no-op for Payload frames.
func (p *Payload) SetKeyLength(uint32) {}

// Control represents a control frame used for signaling protocol actions
// such as ping, shutdown, or handshake operations.
type Control struct {
	Op Op
}

// SetID is a no-op for Control frames.
func (c *Control) SetID(Id uint32) {}

// Write is a no-op for Control frames since they do not carry payload data.
func (c *Control) Write([]byte) (int, error) { return 0, nil }

// Bytes serializes the control frame.
//
// Layout:
// [type][op]
func (c *Control) Bytes() []byte {
	return []byte{byte(TypeControl), byte(c.Op)}
}

// SetOp sets the control operation code.
func (c *Control) SetOp(op Op) { c.Op = op }

// SetKeyLength is a no-op for Control frames.
func (c *Control) SetKeyLength(uint32) {}

// Error represents an error frame returned to a client.
// It carries the request ID and a textual error message.
type Error struct {
	Id     uint32
	Buffer []byte
	offset int
}

// Write copies error message bytes into the internal buffer.
func (e *Error) Write(b []byte) (int, error) {
	n := copy(e.Buffer[e.offset:], b)
	e.offset += n
	return n, nil
}

// SetID assigns the request identifier associated with the error.
func (e *Error) SetID(Id uint32) {
	e.Id = Id
}

// Bytes serializes the error frame.
//
// Layout:
// [type][id][messageLen][message]
func (e *Error) Bytes() []byte {
	b := bytes.NewBuffer([]byte{byte(TypeError)})
	binary.Write(b, binary.LittleEndian, e.Id)
	binary.Write(b, binary.LittleEndian, uint32(len(e.Buffer)))
	b.Write(e.Buffer)
	return b.Bytes()
}

// SetOp is a no-op for Error frames.
func (e *Error) SetOp(op Op) {}

// SetKeyLength is a no-op for Error frames.
func (e *Error) SetKeyLength(uint32) {}

// NewError creates a new error frame from a request ID and error value.
// If err is nil, the error payload will be empty.
func NewError(Id uint32, err error) *Error {
	b := []byte{}
	if err != nil {
		b = append(b, []byte(err.Error())...)
	}
	return &Error{
		Id:     Id,
		Buffer: b,
	}
}

// NewControl creates a new control frame with the given operation.
func NewControl(op Op) *Control {
	return &Control{
		Op: op,
	}
}

// NewPayload creates a new payload frame with a request ID and payload bytes.
func NewPayload(Id uint32, val []byte) *Payload {
	return &Payload{
		Id:     Id,
		Buffer: val,
	}
}

// NewQuery creates an empty query frame.
func NewQuery() *Query {
	return &Query{}
}

// NewFrame is a factory function that constructs a frame
// based on the provided frame type.
func NewFrame(t Type) Frame {
	switch t {
	case TypeControl:
		return NewControl(OpDefault)
	case TypeQuery:
		return NewQuery()
	case TypePayload:
		return NewPayload(0, nil)
	case TypeError:
		return NewError(0, nil)
	}
	return nil
}
