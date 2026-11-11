//go:build !windows

package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
)

// The Unix engine/native-manager fixture is the same imperative boundary used
// by installation_unix_test.go. This is a real filesystem-prefix interruption,
// not a domain mock or a claim of an actual agent lifecycle exchange.
func TestRelocationResumesInterruptedOwnedHookAndPluginBeforeManifest(t *testing.T) {
	for _, sourceKind := range []string{"codex", "opencode"} {
		t.Run(sourceKind, func(t *testing.T) {
			ctx := context.Background()
			project, source, base, legacyDB := relocationFixture(t)
			t.Setenv("AL_CAPTURE_INSTALL_FIXTURE", "1")
			t.Setenv("GORACE", "atexit_sleep_ms=0")
			cache := filepath.Join(t.TempDir(), "native-cache")
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			if err := os.WriteFile(local, []byte(fmt.Sprintf("agenticlens:\n  install_root: %q\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n", cache, source)), 0600); err != nil {
				t.Fatal(err)
			}
			entry, err := adaptersconfig.ResolveEntrypoint(project)
			if err != nil {
				t.Fatal(err)
			}
			binary, err := entry.NativePath()
			if err != nil {
				t.Fatal(err)
			}
			if err = os.MkdirAll(filepath.Dir(binary), 0700); err != nil {
				t.Fatal(err)
			}
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Symlink(executable, binary); err != nil {
				t.Fatal(err)
			}
			wrapperBody, err := os.ReadFile(relocationManagerExecutable(t))
			if err != nil {
				t.Fatal(err)
			}
			wrapper := filepath.Join(t.TempDir(), "wrapper")
			if err = os.WriteFile(wrapper, wrapperBody, 0700); err != nil {
				t.Fatal(err)
			}
			storePath := t.TempDir()
			if sourceKind == "opencode" {
				storePath = filepath.Join(storePath, "source.db")
			}
			installed, err := Install(ctx, InstallOptions{Repository: project, StateDir: source, Source: sourceKind, SourceStore: storePath, Binary: binary, Wrapper: wrapper})
			if err != nil {
				t.Fatal(err)
			}
			legacy := installed.Manifest
			legacy.Protocol = 1
			legacyBody, err := json.Marshal(legacy)
			if err != nil {
				t.Fatal(err)
			}
			anchor := filepath.Join(project, "."+sourceKind, "agenticlens-installation.json")
			for _, path := range []string{manifestPath(source, sourceKind), anchor} {
				if err = os.WriteFile(path, legacyBody, 0600); err != nil {
					t.Fatal(err)
				}
			}
			beforeHook, err := os.ReadFile(legacy.ConfigPath)
			if err != nil {
				t.Fatal(err)
			}
			artifacts := map[string][]byte{
				filepath.Join(project, "workspaces", "T-1", "agentic_metrics.json"): []byte("{\"schema_version\":5,\"ticket_id\":\"T-1\",\"sessions\":[]}\n"),
				filepath.Join(project, "workspaces", "T-1", "SUMMARY.md"):           []byte("Previously committed summary\n"),
			}
			for path, body := range artifacts {
				if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			prefixApplied, interrupted := false, false
			func() {
				defer func() {
					if recover() != nil {
						interrupted = true
					}
				}()
				_, err = relocateState(ctx, RelocationOptions{Project: project, FromDirectory: source, ToBase: base}, func(ctx context.Context, options InstallOptions) (InstallResult, error) {
					return installWithWriter(ctx, options, func(change fileChange) error {
						if writeErr := apply(change); writeErr != nil {
							return writeErr
						}
						if change.path == legacy.ConfigPath {
							prefixApplied = true
							panic("test interruption after the owned hook/plugin replacement")
						}
						return nil
					})
				})
			}()
			if !prefixApplied || !interrupted {
				t.Fatalf("did not interrupt the real integration write prefix: %v", err)
			}
			afterHook, err := os.ReadFile(legacy.ConfigPath)
			if err != nil || bytes.Equal(afterHook, beforeHook) {
				t.Fatalf("owned integration prefix was not written: %v", err)
			}
			target, err := ProjectState(base, project)
			if err != nil {
				t.Fatal(err)
			}
			db, e := openReadOnlyDB(filepath.Join(target, "registry.db"))
			if e != nil {
				t.Fatal(e)
			}
			var scope int64
			if e = db.QueryRow("SELECT id FROM worktrees WHERE directory=?", project).Scan(&scope); e != nil {
				t.Fatal(e)
			}
			coord, e := readCoordinatePath(ctx, db, target, scope)
			if closeErr := db.Close(); e != nil || closeErr != nil {
				t.Fatalf("pending coordinate: %v %v", e, closeErr)
			}
			pendingManifest, err := os.ReadFile(manifestPath(coord, sourceKind))
			if err != nil || !bytes.Equal(pendingManifest, legacyBody) {
				t.Fatalf("interruption passed the manifest checkpoint: %v", err)
			}
			unchangedAnchor, err := os.ReadFile(anchor)
			if err != nil || !bytes.Equal(unchangedAnchor, legacyBody) {
				t.Fatalf("interruption passed the anchor checkpoint: %v", err)
			}
			if _, err = os.Stat(installJournalPath(coord, sourceKind)); err != nil {
				t.Fatal("prefix did not retain its installer journal")
			}
			if pending, openErr := OpenProject(target, project); openErr == nil {
				pending.Close()
				t.Fatal("interrupted relocation became operational")
			}
			var protocol int
			if err = legacyDB.QueryRow("SELECT version FROM capture_meta").Scan(&protocol); err != nil || protocol != -1 {
				t.Fatal("interrupted migration did not preserve source fence")
			}

			// The ordinary public maintenance entrypoint must recover its own
			// exact output prefix, then finish without accepting foreign edits.
			result, err := RelocateState(ctx, RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
			if err != nil || result.Phase != "completed" {
				t.Fatalf("normal public retry could not recover owned prefix: %+v %v", result, err)
			}
			current, err := OpenProject(target, project)
			if err != nil {
				t.Fatal(err)
			}
			defer current.Close()
			rows, err := current.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			assertRelocationRegistrations(t, rows, project)
			assertRelocationContinuations(t, current, project)
			request, err := current.PublicationRequest(ctx)
			if err != nil || request.Generation != 7 || request.Acknowledged != 3 {
				t.Fatalf("retry changed publication acknowledgement: %+v %v", request, err)
			}
			final, err := ReadManifest(coord, sourceKind)
			if err != nil || !final.Active || final.Protocol != ManifestProtocolVersion || final.StateDir != coord {
				t.Fatalf("retry did not finish the owned installation: %+v %v", final, err)
			}
			if err = checkOwnedIntegration(project, sourceKind, final); err != nil {
				t.Fatalf("retry does not own the final integration bytes: %v", err)
			}
			finalHook, err := os.ReadFile(legacy.ConfigPath)
			if err != nil || !strings.Contains(string(finalHook), target) || strings.Contains(string(finalHook), source) {
				t.Fatalf("retry retained an obsolete source location: %v", err)
			}
			for path, body := range artifacts {
				actual, readErr := os.ReadFile(path)
				if readErr != nil || !bytes.Equal(actual, body) {
					t.Fatalf("migration rewrote prior publication artifact %s: %v", path, readErr)
				}
			}
		})
	}
}

func relocationManagerExecutable(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "agentic-lens")
	command := exec.Command("go", "build", "-ldflags", "-X agenticlens/internal/domain.Version=v9.8.7", "-o", path, "./cmd/agentic-lens-manager")
	command.Dir = filepath.Join("..", "..", "..")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build manager: %v: %s", err, output)
	}
	return path
}
