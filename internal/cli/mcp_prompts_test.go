package cli

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSkillPrompts(t *testing.T) {
	// Create a test server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "v0.0.1",
	}, nil)

	// Register prompts (memory disabled for test)
	registerSkillPrompts(server, "/test/repo", false)

	// Create a client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "v0.0.1",
	}, nil)

	// Connect server and client
	t1, t2 := mcp.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("failed to connect server: %v", err)
	}
	defer func() {
		if err := serverSession.Close(); err != nil {
			t.Errorf("failed to close server session: %v", err)
		}
	}()

	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer func() {
		if err := clientSession.Close(); err != nil {
			t.Errorf("failed to close client session: %v", err)
		}
	}()

	// Test 1: List prompts
	t.Run("ListPrompts", func(t *testing.T) {
		var prompts []*mcp.Prompt
		for prompt, err := range clientSession.Prompts(ctx, nil) {
			if err != nil {
				t.Fatalf("prompts iterator error: %v", err)
			}
			prompts = append(prompts, prompt)
		}

		expectedPrompts := []string{
			"mesdx.skill.bugfix",
			"mesdx.skill.refactoring",
			"mesdx.skill.feature_development",
			"mesdx.skill.security_analysis",
		}

		if len(prompts) != len(expectedPrompts) {
			t.Fatalf("expected %d prompts, got %d", len(expectedPrompts), len(prompts))
		}

		promptNames := make(map[string]bool)
		for _, p := range prompts {
			promptNames[p.Name] = true
		}

		for _, expected := range expectedPrompts {
			if !promptNames[expected] {
				t.Errorf("missing prompt: %s", expected)
			}
		}
	})

	// Test 2: Get each prompt
	tests := []struct {
		name string
	}{
		{"mesdx.skill.bugfix"},
		{"mesdx.skill.refactoring"},
		{"mesdx.skill.feature_development"},
		{"mesdx.skill.security_analysis"},
	}

	for _, tt := range tests {
		t.Run("GetPrompt_"+tt.name, func(t *testing.T) {
			result, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
				Name: tt.name,
			})
			if err != nil {
				t.Fatalf("failed to get prompt %s: %v", tt.name, err)
			}

			if result.Description == "" {
				t.Errorf("prompt %s has empty description", tt.name)
			}

			if len(result.Messages) == 0 {
				t.Errorf("prompt %s has no messages", tt.name)
			}

			// Verify the message has content
			for _, msg := range result.Messages {
				if msg.Role != "user" {
					t.Errorf("expected role 'user', got %s", msg.Role)
				}
				textContent, ok := msg.Content.(*mcp.TextContent)
				if !ok {
					t.Errorf("expected TextContent, got %T", msg.Content)
				}
				if textContent.Text == "" {
					t.Errorf("prompt %s has empty text content", tt.name)
				}
			}
		})
	}
}

func TestSkillPromptsWithMemory(t *testing.T) {
	// Create a test server with memory enabled
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "v0.0.1",
	}, nil)

	// Register prompts (memory enabled for test)
	registerSkillPrompts(server, "/test/repo", true)

	// Create a client
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "v0.0.1",
	}, nil)

	// Connect server and client
	t1, t2 := mcp.NewInMemoryTransports()
	ctx := context.Background()

	serverSession, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("failed to connect server: %v", err)
	}
	defer func() {
		if err := serverSession.Close(); err != nil {
			t.Errorf("failed to close server session: %v", err)
		}
	}()

	clientSession, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer func() {
		if err := clientSession.Close(); err != nil {
			t.Errorf("failed to close client session: %v", err)
		}
	}()

	// Test that prompts include memory guidance when enabled
	result, err := clientSession.GetPrompt(ctx, &mcp.GetPromptParams{
		Name: "mesdx.skill.bugfix",
	})
	if err != nil {
		t.Fatalf("failed to get bugfix prompt: %v", err)
	}

	textContent, ok := result.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent")
	}

	// Verify memory-related content is present
	if !contains(textContent.Text, "mesdx.memoryAppend") {
		t.Errorf("bugfix prompt should mention memoryAppend when memory is enabled")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
