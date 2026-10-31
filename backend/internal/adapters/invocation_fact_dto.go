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
	Role                  string               `json:"role"`
	Model                 string               `json:"model"`
	StartedAt             *string              `json:"started_at"`
	CompletedAt           *string              `json:"completed_at"`
	ActiveDurationSeconds int64                `json:"active_duration_seconds"`
	Turns                 int64                `json:"turns"`
	Tokens                InvocationTokensDTO  `json:"tokens"`
	CostUSD               *float64             `json:"cost_usd"`
	CostEvidence          *domain.CostEvidence `json:"cost_evidence"`
	decodedCost           bool
	costPresent           bool
	evidencePresent       bool
	Compactions           int64 `json:"compactions"`
	UserIterations        int64 `json:"user_iterations"`
}

func (dto InvocationFactDTO) ToDomain() domain.InvocationFact {
	evidence := costFactEvidence(dto.CostUSD, dto.CostEvidence)
	return domain.InvocationFact{
		Role:                  domain.AgentRole(dto.Role),
		Model:                 dto.Model,
		StartedAt:             dto.StartedAt,
		CompletedAt:           dto.CompletedAt,
		ActiveDurationSeconds: dto.ActiveDurationSeconds,
		Turns:                 dto.Turns,
		Tokens:                dto.Tokens.ToDomain(),
		CostUSD:               costFactAccumulator(dto.CostUSD, evidence),
		CostEvidence:          evidence,
		Compactions:           dto.Compactions,
		UserIterations:        dto.UserIterations,
	}
}

// InvocationFactDTOFromDomain converts domain.InvocationFact to InvocationFactDTO.
func InvocationFactDTOFromDomain(inv domain.InvocationFact) InvocationFactDTO {
	evidence := domain.EffectiveCost(inv.CostUSD, inv.CostEvidence)
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
		CostUSD:        factCostAmount(inv.CostUSD, evidence),
		CostEvidence:   costEvidencePointer(evidence),
		Compactions:    inv.Compactions,
		UserIterations: inv.UserIterations,
	}
}

func (dto *InvocationFactDTO) UnmarshalJSON(data []byte) error {
	type invocationAlias InvocationFactDTO
	var value invocationAlias
	decoder := newStrictFactDecoder(data)
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if err := requireFactJSONEOF(decoder, "invocation"); err != nil {
		return err
	}
	amountPresent, evidencePresent, err := factCostMemberPresence(data)
	if err != nil {
		return err
	}
	*dto = InvocationFactDTO(value)
	dto.decodedCost = true
	dto.costPresent = amountPresent
	dto.evidencePresent = evidencePresent
	return domain.ValidateCostEvidence(dto.CostUSD, dto.CostEvidence, false)
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
