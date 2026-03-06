package main

import "time"

// Database Configuration - can be overridden via ldflags
var (
	defaultDBFile = "default.db" // Default database file path
)

// Server Configuration - can be overridden via ldflags
var (
	defaultServerAddr   = ":4040"
	defaultReadTimeout  = time.Second * 30 // Default read timeout
	defaultWriteTimeout = time.Second * 30 // Default write timeout
	defaultBufferSize   = 1024             // Default buffer size in bytes
	defaultChannelSize  = 32               // Default channel queue size
)

// Client Configuration - can be overridden via ldflags
var (
	defaultClientReadTimeout  = time.Minute      // Default client read timeout
	defaultClientWriteTimeout = time.Minute      // Default client write timeout
	defaultClientBufferSize   = 1024             // Default client buffer size
	defaultClientChannelSize  = 32               // Default client channel size
	clientPingInterval        = time.Second * 30 // Client ping interval
)

// Store Configuration
const (
	storeCompactionInterval = time.Minute * 5 // Store compaction interval
	storeChannelSize        = 32              // Store transaction channel size
)

// Test Configuration
const (
	testPort1            = ":9999"     // Test server port 1
	testPort2            = ":9998"     // Test server port 2
	testPort3            = ":9997"     // Test server port 3
	testPort4            = ":9996"     // Test server port 4
	testDBFile           = "test.db"   // Test database file name
	testLargePayloadSize = 1024 * 1024 // 1MB test payload
)
