package capture_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestPublicationLeaseFollowsRealRenameAndRejectsOutsideScope(t *testing.T) {
	project := t.TempDir()
	captureProjectConfig(t, project, "Publication movement")
	// Manual publication consumes operational state without interpreting capture policy.
	local := []byte("agenticlens:\n  state_dir: operational-state\ntelemetry:\n  capture:\n    enabled: deliberately-invalid-unused-policy\n")
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), local, 0600); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(project, "records", "arbitrary-home")
	archived := filepath.Join(project, "records", "archive", "renamed-without-ticket")
	if err := os.MkdirAll(original, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(archived), 0700); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(original, "agentic_metrics.json")
	if err := os.WriteFile(ledger, []byte("accepted"), 0600); err != nil {
		t.Fatal(err)
	}
	owner, err := capture.AcquirePublication(context.Background(), ledger)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if err := os.Rename(original, archived); err != nil {
		t.Fatal(err)
	}
	if err := owner.ValidateLocation(); err != nil {
		t.Fatalf("rename within configured scope invalidated held directory: %v", err)
	}
	// A replacement at the old pathname is a distinct directory, never the held destination.
	if err := os.Mkdir(original, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledger, []byte("unrelated replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := capture.WritePublicationFile(ledger, []byte("update held home"), 0600); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(filepath.Join(archived, "agentic_metrics.json")); err != nil || string(body) != "update held home" {
		t.Fatalf("publication lost moved directory: %q %v", body, err)
	}
	if body, err := os.ReadFile(ledger); err != nil || string(body) != "unrelated replacement" {
		t.Fatalf("publication wrote unrelated replacement: %q %v", body, err)
	}
	outside := filepath.Join(project, "outside-selected-root")
	if err := os.Rename(archived, outside); err != nil {
		t.Fatal(err)
	}
	if err := owner.ValidateLocation(); err == nil {
		t.Fatal("outside-scope move remained authorized")
	}
	if _, err := os.Stat(archived); !os.IsNotExist(err) {
		t.Fatalf("validation recreated former archive location: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agenticlens-capture")); !os.IsNotExist(err) {
		t.Fatalf("manual publication initialized capture registry: %v", err)
	}
	if body, err := os.ReadFile(filepath.Join(project, ".agenticlens-config.local.yaml")); err != nil || string(body) != string(local) {
		t.Fatalf("manual publication edited capture/local policy: %v", err)
	}
}

func TestUnavailableOperationalStateStopsBeforeWorkspaceChanges(t *testing.T) {
	project := t.TempDir()
	captureProjectConfig(t, project, "Publication unavailable state")
	state := filepath.Join(project, "state-is-a-file")
	if err := os.WriteFile(state, []byte("unrelated state destination"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte("agenticlens:\n  state_dir: state-is-a-file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(project, "records", "home")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(directory, "agentic_metrics.json")
	accepted := []byte("accepted bytes")
	if err := os.WriteFile(ledger, accepted, 0600); err != nil {
		t.Fatal(err)
	}
	if owner, err := capture.AcquirePublication(context.Background(), ledger); err == nil {
		owner.Close()
		t.Fatal("unavailable state home accepted")
	}
	if body, err := os.ReadFile(ledger); err != nil || string(body) != string(accepted) {
		t.Fatalf("state error changed public bytes: %q %v", body, err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 || entries[0].Name() != "agentic_metrics.json" {
		t.Fatalf("state error staged public transaction files: %v %v", entries, err)
	}
	if body, err := os.ReadFile(state); err != nil || string(body) != "unrelated state destination" {
		t.Fatalf("state initialization overwrote unrelated file: %q %v", body, err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agenticlens-capture")); !os.IsNotExist(err) {
		t.Fatalf("operational-state failure touched capture registry: %v", err)
	}
}
