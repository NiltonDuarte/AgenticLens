package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
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

type manualStateFile struct {
	Mode     fs.FileMode
	Modified time.Time
	Data     string
}

// File contents, names and modification metadata are observable state. Corrupt
// databases and a file used as a parent additionally make registry access fail.
func manualStateSnapshot(t *testing.T, root string) map[string]manualStateFile {
	t.Helper()
	result := map[string]manualStateFile{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		row := manualStateFile{Mode: info.Mode(), Modified: info.ModTime()}
		if info.Mode().IsRegular() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			row.Data = string(data)
		}
		result[relative] = row
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestManualGenerateIgnoresCaptureRegistryAndPolicy(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		for _, state := range []string{"omitted-policy", "no-capture-config", "invalid-capture-only-policy", "missing-registry", "corrupt-registry", "inaccessible-state", "conflicting-binding"} {
			t.Run(source+"/"+state, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				sourceArgs, sourceFile := manualSourceFixture(t, f, source)
				local := "telemetry:\n  default_source: " + source + "\n"
				captureDir := f.state
				switch state {
				case "no-capture-config":
					writeManualStateOnlyConfig(t, f.project)
				case "invalid-capture-only-policy":
					local += "  capture:\n    enabled: nonsense\n    state_dir: 42\n    max_sessions: -1\n    pass_timeout_ms: -1\n"
				case "corrupt-registry":
					if err := os.MkdirAll(capture.RegistryPathDirectory(f.state), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(capture.RegistryPathDirectory(f.state), "registry.db"), []byte("invalid SQLite capture registry"), 0600); err != nil {
						t.Fatal(err)
					}
				case "inaccessible-state":
					blocker := filepath.Join(f.stateBase, "file-parent")
					if err := os.WriteFile(blocker, []byte("not a directory"), 0600); err != nil {
						t.Fatal(err)
					}
					captureDir = filepath.Join(blocker, "capture")
				case "conflicting-binding":
					// A real registry contains an owner already bound to another ticket.
					registry, err := capture.OpenProject(f.state, f.project)
					if err != nil {
						t.Fatal(err)
					}
					store := f.store
					if source == "opencode" {
						store = sourceFile
					}
					id := domain.CaptureIdentity{Source: domain.SourceHarness(source), Store: store, SessionID: "root"}
					row, err := registry.Register(context.Background(), id, sourceFile, "SessionStart")
					if err != nil {
						t.Fatal(err)
					}
					if _, err = registry.Bind(context.Background(), id, "TEST-100", row.Request); err != nil {
						t.Fatal(err)
					}
					if err = registry.Close(); err != nil {
						t.Fatal(err)
					}
				}
				if state != "no-capture-config" {
					if state != "invalid-capture-only-policy" {
						local += fmt.Sprintf("  capture:\n    state_dir: %q\n", captureDir)
					}
					if state == "omitted-policy" {
						local = "telemetry:\n  default_source: " + source + "\n"
					}
					local += fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(f.project, ".publication-state"))
					if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
						t.Fatal(err)
					}
				}
				before := manualStateSnapshot(t, f.stateBase)
				args := []string{"generate", "--source", source, "--session-id", "root", "--ticket-id", "TEST-200", "--ticket-title", "Manual extraction", "--ticket-type", "fix", "--json-output", f.output}
				args = append(args, sourceArgs...)
				if state != "no-capture-config" && state != "omitted-policy" {
					args = append(args, "--capture-dir", captureDir)
				}
				for pass := 0; pass < 2; pass++ {
					command := exec.Command(cliBinaryPath, args...)
					command.Dir = f.project
					if output, err := command.CombinedOutput(); err != nil {
						t.Fatalf("manual pass %d: %v %s", pass, err, output)
					}
					ledger := f.ledger(t)
					if ledger.TicketID == nil || *ledger.TicketID != "TEST-200" || len(ledger.Sessions) != 1 || ledger.Sessions[0].ID != "root" || ledger.Sessions[0].Source != domain.SourceHarness(source) || len(ledger.InvocationFacts()) != 1 || ledger.InvocationFacts()[0].Tokens.Input != 17 {
						t.Fatalf("wrong authoritative source output: %+v", ledger)
					}
					if after := manualStateSnapshot(t, f.stateBase); !reflect.DeepEqual(before, after) {
						t.Fatalf("manual generation changed capture state: before=%+v after=%+v", before, after)
					}
				}
			})
		}
	}
}

func TestAutomaticGenerateExplicitSelectorsCannotCreateBindings(t *testing.T) {
	for _, selector := range []string{"--session-id", "--nonce"} {
		for _, binding := range []string{"unassigned", "conflicting"} {
			t.Run(selector+"/"+binding, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				path := f.write(t, 17)
				registry, err := capture.OpenProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}
				id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
				row, err := registry.Register(context.Background(), id, path, "SessionStart")
				if err != nil {
					t.Fatal(err)
				}
				if binding == "conflicting" {
					if _, err = registry.Bind(context.Background(), id, "TEST-100", row.Request); err != nil {
						t.Fatal(err)
					}
				}
				before, err := registry.List(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if err = registry.Close(); err != nil {
					t.Fatal(err)
				}
				args := []string{"generate", "--capture-auto", "--source", "codex", "--codex-home", f.store, "--capture-dir", f.state, "--ticket-id", "TEST-200", "--ticket-title", "Automatic admission", "--ticket-type", "fix", "--json-output", f.output, selector, "root"}
				command := exec.Command(cliBinaryPath, args...)
				command.Dir = f.project
				output, runErr := command.CombinedOutput()
				if runErr == nil {
					t.Fatalf("automatic selector bypass succeeded: %s", output)
				}
				if !bytes.Contains(output, []byte("capture-auto")) || (!strings.Contains(string(output), "selector") && !strings.Contains(string(output), selector)) {
					t.Fatalf("failed before explicit-selector guard: %s", output)
				}
				if _, err = os.Stat(f.output); !os.IsNotExist(err) {
					t.Fatalf("rejected automatic selector published: %v", err)
				}
				registry, err = capture.OpenReadOnlyProject(f.state, f.project)
				if err != nil {
					t.Fatal(err)
				}
				after, err := registry.List(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if err = registry.Close(); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before, after) {
					t.Fatalf("automatic selector changed binding: before=%+v after=%+v", before, after)
				}
			})
		}
	}
}

// These are synthetic native-format sources; no capture registration is needed.
func manualSourceFixture(t *testing.T, f captureGenerationFixture, source string) ([]string, string) {
	t.Helper()
	if source == "codex" {
		path := f.write(t, 17)
		return []string{"--codex-home", f.store}, path
	}
	database := setupIsolatedEmptyDB(t)
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO session (id,agent,model,time_created,time_updated) VALUES ('root','root','source-model',1000,2000)`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('m','root',1000,2000,'{"role":"assistant","modelID":"source-model","tokens":{"input":17,"output":2},"time":{"created":1000,"completed":2000}}')`); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, database, f.project)
	return []string{"--db-path", database}, database
}

func TestManualGenerateUnavailableOrInvalidSourcesPreserveAcceptedArtifacts(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		for _, failure := range []string{"missing", "invalid", "directory"} {
			t.Run(source+"/"+failure, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				sourceArgs, sourceFile := manualSourceFixture(t, f, source)
				writeManualStateOnlyConfig(t, f.project)
				summary := filepath.Join(f.project, "SUMMARY.md")
				args := append([]string{"generate", "--source", source, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Source error preservation", "--ticket-type", "fix", "--json-output", f.output, "--append", summary}, sourceArgs...)
				run := func() ([]byte, error) {
					cmd := exec.Command(cliBinaryPath, args...)
					cmd.Dir = f.project
					return cmd.CombinedOutput()
				}
				if out, err := run(); err != nil {
					t.Fatalf("valid source control: %v %s", err, out)
				}
				acceptedLedger, acceptedSummary := readCaptureFile(t, f.output), readCaptureFile(t, summary)
				if err := os.Remove(sourceFile); err != nil {
					t.Fatal(err)
				}
				if failure == "invalid" {
					if err := os.WriteFile(sourceFile, []byte("invalid native source"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if failure == "directory" {
					if err := os.Mkdir(sourceFile, 0700); err != nil {
						t.Fatal(err)
					}
				}
				output, err := run()
				if err == nil {
					t.Fatalf("%s selected source became successful empty telemetry: %s", failure, output)
				}
				if !strings.Contains(strings.ToLower(string(output)), "source") && !strings.Contains(string(output), "rollout") {
					t.Fatalf("source failure lacks actionable diagnostic: %s", output)
				}
				if !bytes.Equal(acceptedLedger, readCaptureFile(t, f.output)) || !bytes.Equal(acceptedSummary, readCaptureFile(t, summary)) {
					t.Fatal("source failure changed accepted publication")
				}
			})
		}
	}
}

func TestManualGenerateLowerAndEmptySnapshotsPreserveOtherOwners(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	source := f.write(t, 51)
	writeManualStateOnlyConfig(t, f.project)
	state := filepath.Join(f.stateBase, "must-remain-absent")
	run := func() {
		t.Helper()
		if out, err := directOwnedGenerate(t, f.project, f.store, state, "root", f.output); err != nil {
			t.Fatalf("manual refresh: %v %s", err, out)
		}
	}
	run()
	baseline := f.ledger(t)
	for _, source := range []domain.SourceHarness{domain.SourceCodex, domain.SourceOpenCode} {
		baseline.Sessions = append(baseline.Sessions, domain.SessionFact{ID: string(source) + "-unavailable", Source: source, User: "other-contributor", Invocations: []domain.InvocationFact{{Model: "stored-model", CostEvidence: domain.CostEvidence{State: domain.CostObserved, KnownSubtotalUSD: domain.CostNumber(0)}, Tokens: domain.InvocationTokens{Input: 37}}}, Skills: []domain.SkillFact{{Name: "preserved"}, {Name: "preserved"}}})
	}
	if err := adapters.SaveFactLedger(f.output, baseline); err != nil {
		t.Fatal(err)
	}
	baseline = f.ledger(t)
	for _, tokens := range []int{9, 0} {
		if tokens == 0 {
			if err := os.WriteFile(source, []byte(fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":\"root\",\"cwd\":%q}}\n", f.project)), 0600); err != nil {
				t.Fatal(err)
			}
		} else {
			f.write(t, tokens)
		}
		freshOutput := filepath.Join(f.project, fmt.Sprintf("fresh-%d", tokens), "agentic_metrics.json")
		restoreScope := selectGenerateControlHome(t, f.project, filepath.Dir(freshOutput))
		out, err := directOwnedGenerate(t, f.project, f.store, state, "root", freshOutput)
		restoreScope()
		if err != nil {
			t.Fatalf("fresh authoritative control: %v %s", err, out)
		}
		freshFixture := f
		freshFixture.output = freshOutput
		freshOwner := freshFixture.ledger(t).Sessions[0]
		for repeat := 0; repeat < 2; repeat++ {
			run()
			got := f.ledger(t)
			if len(got.Sessions) != 3 {
				t.Fatalf("refresh lost/duplicated owners: %+v", got.Sessions)
			}
			for _, owner := range got.Sessions {
				if owner.ID == "root" {
					if !reflect.DeepEqual(owner, freshOwner) {
						t.Fatalf("refresh differs from independently extracted authoritative snapshot: got=%+v want=%+v", owner, freshOwner)
					}
					// Codex intentionally retains an unknown-model structural row
					// for metadata-only sessions. It must replace prior measured
					// facts, rather than unioning their invocations into the result.
					if tokens == 0 && (len(owner.Invocations) != 1 || owner.Invocations[0].Model != "" || owner.Invocations[0].Tokens != (domain.InvocationTokens{}) || owner.Invocations[0].CostEvidence.State != domain.CostUnavailable) {
						t.Fatalf("metadata-only source retained stale measured invocation: %+v", owner)
					}
					if tokens > 0 && (len(owner.Invocations) != 1 || owner.Invocations[0].Tokens.Input != int64(tokens)) {
						t.Fatalf("complete lower snapshot not authoritative: %+v", owner)
					}
				} else {
					var expected domain.SessionFact
					for _, original := range baseline.Sessions {
						if original.ID == owner.ID {
							expected = original
						}
					}
					if !reflect.DeepEqual(owner, expected) {
						t.Fatalf("unavailable owner changed: got=%+v want=%+v", owner, expected)
					}
				}
			}
		}
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatalf("manual refresh allocated registry: %v", err)
	}
}

func TestManualGenerateRejectsInvalidExplicitCaptureBudgets(t *testing.T) {
	for _, flag := range []string{"--max-sessions", "--max-source-bytes"} {
		for _, value := range []string{"0", "-1"} {
			t.Run(flag+"="+value, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				f.write(t, 17)
				if err := os.Remove(filepath.Join(f.project, ".agenticlens-config.local.yaml")); err != nil {
					t.Fatal(err)
				}
				command := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", f.store, "--session-id", "root", "--ticket-id", "TEST-100", "--ticket-title", "Explicit budget validation", "--ticket-type", "fix", "--json-output", f.output, flag, value)
				command.Dir = f.project
				output, err := command.CombinedOutput()
				if err == nil || !strings.Contains(string(output), flag+" must be a positive integer") {
					t.Fatalf("invalid explicit budget not rejected: %v %s", err, output)
				}
				if _, err := os.Stat(f.output); !os.IsNotExist(err) {
					t.Fatalf("invalid budget published output: %v", err)
				}
			})
		}
	}
}

func writeManualStateOnlyConfig(t *testing.T, project string) {
	t.Helper()
	body := fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(project, ".publication-state"))
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}
