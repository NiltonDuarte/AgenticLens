package config

import (
	"agenticlens/internal/domain"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCaptureConfigUpdateInsertionPreservationAndCAS(t *testing.T) {
	for _, ext := range []string{"yaml", "yml"} {
		t.Run(ext, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "config."+ext)
			body := "# Preserve comment\nproject_name: Example\nagenticlens:\n  version: v9.8.7\ntelemetry:\n  capture:\n    enabled: false\n    max_sessions: 7\ncustom_policy: retained\n"
			if err := os.WriteFile(path, []byte(body), 0640); err != nil {
				t.Fatal(err)
			}
			preview, err := PrepareCaptureConfigUpdate(path, "chosen-state")
			if err != nil || !preview.Changed || preview.StateDir != "chosen-state" {
				t.Fatalf("preview %+v %v", preview, err)
			}
			current, err := os.ReadFile(path)
			if err != nil || string(current) != body {
				t.Fatalf("preview wrote config: %v", err)
			}
			if err := WriteCaptureConfigUpdate(path, preview); err != nil {
				t.Fatal(err)
			}
			result, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(result, []byte("retained")) || !bytes.Contains(result, []byte("v9.8.7")) {
				t.Fatalf("unrelated fields lost: %s", result)
			}
			if ext == "yaml" && !bytes.Contains(result, []byte("# Preserve comment")) {
				t.Fatal("YAML comment lost")
			}
			cfg, err := decodeCapturePolicy(result)
			if err != nil || cfg.StateDir != "chosen-state" || cfg.Enabled || cfg.MaxSessions != 7 {
				t.Fatalf("capture policy changed: %+v %v", cfg, err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS != "windows" && info.Mode().Perm() != 0640 {
				t.Fatalf("permissions changed to %o", info.Mode().Perm())
			}
			noChange, err := PrepareCaptureConfigUpdate(path, "")
			if err != nil || noChange.Changed || !bytes.Equal(result, noChange.Data) {
				t.Fatalf("existing path rewritten: %+v %v", noChange, err)
			}
			explicit, err := PrepareCaptureConfigUpdate(path, "external-state")
			if err != nil || !explicit.Changed {
				t.Fatalf("explicit replacement %+v %v", explicit, err)
			}
			newer := append(append([]byte(nil), result...), []byte(" ")...)
			if err := os.WriteFile(path, newer, 0640); err != nil {
				t.Fatal(err)
			}
			if err := WriteCaptureConfigUpdate(path, explicit); err == nil || !strings.Contains(err.Error(), "changed after preview") {
				t.Fatalf("stale preview accepted: %v", err)
			}
			current, err = os.ReadFile(path)
			if err != nil || !bytes.Equal(current, newer) {
				t.Fatalf("CAS failure altered file: %v", err)
			}
			files, err := os.ReadDir(root)
			if err != nil || len(files) != 1 {
				t.Fatalf("update leaked temporary files: %v %v", files, err)
			}
		})
	}
}

func TestCaptureUpdateRejectsInvalidValuesAndPreservesExplicitPolicy(t *testing.T) {
	for _, value := range []string{`""`, `"   "`, "null", "42", "false", "{}", "[]"} {
		t.Run(value, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			data := []byte("telemetry:\n  capture:\n    state_dir: " + value + "\n")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := PrepareCaptureConfigUpdate(path, ""); err == nil {
				t.Fatal("invalid setting defaulted")
			}
			actual, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(actual, data) {
				t.Fatalf("failed preview mutated configuration: %v", err)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	original := []byte("telemetry:\n  capture:\n    state_dir: .agenticlens-capture\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	result, err := PrepareCaptureConfigUpdate(path, "")
	if err != nil || result.Changed || !bytes.Equal(result.Data, original) {
		t.Fatalf("explicit path replaced: %+v %v", result, err)
	}
	cfg := fixtureProjectPolicy(t)
	cfg.ProjectName = "Example"
	cfg.Ticket.Prefixes = []domain.TicketPrefix{{Prefix: "AL", Description: "Fixture work"}}
	cfg.CodeMetrics.Production.Include = []string{"src/**"}
	cfg.Telemetry.Capture.StateDir = ""
	cfg.Telemetry.Capture.Enabled = false
	cfg.Telemetry.Capture.MaxSessions = 7
	generated, err := GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(generated), "capture:") {
		t.Fatal("shared writer leaked machine capture policy")
	}
	if _, err := PrepareCaptureConfigUpdate(path+"-absent", ""); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("created missing config: %v", err)
	}
}

func TestCaptureConfigUpdatePreservesPinnedOperationalConfiguration(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, ".agenticlens-config.yaml")
	body := fixtureSharedBody(t)
	if err := os.WriteFile(shared, body, 0600); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(root, ".agenticlens-config.local.yaml")
	before := []byte("telemetry:\n  capture:\n    enabled: false\n    state_dir: before-state\n")
	if err := os.WriteFile(local, before, 0600); err != nil {
		t.Fatal(err)
	}
	preview, err := PrepareCaptureConfigUpdate(local, "fixture-state")
	if err != nil || !preview.Changed {
		t.Fatalf("preview %+v %v", preview, err)
	}
	if err := WriteCaptureConfigUpdateChecked(context.Background(), local, preview, func(CaptureConfigChange) error { return nil }); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(root)
	if err != nil || cfg.AgenticLens.Version != "v9.8.7" || cfg.ProjectName != "Fixture" || cfg.Telemetry.Capture.StateDir != "fixture-state" {
		t.Fatalf("operational configuration changed: %+v %v", cfg, err)
	}
	entry, err := ResolveEntrypoint(root)
	if err != nil || entry.Version != "v9.8.7" {
		t.Fatalf("pin %+v %v", entry, err)
	}
	after, err := os.ReadFile(shared)
	if err != nil || !bytes.Equal(after, body) {
		t.Fatal("local update modified shared policy")
	}
}

// The parent test invokes this subprocess with an actual production update and
// kills it while the OS config lock is held at a precise activation boundary.
func TestCaptureConfigUpdateCrashChild(t *testing.T) {
	path := os.Getenv("AL68_CONFIG_UPDATE_TEST_PATH")
	if path == "" {
		return
	}
	stage := configUpdateStaged
	if os.Getenv("AL68_CONFIG_UPDATE_TEST_STAGE") == "activated" {
		stage = configUpdateActivated
	}
	update, err := PrepareCaptureConfigUpdate(path, "after-crash-state")
	if err != nil {
		t.Fatal(err)
	}
	err = writeCaptureConfigUpdate(path, update, func(current configUpdateStage) error {
		if current != stage {
			return nil
		}
		if err := os.WriteFile(path+".checkpoint", []byte("ready"), 0600); err != nil {
			return err
		}
		var signal [1]byte
		_, err := os.Stdin.Read(signal[:])
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal("subprocess was expected to remain at its checkpoint until killed")
}

func TestCaptureConfigUpdateSurvivesProcessDeath(t *testing.T) {
	for _, stage := range []string{"staged", "activated"} {
		t.Run(stage, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, ".agenticlens-config.local.yaml")
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), fixtureSharedBody(t), 0600); err != nil {
				t.Fatal(err)
			}
			original := []byte("# Preserve local comment\ntelemetry:\n  capture:\n    enabled: false\n    state_dir: before-crash-state\n")
			if err := os.WriteFile(path, original, 0640); err != nil {
				t.Fatal(err)
			}
			prepared, err := PrepareCaptureConfigUpdate(path, "after-crash-state")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCaptureConfigUpdateCrashChild$")
			cmd.Env = append(os.Environ(), "AL68_CONFIG_UPDATE_TEST_PATH="+path, "AL68_CONFIG_UPDATE_TEST_STAGE="+stage)
			var output bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &output
			stdin, err := cmd.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := stdin.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
					t.Error(err)
				}
			}()
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			exited := make(chan error, 1)
			go func() { exited <- cmd.Wait() }()
			deadline := time.NewTimer(10 * time.Second)
			defer deadline.Stop()
			tick := time.NewTicker(10 * time.Millisecond)
			defer tick.Stop()
		waiting:
			for {
				select {
				case err := <-exited:
					t.Fatalf("child exited before checkpoint: %v %s", err, output.String())
				case <-deadline.C:
					if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
						t.Error(err)
					}
					<-exited
					t.Fatal("subprocess failed to reach configuration checkpoint")
				case <-tick.C:
					if _, err := os.Stat(path + ".checkpoint"); err == nil {
						break waiting
					} else if !errors.Is(err, os.ErrNotExist) {
						t.Fatal(err)
					}
				}
			}
			// Readers must always find a complete configuration while the writer holds
			// its OS config lock, including immediately after atomic replacement on Windows.
			cfg, err := Load(root)
			if err != nil {
				t.Fatalf("config disappeared at %s: %v", stage, err)
			}
			want := "before-crash-state"
			if stage == "activated" {
				want = "after-crash-state"
			}
			if cfg.Telemetry.Capture.StateDir != want {
				t.Fatalf("wrong boundary config: %+v", cfg.Telemetry.Capture)
			}
			if _, err := ResolveEntrypoint(root); err != nil {
				t.Fatalf("pin became unusable at %s: %v", stage, err)
			}
			if stage == "staged" {
				contendingPath := path
				if runtime.GOOS == "windows" {
					contendingPath = strings.ToUpper(path)
				}
				if err := WriteCaptureConfigUpdate(contendingPath, prepared); err == nil {
					t.Fatal("concurrent updater bypassed the held configuration lock")
				}
			}
			if err := cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err := <-exited; err == nil {
				t.Fatal("child unexpectedly completed instead of being killed")
			}
			current, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			expected := original
			if stage == "activated" {
				expected = prepared.Data
			}
			if !bytes.Equal(current, expected) {
				t.Fatalf("process death left a partial config at %s", stage)
			}
			retry, err := PrepareCaptureConfigUpdate(path, "resumed-state")
			if err != nil {
				t.Fatal(err)
			}
			if err := WriteCaptureConfigUpdate(path, retry); err != nil {
				t.Fatalf("OS lock survived child death: %v", err)
			}
			cfg, err = Load(root)
			if err != nil || cfg.Telemetry.Capture.StateDir != "resumed-state" {
				t.Fatalf("retry failed: %+v %v", cfg, err)
			}
			if _, err := os.Stat(path + ".update.lock"); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("persistent sidecar lock exists: %v", err)
			}
		})
	}
}

func TestCaptureConfigUpdatePreservesExternalReplacementOnError(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "config.yaml")
	original := []byte("telemetry:\n  capture:\n    state_dir: before\n")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	update, err := PrepareCaptureConfigUpdate(path, "planned")
	if err != nil {
		t.Fatal(err)
	}
	external := []byte("telemetry:\n  capture:\n    state_dir: externally-changed\n")
	err = writeCaptureConfigUpdate(path, update, func(stage configUpdateStage) error {
		if stage != configUpdateStaged {
			return nil
		}
		return os.WriteFile(path, external, 0600)
	})
	if err == nil || !strings.Contains(err.Error(), "changed after preview") {
		t.Fatalf("external replacement was not rejected: %v", err)
	}
	actual, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(actual, external) {
		t.Fatalf("rollback overwrote newer external configuration: %v", err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("normal rejection leaked artifacts: %v %v", entries, err)
	}
}
