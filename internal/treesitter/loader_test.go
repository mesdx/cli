package treesitter

import (
	"testing"
)

func TestLoadLanguage(t *testing.T) {
	// Test loading a supported language
	lang, err := LoadLanguage("go")
	if err != nil {
		t.Fatalf("LoadLanguage(go): %v", err)
	}
	if lang == nil {
		t.Fatal("LoadLanguage(go) returned nil language")
	}
	if lang.Name() != "go" {
		t.Errorf("Language.Name() = %q, want %q", lang.Name(), "go")
	}
}

func TestLoadLanguageUnsupported(t *testing.T) {
	// Test loading an unsupported language
	_, err := LoadLanguage("cobol")
	if err == nil {
		t.Error("LoadLanguage(cobol) should fail for unsupported language")
	}
}

func TestVerifyLanguages(t *testing.T) {
	// Should succeed for supported languages
	err := VerifyLanguages([]string{"go", "python", "typescript"})
	if err != nil {
		t.Errorf("VerifyLanguages(supported): %v", err)
	}

	// Should fail for unsupported languages
	err = VerifyLanguages([]string{"go", "cobol"})
	if err == nil {
		t.Error("VerifyLanguages() should fail when languages are unsupported")
	}
}

func TestRequiredLanguages(t *testing.T) {
	langs := RequiredLanguages()
	expected := []string{"go", "java", "rust", "python", "javascript", "typescript"}
	
	if len(langs) != len(expected) {
		t.Errorf("RequiredLanguages() returned %d languages, want %d", len(langs), len(expected))
	}

	// Check all expected languages are present
	langMap := make(map[string]bool)
	for _, l := range langs {
		langMap[l] = true
	}

	for _, exp := range expected {
		if !langMap[exp] {
			t.Errorf("RequiredLanguages() missing %q", exp)
		}
	}
}

func TestLanguageCache(t *testing.T) {
	// Load a language
	lang1, err := LoadLanguage("rust")
	if err != nil {
		t.Fatalf("LoadLanguage(rust): %v", err)
	}

	// Load again - should return cached instance
	lang2, err := LoadLanguage("rust")
	if err != nil {
		t.Fatalf("LoadLanguage(rust) second time: %v", err)
	}

	// Should be the same instance (pointer equality)
	if lang1 != lang2 {
		t.Error("LoadLanguage should return cached language instance")
	}

	// Clear cache
	CloseAll()

	// Load again - should be a new instance
	lang3, err := LoadLanguage("rust")
	if err != nil {
		t.Fatalf("LoadLanguage(rust) after CloseAll: %v", err)
	}
	if lang3 == lang1 {
		t.Error("LoadLanguage after CloseAll should return new instance")
	}
}
