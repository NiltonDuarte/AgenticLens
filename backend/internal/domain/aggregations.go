package domain

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ticketIDPattern = regexp.MustCompile(`^([a-zA-Z]+)[-_]?([0-9]+)(.*)$`)

// CompareTicketIDs compares two ticket identifiers using natural numerical sequence ordering.
// Conforming ticket identifiers have a letter prefix, optional hyphen/underscore, and numeric sequence.
// Conforming tickets sort before non-conforming strings.
// Returns -1 if a < b, 0 if a == b, and 1 if a > b.
func CompareTicketIDs(a, b string) int {
	cleanA := strings.TrimSpace(a)
	cleanB := strings.TrimSpace(b)

	if cleanA == cleanB {
		return 0
	}
	if cleanA == "" {
		return -1
	}
	if cleanB == "" {
		return 1
	}

	matchA := ticketIDPattern.FindStringSubmatch(cleanA)
	matchB := ticketIDPattern.FindStringSubmatch(cleanB)

	hasMatchA := len(matchA) == 4
	hasMatchB := len(matchB) == 4

	if hasMatchA && hasMatchB {
		prefixA := strings.ToUpper(matchA[1])
		prefixB := strings.ToUpper(matchB[1])
		if prefixA != prefixB {
			if prefixA < prefixB {
				return -1
			}
			return 1
		}

		numA, errA := strconv.ParseInt(matchA[2], 10, 64)
		numB, errB := strconv.ParseInt(matchB[2], 10, 64)
		if errA == nil && errB == nil {
			if numA != numB {
				if numA < numB {
					return -1
				}
				return 1
			}
			// When numerical values are equal (e.g. AL-1 vs AL-01), fall back to raw string comparison.
			return strings.Compare(cleanA, cleanB)
		}
	}

	if hasMatchA && !hasMatchB {
		return -1
	}
	if !hasMatchA && hasMatchB {
		return 1
	}

	return strings.Compare(cleanA, cleanB)
}

type invocationAccumulator struct {
	invocations       int64
	inputTokens       int64
	outputTokens      int64
	cacheReadTokens   int64
	cacheHitRatePct   float64
	costUSD           float64
	peakContextTokens int64
}

func accumulateInvocations(invs []InvocationFact) invocationAccumulator {
	var inp, out, cache, peak int64
	var cost float64

	for _, inv := range invs {
		inp += inv.Tokens.Input
		out += inv.Tokens.Output
		cache += inv.Tokens.CacheRead
		cost += inv.CostUSD
		if inv.Tokens.PeakContext > peak {
			peak = inv.Tokens.PeakContext
		}
	}

	totTok := inp + cache
	var hitPct float64
	if totTok > 0 {
		hitPct = math.Round(float64(cache)/float64(totTok)*1000.0) / 10.0
	}

	return invocationAccumulator{
		invocations:       int64(len(invs)),
		inputTokens:       inp,
		outputTokens:      out,
		cacheReadTokens:   cache,
		cacheHitRatePct:   hitPct,
		costUSD:           math.Round(cost*10000.0) / 10000.0,
		peakContextTokens: peak,
	}
}

// ComputeTotals calculates aggregate invocation metrics from a slice of invocation facts.
func ComputeTotals(invocations []InvocationFact, ceiling ...int64) GroupedMetric {
	effCeiling := ContextCeilingCriticalTokens
	if len(ceiling) > 0 && ceiling[0] > 0 {
		effCeiling = ceiling[0]
	}

	if len(invocations) == 0 {
		return GroupedMetric{
			Invocations:       0,
			InputTokens:       0,
			OutputTokens:      0,
			CacheReadTokens:   0,
			CacheHitRatePct:   0.0,
			CostUSD:           0.0,
			PeakContextTokens: 0,
			ContextHealth:     ContextHealthHealthy,
		}
	}

	byRole := make(map[AgentRole][]InvocationFact)
	for _, inv := range invocations {
		byRole[inv.Role] = append(byRole[inv.Role], inv)
	}

	var totalInvocations int64
	for _, rInvs := range byRole {
		totalInvocations += accumulateInvocations(rInvs).invocations
	}

	acc := accumulateInvocations(invocations)
	healthStatus := GetBaseHealthStatus(acc.peakContextTokens, effCeiling)

	return GroupedMetric{
		Invocations:       totalInvocations,
		InputTokens:       acc.inputTokens,
		OutputTokens:      acc.outputTokens,
		CacheReadTokens:   acc.cacheReadTokens,
		CacheHitRatePct:   acc.cacheHitRatePct,
		CostUSD:           acc.costUSD,
		PeakContextTokens: acc.peakContextTokens,
		ContextHealth:     healthStatus,
	}
}

// ComputeWorkflowEfficiency derives workflow efficiency profile from an atomic fact ledger.
func ComputeWorkflowEfficiency(ledger AgenticFactLedger, roleOrder ...[]string) WorkflowEfficiency {
	return ComputeWorkflowEfficiencyWithCeiling(ledger, ContextCeilingCriticalTokens, roleOrder...)
}

// ComputeWorkflowEfficiencyWithConfig derives workflow efficiency using project configuration.
func ComputeWorkflowEfficiencyWithConfig(ledger AgenticFactLedger, cfg ProjectConfig) WorkflowEfficiency {
	return ComputeWorkflowEfficiencyWithCeiling(ledger, cfg.EffectiveHealthyCeiling(), cfg.RoleOrder)
}

// ComputeWorkflowEfficiencyWithCeiling derives workflow efficiency profile with an explicit context ceiling.
func ComputeWorkflowEfficiencyWithCeiling(ledger AgenticFactLedger, ceiling int64, roleOrder ...[]string) WorkflowEfficiency {
	if ceiling <= 0 {
		ceiling = ContextCeilingCriticalTokens
	}
	tools := ledger.ToolFacts()
	code := ledger.Code

	var activeDurationSeconds int64
	var totReasoning, totOutput, totCompactions int64
	rolePeaks := make(map[AgentRole]int64)

	var rootPeak int64
	var hasRoot bool

	subCosts := make(map[string]float64)
	var totSubCost float64
	subPeaks := make(map[AgentRole]int64)
	roleTurns := make(map[AgentRole][]int64)

	for _, session := range ledger.Sessions {
		isRoot := session.ParentID == nil || strings.TrimSpace(*session.ParentID) == ""
		for _, inv := range session.Invocations {
			activeDurationSeconds += inv.ActiveDurationSeconds
			totReasoning += inv.Tokens.Reasoning
			totOutput += inv.Tokens.Output
			totCompactions += inv.Compactions

			if curr, exists := rolePeaks[inv.Role]; !exists || inv.Tokens.PeakContext > curr {
				rolePeaks[inv.Role] = inv.Tokens.PeakContext
			}

			if isRoot {
				if !hasRoot || inv.Tokens.PeakContext > rootPeak {
					rootPeak = inv.Tokens.PeakContext
				}
				hasRoot = true
			} else {
				role := inv.Role
				if strings.TrimSpace(string(role)) == "" {
					role = RoleUnassignedSubAgent
				}
				subCosts[string(role)] += inv.CostUSD
				totSubCost += inv.CostUSD
				if curr, exists := subPeaks[role]; !exists || inv.Tokens.PeakContext > curr {
					subPeaks[role] = inv.Tokens.PeakContext
				}
				roleTurns[role] = append(roleTurns[role], inv.Turns)
			}
		}
	}

	durHours := activeDurationSeconds / 3600
	durRem := activeDurationSeconds % 3600
	durMins := durRem / 60
	durRemSecs := durRem % 60

	var activeDurationFormatted string
	if durHours > 0 {
		activeDurationFormatted = fmt.Sprintf("%dh %dm", durHours, durMins)
	} else {
		activeDurationFormatted = fmt.Sprintf("%dm %ds", durMins, durRemSecs)
	}

	roleEffortDistributionPct := make(map[string]float64)
	if totSubCost > 0 && len(subCosts) > 0 {
		for r, c := range subCosts {
			if c > 0 {
				roleEffortDistributionPct[r] = math.Round(c/totSubCost*1000.0) / 10.0
			}
		}
	}

	var customOrder []AgentRole
	if len(roleOrder) > 0 && len(roleOrder[0]) > 0 {
		customOrder = make([]AgentRole, len(roleOrder[0]))
		for i, r := range roleOrder[0] {
			customOrder[i] = AgentRole(r)
		}
	}

	var allSubTurns []int64
	var subTurnRoles []AgentRole
	for r := range roleTurns {
		subTurnRoles = append(subTurnRoles, r)
	}
	var orderedRolesForTurns []AgentRole
	if len(customOrder) > 0 {
		orderedRolesForTurns = OrderSubagentRoles(subTurnRoles, customOrder)
	} else {
		orderedRolesForTurns = OrderSubagentRoles(subTurnRoles)
	}
	orderedSet := make(map[AgentRole]struct{}, len(orderedRolesForTurns))
	for _, r := range orderedRolesForTurns {
		orderedSet[r] = struct{}{}
		if turns, ok := roleTurns[r]; ok {
			allSubTurns = append(allSubTurns, turns...)
		}
	}
	sort.Slice(subTurnRoles, func(i, j int) bool {
		return subTurnRoles[i] < subTurnRoles[j]
	})
	for _, r := range subTurnRoles {
		if _, ok := orderedSet[r]; !ok {
			allSubTurns = append(allSubTurns, roleTurns[r]...)
		}
	}

	var avgSubTurns float64
	var peakSubTurns int64
	if len(allSubTurns) > 0 {
		var sumSubTurns int64
		for _, t := range allSubTurns {
			sumSubTurns += t
			if t > peakSubTurns {
				peakSubTurns = t
			}
		}
		avgSubTurns = math.Round(float64(sumSubTurns)/float64(len(allSubTurns))*10.0) / 10.0
	}

	velByRole := make(map[AgentRole]float64)
	for r, turns := range roleTurns {
		if len(turns) > 0 {
			var rSum int64
			for _, t := range turns {
				rSum += t
			}
			velByRole[r] = math.Round(float64(rSum)/float64(len(turns))*10.0) / 10.0
		}
	}

	subagentTurnVelocity := SubagentTurnVelocity{
		AvgTurnsPerInvocation: avgSubTurns,
		PeakTurns:             peakSubTurns,
		ByRole:                velByRole,
	}

	var testRatioStr string
	if code.ProdLinesAdded > 0 {
		testRatioStr = fmt.Sprintf("%.1f:1", float64(code.TestLinesAdded)/float64(code.ProdLinesAdded))
	} else {
		testRatioStr = fmt.Sprintf("%d:0", code.TestLinesAdded)
	}

	testToProductionRatio := TestToCodeRatio{
		Ratio:          testRatioStr,
		TestLinesAdded: code.TestLinesAdded,
		ProdLinesAdded: code.ProdLinesAdded,
	}

	genTotal := totOutput + totReasoning
	var reasoningRatioPct float64
	if genTotal > 0 {
		reasoningRatioPct = math.Round(float64(totReasoning)/float64(genTotal)*1000.0) / 10.0
	}

	codeDelivered := CodeDeliveredStats{
		LinesAdded:   code.LinesAdded,
		LinesDeleted: code.LinesDeleted,
		FilesChanged: code.FilesChanged,
	}

	var totToolCalls, totToolErrors int64
	for _, t := range tools {
		totToolCalls += t.Calls
		totToolErrors += t.Errors
	}

	toolSuccessRate := 100.0
	if totToolCalls > 0 {
		toolSuccessRate = math.Round(float64(totToolCalls-totToolErrors)/float64(totToolCalls)*1000.0) / 10.0
	}

	toolExecution := ToolExecutionHealth{
		TotalCalls:           totToolCalls,
		ErrorsAndCorrections: totToolErrors,
		SuccessRatePct:       toolSuccessRate,
	}

	var rootPeakVal int64
	if hasRoot {
		rootPeakVal = rootPeak
	}
	var rootSatPct float64
	if ceiling > 0 {
		rootSatPct = math.Round(float64(rootPeakVal)/float64(ceiling)*1000.0) / 10.0
	}

	var subPeak int64
	subRoleStr := AgentRole("none")
	var foundSubRole bool

	if len(subPeaks) > 0 {
		var candidateSubRoles []AgentRole
		for r := range subPeaks {
			candidateSubRoles = append(candidateSubRoles, r)
		}
		var orderedCandidateRoles []AgentRole
		if len(customOrder) > 0 {
			orderedCandidateRoles = OrderSubagentRoles(candidateSubRoles, customOrder)
		} else {
			orderedCandidateRoles = OrderSubagentRoles(candidateSubRoles)
		}
		for _, r := range orderedCandidateRoles {
			peak := subPeaks[r]
			if !foundSubRole || peak > subPeak {
				subPeak = peak
				subRoleStr = r
				foundSubRole = true
			}
		}
		if !foundSubRole {
			sort.Slice(candidateSubRoles, func(i, j int) bool {
				return candidateSubRoles[i] < candidateSubRoles[j]
			})
			for _, r := range candidateSubRoles {
				peak := subPeaks[r]
				if !foundSubRole || peak > subPeak {
					subPeak = peak
					subRoleStr = r
					foundSubRole = true
				}
			}
		}
	}

	overallPeak := rootPeakVal
	if subPeak > overallPeak {
		overallPeak = subPeak
	}
	cwhStatus := GetBaseHealthStatus(overallPeak, ceiling)

	contextWindowHealth := ContextWindowHealth{
		HealthyCeilingTokens: ceiling,
		Status:               cwhStatus,
		RootPeakTokens:       rootPeakVal,
		RootSaturationPct:    rootSatPct,
		SubagentsPeakTokens:  subPeak,
		SubagentsPeakRole:    subRoleStr,
		Compactions:          totCompactions,
		PeakByRole:           rolePeaks,
	}

	return WorkflowEfficiency{
		SDLCIterations:            code.SDLCIterations,
		ActiveDurationSeconds:     activeDurationSeconds,
		ActiveDurationFormatted:   activeDurationFormatted,
		RoleEffortDistributionPct: roleEffortDistributionPct,
		SubagentTurnVelocity:      subagentTurnVelocity,
		TestToProductionRatio:     testToProductionRatio,
		ReasoningTokens:           totReasoning,
		ReasoningRatioPct:         reasoningRatioPct,
		CodeDelivered:             codeDelivered,
		ToolExecution:             toolExecution,
		ContextWindowHealth:       contextWindowHealth,
	}
}

// ComputeRoleMetrics calculates aggregated metrics grouped by agent role using default ceiling.
func ComputeRoleMetrics(invocations []InvocationFact, roleOrder ...[]string) []RoleMetric {
	return ComputeRoleMetricsWithCeiling(invocations, ContextCeilingCriticalTokens, roleOrder...)
}

// ComputeRoleMetricsWithCeiling calculates aggregated metrics grouped by agent role with an explicit context ceiling.
func ComputeRoleMetricsWithCeiling(invocations []InvocationFact, ceiling int64, roleOrder ...[]string) []RoleMetric {
	if ceiling <= 0 {
		ceiling = ContextCeilingCriticalTokens
	}

	byRole := make(map[AgentRole][]InvocationFact)
	for _, inv := range invocations {
		byRole[inv.Role] = append(byRole[inv.Role], inv)
	}

	var roles []AgentRole
	for r := range byRole {
		roles = append(roles, r)
	}

	var customOrder []AgentRole
	if len(roleOrder) > 0 && len(roleOrder[0]) > 0 {
		customOrder = make([]AgentRole, len(roleOrder[0]))
		for i, r := range roleOrder[0] {
			customOrder[i] = AgentRole(r)
		}
	}

	var orderedRoles []AgentRole
	if len(customOrder) > 0 {
		orderedRoles = OrderRoles(roles, customOrder)
	} else {
		orderedRoles = OrderRoles(roles)
	}

	var metrics []RoleMetric
	for _, role := range orderedRoles {
		acc := accumulateInvocations(byRole[role])
		health := GetBaseHealthStatus(acc.peakContextTokens, ceiling)

		metrics = append(metrics, RoleMetric{
			Role:              role,
			Invocations:       acc.invocations,
			InputTokens:       acc.inputTokens,
			OutputTokens:      acc.outputTokens,
			CacheReadTokens:   acc.cacheReadTokens,
			CacheHitRatePct:   acc.cacheHitRatePct,
			CostUSD:           acc.costUSD,
			PeakContextTokens: acc.peakContextTokens,
			ContextHealth:     health,
		})
	}

	return metrics
}

// ComputeModelMetrics calculates aggregated metrics grouped by model, sorted by cost descending
// with secondary tie-breaking on model name ascending.
func ComputeModelMetrics(invocations []InvocationFact, ceiling ...int64) []ModelMetric {
	effCeiling := ContextCeilingCriticalTokens
	if len(ceiling) > 0 && ceiling[0] > 0 {
		effCeiling = ceiling[0]
	}

	byModel := make(map[string][]InvocationFact)
	for _, inv := range invocations {
		byModel[inv.Model] = append(byModel[inv.Model], inv)
	}

	var metrics []ModelMetric
	for model, mInvs := range byModel {
		acc := accumulateInvocations(mInvs)
		health := GetBaseHealthStatus(acc.peakContextTokens, effCeiling)

		metrics = append(metrics, ModelMetric{
			Model:             model,
			Invocations:       acc.invocations,
			InputTokens:       acc.inputTokens,
			OutputTokens:      acc.outputTokens,
			CacheReadTokens:   acc.cacheReadTokens,
			CacheHitRatePct:   acc.cacheHitRatePct,
			CostUSD:           acc.costUSD,
			PeakContextTokens: acc.peakContextTokens,
			ContextHealth:     health,
		})
	}

	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].CostUSD != metrics[j].CostUSD {
			return metrics[i].CostUSD > metrics[j].CostUSD
		}
		return metrics[i].Model < metrics[j].Model
	})

	return metrics
}

// ComputeThreadSummaries groups session-owned invocation rows by their root session.
func ComputeThreadSummaries(sessions []SessionFact, roleOrder ...[]string) []ThreadExecutionRecord {
	return ComputeThreadSummariesWithCeiling(sessions, ContextCeilingCriticalTokens, roleOrder...)
}

// ComputeThreadSummariesFromSessions groups sessions by root session into hierarchical thread execution records.
func ComputeThreadSummariesFromSessions(sessions []SessionFact, roleOrder ...[]string) []ThreadExecutionRecord {
	return ComputeThreadSummaries(sessions, roleOrder...)
}

// ComputeThreadSummariesWithCeiling groups session-owned invocation rows by their root session with an explicit context ceiling.
func ComputeThreadSummariesWithCeiling(sessions []SessionFact, ceiling int64, roleOrder ...[]string) []ThreadExecutionRecord {
	if ceiling <= 0 {
		ceiling = ContextCeilingCriticalTokens
	}

	if len(sessions) == 0 {
		return []ThreadExecutionRecord{}
	}

	sessionByID := make(map[string]SessionFact, len(sessions))
	for _, s := range sessions {
		sessionByID[s.ID] = s
	}

	resolveRoot := func(s SessionFact) string {
		curr := s.ID
		visited := make(map[string]struct{})
		for {
			sess, ok := sessionByID[curr]
			if !ok || sess.ParentID == nil || strings.TrimSpace(*sess.ParentID) == "" {
				break
			}
			if _, seen := visited[curr]; seen {
				break
			}
			visited[curr] = struct{}{}
			curr = strings.TrimSpace(*sess.ParentID)
		}
		return curr
	}

	threadGroups := make(map[string][]InvocationFact)
	threadUsers := make(map[string]string)
	for _, s := range sessions {
		rootID := resolveRoot(s)
		threadGroups[rootID] = append(threadGroups[rootID], s.Invocations...)
		if s.ID == rootID && s.User != "" {
			threadUsers[rootID] = s.User
		} else if threadUsers[rootID] == "" && s.User != "" {
			threadUsers[rootID] = s.User
		}
	}

	var records []ThreadExecutionRecord
	for rootID, threadInvs := range threadGroups {
		threadUser := threadUsers[rootID]
		if threadUser == "" {
			threadUser = "unknown"
		}

		var compactions int64
		byRole := make(map[AgentRole][]InvocationFact)
		for _, inv := range threadInvs {
			compactions += inv.Compactions
			byRole[inv.Role] = append(byRole[inv.Role], inv)
		}

		var threadRoles []AgentRole
		for r := range byRole {
			threadRoles = append(threadRoles, r)
		}

		var customOrder []AgentRole
		if len(roleOrder) > 0 && len(roleOrder[0]) > 0 {
			customOrder = make([]AgentRole, len(roleOrder[0]))
			for i, r := range roleOrder[0] {
				customOrder[i] = AgentRole(r)
			}
		}

		var orderedRoles []AgentRole
		if len(customOrder) > 0 {
			orderedRoles = OrderRoles(threadRoles, customOrder)
		} else {
			orderedRoles = OrderRoles(threadRoles)
		}

		var roleMetrics []ThreadRoleMetric
		for _, role := range orderedRoles {
			rInvs := byRole[role]
			byModel := make(map[string][]InvocationFact)
			for _, inv := range rInvs {
				byModel[inv.Model] = append(byModel[inv.Model], inv)
			}

			var modelMetrics []ModelMetric
			for model, mInvs := range byModel {
				var inp, out, cache int64
				var cost float64
				var peak int64

				for _, inv := range mInvs {
					inp += inv.Tokens.Input
					out += inv.Tokens.Output
					cache += inv.Tokens.CacheRead
					cost += inv.CostUSD
					if inv.Tokens.PeakContext > peak {
						peak = inv.Tokens.PeakContext
					}
				}

				totTok := inp + cache
				var hitPct float64
				if totTok > 0 {
					hitPct = math.Round(float64(cache)/float64(totTok)*1000.0) / 10.0
				}
				health := GetBaseHealthStatus(peak, ceiling)

				modelMetrics = append(modelMetrics, ModelMetric{
					Model:             model,
					Invocations:       int64(len(mInvs)),
					InputTokens:       inp,
					OutputTokens:      out,
					CacheReadTokens:   cache,
					CacheHitRatePct:   hitPct,
					CostUSD:           math.Round(cost*10000.0) / 10000.0,
					PeakContextTokens: peak,
					ContextHealth:     health,
				})
			}

			sort.Slice(modelMetrics, func(i, j int) bool {
				if modelMetrics[i].CostUSD != modelMetrics[j].CostUSD {
					return modelMetrics[i].CostUSD > modelMetrics[j].CostUSD
				}
				return modelMetrics[i].Model < modelMetrics[j].Model
			})

			roleMetrics = append(roleMetrics, ThreadRoleMetric{
				Role:   role,
				Models: modelMetrics,
			})
		}

		records = append(records, ThreadExecutionRecord{
			SessionID:   rootID,
			User:        threadUser,
			Roles:       roleMetrics,
			Compactions: compactions,
		})
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].SessionID < records[j].SessionID
	})

	return records
}

// DeriveMetricsLedger projects all aggregations from an atomic fact ledger using full project configuration.
func DeriveMetricsLedger(factLedger AgenticFactLedger, cfg ProjectConfig) AgenticMetricsLedger {
	mappings := cfg.RoleMappings
	if mappings == nil {
		mappings = RoleMappings{}
	}

	invocations := factLedger.InvocationFacts()
	mappedInvocations := make([]InvocationFact, len(invocations))
	for i, inv := range invocations {
		mappedInvocations[i] = inv
		mappedInvocations[i].Role = mappings.ResolveRole(string(inv.Role))
	}

	events := factLedger.EventFacts()
	var mappedEvents []EventFact
	if events != nil {
		mappedEvents = make([]EventFact, len(events))
		for i, ev := range events {
			mappedEvents[i] = ev
			mappedEvents[i].Role = mappings.ResolveRole(string(ev.Role))
		}
	}

	mappedFactLedger := factLedger
	mappedFactLedger.Sessions = make([]SessionFact, len(factLedger.Sessions))
	for i, session := range factLedger.Sessions {
		mappedSession := session
		mappedSession.Invocations = make([]InvocationFact, len(session.Invocations))
		for j, inv := range session.Invocations {
			mappedSession.Invocations[j] = inv
			mappedSession.Invocations[j].Role = mappings.ResolveRole(string(inv.Role))
		}
		if session.Events != nil {
			mappedSession.Events = make([]SessionEventFact, len(session.Events))
			for k, ev := range session.Events {
				mappedSession.Events[k] = ev
				mappedSession.Events[k].Role = mappings.ResolveRole(string(ev.Role))
			}
		}
		mappedFactLedger.Sessions[i] = mappedSession
	}

	effectiveCeiling := cfg.EffectiveHealthyCeiling()

	totals := ComputeTotals(mappedInvocations, effectiveCeiling)
	eff := ComputeWorkflowEfficiencyWithCeiling(mappedFactLedger, effectiveCeiling, cfg.RoleOrder)
	roles := ComputeRoleMetricsWithCeiling(mappedInvocations, effectiveCeiling, cfg.RoleOrder)
	models := ComputeModelMetrics(mappedInvocations, effectiveCeiling)
	threads := ComputeThreadSummariesWithCeiling(mappedFactLedger.Sessions, effectiveCeiling, cfg.RoleOrder)
	anomalies := RefreshAnomalySummaries(mappedFactLedger, cfg.AnomalyThresholds)
	for i := range anomalies {
		if anomalies[i].Context.Role != nil {
			role := mappings.ResolveRole(string(*anomalies[i].Context.Role))
			anomalies[i].Context.Role = &role
		}
		if anomalies[i].Context.SecondaryRole != nil {
			role := mappings.ResolveRole(string(*anomalies[i].Context.SecondaryRole))
			anomalies[i].Context.SecondaryRole = &role
		}
	}

	return AgenticMetricsLedger{
		TicketID:           factLedger.TicketID,
		TicketTitle:        factLedger.TicketTitle,
		TicketType:         factLedger.TicketType,
		StartedAt:          factLedger.StartedAt,
		CompletedAt:        factLedger.CompletedAt,
		UpdatedAt:          factLedger.UpdatedAt,
		Facts:              factLedger,
		WorkflowEfficiency: eff,
		Threads:            threads,
		Totals:             totals,
		Roles:              roles,
		Models:             models,
		Anomalies:          anomalies,
	}
}

// CalculateFleetSummary aggregates high-level metrics across multiple ledgers.
func CalculateFleetSummary(ledgers []AgenticMetricsLedger) FleetSummary {
	totalTickets := int64(len(ledgers))
	if totalTickets == 0 {
		return FleetSummary{
			TotalTickets:             0,
			TotalCostUSD:             0.0,
			TotalInvocations:         0,
			TotalInputTokens:         0,
			TotalOutputTokens:        0,
			TotalCacheReadTokens:     0,
			OverallCacheHitRatePct:   0.0,
			TotalDurationSeconds:     0,
			FormattedTotalDuration:   "0m 0s",
			TotalLinesAdded:          0,
			TotalLinesDeleted:        0,
			TotalFilesChanged:        0,
			TotalReasoningTokens:     0,
			OverallReasoningRatioPct: 0.0,
			RoleCosts:                make(map[AgentRole]float64),
			ModelCosts:               make(map[string]float64),
			AnomalySummary:           ComputeFleetAnomalies(nil),
		}
	}

	var totalCostUSD float64
	var totalInvocations, totalInputTokens, totalOutputTokens, totalCacheReadTokens int64
	var totalDurationSeconds, totalLinesAdded, totalLinesDeleted, totalFilesChanged, totalReasoningTokens int64

	roleCosts := make(map[AgentRole]float64)
	modelCosts := make(map[string]float64)

	factLedgers := make([]AgenticFactLedger, len(ledgers))

	for i, ledger := range ledgers {
		factLedgers[i] = ledger.Facts
		totalCostUSD += ledger.Totals.CostUSD
		totalInvocations += ledger.Totals.Invocations
		totalInputTokens += ledger.Totals.InputTokens
		totalOutputTokens += ledger.Totals.OutputTokens
		totalCacheReadTokens += ledger.Totals.CacheReadTokens
		totalDurationSeconds += ledger.WorkflowEfficiency.ActiveDurationSeconds
		totalLinesAdded += ledger.WorkflowEfficiency.CodeDelivered.LinesAdded
		totalLinesDeleted += ledger.WorkflowEfficiency.CodeDelivered.LinesDeleted
		totalFilesChanged += ledger.WorkflowEfficiency.CodeDelivered.FilesChanged
		totalReasoningTokens += ledger.WorkflowEfficiency.ReasoningTokens

		for _, r := range ledger.Roles {
			roleCosts[r.Role] = math.Round((roleCosts[r.Role]+r.CostUSD)*10000.0) / 10000.0
		}
		for _, m := range ledger.Models {
			modelCosts[m.Model] = math.Round((modelCosts[m.Model]+m.CostUSD)*10000.0) / 10000.0
		}
	}

	anomalySummary := ComputeFleetAnomalies(factLedgers)

	totalCachedOrInput := totalInputTokens + totalCacheReadTokens
	var overallCacheHitRatePct float64
	if totalCachedOrInput > 0 {
		overallCacheHitRatePct = math.Round(float64(totalCacheReadTokens)/float64(totalCachedOrInput)*1000.0) / 10.0
	}

	durHours := totalDurationSeconds / 3600
	durRem := totalDurationSeconds % 3600
	durMins := durRem / 60
	durRemSecs := durRem % 60

	var formattedTotalDuration string
	if durHours > 0 {
		formattedTotalDuration = fmt.Sprintf("%dh %dm", durHours, durMins)
	} else {
		formattedTotalDuration = fmt.Sprintf("%dm %ds", durMins, durRemSecs)
	}

	genTotal := totalOutputTokens + totalReasoningTokens
	var overallReasoningRatioPct float64
	if genTotal > 0 {
		overallReasoningRatioPct = math.Round(float64(totalReasoningTokens)/float64(genTotal)*1000.0) / 10.0
	}

	return FleetSummary{
		TotalTickets:             totalTickets,
		TotalCostUSD:             math.Round(totalCostUSD*10000.0) / 10000.0,
		TotalInvocations:         totalInvocations,
		TotalInputTokens:         totalInputTokens,
		TotalOutputTokens:        totalOutputTokens,
		TotalCacheReadTokens:     totalCacheReadTokens,
		OverallCacheHitRatePct:   overallCacheHitRatePct,
		TotalDurationSeconds:     totalDurationSeconds,
		FormattedTotalDuration:   formattedTotalDuration,
		TotalLinesAdded:          totalLinesAdded,
		TotalLinesDeleted:        totalLinesDeleted,
		TotalFilesChanged:        totalFilesChanged,
		TotalReasoningTokens:     totalReasoningTokens,
		OverallReasoningRatioPct: overallReasoningRatioPct,
		RoleCosts:                roleCosts,
		ModelCosts:               modelCosts,
		AnomalySummary:           anomalySummary,
	}
}

// ComputeLeadTimeSeconds computes turnaround elapsed duration in seconds.
func ComputeLeadTimeSeconds(startedAt, completedAt *string) *int64 {
	if startedAt == nil || completedAt == nil {
		return nil
	}
	startDT, err := time.Parse(time.RFC3339Nano, *startedAt)
	if err != nil {
		startDT, err = time.Parse(time.RFC3339, *startedAt)
		if err != nil {
			return nil
		}
	}
	completedDT, err := time.Parse(time.RFC3339Nano, *completedAt)
	if err != nil {
		completedDT, err = time.Parse(time.RFC3339, *completedAt)
		if err != nil {
			return nil
		}
	}
	diff := int64(completedDT.Sub(startDT).Seconds())
	return &diff
}

// FormatLeadTime formats seconds into human-readable string.
func FormatLeadTime(seconds *int64) string {
	if seconds == nil {
		return "In Progress"
	}
	sec := *seconds
	if sec < 0 {
		return "0s"
	}
	if sec >= 86400 {
		days := sec / 86400
		hours := (sec % 86400) / 3600
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if sec >= 3600 {
		hours := sec / 3600
		mins := (sec % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	mins := sec / 60
	remSecs := sec % 60
	return fmt.Sprintf("%dm %ds", mins, remSecs)
}

// BuildComparisonRows constructs comparison rows sorted chronologically.
func BuildComparisonRows(ledgers []AgenticMetricsLedger) []TicketComparisonRow {
	var rows []TicketComparisonRow
	for _, ledger := range ledgers {
		tid := "unknown"
		if ledger.TicketID != nil && *ledger.TicketID != "" {
			tid = *ledger.TicketID
		}
		eff := ledger.WorkflowEfficiency
		tot := ledger.Totals
		cwh := eff.ContextWindowHealth
		leadSecs := ComputeLeadTimeSeconds(ledger.StartedAt, ledger.CompletedAt)
		title := "N/A"
		if ledger.TicketTitle != nil && *ledger.TicketTitle != "" {
			title = *ledger.TicketTitle
		}

		rows = append(rows, TicketComparisonRow{
			TicketID:                tid,
			UpdatedAt:               ledger.UpdatedAt,
			SDLCIterations:          eff.SDLCIterations,
			ActiveDurationSeconds:   eff.ActiveDurationSeconds,
			ActiveDurationFormatted: eff.ActiveDurationFormatted,
			CostUSD:                 tot.CostUSD,
			CacheHitRatePct:         tot.CacheHitRatePct,
			TestToCodeRatio:         eff.TestToProductionRatio.Ratio,
			LinesAdded:              eff.CodeDelivered.LinesAdded,
			LinesDeleted:            eff.CodeDelivered.LinesDeleted,
			FilesChanged:            eff.CodeDelivered.FilesChanged,
			ToolSuccessRatePct:      eff.ToolExecution.SuccessRatePct,
			ContextHealthStatus:     cwh.Status,
			RootSaturationPct:       cwh.RootSaturationPct,
			PeakContextTokens:       cwh.RootPeakTokens,
			TicketTitle:             title,
			TicketType:              ledger.TicketType,
			StartedAt:               ledger.StartedAt,
			CompletedAt:             ledger.CompletedAt,
			LeadDurationSeconds:     leadSecs,
			LeadDurationFormatted:   FormatLeadTime(leadSecs),
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		rI := rows[i]
		rJ := rows[j]

		// Sort key: (completed_at != nil, completed_at, started_at, ticket_id) reverse
		compIHasVal := rI.CompletedAt != nil
		compJHasVal := rJ.CompletedAt != nil
		if compIHasVal != compJHasVal {
			return compIHasVal // true comes before false in reverse sort
		}

		compIStr := ""
		if rI.CompletedAt != nil {
			compIStr = *rI.CompletedAt
		}
		compJStr := ""
		if rJ.CompletedAt != nil {
			compJStr = *rJ.CompletedAt
		}
		if compIStr != compJStr {
			return compIStr > compJStr
		}

		startIStr := ""
		if rI.StartedAt != nil {
			startIStr = *rI.StartedAt
		}
		startJStr := ""
		if rJ.StartedAt != nil {
			startJStr = *rJ.StartedAt
		}
		if startIStr != startJStr {
			return startIStr > startJStr
		}

		cmp := CompareTicketIDs(rI.TicketID, rJ.TicketID)
		if cmp != 0 {
			return cmp > 0
		}
		return rI.TicketID > rJ.TicketID
	})

	return rows
}

// CalculateRoleEffortRollup groups invocations, tokens, and cost across all tickets by agent role.
func CalculateRoleEffortRollup(ledgers []AgenticMetricsLedger, ceiling ...int64) map[string]GroupedMetric {
	effCeiling := ContextCeilingCriticalTokens
	if len(ceiling) > 0 && ceiling[0] > 0 {
		effCeiling = ceiling[0]
	}

	rollup := make(map[string]GroupedMetric)
	for _, ledger := range ledgers {
		for _, r := range ledger.Roles {
			curr := rollup[string(r.Role)]
			curr.Invocations += r.Invocations
			curr.InputTokens += r.InputTokens
			curr.OutputTokens += r.OutputTokens
			curr.CacheReadTokens += r.CacheReadTokens
			curr.CostUSD = math.Round((curr.CostUSD+r.CostUSD)*10000.0) / 10000.0
			if r.PeakContextTokens > curr.PeakContextTokens {
				curr.PeakContextTokens = r.PeakContextTokens
			}
			rollup[string(r.Role)] = curr
		}
	}

	for role, m := range rollup {
		totTok := m.InputTokens + m.CacheReadTokens
		if totTok > 0 {
			m.CacheHitRatePct = math.Round(float64(m.CacheReadTokens)/float64(totTok)*1000.0) / 10.0
		}
		m.ContextHealth = GetBaseHealthStatus(m.PeakContextTokens, effCeiling)
		rollup[role] = m
	}

	return rollup
}

func parseTimestamp(ts *string) (time.Time, bool) {
	if ts == nil {
		return time.Time{}, false
	}
	clean := strings.TrimSpace(*ts)
	if clean == "" {
		return time.Time{}, false
	}
	// Try RFC3339Nano
	if t, err := time.Parse(time.RFC3339Nano, clean); err == nil {
		return t.UTC(), true
	}
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339, clean); err == nil {
		return t.UTC(), true
	}
	// Try ISO 8601 without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", clean); err == nil {
		return t.UTC(), true
	}
	// Try HTML datetime-local format without seconds
	if t, err := time.Parse("2006-01-02T15:04", clean); err == nil {
		return t.UTC(), true
	}
	// Try date-only YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", clean); err == nil {
		return t.UTC(), true
	}
	return time.Time{}, false
}

// GetLedgerCompletionDate resolves the effective completion date for a ledger.
// It strictly parses completed_at; if nil, empty, or unparseable, it returns a zero time.Time and false.
// It never falls back to updated_at or started_at.
func GetLedgerCompletionDate(ledger AgenticMetricsLedger) (time.Time, bool) {
	if t, ok := parseTimestamp(ledger.CompletedAt); ok {
		return t, true
	}
	return time.Time{}, false
}

// CanonicalTicketType normalizes a raw ticket type string to its canonical form,
// resolving aliases such as "feat" -> "feature", "bug"/"fix"/"bugfix" -> "bug", and "doc"/"docs" -> "docs".
func CanonicalTicketType(raw string) string {
	clean := strings.ToLower(strings.TrimSpace(raw))
	switch clean {
	case "feat", "feature":
		return "feature"
	case "bug", "fix", "bugfix":
		return "bug"
	case "doc", "docs":
		return "docs"
	default:
		return clean
	}
}

// FilterLedgersByTicketTypes filters ledgers whose canonical ticket type matches any of the supplied types.
// If ticketTypes is empty, or contains "all" (case-insensitive) or an empty token, all ledgers are retained.
// Returns a newly allocated slice.
func FilterLedgersByTicketTypes(ledgers []AgenticMetricsLedger, ticketTypes []string) []AgenticMetricsLedger {
	if len(ledgers) == 0 {
		return []AgenticMetricsLedger{}
	}
	if len(ticketTypes) == 0 {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	for _, raw := range ticketTypes {
		clean := strings.ToLower(strings.TrimSpace(raw))
		if clean == "" || clean == "all" {
			result := make([]AgenticMetricsLedger, len(ledgers))
			copy(result, ledgers)
			return result
		}
	}

	targetSet := make(map[string]struct{})
	for _, raw := range ticketTypes {
		canon := CanonicalTicketType(raw)
		if canon != "" {
			targetSet[canon] = struct{}{}
		}
	}

	if len(targetSet) == 0 {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	filtered := make([]AgenticMetricsLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		if ledger.TicketType != nil {
			canon := CanonicalTicketType(string(*ledger.TicketType))
			if _, ok := targetSet[canon]; ok {
				filtered = append(filtered, ledger)
			}
		}
	}

	return filtered
}

// FilterLedgersByTicketType is a convenience wrapper delegating to FilterLedgersByTicketTypes,
// supporting comma-delimited strings.
func FilterLedgersByTicketType(ledgers []AgenticMetricsLedger, ticketType string) []AgenticMetricsLedger {
	clean := strings.TrimSpace(ticketType)
	if clean == "" || strings.EqualFold(clean, "all") {
		return FilterLedgersByTicketTypes(ledgers, nil)
	}
	parts := strings.Split(clean, ",")
	var types []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			types = append(types, trimmed)
		}
	}
	return FilterLedgersByTicketTypes(ledgers, types)
}

func isAllASCIIAlpha(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			return false
		}
	}
	return true
}

// ExtractTicketPrefix extracts the uppercase ticket prefix (e.g. "AL", "GEN", "PO").
// Evaluates conforming patterns (leading letters followed by digits) or leading alphabetic tokens before delimiters.
func ExtractTicketPrefix(ticketID string) string {
	clean := strings.TrimSpace(ticketID)
	if clean == "" {
		return ""
	}
	match := ticketIDPattern.FindStringSubmatch(clean)
	if len(match) >= 2 && match[1] != "" {
		return strings.ToUpper(match[1])
	}
	idx := strings.IndexAny(clean, "-_")
	if idx > 0 {
		candidate := clean[:idx]
		if isAllASCIIAlpha(candidate) {
			return strings.ToUpper(candidate)
		}
	}
	return ""
}

// FilterLedgersByPrefixes filters ledgers whose extracted ticket prefix matches any of the supplied prefixes.
// If prefixes is empty, or contains "all" (case-insensitive) or an empty token, all ledgers are retained.
// Returns a newly allocated slice.
func FilterLedgersByPrefixes(ledgers []AgenticMetricsLedger, prefixes []string) []AgenticMetricsLedger {
	if len(ledgers) == 0 {
		return []AgenticMetricsLedger{}
	}
	if len(prefixes) == 0 {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	for _, raw := range prefixes {
		clean := strings.ToLower(strings.TrimSpace(raw))
		if clean == "" || clean == "all" {
			result := make([]AgenticMetricsLedger, len(ledgers))
			copy(result, ledgers)
			return result
		}
	}

	targetSet := make(map[string]struct{})
	for _, raw := range prefixes {
		clean := strings.ToUpper(strings.TrimSpace(raw))
		if clean != "" {
			targetSet[clean] = struct{}{}
		}
	}

	if len(targetSet) == 0 {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	filtered := make([]AgenticMetricsLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		if ledger.TicketID != nil {
			prefix := ExtractTicketPrefix(*ledger.TicketID)
			if prefix != "" {
				if _, ok := targetSet[prefix]; ok {
					filtered = append(filtered, ledger)
				}
			}
		}
	}

	return filtered
}

// FilterLedgersByPrefix is a convenience wrapper delegating to FilterLedgersByPrefixes,
// supporting comma-delimited strings.
func FilterLedgersByPrefix(ledgers []AgenticMetricsLedger, prefix string) []AgenticMetricsLedger {
	clean := strings.TrimSpace(prefix)
	if clean == "" || strings.EqualFold(clean, "all") {
		return FilterLedgersByPrefixes(ledgers, nil)
	}
	parts := strings.Split(clean, ",")
	var prefixes []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			prefixes = append(prefixes, trimmed)
		}
	}
	return FilterLedgersByPrefixes(ledgers, prefixes)
}

// ExtractDistinctPrefixes returns an array of unique non-empty uppercase ticket prefix strings sorted in ascending alphabetical order.
func ExtractDistinctPrefixes(ledgers []AgenticMetricsLedger) []string {
	if len(ledgers) == 0 {
		return []string{}
	}
	prefixMap := make(map[string]struct{})
	for _, ledger := range ledgers {
		if ledger.TicketID != nil {
			p := ExtractTicketPrefix(*ledger.TicketID)
			if p != "" {
				prefixMap[p] = struct{}{}
			}
		}
	}
	if len(prefixMap) == 0 {
		return []string{}
	}
	results := make([]string, 0, len(prefixMap))
	for p := range prefixMap {
		results = append(results, p)
	}
	sort.Strings(results)
	return results
}

// FilterLedgersByCompletionDate filters ledgers based on their effective completion date within inclusive bounds.
// Ledgers lacking a valid date are excluded when boundaries are active.
// Returns a newly allocated slice.
func FilterLedgersByCompletionDate(ledgers []AgenticMetricsLedger, since, until *time.Time) []AgenticMetricsLedger {
	if len(ledgers) == 0 {
		return []AgenticMetricsLedger{}
	}
	if since == nil && until == nil {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	filtered := make([]AgenticMetricsLedger, 0, len(ledgers))
	for _, ledger := range ledgers {
		t, ok := GetLedgerCompletionDate(ledger)
		if !ok {
			continue
		}
		if since != nil && t.Before(*since) {
			continue
		}
		if until != nil && t.After(*until) {
			continue
		}
		filtered = append(filtered, ledger)
	}

	return filtered
}

func parseISODate(ts *string) (time.Time, bool) {
	if ts == nil || len(*ts) < 10 {
		return time.Time{}, false
	}
	clean := (*ts)[:10]
	t, err := time.Parse("2006-01-02", clean)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// GetTicketEffectiveDate resolves the effective date for a ticket strictly based on completed_at.
// If completed_at is nil, empty, or unparseable, it returns a zero time.Time and false.
// It never falls back to started_at or updated_at.
func GetTicketEffectiveDate(completedAt, startedAt *string, updatedAt string) (time.Time, bool) {
	return parseISODate(completedAt)
}

// ResolveDatePreset resolves preset string into start and end dates.
func ResolveDatePreset(preset string, refDate time.Time) (startDate *time.Time, endDate *time.Time) {
	ref := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)
	switch TimeWindowPreset(preset) {
	case TimeWindowLast24Hours:
		start := refDate.Add(-24 * time.Hour)
		end := refDate
		return &start, &end
	case TimeWindowLast7Days:
		start := ref.AddDate(0, 0, -7)
		return &start, &ref
	case TimeWindowLast30Days:
		start := ref.AddDate(0, 0, -30)
		return &start, &ref
	case TimeWindowLast90Days:
		start := ref.AddDate(0, 0, -90)
		return &start, &ref
	default:
		return nil, nil
	}
}

// FilterLedgersByDateRange filters ledgers based on their effective date within inclusive bounds.
func FilterLedgersByDateRange(ledgers []AgenticMetricsLedger, startDate, endDate *time.Time) []AgenticMetricsLedger {
	if startDate == nil && endDate == nil {
		return ledgers
	}

	var filtered []AgenticMetricsLedger
	for _, ledger := range ledgers {
		t, ok := GetTicketEffectiveDate(ledger.CompletedAt, ledger.StartedAt, ledger.UpdatedAt)
		if !ok {
			continue
		}
		if startDate != nil && t.Before(*startDate) {
			continue
		}
		if endDate != nil && t.After(*endDate) {
			continue
		}
		filtered = append(filtered, ledger)
	}
	return filtered
}

// FilterLedgersByTimeWindow filters ledgers based on a TimeWindow relative to refDate.
// For TimeWindowAll or empty windows, it returns a clone of all ledgers.
func FilterLedgersByTimeWindow(ledgers []AgenticMetricsLedger, window TimeWindow, refDate time.Time) []AgenticMetricsLedger {
	if window == TimeWindowAll || len(ledgers) == 0 {
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	if window == TimeWindow1D {
		start := refDate.Add(-24 * time.Hour)
		var filtered []AgenticMetricsLedger
		for _, ledger := range ledgers {
			t, ok := GetLedgerCompletionDate(ledger)
			if !ok {
				continue
			}
			if !t.Before(start) && !t.After(refDate) {
				filtered = append(filtered, ledger)
			}
		}
		if filtered == nil {
			return []AgenticMetricsLedger{}
		}
		return filtered
	}

	var days int
	switch window {
	case TimeWindow7D:
		days = 7
	case TimeWindow30D:
		days = 30
	case TimeWindow90D:
		days = 90
	default:
		result := make([]AgenticMetricsLedger, len(ledgers))
		copy(result, ledgers)
		return result
	}

	ref := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 23, 59, 59, 999999999, time.UTC)
	start := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days)

	var filtered []AgenticMetricsLedger
	for _, ledger := range ledgers {
		t, ok := GetTicketEffectiveDate(ledger.CompletedAt, ledger.StartedAt, ledger.UpdatedAt)
		if !ok {
			continue
		}
		if !t.Before(start) && !t.After(ref) {
			filtered = append(filtered, ledger)
		}
	}

	if filtered == nil {
		return []AgenticMetricsLedger{}
	}
	return filtered
}

// SortLedgers deterministically sorts a slice of AgenticMetricsLedger according to field and order.
// It returns a newly allocated sorted slice without mutating the original input slice.
func SortLedgers(ledgers []AgenticMetricsLedger, field LedgerSortField, ascending bool) []AgenticMetricsLedger {
	result := make([]AgenticMetricsLedger, len(ledgers))
	copy(result, ledgers)

	if len(result) <= 1 {
		return result
	}

	sort.SliceStable(result, func(i, j int) bool {
		a := result[i]
		b := result[j]

		if field == SortFieldCompletedAt {
			dateA, okA := GetLedgerCompletionDate(a)
			dateB, okB := GetLedgerCompletionDate(b)
			if okA != okB {
				// Nulls last invariant regardless of sort direction:
				// a candidate with a date precedes a candidate without a date.
				return okA
			}
			if okA && okB {
				if !dateA.Equal(dateB) {
					if ascending {
						return dateA.Before(dateB)
					}
					return dateA.After(dateB)
				}
			}
			tidA := ""
			if a.TicketID != nil {
				tidA = *a.TicketID
			}
			tidB := ""
			if b.TicketID != nil {
				tidB = *b.TicketID
			}
			cmp := CompareTicketIDs(tidA, tidB)
			if cmp != 0 {
				return cmp < 0
			}
			return a.UpdatedAt < b.UpdatedAt
		}

		var less bool
		switch field {
		case SortFieldTicketID:
			tidA := ""
			if a.TicketID != nil {
				tidA = *a.TicketID
			}
			tidB := ""
			if b.TicketID != nil {
				tidB = *b.TicketID
			}
			cmp := CompareTicketIDs(tidA, tidB)
			if cmp == 0 {
				less = a.UpdatedAt < b.UpdatedAt
			} else {
				less = cmp < 0
			}
		case SortFieldCost:
			if a.Totals.CostUSD == b.Totals.CostUSD {
				less = a.UpdatedAt < b.UpdatedAt
			} else {
				less = a.Totals.CostUSD < b.Totals.CostUSD
			}
		case SortFieldDuration:
			if a.WorkflowEfficiency.ActiveDurationSeconds == b.WorkflowEfficiency.ActiveDurationSeconds {
				less = a.UpdatedAt < b.UpdatedAt
			} else {
				less = a.WorkflowEfficiency.ActiveDurationSeconds < b.WorkflowEfficiency.ActiveDurationSeconds
			}
		case SortFieldInvocations:
			if a.Totals.Invocations == b.Totals.Invocations {
				less = a.UpdatedAt < b.UpdatedAt
			} else {
				less = a.Totals.Invocations < b.Totals.Invocations
			}
		case SortFieldUpdatedAt:
			fallthrough
		default:
			less = a.UpdatedAt < b.UpdatedAt
		}

		if ascending {
			return less
		}
		return !less
	})

	return result
}

type bucketAccumulator struct {
	invocations           int64
	costUSD               float64
	inputTokens           int64
	outputTokens          int64
	cacheReadTokens       int64
	reasoningTokens       int64
	activeDurationSeconds int64
	peakContextTokens     int64
	linesAdded            int64
	linesDeleted          int64
	filesChanged          int64
}

// AggregateMetricsOverTime aggregates metrics chronologically into hourly, daily, or weekly buckets.
func AggregateMetricsOverTime(
	ledgers []AgenticMetricsLedger,
	granularity string,
	startDate, endDate *time.Time,
) []TimeBucketMetric {
	isHourly := strings.EqualFold(granularity, string(GranularityHourly))
	isWeekly := strings.EqualFold(granularity, string(GranularityWeekly))

	accumulators := make(map[string]*bucketAccumulator)

	getBucketKey := func(d time.Time) string {
		utc := d.UTC()
		if isHourly {
			return utc.Format("2006-01-02T15:00:00Z")
		}
		if isWeekly {
			// Find Monday
			weekday := int(utc.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday is 7 days from last Monday
			}
			monday := utc.AddDate(0, 0, -(weekday - 1))
			return monday.Format("2006-01-02")
		}
		return utc.Format("2006-01-02")
	}

	for _, ledger := range ledgers {
		for _, inv := range ledger.Facts.InvocationFacts() {
			var invDate time.Time
			var ok bool
			if d, okInv := parseTimestamp(inv.CompletedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseTimestamp(inv.StartedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseTimestamp(ledger.CompletedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseTimestamp(ledger.StartedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseTimestamp(&ledger.UpdatedAt); okInv {
				invDate = d
				ok = true
			}
			if !ok {
				continue
			}
			if startDate != nil && invDate.Before(*startDate) {
				continue
			}
			if endDate != nil && invDate.After(*endDate) {
				continue
			}

			bkey := getBucketKey(invDate)
			acc, exists := accumulators[bkey]
			if !exists {
				acc = &bucketAccumulator{}
				accumulators[bkey] = acc
			}
			acc.invocations++
			acc.costUSD += inv.CostUSD
			acc.inputTokens += inv.Tokens.Input
			acc.outputTokens += inv.Tokens.Output
			acc.cacheReadTokens += inv.Tokens.CacheRead
			acc.reasoningTokens += inv.Tokens.Reasoning
			acc.activeDurationSeconds += inv.ActiveDurationSeconds
			if inv.Tokens.PeakContext > acc.peakContextTokens {
				acc.peakContextTokens = inv.Tokens.PeakContext
			}
		}

		var codeDate time.Time
		var codeDateOk bool
		if d, ok := parseTimestamp(ledger.CompletedAt); ok {
			codeDate = d
			codeDateOk = true
		} else if d, ok := parseTimestamp(ledger.StartedAt); ok {
			codeDate = d
			codeDateOk = true
		} else if d, ok := parseTimestamp(&ledger.UpdatedAt); ok {
			codeDate = d
			codeDateOk = true
		}

		if codeDateOk {
			inRange := true
			if startDate != nil && codeDate.Before(*startDate) {
				inRange = false
			}
			if endDate != nil && codeDate.After(*endDate) {
				inRange = false
			}

			if inRange {
				bkey := getBucketKey(codeDate)
				acc, exists := accumulators[bkey]
				if !exists {
					acc = &bucketAccumulator{}
					accumulators[bkey] = acc
				}
				code := ledger.WorkflowEfficiency.CodeDelivered
				acc.linesAdded += code.LinesAdded
				acc.linesDeleted += code.LinesDeleted
				acc.filesChanged += code.FilesChanged
			}
		}
	}

	if len(accumulators) == 0 {
		return []TimeBucketMetric{}
	}

	var sortedKeys []string
	for k := range accumulators {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var runningCumCost float64
	var runningCumInput, runningCumOutput, runningCumLinesAdded, runningCumLinesDeleted, runningCumInvocations int64

	var results []TimeBucketMetric
	for _, bkey := range sortedKeys {
		acc := accumulators[bkey]
		runningCumCost += acc.costUSD
		runningCumInput += acc.inputTokens
		runningCumOutput += acc.outputTokens
		runningCumLinesAdded += acc.linesAdded
		runningCumLinesDeleted += acc.linesDeleted
		runningCumInvocations += acc.invocations

		cachedOrInp := acc.inputTokens + acc.cacheReadTokens
		var hitPct float64
		if cachedOrInp > 0 {
			hitPct = math.Round(float64(acc.cacheReadTokens)/float64(cachedOrInp)*1000.0) / 10.0
		}

		results = append(results, TimeBucketMetric{
			BucketKey:               bkey,
			BucketDate:              bkey,
			Invocations:             acc.invocations,
			CostUSD:                 math.Round(acc.costUSD*10000.0) / 10000.0,
			InputTokens:             acc.inputTokens,
			OutputTokens:            acc.outputTokens,
			CacheReadTokens:         acc.cacheReadTokens,
			ReasoningTokens:         acc.reasoningTokens,
			CacheHitRatePct:         hitPct,
			LinesAdded:              acc.linesAdded,
			LinesDeleted:            acc.linesDeleted,
			FilesChanged:            acc.filesChanged,
			NetLinesAdded:           acc.linesAdded - acc.linesDeleted,
			PeakContextTokens:       acc.peakContextTokens,
			ActiveDurationSeconds:   acc.activeDurationSeconds,
			CumulativeCostUSD:       math.Round(runningCumCost*10000.0) / 10000.0,
			CumulativeInputTokens:   runningCumInput,
			CumulativeOutputTokens:  runningCumOutput,
			CumulativeLinesAdded:    runningCumLinesAdded,
			CumulativeLinesDeleted:  runningCumLinesDeleted,
			CumulativeNetLinesAdded: runningCumLinesAdded - runningCumLinesDeleted,
			CumulativeInvocations:   runningCumInvocations,
		})
	}

	return results
}
