package main_test

import (
	"bytes"
	"context"
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

func TestCaptureAutoKeepsCommittedChildLinksWithoutCaptureState(t *testing.T) {
	for _, tc := range []struct {
		name string
		// queued children have a SubagentStart hint on machine B, so its
		// pass discovers and reads them. The others are unknown to its registry.
		queued []int
	}{
		{name: "parent read alone"},
		{name: "parent read with some children", queued: []int{0, 1}},
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
			// the parent's own SessionStart registration.
			if _, err := os.Stat(filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db")); !os.IsNotExist(err) {
				t.Fatalf("machine B must start without capture state: %v", err)
			}
			if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    max_source_bytes: 163840\n")); err != nil {
				t.Fatal(err)
			}
			family.WriteParent(t, 41, nil)
			s, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: family.ParentID}
			row, err := s.Register(ctx, identity, family.Path(family.ParentID), "SessionStart")
			if err == nil {
				_, err = s.Bind(ctx, identity, "TEST-100", row.Request)
			}
			for _, i := range tc.queued {
				if err == nil {
					err = s.QueueChildStart(ctx, f.store, family.ParentID, family.ChildIDs[i])
				}
			}
			if err = errorsJoinClose(err, s); err != nil {
				t.Fatal(err)
			}

			out := runCLIChildLinks(t, f.project, f.args(t)...)
			published := f.ledger(t)
			if links := cliParentLinks(published, family.ParentID); !slices.Equal(links, family.ChildIDs) {
				t.Fatalf("published links = %q, want every accepted link %q\n%s", links, family.ChildIDs, out)
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
			if !strings.Contains(string(out), "Capture kept 4 accepted delegation child link(s)") {
				t.Fatalf("generation did not report the retained links:\n%s", out)
			}

			// The retained links are unresolved child evidence, including the
			// children this registry never registered, and the latest
			// reconciliation is partial.
			s, err = capture.OpenProject(f.state, f.project)
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
			if latest == nil || latest.Attempt != domain.CaptureAttemptPartial || latest.PendingChildren < 2 {
				t.Fatalf("latest reconciliation = %+v, want partial with the unregistered children unresolved", latest)
			}
			t.Logf("latest reconciliation %s with unresolved child evidence %d", latest.Attempt, latest.PendingChildren)
			if len(tc.queued) == 0 && latest.PendingChildren != cliChildLinkTasks {
				t.Fatalf("unresolved child evidence = %d, want all %d links unknown to this registry", latest.PendingChildren, cliChildLinkTasks)
			}

			// A repeated run with unchanged sources keeps the ledger bytes,
			// including every timestamp.
			before := readCaptureFile(t, f.output)
			runCLIChildLinks(t, f.project, f.args(t)...)
			if after := readCaptureFile(t, f.output); !bytes.Equal(before, after) {
				t.Fatalf("repeated partial generation changed the ledger:\nbefore %s\nafter  %s", before, after)
			}
		})
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
