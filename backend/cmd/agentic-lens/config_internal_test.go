package main

import (
	adaptersconfig "agenticlens/internal/adapters/config"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteConfigPairCollisionAndForce(t *testing.T) {
	dir := t.TempDir()
	first, second := filepath.Join(dir, "shared.yaml"), filepath.Join(dir, "local.yaml")
	if err := writeConfigPair(first, []byte("shared-original"), second, []byte("local-original"), false); err != nil {
		t.Fatal(err)
	}
	err := writeConfigPair(first, []byte("shared-new"), second, []byte("local-new"), false)
	var collision *adaptersconfig.FileExistsError
	if !errors.Is(err, adaptersconfig.ErrFileExists) || !errors.As(err, &collision) || collision.Path != first {
		t.Fatalf("expected typed first-file collision, got %v", err)
	}
	for _, item := range []struct{ path, want string }{{first, "shared-original"}, {second, "local-original"}} {
		got, err := os.ReadFile(item.path)
		if err != nil || string(got) != item.want {
			t.Fatalf("collision modified %s: %q, %v", item.path, got, err)
		}
	}
	if err := writeConfigPair(first, []byte("shared-new"), second, []byte("local-new"), true); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ path, want string }{{first, "shared-new"}, {second, "local-new"}} {
		got, err := os.ReadFile(item.path)
		if err != nil || string(got) != item.want {
			t.Fatalf("force did not replace %s: %q, %v", item.path, got, err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("staging/backup files leaked: %v", entries)
	}
}

// Staging the nested first target creates a directory at the second target.
// This reaches cleanup after staging, rather than failing initial validation.
func TestWriteConfigPairCleansStagedFilesAfterTargetBecomesDirectory(t *testing.T) {
	dir := t.TempDir()
	second := filepath.Join(dir, "local.yaml")
	first := filepath.Join(second, "shared.yaml")
	if err := writeConfigPair(first, []byte("shared"), second, []byte("local"), true); err == nil {
		t.Fatal("expected the second target directory to reject installation")
	}
	if _, err := os.Stat(first); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial first configuration remains: %v", err)
	}
	for _, path := range []string{dir, second} {
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				t.Fatalf("staged or backup file leaked in %s: %s", path, entry.Name())
			}
		}
	}
}

func TestCaptureConfigUpdatePreviewAndApplyPreserveLocalSettings(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".agenticlens-config.yaml")
	localPath := filepath.Join(root, ".agenticlens-config.local.yaml")
	shared := []byte("# Keep project policy\nproject_name: Example\ncustom_policy: retained\n")
	local := []byte("telemetry:\n  capture:\n    state_dir: private-local-state\n")
	if err := os.WriteFile(path, shared, 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, local, 0600); err != nil {
		t.Fatal(err)
	}
	if err := runCaptureConfigUpdate(path, "", true); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil || string(original) != string(shared) {
		t.Fatalf("dry run wrote config: %v", err)
	}
	if err := runCaptureConfigUpdate(path, "", false); err != nil {
		t.Fatal(err)
	}
	changed, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(changed), "state_dir: ~/.agenticlens/state") || !strings.Contains(string(changed), "custom_policy: retained") {
		t.Fatalf("unexpected update %s %v", changed, err)
	}
	currentLocal, err := os.ReadFile(localPath)
	if err != nil || string(currentLocal) != string(local) {
		t.Fatalf("local config changed: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 2 {
		t.Fatalf("update leaked artifacts: %v %v", entries, err)
	}
}
