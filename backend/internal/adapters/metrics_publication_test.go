package adapters

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

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
	if err := BeginPublication(path, []string{path}); err != nil {
		t.Fatal(err)
	}
	if err := SaveMetricsLedger(alias, domain.AgenticMetricsLedger{}); err == nil || !strings.Contains(err.Error(), "interrupted") {
		t.Fatalf("export accepted pending journal: %v", err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(body) {
		t.Fatal("export changed interrupted publication")
	}
}
