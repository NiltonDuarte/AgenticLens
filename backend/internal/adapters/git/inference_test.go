package git_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func TestInferTicketType_PrefixMapping(t *testing.T) {
	t.Parallel()

	configuredPrefixes := []domain.TicketPrefix{
		{Prefix: "PO", TicketType: "docs"},
		{Prefix: "GEN", TicketType: "chore"},
	}

	tests := []struct {
		name     string
		ticketID string
		expected domain.ConventionalTicketType
	}{
		{name: "PO ticket uppercase", ticketID: "PO-5", expected: domain.TicketTypeDocs},
		{name: "po ticket lowercase", ticketID: "po-12", expected: domain.TicketTypeDocs},
		{name: "Po ticket mixed case", ticketID: "Po-99", expected: domain.TicketTypeDocs},
		{name: "GEN ticket uppercase", ticketID: "GEN-42", expected: domain.TicketTypeChore},
		{name: "gen ticket lowercase", ticketID: "gen-1", expected: domain.TicketTypeChore},
		{name: "Gen ticket mixed case", ticketID: "Gen-100", expected: domain.TicketTypeChore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := git.InferTicketType(tc.ticketID, configuredPrefixes...)
			if actual == nil {
				t.Fatalf("expected non-nil ticket type for %s, got nil", tc.ticketID)
			}
			if *actual != tc.expected {
				t.Errorf("InferTicketType(%q) = %q, expected %q", tc.ticketID, *actual, tc.expected)
			}
		})
	}

	t.Run("without prefixes returns nil on unbranched PO and GEN tickets", func(t *testing.T) {
		t.Parallel()
		for _, tid := range []string{"PO-5", "GEN-42", "PO-100", "GEN-1"} {
			res := git.InferTicketType(tid)
			if res != nil {
				t.Errorf("expected nil for %s without prefixes, got %q", tid, *res)
			}
		}
	})

	t.Run("custom prefix resolves to configured type", func(t *testing.T) {
		t.Parallel()
		customPrefixes := []domain.TicketPrefix{
			{Prefix: "CUSTOM", TicketType: "fix"},
			{Prefix: "TASK", TicketType: "chore"},
		}
		actual := git.InferTicketType("CUSTOM-1", customPrefixes...)
		if actual == nil || *actual != "fix" {
			t.Errorf("expected fix for CUSTOM-1, got %v", actual)
		}

		actualTask := git.InferTicketType("TASK-99", customPrefixes...)
		if actualTask == nil || *actualTask != "chore" {
			t.Errorf("expected chore for TASK-99, got %v", actualTask)
		}
	})
}

func TestInferTicketType_EmptyAndWhitespace(t *testing.T) {
	t.Parallel()

	testCases := []string{"", "   ", "\t\n"}
	for _, tc := range testCases {
		res := git.InferTicketType(tc)
		if res != nil {
			t.Errorf("InferTicketType(%q) = %v, expected nil", tc, *res)
		}
	}
}

func TestInferTicketType_UnmatchedReturnsNil(t *testing.T) {
	t.Parallel()

	// An unmatched ticket without a matching branch or PO/GEN prefix must return nil,
	// never falling back to "feature".
	unmatched := []string{
		"NONEXISTENT-TICKET-99999",
		"UNKNOWN-XYZ",
		"RANDOM-404",
	}

	for _, ticketID := range unmatched {
		res := git.InferTicketType(ticketID)
		if res != nil {
			t.Errorf("expected nil for unmatched ticket %s, got %q", ticketID, *res)
		}
	}
}

func TestInferTicketType_BranchPrefixResolution(t *testing.T) {
	t.Parallel()
	branchTests := []struct {
		ticketID   string
		branchList string
		expected   domain.ConventionalTicketType
	}{
		// Canonical types
		{ticketID: "AL-1", branchList: "  feat/AL-1-ci-fixture\n", expected: domain.TicketTypeFeat},
		{ticketID: "AL-30", branchList: "+ fix/AL-30-ci-fixture\n", expected: domain.TicketTypeFix},
		{ticketID: "AL-40", branchList: "  docs/AL-40-architecture\n", expected: domain.TicketTypeDocs},
		{ticketID: "AL-50", branchList: "  chore/AL-50-cleanup\n", expected: domain.TicketTypeChore},
		{ticketID: "AL-16", branchList: "* refactor/AL-16-ci-fixture\n", expected: domain.TicketTypeRefactor},
		{ticketID: "AL-60", branchList: "  test/AL-60-e2e\n", expected: domain.TicketTypeTest},
		{ticketID: "AL-14", branchList: "  remotes/origin/ci/AL-14-ci-fixture\n", expected: domain.TicketTypeCI},
		{ticketID: "AL-70", branchList: "  perf/AL-70-cache-opt\n", expected: domain.TicketTypePerf},
		{ticketID: "AL-80", branchList: "  build/AL-80-goreleaser\n", expected: domain.TicketTypeBuild},

		// Aliases
		{ticketID: "AL-90", branchList: "  feature/AL-90-ui-dashboard\n", expected: domain.TicketTypeFeat},
		{ticketID: "AL-91", branchList: "  bug/AL-91-segfault\n", expected: domain.TicketTypeFix},
		{ticketID: "AL-92", branchList: "  bugfix/AL-92-nil-pointer\n", expected: domain.TicketTypeFix},
		{ticketID: "AL-93", branchList: "  remotes/upstream/doc/AL-93-guide\n", expected: domain.TicketTypeDocs},
	}

	for _, tc := range branchTests {
		res := git.InferTicketTypeFromBranchList(tc.branchList, tc.ticketID)
		if res == nil {
			t.Fatalf("expected non-nil ticket type for branch ticket %s, got nil", tc.ticketID)
		}
		if *res != tc.expected {
			t.Errorf("InferTicketTypeFromBranchList(%q) = %q, expected %q", tc.ticketID, *res, tc.expected)
		}
	}

	// Boundary and edge cases
	t.Run("empty ticketID returns nil", func(t *testing.T) {
		t.Parallel()
		if res := git.InferTicketTypeFromBranchList("  feat/AL-1-feature\n", ""); res != nil {
			t.Errorf("expected nil for empty ticketID, got %v", *res)
		}
		if res := git.InferTicketTypeFromBranchList("  feat/AL-1-feature\n", "   "); res != nil {
			t.Errorf("expected nil for whitespace ticketID, got %v", *res)
		}
	})

	t.Run("empty branch output returns nil", func(t *testing.T) {
		t.Parallel()
		if res := git.InferTicketTypeFromBranchList("", "AL-1"); res != nil {
			t.Errorf("expected nil for empty branch output, got %v", *res)
		}
	})

	t.Run("unmatched ticket branch returns nil", func(t *testing.T) {
		t.Parallel()
		if res := git.InferTicketTypeFromBranchList("  feat/AL-99-other\n", "AL-1"); res != nil {
			t.Errorf("expected nil for unmatched ticket, got %v", *res)
		}
	})

	t.Run("branch without conventional prefix returns nil", func(t *testing.T) {
		t.Parallel()
		if res := git.InferTicketTypeFromBranchList("  random/AL-1-feature\n", "AL-1"); res != nil {
			t.Errorf("expected nil for non-conventional prefix, got %v", *res)
		}
	})
}

func TestInferTicketType_Precedence_BranchOverConfiguredPrefix(t *testing.T) {
	t.Parallel()

	// In-memory branch fixture evaluates conventional commit type directly
	branchOut := "  fix/AL-30-ci-fixture\n"
	branchType := git.InferTicketTypeFromBranchList(branchOut, "AL-30")
	if branchType == nil || *branchType != domain.TicketTypeFix {
		t.Fatalf("expected in-memory branch resolution 'fix', got %v", branchType)
	}

	// Configured prefixes are prioritized over subprocess execution in InferTicketType
	configuredPrefixes := []domain.TicketPrefix{
		{Prefix: "AL", TicketType: "chore"},
	}
	actual := git.InferTicketType("AL-30", configuredPrefixes...)
	if actual == nil || *actual != domain.TicketTypeChore {
		t.Errorf("expected configured prefix 'chore' to take priority in InferTicketType, got %v", actual)
	}

	// Configured prefixes resolve unbranched PO and custom tickets
	poActual := git.InferTicketType("PO-5", domain.TicketPrefix{Prefix: "PO", TicketType: "docs"})
	if poActual == nil || *poActual != domain.TicketTypeDocs {
		t.Errorf("expected configured prefix 'docs' for unbranched PO-5, got %v", poActual)
	}

	unbranchedActual := git.InferTicketType("CUSTOM-99", domain.TicketPrefix{Prefix: "CUSTOM", TicketType: "refactor"})
	if unbranchedActual == nil || *unbranchedActual != domain.TicketTypeRefactor {
		t.Errorf("expected configured prefix 'refactor' for unbranched CUSTOM-99, got %v", unbranchedActual)
	}

	// Unbranched and unconfigured ticket returns nil
	unresolved := git.InferTicketType("CUSTOM-99")
	if unresolved != nil {
		t.Errorf("expected nil for unbranched and unconfigured ticket, got %v", unresolved)
	}
}

func TestInferTicketTypeFromBranchList_Adversarial(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		branchOut  string
		ticketID   string
		expected   *domain.ConventionalTicketType
	}{
		// Empty and whitespace edge cases
		{
			name:      "empty branch output",
			branchOut: "",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "whitespace only branch output",
			branchOut: "   \n\t  \r\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "empty ticket ID",
			branchOut: "  feat/AL-29\n",
			ticketID:  "",
			expected:  nil,
		},
		{
			name:      "whitespace ticket ID",
			branchOut: "  feat/AL-29\n",
			ticketID:  "   \t",
			expected:  nil,
		},

		// Unconventional branch naming: no prefix
		{
			name:      "branch without slash prefix",
			branchOut: "  AL-29\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "active branch without slash prefix",
			branchOut: "* AL-29\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "main and master branches only",
			branchOut: "  main\n* master\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "branch starting with slash",
			branchOut: "  /AL-29\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "unrecognized prefix without conventional mapping",
			branchOut: "  experimental/AL-29\n  user-branch/AL-29\n",
			ticketID:  "AL-29",
			expected:  nil,
		},

		// Nested slashes
		{
			name:      "nested slashes with canonical feat",
			branchOut: "  feat/subsystem/core/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
		{
			name:      "nested slashes with canonical fix",
			branchOut: "  fix/deeply/nested/subsystem/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("fix"),
		},
		{
			name:      "nested slashes with alias documentation",
			branchOut: "  documentation/specs/v2/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("docs"),
		},
		{
			name:      "nested slashes with alias task",
			branchOut: "  task/infra/deps/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("chore"),
		},

		// Remote branches and nested slashes
		{
			name:      "remotes origin with feat and nested slashes",
			branchOut: "  remotes/origin/feat/subsystem/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
		{
			name:      "origin without remotes prefix",
			branchOut: "  origin/fix/auth/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("fix"),
		},
		{
			name:      "remotes upstream with refactor",
			branchOut: "  remotes/upstream/refactor/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("refactor"),
		},
		{
			name:      "upstream without remotes prefix",
			branchOut: "  upstream/ci/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("ci"),
		},
		{
			name:      "multiple slashes stripped cleanly",
			branchOut: "  remotes/origin///feat//subsystem/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},

		// Detached HEAD simulations
		{
			name:      "detached HEAD at commit hash",
			branchOut: "* (HEAD detached at 7b8a9c0)\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "detached HEAD mentioning ticket",
			branchOut: "* (HEAD detached from origin/feat/AL-29)\n",
			ticketID:  "AL-29",
			expected:  nil,
		},
		{
			name:      "detached HEAD with clean branch following",
			branchOut: "* (HEAD detached at 7b8a9c0)\n  feat/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},

		// Multiple ticket matches and ordering
		{
			name:      "first match unclassified, second match valid",
			branchOut: "  experimental/AL-29\n  feat/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
		{
			name:      "first valid match takes precedence",
			branchOut: "  fix/AL-29\n  feat/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("fix"),
		},
		{
			name:      "longer ticket number does not falsely match shorter ticket",
			branchOut: "  feat/AL-290\n  fix/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("fix"),
		},
		{
			name:      "ticket with suffix matches delimiter boundary",
			branchOut: "  feat/AL-29-deep-dive\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
		{
			name:      "ticket with underscore suffix matches",
			branchOut: "  fix/AL-29_hotfix\n",
			ticketID:  "AL-29",
			expected:  strPtr("fix"),
		},

		// False prefix matches (POD-1 vs PO-1, GENTLE-1 vs GEN-1)
		{
			name:      "ticket PO-1 does NOT match branch for POD-1",
			branchOut: "  feat/POD-1\n",
			ticketID:  "PO-1",
			expected:  nil,
		},
		{
			name:      "ticket PO-1 does NOT match branch for PO-10",
			branchOut: "  feat/PO-10\n",
			ticketID:  "PO-1",
			expected:  nil,
		},
		{
			name:      "ticket PO-1 matches docs/PO-1-spec",
			branchOut: "  docs/PO-1-spec\n",
			ticketID:  "PO-1",
			expected:  strPtr("docs"),
		},
		{
			name:      "ticket GEN-1 does NOT match branch for GENTLE-1",
			branchOut: "  chore/GENTLE-1\n",
			ticketID:  "GEN-1",
			expected:  nil,
		},
		{
			name:      "ticket GEN-1 does NOT match branch for GEN-100",
			branchOut: "  chore/GEN-100\n",
			ticketID:  "GEN-1",
			expected:  nil,
		},
		{
			name:      "ticket GEN-1 matches chore/GEN-1-task",
			branchOut: "  chore/GEN-1-task\n",
			ticketID:  "GEN-1",
			expected:  strPtr("chore"),
		},
		{
			name:      "ticket POD-1 does NOT match branch for PO-1",
			branchOut: "  docs/PO-1\n",
			ticketID:  "POD-1",
			expected:  nil,
		},
		{
			name:      "ticket POD-1 matches its own branch feat/POD-1",
			branchOut: "  feat/POD-1\n",
			ticketID:  "POD-1",
			expected:  strPtr("feat"),
		},
		{
			name:      "ticket GENTLE-1 matches its own branch fix/GENTLE-1",
			branchOut: "  fix/GENTLE-1\n",
			ticketID:  "GENTLE-1",
			expected:  strPtr("fix"),
		},

		// Whitespace and branch prefix symbols
		{
			name:      "worktree branch marker plus symbol",
			branchOut: "+ feat/AL-29\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
		{
			name:      "carriage returns in branch list",
			branchOut: "  feat/AL-29\r\n  fix/AL-29\r\n",
			ticketID:  "AL-29",
			expected:  strPtr("feat"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual := git.InferTicketTypeFromBranchList(tc.branchOut, tc.ticketID)
			if tc.expected == nil {
				if actual != nil {
					t.Fatalf("expected nil, got %q", *actual)
				}
			} else {
				if actual == nil {
					t.Fatalf("expected %q, got nil", *tc.expected)
				}
				if *actual != *tc.expected {
					t.Errorf("expected %q, got %q", *tc.expected, *actual)
				}
			}
		})
	}
}

func TestInferTicketType_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	configuredPrefixes := []domain.TicketPrefix{
		{Prefix: "PO", TicketType: "docs"},
		{Prefix: "GEN", TicketType: "chore"},
	}

	t.Run("false prefix tickets without matching branches return nil", func(t *testing.T) {
		t.Parallel()
		falsePrefixes := []string{
			"POD-1",
			"GENTLE-1",
			"POST-99",
			"GENESIS-42",
			"POW-10",
			"GENERAL-5",
			"PO1",
			"GEN1",
			"POLICE-7",
		}
		for _, ticketID := range falsePrefixes {
			res := git.InferTicketType(ticketID, configuredPrefixes...)
			if res != nil {
				t.Errorf("expected nil for false prefix ticket %q, got %q", ticketID, *res)
			}
		}
	})

	t.Run("valid prefix tickets with whitespace and mixed casing resolve correctly", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			input    string
			expected domain.ConventionalTicketType
		}{
			{input: "  PO-1  ", expected: domain.TicketTypeDocs},
			{input: "\tpO-42\n", expected: domain.TicketTypeDocs},
			{input: "\r\nPo-999\t", expected: domain.TicketTypeDocs},
			{input: "  GEN-1  ", expected: domain.TicketTypeChore},
			{input: "\tgEn-42\n", expected: domain.TicketTypeChore},
			{input: "\r\nGen-888\t", expected: domain.TicketTypeChore},
		}
		for _, tc := range cases {
			res := git.InferTicketType(tc.input, configuredPrefixes...)
			if res == nil {
				t.Fatalf("expected non-nil for %q, got nil", tc.input)
			}
			if *res != tc.expected {
				t.Errorf("InferTicketType(%q) = %q, expected %q", tc.input, *res, tc.expected)
			}
		}
	})

	t.Run("ticket with empty or invalid suffix returns nil", func(t *testing.T) {
		t.Parallel()
		invalidSuffixes := []string{
			"PO-",
			"GEN-",
			"PO-   ",
			"GEN-   ",
			"\tPO-\n",
			"\tGEN-\n",
			"PO_",
			"GEN_",
		}
		for _, ticketID := range invalidSuffixes {
			res := git.InferTicketType(ticketID, configuredPrefixes...)
			if res != nil {
				t.Errorf("expected nil for ticket with invalid suffix %q, got %q", ticketID, *res)
			}
		}
	})
}

func strPtr(s domain.ConventionalTicketType) *domain.ConventionalTicketType {
	return &s
}

func TestComputeSDLCIterations_ZeroCallSitesOrExports(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed resolving cwd: %v", err)
	}

	backendDir := cwd
	for {
		if filepath.Base(backendDir) == "backend" {
			break
		}
		cand := filepath.Join(backendDir, "backend")
		if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
			backendDir = cand
			break
		}
		parent := filepath.Dir(backendDir)
		if parent == backendDir {
			t.Fatalf("could not locate backend directory starting from %s", cwd)
		}
		backendDir = parent
	}

	fset := token.NewFileSet()
	var violations []string

	err = filepath.Walk(backendDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return parseErr
		}

		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				if x.Name.Name == "ComputeSDLCIterations" {
					violations = append(violations, "declaration: "+path+":"+fset.Position(x.Pos()).String())
				}
			case *ast.CallExpr:
				if sel, ok := x.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "ComputeSDLCIterations" {
					violations = append(violations, "call: "+path+":"+fset.Position(sel.Pos()).String())
				} else if ident, ok := x.Fun.(*ast.Ident); ok && ident.Name == "ComputeSDLCIterations" {
					violations = append(violations, "call: "+path+":"+fset.Position(ident.Pos()).String())
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "ComputeSDLCIterations" {
					violations = append(violations, "selector: "+path+":"+fset.Position(x.Pos()).String())
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("walking backend failed: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("invariant violation: ComputeSDLCIterations must have zero call sites or exports, found: %v", violations)
	}
}
