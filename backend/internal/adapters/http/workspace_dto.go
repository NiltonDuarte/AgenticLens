package http

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

// ThreadRoleMetricDTO is the concrete nullable cost projection of ThreadRoleMetric.
type ThreadRoleMetricDTO struct {
	Role   domain.AgentRole          `json:"role"`
	Models []adapters.ModelMetricDTO `json:"models"`
}

func ThreadRoleMetricDTOFromDomain(value domain.ThreadRoleMetric) ThreadRoleMetricDTO {
	return ThreadRoleMetricDTO{
		Role:   value.Role,
		Models: adapters.ModelMetricDTOsFromDomain(value.Models),
	}
}

func ThreadRoleMetricDTOsFromDomain(values []domain.ThreadRoleMetric) []ThreadRoleMetricDTO {
	if values == nil {
		return nil
	}
	result := make([]ThreadRoleMetricDTO, len(values))
	for i, value := range values {
		result[i] = ThreadRoleMetricDTOFromDomain(value)
	}
	return result
}

// ThreadExecutionRecordDTO is the concrete nullable cost projection of ThreadExecutionRecord.
type ThreadExecutionRecordDTO struct {
	SessionID   string                `json:"session_id"`
	User        string                `json:"user"`
	Roles       []ThreadRoleMetricDTO `json:"roles"`
	Compactions int64                 `json:"compactions"`
}

func ThreadExecutionRecordDTOFromDomain(value domain.ThreadExecutionRecord) ThreadExecutionRecordDTO {
	return ThreadExecutionRecordDTO{
		SessionID:   value.SessionID,
		User:        value.User,
		Roles:       ThreadRoleMetricDTOsFromDomain(value.Roles),
		Compactions: value.Compactions,
	}
}

func ThreadExecutionRecordDTOsFromDomain(values []domain.ThreadExecutionRecord) []ThreadExecutionRecordDTO {
	if values == nil {
		return nil
	}
	result := make([]ThreadExecutionRecordDTO, len(values))
	for i, value := range values {
		result[i] = ThreadExecutionRecordDTOFromDomain(value)
	}
	return result
}

// SessionFactResponseDTO is the concrete nullable cost projection of SessionFact.
type SessionFactResponseDTO struct {
	ID          string                       `json:"id"`
	ParentID    *string                      `json:"parent_id,omitempty"`
	Source      domain.SourceHarness         `json:"source,omitempty"`
	User        string                       `json:"user,omitempty"`
	Invocations []adapters.InvocationFactDTO `json:"invocations"`
	Tools       []domain.SessionToolFact     `json:"tools"`
	Events      []domain.SessionEventFact    `json:"events,omitempty"`
	Anomalies   []domain.Anomaly             `json:"anomalies"`
}

func SessionFactResponseDTOFromDomain(value domain.SessionFact) SessionFactResponseDTO {
	return SessionFactResponseDTO{
		ID:          value.ID,
		ParentID:    value.ParentID,
		Source:      value.Source,
		User:        value.User,
		Invocations: adapters.InvocationFactDTOsFromDomain(value.Invocations),
		Tools:       value.Tools,
		Events:      value.Events,
		Anomalies:   value.Anomalies,
	}
}

func SessionFactResponseDTOsFromDomain(values []domain.SessionFact) []SessionFactResponseDTO {
	if values == nil {
		return nil
	}
	result := make([]SessionFactResponseDTO, len(values))
	for i, value := range values {
		result[i] = SessionFactResponseDTOFromDomain(value)
	}
	return result
}

// FactLedgerResponseDTO is the concrete nullable cost projection of AgenticFactLedger.
type FactLedgerResponseDTO struct {
	SchemaVersion *int                           `json:"schema_version,omitempty"`
	Migration     *domain.MigrationInfo          `json:"migration,omitempty"`
	TicketID      *string                        `json:"ticket_id"`
	TicketTitle   *string                        `json:"ticket_title"`
	TicketType    *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt     *string                        `json:"started_at"`
	CompletedAt   *string                        `json:"completed_at"`
	UpdatedAt     string                         `json:"updated_at"`
	Code          domain.CodeFacts               `json:"code"`
	Sessions      []SessionFactResponseDTO       `json:"sessions"`
	Anomalies     []domain.Anomaly               `json:"anomalies"`
	Skills        []domain.SkillFact             `json:"skills,omitempty"`
	TaskLinks     []adapters.TaskLinkFactDTO     `json:"task_links,omitempty"`
}

func FactLedgerResponseDTOFromDomain(value domain.AgenticFactLedger) FactLedgerResponseDTO {
	return FactLedgerResponseDTO{
		SchemaVersion: value.SchemaVersion,
		Migration:     value.Migration,
		TicketID:      value.TicketID,
		TicketTitle:   value.TicketTitle,
		TicketType:    value.TicketType,
		StartedAt:     value.StartedAt,
		CompletedAt:   value.CompletedAt,
		UpdatedAt:     value.UpdatedAt,
		Code:          value.Code,
		Sessions:      SessionFactResponseDTOsFromDomain(value.Sessions),
		Anomalies:     value.Anomalies,
		Skills:        value.Skills,
		TaskLinks:     adapters.TaskLinkFactDTOsFromDomain(value.TaskLinks),
	}
}

// MetricsLedgerResponseDTO is the concrete nullable cost projection of AgenticMetricsLedger.
type MetricsLedgerResponseDTO struct {
	TicketID           *string                        `json:"ticket_id"`
	TicketTitle        *string                        `json:"ticket_title"`
	TicketType         *domain.ConventionalTicketType `json:"ticket_type"`
	StartedAt          *string                        `json:"started_at"`
	CompletedAt        *string                        `json:"completed_at"`
	UpdatedAt          string                         `json:"updated_at"`
	Facts              FactLedgerResponseDTO          `json:"facts"`
	WorkflowEfficiency adapters.WorkflowEfficiencyDTO `json:"workflow_efficiency"`
	Threads            []ThreadExecutionRecordDTO     `json:"threads"`
	Totals             adapters.GroupedMetricDTO      `json:"totals"`
	Roles              []adapters.RoleMetricDTO       `json:"roles"`
	Models             []adapters.ModelMetricDTO      `json:"models"`
	Anomalies          []domain.Anomaly               `json:"anomalies,omitempty"`
}

func MetricsLedgerResponseDTOFromDomain(value domain.AgenticMetricsLedger) MetricsLedgerResponseDTO {
	return MetricsLedgerResponseDTO{
		TicketID:           value.TicketID,
		TicketTitle:        value.TicketTitle,
		TicketType:         value.TicketType,
		StartedAt:          value.StartedAt,
		CompletedAt:        value.CompletedAt,
		UpdatedAt:          value.UpdatedAt,
		Facts:              FactLedgerResponseDTOFromDomain(value.Facts),
		WorkflowEfficiency: adapters.WorkflowEfficiencyDTOFromDomain(value.WorkflowEfficiency),
		Threads:            ThreadExecutionRecordDTOsFromDomain(value.Threads),
		Totals:             adapters.GroupedMetricDTOFromDomain(value.Totals),
		Roles:              adapters.RoleMetricDTOsFromDomain(value.Roles),
		Models:             adapters.ModelMetricDTOsFromDomain(value.Models),
		Anomalies:          value.Anomalies,
	}
}

func MetricsLedgerResponseDTOsFromDomain(values []domain.AgenticMetricsLedger) []MetricsLedgerResponseDTO {
	if values == nil {
		return nil
	}
	result := make([]MetricsLedgerResponseDTO, len(values))
	for i, value := range values {
		result[i] = MetricsLedgerResponseDTOFromDomain(value)
	}
	return result
}
