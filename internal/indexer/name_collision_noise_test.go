package indexer

// TestNameCollisionNoise* — exercises the noise-filter and definition-ranking
// helpers against a set of per-language repos that are generated entirely in
// memory (os.MkdirTemp).  This keeps the testdata directory clean while still
// providing thorough multi-language coverage.
//
// For each supported language the test:
//   1. Writes a canonical model file under models/ that defines UserModel as a
//      type (class/struct/etc.).
//   2. Writes N migration files under migrations/ that each create a *local
//      variable* also named UserModel.
//   3. Writes M fixture files under tests/fixtures/ that reference UserModel
//      in helper assignments.
//   4. Writes a realistic consumer file under services/ that imports and uses
//      the real UserModel type (call, type-ref, instantiation).
//
// After indexing the assertions verify:
//   • RankDefinitions returns the models/ type as the top candidate.
//   • The top candidate's confidence exceeds the noisy migration candidates'.
//   • ResolveAndFilterUsages excludes migration/fixture paths from usages
//     returned under default (IncludeNoise=false) filtering.
//   • When IncludeNoise=true, migration/fixture usages are present.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mesdx/cli/internal/db"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeTempFile(t *testing.T, dir, relPath, contents string) {
	t.Helper()
	abs := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", filepath.Dir(abs), err)
	}
	if err := os.WriteFile(abs, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", abs, err)
	}
}

func setupTempRepo(t *testing.T, root string) (*Navigator, string, func()) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "noise_test.db")
	if err := db.Initialize(dbPath); err != nil {
		t.Fatalf("db.Initialize: %v", err)
	}
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	idx := New(d, root)
	if _, err := idx.FullIndex([]string{"."}); err != nil {
		t.Fatalf("FullIndex: %v", err)
	}
	nav := &Navigator{DB: d, ProjectID: idx.Store.ProjectID, RepoRoot: root}
	return nav, root, func() { _ = d.Close() }
}

const numMigrations = 20
const numFixtures = 10

// ---------------------------------------------------------------------------
// Per-language fixture generators
// ---------------------------------------------------------------------------

func writeGoCollisionFixtures(t *testing.T, root string) {
	t.Helper()

	// Primary model type
	writeTempFile(t, root, "models/user_model.go", `package models

type UserModel struct {
	ID    int
	Email string
	Name  string
}

func NewUserModel(id int, email, name string) *UserModel {
	return &UserModel{ID: id, Email: email, Name: name}
}
`)

	// Service consumer (realistic usage of the real UserModel type)
	writeTempFile(t, root, "services/user_service.go", `package services

import "models"

func GetUser(id int) *models.UserModel {
	return models.NewUserModel(id, "user@example.com", "Alice")
}

func SaveUser(u *models.UserModel) error {
	return nil
}
`)

	// Migration files — each declares a local variable named UserModel
	for i := 0; i < numMigrations; i++ {
		writeTempFile(t, root, fmt.Sprintf("migrations/m%04d_user.go", i), fmt.Sprintf(`package migrations

func Migration%04d(db interface{}) {
	UserModel := struct{ id int }{id: %d}
	_ = UserModel
}
`, i, i))
	}

	// Test fixture files — helper variables / shadow definitions
	for i := 0; i < numFixtures; i++ {
		writeTempFile(t, root, fmt.Sprintf("tests/fixtures/user_fixture_%d.go", i), fmt.Sprintf(`package fixtures

func UserModelFixture%d() interface{} {
	UserModel := map[string]interface{}{"id": %d, "email": "test@test.com"}
	return UserModel
}
`, i, i))
	}
}

func writeJavaCollisionFixtures(t *testing.T, root string) {
	t.Helper()

	writeTempFile(t, root, "models/UserModel.java", `package models;

public class UserModel {
    private int id;
    private String email;
    private String name;

    public UserModel(int id, String email, String name) {
        this.id = id;
        this.email = email;
        this.name = name;
    }
    public int getId() { return id; }
    public String getEmail() { return email; }
    public String getName() { return name; }
}
`)

	writeTempFile(t, root, "services/UserService.java", `package services;

import models.UserModel;

public class UserService {
    public UserModel getUser(int id) {
        return new UserModel(id, "user@example.com", "Alice");
    }
    public void saveUser(UserModel u) {
        // persist
    }
    public UserModel createUser(String email, String name) {
        return new UserModel(0, email, name);
    }
}
`)

	for i := 0; i < numMigrations; i++ {
		migrationBody := fmt.Sprintf(`package migrations;

public class Migration%04d {
    public void run(Object db) {
        Object UserModel = new Object();
        System.out.println(UserModel);
    }
}
`, i)
		writeTempFile(t, root, fmt.Sprintf("migrations/Migration%04d.java", i), migrationBody)
	}

	for i := 0; i < numFixtures; i++ {
		writeTempFile(t, root, fmt.Sprintf("tests/fixtures/UserFixture%d.java", i), fmt.Sprintf(`package fixtures;

public class UserFixture%d {
    public static Object build() {
        Object UserModel = new Object();
        return UserModel;
    }
}
`, i))
	}
}

func writeRustCollisionFixtures(t *testing.T, root string) {
	t.Helper()

	writeTempFile(t, root, "models/user_model.rs", `pub struct UserModel {
    pub id: i64,
    pub email: String,
    pub name: String,
}

impl UserModel {
    pub fn new(id: i64, email: &str, name: &str) -> Self {
        UserModel { id, email: email.to_string(), name: name.to_string() }
    }
}
`)

	writeTempFile(t, root, "services/user_service.rs", `use crate::models::user_model::UserModel;

pub fn get_user(id: i64) -> UserModel {
    UserModel::new(id, "user@example.com", "Alice")
}

pub fn save_user(u: &UserModel) -> bool {
    let _id = u.id;
    true
}

pub fn create_user(email: &str, name: &str) -> UserModel {
    UserModel::new(0, email, name)
}
`)

	for i := 0; i < numMigrations; i++ {
		writeTempFile(t, root, fmt.Sprintf("migrations/m%04d_user.rs", i), fmt.Sprintf(`pub fn migration_%04d(db: &dyn std::any::Any) {
    let UserModel = format!("migration_row_%d");
    let _ = UserModel;
}
`, i, i))
	}

	for i := 0; i < numFixtures; i++ {
		writeTempFile(t, root, fmt.Sprintf("tests/fixtures/user_fixture_%d.rs", i), fmt.Sprintf(`pub fn user_model_fixture_%d() -> String {
    let UserModel = format!("fixture_%d@test.com");
    UserModel
}
`, i, i))
	}
}

func writeTypeScriptCollisionFixtures(t *testing.T, root string) {
	t.Helper()

	writeTempFile(t, root, "models/UserModel.ts", `export class UserModel {
    constructor(
        public id: number,
        public email: string,
        public name: string,
    ) {}

    static fromRow(row: Record<string, unknown>): UserModel {
        return new UserModel(Number(row.id), String(row.email), String(row.name));
    }
}
`)

	writeTempFile(t, root, "services/userService.ts", `import { UserModel } from '../models/UserModel';

export function getUser(id: number): UserModel {
    return new UserModel(id, 'user@example.com', 'Alice');
}

export function saveUser(u: UserModel): boolean {
    return u.id > 0;
}

export function createUser(email: string, name: string): UserModel {
    return new UserModel(0, email, name);
}
`)

	for i := 0; i < numMigrations; i++ {
		writeTempFile(t, root, fmt.Sprintf("migrations/m%04d_user.ts", i), fmt.Sprintf(`export async function migration%04d(db: unknown): Promise<void> {
    const UserModel = { id: %d, email: 'mig@db.com' };
    console.log(UserModel);
}
`, i, i))
	}

	for i := 0; i < numFixtures; i++ {
		writeTempFile(t, root, fmt.Sprintf("tests/fixtures/userFixture%d.ts", i), fmt.Sprintf(`export function buildUserModel%d(): Record<string, unknown> {
    const UserModel = { id: %d, email: 'fix@test.com', name: 'Test' };
    return UserModel;
}
`, i, i))
	}
}

func writeJavaScriptCollisionFixtures(t *testing.T, root string) {
	t.Helper()

	writeTempFile(t, root, "models/UserModel.js", `class UserModel {
    constructor(id, email, name) {
        this.id = id;
        this.email = email;
        this.name = name;
    }

    static fromRow(row) {
        return new UserModel(row.id, row.email, row.name);
    }
}

module.exports = { UserModel };
`)

	writeTempFile(t, root, "services/userService.js", `const { UserModel } = require('../models/UserModel');

function getUser(id) {
    return new UserModel(id, 'user@example.com', 'Alice');
}

function saveUser(u) {
    return u instanceof UserModel;
}

function createUser(email, name) {
    return new UserModel(0, email, name);
}

module.exports = { getUser, saveUser, createUser };
`)

	for i := 0; i < numMigrations; i++ {
		writeTempFile(t, root, fmt.Sprintf("migrations/m%04d_user.js", i), fmt.Sprintf(`async function migration%04d(db) {
    const UserModel = { id: %d, email: 'mig@db.com' };
    return UserModel;
}

module.exports = { migration%04d };
`, i, i, i))
	}

	for i := 0; i < numFixtures; i++ {
		writeTempFile(t, root, fmt.Sprintf("tests/fixtures/userFixture%d.js", i), fmt.Sprintf(`function buildUserModel%d() {
    const UserModel = { id: %d, email: 'fix@test.com', name: 'Test' };
    return UserModel;
}

module.exports = { buildUserModel%d };
`, i, i, i))
	}
}

// ---------------------------------------------------------------------------
// Shared assertion helpers
// ---------------------------------------------------------------------------

// assertPrimaryAtTop verifies that the top-ranked definition is from the
// models/ path (not a migration or fixture).
func assertPrimaryAtTop(t *testing.T, lang string, ranked []RankedDefinition) {
	t.Helper()
	if len(ranked) == 0 {
		t.Errorf("[%s] RankDefinitions returned 0 results", lang)
		return
	}
	top := ranked[0]
	norm := filepath.ToSlash(top.Location.Path)
	if !strings.Contains(norm, "models/") {
		t.Errorf("[%s] top-ranked definition is NOT from models/: got %q (kind=%s conf=%.2f)",
			lang, top.Location.Path, top.Kind, top.Confidence)
	}
}

// assertModelTypeKind verifies that the top-ranked result is a type-defining
// symbol (class, struct, type_alias, …) rather than a variable.
func assertModelTypeKind(t *testing.T, lang string, ranked []RankedDefinition) {
	t.Helper()
	if len(ranked) == 0 {
		return
	}
	top := ranked[0]
	if !isTypeKind(top.Kind) {
		t.Errorf("[%s] top-ranked definition has non-type kind %q (expected class/struct/…)", lang, top.Kind)
	}
}

// assertHigherConfidenceThanNoise verifies that the top-ranked models/
// candidate has a higher confidence than every migration candidate.
func assertHigherConfidenceThanNoise(t *testing.T, lang string, ranked []RankedDefinition) {
	t.Helper()
	if len(ranked) == 0 {
		return
	}
	topConf := ranked[0].Confidence
	for _, r := range ranked[1:] {
		norm := filepath.ToSlash(r.Location.Path)
		if strings.Contains(norm, "migrations/") && r.Confidence >= topConf {
			t.Errorf("[%s] migration candidate %q has confidence %.2f >= top %.2f",
				lang, r.Location.Path, r.Confidence, topConf)
		}
	}
}

// assertNoMigrationPathsFiltered asserts that none of the usages in the
// filtered result come from a migrations/ directory.
func assertNoMigrationPathsFiltered(t *testing.T, lang string, usages []ScoredUsageResolved) {
	t.Helper()
	for _, u := range usages {
		norm := filepath.ToSlash(u.Location.Path)
		if strings.Contains(norm, "migrations/") {
			t.Errorf("[%s] filtered usages still contain migration path %q", lang, u.Location.Path)
		}
	}
}

// assertNoFixturePathsFiltered asserts that none of the usages in the
// filtered result come from a tests/fixtures/ directory.
func assertNoFixturePathsFiltered(t *testing.T, lang string, usages []ScoredUsageResolved) {
	t.Helper()
	for _, u := range usages {
		norm := filepath.ToSlash(u.Location.Path)
		if strings.Contains(norm, "fixtures/") {
			t.Errorf("[%s] filtered usages still contain fixture path %q", lang, u.Location.Path)
		}
	}
}

// assertNoisyPathsPresentWhenIncluded verifies that when IncludeNoise=true,
// migration or fixture paths DO appear in the result.
func assertNoisyPathsPresentWhenIncluded(t *testing.T, lang string, usages []ScoredUsageResolved) {
	t.Helper()
	for _, u := range usages {
		norm := filepath.ToSlash(u.Location.Path)
		if strings.Contains(norm, "migrations/") || strings.Contains(norm, "fixtures/") {
			return // found at least one noisy path — good
		}
	}
	t.Errorf("[%s] expected at least one migration/fixture usage when IncludeNoise=true, got none", lang)
}

// assertMoreResultsWithNoise asserts that unfiltered results have more items
// than filtered results.
func assertMoreResultsWithNoise(t *testing.T, lang string, filtered, unfiltered []ScoredUsageResolved) {
	t.Helper()
	if len(unfiltered) <= len(filtered) {
		t.Errorf("[%s] includeNoise=true should produce more usages: filtered=%d unfiltered=%d",
			lang, len(filtered), len(unfiltered))
	}
}

// ---------------------------------------------------------------------------
// Core per-language test runner
// ---------------------------------------------------------------------------

type langFixtureSetup struct {
	lang string
	ext  string
	// writeFixtures writes the files for this language into root.
	writeFixtures func(t *testing.T, root string)
}

func runNameCollisionTest(t *testing.T, tc langFixtureSetup) {
	t.Helper()
	root, err := os.MkdirTemp("", "noise-test-"+tc.lang+"-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(root) })

	tc.writeFixtures(t, root)

	nav, repoRoot, cleanup := setupTempRepo(t, root)
	defer cleanup()

	// -----------------------------------------------------------------------
	// Phase 1: RankDefinitions
	// -----------------------------------------------------------------------
	rawDefs, err := nav.GoToDefinitionByName("UserModel", "", tc.lang)
	if err != nil {
		t.Fatalf("[%s] GoToDefinitionByName: %v", tc.lang, err)
	}
	if len(rawDefs) == 0 {
		t.Fatalf("[%s] no definitions found for UserModel — fixture indexing failed", tc.lang)
	}

	// With filtering (default)
	filteredDefs := RankDefinitions(rawDefs, "", NoiseFilterOptions{})
	assertPrimaryAtTop(t, tc.lang, filteredDefs)
	assertModelTypeKind(t, tc.lang, filteredDefs)

	// With IncludeNoise=true (all candidates, still ranked)
	allDefs := RankDefinitions(rawDefs, "", NoiseFilterOptions{IncludeNoise: true})
	if len(allDefs) < len(filteredDefs) {
		t.Errorf("[%s] includeNoise=true returned fewer candidates than filtered (%d < %d)",
			tc.lang, len(allDefs), len(filteredDefs))
	}

	// The models/ type should still be at or near the top of allDefs.
	assertPrimaryAtTop(t, tc.lang, allDefs)
	assertHigherConfidenceThanNoise(t, tc.lang, allDefs)

	// All migration variable definitions should have lower confidence than the
	// model type.
	if len(filteredDefs) > 0 {
		topConf := filteredDefs[0].Confidence
		if topConf < 0.55 {
			t.Errorf("[%s] top definition confidence %.2f is below threshold 0.55", tc.lang, topConf)
		}
	}

	// There should be fewer definitions after filtering than in the raw set.
	if len(rawDefs) > 1 && len(filteredDefs) >= len(rawDefs) {
		t.Errorf("[%s] filtering removed no candidates: raw=%d filtered=%d",
			tc.lang, len(rawDefs), len(filteredDefs))
	}

	// -----------------------------------------------------------------------
	// Phase 2: ResolveAndFilterUsages
	// -----------------------------------------------------------------------
	rawUsages, err := nav.FindUsagesByName("UserModel", "", tc.lang)
	if err != nil {
		t.Fatalf("[%s] FindUsagesByName: %v", tc.lang, err)
	}
	if len(rawUsages) == 0 {
		t.Fatalf("[%s] no usages found for UserModel — fixture indexing failed", tc.lang)
	}

	// Compute candidate defs from ranked (filtered) set.
	candidateDefs := make([]DefinitionResult, len(filteredDefs))
	for i, rd := range filteredDefs {
		candidateDefs[i] = rd.DefinitionResult
	}
	if len(candidateDefs) == 0 {
		// Fallback to raw if all got filtered.
		candidateDefs = rawDefs
	}

	primaryDef := &candidateDefs[0]

	// Filtered usages (noise removed).
	filteredUsages := ResolveAndFilterUsages(rawUsages, candidateDefs, primaryDef, repoRoot,
		NoiseFilterOptions{})
	assertNoMigrationPathsFiltered(t, tc.lang, filteredUsages)
	assertNoFixturePathsFiltered(t, tc.lang, filteredUsages)

	// Usages with noise included.
	allUsages := ResolveAndFilterUsages(rawUsages, rawDefs, primaryDef, repoRoot,
		NoiseFilterOptions{IncludeNoise: true})
	assertNoisyPathsPresentWhenIncluded(t, tc.lang, allUsages)

	// Filtered should be a strict subset (fewer).
	if len(rawUsages) > 0 {
		assertMoreResultsWithNoise(t, tc.lang, filteredUsages, allUsages)
	}

	// -----------------------------------------------------------------------
	// Phase 3: Confidence field sanity
	// -----------------------------------------------------------------------
	for _, r := range filteredDefs {
		if r.Confidence < 0 || r.Confidence > 1.0 {
			t.Errorf("[%s] confidence %.4f out of [0,1]: %s", tc.lang, r.Confidence, r.Location.Path)
		}
	}
	for _, u := range filteredUsages {
		if u.ResolutionConfidence < 0 || u.ResolutionConfidence > 1.0 {
			t.Errorf("[%s] resolutionConfidence %.4f out of [0,1]: %s",
				tc.lang, u.ResolutionConfidence, u.Location.Path)
		}
	}
}

// ---------------------------------------------------------------------------
// Top-level test entry points (one per language)
// ---------------------------------------------------------------------------

func TestNameCollisionNoise_Go(t *testing.T) {
	runNameCollisionTest(t, langFixtureSetup{
		lang:          "go",
		ext:           ".go",
		writeFixtures: writeGoCollisionFixtures,
	})
}

func TestNameCollisionNoise_Java(t *testing.T) {
	runNameCollisionTest(t, langFixtureSetup{
		lang:          "java",
		ext:           ".java",
		writeFixtures: writeJavaCollisionFixtures,
	})
}

func TestNameCollisionNoise_Rust(t *testing.T) {
	runNameCollisionTest(t, langFixtureSetup{
		lang:          "rust",
		ext:           ".rs",
		writeFixtures: writeRustCollisionFixtures,
	})
}

func TestNameCollisionNoise_TypeScript(t *testing.T) {
	runNameCollisionTest(t, langFixtureSetup{
		lang:          "typescript",
		ext:           ".ts",
		writeFixtures: writeTypeScriptCollisionFixtures,
	})
}

func TestNameCollisionNoise_JavaScript(t *testing.T) {
	runNameCollisionTest(t, langFixtureSetup{
		lang:          "javascript",
		ext:           ".js",
		writeFixtures: writeJavaScriptCollisionFixtures,
	})
}

// ---------------------------------------------------------------------------
// Unit tests for isNoisyPath and RankDefinitions building blocks
// ---------------------------------------------------------------------------

func TestIsNoisyPath(t *testing.T) {
	cases := []struct {
		path  string
		noisy bool
	}{
		{"models/user_model.go", false},
		{"services/user_service.go", false},
		{"migrations/m0001_users.go", true},
		{"migration/add_column.go", true},
		{"db/migrate/v2.go", true},
		{"tests/fixtures/helper.go", true},
		{"testdata/sample.go", true},
		{"test/helpers.go", true},
		{"mocks/user_mock.go", true},
		{"stubs/stub_user.go", true},
		{"generated/pb/user.go", true},
		{"vendor/github.com/foo/bar.go", true},
		{"user_test.go", true},
		{"user.test.ts", true},
		{"user.spec.ts", true},
		{"internal/indexer/navigation.go", false},
	}
	for _, tc := range cases {
		got := isNoisyPath(tc.path)
		if got != tc.noisy {
			t.Errorf("isNoisyPath(%q) = %v, want %v", tc.path, got, tc.noisy)
		}
	}
}

func TestKindPriority_TypeOverVariable(t *testing.T) {
	typeKinds := []string{"class", "struct", "interface", "enum", "trait", "type_alias"}
	valueLike := []string{"variable", "field", "property", "parameter"}

	for _, tk := range typeKinds {
		for _, vk := range valueLike {
			if kindPriority(tk) <= kindPriority(vk) {
				t.Errorf("kindPriority(%q)=%.2f should be > kindPriority(%q)=%.2f",
					tk, kindPriority(tk), vk, kindPriority(vk))
			}
		}
	}
}

func TestRankDefinitions_ValueDroppedWhenTypeExists(t *testing.T) {
	defs := []DefinitionResult{
		{Name: "Foo", Kind: "class", Location: Location{Path: "models/foo.go"}},
		{Name: "Foo", Kind: "variable", Location: Location{Path: "migrations/m001.go"}},
		{Name: "Foo", Kind: "variable", Location: Location{Path: "migrations/m002.go"}},
	}
	ranked := RankDefinitions(defs, "", NoiseFilterOptions{})
	for _, r := range ranked {
		if isValueKind(r.Kind) {
			t.Errorf("value-kind %q should be filtered when type-kind candidate exists", r.Kind)
		}
	}
	if len(ranked) != 1 {
		t.Errorf("expected exactly 1 result after filtering, got %d", len(ranked))
	}
}

func TestRankDefinitions_NeverEmptyFallback(t *testing.T) {
	// All candidates are migration variables — even so, we must not return empty.
	defs := []DefinitionResult{
		{Name: "Foo", Kind: "variable", Location: Location{Path: "migrations/m001.go"}},
		{Name: "Foo", Kind: "variable", Location: Location{Path: "migrations/m002.go"}},
	}
	ranked := RankDefinitions(defs, "", NoiseFilterOptions{})
	if len(ranked) == 0 {
		t.Error("RankDefinitions must not return empty when input is non-empty")
	}
}

func TestRankDefinitions_LocalityBoost(t *testing.T) {
	defs := []DefinitionResult{
		{Name: "Bar", Kind: "function", Location: Location{Path: "services/other.go"}},
		{Name: "Bar", Kind: "function", Location: Location{Path: "models/bar.go"}},
	}
	// filterFile is in models/ — should boost models/bar.go
	ranked := RankDefinitions(defs, "models/bar.go", NoiseFilterOptions{IncludeNoise: true})
	if len(ranked) < 2 {
		t.Skip("fewer than 2 results")
	}
	if !strings.Contains(ranked[0].Location.Path, "models/") {
		t.Errorf("expected models/bar.go at top due to locality boost, got %s", ranked[0].Location.Path)
	}
}

func TestRankDefinitions_IncludeNoiseRetainsAll(t *testing.T) {
	defs := []DefinitionResult{
		{Name: "X", Kind: "class", Location: Location{Path: "models/x.go"}},
		{Name: "X", Kind: "variable", Location: Location{Path: "migrations/m001.go"}},
		{Name: "X", Kind: "variable", Location: Location{Path: "tests/fixtures/x.go"}},
	}
	ranked := RankDefinitions(defs, "", NoiseFilterOptions{IncludeNoise: true})
	if len(ranked) != 3 {
		t.Errorf("IncludeNoise=true should return all 3 candidates, got %d", len(ranked))
	}
}

func TestRankDefinitions_ConfidenceInBounds(t *testing.T) {
	defs := []DefinitionResult{
		{Name: "Z", Kind: "class", Location: Location{Path: "models/z.go"}},
		{Name: "Z", Kind: "variable", Location: Location{Path: "migrations/m001.go"}},
		{Name: "Z", Kind: "function", Location: Location{Path: "services/z.go"}},
	}
	ranked := RankDefinitions(defs, "", NoiseFilterOptions{IncludeNoise: true})
	for _, r := range ranked {
		if r.Confidence < 0 || r.Confidence > 1.0 {
			t.Errorf("confidence %.4f out of [0,1] for %s", r.Confidence, r.Location.Path)
		}
	}
}

func TestNoisePenalty_MigrationHighestPenalty(t *testing.T) {
	migPenalty := noisePenalty("migrations/0045_add_column.py")
	testPenalty := noisePenalty("tests/fixtures/helper.py")
	modelPenalty := noisePenalty("models/user_model.py")

	if modelPenalty != 0 {
		t.Errorf("models/ should have zero penalty, got %.2f", modelPenalty)
	}
	if migPenalty <= testPenalty {
		t.Errorf("migration penalty (%.2f) should exceed test-fixture penalty (%.2f)", migPenalty, testPenalty)
	}
	if migPenalty < 0.4 {
		t.Errorf("migration penalty %.2f too low, want >= 0.4", migPenalty)
	}
}
