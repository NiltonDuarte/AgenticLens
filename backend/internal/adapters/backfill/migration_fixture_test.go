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
	const shared = "project_name: MigrationFixture\nagenticlens:\n  version: v1.2.3\nticket:\n  prefixes: [{prefix: \"AL-\", description: \"Historical fixture tickets\"}]\n  case_insensitive: true\nworkspaces:\n  roots:\n    - path: .\n      discovery:\n        max_depth: 8\n        max_directories: 300\n  artifacts: []\ndatabase:\n  busy_timeout_ms: 5000\nrole_mappings: {}\nrole_order: []\ncode_metrics:\n  base_branch: main\n  production: {include: [], exclude: []}\n  tests: {include: [], exclude: []}\ngeneration:\n  metrics: {skills: false, delegations: false, anomalies: false}\n  feedback: {provisional_threshold_seconds: 0}\n  summary: {enabled: true, filename: SUMMARY.md}\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n  consecutive_errors_streak_warning: 3\n  consecutive_errors_streak_critical: 5\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  turn_usage_warning_tokens: 1000\n  turn_usage_critical_tokens: 2000\n  context_growth_warning_tokens: 500\n  context_growth_critical_tokens: 1000\n  stall_threshold_seconds: 60\n"
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(shared), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), []byte("agenticlens:\n  state_dir: .publication-state\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}
