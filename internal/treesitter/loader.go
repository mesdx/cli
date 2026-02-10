package treesitter

/*
#cgo CFLAGS: -I${SRCDIR}
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>
#include "tree_sitter_api.h"

// Type for tree-sitter language function
typedef const TSLanguage *(*ts_language_fn)(void);

static const TSLanguage *load_language(void *handle) {
    ts_language_fn fn = (ts_language_fn)dlsym(handle, "tree_sitter_language");
    if (fn == NULL) {
        return NULL;
    }
    return fn();
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

// Language maps to a tree-sitter language pointer.
type Language struct {
	handle *C.void
	lang   *C.TSLanguage
	name   string
}

var (
	languageCache = make(map[string]*Language)
	cacheMu       sync.RWMutex
)

// ParserDir returns the directory containing parser dynamic libraries.
// It checks in order:
// 1. MESDX_PARSER_DIR environment variable
// 2. exeDir/../lib/mesdx/parsers (installed alongside binary)
// 3. Returns error if none found
func ParserDir() (string, error) {
	// Check env var first
	if dir := os.Getenv("MESDX_PARSER_DIR"); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
		return "", fmt.Errorf("MESDX_PARSER_DIR is set to %q but directory does not exist", dir)
	}

	// Try relative to executable
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable symlinks: %w", err)
	}

	exeDir := filepath.Dir(exe)
	parserDir := filepath.Join(exeDir, "..", "lib", "mesdx", "parsers")
	parserDir, _ = filepath.Abs(parserDir)

	if _, err := os.Stat(parserDir); err == nil {
		return parserDir, nil
	}

	return "", fmt.Errorf("parser libraries not found; tried MESDX_PARSER_DIR and %s", parserDir)
}

// LoadLanguage loads a tree-sitter language library by name.
// The name should be the language identifier (e.g., "go", "python", "typescript").
func LoadLanguage(langName string) (*Language, error) {
	cacheMu.RLock()
	if lang, ok := languageCache[langName]; ok {
		cacheMu.RUnlock()
		return lang, nil
	}
	cacheMu.RUnlock()

	parserDir, err := ParserDir()
	if err != nil {
		return nil, err
	}

	// Determine library extension
	ext := ".so"
	if runtime.GOOS == "darwin" {
		ext = ".dylib"
	}

	libName := fmt.Sprintf("libtree-sitter-%s%s", langName, ext)
	libPath := filepath.Join(parserDir, libName)

	// Check if library exists
	if _, err := os.Stat(libPath); err != nil {
		return nil, fmt.Errorf("parser library not found: %s", libPath)
	}

	// Load the library
	cPath := C.CString(libPath)
	defer C.free(unsafe.Pointer(cPath))

	handle := C.dlopen(cPath, C.RTLD_NOW|C.RTLD_LOCAL)
	if handle == nil {
		errStr := C.GoString(C.dlerror())
		return nil, fmt.Errorf("failed to load %s: %s", libPath, errStr)
	}

	// Load the language function
	tsLang := C.load_language(handle)
	if tsLang == nil {
		errStr := C.GoString(C.dlerror())
		C.dlclose(handle)
		return nil, fmt.Errorf("failed to load tree_sitter_language from %s: %s", libPath, errStr)
	}

	lang := &Language{
		handle: handle,
		lang:   tsLang,
		name:   langName,
	}

	cacheMu.Lock()
	languageCache[langName] = lang
	cacheMu.Unlock()

	return lang, nil
}

// TSLanguage returns the underlying tree-sitter language pointer.
func (l *Language) TSLanguage() *C.TSLanguage {
	return l.lang
}

// Name returns the language name.
func (l *Language) Name() string {
	return l.name
}

// Close closes the language library handle.
func (l *Language) Close() error {
	if l.handle != nil {
		C.dlclose(l.handle)
		l.handle = nil
	}
	return nil
}

// CloseAll closes all loaded language libraries.
func CloseAll() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	for _, lang := range languageCache {
		_ = lang.Close()
	}
	languageCache = make(map[string]*Language)
}

// VerifyLanguages checks that all required language libraries are available.
// Returns a list of missing libraries or an error describing the issue.
func VerifyLanguages(requiredLangs []string) error {
	parserDir, err := ParserDir()
	if err != nil {
		return err
	}

	var missing []string
	for _, langName := range requiredLangs {
		ext := ".so"
		if runtime.GOOS == "darwin" {
			ext = ".dylib"
		}
		libName := fmt.Sprintf("libtree-sitter-%s%s", langName, ext)
		libPath := filepath.Join(parserDir, libName)

		if _, err := os.Stat(libPath); err != nil {
			missing = append(missing, langName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing parser libraries for languages: %s\n\nParser directory: %s\n\nPlease install the parser libraries. See README.md for installation instructions.",
			strings.Join(missing, ", "), parserDir)
	}

	return nil
}

// RequiredLanguages returns the list of language identifiers required by MesDX.
func RequiredLanguages() []string {
	return []string{"go", "java", "rust", "python", "javascript", "typescript"}
}
