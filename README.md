<div align="center">
  <img src="icon.png" alt="MesDX" width="128" height="128">
  <h1>MesDX</h1>
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

### Quick Install (Recommended)

Use our install scripts that handle everything automatically:

**macOS (Apple Silicon):**
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-darwin-arm64.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-darwin-arm64.sh | bash -s v0.3.0
```

**macOS (Intel):**
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-darwin-amd64.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-darwin-amd64.sh | bash -s v0.3.0
```

**Linux (AMD64):**
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-linux-amd64.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-linux-amd64.sh | bash -s v0.3.0
```

**Linux (ARM64):**
```bash
# Install latest version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-linux-arm64.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/mesdx/cli/main/scripts/install-linux-arm64.sh | bash -s v0.3.0
```

The install script will:
- Download the binary to `~/.local/bin/mesdx`
- Download parser libraries to `~/.local/lib/mesdx/parsers`
- Make the binary executable

### Add to PATH

Make sure `~/.local/bin` is in your PATH:

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

### Verify Installation

```bash
mesdx --version
```

### Manual Installation

If you prefer manual installation:

**Latest version:**
1. Download the binary for your platform from [latest release](https://github.com/mesdx/cli/releases/latest)
2. Download the parser tarball for your platform
3. Extract binary to `~/.local/bin/`
4. Extract parsers to `~/.local/lib/mesdx/parsers/`
5. Make the binary executable: `chmod +x ~/.local/bin/mesdx`

**Specific version:**
1. Visit the [releases page](https://github.com/mesdx/cli/releases) and find your desired version
2. Download the binary for your platform (e.g., `mesdx-darwin-arm64` for version `v0.3.0`)
3. Download the parser tarball for your platform (e.g., `mesdx-parsers-darwin-arm64.tar.gz`)
4. Extract binary to `~/.local/bin/mesdx`
5. Extract parsers to `~/.local/lib/mesdx/parsers/`
6. Make the binary executable: `chmod +x ~/.local/bin/mesdx`

**Example for version v0.3.0 on macOS (Apple Silicon):**
```bash
install_dir="$HOME/.local/bin"
parser_dir="$HOME/.local/lib/mesdx/parsers"
mkdir -p "$install_dir" "$parser_dir"

# Download binary
curl -L https://github.com/mesdx/cli/releases/download/v0.3.0/mesdx-darwin-arm64 -o "$install_dir/mesdx"
chmod +x "$install_dir/mesdx"

# Download and extract parsers
curl -L https://github.com/mesdx/cli/releases/download/v0.3.0/mesdx-parsers-darwin-arm64.tar.gz | tar xz -C "$parser_dir" --strip-components=1
```

### Custom Parser Directory

By default, MesDX looks for parsers in `~/.local/lib/mesdx/parsers`. To use a custom location:

```bash
export MESDX_PARSER_DIR=/path/to/your/parsers
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
