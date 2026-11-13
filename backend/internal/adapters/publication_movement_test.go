package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// Atomic staging, replacement and finalization must all follow the held home.
// A replacement at the previous pathname belongs to somebody else.
func TestPublicationTransactionFollowsRenamedDirectory(t *testing.T) {
	for _, outcome := range []string{"commit", "recover"} {
		t.Run(outcome, func(t *testing.T) {
			project := t.TempDir()
			writePublicationTestPolicy(t, project)
			home := filepath.Join(project, "selected")
			moved := filepath.Join(project, "archive", "arbitrary-label")
			if err := os.MkdirAll(filepath.Dir(moved), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(home, 0700); err != nil {
				t.Fatal(err)
			}
			ledger := filepath.Join(home, "agentic_metrics.json")
			summary := filepath.Join(home, "delivery.md")
			prior := map[string][]byte{"agentic_metrics.json": []byte("accepted ledger\n"), "delivery.md": []byte("accepted human evidence\n")}
			next := map[string][]byte{"agentic_metrics.json": []byte("replacement ledger\n"), "delivery.md": []byte("replacement human evidence\n")}
			for name, body := range prior {
				if err := os.WriteFile(filepath.Join(home, name), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			owner, err := capture.AcquirePublication(context.Background(), ledger, project)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if owner != nil {
					if err := owner.Close(); err != nil {
						t.Error(err)
					}
				}
			})
			paths := []string{ledger, summary}
			if err := BeginPublication(ledger, paths); err != nil {
				t.Fatal(err)
			}
			marker, err := os.ReadFile(journalPath(ledger))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(home, moved); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(home, 0700); err != nil {
				t.Fatal(err)
			}
			sentinel := []byte("unrelated replacement must remain exact\n")
			if err := os.WriteFile(filepath.Join(home, "unrelated"), sentinel, 0600); err != nil {
				t.Fatal(err)
			}
			for name, body := range next {
				if err := durableReplace(filepath.Join(home, name), body); err != nil {
					t.Fatal(err)
				}
			}
			assertPublicationBytes(t, filepath.Join(moved, "agentic_metrics.json.publication-pending"), marker)
			if err := CheckPublication(ledger); err == nil {
				t.Fatal("uncommitted moved set was accepted")
			}
			if outcome == "commit" {
				if err := CommitPublication(ledger, paths); err != nil {
					t.Fatal(err)
				}
			} else {
				for name, want := range next {
					assertPublicationBytes(t, filepath.Join(moved, name), want)
				}
				// Discard the old lease, as a restarted process must resolve the new home.
				if err := owner.Close(); err != nil {
					t.Fatal(err)
				}
				owner = nil
				ledger = filepath.Join(moved, "agentic_metrics.json")
				paths = []string{ledger, filepath.Join(moved, "delivery.md")}
				owner, err = capture.AcquirePublication(context.Background(), ledger, project)
				if err != nil {
					t.Fatal(err)
				}
				if recovered, err := RecoverPublication(ledger, paths); err != nil || !recovered {
					t.Fatalf("moved recovery: %v %v", recovered, err)
				}
				if recovered, err := RecoverPublication(ledger, paths); err != nil || recovered {
					t.Fatalf("repeated recovery: %v %v", recovered, err)
				}
			}
			want := next
			if outcome == "recover" {
				want = prior
			}
			for name, body := range want {
				assertPublicationBytes(t, filepath.Join(moved, name), body)
			}
			if err := CheckPublication(filepath.Join(moved, "agentic_metrics.json")); err != nil {
				t.Fatalf("final moved set pending: %v", err)
			}
			for _, dir := range []string{home, moved} {
				if _, err := os.Stat(filepath.Join(dir, "agentic_metrics.json.publication-pending")); !os.IsNotExist(err) {
					t.Fatalf("pending marker remains at %s: %v", dir, err)
				}
			}
			entries, err := os.ReadDir(home)
			if err != nil || len(entries) != 1 || entries[0].Name() != "unrelated" {
				t.Fatalf("old home was modified: %v %v", entries, err)
			}
			assertPublicationBytes(t, filepath.Join(home, "unrelated"), sentinel)
		})
	}
}

func assertPublicationBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("artifact %s: got %q, want %q: %v", path, got, want, err)
	}
}
