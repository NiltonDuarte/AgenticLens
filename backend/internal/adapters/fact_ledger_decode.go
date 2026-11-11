package adapters

import (
	"encoding/json"
	"fmt"
)

// canonicalFactLedgerDTO decodes the current schema only. Legacy inputs belong
// exclusively to the explicit versioned migration engine.
type canonicalFactLedgerDTO AgenticFactLedgerDTO

func decodeCanonicalFactLedger(data []byte, requiredCostEvidence bool) (AgenticFactLedgerDTO, error) {
	if err := validateCurrentOptionalCollections(data); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	var wire canonicalFactLedgerDTO
	decoder := newStrictFactDecoder(data)
	if err := decoder.Decode(&wire); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	if err := requireFactJSONEOF(decoder, "fact ledger"); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	result := AgenticFactLedgerDTO(wire)
	if err := result.validateCostFields(requiredCostEvidence); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	return result, nil
}

// validateCurrentOptionalCollections validates current wire presence, independently
// of explicit historical metadata cleanup. Null never means an unrecorded array.
func validateCurrentOptionalCollections(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	if err := validateOptionalCategory(root, "anomalies"); err != nil {
		return err
	}
	var sessions []map[string]json.RawMessage
	if raw, exists := root["sessions"]; exists {
		if err := json.Unmarshal(raw, &sessions); err != nil {
			return err
		}
	}
	for i, owner := range sessions {
		for _, key := range []string{"skills", "delegations", "anomalies"} {
			if err := validateOptionalCategory(owner, key); err != nil {
				return fmt.Errorf("sessions[%d]: %w", i, err)
			}
		}
	}
	return nil
}
