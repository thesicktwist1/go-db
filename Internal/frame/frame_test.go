package frame

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Query(t *testing.T) {
	query := &Query{
		Buffer: make([]byte, 3),
	}
	expectedBytes := fmt.Appendf(
		[]byte{2, 2, 21, 0, 0, 0, 3, 0, 0, 0, 3, 0, 0, 0},
		"%s", "key",
	)
	query.SetID(21)
	require.Equal(t, uint32(21), query.ID)

	query.SetKeyLength(3)
	require.Equal(t, uint32(3), query.KeyLen)

	query.SetOp(OpDel)
	require.Equal(t, OpDel, query.Op)

	n, _ := query.Write([]byte("key"))
	require.Equal(t, 3, n)
	require.Equal(t, []byte("key"), query.Buffer)

	require.Equal(t, expectedBytes, query.Bytes())

}

func Test_Control(t *testing.T) {
	control := &Control{}

	control.SetOp(OpPing)
	require.Equal(t, OpPing, control.Op)

	expectedBytes := []byte{1, 5}

	require.Equal(t, expectedBytes, control.Bytes())
	control.SetID(21)          // shouldn't do anything
	control.SetKeyLength(32)   // shouldn't do anything
	control.Write([]byte("d")) // shouldn't do anything
	require.Equal(t, expectedBytes, control.Bytes())

}

func Test_Error(t *testing.T) {
	errFrame := &Error{
		Buffer: make([]byte, 3),
	}

	errFrame.SetID(10)
	require.Equal(t, uint32(10), errFrame.Id)

	n, _ := errFrame.Write([]byte("err"))
	require.Equal(t, 3, n)

	expectedBytes := fmt.Appendf(
		[]byte{4, 10, 0, 0, 0, 3, 0, 0, 0}, "%s", "err")
	require.Equal(t, expectedBytes, errFrame.Bytes())
	errFrame.SetKeyLength(3) // shouldn't do anything
	errFrame.SetOp(OpPing)   // shouldn't do anything
	require.Equal(t, expectedBytes, errFrame.Bytes())
}

func Test_Payload(t *testing.T) {
	payload := &Payload{
		Buffer: make([]byte, 3),
	}

	payload.SetID(10)
	require.Equal(t, uint32(10), payload.Id)

	n, _ := payload.Write([]byte("dat"))
	require.Equal(t, 3, n)

	expectedBytes := fmt.Appendf(
		[]byte{3, 10, 0, 0, 0, 3, 0, 0, 0}, "%s", "dat")
	require.Equal(t, expectedBytes, payload.Bytes())
	payload.SetKeyLength(3) // shouldn't do anything
	payload.SetOp(OpPing)   // shouldn't do anything
	require.Equal(t, expectedBytes, payload.Bytes())
}
