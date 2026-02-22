# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- External source dependency graphs: analyze and visualize cross-repo and external library relationships
- Deeper relation analysis: inbound/outbound dependency scoring now traverses transitive edges for richer impact assessment

### Removed
- Dropped support for Linux 386 and Windows builds to simplify distribution

### Changed
- Migrated to the official `go-tree-sitter` bindings for improved parser stability
- Simplified to a single-binary distribution with statically linked tree-sitter parsers (no external grammar shared libraries required)

## [0.2.0] - 2026-02-09

### Added
- MCP prompts support (`mesdx.skill.*` workflow prompts for bugfix, refactoring, feature development, and security analysis)

### Changed
- Switched license back to AGPL-3.0

## [0.1.3] - 2026-02-09

### Added
- AGPL-3.0 license file

### Changed
- Project renamed from `codeintelx` to `mesdx`
- Memory index improvements for faster full-text search across notes

## [0.1.2] - 2026-02-07

### Fixed
- Self-update binary download and version detection

## [0.1.1] - 2026-02-07

### Fixed
- Release pipeline and artifact upload corrections

## [0.1.0] - 2026-02-07

### Added
- Initial release
- Local-first code intelligence MCP server
- Symbol navigation: `goToDefinition`, `findUsages`
- Dependency graph analysis with impact scoring
- Persistent memory layer for agent context (`memoryAppend`, `memoryRead`, `memorySearch`)
- Self-update mechanism
- Support for Go, Java, Rust, Python, TypeScript, and JavaScript via tree-sitter
- SQLite-backed symbol index with file watcher for real-time updates
- MCP tools: `projectInfo`, `goToDefinition`, `findUsages`, `dependencyGraph`
