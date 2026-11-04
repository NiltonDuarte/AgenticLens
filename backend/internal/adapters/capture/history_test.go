package capture

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func historyRepository(t *testing.T) (string, string, string) {
	t.Helper()
	repo := t.TempDir()
	runHistoryGit(t, repo, "init", "-qb", "main")
	if err := os.WriteFile(filepath.Join(repo, "seed"), []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	runHistoryGit(t, repo, "add", "seed")
	runHistoryGit(t, repo, "-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "seed")
	a, b := filepath.Join(t.TempDir(), "a"), filepath.Join(t.TempDir(), "b")
	runHistoryGit(t, repo, "worktree", "add", "-qb", "a", a)
	runHistoryGit(t, repo, "worktree", "add", "-qb", "b", b)
	return repo, a, b
}
func runHistoryGit(t *testing.T, project string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = project
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestSessionAssociationChecksActualConflictsWithoutPrivateHistoryGate(t *testing.T) {
	for _, tc := range []struct {
		name, ticket         string
		source               domain.SourceHarness
		registered, conflict bool
	}{
		{name: "portable owner has no private registration"},
		{name: "same source and ticket", ticket: "TEST-1", source: domain.SourceCodex, registered: true},
		{name: "wrong ticket", ticket: "TEST-2", source: domain.SourceCodex, registered: true, conflict: true},
		{name: "different harness has an independent native identity", ticket: "TEST-1", source: domain.SourceOpenCode, registered: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := openRegistry(t.TempDir(), "Shared", t.TempDir(), false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := s.Close(); err != nil {
					t.Error(err)
				}
			}()
			selectedStore := t.TempDir()
			if tc.registered {
				id := domain.CaptureIdentity{Source: tc.source, Store: selectedStore, SessionID: "owner"}
				if _, err := s.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
					r.TicketID = tc.ticket
					r.Binding = "explicit_manual_selection"
					return r, nil
				}); err != nil {
					t.Fatal(err)
				}
			}
			err = s.RejectSessionAssociations(context.Background(), "TEST-1", selectedStore, []domain.SessionFact{{ID: "owner", Source: domain.SourceCodex}})
			if (err != nil) != tc.conflict {
				t.Fatalf("association conflict=%t err=%v", tc.conflict, err)
			}
		})
	}
}

func TestWorktreeRepositoryProofEnrichesOnlyUnchangedOwnership(t *testing.T) {
	_, a, _ := historyRepository(t)
	current, err := currentWorktreeOwnership(a)
	if err != nil {
		t.Fatal(err)
	}
	var owner worktreeOwnership
	if err = json.Unmarshal([]byte(current), &owner); err != nil {
		t.Fatal(err)
	}
	if owner.Repository.CommonDirectory == "" || owner.Repository.DirectoryIdentity == "" {
		t.Fatal("new scope lacks durable repository proof")
	}
	prior := owner
	prior.Repository = RepositoryProof{}
	raw, err := json.Marshal(prior)
	if err != nil {
		t.Fatal(err)
	}
	same, err := sameWorktreeOwnership(string(raw), current)
	if err != nil || !same {
		t.Fatalf("verified live enrichment rejected: %v %v", same, err)
	}
	owner.Repository.DirectoryIdentity = "replacement"
	changed, err := json.Marshal(owner)
	if err != nil {
		t.Fatal(err)
	}
	same, err = sameWorktreeOwnership(current, string(changed))
	if err != nil || same {
		t.Fatalf("replaced repository proof accepted: %v %v", same, err)
	}
}
