package adapters_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestAgenticFactLedgerDTOPreservesSkillAndTaskFacts(t *testing.T) {
	version := domain.CurrentSchemaVersion
	want := domain.AgenticFactLedger{
		SchemaVersion: &version,
		Skills:        []domain.SkillFact{{Name: "agentic-lens-dev", TurnIndex: 1, MessageIndex: 0, CallerRole: "coder", Timestamp: "2026-10-31T10:00:00Z", Success: true}},
		TaskLinks:     []domain.TaskLinkFact{{SubagentType: "tester", CostEvidence: domain.UnavailableCost(), Description: "Run tests", ParentTurn: 2, ChildSession: "child-1", Timestamp: "2026-10-31T10:01:00Z"}},
	}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(want))
	if err != nil {
		t.Fatal(err)
	}
	var decoded adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.ToDomain()
	if len(got.Skills) != 1 || got.Skills[0] != want.Skills[0] || len(got.TaskLinks) != 1 || got.TaskLinks[0] != want.TaskLinks[0] {
		t.Fatalf("DTO round trip got %+v, want %+v", got, want)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestCanonicalV4NestedEventsReadCompatibilityAndMetadata(t *testing.T) {
	eventJSON := `{"event_id":"call-1","timestamp":"2026-10-29T10:00:00Z","message_id":"message-1","turn":2,"role":"security-auditor","event_type":"tool_call","tool_name":"bash","status":"error","duration_ms":0,"context_tokens":1234}`
	bareJSON := `{"timestamp":"2026-10-29T09:00:00Z","turn":1,"role":"coder","event_type":"turn_message","status":"completed","context_tokens":0}`
	ledgerJSON := func(collection string) string {
		return `{"schema_version":4,"ticket_id":"FIX-12","sessions":[{"id":"owner","source":"opencode","user":"tester","invocations":[],"tools":[],"anomalies":[]` + collection + `}]}`
	}
	for _, tc := range []struct {
		name, collection string
		count            int
	}{
		{"omitted", "", 0}, {"empty", `,"events":[]`, 0}, {"metadata and missing ID", `,"events":[` + eventJSON + `,` + bareJSON + `]`, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte(ledgerJSON(tc.collection))
			if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
				t.Fatal(err)
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(data, &dto); err != nil {
				t.Fatal(err)
			}
			if err := dto.Validate(); err != nil {
				t.Fatal(err)
			}
			facts := dto.ToDomain()
			events := facts.EventFacts()
			if len(events) != tc.count {
				t.Fatalf("expected %d events, got %#v", tc.count, events)
			}
			if tc.count > 0 {
				want := domain.EventFact{EventID: "call-1", Timestamp: "2026-10-29T10:00:00Z", SessionID: "owner", MessageID: strPtr("message-1"), Turn: 2, Role: "security-auditor", EventType: domain.EventTypeToolCall, ToolName: strPtr("bash"), Status: "error", DurationMs: new(int64), ContextTokens: 1234}
				if !reflect.DeepEqual(events[1], want) || events[0].EventID != "" || events[0].SessionID != "owner" {
					t.Fatalf("nested metadata/owner lost: %#v", events)
				}
			}
			encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), `"events"`) || strings.Contains(string(encoded), `"evidence"`) {
				t.Fatalf("writer stopped compacting: %s", encoded)
			}
		})
	}
	for _, field := range []string{"session_id", "error_message", "prompt_content", "reasoning_content", "tool_input", "tool_output"} {
		t.Run("reject persisted "+field, func(t *testing.T) {
			row := strings.Replace(eventJSON, `"event_id":`, `"`+field+`":"unexpected","event_id":`, 1)
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal([]byte(ledgerJSON(`,"events":[`+row+`]`)), &dto); err == nil {
				t.Fatalf("persisted DTO accepted response-only %s", field)
			}
		})
	}
	for _, key := range []string{"events", "tools", "invocations"} {
		data := []byte(strings.Replace(ledgerJSON(""), `"schema_version":4`, `"schema_version":4,"`+key+`":[]`, 1))
		if err := adapters.ValidateRuntimeFactLedger(data); err == nil {
			t.Fatalf("accepted ambiguous root %s", key)
		}
	}
}

func TestCanonicalRetiredContentFieldsReadCompatibility(t *testing.T) {
	// This is a self-contained ledger in the shape emitted before content
	// classification was removed. Reading it must not require source hydration.
	const raw = `{"schema_version":4,"ticket_id":"TEST-101","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"anomalies":[],"tools":[{"name":"command","calls":2,"errors":1,"content_errors":1}],"events":[{"event_id":"call-1","timestamp":"2026-10-29T10:00:00Z","turn":1,"role":"coder","event_type":"tool_call","tool_name":"command","status":"completed","context_tokens":1234,"has_content_error":true,"content_severity":"medium","content_error_snippet":"syntax error: bad name"}]}]}`
	if err := adapters.ValidateRuntimeFactLedger([]byte(raw)); err != nil {
		t.Fatal(err)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal([]byte(raw), &dto); err != nil {
		t.Fatal(err)
	}
	if err := dto.Validate(); err != nil {
		t.Fatal(err)
	}
	facts := dto.ToDomain()
	tools, events := facts.ToolFacts(), facts.EventFacts()
	if len(facts.Sessions) != 1 || facts.Sessions[0].Source != domain.SourceCodex || len(tools) != 1 || tools[0] != (domain.ToolFact{SessionID: "owner", Name: "command", Calls: 2, Errors: 1}) {
		t.Fatalf("existing tool/source facts lost: %+v", facts)
	}
	if len(events) != 1 || events[0].EventID != "call-1" || events[0].SessionID != "owner" || events[0].Status != "completed" || events[0].ContextTokens != 1234 {
		t.Fatalf("existing event facts lost: %+v", events)
	}
	for _, projection := range []adapters.AgenticFactLedgerDTO{dto, adapters.AgenticFactLedgerDTOFromDomain(facts)} {
		encoded, err := json.Marshal(projection)
		if err != nil {
			t.Fatal(err)
		}
		for _, retired := range []string{"content_errors", "has_content_error", "content_severity", "content_error_snippet"} {
			if strings.Contains(string(encoded), `"`+retired+`"`) {
				t.Fatalf("retired field %s re-emitted: %s", retired, encoded)
			}
		}
	}
	for _, existing := range []string{`"content_errors":1`, `"content_severity":"medium"`} {
		invalid := strings.Replace(raw, existing, existing+`,"unknown_field":true`, 1)
		if err := json.Unmarshal([]byte(invalid), &dto); err == nil || !strings.Contains(err.Error(), "unknown_field") {
			t.Fatalf("unrelated unknown field must still be rejected: %v", err)
		}
	}
}

func TestValidateRuntimeFactLedgerRejectsLegacyVersionsBeforeFacts(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("v%d", version), func(t *testing.T) {
			data := []byte(fmt.Sprintf(`{"schema_version":%d,"invocations":[{"malformed":true}]}`, version))
			err := adapters.ValidateRuntimeFactLedger(data)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("version %d", version)) {
				t.Fatalf("expected version-specific legacy rejection, got %v", err)
			}
			if !strings.Contains(err.Error(), "source-backed") {
				t.Fatalf("expected migration and source-backed backfill guidance, got %v", err)
			}
		})
	}
}

func TestValidateRuntimeFactLedgerRejectsIncompleteV4(t *testing.T) {
	data := []byte(`{"schema_version":4,"sessions":[],"migration":{"status":"partial","steps_pending":["source_backfill"]}}`)
	if err := adapters.ValidateRuntimeFactLedger(data); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("expected incomplete v4 rejection, got %v", err)
	}
}

func TestValidateRuntimeFactLedgerAcceptsCanonicalV4WithoutMigrationMetadata(t *testing.T) {
	base := `{"schema_version":4,"sessions":[{"id":"s1","source":"opencode","user":"u","invocations":[{"role":"orchestrator","model":"m","started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":0,"tokens":{"input":0,"output":0,"cache_read":0,"reasoning":0,"peak_context":0},"cost_usd":0,"compactions":0,"user_iterations":0}],"tools":[],"anomalies":[]}]}`
	if err := adapters.ValidateRuntimeFactLedger([]byte(base)); err != nil {
		t.Fatalf("canonical v4 without workflow metadata should be accepted: %v", err)
	}
	for name, data := range map[string]string{
		"missing source":       strings.Replace(base, `,"source":"opencode"`, "", 1),
		"missing atomic field": strings.Replace(base, `,"user_iterations":0`, "", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if err := adapters.ValidateRuntimeFactLedger([]byte(data)); err == nil {
				t.Fatal("expected canonical v4 shape rejection")
			}
		})
	}
	for _, value := range []string{`"obsolete"`, `""`, `null`, `0`, `{}`} {
		t.Run("reject removed invocation key "+value, func(t *testing.T) {
			data := strings.Replace(base, `"role":`, `"id":`+value+`,"role":`, 1)
			if err := adapters.ValidateRuntimeFactLedger([]byte(data)); err == nil || !strings.Contains(err.Error(), "must not contain id") {
				t.Fatalf("expected explicit removed-key rejection, got %v", err)
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal([]byte(data), &dto); err == nil {
				t.Fatal("DTO decoding accepted the removed invocation key")
			}
		})
	}
}

func TestInvocationSerializationPreservesRowsAndSessionOwnershipWithoutID(t *testing.T) {
	version := domain.CurrentSchemaVersion
	rootID := "opaque-owner"
	started, completed := "2026-10-28T10:00:00Z", "2026-10-28T10:01:00Z"
	rows := []domain.InvocationFact{
		{Role: "coder", Model: "model-a", StartedAt: &started, CompletedAt: &completed, Tokens: domain.InvocationTokens{Input: 10}, CostUSD: 0.25, CostEvidence: domain.ObservedCost(0.25)},
		{Role: "coder", Model: "model-a", StartedAt: &started, CompletedAt: &completed, Tokens: domain.InvocationTokens{Input: 20}, CostUSD: 0.50, CostEvidence: domain.ObservedCost(0.50)},
		{Role: "reviewer", Model: "model-b", StartedAt: &started, CompletedAt: &completed, Tokens: domain.InvocationTokens{Input: 30}, CostUSD: 0.75, CostEvidence: domain.ObservedCost(0.75)},
	}
	facts := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{
		{ID: rootID, Source: domain.SourceCodex, User: "alice", Invocations: rows, Tools: []domain.SessionToolFact{}, Anomalies: []domain.Anomaly{}},
		{ID: "opaque-child", ParentID: &rootID, Source: domain.SourceCodex, User: "alice", Invocations: rows[:1], Tools: []domain.SessionToolFact{}, Anomalies: []domain.Anomaly{}},
	}, Anomalies: []domain.Anomaly{}}
	domainJSON, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	dtoJSON, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"domain": domainJSON, "DTO": dtoJSON} {
		t.Run(name, func(t *testing.T) {
			if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
				t.Fatalf("serialized rows are not canonical v4: %v", err)
			}
			var shape struct {
				Sessions []struct {
					Invocations []map[string]json.RawMessage `json:"invocations"`
				} `json:"sessions"`
			}
			if err := json.Unmarshal(data, &shape); err != nil {
				t.Fatal(err)
			}
			for si, session := range shape.Sessions {
				for ii, row := range session.Invocations {
					if _, exists := row["id"]; exists {
						t.Fatalf("session %d row %d serialized the removed key", si, ii)
					}
				}
			}
			var decoded adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			got := decoded.ToDomain()
			if len(got.Sessions) != 2 || got.Sessions[0].ID != rootID || got.Sessions[1].ID != "opaque-child" || got.Sessions[1].ParentID == nil || *got.Sessions[1].ParentID != rootID {
				t.Fatalf("session ownership changed: %+v", got.Sessions)
			}
			if !reflect.DeepEqual(got.Sessions[0].Invocations, rows) || !reflect.DeepEqual(got.Sessions[1].Invocations, rows[:1]) {
				t.Fatalf("invocation rows changed during serialization: %+v", got.Sessions)
			}
		})
	}
}

func TestAgenticFactLedgerDTOFromDomainPreservesTemporaryMigrationMetadata(t *testing.T) {
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{SchemaVersion: &version}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"migration"`) {
		t.Fatalf("completed ledger with nil migration metadata must omit migration: %s", encoded)
	}
	if err := adapters.ValidateRuntimeFactLedger(encoded); err != nil {
		t.Fatalf("metadata-free canonical v4 ledger should load: %v", err)
	}

	ledger.Migration = &domain.MigrationInfo{
		SchemaVersion:  domain.CurrentSchemaVersion,
		Status:         domain.MigrationStatusPartial,
		StepsCompleted: []string{},
		StepsPending:   []string{domain.MigrationStepSourceBackfill},
	}
	encoded, err = json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"migration"`) {
		t.Fatalf("in-progress migration metadata must be preserved: %s", encoded)
	}
	if err := adapters.ValidateRuntimeFactLedger(encoded); err == nil {
		t.Fatal("runtime must reject a ledger while migration work is pending")
	}
}

func TestValidateRuntimeFactLedgerRejectsIntermediateMigrationState(t *testing.T) {
	base := `{"schema_version":4,"migration":{"status":"partial","steps_completed":[],"steps_pending":["source_backfill"]},"sessions":[]}`
	err := adapters.ValidateRuntimeFactLedger([]byte(base))
	if err == nil || !strings.Contains(err.Error(), "resume") {
		t.Fatalf("expected actionable incomplete migration rejection, got %v", err)
	}
}

func TestAgenticFactLedgerDTO_PersistsSessionAndTicketAnomalyScopes(t *testing.T) {
	sid, tool := "s1", "bash"
	count, calls, rate, threshold := int64(3), int64(10), 30.0, ">20%"
	nested := domain.Anomaly{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{SessionID: &sid, ToolName: &tool, Count: &count}}
	aggregate := domain.Anomaly{RuleID: domain.RuleToolHighFailureRate, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{ToolName: &tool, Count: &count, TotalCalls: &calls, FailureRatePct: &rate, Threshold: &threshold}}
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
	legacyAnomaly, err := json.Marshal(legacy.Sessions[0].Anomalies[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(legacyAnomaly), `"title"`) || strings.Contains(string(legacyAnomaly), `"description"`) {
		t.Fatalf("legacy presentation strings were not normalized away: %s", legacyAnomaly)
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
		RuleID:   domain.RuleToolHighFailureRate,
		Severity: domain.SeverityCritical,
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
	if *dto.Context.ToolName != toolName {
		t.Errorf("context tool name mismatch: %v", dto.Context.ToolName)
	}

	roundDom := dto.ToDomain()
	if roundDom.RuleID != dom.RuleID || roundDom.Severity != dom.Severity {
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

	t.Run("anomalies present preserve emitted rule severity and context", func(t *testing.T) {
		ledger := domain.AgenticMetricsLedger{
			TicketID:  &tid,
			UpdatedAt: "2026-10-18T10:00:00Z",
			Anomalies: []domain.Anomaly{
				{
					RuleID:   domain.RuleToolRetryStreak,
					Severity: domain.SeverityCritical,
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
		if strings.Contains(jsonStr, `"title"`) || strings.Contains(jsonStr, `"description"`) {
			t.Errorf("presentation text must not be persisted: %s", jsonStr)
		}

		var parsed adapters.AgenticMetricsLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if len(parsed.Anomalies) != 1 {
			t.Fatalf("expected 1 parsed anomaly, got %d", len(parsed.Anomalies))
		}

		anomaly := parsed.Anomalies[0]
		if anomaly.RuleID != string(domain.RuleToolRetryStreak) || anomaly.Severity != string(domain.SeverityCritical) || anomaly.Context.ToolName == nil || *anomaly.Context.ToolName != "bash" || anomaly.Context.Turn == nil || *anomaly.Context.Turn != 5 || anomaly.Context.Count == nil || *anomaly.Context.Count != 5 {
			t.Fatalf("emitted anomaly lost rule, severity or context: %+v", anomaly)
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

}

func TestSchemaMigrationDTOBoundaries(t *testing.T) {
	t.Run("fact ledger rejects malformed anomaly collection", func(t *testing.T) {
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(`{"schema_version":4,"sessions":[],"anomalies":"not_an_array"}`), &dto); err == nil {
			t.Fatal("fact input accepted malformed anomaly collection")
		}
	})

	t.Run("malformed nested session records are rejected instead of defaulted", func(t *testing.T) {
		missingParent := "missing"
		cases := []struct {
			name    string
			session adapters.SessionFactDTO
		}{
			{"blank tool name", adapters.SessionFactDTO{ID: "s1", User: "alice", Tools: []adapters.SessionToolFactDTO{{Calls: 1}}}},
			{"negative tool calls", adapters.SessionFactDTO{ID: "s1", User: "alice", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: -1}}}},
			{"errors exceed calls", adapters.SessionFactDTO{ID: "s1", User: "alice", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: 1, Errors: 2}}}},
			{"duplicate tool name", adapters.SessionFactDTO{ID: "s1", User: "alice", Tools: []adapters.SessionToolFactDTO{{Name: "bash", Calls: 1}, {Name: "bash", Calls: 2}}}},
			{"dangling session parent", adapters.SessionFactDTO{ID: "s1", User: "alice", ParentID: &missingParent}},
			{"blank session user", adapters.SessionFactDTO{ID: "s1", User: "   "}},
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
				Role:  tc.role,
				Model: tc.model,
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

	t.Run("empty role and model remain empty without fallback values", func(t *testing.T) {
		t.Parallel()
		dto := adapters.InvocationFactDTO{
			Role:  "",
			Model: "",
		}

		dom := dto.ToDomain()
		if dom.Role != "" {
			t.Errorf("expected empty role to remain empty, got %q", dom.Role)
		}
		if dom.Model != "" {
			t.Errorf("expected empty model to remain empty, got %q", dom.Model)
		}
	})

	t.Run("json unmarshaling with arbitrary roles and models", func(t *testing.T) {
		t.Parallel()
		jsonBlob := `{
			"role": "security-auditor",
			"model": "claude-3-7-sonnet-custom",
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

	if len(parsed.Roles) != 4 {
		t.Fatalf("expected 4 roles, got %d", len(parsed.Roles))
	}
	if parsed.Roles[2].Role != "advisor" || parsed.Roles[3].Role != "po" {
		t.Errorf("custom roles corrupted during serialization: %+v", parsed.Roles)
	}

	if len(parsed.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(parsed.Models))
	}
	if parsed.Models[0].Model != "claude-3-7-sonnet" || parsed.Models[1].Model != "gemini-2.5-pro" {
		t.Errorf("custom models corrupted during serialization: %+v", parsed.Models)
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
			data, err := json.Marshal(metricsDTO)
			if err != nil {
				t.Fatal(err)
			}
			var emitted struct {
				TicketType string `json:"ticket_type"`
			}
			if err := json.Unmarshal(data, &emitted); err != nil {
				t.Fatal(err)
			}
			if emitted.TicketType != string(tt) {
				t.Fatalf("computed export ticket type=%q, want %q", emitted.TicketType, tt)
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

func TestSessionFactDTO_SourceRoundtripAndRequired(t *testing.T) {
	t.Run("session source roundtrips accurately", func(t *testing.T) {
		dto := adapters.AgenticFactLedgerDTO{
			Sessions: []adapters.SessionFactDTO{
				{ID: "01a17362-baa4-72b3-934d-3cb040c899ea", Source: "codex"},
				{ID: "ses_abc123", Source: "opencode"},
			},
		}
		dom := dto.ToDomain()
		if len(dom.Sessions) != 2 {
			t.Fatalf("expected 2 sessions, got %d", len(dom.Sessions))
		}
		if dom.Sessions[0].Source != "codex" {
			t.Errorf("expected codex, got %s", dom.Sessions[0].Source)
		}
		if dom.Sessions[1].Source != "opencode" {
			t.Errorf("expected opencode, got %s", dom.Sessions[1].Source)
		}

		back := adapters.AgenticFactLedgerDTOFromDomain(dom)
		if back.Sessions[0].Source != "codex" {
			t.Errorf("expected codex in DTO, got %s", back.Sessions[0].Source)
		}
		if back.Sessions[1].Source != "opencode" {
			t.Errorf("expected opencode in DTO, got %s", back.Sessions[1].Source)
		}
	})

	t.Run("validation requires session source", func(t *testing.T) {
		dto := adapters.AgenticFactLedgerDTO{
			Sessions: []adapters.SessionFactDTO{
				{ID: "ses_missing_source", Source: ""},
			},
		}
		if err := dto.ValidatePersistedEvidence(); err == nil {
			t.Error("expected error for missing session source, got nil")
		}
	})
}

func TestCostEvidenceBoundaryCompatibility(t *testing.T) {

	legacyBase := `{"schema_version":4,"sessions":[{"id":"owner","source":"codex","user":"alice","invocations":[{"role":"coder","model":"model-a","started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":1,"tokens":{"input":1,"output":0,"cache_read":0,"reasoning":0,"peak_context":1},COST"compactions":0,"user_iterations":0}],"tools":[],"anomalies":[]}]}`
	t.Run("typed task presence and partial facts survive concrete roundtrip", func(t *testing.T) {
		for _, version := range []int{4, 5} {
			for _, tc := range []struct {
				name, fields string
				state        domain.CostState
				amount       float64
			}{
				{"observed zero", `"cost_usd":0`, domain.CostObserved, 0},
				{"absent", ``, domain.CostUnavailable, 0},
				{"null", `"cost_usd":null`, domain.CostUnavailable, 0},
				{"partial", `"cost_usd":null,"cost_evidence":{"state":"partial","known_subtotal_usd":2.5}`, domain.CostPartial, 2.5},
			} {
				t.Run(fmt.Sprintf("v%d/%s", version, tc.name), func(t *testing.T) {
					fields := tc.fields
					if version == 5 && tc.name == "absent" {
						fields = `"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}`
					} else if version == 5 && tc.state != domain.CostPartial {
						subtotal := "null"
						if tc.state == domain.CostObserved {
							subtotal = "0"
						}
						fields += fmt.Sprintf(`,"cost_evidence":{"state":%q,"known_subtotal_usd":%s}`, tc.state, subtotal)
					}
					core := strings.Replace(legacyBase, `"schema_version":4`, fmt.Sprintf(`"schema_version":%d`, version), 1)
					core = strings.Replace(core, "COST", `"cost_usd":null,"cost_evidence":{"state":"partial","known_subtotal_usd":1.25},`, 1)
					task := `{"subagent_type":"coder","description":"Keep source task","parent_turn":8,"child_session_id":"child","timestamp":"2026-11-01T10:00:00Z","duration_ms":12`
					if fields != "" {
						task += "," + fields
					}
					task += "}"
					input := []byte(strings.TrimSuffix(core, "}") + `,"task_links":[` + task + `]}`)
					if version == 5 && tc.name == "absent" {
						if err := adapters.ValidateRuntimeFactLedger(input); err == nil {
							t.Fatal("v5 runtime accepted task with omitted cost_usd")
						}
						var dto adapters.AgenticFactLedgerDTO
						if err := json.Unmarshal(input, &dto); err == nil {
							t.Fatal("v5 DTO accepted task with omitted cost_usd")
						}
						return
					}
					if err := adapters.ValidateRuntimeFactLedger(input); err != nil {
						t.Fatal(err)
					}
					var dto adapters.AgenticFactLedgerDTO
					if err := json.Unmarshal(input, &dto); err != nil {
						t.Fatal(err)
					}
					facts := dto.ToDomain()
					fact := facts.Sessions[0].Invocations[0]
					if fact.CostUSD != 1.25 || fact.CostEvidence.State != domain.CostPartial || fact.CostEvidence.KnownSubtotalUSD == nil || *fact.CostEvidence.KnownSubtotalUSD != 1.25 {
						t.Fatalf("partial accumulator lost: %+v", fact)
					}
					if len(facts.TaskLinks) != 1 || facts.TaskLinks[0].CostEvidence.State != tc.state || facts.TaskLinks[0].CostUSD != tc.amount || facts.TaskLinks[0].Description != "Keep source task" || facts.TaskLinks[0].ChildSession != "child" || facts.TaskLinks[0].ParentTurn != 8 || facts.TaskLinks[0].DurationMs != 12 {
						t.Fatalf("task conversion lost presence/metadata: %+v", facts.TaskLinks)
					}
					encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
					if err != nil {
						t.Fatal(err)
					}
					var wire struct {
						TaskLinks []struct {
							Cost     *float64            `json:"cost_usd"`
							Evidence domain.CostEvidence `json:"cost_evidence"`
						} `json:"task_links"`
						Sessions []struct {
							Invocations []struct {
								Cost *float64 `json:"cost_usd"`
							} `json:"invocations"`
						} `json:"sessions"`
					}
					if err := json.Unmarshal(encoded, &wire); err != nil {
						t.Fatal(err)
					}
					if wire.Sessions[0].Invocations[0].Cost != nil || wire.TaskLinks[0].Evidence.State != tc.state {
						t.Fatalf("partial complete amount leaked: %s", encoded)
					}
					if tc.state == domain.CostObserved {
						if wire.TaskLinks[0].Cost == nil || *wire.TaskLinks[0].Cost != 0 {
							t.Fatalf("explicit task zero lost: %s", encoded)
						}
					} else if wire.TaskLinks[0].Cost != nil {
						t.Fatalf("incomplete task became numeric: %s", encoded)
					}
				})
			}
		}
	})
	t.Run("owning canonical decoders reject malformed trailing and unknown fields", func(t *testing.T) {
		for _, version := range []int{4, 5} {
			base := strings.Replace(legacyBase, `"schema_version":4`, fmt.Sprintf(`"schema_version":%d`, version), 1)
			base = strings.Replace(base, "COST", `"cost_usd":0,"cost_evidence":{"state":"observed","known_subtotal_usd":0},`, 1)
			for _, input := range []string{
				base + `{}`, strings.TrimSuffix(base, "}"),
				strings.TrimSuffix(base, "}") + `,"unexpected":true}`,
				strings.Replace(base, `"id":"owner"`, `"id":"owner","unexpected":true`, 1),
				strings.Replace(base, `"role":"coder"`, `"role":"coder","unexpected":true`, 1),
				strings.Replace(base, `"input":1`, `"input":1,"unexpected":true`, 1),
				strings.Replace(base, `"state":"observed"`, `"state":"observed","unexpected":true`, 1),
				strings.TrimSuffix(base, "}") + `,"task_links":[{"subagent_type":"coder","cost_usd":0,"cost_evidence":{"state":"observed","known_subtotal_usd":0},"unexpected":true}]}`,
			} {
				if err := adapters.ValidateRuntimeFactLedger([]byte(input)); err == nil {
					t.Fatalf("v%d runtime accepted invalid canonical JSON: %s", version, input)
				}
				var dto adapters.AgenticFactLedgerDTO
				if err := json.Unmarshal([]byte(input), &dto); err == nil {
					t.Fatalf("v%d DTO decoder accepted invalid canonical JSON: %s", version, input)
				}
			}
		}
	})
	for _, tc := range []struct {
		name, field string
		state       domain.CostState
		amount      float64
	}{
		{"historical Codex zero", `"cost_usd":0,`, domain.CostObserved, 0},
		{"historical amount", `"cost_usd":7,`, domain.CostObserved, 7},
		{"null", `"cost_usd":null,`, domain.CostUnavailable, 0},
		{"omitted", "", domain.CostUnavailable, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte(strings.Replace(legacyBase, "COST", tc.field, 1))
			if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
				t.Fatal(err)
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(data, &dto); err != nil {
				t.Fatal(err)
			}
			if err := dto.Validate(); err != nil {
				t.Fatal(err)
			}
			inv := dto.ToDomain().Sessions[0].Invocations[0]
			if inv.CostEvidence.State != tc.state || inv.CostUSD != tc.amount {
				t.Fatalf("cost=%+v", inv)
			}
		})
	}
	for _, fields := range []string{
		`"cost_evidence":{"state":"unavailable","known_subtotal_usd":null},`,
		`"cost_usd":null,"cost_evidence":null,`,
		`"cost_usd":0,`,
		`"cost_usd":null,"cost_evidence":{"state":"observed","known_subtotal_usd":0},`,
		`"cost_usd":2,"cost_evidence":{"state":"partial","known_subtotal_usd":2},`,
		`"cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":0},`,
		`"cost_usd":-1,"cost_evidence":{"state":"observed","known_subtotal_usd":-1},`,
	} {
		t.Run("v5 rejects "+fields, func(t *testing.T) {
			data := []byte(strings.Replace(strings.Replace(legacyBase, `"schema_version":4`, `"schema_version":5`, 1), "COST", fields, 1))
			if err := adapters.ValidateRuntimeFactLedger(data); err == nil {
				t.Fatal("runtime accepted contradictory/missing evidence")
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(data, &dto); err == nil {
				t.Fatal("decoder accepted contradictory/missing evidence")
			}
		})
	}
}

// jsonFieldsInOrder retains object encounter order while exposing nested raw
// payloads for the serializer-contract assertions below.
func jsonFieldsInOrder(t *testing.T, data []byte, wanted ...string) map[string]json.RawMessage {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		t.Fatalf("expected JSON object: %s (%v)", data, err)
	}
	keys := []string{}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			t.Fatal(err)
		}
		key, ok := token.(string)
		if !ok {
			t.Fatalf("expected object key, got %v", token)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, key)
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		t.Fatal(err)
	}
	cursor := 0
	for _, key := range wanted {
		for cursor < len(keys) && keys[cursor] != key {
			cursor++
		}
		if cursor == len(keys) {
			t.Fatalf("field order %v does not retain required subsequence %v", keys, wanted)
		}
		cursor++
	}
	return fields
}

func assertFactJSONFieldOrder(t *testing.T, data []byte) {
	t.Helper()
	root := jsonFieldsInOrder(t, data, "schema_version", "ticket_id", "updated_at", "code", "sessions", "anomalies")
	jsonFieldsInOrder(t, root["code"], "lines_added", "lines_deleted", "test_lines_added", "test_lines_deleted", "prod_lines_added", "prod_lines_deleted", "files_changed", "sdlc_iterations")
	var sessions []json.RawMessage
	if err := json.Unmarshal(root["sessions"], &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) == 0 {
		t.Fatal("ordering assertion requires a session")
	}
	session := jsonFieldsInOrder(t, sessions[0], "id", "source", "user", "invocations", "tools", "anomalies")
	var invocations []json.RawMessage
	if err := json.Unmarshal(session["invocations"], &invocations); err != nil {
		t.Fatal(err)
	}
	if len(invocations) == 0 {
		t.Fatal("ordering assertion requires an invocation")
	}
	invocation := jsonFieldsInOrder(t, invocations[0], "role", "model", "started_at", "completed_at", "active_duration_seconds", "turns", "tokens", "cost_usd", "cost_evidence", "compactions", "user_iterations")
	jsonFieldsInOrder(t, invocation["tokens"], "input", "output", "cache_read", "reasoning", "peak_context")
	jsonFieldsInOrder(t, invocation["cost_evidence"], "state", "known_subtotal_usd")
}
