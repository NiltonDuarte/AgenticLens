package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func captureAdmissionFixture(t *testing.T) (captureGenerationFixture, string, string) {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	seedCaptureTicketHome(t, f, "TEST-100")
	f.write(t, 11)
	if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: true\n")); err != nil {
		t.Fatal(err)
	}
	return f, filepath.Dir(filepath.Dir(f.state)), filepath.Join(t.TempDir(), "new-state")
}

func admissionRequest(f captureGenerationFixture) *exec.Cmd {
	command := exec.Command(cliBinaryPath, "capture", "request", "--repo", f.project, "--source", "codex", "--store", f.store, "--session-id", "root")
	command.Dir = f.project
	return command
}

func admissionUpdate(f captureGenerationFixture, next, output string) *exec.Cmd {
	if output == "" {
		output = filepath.Join(f.project, ".agenticlens-config.local.yaml")
	}
	args := []string{"init", "--update-capture-config", "--capture-state-dir", next}
	if output != "" {
		args = append(args, "--output", output)
	}
	command := exec.Command(cliBinaryPath, args...)
	command.Dir = f.project
	return command
}

func admissionBind(t *testing.T, f captureGenerationFixture, request string) {
	t.Helper()
	command := exec.Command(cliBinaryPath, "capture", "bind", "--repo", f.project, "--source", "codex", "--store", f.store, "--session-id", "root", "--request", request, "--ticket-id", "TEST-100", "--ticket-title", "Admission ownership", "--ticket-type", "feat")
	command.Dir = f.project
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("accepted request token became detached: %v %s", err, output)
	}
	// Drain the actual publisher process before TempDir removes its registry.
	policy, err := adaptersconfig.ResolveCapture(f.project, "")
	if err != nil {
		t.Fatal(err)
	}
	state, err := capture.ProjectState(policy.StateDir, f.project)
	if err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	if _, err := store.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.Activity = "idle"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		request, err := store.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		progress, err := capture.WorkerStatus(store.CoordinationDirectory())
		if err != nil {
			t.Fatal(err)
		}
		if request.Generation == request.Acknowledged && progress.PID > 0 {
			exited, err := captureNativeProcessExited(progress.PID)
			if err != nil {
				t.Fatal(err)
			}
			if exited {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("admitted native worker did not finish before fixture removal")
}

func admissionToken(t *testing.T, output []byte) string {
	t.Helper()
	var response struct {
		Request string `json:"request"`
	}
	if err := json.Unmarshal(output, &response); err != nil || response.Request == "" {
		t.Fatalf("missing accepted request token: %v %s", err, output)
	}
	return response.Request
}

func TestNativeCaptureRequestAndConfigUpdateCannotStrandBinding(t *testing.T) {
	for i := 0; i < 12; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			f, oldBase, next := captureAdmissionFixture(t)
			request, update := admissionRequest(f), admissionUpdate(f, next, "")
			var requestOut, updateOut bytes.Buffer
			request.Stdout = &requestOut
			request.Stderr = &requestOut
			update.Stdout = &updateOut
			update.Stderr = &updateOut
			if err := request.Start(); err != nil {
				t.Fatal(err)
			}
			if err := update.Start(); err != nil {
				if e := request.Wait(); e != nil {
					t.Error(e)
				}
				t.Fatal(err)
			}
			requestErr, updateErr := request.Wait(), update.Wait()
			if requestErr != nil {
				t.Fatalf("capture admission failed: %v %s", requestErr, requestOut.Bytes())
			}
			policy, err := adaptersconfig.ResolveCapture(f.project, "")
			if err != nil {
				t.Fatal(err)
			}
			expected := next
			if updateErr != nil {
				expected = oldBase
			}
			if policy.StateDir != expected {
				t.Fatalf("update result/config mismatch: %v %s %+v", updateErr, updateOut.Bytes(), policy)
			}
			registry, err := capture.ProjectState(expected, f.project)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(registry, "registry.db")); err != nil {
				t.Fatalf("accepted handoff is outside selected config: %v", err)
			}
			admissionBind(t, f, admissionToken(t, requestOut.Bytes()))
		})
	}
}

func TestCaptureAdmissionBeforeUpdateRetainsDurableOldHandoff(t *testing.T) {
	f, _, next := captureAdmissionFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lease, err := adaptersconfig.AcquireCaptureAdmission(ctx, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lease.Close(); err != nil {
			t.Error(err)
		}
	}()
	policy, err := lease.ResolveCapture("")
	if err != nil {
		t.Fatal(err)
	}
	command := admissionUpdate(f, next, "")
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	// Allocate and accept the exact first handoff under the same typed lease that
	// native callbacks use; config maintenance cannot check absence in this span.
	registry, err := capture.ProjectState(policy.StateDir, f.project)
	if err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(registry, f.project)
	if err != nil {
		t.Fatal(err)
	}
	row, registerErr := store.Register(ctx, domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}, "", "agent_context_request")
	closeErr := store.Close()
	if registerErr != nil || closeErr != nil {
		t.Fatalf("handoff: %v %v", registerErr, closeErr)
	}
	if err := lease.Close(); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Fatalf("path-only update detached an admitted database: %s", output.Bytes())
	}
	admissionBind(t, f, row.Request)
}

func TestConfigUpdateAdmissionBeforeNativeRequestSelectsNewState(t *testing.T) {
	f, _, next := captureAdmissionFixture(t)
	path := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	preview, err := adaptersconfig.PrepareCaptureConfigUpdate(path, next)
	if err != nil {
		t.Fatal(err)
	}
	checked, activate := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		result <- adaptersconfig.WriteCaptureConfigUpdateChecked(ctx, path, preview, func(change adaptersconfig.CaptureConfigChange) error {
			before, err := adaptersconfig.ResolveCaptureStatePath(change.ProjectRoot, change.PreviousStateDir)
			if err != nil {
				return err
			}
			after, err := adaptersconfig.ResolveCaptureStatePath(change.ProjectRoot, change.StateDir)
			if err != nil {
				return err
			}
			if err := capture.CheckStateRelocation(change.ProjectRoot, before, after); err != nil {
				return err
			}
			close(checked)
			select {
			case <-activate:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-checked:
	case err := <-result:
		t.Fatalf("update admission failed: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	command := admissionRequest(f)
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		close(activate)
		t.Fatal(err)
	}
	close(activate)
	if err := <-result; err != nil {
		if e := command.Wait(); e != nil {
			t.Error(e)
		}
		t.Fatal(err)
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("request did not resolve after update: %v %s", err, output.Bytes())
	}
	policy, err := adaptersconfig.ResolveCapture(f.project, "")
	if err != nil || policy.StateDir != next {
		t.Fatalf("wrong selected state: %+v %v", policy, err)
	}
	admissionBind(t, f, admissionToken(t, output.Bytes()))
	if _, err := os.Stat(filepath.Join(f.state, "registry.db")); !os.IsNotExist(err) {
		t.Fatalf("default-derived stale state was admitted: %v", err)
	}
}

func TestCaptureConfigUpdateGuardsEffectiveLocalSelection(t *testing.T) {
	f, oldBase, next := captureAdmissionFixture(t)
	localPath := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	local := []byte(fmt.Sprintf("agenticlens:\n  state_dir: %q\ntelemetry:\n  capture:\n    max_sessions: 256\n    max_source_bytes: 67108864\n    commit_warning: true\n    enabled: true\n    state_dir: %q\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    pass_timeout_ms: 3000\n", filepath.Join(f.project, ".publication-state"), oldBase))
	local = append(local, []byte(fmt.Sprintf("  default_source: codex\n  sources:\n    codex:\n      home: %q\n", f.store))...)
	if err := os.WriteFile(localPath, local, 0600); err != nil {
		t.Fatal(err)
	}
	request, err := admissionRequest(f).CombinedOutput()
	if err != nil {
		t.Fatalf("request: %v %s", err, request)
	}
	if output, err := admissionUpdate(f, next, filepath.Join(f.project, ".agenticlens-config.yaml")).CombinedOutput(); err == nil {
		t.Fatalf("shared entrypoint bypassed selected populated local state: %s", output)
	}
	if output, err := admissionUpdate(f, next, localPath).CombinedOutput(); err == nil {
		t.Fatalf("local-only movement abandoned accepted work: %s", output)
	}
	current, err := os.ReadFile(localPath)
	if err != nil || !bytes.Equal(current, local) {
		t.Fatalf("refused local replacement modified config: %v", err)
	}
	admissionBind(t, f, admissionToken(t, request))
}

func TestCaptureAdmissionHonorsExistingShortCommandDeadline(t *testing.T) {
	f, _, _ := captureAdmissionFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lease, err := adaptersconfig.AcquireCaptureAdmission(ctx, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lease.Close(); err != nil {
			t.Error(err)
		}
	}()
	command := admissionRequest(f)
	command.Args = append(command.Args, "--timeout", "40ms")
	started := time.Now()
	output, err := command.CombinedOutput()
	if err == nil || !bytes.Contains(output, []byte("context deadline exceeded")) {
		t.Fatalf("held admission did not respect caller deadline: %v %s", err, output)
	}
	if time.Since(started) > 1500*time.Millisecond {
		t.Fatalf("short callback acquired a fresh foreground timeout: %s", time.Since(started))
	}
	if _, err := os.Stat(filepath.Join(f.state, "registry.db")); !os.IsNotExist(err) {
		t.Fatalf("timed-out admission allocated state: %v", err)
	}
}
