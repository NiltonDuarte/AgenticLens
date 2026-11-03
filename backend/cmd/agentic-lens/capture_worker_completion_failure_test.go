package main_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
)

func TestCaptureWorkerCompletionSurvivesFixtureEarlyReturn(t *testing.T) {
	var directory string
	drained := false
	t.Run("returns with native still blocked", func(t *testing.T) {
		f := newCaptureGenerationFixture(t)
		directory = f.project
		if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte("telemetry:\n  capture:\n    enabled: false\n"), 0600); err != nil {
			t.Fatal(err)
		}
		store, err := capture.OpenProject(f.state, f.project)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				t.Error(err)
			}
		}()
		scoped := store.CoordinationDirectory()
		completion := captureWorkerCompletion(t, scoped)
		admission, err := adaptersconfig.AcquireCaptureAdmission(context.Background(), f.project)
		if err != nil {
			t.Fatal(err)
		}
		// Register before launching: a return before normal shutdown must still
		// finish the delayed native before testing removes its directories.
		t.Cleanup(func() {
			defer func() {
				if err := admission.Close(); err != nil {
					t.Error(err)
				}
				captureCleanupWorkers(t, completion)
			}()
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			fallback := make(chan struct{})
			result := make(chan error, 1)
			go func() {
				result <- captureDrainWorkerCompletion(ctx, func(waitCtx context.Context) error {
					if waitCtx.Err() == nil {
						close(fallback)
					}
					return completion(waitCtx)
				})
			}()
			guard, stop := context.WithTimeout(context.Background(), 5*time.Second)
			defer stop()
			select {
			case <-fallback:
			case err := <-result:
				t.Errorf("canceled wait returned before draining native: %v", err)
				return
			case <-guard.Done():
				t.Error("canceled wait did not enter safe native drain")
				return
			}
			select {
			case err := <-result:
				t.Errorf("safe drain completed while native admission was held: %v", err)
				return
			default:
			}
			if _, err := os.Stat(directory); err != nil {
				t.Errorf("fixture disappeared while native drain was pending: %v", err)
			}
			if err := admission.Close(); err != nil {
				t.Error(err)
			}
			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Errorf("safe drain lost initial canceled wait: %v", err)
				}
			case <-guard.Done():
				t.Error("safe native drain did not finish after admission release")
				return
			}
			if _, err := os.Stat(filepath.Join(scoped, "worker.lock")); err != nil {
				t.Errorf("cleanup completed before delayed native reached lease allocation: %v", err)
				return
			}
			drained = true
		})
		if err := capture.LaunchWorker(cliBinaryPath, f.project, scoped); err != nil {
			t.Fatal(err)
		}
		request, err := store.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if request.Generation != request.Acknowledged {
			t.Fatalf("unexpected pending publication: %+v", request)
		}
		owner, err := capture.AcquireWorker(scoped)
		if err != nil || owner == nil {
			t.Fatalf("delayed native should not yet own worker lease: %v", err)
		}
		if err := owner.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(scoped, "worker.lock")); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := completion(ctx); !errors.Is(err, context.Canceled) {
			t.Fatalf("completion accepted a native still blocked at fixture return: %v", err)
		}
		return
	})
	if !drained {
		t.Fatal("fixture returned without completing native cleanup")
	}
	if _, err := os.Stat(directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("fixture directory was not removed after native completion: %v", err)
	}
}
