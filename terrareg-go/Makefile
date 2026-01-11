.PHONY: help build run test lint fmt clean migrate dev docker

# Variables
BINARY_NAME=terrareg
MAIN_PATH=cmd/server/main.go
MIGRATE_PATH=cmd/migrate/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary created at bin/$(BINARY_NAME)"

build-all: ## Build all binaries
	@echo "Building all binaries..."
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Server binary created at bin/$(BINARY_NAME)"

run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GORUN) $(MAIN_PATH)

dev: ## Run in development mode with auto-reload (requires air)
	@echo "Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without auto-reload..."; \
		$(GORUN) $(MAIN_PATH); \
	fi

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./test/integration/...

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		$(GOVET) ./...; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f terrareg.db
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@echo "TODO: Implement migration command"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@echo "TODO: Implement migration rollback"

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration $(NAME)..."
	@echo "TODO: Implement migration creation"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t terrareg:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 5000:5000 terrareg:latest

install: ## Install the binary
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)

# Development helpers
.PHONY: watch
watch: ## Watch for changes and rebuild (requires entr)
	@if command -v entr > /dev/null; then \
		find . -name '*.go' | entr -r make run; \
	else \
		echo "entr not installed. Use 'make dev' instead."; \
	fi

.PHONY: db-reset
db-reset: ## Reset the database (WARNING: deletes all data)
	@echo "Resetting database..."
	rm -f terrareg.db
	@echo "Database reset complete. Run 'make run' to recreate."

.PHONY: db-console
db-console: ## Open SQLite console
	@if [ -f terrareg.db ]; then \
		sqlite3 terrareg.db; \
	else \
		echo "Database file not found. Run 'make run' first."; \
	fi

.DEFAULT_GOAL := help
