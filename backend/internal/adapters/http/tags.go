package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"agenticlens/internal/domain"
)

func parseTagFilters(query url.Values) ([]domain.TagFilter, error) {
	inputs, present := query["tag_filter"]
	if !present {
		return nil, nil
	}
	if len(inputs) != 1 || len(inputs[0]) > 65536 || strings.TrimSpace(inputs[0]) == "" {
		return nil, fmt.Errorf("tag_filter requires one JSON array (at most 65536 bytes)")
	}
	var filters []domain.TagFilter
	decoder := json.NewDecoder(bytes.NewBufferString(inputs[0]))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&filters); err != nil {
		return nil, fmt.Errorf("invalid tag_filter: %w", err)
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("tag_filter must contain one JSON array")
	}
	if filters == nil || len(filters) > 128 {
		return nil, fmt.Errorf("tag_filter must be an array of at most 128 entries")
	}
	seen := make(map[string]bool)
	for _, filter := range filters {
		if filter.Key == "" || strings.TrimSpace(filter.Key) != filter.Key || seen[filter.Key] {
			return nil, fmt.Errorf("tag_filter keys must be nonblank, trimmed and unique")
		}
		seen[filter.Key] = true
		if len(filter.Values) == 0 && !filter.Missing {
			return nil, fmt.Errorf("tag_filter %q must select values or Missing", filter.Key)
		}
		for _, value := range filter.Values {
			if value == "" || strings.TrimSpace(value) != value {
				return nil, fmt.Errorf("tag_filter values must be nonblank trimmed strings")
			}
		}
	}
	return filters, nil
}
