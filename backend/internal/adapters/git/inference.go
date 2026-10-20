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

// InferTicketType infers ticket classification from caller-supplied configured prefixes or git branch prefix.
func InferTicketType(ticketID string, prefixes ...domain.TicketPrefix) *string {
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil
	}

	// 1. Evaluate configured ticket identifier prefixes first (shortcut avoiding subprocess forks)
	if len(prefixes) > 0 {
		if tt, ok := domain.InferTicketTypeFromPrefix(cleanTicket, prefixes); ok {
			s := string(tt)
			return &s
		}
	}

	// 2. Search git branches if not resolved by configured prefixes
	branchOut, err := runGitCommand("branch", "-a", "--list", "*"+cleanTicket+"*")
	if err == nil {
		if res := InferTicketTypeFromBranchList(branchOut, cleanTicket); res != nil {
			return res
		}
	}

	return nil
}
