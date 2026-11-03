package main_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestCaptureInstallReportsPendingAllocationUntilCommittedManifest(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	args := []string{"install", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath}
	invoke := func(dryRun bool) capture.InstallResult {
		t.Helper()
		selected := append([]string(nil), args...)
		if dryRun {
			selected = append(selected, "--dry-run")
		}
		out, err := captureCLI(t, f.project, "", selected...)
		if err != nil {
			t.Fatalf("install dry-run=%t: %v\n%s", dryRun, err, out)
		}
		var result capture.InstallResult
		if err = json.Unmarshal(out, &result); err != nil {
			t.Fatalf("install result: %v\n%s", err, out)
		}
		return result
	}
	for range 2 {
		preview := invoke(true)
		if !preview.StateAllocationPending {
			t.Fatal("unallocated dry-run did not disclose pending allocation")
		}
		if _, err := os.Stat(f.state); !os.IsNotExist(err) {
			t.Fatalf("dry-run allocated registry: %v", err)
		}
	}
	installed := invoke(false)
	if installed.StateAllocationPending {
		t.Fatal("real installation still reports provisional allocation")
	}
	location, err := capture.ExistingLocation(f.state, f.project)
	if err != nil || location.Provisional || location.CoordinationDirectory == "" {
		t.Fatalf("committed location: %+v %v", location, err)
	}
	manifest, err := capture.ReadManifest(location.CoordinationDirectory, "codex")
	if err != nil || !manifest.Active || manifest.StateDir != location.CoordinationDirectory {
		t.Fatalf("manifest did not retain committed scope: %+v %v", manifest, err)
	}
	if strings.Contains(manifest.StateDir, "allocation-pending") || filepath.Base(manifest.StateDir) == "1" {
		t.Fatalf("manifest has a placeholder or numeric allocation: %s", manifest.StateDir)
	}
	preview := invoke(true)
	if preview.StateAllocationPending {
		t.Fatal("existing-scope dry-run still reports pending allocation")
	}
	reopened, err := capture.ExistingLocation(f.state, f.project)
	if err != nil || reopened.CoordinationDirectory != manifest.StateDir {
		t.Fatalf("dry-run changed committed location: %+v %v", reopened, err)
	}
}
