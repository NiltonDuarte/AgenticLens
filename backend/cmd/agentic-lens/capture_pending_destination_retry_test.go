package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureNativeSelectionRechecksConflictsBeforeFirstCommit(t *testing.T) {
	for _, scenario := range []struct {
		name                          string
		removeOriginal, attemptMarker bool
	}{
		{"original tentative directory remains", false, false},
		{"original tentative directory removed", true, false},
		{"original directory removed after marker attempt", true, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			verifyNativeSelectionRetry(t, scenario.removeOriginal, scenario.attemptMarker)
		})
	}
}

func verifyNativeSelectionRetry(t *testing.T, removeOriginal, attemptMarker bool) {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	first := initialPublicationTarget(t, f, "indicated-first")
	second := initialPublicationTarget(t, f, "indicated-second")
	database := setupIsolatedEmptyDB(t)
	t.Setenv("OPENCODE_DB", database)
	policy := fmt.Sprintf("telemetry:\n  default_source: opencode\n  capture:\n    max_lifetime_ms: 1500\n  sources:\n    opencode:\n      database_path: %q\n", database)
	if err := writeCaptureFixtureLocal(t, f, []byte(policy)); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for root, amount := range map[string]int64{"first": 17, "second": 29} {
		if _, err := db.Exec("INSERT INTO session(id,agent,model,time_created,time_updated) VALUES (?,'build','fixture',1000,3000)", root); err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf(`{"role":"assistant","modelID":"fixture","time":{"created":1000,"completed":3000},"tokens":{"input":%d}}`, amount)
		if _, err := db.Exec("INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES (?,?,1000,3000,?)", "m-"+root, root, body); err != nil {
			t.Fatal(err)
		}
	}
	setOpenCodeFixtureProject(t, database, f.project)
	ctx := context.Background()
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	bind := func(root, path string) domain.CaptureRegistration {
		t.Helper()
		id := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: database, SessionID: root}
		observed, err := store.Register(ctx, id, "", "session.idle")
		if err != nil {
			t.Fatal(err)
		}
		binding := capture.BindingTicket{ID: initialPublicationTicket, Title: "Exact conflict fixture", Type: "fix", Destination: filepath.Dir(path)}
		row, err := store.BindNativeOpenCode(ctx, id, binding, observed.Request)
		if err != nil {
			t.Fatal(err)
		}
		return row
	}
	firstBound := bind("first", first)

	// Stop after the real initial destination selection has persisted its
	// receipt but before any canonical publication transaction has begun.
	cfg, err := adaptersconfig.Load(f.project)
	if err != nil {
		t.Fatal(err)
	}
	key := adapters.PublicationTicketKey(initialPublicationTicket, cfg.Ticket)
	lease, err := capture.AcquireInitialPublication(ctx, f.project, key, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, admitted, selectErr := adapters.SelectNativeWorkspacePublication(ctx, cfg, f.project, initialPublicationTicket, "", lease, nil, func() ([]domain.CaptureRegistration, error) {
		return store.List(ctx)
	})
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if selectErr != nil || admitted == nil || admitted.Path != first || admitted.Closed() {
		t.Fatalf("expected an open native receipt for the first indication: %+v %v", admitted, selectErr)
	}
	assertInitialPublicationAbsent(t, first, first+".publication-pending", second)
	bind("second", second)

	output, err := runInitialPublicationWorker(t, f)
	if err == nil || !strings.Contains(string(output), "conflicting ledger destinations") {
		t.Fatalf("old open receipt bypassed conflicting current indications: %v\n%s", err, output)
	}
	assertInitialPublicationAbsent(t, first, first+".publication-pending", second, second+".publication-pending")
	pending := filepath.Join(f.project, captureFixturePendingDir)
	snapshots, err := filepath.Glob(filepath.Join(pending, "*.json"))
	if err != nil || len(snapshots) != 2 {
		t.Fatalf("conflict did not retain both snapshots: %v %v", snapshots, err)
	}
	expected := map[string]int64{"first": 17, "second": 29}
	retained := make(map[string][]byte, len(snapshots))
	for _, path := range snapshots {
		var snapshot adapters.PendingSnapshot
		retained[path] = readCaptureFile(t, path)
		if err := json.Unmarshal(retained[path], &snapshot); err != nil {
			t.Fatal(err)
		}
		amount, known := expected[snapshot.RootSessionID]
		if !known || snapshot.State != "awaiting_destination" || snapshot.TicketID != initialPublicationTicket ||
			len(snapshot.Sessions) != 1 || snapshot.Sessions[0].ID != snapshot.RootSessionID ||
			len(snapshot.Sessions[0].Invocations) != 1 || snapshot.Sessions[0].Invocations[0].Tokens.Input != amount {
			t.Fatalf("blocked publication lost exact root facts: %+v", snapshot)
		}
		delete(expected, snapshot.RootSessionID)
	}
	if attemptMarker {
		err := func() (returnErr error) {
			ticketLease, err := capture.AcquireInitialPublication(ctx, f.project, key, nil)
			if err != nil {
				return err
			}
			defer func() { returnErr = errors.Join(returnErr, ticketLease.Close()) }()
			owner, err := capture.AcquirePublication(ctx, first, f.project)
			if err != nil {
				return err
			}
			defer func() { returnErr = errors.Join(returnErr, owner.Close()) }()
			return owner.MarkInitialPublicationMarkerAttempt(admitted.Link())
		}()
		if err != nil {
			t.Fatal(err)
		}
		// Interruption occurred after the private phase became durable and
		// before the public indicator could be written.
		assertInitialPublicationAbsent(t, first+".publication-pending")
	}
	if removeOriginal {
		if err := os.Remove(filepath.Dir(first)); err != nil {
			t.Fatal(err)
		}
	}
	agreement := capture.BindingTicket{ID: initialPublicationTicket, Destination: filepath.Dir(second)}
	if _, err := store.IndicateNativeOpenCodeDestination(ctx, firstBound.Identity, agreement, capture.NativePublicationDestinationRequest(firstBound)); err != nil {
		t.Fatal(err)
	}
	if attemptMarker {
		output, err := runInitialPublicationWorker(t, f)
		if err == nil || !strings.Contains(string(output), "old native selection") {
			t.Fatalf("a lost marker-attempt destination did not require exact recovery: %v\n%s", err, output)
		}
		assertInitialPublicationAbsent(t, filepath.Dir(first), first, second, second+".publication-pending")
		for path, before := range retained {
			if !bytes.Equal(readCaptureFile(t, path), before) {
				t.Fatalf("blocked retry changed pending facts at %s", path)
			}
		}
		ticketLease, err := capture.AcquireInitialPublication(ctx, f.project, key, nil)
		if err != nil {
			t.Fatal(err)
		}
		current, err := ticketLease.Read()
		if err := errors.Join(err, ticketLease.Close()); err != nil {
			t.Fatal(err)
		}
		if current == nil || current.Path != first || current.Link() != admitted.Link() || current.MarkerPhase != "attempted" {
			t.Fatalf("blocked retry replaced the exact receipt: %+v", current)
		}
		return
	}
	if err := writeCaptureFixtureLocal(t, f, []byte(strings.Replace(policy, "1500", "5000", 1))); err != nil {
		t.Fatal(err)
	}
	if output, err := runInitialPublicationWorker(t, f); err != nil {
		t.Fatalf("agreed second indication did not replace the open receipt: %v\n%s", err, output)
	}
	assertInitialPublicationAbsent(t, first, second+".publication-pending")
	if removeOriginal {
		assertInitialPublicationAbsent(t, filepath.Dir(first))
	}
	assertInitialPublicationAbsent(t, snapshots...)
	f.output = second
	ledger := f.ledger(t)
	if ledger.TicketID == nil || *ledger.TicketID != initialPublicationTicket || len(ledger.Sessions) != 2 {
		t.Fatalf("resolved conflict lost ticket identity or owners: %+v", ledger)
	}
	expected = map[string]int64{"first": 17, "second": 29}
	for _, owner := range ledger.Sessions {
		amount, known := expected[owner.ID]
		if !known || owner.Source != domain.SourceOpenCode || len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != amount {
			t.Fatalf("resolved conflict changed or duplicated an owner: %+v", owner)
		}
		delete(expected, owner.ID)
	}
}
