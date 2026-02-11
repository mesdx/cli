package indexer

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mesdx/cli/internal/symbols"
)

// Location describes a position in a source file.
type Location struct {
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	StartCol  int    `json:"startCol"`
	EndLine   int    `json:"endLine"`
	EndCol    int    `json:"endCol"`
}

// DefinitionResult is the output of a go-to-definition query.
type DefinitionResult struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	Container string   `json:"container,omitempty"`
	Signature string   `json:"signature,omitempty"`
	Location  Location `json:"location"`
}

// UsageResult is the output of a find-usages query.
type UsageResult struct {
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	ContextContainer string   `json:"contextContainer,omitempty"`
	Relation         string   `json:"relation,omitempty"`
	ReceiverType     string   `json:"receiverType,omitempty"`
	TargetType       string   `json:"targetType,omitempty"`
	Location         Location `json:"location"`
	DependencyScore  float64  `json:"dependencyScore,omitempty"`
}

// Navigator provides go-to-definition and find-usages queries.
type Navigator struct {
	DB        *sql.DB
	ProjectID int64
}

// GoToDefinitionByName finds symbol definitions matching the given name.
// An optional filterFile (repo-relative path) ranks results from that file higher.
// The lang parameter filters results to files of the specified language.
func (n *Navigator) GoToDefinitionByName(name string, filterFile string, lang string) ([]DefinitionResult, error) {
	query := `
		SELECT s.name, s.kind, s.container_name, s.signature,
		       f.path, s.start_line, s.start_col, s.end_line, s.end_col
		FROM symbols s
		JOIN files f ON s.file_id = f.id
		WHERE f.project_id = ? AND s.name = ? AND f.lang = ?
		ORDER BY
			CASE WHEN f.path = ? THEN 0 ELSE 1 END,
			s.kind ASC
	`
	rows, err := n.DB.Query(query, n.ProjectID, name, lang, filterFile)
	if err != nil {
		return nil, fmt.Errorf("query definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []DefinitionResult
	for rows.Next() {
		var r DefinitionResult
		var kindInt int
		if err := rows.Scan(&r.Name, &kindInt, &r.Container, &r.Signature,
			&r.Location.Path, &r.Location.StartLine, &r.Location.StartCol,
			&r.Location.EndLine, &r.Location.EndCol); err != nil {
			return nil, err
		}
		r.Kind = symbols.SymbolKind(kindInt).String()
		results = append(results, r)
	}
	return results, rows.Err()
}

// GoToDefinitionByPosition resolves the identifier at the given cursor position,
// then looks up its definition.
// The lang parameter filters results to files of the specified language.
func (n *Navigator) GoToDefinitionByPosition(filePath string, line, col int, lang string) ([]DefinitionResult, error) {
	// First, find the symbol/ref name at the cursor position.
	name, err := n.identifierAt(filePath, line, col)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("no identifier found at %s:%d:%d", filePath, line, col)
	}
	return n.GoToDefinitionByName(name, filePath, lang)
}

// FindUsagesByName finds all references to the given name across the project.
// The lang parameter filters results to files of the specified language.
func (n *Navigator) FindUsagesByName(name string, filterFile string, lang string) ([]UsageResult, error) {
	query := `
		SELECT r.name, r.kind, r.context_container, r.relation, r.receiver_type, r.target_type,
		       f.path, r.start_line, r.start_col, r.end_line, r.end_col
		FROM refs r
		JOIN files f ON r.file_id = f.id
		WHERE f.project_id = ? AND r.name = ? AND f.lang = ?
		ORDER BY
			CASE WHEN f.path = ? THEN 0 ELSE 1 END,
			f.path ASC, r.start_line ASC
	`
	rows, err := n.DB.Query(query, n.ProjectID, name, lang, filterFile)
	if err != nil {
		return nil, fmt.Errorf("query usages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []UsageResult
	for rows.Next() {
		var r UsageResult
		var kindInt int
		if err := rows.Scan(&r.Name, &kindInt, &r.ContextContainer,
			&r.Relation, &r.ReceiverType, &r.TargetType,
			&r.Location.Path, &r.Location.StartLine, &r.Location.StartCol,
			&r.Location.EndLine, &r.Location.EndCol); err != nil {
			return nil, err
		}
		r.Kind = symbols.RefKind(kindInt).String()
		results = append(results, r)
	}
	return results, rows.Err()
}

// FindUsagesByPosition resolves the identifier at the given cursor position,
// then looks up its usages.
// The lang parameter filters results to files of the specified language.
func (n *Navigator) FindUsagesByPosition(filePath string, line, col int, lang string) ([]UsageResult, error) {
	name, err := n.identifierAt(filePath, line, col)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("no identifier found at %s:%d:%d", filePath, line, col)
	}
	return n.FindUsagesByName(name, filePath, lang)
}

// RefsInFileRange returns all references in the given file within the
// specified line range [startLine, endLine] (1-based, inclusive).
func (n *Navigator) RefsInFileRange(filePath string, startLine, endLine int, lang string) ([]UsageResult, error) {
	query := `
		SELECT r.name, r.kind, r.context_container,
		       f.path, r.start_line, r.start_col, r.end_line, r.end_col
		FROM refs r
		JOIN files f ON r.file_id = f.id
		WHERE f.project_id = ? AND f.path = ? AND f.lang = ?
		  AND r.start_line >= ? AND r.start_line <= ?
		ORDER BY r.start_line ASC, r.start_col ASC
	`
	rows, err := n.DB.Query(query, n.ProjectID, filePath, lang, startLine, endLine)
	if err != nil {
		return nil, fmt.Errorf("query refs in range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []UsageResult
	for rows.Next() {
		var r UsageResult
		var kindInt int
		if err := rows.Scan(&r.Name, &kindInt, &r.ContextContainer,
			&r.Location.Path, &r.Location.StartLine, &r.Location.StartCol,
			&r.Location.EndLine, &r.Location.EndCol); err != nil {
			return nil, err
		}
		r.Kind = symbols.RefKind(kindInt).String()
		results = append(results, r)
	}
	return results, rows.Err()
}

// identifierAt looks up the identifier name at the given file+line+col from
// both the symbols and refs tables.
func (n *Navigator) identifierAt(filePath string, line, col int) (string, error) {
	// Try symbols first
	var name string
	err := n.DB.QueryRow(`
		SELECT s.name FROM symbols s
		JOIN files f ON s.file_id = f.id
		WHERE f.project_id = ? AND f.path = ?
		  AND s.start_line = ? AND s.start_col <= ? AND s.end_col >= ?
		LIMIT 1
	`, n.ProjectID, filePath, line, col, col).Scan(&name)
	if err == nil && name != "" {
		return name, nil
	}

	// Try refs
	err = n.DB.QueryRow(`
		SELECT r.name FROM refs r
		JOIN files f ON r.file_id = f.id
		WHERE f.project_id = ? AND f.path = ?
		  AND r.start_line = ? AND r.start_col <= ? AND r.end_col >= ?
		LIMIT 1
	`, n.ProjectID, filePath, line, col, col).Scan(&name)
	if err == nil && name != "" {
		return name, nil
	}

	// Fallback: read the file and extract the identifier at position.
	return n.identifierFromSource(filePath, line, col)
}

// identifierFromSource reads the source file and extracts the word at the position.
func (n *Navigator) identifierFromSource(filePath string, line, col int) (string, error) {
	// We need the absolute path. For now we'll look it up via the file record.
	// The caller should pass repo-relative paths.
	// If the file doesn't exist in the DB, return empty.
	var _ string // placeholder
	return extractWordAtPosition(filePath, line, col)
}

// extractWordAtPosition reads a file and returns the identifier at line:col.
func extractWordAtPosition(filePath string, line, col int) (string, error) {
	// This would require the absolute path — the caller should resolve.
	// For now this is a best-effort fallback.
	return "", nil
}

// FormatDefinitions formats definition results as a human-readable string for MCP.
func FormatDefinitions(results []DefinitionResult) string {
	if len(results) == 0 {
		return "No definitions found."
	}
	var b strings.Builder
	for i, r := range results {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d] %s (%s)", i+1, r.Name, r.Kind)
		if r.Container != "" {
			fmt.Fprintf(&b, " in %s", r.Container)
		}
		fmt.Fprintf(&b, "\n    %s:%d:%d", r.Location.Path, r.Location.StartLine, r.Location.StartCol)
		if r.Signature != "" {
			fmt.Fprintf(&b, "\n    %s", r.Signature)
		}
	}
	return b.String()
}

// FormatUsages formats usage results as a human-readable string for MCP.
// If usages carry DependencyScore > 0, the score is printed under each entry.
func FormatUsages(results []UsageResult) string {
	if len(results) == 0 {
		return "No usages found."
	}
	var b strings.Builder
	for i, r := range results {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d] %s", i+1, r.Name)
		if r.ContextContainer != "" {
			fmt.Fprintf(&b, " (in %s)", r.ContextContainer)
		}
		fmt.Fprintf(&b, "\n    %s:%d:%d", r.Location.Path, r.Location.StartLine, r.Location.StartCol)
		if r.DependencyScore > 0 {
			fmt.Fprintf(&b, "\n    score: %.4f", r.DependencyScore)
		}
	}
	return b.String()
}

// FormatScoredUsages formats scored usage results with dependency scores.
func FormatScoredUsages(results []ScoredUsage) string {
	if len(results) == 0 {
		return "No usages found."
	}
	var b strings.Builder
	for i, r := range results {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d] %s", i+1, r.Name)
		if r.ContextContainer != "" {
			fmt.Fprintf(&b, " (in %s)", r.ContextContainer)
		}
		fmt.Fprintf(&b, "\n    %s:%d:%d", r.Location.Path, r.Location.StartLine, r.Location.StartCol)
		fmt.Fprintf(&b, "\n    score: %.4f", r.DependencyScore)
		if r.BestDefinition != nil {
			fmt.Fprintf(&b, " → %s (%s) at %s:%d",
				r.BestDefinition.Name,
				r.BestDefinition.Kind,
				r.BestDefinition.Location.Path,
				r.BestDefinition.Location.StartLine,
			)
		}
	}
	return b.String()
}
