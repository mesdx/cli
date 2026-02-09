package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateLoadAndSave(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Override the state path for testing
	originalStatePathFunc := statePathFunc
	statePathFunc = func() (string, error) {
		return filepath.Join(tmpDir, "mesdx", stateFileName), nil
	}
	defer func() { statePathFunc = originalStatePathFunc }()

	// Load state (should create new empty state)
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if state.WarnedForVersion == nil {
		t.Error("LoadState() returned state with nil WarnedForVersion map")
	}

	// Modify state
	state.MarkChecked()
	state.MarkWarnedForVersion("v1.2.3")

	// Save state
	if err := state.SaveState(); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Load again and verify
	state2, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() second call error = %v", err)
	}

	if state2.WarnedForVersion["v1.2.3"] != true {
		t.Error("State was not persisted correctly")
	}

	if state2.LastCheck.IsZero() {
		t.Error("LastCheck was not persisted")
	}
}

func TestStateShouldCheck(t *testing.T) {
	state := &State{
		LastCheck:        time.Now().Add(-25 * time.Hour),
		WarnedForVersion: make(map[string]bool),
	}

	// Should check after 24h
	if !state.ShouldCheck(24 * time.Hour) {
		t.Error("ShouldCheck(24h) = false, want true (last check was 25h ago)")
	}

	// Should not check if checked recently
	state.LastCheck = time.Now().Add(-1 * time.Hour)
	if state.ShouldCheck(24 * time.Hour) {
		t.Error("ShouldCheck(24h) = true, want false (last check was 1h ago)")
	}
}

func TestStateShouldWarnForVersion(t *testing.T) {
	state := &State{
		WarnedForVersion: map[string]bool{
			"v1.0.0": true,
		},
	}

	// Should not warn for already-warned version
	if state.ShouldWarnForVersion("v1.0.0") {
		t.Error("ShouldWarnForVersion(v1.0.0) = true, want false (already warned)")
	}

	// Should warn for new version
	if !state.ShouldWarnForVersion("v1.0.1") {
		t.Error("ShouldWarnForVersion(v1.0.1) = false, want true (not warned yet)")
	}
}

func TestStateMarkWarnedForVersion(t *testing.T) {
	state := &State{
		WarnedForVersion: make(map[string]bool),
	}

	version := "v1.2.3"
	if state.WarnedForVersion[version] {
		t.Error("Initial state should not have warned for version")
	}

	state.MarkWarnedForVersion(version)
	if !state.WarnedForVersion[version] {
		t.Error("MarkWarnedForVersion() did not mark version as warned")
	}
}

func TestStateLoadCorruptFile(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()

	// Create a corrupt state file
	stateFile := filepath.Join(tmpDir, "mesdx", stateFileName)
	if err := os.MkdirAll(filepath.Dir(stateFile), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(stateFile, []byte("not valid json{{{"), 0644); err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Override the state path for testing
	originalStatePathFunc := statePathFunc
	statePathFunc = func() (string, error) {
		return stateFile, nil
	}
	defer func() { statePathFunc = originalStatePathFunc }()

	// Should return empty state instead of failing
	state, err := LoadState()
	if err != nil {
		t.Fatalf("LoadState() with corrupt file should not error, got: %v", err)
	}

	if state.WarnedForVersion == nil {
		t.Error("LoadState() should initialize WarnedForVersion map even with corrupt file")
	}
}
