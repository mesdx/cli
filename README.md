<div align="center">
  <img src="icon.png" alt="CodeIntelX" width="128" height="128">
  <h1>CodeIntelX</h1>
</div>

CodeIntelX is a tool which keeps track of your code and serves as a niche layer of code intelligence for Claude Code like tools.

## Installation

### 1. Install CodeIntelX Binary

Download the pre-built binary for your platform from the [latest release](https://github.com/codeintelx/cli/releases/latest).

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-arm64 -o codeintelx
chmod +x codeintelx
sudo mv codeintelx /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-amd64 -o codeintelx
chmod +x codeintelx
sudo mv codeintelx /usr/local/bin/
```

**Linux:**
```bash
# AMD64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-amd64 -o codeintelx

# ARM64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-arm64 -o codeintelx

chmod +x codeintelx
sudo mv codeintelx /usr/local/bin/
```

Verify installation:
```bash
codeintelx --version
```

### 2. Add to Claude Code

1. Open Claude Code → **Settings** → **Developer** → **Edit Config**

2. Add CodeIntelX to your `claude_desktop_config.json`:

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

**Using a specific workspace directory:**

If you want to point the MCP server at a specific workspace directory (useful when working with multiple projects), use the `--cwd` flag:

```json
{
  "mcpServers": {
    "codeintelx": {
      "command": "codeintelx",
      "args": ["mcp", "--cwd", "/path/to/your/workspace"]
    }
  }
}
```

3. Restart Claude Code to load the MCP server.

## Usage

### Initialize a Repository

```bash
cd /path/to/your/repo
codeintelx init
```

This will:
- Detect the repository root
- Prompt you to select source directories to index
- Create relevant configuration within repository, your repository data stays in repository.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
