# Contributing to Codeintelx CLI

Thank you for your interest in contributing to Codeintelx CLI! This document provides guidelines and instructions for developers.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Building

Build the CLI from source:

```bash
cd cli
go build ./cmd/codeintelx
```

The binary will be created in the current directory.

### Installing for Development

Install the CLI to your Go bin directory:

```bash
go install ./cmd/codeintelx
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

## Testing

Run all tests:

```bash
cd cli
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests for a specific package:

```bash
go test ./internal/repo
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

Configuration is stored as JSON in `.codeintelx/config.json` at the repository root.

### Database

SQLite database is stored at `.codeintelx/index.db` using the pure-Go driver (no CGO).

### MCP Server

The MCP server runs over stdio transport and exposes tools for code intelligence queries. Currently implements a stub `codeintelx.projectInfo` tool.

## Future Work

- File-hash based incremental indexing
- Language-aware code parsing (Go, TypeScript, JavaScript, etc.)
- Dependency graph extraction and storage
- Additional MCP tools for symbol search, call graph queries, etc.

## Questions?

If you have questions or need help, please open an issue on the repository.
