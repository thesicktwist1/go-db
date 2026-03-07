package main

import "time"

// Database Configuration
const (
	defaultDBFile = "default.db" // Default database file path
)

// Server Configuration
const (
	defaultServerAddr   = ":4040"
	defaultReadTimeout  = time.Second * 15 // Default read timeout
	defaultWriteTimeout = time.Second * 15 // Default write timeout
	defaultBufferSize   = 1024             // Default buffer size in bytes
	defaultChannelSize  = 32               // Default channel queue size
)

// Client Configuration
const (
	clientPingInterval = time.Second * 5 // Client ping interval
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

// Use client defaults if values are not set
func clientOptsDefaults(opts *ClientOpts) {
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = defaultReadTimeout
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = defaultWriteTimeout
	}
	if opts.BufferSize <= 0 {
		opts.BufferSize = defaultBufferSize
	}
	if opts.ChannelSize <= 0 {
		opts.ChannelSize = defaultChannelSize
	}
}

// Use server defaults if values are not set
func serverOptsDefaults(opts *ServerOpts) {
	if opts.readTimeout == 0 {
		opts.readTimeout = defaultReadTimeout
	}
	if opts.writeTimeout == 0 {
		opts.writeTimeout = defaultReadTimeout
	}
	if opts.bufferSize <= 0 {
		opts.bufferSize = defaultBufferSize
	}
	if opts.channelSize <= 0 {
		opts.channelSize = defaultChannelSize
	}
}

// DefaultClientOpts returns the default client options.
func DefaultClientOpts() ClientOpts {
	return ClientOpts{
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		BufferSize:   defaultBufferSize,
		ChannelSize:  defaultChannelSize,
	}
}
