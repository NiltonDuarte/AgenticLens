package capture_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestReconcileArchivedBoundCodexChildPreservesParent(t *testing.T) {
	for _, sourceParent := range []string{"parent", "conflicting-parent"} {
		t.Run(sourceParent, func(t *testing.T) {
			ctx := context.Background()
			s := openStore(t, t.TempDir())
			home, active, archive := archivedRolloutFixture(t)
			parent := filepath.Join(active, "parent.jsonl")
			child := filepath.Join(active, "child.jsonl")
			writeRollout(t, parent, "parent", "", 13)
			writeRollout(t, child, "child", "parent", 17)
			registerBound(t, s, home, parent, "parent", "TEST-100")
			if err := s.QueueChild(ctx, home, child, "parent"); err != nil {
				t.Fatal(err)
			}
			before := reconcileArchiveFixture(t, s, fixtureByteBudget)
			owner := archivedFixtureOwner(t, before, "child")
			if owner.Binding != "source_parent" || owner.ParentID != "parent" || owner.State != domain.CaptureCurrent {
				t.Fatalf("fixture did not verify the inherited child: %+v", owner)
			}

			writeRollout(t, child, "child", sourceParent, 29)
			archivedChild := filepath.Join(archive, filepath.Base(child))
			if err := os.Rename(child, archivedChild); err != nil {
				t.Fatal(err)
			}
			after := reconcileArchiveFixture(t, s, fixtureByteBudget)
			recovered := archivedFixtureOwner(t, after, "child")
			if recovered.ParentID != owner.ParentID || recovered.TicketID != owner.TicketID || recovered.Binding != owner.Binding {
				t.Fatalf("archive refresh changed inherited attribution: %+v", recovered)
			}
			if sourceParent != owner.ParentID {
				if recovered.State != domain.CaptureFailed || recovered.TranscriptPath != child ||
					!strings.Contains(recovered.Detail, "accepted parent identity") ||
					!reflect.DeepEqual(owner.Snapshot, recovered.Snapshot) {
					t.Fatalf("archive refresh accepted conflicting ancestry: %+v", recovered)
				}
				return
			}

			if recovered.TranscriptPath != archivedChild || recovered.State != domain.CaptureCurrent {
				t.Fatalf("archive refresh did not repair the source path: %+v", recovered)
			}
			if len(recovered.Snapshot) != 1 || len(recovered.Snapshot[0].Invocations) != 1 ||
				recovered.Snapshot[0].Invocations[0].Tokens.Input != 29 || recovered.Snapshot[0].ParentID == nil ||
				*recovered.Snapshot[0].ParentID != "parent" {
				t.Fatalf("archive refresh did not replace the exact child facts: %+v", recovered.Snapshot)
			}
			if repeated := reconcileArchiveFixture(t, s, fixtureByteBudget); !reflect.DeepEqual(after.Sessions, repeated.Sessions) {
				t.Fatal("repeating the archived child refresh changed accepted facts")
			}
		})
	}
}

func TestReconcileArchivedCodexFailuresPreserveOwnerAndRefreshSibling(t *testing.T) {
	for _, mode := range []struct {
		name, detail string
	}{
		{"missing", "codex native session unavailable"},
		{"ambiguous", "ambiguous rollout paths"},
		{"foreign-store", "does not belong to selected source store"},
		{"partial-rollout", "partial codex source parse"},
		{"conflicting-active", "does not match its registered source path"},
		{"unreadable-active", "excessive symlink traversal"},
	} {
		t.Run(mode.name, func(t *testing.T) {
			s := openStore(t, t.TempDir())
			home, active, archive := archivedRolloutFixture(t)
			path := filepath.Join(active, "owner.jsonl")
			sibling := filepath.Join(active, "sibling.jsonl")
			writeRollout(t, path, "owner", "", 17)
			writeRollout(t, sibling, "sibling", "", 13)
			registerBound(t, s, home, path, "owner", "TEST-100")
			registerBound(t, s, home, sibling, "sibling", "TEST-100")
			before := reconcileArchiveFixture(t, s, fixtureByteBudget)
			old := archivedFixtureOwner(t, before, "owner")
			if old.State != domain.CaptureCurrent || len(old.Snapshot) != 1 {
				t.Fatalf("fixture owner was never accepted: %+v", old)
			}

			if mode.name != "conflicting-active" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			replacement := filepath.Join(archive, "owner.jsonl")
			switch mode.name {
			case "missing":
				// A similarly named archive file with the wrong native ID is
				// not the registered owner.
				writeRollout(t, replacement, "other", "", 29)
			case "ambiguous":
				writeRollout(t, replacement, "owner", "", 29)
				writeRollout(t, filepath.Join(archive, "copy.jsonl"), "owner", "", 37)
			case "foreign-store":
				outside := filepath.Join(t.TempDir(), "owner.jsonl")
				writeRollout(t, outside, "owner", "", 29)
				if err := os.Symlink(outside, replacement); err != nil {
					t.Fatal(err)
				}
			case "partial-rollout":
				body := writeRollout(t, replacement, "owner", "", 29)
				if err := os.WriteFile(replacement, append(body, []byte("{malformed\n")...), 0600); err != nil {
					t.Fatal(err)
				}
			case "conflicting-active":
				writeRollout(t, path, "other", "", 37)
				writeRollout(t, replacement, "owner", "", 29)
			case "unreadable-active":
				// A symlink loop fails even as root. The failure is distinct
				// from a missing cached source, despite a valid archived copy.
				if err := os.Symlink(filepath.Base(path), path); err != nil {
					t.Fatal(err)
				}
				writeRollout(t, replacement, "owner", "", 29)
			}
			writeRollout(t, sibling, "sibling", "", 47)
			after := reconcileArchiveFixture(t, s, fixtureByteBudget)
			retained := archivedFixtureOwner(t, after, "owner")
			if retained.State != domain.CaptureFailed || !strings.Contains(strings.ToLower(retained.Detail), mode.detail) ||
				retained.TranscriptPath != path || !reflect.DeepEqual(old.Snapshot, retained.Snapshot) ||
				retained.RefreshedAt != old.RefreshedAt {
				t.Fatalf("failure did not report the rejected source and preserve accepted facts: %+v", retained)
			}
			if len(after.RefreshedSessions) != 1 || after.RefreshedSessions[0].ID != "sibling" ||
				len(after.RefreshedSessions[0].Invocations) != 1 || after.RefreshedSessions[0].Invocations[0].Tokens.Input != 47 {
				t.Fatalf("an independent sibling could not refresh: %+v", after.RefreshedSessions)
			}
		})
	}
}

func TestReconcileArchivedCodexBudgetRetriesWithoutLosingFacts(t *testing.T) {
	s := openStore(t, t.TempDir())
	home, active, archive := archivedRolloutFixture(t)
	path := filepath.Join(active, "owner.jsonl")
	writeRollout(t, path, "owner", "", 17)
	registerBound(t, s, home, path, "owner", "TEST-100")
	before := archivedFixtureOwner(t, reconcileArchiveFixture(t, s, fixtureByteBudget), "owner")
	grown := writeRollout(t, path, "owner", "", 29)
	replacement := filepath.Join(archive, filepath.Base(path))
	if err := os.Rename(path, replacement); err != nil {
		t.Fatal(err)
	}
	// A full discovery fits, but discovery and extraction cannot both fit.
	budget := 2*int64(len(grown)) - 1
	limited := reconcileArchiveFixture(t, s, budget)
	pending := archivedFixtureOwner(t, limited, "owner")
	if pending.State != domain.CapturePending || !strings.Contains(strings.ToLower(pending.Detail), "budget") ||
		pending.TranscriptPath != path || !reflect.DeepEqual(pending.Snapshot, before.Snapshot) ||
		len(limited.RefreshedSessions) != 0 || limited.SourceBytes > budget {
		t.Fatalf("budget-limited archive read was accepted or consumed unbounded work: %+v", limited)
	}
	result := reconcileArchiveFixture(t, s, fixtureByteBudget)
	recovered := archivedFixtureOwner(t, result, "owner")
	if recovered.State != domain.CaptureCurrent || recovered.TranscriptPath != replacement ||
		len(result.RefreshedSessions) != 1 || len(result.RefreshedSessions[0].Invocations) != 1 ||
		result.RefreshedSessions[0].Invocations[0].Tokens.Input != 29 {
		t.Fatalf("archive read was not retried after budget became available: %+v", result)
	}
}

func archivedRolloutFixture(t *testing.T) (home, active, archive string) {
	t.Helper()
	var err error
	home, err = filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	active = filepath.Join(home, "sessions")
	archive = filepath.Join(home, "archived_sessions")
	for _, dir := range []string{active, archive} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	return
}

func reconcileArchiveFixture(t *testing.T, s *capture.Store, budget int64) capture.ReconcileResult {
	t.Helper()
	result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 20, budget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func archivedFixtureOwner(t *testing.T, result capture.ReconcileResult, id string) domain.CaptureRegistration {
	t.Helper()
	for _, row := range result.Registrations {
		if row.Identity.SessionID == id {
			return row
		}
	}
	t.Fatalf("owner %s not present in reconciliation", id)
	return domain.CaptureRegistration{}
}
