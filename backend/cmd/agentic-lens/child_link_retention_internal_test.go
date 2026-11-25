package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex/codexstoretest"
	"agenticlens/internal/domain"
)

// FIX-21 R3-02 and R3-T2: childLinkRetentionOutcome maps a retention result
// to the generation exit code and its stderr line.
func TestChildLinkRetentionOutcome(t *testing.T) {
	busy := sqliteBusyError(t)
	for _, tc := range []struct {
		name      string
		retention capture.ChildLinkRetention
		pending   int
		err       error
		code      int
		message   string
	}{
		{name: "registry contention is busy", err: busy, code: capturePublicationBusyExit, message: "Cannot record retained child links; publication not completed: "},
		{name: "joined contention is busy", err: errors.Join(fmt.Errorf("recording: %w", busy)), code: capturePublicationBusyExit, message: "publication not completed"},
		{name: "an expired wait marked by capture is busy", err: expiredRegistryWait(t), code: capturePublicationBusyExit, message: "publication not completed"},
		{name: "another error fails", err: errors.New("disk full"), code: 1, message: "publication not completed: disk full"},
		{name: "an unmarked expired context fails", err: fmt.Errorf("opening: %w", context.DeadlineExceeded), code: 1, message: "publication not completed"},
		{name: "nothing to report", retention: capture.ChildLinkRetention{}, pending: 3, code: 0, message: ""},
		{
			name:      "added evidence remains partial",
			retention: capture.ChildLinkRetention{Kept: 2, Refused: 1, Added: 3},
			pending:   5,
			message:   "Capture kept 2 accepted delegation child link(s) for ticket TEST-100 that this pass did not re-read, could not place 1 accepted link(s) whose delegation group is shorter than accepted, has no fresh match or was replaced by a confirmed child, and cleared 0 contradicted link(s); unresolved child evidence: 5, so the reconciliation remains partial.",
		},
		{
			name:      "only cleared links",
			retention: capture.ChildLinkRetention{Cleared: 1},
			pending:   0,
			message:   "Capture kept 0 accepted delegation child link(s) for ticket TEST-100 that this pass did not re-read, could not place 0 accepted link(s) whose delegation group is shorter than accepted, has no fresh match or was replaced by a confirmed child, and cleared 1 contradicted link(s); unresolved child evidence: 0.",
		},
		{
			name:      "kept links already counted add nothing",
			retention: capture.ChildLinkRetention{Kept: 1},
			pending:   1,
			message:   "unresolved child evidence: 1.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, message := childLinkRetentionOutcome("TEST-100", tc.retention, tc.pending, tc.err)
			if code != tc.code {
				t.Fatalf("code = %d, want %d (%s)", code, tc.code, message)
			}
			if tc.err == nil && tc.message != "" && !strings.HasSuffix(message, tc.message) {
				t.Fatalf("message = %q, want it to end %q", message, tc.message)
			}
			if tc.err != nil && !strings.Contains(message, tc.message) {
				t.Fatalf("message = %q, want it to contain %q", message, tc.message)
			}
			if tc.message == "" && message != "" {
				t.Fatalf("message = %q, want none", message)
			}
		})
	}
}

// sqliteBusyError returns a real SQLite busy error from a held writer.
func sqliteBusyError(t *testing.T) error {
	t.Helper()
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "busy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(ctx, "CREATE TABLE fixture(id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatal(err)
	}
	owner, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = owner.Close() })
	if _, err := owner.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = owner.ExecContext(ctx, "ROLLBACK") })
	contender, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = contender.Close() })
	if _, err := contender.ExecContext(ctx, "PRAGMA busy_timeout=0"); err != nil {
		t.Fatal(err)
	}
	_, busy := contender.ExecContext(ctx, "INSERT INTO fixture VALUES (1)")
	if busy == nil || !capture.IsRegistryContention(busy) {
		t.Fatalf("fixture did not produce SQLite contention: %v", busy)
	}
	return busy
}

// expiredRegistryWait returns the error capture.RetainAcceptedChildLinks
// gives when its bounded wait expires while another connection holds the
// project registry.
func expiredRegistryWait(t *testing.T) error {
	t.Helper()
	ctx := context.Background()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	shared, err := os.ReadFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	local := "agenticlens:\n  state_dir: publication-state\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: .agenticlens-capture\n    max_sessions: 256\n    max_source_bytes: 67108864\n"
	if err := errors.Join(os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), shared, 0o600), os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0o600)); err != nil {
		t.Fatal(err)
	}
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	created := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	family := codexstoretest.NewTaskFamily(t, home, project, created, 1, 1, 0)
	state := t.TempDir()
	s, err := capture.OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: family.ParentID}
	row, err := s.Register(ctx, identity, family.Path(family.ParentID), "SessionStart")
	if err == nil {
		_, err = s.Bind(ctx, identity, "TEST-100", row.Request)
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(family.Path(family.ChildIDs[0])); err != nil {
		t.Fatal(err)
	}
	result, err := s.ReconcileAt(ctx, "TEST-100", "tester", 16, 64<<20, created.Add(24*time.Hour), domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true})
	if err != nil || len(result.RefreshedSessions) != 1 {
		t.Fatalf("fixture pass = %+v, %v", result.RefreshedSessions, err)
	}
	parent := result.RefreshedSessions[0]
	parent.Delegations = append([]domain.DelegationFact(nil), parent.Delegations...)
	parent.Delegations[0].ChildSession = family.ChildIDs[0]
	db, err := sql.Open("sqlite", filepath.Join(s.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	holder, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = holder.Close() })
	if _, err := holder.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = holder.ExecContext(ctx, "ROLLBACK") }()
	expiring, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()
	_, err = capture.RetainAcceptedChildLinks(expiring, state, project, &result, []domain.SessionFact{parent})
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("fixture must expire its wait on the held registry, got %v", err)
	}
	return err
}
