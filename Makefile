.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the binary
	@go build -o dist/mesdx ./cmd/mesdx
	@echo "✅ Build complete: dist/mesdx"

.PHONY: test
test: ## Run tests
	@go test -v ./...

.PHONY: test-quick
test-quick: ## Run tests (alias for test)
	@go test -v ./...

.PHONY: install
install: build ## Install mesdx binary to ~/.local/bin
	@echo "Installing mesdx to ~/.local/bin..."
	@mkdir -p $$HOME/.local/bin
	@cp dist/mesdx $$HOME/.local/bin/
	@echo "✅ Installation complete"
	@echo ""
	@echo "Make sure ~/.local/bin is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf dist/
	@echo "✓ Cleaned build artifacts"

.PHONY: lint
lint: ## Run linter
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@go fmt ./...
	@gofmt -s -w .

.DEFAULT_GOAL := help
