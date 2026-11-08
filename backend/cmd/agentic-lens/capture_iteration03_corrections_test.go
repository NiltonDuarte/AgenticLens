package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

type correctionCaptureStatus struct {
	InstallationIssues []string             `json:"installation_issues"`
	Verification       domain.CaptureHealth `json:"verification"`
}

func correctionCaptureHistory(t *testing.T) captureGenerationFixture {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n", f.state)
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(f.project, ".codex"), 0700); err != nil {
		t.Fatal(err)
	}
	foreign := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"user-owned sentinel"}]}]},"custom":true}`
	if err := os.WriteFile(filepath.Join(f.project, ".codex", "hooks.json"), []byte(foreign), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store}
	if out, err := captureCLI(t, f.project, "", append([]string{"install"}, args...)...); err != nil {
		t.Fatalf("initial installation: %v %s", err, out)
	}
	completion := captureWorkerCompletion(t, f.coordination(t))
	t.Cleanup(func() { captureCleanupWorkers(t, completion) })
	path := f.write(t, 10)
	payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, string(payload), append([]string{"receive"}, args...)...); err != nil {
		t.Fatalf("actual callback: %v %s", err, out)
	}
	requestArgs := append(append([]string{"request"}, args...), "--session-id", "root")
	out, err := captureCLI(t, f.project, "", requestArgs...)
	if err != nil {
		t.Fatalf("exact request: %v %s", err, out)
	}
	var row domain.CaptureRegistration
	if err := json.Unmarshal(out, &row); err != nil || row.Request == "" {
		t.Fatalf("binding token unavailable: %v %s", err, out)
	}
	bind := append(append([]string{"bind"}, args...), "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "History fixture", "--ticket-type", "feat", "--request", row.Request)
	if out, err := captureCLI(t, f.project, "", bind...); err != nil {
		t.Fatalf("explicit response: %v %s", err, out)
	}
	endPayload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionEnd", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, string(endPayload), append([]string{"receive"}, args...)...); err != nil {
		t.Fatalf("end fixture activity: %s %v", out, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := completion(ctx); err != nil {
		t.Fatalf("fixture native writers still active: %v", err)
	}
	f.generate(t)
	return f
}

func correctionInspection(t *testing.T, f captureGenerationFixture, command string) correctionCaptureStatus {
	t.Helper()
	out, err := correctionSelectedCLI(t, f, command, "--repo", f.project, "--state-dir", f.state)
	if err != nil {
		t.Fatalf("public %s must remain inspectable: %v %s", command, err, out)
	}
	var status correctionCaptureStatus
	if command == "verify" {
		err = json.Unmarshal(out, &status.Verification)
	} else {
		err = json.Unmarshal(out, &status)
	}
	if err != nil {
		t.Fatalf("%s JSON: %v %s", command, err, out)
	}
	return status
}

func correctionSelectedCLI(t *testing.T, f captureGenerationFixture, args ...string) ([]byte, error) {
	t.Helper()
	wrapper := fixtureWrapperFromHook(t, f.project)
	if args[0] == "install" {
		args = append(append([]string{}, args...), "--wrapper", wrapper)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, wrapper, append([]string{"--al-offline", "--al-project", f.project, "capture"}, args...)...)
	command.Dir = f.project
	return command.CombinedOutput()
}

func TestCaptureStaleProvenanceAllowsPublicInspectionAndOwnedRepair(t *testing.T) {
	newVersion := "v9.8.8"
	newNative := filepath.Join(t.TempDir(), "selected-native-b")
	build := exec.Command("go", "build", "-ldflags", "-X agenticlens/internal/domain.Version="+newVersion, "-o", newNative, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("actual native B fixture: %v %s", err, out)
	}
	for _, drift := range []string{"template", "actual release switch"} {
		for _, operation := range []string{"install", "disable", "uninstall", "uninstall-check"} {
			t.Run(drift+"/"+operation, func(t *testing.T) {
				f := correctionCaptureHistory(t)
				if operation == "uninstall-check" {
					if out, err := exec.Command("git", "-C", f.project, "init", "--quiet").CombinedOutput(); err != nil {
						t.Fatalf("disposable Git init: %v %s", err, out)
					}
					if out, err := captureCLI(t, f.project, "", "install-check", "--repo", f.project, "--state-dir", f.state, "--wrapper", fixtureWrapperFromHook(t, f.project)); err != nil {
						t.Fatalf("owned advisory setup: %v %s", err, out)
					}
				}
				if drift == "actual release switch" {
					shared := filepath.Join(f.project, ".agenticlens-config.yaml")
					body := bytes.Replace(readCaptureFile(t, shared), []byte("version: "+domain.DefaultVersion+"\n"), []byte("version: "+newVersion+"\n"), 1)
					if err := os.WriteFile(shared, body, 0600); err != nil {
						t.Fatal(err)
					}
					entry, err := adaptersconfig.ResolveEntrypoint(f.project)
					if err != nil {
						t.Fatal(err)
					}
					selected, err := entry.NativePath()
					if err != nil {
						t.Fatal(err)
					}
					if err := os.MkdirAll(filepath.Dir(selected), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(newNative, selected); err != nil {
						t.Fatal(err)
					}
					var previous capture.Manifest
					if err := json.Unmarshal(readCaptureFile(t, filepath.Join(f.coordination(t), "install-codex.json")), &previous); err != nil {
						t.Fatal(err)
					}
					if previous.Version == newVersion || previous.Binary == newNative {
						t.Fatalf("old installation was unexpectedly relabeled before repair: %+v", previous)
					}
					command := exec.Command(fixtureWrapperFromHook(t, f.project), "--al-offline", "--al-project", f.project, "version", "--short")
					if out, err := command.CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != newVersion {
						t.Fatalf("current wrapper did not execute actual native B: %v %s", err, out)
					}
				} else {
					for _, path := range []string{filepath.Join(f.project, ".codex", "agenticlens-installation.json"), filepath.Join(f.coordination(t), "install-codex.json")} {
						var manifest capture.Manifest
						if err := json.Unmarshal(readCaptureFile(t, path), &manifest); err != nil {
							t.Fatal(err)
						}
						manifest.TemplateVersion = "previous-owned-template"
						body, err := json.Marshal(manifest)
						if err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(path, body, 0600); err != nil {
							t.Fatal(err)
						}
					}
				}
				ledger := readCaptureFile(t, f.output)
				registry := readCaptureFile(t, filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))
				for _, command := range []string{"status", "diagnostics", "verify"} {
					status := correctionInspection(t, f, command)
					if status.Verification.State != "degraded" || (command != "verify" && len(status.InstallationIssues) == 0) {
						t.Fatalf("%s hid stale provenance: %+v", command, status)
					}
				}
				if !bytes.Equal(ledger, readCaptureFile(t, f.output)) || !bytes.Equal(registry, readCaptureFile(t, filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))) {
					t.Fatal("stale read-only inspection changed accepted facts")
				}

				args := []string{operation, "--repo", f.project, "--state-dir", f.state}
				if operation != "uninstall-check" {
					args = append(args, "--source", "codex", "--store", f.store)
				}
				if out, err := correctionSelectedCLI(t, f, args...); err != nil {
					t.Fatalf("stale preflight blocked owned %s: %v %s", operation, err, out)
				}
				if !bytes.Equal(ledger, readCaptureFile(t, f.output)) {
					t.Fatal("repair/removal changed accepted canonical metrics")
				}
				if !bytes.Contains(readCaptureFile(t, filepath.Join(f.project, ".codex", "hooks.json")), []byte("user-owned sentinel")) {
					t.Fatal("owned repair/removal changed unrelated hook")
				}
				if operation == "install" {
					status := correctionInspection(t, f, "status")
					if len(status.InstallationIssues) != 0 {
						t.Fatalf("owned refresh did not repair provenance: %+v", status.InstallationIssues)
					}
					if drift == "actual release switch" {
						manifest, err := capture.ReadManifest(f.coordination(t), "codex")
						if err != nil || manifest.Version != newVersion || manifest.Binary != newNative || status.Verification.State == "verified" {
							t.Fatalf("new installation identity/runtime verification conflated: %+v %+v %v", manifest, status.Verification, err)
						}
					}
				}
			})
		}
	}
}

func TestCaptureMissingIntegrationDiagnosesDriftWithoutErasingHistoricalExchange(t *testing.T) {
	f := correctionCaptureHistory(t)
	before := correctionInspection(t, f, "verify").Verification
	if before.State != "verified" || !before.CallbackObserved || !before.BindingAccepted || !before.SourceAccepted || !before.PublicationAccepted {
		t.Fatalf("historical native exchange was not established: %+v", before)
	}
	ledger := readCaptureFile(t, f.output)
	registry := readCaptureFile(t, filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))
	config := filepath.Join(f.project, ".codex", "hooks.json")
	if err := os.Remove(config); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"status", "diagnostics", "verify"} {
		status := correctionInspection(t, f, command)
		h := status.Verification
		if h.State != "degraded" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted || (command != "verify" && len(status.InstallationIssues) == 0) {
			t.Fatalf("%s hid live drift or erased historical facts: %+v", command, status)
		}
	}
	if out, err := captureCLI(t, f.project, "", "check", "--repo", f.project, "--state-dir", f.state); err != nil || bytes.Count(out, []byte("AgenticLens capture advisory:")) != 1 {
		t.Fatalf("missing owned lifecycle integration not advisory: %v %s", err, out)
	}
	if _, err := os.Stat(config); !os.IsNotExist(err) {
		t.Fatalf("inspection recreated missing integration: %v", err)
	}
	if !bytes.Equal(ledger, readCaptureFile(t, f.output)) || !bytes.Equal(registry, readCaptureFile(t, filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))) {
		t.Fatal("missing-integration diagnosis rewrote accepted facts")
	}
	// Removing callback configuration can be a user's deliberate edit. Public
	// repair must reach the owning adapter and preserve that conflict, rather
	// than restoring removed hooks or failing the compatibility preflight.
	if out, err := captureCLI(t, f.project, "", "install", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store); err == nil || !bytes.Contains(out, []byte("edited or removed")) || !bytes.Contains(out, []byte("configuration preserved")) {
		t.Fatalf("missing owned groups bypassed ownership conflict: %v %s", err, out)
	}
	if _, err := os.Stat(config); !os.IsNotExist(err) || !bytes.Equal(ledger, readCaptureFile(t, f.output)) {
		t.Fatalf("rejected repair recreated removed hooks or changed metrics: %v", err)
	}
}

func TestCapturePendingSourceRetryRetainsFailureUntilExactRead(t *testing.T) {
	f := correctionCaptureHistory(t)
	healthy := func() {
		t.Helper()
		h := correctionInspection(t, f, "verify").Verification
		if h.State != "verified" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
			t.Fatalf("actual exchange is not healthy: %+v", h)
		}
		if out, err := correctionSelectedCLI(t, f, "check", "--repo", f.project, "--state-dir", f.state); err != nil || len(out) != 0 {
			t.Fatalf("ordinary accepted activity became advisory: %v %s", err, out)
		}
	}
	reconcile := func(budget int64, want domain.CaptureState) domain.CaptureRegistration {
		t.Helper()
		if budget == 0 {
			budget = capture.DefaultByteBudget
		}
		store, err := capture.OpenProject(f.state, f.project)
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()
		result, err := store.Reconcile(context.Background(), "TEST-100", "tester", capture.DefaultSessionBudget, budget)
		if err != nil {
			t.Fatalf("source reconciliation: %v", err)
		}
		if len(result.Registrations) != 1 || result.Registrations[0].Identity.SessionID != "root" || result.Registrations[0].State != want {
			t.Fatalf("did not reach actual %s source outcome: %+v", want, result)
		}
		if budget == 1 && result.SourceBytes != 0 {
			t.Fatalf("pending retry unexpectedly accepted source bytes: %+v", result)
		}
		return result.Registrations[0]
	}
	healthy()
	// Pending processing alone is quiet when no confirmed failure exists.
	ordinary := reconcile(1, domain.CapturePending)
	healthy()
	path := filepath.Join(f.store, "sessions", "root.jsonl")
	body := readCaptureFile(t, path)
	ledger := readCaptureFile(t, f.output)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	reconcile(0, domain.CaptureFailed)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	pending := reconcile(1, domain.CapturePending)
	if pending.RefreshedAt != ordinary.RefreshedAt {
		t.Fatalf("budget-pending retry acknowledged a successful read: before=%s after=%s", ordinary.RefreshedAt, pending.RefreshedAt)
	}
	out, err := correctionSelectedCLI(t, f, "status", "--repo", f.project, "--state-dir", f.state)
	if err != nil {
		t.Fatalf("retained failure status: %v %s", err, out)
	}
	var status struct {
		Diagnostics []domain.CaptureDiagnostic `json:"diagnostics"`
		More        bool                       `json:"current_diagnostics_continuation"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		t.Fatal(err)
	}
	failurePresent := false
	for _, diagnostic := range status.Diagnostics {
		if diagnostic.Phase == "reconciliation" && diagnostic.Outcome == "failed" && diagnostic.Reconciliation != nil {
			for _, owner := range diagnostic.Reconciliation.Owners {
				if owner.Identity == pending.Identity && owner.State == domain.CaptureFailed {
					failurePresent = true
				}
			}
		}
	}
	if status.More || !failurePresent {
		t.Fatalf("fixture did not expose complete current source-failure evidence: %s", out)
	}
	if out, err := correctionSelectedCLI(t, f, "check", "--repo", f.project, "--state-dir", f.state); err != nil || !bytes.Contains(out, []byte("source refresh failed")) || bytes.Count(out, []byte("AgenticLens capture advisory:")) != 1 {
		t.Fatalf("pending retry hid confirmed failure in native advisory: %v %s", err, out)
	}
	h := correctionInspection(t, f, "verify").Verification
	if h.State != "degraded" || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
		t.Fatalf("pending retry hid source failure or erased historical exchange: %+v", h)
	}
	if !bytes.Equal(ledger, readCaptureFile(t, f.output)) {
		t.Fatal("failed/pending source inspection changed accepted metrics")
	}
	reconcile(0, domain.CaptureCurrent)
	healthy()
}

func TestCaptureSamePinOtherProjectInstallPreservesAcknowledgedActiveWorker(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	cache, wrappers := t.TempDir(), t.TempDir()
	header := fmt.Sprintf("agenticlens:\n  version: %s\n  install_root: %q\n", domain.DefaultVersion, cache)
	shared := append([]byte(header), readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml"))...)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 1000\n    max_lifetime_ms: 20000\n", f.state)
	localPath := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	if err := os.WriteFile(localPath, []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	checksum := fmt.Sprintf("%x", sha256.Sum256(readCaptureFile(t, cliBinaryPath)))
	installer := filepath.Join("..", "..", "..", ".agents", "skills", "agentic-lens-setup", "scripts", "download.sh")
	install := func(project string, captureSource bool) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()
		args := []string{installer, "--project", project, "--local-bin", cliBinaryPath, "--sha256", checksum, "--dest", wrappers, "--force"}
		if captureSource {
			args = append(args, "--capture-source", "codex", "--capture-store", f.store, "--capture-state-dir", f.state)
		}
		command := exec.CommandContext(ctx, "bash", args...)
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("production explicit installer: %v %s", err, out)
		}
	}
	install(f.project, true)
	entry, err := adaptersconfig.ResolveEntrypoint(f.project)
	if err != nil {
		t.Fatal(err)
	}
	native, err := entry.NativePath()
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(native)
	if err != nil {
		t.Fatal(err)
	}
	f.bind(t, f.write(t, 10))
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	if _, err := store.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.Activity, row.Event = "busy", "UserPromptSubmit"
		row.TicketTitle, row.TicketType = "Active same-pin collection", "feat"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte(strings.Replace(local, "enabled: false", "enabled: true", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	worker := exec.CommandContext(ctx, native, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	log, err := os.CreateTemp(t.TempDir(), "active-worker")
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	defer func() {
		if err := log.Close(); err != nil {
			t.Error(err)
		}
	}()
	worker.Stdout, worker.Stderr = log, log
	if err := worker.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer cancel()
		if err := os.WriteFile(localPath, []byte(local), 0600); err != nil {
			t.Error(err)
		}
		done := make(chan error, 1)
		go func() { done <- worker.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("active worker: %v %s", err, readCaptureFile(t, log.Name()))
			}
		case <-time.After(4 * time.Second):
			cancel()
			<-done
			t.Error("active worker did not stop after explicit disable")
		}
	})
	waitCorrectionLedger(t, f.output, "root", 10)
	deadline := time.Now().Add(4 * time.Second)
	var accepted capture.PublicationRequest
	for time.Now().Before(deadline) {
		accepted, err = store.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if accepted.Generation > 0 && accepted.Generation == accepted.Acknowledged {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if accepted.Generation == 0 || accepted.Generation != accepted.Acknowledged {
		t.Fatalf("active worker had no accepted snapshot: %+v", accepted)
	}
	second := t.TempDir()
	if err := os.WriteFile(filepath.Join(second, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	install(second, false)
	after, err := os.Stat(native)
	if err != nil || !os.SameFile(before, after) || !before.ModTime().Equal(after.ModTime()) {
		t.Fatalf("identical same-pin install changed active native identity: %v", err)
	}
	f.writeAt(t, 31, "2026-01-01T00:00:03Z")
	// No callback or new generation: the already acknowledged active owner must
	// keep refreshing after another project installs that exact cached release.
	waitCorrectionLedger(t, f.output, "root", 31)
	progress, err := capture.WorkerStatus(f.coordination(t))
	if err != nil || progress.PID != worker.Process.Pid {
		t.Fatalf("other project's identical install retired active owner: %+v %v", progress, err)
	}
	request, err := store.PublicationRequest(context.Background())
	if err != nil || request.Generation != accepted.Generation || request.Acknowledged != accepted.Acknowledged {
		t.Fatalf("source growth needed a replacement event/generation: before=%+v after=%+v %v", accepted, request, err)
	}
}
