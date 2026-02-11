# Testing Infrastructure

This document describes the testing infrastructure set up for the Mneme project.

## Overview

The project now has a comprehensive testing infrastructure with:

- Unit tests for core packages
- Test coverage reporting
- Benchmark tests
- CI/CD pipeline with GitHub Actions
- Makefile for easy test execution

## Test Structure

### Test Files

- `internal/logger/logger_test.go` - Tests for logger package
- `internal/config/load_test.go` - Tests for config loading
- `internal/config/save_test.go` - Tests for config saving (to be added)
- `internal/utils/config_test.go` - Tests for utility functions
- `internal/core/config_test.go` - Tests for core configuration types
- `internal/logger/benchmark_test.go` - Existing benchmark tests

### Dependencies

The following testing dependencies have been added to `go.mod`:

- `github.com/stretchr/testify` - Assertion and mocking library
- `github.com/google/go-cmp` - Deep comparison for tests
- `gopkg.in/yaml.v3` - YAML support (indirect)

## Running Tests

### Using Makefile

The `Makefile` provides several test targets:

```bash
# Run all tests with coverage
make test

# Run tests without race detector
make test-short

# Run tests with verbose output
make test-verbose

# Run benchmarks
make test-bench

# Run custom performance benchmarks (table output)
make benchmarks

# Generate HTML coverage report
make test-coverage

# Clean test artifacts
make test-clean
```

### Using Go Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests for specific package
go test ./internal/logger -v

# Run benchmarks
go test -bench=. -benchmem ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
```

## CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/test.yml`) includes:

### Jobs

1. **test** - Runs tests with coverage on Ubuntu
   - Tests with race detector
   - Uploads coverage to Codecov
   - Shows test summary

2. **lint** - Code quality checks
   - Runs `go vet`
   - Checks code formatting

3. **build** - Build verification
   - Builds on Ubuntu, macOS, and Windows
   - Tests the binary

4. **benchmark** - Runs benchmarks
   - Ensures performance tests pass

5. **security** - Security scanning
   - Uses `govulncheck` for vulnerability scanning

6. **test-matrix** - Multi-OS testing
   - Tests on Ubuntu and macOS

### Triggers

The workflow runs on:
- Push to `main` or `master` branches
- Pull requests to `main` or `master` branches

## Test Coverage

### Coverage Report

After running tests with coverage, you can view the report:

```bash
# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# View coverage summary
go tool cover -func=coverage.out
```

### Coverage Files

- `coverage.out` - Raw coverage data
- `coverage.html` - HTML report (generated)
- `coverage.txt` - Text summary (generated)

## Test Categories

### Unit Tests

- **Logger Tests**: Test initialization, logging levels, and output formatting
- **Config Tests**: Test configuration loading, saving, and marshaling
- **Core Tests**: Test data structures and default values
- **Utils Tests**: Test utility functions like path expansion

### Integration Tests

- Config file operations (read/write)
- Path validation and expansion
- TOML marshaling/unmarshaling

### Benchmark Tests

- Logger performance benchmarks
- Comparison with standard library

## Best Practices

### Writing Tests

1. Use subtests with `t.Run()` for better organization
2. Use table-driven tests for multiple scenarios
3. Use temporary directories for file operations
4. Reset global state between tests
5. Use appropriate assertions from testify

### Test Naming

- Use descriptive test names
- Follow the pattern: `TestFunctionName/Scenario`
- Use `t.Run()` for subtests

### Test Isolation

- Each test should be independent
- Clean up resources after tests
- Use `t.TempDir()` for temporary files
- Reset global variables between tests

## Troubleshooting

### Common Issues

1. **Import errors**: Run `go mod tidy` to resolve dependencies
2. **Race conditions**: Use `-race` flag to detect
3. **Coverage not showing**: Ensure tests are passing first
4. **Build failures**: Check for syntax errors or missing imports

### Debugging

```bash
# Run specific test with verbose output
go test -v -run TestFunctionName ./...

# Run with race detector
go test -race ./...

# Debug with delve
dlv test ./internal/logger
```

## Future Improvements

- Add more integration tests
- Add end-to-end tests
- Add fuzz tests for security-critical code
- Add property-based tests
- Add test fixtures and test data
- Add mock implementations for external dependencies

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Coverage](https://go.dev/blog/cover)
- [GitHub Actions](https://docs.github.com/en/actions)
