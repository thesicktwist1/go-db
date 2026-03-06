package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClientServerIntegration(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	dbFile := tempDir + "/test.db"

	// Change to temp directory
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create and start server
	server, err := NewServer(dbFile, ServerOpts{
		ListenAddr:   ":0", // Use random available port
		readTimeout:  time.Second * 5,
		writeTimeout: time.Second * 5,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond) // Give server time to start

	// Get the actual port the server is listening on
	addr := server.GetAddr()

	t.Run("SingleClientOperations", func(t *testing.T) {
		client, err := NewClient(addr)
		require.NoError(t, err)
		defer client.cancel() // Clean up client

		ctx := context.Background()

		// Test SET operation
		err = client.Set(ctx, "testkey", []byte("testvalue"))
		require.NoError(t, err)

		// Test GET operation
		value, err := client.Get(ctx, "testkey")
		require.NoError(t, err)
		require.Equal(t, "testvalue", string(value))

		// Test DELETE operation
		err = client.Delete(ctx, "testkey")
		require.NoError(t, err)

		// Verify key is deleted
		_, err = client.Get(ctx, "testkey")
		require.Error(t, err) // Should return error for non-existent key
	})

	t.Run("MultipleClientsConcurrent", func(t *testing.T) {
		// Create multiple clients
		client1, err := NewClient(addr)
		require.NoError(t, err)
		defer client1.cancel()

		client2, err := NewClient(addr)
		require.NoError(t, err)
		defer client2.cancel()

		ctx := context.Background()

		// Client 1 sets a value
		err = client1.Set(ctx, "sharedkey", []byte("client1value"))
		require.NoError(t, err)

		// Client 2 can read it
		value, err := client2.Get(ctx, "sharedkey")
		require.NoError(t, err)
		require.Equal(t, "client1value", string(value))

		// Client 2 updates the value
		err = client2.Set(ctx, "sharedkey", []byte("client2value"))
		require.NoError(t, err)

		// Client 1 can read the updated value
		value, err = client1.Get(ctx, "sharedkey")
		require.NoError(t, err)
		require.Equal(t, "client2value", string(value))
	})

	t.Run("PersistenceAcrossConnections", func(t *testing.T) {
		// First client sets data
		client1, err := NewClient(addr)
		require.NoError(t, err)

		err = client1.Set(context.Background(), "persistent", []byte("data"))
		require.NoError(t, err)
		client1.cancel() // Disconnect first client

		// Second client connects and can read the data
		client2, err := NewClient(addr)
		require.NoError(t, err)
		defer client2.cancel()

		value, err := client2.Get(context.Background(), "persistent")
		require.NoError(t, err)
		require.Equal(t, "data", string(value))
	})
}

func TestClientTimeout(t *testing.T) {
	tempDir := t.TempDir()
	dbFile := tempDir + "/test.db"

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	server, err := NewServer(dbFile, ServerOpts{
		ListenAddr:   ":0",
		readTimeout:  time.Millisecond * 100, // Very short timeout
		writeTimeout: time.Millisecond * 100,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go server.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	client, err := NewClient(server.GetAddr())
	require.NoError(t, err)
	defer client.cancel()

	// Try to set a very large value that might cause timeout
	largeValue := make([]byte, 1024*1024) // 1MB
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	// This should either succeed or timeout gracefully
	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second)
	defer cancelTimeout()

	err = client.Set(ctxTimeout, "largekey", largeValue)
	// We don't assert on the error since it might succeed or timeout depending on system
	// The important thing is that it doesn't panic or hang indefinitely
	t.Logf("Large value operation result: %v", err)
}
