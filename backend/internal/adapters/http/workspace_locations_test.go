package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
)

func TestLedgerLocationsRemainDistinctAcrossMoveAndUnavailableRecoveryState(t *testing.T) {
	project := t.TempDir()
	isolatedHome := t.TempDir()
	t.Setenv("HOME", isolatedHome)
	cfg := httpFixturePolicy()
	cfg.ConfigPath = filepath.Join(project, ".agenticlens-config.yaml")
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{
		{Path: "current", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 8}},
		{Path: "history", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 8}},
		{Path: "missing-root", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 8}},
	}
	writeHTTPFixturePolicy(t, project, cfg)
	locations := []string{filepath.Join(project, "current", "WRONG-999"), filepath.Join(project, "history", "arbitrary-description")}
	payload := []byte(sampleTestLedgerJSON("TASK-7", "Recorded identity", 10, .1, 1))
	for _, location := range locations {
		if err := os.MkdirAll(location, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(location, "agentic_metrics.json"), payload, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(location, "CONTRACT.md"), []byte(location), 0600); err != nil {
			t.Fatal(err)
		}
	}
	planningOnly := filepath.Join(project, "current", "TASK-8-planning-only")
	if err := os.Mkdir(planningOnly, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planningOnly, "proposal.md"), []byte("No ledger exists."), 0600); err != nil {
		t.Fatal(err)
	}
	server := adaptershttp.NewServerWithConfig(project, fstest.MapFS{}, cfg)
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		return response
	}
	list := get("/api/workspaces")
	if list.Code != http.StatusOK {
		t.Fatalf("partial-root list: %d %s", list.Code, list.Body.String())
	}
	var ledgers []domain.AgenticMetricsLedger
	if err := json.Unmarshal(list.Body.Bytes(), &ledgers); err != nil {
		t.Fatal(err)
	}
	if len(ledgers) != 2 {
		t.Fatalf("duplicate locations collapsed or planning-only row fabricated: %s", list.Body.String())
	}
	seen := map[string]bool{}
	for _, ledger := range ledgers {
		if ledger.TicketID == nil || *ledger.TicketID != "TASK-7" {
			t.Fatalf("folder spelling became identity: %+v", ledger)
		}
		seen[ledger.WorkspacePath] = true
	}
	for _, location := range locations {
		if !seen[location] {
			t.Fatalf("location lost from list: %v", seen)
		}
	}
	if !strings.Contains(list.Header().Get("X-AgenticLens-Workspace-Diagnostics"), "missing-root") {
		t.Fatal("unavailable sibling root omitted from diagnostics")
	}
	if response := get("/api/workspaces/TASK-7"); response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "ambiguous") {
		t.Fatalf("ambiguous identity chose a root: %d %s", response.Code, response.Body.String())
	}
	selected := func(location, suffix string) *httptest.ResponseRecorder {
		return get("/api/workspaces/TASK-7" + suffix + "?workspace=" + url.QueryEscape(location))
	}
	for _, location := range locations {
		response := selected(location, "")
		if response.Code != http.StatusOK {
			t.Fatalf("explicit detail: %d %s", response.Code, response.Body.String())
		}
		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(response.Body.Bytes(), &detail); err != nil {
			t.Fatal(err)
		}
		if detail.WorkspacePath != location || detail.Ledger.WorkspacePath != location || fixtureArtifactContent(detail.Artifacts, "CONTRACT.md") == nil || *fixtureArtifactContent(detail.Artifacts, "CONTRACT.md") != location {
			t.Fatalf("detail/artifacts selected different home: %+v", detail)
		}
		response = selected(location, "/events")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"availability":"unavailable"`) {
			t.Fatalf("selected events fabricated source availability: %d %s", response.Code, response.Body.String())
		}
	}
	moved := filepath.Join(project, "history", "2026-renamed-with-no-ticket")
	if err := os.Rename(locations[0], moved); err != nil {
		t.Fatal(err)
	}
	if response := selected(moved, ""); response.Code != http.StatusOK {
		t.Fatalf("moved ledger inaccessible: %d %s", response.Code, response.Body.String())
	}
	if response := selected(locations[0], ""); response.Code == http.StatusOK {
		t.Fatal("stale location silently selected another duplicate")
	}
	if _, err := os.Stat(locations[0]); !os.IsNotExist(err) {
		t.Fatalf("read recreated old location: %v", err)
	}
	indicator := filepath.Join(moved, "agentic_metrics.json.publication-pending")
	pending := []byte(`{"version":2,"transaction":"0123456789abcdef0123456789abcdef"}`)
	if err := os.WriteFile(indicator, pending, 0600); err != nil {
		t.Fatal(err)
	}
	response := selected(moved, "")
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "publication") {
		t.Fatalf("copied pending set accepted without originating private journal: %d %s", response.Code, response.Body.String())
	}
	if after, err := os.ReadFile(indicator); err != nil || string(after) != string(pending) {
		t.Fatalf("read removed/repaired pending transaction: %v", err)
	}
	if after, err := os.ReadFile(filepath.Join(moved, "agentic_metrics.json")); err != nil || string(after) != string(payload) {
		t.Fatalf("read changed recorded facts: %v", err)
	}
	list = get("/api/workspaces")
	ledgers = nil
	if err := json.Unmarshal(list.Body.Bytes(), &ledgers); err != nil {
		t.Fatal(err)
	}
	if list.Code != http.StatusOK || len(ledgers) != 1 || ledgers[0].WorkspacePath != locations[1] || !strings.Contains(list.Header().Get("X-AgenticLens-Workspace-Diagnostics"), "publication") {
		t.Fatalf("pending sibling hid accepted home or became accepted facts: %d %s", list.Code, list.Body.String())
	}
	if _, err := os.Stat(filepath.Join(isolatedHome, ".agenticlens")); !os.IsNotExist(err) {
		t.Fatalf("portable read initialized operational state: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
		t.Fatalf("portable read wrote private configuration: %v", err)
	}
}

func TestHTTPDiscoveryDepthDefinesScopeAndBudgetReportsPartialResults(t *testing.T) {
	project := t.TempDir()
	selected := filepath.Join(project, "selected")
	child := filepath.Join(selected, "child-with-any-name")
	if err := os.MkdirAll(child, 0700); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ path, id string }{{selected, "ROOT-1"}, {child, "CHILD-2"}} {
		if err := os.WriteFile(filepath.Join(entry.path, "agentic_metrics.json"), []byte(sampleTestLedgerJSON(entry.id, entry.id, 1, 0, 1)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name                 string
		depth, budget, count int
		partial              bool
	}{{"root only", 0, 1, 1, false}, {"root and child", 1, 2, 2, false}, {"budget interruption", 1, 1, 1, true}} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := httpFixturePolicy()
			cfg.ConfigPath = filepath.Join(project, ".agenticlens-config.yaml")
			cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: "selected", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: tc.depth, MaxDirectories: tc.budget}}}
			writeHTTPFixturePolicy(t, project, cfg)
			response := httptest.NewRecorder()
			adaptershttp.NewServerWithConfig(project, fstest.MapFS{}, cfg).Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/workspaces", nil))
			var ledgers []domain.AgenticMetricsLedger
			if err := json.Unmarshal(response.Body.Bytes(), &ledgers); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusOK || len(ledgers) != tc.count {
				t.Fatalf("scope/budget results: %d %s", response.Code, response.Body.String())
			}
			diagnostic := response.Header().Get("X-AgenticLens-Workspace-Diagnostics")
			if strings.Contains(diagnostic, "incomplete") != tc.partial {
				t.Fatalf("depth completion confused with budget interruption: %s", diagnostic)
			}
		})
	}
}
