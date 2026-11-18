package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func inboxCredentialFromHook(t *testing.T, reply capture.HookOutput) string {
	t.Helper()
	if reply.Context == nil || !strings.Contains(reply.Context.Context, "capture submit-binding") {
		t.Fatal("user callback did not supply the sandbox-accessible submission command")
	}
	match := regexp.MustCompile(`--request ['"]?([A-Za-z0-9_+./=-]+)`).FindStringSubmatch(reply.Context.Context)
	if len(match) != 2 {
		t.Fatal("missing response credential in native hook context")
	}
	return match[1]
}

// Exercise hook feedback, the public sandbox command and the autonomous worker
// as separate processes. Only the harness-emitted rollout is a fixture.
func TestCodexInboxWorkerPublishesAuthenticatedBinding(t *testing.T) {
	f := newCodexInboxFixture(t)
	enableCLIFixtureSummary(t, f.project)
	ledgerPath := seedCaptureTicketHome(t, f, "TEST-100")
	summary := filepath.Join(filepath.Dir(ledgerPath), "SUMMARY.md")
	const authored = "Authored delivery evidence\n"
	if err := os.WriteFile(summary, []byte(authored), 0600); err != nil {
		t.Fatal(err)
	}
	codexInboxOperation(t, f.project, "install-user")
	rollout := f.write(t, 73)
	receive := func(event string) capture.HookOutput {
		t.Helper()
		hookInput := capture.HookInput{WorkingDirectory: f.project, Event: event, SessionID: "root", TranscriptPath: rollout}
		payload, err := json.Marshal(hookInput)
		if err != nil {
			t.Fatal(err)
		}
		out, err := captureCLI(t, f.project, string(payload), "user-hook")
		if err != nil {
			t.Fatalf("native %s callback: %s %v", event, out, err)
		}
		var reply capture.HookOutput
		if err := json.Unmarshal(out, &reply); err != nil {
			t.Fatal(err)
		}
		return reply
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	response := receive("SessionStart")
	defer receive("SessionEnd")
	args := []string{"submit-binding", "--repo", f.project, "--store", f.store, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Inbox acceptance", "--ticket-type", "fix", "--request", inboxCredentialFromHook(t, response)}
	if out, err := captureCLI(t, f.project, "", args...); err != nil {
		t.Fatalf("sandbox submission failed: %s %v", out, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	for ctx.Err() == nil {
		rows, err := store.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		body := readCaptureFile(t, ledgerPath)
		var ledger domain.AgenticFactLedger
		if len(rows) == 1 && rows[0].TicketID == "TEST-100" && rows[0].TicketTitle == "Inbox acceptance" && rows[0].TicketType == "fix" && json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) == 1 {
			if ledger.TicketTitle == nil || *ledger.TicketTitle != "Inbox acceptance" || ledger.TicketType == nil || string(*ledger.TicketType) != "fix" || ledger.Sessions[0].ID != "root" {
				t.Fatalf("published ledger does not match the authenticated submission: %+v", ledger)
			}
			assertCaptureRootInput(t, ledger, 73)
			rendered := readCaptureFile(t, summary)
			if bytes.Contains(rendered, []byte(authored)) && bytes.Contains(rendered, []byte("<!-- agentic-lens:metrics:start -->")) {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("worker did not publish the queued authenticated owner to ledger and summary")
}

// Control consumption deterministically at the adapter boundary. The actual
// registry, registration proofs, filesystem and submission command are used.
func TestCodexInboxVerifiesExactOwnerAndRetriesAfterSourceFlush(t *testing.T) {
	f := newCodexInboxFixture(t)
	codexInboxOperation(t, f.project, "install-user")
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx := context.Background()
	inbox := filepath.Join(f.store, "agenticlens-binding-inbox")
	rollout := f.write(t, 73)
	observedSource := readCaptureFile(t, rollout)
	key := func(id, path string) string {
		t.Helper()
		out, err := capture.ReceiveCodexUser(ctx, store, capture.HookInput{SessionID: id, Event: "SessionStart", TranscriptPath: path, WorkingDirectory: f.project}, f.store, cliBinaryPath, f.state)
		if err != nil {
			t.Fatal(err)
		}
		return inboxCredentialFromHook(t, out)
	}
	rootKey := key("root", rollout)
	otherPath := filepath.Join(f.store, "sessions", "other.jsonl")
	if err := os.WriteFile(otherPath, bytes.ReplaceAll(observedSource, []byte(`"root"`), []byte(`"other"`)), 0600); err != nil {
		t.Fatal(err)
	}
	otherKey := key("other", otherPath)
	submit := func(id, credential, title string) {
		t.Helper()
		args := []string{"submit-binding", "--repo", f.project, "--store", f.store, "--session-id", id, "--request", credential, "--ticket-id", "TEST-100", "--ticket-title", title, "--ticket-type", "fix"}
		if out, err := captureCLI(t, f.project, "", args...); err != nil {
			t.Fatalf("submit: %s %v", out, err)
		}
	}
	beforeRows, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeRequest, err := store.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	submit("root", rootKey, "First owner")
	afterRows, err := store.List(ctx)
	if err != nil || !reflect.DeepEqual(beforeRows, afterRows) {
		t.Fatalf("sandbox submission modified private registration state: %v", err)
	}
	afterRequest, err := store.PublicationRequest(ctx)
	if err != nil || beforeRequest != afterRequest {
		t.Fatalf("sandbox submission modified trusted worker progress: %v", err)
	}
	entries, err := os.ReadDir(inbox)
	if err != nil || len(entries) != 1 {
		t.Fatalf("root slot: %v %v", entries, err)
	}
	rootSlot := filepath.Join(inbox, entries[0].Name())
	body := readCaptureFile(t, rootSlot)
	submit("other", otherKey, "Other owner")
	entries, err = os.ReadDir(inbox)
	if err != nil || len(entries) != 2 {
		t.Fatalf("independent slots: %v %v", entries, err)
	}
	var otherSlot string
	for _, e := range entries {
		p := filepath.Join(inbox, e.Name())
		if p != rootSlot {
			otherSlot = p
		}
	}
	rows := func() []domain.CaptureRegistration {
		t.Helper()
		r, err := store.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	for _, r := range rows() {
		if bytes.Contains(body, []byte(r.Request)) || strings.Contains(rootSlot, r.Request) {
			t.Fatal("inbox exposed the registry binding challenge")
		}
	}
	if bytes.Contains(body, []byte(rootKey)) || strings.Contains(rootSlot, rootKey) {
		t.Fatal("inbox exposed reusable submission credential")
	}
	consume := func(sessions int, budget int64, cursor *int) (int, error) {
		t.Helper()
		return store.ConsumeCodexBindings(ctx, f.store, rows(), cursor, sessions, budget)
	}
	for _, field := range [][2]string{{"TEST-100", "TEST-101"}, {"First owner", "Changed owner"}, {"fix", "feat"}} {
		tampered := bytes.Replace(body, []byte(field[0]), []byte(field[1]), 1)
		if bytes.Equal(tampered, body) {
			t.Fatalf("fixture lacks intended field %q", field[0])
		}
		if err := os.WriteFile(rootSlot, tampered, 0600); err != nil {
			t.Fatal(err)
		}
		// A valid envelope copied to another session's expected slot must fail.
		if err := os.WriteFile(otherSlot, body, 0600); err != nil {
			t.Fatal(err)
		}
		idx := 0
		if accepted, err := consume(2, 1<<20, &idx); err == nil || accepted != 0 {
			t.Fatalf("modified field %q or replay at another session bound: %d %v", field[0], accepted, err)
		}
		for _, r := range rows() {
			if r.TicketID != "" || r.TicketTitle != "" {
				t.Fatal("invalid envelope changed durable metadata")
			}
		}
	}
	if err := os.Remove(otherSlot); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(rootSlot); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(rootSlot, 0600); err != nil {
		t.Fatal(err)
	}
	cursor := 0
	if accepted, err := consume(2, 1<<20, &cursor); err == nil || accepted != 0 {
		t.Fatalf("special file admitted: %d %v", accepted, err)
	}
	if err := os.Remove(rootSlot); err != nil {
		t.Fatal(err)
	}
	submit("root", rootKey, "First owner")
	if accepted, err := consume(2, 1, &cursor); err == nil || accepted != 0 {
		t.Fatalf("response exceeded available byte budget: %d %v", accepted, err)
	}
	if err := os.Remove(rollout); err != nil {
		t.Fatal(err)
	}
	if accepted, err := consume(2, 1<<20, &cursor); err == nil || accepted != 0 || !bytes.Equal(body, readCaptureFile(t, rootSlot)) {
		t.Fatalf("unverified source consumed response: %d %v", accepted, err)
	}
	if err := os.WriteFile(rollout, observedSource, 0600); err != nil {
		t.Fatal(err)
	}
	idx := slices.IndexFunc(rows(), func(r domain.CaptureRegistration) bool { return r.Identity.SessionID == "other" })
	if idx < 0 {
		t.Fatal("missing independent owner")
	}
	if n, err := consume(1, 1<<20, &idx); err != nil || n != 0 {
		t.Fatalf("one-slot pass inspected another owner: %d %v", n, err)
	}
	if n, err := consume(1, 1<<20, &idx); err != nil || n != 1 {
		t.Fatalf("continued bounded pass failed to verify restored root: %d %v", n, err)
	}
	for _, r := range rows() {
		if r.Identity.SessionID == "root" && (r.TicketID != "TEST-100" || r.TicketTitle != "First owner" || r.TicketType != "fix") {
			t.Fatal("expected root metadata not committed atomically")
		}
		if r.Identity.SessionID == "other" && r.TicketID != "" {
			t.Fatal("unrelated registration was assigned")
		}
	}
	submit("root", rootKey, "First owner")
	if n, err := consume(2, 1<<20, &cursor); err != nil || n != 0 {
		t.Fatalf("replay was not idempotent: %d %v", n, err)
	}
}
