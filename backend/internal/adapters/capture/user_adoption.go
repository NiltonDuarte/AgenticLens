package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UserCodexInstalled observes global activation only when a preserved legacy
// receipt explicitly deactivated a project integration. It is never an adoption
// prerequisite for configured projects without legacy installation state.
func UserCodexInstalled() (bool, error) {
	home, err := EffectiveCodexHome()
	if err != nil {
		return false, err
	}
	data, err := os.ReadFile(filepath.Join(home, "agenticlens-user-installation.json"))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var receipt userInstallation
	if err := json.Unmarshal(data, &receipt); err != nil {
		return false, err
	}
	if receipt.Protocol != 1 {
		return false, errors.New("unsupported user capture ownership protocol")
	}
	return len(receipt.Groups) > 0, nil
}

func checkLegacyUserConflicts(ctx context.Context, project, migration string) error {
	project, err := ProjectDirectory(project)
	if err != nil {
		return err
	}
	candidates := []string{project}
	// Native Codex project hooks are discovered from the original checkout, so a
	// worktree-local scan alone cannot establish absence of duplicate callbacks.
	output, gitErr := exec.CommandContext(ctx, "git", "-C", project, "worktree", "list", "--porcelain").Output()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if gitErr != nil {
		if info, err := os.Stat(filepath.Join(project, ".git")); err == nil && info.Mode().IsRegular() {
			return fmt.Errorf("cannot establish original checkout legacy hook ownership: %w", gitErr)
		}
	}
	if gitErr == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "worktree ") {
				candidates = append(candidates, strings.TrimPrefix(line, "worktree "))
				break
			}
		}
	}
	if output, err := exec.CommandContext(ctx, "git", "-C", project, "rev-parse", "--show-toplevel").Output(); err == nil {
		candidates = append(candidates, strings.TrimSpace(string(output)))
	}
	if migration != "" {
		migration, err = ProjectDirectory(migration)
		if err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate, err = ProjectDirectory(candidate)
		if err != nil {
			return err
		}
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		if candidate == migration {
			continue
		} // Post-recovery migration planning examines the remaining groups; only
		// exact receipt-attested removed groups are exempt, never the checkout.
		path := filepath.Join(candidate, ".codex", "hooks.json")
		data, err := readEntrypointFile(path, installationHealthFileLimit)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		var document struct {
			Hooks map[string][]json.RawMessage `json:"hooks"`
		}
		if err := json.Unmarshal(data, &document); err != nil {
			return fmt.Errorf("legacy project hooks %s invalid; preserved: %w", path, err)
		}
		if legacyCaptureGroupsRemain(document.Hooks) {
			return fmt.Errorf("active legacy AgenticLens project callbacks at %s would duplicate user hooks; preserve them and explicitly use --migrate-legacy-repo %s for unchanged owned callbacks, or resolve unknown/modified callbacks", path, candidate)
		}
	}
	return nil
}

// legacyCaptureGroupsRemain recognizes the actual legacy receive invocation as
// well as branded callbacks. Unknown executable callbacks remain user-owned;
// identifying a conflict never authorizes deleting their bytes.
func legacyCaptureGroupsRemain(hooks map[string][]json.RawMessage) bool {
	for _, groups := range hooks {
		for _, group := range groups {
			lower := strings.ToLower(string(group))
			if strings.Contains(lower, "agenticlens") || strings.Contains(lower, "agentic-lens") || strings.Contains(lower, strings.ToLower(marker)) {
				return true
			}
			var handlers struct {
				Hooks []struct {
					Type    string `json:"type"`
					Command string `json:"command"`
					Windows string `json:"commandWindows"`
				} `json:"hooks"`
			}
			if json.Unmarshal(group, &handlers) != nil {
				continue
			}
			for _, handler := range handlers.Hooks {
				if handler.Type != "command" {
					continue
				}
				for _, command := range []string{handler.Command, handler.Windows} {
					tokens := strings.Fields(strings.NewReplacer("'", "", "\"", "").Replace(strings.ToLower(command)))
					for i := 1; i < len(tokens); i++ {
						if tokens[i-1] == "capture" && tokens[i] == "receive" || tokens[i-1] == "dispatch" && tokens[i] == "codex-hook" {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
