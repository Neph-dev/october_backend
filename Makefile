# October Backend Makefile
# Following NASA clean code principles

.PHONY: help build run test clean lint format check-security deps seed-data

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the application binary
	@echo "Building application..."
	@go build -o bin/october-server ./cmd/api
	@go build -o bin/seed ./cmd/seed
	@go build -o bin/feed-processor ./cmd/feed-processor
	@echo "Build complete: bin/october-server, bin/seed, bin/feed-processor"

# Run the application
run: ## Run the application in development mode
	@echo "Starting application..."
	@go run ./cmd/api

# Seed database with initial data
seed-data: ## Seed the database with initial company data
	@echo "Seeding database..."
	@go run ./cmd/seed

# Process RSS feeds
process-feeds: ## Process RSS feeds for all companies
	@echo "Processing RSS feeds..."
	@go run ./cmd/feed-processor

# Process RSS feed for specific company
process-feed: ## Process RSS feed for specific company (usage: make process-feed COMPANY="Raytheon Technologies")
	@echo "Processing RSS feed for company: $(COMPANY)"
	@go run ./cmd/feed-processor -company="$(COMPANY)"

# Test all APIs
test-api: ## Test all API endpoints (requires server to be running)
	@echo "Testing API endpoints..."
	@./scripts/test_full_api.sh

# Run with debug logging
debug: ## Run the application with debug logging
	@echo "Starting application with debug logging..."
	@LOG_LEVEL=debug go run ./cmd/api

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	@go clean
	@rm -f bin/october-server bin/seed bin/feed-processor
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Lint the code
lint: ## Run linting tools
	@echo "Running linters..."
	@go vet ./...
	@gofmt -s -l .
	@echo "Lint complete"

# Format the code
format: ## Format Go code
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Format complete"

# Check for security issues
check-security: ## Run security checks
	@echo "Running security checks..."
	@go vet ./...
	@echo "Security check complete"

# Install dependencies
deps: ## Download and install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Verify dependencies
verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "Dependencies verified"

# Update dependencies
update-deps: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated"

# Run the application with live reload (requires air)
dev: ## Run with live reload (requires 'air' tool)
	@echo "Starting development server with live reload..."
	@air

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@echo "Tools installed"

# Docker build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t october-backend .
	@echo "Docker build complete"

# Docker run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env october-backend

# Production build
build-prod: ## Build for production with optimizations
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o bin/october-server ./cmd/api
	@echo "Production build complete"

# Check all (comprehensive check)
check-all: deps verify lint test check-security ## Run all checks
	@echo "All checks complete"