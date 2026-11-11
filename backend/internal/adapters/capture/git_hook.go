package capture

import (
	"agenticlens/internal/adapters/filesystem"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
)

const checkHookHeader = "#!/bin/sh\n# AgenticLens advisory pre-commit v1\n# wrapper-base64: "

// CheckInstallOptions controls explicit, ownership-aware advisory installation.
// Wrapper is the stable common entrypoint, never a version-specific native path.
type CheckInstallOptions struct {
	Repository string
	Wrapper    string
	DryRun     bool
	Remove     bool
}

type CheckInstallResult struct {
	HookPath string `json:"hook_path"`
	Action   string `json:"action"`
	DryRun   bool   `json:"dry_run"`
}

// InstallCheck preserves effective Git hook settings and refuses to compose or
// replace arbitrary hooks. Existing owners can compose the advisory manually.
func InstallCheck(ctx context.Context, options CheckInstallOptions) (CheckInstallResult, error) {
	result := CheckInstallResult{DryRun: options.DryRun}
	repo, err := ProjectDirectory(options.Repository)
	if err != nil {
		return result, fmt.Errorf("resolve advisory repository: %w", err)
	}
	command := filesystem.RepositoryGitCommand(ctx, repo, "rev-parse", "--path-format=absolute", "--git-path", "hooks")
	output, err := command.Output()
	if err != nil {
		return result, fmt.Errorf("resolve effective Git hooks directory: %w", err)
	}
	hooks := strings.TrimSuffix(strings.TrimSuffix(string(output), "\n"), "\r")
	if !filepath.IsAbs(hooks) || strings.ContainsAny(hooks, "\r\n\x00") {
		return result, errors.New("Git returned an unsupported hooks directory")
	}
	result.HookPath = filepath.Join(hooks, "pre-commit")
	before, exists, err := readCheckHook(result.HookPath)
	if err != nil {
		return result, err
	}
	if exists && !ownedCheckHook(before) {
		return result, foreignCheckHook(result.HookPath)
	}
	if options.Remove {
		result.Action = "absent"
		if exists {
			result.Action = "remove"
			if !options.DryRun {
				if err := verifyCheckHook(result.HookPath, before, true); err != nil {
					return result, err
				}
				if err := os.Remove(result.HookPath); err != nil {
					return result, fmt.Errorf("remove owned advisory hook: %w", err)
				}
			}
		}
		return result, nil
	}
	wrapper := options.Wrapper
	if wrapper == "" {
		wrapper, err = adaptersconfig.UserManagerPath()
		if err != nil {
			return result, fmt.Errorf("resolve stable user wrapper: %w", err)
		}
	}
	wrapper, err = ProjectDirectory(wrapper)
	if err != nil || strings.ContainsAny(wrapper, "\r\n\x00") {
		return result, errors.New("advisory wrapper must be a path without control characters")
	}
	after := renderCheckHook(wrapper)
	result.Action = "install"
	if exists {
		result.Action = "update"
		if bytes.Equal(before, after) {
			result.Action = "unchanged"
			return result, nil
		}
	}
	if options.DryRun {
		return result, nil
	}
	if err := writeCheckHook(result.HookPath, before, after, exists); err != nil {
		return result, err
	}
	return result, nil
}

func foreignCheckHook(path string) error {
	return fmt.Errorf("refusing to replace or remove existing or edited hook %s; preserve it and manually compose an advisory call to the stable wrapper with --al-offline --al-project \"$project\" capture check --repo \"$project\", where project is the runtime Git top-level; preserve the existing hook's exit status", path)
}

func readCheckHook(path string) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("inspect advisory hook: %w", err)
	}
	// Do not follow a hook symlink: its target belongs to its existing owner.
	// Symlinked hooks directories remain supported by normal path resolution.
	if !info.Mode().IsRegular() || info.Size() > 64*1024 {
		return nil, true, foreignCheckHook(path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, true, fmt.Errorf("read advisory hook: %w", err)
	}
	return body, true, nil
}

func ownedCheckHook(body []byte) bool {
	if !bytes.HasPrefix(body, []byte(checkHookHeader)) {
		return false
	}
	encoded, script, ok := strings.Cut(string(body[len(checkHookHeader):]), "\n")
	if !ok {
		return false
	}
	wrapper, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(wrapper) == 0 || strings.ContainsAny(string(wrapper), "\r\n\x00") {
		return false
	}
	if bytes.Equal(body, renderCheckHook(string(wrapper))) {
		return true
	}
	// Recognize the exact previously generated template for safe refresh/removal.
	// The retired script is never generated or executed by this compatibility check.
	if base64.StdEncoding.EncodeToString(wrapper) != encoded {
		return false
	}
	invocation := QuoteArgument(filepath.ToSlash(string(wrapper))) + " '--al-offline' '--al-project'"
	if strings.Count(script, invocation) != 1 {
		return false
	}
	normalized := strings.Replace(script, invocation, "__AGENTICLENS_MANAGER__ '--al-offline' '--al-project'", 1)
	return hash([]byte(normalized)) == "b740ee74c5fa10e0cea61e0e43dfa2622995164650065b37842b33c788745c44"
}

func renderCheckHook(wrapper string) []byte {
	command := []string{wrapper, "--al-offline", "--al-project"}
	quoted := make([]string, len(command))
	for i, arg := range command {
		quoted[i] = QuoteArgument(filepath.ToSlash(arg))
	}
	// Managers exec the selected engine, preserving cancellation and exit status.
	// Watchdog descriptors are detached so sleep cannot delay EOF.
	// dd bounds bytes before awk bounds lines, even for malformed native output.
	return []byte(checkHookHeader + base64.StdEncoding.EncodeToString([]byte(wrapper)) + `
project=$(git rev-parse --show-toplevel 2>/dev/null) || {
  printf '%s\n' 'AgenticLens capture advisory: project unavailable; inspect capture status.' >&2
  exit 0
}
if [ ! -f "$project/.agenticlens-config.yaml" ] &&
   [ ! -f "$project/.agenticlens-config.yml" ]; then
  exit 0
fi
(
  ` + strings.Join(quoted, " ") + ` "$project" 'capture' 'check' '--repo' "$project" </dev/null >/dev/null &
  handler=$!
  (
    sleep 3
    kill -TERM "$handler" 2>/dev/null
    sleep 1
    kill -KILL "$handler" 2>/dev/null
  ) >/dev/null 2>&1 &
  watchdog=$!
  wait "$handler" 2>/dev/null
  status=$?
  kill "$watchdog" 2>/dev/null
  wait "$watchdog" 2>/dev/null
  if [ "$status" -ne 0 ]; then
    printf '%s\n' 'AgenticLens capture advisory: check unavailable; inspect capture status or refresh user wrapper setup.' >&2
  fi
) 2>&1 | dd bs=1 count=512 2>/dev/null | LC_ALL=C awk 'NR == 1 { sub(/\r$/, ""); print }' >&2
exit 0
`)
}

func verifyCheckHook(path string, before []byte, existed bool) error {
	current, exists, err := readCheckHook(path)
	if err != nil {
		return err
	}
	if exists != existed || !bytes.Equal(current, before) {
		return fmt.Errorf("advisory hook changed during installation: %s; retry after inspecting its owner", path)
	}
	return nil
}

func writeCheckHook(path string, before, after []byte, existed bool) (resultErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create advisory hooks directory: %w", err)
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".agenticlens-pre-commit-*")
	if err != nil {
		return fmt.Errorf("stage advisory hook: %w", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove advisory staging file: %w", err))
		}
	}()
	_, writeErr := file.Write(after)
	err = errors.Join(writeErr, file.Chmod(0755), file.Close())
	if err != nil {
		return fmt.Errorf("stage advisory hook contents: %w", err)
	}
	if err := verifyCheckHook(path, before, existed); err != nil {
		return err
	}
	if !existed {
		// Linking the fully staged file publishes it without ever truncating a
		// competing hook or exposing partial shell syntax after a write failure.
		if err := os.Link(file.Name(), path); err != nil {
			return fmt.Errorf("publish absent advisory hook without replacement: %w", err)
		}
		return nil
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("replace unchanged owned advisory hook: %w", err)
	}
	return nil
}
