package scmsearch

import (
	"testing"
)

func TestLookupStub(t *testing.T) {
	ids := []string{
		"defs.function.named",
		"defs.class.named",
		"defs.interface.named",
		"refs.type.named",
		"refs.call.named",
		"defs.method.named",
		"refs.write.named",
		"refs.import.named",
	}
	for _, id := range ids {
		s := LookupStub(id)
		if s == nil {
			t.Errorf("stub %q not found in registry", id)
		}
	}
}

func TestLookupStub_NotFound(t *testing.T) {
	s := LookupStub("nonexistent")
	if s != nil {
		t.Error("expected nil for nonexistent stub")
	}
}

func TestListStubs(t *testing.T) {
	stubs := ListStubs()
	if len(stubs) < 8 {
		t.Errorf("expected at least 8 stubs, got %d", len(stubs))
	}
}

func TestStubRender_AllLanguages(t *testing.T) {
	languages := []string{"go", "java", "rust", "python", "typescript", "javascript"}
	for _, s := range ListStubs() {
		for _, lang := range languages {
			args := map[string]string{}
			for _, a := range s.Args {
				args[a] = "TestValue"
			}
			rendered, err := s.Render(lang, args)
			if err != nil {
				t.Errorf("stub %q failed for language %q: %v", s.ID, lang, err)
				continue
			}
			if rendered == "" {
				t.Errorf("stub %q rendered empty for language %q", s.ID, lang)
			}
			if containsPlaceholder(rendered) {
				t.Errorf("stub %q has unresolved placeholders for %q: %s", s.ID, lang, rendered)
			}
		}
	}
}

func containsPlaceholder(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '{' && s[i+1] == '{' {
			return true
		}
	}
	return false
}

func TestStubRender_MissingArg(t *testing.T) {
	s := LookupStub("defs.function.named")
	_, err := s.Render("go", map[string]string{})
	if err == nil {
		t.Error("expected error for missing arg")
	}
}

func TestStubRender_UnsupportedLanguage(t *testing.T) {
	s := LookupStub("defs.function.named")
	_, err := s.Render("haskell", map[string]string{"name": "foo"})
	if err == nil {
		t.Error("expected error for unsupported language")
	}
}
