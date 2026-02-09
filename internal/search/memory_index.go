package search

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search/query"
)

// maxSnippetLen is the maximum length of a search result snippet before truncation.
const maxSnippetLen = 240

// MemoryIndex provides Bleve-backed full-text search over memory markdown files.
// It indexes each memory file as multiple documents, chunked by any markdown
// header (`#` through `######`). Each chunk becomes a separate document for better relevance.
type MemoryIndex struct {
	projectID int64

	indexPath    string
	manifestPath string

	idx      *BleveIndex
	manifest *memoryManifest
}

// MemoryChunkDoc is the indexed document shape for a single chunk of a memory file.
type MemoryChunkDoc struct {
	Kind string `json:"kind"`

	ProjectID  int64  `json:"projectId"`
	ProjectKey string `json:"projectKey"`
	MemoryUID  string `json:"memoryUid"`
	Scope      string `json:"scope"`
	FilePath   string `json:"filePath,omitempty"`
	MdRelPath  string `json:"mdRelPath"`
	Title      string `json:"title,omitempty"`

	Status     string `json:"status"`
	FileStatus string `json:"fileStatus"`

	ChunkOrdinal int    `json:"chunkOrdinal"`
	ChunkHeading string `json:"chunkHeading,omitempty"`
	ChunkText    string `json:"chunkText"`

	SymbolText string `json:"symbolText,omitempty"`
}

// NewMemoryIndex opens (or creates) a memory index rooted at baseDir.
// baseDir is typically `<repoRoot>/.mesdx/search`.
func NewMemoryIndex(projectID int64, baseDir string) (*MemoryIndex, error) {
	indexPath := filepath.Join(baseDir, "memory.bleve")
	manifestPath := filepath.Join(baseDir, "memory-manifest.json")

	m := memoryIndexMapping()

	idx, err := OpenOrCreate(indexPath, m)
	if err != nil {
		return nil, err
	}
	manifest, err := loadMemoryManifest(manifestPath)
	if err != nil {
		_ = idx.Close()
		return nil, err
	}

	return &MemoryIndex{
		projectID:    projectID,
		indexPath:    indexPath,
		manifestPath: manifestPath,
		idx:          idx,
		manifest:     manifest,
	}, nil
}

// NewMemoryIndexReadOnly opens an existing memory index in read-only mode.
// This allows searching while the MCP server has the index open for writing.
func NewMemoryIndexReadOnly(projectID int64, baseDir string) (*MemoryIndex, error) {
	indexPath := filepath.Join(baseDir, "memory.bleve")
	manifestPath := filepath.Join(baseDir, "memory-manifest.json")

	idx, err := OpenReadOnly(indexPath)
	if err != nil {
		return nil, err
	}
	manifest, err := loadMemoryManifest(manifestPath)
	if err != nil {
		_ = idx.Close()
		return nil, err
	}

	return &MemoryIndex{
		projectID:    projectID,
		indexPath:    indexPath,
		manifestPath: manifestPath,
		idx:          idx,
		manifest:     manifest,
	}, nil
}

// Reset clears the index and manifest, then recreates an empty index.
func (m *MemoryIndex) Reset() error {
	if m == nil {
		return nil
	}
	_ = m.idx.Close()

	b, err := Reset(m.indexPath, memoryIndexMapping())
	if err != nil {
		return err
	}
	m.idx = b
	m.manifest = &memoryManifest{ByMdRelPath: map[string][]string{}}
	return saveMemoryManifest(m.manifestPath, m.manifest)
}

func (m *MemoryIndex) Close() error {
	if m == nil {
		return nil
	}
	if err := saveMemoryManifest(m.manifestPath, m.manifest); err != nil {
		log.Printf("warning: failed to save memory manifest on close: %v", err)
	}
	return m.idx.Close()
}

// RemoveByMdRelPath deletes all indexed chunk documents for a memory markdown file.
func (m *MemoryIndex) RemoveByMdRelPath(mdRelPath string) error {
	if m == nil {
		return nil
	}
	docIDs := m.manifest.ByMdRelPath[mdRelPath]
	for _, id := range docIDs {
		_ = m.idx.Index.Delete(id)
	}
	delete(m.manifest.ByMdRelPath, mdRelPath)
	return saveMemoryManifest(m.manifestPath, m.manifest)
}

// IndexMemory indexes a single memory file as multiple chunk documents.
// The caller is responsible for ensuring the memory is "valid" (e.g. file-scoped
// reference exists) before calling this.
func (m *MemoryIndex) IndexMemory(memoryUID, scope, filePath, mdRelPath, title, status, fileStatus string, body string, symbols []string) error {
	if m == nil {
		return nil
	}
	if err := m.RemoveByMdRelPath(mdRelPath); err != nil {
		return err
	}

	chunks := ChunkByHeaders(body)
	symText := strings.Join(symbols, " ")

	batch := m.idx.Index.NewBatch()
	var docIDs []string
	for i, c := range chunks {
		docID := fmt.Sprintf("memory:%d:%s:%d", m.projectID, memoryUID, i)
		doc := MemoryChunkDoc{
			Kind: "memoryChunk",

			ProjectID:  m.projectID,
			ProjectKey: fmt.Sprintf("%d", m.projectID),
			MemoryUID:  memoryUID,
			Scope:      scope,
			FilePath:   filePath,
			MdRelPath:  mdRelPath,
			Title:      title,

			Status:     status,
			FileStatus: fileStatus,

			ChunkOrdinal: i,
			ChunkHeading: c.Heading,
			ChunkText:    c.Text,

			SymbolText: symText,
		}
		batch.Index(docID, doc) //nolint:errcheck
		docIDs = append(docIDs, docID)
	}

	if err := m.idx.Index.Batch(batch); err != nil {
		return fmt.Errorf("bleve batch: %w", err)
	}
	m.manifest.ByMdRelPath[mdRelPath] = docIDs
	return saveMemoryManifest(m.manifestPath, m.manifest)
}

type MemorySearchHit struct {
	MemoryUID string
	Scope     string
	FilePath  string
	MdRelPath string
	Title     string
	Score     float64
	Snippet   string
}

// Search searches the memory index, filtered by scope and filePath (optional).
// It always filters to the configured projectId and excludes deleted records.
func (m *MemoryIndex) Search(queryText, scope, filePath string, limit int) ([]MemorySearchHit, error) {
	if m == nil {
		return nil, nil
	}
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}

	// Relevance query: match across multiple fields with boosts.
	qTitle := bleve.NewMatchQuery(queryText)
	qTitle.SetField("title")
	qTitle.SetBoost(2.0)

	qHead := bleve.NewMatchQuery(queryText)
	qHead.SetField("chunkHeading")
	qHead.SetBoost(1.5)

	qBody := bleve.NewMatchQuery(queryText)
	qBody.SetField("chunkText")
	qBody.SetBoost(1.0)

	qSym := bleve.NewMatchQuery(queryText)
	qSym.SetField("symbolText")
	qSym.SetBoost(1.0)

	relQ := bleve.NewDisjunctionQuery(qTitle, qHead, qBody, qSym)

	// Filters.
	filters := []query.Query{
		bleve.NewTermQuery(fmt.Sprintf("%d", m.projectID)),
		bleve.NewTermQuery("active"),
		bleve.NewTermQuery("active"),
	}
	filters[0].(*query.TermQuery).SetField("projectKey")
	filters[1].(*query.TermQuery).SetField("status")
	filters[2].(*query.TermQuery).SetField("fileStatus")

	if scope != "" {
		tq := bleve.NewTermQuery(scope)
		tq.SetField("scope")
		filters = append(filters, tq)
	}
	if filePath != "" {
		tq := bleve.NewTermQuery(filePath)
		tq.SetField("filePath")
		filters = append(filters, tq)
	}

	finalQ := bleve.NewConjunctionQuery(append([]query.Query{relQ}, filters...)...)

	req := bleve.NewSearchRequestOptions(finalQ, limit, 0, false)
	req.Fields = []string{"memoryUid", "scope", "filePath", "mdRelPath", "title", "chunkText"}

	res, err := m.idx.Index.Search(req)
	if err != nil {
		return nil, fmt.Errorf("bleve search: %w", err)
	}

	// Deduplicate by memoryUid: keep the best chunk hit per memory.
	best := map[string]MemorySearchHit{}
	order := make([]string, 0, len(res.Hits))
	for _, h := range res.Hits {
		memUID, _ := h.Fields["memoryUid"].(string)
		if memUID == "" {
			continue
		}
		scopeV, _ := h.Fields["scope"].(string)
		fileV, _ := h.Fields["filePath"].(string)
		mdV, _ := h.Fields["mdRelPath"].(string)
		titleV, _ := h.Fields["title"].(string)
		chunkText, _ := h.Fields["chunkText"].(string)

		snippet := chunkText
		if len(snippet) > maxSnippetLen {
			snippet = snippet[:maxSnippetLen] + "…"
		}

		cur, ok := best[memUID]
		newHit := MemorySearchHit{
			MemoryUID: memUID,
			Scope:     scopeV,
			FilePath:  fileV,
			MdRelPath: mdV,
			Title:     titleV,
			Score:     h.Score,
			Snippet:   snippet,
		}
		if !ok {
			best[memUID] = newHit
			order = append(order, memUID)
			continue
		}
		if newHit.Score > cur.Score {
			best[memUID] = newHit
		}
	}

	out := make([]MemorySearchHit, 0, len(best))
	for _, id := range order {
		out = append(out, best[id])
	}
	return out, nil
}

// --- Chunking ---

type Chunk struct {
	Heading string
	Text    string
}

// headerPrefix returns the heading text if the line is a markdown header
// (# through ######), along with the full header prefix (e.g. "## ").
// Returns ("", "") if the line is not a header.
func headerPrefix(line string) (headingText string, prefix string) {
	trimmed := strings.TrimLeft(line, "#")
	hashes := len(line) - len(trimmed)
	if hashes < 1 || hashes > 6 {
		return "", ""
	}
	if !strings.HasPrefix(trimmed, " ") {
		return "", ""
	}
	return strings.TrimSpace(trimmed), line[:hashes] + " "
}

// ChunkByHeaders splits markdown by any header (`#` through `######`).
// Each chunk contains the header line (if present) and its following content,
// up to (but not including) the next header. Content before the first header
// becomes a preamble chunk with an empty heading.
func ChunkByHeaders(body string) []Chunk {
	lines := strings.Split(body, "\n")
	var chunks []Chunk
	var cur *Chunk

	flush := func() {
		if cur == nil {
			return
		}
		cur.Text = strings.TrimSpace(cur.Text)
		if cur.Text == "" {
			cur = nil
			return
		}
		chunks = append(chunks, *cur)
		cur = nil
	}

	startNew := func(heading, headerLine string) {
		flush()
		cur = &Chunk{Heading: heading}
		if headerLine != "" {
			cur.Text = headerLine
		}
	}

	// If the file starts without a header, keep a preamble chunk.
	startNew("", "")

	for _, line := range lines {
		if h, _ := headerPrefix(line); h != "" {
			startNew(h, line)
			continue
		}
		if cur == nil {
			startNew("", "")
		}
		if cur.Text != "" {
			cur.Text += "\n"
		}
		cur.Text += line
	}
	flush()

	// As a fallback (empty file), return no chunks.
	if len(chunks) == 1 && chunks[0].Heading == "" && strings.TrimSpace(chunks[0].Text) == "" {
		return nil
	}
	return chunks
}

// EnsureIndexDir exists for callers that want to pre-create directories.
func EnsureIndexDir(baseDir string) error {
	return os.MkdirAll(baseDir, 0755)
}

func memoryIndexMapping() mapping.IndexMapping {
	m := bleve.NewIndexMapping()
	m.TypeField = "kind"
	m.DefaultType = "memoryChunk"

	docMapping := mapping.NewDocumentMapping()

	text := mapping.NewTextFieldMapping()
	text.Store = true

	kw := mapping.NewKeywordFieldMapping()
	kw.Store = true

	num := mapping.NewNumericFieldMapping()
	num.Store = true

	// Fields used for filtering.
	docMapping.AddFieldMappingsAt("projectId", num)
	docMapping.AddFieldMappingsAt("projectKey", kw)
	docMapping.AddFieldMappingsAt("scope", kw)
	docMapping.AddFieldMappingsAt("filePath", kw)
	docMapping.AddFieldMappingsAt("mdRelPath", kw)
	docMapping.AddFieldMappingsAt("memoryUid", kw)
	docMapping.AddFieldMappingsAt("status", kw)
	docMapping.AddFieldMappingsAt("fileStatus", kw)

	// Fields used for relevance.
	docMapping.AddFieldMappingsAt("title", text)
	docMapping.AddFieldMappingsAt("chunkHeading", text)
	docMapping.AddFieldMappingsAt("chunkText", text)
	docMapping.AddFieldMappingsAt("symbolText", text)

	m.AddDocumentMapping("memoryChunk", docMapping)
	return m
}
