package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"agenticlens/internal/domain"
)

// ErrToolSessionIDRequired identifies a fact ledger that has tool aggregates
// without the session identity required by schema v4.
var ErrToolSessionIDRequired = errors.New("tool fact requires session_id")

// InvocationTokensDTO represents the token counts in JSON boundary serialization.
type InvocationTokensDTO struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CacheRead   int64 `json:"cache_read"`
	Reasoning   int64 `json:"reasoning"`
	PeakContext int64 `json:"peak_context"`
}

func (dto InvocationTokensDTO) ToDomain() domain.InvocationTokens {
	return domain.InvocationTokens{
		Input:       dto.Input,
		Output:      dto.Output,
		CacheRead:   dto.CacheRead,
		Reasoning:   dto.Reasoning,
		PeakContext: dto.PeakContext,
	}
}

// InvocationFactDTO represents a single invocation fact in JSON boundary serialization.
type InvocationFactDTO struct {
	Role                  string              `json:"role"`
	Model                 string              `json:"model"`
	StartedAt             *string             `json:"started_at"`
	CompletedAt           *string             `json:"completed_at"`
	ActiveDurationSeconds int64               `json:"active_duration_seconds"`
	Turns                 int64               `json:"turns"`
	Tokens                InvocationTokensDTO `json:"tokens"`
	CostUSD               float64             `json:"cost_usd"`
	Compactions           int64               `json:"compactions"`
	UserIterations        int64               `json:"user_iterations"`
}

func (dto InvocationFactDTO) ToDomain() domain.InvocationFact {
	return domain.InvocationFact{
		Role:                  domain.AgentRole(dto.Role),
		Model:                 dto.Model,
		StartedAt:             dto.StartedAt,
		CompletedAt:           dto.CompletedAt,
		ActiveDurationSeconds: dto.ActiveDurationSeconds,
		Turns:                 dto.Turns,
		Tokens:                dto.Tokens.ToDomain(),
		CostUSD:               dto.CostUSD,
		Compactions:           dto.Compactions,
		UserIterations:        dto.UserIterations,
	}
}

// InvocationFactDTOFromDomain converts domain.InvocationFact to InvocationFactDTO.
func InvocationFactDTOFromDomain(inv domain.InvocationFact) InvocationFactDTO {
	return InvocationFactDTO{
		Role:                  string(inv.Role),
		Model:                 inv.Model,
		StartedAt:             inv.StartedAt,
		CompletedAt:           inv.CompletedAt,
		ActiveDurationSeconds: inv.ActiveDurationSeconds,
		Turns:                 inv.Turns,
		Tokens: InvocationTokensDTO{
			Input:       inv.Tokens.Input,
			Output:      inv.Tokens.Output,
			CacheRead:   inv.Tokens.CacheRead,
			Reasoning:   inv.Tokens.Reasoning,
			PeakContext: inv.Tokens.PeakContext,
		},
		CostUSD:        inv.CostUSD,
		Compactions:    inv.Compactions,
		UserIterations: inv.UserIterations,
	}
}

// ToolFactDTO represents a single tool execution fact in JSON.
type ToolFactDTO struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	Calls     int64  `json:"calls"`
	Errors    int64  `json:"errors"`
}

func (dto ToolFactDTO) ToDomain() domain.ToolFact {
	return domain.ToolFact{
		SessionID: dto.SessionID,
		Name:      dto.Name,
		Calls:     dto.Calls,
		Errors:    dto.Errors,
	}
}

// EventFactDTO represents a single chronological event fact in JSON boundary serialization.
type EventFactDTO struct {
	EventID       string  `json:"event_id"`
	Timestamp     string  `json:"timestamp"`
	SessionID     string  `json:"session_id"`
	Turn          int64   `json:"turn"`
	Role          string  `json:"role"`
	EventType     string  `json:"event_type"`
	ToolName      *string `json:"tool_name,omitempty"`
	Status        string  `json:"status"`
	ErrorMessage  *string `json:"error_message,omitempty"`
	DurationMs    *int64  `json:"duration_ms,omitempty"`
	ContextTokens int64   `json:"context_tokens"`
}

type SessionToolFactDTO struct {
	Name   string `json:"name"`
	Calls  int64  `json:"calls"`
	Errors int64  `json:"errors"`
}

type SessionEventFactDTO struct {
	EventID       string  `json:"event_id"`
	Timestamp     string  `json:"timestamp"`
	MessageID     *string `json:"message_id,omitempty"`
	Turn          int64   `json:"turn"`
	Role          string  `json:"role"`
	EventType     string  `json:"event_type"`
	ToolName      *string `json:"tool_name,omitempty"`
	Status        string  `json:"status"`
	DurationMs    *int64  `json:"duration_ms,omitempty"`
	ContextTokens int64   `json:"context_tokens"`
}

type SessionFactDTO struct {
	ID          string                `json:"id"`
	ParentID    *string               `json:"parent_id,omitempty"`
	Source      string                `json:"source,omitempty"`
	User        string                `json:"user,omitempty"`
	Invocations []InvocationFactDTO   `json:"invocations"`
	Tools       []SessionToolFactDTO  `json:"tools"`
	Events      []SessionEventFactDTO `json:"events,omitempty"`
	Anomalies   []AnomalyDTO          `json:"anomalies"`
}

func (dto EventFactDTO) ToDomain() domain.EventFact {
	role := dto.Role
	if role == "" {
		role = "orchestrator"
	}
	status := dto.Status
	if status == "" {
		status = "completed"
	}
	return domain.EventFact{
		EventID:       dto.EventID,
		Timestamp:     dto.Timestamp,
		SessionID:     dto.SessionID,
		Turn:          dto.Turn,
		Role:          domain.AgentRole(role),
		EventType:     domain.EventType(dto.EventType),
		ToolName:      dto.ToolName,
		Status:        status,
		ErrorMessage:  dto.ErrorMessage,
		DurationMs:    dto.DurationMs,
		ContextTokens: dto.ContextTokens,
	}
}

// EventFactDTOFromDomain converts domain.EventFact to EventFactDTO.
func EventFactDTOFromDomain(fact domain.EventFact) EventFactDTO {
	return EventFactDTO{
		EventID:       fact.EventID,
		Timestamp:     fact.Timestamp,
		SessionID:     fact.SessionID,
		Turn:          fact.Turn,
		Role:          string(fact.Role),
		EventType:     string(fact.EventType),
		ToolName:      fact.ToolName,
		Status:        fact.Status,
		ErrorMessage:  fact.ErrorMessage,
		DurationMs:    fact.DurationMs,
		ContextTokens: fact.ContextTokens,
	}
}

// CodeFactsDTO represents code facts in JSON.
type CodeFactsDTO struct {
	LinesAdded       int64 `json:"lines_added"`
	LinesDeleted     int64 `json:"lines_deleted"`
	TestLinesAdded   int64 `json:"test_lines_added"`
	TestLinesDeleted int64 `json:"test_lines_deleted"`
	ProdLinesAdded   int64 `json:"prod_lines_added"`
	ProdLinesDeleted int64 `json:"prod_lines_deleted"`
	FilesChanged     int64 `json:"files_changed"`
	SDLCIterations   int64 `json:"sdlc_iterations"`
}

func (dto CodeFactsDTO) ToDomain() domain.CodeFacts {
	return domain.CodeFacts{
		LinesAdded:       dto.LinesAdded,
		LinesDeleted:     dto.LinesDeleted,
		TestLinesAdded:   dto.TestLinesAdded,
		TestLinesDeleted: dto.TestLinesDeleted,
		ProdLinesAdded:   dto.ProdLinesAdded,
		ProdLinesDeleted: dto.ProdLinesDeleted,
		FilesChanged:     dto.FilesChanged,
		SDLCIterations:   dto.SDLCIterations,
	}
}

// MigrationInfoDTO represents migration metadata in JSON boundary serialization.
type MigrationInfoDTO struct {
	SchemaVersion  int      `json:"schema_version"`
	Status         string   `json:"status"`
	StepsCompleted []string `json:"steps_completed"`
	StepsPending   []string `json:"steps_pending"`
	LastMigratedAt string   `json:"last_migrated_at"`
	LastMigratedBy string   `json:"last_migrated_by"`
	Notes          string   `json:"notes,omitempty"`
}

func (dto MigrationInfoDTO) ToDomain() domain.MigrationInfo {
	status := domain.MigrationStatus(dto.Status)
	if status == "" {
		status = domain.MigrationStatusCompleted
	}
	stepsCompleted := dto.StepsCompleted
	if stepsCompleted == nil {
		stepsCompleted = []string{}
	}
	stepsPending := dto.StepsPending
	if stepsPending == nil {
		stepsPending = []string{}
	}
	return domain.MigrationInfo{
		SchemaVersion:  dto.SchemaVersion,
		Status:         status,
		StepsCompleted: stepsCompleted,
		StepsPending:   stepsPending,
		LastMigratedAt: dto.LastMigratedAt,
		LastMigratedBy: dto.LastMigratedBy,
		Notes:          dto.Notes,
	}
}

// MigrationInfoDTOFromDomain converts domain.MigrationInfo to MigrationInfoDTO.
func MigrationInfoDTOFromDomain(info *domain.MigrationInfo) *MigrationInfoDTO {
	if info == nil {
		return nil
	}
	return &MigrationInfoDTO{
		SchemaVersion:  info.SchemaVersion,
		Status:         string(info.Status),
		StepsCompleted: info.StepsCompleted,
		StepsPending:   info.StepsPending,
		LastMigratedAt: info.LastMigratedAt,
		LastMigratedBy: info.LastMigratedBy,
		Notes:          info.Notes,
	}
}

// AgenticFactLedgerDTO represents the complete raw agentic_metrics.json schema.
type AgenticFactLedgerDTO struct {
	SchemaVersion *int              `json:"schema_version,omitempty"`
	Source        string            `json:"source,omitempty"`
	Migration     *MigrationInfoDTO `json:"migration,omitempty"`
	TicketID      *string           `json:"ticket_id"`
	TicketTitle   *string           `json:"ticket_title"`
	TicketType    *string           `json:"ticket_type"`
	StartedAt     *string           `json:"started_at"`
	CompletedAt   *string           `json:"completed_at"`
	UpdatedAt     string            `json:"updated_at"`
	Code          CodeFactsDTO      `json:"code"`
	Sessions      []SessionFactDTO  `json:"sessions"`
	Anomalies     []AnomalyDTO      `json:"anomalies"`
}

// ValidateRuntimeFactLedger rejects legacy or explicitly incomplete ledgers before
// any fact collections are decoded by ordinary ingestion paths.
func ValidateRuntimeFactLedger(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("decoding ledger metadata: %w", err)
	}
	var metadata struct {
		SchemaVersion *int `json:"schema_version"`
		Migration     *struct {
			Status         string   `json:"status"`
			StepsCompleted []string `json:"steps_completed"`
			StepsPending   []string `json:"steps_pending"`
		} `json:"migration"`
	}
	if err := json.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("decoding ledger metadata: %w", err)
	}
	if metadata.SchemaVersion == nil {
		return errors.New("ledger schema version is missing; run `agentic-lens migrate --target-version 4` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`")
	}
	if *metadata.SchemaVersion != domain.CurrentSchemaVersion {
		return fmt.Errorf("ledger schema version %d is not runtime-ingestible; run `agentic-lens migrate --target-version 4` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`", *metadata.SchemaVersion)
	}
	for _, key := range []string{"invocations", "tools", "events"} {
		if _, exists := fields[key]; exists {
			return fmt.Errorf("schema v4 ledger contains ambiguous root-level %s; run `agentic-lens migrate --target-version 4` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`", key)
		}
	}
	sessionsJSON, sessionsExists := fields["sessions"]
	if !sessionsExists || string(sessionsJSON) == "null" || !json.Valid(sessionsJSON) {
		return errors.New("canonical schema v4 ledger requires a sessions collection; run `agentic-lens migrate --target-version 4` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`")
	}
	if metadata.Migration != nil && (metadata.Migration.Status != string(domain.MigrationStatusCompleted) || len(metadata.Migration.StepsPending) > 0) {
		return errors.New("schema v4 ledger has incomplete migration metadata; resume the explicit migration and source-backed backfill workflow before runtime ingestion")
	}
	var canonical struct {
		Sessions []json.RawMessage `json:"sessions"`
	}
	if err := json.Unmarshal(sessionsJSON, &canonical.Sessions); err != nil || canonical.Sessions == nil {
		return errors.New("canonical schema v4 sessions must be an array")
	}
	for si, rawSession := range canonical.Sessions {
		var session map[string]json.RawMessage
		if err := json.Unmarshal(rawSession, &session); err != nil {
			return fmt.Errorf("canonical schema v4 session %d is invalid", si)
		}
		for _, required := range []string{"id", "source", "user", "invocations", "tools", "anomalies"} {
			if _, ok := session[required]; !ok {
				return fmt.Errorf("canonical schema v4 session %d is missing required %s", si, required)
			}
		}
		var sessionUser string
		if err := json.Unmarshal(session["user"], &sessionUser); err != nil || strings.TrimSpace(sessionUser) == "" {
			return fmt.Errorf("canonical schema v4 session %d requires a non-empty user", si)
		}
		for _, collection := range []string{"tools", "anomalies"} {
			var values []json.RawMessage
			if err := json.Unmarshal(session[collection], &values); err != nil || values == nil {
				return fmt.Errorf("canonical schema v4 session %d %s must be an array", si, collection)
			}
		}
		var source string
		if err := json.Unmarshal(session["source"], &source); err != nil || (source != string(domain.SourceOpenCode) && source != string(domain.SourceCodex)) {
			return fmt.Errorf("canonical schema v4 session %d requires a supported source (opencode or codex)", si)
		}
		var invocations []json.RawMessage
		if err := json.Unmarshal(session["invocations"], &invocations); err != nil || invocations == nil {
			return fmt.Errorf("canonical schema v4 session %d invocations must be an array", si)
		}
		for ii, rawInvocation := range invocations {
			var invocation map[string]json.RawMessage
			if err := json.Unmarshal(rawInvocation, &invocation); err != nil {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d is invalid", ii, si)
			}
			if _, ok := invocation["id"]; ok {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d must not contain id", ii, si)
			}
			if _, ok := invocation["user"]; ok {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d must not contain user", ii, si)
			}
			if _, ok := invocation["parent_id"]; ok {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d must not contain parent_id", ii, si)
			}
			for _, required := range []string{"role", "model", "started_at", "completed_at", "active_duration_seconds", "turns", "tokens", "cost_usd", "compactions", "user_iterations"} {
				if _, ok := invocation[required]; !ok {
					return fmt.Errorf("canonical schema v4 invocation %d in session %d is missing required %s", ii, si, required)
				}
			}
			var tokens map[string]json.RawMessage
			if err := json.Unmarshal(invocation["tokens"], &tokens); err != nil || tokens == nil {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d requires token counts", ii, si)
			}
			for _, required := range []string{"input", "output", "cache_read", "reasoning", "peak_context"} {
				if _, ok := tokens[required]; !ok {
					return fmt.Errorf("canonical schema v4 invocation %d in session %d is missing token count %s", ii, si, required)
				}
			}
			value, ok := invocation["user_iterations"]
			var count int64
			if !ok || json.Unmarshal(value, &count) != nil || count < 0 {
				return fmt.Errorf("canonical schema v4 invocation %d in session %d requires a non-negative numeric user_iterations (zero is valid)", ii, si)
			}
		}
	}
	return nil
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// UnmarshalJSON rejects legacy root fact collections at the current schema boundary.
// They must be passed through the versioned migration engine instead.
func (dto *AgenticFactLedgerDTO) UnmarshalJSON(data []byte) error {
	type currentDTO AgenticFactLedgerDTO
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, legacy := range []string{"invocations", "tools", "events"} {
		if _, exists := fields[legacy]; exists {
			return fmt.Errorf("legacy root-level %s is not valid in session-owned schema v4; run `agentic-lens migrate`", legacy)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded currentDTO
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values in fact ledger")
		}
		return err
	}
	*dto = AgenticFactLedgerDTO(decoded)
	return nil
}

func toDomainTicketType(raw *string) *domain.ConventionalTicketType {
	if raw == nil {
		return nil
	}
	if tt, ok := domain.NormalizeTicketType(*raw); ok {
		return &tt
	}
	ct := domain.ConventionalTicketType(*raw)
	return &ct
}

func fromDomainTicketType(tt *domain.ConventionalTicketType) *string {
	if tt == nil {
		return nil
	}
	s := string(*tt)
	return &s
}

func toDomainAgentRole(raw *string) *domain.AgentRole {
	if raw == nil {
		return nil
	}
	r := domain.AgentRole(*raw)
	return &r
}

func fromDomainAgentRole(r *domain.AgentRole) *string {
	if r == nil {
		return nil
	}
	s := string(*r)
	return &s
}

func (dto AgenticFactLedgerDTO) ToDomain() domain.AgenticFactLedger {
	sessions := make([]domain.SessionFact, len(dto.Sessions))
	for i, session := range dto.Sessions {
		sessionSource := domain.SourceHarness(strings.TrimSpace(session.Source))
		if sessionSource == "" && dto.Source != "" {
			sessionSource = domain.SourceHarness(strings.TrimSpace(dto.Source))
		}
		user := strings.TrimSpace(session.User)
		parentID := session.ParentID
		sessions[i] = domain.SessionFact{
			ID:       session.ID,
			ParentID: parentID,
			Source:   sessionSource,
			User:     user,
		}
		if session.Anomalies != nil {
			sessions[i].Anomalies = []domain.Anomaly{}
		}
		for _, inv := range session.Invocations {
			sessions[i].Invocations = append(sessions[i].Invocations, inv.ToDomain())
		}
		for _, tool := range session.Tools {
			name := tool.Name
			if name == "" {
				name = "unknown"
			}
			sessions[i].Tools = append(sessions[i].Tools, domain.SessionToolFact{Name: name, Calls: tool.Calls, Errors: tool.Errors})
		}
		for _, event := range session.Events {
			sessions[i].Events = append(sessions[i].Events, domain.SessionEventFact{
				EventID: event.EventID, Timestamp: event.Timestamp, MessageID: event.MessageID, Turn: event.Turn,
				Role: domain.AgentRole(event.Role), EventType: domain.EventType(event.EventType), ToolName: event.ToolName,
				Status: event.Status, DurationMs: event.DurationMs, ContextTokens: event.ContextTokens,
			})
		}
		for _, anomaly := range session.Anomalies {
			detected := anomaly.ToDomain()
			detected.Context.SessionID = nil
			sessions[i].Anomalies = append(sessions[i].Anomalies, detected)
		}
	}

	var mig *domain.MigrationInfo
	if dto.Migration != nil {
		domMig := dto.Migration.ToDomain()
		mig = &domMig
	}

	return domain.AgenticFactLedger{
		SchemaVersion: dto.SchemaVersion,
		Migration:     mig,
		TicketID:      dto.TicketID,
		TicketTitle:   dto.TicketTitle,
		TicketType:    toDomainTicketType(dto.TicketType),
		StartedAt:     dto.StartedAt,
		CompletedAt:   dto.CompletedAt,
		UpdatedAt:     dto.UpdatedAt,
		Code:          dto.Code.ToDomain(),
		Sessions:      sessions,
		Anomalies:     anomaliesToDomain(dto.Anomalies),
	}
}

// Validate enforces required fields at the persisted telemetry boundary.
func (dto AgenticFactLedgerDTO) Validate() error {
	currentSchema := dto.SchemaVersion != nil && *dto.SchemaVersion >= domain.CurrentSchemaVersion
	if currentSchema && dto.Sessions == nil {
		return errors.New("sessions collection is required in schema v4")
	}
	seen := make(map[string]bool, len(dto.Sessions))
	for i, session := range dto.Sessions {
		if strings.TrimSpace(session.ID) == "" {
			return fmt.Errorf("sessions[%d].id is required", i)
		}
		if session.ID != strings.TrimSpace(session.ID) {
			return fmt.Errorf("sessions[%d].id must not have leading or trailing whitespace", i)
		}
		if seen[session.ID] {
			return fmt.Errorf("sessions[%d].id %q is duplicated", i, session.ID)
		}
		seen[session.ID] = true
	}
	for i, session := range dto.Sessions {
		if currentSchema && strings.TrimSpace(session.Source) == "" {
			return fmt.Errorf("sessions[%d].source is required", i)
		}
		if currentSchema && strings.TrimSpace(session.User) == "" {
			return fmt.Errorf("sessions[%d].user is required", i)
		}
		if session.User != strings.TrimSpace(session.User) {
			return fmt.Errorf("sessions[%d].user must not have leading or trailing whitespace", i)
		}
		if currentSchema && (session.Invocations == nil || session.Tools == nil) {
			return fmt.Errorf("sessions[%d] requires invocations and tools collections", i)
		}
		if session.ParentID != nil {
			if *session.ParentID != strings.TrimSpace(*session.ParentID) {
				return fmt.Errorf("sessions[%d].parent_id must not have leading or trailing whitespace", i)
			}
			if !seen[*session.ParentID] {
				return fmt.Errorf("sessions[%d].parent_id %q does not identify a session", i, *session.ParentID)
			}
		}
		for j, tool := range session.Tools {
			if strings.TrimSpace(tool.Name) == "" {
				return fmt.Errorf("sessions[%d].tools[%d].name is required", i, j)
			}
			if tool.Calls < 0 || tool.Errors < 0 || tool.Errors > tool.Calls {
				return fmt.Errorf("sessions[%d].tools[%d] has invalid calls/errors counts", i, j)
			}
		}
		toolNames := make(map[string]bool, len(session.Tools))
		for j, tool := range session.Tools {
			if toolNames[tool.Name] {
				return fmt.Errorf("sessions[%d].tools[%d].name %q is duplicated", i, j, tool.Name)
			}
			toolNames[tool.Name] = true
		}
		for j, anomaly := range session.Anomalies {
			if anomaly.Context.SessionID != nil {
				return fmt.Errorf("sessions[%d].anomalies[%d].context.session_id must be inherited from the owning session", i, j)
			}
		}
	}
	for i, anomaly := range dto.Anomalies {
		if anomaly.Context.SessionID != nil {
			return fmt.Errorf("anomalies[%d].context.session_id is not allowed for ticket-level summaries", i)
		}
	}
	return nil
}

func validateAnomalyEvidence(anomaly AnomalyDTO, ownerSession string, nested bool, knownEvents map[string]EventFactDTO, canonical bool) error {
	seenRefs := make(map[string]bool, len(anomaly.Evidence))
	for i, ref := range anomaly.Evidence {
		if strings.TrimSpace(ref.SessionID) == "" || strings.TrimSpace(ref.EventID) == "" {
			return fmt.Errorf("evidence[%d] requires session_id and event_id", i)
		}
		if nested && ref.SessionID != ownerSession {
			return fmt.Errorf("evidence[%d].session_id must match owning session %q", i, ownerSession)
		}
		key := ref.SessionID + "\x00" + ref.EventID
		if seenRefs[key] {
			return fmt.Errorf("evidence[%d] duplicates session=%q event_id=%q", i, ref.SessionID, ref.EventID)
		}
		seenRefs[key] = true
		_, found := knownEvents[key]
		if canonical && !found {
			return fmt.Errorf("evidence[%d] references missing event session=%q event_id=%q", i, ref.SessionID, ref.EventID)
		}
	}
	if anomaly.RuleID == string(domain.RuleToolHighFailureRate) {
		if anomaly.Context.Count == nil || int64(len(anomaly.Evidence)) != *anomaly.Context.Count {
			return errors.New("high tool failure anomaly must reference every failed event in its count")
		}
		if canonical {
			toolName := ""
			if anomaly.Context.ToolName != nil {
				toolName = strings.TrimSpace(*anomaly.Context.ToolName)
			}
			if toolName == "" {
				return errors.New("high tool failure anomaly requires a tool_name")
			}
			for i, ref := range anomaly.Evidence {
				event := knownEvents[ref.SessionID+"\x00"+ref.EventID]
				if event.EventType != string(domain.EventTypeToolCall) || (event.Status != "error" && event.Status != "failed") || event.ToolName == nil || *event.ToolName != toolName {
					return fmt.Errorf("evidence[%d] must reference a failed tool_call for tool %q", i, toolName)
				}
			}
		}
	}
	return nil
}

// ValidatePersistedEvidence enforces stable source identity on rows about to be written.
// Read validation remains tolerant of pre-iteration v4 ledgers so they can be normalized.
func (dto AgenticFactLedgerDTO) ValidatePersistedEvidence() error {
	for i, session := range dto.Sessions {
		if strings.TrimSpace(session.Source) == "" {
			return fmt.Errorf("sessions[%d].source is required", i)
		}
	}
	return dto.Validate()
}

// ValidateToolFacts enforces session identity for facts before writing a ledger.
func ValidateToolFacts(tools []domain.ToolFact) error {
	for i, tool := range tools {
		if strings.TrimSpace(tool.SessionID) == "" {
			return fmt.Errorf("%w: tools[%d] is missing; regenerate or re-extract this ledger from source logs", ErrToolSessionIDRequired, i)
		}
	}
	return nil
}

// AgenticFactLedgerDTOFromDomain converts a domain AgenticFactLedger to its DTO for serialization.
func AgenticFactLedgerDTOFromDomain(ledger domain.AgenticFactLedger) AgenticFactLedgerDTO {
	sessions := make([]SessionFactDTO, len(ledger.Sessions))
	for i, session := range ledger.Sessions {
		sessionSource := strings.TrimSpace(string(session.Source))
		sessions[i] = SessionFactDTO{
			ID:          session.ID,
			ParentID:    session.ParentID,
			Source:      sessionSource,
			User:        session.User,
			Invocations: []InvocationFactDTO{},
			Tools:       []SessionToolFactDTO{},
			Anomalies:   []AnomalyDTO{},
		}
		for _, inv := range session.Invocations {
			sessions[i].Invocations = append(sessions[i].Invocations, InvocationFactDTOFromDomain(inv))
		}
		for _, tool := range session.Tools {
			sessions[i].Tools = append(sessions[i].Tools, SessionToolFactDTO{Name: tool.Name, Calls: tool.Calls, Errors: tool.Errors})
		}
		for _, anomaly := range session.Anomalies {
			anomaly.Context.SessionID = nil
			sessions[i].Anomalies = append(sessions[i].Anomalies, AnomalyDTOFromDomain(anomaly))
		}
	}

	return AgenticFactLedgerDTO{
		SchemaVersion: ledger.SchemaVersion,
		Migration:     MigrationInfoDTOFromDomain(ledger.Migration),
		TicketID:      ledger.TicketID,
		TicketTitle:   ledger.TicketTitle,
		TicketType:    fromDomainTicketType(ledger.TicketType),
		StartedAt:     ledger.StartedAt,
		CompletedAt:   ledger.CompletedAt,
		UpdatedAt:     ledger.UpdatedAt,
		Code: CodeFactsDTO{
			LinesAdded:       ledger.Code.LinesAdded,
			LinesDeleted:     ledger.Code.LinesDeleted,
			TestLinesAdded:   ledger.Code.TestLinesAdded,
			TestLinesDeleted: ledger.Code.TestLinesDeleted,
			ProdLinesAdded:   ledger.Code.ProdLinesAdded,
			ProdLinesDeleted: ledger.Code.ProdLinesDeleted,
			FilesChanged:     ledger.Code.FilesChanged,
			SDLCIterations:   ledger.Code.SDLCIterations,
		},
		Sessions:  sessions,
		Anomalies: anomaliesFromDomain(ledger.Anomalies),
	}
}

// AnomalyContextDTO represents anomaly contextual metadata in JSON boundary serialization.
type AnomalyContextDTO struct {
	ToolName       *string  `json:"tool_name,omitempty"`
	SessionID      *string  `json:"session_id,omitempty"`
	Turn           *int64   `json:"turn,omitempty"`
	Role           *string  `json:"role,omitempty"`
	SecondaryRole  *string  `json:"secondary_role,omitempty"`
	Count          *int64   `json:"count,omitempty"`
	TotalCalls     *int64   `json:"total_calls,omitempty"`
	TokenDelta     *int64   `json:"token_delta,omitempty"`
	ContextTokens  *int64   `json:"context_tokens,omitempty"`
	Threshold      *string  `json:"threshold,omitempty"`
	FailureRatePct *float64 `json:"failure_rate_pct,omitempty"`
}

func (dto AnomalyContextDTO) ToDomain() domain.AnomalyContext {
	return domain.AnomalyContext{
		ToolName:       dto.ToolName,
		SessionID:      dto.SessionID,
		Turn:           dto.Turn,
		Role:           toDomainAgentRole(dto.Role),
		SecondaryRole:  toDomainAgentRole(dto.SecondaryRole),
		Count:          dto.Count,
		TotalCalls:     dto.TotalCalls,
		TokenDelta:     dto.TokenDelta,
		ContextTokens:  dto.ContextTokens,
		Threshold:      dto.Threshold,
		FailureRatePct: dto.FailureRatePct,
	}
}

// AnomalyContextDTOFromDomain converts domain.AnomalyContext to AnomalyContextDTO.
func AnomalyContextDTOFromDomain(ctx domain.AnomalyContext) AnomalyContextDTO {
	return AnomalyContextDTO{
		ToolName:       ctx.ToolName,
		SessionID:      ctx.SessionID,
		Turn:           ctx.Turn,
		Role:           fromDomainAgentRole(ctx.Role),
		SecondaryRole:  fromDomainAgentRole(ctx.SecondaryRole),
		Count:          ctx.Count,
		TotalCalls:     ctx.TotalCalls,
		TokenDelta:     ctx.TokenDelta,
		ContextTokens:  ctx.ContextTokens,
		Threshold:      ctx.Threshold,
		FailureRatePct: ctx.FailureRatePct,
	}
}

// AnomalyDTO represents a detected domain execution anomaly in JSON boundary serialization.
type AnomalyDTO struct {
	RuleID   string                  `json:"rule_id"`
	Severity string                  `json:"severity"`
	Context  AnomalyContextDTO       `json:"context"`
	Evidence []AnomalyEvidenceRefDTO `json:"evidence,omitempty"`
}

// UnmarshalJSON accepts old presentation fields so strict ledger decoding can
// normalize them away while continuing to reject unrelated unknown fields.
func (dto *AnomalyDTO) UnmarshalJSON(data []byte) error {
	type anomalyDTOAlias AnomalyDTO
	var decoded struct {
		*anomalyDTOAlias
		Title       json.RawMessage `json:"title"`
		Description json.RawMessage `json:"description"`
	}
	decoded.anomalyDTOAlias = (*anomalyDTOAlias)(dto)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	return nil
}

type AnomalyEvidenceRefDTO struct {
	SessionID string `json:"session_id"`
	EventID   string `json:"event_id"`
}

func (dto AnomalyDTO) ToDomain() domain.Anomaly {
	return domain.Anomaly{
		RuleID:   domain.AnomalyRuleID(dto.RuleID),
		Severity: domain.AnomalySeverity(dto.Severity),
		Context:  dto.Context.ToDomain(),
	}
}

// AnomalyDTOFromDomain converts domain.Anomaly to AnomalyDTO.
func AnomalyDTOFromDomain(anomaly domain.Anomaly) AnomalyDTO {
	return AnomalyDTO{
		RuleID:   string(anomaly.RuleID),
		Severity: string(anomaly.Severity),
		Context:  AnomalyContextDTOFromDomain(anomaly.Context),
	}
}

func evidenceRefsToDomain(items []AnomalyEvidenceRefDTO) []domain.AnomalyEvidenceRef {
	if items == nil {
		return nil
	}
	result := make([]domain.AnomalyEvidenceRef, len(items))
	for i, item := range items {
		result[i] = domain.AnomalyEvidenceRef{SessionID: item.SessionID, EventID: item.EventID}
	}
	return result
}

func evidenceRefsFromDomain(items []domain.AnomalyEvidenceRef) []AnomalyEvidenceRefDTO {
	if items == nil {
		return nil
	}
	result := make([]AnomalyEvidenceRefDTO, len(items))
	for i, item := range items {
		result[i] = AnomalyEvidenceRefDTO{SessionID: item.SessionID, EventID: item.EventID}
	}
	return result
}

func anomaliesToDomain(items []AnomalyDTO) []domain.Anomaly {
	if items == nil {
		return nil
	}
	result := make([]domain.Anomaly, len(items))
	for i, item := range items {
		result[i] = item.ToDomain()
	}
	return result
}

func anomaliesFromDomain(items []domain.Anomaly) []AnomalyDTO {
	result := make([]AnomalyDTO, len(items))
	for i, item := range items {
		result[i] = AnomalyDTOFromDomain(item)
	}
	return result
}

// BrittleToolSummaryDTO represents brittle tool aggregates in JSON boundary serialization.
type BrittleToolSummaryDTO struct {
	ToolName       string  `json:"tool_name"`
	AnomalyCount   int64   `json:"anomaly_count"`
	TotalCalls     int64   `json:"total_calls"`
	ErrorCount     int64   `json:"error_count"`
	FailureRatePct float64 `json:"failure_rate_pct"`
}

func (dto BrittleToolSummaryDTO) ToDomain() domain.BrittleToolSummary {
	return domain.BrittleToolSummary{
		ToolName:       dto.ToolName,
		AnomalyCount:   dto.AnomalyCount,
		TotalCalls:     dto.TotalCalls,
		ErrorCount:     dto.ErrorCount,
		FailureRatePct: dto.FailureRatePct,
	}
}

// BrittleToolSummaryDTOFromDomain converts domain.BrittleToolSummary to BrittleToolSummaryDTO.
func BrittleToolSummaryDTOFromDomain(summary domain.BrittleToolSummary) BrittleToolSummaryDTO {
	return BrittleToolSummaryDTO{
		ToolName:       summary.ToolName,
		AnomalyCount:   summary.AnomalyCount,
		TotalCalls:     summary.TotalCalls,
		ErrorCount:     summary.ErrorCount,
		FailureRatePct: summary.FailureRatePct,
	}
}

// FleetAnomalySummaryDTO represents the fleet anomaly rollup in JSON boundary serialization.
type FleetAnomalySummaryDTO struct {
	TotalAnomalies       int64                   `json:"total_anomalies"`
	WarningCount         int64                   `json:"warning_count"`
	CriticalCount        int64                   `json:"critical_count"`
	SeverityCounts       map[string]int64        `json:"severity_counts"`
	RuleCounts           map[string]int64        `json:"rule_counts"`
	TotalTickets         int64                   `json:"total_tickets"`
	TicketsWithAnomalies int64                   `json:"tickets_with_anomalies"`
	AnomalyRatePerTicket float64                 `json:"anomaly_rate_per_ticket"`
	BrittleTools         []BrittleToolSummaryDTO `json:"brittle_tools"`
}

func (dto FleetAnomalySummaryDTO) ToDomain() domain.FleetAnomalySummary {
	severityCounts := make(map[string]int64)
	for k, v := range dto.SeverityCounts {
		severityCounts[k] = v
	}
	ruleCounts := make(map[string]int64)
	for k, v := range dto.RuleCounts {
		ruleCounts[k] = v
	}
	brittleTools := make([]domain.BrittleToolSummary, len(dto.BrittleTools))
	for i, bt := range dto.BrittleTools {
		brittleTools[i] = bt.ToDomain()
	}
	return domain.FleetAnomalySummary{
		TotalAnomalies:       dto.TotalAnomalies,
		WarningCount:         dto.WarningCount,
		CriticalCount:        dto.CriticalCount,
		SeverityCounts:       severityCounts,
		RuleCounts:           ruleCounts,
		TotalTickets:         dto.TotalTickets,
		TicketsWithAnomalies: dto.TicketsWithAnomalies,
		AnomalyRatePerTicket: dto.AnomalyRatePerTicket,
		BrittleTools:         brittleTools,
	}
}

// FleetAnomalySummaryDTOFromDomain converts domain.FleetAnomalySummary to FleetAnomalySummaryDTO.
func FleetAnomalySummaryDTOFromDomain(summary domain.FleetAnomalySummary) FleetAnomalySummaryDTO {
	severityCounts := make(map[string]int64)
	for k, v := range summary.SeverityCounts {
		severityCounts[k] = v
	}
	ruleCounts := make(map[string]int64)
	for k, v := range summary.RuleCounts {
		ruleCounts[k] = v
	}
	brittleTools := make([]BrittleToolSummaryDTO, len(summary.BrittleTools))
	for i, bt := range summary.BrittleTools {
		brittleTools[i] = BrittleToolSummaryDTOFromDomain(bt)
	}
	return FleetAnomalySummaryDTO{
		TotalAnomalies:       summary.TotalAnomalies,
		WarningCount:         summary.WarningCount,
		CriticalCount:        summary.CriticalCount,
		SeverityCounts:       severityCounts,
		RuleCounts:           ruleCounts,
		TotalTickets:         summary.TotalTickets,
		TicketsWithAnomalies: summary.TicketsWithAnomalies,
		AnomalyRatePerTicket: summary.AnomalyRatePerTicket,
		BrittleTools:         brittleTools,
	}
}

// AgenticMetricsLedgerDTO represents the computed metrics ledger format in JSON.
type AgenticMetricsLedgerDTO struct {
	TicketID           *string                   `json:"ticket_id"`
	TicketTitle        *string                   `json:"ticket_title"`
	TicketType         *string                   `json:"ticket_type"`
	StartedAt          *string                   `json:"started_at"`
	CompletedAt        *string                   `json:"completed_at"`
	UpdatedAt          string                    `json:"updated_at"`
	WorkflowEfficiency domain.WorkflowEfficiency `json:"workflow_efficiency"`
	Totals             domain.GroupedMetric      `json:"totals"`
	Roles              []domain.RoleMetric       `json:"roles"`
	Models             []domain.ModelMetric      `json:"models"`
	Anomalies          []AnomalyDTO              `json:"anomalies,omitempty"`
}

// SerializedLedgerDTO is an alias to AgenticMetricsLedgerDTO for backward compatibility.
type SerializedLedgerDTO = AgenticMetricsLedgerDTO

func (dto AgenticMetricsLedgerDTO) ToDomain() domain.AgenticMetricsLedger {
	var anomalies []domain.Anomaly
	if len(dto.Anomalies) > 0 {
		anomalies = make([]domain.Anomaly, len(dto.Anomalies))
		for i, a := range dto.Anomalies {
			anomalies[i] = a.ToDomain()
		}
	}
	return domain.AgenticMetricsLedger{
		TicketID:           dto.TicketID,
		TicketTitle:        dto.TicketTitle,
		TicketType:         toDomainTicketType(dto.TicketType),
		StartedAt:          dto.StartedAt,
		CompletedAt:        dto.CompletedAt,
		UpdatedAt:          dto.UpdatedAt,
		WorkflowEfficiency: dto.WorkflowEfficiency,
		Totals:             dto.Totals,
		Roles:              dto.Roles,
		Models:             dto.Models,
		Anomalies:          anomalies,
	}
}

// SerializeLedger converts a domain AgenticMetricsLedger into an AgenticMetricsLedgerDTO.
func SerializeLedger(ledger domain.AgenticMetricsLedger) AgenticMetricsLedgerDTO {
	var anomalies []AnomalyDTO
	if len(ledger.Anomalies) > 0 {
		anomalies = make([]AnomalyDTO, len(ledger.Anomalies))
		for i, a := range ledger.Anomalies {
			anomalies[i] = AnomalyDTOFromDomain(a)
		}
	}
	return AgenticMetricsLedgerDTO{
		TicketID:           ledger.TicketID,
		TicketTitle:        ledger.TicketTitle,
		TicketType:         fromDomainTicketType(ledger.TicketType),
		StartedAt:          ledger.StartedAt,
		CompletedAt:        ledger.CompletedAt,
		UpdatedAt:          ledger.UpdatedAt,
		WorkflowEfficiency: ledger.WorkflowEfficiency,
		Totals:             ledger.Totals,
		Roles:              ledger.Roles,
		Models:             ledger.Models,
		Anomalies:          anomalies,
	}
}
