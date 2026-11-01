package capture_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func openStore(t *testing.T, directory string) *capture.Store {
	t.Helper()
	s, err := capture.Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}

func TestRegistryBindingRequiresExactSessionRequestAndPreservesConflicts(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s := openStore(t, dir)
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: filepath.Join(t.TempDir(), "source store"), SessionID: "session-one"}
	first, err := s.Register(ctx, id, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	secondID := id
	secondID.SessionID = "session-two"
	second, err := s.Register(ctx, secondID, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if first.Request == "" || first.Request == second.Request || first.TicketID != "" {
		t.Fatalf("invalid independent challenges: %+v %+v", first, second)
	}
	for _, tc := range []struct {
		name     string
		identity domain.CaptureIdentity
		request  string
	}{
		{"missing request", id, ""},
		{"other session token", id, second.Request},
		{"other session identity", secondID, first.Request},
		{"wrong store", domain.CaptureIdentity{Source: id.Source, Store: id.Store + "-other", SessionID: id.SessionID}, first.Request},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := s.Bind(ctx, tc.identity, "TEST-101", tc.request); err == nil {
				t.Fatal("accepted an unrelated request")
			}
		})
	}
	bound, err := s.Bind(ctx, id, "TEST-101", first.Request)
	if err != nil {
		t.Fatal(err)
	}
	if bound.TicketID != "TEST-101" || bound.Binding != "explicit_agent_response" {
		t.Fatalf("invalid binding: %+v", bound)
	}
	if _, err := s.Bind(ctx, id, "TEST-202", first.Request); err == nil {
		t.Fatal("conflicting ticket replaced explicit ownership")
	}
	repeated, err := s.Register(ctx, id, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if repeated.TicketID != "TEST-101" || repeated.Request != first.Request {
		t.Fatalf("resume changed binding/request: %+v", repeated)
	}
	rows, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("failed binding persisted an extra identity: %+v", rows)
	}
	for _, row := range rows {
		if row.Identity == secondID && row.TicketID != "" {
			t.Fatal("independent session stole ticket binding")
		}
	}
}

func TestRegistryTransactionRollbackAndConcurrentWriters(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	a, b := openStore(t, dir), openStore(t, dir)
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
	initial, err := a.Register(ctx, id, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("interrupt before commit")
	if _, err := a.Change(ctx, id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketID = "TEST-LOST"
		return row, rollback
	}); !errors.Is(err, rollback) {
		t.Fatalf("change error=%v", err)
	}
	var wg sync.WaitGroup
	errorsCh := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			store := a
			if n%2 == 1 {
				store = b
			}
			_, err := store.Change(ctx, id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				row.Revision++
				return row, nil
			})
			errorsCh <- err
		}(i)
	}
	wg.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	reopened := openStore(t, dir)
	rows, err := reopened.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Revision != 12 || rows[0].TicketID != "" || rows[0].Request != initial.Request {
		t.Fatalf("lost update/rollback/reopen: %+v", rows)
	}
}

func TestRegistryCrashRecoveryRetainsCommittedRequest(t *testing.T) {
	dir := t.TempDir()
	s := openStore(t, dir)
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: dir, SessionID: "crash-root"}
	row, err := s.Register(context.Background(), id, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary, "-test.run=^TestCaptureInterruptedWriterProcess$")
	cmd.Env = append(os.Environ(), "AL_CAPTURE_CRASH_DIRECTORY="+dir)
	out, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("crash helper did not interrupt transaction: err=%v out=%s", err, out)
	}
	reopened := openStore(t, dir)
	bound, err := reopened.Bind(context.Background(), id, "TEST-100", row.Request)
	if err != nil {
		t.Fatalf("crashed process stranded writer lock or request: %v", err)
	}
	if bound.TicketID != "TEST-100" || bound.Revision != 0 {
		t.Fatalf("uncommitted changes survived crash: %+v", bound)
	}
}

// TestCaptureInterruptedWriterProcess is the subprocess fixture for the recovery test.
func TestCaptureInterruptedWriterProcess(t *testing.T) {
	dir := os.Getenv("AL_CAPTURE_CRASH_DIRECTORY")
	if dir == "" {
		return
	}
	s, err := capture.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.Change(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: dir, SessionID: "crash-root"}, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.Revision = 999
		os.Exit(23)
		return row, nil
	})
	t.Fatalf("crash fixture returned without exiting: %v", err)
}
