package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowFiles_Integrity(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("could not determine repository root: %v", err)
	}

	workflows := []struct {
		name     string
		relPath  string
		required []string
	}{
		{
			name:    "Release Candidate Workflow",
			relPath: ".github/workflows/release-candidate.yml",
			required: []string{
				"name: Release Candidate",
				"branches:",
				"- main",
				"release-candidate:",
				"actions/checkout@v4",
				"actions/setup-go@v5",
				"actions/setup-node@v4",
				"make check",
				"make test",
				"make release-build",
				"make release-package",
				"make release-verify",
				"softprops/action-gh-release@v2",
				"prerelease: true",
				"dist/SHA256SUMS",
			},
		},
		{
			name:    "Stable Release Workflow",
			relPath: ".github/workflows/release-stable.yml",
			required: []string{
				"name: Stable Release",
				"tags:",
				"- 'v*.*.*'",
				"release-stable:",
				"Validate Semantic Versioning",
				"^v[0-9]+\\.[0-9]+\\.[0-9]+$",
				"make check",
				"make test",
				"make release-build",
				"make release-package",
				"make release-verify",
				"softprops/action-gh-release@v2",
				"prerelease: false",
				"make_latest: true",
				"git tag -fa latest",
				"git push -f origin latest",
			},
		},
	}

	for _, wf := range workflows {
		t.Run(wf.name, func(t *testing.T) {
			path := filepath.Join(repoRoot, wf.relPath)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", wf.relPath, err)
			}

			str := string(content)

			for _, req := range wf.required {
				if !strings.Contains(str, req) {
					t.Errorf("%s missing required string %q", wf.relPath, req)
				}
			}
		})
	}
}

func TestMakefile_ReleaseTargetsListedInHelp(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("could not determine repo root: %v", err)
	}

	cmd := exec.Command("make", "help")
	cmd.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("make help failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	requiredTargets := []string{
		"release-build",
		"release-package",
		"release-verify",
		"release-clean",
	}

	for _, target := range requiredTargets {
		if !strings.Contains(output, target) {
			t.Errorf("make help missing required release target %q", target)
		}
	}
}

func TestMakefile_ReleaseVerifyFailsWhenChecksumMissing(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("could not determine repo root: %v", err)
	}

	tempDist := t.TempDir()

	cmd := exec.Command("make", "release-verify", fmt.Sprintf("RELEASE_DIR=%s", tempDist))
	cmd.Dir = repoRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Fatalf("expected make release-verify to fail when SHA256SUMS is missing, but it succeeded")
	}

	if !strings.Contains(stderr.String(), "SHA256SUMS not found") {
		t.Errorf("expected error message to mention 'SHA256SUMS not found', got: %s", stderr.String())
	}
}

func TestMakefile_ReleasePackageFailsWhenBinaryMissing(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("could not determine repo root: %v", err)
	}

	tempDist := t.TempDir()

	cmd := exec.Command("make", "release-package", fmt.Sprintf("RELEASE_DIR=%s", tempDist))
	cmd.Dir = repoRoot
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		t.Fatalf("expected make release-package to fail when binaries are missing, but it succeeded")
	}

	if !strings.Contains(stderr.String(), "not found. Run 'make release-build' first.") {
		t.Errorf("expected error message to direct running 'make release-build', got: %s", stderr.String())
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "Makefile")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find repo root from %s", dir)
}
