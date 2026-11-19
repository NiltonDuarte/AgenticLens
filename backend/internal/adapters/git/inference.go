package git

import (
	"context"
	"os/exec"
	"strings"
)

// GetGitUser retrieves user name strictly from git config (user.name).
func GetGitUser() string {
	return GetGitUserInDir("")
}

// GetGitUserInDir uses the selected project's Git identity configuration.
func GetGitUserInDir(repoRoot string) string {
	return GetGitUserInDirContext(context.Background(), repoRoot)
}

// GetGitUserInDirContext limits the Git identity lookup to the caller's context.
func GetGitUserInDirContext(ctx context.Context, repoRoot string) string {
	cmd := exec.CommandContext(ctx, "git", "config", "user.name")
	cmd.Dir = repoRoot
	name, err := cmd.Output()
	if err == nil {
		clean := strings.TrimSpace(string(name))
		if clean != "" {
			return clean
		}
	}

	return "unknown"
}
