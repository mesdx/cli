# MesDX - Local-First Code Intelligence

[![CI](https://github.com/mesdx/cli/actions/workflows/test.yml/badge.svg)](https://github.com/mesdx/cli/actions/workflows/test.yml)
[![Release](https://github.com/mesdx/cli/actions/workflows/release.yml/badge.svg)](https://github.com/mesdx/cli/actions/workflows/release.yml)

MesDX provides precise, local-first code navigation and analysis via the Model Context Protocol (MCP).

## Features

- **Precise navigation**: Go to definition, find usages with tree-sitter parsing
- **Dependency analysis**: Understand symbol relationships and refactoring impact
- **MCP server mode**: Integrate with Claude Desktop and other MCP clients
- **Memory system**: Persistent project notes and context
- **Multi-language**: Go, Java, Rust, Python, TypeScript, JavaScript

## Quick Install

### macOS (Apple Silicon)

```bash
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-darwin-arm64 -o mesdx
chmod +x mesdx
sudo mv mesdx /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-darwin-amd64 -o mesdx
chmod +x mesdx
sudo mv mesdx /usr/local/bin/
```

### Linux (x86_64)

```bash
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-linux-amd64 -o mesdx
chmod +x mesdx
sudo mv mesdx /usr/local/bin/
```

### Linux (ARM64)

```bash
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-linux-arm64 -o mesdx
chmod +x mesdx
sudo mv mesdx /usr/local/bin/
```

### Install Specific Version

Replace `latest` with a specific version tag (e.g., `v0.2.0`):

```bash
# macOS Apple Silicon
curl -L https://github.com/mesdx/cli/releases/download/v0.2.0/mesdx-darwin-arm64 -o mesdx
chmod +x mesdx
sudo mv mesdx /usr/local/bin/
```

### Verify Installation

```bash
mesdx --version
```

## Usage

### Initialize a Repository

```bash
cd /path/to/your/project
mesdx init
```

This creates a `.mesdx/` directory with an SQLite database for symbols and references.

### Re-index After Changes

```bash
mesdx init --force
```

### MCP Server Mode

Add to Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mesdx": {
      "command": "mesdx",
      "args": ["mcp", "--cwd", "/absolute/path/to/your/repo"]
    }
  }
}
```

Then restart Claude Desktop. You can now use MesDX tools like `mesdx.goToDefinition`, `mesdx.findUsages`, etc.

## Available Commands

```bash
mesdx --help
```

- `mesdx init` - Initialize/index a repository
- `mesdx mcp` - Start MCP server
- `mesdx version` - Show version

## Build from Source

### Prerequisites

- Go 1.21 or later
- Git

### Build

```bash
git clone https://github.com/mesdx/cli.git
cd cli
make build
```

The binary will be in `dist/mesdx`.

### Install Locally

```bash
make install
```

This installs to `~/.local/bin/mesdx`.

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## Architecture

MesDX uses:
- **Tree-sitter** for precise syntax parsing (all languages statically compiled)
- **SQLite** for local symbol/reference storage
- **MCP** for Claude Desktop integration
- **FTS5** for full-text memory search

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [Tree-sitter](https://tree-sitter.github.io/)
