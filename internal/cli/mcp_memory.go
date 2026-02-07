package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codeintelx/cli/internal/memory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- Arg structs ---

// MemoryAppendArgs is the input for memoryAppend.
type MemoryAppendArgs struct {
	Scope   string             `json:"scope"`
	File    string             `json:"file,omitempty"`
	Title   string             `json:"title,omitempty"`
	Content string             `json:"content"`
	Symbols []memory.SymbolRef `json:"symbols,omitempty"`
}

// MemoryReadArgs is the input for memoryRead.
type MemoryReadArgs struct {
	MemoryID  string `json:"memoryId,omitempty"`
	MdRelPath string `json:"mdRelPath,omitempty"`
	Scope     string `json:"scope,omitempty"`
	File      string `json:"file,omitempty"`
}

// MemoryUpdateArgs is the input for memoryUpdate.
type MemoryUpdateArgs struct {
	MemoryID string              `json:"memoryId"`
	Title    *string             `json:"title,omitempty"`
	Content  *string             `json:"content,omitempty"`
	Symbols  *[]memory.SymbolRef `json:"symbols,omitempty"`
}

// MemoryDeleteArgs is the input for memoryDelete.
type MemoryDeleteArgs struct {
	MemoryID string `json:"memoryId"`
}

// MemoryGrepReplaceArgs is the input for memoryGrepReplace.
type MemoryGrepReplaceArgs struct {
	MemoryID    string `json:"memoryId,omitempty"`
	MdRelPath   string `json:"mdRelPath,omitempty"`
	Scope       string `json:"scope,omitempty"`
	File        string `json:"file,omitempty"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

// MemorySearchArgs is the input for memorySearch.
type MemorySearchArgs struct {
	Query string `json:"query"`
	Scope string `json:"scope,omitempty"`
	File  string `json:"file,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// registerMemoryTools registers all memory MCP tools on the given server.
func registerMemoryTools(server *mcp.Server, mgr *memory.Manager) {
	// --- memoryAppend ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memoryAppend",
		Description: "Create a new memory element. Scope can be 'project' or 'file'. For file scope, provide the repo-relative file path. Content is the markdown body. Optionally attach symbol references.",
		InputSchema: memorySchema(MemoryAppendArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemoryAppendArgs) (*mcp.CallToolResult, any, error) {
		if args.Content == "" {
			return mcpError("content is required"), nil, nil
		}

		elem, err := mgr.Append(args.Scope, args.File, args.Title, args.Content, args.Symbols)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}

		text := fmt.Sprintf("Memory created: %s\nScope: %s\nPath: %s", elem.Meta.ID, elem.Meta.Scope, elem.MdRelPath)
		if elem.Meta.File != "" {
			text += fmt.Sprintf("\nFile: %s", elem.Meta.File)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, elem, nil
	})

	// --- memoryRead ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memoryRead",
		Description: "Read a memory element by ID or path. If neither is provided, lists memories matching the optional scope and file filters.",
		InputSchema: memorySchema(MemoryReadArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemoryReadArgs) (*mcp.CallToolResult, any, error) {
		// Single read by ID
		if args.MemoryID != "" {
			elem, err := mgr.Read(args.MemoryID)
			if err != nil {
				return mcpError(err.Error()), nil, nil
			}
			text := formatMemoryElement(elem)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, elem, nil
		}

		// Single read by path
		if args.MdRelPath != "" {
			elem, err := mgr.ReadByPath(args.MdRelPath)
			if err != nil {
				return mcpError(err.Error()), nil, nil
			}
			text := formatMemoryElement(elem)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, elem, nil
		}

		// List
		rows, err := mgr.List(args.Scope, args.File)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}

		text := formatMemoryList(rows)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, rows, nil
	})

	// --- memoryUpdate ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memoryUpdate",
		Description: "Update an existing memory element. Provide memoryId and any fields to update (title, content, symbols). Fields not provided are left unchanged.",
		InputSchema: memorySchema(MemoryUpdateArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemoryUpdateArgs) (*mcp.CallToolResult, any, error) {
		if args.MemoryID == "" {
			return mcpError("memoryId is required"), nil, nil
		}

		elem, err := mgr.Update(args.MemoryID, args.Title, args.Content, args.Symbols)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}

		text := fmt.Sprintf("Memory updated: %s\nPath: %s", elem.Meta.ID, elem.MdRelPath)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, elem, nil
	})

	// --- memoryDelete ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memoryDelete",
		Description: "Soft-delete a memory element. The file is preserved on disk with status set to 'deleted', but it is excluded from search results.",
		InputSchema: memorySchema(MemoryDeleteArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemoryDeleteArgs) (*mcp.CallToolResult, any, error) {
		if args.MemoryID == "" {
			return mcpError("memoryId is required"), nil, nil
		}

		if err := mgr.Delete(args.MemoryID); err != nil {
			return mcpError(err.Error()), nil, nil
		}

		text := fmt.Sprintf("Memory deleted (soft): %s", args.MemoryID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, map[string]interface{}{"deleted": args.MemoryID}, nil
	})

	// --- memoryGrepReplace ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memoryGrepReplace",
		Description: "Regex find-and-replace in a memory element's body. You must identify the target by memoryId or mdRelPath. If you provide scope/file filters instead and they match multiple memories, the tool will fail and return the candidate list for disambiguation.",
		InputSchema: memorySchema(MemoryGrepReplaceArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemoryGrepReplaceArgs) (*mcp.CallToolResult, any, error) {
		if args.Pattern == "" {
			return mcpError("pattern is required"), nil, nil
		}

		// If explicit target is given, proceed directly
		if args.MemoryID != "" || args.MdRelPath != "" {
			result, err := mgr.GrepReplace(args.MemoryID, args.MdRelPath, args.Pattern, args.Replacement)
			if err != nil {
				return mcpError(err.Error()), nil, nil
			}
			text := fmt.Sprintf("Grep/replace on %s: %d replacement(s)", result.MdRelPath, result.Replacements)
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: text}},
			}, result, nil
		}

		// No explicit target — resolve via scope/file filters and enforce single-target
		rows, err := mgr.List(args.Scope, args.File)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}

		// Filter out deleted
		var active []memory.MemoryRow
		for _, r := range rows {
			if r.Status != "deleted" && r.FileStatus != "deleted" {
				active = append(active, r)
			}
		}

		if len(active) == 0 {
			return mcpError("no matching memories found"), nil, nil
		}
		if len(active) > 1 {
			// Ambiguity — fail with candidates
			text := formatAmbiguousMemories(active)
			return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: text}},
					IsError: true,
				}, map[string]interface{}{
					"ambiguous":  true,
					"candidates": active,
					"hint":       "Provide a specific memoryId or mdRelPath to disambiguate.",
				}, nil
		}

		// Exactly one — proceed
		result, err := mgr.GrepReplace(active[0].MemoryUID, "", args.Pattern, args.Replacement)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}
		text := fmt.Sprintf("Grep/replace on %s: %d replacement(s)", result.MdRelPath, result.Replacements)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, result, nil
	})

	// --- memorySearch ---
	mcp.AddTool(server, &mcp.Tool{
		Name:        "codeintelx.memorySearch",
		Description: "Search memory elements using ngram-based text matching. Returns ranked results. Deleted memories and memories referencing deleted files are excluded.",
		InputSchema: memorySchema(MemorySearchArgs{}),
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MemorySearchArgs) (*mcp.CallToolResult, any, error) {
		if args.Query == "" {
			return mcpError("query is required"), nil, nil
		}

		results, err := mgr.Search(args.Query, args.Scope, args.File, args.Limit)
		if err != nil {
			return mcpError(err.Error()), nil, nil
		}

		text := formatSearchResults(results)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: text}},
		}, results, nil
	})
}

// --- Schema builder extension ---

func memorySchema(v interface{}) json.RawMessage {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
	props := schema["properties"].(map[string]interface{})

	switch v.(type) {
	case MemoryAppendArgs:
		schema["required"] = []string{"content"}
		props["scope"] = map[string]interface{}{
			"type":        "string",
			"description": "Memory scope: 'project' (default) or 'file'",
			"enum":        []string{"project", "file"},
			"default":     "project",
		}
		props["file"] = map[string]interface{}{
			"type":        "string",
			"description": "Repo-relative file path (required when scope is 'file')",
		}
		props["title"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional title for the memory",
		}
		props["content"] = map[string]interface{}{
			"type":        "string",
			"description": "Markdown body content of the memory",
		}
		props["symbols"] = map[string]interface{}{
			"type":        "array",
			"description": "Optional symbol references to attach",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{"type": "string"},
					"name":     map[string]interface{}{"type": "string"},
				},
				"required": []string{"language", "name"},
			},
		}

	case MemoryReadArgs:
		props["memoryId"] = map[string]interface{}{
			"type":        "string",
			"description": "Memory ID to read",
		}
		props["mdRelPath"] = map[string]interface{}{
			"type":        "string",
			"description": "Markdown file path relative to memory dir",
		}
		props["scope"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter by scope when listing",
			"enum":        []string{"project", "file"},
		}
		props["file"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter by repo-relative file path when listing",
		}

	case MemoryUpdateArgs:
		schema["required"] = []string{"memoryId"}
		props["memoryId"] = map[string]interface{}{
			"type":        "string",
			"description": "Memory ID to update",
		}
		props["title"] = map[string]interface{}{
			"type":        "string",
			"description": "New title (omit to keep existing)",
		}
		props["content"] = map[string]interface{}{
			"type":        "string",
			"description": "New markdown body (omit to keep existing)",
		}
		props["symbols"] = map[string]interface{}{
			"type":        "array",
			"description": "New symbol references (omit to keep existing, provide empty array to clear)",
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"language": map[string]interface{}{"type": "string"},
					"name":     map[string]interface{}{"type": "string"},
				},
				"required": []string{"language", "name"},
			},
		}

	case MemoryDeleteArgs:
		schema["required"] = []string{"memoryId"}
		props["memoryId"] = map[string]interface{}{
			"type":        "string",
			"description": "Memory ID to soft-delete",
		}

	case MemoryGrepReplaceArgs:
		schema["required"] = []string{"pattern", "replacement"}
		props["memoryId"] = map[string]interface{}{
			"type":        "string",
			"description": "Target memory by ID (preferred for disambiguation)",
		}
		props["mdRelPath"] = map[string]interface{}{
			"type":        "string",
			"description": "Target memory by markdown file path",
		}
		props["scope"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter scope (used only when no explicit target; must resolve to exactly 1)",
		}
		props["file"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter by file (used only when no explicit target; must resolve to exactly 1)",
		}
		props["pattern"] = map[string]interface{}{
			"type":        "string",
			"description": "Regex pattern to search for in the memory body",
		}
		props["replacement"] = map[string]interface{}{
			"type":        "string",
			"description": "Replacement string (supports regex backreferences like $1)",
		}

	case MemorySearchArgs:
		schema["required"] = []string{"query"}
		props["query"] = map[string]interface{}{
			"type":        "string",
			"description": "Search query text",
		}
		props["scope"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter results by scope",
			"enum":        []string{"project", "file"},
		}
		props["file"] = map[string]interface{}{
			"type":        "string",
			"description": "Filter results by repo-relative file path",
		}
		props["limit"] = map[string]interface{}{
			"type":        "integer",
			"description": "Maximum number of results (default: 20)",
			"default":     20,
			"minimum":     1,
			"maximum":     100,
		}
	}

	data, _ := json.Marshal(schema)
	return data
}

// --- Formatters ---

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Error: %s", msg)},
		},
		IsError: true,
	}
}

func formatMemoryElement(elem *memory.MemoryElement) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Memory: %s\n\n", elem.Meta.ID)
	fmt.Fprintf(&b, "- **Scope**: %s\n", elem.Meta.Scope)
	if elem.Meta.File != "" {
		fmt.Fprintf(&b, "- **File**: %s\n", elem.Meta.File)
	}
	if elem.Meta.Title != "" {
		fmt.Fprintf(&b, "- **Title**: %s\n", elem.Meta.Title)
	}
	fmt.Fprintf(&b, "- **Status**: %s\n", elem.Meta.Status)
	fmt.Fprintf(&b, "- **File Status**: %s\n", elem.Meta.FileStatus)
	fmt.Fprintf(&b, "- **Path**: %s\n", elem.MdRelPath)

	if len(elem.Meta.Symbols) > 0 {
		b.WriteString("\n**Symbols**:\n")
		for _, sym := range elem.Meta.Symbols {
			fmt.Fprintf(&b, "- `%s` (%s) [%s]\n", sym.Name, sym.Language, sym.Status)
		}
	}

	if elem.Body != "" {
		b.WriteString("\n---\n\n")
		b.WriteString(elem.Body)
	}
	return b.String()
}

func formatMemoryList(rows []memory.MemoryRow) string {
	if len(rows) == 0 {
		return "No memories found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d memory element(s):\n\n", len(rows))
	for i, r := range rows {
		fmt.Fprintf(&b, "%d. **%s** [%s] scope=%s", i+1, r.MemoryUID, r.Status, r.Scope)
		if r.FilePath != "" {
			fmt.Fprintf(&b, " file=%s", r.FilePath)
		}
		if r.Title != "" {
			fmt.Fprintf(&b, " title=%q", r.Title)
		}
		fmt.Fprintf(&b, "\n   path: %s\n", r.MdRelPath)
	}
	return b.String()
}

func formatSearchResults(results []memory.SearchResult) string {
	if len(results) == 0 {
		return "No results found."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Found %d result(s):\n\n", len(results))
	for i, r := range results {
		fmt.Fprintf(&b, "%d. **%s** (score: %.2f) scope=%s", i+1, r.MemoryUID, r.Score, r.Scope)
		if r.FilePath != "" {
			fmt.Fprintf(&b, " file=%s", r.FilePath)
		}
		if r.Title != "" {
			fmt.Fprintf(&b, " title=%q", r.Title)
		}
		fmt.Fprintf(&b, "\n   path: %s\n", r.MdRelPath)
	}
	return b.String()
}

func formatAmbiguousMemories(rows []memory.MemoryRow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Error: ambiguous target — %d memories matched. Provide a specific memoryId or mdRelPath.\n\n", len(rows))
	for i, r := range rows {
		fmt.Fprintf(&b, "%d. id=%s scope=%s", i+1, r.MemoryUID, r.Scope)
		if r.FilePath != "" {
			fmt.Fprintf(&b, " file=%s", r.FilePath)
		}
		if r.Title != "" {
			fmt.Fprintf(&b, " title=%q", r.Title)
		}
		fmt.Fprintf(&b, "\n   path: %s status=%s\n", r.MdRelPath, r.Status)
	}
	return b.String()
}
