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

type eventIdentity struct {
	sessionID string
	eventID   string
}

type eventOccurrence struct {
	sessionID string
	timestamp string
	turn      int64
	eventType EventType
	toolName  string
}

func occurrenceOf(event EventFact) eventOccurrence {
	tool := ""
	if event.ToolName != nil {
		tool = *event.ToolName
	}
	return eventOccurrence{event.SessionID, event.Timestamp, event.Turn, event.EventType, tool}
}

// ReconcileEvents enriches recorded facts without discarding unmatched occurrences.
// Stable IDs take precedence; missing IDs require a unique occurrence on both sides.
// Inputs represent independent occurrences, not an already reconciled response.
func ReconcileEvents(baseline, enrichment []EventFact) []EventFact {
	return reconcileEvents(baseline, enrichment, false)
}

// ReconcileSourceEvents projects personas from successfully read authoritative
// source events, including an unknown persona, onto matched recorded events.
// Callers must validate source ownership before supplying enrichment. Matching,
// occurrence preservation and payload enrichment follow ReconcileEvents.
// This is a read projection; neither input is mutated or persisted.
func ReconcileSourceEvents(baseline, enrichment []EventFact) []EventFact {
	return reconcileEvents(baseline, enrichment, true)
}

func reconcileEvents(baseline, enrichment []EventFact, sourcePersonas bool) []EventFact {
	result := append([]EventFact{}, baseline...)
	baselineUsed := make([]bool, len(baseline))
	enrichmentUsed := make([]bool, len(enrichment))
	byID := make(map[eventIdentity][]int)
	byOccurrence := make(map[eventOccurrence][]int)
	for index, event := range baseline {
		if event.SessionID == "" {
			continue
		}
		if event.EventID != "" {
			key := eventIdentity{event.SessionID, event.EventID}
			byID[key] = append(byID[key], index)
		}
		key := occurrenceOf(event)
		byOccurrence[key] = append(byOccurrence[key], index)
	}
	for index, event := range enrichment {
		if event.SessionID == "" || event.EventID == "" {
			continue
		}
		key := eventIdentity{event.SessionID, event.EventID}
		candidates := byID[key]
		if len(candidates) == 0 {
			continue
		}
		matched := candidates[0]
		byID[key] = candidates[1:]
		baselineUsed[matched], enrichmentUsed[index] = true, true
		result[matched] = enrichEvent(baseline[matched], event, sourcePersonas)
	}

	// Count before matching so consuming one row cannot resolve an ambiguous group.
	candidate := make([]int, len(enrichment))
	baselineMatches := make([]int, len(baseline))
	for index, event := range enrichment {
		candidate[index] = -1
		if enrichmentUsed[index] || event.SessionID == "" {
			continue
		}
		count := 0
		for _, otherIndex := range byOccurrence[occurrenceOf(event)] {
			other := baseline[otherIndex]
			if baselineUsed[otherIndex] || (event.EventID != "" && other.EventID != "") {
				continue
			}
			if event.MessageID != nil && *event.MessageID != "" && other.MessageID != nil && *other.MessageID != "" && *event.MessageID != *other.MessageID {
				continue
			}
			count++
			candidate[index] = otherIndex
			baselineMatches[otherIndex]++
		}
		if count != 1 {
			candidate[index] = -1
		}
	}
	for index, matched := range candidate {
		if matched >= 0 && baselineMatches[matched] == 1 {
			baselineUsed[matched], enrichmentUsed[index] = true, true
			result[matched] = enrichEvent(baseline[matched], enrichment[index], sourcePersonas)
		}
	}

	for index, event := range enrichment {
		if !enrichmentUsed[index] {
			result = append(result, event)
		}
	}
	return SortEvents(result)
}

func enrichEvent(recorded, local EventFact, sourcePersona bool) EventFact {
	if sourcePersona {
		recorded.Role = local.Role
	}
	if recorded.EventID == "" {
		recorded.EventID = local.EventID
	}
	if recorded.CallID == "" {
		recorded.CallID = local.CallID
	}
	recorded.MessageID = enrichedString(recorded.MessageID, local.MessageID)
	recorded.ErrorMessage = enrichedString(recorded.ErrorMessage, local.ErrorMessage)
	recorded.PromptContent = enrichedString(recorded.PromptContent, local.PromptContent)
	recorded.ReasoningContent = enrichedString(recorded.ReasoningContent, local.ReasoningContent)
	recorded.ToolInput = enrichedString(recorded.ToolInput, local.ToolInput)
	recorded.ToolOutput = enrichedString(recorded.ToolOutput, local.ToolOutput)
	if local.DurationMs != nil {
		recorded.DurationMs = local.DurationMs
	}
	return recorded
}

func enrichedString(recorded, local *string) *string {
	if local != nil && *local != "" {
		return local
	}
	return recorded
}

// DetectAnomaliesWithEvidence attaches request-local source references to summaries.
// The references are intended only for local hydrated API responses.
func DetectAnomaliesWithEvidence(ledger AgenticFactLedger, source SourceEvents, thresholds ...AnomalyThresholdsConfig) ([]Anomaly, error) {
	events := SortEvents(source.Events)
	anomalies := RefreshAnomalySummariesFromSource(ledger, source, thresholds...)
	complete := make(map[string]bool)
	for _, id := range source.CompleteSessionIDs {
		complete[id] = true
	}
	selection := anomalyEvidenceSelection{events: map[int]bool{}, refs: map[int][]AnomalyEvidenceRef{}}
	for index, anomaly := range anomalies {
		if anomaly.Context.SessionID != nil && !complete[*anomaly.Context.SessionID] {
			continue
		}
		// Missing evidence for one retained summary does not hide associations for
		// independently complete source owners.
		selected, err := selectAnomalyEvidence(ledger, events, []Anomaly{anomaly}, thresholds...)
		if err == nil {
			selection.refs[index] = selected.refs[0]
		}
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
// with no complete current event stream. Sparse occurrence projections do not
// establish source-wide anomaly evidence.
func RefreshAnomalySummaries(ledger AgenticFactLedger, thresholds ...AnomalyThresholdsConfig) []Anomaly {
	return RefreshAnomalySummariesFromSource(ledger, SourceEvents{}, thresholds...)
}
func RefreshAnomalySummariesFromSource(ledger AgenticFactLedger, source SourceEvents, thresholds ...AnomalyThresholdsConfig) []Anomaly {
	if len(source.CompleteSessionIDs) == 0 {
		return ledger.AnomalyFacts()
	}
	cached := ledger.AnomalyFacts()
	currentSessions := make(map[string]bool)
	for _, id := range source.CompleteSessionIDs {
		currentSessions[id] = true
	}
	events := []EventFact{}
	for _, event := range source.Events {
		if currentSessions[event.SessionID] {
			events = append(events, event)
		}
	}
	observed := ledger
	observed.Sessions = nil
	for _, owner := range ledger.Sessions {
		if currentSessions[owner.ID] {
			observed.Sessions = append(observed.Sessions, owner)
		}
	}
	detected := DetectAnomaliesFromEvents(observed, events, thresholds...)
	owned := make(map[string]bool, len(ledger.Sessions))
	for _, owner := range ledger.Sessions {
		owned[owner.ID] = true
	}
	groupComplete := func(parent string) bool {
		if !currentSessions[parent] || !owned[parent] {
			return false
		}
		for _, owner := range ledger.Sessions {
			if owner.ParentID != nil && *owner.ParentID == parent && !currentSessions[owner.ID] {
				return false
			}
		}
		return true
	}
	filtered := make([]Anomaly, 0, len(detected))
	for _, anomaly := range detected {
		if anomaly.RuleID == RuleSubagentReworkLoop && anomaly.Context.SessionID != nil && !groupComplete(*anomaly.Context.SessionID) {
			continue
		}
		filtered = append(filtered, anomaly)
	}
	detected = filtered
	allComplete := completeAnomalySourceCoverage(ledger, currentSessions)
	refreshed := make([]Anomaly, 0, len(cached)+len(detected))
	for _, anomaly := range detected {
		if (anomaly.Context.SessionID == nil || *anomaly.Context.SessionID == "") && allComplete {
			anomaly.Evidence = nil
			refreshed = append(refreshed, anomaly)
		}
	}
	for _, anomaly := range cached {
		retain := false
		if anomaly.Context.SessionID == nil {
			retain = !allComplete
		} else {
			sid := *anomaly.Context.SessionID
			retain = !currentSessions[sid] || anomaly.RuleID == RuleSubagentReworkLoop && !groupComplete(sid)
		}
		if retain {
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

// PersistAnomalySummaries normalizes durable summaries without source I/O.
func PersistAnomalySummaries(ledger AgenticFactLedger, thresholds ...AnomalyThresholdsConfig) (AgenticFactLedger, error) {
	return PersistSourceAnomalySummaries(ledger, SourceEvents{}, thresholds...)
}
func PersistSourceAnomalySummaries(ledger AgenticFactLedger, source SourceEvents, thresholds ...AnomalyThresholdsConfig) (AgenticFactLedger, error) {
	if len(source.CompleteSessionIDs) == 0 {
		return ledger, nil
	}
	anomalies := RefreshAnomalySummariesFromSource(ledger, source, thresholds...)
	bySession := make(map[string][]Anomaly)
	ticket := []Anomaly{}
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
	complete := make(map[string]bool, len(source.CompleteSessionIDs))
	for _, id := range source.CompleteSessionIDs {
		complete[id] = true
	}
	ledger.Sessions = append([]SessionFact{}, ledger.Sessions...)
	for i := range ledger.Sessions {
		if !complete[ledger.Sessions[i].ID] {
			continue
		}
		owner := ledger.Sessions[i]
		incompleteGroup := false
		for _, child := range ledger.Sessions {
			if child.ParentID != nil && *child.ParentID == owner.ID && !complete[child.ID] {
				incompleteGroup = true
			}
		}
		if incompleteGroup && owner.Anomalies == nil && len(bySession[owner.ID]) == 0 {
			continue
		}
		ledger.Sessions[i].Anomalies = append([]Anomaly{}, bySession[owner.ID]...)
	}
	allComplete := completeAnomalySourceCoverage(ledger, complete)
	if allComplete {
		ledger.Anomalies = ticket
	}
	return ledger, nil
}
func NormalizeLedgerForRead(ledger AgenticFactLedger, thresholds ...AnomalyThresholdsConfig) (AgenticFactLedger, error) {
	return PersistAnomalySummaries(ledger, thresholds...)
}

type anomalyEvidenceSelection struct {
	events map[int]bool
	refs   map[int][]AnomalyEvidenceRef
}

func selectAnomalyEvidence(ledger AgenticFactLedger, events []EventFact, anomalies []Anomaly, thresholds ...AnomalyThresholdsConfig) (anomalyEvidenceSelection, error) {
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
	thresh := resolveAnomalyThresholds(thresholds...)
	retryDetections := detectToolRetryStreaksWithIndexes(events, thresh)
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
	for _, session := range ledger.Sessions {
		if session.ID != "" && session.ParentID != nil && strings.TrimSpace(*session.ParentID) != "" {
			childParent[session.ID] = strings.TrimSpace(*session.ParentID)
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

// MergeSourceEvents applies complete ephemeral owner snapshots, including empty
// snapshots. It never unions old event occurrences into a refreshed owner.
func MergeSourceEvents(previous, fresh SourceEvents) SourceEvents {
	replaced := make(map[string]bool, len(fresh.CompleteSessionIDs))
	for _, id := range fresh.CompleteSessionIDs {
		replaced[id] = true
	}
	result := SourceEvents{}
	for _, id := range previous.CompleteSessionIDs {
		if !replaced[id] {
			result.CompleteSessionIDs = append(result.CompleteSessionIDs, id)
		}
	}
	result.CompleteSessionIDs = append(result.CompleteSessionIDs, fresh.CompleteSessionIDs...)
	for _, event := range previous.Events {
		if !replaced[event.SessionID] {
			result.Events = append(result.Events, event)
		}
	}
	result.Events = append(result.Events, fresh.Events...)
	for _, evidence := range previous.CheckpointEvidence {
		if !replaced[evidence.SessionID] {
			result.CheckpointEvidence = append(result.CheckpointEvidence, evidence)
		}
	}
	result.CheckpointEvidence = append(result.CheckpointEvidence, fresh.CheckpointEvidence...)
	return result
}

// RetainIncompleteGroupSummaries preserves recorded parent rework summaries
// when a complete parent refresh cannot cover its already recorded children.
// Owner-local summaries still come from each complete source snapshot.
func RetainIncompleteGroupSummaries(previous, fresh []SessionFact) []SessionFact {
	refreshed := make(map[string]bool, len(fresh))
	old := make(map[string]SessionFact, len(previous))
	for _, owner := range fresh {
		refreshed[owner.ID] = true
	}
	incomplete := make(map[string]bool)
	for _, owner := range previous {
		old[owner.ID] = owner
		if owner.ParentID != nil && !refreshed[owner.ID] {
			incomplete[*owner.ParentID] = true
		}
	}
	result := append([]SessionFact{}, fresh...)
	for i, owner := range result {
		if !incomplete[owner.ID] {
			continue
		}
		var summaries []Anomaly
		if old[owner.ID].Anomalies != nil {
			summaries = []Anomaly{}
		}
		for _, anomaly := range owner.Anomalies {
			if anomaly.RuleID != RuleSubagentReworkLoop {
				summaries = append(summaries, anomaly)
			}
		}
		for _, anomaly := range old[owner.ID].Anomalies {
			if anomaly.RuleID == RuleSubagentReworkLoop {
				summaries = append(summaries, anomaly)
			}
		}
		result[i].Anomalies = summaries
	}
	return RetainUnrecordedCategories(previous, result)
}

// PersistSourceAnomalySummariesWithPolicy bypasses analysis for disabled generate
// output without altering retained owners before their equality guard.
func PersistSourceAnomalySummariesWithPolicy(ledger AgenticFactLedger, source SourceEvents, policy GenerationPolicy, thresholds ...AnomalyThresholdsConfig) (AgenticFactLedger, error) {
	if !policy.Anomalies {
		return ledger, nil
	}
	return PersistSourceAnomalySummaries(ledger, source, thresholds...)
}
