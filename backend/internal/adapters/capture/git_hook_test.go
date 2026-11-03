package capture_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
)

func checkGitCommand(repo string, args ...string) *exec.Cmd {
	command := exec.Command("git", append([]string{"-C", repo, "-c", "commit.gpgsign=false"}, args...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_AUTHOR_NAME=Capture Test", "GIT_AUTHOR_EMAIL=capture@example.invalid",
		"GIT_COMMITTER_NAME=Capture Test", "GIT_COMMITTER_EMAIL=capture@example.invalid")
	return command
}

func checkGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	output, err := checkGitCommand(repo, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return string(output)
}

func checkWrite(t *testing.T, path, body string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatal(err)
	}
}

func checkRead(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func checkRepository(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	checkGit(t, repo, "init", "--quiet")
	checkWrite(t, filepath.Join(repo, ".agenticlens-config.yaml"), "agenticlens:\n  version: v1.0.0\n", 0644)
	checkGit(t, repo, "add", ".agenticlens-config.yaml")
	checkGit(t, repo, "commit", "--quiet", "-m", "initial")
	return repo
}

func installCheck(t *testing.T, options capture.CheckInstallOptions) capture.CheckInstallResult {
	t.Helper()
	result, err := capture.InstallCheck(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestGitCheckOwnershipPreviewUpgradeAndRemoval(t *testing.T) {
	repo := checkRepository(t)
	checkGit(t, repo, "config", "core.hooksPath", "custom hooks")
	config := checkRead(t, filepath.Join(repo, ".git", "config"))
	index := checkRead(t, filepath.Join(repo, ".git", "index"))
	paths := t.TempDir()
	physical := filepath.Join(paths, "physical")
	if err := os.MkdirAll(filepath.Join(physical, "child"), 0755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(paths, "alias")
	if err := os.Symlink(filepath.Join(physical, "child"), alias); err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(physical, "wrapper ' with spaces")
	checkWrite(t, wrapper, "#!/bin/sh\nexit 0\n", 0755)
	// Preserve the raw symlink/.. sequence: lexical cleaning selects a different
	// wrapper than filesystem traversal. Alias and physical installs must agree.
	rawWrapper := alias + string(filepath.Separator) + ".." + string(filepath.Separator) + filepath.Base(wrapper)
	options := capture.CheckInstallOptions{Repository: repo, Wrapper: rawWrapper, DryRun: true}
	preview := installCheck(t, options)
	if preview.HookPath != filepath.Join(repo, "custom hooks", "pre-commit") || preview.Action != "install" || !preview.DryRun {
		t.Fatalf("preview: %+v", preview)
	}
	if _, err := os.Stat(filepath.Dir(preview.HookPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("preview created hooks directory: %v", err)
	}
	options.DryRun = false
	installed := installCheck(t, options)
	before := checkRead(t, installed.HookPath)
	info, err := os.Stat(installed.HookPath)
	if err != nil || info.Mode().Perm()&0111 == 0 {
		t.Fatalf("hook is not executable: %v", err)
	}
	if result := installCheck(t, options); result.Action != "unchanged" {
		t.Fatalf("repeat installation: %+v", result)
	}
	options.Wrapper = wrapper
	if result := installCheck(t, options); result.Action != "unchanged" {
		t.Fatalf("physical wrapper alias changed installed identity: %+v", result)
	}
	options.Wrapper = filepath.Join(t.TempDir(), "replacement wrapper")
	options.DryRun = true
	if result := installCheck(t, options); result.Action != "update" || !bytes.Equal(before, checkRead(t, installed.HookPath)) {
		t.Fatalf("update preview changed hook: %+v", result)
	}
	options.DryRun = false
	if result := installCheck(t, options); result.Action != "update" || bytes.Equal(before, checkRead(t, installed.HookPath)) {
		t.Fatalf("owned upgrade failed: %+v", result)
	}
	options.Remove, options.DryRun = true, true
	if result := installCheck(t, options); result.Action != "remove" {
		t.Fatalf("removal preview: %+v", result)
	}
	if _, err := os.Stat(installed.HookPath); err != nil {
		t.Fatalf("preview removed hook: %v", err)
	}
	options.DryRun = false
	if result := installCheck(t, options); result.Action != "remove" {
		t.Fatalf("remove: %+v", result)
	}
	if result := installCheck(t, options); result.Action != "absent" {
		t.Fatalf("repeat removal: %+v", result)
	}
	if _, err := os.Stat(installed.HookPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("owned hook retained: %v", err)
	}
	if !bytes.Equal(config, checkRead(t, filepath.Join(repo, ".git", "config"))) || !bytes.Equal(index, checkRead(t, filepath.Join(repo, ".git", "index"))) {
		t.Fatal("installation changed Git configuration or index")
	}
}

func TestGitCheckPreservesForeignEditedAndSymlinkHooks(t *testing.T) {
	repo := checkRepository(t)
	wrapper := filepath.Join(t.TempDir(), "wrapper")
	options := capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper}
	installed := installCheck(t, options)
	owned := checkRead(t, installed.HookPath)
	foreign := "#!/bin/sh\nprintf '%s\\n' 'foreign sentinel' >&2\nexit 23\n"
	for _, body := range []string{foreign, string(owned) + "# user edit\n"} {
		checkWrite(t, installed.HookPath, body, 0755)
		for _, remove := range []bool{false, true} {
			for _, dryRun := range []bool{false, true} {
				options.Remove, options.DryRun = remove, dryRun
				if _, err := capture.InstallCheck(context.Background(), options); err == nil || !strings.Contains(err.Error(), "manually compose") {
					t.Fatalf("foreign hook not refused remove=%t dry-run=%t: %v", remove, dryRun, err)
				}
				if !bytes.Equal([]byte(body), checkRead(t, installed.HookPath)) {
					t.Fatal("foreign hook changed")
				}
			}
		}
	}
	checkWrite(t, installed.HookPath, foreign, 0755)
	output, err := checkGitCommand(repo, "commit", "--allow-empty", "-m", "blocked by owner").CombinedOutput()
	if err == nil || !strings.Contains(string(output), "foreign sentinel") {
		t.Fatalf("foreign Git behavior lost: %v %s", err, output)
	}
	command := exec.Command("sh", installed.HookPath)
	var exit *exec.ExitError
	if err := command.Run(); !errors.As(err, &exit) || exit.ExitCode() != 23 {
		t.Fatalf("foreign exit changed: %v", err)
	}
	if err := os.Remove(installed.HookPath); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "owned elsewhere")
	checkWrite(t, target, string(owned), 0755)
	if err := os.Symlink(target, installed.HookPath); err != nil {
		t.Fatal(err)
	}
	if _, err := capture.InstallCheck(context.Background(), options); err == nil {
		t.Fatal("installer followed hook symlink")
	}
	if !bytes.Equal(owned, checkRead(t, target)) {
		t.Fatal("symlink target changed")
	}
}

func TestGitCheckRuntimeWorktreeAndSharedHookIsolation(t *testing.T) {
	repo := checkRepository(t)
	shared := t.TempDir()
	alias := filepath.Join(t.TempDir(), "hooks alias")
	if err := os.Symlink(shared, alias); err != nil {
		t.Fatal(err)
	}
	checkGit(t, repo, "config", "core.hooksPath", alias)
	wrapper := filepath.Join(t.TempDir(), "wrapper ' $value with spaces")
	checkWrite(t, wrapper, "#!/bin/sh\n[ \"$1\" = --al-offline ] && [ \"$2\" = --al-project ] && [ \"$4\" = capture ] && [ \"$5\" = check ] && [ \"$6\" = --repo ] && [ \"$3\" = \"$7\" ] || exit 9\nprintf 'checked project=%s\\n' \"$3\" >&2\nprintf 'stdout must remain hidden\\n'\n", 0755)
	installed := installCheck(t, capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper})
	if installed.HookPath != filepath.Join(shared, "pre-commit") {
		t.Fatalf("effective hook directory ignored: %+v", installed)
	}
	if info, err := os.Lstat(alias); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("hook directory alias changed: %v", err)
	}
	worktree := filepath.Join(t.TempDir(), "other worktree")
	checkGit(t, repo, "worktree", "add", "-b", "isolated", worktree)
	for _, project := range []string{repo, worktree} {
		var stderr bytes.Buffer
		command := checkGitCommand(project, "commit", "--quiet", "--allow-empty", "-m", "check context")
		command.Stderr = &stderr
		stdout, err := command.Output()
		if err != nil || len(stdout) != 0 || stderr.String() != "checked project="+filepath.ToSlash(project)+"\n" {
			t.Fatalf("wrong runtime context: %v stdout=%q stderr=%q", err, stdout, stderr.String())
		}
	}
	unrelated := t.TempDir()
	checkGit(t, unrelated, "init", "--quiet")
	checkGit(t, unrelated, "config", "core.hooksPath", alias)
	output := checkGit(t, unrelated, "commit", "--quiet", "--allow-empty", "-m", "unrelated")
	if output != "" {
		t.Fatalf("shared hook warned on unconfigured project: %q", output)
	}
}

func TestGitCheckAdvisoryFailuresOutputAndDeadline(t *testing.T) {
	repo := checkRepository(t)
	wrapper := filepath.Join(t.TempDir(), "wrapper")
	options := capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper}
	installCheck(t, options)
	for _, body := range []string{
		"", // Missing wrapper must not block a real commit.
		"#!/bin/sh\nexit 17\n",
		"#!/bin/sh\nprintf '%0900d\\nsecond diagnostic\\n' 0 >&2\nexit 17\n",
		"#!/bin/sh\ntrap '' TERM\nexec sleep 30\n",
	} {
		if body != "" {
			checkWrite(t, wrapper, body, 0755)
		}
		start := time.Now()
		output := checkGit(t, repo, "commit", "--quiet", "--allow-empty", "-m", "advisory only")
		if elapsed := time.Since(start); elapsed > 8*time.Second {
			t.Fatalf("advisory exceeded deadline: %v", elapsed)
		}
		if output == "" || len(output) > 513 || strings.Count(output, "\n") != 1 || strings.Contains(output, "second diagnostic") {
			t.Fatalf("unbounded or missing advisory: %q", output)
		}
	}
}

func TestGitCheckPowerShellUsesCommonEntrypoint(t *testing.T) {
	repo := checkRepository(t)
	result := installCheck(t, capture.CheckInstallOptions{Repository: repo, Wrapper: filepath.Join(t.TempDir(), "wrapper with spaces.ps1")})
	body := string(checkRead(t, result.HookPath))
	if !strings.Contains(body, "powershell.exe' '-NoLogo' '-NoProfile' '-NonInteractive' '-File'") || strings.Contains(body, "ExecutionPolicy") {
		t.Fatalf("native PowerShell entrypoint not preserved: %s", body)
	}
}
