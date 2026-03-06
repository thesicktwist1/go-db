package frame

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrMalformed         = errors.New("error malformed request")
	ErrInvalidPayloadLen = errors.New("error invalid payload length")
)

var (
	Closing = []byte{byte(TypeControl), byte(OpClosing)}
	Ping    = []byte{byte(TypeControl), byte(OpPing)}
	Pong    = []byte{byte(TypeControl), byte(OpPong)}
)

const headerSize = 4

type Type uint8
type ParserState uint8

const (
	TypeDefault Type = 0
	TypeControl Type = 1
	TypeQuery   Type = 2
	TypePayload Type = 3
	TypeError   Type = 4
)

const (
	StateInit              ParserState = 1
	StateOpParsing         ParserState = 2
	StateKeyLenParsing     ParserState = 3
	StatePayloadLenParsing ParserState = 4
	StatePayloadParsing    ParserState = 5
	StateIdParsing         ParserState = 6
	StateDone              ParserState = 7
)

type Parser struct {
	State ParserState
	Frame
}

type Frame struct {
	Type      Type
	Id        uint32
	Op        Op
	Buffer    bytes.Buffer
	KeyLen    int
	available int
}

func NewFrame() Frame {
	return Frame{
		Type:   TypeDefault,
		Op:     OpDefault,
		Buffer: *new(bytes.Buffer),
	}
}

func NewParser() *Parser {
	return &Parser{
		State: StateInit,
		Frame: NewFrame(),
	}
}

func (p *Parser) Parse(buf []byte) (int, error) {
	readN := 0
outerLoop:
	for {
		if readN >= len(buf) {
			break
		}
		current := buf[readN:]
		switch p.State {
		case StateInit:
			p.Type = Type(current[0])
			readN++
			switch p.Type {
			case TypeError, TypePayload:
				p.State = StateIdParsing
			default:
				p.State = StateOpParsing
			}
		case StateOpParsing:
			p.Op = Op(current[0])
			readN++
			switch p.Op {
			case OpDel, OpGet, OpSet:
				p.State = StateIdParsing
			case OpAuth:
				p.State = StatePayloadLenParsing
			default:
				p.State = StateDone
			}
		case StateIdParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			p.Id = binary.LittleEndian.Uint32(current[:headerSize])
			readN += headerSize
			switch p.Type {
			case TypeError, TypePayload:
				p.State = StatePayloadLenParsing
			default:
				p.State = StateKeyLenParsing
			}
		case StateKeyLenParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			KeyLen := binary.LittleEndian.Uint32(current[:headerSize])
			p.KeyLen = int(KeyLen)
			readN += headerSize
			p.State = StatePayloadLenParsing
		case StatePayloadLenParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			size := binary.LittleEndian.Uint32(current[:headerSize])
			if p.KeyLen > int(size) {
				return 0, ErrMalformed
			}
			if size == 0 {
				p.State = StateDone
			} else {
				p.State = StatePayloadParsing
			}
			readN += headerSize
			p.available = int(size)
		case StatePayloadParsing:
			rem := min(len(current), int(p.available))
			n, _ := p.Buffer.Write(current[:rem])
			p.available -= n
			readN += n
			if p.available == 0 {
				p.State = StateDone
			}
		case StateDone:
			break outerLoop
		default:
			panic("something went really wrong... (invalid state)")
		}

	}
	return readN, nil
}

func Query(Op Op, id uint32, key string, val []byte) []byte {
	buf := bytes.NewBuffer([]byte{byte(TypeQuery), byte(Op)})
	size := len(key) + len(val)
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, uint32(len(key)))
	binary.Write(buf, binary.LittleEndian, uint32(size))
	buf.WriteString(key)
	buf.Write(val)
	return buf.Bytes()
}

func Payload(id uint32, val []byte) []byte {
	buf := bytes.NewBuffer([]byte{byte(TypePayload)})
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, uint32(len(val)))
	buf.Write(val)
	return buf.Bytes()
}

func Error(id uint32, err error) []byte {
	buf := bytes.NewBuffer([]byte{byte(TypeError)})
	binary.Write(buf, binary.LittleEndian, id)
	binary.Write(buf, binary.LittleEndian, uint32(len(err.Error())))
	buf.WriteString(err.Error())
	return buf.Bytes()
}

func (p *Parser) Done() bool {
	return p.State == StateDone
}

func (p *Parser) Reset() {
	p.State = StateInit

	p.Frame.Type = TypeDefault
	p.Frame.Op = OpDefault
	p.Frame.Buffer.Reset()
	p.Frame.available = 0
	p.Frame.KeyLen = 0
}
