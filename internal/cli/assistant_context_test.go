package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectAssistants(t *testing.T) {
	t.Run("NothingDetected", func(t *testing.T) {
		dir := t.TempDir()
		result := detectAssistants(dir)
		for _, a := range result {
			if a.Detected {
				t.Errorf("expected %s to not be detected in empty dir", a.Kind)
			}
		}
	})

	t.Run("ClaudeDetected", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Claude"), 0644); err != nil {
			t.Fatal(err)
		}
		found := findByKind(detectAssistants(dir), AssistantClaude)
		if found == nil || !found.Detected {
			t.Error("Claude should be detected when CLAUDE.md exists")
		}
	})

	t.Run("CursorDetectedViaDotCursor", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".cursor"), 0755); err != nil {
			t.Fatal(err)
		}
		found := findByKind(detectAssistants(dir), AssistantCursor)
		if found == nil || !found.Detected {
			t.Error("Cursor should be detected when .cursor/ exists")
		}
	})

	t.Run("CursorDetectedViaCursorrules", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".cursorrules"), []byte("rules"), 0644); err != nil {
			t.Fatal(err)
		}
		found := findByKind(detectAssistants(dir), AssistantCursor)
		if found == nil || !found.Detected {
			t.Error("Cursor should be detected when .cursorrules exists")
		}
	})

	t.Run("CursorDetectedViaAgentsMd", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("agents"), 0644); err != nil {
			t.Fatal(err)
		}
		found := findByKind(detectAssistants(dir), AssistantCursor)
		if found == nil || !found.Detected {
			t.Error("Cursor should be detected when AGENTS.md exists")
		}
	})

	t.Run("AntigravityDetected", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".agent"), 0755); err != nil {
			t.Fatal(err)
		}
		found := findByKind(detectAssistants(dir), AssistantAntigravity)
		if found == nil || !found.Detected {
			t.Error("Antigravity should be detected when .agent/ exists")
		}
	})

	t.Run("MultipleDetected", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("claude"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, ".cursor"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, ".agent"), 0755); err != nil {
			t.Fatal(err)
		}
		for _, a := range detectAssistants(dir) {
			if !a.Detected {
				t.Errorf("expected %s to be detected when all signals present", a.Kind)
			}
		}
	})
}

func TestWriteAssistantGuidance(t *testing.T) {
	t.Run("Claude", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAssistantGuidance(dir, AssistantClaude); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
		if err != nil {
			t.Fatalf("failed to read CLAUDE.md: %v", err)
		}
		s := string(content)
		if !strings.Contains(s, "claude mcp add mesdx") {
			t.Error("CLAUDE.md missing Claude setup command")
		}
		if !strings.Contains(s, "mesdx.goToDefinition") {
			t.Error("CLAUDE.md missing tool reference")
		}
	})

	t.Run("Cursor_CreateNew", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAssistantGuidance(dir, AssistantCursor); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		path := filepath.Join(dir, ".cursor", "rules", "mesdx.mdc")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		s := string(content)
		if !strings.Contains(s, "alwaysApply: true") {
			t.Error("Cursor rule missing frontmatter alwaysApply")
		}
		if !strings.Contains(s, "mcpServers") {
			t.Error("Cursor rule missing mcpServers config snippet")
		}
		if !strings.Contains(s, ".cursor/mcp.json") {
			t.Error("Cursor rule missing .cursor/mcp.json reference")
		}
		if !strings.Contains(s, "mesdx.goToDefinition") {
			t.Error("Cursor rule missing tool reference")
		}
		if !strings.Contains(s, "mesdx.skill.bugfix") {
			t.Error("Cursor rule missing skill reference")
		}
	})

	t.Run("Cursor_Overwrite", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAssistantGuidance(dir, AssistantCursor); err != nil {
			t.Fatalf("first write failed: %v", err)
		}
		if err := writeAssistantGuidance(dir, AssistantCursor); err != nil {
			t.Fatalf("second write failed: %v", err)
		}
		path := filepath.Join(dir, ".cursor", "rules", "mesdx.mdc")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if strings.Count(string(content), "---") != 2 {
			t.Error("overwrite resulted in duplicate frontmatter")
		}
	})

	t.Run("Antigravity_CreateNew", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAssistantGuidance(dir, AssistantAntigravity); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		path := filepath.Join(dir, ".agent", "rules", "mesdx.mdc")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		s := string(content)
		if !strings.Contains(s, "alwaysApply: true") {
			t.Error("Antigravity rule missing frontmatter alwaysApply")
		}
		if !strings.Contains(s, ".vscode/mcp.json") {
			t.Error("Antigravity rule missing .vscode/mcp.json reference")
		}
		if !strings.Contains(s, `"type": "stdio"`) {
			t.Error("Antigravity rule missing stdio type in MCP config")
		}
		if !strings.Contains(s, "mesdx.goToDefinition") {
			t.Error("Antigravity rule missing tool reference")
		}
		if !strings.Contains(s, "mesdx.skill.bugfix") {
			t.Error("Antigravity rule missing skill reference")
		}
	})

	t.Run("Antigravity_Overwrite", func(t *testing.T) {
		dir := t.TempDir()
		if err := writeAssistantGuidance(dir, AssistantAntigravity); err != nil {
			t.Fatalf("first write failed: %v", err)
		}
		if err := writeAssistantGuidance(dir, AssistantAntigravity); err != nil {
			t.Fatalf("second write failed: %v", err)
		}
		path := filepath.Join(dir, ".agent", "rules", "mesdx.mdc")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		if strings.Count(string(content), "---") != 2 {
			t.Error("overwrite resulted in duplicate frontmatter")
		}
	})

	t.Run("UnknownKind", func(t *testing.T) {
		dir := t.TempDir()
		err := writeAssistantGuidance(dir, "unknown")
		if err == nil {
			t.Error("expected error for unknown assistant kind")
		}
	})
}

func TestGenerateCursorGuidance(t *testing.T) {
	content := generateCursorGuidance()
	checks := []string{
		"---",
		"alwaysApply: true",
		"## MesDX Code Intelligence",
		"mcpServers",
		".cursor/mcp.json",
		"mesdx.goToDefinition",
		"mesdx.skill.bugfix",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("Cursor guidance missing: %q", c)
		}
	}
}

func TestGenerateAntigravityGuidance(t *testing.T) {
	content := generateAntigravityGuidance()
	checks := []string{
		"---",
		"alwaysApply: true",
		"## MesDX Code Intelligence",
		`"type": "stdio"`,
		".vscode/mcp.json",
		"mesdx.goToDefinition",
		"mesdx.skill.bugfix",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Errorf("Antigravity guidance missing: %q", c)
		}
	}
}

func TestMesdxToolsAndSkills(t *testing.T) {
	content := mesdxToolsAndSkills()
	required := []string{
		"mesdx.projectInfo",
		"mesdx.goToDefinition",
		"mesdx.findUsages",
		"mesdx.dependencyGraph",
		"mesdx.memoryAppend",
		"mesdx.memoryRead",
		"mesdx.memorySearch",
		"mesdx.memoryUpdate",
		"mesdx.memoryGrepReplace",
		"mesdx.memoryDelete",
		"mesdx.skill.bugfix",
		"mesdx.skill.refactoring",
		"mesdx.skill.feature_development",
		"mesdx.skill.security_analysis",
	}
	for _, r := range required {
		if !strings.Contains(content, r) {
			t.Errorf("mesdxToolsAndSkills missing: %q", r)
		}
	}
}

func findByKind(infos []AssistantInfo, kind AssistantKind) *AssistantInfo {
	for i := range infos {
		if infos[i].Kind == kind {
			return &infos[i]
		}
	}
	return nil
}
