package ignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	gitignoreFile    = ".gitignore"
	dockerignoreFile = ".dockerignore"
	ignorePattern    = ".codeintelx/"
	commentMarker    = "# codeintelx"
)

var (
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
)

// HandleIgnoreFiles checks for .gitignore and .dockerignore files and prompts
// the user to add .codeintelx/ if it's not already ignored.
func HandleIgnoreFiles(repoRoot string, cmd *cobra.Command) error {
	// Handle .gitignore
	gitignorePath := filepath.Join(repoRoot, gitignoreFile)
	if err := handleIgnoreFile(gitignorePath, "Git", "prevents committing local index/config state to version control", cmd); err != nil {
		return fmt.Errorf("failed to handle .gitignore: %w", err)
	}

	// Handle .dockerignore
	dockerignorePath := filepath.Join(repoRoot, dockerignoreFile)
	if err := handleIgnoreFile(dockerignorePath, "Docker", "prevents sending the local index/db in build context (smaller builds, avoids leaking local state)", cmd); err != nil {
		return fmt.Errorf("failed to handle .dockerignore: %w", err)
	}

	return nil
}

func handleIgnoreFile(filePath, toolName, impact string, cmd *cobra.Command) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, skip
	}

	// Check if already ignored
	alreadyIgnored, err := isAlreadyIgnored(filePath)
	if err != nil {
		return fmt.Errorf("failed to check ignore file: %w", err)
	}

	if alreadyIgnored {
		return nil // Already ignored, skip
	}

	// Prompt user
	var shouldAdd bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Add .codeintelx/ to %s?", toolName)).
				Description(fmt.Sprintf("This %s", impact)).
				Value(&shouldAdd),
		),
	)

	if err := form.Run(); err != nil {
		return fmt.Errorf("interactive prompt failed: %w", err)
	}

	if !shouldAdd {
		cmd.Printf("%s Skipped adding to %s\n", infoStyle.Render("→"), toolName)
		return nil
	}

	// Add ignore entry
	if err := addIgnoreEntry(filePath); err != nil {
		return fmt.Errorf("failed to add ignore entry: %w", err)
	}

	cmd.Printf("✓ Added .codeintelx/ to %s\n", toolName)
	return nil
}

// isAlreadyIgnored checks if .codeintelx/ is already in the ignore file.
func isAlreadyIgnored(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Normalize the line by removing trailing slash for comparison
		normalizedLine := strings.TrimSuffix(line, "/")

		// Check for exact match of .codeintelx (with or without trailing slash)
		if normalizedLine == ".codeintelx" {
			return true, nil
		}

		// Check for patterns that specifically match .codeintelx/ directory
		// We need to be very specific to avoid matching things like "cmd/codeintelx"
		// Valid patterns:
		// - .codeintelx/ (exact)
		// - .codeintelx/** (glob pattern)
		// - **/.codeintelx/ (glob pattern)
		// - .codeintelx (without slash)

		// Only check for patterns that start with ".codeintelx" (not just contain it)
		// This ensures we don't match paths like "cmd/codeintelx" or "some/.codeintelx/other"
		if strings.HasPrefix(normalizedLine, ".codeintelx") {
			// Check if it's exactly ".codeintelx" or starts with ".codeintelx/" or ".codeintelx*"
			// This covers: .codeintelx, .codeintelx/, .codeintelx/**, etc.
			if normalizedLine == ".codeintelx" || strings.HasPrefix(line, ".codeintelx/") || strings.HasPrefix(line, ".codeintelx*") {
				return true, nil
			}
		}

		// Check for glob patterns like "**/.codeintelx/" or "**/.codeintelx"
		// But NOT patterns like "cmd/.codeintelx", "cmd/codeintelx", or "some/.codeintelx/other"
		if strings.Contains(line, "/.codeintelx") {
			// Only match if it's a glob pattern starting with **
			// This matches: **/.codeintelx/, **/.codeintelx
			// But NOT: cmd/.codeintelx, cmd/codeintelx, etc.
			if strings.HasPrefix(line, "**") {
				// Verify it ends with .codeintelx (the directory pattern)
				if strings.HasSuffix(normalizedLine, ".codeintelx") {
					return true, nil
				}
			}
		}
	}

	return false, scanner.Err()
}

// addIgnoreEntry appends .codeintelx/ to the ignore file in an idempotent way.
func addIgnoreEntry(filePath string) error {
	// Double-check it's not already there (race condition protection)
	alreadyIgnored, err := isAlreadyIgnored(filePath)
	if err != nil {
		return fmt.Errorf("failed to re-check ignore file: %w", err)
	}
	if alreadyIgnored {
		return nil // Already there, nothing to do
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add a newline if file doesn't end with one, then add our entry
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() > 0 {
		// Check if last character is newline
		if _, err := file.Seek(stat.Size()-1, 0); err != nil {
			return err
		}
		var lastChar [1]byte
		if _, err := file.Read(lastChar[:]); err == nil && lastChar[0] != '\n' {
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
		if _, err := file.Seek(0, 2); err != nil { // Seek to end
			return err
		}
	}

	// Append ignore entry with comment
	entry := fmt.Sprintf("\n%s\n%s\n", commentMarker, ignorePattern)
	if _, err := file.WriteString(entry); err != nil {
		return err
	}

	return nil
}
