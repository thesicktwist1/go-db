package main

import (
	"bytes"
	"os"
	"testing"
	"thesicktwist1/go-db/Internal/frame"

	"github.com/stretchr/testify/require"
)

func Test_processTask(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	store, err := NewStore("test.db")
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
			transaction: transaction{nil, nil, frame.Frame{
				Type:   frame.TypeQuery,
				Op:     frame.OpGet,
				KeyLen: 3,
				Buffer: *bytes.NewBuffer([]byte("ke")),
			}},
			wantErr: true,
			errType: ErrNotExists,
		},
		{
			name: "Operation Auth (invalid operation)",
			transaction: transaction{nil, nil, frame.Frame{
				Type: frame.TypeQuery,
				Op:   frame.OpAuth,
			}},
			wantErr: true,
			errType: frame.ErrMalformed,
		},
		{
			name: "Operation Set (should be valid)",
			transaction: transaction{nil, nil, frame.Frame{
				Type:   frame.TypeQuery,
				Op:     frame.OpSet,
				KeyLen: 3,
				Buffer: *bytes.NewBuffer([]byte("keynewvalue")),
			}},
		},
		{
			name: "Operation Delete (should be valid)",
			transaction: transaction{nil, nil, frame.Frame{
				Type:   frame.TypeQuery,
				Op:     frame.OpDel,
				KeyLen: 3,
				Buffer: *bytes.NewBuffer([]byte("key")),
			}},
		},
		{
			name: "Operation Delete (invalid key)",
			transaction: transaction{nil, nil, frame.Frame{
				Type:   frame.TypeQuery,
				Op:     frame.OpDel,
				KeyLen: 3,
				Buffer: *bytes.NewBuffer([]byte("k")),
			}},
			wantErr: true,
			errType: ErrNotExists,
		},
		{
			name: "Operation Get (should be valid)",
			transaction: transaction{nil, nil, frame.Frame{
				Type:   frame.TypeQuery,
				Op:     frame.OpGet,
				KeyLen: 7,
				Buffer: *bytes.NewBuffer([]byte("example")),
			}},
			expected: []byte("key"),
		},
	}
	for _, tc := range tests {
		got, err := store.processTask(tc.transaction)
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
