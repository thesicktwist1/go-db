package main

import (
	"os"
	"testing"
	"thesicktwist1/go-db/Internal/frame"

	"github.com/stretchr/testify/require"
)

func Test_executeTransaction(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	store, err := NewStore(testDBFile)
	require.NoError(t, err)
	store.mem["key"] = []byte("example")
	store.mem["example"] = []byte("key")

	tests := []struct {
		name        string
		transaction transaction
		expected    []byte
		wantErr     bool
		errType     error
	}{
		{
			name: "Operation Get (invalid key)",
			transaction: transaction{nil, frame.Query{
				ID:     21,
				Op:     frame.OpGet,
				KeyLen: 3,
				Buffer: []byte("ke"),
			}, nil},
			wantErr: true,
			errType: ErrNotExists,
		},
		{
			name: "Operation Get (should be valid)",
			transaction: transaction{nil, frame.Query{
				ID:     2,
				Op:     frame.OpGet,
				KeyLen: 3,
				Buffer: []byte("key"),
			}, nil},
			expected: []byte("example"),
		},
		{
			name: "Operation Delete (should be valid)",
			transaction: transaction{nil, frame.Query{
				Op:     frame.OpDel,
				KeyLen: 3,
				Buffer: []byte("key"),
			}, nil},
		},
		{
			name: "Operation Set (should be valid)",
			transaction: transaction{nil, frame.Query{
				Op:     frame.OpSet,
				KeyLen: 3,
				Buffer: []byte("keyexample"),
			}, nil},
		},
	}
	for _, tc := range tests {
		got, err := store.executeTransaction(tc.transaction)
		if tc.wantErr {
			require.Errorf(t, err, tc.name)
			require.ErrorIsf(t, err, tc.errType, tc.name)
			require.Equalf(t, tc.expected, got, tc.name)
		} else {
			require.NoErrorf(t, err, tc.name)
			require.Equalf(t, tc.expected, got, tc.name)
		}
	}
}
