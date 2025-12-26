.PHONY: all build install test clean fmt vet lint

# Binary name
BINARY_NAME=ptyee

# Build directory
BUILD_DIR=build

all: build

# Build the CLI tool
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ptyee

# Install the CLI tool
install:
	go install ./cmd/ptyee

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Run the built binary
run: build
	$(BUILD_DIR)/$(BINARY_NAME)