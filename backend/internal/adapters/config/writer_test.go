package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestGenerateConfigYAMLExplicitValuesRoundTrip(t *testing.T) {
	for _, text := range []string{"John's Project", "Project: # tagged", `quote " slash \\`, "null", "true", "line one\nline two"} {
		t.Run(text, func(t *testing.T) {
			cfg := testProjectConfig()
			cfg.ProjectName = text
			cfg.Workspaces.Dir = "records/space # path"
			cfg.Ticket.Prefixes = []domain.TicketPrefix{{Prefix: "T[1]-", Description: text, TicketType: "docs"}}
			cfg.RoleMappings = domain.RoleMappings{"true": "null", "builder: one": text}
			cfg.RoleOrder = []string{"true", text}
			cfg.CodeMetrics.BaseBranch = "origin/release-2.0"
			cfg.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{{Path: "nested/NOTES.md", Title: text, Description: "quoted: # details"}, {Path: "METRICS.md"}}
			cfg.Telemetry.DefaultSource = "codex"
			cfg.Telemetry.Capture.StateDir = "private-machine-state"
			rendered, err := config.GenerateConfigYAML(cfg)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			path := writeConfig(t, root, "chosen.yml", string(rendered))
			got, err := config.LoadFromFile(path)
			if err != nil {
				t.Fatalf("rendered YAML invalid: %v\n%s", err, rendered)
			}
			if got.ProjectName != text || got.Workspaces.Dir != cfg.Workspaces.Dir || !reflect.DeepEqual(got.Ticket, cfg.Ticket) || !reflect.DeepEqual(got.RoleMappings, cfg.RoleMappings) || !reflect.DeepEqual(got.RoleOrder, cfg.RoleOrder) || !reflect.DeepEqual(got.Workspaces.Artifacts, cfg.Workspaces.Artifacts) || !reflect.DeepEqual(got.CodeMetrics, cfg.CodeMetrics) {
				t.Fatalf("scalar/policy round trip changed\n got %+v\nwant %+v", got, cfg)
			}
			if got.Database.BusyTimeoutMs != 0 || got.Generation.Feedback.ProvisionalThresholdSeconds != 0 || got.Generation.Metrics != cfg.Generation.Metrics || got.Generation.Summary != cfg.Generation.Summary {
				t.Fatal("explicit zero/false/summary substituted")
			}
			if strings.Contains(string(rendered), "private-machine-state") || strings.Contains(string(rendered), "default_source:") {
				t.Fatal("shared writer leaked machine policy")
			}
			if !strings.Contains(string(rendered), "#") {
				t.Fatal("generated policy lost explanatory comments")
			}
		})
	}
}

func TestGenerateConfigYAMLRequiredAndInvalidChoices(t *testing.T) {
	cases := []struct {
		name  string
		alter func(*domain.ProjectConfig)
	}{
		{"workspace", func(c *domain.ProjectConfig) { c.Workspaces.Dir = "  " }},
		{"base", func(c *domain.ProjectConfig) { c.CodeMetrics.BaseBranch = "" }},
		{"prefix description", func(c *domain.ProjectConfig) { c.Ticket.Prefixes[0].Description = "" }},
		{"production list", func(c *domain.ProjectConfig) { c.CodeMetrics.Production.Include = nil }},
		{"test glob", func(c *domain.ProjectConfig) { c.CodeMetrics.Tests.Exclude = []string{"["} }},
		{"role mapping", func(c *domain.ProjectConfig) { c.RoleMappings = nil }},
		{"role order", func(c *domain.ProjectConfig) { c.RoleOrder = nil }},
		{"artifact list", func(c *domain.ProjectConfig) { c.Workspaces.Artifacts = nil }},
		{"artifact escape", func(c *domain.ProjectConfig) {
			c.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{{Path: "../escape"}}
		}},
		{"busy timeout", func(c *domain.ProjectConfig) { c.Database.BusyTimeoutMs = -1 }},
		{"activity", func(c *domain.ProjectConfig) { c.Live.ActiveThresholdSeconds = 0 }},
		{"polling", func(c *domain.ProjectConfig) { c.Live.PollIntervalMs = 0 }},
		{"feedback", func(c *domain.ProjectConfig) { c.Generation.Feedback.ProvisionalThresholdSeconds = -1 }},
		{"failure calls", func(c *domain.ProjectConfig) { c.AnomalyThresholds.ErrorRateMinCalls = 0 }},
		{"failure ratio", func(c *domain.ProjectConfig) { c.AnomalyThresholds.ErrorRateThreshold = .9 }},
		{"context order", func(c *domain.ProjectConfig) { c.AnomalyThresholds.ContextCeilingWarningTokens = 999999 }},
		{"usage order", func(c *domain.ProjectConfig) { c.AnomalyThresholds.TurnUsageWarningTokens = 999999 }},
		{"growth order", func(c *domain.ProjectConfig) { c.AnomalyThresholds.ContextGrowthWarningTokens = 999999 }},
		{"release pin", func(c *domain.ProjectConfig) { c.AgenticLens.Version = "dev" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testProjectConfig()
			tc.alter(&cfg)
			if _, err := config.GenerateConfigYAML(cfg); err == nil {
				t.Fatal("invalid explicit writer policy accepted")
			}
		})
	}
}

func TestGenerateConfigYAMLThresholdPrecisionAndIndependentEmptyChoices(t *testing.T) {
	cfg := testProjectConfig()
	cfg.AnomalyThresholds.ErrorRateThreshold = .35123456789
	cfg.AnomalyThresholds.ErrorRateCriticalThreshold = .81234567891
	cfg.CodeMetrics.Production.Include = []string{}
	cfg.CodeMetrics.Production.Exclude = []string{}
	cfg.CodeMetrics.Tests.Include = []string{}
	cfg.CodeMetrics.Tests.Exclude = []string{}
	cfg.Generation.Summary = domain.GenerationSummaryConfig{Enabled: false}
	cfg.RoleMappings = domain.RoleMappings{}
	cfg.RoleOrder = []string{}
	data, err := config.GenerateConfigYAML(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), ".agenticlens-config.yaml")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.AnomalyThresholds != cfg.AnomalyThresholds || !reflect.DeepEqual(got.CodeMetrics, cfg.CodeMetrics) || got.Generation.Summary.Enabled || got.Generation.Summary.Filename != "" || got.RoleMappings == nil || got.RoleOrder == nil {
		t.Fatalf("explicit precision/empty policy changed: %+v", got)
	}
}
