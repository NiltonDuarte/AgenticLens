package adapters_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

func TestAgenticFactLedgerDTOPreservesSkillAndTaskFacts(t *testing.T) {
	version := domain.CurrentSchemaVersion
	want := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceCodex, User: "tester",
		Skills:      []domain.SkillFact{{Name: "agentic-lens-dev", CallerRole: domain.KnownString("coder"), InvokedAt: domain.KnownString("2026-10-31T10:00:00Z")}},
		Delegations: []domain.DelegationFact{{AgentType: domain.KnownString("tester"), Description: "Run tests", ChildSession: "child-1", InvokedAt: domain.KnownString("2026-10-31T10:01:00Z")}},
	}}}

	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(want))
	if err != nil {
		t.Fatal(err)
	}
	var decoded adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.ToDomain()
	if !reflect.DeepEqual(got.SkillFacts(), want.SkillFacts()) || !reflect.DeepEqual(got.DelegationFacts(), want.DelegationFacts()) {
		t.Fatalf("DTO owned facts round trip got %+v want %+v", got, want)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestCurrentLedgerRejectsRetiredEventsAtRootAndSession(t *testing.T) {
	for _, field := range []string{"events", "events_are_partial", "agent_path", "task_links"} {
		for _, value := range []string{"null", "[]", "false", `""`} {
			t.Run(field+"/"+value, func(t *testing.T) {
				session := `{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[],"anomalies":[],"` + field + `":` + value + `}`
				root := `{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[],"` + field + `":` + value + `}`
				for _, input := range []string{root, `{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[` + session + `]}`} {
					if err := adapters.ValidateRuntimeFactLedger([]byte(input)); err == nil {
						t.Fatalf("retired field accepted: %s", input)
					}
				}
			})
		}
	}
}

func TestCanonicalRetiredContentFieldsReadCompatibility(t *testing.T) {
	// This is a self-contained ledger in the shape emitted before content
	// classification was removed. Reading it must not require source hydration.
	const raw = `{"schema_version":6,"ticket_id":"TEST-101","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"anomalies":[],"tools":[{"name":"command","calls":2,"errors":1,"content_errors":1}]}]}`
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
	tools := facts.ToolFacts()
	if len(facts.Sessions) != 1 || facts.Sessions[0].Source != domain.SourceCodex || len(tools) != 1 || tools[0] != (domain.ToolFact{SessionID: "owner", Name: "command", Calls: 2, Errors: 1}) {
		t.Fatalf("existing tool/source facts lost: %+v", facts)
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
	for _, existing := range []string{`"content_errors":1`} {
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
	data := []byte(`{"schema_version":6,"sessions":[],"migration":{"status":"partial","steps_pending":["source_backfill"]}}`)
	if err := adapters.ValidateRuntimeFactLedger(data); err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("expected incomplete v4 rejection, got %v", err)
	}
}

func TestValidateRuntimeFactLedgerAcceptsCanonicalV4WithoutMigrationMetadata(t *testing.T) {
	base := `{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[{"id":"s1","source":"opencode","user":"u","invocations":[{"role":"orchestrator","model":"m","started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":0,"tokens":{"input":0,"output":0,"cache_read":0,"reasoning":0,"peak_context":0},"cost_usd":0,"compactions":0,"user_iterations":0,"cost_evidence":{"state":"observed","known_subtotal_usd":0}}],"tools":[],"anomalies":[]}]}`
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
	for _, value := range []string{`"obsolete"`, `""`, `null`, `0`, `{"schema_version":6}`} {
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
	facts := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: strPtr("TEST-VALIDATION"), Sessions: []domain.SessionFact{
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
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: strPtr("TEST-VALIDATION")}
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
	base := `{"schema_version":6,"migration":{"status":"partial","steps_completed":[],"steps_pending":["source_backfill"]},"sessions":[]}`
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
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{{ID: sid, Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}, Anomalies: []domain.Anomaly{nested}}}, Anomalies: []domain.Anomaly{aggregate}}
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

func TestAgenticFactLedgerDTO_OmitsAnomalyEvidence(t *testing.T) {
	version := domain.CurrentSchemaVersion
	facts := domain.AgenticFactLedger{SchemaVersion: &version, Sessions: []domain.SessionFact{{ID: "s1", Source: domain.SourceCodex, User: "tester", Tools: []domain.SessionToolFact{{Name: "bash", Calls: 1, Errors: 1}}, Anomalies: []domain.Anomaly{{RuleID: domain.RuleToolRetryStreak, Evidence: []domain.AnomalyEvidenceRef{{SessionID: "s1", EventID: "event-1"}}}}}}}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), `"events"`) || strings.Contains(string(encoded), `"evidence"`) {
		t.Fatalf("transient payload serialized: %s", encoded)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(encoded, &dto); err != nil {
		t.Fatal(err)
	}
	if len(dto.ToDomain().Sessions[0].Anomalies) != 1 {
		t.Fatal("stored summary lost")
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
		if err := json.Unmarshal([]byte(`{"schema_version":6,"sessions":[],"anomalies":"not_an_array"}`), &dto); err == nil {
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
	t.Run("current event DTO rejected", func(t *testing.T) {
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal([]byte(`{"schema_version":6,"sessions":[{"id":"s1","events":[]}]}`), &dto); err == nil {
			t.Fatal("current retired event accepted")
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
			{"schema_version float", `{"schema_version": 7.5}`},
			{"events not an array", `{"schema_version": 6, "events": true}`},
			{"events array with invalid element", `{"schema_version": 6, "events": [123]}`},
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
		if err := json.Unmarshal([]byte(`{"schema_version":6}`), &dto); err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		dom := dto.ToDomain()
		if dom.SchemaVersion == nil || *dom.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected current schema_version to be retained")
		}
		if dom.Code.SDLCIterations != 0 {
			t.Errorf("expected zero SDLC iterations to remain zero, got %d", dom.Code.SDLCIterations)
		}
		if len(dom.InvocationFacts()) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(dom.InvocationFacts()))
		}
		if len(dom.ToolFacts()) != 0 {
			t.Errorf("expected 0 tools, got %d", len(dom.ToolFacts()))
		}
	})
}

func TestTicketIDRequiredAtLedgerBoundary(t *testing.T) {
	for _, tc := range []struct {
		name  string
		field string
		valid bool
	}{
		{"absent", "", false},
		{"null", `,"ticket_id":null`, false},
		{"empty", `,"ticket_id":""`, false},
		{"whitespace", `,"ticket_id":" \t\n"`, false},
		{"wrong type", `,"ticket_id":42`, false},
		{"recorded", `,"ticket_id":"RECORDED-1"`, true},
		{"preserved surrounding whitespace", `,"ticket_id":" RECORDED-1 "`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(`{"schema_version":6,"sessions":[]` + tc.field + `}`)
			runtimeErr := adapters.ValidateRuntimeFactLedger(input)
			if (runtimeErr == nil) != tc.valid {
				t.Fatalf("runtime valid=%t, got %v", tc.valid, runtimeErr)
			}
			if !tc.valid && tc.name != "wrong type" && !strings.Contains(runtimeErr.Error(), "ticket_id is required") {
				t.Fatalf("runtime rejected for unrelated reason: %v", runtimeErr)
			}
			var dto adapters.AgenticFactLedgerDTO
			err := json.Unmarshal(input, &dto)
			if err == nil {
				err = dto.Validate()
			}
			if (err == nil) != tc.valid {
				t.Fatalf("typed valid=%t, got %v", tc.valid, err)
			}
			if !tc.valid && tc.name != "wrong type" && !strings.Contains(err.Error(), "ticket_id is required") {
				t.Fatalf("typed validator rejected for unrelated reason: %v", err)
			}
			if tc.name == "preserved surrounding whitespace" && *dto.ToDomain().TicketID != " RECORDED-1 " {
				t.Fatal("recorded identity changed during projection")
			}
		})
	}
}

func TestToolSessionIDRequiredAtLedgerBoundary(t *testing.T) {
	version := domain.CurrentSchemaVersion
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
			dto := adapters.AgenticFactLedgerDTO{SchemaVersion: &version, TicketID: strPtr("TEST-VALIDATION"), Sessions: []adapters.SessionFactDTO{{ID: tc.sessionID, Source: "codex", User: "tester", Invocations: []adapters.InvocationFactDTO{}, Tools: tools}}}
			err := dto.Validate()
			if tc.valid && err != nil {
				t.Fatalf("expected valid session identity, got %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatal("expected missing session identity to be rejected")
			}
		})
	}
	if err := (adapters.AgenticFactLedgerDTO{SchemaVersion: &version, TicketID: strPtr("TEST-VALIDATION"), Sessions: []adapters.SessionFactDTO{}}).Validate(); err != nil {
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
			"schema_version": 6,
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
			"schema_version": 6,
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
				json: `{"schema_version": 6, "ticket_id": "AL-29", "ticket_type": 123}`,
			},
			{
				name: "ticket_type as boolean",
				json: `{"schema_version": 6, "ticket_id": "AL-29", "ticket_type": true}`,
			},
			{
				name: "ticket_type as array",
				json: `{"schema_version": 6, "ticket_id": "AL-29", "ticket_type": ["feat"]}`,
			},
			{
				name: "ticket_type as nested object",
				json: `{"schema_version": 6, "ticket_id": "AL-29", "ticket_type": {"name": "feat"}}`,
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
	t.Run("invocation partial facts survive lean delegation roundtrip", func(t *testing.T) {
		input := []byte(`{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[{"id":"owner","source":"codex","user":"alice","invocations":[{"role":"coder","model":"fixture","started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":0,"tokens":{"input":0,"output":0,"cache_read":0,"reasoning":0,"peak_context":0},"compactions":0,"user_iterations":0,"cost_usd":null,"cost_evidence":{"state":"partial","known_subtotal_usd":1.25}}],"tools":[],"anomalies":[],"delegations":[{"agent_type":"coder","invoked_at":null,"description":"Keep source task","child_session_id":"child"}]}]}`)
		if err := adapters.ValidateRuntimeFactLedger(input); err != nil {
			t.Fatal(err)
		}
		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(input, &dto); err != nil {
			t.Fatal(err)
		}
		facts := dto.ToDomain()
		inv := facts.Sessions[0].Invocations[0]
		if inv.CostEvidence.State != domain.CostPartial || inv.CostEvidence.KnownSubtotalUSD == nil || *inv.CostEvidence.KnownSubtotalUSD != 1.25 {
			t.Fatalf("partial invocation cost lost: %+v", inv)
		}
		encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(facts))
		if err != nil {
			t.Fatal(err)
		}
		var reread adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(encoded, &reread); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(facts.Sessions, reread.ToDomain().Sessions) {
			t.Fatalf("lean metadata/cost changed: %s", encoded)
		}
	})

	t.Run("owning canonical decoders reject malformed trailing and unknown fields", func(t *testing.T) {
		for _, version := range []int{6} {
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
			migrated, err := migration.NewEngine(nil).Migrate(data, 6)
			if err != nil {
				t.Fatal(err)
			}
			inv := migrated.Sessions[0].Invocations[0]

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
			data := []byte(strings.Replace(strings.Replace(legacyBase, `"schema_version":4`, `"schema_version":6,"ticket_id":"TEST-COST"`, 1), "COST", fields, 1))
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
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	order := []string{"schema_version", "ticket_id", "updated_at", "code", "sessions"}
	if _, recorded := fields["anomalies"]; recorded {
		order = append(order, "anomalies")
	}
	root := jsonFieldsInOrder(t, data, order...)
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

func TestOwnedFactsSurviveV5ToV6MigrationSaveReloadAndRejectRetiredRuntimeShapes(t *testing.T) {
	const old = `{"schema_version":5,"ticket_id":"TEST-OWNED","ticket_title":"Explicit migration","sessions":[{"id":"owner","source":"codex","user":"alice","invocations":[],"tools":[]}],"skills":[{"session_id":"owner","name":"kept","success":true,"provenance":{"message_id":"native-message"}},{"name":"discarded","success":true}],"task_links":[{"session_id":"owner","description":"kept","child_session_id":"unavailable","cost_usd":null,"cost_evidence":{"state":"unavailable","known_subtotal_usd":null}}]}`
	if err := adapters.ValidateRuntimeFactLedger([]byte(old)); err == nil || !strings.Contains(strings.ToLower(err.Error()), "migrat") {
		t.Fatalf("old runtime ingestion did not require migration: %v", err)
	}
	richCurrent := strings.Replace(old, `"schema_version":5`, `"schema_version":6`, 1)
	if err := adapters.ValidateRuntimeFactLedger([]byte(richCurrent)); err == nil {
		t.Fatal("unreleased rich6 accepted as current runtime ledger")
	}
	if _, err := migration.NewEngine(nil).Migrate([]byte(richCurrent), 6); err == nil {
		t.Fatal("unreleased rich6 silently converted through a compatibility migration")
	}
	unsupported := strings.Replace(old, `"schema_version":5`, `"schema_version":7`, 1)
	if err := adapters.ValidateRuntimeFactLedger([]byte(unsupported)); err == nil {
		t.Fatal("unsupported schema7 accepted at runtime")
	}

	migrated, err := migration.NewEngine(nil).Migrate([]byte(old), 6)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := adapters.SaveFactLedger(path, migrated); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
		t.Fatal(err)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatal(err)
	}
	got := dto.ToDomain()
	if len(got.SkillFacts()) != 1 || len(got.DelegationFacts()) != 1 || got.Sessions[0].ID != "owner" || got.SkillFacts()[0].Name != "kept" || got.DelegationFacts()[0].ChildSession != "unavailable" {
		t.Fatalf("owned migration facts lost after persistence: %+v", got)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{"skills", "task_links", "unresolved_facts"} {
		if _, ok := fields[retired]; ok {
			t.Fatalf("retired root field persisted: %s", retired)
		}
		for _, value := range []string{"null", "[]", "{}"} {
			bad := bytes.TrimSpace(data)
			bad = append(append(append([]byte{}, bad[:len(bad)-1]...), []byte(fmt.Sprintf(",%q:%s}", retired, value))...), '\n')
			if err := adapters.ValidateRuntimeFactLedger(bad); err == nil {
				t.Fatalf("retired current field %s accepted with %s", retired, value)
			}
		}
	}
}

func TestNativeNamespaceChildProofSurvivesLeanPersistenceAndRejectsAmbiguity(t *testing.T) {
	parentInput := `{"type":"session_meta","payload":{"id":"parent","source":"cli"}}` + "\n" + `{"timestamp":"2026-01-01T10:00:00Z","type":"response_item","payload":{"type":"message","id":"native-activation","role":"user","content":[{"type":"input_text","text":"$inspect-source"}]}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call","call_id":"spawn","name":"spawn_agent","arguments":"{\"agent_type\":\"tester\"}"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call_output","call_id":"spawn","output":"{\"task_name\":\"/root/tester\"}"}}`
	parentInput += "\n" + `{"type":"response_item","payload":{"type":"function_call","call_id":"unrelated-call","name":"exec_command","arguments":"{\"cmd\":\"private-event-body\"}"}}` + "\n" + `{"type":"response_item","payload":{"type":"function_call_output","call_id":"unrelated-call","output":"private-event-body"}}`
	childInput := `{"type":"session_meta","payload":{"id":"child-a","source":{"subagent":{"thread_spawn":{"parent_thread_id":"parent","agent_path":"/root/tester"}}}}}`
	parent, err := codex.ParseRollout(strings.NewReader(parentInput))
	if err != nil {
		t.Fatal(err)
	}
	child, err := codex.ParseRollout(strings.NewReader(childInput))
	if err != nil {
		t.Fatal(err)
	}
	combined, err := codex.CombineTranscripts([]codex.Transcript{parent, child})
	if err != nil {
		t.Fatal(err)
	}
	sessions, err := combined.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	version, ticket, title := 6, "TEST-PROJECTION", "Native namespace remains source evidence"
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: sessions}
	if got := ledger.DelegationFacts(); len(got) != 1 || got[0].ChildSession != "child-a" {
		t.Fatalf("unique native namespace should resolve in read projection: %+v", got)
	}
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := adapters.SaveFactLedger(path, ledger); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{"private-event-body", `"events"`, `"agent_path"`, `"child_reference"`, `"call_id"`} {
		if bytes.Contains(data, []byte(retired)) {
			t.Fatalf("retired source payload persisted: %s", data)
		}
	}
	stored, err := adapters.LoadExistingFactLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := (domain.AgenticFactLedger{Sessions: stored.Sessions}).DelegationFacts(); len(got) != 1 || got[0].ChildSession != "child-a" {
		t.Fatalf("source-proven child lost across reload: %+v", got)
	}
	ambiguous, err := codex.ParseRollout(strings.NewReader(strings.Replace(childInput, "child-a", "child-b", 1)))
	if err != nil {
		t.Fatal(err)
	}
	recombined, err := codex.CombineTranscripts([]codex.Transcript{parent, child, ambiguous})
	if err != nil {
		t.Fatal(err)
	}
	ambiguousSessions, err := recombined.SessionFacts()
	if err != nil {
		t.Fatal(err)
	}
	if got := (domain.AgenticFactLedger{Sessions: ambiguousSessions}).DelegationFacts(); len(got) != 1 || got[0].ChildSession != "" {
		t.Fatalf("ambiguous source relationship guessed: %+v", got)
	}
}

func TestCurrentLeanOccurrencesRequireExactNullableShape(t *testing.T) {
	cases := []struct {
		name, collection, row string
		valid                 bool
	}{
		{"unknown skill", "skills", `{"name":"inspect","caller_role":null,"invoked_at":null}`, true},
		{"recorded skill", "skills", `{"name":"inspect","caller_role":"custom-role","invoked_at":"2026-01-01T00:00:00Z"}`, true},
		{"missing role", "skills", `{"name":"inspect","invoked_at":null}`, false},
		{"missing time", "skills", `{"name":"inspect","caller_role":null}`, false},
		{"empty role", "skills", `{"name":"inspect","caller_role":"","invoked_at":null}`, false},
		{"empty name", "skills", `{"name":"","caller_role":null,"invoked_at":null}`, false},
		{"null name", "skills", `{"name":null,"caller_role":null,"invoked_at":null}`, false},
		{"unknown delegation", "delegations", `{"agent_type":null,"invoked_at":null}`, true},
		{"recorded delegation", "delegations", `{"agent_type":"custom-type","invoked_at":"2026-01-01T00:00:00Z","description":"Run checks","child_session_id":"unavailable"}`, true},
		{"missing type", "delegations", `{"invoked_at":null}`, false},
		{"missing delegation time", "delegations", `{"agent_type":null}`, false},
		{"empty type", "delegations", `{"agent_type":"","invoked_at":null}`, false},
		{"null description", "delegations", `{"agent_type":null,"invoked_at":null,"description":null}`, false},
		{"null child", "delegations", `{"agent_type":null,"invoked_at":null,"child_session_id":null}`, false},
		{"empty child", "delegations", `{"agent_type":null,"invoked_at":null,"child_session_id":""}`, false},
		{"invalid timestamp", "skills", `{"name":"inspect","caller_role":null,"invoked_at":"not-source-time"}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(`{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[],"anomalies":[],"` + tc.collection + `":[` + tc.row + `]}]}`)
			err := adapters.ValidateRuntimeFactLedger(input)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t err=%v input=%s", tc.valid, err, input)
			}
		})
	}
	for _, collection := range []string{"skills", "delegations"} {
		base := `{"name":"inspect","caller_role":null,"invoked_at":null}`
		if collection == "delegations" {
			base = `{"agent_type":null,"invoked_at":null}`
		}
		for _, field := range []string{"success", "outcome", "provenance", "evidence_kind", "path", "timestamp", "session_id", "source", "call_id", "child_reference", "child_agent_id", "child_link_diagnostic", "parent_turn", "duration_ms", "cost_usd", "cost_evidence", "error", "prompt"} {
			t.Run(collection+"/reject "+field, func(t *testing.T) {
				row := strings.TrimSuffix(base, "}") + `,"` + field + `":null}`
				input := []byte(`{"schema_version":6,"ticket_id":"TEST-VALIDATION","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[],"anomalies":[],"` + collection + `":[` + row + `]}]}`)
				if err := adapters.ValidateRuntimeFactLedger(input); err == nil {
					t.Fatalf("retired field accepted: %s", input)
				}
			})
		}
	}
}

func TestFullSourceAnomaliesSurviveLeanPersistenceAndPartialRefresh(t *testing.T) {
	version, ticket, title := 6, "TEST-SPARSE-EVENTS", "Sparse timeline keeps full source anomaly summaries"
	owner := domain.SessionFact{ID: "owner", Source: domain.SourceCodex, User: "tester", Tools: []domain.SessionToolFact{{Name: "bash", Calls: 5, Errors: 3}, {Name: "load_skill", Calls: 3, Errors: 3}}}
	private := "source-only-private-body"
	event := func(id, name, status string, order int) domain.EventFact {
		return domain.EventFact{SessionID: "owner", CallID: id, EventID: "source-" + id, Timestamp: fmt.Sprintf("2026-01-01T10:00:%02dZ", order), Turn: int64(order), Role: "coder", EventType: domain.EventTypeToolCall, ToolName: &name, Status: status, ToolInput: &private, ToolOutput: &private, ErrorMessage: &private}
	}
	events := []domain.EventFact{}
	for i := 0; i < 3; i++ {
		events = append(events, event(fmt.Sprintf("generic-failure-%d", i), "bash", "error", i))
	}
	for i := 0; i < 3; i++ {
		call := fmt.Sprintf("skill-failure-%d", i)
		owner.Skills = append(owner.Skills, domain.SkillFact{Name: "inspect-source"})
		events = append(events, event(call, "load_skill", "error", 3+i*2))
		if i < 2 {
			events = append(events, event(fmt.Sprintf("generic-success-%d", i), "bash", "completed", 4+i*2))
		}
	}
	ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{owner}}
	source := domain.SourceEvents{Events: events, CompleteSessionIDs: []string{"owner"}}
	expected := domain.RefreshAnomalySummariesFromSource(ledger, source)
	ledger, err := domain.PersistSourceAnomalySummaries(ledger, source)
	if err != nil {
		t.Fatal(err)
	}
	retryTools := func(anomalies []domain.Anomaly) []string {
		var tools []string
		for _, anomaly := range anomalies {
			if anomaly.RuleID == domain.RuleToolRetryStreak && anomaly.Context.ToolName != nil {
				tools = append(tools, *anomaly.Context.ToolName)
			}
		}
		return tools
	}
	if !reflect.DeepEqual(retryTools(expected), []string{"bash"}) {
		t.Fatalf("full source control must have real generic retry only, not interleaved skill retries: %+v", expected)
	}
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := adapters.SaveFactLedger(path, ledger); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(private)) {
		t.Fatal("sparse timeline persisted content-bearing source payload")
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatal(err)
	}
	stored := dto.ToDomain()
	if bytes.Contains(data, []byte(`"events"`)) || bytes.Contains(data, []byte(`"events_are_partial"`)) {
		t.Fatalf("source events persisted: %s", data)
	}
	actual := domain.RefreshAnomalySummariesFromSource(stored, domain.SourceEvents{Events: []domain.EventFact{event("partial-one", "load_skill", "error", 1), event("partial-two", "load_skill", "error", 2), event("partial-three", "load_skill", "error", 3)}})
	if !reflect.DeepEqual(retryTools(actual), []string{"bash"}) || !reflect.DeepEqual(actual, expected) {
		t.Fatalf("sparse event recomputation discarded cached full-source anomalies or invented skill streak: actual=%+v expected=%+v", actual, expected)
	}
}

func TestDurationEvidenceRoundTripLegacyAndValidation(t *testing.T) {
	version := domain.CurrentSchemaVersion
	ticket, title := "FIX-DURATION", "Portable numeric duration"
	want := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceCodex, User: "alice", Invocations: []domain.InvocationFact{{Role: "coder", Model: "m", ActiveDurationSeconds: 10, ProvisionalSeconds: 7, Tokens: domain.InvocationTokens{Input: 31}}}}}}
	encoded, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(want))
	if err != nil {
		t.Fatal(err)
	}
	draft := `"duration_evidence":{"state":"provisional","completed_seconds":3,"provisional_seconds":7,"cutoff":"2026-01-01T00:00:10Z","basis":"codex:turn"}`
	wire := func(member string) []byte {
		return []byte(strings.Replace(string(encoded), `"provisional_seconds":7`, member, 1))
	}
	for _, tc := range []struct {
		name, member string
		portion      int64
	}{
		{"original_v6", "", 0},
		{"explicit_zero", `"provisional_seconds":0`, 0},
		{"lean", `"provisional_seconds":7`, 7},
		{"draft", draft, 7},
		{"equal_dual", `"provisional_seconds":7,` + draft, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := wire(tc.member)
			if tc.member == "" {
				raw = bytes.Replace(raw, []byte(",,"), []byte(","), 1)
				raw = bytes.Replace(raw, []byte("{,"), []byte("{"), 1)
			}
			if err := adapters.ValidateRuntimeFactLedger(raw); err != nil {
				t.Fatal(err)
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(raw, &dto); err != nil {
				t.Fatal(err)
			}
			ledger := dto.ToDomain()
			inv := ledger.Sessions[0].Invocations[0]
			if inv.ActiveDurationSeconds != 10 || inv.ProvisionalSeconds != tc.portion || inv.Tokens.Input != 31 {
				t.Fatalf("numeric/owner fact loss: %+v", ledger)
			}
			migrated, err := migration.NewEngine(nil).Migrate(raw, version)
			if err != nil {
				t.Fatal(err)
			}
			if migrated.Sessions[0].ID != ledger.Sessions[0].ID || migrated.Sessions[0].Source != ledger.Sessions[0].Source || migrated.Sessions[0].User != ledger.Sessions[0].User || !reflect.DeepEqual(migrated.Sessions[0].Invocations, ledger.Sessions[0].Invocations) {
				t.Fatalf("migration changed normalized owner: %+v", migrated)
			}
			canonical, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(migrated))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(canonical, []byte(`"duration_evidence"`)) || bytes.Contains(canonical, []byte(`"duration_reasoning"`)) || bytes.Contains(canonical, []byte(`"completed_seconds"`)) {
				t.Fatalf("canonical output kept retired metadata: %s", canonical)
			}
			if bytes.Contains(canonical, []byte(`"provisional_seconds"`)) != (tc.portion > 0) {
				t.Fatalf("nonpositive provisional not omitted: %s", canonical)
			}
			var reloaded adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(canonical, &reloaded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(reloaded.ToDomain().Sessions[0].Invocations, ledger.Sessions[0].Invocations) {
				t.Fatal("save/reload changed numeric facts or identity")
			}
		})
	}
	invalid := map[string]string{
		"negative": `"provisional_seconds":-1`, "fractional": `"provisional_seconds":1.5`, "null": `"provisional_seconds":null`, "string": `"provisional_seconds":"7"`,
		"over_total": `"provisional_seconds":11`, "unsafe_integer": `"provisional_seconds":9007199254740992`,
		"dual_conflict": `"provisional_seconds":6,` + draft, "dual_zero_conflict": `"provisional_seconds":0,` + draft,
		"flat_key_alias":                `"Provisional_Seconds":7`,
		"draft_key_alias":               strings.Replace(draft, `"duration_evidence"`, `"Duration_Evidence"`, 1),
		"total_key_alias":               `"Active_Duration_Seconds":10`,
		"alias_conflicting_legacy_zero": `"Provisional_Seconds":7,"duration_evidence":{"state":"complete","completed_seconds":10,"provisional_seconds":0,"basis":"codex:turn"}`,
	}
	for name, replacement := range map[string][2]string{
		"legacy_negative":                  {`"provisional_seconds":7`, `"provisional_seconds":-7`},
		"legacy_sum":                       {`"completed_seconds":3`, `"completed_seconds":4`},
		"legacy_bad_cutoff":                {`2026-01-01T00:00:10Z`, `not-a-time`},
		"legacy_missing_cutoff":            {`"cutoff":"2026-01-01T00:00:10Z",`, ``},
		"legacy_complete_with_provisional": {`"state":"provisional"`, `"state":"complete"`},
		"legacy_unknown_state":             {`"state":"provisional"`, `"state":"invented"`},
		"legacy_missing_basis":             {`"basis":"codex:turn"`, `"basis":""`},
		"legacy_blank_basis":               {`"basis":"codex:turn"`, `"basis":"  "`},
		"legacy_unknown_member":            {`"basis":"codex:turn"`, `"basis":"codex:turn","invented":true`},
	} {
		invalid[name] = strings.Replace(draft, replacement[0], replacement[1], 1)
	}
	for _, total := range []string{"-1", "null", "1.5", "9007199254740992"} {
		t.Run("invalid_total_"+total, func(t *testing.T) {
			raw := bytes.Replace(encoded, []byte(`"active_duration_seconds":10`), []byte(`"active_duration_seconds":`+total), 1)
			if err := adapters.ValidateRuntimeFactLedger(raw); err == nil {
				t.Fatalf("invalid total accepted: %s", raw)
			}
			if _, err := migration.NewEngine(nil).Migrate(raw, version); err == nil {
				t.Fatalf("migration accepted invalid total %s", total)
			}
		})
	}
	for name, member := range invalid {
		t.Run(name, func(t *testing.T) {
			raw := wire(member)
			if err := adapters.ValidateRuntimeFactLedger(raw); err == nil {
				t.Fatalf("invalid duration accepted: %s", raw)
			}
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(raw, &dto); err == nil {
				t.Fatal("DTO accepted invalid duration")
			}
			if _, err := migration.NewEngine(nil).Migrate(raw, version); err == nil {
				t.Fatal("migration accepted invalid duration")
			}
		})
	}
}

func TestRetainedDraftDurationNormalizesWithoutMutatingBaseline(t *testing.T) {
	raw := []byte(`{"schema_version":6,"ticket_id":"FIX-HISTORY","ticket_title":"Retained duration","sessions":[{"id":"remote","source":"codex","user":"alice","invocations":[{"role":"coder","model":"m","started_at":null,"completed_at":null,"active_duration_seconds":10,"duration_evidence":{"state":"provisional","completed_seconds":3,"provisional_seconds":7,"cutoff":"2026-01-01T00:00:10Z","basis":"codex:turn"},"turns":2,"tokens":{"input":31,"output":2,"cache_read":0,"reasoning":0,"peak_context":31},"cost_usd":null,"cost_evidence":{"state":"unavailable"},"compactions":0,"user_iterations":1}],"tools":[],"anomalies":[]}]}`)
	existing, err := adapters.DecodeExistingFactLedger(raw)
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(existing.PersistedSessions)
	if err != nil {
		t.Fatal(err)
	}
	normalized, err := adapters.NormalizeHistorySessionForPublication(existing.PersistedSessions[0], domain.CurrentSchemaVersion)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(existing.PersistedSessions)
	if !bytes.Equal(before, after) {
		t.Fatal("history normalization mutated retained baseline")
	}
	if !reflect.DeepEqual(normalized.ToDomain(), existing.Sessions[0]) {
		t.Fatal("history normalization changed owned facts")
	}
	ledger := domain.AgenticFactLedger{SchemaVersion: existing.SchemaVersion, TicketID: existing.TicketID, TicketTitle: existing.TicketTitle, Sessions: existing.Sessions}
	path := filepath.Join(t.TempDir(), "copied", "agentic_metrics.json")
	if err := adapters.SaveFactLedgerPreservingHistory(path, ledger, existing.PersistedSessions); err != nil {
		t.Fatal(err)
	}
	canonical, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(canonical, []byte(`"duration_evidence"`)) || !bytes.Contains(canonical, []byte(`"provisional_seconds": 7`)) {
		t.Fatalf("retained draft not canonicalized: %s", canonical)
	}
	reloaded, err := adapters.DecodeExistingFactLedger(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reloaded.Sessions, existing.Sessions) {
		t.Fatalf("history save/reload lost numeric or identity facts: %+v", reloaded.Sessions)
	}
	after, _ = json.Marshal(existing.PersistedSessions)
	if !bytes.Equal(before, after) {
		t.Fatal("history publication mutated baseline")
	}
}

func TestCurrentLedgerGenerationPolicyReadBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, policy string
		valid        bool
		wantCause    string
	}{
		{"historical_absent", "", true, ""},
		{"all_enabled", `,"generation_policy":{"skills":true,"delegations":true,"anomalies":true}`, true, ""},
		{"all_disabled", `,"generation_policy":{"skills":false,"delegations":false,"anomalies":false}`, true, ""},
		{"null", `,"generation_policy":null`, false, "generation_policy requires exactly"},
		{"partial", `,"generation_policy":{"skills":false}`, false, "generation_policy requires exactly"},
		{"wrong_type", `,"generation_policy":{"skills":false,"delegations":"false","anomalies":false}`, false, "generation_policy.delegations must be a boolean"},
		{"retired_spelling", `,"generation_policy":{"skills":false,"task_links":false,"anomalies":false}`, false, "generation_policy.delegations must be a boolean"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"schema_version":6,"ticket_id":"TEST-POLICY","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[]}]` + tc.policy + `}`
			if tc.name != "all_disabled" {
				raw = strings.Replace(raw, `"tools":[]`, `"tools":[],"anomalies":[]`, 1)
			}
			existing, err := adapters.DecodeExistingFactLedger([]byte(raw))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t, decode error=%v", tc.valid, err)
			}
			if !tc.valid && !strings.Contains(err.Error(), tc.wantCause) {
				t.Fatalf("rejected for unrelated cause; want %q, got %v", tc.wantCause, err)
			}
			if tc.valid && existing == nil {
				t.Fatal("valid portable ledger unavailable")
			}
			if tc.valid {
				if existing.TicketID == nil || *existing.TicketID != "TEST-POLICY" || len(existing.Sessions) != 1 || existing.Sessions[0].ID != "owner" {
					t.Fatalf("compatible metadata changed recorded identity: %+v", existing)
				}
				if (existing.Sessions[0].Anomalies == nil) != (tc.name == "all_disabled") {
					t.Fatal("compatible policy changed omitted versus recorded-empty anomalies")
				}
				var dto adapters.AgenticFactLedgerDTO
				if err := json.Unmarshal([]byte(raw), &dto); err != nil {
					t.Fatal(err)
				}
				rewritten, err := json.Marshal(dto)
				if err != nil {
					t.Fatal(err)
				}
				if bytes.Contains(rewritten, []byte(`"generation_policy"`)) || bytes.Contains(rewritten, []byte(`"anomaly_availability"`)) {
					t.Fatalf("read-compatible metadata re-emitted: %s", rewritten)
				}
				if err := adapters.ValidateRuntimeFactLedger(rewritten); err != nil {
					t.Fatalf("metadata-free output rejected: %v", err)
				}
			}
		})
	}
}

func TestCurrentLedgerUnavailableAnomalyBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		valid        bool
		wantCause    string
	}{
		{"omitted_unavailable", `,"anomaly_availability":"unavailable"`, true, ""},
		{"unknown_marker", `,"anomaly_availability":"missing"`, false, "anomaly_availability must be unavailable"},
		{"null_marker", `,"anomaly_availability":null`, false, "anomaly_availability must be unavailable"},
		{"boolean_marker", `,"anomaly_availability":false`, false, "anomaly_availability must be unavailable"},
		{"contradictory_empty", `,"anomaly_availability":"unavailable","anomalies":[]`, false, "unavailable analysis requires omitted anomalies"},
		{"contradictory_null", `,"anomaly_availability":"unavailable","anomalies":null`, false, "anomalies must be an array"},
		{"unmarked_null", `,"anomalies":null`, false, "anomalies must be an array"},
		{"unmarked_measured_empty", `,"anomalies":[]`, true, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := `{"schema_version":6,"ticket_id":"TEST-ANOMALY","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[]` + tc.fields + `}]}`
			existing, err := adapters.DecodeExistingFactLedger([]byte(raw))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t error=%v", tc.valid, err)
			}
			if !tc.valid && !strings.Contains(err.Error(), tc.wantCause) {
				t.Fatalf("rejected for unrelated cause; want %q, got %v", tc.wantCause, err)
			}
			if tc.valid && existing == nil {
				t.Fatal("valid portable ledger unavailable")
			}
			if tc.valid {
				if existing.TicketID == nil || *existing.TicketID != "TEST-ANOMALY" || len(existing.Sessions) != 1 || existing.Sessions[0].ID != "owner" {
					t.Fatalf("compatible metadata changed recorded identity: %+v", existing)
				}
				if (existing.Sessions[0].Anomalies == nil) != (tc.name == "omitted_unavailable") {
					t.Fatal("compatibility read changed unavailable versus measured-empty anomalies")
				}
				var dto adapters.AgenticFactLedgerDTO
				if err := json.Unmarshal([]byte(raw), &dto); err != nil {
					t.Fatal(err)
				}
				rewritten, err := json.Marshal(dto)
				if err != nil {
					t.Fatal(err)
				}
				if bytes.Contains(rewritten, []byte(`"generation_policy"`)) || bytes.Contains(rewritten, []byte(`"anomaly_availability"`)) {
					t.Fatalf("read-compatible metadata re-emitted: %s", rewritten)
				}
				if err := adapters.ValidateRuntimeFactLedger(rewritten); err != nil {
					t.Fatalf("metadata-free output rejected: %v", err)
				}
			}
		})
	}
}

func TestOptionalCollectionPresenceSurvivesCanonicalRoundTrip(t *testing.T) {
	for _, state := range []string{"not recorded", "recorded empty", "recorded values"} {
		t.Run(state, func(t *testing.T) {
			version := domain.CurrentSchemaVersion
			owner := domain.SessionFact{ID: "owner", Source: domain.SourceCodex, User: "tester", Invocations: []domain.InvocationFact{}, Tools: []domain.SessionToolFact{}}
			ticket, title := "TEST-PRESENCE", "Optional collection presence"
			kind := domain.TicketTypeFeat
			ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &title, TicketType: &kind, UpdatedAt: "2026-01-01T00:00:00Z", Sessions: []domain.SessionFact{owner}}
			if state != "not recorded" {
				ledger.Sessions[0].Skills = []domain.SkillFact{}
				ledger.Sessions[0].Delegations = []domain.DelegationFact{}
				ledger.Sessions[0].Anomalies = []domain.Anomaly{}
				ledger.Anomalies = []domain.Anomaly{}
			}
			if state == "recorded values" {
				ledger.Sessions[0].Skills = []domain.SkillFact{{Name: "inspect"}, {Name: "inspect"}}
				ledger.Sessions[0].Delegations = []domain.DelegationFact{{Description: "review"}, {Description: "review"}}
				count := int64(3)
				tool := "saved-only"
				anomaly := domain.Anomaly{RuleID: domain.RuleToolRetryStreak, Severity: domain.SeverityWarning, Context: domain.AnomalyContext{ToolName: &tool, Count: &count}}
				ledger.Sessions[0].Anomalies = []domain.Anomaly{anomaly}
				ledger.Anomalies = []domain.Anomaly{anomaly}
			}
			path := filepath.Join(t.TempDir(), "ledger.json")
			if err := adapters.SaveFactLedger(path, ledger); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := adapters.ValidateRuntimeFactLedger(raw); err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Anomalies json.RawMessage              `json:"anomalies"`
				Sessions  []map[string]json.RawMessage `json:"sessions"`
			}
			if err := json.Unmarshal(raw, &wire); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), `"generation_policy"`) || strings.Contains(string(raw), `"anomaly_availability"`) {
				t.Fatalf("obsolete metadata emitted: %s", raw)
			}
			for name, value := range map[string]json.RawMessage{"root anomalies": wire.Anomalies, "skills": wire.Sessions[0]["skills"], "delegations": wire.Sessions[0]["delegations"], "anomalies": wire.Sessions[0]["anomalies"]} {
				switch state {
				case "not recorded":
					if len(value) != 0 {
						t.Fatalf("%s fabricated recorded state: %s", name, value)
					}
				case "recorded empty":
					if string(value) != "[]" {
						t.Fatalf("%s lost recorded-empty state: %s", name, value)
					}
				case "recorded values":
					if len(value) == 0 || string(value) == "[]" || string(value) == "null" {
						t.Fatalf("%s lost recorded values: %s", name, value)
					}
				}
			}
			loaded, err := adapters.LoadExistingFactLedger(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(loaded.Sessions[0].Skills, ledger.Sessions[0].Skills) || !reflect.DeepEqual(loaded.Sessions[0].Delegations, ledger.Sessions[0].Delegations) || !reflect.DeepEqual(loaded.Sessions[0].Anomalies, ledger.Sessions[0].Anomalies) || !reflect.DeepEqual(loaded.Anomalies, ledger.Anomalies) {
				t.Fatalf("canonical reload changed collection presence/content: %+v", loaded)
			}
		})
	}
}

func TestOptionalCollectionsRejectNullWithoutChangingAbsenceMeaning(t *testing.T) {
	for _, category := range []string{"skills", "delegations", "anomalies"} {
		raw := `{"schema_version":6,"ticket_id":"TEST-PRESENCE","sessions":[{"id":"owner","source":"codex","user":"tester","invocations":[],"tools":[],"` + category + `":null}]}`
		if err := adapters.ValidateRuntimeFactLedger([]byte(raw)); err == nil {
			t.Fatalf("null %s accepted as absent/empty", category)
		}
	}
	raw := `{"schema_version":6,"ticket_id":"TEST-PRESENCE","sessions":[],"anomalies":null}`
	if err := adapters.ValidateRuntimeFactLedger([]byte(raw)); err == nil {
		t.Fatal("null root analysis accepted as absent/empty")
	}
}
