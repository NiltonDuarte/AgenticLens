package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// The gate subprocess forwards real Git unchanged after a deterministic test barrier.
func init() {
	if os.Getenv("AL_CAPTURE_TEST_GIT_HELPER") != "1" {
		return
	}
	if strings.Join(os.Args[1:], " ") == "rev-parse --show-toplevel" {
		connection, err := net.DialTimeout("tcp", os.Getenv("AL_CAPTURE_TEST_GATE_ADDRESS"), 5*time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(91)
		}
		if err := connection.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(92)
		}
		if _, err := connection.Write([]byte("R")); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(93)
		}
		if _, err := io.ReadFull(connection, make([]byte, 1)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(94)
		}
		if err := connection.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(95)
		}
	}
	command := exec.Command(os.Getenv("AL_CAPTURE_TEST_REAL_GIT"), os.Args[1:]...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(96)
	}
	os.Exit(0)
}

type captureGenerationFixture struct{ project, store, state, stateBase, output string }

func newCaptureGenerationFixture(t *testing.T) captureGenerationFixture {
	t.Helper()
	f := captureGenerationFixture{project: t.TempDir(), store: t.TempDir(), state: t.TempDir()}
	t.Setenv("CODEX_HOME", f.store)
	f.stateBase = f.state
	f.output = filepath.Join(f.project, "ledger.json")
	config := "project_name: CaptureFixture\ntelemetry:\n  capture:\n    state_dir: " + strconv.Quote(f.state) + "\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n")); err != nil {
		t.Fatal(err)
	}
	var err error
	f.state, err = capture.ProjectState(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(f.store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f captureGenerationFixture) coordination(t *testing.T) string {
	t.Helper()
	location, err := capture.ExistingLocation(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	return location.CoordinationDirectory
}

func (f captureGenerationFixture) write(t *testing.T, tokens int) string {
	return f.writeAt(t, tokens, "2026-01-01T00:00:02Z")
}

func (f captureGenerationFixture) writeAt(t *testing.T, tokens int, sourceTime string) string {
	t.Helper()
	path := filepath.Join(f.store, "sessions", "root.jsonl")
	contents := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"root","cwd":%q,"timestamp":"2026-01-01T00:00:00Z"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":"root","turn_id":"one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, f.project, sourceTime, tokens)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func (f captureGenerationFixture) bind(t *testing.T, path string) {
	t.Helper()
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
	row, err := store.Register(context.Background(), id, path, "SessionStart")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Bind(context.Background(), id, "TEST-100", row.Request); err != nil {
		t.Fatal(err)
	}
}

func (f captureGenerationFixture) args(t *testing.T, extra ...string) []string {
	explicitOwner := "root"
	for i, arg := range extra {
		if arg == "--session-id" && i+1 < len(extra) {
			explicitOwner = extra[i+1]
		}
	}
	for i, arg := range extra {
		if arg == "--capture-dir" && i+1 < len(extra) {
			dir, err := capture.ProjectState(extra[i+1], f.project)
			if err != nil {
				t.Fatal(err)
			}
			registry, err := capture.OpenProject(dir, f.project)
			if err != nil {
				t.Fatal(err)
			}
			r, err := registry.Register(context.Background(), domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: explicitOwner}, "", "fixture explicit owner")
			if err != nil {
				t.Fatal(err)
			}
			if err = capture.ValidateOwner(context.Background(), f.project, r); err != nil {
				t.Fatal(err)
			}
			if _, err = registry.Bind(context.Background(), r.Identity, "TEST-100", r.Request); err != nil {
				t.Fatal(err)
			}
			_, err = registry.Change(context.Background(), r.Identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				row.ProjectDirectory = f.project
				return row, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if err = registry.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
	return append([]string{"generate", "--source", "codex", "--capture-dir", f.state, "--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "feat", "--json-output", f.output}, extra...)
}

func (f captureGenerationFixture) generate(t *testing.T, extra ...string) domain.AgenticFactLedger {
	t.Helper()
	cmd := exec.Command(cliBinaryPath, f.args(t, extra...)...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate: %v %s", err, out)
	}
	return f.ledger(t)
}

func (f captureGenerationFixture) ledger(t *testing.T) domain.AgenticFactLedger {
	t.Helper()
	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal(readCaptureFile(t, f.output), &ledger); err != nil {
		t.Fatal(err)
	}
	return ledger
}

func assertCaptureRootInput(t *testing.T, ledger domain.AgenticFactLedger, want int64) {
	t.Helper()
	for _, row := range ledger.Sessions {
		if row.ID == "root" {
			if len(row.Invocations) != 1 || row.Invocations[0].Tokens.Input != want {
				t.Fatalf("root snapshot=%+v, want input %d", row, want)
			}
			return
		}
	}
	t.Fatal("missing root session")
}

func captureCLI(t *testing.T, project, input string, args ...string) ([]byte, error) {
	t.Helper()
	local := filepath.Join(project, ".agenticlens-config.local.yaml")
	if _, err := os.Stat(local); os.IsNotExist(err) {
		state := ".agenticlens-capture"
		for i := 0; i+1 < len(args); i++ {
			if args[i] == "--state-dir" {
				state = args[i+1]
			}
		}
		if err := writeCaptureFixtureLocal(t, captureGenerationFixture{project: project, state: state}, []byte("telemetry:\n  capture:\n")); err != nil {
			t.Fatal(err)
		}
	}
	args = prepareCaptureInstall(t, project, args)
	cmd := exec.Command(cliBinaryPath, append([]string{"capture"}, args...)...)
	cmd.Dir = project
	cmd.Stdin = strings.NewReader(input)
	return cmd.CombinedOutput()
}

func readCaptureFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func captureDiagnostics(t *testing.T, project, state string) []domain.CaptureDiagnostic {
	t.Helper()
	out, err := runCaptureNative(t, project, "", "capture", "diagnostics", "--repo", project, "--state-dir", state)
	if err != nil {
		t.Fatalf("capture diagnostics: %v %s", err, out)
	}
	var status struct {
		Diagnostics []domain.CaptureDiagnostic `json:"diagnostics"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		t.Fatalf("diagnostic JSON: %v %s", err, out)
	}
	return status.Diagnostics
}

func assertCapturePublishedOwners(t *testing.T, diagnostics []domain.CaptureDiagnostic, output, summary string, sessions []domain.SessionFact) {
	t.Helper()
	var committed *domain.CaptureDiagnostic
	for i := range diagnostics {
		d := &diagnostics[i]
		if d.Phase == "publication" && (committed == nil || d.Sequence > committed.Sequence) {
			committed = d
		}
	}
	if committed == nil || committed.Outcome != "committed" || committed.ObservedAt == "" || committed.Destination != output || committed.Summary != summary {
		t.Fatalf("publication not durably committed at requested destinations: %+v", committed)
	}
	if len(committed.Owners) != len(sessions) {
		t.Fatalf("published owner count=%d want %d", len(committed.Owners), len(sessions))
	}
	for _, session := range sessions {
		found := false
		for _, owner := range committed.Owners {
			if owner.Source != session.Source || owner.SessionID != session.ID {
				continue
			}
			found = true
			if checkpoint, known := domain.CaptureOwnerCheckpoint(session); known && owner.Checkpoint != checkpoint.UTC().Format(time.RFC3339Nano) {
				t.Fatalf("published checkpoint for %s=%s want %s", session.ID, owner.Checkpoint, checkpoint)
			}
		}
		if !found {
			t.Fatalf("committed diagnostic omitted owner %s/%s", session.Source, session.ID)
		}
	}
}

func TestCaptureCLIInstallerPreservesUserHooksDryRunAndRepeatedRemoval(t *testing.T) {
	project := filepath.Join(t.TempDir(), "adopting repo with spaces")
	if err := os.MkdirAll(filepath.Join(project, ".codex"), 0700); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(project, ".codex", "hooks.json")
	original := []byte(`{"description":"user-owned","hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"user-command"}]}]}}`)
	if err := os.WriteFile(configPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(project, "custom local state")
	store := filepath.Join(t.TempDir(), "source store")
	base := []string{"--repo", project, "--state-dir", state, "--source", "codex", "--store", store, "--binary", cliBinaryPath}
	invoke := func(command string, extra ...string) []byte {
		t.Helper()
		out, err := captureCLI(t, project, "", append(append([]string{command}, base...), extra...)...)
		if err != nil {
			t.Fatalf("%s failed: %v\n%s", command, err, out)
		}
		return out
	}
	invoke("install", "--dry-run")
	if !bytes.Equal(readCaptureFile(t, configPath), original) {
		t.Fatal("dry run changed user hooks")
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("dry run created capture state: %v", err)
	}
	invoke("install")
	installed := readCaptureFile(t, configPath)
	if !bytes.Contains(installed, []byte("user-command")) || !bytes.Contains(installed, []byte("AgenticLens capture")) {
		t.Fatalf("installed hooks lost user/owned entries: %s", installed)
	}
	invoke("install")
	if !bytes.Equal(readCaptureFile(t, configPath), installed) {
		t.Fatal("reinstall duplicated/changed hooks")
	}
	edited := bytes.Replace(installed, []byte("AgenticLens capture"), []byte("locally edited capture"), 1)
	if err := os.WriteFile(configPath, edited, 0600); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"install", "uninstall"} {
		if out, err := captureCLI(t, project, "", append([]string{command}, base...)...); err == nil {
			t.Fatalf("%s overwrote local edit: %s", command, out)
		}
		if !bytes.Equal(readCaptureFile(t, configPath), edited) {
			t.Fatal("conflict handler changed user edits")
		}
	}
	if err := os.WriteFile(configPath, installed, 0600); err != nil {
		t.Fatal(err)
	}
	invoke("disable")
	invoke("disable")
	invoke("uninstall")
	remaining := readCaptureFile(t, configPath)
	if !bytes.Contains(remaining, []byte("user-command")) || bytes.Contains(remaining, []byte("AgenticLens capture")) {
		t.Fatalf("removal damaged ownership: %s", remaining)
	}
	var before, after struct {
		Description string                       `json:"description"`
		Hooks       map[string][]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(original, &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(remaining, &after); err != nil {
		t.Fatal(err)
	}
	if before.Description != after.Description || len(after.Hooks) != 1 || len(after.Hooks["SessionStart"]) != 1 {
		t.Fatalf("unrelated hook config changed: %s", remaining)
	}
	scoped, err := capture.ProjectState(state, project)
	if err != nil {
		t.Fatal(err)
	}
	location, err := capture.ExistingLocation(scoped, project)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := capture.ReadManifest(location.CoordinationDirectory, "codex")
	if err != nil || manifest.Active {
		t.Fatalf("removed installation reported active: %+v %v", manifest, err)
	}
}

func TestCaptureCLIOpenCodePluginOwnershipAndMissingHandler(t *testing.T) {
	project := t.TempDir()
	state := filepath.Join(project, "local-state")
	store := filepath.Join(project, "source.db")
	base := []string{"--repo", project, "--state-dir", state, "--source", "opencode", "--store", store, "--binary", cliBinaryPath}
	if out, err := captureCLI(t, project, "", append([]string{"install"}, base...)...); err != nil {
		t.Fatalf("install: %v %s", err, out)
	}
	pluginPath := filepath.Join(project, ".opencode", "plugins", "agenticlens-capture.ts")
	original := readCaptureFile(t, pluginPath)
	edited := append(append([]byte{}, original...), []byte("\n// user-owned edit\n")...)
	if err := os.WriteFile(pluginPath, edited, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, project, "", append([]string{"uninstall"}, base...)...); err == nil {
		t.Fatalf("removed locally edited plugin: %s", out)
	}
	if !bytes.Equal(readCaptureFile(t, pluginPath), edited) {
		t.Fatal("plugin conflict lost local edit")
	}
	if err := os.WriteFile(pluginPath, original, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := captureCLI(t, project, "", append([]string{"uninstall"}, base...)...); err != nil {
		t.Fatalf("uninstall: %v %s", err, out)
	}
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		t.Fatalf("owned plugin still exists: %v", err)
	}
	missingProject := t.TempDir()
	if out, err := captureCLI(t, missingProject, "", "install", "--repo", missingProject, "--state-dir", filepath.Join(missingProject, "state"), "--source", "codex", "--store", store, "--binary", filepath.Join(missingProject, "absent-handler")); err == nil {
		t.Fatalf("activated absent handler: %s", out)
	}
	if _, err := os.Stat(filepath.Join(missingProject, ".codex", "hooks.json")); !os.IsNotExist(err) {
		t.Fatalf("missing handler left active config: %v", err)
	}
}

func TestCaptureCLIRequestExposesUnassignedTicketToPlugin(t *testing.T) {
	project := newCaptureGenerationFixture(t).project
	out, err := captureCLI(t, project, "", "request", "--source", "opencode", "--store", filepath.Join(project, "source.db"), "--state-dir", filepath.Join(project, "state"), "--session-id", "runtime-session")
	if err != nil {
		t.Fatalf("request: %v %s", err, out)
	}
	var response map[string]json.RawMessage
	if err := json.Unmarshal(out, &response); err != nil {
		t.Fatal(err)
	}
	if string(response["ticket_id"]) != `""` {
		t.Fatalf("plugin cannot distinguish an unassigned response without explicit empty ticket_id: %s", out)
	}
	var token string
	if err := json.Unmarshal(response["request"], &token); err != nil || token == "" {
		t.Fatalf("missing session binding request: %s err=%v", out, err)
	}
}

func TestCaptureCLIGenerateIncludesBoundIndependentSessionsAndSummaryDiagnostics(t *testing.T) {
	project := t.TempDir()
	configuration := "project_name: CaptureFixture\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: custom records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(configuration), 0600); err != nil {
		t.Fatal(err)
	}
	state, store := filepath.Join(project, "state with spaces"), t.TempDir()
	t.Setenv("CODEX_HOME", store)
	fixture := captureGenerationFixture{project: project, state: state, store: store}
	if err := writeCaptureFixtureLocal(t, fixture, []byte("telemetry:\n  capture:\n")); err != nil {
		t.Fatal(err)
	}
	serializeCaptureFixtureWorkers(t, fixture)
	var requestTokens []string
	request := func(id string, tokens int, ticket string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(store, "sessions"), 0700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(store, "sessions", id+".jsonl")
		contents := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"` + id + `","cwd":` + strconv.Quote(project) + `,"timestamp":"2026-01-01T00:00:00Z"}}` + "\n" + `{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"turn-one"}}` + "\n"
		usage := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}`, id, tokens)
		if err := os.WriteFile(path, []byte(contents+usage+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		payload, err := json.Marshal(capture.HookInput{SessionID: id, Event: "SessionStart", TranscriptPath: path})
		if err != nil {
			t.Fatal(err)
		}
		out, err := captureCLI(t, project, string(payload), "receive", "--source", "codex", "--store", store, "--state-dir", state)
		if err != nil {
			t.Fatalf("receive: %v %s", err, out)
		}
		var feedback capture.HookOutput
		if err := json.Unmarshal(out, &feedback); err != nil || feedback.Context == nil {
			t.Fatalf("agent feedback=%s err=%v", out, err)
		}
		if ticket == "" {
			return
		}
		out, err = captureCLI(t, project, "", "request", "--source", "codex", "--store", store, "--state-dir", state, "--session-id", id)
		if err != nil {
			t.Fatalf("request: %v %s", err, out)
		}
		var registration domain.CaptureRegistration
		if err := json.Unmarshal(out, &registration); err != nil {
			t.Fatal(err)
		}
		requestTokens = append(requestTokens, registration.Request)
		if !strings.Contains(feedback.Context.Context, registration.Request) {
			t.Fatal("agent response token differs from durable session request")
		}
		out, err = captureCLI(t, project, "", "bind", "--source", "codex", "--store", store, "--state-dir", state, "--session-id", id, "--request", registration.Request, "--ticket-id", ticket)
		if err != nil {
			t.Fatalf("bind: %v %s", err, out)
		}
	}
	request("independent-one", 10, "TEST-100")
	request("independent-two", 20, "TEST-100")
	request("other-ticket", 50, "TEST-200")
	request("unassigned", 70, "")
	output := filepath.Join(project, "metrics.json")
	summary := filepath.Join(project, "SUMMARY.md")
	if err := os.WriteFile(summary, []byte("# Capture fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	generate := func() domain.AgenticFactLedger {
		t.Helper()
		cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--capture-dir", state, "--ticket-id", "TEST-100", "--ticket-title", "Explicit independent roots", "--ticket-type", "feat", "--json-output", output, "--append", summary)
		cmd.Dir = project
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generate: %v %s", err, out)
		}
		b := readCaptureFile(t, output)
		var ledger domain.AgenticFactLedger
		if err := json.Unmarshal(b, &ledger); err != nil {
			t.Fatal(err)
		}
		summaryBytes := readCaptureFile(t, summary)
		for _, public := range [][]byte{summaryBytes} {
			for _, forbidden := range append(append([]string{store, state}, requestTokens...), "\"transcript_path\"", "\"snapshot\"") {
				if forbidden != "" && bytes.Contains(public, []byte(forbidden)) {
					t.Fatalf("publication leaked private capture data %q", forbidden)
				}
			}
		}
		if bytes.Contains(summaryBytes, []byte("capture_report.json")) || !bytes.Contains(summaryBytes, []byte("Capture observation finalized:")) {
			t.Fatalf("summary lacks current capture coverage or links retired report: %s", summaryBytes)
		}
		if _, err := os.Stat(filepath.Join(project, "capture_report.json")); !os.IsNotExist(err) {
			t.Fatalf("summary-only publication created report: %v", err)
		}
		if ledger.SchemaVersion == nil || *ledger.SchemaVersion != domain.CurrentSchemaVersion || len(ledger.Sessions) != 2 {
			t.Fatalf("wrong canonical result: %+v", ledger)
		}
		for _, session := range ledger.Sessions {
			if session.ParentID != nil || (session.ID != "independent-one" && session.ID != "independent-two") {
				t.Fatalf("guessed identity/parent: %+v", session)
			}
		}
		if got := ledger.InvocationFacts()[0].Tokens.Input + ledger.InvocationFacts()[1].Tokens.Input; got != 30 {
			t.Fatalf("usage=%d want 30", got)
		}
		assertCapturePublishedOwners(t, captureDiagnostics(t, project, state), output, summary, ledger.Sessions)
		return ledger
	}
	first := generate()
	if second := generate(); !reflect.DeepEqual(first.Sessions, second.Sessions) {
		t.Fatal("repeated capture generation duplicated cumulative facts")
	}
	if err := os.Remove(filepath.Join(store, "sessions", "independent-two.jsonl")); err != nil {
		t.Fatal(err)
	}
	if missing := generate(); !reflect.DeepEqual(first.Sessions, missing.Sessions) {
		t.Fatal("unavailable registered source hid accepted facts")
	}
	for _, want := range []string{"retained older facts: 1", "failed: 1"} {
		if !bytes.Contains(readCaptureFile(t, summary), []byte(want)) {
			t.Fatalf("summary hid unavailable-source coverage %q", want)
		}
	}
	failedSource := false
	for _, d := range captureDiagnostics(t, project, state) {
		if d.Phase == "reconciliation" && d.Outcome == "failed" && d.Reconciliation != nil && d.Reconciliation.Cutoff != "" {
			for _, row := range d.Reconciliation.Owners {
				if row.Identity.SessionID == "independent-two" && row.State == domain.CaptureFailed && row.RefreshedAt != "" {
					failedSource = true
				}
			}
		}
	}
	if !failedSource {
		t.Fatal("successful partial publication hid durable source failure/retained snapshot")
	}
}

func TestCaptureGenerationRetainedRegistryCannotOverrideNewerCanonicalOrExplicitSnapshot(t *testing.T) {
	for _, mode := range []string{"newer canonical", "newer explicit"} {
		t.Run(mode, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 10)
			f.bind(t, path)
			assertCaptureRootInput(t, f.generate(t), 10)
			f.writeAt(t, 30, "2026-01-01T00:00:03Z")
			if mode == "newer canonical" {
				newer := f.generate(t, "--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store)
				assertCaptureRootInput(t, newer, 30)
				newer.Sessions = append(newer.Sessions, domain.SessionFact{ID: "other-harness", Source: domain.SourceOpenCode, User: "stored-user", Invocations: []domain.InvocationFact{{Role: "reviewer", Model: "stored-model", Tokens: domain.InvocationTokens{Input: 77}}}})
				newer.Sessions[0].Skills = []domain.SkillFact{{Name: "stored-skill"}, {Name: "stored-skill"}}
				link := domain.DelegationFact{AgentType: domain.KnownString("reviewer"), Description: "stored delegation", ChildSession: "stored-child", InvokedAt: domain.KnownString("2026-01-01T00:00:00Z")}
				newer.Sessions[0].Delegations = []domain.DelegationFact{link, link}
				acceptFixtureOwners(t, f.project, f.state, "TEST-100", newer.Sessions)
				// The owned snapshot above represents an accepted native publication.
				accepted, err := capture.OpenProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}

				if err = accepted.Close(); err != nil {
					t.Fatal(err)
				}
				body, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(newer))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f.output, body, 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				result := f.generate(t)
				assertCaptureRootInput(t, result, 30)
				if !reflect.DeepEqual(result.SkillFacts(), newer.SkillFacts()) {
					t.Fatalf("retained registry lost AL-62 row multiplicity: %+v", result.SkillFacts())
				}
				if !reflect.DeepEqual(result.DelegationFacts(), newer.DelegationFacts()) {
					t.Fatalf("retained registry lost AL-62 delegation multiplicity: %+v", result.DelegationFacts())
				}
				found := false
				for _, row := range result.Sessions {
					if row.ID == "other-harness" {
						found = true
						if row.Invocations[0].Tokens.Input != 77 {
							t.Fatal("other-harness baseline changed")
						}
					}
				}
				if !found {
					t.Fatal("other-harness baseline disappeared")
				}
			} else {
				archive := filepath.Join(f.store, "archived_sessions")
				if err := os.MkdirAll(archive, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(path, filepath.Join(archive, "root.jsonl")); err != nil {
					t.Fatal(err)
				}
				assertCaptureRootInput(t, f.generate(t, "--session-id", "root", "--codex-home", f.store), 30)
			}
		})
	}
	t.Run("successful unchanged older source checkpoint", func(t *testing.T) {
		f := newCaptureGenerationFixture(t)
		path := f.write(t, 10)
		older := readCaptureFile(t, path)
		f.bind(t, path)
		assertCaptureRootInput(t, f.generate(t), 10)
		// Publish from a demonstrably later source checkpoint, then restore the
		// exact older bytes still accepted by the independent local registry.
		later := append(append([]byte{}, older...), []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"two","usage":{"input_tokens":20,"output_tokens":3}}}`+"\n")...)
		if err := os.WriteFile(path, later, 0600); err != nil {
			t.Fatal(err)
		}
		assertCaptureRootInput(t, f.generate(t, "--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store), 30)
		if err := os.WriteFile(path, older, 0600); err != nil {
			t.Fatal(err)
		}
		assertCaptureRootInput(t, f.generate(t), 10)
	})
}

func TestCaptureGenerationCompleteSnapshotsReplaceRegardlessOfSourceOrder(t *testing.T) {
	for _, tc := range []struct {
		name              string
		explicit, unknown bool
	}{
		{"registered equal source checkpoint", false, false},
		{"registered unknown canonical source checkpoint", false, true},
		{"explicit equal source checkpoint", true, false},
		{"explicit unknown canonical source checkpoint", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 10))
			baseline := f.generate(t)
			if tc.unknown {
				baseline.Sessions[0].Invocations[0].CompletedAt = nil
				body, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(baseline))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f.output, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			f.write(t, 30) // Different facts, deliberately no later source timestamp.
			summary := filepath.Join(f.project, "SUMMARY.md")
			args := f.args(t, "--append", summary)
			if tc.explicit {
				args = f.args(t, "--append", summary, "--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store)
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = f.project
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("unordered source refresh: %v %s", err, out)
			}
			result := f.ledger(t)
			assertCaptureRootInput(t, result, 30)
			if bytes.Contains(out, []byte("unresolved precedence")) {
				t.Fatalf("BEH-02 complete exact source was treated as an unordered partial read: %s", out)
			}
			if !tc.explicit {
				registry, err := capture.OpenReadOnlyProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}
				rows, listErr := registry.List(context.Background())
				closeErr := registry.Close()
				if listErr != nil || closeErr != nil {
					t.Fatalf("read capture snapshot: %v %v", listErr, closeErr)
				}
				if len(rows) != 1 || len(rows[0].Snapshot) != 1 || rows[0].Snapshot[0].Invocations[0].Tokens.Input != 30 {
					t.Fatalf("complete registered source did not become accepted snapshot: %+v", rows)
				}
			}

		})
	}
}

func TestCaptureGenerationRejectsExplicitSelectionOfAnotherTicketOwner(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 10))
	cmd := exec.Command(cliBinaryPath, f.args(t, "--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", f.store)...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("explicit selector stole known TEST-100 owner: %s", out)
	}
	if _, err := os.Stat(f.output); !os.IsNotExist(err) {
		t.Fatalf("conflicting selector published ledger: %v", err)
	}
}

func TestCaptureGenerationResolvesStoreAliasesWithoutConflatingDistinctStores(t *testing.T) {
	for _, bindingUsesAlias := range []bool{false, true} {
		t.Run(fmt.Sprintf("stored alias %v", bindingUsesAlias), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			actualStore := f.store
			path := f.write(t, 10)
			alias := filepath.Join(t.TempDir(), "source-alias")
			if err := os.Symlink(actualStore, alias); err != nil {
				t.Fatal(err)
			}
			selected := alias
			if bindingUsesAlias {
				f.store = alias
				selected = actualStore
			}
			f.bind(t, path)
			if bindingUsesAlias {
				// Seed an actual pre-normalization registry row. Current Register
				// canonicalizes aliases, which alone would miss the legacy guard.
				db, err := sql.Open("sqlite", filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := db.Close(); err != nil {
						t.Error(err)
					}
				})
				var raw []byte
				if err := db.QueryRow(`SELECT record FROM registrations WHERE session_id='root'`).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var record map[string]json.RawMessage
				if err := json.Unmarshal(raw, &record); err != nil {
					t.Fatal(err)
				}
				var owner domain.CaptureRegistration
				if err := json.Unmarshal(record["registration"], &owner); err != nil {
					t.Fatal(err)
				}
				owner.Identity.Store = alias
				record["registration"], err = json.Marshal(owner)
				if err != nil {
					t.Fatal(err)
				}
				raw, err = json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := db.Exec(`UPDATE registrations SET identity=?,store=?,record=? WHERE session_id='root'`, capture.IdentityKey(owner.Identity), alias, raw); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(cliBinaryPath, f.args(t, "--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", selected)...)
			cmd.Dir = f.project
			if out, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("actual-store alias bypassed known ticket owner: %s", out)
			}
			if _, err := os.Stat(f.output); !os.IsNotExist(err) {
				t.Fatalf("alias conflict published output: %v", err)
			}
			// A second real store may legitimately reuse a harness-local ID. Its
			// different native source must not inherit the first store's binding.
			distinct := t.TempDir()
			if err := os.MkdirAll(filepath.Join(distinct, "sessions"), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(distinct, "sessions", "root.jsonl"), readCaptureFile(t, path), 0600); err != nil {
				t.Fatal(err)
			}
			cmd = exec.Command(cliBinaryPath, f.args(t, "--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", distinct)...)
			cmd.Dir = f.project
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("truly distinct store was conflated with known owner: %v %s", err, out)
			}
			ledger := f.ledger(t)
			if ledger.TicketID == nil || *ledger.TicketID != "TEST-200" {
				t.Fatalf("distinct-store result has wrong ticket: %+v", ledger.TicketID)
			}
			assertCaptureRootInput(t, ledger, 10)
		})
	}
}

func TestCaptureGenerationOverlappingPublishersCannotRegressFinalSnapshot(t *testing.T) {
	t.Run("registered newer publisher", func(t *testing.T) { testCaptureOverlappingPublishers(t, false, false) })
	t.Run("explicit newer publisher without registry", func(t *testing.T) { testCaptureOverlappingPublishers(t, true, false) })
	t.Run("automatic newer publisher", func(t *testing.T) { testCaptureOverlappingPublishers(t, false, true) })
}

func testCaptureOverlappingPublishers(t *testing.T, explicitNewer, automatic bool) {
	t.Helper()
	f := newCaptureGenerationFixture(t)
	if automatic {
		f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
		if err := writeCaptureFixtureLocal(t, f, []byte("telemetry:\n  capture:\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 10000\n")); err != nil {
			t.Fatal(err)
		}
	}
	f.bind(t, f.write(t, 10))
	f.generate(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	helper, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shimDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(shimDir, "git"), readCaptureFile(t, helper), 0700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	tcp := listener.(*net.TCPListener)
	if err := tcp.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	older := exec.CommandContext(ctx, cliBinaryPath, f.args(t)...)
	older.Dir = f.project
	older.Env = append(os.Environ(), "PATH="+shimDir+string(os.PathListSeparator)+os.Getenv("PATH"), "AL_CAPTURE_TEST_GIT_HELPER=1", "AL_CAPTURE_TEST_GATE_ADDRESS="+listener.Addr().String(), "AL_CAPTURE_TEST_REAL_GIT="+realGit)
	var olderOutput bytes.Buffer
	older.Stdout, older.Stderr = &olderOutput, &olderOutput
	if err := older.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if older.ProcessState == nil {
			cancel()
			if err := older.Wait(); err != nil {
				t.Logf("gate cleanup: %v", err)
			}
		}
	}()
	connection, err := listener.Accept()
	if err != nil {
		t.Fatalf("older publisher did not reach Git gate: %v %s", err, olderOutput.String())
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadFull(connection, make([]byte, 1)); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(f.store, "sessions", "root.jsonl")
	laterSource := append(readCaptureFile(t, path), []byte(`{"timestamp":"2026-01-01T00:00:03Z","type":"token_usage_record","payload":{"thread_id":"root","turn_id":"two","usage":{"input_tokens":40,"output_tokens":3}}}`+"\n")...)
	if err := os.WriteFile(path, laterSource, 0600); err != nil {
		t.Fatal(err)
	}
	newerArgs := f.args(t)
	if explicitNewer {
		newerArgs = f.args(t, "--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store)
	}
	if automatic {
		registry, err := capture.OpenProject(f.state, f.project)
		if err != nil {
			t.Fatal(err)
		}
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
		if _, err = registry.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			r.TicketTitle = "Capture regression"
			r.TicketType = "feat"
			r.Activity = "idle"
			return r, nil
		}); err != nil {
			registry.Close()
			t.Fatal(err)
		}
		if err = registry.RequestPublication(context.Background()); err != nil {
			registry.Close()
			t.Fatal(err)
		}
		if err = registry.Close(); err != nil {
			t.Fatal(err)
		}
		newerArgs = []string{"capture", "worker", "--repo", f.project, "--state-dir", f.state}
	}
	newer := exec.CommandContext(ctx, cliBinaryPath, newerArgs...)
	newer.Dir = f.project
	var newerOutput bytes.Buffer
	newer.Stdout, newer.Stderr = &newerOutput, &newerOutput
	if err := newer.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- newer.Wait() }()
	completed := false
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("newer publisher: %v %s", err, newerOutput.String())
		}
		completed = true
	case <-time.After(time.Second):
		// A publication lock may serialize the newer process behind the gated owner.
		// The deadline releases that valid ordering; it is not a latency assertion.
	}
	if _, err := connection.Write([]byte("C")); err != nil {
		t.Fatal(err)
	}
	if err := older.Wait(); err != nil {
		t.Fatalf("older publisher: %v %s", err, olderOutput.String())
	}
	if !completed {
		if err := <-done; err != nil {
			t.Fatalf("serialized newer publisher: %v %s", err, newerOutput.String())
		}
	}
	assertCaptureRootInput(t, f.ledger(t), 50)
	if _, err := os.Stat(filepath.Join(filepath.Dir(f.output), "capture_report.json")); !os.IsNotExist(err) {
		t.Fatalf("overlapping publisher produced retired report: %v", err)
	}
	if !explicitNewer {
		summary := ""
		if automatic {
			summary = filepath.Join(filepath.Dir(f.output), "SUMMARY.md")
		}
		assertCapturePublishedOwners(t, captureDiagnostics(t, f.project, f.state), f.output, summary, f.ledger(t).Sessions)
	}
}

func TestCaptureReceiveNativeTimeoutBoundsOpenStdin(t *testing.T) {
	project := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliBinaryPath, "capture", "receive", "--source", "codex", "--store", project, "--state-dir", filepath.Join(project, "state"), "--timeout", "50ms")
	cmd.Dir = project
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(input, `{"session_id":"stalled","hook_event_name":"SessionStart"}`); err != nil {
		t.Fatal(err)
	}
	err = cmd.Wait()
	if ctx.Err() != nil {
		t.Fatalf("native timeout did not cover open stdin: %v %s", ctx.Err(), output.String())
	}
	if err == nil {
		t.Fatal("open stdin was accepted without complete bounded handoff")
	}
	s, err := capture.Open(filepath.Join(project, "state"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rows, err := s.List(context.Background())
	if err != nil || len(rows) != 0 {
		t.Fatalf("timed-out input committed event: %+v %v", rows, err)
	}
}

func TestCaptureInstallerRendersMarkerBearingPathsOnceAndRollsBackMidWrite(t *testing.T) {
	t.Run("literal marker paths", func(t *testing.T) {
		binary := filepath.Join(t.TempDir(), "native __STATE__ __STORE__")
		if err := os.WriteFile(binary, readCaptureFile(t, cliBinaryPath), 0700); err != nil {
			t.Fatal(err)
		}
		project := t.TempDir()
		state := filepath.Join(project, `state"quote __BINARY__`)
		store := filepath.Join(project, "store __STATE__")
		out, err := captureCLI(t, project, "", "install", "--source", "opencode", "--repo", project, "--state-dir", state, "--store", store, "--binary", binary)
		if err != nil {
			t.Fatalf("install literal paths: %v %s", err, out)
		}
		plugin := string(readCaptureFile(t, filepath.Join(project, ".opencode", "plugins", "agenticlens-capture.ts")))
		scoped, scopeErr := capture.ProjectState(state, project)
		if scopeErr != nil {
			t.Fatal(scopeErr)
		}
		for name, want := range map[string]string{"wrapper": fixtureWrapperFromHook(t, project), "stateDir": scoped, "sourceStore": store} {
			prefix := "const " + name + " = "
			found := false
			for _, line := range strings.Split(plugin, "\n") {
				if strings.HasPrefix(line, prefix) {
					found = true
					var got string
					if err := json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimPrefix(line, prefix), ";")), &got); err != nil || got != want {
						t.Fatalf("template reinterpreted %s path: got=%q want=%q err=%v line=%s", name, got, want, err, line)
					}
				}
			}
			if !found {
				t.Fatalf("installed plugin lacks %s declaration", name)
			}
		}
	})
	t.Run("rollback after config and ignore writes", func(t *testing.T) {
		project := t.TempDir()
		ignore := filepath.Join(project, ".gitignore")
		original := []byte("# user-owned\ncustom-secret\n")
		if err := os.WriteFile(ignore, original, 0600); err != nil {
			t.Fatal(err)
		}
		config := filepath.Join(project, ".codex", "hooks.json")
		state := filepath.Join(config, "state")
		out, err := captureCLI(t, project, "", "install", "--source", "codex", "--repo", project, "--state-dir", state, "--store", t.TempDir(), "--binary", cliBinaryPath)
		if err == nil {
			t.Fatalf("impossible manifest path unexpectedly succeeded: %s", out)
		}
		if !bytes.Equal(readCaptureFile(t, ignore), original) {
			t.Fatal("mid-write failure lost original ignore file")
		}
		if _, err := os.Stat(config); !os.IsNotExist(err) {
			t.Fatalf("mid-write failure left new active config: %v", err)
		}
	})
}

// This launches the real native handler/worker, but uses source fixtures. It is
// intentionally separate from the real harness support evidence.
func TestCaptureNativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Probe"}, {"config", "user.email", "probe@example.invalid"}, {"commit", "--allow-empty", "-m", "probe"}, {"update-ref", "refs/remotes/origin/main", "HEAD"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = f.project
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	state := filepath.Join(f.project, ".agenticlens-capture")
	writeCaptureSuccessfulPublicationPolicy(t, f.project, state, true)
	receive := func(id, event string) {
		t.Helper()
		payload, _ := json.Marshal(capture.HookInput{SessionID: id, Event: event, TranscriptPath: filepath.Join(f.store, "sessions", id+".jsonl")})
		if out, err := captureCLI(t, f.project, string(payload), "receive", "--source", "codex", "--store", f.store, "--state-dir", state, "--binary", filepath.Join(f.project, "not-a-native-wrapper")); err != nil {
			t.Fatalf("receive: %v %s", err, out)
		}
	}
	write := func(id string, tokens int, stamp string) {
		t.Helper()
		body := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":%q,"cwd":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, id, f.project, stamp, id, tokens)
		if err := os.WriteFile(filepath.Join(f.store, "sessions", id+".jsonl"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"one", "two"} {
		write(id, 10, "2026-01-01T00:00:02Z")
		receive(id, "SessionStart")
		requestCommand := exec.Command(cliBinaryPath, "capture", "request", "--source", "codex", "--store", f.store, "--state-dir", state, "--session-id", id)
		requestCommand.Dir = f.project
		var requestError bytes.Buffer
		requestCommand.Stderr = &requestError
		out, err := requestCommand.Output()
		if err != nil {
			t.Fatalf("request %v %s", err, requestError.String())
		}
		if !strings.Contains(requestError.String(), "detailed_report is deprecated") {
			t.Fatalf("legacy option diagnosis missing from stderr: %s", requestError.String())
		}
		var row domain.CaptureRegistration
		if err = json.Unmarshal(out, &row); err != nil {
			t.Fatal(err)
		}
		if out, err = captureCLI(t, f.project, "", "bind", "--source", "codex", "--store", f.store, "--state-dir", state, "--session-id", id, "--ticket-id", "TEST-100", "--ticket-title", "Explicit context", "--ticket-type", "feat", "--request", row.Request); err != nil {
			t.Fatalf("bind %v %s", err, out)
		}
	}
	defer func() {
		receive("one", "SessionEnd")
		receive("two", "SessionEnd")
		scoped, err := capture.ProjectState(state, f.project)
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
					if e = lock.Close(); e != nil {
						t.Error(e)
					}
					return
				}
			}
			time.Sleep(40 * time.Millisecond)
		}
		t.Error("native worker did not finish terminal idle shutdown before fixture cleanup")
	}()
	output := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	wait := func(want int64) domain.AgenticFactLedger {
		t.Helper()
		deadline := time.Now().Add(8 * time.Second)
		for time.Now().Before(deadline) {
			body, err := os.ReadFile(output)
			if err == nil {
				var ledger domain.AgenticFactLedger
				if json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) == 2 {
					var total int64
					for _, s := range ledger.Sessions {
						for _, i := range s.Invocations {
							total += i.Tokens.Input
						}
					}
					if total == want {
						return ledger
					}
				}
			}
			time.Sleep(40 * time.Millisecond)
		}
		t.Fatalf("automatic publication did not reach %d; output %s", want, output)
		return domain.AgenticFactLedger{}
	}
	first := wait(20)
	if first.CompletedAt == nil || *first.CompletedAt != first.UpdatedAt {
		t.Fatalf("snapshot completion not finalized: %+v", first)
	}
	write("one", 50, "2026-01-01T00:40:02Z")
	// No callback and no manual generate: active periodic refresh must see growth.
	next := wait(60)
	for _, session := range next.Sessions {
		want := "2026-01-01T00:00:02Z"
		if session.ID == "one" {
			want = "2026-01-01T00:40:02Z"
		}
		if len(session.Invocations) != 1 || session.Invocations[0].CompletedAt == nil || *session.Invocations[0].CompletedAt != want {
			t.Fatalf("source invocation time changed: %+v want %s", session, want)
		}
	}
	if next.CompletedAt == nil || *next.CompletedAt != next.UpdatedAt {
		t.Fatalf("active publication did not finalize: %+v", next)
	}
	summary, err := os.ReadFile(filepath.Join(filepath.Dir(output), "SUMMARY.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(summary), "](capture_report.json)") {
		t.Fatal("deprecated detailed_report produced dangling link")
	}
	if _, err = os.Stat(filepath.Join(filepath.Dir(output), "capture_report.json")); !os.IsNotExist(err) {
		t.Fatalf("deprecated detailed_report produced sidecar: %v", err)
	}
}

// Existing capture fixtures explicitly supply a complete selected local policy.
// Tests of omission use config's dedicated whole-file selection cases instead.
func writeCaptureFixtureLocal(t *testing.T, f captureGenerationFixture, body []byte) error {
	t.Helper()
	text := string(body)
	extra := ""
	if !strings.Contains(text, "    enabled:") {
		extra += "    enabled: true\n"
	}
	if !strings.Contains(text, "    state_dir:") {
		state := f.stateBase
		if state == "" {
			state = f.state
		}
		extra += "    state_dir: '" + strings.ReplaceAll(state, "'", "''") + "'\n"
	}
	for _, setting := range []struct{ key, value string }{{"refresh_ms", "100"}, {"debounce_ms", "0"}, {"idle_ms", "300"}, {"max_lifetime_ms", "5000"}, {"pass_timeout_ms", "3000"}} {
		if !strings.Contains(text, "    "+setting.key+":") {
			extra += "    " + setting.key + ": " + setting.value + "\n"
		}
	}
	text = strings.Replace(text, "  capture:\n", "  capture:\n"+extra, 1)
	return os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(text), 0600)
}

// These fixtures test source guards or explicit publication, while autonomous
// worker behavior has its own process tests. Hold the actual worker lease and
// drain every launched child's inherited log handle before releasing its files.
func serializeCaptureFixtureWorkers(t *testing.T, f captureGenerationFixture) {
	t.Helper()
	state, err := capture.ProjectState(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := capture.AcquireWorker(store.CoordinationDirectory())
	if err != nil || lease == nil {
		store.Close()
		t.Fatalf("fixture worker serialization: %v", err)
	}
	completion := captureWorkerCompletion(t, store.CoordinationDirectory())
	t.Cleanup(func() {
		captureCleanupWorkers(t, completion)
		if err := lease.Close(); err != nil {
			t.Error(err)
		}
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
}

func TestAutomaticGenerationUsesSelectedLocalDefaultsRatherThanSharedBudget(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 10))
	assertCaptureRootInput(t, f.generate(t), 10)
	sharedPath := filepath.Join(f.project, ".agenticlens-config.yaml")
	shared := strings.Replace(string(readCaptureFile(t, sharedPath)), "  capture:\n", "  capture:\n    max_source_bytes: 1\n", 1)
	if err := os.WriteFile(sharedPath, []byte(shared), 0600); err != nil {
		t.Fatal(err)
	}
	// Omission means documented automatic defaults, not machine keys from shared YAML.
	local := fmt.Sprintf("telemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n  sources:\n    codex:\n      home: %q\n", f.stateBase, f.store)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	policy, err := adaptersconfig.ResolveCapture(f.project, "")
	if err != nil || policy.MaxSourceBytes != capture.DefaultByteBudget {
		t.Fatalf("worker's local default policy: %+v %v", policy, err)
	}
	f.writeAt(t, 31, "2026-01-01T00:00:03Z")
	assertCaptureRootInput(t, f.generate(t, "--capture-auto"), 31)
}

func TestAutomaticParentAndExplicitOtherTicketChildPublishIndependently(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	rootPath := f.write(t, 10)
	f.bind(t, rootPath)
	childPath := filepath.Join(f.store, "sessions", "child.jsonl")
	body := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"child","cwd":%q,"parent_thread_id":"root","timestamp":"2026-01-01T00:00:00Z"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"child-model","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"child","turn_id":"child-turn","usage":{"input_tokens":31,"output_tokens":2}}}
`, f.project)
	if err := os.WriteFile(childPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "child"}
	row, err := store.Register(context.Background(), id, childPath, "SessionEnd")
	if err != nil {
		store.Close()
		t.Fatal(err)
	}
	if _, err := store.Bind(context.Background(), id, "TEST-200", row.Request); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if err := store.QueueChild(context.Background(), f.store, childPath, "root"); err != nil {
		store.Close()
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("telemetry:\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n  capture:\n", f.store)
	if err := writeCaptureFixtureLocal(t, f, []byte(local)); err != nil {
		t.Fatal(err)
	}
	// Match the worker's registered-only automatic child: no --session-id can
	// broaden an exact-ticket collector into the selected owner's source root.
	generate := func(ticket, output string) domain.AgenticFactLedger {
		t.Helper()
		command := exec.Command(cliBinaryPath, "generate", "--capture-auto", "--capture-dir", f.state, "--ticket-id", ticket, "--ticket-title", "Explicit "+ticket, "--ticket-type", "feat", "--json-output", output, "--append", output+".md")
		command.Dir = f.project
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("automatic %s publication: %s %v", ticket, out, err)
		}
		var ledger domain.AgenticFactLedger
		if err := json.Unmarshal(readCaptureFile(t, output), &ledger); err != nil {
			t.Fatal(err)
		}
		return ledger
	}
	parent := generate("TEST-100", f.output)
	if len(parent.Sessions) != 1 || parent.Sessions[0].ID != "root" || parent.Sessions[0].Invocations[0].Tokens.Input != 10 {
		t.Fatalf("child blocked/crossed parent publication: %+v", parent.Sessions)
	}
	acceptedParent := readCaptureFile(t, f.output)
	child := generate("TEST-200", filepath.Join(f.project, "child-ledger.json"))
	if len(child.Sessions) != 1 || child.Sessions[0].ID != "child" || child.Sessions[0].ParentID == nil || *child.Sessions[0].ParentID != "root" || child.Sessions[0].Invocations[0].Tokens.Input != 31 {
		t.Fatalf("explicit child publication lost source ownership: %+v", child.Sessions)
	}
	if !bytes.Equal(acceptedParent, readCaptureFile(t, f.output)) {
		t.Fatal("child ticket publication changed parent ticket ledger")
	}
	if child.Anomalies != nil || child.Sessions[0].Anomalies == nil || child.CompleteAnomalyCoverage() || child.UnrecordedAnomalyScopes() != 1 {
		t.Fatalf("fresh complete child invented external-parent group coverage: %+v", child)
	}
	summary := string(readCaptureFile(t, filepath.Join(f.project, "child-ledger.json.md")))
	if strings.Contains(summary, "No execution anomalies detected (clean session)") || !strings.Contains(summary, "no clean-session conclusion is available") {
		t.Fatalf("external-parent summary misrepresented partial analysis: %s", summary)
	}
}
