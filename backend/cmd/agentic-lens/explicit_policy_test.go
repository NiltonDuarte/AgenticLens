package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestGenerateExplicitEmptyClassificationAndIndependentOverlap(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	enableCLIFixtureSummary(t, f.project)
	f.bind(t, f.write(t, 17))
	for name, body := range map[string]string{"production.go": "package fixture\n", "production_test.go": "package fixture\n"} {
		path := filepath.Join(f.project, "src", name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "src"}, {"commit", "-m", "source and tests"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = f.project
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	cfg, err := adaptersconfig.LoadShared(f.project)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AgenticLens.Version = "v9.8.7"
	write := func() {
		t.Helper()
		body, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cfg.CodeMetrics.Production.Include = []string{}
	cfg.CodeMetrics.Tests.Include = []string{}
	write()
	got := f.generate(t)
	if got.Code.FilesChanged != 0 || got.Code.LinesAdded != 0 || got.Code.TestLinesAdded != 0 || got.Code.ProdLinesAdded != 0 {
		t.Fatalf("explicit empty categories invented classification: %+v", got.Code)
	}
	before, summaryBefore := readCaptureFile(t, f.output), readCaptureFile(t, filepath.Join(f.project, "SUMMARY.md"))
	cfg.CodeMetrics.Production.Include = []string{"src/**"}
	cfg.CodeMetrics.Production.Exclude = []string{}
	cfg.CodeMetrics.Tests.Include = []string{"src/**"}
	cfg.CodeMetrics.Tests.Exclude = []string{}
	write()
	cmd := exec.Command(cliBinaryPath, f.args(t)...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err == nil || !bytes.Contains(bytes.ToLower(out), []byte("both")) {
		t.Fatalf("independently matched category overlap accepted: %v %s", err, out)
	}
	if !bytes.Equal(before, readCaptureFile(t, f.output)) || !bytes.Equal(summaryBefore, readCaptureFile(t, filepath.Join(f.project, "SUMMARY.md"))) {
		t.Fatal("classification failure changed accepted artifacts")
	}
}

func TestServerOptionalSourcePolicyPreservesRecordedLedger(t *testing.T) {
	for _, scenario := range []string{"source-free", "missing-busy-timeout", "explicit-zero-control"} {
		t.Run(scenario, func(t *testing.T) {
			workspaces := setupTestWorkspace(t)
			project := filepath.Dir(workspaces)
			shared := filepath.Join(project, ".agenticlens-config.yaml")
			body := readCaptureFile(t, shared)
			if scenario != "explicit-zero-control" {
				body = bytes.Replace(body, []byte("database:\n  busy_timeout_ms: 0\n"), nil, 1)
			}
			if scenario == "source-free" {
				if err := os.Remove(filepath.Join(project, ".agenticlens-config.local.yaml")); err != nil {
					t.Fatal(err)
				}
				body = bytes.Replace(body, []byte("live:\n  active_threshold_seconds: 91\n  poll_interval_ms: 333\n"), nil, 1)
			} else {
				source := setupIsolatedEmptyDB(t)
				db, err := sql.Open("sqlite", source)
				if err != nil {
					t.Fatal(err)
				}
				_, err = db.Exec(`INSERT INTO session(id,parent_id,agent,model,title,time_created,time_updated) VALUES('ses-orch',NULL,'orchestrator','model','Source fixture',1760000000000,1760000001000),('ses-coder','ses-orch','coder','model','Child',1760000000000,1760000001000);INSERT INTO message(id,session_id,time_created,time_updated,data) VALUES('m1','ses-orch',1760000000000,1760000001000,'{"role":"user","time":{"created":1760000000000}}'),('m2','ses-coder',1760000000000,1760000001000,'{"role":"user","time":{"created":1760000000000}}')`)
				if err != nil {
					t.Fatal(err)
				}
				if err = db.Close(); err != nil {
					t.Fatal(err)
				}
				setOpenCodeFixtureProject(t, source, project)
				local := fmt.Sprintf("telemetry:\n  sources:\n    opencode:\n      database_path: %q\n", source)
				if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(shared, body, 0600); err != nil {
				t.Fatal(err)
			}
			ledgerPath := filepath.Join(workspaces, "TICKET-1", "agentic_metrics.json")
			before := readCaptureFile(t, ledgerPath)
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			port := listener.Addr().(*net.TCPAddr).Port
			if err = listener.Close(); err != nil {
				t.Fatal(err)
			}
			logPath := filepath.Join(t.TempDir(), "server.log")
			log, err := os.Create(logPath)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, "server", "--workspaces", workspaces, "--port", fmt.Sprint(port))
			cmd.Dir = project
			cmd.Stdout = log
			cmd.Stderr = log
			if err = cmd.Start(); err != nil {
				log.Close()
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait(); _ = log.Close() })
			address := fmt.Sprintf("http://127.0.0.1:%d", port)
			response := waitForServerReady(t, address+"/api/workspaces/TICKET-1", 3*time.Second)
			if response.StatusCode != http.StatusOK {
				t.Fatalf("recorded ledger hidden: %d", response.StatusCode)
			}
			response.Body.Close()
			response, err = http.Get(address + "/api/workspaces/TICKET-1/events")
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			var events domain.WorkspaceEventsResponse
			if err = json.NewDecoder(response.Body).Decode(&events); err != nil {
				t.Fatal(err)
			}
			if scenario == "explicit-zero-control" {
				if !events.IsHydrated || events.Availability != domain.TraceComplete || len(events.Coverage.CompleteSessionIDs) != 2 {
					t.Fatalf("valid explicit-zero source policy did not hydrate: %+v; log=%s", events, readCaptureFile(t, logPath))
				}
			} else if len(events.Events) != 0 || events.Availability != domain.TraceUnavailable || events.DataSource != domain.DataSourceUnavailable {
				t.Fatalf("unused/missing source policy invented hydration: %+v", events)
			}
			logBytes := readCaptureFile(t, logPath)
			if scenario == "missing-busy-timeout" && !bytes.Contains(logBytes, []byte("database.busy_timeout_ms")) {
				t.Fatalf("missing busy timeout silently became zero: %s", logBytes)
			}
			if scenario == "source-free" && strings.Contains(string(logBytes), "poll_interval_ms") {
				t.Fatalf("source-free server consumed polling policy: %s", logBytes)
			}
			if !bytes.Equal(before, readCaptureFile(t, ledgerPath)) {
				t.Fatal("optional source handling mutated recorded ledger")
			}
		})
	}
}

func TestGenerateSharedSummaryNameAppendPrecedenceAndAdmission(t *testing.T) {
	f := newCaptureGenerationFixture(t)
	f.output = filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
	f.bind(t, f.write(t, 17))
	cfg, err := adaptersconfig.LoadShared(f.project)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AgenticLens.Version = "v9.8.7"
	cfg.Generation.Summary.Enabled = true
	cfg.Generation.Summary.Filename = "OBSERVATIONS.md"
	write := func() {
		t.Helper()
		body, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(f.project, ".agenticlens-config.yaml"), body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write()
	configured := filepath.Join(filepath.Dir(f.output), "OBSERVATIONS.md")
	f.generate(t)
	if !bytes.Contains(readCaptureFile(t, configured), []byte("### Workflow Efficiency")) {
		t.Fatal("configured summary basename not used beside ledger")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(f.output), "SUMMARY.md")); !os.IsNotExist(err) {
		t.Fatalf("invented summary filename: %v", err)
	}
	configuredBefore := readCaptureFile(t, configured)
	override := filepath.Join(f.project, "manual-summary.md")
	cfg.Generation.Summary.Enabled = false
	write()
	f.writeAt(t, 31, "2026-01-01T00:00:03Z")
	f.generate(t, "--append", override)
	if !bytes.Contains(readCaptureFile(t, override), []byte("### Workflow Efficiency")) || !bytes.Equal(configuredBefore, readCaptureFile(t, configured)) {
		t.Fatal("explicit append did not override disabled configured summary destination")
	}
	diagnostics := captureDiagnostics(t, f.project, f.state)
	before := readCaptureFile(t, f.output)
	cmd := exec.Command(cliBinaryPath, f.args(t, "--append", f.output)...)
	cmd.Dir = f.project
	if out, err := cmd.CombinedOutput(); err == nil || !bytes.Contains(out, []byte("collides")) {
		t.Fatalf("summary collision accepted: %v %s", err, out)
	}
	if !bytes.Equal(before, readCaptureFile(t, f.output)) {
		t.Fatal("summary preflight mutated ledger")
	}
	after := captureDiagnostics(t, f.project, f.state)
	if !reflect.DeepEqual(diagnostics, after) {
		t.Fatal("invalid destination reached publication authorization/journal")
	}
	if _, err := os.Stat(f.output + ".publication-pending"); !os.IsNotExist(err) {
		t.Fatalf("invalid destination created publication journal: %v", err)
	}
}

func TestBackfillFreshGitFailurePreservesPreparedMigrationBytes(t *testing.T) {
	for _, version := range []int{5, 6} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			workspaces := setupTestWorkspace(t)
			project := filepath.Dir(workspaces)
			path := filepath.Join(workspaces, "TICKET-1", "agentic_metrics.json")
			var root map[string]json.RawMessage
			if err := json.Unmarshal(readCaptureFile(t, path), &root); err != nil {
				t.Fatal(err)
			}
			delete(root, "migration")
			root["schema_version"] = json.RawMessage(fmt.Sprint(version))
			before, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, before, 0600); err != nil {
				t.Fatal(err)
			}
			summary := filepath.Join(filepath.Dir(path), "SUMMARY.md")
			summaryBefore := []byte("# Human evidence\nKeep the accepted summary.\n")
			if err = os.WriteFile(summary, summaryBefore, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, "backfill", "--source", "opencode", "--workspaces", workspaces, "--ticket-id", "TICKET-1", "--migrate-schema", "--update-summaries", "--recompute-git", "--code-ref", "missing-selected-ref")
			cmd.Dir = project
			out, runErr := cmd.CombinedOutput()
			if !bytes.Contains(out, []byte("missing-selected-ref")) || !bytes.Contains(out, []byte("requested Git measurement failed before publication")) || !bytes.Contains(out, []byte("Updated: 0")) {
				t.Fatalf("fresh Git failure not visible in batch result: %v %s", runErr, out)
			}
			if !bytes.Equal(before, readCaptureFile(t, path)) || !bytes.Equal(summaryBefore, readCaptureFile(t, summary)) {
				t.Fatal("failed fresh calculation published migration/summary changes")
			}
			if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
				t.Fatalf("failed fresh calculation installed migration backup: %v", err)
			}
		})
	}
}
