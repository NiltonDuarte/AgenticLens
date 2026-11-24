package capture_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestBindArchivedUnboundCodex(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "exact project"},
		{name: "foreign project", want: "conflicts with configured project"},
		{name: "ambiguous owner", want: "ambiguous rollout paths"},
		{name: "changed parent", want: "conflicts with the accepted parent identity"},
		{name: "archive outside store", want: "escapes its registered store"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			captureProjectConfig(t, project, "ArchivedBindingFixture")
			s, err := capture.OpenProject(t.TempDir(), project)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			home, active, archive := archivedRolloutFixture(t)
			path := filepath.Join(active, "root.jsonl")
			archived := filepath.Join(archive, filepath.Base(path))
			writeSource := func(path, cwd, parent string) {
				t.Helper()
				body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"root","cwd":%q,"parent_thread_id":%q,"timestamp":"2026-01-01T00:00:00Z"}}`+"\n"+
					`{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn"}}`+"\n"+
					`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"turn","usage":{"input_tokens":17}}}`+"\n", cwd, parent)
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			writeSource(path, project, "")
			if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path}, home, "/native", s.CoordinationDirectory()); err != nil {
				t.Fatal(err)
			}
			observed, err := s.Reconcile(ctx, "", "tester", 16, fixtureByteBudget, fixtureGenerationPolicy())
			if err != nil {
				t.Fatal(err)
			}
			before := archivedFixtureOwner(t, observed, "root")
			if before.TicketID != "" || before.ProjectProof != domain.CaptureProjectVerified || before.TranscriptPath != path || len(before.Snapshot) != 1 {
				t.Fatal("fixture must have a verified unbound capture at its original path")
			}
			if err := os.Rename(path, archived); err != nil {
				t.Fatal(err)
			}
			switch tc.name {
			case "foreign project":
				writeSource(archived, t.TempDir(), "")
			case "ambiguous owner":
				writeSource(filepath.Join(archive, "copy.jsonl"), project, "")
			case "changed parent":
				writeSource(archived, project, "replacement-parent")
			case "archive outside store":
				external := t.TempDir()
				externalPath := filepath.Join(external, filepath.Base(archived))
				if err := os.Rename(archived, externalPath); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(archive); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(external, archive); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(externalPath, filepath.Join(active, "alias.jsonl")); err != nil {
					t.Fatal(err)
				}
			}
			bound, bindErr := s.Bind(ctx, before.Identity, "TEST-100", before.Request)
			rows, err := s.List(ctx)
			if err != nil || len(rows) != 1 {
				t.Fatalf("expected one persisted registration: got %d, err=%v", len(rows), err)
			}
			if tc.want != "" {
				if !errors.Is(bindErr, capture.ErrBindingProofPending) || !strings.Contains(bindErr.Error(), tc.want) {
					t.Fatalf("binding rejected without checking the archive: want %q, got %v", tc.want, bindErr)
				}
				if !reflect.DeepEqual(rows[0], before) {
					t.Fatal("rejected archive changed the persisted registration or captured facts")
				}
				return
			}
			if bindErr != nil {
				t.Fatalf("late binding did not recover the archived source: %v", bindErr)
			}
			if bound.Identity != before.Identity || bound.TicketID != "TEST-100" || bound.TranscriptPath != archived ||
				bound.Binding != "explicit_agent_response" || bound.BindingAcceptedAt == "" || !reflect.DeepEqual(rows[0], bound) ||
				!reflect.DeepEqual(before.Snapshot, bound.Snapshot) {
				t.Fatal("binding did not persist the verified archived path and preserve captured facts")
			}
		})
	}
}

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

func TestReconcileArchivedCodexRootRejectsNewParent(t *testing.T) {
	s := openStore(t, t.TempDir())
	home, active, archive := archivedRolloutFixture(t)
	path := filepath.Join(active, "root.jsonl")
	writeRollout(t, path, "root", "", 17)
	registerBound(t, s, home, path, "root", "TEST-100")
	before := archivedFixtureOwner(t, reconcileArchiveFixture(t, s, fixtureByteBudget), "root")
	if before.State != domain.CaptureCurrent || before.ParentID != "" ||
		len(before.Snapshot) != 1 || before.Snapshot[0].ParentID != nil {
		t.Fatalf("fixture did not accept a root: %+v", before)
	}

	writeRollout(t, path, "root", "new-parent", 29)
	if err := os.Rename(path, filepath.Join(archive, filepath.Base(path))); err != nil {
		t.Fatal(err)
	}
	result := reconcileArchiveFixture(t, s, fixtureByteBudget)
	got := archivedFixtureOwner(t, result, "root")
	if got.State != domain.CaptureFailed || !strings.Contains(got.Detail, "accepted parent identity") ||
		got.ParentID != "" || got.TranscriptPath != path || !reflect.DeepEqual(got.Snapshot, before.Snapshot) ||
		len(result.RefreshedSessions) != 0 {
		t.Fatalf("archive changed the accepted root's ancestry or facts: %+v", got)
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

func TestReconcileArchivedCodexLateRegistryChangeExcludesSourceEvidence(t *testing.T) {
	const stableDelegation = `{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"accepted-spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"coder\"}"}}
{"timestamp":"2026-01-01T00:00:04Z","type":"response_item","payload":{"type":"function_call_output","call_id":"accepted-spawn","output":"{\"session_id\":\"sibling\"}"}}
`
	for _, mode := range []string{"accepted", "superseded-before-annotations", "superseded-after-annotations"} {
		t.Run(mode, func(t *testing.T) {
			s := openStore(t, t.TempDir())
			home, active, archive := archivedRolloutFixture(t)
			parentPath := filepath.Join(active, "parent.jsonl")
			childPath := filepath.Join(active, "child.jsonl")
			siblingPath := filepath.Join(active, "sibling.jsonl")
			priorParent := writeRollout(t, parentPath, "parent", "", 13)
			if err := os.WriteFile(parentPath, append(priorParent, []byte(stableDelegation)...), 0600); err != nil {
				t.Fatal(err)
			}
			writeRollout(t, childPath, "child", "parent", 17)
			writeRollout(t, siblingPath, "sibling", "parent", 19)
			registerBound(t, s, home, parentPath, "parent", "TEST-100")
			registerBound(t, s, home, childPath, "child", "TEST-100")
			registerBound(t, s, home, siblingPath, "sibling", "TEST-100")
			prior := reconcileArchiveFixture(t, s, fixtureByteBudget)
			before := archivedFixtureOwner(t, prior, "child")
			if before.State != domain.CaptureCurrent || before.ParentID != "parent" {
				t.Fatalf("fixture did not establish an accepted child checkpoint: %+v", before)
			}
			stableParent := archivedFixtureOwner(t, prior, "parent")
			if len(stableParent.Snapshot) != 1 || len(stableParent.Snapshot[0].Delegations) != 1 ||
				stableParent.Snapshot[0].Delegations[0].ChildSession != "sibling" {
				t.Fatalf("fixture did not accept the independent delegation: %+v", stableParent.Snapshot)
			}

			// The opaque agent ID is promoted to a native child link only
			// when a matching child's source participates in grouping.
			parentBody := writeRollout(t, parentPath, "parent", "", 23)
			parentBody = append(parentBody, []byte(stableDelegation)...)
			parentBody = append(parentBody, []byte(`{"timestamp":"2026-01-01T00:00:05Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}
{"timestamp":"2026-01-01T00:00:06Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":"{\"agent_id\":\"child\"}"}}
`)...)
			if err := os.WriteFile(parentPath, parentBody, 0600); err != nil {
				t.Fatal(err)
			}
			childBody := writeRollout(t, childPath, "child", "parent", 29)
			childBody = append(childBody, []byte(`{"timestamp":"2026-01-01T00:00:05Z","type":"response_item","payload":{"type":"function_call","call_id":"archived-check","name":"run_check","arguments":"{}"}}
{"timestamp":"2026-01-01T00:00:06Z","type":"response_item","payload":{"type":"function_call_output","call_id":"archived-check","output":"completed"}}
`)...)
			if err := os.WriteFile(childPath, childBody, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(childPath, filepath.Join(archive, filepath.Base(childPath))); err != nil {
				t.Fatal(err)
			}
			writeRollout(t, siblingPath, "sibling", "parent", 47)
			if mode != "accepted" {
				supersedeArchiveCheckpoint(t, s, mode == "superseded-after-annotations")
			}

			result := reconcileArchiveFixture(t, s, fixtureByteBudget)
			gotChild := archivedFixtureOwner(t, result, "child")
			gotParent := archivedFixtureOwner(t, result, "parent")
			wantInput, wantLink, wantState := int64(29), "child", domain.CaptureCurrent
			if mode != "accepted" {
				wantInput, wantLink, wantState = 41, "", domain.CapturePending
				if gotChild.SourceFingerprint != "newer-checkpoint" {
					t.Fatalf("fixture did not supersede the archive observation: %+v", gotChild)
				}
				if !strings.Contains(gotChild.Detail, "retry") {
					t.Errorf("superseded archive read had no retry diagnostic: %q", gotChild.Detail)
				}
			}
			if gotChild.State != wantState {
				t.Errorf("child state after archive reconciliation = %s, want %s", gotChild.State, wantState)
			}
			if len(gotChild.Snapshot) != 1 || len(gotChild.Snapshot[0].Invocations) != 1 ||
				gotChild.Snapshot[0].Invocations[0].Tokens.Input != wantInput {
				t.Fatalf("unexpected child snapshot after archive reconciliation: %+v", gotChild)
			}
			if len(gotParent.Snapshot) != 1 {
				t.Fatalf("parent snapshot was not retained: %+v", gotParent.Snapshot)
			}
			if len(gotParent.Snapshot[0].Invocations) != 1 ||
				gotParent.Snapshot[0].Invocations[0].Tokens.Input != 23 || len(gotParent.Snapshot[0].Delegations) != 2 ||
				gotParent.Snapshot[0].Delegations[0].ChildSession != "sibling" || gotParent.Snapshot[0].Delegations[1].ChildSession != wantLink {
				t.Errorf("parent did not refresh independently of stale child evidence: %+v", gotParent.Snapshot)
			}
			wantEvidence := mode == "accepted"
			source := result.SourceEvents
			containsChildEvent := slices.ContainsFunc(source.Events, func(event domain.EventFact) bool { return event.SessionID == "child" })
			containsChildCheckpoint := slices.ContainsFunc(source.CheckpointEvidence, func(checkpoint domain.CheckpointInvocationEvidence) bool {
				return checkpoint.SessionID == "child"
			})
			containsParentLink := slices.ContainsFunc(source.Events, func(event domain.EventFact) bool {
				return event.SessionID == "parent" && event.CallID == "spawn" && event.ChildSessionID == "child"
			})
			containsChildRefresh := slices.ContainsFunc(result.RefreshedSessions, func(row domain.SessionFact) bool { return row.ID == "child" })
			if containsChildEvent != wantEvidence || containsChildCheckpoint != wantEvidence || containsParentLink != wantEvidence ||
				slices.Contains(source.CompleteSessionIDs, "child") != wantEvidence || containsChildRefresh != wantEvidence {
				t.Errorf("unexpected child evidence in publication: events=%v checkpoint=%v parent-link=%v complete=%v refreshed=%v",
					containsChildEvent, containsChildCheckpoint, containsParentLink, source.CompleteSessionIDs, containsChildRefresh)
			}
			if !slices.ContainsFunc(source.Events, func(event domain.EventFact) bool {
				return event.SessionID == "parent" && event.CallID == "accepted-spawn" && event.ChildSessionID == "sibling"
			}) || !slices.Contains(source.CompleteSessionIDs, "sibling") ||
				!slices.ContainsFunc(result.RefreshedSessions, func(row domain.SessionFact) bool {
					return row.ID == "sibling" && len(row.Invocations) == 1 && row.Invocations[0].Tokens.Input == 47
				}) {
				t.Error("independent sibling lost its existing linkage or eligible refresh")
			}
			if !slices.ContainsFunc(result.RefreshedSessions, func(row domain.SessionFact) bool {
				return row.ID == "parent" && reflect.DeepEqual(row, gotParent.Snapshot[0])
			}) {
				t.Error("publication omitted the independently refreshed parent or used a stale projection")
			}
			if !slices.ContainsFunc(result.Sessions, func(row domain.SessionFact) bool {
				return row.ID == "child" && reflect.DeepEqual(row, gotChild.Snapshot[0])
			}) {
				t.Error("publication dropped the accepted child snapshot")
			}
		})
	}
}

func supersedeArchiveCheckpoint(t *testing.T, s *capture.Store, afterAnnotations bool) {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(s.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	// The trigger forces the same registry state as an intervening accepted
	// checkpoint without a producer hook or a goroutine timing dependency.
	pathChanged := "json_extract(OLD.record, '$.path') != json_extract(NEW.record, '$.path')"
	prelude, triggerGuard, updateGuard := "", pathChanged, ""
	if afterAnnotations {
		if _, err := db.Exec(`CREATE TABLE archive_annotation_writes (count INTEGER NOT NULL);
INSERT INTO archive_annotation_writes VALUES (0);`); err != nil {
			t.Fatal(err)
		}
		// Grouping writes the child's delegation annotation first, then its
		// anomaly annotation. Replacing the checkpoint on the second write
		// leaves grouping complete when final publication loads the rows.
		triggerGuard = "NOT (" + pathChanged + ") AND instr(json_extract(NEW.record, '$.path'), '/archived_sessions/') != 0"
		prelude = "UPDATE archive_annotation_writes SET count = count + 1;"
		updateGuard = " AND (SELECT count FROM archive_annotation_writes) = 2"
	}
	_, err = db.Exec(`CREATE TRIGGER supersede_archive_checkpoint AFTER UPDATE ON registrations
WHEN NEW.session_id = 'child' AND (` + triggerGuard + `)
BEGIN
	` + prelude + `
	UPDATE registrations
	SET record = json_set(CAST(NEW.record AS TEXT),
		'$.registration.source_fingerprint', 'newer-checkpoint',
		'$.registration.refreshed_at', '2026-01-02T00:00:00Z',
		'$.registration.revision', json_extract(NEW.record, '$.registration.revision') + 1,
		'$.snapshot[0].invocations[0].tokens.input', 41)
	WHERE worktree_id = NEW.worktree_id AND identity = NEW.identity` + updateGuard + `;
END`)
	if err != nil {
		t.Fatalf("install fixture checkpoint trigger: %v", err)
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

// FIX-21: an unassigned Codex child whose rollout Codex moved to
// archived_sessions/ is recovered on the same exact proof that binding uses,
// or isolated with a typed reason. The row stays unassigned either way.
func TestReconcileArchivedUnassignedCodexChild(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reason capture.IsolationReason
		detail string
	}{
		{name: "recovered"},
		{name: "missing", reason: capture.IsolationUnavailable, detail: "unavailable"},
		{name: "ambiguous", reason: capture.IsolationAmbiguous, detail: "ambiguous rollout paths"},
		{name: "foreign project", reason: capture.IsolationConflicting, detail: "conflicts with configured project"},
		{name: "conflicting parent", reason: capture.IsolationConflicting, detail: "accepted parent identity"},
		{name: "archive outside store", reason: capture.IsolationConflicting, detail: "escapes its registered store"},
		// Errors other than not-found still fail the row, as before.
		{name: "unreadable registered path", detail: "excessive symlink traversal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			captureProjectConfig(t, project, "ArchivedUnassignedFixture")
			s, err := capture.OpenProject(t.TempDir(), project)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			})
			home, active, archive := archivedRolloutFixture(t)
			path := filepath.Join(active, "child.jsonl")
			archived := filepath.Join(archive, filepath.Base(path))
			writeSource := func(path, id, cwd, parent string, tokens int) {
				t.Helper()
				body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q,"parent_thread_id":%q,"timestamp":"2026-01-01T00:00:00Z"}}`+"\n"+
					`{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn"}}`+"\n"+
					`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn","usage":{"input_tokens":%d}}}`+"\n", id, cwd, parent, id, tokens)
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			writeSource(path, "child", project, "parent", 17)
			identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: "child"}
			if _, err := s.Register(ctx, identity, path, "SessionStart"); err != nil {
				t.Fatal(err)
			}
			unassignedPass := func() capture.ReconcileResult {
				t.Helper()
				result, err := s.Reconcile(ctx, "", "tester", 16, fixtureByteBudget, fixtureGenerationPolicy())
				if err != nil {
					t.Fatal(err)
				}
				return result
			}
			before := archivedFixtureOwner(t, unassignedPass(), "child")
			if before.TicketID != "" || before.State != domain.CaptureUnassigned || before.ParentID != "parent" || len(before.Snapshot) != 1 {
				t.Fatalf("fixture must accept an unassigned child of parent: %+v", before)
			}

			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			switch tc.name {
			case "recovered":
				writeSource(archived, "child", project, "parent", 29)
			case "missing":
				// A similarly named archive file with another identity is not it.
				writeSource(archived, "other", project, "parent", 29)
			case "ambiguous":
				writeSource(archived, "child", project, "parent", 29)
				writeSource(filepath.Join(archive, "copy.jsonl"), "child", project, "parent", 37)
			case "foreign project":
				writeSource(archived, "child", t.TempDir(), "parent", 29)
			case "conflicting parent":
				writeSource(archived, "child", project, "replacement-parent", 29)
			case "archive outside store":
				external := t.TempDir()
				externalPath := filepath.Join(external, filepath.Base(archived))
				writeSource(externalPath, "child", project, "parent", 29)
				if err := os.Remove(archive); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(external, archive); err != nil {
					t.Fatal(err)
				}
				// The only proving rollout is reachable from the store through an
				// alias whose physical file lies outside it.
				if err := os.Symlink(externalPath, filepath.Join(active, "alias.jsonl")); err != nil {
					t.Fatal(err)
				}
			case "unreadable registered path":
				if err := os.Symlink(filepath.Base(path), path); err != nil {
					t.Fatal(err)
				}
				writeSource(archived, "child", project, "parent", 29)
			}
			result := unassignedPass()
			got := archivedFixtureOwner(t, result, "child")
			if got.TicketID != "" || got.ParentID != before.ParentID || got.Binding != before.Binding {
				t.Fatalf("unassigned pass changed attribution: %+v", got)
			}
			var isolated []capture.IsolatedSource
			for _, source := range result.IsolatedSources {
				if source.Identity.SessionID == "child" {
					isolated = append(isolated, source)
				}
			}
			switch {
			case tc.name == "recovered":
				if got.State != domain.CaptureUnassigned || got.TranscriptPath != archived || len(isolated) != 0 ||
					len(got.Snapshot) != 1 || len(got.Snapshot[0].Invocations) != 1 || got.Snapshot[0].Invocations[0].Tokens.Input != 29 {
					t.Fatalf("archived unassigned child not recovered: %+v isolated=%+v", got, isolated)
				}
				if repeated := archivedFixtureOwner(t, unassignedPass(), "child"); !reflect.DeepEqual(repeated.Snapshot, got.Snapshot) || repeated.TranscriptPath != archived {
					t.Fatal("repeating the recovered pass changed the archived child")
				}
				// Its facts stay available for binding, at the recovered path.
				bound, err := s.Bind(ctx, identity, "TEST-100", got.Request)
				if err != nil || bound.TranscriptPath != archived || !reflect.DeepEqual(bound.Snapshot, got.Snapshot) {
					t.Fatalf("recovered child could not be bound with its facts: %+v %v", bound, err)
				}
			case tc.reason != "":
				if len(isolated) != 1 || isolated[0].Reason != tc.reason || !strings.Contains(strings.ToLower(isolated[0].Detail), tc.detail) {
					t.Fatalf("isolated = %+v, want one %s source naming %q", isolated, tc.reason, tc.detail)
				}
				if got.State != domain.CapturePending || got.TranscriptPath != path || !reflect.DeepEqual(got.Snapshot, before.Snapshot) || got.Detail != isolated[0].Detail {
					t.Fatalf("isolated row did not keep its accepted facts and path: %+v", got)
				}
			default:
				if got.State != domain.CaptureFailed || len(isolated) != 0 || !strings.Contains(strings.ToLower(got.Detail), tc.detail) || !reflect.DeepEqual(got.Snapshot, before.Snapshot) {
					t.Fatalf("a non-not-found error must still fail with facts kept: %+v isolated=%+v", got, isolated)
				}
			}
		})
	}
}
