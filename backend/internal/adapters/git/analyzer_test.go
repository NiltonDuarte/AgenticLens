package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func TestParseNumstatOutput(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	// Test branch inference purely in memory without mutating disk branches
	branchList := "  feat/AL-11-ci-fixture\n"
	ticketType := git.InferTicketTypeFromBranchList(branchList, "AL-11")
	if ticketType == nil || *ticketType != "feat" {
		t.Errorf("expected feat ticket type for AL-11, got %v", ticketType)
	}

	iters := git.ComputeSDLCIterations("AL-11")
	if iters != 1 {
		t.Errorf("expected sdlc iterations == 1, got %d", iters)
	}

	user := git.GetGitUser()
	if user == "" {
		t.Errorf("expected non-empty git user")
	}
	if strings.Contains(user, "@") {
		t.Errorf("expected git user not to leak full email domain, got: %s", user)
	}
}

func TestGitAnalyzer_AdversarialEdgeCases(t *testing.T) {
	t.Run("empty or whitespace ticket ID returns empty numstats and safe defaults", func(t *testing.T) {
		entries, err := git.GetTicketNumstat("", "main", []string{"backend"}, nil)
		if err != nil {
			t.Fatalf("unexpected error for empty ticket ID: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries for empty ticket ID, got %d", len(entries))
		}

		entriesWs, err := git.GetTicketNumstat("   ", "main", []string{"backend"}, nil)
		if err != nil {
			t.Fatalf("unexpected error for whitespace ticket ID: %v", err)
		}
		if len(entriesWs) != 0 {
			t.Errorf("expected 0 entries for whitespace ticket ID, got %d", len(entriesWs))
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
		entries, err := git.GetTicketNumstat("NON-EXISTENT-TICKET-XYZ-999", "main", []string{"backend"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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

		facts, err := git.ComputeCodeFacts("NON-EXISTENT-TICKET-XYZ-999", nil, &fallbackCode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if facts.LinesAdded != 100 || facts.LinesDeleted != 20 || facts.FilesChanged != 3 || facts.SDLCIterations != 2 {
			t.Errorf("expected fallback code facts to be preserved, got %+v", facts)
		}

		// When fallback is nil, it should return zeroed CodeFacts with sdlc iterations = 1
		zeroFacts, err := git.ComputeCodeFacts("NON-EXISTENT-TICKET-XYZ-999", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
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

	t.Run("GetWorkspaceCommitTimeRange", func(t *testing.T) {
		minTs, maxTs, found := git.GetWorkspaceCommitTimeRange("")
		if found || minTs != 0 || maxTs != 0 {
			t.Errorf("expected not found for empty workspace, got found=%v", found)
		}

		minTs, maxTs, found = git.GetWorkspaceCommitTimeRange("workspaces/PO-3")
		if !found || minTs <= 0 || maxTs < minTs {
			t.Errorf("expected valid range for workspaces/PO-3, got min=%d max=%d found=%v", minTs, maxTs, found)
		}
	})
}

func TestLoadConfig_CentralizedIntegration(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	cfgDir := filepath.Join(tempDir, ".agenticlens")
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := `
project_name: GitAnalyzerTest
ticket:
  prefixes:
    - prefix: AL
workspaces:
  src_folders:
    - src
code_metrics:
  include:
    - "custom/include/**"
  exclude:
    - "custom/exclude/**"
  test_patterns:
    - "*_test_custom.go"
  prod_patterns:
    - "*.custom_prod"
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := git.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil ClassificationConfig")
	}

	if len(cfg.Include) != 1 || cfg.Include[0] != "custom/include/**" {
		t.Errorf("unexpected include: %v", cfg.Include)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "custom/exclude/**" {
		t.Errorf("unexpected exclude: %v", cfg.Exclude)
	}
	if len(cfg.TestPatterns) != 1 || cfg.TestPatterns[0] != "*_test_custom.go" {
		t.Errorf("unexpected test_patterns: %v", cfg.TestPatterns)
	}
	if len(cfg.ProdPatterns) != 1 || cfg.ProdPatterns[0] != "*.custom_prod" {
		t.Errorf("unexpected prod_patterns: %v", cfg.ProdPatterns)
	}

	// Verify classification applies loaded patterns
	catInc := git.ClassifyFile("custom/include/file.xyz", cfg)
	if catInc != git.CategoryProd {
		t.Errorf("expected CategoryProd for custom/include/file.xyz, got %s", catInc)
	}

	catExc := git.ClassifyFile("custom/exclude/file.go", cfg)
	if catExc != git.CategoryIgnored {
		t.Errorf("expected CategoryIgnored for custom/exclude/file.go, got %s", catExc)
	}
}

func initTestGitRepo(t *testing.T, dir string, branch string) {
	t.Helper()
	runCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, string(out))
		}
	}
	runCmd("init", "-b", branch)
	runCmd("config", "user.name", "TestUser")
	runCmd("config", "user.email", "test@example.com")

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "README.md")
	runCmd("commit", "-m", "initial commit")
}

func TestResolveBaseBranch_Configured(t *testing.T) {
	t.Run("configured branch exists succeeds", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		cmd := exec.Command("git", "checkout", "-b", "release-1.0")
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed creating branch release-1.0: %v (%s)", err, string(out))
		}

		cfg := &domain.ProjectConfig{
			CodeMetrics: domain.CodeMetricsConfig{
				BaseBranch: "release-1.0",
			},
		}
		branch, err := git.ResolveBaseBranch(tempDir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "release-1.0" {
			t.Errorf("expected release-1.0, got %q", branch)
		}
	})

	t.Run("configured branch with whitespace trimmed succeeds", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		cmd := exec.Command("git", "checkout", "-b", "release-2.0")
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed creating branch release-2.0: %v (%s)", err, string(out))
		}

		cfg := &domain.ProjectConfig{
			CodeMetrics: domain.CodeMetricsConfig{
				BaseBranch: "  release-2.0  ",
			},
		}
		branch, err := git.ResolveBaseBranch(tempDir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "release-2.0" {
			t.Errorf("expected release-2.0, got %q", branch)
		}
	})

	t.Run("configured empty string falls through to candidate probing", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		cfg := &domain.ProjectConfig{
			CodeMetrics: domain.CodeMetricsConfig{
				BaseBranch: "   ",
			},
		}
		branch, err := git.ResolveBaseBranch(tempDir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("expected main, got %q", branch)
		}
	})
}

func TestResolveBaseBranch_Configured_Missing(t *testing.T) {
	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "main")

	cfg := &domain.ProjectConfig{
		CodeMetrics: domain.CodeMetricsConfig{
			BaseBranch: "nonexistent-branch",
		},
	}
	branch, err := git.ResolveBaseBranch(tempDir, cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent configured branch, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-branch") {
		t.Errorf("expected error mentioning nonexistent-branch, got: %v", err)
	}
	if branch != "" {
		t.Errorf("expected empty branch string on error, got %q", branch)
	}
	if branch == "HEAD" {
		t.Error("prohibition violation: returned HEAD on missing configured branch")
	}
}

func TestResolveBaseBranch_SymbolicRef(t *testing.T) {
	t.Run("valid symbolic ref detects remote tracking branch", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "trunk")

		cmd := exec.Command("git", "update-ref", "refs/remotes/origin/trunk", "HEAD")
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed update-ref: %v (%s)", err, string(out))
		}
		cmd = exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk")
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed symbolic-ref: %v (%s)", err, string(out))
		}

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "origin/trunk" {
			t.Errorf("expected origin/trunk, got %q", branch)
		}
	})

	t.Run("broken symbolic ref falls through to candidate probes", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/nonexistent")
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("failed symbolic-ref: %v (%s)", err, string(out))
		}

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("expected main after broken symbolic ref fallthrough, got %q", branch)
		}
	})
}

func TestResolveBaseBranch_ProbeCandidates(t *testing.T) {
	t.Run("probe candidate priority 1: origin/main beats all candidates", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		runCmd := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = tempDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
			}
		}
		runCmd("update-ref", "refs/remotes/origin/main", "HEAD")
		runCmd("update-ref", "refs/remotes/origin/master", "HEAD")
		runCmd("branch", "master", "HEAD")

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "origin/main" {
			t.Errorf("expected origin/main, got %q", branch)
		}
	})

	t.Run("probe candidate priority 2: main beats origin/master and master", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		runCmd := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = tempDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
			}
		}
		runCmd("update-ref", "refs/remotes/origin/master", "HEAD")
		runCmd("branch", "master", "HEAD")

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "main" {
			t.Errorf("expected main, got %q", branch)
		}
	})

	t.Run("probe candidate priority 3: origin/master beats master", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "dev-temp")

		runCmd := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = tempDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
			}
		}
		runCmd("update-ref", "refs/remotes/origin/master", "HEAD")
		runCmd("branch", "master", "HEAD")

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "origin/master" {
			t.Errorf("expected origin/master, got %q", branch)
		}
	})

	t.Run("probe candidate priority 4: master succeeds when only candidate", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "master")

		branch, err := git.ResolveBaseBranch(tempDir, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if branch != "master" {
			t.Errorf("expected master, got %q", branch)
		}
	})
}

func TestResolveBaseBranch_Unresolvable_FailureInvariant(t *testing.T) {
	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "dev-temp")

	branch, err := git.ResolveBaseBranch(tempDir, nil)
	if err == nil {
		t.Fatalf("expected error for unresolvable branch, got branch: %q", branch)
	}
	expectedMsg := "unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
	if branch != "" {
		t.Errorf("expected empty branch string on failure, got %q", branch)
	}
	if branch == "HEAD" {
		t.Error("prohibition violation: returned HEAD as base branch")
	}
}

func TestValidateSrcFolders(t *testing.T) {
	expectedMsg := "workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml"

	t.Run("nil slice fails with exact configuration prompt", func(t *testing.T) {
		err := git.ValidateSrcFolders(nil)
		if err == nil {
			t.Fatal("expected error for nil src_folders, got nil")
		}
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("empty slice fails with exact configuration prompt", func(t *testing.T) {
		err := git.ValidateSrcFolders([]string{})
		if err == nil {
			t.Fatal("expected error for empty src_folders, got nil")
		}
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("whitespace only slices fail with exact configuration prompt", func(t *testing.T) {
		whitespaceCases := [][]string{
			{"   "},
			{"\t", "\n"},
			{"", "  ", "\t", "\r\n"},
		}
		for _, tc := range whitespaceCases {
			err := git.ValidateSrcFolders(tc)
			if err == nil {
				t.Fatalf("expected error for whitespace src_folders %v, got nil", tc)
			}
			if !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
			}
		}
	})

	t.Run("valid slice succeeds", func(t *testing.T) {
		err := git.ValidateSrcFolders([]string{"backend", "web"})
		if err != nil {
			t.Errorf("unexpected error for valid src_folders: %v", err)
		}
	})

	t.Run("slice with empty strings but at least one valid folder succeeds", func(t *testing.T) {
		err := git.ValidateSrcFolders([]string{"", "   ", "services", "\t"})
		if err != nil {
			t.Errorf("unexpected error for mixed src_folders: %v", err)
		}
	})
}

func TestIsMonitoredSourcePath(t *testing.T) {
	srcFolders := []string{"app", "pkg", "services"}
	cfgWithInclude := &git.ClassificationConfig{
		Include: []string{"main.py", "docker-compose.yml", "*.config.ts"},
	}
	cfgWithProd := &git.ClassificationConfig{
		ProdPatterns: []string{"*.vue", "core/**"},
	}

	tests := []struct {
		name       string
		path       string
		srcFolders []string
		cfg        *git.ClassificationConfig
		expected   bool
	}{
		// Custom folder structures: app/, pkg/, services/
		{"app root subfile", "app/main.go", srcFolders, nil, true},
		{"app nested file", "app/controllers/user.go", srcFolders, nil, true},
		{"pkg root subfile", "pkg/util.go", srcFolders, nil, true},
		{"pkg nested file", "pkg/storage/db.go", srcFolders, nil, true},
		{"services root subfile", "services/worker.go", srcFolders, nil, true},
		{"services nested file", "services/auth/jwt.go", srcFolders, nil, true},
		{"exact folder name app", "app", srcFolders, nil, true},
		{"exact folder name services", "services", srcFolders, nil, true},
		{"exact folder name pkg", "pkg", srcFolders, nil, true},

		// Leading dot and slashes normalization
		{"dot slash prefix", "./app/main.go", srcFolders, nil, true},
		{"leading slash prefix", "/services/api/handler.go", srcFolders, nil, true},

		// Configured folders with slashes inside slice
		{"folders with leading and trailing slashes", "app/server.go", []string{"/app/", "pkg/", "/services"}, nil, true},
		{"folders with leading and trailing slashes pkg", "pkg/helper.go", []string{"/app/", "pkg/", "/services"}, nil, true},
		{"folders with leading and trailing slashes services", "services/worker.go", []string{"/app/", "pkg/", "/services"}, nil, true},

		// Boundary false prefix matches (must NOT match)
		{"file with app prefix but not in app folder", "app_helper.go", srcFolders, nil, false},
		{"folder starting with app name", "application/main.go", srcFolders, nil, false},
		{"folder starting with services name", "services_old/file.go", srcFolders, nil, false},
		{"folder starting with pkg name", "pkg-utils/main.go", srcFolders, nil, false},

		// Root files matching cfg.Include
		{"root python file matching include", "main.py", []string{"backend"}, cfgWithInclude, true},
		{"docker compose file matching include", "docker-compose.yml", []string{"backend"}, cfgWithInclude, true},
		{"config typescript file matching include glob", "vite.config.ts", []string{"backend"}, cfgWithInclude, true},

		// Files matching cfg.ProdPatterns
		{"vue component matching prod pattern", "web/src/view.vue", []string{"backend"}, cfgWithProd, true},
		{"cpp file inside core folder matching prod pattern", "core/engine.cpp", []string{"backend"}, cfgWithProd, true},

		// Non-monitored files
		{"documentation markdown file", "docs/spec.md", srcFolders, nil, false},
		{"random data binary file", "random/data.bin", srcFolders, nil, false},
		{"root notes file", "notes.txt", srcFolders, nil, false},
		{"workspace contract file", "workspaces/AL-39/CONTRACT.md", srcFolders, nil, false},
		{"github workflow file", ".github/workflows/ci.yml", srcFolders, nil, false},

		// Empty and whitespace paths
		{"empty path", "", srcFolders, nil, false},
		{"whitespace path", "   ", srcFolders, nil, false},
		{"tab path", "\t", srcFolders, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := git.IsMonitoredSourcePath(tt.path, tt.srcFolders, tt.cfg)
			if actual != tt.expected {
				t.Errorf("IsMonitoredSourcePath(%q, %v) = %v, expected %v", tt.path, tt.srcFolders, actual, tt.expected)
			}
		})
	}
}

func TestGetTicketNumstat_DynamicBranchAndDuplicateConditionsEliminated(t *testing.T) {
	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "master")

	runCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
		}
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

	// 1. Branch for TICK-SVC touching custom monitored folder "services" and unmonitored "docs"
	runCmd("checkout", "-b", "feat/TICK-SVC-billing")
	svcFile := filepath.Join(tempDir, "services", "billing", "charge.go")
	if err := os.MkdirAll(filepath.Dir(svcFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(svcFile, []byte("package billing\nfunc Charge() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	docFile := filepath.Join(tempDir, "docs", "guide.md")
	if err := os.MkdirAll(filepath.Dir(docFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docFile, []byte("# Billing Guide\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "services/billing/charge.go", "docs/guide.md")
	runCmd("commit", "-m", "feat(TICK-SVC): billing charge service")

	// 2. Branch for TICK-UNMON modifying ONLY unmonitored files (random and docs)
	runCmd("checkout", "master")
	runCmd("checkout", "-b", "feat/TICK-UNMON-notes")
	randomFile := filepath.Join(tempDir, "random", "notes.txt")
	if err := os.MkdirAll(filepath.Dir(randomFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(randomFile, []byte("internal notes\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "random/notes.txt")
	runCmd("commit", "-m", "docs(TICK-UNMON): notes only")

	// 3. Branch for TICK-BACKEND modifying "backend/legacy.go"
	runCmd("checkout", "master")
	runCmd("checkout", "-b", "feat/TICK-BACKEND-legacy")
	backendFile := filepath.Join(tempDir, "backend", "legacy.go")
	if err := os.MkdirAll(filepath.Dir(backendFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(backendFile, []byte("package backend\nfunc Old() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "backend/legacy.go")
	runCmd("commit", "-m", "feat(TICK-BACKEND): backend legacy")

	// Switch back to master so branch-level checks (Step 3) are exercised
	runCmd("checkout", "master")

	// Verification 1: TICK-SVC diffed against master with srcFolders ["services"] captures charge.go
	svcEntries, err := git.GetTicketNumstat("TICK-SVC", "master", []string{"services"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(svcEntries) == 0 {
		t.Fatal("expected entries for custom source folder services, got 0")
	}
	foundSvc := false
	for _, e := range svcEntries {
		if e.Path == "services/billing/charge.go" {
			foundSvc = true
			if e.Added <= 0 {
				t.Errorf("expected added lines > 0, got %d", e.Added)
			}
		}
	}
	if !foundSvc {
		t.Errorf("expected services/billing/charge.go in numstat entries, got %+v", svcEntries)
	}

	// Verification 2: TICK-UNMON modifying ONLY unmonitored files returns 0 numstat entries
	unmonitoredEntries, err := git.GetTicketNumstat("TICK-UNMON", "master", []string{"services"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unmonitoredEntries) != 0 {
		t.Errorf("expected 0 entries when branch only modifies unmonitored paths, got %d: %+v", len(unmonitoredEntries), unmonitoredEntries)
	}

	// Verification 3: TICK-BACKEND with srcFolders ["services"] returns 0 entries
	// This proves that the old hardcoded list of 6 folders was removed and configuration is respected
	backendEntries, err := git.GetTicketNumstat("TICK-BACKEND", "master", []string{"services"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(backendEntries) != 0 {
		t.Errorf("expected 0 entries for backend path when src_folders is ['services'], got %d: %+v", len(backendEntries), backendEntries)
	}
}

func TestGetTicketNumstat_DynamicBranch(t *testing.T) {
	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "master")

	runCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, string(out))
		}
	}

	runCmd("checkout", "-b", "feat/AL-39-decouple")
	appFile := filepath.Join(tempDir, "app", "feature.go")
	if err := os.MkdirAll(filepath.Dir(appFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(appFile, []byte("package app\nfunc Feature() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "app/feature.go")
	runCmd("commit", "-m", "feat(AL-39): decouple git analyzer")

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

	entries, err := git.GetTicketNumstat("AL-39", "master", []string{"app"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry for ticket AL-39 diffed against master, got 0")
	}
	found := false
	for _, e := range entries {
		if e.Path == "app/feature.go" {
			found = true
			if e.Added <= 0 {
				t.Errorf("expected added lines > 0, got %d", e.Added)
			}
		}
	}
	if !found {
		t.Errorf("expected app/feature.go in entries, got %+v", entries)
	}
}

func TestGetTicketNumstat_BaseBranchFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "main")

	runCmd := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = tempDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, string(out))
		}
	}

	runCmd("checkout", "-b", "feat/AL-39-cfg-branch")
	appFile := filepath.Join(tempDir, "app", "feature.go")
	if err := os.MkdirAll(filepath.Dir(appFile), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(appFile, []byte("package app\nfunc Feature() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCmd("add", "app/feature.go")
	runCmd("commit", "-m", "feat(AL-39): test config base branch")

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

	t.Run("unresolvable base branch in config returns validation error", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			BaseBranch: "nonexistent-branch",
			SrcFolders: []string{"app"},
		}
		_, err := git.GetTicketNumstat("AL-39", "", []string{"app"}, cfg)
		if err == nil {
			t.Fatal("expected error for unresolvable base branch in config, got nil")
		}
		expectedPrompt := "configured base branch \"nonexistent-branch\" not found"
		if !strings.Contains(err.Error(), expectedPrompt) {
			t.Errorf("expected error to contain %q, got %q", expectedPrompt, err.Error())
		}
	})

	t.Run("valid base branch in config resolves and succeeds", func(t *testing.T) {
		cfg := &git.ClassificationConfig{
			BaseBranch: "main",
			SrcFolders: []string{"app"},
		}
		entries, err := git.GetTicketNumstat("AL-39", "", []string{"app"}, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) == 0 {
			t.Fatal("expected at least 1 entry for ticket AL-39 with config base branch, got 0")
		}
	})
}

func TestComputeCodeFacts_Invariants(t *testing.T) {
	t.Run("empty src_folders returns explicit invariant error", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

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

		cfg := &git.ClassificationConfig{
			SrcFolders: []string{},
		}
		_, err = git.ComputeCodeFacts("TICK-1", cfg, nil)
		if err == nil {
			t.Fatal("expected error for empty src_folders, got nil")
		}
		expectedMsg := "workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("unresolvable base branch returns explicit invariant error", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "dev-temp")

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

		cfg := &git.ClassificationConfig{
			SrcFolders: []string{"backend"},
		}
		_, err = git.ComputeCodeFacts("TICK-1", cfg, nil)
		if err == nil {
			t.Fatal("expected error for unresolvable base branch, got nil")
		}
		expectedMsg := "unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("custom workspaces_dir resolves iterations in ComputeCodeFacts", func(t *testing.T) {
		tempDir := t.TempDir()
		initTestGitRepo(t, tempDir, "main")

		customWs := "custom_tickets"
		ticketID := "TICK-100"

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

		cfg := &git.ClassificationConfig{
			SrcFolders:    []string{"backend"},
			WorkspacesDir: customWs,
		}
		facts, err := git.ComputeCodeFacts(ticketID, cfg, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if facts.SDLCIterations != 1 {
			t.Errorf("expected 1 iteration without filesystem lookup, got %d", facts.SDLCIterations)
		}

		fallback := &domain.CodeFacts{
			LinesAdded:     10,
			FilesChanged:   1,
			SDLCIterations: 4,
		}
		factsWithFallback, err := git.ComputeCodeFacts(ticketID, cfg, fallback)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if factsWithFallback.SDLCIterations != 4 {
			t.Errorf("expected 4 iterations preserved from fallback, got %d", factsWithFallback.SDLCIterations)
		}
	})
}
