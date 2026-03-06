# Go-DB: A Redis-Style Database

A high-performance, Redis-compatible in-memory database written in Go, designed for simplicity, speed, and reliability.

## 🎯 Project Goal

Go-DB aims to provide a lightweight, Redis-style database that combines the simplicity of Redis with Go's performance and concurrency model. The project focuses on:

- **High Performance**: In-memory operations with efficient TCP-based communication
- **Persistence**: Write-ahead logging for data durability
- **Concurrency**: Goroutine-based architecture for handling multiple clients
- **Simplicity**: Clean, maintainable codebase with minimal dependencies
- **Redis Compatibility**: Familiar GET/SET/DELETE operations with Redis-like protocol

## 🚀 Features

### Core Operations
- **GET** - Retrieve values by key
- **SET** - Store key-value pairs
- **DELETE** - Remove keys from the database
- **PING/PONG** - Connection health checks

### Architecture
- **Client-Server Model**: TCP-based communication between clients and server
- **In-Memory Storage**: Fast access with hash map backend
- **Write-Ahead Logging**: Persistent storage with configurable sync intervals
- **Concurrent Processing**: Goroutine-based request handling
- **Connection Management**: Automatic peer management and cleanup

### Technical Highlights
- **Custom Protocol**: Efficient binary frame-based communication
- **Structured Logging**: Comprehensive logging with slog
- **Graceful Shutdown**: Proper cleanup on termination signals
- **Timeout Management**: Configurable read/write timeouts
- **Error Handling**: Robust error propagation and recovery

## 🏗️ Architecture Overview

```
┌─────────────┐    TCP     ┌─────────────┐
│   Client    │◄──────────►│   Server    │
│             │            │             │
│ • GET/SET   │            │ • Peer Mgmt │
│ • DELETE    │            │ • Frame     │
│ • PING      │            │   Parsing   │
└─────────────┘            └─────────────┘
                                │
                                ▼
                       ┌─────────────┐
                       │    Store    │
                       │             │
                       │ • In-Memory │
                       │ • WAL       │
                       │ • Sync      │
                       └─────────────┘
```

### Components

- **Server**: Main database server handling connections and requests
- **Client**: TCP client for database operations
- **Store**: In-memory storage with persistence layer
- **Frame**: Binary protocol for network communication
- **Logs**: Write-ahead logging system
- **Peer**: Connection abstraction for client management

## 🛠️ Quick Start

### Prerequisites
- Go 1.19 or later

### Installation
```bash
git clone https://github.com/thesicktwist1/go-db.git
cd go-db
go mod tidy
```

### Running the Server
```bash
# Start the database server
make run
# or
go run .
```

The server will start on `:4040` by default.

### Basic Usage Example
```go
// Connect to the database
client, err := NewClient(":4040")
if err != nil {
    log.Fatal(err)
}

// Set a value
err = client.Set(context.Background(), "key", []byte("value"))
if err != nil {
    log.Fatal(err)
}

// Get the value
value, err := client.Get(context.Background(), "key")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(value)) // Output: value

// Delete the key
err = client.Delete(context.Background(), "key")
```

## 📁 Project Structure

```
go-db/
├── main.go              # Application entry point
├── client.go            # TCP client implementation
├── server.go            # TCP server implementation
├── store.go             # In-memory storage engine
├── peer.go              # Peer connection management
├── pid.go               # Process ID utilities
├── Internal/
│   ├── frame/           # Binary protocol frames
│   │   ├── op.go        # Operation definitions
│   │   ├── parser.go    # Frame parsing
│   │   └── op_test.go   # Operation tests
│   └── logs/            # Write-ahead logging
│       ├── logs.go      # Log implementation
│       └── logs_test.go # Log tests
├── makefile             # Build automation
├── go.mod               # Go module definition
└── README.md            # This file
```

## 🧪 Testing

Run the test suite:
```bash
make test
# or
go test ./...
```

## 🎯 Design Philosophy

### Simplicity First
- Minimal dependencies (only testify for testing)
- Clean, idiomatic Go code
- Straightforward architecture

### Performance Focus
- In-memory operations for speed
- Efficient binary protocol
- Concurrent request processing

### Reliability
- Persistent storage via WAL
- Graceful error handling
- Connection lifecycle management

## 🔮 Future Enhancements

- [ ] Authentication system
- [ ] Replication support
- [ ] Clustering capabilities
- [ ] Additional data types (lists, sets, hashes)
- [ ] Pub/Sub functionality
- [ ] Snapshotting and backup
- [ ] Metrics and monitoring

## 📝 License

This project is open source. See LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

---

*Built with ❤️ in Go*</content>
<parameter name="filePath">/home/thesicktwist1/workspace/github.com/thesicktwist1/go-db/README.md
