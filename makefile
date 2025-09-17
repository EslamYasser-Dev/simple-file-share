.PHONY: proto build windows-build run run-backend run-frontend run-frontend-background test clean push-tag help

# Metadata
BIN_NAME        = bin/file-share
REL_MAIN_PATH   = cmd/server/main.go
DATE            := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
VERSION         ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

.DEFAULT_GOAL := help

## Generate Go code from proto file (optional)
proto: ## Generate Go code from .proto
	@echo "📦 Generating Go code from Proto file..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_SRC)

build: ## Build Linux amd64 binary into bin/
	@echo "🔨 Building server..."
	@cd backend && go mod tidy
	@cd backend && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.BuildDate=$(DATE)" \
		-o ../$(BIN_NAME)-$(VERSION) $(REL_MAIN_PATH)
	@echo "✅ Build complete: $(BIN_NAME)-$(VERSION)"

windows-build: ## Build Windows amd64 binary into bin/
	@echo "🔨 Building server for Windows..."
	@cd backend && go mod tidy
	@cd backend && GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.BuildDate=$(DATE)" \
		-o ../$(BIN_NAME).exe $(REL_MAIN_PATH)
	@echo "✅ Build complete: $(BIN_NAME).exe"

## Run the server
run-backend: build 
	@clear
	@echo "🚀 Running file-share server..."
	@./$(BIN_NAME)-$(VERSION)
	@echo "👋 Server stopped."

## Alias for run-backend
run: run-backend ## Alias: run server (builds first)


run-frontend-background: ## Run frontend dev server in background
	@$(MAKE) run-frontend &




run-frontend: ## Start frontend dev server
	@(cd frontend && npm install && npm run dev)
## Run unit tests with coverage report
test: ## Run backend tests with coverage report
	@echo "🧪 Running tests..."
	@(cd backend && go test -v ./... -coverprofile=coverage.out)
	@cd backend && go tool cover -html=coverage.out -o ../coverage.html
	@echo "✅ Tests complete. Open coverage.html for report."

## Remove binaries and coverage files
clean: ## Clean all build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/* coverage.html backend/coverage.out
	@echo "✅ Clean complete."

## Commit and push with tag
push-tag: ## Commit changes and push new Git tag
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ ERROR: VERSION is not set. Usage: make push-tag VERSION=x.y.z"; \
		exit 1; \
	fi
	@git pull origin master
	@git add .
	@if git diff --cached --quiet; then \
		echo "✅ No changes to commit."; \
	else \
		git commit -m "Release $(VERSION)"; \
		echo "✅ Changes committed with message: 'Release $(VERSION)'"; \
	fi
	@echo "📤 Pushing changes to remote..."
	@git push origin $$(git rev-parse --abbrev-ref HEAD)
	@echo "🏷️  Creating and pushing tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) pushed successfully."

## Show all available targets
help:
	@echo "🚀 Makefile Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'