package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// AnomalyContext captures rule-specific parameters, thresholds, and execution context.
type AnomalyContext struct {
	ToolName       *string  `json:"tool_name,omitempty"`
	SessionID      *string  `json:"session_id,omitempty"`
	Turn           *int64   `json:"turn,omitempty"`
	Role           *string  `json:"role,omitempty"`
	SecondaryRole  *string  `json:"secondary_role,omitempty"`
	Count          *int64   `json:"count,omitempty"`
	TotalCalls     *int64   `json:"total_calls,omitempty"`
	TokenDelta     *int64   `json:"token_delta,omitempty"`
	ContextTokens  *int64   `json:"context_tokens,omitempty"`
	Threshold      *string  `json:"threshold,omitempty"`
	FailureRatePct *float64 `json:"failure_rate_pct,omitempty"`
}

// Anomaly represents a single detected execution health issue or operational degradation.
type Anomaly struct {
	RuleID      AnomalyRuleID   `json:"rule_id"`
	Severity    AnomalySeverity `json:"severity"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Context     AnomalyContext  `json:"context"`
}

func strPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

// detectToolRetryStreaks detects consecutive execution failures of the same tool in the event stream,
// strictly scoped per session ID.
func detectToolRetryStreaks(events []EventFact) []Anomaly {
	if len(events) == 0 {
		return nil
	}

	var anomalies []Anomaly
	var activeTool string
	var activeSessionID string
	var consecutiveErrors int64
	var lastTurn int64

	emitStreak := func() {
		if consecutiveErrors >= 3 && activeTool != "" && activeSessionID != "" {
			sev := SeverityWarning
			thresholdStr := "3"
			if consecutiveErrors >= 5 {
				sev = SeverityCritical
				thresholdStr = "5"
			}
			anomaly := Anomaly{
				RuleID:   RuleToolRetryStreak,
				Severity: sev,
				Title:    fmt.Sprintf("Tool retry streak: %s (%d consecutive errors)", activeTool, consecutiveErrors),
				Description: fmt.Sprintf("Tool '%s' encountered %d consecutive execution failures (session: %s, turn: %d)",
					activeTool, consecutiveErrors, activeSessionID, lastTurn),
				Context: AnomalyContext{
					ToolName:  strPtr(activeTool),
					Count:     int64Ptr(consecutiveErrors),
					SessionID: strPtr(activeSessionID),
					Turn:      int64Ptr(lastTurn),
					Threshold: strPtr(thresholdStr),
				},
			}
			anomalies = append(anomalies, anomaly)
		}
	}

	for _, event := range events {
		if event.EventType != EventTypeToolCall || event.ToolName == nil || *event.ToolName == "" {
			continue
		}

		tool := *event.ToolName

		if event.Status == "error" {
			if tool == activeTool && event.SessionID == activeSessionID {
				consecutiveErrors++
				lastTurn = event.Turn
			} else {
				emitStreak()
				activeTool = tool
				activeSessionID = event.SessionID
				consecutiveErrors = 1
				lastTurn = event.Turn
			}
		} else {
			// Status is completed or non-error
			emitStreak()
			activeTool = ""
			activeSessionID = ""
			consecutiveErrors = 0
		}
	}

	emitStreak()
	return anomalies
}

// detectToolHighFailureRates detects tools with an excessively high failure rate across the ticket.
func detectToolHighFailureRates(ledger AgenticFactLedger) []Anomaly {
	type toolStats struct {
		calls  int64
		errors int64
	}

	statsMap := make(map[string]*toolStats)

	if len(ledger.Tools) > 0 {
		for _, t := range ledger.Tools {
			if t.Name == "" {
				continue
			}
			statsMap[t.Name] = &toolStats{
				calls:  t.Calls,
				errors: t.Errors,
			}
		}
	} else if len(ledger.Events) > 0 {
		for _, e := range ledger.Events {
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
			if e.Status == "error" {
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
				Title:    fmt.Sprintf("High tool failure rate: %s (%.1f%%)", name, pct),
				Description: fmt.Sprintf("Tool '%s' failed %d out of %d calls (%.1f%% failure rate, threshold: %s)",
					name, st.errors, st.calls, pct, thresholdStr),
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
func detectContextSaturationSpikes(ledger AgenticFactLedger) []Anomaly {
	var anomalies []Anomaly

	if len(ledger.Events) > 0 {
		sessionEvents := make(map[string][]EventFact)
		for _, e := range ledger.Events {
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

		ceilingReported := make(map[string]bool)

		for _, sid := range sessionIDs {
			events := sessionEvents[sid]
			var hasPrev bool
			var prevTokens int64

			for _, ev := range events {
				// 1. Single-turn jump detection
				if hasPrev {
					delta := ev.ContextTokens - prevTokens
					if delta > 40_000 {
						sev := SeverityWarning
						thresh := "40000"
						if delta > 80_000 {
							sev = SeverityCritical
							thresh = "80000"
						}
						anomaly := Anomaly{
							RuleID:   RuleContextSaturationSpike,
							Severity: sev,
							Title:    fmt.Sprintf("Context saturation spike (+%s tokens)", FormatTokensK(delta)),
							Description: fmt.Sprintf("Context tokens surged by %s tokens in turn %d (session: %s, total: %s tokens, threshold: %s)",
								FormatTokensK(delta), ev.Turn, ev.SessionID, FormatTokensK(ev.ContextTokens), thresh),
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

				// 2. Critical ceiling breach detection (160k tokens)
				if ev.ContextTokens >= 160_000 && !ceilingReported[sid] {
					ceilingReported[sid] = true
					anomaly := Anomaly{
						RuleID:   RuleContextSaturationSpike,
						Severity: SeverityCritical,
						Title:    fmt.Sprintf("Context ceiling breached (%s tokens)", FormatTokensK(ev.ContextTokens)),
						Description: fmt.Sprintf("Context tokens reached %s in turn %d (session: %s), breaching 160k critical ceiling",
							FormatTokensK(ev.ContextTokens), ev.Turn, ev.SessionID),
						Context: AnomalyContext{
							SessionID:     strPtr(ev.SessionID),
							Turn:          int64Ptr(ev.Turn),
							ContextTokens: int64Ptr(ev.ContextTokens),
							Threshold:     strPtr("160000"),
						},
					}
					anomalies = append(anomalies, anomaly)
				}
			}
		}
	} else {
		// v1 Fallback: inspect Invocations for PeakContext >= 160,000
		for _, inv := range ledger.Invocations {
			if inv.Tokens.PeakContext >= 160_000 {
				anomaly := Anomaly{
					RuleID:   RuleContextSaturationSpike,
					Severity: SeverityCritical,
					Title:    fmt.Sprintf("Context ceiling breached (%s tokens)", FormatTokensK(inv.Tokens.PeakContext)),
					Description: fmt.Sprintf("Context tokens reached %s in role '%s' (session: %s), breaching 160k critical ceiling",
						FormatTokensK(inv.Tokens.PeakContext), inv.Role, inv.ID),
					Context: AnomalyContext{
						SessionID:     strPtr(inv.ID),
						Role:          strPtr(inv.Role),
						ContextTokens: int64Ptr(inv.Tokens.PeakContext),
						Threshold:     strPtr("160000"),
					},
				}
				anomalies = append(anomalies, anomaly)
			}
		}
	}

	return anomalies
}

// detectSubagentReworkLoops detects cyclic ping-pong role transitions between worker subagents.
func detectSubagentReworkLoops(ledger AgenticFactLedger) []Anomaly {
	isChildSession := make(map[string]bool)
	for _, inv := range ledger.Invocations {
		if inv.ID != "" {
			isChildSession[inv.ID] = inv.ParentID != nil
		}
	}

	var roles []string

	if len(ledger.Events) > 0 {
		for _, e := range ledger.Events {
			if isChildSession[e.SessionID] {
				r := strings.ToLower(strings.TrimSpace(e.Role))
				if r != "" && (len(roles) == 0 || roles[len(roles)-1] != r) {
					roles = append(roles, r)
				}
			}
		}
	}

	if len(roles) == 0 {
		for _, inv := range ledger.Invocations {
			if inv.ParentID != nil {
				r := strings.ToLower(strings.TrimSpace(inv.Role))
				if r != "" && (len(roles) == 0 || roles[len(roles)-1] != r) {
					roles = append(roles, r)
				}
			}
		}
	}

	var anomalies []Anomaly
	n := len(roles)
	i := 0
	for i < n-1 {
		a := roles[i]
		b := roles[i+1]
		if a == b {
			i++
			continue
		}

		cycles := 0
		pos := i
		for pos+1 < n && roles[pos] == a && roles[pos+1] == b {
			cycles++
			pos += 2
		}

		if cycles >= 3 {
			sev := SeverityWarning
			thresh := "3"
			if cycles >= 5 {
				sev = SeverityCritical
				thresh = "5"
			}
			anomaly := Anomaly{
				RuleID:   RuleSubagentReworkLoop,
				Severity: sev,
				Title:    fmt.Sprintf("Subagent rework loop: %s <-> %s (%d cycles)", a, b, cycles),
				Description: fmt.Sprintf("Subagent roles '%s' and '%s' alternated for %d consecutive cycles indicating thrashing",
					a, b, cycles),
				Context: AnomalyContext{
					Role:          strPtr(a),
					SecondaryRole: strPtr(b),
					Count:         int64Ptr(int64(cycles)),
					Threshold:     strPtr(thresh),
				},
			}
			anomalies = append(anomalies, anomaly)
			i = pos
		} else {
			i++
		}
	}

	return anomalies
}

// DetectAnomalies evaluates deterministic operational rules against the fact ledger.
func DetectAnomalies(ledger AgenticFactLedger) []Anomaly {
	var anomalies []Anomaly

	anomalies = append(anomalies, detectToolRetryStreaks(ledger.Events)...)
	anomalies = append(anomalies, detectToolHighFailureRates(ledger)...)
	anomalies = append(anomalies, detectContextSaturationSpikes(ledger)...)
	anomalies = append(anomalies, detectSubagentReworkLoops(ledger)...)

	if len(anomalies) == 0 {
		return []Anomaly{}
	}

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

		return a.Title < b.Title
	})

	return anomalies
}

// ComputeFleetAnomalies aggregates operational anomalies, severity counts, and brittle tools across all ledgers.
func ComputeFleetAnomalies(ledgers []AgenticFactLedger) FleetAnomalySummary {
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
		ticketAnomalies := DetectAnomalies(ledger)
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

		if len(ledger.Events) > 0 {
			for _, e := range ledger.Events {
				if e.EventType == EventTypeToolCall && e.ToolName != nil && *e.ToolName != "" {
					tm := getTool(*e.ToolName)
					tm.calls++
					if e.Status == "error" {
						tm.errors++
					}
				}
			}
		} else {
			for _, t := range ledger.Tools {
				if t.Name != "" {
					tm := getTool(t.Name)
					tm.calls += t.Calls
					tm.errors += t.Errors
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
