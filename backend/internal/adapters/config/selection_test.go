package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectedLocalOriginAcrossLinkedWorktrees(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "base")
	one := filepath.Join(root, "one")
	two := filepath.Join(root, "two")
	if err := os.Mkdir(base, 0700); err != nil {
		t.Fatal(err)
	}
	git := func(dir string, args ...string) {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s %v", args, out, err)
		}
	}
	write := func(path, body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	git(base, "init", "-b", "main")
	shared := "project_name: Portable\nagenticlens:\n  version: v9.8.7\nworkspaces:\n  dir: custom-records\nticket:\n  prefixes: [CUSTOM]\ncode_metrics:\n  source_roots: [custom-src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	write(filepath.Join(base, ".agenticlens-config.yaml"), shared)
	git(base, "add", ".agenticlens-config.yaml")
	git(base, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-m", "fixture")
	git(base, "worktree", "add", "-b", "one", one)
	git(base, "worktree", "add", "-b", "two", two)
	if _, err := InitialLocalYAML(one, "true", "codex", "source", ""); err == nil {
		t.Fatal("guided setup created worktree-local policy")
	}
	if _, err := os.Stat(filepath.Join(one, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
		t.Fatalf("worktree setup wrote local file: %v", err)
	}
	nested := filepath.Join(one, "src", "nested")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	localPath := filepath.Join(base, ".agenticlens-config.local.yaml")
	write(localPath, "telemetry:\n  capture:\n    enabled: true\n    state_dir: base-state\n    max_sessions: 11\n  sources:\n    codex:\n      home: base-codex\n")
	for _, cwd := range []string{base, nested, two} {
		selection, err := SelectConfig(cwd)
		if err != nil {
			t.Fatal(err)
		}
		expected := cwd
		if cwd == nested {
			expected = one
		}
		if selection.Project != expected || selection.LocalPath != localPath {
			t.Fatalf("wrong whole-file selection: %+v", selection)
		}
		policy, err := ResolveCapture(cwd, "")
		if err != nil || policy.StateDir != filepath.Join(base, "base-state") || policy.MaxSessions != 11 {
			t.Fatalf("base-relative policy: %+v %v", policy, err)
		}
		cfg, err := Load(cwd)
		if err != nil || cfg.Telemetry.Sources.CodexHome != filepath.Join(base, "base-codex") || cfg.Workspaces.Dir != "custom-records" || cfg.Ticket.Prefixes[0].Prefix != "CUSTOM" || cfg.CodeMetrics.SourceRoots[0] != "custom-src" {
			t.Fatalf("project/source origin lost: %+v %v", cfg, err)
		}
	}
	local := filepath.Join(one, ".agenticlens-config.local.yaml")
	write(local, "telemetry:\n  capture:\n    enabled: false\n    state_dir: local-state\n")
	policy, err := ResolveCapture(nested, "")
	if err != nil || policy.Enabled || policy.StateDir != filepath.Join(one, "local-state") || policy.MaxSessions != 256 {
		t.Fatalf("base key leaked into local: %+v %v", policy, err)
	}
	cfg, err := Load(nested)
	if err != nil || cfg.Telemetry.Sources.CodexHome != "" {
		t.Fatalf("source inherited from base: %+v %v", cfg.Telemetry.Sources, err)
	}
	write(local, "telemetry:\n  capture:\n    enabled: false\n")
	if _, err := ResolveCapture(nested, ""); err == nil || !strings.Contains(err.Error(), "state_dir") {
		t.Fatalf("incomplete local inherited base: %v", err)
	}
	write(local, "telemetry: [broken")
	if _, err := ResolveCapture(nested, ""); err == nil || !strings.Contains(err.Error(), local) {
		t.Fatalf("invalid local fell back: %v", err)
	}
	write(local, "telemetry:\n  capture:\n    enabled: false\n    state_dir: local-state\n")
	if err := os.Remove(localPath); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveCapture(nested, ""); err != nil {
		t.Fatalf("valid worktree policy required base: %v", err)
	}
	if _, err := ResolveCapture(two, ""); err == nil || !strings.Contains(err.Error(), localPath) {
		t.Fatalf("missing fallback hidden: %v", err)
	}
	if err := os.Remove(filepath.Join(one, ".agenticlens-config.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, err := SelectConfig(nested); err != ErrConfigNotFound {
		t.Fatalf("base shared used as fallback: %v", err)
	}
	for _, path := range []string{filepath.Join(base, "base-state"), filepath.Join(one, "local-state")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("read allocated state %s: %v", path, err)
		}
	}
}

func TestCaptureEnablementMustBeExplicitBoolean(t *testing.T) {
	for _, value := range []string{"missing", "null", "\"false\"", "0", "[]"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte("project_name: Fixture\n"), 0600); err != nil {
				t.Fatal(err)
			}
			body := "telemetry:\n  capture:\n    state_dir: unused\n"
			if value != "missing" {
				body += "    enabled: " + value + "\n"
			}
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := ResolveCapture(root, ""); err == nil {
				t.Fatalf("nonboolean accepted: %s", value)
			}
		})
	}
}
