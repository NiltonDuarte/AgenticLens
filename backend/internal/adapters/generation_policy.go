package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"

	"agenticlens/internal/domain"
)

// ProjectGenerationDTO is the final output transform, after retained owners have
// passed their original identity/equality checks. It never changes caller slices.
func ProjectGenerationDTO(dto AgenticFactLedgerDTO) AgenticFactLedgerDTO {
	if dto.GenerationPolicy == nil {
		return dto
	}
	policy := *dto.GenerationPolicy
	dto.Sessions = append([]SessionFactDTO{}, dto.Sessions...)
	if !policy.Anomalies {
		dto.Anomalies = nil
	}
	for i := range dto.Sessions {
		if policy.Anomalies && dto.Sessions[i].Anomalies == nil {
			dto.Sessions[i].AnomalyAvailability = domain.AnomalyUnavailable
		}
		if !policy.Skills {
			dto.Sessions[i].Skills = nil
		}
		if !policy.Delegations {
			dto.Sessions[i].Delegations = nil
		}
		if !policy.Anomalies || dto.Sessions[i].AnomalyAvailability == domain.AnomalyUnavailable {
			dto.Sessions[i].Anomalies = nil
		}
	}
	return dto
}

// MarshalJSON preserves enabled empty anomaly arrays while intentionally omitting
// disabled root and session analysis. Policy metadata explains that absence.
func (dto AgenticFactLedgerDTO) MarshalJSON() ([]byte, error) {
	type wire AgenticFactLedgerDTO
	dto = ProjectGenerationDTO(dto)
	data, err := json.Marshal(wire(dto))
	if err != nil {
		return nil, err
	}
	return OmitUnrecordedAnomaliesJSON(data, dto.GenerationPolicy)
}

// OmitUnrecordedAnomaliesJSON applies the same additive omission contract to HTTP
// fact projections, including unavailable owners after re-enabling. It touches
// only root/session category fields, never raw text.
func OmitUnrecordedAnomaliesJSON(data []byte, policy *domain.GenerationPolicy) ([]byte, error) {

	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	disabled := policy != nil && !policy.Anomalies
	if disabled {
		delete(root, "anomalies")
	}
	var sessions []map[string]json.RawMessage
	if err := json.Unmarshal(root["sessions"], &sessions); err != nil {
		return nil, err
	}
	changed := disabled
	for _, session := range sessions {
		if disabled || string(session["anomaly_availability"]) == `"unavailable"` {
			delete(session, "anomalies")
			changed = true
		}
	}
	if !changed {
		return data, nil
	}
	encoded, err := json.Marshal(sessions)
	if err != nil {
		return nil, err
	}
	root["sessions"] = encoded
	return json.Marshal(root)
}

func validateGenerationPolicyJSON(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	raw, exists := root["generation_policy"]
	if !exists {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return fmt.Errorf("generation_policy must be an object containing skills, delegations and anomalies booleans")
	}
	for _, key := range []string{"skills", "delegations", "anomalies"} {
		value := bytes.TrimSpace(fields[key])
		if !bytes.Equal(value, []byte("true")) && !bytes.Equal(value, []byte("false")) {
			return fmt.Errorf("generation_policy.%s must be a boolean", key)
		}
	}
	return nil
}
