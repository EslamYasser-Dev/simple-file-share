.PHONY: build clean test

# Variables
BIN_NAME = bin/file-share
REL_MAIN_PATH = cmd/server/main.go
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
BIN_LINUX = $(BIN_NAME)-$(VERSION)-linux

.DEFAULT_GOAL := help

# Build Linux binary
build:
	@echo "🔨 Building server..."
	@cd backend && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" \
		-o ../$(BIN_LINUX) $(REL_MAIN_PATH)
	@echo "✅ Build complete: $(BIN_LINUX)"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/* coverage.html backend/coverage.out
	@echo "✅ Clean complete."

# Run tests with coverage
test:
	@echo "🧪 Running tests..."
	@(cd backend && go test -v ./... -coverprofile=coverage.out -covermode=count)
	@cd backend && go tool cover -func=coverage.out | grep "total:"
	@cd backend && go tool cover -html=coverage.out -o ../coverage.html
	@echo "✅ Tests complete. Open coverage.html for detailed report."

# Run the server in development mode
run: build
	@chmod +x $(BIN_LINUX)
	@echo "🚀 Starting server in development mode..."
	@./$(BIN_LINUX)

# Show help
help:
	@echo "🚀 Available targets:"
	@echo "  build    - Build Linux binary"
	@echo "  test     - Run tests with coverage"
	@echo "  clean    - Clean build artifacts"
	@echo "  run      - Run the server in development mode"
	@echo "  help     - Show this help message"

