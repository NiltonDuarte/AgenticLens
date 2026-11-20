package capture

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestCodexCaptureControlInboxAuthenticatesScopeAndIsolatesStaleReplays(t *testing.T) {
	ctx := context.Background()
	project, home, state := t.TempDir(), protectedUserFixtureRoot(t), t.TempDir()
	shared, err := fixtureSharedYAML(t, "Capture control inbox", "records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("telemetry:\n  default_source: codex\n  capture:\n    enabled: true\n    state_dir: %q\n    max_sessions: 16\n    max_source_bytes: 67108864\n", state)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_HOME", home)
	manager, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InstallUserCodex(ctx, UserInstallOptions{Home: home, Entrypoint: manager}); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(home, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	store, err := OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	register := func(id string) HookOutput {
		t.Helper()
		rollout := filepath.Join(home, "sessions", id+".jsonl")
		native := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q}}`+"\n"+
			`{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn"}}`+"\n"+
			`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn","usage":{"input_tokens":7}}}`+"\n", id, project, id)
		if err := os.WriteFile(rollout, []byte(native), 0600); err != nil {
			t.Fatal(err)
		}
		output, err := ReceiveCodexUser(ctx, store, HookInput{SessionID: id, Event: "SessionStart", WorkingDirectory: project, TranscriptPath: rollout}, home, manager, state)
		if err != nil {
			t.Fatal(err)
		}
		return output
	}
	native := register("controlled-root")
	register("other-root")
	rows, err := store.List(ctx)
	if err != nil || len(rows) != 2 {
		t.Fatalf("expected independently verified roots: %d, %v", len(rows), err)
	}
	var root domain.CaptureRegistration
	for _, row := range rows {
		if row.Identity.SessionID == "controlled-root" {
			root = row
		}
	}
	invitation := func(action string) NativeCaptureControl {
		t.Helper()
		request, err := store.CaptureControlInvitation(ctx, root)
		if err != nil || request == nil || request.Action != action {
			t.Fatalf("missing verified %s invitation: %v", action, err)
		}
		return *request
	}
	discard := invitation("discard")
	if native.Context == nil || !strings.Contains(native.Context.Context, discard.Request) || !strings.Contains(native.Context.Context, "--capture-action discard") {
		t.Fatal("native hook did not offer its own exact authenticated discard")
	}
	stateOf := func(id string, want CaptureControlState) CaptureControl {
		t.Helper()
		status, err := store.CaptureControl(ctx, id)
		if err != nil || status.State != want {
			t.Fatalf("capture control state = %q, want %q: %v", status.State, want, err)
		}
		return status
	}
	submit := func(id, action, request string) {
		t.Helper()
		if err := SubmitCodexCaptureControl(project, home, root.Identity.SessionID, id, action, request); err != nil {
			t.Fatal(err)
		}
	}
	consume := func(want int, failure string) {
		t.Helper()
		cursor := 0
		got, err := store.ConsumeCodexBindings(ctx, home, []domain.CaptureRegistration{root}, &cursor, 16, 1<<20)
		if got != want || (failure == "" && err != nil) || (failure != "" && (err == nil || !strings.Contains(err.Error(), failure))) {
			t.Fatalf("worker accepted %d, want %d, failure %q: %v", got, want, failure, err)
		}
	}
	other := StableCaptureID(domain.SourceCodex, "other-root")
	submit(other, discard.Action, discard.Request)
	consume(0, "invalid Codex capture control envelope")
	stateOf(other, CaptureControlActive)
	submit(discard.CaptureID, "restore", discard.Request)
	consume(0, "authentication failed")
	stateOf(discard.CaptureID, CaptureControlActive)

	submit(discard.CaptureID, discard.Action, discard.Request)
	consume(1, "")
	discarded := stateOf(discard.CaptureID, CaptureControlDiscarded)
	restore := invitation("restore")
	submit(restore.CaptureID, restore.Action, restore.Request)
	inbox, err := CodexBindingInbox(home)
	if err != nil {
		t.Fatal(err)
	}
	slot, _, revision, err := parseControlRequest(restore.Request)
	if err != nil || revision != discarded.Revision {
		t.Fatalf("restore receipt carries the wrong revision: %d, %v", revision, err)
	}
	currentSlot := filepath.Join(inbox, controlInboxName(slot))
	queued, err := os.ReadFile(currentSlot)
	if err != nil {
		t.Fatal(err)
	}
	submit(discard.CaptureID, discard.Action, discard.Request)
	if after, err := os.ReadFile(currentSlot); err != nil || !bytes.Equal(after, queued) {
		t.Fatalf("older admission overwrote the current queued restoration: %v", err)
	}
	consume(1, "")
	restored := stateOf(discard.CaptureID, CaptureControlRestoring)
	if restored.Revision <= discarded.Revision {
		t.Fatalf("restore did not advance revision: %d to %d", discarded.Revision, restored.Revision)
	}
	result, err := store.Reconcile(ctx, "", "fixture", 16, 64<<20, fixtureGenerationPolicy())
	if err != nil || len(result.RestoresReady) != 1 || result.RestoresReady[0].CaptureID != discard.CaptureID {
		t.Fatalf("exact native history did not provide restore coverage: %v", err)
	}
	if err := store.CompleteCaptureRestores(ctx, result.RestoresReady); err != nil {
		t.Fatal(err)
	}
	for _, expired := range []NativeCaptureControl{discard, restore} {
		submit(expired.CaptureID, expired.Action, expired.Request)
		consume(0, "")
		if active := stateOf(expired.CaptureID, CaptureControlActive); active.Revision != restored.Revision {
			t.Fatalf("older admission changed the restored capture revision: %d", active.Revision)
		}
	}
}
