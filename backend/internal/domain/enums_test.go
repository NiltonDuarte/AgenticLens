package domain_test

import (
	"testing"

	"agenticlens/internal/domain"
)

func TestCanonicalTicketTypes(t *testing.T) {
	t.Parallel()

	expected := []domain.ConventionalTicketType{
		domain.TicketTypeFeat,
		domain.TicketTypeFix,
		domain.TicketTypeDocs,
		domain.TicketTypeChore,
		domain.TicketTypeRefactor,
		domain.TicketTypeTest,
		domain.TicketTypeCI,
		domain.TicketTypePerf,
		domain.TicketTypeBuild,
	}

	if len(domain.CanonicalTicketTypes) != len(expected) {
		t.Fatalf("expected %d canonical ticket types, got %d", len(expected), len(domain.CanonicalTicketTypes))
	}

	for i, exp := range expected {
		if domain.CanonicalTicketTypes[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, domain.CanonicalTicketTypes[i])
		}
	}
}

func TestNormalizeTicketType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		expected   domain.ConventionalTicketType
		expectedOk bool
	}{
		// Canonical types
		{name: "canonical feat", input: "feat", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "canonical fix", input: "fix", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "canonical docs", input: "docs", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "canonical chore", input: "chore", expected: domain.TicketTypeChore, expectedOk: true},
		{name: "canonical refactor", input: "refactor", expected: domain.TicketTypeRefactor, expectedOk: true},
		{name: "canonical test", input: "test", expected: domain.TicketTypeTest, expectedOk: true},
		{name: "canonical ci", input: "ci", expected: domain.TicketTypeCI, expectedOk: true},
		{name: "canonical perf", input: "perf", expected: domain.TicketTypePerf, expectedOk: true},
		{name: "canonical build", input: "build", expected: domain.TicketTypeBuild, expectedOk: true},

		// Aliases
		{name: "alias feature", input: "feature", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "alias bug", input: "bug", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "alias bugfix", input: "bugfix", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "alias doc", input: "doc", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "alias documentation", input: "documentation", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "alias task", input: "task", expected: domain.TicketTypeChore, expectedOk: true},
		{name: "alias tests", input: "tests", expected: domain.TicketTypeTest, expectedOk: true},
		{name: "alias performance", input: "performance", expected: domain.TicketTypePerf, expectedOk: true},

		// Whitespace and case insensitivity
		{name: "uppercase FEAT", input: "FEAT", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "mixed case Feature with whitespace", input: "  Feature  ", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "mixed case FeAt with whitespace", input: "  FeAt  ", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "uppercase BUGFIX", input: "BUGFIX", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "mixed case BuGfIx", input: "\tBuGfIx\n", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "mixed case DoCs", input: "DoCs", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "uppercase DOCS", input: "DOCS", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "uppercase Documentation", input: "DoCuMeNtAtIoN", expected: domain.TicketTypeDocs, expectedOk: true},
		{name: "mixed case TaSk", input: "  TaSk  ", expected: domain.TicketTypeChore, expectedOk: true},
		{name: "uppercase CI with whitespace and control chars", input: "\n\r\t  cI  \t", expected: domain.TicketTypeCI, expectedOk: true},
		{name: "mixed case PeRf with whitespace", input: "\tPeRf\n", expected: domain.TicketTypePerf, expectedOk: true},
		{name: "mixed case PeRfOrMaNcE", input: "PeRfOrMaNcE", expected: domain.TicketTypePerf, expectedOk: true},
		{name: "mixed case bUiLd with whitespace", input: "  bUiLd  ", expected: domain.TicketTypeBuild, expectedOk: true},
		{name: "mixed case ReFaCtOr with whitespace and newlines", input: "\n\r  ReFaCtOr \t", expected: domain.TicketTypeRefactor, expectedOk: true},
		{name: "mixed case bUg", input: "bUg", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "mixed case tEsTs with whitespace", input: "  tEsTs \t", expected: domain.TicketTypeTest, expectedOk: true},
		{name: "mixed case fEaTuRe with padding", input: "   fEaTuRe   ", expected: domain.TicketTypeFeat, expectedOk: true},
		{name: "carriage return fix", input: "\r\n\t  fIx  \r\n", expected: domain.TicketTypeFix, expectedOk: true},
		{name: "uppercase BUILD", input: "BUILD", expected: domain.TicketTypeBuild, expectedOk: true},

		// Invalid values
		{name: "empty string", input: "", expected: "", expectedOk: false},
		{name: "whitespace only", input: "   \t\n", expected: "", expectedOk: false},
		{name: "unknown type", input: "unknown", expected: "", expectedOk: false},
		{name: "epic", input: "epic", expected: "", expectedOk: false},
		{name: "story", input: "story", expected: "", expectedOk: false},
		{name: "spike", input: "spike", expected: "", expectedOk: false},
		{name: "numeric string", input: "123", expected: "", expectedOk: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual, ok := domain.NormalizeTicketType(tc.input)
			if ok != tc.expectedOk {
				t.Fatalf("NormalizeTicketType(%q) ok = %v, expected %v", tc.input, ok, tc.expectedOk)
			}
			if actual != tc.expected {
				t.Errorf("NormalizeTicketType(%q) = %q, expected %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestIsValidTicketType(t *testing.T) {
	t.Parallel()

	validCases := []string{
		"feat", "feature", "fix", "bug", "bugfix", "docs", "doc", "documentation",
		"chore", "task", "refactor", "test", "tests", "ci", "perf", "performance", "build",
		"  feat  ", "FEAT", "Chore", "  FeAt  ", "BUGFIX", "DoCs", "\n\r\t  cI  \t",
		"PeRfOrMaNcE", "DoCuMeNtAtIoN", "\tPeRf\n", "  bUiLd  ",
	}

	for _, input := range validCases {
		if !domain.IsValidTicketType(input) {
			t.Errorf("expected IsValidTicketType(%q) to be true", input)
		}
	}

	invalidCases := []string{
		"", "   ", "unknown", "epic", "story", "bug-fix", "feats", "release",
		"feat-1", "docs-spec", "chore/subsystem", "fix123", "   \t\n\r   ",
	}

	for _, input := range invalidCases {
		if domain.IsValidTicketType(input) {
			t.Errorf("expected IsValidTicketType(%q) to be false", input)
		}
	}
}

func TestOrderSubagentRoles_OrderingAndPreservation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       []string
		customOrder []string
		expected    []string
	}{
		{
			name:        "empty input slice",
			input:       []string{},
			customOrder: nil,
			expected:    nil,
		},
		{
			name:        "slice with only whitespace and empty strings",
			input:       []string{"", "   ", "\t\n"},
			customOrder: nil,
			expected:    nil,
		},
		{
			name:        "alphabetical order when no custom order provided",
			input:       []string{"reviewer", "tester", "coder", "planner"},
			customOrder: nil,
			expected:    []string{"coder", "planner", "reviewer", "tester"},
		},
		{
			name:        "custom order priority with alphabetical remainder",
			input:       []string{"reviewer", "tester", "coder", "planner", "zeta", "alpha"},
			customOrder: []string{"planner", "coder", "tester", "reviewer"},
			expected:    []string{"planner", "coder", "tester", "reviewer", "alpha", "zeta"},
		},
		{
			name:        "custom order with duplicates and whitespace",
			input:       []string{" coder ", "planner", "coder", " tester\t", "reviewer", "planner"},
			customOrder: []string{"planner", "coder", "tester", "reviewer"},
			expected:    []string{"planner", "coder", "tester", "reviewer"},
		},
		{
			name:        "mixed custom order and remaining roles sorted alphabetically",
			input:       []string{"advisor", "reviewer", "orchestrator", "general", "build", "root", "po", "planner", "advisor", "General", "BUILD"},
			customOrder: []string{"planner", "reviewer"},
			expected:    []string{"planner", "reviewer", "BUILD", "General", "advisor", "build", "general", "orchestrator", "po", "root"},
		},
		{
			name:        "purely custom roles sorted alphabetically when no custom order",
			input:       []string{"zebra", "alpha", "beta"},
			customOrder: nil,
			expected:    []string{"alpha", "beta", "zebra"},
		},
		{
			name:        "unassigned role included among roles",
			input:       []string{"unassigned", "planner", "coder"},
			customOrder: nil,
			expected:    []string{"coder", "planner", "unassigned"},
		},
		{
			name:        "nil input",
			input:       nil,
			customOrder: nil,
			expected:    nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			if tc.customOrder != nil {
				got = domain.OrderSubagentRoles(tc.input, tc.customOrder)
			} else {
				got = domain.OrderSubagentRoles(tc.input)
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("expected len %d, got %d: %v", len(tc.expected), len(got), got)
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Errorf("at index %d: expected %s, got %s", i, tc.expected[i], got[i])
				}
			}
		})
	}
}

func TestOrderRoles_OrderingAndPreservation(t *testing.T) {
	t.Parallel()

	input := []string{"tester", "coder", "orchestrator", "custom", "reviewer", "planner"}
	customOrder := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}

	gotCustom := domain.OrderRoles(input, customOrder)
	expectedCustom := []string{"orchestrator", "planner", "coder", "tester", "reviewer", "custom"}
	if len(gotCustom) != len(expectedCustom) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedCustom), len(gotCustom), gotCustom)
	}
	for i := range expectedCustom {
		if gotCustom[i] != expectedCustom[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedCustom[i], gotCustom[i])
		}
	}

	gotAlpha := domain.OrderRoles(input)
	expectedAlpha := []string{"coder", "custom", "orchestrator", "planner", "reviewer", "tester"}
	if len(gotAlpha) != len(expectedAlpha) {
		t.Fatalf("expected len %d, got %d: %v", len(expectedAlpha), len(gotAlpha), gotAlpha)
	}
	for i := range expectedAlpha {
		if gotAlpha[i] != expectedAlpha[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expectedAlpha[i], gotAlpha[i])
		}
	}
}

func TestOrderRoles_And_OrderSubagentRoles_AdversarialComprehensive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		input       []string
		customOrder [][]string
		expected    []string
	}

	testCases := []testCase{
		{
			name:        "nil input returns nil",
			input:       nil,
			customOrder: nil,
			expected:    nil,
		},
		{
			name:        "empty input slice returns nil",
			input:       []string{},
			customOrder: nil,
			expected:    nil,
		},
		{
			name:        "whitespace-only input elements return nil",
			input:       []string{"", "  ", "\t", "\n", "\r\n"},
			customOrder: nil,
			expected:    nil,
		},
		{
			name:        "omitted customOrder sorts strictly alphabetically",
			input:       []string{"developer", "architect", "researcher", "lead-qa"},
			customOrder: nil,
			expected:    []string{"architect", "developer", "lead-qa", "researcher"},
		},
		{
			name:        "nil customOrder slice sorts strictly alphabetically",
			input:       []string{"developer", "architect", "researcher", "lead-qa"},
			customOrder: [][]string{nil},
			expected:    []string{"architect", "developer", "lead-qa", "researcher"},
		},
		{
			name:        "empty customOrder slice sorts strictly alphabetically",
			input:       []string{"developer", "architect", "researcher", "lead-qa"},
			customOrder: [][]string{{}},
			expected:    []string{"architect", "developer", "lead-qa", "researcher"},
		},
		{
			name:        "whitespace-only customOrder slice sorts strictly alphabetically",
			input:       []string{"developer", "architect", "researcher", "lead-qa"},
			customOrder: [][]string{{"", "  ", "\t"}},
			expected:    []string{"architect", "developer", "lead-qa", "researcher"},
		},
		{
			name:        "input with duplicate roles and whitespace trimmed and deduplicated alphabetically",
			input:       []string{" developer ", "architect", "DEVELOPER", "  developer  ", "architect\t", "lead-qa"},
			customOrder: nil,
			expected:    []string{"DEVELOPER", "architect", "developer", "lead-qa"},
		},
		{
			name:        "customOrder priority first followed by alphabetical remainder",
			input:       []string{"developer", "architect", "researcher", "lead-qa", "tester", "coder"},
			customOrder: [][]string{{"lead-qa", "developer"}},
			expected:    []string{"lead-qa", "developer", "architect", "coder", "researcher", "tester"},
		},
		{
			name:        "customOrder with unlisted roles not present in input are cleanly ignored",
			input:       []string{"developer", "researcher"},
			customOrder: [][]string{{"missing-1", "researcher", "missing-2", "developer", "missing-3"}},
			expected:    []string{"researcher", "developer"},
		},
		{
			name:        "customOrder with duplicate roles in custom order deduplicated without repeat",
			input:       []string{"developer", "researcher", "architect"},
			customOrder: [][]string{{"researcher", "researcher", "architect", "researcher"}},
			expected:    []string{"researcher", "architect", "developer"},
		},
		{
			name:        "customOrder matching zero input roles falls back to full alphabetical sort",
			input:       []string{"developer", "researcher", "architect"},
			customOrder: [][]string{{"planner", "coder", "tester", "reviewer"}},
			expected:    []string{"architect", "developer", "researcher"},
		},
		{
			name:        "customOrder matching all input roles in reverse order preserves exact customOrder",
			input:       []string{"architect", "developer", "lead-qa", "researcher"},
			customOrder: [][]string{{"researcher", "lead-qa", "developer", "architect"}},
			expected:    []string{"researcher", "lead-qa", "developer", "architect"},
		},
		{
			name:        "customOrder with whitespace trimmed around role names",
			input:       []string{"developer", "researcher", "architect"},
			customOrder: [][]string{{"  researcher  ", "\tarchitect\n"}},
			expected:    []string{"researcher", "architect", "developer"},
		},
		{
			name:        "already sorted input preserves alphabetical order",
			input:       []string{"alpha", "beta", "gamma", "omega"},
			customOrder: nil,
			expected:    []string{"alpha", "beta", "gamma", "omega"},
		},
		{
			name:        "reverse sorted input sorts ascending alphabetically",
			input:       []string{"omega", "gamma", "beta", "alpha"},
			customOrder: nil,
			expected:    []string{"alpha", "beta", "gamma", "omega"},
		},
		{
			name:        "single role input returned as single element",
			input:       []string{"developer"},
			customOrder: nil,
			expected:    []string{"developer"},
		},
		{
			name:        "single role input with customOrder matching",
			input:       []string{"developer"},
			customOrder: [][]string{{"developer"}},
			expected:    []string{"developer"},
		},
		{
			name:        "single role input with customOrder not matching",
			input:       []string{"developer"},
			customOrder: [][]string{{"architect"}},
			expected:    []string{"developer"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("OrderRoles/"+tc.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			if tc.customOrder != nil {
				got = domain.OrderRoles(tc.input, tc.customOrder...)
			} else {
				got = domain.OrderRoles(tc.input)
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("OrderRoles len mismatch: expected %d (%v), got %d (%v)", len(tc.expected), tc.expected, len(got), got)
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Errorf("OrderRoles at [%d]: expected %q, got %q", i, tc.expected[i], got[i])
				}
			}
		})

		t.Run("OrderSubagentRoles/"+tc.name, func(t *testing.T) {
			t.Parallel()
			var got []string
			if tc.customOrder != nil {
				got = domain.OrderSubagentRoles(tc.input, tc.customOrder...)
			} else {
				got = domain.OrderSubagentRoles(tc.input)
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("OrderSubagentRoles len mismatch: expected %d (%v), got %d (%v)", len(tc.expected), tc.expected, len(got), got)
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Errorf("OrderSubagentRoles at [%d]: expected %q, got %q", i, tc.expected[i], got[i])
				}
			}
		})
	}
}

func TestOrderRoles_GenericAgentRoleType(t *testing.T) {
	t.Parallel()

	input := []domain.AgentRole{
		domain.AgentRole("tester"),
		domain.AgentRole("architect"),
		domain.AgentRole("coder"),
		domain.AgentRole("security-auditor"),
		domain.AgentRole("planner"),
		domain.AgentRole("orchestrator"),
	}

	customOrder := []domain.AgentRole{
		domain.AgentRole("orchestrator"),
		domain.AgentRole("planner"),
		domain.AgentRole("coder"),
		domain.AgentRole("tester"),
	}

	ordered := domain.OrderRoles(input, customOrder)
	expected := []domain.AgentRole{
		domain.AgentRole("orchestrator"),
		domain.AgentRole("planner"),
		domain.AgentRole("coder"),
		domain.AgentRole("tester"),
		domain.AgentRole("architect"),
		domain.AgentRole("security-auditor"),
	}

	if len(ordered) != len(expected) {
		t.Fatalf("expected len %d, got %d: %v", len(expected), len(ordered), ordered)
	}
	for i := range expected {
		if ordered[i] != expected[i] {
			t.Errorf("at index %d: expected %s, got %s", i, expected[i], ordered[i])
		}
	}

	// Subagent ordering with AgentRole
	subOrdered := domain.OrderSubagentRoles(input, customOrder)
	if len(subOrdered) != len(expected) {
		t.Fatalf("expected subagent len %d, got %d: %v", len(expected), len(subOrdered), subOrdered)
	}
	for i := range expected {
		if subOrdered[i] != expected[i] {
			t.Errorf("subagent at index %d: expected %s, got %s", i, expected[i], subOrdered[i])
		}
	}
}

func TestAgentRole_FallbackConstants(t *testing.T) {
	t.Parallel()

	if domain.RoleRootAgent != "RootAgent" {
		t.Errorf("expected RoleRootAgent to be 'RootAgent', got %q", domain.RoleRootAgent)
	}
	if domain.RoleUnassignedSubAgent != "UnassignedSubAgent" {
		t.Errorf("expected RoleUnassignedSubAgent to be 'UnassignedSubAgent', got %q", domain.RoleUnassignedSubAgent)
	}
}
