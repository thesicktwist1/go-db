# Go-DB

Redis-style in-memory database in Go with persistence and concurrent client support.

## Features

### Core Operations
- **GET** - Retrieve value by key
- **SET** - Store key-value pairs
- **DELETE** - Remove keys from database
- **PING/PONG** - Connection health checks

### Technical Features
- **In-Memory Storage** - Fast access with O(1) hash map operations
- **Persistence** - Append only logging for durability
- **TCP Communication** - Binary protocol over TCP for efficiency
- **Concurrent Clients** - Goroutine-based handling of multiple connections
- **Configurable Timeouts** - Customizable read/write deadlines
- **Graceful Shutdown** - Proper cleanup on termination signals
- **Structured Logging** - slog for comprehensive operations logging

## Requirements
- Go 1.19+
- No external dependencies (except testify for tests)

## Quick Start

### Installation
```bash
git clone https://github.com/thesicktwist1/go-db.git
cd go-db
go mod tidy
```

### Running Server
```bash
# Using make
make run

# Or directly
go run .
```

Server listens on `:4040` by default.

## Configuration

### Server Options
```go
server, _ := NewServer("database.db", ServerOpts{
    ListenAddr:   ":4040",
    readTimeout:  time.Minute,
    writeTimeout: time.Minute,
})
server.Start(ctx)
```

### Client Options
```go
// Default options (1 minute timeouts)
client, _ := NewClient(":4040", DefaultClientOpts())

// Custom timeouts
client, _ := NewClient(":4040", ClientOpts{
    ReadTimeout:  time.Second * 30,
    WriteTimeout: time.Second * 30,
})
```

## Usage Examples

### Basic Operations
```go
import (
    "context"
    "log"
)

client, err := NewClient(":4040", DefaultClientOpts())
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()

// Set a value
err = client.Set(ctx, "username", []byte("john"))

// Get a value
value, err := client.Get(ctx, "username")
if err != nil {
    log.Fatal(err)
}

// Delete a key
err = client.Delete(ctx, "username")
```

### With Context Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

value, err := client.Get(ctx, "key")
```

## Architecture

```
┌─────────────────────────────────────────┐
│         TCP Clients (Multiple)          │
└────┬────────────────────────────┬───────┘
     │                            │
     └────────────┬───────────────┘
                  │ (TCP Connections)
         ┌────────▼────────┐
         │   Server        │
         │ (Peer Manager)  │
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │   Store         │
         │ (In-Memory)     │
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │       Log       │
         │ (Persistence)   │
         └─────────────────┘
```

## Data Storage

- **Key**: String
- **Value**: []byte (binary data)
- **Storage**: Hash map in memory
- **Persistence**: Write-ahead log to disk

## Commands

```bash
make run      # Start server
make test     # Run unit and integration tests
make build    # Build binary
make clean    # Clean build artifacts
make fmt      # Format code
make help     # Show available commands
```

## Testing

Comprehensive test coverage:
- **Unit tests** for individual components (frame parsing, storage operations)
- **Integration tests** for client-server communication
- **Concurrent access** tests with multiple clients
- **Persistence** tests across connections
- **Timeout** tests with large payloads

Run tests:
```bash
make test
# or
go test -v ./...
```

## File Structure

```
├── main.go              # Server entry point
├── client.go            # TCP client implementation
├── server.go            # TCP server implementation
├── store.go             # In-memory storage engine
├── peer.go              # Connection handler
├── pid.go               # Process ID utilities
├── integration_test.go  # Client-server integration tests
├── store_test.go        # Storage operation tests
├── Internal/
│   ├── frame/           # Protocol implementation
│   │   ├── op.go        # Operation types
│   │   ├── parser.go    # Frame parsing
│   │   └── op_test.go   # Operation tests
│   └── logs/            # Persistence layer
│       ├── logs.go      # Append-Only implementation
│       └── logs_test.go # Log tests
├── makefile             # Build commands
├── go.mod               # Go module
└── README.md            # This file
```

## How It Works

1. **Server Start** - Listens on configured address, starts background store processor
2. **Client Connect** - Client connects via TCP, starts read/write loops
3. **Request** - Client sends operation (GET/SET/DELETE) in binary frame format
4. **Processing** - Server queues transaction, store executes it
5. **Response** - Result sent back to client
6. **Persistence** - SET/DELETE operations logged to file
7. **Cleanup** - Graceful shutdown on signals (SIGTERM, SIGINT)

## Performance Notes

- **Latency**: Sub-millisecond for in-memory operations
- **Throughput**: Limited by network and concurrent request handling
- **Memory**: In-memory storage grows with data size

## Limitations

- Single-machine only (no replication)
- No clustering support
- No advanced data types (lists, sets, hashes)
- No pub/sub functionality
- No authentication/ACL</content>

