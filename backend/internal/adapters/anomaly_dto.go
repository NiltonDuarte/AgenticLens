package adapters

import (
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
)

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

type AnomalyEvidenceRefDTO struct {
	SessionID string `json:"session_id"`
	EventID   string `json:"event_id"`
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
