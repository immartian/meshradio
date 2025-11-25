# MeshRadio Makefile

.PHONY: all build clean run install test help build-audio build-full check-audio install-audio-deps

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

# Build with audio support (auto-detect)
build-audio:
	@echo "Building MeshRadio with audio support..."
	@if pkg-config --exists portaudio-2.0 2>/dev/null && pkg-config --exists opus 2>/dev/null; then \
		echo "✅ Found PortAudio and Opus - building with full audio support"; \
		$(GO) build -tags "portaudio opus" $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR); \
	elif pkg-config --exists portaudio-2.0 2>/dev/null; then \
		echo "✅ Found PortAudio - building with real I/O (no compression)"; \
		$(GO) build -tags "portaudio" $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR); \
	elif pkg-config --exists opus 2>/dev/null; then \
		echo "✅ Found Opus - building with compression (simulated I/O)"; \
		$(GO) build -tags "opus" $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR); \
	else \
		echo "⚠️  No audio libraries found - building with simulated audio"; \
		echo "   Run 'make install-audio-deps' to enable real audio"; \
		$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR); \
	fi
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Force build with full audio support
build-full:
	@echo "Building with full audio support..."
	$(GO) build -tags "portaudio opus" $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Check audio library status
check-audio:
	@echo "Checking audio dependencies..."
	@bash scripts/check-deps.sh

# Install audio dependencies
install-audio-deps:
	@echo "Installing audio dependencies..."
	@bash scripts/install-audio-deps.sh

# Install Go audio bindings
install-go-audio:
	@echo "Installing Go audio bindings..."
	$(GO) get github.com/gordonklaus/portaudio
	$(GO) get gopkg.in/hraban/opus.v2
	@echo "Go bindings installed!"

# Show help
help:
	@echo "MeshRadio Makefile"
	@echo ""
	@echo "Basic Commands:"
	@echo "  make build              - Build with simulated audio"
	@echo "  make build-audio        - Build with audio (auto-detect libraries)"
	@echo "  make build-full         - Build with full audio support"
	@echo "  make run                - Build and run"
	@echo "  make clean              - Remove build artifacts"
	@echo ""
	@echo "Audio Setup:"
	@echo "  make check-audio        - Check audio library status"
	@echo "  make install-audio-deps - Install PortAudio and Opus"
	@echo "  make install-go-audio   - Install Go audio bindings"
	@echo ""
	@echo "Development:"
	@echo "  make test               - Run tests"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Lint code"
	@echo "  make dev                - Build with race detector"
	@echo ""
	@echo "Installation:"
	@echo "  make install            - Install to /usr/local/bin"
	@echo "  make deps               - Install Go dependencies"
