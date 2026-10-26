package domain

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// SortEvents returns a stable chronological copy of the event stream.
func SortEvents(events []EventFact) []EventFact {
	result := append([]EventFact{}, events...)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Timestamp != result[j].Timestamp {
			return result[i].Timestamp < result[j].Timestamp
		}
		return result[i].Turn < result[j].Turn
	})
	return result
}

// DetectAnomaliesWithEvidence attaches request-local source references to summaries.
// The references are intended only for local hydrated API responses.
func DetectAnomaliesWithEvidence(ledger AgenticFactLedger) ([]Anomaly, error) {
	events := SortEvents(ledger.EventFacts())
	anomalies := RefreshAnomalySummaries(ledger)
	selection, err := selectAnomalyEvidence(ledger, events, anomalies)
	if err != nil {
		return nil, err
	}
	for index := range anomalies {
		refs := []AnomalyEvidenceRef{}
		for _, ref := range selection.refs[index] {
			found := false
			for _, existing := range refs {
				if existing == ref {
					found = true
					break
				}
			}
			if !found {
				refs = append(refs, ref)
			}
		}
		anomalies[index].Evidence = refs
	}
	return anomalies, nil
}

// RefreshAnomalySummaries recomputes summaries for available source events and
// merged tool totals while retaining cached session summaries only for sessions
// with no current event stream.
func RefreshAnomalySummaries(ledger AgenticFactLedger) []Anomaly {
	cached := ledger.AnomalyFacts()
	events := ledger.EventFacts()
	currentSessions := make(map[string]bool)
	for _, event := range events {
		if event.SessionID != "" {
			currentSessions[event.SessionID] = true
		}
	}
	detected := DetectAnomalies(ledger)
	refreshed := make([]Anomaly, 0, len(cached)+len(detected))
	for _, anomaly := range detected {
		if anomaly.Context.SessionID == nil || *anomaly.Context.SessionID == "" {
			anomaly.Evidence = nil
			refreshed = append(refreshed, anomaly)
		}
	}
	for _, anomaly := range cached {
		if anomaly.Context.SessionID != nil && !currentSessions[*anomaly.Context.SessionID] {
			anomaly.Evidence = nil
			refreshed = append(refreshed, anomaly)
		}
	}
	for _, anomaly := range detected {
		if anomaly.Context.SessionID != nil && *anomaly.Context.SessionID != "" {
			anomaly.Evidence = nil
			refreshed = append(refreshed, anomaly)
		}
	}
	seen := make(map[string]bool, len(refreshed))
	deduped := make([]Anomaly, 0, len(refreshed))
	for _, anomaly := range refreshed {
		k := anomalyDeduplicationKey(anomaly)
		if !seen[k] {
			seen[k] = true
			deduped = append(deduped, anomaly)
		}
	}
	sortAnomalies(deduped)
	return deduped
}

func anomalyDeduplicationKey(a Anomaly) string {
	var sid, role, tool string
	var turn, count, tokens, delta, totalCalls int64
	var thresh string
	if a.Context.SessionID != nil {
		sid = *a.Context.SessionID
	}
	if a.Context.Role != nil {
		role = string(*a.Context.Role)
	}
	if a.Context.ToolName != nil {
		tool = *a.Context.ToolName
	}
	if a.Context.Turn != nil {
		turn = *a.Context.Turn
	}
	if a.Context.Count != nil {
		count = *a.Context.Count
	}
	if a.Context.ContextTokens != nil {
		tokens = *a.Context.ContextTokens
	}
	if a.Context.TokenDelta != nil {
		delta = *a.Context.TokenDelta
	}
	if a.Context.TotalCalls != nil {
		totalCalls = *a.Context.TotalCalls
	}
	if a.Context.Threshold != nil {
		thresh = *a.Context.Threshold
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%d|%d|%d|%d|%s",
		a.RuleID, a.Severity, sid, role, tool, turn, count, tokens, delta, totalCalls, thresh)
}

// PersistAnomalySummaries derives summaries from complete in-memory facts, then
// removes event rows and evidence references from the canonical ledger.
func PersistAnomalySummaries(ledger AgenticFactLedger) (AgenticFactLedger, error) {
	anomalies := RefreshAnomalySummaries(ledger)
	bySession := make(map[string][]Anomaly)
	var ticket []Anomaly
	for _, anomaly := range anomalies {
		anomaly.Evidence = nil
		if anomaly.Context.SessionID == nil || *anomaly.Context.SessionID == "" {
			ticket = append(ticket, anomaly)
			continue
		}
		sid := *anomaly.Context.SessionID
		anomaly.Context.SessionID = nil
		bySession[sid] = append(bySession[sid], anomaly)
	}
	sessions, err := SessionsFromFacts(ledger.InvocationFacts(), ledger.ToolFacts(), nil)
	if err != nil {
		return AgenticFactLedger{}, err
	}
	origSourceByID := make(map[string]string, len(ledger.Sessions))
	for _, orig := range ledger.Sessions {
		origSourceByID[orig.ID] = orig.Source
	}
	seen := make(map[string]bool, len(sessions))
	for _, session := range sessions {
		seen[session.ID] = true
	}
	for _, original := range ledger.Sessions {
		if !seen[original.ID] {
			sessions = append(sessions, SessionFact{ID: original.ID, ParentID: original.ParentID, Source: original.Source, Invocations: []InvocationFact{}, Tools: []SessionToolFact{}, Events: nil})
			seen[original.ID] = true
		}
	}
	for i := range sessions {
		if src, ok := origSourceByID[sessions[i].ID]; ok && src != "" {
			sessions[i].Source = src
		} else if ledger.Source != "" {
			sessions[i].Source = ledger.Source
		}
		sessions[i].Anomalies = bySession[sessions[i].ID]
		if sessions[i].Anomalies == nil {
			sessions[i].Anomalies = []Anomaly{}
		}
	}
	ledger.Sessions = sessions
	ledger.Anomalies = ticket
	return ledger, nil
}

type anomalyEvidenceSelection struct {
	events map[int]bool
	refs   map[int][]AnomalyEvidenceRef
}

func selectAnomalyEvidence(ledger AgenticFactLedger, events []EventFact, anomalies []Anomaly) (anomalyEvidenceSelection, error) {
	selection := anomalyEvidenceSelection{events: make(map[int]bool), refs: make(map[int][]AnomalyEvidenceRef)}
	add := func(anomalyIndex, eventIndex int) {
		selection.events[eventIndex] = true
		event := events[eventIndex]
		if event.SessionID == "" || event.EventID == "" {
			return
		}
		ref := AnomalyEvidenceRef{SessionID: event.SessionID, EventID: event.EventID}
		for _, existing := range selection.refs[anomalyIndex] {
			if existing == ref {
				return
			}
		}
		selection.refs[anomalyIndex] = append(selection.refs[anomalyIndex], ref)
	}
	reworkSummaryCounts := make(map[string]int)
	for _, anomaly := range anomalies {
		if anomaly.RuleID == RuleSubagentReworkLoop {
			if key := reworkSummaryKey(anomaly); key != "" {
				reworkSummaryCounts[key]++
			}
		}
	}
	reworkCandidates := make(map[string][]reworkEvidence)
	for _, candidate := range reworkRoleEvidence(ledger, events) {
		reworkCandidates[candidate.key] = append(reworkCandidates[candidate.key], candidate)
	}
	reworkCandidateOffset := make(map[string]int)
	retryDetections := detectToolRetryStreaksWithIndexes(events)
	retryDetectionUsed := make([]bool, len(retryDetections))
	for anomalyIndex, anomaly := range anomalies {
		sid, turn := "", int64(-1)
		if anomaly.Context.SessionID != nil {
			sid = *anomaly.Context.SessionID
		}
		if anomaly.Context.Turn != nil {
			turn = *anomaly.Context.Turn
		}
		switch anomaly.RuleID {
		case RuleToolRetryStreak:
			tool := ""
			if anomaly.Context.ToolName != nil {
				tool = *anomaly.Context.ToolName
			}
			needed := int64(3)
			if anomaly.Context.Count != nil {
				needed = *anomaly.Context.Count
			}
			if len(anomaly.Evidence) > 0 {
				if int64(len(anomaly.Evidence)) != needed {
					return selection, fmt.Errorf("anomaly %s session=%q tool=%q requires %d retry evidence refs; found %d", anomaly.RuleID, sid, tool, needed, len(anomaly.Evidence))
				}
				for _, ref := range anomaly.Evidence {
					found := false
					for eventIndex, event := range events {
						if event.SessionID == ref.SessionID && event.EventID == ref.EventID {
							if ref.SessionID != sid || event.EventType != EventTypeToolCall || event.ToolName == nil || *event.ToolName != tool || (event.Status != "error" && event.Status != "failed") {
								return selection, fmt.Errorf("anomaly %s evidence session=%q event_id=%q does not match its detected retry streak", anomaly.RuleID, ref.SessionID, ref.EventID)
							}
							add(anomalyIndex, eventIndex)
							found = true
							break
						}
					}
					if !found {
						return selection, fmt.Errorf("anomaly %s references missing retry event session=%q event_id=%q", anomaly.RuleID, ref.SessionID, ref.EventID)
					}
				}
				continue
			}
			matched := -1
			for candidateIndex, candidate := range retryDetections {
				if retryDetectionUsed[candidateIndex] {
					continue
				}
				candidateAnomaly := candidate.anomaly
				if candidateAnomaly.Context.SessionID == nil || candidateAnomaly.Context.ToolName == nil || candidateAnomaly.Context.Turn == nil || candidateAnomaly.Context.Count == nil {
					continue
				}
				if *candidateAnomaly.Context.SessionID == sid && *candidateAnomaly.Context.ToolName == tool && *candidateAnomaly.Context.Turn == turn && *candidateAnomaly.Context.Count == needed {
					matched = candidateIndex
					break
				}
			}
			if matched < 0 {
				return selection, fmt.Errorf("anomaly %s session=%q tool=%q turn=%d requires a matching retry streak of %d events; found none", anomaly.RuleID, sid, tool, turn, needed)
			}
			retryDetectionUsed[matched] = true
			for _, eventIndex := range retryDetections[matched].eventIndexes {
				add(anomalyIndex, eventIndex)
			}
		case RuleToolHighFailureRate:
			if anomaly.Context.ToolName == nil || anomaly.Context.Count == nil {
				return selection, fmt.Errorf("ticket anomaly %s is missing tool/error count context", anomaly.RuleID)
			}
			var indices []int
			for i, event := range events {
				if event.EventType == EventTypeToolCall && event.ToolName != nil && *event.ToolName == *anomaly.Context.ToolName && (event.Status == "error" || event.Status == "failed") {
					indices = append(indices, i)
				}
			}
			if int64(len(indices)) != *anomaly.Context.Count {
				return selection, fmt.Errorf("ticket anomaly %s tool=%q requires %d failed tool-call evidence events but found %d", anomaly.RuleID, *anomaly.Context.ToolName, *anomaly.Context.Count, len(indices))
			}
			for _, index := range indices {
				add(anomalyIndex, index)
			}
		case RuleContextSaturationSpike:
			if anomaly.Context.Turn == nil || anomaly.Context.ContextTokens == nil {
				continue
			}
			if sid == "" {
				return selection, fmt.Errorf("anomaly %s turn=%d has no owning session", anomaly.RuleID, turn)
			}
			current := -1
			for i, event := range events {
				if event.SessionID == sid && event.Turn == turn && event.ContextTokens == *anomaly.Context.ContextTokens {
					current = i
				}
			}
			if current < 0 {
				return selection, fmt.Errorf("anomaly %s session=%q turn=%d has no matching context evidence event", anomaly.RuleID, sid, turn)
			}
			if anomaly.Context.TokenDelta != nil {
				previous := -1
				for i := current - 1; i >= 0; i-- {
					if events[i].SessionID == sid {
						previous = i
						break
					}
				}
				if previous < 0 {
					return selection, fmt.Errorf("anomaly %s session=%q turn=%d has no before-context evidence event", anomaly.RuleID, sid, turn)
				}
				add(anomalyIndex, previous)
			}
			add(anomalyIndex, current)
		case RuleSubagentReworkLoop:
			key := reworkSummaryKey(anomaly)
			if key == "" {
				continue
			}
			candidates := reworkCandidates[key]
			if len(candidates) != reworkSummaryCounts[key] {
				continue
			}
			offset := reworkCandidateOffset[key]
			if offset >= len(candidates) {
				continue
			}
			for _, index := range candidates[offset].eventIndexes {
				add(anomalyIndex, index)
			}
			reworkCandidateOffset[key] = offset + 1
		}
	}
	return selection, nil
}

type reworkEvidence struct {
	key          string
	eventIndexes []int
}

func reworkSummaryKey(anomaly Anomaly) string {
	if anomaly.Context.Role == nil || anomaly.Context.SecondaryRole == nil || anomaly.Context.Count == nil {
		return ""
	}
	return reworkEvidenceKey(string(*anomaly.Context.Role), string(*anomaly.Context.SecondaryRole), *anomaly.Context.Count)
}

func reworkEvidenceKey(first, second string, cycles int64) string {
	return strings.ToLower(strings.TrimSpace(first)) + "\x00" + strings.ToLower(strings.TrimSpace(second)) + "\x00" + strconv.FormatInt(cycles, 10)
}

// reworkRoleEvidence mirrors the event-sequence branch of the detector and returns each
// minimal alternating child-role sequence that reaches the recorded cycle count.
func reworkRoleEvidence(ledger AgenticFactLedger, events []EventFact) []reworkEvidence {
	childParent := make(map[string]string)
	for _, invocation := range ledger.InvocationFacts() {
		if invocation.ID != "" && invocation.ParentID != nil && strings.TrimSpace(*invocation.ParentID) != "" {
			childParent[invocation.ID] = strings.TrimSpace(*invocation.ParentID)
		}
	}
	type roleEvent struct {
		index int
		role  string
	}
	groups := make(map[string][]roleEvent)
	for index, event := range events {
		parent, ok := childParent[event.SessionID]
		if !ok {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(string(event.Role)))
		if role == "" {
			continue
		}
		sequence := groups[parent]
		if len(sequence) > 0 && sequence[len(sequence)-1].role == role {
			continue
		}
		groups[parent] = append(sequence, roleEvent{index: index, role: role})
	}
	parents := make([]string, 0, len(groups))
	for parent := range groups {
		parents = append(parents, parent)
	}
	sort.Strings(parents)
	var evidence []reworkEvidence
	for _, parent := range parents {
		sequence := groups[parent]
		for i := 0; i < len(sequence)-1; {
			first, second := sequence[i].role, sequence[i+1].role
			cycles, end := int64(0), i
			for end+1 < len(sequence) && sequence[end].role == first && sequence[end+1].role == second {
				cycles++
				end += 2
			}
			if cycles >= 3 {
				indices := make([]int, 0, end-i)
				for pair := i; pair < end; pair++ {
					indices = append(indices, sequence[pair].index)
				}
				evidence = append(evidence, reworkEvidence{key: reworkEvidenceKey(first, second, cycles), eventIndexes: indices})
				i = end
			} else {
				i++
			}
		}
	}
	return evidence
}
