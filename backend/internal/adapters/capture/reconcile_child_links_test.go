package capture_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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

// FIX-21 child link retention through partial capture refresh.
//
// One Codex parent delegates four children through task-name-only
// spawn_agent results, so a link resolves only when the child's rollout is
// read in the same pass. Children 3 and 4 are larger than the partial-pass
// byte budget, which reproduces a pass that reads the parent with only some
// of its children.

const (
	childLinkTicket = "TEST-100"
	childLinkTasks  = 4
	// childLinkLargeBytes is larger than childLinkPartialBudget, so a partial
	// pass can never extract children 3 and 4.
	childLinkLargeBytes          = 300 * 1024
	childLinkPartialBudget int64 = 160 * 1024
)

var childLinkCreated = time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)

// childLinkFamily keeps the fixture's short local names over the shared
// task-name family writer.
type childLinkFamily struct {
	*codexstoretest.TaskFamily
	home     string
	parent   string
	children []string
}

func newChildLinkFamily(t *testing.T) childLinkFamily {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	family := codexstoretest.NewTaskFamily(t, home, "/fixture/project", childLinkCreated, childLinkTasks, 2, childLinkLargeBytes)
	return childLinkFamily{TaskFamily: family, home: home, parent: family.ParentID, children: family.ChildIDs}
}

func (f childLinkFamily) writeParent(t *testing.T, tokens int, outputs ...map[int]string) {
	t.Helper()
	var replaced map[int]string
	if len(outputs) > 0 {
		replaced = outputs[0]
	}
	f.WriteParent(t, tokens, replaced)
}

func (f childLinkFamily) writeChild(t *testing.T, i int, parent string, tokens int) {
	t.Helper()
	f.WriteChild(t, i, parent, tokens)
}

type childLinkRegistry struct {
	store  *capture.Store
	family childLinkFamily
}

// newChildLinkRegistry binds the parent and queues one SubagentStart hint per
// child, as Codex lifecycle hooks do, then settles a full pass.
func newChildLinkRegistry(t *testing.T) (childLinkRegistry, capture.ReconcileResult) {
	t.Helper()
	ctx := context.Background()
	family := newChildLinkFamily(t)
	registry := childLinkRegistry{store: openStore(t, t.TempDir()), family: family}
	registerBound(t, registry.store, family.home, family.Path(family.parent), family.parent, childLinkTicket)
	for _, child := range family.children {
		if err := registry.store.QueueChildStart(ctx, family.home, family.parent, child); err != nil {
			t.Fatal(err)
		}
	}
	var full capture.ReconcileResult
	for pass := 0; pass < 8; pass++ {
		full = registry.reconcile(t, fixtureByteBudget)
		links, _ := childLinkLinks(full.RefreshedSessions, family.parent)
		if !full.DiscoveryPending && full.PendingChildren == 0 && slices.Equal(links, family.children) {
			return registry, full
		}
	}
	t.Fatalf("a full capture never linked every child: pending=%d discovery=%t sessions=%+v", full.PendingChildren, full.DiscoveryPending, full.Sessions)
	return registry, full
}

func (r childLinkRegistry) reconcile(t *testing.T, budget int64) capture.ReconcileResult {
	t.Helper()
	result, err := r.store.Reconcile(context.Background(), childLinkTicket, "tester", 20, budget, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceBytes > budget {
		t.Fatalf("pass charged %d above its %d budget", result.SourceBytes, budget)
	}
	return result
}

// snapshots returns every registered owner's persisted snapshot by session.
func (r childLinkRegistry) snapshots(t *testing.T) map[string][]domain.SessionFact {
	t.Helper()
	rows, err := r.store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][]domain.SessionFact{}
	for _, row := range rows {
		result[row.Identity.SessionID] = row.Snapshot
	}
	return result
}

func (r childLinkRegistry) latestReconciliation(t *testing.T) domain.CaptureReconciliationDiagnostic {
	t.Helper()
	diagnostics, err := r.store.Diagnostics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := len(diagnostics) - 1; i >= 0; i-- {
		if diagnostics[i].TicketID == childLinkTicket && diagnostics[i].Reconciliation != nil {
			return *diagnostics[i].Reconciliation
		}
	}
	t.Fatal("no reconciliation diagnostic recorded")
	return domain.CaptureReconciliationDiagnostic{}
}

// childLinkLinks returns the parent's delegation links in delegation order.
func childLinkLinks(sessions []domain.SessionFact, parent string) ([]string, []domain.DelegationFact) {
	for _, session := range sessions {
		if session.ID != parent {
			continue
		}
		links := []string{}
		for _, delegation := range session.Delegations {
			links = append(links, delegation.ChildSession)
		}
		return links, session.Delegations
	}
	return nil, nil
}

func refreshedIDs(result capture.ReconcileResult) map[string]bool {
	ids := map[string]bool{}
	for _, session := range result.RefreshedSessions {
		ids[session.ID] = true
	}
	return ids
}

func spawnEventChildren(result capture.ReconcileResult, parent string) map[string]string {
	children := map[string]string{}
	for _, event := range result.SourceEvents.Events {
		if event.SessionID == parent && strings.HasPrefix(event.CallID, "call-spawn-") {
			if previous, seen := children[event.CallID]; !seen || previous == "" {
				children[event.CallID] = event.ChildSessionID
			}
		}
	}
	return children
}

func parentInput(t *testing.T, sessions []domain.SessionFact, parent string) int64 {
	t.Helper()
	for _, session := range sessions {
		if session.ID == parent && len(session.Invocations) == 1 {
			return session.Invocations[0].Tokens.Input
		}
	}
	t.Fatalf("parent %s has no single invocation in %+v", parent, sessions)
	return 0
}

func TestReconcileRetainsAcceptedChildLinksThroughPartialRefresh(t *testing.T) {
	// Child 3 is already pending child work for the parent, either as a queued
	// SubagentStart naming it or as a queued child path registered to it.
	for _, queue := range []string{"queued start", "queued child path"} {
		t.Run(queue, func(t *testing.T) { partialRefreshRetainsLinks(t, queue == "queued child path") })
	}
}

func partialRefreshRetainsLinks(t *testing.T, byPath bool) {
	registry, full := newChildLinkRegistry(t)
	family := registry.family
	accepted := registry.snapshots(t)
	acceptedLinks, acceptedDelegations := childLinkLinks(accepted[family.parent], family.parent)
	if !slices.Equal(acceptedLinks, family.children) {
		t.Fatalf("full pass links = %q, want every child %q", acceptedLinks, family.children)
	}
	if full.UnresolvedParentLinks != 0 || registry.latestReconciliation(t).Attempt != domain.CaptureAttemptCompleted {
		t.Fatalf("full pass did not complete: %+v", registry.latestReconciliation(t))
	}

	// The parent grows, and child 3 is already counted as pending child work
	// for this parent.
	family.writeParent(t, 41)
	queue := registry.store.QueueChildStart(context.Background(), family.home, family.parent, family.children[2])
	if byPath {
		queue = registry.store.QueueChild(context.Background(), family.home, family.Path(family.children[2]), family.parent)
	}
	if queue != nil {
		t.Fatal(queue)
	}
	partial := registry.reconcile(t, childLinkPartialBudget)
	read := refreshedIDs(partial)
	if !read[family.parent] || !read[family.children[0]] || !read[family.children[1]] || read[family.children[2]] || read[family.children[3]] {
		t.Fatalf("fixture must read the parent with only children 1 and 2, read %v", read)
	}
	if got := parentInput(t, partial.RefreshedSessions, family.parent); got != 41 {
		t.Fatalf("parent was not refreshed from its grown source: input %d", got)
	}

	t.Run("every accepted link is kept in the registry and the refreshed owners", func(t *testing.T) {
		stored := registry.snapshots(t)
		links, delegations := childLinkLinks(stored[family.parent], family.parent)
		if !slices.Equal(links, family.children) {
			t.Fatalf("registry links = %q, want %q", links, family.children)
		}
		if !reflect.DeepEqual(delegations, acceptedDelegations) {
			t.Fatalf("delegations changed beyond their links: %+v, want %+v", delegations, acceptedDelegations)
		}
		if links, _ := childLinkLinks(partial.RefreshedSessions, family.parent); !slices.Equal(links, family.children) {
			t.Fatalf("refreshed parent links = %q, want %q", links, family.children)
		}
		if links, _ := childLinkLinks(partial.Sessions, family.parent); !slices.Equal(links, family.children) {
			t.Fatalf("published parent links = %q, want %q", links, family.children)
		}
		for _, child := range family.children[2:] {
			if !reflect.DeepEqual(stored[child], accepted[child]) {
				t.Fatalf("unread child %s lost its accepted snapshot", child)
			}
		}
	})

	t.Run("retained links are unresolved child evidence counted once", func(t *testing.T) {
		// Child 3: queued start and retained link, counted once. Child 4:
		// retained link only. Children 1 and 2 were re-proved.
		if partial.PendingChildren != 2 {
			t.Fatalf("unresolved child evidence = %d, want 2", partial.PendingChildren)
		}
		if partial.UnresolvedParentLinks != full.UnresolvedParentLinks {
			t.Fatalf("retained links changed parent links: %d, was %d", partial.UnresolvedParentLinks, full.UnresolvedParentLinks)
		}
		latest := registry.latestReconciliation(t)
		if latest.Attempt != domain.CaptureAttemptPartial || latest.PendingChildren != 2 {
			t.Fatalf("latest reconciliation = %s with %d unresolved children, want partial with 2", latest.Attempt, latest.PendingChildren)
		}
		if summary := capture.CoverageSummary(partial); !strings.Contains(summary, "Unresolved child evidence: 2;") {
			t.Fatalf("coverage summary does not report the retained links: %s", summary)
		}
	})

	t.Run("spawn events carry only freshly resolved children", func(t *testing.T) {
		events := spawnEventChildren(partial, family.parent)
		want := map[string]string{"call-spawn-01": family.children[0], "call-spawn-02": family.children[1], "call-spawn-03": "", "call-spawn-04": ""}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("spawn event children = %v, want %v", events, want)
		}
	})

	t.Run("a repeated partial pass leaves every snapshot identical", func(t *testing.T) {
		before := registry.snapshots(t)
		repeated := registry.reconcile(t, childLinkPartialBudget)
		if !reflect.DeepEqual(registry.snapshots(t), before) {
			t.Fatal("repeating the partial pass changed accepted snapshots")
		}
		if repeated.PendingChildren != partial.PendingChildren || repeated.UnresolvedParentLinks != partial.UnresolvedParentLinks {
			t.Fatalf("repeated coverage = %d/%d, want %d/%d", repeated.PendingChildren, repeated.UnresolvedParentLinks, partial.PendingChildren, partial.UnresolvedParentLinks)
		}
		if links, _ := childLinkLinks(repeated.RefreshedSessions, family.parent); !slices.Equal(links, family.children) {
			t.Fatalf("repeated refreshed links = %q", links)
		}
	})

	t.Run("a full pass re-proves every link and completes", func(t *testing.T) {
		var settled capture.ReconcileResult
		for pass := 0; pass < 6; pass++ {
			settled = registry.reconcile(t, fixtureByteBudget)
			if settled.PendingChildren == 0 && !settled.DiscoveryPending {
				break
			}
		}
		if settled.PendingChildren != 0 || settled.DiscoveryPending {
			t.Fatalf("full pass did not resolve the retained links: pending=%d discovery=%t", settled.PendingChildren, settled.DiscoveryPending)
		}
		links, _ := childLinkLinks(registry.snapshots(t)[family.parent], family.parent)
		if !slices.Equal(links, family.children) {
			t.Fatalf("full pass links = %q", links)
		}
	})
}

func TestReconcileFreshContradictingChildClearsItsRetainedLink(t *testing.T) {
	for _, tc := range []struct {
		name string
		// output replaces child 2's spawn result; empty keeps the task name.
		// An explicit child session is the case for which the Codex adapter
		// raises its contradiction warning; capture does not surface it.
		output string
	}{
		{name: "task-name result"},
		{name: "explicit child session result", output: "session"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry, _ := newChildLinkRegistry(t)
			family := registry.family
			outputs := map[int]string{}
			if tc.output != "" {
				outputs[1] = fmt.Sprintf(`{"session_id":%q}`, family.children[1])
			}
			family.writeParent(t, 41, outputs)
			// Child 2 now names another parent in its own metadata and is read
			// with the parent. Children 3 and 4 are not read.
			family.writeChild(t, 1, "019abcde-0000-7000-8000-00000000ffff", 22)
			partial := registry.reconcile(t, childLinkPartialBudget)
			read := refreshedIDs(partial)
			if !read[family.parent] || !read[family.children[1]] || read[family.children[2]] || read[family.children[3]] {
				t.Fatalf("fixture must read the parent and child 2 but not children 3 and 4, read %v", read)
			}
			want := []string{family.children[0], "", family.children[2], family.children[3]}
			if links, _ := childLinkLinks(registry.snapshots(t)[family.parent], family.parent); !slices.Equal(links, want) {
				t.Fatalf("registry links = %q, want %q", links, want)
			}
			if links, _ := childLinkLinks(partial.RefreshedSessions, family.parent); !slices.Equal(links, want) {
				t.Fatalf("refreshed links = %q, want %q", links, want)
			}
			// Only children 3 and 4 are retained without re-proof.
			if partial.PendingChildren != 2 {
				t.Fatalf("unresolved child evidence = %d, want 2", partial.PendingChildren)
			}
			if repeated := registry.reconcile(t, childLinkPartialBudget); !slices.Equal(func() []string { l, _ := childLinkLinks(repeated.RefreshedSessions, family.parent); return l }(), want) {
				t.Fatal("repeating the pass restored the contradicted link")
			}
		})
	}
}

// FIX-21 DATA-01: a pass that stops after the parent's own snapshot commit
// and before the family annotation commit must not leave the registry with
// fewer accepted links. The annotation write is made to fail with a registry
// trigger, the same seam the archived checkpoint tests use: the trigger aborts
// the first update of the parent's row after its refreshed snapshot (input
// 41) was committed.
func TestReconcileInterruptedAnnotationKeepsRegistryLinks(t *testing.T) {
	registry, _ := newChildLinkRegistry(t)
	family := registry.family
	db, err := sql.Open("sqlite", filepath.Join(registry.store.RegistryDirectory(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	input := "json_extract(%s.record, '$.snapshot[0].invocations[0].tokens.input')"
	if _, err := db.Exec(`CREATE TRIGGER interrupt_family_annotation BEFORE UPDATE ON registrations
WHEN NEW.session_id = '` + family.parent + `' AND ` + fmt.Sprintf(input, "OLD") + ` = 41 AND ` + fmt.Sprintf(input, "NEW") + ` = 41
BEGIN SELECT RAISE(ABORT, 'fixture interrupts the family annotation'); END`); err != nil {
		t.Fatal(err)
	}
	family.writeParent(t, 41)
	_, err = registry.store.Reconcile(context.Background(), childLinkTicket, "tester", 20, childLinkPartialBudget, fixtureGenerationPolicy())
	if err == nil || !strings.Contains(err.Error(), "fixture interrupts the family annotation") {
		t.Fatalf("the fixture must interrupt the pass at the family annotation, got %v", err)
	}
	stored := registry.snapshots(t)[family.parent]
	if got := parentInput(t, stored, family.parent); got != 41 {
		t.Fatalf("the parent's own snapshot must have committed before the interruption, input %d", got)
	}
	if links, _ := childLinkLinks(stored, family.parent); !slices.Equal(links, family.children) {
		t.Fatalf("interrupted pass left registry links %q, want every accepted link %q", links, family.children)
	}
	if _, err := db.Exec(`DROP TRIGGER interrupt_family_annotation`); err != nil {
		t.Fatal(err)
	}
	// The next partial pass still retains the unread children's links and
	// reports them as unresolved.
	next := registry.reconcile(t, childLinkPartialBudget)
	if links, _ := childLinkLinks(registry.snapshots(t)[family.parent], family.parent); !slices.Equal(links, family.children) {
		t.Fatalf("follow-up pass registry links = %q", links)
	}
	if links, _ := childLinkLinks(next.RefreshedSessions, family.parent); !slices.Equal(links, family.children) {
		t.Fatalf("follow-up pass refreshed links = %q", links)
	}
	if next.PendingChildren != 2 {
		t.Fatalf("follow-up unresolved child evidence = %d, want children 3 and 4", next.PendingChildren)
	}
}

// FIX-21 ordinal group retention: two spawns share agent type, description
// and invocation time, so their accepted links can only be placed by position.
// Their agent_id results resolve only when the children are read. Two
// consecutive partial passes keep both links in the registry and in the
// published owners, and count them as unresolved each time.
func TestReconcileAmbiguousGroupLinksSurviveRepeatedPartialPasses(t *testing.T) {
	s := openStore(t, t.TempDir())
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	active := filepath.Join(home, "sessions")
	if err := os.MkdirAll(active, 0o700); err != nil {
		t.Fatal(err)
	}
	parentPath := filepath.Join(active, "group-parent.jsonl")
	writeParent := func(tokens int) {
		t.Helper()
		var body strings.Builder
		fmt.Fprintf(&body, `{"timestamp":"2026-11-24T09:00:00Z","type":"session_meta","payload":{"id":"group-parent","timestamp":"2026-11-24T09:00:00Z","cwd":"/fixture/project","source":"cli"}}`+"\n")
		fmt.Fprintf(&body, `{"timestamp":"2026-11-24T09:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}`+"\n")
		for i, child := range []string{"group-child-1", "group-child-2"} {
			// Same agent type, arguments and timestamp: one key group.
			fmt.Fprintf(&body, `{"timestamp":"2026-11-24T09:00:05Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn-%d","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}`+"\n", i+1)
			fmt.Fprintf(&body, `{"timestamp":"2026-11-24T09:00:06Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn-%d","output":"{\"agent_id\":\"%s\"}"}}`+"\n", i+1, child)
		}
		fmt.Fprintf(&body, `{"timestamp":"2026-11-24T09:03:00Z","type":"token_usage_record","payload":{"thread_id":"group-parent","turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}`+"\n", tokens)
		if err := os.WriteFile(parentPath, []byte(body.String()), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeParent(11)
	registerBound(t, s, home, parentPath, "group-parent", childLinkTicket)
	children := []string{"group-child-1", "group-child-2"}
	for _, child := range children {
		path := filepath.Join(active, child+".jsonl")
		// Each child is larger than the partial budget.
		text := strings.Repeat("synthetic child output ", childLinkLargeBytes/23+1)
		body := fmt.Sprintf(`{"timestamp":"2026-11-24T09:00:06Z","type":"session_meta","payload":{"id":%q,"parent_thread_id":"group-parent","timestamp":"2026-11-24T09:00:06Z","cwd":"/fixture/project"}}`+"\n"+
			`{"timestamp":"2026-11-24T09:00:07Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}`+"\n"+
			`{"timestamp":"2026-11-24T09:00:08Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":%q}]}}`+"\n"+
			`{"timestamp":"2026-11-24T09:00:09Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":5,"output_tokens":2}}}`+"\n", child, text, child)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		registerBound(t, s, home, path, child, childLinkTicket)
	}
	registry := childLinkRegistry{store: s}
	full := registry.reconcile(t, fixtureByteBudget)
	if links, delegations := childLinkLinks(full.RefreshedSessions, "group-parent"); !slices.Equal(links, children) || delegations[0].InvokedAt == nil || *delegations[0].InvokedAt != *delegations[1].InvokedAt {
		t.Fatalf("fixture must accept two links in one key group: %q", links)
	}
	for pass := 1; pass <= 2; pass++ {
		writeParent(40 + pass)
		partial := registry.reconcile(t, childLinkPartialBudget)
		read := refreshedIDs(partial)
		if !read["group-parent"] || read[children[0]] || read[children[1]] {
			t.Fatalf("pass %d must read the parent without its children, read %v", pass, read)
		}
		if links, _ := childLinkLinks(partial.RefreshedSessions, "group-parent"); !slices.Equal(links, children) {
			t.Fatalf("pass %d published group links = %q, want %q", pass, links, children)
		}
		if links, _ := childLinkLinks(registry.snapshots(t)["group-parent"], "group-parent"); !slices.Equal(links, children) {
			t.Fatalf("pass %d registry group links = %q, want %q", pass, links, children)
		}
		if partial.PendingChildren != 2 {
			t.Fatalf("pass %d unresolved child evidence = %d, want both carried links", pass, partial.PendingChildren)
		}
	}
}
