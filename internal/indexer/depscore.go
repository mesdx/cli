package indexer

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Scored usage types
// ---------------------------------------------------------------------------

// ScoredUsage extends UsageResult with dependency scoring information.
type ScoredUsage struct {
	UsageResult
	DependencyScore float64           `json:"dependencyScore"`
	BestDefinition  *DefinitionResult `json:"bestDefinition,omitempty"`
}

// ---------------------------------------------------------------------------
// Dependency graph types
// ---------------------------------------------------------------------------

// DepGraphNode represents a node in the dependency graph.
type DepGraphNode struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Path      string `json:"path"`
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Container string `json:"container,omitempty"`
	Signature string `json:"signature,omitempty"`
}

// DepGraphEdge represents a directed edge in the symbol dependency graph.
type DepGraphEdge struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Kind     string  `json:"kind"` // "inbound" or "outbound"
	Score    float64 `json:"score"`
	Count    int     `json:"count"`
	FilePath string  `json:"filePath,omitempty"`
}

// FileGraphEdge represents a directed edge in the file-level graph.
type FileGraphEdge struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Score float64 `json:"score"`
	Count int     `json:"count"`
}

// SymbolGraph contains the symbol-level dependency graph.
type SymbolGraph struct {
	Nodes []DepGraphNode `json:"nodes"`
	Edges []DepGraphEdge `json:"edges"`
}

// DependencyGraph is the full output of the dependency graph tool.
type DependencyGraph struct {
	PrimaryDefinition    *DefinitionResult  `json:"primaryDefinition"`
	DefinitionCandidates []DefinitionResult `json:"definitionCandidates"`
	SymbolGraph          SymbolGraph        `json:"symbolGraph"`
	FileGraph            []FileGraphEdge    `json:"fileGraph"`
	Usages               []ScoredUsage      `json:"usages,omitempty"`
}

// ---------------------------------------------------------------------------
// Scoring constants (feature weights)
// ---------------------------------------------------------------------------

const (
	boostSameFile       = 3.0
	boostSameDir        = 1.5
	boostContainerMatch = 1.5
	boostKindMatch      = 2.0 // lexical context matches definition kind
	boostRefKindMatch   = 2.5 // structured ref kind matches definition kind
	boostImportRef      = 3.0 // import references are very strong signals
	boostUniqueMin      = 1.0 // minimum uniqueness weight
)

// ---------------------------------------------------------------------------
// Lexical-context patterns (used for kind compatibility)
// ---------------------------------------------------------------------------

var (
	patNew       = regexp.MustCompile(`\bnew\s+\w+`)
	patCall      = regexp.MustCompile(`\w+\s*\(`)
	patExtends   = regexp.MustCompile(`\b(?:extends|implements)\s+\w+`)
	patTypeHint  = regexp.MustCompile(`(?::\s*\w+|->)\s*\w+`)
	patDotAccess = regexp.MustCompile(`\.\w+`)
	patInherit   = regexp.MustCompile(`\b(?:extends|implements|:\s*public|:\s*private|:\s*protected)\s+\w+`)
)

// ---------------------------------------------------------------------------
// Core scoring
// ---------------------------------------------------------------------------

// ScoreUsages computes a DependencyScore for each usage against the given
// candidate definitions. primaryDef may be nil (name-only lookups); when
// provided it biases the score toward that definition.
func ScoreUsages(
	usages []UsageResult,
	candidates []DefinitionResult,
	primaryDef *DefinitionResult,
	repoRoot string,
) []ScoredUsage {
	if len(candidates) == 0 {
		// No candidates → score 0 for all usages.
		scored := make([]ScoredUsage, len(usages))
		for i, u := range usages {
			scored[i] = ScoredUsage{UsageResult: u, DependencyScore: 0}
		}
		return scored
	}

	scored := make([]ScoredUsage, 0, len(usages))
	// Cache source lines by path to avoid repeated reads.
	lineCache := map[string]string{}

	for _, usage := range usages {
		best, bestDef := scoreOneUsage(usage, candidates, primaryDef, repoRoot, lineCache)
		su := ScoredUsage{
			UsageResult:     usage,
			DependencyScore: best,
		}
		if bestDef != nil {
			defCopy := *bestDef
			su.BestDefinition = &defCopy
		}
		scored = append(scored, su)
	}
	return scored
}

// scoreOneUsage computes the dependency score for a single usage against
// all candidate definitions and returns (score, bestDef).
func scoreOneUsage(
	usage UsageResult,
	candidates []DefinitionResult,
	primaryDef *DefinitionResult,
	repoRoot string,
	lineCache map[string]string,
) (float64, *DefinitionResult) {
	numCandidates := float64(len(candidates))

	// Read the source line for lexical-context analysis.
	srcLine := getSourceLine(repoRoot, usage.Location.Path, usage.Location.StartLine, lineCache)

	weights := make([]float64, len(candidates))
	for i, def := range candidates {
		w := boostUniqueMin

		// 1. Uniqueness prior — fewer candidates ⇒ higher base.
		w *= 1.0 / math.Sqrt(numCandidates)

		// 2. Same-file boost.
		if usage.Location.Path == def.Location.Path {
			w *= boostSameFile
		} else if sameDir(usage.Location.Path, def.Location.Path) {
			// 3. Same-directory boost (only if not same file).
			w *= boostSameDir
		}

		// 4. Container match boost.
		if usage.ContextContainer != "" && def.Container != "" &&
			usage.ContextContainer == def.Container {
			w *= boostContainerMatch
		}

		// 5. Kind compatibility from lexical context on the source line.
		w *= kindCompatibility(srcLine, usage.Name, def.Kind)

		// 6. Structured ref kind boost.
		w *= refKindCompatibility(usage.Kind, def.Kind)

		weights[i] = w
	}

	// Normalize weights → probabilities.
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		return 0, nil
	}

	probs := make([]float64, len(weights))
	for i, w := range weights {
		probs[i] = w / total
	}

	// Find best score.
	if primaryDef != nil {
		// Cursor-based: compute P(primaryDef | ref).
		for i, def := range candidates {
			if def.Location.Path == primaryDef.Location.Path &&
				def.Location.StartLine == primaryDef.Location.StartLine &&
				def.Location.StartCol == primaryDef.Location.StartCol {
				return round4(probs[i]), &candidates[i]
			}
		}
		// Primary not among candidates — fall through to best.
	}

	// Name-only: return max probability and its definition.
	bestIdx := 0
	for i := 1; i < len(probs); i++ {
		if probs[i] > probs[bestIdx] {
			bestIdx = i
		}
	}
	return round4(probs[bestIdx]), &candidates[bestIdx]
}

// ---------------------------------------------------------------------------
// Signal helpers
// ---------------------------------------------------------------------------

// sameDir checks if two repo-relative paths share the same parent directory.
func sameDir(a, b string) bool {
	return filepath.Dir(a) == filepath.Dir(b)
}

// kindCompatibility returns a multiplier based on whether the lexical context
// surrounding the usage name on its source line is compatible with the
// definition kind.
func kindCompatibility(srcLine, name, defKind string) float64 {
	if srcLine == "" {
		return 1.0 // no context → neutral
	}

	isType := defKind == "class" || defKind == "struct" || defKind == "interface" ||
		defKind == "enum" || defKind == "type_alias" || defKind == "trait"
	isCallable := defKind == "function" || defKind == "method" || defKind == "constructor"
	isMember := defKind == "field" || defKind == "property" || defKind == "method"

	// Check for `new Name`
	if patNew.MatchString(srcLine) && strings.Contains(srcLine, name) {
		if isType || defKind == "constructor" {
			return boostKindMatch
		}
		return 0.5 // unlikely
	}

	// Check for extends/implements
	if patExtends.MatchString(srcLine) && strings.Contains(srcLine, name) {
		if isType {
			return boostKindMatch
		}
		return 0.5
	}

	// Check for dot access `.Name`
	if strings.Contains(srcLine, "."+name) {
		if isMember {
			return boostKindMatch
		}
		return 0.8
	}

	// Check for call `Name(`
	callPat := name + "("
	if strings.Contains(srcLine, callPat) {
		if isCallable {
			return boostKindMatch
		}
		return 0.7
	}

	// Check for type hints `: Name` or `-> Name`
	if patTypeHint.MatchString(srcLine) && strings.Contains(srcLine, name) {
		if isType {
			return boostKindMatch * 0.8
		}
	}

	return 1.0
}

// refKindCompatibility returns a multiplier based on the structured ref kind
// (when available) and the definition kind.
func refKindCompatibility(refKind, defKind string) float64 {
	switch refKind {
	case "import":
		return boostImportRef // imports are very strong
	case "type_ref":
		if defKind == "class" || defKind == "struct" || defKind == "interface" ||
			defKind == "enum" || defKind == "type_alias" || defKind == "trait" {
			return boostRefKindMatch
		}
		return 0.6
	case "inherit":
		if defKind == "class" || defKind == "interface" || defKind == "trait" {
			return boostRefKindMatch
		}
		return 0.5
	case "call":
		if defKind == "function" || defKind == "method" || defKind == "constructor" {
			return boostRefKindMatch
		}
		return 0.6
	case "annotation":
		return 1.2
	default:
		return 1.0 // "other", "read", "write" → neutral
	}
}

// getSourceLine reads a single source line from the file. Results are cached.
func getSourceLine(repoRoot, relPath string, line int, cache map[string]string) string {
	key := relPath + ":" + itoa(line)
	if v, ok := cache[key]; ok {
		return v
	}

	absPath := safeJoinRepoPath(repoRoot, relPath)
	if absPath == "" {
		return ""
	}

	content, err := readSingleLine(absPath, line)
	if err != nil {
		return ""
	}
	cache[key] = content
	return content
}

// ---------------------------------------------------------------------------
// Adjacency-aware grouping and sorting
// ---------------------------------------------------------------------------

// usageGroup represents a group of adjacent usages in the same file.
type usageGroup struct {
	usages    []ScoredUsage
	maxScore  float64
	filePath  string
	startLine int
	endLine   int
}

// GroupAndSortUsages groups adjacent usages (same file, within gapLines of
// each other) and sorts groups descending by max dependency score, then
// within each group by line ascending.
func GroupAndSortUsages(usages []ScoredUsage, gapLines int) []ScoredUsage {
	if len(usages) == 0 {
		return usages
	}
	if gapLines <= 0 {
		gapLines = 3 // default adjacency gap
	}

	// Step 1: group by file then merge adjacent.
	byFile := map[string][]ScoredUsage{}
	for _, u := range usages {
		byFile[u.Location.Path] = append(byFile[u.Location.Path], u)
	}

	var groups []usageGroup
	for path, fileUsages := range byFile {
		// Sort by line within file.
		sort.Slice(fileUsages, func(i, j int) bool {
			return fileUsages[i].Location.StartLine < fileUsages[j].Location.StartLine
		})

		// Merge adjacent.
		g := usageGroup{
			usages:    []ScoredUsage{fileUsages[0]},
			maxScore:  fileUsages[0].DependencyScore,
			filePath:  path,
			startLine: fileUsages[0].Location.StartLine,
			endLine:   fileUsages[0].Location.EndLine,
		}
		for i := 1; i < len(fileUsages); i++ {
			u := fileUsages[i]
			if u.Location.StartLine <= g.endLine+gapLines {
				// Adjacent — merge into current group.
				g.usages = append(g.usages, u)
				if u.DependencyScore > g.maxScore {
					g.maxScore = u.DependencyScore
				}
				if u.Location.EndLine > g.endLine {
					g.endLine = u.Location.EndLine
				}
			} else {
				groups = append(groups, g)
				g = usageGroup{
					usages:    []ScoredUsage{u},
					maxScore:  u.DependencyScore,
					filePath:  path,
					startLine: u.Location.StartLine,
					endLine:   u.Location.EndLine,
				}
			}
		}
		groups = append(groups, g)
	}

	// Step 2: sort groups descending by maxScore.
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].maxScore != groups[j].maxScore {
			return groups[i].maxScore > groups[j].maxScore
		}
		// Tie-break: file path then start line.
		if groups[i].filePath != groups[j].filePath {
			return groups[i].filePath < groups[j].filePath
		}
		return groups[i].startLine < groups[j].startLine
	})

	// Step 3: flatten groups back into a slice.
	result := make([]ScoredUsage, 0, len(usages))
	for _, g := range groups {
		result = append(result, g.usages...)
	}
	return result
}

// ---------------------------------------------------------------------------
// Dependency graph builder
// ---------------------------------------------------------------------------

// BuildDependencyGraph constructs a symbol-level dependency graph and a
// collapsed file-level graph for the given primary definition.
func BuildDependencyGraph(
	nav *Navigator,
	primaryDef *DefinitionResult,
	candidates []DefinitionResult,
	lang string,
	repoRoot string,
	maxDepth int,
	minScore float64,
	maxUsages int,
) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		PrimaryDefinition:    primaryDef,
		DefinitionCandidates: candidates,
		SymbolGraph: SymbolGraph{
			Nodes: []DepGraphNode{},
			Edges: []DepGraphEdge{},
		},
		FileGraph: []FileGraphEdge{},
	}

	if primaryDef == nil {
		return graph, nil
	}

	// Add primary definition as a node.
	primaryNodeID := nodeID(primaryDef.Location.Path, primaryDef.Name, primaryDef.Location.StartLine)
	graph.SymbolGraph.Nodes = append(graph.SymbolGraph.Nodes, DepGraphNode{
		ID:        primaryNodeID,
		Name:      primaryDef.Name,
		Kind:      primaryDef.Kind,
		Path:      primaryDef.Location.Path,
		StartLine: primaryDef.Location.StartLine,
		EndLine:   primaryDef.Location.EndLine,
		Container: primaryDef.Container,
		Signature: primaryDef.Signature,
	})

	// -----------------------------------------------------------------------
	// Inbound edges: usages of this symbol → primaryDef
	// -----------------------------------------------------------------------
	usages, err := nav.FindUsagesByName(primaryDef.Name, "", lang)
	if err != nil {
		return graph, err
	}
	if maxUsages > 0 && len(usages) > maxUsages {
		usages = usages[:maxUsages]
	}

	scored := ScoreUsages(usages, candidates, primaryDef, repoRoot)
	scored = GroupAndSortUsages(scored, 3)

	// Filter by minScore.
	filtered := make([]ScoredUsage, 0, len(scored))
	for _, su := range scored {
		if su.DependencyScore >= minScore {
			filtered = append(filtered, su)
		}
	}
	graph.Usages = filtered

	// Aggregate inbound edges per file.
	inbound := map[string]*DepGraphEdge{}
	for _, su := range filtered {
		key := su.Location.Path
		if e, ok := inbound[key]; ok {
			e.Count++
			if su.DependencyScore > e.Score {
				e.Score = su.DependencyScore
			}
		} else {
			inbound[key] = &DepGraphEdge{
				From:     su.Location.Path,
				To:       primaryNodeID,
				Kind:     "inbound",
				Score:    su.DependencyScore,
				Count:    1,
				FilePath: su.Location.Path,
			}
		}
	}
	for _, e := range inbound {
		e.Score = round4(e.Score)
		graph.SymbolGraph.Edges = append(graph.SymbolGraph.Edges, *e)
	}

	// -----------------------------------------------------------------------
	// Outbound edges: refs inside the definition span → other symbols
	// -----------------------------------------------------------------------
	if maxDepth >= 1 {
		outbound, err := computeOutbound(nav, primaryDef, primaryNodeID, lang, repoRoot)
		if err == nil {
			for _, node := range outbound.nodes {
				graph.SymbolGraph.Nodes = append(graph.SymbolGraph.Nodes, node)
			}
			for _, edge := range outbound.edges {
				graph.SymbolGraph.Edges = append(graph.SymbolGraph.Edges, edge)
			}
		}
	}

	// -----------------------------------------------------------------------
	// Collapse to file graph
	// -----------------------------------------------------------------------
	graph.FileGraph = collapseToFileGraph(primaryDef.Location.Path, graph.SymbolGraph.Edges)

	return graph, nil
}

// outboundResult holds intermediate outbound dependency data.
type outboundResult struct {
	nodes []DepGraphNode
	edges []DepGraphEdge
}

// computeOutbound finds refs inside the definition span and resolves them
// to their best candidate definitions, producing outbound symbol edges.
func computeOutbound(
	nav *Navigator,
	def *DefinitionResult,
	defNodeID string,
	lang string,
	repoRoot string,
) (*outboundResult, error) {
	refs, err := nav.RefsInFileRange(
		def.Location.Path,
		def.Location.StartLine,
		def.Location.EndLine,
		lang,
	)
	if err != nil {
		return nil, err
	}

	// Deduplicate ref names.
	seen := map[string]bool{}
	var uniqueNames []string
	for _, r := range refs {
		if r.Name == def.Name {
			continue // skip self-references
		}
		if !seen[r.Name] {
			seen[r.Name] = true
			uniqueNames = append(uniqueNames, r.Name)
		}
	}

	result := &outboundResult{}
	for _, refName := range uniqueNames {
		// Find candidate definitions for this ref name.
		defs, err := nav.GoToDefinitionByName(refName, def.Location.Path, lang)
		if err != nil || len(defs) == 0 {
			continue
		}

		// Pick the best candidate (simple heuristic: same file > same dir > first).
		best := pickBestCandidate(defs, def.Location.Path)
		if best == nil {
			continue
		}

		targetNodeID := nodeID(best.Location.Path, best.Name, best.Location.StartLine)
		// Avoid duplicate nodes.
		result.nodes = append(result.nodes, DepGraphNode{
			ID:        targetNodeID,
			Name:      best.Name,
			Kind:      best.Kind,
			Path:      best.Location.Path,
			StartLine: best.Location.StartLine,
			EndLine:   best.Location.EndLine,
			Container: best.Container,
			Signature: best.Signature,
		})

		// Count how many refs to this name inside the span.
		count := 0
		for _, r := range refs {
			if r.Name == refName {
				count++
			}
		}

		result.edges = append(result.edges, DepGraphEdge{
			From:     defNodeID,
			To:       targetNodeID,
			Kind:     "outbound",
			Score:    1.0 / math.Sqrt(float64(len(defs))), // uniqueness-based score
			Count:    count,
			FilePath: best.Location.Path,
		})
	}
	return result, nil
}

// pickBestCandidate returns the definition most likely referenced from
// contextFile (same file > same dir > first in list).
func pickBestCandidate(defs []DefinitionResult, contextFile string) *DefinitionResult {
	if len(defs) == 0 {
		return nil
	}
	// Prefer same file.
	for i := range defs {
		if defs[i].Location.Path == contextFile {
			return &defs[i]
		}
	}
	// Prefer same directory.
	contextDir := filepath.Dir(contextFile)
	for i := range defs {
		if filepath.Dir(defs[i].Location.Path) == contextDir {
			return &defs[i]
		}
	}
	return &defs[0]
}

// collapseToFileGraph derives file-level edges from symbol-level edges.
func collapseToFileGraph(primaryFile string, edges []DepGraphEdge) []FileGraphEdge {
	type fileEdgeKey struct{ from, to string }
	agg := map[fileEdgeKey]*FileGraphEdge{}

	for _, e := range edges {
		var from, to string
		switch e.Kind {
		case "inbound":
			from = e.FilePath
			to = primaryFile
		case "outbound":
			from = primaryFile
			to = e.FilePath
		default:
			continue
		}
		if from == to {
			continue // skip self-edges
		}
		key := fileEdgeKey{from, to}
		if fe, ok := agg[key]; ok {
			fe.Count += e.Count
			if e.Score > fe.Score {
				fe.Score = e.Score
			}
		} else {
			agg[key] = &FileGraphEdge{
				From:  from,
				To:    to,
				Score: e.Score,
				Count: e.Count,
			}
		}
	}

	result := make([]FileGraphEdge, 0, len(agg))
	for _, fe := range agg {
		fe.Score = round4(fe.Score)
		result = append(result, *fe)
	}

	// Sort for deterministic output.
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score != result[j].Score {
			return result[i].Score > result[j].Score
		}
		if result[i].From != result[j].From {
			return result[i].From < result[j].From
		}
		return result[i].To < result[j].To
	})

	return result
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

// nodeID generates a deterministic ID for a graph node.
func nodeID(path, name string, line int) string {
	return path + ":" + name + ":" + itoa(line)
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	// Reverse.
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

// round4 rounds a float to 4 decimal places.
func round4(f float64) float64 {
	return math.Round(f*10000) / 10000
}

// safeJoinRepoPath safely joins repoRoot and relPath, rejecting path traversal.
func safeJoinRepoPath(repoRoot, relPath string) string {
	cleanRel := filepath.Clean(relPath)
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return ""
	}
	return filepath.Join(repoRoot, cleanRel)
}

// readSingleLine reads a single 1-based line from a file.
func readSingleLine(absPath string, lineNum int) (string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	current := 0
	for scanner.Scan() {
		current++
		if current == lineNum {
			return scanner.Text(), nil
		}
	}
	return "", scanner.Err()
}
