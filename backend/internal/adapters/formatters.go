package adapters

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

// FormatISODatetime formats ISO timestamp string into standard YYYY-MM-DD HH:MM UTC format.
func FormatISODatetime(ts *string) string {
	if ts == nil || strings.TrimSpace(*ts) == "" {
		return "N/A"
	}
	clean := strings.TrimSpace(*ts)
	dt, err := time.Parse(time.RFC3339Nano, clean)
	if err != nil {
		dt, err = time.Parse(time.RFC3339, clean)
		if err != nil {
			return clean
		}
	}
	return dt.UTC().Format("2006-01-02 15:04 UTC")
}

// FormatCurrency formats USD currency amount.
func FormatCurrency(amount float64) string {
	if amount < 10.0 {
		return fmt.Sprintf("$%.4f", amount)
	}
	return fmt.Sprintf("$%.2f", amount)
}

// FormatPercentage formats percentage value.
func FormatPercentage(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

// FormatDuration formats duration in seconds into human-readable hours and minutes or minutes and seconds.
func FormatDuration(seconds int64) string {
	hours := seconds / 3600
	rem := seconds % 3600
	mins := rem / 60
	secs := rem % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}

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

// FormatCompactNumber formats integer into compact SI notation (e.g. 11.9k, 1.5M).
func FormatCompactNumber(val int64) string {
	absVal := val
	if absVal < 0 {
		absVal = -absVal
	}
	var formatted string
	if absVal >= 1_000_000 {
		num := float64(absVal) / 1_000_000.0
		if num == math.Floor(num) {
			formatted = fmt.Sprintf("%.0fM", num)
		} else {
			formatted = fmt.Sprintf("%.1fM", num)
		}
	} else if absVal >= 1_000 {
		num := float64(absVal) / 1_000.0
		if num == math.Floor(num) {
			formatted = fmt.Sprintf("%.0fk", num)
		} else {
			formatted = fmt.Sprintf("%.1fk", num)
		}
	} else {
		formatted = fmt.Sprintf("%d", absVal)
	}

	if val < 0 {
		return "-" + formatted
	}
	return formatted
}

// FormatCodeDelivery formats code delivered lines added and deleted with compact notation.
func FormatCodeDelivery(linesAdded, linesDeleted int64) string {
	return fmt.Sprintf("+%s / -%s", FormatCompactNumber(linesAdded), FormatCompactNumber(linesDeleted))
}

// FormatTicketLabel formats combined ticket identifier and human-readable title.
func FormatTicketLabel(ticketID, ticketTitle *string) string {
	idClean := ""
	if ticketID != nil {
		idClean = strings.TrimSpace(*ticketID)
	}
	titleClean := ""
	if ticketTitle != nil {
		titleClean = strings.TrimSpace(*ticketTitle)
	}
	if idClean != "" && titleClean != "" {
		return fmt.Sprintf("%s - %s", idClean, titleClean)
	}
	if idClean != "" {
		return idClean
	}
	if titleClean != "" {
		return titleClean
	}
	return "Unknown"
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
	for _, i := range ledger.Facts.InvocationFacts() {
		if i.User != "" && i.User != "unknown" {
			contributorSet[i.User] = struct{}{}
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

	lines := []string{
		"### Workflow Efficiency",
		fmt.Sprintf("- **Software Development Life Cycle Iterations:** %d", we.SDLCIterations),
		fmt.Sprintf("- **Active Agent Duration:** %s (excludes user wait time)", we.ActiveDurationFormatted),
		fmt.Sprintf(
			"- **Context Window Health:** %s — Root agent peak %s tokens (%s of %s healthy ceiling, %d compactions) | Subagents peak %s (%s)",
			cwh.Status,
			rootTokStr,
			rootPctStr,
			domain.FormatTokensK(domain.HealthyCeilingTokens),
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
			lines = append(lines, fmt.Sprintf("- [%s] **%s**: %s — %s", a.Severity, a.RuleID, a.Title, a.Description))
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
			"| **%s** | %d | %s | %s | %s | %.1f%% | %s | %s | $%.2f |",
			r.Role,
			r.Invocations,
			FormatNumber(r.InputTokens),
			FormatNumber(r.OutputTokens),
			FormatNumber(r.CacheReadTokens),
			r.CacheHitRatePct,
			domain.FormatTokensK(r.PeakContextTokens),
			domain.GetHealthLabel(r.PeakContextTokens, domain.HealthyCeilingTokens),
			r.CostUSD,
		))
	}

	lines = append(lines, fmt.Sprintf(
		"| **Total** | **%d** | **%s** | **%s** | **%s** | **%.1f%%** | **%s** | **%s** | **$%.2f** |",
		totals.Invocations,
		FormatNumber(totals.InputTokens),
		FormatNumber(totals.OutputTokens),
		FormatNumber(totals.CacheReadTokens),
		totals.CacheHitRatePct,
		domain.FormatTokensK(totals.PeakContextTokens),
		totals.ContextHealth,
		totals.CostUSD,
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
		return sortedModels[i].CostUSD > sortedModels[j].CostUSD
	})

	for _, m := range sortedModels {
		lines = append(lines, fmt.Sprintf(
			"| **%s** | %d | %s | %s | %s | %.1f%% | $%.2f |",
			m.Model,
			m.Invocations,
			FormatNumber(m.InputTokens),
			FormatNumber(m.OutputTokens),
			FormatNumber(m.CacheReadTokens),
			m.CacheHitRatePct,
			m.CostUSD,
		))
	}

	lines = append(lines,
		fmt.Sprintf(
			"| **Total** | **%d** | **%s** | **%s** | **%s** | **%.1f%%** | **$%.2f** |",
			totals.Invocations,
			FormatNumber(totals.InputTokens),
			FormatNumber(totals.OutputTokens),
			FormatNumber(totals.CacheReadTokens),
			totals.CacheHitRatePct,
			totals.CostUSD,
		),
		"",
	)

	return strings.Join(lines, "\n")
}
