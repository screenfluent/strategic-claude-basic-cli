.PHONY: build test clean install run lint fmt fmt-check lint-strict pre-commit-check help

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

# Format code (matches pre-commit formatting)
fmt:
	@echo "Formatting Go code with goimports..."
	goimports -w .
	@echo "Tidying Go modules..."
	$(GOMOD) tidy
	@echo "Fixing trailing whitespace..."
	@find . -name "*.go" -exec sed -i 's/[[:space:]]*$$//' {} \;

# Check formatting without making changes
fmt-check:
	@echo "Checking Go formatting..."
	@if [ -n "$$(goimports -l .)" ]; then \
		echo "The following files need formatting:"; \
		goimports -l .; \
		echo "Run 'make fmt' to fix formatting issues."; \
		exit 1; \
	fi
	@echo "Checking Go modules..."
	@if ! $(GOMOD) tidy -diff; then \
		echo "Go modules need tidying. Run 'make fmt' to fix."; \
		exit 1; \
	fi

# Run linter (original target)
lint:
	golangci-lint run

# Run comprehensive linting (matches pre-commit strictness)
lint-strict:
	@echo "Running comprehensive linting checks..."
	@echo "1. Checking Go formatting and imports..."
	@$(MAKE) fmt-check
	@echo "2. Running golangci-lint..."
	golangci-lint run
	@echo "3. Checking for merge conflicts..."
	@if find . -name "*.go" -exec grep -l "<<<<<<< HEAD\|=======" {} \; | head -1 | grep -q .; then \
		echo "Merge conflict markers found in Go files!"; \
		exit 1; \
	fi
	@echo "All linting checks passed!"

# Run all pre-commit checks locally
pre-commit-check:
	@echo "Running all pre-commit validation checks..."
	@echo "1. Formatting checks..."
	@$(MAKE) fmt-check
	@echo "2. Building project..."
	@$(MAKE) build
	@echo "3. Running tests..."
	@$(MAKE) test
	@echo "4. Running golangci-lint..."
	golangci-lint run --timeout=5m
	@echo "5. Additional validation checks..."
	@if find . -name "*.go" -exec grep -l "<<<<<<< HEAD\|=======" {} \; | head -1 | grep -q .; then \
		echo "Merge conflict markers found in Go files!"; \
		exit 1; \
	fi
	@echo "All pre-commit checks passed! ✅"

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
	@echo ""
	@echo "Building & Running:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  install       - Install the binary to GOPATH/bin"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt           - Format code (goimports + mod tidy + whitespace)"
	@echo "  fmt-check     - Check formatting without making changes"
	@echo "  lint          - Run golangci-lint (basic)"
	@echo "  lint-strict   - Run comprehensive linting (matches pre-commit)"
	@echo "  pre-commit-check - Run all pre-commit validations locally"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo ""
	@echo "  help          - Show this help message"
