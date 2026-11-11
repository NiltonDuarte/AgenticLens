package domain_test

import (
	"agenticlens/internal/domain"
	"strings"
	"testing"
)

// testProjectConfig is explicit fixture policy, not runtime defaults.
func testProjectConfig() domain.ProjectConfig {
	return domain.ProjectConfig{
		ProjectName: "Fixture", AgenticLens: domain.EntrypointConfig{Version: "v9.8.7"},
		Ticket: domain.TicketConfig{CaseInsensitive: true, Prefixes: []domain.TicketPrefix{
			{Prefix: "AL", Description: "Development"}, {Prefix: "GEN", Description: "General", TicketType: "chore"}, {Prefix: "PO", Description: "Planning", TicketType: "docs"},
		}},
		Workspaces:        domain.WorkspacesConfig{Dir: "records", Artifacts: []domain.WorkspaceArtifactConfig{}},
		Database:          domain.DatabaseConfig{BusyTimeoutMs: 5000},
		Live:              domain.LiveConfig{ActiveThresholdSeconds: 300, PollIntervalMs: 2000},
		Generation:        domain.GenerationConfig{Metrics: domain.GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}, Feedback: domain.GenerationFeedbackConfig{ProvisionalThresholdSeconds: 180}, Summary: domain.GenerationSummaryConfig{Enabled: false}},
		AnomalyThresholds: domain.AnomalyThresholdsConfig{StallThresholdSeconds: 120, ErrorRateMinCalls: 5, ErrorRateThreshold: 0.25, ErrorRateCriticalThreshold: 0.70, ConsecutiveErrorsStreakWarning: 3, ConsecutiveErrorsStreakCritical: 5, ContextCeilingWarningTokens: 160000, ContextCeilingCriticalTokens: 200000, TurnUsageWarningTokens: 30000, TurnUsageCriticalTokens: 60000, ContextGrowthWarningTokens: 40000, ContextGrowthCriticalTokens: 80000, ToolThrashingMinFailures: 3, ToolThrashingCriticalFailures: 5},
		CodeMetrics:       domain.CodeMetricsConfig{BaseBranch: "main", Production: domain.CodeCategoryConfig{Include: []string{"**/*.go"}, Exclude: []string{"**/*_test.go"}}, Tests: domain.CodeCategoryConfig{Include: []string{"**/*_test.go"}, Exclude: []string{}}},
		RoleMappings:      domain.RoleMappings{"orchestrator": "orchestrator", "build": "orchestrator", "planner": "planner", "coder": "coder", "tester": "tester", "reviewer": "reviewer"},
		RoleOrder:         []string{"orchestrator", "planner", "coder", "tester", "reviewer"},
	}
}

func testMetricsProjectConfig() domain.ProjectConfig {
	cfg := testProjectConfig()
	cfg.RoleMappings = domain.RoleMappings{}
	cfg.RoleOrder = []string{}
	return cfg
}

func TestNormalizeTicketIDPreservesLiteralIdentity(t *testing.T) {
	cfg := domain.TicketConfig{CaseInsensitive: true, Prefixes: []domain.TicketPrefix{
		{Prefix: "T"}, {Prefix: "T-"}, {Prefix: "T_"}, {Prefix: "_TEAM"}, {Prefix: "AL2-"}, {Prefix: "PR.FIX"}, {Prefix: "C++"}, {Prefix: "A$B"}, {Prefix: "T[1]"}, {Prefix: "^STAR"}, {Prefix: "T1-"}, {Prefix: "AL-"}, {Prefix: "GEN-"}, {Prefix: "PO-"},
	}}
	for _, tc := range []struct{ input, want string }{
		{"T01", "T01"}, {"T-01", "T-01"}, {"T_01", "T_01"}, {"_TEAM01", "_TEAM01"}, {"AL2-0007", "AL2-0007"}, {"al2-0007", "al2-0007"},
		{"pr.fix-10", "pr.fix-10"}, {"C++-01", "C++-01"}, {"a$b-42", "a$b-42"}, {"t[1]-5", "t[1]-5"}, {"^star-8", "^star-8"},
		{"  AL-21  ", "AL-21"}, {"\t\nAL-100\r\n ", "AL-100"}, {"feat/al-21-yaml-config", "al-21"}, {"AL-123_patch", "AL-123"}, {"refs/heads/al-9-trace", "al-9"},
		{"workspaces/PO-3/agentic_metrics.json", "PO-3"}, {"bugfix/GEN-33-db-lock", "GEN-33"}, {"T1-02", "T1-02"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := domain.NormalizeTicketID(tc.input, cfg)
			if !ok || got != tc.want {
				t.Fatalf("got %q,%v want %q", got, ok, tc.want)
			}
		})
	}
	for _, input := range []string{"", "  ", "random-text", "AL-", "AL-abc", "fooal-21", "al-21foo", "prafix-10", "c123-99", "ab-42", "t1x-5", "star-8"} {
		if got, ok := domain.NormalizeTicketID(input, cfg); ok || got != "" {
			t.Errorf("unexpected recognition of %q: %q,%v", input, got, ok)
		}
	}
	cfg.CaseInsensitive = false
	if _, ok := domain.NormalizeTicketID("al-21", cfg); ok {
		t.Fatal("case-sensitive policy matched lowercase ID")
	}
	if got, ok := domain.NormalizeTicketID("AL-021", cfg); !ok || got != "AL-021" {
		t.Fatalf("case-sensitive literal recognition lost spelling: %q %v", got, ok)
	}
	for _, prefixes := range [][]domain.TicketPrefix{nil, {}, {{Prefix: "  "}}} {
		if got, ok := domain.NormalizeTicketID("AL-21", domain.TicketConfig{Prefixes: prefixes, CaseInsensitive: true}); ok || got != "" {
			t.Fatalf("invented prefix: %q %v", got, ok)
		}
	}
}

func TestLiteralPrefixMembershipOverlaps(t *testing.T) {
	ids := []string{"T01", "T02", "T-01", "T-02", "T_01", "T_02"}
	for _, tc := range []struct {
		prefix string
		want   []string
	}{
		{"T", ids}, {"T-", []string{"T-01", "T-02"}}, {"T_", []string{"T_01", "T_02"}},
	} {
		var got []string
		for _, id := range ids {
			if domain.MatchTicketPrefix(id, tc.prefix, false) {
				got = append(got, id)
			}
		}
		if strings.Join(got, ",") != strings.Join(tc.want, ",") {
			t.Fatalf("%q membership %v want %v", tc.prefix, got, tc.want)
		}
	}
}

func TestTicketTypeMappingRequiresAgreementAndExplicitCase(t *testing.T) {
	for _, tc := range []struct {
		name, id    string
		insensitive bool
		prefixes    []domain.TicketPrefix
		want        domain.ConventionalTicketType
		found, err  bool
	}{
		{"untyped", "T01", true, []domain.TicketPrefix{{Prefix: "T"}}, "", false, false},
		{"literal plain", "T01", false, []domain.TicketPrefix{{Prefix: "T", TicketType: "feat"}}, domain.TicketTypeFeat, true, false},
		{"literal hyphen", "T-01", false, []domain.TicketPrefix{{Prefix: "T-", TicketType: "docs"}}, domain.TicketTypeDocs, true, false},
		{"literal underscore", "T_01", false, []domain.TicketPrefix{{Prefix: "T_", TicketType: "fix"}}, domain.TicketTypeFix, true, false},
		{"case insensitive", "t-01", true, []domain.TicketPrefix{{Prefix: "T-", TicketType: "docs"}}, domain.TicketTypeDocs, true, false},
		{"case sensitive", "t-01", false, []domain.TicketPrefix{{Prefix: "T-", TicketType: "docs"}}, "", false, false},
		{"agree", "T-01", true, []domain.TicketPrefix{{Prefix: "T", TicketType: "docs"}, {Prefix: "T-", TicketType: "docs"}}, domain.TicketTypeDocs, true, false},
		{"conflict", "T-01", true, []domain.TicketPrefix{{Prefix: "T", TicketType: "fix"}, {Prefix: "T-", TicketType: "docs"}}, "", false, true},
		{"invalid type", "T01", false, []domain.TicketPrefix{{Prefix: "T", TicketType: "invalid"}}, "", false, true},
		{"unknown", "OTHER-1", true, []domain.TicketPrefix{{Prefix: "T", TicketType: "feat"}}, "", false, false},
		{"false prefix", "TEAM-01", true, []domain.TicketPrefix{{Prefix: "T", TicketType: "feat"}}, "", false, false},
		{"empty suffix", "T-", true, []domain.TicketPrefix{{Prefix: "T", TicketType: "feat"}}, "", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := (domain.TicketConfig{Prefixes: tc.prefixes, CaseInsensitive: tc.insensitive}).ResolveTicketTypeMapping(tc.id)
			if (err != nil) != tc.err || ok != tc.found || got != tc.want {
				t.Fatalf("got %q,%v,%v want %q,%v,error=%v", got, ok, err, tc.want, tc.found, tc.err)
			}
			if tc.name == "conflict" && (!strings.Contains(err.Error(), "T=fix") || !strings.Contains(err.Error(), "T-=docs") || !strings.Contains(err.Error(), "--ticket-type")) {
				t.Fatalf("conflict guidance incomplete: %v", err)
			}
		})
	}
}

func TestRoleMappingsPreserveUnmappedSourceNames(t *testing.T) {
	explicit := domain.RoleMappings{"build": "orchestrator", "Planner": "planning display", "spaced": " Role Display ", "": "ignored blank input"}
	for _, tc := range []struct{ raw, want string }{
		{"build", "orchestrator"}, {"Build", "Build"}, {"Planner", "planning display"}, {"planner", "planner"}, {"spaced", " Role Display "}, {" Coder ", " Coder "}, {"custom-developer", "custom-developer"}, {"RootAgent", "RootAgent"}, {"UnassignedSubAgent", "UnassignedSubAgent"}, {"UNKNOWN", "UNKNOWN"}, {"", ""}, {"   ", ""}, {"\t\n", ""},
	} {
		if got := explicit.ResolveRole(tc.raw); string(got) != tc.want {
			t.Errorf("role %q got%q want%q", tc.raw, got, tc.want)
		}
	}
	for _, mapping := range []domain.RoleMappings{nil, {}} {
		if got := mapping.ResolveRole("Reviewer-X"); got != "Reviewer-X" {
			t.Errorf("unmapped role identity rewritten: %q", got)
		}
	}
}
