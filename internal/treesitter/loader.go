package treesitter

import (
	"fmt"
	"sync"
	"unsafe"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

// Language wraps a tree-sitter language.
type Language struct {
	lang *tree_sitter.Language
	name string
}

var (
	languageCache = make(map[string]*Language)
	cacheMu       sync.RWMutex
)

// languageMap contains all statically compiled language parsers.
var languageMap = map[string]func() unsafe.Pointer{
	"go":         tree_sitter_go.Language,
	"java":       tree_sitter_java.Language,
	"rust":       tree_sitter_rust.Language,
	"python":     tree_sitter_python.Language,
	"javascript": tree_sitter_javascript.Language,
	"typescript": tree_sitter_typescript.LanguageTypescript,
	"tsx":        tree_sitter_typescript.LanguageTSX,
}

// LoadLanguage loads a tree-sitter language by name.
// The name should be the language identifier (e.g., "go", "python", "typescript").
func LoadLanguage(langName string) (*Language, error) {
	cacheMu.RLock()
	if lang, ok := languageCache[langName]; ok {
		cacheMu.RUnlock()
		return lang, nil
	}
	cacheMu.RUnlock()

	// Get the language function from the static map
	langFunc, ok := languageMap[langName]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", langName)
	}

	// Call the language function to get the pointer
	langPtr := langFunc()
	if langPtr == nil {
		return nil, fmt.Errorf("failed to get language pointer for %s", langName)
	}

	// Create tree-sitter language wrapper
	tsLang := tree_sitter.NewLanguage(langPtr)

	lang := &Language{
		lang: tsLang,
		name: langName,
	}

	cacheMu.Lock()
	languageCache[langName] = lang
	cacheMu.Unlock()

	return lang, nil
}

// TSLanguage returns the underlying tree-sitter language.
func (l *Language) TSLanguage() *tree_sitter.Language {
	return l.lang
}

// Name returns the language name.
func (l *Language) Name() string {
	return l.name
}

// Close is a no-op for statically linked languages.
// Kept for API compatibility.
func (l *Language) Close() error {
	return nil
}

// CloseAll clears the language cache.
// Kept for API compatibility.
func CloseAll() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	languageCache = make(map[string]*Language)
}

// VerifyLanguages checks that all required languages are available.
// With static linking, this just checks against the language map.
func VerifyLanguages(requiredLangs []string) error {
	var missing []string
	for _, langName := range requiredLangs {
		if _, ok := languageMap[langName]; !ok {
			missing = append(missing, langName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("unsupported languages: %v", missing)
	}

	return nil
}

// RequiredLanguages returns the list of language identifiers required by MesDX.
func RequiredLanguages() []string {
	return []string{"go", "java", "rust", "python", "javascript", "typescript"}
}
