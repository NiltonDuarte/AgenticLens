package main_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

// FIX-21: an unassigned Codex rollout that moved and cannot be proven
// anywhere in its store is isolated. The worker completes its pass, other
// tickets still publish, and it exits when idle instead of retrying the row
// until its lifetime is exhausted.
func TestCaptureWorkerIsolatesUnrecoverableUnassignedRollout(t *testing.T) {
	const ticket = "TEST-100"
	f := newCaptureGenerationFixture(t)
	f.output = seedCaptureTicketHome(t, f, ticket)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.state)
	verifiedPath := prepareCaptureOwner(t, f, "verified", ticket, 29)
	loosePath := filepath.Join(f.store, "sessions", "loose.jsonl")
	writeLoose := func(tokens int) {
		t.Helper()
		body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"loose","cwd":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"loose","turn_id":"one","usage":{"input_tokens":%d}}}
`, f.project, tokens)
		if err := os.WriteFile(loosePath, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeLoose(13)
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := registry.Close(); err != nil {
			t.Error(err)
		}
	})
	ctx := context.Background()
	loose := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "loose"}
	if _, err = registry.Register(ctx, loose, loosePath, "SessionEnd"); err != nil {
		t.Fatal(err)
	}
	if err = registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if out, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state); err != nil {
		t.Fatalf("fixture worker pass: %v %s", err, out)
	}
	accepted := func() domain.CaptureRegistration {
		t.Helper()
		rows, err := registry.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if row.Identity.SessionID == "loose" {
				return row
			}
		}
		t.Fatal("unassigned row missing")
		return domain.CaptureRegistration{}
	}
	before := accepted()
	if before.TicketID != "" || len(before.Snapshot) != 1 {
		t.Fatalf("fixture must accept the unassigned rollout: %+v", before)
	}

	// Codex moved the rollout somewhere this store cannot prove it, and the
	// ticket owner has new source facts to publish.
	if err := os.Remove(loosePath); err != nil {
		t.Fatal(err)
	}
	body := strings.Replace(string(readCaptureFile(t, verifiedPath)), `"input_tokens":29`, `"input_tokens":31`, 1)
	if err := os.WriteFile(verifiedPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err = registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	// The harness allows three times the policy's 10 s lifetime, so only the
	// worker's own outcome decides: a worker that retried the isolated row
	// would end with its lifetime exhausted, however loaded the host is.
	out, err := runCaptureWorkerWithin(t, f.project, 30*time.Second, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	if err != nil || strings.Contains(string(out), "lifetime exhausted") {
		t.Fatalf("an isolated unassigned row failed the worker: %v\n%s", err, out)
	}
	assertCaptureRootInputFor(t, f.ledger(t), "verified", 31)
	got := accepted()
	if got.TicketID != "" || got.State != domain.CapturePending || got.TranscriptPath != loosePath ||
		!strings.Contains(got.Detail, "unavailable") || !strings.Contains(got.Detail, "isolated") {
		t.Fatalf("isolated row does not show its unavailable source: %+v", got)
	}
	if len(got.Snapshot) != 1 || got.Snapshot[0].Invocations[0].Tokens.Input != before.Snapshot[0].Invocations[0].Tokens.Input {
		t.Fatalf("isolated row lost its accepted facts: %+v", got.Snapshot)
	}
	// A later integration event retries it: the rollout reappears in the
	// archive and the next pass recovers it, still unassigned.
	archived := filepath.Join(f.store, "archived_sessions", "loose.jsonl")
	if err := os.MkdirAll(filepath.Dir(archived), 0700); err != nil {
		t.Fatal(err)
	}
	loosePath = archived
	writeLoose(17)
	if err = registry.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if out, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state); err != nil {
		t.Fatalf("retry pass: %v %s", err, out)
	}
	if recovered := accepted(); recovered.TicketID != "" || recovered.TranscriptPath != archived || recovered.Snapshot[0].Invocations[0].Tokens.Input != 17 {
		t.Fatalf("retried row not recovered from the archive: %+v", recovered)
	}
}

func assertCaptureRootInputFor(t *testing.T, ledger domain.AgenticFactLedger, id string, want int64) {
	t.Helper()
	for _, row := range ledger.Sessions {
		if row.ID == id {
			if len(row.Invocations) != 1 || row.Invocations[0].Tokens.Input != want {
				t.Fatalf("%s snapshot=%+v, want input %d", id, row, want)
			}
			return
		}
	}
	t.Fatalf("missing %s session", id)
}

// runCaptureWorkerWithin runs a native capture command like runCaptureNative,
// with an explicit harness timeout.
func runCaptureWorkerWithin(t *testing.T, cwd string, timeout time.Duration, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if len(args) > 1 && args[0] == "capture" {
		args = append([]string{"capture"}, prepareCaptureInstall(t, cwd, args[1:])...)
	}
	cmd := exec.CommandContext(ctx, cliBinaryPath, args...)
	cmd.Dir = cwd
	return cmd.CombinedOutput()
}
