.PHONY: test coverage bench lint clean help

# Default target
all: test lint

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem -run=^$$ ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Run fuzz tests
fuzz:
	@echo "Running fuzz tests..."
	go test -fuzz=FuzzParse -fuzztime=30s

# Clean generated files
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html

# Run all quality checks
check: test lint
	@echo "All checks passed!"

# Install tools
tools:
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  test     - Run tests with race detection"
	@echo "  coverage - Run tests with coverage report"
	@echo "  bench    - Run benchmarks"
	@echo "  lint     - Run linter"
	@echo "  fuzz     - Run fuzz tests"
	@echo "  clean    - Clean generated files"
	@echo "  check    - Run tests and linting"
	@echo "  tools    - Install development tools"
	@echo "  help     - Show this help"