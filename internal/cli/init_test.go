package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateClaudeMd(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	claudeMdPath := filepath.Join(tmpDir, "CLAUDE.md")

	t.Run("CreateNew", func(t *testing.T) {
		// Test creating a new CLAUDE.md
		err := updateClaudeMd(claudeMdPath, tmpDir, false)
		if err != nil {
			t.Fatalf("failed to create CLAUDE.md: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(claudeMdPath); os.IsNotExist(err) {
			t.Fatal("CLAUDE.md was not created")
		}

		// Read and verify content
		content, err := os.ReadFile(claudeMdPath)
		if err != nil {
			t.Fatalf("failed to read CLAUDE.md: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, claudeMdBeginMarker) {
			t.Error("CLAUDE.md missing begin marker")
		}
		if !strings.Contains(contentStr, claudeMdEndMarker) {
			t.Error("CLAUDE.md missing end marker")
		}
		if !strings.Contains(contentStr, "mesdx.skill.bugfix") {
			t.Error("CLAUDE.md missing skill prompt reference")
		}
		if !strings.Contains(contentStr, "claude mcp add mesdx") {
			t.Error("CLAUDE.md missing setup command")
		}
		// Verify it starts with the marker (no wrapper content)
		if !strings.HasPrefix(strings.TrimSpace(contentStr), claudeMdBeginMarker) {
			t.Error("CLAUDE.md should start with begin marker when created new")
		}
	})

	t.Run("UpdateExistingWithoutMarkers", func(t *testing.T) {
		// Create a CLAUDE.md without markers
		initialContent := `# My Project

Some existing content.
`
		err := os.WriteFile(claudeMdPath, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("failed to write initial CLAUDE.md: %v", err)
		}

		// Update it
		err = updateClaudeMd(claudeMdPath, tmpDir, true)
		if err != nil {
			t.Fatalf("failed to update CLAUDE.md: %v", err)
		}

		// Verify content was preserved and markers added
		content, err := os.ReadFile(claudeMdPath)
		if err != nil {
			t.Fatalf("failed to read updated CLAUDE.md: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "My Project") {
			t.Error("CLAUDE.md lost original content")
		}
		if !strings.Contains(contentStr, claudeMdBeginMarker) {
			t.Error("CLAUDE.md missing begin marker after update")
		}
		if !strings.Contains(contentStr, claudeMdEndMarker) {
			t.Error("CLAUDE.md missing end marker after update")
		}
	})

	t.Run("UpdateExistingWithMarkers", func(t *testing.T) {
		// Create a CLAUDE.md with existing markers and old content
		initialContent := `# My Project

` + claudeMdBeginMarker + `
Old MesDX content that should be replaced.
` + claudeMdEndMarker + `

More content.
`
		err := os.WriteFile(claudeMdPath, []byte(initialContent), 0644)
		if err != nil {
			t.Fatalf("failed to write initial CLAUDE.md: %v", err)
		}

		// Update it
		err = updateClaudeMd(claudeMdPath, tmpDir, true)
		if err != nil {
			t.Fatalf("failed to update CLAUDE.md: %v", err)
		}

		// Verify content
		content, err := os.ReadFile(claudeMdPath)
		if err != nil {
			t.Fatalf("failed to read updated CLAUDE.md: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "My Project") {
			t.Error("CLAUDE.md lost original content")
		}
		if !strings.Contains(contentStr, "More content") {
			t.Error("CLAUDE.md lost content after markers")
		}
		if strings.Contains(contentStr, "Old MesDX content") {
			t.Error("CLAUDE.md still contains old content between markers")
		}
		if !strings.Contains(contentStr, "mesdx.skill.bugfix") {
			t.Error("CLAUDE.md missing updated content")
		}
	})
}

func TestUpdateManagedSection(t *testing.T) {
	tests := []struct {
		name            string
		existing        string
		managedContent  string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:           "NoMarkers",
			existing:       "# Title\n\nContent",
			managedContent: "New managed content",
			wantContains: []string{
				"# Title",
				"Content",
				claudeMdBeginMarker,
				"New managed content",
				claudeMdEndMarker,
			},
		},
		{
			name: "ExistingMarkers",
			existing: "# Title\n" +
				claudeMdBeginMarker + "\nOld content\n" + claudeMdEndMarker + "\nAfter",
			managedContent: "New managed content",
			wantContains: []string{
				"# Title",
				"New managed content",
				"After",
			},
			wantNotContains: []string{
				"Old content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateManagedSection(tt.existing, tt.managedContent)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result missing expected content: %q", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("result contains unwanted content: %q", notWant)
				}
			}
		})
	}
}
