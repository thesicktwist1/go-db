.PHONY: test run build clean fmt help

# Build the binary
build:
	go build -o bin/go-db .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	go clean
	rm -rf bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  run       - Run the application"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format code"
	@echo "  clean     - Clean build artifacts"
	@echo "  help      - Show this help message"

