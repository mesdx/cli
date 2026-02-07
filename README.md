<div align="center">
  <img src="icon.png" alt="CodeIntelX" width="128" height="128">
  <h1>CodeIntelX</h1>
</div>

CodeIntelX is a **local-first code intelligence MCP server** that indexes your repository and exposes reliable, structured navigation tools to agents like **Claude Code** (and other MCP clients). 🔍🧭

## Why CodeIntelX (MCP)?

- **🔎 Precise symbol navigation**: resolve definitions/usages from your *actual codebase* instead of guessing.
- **🧩 Impact analysis**: inspect inbound/outbound dependencies for safer refactors.
- **⚡ Always up to date**: runs a background file watcher and reconciles indexes as files change.
- **🗂️ Repo-scoped & local**: stores data in your repo’s `.codeintelx/` (no system-wide daemon, no sudo required).
- **🧠 Long-term “memory” for agents (optional)**: keep durable, *searchable* markdown knowledge alongside the repo—design decisions, gotchas, runbooks, TODOs, and context that shouldn’t live in code comments. It supports project-wide or file-scoped notes, symbol references, and fast text search so an agent can regain context across sessions.

### MCP tools you get

- **📦 `codeintelx.projectInfo`**: repo root, configured source roots, DB path.
- **🧭 `codeintelx.goToDefinition`**: go-to-definition by cursor (`filePath + line + column`) or by `symbolName`.
- **🔁 `codeintelx.findUsages`**: find and score usages across the codebase (cursor-based or name-based).
- **🧩 `codeintelx.dependencyGraph`**: inbound/outbound symbol dependencies (great for refactor risk checks).

Supported languages: **Go, Java, Rust, Python, TypeScript, JavaScript**.

## Installation

### 1. Install CodeIntelX Binary

Download the pre-built binary for your platform from the [latest release](https://github.com/codeintelx/cli/releases/latest).

**macOS (Apple Silicon):**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-arm64 -o "$install_dir/codeintelx"
chmod +x "$install_dir/codeintelx"
```

**macOS (Intel):**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-darwin-amd64 -o "$install_dir/codeintelx"
chmod +x "$install_dir/codeintelx"
```

**Linux:**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"

# Choose ONE:
# AMD64
curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-amd64 -o "$install_dir/codeintelx"

# ARM64
# curl -L https://github.com/codeintelx/cli/releases/latest/download/codeintelx-linux-arm64 -o "$install_dir/codeintelx"

chmod +x "$install_dir/codeintelx"
```

Make sure `~/.local/bin` is on your `PATH`:

**zsh:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**bash:**
```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Verify installation:
```bash
codeintelx --version
```

### 2. Add to Claude Code

Claude Code can add MCP servers via the CLI (recommended). See the official [Claude Code MCP docs](https://code.claude.com/docs/en/mcp).

Add CodeIntelX as a **stdio** MCP server:

```bash
claude mcp add codeintelx --transport stdio -- codeintelx mcp
```

**Using a specific workspace directory:**

If you want to point the MCP server at a specific workspace directory (useful when working with multiple projects), use the `--cwd` flag:

```bash
claude mcp add codeintelx --transport stdio -- codeintelx mcp --cwd /path/to/your/workspace
```

Restart Claude Code to load the MCP server.

## Usage

### Initialize a Repository

```bash
cd /path/to/your/repo
codeintelx init
```

This will:
- **🗂️ Detect** the repository root and create `.codeintelx/config.json` + a local DB in `.codeintelx/`
- **🔎 Index** the source directories you select (symbols + references)
- **🧠 Create/index** a repo-relative markdown “memory” directory (default: `docs/codeintelx-memory`)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.
