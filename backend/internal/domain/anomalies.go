package domain

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// AnomalyContext captures rule-specific parameters, thresholds, and execution context.
type AnomalyContext struct {
	ToolName       *string    `json:"tool_name,omitempty"`
	SessionID      *string    `json:"session_id,omitempty"`
	Turn           *int64     `json:"turn,omitempty"`
	Role           *AgentRole `json:"role,omitempty"`
	SecondaryRole  *AgentRole `json:"secondary_role,omitempty"`
	Count          *int64     `json:"count,omitempty"`
	TotalCalls     *int64     `json:"total_calls,omitempty"`
	TokenDelta     *int64     `json:"token_delta,omitempty"`
	ContextTokens  *int64     `json:"context_tokens,omitempty"`
	Threshold      *string    `json:"threshold,omitempty"`
	FailureRatePct *float64   `json:"failure_rate_pct,omitempty"`
}

// Anomaly represents a single detected execution health issue or operational degradation.
type Anomaly struct {
	RuleID   AnomalyRuleID        `json:"rule_id"`
	Severity AnomalySeverity      `json:"severity"`
	Context  AnomalyContext       `json:"context"`
	Evidence []AnomalyEvidenceRef `json:"evidence,omitempty"`
}

// AnomalyEvidenceRef identifies a source event without duplicating its session or event payload.
type AnomalyEvidenceRef struct {
	SessionID string `json:"session_id"`
	EventID   string `json:"event_id"`
}

// AnomalyDisplayText derives concise presentation strings from canonical anomaly facts.
// It is intentionally outside the persisted Anomaly shape.
func AnomalyDisplayText(anomaly Anomaly) (string, string) {
	ctx := anomaly.Context
	tool := "tool"
	if ctx.ToolName != nil && strings.TrimSpace(*ctx.ToolName) != "" {
		tool = *ctx.ToolName
	}
	count := int64(0)
	if ctx.Count != nil {
		count = *ctx.Count
	}
	switch anomaly.RuleID {
	case RuleToolRetryStreak:
		if ctx.ToolName != nil && ctx.Count != nil {
			return fmt.Sprintf("Tool retry streak: %s (%d consecutive errors)", tool, count), fmt.Sprintf("Tool '%s' encountered %d consecutive execution failures.", tool, count)
		}
	case RuleToolHighFailureRate:
		if ctx.ToolName != nil && ctx.Count != nil && ctx.TotalCalls != nil && ctx.FailureRatePct != nil {
			threshold := ctx.Threshold
			thresholdText := "configured threshold"
			if threshold != nil && *threshold != "" {
				thresholdText = *threshold
			}
			return fmt.Sprintf("High tool failure rate: %s (%.1f%%)", tool, *ctx.FailureRatePct), fmt.Sprintf("Tool '%s' failed %d out of %d calls (%.1f%% failure rate, threshold: %s).", tool, count, *ctx.TotalCalls, *ctx.FailureRatePct, thresholdText)
		}
	case RuleContextSaturationSpike:
		if ctx.ContextTokens != nil && ctx.Threshold != nil && *ctx.Threshold != "" && ctx.TokenDelta == nil {
			ceilingName := *ctx.Threshold
			if val, err := strconv.ParseInt(*ctx.Threshold, 10, 64); err == nil {
				ceilingName = FormatTokensK(val)
			}
			sevLabel := "critical"
			if anomaly.Severity == SeverityWarning {
				sevLabel = "warning"
			}
			return fmt.Sprintf("Context ceiling breached (%s tokens)", FormatTokensK(*ctx.ContextTokens)), fmt.Sprintf("Context tokens reached %s, breaching the %s %s ceiling.", FormatTokensK(*ctx.ContextTokens), ceilingName, sevLabel)
		}
		if ctx.TokenDelta != nil && ctx.ContextTokens != nil {
			return fmt.Sprintf("Context saturation spike (+%s tokens)", FormatTokensK(*ctx.TokenDelta)), fmt.Sprintf("Context tokens surged by %s to %s tokens.", FormatTokensK(*ctx.TokenDelta), FormatTokensK(*ctx.ContextTokens))
		}
	case RuleSubagentReworkLoop:
		if ctx.Role != nil && ctx.SecondaryRole != nil && ctx.Count != nil {
			return fmt.Sprintf("Subagent rework loop: %s <-> %s (%d cycles)", *ctx.Role, *ctx.SecondaryRole, count), fmt.Sprintf("Subagent roles '%s' and '%s' alternated for %d consecutive cycles.", *ctx.Role, *ctx.SecondaryRole, count)
		}
	}
	name := strings.ToLower(strings.ReplaceAll(string(anomaly.RuleID), "_", " "))
	if name == "" {
		name = "execution anomaly"
	}
	return strings.ToUpper(name[:1]) + name[1:], fmt.Sprintf("%s detected; some diagnostic context is unavailable.", name)
}

func strPtr(s string) *string {
	return &s
}

func agentRolePtr(r AgentRole) *AgentRole {
	return &r
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// detectToolRetryStreaks detects consecutive execution failures of the same tool in the event stream,
// strictly scoped per session ID.
type retryStreakDetection struct {
	anomaly      Anomaly
	eventIndexes []int
}

func detectToolRetryStreaksWithIndexes(events []EventFact, thresholds AnomalyThresholdsConfig) []retryStreakDetection {
	if len(events) == 0 {
		return nil
	}
	var detections []retryStreakDetection
	bySession := make(map[string][]int)
	for index, event := range events {
		if strings.TrimSpace(event.SessionID) != "" {
			bySession[event.SessionID] = append(bySession[event.SessionID], index)
		}
	}
	sessionIDs := make([]string, 0, len(bySession))
	for sessionID := range bySession {
		sessionIDs = append(sessionIDs, sessionID)
	}
	sort.Strings(sessionIDs)

	warnStreak := int64(thresholds.ConsecutiveErrorsStreakWarning)
	if warnStreak <= 0 {
		warnStreak = ConsecutiveToolErrorsStreakWarning
	}
	critStreak := int64(thresholds.ConsecutiveErrorsStreakCritical)
	if critStreak <= 0 {
		critStreak = ConsecutiveToolErrorsStreakCritical
	}

	for _, sessionID := range sessionIDs {
		indices := bySession[sessionID]
		sort.SliceStable(indices, func(i, j int) bool {
			a, b := events[indices[i]], events[indices[j]]
			if a.Timestamp != b.Timestamp {
				return a.Timestamp < b.Timestamp
			}
			return a.Turn < b.Turn
		})
		var activeTool string
		var consecutiveErrors int64
		var lastTurn int64
		var activeEventIndexes []int
		emitStreak := func() {
			if consecutiveErrors >= warnStreak && activeTool != "" {
				severity, threshold := SeverityWarning, strconv.FormatInt(warnStreak, 10)
				if consecutiveErrors >= critStreak {
					severity, threshold = SeverityCritical, strconv.FormatInt(critStreak, 10)
				}
				var evidence []AnomalyEvidenceRef
				for _, eventIndex := range activeEventIndexes {
					event := events[eventIndex]
					if event.SessionID != "" && event.EventID != "" {
						evidence = append(evidence, AnomalyEvidenceRef{SessionID: event.SessionID, EventID: event.EventID})
					}
				}
				detections = append(detections, retryStreakDetection{
					anomaly: Anomaly{
						RuleID: RuleToolRetryStreak, Severity: severity,
						Context:  AnomalyContext{ToolName: strPtr(activeTool), Count: int64Ptr(consecutiveErrors), SessionID: strPtr(sessionID), Turn: int64Ptr(lastTurn), Threshold: strPtr(threshold)},
						Evidence: evidence,
					},
					eventIndexes: append([]int{}, activeEventIndexes...),
				})
			}
			activeEventIndexes = nil
		}
		for _, eventIndex := range indices {
			event := events[eventIndex]
			if event.EventType != EventTypeToolCall || event.ToolName == nil || *event.ToolName == "" {
				continue
			}
			tool := *event.ToolName
			if event.Status == "error" || event.Status == "failed" {
				if tool == activeTool {
					consecutiveErrors++
					lastTurn = event.Turn
					activeEventIndexes = append(activeEventIndexes, eventIndex)
				} else {
					emitStreak()
					activeTool = tool
					consecutiveErrors = 1
					lastTurn = event.Turn
					activeEventIndexes = []int{eventIndex}
				}
			} else {
				emitStreak()
				activeTool = ""
				consecutiveErrors = 0
			}
		}
		emitStreak()
	}
	return detections
}

func detectToolRetryStreaks(events []EventFact, thresholds AnomalyThresholdsConfig) []Anomaly {
	detections := detectToolRetryStreaksWithIndexes(events, thresholds)
	anomalies := make([]Anomaly, 0, len(detections))
	for _, detection := range detections {
		anomalies = append(anomalies, detection.anomaly)
	}
	return anomalies
}

// detectToolHighFailureRates detects tools with an excessively high failure rate across the ticket.
func detectToolHighFailureRates(ledger AgenticFactLedger, thresholds AnomalyThresholdsConfig) []Anomaly {
	type toolStats struct {
		calls  int64
		errors int64
	}

	statsMap := make(map[string]*toolStats)

	if len(ledger.ToolFacts()) > 0 {
		for _, t := range ledger.ToolFacts() {
			if t.Name == "" {
				continue
			}
			st, ok := statsMap[t.Name]
			if !ok {
				st = &toolStats{}
				statsMap[t.Name] = st
			}
			st.calls += t.Calls
			st.errors += t.Errors
		}
	} else if len(ledger.EventFacts()) > 0 {
		for _, e := range ledger.EventFacts() {
			if e.EventType != EventTypeToolCall || e.ToolName == nil || *e.ToolName == "" {
				continue
			}
			name := *e.ToolName
			st, ok := statsMap[name]
			if !ok {
				st = &toolStats{}
				statsMap[name] = st
			}
			st.calls++
			if e.Status == "error" || e.Status == "failed" {
				st.errors++
			}
		}
	}

	var toolNames []string
	for name := range statsMap {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	var anomalies []Anomaly
	for _, name := range toolNames {
		st := statsMap[name]
		if st.calls < 5 {
			continue
		}

		rate := float64(st.errors) / float64(st.calls)
		if rate > 0.20 {
			sev := SeverityWarning
			thresholdStr := ">20%"
			if rate >= 0.50 {
				sev = SeverityCritical
				thresholdStr = ">=50%"
			}
			pct := math.Round(rate*1000.0) / 10.0
			anomaly := Anomaly{
				RuleID:   RuleToolHighFailureRate,
				Severity: sev,
				Context: AnomalyContext{
					ToolName:       strPtr(name),
					Count:          int64Ptr(st.errors),
					TotalCalls:     int64Ptr(st.calls),
					FailureRatePct: float64Ptr(pct),
					Threshold:      strPtr(thresholdStr),
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	return anomalies
}

// detectContextSaturationSpikes detects single-turn context surges and critical ceiling breaches.
func detectContextSaturationSpikes(ledger AgenticFactLedger, thresholds AnomalyThresholdsConfig) []Anomaly {
	var anomalies []Anomaly

	warnDelta := thresholds.TokenVelocityWarningDeltaTokens
	if warnDelta <= 0 {
		warnDelta = TokenVelocityWarningDeltaTokens
	}
	critDelta := thresholds.TokenVelocityCriticalDeltaTokens
	if critDelta <= 0 {
		critDelta = TokenVelocityCriticalDeltaTokens
	}
	warnCeiling := thresholds.ContextCeilingWarningTokens
	if warnCeiling <= 0 {
		warnCeiling = ContextCeilingWarningTokens
	}
	critCeiling := thresholds.ContextCeilingCriticalTokens
	if critCeiling <= 0 {
		critCeiling = ContextCeilingCriticalTokens
	}

	if len(ledger.EventFacts()) > 0 {
		sessionEvents := make(map[string][]EventFact)
		for _, e := range ledger.EventFacts() {
			sid := e.SessionID
			if sid == "" {
				sid = "default"
			}
			sessionEvents[sid] = append(sessionEvents[sid], e)
		}

		var sessionIDs []string
		for sid := range sessionEvents {
			sessionIDs = append(sessionIDs, sid)
		}
		sort.Strings(sessionIDs)

		warnCeilingReported := make(map[string]bool)
		critCeilingReported := make(map[string]bool)

		for _, sid := range sessionIDs {
			events := sessionEvents[sid]
			var hasPrev bool
			var prevTokens int64

			for _, ev := range events {
				// 1. Single-turn jump detection
				if hasPrev {
					delta := ev.ContextTokens - prevTokens
					if delta > warnDelta {
						sev := SeverityWarning
						thresh := fmt.Sprintf("%d", warnDelta)
						if delta > critDelta {
							sev = SeverityCritical
							thresh = fmt.Sprintf("%d", critDelta)
						}
						anomaly := Anomaly{
							RuleID:   RuleContextSaturationSpike,
							Severity: sev,
							Context: AnomalyContext{
								SessionID:     strPtr(ev.SessionID),
								Turn:          int64Ptr(ev.Turn),
								TokenDelta:    int64Ptr(delta),
								ContextTokens: int64Ptr(ev.ContextTokens),
								Threshold:     strPtr(thresh),
							},
						}
						anomalies = append(anomalies, anomaly)
					}
				}

				prevTokens = ev.ContextTokens
				hasPrev = true

				// 2. Ceiling breach detection (warning tokens and critical tokens)
				if ev.ContextTokens >= critCeiling && !critCeilingReported[sid] {
					critCeilingReported[sid] = true
					warnCeilingReported[sid] = true
					anomaly := Anomaly{
						RuleID:   RuleContextSaturationSpike,
						Severity: SeverityCritical,
						Context: AnomalyContext{
							SessionID:     strPtr(ev.SessionID),
							Turn:          int64Ptr(ev.Turn),
							ContextTokens: int64Ptr(ev.ContextTokens),
							Threshold:     strPtr(fmt.Sprintf("%d", critCeiling)),
						},
					}
					anomalies = append(anomalies, anomaly)
				} else if ev.ContextTokens >= warnCeiling && !warnCeilingReported[sid] {
					warnCeilingReported[sid] = true
					anomaly := Anomaly{
						RuleID:   RuleContextSaturationSpike,
						Severity: SeverityWarning,
						Context: AnomalyContext{
							SessionID:     strPtr(ev.SessionID),
							Turn:          int64Ptr(ev.Turn),
							ContextTokens: int64Ptr(ev.ContextTokens),
							Threshold:     strPtr(fmt.Sprintf("%d", warnCeiling)),
						},
					}
					anomalies = append(anomalies, anomaly)
				}
			}
		}
	} else {
		// v1 Fallback: inspect session-owned invocations against warning and critical ceilings.
		for _, session := range ledger.Sessions {
			for _, inv := range session.Invocations {
				severity, threshold := SeverityWarning, warnCeiling
				if inv.Tokens.PeakContext >= critCeiling {
					severity, threshold = SeverityCritical, critCeiling
				} else if inv.Tokens.PeakContext < warnCeiling {
					continue
				}
				var sid *string
				if session.ID != "" {
					sid = strPtr(session.ID)
				}
				anomaly := Anomaly{
					RuleID:   RuleContextSaturationSpike,
					Severity: severity,
					Context: AnomalyContext{
						SessionID:     sid,
						Role:          agentRolePtr(AgentRole(inv.Role)),
						ContextTokens: int64Ptr(inv.Tokens.PeakContext),
						Threshold:     strPtr(fmt.Sprintf("%d", threshold)),
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

// detectSubagentReworkLoops detects cyclic role alternation within each explicit
// invocation parent-child group. Independent child groups never share a role stream.
func detectSubagentReworkLoops(ledger AgenticFactLedger) []Anomaly {
	childParent := make(map[string]string)
	invocationRoles := make(map[string][]string)
	for _, session := range ledger.Sessions {
		if session.ParentID == nil {
			continue
		}
		parent := strings.TrimSpace(*session.ParentID)
		if parent == "" {
			continue
		}
		if session.ID != "" {
			childParent[session.ID] = parent
		}
		for _, invocation := range session.Invocations {
			role := strings.ToLower(strings.TrimSpace(string(invocation.Role)))
			if role != "" {
				invocationRoles[parent] = appendDistinctRole(invocationRoles[parent], role)
			}
		}
	}

	eventRoles := make(map[string][]string)
	for _, event := range ledger.EventFacts() {
		parent, child := childParent[event.SessionID]
		if !child {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(string(event.Role)))
		if role == "" {
			continue
		}
		eventRoles[parent] = appendDistinctRole(eventRoles[parent], role)
	}

	parents := make([]string, 0, len(childParent))
	seenParent := make(map[string]bool)
	for _, parent := range childParent {
		if !seenParent[parent] {
			parents = append(parents, parent)
			seenParent[parent] = true
		}
	}
	for parent := range invocationRoles {
		if !seenParent[parent] {
			parents = append(parents, parent)
			seenParent[parent] = true
		}
	}
	sort.Strings(parents)
	var anomalies []Anomaly
	for _, parent := range parents {
		roles := eventRoles[parent]
		if len(roles) == 0 {
			roles = invocationRoles[parent]
		}
		for i := 0; i < len(roles)-1; {
			a, b := roles[i], roles[i+1]
			if a == b {
				i++
				continue
			}
			cycles, pos := int64(0), i
			for pos+1 < len(roles) && roles[pos] == a && roles[pos+1] == b {
				cycles++
				pos += 2
			}
			if cycles >= 3 {
				severity, threshold := SeverityWarning, "3"
				if cycles >= 5 {
					severity, threshold = SeverityCritical, "5"
				}
				anomalies = append(anomalies, Anomaly{
					RuleID: RuleSubagentReworkLoop, Severity: severity,
					Context: AnomalyContext{Role: agentRolePtr(AgentRole(a)), SecondaryRole: agentRolePtr(AgentRole(b)), Count: int64Ptr(cycles), Threshold: strPtr(threshold)},
				})
				i = pos
			} else {
				i++
			}
		}
	}
	return anomalies
}

func appendDistinctRole(roles []string, role string) []string {
	if role == "" || (len(roles) > 0 && roles[len(roles)-1] == role) {
		return roles
	}
	return append(roles, role)
}

func resolveAnomalyThresholds(thresholds ...AnomalyThresholdsConfig) AnomalyThresholdsConfig {
	def := DefaultAnomalyThresholdsConfig()
	if len(thresholds) == 0 {
		return def
	}
	t := thresholds[0]
	if t.StallThresholdSeconds <= 0 {
		t.StallThresholdSeconds = def.StallThresholdSeconds
	}
	if t.ErrorRateThreshold <= 0 {
		t.ErrorRateThreshold = def.ErrorRateThreshold
	}
	if t.ContextExhaustionPct <= 0 {
		t.ContextExhaustionPct = def.ContextExhaustionPct
	}
	if t.ConsecutiveErrorsStreakWarning <= 0 {
		t.ConsecutiveErrorsStreakWarning = def.ConsecutiveErrorsStreakWarning
	}
	if t.ConsecutiveErrorsStreakCritical <= 0 {
		t.ConsecutiveErrorsStreakCritical = def.ConsecutiveErrorsStreakCritical
	}
	if t.ContextCeilingWarningTokens <= 0 {
		t.ContextCeilingWarningTokens = def.ContextCeilingWarningTokens
	}
	if t.ContextCeilingCriticalTokens <= 0 {
		t.ContextCeilingCriticalTokens = def.ContextCeilingCriticalTokens
	}
	if t.TokenVelocityWarningTokens <= 0 {
		t.TokenVelocityWarningTokens = def.TokenVelocityWarningTokens
	}
	if t.TokenVelocityCriticalTokens <= 0 {
		t.TokenVelocityCriticalTokens = def.TokenVelocityCriticalTokens
	}
	if t.TokenVelocityWarningDeltaTokens <= 0 {
		t.TokenVelocityWarningDeltaTokens = def.TokenVelocityWarningDeltaTokens
	}
	if t.TokenVelocityCriticalDeltaTokens <= 0 {
		t.TokenVelocityCriticalDeltaTokens = def.TokenVelocityCriticalDeltaTokens
	}
	if t.ToolThrashingMinFailures <= 0 {
		t.ToolThrashingMinFailures = def.ToolThrashingMinFailures
	}
	if t.ToolThrashingCriticalFailures <= 0 {
		t.ToolThrashingCriticalFailures = def.ToolThrashingCriticalFailures
	}
	return t
}

// DetectAnomalies evaluates deterministic operational rules against the fact ledger.
func DetectAnomalies(ledger AgenticFactLedger, thresholds ...AnomalyThresholdsConfig) []Anomaly {
	cfg := resolveAnomalyThresholds(thresholds...)
	var anomalies []Anomaly

	anomalies = append(anomalies, detectToolRetryStreaks(ledger.EventFacts(), cfg)...)
	anomalies = append(anomalies, detectToolHighFailureRates(ledger, cfg)...)
	anomalies = append(anomalies, detectContextSaturationSpikes(ledger, cfg)...)
	anomalies = append(anomalies, detectSubagentReworkLoops(ledger)...)

	if len(anomalies) == 0 {
		return []Anomaly{}
	}

	sortAnomalies(anomalies)

	return anomalies
}

func sortAnomalies(anomalies []Anomaly) {
	sort.SliceStable(anomalies, func(i, j int) bool {
		a := anomalies[i]
		b := anomalies[j]

		// 1. Primary sort: Severity descending (CRITICAL before WARNING)
		if a.Severity != b.Severity {
			return a.Severity == SeverityCritical
		}

		// 2. Secondary sort: RuleID ascending
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}

		// 3. Tertiary sort: Turn ascending (or ToolName ascending when turn is absent)
		if a.Context.Turn != nil && b.Context.Turn != nil {
			if *a.Context.Turn != *b.Context.Turn {
				return *a.Context.Turn < *b.Context.Turn
			}
		} else if a.Context.Turn != nil && b.Context.Turn == nil {
			return true
		} else if a.Context.Turn == nil && b.Context.Turn != nil {
			return false
		}

		toolA := ""
		if a.Context.ToolName != nil {
			toolA = *a.Context.ToolName
		}
		toolB := ""
		if b.Context.ToolName != nil {
			toolB = *b.Context.ToolName
		}
		if toolA != toolB {
			return toolA < toolB
		}

		return string(a.RuleID) < string(b.RuleID)
	})
}

// ComputeFleetAnomalies aggregates operational anomalies, severity counts, and brittle tools across all ledgers.
func ComputeFleetAnomalies(ledgers []AgenticFactLedger, thresholds ...AnomalyThresholdsConfig) FleetAnomalySummary {
	severityCounts := map[string]int64{
		string(SeverityWarning):  0,
		string(SeverityCritical): 0,
	}
	ruleCounts := map[string]int64{
		string(RuleToolRetryStreak):        0,
		string(RuleToolHighFailureRate):    0,
		string(RuleContextSaturationSpike): 0,
		string(RuleSubagentReworkLoop):     0,
	}

	if len(ledgers) == 0 {
		return FleetAnomalySummary{
			TotalAnomalies:       0,
			WarningCount:         0,
			CriticalCount:        0,
			SeverityCounts:       severityCounts,
			RuleCounts:           ruleCounts,
			TotalTickets:         0,
			TicketsWithAnomalies: 0,
			AnomalyRatePerTicket: 0.0,
			BrittleTools:         []BrittleToolSummary{},
		}
	}

	type toolMetrics struct {
		calls     int64
		errors    int64
		anomalies int64
	}
	toolStats := make(map[string]*toolMetrics)
	getTool := func(name string) *toolMetrics {
		st, ok := toolStats[name]
		if !ok {
			st = &toolMetrics{}
			toolStats[name] = st
		}
		return st
	}

	var totalAnomalies, warningCount, criticalCount, ticketsWithAnomalies int64

	for _, ledger := range ledgers {
		ticketAnomalies := RefreshAnomalySummaries(ledger, thresholds...)
		if len(ticketAnomalies) > 0 {
			ticketsWithAnomalies++
		}
		for _, a := range ticketAnomalies {
			totalAnomalies++
			if a.Severity == SeverityWarning {
				warningCount++
				severityCounts[string(SeverityWarning)]++
			} else if a.Severity == SeverityCritical {
				criticalCount++
				severityCounts[string(SeverityCritical)]++
			}
			ruleCounts[string(a.RuleID)]++
			if a.Context.ToolName != nil && *a.Context.ToolName != "" {
				getTool(*a.Context.ToolName).anomalies++
			}
		}

		if len(ledger.ToolFacts()) > 0 {
			for _, t := range ledger.ToolFacts() {
				if t.Name != "" {
					tm := getTool(t.Name)
					tm.calls += t.Calls
					tm.errors += t.Errors
				}
			}
		} else if len(ledger.EventFacts()) > 0 {
			for _, e := range ledger.EventFacts() {
				if e.EventType == EventTypeToolCall && e.ToolName != nil && *e.ToolName != "" {
					tm := getTool(*e.ToolName)
					tm.calls++
					if e.Status == "error" {
						tm.errors++
					}
				}
			}
		}
	}

	totalTickets := int64(len(ledgers))
	var anomalyRatePerTicket float64
	if totalTickets > 0 {
		anomalyRatePerTicket = math.Round((float64(totalAnomalies)/float64(totalTickets))*100.0) / 100.0
	}

	brittleTools := make([]BrittleToolSummary, 0)
	for toolName, tm := range toolStats {
		if tm.errors > 0 || tm.anomalies > 0 {
			var failRate float64
			if tm.calls > 0 {
				failRate = math.Round((float64(tm.errors)/float64(tm.calls))*1000.0) / 10.0
			}
			brittleTools = append(brittleTools, BrittleToolSummary{
				ToolName:       toolName,
				AnomalyCount:   tm.anomalies,
				TotalCalls:     tm.calls,
				ErrorCount:     tm.errors,
				FailureRatePct: failRate,
			})
		}
	}

	sort.SliceStable(brittleTools, func(i, j int) bool {
		if brittleTools[i].AnomalyCount != brittleTools[j].AnomalyCount {
			return brittleTools[i].AnomalyCount > brittleTools[j].AnomalyCount
		}
		if brittleTools[i].ErrorCount != brittleTools[j].ErrorCount {
			return brittleTools[i].ErrorCount > brittleTools[j].ErrorCount
		}
		if brittleTools[i].FailureRatePct != brittleTools[j].FailureRatePct {
			return brittleTools[i].FailureRatePct > brittleTools[j].FailureRatePct
		}
		return brittleTools[i].ToolName < brittleTools[j].ToolName
	})

	return FleetAnomalySummary{
		TotalAnomalies:       totalAnomalies,
		WarningCount:         warningCount,
		CriticalCount:        criticalCount,
		SeverityCounts:       severityCounts,
		RuleCounts:           ruleCounts,
		TotalTickets:         totalTickets,
		TicketsWithAnomalies: ticketsWithAnomalies,
		AnomalyRatePerTicket: anomalyRatePerTicket,
		BrittleTools:         brittleTools,
	}
}
