package memory

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// MesdxMeta is the top-level mesdx metadata stored in YAML frontmatter.
type MesdxMeta struct {
	ID         string      `yaml:"id" json:"id"`
	Scope      string      `yaml:"scope" json:"scope"`
	File       string      `yaml:"file,omitempty" json:"file,omitempty"`
	Title      string      `yaml:"title,omitempty" json:"title,omitempty"`
	Status     string      `yaml:"status" json:"status"`
	FileStatus string      `yaml:"fileStatus" json:"fileStatus"`
	Symbols    []SymbolRef `yaml:"symbols,omitempty" json:"symbols,omitempty"`
}

// SymbolRef is a reference to a code symbol stored in the frontmatter.
type SymbolRef struct {
	Language       string `yaml:"language" json:"language"`
	Name           string `yaml:"name" json:"name"`
	Status         string `yaml:"status" json:"status"`
	LastResolvedAt string `yaml:"lastResolvedAt,omitempty" json:"lastResolvedAt,omitempty"`
}

// frontmatterWrapper wraps MesdxMeta for YAML marshaling.
type frontmatterWrapper struct {
	Mesdx MesdxMeta `yaml:"mesdx"`
}

// ParseMarkdown parses a markdown file with YAML frontmatter.
// Returns the parsed metadata, the body text, and any error.
// If the frontmatter is not parsable, returns an error (caller should handle salvage).
func ParseMarkdown(data []byte) (*MesdxMeta, string, error) {
	fmRaw, body, hasFM := splitFrontmatter(string(data))
	if !hasFM {
		return nil, string(data), fmt.Errorf("no frontmatter found")
	}

	var wrapper frontmatterWrapper
	if err := yaml.Unmarshal([]byte(fmRaw), &wrapper); err != nil {
		return nil, string(data), fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	meta := &wrapper.Mesdx
	if meta.ID == "" {
		return nil, string(data), fmt.Errorf("missing mesdx.id in frontmatter")
	}

	// Apply defaults for missing fields
	if meta.Scope == "" {
		meta.Scope = "project"
	}
	if meta.Status == "" {
		meta.Status = "active"
	}
	if meta.FileStatus == "" {
		meta.FileStatus = "active"
	}
	for i := range meta.Symbols {
		if meta.Symbols[i].Status == "" {
			meta.Symbols[i].Status = "active"
		}
	}

	return meta, body, nil
}

// WriteMarkdown writes a markdown file with YAML frontmatter.
func WriteMarkdown(meta *MesdxMeta, body string) ([]byte, error) {
	wrapper := frontmatterWrapper{Mesdx: *meta}
	fmBytes, err := yaml.Marshal(&wrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fmBytes)
	buf.WriteString("---\n")
	if body != "" {
		buf.WriteString("\n")
		buf.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			buf.WriteString("\n")
		}
	}

	return buf.Bytes(), nil
}

// splitFrontmatter splits markdown content into frontmatter YAML and body.
func splitFrontmatter(content string) (fmRaw string, body string, hasFM bool) {
	const sep = "---"

	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != sep {
		return "", content, false
	}

	// Find closing separator (start from line 1)
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == sep {
			fmRaw = strings.Join(lines[1:i], "\n")
			body = strings.Join(lines[i+1:], "\n")
			body = strings.TrimLeft(body, "\n")
			return fmRaw, body, true
		}
	}

	return "", content, false
}

// NewMemoryID generates a new stable memory ID.
func NewMemoryID() string {
	return uuid.New().String()
}

// NewMeta creates a new MesdxMeta with defaults.
func NewMeta(scope, filePath, title string, symbols []SymbolRef) *MesdxMeta {
	meta := &MesdxMeta{
		ID:         NewMemoryID(),
		Scope:      scope,
		File:       filePath,
		Title:      title,
		Status:     "active",
		FileStatus: "active",
		Symbols:    symbols,
	}

	// Initialize symbol statuses
	now := time.Now().UTC().Format(time.RFC3339)
	for i := range meta.Symbols {
		if meta.Symbols[i].Status == "" {
			meta.Symbols[i].Status = "active"
		}
		meta.Symbols[i].LastResolvedAt = now
	}

	return meta
}
