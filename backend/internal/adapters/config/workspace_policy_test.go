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
	body, err := config.InitialLocalYAMLWithSettings(project, "", "false", "codex", filepath.Join(project, "codex-source"), "", captureState, publicationState)
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
	repeated, err := config.InitialLocalYAMLWithSettings(project, "", "false", "codex", "", "", captureState, filepath.Join(project, "different-publication-state"))
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
