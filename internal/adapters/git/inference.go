package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GetGitUser retrieves user email or user name from git config.
func GetGitUser() string {
	email, err := runGitCommand("config", "user.email")
	if err == nil {
		clean := strings.TrimSpace(email)
		if clean != "" {
			return clean
		}
	}

	name, err := runGitCommand("config", "user.name")
	if err == nil {
		clean := strings.TrimSpace(name)
		if clean != "" {
			return clean
		}
	}

	return "unknown"
}

func findWorkspaceFile(ticketID, filename string) ([]byte, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, false
	}
	curr := cwd
	for {
		candidate := filepath.Join(curr, "workspaces", ticketID, filename)
		if data, err := os.ReadFile(candidate); err == nil {
			return data, true
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return nil, false
}

func findWorkspaceDir(ticketID, subDir string) (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", false
	}
	curr := cwd
	for {
		candidate := filepath.Join(curr, "workspaces", ticketID, subDir)
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			break
		}
		curr = parent
	}
	return "", false
}

// ComputeSDLCIterations calculates iterations count from ticket workspace iterations directory.
func ComputeSDLCIterations(ticketID string) int64 {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return 1
	}

	if iterDir, ok := findWorkspaceDir(cleanTicket, "iterations"); ok {
		entries, err := os.ReadDir(iterDir)
		if err == nil {
			var count int64
			for _, e := range entries {
				if e.IsDir() {
					count++
				}
			}
			if count > 0 {
				return count
			}
		}
	}

	return 1
}

// InferTicketTitle extracts a human-readable title from SUMMARY.md, CONTRACT.md, or git log.
func InferTicketTitle(ticketID string) *string {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil
	}

	// 1. Check SUMMARY.md
	if data, ok := findWorkspaceFile(cleanTicket, "SUMMARY.md"); ok {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			first := strings.TrimSpace(lines[0])
			clean := cleanHeadingTitle(first, cleanTicket, "Summary")
			if clean != "" {
				return &clean
			}
		}
	}

	// 2. Check CONTRACT.md
	if data, ok := findWorkspaceFile(cleanTicket, "CONTRACT.md"); ok {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			first := strings.TrimSpace(lines[0])
			clean := cleanHeadingTitle(first, cleanTicket, "Contract")
			if clean != "" {
				return &clean
			}
		}
	}

	// 3. Search git log commits mentioning ticket
	logOut, err := runGitCommand("log", "--all", "--grep="+cleanTicket, "--oneline", "-n", "15")
	if err == nil {
		lines := strings.Split(logOut, "\n")
		metricsRegex := regexp.MustCompile(`(?i)\b(refactor|feat|chore)\((agentic/)?metrics\)`)
		ticketPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(cleanTicket) + `[:\s\-]+([^(#\n]+)`)
		slcpPattern := regexp.MustCompile(`(?i)^\[SLCP\]\s*`)

		for _, line := range lines {
			if metricsRegex.MatchString(line) {
				continue
			}
			m := ticketPattern.FindStringSubmatch(line)
			if len(m) >= 2 {
				clean := strings.TrimSpace(m[1])
				clean = slcpPattern.ReplaceAllString(clean, "")
				clean = strings.Trim(clean, ": -")
				if clean != "" && !strings.HasPrefix(clean, "DPS-") {
					return &clean
				}
			}
		}
	}

	return nil
}

func cleanHeadingTitle(line, ticketID, docType string) string {
	clean := regexp.MustCompile(`^#\s*`).ReplaceAllString(line, "")
	clean = regexp.MustCompile(`(?i)Delivery\s+Summary:?\s*`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`(?i)`+docType+`:?\s*`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`(?i)`+regexp.QuoteMeta(ticketID)+`:?\s*`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`(?i)\s*-\s*`+docType+`$`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`(?i)\s*Delivery\s+Summary$`).ReplaceAllString(clean, "")
	return strings.Trim(strings.TrimSpace(clean), ": -")
}

// InferTicketType infers ticket classification (feature, bug, chore, refactor, documentation) from git.
func InferTicketType(ticketID string) *string {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil
	}

	// 1. Search git branches
	branchOut, err := runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*")
	if err == nil {
		for _, line := range strings.Split(branchOut, "\n") {
			clean := strings.TrimSpace(strings.TrimLeft(line, "*+ "))
			if strings.Contains(clean, "/") {
				parts := strings.Split(clean, "/")
				var filtered []string
				for _, p := range parts {
					if p != "remotes" {
						filtered = append(filtered, p)
					}
				}
				prefix := ""
				if len(filtered) > 0 {
					if filtered[0] != "origin" {
						prefix = strings.ToLower(filtered[0])
					} else if len(filtered) > 1 {
						prefix = strings.ToLower(filtered[1])
					}
				}
				res := mapPrefixToType(prefix)
				if res != nil {
					return res
				}
			}
		}
	}

	// 2. Search git log commits mentioning ticket
	logOut, err := runGitCommand("log", "--all", "--grep="+cleanTicket, "--oneline", "-n", "5")
	if err == nil {
		for _, line := range strings.Split(logOut, "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "feat(") || strings.Contains(lower, "feat:") || strings.Contains(lower, "feature") {
				feat := "feature"
				return &feat
			}
			if strings.Contains(lower, "fix(") || strings.Contains(lower, "fix:") || strings.Contains(lower, "bug") {
				bug := "bug"
				return &bug
			}
			if strings.Contains(lower, "chore(") || strings.Contains(lower, "chore:") || strings.Contains(lower, "ci(") || strings.Contains(lower, "ci:") {
				chore := "chore"
				return &chore
			}
			if strings.Contains(lower, "refactor(") || strings.Contains(lower, "refactor:") {
				ref := "refactor"
				return &ref
			}
		}
	}

	feat := "feature"
	return &feat
}

func mapPrefixToType(prefix string) *string {
	var val string
	switch prefix {
	case "feat", "feature":
		val = "feature"
	case "fix", "bug", "bugfix":
		val = "bug"
	case "chore", "ci", "task":
		val = "chore"
	case "refactor":
		val = "refactor"
	case "docs":
		val = "documentation"
	default:
		return nil
	}
	return &val
}
