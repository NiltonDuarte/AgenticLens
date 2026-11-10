package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCapturePolicyWholeLocalFileAndProjectRelativeState(t *testing.T) {
	project := t.TempDir()
	for name, body := range map[string]string{".agenticlens-config.yaml": "telemetry:\n  capture:\n    refresh_ms: 800\n    state_dir: portable\n    detailed_report: true\n", ".agenticlens-config.local.yaml": "telemetry:\n  capture:\n    enabled: true\n    state_dir: local\n    max_sessions: 7\n"} {
		if err := os.WriteFile(filepath.Join(project, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := ResolveCapture(project, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RefreshMS != 15000 || cfg.MaxSessions != 7 || cfg.DetailedReport || cfg.DeprecatedDetailedReport || cfg.StateDir != filepath.Join(project, "local") {
		t.Fatalf("shared capture values leaked into selected local policy: %+v", cfg)
	}
	override, err := ResolveCapture(project, "explicit")
	if err != nil || override.StateDir != filepath.Join(project, "explicit") {
		t.Fatalf("CLI override: %+v %v", override, err)
	}
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    enabled: true\n    state_dir: local\n    max_sessions: -1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = ResolveCapture(project, ""); err == nil {
		t.Fatal("invalid work bound accepted")
	}
}

func TestCaptureOwningProjectFormatsAndNestedBoundary(t *testing.T) {
	for _, ext := range []string{"yaml", "yml", "json"} {
		t.Run(ext, func(t *testing.T) {
			project := t.TempDir()
			child := filepath.Join(project, "src")
			if err := os.Mkdir(child, 0700); err != nil {
				t.Fatal(err)
			}
			body := "telemetry:\n  capture:\n    state_dir: configured\n    enabled: false\n"
			if ext == "json" {
				body = `{"telemetry":{"capture":{"state_dir":"configured","enabled":false}}}`
			}
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config."+ext), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    enabled: false\n    state_dir: configured\n"), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := ResolveCapture(child, "")
			if err != nil || cfg.StateDir != filepath.Join(project, "configured") || cfg.Enabled {
				t.Fatalf("owning policy %+v %v", cfg, err)
			}
			if err := os.Mkdir(filepath.Join(child, ".git"), 0700); err != nil {
				t.Fatal(err)
			}
			root, err := ProjectRoot(child)
			if err != nil || root != child {
				t.Fatalf("crossed nested Git boundary %s %v", root, err)
			}
		})
	}
}

func TestRequiredCaptureStateInSelectedLocalPolicy(t *testing.T) {
	for _, value := range []string{"missing", `""`, `"   "`, "null", "42", "false", "{}", "[]"} {
		t.Run(value, func(t *testing.T) {
			root := t.TempDir()
			home := filepath.Join(root, "uncreated-home")
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte("project_name: Example\n"), 0600); err != nil {
				t.Fatal(err)
			}
			body := "telemetry:\n  capture:\n    enabled: false\n"
			if value != "missing" {
				body += "    state_dir: " + value + "\n"
			}
			path := filepath.Join(root, ".agenticlens-config.local.yaml")
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := ResolveCapture(root, "cli-cannot-repair-missing-setting"); err == nil || !strings.Contains(err.Error(), "telemetry.capture.state_dir") {
				t.Fatalf("invalid local state was accepted or diagnostic lost: %v", err)
			}
			actual, err := os.ReadFile(path)
			if err != nil || string(actual) != body {
				t.Fatalf("configuration changed: %v", err)
			}
			if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("invalid read created home: %v", err)
			}
		})
	}
}

func TestLocalOnlyCaptureStateAndMissingFilesystemDoNotNeedFallback(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, ".agenticlens-config.yaml")
	body := "project_name: Example\nticket:\n  prefixes: [AL]\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(shared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(root, "uncreated-home")
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	local := "telemetry:\n  capture:\n    enabled: true\n    state_dir: ~/.agenticlens/state\n    max_sessions: 7\n"
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(root)
	if err != nil || cfg.Telemetry.Capture.StateDir != "~/.agenticlens/state" || cfg.Telemetry.Capture.MaxSessions != 7 {
		t.Fatalf("effective local config: %+v %v", cfg, err)
	}
	policy, err := ResolveCapture(root, "")
	if err != nil || policy.StateDir != filepath.Join(home, ".agenticlens", "state") {
		t.Fatalf("explicit home expansion: %+v %v", policy, err)
	}
	if _, err := os.Stat(home); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read created the configured directory: %v", err)
	}
	if _, err := LoadFromFile(shared); err != nil {
		t.Fatalf("manual single-file reader requires capture state: %v", err)
	}
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	if _, err := ResolveCapture(root, ""); err == nil || !strings.Contains(err.Error(), "home") {
		t.Fatalf("unavailable home was silently replaced: %v", err)
	}
}
