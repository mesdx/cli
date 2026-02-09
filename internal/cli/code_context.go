package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mesdx/cli/internal/indexer"
)

const (
	maxCodeChars = 100000 // Max total chars in concatenated code output
)

// codeWindow represents a contiguous range of lines in a file
type codeWindow struct {
	startLine int
	endLine   int
}

// fetchDefinitionsCode reads and concatenates code for all definitions.
// It expands the start line backward to include leading doc comments/annotations.
func fetchDefinitionsCode(repoRoot string, results []indexer.DefinitionResult) (string, error) {
	var sb strings.Builder
	totalChars := 0

	// Cache file lines by path to avoid repeated reads
	fileCache := map[string][]string{}

	for i, def := range results {
		if i > 0 {
			sb.WriteString("\n\n")
			totalChars += 2
		}

		absPath := safeJoinPath(repoRoot, def.Location.Path)
		if absPath == "" {
			continue
		}

		// Read file lines (cached)
		fileLines, ok := fileCache[absPath]
		if !ok {
			var err error
			fileLines, err = readFileAllLines(absPath)
			if err != nil {
				continue
			}
			fileCache[absPath] = fileLines
		}

		// Expand start to include leading doc comments/decorators
		lang := indexer.DetectLang(def.Location.Path)
		startLine := indexer.FindDocStartLine(fileLines, def.Location.StartLine, lang)
		endLine := def.Location.EndLine
		if endLine > len(fileLines) {
			endLine = len(fileLines)
		}

		// Write header
		header := fmt.Sprintf("--- %s:%d-%d (%s %s) ---\n",
			def.Location.Path,
			startLine,
			endLine,
			def.Name,
			def.Kind,
		)
		sb.WriteString(header)
		totalChars += len(header)

		// Write lines
		for line := startLine; line <= endLine; line++ {
			lineStr := fmt.Sprintf("%6d| %s\n", line, fileLines[line-1])
			if totalChars+len(lineStr) > maxCodeChars {
				remaining := maxCodeChars - totalChars
				if remaining > 0 {
					sb.WriteString(lineStr[:remaining])
				}
				sb.WriteString("\n... [truncated: output too large] ...")
				return sb.String(), nil
			}
			sb.WriteString(lineStr)
			totalChars += len(lineStr)
		}
	}

	return sb.String(), nil
}

// fetchUsagesCode reads and concatenates code for all usages with context
func fetchUsagesCode(repoRoot string, results []indexer.UsageResult, linesAround int) (string, error) {
	// Group usages by file
	fileUsages := make(map[string][]indexer.UsageResult)
	for _, usage := range results {
		fileUsages[usage.Location.Path] = append(fileUsages[usage.Location.Path], usage)
	}

	var sb strings.Builder
	totalChars := 0
	isFirst := true

	// Process each file
	for path, usages := range fileUsages {
		// Sort by line
		sort.Slice(usages, func(i, j int) bool {
			return usages[i].Location.StartLine < usages[j].Location.StartLine
		})

		// Compute windows with context
		windows := make([]codeWindow, 0, len(usages))
		for _, usage := range usages {
			start := usage.Location.StartLine - linesAround
			end := usage.Location.EndLine + linesAround
			if start < 1 {
				start = 1
			}
			windows = append(windows, codeWindow{startLine: start, endLine: end})
		}

		// Merge overlapping/adjacent windows
		merged := mergeWindows(windows)

		// Read file once
		absPath := safeJoinPath(repoRoot, path)
		if absPath == "" {
			continue
		}

		fileLines, err := readFileAllLines(absPath)
		if err != nil {
			continue
		}

		// Extract and output each merged window
		for _, window := range merged {
			if !isFirst {
				sb.WriteString("\n\n")
				totalChars += 2
			}
			isFirst = false

			// Clamp to file bounds
			if window.endLine > len(fileLines) {
				window.endLine = len(fileLines)
			}

			// Write header
			header := fmt.Sprintf("--- %s:%d-%d ---\n", path, window.startLine, window.endLine)
			sb.WriteString(header)
			totalChars += len(header)

			// Write lines
			for line := window.startLine; line <= window.endLine; line++ {
				if line > len(fileLines) {
					break
				}
				lineStr := fmt.Sprintf("%6d| %s\n", line, fileLines[line-1])

				if totalChars+len(lineStr) > maxCodeChars {
					remaining := maxCodeChars - totalChars
					if remaining > 0 {
						sb.WriteString(lineStr[:remaining])
					}
					sb.WriteString("\n... [truncated: output too large] ...")
					return sb.String(), nil
				}

				sb.WriteString(lineStr)
				totalChars += len(lineStr)
			}
		}
	}

	return sb.String(), nil
}

// mergeWindows merges overlapping or adjacent windows (no gap)
func mergeWindows(windows []codeWindow) []codeWindow {
	if len(windows) == 0 {
		return nil
	}

	// Sort by start line
	sort.Slice(windows, func(i, j int) bool {
		return windows[i].startLine < windows[j].startLine
	})

	merged := []codeWindow{windows[0]}

	for i := 1; i < len(windows); i++ {
		last := &merged[len(merged)-1]
		curr := windows[i]

		// Merge if overlapping or adjacent (no gap between them)
		if curr.startLine <= last.endLine+1 {
			if curr.endLine > last.endLine {
				last.endLine = curr.endLine
			}
		} else {
			merged = append(merged, curr)
		}
	}

	return merged
}

// safeJoinPath safely joins repoRoot and relPath, rejecting path traversal
func safeJoinPath(repoRoot, relPath string) string {
	// Clean the relative path
	cleanRel := filepath.Clean(relPath)

	// Reject if it tries to escape
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return ""
	}

	absPath := filepath.Join(repoRoot, cleanRel)

	// Double-check it's still under repoRoot
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return ""
	}
	absPath, err = filepath.Abs(absPath)
	if err != nil {
		return ""
	}

	if !strings.HasPrefix(absPath, absRoot) {
		return ""
	}

	return absPath
}

// readFileLines reads specific lines from a file (1-indexed, inclusive)
func readFileLines(path string, startLine, endLine int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	var sb strings.Builder
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		if lineNum >= startLine && lineNum <= endLine {
			sb.WriteString(fmt.Sprintf("%6d| %s\n", lineNum, scanner.Text()))
		}
		if lineNum > endLine {
			break
		}
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// readFileAllLines reads all lines from a file into a slice (0-indexed slice, but conceptually 1-indexed lines)
func readFileAllLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
