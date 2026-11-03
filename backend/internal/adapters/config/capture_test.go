package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCapturePolicyPrecedenceAndProjectRelativeState(t *testing.T) {
	project := t.TempDir()
	for name, body := range map[string]string{".agenticlens-config.yaml": "telemetry:\n  capture:\n    refresh_ms: 800\n    state_dir: portable\n    detailed_report: true\n", ".agenticlens-config.local.yaml": "telemetry:\n  capture:\n    state_dir: local\n    max_sessions: 7\n"} {
		if err := os.WriteFile(filepath.Join(project, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := ResolveCapture(project, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RefreshMS != 800 || cfg.MaxSessions != 7 || cfg.DetailedReport || !cfg.DeprecatedDetailedReport || cfg.StateDir != filepath.Join(project, "local") {
		t.Fatalf("wrong overlay: %+v", cfg)
	}
	override, err := ResolveCapture(project, "explicit")
	if err != nil || override.StateDir != filepath.Join(project, "explicit") {
		t.Fatalf("CLI override: %+v %v", override, err)
	}
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    max_sessions: -1\n"), 0600); err != nil {
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
