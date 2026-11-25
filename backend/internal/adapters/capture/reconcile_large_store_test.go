package capture

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/codex/codexstoretest"
	"agenticlens/internal/domain"
)

// FIX-21 discovery budget on a large Codex store.
//
// The byte oracle is the fixture's known file sizes. Under the accepted rules
// a header read costs the bytes the header reader actually consumes: whole
// 4 KiB reads until the first session_meta line ends, never past the file or
// the 64 KiB structural bound. Each header is read at most once per pass, and
// an owner's extraction costs its file size. evidence/TESTS.md keeps the
// pre-change baseline these assertions replace.

const (
	largeStoreTicket = "TEST-100"
	largeStoreUser   = "large-store fixture"
	// incidentByteBudget is the incident's max_source_bytes at divisor 1.
	incidentByteBudget int64 = 512 << 20
	// largeStoreSessionBudget lets an exact lookup list the whole store, as the
	// incident's byte-bound failure implies.
	largeStoreSessionBudget = 4096
	// exampleSessionBudget is .agenticlens-config.local.example.yaml's value.
	exampleSessionBudget = 256
	unboundedBytes       = int64(1) << 40
	unboundedSessions    = 1 << 20
	// headerReadChunk is the header reader's read granularity.
	headerReadChunk = 4096
)

// largeStoreDivisor scales the store and its budget down together. The
// default keeps the package's race-enabled run reasonable; divisor 1 is the
// incident-scale store of about 1,900 rollouts and 3 GB.
var largeStoreDivisor = flag.Int("fix21.large-store-divisor", 4, "divisor applied to the FIX-21 incident-scale store and budget")

type largeStoreFixture struct {
	store   codexstoretest.LargeStore
	project string
	cutoff  time.Time
}

func newLargeStoreFixture(t *testing.T, spec func(project string) codexstoretest.LargeStoreSpec, byteBudget int64) largeStoreFixture {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data, err := fixtureSharedYAML(t, "Large Codex store", "records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("agenticlens:\n  state_dir: publication-state\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: .agenticlens-capture\n    max_sessions: %d\n    max_source_bytes: %d\n", largeStoreSessionBudget, byteBudget)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0o600); err != nil {
		t.Fatal(err)
	}
	generated := spec(project)
	return largeStoreFixture{store: codexstoretest.WriteLargeStore(t, generated), project: project, cutoff: generated.RootCreated.Add(24 * time.Hour)}
}

// smallStoreSpec keeps the incident store's names, dates and header size
// with a few unrelated rollouts and the given family sizes, so a byte budget
// can be set around single owners and headers.
func smallStoreSpec(rootBytes, childBytes int64) func(string) codexstoretest.LargeStoreSpec {
	return func(project string) codexstoretest.LargeStoreSpec {
		spec := codexstoretest.IncidentScaleSpec(project, 16)
		spec.RootBytes, spec.ChildBytes = rootBytes, childBytes
		return spec
	}
}

// openRegistry opens a fresh capture registry over the shared, read-only
// Codex home and binds the root. queued children get a SubagentStart hint, as
// Codex lifecycle hooks send; bound children are registered and bound
// directly, as an earlier SessionStart would have.
func (fx largeStoreFixture) openRegistry(t *testing.T, bytes int64, queued, bound []int) *Store {
	t.Helper()
	ctx := context.Background()
	s, err := OpenProject(t.TempDir(), fx.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := s.SetSourceBudgets(largeStoreSessionBudget, bytes); err != nil {
		t.Fatal(err)
	}
	register := func(id, path string) {
		t.Helper()
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: fx.store.Home, SessionID: id}
		row, err := s.Register(ctx, identity, path, "SessionStart")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.Bind(ctx, identity, largeStoreTicket, row.Request); err != nil {
			t.Fatal(err)
		}
	}
	register(fx.store.RootID, fx.store.RootPath)
	for _, i := range bound {
		register(fx.store.ChildIDs[i], fx.store.ChildPaths[i])
	}
	for _, i := range queued {
		if err := s.QueueChildStart(ctx, fx.store.Home, fx.store.RootID, fx.store.ChildIDs[i]); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

func allChildren(n int) []int {
	indexes := make([]int, n)
	for i := range indexes {
		indexes[i] = i
	}
	return indexes
}

// headerCost is the oracle for one header read.
func headerCost(file codexstoretest.File) int64 {
	chunks := (file.HeaderBytes + headerReadChunk - 1) / headerReadChunk
	return min(file.Size, chunks*headerReadChunk, int64(MaxPayloadBytes))
}

// unconventional reports the names outside the rollout convention. Capture
// exact lookups read them for every owner; names with another UUID are skipped.
func unconventional(file codexstoretest.File) bool {
	switch file.Note {
	case codexstoretest.NoteNoUUID, codexstoretest.NoteCustomName, codexstoretest.NoteNonDateDir, codexstoretest.NoteWindowNoUUID:
		return true
	}
	return false
}

// sweepOnlyHeaders are the in-window rollouts that only a child sweep reads:
// unrelated rollouts created after the root and the non-v7 name in the root's
// day directory, which exact lookups skip for its other UUID.
func (fx largeStoreFixture) sweepOnlyHeaders() int64 {
	var total int64
	for _, file := range fx.store.Files {
		if file.Later || file.Note == codexstoretest.NoteWindowNonV7 {
			total += headerCost(file)
		}
	}
	return total
}

func (fx largeStoreFixture) unconventionalHeaders() int64 {
	var total int64
	for _, file := range fx.store.Files {
		if unconventional(file) {
			total += headerCost(file)
		}
	}
	return total
}

func (fx largeStoreFixture) familyHeaders() int64 {
	var total int64
	for _, file := range fx.store.Files {
		if file.Family {
			total += headerCost(file)
		}
	}
	return total
}

func rootLinks(sessions []domain.SessionFact, rootID string) (links, delegations int) {
	for _, session := range sessions {
		if session.ID != rootID {
			continue
		}
		for _, delegation := range session.Delegations {
			delegations++
			if delegation.ChildSession != "" {
				links++
			}
		}
	}
	return links, delegations
}

func reconcileLarge(t *testing.T, s *Store, fx largeStoreFixture, sessions int, bytes int64) ReconcileResult {
	t.Helper()
	result, err := s.ReconcileAt(context.Background(), largeStoreTicket, largeStoreUser, sessions, bytes, fx.cutoff, fixtureGenerationPolicy())
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceBytes > bytes {
		t.Fatalf("pass charged %d above the %d budget", result.SourceBytes, bytes)
	}
	return result
}

func registrationsByID(result ReconcileResult) map[string]domain.CaptureRegistration {
	rows := map[string]domain.CaptureRegistration{}
	for _, row := range result.Registrations {
		rows[row.Identity.SessionID] = row
	}
	return rows
}

func latestReconciliation(t *testing.T, s *Store) domain.CaptureDiagnostic {
	t.Helper()
	diagnostics, err := s.Diagnostics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := len(diagnostics) - 1; i >= 0; i-- {
		if diagnostics[i].Reconciliation != nil && diagnostics[i].TicketID == largeStoreTicket {
			return diagnostics[i]
		}
	}
	t.Fatal("no reconciliation diagnostic")
	return domain.CaptureDiagnostic{}
}

func TestLargeStoreFamilyWithinCaptureBudget(t *testing.T) {
	divisor := *largeStoreDivisor
	if divisor < 1 || divisor > 16 {
		t.Fatalf("-fix21.large-store-divisor=%d must be from 1 through 16", divisor)
	}
	byteBudget := incidentByteBudget / int64(divisor)
	fx := newLargeStoreFixture(t, func(project string) codexstoretest.LargeStoreSpec {
		spec := codexstoretest.IncidentScaleSpec(project, divisor)
		// Unrelated sessions continue after the root, and two irregular names
		// sit in the root's own day directory, inside the sweep's window.
		spec.LaterUnrelated, spec.LaterSpan, spec.WindowIrregular = 6, 90*time.Minute, true
		return spec
	}, byteBudget)
	store := fx.store
	family := func(f codexstoretest.File) bool { return f.Family }
	familyBytes := store.TotalBytes(family)
	nonDate := store.Noted(codexstoretest.NoteNonDateDir)
	t.Logf("fixture: divisor=%d files=%d store bytes=%d family bytes=%d budget=%d", divisor, len(store.Files), store.TotalBytes(nil), familyBytes, byteBudget)

	t.Run("fixture family resolves completely", func(t *testing.T) {
		var transcripts []codex.Transcript
		for _, path := range append([]string{store.RootPath}, store.ChildPaths...) {
			transcript, err := codex.ParseRolloutPathAt(path, fx.cutoff)
			if err != nil || len(transcript.Warnings) > 0 {
				t.Fatalf("family rollout %s is not clean: %v %v", path, err, transcript.Warnings)
			}
			transcripts = append(transcripts, transcript)
		}
		if hints := transcripts[0].ChildDiscoveryHints(); len(hints) != 0 {
			t.Fatalf("root delegations must be task-name-only, got hints %v", hints)
		}
		combined, err := codex.CombineTranscriptsWithPolicy(transcripts, fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		sessions, err := combined.SessionFacts(fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		if links, delegations := rootLinks(sessions, store.RootID); links != len(store.ChildIDs) || delegations != len(store.ChildIDs) {
			t.Fatalf("complete family resolves %d of %d delegations", links, delegations)
		}
	})

	t.Run("capture exact lookup reads only matching and unconventional names", func(t *testing.T) {
		for _, tc := range []struct {
			name, id  string
			ambiguous bool
		}{
			{"root", store.RootID, false},
			{"child", store.ChildIDs[len(store.ChildIDs)-1], false},
			{"same-UUID copy", store.DuplicateID, true},
		} {
			// Oracle: sorted kept paths are read until the owner is decided.
			var kept []codexstoretest.File
			for _, file := range store.Files {
				if unconventional(file) || strings.Contains(filepath.Base(file.Path), tc.id) {
					kept = append(kept, file)
				}
			}
			sort.Slice(kept, func(i, j int) bool { return kept[i].Path < kept[j].Path })
			var wantBytes int64
			wantClaims, matches := 0, 0
			for _, file := range kept {
				wantClaims++
				wantBytes += headerCost(file)
				if file.SessionID == tc.id {
					if matches++; matches == 2 {
						break
					}
				}
			}
			for _, sessions := range []int{unboundedSessions, exampleSessionBudget} {
				budget := &codex.DiscoveryBudget{MaxSessions: sessions, MaxSourceBytes: unboundedBytes}
				path, err := codex.DiscoverExactCaptureRolloutPath(store.Home, tc.id, budget, codex.NewHeaderCache())
				if tc.ambiguous != errors.Is(err, codex.ErrOwnerAmbiguous) || (!tc.ambiguous && (err != nil || store.File(path).SessionID != tc.id)) {
					t.Fatalf("%s with max_sessions %d: %q %v", tc.name, sessions, path, err)
				}
				if budget.SessionsRead != wantClaims || budget.SourceBytesRead != wantBytes {
					t.Fatalf("%s: claims=%d bytes=%d, want %d and %d", tc.name, budget.SessionsRead, budget.SourceBytesRead, wantClaims, wantBytes)
				}
			}
		}
		// Live exact lookup is unchanged: it claims and reads every rollout.
		var oracle int64
		for _, file := range store.Files {
			if !file.Compressed {
				oracle += min(file.Size, int64(MaxPayloadBytes))
			}
		}
		budget := &codex.DiscoveryBudget{MaxSessions: unboundedSessions, MaxSourceBytes: unboundedBytes}
		if _, err := codex.DiscoverExactRolloutPathWithBudget(store.Home, store.RootID, budget); err != nil {
			t.Fatal(err)
		}
		if gzip := budget.SourceBytesRead - oracle; budget.SessionsRead != len(store.Files) || gzip <= 0 || gzip > int64(MaxPayloadBytes) {
			t.Fatalf("live lookup changed: claims=%d bytes=%d, want %d claims and %d plus one gzip header", budget.SessionsRead, budget.SourceBytesRead, len(store.Files), oracle)
		}
	})

	t.Run("child sweep reads only the window and takes no session slots", func(t *testing.T) {
		s := fx.openRegistry(t, byteBudget, allChildren(len(store.ChildIDs)), nil)
		budget := &workBudget{bytes: unboundedBytes, sessions: unboundedSessions, headers: codex.NewHeaderCache()}
		if _, err := s.recoverChildren(context.Background(), largeStoreTicket, budget, workStarts); err != nil {
			t.Fatal(err)
		}
		// The root and its children are in the parent's day directory with the
		// later unrelated rollouts and the in-window irregular names, and the
		// non-date directory is always listed. Every earlier rollout, including
		// unconventional names inside earlier day directories, costs nothing.
		want := fx.familyHeaders() + headerCost(nonDate) + headerCost(store.Noted(codexstoretest.NoteWindowNoUUID)) + fx.sweepOnlyHeaders()
		if budget.readBytes != want || budget.sessions != unboundedSessions || budget.attempts != 0 {
			t.Fatalf("sweep charged %d bytes, %d slots and %d attempts; want %d bytes and none", budget.readBytes, unboundedSessions-budget.sessions, budget.attempts, want)
		}
		rows, err := s.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		inherited := 0
		for _, row := range rows {
			if row.ParentID == store.RootID && row.TicketID == largeStoreTicket {
				inherited++
			}
		}
		if inherited != len(store.ChildIDs) {
			t.Fatalf("sweep inherited %d of %d children", inherited, len(store.ChildIDs))
		}
	})

	for _, sessions := range []int{largeStoreSessionBudget, exampleSessionBudget} {
		t.Run(fmt.Sprintf("family extracted within budget at max_sessions %d", sessions), func(t *testing.T) {
			s := fx.openRegistry(t, byteBudget, allChildren(len(store.ChildIDs)), nil)
			// Pass 1 reads the root, discovers every child and extracts them in
			// the final refresh. Every family header and every unconventional
			// name is charged once: the unconventional rollouts the sweep visits
			// were already read by the root's exact lookup in the same pass.
			// The sweep alone reads the later and non-v7 in-window rollouts.
			first := reconcileLarge(t, s, fx, sessions, byteBudget)
			if want := familyBytes + fx.familyHeaders() + fx.unconventionalHeaders() + fx.sweepOnlyHeaders(); first.SourceBytes != want || first.SourceAttempts != 1+len(store.ChildIDs) {
				t.Fatalf("pass 1: bytes=%d attempts=%d, want %d and %d", first.SourceBytes, first.SourceAttempts, want, 1+len(store.ChildIDs))
			}
			if len(first.RefreshedSessions) != 1+len(store.ChildIDs) {
				t.Fatalf("pass 1 refreshed %d owners, want the whole family", len(first.RefreshedSessions))
			}
			// Pass 2 reads the family together and links every child. Headers
			// are pass-local, so each is read once again.
			second := reconcileLarge(t, s, fx, sessions, byteBudget)
			if want := familyBytes + fx.familyHeaders() + fx.unconventionalHeaders(); second.SourceBytes != want || second.SourceAttempts != 1+len(store.ChildIDs) {
				t.Fatalf("pass 2: bytes=%d attempts=%d, want %d and %d", second.SourceBytes, second.SourceAttempts, want, 1+len(store.ChildIDs))
			}
			if links, _ := rootLinks(second.Sessions, store.RootID); links != len(store.ChildIDs) {
				t.Fatalf("pass 2 links %d of %d children", links, len(store.ChildIDs))
			}
			latest := latestReconciliation(t, s)
			if second.DiscoveryPending || second.PendingChildren != 0 || second.UnresolvedParentLinks != 0 || latest.Reconciliation.Attempt != domain.CaptureAttemptCompleted {
				t.Fatalf("pass 2 did not complete: %+v", latest.Reconciliation)
			}
			t.Logf("max_sessions %d: pass 1 %d bytes, pass 2 %d bytes (identification %d); baseline charged the full %d", sessions, first.SourceBytes, second.SourceBytes, second.SourceBytes-familyBytes, byteBudget)
		})
	}
}

func TestLargeStoreIdentificationAllowance(t *testing.T) {
	ctx := context.Background()
	h := func(fx largeStoreFixture, path string) int64 { return headerCost(fx.store.File(path)) }

	t.Run("extraction ahead of discovery keeps the reserved quarter", func(t *testing.T) {
		const budgetBytes int64 = 1 << 20
		fx := newLargeStoreFixture(t, smallStoreSpec(256<<10, 256<<10), budgetBytes)
		store := fx.store
		last := len(store.ChildIDs) - 1
		// Known owners could consume the whole budget; the last child is known
		// only from a SubagentStart hint.
		s := fx.openRegistry(t, budgetBytes, []int{last}, allChildren(last))
		if known := store.TotalBytes(func(f codexstoretest.File) bool { return f.Family && f.SessionID != store.ChildIDs[last] }); known <= budgetBytes {
			t.Fatalf("fixture known owners %d must exceed the %d budget", known, budgetBytes)
		}
		revision, err := s.captureControlRevision(ctx)
		if err != nil {
			t.Fatal(err)
		}
		// Phase 0 order: the refresh class runs before child discovery.
		budget := &workBudget{bytes: budgetBytes, sessions: largeStoreSessionBudget, controlRevision: revision, openCodeAttempts: map[string]bool{}, headers: codex.NewHeaderCache(), split: newBudgetSplit(budgetBytes)}
		result := ReconcileResult{StartedAt: fx.cutoff.UTC().Format(time.RFC3339Nano)}
		attempted := map[string]bool{}
		rows, accepted, err := s.refreshTicketOwners(ctx, largeStoreTicket, largeStoreUser, budget, &result, attempted, map[string]domain.CaptureRegistration{}, fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		quarter := budgetBytes / 4
		if budget.readBytes > budgetBytes-quarter || budget.split.identification != quarter {
			t.Fatalf("extraction read %d of %d and left identification %d; want at most %d and the whole quarter %d", budget.readBytes, budgetBytes, budget.split.identification, budgetBytes-quarter, quarter)
		}
		deferred := 0
		for _, row := range rows {
			if row.TicketID == largeStoreTicket && !accepted[IdentityKey(row.Identity)] {
				deferred++
				if attempted[IdentityKey(row.Identity)] || row.State != domain.CapturePending {
					t.Fatalf("owner refused only by the reservation was not deferred: %+v", row)
				}
			}
		}
		if deferred == 0 || budget.attempts != len(accepted) {
			t.Fatalf("deferred %d owners with %d attempts for %d extracted; deferred slots must be refunded", deferred, budget.attempts, len(accepted))
		}
		beforeSweep := budget.readBytes
		if _, err := s.recoverChildren(ctx, largeStoreTicket, budget, workStarts); err != nil {
			t.Fatal(err)
		}
		// The sweep reads, from the reserved quarter, the hinted child's header
		// and those of deferred owners whose preflight did not start.
		child := h(fx, store.ChildPaths[last])
		if got := budget.readBytes - beforeSweep; got < child || got > quarter || got%child != 0 || budget.split.identification != quarter-got {
			t.Fatalf("sweep charged %d, identification left %d; want whole child headers within the %d quarter", got, budget.split.identification, quarter)
		}
		inherited, err := s.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range inherited {
			found = found || (row.Identity.SessionID == store.ChildIDs[last] && row.ParentID == store.RootID && row.TicketID == largeStoreTicket)
		}
		if !found {
			t.Fatal("the hinted child was not discovered after extraction ran first")
		}
		budget.releaseSplit()
		_, leftover, err := s.refreshTicketOwners(ctx, largeStoreTicket, largeStoreUser, budget, &result, attempted, map[string]domain.CaptureRegistration{}, fixtureGenerationPolicy())
		if err != nil {
			t.Fatal(err)
		}
		// Each family header and unconventional name is charged once in the
		// pass, plus the size of every owner extracted in either refresh.
		extracted := fx.familyHeaders() + fx.unconventionalHeaders()
		for _, file := range store.Files {
			key := IdentityKey(domain.CaptureIdentity{Source: domain.SourceCodex, Store: store.Home, SessionID: file.SessionID})
			if file.Family && (accepted[key] || leftover[key]) {
				extracted += file.Size
			}
		}
		if budget.readBytes != extracted || budget.readBytes > budgetBytes {
			t.Fatalf("pass charged %d (leftover extracted %d owners), want %d within %d", budget.readBytes, len(leftover), extracted, budgetBytes)
		}
	})

	t.Run("unused identification bytes reach the final refresh", func(t *testing.T) {
		fx := newLargeStoreFixture(t, smallStoreSpec(512<<10, 64<<10), unboundedBytes>>20)
		store := fx.store
		root := store.File(store.RootPath)
		// The root alone needs more than extraction's three quarters, and fits
		// only with identification's unused quarter. One session: the deferred
		// root's slot must be refunded for the final refresh.
		need := h(fx, store.RootPath) + fx.unconventionalHeaders() + root.Size
		budgetBytes := need + 1024
		if need <= budgetBytes-budgetBytes/4 {
			t.Fatalf("fixture root %d must exceed extraction's share of %d", need, budgetBytes)
		}
		s := fx.openRegistry(t, budgetBytes, nil, nil)
		result := reconcileLarge(t, s, fx, 1, budgetBytes)
		if result.SourceBytes != need || result.SourceAttempts != 1 || len(result.RefreshedSessions) != 1 || result.RefreshedSessions[0].ID != store.RootID {
			t.Fatalf("bytes=%d attempts=%d refreshed=%d; want the root read with %d bytes in one attempt", result.SourceBytes, result.SourceAttempts, len(result.RefreshedSessions), need)
		}
		if row := registrationsByID(result)[store.RootID]; row.State != domain.CaptureCurrent {
			t.Fatalf("root = %s: %s", row.State, row.Detail)
		}
	})

	t.Run("identification stops at its ceiling while known owners are extracted", func(t *testing.T) {
		const minimumHeaderRead = int64(MaxPayloadBytes)
		fx := newLargeStoreFixture(t, smallStoreSpec(256<<10, 256<<10), unboundedBytes>>20)
		store := fx.store
		child := h(fx, store.ChildPaths[0])
		// Identification's quarter starts four child header reads; the next
		// read would not fit under the 64 KiB start rule.
		reads := 4
		quarter := int64(reads-1)*child + minimumHeaderRead
		budgetBytes := 4 * quarter
		root := store.File(store.RootPath)
		rootCost := h(fx, store.RootPath) + fx.unconventionalHeaders() + root.Size
		if rootCost > budgetBytes-quarter || budgetBytes-rootCost-int64(reads)*child >= store.File(store.ChildPaths[0]).Size {
			t.Fatalf("fixture sizes do not isolate the identification ceiling: root %d, budget %d", rootCost, budgetBytes)
		}
		s := fx.openRegistry(t, budgetBytes, allChildren(len(store.ChildIDs)), nil)
		result := reconcileLarge(t, s, fx, largeStoreSessionBudget, budgetBytes)
		if want := rootCost + int64(reads)*child; result.SourceBytes != want {
			t.Fatalf("bytes=%d, want root %d plus %d child headers", result.SourceBytes, rootCost, reads)
		}
		rows := registrationsByID(result)
		if rows[store.RootID].State != domain.CaptureCurrent {
			t.Fatalf("known root not extracted: %+v", rows[store.RootID])
		}
		for i, id := range store.ChildIDs {
			row, known := rows[id]
			if known != (i < reads) {
				t.Fatalf("child %d discovered=%t, want the first %d discovered", i, known, reads)
			}
			if known && row.State == domain.CaptureFailed {
				t.Fatalf("child %d failed instead of pending: %s", i, row.Detail)
			}
		}
		latest := latestReconciliation(t, s)
		if !result.DiscoveryPending || result.PendingChildren != len(store.ChildIDs)-reads || latest.Outcome == "failed" || latest.Reconciliation.Attempt != domain.CaptureAttemptPartial {
			t.Fatalf("discovery pending=%t children=%d outcome=%s attempt=%s; want pending, %d children, not failed", result.DiscoveryPending, result.PendingChildren, latest.Outcome, latest.Reconciliation.Attempt, len(store.ChildIDs)-reads)
		}
		// The next pass resumes discovery where identification stopped.
		next := reconcileLarge(t, s, fx, largeStoreSessionBudget, budgetBytes)
		discovered := 0
		for _, id := range store.ChildIDs {
			if _, ok := registrationsByID(next)[id]; ok {
				discovered++
			}
		}
		if discovered <= reads {
			t.Fatalf("the next pass did not continue discovery: %d children known", discovered)
		}
	})
}

// A child created just after local midnight sits in a later machine-local day
// directory than its parent's UTC date. The sweep still finds every child,
// and day directories that end more than 14 hours before the parent are not
// listed: unconventional names inside them, which are always read when listed,
// cost nothing.
func TestLargeStoreChildSweepAcrossLocalMidnight(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name             string
		through, created time.Time
	}{
		{"CET", time.Date(2026, 11, 23, 23, 0, 0, 0, berlin), time.Date(2026, 11, 24, 23, 59, 55, 0, berlin)},
		{"CEST month end", time.Date(2026, 6, 29, 23, 0, 0, 0, berlin), time.Date(2026, 6, 30, 23, 59, 55, 0, berlin)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newLargeStoreFixture(t, func(project string) codexstoretest.LargeStoreSpec {
				spec := codexstoretest.IncidentScaleSpec(project, 16)
				spec.UnrelatedThrough, spec.RootCreated = tc.through, tc.created
				return spec
			}, incidentByteBudget/16)
			store := fx.store
			first := store.File(store.ChildPaths[0])
			if local, utc := tc.created.Add(7*time.Second).Format("2006/01/02"), tc.created.UTC().Format("2006/01/02"); !strings.Contains(filepath.ToSlash(first.Path), "/sessions/"+local+"/") || local == utc {
				t.Fatalf("fixture child %s must sit in local day %s, after the parent's UTC day %s", first.Path, local, utc)
			}
			s := fx.openRegistry(t, incidentByteBudget/16, allChildren(len(store.ChildIDs)), nil)
			budget := &workBudget{bytes: unboundedBytes, sessions: unboundedSessions, headers: codex.NewHeaderCache()}
			if _, err := s.recoverChildren(context.Background(), largeStoreTicket, budget, workStarts); err != nil {
				t.Fatal(err)
			}
			if want := fx.familyHeaders() + headerCost(store.Noted(codexstoretest.NoteNonDateDir)); budget.readBytes != want {
				t.Fatalf("sweep charged %d, want %d: a pruned directory was listed or a family header was skipped", budget.readBytes, want)
			}
			rows, err := s.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			inherited := 0
			for _, row := range rows {
				if row.ParentID == store.RootID {
					inherited++
				}
			}
			if inherited != len(store.ChildIDs) {
				t.Fatalf("sweep inherited %d of %d children", inherited, len(store.ChildIDs))
			}
		})
	}
}

// relocateStore moves a generated store to home, which must not exist yet.
// Rollout contents never name their store, so only recorded paths change.
func relocateStore(t *testing.T, store codexstoretest.LargeStore, home string) codexstoretest.LargeStore {
	t.Helper()
	if err := os.Rename(store.Home, home); err != nil {
		t.Fatal(err)
	}
	move := func(path string) string { return home + strings.TrimPrefix(path, store.Home) }
	moved := store
	moved.Home, moved.RootPath = home, move(store.RootPath)
	moved.Files = append([]codexstoretest.File(nil), store.Files...)
	for i := range moved.Files {
		moved.Files[i].Path = move(moved.Files[i].Path)
	}
	moved.ChildPaths = nil
	for _, path := range store.ChildPaths {
		moved.ChildPaths = append(moved.ChildPaths, move(path))
	}
	return moved
}

// An owner deferred only by the identification reservation is retried with
// its refunded slot before a child discovered later in the same pass. With a
// single session the order decides which of them is read, so the case runs
// with the store placed where the child's registry key sorts after the
// root's, and where it sorts before.
func TestLargeStoreDeferredOwnerPrecedesLaterDiscoveryInFinalRefresh(t *testing.T) {
	for _, childAfterRoot := range []bool{true, false} {
		t.Run(fmt.Sprintf("child key after root key %t", childAfterRoot), func(t *testing.T) {
			fx := newLargeStoreFixture(t, func(project string) codexstoretest.LargeStoreSpec {
				spec := smallStoreSpec(512<<10, 64<<10)(project)
				spec.ChildDelays = spec.ChildDelays[:1]
				return spec
			}, unboundedBytes>>20)
			parent := t.TempDir()
			for n := 0; ; n++ {
				if n == 256 {
					t.Fatal("no store location gives the requested key order")
				}
				home := filepath.Join(parent, fmt.Sprintf("codex-home-%03d", n))
				canonical, err := CanonicalStore(parent)
				if err != nil {
					t.Fatal(err)
				}
				canonical = filepath.Join(canonical, filepath.Base(home))
				key := func(id string) string {
					return IdentityKey(domain.CaptureIdentity{Source: domain.SourceCodex, Store: canonical, SessionID: id})
				}
				if (key(fx.store.ChildIDs[0]) > key(fx.store.RootID)) == childAfterRoot {
					fx.store = relocateStore(t, fx.store, home)
					break
				}
			}
			store := fx.store
			h := func(path string) int64 { return headerCost(store.File(path)) }
			// The root needs more than extraction's three quarters during the
			// rotating rounds, and the child's discovery fits in the reserved
			// quarter; what remains afterwards covers the root.
			rootNeed := h(store.RootPath) + fx.unconventionalHeaders() + store.File(store.RootPath).Size
			budgetBytes := rootNeed + h(store.ChildPaths[0]) + 1024
			if rootNeed <= budgetBytes-budgetBytes/4 || budgetBytes/4 < int64(MaxPayloadBytes) {
				t.Fatalf("fixture sizes do not defer the root: need %d, budget %d", rootNeed, budgetBytes)
			}
			s := fx.openRegistry(t, budgetBytes, []int{0}, nil)
			result := reconcileLarge(t, s, fx, 1, budgetBytes)
			rows := registrationsByID(result)
			child, discovered := rows[store.ChildIDs[0]]
			if !discovered || child.ParentID != store.RootID {
				t.Fatalf("the child was not discovered in the same pass: %+v", rows)
			}
			if rows[store.RootID].State != domain.CaptureCurrent || len(result.RefreshedSessions) != 1 || result.RefreshedSessions[0].ID != store.RootID {
				t.Fatalf("the deferred root lost its refunded slot to the later child: root %s (%s), refreshed %d", rows[store.RootID].State, rows[store.RootID].Detail, len(result.RefreshedSessions))
			}
			if want := rootNeed + h(store.ChildPaths[0]); result.SourceAttempts != 1 || result.SourceBytes != want {
				t.Fatalf("attempts=%d bytes=%d, want 1 and %d", result.SourceAttempts, result.SourceBytes, want)
			}
			if child.State == domain.CaptureCurrent {
				t.Fatalf("the child was read without a session slot: %+v", child)
			}
		})
	}
}
