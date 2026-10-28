package domain

import (
	"fmt"
	"math"
)

// CurrentSchemaVersion defines the active telemetry schema version.
const CurrentSchemaVersion int = 4

// Migration step identifiers.
const (
	MigrationStepGitCodeFacts    = "git_code_facts"
	MigrationStepOpenCodeDB      = "opencode_db"
	MigrationStepSanitizedEvents = "schema_v3_sanitized"
	MigrationStepUserIterations  = "schema_v4_user_iterations"
	MigrationStepSourceBackfill  = "source_backfill"
)

// HealthyCeilingTokens is the threshold constant for a healthy context window (200,000 tokens).
const HealthyCeilingTokens int64 = 200_000

// FormatTokensK formats token count into human-readable compact notation (e.g., 900.8k, 200k, 1.2M, 450).
func FormatTokensK(toks int64) string {
	if toks >= 1_000_000 {
		val := float64(toks) / 1_000_000.0
		if val == math.Floor(val) {
			return fmt.Sprintf("%.0fM", val)
		}
		return fmt.Sprintf("%.1fM", val)
	}
	if toks >= 1_000 {
		val := float64(toks) / 1_000.0
		if val == math.Floor(val) {
			return fmt.Sprintf("%.0fk", val)
		}
		return fmt.Sprintf("%.1fk", val)
	}
	return fmt.Sprintf("%d", toks)
}

// GetBaseHealthStatus returns the base context health status without percentage annotations.
func GetBaseHealthStatus(toks int64, ceiling int64) ContextHealthStatus {
	if ceiling <= 0 {
		ceiling = HealthyCeilingTokens
	}
	if toks > 2*ceiling {
		return ContextHealthCritical
	}
	if toks > ceiling {
		return ContextHealthBloated
	}
	if toks >= ceiling/2 {
		return ContextHealthElevated
	}
	return ContextHealthHealthy
}

// GetHealthLabel formats a granular health label including saturation percentage for bloated/critical states.
func GetHealthLabel(toks int64, ceiling int64) string {
	if ceiling <= 0 {
		ceiling = HealthyCeilingTokens
	}
	if toks > 2*ceiling {
		pct := math.Round(float64(toks) / float64(ceiling) * 100.0)
		return fmt.Sprintf("CRITICAL (%.0f%%)", pct)
	}
	if toks > ceiling {
		pct := math.Round(float64(toks) / float64(ceiling) * 100.0)
		return fmt.Sprintf("BLOATED (%.0f%%)", pct)
	}
	if toks >= ceiling/2 {
		return string(ContextHealthElevated)
	}
	return string(ContextHealthHealthy)
}
