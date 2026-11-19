package adapters

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerationPublicationRejectsReplacementBeforeAdmission(t *testing.T) {
	for _, mode := range []string{"implicit", "explicit", "initial"} {
		t.Run(mode, func(t *testing.T) {
			project := t.TempDir()
			writePublicationTestPolicy(t, project)
			_, cfg, err := LoadWorkspaceProject(project)
			if err != nil {
				t.Fatal(err)
			}
			const ticket = "T-120"
			selectedDir, movedDir := filepath.Join(project, "selected"), filepath.Join(project, "moved")
			const ledgerName = "agentic_metrics.json"
			path := filepath.Join(selectedDir, ledgerName)
			if err := os.Mkdir(selectedDir, 0700); err != nil {
				t.Fatal(err)
			}
			original := []byte(`{"schema_version":6,"ticket_id":"T-120","ticket_title":"Selected","sessions":[]}`)
			replacement := []byte(`{"schema_version":6,"ticket_id":"T-120","ticket_title":"Replacement","sessions":[]}`)
			if mode != "initial" {
				if err := os.WriteFile(path, original, 0600); err != nil {
					t.Fatal(err)
				}
			}
			explicit := path
			if mode == "implicit" {
				explicit = ""
			}
			location, err := SelectWorkspacePublication(cfg, ticket, explicit)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := location.Close(); err != nil {
					t.Error(err)
				}
			})
			if err := os.Rename(selectedDir, movedDir); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(selectedDir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, replacement, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			owner, err := location.AcquirePublication(ctx, project)
			if owner != nil {
				if closeErr := owner.Close(); closeErr != nil {
					t.Error(closeErr)
				}
			}
			if owner != nil || err == nil || !strings.Contains(err.Error(), "retry") {
				t.Fatalf("write admission accepted a stale directory or concealed retry: %v", err)
			}
			if current, err := os.ReadFile(path); err != nil || !bytes.Equal(current, replacement) {
				t.Fatalf("replacement changed: %s, %v", current, err)
			}
			movedLedger := filepath.Join(movedDir, ledgerName)
			if mode == "initial" {
				if _, err := os.Stat(movedLedger); !os.IsNotExist(err) {
					t.Fatalf("initial selection unexpectedly published at the old directory: %v", err)
				}
				return
			}
			if current, err := os.ReadFile(movedLedger); err != nil || !bytes.Equal(current, original) {
				t.Fatalf("moved original changed: %s, %v", current, err)
			}
			// Fresh explicit selection is legitimate even with the same-ID
			// replacement still present in a sibling directory.
			fresh, err := SelectWorkspacePublication(cfg, ticket, movedLedger)
			if err != nil {
				t.Fatal(err)
			}
			if owner, err := fresh.AcquirePublication(ctx, project); err != nil {
				t.Error(err)
			} else if err := owner.Close(); err != nil {
				t.Error(err)
			}
			if err := fresh.Close(); err != nil {
				t.Error(err)
			}
		})
	}
}
