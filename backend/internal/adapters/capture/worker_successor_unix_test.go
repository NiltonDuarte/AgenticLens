//go:build linux || darwin

package capture_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

func TestWorkerSuccessorRequiresExactReadyInstallationAndOriginalState(t *testing.T) {
	for _, scenario := range []string{"ready", "state drift", "manifest version", "manifest namespace", "missing native", "disabled integration"} {
		t.Run(scenario, func(t *testing.T) {
			project := t.TempDir()
			binary, wrapper := installFixture(t, project)
			state, err := capture.ProjectState(filepath.Join(project, capture.DefaultDirectory), project)
			if err != nil {
				t.Fatal(err)
			}
			registry, err := capture.OpenProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			state = registry.CoordinationDirectory()
			if err = registry.Close(); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err = capture.Install(ctx, capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "codex", SourceStore: t.TempDir()}); err != nil {
				t.Fatal(err)
			}
			manifest, err := capture.ReadManifest(state, "codex")
			if err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "state drift":
				if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    enabled: true\n    state_dir: moved-state\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "manifest version":
				manifest.Version = "v9.8.8"
			case "manifest namespace":
				manifest.StateDir = t.TempDir()
			case "missing native":
				if err = os.Remove(binary); err != nil {
					t.Fatal(err)
				}
			case "disabled integration":
				manifest.Active = false
			}
			if scenario == "manifest version" || scenario == "manifest namespace" || scenario == "disabled integration" {
				body, err := json.Marshal(manifest)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(state, "install-codex.json"), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			selected, err := capture.WorkerSuccessor(ctx, project, state)
			if scenario == "ready" {
				if err != nil || selected != binary {
					t.Fatalf("ready exact successor: %s %v", selected, err)
				}
			} else if err == nil || selected != "" {
				t.Fatalf("unsafe successor accepted: %s %v", selected, err)
			}
			if scenario == "state drift" && !strings.Contains(err.Error(), "original namespace") {
				t.Fatalf("state recovery guidance absent: %v", err)
			}
			if _, err = os.Stat(filepath.Join(state, "registry.db")); !os.IsNotExist(err) {
				t.Fatalf("successor validation created/migrated registry: %v", err)
			}
		})
	}
}
