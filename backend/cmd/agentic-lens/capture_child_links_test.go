package main_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex/codexstoretest"
	"agenticlens/internal/domain"
)

// FIX-21: a ledger committed with every delegation link is refreshed by
// generate --capture-auto on a machine whose capture registry has never seen
// the children. The parent's task-name-only spawn results resolve only
// against children read in the same pass.

const (
	cliChildLinkTasks = 4
	// cliChildLinkLarge children cannot be extracted within cliChildLinkBudget.
	cliChildLinkLarge  = 300 * 1024
	cliChildLinkBudget = 160 * 1024
)

func cliParentLinks(ledger domain.AgenticFactLedger, parent string) []string {
	for _, session := range ledger.Sessions {
		if session.ID == parent {
			links := []string{}
			for _, delegation := range session.Delegations {
				links = append(links, delegation.ChildSession)
			}
			return links
		}
	}
	return nil
}

func cliSession(ledger domain.AgenticFactLedger, id string) (domain.SessionFact, bool) {
	for _, session := range ledger.Sessions {
		if session.ID == id {
			return session, true
		}
	}
	return domain.SessionFact{}, false
}

func runCLIChildLinks(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(cliBinaryPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %v\n%s", args[:3], err, out)
	}
	return out
}

// cliChildLinkRun is what one generate --capture-auto run on machine B must
// report: links kept without re-proof, accepted links it could not place, and
// the exact unresolved child evidence of the latest reconciliation.
type cliChildLinkRun struct {
	kept, refused, pending int
}

func TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState(t *testing.T) {
	for _, tc := range []struct {
		name string
		// queued children have a SubagentStart hint on machine B; children 1
		// and 2 are small and discovered, children 3 and 4 are too large for
		// the budget's identification share and stay unknown.
		queued []int
		// queuedPaths are child rollout paths queued on machine B before the
		// child is registered there.
		queuedPaths []int
		// contradicted children are registered on machine B and now name
		// another parent in their own metadata.
		contradicted []int
		// ambiguous makes the committed ledger give delegation 4 the agent
		// type, description and invocation time of delegation 3.
		ambiguous bool
		links     func(children []string) []string
		runs      []cliChildLinkRun
	}{
		{
			name:  "parent read alone",
			links: func(c []string) []string { return c },
			runs:  []cliChildLinkRun{{kept: 4, pending: 4}, {kept: 4, pending: 4}},
		},
		{
			// DATA-03: the queued path of the unregistered child 4 and its
			// retained link are one unresolved child.
			name:        "queued path of an unregistered child",
			queuedPaths: []int{3},
			links:       func(c []string) []string { return c },
			runs:        []cliChildLinkRun{{kept: 4, pending: 4}, {kept: 4, pending: 4}},
		},
		{
			// TEST-02: the first run discovers children 1 and 2 and reads them
			// after the parent. The second run reads them with the parent, so
			// their links resolve freshly, and keeps children 3 and 4. Child
			// 3's queued hint and its kept link count once.
			name:   "some children read with the parent on the second run",
			queued: []int{0, 1, 2},
			links:  func(c []string) []string { return c },
			runs:   []cliChildLinkRun{{kept: 4, pending: 4}, {kept: 2, pending: 2}},
		},
		{
			// TEST-06: fresh child metadata naming another parent clears the
			// accepted link through --capture-auto.
			name:         "contradicting child read on machine B",
			contradicted: []int{1},
			links:        func(c []string) []string { return []string{c[0], "", c[2], c[3]} },
			runs:         []cliChildLinkRun{{kept: 3, pending: 3}, {kept: 3, pending: 3}},
		},
		{
			// DATA-02: accepted links on ambiguous delegations are not moved,
			// and the run that refuses them reports them as unresolved. Its
			// publication holds the fresh delegations, so the next run has no
			// such accepted link left to refuse.
			name:      "ambiguous accepted provenance",
			ambiguous: true,
			links:     func(c []string) []string { return []string{c[0], c[1], "", ""} },
			runs:      []cliChildLinkRun{{kept: 2, refused: 2, pending: 4}, {kept: 2, refused: 0, pending: 2}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			f := newCaptureGenerationFixture(t)
			created := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
			family := codexstoretest.NewTaskFamily(t, f.store, f.project, created, cliChildLinkTasks, 2, cliChildLinkLarge)

			// Machine A: a full manual generation reads the parent with every
			// child and links each delegation from the current source alone.
			manual := []string{"generate", "--source", "codex", "--codex-home", f.store, "--capture-dir", filepath.Join(t.TempDir(), "machine-a"), "--session-id", family.ParentID,
				"--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "fix", "--json-output", f.output}
			runCLIChildLinks(t, f.project, manual...)
			committed := f.ledger(t)
			if links := cliParentLinks(committed, family.ParentID); !slices.Equal(links, family.ChildIDs) {
				t.Fatalf("full manual generation links = %q, want every child %q", links, family.ChildIDs)
			}
			// The committed ledger also holds an unrelated Codex owner and a
			// fact from another harness that this machine cannot read.
			parent := "unrelated-parent"
			committed.Sessions = append(committed.Sessions,
				domain.SessionFact{ID: "unrelated-root", Source: domain.SourceCodex, User: "machine-a", Invocations: []domain.InvocationFact{{Role: "coder", Turns: 1, Tokens: domain.InvocationTokens{Input: 5}}}, Tools: []domain.SessionToolFact{}, Delegations: []domain.DelegationFact{{Description: "unrelated", ChildSession: "unrelated-child"}}},
				domain.SessionFact{ID: "opencode-child", ParentID: &parent, Source: domain.SourceOpenCode, User: "machine-a", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}},
			)
			if tc.ambiguous {
				for i := range committed.Sessions {
					if committed.Sessions[i].ID == family.ParentID {
						d := committed.Sessions[i].Delegations
						d[3].AgentType, d[3].Description, d[3].InvokedAt = d[2].AgentType, d[2].Description, d[2].InvokedAt
					}
				}
			}
			if err := adapters.SaveFactLedger(f.output, committed); err != nil {
				t.Fatal(err)
			}
			committed = f.ledger(t)
			for _, args := range [][]string{{"add", "agentic_metrics.json"}, {"commit", "-qm", "committed ledger"}} {
				cmd := exec.Command("git", args...)
				cmd.Dir = f.project
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v %s", args, err, out)
				}
			}

			// Machine B: empty capture state, a partial byte budget, and only
			// the parent's own SessionStart registration plus the case's hints.
			if _, err := os.Stat(filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db")); !os.IsNotExist(err) {
				t.Fatalf("machine B must start without capture state: %v", err)
			}
			if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    max_source_bytes: 163840\n")); err != nil {
				t.Fatal(err)
			}
			family.WriteParent(t, 41, nil)
			const elsewhere = "019abcde-0000-7000-8000-00000000ffff"
			for _, i := range tc.contradicted {
				family.WriteChild(t, i, elsewhere, 22)
			}
			s, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			register := func(id string) error {
				identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: id}
				row, err := s.Register(ctx, identity, family.Path(id), "SessionStart")
				if err == nil {
					_, err = s.Bind(ctx, identity, "TEST-100", row.Request)
				}
				return err
			}
			err = register(family.ParentID)
			for _, i := range tc.contradicted {
				err = errorsJoin(err, register(family.ChildIDs[i]))
			}
			for _, i := range tc.queued {
				err = errorsJoin(err, s.QueueChildStart(ctx, f.store, family.ParentID, family.ChildIDs[i]))
			}
			for _, i := range tc.queuedPaths {
				err = errorsJoin(err, s.QueueChild(ctx, f.store, family.Path(family.ChildIDs[i]), family.ParentID))
			}
			if err = errorsJoinClose(err, s); err != nil {
				t.Fatal(err)
			}

			var previous []byte
			for run, want := range tc.runs {
				out := runCLIChildLinks(t, f.project, f.args(t)...)
				published := f.ledger(t)
				if links := cliParentLinks(published, family.ParentID); !slices.Equal(links, tc.links(family.ChildIDs)) {
					t.Fatalf("run %d published links = %q, want %q\n%s", run+1, links, tc.links(family.ChildIDs), out)
				}
				refreshedParent, _ := cliSession(published, family.ParentID)
				if len(refreshedParent.Invocations) != 1 || refreshedParent.Invocations[0].Tokens.Input != 41 {
					t.Fatalf("parent was not refreshed from machine B's source: %+v", refreshedParent.Invocations)
				}
				for _, id := range []string{"unrelated-root", "opencode-child", family.ChildIDs[2], family.ChildIDs[3]} {
					want, _ := cliSession(committed, id)
					got, ok := cliSession(published, id)
					if !ok || !reflect.DeepEqual(got, want) {
						t.Fatalf("owner %s not read on machine B changed: %+v, want %+v", id, got, want)
					}
				}
				for _, i := range tc.contradicted {
					if child, _ := cliSession(published, family.ChildIDs[i]); child.ParentID == nil || *child.ParentID != elsewhere {
						t.Fatalf("contradicting child was not published from its fresh metadata: %+v", child)
					}
				}
				line := fmt.Sprintf("Capture kept %d accepted delegation child link(s)", want.kept)
				if !strings.Contains(string(out), line) || !strings.Contains(string(out), fmt.Sprintf("could not place %d accepted link(s)", want.refused)) {
					t.Fatalf("run %d did not report %d kept and %d refused links:\n%s", run+1, want.kept, want.refused, out)
				}

				// Kept and refused links are unresolved child evidence, counted
				// once beside queued work, and the latest reconciliation is
				// partial.
				s, err := capture.OpenProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}
				diagnostics, err := s.Diagnostics(ctx)
				rows, listErr := s.List(ctx)
				if err = errorsJoinClose(errorsJoin(err, listErr), s); err != nil {
					t.Fatal(err)
				}
				var latest *domain.CaptureReconciliationDiagnostic
				for i := len(diagnostics) - 1; i >= 0 && latest == nil; i-- {
					if diagnostics[i].TicketID == "TEST-100" && diagnostics[i].Reconciliation != nil {
						latest = diagnostics[i].Reconciliation
					}
				}
				registered := map[string]bool{}
				for _, row := range rows {
					registered[row.Identity.SessionID] = true
				}
				if registered[family.ChildIDs[2]] || registered[family.ChildIDs[3]] {
					t.Fatal("fixture must leave children 3 and 4 unknown to machine B's registry")
				}
				if latest == nil || latest.Attempt != domain.CaptureAttemptPartial || latest.PendingChildren != want.pending {
					t.Fatalf("run %d latest reconciliation = %+v, want partial with %d unresolved children", run+1, latest, want.pending)
				}
				// A repeated run with unchanged sources keeps the ledger bytes,
				// including every timestamp.
				current := readCaptureFile(t, f.output)
				if previous != nil && !bytes.Equal(previous, current) {
					t.Fatalf("run %d changed the ledger:\nbefore %s\nafter  %s", run+1, previous, current)
				}
				previous = current
			}
		})
	}
}

// FIX-21 TEST-06: manual generation resolves links from the current source
// alone. A link to a child whose rollout is gone is left empty, not carried
// forward from the existing ledger.
func TestManualGenerateLeavesLinkToMissingChildEmpty(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	family := codexstoretest.NewTaskFamily(t, f.store, f.project, time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC), cliChildLinkTasks, cliChildLinkTasks, 0)
	manual := []string{"generate", "--source", "codex", "--codex-home", f.store, "--capture-dir", filepath.Join(t.TempDir(), "manual"), "--session-id", family.ParentID,
		"--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "fix", "--json-output", f.output}
	runCLIChildLinks(t, f.project, manual...)
	if links := cliParentLinks(f.ledger(t), family.ParentID); !slices.Equal(links, family.ChildIDs) {
		t.Fatalf("full manual generation links = %q", links)
	}
	if err := os.Remove(family.Path(family.ChildIDs[3])); err != nil {
		t.Fatal(err)
	}
	out := runCLIChildLinks(t, f.project, manual...)
	want := []string{family.ChildIDs[0], family.ChildIDs[1], family.ChildIDs[2], ""}
	if links := cliParentLinks(f.ledger(t), family.ParentID); !slices.Equal(links, want) {
		t.Fatalf("manual generation links = %q, want %q", links, want)
	}
	if strings.Contains(string(out), "Capture kept") {
		t.Fatalf("manual generation ran capture retention:\n%s", out)
	}
}

func errorsJoin(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func errorsJoinClose(err error, s *capture.Store) error {
	return errorsJoin(err, s.Close())
}
