package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	t.Run("publication", func(t *testing.T) { verifyLateCodexSubmission(t, 1) })
	t.Run("one slot per pass across worker lifetimes", func(t *testing.T) { verifyLateCodexSubmission(t, 12) })
}

func verifyLateCodexSubmission(t *testing.T, ownerCount int) {
	t.Helper()
	f := newCodexInboxFixture(t)
	prepareCaptureInstall(t, f.project, []string{"install", "--repo", f.project})
	enableCLIFixtureSummary(t, f.project)
	ledgerPath := seedCaptureTicketHome(t, f, "TEST-100")
	summary := filepath.Join(filepath.Dir(ledgerPath), "SUMMARY.md")
	const authored = "Authored delivery evidence\n"
	if err := os.WriteFile(summary, []byte(authored), 0600); err != nil {
		t.Fatal(err)
	}
	codexInboxOperation(t, f.project, "install-user")
	// Pick the last registration in the store's order so one worker pass
	// (one session) cannot reach this submission from the beginning.
	var owners []domain.CaptureIdentity
	for i := 0; i < ownerCount; i++ {
		owners = append(owners, domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: fmt.Sprintf("owner-%02d", i)})
	}
	slices.SortFunc(owners, func(a, b domain.CaptureIdentity) int {
		return strings.Compare(capture.IdentityKey(a), capture.IdentityKey(b))
	})
	target := owners[len(owners)-1]
	template := readCaptureFile(t, f.write(t, 73))
	transcript := func(id string) string {
		t.Helper()
		path := filepath.Join(f.store, "sessions", id+".jsonl")
		if err := os.WriteFile(path, bytes.ReplaceAll(template, []byte(`"root"`), []byte(fmt.Sprintf("%q", id))), 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	rollout := transcript(target.SessionID)
	receive := func(event string) capture.HookOutput {
		t.Helper()
		hookInput := capture.HookInput{WorkingDirectory: f.project, Event: event, SessionID: target.SessionID, TranscriptPath: rollout}
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
	// Wait for the initial callback to finish its first publication pass.
	// A user can answer the request after the configured 300 ms idle interval.
	ackDeadline := time.Now().Add(5 * time.Second)
	for {
		progress, err := store.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if progress.Generation > 0 && progress.Generation == progress.Acknowledged {
			lock, err := capture.AcquireWorker(store.CoordinationDirectory())
			if err != nil {
				t.Fatal(err)
			}
			if lock != nil {
				if err := lock.Close(); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		if time.Now().After(ackDeadline) {
			t.Fatalf("worker did not acknowledge initial callback and release its lease: %+v", progress)
		}
		time.Sleep(25 * time.Millisecond)
	}
	// The callback worker has exited. Queue earlier unassigned registrations
	// without waking a worker. At 100 ms per pass, the 12th slot also falls
	// beyond the 700 ms lifetime of the first worker launched by PostToolUse.
	if ownerCount > 1 {
		if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    max_sessions: 1\n    max_lifetime_ms: 700\n")); err != nil {
			t.Fatal(err)
		}
	}
	for _, owner := range owners[:len(owners)-1] {
		if _, err := capture.ReceiveCodexUser(context.Background(), store, capture.HookInput{SessionID: owner.SessionID, Event: "SessionStart", WorkingDirectory: f.project, TranscriptPath: transcript(owner.SessionID)}, f.store, cliBinaryPath, f.state); err != nil {
			t.Fatal(err)
		}
	}
	args := []string{"submit-binding", "--repo", f.project, "--store", f.store, "--session-id", target.SessionID, "--ticket-id", "TEST-100", "--ticket-title", "Inbox acceptance", "--ticket-type", "fix", "--request", inboxCredentialFromHook(t, response)}
	if out, err := captureCLI(t, f.project, "", args...); err != nil {
		t.Fatalf("sandbox submission failed: %s %v", out, err)
	}
	// Model the PostToolUse payload generated by Codex after the Bash call.
	postToolUse := func(command string) {
		t.Helper()
		event := struct {
			capture.HookInput
			ToolName  string `json:"tool_name"`
			ToolInput struct {
				Command string `json:"command"`
			} `json:"tool_input"`
			ToolResponse struct {
				ExitCode int `json:"exit_code"`
			} `json:"tool_response"`
		}{HookInput: capture.HookInput{WorkingDirectory: f.project, Event: "PostToolUse", SessionID: target.SessionID, TranscriptPath: rollout}, ToolName: "Bash"}
		event.ToolInput.Command = command
		data, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(managerBinaryPath, "codex-binding-wake")
		cmd.Dir = f.project
		cmd.Stdin = bytes.NewReader(data)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("native wake callback failed: %s %v", out, err)
		}
	}
	before, err := store.PublicationRequest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	postToolUse("git status --short")
	if after, err := store.PublicationRequest(context.Background()); err != nil || after != before {
		t.Fatalf("unrelated tool requested publication: %+v %v", after, err)
	}
	_, prompt, ok := strings.Cut(response.Context.Context, "execute: ")
	if !ok {
		t.Fatal("missing native submission command")
	}
	prefix, _, ok := strings.Cut(prompt, " --ticket-id ")
	if !ok {
		t.Fatal("missing native ticket arguments")
	}
	postToolUse(prefix + " --ticket-id TEST-100 --ticket-title 'Inbox acceptance' --ticket-type fix")
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	for ctx.Err() == nil {
		rows, err := store.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		body := readCaptureFile(t, ledgerPath)
		var ledger domain.AgenticFactLedger
		var row domain.CaptureRegistration
		for _, candidate := range rows {
			if candidate.Identity == target {
				row = candidate
			} else if candidate.TicketID != "" {
				t.Fatal("bound an unrelated registration")
			}
		}
		if len(rows) == len(owners) && row.TicketID == "TEST-100" && row.TicketTitle == "Inbox acceptance" && row.TicketType == "fix" {
			if ownerCount > 1 {
				// The one-session budget exercises delivery of the binding. A
				// complete scan of these twelve rollouts for ledger generation
				// needs a separate, larger source budget.
				return
			}
			if json.Unmarshal(body, &ledger) != nil || len(ledger.Sessions) != 1 {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			if ledger.TicketTitle == nil || *ledger.TicketTitle != "Inbox acceptance" || ledger.TicketType == nil || string(*ledger.TicketType) != "fix" || ledger.Sessions[0].ID != target.SessionID {
				t.Fatalf("published ledger does not match the authenticated submission: %+v", ledger)
			}
			if got := ledger.Sessions[0].Invocations; len(got) != 1 || got[0].Tokens.Input != 73 {
				t.Fatalf("missing source-backed usage: %+v", got)
			}
			rendered := readCaptureFile(t, summary)
			if bytes.Contains(rendered, []byte(authored)) && bytes.Contains(rendered, []byte("<!-- agentic-lens:metrics:start -->")) {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	rows, listErr := store.List(context.Background())
	for _, row := range rows {
		if row.Identity == target {
			t.Logf("target binding=%q ticket=%q state=%q", row.Binding, row.TicketID, row.State)
		}
	}
	prog, statusErr := capture.WorkerStatus(store.CoordinationDirectory())
	gen, genErr := store.PublicationRequest(context.Background())
	t.Fatalf("worker did not publish queued owner: progress=%+v publication=%+v readErrors=(%v,%v,%v)", prog, gen, listErr, statusErr, genErr)
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
	// An existing installation can outlive its project's local settings.
	// Submission must recheck separation before writing any response.
	localPath := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	original := readCaptureFile(t, localPath)
	overlap := bytes.Replace(original, []byte(f.stateBase), []byte(filepath.Join(inbox, "changed-state")), 1)
	if bytes.Equal(original, overlap) {
		t.Fatal("fixture did not change its capture state")
	}
	if err := os.WriteFile(localPath, overlap, 0600); err != nil {
		t.Fatal(err)
	}
	out, submitErr := captureCLI(t, f.project, "", "submit-binding", "--repo", f.project, "--store", f.store, "--session-id", "root", "--request", rootKey, "--ticket-id", "TEST-100", "--ticket-title", "Root")
	if submitErr == nil || !strings.Contains(strings.ToLower(string(out)), "overlap") {
		t.Fatalf("stale settings allowed submission or failed for the wrong reason: %s %v", out, submitErr)
	}
	if files, err := os.ReadDir(inbox); err != nil || len(files) != 0 {
		t.Fatalf("submission touched unavailable inbox: %v %v", files, err)
	}
	if err := os.WriteFile(localPath, original, 0600); err != nil {
		t.Fatal(err)
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

func TestOpenCodePublishesWithUnavailableCodexInbox(t *testing.T) {
	f := newCodexInboxFixture(t)
	ledger := seedCaptureTicketHome(t, f, "TEST-200")
	dbPath := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", dbPath)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("OPENCODE_CONFIG_DIR", t.TempDir())
	local := fmt.Sprintf("telemetry:\n  capture:\n  sources:\n    codex:\n      home: %q\n    opencode:\n      database_path: %q\n", f.store, dbPath)
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO session (id,agent,model,time_created,time_updated) VALUES ('opencode-root','root','source-model',1767225600000,1767225602000);
INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('m1','opencode-root',1767225601000,1767225602000,'{"role":"assistant","modelID":"source-model","time":{"created":1767225601000,"completed":1767225602000},"tokens":{"input":43,"output":2}}')`); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, dbPath, f.project)
	codexInboxOperation(t, f.project, "install-user")
	if out, err := captureCLI(t, f.project, "", "install-user", "--source", "opencode", "--repo", f.project, "--entrypoint", managerBinaryPath); err != nil {
		t.Fatalf("OpenCode fixture install: %s %v", out, err)
	}
	if err := os.Remove(filepath.Join(f.store, "agenticlens-binding-inbox")); err != nil {
		t.Fatal(err)
	}
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	completion := captureWorkerCompletion(t, registry.CoordinationDirectory())
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	payload, err := json.Marshal(capture.HookInput{WorkingDirectory: f.project, SessionID: "opencode-root", Event: "session.idle"})
	if err != nil {
		t.Fatal(err)
	}
	call := func(args ...string) []byte {
		t.Helper()
		out, err := captureCLI(t, f.project, string(payload), append([]string{"opencode-hook"}, args...)...)
		if err != nil {
			t.Fatalf("OpenCode callback %s failed: %s, %v", args[0], out, err)
		}
		return out
	}
	var reg domain.CaptureRegistration
	if err := json.Unmarshal(call("receive"), &reg); err != nil || reg.Request == "" || reg.Identity.SessionID != "opencode-root" {
		t.Fatalf("OpenCode receive did not return verified owner: %v", err)
	}
	call("bind", "--ticket-id", "TEST-200", "--ticket-title", "Independent OpenCode capture", "--ticket-type", "fix", "--request", reg.Request)
	waitCorrectionLedger(t, ledger, "opencode-root", 43)
}

func TestCodexInboxPassSharesBytesAcrossOwnerProofs(t *testing.T) {
	f := newCodexInboxFixture(t)
	codexInboxOperation(t, f.project, "install-user")
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	inbox := filepath.Join(f.store, "agenticlens-binding-inbox")
	// Each identically sized file contains only its exact session metadata.
	var sourceBytes int
	for _, id := range []string{"owner-A", "owner-B"} {
		source := []byte(fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", id, f.project))
		path := filepath.Join(f.store, "sessions", id+".jsonl")
		if err := os.WriteFile(path, source, 0600); err != nil {
			t.Fatal(err)
		}
		reply, err := capture.ReceiveCodexUser(ctx, s, capture.HookInput{SessionID: id, Event: "SessionStart", TranscriptPath: path, WorkingDirectory: f.project}, f.store, cliBinaryPath, f.state)
		if err != nil {
			t.Fatal(err)
		}
		out, err := captureCLI(t, f.project, "", "submit-binding", "--repo", f.project, "--store", f.store, "--session-id", id, "--request", inboxCredentialFromHook(t, reply), "--ticket-id", "TEST-100", "--ticket-title", id)
		if err != nil {
			t.Fatalf("submission for %s: %s %v", id, out, err)
		}
		sourceBytes = len(source)
	}
	slots, err := os.ReadDir(inbox)
	if err != nil || len(slots) != 2 {
		t.Fatalf("expected two submitted records: %v %v", slots, err)
	}
	images := map[string][]byte{}
	allowance := int64(sourceBytes)
	for _, slot := range slots {
		path := filepath.Join(inbox, slot.Name())
		images[path] = readCaptureFile(t, path)
		allowance += int64(len(images[path]))
	}
	extra := filepath.Join(inbox, "operator-note")
	images[extra] = []byte("pending operator work")
	if err := os.WriteFile(extra, images[extra], 0600); err != nil {
		t.Fatal(err)
	}
	cursor := 0
	pass := func() (int, error) {
		rows, err := s.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return s.ConsumeCodexBindings(ctx, f.store, rows, &cursor, 2, allowance)
	}
	if accepted, _ := pass(); accepted != 1 {
		t.Fatalf("two envelopes plus one owner's source must bind exactly one owner; accepted %d", accepted)
	}
	if accepted, err := pass(); err != nil || accepted != 1 {
		t.Fatalf("remaining owner did not complete next pass: %d %v", accepted, err)
	}
	if accepted, err := pass(); err != nil || accepted != 0 {
		t.Fatalf("reprocessed a completed owner: %d %v", accepted, err)
	}
	for path, want := range images {
		if got := readCaptureFile(t, path); !bytes.Equal(got, want) {
			t.Fatalf("consumer changed inbox entry %s", filepath.Base(path))
		}
	}
}
