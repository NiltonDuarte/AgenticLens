package capture_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func writeRollout(t *testing.T, path, id, parent string, tokens int) []byte {
	t.Helper()
	metadata, err := json.Marshal(struct {
		ID        string `json:"id"`
		Parent    string `json:"parent_thread_id,omitempty"`
		Timestamp string `json:"timestamp"`
	}{id, parent, "2026-01-01T00:00:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	contents := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":%s}\n"+`{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"turn-one"}}`+"\n"+`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}`+"\n", metadata, id, tokens)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return []byte(contents)
}

func registerBound(t *testing.T, s *capture.Store, store, path, id, ticket string) domain.CaptureRegistration {
	t.Helper()
	row, err := s.Register(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: id}, path, "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	row, err = s.Bind(context.Background(), row.Identity, ticket, row.Request)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

func TestReconcilePreservesAcceptedSnapshotsThroughResumeTruncationAndMissingSource(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	path := filepath.Join(store, "root.jsonl")
	original := writeRollout(t, path, "root", "", 10)
	row := registerBound(t, s, store, path, "root", "TEST-100")
	reconcile := func() capture.ReconcileResult {
		t.Helper()
		result, err := s.Reconcile(ctx, "TEST-100", "tester", 20, capture.DefaultByteBudget)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := reconcile()
	if len(first.Sessions) != 1 || len(first.Sessions[0].Invocations) != 1 || first.Sessions[0].Invocations[0].Tokens.Input != 10 {
		t.Fatalf("source not captured: %+v", first)
	}
	if first.Sessions[0].Invocations[0].CostEvidence.State != domain.CostUnavailable {
		t.Fatal("missing cost became known zero")
	}
	if second := reconcile(); !reflect.DeepEqual(first.Sessions, second.Sessions) {
		t.Fatal("unchanged source duplicated facts")
	}
	growth := append(append([]byte{}, original...), []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"turn-two","usage":{"input_tokens":20,"output_tokens":3}}}`+"\n")...)
	if err := os.WriteFile(path, growth, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Register(ctx, row.Identity, path, "Stop"); err != nil {
		t.Fatal(err)
	}
	grown := reconcile()
	if len(grown.Sessions[0].Invocations) != 1 || grown.Sessions[0].Invocations[0].Tokens.Input != 30 {
		t.Fatalf("resume cumulative facts wrong: %+v", grown.Sessions)
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	truncated := reconcile()
	if !reflect.DeepEqual(grown.Sessions, truncated.Sessions) || truncated.Registrations[0].State != domain.CapturePending {
		t.Fatalf("truncation lost facts or completeness: %+v", truncated)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	missing := reconcile()
	if !reflect.DeepEqual(grown.Sessions, missing.Sessions) || missing.Registrations[0].State != domain.CaptureFailed {
		t.Fatalf("missing source erased facts: %+v", missing)
	}
}

func TestReconcileVerifiesChildSourceAndIsolatesIndependentTickets(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	rootPath, childPath := filepath.Join(store, "root.jsonl"), filepath.Join(store, "child.jsonl")
	writeRollout(t, rootPath, "root", "", 10)
	writeRollout(t, childPath, "source-child", "root", 3)
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	if err := s.QueueChild(ctx, store, childPath, "root"); err != nil {
		t.Fatal(err)
	}
	otherPath := filepath.Join(store, "other.jsonl")
	writeRollout(t, otherPath, "other", "", 7)
	registerBound(t, s, store, otherPath, "other", "TEST-200")
	unboundPath := filepath.Join(store, "unbound.jsonl")
	writeRollout(t, unboundPath, "unbound", "", 9)
	if _, err := s.Register(ctx, domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: "unbound"}, unboundPath, "SessionStart"); err != nil {
		t.Fatal(err)
	}
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 20, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]domain.SessionFact{}
	for _, session := range result.Sessions {
		byID[session.ID] = session
	}
	if len(byID) != 2 || byID["root"].ID == "" || byID["source-child"].ParentID == nil || *byID["source-child"].ParentID != "root" {
		t.Fatalf("wrong exact source tree: %+v", result)
	}
	for _, r := range result.Registrations {
		if r.Identity.SessionID == "unbound" {
			t.Fatalf("unassigned local registration leaked into ticket report: %+v", r)
		}
		if r.Identity.SessionID == "other" {
			t.Fatal("another ticket leaked into report")
		}
	}
	local, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range local {
		if r.Identity.SessionID == "unbound" {
			found = true
			if r.TicketID != "" {
				t.Fatal("guessed unassigned local ownership")
			}
		}
	}
	if !found {
		t.Fatal("unassigned session disappeared from local diagnostics")
	}
}

func TestReportDoesNotPublishLocalPathsOrBindingToken(t *testing.T) {
	registration := domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/private-source-store", SessionID: "root"}, TicketID: "TEST-100", Request: "private-binding-token", TranscriptPath: "/private-transcript", Snapshot: []domain.SessionFact{{ID: "root"}}}
	result := capture.ReconcileResult{Protocol: 1, Cutoff: "2026-01-01T00:00:00Z", TicketID: "TEST-100", Registrations: []domain.CaptureRegistration{registration}}
	b, err := capture.ReportJSON(result, "ledger-digest")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private-source-store", "private-binding-token", "private-transcript", "snapshot"} {
		if strings.Contains(string(b), forbidden) {
			t.Fatalf("report leaked %q: %s", forbidden, b)
		}
	}
	if !strings.Contains(string(b), `"ledger_sha256": "ledger-digest"`) || !strings.Contains(string(b), `"session_id": "root"`) {
		t.Fatalf("missing digest/identity: %s", b)
	}
	if result.Registrations[0].Identity.Store != "/private-source-store" || result.Registrations[0].Request != "private-binding-token" {
		t.Fatal("report redaction mutated local registration")
	}
}

func TestReconcileChargesEachSourceOnceAndContinuesBudgetedWork(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	path := filepath.Join(store, "root.jsonl")
	contents := writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 1, int64(len(contents)))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Registrations) != 1 || result.Registrations[0].State != domain.CaptureCurrent {
		t.Fatalf("exact one-source budget was charged again: %+v", result)
	}
	for _, id := range []string{"second", "third"} {
		path := filepath.Join(store, id+".jsonl")
		writeRollout(t, path, id, "", 3)
		registerBound(t, s, store, path, id, "TEST-100")
	}
	for i := 0; i < 3; i++ {
		if _, err := s.Reconcile(ctx, "TEST-100", "tester", 1, capture.DefaultByteBudget); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if len(row.Snapshot) != 1 {
			t.Fatalf("budget continuation starved owner %s: %+v", row.Identity.SessionID, rows)
		}
	}
}

func TestReconcileRejectsCanonicalIdentityCollisionAcrossStores(t *testing.T) {
	s := openStore(t, t.TempDir())
	for _, tokens := range []int{10, 20} {
		store := t.TempDir()
		path := filepath.Join(store, "same-session.jsonl")
		writeRollout(t, path, "same-session", "", tokens)
		registerBound(t, s, store, path, "same-session", "TEST-100")
	}
	if result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 20, capture.DefaultByteBudget); err == nil {
		t.Fatalf("canonical v5 cannot distinguish these owners: %+v", result.Sessions)
	}
}

func TestReconcilePreservesFactsWhenSourceParsingIsPartial(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	path := filepath.Join(store, "root.jsonl")
	original := writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	before, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	partial := writeRollout(t, path, "root", "", 1)
	partial = append(partial, []byte(strings.Repeat(" ", len(original))+"{\"type\":\n")...)
	if err := os.WriteFile(path, partial, 0600); err != nil {
		t.Fatal(err)
	}
	after, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before.Sessions, after.Sessions) || after.Registrations[0].State == domain.CaptureCurrent {
		t.Fatalf("partial parse replaced authoritative snapshot or claimed current: %+v", after)
	}
}

func TestReconcileFailedReadsConsumeSharedByteBudget(t *testing.T) {
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	var size int64
	for _, id := range []string{"owner-a", "owner-b"} {
		path := filepath.Join(store, id+".jsonl")
		contents := append(writeRollout(t, path, id, "", 10), []byte("{\"type\":\n")...)
		if err := os.WriteFile(path, contents, 0600); err != nil {
			t.Fatal(err)
		}
		size = int64(len(contents))
		registerBound(t, s, store, path, id, "TEST-100")
	}
	result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 2, size)
	if err != nil {
		t.Fatal(err)
	}
	failed, pending := 0, 0
	for _, row := range result.Registrations {
		if row.State == domain.CaptureFailed {
			failed++
		}
		if row.State == domain.CapturePending {
			pending++
		}
	}
	if failed != 1 || pending != 1 {
		t.Fatalf("failed parsing escaped the shared byte budget: failed=%d pending=%d rows=%+v", failed, pending, result.Registrations)
	}
	if result.SourceBytes != size || result.SourceAttempts > 2 {
		t.Fatalf("incorrect failed-read accounting: bytes=%d attempts=%d budget=%d", result.SourceBytes, result.SourceAttempts, size)
	}
}

func TestReconcileRecoversSourceProvenChildAfterStartWithoutStop(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	sessions := filepath.Join(store, "sessions")
	if err := os.MkdirAll(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(sessions, "root.jsonl")
	writeRollout(t, rootPath, "root", "", 10)
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SubagentStart", TranscriptPath: rootPath, AgentID: "unproven-interface-agent"}, store, "/binary", "/state"); err != nil {
		t.Fatal(err)
	}
	writeRollout(t, filepath.Join(sessions, "child.jsonl"), "actual-source-child", "root", 3)
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range result.Sessions {
		if row.ID == "unproven-interface-agent" {
			t.Fatal("guessed harness ID from interface agent ID")
		}
		if row.ID == "actual-source-child" {
			found = true
			if row.ParentID == nil || *row.ParentID != "root" {
				t.Fatalf("child lacks source-proven parent: %+v", row)
			}
		}
	}
	if !found {
		t.Fatalf("start evidence lost before stop: %+v", result)
	}
}

func TestReconcileUnassignedRootsDoNotConsumeTicketWorkOrPublishIdentities(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	boundID := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store}
	for n := 0; ; n++ {
		boundID.SessionID = fmt.Sprintf("bound-%d", n)
		if strings.HasPrefix(capture.IdentityKey(boundID), "f") {
			break
		}
	}
	path := filepath.Join(store, "bound.jsonl")
	contents := writeRollout(t, path, boundID.SessionID, "", 10)
	registerBound(t, s, store, path, boundID.SessionID, "TEST-100")
	unassigned := []string{}
	for n := 0; len(unassigned) < 4; n++ {
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: fmt.Sprintf("unassigned-%d", n)}
		if capture.IdentityKey(id) >= capture.IdentityKey(boundID) {
			continue
		}
		path := filepath.Join(store, id.SessionID+".jsonl")
		writeRollout(t, path, id.SessionID, "", 99)
		if _, err := s.Register(ctx, id, path, "SessionStart"); err != nil {
			t.Fatal(err)
		}
		unassigned = append(unassigned, id.SessionID)
	}
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 1, int64(len(contents)))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].ID != boundID.SessionID || len(result.Registrations) != 1 || result.Registrations[0].State != domain.CaptureCurrent {
		t.Fatalf("unassigned roots starved ticket: %+v", result)
	}
	report, err := capture.ReportJSON(result, "digest")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range unassigned {
		if strings.Contains(string(report), id) {
			t.Fatalf("ticket report disclosed unrelated unassigned session %q", id)
		}
	}
	local, err := s.List(ctx)
	if err != nil || len(local) != 5 {
		t.Fatalf("local discovery lost unassigned roots: %+v %v", local, err)
	}
}

func TestReconcileReportsCompletionAfterOwnerCheckpoint(t *testing.T) {
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	path := filepath.Join(store, "root.jsonl")
	writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	start := time.Now().UTC()
	result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, capture.DefaultByteBudget)
	end := time.Now().UTC()
	if err != nil {
		t.Fatal(err)
	}
	cutoff, err := time.Parse(time.RFC3339Nano, result.Cutoff)
	if err != nil {
		t.Fatal(err)
	}
	if cutoff.Before(start) || cutoff.After(end) {
		t.Fatalf("completion boundary outside actual pass: %s", result.Cutoff)
	}
	data, err := json.Marshal(result.Registrations[0])
	if err != nil {
		t.Fatal(err)
	}
	var checkpoint struct {
		RefreshedAt string `json:"refreshed_at"`
	}
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		t.Fatal(err)
	}
	refreshed, err := time.Parse(time.RFC3339Nano, checkpoint.RefreshedAt)
	if err != nil || refreshed.Before(start) || refreshed.After(cutoff) {
		t.Fatalf("owner read checkpoint %q is not bounded by truthful pass completion %q: %v", checkpoint.RefreshedAt, result.Cutoff, err)
	}
	started, err := time.Parse(time.RFC3339Nano, result.StartedAt)
	if err != nil || started.Before(start) || started.After(refreshed) || result.CutoffSemantics == "" {
		t.Fatalf("missing truthful pass start/completion semantics: %+v err=%v", result, err)
	}
}

func TestQueuedCodexChildRecoveryContinuesPastMalformedAndMismatchedMetadata(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	rootPath := filepath.Join(store, "root.jsonl")
	writeRollout(t, rootPath, "root", "", 10)
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	baseline, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.Sessions) != 1 {
		t.Fatal("fixture did not establish accepted root")
	}
	paths := []string{filepath.Join(store, "a-malformed.jsonl"), filepath.Join(store, "b-malformed.jsonl"), filepath.Join(store, "y-unrelated.jsonl"), filepath.Join(store, "z-valid.jsonl")}
	for _, path := range paths[:2] {
		if err := os.WriteFile(path, []byte("{\"type\":\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeRollout(t, paths[2], "unrelated-child", "different-parent", 99)
	writeRollout(t, paths[3], "actual-source-child", "root", 3)
	for _, path := range paths {
		if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SubagentStop", TranscriptPath: rootPath, AgentTranscriptPath: path}, store, "/binary", "/state"); err != nil {
			t.Fatal(err)
		}
	}
	recovered := false
	var latest capture.ReconcileResult
	for pass := 0; pass < 20; pass++ {
		latest, err = s.Reconcile(ctx, "TEST-100", "tester", 2, capture.DefaultByteBudget)
		if err != nil {
			t.Fatal(err)
		}
		if latest.SourceAttempts > 2 {
			t.Fatalf("recovery exceeded shared pass allowance: %+v", latest)
		}
		for _, row := range latest.Sessions {
			if row.ID == "unrelated-child" {
				t.Fatal("queued metadata with another source parent was bound")
			}
			if row.ID == "actual-source-child" && len(row.Invocations) == 1 {
				if row.ParentID == nil || *row.ParentID != "root" || row.Invocations[0].Tokens.Input != 3 {
					t.Fatalf("recovered child lost source proof/facts: %+v", row)
				}
				recovered = true
			}
		}
		if recovered {
			break
		}
	}
	if !recovered {
		t.Fatal("malformed or mismatched queued metadata permanently starved the valid child")
	}
	if latest.PendingChildren != 3 {
		t.Fatalf("unresolved metadata was discarded or successful child stayed pending: %d", latest.PendingChildren)
	}
}

func TestReconcileSchedulerEventuallyRefreshesOwnerWhenFailedMetadataConsumesByteAllowance(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	rootPath := filepath.Join(store, "root.jsonl")
	old := []byte(`{"type":"session_meta","payload":{"id":"root","timestamp":"2026-01-01T00:00:00Z"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"one","usage":{"input_tokens":10,"output_tokens":2}}}
`)
	if err := os.WriteFile(rootPath, old, 0600); err != nil {
		t.Fatal(err)
	}
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	before, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Sessions) != 1 || before.Sessions[0].Invocations[0].Tokens.Input != 10 {
		t.Fatal("fixture did not establish the old owner snapshot")
	}
	bad := filepath.Join(store, "a-malformed.jsonl")
	if err := os.WriteFile(bad, []byte("{"+strings.Repeat("x", 1023)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SubagentStop", TranscriptPath: rootPath, AgentTranscriptPath: bad}, store, "/binary", "/state"); err != nil {
		t.Fatal(err)
	}
	growing := append(append([]byte{}, old...), []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"two","usage":{"input_tokens":20,"output_tokens":3}}}`+"\n")...)
	if len(growing) != 509 {
		t.Fatalf("review schedule requires a 509-byte fitting owner, got %d", len(growing))
	}
	if err := os.WriteFile(rootPath, growing, 0600); err != nil {
		t.Fatal(err)
	}
	progress := false
	for pass := 0; pass < 8; pass++ {
		result, err := s.Reconcile(ctx, "TEST-100", "tester", 2, 1024)
		if err != nil {
			t.Fatal(err)
		}
		if result.SourceBytes > 1024 || result.SourceAttempts > 2 {
			t.Fatalf("scheduler exceeded unchanged pass budget: %+v", result)
		}
		if result.PendingChildren < 1 {
			t.Fatal("scheduler discarded unresolved malformed metadata to make progress")
		}
		for _, row := range result.Registrations {
			if row.Identity.SessionID == "root" && len(row.Snapshot) == 1 && len(row.Snapshot[0].Invocations) == 1 && row.Snapshot[0].Invocations[0].Tokens.Input == 30 {
				if row.RefreshedAt == before.Registrations[0].RefreshedAt {
					t.Fatal("new source facts retained an old successful-read checkpoint")
				}
				progress = true
			}
		}
		if progress {
			break
		}
	}
	if !progress {
		t.Fatal("failed metadata consumed all bytes on every bounded pass and starved a fitting valid owner")
	}
}

func TestReconcileSchedulerEventuallyDiscoversStartOnlyChildWithPersistentFailedStop(t *testing.T) {
	for _, withFailedStop := range []bool{true, false} {
		t.Run(fmt.Sprintf("failed stop %v", withFailedStop), func(t *testing.T) {
			ctx := context.Background()
			s := openStore(t, t.TempDir())
			store := t.TempDir()
			sessions := filepath.Join(store, "sessions")
			if err := os.MkdirAll(sessions, 0700); err != nil {
				t.Fatal(err)
			}
			rootPath := filepath.Join(sessions, "root.jsonl")
			writeRollout(t, rootPath, "root", "", 10)
			registerBound(t, s, store, rootPath, "root", "TEST-100")
			if _, err := s.Reconcile(ctx, "TEST-100", "tester", 10, capture.DefaultByteBudget); err != nil {
				t.Fatal(err)
			}
			if withFailedStop {
				bad := filepath.Join(sessions, "a-malformed.jsonl")
				if err := os.WriteFile(bad, []byte("{\"type\":\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SubagentStop", TranscriptPath: rootPath, AgentTranscriptPath: bad}, store, "/binary", "/state"); err != nil {
					t.Fatal(err)
				}
			}
			writeRollout(t, filepath.Join(sessions, "z-valid.jsonl"), "child-only-start", "root", 3)
			if _, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "root", Event: "SubagentStart", TranscriptPath: rootPath, AgentID: "opaque-child-interface"}, store, "/binary", "/state"); err != nil {
				t.Fatal(err)
			}
			progress := false
			for pass := 0; pass < 20; pass++ {
				result, err := s.Reconcile(ctx, "TEST-100", "tester", 2, capture.DefaultByteBudget)
				if err != nil {
					t.Fatal(err)
				}
				if result.SourceAttempts > 2 || result.SourceBytes > capture.DefaultByteBudget {
					t.Fatalf("cross-queue progress exceeded its unchanged allowance: %+v", result)
				}
				if withFailedStop && result.PendingChildren < 1 {
					t.Fatal("scheduler dropped failed stop evidence")
				}
				for _, row := range result.Sessions {
					if row.ID == "opaque-child-interface" {
						t.Fatal("scheduler substituted an interface ID for source identity")
					}
					if row.ID == "child-only-start" && len(row.Invocations) == 1 {
						if row.ParentID == nil || *row.ParentID != "root" || row.Invocations[0].Tokens.Input != 3 {
							t.Fatalf("start-only child lacks authoritative source facts: %+v", row)
						}
						progress = true
					}
				}
				if progress {
					break
				}
			}
			if !progress {
				t.Fatal("persistent stop work permanently starved source-proven start-only child discovery")
			}
		})
	}
}
