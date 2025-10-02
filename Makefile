# Go Build Configuration
GO_VERSION ?= $(shell ./scripts/get-go-version.sh)
BINARY_NAME = dev-stack
BUILD_DIR = build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X github.com/isaacgarza/dev-stack/internal/cli.version=$(VERSION) \
                   -X github.com/isaacgarza/dev-stack/internal/cli.commit=$(COMMIT) \
                   -X github.com/isaacgarza/dev-stack/internal/cli.date=$(BUILD_DATE)"



# Platform detection
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: help build build-all install clean test test-go lint lint-go deps deps-go docs

## Default target
all: build

## Go targets
build: deps-go ## Build the Go binary for current platform
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/dev-stack

build-all: deps-go ## Build binaries for all supported platforms
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/dev-stack
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/dev-stack
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/dev-stack
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/dev-stack
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/dev-stack
	@echo "Binaries built in $(BUILD_DIR)/"

install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/dev-stack

deps-go: ## Download Go dependencies
	@echo "Downloading Go dependencies..."
	go mod download
	go mod tidy

test-go: deps-go ## Run Go tests
	@echo "Running Go tests..."
	go test -v -race -coverprofile=coverage.out $(shell go list ./... | grep -v '/tests/integration')

test-go-integration: build ## Run Go integration tests
	@echo "Running Go integration tests..."
	@if find ./tests -name "*_test.go" 2>/dev/null | grep -q .; then \
		cd tests/integration && go test -v -tags=integration .; \
	else \
		echo "No Go integration tests found in ./tests/ directory, skipping..."; \
	fi

lint-go: ## Run Go linting
	@echo "Running Go linting..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint v2..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.5.0; \
	fi
	$(shell go env GOPATH)/bin/golangci-lint run ./...

fmt-go: ## Format Go code
	@echo "Formatting Go code..."
	@if [ ! -f $(shell go env GOPATH)/bin/goimports ]; then \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	fi
	go fmt ./...
	$(shell go env GOPATH)/bin/goimports -w .

vet-go: ## Run Go vet
	@echo "Running go vet..."
	go vet ./...



docs: build ## Generate documentation from YAML manifests
	@echo "Generating documentation from YAML manifests..."
	./$(BUILD_DIR)/$(BINARY_NAME) docs --verbose

## Combined targets
deps: deps-go ## Download all dependencies

test: test-go ## Run all tests

lint: lint-go ## Run all linting

## Development targets
dev: build ## Build and run in development mode
	@echo "Running $(BINARY_NAME) in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

watch: ## Watch for changes and rebuild (requires entr)
	@if ! command -v entr >/dev/null 2>&1; then \
		echo "Error: entr is not installed. Install it with your package manager."; \
		exit 1; \
	fi
	@echo "Watching for changes... (requires entr)"
	find . -name "*.go" | entr -r make build

## Release targets
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@git diff --exit-code || (echo "Error: Uncommitted changes found"; exit 1)
	@git diff --cached --exit-code || (echo "Error: Staged changes found"; exit 1)
	@echo "Ready for release!"

## Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t dev-stack:$(VERSION) .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it dev-stack:$(VERSION)

## Cleanup targets
clean: ## Remove build artifacts and generated files
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out
	rm -f lint.log

	go clean

clean-all: clean ## Remove all generated files including dependencies
	go clean -modcache
	rm -rf vendor/

## Version and info targets
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(shell go version)"

info: ## Show build information
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Build Directory: $(BUILD_DIR)"
	@echo "Target OS/Arch: $(GOOS)/$(GOARCH)"
	@echo "Go Version: $(shell go version)"
	@echo "Git Status: $(shell git status --porcelain | wc -l) uncommitted changes"

## Version management targets
sync-version: ## Sync Go version across all config files
	@echo "Syncing Go version across configuration files..."
	./scripts/sync-go-version.sh --fix

check-version: ## Check if Go versions are consistent across config files
	@echo "Checking Go version consistency..."
	./scripts/sync-go-version.sh --check

show-go-version: ## Show the current Go version from .go-version
	@echo "Current Go version: $(GO_VERSION)"
	@echo "Matrix versions for CI: $(shell ./scripts/get-go-version.sh --github-matrix)"

## Release configuration targets
generate-release-configs: ## Generate all release configuration files from central config
	@echo "Generating release configuration files..."
	@./scripts/generate-release-configs.sh

validate-release-configs: ## Validate release configuration files are up to date
	@echo "Validating release configuration files..."
	@./scripts/generate-release-configs.sh
	@if git diff --exit-code .commitlintrc.json .release-please-config.json; then \
		echo "✅ Release configuration files are up to date"; \
	else \
		echo "❌ Release configuration files are out of date"; \
		echo "Run 'make generate-release-configs' to update them"; \
		exit 1; \
	fi



check-release-deps: ## Check if release dependencies are installed
	@echo "Checking release dependencies..."
	@if ! command -v yq >/dev/null 2>&1; then \
		echo "❌ yq is required but not installed."; \
		echo "Install with: brew install yq (macOS) or go install github.com/mikefarah/yq/v4@latest"; \
		exit 1; \
	fi
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "❌ jq is required but not installed."; \
		echo "Install with: brew install jq (macOS) or apt-get install jq (Ubuntu)"; \
		exit 1; \
	fi
	@echo "✅ All release dependencies are installed"

release-setup: check-release-deps generate-release-configs ## Complete release automation setup
	@echo "🚀 Release automation setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Commit the generated configuration files"
	@echo "  2. Push to main branch to enable Release Please"
	@echo "  3. Use conventional commits for automatic releases"

## Help target
help: ## Show this help message
	@echo "dev-stack Makefile"
	@echo ""
	@echo "Go Targets (Primary):"
	@echo "  build          - Build the Go binary for current platform"
	@echo "  build-all      - Build binaries for all supported platforms"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  test-go        - Run Go tests"
	@echo "  lint-go        - Run Go linting"
	@echo "  fmt-go         - Format Go code"
	@echo "  deps-go        - Download Go dependencies"
	@echo ""
	@echo "Development Targets:"
	@echo "  dev            - Build and run in development mode"
	@echo "  watch          - Watch for changes and rebuild"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo ""
	@echo "Documentation Targets:"
	@echo "  docs           - Generate documentation from YAML manifests"
	@echo ""

	@echo ""
	@echo "Combined Targets:"
	@echo "  test           - Run all tests"
	@echo "  lint           - Run all linting"
	@echo "  deps           - Download all dependencies"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean          - Remove build artifacts"
	@echo "  clean-all      - Remove all generated files"
	@echo "  version        - Show version information"
	@echo "  info           - Show build information"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Version Management:"
	@echo "  sync-version   - Sync Go version across all config files"
	@echo "  check-version  - Check Go version consistency"
	@echo "  show-go-version - Show current Go version and CI matrix"
	@echo ""
	@echo "Release Management:"
	@echo "  generate-release-configs - Generate release configuration files"
	@echo "  validate-release-configs - Validate configuration files are current"
	@echo "  check-release-deps      - Check if release dependencies are installed"
	@echo "  release-setup           - Complete release automation setup"
	@echo ""
	@echo "Usage Examples:"
	@echo "  make build                    # Build for current platform"
	@echo "  make build-all               # Build for all platforms"
	@echo "  make test                    # Run tests"
	@echo "  make GOOS=linux GOARCH=amd64 build  # Cross-compile"
	@echo "  make release-setup           # Set up release automation"
