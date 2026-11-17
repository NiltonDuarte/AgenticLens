package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceArtifactSelectionRejectsReplacement(t *testing.T) {
	for _, mode := range []string{"stable", "replacement", "pending replacement"} {
		t.Run(mode, func(t *testing.T) {
			replace := mode != "stable"
			project := t.TempDir()
			writePublicationTestPolicy(t, project)
			a, b, moved := filepath.Join(project, "selected"), filepath.Join(project, "replacement"), filepath.Join(project, "moved")
			for _, v := range []struct{ path, title string }{{a, "Copy A"}, {b, "Copy B"}} {
				if err := os.Mkdir(v.path, 0700); err != nil {
					t.Fatal(err)
				}
				body := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":"Recorded-A","ticket_title":%q,"sessions":[]}`, v.title))
				if err := os.WriteFile(filepath.Join(v.path, "agentic_metrics.json"), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			selected, err := LoadWorkspaceArtifactSelection("Recorded-A", project, a)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := selected.Close(); err != nil {
					t.Error(err)
				}
			})
			if *selected.Ledger.TicketTitle != "Copy A" {
				t.Fatal("wrong initial selection")
			}
			if replace {
				if err := os.Rename(a, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(b, a); err != nil {
					t.Fatal(err)
				}
			}
			output := filepath.Join(a, "export.json")
			ledgerPath := filepath.Join(a, "agentic_metrics.json")
			before, err := os.ReadFile(ledgerPath)
			if err != nil {
				t.Fatal(err)
			}
			marker := ledgerPath + ".publication-pending"
			var pending []byte
			if mode == "pending replacement" {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				owner, err := capture.AcquirePublication(ctx, ledgerPath, project)
				if err != nil {
					t.Fatal(err)
				}
				if err := BeginPublication(ledgerPath, []string{ledgerPath, output}); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(output, []byte("interrupted replacement export"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := owner.Close(); err != nil {
					t.Fatal(err)
				}
				pending, err = os.ReadFile(marker)
				if err != nil {
					t.Fatal(err)
				}
			}
			err = PublishWorkspaceArtifacts(project, selected, output, "")
			if !replace {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("stale physical selection accepted")
			}
			after, readErr := os.ReadFile(ledgerPath)
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatalf("replacement source changed: %v", readErr)
			}
			if mode == "pending replacement" {
				got, err := os.ReadFile(marker)
				if err != nil || !bytes.Equal(got, pending) {
					t.Fatalf("replacement pending marker changed: %v", err)
				}
				got, err = os.ReadFile(output)
				if err != nil || string(got) != "interrupted replacement export" {
					t.Fatalf("stale selection recovered replacement: %s %v", got, err)
				}
				if err := RecoverWorkspaceArtifacts(project, a, output, ""); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("replacement export mutated: %v", err)
			}
			fresh, err := LoadWorkspaceArtifactSelection("Recorded-A", project, a)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := fresh.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := PublishWorkspaceArtifacts(project, fresh, output, ""); err != nil {
				t.Fatal(err)
			}
			if err := fresh.Close(); err != nil {
				t.Fatal(err)
			}
			if err := PublishWorkspaceArtifacts(project, fresh, output, ""); err == nil {
				t.Fatal("closed selection admitted")
			}
			if err := PublishWorkspaceArtifacts(project, &WorkspaceArtifactSelection{}, output, ""); err == nil {
				t.Fatal("uninitialized selection admitted")
			}

		})
	}
}
