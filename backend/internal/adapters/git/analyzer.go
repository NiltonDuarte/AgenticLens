package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"agenticlens/internal/domain"
)

// NumstatEntry retains one NUL-delimited Git numstat record. Rename records
// belong to their destination. Binary records have no line measurements.
type NumstatEntry struct {
	Added   int64
	Deleted int64
	Path    string
	Binary  bool
}

// ParseNumstatOutput parses git diff --numstat -z without interpreting or
// trimming filenames. Malformed records fail the measurement instead of being
// silently omitted.
func ParseNumstatOutput(stdout string) ([]NumstatEntry, error) {
	entries := make([]NumstatEntry, 0)
	if stdout == "" {
		return entries, nil
	}
	if !strings.HasSuffix(stdout, "\x00") {
		return nil, fmt.Errorf("Git numstat output is not NUL-terminated")
	}
	records := strings.Split(stdout[:len(stdout)-1], "\x00")
	for i := 0; i < len(records); i++ {
		fields := strings.SplitN(records[i], "\t", 3)
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid Git numstat record %d: expected line counts and a path", i+1)
		}
		filePath := fields[2]
		if filePath == "" {
			if i+2 >= len(records) || records[i+1] == "" || records[i+2] == "" {
				return nil, fmt.Errorf("invalid Git numstat rename record %d: missing source or destination", i+1)
			}
			filePath = records[i+2]
			i += 2
		}
		entry := NumstatEntry{Path: filePath}
		if fields[0] == "-" && fields[1] == "-" {
			entry.Binary = true
		} else {
			var err error
			entry.Added, err = parseLineCount(fields[0])
			if err != nil {
				return nil, fmt.Errorf("invalid added line count for %q: %w", filePath, err)
			}
			entry.Deleted, err = parseLineCount(fields[1])
			if err != nil {
				return nil, fmt.Errorf("invalid deleted line count for %q: %w", filePath, err)
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func parseLineCount(value string) (int64, error) {
	if value == "" {
		return 0, fmt.Errorf("expected a nonnegative decimal integer")
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("expected a nonnegative decimal integer, got %q", value)
		}
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func runGitCommandInDir(dir string, args ...string) (string, error) {
	return runGitCommandWithEnvironment(dir, nil, args...)
}

func runMetricGitCommandInDir(dir string, args ...string) (string, error) {
	return runGitCommandWithEnvironment(dir, historyGitEnvironment(os.Environ()), args...)
}

func runGitCommandWithEnvironment(dir string, environment []string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if environment != nil {
		cmd.Env = environment
	}
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("git %s: %s: %w", args[0], detail, err)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}
	return stdout.String(), nil
}

func runGitCommand(args ...string) (string, error) {
	return runGitCommandInDir("", args...)
}

func resolveCommit(repoRoot, ref string) (string, error) {
	resolved, err := runMetricGitCommandInDir(repoRoot, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resolved), nil
}

func configuredBase(repoRoot string, cfg *domain.ProjectConfig) (string, string, error) {
	if cfg == nil || strings.TrimSpace(cfg.CodeMetrics.BaseBranch) == "" {
		return "", "", fmt.Errorf("code_metrics.base_branch is required: configure an explicit comparison ref in shared project configuration")
	}
	branch := strings.TrimSpace(cfg.CodeMetrics.BaseBranch)
	commit, err := resolveCommit(repoRoot, branch)
	if err != nil {
		return "", "", fmt.Errorf("code_metrics.base_branch %q does not resolve to a commit in the target repository: %w", branch, err)
	}
	return branch, commit, nil
}

// ResolveBaseBranch validates only the explicitly configured comparison ref.
// It neither discovers other branches nor fetches unavailable refs.
func ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error) {
	branch, _, err := configuredBase(repoRoot, cfg)
	return branch, err
}

// ComputeCodeFactsAtRef measures an explicitly selected committed target in the
// invocation repository. Callers selecting another project use the directory
// variant so repository ownership never depends on global working-directory changes.
func ComputeCodeFactsAtRef(codeRef string, cfg *ClassificationConfig) (domain.CodeFacts, error) {
	return ComputeCodeFactsAtRefInDir("", codeRef, cfg)
}

// ComputeCodeFactsAtRefInDir measures the target's net committed contribution
// since its merge base with the configured comparison ref. Only explicitly
// classified production/test files contribute to the totals.
func ComputeCodeFactsAtRefInDir(repoRoot, codeRef string, cfg *ClassificationConfig) (domain.CodeFacts, error) {
	ref := strings.TrimSpace(codeRef)
	if ref == "" {
		return domain.CodeFacts{}, fmt.Errorf("code ref is required; pass --code-ref with a branch, tag or commit")
	}
	if _, err := runMetricGitCommandInDir(repoRoot, "rev-parse", "--show-toplevel"); err != nil {
		return domain.CodeFacts{}, fmt.Errorf("cannot measure code ref %q in the target Git repository: %w", ref, err)
	}
	if cfg == nil {
		loaded, err := LoadConfig(repoRoot)
		if err != nil {
			return domain.CodeFacts{}, fmt.Errorf("loading Git classification configuration: %w", err)
		}
		cfg = loaded
	}
	if err := ValidateClassificationConfig(cfg); err != nil {
		return domain.CodeFacts{}, err
	}
	project := &domain.ProjectConfig{CodeMetrics: domain.CodeMetricsConfig{BaseBranch: cfg.BaseBranch}}
	baseName, baseCommit, err := configuredBase(repoRoot, project)
	if err != nil {
		return domain.CodeFacts{}, err
	}
	targetCommit, err := resolveCommit(repoRoot, ref)
	if err != nil {
		return domain.CodeFacts{}, fmt.Errorf("code ref %q does not resolve to a commit in the target repository: %w", ref, err)
	}
	mergeBase, err := runMetricGitCommandInDir(repoRoot, "merge-base", baseCommit, targetCommit)
	if err != nil {
		return domain.CodeFacts{}, fmt.Errorf("finding merge base of code ref %q and configured base %q: %w", ref, baseName, err)
	}
	out, err := runMetricGitCommandInDir(repoRoot, "diff", "--numstat", "-z", "--no-ext-diff", "--no-textconv", "--find-renames", strings.TrimSpace(mergeBase), targetCommit, "--")
	if err != nil {
		return domain.CodeFacts{}, fmt.Errorf("comparing code ref %q with configured base %q: %w", ref, baseName, err)
	}
	entries, err := ParseNumstatOutput(out)
	if err != nil {
		return domain.CodeFacts{}, err
	}
	var facts domain.CodeFacts
	for _, entry := range entries {
		category, err := classifyFile(entry.Path, cfg)
		if err != nil {
			return domain.CodeFacts{}, err
		}
		if category == CategoryOther {
			continue
		}
		facts.FilesChanged++
		if entry.Binary {
			continue
		}
		facts.LinesAdded += entry.Added
		facts.LinesDeleted += entry.Deleted
		switch category {
		case CategoryTest:
			facts.TestLinesAdded += entry.Added
			facts.TestLinesDeleted += entry.Deleted
		case CategoryProd:
			facts.ProdLinesAdded += entry.Added
			facts.ProdLinesDeleted += entry.Deleted
		}
	}
	return facts, nil
}

// GetWorkspaceCommitTimeRange returns the minimum and maximum commit timestamps (in Unix seconds) for commits touching workspaceDir.
func GetWorkspaceCommitTimeRange(workspaceDir string) (minUnixSec int64, maxUnixSec int64, found bool) {
	cleanDir := strings.TrimSpace(workspaceDir)
	if cleanDir == "" {
		return 0, 0, false
	}
	cleanDir = strings.TrimPrefix(cleanDir, "./")
	cleanDir = strings.TrimPrefix(cleanDir, "/")
	output, err := runGitCommand("log", "--format=%ct", "--", ":(top)"+cleanDir)
	if err != nil {
		return 0, 0, false
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var minTs, maxTs int64
	count := 0
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "" {
			continue
		}
		ts, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil || ts <= 0 {
			continue
		}
		if count == 0 {
			minTs = ts
			maxTs = ts
		} else {
			if ts < minTs {
				minTs = ts
			}
			if ts > maxTs {
				maxTs = ts
			}
		}
		count++
	}
	if count == 0 {
		return 0, 0, false
	}
	return minTs, maxTs, true
}
