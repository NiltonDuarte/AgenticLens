package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func resolveArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "amd64"
	case "arm64":
		return "arm64"
	default:
		return runtime.GOARCH
	}
}

func resolveOS() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

func TestWrapper_DispatchesToMatchingPlatformBinary(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	tempDir := t.TempDir()
	wrapperSrc := filepath.Join(repoRoot, "scripts", "agentic-lens-wrapper.sh")
	wrapperDst := filepath.Join(tempDir, "agentic-lens")

	content, err := os.ReadFile(wrapperSrc)
	if err != nil {
		t.Fatalf("failed reading wrapper script: %v", err)
	}
	if err := os.WriteFile(wrapperDst, content, 0755); err != nil {
		t.Fatalf("failed writing wrapper script: %v", err)
	}

	currentOS := resolveOS()
	currentArch := resolveArch()
	mockBinName := fmt.Sprintf("agentic-lens-%s-%s", currentOS, currentArch)
	mockBinPath := filepath.Join(tempDir, mockBinName)

	mockScript := `#!/bin/sh
echo "mock platform binary executed with: $@"
`
	if err := os.WriteFile(mockBinPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("failed writing mock binary: %v", err)
	}

	cmd := exec.Command(wrapperDst, "arg1", "arg2")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "AGENTIC_LENS_NO_UPDATE_CHECK=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper execution failed: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "mock platform binary executed with: arg1 arg2") {
		t.Errorf("unexpected wrapper execution output: %s", out)
	}
}

func TestWrapper_UpdateNoticeSurfacesOnNewerVersion(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	tempDir := t.TempDir()
	wrapperSrc := filepath.Join(repoRoot, "scripts", "agentic-lens-wrapper.sh")
	wrapperDst := filepath.Join(tempDir, "agentic-lens")

	content, err := os.ReadFile(wrapperSrc)
	if err != nil {
		t.Fatalf("failed reading wrapper script: %v", err)
	}
	if err := os.WriteFile(wrapperDst, content, 0755); err != nil {
		t.Fatalf("failed writing wrapper script: %v", err)
	}

	currentOS := resolveOS()
	currentArch := resolveArch()
	mockBinName := fmt.Sprintf("agentic-lens-%s-%s", currentOS, currentArch)
	mockBinPath := filepath.Join(tempDir, mockBinName)

	mockScript := `#!/bin/sh
if [ "$1" = "version" ] && [ "$2" = "--short" ]; then
  echo "v1.0.0"
  exit 0
fi
echo "mock output"
`
	if err := os.WriteFile(mockBinPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("failed writing mock binary: %v", err)
	}

	// Write a newer version cache
	latestFile := filepath.Join(tempDir, ".latest_version")
	if err := os.WriteFile(latestFile, []byte("v1.2.0\n"), 0644); err != nil {
		t.Fatalf("failed writing .latest_version: %v", err)
	}

	cmd := exec.Command(wrapperDst, "version", "--short")
	cmd.Dir = tempDir
	// Clear disable check env var
	env := filterEnv(os.Environ(), "AGENTIC_LENS_NO_UPDATE_CHECK")
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper execution failed: %v, stderr: %s", err, stderr.String())
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "Notice: AgenticLens update available: v1.2.0 (current: v1.0.0)") {
		t.Errorf("expected update notice on stderr, got: %s", errOut)
	}
}

func TestWrapper_UpdateNoticeSuppressedWhenDisabled(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	tempDir := t.TempDir()
	wrapperSrc := filepath.Join(repoRoot, "scripts", "agentic-lens-wrapper.sh")
	wrapperDst := filepath.Join(tempDir, "agentic-lens")

	content, err := os.ReadFile(wrapperSrc)
	if err != nil {
		t.Fatalf("failed reading wrapper script: %v", err)
	}
	if err := os.WriteFile(wrapperDst, content, 0755); err != nil {
		t.Fatalf("failed writing wrapper script: %v", err)
	}

	currentOS := resolveOS()
	currentArch := resolveArch()
	mockBinName := fmt.Sprintf("agentic-lens-%s-%s", currentOS, currentArch)
	mockBinPath := filepath.Join(tempDir, mockBinName)

	mockScript := `#!/bin/sh
if [ "$1" = "version" ] && [ "$2" = "--short" ]; then
  echo "v1.0.0"
  exit 0
fi
echo "mock output"
`
	if err := os.WriteFile(mockBinPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("failed writing mock binary: %v", err)
	}

	latestFile := filepath.Join(tempDir, ".latest_version")
	if err := os.WriteFile(latestFile, []byte("v1.2.0\n"), 0644); err != nil {
		t.Fatalf("failed writing .latest_version: %v", err)
	}

	cmd := exec.Command(wrapperDst, "version", "--short")
	cmd.Dir = tempDir
	cmd.Env = append(os.Environ(), "AGENTIC_LENS_NO_UPDATE_CHECK=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("wrapper execution failed: %v, stderr: %s", err, stderr.String())
	}

	errOut := stderr.String()
	if strings.Contains(errOut, "Notice: AgenticLens update available") {
		t.Errorf("expected update notice to be suppressed when AGENTIC_LENS_NO_UPDATE_CHECK=1, got: %s", errOut)
	}
}

func TestDownloader_DryRunModes(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	downloaderScript := filepath.Join(repoRoot, "scripts", "download-agentic-lens.sh")

	// 1. Basic dry run
	cmd1 := exec.Command(downloaderScript, "--dry-run")
	var out1 bytes.Buffer
	cmd1.Stdout = &out1
	if err := cmd1.Run(); err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	s1 := out1.String()
	if !strings.Contains(s1, "Target Wrapper:") || !strings.Contains(s1, "Binaries to install:") {
		t.Errorf("expected wrapper and binary list in dry-run, got: %s", s1)
	}

	// 2. Include host dry run
	cmd2 := exec.Command(downloaderScript, "--dry-run", "--include-host", "darwin-arm64")
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	if err := cmd2.Run(); err != nil {
		t.Fatalf("dry-run with include-host failed: %v", err)
	}
	s2 := out2.String()
	if !strings.Contains(s2, "agentic-lens-darwin-arm64") {
		t.Errorf("expected darwin-arm64 in download list with --include-host, got: %s", s2)
	}

	// 3. All platforms dry run
	cmd3 := exec.Command(downloaderScript, "--dry-run", "--all-platforms")
	var out3 bytes.Buffer
	cmd3.Stdout = &out3
	if err := cmd3.Run(); err != nil {
		t.Fatalf("dry-run with all-platforms failed: %v", err)
	}
	s3 := out3.String()
	for _, expectedBin := range []string{
		"agentic-lens-linux-amd64",
		"agentic-lens-linux-arm64",
		"agentic-lens-darwin-amd64",
		"agentic-lens-darwin-arm64",
		"agentic-lens-windows-amd64.exe",
	} {
		if !strings.Contains(s3, expectedBin) {
			t.Errorf("expected %s in --all-platforms output, got: %s", expectedBin, s3)
		}
	}
}

func TestDownloader_LocalBinaryAndWrapperInstallation(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	tempDir := t.TempDir()
	mockBinPath := filepath.Join(tempDir, "mock-build")
	mockScript := `#!/bin/sh
echo "mock binary local installation test"
`
	if err := os.WriteFile(mockBinPath, []byte(mockScript), 0755); err != nil {
		t.Fatalf("failed writing mock build binary: %v", err)
	}

	destDir := filepath.Join(tempDir, "install_bin")
	downloaderScript := filepath.Join(repoRoot, "scripts", "download-agentic-lens.sh")

	cmd := exec.Command(downloaderScript, "--dest", destDir, "--local-bin", mockBinPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("local-bin install failed: %v, stderr: %s", err, stderr.String())
	}

	// Verify wrapper is installed
	wrapperPath := filepath.Join(destDir, "agentic-lens")
	info, err := os.Stat(wrapperPath)
	if err != nil {
		t.Fatalf("agentic-lens wrapper was not created: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("wrapper script is not executable")
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(destDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); err != nil {
		t.Errorf("protective .gitignore was not created: %v", err)
	}

	// Verify wrapper execution works
	execCmd := exec.Command(wrapperPath)
	execCmd.Dir = destDir
	execCmd.Env = append(os.Environ(), "AGENTIC_LENS_NO_UPDATE_CHECK=1")
	var execOut bytes.Buffer
	execCmd.Stdout = &execOut
	if err := execCmd.Run(); err != nil {
		t.Fatalf("installed wrapper execution failed: %v", err)
	}
	if !strings.Contains(execOut.String(), "mock binary local installation test") {
		t.Errorf("unexpected output from installed wrapper: %s", execOut.String())
	}
}

func TestDownloader_SkillScriptParity(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	skillDownloader := filepath.Join(repoRoot, ".opencode", "skills", "agentic-lens-setup", "scripts", "download.sh")
	if _, err := os.Stat(skillDownloader); err != nil {
		t.Fatalf("skill downloader script not found: %v", err)
	}

	cmd := exec.Command(skillDownloader, "--dry-run", "--all-platforms")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("skill downloader dry-run failed: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "agentic-lens-darwin-arm64") || !strings.Contains(out, "agentic-lens-linux-amd64") {
		t.Errorf("expected multi-platform binary list in skill downloader output, got: %s", out)
	}
}

func TestZeroEmojisAcrossScriptsAndGuides(t *testing.T) {
	repoRoot, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed finding repo root: %v", err)
	}

	filesToCheck := []string{
		filepath.Join(repoRoot, "scripts", "agentic-lens-wrapper.sh"),
		filepath.Join(repoRoot, "scripts", "download-agentic-lens.sh"),
		filepath.Join(repoRoot, ".opencode", "skills", "agentic-lens-setup", "scripts", "agentic-lens-wrapper.sh"),
		filepath.Join(repoRoot, ".opencode", "skills", "agentic-lens-setup", "scripts", "download.sh"),
		filepath.Join(repoRoot, ".opencode", "skills", "agentic-lens-setup", "SKILL.md"),
		filepath.Join(repoRoot, "docs", "SETUP_MANUAL.md"),
		filepath.Join(repoRoot, "alspaces", "GEN-7", "CONTRACT.md"),
	}

	emojiRegex := regexp.MustCompile(`[\x{1F300}-\x{1F9FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{1F1E6}-\x{1F1FF}]`)

	for _, file := range filesToCheck {
		t.Run(filepath.Base(file), func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed reading file %s: %v", file, err)
			}
			if emojiRegex.Match(content) {
				t.Errorf("file %s contains forbidden emojis", file)
			}
		})
	}
}

func filterEnv(environ []string, keyToOmit string) []string {
	prefix := keyToOmit + "="
	var res []string
	for _, env := range environ {
		if !strings.HasPrefix(env, prefix) {
			res = append(res, env)
		}
	}
	return res
}
