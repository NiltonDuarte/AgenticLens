package domain

import (
	"fmt"
	"strings"
)

// NormalizeTags validates one input layer, rejecting keys which collide after trimming.
func NormalizeTags(input map[string]string) (map[string]string, error) {
	if input == nil {
		return nil, nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		k, v := strings.TrimSpace(key), strings.TrimSpace(value)
		if k == "" || v == "" {
			return nil, fmt.Errorf("tags require nonblank string keys and values")
		}
		if _, exists := result[k]; exists {
			return nil, fmt.Errorf("duplicate normalized tag key %q", k)
		}
		result[k] = v
	}
	return result, nil
}

// ValidateTags rejects malformed persisted values rather than relabeling history on read.
func ValidateTags(tags map[string]string) error {
	normalized, err := NormalizeTags(tags)
	if err != nil {
		return err
	}
	for key, value := range tags {
		if normalized[key] != value {
			return fmt.Errorf("tag %q must have trimmed key and value", key)
		}
	}
	return nil
}

// MergeTags returns an independent map, preserving lower-layer keys absent from overrides.
func MergeTags(base, overrides map[string]string) map[string]string {
	if base == nil && overrides == nil {
		return nil
	}
	result := make(map[string]string, len(base)+len(overrides))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range overrides {
		result[k] = v
	}
	return result
}

// TagFilter selects literal values or absence for one ticket-level key.
type TagFilter struct {
	Key     string   `json:"key"`
	Values  []string `json:"values"`
	Missing bool     `json:"missing"`
}

func MatchTicketTags(tags map[string]string, filters []TagFilter) bool {
	for _, filter := range filters {
		value, exists := tags[filter.Key]
		matched := !exists && filter.Missing
		if exists {
			for _, candidate := range filter.Values {
				if value == candidate {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func FilterLedgersByTags(ledgers []AgenticMetricsLedger, filters []TagFilter) []AgenticMetricsLedger {
	if len(filters) == 0 {
		return ledgers
	}
	result := make([]AgenticMetricsLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		if MatchTicketTags(ledger.Facts.Tags, filters) {
			result = append(result, ledger)
		}
	}
	return result
}
