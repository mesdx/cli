package indexer

import (
	"path/filepath"
	"strings"
)

// Lang represents a supported programming language.
type Lang string

const (
	LangGo         Lang = "go"
	LangJava       Lang = "java"
	LangRust       Lang = "rust"
	LangPython     Lang = "python"
	LangTypeScript Lang = "typescript"
	LangJavaScript Lang = "javascript"
	LangUnknown    Lang = ""
)

// extMap maps file extensions to languages.
var extMap = map[string]Lang{
	".go":   LangGo,
	".java": LangJava,
	".rs":   LangRust,
	".py":   LangPython,
	".pyi":  LangPython,
	".ts":   LangTypeScript,
	".tsx":  LangTypeScript,
	".js":   LangJavaScript,
	".jsx":  LangJavaScript,
	".mjs":  LangJavaScript,
	".cjs":  LangJavaScript,
	".mts":  LangTypeScript,
	".cts":  LangTypeScript,
}

// DetectLang returns the language for a given file path based on extension.
func DetectLang(path string) Lang {
	ext := strings.ToLower(filepath.Ext(path))
	if l, ok := extMap[ext]; ok {
		return l
	}
	return LangUnknown
}

// SupportedExtensions returns all file extensions we index.
func SupportedExtensions() []string {
	exts := make([]string, 0, len(extMap))
	for e := range extMap {
		exts = append(exts, e)
	}
	return exts
}
