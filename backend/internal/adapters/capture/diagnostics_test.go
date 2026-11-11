package capture_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func captureStateFiles(t *testing.T, directory string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "registry.db-shm") {
			return nil
		}
		if strings.HasSuffix(path, "registry.db-wal") && len(body) == 0 {
			return nil
		}
		result[path] = string(body)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCaptureReadOnlyDiagnosticsPreserveSettledFilesAndReadCommittedWAL(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(map[bool]string{false: "settled", true: "active-wal"}[active], func(t *testing.T) {
			ctx := context.Background()
			project := t.TempDir()
			captureProjectConfig(t, project, "DiagnosticsFixture")
			state := t.TempDir()
			writer, err := capture.OpenProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
			registered, err := writer.Register(ctx, id, "", "SessionStart")
			if err != nil {
				t.Fatal(err)
			}
			if err = writer.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "committed"}); err != nil {
				t.Fatal(err)
			}
			if !active {
				if err = writer.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				defer func() {
					if err := writer.Close(); err != nil {
						t.Error(err)
					}
				}()
			}
			before := captureStateFiles(t, state)
			reader, err := capture.OpenReadOnlyProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			rows, err := reader.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].Request != registered.Request {
				t.Fatalf("committed handoff omitted: %+v", rows)
			}
			diagnostics, err := reader.Diagnostics(ctx)
			if err != nil || len(diagnostics) != 1 {
				t.Fatalf("diagnostics=%+v %v", diagnostics, err)
			}
			if _, err = reader.Register(ctx, id, "", "Stop"); err == nil {
				t.Fatal("read-only connection accepted a write")
			}
			if _, err = capture.WorkerStatus(state); err != nil {
				t.Fatal(err)
			}
			if err = reader.Close(); err != nil {
				t.Fatal(err)
			}
			after := captureStateFiles(t, state)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("inspection changed registry facts/schema/committed WAL or created explicit worker files")
			}
		})
	}
	absent := filepath.Join(t.TempDir(), "absent")
	absentProject := t.TempDir()
	captureProjectConfig(t, absentProject, "AbsentFixture")
	if _, err := capture.OpenReadOnlyProject(absent, absentProject); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing registry: %v", err)
	}
	if _, err := os.Stat(absent); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("inspection created namespace: %v", err)
	}
}

func TestCaptureDiagnosticRetentionKeepsLastSuccessAndBoundedFailures(t *testing.T) {
	s := openStore(t, t.TempDir())
	ctx := context.Background()
	if err := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "committed", Destination: "accepted.json"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 40; i++ {
		if err := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "failed"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "staging", Destination: "unfinished.json"}); err != nil {
		t.Fatal(err)
	}
	diagnostics, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	failures, successes, staging := 0, 0, 0
	for _, d := range diagnostics {
		switch d.Outcome {
		case "failed":
			failures++
		case "committed":
			successes++
			if d.Destination != "accepted.json" {
				t.Fatal("last success overwritten")
			}
		case "staging":
			staging++
		}
	}
	if failures != 32 || successes != 1 || staging != 1 {
		t.Fatalf("retention failure: failed=%d success=%d staging=%d", failures, successes, staging)
	}
}

func TestCallbackProjectProofRetainsUnflushedAndForeignHandoffs(t *testing.T) {
	for _, relation := range []string{"unflushed", "foreign", "own"} {
		t.Run(relation, func(t *testing.T) {
			ctx := context.Background()
			project := t.TempDir()
			captureProjectConfig(t, project, "DiagnosticsFixture")
			state := t.TempDir()
			home := t.TempDir()
			store, err := capture.OpenProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := store.Close(); err != nil {
					t.Error(err)
				}
			}()
			path := filepath.Join(home, "sessions", "root.jsonl")
			if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			cwd := project
			if relation == "foreign" {
				cwd = t.TempDir()
			}
			writeMetadata := func() {
				t.Helper()
				body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"root","cwd":%q,"timestamp":"2026-01-01T00:00:00Z"}}`+"\n", cwd)
				// Invalid trailing data proves callback verification only reads metadata,
				// rather than performing an unbounded transcript extraction before binding.
				if err := os.WriteFile(path, []byte(body+"not a transcript body\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if relation != "unflushed" {
				writeMetadata()
			}
			response, err := capture.ReceiveCodex(ctx, store, capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path}, home, "/native", state)
			if err != nil {
				t.Fatal(err)
			}
			if !response.Registered {
				t.Fatal("durable handoff not acknowledged")
			}
			rows, err := store.List(ctx)
			if err != nil || len(rows) != 1 {
				t.Fatalf("lost minimal handoff: %+v %v", rows, err)
			}
			row := rows[0]
			if row.Request == "" || row.TicketID != "" || row.FirstObservedAt == "" {
				t.Fatalf("handoff guessed or omitted identity: %+v", row)
			}
			if relation == "own" {
				if row.ProjectProof != domain.CaptureProjectVerified || row.ProjectDirectory != project {
					t.Fatalf("own metadata unverified: %+v", row)
				}
			} else {
				if row.ProjectProof != domain.CaptureProjectUnverified || row.ProjectDirectory != "" {
					t.Fatalf("foreign/unflushed source acquired project proof: %+v", row)
				}
				if issues := domain.CaptureAdvisoryIssues(domain.CaptureAdvisoryEvidence{Registrations: rows}); len(issues) != 0 {
					t.Fatalf("unverified source became project warning: %v", issues)
				}
				if relation == "foreign" {
					if _, err = store.Bind(ctx, row.Identity, "TEST-100", row.Request); err == nil {
						t.Fatal("foreign source bound to project")
					}
				}
			}
			if relation == "unflushed" {
				writeMetadata()
				if err = store.RefreshProjectProofs(ctx, 1, capture.MaxPayloadBytes); err != nil {
					t.Fatal(err)
				}
				after, err := store.List(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if len(after) != 1 || after[0].ProjectProof != domain.CaptureProjectVerified || after[0].Request != row.Request || after[0].TicketID != "" {
					t.Fatalf("exact known handoff not recovered: %+v", after)
				}
				if issues := domain.CaptureAdvisoryIssues(domain.CaptureAdvisoryEvidence{Registrations: after}); len(issues) != 1 || !strings.Contains(issues[0], "unbound") {
					t.Fatalf("proven missing binding not actionable: %v", issues)
				}
			}
		})
	}
}

func TestInitialCaptureHandoffInterruptionBeforeAndAfterCommit(t *testing.T) {
	for _, phase := range []string{"before", "after"} {
		t.Run(phase, func(t *testing.T) {
			project, state, home := t.TempDir(), t.TempDir(), t.TempDir()
			captureProjectConfig(t, project, "InterruptedFixture")
			sourcePath := filepath.Join(home, "sessions", "root.jsonl")
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^TestInitialCaptureHandoffProcess$")
			command.Env = append(os.Environ(), "AL_CAPTURE_FIRST_HANDOFF="+phase, "AL_CAPTURE_FIRST_PROJECT="+project, "AL_CAPTURE_FIRST_STATE="+state, "AL_CAPTURE_FIRST_HOME="+home, "AL_CAPTURE_FIRST_PATH="+sourcePath)
			input, err := command.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			output, err := command.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			var stderr bytes.Buffer
			command.Stderr = &stderr
			if err = command.Start(); err != nil {
				t.Fatal(err)
			}
			ready := bufio.NewScanner(output)
			if !ready.Scan() || ready.Text() != "handoff-"+phase {
				if err := command.Process.Kill(); err != nil {
					t.Error(err)
				}
				t.Fatalf("handoff boundary not reached: %s %v", stderr.String(), ready.Err())
			}
			if err = command.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			if err = command.Wait(); err == nil {
				t.Fatal("interrupted process exited successfully")
			}
			if err = input.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
				t.Fatal(err)
			}
			reopened, err := capture.OpenProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := reopened.Close(); err != nil {
					t.Error(err)
				}
			}()
			rows, err := reopened.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if phase == "before" {
				if len(rows) != 0 {
					t.Fatalf("uncommitted initial identity survived: %+v", rows)
				}
				return
			}
			if len(rows) != 1 || rows[0].Request == "" || rows[0].TicketID != "" || rows[0].TranscriptPath != sourcePath || rows[0].ProjectDirectory != "" {
				t.Fatalf("accepted minimal handoff lost or fabricated proof/binding: %+v", rows)
			}
			if err = os.MkdirAll(filepath.Dir(sourcePath), 0700); err != nil {
				t.Fatal(err)
			}
			body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"root","cwd":%q,"timestamp":"2026-01-01T00:00:00Z"}}`+"\n", project)
			if err = os.WriteFile(sourcePath, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if err = reopened.RefreshProjectProofs(context.Background(), 1, capture.MaxPayloadBytes); err != nil {
				t.Fatal(err)
			}
			recovered, err := reopened.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(recovered) != 1 || recovered[0].Request != rows[0].Request || recovered[0].ProjectProof != domain.CaptureProjectVerified || recovered[0].TicketID != "" {
				t.Fatalf("exact surviving handoff did not recover: %+v", recovered)
			}
		})
	}
}

// TestInitialCaptureHandoffProcess exposes the actual registry commit boundary
// over a pipe; the parent kills it before metadata proof or worker launch.
func TestInitialCaptureHandoffProcess(t *testing.T) {
	phase := os.Getenv("AL_CAPTURE_FIRST_HANDOFF")
	if phase == "" {
		return
	}
	store, err := capture.OpenProject(os.Getenv("AL_CAPTURE_FIRST_STATE"), os.Getenv("AL_CAPTURE_FIRST_PROJECT"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	}()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: os.Getenv("AL_CAPTURE_FIRST_HOME"), SessionID: "root"}
	block := func() error {
		if _, err := fmt.Fprintln(os.Stdout, "handoff-"+phase); err != nil {
			return err
		}
		_, err := io.ReadFull(os.Stdin, make([]byte, 1))
		return err
	}
	if phase == "before" {
		_, err = store.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) { return row, block() })
	} else {
		_, err = store.Register(context.Background(), id, os.Getenv("AL_CAPTURE_FIRST_PATH"), "SessionStart")
		if err == nil {
			err = block()
		}
	}
	t.Fatalf("interruption boundary unexpectedly returned: %v", err)
}

func TestBindingRequestCannotReuseOldCallbackForNewInstallation(t *testing.T) {
	state := t.TempDir()
	store := openStore(t, state)
	home := t.TempDir()
	ctx := context.Background()
	manifestPath := filepath.Join(state, "install-codex.json")
	writeInstall := func(version string) {
		t.Helper()
		body, err := json.Marshal(struct {
			Active      bool   `json:"active"`
			SourceStore string `json:"source_store"`
			Version     string `json:"version"`
		}{true, home, version})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(manifestPath, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeInstall("first")
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: "root"}
	callback, err := store.Register(ctx, id, "", "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if callback.InstallationID == "" || callback.CallbackObservedAt == "" {
		t.Fatalf("callback evidence missing: %+v", callback)
	}
	writeInstall("second")
	request, err := store.Register(ctx, id, "", "agent_context_request")
	if err != nil {
		t.Fatal(err)
	}
	current, _, err := capture.InstallationIdentity(state, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if request.InstallationID != callback.InstallationID || request.InstallationID == current || request.CallbackObservedAt != callback.CallbackObservedAt {
		t.Fatalf("request fabricated callback for current install: callback=%+v request=%+v", callback, request)
	}
	observed, err := store.Register(ctx, id, "", "UserPromptSubmit")
	if err != nil {
		t.Fatal(err)
	}
	if observed.InstallationID != current {
		t.Fatalf("new actual callback did not identify current install: %+v", observed)
	}
}

func TestDiagnosticCurrentFailuresSurvivePendingAndUnrelatedHistory(t *testing.T) {
	s := openStore(t, t.TempDir())
	ctx := context.Background()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
	if _, err := s.Change(ctx, id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		row.TicketID = "TEST-100"
		return row, nil
	}); err != nil {
		t.Fatal(err)
	}
	record := func(value domain.CaptureDiagnostic) {
		t.Helper()
		if err := s.RecordDiagnostic(ctx, value); err != nil {
			t.Fatal(err)
		}
	}
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "failed"})
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "preparing"})
	failedOwner := domain.CaptureRegistration{Identity: id, State: domain.CaptureFailed}
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "reconciliation", Outcome: "failed", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{failedOwner}}})
	other := failedOwner
	other.Identity.SessionID = "other"
	pending := failedOwner
	pending.State = domain.CapturePending
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "reconciliation", Outcome: "failed", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{pending, other}}})
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{pending}}})
	for i := 0; i < 40; i++ {
		record(domain.CaptureDiagnostic{TicketID: "TEST-200", Phase: "publication", Outcome: "failed"})
	}
	feedback, err := s.SessionFailureFeedback(ctx, id, fixtureByteBudget)
	if err != nil || !strings.Contains(feedback, "publication failed") || !strings.Contains(feedback, "source reconciliation failed") {
		t.Fatalf("unresolved terminal evidence erased by pending/unrelated churn: %q %v", feedback, err)
	}
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "committed"})
	successful := pending
	successful.State = domain.CaptureCurrent
	successful.RefreshedAt = "2026-01-01T00:00:00Z"
	record(domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptPartial, Owners: []domain.CaptureRegistration{successful}}})
	feedback, err = s.SessionFailureFeedback(ctx, id, fixtureByteBudget)
	if err != nil || feedback != "" {
		t.Fatalf("relevant terminal success did not supersede failure: %q %v", feedback, err)
	}
	diagnostics, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	retained := false
	for _, value := range diagnostics {
		if value.TicketID == "TEST-100" && value.Phase == "publication" && value.Outcome == "failed" {
			retained = true
		}
	}
	if !retained {
		t.Fatal("last failure fact removed after terminal recovery")
	}
}

func TestPublishedOwnerAssociationUsesSelectedFactsAndSourceInstallation(t *testing.T) {
	for _, scenario := range []string{"matching", "different equal checkpoint", "absent snapshot"} {
		t.Run(scenario, func(t *testing.T) {
			s := openStore(t, t.TempDir())
			ctx := context.Background()
			id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "root"}
			stamp := "2026-01-01T00:00:02Z"
			selected := domain.SessionFact{ID: "root", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{CompletedAt: &stamp, Tokens: domain.InvocationTokens{Input: 10}}}}
			snapshot := selected
			if scenario == "different equal checkpoint" {
				snapshot.Invocations = append([]domain.InvocationFact{}, selected.Invocations...)
				snapshot.Invocations[0].Tokens.Input = 300
			}
			if _, err := s.Change(ctx, id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				row.TicketID = "TEST-100"
				row.InstallationID = "latest-callback-installation"
				row.SourceInstallationID = "accepted-source-installation"
				row.SourceFingerprint = "accepted-source-fingerprint"
				if scenario != "absent snapshot" {
					row.Snapshot = []domain.SessionFact{snapshot}
				}
				return row, nil
			}); err != nil {
				t.Fatal(err)
			}
			if err := s.RecordPublication(ctx, "TEST-100", "committed", "metrics.json", "", []domain.SessionFact{selected}, ""); err != nil {
				t.Fatal(err)
			}
			diagnostics, err := s.Diagnostics(ctx)
			if err != nil {
				t.Fatal(err)
			}
			published := diagnostics[0].Owners[0]
			if published.Checkpoint != stamp {
				t.Fatalf("actual selected checkpoint lost: %+v", published)
			}
			if scenario == "matching" {
				if published.Association != "matching_registry_snapshot" || published.InstallationID != "accepted-source-installation" || published.SourceFingerprint != "accepted-source-fingerprint" || published.Store != id.Store {
					t.Fatalf("matching source snapshot association lost/borrowed latest callback: %+v", published)
				}
			} else if published.Association != "unknown" || published.InstallationID != "" || published.SourceFingerprint != "" || published.Store != "" {
				t.Fatalf("rejected/unknown snapshot lent provenance: %+v", published)
			}
		})
	}
}

func TestDiagnosticInspectionBoundsRowsBytesAndPreservesRelevantCurrentState(t *testing.T) {
	s := openStore(t, t.TempDir())
	ctx := context.Background()
	record := func(value domain.CaptureDiagnostic) {
		t.Helper()
		if err := s.RecordDiagnostic(ctx, value); err != nil {
			t.Fatal(err)
		}
	}
	record(domain.CaptureDiagnostic{TicketID: "TEST-current", Phase: "publication", Outcome: "committed"})
	record(domain.CaptureDiagnostic{TicketID: "TEST-current", Phase: "publication", Outcome: "failed"})
	record(domain.CaptureDiagnostic{TicketID: "TEST-current", Phase: "publication", Outcome: "preparing"})
	for index := 0; index < 80; index++ {
		record(domain.CaptureDiagnostic{TicketID: fmt.Sprintf("HISTORY-%d", index), Phase: "publication", Outcome: "committed"})
	}
	current, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-current"}, Limit: 8, MaxBytes: 1024 * 1024})
	if err != nil {
		t.Fatal(err)
	}
	if current.More || len(current.Records) != 3 {
		t.Fatalf("relevant current state scaled with historical ticket count: %+v", current)
	}
	for _, value := range current.Records {
		if value.TicketID != "TEST-current" {
			t.Fatalf("unrelated history included in current advisory work: %+v", value)
		}
	}
	first, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Limit: 3, MaxBytes: 1024 * 1024})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Limit: 3, Before: first.NextBefore, MaxBytes: 1024 * 1024})
	if err != nil {
		t.Fatal(err)
	}
	if !first.More || !second.More || len(first.Records) != 3 || len(second.Records) != 3 || second.Records[0].Sequence >= first.NextBefore {
		t.Fatalf("bounded history cursor lost continuation: first=%+v second=%+v", first, second)
	}
	small, err := s.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-current"}, Limit: 8, MaxBytes: 1})
	if err != nil || !small.More || len(small.Records) != 0 || small.Bytes != 0 {
		t.Fatalf("byte budget silently exceeded/claimed complete: %+v %v", small, err)
	}
	// Unrelated historical latest/last-success rows remain intact after bounded reads.
	all, err := s.Diagnostics(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 83 {
		t.Fatalf("bounded reads evicted required ticket facts: %d", len(all))
	}
}
