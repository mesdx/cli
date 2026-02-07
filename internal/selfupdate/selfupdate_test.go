package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAssetNameForPlatform(t *testing.T) {
	expected := "codeintelx-" + runtime.GOOS + "-" + runtime.GOARCH
	actual := assetNameForPlatform()

	if actual != expected {
		t.Errorf("assetNameForPlatform() = %q, want %q", actual, expected)
	}
}

func TestFindAsset(t *testing.T) {
	assets := []Asset{
		{Name: "codeintelx-darwin-arm64", BrowserDownloadURL: "https://example.com/darwin-arm64"},
		{Name: "codeintelx-darwin-amd64", BrowserDownloadURL: "https://example.com/darwin-amd64"},
		{Name: "codeintelx-linux-amd64", BrowserDownloadURL: "https://example.com/linux-amd64"},
	}

	tests := []struct {
		name      string
		assetName string
		want      string
		wantNil   bool
	}{
		{
			name:      "find darwin-arm64",
			assetName: "codeintelx-darwin-arm64",
			want:      "https://example.com/darwin-arm64",
			wantNil:   false,
		},
		{
			name:      "find linux-amd64",
			assetName: "codeintelx-linux-amd64",
			want:      "https://example.com/linux-amd64",
			wantNil:   false,
		},
		{
			name:      "not found",
			assetName: "codeintelx-windows-amd64",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := findAsset(assets, tt.assetName)
			if tt.wantNil {
				if asset != nil {
					t.Errorf("findAsset() = %v, want nil", asset)
				}
			} else {
				if asset == nil {
					t.Errorf("findAsset() = nil, want non-nil")
				} else if asset.BrowserDownloadURL != tt.want {
					t.Errorf("findAsset().BrowserDownloadURL = %q, want %q", asset.BrowserDownloadURL, tt.want)
				}
			}
		})
	}
}

func TestCanWriteToPath(t *testing.T) {
	// Create a temp directory we can write to
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-executable")

	// Create a test file
	if err := os.WriteFile(tmpFile, []byte("test"), 0755); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should be able to write to temp directory
	if !canWriteToPath(tmpFile) {
		t.Errorf("canWriteToPath(%q) = false, want true (temp dir should be writable)", tmpFile)
	}

	// Test with a path in a non-writable location (if not root)
	if os.Getuid() != 0 {
		nonWritablePath := "/usr/local/bin/codeintelx"
		// Only test if the directory actually exists
		if _, err := os.Stat("/usr/local/bin"); err == nil {
			if canWriteToPath(nonWritablePath) {
				// This might pass if user has write access, which is OK
				t.Logf("Note: User has write access to /usr/local/bin")
			}
		}
	}
}

func TestCheckAndUpdate_SkipWhenDisabled(t *testing.T) {
	// Set env var to disable
	_ = os.Setenv("CODEINTELX_NO_SELF_UPDATE", "1")
	defer func() { _ = os.Unsetenv("CODEINTELX_NO_SELF_UPDATE") }()

	// Should not error and should skip
	err := CheckAndUpdate("v0.1.0")
	if err != nil {
		t.Errorf("CheckAndUpdate() with disabled flag returned error: %v", err)
	}
}

func TestCheckAndUpdate_SkipDevVersion(t *testing.T) {
	tests := []string{"", "dev", "main", "0.1.0"} // versions without "v" prefix

	for _, version := range tests {
		t.Run("version="+version, func(t *testing.T) {
			err := CheckAndUpdate(version)
			if err != nil {
				t.Errorf("CheckAndUpdate(%q) returned error: %v", version, err)
			}
		})
	}
}
