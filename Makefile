# MeshRadio Makefile

.PHONY: all build clean run install test help

# Variables
BINARY_NAME=meshradio
GO=go
GOFLAGS=-v
BUILD_DIR=.
CMD_DIR=./cmd/meshradio

all: build

# Build the binary
build:
	@echo "Building MeshRadio..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build
	@echo "Starting MeshRadio..."
	./$(BINARY_NAME)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	$(GO) clean

# Install to system
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed!"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Development build with race detector
dev:
	@echo "Building development version..."
	$(GO) build -race $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Show help
help:
	@echo "MeshRadio Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build    - Build the binary"
	@echo "  make run      - Build and run"
	@echo "  make install  - Install to system"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Install dependencies"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Lint code"
	@echo "  make dev      - Build with race detector"
	@echo "  make help     - Show this help"
