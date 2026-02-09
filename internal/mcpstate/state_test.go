package mcpstate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateFileLifecycle(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "mcpstate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test 1: No state file initially
	running, state, err := IsRunning(tmpDir)
	if err != nil {
		t.Errorf("IsRunning should not error on missing file: %v", err)
	}
	if running {
		t.Error("IsRunning should return false when no state file exists")
	}
	if state != nil {
		t.Error("State should be nil when no state file exists")
	}

	// Test 2: Create state file
	if err := CreateStateFile(tmpDir); err != nil {
		t.Fatalf("CreateStateFile failed: %v", err)
	}

	// Verify state file was created
	statePath := StatePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("State file was not created")
	}

	// Test 3: Check running state (should be true for current process)
	running, state, err = IsRunning(tmpDir)
	if err != nil {
		t.Errorf("IsRunning failed: %v", err)
	}
	if !running {
		t.Error("IsRunning should return true for current process")
	}
	if state == nil {
		t.Fatal("State should not be nil")
	}
	if state.PID != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), state.PID)
	}
	if state.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}

	// Test 4: Remove state file
	if err := RemoveStateFile(tmpDir); err != nil {
		t.Errorf("RemoveStateFile failed: %v", err)
	}

	// Verify state file was removed
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("State file was not removed")
	}

	// Test 5: Check after removal
	running, state, err = IsRunning(tmpDir)
	if err != nil {
		t.Errorf("IsRunning should not error after removal: %v", err)
	}
	if running {
		t.Error("IsRunning should return false after removal")
	}
	if state != nil {
		t.Error("State should be nil after removal")
	}
}

func TestStateFileWithDeadProcess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpstate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a state file with a PID that's very unlikely to exist
	// Using a very high PID that should not exist
	statePath := StatePath(tmpDir)
	data := []byte(`{"pid": 999999, "startedAt": "2026-02-09T10:00:00Z"}`)
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		t.Fatalf("failed to write test state file: %v", err)
	}

	// IsRunning should detect the process is dead and clean up the file
	running, resultState, err := IsRunning(tmpDir)
	if err != nil {
		t.Errorf("IsRunning should not error on dead process: %v", err)
	}
	if running {
		t.Error("IsRunning should return false for dead process")
	}
	if resultState != nil {
		t.Error("State should be nil for dead process")
	}

	// State file should be removed
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("State file for dead process should be removed")
	}
}

func TestStateFileCorrupted(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcpstate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a corrupted state file
	statePath := StatePath(tmpDir)
	if err := os.WriteFile(statePath, []byte("corrupted json{{{"), 0644); err != nil {
		t.Fatalf("failed to write test state file: %v", err)
	}

	// IsRunning should handle corruption gracefully
	running, state, err := IsRunning(tmpDir)
	if err != nil {
		t.Errorf("IsRunning should not error on corrupted file: %v", err)
	}
	if running {
		t.Error("IsRunning should return false for corrupted file")
	}
	if state != nil {
		t.Error("State should be nil for corrupted file")
	}

	// Corrupted file should be removed
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("Corrupted state file should be removed")
	}
}

func TestStatePath(t *testing.T) {
	testDir := "/test/dir"
	expected := filepath.Join("/test/dir", stateFileName)
	result := StatePath(testDir)
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
