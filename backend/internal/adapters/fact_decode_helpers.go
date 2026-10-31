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
func factCostMemberPresence(data []byte) (amount, evidence bool, err error) {
	var fields struct {
		CostUSD      json.RawMessage `json:"cost_usd"`
		CostEvidence json.RawMessage `json:"cost_evidence"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return false, false, err
	}
	return len(fields.CostUSD) > 0, len(fields.CostEvidence) > 0, nil
}
