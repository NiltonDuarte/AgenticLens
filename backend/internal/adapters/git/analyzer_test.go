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
		// Ignored files: lockfiles, dotfiles, and configured workspaces
		{"go.sum", nil, git.CategoryIgnored},
		{"package-lock.json", nil, git.CategoryIgnored},
		{"pnpm-lock.yaml", nil, git.CategoryIgnored},
		{"yarn.lock", nil, git.CategoryIgnored},
		{".gitignore", nil, git.CategoryIgnored},
		{".env", nil, git.CategoryIgnored},
		{"workspaces/AL-11/CONTRACT.md", &git.ClassificationConfig{WorkspacesDir: "workspaces"}, git.CategoryIgnored},
		{"workspaces/AL-11/agentic_metrics.json", &git.ClassificationConfig{WorkspacesDir: "workspaces"}, git.CategoryIgnored},
		{"workspaces/file.go", &git.ClassificationConfig{WorkspacesDir: "workspaces"}, git.CategoryIgnored},
		{"workspaces/spec.md", &git.ClassificationConfig{WorkspacesDir: "workspaces"}, git.CategoryIgnored},

		// Other files: markdown, json, lock without specific lockfile names, or workspaces with nil config
		{"README.md", nil, git.CategoryOther},
		{"docs/ARCH.md", nil, git.CategoryOther},
		{"package.json", nil, git.CategoryOther},
		{"test.md", nil, git.CategoryOther},
		{"data.json", nil, git.CategoryOther},
		{"tool.lock", nil, git.CategoryOther},
		{"workspaces/AL-11/CONTRACT.md", nil, git.CategoryOther},
		{"workspaces/AL-11/agentic_metrics.json", nil, git.CategoryOther},
		{"workspaces/spec.md", nil, git.CategoryOther},

		// Test files
		{"internal/adapters/git/analyzer_test.go", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},
		{"tests/test_api.py", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},
		{"src/components/Card.test.tsx", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},
		{"src/utils/math.spec.ts", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},
		{"tests/unit/test_runner.py", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},
		{"__tests__/dashboard.test.js", &git.ClassificationConfig{TestPatterns: []string{"**"}}, git.CategoryTest},

		// Production files
		{"internal/domain/models.go", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"workspaces/file.go", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"cmd/agentic-lens/main.go", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"web/src/App.tsx", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"web/src/main.ts", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"web/index.html", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"web/src/style.css", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"Makefile", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"scripts/setup.sh", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},
		{"db/schema.sql", &git.ClassificationConfig{ProdPatterns: []string{"**"}}, git.CategoryProd},

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

	// Test branch inference purely in memory without mutating disk branches
	branchList := "  feat/AL-11-ci-fixture\n"
	ticketType := git.InferTicketTypeFromBranchList(branchList, "AL-11")
	if ticketType == nil || *ticketType != "feat" {
		t.Errorf("expected feat ticket type for AL-11, got %v", ticketType)
	}

	user := git.GetGitUser()
	if user == "" {
		t.Errorf("expected non-empty git user")
	}
	if strings.Contains(user, "@") {
		t.Errorf("expected git user not to leak full email domain, got: %s", user)
	}
	if user != "TestUser" {
		t.Errorf("expected git user 'TestUser' from isolated config, got: %s", user)
	}
}

func TestGitAnalyzer_AdversarialEdgeCases(t *testing.T) {
	malformedText := "\nnot_a_number\tnot_a_number\tfile.go\n10\n10\t20\n-\t-\tbinary.png\n\t\n15\t5\n-5\t-10\tnegative.go\n"
	parsed := git.ParseNumstatOutput(malformedText)
	if len(parsed) != 1 || parsed[0].Path != "negative.go" {
		t.Fatalf("expected only the three-field numeric row to parse, got %+v", parsed)
	}

	tempDir := t.TempDir()
	initTestGitRepo(t, tempDir, "main")
	file := filepath.Join(tempDir, "tickets", "AL-1", "SUMMARY.md")
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("# Summary\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "tickets/AL-1/SUMMARY.md")
	cmd.Dir = tempDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v (%s)", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "add fixture summary")
	cmd.Dir = tempDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v (%s)", err, out)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	minTs, maxTs, found := git.GetWorkspaceCommitTimeRange("tickets/AL-1")
	if !found || minTs <= 0 || maxTs < minTs {
		t.Fatalf("expected valid workspace commit range, got min=%d max=%d found=%v", minTs, maxTs, found)
	}
}

func TestLoadConfig_CentralizedIntegration(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	yamlContent := `
project_name: GitAnalyzerTest
ticket:
  prefixes:
    - prefix: AL
workspaces:
code_metrics:
  source_roots:
    - src
  include:
    - "custom/include/**"
  exclude:
    - "custom/exclude/**"
  test_patterns:
    - "*_test_custom.go"
  prod_patterns:
    - "*.custom_prod"
anomaly_thresholds:
  context_ceiling_warning_tokens: 160000
  context_ceiling_critical_tokens: 200000
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(yamlContent), 0644); err != nil {
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

	if got := git.ClassifyFile("checks/example_test_custom.go", cfg); got != git.CategoryTest {
		t.Fatalf("loaded test pattern: %s", got)
	}
	if got := git.ClassifyFile("runtime/example.custom_prod", cfg); got != git.CategoryProd {
		t.Fatalf("loaded production pattern: %s", got)
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

func TestValidateSourceRoots(t *testing.T) {
	expectedMsg := "code_metrics.source_roots is empty or unconfigured: configure 'code_metrics.source_roots' in .agenticlens-config.yaml"

	t.Run("nil slice fails with exact configuration prompt", func(t *testing.T) {
		err := git.ValidateSourceRoots(nil)
		if err == nil {
			t.Fatal("expected error for nil code_metrics.source_roots, got nil")
		}
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("empty slice fails with exact configuration prompt", func(t *testing.T) {
		err := git.ValidateSourceRoots([]string{})
		if err == nil {
			t.Fatal("expected error for empty code_metrics.source_roots, got nil")
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
			err := git.ValidateSourceRoots(tc)
			if err == nil {
				t.Fatalf("expected error for whitespace source_roots %v, got nil", tc)
			}
			if !strings.Contains(err.Error(), expectedMsg) {
				t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
			}
		}
	})

	t.Run("valid slice succeeds", func(t *testing.T) {
		err := git.ValidateSourceRoots([]string{"backend", "web"})
		if err != nil {
			t.Errorf("unexpected error for valid source_roots: %v", err)
		}
	})

	t.Run("slice with empty strings but at least one valid folder succeeds", func(t *testing.T) {
		err := git.ValidateSourceRoots([]string{"", "   ", "services", "\t"})
		if err != nil {
			t.Errorf("unexpected error for mixed source_roots: %v", err)
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
			cfg := tt.cfg
			if cfg == nil {
				cfg = &git.ClassificationConfig{SourceRoots: tt.srcFolders}
			}
			actual := git.IsMonitoredSourcePath(tt.path, cfg)
			if actual != tt.expected {
				t.Errorf("IsMonitoredSourcePath(%q, %v) = %v, expected %v", tt.path, tt.srcFolders, actual, tt.expected)
			}
		})
	}
}

func TestExplicitCodeRefDoesNotDependOnTicketOrBranchNaming(t *testing.T) {
	repo := t.TempDir()
	initTestGitRepo(t, repo, "baseline")
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(before) })
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("checkout", "-b", "research/adapter")
	if err := os.Mkdir("src", 0700); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		"src/app.go":               "package app\n",
		"src/app_test.go":          "package app\n",
		"src/testdata/input.jsonl": "{}\n",
		"runtime/service.custom":   "fixture\n",
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("commit", "-m", "implementation under a ticket-free branch")
	cfg := &git.ClassificationConfig{
		BaseBranch: "baseline", SourceRoots: []string{"src"},
		TestPatterns: []string{"**/*_test.go", "**/testdata/**"},
		ProdPatterns: []string{"src/**/*.go", "runtime/**"},
	}
	facts, err := git.ComputeCodeFactsAtRef("HEAD", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if facts.LinesAdded != 4 || facts.FilesChanged != 4 || facts.TestLinesAdded != 2 || facts.ProdLinesAdded != 2 {
		t.Fatalf("explicit ref did not produce expected classified facts: %+v", facts)
	}
	if _, err := git.ComputeCodeFactsAtRef("", cfg); err == nil {
		t.Fatal("empty code ref must fail instead of invoking ticket discovery")
	}
	if _, err := git.ComputeCodeFactsAtRef("missing-ref", cfg); err == nil {
		t.Fatal("invalid code ref silently accepted")
	}
}
