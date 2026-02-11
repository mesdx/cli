# Contributing to MesDX

Thank you for your interest in contributing to MesDX!

## Development Setup

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Git** - For version control
- **golangci-lint** (optional) - For linting

### Quick Start

1. **Clone the repository**

```bash
git clone https://github.com/mesdx/cli.git
cd cli
```

2. **Build the project**

```bash
make build
```

This compiles the binary to `dist/mesdx` with all tree-sitter parsers statically linked.

3. **Run tests**

```bash
make test
```

4. **Install locally for testing**

```bash
make install
```

This installs to `~/.local/bin/mesdx`.

## Project Structure

```
cli/
├── cmd/mesdx/           # Main CLI entry point
├── internal/
│   ├── cli/             # CLI commands
│   ├── db/              # SQLite database layer
│   ├── indexer/         # Symbol indexing engine
│   ├── treesitter/      # Tree-sitter parsing (static bindings)
│   ├── memory/          # Memory/notes system
│   ├── search/          # FTS5 search
│   └── selfupdate/      # Auto-update functionality
├── Makefile             # Build targets
└── go.mod               # Go dependencies
```

## Architecture

### Parsing

MesDX uses **tree-sitter** for precise syntax parsing. All language parsers are **statically compiled** into the binary using official Go bindings:

- `github.com/tree-sitter/tree-sitter-go/bindings/go`
- `github.com/tree-sitter/tree-sitter-java/bindings/go`
- `github.com/tree-sitter/tree-sitter-rust/bindings/go`
- `github.com/tree-sitter/tree-sitter-python/bindings/go`
- `github.com/tree-sitter/tree-sitter-javascript/bindings/go`
- `github.com/tree-sitter/tree-sitter-typescript/bindings/go`

Query files (`.scm`) define patterns for extracting symbols and references:
- `internal/treesitter/queries/*.scm`

### Adding a New Language

To add support for a new language:

1. **Add the Go binding dependency**

```bash
go get github.com/tree-sitter/tree-sitter-LANGUAGE/bindings/go@latest
```

2. **Update `internal/treesitter/loader.go`**

Add an import and entry to `languageMap`:

```go
import tree_sitter_LANGUAGE "github.com/tree-sitter/tree-sitter-LANGUAGE/bindings/go"

var languageMap = map[string]func() unsafe.Pointer{
    // ... existing languages
    "LANGUAGE": tree_sitter_LANGUAGE.Language,
}
```

3. **Create a query file**

Create `internal/treesitter/queries/LANGUAGE.scm` with tree-sitter queries for:
- Symbol definitions (`@def.*`)
- References (`@ref.*`)

See existing `.scm` files for examples.

4. **Update `RequiredLanguages()` in `loader.go`**

5. **Add tests** in `internal/treesitter/extractor_test.go`

6. **Update language detection** in `internal/indexer/indexer.go` if needed

### Indexing Flow

1. **Discovery**: `indexer.Indexer.IndexAll()` walks the file tree
2. **Parsing**: Each file is parsed by the appropriate tree-sitter parser
3. **Extraction**: Query patterns extract symbols and references
4. **Storage**: Symbols/refs are stored in SQLite (`.mesdx/mesdx.db`)
5. **FTS**: Full-text search index is updated for memories

### Database Schema

See `internal/db/migrations.go` for the full schema. Key tables:

- `symbols` - Symbol definitions (functions, classes, etc.)
- `refs` - Symbol references (calls, uses)
- `file_hashes` - Track file changes for incremental updates
- `memory_elements` - Project notes
- `memory_fts` - Full-text search index

## Testing

### Run All Tests

```bash
make test
```

### Run Specific Tests

```bash
go test -v ./internal/indexer -run TestGoParser
```

### Test Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Style

- Run `go fmt` before committing
- Follow standard Go conventions
- Use `golangci-lint` for additional checks

```bash
make fmt
make lint
```

## Making Changes

1. **Create a feature branch**

```bash
git checkout -b feature/my-feature
```

2. **Make your changes**

- Write tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

3. **Test thoroughly**

```bash
make test
make build
./dist/mesdx init # Test on a real repository
```

4. **Commit**

Use clear, descriptive commit messages:

```bash
git add .
git commit -m "Add support for Ruby parsing"
```

5. **Push and create a PR**

```bash
git push origin feature/my-feature
```

Then create a Pull Request on GitHub.

## Release Process

Releases are automated via GitHub Actions when a tag is pushed:

1. Maintainer creates a new tag: `git tag v0.x.0`
2. Push tag: `git push origin v0.x.0`
3. CI builds binaries for all platforms
4. Binaries are attached to the GitHub release

**Note**: Tags containing `rc` or `test` are marked as prereleases.

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/mesdx/cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mesdx/cli/discussions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
