<!-- mesdx:begin -->
## MesDX Code Intelligence

MesDX provides precise, local-first code navigation and analysis via the Model Context Protocol (MCP).

### Setup

Add MesDX to Claude Code with this command:

```bash
claude mcp add mesdx --transport stdio -- mesdx mcp --cwd <REPO_ROOT>
```

Replace `<REPO_ROOT>` with the absolute path to this repository.


### Available Tools

**Navigation & Definition Lookup**
- `mesdx.projectInfo` — Get repo root, source roots, and database path
- `mesdx.goToDefinition` — Find symbol definitions by cursor position or name
- `mesdx.findUsages` — Find all references to a symbol with dependency scoring

**Impact Analysis**
- `mesdx.dependencyGraph` — Analyze inbound/outbound dependencies for refactor risk assessment

**Memory (Persistent Context)**
- `mesdx.memoryAppend` — Create project or file-scoped markdown notes
- `mesdx.memoryRead` — Read or list memory elements
- `mesdx.memorySearch` — Full-text search across memories
- `mesdx.memoryUpdate` — Update existing memories
- `mesdx.memoryGrepReplace` — Regex find-and-replace in memories
- `mesdx.memoryDelete` — Soft-delete memories

### Workflow Skills

For detailed step-by-step guidance on specific workflows, use MCP prompts:

- `mesdx.skill.bugfix` — Navigate, analyze, and document bug fixes
- `mesdx.skill.refactoring` — Safe refactoring with impact analysis
- `mesdx.skill.feature_development` — Plan and implement new features
- `mesdx.skill.security_analysis` — Find and document security issues

**Supported languages:** Go, Java, Rust, Python, TypeScript, JavaScript

<!-- mesdx:end -->
