# Makefile for Mneme project

# Variables
GO := go
GOFLAGS := 
TESTFLAGS := -race
COVERAGEFLAGS := -race -coverprofile=/tmp/coverage.out -covermode=atomic
BENCHFLAGS := -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
COVERAGE_HTML := coverage.html
COVERAGE_TXT := coverage.txt

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help:
	@echo "Mneme Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make test          - Run all tests with coverage"
	@echo "  make test-short    - Run tests without race detector"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-bench    - Run benchmarks"
	@echo "  make test-coverage - Generate HTML coverage report"
	@echo "  make test-clean    - Clean test artifacts"
	@echo "  make build         - Build the project"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make lint          - Run linter"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Run go vet"
	@echo "  make all           - Run all checks and build"
	@echo ""

# Test targets
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test $(TESTFLAGS) ./... 2>&1 | tee /tmp/test_output.txt | grep -E "^(ok|FAIL|---|	--- )" || true
	@echo ""
	@echo "=== Test Summary ==="
	@echo "Passed: $$(grep -c '^ok' /tmp/test_output.txt || echo 0)"
	@echo "Failed: $$(grep -c '^FAIL' /tmp/test_output.txt || echo 0)"
	@if grep -q '^FAIL' /tmp/test_output.txt; then exit 1; fi

.PHONY: test-verbose
test-verbose:
	@echo "Running tests (verbose mode)..."
	$(GO) test -v $(TESTFLAGS) ./...
	@echo ""

.PHONY: test-short
test-short:
	@echo "Running tests (short mode, no race detector)..."
	$(GO) test ./...
	@echo ""

.PHONY: test-bench
test-bench:
	@echo "Running benchmarks..."
	$(GO) test $(GOFLAGS) $(BENCHFLAGS) ./...
	@echo "Benchmark results saved to cpu.prof and mem.prof"
	@echo ""

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GO) test $(COVERAGEFLAGS) ./... 2>&1 | grep -E "^(ok|FAIL|coverage)" || true
	@echo ""
	@echo "Generating coverage report..."
	$(GO) tool cover -html=/tmp/coverage.out -o $(COVERAGE_HTML)
	$(GO) tool cover -func=/tmp/coverage.out -o $(COVERAGE_TXT)
	@echo "Coverage report: $(COVERAGE_HTML)"
	@echo "Coverage summary: $(COVERAGE_TXT)"
	@echo ""

.PHONY: test-clean
test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out $(COVERAGE_HTML) $(COVERAGE_TXT) cpu.prof mem.prof *.bench *.bench.out
	@echo ""

# Build targets
.PHONY: build
build:
	@echo "Building project..."
	$(GO) build -o mneme ./cmd/mneme
	@echo "Build complete: ./mneme"
	@echo ""

.PHONY: build-dev
build-dev:
	@echo "Building project (dev mode)..."
	$(GO) build -o mneme -gcflags="-N -l" ./cmd/mneme
	@echo "Build complete: ./mneme"
	@echo ""

.PHONY: build-linux
build-linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/mneme-linux-amd64 ./cmd/mneme
	@echo "Build complete: dist/mneme-linux-amd64"

.PHONY: build-windows
build-windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/mneme-windows-amd64.exe ./cmd/mneme
	@echo "Build complete: dist/mneme-windows-amd64.exe"

.PHONY: build-macos
build-macos:
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o dist/mneme-darwin-amd64 ./cmd/mneme
	@echo "Build complete: dist/mneme-darwin-amd64"
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-s -w" -o dist/mneme-darwin-arm64 ./cmd/mneme
	@echo "Build complete: dist/mneme-darwin-arm64"

.PHONY: release
release: clean build-linux build-windows build-macos
	@echo "Release build complete. Artifacts in dist/"


# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f mneme *.test *.out *.prof *.bench *.bench.out
	$(GO) clean -cache -testcache
	@echo ""

# Code quality targets
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GO) vet ./...
	@echo ""

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo ""

.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo ""

# All checks
.PHONY: all
all: fmt vet test build
	@echo "All checks passed!"
	@echo ""

# CI/CD targets
.PHONY: ci-test
ci-test:
	@echo "Running CI tests..."
	$(GO) test $(GOFLAGS) -race -coverprofile=/tmp/coverage.out -covermode=atomic ./...
	$(GO) tool cover -func=/tmp/coverage.out
	@echo ""

.PHONY: ci-lint
ci-lint:
	@echo "Running CI lint..."
	$(GO) vet ./...
	@echo ""

.PHONY: ci-build
ci-build:
	@echo "Running CI build..."
	$(GO) build -o mneme ./cmd/mneme
	@echo ""

# Development helpers
.PHONY: watch
watch:
	@echo "Installing air for live reload..."
	@command -v air >/dev/null 2>&1 || $(GO) install github.com/cosmtrek/air@latest
	@echo "Starting live reload..."
	air -c .air.toml

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy
	@echo ""

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo ""

# Test summary
.PHONY: test-summary
test-summary: test
	@echo "Test Summary:"
	@echo "  - Coverage report: coverage.html"
	@echo "  - Coverage summary: coverage.txt"
	@echo "  - Run 'make test-coverage' to view detailed coverage"
	@echo ""
