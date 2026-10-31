package migration

import (
	"bytes"
	"encoding/json"
	"fmt"

	"agenticlens/internal/domain"
)

// decodeCostPresence preserves legacy null/omission before numeric decoding and
// validates v5 evidence without importing an adapter into the migration core.
func decodeCostPresence(data []byte, legacy bool) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return data, nil
	}
	switch trimmed[0] {
	case '[':
		var rows []json.RawMessage
		if err := json.Unmarshal(data, &rows); err != nil {
			return nil, err
		}
		for i, row := range rows {
			value, err := decodeCostPresence(row, legacy)
			if err != nil {
				return nil, err
			}
			rows[i] = value
		}
		return json.Marshal(rows)
	case '{':
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return nil, err
		}
		_, tokens := fields["tokens"]
		_, model := fields["model"]
		_, role := fields["role"]
		_, task := fields["subagent_type"]
		if ((tokens || model) && role) || task {
			var amount *float64
			raw, hasCost := fields["cost_usd"]
			if hasCost {
				if err := json.Unmarshal(raw, &amount); err != nil {
					return nil, fmt.Errorf("decoding cost: %w", err)
				}
			}
			if amount != nil && !domain.ValidCostAmount(*amount) {
				return nil, fmt.Errorf("invalid cost amount")
			}
			var evidence domain.CostEvidence
			if raw, exists := fields["cost_evidence"]; exists && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				if err := json.Unmarshal(raw, &evidence); err != nil {
					return nil, err
				}
			}
			if evidence.State == "" && legacy {
				evidence = domain.UnavailableCost()
				if amount != nil {
					evidence = domain.ObservedCost(*amount)
				}
			}
			if evidence.State == "" {
				return nil, fmt.Errorf("schema v5 requires cost_evidence")
			}
			if !hasCost && !legacy {
				return nil, fmt.Errorf("schema v5 requires cost_usd")
			}
			switch evidence.State {
			case domain.CostObserved:
				if amount == nil || evidence.KnownSubtotalUSD == nil || *amount != *evidence.KnownSubtotalUSD {
					return nil, fmt.Errorf("observed cost requires matching numeric amount and subtotal")
				}
			case domain.CostPartial:
				if amount != nil || evidence.KnownSubtotalUSD == nil || !domain.ValidCostAmount(*evidence.KnownSubtotalUSD) {
					return nil, fmt.Errorf("partial cost requires null amount and valid subtotal")
				}
			case domain.CostUnavailable:
				if amount != nil || evidence.KnownSubtotalUSD != nil {
					return nil, fmt.Errorf("unavailable cost requires null amounts")
				}
			default:
				return nil, fmt.Errorf("invalid cost evidence state %q", evidence.State)
			}
			encoded, err := json.Marshal(evidence)
			if err != nil {
				return nil, err
			}
			fields["cost_evidence"] = encoded
			if evidence.State == domain.CostPartial {
				subtotal, err := json.Marshal(evidence.KnownSubtotalUSD)
				if err != nil {
					return nil, err
				}
				fields["cost_usd"] = subtotal
			}
		}
		for key, value := range fields {
			updated, err := decodeCostPresence(value, legacy)
			if err != nil {
				return nil, err
			}
			fields[key] = updated
		}
		return json.Marshal(fields)
	default:
		return data, nil
	}
}
