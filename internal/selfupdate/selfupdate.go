package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubRepo = "mesdx/cli"
	githubAPI  = "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	timeout    = 10 * time.Second
)

// Release represents a GitHub release
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckAndUpdate checks for a newer version and attempts to update
// Returns true if an update was performed, false otherwise
func CheckAndUpdate(currentVersion string) error {
	// Skip if disabled
	if os.Getenv("MESDX_NO_SELF_UPDATE") == "1" {
		return nil
	}

	// Skip dev builds
	if currentVersion == "" || currentVersion == "dev" || !strings.HasPrefix(currentVersion, "v") {
		return nil
	}

	// Load state
	state, err := LoadState()
	if err != nil {
		// Don't fail the command if we can't load state
		return nil
	}

	// Check TTL
	if !state.ShouldCheck(defaultTTL) {
		return nil
	}

	// Fetch latest release
	release, err := fetchLatestRelease()
	if err != nil {
		// Update check time even on error to avoid hammering GitHub
		state.MarkChecked()
		_ = state.SaveState()
		return nil // Don't fail the command
	}

	// Update check time
	state.MarkChecked()
	_ = state.SaveState()

	// Check if newer version available
	if release.TagName == currentVersion {
		return nil // Already up to date
	}

	// Find the correct asset for this platform
	assetName := assetNameForPlatform()
	asset := findAsset(release.Assets, assetName)
	if asset == nil {
		// No asset for this platform
		return nil
	}

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return nil
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return nil
	}

	// Check if we can write to the executable location
	if !canWriteToPath(exePath) {
		// Only warn once per version
		if state.ShouldWarnForVersion(release.TagName) {
			printManualUpdateInstructions(currentVersion, release.TagName, assetName, asset.BrowserDownloadURL)
			state.MarkWarnedForVersion(release.TagName)
			_ = state.SaveState()
		}
		return nil
	}

	// Attempt in-place update
	if err := downloadAndReplace(asset.BrowserDownloadURL, exePath); err != nil {
		// Don't fail the command, but we could log if we had a logger
		return nil
	}

	fmt.Printf("✓ Updated mesdx from %s to %s\n", currentVersion, release.TagName)
	fmt.Println("  Please re-run your command to use the new version.")

	return nil
}

func fetchLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(githubAPI)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func assetNameForPlatform() string {
	return fmt.Sprintf("mesdx-%s-%s", runtime.GOOS, runtime.GOARCH)
}

func findAsset(assets []Asset, name string) *Asset {
	for _, asset := range assets {
		if asset.Name == name {
			return &asset
		}
	}
	return nil
}

func canWriteToPath(path string) bool {
	// Try to open for writing to check permissions
	// We'll try opening the directory with a test file
	dir := filepath.Dir(path)
	testFile := filepath.Join(dir, ".mesdx-write-test")

	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(testFile)

	return true
}

func downloadAndReplace(url, targetPath string) error {
	// Download to temp file in same directory for atomic rename
	dir := filepath.Dir(targetPath)
	tmpFile, err := os.CreateTemp(dir, ".mesdx-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }() // Clean up on error

	// Download
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		_ = tmpFile.Close()
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Copy to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return err
	}

	return nil
}

func printManualUpdateInstructions(currentVersion, newVersion, assetName, downloadURL string) {
	fmt.Printf("\n")
	fmt.Printf("╭─────────────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│ Update Available: %s → %s                          \n", currentVersion, newVersion)
	fmt.Printf("╰─────────────────────────────────────────────────────────────────╯\n")
	fmt.Printf("\n")
	fmt.Printf("The mesdx binary is installed in a location that requires\n")
	fmt.Printf("elevated permissions to update. To update manually, run:\n")
	fmt.Printf("\n")
	fmt.Printf("  curl -L %s -o /tmp/mesdx\n", downloadURL)
	fmt.Printf("  chmod +x /tmp/mesdx\n")

	// Try to detect install location
	exePath, err := os.Executable()
	if err == nil {
		exePath, _ = filepath.EvalSymlinks(exePath)
		fmt.Printf("  sudo mv /tmp/mesdx %s\n", exePath)
	} else {
		fmt.Printf("  sudo mv /tmp/mesdx /usr/local/bin/mesdx\n")
	}

	fmt.Printf("\n")
}
