package indexer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mesdx/cli/internal/db"
)

// sharedNav is a Navigator backed by a single fully-indexed DB built once for
// the whole test binary. All read-only test helpers use this instead of
// spinning up their own FullIndex, which was the primary cause of the test
// suite taking 3+ minutes under -race.
var sharedNav *Navigator

// sharedRepoRoot is the testdata directory used by sharedNav.
var sharedRepoRoot string

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	wd, err := os.Getwd()
	if err != nil {
		panic("TestMain: os.Getwd: " + err.Error())
	}
	repoRoot := filepath.Join(wd, "testdata")
	if _, err := os.Stat(repoRoot); err != nil {
		panic("TestMain: testdata dir not found at " + repoRoot)
	}

	tmpDir, err := os.MkdirTemp("", "indexer-shared-test-*")
	if err != nil {
		panic("TestMain: MkdirTemp: " + err.Error())
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	dbPath := filepath.Join(tmpDir, "shared.db")
	if err := db.Initialize(dbPath); err != nil {
		panic("TestMain: db.Initialize: " + err.Error())
	}
	d, err := db.Open(dbPath)
	if err != nil {
		panic("TestMain: db.Open: " + err.Error())
	}
	defer func() { _ = d.Close() }()

	idx := New(d, repoRoot)
	if _, err := idx.FullIndex([]string{"."}); err != nil {
		panic("TestMain: FullIndex: " + err.Error())
	}

	sharedNav = &Navigator{DB: d, ProjectID: idx.Store.ProjectID, RepoRoot: repoRoot}
	sharedRepoRoot = repoRoot

	return m.Run()
}
