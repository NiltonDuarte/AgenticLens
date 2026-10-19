package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestDefaultProjectConfig(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultProjectConfig()

	if cfg.ProjectName != "AgenticLens" {
		t.Errorf("expected ProjectName AgenticLens, got %s", cfg.ProjectName)
	}
	if len(cfg.Ticket.Prefixes) != 3 {
		t.Fatalf("expected 3 ticket prefixes, got %d: %v", len(cfg.Ticket.Prefixes), cfg.Ticket.Prefixes)
	}
	expectedPrefixes := []struct {
		prefix     string
		desc       string
		ticketType string
	}{
		{"AL", "Core AgenticLens development features, bugfixes, and technical initiatives", ""},
		{"GEN", "General multi-agent engineering tasks and ad-hoc untracked work", "chore"},
		{"PO", "Product Owner specifications, requirements definitions, and roadmap breakdowns", "docs"},
	}
	for i, exp := range expectedPrefixes {
		if cfg.Ticket.Prefixes[i].Prefix != exp.prefix {
			t.Errorf("expected prefix[%d] %s, got %s", i, exp.prefix, cfg.Ticket.Prefixes[i].Prefix)
		}
		if cfg.Ticket.Prefixes[i].Description != exp.desc {
			t.Errorf("expected description[%d] %q, got %q", i, exp.desc, cfg.Ticket.Prefixes[i].Description)
		}
		if cfg.Ticket.Prefixes[i].TicketType != exp.ticketType {
			t.Errorf("expected ticketType[%d] %q, got %q", i, exp.ticketType, cfg.Ticket.Prefixes[i].TicketType)
		}
	}
	names := cfg.Ticket.PrefixNames()
	if len(names) != 3 || names[0] != "AL" || names[1] != "GEN" || names[2] != "PO" {
		t.Errorf("unexpected PrefixNames: %v", names)
	}
	if !cfg.Ticket.CaseInsensitive {
		t.Errorf("expected CaseInsensitive to be true")
	}
	if len(cfg.Ticket.CustomPatterns) != 0 {
		t.Errorf("expected empty custom patterns, got %v", cfg.Ticket.CustomPatterns)
	}
	if cfg.Workspaces.Dir != "workspaces" {
		t.Errorf("expected Workspaces.Dir workspaces, got %s", cfg.Workspaces.Dir)
	}
	if len(cfg.Workspaces.SrcFolders) != 2 || cfg.Workspaces.SrcFolders[0] != "web" || cfg.Workspaces.SrcFolders[1] != "backend" {
		t.Errorf("unexpected Workspaces.SrcFolders: %v", cfg.Workspaces.SrcFolders)
	}
	if cfg.Database.Path != "~/.local/share/opencode/opencode.db" {
		t.Errorf("unexpected Database.Path: %s", cfg.Database.Path)
	}
	if cfg.Database.QueryTimeoutMs != 5000 {
		t.Errorf("expected QueryTimeoutMs 5000, got %d", cfg.Database.QueryTimeoutMs)
	}
	if cfg.Database.BusyTimeoutMs != 5000 {
		t.Errorf("expected BusyTimeoutMs 5000, got %d", cfg.Database.BusyTimeoutMs)
	}
	if !cfg.Database.ReadOnly {
		t.Errorf("expected Database.ReadOnly to be true")
	}
	if cfg.Live.ActiveThresholdSeconds != 300 {
		t.Errorf("expected ActiveThresholdSeconds 300, got %d", cfg.Live.ActiveThresholdSeconds)
	}
	if cfg.Live.PollIntervalMs != 2000 {
		t.Errorf("expected PollIntervalMs 2000, got %d", cfg.Live.PollIntervalMs)
	}
	if cfg.Live.MaxTraceBufferMessages != 500 {
		t.Errorf("expected MaxTraceBufferMessages 500, got %d", cfg.Live.MaxTraceBufferMessages)
	}
	if cfg.Live.AnomalyThresholds.StallThresholdSeconds != 120 {
		t.Errorf("expected StallThresholdSeconds 120, got %d", cfg.Live.AnomalyThresholds.StallThresholdSeconds)
	}
	if cfg.Live.AnomalyThresholds.ErrorRateThreshold != 0.25 {
		t.Errorf("expected ErrorRateThreshold 0.25, got %f", cfg.Live.AnomalyThresholds.ErrorRateThreshold)
	}
	if cfg.Live.AnomalyThresholds.ContextExhaustionPct != 0.85 {
		t.Errorf("expected ContextExhaustionPct 0.85, got %f", cfg.Live.AnomalyThresholds.ContextExhaustionPct)
	}
	if len(cfg.RoleMappings) == 0 {
		t.Errorf("expected non-empty RoleMappings")
	}
	expectedRoleOrder := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}
	if len(cfg.RoleOrder) != len(expectedRoleOrder) {
		t.Fatalf("expected %d RoleOrder items, got %d", len(expectedRoleOrder), len(cfg.RoleOrder))
	}
	for i, r := range expectedRoleOrder {
		if cfg.RoleOrder[i] != r {
			t.Errorf("expected RoleOrder[%d] %s, got %s", i, r, cfg.RoleOrder[i])
		}
	}
}

func TestNormalizeTicketID_ExactMatches(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultProjectConfig().Ticket

	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"AL-21", "AL-21", true},
		{"GEN-100", "GEN-100", true},
		{"AL-1", "AL-1", true},
		{"GEN-42", "GEN-42", true},
		{"PO-1", "PO-1", true},
		{"PO-4", "PO-4", true},
	}

	for _, tt := range tests {
		got, ok := domain.NormalizeTicketID(tt.input, cfg)
		if ok != tt.valid || got != tt.expected {
			t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, %v)", tt.input, got, ok, tt.expected, tt.valid)
		}
	}
}

func TestNormalizeTicketID_CaseInsensitivity(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultProjectConfig().Ticket

	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"al-21", "AL-21", true},
		{"Al-21", "AL-21", true},
		{"aL-21", "AL-21", true},
		{"gen-5", "GEN-5", true},
		{"gEn-99", "GEN-99", true},
		{"po-1", "PO-1", true},
		{"Po-2", "PO-2", true},
		{"pO-3", "PO-3", true},
		{"PO-4", "PO-4", true},
	}

	for _, tt := range tests {
		got, ok := domain.NormalizeTicketID(tt.input, cfg)
		if ok != tt.valid || got != tt.expected {
			t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, %v)", tt.input, got, ok, tt.expected, tt.valid)
		}
	}
}

func TestNormalizeTicketID_BranchAndTokenPatterns(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultProjectConfig().Ticket

	tests := []struct {
		input    string
		expected string
		valid    bool
	}{
		{"feat/al-21-yaml-config", "AL-21", true},
		{"bugfix/GEN-33-db-lock", "GEN-33", true},
		{"refs/heads/al-9-trace", "AL-9", true},
		{"workspaces/AL-15", "AL-15", true},
		{"feature/AL-100", "AL-100", true},
		{"chore/al-4-cleanup", "AL-4", true},
		{"spec/po-1-rewrite-core", "PO-1", true},
		{"feature/po-2-planning", "PO-2", true},
		{"workspaces/PO-3/agentic_metrics.json", "PO-3", true},
	}

	for _, tt := range tests {
		got, ok := domain.NormalizeTicketID(tt.input, cfg)
		if ok != tt.valid || got != tt.expected {
			t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, %v)", tt.input, got, ok, tt.expected, tt.valid)
		}
	}
}

func TestNormalizeTicketID_CaseSensitiveMode(t *testing.T) {
	t.Parallel()
	cfg := domain.TicketConfig{
		Prefixes: []domain.TicketPrefix{
			{Prefix: "AL"},
			{Prefix: "GEN"},
			{Prefix: "PO"},
		},
		CaseInsensitive: false,
	}

	got, ok := domain.NormalizeTicketID("al-21", cfg)
	if ok || got != "" {
		t.Errorf("expected al-21 to fail in case sensitive mode, got (%q, %v)", got, ok)
	}

	got, ok = domain.NormalizeTicketID("AL-21", cfg)
	if !ok || got != "AL-21" {
		t.Errorf("expected AL-21 to match in case sensitive mode, got (%q, %v)", got, ok)
	}

	got, ok = domain.NormalizeTicketID("po-1", cfg)
	if ok || got != "" {
		t.Errorf("expected po-1 to fail in case sensitive mode, got (%q, %v)", got, ok)
	}

	got, ok = domain.NormalizeTicketID("PO-1", cfg)
	if !ok || got != "PO-1" {
		t.Errorf("expected PO-1 to match in case sensitive mode, got (%q, %v)", got, ok)
	}
}

func TestNormalizeTicketID_NegativeCases(t *testing.T) {
	t.Parallel()
	cfg := domain.DefaultProjectConfig().Ticket

	tests := []string{
		"PROJ-123",
		"random-text",
		"ticket-",
		"",
		"   ",
		"AL-",
		"fooal-21",
		"al-21foo",
	}

	for _, input := range tests {
		got, ok := domain.NormalizeTicketID(input, cfg)
		if ok || got != "" {
			t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected empty and false", input, got, ok)
		}
	}
}

func TestNormalizeTicketID_CustomPatterns(t *testing.T) {
	t.Parallel()
	cfg := domain.TicketConfig{
		Prefixes: []domain.TicketPrefix{
			{Prefix: "AL"},
		},
		CaseInsensitive: true,
		CustomPatterns: []string{
			`ISSUE-(\d+)`,
		},
	}

	got, ok := domain.NormalizeTicketID("ISSUE-456", cfg)
	if !ok || got != "456" {
		t.Errorf("expected custom pattern match 456, got (%q, %v)", got, ok)
	}
}

func TestNormalizeTicketID_EmptyPrefixFallback(t *testing.T) {
	t.Parallel()
	cfg := domain.TicketConfig{
		Prefixes:        []domain.TicketPrefix{},
		CaseInsensitive: true,
	}

	got, ok := domain.NormalizeTicketID("al-10", cfg)
	if !ok || got != "AL-10" {
		t.Errorf("expected fallback to AL prefix, got (%q, %v)", got, ok)
	}

	got, ok = domain.NormalizeTicketID("gen-42", cfg)
	if !ok || got != "GEN-42" {
		t.Errorf("expected fallback to GEN prefix, got (%q, %v)", got, ok)
	}

	got, ok = domain.NormalizeTicketID("po-4", cfg)
	if !ok || got != "PO-4" {
		t.Errorf("expected fallback to PO prefix, got (%q, %v)", got, ok)
	}
}

func TestRoleMappings_ResolveRole(t *testing.T) {
	t.Parallel()
	mappings := domain.DefaultProjectConfig().RoleMappings

	// Default role mappings
	if mappings.ResolveRole("planner") != domain.RolePlanner {
		t.Errorf("expected RolePlanner, got %s", mappings.ResolveRole("planner"))
	}
	if mappings.ResolveRole("coder") != domain.RoleCoder {
		t.Errorf("expected RoleCoder, got %s", mappings.ResolveRole("coder"))
	}
	if mappings.ResolveRole("tester") != domain.RoleTester {
		t.Errorf("expected RoleTester, got %s", mappings.ResolveRole("tester"))
	}
	if mappings.ResolveRole("reviewer") != domain.RoleReviewer {
		t.Errorf("expected RoleReviewer, got %s", mappings.ResolveRole("reviewer"))
	}
	if mappings.ResolveRole("orchestrator") != domain.RoleOrchestrator {
		t.Errorf("expected RoleOrchestrator, got %s", mappings.ResolveRole("orchestrator"))
	}
	if mappings.ResolveRole("build") != domain.RoleOrchestrator {
		t.Errorf("expected RoleOrchestrator for build, got %s", mappings.ResolveRole("build"))
	}
	if mappings.ResolveRole("root") != domain.AgentRole("root") {
		t.Errorf("expected unmapped role root for root, got %s", mappings.ResolveRole("root"))
	}

	// Empty string safety
	if mappings.ResolveRole("") != domain.RoleUnassigned {
		t.Errorf("expected RoleUnassigned for empty string, got %s", mappings.ResolveRole(""))
	}
	if mappings.ResolveRole("   ") != domain.RoleUnassigned {
		t.Errorf("expected RoleUnassigned for whitespace, got %s", mappings.ResolveRole("   "))
	}

	// Custom mapping
	customMappings := domain.RoleMappings{
		"agent-qa": "tester",
		"architect": "planner",
	}
	if customMappings.ResolveRole("agent-qa") != domain.RoleTester {
		t.Errorf("expected RoleTester for agent-qa, got %s", customMappings.ResolveRole("agent-qa"))
	}
	if customMappings.ResolveRole("architect") != domain.RolePlanner {
		t.Errorf("expected RolePlanner for architect, got %s", customMappings.ResolveRole("architect"))
	}

	// Unknown role cast
	if customMappings.ResolveRole("unseen-agent") != domain.AgentRole("unseen-agent") {
		t.Errorf("expected unseen-agent cast, got %s", customMappings.ResolveRole("unseen-agent"))
	}

	// Nil mappings safety
	var nilMappings domain.RoleMappings
	if nilMappings.ResolveRole("coder") != domain.RoleCoder {
		t.Errorf("expected RoleCoder from nil mappings, got %s", nilMappings.ResolveRole("coder"))
	}
}

func TestNormalizeTicketID_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("mixed casing and leading zeros", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig().Ticket
		cases := []struct {
			input    string
			expected string
		}{
			{"al-1", "AL-1"},
			{"AL-01", "AL-01"},
			{"Al-999", "AL-999"},
			{"aL-00042", "AL-00042"},
			{"al-0", "AL-0"},
			{"gen-0", "GEN-0"},
			{"GeN-42", "GEN-42"},
		}
		for _, tc := range cases {
			got, ok := domain.NormalizeTicketID(tc.input, cfg)
			if !ok || got != tc.expected {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, true)", tc.input, got, ok, tc.expected)
			}
		}
	})

	t.Run("surrounding whitespace and control characters", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig().Ticket
		cases := []struct {
			input    string
			expected string
		}{
			{"  AL-21  ", "AL-21"},
			{"\t\nAL-100\r\n ", "AL-100"},
			{"   al-1   ", "AL-1"},
			{"\r\n\tgen-99\t ", "GEN-99"},
		}
		for _, tc := range cases {
			got, ok := domain.NormalizeTicketID(tc.input, cfg)
			if !ok || got != tc.expected {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, true)", tc.input, got, ok, tc.expected)
			}
		}
	})

	t.Run("prefixes containing special regex characters", func(t *testing.T) {
		cfg := domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "PR.FIX"},
				{Prefix: "C++"},
				{Prefix: "A$B"},
				{Prefix: "FOO+BAR"},
				{Prefix: "T[1]"},
				{Prefix: "^STAR"},
			},
			CaseInsensitive: true,
		}

		validCases := []struct {
			input    string
			expected string
		}{
			{"pr.fix-10", "PR.FIX-10"},
			{"PR.FIX-99", "PR.FIX-99"},
			{"c++-99", "C++-99"},
			{"C++-01", "C++-01"},
			{"a$b-42", "A$B-42"},
			{"A$B-7", "A$B-7"},
			{"foo+bar-7", "FOO+BAR-7"},
			{"t[1]-5", "T[1]-5"},
			{"^star-8", "^STAR-8"},
		}
		for _, tc := range validCases {
			got, ok := domain.NormalizeTicketID(tc.input, cfg)
			if !ok || got != tc.expected {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, true)", tc.input, got, ok, tc.expected)
			}
		}

		// Ensure special characters were properly escaped and not treated as regex operators
		invalidCases := []string{
			"prafix-10", // dot should not match 'a'
			"c123-99",   // plus should not match repeated characters
			"c-99",
			"ab-42", // dollar should not match empty
			"t1-5",  // brackets should not match character class
			"star-8",
		}
		for _, input := range invalidCases {
			got, ok := domain.NormalizeTicketID(input, cfg)
			if ok || got != "" {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected false and empty", input, got, ok)
			}
		}
	})

	t.Run("varied branch names and path expressions", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig().Ticket
		cases := []struct {
			input    string
			expected string
		}{
			{"feature/AL-123_patch", "AL-123"},
			{"bugfix/al-99-fix", "AL-99"},
			{"chore/GEN-42-bump", "GEN-42"},
			{"release/v2.1.0/AL-7", "AL-7"},
			{"al-55", "AL-55"},
			{"user/john/al-12-test-branch", "AL-12"},
			{"HEAD->feat/AL-88", "AL-88"},
			{"origin/feat/al-12", "AL-12"},
			{"workspaces/AL-15/agentic_metrics.json", "AL-15"},
			{"refs/pull/42/merge-al-99-done", "AL-99"},
			{"spec/po-1-rewrite-core", "PO-1"},
			{"feature/po-2-planning", "PO-2"},
			{"workspaces/PO-3/agentic_metrics.json", "PO-3"},
		}
		for _, tc := range cases {
			got, ok := domain.NormalizeTicketID(tc.input, cfg)
			if !ok || got != tc.expected {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected (%q, true)", tc.input, got, ok, tc.expected)
			}
		}
	})

	t.Run("custom pattern regexes with and without groups", func(t *testing.T) {
		cfg := domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "AL"},
			},
			CaseInsensitive: true,
			CustomPatterns: []string{
				`ISSUE-(\d+)`,
				`(?:JIRA|PROJ)-\d+`,
				`[unclosed(regex`, // invalid regex should be skipped without panic
				"",                // empty string should be skipped
				"   ",             // whitespace should be skipped
			},
		}

		// Pattern with capture group returns group 1
		got, ok := domain.NormalizeTicketID("ISSUE-456", cfg)
		if !ok || got != "456" {
			t.Errorf("expected 456, got (%q, %v)", got, ok)
		}

		// Pattern without capture group returns full match
		got, ok = domain.NormalizeTicketID("JIRA-1234", cfg)
		if !ok || got != "JIRA-1234" {
			t.Errorf("expected JIRA-1234, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("PROJ-999", cfg)
		if !ok || got != "PROJ-999" {
			t.Errorf("expected PROJ-999, got (%q, %v)", got, ok)
		}
	})

	t.Run("hostile malformed inputs and non-matching numbers", func(t *testing.T) {
		cfg := domain.DefaultProjectConfig().Ticket
		invalidInputs := []string{
			"",
			"   ",
			"\t\n\r",
			"AL-",
			"-123",
			"AL--123",
			"AL-abc",
			"AL-12a",
			"fooal-21",
			"al-21foo",
			"AL",
			"AL 21",
			"AL_21",
			"!@#$%^&*()",
			"12345",
			"9999999",
			"OTHER-123",
			"XYZ-1",
		}
		for _, input := range invalidInputs {
			got, ok := domain.NormalizeTicketID(input, cfg)
			if ok || got != "" {
				t.Errorf("NormalizeTicketID(%q) = (%q, %v), expected false and empty", input, got, ok)
			}
		}
	})

	t.Run("prefix collision and length priority", func(t *testing.T) {
		cfg := domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "AL"},
				{Prefix: "ALTEST"},
				{Prefix: "PO"},
				{Prefix: "POSTORE"},
			},
			CaseInsensitive: true,
		}

		got, ok := domain.NormalizeTicketID("altest-100", cfg)
		if !ok || got != "ALTEST-100" {
			t.Errorf("expected ALTEST-100, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("al-100", cfg)
		if !ok || got != "AL-100" {
			t.Errorf("expected AL-100, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("postore-99", cfg)
		if !ok || got != "POSTORE-99" {
			t.Errorf("expected POSTORE-99, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("po-99", cfg)
		if !ok || got != "PO-99" {
			t.Errorf("expected PO-99, got (%q, %v)", got, ok)
		}
	})

	t.Run("prefixes with trailing separators or whitespace", func(t *testing.T) {
		cfg := domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "AL-"},
				{Prefix: "GEN_"},
				{Prefix: "  PO  "},
				{Prefix: "  TICK  "},
			},
			CaseInsensitive: true,
		}

		got, ok := domain.NormalizeTicketID("al-50", cfg)
		if !ok || got != "AL-50" {
			t.Errorf("expected AL-50, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("gen-12", cfg)
		if !ok || got != "GEN-12" {
			t.Errorf("expected GEN-12, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("po-5", cfg)
		if !ok || got != "PO-5" {
			t.Errorf("expected PO-5, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("tick-9", cfg)
		if !ok || got != "TICK-9" {
			t.Errorf("expected TICK-9, got (%q, %v)", got, ok)
		}
	})

	t.Run("prefixes slice with only whitespace falls back to defaults", func(t *testing.T) {
		cfg := domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{Prefix: "  "},
				{Prefix: ""},
			},
			CaseInsensitive: true,
		}

		got, ok := domain.NormalizeTicketID("al-10", cfg)
		if !ok || got != "AL-10" {
			t.Errorf("expected AL-10 via fallback, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("gen-42", cfg)
		if !ok || got != "GEN-42" {
			t.Errorf("expected GEN-42 via fallback, got (%q, %v)", got, ok)
		}

		got, ok = domain.NormalizeTicketID("po-1", cfg)
		if !ok || got != "PO-1" {
			t.Errorf("expected PO-1 via fallback, got (%q, %v)", got, ok)
		}
	})
}

func TestRoleMappings_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("casing and whitespace variations", func(t *testing.T) {
		mappings := domain.DefaultProjectConfig().RoleMappings

		tests := []struct {
			input    string
			expected domain.AgentRole
		}{
			{"PLANNER", domain.RolePlanner},
			{"Planner", domain.RolePlanner},
			{"  planner  ", domain.RolePlanner},
			{"\tplanner\n", domain.RolePlanner},
			{"CODER", domain.RoleCoder},
			{"Coder", domain.RoleCoder},
			{"  coder  ", domain.RoleCoder},
			{"TESTER", domain.RoleTester},
			{"Tester", domain.RoleTester},
			{"tEsTeR", domain.RoleTester},
			{"REVIEWER", domain.RoleReviewer},
			{"Reviewer", domain.RoleReviewer},
			{"ORCHESTRATOR", domain.RoleOrchestrator},
			{"Orchestrator", domain.RoleOrchestrator},
			{"BUILD", domain.RoleOrchestrator},
			{"ROOT", domain.AgentRole("root")},
			{"  build  ", domain.RoleOrchestrator},
			{"  root  ", domain.AgentRole("root")},
		}

		for _, tt := range tests {
			got := mappings.ResolveRole(tt.input)
			if got != tt.expected {
				t.Errorf("ResolveRole(%q) = %s, expected %s", tt.input, got, tt.expected)
			}
		}
	})

	t.Run("empty and whitespace strings default to unassigned", func(t *testing.T) {
		mappings := domain.DefaultProjectConfig().RoleMappings

		emptyInputs := []string{"", " ", "   ", "\t", "\n", "\r\n"}
		for _, input := range emptyInputs {
			got := mappings.ResolveRole(input)
			if got != domain.RoleUnassigned {
				t.Errorf("ResolveRole(%q) = %s, expected RoleUnassigned", input, got)
			}
		}
	})

	t.Run("custom overrides with mixed casing and fallback", func(t *testing.T) {
		custom := domain.RoleMappings{
			"coder":          "custom-developer",
			"Agent-QA":       "tester",
			"automation-bot": "REVIEWER",
		}

		if custom.ResolveRole("coder") != domain.AgentRole("custom-developer") {
			t.Errorf("expected custom-developer, got %s", custom.ResolveRole("coder"))
		}
		if custom.ResolveRole("agent-qa") != domain.RoleTester {
			t.Errorf("expected RoleTester for agent-qa, got %s", custom.ResolveRole("agent-qa"))
		}
		if custom.ResolveRole("AGENT-QA") != domain.RoleTester {
			t.Errorf("expected RoleTester for AGENT-QA, got %s", custom.ResolveRole("AGENT-QA"))
		}
		if custom.ResolveRole("automation-bot") != domain.RoleReviewer {
			t.Errorf("expected RoleReviewer for automation-bot, got %s", custom.ResolveRole("automation-bot"))
		}
		if custom.ResolveRole("unknown-persona") != domain.AgentRole("unknown-persona") {
			t.Errorf("expected unknown-persona cast, got %s", custom.ResolveRole("unknown-persona"))
		}
		if custom.ResolveRole("  UNKNOWN-PERSONA  ") != domain.AgentRole("unknown-persona") {
			t.Errorf("expected lowercase unknown-persona cast, got %s", custom.ResolveRole("  UNKNOWN-PERSONA  "))
		}
	})

	t.Run("nil and empty role mappings resilience", func(t *testing.T) {
		var nilMap domain.RoleMappings
		if nilMap.ResolveRole("planner") != domain.RolePlanner {
			t.Errorf("expected RolePlanner, got %s", nilMap.ResolveRole("planner"))
		}
		if nilMap.ResolveRole("coder") != domain.RoleCoder {
			t.Errorf("expected RoleCoder, got %s", nilMap.ResolveRole("coder"))
		}
		if nilMap.ResolveRole("") != domain.RoleUnassigned {
			t.Errorf("expected RoleUnassigned for empty string, got %s", nilMap.ResolveRole(""))
		}
		if nilMap.ResolveRole("arbitrary") != domain.AgentRole("arbitrary") {
			t.Errorf("expected arbitrary cast, got %s", nilMap.ResolveRole("arbitrary"))
		}

		emptyMap := domain.RoleMappings{}
		if emptyMap.ResolveRole("reviewer") != domain.RoleReviewer {
			t.Errorf("expected RoleReviewer, got %s", emptyMap.ResolveRole("reviewer"))
		}
		if emptyMap.ResolveRole("") != domain.RoleUnassigned {
			t.Errorf("expected RoleUnassigned for empty string, got %s", emptyMap.ResolveRole(""))
		}
	})
}

func TestInferTicketTypeFromPrefix(t *testing.T) {
	t.Parallel()

	defaultCfg := domain.DefaultProjectConfig().Ticket

	tests := []struct {
		name       string
		ticketID   string
		prefixes   []domain.TicketPrefix
		expected   domain.ConventionalTicketType
		expectedOk bool
	}{
		// Default configured prefixes: PO -> docs, GEN -> chore, AL -> empty
		{name: "PO-1 uppercase", ticketID: "PO-1", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "po-42 lowercase", ticketID: "po-42", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "Po-100 mixed case", ticketID: "Po-100", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "PO-5 with whitespace", ticketID: "  PO-5  ", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "PO-ARCH prefix", ticketID: "PO-ARCH", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "PO-DOCS-SPEC prefix", ticketID: "PO-DOCS-SPEC", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "po-999 with carriage return and tabs", ticketID: "\r\n  po-999  \t", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "PO_1 underscore delimiter", ticketID: "PO_1", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "po_42 lowercase underscore", ticketID: "po_42", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeDocs, expectedOk: true},

		{name: "GEN-100 uppercase", ticketID: "GEN-100", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "gen-5 lowercase", ticketID: "gen-5", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "Gen-42 mixed case", ticketID: "Gen-42", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "GEN-1 with whitespace", ticketID: "\tGEN-1\n", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "GEN-INFRA-RENAME prefix", ticketID: "GEN-INFRA-RENAME", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "gen-888 with carriage return and tabs", ticketID: "\r\n  gen-888  \t", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},
		{name: "GEN_1 underscore delimiter", ticketID: "GEN_1", prefixes: defaultCfg.Prefixes, expected: domain.TicketTypeChore, expectedOk: true},

		// AL has no configured ticket_type
		{name: "AL-1 has no configured ticket type", ticketID: "AL-1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "al-11 lowercase no ticket type", ticketID: "al-11", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},

		// False prefix matches (MUST NOT match PO or GEN)
		{name: "false prefix POD-1", ticketID: "POD-1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix GENTLE-1", ticketID: "GENTLE-1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix POST-1", ticketID: "POST-1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix GENESIS-1", ticketID: "GENESIS-1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix POW-10", ticketID: "POW-10", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix GENERAL-5", ticketID: "GENERAL-5", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix PO1 without delimiter", ticketID: "PO1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix GEN1 without delimiter", ticketID: "GEN1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix POLICE-99", ticketID: "POLICE-99", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "false prefix GENETIC-42", ticketID: "GENETIC-42", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},

		// Delimiter without suffix or whitespace suffix
		{name: "just PO- with empty suffix", ticketID: "PO-", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "just GEN- with empty suffix", ticketID: "GEN-", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "PO- with trailing spaces only", ticketID: "PO-   ", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "GEN- with trailing spaces only", ticketID: "GEN-   ", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "PO- with control characters only", ticketID: "\tPO-\n", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "GEN- with control characters only", ticketID: "\tGEN-\n", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "just PO_ with empty suffix", ticketID: "PO_", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "just GEN_ with empty suffix", ticketID: "GEN_", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},

		// Non-supported delimiters (slash, dot, etc.)
		{name: "PO/1 slash delimiter", ticketID: "PO/1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "GEN/1 slash delimiter", ticketID: "GEN/1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "PO.1 dot delimiter", ticketID: "PO.1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "GEN.1 dot delimiter", ticketID: "GEN.1", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},

		// Empty/whitespace ticketID
		{name: "empty ticket", ticketID: "", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "whitespace only", ticketID: "   ", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "just PO without delimiter", ticketID: "PO", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},
		{name: "just GEN without delimiter", ticketID: "GEN", prefixes: defaultCfg.Prefixes, expected: "", expectedOk: false},

		// Empty prefixes slice
		{name: "empty prefixes slice", ticketID: "PO-1", prefixes: nil, expected: "", expectedOk: false},

		// Custom prefixes configuration
		{
			name:     "custom DOC prefix to docs",
			ticketID: "DOC-101",
			prefixes: []domain.TicketPrefix{
				{Prefix: "DOC", TicketType: "docs"},
			},
			expected:   domain.TicketTypeDocs,
			expectedOk: true,
		},
		{
			name:     "custom TASK prefix to chore",
			ticketID: "TASK-42",
			prefixes: []domain.TicketPrefix{
				{Prefix: "TASK", TicketType: "chore"},
			},
			expected:   domain.TicketTypeChore,
			expectedOk: true,
		},
		{
			name:     "custom BUG prefix to fix",
			ticketID: "BUG-99",
			prefixes: []domain.TicketPrefix{
				{Prefix: "BUG", TicketType: "fix"},
			},
			expected:   domain.TicketTypeFix,
			expectedOk: true,
		},
		{
			name:     "custom FEAT prefix with alias feature",
			ticketID: "FEAT-7",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FEAT", TicketType: "feature"},
			},
			expected:   domain.TicketTypeFeat,
			expectedOk: true,
		},
		{
			name:     "custom prefix with trailing hyphen in config",
			ticketID: "SPEC-5",
			prefixes: []domain.TicketPrefix{
				{Prefix: "SPEC-", TicketType: "docs"},
			},
			expected:   domain.TicketTypeDocs,
			expectedOk: true,
		},
		{
			name:     "custom prefix with trailing underscore in config",
			ticketID: "SPEC_5",
			prefixes: []domain.TicketPrefix{
				{Prefix: "SPEC_", TicketType: "docs"},
			},
			expected:   domain.TicketTypeDocs,
			expectedOk: true,
		},
		{
			name:     "custom prefix with invalid ticket type",
			ticketID: "INVALID-1",
			prefixes: []domain.TicketPrefix{
				{Prefix: "INVALID", TicketType: "nonexistent-type"},
			},
			expected:   "",
			expectedOk: false,
		},
		{
			name:     "empty prefix in config is skipped",
			ticketID: "FOO-1",
			prefixes: []domain.TicketPrefix{
				{Prefix: "", TicketType: "feat"},
				{Prefix: "   ", TicketType: "fix"},
				{Prefix: "---", TicketType: "docs"},
				{Prefix: "___", TicketType: "chore"},
			},
			expected:   "",
			expectedOk: false,
		},
		{
			name:     "delimiter test FOO_1 with underscore delimiter",
			ticketID: "FOO_1",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   domain.TicketTypeFeat,
			expectedOk: true,
		},
		{
			name:     "delimiter test FOO-1 with hyphen delimiter",
			ticketID: "FOO-1",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   domain.TicketTypeFeat,
			expectedOk: true,
		},
		{
			name:     "delimiter test FOOBAR-1 does NOT match prefix FOO",
			ticketID: "FOOBAR-1",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   "",
			expectedOk: false,
		},
		{
			name:     "non-numeric suffix FOO-BAR matches prefix FOO",
			ticketID: "FOO-BAR",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   domain.TicketTypeFeat,
			expectedOk: true,
		},
		{
			name:     "non-numeric suffix FOO_BAR_EXTRA matches prefix FOO",
			ticketID: "FOO_BAR_EXTRA",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   domain.TicketTypeFeat,
			expectedOk: true,
		},
		{
			name:     "bare FOO without delimiter returns false",
			ticketID: "FOO",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   "",
			expectedOk: false,
		},
		{
			name:     "FOO- with empty suffix returns false",
			ticketID: "FOO-",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   "",
			expectedOk: false,
		},
		{
			name:     "FOO_ with empty suffix returns false",
			ticketID: "FOO_",
			prefixes: []domain.TicketPrefix{
				{Prefix: "FOO", TicketType: "feat"},
			},
			expected:   "",
			expectedOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual, ok := domain.InferTicketTypeFromPrefix(tc.ticketID, tc.prefixes)
			if ok != tc.expectedOk {
				t.Fatalf("InferTicketTypeFromPrefix(%q) ok = %v, expected %v", tc.ticketID, ok, tc.expectedOk)
			}
			if actual != tc.expected {
				t.Errorf("InferTicketTypeFromPrefix(%q) = %q, expected %q", tc.ticketID, actual, tc.expected)
			}
		})
	}
}

func TestTicketConfig_InferTicketType(t *testing.T) {
	t.Parallel()

	cfg := domain.TicketConfig{
		Prefixes: []domain.TicketPrefix{
			{Prefix: "PO", TicketType: "docs"},
			{Prefix: "GEN", TicketType: "chore"},
			{Prefix: "HOTFIX", TicketType: "fix"},
		},
	}

	got, ok := cfg.InferTicketType("PO-42")
	if !ok || got != domain.TicketTypeDocs {
		t.Errorf("expected (docs, true), got (%q, %v)", got, ok)
	}

	got, ok = cfg.InferTicketType("gen-100")
	if !ok || got != domain.TicketTypeChore {
		t.Errorf("expected (chore, true), got (%q, %v)", got, ok)
	}

	got, ok = cfg.InferTicketType("HOTFIX_7")
	if !ok || got != domain.TicketTypeFix {
		t.Errorf("expected (fix, true), got (%q, %v)", got, ok)
	}

	got, ok = cfg.InferTicketType("UNKNOWN-1")
	if ok || got != "" {
		t.Errorf("expected ('', false), got (%q, %v)", got, ok)
	}
}

func TestRoleMappings_ResolveRole_AdversarialUnmappedAndWhitespace(t *testing.T) {
	t.Parallel()

	t.Run("arbitrary unmapped roles resolve to lowercase cleaned AgentRole", func(t *testing.T) {
		t.Parallel()
		defaultMappings := domain.DefaultProjectConfig().RoleMappings

		tests := []struct {
			input    string
			expected domain.AgentRole
		}{
			{"developer", domain.AgentRole("developer")},
			{"DEVELOPER", domain.AgentRole("developer")},
			{"  developer  ", domain.AgentRole("developer")},
			{"DeVeLoPeR", domain.AgentRole("developer")},
			{"\tdeveloper\n", domain.AgentRole("developer")},
			{"researcher", domain.AgentRole("researcher")},
			{"RESEARCHER", domain.AgentRole("researcher")},
			{"  researcher  ", domain.AgentRole("researcher")},
			{"ReSeArChEr", domain.AgentRole("researcher")},
			{"architect", domain.AgentRole("architect")},
			{"ARCHITECT", domain.AgentRole("architect")},
			{"  architect  ", domain.AgentRole("architect")},
			{"ArChItEcT", domain.AgentRole("architect")},
			{"lead-qa", domain.AgentRole("lead-qa")},
			{"LEAD-QA", domain.AgentRole("lead-qa")},
			{"Lead-QA", domain.AgentRole("lead-qa")},
			{"  lead-qa  ", domain.AgentRole("lead-qa")},
			{"\tLead-QA\r\n", domain.AgentRole("lead-qa")},
			{"security-engineer", domain.AgentRole("security-engineer")},
			{"data-analyst", domain.AgentRole("data-analyst")},
			{"devops", domain.AgentRole("devops")},
			{"sre", domain.AgentRole("sre")},
			{"  SRE  ", domain.AgentRole("sre")},
			{"product-manager", domain.AgentRole("product-manager")},
		}

		for _, tc := range tests {
			got := defaultMappings.ResolveRole(tc.input)
			if got != tc.expected {
				t.Errorf("defaultMappings.ResolveRole(%q) = %q, expected %q", tc.input, got, tc.expected)
			}
		}
	})

	t.Run("empty and whitespace variations across mappings instances resolve to RoleUnassigned", func(t *testing.T) {
		t.Parallel()

		whitespaceInputs := []string{
			"",
			" ",
			"  ",
			"   ",
			"\t",
			"\n",
			"\r\n",
			"\t\t",
			"  \t  \n  \r  ",
		}

		mappingInstances := []struct {
			name     string
			mappings domain.RoleMappings
		}{
			{"default mappings", domain.DefaultProjectConfig().RoleMappings},
			{"nil mappings", nil},
			{"empty mappings", domain.RoleMappings{}},
			{"custom mappings", domain.RoleMappings{"custom": "coder"}},
		}

		for _, mi := range mappingInstances {
			for _, input := range whitespaceInputs {
				got := mi.mappings.ResolveRole(input)
				if got != domain.RoleUnassigned {
					t.Errorf("%s: ResolveRole(%q) = %q, expected RoleUnassigned (%q)", mi.name, input, got, domain.RoleUnassigned)
				}
			}
		}
	})

	t.Run("nil and empty mappings return unmapped cleaned roles", func(t *testing.T) {
		t.Parallel()

		unmappedRoles := []string{"developer", "researcher", "architect", "lead-qa"}
		var nilMap domain.RoleMappings
		emptyMap := domain.RoleMappings{}

		for _, role := range unmappedRoles {
			if got := nilMap.ResolveRole(role); got != domain.AgentRole(role) {
				t.Errorf("nilMap.ResolveRole(%q) = %q, expected %q", role, got, role)
			}
			if got := emptyMap.ResolveRole(role); got != domain.AgentRole(role) {
				t.Errorf("emptyMap.ResolveRole(%q) = %q, expected %q", role, got, role)
			}
		}
	})

	t.Run("custom mappings partial override with unmapped fallback and key normalization", func(t *testing.T) {
		t.Parallel()

		custom := domain.RoleMappings{
			"Architect":    "planner",
			"  Lead-QA  ":  "tester",
			"reviewer-bot": "  REVIEWER  ",
			"empty-target": "   ",
		}

		if got := custom.ResolveRole("architect"); got != domain.RolePlanner {
			t.Errorf("expected RolePlanner for architect, got %q", got)
		}
		if got := custom.ResolveRole("ARCHITECT"); got != domain.RolePlanner {
			t.Errorf("expected RolePlanner for ARCHITECT, got %q", got)
		}
		if got := custom.ResolveRole("  architect  "); got != domain.RolePlanner {
			t.Errorf("expected RolePlanner for '  architect  ', got %q", got)
		}
		if got := custom.ResolveRole("lead-qa"); got != domain.RoleTester {
			t.Errorf("expected RoleTester for lead-qa, got %q", got)
		}
		if got := custom.ResolveRole("LEAD-QA"); got != domain.RoleTester {
			t.Errorf("expected RoleTester for LEAD-QA, got %q", got)
		}
		if got := custom.ResolveRole("reviewer-bot"); got != domain.RoleReviewer {
			t.Errorf("expected RoleReviewer for reviewer-bot, got %q", got)
		}

		if got := custom.ResolveRole("empty-target"); got != domain.AgentRole("empty-target") {
			t.Errorf("expected empty-target fallback, got %q", got)
		}

		if got := custom.ResolveRole("developer"); got != domain.AgentRole("developer") {
			t.Errorf("expected developer, got %q", got)
		}
		if got := custom.ResolveRole("  DEVELOPER  "); got != domain.AgentRole("developer") {
			t.Errorf("expected developer for '  DEVELOPER  ', got %q", got)
		}
		if got := custom.ResolveRole("researcher"); got != domain.AgentRole("researcher") {
			t.Errorf("expected researcher, got %q", got)
		}
		if got := custom.ResolveRole("ReSeArChEr"); got != domain.AgentRole("researcher") {
			t.Errorf("expected researcher for ReSeArChEr, got %q", got)
		}
	})
}
