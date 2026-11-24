package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestWorkspaceRootsRequireExplicitScopeAndRejectRetiredShapes(t *testing.T) {
	base := fixtureSharedBody(t)
	valid := `[{path: arbitrary/location, discovery: {max_depth: 0, max_directories: 300}}]`
	cfg, err := loadBody(t, replaceSetting(t, base, "workspaces.roots", valid))
	if err != nil || len(cfg.Workspaces.Roots) != 1 || cfg.Workspaces.Roots[0].Discovery.MaxDepth != 0 {
		t.Fatalf("explicit root-only depth was lost: %+v %v", cfg.Workspaces.Roots, err)
	}
	for _, roots := range []string{
		`[{path: records}]`,
		`[{path: records, discovery: {max_directories: 300}}]`,
		`[{path: records, discovery: {max_depth: 0}}]`,
		`[{path: records, discovery: {max_depth: -1, max_directories: 300}}]`,
		`[{path: records, discovery: {max_depth: 0, max_directories: 0}}]`,
		`[{path: records, discovery: {max_depth: "0", max_directories: 300}}]`,
		`[{path: records, discovery: {max_depth: 0, max_directories: "300"}}]`,
		`[{path: records, id: old-root, discovery: {max_depth: 0, max_directories: 300}}]`,
		`[{path: records, discovery: {min_depth: 0, max_depth: 0, max_directories: 300}}]`,
		`[{path: records, discovery: {max_depth: 0, max_depth: 1, max_directories: 300}}]`,
		`[{path: "", discovery: {max_depth: 0, max_directories: 300}}]`,
	} {
		t.Run(roots, func(t *testing.T) {
			if _, err := loadBody(t, replaceSetting(t, base, "workspaces.roots", roots)); err == nil || (!strings.Contains(err.Error(), "workspaces") && !strings.Contains(err.Error(), "max_depth")) {
				t.Fatalf("invalid root policy accepted or unidentified: %v", err)
			}
		})
	}
	for _, policy := range []string{
		`{dir: records, artifacts: []}`,
		`{roots: ` + valid + `, dir: records, artifacts: []}`,
		`{roots: ` + valid + `, discovery: {max_depth: 0, max_directories: 300}, artifacts: []}`,
	} {
		if _, err := loadBody(t, replaceSetting(t, base, "workspaces", policy)); err == nil {
			t.Fatalf("retired configuration accepted: %s", policy)
		}
	}
}

func TestWorkspaceWriterPreservesPerRootScopeWithoutRetiredFields(t *testing.T) {
	cfg := testProjectConfig()
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{
		{Path: "direct-home", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 300}},
		{Path: "nested/selected", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 2, MaxDirectories: 17}, Artifacts: []domain.WorkspaceArtifactConfig{{Path: "delivery.md", Title: "Delivery"}}},
	}
	body, err := config.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := loadBody(t, string(body))
	if err != nil || !reflect.DeepEqual(loaded.Workspaces.Roots, cfg.Workspaces.Roots) {
		t.Fatalf("writer changed explicit root scope/artifacts: %+v %v", loaded.Workspaces.Roots, err)
	}
	for _, retired := range []string{"min_depth:", "id:", "\n  dir:"} {
		if strings.Contains(string(body), retired) {
			t.Fatalf("writer emitted retired setting %q: %s", retired, body)
		}
	}
	if !strings.Contains(string(body), "max_depth: 0") || !strings.Contains(string(body), "max_directories: 300") {
		t.Fatalf("writer omitted explicit root-only values: %s", body)
	}
}

func TestPublicationStateIsExplicitLocalAndReadOnlyResolutionDoesNotInitialize(t *testing.T) {
	project := t.TempDir()
	writeConfig(t, project, ".agenticlens-config.yaml", fixtureSharedBody(t))
	if _, err := config.ResolvePublicationState(project); err == nil {
		t.Fatal("missing operational state silently defaulted")
	}
	local := writeConfig(t, project, ".agenticlens-config.local.yaml", "agenticlens:\n  state_dir: private-publication\ntelemetry:\n  capture:\n    enabled: invalid-unused-capture-setting\n")
	state, err := config.ResolvePublicationState(project)
	if err != nil || state != filepath.Join(project, "private-publication") {
		t.Fatalf("operational state required capture or lost local origin: %q %v", state, err)
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("state resolution initialized a private directory: %v", err)
	}
	if _, err := config.LoadShared(project); err != nil {
		t.Fatalf("shared-only read interpreted private capture state: %v", err)
	}
	if err := os.WriteFile(local, []byte("agenticlens: [broken-unused-local-state]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadShared(project); err != nil {
		t.Fatalf("portable read depended on broken private state: %v", err)
	}
	if _, err := config.ResolvePublicationState(project); err == nil {
		t.Fatal("writer operational-state resolution accepted malformed local policy")
	}
	body := strings.Replace(fixtureSharedBody(t), "  version: v9.8.7", "  version: v9.8.7\n  state_dir: private", 1)
	if _, err := loadBody(t, body); err == nil {
		t.Fatal("private operational state accepted in shared policy")
	}
}

func TestLocalWriterEmitsPublicationStateSeparatelyAndPreservesSelection(t *testing.T) {
	project := t.TempDir()
	captureState := filepath.Join(project, "capture-registry")
	publicationState := filepath.Join(project, "publication-state")
	body, err := config.InitialLocalYAMLWithSettings(project, "", "false", "codex", filepath.Join(project, "codex-source"), "", "", captureState, publicationState)
	if err != nil {
		t.Fatal(err)
	}
	writeConfig(t, project, ".agenticlens-config.yaml", fixtureSharedBody(t))
	writeConfig(t, project, ".agenticlens-config.local.yaml", string(body))
	selected, err := config.ResolvePublicationState(project)
	if err != nil || selected != publicationState {
		t.Fatalf("writer omitted explicit operational state: %q %v", selected, err)
	}
	if !strings.Contains(string(body), captureState) {
		t.Fatalf("writer lost separate capture policy: %s", body)
	}
	repeated, err := config.InitialLocalYAMLWithSettings(project, "", "false", "codex", "", "", "", captureState, filepath.Join(project, "different-publication-state"))
	if err != nil {
		t.Fatal(err)
	}
	if string(repeated) != string(body) {
		t.Fatalf("repeated setup overwrote selected state:\n%s", repeated)
	}
	for _, path := range []string{captureState, publicationState} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("configuration writer initialized runtime storage %s: %v", path, err)
		}
	}
}

// The local writer accepts an explicit Claude Code source, round-trips through
// the real loader, and fills only missing choices. Paths are synthetic fixtures.
func TestLocalWriterRecordsClaudeSourceAndPreservesExistingValues(t *testing.T) {
	writeLocal := func(t *testing.T, project string, body []byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), body, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("fresh setup", func(t *testing.T) {
		project := t.TempDir()
		store := filepath.Join(project, "claude-synthetic-store")
		body, err := config.InitialLocalYAMLWithSettings(project, "", "false", "claude", "", store, "", filepath.Join(project, "capture"))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(body), "codex") || strings.Contains(string(body), "database_path") {
			t.Fatalf("Claude Code setup wrote another harness store:\n%s", body)
		}
		writeConfig(t, project, ".agenticlens-config.yaml", fixtureSharedBody(t))
		writeLocal(t, project, body)
		cfg, err := config.Load(project)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Telemetry.DefaultSource != "claude" || cfg.Telemetry.Sources.ClaudeHome != store {
			t.Fatalf("writer output did not round-trip: %+v", cfg.Telemetry)
		}
		repeated, err := config.InitialLocalYAMLWithSettings(project, "", "false", "claude", "", filepath.Join(project, "different-store"), "", filepath.Join(project, "capture"))
		if err != nil || string(repeated) != string(body) {
			t.Fatalf("repeated setup replaced the selected store: %v\n%s", err, repeated)
		}
	})
	t.Run("missing store", func(t *testing.T) {
		project := t.TempDir()
		_, err := config.InitialLocalYAMLWithSettings(project, "", "false", "claude", filepath.Join(project, "codex"), "", "", filepath.Join(project, "capture"))
		if err == nil || !strings.Contains(err.Error(), "telemetry.sources.claude.home") {
			t.Fatalf("a Claude Code source without its store must fail: %v", err)
		}
	})
	t.Run("existing selection preserved", func(t *testing.T) {
		project := t.TempDir()
		existing := []byte("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: /synthetic/codex\n  capture:\n    enabled: false\n    state_dir: /synthetic/capture\nagenticlens:\n  state_dir: /synthetic/publication\n")
		writeLocal(t, project, existing)
		body, err := config.InitialLocalYAMLWithSettings(project, "", "false", "claude", "", "/synthetic/claude", "", "/synthetic/capture")
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != string(existing) {
			t.Fatalf("valid existing source selection must be preserved:\n%s", body)
		}
	})
	t.Run("settings candidate", func(t *testing.T) {
		project := t.TempDir()
		candidate := filepath.Join(t.TempDir(), "local-settings.yaml")
		if err := os.WriteFile(candidate, []byte("telemetry:\n  default_source: claude\n  sources:\n    claude:\n      home: /synthetic/claude-from-settings\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		body, err := config.InitialLocalYAMLWithSettings(project, candidate, "false", "", "", "", "", "/synthetic/capture")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "default_source: claude") || !strings.Contains(string(body), "home: /synthetic/claude-from-settings") {
			t.Fatalf("approved settings candidate lost the Claude Code store:\n%s", body)
		}
	})
}
