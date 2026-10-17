package git

import (
	"bytes"
	"os/exec"
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

func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// GetTicketNumstat retrieves git diff numstat entries for a ticket.
func GetTicketNumstat(ticketID string) []NumstatEntry {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return []NumstatEntry{}
	}

	excludeArgs := []string{
		"--",
		":(exclude)*.md",
		":(exclude)*.json",
		":(exclude)*.lock",
	}

	// 1. Check if active branch is working on this ticket
	currBranch, err := runGitCommand("branch", "--show-current")
	if err == nil {
		currBranch = strings.TrimSpace(currBranch)
		if currBranch != "" && strings.Contains(strings.ToLower(currBranch), strings.ToLower(cleanTicket)) {
			// Try merge-base with origin/main
			args := append([]string{"diff", "--numstat", "origin/main...HEAD"}, excludeArgs...)
			out, err := runGitCommand(args...)
			if err == nil {
				parsed := ParseNumstatOutput(out)
				if len(parsed) > 0 {
					return parsed
				}
			}

			// Try diff with origin/main directly including uncommitted work
			argsDirect := append([]string{"diff", "--numstat", "origin/main"}, excludeArgs...)
			outDirect, errDirect := runGitCommand(argsDirect...)
			if errDirect == nil {
				parsed := ParseNumstatOutput(outDirect)
				if len(parsed) > 0 {
					return parsed
				}
			}
		}
	}

	// 2. Check merged commits on origin/main where ticket appears in subject line
	logOut, err := runGitCommand("log", "origin/main", "--grep="+cleanTicket, "--format=%H\t%s", "-n", "25")
	if err == nil {
		lines := strings.Split(logOut, "\n")
		ticketRegex, reErr := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(cleanTicket) + `\b`)
		for _, line := range lines {
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) == 2 {
				sha := parts[0]
				subject := parts[1]
				if reErr == nil && ticketRegex.MatchString(subject) {
					diffArgs := append([]string{"diff", "--numstat", sha + "^1", sha}, excludeArgs...)
					out, err := runGitCommand(diffArgs...)
					if err == nil {
						parsed := ParseNumstatOutput(out)
						if len(parsed) > 0 {
							return parsed
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
				diffArgs := append([]string{"diff", "--numstat", "origin/main..." + bClean}, excludeArgs...)
				out, err := runGitCommand(diffArgs...)
				if err == nil {
					parsed := ParseNumstatOutput(out)
					if len(parsed) > 0 {
						for _, r := range parsed {
							if strings.HasPrefix(r.Path, "src/") ||
								strings.HasPrefix(r.Path, "internal/") ||
								strings.HasPrefix(r.Path, "cmd/") ||
								strings.HasPrefix(r.Path, "web/") ||
								strings.HasPrefix(r.Path, "tests/") {
								return parsed
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
				diffArgs := append([]string{"diff", "--numstat", sha + "^1", sha}, excludeArgs...)
				out, err := runGitCommand(diffArgs...)
				if err == nil {
					parsed := ParseNumstatOutput(out)
					if len(parsed) > 0 {
						for _, r := range parsed {
							if strings.HasPrefix(r.Path, "src/") ||
								strings.HasPrefix(r.Path, "internal/") ||
								strings.HasPrefix(r.Path, "cmd/") ||
								strings.HasPrefix(r.Path, "web/") ||
								strings.HasPrefix(r.Path, "tests/") {
								return parsed
							}
						}
					}
				}
			}
		}
	}

	// 5. Fallback for active ticket working tree: if working directory has changes and matches active branch
	if currBranch != "" && strings.Contains(strings.ToLower(currBranch), strings.ToLower(cleanTicket)) {
		headDiffArgs := append([]string{"diff", "--numstat", "HEAD"}, excludeArgs...)
		headOut, err := runGitCommand(headDiffArgs...)
		if err == nil {
			parsed := ParseNumstatOutput(headOut)
			if len(parsed) > 0 {
				return parsed
			}
		}
	}

	return []NumstatEntry{}
}

// ComputeCodeFacts computes code changes, test/prod breakdown, and SDLC iterations from git diff.
func ComputeCodeFacts(ticketID string, cfg *ClassificationConfig, fallback *domain.CodeFacts) domain.CodeFacts {
	numstat := GetTicketNumstat(ticketID)
	if len(numstat) == 0 && fallback != nil {
		return *fallback
	}

	var linesAdded, linesDeleted int64
	var testLinesAdded, testLinesDeleted int64
	var prodLinesAdded, prodLinesDeleted int64
	var filesChanged int64

	for _, entry := range numstat {
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

	sdlcIters := ComputeSDLCIterations(ticketID)

	return domain.CodeFacts{
		LinesAdded:       linesAdded,
		LinesDeleted:     linesDeleted,
		TestLinesAdded:   testLinesAdded,
		TestLinesDeleted: testLinesDeleted,
		ProdLinesAdded:   prodLinesAdded,
		ProdLinesDeleted: prodLinesDeleted,
		FilesChanged:     filesChanged,
		SDLCIterations:   sdlcIters,
	}
}
