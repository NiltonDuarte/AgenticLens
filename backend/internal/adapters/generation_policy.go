package adapters

import (
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
	"fmt"
)

func validateOptionalCategory(fields map[string]json.RawMessage, key string) error {
	raw, exists := fields[key]
	if !exists {
		return nil
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil || rows == nil {
		return fmt.Errorf("%s must be an array when recorded; omit it when not recorded", key)
	}
	return nil
}

// normalizeFormerGenerationMetadata accepts only the previously supported AL-66
// metadata. It validates contradictions, then removes the reasons without
// synthesizing any category arrays. New writes never include these members.
func normalizeFormerGenerationMetadata(data []byte) ([]byte, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	changed := false
	var policy *domain.GenerationPolicy
	if raw, exists := root["generation_policy"]; exists {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || len(fields) != 3 {
			return nil, fmt.Errorf("generation_policy requires exactly skills, delegations and anomalies booleans")
		}
		for _, key := range []string{"skills", "delegations", "anomalies"} {
			value := bytes.TrimSpace(fields[key])
			if !bytes.Equal(value, []byte("true")) && !bytes.Equal(value, []byte("false")) {
				return nil, fmt.Errorf("generation_policy.%s must be a boolean", key)
			}
		}
		if err := json.Unmarshal(raw, &policy); err != nil {
			return nil, err
		}
		if !policy.Anomalies {
			if _, exists := root["anomalies"]; exists {
				return nil, fmt.Errorf("former disabled anomaly policy requires omitted root anomalies")
			}
		}
		delete(root, "generation_policy")
		changed = true
	}
	if err := validateOptionalCategory(root, "anomalies"); err != nil {
		return nil, err
	}
	var sessions []map[string]json.RawMessage
	if raw, exists := root["sessions"]; exists {
		if err := json.Unmarshal(raw, &sessions); err != nil {
			return nil, err
		}
		for i, owner := range sessions {
			for _, key := range []string{"skills", "delegations", "anomalies"} {
				if err := validateOptionalCategory(owner, key); err != nil {
					return nil, fmt.Errorf("sessions[%d]: %w", i, err)
				}
			}
			if raw, exists := owner["anomaly_availability"]; exists {
				var availability string
				if err := json.Unmarshal(raw, &availability); err != nil || availability != "unavailable" {
					return nil, fmt.Errorf("sessions[%d].anomaly_availability must be unavailable", i)
				}
				if _, exists := owner["anomalies"]; exists {
					return nil, fmt.Errorf("sessions[%d] unavailable analysis requires omitted anomalies", i)
				}
				delete(owner, "anomaly_availability")
				changed = true
			}
			if policy != nil {
				for key, enabled := range map[string]bool{"skills": policy.Skills, "delegations": policy.Delegations, "anomalies": policy.Anomalies} {
					if _, exists := owner[key]; exists && !enabled {
						return nil, fmt.Errorf("sessions[%d].%s contradicts former disabled policy", i, key)
					}
				}
			}
		}
	}
	if !changed {
		return data, nil
	}
	if sessions != nil {
		encoded, err := json.Marshal(sessions)
		if err != nil {
			return nil, err
		}
		root["sessions"] = encoded
	}
	return json.Marshal(root)
}

// NormalizeFormerGenerationMetadata validates original JSON members before
// removing supported transitional metadata. Only explicit migrate/backfill preparation invokes this cleanup; ordinary reads
// require current metadata without normalization.
func NormalizeFormerGenerationMetadata(data []byte) ([]byte, error) {
	if _, err := FormatLedgerJSON(data); err != nil {
		return nil, fmt.Errorf("validating original ledger JSON: %w", err)
	}
	return normalizeFormerGenerationMetadata(data)
}
