package migration

import (
	"agenticlens/internal/domain"
	"encoding/json"
)

type legacySkillFact struct {
	removedFields      int
	EvidenceReferences []domain.FactProvenance `json:"evidence_references,omitempty"`
	SessionID          string                  `json:"session_id,omitempty"`
	Source             domain.SourceHarness    `json:"source,omitempty"`
	Provenance         domain.FactProvenance   `json:"provenance,omitzero"`
	EvidenceKind       string                  `json:"evidence_kind,omitempty"`
	Outcome            domain.FactOutcome      `json:"outcome,omitempty"`
	Path               string                  `json:"path,omitempty"`
	Name               string                  `json:"name"`
	TurnIndex          int                     `json:"turn_index"`
	MessageIndex       int                     `json:"message_index"`
	CallerRole         string                  `json:"caller_role"`
	Timestamp          string                  `json:"timestamp"`
	Success            bool                    `json:"success"`
	Error              string                  `json:"error,omitempty"`
}

// legacyTaskLinkFact records a parent delegation occurrence and its explicit child reference.
type legacyTaskLinkFact struct {
	removedFields       int
	ChildAgentID        string                `json:"child_agent_id,omitempty"`
	ChildLinkDiagnostic string                `json:"child_link_diagnostic,omitempty"`
	ChildReference      string                `json:"child_reference,omitempty"`
	SessionID           string                `json:"session_id,omitempty"`
	Source              domain.SourceHarness  `json:"source,omitempty"`
	Provenance          domain.FactProvenance `json:"provenance,omitzero"`
	Outcome             domain.FactOutcome    `json:"outcome,omitempty"`
	Error               string                `json:"error,omitempty"`
	SubagentType        string                `json:"subagent_type"`
	Description         string                `json:"description"`
	ParentTurn          int                   `json:"parent_turn"`
	ChildSession        string                `json:"child_session_id,omitempty"`
	Timestamp           string                `json:"timestamp"`
	DurationMs          int64                 `json:"duration_ms,omitempty"`
	CostUSD             float64               `json:"cost_usd"`
	CostEvidence        domain.CostEvidence   `json:"cost_evidence"`
}

type legacySessionEventFact struct {
	CallID           string           `json:"call_id,omitempty"`
	EventID          string           `json:"event_id"`
	Timestamp        string           `json:"timestamp"`
	MessageID        *string          `json:"message_id,omitempty"`
	Turn             int64            `json:"turn"`
	Role             domain.AgentRole `json:"role"`
	EventType        domain.EventType `json:"event_type"`
	ToolName         *string          `json:"tool_name,omitempty"`
	Status           string           `json:"status"`
	ErrorMessage     *string          `json:"error_message,omitempty"`
	DurationMs       *int64           `json:"duration_ms,omitempty"`
	ContextTokens    int64            `json:"context_tokens"`
	PromptContent    *string          `json:"prompt_content,omitempty"`
	ReasoningContent *string          `json:"reasoning_content,omitempty"`
	ToolInput        *string          `json:"tool_input,omitempty"`
	ToolOutput       *string          `json:"tool_output,omitempty"`
}

// legacySessionFact owns all raw telemetry attributed to a single agent session.
type legacySessionFact struct {
	partialMarkerPresent bool
	agentPathPresent     bool
	EventsArePartial     bool                     `json:"events_are_partial,omitempty"`
	AgentPath            string                   `json:"agent_path,omitempty"`
	Skills               []legacySkillFact        `json:"skills,omitempty"`
	TaskLinks            []legacyTaskLinkFact     `json:"task_links,omitempty"`
	ID                   string                   `json:"id"`
	ParentID             *string                  `json:"parent_id,omitempty"`
	Source               domain.SourceHarness     `json:"source,omitempty"`
	User                 string                   `json:"user,omitempty"`
	Invocations          []domain.InvocationFact  `json:"invocations"`
	Tools                []domain.SessionToolFact `json:"tools"`
	Events               []legacySessionEventFact `json:"events,omitempty"`
	Anomalies            []domain.Anomaly         `json:"anomalies"`
}

type legacyLedger struct {
	SchemaVersion *int                           `json:"schema_version,omitempty"`
	Migration     *domain.MigrationInfo          `json:"migration,omitempty"`
	TicketID      *string                        `json:"ticket_id"`
	TicketTitle   *string                        `json:"ticket_title"`
	TicketType    *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt     *string                        `json:"started_at"`
	CompletedAt   *string                        `json:"completed_at"`
	UpdatedAt     string                         `json:"updated_at"`
	Code          domain.CodeFacts               `json:"code"`
	Sessions      []legacySessionFact            `json:"sessions"`
	Anomalies     []domain.Anomaly               `json:"anomalies"`
	Skills        []legacySkillFact              `json:"-"`
	TaskLinks     []legacyTaskLinkFact           `json:"-"`
}

func (ledger legacyLedger) current() domain.AgenticFactLedger {
	result := domain.AgenticFactLedger{SchemaVersion: ledger.SchemaVersion, Migration: ledger.Migration, TicketID: ledger.TicketID, TicketTitle: ledger.TicketTitle, TicketType: ledger.TicketType, StartedAt: ledger.StartedAt, CompletedAt: ledger.CompletedAt, UpdatedAt: ledger.UpdatedAt, Code: ledger.Code, Anomalies: ledger.Anomalies, Sessions: []domain.SessionFact{}}
	for _, session := range ledger.Sessions {
		owner := domain.SessionFact{ID: session.ID, ParentID: session.ParentID, Source: session.Source, User: session.User, Invocations: session.Invocations, Tools: session.Tools, Anomalies: session.Anomalies}
		for _, skill := range session.Skills {
			owner.Skills = append(owner.Skills, domain.SkillFact{Name: skill.Name, CallerRole: domain.KnownString(skill.CallerRole), InvokedAt: domain.KnownString(skill.Timestamp)})
		}
		for _, link := range session.TaskLinks {
			owner.Delegations = append(owner.Delegations, domain.DelegationFact{AgentType: domain.KnownString(link.SubagentType), Description: link.Description, InvokedAt: domain.KnownString(link.Timestamp), ChildSession: link.ChildSession})
		}
		result.Sessions = append(result.Sessions, owner)
	}
	return result
}
func (ledger legacyLedger) InvocationFacts() []domain.InvocationFact {
	return ledger.current().InvocationFacts()
}
func (ledger legacyLedger) ToolFacts() []domain.ToolFact   { return ledger.current().ToolFacts() }
func (ledger legacyLedger) AnomalyFacts() []domain.Anomaly { return ledger.current().AnomalyFacts() }
func (ledger legacyLedger) EventFacts() []domain.EventFact {
	result := []domain.EventFact{}
	for _, session := range ledger.Sessions {
		for _, event := range session.Events {
			result = append(result, domain.EventFact{EventID: event.EventID, CallID: event.CallID, Timestamp: event.Timestamp, SessionID: session.ID, MessageID: event.MessageID, Turn: event.Turn, Role: event.Role, EventType: event.EventType, ToolName: event.ToolName, Status: event.Status, ErrorMessage: event.ErrorMessage, DurationMs: event.DurationMs, ContextTokens: event.ContextTokens})
		}
	}
	return result
}
func legacySessionsFromFacts(owned []legacySessionFact, tools []domain.ToolFact, events []domain.EventFact) ([]legacySessionFact, error) {
	base := legacyLedger{Sessions: owned}
	sessions, err := domain.SessionsFromFacts(base.current().Sessions, tools, events)
	if err != nil {
		return nil, err
	}
	result := make([]legacySessionFact, len(sessions))
	byID := make(map[string]legacySessionFact)
	for _, owner := range owned {
		byID[owner.ID] = owner
	}
	for i, owner := range sessions {
		previous := byID[owner.ID]
		previous.ID = owner.ID
		previous.ParentID = owner.ParentID
		previous.Source = owner.Source
		previous.User = owner.User
		previous.Invocations = owner.Invocations
		previous.Tools = owner.Tools
		previous.Events = nil
		result[i] = previous
	}
	for i := range result {
		for _, event := range events {
			if event.SessionID == result[i].ID {
				result[i].Events = append(result[i].Events, legacySessionEventFact{EventID: event.EventID, CallID: event.CallID, Timestamp: event.Timestamp, MessageID: event.MessageID, Turn: event.Turn, Role: event.Role, EventType: event.EventType, ToolName: event.ToolName, Status: event.Status, ErrorMessage: event.ErrorMessage, DurationMs: event.DurationMs, ContextTokens: event.ContextTokens})
			}
		}
	}
	return result, nil
}
func legacyPersistAnomalySummaries(ledger legacyLedger, thresholds ...domain.AnomalyThresholdsConfig) (legacyLedger, error) {
	// Only explicitly complete old event collections may replace a cached owner summary.
	source := domain.SourceEvents{Events: ledger.EventFacts()}
	for _, owner := range ledger.Sessions {
		if owner.Events != nil && !owner.EventsArePartial {
			source.CompleteSessionIDs = append(source.CompleteSessionIDs, owner.ID)
		}
	}
	policy := historicalV2AnomalyPolicy()
	if len(thresholds) > 0 {
		policy = thresholds[0]
	}
	current, err := persistHistoricalAnomalySummaries(ledger.current(), source, policy)
	if err != nil {
		return legacyLedger{}, err
	}
	result := ledger
	result.Sessions = append([]legacySessionFact{}, ledger.Sessions...)
	for i := range result.Sessions {
		result.Sessions[i].Anomalies = current.Sessions[i].Anomalies
		result.Sessions[i].Events = append([]legacySessionEventFact(nil), result.Sessions[i].Events...)
		for j := range result.Sessions[i].Events {
			event := &result.Sessions[i].Events[j]
			event.ErrorMessage = nil
			event.PromptContent = nil
			event.ReasoningContent = nil
			event.ToolInput = nil
			event.ToolOutput = nil
		}
	}
	result.Anomalies = current.Anomalies
	return result, nil
}

// Input aliases expose only the explicit migration boundary, never runtime facts.
type LedgerInput = legacyLedger
type SessionInput = legacySessionFact
type SkillInput = legacySkillFact
type DelegationInput = legacyTaskLinkFact
type EventInput = legacySessionEventFact

func countRetiredFields(data []byte, kept []string) (int, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return 0, err
	}
	count := 0
	for key := range fields {
		keep := false
		for _, name := range kept {
			if key == name {
				keep = true
			}
		}
		if !keep {
			count++
		}
	}
	return count, nil
}
func (value *legacySkillFact) UnmarshalJSON(data []byte) error {
	type alias legacySkillFact
	var input alias
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	count, err := countRetiredFields(data, []string{"name", "caller_role", "timestamp"})
	if err != nil {
		return err
	}
	*value = legacySkillFact(input)
	value.removedFields = count
	return nil
}
func (value *legacySessionFact) UnmarshalJSON(data []byte) error {
	type alias legacySessionFact
	var input alias
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*value = legacySessionFact(input)
	_, value.partialMarkerPresent = fields["events_are_partial"]
	_, value.agentPathPresent = fields["agent_path"]
	return nil
}

// persistHistoricalAnomalySummaries freezes the registered historical sanitizer's
// owner/presence rules. Runtime validation is intentionally not a migration input:
// this transform predates failure-rate configuration and never invents that rule.
func persistHistoricalAnomalySummaries(ledger domain.AgenticFactLedger, source domain.SourceEvents, policy domain.AnomalyThresholdsConfig) (domain.AgenticFactLedger, error) {
	if len(source.CompleteSessionIDs) == 0 {
		return ledger, nil
	}
	anomalies := domain.RefreshAnomalySummariesFromSource(ledger, source, policy)
	rateAvailable := domain.HasConfiguredFailureRatePolicy(policy)
	bySession := map[string][]domain.Anomaly{}
	ticket := []domain.Anomaly{}
	for _, a := range anomalies {
		a.Evidence = nil
		if a.Context.SessionID == nil || *a.Context.SessionID == "" {
			ticket = append(ticket, a)
			continue
		}
		id := *a.Context.SessionID
		a.Context.SessionID = nil
		bySession[id] = append(bySession[id], a)
	}
	complete := map[string]bool{}
	for _, id := range source.CompleteSessionIDs {
		complete[id] = true
	}
	ledger.Sessions = append([]domain.SessionFact{}, ledger.Sessions...)
	for i, owner := range ledger.Sessions {
		if !complete[owner.ID] {
			continue
		}
		incompleteGroup := false
		for _, child := range ledger.Sessions {
			if child.ParentID != nil && *child.ParentID == owner.ID && !complete[child.ID] {
				incompleteGroup = true
			}
		}
		if incompleteGroup && owner.Anomalies == nil && len(bySession[owner.ID]) == 0 {
			continue
		}
		if !rateAvailable && len(bySession[owner.ID]) == 0 && owner.Anomalies == nil {
			continue
		}
		ledger.Sessions[i].Anomalies = append([]domain.Anomaly{}, bySession[owner.ID]...)
	}
	allComplete := len(ledger.Sessions) > 0 && !ledger.HasExternalParentReference()
	for _, owner := range ledger.Sessions {
		if !complete[owner.ID] {
			allComplete = false
		}
		if owner.ParentID != nil && *owner.ParentID != "" && !complete[*owner.ParentID] {
			allComplete = false
		}
	}
	if allComplete && (rateAvailable || len(ticket) > 0 || ledger.Anomalies != nil) {
		ledger.Anomalies = ticket
	}
	return ledger, nil
}
