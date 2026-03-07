package frame

import (
	"encoding/binary"
	"errors"
)

var (
	ErrMalformed         = errors.New("error malformed request")
	ErrInvalidPayloadLen = errors.New("error invalid payload length")
	ErrFrameTooLarge     = errors.New("Frame exceeds maximum allowed size")
	ErrInvalidFrameType  = errors.New("invalid Frame type")
	ErrInvalidOperation  = errors.New("invalid operation")
	ErrKeyTooLarge       = errors.New("key size exceeds maximum allowed size")
	ErrValueTooLarge     = errors.New("value size exceeds maximum allowed size")
)

var (
	Closing = []byte{byte(TypeControl), byte(OpClosing)}
	Ping    = []byte{byte(TypeControl), byte(OpPing)}
	Pong    = []byte{byte(TypeControl), byte(OpPong)}
)

const headerSize = 4

// Frame size limits for security and performance
const (
	maxFrameSize   = 1024 * 1024 * 10 // 10MB max Frame size
	maxKeySize     = 1024 * 64        // 64KB max key size
	maxValueSize   = 1024 * 1024 * 10 // 10MB max value size
	maxPayloadSize = maxKeySize + maxValueSize
)

type ParserState uint8

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
	state     ParserState
	available int
	Frame     Frame
}

func NewParser() *Parser {
	return &Parser{
		state: StateInit,
	}
}

func (p *Parser) Parse(buf []byte) (int, error) {
	// Validate input buffer
	if len(buf) > maxFrameSize {
		return 0, ErrFrameTooLarge
	}
	readN := 0
outerLoop:
	for {
		if readN >= len(buf) {
			break
		}
		current := buf[readN:]
		switch p.state {
		case StateInit:
			frameType := Type(current[0])
			if frameType < TypeDefault || frameType > TypeError {
				return 0, ErrInvalidFrameType
			}
			readN++
			p.Frame = NewFrame(frameType)
			switch frameType {
			case TypePayload, TypeError:
				p.state = StateIdParsing
			default:
				p.state = StateOpParsing
			}
		case StateOpParsing:
			op := Op(current[0])
			if op > OpClosing {
				return 0, ErrInvalidOperation
			}
			p.Frame.SetOp(op)
			readN++
			switch op {
			case OpDel, OpGet, OpSet:
				p.state = StateIdParsing
			case OpAuth:
				p.state = StatePayloadLenParsing
			default:
				p.state = StateDone
			}
		case StateIdParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			id := binary.LittleEndian.Uint32(current[:headerSize])
			readN += headerSize
			p.Frame.SetID(id)
			switch p.Frame.(type) {
			case *Error, *Payload:
				p.state = StatePayloadLenParsing
			default:
				p.state = StateKeyLenParsing
			}
		case StateKeyLenParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			KeyLen := binary.LittleEndian.Uint32(current[:headerSize])
			if KeyLen > maxKeySize {
				return 0, ErrKeyTooLarge
			}
			p.Frame.SetKeyLength(KeyLen)
			p.state = StatePayloadLenParsing
			readN += headerSize
		case StatePayloadLenParsing:
			if len(current) < headerSize {
				return -1, nil
			}
			size := binary.LittleEndian.Uint32(current[:headerSize])
			switch q := p.Frame.(type) {
			case *Query:
				if q.KeyLen > size {
					return 0, ErrMalformed
				}
				q.Buffer = make([]byte, size)
			case *Payload:
				q.Buffer = make([]byte, size)
			case *Error:
				q.Buffer = make([]byte, size)
			default:
			}
			if int(size) > maxPayloadSize {
				return 0, ErrInvalidPayloadLen
			}
			if size == 0 {
				p.state = StateDone
			} else {
				p.state = StatePayloadParsing
			}
			readN += headerSize
			p.available = int(size)
		case StatePayloadParsing:
			rem := min(len(current), int(p.available))
			n, _ := p.Frame.Write(current[:rem])
			p.available -= n
			readN += n
			if p.available == 0 {
				p.state = StateDone
			}
		case StateDone:
			break outerLoop
		default:
			panic("something went really wrong... (invalid state)")
		}

	}
	return readN, nil
}

func (p *Parser) Done() bool {
	return p.state == StateDone
}

func (p *Parser) Reset() {
	p.state = StateInit
	p.available = 0
}
