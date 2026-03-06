# Go-DB

Redis-style in-memory database in Go.

## Features
- GET/SET/DELETE operations
- In-memory storage with persistence (WAL)
- TCP client-server architecture
- Concurrent request handling

## Quick Start
```bash
git clone https://github.com/thesicktwist1/go-db.git
cd go-db && go run .
```

Server runs on `:4040`.

## Usage
```go
client, _ := NewClient(":4040")
client.Set(ctx, "key", []byte("value"))
value, _ := client.Get(ctx, "key")
client.Delete(ctx, "key")
```

## Commands
```bash
make run    # Start server
make test   # Run tests
make build  # Build binary
make clean  # Clean up
make fmt    # Format code
```

## Architecture
Client ↔ Server ↔ Store (with WAL persistence)</content>
<parameter name="filePath">/home/thesicktwist1/workspace/github.com/thesicktwist1/go-db/README.md
