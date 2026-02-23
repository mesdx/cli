package indexer

import "strings"

// ---------------------------------------------------------------------------
// Coupling-strength scoring
// ---------------------------------------------------------------------------

// couplingWeights maps ref kinds to their base coupling strength [0,1].
// These weights reflect how tightly a usage couples the caller to the
// definition — independent of how many definitions or usages exist.
var couplingWeights = map[string]float64{
	"inherit":    1.00, // structural dependency: inheritance / trait impl
	"call":       0.65, // direct function or method call
	"write":      0.55, // assignment of a value of this type
	"read":       0.50, // reading a field or variable of this type
	"import":     0.40, // importing the package/module that provides this symbol
	"type_ref":   0.25, // type annotation, parameter type, return type
	"annotation": 0.20, // decorator / attribute annotation
	"other":      0.10, // generic identifier reference (catch-all)
}

// CouplingScore returns the coupling strength [0,1] for a single usage.
//
// The score reflects how hard it would be to change the referenced symbol
// without touching this usage site:
//   - 1.0 = structural dependency (inheritance, instantiation)
//   - 0.6 = direct call coupling
//   - 0.1 = casual mention (instanceof check, comment-like identifier)
//
// Scoring is NOT normalized by usage count or candidate count, so it remains
// meaningful even when the symbol has thousands of usages.
func CouplingScore(usage UsageResult, srcLine string) float64 {
	base := couplingWeights[usage.Kind]
	if base == 0 {
		base = 0.10 // default for unrecognised kinds
	}

	// Structural-relation override (set by tree-sitter inherit/implements captures).
	switch usage.Relation {
	case "inherits", "implements":
		return 1.0
	case "prototype":
		if base < 0.80 {
			base = 0.80
		}
	}

	// Lexical escalation: detect instantiation / inheritance patterns that
	// the structured ref kind may not have captured (especially in JS/TS/Go
	// where these surface as generic "identifier" or "type_ref" refs).
	if base < 0.99 && srcLine != "" {
		name := usage.Name
		base = lexicalEscalate(base, name, srcLine)
	}

	if base > 1.0 {
		return 1.0
	}
	return base
}

// lexicalEscalate raises the coupling score when the source line pattern
// indicates a stronger coupling than the ref kind alone suggests.
func lexicalEscalate(base float64, name, srcLine string) float64 {
	// ── Inheritance / interface implementation ────────────────────────────
	// Matches: "extends Name", "implements Name" (Java, TS, JS)
	if strings.Contains(srcLine, "extends "+name) ||
		strings.Contains(srcLine, "implements "+name) {
		return 1.0
	}

	// ── Instantiation ─────────────────────────────────────────────────────
	const instScore = 0.95

	// "new Name(" or "new Name{" — Java, JS, TS, C#
	if strings.Contains(srcLine, "new "+name+"(") ||
		strings.Contains(srcLine, "new "+name+"{") ||
		strings.Contains(srcLine, "new "+name+" {") {
		if base < instScore {
			return instScore
		}
		return base
	}

	// "&Name{" — Go pointer struct literal
	if strings.Contains(srcLine, "&"+name+"{") ||
		strings.Contains(srcLine, "&"+name+" {") {
		if base < instScore {
			return instScore
		}
		return base
	}

	// "Name{" or "Name {" — Go / Rust struct literal
	if strings.Contains(srcLine, name+"{") ||
		strings.Contains(srcLine, name+" {") {
		if base < instScore {
			return instScore
		}
		return base
	}

	// "Name::new" — Rust associated constructor
	if strings.Contains(srcLine, name+"::new") ||
		strings.Contains(srcLine, name+"::New") {
		if base < instScore {
			return instScore
		}
		return base
	}

	return base
}

// CoupleUsages computes per-usage coupling strengths for a slice of usages.
//
// Unlike ScoreUsages (which computes definition-resolution confidence),
// CoupleUsages does NOT depend on the number of usages or candidate
// definitions. Every score in [0,1] reflects only the coupling nature of
// that specific usage site, making scores comparable across symbols with
// very different usage counts.
func CoupleUsages(usages []UsageResult, repoRoot string) []ScoredUsage {
	lineCache := map[string]string{}
	result := make([]ScoredUsage, len(usages))
	for i, u := range usages {
		srcLine := getSourceLine(repoRoot, u.Location.Path, u.Location.StartLine, lineCache)
		result[i] = ScoredUsage{
			UsageResult:     u,
			DependencyScore: round4(CouplingScore(u, srcLine)),
		}
	}
	return result
}
