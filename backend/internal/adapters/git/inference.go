package git

import "strings"

// GetGitUser retrieves user name strictly from git config (user.name).
func GetGitUser() string {
	return GetGitUserInDir("")
}

// GetGitUserInDir uses the selected project's Git identity configuration.
func GetGitUserInDir(repoRoot string) string {
	name, err := runGitCommandInDir(repoRoot, "config", "user.name")
	if err == nil {
		clean := strings.TrimSpace(name)
		if clean != "" {
			return clean
		}
	}

	return "unknown"
}
