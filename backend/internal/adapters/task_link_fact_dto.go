package adapters

import (
	"agenticlens/internal/domain"
)

// TaskLinkFactDTO retains the presence of the optional task cost independently
// of its value. Task links never contribute to invocation aggregates.
type TaskLinkFactDTO struct {
	SubagentType    string               `json:"subagent_type"`
	Description     string               `json:"description"`
	ParentTurn      int                  `json:"parent_turn"`
	ChildSession    string               `json:"child_session_id,omitempty"`
	Timestamp       string               `json:"timestamp"`
	DurationMs      int64                `json:"duration_ms,omitempty"`
	CostUSD         *float64             `json:"cost_usd"`
	CostEvidence    *domain.CostEvidence `json:"cost_evidence"`
	decodedCost     bool
	costPresent     bool
	evidencePresent bool
}

func (dto *TaskLinkFactDTO) UnmarshalJSON(data []byte) error {
	type taskAlias TaskLinkFactDTO
	var value taskAlias
	decoder := newStrictFactDecoder(data)
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := requireFactJSONEOF(decoder, "task link"); err != nil {
		return err
	}
	amountPresent, evidencePresent, err := factCostMemberPresence(data)
	if err != nil {
		return err
	}
	*dto = TaskLinkFactDTO(value)
	dto.decodedCost = true
	dto.costPresent = amountPresent
	dto.evidencePresent = evidencePresent
	return domain.ValidateCostEvidence(dto.CostUSD, dto.CostEvidence, false)
}

func (dto TaskLinkFactDTO) ToDomain() domain.TaskLinkFact {
	evidence := costFactEvidence(dto.CostUSD, dto.CostEvidence)
	return domain.TaskLinkFact{SubagentType: dto.SubagentType, Description: dto.Description, ParentTurn: dto.ParentTurn, ChildSession: dto.ChildSession, Timestamp: dto.Timestamp, DurationMs: dto.DurationMs, CostUSD: costFactAccumulator(dto.CostUSD, evidence), CostEvidence: evidence}
}

func TaskLinkFactDTOFromDomain(value domain.TaskLinkFact) TaskLinkFactDTO {
	evidence := value.CostEvidence
	if evidence.State == "" {
		evidence = domain.UnavailableCost()
		if value.CostUSD != 0 {
			evidence = domain.ObservedCost(value.CostUSD)
		}
	}
	return TaskLinkFactDTO{SubagentType: value.SubagentType, Description: value.Description, ParentTurn: value.ParentTurn, ChildSession: value.ChildSession, Timestamp: value.Timestamp, DurationMs: value.DurationMs, CostUSD: factCostAmount(value.CostUSD, evidence), CostEvidence: costEvidencePointer(evidence)}
}

func TaskLinkFactDTOsFromDomain(values []domain.TaskLinkFact) []TaskLinkFactDTO {
	if values == nil {
		return nil
	}
	result := make([]TaskLinkFactDTO, len(values))
	for i, value := range values {
		result[i] = TaskLinkFactDTOFromDomain(value)
	}
	return result
}

func TaskLinkFactDTOsToDomain(values []TaskLinkFactDTO) []domain.TaskLinkFact {
	if values == nil {
		return nil
	}
	result := make([]domain.TaskLinkFact, len(values))
	for i, value := range values {
		result[i] = value.ToDomain()
	}
	return result
}
