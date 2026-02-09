package mcpstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const stateFileName = "mcp.state"

// State represents the MCP server state file
type State struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"startedAt"`
}

// StatePath returns the path to the MCP state file
func StatePath(mesdxDir string) string {
	return filepath.Join(mesdxDir, stateFileName)
}

// CreateStateFile writes the MCP state file indicating the server is running
func CreateStateFile(mesdxDir string) error {
	state := State{
		PID:       os.Getpid(),
		StartedAt: time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	statePath := StatePath(mesdxDir)
	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// RemoveStateFile removes the MCP state file
func RemoveStateFile(mesdxDir string) error {
	statePath := StatePath(mesdxDir)
	if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove state file: %w", err)
	}
	return nil
}

// IsRunning checks if the MCP server is currently running by checking the state file
// and verifying the process is still alive
func IsRunning(mesdxDir string) (bool, *State, error) {
	statePath := StatePath(mesdxDir)

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// State file is corrupted, remove it
		_ = os.Remove(statePath)
		return false, nil, nil
	}

	// Check if the process is still running
	process, err := os.FindProcess(state.PID)
	if err != nil {
		// Process doesn't exist, remove stale state file
		_ = os.Remove(statePath)
		return false, nil, nil
	}

	// On Unix systems, Signal(0) can be used to check if process is alive
	// without actually sending a signal
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process is not running, remove stale state file
		_ = os.Remove(statePath)
		return false, nil, nil
	}

	return true, &state, nil
}
