package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agenticlens/internal/domain"
)

// GetGitUser retrieves user name strictly from git config (user.name).
func GetGitUser() string {
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

// InferTicketTypeFromBranchList parses git branch list output to resolve conventional commit type.
func InferTicketTypeFromBranchList(branchOut, ticketID string) *string {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil
	}

	ticketPattern := regexp.MustCompile(`(?i)(?:^|[/_.-])` + regexp.QuoteMeta(cleanTicket) + `(?:[/_.-]|$)`)
	for _, line := range strings.Split(branchOut, "\n") {
		clean := strings.TrimSpace(strings.TrimLeft(line, "*+ "))
		if clean == "" || !ticketPattern.MatchString(clean) {
			continue
		}
		if strings.Contains(clean, "/") {
			parts := strings.Split(clean, "/")
			var filtered []string
			for _, p := range parts {
				if p != "remotes" && p != "" {
					filtered = append(filtered, p)
				}
			}
			if len(filtered) > 0 {
				prefix := strings.ToLower(filtered[0])
				if tt, ok := domain.NormalizeTicketType(prefix); ok {
					s := string(tt)
					return &s
				}
				// If first segment was a remote identifier (e.g. origin, upstream), check the next segment
				if len(filtered) > 1 {
					if tt, ok := domain.NormalizeTicketType(strings.ToLower(filtered[1])); ok {
						s := string(tt)
						return &s
					}
				}
			}
		}
	}
	return nil
}

// InferTicketType infers ticket classification from git branch prefix or caller-supplied configured prefixes.
func InferTicketType(ticketID string, prefixes ...domain.TicketPrefix) *string {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil
	}

	// 1. Search git branches (Tier 2)
	branchOut, err := runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*")
	if err == nil {
		if res := InferTicketTypeFromBranchList(branchOut, cleanTicket); res != nil {
			return res
		}
	}

	// 2. Search configured ticket identifier prefix backup (Tier 3)
	if len(prefixes) > 0 {
		if tt, ok := domain.InferTicketTypeFromPrefix(cleanTicket, prefixes); ok {
			s := string(tt)
			return &s
		}
	}

	return nil
}
