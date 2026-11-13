package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func TestParseNumstatOutputPreservesPathsAndRenames(t *testing.T) {
	t.Parallel()
	raw := "12\t4\tinternal/analyzer.go\x00" +
		"3\t1\t path with\tspaces\nand newline.go \x00" +
		"-\t-\timage.png\x00" +
		"2\t1\t\x00old/name.go\x00tests/new\tname.go\x00"
	got, err := git.ParseNumstatOutput(raw)
	want := []git.NumstatEntry{
		{Added: 12, Deleted: 4, Path: "internal/analyzer.go"},
		{Added: 3, Deleted: 1, Path: " path with\tspaces\nand newline.go "},
		{Path: "image.png", Binary: true},
		{Added: 2, Deleted: 1, Path: "tests/new\tname.go"},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseNumstatOutput = %#v, %v; want %#v", got, err, want)
	}
	if entries, err := git.ParseNumstatOutput(""); err != nil || len(entries) != 0 {
		t.Fatalf("successful empty diff: %#v, %v", entries, err)
	}
}

func TestParseNumstatOutputRejectsMalformedMeasurements(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"1\t2\tfile.go", "\x00", "1\t2\x00", "x\t1\tfile.go\x00", "-1\t2\tfile.go\x00",
		"1\t-\tfile.go\x00", "-\t1\tfile.go\x00", "9223372036854775808\t1\tfile.go\x00",
		"1\t2\t\x00", "1\t2\t\x00old.go\x00", "1\t2\t\x00\x00new.go\x00", "1\t2\t\x00old.go\x00\x00",
	} {
		if entries, err := git.ParseNumstatOutput(raw); err == nil {
			t.Errorf("malformed %q accepted as %#v", raw, entries)
		}
	}
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, dir, name string, contents []byte) {
	t.Helper()
	full := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, contents, 0600); err != nil {
		t.Fatal(err)
	}
}

func initTestGitRepo(t *testing.T, dir, branch string) {
	t.Helper()
	gitRun(t, dir, "init", "-b", branch)
	gitRun(t, dir, "config", "user.name", "TestUser")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	writeFile(t, dir, "README.md", []byte("# Test Repo\n"))
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "initial commit")
}

func commitFiles(t *testing.T, repo, message string) {
	t.Helper()
	gitRun(t, repo, "add", "--all")
	gitRun(t, repo, "commit", "-m", message)
}

func withWorkingDirectory(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func TestGitUserAndWorkspaceCommitRange(t *testing.T) {
	repo := t.TempDir()
	initTestGitRepo(t, repo, "main")
	writeFile(t, repo, "tickets/CASE-1/SUMMARY.md", []byte("# Summary\n"))
	commitFiles(t, repo, "ticket summary")
	withWorkingDirectory(t, repo)
	if user := git.GetGitUser(); user != "TestUser" || strings.Contains(user, "@") {
		t.Fatalf("Git user should use isolated configured name: %q", user)
	}
	other := t.TempDir()
	initTestGitRepo(t, other, "main")
	gitRun(t, other, "config", "user.name", "TargetProjectUser")
	if user := git.GetGitUserInDir(other); user != "TargetProjectUser" {
		t.Fatalf("foreign target borrowed invocation-project user: %q", user)
	}
	min, max, found := git.GetWorkspaceCommitTimeRange("tickets/CASE-1")
	if !found || min <= 0 || max < min {
		t.Fatalf("workspace commit range: min=%d max=%d found=%v", min, max, found)
	}
	if _, _, found := git.GetWorkspaceCommitTimeRange("tickets/CASE-2"); found {
		t.Fatal("uncommitted workspace must not acquire a commit range")
	}
}

func TestResolveBaseBranchRequiresOnlyConfiguredRef(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	initTestGitRepo(t, repo, "main")
	gitRun(t, repo, "branch", "release-1.0")
	gitRun(t, repo, "branch", "master")
	gitRun(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitRun(t, repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	for _, tc := range []struct {
		base string
		want string
	}{
		{"release-1.0", "release-1.0"},
		{"  release-1.0  ", "release-1.0"},
		{"origin/main", "origin/main"},
		{"", ""},
		{"  ", ""},
		{"missing-branch", ""},
		{"--all", ""},
	} {
		cfg := &domain.ProjectConfig{CodeMetrics: domain.CodeMetricsConfig{BaseBranch: tc.base}}
		got, err := git.ResolveBaseBranch(repo, cfg)
		if tc.want == "" {
			if err == nil || got != "" || !strings.Contains(err.Error(), "code_metrics.base_branch") {
				t.Errorf("base %q used another available ref: %q, %v", tc.base, got, err)
			}
		} else if err != nil || got != tc.want {
			t.Errorf("base %q: %q, %v; want %q", tc.base, got, err, tc.want)
		}
	}
	if got, err := git.ResolveBaseBranch(repo, nil); err == nil || got != "" {
		t.Fatalf("missing configuration selected another branch: %q, %v", got, err)
	}
}

func TestComputeCodeFactsUsesNetContributionAcrossBaseMerges(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	initTestGitRepo(t, repo, "baseline")
	writeFile(t, repo, "src/main.go", []byte("one\ntwo\nthree\n"))
	commitFiles(t, repo, "baseline source")
	gitRun(t, repo, "checkout", "-b", "research/adapter")
	writeFile(t, repo, "src/main.go", []byte("one\nthree\n"))
	writeFile(t, repo, "src/feature.go", []byte("a\nb\nc\n"))
	writeFile(t, repo, "src/feature_test.go", []byte("test a\ntest b\n"))
	writeFile(t, repo, "src/reverted.go", []byte("temporary\n"))
	writeFile(t, repo, "docs/feature.md", []byte("documentation\n"))
	commitFiles(t, repo, "branch code, tests and temporary change")
	if err := os.Remove(filepath.Join(repo, "src/reverted.go")); err != nil {
		t.Fatal(err)
	}
	commitFiles(t, repo, "revert temporary code")
	historical := gitRun(t, repo, "rev-parse", "HEAD")
	gitRun(t, repo, "checkout", "baseline")
	writeFile(t, repo, "src/upstream.go", []byte("upstream a\nupstream b\nupstream c\nupstream d\n"))
	commitFiles(t, repo, "incoming base source")
	gitRun(t, repo, "checkout", "research/adapter")
	gitRun(t, repo, "merge", "baseline", "--no-edit")
	cfg := classification([]string{"src/**/*.go"}, []string{"**/*_test.go"}, []string{"**/*_test.go"}, nil)
	want := domain.CodeFacts{LinesAdded: 5, LinesDeleted: 1, FilesChanged: 3, ProdLinesAdded: 3, ProdLinesDeleted: 1, TestLinesAdded: 2}
	assertFacts := func(ref string) {
		t.Helper()
		got, err := git.ComputeCodeFactsAtRefInDir(repo, ref, cfg)
		if err != nil || got != want {
			t.Fatalf("net contribution at %q: %+v, %v; want %+v", ref, got, err, want)
		}
	}
	assertFacts("HEAD")
	assertFacts(historical)
	writeFile(t, repo, "src/feature.go", []byte("uncommitted change\n"))
	gitRun(t, repo, "add", "src/feature.go")
	assertFacts("HEAD")
	gitRun(t, repo, "reset", "--hard", "HEAD")
	gitRun(t, repo, "checkout", "baseline")
	writeFile(t, repo, "src/next.go", []byte("later upstream\n"))
	commitFiles(t, repo, "advance base again")
	gitRun(t, repo, "checkout", "research/adapter")
	assertFacts("HEAD")
	gitRun(t, repo, "merge", "baseline", "--no-edit")
	assertFacts("HEAD")
	if _, err := git.ComputeCodeFactsAtRefInDir(repo, "", cfg); err == nil {
		t.Fatal("empty explicit target accepted")
	}
	if _, err := git.ComputeCodeFactsAtRefInDir(repo, "missing-ref", cfg); err == nil {
		t.Fatal("unavailable explicit target accepted")
	}
}

func TestCodeFactsPreserveUnusualPathsAndDestinationOwnedRename(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	initTestGitRepo(t, repo, "baseline")
	writeFile(t, repo, "src/source.go", []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n"))
	commitFiles(t, repo, "baseline source")
	gitRun(t, repo, "checkout", "-b", "topic")
	if err := os.MkdirAll(filepath.Join(repo, "tests"), 0700); err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "mv", "src/source.go", "tests/moved\tcode.go")
	writeFile(t, repo, "tests/moved\tcode.go", []byte("one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\neleven\n"))
	for _, name := range []string{"src/tab\tname.go", "src/line\nname.go", "src/ spaced name.go "} {
		writeFile(t, repo, name, []byte("source\n"))
	}
	writeFile(t, repo, "src/binary.bin", []byte{0, 1, 2, 0, 3})
	writeFile(t, repo, "docs/ignored.go", []byte("ignored\n"))
	commitFiles(t, repo, "rename into tests and add unusual source names")
	cfg := classification([]string{"src/**"}, nil, []string{"tests/**"}, nil)
	facts, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", cfg)
	want := domain.CodeFacts{LinesAdded: 4, FilesChanged: 5, ProdLinesAdded: 3, TestLinesAdded: 1}
	if err != nil || facts != want {
		t.Fatalf("lossless destination-owned diff: %+v, %v; want %+v", facts, err, want)
	}
	cfg.Tests.Include = []string{"**/*.go"}
	if facts, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", cfg); err == nil || facts != (domain.CodeFacts{}) || !strings.Contains(err.Error(), "classification conflict") {
		t.Fatalf("overlap must fail entire measurement: %+v, %v", facts, err)
	}
}

func TestCodeFactsRequiredPolicyAndRepositoryFailures(t *testing.T) {
	t.Parallel()
	cfg := classification([]string{"**/*.go"}, nil, nil, nil)
	if facts, err := git.ComputeCodeFactsAtRefInDir(t.TempDir(), "HEAD", cfg); err == nil || facts != (domain.CodeFacts{}) {
		t.Fatalf("outside Git fabricated a successful measurement: %+v, %v", facts, err)
	}
	repo := t.TempDir()
	initTestGitRepo(t, repo, "main")
	if facts, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", nil); err == nil || facts != (domain.CodeFacts{}) {
		t.Fatalf("missing selected config fabricated measurement: %+v, %v", facts, err)
	}
	cfg.BaseBranch = ""
	if _, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", cfg); err == nil {
		t.Fatal("missing base silently used available main")
	}
	cfg.BaseBranch = "main"
	cfg.Tests.Exclude = nil
	if _, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", cfg); err == nil || !strings.Contains(err.Error(), "tests.exclude") {
		t.Fatalf("missing required list accepted for empty diff: %v", err)
	}
	cfg.Tests.Exclude = []string{}
	cfg.Production.Include = []string{"[invalid"}
	if _, err := git.ComputeCodeFactsAtRefInDir(repo, "HEAD", cfg); err == nil || !strings.Contains(err.Error(), "production.include") {
		t.Fatalf("invalid glob accepted for empty diff: %v", err)
	}
}

func TestCodeFactsUsesSelectedRepositoryDespiteInheritedGitOverrides(t *testing.T) {
	selected := t.TempDir()
	other := t.TempDir()
	for _, repo := range []string{selected, other} {
		initTestGitRepo(t, repo, "baseline")
		gitRun(t, repo, "checkout", "-b", "topic")
	}
	writeFile(t, selected, "src/selected.go", []byte("selected\n"))
	commitFiles(t, selected, "selected contribution")
	writeFile(t, other, "src/unrelated.go", []byte("one\ntwo\nthree\n"))
	commitFiles(t, other, "unrelated contribution")
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	t.Setenv("GIT_INDEX_FILE", filepath.Join(other, ".git", "index"))
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.bare")
	t.Setenv("GIT_CONFIG_VALUE_0", "true")
	cfg := classification([]string{"src/**"}, nil, nil, nil)
	facts, err := git.ComputeCodeFactsAtRefInDir(selected, "HEAD", cfg)
	want := domain.CodeFacts{LinesAdded: 1, FilesChanged: 1, ProdLinesAdded: 1}
	if err != nil || facts != want {
		t.Fatalf("inherited repository settings redirected measurement: %+v, %v; want %+v", facts, err, want)
	}
	if facts, err := git.ComputeCodeFactsAtRefInDir(t.TempDir(), "HEAD", cfg); err == nil || facts != (domain.CodeFacts{}) {
		t.Fatalf("unavailable selected repository used inherited repository: %+v, %v", facts, err)
	}
}

func TestLoadConfigCentralizedClassification(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	const content = "project_name: GitAnalyzerTest\nagenticlens:\n  version: v1.2.3\nticket:\n  prefixes:\n    - prefix: CASE-\n      description: Isolated test tickets\n  case_insensitive: false\nworkspaces:\n  roots:\n    - path: tickets\n      discovery:\n        max_depth: 8\n        max_directories: 300\n  artifacts: []\nrole_mappings: {}\nrole_order: []\ngeneration:\n  metrics:\n    skills: false\n    delegations: false\n    anomalies: false\n  feedback:\n    provisional_threshold_seconds: 0\n  summary:\n    enabled: false\ncode_metrics:\n  base_branch: missing-but-not-needed-for-classification\n  production:\n    include: [\"custom/include/**\", \"*.custom_prod\"]\n    exclude: [\"custom/include/excluded/**\"]\n  tests:\n    include: [\"*_test_custom.go\"]\n    exclude: []\ndatabase:\n  busy_timeout_ms: 17\nlive:\n  active_threshold_seconds: 60\n  poll_interval_ms: 100\nanomaly_thresholds:\n  stall_threshold_seconds: 15\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.7\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n  consecutive_errors_streak_warning: 2\n  consecutive_errors_streak_critical: 4\n  tool_thrashing_min_failures: 3\n  tool_thrashing_critical_failures: 5\n  turn_usage_warning_tokens: 1000\n  turn_usage_critical_tokens: 2000\n  context_growth_warning_tokens: 500\n  context_growth_critical_tokens: 1000\n"
	writeFile(t, repo, ".agenticlens-config.yaml", []byte(content))
	cfg, err := git.LoadConfig(repo)
	if err != nil {
		t.Fatalf("load explicit current policy: %v", err)
	}
	if cfg.BaseBranch != "missing-but-not-needed-for-classification" || !reflect.DeepEqual(cfg.Production.Include, []string{"custom/include/**", "*.custom_prod"}) || !reflect.DeepEqual(cfg.Production.Exclude, []string{"custom/include/excluded/**"}) || !reflect.DeepEqual(cfg.Tests.Include, []string{"*_test_custom.go"}) || cfg.Tests.Exclude == nil || len(cfg.Tests.Exclude) != 0 {
		t.Fatalf("loaded policy changed explicit lists/base: %+v", cfg)
	}
	for _, tc := range []struct {
		name string
		want git.FileCategory
	}{
		{"checks/example_test_custom.go", git.CategoryTest},
		{"runtime/example.custom_prod", git.CategoryProd},
		{"custom/include/file.xyz", git.CategoryProd},
		{"custom/include/excluded/file.xyz", git.CategoryOther},
	} {
		if got, err := git.ClassifyFile(tc.name, cfg); err != nil || got != tc.want {
			t.Errorf("loaded %q: %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}
}
