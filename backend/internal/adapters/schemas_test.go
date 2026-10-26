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

func TestAgenticFactLedgerDTO_PersistsSessionAndTicketAnomalyScopes(t *testing.T) {
	sid, tool := "s1", "bash"
	count, calls, rate, threshold := int64(3), int64(10), 30.0, ">20%"
	nested := domain.Anomaly{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning, Title: "Retry", Description: "three retries", Context: domain.AnomalyContext{SessionID: &sid, ToolName: &tool, Count: &count}}
	aggregate := domain.Anomaly{RuleID: domain.RuleToolHighFailureRate, Severity: domain.SeverityWarning, Title: "High tool failure rate", Description: "aggregate", Context: domain.AnomalyContext{ToolName: &tool, Count: &count, TotalCalls: &calls, FailureRatePct: &rate, Threshold: &threshold}}
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{{ID: sid, Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}, Events: []domain.SessionEventFact{}, Anomalies: []domain.Anomaly{nested}}}, Anomalies: []domain.Anomaly{aggregate}}
	data, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), `"session_id":"s1"`) {
		t.Fatalf("nested summary repeats its owning session id: %s", data)
	}
	var decoded adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.ToDomain()
	facts := got.AnomalyFacts()
	if len(facts) != 2 || facts[1].Context.SessionID == nil {
		t.Fatalf("summary scope was not restored on read: %#v", facts)
	}
	if got.Anomalies[0].Context.SessionID != nil {
		t.Fatalf("ticket-level summary gained a session id: %#v", got.Anomalies[0])
	}
}

func TestAgenticFactLedgerDTO_OmitsEventsAndAnomalyEvidence(t *testing.T) {
	version := domain.CurrentSchemaVersion
	tool := "bash"
	event := domain.SessionEventFact{EventID: "event-1", Timestamp: "2026-10-26T10:00:00Z", Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"}
	facts := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{{ID: "s1", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{{Name: tool, Calls: 1, Errors: 1}}, Events: []domain.SessionEventFact{event}, Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Evidence: []domain.AnomalyEvidenceRef{{SessionID: "s1", EventID: "event-1"}}}}}}}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "events") || strings.Contains(string(encoded), "evidence") {
		t.Fatalf("compact DTO serialized transient rows or refs: %s", encoded)
	}
	var legacy adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal([]byte(`{"schema_version":4,"sessions":[{"id":"s1","invocations":[],"tools":[],"events":[{"event_id":"event-1","timestamp":"2026-10-26T10:00:00Z","role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"}],"anomalies":[{"rule_id":"TOOL_RETRY_STREAK","severity":"WARNING","title":"retry","description":"retry","context":{},"evidence":[{"session_id":"s1","event_id":"event-1"}]}]}]}`), &legacy); err != nil {
		t.Fatalf("legacy rows should decode for normalization: %v", err)
	}
	if len(legacy.Sessions[0].Events) != 1 || len(legacy.Sessions[0].Anomalies[0].Evidence) != 1 {
		t.Fatal("legacy rows weren't available to the normalizer")
	}
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
		role := domain.AgentRole("coder")
		secRole := domain.AgentRole("tester")
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
			*dto.Role != string(role) || *dto.SecondaryRole != string(secRole) || *dto.Count != count ||
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

func TestSchemaMigrationDTOBoundaries(t *testing.T) {
	t.Run("malformed nested session records are rejected instead of defaulted", func(t *testing.T) {
		missingParent := "missing"
		cases := []struct {
			name    string
			session adapters.SessionFactDTO
		}{
			{"blank tool name", adapters.SessionFactDTO{ID: "s1", Tools: []adapters.SessionToolFactDTO{{Calls: 1}}}},
			{"negative tool calls", adapters.SessionFactDTO{ID: "s1", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: -1}}}},
			{"errors exceed calls", adapters.SessionFactDTO{ID: "s1", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: 1, Errors: 2}}}},
			{"duplicate tool name", adapters.SessionFactDTO{ID: "s1", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: 1}, {Name: "bash", Calls: 2}}}},
			{"dangling invocation parent", adapters.SessionFactDTO{ID: "s1", Invocations: []adapters.InvocationFactDTO{{ID: "i1", ParentID: &missingParent}}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				dto := adapters.AgenticFactLedgerDTO{Sessions: []adapters.SessionFactDTO{tc.session}}
				if err := dto.Validate(); err == nil {
					t.Fatal("expected malformed nested record to fail validation")
				}
			})
		}
	})
	t.Run("legacy event DTO decodes but output omits rows", func(t *testing.T) {
		var decoded adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(`{"schema_version":4,"sessions":[{"id":"s1","invocations":[],"tools":[],"events":[{"event_id":"e1","timestamp":"2026-10-18T10:00:00Z","role":"coder","event_type":"tool_call","status":"error"}]}]}`), &decoded); err != nil {
			t.Fatalf("expected legacy event rows to be readable for compaction: %v", err)
		}
		data, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(decoded.ToDomain()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"events"`) || strings.Contains(string(data), `"evidence"`) {
			t.Fatalf("canonical DTO serialized event rows or evidence refs: %s", data)
		}
	})

	t.Run("empty ledger serializes a sessions collection without root facts", func(t *testing.T) {
		v4 := 4
		tid := "AL-V4-EMPTY"
		ledger := domain.AgenticFactLedger{
			SchemaVersion: &v4,
			TicketID:      &tid,
			UpdatedAt:     "2026-10-18T10:00:00Z", Sessions: []domain.SessionFact{},
		}

		dto := adapters.AgenticFactLedgerDTOFromDomain(ledger)
		if dto.Sessions == nil {
			t.Fatal("expected DTO.Sessions to be initialized")
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"sessions":[]`) || strings.Contains(jsonStr, `"invocations"`) || strings.Contains(jsonStr, `"tools"`) || strings.Contains(jsonStr, `"events"`) {
			t.Errorf("expected an empty sessions collection with no root facts, got %s", jsonStr)
		}
	})

	t.Run("session event serialization omits repeated session identity", func(t *testing.T) {
		sid := "session-fixture"
		tool := "bash"
		ledger := domain.AgenticFactLedger{Sessions: sessionFacts(nil,
			[]domain.ToolFact{{SessionID: sid, Name: tool, Calls: 2}},
			[]domain.EventFact{{SessionID: sid, EventType: domain.EventTypeToolCall, ToolName: &tool, Status: "error"}})}
		data, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(data), `"id":"session-fixture"`) != 1 || strings.Contains(string(data), `"session_id"`) {
			t.Fatalf("expected session identity only on the session, got %s", data)
		}
	})

	t.Run("boundary unmarshaling with malformed schema types", func(t *testing.T) {
		corruptedInputs := []struct {
			name    string
			payload string
		}{
			{"schema_version string", `{"schema_version": "2"}`},
			{"schema_version float", `{"schema_version": 2.5}`},
			{"events not an array", `{"schema_version": 2, "events": true}`},
			{"events array with invalid element", `{"schema_version": 2, "events": [123]}`},
			{"code not an object", `{"code": "not_an_object"}`},
			{"invocations not an array", `{"invocations": {}}`},
			{"migration not an object", `{"migration": ["git_code_facts"]}`},
		}

		for _, tc := range corruptedInputs {
			t.Run(tc.name, func(t *testing.T) {
				var dto adapters.AgenticFactLedgerDTO
				err := json.Unmarshal([]byte(tc.payload), &dto)
				if err == nil {
					t.Errorf("expected unmarshal error for %s, got nil", tc.name)
				}
			})
		}
	})

	t.Run("empty json object ToDomain defaults safely without panic", func(t *testing.T) {
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(`{}`), &dto); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		dom := dto.ToDomain()
		if dom.SchemaVersion != nil {
			t.Errorf("expected nil schema_version for empty JSON")
		}
		if dom.Code.SDLCIterations != 0 {
			t.Errorf("expected zero SDLC iterations to remain zero, got %d", dom.Code.SDLCIterations)
		}
		if len(dom.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got len %d", len(dom.EventFacts()))
		}
		if len(dom.InvocationFacts()) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(dom.InvocationFacts()))
		}
		if len(dom.ToolFacts()) != 0 {
			t.Errorf("expected 0 tools, got %d", len(dom.ToolFacts()))
		}
	})
}

func TestToolSessionIDRequiredAtLedgerBoundary(t *testing.T) {
	cases := []struct {
		name      string
		sessionID string
		valid     bool
	}{
		{name: "missing"},
		{name: "empty", sessionID: ""},
		{name: "whitespace", sessionID: " \t\n"},
		{name: "present", sessionID: "thread-fixture", valid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tools := []adapters.SessionToolFactDTO{{Name: "command", Calls: 1}}
			dto := adapters.AgenticFactLedgerDTO{Sessions: []adapters.SessionFactDTO{{ID: tc.sessionID, Invocations: []adapters.InvocationFactDTO{}, Tools: tools, Events: []adapters.SessionEventFactDTO{}}}}
			err := dto.Validate()
			if tc.valid && err != nil {
				t.Fatalf("expected valid session identity, got %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected missing session identity to be rejected")
			}
		})
	}
	if err := (adapters.AgenticFactLedgerDTO{Sessions: []adapters.SessionFactDTO{}}).Validate(); err != nil {
		t.Fatalf("empty tools collection should be valid: %v", err)
	}
}

func TestCodeFactsDTOToDomainPreservesZeroIterations(t *testing.T) {
	got := (adapters.CodeFactsDTO{SDLCIterations: 0}).ToDomain()
	if got.SDLCIterations != 0 {
		t.Fatalf("expected exact zero SDLC iterations, got %d", got.SDLCIterations)
	}
}

func TestInvocationFactDTO_DynamicRolesAndModels(t *testing.T) {
	t.Parallel()

	t.Run("arbitrary custom roles and model names preserved without mutation", func(t *testing.T) {
		t.Parallel()
		customCases := []struct {
			role  string
			model string
		}{
			{role: "advisor", model: "claude-3-7-sonnet"},
			{role: "po", model: "gemini-2.5-pro"},
			{role: "researcher", model: "deepseek-r1"},
			{role: "security-auditor", model: "gpt-4o-custom"},
			{role: "domain-expert", model: "meta-llama-3.3-70b"},
		}

		for _, tc := range customCases {
			dto := adapters.InvocationFactDTO{
				ID:    "inv-1",
				Role:  tc.role,
				Model: tc.model,
				User:  "test-user",
			}

			dom := dto.ToDomain()
			if string(dom.Role) != tc.role {
				t.Errorf("expected role %q preserved, got %q", tc.role, dom.Role)
			}
			if dom.Model != tc.model {
				t.Errorf("expected model %q preserved, got %q", tc.model, dom.Model)
			}
		}
	})

	t.Run("empty role model and user remain empty without fallback values", func(t *testing.T) {
		t.Parallel()
		dto := adapters.InvocationFactDTO{
			ID:    "inv-empty",
			Role:  "",
			Model: "",
			User:  "",
		}

		dom := dto.ToDomain()
		if dom.Role != "" {
			t.Errorf("expected empty role to remain empty, got %q", dom.Role)
		}
		if dom.Model != "" {
			t.Errorf("expected empty model to remain empty, got %q", dom.Model)
		}
		if dom.User != "" {
			t.Errorf("expected empty user to remain empty, got %q", dom.User)
		}
	})

	t.Run("json unmarshaling with arbitrary roles and models", func(t *testing.T) {
		t.Parallel()
		jsonBlob := `{
			"id": "ses-dyn-1",
			"role": "security-auditor",
			"model": "claude-3-7-sonnet-custom",
			"user": "alice",
			"cost_usd": 1.75,
			"turns": 6,
			"tokens": {
				"input": 15000,
				"output": 3000,
				"cache_read": 8000,
				"reasoning": 1000,
				"peak_context": 115000
			}
		}`

		var dto adapters.InvocationFactDTO
		if err := json.Unmarshal([]byte(jsonBlob), &dto); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}

		dom := dto.ToDomain()
		if dom.Role != "security-auditor" {
			t.Errorf("expected role 'security-auditor', got %q", dom.Role)
		}
		if dom.Model != "claude-3-7-sonnet-custom" {
			t.Errorf("expected model 'claude-3-7-sonnet-custom', got %q", dom.Model)
		}
		if dom.CostUSD != 1.75 {
			t.Errorf("expected cost 1.75, got %f", dom.CostUSD)
		}
		if dom.Tokens.PeakContext != 115000 {
			t.Errorf("expected peak context 115000, got %d", dom.Tokens.PeakContext)
		}
	})
}

func TestEventFactDTO_DynamicRoles(t *testing.T) {
	t.Parallel()

	t.Run("custom roles preserved in event stream", func(t *testing.T) {
		t.Parallel()
		dto := adapters.EventFactDTO{
			Timestamp: "2026-10-18T12:00:00Z",
			SessionID: "s-1",
			Turn:      1,
			Role:      "security-auditor",
			EventType: "tool_call",
			Status:    "completed",
		}

		dom := dto.ToDomain()
		if dom.Role != "security-auditor" {
			t.Errorf("expected role 'security-auditor', got %q", dom.Role)
		}
	})

	t.Run("empty role and status default safely", func(t *testing.T) {
		t.Parallel()
		dto := adapters.EventFactDTO{
			Timestamp: "2026-10-18T12:00:00Z",
			SessionID: "s-1",
			Turn:      1,
			Role:      "",
			EventType: "turn_start",
			Status:    "",
		}

		dom := dto.ToDomain()
		if dom.Role != "orchestrator" {
			t.Errorf("expected default 'orchestrator', got %q", dom.Role)
		}
		if dom.Status != "completed" {
			t.Errorf("expected default 'completed', got %q", dom.Status)
		}
	})
}

func TestAgenticMetricsLedgerDTO_DynamicRolesAndModelsSerialization(t *testing.T) {
	t.Parallel()
	tid := "AL-DYN-SERIAL"
	ledger := domain.AgenticMetricsLedger{
		TicketID:  &tid,
		UpdatedAt: "2026-10-18T12:00:00Z",
		Roles: []domain.RoleMetric{
			{Role: "orchestrator", CostUSD: 1.0},
			{Role: "coder", CostUSD: 2.0},
			{Role: "advisor", CostUSD: 1.5},
			{Role: "po", CostUSD: 0.5},
		},
		Models: []domain.ModelMetric{
			{Model: "claude-3-7-sonnet", CostUSD: 3.5},
			{Model: "gemini-2.5-pro", CostUSD: 1.5},
		},
	}

	dto := adapters.SerializeLedger(ledger)
	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed adapters.AgenticMetricsLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	roundDom := parsed.ToDomain()
	if len(roundDom.Roles) != 4 {
		t.Fatalf("expected 4 roles, got %d", len(roundDom.Roles))
	}
	if roundDom.Roles[2].Role != "advisor" || roundDom.Roles[3].Role != "po" {
		t.Errorf("custom roles corrupted during serialization: %+v", roundDom.Roles)
	}

	if len(roundDom.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(roundDom.Models))
	}
	if roundDom.Models[0].Model != "claude-3-7-sonnet" || roundDom.Models[1].Model != "gemini-2.5-pro" {
		t.Errorf("custom models corrupted during serialization: %+v", roundDom.Models)
	}
}

func TestTicketTypeBoundarySchemaAndSerialization(t *testing.T) {
	t.Parallel()

	t.Run("explicit ticket_type null preserves nil pointer across round-trip", func(t *testing.T) {
		t.Parallel()
		jsonBlob := `{
			"schema_version": 2,
			"ticket_id": "AL-29",
			"ticket_type": null,
			"code": {
				"lines_added": 0,
				"lines_deleted": 0,
				"files_changed": 0,
				"sdlc_iterations": 1
			},
			"sessions": []
		}`

		var factDTO adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(jsonBlob), &factDTO); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if factDTO.TicketType != nil {
			t.Fatalf("expected nil TicketType in fact DTO, got %q", *factDTO.TicketType)
		}

		dom := factDTO.ToDomain()
		if dom.TicketType != nil {
			t.Fatalf("expected nil TicketType in domain fact ledger, got %q", *dom.TicketType)
		}

		serializedDTO := adapters.AgenticFactLedgerDTOFromDomain(dom)
		data, err := json.Marshal(serializedDTO)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}
		if !strings.Contains(string(data), `"ticket_type":null`) && !strings.Contains(string(data), `"ticket_type": null`) {
			t.Errorf("expected serialized JSON to contain 'ticket_type: null', got:\n%s", string(data))
		}
	})

	t.Run("omitted ticket_type defaults to nil pointer", func(t *testing.T) {
		t.Parallel()
		jsonBlob := `{
			"schema_version": 2,
			"ticket_id": "AL-29",
			"code": {
				"lines_added": 0,
				"lines_deleted": 0,
				"files_changed": 0,
				"sdlc_iterations": 1
			},
			"sessions": []
		}`

		var factDTO adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(jsonBlob), &factDTO); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if factDTO.TicketType != nil {
			t.Fatalf("expected nil TicketType when omitted, got %q", *factDTO.TicketType)
		}
	})

	t.Run("canonical ticket_type values round-trip through metrics and fact DTOs", func(t *testing.T) {
		t.Parallel()
		types := domain.CanonicalTicketTypes
		convPtr := func(tt domain.ConventionalTicketType) *domain.ConventionalTicketType { return &tt }
		for _, tt := range types {
			domFact := domain.AgenticFactLedger{
				TicketID:   strPtr("AL-29"),
				TicketType: convPtr(tt),
				Code: domain.CodeFacts{
					SDLCIterations: 1,
				},
			}
			dto := adapters.AgenticFactLedgerDTOFromDomain(domFact)
			if dto.TicketType == nil || *dto.TicketType != string(tt) {
				t.Errorf("expected %q in FactDTO, got %v", tt, dto.TicketType)
			}
			backToDom := dto.ToDomain()
			if backToDom.TicketType == nil || *backToDom.TicketType != tt {
				t.Errorf("expected %q back in domain, got %v", tt, backToDom.TicketType)
			}

			domMetrics := domain.AgenticMetricsLedger{
				TicketID:   strPtr("AL-29"),
				TicketType: convPtr(tt),
			}
			metricsDTO := adapters.SerializeLedger(domMetrics)
			if metricsDTO.TicketType == nil || *metricsDTO.TicketType != string(tt) {
				t.Errorf("expected %q in MetricsDTO, got %v", tt, metricsDTO.TicketType)
			}
			metricsBackToDom := metricsDTO.ToDomain()
			if metricsBackToDom.TicketType == nil || *metricsBackToDom.TicketType != tt {
				t.Errorf("expected %q back in metrics domain, got %v", tt, metricsBackToDom.TicketType)
			}
		}
	})

	t.Run("corrupted ticket_type formats return unmarshal error", func(t *testing.T) {
		t.Parallel()
		corruptedCases := []struct {
			name string
			json string
		}{
			{
				name: "ticket_type as integer number",
				json: `{"schema_version": 2, "ticket_id": "AL-29", "ticket_type": 123}`,
			},
			{
				name: "ticket_type as boolean",
				json: `{"schema_version": 2, "ticket_id": "AL-29", "ticket_type": true}`,
			},
			{
				name: "ticket_type as array",
				json: `{"schema_version": 2, "ticket_id": "AL-29", "ticket_type": ["feat"]}`,
			},
			{
				name: "ticket_type as nested object",
				json: `{"schema_version": 2, "ticket_id": "AL-29", "ticket_type": {"name": "feat"}}`,
			},
		}

		for _, tc := range corruptedCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				var factDTO adapters.AgenticFactLedgerDTO
				if err := json.Unmarshal([]byte(tc.json), &factDTO); err == nil {
					t.Fatalf("expected unmarshal error for %s, but succeeded with %+v", tc.name, factDTO)
				}

				var metricsDTO adapters.AgenticMetricsLedgerDTO
				if err := json.Unmarshal([]byte(tc.json), &metricsDTO); err == nil {
					t.Fatalf("expected unmarshal error for %s on metrics DTO, but succeeded with %+v", tc.name, metricsDTO)
				}
			})
		}
	})
}

func TestLiveSchema_AnomalyAlertSerializationAndParsing(t *testing.T) {
	t.Run("valid alert JSON unmarshals and roundtrips cleanly", func(t *testing.T) {
		payload := `{
			"session_id": "ses_test_1",
			"rule_name": "ConsecutiveToolErrors",
			"severity": "HIGH",
			"message": "Tool execution error streak reached 3 consecutive failures without recovery (last failing tool: bash).",
			"timestamp": "2026-10-19T10:00:00Z",
			"trigger_event": "part_3: exit status 1"
		}`

		var alert domain.LiveAnomalyAlert
		if err := json.Unmarshal([]byte(payload), &alert); err != nil {
			t.Fatalf("failed unmarshaling valid alert: %v", err)
		}

		if alert.SessionID != "ses_test_1" {
			t.Errorf("expected session_id ses_test_1, got %s", alert.SessionID)
		}
		if alert.RuleName != domain.RuleConsecutiveToolErrors {
			t.Errorf("expected rule %s, got %s", domain.RuleConsecutiveToolErrors, alert.RuleName)
		}
		if alert.Severity != domain.LiveSeverityHigh {
			t.Errorf("expected severity HIGH, got %s", alert.Severity)
		}

		marshaled, err := json.Marshal(alert)
		if err != nil {
			t.Fatalf("failed marshaling alert: %v", err)
		}
		if !strings.Contains(string(marshaled), `"rule_name":"ConsecutiveToolErrors"`) {
			t.Errorf("expected rule_name in marshaled JSON, got %s", string(marshaled))
		}
	})

	t.Run("malformed alert JSON fails with descriptive error", func(t *testing.T) {
		corrupted := `{ "session_id": "s1", "rule_name": `
		var alert domain.LiveAnomalyAlert
		if err := json.Unmarshal([]byte(corrupted), &alert); err == nil {
			t.Fatal("expected unmarshaling error for truncated json, got nil")
		}
	})

	t.Run("alert with unexpected field types fails cleanly", func(t *testing.T) {
		badType := `{ "session_id": 12345, "rule_name": "ConsecutiveToolErrors" }`
		var alert domain.LiveAnomalyAlert
		if err := json.Unmarshal([]byte(badType), &alert); err == nil {
			t.Fatal("expected unmarshaling error for integer session_id, got nil")
		}
	})
}

func TestLiveSchema_LiveTraceViewBoundary(t *testing.T) {
	t.Run("empty slice fields serialize as empty array [] and not null", func(t *testing.T) {
		trace := domain.LiveTraceView{
			SessionID:       "s_empty",
			Messages:        []domain.LiveMessageFact{},
			TotalToolCalls:  0,
			FailedToolCalls: 0,
			RecentErrors:    []domain.LivePartFact{},
			Anomalies:       []domain.LiveAnomalyAlert{},
		}

		data, err := json.Marshal(trace)
		if err != nil {
			t.Fatalf("failed marshaling trace view: %v", err)
		}

		str := string(data)
		if strings.Contains(str, `"messages":null`) {
			t.Errorf("messages serialized as null instead of []: %s", str)
		}
		if strings.Contains(str, `"recent_errors":null`) {
			t.Errorf("recent_errors serialized as null instead of []: %s", str)
		}
		if strings.Contains(str, `"anomalies":null`) {
			t.Errorf("anomalies serialized as null instead of []: %s", str)
		}
	})

	t.Run("unmarshaling handles negative and extreme token values without panic", func(t *testing.T) {
		payload := `{
			"session_id": "s_extreme",
			"messages": [
				{
					"id": "m1",
					"role": "assistant",
					"content": "extreme test",
					"timestamp": "2026-10-19T10:00:00Z",
					"context_tokens": 9223372036854775807,
					"input_tokens": 1000,
					"output_tokens": 500,
					"cache_read_tokens": 0,
					"reasoning_tokens": 0,
					"total_tokens": 1500,
					"parts": []
				}
			],
			"total_tool_calls": 0,
			"failed_tool_calls": 0,
			"recent_errors": [],
			"anomalies": []
		}`

		var trace domain.LiveTraceView
		if err := json.Unmarshal([]byte(payload), &trace); err != nil {
			t.Fatalf("failed unmarshaling extreme token values: %v", err)
		}

		if len(trace.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(trace.Messages))
		}
		if trace.Messages[0].ContextTokens != 9223372036854775807 {
			t.Errorf("expected max int64 context tokens, got %d", trace.Messages[0].ContextTokens)
		}
	})
}

func TestLiveSchema_LiveSessionTreeBoundary(t *testing.T) {
	t.Run("empty slice fields serialize as empty array [] and not null", func(t *testing.T) {
		tree := domain.LiveSessionTree{
			RootSessions:        []*domain.LiveSessionNode{},
			LastRefreshedAt:     "2026-10-19T10:00:00Z",
			TotalActiveSessions: 0,
			ActiveAnomalies:     []domain.LiveAnomalyAlert{},
		}

		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatalf("failed marshaling session tree: %v", err)
		}

		str := string(data)
		if strings.Contains(str, `"root_sessions":null`) {
			t.Errorf("root_sessions serialized as null instead of []: %s", str)
		}
		if strings.Contains(str, `"active_anomalies":null`) {
			t.Errorf("active_anomalies serialized as null instead of []: %s", str)
		}
	})

	t.Run("nested sessions roundtrip correctly", func(t *testing.T) {
		parentID := "ses_parent"
		child := &domain.LiveSessionNode{
			ID:               "ses_child",
			ParentID:         &parentID,
			Agent:            "coder",
			Role:             domain.AgentRole("coder"),
			Model:            "gemini-3.8-flash",
			Title:            "Child session",
			TimeCreated:      "2026-10-19T10:05:00Z",
			TimeUpdated:      "2026-10-19T10:10:00Z",
			Status:           "running",
			TicketConfidence: 1.0,
			Anomalies:        []domain.LiveAnomalyAlert{},
			SubSessions:      []*domain.LiveSessionNode{},
		}
		root := &domain.LiveSessionNode{
			ID:               parentID,
			Agent:            "build",
			Role:             domain.AgentRole("orchestrator"),
			Model:            "gemini-3.8-flash",
			Title:            "Root session",
			TimeCreated:      "2026-10-19T10:00:00Z",
			TimeUpdated:      "2026-10-19T10:15:00Z",
			Status:           "running",
			TicketConfidence: 1.0,
			Anomalies:        []domain.LiveAnomalyAlert{},
			SubSessions:      []*domain.LiveSessionNode{child},
		}
		tree := domain.LiveSessionTree{
			RootSessions:        []*domain.LiveSessionNode{root},
			LastRefreshedAt:     "2026-10-19T10:15:01Z",
			TotalActiveSessions: 2,
			ActiveAnomalies:     []domain.LiveAnomalyAlert{},
		}

		data, err := json.Marshal(tree)
		if err != nil {
			t.Fatalf("failed marshaling nested tree: %v", err)
		}

		var roundTrip domain.LiveSessionTree
		if err := json.Unmarshal(data, &roundTrip); err != nil {
			t.Fatalf("failed unmarshaling nested tree: %v", err)
		}

		if len(roundTrip.RootSessions) != 1 {
			t.Fatalf("expected 1 root session, got %d", len(roundTrip.RootSessions))
		}
		if len(roundTrip.RootSessions[0].SubSessions) != 1 {
			t.Fatalf("expected 1 sub-session, got %d", len(roundTrip.RootSessions[0].SubSessions))
		}
		if roundTrip.RootSessions[0].SubSessions[0].ID != "ses_child" {
			t.Errorf("expected child ID ses_child, got %s", roundTrip.RootSessions[0].SubSessions[0].ID)
		}
	})
}
