package main_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writeCLISharedPolicy(t *testing.T, project string) {
	t.Helper()
	body, err := os.ReadFile("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
}

// Fresh metrics always measure a real committed Git comparison, including zero
// net contribution. This helper does not alter histories authored by a test.
func initCLIGitFixture(t *testing.T, project string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(project, ".git")); err == nil {
		return
	}
	for _, args := range [][]string{{"init", "-qb", "main"}, {"config", "user.name", "CLI Fixture"}, {"config", "user.email", "cli@example.invalid"}, {"commit", "--allow-empty", "-m", "fixture base"}, {"branch", "fixture-base"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = project
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
}

func cliFixtureSettings(t *testing.T) string {
	t.Helper()
	absolute, err := filepath.Abs("testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return absolute
}

// A recognized native owner with zero turns and no model observations is a
// complete empty source snapshot. The command fixture fills its actual cwd.
func setupCLIEmptyCodexRoot(t *testing.T) string {
	t.Helper()
	home := filepath.Join(t.TempDir(), "empty-codex-source")
	if err := os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	return home
}

func enableCLIFixtureSummary(t *testing.T, project string) {
	t.Helper()
	path := filepath.Join(project, ".agenticlens-config.yaml")
	body := readCaptureFile(t, path)
	body = bytes.Replace(body, []byte("  summary:\n    enabled: false"), []byte("  summary:\n    enabled: true"), 1)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
}
