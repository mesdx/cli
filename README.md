<div align="center">
  <img src="icon.png" alt="MesDX" width="128" height="128">
  <h1>MesDX</h1>
  
  [![Test](https://github.com/mesdx/cli/actions/workflows/test.yml/badge.svg)](https://github.com/mesdx/cli/actions/workflows/test.yml)
  [![Release](https://github.com/mesdx/cli/actions/workflows/release.yml/badge.svg)](https://github.com/mesdx/cli/actions/workflows/release.yml)
</div>

MesDX is a **local-first code intelligence MCP server** that indexes your repository and exposes reliable, structured navigation tools to agents like **Claude Code** (and other MCP clients). 🔍🧭

## Why MesDX (MCP)?

- **🔎 Precise symbol navigation**: resolve definitions/usages from your *actual codebase* instead of guessing.
- **🧩 Impact analysis**: inspect inbound/outbound dependencies for safer refactors.
- **⚡ Always up to date**: runs a background file watcher and reconciles indexes as files change.
- **🗂️ Repo-scoped & local**: stores data in your repo’s `.mesdx/` (no system-wide daemon, no sudo required).
- **🧠 Long-term “memory” for agents (optional)**: keep durable, *searchable* markdown knowledge alongside the repo—design decisions, gotchas, runbooks, TODOs, and context that shouldn’t live in code comments. It supports project-wide or file-scoped notes, symbol references, and fast text search so an agent can regain context across sessions.

### MCP tools you get

- **📦 `mesdx.projectInfo`**: repo root, configured source roots, DB path.
- **🧭 `mesdx.goToDefinition`**: go-to-definition by cursor (`filePath + line + column`) or by `symbolName`.
- **🔁 `mesdx.findUsages`**: find and score usages across the codebase (cursor-based or name-based).
- **🧩 `mesdx.dependencyGraph`**: inbound/outbound symbol dependencies (great for refactor risk checks).

Supported languages: **Go, Java, Rust, Python, TypeScript, JavaScript**.

## Installation

### 1. Install MesDX Binary

Download the pre-built binary for your platform from the [latest release](https://github.com/mesdx/cli/releases/latest).

**macOS (Apple Silicon):**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-darwin-arm64 -o "$install_dir/mesdx"
chmod +x "$install_dir/mesdx"
```

**macOS (Intel):**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-darwin-amd64 -o "$install_dir/mesdx"
chmod +x "$install_dir/mesdx"
```

**Linux:**
```bash
install_dir="$HOME/.local/bin"
mkdir -p "$install_dir"

# Choose ONE:
# AMD64
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-linux-amd64 -o "$install_dir/mesdx"

# ARM64
# curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-linux-arm64 -o "$install_dir/mesdx"

chmod +x "$install_dir/mesdx"
```

Make sure `~/.local/bin` is on your `PATH`:

**zsh:**
```bash
echo -e '\nexport PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**bash:**
```bash
echo -e '\nexport PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

Verify installation:
```bash
mesdx --version
```

**Optional: Verify checksum**

For security, you can verify the SHA256 checksum:

```bash
# Download the checksum file
curl -L https://github.com/mesdx/cli/releases/latest/download/mesdx-darwin-arm64.sha256 -o /tmp/mesdx.sha256

# Verify (macOS)
cd "$install_dir" && shasum -a 256 -c /tmp/mesdx.sha256

# Verify (Linux)
cd "$install_dir" && sha256sum -c /tmp/mesdx.sha256
```

### 2. Add to Claude Code

Claude Code can add MCP servers via the CLI (recommended). See the official [Claude Code MCP docs](https://code.claude.com/docs/en/mcp).

Add MesDX as a **stdio** MCP server:

```bash
claude mcp add mesdx --transport stdio -- mesdx mcp
```

**Using a specific workspace directory:**

If you want to point the MCP server at a specific workspace directory (useful when working with multiple projects), use the `--cwd` flag:

```bash
claude mcp add mesdx --transport stdio -- mesdx mcp --cwd /path/to/your/workspace
```

Restart Claude Code to load the MCP server.

## Usage

### Initialize a Repository

```bash
cd /path/to/your/repo
mesdx init
```

This will:
- **🗂️ Detect** the repository root and create `.mesdx/config.json` + a local DB in `.mesdx/`
- **🔎 Index** the source directories you select (symbols + references)
- **🧠 Create/index** a repo-relative markdown “memory” directory (default: `docs/mesdx-memory`)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## License

This project is licensed under AGPL-3.0.

If you modify and run it as a network service, you must publish your changes.
This ensures improvements remain open for everyone.

See [`LICENSE`](LICENSE).
