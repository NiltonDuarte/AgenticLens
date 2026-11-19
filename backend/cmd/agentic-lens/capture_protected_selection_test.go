package main_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
)

func TestProtectedPublicationRejectsReplacedSelection(t *testing.T) {
	f := newCodexInboxFixture(t)
	codexInboxOperation(t, f.project, "install-user")
	cmd := exec.Command(fixtureCaptureEngine(), "capture", "admit-protected-worktree", "--repo", f.project, "--protected-codex-home", f.store)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("admit disposable project: %s %v", out, err)
	}
	protection, err := capture.AdmitProtectedPublication(f.project, f.store)
	if err != nil {
		t.Fatal(err)
	}
	project, cfg, err := adapters.LoadWorkspaceProject(f.project)
	if err != nil {
		t.Fatal(err)
	}
	const ticket = "TEST-100"
	ledger := seedCaptureTicketHome(t, f, ticket)
	original := readCaptureFile(t, ledger)
	selected, err := adapters.SelectWorkspacePublication(cfg, ticket, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := selected.Close(); err != nil {
			t.Error(err)
		}
	})
	selectedDir := filepath.Dir(ledger)
	movedDir := filepath.Join(filepath.Dir(selectedDir), "moved")
	if err := os.Rename(selectedDir, movedDir); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(selectedDir, 0700); err != nil {
		t.Fatal(err)
	}
	replacement := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":%q,"ticket_title":"Replacement","sessions":[]}`, ticket))
	if err := os.WriteFile(ledger, replacement, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stale, err := selected.AcquirePublicationWithProtection(ctx, project, &protection)
	if stale != nil {
		if closeErr := stale.Close(); closeErr != nil {
			t.Error(closeErr)
		}
	}
	if stale != nil || err == nil || !strings.Contains(err.Error(), "workspace changed after selection; resolve and retry") {
		t.Fatalf("protected publication did not reject the stale selection with a retry diagnostic: %v", err)
	}
	if !bytes.Equal(readCaptureFile(t, ledger), replacement) || !bytes.Equal(readCaptureFile(t, filepath.Join(movedDir, filepath.Base(ledger))), original) {
		t.Fatal("stale protected admission changed a ledger")
	}

	fresh, err := adapters.SelectWorkspacePublication(cfg, ticket, ledger)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fresh.Close(); err != nil {
			t.Error(err)
		}
	})
	owner, err := fresh.AcquirePublicationWithProtection(ctx, project, &protection)
	if err != nil {
		t.Fatalf("admit a fresh selection under the same protected state: %v", err)
	}
	t.Cleanup(func() {
		if err := owner.Close(); err != nil {
			t.Error(err)
		}
	})
	if owner.StateHome() != protection.State {
		t.Fatalf("fresh selection used state %s, want %s", owner.StateHome(), protection.State)
	}
	if err := owner.CheckPrivatePath(filepath.Join(project, "journal.json")); err == nil {
		t.Fatal("fresh selection allowed a journal in the writable workspace")
	}
}
