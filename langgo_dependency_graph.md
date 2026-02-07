# LangGo Dependency Graph

## Overview
`LangGo` is a constant of type `Lang` defined in `internal/indexer/lang.go:12` with the value `"go"`.

## Dependency Graph

```mermaid
graph TD
    %% LangGo definition
    LangGo["LangGo<br/>(constant)<br/>lang.go:12"]
    
    %% What LangGo depends on
    LangType["Lang<br/>(type)<br/>lang.go:9"]
    
    %% What depends on LangGo
    LangGoFile["lang.go:23<br/>(extMap)"]
    ParserFile["parser.go:15<br/>(parserRegistry)"]
    DefinitionSpan["definition_span.go:114<br/>(isDocLine switch)"]
    DefinitionSpanTest["definition_span_test.go:769,813<br/>(FindDocStartLine tests)"]
    IndexerTest["indexer_test.go:270<br/>(DetectLang test)"]
    
    %% Dependencies (outbound)
    LangType -->|"type of"| LangGo
    
    %% Usages (inbound)
    LangGoFile -->|"uses"| LangGo
    ParserFile -->|"uses"| LangGo
    DefinitionSpan -->|"uses"| LangGo
    DefinitionSpanTest -->|"uses"| LangGo
    IndexerTest -->|"uses"| LangGo
    
    %% Styling
    classDef constant fill:#e1f5ff,stroke:#01579b,stroke-width:2px
    classDef type fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef usage fill:#f3e5f5,stroke:#4a148c,stroke-width:1px
    
    class LangGo constant
    class LangType type
    class LangGoFile,ParserFile,DefinitionSpan,DefinitionSpanTest,IndexerTest usage
```

## Detailed Usage Locations

### 1. **lang.go:23** - Extension Map
```go
var extMap = map[string]Lang{
    ".go":   LangGo,
    // ...
}
```
Maps the `.go` file extension to `LangGo`.

### 2. **parser.go:15** - Parser Registry
```go
parserRegistry[LangGo] = &GoParser{}
```
Registers the Go parser implementation for `LangGo`.

### 3. **definition_span.go:114** - Documentation Line Detection
```go
func isDocLine(trimmed string, lang Lang) bool {
    switch lang {
    case LangGo:
        return strings.HasPrefix(trimmed, "//") ||
            strings.HasPrefix(trimmed, "/*") ||
            // ...
    }
}
```
Used to detect Go-style documentation comments.

### 4. **definition_span_test.go:769,813** - Test Cases
Used in test cases for `FindDocStartLine` function with Go code examples.

### 5. **indexer_test.go:270** - Language Detection Test
```go
{"foo.go", LangGo},
```
Used in test cases for `DetectLang` function.

## Summary

- **Outbound Dependencies (1):**
  - `Lang` type (the type that `LangGo` is an instance of)

- **Inbound Dependencies (5 files, 6 usages):**
  - `lang.go` - Extension mapping
  - `parser.go` - Parser registration
  - `definition_span.go` - Documentation detection
  - `definition_span_test.go` - Test cases (2 usages)
  - `indexer_test.go` - Test cases

All usages have a dependency score of 1.0000, indicating strong, direct dependencies.
