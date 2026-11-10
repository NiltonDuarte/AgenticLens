package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestCaptureLaterAgentChannelReportsBoundPublicationFailure(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	path := f.write(t, 10)
	f.bind(t, path)
	held, err := capture.AcquireWorker(f.coordination(t))
	if err != nil || held == nil {
		t.Fatalf("fixture worker ownership: %v", err)
	}
	defer func() {
		if held != nil {
			if err := held.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	if out, err := captureCLI(t, f.project, "", "install", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath); err != nil {
		t.Fatalf("installed fixture: %v %s", err, out)
	}
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 1000\n    pass_timeout_ms: 500\n", f.state)
	if err = writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	args := []string{"--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", f.store}
	input, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, f.project, string(input), append([]string{"receive"}, args...)...); err != nil {
		t.Fatalf("initial actual callback: %v %s", err, out)
	}
	out, err := captureCLI(t, f.project, "", append(append([]string{"request"}, args...), "--session-id", "root")...)
	if err != nil {
		t.Fatalf("request: %v %s", err, out)
	}
	var registration domain.CaptureRegistration
	if err = json.Unmarshal(out, &registration); err != nil {
		t.Fatal(err)
	}
	bind := append(append([]string{"bind"}, args...), "--session-id", "root", "--request", registration.Request, "--ticket-id", "TEST-100", "--ticket-title", "Failure feedback", "--ticket-type", "feat")
	if out, err := captureCLI(t, f.project, "", bind...); err != nil {
		t.Fatalf("current installed binding: %v %s", err, out)
	}
	summary := filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
	f.generate(t, "--append", summary)
	before := readCaptureFile(t, f.output)
	beforeSummary := readCaptureFile(t, summary)
	f.writeAt(t, 31, "2026-01-01T00:00:03Z")
	if err = os.Remove(summary); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(summary, 0700); err != nil {
		t.Fatal(err)
	}
	if err = held.Close(); err != nil {
		t.Fatal(err)
	}
	held = nil
	out, err = runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	if err == nil || !bytes.Contains(out, []byte("pending")) {
		t.Fatalf("actual native async publication did not fail: %v %s", err, out)
	}
	if !bytes.Equal(before, readCaptureFile(t, f.output)) {
		t.Fatal("failed async replacement changed accepted ledger")
	}
	failed := false
	for _, diagnostic := range captureDiagnostics(t, f.project, f.state) {
		if diagnostic.TicketID == "TEST-100" && diagnostic.Phase == "publication" && diagnostic.Outcome == "failed" {
			failed = true
		}
	}
	if !failed {
		t.Fatal("real worker failure was not durably diagnosed")
	}
	held, err = capture.AcquireWorker(f.coordination(t))
	if err != nil || held == nil {
		t.Fatalf("failed worker ownership not released: %v", err)
	}
	for _, channel := range []string{"receive", "request"} {
		t.Run(channel, func(t *testing.T) {
			input := ""
			command := append(append([]string{channel}, args...), "--session-id", "root")
			if channel == "receive" {
				payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "UserPromptSubmit", TranscriptPath: path})
				if err != nil {
					t.Fatal(err)
				}
				input = string(payload)
			}
			out, err := captureCLI(t, f.project, input, command...)
			if err != nil {
				t.Fatalf("later working agent channel: %v %s", err, out)
			}
			var response struct {
				Message  string               `json:"systemMessage"`
				Feedback string               `json:"capture_feedback"`
				Context  *capture.HookContext `json:"hookSpecificOutput"`
			}
			if err = json.Unmarshal(out, &response); err != nil {
				t.Fatal(err)
			}
			feedback := response.Message + response.Feedback
			if !strings.Contains(feedback, "publication failed") || !strings.Contains(feedback, "capture diagnostics") {
				t.Fatalf("durable publication failure silently lost at later %s: %s", channel, out)
			}
			if channel == "receive" && (response.Context == nil || response.Context.Event != "UserPromptSubmit" || !strings.Contains(response.Context.Context, "publication failed")) {
				t.Fatalf("confirmed failure did not reach supported agent additionalContext: %s", out)
			}
			for _, private := range []string{f.project, f.store, f.state, registration.Request} {
				if strings.Contains(feedback, private) {
					t.Fatalf("failure feedback leaked private detail %q: %s", private, feedback)
				}
			}
		})
	}
	query := func(session, store string) []byte {
		t.Helper()
		out, err := captureCLI(t, f.project, "", "request", "--repo", f.project, "--state-dir", f.state, "--source", "codex", "--store", store, "--session-id", session)
		if err != nil {
			t.Fatalf("exact request control: %v %s", err, out)
		}
		return out
	}
	assertQuiet := func(out []byte) {
		t.Helper()
		var response struct {
			Feedback string `json:"capture_feedback"`
			Message  string `json:"systemMessage"`
		}
		if err := json.Unmarshal(out, &response); err != nil {
			t.Fatal(err)
		}
		if response.Feedback != "" || response.Message != "" {
			t.Fatalf("nonfailure context received unrelated/obsolete warning: %s", out)
		}
	}
	prepareCaptureOwner(t, f, "other", "TEST-200", 9)
	assertQuiet(query("other", f.store))
	if err = writeCaptureFixtureLocal(t, f, []byte(strings.Replace(local, "enabled: true", "enabled: false", 1))); err != nil {
		t.Fatal(err)
	}
	if out := query("root", f.store); strings.TrimSpace(string(out)) != "" {
		t.Fatalf("disabled request did not stop before capture IO: %s", out)
	}
	disabledInput, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "UserPromptSubmit", TranscriptPath: path})
	if err != nil {
		t.Fatal(err)
	}
	disabledOut, err := captureCLI(t, f.project, string(disabledInput), append([]string{"receive"}, args...)...)
	if err != nil {
		t.Fatalf("disabled callback: %v %s", err, disabledOut)
	}
	assertQuiet(disabledOut)
	if err = writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(summary); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(summary, beforeSummary, 0600); err != nil {
		t.Fatal(err)
	}
	recovered := f.generate(t, "--append", summary)
	assertCaptureRootInput(t, recovered, 31)
	assertQuiet(query("root", f.store))
	recoveredOut, err := captureCLI(t, f.project, string(disabledInput), append([]string{"receive"}, args...)...)
	if err != nil {
		t.Fatalf("recovered callback: %v %s", err, recoveredOut)
	}
	assertQuiet(recoveredOut)

	// A confirmed source failure remains visible even when publishing retained
	// available facts succeeds, and a budget-pending retry cannot clear it.
	// This phase must complete real generation, unlike the earlier intentionally
	// failing publication phase with its short lifetime/pass policy.
	writeCaptureSuccessfulPublicationPolicy(t, f.project, f.state, false)
	sourceBody := readCaptureFile(t, path)
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	registry, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = registry.Change(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.Activity = "idle"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	// Queue a pass after removing the source. Earlier callbacks may have
	// detached a worker that has not yet tried the lease held by this fixture.
	// Either that child or the explicit worker can drain this generation.
	if err = registry.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	partialRequest, err := registry.PublicationRequest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err = held.Close(); err != nil {
		t.Fatal(err)
	}
	held = nil
	if out, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state); err != nil {
		t.Fatalf("available retained facts publication failed: %v %s", err, out)
	}
	// A successful competing invocation can return without owning the lease.
	// Observe the actual generation acknowledgement, then hold the released
	// lease so subsequent feedback controls cannot race another source pass.
	completion := time.NewTimer(5 * time.Second)
	defer completion.Stop()
	poll := time.NewTicker(10 * time.Millisecond)
	defer poll.Stop()
	for held == nil {
		request, err := registry.PublicationRequest(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if request.Acknowledged >= partialRequest.Generation && request.Generation == request.Acknowledged {
			held, err = capture.AcquireWorker(f.coordination(t))
			if err != nil {
				t.Fatal(err)
			}
			if held != nil {
				break
			}
		}
		select {
		case <-completion.C:
			progress, progressErr := capture.WorkerStatus(f.coordination(t))
			t.Fatalf("partial-source generation %d did not finish: request=%+v worker=%+v err=%v log=%s", partialRequest.Generation, request, progress, progressErr, readCaptureFile(t, filepath.Join(f.coordination(t), "worker.log")))
		case <-poll.C:
		}
	}
	sourceResponse := query("root", f.store)
	var sourceFeedback struct {
		Feedback string `json:"capture_feedback"`
	}
	if err = json.Unmarshal(sourceResponse, &sourceFeedback); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sourceFeedback.Feedback, "source reconciliation failed") || strings.Contains(sourceFeedback.Feedback, "publication failed") {
		t.Fatalf("confirmed source failure conflated with successful partial publication: %s", sourceResponse)
	}
	if _, err = registry.Reconcile(context.Background(), "TEST-100", "fixture", 1, 1); err != nil {
		t.Fatal(err)
	}
	if err = registry.Close(); err != nil {
		t.Fatal(err)
	}
	pendingResponse := query("root", f.store)
	if err = json.Unmarshal(pendingResponse, &sourceFeedback); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sourceFeedback.Feedback, "source reconciliation failed") {
		t.Fatalf("pending retry falsely acknowledged confirmed source failure: %s", pendingResponse)
	}
	if err = os.WriteFile(path, sourceBody, 0600); err != nil {
		t.Fatal(err)
	}
	f.generate(t, "--append", summary)
	assertQuiet(query("root", f.store))
	// Exact store lookup must not borrow another registration's ticket.
	reader, err := capture.OpenReadOnlyProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	foreignFeedback, err := reader.SessionFailureFeedback(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"})
	closeErr := reader.Close()
	if err != nil || closeErr != nil || foreignFeedback != "" {
		t.Fatalf("foreign exact store inherited failure: %q %v %v", foreignFeedback, err, closeErr)
	}
	// Drain callback handoffs before fixture cleanup, using request ACK and
	// released ownership as the completion oracle rather than a fixed delay.
	if err = held.Close(); err != nil {
		t.Fatal(err)
	}
	held = nil
	if out, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state); err != nil {
		t.Fatalf("final feedback handoff drain: %v %s", err, out)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		reader, err := capture.OpenReadOnlyProject(f.state, f.project)
		if err != nil {
			t.Fatal(err)
		}
		request, err := reader.PublicationRequest(context.Background())
		closeErr := reader.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("drain request: %v %v", err, closeErr)
		}
		if request.Generation == request.Acknowledged {
			lock, err := capture.AcquireWorker(f.coordination(t))
			if err != nil {
				t.Fatal(err)
			}
			if lock != nil {
				if err = lock.Close(); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("native feedback callback workers did not finish: %s", readCaptureFile(t, filepath.Join(f.coordination(t), "worker.log")))
}
