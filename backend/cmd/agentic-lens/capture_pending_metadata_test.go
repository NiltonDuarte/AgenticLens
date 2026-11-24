package main_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCapturePendingRootsRetainTheirMetadataWhenTicketTypesConflict(t *testing.T) {
	verifyPendingMetadataConflict(t, "fix", "conflicting explicit types")
}

func TestCapturePendingRootsRefreshWhenOnlyTicketTitlesConflict(t *testing.T) {
	verifyPendingMetadataConflict(t, "feat", "conflicting explicit titles")
}

func verifyPendingMetadataConflict(t *testing.T, secondKind, failure string) {
	t.Helper()
	fixture := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, fixture.project, fixture.stateBase)
	fixture.output = initialPublicationTarget(t, fixture, "selected")
	prepareCaptureOwner(t, fixture, "first", initialPublicationTicket, 17)
	prepareCaptureOwner(t, fixture, "second", initialPublicationTicket, 29)

	pendingPattern := filepath.Join(fixture.project, captureFixturePendingDir, "*.json")
	var priorPaths []string
	secondInput := int64(29)
	if secondKind == "feat" {
		// The roots already have pending facts when the prepared destination
		// and differing title arrive. Reconciliation must still refresh them.
		if output, err := runInitialPublicationWorker(t, fixture); err != nil {
			t.Fatalf("cannot capture the prior pending facts: %v\n%s", err, output)
		}
		var err error
		priorPaths, err = filepath.Glob(pendingPattern)
		if err != nil || len(priorPaths) != 2 {
			t.Fatalf("expected two prior pending snapshots: %v %v", priorPaths, err)
		}
		secondInput = 31
		prepareCaptureOwner(t, fixture, "second", initialPublicationTicket, int(secondInput))
	}
	prepareInitialPublication(t, fixture, fixture.output)

	store, err := capture.OpenProject(fixture.state, fixture.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: fixture.store, SessionID: "second"}
	if _, err := store.Change(context.Background(), identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketTitle = "Second accepted native title"
		row.TicketType = secondKind
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}

	local := filepath.Join(fixture.project, ".agenticlens-config.local.yaml")
	policy := strings.Replace(string(readCaptureFile(t, local)), "max_lifetime_ms: 10000", "max_lifetime_ms: 1500", 1)
	if err := writeCaptureFileAtomic(local, []byte(policy), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := runInitialPublicationWorker(t, fixture)
	if err == nil || !strings.Contains(string(output), failure) {
		t.Fatalf("the worker did not report blocked publication: %v\n%s", err, output)
	}
	assertInitialPublicationAbsent(t, fixture.output, fixture.output+".publication-pending", filepath.Join(filepath.Dir(fixture.output), "SUMMARY.md"))

	paths, err := filepath.Glob(pendingPattern)
	if err != nil || len(paths) != 2 {
		t.Fatalf("expected two pending roots, got paths %v, error %v", paths, err)
	}
	if priorPaths != nil && !reflect.DeepEqual(paths, priorPaths) {
		t.Fatalf("refresh changed pending identity: before=%v after=%v", priorPaths, paths)
	}
	expected := map[string]struct {
		title string
		kind  string
		input int64
	}{
		"first":  {title: "Explicit " + initialPublicationTicket, kind: "feat", input: 17},
		"second": {title: "Second accepted native title", kind: secondKind, input: secondInput},
	}
	for _, path := range paths {
		var snapshot adapters.PendingSnapshot
		if err := json.Unmarshal(readCaptureFile(t, path), &snapshot); err != nil {
			t.Fatal(err)
		}
		metadata, expectedRoot := expected[snapshot.RootSessionID]
		if !expectedRoot || snapshot.TicketID != initialPublicationTicket || snapshot.TicketTitle != metadata.title ||
			snapshot.TicketType != metadata.kind || snapshot.State != "awaiting_destination" ||
			len(snapshot.Sessions) != 1 || snapshot.Sessions[0].ID != snapshot.RootSessionID ||
			len(snapshot.Sessions[0].Invocations) != 1 || snapshot.Sessions[0].Invocations[0].Tokens.Input != metadata.input {
			t.Fatalf("pending root lost its own binding or metrics: %+v", snapshot)
		}
		delete(expected, snapshot.RootSessionID)
	}
}

func TestCaptureCanonicalMetadataSurvivesConflictingNativeTicketTypes(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.stateBase)
	f.output = initialPublicationTarget(t, f, "existing-ledger")
	const title = "Accepted canonical title"
	var imported domain.AgenticFactLedger
	if err := json.Unmarshal([]byte(sampleFactJSON(initialPublicationTicket, title, 18, 0)), &imported); err != nil {
		t.Fatal(err)
	}
	kind := domain.TicketTypeDocs
	imported.TicketType = &kind
	data, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(imported))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.output, data, 0600); err != nil {
		t.Fatal(err)
	}
	accepted := f.ledger(t)

	prepareCaptureOwner(t, f, "first", initialPublicationTicket, 17)
	prepareCaptureOwner(t, f, "second", initialPublicationTicket, 29)
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	second := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "second"}
	if _, err := store.Change(context.Background(), second, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketTitle, row.TicketType = "Other native title", "fix"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("metadata conflict blocked refresh of the unique existing ledger: %v\n%s", err, output)
	}

	got := f.ledger(t)
	if got.TicketID == nil || *got.TicketID != initialPublicationTicket || got.TicketTitle == nil || *got.TicketTitle != title ||
		got.TicketType == nil || *got.TicketType != kind || len(got.Sessions) != len(accepted.Sessions)+2 {
		t.Fatalf("refresh changed accepted metadata or lost owners: %+v", got)
	}
	sessions := make(map[string]domain.SessionFact, len(got.Sessions))
	for _, session := range got.Sessions {
		if _, duplicate := sessions[session.ID]; duplicate {
			t.Fatalf("refresh duplicated %s", session.ID)
		}
		sessions[session.ID] = session
	}
	for _, prior := range accepted.Sessions {
		if !reflect.DeepEqual(sessions[prior.ID], prior) {
			t.Fatalf("refresh changed accepted historical owner %s: before=%+v after=%+v", prior.ID, prior, sessions[prior.ID])
		}
	}
	for owner, input := range map[string]int64{"first": 17, "second": 29} {
		if session := sessions[owner]; session.Source != domain.SourceCodex || len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != input {
			t.Fatalf("refresh omitted verified native facts for %s: %+v", owner, session)
		}
	}
	assertInitialPublicationAbsent(t, f.output+".publication-pending")
}
