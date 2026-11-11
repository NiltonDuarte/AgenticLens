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
	adaptersconfig "agenticlens/internal/adapters/config"
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

func TestReconcileReplacesCompleteRestoredSnapshotAndPreservesMissingSource(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store, "sessions", "root.jsonl")
	original := writeRollout(t, path, "root", "", 10)
	row := registerBound(t, s, store, path, "root", "TEST-100")
	reconcile := func() capture.ReconcileResult {
		t.Helper()
		result, err := s.Reconcile(ctx, "TEST-100", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy())
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
	if len(truncated.Sessions) != 1 || truncated.Sessions[0].Invocations[0].Tokens.Input != 10 || truncated.Registrations[0].State == domain.CapturePending {
		t.Fatalf("BEH-02 complete restored smaller source was rejected as incomplete: %+v", truncated)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	missing := reconcile()
	if !reflect.DeepEqual(truncated.Sessions, missing.Sessions) || missing.Registrations[0].State != domain.CaptureFailed {
		t.Fatalf("missing source erased facts: %+v", missing)
	}
}

func TestReconcileVerifiesChildSourceAndIsolatesIndependentTickets(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	rootPath, childPath := filepath.Join(store, "sessions", "root.jsonl"), filepath.Join(store, "sessions", "child.jsonl")
	writeRollout(t, rootPath, "root", "", 10)
	writeRollout(t, childPath, "source-child", "root", 3)
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	if err := s.QueueChild(ctx, store, childPath, "root"); err != nil {
		t.Fatal(err)
	}
	otherPath := filepath.Join(store, "sessions", "other.jsonl")
	writeRollout(t, otherPath, "other", "", 7)
	registerBound(t, s, store, otherPath, "other", "TEST-200")
	unboundPath := filepath.Join(store, "sessions", "unbound.jsonl")
	writeRollout(t, unboundPath, "unbound", "", 9)
	if _, err := s.Register(ctx, domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: "unbound"}, unboundPath, "SessionStart"); err != nil {
		t.Fatal(err)
	}
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy())
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

func TestCoverageSummaryDoesNotPublishLocalPathsOrBindingToken(t *testing.T) {
	registration := domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/private-source-store", SessionID: "root"}, TicketID: "TEST-100", Request: "private-binding-token", TranscriptPath: "/private-transcript", Snapshot: []domain.SessionFact{{ID: "root"}}}
	result := capture.ReconcileResult{Protocol: 1, Cutoff: "2026-01-01T00:00:00Z", TicketID: "TEST-100", Registrations: []domain.CaptureRegistration{registration}}
	summary := capture.CoverageSummary(result)
	for _, forbidden := range []string{"private-source-store", "private-binding-token", "private-transcript", `"snapshot"`, "capture_report.json"} {
		if strings.Contains(summary, forbidden) {
			t.Fatalf("summary leaked obsolete/private detail %q: %s", forbidden, summary)
		}
	}
	if !strings.Contains(summary, "codex/root") || !strings.Contains(summary, "unobserved sessions") || !strings.Contains(summary, "checkpoint unknown") {
		t.Fatalf("summary omitted owner/coverage/freshness uncertainty: %s", summary)
	}
	if result.Registrations[0].Identity.Store != "/private-source-store" || result.Registrations[0].Request != "private-binding-token" {
		t.Fatal("summary mutated private registration")
	}
}

func TestReconcileChargesEachSourceOnceAndContinuesBudgetedWork(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store, "sessions", "root.jsonl")
	contents := writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 1, int64(len(contents)), fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Registrations) != 1 || result.Registrations[0].State != domain.CaptureCurrent {
		t.Fatalf("exact one-source budget was charged again: %+v", result)
	}
	for _, id := range []string{"second", "third"} {
		path := filepath.Join(store, "sessions", id+".jsonl")
		writeRollout(t, path, id, "", 3)
		registerBound(t, s, store, path, id, "TEST-100")
	}
	for i := 0; i < 3; i++ {
		if _, err := s.Reconcile(ctx, "TEST-100", "tester", 1, fixtureByteBudget, fixtureGenerationPolicy()); err != nil {
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
		if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(store, "sessions", "same-session.jsonl")
		writeRollout(t, path, "same-session", "", tokens)
		registerBound(t, s, store, path, "same-session", "TEST-100")
	}
	if result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy()); err == nil {
		t.Fatalf("canonical v5 cannot distinguish these owners: %+v", result.Sessions)
	}
}

func TestReconcilePreservesFactsWhenSourceParsingIsPartial(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store, "sessions", "root.jsonl")
	original := writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	before, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	partial := writeRollout(t, path, "root", "", 1)
	partial = append(partial, []byte(strings.Repeat(" ", len(original))+"{\"type\":\n")...)
	if err := os.WriteFile(path, partial, 0600); err != nil {
		t.Fatal(err)
	}
	after, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	var size int64
	for _, id := range []string{"owner-a", "owner-b"} {
		path := filepath.Join(store, "sessions", id+".jsonl")
		contents := append(writeRollout(t, path, id, "", 10), []byte("{\"type\":\n")...)
		if err := os.WriteFile(path, contents, 0600); err != nil {
			t.Fatal(err)
		}
		size = int64(len(contents))
		registerBound(t, s, store, path, id, "TEST-100")
	}
	result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 2, size, fixtureGenerationPolicy())
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
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
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	boundID := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store}
	for n := 0; ; n++ {
		boundID.SessionID = fmt.Sprintf("bound-%d", n)
		if strings.HasPrefix(capture.IdentityKey(boundID), "f") {
			break
		}
	}
	path := filepath.Join(store, "sessions", "bound.jsonl")
	contents := writeRollout(t, path, boundID.SessionID, "", 10)
	registerBound(t, s, store, path, boundID.SessionID, "TEST-100")
	unassigned := []string{}
	for n := 0; len(unassigned) < 4; n++ {
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: fmt.Sprintf("unassigned-%d", n)}
		if capture.IdentityKey(id) >= capture.IdentityKey(boundID) {
			continue
		}
		path := filepath.Join(store, "sessions", id.SessionID+".jsonl")
		writeRollout(t, path, id.SessionID, "", 99)
		if _, err := s.Register(ctx, id, path, "SessionStart"); err != nil {
			t.Fatal(err)
		}
		unassigned = append(unassigned, id.SessionID)
	}
	result, err := s.Reconcile(ctx, "TEST-100", "tester", 1, int64(len(contents)), fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].ID != boundID.SessionID || len(result.Registrations) != 1 || result.Registrations[0].State != domain.CaptureCurrent {
		t.Fatalf("unassigned roots starved ticket: %+v", result)
	}
	summary := capture.CoverageSummary(result)
	for _, id := range unassigned {
		if strings.Contains(summary, id) {
			t.Fatalf("ticket summary disclosed unrelated unassigned session %q", id)
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store, "sessions", "root.jsonl")
	writeRollout(t, path, "root", "", 10)
	registerBound(t, s, store, path, "root", "TEST-100")
	start := time.Now().UTC()
	result, err := s.Reconcile(context.Background(), "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(store, "sessions", "root.jsonl")
	writeRollout(t, rootPath, "root", "", 10)
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	baseline, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(baseline.Sessions) != 1 {
		t.Fatal("fixture did not establish accepted root")
	}
	paths := []string{filepath.Join(store, "sessions", "a-malformed.jsonl"), filepath.Join(store, "sessions", "b-malformed.jsonl"), filepath.Join(store, "sessions", "y-unrelated.jsonl"), filepath.Join(store, "sessions", "z-valid.jsonl")}
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
		latest, err = s.Reconcile(ctx, "TEST-100", "tester", 2, fixtureByteBudget, fixtureGenerationPolicy())
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
	if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	rootPath := filepath.Join(store, "sessions", "root.jsonl")
	old := []byte(`{"type":"session_meta","payload":{"id":"root","timestamp":"2026-01-01T00:00:00Z"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"one","usage":{"input_tokens":10,"output_tokens":2}}}
`)
	if err := os.WriteFile(rootPath, old, 0600); err != nil {
		t.Fatal(err)
	}
	registerBound(t, s, store, rootPath, "root", "TEST-100")
	before, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Sessions) != 1 || before.Sessions[0].Invocations[0].Tokens.Input != 10 {
		t.Fatal("fixture did not establish the old owner snapshot")
	}
	bad := filepath.Join(store, "sessions", "a-malformed.jsonl")
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
		result, err := s.Reconcile(ctx, "TEST-100", "tester", 2, 1024, fixtureGenerationPolicy())
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
			if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
				t.Fatal(err)
			}
			sessions := filepath.Join(store, "sessions")
			if err := os.MkdirAll(sessions, 0700); err != nil {
				t.Fatal(err)
			}
			rootPath := filepath.Join(sessions, "root.jsonl")
			writeRollout(t, rootPath, "root", "", 10)
			registerBound(t, s, store, rootPath, "root", "TEST-100")
			if _, err := s.Reconcile(ctx, "TEST-100", "tester", 10, fixtureByteBudget, fixtureGenerationPolicy()); err != nil {
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
				result, err := s.Reconcile(ctx, "TEST-100", "tester", 2, fixtureByteBudget, fixtureGenerationPolicy())
				if err != nil {
					t.Fatal(err)
				}
				if result.SourceAttempts > 2 || result.SourceBytes > fixtureByteBudget {
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

func TestReconcileFatalCollisionRetainsCollectionFailureUntilCompletedRecovery(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	captureProjectConfig(t, project, "ReconcileFixture")
	s, openErr := capture.OpenProject(t.TempDir(), project)
	if openErr != nil {
		t.Fatal(openErr)
	}
	t.Cleanup(func() {
		if e := s.Close(); e != nil {
			t.Error(e)
		}
	})
	writeProjectRollout := func(path string, tokens int) {
		body := writeRollout(t, path, "root", "", tokens)
		body = []byte(strings.Replace(string(body), `"id":"root"`, fmt.Sprintf(`"id":"root","cwd":%q`, project), 1))
		if e := os.WriteFile(path, body, 0600); e != nil {
			t.Fatal(e)
		}
	}
	firstStore := t.TempDir()
	if err := os.MkdirAll(filepath.Join(firstStore, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	firstPath := filepath.Join(firstStore, "sessions", "root.jsonl")
	writeProjectRollout(firstPath, 10)
	first := registerBound(t, s, firstStore, firstPath, "root", "TEST-100")
	_, err := s.Change(ctx, first.Identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.InstallationID = "install"
		r.BindingInstallationID = "install"
		return r, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := s.Reconcile(ctx, "TEST-100", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RecordPublication(ctx, "TEST-100", "committed", "metrics.json", "", baseline.Sessions, ""); err != nil {
		t.Fatal(err)
	}
	secondStore := t.TempDir()
	if err = os.MkdirAll(filepath.Join(secondStore, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	secondPath := filepath.Join(secondStore, "sessions", "root.jsonl")
	writeProjectRollout(secondPath, 30)
	second := registerBound(t, s, secondStore, secondPath, "root", "TEST-100")
	result, err := s.Reconcile(ctx, "TEST-100", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err == nil || !strings.Contains(err.Error(), "identity collision") {
		t.Fatalf("collision not reached: %+v %v", result, err)
	}
	t.Logf("actual rejected reconciliation: %v; result owners=%d", err, len(result.Registrations))
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var selected domain.CaptureRegistration
	for _, r := range rows {
		if r.Identity == first.Identity {
			selected = r
		}
	}
	page, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-100"}, ExactOwners: []capture.DiagnosticOwner{{TicketID: "TEST-100", Identity: first.Identity}}, Limit: 18, MaxBytes: fixtureByteBudget})
	if err != nil || !page.MetadataComplete || page.More {
		t.Fatalf("current evidence incomplete: %+v %v", page, err)
	}
	h := domain.CaptureVerification([]domain.CaptureRegistration{selected}, page.Records)
	issues := domain.CaptureAdvisoryIssues(domain.CaptureAdvisoryEvidence{Registrations: []domain.CaptureRegistration{selected}, Diagnostics: page.Records, Publications: []domain.CapturePublicationEvidence{{TicketID: "TEST-100", Available: true, Sessions: baseline.Sessions, PublishedAt: baseline.Cutoff}}, Caller: &first.Identity})
	t.Logf("current complete=%t more=%t advisory=%v", page.MetadataComplete, page.More, issues)
	feedback, err := s.SessionFailureFeedback(ctx, first.Identity, fixtureByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	failures := domain.CaptureCurrentFailures(selected, ds)
	t.Logf("health=%+v feedback=%q failures=%+v latest=%+v", h, feedback, failures, ds[len(ds)-1])
	if h.State != "degraded" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted || !failures.Collection || failures.Source || !strings.Contains(feedback, "whole-collection reconciliation failed") || len(issues) == 0 {
		t.Fatalf("fatal collection evidence lost or individual success mislabeled: %+v %q %+v %v", h, feedback, failures, issues)
	}
	assertCollection := func(wantFailed bool) {
		t.Helper()
		current, e := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-100"}, ExactOwners: []capture.DiagnosticOwner{{TicketID: "TEST-100", Identity: first.Identity}}, Limit: 24, MaxBytes: fixtureByteBudget})
		if e != nil || !current.MetadataComplete || current.More {
			t.Fatalf("current collection evidence unavailable: %+v %v", current, e)
		}
		f := domain.CaptureCurrentFailures(selected, current.Records)
		message, e := s.SessionFailureFeedback(ctx, first.Identity, fixtureByteBudget)
		health := domain.CaptureVerification([]domain.CaptureRegistration{selected}, current.Records)
		if e != nil || f.Collection != wantFailed || (health.State == "degraded") != wantFailed || strings.Contains(message, "whole-collection reconciliation failed") != wantFailed {
			t.Fatalf("collection recovery wrong: %+v %+v %q %v", f, health, message, e)
		}
	}
	// Individual successful reads and pending/partial passes cannot acknowledge
	// aggregate validation failure, even after recent-history churn.
	for i := 0; i < 40; i++ {
		if e := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{selected}}}); e != nil {
			t.Fatal(e)
		}
		if e := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "OTHER-1", Phase: "publication", Outcome: "failed"}); e != nil {
			t.Fatal(e)
		}
	}
	assertCollection(true)
	// Repair the fixture's explicit collection assignment; preserve both stored
	// owners/facts. This trusted Store mutation models operator state repair,
	// and does not add or claim a public session-rebinding command.
	if _, e := s.Change(ctx, second.Identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketID = "TEST-200"
		return r, nil
	}); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Reconcile(ctx, "TEST-100", "reviewer", 1, 1, fixtureGenerationPolicy()); e != nil {
		t.Fatal(e)
	}
	assertCollection(true)
	recovered, e := s.Reconcile(ctx, "TEST-100", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if e != nil || len(recovered.Sessions) != 1 || recovered.Sessions[0].ID != "root" {
		t.Fatalf("whole collection repair did not complete: %+v %v", recovered, e)
	}
	assertCollection(false)
}

func TestReconcileSuccessfulReadPreservesParentConflictAndNormalPendingRecovery(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	captureProjectConfig(t, project, "ReconcileFixture")
	s, err := capture.OpenProject(t.TempDir(), project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	store := t.TempDir()
	if err = os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	write := func(id, parent string) string {
		path := filepath.Join(store, "sessions", id+".jsonl")
		body := writeRollout(t, path, id, parent, 10)
		body = []byte(strings.Replace(string(body), fmt.Sprintf(`"id":%q`, id), fmt.Sprintf(`"id":%q,"cwd":%q`, id, project), 1))
		if err := os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	// Positive control: an unchanged successful read may normalize retained
	// lifecycle Pending, without changing accepted source facts or registration.
	normal := registerBound(t, s, store, write("normal", ""), "normal", "TEST-300")
	original, e := s.Reconcile(ctx, "TEST-300", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Register(ctx, normal.Identity, normal.TranscriptPath, "UserPromptSubmit"); e != nil {
		t.Fatal(e)
	}
	normalRead, e := s.Reconcile(ctx, "TEST-300", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if e != nil || len(normalRead.Registrations) != 1 || normalRead.Registrations[0].State != domain.CaptureCurrent || normalRead.SourceAttempts < 1 {
		t.Fatalf("complete unchanged-read control not reached: %+v %v", normalRead, e)
	}
	normalDiagnostics, e := s.Diagnostics(ctx)
	if e != nil {
		t.Fatal(e)
	}
	normalLast := normalDiagnostics[len(normalDiagnostics)-1]
	if normalLast.Reconciliation.Attempt != domain.CaptureAttemptCompleted || normalLast.Reconciliation.Owners[0].State != domain.CaptureCurrent || !reflect.DeepEqual(original.Sessions, normalRead.Sessions) {
		t.Fatalf("safe pending normalization regressed: %+v", normalLast)
	}
	// Establish actual inherited ownership, then accept a complete source refresh
	// naming a different, not-yet-registered parent. Its later explicit binding
	// can contradict inherited ownership; explicit child choices remain independent.
	registerBound(t, s, store, write("original-parent", ""), "original-parent", "TEST-200")
	childPath := write("child", "original-parent")
	if err = s.QueueChild(ctx, store, childPath, "original-parent"); err != nil {
		t.Fatal(err)
	}
	inherited, err := s.Reconcile(ctx, "TEST-200", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	findChild := func(owners []domain.CaptureRegistration) domain.CaptureRegistration {
		t.Helper()
		for _, owner := range owners {
			if owner.Identity.SessionID == "child" {
				return owner
			}
		}
		t.Fatal("source-proven child registration missing")
		return domain.CaptureRegistration{}
	}
	child := findChild(inherited.Registrations)
	if child.Binding != "source_parent" || child.TicketID != "TEST-200" || child.ParentID != "original-parent" || child.State != domain.CaptureCurrent {
		t.Fatalf("actual inherited control not reached: %+v", child)
	}
	write("child", "parent")
	if _, err = s.Register(ctx, child.Identity, childPath, "SessionStart"); err != nil {
		t.Fatal(err)
	}
	baseline, err := s.Reconcile(ctx, "TEST-200", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil || len(baseline.Registrations) != 2 {
		t.Fatalf("baseline unavailable: %+v %v", baseline, err)
	}
	baselineChild := findChild(baseline.Registrations)
	if baselineChild.State != domain.CaptureCurrent || baselineChild.Binding != "source_parent" || baselineChild.ParentID != "parent" || len(baselineChild.Snapshot) != 1 || baselineChild.Snapshot[0].ParentID == nil || *baselineChild.Snapshot[0].ParentID != "parent" {
		t.Fatalf("complete source parent refresh not accepted: %+v", baselineChild)
	}
	if err = s.RecordPublication(ctx, "TEST-200", "committed", "metrics.json", "", baseline.Sessions, ""); err != nil {
		t.Fatal(err)
	}
	parent := registerBound(t, s, store, write("parent", ""), "parent", "TEST-100")
	parentRead, err := s.Reconcile(ctx, "TEST-100", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil || len(parentRead.Registrations) != 1 {
		t.Fatal(err)
	}
	result, err := s.Reconcile(ctx, "TEST-200", "reviewer", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	resultChild := findChild(result.Registrations)
	if len(result.Registrations) != 2 || resultChild.State != domain.CaptureFailed || !strings.Contains(resultChild.Detail, "Conflicting source-proven parent") {
		t.Fatalf("did not reach supported parent conflict: %+v", result)
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var selected domain.CaptureRegistration
	for _, r := range rows {
		if r.Identity == child.Identity {
			selected = r
		}
		if r.Identity == parent.Identity && !reflect.DeepEqual(r, parentRead.Registrations[0]) {
			t.Fatal("child conflict mutated the independently bound parent")
		}
	}
	if selected.TicketID != baselineChild.TicketID || selected.Binding != baselineChild.Binding || selected.ParentID != baselineChild.ParentID || !reflect.DeepEqual(selected.Snapshot, baselineChild.Snapshot) {
		t.Fatal("child conflict changed inherited attribution or accepted facts")
	}
	page, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-200"}, ExactOwners: []capture.DiagnosticOwner{{TicketID: "TEST-200", Identity: child.Identity}}, Limit: 24, MaxBytes: fixtureByteBudget})
	if err != nil || !page.MetadataComplete || page.More {
		t.Fatalf("metadata incomplete: %+v %v", page, err)
	}
	h := domain.CaptureVerification([]domain.CaptureRegistration{selected}, page.Records)
	feedback, err := s.SessionFailureFeedback(ctx, child.Identity, fixtureByteBudget)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	latest := ds[len(ds)-1]
	issues := domain.CaptureAdvisoryIssues(domain.CaptureAdvisoryEvidence{Registrations: []domain.CaptureRegistration{selected}, Diagnostics: page.Records, Publications: []domain.CapturePublicationEvidence{{TicketID: "TEST-200", Available: true, Sessions: baseline.Sessions, PublishedAt: baseline.Cutoff}}, Caller: &child.Identity})
	if !reflect.DeepEqual(baseline.Sessions, result.Sessions) {
		t.Fatal("numeric/source facts changed")
	}
	if result.SourceAttempts < 1 || selected.RefreshedAt == baselineChild.RefreshedAt || selected.SourceFingerprint != baselineChild.SourceFingerprint {
		t.Fatal("conflict did not follow a successful unchanged read")
	}
	diagnosticChild := findChild(latest.Reconciliation.Owners)
	if latest.Outcome != "failed" || latest.Reconciliation.Attempt != domain.CaptureAttemptPartial || !strings.Contains(diagnosticChild.Detail, "Conflicting source-proven parent") {
		t.Fatalf("post-read validation was normalized or acknowledged: %+v", latest)
	}
	if !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
		t.Fatalf("historical exchange flags lost: %+v", h)
	}

	t.Logf("result state=%s detail=%q; stored state=%s; diagnostic outcome=%s attempt=%s owner state=%s; health=%+v feedback=%q advisory=%v", resultChild.State, resultChild.Detail, selected.State, latest.Outcome, latest.Reconciliation.Attempt, diagnosticChild.State, h, feedback, issues)
	if latest.Reconciliation.Attempt == domain.CaptureAttemptCompleted || diagnosticChild.State != domain.CaptureFailed || h.State != "degraded" || feedback == "" || len(issues) == 0 {
		t.Fatalf("inherited parent binding conflict rewritten to successful completion: health=%+v feedback=%q advisory=%v", h, feedback, issues)
	}
}

func TestReconcileExplicitChildTicketDoesNotBlockOrReassignParent(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	parentPath := filepath.Join(source, "sessions", "parent.jsonl")
	childPath := filepath.Join(source, "sessions", "child.jsonl")
	writeRollout(t, parentPath, "parent", "", 10)
	writeRollout(t, childPath, "child", "parent", 31)
	registerBound(t, s, source, parentPath, "parent", "TEST-100")
	registerBound(t, s, source, childPath, "child", "TEST-200")
	if err := s.QueueChild(ctx, source, childPath, "parent"); err != nil {
		t.Fatal(err)
	}
	parent, err := s.Reconcile(ctx, "TEST-100", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatalf("explicit child blocked parent: %v", err)
	}
	if len(parent.Sessions) != 1 || parent.Sessions[0].ID != "parent" || parent.Sessions[0].Invocations[0].Tokens.Input != 10 {
		t.Fatalf("child crossed into parent ticket: %+v", parent.Sessions)
	}
	child, err := s.Reconcile(ctx, "TEST-200", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatalf("child's own ticket blocked: %v", err)
	}
	if len(child.Sessions) != 1 || child.Sessions[0].ID != "child" || child.Sessions[0].ParentID == nil || *child.Sessions[0].ParentID != "parent" || child.Sessions[0].Invocations[0].Tokens.Input != 31 {
		t.Fatalf("explicit child identity/source ownership lost: %+v", child.Sessions)
	}
	if len(child.Registrations) != 1 || child.Registrations[0].State != domain.CaptureCurrent || child.Registrations[0].Binding != "explicit_agent_response" || child.Registrations[0].TicketID != "TEST-200" {
		t.Fatalf("explicit child choice became a parent conflict: %+v", child.Registrations)
	}
	diagnostics, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	latest := diagnostics[len(diagnostics)-1]
	if latest.TicketID != "TEST-200" || latest.Outcome != "accepted" || latest.Reconciliation == nil || latest.Reconciliation.Attempt != domain.CaptureAttemptCompleted || len(latest.Reconciliation.Owners) != 1 || latest.Reconciliation.Owners[0].State != domain.CaptureCurrent {
		t.Fatalf("explicit child reconciliation did not complete: %+v", latest)
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.Identity.SessionID == "child" && (row.TicketID != "TEST-200" || row.Binding != "explicit_agent_response" || row.ParentID != "parent") {
			t.Fatalf("explicit child binding rewritten: %+v", row)
		}
	}
	again, err := s.Reconcile(ctx, "TEST-100", "tester", 20, fixtureByteBudget, fixtureGenerationPolicy())
	if err != nil || len(again.Sessions) != 1 || again.Sessions[0].ID != "parent" {
		t.Fatalf("settled child work blocked/reassigned repeated parent: %+v %v", again.Sessions, err)
	}
}

func TestProjectReconciliationReceivesConfiguredAnomalyPolicy(t *testing.T) {
	project := t.TempDir()
	state := filepath.Join(project, "capture-state")
	data, err := fixtureSharedYAML(t, "ReconcilePolicy", "records")
	if err != nil {
		t.Fatal(err)
	}
	shared := filepath.Join(project, ".agenticlens-config.yaml")
	if err := os.WriteFile(shared, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    enabled: true\n    state_dir: capture-state\n    max_sessions: 256\n    max_source_bytes: 67108864\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := adaptersconfig.LoadShared(project)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AnomalyThresholds.ErrorRateMinCalls = 3
	cfg.AnomalyThresholds.ErrorRateThreshold = .40
	cfg.AnomalyThresholds.ErrorRateCriticalThreshold = .80
	cfg.AnomalyThresholds.ConsecutiveErrorsStreakWarning = 2
	cfg.AnomalyThresholds.ConsecutiveErrorsStreakCritical = 2
	data, err = adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shared, data, 0600); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "sessions", "source.jsonl")
	source := writeRollout(t, path, "policy-owner", "", 5)
	source = []byte(strings.Replace(string(source), `"payload":{"id":`, fmt.Sprintf(`"payload":{"cwd":%q,"id":`, project), 1))
	for i, exit := range []int{1, 1, 0} {
		source = append(source, []byte(fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:0%dZ","type":"event_msg","payload":{"type":"exec_command_end","call_id":"call-%d","turn_id":"turn-one","exit_code":%d}}`+"\n", i+3, i, exit))...)
	}
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	registry, err := capture.OpenProjectContext(context.Background(), state, project)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	registerBound(t, registry, home, path, "policy-owner", "TEST-100")
	result, err := registry.Reconcile(context.Background(), "TEST-100", "tester", 10, fixtureByteBudget, domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].Anomalies == nil {
		t.Fatalf("configured analysis was not recorded: %+v", result)
	}
	found := false
	for _, a := range result.Sessions[0].Anomalies {
		found = found || (a.RuleID == domain.RuleToolRetryStreak && a.Severity == domain.SeverityCritical && a.Context.Count != nil && *a.Context.Count == 2)
	}
	if !found {
		t.Fatalf("reconciliation lost configured two-error critical rule: %+v", result.Sessions[0].Anomalies)
	}
}
