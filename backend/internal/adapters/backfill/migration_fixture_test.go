package backfill_test

import (
	"os"
	"path/filepath"
	"testing"
)

// Migration fixtures own their project policy and never borrow the developer's
// checkout. No source or capture state is required by representation/schema work.
func migrationTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	const shared = `project_name: MigrationFixture
agenticlens:
  version: v1.2.3
ticket:
  prefixes: [{prefix: "AL-", description: "Historical fixture tickets"}]
  case_insensitive: true
workspaces:
  dir: .
  artifacts: []
role_mappings: {}
role_order: []
code_metrics:
  base_branch: main
  production: {include: [], exclude: []}
  tests: {include: [], exclude: []}
generation:
  metrics: {skills: false, delegations: false, anomalies: false}
  feedback: {provisional_threshold_seconds: 0}
  summary: {enabled: true, filename: SUMMARY.md}
anomaly_thresholds:
  error_rate_min_calls: 5
  error_rate_threshold: 0.25
  error_rate_critical_threshold: 0.70
  context_ceiling_warning_tokens: 160000
  context_ceiling_critical_tokens: 200000
  consecutive_errors_streak_warning: 3
  consecutive_errors_streak_critical: 5
  tool_thrashing_min_failures: 3
  tool_thrashing_critical_failures: 5
  turn_usage_warning_tokens: 1000
  turn_usage_critical_tokens: 2000
  context_growth_warning_tokens: 500
  context_growth_critical_tokens: 1000
  stall_threshold_seconds: 60
`
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(shared), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}
