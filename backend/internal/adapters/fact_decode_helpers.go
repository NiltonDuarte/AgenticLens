package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func newStrictFactDecoder(data []byte) *json.Decoder {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder
}

func requireFactJSONEOF(decoder *json.Decoder, owner string) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values in %s", owner)
		}
		return fmt.Errorf("decoding trailing %s data: %w", owner, err)
	}
	return nil
}

// factCostMemberPresence observes only the declared cost fields of an invocation
// or task object; null is present and remains distinct from omission.
func factCostMemberPresence(data []byte) (amount, evidence, estimateNull bool, err error) {
	var fields struct {
		CostUSD      json.RawMessage `json:"cost_usd"`
		CostEvidence json.RawMessage `json:"cost_evidence"`
		CostEstimate json.RawMessage `json:"cost_estimate"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return false, false, false, err
	}
	isEstimateNull := len(fields.CostEstimate) > 0 && bytes.Equal(bytes.TrimSpace(fields.CostEstimate), []byte("null"))
	return len(fields.CostUSD) > 0, len(fields.CostEvidence) > 0, isEstimateNull, nil
}
