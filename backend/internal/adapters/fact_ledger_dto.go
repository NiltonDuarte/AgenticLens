package adapters

import (
	"agenticlens/internal/domain"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

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
	SchemaVersion *int               `json:"schema_version,omitempty"`
	Source        string             `json:"source,omitempty"`
	Migration     *MigrationInfoDTO  `json:"migration,omitempty"`
	TicketID      *string            `json:"ticket_id"`
	TicketTitle   *string            `json:"ticket_title"`
	TicketType    *string            `json:"ticket_type"`
	StartedAt     *string            `json:"started_at"`
	CompletedAt   *string            `json:"completed_at"`
	UpdatedAt     string             `json:"updated_at"`
	Code          CodeFactsDTO       `json:"code"`
	Sessions      []SessionFactDTO   `json:"sessions"`
	Anomalies     []AnomalyDTO       `json:"anomalies"`
	Skills        []domain.SkillFact `json:"skills,omitempty"`
	TaskLinks     []TaskLinkFactDTO  `json:"task_links,omitempty"`
}

// UnmarshalJSON rejects legacy root fact collections at the current schema boundary.
// They must be passed through the versioned migration engine instead.
func (dto *AgenticFactLedgerDTO) UnmarshalJSON(data []byte) error {
	var version struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}
	if version.SchemaVersion != nil && (*version.SchemaVersion < 1 || *version.SchemaVersion > 5) {
		return fmt.Errorf("unsupported canonical fact schema version %d", *version.SchemaVersion)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, legacy := range []string{"invocations", "tools", "events"} {
		if _, exists := fields[legacy]; exists {
			return fmt.Errorf("legacy root-level %s is not valid in session-owned schema v4; run `agentic-lens migrate`", legacy)
		}
	}
	var decoded AgenticFactLedgerDTO
	var err error
	if version.SchemaVersion != nil && *version.SchemaVersion == 5 {
		decoded, err = decodeSchemaV5FactLedger(data)
	} else {
		// Older canonical shapes can be decoded by explicit migration tooling. The
		// ordinary runtime gate accepts only canonical v4 and v5.
		decoded, err = decodeSchemaV4FactLedger(data)
	}
	if err != nil {
		return err
	}
	*dto = decoded
	return nil
}

func (dto AgenticFactLedgerDTO) ToDomain() domain.AgenticFactLedger {
	sessions := make([]domain.SessionFact, len(dto.Sessions))
	for i, session := range dto.Sessions {
		sessions[i] = session.ToDomain()
		if sessions[i].Source == "" && dto.Source != "" {
			sessions[i].Source = domain.SourceHarness(strings.TrimSpace(dto.Source))
		}
	}

	var mig *domain.MigrationInfo
	if dto.Migration != nil {
		domMig := dto.Migration.ToDomain()
		mig = &domMig
	}

	skills := dto.Skills
	if skills == nil {
		skills = []domain.SkillFact{}
	}
	taskLinks := TaskLinkFactDTOsToDomain(dto.TaskLinks)
	if taskLinks == nil {
		taskLinks = []domain.TaskLinkFact{}
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
		Skills:        skills,
		TaskLinks:     taskLinks,
	}
}

// Validate enforces required fields at the persisted telemetry boundary.
func (dto AgenticFactLedgerDTO) Validate() error {
	currentSchema := dto.SchemaVersion != nil && *dto.SchemaVersion >= 4
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
	return dto.validateCostFields(dto.SchemaVersion != nil && *dto.SchemaVersion >= 5)
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

// AgenticFactLedgerDTOFromDomain converts a domain AgenticFactLedger to its DTO for serialization.
func AgenticFactLedgerDTOFromDomain(ledger domain.AgenticFactLedger) AgenticFactLedgerDTO {
	sessions := make([]SessionFactDTO, len(ledger.Sessions))
	for i, session := range ledger.Sessions {
		sessions[i] = SessionFactDTOFromDomain(session)
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
		Skills:    ledger.Skills,
		TaskLinks: TaskLinkFactDTOsFromDomain(ledger.TaskLinks),
	}
}

func (dto AgenticFactLedgerDTO) validateCostFields(required bool) error {
	for si, session := range dto.Sessions {
		for ii, inv := range session.Invocations {
			if required && inv.decodedCost && (!inv.costPresent || !inv.evidencePresent) {
				return fmt.Errorf("sessions[%d].invocations[%d] requires cost_usd and cost_evidence", si, ii)
			}
			if err := domain.ValidateCostEvidence(inv.CostUSD, inv.CostEvidence, required); err != nil {
				return fmt.Errorf("sessions[%d].invocations[%d]: %w", si, ii, err)
			}
		}
	}
	for ti, task := range dto.TaskLinks {
		if required && task.decodedCost && (!task.costPresent || !task.evidencePresent) {
			return fmt.Errorf("task_links[%d] requires cost_usd and cost_evidence", ti)
		}
		if err := domain.ValidateCostEvidence(task.CostUSD, task.CostEvidence, required); err != nil {
			return fmt.Errorf("task_links[%d]: %w", ti, err)
		}
	}
	return nil
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
		return errors.New("ledger schema version is missing; run `agentic-lens migrate --target-version 5` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`")
	}
	if *metadata.SchemaVersion != 4 && *metadata.SchemaVersion != domain.CurrentSchemaVersion {
		return fmt.Errorf("ledger schema version %d is not runtime-ingestible; run `agentic-lens migrate --target-version 5` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`", *metadata.SchemaVersion)
	}
	for _, key := range []string{"invocations", "tools", "events"} {
		if _, exists := fields[key]; exists {
			return fmt.Errorf("schema v4 ledger contains ambiguous root-level %s; run `agentic-lens migrate --target-version 5` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`", key)
		}
	}
	sessionsJSON, sessionsExists := fields["sessions"]
	if !sessionsExists || string(sessionsJSON) == "null" || !json.Valid(sessionsJSON) {
		return errors.New("canonical schema v4/v5 ledger requires a sessions collection; run `agentic-lens migrate --target-version 5` followed by source-backed `agentic-lens backfill --recompute-db --source <opencode|codex>`")
	}
	if metadata.Migration != nil && (metadata.Migration.Status != string(domain.MigrationStatusCompleted) || len(metadata.Migration.StepsPending) > 0) {
		return errors.New("schema v4/v5 ledger has incomplete migration metadata; resume the explicit migration and source-backed backfill workflow before runtime ingestion")
	}
	var canonical struct {
		Sessions []json.RawMessage `json:"sessions"`
	}
	if err := json.Unmarshal(sessionsJSON, &canonical.Sessions); err != nil || canonical.Sessions == nil {
		return errors.New("canonical schema v4/v5 sessions must be an array")
	}
	for si, rawSession := range canonical.Sessions {
		var session map[string]json.RawMessage
		if err := json.Unmarshal(rawSession, &session); err != nil {
			return fmt.Errorf("canonical schema v4/v5 session %d is invalid", si)
		}
		for _, required := range []string{"id", "source", "user", "invocations", "tools", "anomalies"} {
			if _, ok := session[required]; !ok {
				return fmt.Errorf("canonical schema v4/v5 session %d is missing required %s", si, required)
			}
		}
		var sessionUser string
		if err := json.Unmarshal(session["user"], &sessionUser); err != nil || strings.TrimSpace(sessionUser) == "" {
			return fmt.Errorf("canonical schema v4/v5 session %d requires a non-empty user", si)
		}
		for _, collection := range []string{"tools", "anomalies"} {
			var values []json.RawMessage
			if err := json.Unmarshal(session[collection], &values); err != nil || values == nil {
				return fmt.Errorf("canonical schema v4/v5 session %d %s must be an array", si, collection)
			}
		}
		var source string
		if err := json.Unmarshal(session["source"], &source); err != nil || (source != string(domain.SourceOpenCode) && source != string(domain.SourceCodex)) {
			return fmt.Errorf("canonical schema v4/v5 session %d requires a supported source (opencode or codex)", si)
		}
		var invocations []json.RawMessage
		if err := json.Unmarshal(session["invocations"], &invocations); err != nil || invocations == nil {
			return fmt.Errorf("canonical schema v4/v5 session %d invocations must be an array", si)
		}
		for ii, rawInvocation := range invocations {
			var invocation map[string]json.RawMessage
			if err := json.Unmarshal(rawInvocation, &invocation); err != nil {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d is invalid", ii, si)
			}
			if _, ok := invocation["id"]; ok {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d must not contain id", ii, si)
			}
			if _, ok := invocation["user"]; ok {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d must not contain user", ii, si)
			}
			if _, ok := invocation["parent_id"]; ok {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d must not contain parent_id", ii, si)
			}
			for _, required := range []string{"role", "model", "started_at", "completed_at", "active_duration_seconds", "turns", "tokens", "compactions", "user_iterations"} {
				if _, ok := invocation[required]; !ok {
					return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d is missing required %s", ii, si, required)
				}
			}
			var tokens map[string]json.RawMessage
			if err := json.Unmarshal(invocation["tokens"], &tokens); err != nil || tokens == nil {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d requires token counts", ii, si)
			}
			for _, required := range []string{"input", "output", "cache_read", "reasoning", "peak_context"} {
				if _, ok := tokens[required]; !ok {
					return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d is missing token count %s", ii, si, required)
				}
			}
			value, ok := invocation["user_iterations"]
			var count int64
			if !ok || json.Unmarshal(value, &count) != nil || count < 0 {
				return fmt.Errorf("canonical schema v4/v5 invocation %d in session %d requires a non-negative numeric user_iterations (zero is valid)", ii, si)
			}
		}
	}
	var decoded AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	return decoded.validateCostFields(*metadata.SchemaVersion >= 5)
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
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
