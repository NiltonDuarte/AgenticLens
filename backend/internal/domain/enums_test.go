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
		name        string
		input       string
		expected    domain.ConventionalTicketType
		expectedOk  bool
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
