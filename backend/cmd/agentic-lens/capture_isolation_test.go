package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureManualSelectorsRequireExactSourceProject(t *testing.T) {
	for _, relation := range []string{"different directory", "nested directory", "true alias"} {
		for _, selector := range []string{"session", "nonce"} {
			t.Run(relation+"/"+selector, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				path := f.write(t, 23)
				selectedProject := t.TempDir()
				if relation == "nested directory" {
					selectedProject = filepath.Join(f.project, "nested")
					if err := os.Mkdir(selectedProject, 0700); err != nil {
						t.Fatal(err)
					}
				}
				if relation == "true alias" {
					selectedProject = filepath.Join(t.TempDir(), "alias")
					if err := os.Symlink(f.project, selectedProject); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.WriteFile(filepath.Join(selectedProject, ".agenticlens-config.yaml"), readCaptureFile(t, filepath.Join(f.project, ".agenticlens-config.yaml")), 0600); err != nil {
						t.Fatal(err)
					}
					if err := os.MkdirAll(filepath.Join(selectedProject, "records"), 0700); err != nil {
						t.Fatal(err)
					}
					writeManualStateOnlyConfig(t, selectedProject)
				}
				file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, err = file.WriteString("{\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":[{\"type\":\"input_text\",\"text\":\"isolation-nonce\"}]}}\n")
				if err != nil {
					t.Fatal(err)
				}
				if err = file.Close(); err != nil {
					t.Fatal(err)
				}
				output := filepath.Join(selectedProject, "agentic_metrics.json")
				args := []string{"generate", "--source", "codex", "--codex-home", f.store, "--capture-dir", f.state, "--ticket-id", "TEST-100", "--ticket-title", "Exact project", "--ticket-type", "feat", "--json-output", output}
				if selector == "session" {
					args = append(args, "--session-id", "root")
				} else {
					args = append(args, "--nonce", "isolation-nonce")
				}
				// Each project shares a state base, not another project's already-scoped namespace.
				args[6] = filepath.Dir(filepath.Dir(f.state))
				cmd := exec.Command(cliBinaryPath, args...)
				cmd.Dir = selectedProject
				out, err := cmd.CombinedOutput()
				if relation != "true alias" {
					if err == nil {
						t.Fatalf("foreign source imported: %s", out)
					}
					if _, e := os.Stat(output); !os.IsNotExist(e) {
						t.Fatalf("rejected owner wrote ledger: %v", e)
					}
					return
				}
				if err != nil {
					t.Fatalf("physical project alias rejected: %v %s", err, out)
				}
				var ledger domain.AgenticFactLedger
				if err = json.Unmarshal(readCaptureFile(t, output), &ledger); err != nil {
					t.Fatal(err)
				}
				assertCaptureRootInput(t, ledger, 23)
			})
		}
	}
}

func TestCaptureCopiedCanonicalBaselineCannotCrossProjects(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 31))
	accepted := f.generate(t)
	other := newCaptureGenerationFixture(t)
	if err := os.WriteFile(other.output, readCaptureFile(t, f.output), 0600); err != nil {
		t.Fatal(err)
	}
	before := readCaptureFile(t, other.output)
	cmd := exec.Command(cliBinaryPath, other.args(t, "--session-id", "fixture-empty-root")...)
	cmd.Dir = other.project
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("copied baseline accepted as local: %s", out)
	}
	if !bytes.Equal(before, readCaptureFile(t, other.output)) {
		t.Fatal("rejected baseline changed")
	}
	assertCaptureRootInput(t, accepted, 31)
}

// The journal is the actual production recovery API; partial writes represent
// process death after each externally visible member, with no in-memory recovery.
func TestCaptureCLIRestartsInterruptedArtifactPublication(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		pathsCount := 2
		if legacy {
			pathsCount = 3
		}
		for writes := 0; writes <= pathsCount; writes++ {
			t.Run(fmt.Sprintf("legacy=%t/writes=%d", legacy, writes), func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				f.bind(t, f.write(t, 19))
				before := f.generate(t)
				summary := filepath.Join(f.project, "SUMMARY.md")
				report := filepath.Join(f.project, "capture_report.json")
				retired := []byte("historical report retained for explicit user removal")
				paths := []string{f.output, summary}
				if legacy {
					if err := os.WriteFile(report, retired, 0600); err != nil {
						t.Fatal(err)
					}
					paths = []string{f.output, report, summary}
				}
				if err := os.WriteFile(summary, []byte("# Preserved user introduction\n"), 0600); err != nil {
					t.Fatal(err)
				}
				lock, err := capture.AcquirePublication(context.Background(), f.output)
				if err != nil {
					t.Fatal(err)
				}
				if err = adapters.BeginPublication(f.output, paths); err != nil {
					lock.Close()
					t.Fatal(err)
				}
				for _, path := range paths[:writes] {
					if err = os.WriteFile(path, []byte("interrupted replacement"), 0600); err != nil {
						lock.Close()
						t.Fatal(err)
					}
				}
				if err = lock.Close(); err != nil {
					t.Fatal(err)
				}
				if err = adapters.CheckPublication(f.output); err == nil {
					t.Fatal("interrupted set claimed complete")
				}
				if legacy {
					// Retired report membership requires explicit recovery with the
					// complete accepted set; current CLI must not infer an adapter.
					unchanged := map[string][]byte{}
					for _, path := range append(paths, f.output+".publication-pending") {
						unchanged[path] = readCaptureFile(t, path)
					}
					out, err := runCaptureNative(t, f.project, "", f.args(t, "--append", summary)...)
					if err == nil || !bytes.Contains(out, []byte("complete authorized publication set")) {
						t.Fatalf("retired membership silently adapted: %v %s", err, out)
					}
					for path, body := range unchanged {
						if !bytes.Equal(body, readCaptureFile(t, path)) {
							t.Fatalf("rejected recovery changed %s", path)
						}
					}
					return
				}
				after := f.generate(t, "--append", summary)
				if !reflect.DeepEqual(before.Sessions, after.Sessions) {
					t.Fatalf("restart lost accepted source: %+v", after.Sessions)
				}
				if !bytes.Contains(readCaptureFile(t, summary), []byte("# Preserved user introduction")) {
					t.Fatal("restart lost user summary")
				}
				if err = adapters.CheckPublication(f.output); err != nil {
					t.Fatalf("restart did not commit: %v", err)
				}
				if bytes.Contains(readCaptureFile(t, summary), []byte("](capture_report.json)")) {
					t.Fatal("recovered summary linked retired report")
				}
				if _, err := os.Stat(report); !os.IsNotExist(err) {
					t.Fatalf("modern publication created report: %v", err)
				}
				assertCapturePublishedOwners(t, captureDiagnostics(t, f.project, f.state), f.output, summary, after.Sessions)
			})
		}
	}
}

func TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects(t *testing.T) {
	first := newCaptureGenerationFixture(t)
	second := newCaptureGenerationFixture(t)
	second.store = first.store
	base := t.TempDir()
	fixtures := []captureGenerationFixture{first, second}
	for index, f := range fixtures {
		seedCaptureTicketHome(t, f, "TEST-100")
		writeCaptureSuccessfulPublicationPolicy(t, f.project, base)
		scoped, err := capture.ProjectState(base, f.project)
		if err != nil {
			t.Fatal(err)
		}
		store, err := capture.OpenProject(scoped, f.project)
		if err != nil {
			t.Fatal(err)
		}
		completion := captureWorkerCompletion(t, store.CoordinationDirectory())
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
		id := fmt.Sprintf("project-%d", index)
		path := filepath.Join(f.store, "sessions", id+".jsonl")
		body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, id, f.project, id, 10+index)
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		callback := func(event string) {
			t.Helper()
			payload, _ := json.Marshal(capture.HookInput{SessionID: id, Event: event, TranscriptPath: path})
			if out, err := captureCLI(t, f.project, string(payload), "receive", "--source", "codex", "--store", f.store, "--state-dir", base); err != nil {
				t.Fatalf("callback: %v %s", err, out)
			}
		}
		// Register teardown before the first callback can launch a worker.
		t.Cleanup(func() {
			defer func() {
				// Stop future passes/handoffs even if a callback or assertion failed.
				if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: false\n")); err != nil {
					t.Error(err)
				}
				captureCleanupWorkers(t, completion)
			}()
			callback("SessionEnd")
			scoped, err := capture.ProjectState(base, f.project)
			if err != nil {
				t.Error(err)
				return
			}
			store, err := capture.OpenProject(scoped, f.project)
			if err != nil {
				t.Error(err)
				return
			}
			defer store.Close()
			deadline := time.Now().Add(5 * time.Second)
			for time.Now().Before(deadline) {
				req, e := store.PublicationRequest(context.Background())
				if e != nil {
					t.Error(e)
					return
				}
				if req.Generation == req.Acknowledged {
					lock, e := capture.AcquireWorker(scoped)
					if e != nil {
						t.Error(e)
						return
					}
					if lock != nil {
						if err := lock.Close(); err != nil {
							t.Error(err)
						}
						return
					}
				}
				time.Sleep(20 * time.Millisecond)
			}
			t.Error("shared-store worker did not stop before cleanup")
		})
		callback("SessionStart")
		out, err := captureCLI(t, f.project, "", "request", "--source", "codex", "--store", f.store, "--state-dir", base, "--session-id", id)
		if err != nil {
			t.Fatalf("request: %v %s", err, out)
		}
		var row domain.CaptureRegistration
		if err = json.Unmarshal(out, &row); err != nil {
			t.Fatal(err)
		}
		if out, err = captureCLI(t, f.project, "", "bind", "--source", "codex", "--store", f.store, "--state-dir", base, "--session-id", id, "--request", row.Request, "--ticket-id", "TEST-100", "--ticket-title", "Shared ticket", "--ticket-type", "feat"); err != nil {
			t.Fatalf("bind: %v %s", err, out)
		}
	}
	for index, f := range fixtures {
		output := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
		deadline := time.Now().Add(8 * time.Second)
		ready := false
		for time.Now().Before(deadline) {
			body, err := os.ReadFile(output)
			var ledger domain.AgenticFactLedger
			if err == nil && json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) > 0 {
				if len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != fmt.Sprintf("project-%d", index) || len(ledger.Sessions[0].Invocations) != 1 || ledger.Sessions[0].Invocations[0].Tokens.Input != int64(10+index) {
					t.Fatalf("cross-project automatic import: %+v", ledger.Sessions)
				}
				ready = true
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if !ready {
			t.Fatalf("automatic publication missing for project %d", index)
		}
	}
}

func TestCaptureWorkerCompletionIncludesDelayedNativeStartup(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: true\n")); err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	scoped := store.CoordinationDirectory()
	completion := captureWorkerCompletion(t, scoped)
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Existing production admission holds the native before registry/lease
	// acquisition, without altering its executable or adding runtime test hooks.
	admission, err := adaptersconfig.AcquireCaptureAdmission(context.Background(), f.project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := admission.Close(); err != nil {
			t.Error(err)
		}
		captureCleanupWorkers(t, completion)
	})
	if err := capture.LaunchWorker(cliBinaryPath, f.project, scoped); err != nil {
		t.Fatal(err)
	}
	request, err := store.PublicationRequest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	owner, err := capture.AcquireWorker(scoped)
	if err != nil || owner == nil {
		t.Fatalf("delayed native should not yet own worker lease: %v", err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if request.Generation != request.Acknowledged {
		t.Fatalf("unexpected pending publication: %+v", request)
	}
	// The former ACK/lease oracle passes here, but a known native still holds
	// the log and will create worker.lock after admission is released.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := completion(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("completion accepted a native blocked before startup: %v", err)
	}
	if err := os.Remove(filepath.Join(scoped, "worker.lock")); err != nil {
		t.Fatal(err)
	}
	if err := admission.Close(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := completion(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(scoped, "worker.lock")); err != nil {
		t.Fatalf("native did not reach lease allocation after delayed startup: %v", err)
	}
}

// A failed diagnostic deadline must not allow TempDir cleanup to overtake a
// native writer. The suite timeout remains the final stuck-process diagnosis.
func captureDrainWorkerCompletion(ctx context.Context, completion func(context.Context) error) error {
	err := completion(ctx)
	if err == nil {
		return nil
	}
	return errors.Join(err, completion(context.Background()))
}

func captureCleanupWorkers(t *testing.T, completion func(context.Context) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := captureDrainWorkerCompletion(ctx, func(waitCtx context.Context) error {
		err := completion(waitCtx)
		if err != nil {
			t.Logf("native writer completion failed: %v; waiting before fixture removal", err)
		}
		return err
	}); err != nil {
		t.Errorf("native writer cleanup exceeded its diagnostic deadline: %v", err)
		if finalErr := completion(context.Background()); finalErr != nil {
			t.Errorf("native completion observer failed: %v; retaining fixtures until suite timeout", finalErr)
			// An unavailable observer cannot authorize removal beneath a writer.
			select {}
		}
	}
}

func TestCaptureDescendantOwnershipComesFromEachNativeSession(t *testing.T) {
	for _, foreign := range []bool{false, true} {
		t.Run(fmt.Sprintf("foreign-%v", foreign), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.write(t, 23)
			directory := f.project
			if foreign {
				directory = t.TempDir()
			}
			child := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"child","cwd":%q,"source":{"subagent":{"thread_spawn":{"parent_thread_id":"root","depth":1}}}}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"child","turn_id":"child-turn","usage":{"input_tokens":70,"output_tokens":4}}}
`, directory)
			if err := os.WriteFile(filepath.Join(f.store, "sessions", "child.jsonl"), []byte(child), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, f.args(t, "--session-id", "root", "--codex-home", f.store)...)
			cmd.Dir = f.project
			out, err := cmd.CombinedOutput()
			if foreign && err != nil {
				if !bytes.Contains(out, []byte("project directory conflicts")) {
					t.Fatalf("failed before testing descendant ownership: %v %s", err, out)
				}
				if _, err = os.Stat(f.output); !os.IsNotExist(err) {
					t.Fatalf("rejected extraction published output: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("same-project source extraction: %v %s", err, out)
			}
			ledger := f.ledger(t)
			assertCaptureRootInput(t, ledger, 23)
			want := 2
			if foreign {
				want = 1
			}
			if len(ledger.Sessions) != want {
				t.Fatalf("wrong descendants: %+v", ledger.Sessions)
			}
			for _, owner := range ledger.Sessions {
				if owner.ID == "child" && (foreign || owner.ParentID == nil || *owner.ParentID != "root" || len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != 70) {
					t.Fatalf("child ownership or usage lost: %+v", owner)
				}
			}
		})
	}
}

func TestCaptureInstallerUsesLocalAndExplicitStatePolicy(t *testing.T) {
	for _, level := range []string{"local", "cli"} {
		t.Run(level, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			config := filepath.Join(f.project, ".agenticlens-config.yaml")
			lines := bytes.Split(readCaptureFile(t, config), []byte("\n"))
			for i, line := range lines {
				if bytes.HasPrefix(line, []byte("    state_dir:")) {
					lines[i] = []byte("    state_dir: shared-state")
				}
			}
			body := bytes.Join(lines, []byte("\n"))
			if err := os.WriteFile(config, body, 0600); err != nil {
				t.Fatal(err)
			}
			want := "shared-state"
			if level != "shared" {
				want = "local-state"
				if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: false\n    state_dir: local-state\n")); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"install", "--source", "opencode", "--repo", f.project, "--store", f.store, "--binary", cliBinaryPath}
			if level == "cli" {
				want = "explicit-state"
				args = append(args, "--state-dir", want)
			}
			if out, err := captureCLI(t, f.project, "", args...); err != nil {
				t.Fatalf("install: %v %s", err, out)
			}
			expected, err := capture.ProjectState(filepath.Join(f.project, want), f.project)
			if err != nil {
				t.Fatal(err)
			}
			plugin := readCaptureFile(t, filepath.Join(f.project, ".opencode", "plugins", "agenticlens-capture.ts"))
			quoted, _ := json.Marshal(expected)
			if !bytes.Contains(plugin, append([]byte("const stateDir = "), quoted...)) {
				t.Fatalf("installed path did not honor %s policy: expected %s\n%s", level, expected, plugin)
			}
			quoted, _ = json.Marshal(fixtureWrapperFromHook(t, f.project))
			if !bytes.Contains(plugin, append([]byte("const wrapper = "), quoted...)) {
				t.Fatal("installed callback did not retain the common project-pinned wrapper")
			}
		})
	}
}
