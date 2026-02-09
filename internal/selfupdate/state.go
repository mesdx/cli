package selfupdate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	stateFileName = "mesdx-update.json"
	defaultTTL    = 24 * time.Hour
)

// statePathFunc allows overriding for tests
var statePathFunc = defaultStatePath

// State tracks when we last checked for updates and which versions we've warned about
type State struct {
	LastCheck        time.Time       `json:"last_check"`
	WarnedForVersion map[string]bool `json:"warned_for_version,omitempty"`
}

// LoadState reads the cached state from UserCacheDir
func LoadState() (*State, error) {
	path, err := statePathFunc()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return &State{
				WarnedForVersion: make(map[string]bool),
			}, nil
		}
		return nil, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		// If corrupt, return empty state
		return &State{
			WarnedForVersion: make(map[string]bool),
		}, nil
	}

	if state.WarnedForVersion == nil {
		state.WarnedForVersion = make(map[string]bool)
	}

	return &state, nil
}

// SaveState writes the state to UserCacheDir
func (s *State) SaveState() error {
	path, err := statePathFunc()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ShouldCheck returns true if we should check for updates based on TTL
func (s *State) ShouldCheck(ttl time.Duration) bool {
	return time.Since(s.LastCheck) >= ttl
}

// MarkChecked updates the last check time
func (s *State) MarkChecked() {
	s.LastCheck = time.Now()
}

// ShouldWarnForVersion returns true if we haven't warned about this version yet
func (s *State) ShouldWarnForVersion(version string) bool {
	return !s.WarnedForVersion[version]
}

// MarkWarnedForVersion records that we've warned about this version
func (s *State) MarkWarnedForVersion(version string) {
	s.WarnedForVersion[version] = true
}

func defaultStatePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "mesdx", stateFileName), nil
}
