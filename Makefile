.PHONY: help install dev dev-backend dev-frontend build build-backend build-frontend test test-safety clean docs docs-build docs-serve docs-clean run-backend run-frontend stop db-reset

# Default target
.DEFAULT_GOAL := help

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Project variables
BINARY_NAME := veidly
BACKEND_DIR := backend
FRONTEND_DIR := frontend
DOCS_DIR := docs
BUILD_DIR := build

help: ## Show this help message
	@echo "$(BLUE)Veidly - Makefile Commands$(NC)"
	@echo "=============================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ============================================================================
# Installation & Setup
# ============================================================================

install: ## Install all dependencies (backend + frontend)
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@echo "$(YELLOW)→ Installing Go dependencies$(NC)"
	@cd $(BACKEND_DIR) && go mod download && go mod tidy
	@echo "$(YELLOW)→ Installing frontend dependencies$(NC)"
	@cd $(FRONTEND_DIR) && npm install
	@echo "$(GREEN)✓ All dependencies installed$(NC)"

setup: install setup-env ## Complete setup (install deps + create .env files)
	@echo "$(GREEN)✓ Setup complete! Run 'make dev' to start development$(NC)"

setup-env: ## Create .env files from examples if they don't exist
	@echo "$(BLUE)Setting up environment files...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(YELLOW)→ Created .env from .env.example$(NC)"; \
		echo "$(YELLOW)⚠  IMPORTANT: Edit .env and set secrets$(NC)"; \
		echo "$(YELLOW)   Generate secrets with: make generate-secrets$(NC)"; \
	else \
		echo "$(GREEN)✓ .env already exists$(NC)"; \
	fi
	@if [ ! -f $(FRONTEND_DIR)/.env ]; then \
		cp $(FRONTEND_DIR)/.env.example $(FRONTEND_DIR)/.env; \
		echo "$(YELLOW)→ Created frontend/.env from .env.example$(NC)"; \
	else \
		echo "$(GREEN)✓ frontend/.env already exists$(NC)"; \
	fi
	@echo "$(GREEN)✓ Environment files ready$(NC)"

check-env: ## Check if required environment variables are set
	@echo "$(BLUE)Checking environment configuration...$(NC)"
	@MISSING=0; \
	if [ ! -f .env ]; then \
		echo "$(RED)✗ .env file not found$(NC)"; \
		echo "  Run: make setup-env"; \
		MISSING=1; \
	else \
		if ! grep -q "JWT_SECRET=.*[a-zA-Z0-9]" .env; then \
			echo "$(RED)✗ JWT_SECRET not set in .env$(NC)"; \
			echo "  Generate: openssl rand -base64 43"; \
			MISSING=1; \
		else \
			echo "$(GREEN)✓ JWT_SECRET configured$(NC)"; \
		fi; \
	fi; \
	if [ $$MISSING -eq 0 ]; then \
		echo "$(GREEN)✓ Environment configuration looks good$(NC)"; \
	else \
		exit 1; \
	fi

# ============================================================================
# Development
# ============================================================================

dev: check-env check-deps ## Run both backend and frontend in development mode (parallel)
	@echo "$(BLUE)Starting development servers...$(NC)"
	@echo "$(YELLOW)Backend will run on http://localhost:8080$(NC)"
	@echo "$(YELLOW)Frontend will run on http://localhost:5173$(NC)"
	@echo "$(YELLOW)Health check: http://localhost:8080/health$(NC)"
	@echo ""
	@$(MAKE) -j2 dev-backend dev-frontend

check-deps: ## Check if dependencies are installed
	@MISSING=0; \
	if [ ! -d "$(BACKEND_DIR)/vendor" ] && [ ! -f "$(BACKEND_DIR)/go.sum" ]; then \
		echo "$(RED)✗ Backend dependencies not installed$(NC)"; \
		MISSING=1; \
	fi; \
	if [ ! -d "$(FRONTEND_DIR)/node_modules" ]; then \
		echo "$(RED)✗ Frontend dependencies not installed$(NC)"; \
		echo "$(YELLOW)  Run: make install$(NC)"; \
		MISSING=1; \
	fi; \
	if [ $$MISSING -eq 1 ]; then \
		exit 1; \
	fi

dev-backend: ## Run backend in development mode with auto-reload
	@echo "$(BLUE)Starting backend server...$(NC)"
	@if [ -f .env ]; then \
		export $$(cat .env | grep -v '^#' | xargs) && cd $(BACKEND_DIR) && go run .; \
	else \
		echo "$(YELLOW)⚠  No .env file found, using defaults$(NC)"; \
		cd $(BACKEND_DIR) && go run .; \
	fi

dev-frontend: ## Run frontend in development mode with hot reload
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	@cd $(FRONTEND_DIR) && npm run dev

dev-quick: ## Quick start (skip env check, for rapid iteration)
	@echo "$(BLUE)Quick starting development servers...$(NC)"
	@$(MAKE) -j2 dev-backend dev-frontend

run-backend: ## Run compiled backend binary
	@if [ ! -f $(BINARY_NAME) ]; then \
		echo "$(RED)✗ Binary not found. Run 'make build-backend' first$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running backend...$(NC)"
	./$(BINARY_NAME)

run-frontend: ## Serve built frontend (requires build first)
	@if [ ! -d "$(FRONTEND_DIR)/dist" ]; then \
		echo "$(RED)✗ Frontend build not found. Run 'make build-frontend' first$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Serving frontend on http://localhost:8000$(NC)"
	@cd $(FRONTEND_DIR)/dist && python3 -m http.server 8000

# ============================================================================
# Building
# ============================================================================

build: build-backend build-frontend ## Build both backend and frontend for production

build-backend: ## Build backend binary
	@echo "$(BLUE)Building backend...$(NC)"
	@cd $(BACKEND_DIR) && go build -o ../$(BINARY_NAME) -ldflags="-s -w" .
	@echo "$(GREEN)✓ Backend built: ./$(BINARY_NAME)$(NC)"

build-frontend: ## Build frontend for production
	@echo "$(BLUE)Building frontend...$(NC)"
	@cd $(FRONTEND_DIR) && npm run build
	@echo "$(GREEN)✓ Frontend built: ./$(FRONTEND_DIR)/dist/$(NC)"

# ============================================================================
# Testing
# ============================================================================

test: ## Run all tests (backend + frontend)
	@echo "$(BLUE)Running backend tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "$(BLUE)Backend Test Coverage:$(NC)"
	@cd $(BACKEND_DIR) && go tool cover -func=coverage.out | grep total | awk '{print "$(GREEN)" $$3 " of statements covered$(NC)"}'
	@echo ""
	@echo "$(BLUE)Running frontend tests...$(NC)"
	@cd $(FRONTEND_DIR) && npm test
	@echo "$(GREEN)✓ All tests completed$(NC)"

test-quick: ## Quick test (no race detection, for rapid feedback)
	@echo "$(BLUE)Running quick tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -short ./...
	@echo "$(GREEN)✓ Quick tests passed$(NC)"

test-backend: ## Run backend tests only
	@echo "$(BLUE)Running backend tests...$(NC)"
	@cd $(BACKEND_DIR) && go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "$(BLUE)Test Coverage:$(NC)"
	@cd $(BACKEND_DIR) && go tool cover -func=coverage.out | grep total | awk '{print "$(GREEN)" $$3 " of statements covered$(NC)"}'

test-backend-quick: ## Run backend tests quickly (no race detection)
	@echo "$(BLUE)Running backend tests (quick)...$(NC)"
	@cd $(BACKEND_DIR) && go test -v ./...

test-frontend: ## Run frontend tests only
	@echo "$(BLUE)Running frontend tests...$(NC)"
	@cd $(FRONTEND_DIR) && npm test

test-frontend-coverage: ## Run frontend tests with coverage report
	@echo "$(BLUE)Running frontend tests with coverage...$(NC)"
	@cd $(FRONTEND_DIR) && npm run test:coverage

test-coverage: test ## Run tests and show coverage report in browser
	@cd $(BACKEND_DIR) && go tool cover -html=coverage.out

test-safety: ## Verify test safety (production DB untouched)
	@echo "$(BLUE)Verifying test safety...$(NC)"
	@./verify_test_safety.sh

test-watch: ## Run tests in watch mode (requires entr)
	@if ! command -v entr > /dev/null; then \
		echo "$(RED)✗ entr not installed. Install with: brew install entr$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Watching for changes...$(NC)"
	@find $(BACKEND_DIR) -name '*.go' | entr -c make test

# ============================================================================
# Utilities
# ============================================================================

generate-secrets: ## Generate secure random secrets for .env
	@echo "$(BLUE)Generating secure secrets...$(NC)"
	@echo ""
	@echo "$(YELLOW)JWT_SECRET (43+ chars, copy to .env):$(NC)"
	@openssl rand -base64 43
	@echo ""
	@echo "$(YELLOW)ADMIN_PASSWORD (copy to .env):$(NC)"
	@openssl rand -base64 24 | tr -d "=+/" | cut -c1-20
	@echo ""
	@echo "$(GREEN)✓ Secrets generated. Add them to your .env file$(NC)"

logs: ## Tail application logs (if running)
	@echo "$(BLUE)Tailing logs...$(NC)"
	@tail -f veidly.log 2>/dev/null || echo "$(YELLOW)⚠  No log file found. Is the server running?$(NC)"

ps: ## Show running Veidly processes
	@echo "$(BLUE)Veidly Processes:$(NC)"
	@echo ""
	@ps aux | grep -E "(go run|vite|veidly)" | grep -v grep || echo "$(YELLOW)No processes running$(NC)"

# ============================================================================
# Database
# ============================================================================

db-reset: ## Reset database (WARNING: deletes all data)
	@echo "$(YELLOW)⚠ This will delete all data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -f veidly.db; \
		echo "$(GREEN)✓ Database deleted. Will be recreated on next run.$(NC)"; \
	else \
		echo "$(BLUE)Cancelled.$(NC)"; \
	fi

db-backup: ## Backup database to timestamped file
	@if [ ! -f veidly.db ]; then \
		echo "$(RED)✗ Database not found$(NC)"; \
		exit 1; \
	fi
	@TIMESTAMP=$$(date +%Y%m%d_%H%M%S); \
	cp veidly.db "veidly_backup_$$TIMESTAMP.db"; \
	echo "$(GREEN)✓ Backup created: veidly_backup_$$TIMESTAMP.db$(NC)"

# ============================================================================
# Documentation
# ============================================================================

docs: docs-build docs-serve ## Build and serve documentation

docs-build: ## Build Antora documentation
	@echo "$(BLUE)Building documentation...$(NC)"
	@if command -v docker > /dev/null; then \
		echo "$(YELLOW)→ Using Docker$(NC)"; \
		docker run --rm -v $$(pwd):/antora antora/antora:latest antora-playbook.yml; \
	elif command -v antora > /dev/null; then \
		echo "$(YELLOW)→ Using local Antora$(NC)"; \
		antora antora-playbook.yml; \
	else \
		echo "$(RED)✗ Neither Docker nor Antora found$(NC)"; \
		echo "$(YELLOW)Install Antora: npm install -g @antora/cli @antora/site-generator$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Documentation built: ./$(BUILD_DIR)/site/$(NC)"

docs-serve: ## Serve documentation (requires docs-build first)
	@if [ ! -d "$(BUILD_DIR)/site" ]; then \
		echo "$(RED)✗ Documentation not built. Run 'make docs-build' first$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Serving documentation on http://localhost:8000$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	@cd $(BUILD_DIR)/site && python3 -m http.server 8000

docs-docker: ## Build and serve docs using Docker Compose
	@echo "$(BLUE)Building documentation with Docker Compose...$(NC)"
	@docker-compose -f docker-compose.docs.yml run --rm antora
	@echo "$(BLUE)Starting documentation server...$(NC)"
	@echo "$(GREEN)Documentation: http://localhost:8000$(NC)"
	@docker-compose -f docker-compose.docs.yml up docs-server

docs-clean: ## Clean documentation build files
	@echo "$(BLUE)Cleaning documentation build...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✓ Documentation build cleaned$(NC)"

# ============================================================================
# Cleaning
# ============================================================================

clean: ## Clean all build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -rf $(FRONTEND_DIR)/dist
	@rm -rf $(FRONTEND_DIR)/node_modules/.vite
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	@rm -f veidly-test-suite.db
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-all: clean ## Clean everything including dependencies
	@echo "$(BLUE)Cleaning all dependencies...$(NC)"
	@rm -rf $(FRONTEND_DIR)/node_modules
	@cd $(BACKEND_DIR) && go clean -cache -modcache -testcache
	@echo "$(GREEN)✓ Everything cleaned$(NC)"

# ============================================================================
# Docker
# ============================================================================

docker-build: ## Build Docker image (if Dockerfile exists)
	@if [ ! -f Dockerfile ]; then \
		echo "$(YELLOW)⚠ No Dockerfile found$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t veidly:latest .
	@echo "$(GREEN)✓ Docker image built: veidly:latest$(NC)"

docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	@docker run -p 8080:8080 -v $$(pwd)/veidly.db:/app/veidly.db veidly:latest

# ============================================================================
# Linting & Formatting
# ============================================================================

lint: ## Run linters
	@echo "$(BLUE)Running linters...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		cd $(BACKEND_DIR) && golangci-lint run ./...; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not installed, using go vet$(NC)"; \
		cd $(BACKEND_DIR) && go vet ./...; \
	fi
	@cd $(FRONTEND_DIR) && npm run lint 2>/dev/null || echo "$(YELLOW)⚠ Frontend linting not configured$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	@cd $(BACKEND_DIR) && go fmt ./...
	@cd $(FRONTEND_DIR) && npm run format 2>/dev/null || echo "$(YELLOW)⚠ Frontend formatting not configured$(NC)"
	@echo "$(GREEN)✓ Code formatted$(NC)"

# ============================================================================
# Git Helpers
# ============================================================================

status: ## Show git status and project info
	@echo "$(BLUE)Project Status$(NC)"
	@echo "=============="
	@echo ""
	@echo "$(YELLOW)Git Status:$(NC)"
	@git status -s
	@echo ""
	@echo "$(YELLOW)Backend:$(NC)"
	@if [ -f $(BINARY_NAME) ]; then \
		echo "  Binary: $(GREEN)✓$(NC) $(BINARY_NAME)"; \
	else \
		echo "  Binary: $(RED)✗$(NC) Not built"; \
	fi
	@echo ""
	@echo "$(YELLOW)Frontend:$(NC)"
	@if [ -d "$(FRONTEND_DIR)/dist" ]; then \
		echo "  Build: $(GREEN)✓$(NC) $(FRONTEND_DIR)/dist"; \
	else \
		echo "  Build: $(RED)✗$(NC) Not built"; \
	fi
	@echo ""
	@echo "$(YELLOW)Documentation:$(NC)"
	@if [ -d "$(BUILD_DIR)/site" ]; then \
		echo "  Build: $(GREEN)✓$(NC) $(BUILD_DIR)/site"; \
	else \
		echo "  Build: $(RED)✗$(NC) Not built"; \
	fi
	@echo ""
	@echo "$(YELLOW)Database:$(NC)"
	@if [ -f veidly.db ]; then \
		echo "  File: $(GREEN)✓$(NC) veidly.db ($$(du -h veidly.db | cut -f1))"; \
	else \
		echo "  File: $(YELLOW)⚠$(NC) Not created yet"; \
	fi

commit: ## Stage and commit changes (prompts for message)
	@git add .
	@git status -s
	@echo ""
	@read -p "Commit message: " msg; \
	git commit -m "$$msg"

# ============================================================================
# Production Deployment
# ============================================================================

deploy-check: ## Check if ready for deployment
	@echo "$(BLUE)Checking deployment readiness...$(NC)"
	@echo ""
	@ERRORS=0; \
	if [ ! -f $(BINARY_NAME) ]; then \
		echo "$(RED)✗ Backend not built$(NC)"; \
		ERRORS=$$((ERRORS+1)); \
	else \
		echo "$(GREEN)✓ Backend binary exists$(NC)"; \
	fi; \
	if [ ! -d "$(FRONTEND_DIR)/dist" ]; then \
		echo "$(RED)✗ Frontend not built$(NC)"; \
		ERRORS=$$((ERRORS+1)); \
	else \
		echo "$(GREEN)✓ Frontend build exists$(NC)"; \
	fi; \
	if ! (cd $(BACKEND_DIR) && go test -short ./...) > /dev/null 2>&1; then \
		echo "$(RED)✗ Tests failing$(NC)"; \
		ERRORS=$$((ERRORS+1)); \
	else \
		echo "$(GREEN)✓ Tests passing$(NC)"; \
	fi; \
	if [ $$ERRORS -eq 0 ]; then \
		echo ""; \
		echo "$(GREEN)✓ Ready for deployment!$(NC)"; \
	else \
		echo ""; \
		echo "$(RED)✗ Not ready for deployment ($$ERRORS issues)$(NC)"; \
		exit 1; \
	fi

# ============================================================================
# Quick Commands
# ============================================================================

all: install build test docs-build ## Install, build, test, and build docs

quick-start: setup dev ## Complete setup and start development servers

start: dev-quick ## Alias for dev-quick (fastest way to start)

restart: ## Stop and restart development (useful after changes)
	@echo "$(YELLOW)⚠  Stopping servers (if running)...$(NC)"
	@pkill -f "go run" || true
	@pkill -f "vite" || true
	@sleep 1
	@$(MAKE) dev-quick

prod: build deploy-check ## Build everything and check deployment readiness

# Developer quick commands (one-letter aliases for speed)
b: build ## Quick build (alias: b)
t: test-quick ## Quick test (alias: t)
d: dev-quick ## Quick dev (alias: d)
c: clean ## Quick clean (alias: c)
