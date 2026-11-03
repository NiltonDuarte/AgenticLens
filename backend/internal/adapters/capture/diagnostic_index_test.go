package capture

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func indexTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
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
func legacyDiagnostic(t *testing.T, s *Store, d domain.CaptureDiagnostic) int64 {
	t.Helper()
	body, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.db.Exec(`INSERT INTO capture_diagnostics(worktree_id,ticket,phase,outcome,record)VALUES(?,?,?,?,?)`, s.worktreeID, d.TicketID, d.Phase, d.Outcome, body)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := r.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return seq
}
func TestDiagnosticIndexBackfillBudgetResumeAndLegacyWriterGap(t *testing.T) {
	s := indexTestStore(t)
	ctx := context.Background()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/store", SessionID: "owner"}
	old := domain.CaptureDiagnostic{TicketID: "T-1", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Owners: []domain.CaptureRegistration{{Identity: id, State: domain.CaptureCurrent, RefreshedAt: "2026-01-01T00:00:00Z"}}}}
	legacyDiagnostic(t, s, old)
	legacyDiagnostic(t, s, old)
	tiny, err := s.BackfillDiagnosticIndex(ctx, 1, 1)
	if err != nil || tiny.Complete || tiny.Rows != 0 {
		t.Fatalf("byte gate: %+v %v", tiny, err)
	}
	page, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Owners: []domain.CaptureIdentity{id}, Limit: 18, MaxBytes: 1024 * 1024})
	if err != nil || page.MetadataComplete || !page.More {
		t.Fatalf("unindexed direct writer certified: %+v %v", page, err)
	}
	progress, err := s.BackfillDiagnosticIndex(ctx, 1, 1024*1024)
	if err != nil || progress.Complete || progress.Rows != 1 {
		t.Fatalf("row gate: %+v %v", progress, err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(s.directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	newer := old
	newer.Outcome = "failed"
	newer.Reconciliation = &domain.CaptureReconciliationDiagnostic{Owners: []domain.CaptureRegistration{{Identity: id, State: domain.CaptureFailed}}}
	if err = s.RecordDiagnostic(ctx, newer); err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.BackfillDiagnosticIndex(canceled, 10, 1024*1024); err == nil {
		t.Fatal("canceled maintenance succeeded")
	}
	progress, err = s.BackfillDiagnosticIndex(ctx, 10, 1024*1024)
	if err != nil || !progress.Complete {
		t.Fatalf("resume: %+v %v", progress, err)
	}
	page, err = s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Owners: []domain.CaptureIdentity{id}, Limit: 18, MaxBytes: 1024 * 1024})
	if err != nil || !page.MetadataComplete || page.More || !domain.CaptureCurrentFailures(domain.CaptureRegistration{Identity: id, TicketID: "T-1"}, page.Records).Source {
		t.Fatalf("older backfill erased newer terminal: %+v %v", page, err)
	}
	// A legacy writer can append after the index claimed completion. Both the
	// read-only detector and the next typed write must preserve this unknown gap.
	legacyDiagnostic(t, s, old)
	if err = s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "T-2", Phase: "worker", Outcome: "completed"}); err != nil {
		t.Fatal(err)
	}
	page, err = s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Limit: 18, MaxBytes: 1024 * 1024})
	if err != nil || page.MetadataComplete || !page.More {
		t.Fatalf("postcomplete gap hidden: %+v %v", page, err)
	}
	progress, err = s.BackfillDiagnosticIndex(ctx, 10, 1024*1024)
	if err != nil || !progress.Complete {
		t.Fatalf("gap repair: %+v %v", progress, err)
	}
}
func TestDiagnosticHeadLookupWorkIndependentOfRotatingOwnerHistory(t *testing.T) {
	for _, count := range []int{32, 128, 256} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			s := indexTestStore(t)
			ctx := context.Background()
			owners := make([]domain.CaptureRegistration, count)
			for i := range owners {
				owners[i] = domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/store", SessionID: fmt.Sprint(i)}, State: domain.CapturePending}
			}
			for i := range owners {
				owners[i].State = domain.CaptureCurrent
				owners[i].RefreshedAt = "2026-01-01T00:00:00Z"
				legacyDiagnostic(t, s, domain.CaptureDiagnostic{TicketID: "T-1", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Owners: owners}})
				owners[i].State = domain.CapturePending
				owners[i].RefreshedAt = ""
			}
			progress, err := s.BackfillDiagnosticIndex(ctx, count+1, 64*1024*1024)
			if err != nil || !progress.Complete {
				t.Fatalf("backfill: %+v %v", progress, err)
			}
			page, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Owners: []domain.CaptureIdentity{owners[0].Identity}, Limit: 18, MaxBytes: 1})
			if err != nil || !page.MetadataComplete || !page.More || page.Bytes != 0 || len(page.Records) != 0 || page.HeadLookups != 8 {
				t.Fatalf("bounded complete-index work: %+v %v", page, err)
			}
			rows, err := s.db.QueryContext(ctx, "EXPLAIN QUERY PLAN "+diagnosticHeadQuery, s.worktreeID, "T-1", "reconciliation", IdentityKey(owners[0].Identity))
			if err != nil {
				t.Fatal(err)
			}
			plan := ""
			for rows.Next() {
				var a, b, c int
				var detail string
				if err = rows.Scan(&a, &b, &c, &detail); err != nil {
					t.Fatal(err)
				}
				plan += detail
			}
			if err = rows.Err(); err != nil {
				t.Fatal(err)
			}
			if err = rows.Close(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(plan, "SEARCH capture_diagnostic_heads USING INDEX") || strings.Contains(plan, "SCAN ") {
				t.Fatalf("head query lacks prefix seek: %s", plan)
			}
		})
	}
}
func TestDiagnosticReadLegacyMissingMetadataNeverRepairsSchema(t *testing.T) {
	s := indexTestStore(t)
	ctx := context.Background()
	legacyDiagnostic(t, s, domain.CaptureDiagnostic{TicketID: "T-1", Phase: "publication", Outcome: "failed"})
	for _, table := range []string{"capture_diagnostic_index", "capture_diagnostic_heads", "capture_diagnostic_failures"} {
		if _, err := s.db.Exec("DROP TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.db.Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	page, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Limit: 18, MaxBytes: 1024})
	if err != nil || !page.Available || page.MetadataComplete || !page.More || len(page.Records) != 0 {
		t.Fatalf("legacy current: %+v %v", page, err)
	}
	history, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Limit: 1, MaxBytes: 1024})
	if err != nil || len(history.Records) != 1 {
		t.Fatalf("legacy history: %+v %v", history, err)
	}
	var n int
	if err = s.db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE name='capture_diagnostic_heads'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("inspection repaired schema: %d %v", n, err)
	}
}

func TestDiagnosticLegacyPayloadDoesNotDelayDurableRegistration(t *testing.T) {
	directory := t.TempDir()
	s, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	// Old private rows are deliberately undecodable. Opening and registering must
	// not parse them or perform a historical backfill before the durable handoff.
	if _, err = s.db.Exec(`INSERT INTO capture_diagnostics(worktree_id,ticket,phase,outcome,record)VALUES(?,'T-1','reconciliation','failed',?)`, s.worktreeID, strings.Repeat("invalid", 1024*1024)); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"capture_diagnostic_index", "capture_diagnostic_heads", "capture_diagnostic_failures"} {
		if _, err = s.db.Exec("DROP TABLE " + table); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "new-handoff"}
	if _, err = s.Register(context.Background(), id, "", "SessionStart"); err != nil {
		t.Fatalf("registration waited for legacy decode: %v", err)
	}
	var complete, cursor int
	if err = s.db.QueryRow(`SELECT complete,cursor FROM capture_diagnostic_index`).Scan(&complete, &cursor); err != nil || complete != 0 || cursor != 0 {
		t.Fatalf("Open/Register backfilled: complete=%d cursor=%d %v", complete, cursor, err)
	}
	if _, err = s.BackfillDiagnosticIndex(context.Background(), 1, 16*1024*1024); err == nil {
		t.Fatal("explicit maintenance falsely indexed corrupt legacy payload")
	}
	// Failure to index the old diagnostic does not undo the accepted handoff.
	var count int
	if err = s.db.QueryRow(`SELECT count(*) FROM registrations WHERE identity=?`, IdentityKey(id)).Scan(&count); err != nil || count != 1 {
		t.Fatalf("handoff lost: %d %v", count, err)
	}
}

func TestDiagnosticHistoryOversizedRowDoesNotAdvanceCursor(t *testing.T) {
	s := indexTestStore(t)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if err := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: fmt.Sprintf("T-%d", i), Phase: "publication", Outcome: "committed"}); err != nil {
			t.Fatal(err)
		}
	}
	for _, before := range []int64{0, 2} {
		page, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Before: before, Limit: 1, MaxBytes: 1})
		if err != nil || !page.More || len(page.Records) != 0 || page.NextBefore != before {
			t.Fatalf("oversized row skipped: %+v %v", page, err)
		}
		retry, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Before: page.NextBefore, Limit: 1, MaxBytes: 1024 * 1024})
		if err != nil || len(retry.Records) != 1 || (before > 0 && retry.Records[0].Sequence >= before) {
			t.Fatalf("budget retry: %+v %v", retry, err)
		}
	}
}

func TestDiagnosticAttemptMetadataLegacyAndOlderWriterStayUnknown(t *testing.T) {
	s := indexTestStore(t)
	ctx := context.Background()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/store", SessionID: "root"}
	owner := domain.CaptureRegistration{Identity: id, TicketID: "T-1", State: domain.CaptureCurrent, RefreshedAt: "2026-01-01T00:00:00Z"}
	fatal := domain.CaptureDiagnostic{TicketID: "T-1", Phase: "reconciliation", Outcome: "failed", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptFailed, Owners: []domain.CaptureRegistration{owner}}}
	if err := s.RecordDiagnostic(ctx, fatal); err != nil {
		t.Fatal(err)
	}
	// Candidate03 heads had no whole-attempt freshness watermark. Read-only access
	// must not silently trust or upgrade that already-complete old index.
	if _, err := s.db.Exec(`DELETE FROM capture_diagnostic_heads WHERE owner=?`, attemptIndexOwner); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	page, err := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, ExactOwners: []DiagnosticOwner{{TicketID: "T-1", Identity: id}}, Limit: 24, MaxBytes: 1024 * 1024})
	if err != nil || page.MetadataComplete || !page.More {
		t.Fatalf("legacy metadata certified: %+v %v", page, err)
	}
	var columns int
	if err = s.db.QueryRow(`SELECT count(*) FROM capture_diagnostic_heads WHERE owner=?`, attemptIndexOwner).Scan(&columns); err != nil || columns != 0 {
		t.Fatalf("read-only migrated schema: %d %v", columns, err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(s.directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	small, err := s.BackfillDiagnosticIndex(ctx, 1, 1)
	if err != nil || small.Complete || small.Rows != 0 {
		t.Fatalf("upgrade ignored backfill budget: %+v %v", small, err)
	}
	complete, err := s.BackfillDiagnosticIndex(ctx, 4, 1024*1024)
	if err != nil || !complete.Complete {
		t.Fatalf("upgrade resume: %+v %v", complete, err)
	}
	if _, err = s.db.Exec(`INSERT OR IGNORE INTO capture_diagnostic_index SELECT 1,0,coalesce(max(sequence),0),coalesce(max(sequence),0),CASE WHEN count(*)=0 THEN 1 ELSE 0 END FROM (SELECT sequence FROM capture_diagnostics ORDER BY sequence DESC LIMIT 1)`); err == nil {
		t.Fatal("unscoped schema1 initialization was accepted by schema2")
	}
	// An explicitly scoped direct writer can still bypass the attempt watermark.
	// Its append must invalidate freshness even when it claims complete.
	legacy := fatal
	legacy.Reconciliation = &domain.CaptureReconciliationDiagnostic{Owners: []domain.CaptureRegistration{owner}}
	seq := legacyDiagnostic(t, s, legacy)
	if _, err = s.db.Exec(`UPDATE capture_diagnostic_index SET known_max=?,complete=1`, seq); err != nil {
		t.Fatal(err)
	}
	page, err = s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Limit: 24, MaxBytes: 1024 * 1024})
	if err != nil || page.MetadataComplete || !page.More {
		t.Fatalf("older writer gap certified: %+v %v", page, err)
	}
	partial := fatal
	partial.Outcome = "accepted"
	partial.Reconciliation = &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{owner}}
	if err = s.RecordDiagnostic(ctx, partial); err != nil {
		t.Fatal(err)
	}
	complete, err = s.BackfillDiagnosticIndex(ctx, 4, 1024*1024)
	if err != nil || !complete.Complete {
		t.Fatalf("old writer backfill: %+v %v", complete, err)
	}
	page, err = s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Limit: 24, MaxBytes: 1024 * 1024})
	failures := domain.CaptureCurrentFailures(owner, page.Records)
	if err != nil || !page.MetadataComplete || page.More || !failures.Collection || !failures.AttemptUnknown {
		t.Fatalf("legacy ambiguity erased typed fatal: %+v %+v %v", page, failures, err)
	}
	partial.Reconciliation.Attempt = domain.CaptureAttemptCompleted
	if err = s.RecordDiagnostic(ctx, partial); err != nil {
		t.Fatal(err)
	}
	page, err = s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"T-1"}, Limit: 24, MaxBytes: 1024 * 1024})
	failures = domain.CaptureCurrentFailures(owner, page.Records)
	if err != nil || failures.Collection || failures.AttemptUnknown {
		t.Fatalf("completed recovery remains unresolved: %+v %v", failures, err)
	}
}

func TestDiagnosticIndexScopeDoesNotScanOrCertifyAnotherWorktree(t *testing.T) {
	directory := t.TempDir()
	first, err := openRegistry(directory, "Shared diagnostic project", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := openRegistry(directory, "Shared diagnostic project", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	ctx := context.Background()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: "/store", SessionID: "same-owner"}
	diagnostic := domain.CaptureDiagnostic{TicketID: "TEST-1", Phase: "reconciliation", Outcome: "failed", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptFailed, Owners: []domain.CaptureRegistration{{Identity: id, TicketID: "TEST-1", State: domain.CaptureFailed}}}}
	if err = first.RecordDiagnostic(ctx, diagnostic); err != nil {
		t.Fatal(err)
	}
	// A corrupt direct write belongs only to the second worktree. Inspection of
	// the first scope must neither decode it nor claim to have indexed it.
	if _, err = second.db.Exec(`INSERT INTO capture_diagnostics(worktree_id,ticket,phase,outcome,record) VALUES(?,'TEST-1','reconciliation','failed',?)`, second.worktreeID, strings.Repeat("not-json", 1024)); err != nil {
		t.Fatal(err)
	}
	options := DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-1"}, Owners: []domain.CaptureIdentity{id}, Limit: 18, MaxBytes: 1024 * 1024}
	own, err := first.ReadDiagnostics(ctx, options)
	if err != nil || !own.MetadataComplete || own.More || len(own.Records) != 1 {
		t.Fatalf("foreign history poisoned first scope: %+v %v", own, err)
	}
	progress, err := first.BackfillDiagnosticIndex(ctx, 1, 1)
	if err != nil || !progress.Complete || progress.Rows != 0 {
		t.Fatalf("first scope consumed foreign maintenance work: %+v %v", progress, err)
	}
	foreign, err := second.ReadDiagnostics(ctx, options)
	if err != nil || foreign.MetadataComplete || !foreign.More {
		t.Fatalf("foreign gap silently acknowledged: %+v %v", foreign, err)
	}
	if _, err = second.BackfillDiagnosticIndex(ctx, 1, 1024*1024); err == nil {
		t.Fatal("own corrupt history was certified as indexed")
	}
	own, err = first.ReadDiagnostics(ctx, options)
	if err != nil || !own.MetadataComplete || len(own.Records) != 1 {
		t.Fatalf("foreign failed maintenance changed first scope: %+v %v", own, err)
	}
}
