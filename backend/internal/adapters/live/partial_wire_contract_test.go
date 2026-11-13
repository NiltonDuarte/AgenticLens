package live_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"agenticlens/internal/adapters/config"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

// Recorded detail projects anomaly roles; partial source hydration retains the
// unavailable owner's raw facts. Both wires describe one recorded occurrence.
func TestNativePartialAnomalyWiresMatchFrontendFixtures(t *testing.T) {
	root := t.TempDir()
	cfg := testProjectConfig()
	cfg.AgenticLens.Version = "v9.8.7"
	cfg.Workspaces.Roots[0].Path = "."
	cfg.Ticket = domain.TicketConfig{Prefixes: []domain.TicketPrefix{{Prefix: "T-", Description: "Wire fixture"}}, CaseInsensitive: false}
	cfg.RoleMappings = domain.RoleMappings{"builder": "planner", "planner": "other", "source-tester": "tester", "tester": "helper"}
	cfg.Generation.Metrics = domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
	shared, err := config.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	ledger := []byte(`{"schema_version":6,"migration":{"schema_version":6,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},"ticket_id":"T-01","ticket_title":"Partial role wire","ticket_type":"feat","started_at":"2026-10-17T10:00:00Z","completed_at":"2026-10-17T11:00:00Z","updated_at":"2026-10-17T11:00:00Z","code":{"lines_added":0,"lines_deleted":0,"test_lines_added":0,"test_lines_deleted":0,"prod_lines_added":0,"prod_lines_deleted":0,"files_changed":0,"sdlc_iterations":0},"sessions":[{"id":"root-missing","source":"codex","user":"fixture","invocations":[{"role":"builder","model":"source-model","started_at":null,"completed_at":null,"active_duration_seconds":10,"turns":1,"tokens":{"input":17,"output":3,"cache_read":0,"reasoning":0,"peak_context":17},"cost_usd":0.1,"cost_evidence":{"state":"observed","known_subtotal_usd":0.1},"compactions":0,"user_iterations":1}],"tools":[{"name":"bash","calls":1,"errors":0}],"anomalies":[{"rule_id":"SUBAGENT_REWORK_LOOP","severity":"WARNING","context":{"role":"builder","secondary_role":"source-tester","count":3}}]},{"id":"root-present","source":"codex","user":"fixture","invocations":[{"role":"builder","model":"source-model","started_at":null,"completed_at":null,"active_duration_seconds":10,"turns":1,"tokens":{"input":17,"output":3,"cache_read":0,"reasoning":0,"peak_context":17},"cost_usd":0.1,"cost_evidence":{"state":"observed","known_subtotal_usd":0.1},"compactions":0,"user_iterations":1}],"tools":[{"name":"bash","calls":1,"errors":0}],"anomalies":[]}]}`)
	ticketDir := filepath.Join(root, "T-01")
	if err := os.Mkdir(ticketDir, 0700); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(ticketDir, "agentic_metrics.json")
	if err := os.WriteFile(ledgerPath, ledger, 0600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	rollout := filepath.Join(home, "sessions", "2026", "10", "26", "rollout-present.jsonl")
	if err := os.MkdirAll(filepath.Dir(rollout), 0700); err != nil {
		t.Fatal(err)
	}
	native := []byte("{\"type\":\"session_meta\",\"payload\":{\"id\":\"root-present\",\"session_id\":\"root-present\",\"timestamp\":\"2026-10-26T09:00:00Z\"}}\n")
	if err := os.WriteFile(rollout, native, 0600); err != nil {
		t.Fatal(err)
	}
	service := live.NewLiveServiceWithCodexHome(nil, cfg, home)
	defer service.Close()
	server := adaptershttp.NewServerWithConfig(root, fstest.MapFS{"index.html": {Data: []byte("dashboard")}}, cfg, service)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()
	get := func(route string) []byte {
		t.Helper()
		response, err := ts.Client().Get(ts.URL + route)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatalf("HTTP %s status=%d body=%s", route, response.StatusCode, body)
		}
		return body
	}
	detailWire := get("/api/workspaces/T-01")
	eventsWire := get("/api/workspaces/T-01/events")
	var detail adaptershttp.WorkspaceDetailResponse
	var events domain.WorkspaceEventsResponse
	if err := json.Unmarshal(detailWire, &detail); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(eventsWire, &events); err != nil {
		t.Fatal(err)
	}
	if events.Availability != domain.TracePartial || !events.IsHydrated || events.DataSource != domain.DataSourceLocalCodex || !reflect.DeepEqual(events.Coverage.RequiredSessionIDs, []string{"root-missing", "root-present"}) || !reflect.DeepEqual(events.Coverage.CompleteSessionIDs, []string{"root-present"}) || !reflect.DeepEqual(events.Coverage.Diagnostics, []string{"root-missing: Codex native session unavailable"}) {
		t.Fatalf("native owner coverage changed: %+v", events)
	}
	if len(events.Events) != 0 || len(events.Anomalies) != 1 || len(detail.Anomalies) != 1 || len(detail.Ledger.Anomalies) != 1 {
		t.Fatalf("expected one retained recorded anomaly, not invented events: detail=%+v events=%+v", detail.Anomalies, events)
	}
	assertContext := func(anomaly domain.Anomaly, owner, role, secondary string) {
		t.Helper()
		context := anomaly.Context
		if anomaly.RuleID != domain.RuleSubagentReworkLoop || context.Role == nil || string(*context.Role) != role || context.SecondaryRole == nil || string(*context.SecondaryRole) != secondary || context.Count == nil || *context.Count != 3 {
			t.Fatalf("recorded occurrence changed: %+v", anomaly)
		}
		if owner == "" {
			if context.SessionID != nil {
				t.Fatalf("raw owner-local context gained synthetic owner: %+v", context)
			}
		} else if context.SessionID == nil || *context.SessionID != owner {
			t.Fatalf("wrong owner context: %+v", context)
		}
	}
	assertContext(detail.Anomalies[0], "root-missing", "planner", "tester")
	assertContext(detail.Ledger.Anomalies[0], "root-missing", "planner", "tester")
	assertContext(events.Anomalies[0], "root-missing", "builder", "source-tester")
	facts := detail.Ledger.Facts.Sessions
	if len(facts) != 2 || facts[0].ID != "root-missing" || facts[1].ID != "root-present" || len(facts[0].Anomalies) != 1 || len(facts[0].Invocations) != 1 || facts[0].Invocations[0].Role != "builder" {
		t.Fatalf("raw canonical owners/persona changed: %+v", facts)
	}
	assertContext(facts[0].Anomalies[0], "", "builder", "source-tester")
	ownedRaw := facts[0].Anomalies[0]
	ownerID := facts[0].ID
	ownedRaw.Context.SessionID = &ownerID
	if !reflect.DeepEqual(ownedRaw, events.Anomalies[0]) {
		t.Fatalf("retained source occurrence differs from canonical raw owner fact: raw=%+v events=%+v", ownedRaw, events.Anomalies[0])
	}

	if after, err := os.ReadFile(ledgerPath); err != nil || !bytes.Equal(after, ledger) {
		t.Fatalf("partial hydration modified stored facts: %v", err)
	}
	// Fixed persisted/source values make these payloads deterministic. No fields,
	// owner ids, diagnostics or raw role values are normalized before comparison.
	for _, wire := range []struct {
		name string
		body []byte
	}{{"partial-detail-wire.json", detailWire}, {"partial-events-wire.json", eventsWire}} {
		var actual bytes.Buffer
		if err := json.Compact(&actual, wire.body); err != nil {
			t.Fatal(err)
		}
		fixture := filepath.Join("..", "..", "..", "..", "web", "src", "test", "fixtures", wire.name)
		expected, err := os.ReadFile(fixture)
		if err != nil {
			t.Errorf("shared fixture %s: %v; producer JSON=%s", wire.name, err, actual.Bytes())
			continue
		}
		var compactExpected bytes.Buffer
		if err := json.Compact(&compactExpected, expected); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(actual.Bytes(), compactExpected.Bytes()) {
			t.Errorf("shared fixture %s differs: producer JSON=%s expected=%s", wire.name, actual.Bytes(), compactExpected.Bytes())
		}
	}
}
