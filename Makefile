# zkTest1 Makefile - Zero-Knowledge Proof Balance Verification

.PHONY: test test-unit test-integration test-e2e test-all test-short test-verbose test-coverage clean build run benchmark help

# Default Go command
GO := go

# Test configurations
TEST_TIMEOUT := 30m
TEST_FLAGS := -v -timeout $(TEST_TIMEOUT)
SHORT_FLAGS := -short -timeout 2m

# Build the application
build:
	@echo "Building zkTest1..."
	$(GO) build -o zktest1 .

# Run the application
run:
	@echo "Starting zkTest1 server..."
	$(GO) run .

# Run all tests (includes slow ZK proof tests)
test-all:
	@echo "Running all tests (including slow ZK proof generation/verification)..."
	$(GO) test $(TEST_FLAGS) ./...

# Run unit tests only (fast tests)
test-unit:
	@echo "Running unit tests..."
	$(GO) test $(TEST_FLAGS) -run "Test.*Circuit|Test.*Proof" ./...

# Run integration tests (API endpoint tests)
test-integration:
	@echo "Running integration tests..."
	$(GO) test $(TEST_FLAGS) -run "Test.*API|Test.*Store|Test.*Generate|Test.*Validate" ./...

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	$(GO) test $(TEST_FLAGS) -run "TestEndToEnd|TestConcurrent|TestEdge" ./...

# Run short/fast tests only (skips slow ZK operations)
test-short:
	@echo "Running short tests (skipping slow ZK proof operations)..."
	$(GO) test $(SHORT_FLAGS) ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	$(GO) test -v -timeout $(TEST_TIMEOUT) ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage report..."
	$(GO) test -v -timeout $(TEST_TIMEOUT) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem -timeout $(TEST_TIMEOUT) ./...

# Run benchmarks for proof operations only
benchmark-proof:
	@echo "Running proof benchmarks..."
	$(GO) test -bench="BenchmarkProof|BenchmarkEndToEnd" -benchmem -timeout $(TEST_TIMEOUT) ./...

# Default test target (reasonable for CI/development)
test: test-short

# Clean build artifacts and test outputs
clean:
	@echo "Cleaning build artifacts..."
	rm -f zktest1
	rm -f coverage.out coverage.html
	rm -f *.prof

# Lint the code (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo >&2 "golangci-lint is required but not installed. Visit: https://golangci-lint.run/usage/install/"; exit 1; }
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Check for code formatting issues
fmt-check:
	@echo "Checking code formatting..."
	@test -z $$($(GO) fmt ./...) || (echo "Code is not formatted. Run 'make fmt'"; exit 1)

# Tidy go modules
tidy:
	@echo "Tidying go modules..."
	$(GO) mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download

# Run static analysis
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Full check (format, lint, test)
check: fmt-check vet lint test

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Installing golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	@echo "Development setup complete!"

# Docker build (if you want to containerize later)
docker-build:
	@echo "Building Docker image..."
	docker build -t zktest1 .

# Display help
help:
	@echo "zkTest1 - Zero-Knowledge Proof Balance Verification"
	@echo ""
	@echo "Available commands:"
	@echo "  build           Build the application binary"
	@echo "  run             Run the application server"
	@echo "  test            Run short tests (default, good for development)"
	@echo "  test-all        Run all tests including slow ZK proof tests"
	@echo "  test-unit       Run unit tests only"
	@echo "  test-integration Run API integration tests"
	@echo "  test-e2e        Run end-to-end workflow tests"
	@echo "  test-short      Run fast tests only (skip slow ZK operations)"
	@echo "  test-verbose    Run tests with verbose output"
	@echo "  test-coverage   Generate test coverage report"
	@echo "  benchmark       Run all benchmarks"
	@echo "  benchmark-proof Run proof-specific benchmarks"
	@echo "  clean           Clean build artifacts"
	@echo "  lint            Run code linter"
	@echo "  fmt             Format code"
	@echo "  fmt-check       Check code formatting"
	@echo "  vet             Run go vet"
	@echo "  tidy            Tidy go modules"
	@echo "  deps            Download dependencies"
	@echo "  check           Run full check (format, lint, test)"
	@echo "  dev-setup       Set up development environment"
	@echo "  help            Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make test              # Quick tests for development"
	@echo "  make test-all          # Full test suite (slow)"
	@echo "  make test-coverage     # Tests with coverage report"
	@echo "  make benchmark-proof   # Benchmark ZK proof operations"

# Set default target
.DEFAULT_GOAL := help