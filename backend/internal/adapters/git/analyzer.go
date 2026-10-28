package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"agenticlens/internal/domain"
)

// NumstatEntry represents lines added, lines deleted, and file path from git diff --numstat.
type NumstatEntry struct {
	Added   int64
	Deleted int64
	Path    string
}

// ParseNumstatOutput parses raw git diff --numstat stdout into a list of NumstatEntry.
func ParseNumstatOutput(stdout string) []NumstatEntry {
	var entries []NumstatEntry
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) >= 3 {
			added, err1 := strconv.ParseInt(parts[0], 10, 64)
			deleted, err2 := strconv.ParseInt(parts[1], 10, 64)
			if err1 == nil && err2 == nil {
				// Rejoin path in case it had spaces
				filePath := strings.Join(parts[2:], " ")
				entries = append(entries, NumstatEntry{
					Added:   added,
					Deleted: deleted,
					Path:    filePath,
				})
			}
		}
	}
	return entries
}

func runGitCommandInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

func runGitCommand(args ...string) (string, error) {
	return runGitCommandInDir("", args...)
}

// ResolveBaseBranch resolves the upstream comparison branch using a 4-tier strategy.
func ResolveBaseBranch(repoRoot string, cfg *domain.ProjectConfig) (string, error) {
	// Tier 1: Explicit Configuration Tier
	if cfg != nil {
		configured := strings.TrimSpace(cfg.CodeMetrics.BaseBranch)
		if configured != "" {
			_, err := runGitCommandInDir(repoRoot, "rev-parse", "--verify", "--quiet", configured+"^{commit}")
			if err == nil {
				return configured, nil
			}
			return "", fmt.Errorf("configured base branch %q not found: cannot resolve upstream base branch in repository", configured)
		}
	}

	// Tier 2: Symbolic Reference Remote Tracking Tier
	symOut, err := runGitCommandInDir(repoRoot, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(symOut)
		ref = strings.TrimPrefix(ref, "refs/remotes/")
		if ref != "" {
			_, err := runGitCommandInDir(repoRoot, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
			if err == nil {
				return ref, nil
			}
		}
	}

	// Tier 3: Candidate Probing Tier
	candidates := []string{
		path.Join("origin", "main"),
		"main",
		path.Join("origin", "master"),
		"master",
	}
	for _, candidate := range candidates {
		_, err := runGitCommandInDir(repoRoot, "rev-parse", "--verify", "--quiet", candidate+"^{commit}")
		if err == nil {
			return candidate, nil
		}
	}

	// Tier 4: Explicit Failure Invariant
	return "", fmt.Errorf("unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml")
}

// ValidateSrcFolders verifies that workspaces.src_folders contains at least one non-empty source directory.
func ValidateSrcFolders(srcFolders []string) error {
	if len(srcFolders) == 0 {
		return fmt.Errorf("workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml")
	}
	hasNonEmpty := false
	for _, folder := range srcFolders {
		if strings.TrimSpace(folder) != "" {
			hasNonEmpty = true
			break
		}
	}
	if !hasNonEmpty {
		return fmt.Errorf("workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml")
	}
	return nil
}

// ValidateSourceRoots verifies that code_metrics.source_roots contains at least one non-empty source directory.
func ValidateSourceRoots(sourceRoots []string) error {
	if len(sourceRoots) == 0 {
		return fmt.Errorf("code_metrics.source_roots is empty or unconfigured: configure 'code_metrics.source_roots' in .agenticlens-config.yaml")
	}
	hasNonEmpty := false
	for _, folder := range sourceRoots {
		if strings.TrimSpace(folder) != "" {
			hasNonEmpty = true
			break
		}
	}
	if !hasNonEmpty {
		return fmt.Errorf("code_metrics.source_roots is empty or unconfigured: configure 'code_metrics.source_roots' in .agenticlens-config.yaml")
	}
	return nil
}

// IsMonitoredSourcePath reports whether a file path falls within configured source folders or configured test/production inclusion patterns.
func IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool {
	cleanPath := filepath.ToSlash(strings.TrimSpace(path))
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	if cleanPath == "" {
		return false
	}

	folders := srcFolders
	if len(folders) == 0 && cfg != nil {
		if len(cfg.SourceRoots) > 0 {
			folders = cfg.SourceRoots
		} else {
			folders = cfg.SrcFolders
		}
	}

	for _, folder := range folders {
		normFolder := filepath.ToSlash(strings.TrimSpace(folder))
		normFolder = strings.Trim(normFolder, "/")
		if normFolder == "" {
			continue
		}
		if cleanPath == normFolder || strings.HasPrefix(cleanPath, normFolder+"/") {
			return true
		}
	}

	if cfg != nil {
		for _, pattern := range cfg.Include {
			if matchPattern(pattern, cleanPath) {
				return true
			}
		}
		for _, pattern := range cfg.ProdPatterns {
			if matchPattern(pattern, cleanPath) {
				return true
			}
		}
		for _, pattern := range cfg.TestPatterns {
			if matchPattern(pattern, cleanPath) {
				return true
			}
		}
	}

	return false
}

// ComputeCodeFactsAtRef measures committed code changes at an explicit Git ref.
// Ticket names and branch-name heuristics are intentionally not used.
func ComputeCodeFactsAtRef(codeRef string, cfg *ClassificationConfig) (domain.CodeFacts, error) {
	ref := strings.TrimSpace(codeRef)
	if ref == "" {
		return domain.CodeFacts{}, fmt.Errorf("code ref is required; pass --code-ref with a branch or commit")
	}
	if _, err := runGitCommand("rev-parse", "--show-toplevel"); err != nil {
		if ref == "HEAD" {
			return domain.CodeFacts{SDLCIterations: 1}, nil
		}
		return domain.CodeFacts{}, fmt.Errorf("cannot resolve code ref outside a Git repository")
	}
	if cfg == nil {
		loaded, err := LoadConfig("")
		if err == nil {
			cfg = loaded
		}
	}
	var srcs []string
	if cfg != nil {
		if len(cfg.SourceRoots) > 0 {
			srcs = cfg.SourceRoots
		} else {
			srcs = cfg.SrcFolders
		}
	}
	if cfg == nil || len(srcs) == 0 || ValidateSrcFolders(srcs) != nil {
		return domain.CodeFacts{}, fmt.Errorf("code metrics require valid configured source folders")
	}

	projectConfig := &domain.ProjectConfig{CodeMetrics: domain.CodeMetricsConfig{BaseBranch: cfg.BaseBranch}}
	baseBranch, err := ResolveBaseBranch("", projectConfig)
	if err != nil {
		return domain.CodeFacts{}, err
	}
	resolved, err := runGitCommand("rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
	if err != nil {
		return domain.CodeFacts{}, fmt.Errorf("invalid code ref %q", ref)
	}
	out, err := runGitCommand("diff", "--numstat", baseBranch+"..."+strings.TrimSpace(resolved), "--")
	if err != nil {
		return domain.CodeFacts{}, fmt.Errorf("comparing code ref %q with %q: %w", ref, baseBranch, err)
	}

	var facts domain.CodeFacts
	facts.SDLCIterations = 1
	for _, entry := range ParseNumstatOutput(out) {
		if !IsMonitoredSourcePath(entry.Path, srcs, cfg) {
			continue
		}
		category := ClassifyFile(entry.Path, cfg)
		if category == CategoryIgnored {
			continue
		}
		facts.LinesAdded += entry.Added
		facts.LinesDeleted += entry.Deleted
		facts.FilesChanged++
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
