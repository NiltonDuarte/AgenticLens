package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestCodexDestinationInboxAuthenticatesPurposeTicketAndRevision(t *testing.T) {
	ctx := context.Background()
	project, home, state := t.TempDir(), protectedUserFixtureRoot(t), t.TempDir()
	configuration, err := fixtureSharedYAML(t, "Destination inbox", "records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), configuration, 0600); err != nil {
		t.Fatal(err)
	}
	policyBody := fmt.Sprintf("telemetry:\n  default_source: codex\n  capture:\n    enabled: true\n    state_dir: %q\n    max_sessions: 16\n    max_source_bytes: 67108864\n", state)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(policyBody), 0600); err != nil {
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
	inbox, err := CodexBindingInbox(home)
	if err != nil {
		t.Fatal(err)
	}
	policy := store.configuration.Ticket
	const target = "records/selected"
	const corrected = "records/corrected"
	ticket := BindingTicket{ID: "T-100", Title: "Explicit fixture ticket", Type: "fix"}

	lookup := func(id string) domain.CaptureRegistration {
		t.Helper()
		rows, err := store.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			if row.Identity.SessionID == id {
				return row
			}
		}
		t.Fatalf("could not find exact source registration %s", id)
		return domain.CaptureRegistration{}
	}
	handoff := func(id, event string) HookOutput {
		t.Helper()
		rollout := filepath.Join(home, "sessions", id+".jsonl")
		if event == "SessionStart" {
			body := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", id, project)
			if err := os.WriteFile(rollout, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
		input := HookInput{SessionID: id, WorkingDirectory: project, Event: event, TranscriptPath: rollout}
		reply, err := ReceiveCodexUser(ctx, store, input, home, manager, state)
		if err != nil {
			t.Fatal(err)
		}
		return reply
	}
	requirePrompt := func(reply HookOutput, credential, option string) {
		t.Helper()
		if reply.Context == nil || credential == "" || !strings.Contains(reply.Context.Context, credential) || !strings.Contains(reply.Context.Context, option) {
			t.Fatalf("native prompt did not present expected credential and %s", option)
		}
	}
	consume := func(row domain.CaptureRegistration, want int, failure string) {
		t.Helper()
		var cursor int
		got, err := store.ConsumeCodexBindings(ctx, home, []domain.CaptureRegistration{row}, &cursor, 16, 1<<20)
		if got != want || (failure == "" && err != nil) || (failure != "" && (err == nil || !strings.Contains(err.Error(), failure))) {
			t.Fatalf("consume %s: accepted=%d, want=%d, failure=%q, err=%v", row.Identity.SessionID, got, want, failure, err)
		}
	}
	envelope := func(path string) (bindingEnvelope, []byte) {
		t.Helper()
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var value bindingEnvelope
		if err := json.Unmarshal(body, &value); err != nil {
			t.Fatal(err)
		}
		return value, body
	}
	submitLate := func(id, token, ticketID, destination string) {
		t.Helper()
		if err := SubmitCodexDestination(project, home, id, token, BindingTicket{ID: ticketID, Destination: destination}, policy); err != nil {
			t.Fatal(err)
		}
	}

	const legacyID = "previous-inbox-v1"
	legacyPrompt := handoff(legacyID, "SessionStart")
	legacy := lookup(legacyID)
	initialRequest := inboxCredential(project, legacy)
	requirePrompt(legacyPrompt, initialRequest, "--ticket-title")
	if err := SubmitCodexBinding(project, home, legacyID, initialRequest, ticket, policy); err != nil {
		t.Fatal(err)
	}
	if got, _ := envelope(filepath.Join(inbox, slotName(initialRequest))); got.Version != 1 || got.Purpose != "" {
		t.Fatalf("initial binding without destination changed the v1 wire envelope: %+v", got)
	}
	consume(legacy, 1, "")
	legacy = lookup(legacyID)
	if legacy.TicketID != ticket.ID || legacy.TicketTitle != ticket.Title || legacy.TicketType != ticket.Type || legacy.PublicationDestination != "" {
		t.Fatalf("v1 binding with no destination was not preserved: %+v", legacy)
	}

	lateRequest := NativePublicationDestinationRequest(legacy)
	if lateRequest == "" {
		t.Fatalf("verified binding lacks a destination request: proof=%s binding=%q accepted=%t callback=%t", legacy.ProjectProof, legacy.Binding, legacy.BindingAcceptedAt != "", legacy.CallbackObservedAt != "")
	}
	latePrompt := handoff(legacyID, "UserPromptSubmit")
	legacy = lookup(legacyID)
	if current := NativePublicationDestinationRequest(legacy); current != lateRequest {
		t.Fatalf("prompt changed the request eligibility: proof=%s detail=%s", legacy.ProjectProof, legacy.Detail)
	}
	requirePrompt(latePrompt, lateRequest, "--destination-only")
	if initialRequest == lateRequest {
		t.Fatal("the initial ticket-binding credential authorized destination indication")
	}
	submitLate(legacyID, initialRequest, ticket.ID, target)
	consume(legacy, 0, "")
	submitLate(legacyID, lateRequest, "T-200", target)
	consume(legacy, 0, "exact bound ticket")
	if after := lookup(legacyID); after.PublicationDestination != "" || after.TicketID != ticket.ID || after.DestinationRevision != 0 {
		t.Fatalf("incorrect ticket indication changed the binding: %+v", after)
	}
	submitLate(legacyID, lateRequest, ticket.ID, target)
	slot := filepath.Join(inbox, destinationSlotName(lateRequest))
	got, signed := envelope(slot)
	if got.Version != 2 || got.Purpose != destinationIndication {
		t.Fatalf("late response did not use the bound destination protocol: %+v", got)
	}
	tampered := bytes.Replace(signed, []byte(target), []byte(corrected), 1)
	if err := os.WriteFile(slot, tampered, 0600); err != nil {
		t.Fatal(err)
	}
	consume(legacy, 0, "binding authentication failed")
	if after := lookup(legacyID); after.PublicationDestination != "" {
		t.Fatalf("tampering changed the destination: %+v", after)
	}
	submitLate(legacyID, lateRequest, ticket.ID, target)
	consume(legacy, 1, "")
	first := lookup(legacyID)
	correctionRequest := NativePublicationDestinationRequest(first)
	if first.DestinationRevision != 1 || first.PublicationDestination != target || correctionRequest == "" || correctionRequest == lateRequest {
		t.Fatalf("accepted destination did not retire its credential: %+v", first)
	}
	submitLate(legacyID, correctionRequest, ticket.ID, corrected)
	consume(first, 1, "")
	// A copy of a prior reader's row still selects the old signed slot.
	// Consumption must recheck the current revision inside the registry write.
	consume(legacy, 0, "exact bound ticket")
	if after := lookup(legacyID); after.PublicationDestination != corrected || after.DestinationRevision != 2 || after.TicketTitle != ticket.Title || after.TicketType != ticket.Type {
		t.Fatalf("stale replay changed the corrected destination or ticket metadata: %+v", after)
	}

	const freshID = "initial-inbox-v2"
	freshPrompt := handoff(freshID, "SessionStart")
	fresh := lookup(freshID)
	freshRequest := inboxCredential(project, fresh)
	requirePrompt(freshPrompt, freshRequest, "--destination-directory")
	ticket.Destination = target
	if err := SubmitCodexBinding(project, home, freshID, freshRequest, ticket, policy); err != nil {
		t.Fatal(err)
	}
	if got, _ := envelope(filepath.Join(inbox, slotName(freshRequest))); got.Version != 2 || got.Purpose != bindingWithDestination {
		t.Fatalf("initial binding with destination did not use its own v2 purpose: %+v", got)
	}
	consume(fresh, 1, "")
	if after := lookup(freshID); after.TicketID != ticket.ID || after.PublicationDestination != target || after.DestinationRevision != 0 {
		t.Fatalf("initial explicit hint and binding were not recorded together: %+v", after)
	}

	const partialID = "legacy-partial-binding"
	handoff(partialID, "SessionStart")
	partial := lookup(partialID)
	initialPartialRequest := inboxCredential(project, partial)
	typeless := BindingTicket{ID: ticket.ID, Title: ticket.Title}
	if err := SubmitCodexBinding(project, home, partialID, initialPartialRequest, typeless, policy); err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("new typeless native response did not fail validation: %v", err)
	}
	typelessSlot := filepath.Join(inbox, slotName(initialPartialRequest))
	if _, err := os.Lstat(typelessSlot); !os.IsNotExist(err) {
		t.Fatalf("invalid typeless response touched the inbox: %v", err)
	}
	// A response queued by the previous writer can pass authentication with
	// the v1 wire shape. It must still satisfy the new explicit-type contract.
	old := bindingEnvelope{Version: 1, Ticket: typeless}
	old.MAC = old.signature(initialPartialRequest, project, partial.Identity)
	legacyBody, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(typelessSlot, legacyBody, 0600); err != nil {
		t.Fatal(err)
	}
	consume(partial, 0, "type")
	if row := lookup(partialID); row.TicketID != "" {
		t.Fatalf("a legacy typeless message created a new binding: %+v", row)
	}

	// The historical direct bind allowed a durable ticket without a type.
	// A native response can complete the same ticket after upgrade.
	if _, err := store.Bind(ctx, partial.Identity, typeless.ID, partial.Request); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Change(ctx, partial.Identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketTitle = typeless.Title
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	partialPrompt := handoff(partialID, "UserPromptSubmit")
	partial = lookup(partialID)
	partialRequest := inboxCredential(project, partial)
	requirePrompt(partialPrompt, partialRequest, "--ticket-type")
	if NativePublicationDestinationRequest(partial) != "" {
		t.Fatal("the partial binding had a destination-only credential")
	}
	typeless.Type = "fix"
	otherTicket := typeless
	otherTicket.ID = "T-200"
	if err := SubmitCodexBinding(project, home, partialID, partialRequest, otherTicket, policy); err != nil {
		t.Fatal(err)
	}
	consume(partial, 0, "already bound")
	if err := SubmitCodexBinding(project, home, partialID, partialRequest, typeless, policy); err != nil {
		t.Fatal(err)
	}
	consume(partial, 1, "")
	if row := lookup(partialID); row.TicketID != typeless.ID || row.TicketTitle != typeless.Title || row.TicketType != "fix" || NativePublicationDestinationRequest(row) == "" {
		t.Fatalf("completing the native binding lost its assigned ticket or destination request: %+v", row)
	}
}
