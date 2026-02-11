# Scripts

This directory contains helper scripts for MesDX development and distribution.

## Available Scripts

Currently, all build and installation is handled through the `Makefile` and standard Go tooling. No additional scripts are needed.

## Historical Note

Previous versions required separate parser library builds and distribution. As of v0.2.0, all tree-sitter parsers are **statically compiled** into the binary using official Go bindings. This eliminated the need for:

- `build-parsers.sh` - Parser library compilation
- `setup-grammars.sh` - Git submodule management  
- `init-submodules.sh` - Submodule initialization
- `install-*.sh` - Platform-specific installers with parser libraries
- `build-all.sh` - Unified build orchestration

## Current Build Process

Simply run:

```bash
make build
```

This produces a single, self-contained binary in `dist/mesdx` with all parsers included.
