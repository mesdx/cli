# Codeintelx CLI

Code intelligence CLI and MCP server for indexing codebases and exposing code intelligence via Model Context Protocol.

## Installation

### From GitHub Releases (Recommended)

Download the pre-built binary for your platform from the [latest release](https://github.com/codeintelx/cli/releases/latest).

#### macOS (M1/Apple Silicon)

```bash
# Download for macOS ARM64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-arm64 -o codeintelx

# Make it executable
chmod +x codeintelx

# Move to a directory in your PATH
sudo mv codeintelx /usr/local/bin/
```

#### macOS (Intel)

```bash
# Download for macOS AMD64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-amd64 -o codeintelx

# Make it executable
chmod +x codeintelx

# Move to a directory in your PATH
sudo mv codeintelx /usr/local/bin/
```

#### Linux

```bash
# Download for Linux (AMD64)
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-amd64 -o codeintelx

# Or for ARM64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-arm64 -o codeintelx

# Make it executable
chmod +x codeintelx

# Move to a directory in your PATH
sudo mv codeintelx /usr/local/bin/
```

#### Verify Installation

```bash
codeintelx --version
```

Or check the help:

```bash
codeintelx --help
```

### From Source

If you prefer to build from source:

```bash
git clone https://github.com/codeintelx/cli.git
cd cli
go build ./cmd/codeintelx
```

The binary will be created in the current directory. You can install it globally:

```bash
go install ./cmd/codeintelx
```

## Usage

### Initialize a Repository

Navigate to your repository root and run:

```bash
codeintelx init
```

This will:
1. Detect the repository root (looks for `.git` directory or uses current directory)
2. Prompt you to select source directories to index
3. Create `.codeintelx/` directory with:
   - `config.json` - Configuration file
   - `index.db` - SQLite database for code intelligence data
4. Optionally update `.gitignore` and `.dockerignore` to exclude `.codeintelx/`

### Start MCP Server

To start the MCP server for Claude Code integration:

```bash
codeintelx mcp
```

The server runs over stdio and exposes tools for code intelligence queries.

## Claude Code Integration

Add the following to your Claude Code MCP configuration (typically in `~/.config/claude-desktop/mcp.json` or similar):

```json
{
  "mcpServers": {
    "codeintelx": {
      "command": "codeintelx",
      "args": ["mcp"]
    }
  }
}
```

Make sure `codeintelx` is in your PATH, or use the full path to the binary.

## Project Structure

When you initialize a repository, codeintelx creates the following structure:

```
.codeintelx/
├── config.json    # Repository configuration
└── index.db       # SQLite database for code intelligence
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
