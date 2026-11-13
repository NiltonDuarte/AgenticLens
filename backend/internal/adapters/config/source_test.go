package config_test

import (
	"agenticlens/internal/adapters/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveTelemetrySourcesRequiresExplicitSelection(t *testing.T) {
	for _, tc := range []struct {
		explicit, configured, want string
		valid                      bool
	}{{"codex", "opencode", "codex", true}, {"", "opencode", "opencode", true}, {" OPENCODE ", "codex", "opencode", true}, {"", "", "", false}, {"bad", "codex", "", false}, {"", "bad", "", false}} {
		got, err := config.ResolveSource(tc.explicit, tc.configured)
		if (err == nil) != tc.valid || got != tc.want {
			t.Errorf("selection %+v got %q %v", tc, got, err)
		}
	}
	for _, resolve := range []func(string, string) (string, error){config.ResolveOpenCodePath, config.ResolveCodexHome} {
		for _, tc := range []struct{ explicit, configured, want string }{{"selected-missing", "usable-configured", "selected-missing"}, {"", "explicit-configured", "explicit-configured"}, {" selected ", "other", "selected"}} {
			got, err := resolve(tc.explicit, tc.configured)
			if err != nil || got != tc.want {
				t.Fatalf("path choice changed: %q %v", got, err)
			}
		}
		if _, err := resolve("", ""); err == nil {
			t.Fatal("invented source store")
		}
		if _, err := resolve("bad\x00path", "usable"); err == nil {
			t.Fatal("invalid selected path used configured alternative")
		}
	}
}

func TestExplicitHomeAndPhysicalConfiguredPaths(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
	got, err := config.ResolveCodexHome("~/chosen", "other")
	if err != nil || got != filepath.Join(home, "chosen") {
		t.Fatalf("explicit home %q %v", got, err)
	}
	origin := filepath.Join(root, "origin")
	target := filepath.Join(root, "target")
	if err := os.MkdirAll(filepath.Join(target, "inner"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(origin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(target, "inner"), filepath.Join(origin, "link")); err != nil {
		t.Fatal(err)
	}
	writeConfig(t, target, "source", "selected")
	writeConfig(t, origin, "source", "decoy")
	resolved, err := config.ResolveConfiguredPath(origin, "link/../source")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(resolved)
	if err != nil || string(data) != "selected" {
		t.Fatalf("symlink/.. semantics changed: %s %q %v", resolved, data, err)
	}
	absolute := filepath.Join(target, "source")
	got, err = config.ResolveConfiguredPath(origin, absolute)
	if err != nil || got != absolute {
		t.Fatalf("absolute path changed: %q %v", got, err)
	}
	t.Setenv("HOME", "")
	if _, err := config.ResolveCodexHome("~/chosen", absolute); err == nil || !strings.Contains(err.Error(), "home") {
		t.Fatalf("unavailable expansion selected alternative: %v", err)
	}
}
