package scmsearch

import (
	"fmt"
	"strings"
)

// FormatResults produces a human-readable text summary of search results.
func FormatResults(r *SearchResult) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# SCM Search Results\n\n")
	fmt.Fprintf(&b, "**%d matches** across %d files (scanned %d files in %dms)\n",
		r.Summary.MatchesReturned, r.Summary.FilesMatched, r.Summary.FilesScanned, r.Summary.DurationMs)
	fmt.Fprintf(&b, "Cache: %d hits, %d misses\n\n", r.Summary.CacheHits, r.Summary.CacheMisses)

	if len(r.Matches) == 0 {
		b.WriteString("No matches found.\n")
		return b.String()
	}

	currentFile := ""
	for i, m := range r.Matches {
		if m.FilePath != currentFile {
			if currentFile != "" {
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "## %s\n\n", m.FilePath)
			currentFile = m.FilePath
		}

		fmt.Fprintf(&b, "%d. **@%s** `%s` at L%d:%d", i+1, m.CaptureName, m.NodeType, m.StartLine, m.StartCol)
		if m.EndLine != m.StartLine {
			fmt.Fprintf(&b, "–L%d:%d", m.EndLine, m.EndCol)
		}
		b.WriteString("\n")

		if m.TextSnippet != "" {
			fmt.Fprintf(&b, "   Text: `%s`\n", m.TextSnippet)
		}
		if len(m.ASTParents) > 0 {
			fmt.Fprintf(&b, "   AST: %s\n", strings.Join(m.ASTParents, " → "))
		}

		if len(m.ContextBefore) > 0 || m.Line != "" || len(m.ContextAfter) > 0 {
			b.WriteString("   ```\n")
			for _, l := range m.ContextBefore {
				fmt.Fprintf(&b, "   %s\n", l)
			}
			if m.Line != "" {
				fmt.Fprintf(&b, " → %s\n", m.Line)
			}
			for _, l := range m.ContextAfter {
				fmt.Fprintf(&b, "   %s\n", l)
			}
			b.WriteString("   ```\n")
		}
	}

	return b.String()
}
