# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.1] - 2026-02-23
### Added
- **Noise filtering for definitions and usages**: Go-to-definition and find-usages rank candidates by confidence and filter low-confidence noise (e.g. same-name variables or parameters), preferring the primary type or function definition in name-collision cases
- **Coupling-strength scoring**: Usage sites are scored by coupling strength (e.g. inheritance and calls vs casual references); dependency graph and find-usages use these scores for impact and coupling metrics
- **Richer dependency graph**: Inbound and outbound sections include per-edge coupling/impact scores, normalized scoring, and improved handling of transitive relationships for impact assessment
- **Python string type annotations**: Type references inside Python string annotations (e.g. `def f() -> "MyClass"`, `x: "MyClass"`) are now indexed and resolved for go-to-definition and find-usages
- **Deeper function and symbol navigation**: Improved resolution of definitions and usages across files and in complex name-collision scenarios
- **Added scm search tool**: SCM search tool give agents ability to search codebase with treesitter SCM queries, this gives them higher hits when doing usage analysis.

### Changed
- **Dependency scoring**: Inbound and outbound weights are normalized to probabilities for more consistent impact and coupling scores across the dependency graph

## [0.3.1] - 2026-02-22
### Added
- Added better support for cursor and antigravity.

## [0.3.0] - 2026-02-22
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
