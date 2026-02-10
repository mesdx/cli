# Build Scripts

This directory contains build, setup, and installation scripts for MesDX.

## Build Scripts

### `build-all.sh`
**One-stop build script** - builds everything in correct order.

**Usage:**
```bash
bash scripts/build-all.sh
```

**What it does:**
1. Downloads tree-sitter API headers
2. Initializes grammar submodules
3. Builds parser libraries
4. Builds Go binary

**Output:**
- `dist/mesdx` - the binary
- `dist/parsers/` - parser libraries

### `setup-tree-sitter.sh`
Downloads the tree-sitter API header file needed for CGO compilation.

### `setup-grammars.sh`
Initializes tree-sitter grammar submodules from `third_party/`.

### `build-parsers.sh`
Compiles tree-sitter parser libraries for all supported languages. Creates:
- macOS: `libtree-sitter-*.dylib`
- Linux: `libtree-sitter-*.so`

## Installation Scripts

Platform-specific install scripts that download and install MesDX:

### `install-darwin-arm64.sh`
Installs MesDX on macOS (Apple Silicon).

### `install-darwin-amd64.sh`
Installs MesDX on macOS (Intel).

### `install-linux-amd64.sh`
Installs MesDX on Linux (AMD64).

### `install-linux-arm64.sh`
Installs MesDX on Linux (ARM64).

**Usage:**
```bash
# Install latest release
bash scripts/install-darwin-arm64.sh

# Install specific version
bash scripts/install-darwin-arm64.sh v0.3.0
```

Each script:
- Downloads the binary to `~/.local/bin/mesdx`
- Downloads and extracts parser libraries to `~/.local/lib/mesdx/parsers/`
- Makes the binary executable

## Quick Start

For development:
```bash
make build    # Runs build-all.sh
make test     # Runs tests
make install  # Installs to ~/.local
```

## CI/CD

The GitHub Actions workflow (`release.yml`) uses these scripts:

1. **Build stage**: Runs `build-all.sh` on each platform
2. **Release stage**: Packages binaries and parser tarballs
3. **Install scripts**: Users download and run platform-specific install scripts
