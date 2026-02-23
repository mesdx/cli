package scmsearch

import (
	"testing"

	"github.com/mesdx/cli/internal/treesitter"
)

func TestQueryCache_PutAndGet(t *testing.T) {
	c := NewQueryCache(4)
	defer c.Close()

	lang, err := treesitter.LoadLanguage("go")
	if err != nil {
		t.Fatal(err)
	}
	q, err := treesitter.NewQuery(lang, `(identifier) @id`)
	if err != nil {
		t.Fatal(err)
	}

	c.Put("go", `(identifier) @id`, lang, q)

	got := c.Get("go", `(identifier) @id`)
	if got == nil {
		t.Error("expected cached query, got nil")
	}

	h, m := c.Stats()
	if h != 1 {
		t.Errorf("expected 1 hit, got %d", h)
	}
	if m != 0 {
		t.Errorf("expected 0 misses after hit, got %d", m)
	}
}

func TestQueryCache_Miss(t *testing.T) {
	c := NewQueryCache(4)
	defer c.Close()

	got := c.Get("go", `(nonexistent)`)
	if got != nil {
		t.Error("expected nil for cache miss")
	}

	_, m := c.Stats()
	if m != 1 {
		t.Errorf("expected 1 miss, got %d", m)
	}
}

func TestQueryCache_Eviction(t *testing.T) {
	c := NewQueryCache(2)
	defer c.Close()

	lang, err := treesitter.LoadLanguage("go")
	if err != nil {
		t.Fatal(err)
	}

	q1, _ := treesitter.NewQuery(lang, `(identifier) @a`)
	q2, _ := treesitter.NewQuery(lang, `(identifier) @b`)
	q3, _ := treesitter.NewQuery(lang, `(identifier) @c`)

	c.Put("go", `(identifier) @a`, lang, q1)
	c.Put("go", `(identifier) @b`, lang, q2)
	c.Put("go", `(identifier) @c`, lang, q3)

	if c.Get("go", `(identifier) @a`) != nil {
		t.Error("expected first entry to be evicted")
	}
	if c.Get("go", `(identifier) @b`) == nil {
		t.Error("expected second entry to still be cached")
	}
	if c.Get("go", `(identifier) @c`) == nil {
		t.Error("expected third entry to still be cached")
	}
}

func TestQueryCache_DuplicatePut(t *testing.T) {
	c := NewQueryCache(4)
	defer c.Close()

	lang, err := treesitter.LoadLanguage("go")
	if err != nil {
		t.Fatal(err)
	}

	q, _ := treesitter.NewQuery(lang, `(identifier) @id`)
	c.Put("go", `(identifier) @id`, lang, q)
	c.Put("go", `(identifier) @id`, lang, q)

	if len(c.entries) != 1 {
		t.Errorf("expected 1 entry after duplicate put, got %d", len(c.entries))
	}
}
