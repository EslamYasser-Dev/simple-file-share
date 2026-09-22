# ==============================================================================
# Simple File Share — developer & CI targets
# ==============================================================================

SHELL := /bin/sh

# ---- Project layout ----------------------------------------------------------
BACKEND_DIR  := backend
FRONTEND_DIR := frontend
BIN_DIR      := bin
BUILD_DIR    := build

# ---- Binary / version --------------------------------------------------------
APP_NAME     := file-share
VERSION      ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BIN_NAME     := $(BIN_DIR)/$(APP_NAME)
BIN_LINUX    := $(BIN_NAME)-$(VERSION)-linux

# ---- Go ----------------------------------------------------------------------
GO           ?= go
SERVER_PKG   := ./cmd/server
LDFLAGS      := -s -w
COVER_FILE   := $(BACKEND_DIR)/coverage.out
COVER_MIN    ?= 50

# ---- Node --------------------------------------------------------------------
NPM          ?= npm

# ---- Docker ------------------------------------------------------------------
IMAGE_NAME   ?= simple-file-share
IMAGE_TAG    ?= $(VERSION)
COMPOSE      ?= docker compose

.DEFAULT_GOAL := help

# ==============================================================================
# Help
# ==============================================================================
.PHONY: help
help: ## Show this help
	@printf "\n\033[1mSimple File Share\033[0m — available targets:\n\n"
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\n"

# ==============================================================================
# Setup
# ==============================================================================
.PHONY: setup
setup: backend-deps frontend-deps ## Install backend + frontend dependencies

.PHONY: backend-deps
backend-deps: ## Download Go modules
	@echo "📦 Downloading Go modules..."
	@cd $(BACKEND_DIR) && $(GO) mod download

.PHONY: frontend-deps
frontend-deps: ## Install frontend dependencies (npm ci)
	@echo "📦 Installing frontend dependencies..."
	@cd $(FRONTEND_DIR) && $(NPM) ci

# ==============================================================================
# Formatting & static analysis
# ==============================================================================
.PHONY: fmt
fmt: ## Format Go sources
	@echo "🎨 Formatting Go sources..."
	@cd $(BACKEND_DIR) && $(GO) fmt ./...

.PHONY: fmt-check
fmt-check: ## Fail if Go sources are not gofmt-clean
	@echo "🔍 Checking Go formatting..."
	@cd $(BACKEND_DIR) && test -z "$$(gofmt -l .)"

.PHONY: tidy
tidy: ## Tidy Go module files
	@cd $(BACKEND_DIR) && $(GO) mod tidy

.PHONY: vet
vet: ## Run go vet
	@echo "🔬 Running go vet..."
	@cd $(BACKEND_DIR) && $(GO) vet ./...

.PHONY: lint-frontend
lint-frontend: ## Lint the frontend
	@echo "🔬 Linting frontend..."
	@cd $(FRONTEND_DIR) && $(NPM) run lint

.PHONY: lint
lint: fmt-check vet lint-frontend ## Run every linter

# ==============================================================================
# Build
# ==============================================================================
.PHONY: build
build: build-linux ## Alias for build-linux

.PHONY: build-linux
build-linux: ## Build the static Linux server binary
	@echo "🔨 Building server ($(VERSION))..."
	@mkdir -p $(BIN_DIR)
	@cd $(BACKEND_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		$(GO) build -ldflags="$(LDFLAGS)" -o ../$(BIN_LINUX) $(SERVER_PKG)
	@echo "✅ Built $(BIN_LINUX)"

.PHONY: build-local
build-local: ## Build the server for the host platform
	@echo "🔨 Building local server..."
	@mkdir -p $(BIN_DIR)
	@cd $(BACKEND_DIR) && $(GO) build -ldflags="$(LDFLAGS)" -o ../$(BIN_NAME) $(SERVER_PKG)
	@echo "✅ Built $(BIN_NAME)"

.PHONY: frontend-build
frontend-build: ## Build the React frontend into frontend/dist
	@echo "🔨 Building frontend..."
	@cd $(FRONTEND_DIR) && $(NPM) run build
	@echo "✅ Frontend build complete"

.PHONY: build-all
build-all: build-linux frontend-build ## Build backend + frontend

# ==============================================================================
# Test & coverage
# ==============================================================================
.PHONY: test
test: test-backend test-frontend ## Run all tests

.PHONY: test-backend
test-backend: ## Run backend tests with coverage
	@echo "🧪 Running backend tests..."
	@cd $(BACKEND_DIR) && $(GO) test -v ./... -coverprofile=coverage.out -covermode=count
	@cd $(BACKEND_DIR) && $(GO) tool cover -func=coverage.out | grep "total:"

.PHONY: test-frontend
test-frontend: ## Run frontend tests (skipped when no test script)
	@echo "🧪 Running frontend tests..."
	@cd $(FRONTEND_DIR) && $(NPM) run test --if-present

.PHONY: coverage
coverage: test-backend ## Generate the HTML coverage report
	@cd $(BACKEND_DIR) && $(GO) tool cover -html=coverage.out -o ../coverage.html
	@echo "✅ Coverage report: coverage.html"

.PHONY: coverage-check
coverage-check: ## Enforce a minimum backend coverage percentage
	@cd $(BACKEND_DIR) && $(GO) test ./... -coverprofile=coverage.out -covermode=count > /dev/null
	@cd $(BACKEND_DIR) && $(GO) tool cover -func=coverage.out | grep "total:" | \
		awk -v min=$(COVER_MIN) '{ gsub(/%/,"",$$3); if ($$3+0 < min) { printf "❌ Coverage %s%% < %s%%\n", $$3, min; exit 1 } else { printf "✅ Coverage %s%% >= %s%%\n", $$3, min } }'

# ==============================================================================
# Run
# ==============================================================================
.PHONY: run
run: build-local ## Run the server in development mode
	@echo "🚀 Starting server (development)..."
	@APP_ENV=development ./$(BIN_NAME)

.PHONY: run-prod
run-prod: build-local frontend-build ## Run the server in production mode
	@echo "🚀 Starting server (production)..."
	@APP_ENV=production STATIC_DIR=$(FRONTEND_DIR)/dist ./$(BIN_NAME)

.PHONY: dev-frontend
dev-frontend: ## Start the Vite dev server
	@cd $(FRONTEND_DIR) && $(NPM) run dev

# ==============================================================================
# Docker
# ==============================================================================
.PHONY: docker-build
docker-build: ## Build the Docker image
	@docker build -f dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_NAME):latest .

.PHONY: docker-up
docker-up: ## Start services with Docker Compose
	@$(COMPOSE) up --build -d

.PHONY: docker-down
docker-down: ## Stop Docker Compose services
	@$(COMPOSE) down

.PHONY: docker-logs
docker-logs: ## Tail Docker Compose logs
	@$(COMPOSE) logs -f

# ==============================================================================
# GitHub Pages
# ==============================================================================
# The Pages UI is static; the Go API must be reachable at VITE_API_URL.
#   make deploy-pages VITE_API_URL=https://api.example.com
.PHONY: deploy-pages
deploy-pages: ## Build the frontend and publish it to GitHub Pages
	@echo "🌐 Deploying frontend to GitHub Pages..."
	@bash scripts/deploy-pages.sh
	@echo "✅ Pages deploy complete"

# ==============================================================================
# Housekeeping
# ==============================================================================
.PHONY: clean
clean: ## Remove build, coverage, and frontend dist artifacts
	@echo "🧹 Cleaning artifacts..."
	@rm -rf $(BIN_DIR)/* $(BUILD_DIR) $(COVER_FILE) coverage.html $(FRONTEND_DIR)/dist
	@echo "✅ Clean complete"

.PHONY: clean-all
clean-all: clean ## Clean plus frontend node_modules
	@rm -rf $(FRONTEND_DIR)/node_modules
	@echo "✅ Deep clean complete"

# ==============================================================================
# CI
# ==============================================================================
.PHONY: ci
ci: fmt-check vet test-backend lint-frontend frontend-build ## Full CI pipeline
