package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

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

func factCostAmount(amount float64, evidence domain.CostEvidence) *float64 {
	if evidence.State == domain.CostObserved {
		return domain.CostNumber(amount)
	}
	return nil
}
func costFactEvidence(amount *float64, evidence *domain.CostEvidence) domain.CostEvidence {
	if evidence != nil {
		return *evidence
	}
	if amount != nil {
		return domain.ObservedCost(*amount)
	}
	return domain.UnavailableCost()
}
func costFactAccumulator(amount *float64, evidence domain.CostEvidence) float64 {
	if evidence.State == domain.CostPartial && evidence.KnownSubtotalUSD != nil {
		return *evidence.KnownSubtotalUSD
	}
	if amount != nil {
		return *amount
	}
	return 0
}
func (dto InvocationFactDTO) evidence() domain.CostEvidence {
	return costFactEvidence(dto.CostUSD, dto.CostEvidence)
}
func (dto InvocationFactDTO) internalCost() float64 {
	return costFactAccumulator(dto.CostUSD, dto.evidence())
}

func (dto *InvocationFactDTO) UnmarshalJSON(data []byte) error {
	type invocationAlias InvocationFactDTO
	var value invocationAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values in invocation")
		}
		return err
	}
	var presence struct {
		CostUSD      json.RawMessage `json:"cost_usd"`
		CostEvidence json.RawMessage `json:"cost_evidence"`
	}
	if err := json.Unmarshal(data, &presence); err != nil {
		return err
	}
	*dto = InvocationFactDTO(value)
	dto.decodedCost = true
	dto.costPresent = len(presence.CostUSD) > 0
	dto.evidencePresent = len(presence.CostEvidence) > 0
	return domain.ValidateCostEvidence(dto.CostUSD, dto.CostEvidence, false)
}
func (dto *TaskLinkFactDTO) UnmarshalJSON(data []byte) error {
	type taskAlias TaskLinkFactDTO
	var value taskAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values in task link")
		}
		return err
	}
	var presence struct {
		CostUSD      json.RawMessage `json:"cost_usd"`
		CostEvidence json.RawMessage `json:"cost_evidence"`
	}
	if err := json.Unmarshal(data, &presence); err != nil {
		return err
	}
	*dto = TaskLinkFactDTO(value)
	dto.decodedCost = true
	dto.costPresent = len(presence.CostUSD) > 0
	dto.evidencePresent = len(presence.CostEvidence) > 0
	return domain.ValidateCostEvidence(dto.CostUSD, dto.CostEvidence, false)
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
func InvocationFactDTOsFromDomain(values []domain.InvocationFact) []InvocationFactDTO {
	if values == nil {
		return nil
	}
	result := make([]InvocationFactDTO, len(values))
	for i, value := range values {
		result[i] = InvocationFactDTOFromDomain(value)
	}
	return result
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
