package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/codex/codexstoretest"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// FIX-21 child sweep bookkeeping at the identification ceiling, driven
// through recoverChildren with an exact identification share.

const sweepTicket = "TEST-100"

var sweepCreated = time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)

func openSweepStore(t *testing.T) *Store {
	t.Helper()
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(t.TempDir(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := s.SetSourceBudgets(256, 64<<20); err != nil {
		t.Fatal(err)
	}
	return s
}

func bindSweepParent(t *testing.T, s *Store, home, id, path string) domain.CaptureRegistration {
	t.Helper()
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: id}
	row, err := s.Register(context.Background(), identity, path, "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	row, err = s.Bind(context.Background(), identity, sweepTicket, row.Request)
	if err != nil {
		t.Fatal(err)
	}
	return row
}

// identificationBudget allows exactly bytes of header identification and an
// ample extraction share, as the rotating rounds of a pass would.
func identificationBudget(bytes int64) *workBudget {
	return &workBudget{bytes: 1 << 30, sessions: 64, openCodeAttempts: map[string]bool{}, headers: codex.NewHeaderCache(), split: &budgetSplit{identification: bytes, extraction: 1 << 29}}
}

func (s *Store) sweepCursor(t *testing.T, parent domain.CaptureRegistration) (string, bool) {
	t.Helper()
	var path string
	err := s.db.QueryRow(`SELECT path FROM discovery_cursor WHERE worktree_id=? AND parent=?`, s.worktreeID, IdentityKey(parent.Identity)).Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return path, true
}

func (s *Store) sweepStarts(t *testing.T, parent domain.CaptureRegistration) int {
	t.Helper()
	var count int
	if err := s.db.QueryRow(`SELECT count(*) FROM pending_starts WHERE worktree_id=? AND parent=?`, s.worktreeID, parent.Identity.SessionID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func (s *Store) inheritedChildren(t *testing.T, parent string) map[string]bool {
	t.Helper()
	rows, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	children := map[string]bool{}
	for _, row := range rows {
		if row.ParentID == parent && row.TicketID == sweepTicket {
			children[row.Identity.SessionID] = true
		}
	}
	return children
}

// BEH-01: the second parent resumes from a file cursor, and the first parent
// has left identification below the 64 KiB start bound of its next candidate.
// The cursor must stay where it was, the sweep pending, and a later automatic
// pass (ReconcileAt, which never reopens a closed sweep) must find every child.
func TestResumedSweepAtTheCeilingKeepsItsFileCursor(t *testing.T) {
	ctx := context.Background()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// Children carry large bodies, so each header read needs 64 KiB of
	// allowance to start but costs only one 4 KiB read.
	families := []*codexstoretest.TaskFamily{
		codexstoretest.NewTaskFamily(t, home, "/fixture/project", sweepCreated, 4, 0, 70<<10),
		codexstoretest.NewTaskFamily(t, home, "/fixture/project", sweepCreated.Add(10*time.Minute), 4, 0, 70<<10),
	}
	s := openSweepStore(t)
	parents := map[string]domain.CaptureRegistration{}
	byParent := map[string]*codexstoretest.TaskFamily{}
	for _, family := range families {
		parents[family.ParentID] = bindSweepParent(t, s, home, family.ParentID, family.Path(family.ParentID))
		byParent[family.ParentID] = family
	}
	// Parents are swept in registry key order; the second one gets a real
	// file cursor first, from a pass in which it is the only eligible parent.
	order := []string{families[0].ParentID, families[1].ParentID}
	sort.Slice(order, func(i, j int) bool {
		return IdentityKey(parents[order[i]].Identity) < IdentityKey(parents[order[j]].Identity)
	})
	first, second := parents[order[0]], parents[order[1]]
	queue := func(parent domain.CaptureRegistration) {
		t.Helper()
		for _, child := range byParent[parent.Identity.SessionID].ChildIDs {
			if err := s.QueueChildStart(ctx, home, parent.Identity.SessionID, child); err != nil {
				t.Fatal(err)
			}
		}
	}
	share := int64(MaxPayloadBytes) + 4096
	queue(second)
	if _, err := s.recoverChildren(ctx, sweepTicket, identificationBudget(share), workStarts); err != nil {
		t.Fatal(err)
	}
	resumeAt, exists := s.sweepCursor(t, second)
	if !exists || !strings.HasSuffix(resumeAt, ".jsonl") || s.sweepStarts(t, second) == 0 {
		t.Fatalf("fixture must leave the second parent pending at a file cursor, got %q (exists %t, %d hints)", resumeAt, exists, s.sweepStarts(t, second))
	}
	hints := s.sweepStarts(t, second)
	known := len(s.inheritedChildren(t, second.Identity.SessionID))

	queue(first)
	budget := identificationBudget(share)
	if _, err := s.recoverChildren(ctx, sweepTicket, budget, workStarts); err != nil {
		t.Fatal(err)
	}
	if cursor, _ := s.sweepCursor(t, first); cursor == "" {
		t.Fatal("fixture must leave the first parent pending as well")
	}
	if !budget.discoveryPending {
		t.Fatal("a sweep stopped at the ceiling must leave discovery pending")
	}
	cursor, exists := s.sweepCursor(t, second)
	if !exists || cursor != resumeAt {
		t.Fatalf("resumed sweep that admitted nothing moved its cursor from %q to %q (row %t)", resumeAt, cursor, exists)
	}
	if s.sweepStarts(t, second) != hints || len(s.inheritedChildren(t, second.Identity.SessionID)) != known {
		t.Fatal("a sweep that admitted nothing changed its hints or discoveries")
	}

	// Automatic passes resume both sweeps and find every child.
	for pass := 0; pass < 6; pass++ {
		result, err := s.ReconcileAt(ctx, sweepTicket, "tester", 64, 64<<20, sweepCreated.Add(24*time.Hour), fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		if !result.DiscoveryPending {
			break
		}
	}
	for _, family := range families {
		found := s.inheritedChildren(t, family.ParentID)
		for i, child := range family.ChildIDs {
			if !found[child] {
				t.Fatalf("child %d of %s was never discovered by automatic passes", i, family.ParentID)
			}
		}
	}
}

// writeNonV7Parent writes a parent whose identity is a version 4 UUID, so its
// creation instant can only come from its session_meta timestamp.
func writeNonV7Parent(t *testing.T, home string, created time.Time) (string, string) {
	t.Helper()
	id := codexstoretest.UUID(created, 4, rand.New(rand.NewPCG(7, 7)))
	path := codexstoretest.ConventionalPath(home, created, id, false)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":"/fixture/project","source":"cli"}}
{"timestamp":%q,"type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}
`, codexstoretest.Timestamp(created), id, codexstoretest.Timestamp(created), codexstoretest.Timestamp(created.Add(time.Second)))
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return id, path
}

// writeUnrelatedV7 writes a small conventional rollout of another session.
func writeUnrelatedV7(t *testing.T, home string, created time.Time, seed uint64) string {
	t.Helper()
	id := codexstoretest.UUID(created, 7, rand.New(rand.NewPCG(seed, seed)))
	path := codexstoretest.ConventionalPath(home, created, id, false)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":"/fixture/other","source":"cli"}}`+"\n", codexstoretest.Timestamp(created), id, codexstoretest.Timestamp(created))
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() >= 4096 {
		t.Fatalf("%s must be smaller than one header read chunk", path)
	}
	return info.Size()
}

// nonV7Family writes a non-v7 parent with three conventional v7 children, an
// earlier rollout in the parent's own day directory, one in an earlier day
// directory and a later unrelated rollout inside the window.
func nonV7Family(t *testing.T) (home, parentID, parentPath string, family *codexstoretest.TaskFamily, earlier []string, later string) {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	family = codexstoretest.NewTaskFamily(t, home, "/fixture/project", sweepCreated, 3, 3, 0)
	if err := os.Remove(family.Path(family.ParentID)); err != nil {
		t.Fatal(err)
	}
	parentID, parentPath = writeNonV7Parent(t, home, sweepCreated)
	for i := range family.ChildIDs {
		family.WriteChild(t, i, parentID, 20+i)
	}
	earlier = []string{writeUnrelatedV7(t, home, sweepCreated.Add(-time.Hour), 31), writeUnrelatedV7(t, home, sweepCreated.AddDate(0, 0, -3), 32)}
	later = writeUnrelatedV7(t, home, sweepCreated.Add(20*time.Minute), 33)
	return home, parentID, parentPath, family, earlier, later
}

// TEST-03: without a UUIDv7 identity, the parent's session_meta timestamp
// sets the window. Its children are found, the later unrelated rollout is
// read, and the earlier rollouts cost nothing.
func TestChildSweepWindowFromNonV7ParentTimestamp(t *testing.T) {
	ctx := context.Background()
	home, parentID, parentPath, family, earlier, later := nonV7Family(t)
	s := openSweepStore(t)
	parent := bindSweepParent(t, s, home, parentID, parentPath)
	for _, child := range family.ChildIDs {
		if err := s.QueueChildStart(ctx, home, parentID, child); err != nil {
			t.Fatal(err)
		}
	}
	budget := &workBudget{bytes: 1 << 30, sessions: 64, openCodeAttempts: map[string]bool{}, headers: codex.NewHeaderCache()}
	if _, err := s.recoverChildren(ctx, sweepTicket, budget, workStarts); err != nil {
		t.Fatal(err)
	}
	want := fileSize(t, parentPath) + fileSize(t, later)
	for _, child := range family.ChildIDs {
		want += fileSize(t, family.Path(child))
	}
	for _, path := range earlier {
		fileSize(t, path)
	}
	if budget.readBytes != want {
		t.Fatalf("sweep charged %d, want %d: the parent header, its children and the later rollout only", budget.readBytes, want)
	}
	found := s.inheritedChildren(t, parentID)
	for i, child := range family.ChildIDs {
		if !found[child] {
			t.Fatalf("child %d of the non-v7 parent was pruned", i)
		}
	}
	if cursor, exists := s.sweepCursor(t, parent); !exists || cursor != "" || s.sweepStarts(t, parent) != 0 {
		t.Fatalf("completed sweep left cursor %q (row %t) and %d hints", cursor, exists, s.sweepStarts(t, parent))
	}
}

// BEH-01: a brand-new sweep that admits nothing writes no cursor row, so it is
// not closed. Reading the non-v7 parent's own header for its window uses the
// whole identification share before any candidate.
func TestBrandNewSweepAdmittingNothingWritesNoCursor(t *testing.T) {
	ctx := context.Background()
	home, parentID, parentPath, family, _, _ := nonV7Family(t)
	s := openSweepStore(t)
	parent := bindSweepParent(t, s, home, parentID, parentPath)
	for _, child := range family.ChildIDs {
		if err := s.QueueChildStart(ctx, home, parentID, child); err != nil {
			t.Fatal(err)
		}
	}
	budget := identificationBudget(fileSize(t, parentPath))
	if _, err := s.recoverChildren(ctx, sweepTicket, budget, workStarts); err != nil {
		t.Fatal(err)
	}
	if budget.split.identification != 0 || !budget.discoveryPending {
		t.Fatalf("fixture must spend the share on the parent header and leave discovery pending: left %d", budget.split.identification)
	}
	if cursor, exists := s.sweepCursor(t, parent); exists {
		t.Fatalf("a sweep that admitted nothing wrote cursor %q", cursor)
	}
	if len(s.inheritedChildren(t, parentID)) != 0 || s.sweepStarts(t, parent) != len(family.ChildIDs) {
		t.Fatal("a sweep that admitted nothing changed discovery state")
	}
	if _, err := s.recoverChildren(ctx, sweepTicket, identificationBudget(1<<20), workStarts); err != nil {
		t.Fatal(err)
	}
	if found := s.inheritedChildren(t, parentID); len(found) != len(family.ChildIDs) {
		t.Fatalf("the next sweep found %d of %d children", len(found), len(family.ChildIDs))
	}
}
