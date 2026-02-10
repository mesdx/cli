# Contributing to MesDX CLI

Thank you for your interest in contributing to MesDX CLI! This document provides guidelines and instructions for developers.

## Development Setup

### Prerequisites

- Go 1.25 or later
- Git
- C compiler (gcc/clang) for building tree-sitter parsers
- Make (optional, for convenience)

### Quick Start

```bash
# One command to build everything
make build

# Run tests
make test

# Install locally
make install
```

That's it! The `make build` command will:
1. Download tree-sitter headers
2. Initialize grammar submodules
3. Build parser libraries
4. Build the Go binary

### Manual Setup

If you don't have Make:

```bash
# Build everything (headers + parsers + binary)
bash scripts/build-all.sh

# Run tests
export MESDX_PARSER_DIR=$(pwd)/dist/parsers
go test -v ./...
```

### Install for Development

```bash
make install
```

This installs:
- Binary to `~/.local/bin/mesdx`
- Parsers to `~/.local/lib/mesdx/parsers/`

Make sure `~/.local/bin` is in your PATH.

## Testing

Run all tests:

```bash
make test  # Builds everything if needed, then runs tests
```

Run tests without rebuilding (faster):

```bash
make test-quick
```

Run tests for a specific package:

```bash
export MESDX_PARSER_DIR=$(pwd)/dist/parsers
go test -v ./internal/repo
```

### Adding Dependencies

Add a new dependency:

```bash
go get <package-path>
```

Update dependencies:

```bash
go mod tidy
```

## Code Style

- Follow standard Go formatting: `gofmt` or `goimports`
- Use `golangci-lint` for linting (if configured)
- Write clear, descriptive commit messages
- Add comments for exported functions and types

## Architecture Notes

### Repository Detection

The `internal/repo` package handles repository root detection:
- First checks for `.git` directory in current or parent directories
- Falls back to current working directory if no `.git` found

### Configuration

Configuration is stored as JSON in `.mesdx/config.json` at the repository root.

### Database

SQLite database is stored at `.mesdx/index.db` using the pure-Go driver (no CGO).

### Parsing

MesDX uses tree-sitter for accurate, language-aware parsing:

- Parser libraries are loaded dynamically at runtime via `dlopen`
- Each language has its own `.dylib` (macOS) or `.so` (Linux) file
- Query files in `internal/treesitter/queries/*.scm` define how to extract symbols and references
- The `internal/treesitter` package handles library loading and parsing
- Grammar sources are tracked as git submodules in `third_party/`

To add support for a new language:
1. Add the grammar as a submodule in `.gitmodules`
2. Update `scripts/build-parsers.sh` to build it
3. Create a query file in `internal/treesitter/queries/<lang>.scm`
4. Update `internal/treesitter/extractor.go` to include the new language
5. Add the language to `internal/indexer/lang.go` and the parser registry

### MCP Server

The MCP server runs over stdio transport and exposes tools for code intelligence queries.