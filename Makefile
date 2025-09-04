# KubeMigrate API Server Makefile

# 변수 정의
APP_NAME := kubemigrate-api
BUILD_DIR := bin
BINARY_NAME := kubemigrate-cli
DOCKER_IMAGE := kubemigrate-api-server
DOCKER_TAG := latest
GO_VERSION := 1.24

# 빌드 정보
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go 빌드 플래그
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"
LDFLAGS_OPTIMIZED := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# 출력 색상
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

.PHONY: help build clean docker docker-build docker-run dev lint format deps swagger

# 기본 타겟
all: clean deps lint build

# 도움말 타겟
help: ## 도움말 메시지 표시
	@echo "$(BLUE)KubeMigrate API Server - Available targets:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'

# 빌드 타겟
build: ## 애플리케이션 바이너리 빌드
	@echo "$(YELLOW)Building $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

build-optimized: ## 최적화된 바이너리 빌드 (작은 크기)
	@echo "$(YELLOW)Building optimized $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS_OPTIMIZED) -o $(BUILD_DIR)/$(BINARY_NAME)-optimized ./cmd/main.go
	@echo "$(GREEN)✓ Built optimized $(BUILD_DIR)/$(BINARY_NAME)-optimized$(RESET)"

build-compressed: ## UPX로 바이너리 빌드 및 압축
	@echo "$(YELLOW)Building and compressing $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS_OPTIMIZED) -o $(BUILD_DIR)/$(BINARY_NAME)-temp ./cmd/main.go
	@upx --best --lzma $(BUILD_DIR)/$(BINARY_NAME)-temp -o $(BUILD_DIR)/$(BINARY_NAME)-compressed
	@rm $(BUILD_DIR)/$(BINARY_NAME)-temp
	@echo "$(GREEN)✓ Built and compressed $(BUILD_DIR)/$(BINARY_NAME)-compressed$(RESET)"

build-linux: ## Linux용 빌드
	@echo "$(YELLOW)Building for Linux...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(RESET)"

build-darwin: ## macOS용 빌드
	@echo "$(YELLOW)Building for macOS...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/main.go
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-darwin-*$(RESET)"

build-windows: ## Windows용 빌드
	@echo "$(YELLOW)Building for Windows...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(RESET)"

build-all: build-linux build-darwin build-windows ## 모든 플랫폼용 빌드

# 아키텍처별 Docker 빌드
docker-build-amd64: ## AMD64용 Docker 이미지 빌드
	@echo "$(YELLOW)Building AMD64 Docker image...$(RESET)"
	@docker build --platform linux/amd64 -t $(DOCKER_IMAGE):amd64 .
	@echo "$(GREEN)✓ Built $(DOCKER_IMAGE):amd64$(RESET)"

docker-build-arm64: ## ARM64용 Docker 이미지 빌드
	@echo "$(YELLOW)Building ARM64 Docker image...$(RESET)"
	@docker build --platform linux/arm64 -t $(DOCKER_IMAGE):arm64 .
	@echo "$(GREEN)✓ Built $(DOCKER_IMAGE):arm64$(RESET)"

docker-build-darwin-amd64: ## macOS Intel용 Docker 이미지 빌드
	@echo "$(YELLOW)Building macOS Intel Docker image...$(RESET)"
	@docker build -f docker/Dockerfile.darwin-amd64 -t $(DOCKER_IMAGE):darwin-amd64 .
	@echo "$(GREEN)✓ Built $(DOCKER_IMAGE):darwin-amd64$(RESET)"

docker-build-darwin-arm64: ## macOS Apple Silicon용 Docker 이미지 빌드
	@echo "$(YELLOW)Building macOS Apple Silicon Docker image...$(RESET)"
	@docker build -f docker/Dockerfile.darwin-arm64 -t $(DOCKER_IMAGE):darwin-arm64 .
	@echo "$(GREEN)✓ Built $(DOCKER_IMAGE):darwin-arm64$(RESET)"

docker-build-all-arch: docker-build-amd64 docker-build-arm64 docker-build-darwin-amd64 docker-build-darwin-arm64 ## 모든 아키텍처용 Docker 이미지 빌드

# 개발 타겟
dev: ## 개발 모드로 애플리케이션 실행
	@echo "$(YELLOW)Starting development server...$(RESET)"
	@go run ./cmd/main.go

run: build ## 빌드 후 애플리케이션 실행
	@echo "$(YELLOW)Running $(APP_NAME)...$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

runWithSwagger: build swagger ## Swagger와 함께 빌드 후 애플리케이션 실행
	@echo "$(YELLOW)Running $(APP_NAME)...$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

# 코드 품질 타겟
lint: ## 린터 실행
	@echo "$(YELLOW)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
	fi

format: ## 코드 포맷팅
	@echo "$(YELLOW)Formatting code...$(RESET)"
	@go fmt ./...
	@go mod tidy
	@echo "$(GREEN)✓ Code formatted$(RESET)"

# Docker 타겟
docker-build: ## Docker 이미지 빌드
	@echo "$(YELLOW)Building Docker image...$(RESET)"
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✓ Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

docker-run: docker-build ## Docker 이미지 빌드 후 컨테이너 실행
	@echo "$(YELLOW)Running Docker container...$(RESET)"
	@docker run --rm -p 9091:9091 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose 타겟
compose-up: docker-build ## docker-compose로 서비스 시작
	@echo "$(YELLOW)Starting services with docker-compose...$(RESET)"
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "$(GREEN)✓ Services started$(RESET)"

compose-down: ## docker-compose로 서비스 중지
	@echo "$(YELLOW)Stopping services with docker-compose...$(RESET)"
	@docker-compose -f docker-compose.dev.yml down
	@echo "$(GREEN)✓ Services stopped$(RESET)"

compose-logs: ## docker-compose 로그 표시
	@docker-compose -f docker-compose.dev.yml logs -f

# 의존성 관리
deps: ## 의존성 다운로드
	@echo "$(YELLOW)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

deps-update: ## 의존성 업데이트
	@echo "$(YELLOW)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

# Swagger 타겟
swagger: ## Swagger 문서 생성
	@echo "$(YELLOW)Generating Swagger documentation...$(RESET)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/main.go -o docs/swagger --parseDependency --parseInternal; \
		echo "$(GREEN)✓ Swagger documentation generated$(RESET)"; \
	else \
		echo "$(RED)swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest$(RESET)"; \
	fi

# 정리 타겟
clean: ## 빌드 아티팩트 정리
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean completed$(RESET)"

clean-docker: ## Docker 이미지 및 컨테이너 정리
	@echo "$(YELLOW)Cleaning Docker artifacts...$(RESET)"
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker system prune -f
	@echo "$(GREEN)✓ Docker cleanup completed$(RESET)"

# 릴리스 타겟
release: clean deps lint test build-all ## 릴리스 준비
	@echo "$(GREEN)✓ Release ready$(RESET)"

# 버전 정보
version: ## 버전 정보 표시
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell go version)"

# 현재 상태 표시
status: ## 프로젝트 상태 표시
	@echo "$(BLUE)Project Status:$(RESET)"
	@echo "Git Branch: $(shell git branch --show-current)"
	@echo "Git Status: $(shell git status --porcelain | wc -l | tr -d ' ') files changed"
	@echo "Go Version: $(shell go version | cut -d' ' -f3)"
	@echo "Dependencies: $(shell go list -m all | wc -l | tr -d ' ') modules"
