package indexer

import (
	"path/filepath"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Noise-filter options
// ---------------------------------------------------------------------------

// NoiseFilterOptions controls how RankDefinitions and ResolveAndFilterUsages behave.
type NoiseFilterOptions struct {
	// MinConfidence is the minimum confidence score a definition must reach to
	// be included after filtering. Defaults to 0.55 when zero.
	MinConfidence float64

	// IncludeNoise, when true, skips all filtering and returns every candidate
	// as-is (with confidence scores still computed). This lets callers compare
	// filtered vs unfiltered results.
	IncludeNoise bool
}

func (o NoiseFilterOptions) minConf() float64 {
	if o.MinConfidence <= 0 {
		return 0.55
	}
	return o.MinConfidence
}

// ---------------------------------------------------------------------------
// Ranked definition
// ---------------------------------------------------------------------------

// RankedDefinition augments a DefinitionResult with a computed confidence
// score in [0,1] that reflects how likely this candidate is the primary
// definition the user was looking for (as opposed to noise).
type RankedDefinition struct {
	DefinitionResult
	Confidence float64 `json:"confidence"`
}

// ---------------------------------------------------------------------------
// Kind priority table
// ---------------------------------------------------------------------------

// kindPriority returns a base confidence bonus based on symbol kind.
// Type-defining symbols (class, struct, interface, …) get the highest boost;
// value-level variables and local assignments get none.
func kindPriority(kind string) float64 {
	switch kind {
	// Strongly type-defining
	case "class", "interface", "struct", "enum", "trait":
		return 0.40
	// Weakly type-defining
	case "type_alias":
		return 0.30
	// Callable definitions
	case "function", "method", "constructor":
		return 0.20
	// Named constants (Python-style UPPER_CASE)
	case "constant":
		return 0.10
	// Pure noise kinds in cross-name collision scenarios
	case "variable", "field", "property", "parameter":
		return 0.0
	// Package/module – neutral
	case "package", "module":
		return 0.05
	default:
		return 0.05
	}
}

// isTypeKind reports whether kind is a type-level definition (class, struct, …).
func isTypeKind(kind string) bool {
	switch kind {
	case "class", "interface", "struct", "enum", "trait", "type_alias":
		return true
	}
	return false
}

// isValueKind reports whether kind is a pure value-level definition.
func isValueKind(kind string) bool {
	switch kind {
	case "variable", "field", "property", "parameter":
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Noisy-path detection
// ---------------------------------------------------------------------------

// noisySegments is the set of path segments (directory names) treated as
// "noise" when they appear anywhere inside a file path.
// These correspond to migration files, test fixtures, generated code, etc.
var noisySegments = map[string]bool{
	// Database migrations
	"migrations":  true,
	"migration":   true,
	"migrate":     true,
	"db_migrate":  true,
	"db_migration": true,
	"alembic":     true,
	"flyway":      true,
	"liquibase":   true,
	// Test artefacts
	"testdata":    true,
	"test":        true,
	"tests":       true,
	"__tests__":   true,
	"spec":        true,
	"specs":       true,
	// Fixture / mock helpers
	"fixtures":    true,
	"__fixtures__": true,
	"mocks":       true,
	"stubs":       true,
	"fake":        true,
	"fakes":       true,
	// Generated / vendor
	"generated":   true,
	"gen":         true,
	"vendor":      true,
	"pb":          true, // proto-generated
	"proto":       true,
}

// noisySuffixes is a list of file-name suffix patterns that indicate a noise
// file (language-agnostic). Each entry must match path.Base(filePath).
var noisySuffixes = []string{
	// Go test files
	"_test.go",
	// JS / TS test / spec files
	".test.ts", ".test.tsx", ".test.js", ".test.jsx",
	".spec.ts", ".spec.tsx", ".spec.js", ".spec.jsx",
	// Rust integration tests live under tests/
}

// isNoisyPath returns true if the path is considered "noisy" (migrations,
// test fixtures, generated code, etc.).
func isNoisyPath(relPath string) bool {
	// Normalise separators for cross-platform correctness.
	norm := filepath.ToSlash(relPath)
	lower := strings.ToLower(norm)

	// Check each directory segment.
	segments := strings.Split(lower, "/")
	for _, seg := range segments {
		if noisySegments[seg] {
			return true
		}
	}

	// Check file-name suffixes.
	base := filepath.Base(lower)
	for _, suf := range noisySuffixes {
		if strings.HasSuffix(base, suf) {
			return true
		}
	}

	return false
}

// noisePenalty returns a penalty in [0, 0.45] for the path. A higher value
// means more confident it is noise.
func noisePenalty(relPath string) float64 {
	norm := filepath.ToSlash(relPath)
	lower := strings.ToLower(norm)

	// Migration directories carry the highest penalty.
	for _, seg := range []string{"migrations", "migration", "migrate", "db_migrate", "alembic", "flyway", "liquibase"} {
		if containsSegment(lower, seg) {
			return 0.45
		}
	}

	// Fixture / mock directories.
	for _, seg := range []string{"fixtures", "__fixtures__", "mocks", "stubs", "fake", "fakes"} {
		if containsSegment(lower, seg) {
			return 0.40
		}
	}

	// Test directories and generated code.
	for _, seg := range []string{"testdata", "test", "tests", "__tests__", "spec", "specs", "generated", "gen", "vendor", "pb", "proto"} {
		if containsSegment(lower, seg) {
			return 0.35
		}
	}

	// Noisy file suffixes (lighter penalty – could be a test file that also has the real def).
	base := filepath.Base(lower)
	for _, suf := range noisySuffixes {
		if strings.HasSuffix(base, suf) {
			return 0.30
		}
	}

	return 0.0
}

// containsSegment returns true if any path segment equals seg.
func containsSegment(lowerPath, seg string) bool {
	for _, s := range strings.Split(lowerPath, "/") {
		if s == seg {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Locality boost
// ---------------------------------------------------------------------------

// localityBonus returns a bonus when the candidate lives close to filterFile.
func localityBonus(candidatePath, filterFile string) float64 {
	if filterFile == "" {
		return 0
	}
	if candidatePath == filterFile {
		return 0.25
	}
	if filepath.Dir(candidatePath) == filepath.Dir(filterFile) {
		return 0.12
	}
	return 0
}

// ---------------------------------------------------------------------------
// Core: RankDefinitions
// ---------------------------------------------------------------------------

// RankDefinitions scores and (optionally) filters a set of candidate
// definitions for the given name, returning them in descending confidence order.
//
// The algorithm is language-agnostic:
//  1. Assign a base confidence from symbol kind.
//  2. Add a locality bonus if the candidate is near filterFile.
//  3. Subtract a noise penalty for migration / test / generated paths.
//  4. Clamp to [0, 1].
//
// After scoring, if IncludeNoise is false:
//   - Value-kind candidates are dropped when any type-kind candidate exists.
//   - Candidates below MinConfidence are dropped.
//   - If filtering removes everything the top-scored candidate is kept (never
//     empty when input is non-empty).
func RankDefinitions(defs []DefinitionResult, filterFile string, opts NoiseFilterOptions) []RankedDefinition {
	if len(defs) == 0 {
		return nil
	}

	ranked := make([]RankedDefinition, len(defs))
	for i, d := range defs {
		conf := 0.5 // balanced base
		conf += kindPriority(d.Kind)
		conf += localityBonus(d.Location.Path, filterFile)
		conf -= noisePenalty(d.Location.Path)
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}
		ranked[i] = RankedDefinition{
			DefinitionResult: d,
			Confidence:       round4(conf),
		}
	}

	// Sort descending by confidence, then by kind priority as tiebreak.
	sort.SliceStable(ranked, func(i, j int) bool {
		ci, cj := ranked[i].Confidence, ranked[j].Confidence
		if ci != cj {
			return ci > cj
		}
		// Secondary tiebreak: kind priority
		return kindPriority(ranked[i].Kind) > kindPriority(ranked[j].Kind)
	})

	if opts.IncludeNoise {
		return ranked
	}

	// --- Noise filtering ---

	// Rule 1: If any type-kind candidate exists, discard value-kind candidates.
	hasType := false
	for _, r := range ranked {
		if isTypeKind(r.Kind) {
			hasType = true
			break
		}
	}
	if hasType {
		filtered := ranked[:0]
		for _, r := range ranked {
			if !isValueKind(r.Kind) {
				filtered = append(filtered, r)
			}
		}
		ranked = filtered
	}

	// Rule 2: Drop candidates below MinConfidence.
	minConf := opts.minConf()
	filtered := ranked[:0]
	for _, r := range ranked {
		if r.Confidence >= minConf {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) > 0 {
		ranked = filtered
	}
	// If filtering would produce empty results, keep the top-ranked entry
	// (never return empty when input was non-empty).

	return ranked
}

// ---------------------------------------------------------------------------
// Core: ResolveAndFilterUsages
// ---------------------------------------------------------------------------

// ScoredUsageResolved extends ScoredUsage with a per-usage resolution
// confidence (how likely this usage site refers to the given primaryDef).
type ScoredUsageResolved struct {
	ScoredUsage
	ResolutionConfidence float64 `json:"resolutionConfidence"`
}

// ResolveAndFilterUsages scores each usage against candidate definitions using
// the definition-resolution model (ScoreUsages), attaches the resulting
// resolution confidence, and then — unless IncludeNoise is true — filters out
// usages that are likely from a different same-name symbol.
//
// Two complementary filters are applied:
//
//  1. Path-noise filter: if the primary definition lives in a non-noisy path
//     (e.g. models/) and a usage comes from a noisy path (migrations/, tests/
//     fixtures/, generated/, …), that usage is dropped.  This is the primary
//     filter and works even when there is only one candidate definition
//     (ScoreUsages degenerates to 1.0 for all usages in that case).
//
//  2. Resolution-confidence filter: when there are multiple candidate
//     definitions ScoreUsages assigns a meaningful per-usage probability.
//     Usages whose best-matching definition is NOT primaryDef (i.e. they
//     likely reference a different same-name symbol) are dropped when their
//     confidence is below the threshold.
//
// The returned slice retains the original ordering; callers should apply
// GroupAndSortUsages after this call if needed.
func ResolveAndFilterUsages(
	usages []UsageResult,
	candidates []DefinitionResult,
	primaryDef *DefinitionResult,
	repoRoot string,
	opts NoiseFilterOptions,
) []ScoredUsageResolved {
	if len(usages) == 0 {
		return nil
	}

	// Resolution-confidence scoring using the existing depscore machinery.
	resSrc := ScoreUsages(usages, candidates, primaryDef, repoRoot)

	resolved := make([]ScoredUsageResolved, len(resSrc))
	for i, su := range resSrc {
		resolved[i] = ScoredUsageResolved{
			ScoredUsage:          su,
			ResolutionConfidence: su.DependencyScore,
		}
	}

	if opts.IncludeNoise {
		return resolved
	}

	// Determine whether the primary definition is itself in a "clean" path.
	// If it is noisy we cannot reliably filter usage paths by context.
	primaryIsClean := primaryDef != nil && !isNoisyPath(primaryDef.Location.Path)

	// ---------- Filter 1: path-noise filter ----------
	// Drop usages from noisy paths when the primary is non-noisy.
	if primaryIsClean {
		filtered := make([]ScoredUsageResolved, 0, len(resolved))
		for _, r := range resolved {
			if !isNoisyPath(r.Location.Path) {
				filtered = append(filtered, r)
			}
		}
		if len(filtered) > 0 {
			resolved = filtered
		}
		// Never-empty safeguard: if path filter removed everything keep all.
	}

	// ---------- Filter 2: resolution-confidence filter ----------
	// Only meaningful when there are at least 2 candidates; with a single
	// candidate ScoreUsages always normalises to 1.0 so the threshold is
	// trivially satisfied, and we rely solely on the path filter above.
	if primaryDef != nil && len(candidates) >= 2 {
		threshold := opts.minConf()
		filtered := make([]ScoredUsageResolved, 0, len(resolved))
		for _, r := range resolved {
			if r.ResolutionConfidence < threshold {
				continue
			}
			// BestDefinition must point at the primary (same file+line+col).
			if r.BestDefinition != nil &&
				(r.BestDefinition.Location.Path != primaryDef.Location.Path ||
					r.BestDefinition.Location.StartLine != primaryDef.Location.StartLine ||
					r.BestDefinition.Location.StartCol != primaryDef.Location.StartCol) {
				continue
			}
			filtered = append(filtered, r)
		}
		if len(filtered) > 0 {
			resolved = filtered
		}
	}

	return resolved
}
