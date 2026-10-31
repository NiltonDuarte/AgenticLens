package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// orderedCostObject retains object member order while cost fields are projected.
// Raw values preserve numeric literals and avoid decoding unrelated payloads.
type orderedCostObject struct {
	keys   []string
	fields map[string]json.RawMessage
}

func decodeOrderedCostObject(data []byte) (orderedCostObject, error) {
	result := orderedCostObject{fields: make(map[string]json.RawMessage)}
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil {
		return result, err
	}
	if opening != json.Delim('{') {
		return result, fmt.Errorf("cost JSON object is required")
	}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return result, err
		}
		key, ok := token.(string)
		if !ok {
			return result, fmt.Errorf("cost JSON object member name is required")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return result, err
		}
		if _, exists := result.fields[key]; !exists {
			result.keys = append(result.keys, key)
		}
		result.fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return result, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return result, fmt.Errorf("multiple JSON values in cost object")
		}
		return result, err
	}
	return result, nil
}

func (object *orderedCostObject) insertMissingKey(key string, after ...string) {
	if _, exists := object.fields[key]; !exists {
		return
	}
	for _, existing := range object.keys {
		if existing == key {
			return
		}
	}
	for _, anchor := range after {
		for i, existing := range object.keys {
			if existing == anchor {
				object.keys = append(object.keys, "")
				copy(object.keys[i+2:], object.keys[i+1:])
				object.keys[i+1] = key
				return
			}
		}
	}
	object.keys = append(object.keys, key)
}

func (object orderedCostObject) marshal() ([]byte, error) {
	// New legacy fields have a stable home beside their related amount.
	object.insertMissingKey("cost_usd", "tokens", "model", "duration_ms", "timestamp")
	object.insertMissingKey("cost_evidence", "cost_usd")
	object.insertMissingKey("cumulative_cost_usd", "cost_usd")
	// Remaining inserted fields are deterministic, without reordering existing keys.
	known := make(map[string]bool, len(object.keys))
	for _, key := range object.keys {
		known[key] = true
	}
	var added []string
	for key := range object.fields {
		if !known[key] {
			added = append(added, key)
		}
	}
	sort.Strings(added)
	object.keys = append(object.keys, added...)
	var output bytes.Buffer
	output.WriteByte('{')
	for i, key := range object.keys {
		if i > 0 {
			output.WriteByte(',')
		}
		encoded, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		output.Write(encoded)
		output.WriteByte(':')
		output.Write(object.fields[key])
	}
	output.WriteByte('}')
	return output.Bytes(), nil
}

// retainCostMetadataOrder replaces computed metadata without reordering its
// existing fields. Any new fields follow the replacement's typed field order.
func retainCostMetadataOrder(original, replacement []byte) ([]byte, error) {
	if len(original) == 0 || bytes.Equal(bytes.TrimSpace(original), []byte("null")) {
		return replacement, nil
	}
	previous, err := decodeOrderedCostObject(original)
	if err != nil {
		return nil, err
	}
	updated, err := decodeOrderedCostObject(replacement)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(updated.keys))
	retained := make(map[string]bool, len(updated.keys))
	for _, key := range previous.keys {
		if _, exists := updated.fields[key]; exists {
			keys = append(keys, key)
			retained[key] = true
		}
	}
	for _, key := range updated.keys {
		if !retained[key] {
			keys = append(keys, key)
		}
	}
	updated.keys = keys
	return updated.marshal()
}
