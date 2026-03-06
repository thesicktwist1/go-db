package logs

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_copyFile(t *testing.T) {
	mem := make(map[string][]byte)
	tempDir := t.TempDir()
	path := tempDir + "sync_test"
	log, err := New(path, mem)
	require.NoError(t, err)

	line := formatLine([]byte("keyexample"), 3)

	_, err = log.file.Write(line)
	require.NoError(t, err)

	copy, err := log.copyFile()
	require.NoError(t, err)

	copy, err = os.OpenFile(copy.Name(), os.O_APPEND|os.O_RDWR, 0644)
	require.NoError(t, err)

	_, err = copy.Seek(0, 0)
	require.NoError(t, err)

	got := make([]byte, len(line))
	n, err := copy.Read(got)
	require.NoError(t, err)
	require.Equal(t, len(line), n)

	err = os.Remove(copy.Name())
	require.NoError(t, err)
}

func Test_sync(t *testing.T) {
	mem := make(map[string][]byte)
	tempDir := t.TempDir()
	path := tempDir + "sync_test"
	log, err := New(path, mem)
	require.NoError(t, err)

	line := formatLine([]byte("keyexample"), 3)

	select {
	case log.writesCh <- line:
	default:
		t.Fatal("something went wrong")
	}

	err = log.sync()
	require.NoError(t, err)

	got := make([]byte, len(line))

	_, err = log.file.Seek(0, 0)
	require.NoError(t, err)

	n, _ := log.file.Read(got)
	require.Equal(t, len(line), n)
	require.Equal(t, line, got)
}

func Test_read(t *testing.T) {
	mem := make(map[string][]byte)
	file, err := os.CreateTemp(os.TempDir(), "read_test.db")
	require.NoError(t, err)
	line := formatLine([]byte("keyexample"), 3)

	_, err = file.Write(line)
	require.NoError(t, err)

	n, err := read(file, mem)
	require.NoError(t, err)
	require.Equal(t, len(line), n)

	expected := []byte("example")
	got, exists := mem["key"]
	require.True(t, exists)
	require.Equal(t, expected, got)
}

func Test_formatLine(t *testing.T) {
	expectedPayload := []byte("key:example")
	keyLen := 3
	line := formatLine(expectedPayload, keyLen)
	line = bytes.TrimRight(line, newLine)

	expected := uint32(957925644)
	got := binary.LittleEndian.Uint32(line[len(line)-headerLen:])

	require.Equal(t, expected, got, "%d != %d", expected, got)

	expectedKeyLen := uint32(3)
	gotKeyLen := binary.LittleEndian.Uint32(line[:headerLen])
	require.Equalf(t, expectedKeyLen, gotKeyLen, "%d != %d", expectedKeyLen, gotKeyLen)

	gotPayload := line[headerLen : len(line)-headerLen]
	require.Equal(t, expectedPayload, gotPayload)
}

func Test_readLine(t *testing.T) {
	expectedPayload := []byte("keyexample")
	keyLen := 3
	line := formatLine(expectedPayload, keyLen)
	line = bytes.TrimRight(line, newLine)
	expectedN := len(line)
	mem := make(map[string][]byte)
	gotN, err := readLine(line, mem)
	require.Equal(t, expectedN, gotN)
	require.NoError(t, err)

	expected := []byte("example")
	got, exist := mem["key"]
	require.True(t, exist)
	require.Equal(t, expected, got)
}
