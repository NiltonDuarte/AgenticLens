package adapters_test

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspacePendingSiblingCannotProveIdentity(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	a, b, c := filepath.Join(project, "homes", "a"), filepath.Join(project, "homes", "b"), filepath.Join(project, "homes", "c")
	writeDiscoveryLedger(t, a, "Recorded-25")
	raw := writeDiscoveryLedger(t, b, "Recorded-25")
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "homes", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 10}}}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	marker := filepath.Join(b, "agentic_metrics.json.publication-pending")
	pending := []byte(`{"version":2,"transaction":"0123456789abcdef0123456789abcdef"}`)
	if err := os.WriteFile(marker, pending, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-25", ""); err == nil {
		t.Fatal("pending sibling established implicit uniqueness")
	}
	report := adapters.ScanWorkspaces(project)
	if report.FatalError != "" || len(report.Ledgers) != 1 || len(report.Diagnostics) == 0 {
		t.Fatalf("partial fleet lost healthy row/diagnostic: %+v", report)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-25", a); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(a, "agentic_metrics.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(c, 0700); err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateInitialLedgerDestination(cfg, "Recorded-25", filepath.Join(c, "agentic_metrics.json")); err == nil {
		t.Fatal("sole pending home permitted first generation elsewhere")
	}
	for path, want := range map[string][]byte{marker: pending, filepath.Join(b, "agentic_metrics.json"): raw} {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("pending bytes changed: %s %v", path, err)
		}
	}
	// Definitively malformed siblings are different from operationally unreadable ones.
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(b, "agentic_metrics.json"), []byte(`{broken`), 0600); err != nil {
		t.Fatal(err)
	}
	writeDiscoveryLedger(t, a, "Recorded-25")
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-25", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := adapters.ResolveMaintenanceLocation(cfg, "Recorded-25", ""); err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateInitialLedgerDestination(cfg, "Recorded-26", filepath.Join(c, "agentic_metrics.json")); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceMaintenanceInvalidAndHistoricalSelection(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	home := filepath.Join(project, "homes", "legacy")
	writeDiscoveryLedger(t, home, "Recorded-25")
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "homes", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 10}}}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	path := filepath.Join(home, "agentic_metrics.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"ticket_id":"Recorded-25","sessions":[]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := adapters.ResolveMaintenanceLocation(cfg, "Recorded-25", home); err != nil {
		t.Fatalf("historical explicit maintenance blocked: %v", err)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-25", home); err == nil {
		t.Fatal("runtime accepted historical schema")
	}
	if err := os.WriteFile(path, []byte(`{"ticket_id":42}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := adapters.ResolveMaintenanceLocation(cfg, "Recorded-25", home)
	var resolution *adapters.WorkspaceResolutionError
	if !errors.As(err, &resolution) || resolution.Outcome != "invalid_ledger" {
		t.Fatalf("invalid maintenance classified operationally unavailable: %v", err)
	}
}

func TestWorkspaceAliasUsesPhysicalDepth(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	deep := filepath.Join(project, "homes", "nested", "deep")
	writeDiscoveryLedger(t, deep, "Recorded-99")
	alias := filepath.Join(project, "homes", "alias")
	if err := os.Symlink(deep, alias); err != nil {
		t.Fatal(err)
	}
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "homes", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 10}}}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	report := adapters.ScanWorkspaces(project)
	if report.FatalError != "" || len(report.Ledgers) != 0 {
		t.Fatalf("alias advertised physically excluded target: %+v", report)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-99", alias); err == nil {
		t.Fatal("alias escaped physical depth")
	}
	cfg.Workspaces.Roots = append(cfg.Workspaces.Roots, domain.WorkspaceRootConfig{Path: deep, Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}})
	cfg = writeDiscoveryPolicy(t, project, cfg)
	report = adapters.ScanWorkspaces(project)
	if report.FatalError != "" || len(report.Ledgers) != 1 {
		t.Fatalf("independent coverage lost target: %+v", report)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-99", alias); err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateLedgerDestinationScope(cfg, filepath.Join(deep, "agentic_metrics.json")); err != nil {
		t.Fatal(err)
	}
}
