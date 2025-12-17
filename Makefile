.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down

# Variables
APP_NAME=insurance-portal-api
DOCKER_IMAGE=$(APP_NAME):latest
PORT=8080

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) cmd/api/main.go
	@echo "Build complete!"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p $(PORT):$(PORT) --env-file .env $(DOCKER_IMAGE)

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	@# TODO: Add migration command
	@echo "Migrations complete"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@# TODO: Add rollback command
	@echo "Rollback complete"

db-setup: ## Setup database (create DB and run migrations)
	@echo "Setting up database..."
	@# TODO: Add database setup commands
	@echo "Database setup complete"

dev: ## Run development server with hot reload
	@echo "Starting development server..."
	@# Requires air: go install github.com/cosmtrek/air@latest
	@air

.DEFAULT_GOAL := help