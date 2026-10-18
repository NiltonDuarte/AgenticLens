package domain

import (
	"sort"
)

// SortEvents returns a new slice containing the events sorted stably by Timestamp ascending,
// and by Turn ascending when timestamps are equal.
// If events is nil or empty, an allocated empty non-nil slice is returned.
// The original slice is not modified.
func SortEvents(events []EventFact) []EventFact {
	if len(events) == 0 {
		return []EventFact{}
	}

	result := make([]EventFact, len(events))
	copy(result, events)

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Timestamp != result[j].Timestamp {
			return result[i].Timestamp < result[j].Timestamp
		}
		return result[i].Turn < result[j].Turn
	})

	return result
}
