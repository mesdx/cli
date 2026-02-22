package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// AssistantKind represents a supported AI coding assistant.
type AssistantKind string

const (
	AssistantClaude      AssistantKind = "claude"
	AssistantCursor      AssistantKind = "cursor"
	AssistantAntigravity AssistantKind = "antigravity"
)

// AssistantInfo holds detection results and metadata for an AI assistant.
type AssistantInfo struct {
	Kind        AssistantKind
	DisplayName string
	TargetPath  string // relative to repo root
	Detected    bool
}

var assistantRegistry = []struct {
	Kind        AssistantKind
	DisplayName string
	TargetPath  string
}{
	{AssistantClaude, "Claude Code", "CLAUDE.md"},
	{AssistantCursor, "Cursor", filepath.Join(".cursor", "rules", "mesdx.mdc")},
	{AssistantAntigravity, "Antigravity", filepath.Join(".agent", "rules", "mesdx.mdc")},
}

// detectAssistants checks the repo root for heuristic signals of each AI assistant.
func detectAssistants(repoRoot string) []AssistantInfo {
	result := make([]AssistantInfo, 0, len(assistantRegistry))
	for _, a := range assistantRegistry {
		result = append(result, AssistantInfo{
			Kind:        a.Kind,
			DisplayName: a.DisplayName,
			TargetPath:  a.TargetPath,
			Detected:    isAssistantPresent(repoRoot, a.Kind),
		})
	}
	return result
}

// isAssistantPresent checks for heuristic signals that an AI assistant is used in the repo.
//
// Heuristics (current as of 2026):
//   - Claude Code: CLAUDE.md in repo root
//   - Cursor: .cursor/ dir, .cursorrules file, or AGENTS.md
//   - Antigravity: .agent/ dir (Google Antigravity stores rules in .agent/rules/)
func isAssistantPresent(repoRoot string, kind AssistantKind) bool {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(repoRoot, rel))
		return err == nil
	}
	switch kind {
	case AssistantClaude:
		return exists("CLAUDE.md")
	case AssistantCursor:
		return exists(".cursor") || exists(".cursorrules") || exists("AGENTS.md")
	case AssistantAntigravity:
		return exists(".agent")
	}
	return false
}

// promptAndUpdateAssistantGuidance detects assistants, presents a multi-select
// (pre-selecting detected ones), confirms with the user, then writes guidance.
func promptAndUpdateAssistantGuidance(cmd *cobra.Command, repoRoot string) error {
	assistants := detectAssistants(repoRoot)

	var options []huh.Option[string]
	var defaults []string
	for _, a := range assistants {
		label := a.DisplayName
		if a.Detected {
			label += " (detected)"
		}
		options = append(options, huh.NewOption(label, string(a.Kind)))
		if a.Detected {
			defaults = append(defaults, string(a.Kind))
		}
	}
	if len(defaults) == 0 {
		defaults = []string{string(AssistantClaude)}
	}

	selected := append([]string{}, defaults...)
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Set up MesDX guidance for which AI assistants?").
				Description("Detected assistants are pre-selected. Space to toggle, Enter to confirm.").
				Options(options...).
				Value(&selected),
		),
	)
	if err := selectForm.Run(); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	if len(selected) == 0 {
		return nil
	}

	byKind := make(map[AssistantKind]AssistantInfo, len(assistants))
	for _, a := range assistants {
		byKind[a.Kind] = a
	}

	var paths []string
	for _, s := range selected {
		paths = append(paths, byKind[AssistantKind(s)].TargetPath)
	}

	var confirmed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Write MesDX guidance to the following files?").
				Description(strings.Join(paths, ", ")).
				Value(&confirmed),
		),
	)
	if err := confirmForm.Run(); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	if !confirmed {
		return nil
	}

	for _, s := range selected {
		kind := AssistantKind(s)
		info := byKind[kind]
		if err := writeAssistantGuidance(repoRoot, kind); err != nil {
			cmd.Printf("%s Warning: failed to write %s guidance: %v\n",
				infoStyle.Render("!"), info.DisplayName, err)
			continue
		}
		cmd.Printf("%s Written MesDX guidance to %s\n", successStyle.Render("✓"), info.TargetPath)
	}
	return nil
}

// writeAssistantGuidance writes or updates the guidance file for the given assistant.
func writeAssistantGuidance(repoRoot string, kind AssistantKind) error {
	switch kind {
	case AssistantClaude:
		path := filepath.Join(repoRoot, "CLAUDE.md")
		_, err := os.Stat(path)
		return updateClaudeMd(path, repoRoot, err == nil)
	case AssistantCursor:
		return writeMdcFile(repoRoot, filepath.Join(".cursor", "rules", "mesdx.mdc"), generateCursorGuidance())
	case AssistantAntigravity:
		return writeMdcFile(repoRoot, filepath.Join(".agent", "rules", "mesdx.mdc"), generateAntigravityGuidance())
	}
	return fmt.Errorf("unknown assistant kind: %s", kind)
}

// writeMdcFile creates parent directories and writes an .mdc rule file.
func writeMdcFile(repoRoot, relPath, content string) error {
	absPath := filepath.Join(repoRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
	}
	return os.WriteFile(absPath, []byte(content), 0644)
}

// bt wraps s in backticks for markdown inline code.
func bt(s string) string { return "`" + s + "`" }

// mesdxToolsAndSkills returns the shared tools/skills markdown used across all assistant guidance files.
func mesdxToolsAndSkills() string {
	return `### Available Tools

**Navigation & Definition Lookup**
- ` + bt("mesdx.projectInfo") + ` — Get repo root, source roots, and database path
- ` + bt("mesdx.goToDefinition") + ` — Find symbol definitions by cursor position or name
- ` + bt("mesdx.findUsages") + ` — Find all references to a symbol with dependency scoring

**Impact Analysis**
- ` + bt("mesdx.dependencyGraph") + ` — Analyze inbound/outbound dependencies for refactor risk assessment

**Memory (Persistent Context)**
- ` + bt("mesdx.memoryAppend") + ` — Create project or file-scoped markdown notes
- ` + bt("mesdx.memoryRead") + ` — Read or list memory elements
- ` + bt("mesdx.memorySearch") + ` — Full-text search across memories
- ` + bt("mesdx.memoryUpdate") + ` — Update existing memories
- ` + bt("mesdx.memoryGrepReplace") + ` — Regex find-and-replace in memories
- ` + bt("mesdx.memoryDelete") + ` — Soft-delete memories

### Workflow Skills

For detailed step-by-step guidance on specific workflows, use MCP prompts:

- ` + bt("mesdx.skill.bugfix") + ` — Navigate, analyze, and document bug fixes
- ` + bt("mesdx.skill.refactoring") + ` — Safe refactoring with impact analysis
- ` + bt("mesdx.skill.feature_development") + ` — Plan and implement new features
- ` + bt("mesdx.skill.security_analysis") + ` — Find and document security issues

**Supported languages:** Go, Java, Rust, Python, TypeScript, JavaScript`
}

func generateCursorGuidance() string {
	return `---
description: "MesDX Code Intelligence - local-first code navigation and analysis via MCP"
alwaysApply: true
---

## MesDX Code Intelligence

MesDX provides precise, local-first code navigation and analysis via the Model Context Protocol (MCP).

### Setup

Add MesDX as a project-wide MCP server by creating or updating ` + bt(".cursor/mcp.json") + `:

` + "```json" + `
{
  "mcpServers": {
    "mesdx": {
      "command": "mesdx",
      "args": ["mcp", "--cwd", "<REPO_ROOT>"]
    }
  }
}
` + "```" + `

Replace ` + bt("<REPO_ROOT>") + ` with the absolute path to this repository. Restart Cursor to load the MCP server.

` + mesdxToolsAndSkills() + `
`
}

func generateAntigravityGuidance() string {
	return `---
description: "MesDX Code Intelligence - local-first code navigation and analysis via MCP"
alwaysApply: true
---

## MesDX Code Intelligence

MesDX provides precise, local-first code navigation and analysis via the Model Context Protocol (MCP).

### Setup

Add MesDX as a project-wide MCP server by creating or updating ` + bt(".vscode/mcp.json") + `:

` + "```json" + `
{
  "servers": {
    "mesdx": {
      "type": "stdio",
      "command": "mesdx",
      "args": ["mcp", "--cwd", "<REPO_ROOT>"]
    }
  }
}
` + "```" + `

Replace ` + bt("<REPO_ROOT>") + ` with the absolute path to this repository.

` + mesdxToolsAndSkills() + `
`
}
