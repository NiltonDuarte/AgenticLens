package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"

	"agenticlens/internal/domain"
)

// NormalizeCostJSON projects internal numeric accumulators to nullable wire costs.
// It operates only on explicitly tagged cost evidence/summary objects.
func NormalizeCostJSON(data []byte) ([]byte, error) { return transformCostJSON(data, false) }

// NormalizeLegacyCostJSON preserves numeric presence before legacy scalar decode.
func NormalizeLegacyCostJSON(data []byte) ([]byte, error) { return transformCostJSON(data, true) }

func transformCostJSON(data []byte, legacy bool) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty cost JSON")
	}
	switch trimmed[0] {
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, err
		}
		for i, item := range items {
			updated, err := transformCostJSON(item, legacy)
			if err != nil {
				return nil, err
			}
			items[i] = updated
		}
		return json.Marshal(items)
	case '{':
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return nil, err
		}
		// Recurse first so containing objects never hide nested facts.
		for key, value := range fields {
			if len(bytes.TrimSpace(value)) > 0 && (bytes.TrimSpace(value)[0] == '{' || bytes.TrimSpace(value)[0] == '[') {
				updated, err := transformCostJSON(value, legacy)
				if err != nil {
					return nil, err
				}
				fields[key] = updated
			}
		}
		_, invocation := fields["tokens"]
		_, model := fields["model"]
		_, role := fields["role"]
		_, task := fields["subagent_type"]
		if legacy && (((invocation || model) && role) || task) {
			var prior domain.CostEvidence
			if raw, exists := fields["cost_evidence"]; exists {
				if err := json.Unmarshal(raw, &prior); err != nil {
					return nil, err
				}
			}
			if prior.State == "" {
				var amount *float64
				if value, exists := fields["cost_usd"]; exists {
					if err := json.Unmarshal(value, &amount); err != nil {
						return nil, fmt.Errorf("invalid legacy cost: %w", err)
					}
				}
				evidence := domain.UnavailableCost()
				if amount != nil {
					if !domain.ValidCostAmount(*amount) {
						return nil, fmt.Errorf("invalid legacy cost amount")
					}
					evidence = domain.ObservedCost(*amount)
				}
				encoded, err := json.Marshal(evidence)
				if err != nil {
					return nil, err
				}
				fields["cost_evidence"] = encoded
			}
			if _, exists := fields["cost_usd"]; !exists {
				fields["cost_usd"] = json.RawMessage("null")
			}
		}
		if !legacy {
			if value, exists := fields["cost_evidence"]; exists {
				var evidence domain.CostEvidence
				if err := json.Unmarshal(value, &evidence); err != nil {
					return nil, err
				}
				if evidence.State == domain.CostPartial || evidence.State == domain.CostUnavailable {
					fields["cost_usd"] = json.RawMessage("null")
				}
			}
			for _, pair := range [][2]string{{"cost_summary", "cost_usd"}, {"cumulative_cost_summary", "cumulative_cost_usd"}} {
				if value, exists := fields[pair[0]]; exists {
					var summary domain.CostSummary
					if err := json.Unmarshal(value, &summary); err != nil {
						return nil, err
					}
					key := pair[1]
					if _, fleet := fields["total_cost_usd"]; fleet && key == "cost_usd" {
						key = "total_cost_usd"
					}
					if summary.State == "" {
						var amount float64
						if raw, exists := fields[key]; exists {
							if err := json.Unmarshal(raw, &amount); err != nil {
								return nil, err
							}
						}
						var count int64
						if raw, exists := fields["invocations"]; exists {
							if err := json.Unmarshal(raw, &count); err != nil {
								return nil, err
							}
						}
						if raw, exists := fields["total_invocations"]; exists {
							if err := json.Unmarshal(raw, &count); err != nil {
								return nil, err
							}
						}
						summary = domain.EffectiveCostSummary(summary, amount, count)
						encoded, err := json.Marshal(summary)
						if err != nil {
							return nil, err
						}
						fields[pair[0]] = encoded
					}
					encoded, err := json.Marshal(summary.CompleteAmount())
					if err != nil {
						return nil, err
					}
					fields[key] = encoded
				}
			}
			for _, pair := range [][2]string{{"role_cost_summaries", "role_costs"}, {"model_cost_summaries", "model_costs"}} {
				if value, exists := fields[pair[0]]; exists {
					var summaries map[string]domain.CostSummary
					if err := json.Unmarshal(value, &summaries); err != nil {
						return nil, err
					}
					amounts := make(map[string]*float64, len(summaries))
					for key, summary := range summaries {
						amounts[key] = summary.CompleteAmount()
					}
					encoded, err := json.Marshal(amounts)
					if err != nil {
						return nil, err
					}
					fields[pair[1]] = encoded
				}
			}
			if reason, exists := fields["role_effort_unavailable_reason"]; exists && string(reason) != "\"\"" {
				var distribution map[string]json.RawMessage
				if err := json.Unmarshal(fields["role_effort_distribution_pct"], &distribution); err != nil {
					return nil, err
				}
				for key := range distribution {
					distribution[key] = json.RawMessage("null")
				}
				encoded, err := json.Marshal(distribution)
				if err != nil {
					return nil, err
				}
				fields["role_effort_distribution_pct"] = encoded
			}
		}
		return json.Marshal(fields)
	default:
		return data, nil
	}
}

func validateCost(amount *float64, evidence *domain.CostEvidence, required bool) error {
	if amount != nil && !domain.ValidCostAmount(*amount) {
		return fmt.Errorf("cost_usd must be a finite non-negative number")
	}
	if evidence == nil {
		if required {
			return fmt.Errorf("cost_evidence is required in schema v5")
		}
		return nil
	}
	if evidence.KnownSubtotalUSD != nil && !domain.ValidCostAmount(*evidence.KnownSubtotalUSD) {
		return fmt.Errorf("known_subtotal_usd must be finite and non-negative")
	}
	switch evidence.State {
	case domain.CostObserved:
		if amount == nil || evidence.KnownSubtotalUSD == nil || *amount != *evidence.KnownSubtotalUSD {
			return fmt.Errorf("observed cost requires matching numeric cost_usd and known_subtotal_usd")
		}
	case domain.CostPartial:
		if amount != nil || evidence.KnownSubtotalUSD == nil {
			return fmt.Errorf("partial cost requires null cost_usd and numeric known_subtotal_usd")
		}
	case domain.CostUnavailable:
		if amount != nil || evidence.KnownSubtotalUSD != nil {
			return fmt.Errorf("unavailable cost requires null amounts")
		}
	default:
		return fmt.Errorf("invalid cost_evidence state %q", evidence.State)
	}
	return nil
}

// validateWireCostFields checks original wire values before scalar decoding.
func validateWireCostFields(data []byte, required bool) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}
	if trimmed[0] == '[' {
		var values []json.RawMessage
		if err := json.Unmarshal(data, &values); err != nil {
			return err
		}
		for _, value := range values {
			if err := validateWireCostFields(value, required); err != nil {
				return err
			}
		}
	} else if trimmed[0] == '{' {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			return err
		}
		_, tokens := fields["tokens"]
		_, model := fields["model"]
		_, role := fields["role"]
		_, task := fields["subagent_type"]
		if ((tokens || model) && role) || task {
			var amount *float64
			if value, exists := fields["cost_usd"]; exists {
				if err := json.Unmarshal(value, &amount); err != nil {
					return err
				}
			} else if required {
				return fmt.Errorf("cost_usd is required in schema v5")
			}
			var evidence *domain.CostEvidence
			if value, exists := fields["cost_evidence"]; exists {
				if err := json.Unmarshal(value, &evidence); err != nil {
					return err
				}
			}
			if err := validateCost(amount, evidence, required); err != nil {
				return err
			}
		}
		for _, value := range fields {
			if err := validateWireCostFields(value, required); err != nil {
				return err
			}
		}
	}
	return nil
}
