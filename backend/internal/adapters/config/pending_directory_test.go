package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestPendingDirectorySetupUsesExplicitChoice(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".agenticlens-config.yaml")
	cfg := testProjectConfig()

	selected, err := config.PendingDirectoryForSetup(path, cfg)
	if err != nil || selected != ".agenticlens/pending/" {
		t.Fatalf("new setup selected %q: %v", selected, err)
	}
	body, err := config.GenerateConfigYAML(cfg)
	if err != nil || !strings.Contains(string(body), "pending_dir: .agenticlens/pending/") {
		t.Fatalf("shared writer omitted explicit default: %s\n%v", body, err)
	}

	const custom = "tracking/pending snapshots/"
	writeConfig(t, root, filepath.Base(path), "agenticlens:\n  version: v9.8.7\n  pending_dir: '"+custom+"'\n")
	selected, err = config.PendingDirectoryForSetup(path, cfg)
	if err != nil || selected != custom {
		t.Fatalf("updating unrelated setup choice changed pending directory to %q: %v", selected, err)
	}
	cfg.AgenticLens.PendingDir = "replacement/pending"
	selected, err = config.PendingDirectoryForSetup(path, cfg)
	if err != nil || selected != cfg.AgenticLens.PendingDir {
		t.Fatalf("setup discarded explicit replacement %q: %v", selected, err)
	}
}

func TestResolvePendingDirectoryUsesPhysicalProjectAndLocalOrigins(t *testing.T) {
	root := t.TempDir()
	inner := filepath.Join(root, "nested")
	if err := os.Mkdir(inner, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := testProjectConfig()
	cfg.AgenticLens.PendingDir = "snapshots/pending"
	body, err := config.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	writeConfig(t, root, ".agenticlens-config.yaml", string(body))
	writeConfig(t, root, ".agenticlens-config.local.yaml", "agenticlens:\n  state_dir: runtime/publication\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: runtime/capture\n")
	cfg, err = config.Load(inner)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := config.ResolvePendingDirectory(cfg, inner)
	if err != nil || selected != filepath.Join(root, "snapshots", "pending") {
		t.Fatalf("nested cwd selected %q: %v", selected, err)
	}
	if _, err := os.Stat(selected); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("read initialized pending directory: %v", err)
	}
	cfg.LocalConfigPath = filepath.Join(root, "nested", ".agenticlens-config.local.yaml")
	cfg.AgenticLens.StateDir = "../snapshots/pending/journal"
	if _, err := config.ResolvePendingDirectory(cfg, inner); err == nil || !strings.Contains(err.Error(), "agenticlens.state_dir") {
		t.Fatalf("pending ancestor of publication state accepted: %v", err)
	}
	cfg.AgenticLens.StateDir = "../runtime/publication"
	cfg.Telemetry.Capture.StateDir = "../snapshots"
	if _, err := config.ResolvePendingDirectory(cfg, inner); err == nil || !strings.Contains(err.Error(), "telemetry.capture.state_dir") {
		t.Fatalf("pending descendant of capture state accepted: %v", err)
	}
}

func TestResolvePendingDirectoryRejectsUnsafeTargets(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "file", "contents")
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"", " ", "one\ntwo", ".", "..", outside, "escape/pending", ".git/pending", "file", "file/child"} {
		t.Run(value, func(t *testing.T) {
			cfg := domain.ProjectConfig{AgenticLens: domain.EntrypointConfig{PendingDir: value}}
			if _, err := config.ResolvePendingDirectory(cfg, root); err == nil {
				t.Fatal("unsafe pending path accepted")
			}
		})
	}
}

func TestPendingDirectoryMissingLegacyConfigAndInvalidSharedOrLocal(t *testing.T) {
	root := t.TempDir()
	body := strings.Replace(fixtureSharedBody(t), "  pending_dir: .agenticlens/pending/\n", "", 1)
	shared := writeConfig(t, root, ".agenticlens-config.yaml", body)
	cfg, err := config.LoadManualGeneration(root)
	if err != nil {
		t.Fatalf("missing pending setting broke unrelated legacy read: %v", err)
	}
	if _, err := config.ResolvePendingDirectory(cfg, root); err == nil || !strings.Contains(err.Error(), "agenticlens.pending_dir") {
		t.Fatalf("pending read inferred an absent setting: %v", err)
	}
	for _, value := range []string{"''", "null", "42", "[pending]"} {
		modified := strings.Replace(body, "  version: v9.8.7\n", "  version: v9.8.7\n  pending_dir: "+value+"\n", 1)
		writeConfig(t, root, filepath.Base(shared), modified)
		if _, err := config.LoadFromFile(shared); err == nil || !strings.Contains(err.Error(), "agenticlens.pending_dir") {
			t.Fatalf("invalid value %s was not rejected: %v", value, err)
		}
		if pin, err := config.ReadReleasePin(shared); err != nil || pin != "v9.8.7" {
			t.Fatalf("could not read pin for explicit setup repair: %q %v", pin, err)
		}
	}
	writeConfig(t, root, filepath.Base(shared), body)
	writeConfig(t, root, ".agenticlens-config.local.yaml", "agenticlens:\n  pending_dir: shadow\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: state\n")
	if _, err := config.ResolveCapture(root, ""); err == nil || !strings.Contains(err.Error(), "belongs only in shared YAML") {
		t.Fatalf("capture accepted machine-specific pending policy: %v", err)
	}
}
