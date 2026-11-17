package adapters

import (
	"agenticlens/internal/adapters/capture"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Real first-use creation also supplies a syscall-tracing entrypoint: any
// initialization failure must precede the public pending marker and members.
func TestPublicationPrivateStateInitialization(t *testing.T) {
	for _, scenario := range []string{"fresh-state", "new-project", "blocked-state", "blocked-project"} {
		t.Run(scenario, func(t *testing.T) {
			// Keep syscall fault-injection occurrence counts stable for this
			// diagnostic entrypoint; the test itself uses the real filesystem.
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()
			project := t.TempDir()
			writePublicationTestPolicy(t, project)
			state := filepath.Join(t.TempDir(), "new", "state")
			local := fmt.Sprintf("agenticlens:\n  state_dir: %q\n", state)
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
				t.Fatal(err)
			}
			ledger, summary := filepath.Join(project, "agentic_metrics.json"), filepath.Join(project, "delivery.md")
			accepted := map[string][]byte{ledger: []byte("prior accepted ledger\n"), summary: []byte("prior accepted human evidence\n")}
			for path, body := range accepted {
				if err := os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "blocked-state" {
				if err := os.WriteFile(filepath.Dir(state), []byte("state obstruction"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if scenario == "new-project" || scenario == "blocked-project" {
				base := filepath.Join(state, "publication", "journals")
				if err := os.MkdirAll(base, 0700); err != nil {
					t.Fatal(err)
				}
				if scenario == "blocked-project" {
					physical, err := filepath.EvalSymlinks(project)
					if err != nil {
						t.Fatal(err)
					}
					digest := sha256.Sum256([]byte(physical))
					if err := os.WriteFile(filepath.Join(base, hex.EncodeToString(digest[:])), []byte("namespace obstruction"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			}
			owner, err := capture.AcquirePublication(context.Background(), ledger, project)
			if err == nil {
				defer func() {
					if err := owner.Close(); err != nil {
						t.Error(err)
					}
				}()
				err = BeginPublication(ledger, []string{ledger, summary})
			}
			for path, body := range accepted {
				assertPublicationBytes(t, path, body)
			}
			blocked := scenario == "blocked-state" || scenario == "blocked-project"
			if err != nil {
				if _, pendingErr := os.Stat(journalPath(ledger)); !os.IsNotExist(pendingErr) {
					t.Fatalf("failed private initialization exposed pending marker: %v", pendingErr)
				}
				t.Logf("private initialization failed before pending marker and preserved exact accepted bytes: %v", err)
				if !blocked {
					t.Fatal(err)
				}
				return
			}
			if blocked {
				t.Fatal("obstructed private initialization was accepted")
			}
			if err := CheckPublication(ledger); err == nil {
				t.Fatal("begun publication has no pending marker")
			}
			if err := durableReplace(ledger, []byte("replacement ledger\n")); err != nil {
				t.Fatal(err)
			}
			if err := durableReplace(summary, []byte("replacement human evidence\n")); err != nil {
				t.Fatal(err)
			}
			if err := CommitPublication(ledger, []string{ledger, summary}); err != nil {
				t.Fatal(err)
			}
			assertPublicationBytes(t, ledger, []byte("replacement ledger\n"))
			assertPublicationBytes(t, summary, []byte("replacement human evidence\n"))
			if err := CheckPublication(ledger); err != nil {
				t.Fatal(err)
			}
		})
	}
}
