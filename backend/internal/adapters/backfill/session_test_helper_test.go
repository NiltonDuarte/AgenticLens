package backfill

import (
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func sessionFacts(invocations []domain.InvocationFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	for i := range invocations {
		if invocations[i].Role == "" {
			invocations[i].Role = "orchestrator"
		}
		if invocations[i].Model == "" {
			invocations[i].Model = "fixture-model"
		}
		if invocations[i].StartedAt == nil {
			value := "fixture-started"
			invocations[i].StartedAt = &value
		}
		if invocations[i].CompletedAt == nil {
			value := "fixture-completed"
			invocations[i].CompletedAt = &value
		}
	}
	for i := range tools {
		if tools[i].SessionID == "" {
			tools[i].SessionID = "fixture-session"
		}
	}
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
	}
	sessions, err := domain.SessionsFromFacts(fixtureInvocationSessions(invocations), tools, events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].User == "" {
			sessions[i].User = "fixture-user"
		}
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	return sessions
}

type testInvFact struct {
	SessionID             string
	ParentID              *string
	Role                  domain.AgentRole
	Model                 string
	User                  string
	StartedAt             *string
	CompletedAt           *string
	ActiveDurationSeconds int64
	Turns                 int64
	Tokens                domain.InvocationTokens
	CostUSD               float64
	Compactions           int64
	UserIterations        int64
}

func sessionFactsWithInvs(invs []testInvFact, tools []domain.ToolFact, events []domain.EventFact) []domain.SessionFact {
	var sessions []domain.SessionFact
	for i, inv := range invs {
		sessID := inv.SessionID
		if sessID == "" {
			sessID = fmt.Sprintf("fixture-session-%d", i)
		}
		sessUser := inv.User
		if sessUser == "" {
			sessUser = "fixture-user"
		}
		model := inv.Model
		if model == "" {
			model = "fixture-model"
		}
		startedAt := inv.StartedAt
		if startedAt == nil {
			v := "fixture-started"
			startedAt = &v
		}
		completedAt := inv.CompletedAt
		if completedAt == nil {
			v := "fixture-completed"
			completedAt = &v
		}
		sess := domain.SessionFact{
			ID:       sessID,
			ParentID: inv.ParentID,
			Source:   domain.SourceOpenCode,
			User:     sessUser,
			Invocations: []domain.InvocationFact{
				{
					Role:                  inv.Role,
					Model:                 model,
					StartedAt:             startedAt,
					CompletedAt:           completedAt,
					ActiveDurationSeconds: inv.ActiveDurationSeconds,
					Turns:                 inv.Turns,
					Tokens:                inv.Tokens,
					CostUSD:               inv.CostUSD,
					Compactions:           inv.Compactions,
					UserIterations:        inv.UserIterations,
				},
			},
			Tools:     []domain.SessionToolFact{},
			Anomalies: []domain.Anomaly{},
		}
		sessions = append(sessions, sess)
	}
	for _, tool := range tools {
		for i := range sessions {
			if sessions[i].ID == tool.SessionID {
				sessions[i].Tools = append(sessions[i].Tools, domain.SessionToolFact{
					Name:   tool.Name,
					Calls:  tool.Calls,
					Errors: tool.Errors,
				})
			}
		}
	}
	return sessions
}

func setEventFacts(ledger *domain.AgenticFactLedger, events []domain.EventFact) {
	for i := range events {
		if events[i].SessionID == "" {
			events[i].SessionID = "fixture-session"
		}
		if events[i].Timestamp == "" {
			events[i].Timestamp = fmt.Sprintf("fixture-%06d", i)
		}
	}
	sessions, err := domain.SessionsFromFacts(ledger.Sessions, ledger.ToolFacts(), events)
	if err != nil {
		panic(err)
	}
	for i := range sessions {
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
	}
	ledger.Sessions = sessions
}

func fixtureInvocationSessions(invocations []domain.InvocationFact) []domain.SessionFact {
	if len(invocations) == 0 {
		return nil
	}
	return []domain.SessionFact{{ID: "fixture-session", Source: domain.SourceOpenCode, User: "fixture-user", Invocations: invocations}}
}

// A complete target project and explicit local budgets make source tests independent
// of the calling checkout and machine. These values are fixture choices only.
func writeBackfillProject(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, ".agenticlens-config.yaml")
	if _, err := os.Stat(path); err == nil {
		return
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ProjectName = "Backfill fixture"
	cfg.Workspaces.Dir = "."
	cfg.CodeMetrics.BaseBranch = "main"
	cfg.Generation.Metrics = domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
	cfg.Generation.Summary = domain.GenerationSummaryConfig{Enabled: true, Filename: "SUMMARY.md"}
	data, err := adaptersconfig.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	local := []byte("telemetry:\n  sources:\n    codex:\n      home: .\n  capture:\n    enabled: false\n    state_dir: fixture-state\n    max_sessions: 256\n    max_source_bytes: 67108864\n")
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), local, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
}

func writeBackfillRollout(t *testing.T, dir, id, parent, role, model string, tools map[string]int, suffix string) string {
	t.Helper()
	meta := struct {
		ID     string `json:"id"`
		Parent string `json:"parent_thread_id,omitempty"`
		Role   string `json:"agent_role,omitempty"`
	}{id, parent, role}
	encoded, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:00Z\",\"type\":\"session_meta\",\"payload\":%s}\n", encoded)
	body += fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:01Z\",\"type\":\"token_usage_record\",\"payload\":{\"model\":%q,\"usage\":{\"input_tokens\":1}}}\n", model)
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		for i := 0; i < tools[name]; i++ {
			call := fmt.Sprintf("%s-%s-%d", id, name, i)
			body += fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:02Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"function_call\",\"call_id\":%q,\"name\":%q,\"arguments\":\"{}\"}}\n", call, name)
			body += fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:03Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"function_call_output\",\"call_id\":%q,\"output\":\"success\"}}\n", call)
		}
	}
	path := filepath.Join(dir, id+".jsonl")
	if err := os.WriteFile(path, []byte(body+suffix), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func newBackfillProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeBackfillProject(t, root)
	return root
}

func fixtureCompleteAnomalyPolicy() domain.AnomalyThresholdsConfig {
	return domain.AnomalyThresholdsConfig{ErrorRateMinCalls: 5, ErrorRateThreshold: .35, ErrorRateCriticalThreshold: .70, ContextCeilingWarningTokens: 160000, ContextCeilingCriticalTokens: 200000, ConsecutiveErrorsStreakWarning: 3, ConsecutiveErrorsStreakCritical: 5, ToolThrashingMinFailures: 3, ToolThrashingCriticalFailures: 5, TurnUsageWarningTokens: 30000, TurnUsageCriticalTokens: 60000, ContextGrowthWarningTokens: 40000, ContextGrowthCriticalTokens: 80000}
}
