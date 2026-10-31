package adapters

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"agenticlens/internal/domain"
)

// FormatNumber formats an integer with thousands separator commas.
func FormatNumber(val int64) string {
	in := fmt.Sprintf("%d", val)
	isNeg := false
	if strings.HasPrefix(in, "-") {
		isNeg = true
		in = in[1:]
	}
	n := len(in)
	if n <= 3 {
		if isNeg {
			return "-" + in
		}
		return in
	}

	var sb strings.Builder
	if isNeg {
		sb.WriteByte('-')
	}
	rem := n % 3
	if rem > 0 {
		sb.WriteString(in[:rem])
		if n > rem {
			sb.WriteByte(',')
		}
	}
	for i := rem; i < n; i += 3 {
		sb.WriteString(in[i : i+3])
		if i+3 < n {
			sb.WriteByte(',')
		}
	}
	return sb.String()
}

// titleCase converts lowercase role strings to Title Case (e.g., "orchestrator" -> "Orchestrator").
func titleCase[T ~string](s T) string {
	str := string(s)
	if str == "" {
		return ""
	}
	return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}

// FormatMarkdownMetricsTable reconstructs the exact markdown review table from domain ledger.
func FormatMarkdownMetricsTable(ledger domain.AgenticMetricsLedger, roleOrder ...[]string) string {
	we := ledger.WorkflowEfficiency
	cwh := we.ContextWindowHealth
	totals := ledger.Totals

	contributorSet := make(map[string]struct{})
	for _, s := range ledger.Facts.Sessions {
		if s.User != "" && s.User != "unknown" {
			contributorSet[s.User] = struct{}{}
		}
	}
	var contributors []string
	for user := range contributorSet {
		contributors = append(contributors, user)
	}
	sort.Strings(contributors)

	var distRoles []string
	for role := range we.RoleEffortDistributionPct {
		distRoles = append(distRoles, role)
	}
	orderedDistRoles := domain.OrderSubagentRoles(distRoles, roleOrder...)

	var roleDistParts []string
	for _, role := range orderedDistRoles {
		pct := we.RoleEffortDistributionPct[role]
		if pct > 0 {
			roleDistParts = append(roleDistParts, fmt.Sprintf("%s %.1f%%", titleCase(role), pct))
		}
	}
	roleDistStr := "N/A"
	if we.RoleEffortUnavailableReason != "" {
		roleDistStr = "Unavailable: " + we.RoleEffortUnavailableReason
	}
	if len(roleDistParts) > 0 {
		roleDistStr = strings.Join(roleDistParts, ", ")
	}

	var velRoles []domain.AgentRole
	for role := range we.SubagentTurnVelocity.ByRole {
		velRoles = append(velRoles, role)
	}
	var customVelOrder []domain.AgentRole
	if len(roleOrder) > 0 && len(roleOrder[0]) > 0 {
		customVelOrder = make([]domain.AgentRole, len(roleOrder[0]))
		for i, r := range roleOrder[0] {
			customVelOrder[i] = domain.AgentRole(r)
		}
	}
	var orderedVelRoles []domain.AgentRole
	if len(customVelOrder) > 0 {
		orderedVelRoles = domain.OrderSubagentRoles(velRoles, customVelOrder)
	} else {
		orderedVelRoles = domain.OrderSubagentRoles(velRoles)
	}

	var velParts []string
	for _, role := range orderedVelRoles {
		turns := we.SubagentTurnVelocity.ByRole[role]
		if turns > 0 {
			velParts = append(velParts, fmt.Sprintf("%s %.1f", titleCase(role), turns))
		}
	}
	velStr := ""
	if len(velParts) > 0 {
		velStr = fmt.Sprintf(" | %s", strings.Join(velParts, ", "))
	}

	rootTokStr := domain.FormatTokensK(cwh.RootPeakTokens)
	subTokStr := domain.FormatTokensK(cwh.SubagentsPeakTokens)
	rootPctStr := fmt.Sprintf("%.0f%%", math.Round(cwh.RootSaturationPct))

	subPeakRoleStr := titleCase(cwh.SubagentsPeakRole)

	effectiveCeiling := cwh.HealthyCeilingTokens
	if effectiveCeiling <= 0 {
		effectiveCeiling = domain.ContextCeilingCriticalTokens
	}

	lines := []string{
		"### Workflow Efficiency",
		fmt.Sprintf("- **Software Development Life Cycle Iterations:** %d", we.SDLCIterations),
		fmt.Sprintf("- **Active Agent Duration:** %s (excludes user wait time)", we.ActiveDurationFormatted),
		fmt.Sprintf(
			"- **Context Window Health:** %s — Root agent peak %s tokens (%s of %s healthy ceiling, %d compactions) | Subagents peak %s (%s)",
			cwh.Status,
			rootTokStr,
			rootPctStr,
			domain.FormatTokensK(effectiveCeiling),
			cwh.Compactions,
			subTokStr,
			subPeakRoleStr,
		),
		fmt.Sprintf("- **Role Effort Distribution (Subagents):** %s", roleDistStr),
		fmt.Sprintf(
			"- **Subagent Turn Velocity & Thrashing:** %.1f avg turns/run (peak: %d turns)%s",
			we.SubagentTurnVelocity.AvgTurnsPerInvocation,
			we.SubagentTurnVelocity.PeakTurns,
			velStr,
		),
		fmt.Sprintf(
			"- **Test-to-Production Code Ratio:** %s (+%s test lines / +%s prod lines)",
			we.TestToProductionRatio.Ratio,
			FormatNumber(we.TestToProductionRatio.TestLinesAdded),
			FormatNumber(we.TestToProductionRatio.ProdLinesAdded),
		),
		fmt.Sprintf(
			"- **Code Delivered:** +%s / -%s lines across %d code files (excluding docs)",
			FormatNumber(we.CodeDelivered.LinesAdded),
			FormatNumber(we.CodeDelivered.LinesDeleted),
			we.CodeDelivered.FilesChanged,
		),
		fmt.Sprintf(
			"- **Reasoning Effort:** %s reasoning tokens (%.1f%% of generated output)",
			FormatNumber(we.ReasoningTokens),
			we.ReasoningRatioPct,
		),
		fmt.Sprintf(
			"- **Tool Health & Self-Correction:** %s tool calls (%d self-corrections, %.1f%% success rate)",
			FormatNumber(we.ToolExecution.TotalCalls),
			we.ToolExecution.ErrorsAndCorrections,
			we.ToolExecution.SuccessRatePct,
		),
		"",
		"### Anomalies & Health Flags",
	}

	if len(ledger.Anomalies) == 0 {
		lines = append(lines, "- **Status:** No execution anomalies detected (clean session).")
	} else {
		var critCount, warnCount int
		for _, a := range ledger.Anomalies {
			if a.Severity == domain.SeverityCritical {
				critCount++
			} else {
				warnCount++
			}
		}

		var breakdownParts []string
		if critCount > 0 {
			breakdownParts = append(breakdownParts, fmt.Sprintf("%d %s", critCount, domain.SeverityCritical))
		}
		if warnCount > 0 {
			breakdownParts = append(breakdownParts, fmt.Sprintf("%d %s", warnCount, domain.SeverityWarning))
		}

		statusNoun := "execution anomalies"
		if len(ledger.Anomalies) == 1 {
			statusNoun = "execution anomaly"
		}
		lines = append(lines, fmt.Sprintf("- **Status:** %d %s detected (%s).", len(ledger.Anomalies), statusNoun, strings.Join(breakdownParts, ", ")))

		for _, a := range ledger.Anomalies {
			title, description := domain.AnomalyDisplayText(a)
			lines = append(lines, fmt.Sprintf("- [%s] **%s**: %s — %s", a.Severity, a.RuleID, title, description))
		}
	}

	lines = append(lines,
		"",
		"### Agentic Execution Metrics",
		"",
	)

	if len(contributors) > 0 {
		lines = append(lines, fmt.Sprintf("**Contributors:** %s", strings.Join(contributors, ", ")), "")
	}

	lines = append(lines,
		"#### By Subagent Role",
		"| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |",
		"|---|---|---|---|---|---|---|---|---|",
	)

	for _, r := range ledger.Roles {
		lines = append(lines, fmt.Sprintf(
			"| **%s** | %d | %s | %s | %s | %.1f%% | %s | %s | %s |",
			r.Role,
			r.Invocations,
			FormatNumber(r.InputTokens),
			FormatNumber(r.OutputTokens),
			FormatNumber(r.CacheReadTokens),
			r.CacheHitRatePct,
			domain.FormatTokensK(r.PeakContextTokens),
			domain.GetHealthLabel(r.PeakContextTokens, effectiveCeiling),
			FormatCostSummary(domain.EffectiveCostSummary(r.CostSummary, r.CostUSD, r.Invocations)),
		))
	}

	lines = append(lines, fmt.Sprintf(
		"| **Total** | **%d** | **%s** | **%s** | **%s** | **%.1f%%** | **%s** | **%s** | **%s** |",
		totals.Invocations,
		FormatNumber(totals.InputTokens),
		FormatNumber(totals.OutputTokens),
		FormatNumber(totals.CacheReadTokens),
		totals.CacheHitRatePct,
		domain.FormatTokensK(totals.PeakContextTokens),
		totals.ContextHealth,
		FormatCostSummary(domain.EffectiveCostSummary(totals.CostSummary, totals.CostUSD, totals.Invocations)),
	))

	lines = append(lines,
		"",
		"#### By Model",
		"| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |",
		"|---|---|---|---|---|---|---|",
	)

	sortedModels := make([]domain.ModelMetric, len(ledger.Models))
	copy(sortedModels, ledger.Models)
	sort.Slice(sortedModels, func(i, j int) bool {
		a := domain.EffectiveCostSummary(sortedModels[i].CostSummary, sortedModels[i].CostUSD, sortedModels[i].Invocations).CompleteAmount()
		b := domain.EffectiveCostSummary(sortedModels[j].CostSummary, sortedModels[j].CostUSD, sortedModels[j].Invocations).CompleteAmount()
		if (a == nil) != (b == nil) {
			return a != nil
		}
		if a == nil || *a == *b {
			return sortedModels[i].Model < sortedModels[j].Model
		}
		return *a > *b
	})

	for _, m := range sortedModels {
		lines = append(lines, fmt.Sprintf(
			"| **%s** | %d | %s | %s | %s | %.1f%% | %s |",
			m.Model,
			m.Invocations,
			FormatNumber(m.InputTokens),
			FormatNumber(m.OutputTokens),
			FormatNumber(m.CacheReadTokens),
			m.CacheHitRatePct,
			FormatCostSummary(domain.EffectiveCostSummary(m.CostSummary, m.CostUSD, m.Invocations)),
		))
	}

	lines = append(lines,
		fmt.Sprintf(
			"| **Total** | **%d** | **%s** | **%s** | **%s** | **%.1f%%** | **%s** |",
			totals.Invocations,
			FormatNumber(totals.InputTokens),
			FormatNumber(totals.OutputTokens),
			FormatNumber(totals.CacheReadTokens),
			totals.CacheHitRatePct,
			FormatCostSummary(domain.EffectiveCostSummary(totals.CostSummary, totals.CostUSD, totals.Invocations)),
		),
		"",
	)

	return strings.Join(lines, "\n")
}

// FormatCostSummary keeps reported zero distinct from missing or partial cost.
func FormatCostSummary(summary domain.CostSummary) string {
	summary = domain.FinalizeCostSummary(summary)
	if summary.State == domain.CostSummaryEmpty {
		return "No invocation data"
	}
	if amount := summary.CompleteAmount(); amount != nil {
		return fmt.Sprintf("$%.2f", *amount)
	}
	coverage := fmt.Sprintf("%d observed, %d partial, %d unavailable / %d invocations", summary.FullyObserved, summary.Partial, summary.Unavailable, summary.Total)
	if summary.KnownSubtotalUSD != nil {
		return fmt.Sprintf("$%.2f known subtotal (partial; %s)", *summary.KnownSubtotalUSD, coverage)
	}
	return "Unavailable (" + coverage + ")"
}
