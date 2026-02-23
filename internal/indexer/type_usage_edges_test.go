package indexer

import (
	"strings"
	"testing"
)

// TestFindUsages_TypeUsageEdges_AllLangs verifies that findUsages can locate
// CoreModel when it appears inside generic and container type expressions in
// all supported statically-typed languages, and as runtime usage patterns in
// JavaScript.
//
// Each sub-test asserts that FindUsagesByName("CoreModel", "", lang) returns
// at least one usage whose path ends with the corresponding type_usage_edges
// fixture file.
func TestFindUsages_TypeUsageEdges_AllLangs(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	cases := []struct {
		lang     string
		wantFile string // suffix of expected fixture path
	}{
		{"go", "go/type_usage_edges.go"},
		{"java", "type_usage_edges/TypeUsageEdges.java"},
		{"rust", "rust/type_usage_edges.rs"},
		{"typescript", "ts/type_usage_edges.ts"},
		{"javascript", "js/type_usage_edges.js"},
	}

	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			usages, err := nav.FindUsagesByName("CoreModel", "", tc.lang)
			if err != nil {
				t.Fatalf("FindUsagesByName(%q): %v", tc.lang, err)
			}
			if len(usages) == 0 {
				t.Fatalf("findUsages returned no CoreModel usages for %s", tc.lang)
			}

			// Check that at least one usage comes from the edge fixture.
			foundInEdge := false
			for _, u := range usages {
				if strings.HasSuffix(u.Location.Path, tc.wantFile) {
					foundInEdge = true
					t.Logf("  found: %s  kind=%s  line=%d col=%d",
						u.Location.Path, u.Kind, u.Location.StartLine, u.Location.StartCol)
					break
				}
			}
			if !foundInEdge {
				t.Errorf("expected at least one CoreModel usage in %s; files found:", tc.wantFile)
				paths := map[string]bool{}
				for _, u := range usages {
					paths[u.Location.Path] = true
				}
				for p := range paths {
					t.Logf("  %s", p)
				}
			}
		})
	}
}

// TestFindUsages_TypeUsageEdges_Go_TypePositions verifies that CoreModel refs
// from the Go fixture are found specifically in type positions (slices, maps,
// pointers, channel, local variables).
func TestFindUsages_TypeUsageEdges_Go_TypePositions(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("CoreModel", "", "go")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inEdge []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "go/type_usage_edges.go") {
			inEdge = append(inEdge, u)
		}
	}
	if len(inEdge) == 0 {
		t.Fatal("no CoreModel usages found in go/type_usage_edges.go")
	}
	// The fixture has CoreModel in at least 10 type positions.
	if len(inEdge) < 10 {
		t.Errorf("expected >= 10 CoreModel usages in type_usage_edges.go, got %d", len(inEdge))
	}
	t.Logf("Go type_usage_edges.go: %d CoreModel usages", len(inEdge))
}

// TestFindUsages_TypeUsageEdges_Java_Generics verifies that CoreModel refs
// from the Java fixture are found when CoreModel appears inside generic type
// parameters (List<CoreModel>, Map<String, CoreModel>, Optional<CoreModel>).
func TestFindUsages_TypeUsageEdges_Java_Generics(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("CoreModel", "", "java")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inEdge []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "TypeUsageEdges.java") {
			inEdge = append(inEdge, u)
		}
	}
	if len(inEdge) == 0 {
		t.Fatal("no CoreModel usages found in TypeUsageEdges.java")
	}
	// The fixture has CoreModel in at least 8 generic type positions.
	if len(inEdge) < 8 {
		t.Errorf("expected >= 8 CoreModel usages in TypeUsageEdges.java, got %d", len(inEdge))
	}
	t.Logf("Java TypeUsageEdges.java: %d CoreModel usages", len(inEdge))
}

// TestFindUsages_TypeUsageEdges_Rust_Generics verifies that CoreModel refs from
// the Rust fixture are found when CoreModel appears inside Vec<>, Option<>,
// Result<>, Box<>, and tuple type positions.
func TestFindUsages_TypeUsageEdges_Rust_Generics(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("CoreModel", "", "rust")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inEdge []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "rust/type_usage_edges.rs") {
			inEdge = append(inEdge, u)
		}
	}
	if len(inEdge) == 0 {
		t.Fatal("no CoreModel usages found in rust/type_usage_edges.rs")
	}
	// The fixture has CoreModel in at least 8 generic type positions.
	if len(inEdge) < 8 {
		t.Errorf("expected >= 8 CoreModel usages in type_usage_edges.rs, got %d", len(inEdge))
	}
	t.Logf("Rust type_usage_edges.rs: %d CoreModel usages", len(inEdge))
}

// TestFindUsages_TypeUsageEdges_TypeScript_Generics verifies that CoreModel is
// found when it appears in Array<>, Promise<>, ReadonlyArray<>, Record<>,
// Map<>, Set<>, union types, and interface / type-alias positions.
func TestFindUsages_TypeUsageEdges_TypeScript_Generics(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("CoreModel", "", "typescript")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inEdge []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "ts/type_usage_edges.ts") {
			inEdge = append(inEdge, u)
		}
	}
	if len(inEdge) == 0 {
		t.Fatal("no CoreModel usages found in ts/type_usage_edges.ts")
	}
	// The fixture has CoreModel in at least 10 type positions.
	if len(inEdge) < 10 {
		t.Errorf("expected >= 10 CoreModel usages in type_usage_edges.ts, got %d", len(inEdge))
	}
	t.Logf("TypeScript type_usage_edges.ts: %d CoreModel usages", len(inEdge))
}

// TestFindUsages_TypeUsageEdges_JavaScript_Runtime verifies that CoreModel is
// found in the JavaScript fixture through runtime usage patterns (instanceof,
// new, function references).
func TestFindUsages_TypeUsageEdges_JavaScript_Runtime(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("CoreModel", "", "javascript")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inEdge []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "js/type_usage_edges.js") {
			inEdge = append(inEdge, u)
		}
	}
	if len(inEdge) == 0 {
		t.Fatal("no CoreModel usages found in js/type_usage_edges.js")
	}
	t.Logf("JavaScript type_usage_edges.js: %d CoreModel usages", len(inEdge))
}
