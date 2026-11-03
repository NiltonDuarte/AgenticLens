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
func persistHistoryAssociation(t *testing.T, s *Store, id, ticket, store string) {
	t.Helper()
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: id}
	// This adapter fixture models metadata already verified by Bind; CLI tests use
	// actual source-backed registration/binding and committed publication.
	_, err := s.Change(context.Background(), identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.ProjectProof = domain.CaptureProjectVerified
		r.ProjectDirectory = s.project
		r.Binding = "explicit_agent_response"
		r.TicketID = ticket
		return r, nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
func TestHistoryProvenanceRequiresActualRepositoryAndUnambiguousBinding(t *testing.T) {
	for _, kind := range []string{"deleted source and worktree", "latest source proof unavailable", "same remote unrelated repo", "copied Git objects", "missing proof after deletion", "wrong ticket", "competing store", "competing missing proof", "competing source", "unverified source", "current owner", "changed common identity", "replaced common directory"} {
		t.Run(kind, func(t *testing.T) {
			repo, a, b := historyRepository(t)
			registry := t.TempDir()
			source := t.TempDir()
			foreign, err := openRegistry(registry, "Shared", a, false)
			if err != nil {
				t.Fatal(err)
			}
			persistHistoryAssociation(t, foreign, "prior", "TEST-1", source)
			current, err := openRegistry(registry, "Shared", b, false)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := current.Close(); err != nil {
					t.Error(err)
				}
			}()
			proof, err := current.RepositoryProof(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "deleted source and worktree":
				runHistoryGit(t, repo, "worktree", "remove", "--force", a)
				if err = os.RemoveAll(source); err != nil {
					t.Fatal(err)
				}
			case "same remote unrelated repo", "copied Git objects":
				other := t.TempDir()
				runHistoryGit(t, other, "init", "-qb", "main")
				runHistoryGit(t, repo, "remote", "add", "origin", "https://example.invalid/same.git")
				runHistoryGit(t, other, "remote", "add", "origin", "https://example.invalid/same.git")
				if kind == "copied Git objects" {
					runHistoryGit(t, other, "fetch", repo, "HEAD")
					runHistoryGit(t, other, "reset", "--hard", "FETCH_HEAD")
				}
				// Replace only foreign provenance with a genuinely observed unrelated repo.
				ownership, err := currentWorktreeOwnership(other)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = foreign.db.Exec("UPDATE worktree_ownership SET ownership=? WHERE worktree_id=?", ownership, foreign.worktreeID); err != nil {
					t.Fatal(err)
				}
			case "missing proof after deletion":
				var body string
				if err = foreign.db.QueryRow("SELECT ownership FROM worktree_ownership WHERE worktree_id=?", foreign.worktreeID).Scan(&body); err != nil {
					t.Fatal(err)
				}
				var owner worktreeOwnership
				if err = json.Unmarshal([]byte(body), &owner); err != nil {
					t.Fatal(err)
				}
				owner.Repository = RepositoryProof{}
				raw, err := json.Marshal(owner)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = foreign.db.Exec("UPDATE worktree_ownership SET ownership=? WHERE worktree_id=?", raw, foreign.worktreeID); err != nil {
					t.Fatal(err)
				}
				runHistoryGit(t, repo, "worktree", "remove", "--force", a)
			case "wrong ticket":
				persistHistoryAssociation(t, foreign, "prior", "TEST-2", source)
			case "competing store":
				persistHistoryAssociation(t, foreign, "prior", "TEST-1", t.TempDir())
			case "competing missing proof":
				extra := t.TempDir()
				runHistoryGit(t, repo, "worktree", "add", "-qb", "extra", filepath.Join(extra, "worktree"))
				unknown, err := openRegistry(registry, "Shared", filepath.Join(extra, "worktree"), false)
				if err != nil {
					t.Fatal(err)
				}
				persistHistoryAssociation(t, unknown, "prior", "TEST-1", t.TempDir())
				if _, err = unknown.db.Exec("DELETE FROM worktree_ownership WHERE worktree_id=?", unknown.worktreeID); err != nil {
					t.Fatal(err)
				}
				if err = unknown.Close(); err != nil {
					t.Fatal(err)
				}
			case "competing source":
				identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: source, SessionID: "prior"}
				if _, err = foreign.Change(context.Background(), identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
					r.TicketID = "TEST-1"
					return r, nil
				}); err != nil {
					t.Fatal(err)
				}
			case "unverified source", "latest source proof unavailable":
				identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: source, SessionID: "prior"}
				if _, err = foreign.Change(context.Background(), identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
					r.ProjectProof = domain.CaptureProjectUnverified
					if kind == "unverified source" {
						r.ProjectDirectory = ""
					}
					return r, nil
				}); err != nil {
					t.Fatal(err)
				}
			case "current owner":
				persistHistoryAssociation(t, current, "prior", "TEST-1", source)
			case "changed common identity":
				proof.DirectoryIdentity = "replacement"
			case "replaced common directory":
				common := filepath.Join(repo, ".git")
				if err = os.Rename(common, common+".old"); err != nil {
					t.Fatal(err)
				}
				if err = os.Mkdir(common, 0700); err != nil {
					t.Fatal(err)
				}
				// Preserve administrative-directory identity while replacing its common owner.
				if err = os.Rename(filepath.Join(common+".old", "worktrees"), filepath.Join(common, "worktrees")); err != nil {
					t.Fatal(err)
				}
			}
			if err = foreign.Close(); err != nil {
				t.Fatal(err)
			}
			owners, err := current.AuthorizeCommittedHistory(context.Background(), "TEST-1", []domain.SessionFact{{ID: "prior", Source: domain.SourceCodex}}, proof)
			if kind == "deleted source and worktree" || kind == "latest source proof unavailable" {
				if err != nil || len(owners) != 1 {
					t.Fatalf("proven deleted history rejected: %+v %v", owners, err)
				}
			} else if err == nil {
				t.Fatalf("unsafe %s history accepted: %+v", kind, owners)
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

func TestAuxiliaryEvidencePreservesSnapshotOwnersAndRepeatedOccurrences(t *testing.T) {
	s, err := openRegistry(t.TempDir(), "Fixture", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	}()
	source := t.TempDir()
	for _, id := range []string{"one", "two", "other-ticket"} {
		ticket := "TEST-1"
		if id == "other-ticket" {
			ticket = "TEST-2"
		}
		persistHistoryAssociation(t, s, id, ticket, source)
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: source, SessionID: id}
		if _, err = s.Change(context.Background(), identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			skill := domain.SkillFact{FactProvenance: domain.FactProvenance{SourceInvocationID: "shared-call"}, Name: "shared-skill", Outcome: "unavailable"}
			link := domain.TaskLinkFact{FactProvenance: domain.FactProvenance{SourceInvocationID: "shared-spawn"}, ChildSession: "unavailable-child", CostEvidence: domain.UnavailableCost()}
			row.Snapshot = []domain.SessionFact{{ID: id, Source: domain.SourceCodex, Skills: []domain.SkillFact{skill, skill}, TaskLinks: []domain.TaskLinkFact{link}}}
			if id == "one" {
				row.Skills = []domain.SkillFact{{Name: "unresolved-legacy"}}
			}
			return row, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	facts, err := s.AuxiliaryEvidence(context.Background(), "TEST-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(facts) != 1 || len(facts[0].Skills) != 5 || len(facts[0].TaskLinks) != 2 {
		t.Fatalf("accepted snapshot facts missing, deduplicated or mixed across tickets: %+v", facts)
	}
	counts := map[string]int{}
	for _, skill := range facts[0].Skills {
		if skill.Name == "unresolved-legacy" {
			if skill.SessionID != "" || skill.Source != "" {
				t.Fatal("legacy row gained fabricated ownership")
			}
			continue
		}
		if skill.Source != domain.SourceCodex || skill.SourceInvocationID != "shared-call" || skill.Outcome != "unavailable" {
			t.Fatalf("snapshot source evidence changed: %+v", skill)
		}
		counts[skill.SessionID]++
	}
	if counts["one"] != 2 || counts["two"] != 2 || len(counts) != 2 {
		t.Fatalf("equal-looking calls crossed owners or lost multiplicity: %+v", counts)
	}
	for _, link := range facts[0].TaskLinks {
		if (link.SessionID != "one" && link.SessionID != "two") || link.Source != domain.SourceCodex || link.SourceInvocationID != "shared-spawn" || link.ChildSession != "unavailable-child" {
			t.Fatalf("snapshot delegation owner/source evidence changed: %+v", link)
		}
	}
}
