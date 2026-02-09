# Dependency Graph for Navigator (struct)

**Location**: internal/indexer/navigation.go:38:5
**Signature**: `type Navigator struct`

**Definition Candidates**: 1
- [1] Navigator (struct) at internal/indexer/navigation.go:38

## File Dependency Graph

```mermaid
graph LR
    F0["internal/cli/mcp.go"]
    F1["internal/indexer/navigation.go"]
    F0 -->|"0.95 (12)"| F1
```

## Symbol Dependency Graph

**Nodes**: 8 | **Edges**: 15 (12 inbound, 3 outbound)

```mermaid
graph TD
    internal_indexer_navigation_go_Navigator_38[["Navigator (struct)"]]
    style internal_indexer_navigation_go_Navigator_38 fill:#f9f,stroke:#333,stroke-width:3px
    internal_cli_mcp_go["internal/cli/mcp.go"]
    internal_cli_mcp_go -->|"0.95"| internal_indexer_navigation_go_Navigator_38
    internal_indexer_navigation_go_Navigator_38 -->|"0.87"| internal_indexer_store_go_Store_12
    internal_indexer_store_go_Store_12["Store"]
    internal_indexer_navigation_go_Navigator_38 -->|"0.92"| database_sql_DB_0
    database_sql_DB_0["DB"]
    internal_indexer_navigation_go_Navigator_38 -->|"0.85"| internal_indexer_navigation_go_GoToDefinitionByName_43
    internal_indexer_navigation_go_GoToDefinitionByName_43["GoToDefinitionByName"]
```

## Top Scored Usages

**Total**: 12 usages

1. **Navigator** at `internal/cli/mcp.go:141` (score: 0.9500)
   - Context: runMcp
2. **Navigator** at `internal/cli/mcp.go:265` (score: 0.9500)
   - Context: runMcp
3. **Navigator** at `internal/cli/mcp.go:295` (score: 0.9200)
   - Context: runMcp
4. **Navigator** at `internal/cli/mcp.go:302` (score: 0.9200)
   - Context: runMcp
5. **Navigator** at `internal/cli/mcp.go:326` (score: 0.8900)
   - Context: runMcp
6. **Navigator** at `internal/cli/mcp.go:334` (score: 0.8900)
   - Context: runMcp
7. **Navigator** at `internal/cli/mcp.go:370` (score: 0.8700)
   - Context: runMcp
8. **Navigator** at `internal/cli/mcp.go:385` (score: 0.8700)
   - Context: runMcp
9. **Navigator** at `internal/cli/mcp.go:404` (score: 0.8500)
   - Context: runMcp
10. **Navigator** at `internal/cli/mcp.go:425` (score: 0.8200)
   - Context: runMcp

*... and 2 more usages*

---

# Example: Complex Dependency Graph

This shows what the output looks like for a symbol with many dependencies:

## File Dependency Graph

```mermaid
graph LR
    F0["internal/cli/mcp.go"]
    F1["internal/indexer/navigation.go"]
    F2["internal/indexer/depscore.go"]
    F3["cmd/mesdx/main.go"]
    F0 -->|"0.92 (8)"| F1
    F2 -->|"0.88 (5)"| F1
    F3 -->|"0.75 (2)"| F0
```

## Symbol Dependency Graph

```mermaid
graph TD
    pkg_myapp_go_ProcessData_42[["ProcessData (function)"]]
    style pkg_myapp_go_ProcessData_42 fill:#f9f,stroke:#333,stroke-width:3px
    pkg_api_handler_go["pkg/api/handler.go"]
    pkg_api_handler_go -->|"0.95"| pkg_myapp_go_ProcessData_42
    pkg_worker_processor_go["pkg/worker/process..."]
    pkg_worker_processor_go -->|"0.92"| pkg_myapp_go_ProcessData_42
    pkg_tests_integration_go["pkg/tests/integra..."]
    pkg_tests_integration_go -->|"0.88"| pkg_myapp_go_ProcessData_42
    more_inbound["... 12 more files"]
    more_inbound -.-> pkg_myapp_go_ProcessData_42
    pkg_myapp_go_ProcessData_42 -->|"0.91"| pkg_util_validate_go_Validate_10
    pkg_util_validate_go_Validate_10["Validate"]
    pkg_myapp_go_ProcessData_42 -->|"0.89"| pkg_db_query_go_QueryRows_25
    pkg_db_query_go_QueryRows_25["QueryRows"]
    pkg_myapp_go_ProcessData_42 -->|"0.87"| pkg_logger_log_go_Info_5
    pkg_logger_log_go_Info_5["Info"]
```

---

## Key Features of the Mermaid Output

1. **Two Graph Views**:
   - **File Dependency Graph** (LR - left to right): Shows which files depend on which files
   - **Symbol Dependency Graph** (TD - top to bottom): Shows the primary symbol with its inbound (who uses it) and outbound (what it uses) dependencies

2. **Visual Styling**:
   - Primary definition is highlighted with a thick border and distinctive color
   - Edge labels show dependency scores and usage counts
   - Nodes are automatically sanitized for Mermaid compatibility

3. **Smart Truncation**:
   - File paths are shortened for readability (max 30-40 chars)
   - Graphs are limited to top 15 edges by score to avoid clutter
   - "... N more" nodes indicate truncated content

4. **LLM-Native**:
   - Markdown formatting with sections
   - Mermaid diagrams can be directly rendered by most LLMs and markdown viewers
   - Structured data is also available in the JSON response for programmatic use

5. **Scored Usages**:
   - Listed with context container
   - Sorted by dependency score (descending)
   - Shows top 20 with summary count

This format makes it much easier for LLMs to:
- Visualize the dependency structure
- Understand risk when refactoring
- Identify high-confidence vs low-confidence dependencies
- See the full context of how a symbol is used across the codebase
