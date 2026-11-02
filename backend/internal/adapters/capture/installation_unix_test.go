//go:build !windows

package capture_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestInstalledStateDriftAliasRemovalAndDryRun(t *testing.T) {
	project, state := t.TempDir(), filepath.Join(t.TempDir(), "state with ' quote")
	binary := filepath.Join(t.TempDir(), "native")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '1\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Source: "codex", SourceStore: t.TempDir()}
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(project, ".codex", "agenticlens-installation.json")); err != nil {
		t.Fatal(err)
	}
	states, err := capture.InstalledStates(project)
	if err != nil || len(states) != 1 || states[0] != state {
		t.Fatalf("discovery %v %v", states, err)
	}
	if err := capture.CheckInstalledState(project, t.TempDir(), false); err == nil {
		t.Fatal("state drift hidden")
	}
	if err := capture.CheckInstalledState(project, t.TempDir(), true); err != nil {
		t.Fatal("intentional override rejected", err)
	}
	alias := filepath.Join(t.TempDir(), "project")
	if err := os.Symlink(project, alias); err != nil {
		t.Fatal(err)
	}
	options.Repository = alias
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal("alias reinstall", err)
	}
	options.Remove, options.DryRun = true, true
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if enabled, err := capture.AutomaticEnabled(state); err != nil || !enabled {
		t.Fatal("dry-run deactivated", err)
	}
	options.DryRun = false
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal("alias removal", err)
	}
	if enabled, err := capture.AutomaticEnabled(state); err != nil || enabled {
		t.Fatal("removal remained active", err)
	}
	options.Repository = t.TempDir()
	if _, err := capture.Install(context.Background(), options); err == nil {
		t.Fatal("foreign project removal accepted")
	}
}

func TestLegacyOpenCodeInstallationDiscovery(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary := filepath.Join(t.TempDir(), "native")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '1\\n'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Source: "opencode", SourceStore: filepath.Join(t.TempDir(), "source.db")}
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(project, ".opencode", "agenticlens-installation.json")); err != nil {
		t.Fatal(err)
	}
	states, err := capture.InstalledStates(project)
	if err != nil || len(states) != 1 || states[0] != state {
		t.Fatalf("legacy plugin discovery %v %v", states, err)
	}
}
