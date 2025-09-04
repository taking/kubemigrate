# Velero API Server Makefile

# Variables
APP_NAME := velero-api
BUILD_DIR := bin
BINARY_NAME := velero-cli
DOCKER_IMAGE := velero-api-server
DOCKER_TAG := latest
GO_VERSION := 1.24

# Build info
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

.PHONY: help build clean docker docker-build docker-run dev lint format deps swagger

# Default target
all: clean deps lint build

# Help target
help: ## Show this help message
	@echo "$(BLUE)Velero API Server - Available targets:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# Build targets
build: ## Build the application binary
	@echo "$(YELLOW)Building $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

build-linux: ## Build for Linux
	@echo "$(YELLOW)Building for Linux...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(RESET)"

build-darwin: ## Build for macOS
	@echo "$(YELLOW)Building for macOS...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/main.go
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-darwin-*$(RESET)"

build-windows: ## Build for Windows
	@echo "$(YELLOW)Building for Windows...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(RESET)"

build-all: build-linux build-darwin build-windows ## Build for all platforms

# Development targets
dev: ## Run the application in development mode
	@echo "$(YELLOW)Starting development server...$(RESET)"
	@go run ./cmd/main.go

run: build ## Build and run the application
	@echo "$(YELLOW)Running $(APP_NAME)...$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

runWithSwagger: build swagger ## Build and run the application with Swagger
	@echo "$(YELLOW)Running $(APP_NAME)...$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Code quality targets
lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
	fi

format: ## Format code
	@echo "$(YELLOW)Formatting code...$(RESET)"
	@go fmt ./...
	@go mod tidy
	@echo "$(GREEN)✓ Code formatted$(RESET)"

# Docker targets
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(RESET)"
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(YELLOW)Running Docker container...$(RESET)"
	@docker run --rm -p 9091:9091 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose targets
compose-up: docker-build ## Start services with docker-compose
	@echo "$(YELLOW)Starting services with docker-compose...$(RESET)"
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "$(GREEN)✓ Services started$(RESET)"

compose-down: ## Stop services with docker-compose
	@echo "$(YELLOW)Stopping services with docker-compose...$(RESET)"
	@docker-compose -f docker-compose.dev.yml down
	@echo "$(GREEN)✓ Services stopped$(RESET)"

compose-logs: ## Show docker-compose logs
	@docker-compose -f docker-compose.dev.yml logs -f

# Dependency management
deps: ## Download dependencies
	@echo "$(YELLOW)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

deps-update: ## Update dependencies
	@echo "$(YELLOW)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

# Swagger targets
swagger: ## Generate Swagger documentation
	@echo "$(YELLOW)Generating Swagger documentation...$(RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/main.go -o docs/swagger --parseDependency --parseInternal; \
		echo "$(GREEN)✓ Swagger documentation generated$(RESET)"; \
	else \
		echo "$(RED)swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest$(RESET)"; \
	fi

# Cleanup targets
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean completed$(RESET)"

clean-docker: ## Clean Docker images and containers
	@echo "$(YELLOW)Cleaning Docker artifacts...$(RESET)"
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker system prune -f
	@echo "$(GREEN)✓ Docker cleanup completed$(RESET)"

# Release targets
release: clean deps lint test build-all ## Prepare a release
	@echo "$(GREEN)✓ Release ready$(RESET)"

# Version info
version: ## Show version information
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

# Show current status
status: ## Show project status
	@echo "$(BLUE)Project Status:$(RESET)"
	@echo "Git Branch: $(shell git branch --show-current)"
	@echo "Git Status: $(shell git status --porcelain | wc -l | tr -d ' ') files changed"
	@echo "Go Version: $(shell go version | cut -d' ' -f3)"
	@echo "Dependencies: $(shell go list -m all | wc -l | tr -d ' ') modules"
