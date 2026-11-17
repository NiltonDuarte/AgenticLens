package adapters_test

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestLoadTicketWorkspaceRejectsReplacementIdentity(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	first, second := filepath.Join(project, "selected"), filepath.Join(project, "replacement")
	writeDiscoveryLedger(t, first, "Recorded-A")
	writeDiscoveryLedger(t, second, "Recorded-B")
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "selected", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}}}
	writeDiscoveryPolicy(t, project, cfg)
	check := func() {
		t.Helper()
		ledger, err := adapters.LoadTicketWorkspace("Recorded-A", project)
		if err != nil || ledger == nil || ledger.TicketID == nil || *ledger.TicketID != "Recorded-A" {
			t.Fatalf("stable identity control failed: %+v %v", ledger, err)
		}
	}
	check()
	var stop atomic.Bool
	done := make(chan error, 1)
	go func() {
		for !stop.Load() {
			if err := unix.Renameat2(unix.AT_FDCWD, first, unix.AT_FDCWD, second, unix.RENAME_EXCHANGE); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	joined := false
	t.Cleanup(func() {
		if !joined {
			stop.Store(true)
			if err := <-done; err != nil {
				t.Error(err)
			}
		}
	})
	successful := 0
	for i := 0; i < 1500; i++ {
		ledger, err := adapters.LoadTicketWorkspace("Recorded-A", project)
		if err != nil {
			continue
		}
		successful++
		if ledger == nil || ledger.TicketID == nil || *ledger.TicketID != "Recorded-A" {
			t.Fatalf("replacement returned another ticket at read %d: %+v", i, ledger)
		}
	}
	stop.Store(true)
	swapErr := <-done
	joined = true
	if swapErr != nil {
		t.Fatal(swapErr)
	}
	// Both complete success and explicit errors are valid during replacement.
	// Restore the requested identity and require success after the real race.
	if _, err := adapters.LoadTicketWorkspace("Recorded-A", project); err != nil {
		if err := unix.Renameat2(unix.AT_FDCWD, first, unix.AT_FDCWD, second, unix.RENAME_EXCHANGE); err != nil {
			t.Fatal(err)
		}
	}
	check()
	t.Logf("identity remained correct across 1500 replacement reads (%d successful reads)", successful)
}

// A mutable selection may initially resolve either copy. Successful detail reads
// must still pair the ledger and artifact from one physical directory.
func TestLoadTicketWorkspaceDetailKeepsSameIDCopyCoherent(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	first, second := filepath.Join(project, "selected"), filepath.Join(project, "replacement")
	for _, copy := range []struct{ path, title string }{{first, "Copy A"}, {second, "Copy B"}} {
		writeDiscoveryLedger(t, copy.path, "Recorded-A")
		body := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":"Recorded-A","ticket_title":%q,"sessions":[]}`, copy.title))
		for name, contents := range map[string][]byte{"agentic_metrics.json": body, "evidence.md": []byte(copy.title)} {
			if err := os.WriteFile(filepath.Join(copy.path, name), contents, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: ".", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 3}}}
	cfg.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{{Path: "evidence.md", Title: "Evidence"}}
	writeDiscoveryPolicy(t, project, cfg)
	check := func() bool {
		t.Helper()
		ledger, artifacts, err := adapters.LoadTicketWorkspaceDetail("Recorded-A", project, first)
		if err != nil {
			return false
		}
		if ledger == nil || ledger.TicketID == nil || *ledger.TicketID != "Recorded-A" || ledger.TicketTitle == nil || (*ledger.TicketTitle != "Copy A" && *ledger.TicketTitle != "Copy B") || len(artifacts.Items) != 1 || artifacts.Items[0].Content == nil || *artifacts.Items[0].Content != *ledger.TicketTitle {
			t.Fatalf("same-ID detail mixed physical copies: ledger=%+v artifacts=%+v", ledger, artifacts)
		}
		return true
	}
	if !check() {
		t.Fatal("stable detail control failed before replacement")
	}
	var stop atomic.Bool
	done := make(chan error, 1)
	go func() {
		for !stop.Load() {
			if err := unix.Renameat2(unix.AT_FDCWD, first, unix.AT_FDCWD, second, unix.RENAME_EXCHANGE); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	joined := false
	t.Cleanup(func() {
		if !joined {
			stop.Store(true)
			if err := <-done; err != nil {
				t.Error(err)
			}
		}
	})
	successful := 0
	for i := 0; i < 400; i++ {
		if check() {
			successful++
		}
	}
	stop.Store(true)
	swapErr := <-done
	joined = true
	if swapErr != nil {
		t.Fatal(swapErr)
	}
	if !check() {
		t.Fatal("stable detail control failed after replacement")
	}
	t.Logf("same-ID ledger/artifact pairs stayed coherent across 400 reads (%d successful reads)", successful)
}
