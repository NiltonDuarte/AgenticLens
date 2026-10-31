package migration

import (
	"encoding/json"
	"fmt"

	"agenticlens/internal/domain"
)

// Metadata is identical across supported schema families. Cost-bearing
// collections are decoded separately so nullable amounts never become zero
// before their evidence has been validated.
type ledgerMetadataInput struct {
	SchemaVersion *int                           `json:"schema_version"`
	Migration     *domain.MigrationInfo          `json:"migration"`
	TicketID      *string                        `json:"ticket_id"`
	TicketTitle   *string                        `json:"ticket_title"`
	TicketType    *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt     *string                        `json:"started_at"`
	CompletedAt   *string                        `json:"completed_at"`
	UpdatedAt     string                         `json:"updated_at"`
	Code          domain.CodeFacts               `json:"code"`
	Anomalies     []domain.Anomaly               `json:"anomalies"`
	Skills        []domain.SkillFact             `json:"skills"`
}

type schemaV1ToV4LedgerInput struct {
	ledgerMetadataInput
	ledgerCollectionsInput
}

// Schema v5 retains the collection layout but requires cost members and
// evidence. A distinct decoder selects that policy before domain assembly.
type schemaV5LedgerInput struct {
	ledgerMetadataInput
	ledgerCollectionsInput
}

type ledgerCollectionsInput struct {
	Sessions    []sessionInput          `json:"sessions"`
	TaskLinks   []taskLinkInput         `json:"task_links"`
	Source      domain.SourceHarness    `json:"source"`
	Invocations []legacyInvocationInput `json:"invocations"`
	Tools       []domain.ToolFact       `json:"tools"`
	Events      []domain.EventFact      `json:"events"`
}

// memberPresence records a named schema member, including an explicit null.
// Its value is decoded by the corresponding concrete input struct.
type memberPresence bool

func (present *memberPresence) UnmarshalJSON(_ []byte) error {
	*present = true
	return nil
}

type sessionInput struct {
	ID          string                    `json:"id"`
	ParentID    *string                   `json:"parent_id"`
	Source      domain.SourceHarness      `json:"source"`
	User        string                    `json:"user"`
	Invocations []invocationInput         `json:"invocations"`
	Tools       []domain.SessionToolFact  `json:"tools"`
	Events      []domain.SessionEventFact `json:"events"`
	Anomalies   []domain.Anomaly          `json:"anomalies"`
}

type invocationInput struct {
	Role                  domain.AgentRole        `json:"role"`
	Model                 string                  `json:"model"`
	StartedAt             *string                 `json:"started_at"`
	CompletedAt           *string                 `json:"completed_at"`
	ActiveDurationSeconds int64                   `json:"active_duration_seconds"`
	Turns                 int64                   `json:"turns"`
	Tokens                domain.InvocationTokens `json:"tokens"`
	CostUSD               json.RawMessage         `json:"cost_usd"`
	CostEvidence          *domain.CostEvidence    `json:"cost_evidence"`
	Compactions           int64                   `json:"compactions"`
	UserIterations        int64                   `json:"user_iterations"`
}

// Legacy flat ledgers stored session metadata alongside each invocation row.
// Move that metadata to its owner without interpreting synthetic ID suffixes.
type legacyInvocationInput struct {
	invocationInput
	SessionID string  `json:"id"`
	ParentID  *string `json:"parent_id"`
	User      string  `json:"user"`
}

type taskLinkInput struct {
	SubagentType string               `json:"subagent_type"`
	Description  string               `json:"description"`
	ParentTurn   int                  `json:"parent_turn"`
	ChildSession string               `json:"child_session_id"`
	Timestamp    string               `json:"timestamp"`
	DurationMs   int64                `json:"duration_ms"`
	CostUSD      json.RawMessage      `json:"cost_usd"`
	CostEvidence *domain.CostEvidence `json:"cost_evidence"`
}

type legacyFactCollections struct {
	Source      domain.SourceHarness
	Invocations []legacySessionInvocation
	Tools       []domain.ToolFact
	Events      []domain.EventFact
}

type legacySessionInvocation struct {
	domain.InvocationFact
	SessionID string
	ParentID  *string
	User      string
}

func decodeVersionedLedger(input []byte, version int) (domain.AgenticFactLedger, legacyFactCollections, bool, error) {
	var metadata ledgerMetadataInput
	var collections ledgerCollectionsInput
	if version < 5 {
		var legacy schemaV1ToV4LedgerInput
		if err := json.Unmarshal(input, &legacy); err != nil {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		}
		metadata, collections = legacy.ledgerMetadataInput, legacy.ledgerCollectionsInput
	} else {
		var current schemaV5LedgerInput
		if err := json.Unmarshal(input, &current); err != nil {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		}
		metadata, collections = current.ledgerMetadataInput, current.ledgerCollectionsInput
	}

	// Probe only named schema keys. Collection presence (including explicit
	// null) controls the mixed-shape guard; it never identifies arbitrary
	// objects as invocation/task records.
	var keys struct {
		Sessions    memberPresence `json:"sessions"`
		Invocations memberPresence `json:"invocations"`
		Tools       memberPresence `json:"tools"`
		Events      memberPresence `json:"events"`
	}
	if err := json.Unmarshal(input, &keys); err != nil {
		return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	hasSessions := bool(keys.Sessions)
	for _, field := range []struct {
		name    string
		present memberPresence
	}{
		{"invocations", keys.Invocations}, {"tools", keys.Tools}, {"events", keys.Events},
	} {
		if hasSessions && bool(field.present) {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("mixed ledger shape: canonical sessions cannot coexist with root-level legacy %s; remove ambiguity before migration", field.name)
		}
	}
	if hasSessions && version >= 4 {
		var canonical struct {
			Sessions []struct {
				Invocations []struct {
					ID memberPresence `json:"id"`
				} `json:"invocations"`
			} `json:"sessions"`
		}
		if err := json.Unmarshal(input, &canonical); err != nil {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("decoding canonical sessions: %w", err)
		}
		for si, session := range canonical.Sessions {
			for ii, row := range session.Invocations {
				if row.ID {
					return domain.AgenticFactLedger{}, legacyFactCollections{}, false, fmt.Errorf("canonical schema v%d invocation %d in session %d must not contain id", version, ii, si)
				}
			}
		}
	}

	ledger, legacy, err := collections.toDomain(metadata, version >= 5)
	return ledger, legacy, hasSessions, err
}

func (input ledgerCollectionsInput) toDomain(metadata ledgerMetadataInput, required bool) (domain.AgenticFactLedger, legacyFactCollections, error) {
	ledger := domain.AgenticFactLedger{
		SchemaVersion: metadata.SchemaVersion, Migration: metadata.Migration,
		TicketID: metadata.TicketID, TicketTitle: metadata.TicketTitle, TicketType: metadata.TicketType,
		StartedAt: metadata.StartedAt, CompletedAt: metadata.CompletedAt, UpdatedAt: metadata.UpdatedAt,
		Code: metadata.Code, Anomalies: metadata.Anomalies, Skills: metadata.Skills,
	}
	if input.Sessions != nil {
		ledger.Sessions = make([]domain.SessionFact, len(input.Sessions))
	}
	for si, session := range input.Sessions {
		owner := domain.SessionFact{
			ID: session.ID, ParentID: session.ParentID, Source: session.Source, User: session.User,
			Tools: session.Tools, Events: session.Events, Anomalies: session.Anomalies,
		}
		if session.Invocations != nil {
			owner.Invocations = make([]domain.InvocationFact, len(session.Invocations))
		}
		for ii, row := range session.Invocations {
			invocation, err := row.toDomain(required)
			if err != nil {
				return domain.AgenticFactLedger{}, legacyFactCollections{}, fmt.Errorf("sessions[%d].invocations[%d]: %w", si, ii, err)
			}
			owner.Invocations[ii] = invocation
		}
		ledger.Sessions[si] = owner
	}
	if input.TaskLinks != nil {
		ledger.TaskLinks = make([]domain.TaskLinkFact, len(input.TaskLinks))
	}
	for i, task := range input.TaskLinks {
		amount, evidence, err := decodeKnownCost(task.CostUSD, task.CostEvidence, required)
		if err != nil {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, fmt.Errorf("task_links[%d]: %w", i, err)
		}
		ledger.TaskLinks[i] = domain.TaskLinkFact{
			SubagentType: task.SubagentType, Description: task.Description, ParentTurn: task.ParentTurn,
			ChildSession: task.ChildSession, Timestamp: task.Timestamp, DurationMs: task.DurationMs,
			CostUSD: amount, CostEvidence: evidence,
		}
	}
	legacy := legacyFactCollections{Source: input.Source, Tools: input.Tools, Events: input.Events}
	if input.Invocations != nil {
		legacy.Invocations = make([]legacySessionInvocation, len(input.Invocations))
	}
	for i, row := range input.Invocations {
		invocation, err := row.invocationInput.toDomain(required)
		if err != nil {
			return domain.AgenticFactLedger{}, legacyFactCollections{}, fmt.Errorf("invocations[%d]: %w", i, err)
		}
		legacy.Invocations[i] = legacySessionInvocation{
			InvocationFact: invocation, SessionID: row.SessionID, ParentID: row.ParentID, User: row.User,
		}
	}
	return ledger, legacy, nil
}

func (input invocationInput) toDomain(required bool) (domain.InvocationFact, error) {
	amount, evidence, err := decodeKnownCost(input.CostUSD, input.CostEvidence, required)
	if err != nil {
		return domain.InvocationFact{}, err
	}
	return domain.InvocationFact{
		Role: input.Role, Model: input.Model, StartedAt: input.StartedAt, CompletedAt: input.CompletedAt,
		ActiveDurationSeconds: input.ActiveDurationSeconds, Turns: input.Turns, Tokens: input.Tokens,
		CostUSD: amount, CostEvidence: evidence, Compactions: input.Compactions, UserIterations: input.UserIterations,
	}, nil
}

// decodeKnownCost handles only a declared nullable cost_usd member. The domain
// accumulator receives a partial subtotal only after its wire evidence passes
// validation; legacy null/omission receives explicit unavailable evidence.
func decodeKnownCost(raw json.RawMessage, evidence *domain.CostEvidence, required bool) (float64, domain.CostEvidence, error) {
	if required && len(raw) == 0 {
		return 0, domain.CostEvidence{}, fmt.Errorf("schema v5 requires cost_usd")
	}
	var amount *float64
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &amount); err != nil {
			return 0, domain.CostEvidence{}, fmt.Errorf("decoding cost_usd: %w", err)
		}
	}
	if !required && evidence != nil && evidence.State == "" {
		evidence = nil
	}
	if err := domain.ValidateCostEvidence(amount, evidence, required); err != nil {
		return 0, domain.CostEvidence{}, err
	}
	if evidence == nil {
		if amount == nil {
			return 0, domain.UnavailableCost(), nil
		}
		return *amount, domain.ObservedCost(*amount), nil
	}
	if evidence.State == domain.CostUnavailable {
		return 0, *evidence, nil
	}
	return *evidence.KnownSubtotalUSD, *evidence, nil
}
