package adapters_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// Each case authors a complete project configuration and exercises the durable
// source-summary writer, disk reload, detail projection and fleet consumers.
func TestConfiguredFailureRateSurvivesPersistenceAndFleet(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		minimum, calls, errors int
		warning, critical      float64
		severity               domain.AnomalySeverity
		threshold              string
	}{
		{"below warning", 5, 20, 6, .35, .70, "", ""},
		{"exact warning", 5, 20, 7, .35, .70, domain.SeverityWarning, ">=35%"},
		{"below critical", 5, 20, 13, .35, .70, domain.SeverityWarning, ">=35%"},
		{"exact critical", 5, 20, 14, .35, .70, domain.SeverityCritical, ">=70%"},
		{"custom minimum", 21, 20, 20, .35, .70, "", ""},
		{"lower policy", 3, 10, 2, .10, .20, domain.SeverityCritical, ">=20%"},
		{"higher policy", 3, 10, 7, .60, .90, domain.SeverityWarning, ">=60%"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			workspaces := filepath.Join(root, "tickets")
			project := fixtureReadPolicy(t)
			project.Workspaces.Roots[0].Path = "tickets"
			project.Ticket = domain.TicketConfig{CaseInsensitive: false, Prefixes: []domain.TicketPrefix{{Prefix: "FX", Description: "Policyfixture"}}}
			project.AnomalyThresholds.ErrorRateMinCalls = tc.minimum
			project.AnomalyThresholds.ErrorRateThreshold = tc.warning
			project.AnomalyThresholds.ErrorRateCriticalThreshold = tc.critical
			cfg, err := adaptersconfig.GenerateConfigYAML(project)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), cfg, 0600); err != nil {
				t.Fatal(err)
			}
			policy := project.AnomalyThresholds
			tool, tid, title := "read", "FX-1", "Configured rate"
			events := []domain.EventFact{}
			parts := []domain.LivePartFact{}
			for i := 0; i < tc.calls; i++ {
				status := "completed"
				isError := i < tc.errors
				if isError {
					status = "error"
				}
				events = append(events, domain.EventFact{EventID: fmt.Sprintf("event-%d", i), SessionID: "local", Timestamp: fmt.Sprintf("2026-11-08T10:00:%02dZ", i), Turn: int64(i + 1), EventType: domain.EventTypeToolCall, ToolName: &tool, Status: status, Role: "coder"})
				parts = append(parts, domain.LivePartFact{ID: fmt.Sprintf("part-%d", i), Type: "tool", ToolName: &tool, Status: status, IsError: isError})
			}
			local := sessionFactsWithInvs([]testInvFact{{SessionID: "local", Role: "coder", Model: "model"}}, []domain.ToolFact{{SessionID: "local", Name: tool, Calls: int64(tc.calls), Errors: int64(tc.errors)}}, nil)[0]
			local.Anomalies = []domain.Anomaly{}
			remote := sessionFactsWithInvs([]testInvFact{{SessionID: "unavailable-other-machine", Role: "tester", Model: "remote-model"}}, nil, nil)[0]
			remote.Source = domain.SourceCodex
			remote.User = "other-developer"
			remote.Anomalies = nil
			version := domain.CurrentSchemaVersion
			facts := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &tid, TicketTitle: &title, Sessions: []domain.SessionFact{local}}
			persisted, err := domain.PersistSourceAnomalySummaries(facts, domain.SourceEvents{Events: events, CompleteSessionIDs: []string{"local"}}, policy)
			if err != nil {
				t.Fatal(err)
			}
			persisted.Sessions = append(persisted.Sessions, remote)
			persisted, err = domain.PersistSourceAnomalySummaries(persisted, domain.SourceEvents{Events: events, CompleteSessionIDs: []string{"local"}}, policy)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(persisted.Sessions[1], remote) {
				t.Fatalf("unavailable mixed-harness owner changed: %+v", persisted.Sessions[1])
			}
			rates := func(anomalies []domain.Anomaly) []domain.Anomaly {
				var out []domain.Anomaly
				for _, a := range anomalies {
					if a.RuleID == domain.RuleToolHighFailureRate {
						out = append(out, a)
					}
				}
				return out
			}
			expected := 0
			if tc.severity != "" {
				expected = 1
			}
			assertRate := func(label string, anomalies []domain.Anomaly) {
				t.Helper()
				got := rates(anomalies)
				if len(got) != expected {
					t.Fatalf("%s rate summaries=%+v expected=%d", label, got, expected)
				}
				if expected == 1 && (got[0].Severity != tc.severity || got[0].Context.Threshold == nil || *got[0].Context.Threshold != tc.threshold) {
					t.Fatalf("%s incorrect policy/severity: %+v", label, got)
				}
			}
			assertRate("complete source persistence", persisted.AnomalyFacts())
			live := domain.EvaluateLiveToolFailureRates(&domain.LiveTraceView{SessionID: "local", Messages: []domain.LiveMessageFact{{Timestamp: "2026-11-08T10:00:00Z", Parts: parts}}}, policy)
			if len(live) != expected || (expected == 1 && (string(live[0].Severity) != string(tc.severity) || !strings.Contains(live[0].Message, tc.threshold))) {
				t.Fatalf("live/offline policy mismatch: %+v", live)
			}
			target := filepath.Join(workspaces, tid, "agentic_metrics.json")
			if err := saveFixtureFactLedger(t,target, persisted); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(target)
			if err != nil {
				t.Fatal(err)
			}
			detail, err := adapters.LoadTicketWorkspace(tid, workspaces)
			if err != nil {
				t.Fatal(err)
			}
			assertRate("detail after event pruning", detail.Anomalies)
			report := adapters.ScanWorkspaces(workspaces)
			if report.FatalError != "" || len(report.Diagnostics) != 0 || len(report.Ledgers) != 1 {
				t.Fatalf("saved ledger scan failed: %+v", report)
			}
			fleet := domain.CalculateFleetSummary(report.Ledgers)
			if fleet.AnomalySummary.RuleCounts[string(domain.RuleToolHighFailureRate)] != int64(expected) {
				t.Fatalf("fleet changed configured projection: %+v", fleet.AnomalySummary)
			}
			after, err := os.ReadFile(target)
			if err != nil || string(after) != string(body) {
				t.Fatalf("read projections rewrote persisted evidence: %v", err)
			}
		})
	}
}
