package capture

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func relocationFixture(t *testing.T) (string, string, string, *sql.DB) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	project := filepath.Join(root, "project")
	source := filepath.Join(root, "old-state")
	base := filepath.Join(root, "new-state")
	if err := os.MkdirAll(project, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(source, 0700); err != nil {
		t.Fatal(err)
	}
	writeRelocationConfig(t, project, source)
	db, err := sql.Open("sqlite", filepath.Join(source, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err = db.Exec("PRAGMA journal_mode=WAL; CREATE TABLE capture_meta(version INTEGER NOT NULL); INSERT INTO capture_meta VALUES(1); CREATE TABLE project_owner(singleton INTEGER PRIMARY KEY,directory TEXT NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO project_owner VALUES(1,?)", project); err != nil {
		t.Fatal(err)
	}
	for _, table := range relocationTables {
		if _, err = db.Exec("CREATE TABLE " + table.name + "(" + table.definition + ")"); err != nil {
			t.Fatal(err)
		}
	}
	reg := domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: domain.SourceCodex, Store: filepath.Join(root, "source-home"), SessionID: "registered-before-kill"}, State: domain.CaptureUnassigned, Request: "retained-explicit-binding-challenge"}
	data, err := encode(reg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO registrations VALUES(?,?,?,?,?,?)", IdentityKey(reg.Identity), "codex", reg.Identity.Store, reg.Identity.SessionID, "", data); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO publication_request VALUES(1,7,3,123);INSERT INTO pending_starts VALUES('store','parent','agent'); INSERT INTO capture_diagnostics VALUES(4,'T-1','reconciliation','failed','{}'); INSERT INTO capture_diagnostics VALUES(9,'T-1','reconciliation','accepted','{}'); INSERT INTO capture_diagnostic_heads VALUES('T-1','reconciliation','','accepted',9); INSERT INTO capture_diagnostic_failures VALUES(4); INSERT INTO capture_diagnostic_index VALUES(1,4,9,9,0)"); err != nil {
		t.Fatal(err)
	}
	accepted := relocationAcceptedSnapshot(project)
	data, err = encode(accepted)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO registrations VALUES(?,?,?,?,?,?)", IdentityKey(accepted.Identity), string(accepted.Identity.Source), accepted.Identity.Store, accepted.Identity.SessionID, accepted.TicketID, data); err != nil {
		t.Fatal(err)
	}
	key := IdentityKey(accepted.Identity)
	if _, err = db.Exec("INSERT INTO discovery_cursor VALUES(?, 'sessions/next-discovery.jsonl'); INSERT INTO child_cursor VALUES('queued',?,'sessions/next-child.jsonl',1); INSERT INTO project_proof_cursor VALUES(1,?); INSERT INTO publication_cursor VALUES(1,'T-1'); INSERT INTO reconcile_cursor VALUES('T-1',?); INSERT INTO recovery_phase VALUES('T-1',2)", key, key, key, key); err != nil {
		t.Fatal(err)
	}
	return project, source, base, db
}

// Accepted source facts are separate from the unassigned first handoff: copying
// pending work must neither invent a snapshot nor discard an already accepted one.
func relocationAcceptedSnapshot(project string) domain.CaptureRegistration {
	start, end := "2026-01-02T03:04:05Z", "2026-01-02T03:05:05Z"
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: filepath.Join(filepath.Dir(project), "source-home"), SessionID: "accepted-source-owner"}
	return domain.CaptureRegistration{
		Identity: id, ProjectProof: domain.CaptureProjectVerified, ProjectDirectory: project,
		TicketID: "T-1", TicketTitle: "Accepted capture", TicketType: "feat", Binding: "explicit_agent_response",
		Request: "retained-accepted-binding-challenge", State: domain.CaptureCurrent, Event: "Stop",
		TranscriptPath: filepath.Join(id.Store, "sessions", "accepted.jsonl"), Revision: 8192,
		SourceFingerprint: "8192:1767323105000000000", UpdatedAt: end, RefreshedAt: "2026-01-02T03:06:00Z",
		Snapshot: []domain.SessionFact{{
			ID: id.SessionID, Source: id.Source, User: "fixture-user",
			Invocations: []domain.InvocationFact{{Role: "coder", Model: "recorded-model", StartedAt: &start, CompletedAt: &end,
				ActiveDurationSeconds: 60, Turns: 3, Tokens: domain.InvocationTokens{Input: 23, Output: 7, CacheRead: 11, Reasoning: 2, PeakContext: 41},
				CostUSD: 0, CostEvidence: domain.ObservedCost(0), UserIterations: 2}},
			Tools: []domain.SessionToolFact{{Name: "exec", Calls: 4, Errors: 1}}, Anomalies: []domain.Anomaly{},
		}},
	}
}

func assertRelocationRegistrations(t *testing.T, rows []domain.CaptureRegistration, project string) {
	t.Helper()
	if len(rows) != 2 {
		t.Fatalf("accepted snapshot or pending handoff lost/duplicated: %+v", rows)
	}
	found := map[string]domain.CaptureRegistration{}
	for _, row := range rows {
		found[row.Identity.SessionID] = row
	}
	pending := found["registered-before-kill"]
	if pending.Request != "retained-explicit-binding-challenge" || pending.State != domain.CaptureUnassigned || pending.Binding != "" || pending.TicketID != "" || len(pending.Snapshot) != 0 {
		t.Fatalf("pending handoff lost or promoted to accepted facts: %+v", pending)
	}
	if want, got := relocationAcceptedSnapshot(project), found["accepted-source-owner"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("accepted source snapshot/binding/checkpoints changed:\n got %+v\nwant %+v", got, want)
	}
}

// These independent queries/expected values intentionally do not use the
// production relocationTables list, so omitting any import leaves a failing oracle.
func relocationContinuationValues(t *testing.T, store *Store) []string {
	t.Helper()
	queries := []struct{ table, value string }{
		{"discovery_cursor", "parent||'|'||path"},
		{"child_cursor", "kind||'|'||parent||'|'||key||'|'||more"},
		{"project_proof_cursor", "singleton||'|'||identity"},
		{"publication_cursor", "singleton||'|'||ticket"},
		{"reconcile_cursor", "ticket||'|'||identity"},
		{"recovery_phase", "ticket||'|'||phase"},
	}
	values := make([]string, 0, len(queries))
	for _, query := range queries {
		var count int
		if err := store.db.QueryRow("SELECT count(*) FROM "+query.table+" WHERE worktree_id=?", store.worktreeID).Scan(&count); err != nil || count != 1 {
			t.Fatalf("%s scope %d row count=%d err=%v", query.table, store.worktreeID, count, err)
		}
		var value string
		if err := store.db.QueryRow("SELECT "+query.value+" FROM "+query.table+" WHERE worktree_id=?", store.worktreeID).Scan(&value); err != nil {
			t.Fatalf("%s scope %d: %v", query.table, store.worktreeID, err)
		}
		values = append(values, value)
	}
	return values
}

func assertRelocationContinuations(t *testing.T, store *Store, project string) {
	t.Helper()
	key := IdentityKey(relocationAcceptedSnapshot(project).Identity)
	want := []string{key + "|sessions/next-discovery.jsonl", "queued|" + key + "|sessions/next-child.jsonl|1", "1|" + key, "1|T-1", "T-1|" + key, "T-1|2"}
	if got := relocationContinuationValues(t, store); !reflect.DeepEqual(got, want) {
		t.Fatalf("continuation values changed in exact scope %d: got %q want %q", store.worktreeID, got, want)
	}
}

func writeRelocationConfig(t *testing.T, project, state string) {
	t.Helper()
	data, err := fixtureSharedYAML(t, "Shared", "workspaces")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, ".agenticlens-config.yaml")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n", state)
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestRelocationPreservesScopedFactsAndFencesAlreadyOpenLegacyWriter(t *testing.T) {
	project, source, base, old := relocationFixture(t)
	other := filepath.Join(filepath.Dir(project), "other")
	if err := os.MkdirAll(other, 0700); err != nil {
		t.Fatal(err)
	}
	writeRelocationConfig(t, other, base)
	target, err := ProjectState(base, other)
	if err != nil {
		t.Fatal(err)
	}
	peer, err := OpenProject(target, other)
	if err != nil {
		t.Fatal(err)
	}
	peerID := peer.worktreeID
	if _, err = peer.db.Exec("INSERT INTO discovery_cursor VALUES(?, 'peer-parent', 'peer-discovery'); INSERT INTO child_cursor VALUES(?,'queued','peer-parent','peer-child',1); INSERT INTO project_proof_cursor VALUES(?,1,'peer-proof'); INSERT INTO publication_cursor VALUES(?,1,'PEER-29'); INSERT INTO reconcile_cursor VALUES(?,'PEER-29','peer-reconcile'); INSERT INTO recovery_phase VALUES(?,'PEER-29',1)", peer.worktreeID, peer.worktreeID, peer.worktreeID, peer.worktreeID, peer.worktreeID, peer.worktreeID); err != nil {
		t.Fatal(err)
	}
	peerContinuations := relocationContinuationValues(t, peer)
	if err = peer.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: "T-1", Phase: "publication", Outcome: "failed", Detail: "other-worktree"}); err != nil {
		t.Fatal(err)
	}
	if err = peer.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = peer.Close(); err != nil {
		t.Fatal(err)
	}
	result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "completed" || result.WorktreeID == peerID {
		t.Fatalf("unexpected scoped result: %+v", result)
	}
	if _, err = old.Exec("INSERT INTO pending_starts VALUES('new','parent','agent')"); err == nil || !strings.Contains(err.Error(), "relocated") {
		t.Fatalf("already-open old writer not fenced: %v", err)
	}
	if _, err = old.Exec("UPDATE capture_meta SET version=1"); err == nil {
		t.Fatal("old metadata guard could be reset")
	}
	current, err := OpenProject(target, project)
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	rows, err := current.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertRelocationRegistrations(t, rows, project)
	assertRelocationContinuations(t, current, project)
	progress, err := current.PublicationRequest(context.Background())
	if err != nil || progress.Generation != 7 || progress.Acknowledged != 3 {
		t.Fatalf("pending generation lost: %+v %v", progress, err)
	}
	var seq int64
	if err = current.db.QueryRow("SELECT sequence FROM capture_diagnostic_heads WHERE worktree_id=? AND ticket='T-1' AND kind='accepted'", current.worktreeID).Scan(&seq); err != nil || seq != 10 {
		t.Fatalf("diagnostic association not remapped: %d %v", seq, err)
	}
	peer, err = OpenProject(target, other)
	if err != nil {
		t.Fatal(err)
	}
	defer peer.Close()
	p, err := peer.PublicationRequest(context.Background())
	if err != nil || p.Generation != 1 || p.Acknowledged != 0 {
		t.Fatalf("other worktree changed: %+v %v", p, err)
	}
	if got := relocationContinuationValues(t, peer); !reflect.DeepEqual(got, peerContinuations) {
		t.Fatalf("migration changed peer continuation state: got %q want %q", got, peerContinuations)
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil || cfg.Telemetry.Capture.StateDir != base {
		t.Fatalf("configuration cutover missing: %+v %v", cfg.Telemetry.Capture, err)
	}
	for _, name := range []string{"registry.db", "relocation-prior.db", "relocation-fenced.db", "relocation.json"} {
		if _, err = os.Stat(filepath.Join(source, name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base}); err != nil {
		t.Fatalf("completed retry not idempotent: %v", err)
	}
	assertRelocationContinuations(t, current, project)
	rows, err = current.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertRelocationRegistrations(t, rows, project)
	if got := relocationContinuationValues(t, peer); !reflect.DeepEqual(got, peerContinuations) {
		t.Fatalf("completed retry changed peer continuation state: %q", got)
	}
}
func TestRelocationPreviewAndBusyWorkerDoNotChangeSource(t *testing.T) {
	project, source, base, db := relocationFixture(t)
	before, err := os.ReadFile(filepath.Join(project, ".agenticlens-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base, DryRun: true})
	if err != nil || !result.DryRun {
		t.Fatalf("preview: %+v %v", result, err)
	}
	if _, err = os.Stat(base); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preview created state: %v", err)
	}
	worker, err := AcquireWorker(source)
	if err != nil || worker == nil {
		t.Fatal(err)
	}
	_, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if e := worker.Close(); e != nil {
		t.Fatal(e)
	}
	if err == nil || !strings.Contains(err.Error(), "worker is active") {
		t.Fatalf("active worker accepted: %v", err)
	}
	var version int
	if err = db.QueryRow("SELECT version FROM capture_meta").Scan(&version); err != nil || version != 1 {
		t.Fatalf("source changed: %d %v", version, err)
	}
	after, err := os.ReadFile(filepath.Join(project, ".agenticlens-config.yaml"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("configuration changed before migration")
	}
}
func TestRelocationResumesPendingCutoverAndRetainsDeletedWorktreeDiagnostics(t *testing.T) {
	project, source, base, _ := relocationFixture(t)
	target, err := ProjectState(base, project)
	if err != nil {
		t.Fatal(err)
	}
	allocated, err := OpenProject(target, project)
	if err != nil {
		t.Fatal(err)
	}
	coord := allocated.CoordinationDirectory()
	if err = allocated.Close(); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(coord, "prior-worker.log")
	if err = os.MkdirAll(blocker, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(source, "worker.log"), []byte("retained failed worker"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if err == nil {
		t.Fatal("expected staged evidence failure")
	}
	if _, err = OpenProject(target, project); err == nil {
		t.Fatal("pending destination accepted ordinary operations")
	}
	if err = os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	if _, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base}); err != nil {
		t.Fatalf("resume: %v", err)
	}
	resumed, err := OpenProject(target, project)
	if err != nil {
		t.Fatal(err)
	}
	assertRelocationContinuations(t, resumed, project)
	rows, err := resumed.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	assertRelocationRegistrations(t, rows, project)
	if err = resumed.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.RemoveAll(project); err != nil {
		t.Fatal(err)
	}
	view, err := InspectRetainedState(context.Background(), target, "Shared", project)
	if err != nil || len(view.Registrations) != 2 || view.Publication.Generation != 7 {
		t.Fatalf("deleted-worktree state unavailable: %+v %v", view, err)
	}
	assertRelocationRegistrations(t, view.Registrations, project)
	if _, err = os.Stat(project); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("inspection recreated deleted worktree")
	}
}
func TestRelocationRefusesUnknownSchemaBeforeFencing(t *testing.T) {
	project, source, base, db := relocationFixture(t)
	if _, err := db.Exec("CREATE TABLE unknown_future_data(id INTEGER)"); err != nil {
		t.Fatal(err)
	}
	_, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if err == nil || !strings.Contains(err.Error(), "unknown legacy table") {
		t.Fatalf("unknown schema accepted: %v", err)
	}
	var version int
	if e := db.QueryRow("SELECT version FROM capture_meta").Scan(&version); e != nil || version != 1 {
		t.Fatal("unknown source was changed")
	}
}

func TestRelocationBusySQLiteWriterPreservesSource(t *testing.T) {
	project, source, base, db := relocationFixture(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec("UPDATE publication_request SET generation=generation"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	_, err = RelocateState(ctx, RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
	if e := tx.Rollback(); e != nil {
		t.Fatal(e)
	}
	if err == nil {
		t.Fatal("busy source writer did not prevent relocation")
	}
	var version int
	if e := db.QueryRow("SELECT version FROM capture_meta").Scan(&version); e != nil || version != 1 {
		t.Fatal("busy source was fenced")
	}
}

func TestRelocationRefusesSharedProtocolTwoWithoutChangingPeers(t *testing.T) {
	project, _, base, _ := relocationFixture(t)
	logical, err := ProjectState(base, project)
	if err != nil {
		t.Fatal(err)
	}
	store, err := OpenProject(logical, project)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: logical, ToBase: filepath.Join(t.TempDir(), "unsupported")})
	if err == nil || !strings.Contains(err.Error(), "protocol-2") {
		t.Fatalf("shared relocation boundary unclear: %v", err)
	}
	err = CheckStateRelocation(project, base, filepath.Join(t.TempDir(), "unsupported"))
	if err == nil || !strings.Contains(err.Error(), "protocol-2") {
		t.Fatalf("configuration abandonment allowed: %v", err)
	}
	store, err = OpenProject(logical, project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	request, err := store.PublicationRequest(context.Background())
	if err != nil || request.Generation != 1 {
		t.Fatalf("shared data changed: %+v %v", request, err)
	}
}

func TestAdministrativeRelocationGuardReadsIncompleteConfigWithoutLoadFallback(t *testing.T) {
	project, source, base, db := relocationFixture(t)
	path := filepath.Join(project, ".agenticlens-config.yaml")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(body), "\n")
	kept := lines[:0]
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "base_branch:") {
			kept = append(kept, line)
		}
	}
	if err = os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0600); err != nil {
		t.Fatal(err)
	}
	err = CheckStateRelocation(project, source, base)
	if err == nil || !strings.Contains(err.Error(), "state exists") {
		t.Fatalf("administrative preview did not protect old state: %v", err)
	}
	if err = CheckStateRelocation(project, source, source); err != nil {
		t.Fatalf("declaring unchanged known source was blocked: %v", err)
	}
	var generations int
	if err = db.QueryRow("SELECT generation FROM publication_request").Scan(&generations); err != nil || generations != 7 {
		t.Fatal("administrative guard changed pending work")
	}
	if err = CheckStateRelocation(project, filepath.Join(project, ".agenticlens-capture"), base); err != nil {
		t.Fatalf("missing-key repair with no prior implicit state failed: %v", err)
	}
}

func TestRelocationExplicitPathsUseSelectedProjectAndHome(t *testing.T) {
	project, _, _, _ := relocationFixture(t)
	t.Setenv("HOME", filepath.Dir(project))
	result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: "../old-state", ToBase: "../new-state"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != filepath.Join(filepath.Dir(project), "old-state") {
		t.Fatalf("source used caller directory: %+v", result)
	}
	view, err := InspectRetainedState(context.Background(), "~/new-state/projects/Shared", "Shared", project)
	if err != nil || len(view.Registrations) != 2 {
		t.Fatalf("explicit home path unavailable: %+v %v", view, err)
	}
}

func TestRelocationPreservesDisabledLegacyInstallationAndRefreshesOwnedAnchor(t *testing.T) {
	for _, scenario := range []string{"normal", "manifest-written-anchor-old", "foreign-old-anchor"} {
		t.Run(scenario, func(t *testing.T) {
			project, source, base, _ := relocationFixture(t)
			codex := filepath.Join(project, ".codex")
			if err := os.MkdirAll(codex, 0700); err != nil {
				t.Fatal(err)
			}
			manifest := Manifest{Protocol: 1, Source: "codex", Repository: project, StateDir: source, SourceStore: t.TempDir(), ConfigPath: filepath.Join(codex, "hooks.json"), Active: false}
			body, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			anchorPath := filepath.Join(codex, "agenticlens-installation.json")
			for _, path := range []string{manifestPath(source, "codex"), anchorPath} {
				if err = os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(manifest.ConfigPath, []byte("{}"), 0600); err != nil {
				t.Fatal(err)
			}
			target, err := ProjectState(base, project)
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "manifest-written-anchor-old" {
				// Model the durable checkpoint after data import and manifest replacement,
				// before the project anchor changes. The original anchor is never edited.
				allocated, e := OpenProject(target, project)
				if e != nil {
					t.Fatal(e)
				}
				coord := allocated.CoordinationDirectory()
				if e = allocated.Close(); e != nil {
					t.Fatal(e)
				}
				targetManifest := manifestPath(coord, "codex")
				if err = os.MkdirAll(targetManifest, 0700); err != nil {
					t.Fatal(err)
				}
				if _, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base}); err == nil {
					t.Fatal("expected staging obstruction")
				}
				if err = os.Remove(targetManifest); err != nil {
					t.Fatal(err)
				}
				checkpoint := manifest
				checkpoint.Protocol = 2
				checkpoint.StateDir = filepath.Dir(targetManifest)
				checkpointBody, e := json.Marshal(checkpoint)
				if e != nil {
					t.Fatal(e)
				}
				if err = os.WriteFile(targetManifest, checkpointBody, 0600); err != nil {
					t.Fatal(err)
				}
				unchanged, e := os.ReadFile(anchorPath)
				if e != nil || !bytes.Equal(unchanged, body) {
					t.Fatal("pending fixture changed old anchor")
				}
			}
			if scenario == "foreign-old-anchor" {
				foreign := manifest
				foreign.StateDir = t.TempDir()
				foreignBody, e := json.Marshal(foreign)
				if e != nil {
					t.Fatal(e)
				}
				if err = os.WriteFile(anchorPath, foreignBody, 0600); err != nil {
					t.Fatal(err)
				}
				if _, err = RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base}); err == nil {
					t.Fatal("foreign legacy anchor was accepted")
				}
				after, e := os.ReadFile(anchorPath)
				if e != nil || !bytes.Equal(after, foreignBody) {
					t.Fatal("foreign legacy anchor was overwritten")
				}
				if _, err = OpenProject(target, project); err == nil {
					t.Fatal("foreign-anchor failure activated destination")
				}
				return
			}
			result, err := RelocateState(context.Background(), RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
			if err != nil {
				t.Fatal(err)
			}
			store, err := OpenProject(result.Destination, project)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			current, err := ReadManifest(store.directory, "codex")
			if err != nil || current.Protocol != 2 || current.Active || current.StateDir != store.directory {
				t.Fatalf("disabled installation changed: %+v %v", current, err)
			}
			anchor, err := os.ReadFile(anchorPath)
			if err != nil {
				t.Fatal(err)
			}
			var anchored Manifest
			if err = json.Unmarshal(anchor, &anchored); err != nil || anchored.Protocol != 2 || anchored.StateDir != store.directory {
				t.Fatalf("legacy anchor not refreshed: %+v %v", anchored, err)
			}
		})
	}
}

// This reproduces the architecture review's actual allocation/OpenProject/Register
// interleaving, with and without a handle opened before reservation.
func TestRelocationReservationFencesAllocationBeforeSourceRetirement(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint("existing-handle-", existing), func(t *testing.T) {
			ctx := context.Background()
			project, source, base, legacy := relocationFixture(t)
			// An operator may already have explicitly selected the destination.
			writeRelocationConfig(t, project, base)
			target, err := ProjectState(base, project)
			if err != nil {
				t.Fatal(err)
			}
			peerProject := filepath.Join(filepath.Dir(project), "peer")
			if err = os.Mkdir(peerProject, 0700); err != nil {
				t.Fatal(err)
			}
			writeRelocationConfig(t, peerProject, base)
			peer, err := OpenProject(target, peerProject)
			if err != nil {
				t.Fatal(err)
			}
			defer peer.Close()
			if err = peer.RequestPublication(ctx); err != nil {
				t.Fatal(err)
			}
			var before *Store
			if existing {
				before, err = OpenProject(target, project)
				if err != nil {
					t.Fatal(err)
				}
				defer before.Close()
			}
			journal := relocationJournal{ID: "atomic-reservation", Project: project, Name: "Shared", Destination: target, Phase: "prepared"}
			if err = writeRelocationJournal(filepath.Join(source, "relocation.json"), journal); err != nil {
				t.Fatal(err)
			}
			reserved, err := openRegistryForRelocation(target, "Shared", project, journal.ID)
			if err != nil {
				t.Fatal(err)
			}
			defer reserved.Close()
			var version int
			if err = legacy.QueryRow("SELECT version FROM capture_meta").Scan(&version); err != nil || version != 1 {
				t.Fatal("allocation retired source before its fence")
			}
			policy, err := adaptersconfig.ResolveCapture(project, "")
			if err != nil {
				t.Fatal(err)
			}
			resolved, err := ProjectState(policy.StateDir, project)
			if err != nil || resolved != target {
				t.Fatalf("destination-configured path changed: %s %v", resolved, err)
			}
			if ordinary, e := OpenProject(resolved, project); e == nil {
				ordinary.Close()
				t.Fatal("allocated pending scope became operational")
			}
			if before != nil {
				id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "allocation-gap-handoff"}
				if _, err = before.Register(ctx, id, "", "UserPromptSubmit"); err == nil {
					t.Fatal("already-open destination handle accepted pending work")
				}
			}
			if wrong, e := openRegistryForRelocation(target, "Shared", project, "different-migration"); e == nil {
				wrong.Close()
				t.Fatal("pending activation bypass accepted another migration identity")
			}
			if err = peer.RequestPublication(ctx); err != nil {
				t.Fatalf("migration fenced peer scope: %v", err)
			}
			if err = fenceLegacyRegistry(ctx, legacy, project); err != nil {
				t.Fatal(err)
			}
			if err = importRelocation(ctx, reserved, source, journal.ID); err != nil {
				t.Fatal(err)
			}
			if err = reserved.Close(); err != nil {
				t.Fatal(err)
			}
			journal.Phase = "imported"
			if err = writeRelocationJournal(filepath.Join(source, "relocation.json"), journal); err != nil {
				t.Fatal(err)
			}
			if _, err = RelocateState(ctx, RelocationOptions{Project: project, FromDirectory: source, ToBase: base}); err != nil {
				t.Fatalf("reserved migration could not resume: %v", err)
			}
			final, err := OpenProject(target, project)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := final.List(ctx)
			if closeErr := final.Close(); closeErr != nil {
				t.Fatal(closeErr)
			}
			if err != nil {
				t.Fatal(err)
			}
			assertRelocationRegistrations(t, rows, project)
			request, err := peer.PublicationRequest(ctx)
			if err != nil || request.Generation != 2 || request.Acknowledged != 0 {
				t.Fatalf("peer acknowledgement changed: %+v %v", request, err)
			}
		})
	}
}

func TestRelocationRejectsReplacedDestinationOwnerAtEveryReservationBoundary(t *testing.T) {
	for _, boundary := range []string{"preview", "allocation", "import", "activation"} {
		t.Run(boundary, func(t *testing.T) {
			ctx := context.Background()
			project, source, base, legacy := relocationFixture(t)
			target, err := ProjectState(base, project)
			if err != nil {
				t.Fatal(err)
			}
			prior, err := OpenProject(target, project)
			if err != nil {
				t.Fatal(err)
			}
			if err = prior.Close(); err != nil {
				t.Fatal(err)
			}
			if err = previewRelocationDestination(ctx, target, "Shared", project, ""); err != nil {
				t.Fatal(err)
			}
			var reserved *Store
			if boundary == "import" || boundary == "activation" {
				reserved, err = openRegistryForRelocation(target, "Shared", project, "owner-check")
				if err != nil {
					t.Fatal(err)
				}
				defer reserved.Close()
			}
			if boundary == "activation" {
				if err = fenceLegacyRegistry(ctx, legacy, project); err != nil {
					t.Fatal(err)
				}
				if err = importRelocation(ctx, reserved, source, "owner-check"); err != nil {
					t.Fatal(err)
				}
			}
			configPath := filepath.Join(project, ".agenticlens-config.yaml")
			body, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			// Preserve the original inode so this cannot pass through inode reuse.
			if err = os.Rename(project, project+"-original"); err != nil {
				t.Fatal(err)
			}
			if err = os.Mkdir(project, 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(configPath, body, 0600); err != nil {
				t.Fatal(err)
			}
			local, err := os.ReadFile(filepath.Join(project+"-original", ".agenticlens-config.local.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), local, 0600); err != nil {
				t.Fatal(err)
			}
			if normal, e := OpenProject(target, project); e == nil {
				normal.Close()
				t.Fatal("negative ownership control accepted replacement")
			}
			switch boundary {
			case "preview":
				_, err = RelocateState(ctx, RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
			case "allocation":
				var rejected *Store
				rejected, err = openRegistryForRelocation(target, "Shared", project, "owner-check")
				if rejected != nil {
					rejected.Close()
				}
			case "import":
				err = importRelocation(ctx, reserved, source, "owner-check")
			case "activation":
				err = activateRelocation(ctx, reserved, "owner-check")
			}
			if err == nil || !strings.Contains(err.Error(), "ownership") {
				t.Fatalf("%s bypass adopted replacement owner: %v", boundary, err)
			}
			var version int
			expectedVersion := 1
			if boundary == "activation" {
				expectedVersion = -1
			}
			if err = legacy.QueryRow("SELECT version FROM capture_meta").Scan(&version); err != nil || version != expectedVersion {
				t.Fatalf("owner refusal changed source at %s: %d %v", boundary, version, err)
			}
			if boundary == "activation" {
				var state string
				if err = reserved.db.QueryRow("SELECT state FROM worktree_relocations WHERE worktree_id=?", reserved.worktreeID).Scan(&state); err != nil || state != "pending-imported" {
					t.Fatalf("owner refusal activated imported scope: %s %v", state, err)
				}
			}
			after, err := os.ReadFile(configPath)
			if err != nil || !bytes.Equal(after, body) {
				t.Fatal("rejected destination changed configuration")
			}
		})
	}
}

func TestRelocationReservationRefusesPriorDestinationWorkBeforeRetiringSource(t *testing.T) {
	for _, conflict := range []string{"active-worker", "accepted-handoff"} {
		t.Run(conflict, func(t *testing.T) {
			ctx := context.Background()
			project, source, base, legacy := relocationFixture(t)
			target, err := ProjectState(base, project)
			if err != nil {
				t.Fatal(err)
			}
			existing, err := OpenProject(target, project)
			if err != nil {
				t.Fatal(err)
			}
			defer existing.Close()
			if err = previewRelocationDestination(ctx, target, "Shared", project, ""); err != nil {
				t.Fatal(err)
			}
			var worker *PublicationLock
			if conflict == "active-worker" {
				worker, err = AcquireWorker(existing.directory)
				if err != nil || worker == nil {
					t.Fatal(err)
				}
				defer worker.Close()

			} else {
				id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "accepted-before-reservation"}
				if _, err = existing.Register(ctx, id, "", "UserPromptSubmit"); err != nil {
					t.Fatal(err)
				}
			}
			// Allocation's transaction must recheck after the readonly preview.
			reserved, err := openRegistryForRelocation(target, "Shared", project, "refused-reservation")
			if reserved != nil {
				reserved.Close()
			}
			if err == nil {
				t.Fatal("conflicting destination was reserved")
			}
			var version, markers int
			if err = legacy.QueryRow("SELECT version FROM capture_meta").Scan(&version); err != nil || version != 1 {
				t.Fatal("refused reservation retired source")
			}
			if err = existing.db.QueryRow("SELECT count(*) FROM worktree_relocations WHERE worktree_id=?", existing.worktreeID).Scan(&markers); err != nil || markers != 0 {
				t.Fatalf("failed reservation left a pending fence: %d %v", markers, err)
			}
			// The failed transaction must not fence the ordinary existing handle.
			if err = existing.RequestPublication(ctx); err != nil {
				t.Fatalf("refusal stranded destination operations: %v", err)
			}
			rows, err := existing.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if conflict == "accepted-handoff" && (len(rows) != 1 || rows[0].Identity.SessionID != "accepted-before-reservation") {
				t.Fatal("reservation refusal changed accepted destination work")
			}
			if _, err = os.Stat(filepath.Join(source, "relocation-fenced.db")); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("reservation refusal created a fenced source backup")
			}
		})
	}
}

func TestAdministrativeRelocationGuardUsesCanonicalConfiguredName(t *testing.T) {
	for _, input := range []struct {
		label, projectField, expected string
	}{
		{"plain", "project_name: SpacedName", "SpacedName"},
		{"spaced", "project_name: ' SpacedName '", "SpacedName"},
		{"explicit-string-scalar", "project_name: !!str 12345", "12345"},
	} {
		t.Run(input.label, func(t *testing.T) {
			root := t.TempDir()
			project, oldBase, newBase := filepath.Join(root, "project"), filepath.Join(root, "old"), filepath.Join(root, "new")
			if err := os.Mkdir(project, 0700); err != nil {
				t.Fatal(err)
			}
			writeRelocationConfig(t, project, oldBase)
			path := filepath.Join(project, ".agenticlens-config.yaml")
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(body, []byte("project_name: Shared")) {
				t.Fatal("fixture project field missing")
			}
			body = bytes.Replace(body, []byte("project_name: Shared"), []byte(input.projectField), 1)
			if err = os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := adaptersconfig.Load(project)
			if err != nil || cfg.ProjectName != input.expected {
				t.Fatalf("canonical operational name: %q %v", cfg.ProjectName, err)
			}
			target, err := ProjectState(oldBase, project)
			if err != nil {
				t.Fatal(err)
			}
			if filepath.Base(target) != input.expected {
				t.Fatalf("operational namespace differed: %s", target)
			}
			store, err := OpenProject(target, project)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "canonical-name-handoff"}
			registration, err := store.Register(context.Background(), id, "", "UserPromptSubmit")
			if err != nil {
				t.Fatal(err)
			}
			err = CheckStateRelocation(project, oldBase, newBase)
			if err == nil || !strings.Contains(err.Error(), "protocol-2") {
				t.Fatalf("guard missed the canonical populated namespace: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, body) {
				t.Fatal("administrative guard changed configuration")
			}
			rows, err := store.List(context.Background())
			if err != nil || len(rows) != 1 || rows[0].Request != registration.Request {
				t.Fatal("administrative guard changed accepted state")
			}
			if _, err = os.Stat(newBase); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("guard created the proposed state location")
			}
			// Raw repair still finds the name without required code-metric policy;
			// it must not depend on a full operational configuration load.
			{
				lines := strings.Split(string(body), "\n")
				kept := lines[:0]
				for _, line := range lines {
					if !strings.HasPrefix(strings.TrimSpace(line), "base_branch:") {
						kept = append(kept, line)
					}
				}
				incomplete := []byte(strings.Join(kept, "\n"))
				if err = os.WriteFile(path, incomplete, 0600); err != nil {
					t.Fatal(err)
				}
				if err = CheckStateRelocation(project, oldBase, newBase); err == nil || !strings.Contains(err.Error(), "protocol-2") {
					t.Fatalf("raw repair lost canonical namespace: %v", err)
				}
				after, err = os.ReadFile(path)
				if err != nil || !bytes.Equal(after, incomplete) {
					t.Fatal("raw name check rewrote incomplete configuration")
				}
			}
		})
	}
}

func TestAdministrativeRelocationRejectsUnsupportedNameShapesWithoutWrites(t *testing.T) {
	for _, field := range []string{
		"label: &label SpacedName\nproject_name: *label",
		"defaults: &defaults\n  project_name: SpacedName\n<<: *defaults",
		"project_name: 12345",
	} {
		t.Run(field, func(t *testing.T) {
			project := t.TempDir()
			oldBase := t.TempDir()
			newBase := filepath.Join(t.TempDir(), "not-created")
			writeRelocationConfig(t, project, oldBase)
			path := filepath.Join(project, ".agenticlens-config.yaml")
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			body = bytes.Replace(body, []byte("project_name: Shared"), []byte(field), 1)
			if err = os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = adaptersconfig.Load(project); err == nil {
				t.Fatal("unsupported typed name accepted by operational reader")
			}
			if err = CheckStateRelocation(project, oldBase, newBase); err == nil {
				t.Fatal("unsupported name bypassed maintenance validation")
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, body) {
				t.Fatal("rejected name changed configuration")
			}
			if _, err = os.Stat(newBase); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("rejected name allocated a new state directory")
			}
		})
	}
}
