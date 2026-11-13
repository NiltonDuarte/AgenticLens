package adapters

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPublishWorkspaceNestedSummaryPreservesSourceFacts(t *testing.T) {
	project := t.TempDir()
	path := putFormatLedger(t, project, "T-1", `{"schema_version":6,"ticket_id":"T-1","sessions":[{"id":"unavailable-owner","source":"codex","user":"other-machine","invocations":[],"tools":[{"name":"read","calls":2,"errors":0}],"anomalies":[]}]}`)
	prior, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := LoadTicketWorkspace("T-1", project)
	if err != nil {
		t.Fatal(err)
	}
	summary := filepath.Join(filepath.Dir(path), "reports", "nested", "SUMMARY.md")
	if err := PublishWorkspaceArtifacts(project, *selected, "", summary); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(summary)
	if err != nil || !bytes.Contains(first, []byte("Workflow Efficiency")) {
		t.Fatalf("selected nested summary missing: %q %v", first, err)
	}
	if err := PublishWorkspaceArtifacts(project, *selected, "", summary); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(summary)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("nested summary changed on repeat: %q %v", second, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(prior, after) {
		t.Fatalf("summary publication changed source facts: %q %v", after, err)
	}
	if err := CheckPublication(path); err != nil {
		t.Fatalf("successful nested publication left pending state: %v", err)
	}
}

func TestSaveMetricsLedgerOwnsPhysicalOutput(t *testing.T) {
	root := t.TempDir()
	path := putFormatLedger(t, root, "T-1", formatFixture)
	alias := filepath.Join(root, "alias.json")
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	owner, err := capture.AcquirePublication(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	held := true
	defer func() {
		if held {
			if err := owner.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	deadline, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx := &waitingFormatContext{Context: deadline, waiting: make(chan struct{})}
	done := make(chan error, 1)
	id, title := "T-1", "new export snapshot"
	ledger := domain.AgenticMetricsLedger{TicketID: &id, TicketTitle: &title}
	go func() { done <- saveMetricsLedger(ctx, alias, ledger) }()
	select {
	case <-ctx.waiting:
	case err := <-done:
		t.Fatalf("export bypassed output ownership: %v", err)
	case <-deadline.Done():
		t.Fatal("export never attempted ownership")
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	held = false
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-deadline.Done():
		t.Fatal("export did not finish")
	}
	info, err := os.Lstat(alias)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("export replaced physical alias: %v %v", info, err)
	}
	body, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(body), title) {
		t.Fatalf("export did not publish target: %s %v", body, err)
	}
	if err := beginTestPublication(t, path, []string{path}); err != nil {
		t.Fatal(err)
	}
	if err := capture.PublicationOwner(path).Close(); err != nil {
		t.Fatal(err)
	}
	if err := SaveMetricsLedger(alias, domain.AgenticMetricsLedger{}); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("export accepted pending journal: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(body) {
		t.Fatal("export changed interrupted publication")
	}
}
