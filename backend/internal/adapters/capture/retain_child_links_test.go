package capture_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex/codexstoretest"
	"agenticlens/internal/domain"
)

// FIX-21: capture.RetainAcceptedChildLinks applies the retention rule once
// for generation, against accepted facts from another store such as the
// committed ledger, and records the unresolved evidence it adds.

func retentionOwner(id, parent string, delegations ...domain.DelegationFact) domain.SessionFact {
	owner := domain.SessionFact{ID: id, Source: domain.SourceCodex, Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}, Delegations: delegations}
	if parent != "" {
		owner.ParentID = &parent
	}
	return owner
}

func retentionDelegation(task string, at string, child string) domain.DelegationFact {
	agent := "worker"
	return domain.DelegationFact{AgentType: &agent, Description: task, InvokedAt: &at, ChildSession: child}
}

func retentionLinks(sessions []domain.SessionFact, id string) []string {
	links, _ := childLinkLinks(sessions, id)
	return links
}

type retentionFixture struct {
	project, state string
	family         *codexstoretest.TaskFamily
	store          *capture.Store
	result         capture.ReconcileResult
	accepted       []domain.SessionFact
}

// newRetentionFixture reads only the parent of a task-name family through a
// real project registry, and builds the accepted ledger facts that link every
// child, as a committed ledger from another machine would.
func newRetentionFixture(t *testing.T) retentionFixture {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	captureProjectConfig(t, project, "RetentionFixture")
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	family := codexstoretest.NewTaskFamily(t, home, project, childLinkCreated, childLinkTasks, 2, childLinkLargeBytes)
	state := t.TempDir()
	s, err := capture.OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	registerBound(t, s, home, family.Path(family.ParentID), family.ParentID, childLinkTicket)
	result, err := s.ReconcileAt(context.Background(), childLinkTicket, "tester", 20, childLinkPartialBudget, childLinkCreated.Add(24*time.Hour), fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if latest := latestTicketReconciliation(t, s); latest.Attempt != domain.CaptureAttemptCompleted || result.PendingChildren != 0 {
		t.Fatalf("fixture pass must complete before retention: %+v", latest)
	}
	parent := result.RefreshedSessions[0]
	if parent.ID != family.ParentID || !slices.Equal(retentionLinks(result.RefreshedSessions, family.ParentID), []string{"", "", "", ""}) {
		t.Fatalf("fixture must refresh the parent alone with unresolved task names: %+v", result.RefreshedSessions)
	}
	linked := parent
	linked.Delegations = append([]domain.DelegationFact(nil), parent.Delegations...)
	accepted := []domain.SessionFact{}
	for i, child := range family.ChildIDs {
		linked.Delegations[i].ChildSession = child
		accepted = append(accepted, retentionOwner(child, family.ParentID))
	}
	return retentionFixture{project: project, state: state, family: family, store: s, result: result, accepted: append([]domain.SessionFact{linked}, accepted...)}
}

func latestTicketReconciliation(t *testing.T, s *capture.Store) domain.CaptureReconciliationDiagnostic {
	t.Helper()
	diagnostics, err := s.Diagnostics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := len(diagnostics) - 1; i >= 0; i-- {
		if diagnostics[i].TicketID == childLinkTicket && diagnostics[i].Reconciliation != nil {
			return *diagnostics[i].Reconciliation
		}
	}
	t.Fatal("no reconciliation diagnostic")
	return domain.CaptureReconciliationDiagnostic{}
}

func TestRetainAcceptedChildLinksCountsRefusedLinksOnceBesideFreshLinks(t *testing.T) {
	fx := newRetentionFixture(t)
	family := fx.family
	// Children 3 and 4 have the same provenance key in the accepted ledger, so
	// their links are refused rather than moved. Child 3 is also queued work;
	// its hint reopens the sweep, which inherits children 1 and 2, so they are
	// read with the parent and resolve freshly.
	accepted := append([]domain.SessionFact(nil), fx.accepted...)
	parent := accepted[0]
	parent.Delegations = append([]domain.DelegationFact(nil), parent.Delegations...)
	parent.Delegations[3].AgentType, parent.Delegations[3].Description, parent.Delegations[3].InvokedAt = parent.Delegations[2].AgentType, parent.Delegations[2].Description, parent.Delegations[2].InvokedAt
	accepted[0] = parent
	if err := fx.store.QueueChildStart(context.Background(), family.Home, family.ParentID, family.ChildIDs[2]); err != nil {
		t.Fatal(err)
	}
	result, err := fx.store.ReconcileAt(context.Background(), childLinkTicket, "tester", 20, childLinkPartialBudget, childLinkCreated.Add(24*time.Hour), fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	queued := result.PendingChildren
	if queued != 1 {
		t.Fatalf("fixture must count child 3's queued start once, got %d", queued)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	retention, err := capture.RetainAcceptedChildLinks(ctx, fx.state, fx.project, &result, accepted)
	if err != nil {
		t.Fatal(err)
	}
	if retention.Kept != 0 || retention.Refused != 2 || retention.Cleared != 0 || retention.Added != 1 {
		t.Fatalf("retention = %+v, want 2 refused and only child 4 added beside child 3's queued start", retention)
	}
	if got := retentionLinks(result.RefreshedSessions, family.ParentID); !slices.Equal(got, []string{family.ChildIDs[0], family.ChildIDs[1], "", ""}) {
		t.Fatalf("links = %q: refused links must not move", got)
	}
	if result.PendingChildren != 2 || result.UnresolvedParentLinks != 0 {
		t.Fatalf("unresolved child evidence = %d, parent links = %d; want 2 and 0", result.PendingChildren, result.UnresolvedParentLinks)
	}
	if latest := latestTicketReconciliation(t, fx.store); latest.Attempt != domain.CaptureAttemptPartial || latest.PendingChildren != 2 {
		t.Fatalf("latest reconciliation = %s with %d unresolved children, want partial with 2", latest.Attempt, latest.PendingChildren)
	}
	// Applying it again to the same request counts nothing twice.
	again, err := capture.RetainAcceptedChildLinks(ctx, fx.state, fx.project, &result, accepted)
	if err != nil || again.Added != 0 || result.PendingChildren != 2 {
		t.Fatalf("second application = %+v, %v; unresolved %d", again, err, result.PendingChildren)
	}
}

func TestRetainAcceptedChildLinksRewritesRestoredSessions(t *testing.T) {
	const at1, at2 = "2026-11-24T09:00:07.000Z", "2026-11-24T09:00:30.000Z"
	accepted := []domain.SessionFact{
		retentionOwner("P", "", retentionDelegation("task-01", at1, "C1"), retentionDelegation("task-02", at2, "C2")),
		retentionOwner("C1", "P"), retentionOwner("C2", "P"),
	}
	// The parent is restored from registry state without its links. Child 2 is
	// refreshed in the same request and now names another parent.
	result := capture.ReconcileResult{
		RestoredSessions:  []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-01", at1, ""), retentionDelegation("task-02", at2, ""))},
		RefreshedSessions: []domain.SessionFact{retentionOwner("C2", "Q")},
	}
	retention, err := capture.RetainAcceptedChildLinks(context.Background(), t.TempDir(), t.TempDir(), &result, accepted)
	if err != nil {
		t.Fatal(err)
	}
	if got := retentionLinks(result.RestoredSessions, "P"); !slices.Equal(got, []string{"C1", ""}) {
		t.Fatalf("restored links = %q, want the accepted C1 and no contradicted C2", got)
	}
	if retention.Kept != 1 || retention.Added != 1 || result.PendingChildren != 1 {
		t.Fatalf("retention = %+v, unresolved %d; want one kept link", retention, result.PendingChildren)
	}
	if !reflect.DeepEqual(result.RefreshedSessions, []domain.SessionFact{retentionOwner("C2", "Q")}) {
		t.Fatalf("refreshed child changed: %+v", result.RefreshedSessions)
	}
}

func TestRetainAcceptedChildLinksUsesRegistrySessionsAsChildEvidence(t *testing.T) {
	const at1 = "2026-11-24T09:00:07.000Z"
	accepted := []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-01", at1, "C1")), retentionOwner("C1", "P")}
	for _, tc := range []struct {
		name     string
		fresh    string
		sessions []domain.SessionFact
		want     string
		kept     int
	}{
		// The registry proves C9, which the older ledger does not know.
		{"registry proves the fresh child", "C9", []domain.SessionFact{retentionOwner("C9", "P")}, "C9", 0},
		// The registry's record of C1 now names another parent.
		{"registry contradicts the ledger link", "", []domain.SessionFact{retentionOwner("C1", "Q")}, "", 0},
		{"registry agrees with the ledger", "", []domain.SessionFact{retentionOwner("C1", "P")}, "C1", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := capture.ReconcileResult{
				RefreshedSessions: []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-01", at1, tc.fresh))},
				Sessions:          tc.sessions,
			}
			retention, err := capture.RetainAcceptedChildLinks(context.Background(), t.TempDir(), t.TempDir(), &result, accepted)
			if err != nil {
				t.Fatal(err)
			}
			if got := retentionLinks(result.RefreshedSessions, "P"); !slices.Equal(got, []string{tc.want}) || retention.Kept != tc.kept {
				t.Fatalf("links = %q kept = %d, want %q and %d", got, retention.Kept, tc.want, tc.kept)
			}
		})
	}
}

func TestRetainAcceptedChildLinksReportsRegistryContention(t *testing.T) {
	fx := newRetentionFixture(t)
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(fx.store.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	holder, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	if _, err := holder.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	result := fx.result
	bounded, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = capture.RetainAcceptedChildLinks(bounded, fx.state, fx.project, &result, fx.accepted)
	if err == nil || !capture.IsRegistryContention(err) {
		t.Fatalf("a held registry writer must surface as contention, got %v", err)
	}
	if _, err := holder.ExecContext(ctx, "ROLLBACK"); err != nil {
		t.Fatal(err)
	}
	if latest := latestTicketReconciliation(t, fx.store); latest.Attempt != domain.CaptureAttemptCompleted {
		t.Fatalf("a contended write recorded %s", latest.Attempt)
	}
	// Once the writer releases, a new request records its partial state. It
	// has its own bounded context: the contended call above may have used
	// most of the first one.
	retry, err := fx.store.ReconcileAt(ctx, childLinkTicket, "tester", 20, childLinkPartialBudget, childLinkCreated.Add(24*time.Hour), fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Second)
	defer retryCancel()
	if _, err := capture.RetainAcceptedChildLinks(retryCtx, fx.state, fx.project, &retry, fx.accepted); err != nil {
		t.Fatal(err)
	}
	if latest := latestTicketReconciliation(t, fx.store); latest.Attempt != domain.CaptureAttemptPartial || latest.PendingChildren != 4 {
		t.Fatalf("latest reconciliation = %+v, want partial with 4 unresolved children", latest)
	}
}

// FIX-21 R2-T4: when recording the partial reconciliation fails after the
// registry opened, the error is returned and nothing is recorded as partial
// or complete in its place. A registry trigger rejects the diagnostic insert.
func TestRetainAcceptedChildLinksReturnsDiagnosticWriteFailure(t *testing.T) {
	fx := newRetentionFixture(t)
	db, err := sql.Open("sqlite", filepath.Join(fx.store.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TRIGGER reject_partial_diagnostic BEFORE INSERT ON capture_diagnostics
WHEN NEW.phase = 'reconciliation' BEGIN SELECT RAISE(ABORT, 'fixture rejects the partial diagnostic'); END`); err != nil {
		t.Fatal(err)
	}
	result := fx.result
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	retention, err := capture.RetainAcceptedChildLinks(ctx, fx.state, fx.project, &result, fx.accepted)
	if err == nil || !strings.Contains(err.Error(), "fixture rejects the partial diagnostic") {
		t.Fatalf("a failed diagnostic write must be returned, got %v", err)
	}
	if retention.Added != 4 || result.PendingChildren != 4 {
		t.Fatalf("retention = %+v, unresolved %d; the request still counts its kept links", retention, result.PendingChildren)
	}
	if latest := latestTicketReconciliation(t, fx.store); latest.Attempt != domain.CaptureAttemptCompleted || latest.PendingChildren != 0 {
		t.Fatalf("latest reconciliation changed despite the failed write: %+v", latest)
	}
}

// FIX-21 R2-T8: a link kept for one owner list and refused for the other is
// reported as kept, not as both.
func TestRetainAcceptedChildLinksReportsALinkKeptAndRefusedAsKept(t *testing.T) {
	const at1, at2 = "2026-11-24T09:00:07.000Z", "2026-11-24T09:00:30.000Z"
	accepted := []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-01", at1, "C1"))}
	// The refreshed parent keeps C1 at its delegation. The restored copy of
	// the same parent has no delegation with that key, so C1 is unmatched.
	result := capture.ReconcileResult{
		RefreshedSessions: []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-01", at1, ""))},
		RestoredSessions:  []domain.SessionFact{retentionOwner("P", "", retentionDelegation("task-02", at2, ""))},
	}
	retention, err := capture.RetainAcceptedChildLinks(context.Background(), t.TempDir(), t.TempDir(), &result, accepted)
	if err != nil {
		t.Fatal(err)
	}
	if retention.Kept != 1 || retention.Refused != 0 || retention.Added != 1 || result.PendingChildren != 1 {
		t.Fatalf("retention = %+v, unresolved %d; want one kept link, no refusal, counted once", retention, result.PendingChildren)
	}
}
