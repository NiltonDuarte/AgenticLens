package config_test

import (
	"testing"

	"agenticlens/internal/adapters/config"
)

func TestResolveTelemetrySources(t *testing.T) {
	t.Run("source selection precedence and validation", func(t *testing.T) {
		got, err := config.ResolveSource("codex", "opencode")
		if err != nil || got != "codex" {
			t.Fatalf("explicit source should win: got=%q err=%v", got, err)
		}
		got, err = config.ResolveSource("", "codex")
		if err != nil || got != "codex" {
			t.Fatalf("configured source should win: got=%q err=%v", got, err)
		}
		got, err = config.ResolveSource("", "")
		if err != nil || got != "opencode" {
			t.Fatalf("built-in source should be OpenCode: got=%q err=%v", got, err)
		}
		if _, err := config.ResolveSource("both", ""); err == nil {
			t.Fatal("single-source selection accepted both")
		}
	})
	if got := config.ResolveOpenCodePath(" cli ", "local", "legacy"); got != "cli" {
		t.Fatalf("explicit path precedence: %q", got)
	}
	if got := config.ResolveOpenCodePath("", "local", "legacy"); got != "local" {
		t.Fatalf("local path precedence: %q", got)
	}
	if got := config.ResolveOpenCodePath("", "", "legacy"); got != "legacy" {
		t.Fatalf("legacy path precedence: %q", got)
	}
	if got := config.ResolveCodexHome("cli", "local", "env", "builtin"); got != "cli" {
		t.Fatalf("Codex path precedence: %q", got)
	}
	if got := config.ResolveCodexHome("", "local", "env", "builtin"); got != "local" {
		t.Fatalf("Codex local path precedence: %q", got)
	}
}
