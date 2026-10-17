package adapters

import (
	"agenticlens/internal/domain"
)

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
	ID                    string              `json:"id"`
	ParentID              *string             `json:"parent_id"`
	Role                  string              `json:"role"`
	Model                 string              `json:"model"`
	User                  string              `json:"user"`
	StartedAt             *string             `json:"started_at"`
	CompletedAt           *string             `json:"completed_at"`
	ActiveDurationSeconds int64               `json:"active_duration_seconds"`
	Turns                 int64               `json:"turns"`
	Tokens                InvocationTokensDTO `json:"tokens"`
	CostUSD               float64             `json:"cost_usd"`
	Compactions           int64               `json:"compactions"`
}

func (dto InvocationFactDTO) ToDomain() domain.InvocationFact {
	role := dto.Role
	if role == "" {
		role = "orchestrator"
	}
	model := dto.Model
	if model == "" {
		model = "default"
	}
	user := dto.User
	if user == "" {
		user = "unknown"
	}
	return domain.InvocationFact{
		ID:                    dto.ID,
		ParentID:              dto.ParentID,
		Role:                  role,
		Model:                 model,
		User:                  user,
		StartedAt:             dto.StartedAt,
		CompletedAt:           dto.CompletedAt,
		ActiveDurationSeconds: dto.ActiveDurationSeconds,
		Turns:                 dto.Turns,
		Tokens:                dto.Tokens.ToDomain(),
		CostUSD:               dto.CostUSD,
		Compactions:           dto.Compactions,
	}
}

// ToolFactDTO represents a single tool execution fact in JSON.
type ToolFactDTO struct {
	Name   string `json:"name"`
	Calls  int64  `json:"calls"`
	Errors int64  `json:"errors"`
}

func (dto ToolFactDTO) ToDomain() domain.ToolFact {
	name := dto.Name
	if name == "" {
		name = "unknown"
	}
	return domain.ToolFact{
		Name:   name,
		Calls:  dto.Calls,
		Errors: dto.Errors,
	}
}

// CodeFactsDTO represents code facts in JSON.
type CodeFactsDTO struct {
	LinesAdded      int64 `json:"lines_added"`
	LinesDeleted    int64 `json:"lines_deleted"`
	TestLinesAdded  int64 `json:"test_lines_added"`
	TestLinesDeleted int64 `json:"test_lines_deleted"`
	ProdLinesAdded  int64 `json:"prod_lines_added"`
	ProdLinesDeleted int64 `json:"prod_lines_deleted"`
	FilesChanged    int64 `json:"files_changed"`
	SDLCIterations  int64 `json:"sdlc_iterations"`
}

func (dto CodeFactsDTO) ToDomain() domain.CodeFacts {
	sdlc := dto.SDLCIterations
	if sdlc <= 0 {
		sdlc = 1
	}
	return domain.CodeFacts{
		LinesAdded:      dto.LinesAdded,
		LinesDeleted:    dto.LinesDeleted,
		TestLinesAdded:  dto.TestLinesAdded,
		TestLinesDeleted: dto.TestLinesDeleted,
		ProdLinesAdded:  dto.ProdLinesAdded,
		ProdLinesDeleted: dto.ProdLinesDeleted,
		FilesChanged:    dto.FilesChanged,
		SDLCIterations:  sdlc,
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
	SchemaVersion *int                `json:"schema_version,omitempty"`
	Migration     *MigrationInfoDTO   `json:"migration,omitempty"`
	TicketID      *string             `json:"ticket_id"`
	TicketTitle   *string             `json:"ticket_title"`
	TicketType    *string             `json:"ticket_type"`
	StartedAt     *string             `json:"started_at"`
	CompletedAt   *string             `json:"completed_at"`
	UpdatedAt     string              `json:"updated_at"`
	Code          CodeFactsDTO        `json:"code"`
	Invocations   []InvocationFactDTO `json:"invocations"`
	Tools         []ToolFactDTO       `json:"tools"`
}

func (dto AgenticFactLedgerDTO) ToDomain() domain.AgenticFactLedger {
	invs := make([]domain.InvocationFact, len(dto.Invocations))
	for i, inv := range dto.Invocations {
		invs[i] = inv.ToDomain()
	}
	tools := make([]domain.ToolFact, len(dto.Tools))
	for i, t := range dto.Tools {
		tools[i] = t.ToDomain()
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
		TicketType:    dto.TicketType,
		StartedAt:     dto.StartedAt,
		CompletedAt:   dto.CompletedAt,
		UpdatedAt:     dto.UpdatedAt,
		Code:          dto.Code.ToDomain(),
		Invocations:   invs,
		Tools:         tools,
	}
}

// AgenticFactLedgerDTOFromDomain converts a domain AgenticFactLedger to its DTO for serialization.
func AgenticFactLedgerDTOFromDomain(ledger domain.AgenticFactLedger) AgenticFactLedgerDTO {
	invs := make([]InvocationFactDTO, len(ledger.Invocations))
	for i, inv := range ledger.Invocations {
		invs[i] = InvocationFactDTO{
			ID:                    inv.ID,
			ParentID:              inv.ParentID,
			Role:                  inv.Role,
			Model:                 inv.Model,
			User:                  inv.User,
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
			CostUSD:     inv.CostUSD,
			Compactions: inv.Compactions,
		}
	}

	tools := make([]ToolFactDTO, len(ledger.Tools))
	for i, t := range ledger.Tools {
		tools[i] = ToolFactDTO{
			Name:   t.Name,
			Calls:  t.Calls,
			Errors: t.Errors,
		}
	}

	return AgenticFactLedgerDTO{
		SchemaVersion: ledger.SchemaVersion,
		Migration:     MigrationInfoDTOFromDomain(ledger.Migration),
		TicketID:      ledger.TicketID,
		TicketTitle:   ledger.TicketTitle,
		TicketType:    ledger.TicketType,
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
		Invocations: invs,
		Tools:       tools,
	}
}

// SerializedLedgerDTO represents the JSON output format produced by the CLI.
type SerializedLedgerDTO struct {
	TicketID           *string                    `json:"ticket_id"`
	TicketTitle        *string                    `json:"ticket_title"`
	TicketType         *string                    `json:"ticket_type"`
	StartedAt          *string                    `json:"started_at"`
	CompletedAt        *string                    `json:"completed_at"`
	UpdatedAt          string                     `json:"updated_at"`
	WorkflowEfficiency domain.WorkflowEfficiency  `json:"workflow_efficiency"`
	Totals             domain.GroupedMetric       `json:"totals"`
	Roles              []domain.RoleMetric        `json:"roles"`
	Models             []domain.ModelMetric       `json:"models"`
}

// SerializeLedger converts a domain AgenticMetricsLedger into a SerializedLedgerDTO.
func SerializeLedger(ledger domain.AgenticMetricsLedger) SerializedLedgerDTO {
	return SerializedLedgerDTO{
		TicketID:           ledger.TicketID,
		TicketTitle:        ledger.TicketTitle,
		TicketType:         ledger.TicketType,
		StartedAt:          ledger.StartedAt,
		CompletedAt:        ledger.CompletedAt,
		UpdatedAt:          ledger.UpdatedAt,
		WorkflowEfficiency: ledger.WorkflowEfficiency,
		Totals:             ledger.Totals,
		Roles:              ledger.Roles,
		Models:             ledger.Models,
	}
}
