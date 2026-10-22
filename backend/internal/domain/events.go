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

// PruneEventsForPersistence filters out routine turn messages and successful tool calls,
// retaining bare error events, context compaction events, and anomaly-associated events.
// For all retained events, ErrorMessage, PromptContent, ReasoningContent, ToolInput, and ToolOutput are explicitly stripped to nil.
// If ledger.Events is nil or empty, an allocated empty non-nil slice is returned.
func PruneEventsForPersistence(ledger AgenticFactLedger) []EventFact {
	if len(ledger.Events) == 0 {
		return []EventFact{}
	}

	anomalies := DetectAnomalies(ledger)

	type anomalyKey struct {
		sessionID string
		turn      int64
	}
	anomalyTurns := make(map[anomalyKey]bool)
	wildcardTurns := make(map[int64]bool)

	for _, a := range anomalies {
		if a.Context.Turn != nil {
			if a.Context.SessionID != nil && *a.Context.SessionID != "" {
				anomalyTurns[anomalyKey{sessionID: *a.Context.SessionID, turn: *a.Context.Turn}] = true
			} else {
				wildcardTurns[*a.Context.Turn] = true
			}
		}
	}

	var retained []EventFact
	for _, ev := range ledger.Events {
		isError := ev.Status == "error" || ev.Status == "failed"
		isCompaction := ev.EventType == EventTypeCompaction
		isAnomalyLinked := wildcardTurns[ev.Turn] || (ev.SessionID != "" && anomalyTurns[anomalyKey{sessionID: ev.SessionID, turn: ev.Turn}])

		if isError || isCompaction || isAnomalyLinked {
			sanitized := ev
			sanitized.ErrorMessage = nil
			sanitized.PromptContent = nil
			sanitized.ReasoningContent = nil
			sanitized.ToolInput = nil
			sanitized.ToolOutput = nil
			retained = append(retained, sanitized)
		}
	}

	return SortEvents(retained)
}
