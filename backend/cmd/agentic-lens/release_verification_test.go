package main_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseStable_SemVerRegexValidation(t *testing.T) {
	// The regex pattern enforced in .github/workflows/release-stable.yml
	semVerRegex := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

	validCases := []struct {
		tag  string
		desc string
	}{
		{"v0.1.0", "initial development version"},
		{"v1.0.0", "first major release"},
		{"v2.10.3", "multi-digit minor and patch"},
		{"v100.200.300", "large version numbers"},
		{"v0.0.1", "patch only release"},
		{"v99.99.99", "double digit components"},
	}

	for _, tc := range validCases {
		t.Run("valid_"+tc.tag, func(t *testing.T) {
			if !semVerRegex.MatchString(tc.tag) {
				t.Errorf("expected tag %q (%s) to match SemVer regex, but it was rejected", tc.tag, tc.desc)
			}
		})
	}

	invalidCases := []struct {
		tag  string
		desc string
	}{
		{"1.0.0", "missing v prefix"},
		{"v1.0", "missing patch component"},
		{"v1", "missing minor and patch components"},
		{"v1.0.0-beta", "prerelease tag rejected by stable workflow"},
		{"v1.0.0-rc.1", "release candidate rejected by stable workflow"},
		{"v1.0.0+build.123", "build metadata suffix rejected by stable workflow"},
		{"v1.0.0.0", "four components"},
		{"v-1.0.0", "negative component"},
		{"v1.0.0-dirty", "dirty git suffix"},
		{"latest", "floating tag string"},
		{"v", "only v prefix"},
		{"", "empty string"},
		{" v1.0.0", "leading space"},
		{"v1.0.0 ", "trailing space"},
		{"v1.0.0\n", "trailing newline"},
		{"v1.a.0", "alphanumeric component"},
	}

	for _, tc := range invalidCases {
		t.Run("invalid_"+tc.tag, func(t *testing.T) {
			if semVerRegex.MatchString(tc.tag) {
				t.Errorf("expected tag %q (%s) to be rejected by SemVer regex, but it matched", tc.tag, tc.desc)
			}
		})
	}
}

func TestReleaseCandidate_VersionFormatRegex(t *testing.T) {
	// Pattern enforced for release candidates: vMAJOR.MINOR.PATCH-rc.RUN_NUMBER+SHORT_SHA
	rcRegex := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+\+[0-9a-fA-F]+$`)

	validCases := []struct {
		version string
		desc    string
	}{
		{"v0.1.0-rc.1+a1b2c3d", "standard rc with run 1 and 7-char sha"},
		{"v1.0.0-rc.42+7b8a9c0", "rc with run 42"},
		{"v2.5.12-rc.100+abcdef0", "rc with 3-digit run number"},
		{"v0.2.0-rc.1+12345678", "8-character git sha"},
	}

	for _, tc := range validCases {
		t.Run("valid_"+tc.version, func(t *testing.T) {
			if !rcRegex.MatchString(tc.version) {
				t.Errorf("expected version %q (%s) to match RC regex, but it failed", tc.version, tc.desc)
			}
		})
	}

	invalidCases := []struct {
		version string
		desc    string
	}{
		{"v0.1.0", "stable version missing rc suffix"},
		{"v0.1.0-rc.1", "missing commit sha metadata"},
		{"v0.1.0+a1b2c3d", "missing rc component"},
		{"0.1.0-rc.1+a1b2c3d", "missing v prefix"},
		{"v0.1.0-rc+a1b2c3d", "missing run number in rc suffix"},
		{"v0.1.0-rc.abc+a1b2c3d", "non-numeric run number"},
		{"v0.1.0-rc.1+xyz!@#", "invalid hex characters in sha"},
	}

	for _, tc := range invalidCases {
		t.Run("invalid_"+tc.version, func(t *testing.T) {
			if rcRegex.MatchString(tc.version) {
				t.Errorf("expected version %q (%s) to be rejected by RC regex, but it matched", tc.version, tc.desc)
			}
		})
	}
}

func TestReleaseCandidate_BaseTagDerivation(t *testing.T) {
	// Simulates the base tag extraction logic:
	// BASE_TAG=$(echo "$BASE_TAG" | sed -E 's/(-|\+).*//')
	// if [[ ! "$BASE_TAG" =~ ^v ]]; then BASE_TAG="v${BASE_TAG}"; fi
	stripMetadata := regexp.MustCompile(`(-|\+).*`)

	tests := []struct {
		input    string
		expected string
	}{
		{"v0.1.0", "v0.1.0"},
		{"v1.2.3", "v1.2.3"},
		{"v0.1.0-rc.1", "v0.1.0"},
		{"v0.1.0-rc.42+a1b2c3d", "v0.1.0"},
		{"v2.0.0+build123", "v2.0.0"},
		{"v0.3.0-alpha.1", "v0.3.0"},
		{"0.1.0", "v0.1.0"},
		{"1.0.0-rc.2", "v1.0.0"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			cleaned := stripMetadata.ReplaceAllString(tc.input, "")
			if !strings.HasPrefix(cleaned, "v") {
				cleaned = "v" + cleaned
			}
			if cleaned != tc.expected {
				t.Errorf("base tag derivation for %q = %q, want %q", tc.input, cleaned, tc.expected)
			}
		})
	}
}

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

func TestChecksumVerification_DetectsCorruptedArchives(t *testing.T) {
	tempDir := t.TempDir()

	archiveName := "agentic-lens-test-linux-amd64.tar.gz"
	archivePath := filepath.Join(tempDir, archiveName)
	originalContent := []byte("valid mock archive content 123456789")
	if err := os.WriteFile(archivePath, originalContent, 0644); err != nil {
		t.Fatalf("failed writing test archive: %v", err)
	}

	hash := sha256.Sum256(originalContent)
	hashHex := fmt.Sprintf("%x", hash)
	sumsContent := fmt.Sprintf("%s  %s\n", hashHex, archiveName)
	sumsPath := filepath.Join(tempDir, "SHA256SUMS")
	if err := os.WriteFile(sumsPath, []byte(sumsContent), 0644); err != nil {
		t.Fatalf("failed writing SHA256SUMS: %v", err)
	}

	// Verify original content matches
	cmd := exec.Command("sha256sum", "-c", "SHA256SUMS")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("sha256sum verification failed on unmodified file: %v", err)
	}

	// Corrupt the archive content
	corruptedContent := []byte("corrupted modified archive payload 987654321")
	if err := os.WriteFile(archivePath, corruptedContent, 0644); err != nil {
		t.Fatalf("failed overwriting test archive: %v", err)
	}

	// Verify corrupted content fails checksum check
	corruptCmd := exec.Command("sha256sum", "-c", "SHA256SUMS")
	corruptCmd.Dir = tempDir
	var stderr bytes.Buffer
	corruptCmd.Stderr = &stderr
	runErr := corruptCmd.Run()
	if runErr == nil {
		t.Fatalf("expected sha256sum to fail on corrupted archive, but it succeeded")
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
