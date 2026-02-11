.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Initialize grammar submodules
	@echo "Initializing grammar submodules..."
	@bash scripts/setup-grammars.sh

.PHONY: build
build: setup ## Build everything (parsers + binary)
	@bash scripts/build-all.sh

.PHONY: test
test: build ## Run tests
	@export MESDX_PARSER_DIR=$$(pwd)/dist/parsers && go test -v ./...

.PHONY: test-quick
test-quick: ## Run tests without rebuilding (assumes already built)
	@export MESDX_PARSER_DIR=$$(pwd)/dist/parsers && go test -v ./...

.PHONY: install
install: build ## Install mesdx binary and parsers to ~/.local
	@echo "Installing mesdx to ~/.local/bin..."
	@mkdir -p $$HOME/.local/bin
	@cp dist/mesdx $$HOME/.local/bin/
	@echo "Installing parsers to ~/.local/lib/mesdx/parsers..."
	@mkdir -p $$HOME/.local/lib/mesdx/parsers
	@cp dist/parsers/* $$HOME/.local/lib/mesdx/parsers/
	@echo "✓ Installation complete"
	@echo ""
	@echo "Make sure ~/.local/bin is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf dist/
	@echo "✓ Cleaned build artifacts"

.PHONY: clean-all
clean-all: clean ## Clean everything including grammar submodules
	@git submodule deinit -f third_party/
	@rm -rf third_party/
	@echo "✓ Cleaned all artifacts and submodules"

.PHONY: lint
lint: ## Run linter
	@golangci-lint run

.PHONY: fmt
fmt: ## Format code
	@go fmt ./...
	@gofmt -s -w .

.DEFAULT_GOAL := help
