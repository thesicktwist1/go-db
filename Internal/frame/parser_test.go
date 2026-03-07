package frame

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func testDeleteQuery() []byte {
	return fmt.Appendf(
		[]byte{2, 2, 10, 0, 0, 0, 3, 0, 0, 0, 3, 0, 0, 0}, "%s", "key",
	)
}

func Test_ChunkParse(t *testing.T) {
	parser := NewParser()

	// parsing a control request
	buf := []byte{1}

	readN, err := parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.state, StateOpParsing)
	control, ok := parser.Frame.(*Control)
	require.True(t, ok)

	buf = []byte{5}

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.state, StateDone)
	require.Equal(t, OpPing, control.Op)

	parser.Reset()

	// parsing a query frame (set)
	buf = []byte{2}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.state, StateOpParsing)

	buf = []byte{3}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 1, readN)
	require.Equal(t, parser.state, StateIdParsing)

	buf = []byte{21, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, -1, readN)
	require.Equal(t, parser.state, StateIdParsing)

	buf = []byte{21, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.state, StateKeyLenParsing)

	buf = []byte{3, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.state, StatePayloadLenParsing)

	buf = []byte{8, 0, 0, 0}
	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 4, readN)
	require.Equal(t, parser.state, StatePayloadParsing)

	buf = []byte("keyva")

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 5, readN)
	require.Equal(t, parser.state, StatePayloadParsing)

	buf = []byte("lue more data")

	readN, err = parser.Parse(buf)
	require.NoError(t, err)
	require.Equal(t, 3, readN)
	require.Equal(t, parser.state, StateDone)

	expected := &Query{
		Op:     OpSet,
		ID:     21,
		KeyLen: 3,
		Buffer: []byte("keyvalue"),
		offset: 8,
	}
	require.Equal(t, expected, parser.Frame)
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
			input:         testDeleteQuery(),
			expectedReadN: len(testDeleteQuery()),
			expected: &Query{
				Op:     OpDel,
				ID:     10,
				KeyLen: 3,
				Buffer: []byte("key"),
				offset: 3,
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
			require.Equal(t, StateDone, parser.state)
		}
		parser.Reset()
	}
}

func Test_FrameSizeValidation(t *testing.T) {
	parser := NewParser()

	// Test oversized frame
	oversizedFrame := make([]byte, maxPayloadSize)
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
