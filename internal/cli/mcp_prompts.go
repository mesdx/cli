package cli

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerSkillPrompts registers MCP prompts for common workflows.
func registerSkillPrompts(server *mcp.Server, repoRoot string, memEnabled bool) {
	// Bugfix skill
	server.AddPrompt(&mcp.Prompt{
		Name:        "mesdx.skill.bugfix",
		Description: "Step-by-step guidance for investigating and fixing bugs using MesDX tools",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return bugfixPromptHandler(req, memEnabled)
	})

	// Refactoring skill
	server.AddPrompt(&mcp.Prompt{
		Name:        "mesdx.skill.refactoring",
		Description: "Safe refactoring workflow with impact analysis using MesDX tools",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return refactoringPromptHandler(req, memEnabled)
	})

	// Feature development skill
	server.AddPrompt(&mcp.Prompt{
		Name:        "mesdx.skill.feature_development",
		Description: "Plan and implement new features using MesDX navigation and analysis tools",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return featureDevPromptHandler(req, memEnabled)
	})

	// Security analysis skill
	server.AddPrompt(&mcp.Prompt{
		Name:        "mesdx.skill.security_analysis",
		Description: "Find and document security vulnerabilities using MesDX graph and memory tools",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return securityAnalysisPromptHandler(req, memEnabled)
	})

	// SCM Search skill
	server.AddPrompt(&mcp.Prompt{
		Name:        "mesdx.skill.scm_search",
		Description: "Write and use Tree-sitter SCM queries to search code structurally with mesdx.scmSearch",
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		return scmSearchPromptHandler(req)
	})
}

func bugfixPromptHandler(req *mcp.GetPromptRequest, memEnabled bool) (*mcp.GetPromptResult, error) {
	memoryGuidance := ""
	if memEnabled {
		memoryGuidance = `

**5. Document Your Investigation (Memory)**
   - Create a file-scoped memory to track your hypothesis and findings:
     ` + "`mesdx.memoryAppend`" + ` with scope="file", file="path/to/buggy/file.ext"
   - Include: reproduction steps, hypothesis, touched files, verification plan
   - **Why**: Preserves context if interrupted; helps reviewers understand the fix`
	}

	content := `# Bugfix Workflow with MesDX

Use this systematic approach to investigate and fix bugs efficiently.

## Step-by-Step Process

**1. Locate the Bug Source**
   - Use ` + "`mesdx.goToDefinition`" + ` to find the function/class definition
   - Provide either:
     - Cursor position: filePath + line + column (from error stacktrace)
     - Symbol name: symbolName + language
   - **Why**: Confirms the exact definition and its signature

**2. Find All Call Sites**
   - Use ` + "`mesdx.findUsages`" + ` on the buggy symbol
   - Set fetchCodeLinesAround=3 to see context around each usage
   - Results are scored (0-1) by confidence; review high-scoring usages first
   - **Why**: Identifies where the bug manifests and potential side effects

**3. Analyze Blast Radius**
   - Use ` + "`mesdx.dependencyGraph`" + ` to understand who depends on this symbol
   - Review inbound edges (what calls it) and outbound edges (what it calls)
   - Set maxDepth=1 for immediate dependencies, maxDepth=2 for transitive
   - **Why**: Ensures your fix won't break dependent code` + memoryGuidance + `

**6. Verify the Fix**
   - After fixing, re-run ` + "`mesdx.findUsages`" + ` to confirm all call sites still make sense
   - Check ` + "`mesdx.dependencyGraph`" + ` hasn't introduced new unexpected dependencies

## Tool Summary

- ` + "`mesdx.goToDefinition`" + ` → Find exact definition
- ` + "`mesdx.findUsages`" + ` → Map all call sites with context
- ` + "`mesdx.dependencyGraph`" + ` → Impact analysis (blast radius)`

	if memEnabled {
		content += `
- ` + "`mesdx.memoryAppend`" + ` → Document investigation progress`
	}

	return &mcp.GetPromptResult{
		Description: "Bugfix workflow guidance using MesDX tools for navigation and impact analysis",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: content,
				},
			},
		},
	}, nil
}

func refactoringPromptHandler(req *mcp.GetPromptRequest, memEnabled bool) (*mcp.GetPromptResult, error) {
	memoryWorkflow := ""
	if memEnabled {
		memoryWorkflow = `

**Memory-Driven Refactoring**

1. **Gather Context First**
   - ` + "`mesdx.memorySearch`" + ` with keywords related to the code area
   - **Why**: Discover existing design decisions, gotchas, or previous refactor notes

2. **Document Your Plan**
   - ` + "`mesdx.memoryAppend`" + ` (scope="project" or file-scoped for targeted changes)
   - Include: refactor goal, invariants to preserve, affected modules
   - **Why**: Creates a checkpoint; reviewers understand intent

3. **Update as You Go**
   - ` + "`mesdx.memoryUpdate`" + ` or ` + "`mesdx.memoryGrepReplace`" + ` to reflect progress
   - **Why**: Maintains living documentation of the refactor state`
	}

	content := `# Safe Refactoring with MesDX

Minimize risk when renaming, moving, or restructuring code.

## Core Principle

**Always analyze dependencies BEFORE making changes.**

## Step-by-Step Refactoring

**1. Understand Current Dependencies**
   - Use ` + "`mesdx.dependencyGraph`" + ` on the symbol/function you want to refactor
   - **Inbound edges**: Who depends on this? (breaking these = breaking change)
   - **Outbound edges**: What does this depend on? (changing these = scope of work)
   - Set minScore=0.3 to filter out low-confidence matches
   - **Why**: Quantifies refactor risk and scope

**2. Verify All Usages**
   - Use ` + "`mesdx.findUsages`" + ` to list every reference
   - Set fetchCodeLinesAround=5 for sufficient context
   - Group adjacent usages to understand local usage patterns
   - **Why**: Ensures you update all call sites correctly

**3. Check for Ambiguity**
   - If ` + "`mesdx.goToDefinition`" + ` by name returns multiple candidates, you have overloads
   - Plan renames carefully to avoid conflicts
   - **Why**: Prevents accidental shadowing or naming collisions

**4. Refactor Incrementally**
   - For large refactors, work file-by-file or module-by-module
   - Re-run ` + "`mesdx.dependencyGraph`" + ` after each stage to verify impact
   - **Why**: Reduces risk; easier to isolate issues` + memoryWorkflow + `

## Tool Priority

1. ` + "`mesdx.dependencyGraph`" + ` — **Always first** (understand impact)
2. ` + "`mesdx.findUsages`" + ` — Verify completeness
3. ` + "`mesdx.goToDefinition`" + ` — Disambiguate overloads`

	if memEnabled {
		content += `
4. ` + "`mesdx.memorySearch`" + ` / ` + "`mesdx.memoryAppend`" + ` — Context & documentation`
	}

	return &mcp.GetPromptResult{
		Description: "Safe refactoring workflow with dependency analysis and optional memory tracking",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: content,
				},
			},
		},
	}, nil
}

func featureDevPromptHandler(req *mcp.GetPromptRequest, memEnabled bool) (*mcp.GetPromptResult, error) {
	memoryAnnotation := ""
	if memEnabled {
		memoryAnnotation = `

**Memory for Feature Context**

1. **Create a Project-Level Feature Note**
   - ` + "`mesdx.memoryAppend`" + ` with scope="project"
   - Title: "Feature: [Feature Name]"
   - Content: goal, design decisions, affected modules
   - **Why**: Centralizes feature context across sessions

2. **Add File-Scoped Notes for Key Files**
   - ` + "`mesdx.memoryAppend`" + ` with scope="file" for entrypoints or critical logic
   - **Why**: Annotates implementation details at the file level

3. **Search Before You Build**
   - ` + "`mesdx.memorySearch`" + ` for related features or patterns
   - **Why**: Reuse existing patterns; avoid reinventing the wheel`
	}

	content := `# Feature Development with MesDX

Build new features with comprehensive understanding of the codebase.

## Development Workflow

**1. Orient Yourself (Project Structure)**
   - Use ` + "`mesdx.projectInfo`" + ` to see repo root and configured source roots
   - **Why**: Understand the project layout and indexing scope

**2. Find Extension Points**
   - Use ` + "`mesdx.goToDefinition`" + ` to locate base classes, interfaces, or plugin hooks
   - Search by symbolName if you know the interface/class name
   - **Why**: Identifies where to hook in your new functionality

**3. Understand Existing Usage Patterns**
   - Use ` + "`mesdx.findUsages`" + ` on extension points to see how others have extended them
   - Set fetchCodeLinesAround=10 for fuller examples
   - **Why**: Maintains consistency with existing code patterns

**4. Assess Integration Impact**
   - Use ` + "`mesdx.dependencyGraph`" + ` on modules your feature will integrate with
   - Check inbound dependencies to understand who else relies on these modules
   - **Why**: Prevents accidental breaking changes to existing features

**5. Implement Incrementally**
   - After adding new symbols, use ` + "`mesdx.findUsages`" + ` to verify expected call sites
   - Use ` + "`mesdx.dependencyGraph`" + ` to confirm integration points are wired correctly
   - **Why**: Catches integration bugs early` + memoryAnnotation + `

## Tool Sequence

1. ` + "`mesdx.projectInfo`" + ` → Understand structure
2. ` + "`mesdx.goToDefinition`" + ` → Find extension points
3. ` + "`mesdx.findUsages`" + ` → Learn patterns
4. ` + "`mesdx.dependencyGraph`" + ` → Assess impact`

	if memEnabled {
		content += `
5. ` + "`mesdx.memoryAppend`" + ` / ` + "`mesdx.memorySearch`" + ` → Track & reuse context`
	}

	return &mcp.GetPromptResult{
		Description: "Feature development workflow using MesDX navigation, impact analysis, and memory",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: content,
				},
			},
		},
	}, nil
}

func securityAnalysisPromptHandler(req *mcp.GetPromptRequest, memEnabled bool) (*mcp.GetPromptResult, error) {
	memoryForSecurity := ""
	if memEnabled {
		memoryForSecurity = `

**Memory for Security Findings**

Use structured memories to track vulnerabilities:

1. **Create Finding Records**
   - ` + "`mesdx.memoryAppend`" + ` with consistent structure:
     ` + "```" + `
     ## Finding: [Type] in [Component]
     **Impact**: [Severity + Description]
     **Evidence**: [File:Line references]
     **Mitigation**: [Fix approach or status]
     ` + "```" + `

2. **Scope Appropriately**
   - Use scope="project" for systemic issues (e.g., lack of input validation)
   - Use scope="file" for localized vulnerabilities (e.g., SQL injection in one route)

3. **Search for Patterns**
   - ` + "`mesdx.memorySearch`" + ` to find similar past findings or mitigation notes
   - **Why**: Consistency in remediation; avoid duplicate work`
	}

	content := `# Security Analysis with MesDX

Identify and document security vulnerabilities systematically.

## Analysis Approach

**1. Identify Security-Sensitive Code**

Start by finding symbols related to:
- Authentication/authorization (` + "`login`, `authenticate`, `checkPermission`" + `)
- Input handling (` + "`parseInput`, `validateRequest`, `sanitize`" + `)
- Data access (` + "`query`, `execute`, `fetch`" + `, SQL/database calls)
- Cryptography (` + "`encrypt`, `hash`, `sign`" + `)
- External communication (` + "`fetch`, `request`, `httpClient`" + `)

Use ` + "`mesdx.goToDefinition`" + ` by symbolName to locate these functions.

**2. Trace Data Flows**
   - Use ` + "`mesdx.findUsages`" + ` on input sources (e.g., HTTP handlers, parsers)
   - Set fetchCodeLinesAround=5 to see how data is handled
   - **Why**: Identifies where untrusted input flows into sensitive operations

**3. Analyze Dependency Paths**
   - Use ` + "`mesdx.dependencyGraph`" + ` on sensitive functions
   - **Inbound edges**: What code paths lead to this sensitive operation?
   - **Outbound edges**: What sensitive resources does this access?
   - Set maxDepth=2 to trace multi-hop paths
   - **Why**: Finds indirect vulnerabilities (e.g., user input → parser → SQL query)

**4. Check Authorization Boundaries**
   - Use ` + "`mesdx.findUsages`" + ` on authorization checks
   - Verify that sensitive operations are always preceded by permission checks
   - **Why**: Detects missing authorization (CWE-284)

**5. Review Cryptographic Usage**
   - Use ` + "`mesdx.findUsages`" + ` on crypto functions
   - Check for weak algorithms, hardcoded keys, or improper initialization
   - **Why**: Cryptographic failures (OWASP Top 10 #2)` + memoryForSecurity + `

## Common Vulnerability Patterns

**SQL Injection**: ` + "`mesdx.findUsages`" + ` on ` + "`query`, `execute`" + ` → check for string concatenation
**XSS**: ` + "`mesdx.findUsages`" + ` on output/rendering functions → verify escaping
**IDOR**: ` + "`mesdx.dependencyGraph`" + ` on data access → verify authorization checks upstream
**Secrets in Code**: ` + "`mesdx.findUsages`" + ` on ` + "`password`, `apiKey`, `secret`" + ` → check for hardcoding

## Tool Strategy

1. ` + "`mesdx.goToDefinition`" + ` → Locate security-sensitive symbols
2. ` + "`mesdx.findUsages`" + ` → Trace data sources and sinks
3. ` + "`mesdx.dependencyGraph`" + ` → Map attack paths (multi-hop)`

	if memEnabled {
		content += `
4. ` + "`mesdx.memoryAppend`" + ` → Document findings with consistent structure`
	}

	return &mcp.GetPromptResult{
		Description: "Security analysis workflow using MesDX graph tools and structured memory notes",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: content,
				},
			},
		},
	}, nil
}

func scmSearchPromptHandler(req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := `# Structural Code Search with Tree-sitter (mesdx.scmSearch)

Search your codebase by AST structure, not just text. This tool runs Tree-sitter
S-expression queries in parallel across all source files.

## Quick Start: Use a Stub

Stubs are predefined, per-language query templates. Just supply a name and arguments.

| Stub ID | Args | Description |
|---------|------|-------------|
| ` + "`defs.function.named`" + ` | {name} | Find function definitions by name |
| ` + "`defs.class.named`" + ` | {name} | Find class/struct definitions by name |
| ` + "`defs.interface.named`" + ` | {name} | Find interface/trait definitions by name |
| ` + "`defs.method.named`" + ` | {name} | Find method definitions by name |
| ` + "`refs.type.named`" + ` | {name} | Find type references by name |
| ` + "`refs.call.named`" + ` | {name} | Find call sites of a function by name |
| ` + "`refs.write.named`" + ` | {name} | Find assignment sites for a variable by name |
| ` + "`refs.import.named`" + ` | {name} | Find import statements referencing a name |

**Example** — Find all definitions of ` + "`NewPerson`" + `:
` + "```" + `
mesdx.scmSearch(language="go", stubName="defs.function.named", stubArgs={"name": "NewPerson"})
` + "```" + `

## Writing Raw SCM Queries

Tree-sitter queries use S-expressions (Lisp-like) to match AST node patterns.
Each parenthesized form matches a node type; ` + "`name:`" + ` matches a named child field.
Prefix a node with ` + "`@captureName`" + ` to capture it in results.

### Syntax Cheat Sheet

` + "```scheme" + `
;; Match any function declaration, capture its name
(function_declaration name: (identifier) @fn)

;; Match calls where the callee is literally "foo"
(call_expression
  function: (identifier) @call
  (#eq? @call "foo"))

;; Match class names starting with "Base"
(class_declaration
  name: (type_identifier) @cls
  (#match? @cls "^Base"))

;; Match any of several values
(identifier) @id (#any-of? @id "read" "write" "close")
` + "```" + `

### Predicates

- **` + "`#eq?`" + `** ` + "`@capture \"literal\"`" + ` — exact string match
- **` + "`#match?`" + `** ` + "`@capture \"regex\"`" + ` — regex match (Rust regex syntax)
- **` + "`#any-of?`" + `** ` + "`@capture \"a\" \"b\" \"c\"`" + ` — match any of the listed values

### Common Node Types by Language

**Go**: ` + "`function_declaration`" + `, ` + "`method_declaration`" + `, ` + "`type_declaration`" + `, ` + "`type_spec`" + `, ` + "`struct_type`" + `, ` + "`interface_type`" + `, ` + "`call_expression`" + `, ` + "`identifier`" + `, ` + "`type_identifier`" + `, ` + "`field_identifier`" + `
**TypeScript**: ` + "`function_declaration`" + `, ` + "`class_declaration`" + `, ` + "`interface_declaration`" + `, ` + "`method_definition`" + `, ` + "`call_expression`" + `, ` + "`identifier`" + `, ` + "`type_identifier`" + `, ` + "`property_identifier`" + `
**Python**: ` + "`function_definition`" + `, ` + "`class_definition`" + `, ` + "`call`" + `, ` + "`assignment`" + `, ` + "`identifier`" + `, ` + "`attribute`" + `, ` + "`decorator`" + `
**Java**: ` + "`class_declaration`" + `, ` + "`method_declaration`" + `, ` + "`method_invocation`" + `, ` + "`interface_declaration`" + `, ` + "`identifier`" + `, ` + "`type_identifier`" + `
**Rust**: ` + "`function_item`" + `, ` + "`struct_item`" + `, ` + "`enum_item`" + `, ` + "`trait_item`" + `, ` + "`impl_item`" + `, ` + "`call_expression`" + `, ` + "`identifier`" + `, ` + "`type_identifier`" + `

### Tips

1. **Start broad, then restrict.** Begin with just a node type (e.g. ` + "`(function_declaration)`" + `), then add field constraints and predicates.
2. **Use ` + "`includeGlobs`" + `** to limit scope (e.g. ` + "`[\"*.go\"]`" + ` or ` + "`[\"src/api/*.ts\"]`" + `).
3. **Capture multiple things** in one query by adding more ` + "`@capture`" + ` names.
4. **Check ` + "`astParents`" + `** in results to understand nesting context.
5. **Combine with ` + "`mesdx.goToDefinition`" + `** or ` + "`mesdx.findUsages`" + `** for deeper analysis after identifying locations.

## Output

Each match includes:
- File path, line/column range
- Capture name and AST node type
- Matched text snippet
- Source line with surrounding context (configurable via ` + "`contextLines`" + `)
- AST parent chain (configurable depth via ` + "`astParentDepth`" + `)
`

	return &mcp.GetPromptResult{
		Description: "Guide for writing and using Tree-sitter SCM queries with mesdx.scmSearch",
		Messages: []*mcp.PromptMessage{
			{
				Role: "user",
				Content: &mcp.TextContent{
					Text: content,
				},
			},
		},
	}, nil
}
