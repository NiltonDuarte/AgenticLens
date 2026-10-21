package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// Canonical live anomaly rule names.
const (
	RuleConsecutiveToolErrors  = "ConsecutiveToolErrors"
	RuleContextCeilingApproach = "ContextCeilingApproach"
	RuleTokenVelocityRunaway   = "TokenVelocityRunaway"
	RuleToolThrashing          = "ToolThrashing"
)

// Live anomaly severity tiers.
const (
	LiveSeverityWarning  = "WARNING"
	LiveSeverityHigh     = "HIGH"
	LiveSeverityCritical = "CRITICAL"
)

// Default thresholds for live anomaly detection rules.
const (
	ConsecutiveToolErrorsStreakWarning   = 3
	ConsecutiveToolErrorsStreakCritical  = 5
	ContextCeilingApproachRatioWarning  = 0.80
	ContextCeilingApproachRatioCritical = 1.00
	TokenVelocityWarningTokens          int64 = 30_000
	TokenVelocityCriticalTokens         int64 = 60_000
	TokenVelocityWarningDeltaTokens     int64 = 40_000
	TokenVelocityCriticalDeltaTokens    int64 = 80_000
	ToolThrashingMinFailures            = 3
	ToolThrashingCriticalFailures       = 5
)

// LiveAnomalyAlert represents an active health degradation or operational failure alert in a live session.
type LiveAnomalyAlert struct {
	SessionID    string `json:"session_id"`
	RuleName     string `json:"rule_name"`
	Severity     string `json:"severity"`
	Message      string `json:"message"`
	Timestamp    string `json:"timestamp"`
	TriggerEvent string `json:"trigger_event"`
	Count        *int64 `json:"count,omitempty"`
}

type toolCallRecord struct {
	toolName  string
	isError   bool
	partID    string
	timestamp string
}

// EvaluateConsecutiveToolErrors evaluates the chronological sequence of tool parts in LiveTraceView.
// It triggers when the count of consecutive tool execution failures reaches or exceeds 3.
// Any successful tool execution immediately resets the consecutive error counter.
func EvaluateConsecutiveToolErrors(trace *LiveTraceView) []LiveAnomalyAlert {
	if trace == nil || len(trace.Messages) == 0 {
		return []LiveAnomalyAlert{}
	}

	streak := 0
	var lastFailingPart *LivePartFact
	var lastFailingTimestamp string

	for _, msg := range trace.Messages {
		for i := range msg.Parts {
			part := &msg.Parts[i]
			if part.Type != "tool" {
				continue
			}
			if part.IsError {
				streak++
				lastFailingPart = part
				lastFailingTimestamp = msg.Timestamp
			} else {
				streak = 0
				lastFailingPart = nil
				lastFailingTimestamp = ""
			}
		}
	}

	if streak < ConsecutiveToolErrorsStreakWarning || lastFailingPart == nil {
		return []LiveAnomalyAlert{}
	}

	severity := LiveSeverityHigh
	if streak >= ConsecutiveToolErrorsStreakCritical {
		severity = LiveSeverityCritical
	}

	toolName := "unknown"
	if lastFailingPart.ToolName != nil && strings.TrimSpace(*lastFailingPart.ToolName) != "" {
		toolName = strings.TrimSpace(*lastFailingPart.ToolName)
	}

	msg := fmt.Sprintf("Tool execution error streak reached %d consecutive failures without recovery (last failing tool: %s).", streak, toolName)

	triggerEvent := lastFailingPart.ID
	var errSnippet string
	if lastFailingPart.ErrorMessage != nil && strings.TrimSpace(*lastFailingPart.ErrorMessage) != "" {
		errSnippet = strings.TrimSpace(*lastFailingPart.ErrorMessage)
	} else if strings.TrimSpace(lastFailingPart.OutputSnippet) != "" {
		errSnippet = strings.TrimSpace(lastFailingPart.OutputSnippet)
	}
	if errSnippet != "" {
		if len(errSnippet) > 256 {
			errSnippet = errSnippet[:256]
		}
		triggerEvent = fmt.Sprintf("%s: %s", lastFailingPart.ID, errSnippet)
	}

	streakCount := int64(streak)
	return []LiveAnomalyAlert{
		{
			SessionID:    trace.SessionID,
			RuleName:     RuleConsecutiveToolErrors,
			Severity:     severity,
			Message:      msg,
			Timestamp:    lastFailingTimestamp,
			TriggerEvent: triggerEvent,
			Count:        &streakCount,
		},
	}
}

// EvaluateContextCeilingApproach warns operators when an active session's context token accumulation
// approaches (>=80%) or breaches (>=100%) the defined healthy context window ceiling.
func EvaluateContextCeilingApproach(trace *LiveTraceView, ceilingTokens int64) []LiveAnomalyAlert {
	if trace == nil || len(trace.Messages) == 0 {
		return []LiveAnomalyAlert{}
	}

	if ceilingTokens <= 0 {
		ceilingTokens = HealthyCeilingTokens
	}

	warnThreshold := int64(float64(ceilingTokens) * ContextCeilingApproachRatioWarning)
	critThreshold := ceilingTokens

	var peakContext int64
	for i := range trace.Messages {
		if trace.Messages[i].ContextTokens > peakContext {
			peakContext = trace.Messages[i].ContextTokens
		}
	}

	if peakContext < warnThreshold {
		return []LiveAnomalyAlert{}
	}

	severity := LiveSeverityWarning
	targetThreshold := warnThreshold
	if peakContext >= critThreshold {
		severity = LiveSeverityCritical
		targetThreshold = critThreshold
	}

	var firstBreachMsg *LiveMessageFact
	for i := range trace.Messages {
		if trace.Messages[i].ContextTokens >= targetThreshold {
			firstBreachMsg = &trace.Messages[i]
			break
		}
	}

	pct := math.Round(float64(peakContext) / float64(ceilingTokens) * 100.0)
	ceilingStr := FormatTokensK(ceilingTokens)
	msg := fmt.Sprintf("Context window utilization reached %s tokens (%.0f%% of healthy ceiling %s tokens).", FormatTokensK(peakContext), pct, ceilingStr)

	var triggerEvent, ts string
	if firstBreachMsg != nil {
		ts = firstBreachMsg.Timestamp
		if firstBreachMsg.ID != "" {
			triggerEvent = firstBreachMsg.ID
		} else {
			triggerEvent = firstBreachMsg.Timestamp
		}
	}

	return []LiveAnomalyAlert{
		{
			SessionID:    trace.SessionID,
			RuleName:     RuleContextCeilingApproach,
			Severity:     severity,
			Message:      msg,
			Timestamp:    ts,
			TriggerEvent: triggerEvent,
		},
	}
}

// EvaluateTokenVelocityRunaway detects sudden token consumption surges in a single turn.
func EvaluateTokenVelocityRunaway(trace *LiveTraceView, warningThreshold int64, criticalThreshold int64) []LiveAnomalyAlert {
	if trace == nil || len(trace.Messages) == 0 {
		return []LiveAnomalyAlert{}
	}

	if warningThreshold <= 0 {
		warningThreshold = TokenVelocityWarningTokens
	}
	if criticalThreshold <= 0 {
		criticalThreshold = TokenVelocityCriticalTokens
	}

	warnDeltaThreshold := warningThreshold * 4 / 3
	if warnDeltaThreshold < warningThreshold {
		warnDeltaThreshold = TokenVelocityWarningDeltaTokens
	}
	critDeltaThreshold := criticalThreshold * 4 / 3
	if critDeltaThreshold < criticalThreshold {
		critDeltaThreshold = TokenVelocityCriticalDeltaTokens
	}

	var alerts []LiveAnomalyAlert
	var prevContext int64
	var hasPrev bool

	for i := range trace.Messages {
		msg := &trace.Messages[i]
		if msg.Role != "" && msg.Role != "assistant" {
			continue
		}
		turnNum := i + 1

		var delta int64
		if hasPrev {
			delta = msg.ContextTokens - prevContext
			if delta < 0 {
				delta = 0
			}
		}

		var severity string
		var surgingTokens int64
		var thresholdUsed int64

		if msg.TotalTokens > criticalThreshold || (hasPrev && delta > critDeltaThreshold) {
			severity = LiveSeverityCritical
			if msg.TotalTokens > criticalThreshold {
				surgingTokens = msg.TotalTokens
				thresholdUsed = criticalThreshold
			} else {
				surgingTokens = delta
				thresholdUsed = critDeltaThreshold
			}
		} else if msg.TotalTokens > warningThreshold || (hasPrev && delta > warnDeltaThreshold) {
			severity = LiveSeverityWarning
			if msg.TotalTokens > warningThreshold {
				surgingTokens = msg.TotalTokens
				thresholdUsed = warningThreshold
			} else {
				surgingTokens = delta
				thresholdUsed = warnDeltaThreshold
			}
		}

		if severity != "" {
			alertMsg := fmt.Sprintf("Abnormally high token velocity of %s tokens consumed in turn %d (threshold: %s tokens).", FormatTokensK(surgingTokens), turnNum, FormatTokensK(thresholdUsed))
			trigger := fmt.Sprintf("%s (turn %d)", msg.ID, turnNum)
			if msg.ID == "" {
				trigger = fmt.Sprintf("turn %d", turnNum)
			}
			alerts = append(alerts, LiveAnomalyAlert{
				SessionID:    trace.SessionID,
				RuleName:     RuleTokenVelocityRunaway,
				Severity:     severity,
				Message:      alertMsg,
				Timestamp:    msg.Timestamp,
				TriggerEvent: trigger,
			})
		}

		prevContext = msg.ContextTokens
		hasPrev = true
	}

	if len(alerts) == 0 {
		return []LiveAnomalyAlert{}
	}
	return alerts
}

// EvaluateToolThrashing evaluates the sequence of tool calls in LiveTraceView, tracking repeat invocations
// of the identical tool name where executions fail without intermediate success.
func EvaluateToolThrashing(trace *LiveTraceView) []LiveAnomalyAlert {
	if trace == nil || len(trace.Messages) == 0 {
		return []LiveAnomalyAlert{}
	}

	var calls []toolCallRecord
	for _, msg := range trace.Messages {
		for _, part := range msg.Parts {
			if part.Type != "tool" {
				continue
			}
			tName := "unknown"
			if part.ToolName != nil && strings.TrimSpace(*part.ToolName) != "" {
				tName = strings.TrimSpace(*part.ToolName)
			}
			calls = append(calls, toolCallRecord{
				toolName:  tName,
				isError:   part.IsError,
				partID:    part.ID,
				timestamp: msg.Timestamp,
			})
		}
	}

	if len(calls) == 0 {
		return []LiveAnomalyAlert{}
	}

	type thrashCandidate struct {
		toolName  string
		failCount int
		partIDs   []string
		timestamp string
	}

	candidates := make(map[string]thrashCandidate)

	// Check 1: Consecutive failure streak of the exact same tool ending at the end of the trace
	lastCall := calls[len(calls)-1]
	if lastCall.isError {
		consecStreak := 0
		var consecParts []string
		var consecTs string

		for i := len(calls) - 1; i >= 0; i-- {
			if calls[i].toolName == lastCall.toolName && calls[i].isError {
				consecStreak++
				consecParts = append(consecParts, calls[i].partID)
				if consecTs == "" {
					consecTs = calls[i].timestamp
				}
			} else {
				break
			}
		}

		if consecStreak >= ToolThrashingMinFailures {
			// Reverse part IDs to chronological order
			for i, j := 0, len(consecParts)-1; i < j; i, j = i+1, j-1 {
				consecParts[i], consecParts[j] = consecParts[j], consecParts[i]
			}
			candidates[lastCall.toolName] = thrashCandidate{
				toolName:  lastCall.toolName,
				failCount: consecStreak,
				partIDs:   consecParts,
				timestamp: consecTs,
			}
		}
	}

	// Check 2: Within the most recent window of 5 tool calls
	windowStart := len(calls) - 5
	if windowStart < 0 {
		windowStart = 0
	}
	recentWindow := calls[windowStart:]

	// Find distinct tools in recent window
	toolSet := make(map[string]bool)
	for _, c := range recentWindow {
		toolSet[c.toolName] = true
	}

	for t := range toolSet {
		// Collect calls for tool t in the recent window
		var failuresSinceSuccess []string
		var lastFailTs string
		hadInterveningSuccess := false

		for _, c := range recentWindow {
			if c.toolName != t {
				continue
			}
			if c.isError {
				failuresSinceSuccess = append(failuresSinceSuccess, c.partID)
				lastFailTs = c.timestamp
			} else {
				// Intervening success resets failure accumulation for this tool in the window
				hadInterveningSuccess = true
				failuresSinceSuccess = nil
			}
		}

		if !hadInterveningSuccess && len(failuresSinceSuccess) >= ToolThrashingMinFailures {
			existing, exists := candidates[t]
			if !exists || len(failuresSinceSuccess) > existing.failCount {
				candidates[t] = thrashCandidate{
					toolName:  t,
					failCount: len(failuresSinceSuccess),
					partIDs:   failuresSinceSuccess,
					timestamp: lastFailTs,
				}
			}
		}
	}

	if len(candidates) == 0 {
		return []LiveAnomalyAlert{}
	}

	// Sort candidate tool names deterministically
	toolNames := make([]string, 0, len(candidates))
	for t := range candidates {
		toolNames = append(toolNames, t)
	}
	sort.Strings(toolNames)

	var alerts []LiveAnomalyAlert
	for _, t := range toolNames {
		cand := candidates[t]
		severity := LiveSeverityWarning
		if cand.failCount >= ToolThrashingCriticalFailures {
			severity = LiveSeverityCritical
		}

		msg := fmt.Sprintf("Tool thrashing detected on tool %s with %d failing invocations without progress.", cand.toolName, cand.failCount)
		trigger := fmt.Sprintf("%s: %s", cand.toolName, strings.Join(cand.partIDs, ", "))
		failCount := int64(cand.failCount)

		alerts = append(alerts, LiveAnomalyAlert{
			SessionID:    trace.SessionID,
			RuleName:     RuleToolThrashing,
			Severity:     severity,
			Message:      msg,
			Timestamp:    cand.timestamp,
			TriggerEvent: trigger,
			Count:        &failCount,
		})
	}

	return alerts
}

func liveSeverityRank(sev string) int {
	switch sev {
	case LiveSeverityCritical:
		return 3
	case LiveSeverityHigh:
		return 2
	case LiveSeverityWarning:
		return 1
	default:
		return 0
	}
}

// SortLiveAnomalies sorts live anomaly alerts deterministically:
// Primary sort: Severity descending (CRITICAL > HIGH > WARNING).
// Secondary sort: RuleName ascending.
// Tertiary sort: Timestamp descending.
// Quaternary sort: SessionID ascending, then TriggerEvent ascending.
func SortLiveAnomalies(alerts []LiveAnomalyAlert) {
	sort.SliceStable(alerts, func(i, j int) bool {
		rI := liveSeverityRank(alerts[i].Severity)
		rJ := liveSeverityRank(alerts[j].Severity)
		if rI != rJ {
			return rI > rJ
		}
		if alerts[i].RuleName != alerts[j].RuleName {
			return alerts[i].RuleName < alerts[j].RuleName
		}
		if alerts[i].Timestamp != alerts[j].Timestamp {
			return alerts[i].Timestamp > alerts[j].Timestamp
		}
		if alerts[i].SessionID != alerts[j].SessionID {
			return alerts[i].SessionID < alerts[j].SessionID
		}
		return alerts[i].TriggerEvent < alerts[j].TriggerEvent
	})
}

// EvaluateLiveAnomalies is the master composite evaluator evaluating all active health rules.
// Safely handles nil tree or nil trace arguments.
// Returns a deterministically ordered slice sorted primarily by Severity descending,
// secondarily by RuleName ascending, and tertiary by Timestamp descending.
func EvaluateLiveAnomalies(tree *LiveSessionTree, trace *LiveTraceView) []LiveAnomalyAlert {
	var collected []LiveAnomalyAlert

	if trace != nil {
		collected = append(collected, EvaluateConsecutiveToolErrors(trace)...)
		collected = append(collected, EvaluateContextCeilingApproach(trace, HealthyCeilingTokens)...)
		collected = append(collected, EvaluateTokenVelocityRunaway(trace, TokenVelocityWarningTokens, TokenVelocityCriticalTokens)...)
		collected = append(collected, EvaluateToolThrashing(trace)...)
	}

	if tree != nil {
		collected = append(collected, tree.ActiveAnomalies...)
		var collectFromNodes func(nodes []*LiveSessionNode)
		collectFromNodes = func(nodes []*LiveSessionNode) {
			for _, n := range nodes {
				if n == nil {
					continue
				}
				collected = append(collected, n.Anomalies...)
				collectFromNodes(n.SubSessions)
			}
		}
		collectFromNodes(tree.RootSessions)
	}

	if len(collected) == 0 {
		return []LiveAnomalyAlert{}
	}

	// Deduplicate identical alerts
	seen := make(map[string]bool, len(collected))
	unique := make([]LiveAnomalyAlert, 0, len(collected))
	for _, a := range collected {
		key := fmt.Sprintf("%s|%s|%s|%s|%s", a.SessionID, a.RuleName, a.Severity, a.Timestamp, a.TriggerEvent)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, a)
		}
	}

	SortLiveAnomalies(unique)
	return unique
}
