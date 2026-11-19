package main_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCaptureArchivedCodexRolloutRefresh(t *testing.T) {
	for _, first := range []string{"reconcile", "generate"} {
		t.Run(first, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			serializeCaptureFixtureWorkers(t, f)
			f.bind(t, f.write(t, 17))
			assertCaptureRootInput(t, f.generate(t), 17)
			registry, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := registry.Close(); err != nil {
					t.Error(err)
				}
			}()
			readOwner := func() domain.CaptureRegistration {
				t.Helper()
				rows, err := registry.List(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) != 1 {
					t.Fatalf("expected one registered owner, got %d", len(rows))
				}
				return rows[0]
			}
			prior := readOwner()
			if prior.TicketID != "TEST-100" || prior.Binding == "" {
				t.Fatalf("fixture owner lacks a binding: %+v", prior)
			}

			active := f.writeAt(t, 29, "2026-01-01T00:00:03Z")
			archive := filepath.Join(f.store, "archived_sessions", filepath.Base(active))
			if err := os.Mkdir(filepath.Dir(archive), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(active, archive); err != nil {
				t.Fatal(err)
			}
			archive, err = filepath.EvalSymlinks(archive)
			if err != nil {
				t.Fatal(err)
			}

			if first == "reconcile" {
				out, err := runCaptureNative(t, f.project, "", "capture", "reconcile", "--repo", f.project, "--source", "codex", "--store", f.store, "--ticket-id", "TEST-100", "--state-dir", f.state)
				if err != nil {
					t.Fatalf("capture reconcile: %v %s", err, out)
				}
				var result capture.ReconcileResult
				if err := json.Unmarshal(out, &result); err != nil {
					t.Fatalf("decode reconcile result: %v %s", err, out)
				}
				if len(result.Registrations) != 1 || result.Registrations[0].State != domain.CaptureCurrent {
					t.Fatalf("reconcile did not refresh archived owner: %s", out)
				}
				// The native reconcile must repair the registry before generate
				// can recover the path independently.
				if owner := readOwner(); owner.TranscriptPath != archive {
					t.Fatalf("reconcile saved path %q, want %q", owner.TranscriptPath, archive)
				}
				assertCaptureRootInput(t, f.ledger(t), 17)
			}

			published := f.generate(t)
			assertCaptureRootInput(t, published, 29)
			if len(published.Sessions) != 1 {
				t.Fatalf("archive refresh changed the number of published sessions: %d", len(published.Sessions))
			}
			owner := readOwner()
			if owner.TranscriptPath != archive || owner.State != domain.CaptureCurrent ||
				owner.Identity != prior.Identity || owner.TicketID != prior.TicketID ||
				owner.Binding != prior.Binding || owner.Request != prior.Request || owner.ParentID != prior.ParentID {
				t.Fatalf("archived refresh changed attribution or failed to repair the path: %+v", owner)
			}
			if rerun := f.generate(t); !reflect.DeepEqual(published.Sessions, rerun.Sessions) {
				t.Fatalf("repeated archive refresh changed accepted sessions: got %+v, want %+v", rerun.Sessions, published.Sessions)
			}
		})
	}
}
