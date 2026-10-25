package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
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

// IsMonitoredSourcePath reports whether a file path falls within configured source folders or configured test/production inclusion patterns.
func IsMonitoredSourcePath(path string, srcFolders []string, cfg *ClassificationConfig) bool {
	cleanPath := filepath.ToSlash(strings.TrimSpace(path))
	cleanPath = strings.TrimPrefix(cleanPath, "./")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	if cleanPath == "" {
		return false
	}

	for _, folder := range srcFolders {
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
		for _, pattern := range cfg.TestPatterns {
			if matchPattern(pattern, cleanPath) {
				return true
			}
		}
		for _, pattern := range cfg.ProdPatterns {
			if matchPattern(pattern, cleanPath) {
				return true
			}
		}
	}

	return false
}

// GetTicketNumstat retrieves git diff numstat entries for a ticket using dynamic base branch and source folder filters.
func GetTicketNumstat(ticketID string, baseBranch string, srcFolders []string, cfg *ClassificationConfig) ([]NumstatEntry, error) {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return []NumstatEntry{}, nil
	}

	_, err := runGitCommand("rev-parse", "--show-toplevel")
	if err != nil {
		return []NumstatEntry{}, nil
	}

	if strings.TrimSpace(baseBranch) == "" {
		var projCfg *domain.ProjectConfig
		if cfg != nil && strings.TrimSpace(cfg.BaseBranch) != "" {
			projCfg = &domain.ProjectConfig{
				CodeMetrics: domain.CodeMetricsConfig{
					BaseBranch: strings.TrimSpace(cfg.BaseBranch),
				},
			}
		}
		resolved, err := ResolveBaseBranch("", projCfg)
		if err != nil {
			return nil, err
		}
		baseBranch = resolved
	}

	if len(srcFolders) == 0 && cfg != nil && len(cfg.SrcFolders) > 0 {
		srcFolders = cfg.SrcFolders
	}
	if err := ValidateSrcFolders(srcFolders); err != nil {
		return nil, err
	}

	// 1. Check if active branch is working on this ticket
	currBranch, err := runGitCommand("branch", "--show-current")
	if err == nil {
		currBranch = strings.TrimSpace(currBranch)
		if currBranch != "" && strings.Contains(strings.ToLower(currBranch), strings.ToLower(cleanTicket)) {
			// Try merge-base with base branch
			out, err := runGitCommand("diff", "--numstat", baseBranch+"...HEAD")
			if err == nil {
				parsed := ParseNumstatOutput(out)
				if len(parsed) > 0 {
					return parsed, nil
				}
			}

			// Try diff with base branch directly including uncommitted work
			outDirect, errDirect := runGitCommand("diff", "--numstat", baseBranch)
			if errDirect == nil {
				parsed := ParseNumstatOutput(outDirect)
				if len(parsed) > 0 {
					return parsed, nil
				}
			}
		}
	}

	// 2. Check merged commits on base branch where ticket appears in subject line
	logOut, err := runGitCommand("log", baseBranch, "--grep="+cleanTicket, "--format=%H\t%s", "-n", "25")
	if err == nil {
		lines := strings.Split(logOut, "\n")
		ticketRegex, reErr := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(cleanTicket) + `\b`)
		for _, line := range lines {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				sha := parts[0]
				subject := parts[1]
				if reErr == nil && ticketRegex.MatchString(subject) {
					out, err := runGitCommand("diff", "--numstat", sha+"^1", sha)
					if err == nil {
						parsed := ParseNumstatOutput(out)
						if len(parsed) > 0 {
							return parsed, nil
						}
					}
				}
			}
		}
	}

	// 3. Check any local or remote feature branch matching ticket ID
	branchOut, err := runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*")
	if err == nil {
		for _, b := range strings.Split(branchOut, "\n") {
			bClean := strings.TrimSpace(strings.TrimLeft(b, "*+ "))
			if bClean != "" && !strings.Contains(bClean, "HEAD") {
				out, err := runGitCommand("diff", "--numstat", baseBranch+"..."+bClean)
				if err == nil {
					parsed := ParseNumstatOutput(out)
					if len(parsed) > 0 {
						for _, r := range parsed {
							if IsMonitoredSourcePath(r.Path, srcFolders, cfg) {
								return parsed, nil
							}
						}
					}
				}
			}
		}
	}

	// 4. Check merged commits mentioning ticket in body
	if logOut != "" {
		metricsRegex := regexp.MustCompile(`(?i)\b(refactor|feat|chore)\((agentic/)?metrics\)`)
		for _, line := range strings.Split(logOut, "\n") {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				sha := parts[0]
				subject := parts[1]
				if metricsRegex.MatchString(subject) {
					continue
				}
				out, err := runGitCommand("diff", "--numstat", sha+"^1", sha)
				if err == nil {
					parsed := ParseNumstatOutput(out)
					if len(parsed) > 0 {
						for _, r := range parsed {
							if IsMonitoredSourcePath(r.Path, srcFolders, cfg) {
								return parsed, nil
							}
						}
					}
				}
			}
		}
	}

	// 5. Fallback for active ticket working tree: if working directory has changes and matches active branch
	if currBranch != "" && strings.Contains(strings.ToLower(currBranch), strings.ToLower(cleanTicket)) {
		headOut, err := runGitCommand("diff", "--numstat", "HEAD")
		if err == nil {
			parsed := ParseNumstatOutput(headOut)
			if len(parsed) > 0 {
				return parsed, nil
			}
		}
	}

	return []NumstatEntry{}, nil
}

// ComputeCodeFacts computes code changes, test/prod breakdown, and SDLC iterations from git diff.
func ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) (domain.CodeFacts, error) {
	return ComputeCodeFactsAtRef(ticketID, cfg, fallback, "")
}

// ComputeCodeFactsAtRef uses an explicitly selected branch or commit when ticket
// naming cannot identify the work. It measures committed changes from merge-base.
func ComputeCodeFactsAtRef(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts, codeRef string) (domain.CodeFacts, error) {
	// 1. Return safe defaults or fallback if not in a git repository
	_, gitErr := runGitCommand("rev-parse", "--show-toplevel")
	if gitErr != nil {
		if codeRef != "" && strings.TrimSpace(codeRef) != "HEAD" {
			return domain.CodeFacts{}, fmt.Errorf("cannot resolve code ref outside a Git repository")
		}
		if fallback != nil {
			return *fallback, nil
		}
		return domain.CodeFacts{
			LinesAdded:       0,
			LinesDeleted:     0,
			TestLinesAdded:   0,
			TestLinesDeleted: 0,
			ProdLinesAdded:   0,
			ProdLinesDeleted: 0,
			FilesChanged:     0,
			SDLCIterations:   1,
		}, nil
	}

	// 2. Ticket discovery needs an ID; an explicit code ref does not.
	if strings.TrimSpace(ticketID) == "" && strings.TrimSpace(codeRef) == "" {
		if fallback != nil {
			return *fallback, nil
		}
		return domain.CodeFacts{
			LinesAdded:       0,
			LinesDeleted:     0,
			TestLinesAdded:   0,
			TestLinesDeleted: 0,
			ProdLinesAdded:   0,
			ProdLinesDeleted: 0,
			FilesChanged:     0,
			SDLCIterations:   1,
		}, nil
	}

	if cfg == nil {
		loaded, err := LoadConfig("")
		if err == nil && loaded != nil {
			cfg = loaded
		}
	}

	if cfg == nil || len(cfg.SrcFolders) == 0 || ValidateSrcFolders(cfg.SrcFolders) != nil {
		if fallback != nil {
			return *fallback, nil
		}
		return domain.CodeFacts{
			LinesAdded:       0,
			LinesDeleted:     0,
			TestLinesAdded:   0,
			TestLinesDeleted: 0,
			ProdLinesAdded:   0,
			ProdLinesDeleted: 0,
			FilesChanged:     0,
			SDLCIterations:   1,
		}, nil
	}

	projCfg := &domain.ProjectConfig{
		CodeMetrics: domain.CodeMetricsConfig{
			BaseBranch: cfg.BaseBranch,
		},
	}
	baseBranch, err := ResolveBaseBranch("", projCfg)
	if err != nil {
		return domain.CodeFacts{}, err
	}

	var numstat []NumstatEntry
	if codeRef != "" {
		// Resolve to a commit hash before constructing the revision range.
		resolved, resolveErr := runGitCommand("rev-parse", "--verify", "--end-of-options", codeRef+"^{commit}")
		if resolveErr != nil {
			return domain.CodeFacts{}, fmt.Errorf("invalid code ref %q", codeRef)
		}
		out, diffErr := runGitCommand("diff", "--numstat", baseBranch+"..."+strings.TrimSpace(resolved), "--")
		if diffErr != nil {
			return domain.CodeFacts{}, fmt.Errorf("comparing code ref %q with %q: %w", codeRef, baseBranch, diffErr)
		}
		numstat = ParseNumstatOutput(out)
	} else {
		numstat, err = GetTicketNumstat(ticketID, baseBranch, cfg.SrcFolders, cfg)
	}
	if err != nil {
		return domain.CodeFacts{}, err
	}
	if len(numstat) == 0 && fallback != nil && codeRef == "" {
		return *fallback, nil
	}

	var linesAdded, linesDeleted int64
	var testLinesAdded, testLinesDeleted int64
	var prodLinesAdded, prodLinesDeleted int64
	var filesChanged int64

	for _, entry := range numstat {
		if !IsMonitoredSourcePath(entry.Path, cfg.SrcFolders, cfg) {
			continue
		}
		category := ClassifyFile(entry.Path, cfg)
		if category == CategoryIgnored {
			continue
		}

		linesAdded += entry.Added
		linesDeleted += entry.Deleted
		filesChanged++

		switch category {
		case CategoryTest:
			testLinesAdded += entry.Added
			testLinesDeleted += entry.Deleted
		case CategoryProd:
			prodLinesAdded += entry.Added
			prodLinesDeleted += entry.Deleted
		}
	}

	sdlcIters := int64(1)
	if fallback != nil && fallback.SDLCIterations > 0 {
		sdlcIters = fallback.SDLCIterations
	}

	return domain.CodeFacts{
		LinesAdded:       linesAdded,
		LinesDeleted:     linesDeleted,
		TestLinesAdded:   testLinesAdded,
		TestLinesDeleted: testLinesDeleted,
		ProdLinesAdded:   prodLinesAdded,
		ProdLinesDeleted: prodLinesDeleted,
		FilesChanged:     filesChanged,
		SDLCIterations:   sdlcIters,
	}, nil
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
