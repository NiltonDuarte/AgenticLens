package adapters

import (
	"agenticlens/internal/domain"
)

// AgenticMetricsLedgerDTO represents the computed metrics ledger format in JSON.
type AgenticMetricsLedgerDTO struct {
	TicketID           *string               `json:"ticket_id"`
	TicketTitle        *string               `json:"ticket_title"`
	TicketType         *string               `json:"ticket_type"`
	StartedAt          *string               `json:"started_at"`
	CompletedAt        *string               `json:"completed_at"`
	UpdatedAt          string                `json:"updated_at"`
	WorkflowEfficiency WorkflowEfficiencyDTO `json:"workflow_efficiency"`
	Totals             GroupedMetricDTO      `json:"totals"`
	Roles              []RoleMetricDTO       `json:"roles"`
	Models             []ModelMetricDTO      `json:"models"`
	Anomalies          []AnomalyDTO          `json:"anomalies,omitzero"`
}

// SerializeLedger converts a domain AgenticMetricsLedger into an AgenticMetricsLedgerDTO.
func SerializeLedger(ledger domain.AgenticMetricsLedger) AgenticMetricsLedgerDTO {
	anomalies := anomaliesFromDomain(ledger.Anomalies)
	return AgenticMetricsLedgerDTO{
		TicketID:           ledger.TicketID,
		TicketTitle:        ledger.TicketTitle,
		TicketType:         fromDomainTicketType(ledger.TicketType),
		StartedAt:          ledger.StartedAt,
		CompletedAt:        ledger.CompletedAt,
		UpdatedAt:          ledger.UpdatedAt,
		WorkflowEfficiency: WorkflowEfficiencyDTOFromDomain(ledger.WorkflowEfficiency),
		Totals:             GroupedMetricDTOFromDomain(ledger.Totals),
		Roles:              RoleMetricDTOsFromDomain(ledger.Roles),
		Models:             ModelMetricDTOsFromDomain(ledger.Models),
		Anomalies:          anomalies,
	}
}
