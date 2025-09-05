.PHONY: build test clean install run lint help

# Binary name
BINARY_NAME=strategic-claude-basic-cli
BUILD_DIR=bin

# Version information
VERSION ?= 0.1.0
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Build the application
build:
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/$(BINARY_NAME)

# Test the application
test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f coverage.out

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary
install: build
	@if [ -z "$(GOPATH)" ]; then \
		echo "Installing to $(HOME)/go/bin/$(BINARY_NAME)"; \
		mkdir -p $(HOME)/go/bin; \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/go/bin/$(BINARY_NAME); \
	else \
		echo "Installing to $(GOPATH)/bin/$(BINARY_NAME)"; \
		mkdir -p $(GOPATH)/bin; \
		cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME); \
	fi

# Run the application
run:
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/$(BINARY_NAME)
	./$(BUILD_DIR)/$(BINARY_NAME)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build with dependencies
build: $(BUILD_DIR) deps

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  install      - Install the binary to GOPATH/bin"
	@echo "  run          - Build and run the application"
	@echo "  help         - Show this help message"
