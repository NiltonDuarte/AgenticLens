package capture_test

import (
	"context"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestCodexHookFeedbackBindsOnlyAnExplicitResponse(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	store := t.TempDir()
	input, err := capture.ReadHook(strings.NewReader(`{"session_id":"exact-root","hook_event_name":"SessionStart","model":"untrusted-model","ticket_id":"GUESS-123","parent_id":"invented","secret":"do-not-store"}`))
	if err != nil {
		t.Fatal(err)
	}
	response, err := capture.ReceiveCodex(ctx, s, input, store, "/release binary/agentic-lens", "/local state")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].TicketID != "" || rows[0].ParentID != "" {
		t.Fatalf("untrusted hook fields assigned ownership: %+v", rows)
	}
	if response.Context == nil || response.Context.Event != "SessionStart" {
		t.Fatalf("missing agent context: %+v", response)
	}
	prompt := response.Context.Context
	for _, want := range []string{"exact-root", rows[0].Request, "capture bind", "--session-id", "--store", "--ticket-id", "Do not infer"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("agent context missing %q: %s", want, prompt)
		}
	}
	if strings.Contains(prompt, "do-not-store") || strings.Contains(prompt, "GUESS-123") {
		t.Fatal("raw/untrusted payload leaked into agent context")
	}
	bound, err := s.Bind(ctx, rows[0].Identity, "TEST-100", rows[0].Request)
	if err != nil {
		t.Fatal(err)
	}
	if bound.TicketID != "TEST-100" {
		t.Fatalf("explicit response not accepted: %+v", bound)
	}
	repeated, err := capture.ReceiveCodex(ctx, s, input, store, "/release binary/agentic-lens", "/local state")
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Context != nil {
		t.Fatalf("bound session got another ticket request: %+v", repeated)
	}
}

func TestSubagentHookDoesNotInventChildIdentityFromParentSession(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	out, err := capture.ReceiveCodex(ctx, s, capture.HookInput{SessionID: "parent", Event: "SubagentStart", AgentID: "unproven-agent-id"}, t.TempDir(), "/binary", "/state")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Identity.SessionID != "parent" || rows[0].ParentID != "" || rows[0].TicketID != "" {
		t.Fatalf("fabricated child association: %+v", rows)
	}
	if out.Context == nil || !strings.Contains(out.Context.Context, "Do not bind using the parent's ID") {
		t.Fatalf("child lacks exact-identity warning: %+v", out)
	}
}

func TestHookInputBoundsAndUnsupportedEvent(t *testing.T) {
	for _, payload := range []string{"{", `{}`, `{"session_id":"root"}`, strings.Repeat(" ", capture.MaxPayloadBytes+1)} {
		if _, err := capture.ReadHook(strings.NewReader(payload)); err == nil {
			t.Fatalf("accepted invalid/oversized payload %d bytes", len(payload))
		}
	}
	s := openStore(t, t.TempDir())
	if _, err := capture.ReceiveCodex(context.Background(), s, capture.HookInput{SessionID: "root", Event: "InventedEvent"}, t.TempDir(), "/binary", "/state"); err == nil {
		t.Fatal("accepted invented lifecycle event")
	}
	rows, err := s.List(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("invalid callback persisted: rows=%+v err=%v", rows, err)
	}
}
