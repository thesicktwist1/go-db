package frame

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ChunkParse(t *testing.T) {
	parser := NewParser()

	// parsing a control request
	buf := []byte{1}

	readN, err := parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.State, StateOpParsing)
	require.Equal(t, parser.Frame.Type, TypeControl)

	buf = []byte{5}

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.State, StateDone)
	require.Equal(t, parser.Frame.Op, OpPing)

	parser.Reset()

	// parsing a query request (set)
	buf = []byte{2}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.State, StateOpParsing)

	buf = []byte{3}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.State, StateIdParsing)

	buf = []byte{0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, -1, readN)
	require.Equal(t, parser.State, StateIdParsing)

	buf = []byte{21, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.State, StateKeyLenParsing)

	buf = []byte{3, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.State, StatePayloadLenParsing)

	buf = []byte{8, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.State, StatePayloadParsing)

	buf = []byte("keyva")

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 5, readN)
	require.Equal(t, parser.State, StatePayloadParsing)

	buf = []byte("lue more data")

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 3, readN)
	require.Equal(t, parser.State, StateDone)
}

func Test_Parse(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name          string
		input         []byte
		expected      Frame
		expectedReadN int
		wantErr       bool
		errType       error
	}{
		{
			name:          "Delete Query Frame",
			input:         Query(OpDel, 10, "key", []byte("example")),
			expectedReadN: 24,
			expected: Frame{
				Type:      TypeQuery,
				Id:        10,
				Op:        OpDel,
				Buffer:    *bytes.NewBuffer([]byte("keyexample")),
				KeyLen:    3,
				available: 0,
			},
		},
		{
			name:          "Payload Frame",
			input:         Payload(21, []byte("example")),
			expectedReadN: 16,
			expected: Frame{
				Type:      TypePayload,
				Id:        21,
				Op:        0,
				Buffer:    *bytes.NewBuffer([]byte("example")),
				KeyLen:    0,
				available: 0,
			},
		},
		{
			name:          "Empty Payload Frame",
			input:         Payload(21, nil),
			expectedReadN: 9,
			expected: Frame{
				Type:      TypePayload,
				Id:        21,
				Op:        0,
				Buffer:    *bytes.NewBuffer([]byte{}),
				KeyLen:    0,
				available: 0,
			},
		},
		{
			name:          "Ping Frame",
			input:         []byte{1, 5},
			expectedReadN: 2,
			expected: Frame{
				Type:      TypeControl,
				Id:        21,
				Op:        OpPing,
				Buffer:    *bytes.NewBuffer([]byte{}),
				KeyLen:    0,
				available: 0,
			},
		},
		{
			name:    "Malformed Frame (key length is bigger than size(key and payload))",
			input:   []byte{2, 1, 42, 0, 0, 0, 20, 0, 0, 0, 4, 0, 0, 0},
			wantErr: true,
			errType: ErrMalformed,
		},
	}
	for _, tc := range tests {
		n, err := parser.Parse(tc.input)
		if tc.wantErr {
			require.Error(t, err)
			require.ErrorIs(t, err, tc.errType)
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.expectedReadN, n)
			require.Equal(t, tc.expected, parser.Frame)
			require.Equal(t, StateDone, parser.State)
		}
		parser.Reset()
	}
}

func Test_Reset(t *testing.T) {
	parser := &Parser{
		State: StateDone,
		Frame: Frame{
			Type:   TypeControl,
			Op:     OpPing,
			Buffer: *bytes.NewBuffer([]byte("example")),
			KeyLen: 7,
		},
	}

	parser.Reset()

	expected := Frame{
		Type:      TypeDefault,
		Op:        OpDefault,
		Buffer:    *bytes.NewBuffer([]byte{}),
		available: 0,
		KeyLen:    0,
	}

	require.Equal(t, expected, parser.Frame)

}

func Test_NewPayload(t *testing.T) {
	expected := fmt.Appendf([]byte{3, 10, 0, 0, 0, 7, 0, 0, 0}, "%s", "example")
	got := Payload(10, []byte("example"))

	require.Equal(t, expected, got)
}

func Test_NewQuery(t *testing.T) {
	p := "keyexample"
	expected := fmt.Appendf([]byte{2, 1, 10, 0, 0, 0, 3, 0, 0, 0, 10, 0, 0, 0}, "%s", p)
	got := Query(OpGet, 10, "key", []byte("example"))

	require.Equal(t, expected, got)
}

func Test_NewError(t *testing.T) {
	expected := fmt.Appendf([]byte{4, 10, 0, 0, 0, 5, 0, 0, 0}, "%s", "error")

	got := Error(10, errors.New("error"))
	require.Equal(t, expected, got)
}

func Test_FrameSizeValidation(t *testing.T) {
	parser := NewParser()

	// Test oversized frame
	oversizedFrame := make([]byte, maxFrameSize+1)
	_, err := parser.Parse(oversizedFrame)
	require.Error(t, err)
	require.Equal(t, ErrFrameTooLarge, err)

	parser.Reset()

	// Test invalid frame type
	invalidTypeFrame := []byte{255} // Invalid type
	_, err = parser.Parse(invalidTypeFrame)
	require.Error(t, err)
	require.Equal(t, ErrInvalidFrameType, err)

	parser.Reset()

	// Test invalid operation
	invalidOpFrame := []byte{2, 255} // Valid type, invalid op
	_, err = parser.Parse(invalidOpFrame)
	require.Error(t, err)
	require.Equal(t, ErrInvalidOperation, err)
}
