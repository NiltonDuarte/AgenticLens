package adapters_test

import (
	"encoding/json"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestAnomalyContextDTO_BidirectionalMapping(t *testing.T) {
	t.Run("fully populated context maps to and from domain", func(t *testing.T) {
		tool := "bash"
		sess := "ses-10"
		turn := int64(4)
		role := "coder"
		secRole := "tester"
		count := int64(5)
		totalCalls := int64(10)
		delta := int64(45000)
		ctxTokens := int64(120000)
		thresh := ">=50%"
		failPct := 50.0

		domCtx := domain.AnomalyContext{
			ToolName:       &tool,
			SessionID:      &sess,
			Turn:           &turn,
			Role:           &role,
			SecondaryRole:  &secRole,
			Count:          &count,
			TotalCalls:     &totalCalls,
			TokenDelta:     &delta,
			ContextTokens:  &ctxTokens,
			Threshold:      &thresh,
			FailureRatePct: &failPct,
		}

		dto := adapters.AnomalyContextDTOFromDomain(domCtx)
		if *dto.ToolName != tool || *dto.SessionID != sess || *dto.Turn != turn ||
			*dto.Role != role || *dto.SecondaryRole != secRole || *dto.Count != count ||
			*dto.TotalCalls != totalCalls || *dto.TokenDelta != delta || *dto.ContextTokens != ctxTokens ||
			*dto.Threshold != thresh || *dto.FailureRatePct != failPct {
			t.Fatalf("AnomalyContextDTOFromDomain lost fields: %+v", dto)
		}

		roundDom := dto.ToDomain()
		if *roundDom.ToolName != tool || *roundDom.SessionID != sess || *roundDom.Turn != turn ||
			*roundDom.Role != role || *roundDom.SecondaryRole != secRole || *roundDom.Count != count ||
			*roundDom.TotalCalls != totalCalls || *roundDom.TokenDelta != delta || *roundDom.ContextTokens != ctxTokens ||
			*roundDom.Threshold != thresh || *roundDom.FailureRatePct != failPct {
			t.Fatalf("ToDomain lost fields: %+v", roundDom)
		}
	})

	t.Run("nil context fields map cleanly without panic", func(t *testing.T) {
		domCtx := domain.AnomalyContext{}
		dto := adapters.AnomalyContextDTOFromDomain(domCtx)
		if dto.ToolName != nil || dto.SessionID != nil || dto.Turn != nil ||
			dto.Role != nil || dto.SecondaryRole != nil || dto.Count != nil ||
			dto.TotalCalls != nil || dto.TokenDelta != nil || dto.ContextTokens != nil ||
			dto.Threshold != nil || dto.FailureRatePct != nil {
			t.Fatalf("expected all fields nil in DTO, got: %+v", dto)
		}

		roundDom := dto.ToDomain()
		if roundDom.ToolName != nil || roundDom.SessionID != nil || roundDom.Turn != nil {
			t.Fatalf("expected all fields nil in domain context, got: %+v", roundDom)
		}
	})
}

func TestAnomalyDTO_BidirectionalMapping(t *testing.T) {
	toolName := "git"
	turn := int64(3)
	count := int64(2)

	dom := domain.Anomaly{
		RuleID:      domain.RuleToolHighFailureRate,
		Severity:    domain.SeverityCritical,
		Title:       "High tool failure rate: git (50.0%)",
		Description: "Tool 'git' failed 3 out of 6 calls (50.0% failure rate, threshold: >=50%)",
		Context: domain.AnomalyContext{
			ToolName: &toolName,
			Turn:     &turn,
			Count:    &count,
		},
	}

	dto := adapters.AnomalyDTOFromDomain(dom)
	if dto.RuleID != string(domain.RuleToolHighFailureRate) {
		t.Errorf("expected RuleID %s, got %s", domain.RuleToolHighFailureRate, dto.RuleID)
	}
	if dto.Severity != string(domain.SeverityCritical) {
		t.Errorf("expected Severity %s, got %s", domain.SeverityCritical, dto.Severity)
	}
	if dto.Title != dom.Title || dto.Description != dom.Description {
		t.Errorf("title/description mismatch in DTO: %+v", dto)
	}
	if *dto.Context.ToolName != toolName {
		t.Errorf("context tool name mismatch: %v", dto.Context.ToolName)
	}

	roundDom := dto.ToDomain()
	if roundDom.RuleID != dom.RuleID || roundDom.Severity != dom.Severity ||
		roundDom.Title != dom.Title || roundDom.Description != dom.Description {
		t.Fatalf("round-trip domain mismatch: %+v", roundDom)
	}
	if *roundDom.Context.ToolName != toolName {
		t.Fatalf("round-trip context tool name mismatch: %v", roundDom.Context.ToolName)
	}
}

func TestAgenticMetricsLedgerDTO_AnomaliesSerialization(t *testing.T) {
	tid := "AL-SCHEMA-TEST"
	tool := "bash"
	turn := int64(5)
	count := int64(5)

	t.Run("anomalies present serialize to JSON and round-trip cleanly", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID:  &tid,
			UpdatedAt: "2026-10-18T10:00:00Z",
			Anomalies: []domain.Anomaly{
				{
					RuleID:      domain.RuleToolRetryStreak,
					Severity:    domain.SeverityCritical,
					Title:       "Tool retry streak: bash (5 consecutive errors)",
					Description: "Tool 'bash' encountered 5 consecutive execution failures (session: s1, turn: 5)",
					Context: domain.AnomalyContext{
						ToolName: &tool,
						Turn:     &turn,
						Count:    &count,
					},
				},
			},
		}

		dto := adapters.SerializeLedger(ledger)
		if len(dto.Anomalies) != 1 {
			t.Fatalf("expected 1 anomaly in DTO, got %d", len(dto.Anomalies))
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal DTO: %v", err)
		}
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"anomalies"`) {
			t.Errorf("expected anomalies in JSON output: %s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"TOOL_RETRY_STREAK"`) || !strings.Contains(jsonStr, `"CRITICAL"`) {
			t.Errorf("expected TOOL_RETRY_STREAK and CRITICAL in JSON: %s", jsonStr)
		}

		var parsed adapters.AgenticMetricsLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if len(parsed.Anomalies) != 1 {
			t.Fatalf("expected 1 parsed anomaly, got %d", len(parsed.Anomalies))
		}

		roundDom := parsed.ToDomain()
		if len(roundDom.Anomalies) != 1 {
			t.Fatalf("expected 1 domain anomaly, got %d", len(roundDom.Anomalies))
		}
		if roundDom.Anomalies[0].RuleID != domain.RuleToolRetryStreak {
			t.Errorf("expected RuleToolRetryStreak, got %s", roundDom.Anomalies[0].RuleID)
		}
	})

	t.Run("anomalies omitted when empty due to omitempty tag", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID:  &tid,
			UpdatedAt: "2026-10-18T10:00:00Z",
			Anomalies: nil,
		}

		dto := adapters.SerializeLedger(ledger)
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if strings.Contains(string(data), `"anomalies"`) {
			t.Errorf("expected anomalies to be omitted when empty: %s", string(data))
		}
	})

	t.Run("backward compatibility: legacy JSON without anomalies decodes safely", func(t *testing.T) {
		legacyJSON := `{
			"ticket_id": "AL-LEGACY",
			"updated_at": "2026-10-18T10:00:00Z",
			"totals": {"invocations": 1, "cost_usd": 0.05}
		}`

		var dto adapters.AgenticMetricsLedgerDTO
		if err := json.Unmarshal([]byte(legacyJSON), &dto); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if dto.Anomalies != nil {
			t.Errorf("expected nil Anomalies in DTO, got: %+v", dto.Anomalies)
		}

		dom := dto.ToDomain()
		if len(dom.Anomalies) != 0 {
			t.Errorf("expected empty domain Anomalies slice, got len %d", len(dom.Anomalies))
		}
	})

	t.Run("corrupted anomalies field returns descriptive unmarshal error", func(t *testing.T) {
		badJSON := `{
			"ticket_id": "AL-BAD",
			"updated_at": "2026-10-18T10:00:00Z",
			"anomalies": "not_an_array"
		}`

		var dto adapters.AgenticMetricsLedgerDTO
		err := json.Unmarshal([]byte(badJSON), &dto)
		if err == nil {
			t.Fatal("expected unmarshal error for non-array anomalies, got nil")
		}
	})
}

func TestBrittleToolSummaryDTO_BidirectionalMapping(t *testing.T) {
	t.Run("fully populated brittle tool maps bidirectionally", func(t *testing.T) {
		dom := domain.BrittleToolSummary{
			ToolName:       "bash",
			AnomalyCount:   2,
			TotalCalls:     25,
			ErrorCount:     6,
			FailureRatePct: 24.0,
		}

		dto := adapters.BrittleToolSummaryDTOFromDomain(dom)
		if dto.ToolName != "bash" || dto.AnomalyCount != 2 || dto.TotalCalls != 25 ||
			dto.ErrorCount != 6 || dto.FailureRatePct != 24.0 {
			t.Fatalf("DTO mapping mismatch: %+v", dto)
		}

		roundDom := dto.ToDomain()
		if roundDom.ToolName != dom.ToolName || roundDom.AnomalyCount != dom.AnomalyCount ||
			roundDom.TotalCalls != dom.TotalCalls || roundDom.ErrorCount != dom.ErrorCount ||
			roundDom.FailureRatePct != dom.FailureRatePct {
			t.Fatalf("round-trip domain mismatch: %+v", roundDom)
		}
	})

	t.Run("json serialization round-trip", func(t *testing.T) {
		dto := adapters.BrittleToolSummaryDTO{
			ToolName:       "git",
			AnomalyCount:   1,
			TotalCalls:     10,
			ErrorCount:     3,
			FailureRatePct: 30.0,
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var decoded adapters.BrittleToolSummaryDTO
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if decoded != dto {
			t.Fatalf("expected decoded %+v, got %+v", dto, decoded)
		}
	})
}

func TestFleetAnomalySummaryDTO_BidirectionalMapping(t *testing.T) {
	t.Run("fully populated fleet anomaly summary maps bidirectionally", func(t *testing.T) {
		dom := domain.FleetAnomalySummary{
			TotalAnomalies:       4,
			WarningCount:         3,
			CriticalCount:        1,
			SeverityCounts:       map[string]int64{"WARNING": 3, "CRITICAL": 1},
			RuleCounts:           map[string]int64{"TOOL_RETRY_STREAK": 2, "TOOL_HIGH_FAILURE_RATE": 1, "CONTEXT_SATURATION_SPIKE": 1, "SUBAGENT_REWORK_LOOP": 0},
			TotalTickets:         5,
			TicketsWithAnomalies: 2,
			AnomalyRatePerTicket: 0.8,
			BrittleTools: []domain.BrittleToolSummary{
				{
					ToolName:       "bash",
					AnomalyCount:   2,
					TotalCalls:     20,
					ErrorCount:     5,
					FailureRatePct: 25.0,
				},
			},
		}

		dto := adapters.FleetAnomalySummaryDTOFromDomain(dom)
		if dto.TotalAnomalies != 4 || dto.WarningCount != 3 || dto.CriticalCount != 1 ||
			dto.TotalTickets != 5 || dto.TicketsWithAnomalies != 2 || dto.AnomalyRatePerTicket != 0.8 {
			t.Fatalf("DTO fields mismatch: %+v", dto)
		}
		if len(dto.BrittleTools) != 1 || dto.BrittleTools[0].ToolName != "bash" {
			t.Fatalf("DTO brittle tools mismatch: %+v", dto.BrittleTools)
		}

		roundDom := dto.ToDomain()
		if roundDom.TotalAnomalies != dom.TotalAnomalies ||
			roundDom.WarningCount != dom.WarningCount ||
			roundDom.CriticalCount != dom.CriticalCount ||
			roundDom.TotalTickets != dom.TotalTickets ||
			roundDom.TicketsWithAnomalies != dom.TicketsWithAnomalies ||
			roundDom.AnomalyRatePerTicket != dom.AnomalyRatePerTicket {
			t.Fatalf("round-trip domain mismatch: %+v", roundDom)
		}
		if len(roundDom.BrittleTools) != 1 || roundDom.BrittleTools[0].ToolName != "bash" {
			t.Fatalf("round-trip brittle tools mismatch: %+v", roundDom.BrittleTools)
		}
		if roundDom.SeverityCounts["WARNING"] != 3 || roundDom.SeverityCounts["CRITICAL"] != 1 {
			t.Fatalf("round-trip severity counts mismatch: %+v", roundDom.SeverityCounts)
		}
		if roundDom.RuleCounts["TOOL_RETRY_STREAK"] != 2 {
			t.Fatalf("round-trip rule counts mismatch: %+v", roundDom.RuleCounts)
		}
	})

	t.Run("json serialization round-trip", func(t *testing.T) {
		dto := adapters.FleetAnomalySummaryDTO{
			TotalAnomalies:       0,
			WarningCount:         0,
			CriticalCount:        0,
			SeverityCounts:       map[string]int64{"WARNING": 0, "CRITICAL": 0},
			RuleCounts:           map[string]int64{"TOOL_RETRY_STREAK": 0},
			TotalTickets:         0,
			TicketsWithAnomalies: 0,
			AnomalyRatePerTicket: 0.0,
			BrittleTools:         []adapters.BrittleToolSummaryDTO{},
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var decoded adapters.FleetAnomalySummaryDTO
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		if decoded.TotalAnomalies != 0 || len(decoded.BrittleTools) != 0 {
			t.Fatalf("expected empty decoded summary, got: %+v", decoded)
		}
	})
}
