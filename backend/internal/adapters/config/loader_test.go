package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

func fixtureSharedBody(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func testProjectConfig() domain.ProjectConfig {
	return domain.ProjectConfig{ProjectName: "Fixture", AgenticLens: domain.EntrypointConfig{Version: "v9.8.7"}, Ticket: domain.TicketConfig{Prefixes: []domain.TicketPrefix{{Prefix: "T-", Description: "Development", TicketType: "feat"}}, CaseInsensitive: false}, Workspaces: domain.WorkspacesConfig{Dir: "records", Artifacts: []domain.WorkspaceArtifactConfig{}}, CodeMetrics: domain.CodeMetricsConfig{BaseBranch: "main", Production: domain.CodeCategoryConfig{Include: []string{"src/**/*.go"}, Exclude: []string{"**/*_test.go"}}, Tests: domain.CodeCategoryConfig{Include: []string{"**/*_test.go"}, Exclude: []string{}}}, Generation: domain.GenerationConfig{Metrics: domain.GenerationPolicy{Skills: false, Delegations: true, Anomalies: true}, Feedback: domain.GenerationFeedbackConfig{ProvisionalThresholdSeconds: 0}, Summary: domain.GenerationSummaryConfig{Enabled: true, Filename: "METRICS.md"}}, Database: domain.DatabaseConfig{BusyTimeoutMs: 0}, Live: domain.LiveConfig{ActiveThresholdSeconds: 91, PollIntervalMs: 333}, AnomalyThresholds: domain.AnomalyThresholdsConfig{StallThresholdSeconds: 47, ErrorRateMinCalls: 7, ErrorRateThreshold: .31, ErrorRateCriticalThreshold: .81, ContextCeilingWarningTokens: 170001, ContextCeilingCriticalTokens: 220003, ConsecutiveErrorsStreakWarning: 4, ConsecutiveErrorsStreakCritical: 9, ToolThrashingMinFailures: 2, ToolThrashingCriticalFailures: 6, TurnUsageWarningTokens: 31001, TurnUsageCriticalTokens: 61003, ContextGrowthWarningTokens: 41001, ContextGrowthCriticalTokens: 81003}, RoleMappings: domain.RoleMappings{}, RoleOrder: []string{}}
}
func writeConfig(t *testing.T, root, name, body string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func replaceSetting(t *testing.T, body, key, value string) string {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(key, ".")
	node := doc.Content[0]
	for i, part := range parts {
		found := false
		for j := 0; j < len(node.Content); j += 2 {
			if node.Content[j].Value != part {
				continue
			}
			found = true
			if i == len(parts)-1 {
				if value == "<missing>" {
					node.Content = append(node.Content[:j], node.Content[j+2:]...)
				} else {
					var v yaml.Node
					if err := yaml.Unmarshal([]byte(value), &v); err != nil {
						t.Fatal(err)
					}
					node.Content[j+1] = v.Content[0]
				}
			} else {
				node = node.Content[j+1]
			}
			break
		}
		if !found {
			t.Fatalf("fixture setting absent: %s", key)
		}
	}
	var out strings.Builder
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		t.Fatal(err)
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
	return out.String()
}
func loadBody(t *testing.T, body string) (domain.ProjectConfig, error) {
	t.Helper()
	path := writeConfig(t, t.TempDir(), ".agenticlens-config.yaml", body)
	return config.LoadFromFile(path)
}

func TestLoadExplicitYAMLValuesAndSourceFreePolicy(t *testing.T) {
	for _, suffix := range []string{"yaml", "yml"} {
		t.Run(suffix, func(t *testing.T) {
			root := t.TempDir()
			path := writeConfig(t, root, ".agenticlens-config."+suffix, fixtureSharedBody(t))
			cfg, err := config.Load(root)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.ConfigPath != path || cfg.ProjectName != "Fixture" || cfg.Ticket.CaseInsensitive || cfg.Ticket.Prefixes[1].Prefix != "T-" || cfg.Ticket.Prefixes[2].Prefix != "_TEAM" {
				t.Fatalf("explicit identity/origin lost: %+v", cfg)
			}
			if cfg.Database.BusyTimeoutMs != 0 || cfg.Generation.Feedback.ProvisionalThresholdSeconds != 0 || cfg.Generation.Metrics.Skills || !cfg.Generation.Metrics.Delegations || cfg.Generation.Summary.Filename != "METRICS.md" {
				t.Fatalf("explicit zero/false/output lost: %+v", cfg)
			}
			if cfg.RoleMappings == nil || len(cfg.RoleMappings) != 0 || cfg.RoleOrder == nil || len(cfg.RoleOrder) != 0 || cfg.Workspaces.Artifacts == nil {
				t.Fatal("explicit empty collections changed")
			}
			if cfg.CodeMetrics.BaseBranch != "fixture-base" || cfg.CodeMetrics.Production.Include[0] != "src/**/*.go" || cfg.AnomalyThresholds.ContextCeilingCriticalTokens != 220003 {
				t.Fatal("configured comparison/classification/threshold changed")
			}
		})
	}
	body := fixtureSharedBody(t)
	for _, key := range []string{"database", "live", "anomaly_thresholds", "generation"} {
		body = replaceSetting(t, body, key, "<missing>")
	}
	cfg, err := loadBody(t, body)
	if err != nil {
		t.Fatalf("source-free shared read required unused settings: %v", err)
	}
	for _, tc := range []struct {
		key      string
		validate func(domain.ProjectConfig) error
	}{{"database.busy_timeout_ms", config.ValidateDatabasePolicy}, {"generation.metrics.skills", config.ValidateGenerationPolicy}, {"generation.feedback.provisional_threshold_seconds", config.ValidateFeedbackPolicy}, {"generation.summary.enabled", config.ValidateSummaryPolicy}, {"live.poll_interval_ms", config.ValidatePollingPolicy}, {"live.active_threshold_seconds", config.ValidateActivityPolicy}, {"anomaly_thresholds.error_rate_min_calls", config.ValidateAnomalyPolicy}} {
		if err := tc.validate(cfg); err == nil || !strings.Contains(err.Error(), tc.key) {
			t.Errorf("consumer did not require %s: %v", tc.key, err)
		}
	}
}

func TestSharedDiscoveryAmbiguityNearestAndGitBoundary(t *testing.T) {
	root := t.TempDir()
	body := fixtureSharedBody(t)
	a := writeConfig(t, root, ".agenticlens-config.yaml", body)
	b := writeConfig(t, root, ".agenticlens-config.yml", body)
	if _, err := config.Load(root); err == nil || !strings.Contains(err.Error(), a) || !strings.Contains(err.Error(), b) {
		t.Fatalf("dual names not rejected: %v", err)
	}
	if err := os.Remove(b); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(filepath.Join(nested, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	got, err := config.FindConfigFile(filepath.Join(nested, "src"))
	if err != nil || got != a {
		t.Fatalf("nearest discovery %q %v", got, err)
	}
	bad := writeConfig(t, nested, ".agenticlens-config.yml", "project_name: [broken")
	if _, err := config.Load(filepath.Join(nested, "src")); err == nil || !strings.Contains(err.Error(), bad) {
		t.Fatalf("invalid nearest bypassed: %v", err)
	}
	if err := os.Remove(bad); err != nil {
		t.Fatal(err)
	}
	writeConfig(t, nested, ".git", "gitdir: unrelated\n")
	if _, err := config.Load(filepath.Join(nested, "src")); !errors.Is(err, config.ErrConfigNotFound) {
		t.Fatalf("crossed Git boundary: %v", err)
	}
	if err := os.Remove(filepath.Join(nested, ".git")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(nested, ".agenticlens-config.yaml"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(nested); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("invalid selected directory bypassed: %v", err)
	}
}

func TestLoadExplicitFileAndErrorBoundaries(t *testing.T) {
	for _, name := range []string{"chosen.yaml", "chosen.yml"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			path := writeConfig(t, root, name, fixtureSharedBody(t))
			if _, err := config.LoadFromFile(path); err != nil {
				t.Fatal(err)
			}
			if _, err := config.Load(root); !errors.Is(err, config.ErrConfigNotFound) {
				t.Fatalf("custom basename automatically discovered: %v", err)
			}
		})
	}
	if _, err := config.LoadFromFile(filepath.Join(t.TempDir(), "missing.yaml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing file error not wrapped: %v", err)
	}
	for _, body := range []string{"", "[]", "project_name: [bad", "project_name: A\nproject_name: B\n"} {
		if _, err := loadBody(t, body); err == nil {
			t.Fatalf("invalid document accepted %q", body)
		}
	}
	if _, err := loadBody(t, replaceSetting(t, fixtureSharedBody(t), "project_name", "<missing>")); !errors.Is(err, config.ErrInvalidConfig) {
		t.Fatalf("required configuration identity error lost: %v", err)
	}
}

func TestRequiredCurrentSharedShapeAndIntentionalEmptyLists(t *testing.T) {
	base := fixtureSharedBody(t)
	for _, key := range []string{"ticket.prefixes", "ticket.case_insensitive", "workspaces.artifacts", "code_metrics.base_branch", "code_metrics.production.include", "code_metrics.production.exclude", "code_metrics.tests.include", "code_metrics.tests.exclude", "role_mappings", "role_order"} {
		for _, value := range []string{"<missing>", "null"} {
			t.Run(key+"/"+value, func(t *testing.T) {
				body := replaceSetting(t, base, key, value)
				_, err := loadBody(t, body)
				if err == nil || !strings.Contains(err.Error(), strings.Split(key, ".")[0]) {
					t.Fatalf("missing/invalid required %s accepted or unidentified: %v", key, err)
				}
			})
		}
	}
	// A caller may select --workspaces instead; an explicitly invalid value
	// still cannot bypass ordinary shared configuration validation.
	withoutWorkspace := replaceSetting(t, base, "workspaces.dir", "<missing>")
	if _, err := loadBody(t, withoutWorkspace); err != nil {
		t.Fatalf("optional configured root blocked explicit CLI selection: %v", err)
	}
	if _, err := loadBody(t, replaceSetting(t, base, "workspaces.dir", "null")); err == nil || !strings.Contains(err.Error(), "workspaces.dir") {
		t.Fatalf("invalid configured workspace root accepted: %v", err)
	}
	for _, key := range []string{"code_metrics.production.include", "code_metrics.production.exclude", "code_metrics.tests.include", "code_metrics.tests.exclude"} {
		for _, value := range []string{"{}", "['']", "['[']", "[17]"} {
			t.Run(key+value, func(t *testing.T) {
				if _, err := loadBody(t, replaceSetting(t, base, key, value)); err == nil || !strings.Contains(err.Error(), key) {
					t.Fatalf("bad list/glob %s accepted: %v", key, err)
				}
			})
		}
	}
	body := base
	for _, key := range []string{"code_metrics.production.include", "code_metrics.production.exclude", "code_metrics.tests.include", "code_metrics.tests.exclude"} {
		body = replaceSetting(t, body, key, "[]")
	}
	cfg, err := loadBody(t, body)
	if err != nil || cfg.CodeMetrics.Production.Include == nil || cfg.CodeMetrics.Tests.Exclude == nil {
		t.Fatalf("intentional empty lists rejected: %+v %v", cfg.CodeMetrics, err)
	}
	for _, tc := range []struct{ key, value string }{{"ticket.case_insensitive", "'false'"}, {"ticket.prefixes", "[{prefix: T}]"}, {"ticket.prefixes", "[{prefix: '', description: Work}]"}, {"ticket.prefixes", "[{prefix: T, description: '', ticket_type: feat}]"}, {"ticket.prefixes", "[{prefix: T, description: Work, ticket_type: invalid}]"}, {"role_mappings", "{'': coder}"}, {"role_mappings", "{custom: ''}"}, {"role_mappings", "{custom: 5}"}, {"role_order", "['']"}, {"role_order", "{}"}, {"workspaces.artifacts", "[{path: '../outside.md'}]"}, {"workspaces.artifacts", "[{path: '/outside.md'}]"}, {"workspaces.artifacts", "[{path: ''}]"}, {"workspaces.artifacts", "[{}]"}} {
		if _, err := loadBody(t, replaceSetting(t, base, tc.key, tc.value)); err == nil {
			t.Errorf("malformed %s accepted %s", tc.key, tc.value)
		}
	}
	cfg, err = loadBody(t, replaceSetting(t, base, "workspaces.artifacts", "[{path: nested/NOTES.md}]"))
	if err != nil || cfg.Workspaces.Artifacts[0].Title != "" {
		t.Fatalf("optional artifact title was required/invented: %+v %v", cfg.Workspaces.Artifacts, err)
	}
}

func TestOperationPolicyTypeBoundsAndOrdering(t *testing.T) {
	base := fixtureSharedBody(t)
	for _, tc := range []struct {
		key      string
		values   []string
		validate func(domain.ProjectConfig) error
	}{
		{"database.busy_timeout_ms", []string{"<missing>", "null", "-1", "1.5", "'0'"}, config.ValidateDatabasePolicy},
		{"live.active_threshold_seconds", []string{"<missing>", "0", "-1", "3.0"}, config.ValidateActivityPolicy},
		{"live.poll_interval_ms", []string{"<missing>", "0", "-1"}, config.ValidatePollingPolicy},
		{"anomaly_thresholds.stall_threshold_seconds", []string{"<missing>", "0"}, config.ValidateActivityPolicy},
		{"anomaly_thresholds.error_rate_min_calls", []string{"<missing>", "0", "7.0"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.error_rate_threshold", []string{"<missing>", "0", "1.1", ".nan", ".inf", "0.9"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.error_rate_critical_threshold", []string{"<missing>", "0", "0.2"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.context_ceiling_warning_tokens", []string{"<missing>", "0", "220004"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.consecutive_errors_streak_warning", []string{"<missing>", "0", "10"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.tool_thrashing_min_failures", []string{"<missing>", "0", "7"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.turn_usage_warning_tokens", []string{"<missing>", "0", "61004"}, config.ValidateAnomalyPolicy},
		{"anomaly_thresholds.context_growth_warning_tokens", []string{"<missing>", "0", "81004"}, config.ValidateAnomalyPolicy},
	} {
		for _, value := range tc.values {
			t.Run(tc.key+value, func(t *testing.T) {
				body := replaceSetting(t, base, tc.key, value)
				cfg, err := loadBody(t, body)
				if err == nil {
					err = tc.validate(cfg)
				}
				if err == nil {
					t.Fatalf("invalid consumed policy accepted: %s=%s", tc.key, value)
				}
			})
		}
	}
	for _, key := range []string{"generation.metrics.skills", "generation.metrics.delegations", "generation.metrics.anomalies"} {
		for _, value := range []string{"<missing>", "null", "'false'", "0", "[]", "{}"} {
			t.Run(key+value, func(t *testing.T) {
				cfg, err := loadBody(t, replaceSetting(t, base, key, value))
				if err == nil {
					err = config.ValidateGenerationPolicy(cfg)
				}
				if err == nil {
					t.Fatalf("invalid generation switch accepted %s=%s", key, value)
				}
			})
		}
	}
	body := replaceSetting(t, base, "anomaly_thresholds.error_rate_threshold", "1")
	body = replaceSetting(t, body, "anomaly_thresholds.error_rate_critical_threshold", "1")
	cfg, err := loadBody(t, body)
	if err != nil || config.ValidateAnomalyPolicy(cfg) != nil {
		t.Fatalf("equal unit ratios rejected: %v", err)
	}
}

func TestLoadLocalTelemetrySelectionIsIndependent(t *testing.T) {
	root := t.TempDir()
	shared := writeConfig(t, root, ".agenticlens-config.yaml", fixtureSharedBody(t))
	local := writeConfig(t, root, ".agenticlens-config.local.yaml", "telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: selected-codex\n    opencode:\n      database_path: selected.db\n  capture:\n    enabled: false\n    state_dir: selected-state\n    max_sessions: 7\n")
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Telemetry.DefaultSource != "codex" || cfg.Telemetry.Sources.CodexHome != filepath.Join(root, "selected-codex") || cfg.Telemetry.Sources.OpenCodeDatabasePath != filepath.Join(root, "selected.db") || cfg.LocalConfigPath != local {
		t.Fatalf("local source origin lost: %+v", cfg)
	}
	raw, err := os.ReadFile(shared)
	if err != nil {
		t.Fatal(err)
	}
	writeConfig(t, root, ".agenticlens-config.local.yaml", "telemetry: [broken")
	if _, err := config.Load(root); err == nil || !strings.Contains(err.Error(), local) {
		t.Fatalf("invalid selected local bypassed: %v", err)
	}
	read, err := config.LoadShared(root)
	if err != nil || !reflect.DeepEqual(read.Ticket, cfg.Ticket) {
		t.Fatalf("source-free shared read required local: %+v %v", read, err)
	}
	after, err := os.ReadFile(shared)
	if err != nil || string(after) != string(raw) {
		t.Fatal("read modified shared policy")
	}
}

func TestDiscoveryDoesNotBypassPresentAliasesOrDanglingFiles(t *testing.T) {
	root := t.TempDir()
	body := fixtureSharedBody(t)
	shared := writeConfig(t, root, ".agenticlens-config.yaml", body)
	alias := filepath.Join(root, ".agenticlens-config.yml")
	if err := os.Symlink(shared, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(root); err == nil || !strings.Contains(err.Error(), shared) || !strings.Contains(err.Error(), alias) {
		t.Fatalf("same-file aliases bypassed ambiguity: %v", err)
	}
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}
	dangling := filepath.Join(nested, ".agenticlens-config.yaml")
	if err := os.Symlink(filepath.Join(nested, "absent-target"), dangling); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(nested); err == nil || !strings.Contains(err.Error(), dangling) {
		t.Fatalf("dangling nearest file fell through to ancestor: %v", err)
	}
}

func TestPolicyPresenceFollowsStructuralYAMLKeys(t *testing.T) {
	for _, tc := range []struct {
		key, value string
		validate   func(domain.ProjectConfig) error
	}{
		{"ticket.case_insensitive", "true", nil},
		{"generation.metrics.skills", "false", config.ValidateGenerationPolicy},
		{"database.busy_timeout_ms", "0", config.ValidateDatabasePolicy},
	} {
		t.Run(tc.key, func(t *testing.T) {
			base := fixtureSharedBody(t)
			for _, prefix := range []bool{false, true} {
				spoof := tc.key + ": " + tc.value + "\n"
				missing := replaceSetting(t, base, tc.key, "<missing>")
				if prefix {
					missing = spoof + missing
				} else {
					missing += spoof
				}
				cfg, err := loadBody(t, missing)
				if err == nil && tc.validate != nil {
					err = tc.validate(cfg)
				}
				if err == nil || !strings.Contains(err.Error(), tc.key) {
					t.Fatalf("literal dotted key supplied absent structural policy: %v", err)
				}
			}
			// A literal dotted sibling must not replace a deliberate false/zero.
			body := replaceSetting(t, base, tc.key, tc.value) + tc.key + ": wrong-type\n"
			cfg, err := loadBody(t, body)
			if err != nil {
				t.Fatal(err)
			}
			if tc.validate != nil {
				if err := tc.validate(cfg); err != nil {
					t.Fatal(err)
				}
			}
			if tc.key == "generation.metrics.skills" && cfg.Generation.Metrics.Skills {
				t.Fatal("explicit false replaced")
			}
			if tc.key == "database.busy_timeout_ms" && cfg.Database.BusyTimeoutMs != 0 {
				t.Fatal("explicit zero replaced")
			}
		})
	}
	root := t.TempDir()
	writeConfig(t, root, ".agenticlens-config.yaml", fixtureSharedBody(t))
	writeConfig(t, root, ".agenticlens-config.local.yaml", "telemetry:\n  capture:\n    state_dir: state\ntelemetry.capture.enabled: false\n")
	if _, err := config.ResolveCapture(root, ""); err == nil || !strings.Contains(err.Error(), "enabled") {
		t.Fatalf("dotted local key satisfied required capture bool: %v", err)
	}
}
