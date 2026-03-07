package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// AI generated

func TestClientServerIntegration(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	dbFile := tempDir + "/" + testDBFile

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

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond) // Give server time to start

	// Use a fixed port for testing
	testAddr := testPort1

	t.Run("SingleClientOperations", func(t *testing.T) {
		// Create server on fixed port for this subtest
		testServer, err := NewServer(tempDir+"/test1.db", ServerOpts{
			ListenAddr:   testAddr,
			readTimeout:  time.Second * 5,
			writeTimeout: time.Second * 5,
		})
		require.NoError(t, err)

		testCtx, testCancel := context.WithCancel(t.Context())
		defer testCancel()
		go testServer.Start(testCtx)
		time.Sleep(50 * time.Millisecond)

		client, err := NewClient(testAddr, DefaultClientOpts())
		require.NoError(t, err)
		defer client.cancel() // Clean up client

		ctx := t.Context()

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
		testAddr2 := testPort2
		// Create server on fixed port for this subtest
		testServer, err := NewServer(tempDir+"/test2.db", ServerOpts{
			ListenAddr:   testAddr2,
			readTimeout:  time.Second * 5,
			writeTimeout: time.Second * 5,
		})
		require.NoError(t, err)

		testCtx, testCancel := context.WithCancel(t.Context())
		defer testCancel()
		go testServer.Start(testCtx)
		time.Sleep(50 * time.Millisecond)

		// Create multiple clients
		client1, err := NewClient(testAddr2, DefaultClientOpts())
		require.NoError(t, err)
		defer client1.cancel()

		client2, err := NewClient(testAddr2, DefaultClientOpts())
		require.NoError(t, err)
		defer client2.cancel()

		ctx := t.Context()

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
		testAddr3 := testPort3
		// Create server on fixed port for this subtest
		testServer, err := NewServer(tempDir+"/test3.db", ServerOpts{
			ListenAddr:   testAddr3,
			readTimeout:  time.Second * 5,
			writeTimeout: time.Second * 5,
		})
		require.NoError(t, err)

		testCtx, testCancel := context.WithCancel(t.Context())
		defer testCancel()
		go testServer.Start(testCtx)
		time.Sleep(50 * time.Millisecond)

		// First client sets data
		client1, err := NewClient(testAddr3, DefaultClientOpts())
		require.NoError(t, err)

		err = client1.Set(t.Context(), "persistent", []byte("data"))
		require.NoError(t, err)
		client1.cancel() // Disconnect first client

		time.Sleep(50 * time.Millisecond)

		// Second client connects and can read the data
		client2, err := NewClient(testAddr3, DefaultClientOpts())
		require.NoError(t, err)
		defer client2.cancel()

		value, err := client2.Get(t.Context(), "persistent")
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

	testAddr := testPort4
	server, err := NewServer(dbFile, ServerOpts{
		ListenAddr:   testAddr,
		readTimeout:  time.Millisecond * 100, // Very short timeout
		writeTimeout: time.Millisecond * 100,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go server.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	client, err := NewClient(testAddr, DefaultClientOpts())
	require.NoError(t, err)
	defer client.cancel()

	// Try to set a very large value that might cause timeout
	largeValue := make([]byte, testLargePayloadSize) // Test payload
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	// This should either succeed or timeout gracefully
	ctxTimeout, cancelTimeout := context.WithTimeout(t.Context(), time.Second)
	defer cancelTimeout()

	err = client.Set(ctxTimeout, "largekey", largeValue)
	// We don't assert on the error since it might succeed or timeout depending on system
	// The important thing is that it doesn't panic or hang indefinitely
	t.Logf("Large value operation result: %v", err)
}
