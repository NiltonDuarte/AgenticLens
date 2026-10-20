package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func createTestBranch(t *testing.T, branchName string) {
	t.Helper()
	_ = exec.Command("git", "branch", "-D", branchName).Run()
	cmd := exec.Command("git", "branch", branchName, "HEAD")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create test branch %s: %v (output: %s)", branchName, err, string(out))
	}
	t.Cleanup(func() {
		_ = exec.Command("git", "branch", "-D", branchName).Run()
	})
}

func TestInferTicketType_PrefixMapping(t *testing.T) {
	t.Parallel()

	configuredPrefixes := []domain.TicketPrefix{
		{Prefix: "PO", TicketType: "docs"},
		{Prefix: "GEN", TicketType: "chore"},
	}

	tests := []struct {
		name     string
		ticketID string
		expected string
	}{
		{name: "PO ticket uppercase", ticketID: "PO-5", expected: "docs"},
		{name: "po ticket lowercase", ticketID: "po-12", expected: "docs"},
		{name: "Po ticket mixed case", ticketID: "Po-99", expected: "docs"},
		{name: "GEN ticket uppercase", ticketID: "GEN-42", expected: "chore"},
		{name: "gen ticket lowercase", ticketID: "gen-1", expected: "chore"},
		{name: "Gen ticket mixed case", ticketID: "Gen-100", expected: "chore"},
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
	// Verify branches in repo are resolved to their conventional commit types
	branchTests := []struct {
		ticketID   string
		branchName string
		expected   string
	}{
		{ticketID: "AL-1", branchName: "feat/AL-1-ci-fixture", expected: "feat"},
		{ticketID: "AL-14", branchName: "ci/AL-14-ci-fixture", expected: "ci"},
		{ticketID: "AL-16", branchName: "refactor/AL-16-ci-fixture", expected: "refactor"},
		{ticketID: "AL-30", branchName: "fix/AL-30-ci-fixture", expected: "fix"},
	}

	for _, tc := range branchTests {
		createTestBranch(t, tc.branchName)
		res := git.InferTicketType(tc.ticketID)
		if res == nil {
			t.Fatalf("expected non-nil ticket type for branch ticket %s, got nil", tc.ticketID)
		}
		if *res != tc.expected {
			t.Errorf("InferTicketType(%q) = %q, expected %q", tc.ticketID, *res, tc.expected)
		}
	}
}

func TestInferTicketType_Precedence_BranchOverConfiguredPrefix(t *testing.T) {
	// Create transient branch fixture for AL-30
	createTestBranch(t, "fix/AL-30-ci-fixture")

	// Supplying a configured prefix mapping AL -> chore MUST NOT override the branch type "fix".
	conflictingPrefixes := []domain.TicketPrefix{
		{Prefix: "AL", TicketType: "chore"},
	}

	actual := git.InferTicketType("AL-30", conflictingPrefixes...)
	if actual == nil {
		t.Fatal("expected non-nil ticket type for AL-30, got nil")
	}
	if *actual != "fix" {
		t.Errorf("expected branch prefix 'fix' to take priority over configured 'chore', got %q", *actual)
	}

	// Conversely, when branch is absent (e.g. PO-5 or custom unbranched ticket),
	// configured prefixes act as the backup.
	poActual := git.InferTicketType("PO-5", domain.TicketPrefix{Prefix: "PO", TicketType: "docs"})
	if poActual == nil || *poActual != "docs" {
		t.Errorf("expected backup configured prefix 'docs' for unbranched PO-5, got %v", poActual)
	}

	unbranchedActual := git.InferTicketType("CUSTOM-99", domain.TicketPrefix{Prefix: "CUSTOM", TicketType: "refactor"})
	if unbranchedActual == nil || *unbranchedActual != "refactor" {
		t.Errorf("expected backup configured prefix 'refactor' for unbranched CUSTOM-99, got %v", unbranchedActual)
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
		expected   *string
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
			expected string
		}{
			{input: "  PO-1  ", expected: "docs"},
			{input: "\tpO-42\n", expected: "docs"},
			{input: "\r\nPo-999\t", expected: "docs"},
			{input: "  GEN-1  ", expected: "chore"},
			{input: "\tgEn-42\n", expected: "chore"},
			{input: "\r\nGen-888\t", expected: "chore"},
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

func strPtr(s string) *string {
	return &s
}

func TestComputeSDLCIterations_CustomWorkspacesDir(t *testing.T) {
	tempDir := t.TempDir()
	customWs := "my_custom_ws"
	ticketID := "CUSTOM-1"
	iterDir := filepath.Join(tempDir, customWs, ticketID, "iterations")
	if err := os.MkdirAll(filepath.Join(iterDir, "01"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(iterDir, "02"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(iterDir, "03"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(iterDir, "04"), 0755); err != nil {
		t.Fatal(err)
	}

	// Adversarial: non-directory files inside iterations/ must not be counted
	readmeFile := filepath.Join(iterDir, "README.md")
	if err := os.WriteFile(readmeFile, []byte("iteration notes"), 0644); err != nil {
		t.Fatal(err)
	}
	gitkeepFile := filepath.Join(iterDir, ".gitkeep")
	if err := os.WriteFile(gitkeepFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	iters := git.ComputeSDLCIterations(ticketID, customWs)
	if iters != 4 {
		t.Errorf("expected exactly 4 iterations (excluding files) in custom workspace, got %d", iters)
	}

	// Slashes and whitespace in workspacesDir argument must normalize correctly
	itersNormalized := git.ComputeSDLCIterations(ticketID, "  /"+customWs+"/  ")
	if itersNormalized != 4 {
		t.Errorf("expected 4 iterations with normalized workspacesDir, got %d", itersNormalized)
	}

	// Default unconfigured workspace dir (which does not exist for CUSTOM-1) returns 1
	defaultIters := git.ComputeSDLCIterations(ticketID)
	if defaultIters != 1 {
		t.Errorf("expected 1 iteration for unconfigured workspace dir (not found), got %d", defaultIters)
	}

	// Non-existent ticket returns 1
	missingIters := git.ComputeSDLCIterations("NONEXISTENT-999", customWs)
	if missingIters != 1 {
		t.Errorf("expected 1 iteration for missing ticket, got %d", missingIters)
	}
}
