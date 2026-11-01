package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
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

type captureGenerationFixture struct{ project, store, state, output string }

func newCaptureGenerationFixture(t *testing.T) captureGenerationFixture {
	t.Helper()
	f := captureGenerationFixture{project: t.TempDir(), store: t.TempDir(), state: t.TempDir()}
	f.output = filepath.Join(f.project, "ledger.json")
	config := "project_name: CaptureFixture\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(f.store, "sessions"), 0700); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f captureGenerationFixture) write(t *testing.T, tokens int) string {
	return f.writeAt(t, tokens, "2026-01-01T00:00:02Z")
}

func (f captureGenerationFixture) writeAt(t *testing.T, tokens int, sourceTime string) string {
	t.Helper()
	path := filepath.Join(f.store, "sessions", "root.jsonl")
	contents := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"root","timestamp":"2026-01-01T00:00:00Z"}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":"root","turn_id":"one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, sourceTime, tokens)
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func (f captureGenerationFixture) bind(t *testing.T, path string) {
	t.Helper()
	store, err := capture.Open(f.state)
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

func (f captureGenerationFixture) args(extra ...string) []string {
	return append([]string{"generate", "--source", "codex", "--capture-dir", f.state, "--ticket-id", "TEST-100", "--ticket-title", "Capture regression", "--ticket-type", "feat", "--json-output", f.output}, extra...)
}

func (f captureGenerationFixture) generate(t *testing.T, extra ...string) domain.AgenticFactLedger {
	t.Helper()
	cmd := exec.Command(cliBinaryPath, f.args(extra...)...)
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
	manifest, err := capture.ReadManifest(state, "codex")
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
	project := t.TempDir()
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

func TestCaptureCLIGenerateIncludesBoundIndependentSessionsAndDigestReport(t *testing.T) {
	project := t.TempDir()
	configuration := "project_name: CaptureFixture\nticket:\n  prefixes: [TEST]\nworkspaces:\n  dir: custom records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(configuration), 0600); err != nil {
		t.Fatal(err)
	}
	state, store := filepath.Join(project, "state with spaces"), t.TempDir()
	var requestTokens []string
	request := func(id string, tokens int, ticket string) {
		t.Helper()
		path := filepath.Join(store, id+".jsonl")
		contents := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"` + id + `","timestamp":"2026-01-01T00:00:00Z"}}` + "\n" + `{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"turn-one"}}` + "\n"
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
		var report struct {
			LedgerSHA256  string                       `json:"ledger_sha256"`
			Cutoff        string                       `json:"cutoff"`
			Registrations []domain.CaptureRegistration `json:"registrations"`
		}
		reportBytes := readCaptureFile(t, filepath.Join(project, "capture_report.json"))
		if err := json.Unmarshal(reportBytes, &report); err != nil {
			t.Fatal(err)
		}
		for _, public := range [][]byte{reportBytes, readCaptureFile(t, summary)} {
			for _, forbidden := range append(append([]string{store, state}, requestTokens...), "\"transcript_path\"", "\"snapshot\"") {
				if forbidden != "" && bytes.Contains(public, []byte(forbidden)) {
					t.Fatalf("publication leaked private capture data %q", forbidden)
				}
			}
		}
		digest := sha256.Sum256(b)
		if report.LedgerSHA256 != hex.EncodeToString(digest[:]) || report.Cutoff == "" {
			t.Fatalf("inconsistent report: %+v", report)
		}
		if ledger.SchemaVersion == nil || *ledger.SchemaVersion != 5 || len(ledger.Sessions) != 2 {
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
		return ledger
	}
	first := generate()
	if second := generate(); !reflect.DeepEqual(first.Sessions, second.Sessions) {
		t.Fatal("repeated capture generation duplicated cumulative facts")
	}
	if err := os.Remove(filepath.Join(store, "independent-two.jsonl")); err != nil {
		t.Fatal(err)
	}
	if missing := generate(); !reflect.DeepEqual(first.Sessions, missing.Sessions) {
		t.Fatal("unavailable registered source hid accepted facts")
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
				newer.Skills = []domain.SkillFact{{Name: "stored-skill"}, {Name: "stored-skill"}}
				link := domain.TaskLinkFact{SubagentType: "reviewer", Description: "stored delegation", ChildSession: "stored-child", Timestamp: "2026-01-01T00:00:00Z", CostEvidence: domain.CostEvidence{State: domain.CostUnavailable}}
				newer.TaskLinks = []domain.TaskLinkFact{link, link}
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
				if !reflect.DeepEqual(result.Skills, newer.Skills) {
					t.Fatalf("retained registry lost AL-62 row multiplicity: %+v", result.Skills)
				}
				if !reflect.DeepEqual(result.TaskLinks, newer.TaskLinks) {
					t.Fatalf("retained registry lost AL-62 delegation multiplicity: %+v", result.TaskLinks)
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
		assertCaptureRootInput(t, f.generate(t), 30)
	})
}

func TestCaptureGenerationPreservesDifferingSnapshotsWithoutProvenSourceOrder(t *testing.T) {
	for _, tc := range []struct {
		name              string
		explicit, unknown bool
	}{
		{"registered equal source checkpoint", false, false},
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
			args := f.args()
			if tc.explicit {
				args = f.args("--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store)
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = f.project
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("unordered source refresh: %v %s", err, out)
			}
			result := f.ledger(t)
			if !reflect.DeepEqual(result.Sessions, baseline.Sessions) {
				t.Fatalf("unordered source facts replaced accepted whole owner: before=%+v after=%+v", baseline.Sessions, result.Sessions)
			}
			if tc.explicit {
				if !bytes.Contains(out, []byte("unresolved")) || !bytes.Contains(out, []byte("checkpoint")) {
					t.Fatalf("unordered explicit source did not report unresolved precedence: %s", out)
				}
			} else {
				var report struct {
					Registrations []domain.CaptureRegistration `json:"registrations"`
				}
				if err := json.Unmarshal(readCaptureFile(t, filepath.Join(f.project, "capture_report.json")), &report); err != nil {
					t.Fatal(err)
				}
				if len(report.Registrations) != 1 || report.Registrations[0].State != domain.CapturePending {
					t.Fatalf("unordered registered refresh claimed current coverage: %+v", report)
				}
			}
		})
	}
}

func TestCaptureGenerationRejectsExplicitSelectionOfAnotherTicketOwner(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.bind(t, f.write(t, 10))
	cmd := exec.Command(cliBinaryPath, f.args("--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", f.store)...)
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
				db, err := sql.Open("sqlite", filepath.Join(f.state, "registry.db"))
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
			cmd := exec.Command(cliBinaryPath, f.args("--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", selected)...)
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
			cmd = exec.Command(cliBinaryPath, f.args("--ticket-id", "TEST-200", "--session-id", "root", "--codex-home", distinct)...)
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
	t.Run("registered newer publisher", func(t *testing.T) { testCaptureOverlappingPublishers(t, false) })
	t.Run("explicit newer publisher without registry", func(t *testing.T) { testCaptureOverlappingPublishers(t, true) })
}

func testCaptureOverlappingPublishers(t *testing.T, explicitNewer bool) {
	t.Helper()
	f := newCaptureGenerationFixture(t)
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
	older := exec.CommandContext(ctx, cliBinaryPath, f.args()...)
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
	newerArgs := f.args()
	if explicitNewer {
		newerArgs = f.args("--capture-dir", filepath.Join(f.project, "absent-registry"), "--session-id", "root", "--codex-home", f.store)
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
	ledgerBytes := readCaptureFile(t, f.output)
	hash := sha256.Sum256(ledgerBytes)
	var report struct {
		Digest string `json:"ledger_sha256"`
	}
	if err := json.Unmarshal(readCaptureFile(t, filepath.Join(f.project, "capture_report.json")), &report); err != nil {
		t.Fatal(err)
	}
	if !explicitNewer && report.Digest != hex.EncodeToString(hash[:]) {
		t.Fatal("overlapping publication left a mismatched ledger/report pair")
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
		for name, want := range map[string]string{"binary": binary, "stateDir": state, "sourceStore": store} {
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
