package adapters_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestWorkspaceDiscoveryDeduplicatesPhysicalScopesAndReadsAbsoluteRoots(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	external := discoveryPhysicalDirectory(t, t.TempDir())
	home := t.TempDir()
	t.Setenv("HOME", home)
	container := filepath.Join(project, "container")
	location := filepath.Join(container, "FALSE-999")
	inside := writeDiscoveryLedger(t, location, "Recorded-7")
	outside := writeDiscoveryLedger(t, external, "External-8")
	// Many lexical aliases must consume one physical directory from the budget.
	for i := 0; i < 5; i++ {
		if err := os.Symlink(location, filepath.Join(container, fmt.Sprintf("alias-%d", i))); err != nil {
			t.Fatal(err)
		}
	}
	configuredAlias := filepath.Join(project, "selected-alias")
	if err := os.Symlink(location, configuredAlias); err != nil {
		t.Fatal(err)
	}
	outsideAlias := filepath.Join(container, "outside-alias")
	if err := os.Symlink(external, outsideAlias); err != nil {
		t.Fatal(err)
	}
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{
		{Path: "container", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 2}},
		{Path: "container/FALSE-999", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}},
		{Path: "selected-alias", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}},
		{Path: external, Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}},
		{Path: "unavailable-sibling", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}},
	}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	report := adapters.ScanWorkspaces(project)
	if report.FatalError != "" || len(report.Ledgers) != 2 {
		t.Fatalf("aliases duplicated a physical ledger or absolute root was hidden: %+v", report)
	}
	seen := map[string]string{}
	for _, ledger := range report.Ledgers {
		if ledger.TicketID == nil {
			t.Fatal("ledger lost recorded identity")
		}
		seen[*ledger.TicketID] = ledger.WorkspacePath
	}
	if seen["Recorded-7"] != location || seen["External-8"] != external {
		t.Fatalf("configured paths or folder names changed recorded identity: %+v", seen)
	}
	outsideDiagnostic, missingDiagnostic := false, false
	for _, diagnostic := range report.Diagnostics {
		if strings.Contains(diagnostic.Message, "budget") {
			t.Fatalf("physical aliases exhausted the directory budget: %+v", diagnostic)
		}
		outsideDiagnostic = outsideDiagnostic || diagnostic.Path == outsideAlias && strings.Contains(diagnostic.Message, "outside")
		missingDiagnostic = missingDiagnostic || strings.Contains(diagnostic.Path, "unavailable-sibling")
	}
	if !outsideDiagnostic || !missingDiagnostic {
		t.Fatalf("scope/partial diagnostics lost configured context: %+v", report.Diagnostics)
	}
	resolved, err := adapters.ResolveTicketLocation(cfg, "Recorded-7", configuredAlias)
	if err != nil || resolved.Path != location {
		t.Fatalf("explicit physical alias did not select recorded home: %+v %v", resolved, err)
	}
	if _, err := adapters.ResolveTicketLocation(cfg, "FALSE-999", ""); err == nil {
		t.Fatal("folder name invented a ticket identity")
	}
	loaded, err := adapters.LoadTicketWorkspace("External-8", project, external)
	if err != nil || loaded == nil || loaded.WorkspacePath != external || len(loaded.Facts.Sessions) != 1 || len(loaded.Facts.Sessions[0].Tools) != 1 || loaded.Facts.Sessions[0].Tools[0].Calls != 2 {
		t.Fatalf("absolute portable read required private state or lost facts: %+v %v", loaded, err)
	}
	for path, want := range map[string][]byte{filepath.Join(location, "agentic_metrics.json"): inside, filepath.Join(external, "agentic_metrics.json"): outside} {
		if after, err := os.ReadFile(path); err != nil || !bytes.Equal(after, want) {
			t.Fatalf("read changed source ledger %s: %v", path, err)
		}
	}
	if entries, err := os.ReadDir(home); err != nil || len(entries) != 0 {
		t.Fatalf("read initialized private user state: %v %v", entries, err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
		t.Fatalf("read initialized local configuration: %v", err)
	}
}

func TestWorkspaceDiscoveryRejectsConflictingArtifactPolicyWithoutRootPrecedence(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	first := filepath.Join(project, "homes", "FIRST-900")
	second := filepath.Join(project, "homes", "unrelated-description")
	firstBytes := writeDiscoveryLedger(t, first, "Recorded-11")
	secondBytes := writeDiscoveryLedger(t, second, "Recorded-11")
	proposal := []domain.WorkspaceArtifactConfig{{Path: "proposal.md", Title: "Proposal"}}
	design := []domain.WorkspaceArtifactConfig{{Path: "design.md", Title: "Design"}}
	broad := domain.WorkspaceRootConfig{Path: "homes", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 3}, Artifacts: proposal}
	narrow := domain.WorkspaceRootConfig{Path: "homes/FIRST-900", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 1}, Artifacts: design}
	cfg := fixtureReadPolicy(t)
	for _, order := range [][]domain.WorkspaceRootConfig{{broad, narrow}, {narrow, broad}} {
		cfg.Workspaces.Roots = order
		cfg = writeDiscoveryPolicy(t, project, cfg)
		report := adapters.ScanWorkspaces(project)
		if report.FatalError != "" || len(report.Ledgers) != 1 || report.Ledgers[0].WorkspacePath != second {
			t.Fatalf("root precedence accepted contradictory artifact policy or hid sibling: %+v", report)
		}
		if len(report.Diagnostics) != 1 || report.Diagnostics[0].Path != filepath.Join(first, "agentic_metrics.json") || !strings.Contains(report.Diagnostics[0].Message, "artifact policy") {
			t.Fatalf("conflicting policy lost actual location: %+v", report.Diagnostics)
		}
		if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-11", first); err == nil {
			t.Fatal("explicit location bypassed conflicting artifact policy")
		}
	}
	// Once the conflicting policy is corrected, both valid locations require selection.
	narrow.Artifacts = proposal
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{broad, narrow}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	_, err := adapters.ResolveTicketLocation(cfg, "Recorded-11", "")
	var conflict *adapters.WorkspaceResolutionError
	if !errors.As(err, &conflict) || conflict.Outcome != "ambiguous" || len(conflict.Candidates) != 2 {
		t.Fatalf("distinct ledgers were merged or selected by root order: %v", err)
	}
	for _, location := range []string{first, second} {
		resolved, err := adapters.ResolveTicketLocation(cfg, "Recorded-11", location)
		if err != nil || resolved.Path != location || len(resolved.Artifacts) != 1 || resolved.Artifacts[0].Path != "proposal.md" {
			t.Fatalf("explicit location/policy selection changed: %+v %v", resolved, err)
		}
	}
	for path, want := range map[string][]byte{filepath.Join(first, "agentic_metrics.json"): firstBytes, filepath.Join(second, "agentic_metrics.json"): secondBytes} {
		if after, err := os.ReadFile(path); err != nil || !bytes.Equal(after, want) {
			t.Fatalf("policy resolution changed ledger bytes: %v", err)
		}
	}
}

func discoveryPhysicalDirectory(t *testing.T, path string) string {
	t.Helper()
	physical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return physical
}

func TestWorkspaceIncompleteDiscoveryCannotProveUniqueOrAbsentIdentity(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	visible := filepath.Join(project, "records")
	hidden := filepath.Join(visible, "unvisited-copy")
	initial := filepath.Join(visible, "new-home")
	visibleBody := writeDiscoveryLedger(t, visible, "Recorded-23")
	hiddenBody := writeDiscoveryLedger(t, hidden, "Recorded-23")
	if err := os.Mkdir(initial, 0700); err != nil {
		t.Fatal(err)
	}
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "records", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 1}}}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	report := adapters.ScanWorkspaces(project)
	if report.FatalError != "" || len(report.Ledgers) != 1 || len(report.Diagnostics) == 0 {
		t.Fatalf("partial listing lost observed ledger/diagnostic: %+v", report)
	}
	for _, maintenance := range []bool{false, true} {
		resolve := adapters.ResolveTicketLocation
		if maintenance {
			resolve = adapters.ResolveMaintenanceLocation
		}
		_, err := resolve(cfg, "Recorded-23", "")
		var resolution *adapters.WorkspaceResolutionError
		if !errors.As(err, &resolution) || resolution.Outcome != "unavailable" || len(resolution.Diagnostics) == 0 || len(resolution.Candidates) != 1 {
			t.Fatalf("incomplete discovery claimed uniqueness (maintenance=%v): %v", maintenance, err)
		}
		for _, selected := range []string{visible, hidden} {
			location, err := resolve(cfg, "Recorded-23", selected)
			if err != nil || location.Path != selected {
				t.Fatalf("explicit existing home unavailable (maintenance=%v): %+v %v", maintenance, location, err)
			}
		}
	}
	for _, ticket := range []string{"Recorded-23", "New-24"} {
		if err := adapters.ValidateInitialLedgerDestination(cfg, ticket, filepath.Join(initial, "agentic_metrics.json")); err == nil {
			t.Fatalf("incomplete discovery accepted first publication for %s", ticket)
		}
	}
	if _, err := os.Stat(filepath.Join(initial, "agentic_metrics.json")); !os.IsNotExist(err) {
		t.Fatalf("validation created a ledger: %v", err)
	}
	for path, want := range map[string][]byte{filepath.Join(visible, "agentic_metrics.json"): visibleBody, filepath.Join(hidden, "agentic_metrics.json"): hiddenBody} {
		if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, want) {
			t.Fatalf("discovery changed accepted bytes at %s: %v", path, err)
		}
	}
	// Within the configured depth, a complete scan sees ambiguity; root-only
	// discovery ends normally at depth zero and can establish uniqueness.
	cfg.Workspaces.Roots[0].Discovery.MaxDirectories = 20
	cfg = writeDiscoveryPolicy(t, project, cfg)
	if _, err := adapters.ResolveTicketLocation(cfg, "Recorded-23", ""); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("complete duplicate scan: %v", err)
	}
	cfg.Workspaces.Roots[0].Discovery.MaxDepth = 0
	cfg.Workspaces.Roots[0].Discovery.MaxDirectories = 1
	cfg = writeDiscoveryPolicy(t, project, cfg)
	if location, err := adapters.ResolveTicketLocation(cfg, "Recorded-23", ""); err != nil || location.Path != visible {
		t.Fatalf("normal depth boundary called incomplete: %+v %v", location, err)
	}
}

func TestWorkspaceInspectedInvalidSiblingDoesNotConcealDuplicate(t *testing.T) {
	project := discoveryPhysicalDirectory(t, t.TempDir())
	home := filepath.Join(project, "records", "valid")
	writeDiscoveryLedger(t, home, "Recorded-25")
	invalid := filepath.Join(project, "records", "invalid")
	if err := os.MkdirAll(invalid, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(invalid, "agentic_metrics.json"), []byte(`{"schema_version":999}`), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := fixtureReadPolicy(t)
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "records", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 3}}}
	cfg = writeDiscoveryPolicy(t, project, cfg)
	location, err := adapters.ResolveTicketLocation(cfg, "Recorded-25", "")
	if err != nil || location.Path != home || len(location.Diagnostics) == 0 {
		t.Fatalf("inspected invalid sibling hid a unique valid ledger: %+v %v", location, err)
	}
}
func writeDiscoveryLedger(t *testing.T, directory, ticket string) []byte {
	t.Helper()
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	body := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":%q,"sessions":[{"id":"recorded-unavailable-owner","source":"codex","user":"other-machine","invocations":[],"tools":[{"name":"read","calls":2,"errors":0}],"anomalies":[]}]}`, ticket))
	if err := os.WriteFile(filepath.Join(directory, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	return body
}
func writeDiscoveryPolicy(t *testing.T, project string, cfg domain.ProjectConfig) domain.ProjectConfig {
	t.Helper()
	body, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := adaptersconfig.LoadShared(project)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}
