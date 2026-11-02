package main_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func runCaptureNative(t *testing.T, cwd, input string, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliBinaryPath, args...)
	cmd.Dir = cwd
	cmd.Stdin = strings.NewReader(input)
	return cmd.CombinedOutput()
}
func writeCapturePolicy(t *testing.T, f captureGenerationFixture, enabled bool, debounce int) {
	t.Helper()
	b := fmt.Sprintf("telemetry:\n  capture:\n    enabled: %t\n    refresh_ms: 100\n    debounce_ms: %d\n    idle_ms: 200\n    max_lifetime_ms: 2000\n", enabled, debounce)
	if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(b), 0600); err != nil {
		t.Fatal(err)
	}
}
func prepareCaptureOwner(t *testing.T, f captureGenerationFixture, id, ticket string, tokens int) string {
	t.Helper()
	path := filepath.Join(f.store, "sessions", id+".jsonl")
	body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":%q,"cwd":%q}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"one","usage":{"input_tokens":%d}}}
`, id, f.project, id, tokens)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: id}
	row, err := s.Register(context.Background(), identity, path, "SessionEnd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Bind(context.Background(), identity, ticket, row.Request); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Change(context.Background(), identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle = "Explicit " + ticket
		r.TicketType = "feat"
		r.Activity = "idle"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestCaptureCopiedBaselineCannotCrossTicketEvenWithCorrectedHeader(t *testing.T) {
	for _, corrected := range []bool{false, true} {
		t.Run(fmt.Sprint(corrected), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 31))
			f.generate(t)
			before := readCaptureFile(t, f.output)
			if corrected {
				var ledger domain.AgenticFactLedger
				if err := json.Unmarshal(before, &ledger); err != nil {
					t.Fatal(err)
				}
				ticket := "TEST-200"
				ledger.TicketID = &ticket
				var err error
				before, err = json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(f.output, before, 0600); err != nil {
					t.Fatal(err)
				}
			}
			out, err := runCaptureNative(t, f.project, "", f.args(t, "--ticket-id", "TEST-200", "--capture-auto", "--allow-empty")...)
			if err == nil || !bytes.Contains(bytes.ToLower(out), []byte("ticket")) {
				t.Errorf("ticket guard: %v %s", err, out)
			}
			if !bytes.Equal(before, readCaptureFile(t, f.output)) {
				t.Error("rejected baseline changed accepted bytes")
			}
		})
	}
}
func TestCaptureStatusBeforeFirstPublicationRequest(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprint(legacy), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 19))
			writeCapturePolicy(t, f, false, 0)
			if legacy {
				db, err := sql.Open("sqlite", filepath.Join(f.state, "registry.db"))
				if err != nil {
					t.Fatal(err)
				}
				if _, err = db.Exec("DROP TABLE IF EXISTS publication_request"); err != nil {
					t.Fatal(err)
				}
				db.Close()
			}
			out, err := runCaptureNative(t, f.project, "", "capture", "status", "--repo", f.project, "--state-dir", f.state)
			if err != nil {
				t.Fatalf("prepublication status: %v %s", err, out)
			}
			var status struct {
				Registrations []domain.CaptureRegistration `json:"registrations"`
				Publication   capture.PublicationRequest   `json:"publication"`
			}
			if err = json.Unmarshal(out, &status); err != nil {
				t.Fatal(err)
			}
			if len(status.Registrations) != 1 || status.Registrations[0].TicketID != "TEST-100" || status.Publication.Generation != 0 {
				t.Fatalf("lost registration or invented publication: %s", out)
			}
		})
	}
}
func TestCaptureGenerateUsesOwningConfigurationRoot(t *testing.T) {
	for _, format := range []string{"yaml", "yml", "json"} {
		t.Run(format, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.write(t, 23)
			old := filepath.Join(f.project, ".agenticlens-config.yaml")
			body := readCaptureFile(t, old)
			if format == "json" {
				body = []byte(`{"project_name":"CaptureFixture","ticket":{"prefixes":["TEST"]},"workspaces":{"dir":"records"},"code_metrics":{"source_roots":["src"]},"anomaly_thresholds":{"context_ceiling_warning_tokens":160000,"context_ceiling_critical_tokens":200000}}`)
			}
			if format != "yaml" {
				if err := os.Remove(old); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config."+format), body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			sub := filepath.Join(f.project, "src")
			if err := os.Mkdir(sub, 0700); err != nil {
				t.Fatal(err)
			}
			args := f.args(t, "--session-id", "root", "--codex-home", f.store)
			out, err := runCaptureNative(t, sub, "", args...)
			if err != nil {
				t.Fatalf("owning root from subdir: %v %s", err, out)
			}
			assertCaptureRootInput(t, f.ledger(t), 23)
			before := readCaptureFile(t, f.output)
			if err = os.WriteFile(filepath.Join(sub, ".agenticlens-config."+format), body, 0600); err != nil {
				t.Fatal(err)
			}
			out, err = runCaptureNative(t, sub, "", args...)
			if err == nil || !bytes.Contains(out, []byte("project")) {
				t.Errorf("nested configured project imported parent source: %v %s", err, out)
			}
			if !bytes.Equal(before, readCaptureFile(t, f.output)) {
				t.Error("nested rejection changed ledger")
			}
		})
	}
}
func TestCaptureRejectsOutsideRecoveryJournalBeforeAnyMutation(t *testing.T) {
	for _, present := range []bool{true, false} {
		t.Run(fmt.Sprint(present), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 19))
			f.generate(t)
			before := readCaptureFile(t, f.output)
			victim := filepath.Join(t.TempDir(), "victim")
			accepted := []byte("outside accepted bytes")
			if err := os.WriteFile(victim, accepted, 0600); err != nil {
				t.Fatal(err)
			}
			if err := adapters.BeginPublication(f.output, []string{f.output, filepath.Join(f.project, "capture_report.json")}); err != nil {
				t.Fatal(err)
			}
			journal := f.output + ".publication-pending"
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(readCaptureFile(t, journal), &raw); err != nil {
				t.Fatal(err)
			}
			type member struct {
				Path    string `json:"path"`
				Present bool   `json:"present"`
				Data    []byte `json:"data"`
			}
			var members []member
			if err := json.Unmarshal(raw["artifacts"], &members); err != nil {
				t.Fatal(err)
			}
			members[0].Data = []byte("must not restore authorized first member")
			members[1] = member{victim, present, []byte("attacker replacement")}
			var err error
			raw["artifacts"], err = json.Marshal(members)
			if err != nil {
				t.Fatal(err)
			}
			body, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(journal, body, 0600); err != nil {
				t.Fatal(err)
			}
			out, err := runCaptureNative(t, f.project, "", f.args(t)...)
			if err == nil || !(bytes.Contains(bytes.ToLower(out), []byte("authoriz")) || bytes.Contains(bytes.ToLower(out), []byte("destination"))) {
				t.Errorf("journal guard: %v %s", err, out)
			}
			if !bytes.Equal(before, readCaptureFile(t, f.output)) {
				t.Error("partially restored ledger before validating members")
			}
			if !bytes.Equal(accepted, readCaptureFile(t, victim)) {
				t.Error("mutated outside victim")
			}
			if !bytes.Equal(body, readCaptureFile(t, journal)) {
				t.Error("rejected journal not preserved")
			}
		})
	}
}
func TestCaptureRejectsDanglingOutsideSummaryBeforePublicationOrRecovery(t *testing.T) {
	for _, withJournal := range []bool{false, true} {
		t.Run(fmt.Sprint(withJournal), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 19))
			f.generate(t)
			before := readCaptureFile(t, f.output)
			report := filepath.Join(f.project, "capture_report.json")
			beforeReport := readCaptureFile(t, report)
			victim := filepath.Join(t.TempDir(), "absent-summary.md")
			summary := filepath.Join(f.project, "SUMMARY.md")
			if err := os.Symlink(victim, summary); err != nil {
				t.Fatal(err)
			}
			journal := f.output + ".publication-pending"
			var body []byte
			if withJournal {
				// Complete, ledger-bound metadata poisons only the destination
				// accepted through the dangling caller-supplied summary symlink.
				type member struct {
					Path    string `json:"path"`
					Present bool   `json:"present"`
					Data    []byte `json:"data"`
				}
				var err error
				body, err = json.Marshal(struct {
					Version     int      `json:"version"`
					Ledger      string   `json:"ledger"`
					MemberCount int      `json:"member_count"`
					Artifacts   []member `json:"artifacts"`
				}{1, f.output, 3, []member{{f.output, true, before}, {report, true, beforeReport}, {victim, true, []byte("untrusted journal payload")}}})
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(journal, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			out, err := runCaptureNative(t, f.project, "", f.args(t, "--append", summary)...)
			if err == nil || !bytes.Contains(out, []byte("escapes project")) {
				t.Errorf("dangling destination accepted: %v %s", err, out)
			}
			if _, err := os.Lstat(victim); !os.IsNotExist(err) {
				t.Errorf("outside victim created: %v", err)
			}
			if !bytes.Equal(before, readCaptureFile(t, f.output)) || !bytes.Equal(beforeReport, readCaptureFile(t, report)) {
				t.Error("rejected destination mutated accepted artifacts")
			}
			if withJournal && !bytes.Equal(body, readCaptureFile(t, journal)) {
				t.Error("rejected journal changed")
			}
			if !withJournal {
				if _, err := os.Lstat(journal); !os.IsNotExist(err) {
					t.Errorf("rejected destination created journal: %v", err)
				}
			}
		})
	}
}

func TestCaptureRejectsSymlinkDotDotOutsideSummary(t *testing.T) {
	for _, existing := range []bool{false, true} {
		for _, indirect := range []bool{false, true} {
			t.Run(fmt.Sprintf("existing=%t/indirect=%t", existing, indirect), func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				f.bind(t, f.write(t, 19))
				f.generate(t)
				before := readCaptureFile(t, f.output)
				report := filepath.Join(f.project, "capture_report.json")
				beforeReport := readCaptureFile(t, report)
				outside := t.TempDir()
				child := filepath.Join(outside, "child")
				if err := os.Mkdir(child, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(child, filepath.Join(f.project, "hop")); err != nil {
					t.Fatal(err)
				}
				victim := filepath.Join(outside, "summary.md")
				sentinel := []byte("outside accepted bytes")
				if existing {
					if err := os.WriteFile(victim, sentinel, 0600); err != nil {
						t.Fatal(err)
					}
				}
				// Preserve .. until the kernel follows hop; filepath.Join would
				// erase the trigger before it reaches the CLI.
				appendPath := f.project + string(filepath.Separator) + "hop/../summary.md"
				if indirect {
					appendPath = filepath.Join(f.project, "SUMMARY.md")
					if err := os.Symlink("hop/../summary.md", appendPath); err != nil {
						t.Fatal(err)
					}
				}
				out, err := runCaptureNative(t, f.project, "", f.args(t, "--append", appendPath)...)
				if err == nil || !bytes.Contains(out, []byte("escapes project")) {
					t.Errorf("physical traversal accepted: %v %s", err, out)
				}
				if existing {
					if !bytes.Equal(sentinel, readCaptureFile(t, victim)) {
						t.Error("outside victim modified")
					}
				} else if _, err := os.Lstat(victim); !os.IsNotExist(err) {
					t.Errorf("outside victim created: %v", err)
				}
				if !bytes.Equal(before, readCaptureFile(t, f.output)) || !bytes.Equal(beforeReport, readCaptureFile(t, report)) {
					t.Error("rejected traversal mutated accepted artifacts")
				}
				if _, err := os.Lstat(f.output + ".publication-pending"); !os.IsNotExist(err) {
					t.Errorf("rejected traversal created journal: %v", err)
				}
			})
		}
	}
}

func TestCapturePublishesThroughPhysicalArtifactAliases(t *testing.T) {
	for _, kind := range []string{"ledger", "report", "summary", "all"} {
		for _, state := range []string{"missing", "existing", "missing-parent"} {
			t.Run(kind+"/"+state, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				f.bind(t, f.write(t, 19))
				summary := filepath.Join(f.project, "SUMMARY.md")
				report := filepath.Join(f.project, "capture_report.json")
				if state == "existing" {
					f.generate(t, "--append", summary)
				}
				aliases := map[string]string{}
				for name, path := range map[string]string{"ledger": f.output, "report": report, "summary": summary} {
					if name != kind && kind != "all" {
						continue
					}
					parent := f.project
					if state == "missing-parent" {
						parent = filepath.Join(parent, "physical-"+name, "nested")
					}
					target := filepath.Join(parent, "physical-"+filepath.Base(path))
					if state == "existing" {
						if err := os.Rename(path, target); err != nil {
							t.Fatal(err)
						}
					}
					if err := os.Symlink(target, path); err != nil {
						t.Fatal(err)
					}
					aliases[path] = target
				}
				for _, tokens := range []int{19, 29} {
					checkpoint := "2026-01-01T00:00:02Z"
					if tokens == 29 {
						checkpoint = "2026-01-01T00:00:03Z"
					}
					f.writeAt(t, tokens, checkpoint)
					assertCaptureRootInput(t, f.generate(t, "--append", summary), int64(tokens))
					for alias, target := range aliases {
						info, err := os.Lstat(alias)
						if err != nil || info.Mode()&os.ModeSymlink == 0 {
							t.Fatalf("artifact alias replaced: %s %v", alias, err)
						}
						if len(readCaptureFile(t, target)) == 0 {
							t.Errorf("physical artifact absent: %s", target)
						}
					}
					ledgerTarget := f.output
					if target, ok := aliases[f.output]; ok {
						ledgerTarget = target
					}
					for _, path := range []string{f.output, ledgerTarget} {
						if _, err := os.Lstat(path + ".publication-pending"); !os.IsNotExist(err) {
							t.Errorf("orphaned journal for %s: %v", path, err)
						}
					}
					digest := fmt.Sprintf("%x", sha256.Sum256(readCaptureFile(t, f.output)))
					if !bytes.Contains(readCaptureFile(t, report), []byte(digest)) {
						t.Error("report does not match physical ledger")
					}
					if !bytes.Contains(readCaptureFile(t, summary), []byte("### Workflow Efficiency")) {
						t.Error("physical summary not updated")
					}
				}
			})
		}
	}
}

func TestCaptureRejectsMissingLeafOutputAppendAlias(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			f.bind(t, f.write(t, 19))
			alias := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(f.project, alias); err != nil {
				t.Fatal(err)
			}
			var before []byte
			if existing {
				f.generate(t)
				before = readCaptureFile(t, f.output)
			}
			out, err := runCaptureNative(t, f.project, "", f.args(t, "--append", filepath.Join(alias, filepath.Base(f.output)))...)
			if err == nil || !bytes.Contains(bytes.ToLower(out), []byte("overlap")) {
				t.Errorf("alias guard: %v %s", err, out)
			}
			if existing {
				if !bytes.Equal(before, readCaptureFile(t, f.output)) {
					t.Error("alias rejection changed bytes")
				}
			} else if _, err = os.Stat(f.output); !os.IsNotExist(err) {
				t.Errorf("missing leaf alias wrote output: %v", err)
			}
		})
	}
}

func startCorrectionWorker(t *testing.T, f captureGenerationFixture) {
	t.Helper()
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.RequestPublication(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	cmd := exec.CommandContext(ctx, cliBinaryPath, "capture", "worker", "--repo", f.project, "--state-dir", f.state)
	cmd.Dir = f.project
	log, err := os.CreateTemp(t.TempDir(), "worker")
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = log
	cmd.Stderr = log
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	t.Cleanup(func() {
		writeCapturePolicy(t, f, false, 0)
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("worker: %v %s", err, readCaptureFile(t, log.Name()))
			}
		case <-time.After(4 * time.Second):
			cancel()
			<-done
			t.Error("worker did not stop after disabling")
		}
		cancel()
		log.Close()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			lock, err := capture.AcquireWorker(f.state)
			if err != nil {
				t.Error(err)
				return
			}
			if lock != nil {
				lock.Close()
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Error("automatic worker chain retained ownership after disable")
	})
}
func waitCorrectionLedger(t *testing.T, path, id string, tokens int64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		body, err := os.ReadFile(path)
		var l domain.AgenticFactLedger
		if err == nil && json.Unmarshal(body, &l) == nil && len(l.Sessions) == 1 && l.Sessions[0].ID == id && len(l.Sessions[0].Invocations) == 1 && l.Sessions[0].Invocations[0].Tokens.Input == tokens {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no source-backed ledger for %s input=%d at %s", id, tokens, path)
}
func TestCaptureContinuousNativeCallbacksDoNotStarveAllTickets(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	root := prepareCaptureOwner(t, f, "root", "TEST-100", 10)
	prepareCaptureOwner(t, f, "second", "TEST-200", 20)
	// 100ms coalescing with 100ms refresh: supported events remain closer together
	// than debounce, yet the periodic refresh must publish before the stream ends.
	writeCapturePolicy(t, f, true, 100)
	startCorrectionWorker(t, f)
	first := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	second := filepath.Join(f.project, "records", "TEST-200", "agentic_metrics.json")
	deadline := time.Now().Add(1800 * time.Millisecond)
	callbacks := 0
	published := false
	for time.Now().Before(deadline) {
		payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "UserPromptSubmit", TranscriptPath: root})
		if err != nil {
			t.Fatal(err)
		}
		out, err := runCaptureNative(t, f.project, string(payload), "capture", "receive", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state)
		if err != nil {
			t.Fatalf("callback: %v %s", err, out)
		}
		callbacks++
		if _, e := os.Stat(first); e == nil {
			if _, e = os.Stat(second); e == nil {
				published = true
				break
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
	if !published {
		t.Fatalf("continuous stream (%d callbacks) starved periodic publication of both tickets", callbacks)
	}
	waitCorrectionLedger(t, first, "root", 10)
	waitCorrectionLedger(t, second, "second", 20)
}
func TestCaptureLiveDisableStopsWritesAndManualRecoverySurvives(t *testing.T) {
	for _, mode := range []string{"yaml", "disable", "uninstall", "dry-run"} {
		t.Run(mode, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			prepareCaptureOwner(t, f, "root", "TEST-100", 10)
			writeCapturePolicy(t, f, false, 0)
			if out, err := runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--binary", cliBinaryPath); err != nil {
				t.Fatalf("install: %v %s", err, out)
			}
			writeCapturePolicy(t, f, true, 0)
			startCorrectionWorker(t, f)
			output := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
			waitCorrectionLedger(t, output, "root", 10)
			if mode == "yaml" {
				writeCapturePolicy(t, f, false, 0)
			} else {
				args := []string{"capture", mode, "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state}
				if mode == "dry-run" {
					args[1] = "disable"
					args = append(args, "--dry-run")
				}
				if out, err := runCaptureNative(t, f.project, "", args...); err != nil {
					t.Fatalf("%s: %v %s", mode, err, out)
				}
			}
			if mode == "dry-run" {
				f.writeAt(t, 30, "2026-01-01T00:00:03Z")
				waitCorrectionLedger(t, output, "root", 30)
				return
			}
			// Allow the already captured publication to finish, then require stable bytes
			// over multiple refresh intervals and beyond the finite owner's lifetime.
			time.Sleep(300 * time.Millisecond)
			before := readCaptureFile(t, output)
			f.writeAt(t, 30, "2026-01-01T00:00:03Z")
			time.Sleep(2100 * time.Millisecond)
			if !bytes.Equal(before, readCaptureFile(t, output)) {
				t.Error("disabled capture continued writing or renewing")
			}
			lock, err := capture.AcquireWorker(f.state)
			if err != nil || lock == nil {
				t.Fatalf("disabled worker retained ownership: %v", err)
			}
			lock.Close()
			// Explicit manual recovery remains supported with the accepted source binding.
			f.output = output
			ledger := f.generate(t)
			assertCaptureRootInput(t, ledger, 30)
		})
	}
}
func TestCaptureInstallerUsesSelectedProjectSourceAndTrueAliases(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	caller := newCaptureGenerationFixture(t)
	for _, fixture := range []captureGenerationFixture{f, caller} {
		cfg := filepath.Join(fixture.project, ".agenticlens-config.local.yaml")
		b := []byte(fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n  sources:\n    codex:\n      home: %q\n", fixture.store))
		if err := os.WriteFile(cfg, b, 0600); err != nil {
			t.Fatal(err)
		}

	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(f.project, alias); err != nil {
		t.Fatal(err)
	}
	for _, repo := range []string{f.project, alias} {
		out, err := runCaptureNative(t, caller.project, "", "capture", "install", "--repo", repo, "--source", "codex", "--state-dir", f.state, "--binary", cliBinaryPath)
		if err != nil {
			t.Fatalf("install selected project/alias: %v %s", err, out)
		}
		hooks := readCaptureFile(t, filepath.Join(f.project, ".codex", "hooks.json"))
		if !bytes.Contains(hooks, []byte(f.store)) || bytes.Contains(hooks, []byte(caller.store)) {
			t.Fatalf("source fallback used caller instead of --repo: %s", hooks)
		}
	}
	out, err := runCaptureNative(t, caller.project, "", "capture", "uninstall", "--repo", alias, "--source", "codex", "--state-dir", f.state)
	if err != nil {
		t.Fatalf("remove physical alias: %v %s", err, out)
	}
	if hooks, err := os.ReadFile(filepath.Join(f.project, ".codex", "hooks.json")); err == nil && bytes.Contains(hooks, []byte("AgenticLens capture")) {
		t.Fatalf("alias removal left owned hooks: %s", hooks)
	}
}
func TestCaptureSuppliedTranscriptRequiresSelectedHomeIncludingQueuedChildren(t *testing.T) {
	for _, relation := range []string{"foreign", "foreign symlink", "archive", "home alias"} {
		t.Run(relation, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 23)
			selected := f.store
			if relation == "foreign" || relation == "foreign symlink" {
				selected = t.TempDir()
				if relation == "foreign symlink" {
					dir := filepath.Join(selected, "sessions")
					if err := os.Mkdir(dir, 0700); err != nil {
						t.Fatal(err)
					}
					link := filepath.Join(dir, "root.jsonl")
					if err := os.Symlink(path, link); err != nil {
						t.Fatal(err)
					}
					path = link
				}
			}
			if relation == "archive" {
				dir := filepath.Join(f.store, "archived_sessions")
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				newPath := filepath.Join(dir, "root.jsonl")
				if err := os.Rename(path, newPath); err != nil {
					t.Fatal(err)
				}
				path = newPath
			}
			if relation == "home alias" {
				alias := filepath.Join(t.TempDir(), "home")
				if err := os.Symlink(f.store, alias); err != nil {
					t.Fatal(err)
				}
				selected = alias
			}
			writeCapturePolicy(t, f, false, 0)
			payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
			if err != nil {
				t.Fatal(err)
			}
			out, err := runCaptureNative(t, f.project, string(payload), "capture", "receive", "--repo", f.project, "--source", "codex", "--store", selected, "--state-dir", f.state)
			if err == nil {
				reqOut, reqErr := runCaptureNative(t, f.project, "", "capture", "request", "--repo", f.project, "--source", "codex", "--store", selected, "--state-dir", f.state, "--session-id", "root")
				if reqErr != nil {
					t.Fatalf("request fixture: %v %s", reqErr, reqOut)
				}
				var row domain.CaptureRegistration
				if e := json.Unmarshal(reqOut, &row); e != nil {
					t.Fatal(e)
				}
				out, err = runCaptureNative(t, f.project, "", "capture", "bind", "--repo", f.project, "--source", "codex", "--store", selected, "--state-dir", f.state, "--session-id", "root", "--request", row.Request, "--ticket-id", "TEST-100", "--ticket-title", "Store proof", "--ticket-type", "feat")
			}
			foreign := strings.HasPrefix(relation, "foreign")
			if foreign {
				if err == nil || !(bytes.Contains(bytes.ToLower(out), []byte("store")) || bytes.Contains(bytes.ToLower(out), []byte("home"))) {
					t.Fatalf("wrong-home guard: %v %s", err, out)
				}
				return
			}
			if err != nil {
				t.Fatalf("legitimate archive/alias: %v %s", err, out)
			}
		})
	}
}

func TestCaptureRejectsRawStoreTraversalAndPreservesAcceptedOwner(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	root := f.write(t, 31)
	f.bind(t, root)
	f.generate(t)
	acceptedLedger := readCaptureFile(t, f.output)
	report := filepath.Join(f.project, "capture_report.json")
	acceptedReport := readCaptureFile(t, report)
	writeCapturePolicy(t, f, false, 0)
	outside := t.TempDir()
	if err := os.Mkdir(filepath.Join(outside, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(outside, "root.jsonl")
	body := bytes.ReplaceAll(readCaptureFile(t, root), []byte(`"input_tokens":31`), []byte(`"input_tokens":73`))
	body = bytes.ReplaceAll(body, []byte("00:00:02Z"), []byte("00:00:03Z"))
	if err := os.WriteFile(foreign, body, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "child"), filepath.Join(f.store, "sessions", "hop")); err != nil {
		t.Fatal(err)
	}
	raw := f.store + "/sessions/hop/../root.jsonl"
	for _, path := range []string{foreign, raw} {
		payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: path})
		if err != nil {
			t.Fatal(err)
		}
		out, err := runCaptureNative(t, f.project, string(payload), "capture", "receive", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state)
		if err == nil || !bytes.Contains(out, []byte("source store")) {
			t.Errorf("foreign source accepted via %s: %v %s", path, err, out)
		}
	}
	// Model a previously persisted row reaching reconciliation without a fresh
	// receive guard. Explicit binding itself does not assert source validity.
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
	row, err := s.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TranscriptPath = raw
		return r, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := runCaptureNative(t, f.project, "", "capture", "bind", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state, "--session-id", "root", "--ticket-id", "TEST-100", "--request", row.Request)
	if err != nil {
		t.Fatalf("explicit binding: %v %s", err, out)
	}
	out, err = runCaptureNative(t, f.project, "", "capture", "reconcile", "--repo", f.project, "--ticket-id", "TEST-100", "--state-dir", f.state)
	if err != nil || !bytes.Contains(out, []byte(`"state":"failed"`)) || !bytes.Contains(out, []byte("previous facts preserved")) {
		t.Errorf("rejected refresh not visible in reconciliation: %v %s", err, out)
	}
	out, err = runCaptureNative(t, f.project, "", f.args(t)...)
	if err == nil || !bytes.Contains(out, []byte("source store")) {
		t.Errorf("baseline accepted outside-store association: %v %s", err, out)
	}
	assertCaptureRootInput(t, f.ledger(t), 31)
	if !bytes.Equal(acceptedLedger, readCaptureFile(t, f.output)) || !bytes.Equal(acceptedReport, readCaptureFile(t, report)) {
		t.Error("rejected source changed accepted publication")
	}
}

func TestCaptureInstalledRawSourceAliasesMatchNativeRead(t *testing.T) {
	for _, spelling := range []string{"absolute", "relative", "home"} {
		for _, explicit := range []bool{false, true} {
			if spelling == "home" && explicit {
				continue
			}
			t.Run(fmt.Sprintf("%s/explicit=%t", spelling, explicit), func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				root := f.write(t, 31)
				child := filepath.Join(f.store, "child")
				if err := os.Mkdir(child, 0700); err != nil {
					t.Fatal(err)
				}
				aliasParent := t.TempDir()
				raw := aliasParent + "/hop/.."
				if spelling == "relative" {
					aliasParent = filepath.Join(f.project, "source-alias")
					raw = "source-alias/hop/.."
				}
				if spelling == "home" {
					home := t.TempDir()
					t.Setenv("HOME", home)
					aliasParent = filepath.Join(home, "source-alias")
					raw = "~/source-alias/hop/.."
				}
				if err := os.MkdirAll(aliasParent, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(child, filepath.Join(aliasParent, "hop")); err != nil {
					t.Fatal(err)
				}
				cfg := fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n  sources:\n    codex:\n      home: %q\n", f.state, raw)
				if err := os.WriteFile(filepath.Join(f.project, ".agenticlens-config.local.yaml"), []byte(cfg), 0600); err != nil {
					t.Fatal(err)
				}
				payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SessionStart", TranscriptPath: root})
				if err != nil {
					t.Fatal(err)
				}
				install := []string{"capture", "install", "--repo", f.project, "--source", "codex", "--state-dir", f.state, "--binary", cliBinaryPath}
				if explicit {
					install = append(install, "--store", raw)
				}
				for attempt := 0; attempt < 2; attempt++ {
					out, err := runCaptureNative(t, f.project, "", install...)
					if err != nil {
						t.Fatalf("install/reinstall: %v %s", err, out)
					}
					var hooks struct {
						Hooks map[string][]struct {
							Hooks []struct {
								Command string `json:"command"`
							} `json:"hooks"`
						} `json:"hooks"`
					}
					if err = json.Unmarshal(readCaptureFile(t, filepath.Join(f.project, ".codex", "hooks.json")), &hooks); err != nil {
						t.Fatal(err)
					}
					command := hooks.Hooks["SessionStart"][0].Hooks[0].Command
					cmd := exec.Command("sh", "-c", command)
					cmd.Dir = f.project
					cmd.Stdin = bytes.NewReader(payload)
					out, err = cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("installed hook disagrees with physical source: %v %s", err, out)
					}
					assertCaptureRootInput(t, f.generate(t, "--session-id", "root"), 31)
				}
			})
		}
	}
}

func TestCaptureInstalledStateDriftIsDiagnosedAndRecoveryPreservesOldRegistry(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	local := filepath.Join(f.project, ".agenticlens-config.local.yaml")
	set := func(name string) {
		t.Helper()
		if err := os.WriteFile(local, []byte("telemetry:\n  capture:\n    enabled: false\n    state_dir: "+name+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	set("old-state")
	old, err := capture.ProjectState(filepath.Join(f.project, "old-state"), f.project)
	if err != nil {
		t.Fatal(err)
	}
	f.state = old
	prepareCaptureOwner(t, f, "root", "TEST-100", 19)
	out, err := runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath)
	if err != nil {
		t.Fatalf("initial install: %v %s", err, out)
	}
	before := readCaptureFile(t, filepath.Join(old, "registry.db"))
	hooks := readCaptureFile(t, filepath.Join(f.project, ".codex", "hooks.json"))
	set("new-state")
	out, err = runCaptureNative(t, f.project, "", "capture", "status", "--repo", f.project)
	if err == nil || !bytes.Contains(out, []byte("installed")) || !bytes.Contains(out, []byte("state")) {
		t.Errorf("default status concealed drift: %v %s", err, out)
	}
	out, err = runCaptureNative(t, f.project, "", "capture", "status", "--repo", f.project, "--state-dir", old)
	if err != nil || !bytes.Contains(out, []byte("TEST-100")) || !bytes.Contains(bytes.ToLower(out), []byte("drift")) {
		t.Errorf("old-state explicit recovery failed/undiagnosed: %v %s", err, out)
	}
	out, err = runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath)
	if err == nil {
		t.Error("implicit reinstall silently abandoned old registry")
	}
	if !bytes.Equal(hooks, readCaptureFile(t, filepath.Join(f.project, ".codex", "hooks.json"))) {
		t.Error("rejected reinstall mutated hooks")
	}
	// Restore configured state then reinstall the same owned callbacks safely.
	set("old-state")
	out, err = runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", "codex", "--store", f.store, "--binary", cliBinaryPath)
	if err != nil {
		t.Fatalf("safe reinstall: %v %s", err, out)
	}
	if !bytes.Equal(before, readCaptureFile(t, filepath.Join(old, "registry.db"))) {
		t.Error("status/reinstall changed source registry bytes")
	}
}

func TestCaptureQueuedForeignChildCannotImportFromAnotherHome(t *testing.T) {
	for _, rawTraversal := range []bool{false, true} {
		t.Run(fmt.Sprint(rawTraversal), func(t *testing.T) {

			f := newCaptureGenerationFixture(t)
			root := prepareCaptureOwner(t, f, "root", "TEST-100", 23)
			writeCapturePolicy(t, f, false, 0)
			foreign := t.TempDir()
			child := filepath.Join(foreign, "child.jsonl")
			body := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"child","cwd":%q,"source":{"subagent":{"thread_spawn":{"parent_thread_id":"root","depth":1}}}}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"source-model","turn_id":"child-turn"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":"child","turn_id":"child-turn","usage":{"input_tokens":70}}}
`, f.project)
			if err := os.WriteFile(child, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if rawTraversal {
				nested := filepath.Join(foreign, "nested")
				if err := os.Mkdir(nested, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(nested, filepath.Join(f.store, "sessions", "hop")); err != nil {
					t.Fatal(err)
				}
				child = f.store + "/sessions/hop/../child.jsonl"
			}
			payload, err := json.Marshal(capture.HookInput{SessionID: "root", Event: "SubagentStop", TranscriptPath: root, AgentTranscriptPath: child})
			if err != nil {
				t.Fatal(err)
			}
			out, err := runCaptureNative(t, f.project, string(payload), "capture", "receive", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state)
			if err != nil && !bytes.Contains(bytes.ToLower(out), []byte("store")) {
				t.Fatalf("unexpected callback error: %v %s", err, out)
			}
			ledger := f.generate(t)
			assertCaptureRootInput(t, ledger, 23)
			for _, row := range ledger.Sessions {
				if row.ID == "child" {
					t.Fatalf("foreign queued child imported: %+v", row)
				}
			}

		})
	}
}

func TestCaptureDisablingOneIntegrationKeepsOtherSourceAutomatic(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	writeCapturePolicy(t, f, false, 0)
	dbPath := setupIsolatedEmptyDB(t)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`INSERT INTO session (id,agent,model,time_created,time_updated) VALUES ('enabled-root','root','source-model',1767225600000,1767225602000)`); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, dbPath, f.project)
	if _, err = db.Exec(`INSERT INTO message (id,session_id,time_created,time_updated,data) VALUES ('m1','enabled-root',1767225601000,1767225602000,'{"role":"assistant","modelID":"source-model","time":{"created":1767225601000,"completed":1767225602000},"tokens":{"input":43,"output":2}}')`); err != nil {
		t.Fatal(err)
	}
	s, err := capture.OpenProject(f.state, f.project)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: dbPath, SessionID: "enabled-root"}
	row, err := s.Register(context.Background(), id, "", "session.idle")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Bind(context.Background(), id, "TEST-200", row.Request); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		r.TicketTitle = "Other enabled source"
		r.TicketType = "feat"
		r.Activity = "idle"
		return r, nil
	}); err != nil {
		t.Fatal(err)
	}
	s.Close()
	for _, source := range []string{"codex", "opencode"} {
		store := f.store
		if source == "opencode" {
			store = dbPath
		}
		out, err := runCaptureNative(t, f.project, "", "capture", "install", "--repo", f.project, "--source", source, "--store", store, "--state-dir", f.state, "--binary", cliBinaryPath)
		if err != nil {
			t.Fatalf("install %s: %v %s", source, err, out)
		}
	}
	out, err := runCaptureNative(t, f.project, "", "capture", "disable", "--repo", f.project, "--source", "codex", "--store", f.store, "--state-dir", f.state)
	if err != nil {
		t.Fatalf("disable: %v %s", err, out)
	}
	writeCapturePolicy(t, f, true, 0)
	startCorrectionWorker(t, f)
	waitCorrectionLedger(t, filepath.Join(f.project, "records", "TEST-200", "agentic_metrics.json"), "enabled-root", 43)
}
