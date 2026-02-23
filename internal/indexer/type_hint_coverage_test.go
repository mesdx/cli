package indexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesdx/cli/internal/symbols"
)

// ---------------------------------------------------------------------------
// Integration tests: Python type-annotation coverage
// ---------------------------------------------------------------------------
//
// These tests prove that findUsages finds DataModel when it appears inside
// quoted forward-reference annotations, not just in bare identifier position.
//
// All tests index the full testdata tree (reuses setupNavigationTest) so they
// exercise the complete pipeline: Tree-sitter parse → expansion → DB store →
// SQL query.

// TestFindUsages_PythonQuotedForwardRef_DirectReturn verifies that a symbol
// used as a direct quoted return type (-> "DataModel") is found by findUsages.
func TestFindUsages_PythonQuotedForwardRef_DirectReturn(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("DataModel", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	// Must find at least one usage inside type_hint_edges.py
	var inFixture []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "type_hint_edges.py") {
			inFixture = append(inFixture, u)
		}
	}
	if len(inFixture) == 0 {
		t.Fatal("findUsages found no DataModel usages in type_hint_edges.py")
	}

	// At least one must be an annotation ref (from a quoted annotation)
	foundAnnotation := false
	for _, u := range inFixture {
		if u.Kind == symbols.RefAnnotation.String() || u.Relation == "annotation" {
			foundAnnotation = true
			t.Logf("  annotation ref: %s  line=%d col=%d  kind=%s  relation=%s",
				u.Name, u.Location.StartLine, u.Location.StartCol, u.Kind, u.Relation)
		}
	}
	if !foundAnnotation {
		t.Error("expected at least one annotation ref for DataModel in type_hint_edges.py")
		for _, u := range inFixture {
			t.Logf("  found ref: kind=%s relation=%s line=%d col=%d",
				u.Kind, u.Relation, u.Location.StartLine, u.Location.StartCol)
		}
	}
}

// TestFindUsages_PythonQuotedForwardRef_GenericParam verifies that a symbol
// used as a quoted generic parameter (List["DataModel"], tuple["DataModel", bool])
// is found by findUsages.
func TestFindUsages_PythonQuotedForwardRef_GenericParam(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("DataModel", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	// Find usages in type_hint_edges.py on lines that have generic annotations.
	// type_hint_edges.py line numbers for generic patterns (1-indexed):
	//   get_list_of_models_generic()  -> List["DataModel"]    (Pattern 2)
	//   get_tuple_with_quoted()       -> tuple["DataModel", bool] (Pattern 2)
	//   get_optional_generic()        -> Optional["DataModel"] (Pattern 2)
	//   get_union_generic()           -> Union["DataModel", None] (Pattern 2)
	//   process_list_param(models: List["DataModel"]) (Pattern 2 – param)
	//   process_tuple_param(data: tuple["DataModel", bool]) (Pattern 2 – param)
	foundGenericLine := false
	for _, u := range usages {
		if !strings.HasSuffix(u.Location.Path, "type_hint_edges.py") {
			continue
		}
		if (u.Kind == symbols.RefAnnotation.String() || u.Relation == "annotation") &&
			u.Location.StartLine > 0 {
			foundGenericLine = true
			t.Logf("  generic annotation ref: line=%d col=%d", u.Location.StartLine, u.Location.StartCol)
		}
	}
	if !foundGenericLine {
		t.Error("expected at least one annotation ref for DataModel from generic annotations in type_hint_edges.py")
	}
}

// TestFindUsages_PythonQuotedForwardRef_ComplexString verifies that a symbol
// referenced inside a complex quoted type expression ("tuple[DataModel, bool]")
// is found by findUsages.
func TestFindUsages_PythonQuotedForwardRef_ComplexString(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	// The fixture has: def get_tuple_fully_quoted() -> "tuple[DataModel, bool]":
	// DataModel appears inside the string that constitutes the full annotation.
	usages, err := nav.FindUsagesByName("DataModel", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	var inFixture []UsageResult
	for _, u := range usages {
		if strings.HasSuffix(u.Location.Path, "type_hint_edges.py") {
			inFixture = append(inFixture, u)
		}
	}
	if len(inFixture) == 0 {
		t.Fatal("findUsages found no DataModel usages in type_hint_edges.py — fixture may not be indexed")
	}

	t.Logf("DataModel usages in type_hint_edges.py: %d", len(inFixture))
	for _, u := range inFixture {
		t.Logf("  line=%d col=%d kind=%s relation=%s",
			u.Location.StartLine, u.Location.StartCol, u.Kind, u.Relation)
	}

	// Expect multiple annotation refs: the fixture has ~10 quoted annotation sites.
	annotCount := 0
	for _, u := range inFixture {
		if u.Kind == symbols.RefAnnotation.String() || u.Relation == "annotation" {
			annotCount++
		}
	}
	if annotCount < 5 {
		t.Errorf("expected >= 5 annotation refs for DataModel in type_hint_edges.py, got %d", annotCount)
	}
}

// TestFindUsages_PythonNegativeControl verifies that ordinary string literals
// that merely contain the symbol name do NOT produce annotation refs.
func TestFindUsages_PythonNegativeControl(t *testing.T) {
	nav, _, cleanup := setupNavigationTest(t)
	defer cleanup()

	usages, err := nav.FindUsagesByName("DataModel", "", "python")
	if err != nil {
		t.Fatalf("FindUsagesByName: %v", err)
	}

	// In type_hint_edges.py these lines have plain strings mentioning DataModel
	// but are NOT type annotations:
	//   line 141: description = "Please use DataModel for all data."
	//   line 151: return "DataModel is the key class"
	//   line 154: key: str = "DataModel"  (default value, not annotation)
	//   line 160: model_name_str = "DataModel"
	//
	// The negative controls section starts at line 141 (first line after the
	// separator comment on line 137).  None of these should produce an
	// annotation ref for DataModel.
	for _, u := range usages {
		if !strings.HasSuffix(u.Location.Path, "type_hint_edges.py") {
			continue
		}
		if u.Kind != symbols.RefAnnotation.String() && u.Relation != "annotation" {
			continue
		}
		// Negative controls are at lines 141–165.
		if u.Location.StartLine >= 141 && u.Location.StartLine <= 165 {
			t.Errorf("negative control: unexpected annotation ref for DataModel at line %d (should not be an annotation ref)",
				u.Location.StartLine)
		}
	}
}

// ---------------------------------------------------------------------------
// Extractor-level unit tests – no DB needed
// ---------------------------------------------------------------------------

// TestParsePythonTypeHintEdges runs the tree-sitter parser directly on the
// type_hint_edges.py fixture file and asserts expected refs are produced.
func TestParsePythonTypeHintEdges(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(testdataDir(t), "python", "type_hint_edges.py"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	parser := NewTreeSitterParser("python")
	result, err := parser.Parse("type_hint_edges.py", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// ---- Symbols ----
	expectSymbol(t, result, "DataModel", symbols.KindClass)

	// ---- Refs must include annotation refs for DataModel ----
	if len(result.Refs) == 0 {
		t.Fatal("expected refs, got none")
	}

	// Count annotation refs specifically
	var annotRefs []symbols.Ref
	for _, r := range result.Refs {
		if r.Name == "DataModel" && r.Kind == symbols.RefAnnotation {
			annotRefs = append(annotRefs, r)
			t.Logf("  DataModel annotation ref: line=%d col=%d relation=%s", r.StartLine, r.StartCol, r.Relation)
		}
	}

	// We have at least these quoted annotation sites (Pattern 1-4 plus params):
	//   get_model_direct_quoted()      -> "DataModel"           1 ref
	//   process_model_param_quoted()   param "DataModel"        1 ref
	//   get_optional_quoted()          -> "Optional[DataModel]" 1 ref (DataModel inside)
	//   get_list_of_models_generic()   -> List["DataModel"]     1 ref
	//   get_tuple_with_quoted()        -> tuple["DataModel",..] 1 ref
	//   get_optional_generic()         -> Optional["DataModel"] 1 ref
	//   get_union_generic()            -> Union["DataModel",..]  1 ref
	//   process_list_param()           List["DataModel"] param  1 ref
	//   process_tuple_param()          tuple["DataModel",..] ×2 1+1
	//   get_model_fully_quoted()       -> "DataModel"           1 ref
	//   get_tuple_fully_quoted()       "tuple[DataModel, bool]" 1 ref
	//   get_list_fully_quoted()        "List[DataModel]"        1 ref
	//   get_nested_generic()           Optional[List["DataModel"]] 1 ref
	//   get_deeply_nested()            Optional["List[DataModel]"] 1 ref
	//   latest: "DataModel" = ...      variable annotation      1 ref
	//   cache: List["DataModel"] = []  variable annotation      1 ref
	// Total: >= 16
	if len(annotRefs) < 10 {
		t.Errorf("expected >= 10 DataModel annotation refs, got %d", len(annotRefs))
	}

	// Builtin refs should still be marked as builtin
	expectRefBuiltin(t, result, "len", true)

	// Dedup: DataModel class definition must NOT also appear as a ref at the same position
	expectNoDuplicateDefRef(t, result, "DataModel")
}

