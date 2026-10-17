package git_test

import (
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func TestParseNumstatOutput(t *testing.T) {
	raw := `
12	4	internal/adapters/git/analyzer.go
100	0	internal/adapters/git/analyzer_test.go
-	-	image.png
3	1	path with spaces/file.go
`
	entries := git.ParseNumstatOutput(raw)
	if len(entries) != 3 {
		t.Fatalf("expected 3 parsed entries, got %d", len(entries))
	}

	if entries[0].Added != 12 || entries[0].Deleted != 4 || entries[0].Path != "internal/adapters/git/analyzer.go" {
		t.Errorf("unexpected entry 0: %+v", entries[0])
	}
	if entries[1].Added != 100 || entries[1].Deleted != 0 || entries[1].Path != "internal/adapters/git/analyzer_test.go" {
		t.Errorf("unexpected entry 1: %+v", entries[1])
	}
	if entries[2].Added != 3 || entries[2].Deleted != 1 || entries[2].Path != "path with spaces/file.go" {
		t.Errorf("unexpected entry 2: %+v", entries[2])
	}
}

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		path     string
		cfg      *git.ClassificationConfig
		expected git.FileCategory
	}{
		// Ignored files
		{"README.md", nil, git.CategoryIgnored},
		{"docs/ARCH.md", nil, git.CategoryIgnored},
		{"package.json", nil, git.CategoryIgnored},
		{"go.sum", nil, git.CategoryIgnored},
		{"package-lock.json", nil, git.CategoryIgnored},
		{"pnpm-lock.yaml", nil, git.CategoryIgnored},
		{"workspaces/AL-11/CONTRACT.md", nil, git.CategoryIgnored},
		{"workspaces/AL-11/agentic_metrics.json", nil, git.CategoryIgnored},
		{".gitignore", nil, git.CategoryIgnored},

		// Test files
		{"internal/adapters/git/analyzer_test.go", nil, git.CategoryTest},
		{"tests/test_api.py", nil, git.CategoryTest},
		{"src/components/Card.test.tsx", nil, git.CategoryTest},
		{"src/utils/math.spec.ts", nil, git.CategoryTest},
		{"tests/unit/test_runner.py", nil, git.CategoryTest},
		{"__tests__/dashboard.test.js", nil, git.CategoryTest},

		// Production files
		{"internal/domain/models.go", nil, git.CategoryProd},
		{"cmd/agentic-lens/main.go", nil, git.CategoryProd},
		{"web/src/App.tsx", nil, git.CategoryProd},
		{"web/src/main.ts", nil, git.CategoryProd},
		{"web/index.html", nil, git.CategoryProd},
		{"web/src/style.css", nil, git.CategoryProd},
		{"Makefile", nil, git.CategoryProd},
		{"scripts/setup.sh", nil, git.CategoryProd},
		{"db/schema.sql", nil, git.CategoryProd},

		// Config overrides
		{
			"custom/special.xyz",
			&git.ClassificationConfig{
				ProdPatterns: []string{"custom/*.xyz"},
			},
			git.CategoryProd,
		},
		{
			"internal/domain/models.go",
			&git.ClassificationConfig{
				Exclude: []string{"internal/domain/**"},
			},
			git.CategoryIgnored,
		},
		{
			"custom/test.xyz",
			&git.ClassificationConfig{
				TestPatterns: []string{"custom/test.*"},
			},
			git.CategoryTest,
		},
	}

	for _, tt := range tests {
		actual := git.ClassifyFile(tt.path, tt.cfg)
		if actual != tt.expected {
			t.Errorf("ClassifyFile(%q) = %s, expected %s", tt.path, actual, tt.expected)
		}
	}
}

func TestInferenceOnActiveWorkspace(t *testing.T) {
	// Active ticket in this worktree is AL-11
	title := git.InferTicketTitle("AL-11")
	if title == nil || *title == "" {
		t.Errorf("expected non-empty title for AL-11, got nil")
	} else if *title != "Native Go Telemetry Extractor and Active Session Metric Generation" {
		t.Logf("inferred title for AL-11: %s", *title)
	}

	ticketType := git.InferTicketType("AL-11")
	if ticketType == nil || *ticketType != "feature" {
		t.Errorf("expected feature ticket type for AL-11, got %v", ticketType)
	}

	iters := git.ComputeSDLCIterations("AL-11")
	if iters < 1 {
		t.Errorf("expected sdlc iterations >= 1, got %d", iters)
	}

	user := git.GetGitUser()
	if user == "" {
		t.Errorf("expected non-empty git user")
	}
}

func TestGitAnalyzer_AdversarialEdgeCases(t *testing.T) {
	t.Run("empty or whitespace ticket ID returns empty numstats and safe defaults", func(t *testing.T) {
		entries := git.GetTicketNumstat("")
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for empty ticket ID, got %d", len(entries))
		}

		entriesWs := git.GetTicketNumstat("   ")
		if len(entriesWs) != 0 {
			t.Errorf("expected 0 entries for whitespace ticket ID, got %d", len(entriesWs))
		}

		title := git.InferTicketTitle("")
		if title != nil {
			t.Errorf("expected nil title for empty ticket ID, got %v", title)
		}

		tType := git.InferTicketType("")
		if tType != nil {
			t.Errorf("expected nil type for empty ticket ID, got %v", tType)
		}

		iters := git.ComputeSDLCIterations("")
		if iters != 1 {
			t.Errorf("expected 1 iteration for empty ticket ID, got %d", iters)
		}
	})

	t.Run("missing git revisions and non-existent ticket numstat fallback", func(t *testing.T) {
		// Non-existent ticket should return empty numstat without panicking
		entries := git.GetTicketNumstat("NON-EXISTENT-TICKET-XYZ-999")
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for non-existent ticket, got %d", len(entries))
		}

		fallback := &git.NumstatEntry{Added: 10, Deleted: 5, Path: "fallback.go"}
		_ = fallback
	})

	t.Run("ComputeCodeFacts with fallback and empty git numstat", func(t *testing.T) {
		// When numstat is empty, fallback CodeFacts should be returned untouched
		fallbackCode := domain.CodeFacts{
			LinesAdded:       100,
			LinesDeleted:     20,
			TestLinesAdded:   40,
			TestLinesDeleted: 5,
			ProdLinesAdded:   60,
			ProdLinesDeleted: 15,
			FilesChanged:     3,
			SDLCIterations:   2,
		}

		facts := git.ComputeCodeFacts("NON-EXISTENT-TICKET-XYZ-999", nil, &fallbackCode)
		if facts.LinesAdded != 100 || facts.LinesDeleted != 20 || facts.FilesChanged != 3 || facts.SDLCIterations != 2 {
			t.Errorf("expected fallback code facts to be preserved, got %+v", facts)
		}

		// When fallback is nil, it should return zeroed CodeFacts with sdlc iterations = 1
		zeroFacts := git.ComputeCodeFacts("NON-EXISTENT-TICKET-XYZ-999", nil, nil)
		if zeroFacts.LinesAdded != 0 || zeroFacts.LinesDeleted != 0 || zeroFacts.FilesChanged != 0 {
			t.Errorf("expected zero code facts, got %+v", zeroFacts)
		}
		if zeroFacts.SDLCIterations != 1 {
			t.Errorf("expected 1 sdlc iteration for missing workspace, got %d", zeroFacts.SDLCIterations)
		}
	})

	t.Run("corrupted or malformed numstat text lines", func(t *testing.T) {
		malformedText := `
not_a_number	not_a_number	file.go
10
10	20
-	-	binary.png
	
15	5
-5	-10	negative.go
`
		parsed := git.ParseNumstatOutput(malformedText)
		// Only line "-5 -10 negative.go" has 3 parts and valid int parsing
		if len(parsed) != 1 {
			t.Fatalf("expected 1 parsed entry from malformed text, got %d: %+v", len(parsed), parsed)
		}
		if parsed[0].Path != "negative.go" {
			t.Errorf("expected negative.go, got %s", parsed[0].Path)
		}
	})
}

