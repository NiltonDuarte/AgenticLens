package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	_ "modernc.org/sqlite"
)

var cliBinaryPath string

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "agentic-lens-test-bin-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	binPath := filepath.Join(tempDir, "agentic-lens")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out))
	}
	cliBinaryPath = binPath

	os.Exit(m.Run())
}

func sampleFactJSON(ticketID, title string, duration int64, cost float64) string {
	return `{
		"ticket_id": "` + ticketID + `",
		"ticket_title": "` + title + `",
		"ticket_type": "feature",
		"started_at": "2026-10-17T10:00:00Z",
		"completed_at": "2026-10-17T11:00:00Z",
		"updated_at": "2026-10-17T11:00:00Z",
		"code": {
			"lines_added": 120,
			"lines_deleted": 20,
			"test_lines_added": 60,
			"test_lines_deleted": 0,
			"prod_lines_added": 60,
			"prod_lines_deleted": 20,
			"files_changed": 4,
			"sdlc_iterations": 2
		},
		"invocations": [
			{
				"id": "ses-orch",
				"role": "orchestrator",
				"model": "claude-3-sonnet",
				"user": "test-engineer",
				"active_duration_seconds": ` + fmt.Sprintf("%d", duration) + `,
				"turns": 6,
				"tokens": {
					"input": 5000,
					"output": 1000,
					"cache_read": 10000,
					"reasoning": 300,
					"peak_context": 80000
				},
				"cost_usd": ` + "0.25" + `,
				"compactions": 0
			},
			{
				"id": "ses-coder",
				"parent_id": "ses-orch",
				"role": "coder",
				"model": "claude-3-opus",
				"user": "test-engineer",
				"active_duration_seconds": 60,
				"turns": 4,
				"tokens": {
					"input": 8000,
					"output": 2000,
					"cache_read": 12000,
					"reasoning": 500,
					"peak_context": 150000
				},
				"cost_usd": ` + "0.75" + `,
				"compactions": 0
			}
		],
		"tools": [
			{
				"name": "bash",
				"calls": 10,
				"errors": 1
			}
		]
	}`
}

func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed to create workspaces dir: %v", err)
	}

	t1Dir := filepath.Join(wsDir, "TICKET-1")
	if err := os.MkdirAll(t1Dir, 0755); err != nil {
		t.Fatalf("failed to create TICKET-1 dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(t1Dir, "agentic_metrics.json"), []byte(sampleFactJSON("TICKET-1", "First Feature", 120, 1.00)), 0644); err != nil {
		t.Fatalf("failed to write TICKET-1 ledger: %v", err)
	}

	t2Dir := filepath.Join(wsDir, "TICKET-2")
	if err := os.MkdirAll(t2Dir, 0755); err != nil {
		t.Fatalf("failed to create TICKET-2 dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(t2Dir, "agentic_metrics.json"), []byte(sampleFactJSON("TICKET-2", "Second Feature", 240, 2.50)), 0644); err != nil {
		t.Fatalf("failed to write TICKET-2 ledger: %v", err)
	}

	return wsDir
}

func TestCLIListCommand(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("CLI list failed: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Found 2 ticket workspaces") {
		t.Errorf("expected 2 ticket workspaces in output, got: %s", out)
	}
	if !strings.Contains(out, "TICKET-1") || !strings.Contains(out, "TICKET-2") {
		t.Errorf("expected TICKET-1 and TICKET-2 in list, got: %s", out)
	}
	if !strings.Contains(out, "First Feature") || !strings.Contains(out, "Second Feature") {
		t.Errorf("expected titles in list, got: %s", out)
	}
}

func TestCLITicketFilterAndMarkdownOutput(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	t.Run("specific ticket id outputs markdown", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-1")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "### Workflow Efficiency") {
			t.Errorf("expected Workflow Efficiency section, got: %s", out)
		}
		if !strings.Contains(out, "### Agentic Execution Metrics") {
			t.Errorf("expected Agentic Execution Metrics section, got: %s", out)
		}
		if !strings.Contains(out, "| **orchestrator** |") {
			t.Errorf("expected orchestrator in roles table, got: %s", out)
		}
		if !strings.Contains(out, "| **coder** |") {
			t.Errorf("expected coder in roles table, got: %s", out)
		}
	})

	t.Run("non-existent ticket id returns exit code 1", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "DOES-NOT-EXIST")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected error for non-existent ticket, got success")
		}
		if !strings.Contains(stderr.String(), "Error: Ticket 'DOES-NOT-EXIST' not found") {
			t.Errorf("expected ticket not found error, got: %s", stderr.String())
		}
	})

	t.Run("server subcommand starts and responds", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "server", "--port", "19090", "--workspaces", wsDir)
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand: %v", err)
		}
		defer func() {
			if killErr := cmd.Process.Kill(); killErr != nil {
				// Process might have already terminated
			}
			if waitErr := cmd.Wait(); waitErr != nil {
				// Process termination error expected after kill
			}
		}()

		// Wait briefly for server startup
		var resp *http.Response
		var err error
		for i := 0; i < 20; i++ {
			time.Sleep(100 * time.Millisecond)
			resp, err = http.Get("http://localhost:19090/api/fleet")
			if err == nil {
				break
			}
		}
		if err != nil {
			t.Fatalf("failed querying server: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from /api/fleet, got %d", resp.StatusCode)
		}
	})
}

func TestCLIJsonOutput(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	outJSONPath := filepath.Join(t.TempDir(), "nested", "computed.json")

	cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-1", "--json-output", outJSONPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("CLI json-output failed: %v, stderr: %s", err, stderr.String())
	}

	data, err := os.ReadFile(outJSONPath)
	if err != nil {
		t.Fatalf("failed reading generated JSON: %v", err)
	}

	var parsed adapters.SerializedLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid json generated: %v", err)
	}

	if parsed.TicketID == nil || *parsed.TicketID != "TICKET-1" {
		t.Errorf("expected ticket_id TICKET-1, got %v", parsed.TicketID)
	}
	if parsed.WorkflowEfficiency.ContextWindowHealth.Status == "" {
		t.Errorf("expected populated workflow_efficiency in JSON output")
	}
	if parsed.Totals.Invocations == 0 {
		t.Errorf("expected non-zero totals in JSON output")
	}
}

func TestCLIAppendIdempotency(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	summaryFile := filepath.Join(t.TempDir(), "SUMMARY.md")

	initialContent := "# Ticket Summary\n\nSome introductory remarks.\n\n## Acceptance Criteria\n- [x] All done\n"
	if err := os.WriteFile(summaryFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// First append invocation
	cmd1 := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-1", "--append", summaryFile)
	if out, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("first append failed: %v, output: %s", err, string(out))
	}

	firstContent, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	firstStr := string(firstContent)
	if !strings.Contains(firstStr, "### Workflow Efficiency") {
		t.Fatalf("summary file missing metrics table after first append")
	}
	if !strings.Contains(firstStr, "## Acceptance Criteria") {
		t.Fatalf("summary file lost acceptance criteria")
	}

	// Second append invocation (should be strictly idempotent)
	cmd2 := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-1", "--append", summaryFile)
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("second append failed: %v, output: %s", err, string(out))
	}

	secondContent, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatal(err)
	}
	secondStr := string(secondContent)

	if secondStr != firstStr {
		t.Errorf("summary file append is not idempotent across repeated invocations:\n--- First ---\n%s\n--- Second ---\n%s", firstStr, secondStr)
	}

	// Verify count of headers is exactly 1
	if strings.Count(secondStr, "### Workflow Efficiency") != 1 {
		t.Errorf("expected exactly 1 '### Workflow Efficiency' heading, found %d", strings.Count(secondStr, "### Workflow Efficiency"))
	}
	if strings.Count(secondStr, "### Agentic Execution Metrics") != 1 {
		t.Errorf("expected exactly 1 '### Agentic Execution Metrics' heading, found %d", strings.Count(secondStr, "### Agentic Execution Metrics"))
	}
}

func TestCLINoValidLedgersExit(t *testing.T) {
	emptyDir := t.TempDir()

	cmd := exec.Command(cliBinaryPath, "--workspaces", emptyDir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected failure exit on empty workspaces dir")
	}
	if !strings.Contains(stderr.String(), "No valid agentic_metrics.json ledgers found") {
		t.Errorf("expected 'No valid agentic_metrics.json ledgers found' error, got: %s", stderr.String())
	}
}

func TestCLIServerSubcommandEdgeCases(t *testing.T) {
	t.Run("server with invalid flag returns exit code non-zero", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "server", "--unrecognized-flag-xyz")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected exit error for invalid flag")
		}
	})

	t.Run("server with invalid port format returns exit code non-zero", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "server", "--port", "not-a-port")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected exit error for invalid port format")
		}
	})
}

func TestCLIGenerateCommand(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_opencode.db")

	// Set up SQLite fixture database
	schema := `
		CREATE TABLE session (
			id TEXT PRIMARY KEY,
			parent_id TEXT,
			agent TEXT,
			model TEXT,
			title TEXT,
			time_created INTEGER,
			time_updated INTEGER
		);
		CREATE TABLE message (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
		CREATE TABLE part (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			message_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES
		('ses_root_gen', NULL, 'build', 'gemini-3.8-flash', 'Task AL-999 New Generator', 1760000000000, 1760000050000),
		('ses_child_gen', 'ses_root_gen', 'coder', 'gemini-3.8-flash', NULL, 1760000010000, 1760000045000);

		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_nonce', 'ses_child_gen', 'msg_1', 1760000015000, 1760000015000, '{"type":"tool","tool":"bash","state":{"input":"nonce_test_999"}}'),
		('p_comp', 'ses_root_gen', 'msg_1', 1760000020000, 1760000020000, '{"type":"compaction"}');

		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('msg_root', 'ses_root_gen', 1760000005000, 1760000025000, '{"role":"assistant","modelID":"gemini-3.8-flash","time":{"created":1760000005000,"completed":1760000025000},"tokens":{"input":1500,"output":300,"cache":{"read":4500},"reasoning":50},"cost":0.02}'),
		('msg_child', 'ses_child_gen', 1760000015000, 1760000040000, '{"role":"assistant","modelID":"gemini-3.8-flash","time":{"created":1760000015000,"completed":1760000040000},"tokens":{"input":2500,"output":500,"cache":{"read":7500},"reasoning":100},"cost":0.08}');
	`

	// Use pure Go sqlite driver to create fixture
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite fixture db: %v", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("failed creating schema in fixture db: %v", err)
	}
	db.Close()

	outJSON := filepath.Join(tempDir, "agentic_metrics.json")
	summaryMD := filepath.Join(tempDir, "SUMMARY.md")
	if err := os.WriteFile(summaryMD, []byte("# Initial Summary\n\nSome introductory remarks.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--ticket-id", "AL-999",
		"--nonce", "nonce_test_999",
		"--json-output", outJSON,
		"--append", summaryMD,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
	}

	// Verify JSON output
	data, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("failed reading output JSON: %v", err)
	}

	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed parsing output JSON: %v", err)
	}

	if parsed.TicketID == nil || *parsed.TicketID != "AL-999" {
		t.Errorf("expected ticket_id AL-999, got %v", parsed.TicketID)
	}
	if len(parsed.Invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(parsed.Invocations))
	}

	// Verify summary markdown
	mdData, err := os.ReadFile(summaryMD)
	if err != nil {
		t.Fatalf("failed reading updated summary: %v", err)
	}
	mdStr := string(mdData)
	if !strings.Contains(mdStr, "### Workflow Efficiency") {
		t.Errorf("expected Workflow Efficiency in updated summary")
	}
	if !strings.Contains(mdStr, "### Agentic Execution Metrics") {
		t.Errorf("expected Agentic Execution Metrics in updated summary")
	}
	if !strings.Contains(mdStr, "| **orchestrator** |") || !strings.Contains(mdStr, "| **coder** |") {
		t.Errorf("expected orchestrator and coder rows in markdown table")
	}

	// Verify idempotency
	cmd2 := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--ticket-id", "AL-999",
		"--nonce", "nonce_test_999",
		"--json-output", outJSON,
		"--append", summaryMD,
	)
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("second generate invocation failed: %v, out: %s", err, string(out))
	}

	mdData2, err := os.ReadFile(summaryMD)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(mdData2), "### Workflow Efficiency") != 1 {
		t.Errorf("expected exactly 1 Workflow Efficiency header after second run, found %d", strings.Count(string(mdData2), "### Workflow Efficiency"))
	}
}

func TestCLIGenerate_AdversarialEdgeCases(t *testing.T) {
	t.Run("generate with non-existent database falls back gracefully without crash", func(t *testing.T) {
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "fallback.json")
		nonExistentDB := filepath.Join(tempDir, "missing.db")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", nonExistentDB,
			"--ticket-id", "AL-FALLBACK",
			"--json-output", outJSON,
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected command to complete gracefully without crash, err: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid json generated: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-FALLBACK" {
			t.Errorf("expected AL-FALLBACK ticket ID, got %v", parsed.TicketID)
		}
		if len(parsed.Invocations) != 0 {
			t.Errorf("expected 0 invocations for missing DB, got %d", len(parsed.Invocations))
		}
	})

	t.Run("generate with corrupted existing JSON output preserves execution and safely recovers", func(t *testing.T) {
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "corrupted_existing.json")
		if err := os.WriteFile(outJSON, []byte("{broken json content!"), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", filepath.Join(tempDir, "none.db"),
			"--ticket-id", "AL-RECOVER",
			"--json-output", outJSON,
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected recovery and overwrite on corrupt existing json, err: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("recovered file must be valid JSON: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-RECOVER" {
			t.Errorf("expected AL-RECOVER, got %v", parsed.TicketID)
		}
	})

	t.Run("generate with zero tokens and zero duration produces valid markdown", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "zero.db")

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO session VALUES ('ses_zero', NULL, 'build', 'default_model', 'Zero Ticket', 1000, 1000);
			INSERT INTO message VALUES ('msg_zero', 'ses_zero', 1000, 1000, '{"role":"assistant","tokens":{"input":0,"output":0},"cost":0.0}');
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		summaryMD := filepath.Join(tempDir, "SUMMARY.md")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-ZERO",
			"--session-id", "ses_zero",
			"--append", summaryMD,
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed on zero tokens: %v, stderr: %s", err, stderr.String())
		}

		mdData, err := os.ReadFile(summaryMD)
		if err != nil {
			t.Fatal(err)
		}
		mdStr := string(mdData)
		if !strings.Contains(mdStr, "HEALTHY") {
			t.Errorf("expected HEALTHY health status on zero tokens, got:\n%s", mdStr)
		}
		if !strings.Contains(mdStr, "0.0%") {
			t.Errorf("expected 0.0%% cache hit rate / reasoning ratio on zero tokens")
		}
	})
}
