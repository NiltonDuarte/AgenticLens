package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestCaptureAdmissionSerializesEffectiveLocalConfigGuard(t *testing.T) {
	root := t.TempDir()
	shared := filepath.Join(root, ".agenticlens-config.yaml")
	local := filepath.Join(root, ".agenticlens-config.local.yaml")
	data, err := fixtureSharedYAML(t, "Example", "records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(shared, data, 0600); err != nil {
		t.Fatal(err)
	}
	localData := []byte("telemetry:\n  capture:\n    enabled: false\n    state_dir: machine-local-state\n")
	if err := os.WriteFile(local, localData, 0600); err != nil {
		t.Fatal(err)
	}
	preview, err := PrepareCaptureConfigUpdate(local, "new-local-state")
	if err != nil {
		t.Fatal(err)
	}
	lease, err := AcquireCaptureAdmission(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lease.Close(); err != nil {
			t.Error(err)
		}
	}()
	var called atomic.Bool
	result := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
		defer cancel()
		result <- WriteCaptureConfigUpdateChecked(ctx, local, preview, func(CaptureConfigChange) error { called.Store(true); return nil })
	}()
	if err := <-result; !errors.Is(err, context.DeadlineExceeded) || called.Load() {
		t.Fatalf("checked guard bypassed active admission: %v guard=%t", err, called.Load())
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if err := WriteCaptureConfigUpdateChecked(context.Background(), local, preview, func(change CaptureConfigChange) error {
		called.Store(true)
		if change.PreviousStateDir != filepath.Join(root, "machine-local-state") || change.StateDir != filepath.Join(root, "new-local-state") || change.ProjectRoot != root {
			t.Fatalf("guard did not inspect effective local overrides: %+v", change)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !called.Load() {
		t.Fatal("operational state guard was omitted")
	}
	cfg, err := Load(root)
	if err != nil || cfg.Telemetry.Capture.StateDir != "new-local-state" {
		t.Fatalf("local setting changed: %+v %v", cfg, err)
	}
	actual, err := os.ReadFile(local)
	if err != nil || string(actual) != string(preview.Data) {
		t.Fatalf("local config was rewritten: %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if admission, err := AcquireCaptureAdmission(cancelled, root); !errors.Is(err, context.Canceled) || admission != nil {
		if admission != nil {
			if e := admission.Close(); e != nil {
				t.Error(e)
			}
		}
		t.Fatalf("cancelled admission succeeded: %v", err)
	}
}
