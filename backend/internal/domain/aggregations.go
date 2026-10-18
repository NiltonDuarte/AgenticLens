package domain

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

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
	sids := make(map[string]struct{})
	var inp, out, cache, peak int64
	var cost float64

	for _, inv := range invs {
		sids[inv.ID] = struct{}{}
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
		invocations:       int64(len(sids)),
		inputTokens:       inp,
		outputTokens:      out,
		cacheReadTokens:   cache,
		cacheHitRatePct:   hitPct,
		costUSD:           math.Round(cost*10000.0) / 10000.0,
		peakContextTokens: peak,
	}
}

// ComputeTotals calculates aggregate invocation metrics from a slice of invocation facts.
func ComputeTotals(invocations []InvocationFact) GroupedMetric {
	if len(invocations) == 0 {
		return GroupedMetric{
			Invocations:       0,
			InputTokens:       0,
			OutputTokens:      0,
			CacheReadTokens:   0,
			CacheHitRatePct:   0.0,
			CostUSD:           0.0,
			PeakContextTokens: 0,
			ContextHealth:     string(ContextHealthHealthy),
		}
	}

	byRole := make(map[string][]InvocationFact)
	for _, inv := range invocations {
		byRole[inv.Role] = append(byRole[inv.Role], inv)
	}

	var totalInvocations int64
	for _, rInvs := range byRole {
		totalInvocations += accumulateInvocations(rInvs).invocations
	}

	acc := accumulateInvocations(invocations)
	healthStatus := GetBaseHealthStatus(acc.peakContextTokens, HealthyCeilingTokens)

	return GroupedMetric{
		Invocations:       totalInvocations,
		InputTokens:       acc.inputTokens,
		OutputTokens:      acc.outputTokens,
		CacheReadTokens:   acc.cacheReadTokens,
		CacheHitRatePct:   acc.cacheHitRatePct,
		CostUSD:           acc.costUSD,
		PeakContextTokens: acc.peakContextTokens,
		ContextHealth:     string(healthStatus),
	}
}

// ComputeWorkflowEfficiency derives workflow efficiency profile from an atomic fact ledger.
func ComputeWorkflowEfficiency(ledger AgenticFactLedger) WorkflowEfficiency {
	invs := ledger.Invocations
	tools := ledger.Tools
	code := ledger.Code

	var activeDurationSeconds int64
	roleCosts := make(map[string]float64)
	rolePeaks := make(map[string]int64)
	roleTurns := make(map[string][]int64)
	var totReasoning, totOutput, totCompactions int64

	for _, inv := range invs {
		activeDurationSeconds += inv.ActiveDurationSeconds
		roleCosts[inv.Role] += inv.CostUSD
		if inv.Tokens.PeakContext > rolePeaks[inv.Role] {
			rolePeaks[inv.Role] = inv.Tokens.PeakContext
		}
		if inv.ParentID != nil {
			roleTurns[inv.Role] = append(roleTurns[inv.Role], inv.Turns)
		}
		totReasoning += inv.Tokens.Reasoning
		totOutput += inv.Tokens.Output
		totCompactions += inv.Compactions
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

	subCosts := make(map[string]float64)
	var totSubCost float64
	for _, r := range SubagentRoleOrder {
		if c, ok := roleCosts[r]; ok {
			subCosts[r] = c
			totSubCost += c
		}
	}

	roleEffortDistributionPct := make(map[string]float64)
	for _, r := range SubagentRoleOrder {
		if c, ok := subCosts[r]; ok && totSubCost > 0 {
			roleEffortDistributionPct[r] = math.Round(c/totSubCost*1000.0) / 10.0
		}
	}

	var allSubTurns []int64
	for _, r := range SubagentRoleOrder {
		if turns, ok := roleTurns[r]; ok {
			allSubTurns = append(allSubTurns, turns...)
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

	velByRole := make(map[string]float64)
	for _, r := range SubagentRoleOrder {
		if turns, ok := roleTurns[r]; ok && len(turns) > 0 {
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

	orchPeak := rolePeaks[string(RoleOrchestrator)]
	var orchSatPct float64
	if HealthyCeilingTokens > 0 {
		orchSatPct = math.Round(float64(orchPeak)/float64(HealthyCeilingTokens)*1000.0) / 10.0
	}

	var subPeak int64
	subRoleStr := "none"
	var foundSubRole bool
	for _, r := range SubagentRoleOrder {
		if peak, ok := rolePeaks[r]; ok {
			if !foundSubRole || peak > subPeak {
				subPeak = peak
				subRoleStr = r
				foundSubRole = true
			}
		}
	}

	overallPeak := orchPeak
	if subPeak > overallPeak {
		overallPeak = subPeak
	}
	cwhStatus := GetBaseHealthStatus(overallPeak, HealthyCeilingTokens)

	contextWindowHealth := ContextWindowHealth{
		HealthyCeilingTokens:      HealthyCeilingTokens,
		Status:                    string(cwhStatus),
		OrchestratorPeakTokens:    orchPeak,
		OrchestratorSaturationPct: orchSatPct,
		SubagentsPeakTokens:       subPeak,
		SubagentsPeakRole:         subRoleStr,
		Compactions:               totCompactions,
		PeakByRole:                rolePeaks,
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

// ComputeRoleMetrics calculates aggregated metrics grouped by agent role in canonical order.
func ComputeRoleMetrics(invocations []InvocationFact) []RoleMetric {
	byRole := make(map[string][]InvocationFact)
	for _, inv := range invocations {
		byRole[inv.Role] = append(byRole[inv.Role], inv)
	}

	canonicalSet := make(map[string]struct{})
	for _, r := range CanonicalRoleOrder {
		canonicalSet[r] = struct{}{}
	}

	var orderedRoles []string
	for _, r := range CanonicalRoleOrder {
		if _, ok := byRole[r]; ok {
			orderedRoles = append(orderedRoles, r)
		}
	}

	var customRoles []string
	for r := range byRole {
		if _, ok := canonicalSet[r]; !ok {
			customRoles = append(customRoles, r)
		}
	}
	sort.Strings(customRoles)
	orderedRoles = append(orderedRoles, customRoles...)

	var metrics []RoleMetric
	for _, role := range orderedRoles {
		acc := accumulateInvocations(byRole[role])
		health := GetHealthLabel(acc.peakContextTokens, HealthyCeilingTokens)

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

// ComputeModelMetrics calculates aggregated metrics grouped by model, sorted by cost descending.
func ComputeModelMetrics(invocations []InvocationFact) []ModelMetric {
	byModel := make(map[string][]InvocationFact)
	for _, inv := range invocations {
		byModel[inv.Model] = append(byModel[inv.Model], inv)
	}

	var metrics []ModelMetric
	for model, mInvs := range byModel {
		acc := accumulateInvocations(mInvs)
		health := GetHealthLabel(acc.peakContextTokens, HealthyCeilingTokens)

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
		return metrics[i].CostUSD > metrics[j].CostUSD
	})

	return metrics
}

// ComputeThreadSummaries groups invocations by root session into hierarchical thread execution records.
func ComputeThreadSummaries(invocations []InvocationFact) []ThreadExecutionRecord {
	if len(invocations) == 0 {
		return []ThreadExecutionRecord{}
	}

	idToParent := make(map[string]*string)
	idToUser := make(map[string]string)
	for _, inv := range invocations {
		idToParent[inv.ID] = inv.ParentID
		idToUser[inv.ID] = inv.User
	}

	resolveRoot := func(inv InvocationFact) string {
		curr := inv.ID
		visited := make(map[string]struct{})
		for {
			parentPtr, ok := idToParent[curr]
			if !ok || parentPtr == nil {
				break
			}
			if _, seen := visited[curr]; seen {
				break
			}
			visited[curr] = struct{}{}
			curr = *parentPtr
		}
		if _, seen := visited[curr]; !seen {
			if _, ok := idToParent[curr]; !ok && inv.ParentID != nil {
				return *inv.ParentID
			}
		}
		return curr
	}

	threadGroups := make(map[string][]InvocationFact)
	for _, inv := range invocations {
		rootID := resolveRoot(inv)
		threadGroups[rootID] = append(threadGroups[rootID], inv)
	}

	canonicalSet := make(map[string]struct{})
	for _, r := range CanonicalRoleOrder {
		canonicalSet[r] = struct{}{}
	}

	var records []ThreadExecutionRecord
	for rootID, threadInvs := range threadGroups {
		threadUser, ok := idToUser[rootID]
		if !ok || threadUser == "" {
			if len(threadInvs) > 0 && threadInvs[0].User != "" {
				threadUser = threadInvs[0].User
			} else {
				threadUser = "unknown"
			}
		}

		var compactions int64
		byRole := make(map[string][]InvocationFact)
		for _, inv := range threadInvs {
			compactions += inv.Compactions
			byRole[inv.Role] = append(byRole[inv.Role], inv)
		}

		var orderedRoles []string
		for _, r := range CanonicalRoleOrder {
			if _, exists := byRole[r]; exists {
				orderedRoles = append(orderedRoles, r)
			}
		}

		var customRoles []string
		for r := range byRole {
			if _, exists := canonicalSet[r]; !exists {
				customRoles = append(customRoles, r)
			}
		}
		sort.Strings(customRoles)
		orderedRoles = append(orderedRoles, customRoles...)

		var roleMetrics []ThreadRoleMetric
		for _, role := range orderedRoles {
			rInvs := byRole[role]
			byModel := make(map[string][]InvocationFact)
			for _, inv := range rInvs {
				byModel[inv.Model] = append(byModel[inv.Model], inv)
			}

			var modelMetrics []ModelMetric
			for model, mInvs := range byModel {
				sids := make(map[string]struct{})
				var inp, out, cache int64
				var cost float64
				var peak int64

				for _, inv := range mInvs {
					sids[inv.ID] = struct{}{}
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
				health := GetHealthLabel(peak, HealthyCeilingTokens)

				modelMetrics = append(modelMetrics, ModelMetric{
					Model:             model,
					Invocations:       int64(len(sids)),
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
				return modelMetrics[i].CostUSD > modelMetrics[j].CostUSD
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

// DeriveMetricsLedger projects all aggregations from an atomic fact ledger.
func DeriveMetricsLedger(factLedger AgenticFactLedger) AgenticMetricsLedger {
	totals := ComputeTotals(factLedger.Invocations)
	eff := ComputeWorkflowEfficiency(factLedger)
	roles := ComputeRoleMetrics(factLedger.Invocations)
	models := ComputeModelMetrics(factLedger.Invocations)
	threads := ComputeThreadSummaries(factLedger.Invocations)
	anomalies := DetectAnomalies(factLedger)

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
			RoleCosts:                make(map[string]float64),
			ModelCosts:               make(map[string]float64),
			AnomalySummary:           ComputeFleetAnomalies(nil),
		}
	}

	var totalCostUSD float64
	var totalInvocations, totalInputTokens, totalOutputTokens, totalCacheReadTokens int64
	var totalDurationSeconds, totalLinesAdded, totalLinesDeleted, totalFilesChanged, totalReasoningTokens int64

	roleCosts := make(map[string]float64)
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
			TicketID:                  tid,
			UpdatedAt:                 ledger.UpdatedAt,
			SDLCIterations:            eff.SDLCIterations,
			ActiveDurationSeconds:     eff.ActiveDurationSeconds,
			ActiveDurationFormatted:   eff.ActiveDurationFormatted,
			CostUSD:                   tot.CostUSD,
			CacheHitRatePct:           tot.CacheHitRatePct,
			TestToCodeRatio:           eff.TestToProductionRatio.Ratio,
			LinesAdded:                eff.CodeDelivered.LinesAdded,
			LinesDeleted:              eff.CodeDelivered.LinesDeleted,
			FilesChanged:              eff.CodeDelivered.FilesChanged,
			ToolSuccessRatePct:        eff.ToolExecution.SuccessRatePct,
			ContextHealthStatus:       cwh.Status,
			OrchestratorSaturationPct: cwh.OrchestratorSaturationPct,
			PeakContextTokens:         cwh.OrchestratorPeakTokens,
			TicketTitle:               title,
			StartedAt:                 ledger.StartedAt,
			CompletedAt:               ledger.CompletedAt,
			LeadDurationSeconds:       leadSecs,
			LeadDurationFormatted:     FormatLeadTime(leadSecs),
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

		return rI.TicketID > rJ.TicketID
	})

	return rows
}

// CalculateRoleEffortRollup groups invocations, tokens, and cost across all tickets by agent role.
func CalculateRoleEffortRollup(ledgers []AgenticMetricsLedger) map[string]GroupedMetric {
	rollup := make(map[string]GroupedMetric)
	for _, ledger := range ledgers {
		for _, r := range ledger.Roles {
			curr := rollup[r.Role]
			curr.Invocations += r.Invocations
			curr.InputTokens += r.InputTokens
			curr.OutputTokens += r.OutputTokens
			curr.CacheReadTokens += r.CacheReadTokens
			curr.CostUSD = math.Round((curr.CostUSD+r.CostUSD)*10000.0) / 10000.0
			if r.PeakContextTokens > curr.PeakContextTokens {
				curr.PeakContextTokens = r.PeakContextTokens
			}
			rollup[r.Role] = curr
		}
	}

	for role, m := range rollup {
		totTok := m.InputTokens + m.CacheReadTokens
		if totTok > 0 {
			m.CacheHitRatePct = math.Round(float64(m.CacheReadTokens)/float64(totTok)*1000.0) / 10.0
		}
		m.ContextHealth = GetHealthLabel(m.PeakContextTokens, HealthyCeilingTokens)
		rollup[role] = m
	}

	return rollup
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

// GetTicketEffectiveDate resolves the effective date for a ticket (completed_at -> started_at -> updated_at).
func GetTicketEffectiveDate(completedAt, startedAt *string, updatedAt string) (time.Time, bool) {
	if t, ok := parseISODate(completedAt); ok {
		return t, true
	}
	if t, ok := parseISODate(startedAt); ok {
		return t, true
	}
	return parseISODate(&updatedAt)
}

// ResolveDatePreset resolves preset string into start and end dates.
func ResolveDatePreset(preset string, refDate time.Time) (startDate *time.Time, endDate *time.Time) {
	ref := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)
	switch TimeWindowPreset(preset) {
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

		var less bool
		switch field {
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

// AggregateMetricsOverTime aggregates metrics chronologically into daily, weekly, or cumulative buckets.
func AggregateMetricsOverTime(
	ledgers []AgenticMetricsLedger,
	granularity string,
	startDate, endDate *time.Time,
) []TimeBucketMetric {
	isWeekly := strings.ToLower(granularity) == string(GranularityWeekly) || strings.ToLower(granularity) == "weekly"

	accumulators := make(map[string]*bucketAccumulator)

	getBucketKey := func(d time.Time) string {
		if isWeekly {
			// Find Monday
			weekday := int(d.Weekday())
			if weekday == 0 {
				weekday = 7 // Sunday is 7 days from last Monday
			}
			monday := d.AddDate(0, 0, -(weekday - 1))
			return monday.Format("2006-01-02")
		}
		return d.Format("2006-01-02")
	}

	for _, ledger := range ledgers {
		for _, inv := range ledger.Facts.Invocations {
			var invDate time.Time
			var ok bool
			if d, okInv := parseISODate(inv.CompletedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseISODate(inv.StartedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseISODate(ledger.CompletedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseISODate(ledger.StartedAt); okInv {
				invDate = d
				ok = true
			} else if d, okInv := parseISODate(&ledger.UpdatedAt); okInv {
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
		if d, ok := parseISODate(ledger.CompletedAt); ok {
			codeDate = d
			codeDateOk = true
		} else if d, ok := parseISODate(ledger.StartedAt); ok {
			codeDate = d
			codeDateOk = true
		} else if d, ok := parseISODate(&ledger.UpdatedAt); ok {
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
			BucketKey:              bkey,
			BucketDate:             bkey,
			Invocations:            acc.invocations,
			CostUSD:                math.Round(acc.costUSD*10000.0) / 10000.0,
			InputTokens:            acc.inputTokens,
			OutputTokens:           acc.outputTokens,
			CacheReadTokens:        acc.cacheReadTokens,
			ReasoningTokens:        acc.reasoningTokens,
			CacheHitRatePct:        hitPct,
			LinesAdded:             acc.linesAdded,
			LinesDeleted:           acc.linesDeleted,
			FilesChanged:           acc.filesChanged,
			NetLinesAdded:          acc.linesAdded - acc.linesDeleted,
			PeakContextTokens:      acc.peakContextTokens,
			ActiveDurationSeconds:  acc.activeDurationSeconds,
			CumulativeCostUSD:      math.Round(runningCumCost*10000.0) / 10000.0,
			CumulativeInputTokens:  runningCumInput,
			CumulativeOutputTokens: runningCumOutput,
			CumulativeLinesAdded:   runningCumLinesAdded,
			CumulativeLinesDeleted: runningCumLinesDeleted,
			CumulativeInvocations:  runningCumInvocations,
		})
	}

	return results
}
