# ==============================================================================
# Simple File Share — backend developer & CI targets
# ==============================================================================
#
# This repository is backend-only. The web clients live in sibling repos:
#   frontend/  React + Vite SPA
#   landing/   Next.js landing site
#   mobile/    Flutter app           → stores
# ==============================================================================

SHELL := /bin/sh

# ---- Project layout ----------------------------------------------------------
BACKEND_DIR  := backend
BIN_DIR      := bin
BUILD_DIR    := build

# ---- Binary / version --------------------------------------------------------
APP_NAME     := app
VERSION      ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BIN_NAME     := $(BIN_DIR)/$(APP_NAME)
BIN_LINUX    := $(BIN_NAME)-$(VERSION)-linux

# ---- Go ----------------------------------------------------------------------
GO           ?= go
SERVER_PKG   := ./cmd/server
LDFLAGS      := -s -w
COVER_FILE   := $(BACKEND_DIR)/coverage.out
COVER_MIN    ?= 40

# ---- Docker ------------------------------------------------------------------
IMAGE_NAME   ?= simple-file-share
IMAGE_TAG    ?= $(VERSION)
COMPOSE      ?= podman compose

.DEFAULT_GOAL := help

# ==============================================================================
# Help
# ==============================================================================
.PHONY: help
help: ## Show this help
	@printf "\n\033[1mSimple File Share (backend)\033[0m — available targets:\n\n"
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@printf "\n"

# ==============================================================================
# Setup
# ==============================================================================
.PHONY: setup
setup: backend-deps ## Install backend dependencies

.PHONY: backend-deps
backend-deps: ## Download Go modules
	@echo "📦 Downloading Go modules..."
	@cd $(BACKEND_DIR) && $(GO) mod download

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

.PHONY: tidy-check
tidy-check: ## Fail if go.mod/go.sum are not tidy
	@echo "📦 Checking module tidiness..."
	@cd $(BACKEND_DIR) && $(GO) mod tidy && git diff --exit-code -- go.mod go.sum

.PHONY: vet
vet: ## Run go vet
	@echo "🔬 Running go vet..."
	@cd $(BACKEND_DIR) && $(GO) vet ./...

.PHONY: lint
lint: fmt-check vet ## Run every linter

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

.PHONY: build-check
build-check: ## Compile every package without emitting a binary
	@echo "🔨 Checking compilation..."
	@cd $(BACKEND_DIR) && $(GO) build ./...

# ==============================================================================
# Test & coverage
# ==============================================================================
.PHONY: test
test: test-backend ## Run all tests

.PHONY: test-backend
test-backend: ## Run backend tests with coverage
	@echo "🧪 Running backend tests..."
	@cd $(BACKEND_DIR) && $(GO) test -v ./... -coverprofile=coverage.out -covermode=count
	@cd $(BACKEND_DIR) && $(GO) tool cover -func=coverage.out | grep "total:"

.PHONY: test-race
test-race: ## Run backend tests with the race detector (what CI uses)
	@echo "🧪 Running backend tests (race)..."
	@cd $(BACKEND_DIR) && $(GO) test -race -covermode=atomic -coverprofile=coverage.out ./...

.PHONY: coverage-html
coverage-html: ## Render coverage.html from the existing coverage.out
	@cd $(BACKEND_DIR) && $(GO) tool cover -html=coverage.out -o ../coverage.html
	@echo "✅ Coverage report: coverage.html"

.PHONY: coverage
coverage: test-backend coverage-html ## Run tests and generate the HTML coverage report

.PHONY: coverage-floor
coverage-floor: ## Fail if the existing coverage.out is below COVER_MIN
	@cd $(BACKEND_DIR) && $(GO) tool cover -func=coverage.out | grep "total:" | \
		awk -v min=$(COVER_MIN) '{ gsub(/%/,"",$$3); if ($$3+0 < min) { printf "❌ Coverage %s%% < %s%%\n", $$3, min; exit 1 } else { printf "✅ Coverage %s%% >= %s%%\n", $$3, min } }'

.PHONY: coverage-check
coverage-check: test-backend coverage-floor ## Run tests and enforce the minimum coverage percentage

# ==============================================================================
# Run
# ==============================================================================
.PHONY: run
run: build-local ## Run the server in development mode
	@echo "🚀 Starting server (development)..."
	@APP_ENV=development ./$(BIN_NAME)

.PHONY: run-prod
run-prod: build-local ## Run the server in production mode (API only)
	@echo "🚀 Starting server (production)..."
	@APP_ENV=production ./$(BIN_NAME)

# ==============================================================================
# Containers (Podman)
# ==============================================================================
.PHONY: podman-build
podman-build: ## Build the container image with Podman
	@podman build -f Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_NAME):latest .

.PHONY: podman-up
podman-up: ## Start services with Podman Compose
	@$(COMPOSE) up --build -d

.PHONY: podman-down
podman-down: ## Stop Podman Compose services
	@$(COMPOSE) down

.PHONY: podman-logs
podman-logs: ## Tail Podman Compose logs
	@$(COMPOSE) logs -f

# ==============================================================================
# Housekeeping
# ==============================================================================
.PHONY: clean
clean: ## Remove build and coverage artifacts
	@echo "🧹 Cleaning artifacts..."
	@rm -rf $(BIN_DIR)/* $(BUILD_DIR) $(COVER_FILE) coverage.html
	@echo "✅ Clean complete"

# ==============================================================================
# CI
# ==============================================================================
.PHONY: ci
ci: fmt-check tidy-check vet build-check test-race coverage-html coverage-floor ## Full CI pipeline — this is what GitHub Actions runs
