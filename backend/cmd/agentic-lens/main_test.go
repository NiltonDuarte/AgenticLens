package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
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
		os.RemoveAll(tempDir)
		panic(string(out))
	}
	cliBinaryPath = binPath

	code := m.Run()
	os.RemoveAll(tempDir)
	os.Exit(code)
}

func sampleFactJSON(ticketID, title string, duration int64, cost float64) string {
	parent := "ses-orch"
	ledger := domain.AgenticFactLedger{
		SchemaVersion: func() *int { v := domain.CurrentSchemaVersion; return &v }(),
		Migration:     &domain.MigrationInfo{SchemaVersion: domain.CurrentSchemaVersion, Status: domain.MigrationStatusCompleted, StepsCompleted: []string{domain.MigrationStepSourceBackfill}, StepsPending: []string{}},
		TicketID:      &ticketID, TicketTitle: &title,
		TicketType:  func() *domain.ConventionalTicketType { v := domain.TicketTypeFeat; return &v }(),
		StartedAt:   func() *string { v := "2026-10-17T10:00:00Z"; return &v }(),
		CompletedAt: func() *string { v := "2026-10-17T11:00:00Z"; return &v }(),
		UpdatedAt:   "2026-10-17T11:00:00Z",
		Code:        domain.CodeFacts{LinesAdded: 120, LinesDeleted: 20, TestLinesAdded: 60, ProdLinesAdded: 60, ProdLinesDeleted: 20, FilesChanged: 4, SDLCIterations: 2},
		Sessions: []domain.SessionFact{
			{ID: "ses-orch", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{ID: "ses-orch", Role: "orchestrator", Model: "claude-3-sonnet", User: "test-engineer", StartedAt: func() *string { v := "2026-10-17T10:00:00Z"; return &v }(), CompletedAt: func() *string { v := "2026-10-17T11:00:00Z"; return &v }(), ActiveDurationSeconds: duration, Turns: 6, Tokens: domain.InvocationTokens{Input: 5000, Output: 1000, CacheRead: 10000, Reasoning: 300, PeakContext: 80000}, CostUSD: cost, UserIterations: 0}}, Tools: []domain.SessionToolFact{{Name: "bash", Calls: 10, Errors: 1}}, Events: []domain.SessionEventFact{}, Anomalies: []domain.Anomaly{}},
			{ID: "ses-coder", Source: domain.SourceOpenCode, ParentID: &parent, Invocations: []domain.InvocationFact{{ID: "ses-coder", ParentID: &parent, Role: "coder", Model: "claude-3-opus", User: "test-engineer", StartedAt: func() *string { v := "2026-10-17T10:00:00Z"; return &v }(), CompletedAt: func() *string { v := "2026-10-17T11:00:00Z"; return &v }(), ActiveDurationSeconds: 60, Turns: 4, Tokens: domain.InvocationTokens{Input: 8000, Output: 2000, CacheRead: 12000, Reasoning: 500, PeakContext: 150000}, CostUSD: 0.75, UserIterations: 0}}, Events: []domain.SessionEventFact{}, Anomalies: []domain.Anomaly{}},
		},
	}
	data, err := json.Marshal(adapters.AgenticFactLedgerDTOFromDomain(ledger))
	if err != nil {
		panic(err)
	}
	return string(data)
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

func waitForServerReady(t *testing.T, targetURL string, timeout time.Duration) *http.Response {
	t.Helper()
	u, err := url.Parse(targetURL)
	if err != nil {
		t.Fatalf("invalid target url: %v", err)
	}
	host := u.Host
	if !strings.Contains(host, ":") {
		host = net.JoinHostPort(host, "80")
	}

	deadline := time.Now().Add(timeout)
	delay := 100 * time.Microsecond
	maxDelay := 5 * time.Millisecond

	for {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for server at %s after %v", targetURL, timeout)
		}

		conn, dialErr := net.DialTimeout("tcp", host, 10*time.Millisecond)
		if dialErr == nil {
			_ = conn.Close()
			resp, httpErr := http.Get(targetURL)
			if httpErr == nil {
				return resp
			}
		}

		select {
		case <-time.After(delay):
		}
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
}

func setupIsolatedEmptyDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "isolated_empty_opencode.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open isolated test db: %v", err)
	}
	defer db.Close()

	schema := `
		CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed creating schema in isolated test db: %v", err)
	}
	return dbPath
}

func TestCLIListCommand(t *testing.T) {
	t.Parallel()
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
	if !strings.Contains(out, "Anomalies") {
		t.Errorf("expected Anomalies column header in list, got: %s", out)
	}
	if !strings.Contains(out, "TICKET-1") || !strings.Contains(out, "TICKET-2") {
		t.Errorf("expected TICKET-1 and TICKET-2 in list, got: %s", out)
	}
	if !strings.Contains(out, "First Feature") || !strings.Contains(out, "Second Feature") {
		t.Errorf("expected titles in list, got: %s", out)
	}
}

func TestCLIAnomaliesCommand(t *testing.T) {
	t.Parallel()
	wsDir := setupTestWorkspace(t)

	// Add a ticket with an anomaly
	tanomDir := filepath.Join(wsDir, "AL-ANOMALY")
	if err := os.MkdirAll(tanomDir, 0755); err != nil {
		t.Fatalf("failed to create TICKET-ANOMALY dir: %v", err)
	}
	anomalyFactJSON := `{
		"schema_version": 4,
		"migration": {"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},
		"ticket_id": "AL-ANOMALY",
		"ticket_title": "Failing Feature",
		"updated_at": "2026-10-18T10:00:00Z",
		"sessions": [{
			"id": "ses-anom",
			"source": "opencode",
			"invocations": [],
			"events": [],
			"anomalies": [],
			"tools": [
			{
				"name": "bash",
				"calls": 10,
				"errors": 6
			}
			]
		}]
	}`
	if err := os.WriteFile(filepath.Join(tanomDir, "agentic_metrics.json"), []byte(anomalyFactJSON), 0644); err != nil {
		t.Fatalf("failed to write TICKET-ANOMALY ledger: %v", err)
	}

	t.Run("clean ticket shows no anomalies message", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-1", "--anomalies")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI --anomalies failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "No anomalies detected for ticket TICKET-1.") {
			t.Errorf("expected no anomalies message, got: %s", out)
		}
	})

	t.Run("ticket with anomalies formats table", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "AL-ANOMALY", "--anomalies")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI --anomalies failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "Found 1 anomalies for ticket AL-ANOMALY:") {
			t.Errorf("expected header with count, got: %s", out)
		}
		if !strings.Contains(out, "Severity") || !strings.Contains(out, "Rule ID") {
			t.Errorf("expected table columns, got: %s", out)
		}
		if !strings.Contains(out, "CRITICAL") || !strings.Contains(out, "TOOL_HIGH_FAILURE_RATE") {
			t.Errorf("expected CRITICAL TOOL_HIGH_FAILURE_RATE in output, got: %s", out)
		}
	})

	t.Run("list output shows anomaly counts", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--list")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI --list failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "Anomalies") {
			t.Errorf("expected Anomalies column in list, got: %s", out)
		}
		if !strings.Contains(out, "1 crit") {
			t.Errorf("expected '1 crit' in list output, got: %s", out)
		}
		if !strings.Contains(out, "0") {
			t.Errorf("expected '0' for clean tickets in list output, got: %s", out)
		}
	})

	t.Run("generate rejects legacy aggregate without writing ledger or summary", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "anom_gen.json")
		// Seed an existing JSON ledger with high failure rate
		seedJSON := `{
			"ticket_id": "AL-GEN-ANOM",
			"ticket_title": "Test Fixture Ticket",
			"tools": [
				{"session_id": "session-fixture", "name": "bash", "calls": 8, "errors": 6}
			]
		}`
		if err := os.WriteFile(outJSON, []byte(seedJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", filepath.Join(tempDir, "missing.db"),
			"--ticket-id", "AL-GEN-ANOM",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--append", filepath.Join(tempDir, "SUMMARY.md"),
			"--anomalies",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			t.Fatal("expected legacy ledger ingestion to fail")
		}
		if !strings.Contains(stderr.String(), "schema version is missing") || !strings.Contains(stderr.String(), "source-backed") {
			t.Fatalf("expected actionable legacy schema rejection, got stderr: %s", stderr.String())
		}
		content, err := os.ReadFile(outJSON)
		if err != nil || string(content) != seedJSON {
			t.Fatalf("rejected ingestion modified ledger: err=%v content=%s", err, content)
		}
		if _, err := os.Stat(filepath.Join(tempDir, "SUMMARY.md")); !os.IsNotExist(err) {
			t.Fatalf("rejected ingestion should not create summary, stat err=%v", err)
		}
	})

	t.Run("generate subcommand with --anomalies flag on clean ticket", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "clean_gen.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", filepath.Join(tempDir, "missing.db"),
			"--ticket-id", "AL-GEN-CLEAN",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--anomalies",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate --anomalies clean failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "No anomalies detected for ticket AL-GEN-CLEAN.") {
			t.Errorf("expected no anomalies message, got: %s", out)
		}
	})
}

func TestCLITicketFilterAndMarkdownOutput(t *testing.T) {
	t.Parallel()
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

		// Wait for server readiness
		resp := waitForServerReady(t, "http://localhost:19090/api/fleet", 2*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from /api/fleet, got %d", resp.StatusCode)
		}
	})
}

func TestCLIJsonOutput(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()

	t.Run("server with invalid flag returns exit code non-zero", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "server", "--unrecognized-flag-xyz")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected exit error for invalid flag")
		}
	})

	t.Run("server with invalid port format returns exit code non-zero", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "server", "--port", "not-a-port")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected exit error for invalid port format")
		}
	})

	t.Run("server default port is 8085 in help text", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "server", "--help")
		var env []string
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PORT=") {
				env = append(env, e)
			}
		}
		cmd.Env = env
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		_ = cmd.Run()
		if !strings.Contains(stderr.String(), "default 8085") {
			t.Errorf("expected default 8085 in server help text, got: %s", stderr.String())
		}
	})
}

func TestServer_FailsWithoutWorkspacesConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfgContent := "project_name: Test\nticket:\n  prefixes:\n    - prefix: TEST\nworkspaces:\n  src_folders:\n    - src\n"
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(cliBinaryPath, "server")
	cmd.Dir = tempDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected server to fail when workspaces config is missing, got nil")
	}

	expectedErr := "Error: workspaces directory must be configured in project configuration (workspaces.dir) or provided via --workspaces flag. Server cannot start without this information."
	if !strings.Contains(stderr.String(), expectedErr) {
		t.Errorf("expected stderr to contain %q, got: %s", expectedErr, stderr.String())
	}
}

func TestServer_FailsWhenWorkspacesDirDoesNotExist(t *testing.T) {
	nonExistentPath := "/nonexistent/workspaces/dir/xyz"
	cmd := exec.Command(cliBinaryPath, "server", "--workspaces", nonExistentPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected server to fail when workspaces dir does not exist, got nil")
	}

	expectedErr := fmt.Sprintf("Error: workspaces directory %q does not exist. Server cannot start.", nonExistentPath)
	if !strings.Contains(stderr.String(), expectedErr) {
		t.Errorf("expected stderr to contain %q, got: %s", expectedErr, stderr.String())
	}
}

func TestCLIServerPortEnvironment(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	cmd := exec.Command(cliBinaryPath, "server", "--workspaces", wsDir)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed starting server subcommand with PORT env: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	resp := waitForServerReady(t, fmt.Sprintf("http://localhost:%d/api/fleet", port), 2*time.Second)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /api/fleet on port %d, got %d", port, resp.StatusCode)
	}
}

func TestCLIServerPortFlagPrecedence(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	t.Run("valid flag overrides valid PORT env", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to find free port: %v", err)
		}
		flagPort := ln.Addr().(*net.TCPAddr).Port
		ln.Close()

		envPort := flagPort + 1
		if envPort > 65530 {
			envPort = flagPort - 1
		}

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(flagPort), "--workspaces", wsDir)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", envPort))

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand: %v", err)
		}
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		resp := waitForServerReady(t, fmt.Sprintf("http://localhost:%d/api/fleet", flagPort), 2*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from /api/fleet on flag port %d, got %d", flagPort, resp.StatusCode)
		}
	})

	t.Run("valid flag overrides invalid PORT env", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to find free port: %v", err)
		}
		flagPort := ln.Addr().(*net.TCPAddr).Port
		ln.Close()

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(flagPort), "--workspaces", wsDir)
		cmd.Env = append(os.Environ(), "PORT=not-a-valid-port")

		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand: %v", err)
		}
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		resp := waitForServerReady(t, fmt.Sprintf("http://localhost:%d/api/fleet", flagPort), 2*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from /api/fleet on flag port %d, got %d", flagPort, resp.StatusCode)
		}
	})
}

func TestCLIServerInvalidPortRange(t *testing.T) {
	t.Parallel()
	wsDir := setupTestWorkspace(t)

	testCases := []struct {
		name string
		args []string
		env  []string
	}{
		{
			name: "port 0 via flag",
			args: []string{"server", "--port", "0", "--workspaces", wsDir},
		},
		{
			name: "port 70000 via flag",
			args: []string{"server", "--port", "70000", "--workspaces", wsDir},
		},
		{
			name: "port 65536 via flag",
			args: []string{"server", "--port", "65536", "--workspaces", wsDir},
		},
		{
			name: "negative port via flag",
			args: []string{"server", "--port", "-5", "--workspaces", wsDir},
		},
		{
			name: "non-number port via flag",
			args: []string{"server", "--port", "not-a-number", "--workspaces", wsDir},
		},
		{
			name: "port 0 via PORT env",
			args: []string{"server", "--workspaces", wsDir},
			env:  []string{"PORT=0"},
		},
		{
			name: "port 70000 via PORT env",
			args: []string{"server", "--workspaces", wsDir},
			env:  []string{"PORT=70000"},
		},
		{
			name: "port 65536 via PORT env",
			args: []string{"server", "--workspaces", wsDir},
			env:  []string{"PORT=65536"},
		},
		{
			name: "negative port via PORT env",
			args: []string{"server", "--workspaces", wsDir},
			env:  []string{"PORT=-10"},
		},
		{
			name: "non-number port via PORT env",
			args: []string{"server", "--workspaces", wsDir},
			env:  []string{"PORT=not-a-number"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(cliBinaryPath, tc.args...)
			if len(tc.env) > 0 {
				cmd.Env = append(os.Environ(), tc.env...)
			}
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected error for %s, got success", tc.name)
			}
			errOutput := stderr.String()
			if !strings.Contains(errOutput, "invalid port") &&
				!strings.Contains(errOutput, "out of valid range") &&
				!strings.Contains(errOutput, "parse error") {
				t.Errorf("expected invalid port diagnostic message, got: %s", errOutput)
			}
		})
	}
}

func TestCLIServerAddressAlreadyInUse(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	t.Run("conflict diagnostic via --port flag", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to listen on test port: %v", err)
		}
		defer ln.Close()

		port := ln.Addr().(*net.TCPAddr).Port

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(port), "--workspaces", wsDir)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err == nil {
			t.Fatalf("expected error when binding to already occupied port %d, got success", port)
		}

		out := stderr.String()
		if !strings.Contains(out, "already in use") {
			t.Errorf("expected 'already in use' in stderr, got: %s", out)
		}
		if !strings.Contains(out, "--port or the PORT environment variable") {
			t.Errorf("expected suggestion to use --port or PORT in stderr, got: %s", out)
		}
	})

	t.Run("conflict diagnostic via PORT environment variable", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to listen on test port: %v", err)
		}
		defer ln.Close()

		port := ln.Addr().(*net.TCPAddr).Port

		cmd := exec.Command(cliBinaryPath, "server", "--workspaces", wsDir)
		cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err == nil {
			t.Fatalf("expected error when binding to already occupied port %d via PORT env, got success", port)
		}

		out := stderr.String()
		if !strings.Contains(out, "already in use") {
			t.Errorf("expected 'already in use' in stderr, got: %s", out)
		}
		if !strings.Contains(out, "--port or the PORT environment variable") {
			t.Errorf("expected suggestion to use --port or PORT in stderr, got: %s", out)
		}
	})
}

func TestCLIGenerateCommand(t *testing.T) {
	t.Parallel()
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
		"--ticket-title", "Test Fixture Ticket",
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
	if parsed.SchemaVersion == nil || *parsed.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, parsed.SchemaVersion)
	}
	if len(parsed.ToDomain().InvocationFacts()) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(parsed.ToDomain().InvocationFacts()))
	}
	if strings.Contains(string(data), `"error_message"`) {
		t.Errorf("expected zero error_message in output JSON: %s", string(data))
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
		"--ticket-title", "Test Fixture Ticket",
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

func TestCLIGenerate_SparseAnomalyEvidenceAndZeroErrorMessage(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_prune.db")

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
		('ses_root', NULL, 'build', 'gemini-3.8-flash', 'Task AL-888 Feature', 1760000000000, 1760000050000),
		('ses_child', 'ses_root', 'coder', 'gemini-3.8-flash', NULL, 1760000010000, 1760000045000);

		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES
		('p_nonce', 'ses_child', 'msg_1', 1760000015000, 1760000015000, '{"type":"tool","tool":"bash","state":{"input":"nonce_test_888"}}'),
		('p_ok_tool', 'ses_child', 'msg_1', 1760000016000, 1760000017000, '{"type":"tool","tool":"read_file","state":{"status":"completed"}}'),
		('p_err_tool', 'ses_child', 'msg_1', 1760000018000, 1760000019000, '{"type":"tool","tool":"execute_cmd","state":{"status":"error","error":{"message":"API_KEY=secret_xyz failed to connect to database postgres://admin:secret123@prod.db"}}}'),
		('p_comp', 'ses_root', 'msg_1', 1760000020000, 1760000020000, '{"type":"compaction"}');

		INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('msg_root', 'ses_root', 1760000005000, 1760000025000, '{"role":"assistant","modelID":"gemini-3.8-flash","time":{"created":1760000005000,"completed":1760000025000},"tokens":{"input":1500,"output":300,"cache":{"read":4500},"reasoning":50},"cost":0.02}'),
		('msg_child', 'ses_child', 1760000015000, 1760000040000, '{"role":"assistant","modelID":"gemini-3.8-flash","time":{"created":1760000015000,"completed":1760000040000},"tokens":{"input":2500,"output":500,"cache":{"read":7500},"reasoning":100},"cost":0.08}');
	`

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite fixture db: %v", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		t.Fatalf("failed creating schema: %v", err)
	}
	db.Close()

	outJSON := filepath.Join(tempDir, "agentic_metrics.json")
	summaryMD := filepath.Join(tempDir, "SUMMARY.md")
	if err := os.WriteFile(summaryMD, []byte("# Ticket AL-888 Summary\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--ticket-id", "AL-888",
		"--ticket-title", "Test Fixture Ticket",
		"--nonce", "nonce_test_888",
		"--json-output", outJSON,
		"--append", summaryMD,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v, stderr: %s, stdout: %s", err, stderr.String(), stdout.String())
	}

	rawJSON, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("failed reading output JSON: %v", err)
	}

	// 1. File size must be well under 5 KB
	if len(rawJSON) >= 5120 {
		t.Errorf("expected output JSON under 5120 bytes, got %d", len(rawJSON))
	}

	// 2. Raw JSON must contain zero sensitive error strings and zero "error_message" keys
	if strings.Contains(string(rawJSON), "API_KEY=secret_xyz") {
		t.Errorf("output JSON leaked secret error string: %s", string(rawJSON))
	}
	if strings.Contains(string(rawJSON), "postgres://admin:secret123") {
		t.Errorf("output JSON leaked connection string: %s", string(rawJSON))
	}
	if strings.Contains(string(rawJSON), `"error_message"`) {
		t.Errorf("output JSON must omit error_message key when nil: %s", string(rawJSON))
	}

	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		t.Fatalf("failed unmarshaling output JSON: %v", err)
	}

	// 3. Schema version must be 3
	if parsed.SchemaVersion == nil || *parsed.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, parsed.SchemaVersion)
	}

	// 4. In-memory derived metrics (invocations, tokens, cost) are fully preserved
	if len(parsed.ToDomain().InvocationFacts()) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(parsed.ToDomain().InvocationFacts()))
	}

	// 5. Events contain direct anomaly evidence only; isolated errors and compactions are omitted.
	var persistedEventCount int
	for _, ev := range parsed.ToDomain().EventFacts() {
		persistedEventCount++
		if ev.ErrorMessage != nil {
			t.Errorf("expected event ErrorMessage to be nil, got: %s", *ev.ErrorMessage)
		}
	}
	if persistedEventCount != 0 {
		t.Errorf("expected no unrelated event evidence to persist, got %d rows", persistedEventCount)
	}
	var executeCalls, executeErrors int64
	for _, fact := range parsed.ToDomain().ToolFacts() {
		if fact.Name == "execute_cmd" {
			executeCalls += fact.Calls
			executeErrors += fact.Errors
		}
	}
	if executeCalls != 1 || executeErrors != 1 {
		t.Errorf("isolated error totals were lost: calls=%d errors=%d", executeCalls, executeErrors)
	}
}

func TestCLIGenerate_HighFailureAggregateKeepsSummaryWithoutEventRows(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "aggregate.json")
	seedJSON := `{
		"schema_version":4,
		"migration":{"schema_version":4,"status":"completed","steps_completed":["source_backfill"],"steps_pending":[]},
		"ticket_id":"AL-GEN-EVIDENCE",
		"ticket_title":"Aggregate evidence fixture",
		"sessions":[{
			"id":"session-fixture",
			"source":"opencode",
			"invocations":[],
			"tools":[{"name":"bash","calls":8,"errors":6}],
			"anomalies":[],
			"events":[
				{"event_id":"failed-1","timestamp":"2026-10-26T10:00:01Z","turn":1,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"failed-2","timestamp":"2026-10-26T10:00:02Z","turn":2,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"failed-3","timestamp":"2026-10-26T10:00:03Z","turn":3,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"failed-4","timestamp":"2026-10-26T10:00:04Z","turn":4,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"failed-5","timestamp":"2026-10-26T10:00:05Z","turn":5,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"failed-6","timestamp":"2026-10-26T10:00:06Z","turn":6,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"error"},
				{"event_id":"success-1","timestamp":"2026-10-26T10:00:07Z","turn":7,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"completed"},
				{"event_id":"success-2","timestamp":"2026-10-26T10:00:08Z","turn":8,"role":"coder","event_type":"tool_call","tool_name":"bash","status":"completed"}
			]
		}]
	}`
	if err := os.WriteFile(outJSON, []byte(seedJSON), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, "generate", "--db-path", filepath.Join(tempDir, "missing.db"), "--ticket-id", "AL-GEN-EVIDENCE", "--json-output", outJSON, "--allow-empty")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate with direct aggregate evidence failed: %v stderr=%s stdout=%s", err, stderr.String(), stdout.String())
	}
	raw, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatal(err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatal(err)
	}
	facts := parsed.ToDomain()
	var highFailure *domain.Anomaly
	for i := range facts.Anomalies {
		if facts.Anomalies[i].RuleID == domain.RuleToolHighFailureRate {
			highFailure = &facts.Anomalies[i]
			break
		}
	}
	if highFailure == nil {
		t.Fatalf("expected high failure summary, got %#v", facts.Anomalies)
	}
	if len(highFailure.Evidence) != 0 {
		t.Fatalf("persisted aggregate summary retained refs: %#v", highFailure.Evidence)
	}
	if len(facts.EventFacts()) != 0 {
		t.Fatalf("persisted ledger retained events: %#v", facts.EventFacts())
	}
}

func TestCLIGenerate_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("generate with non-existent database falls back gracefully without crash", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "fallback.json")
		nonExistentDB := filepath.Join(tempDir, "missing.db")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", nonExistentDB,
			"--ticket-id", "AL-FALLBACK",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--allow-empty",
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
		if len(parsed.ToDomain().InvocationFacts()) != 0 {
			t.Errorf("expected 0 invocations for missing DB, got %d", len(parsed.ToDomain().InvocationFacts()))
		}
	})

	t.Run("generate with corrupted existing JSON output fails closed", func(t *testing.T) {
		t.Parallel()
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
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--allow-empty",
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			t.Fatal("expected corrupt existing ledger to be rejected")
		}
		if !strings.Contains(stderr.String(), "decoding ledger metadata") {
			t.Fatalf("expected malformed ledger error, got: %s", stderr.String())
		}
		data, err := os.ReadFile(outJSON)
		if err != nil || string(data) != "{broken json content!" {
			t.Fatalf("rejected ingestion changed corrupt ledger: err=%v content=%s", err, data)
		}
	})

	t.Run("generate with zero tokens and zero duration produces valid markdown", func(t *testing.T) {
		t.Parallel()
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
			"--ticket-title", "Test Fixture Ticket",
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

	t.Run("generate enforces mandatory non-zero completed_at in JSON output", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "completed_test.db")

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO session VALUES ('ses_comp', NULL, 'build', 'default_model', 'Completed At Test', 1000, 1000);
			INSERT INTO message VALUES ('msg_comp', 'ses_comp', 1000, 1000, '{"role":"assistant","tokens":{"input":10,"output":5},"cost":0.01}');
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		outJSON := filepath.Join(tempDir, "agentic_metrics.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-COMPLETED",
			"--ticket-title", "Completed At Title",
			"--session-id", "ses_comp",
			"--json-output", outJSON,
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var fact domain.AgenticFactLedger
		if err := json.Unmarshal(data, &fact); err != nil {
			t.Fatalf("failed unmarshaling fact ledger: %v", err)
		}

		if fact.CompletedAt == nil || strings.TrimSpace(*fact.CompletedAt) == "" {
			t.Fatalf("expected non-nil non-empty CompletedAt, got %v", fact.CompletedAt)
		}
	})

	t.Run("generate respects explicit valid completed_at flag and sanitizes whitespace", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "explicit_completed.db")

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO session VALUES ('ses_explicit', NULL, 'build', 'default_model', 'Explicit Test', 1000, 1000);
			INSERT INTO message VALUES ('msg_explicit', 'ses_explicit', 1000, 1000, '{"role":"assistant","tokens":{"input":5,"output":2},"cost":0.005}');
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		explicitTS := "2026-10-21T15:00:00.000000+00:00"
		outJSON := filepath.Join(tempDir, "agentic_metrics_explicit.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-EXPLICIT-COMPLETED",
			"--ticket-title", "Explicit Title",
			"--session-id", "ses_explicit",
			"--completed-at", explicitTS,
			"--json-output", outJSON,
		)

		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("generate with explicit completed-at failed: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var fact domain.AgenticFactLedger
		if err := json.Unmarshal(data, &fact); err != nil {
			t.Fatal(err)
		}
		if fact.CompletedAt == nil || *fact.CompletedAt != explicitTS {
			t.Errorf("expected completed_at %s, got %v", explicitTS, fact.CompletedAt)
		}

		// Also verify whitespace-only completed-at falls back to non-zero ISO timestamp
		outJSONWhitespace := filepath.Join(tempDir, "agentic_metrics_ws.json")
		cmdWS := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-WS-COMPLETED",
			"--ticket-title", "WS Title",
			"--session-id", "ses_explicit",
			"--completed-at", "   ",
			"--json-output", outJSONWhitespace,
		)
		var stderrWS bytes.Buffer
		cmdWS.Stderr = &stderrWS
		if err := cmdWS.Run(); err != nil {
			t.Fatalf("generate with whitespace completed-at failed: %v, stderr: %s", err, stderrWS.String())
		}
		dataWS, err := os.ReadFile(outJSONWhitespace)
		if err != nil {
			t.Fatal(err)
		}
		var factWS domain.AgenticFactLedger
		if err := json.Unmarshal(dataWS, &factWS); err != nil {
			t.Fatal(err)
		}
		if factWS.CompletedAt == nil || strings.TrimSpace(*factWS.CompletedAt) == "" {
			t.Errorf("expected non-empty resolved completed_at on whitespace flag, got %v", factWS.CompletedAt)
		}
	})

	t.Run("generate falls back to cached code facts on git error when available", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "agentic_metrics.json")
		dbPath := filepath.Join(tempDir, "test.db")

		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO session VALUES ('ses_fallback', NULL, 'build', 'default_model', 'Fallback Test', 1000, 1000);
			INSERT INTO message VALUES ('msg_fallback', 'ses_fallback', 1000, 1000, '{"role":"assistant","tokens":{"input":10,"output":5},"cost":0.01}');
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		tid := "AL-GIT-FALLBACK"
		ver := domain.CurrentSchemaVersion
		existingDTO := adapters.AgenticFactLedgerDTO{
			TicketID:      &tid,
			SchemaVersion: &ver,
			Migration: &adapters.MigrationInfoDTO{
				SchemaVersion:  domain.CurrentSchemaVersion,
				Status:         string(domain.MigrationStatusCompleted),
				StepsCompleted: []string{domain.MigrationStepSourceBackfill},
				StepsPending:   []string{},
			},
			Sessions:      []adapters.SessionFactDTO{},
			Code: adapters.CodeFactsDTO{
				LinesAdded:       123,
				LinesDeleted:     45,
				ProdLinesAdded:   100,
				ProdLinesDeleted: 40,
				TestLinesAdded:   23,
				TestLinesDeleted: 5,
				FilesChanged:     3,
				SDLCIterations:   4,
			},
		}
		existingData, err := json.Marshal(existingDTO)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(outJSON, existingData, 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-GIT-FALLBACK",
			"--ticket-title", "Fallback Git Code Facts",
			"--ticket-type", "feat",
			"--session-id", "ses_fallback",
			"--code-ref", "invalid-nonexistent-git-branch-xyz-999",
			"--json-output", outJSON,
		)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected generate to succeed using cached code facts fallback, err=%v, stderr=%s", err, stderr.String())
		}

		if !strings.Contains(stderr.String(), "WARNING: Failed to compute git code facts") {
			t.Errorf("expected warning in stderr about git facts failure, got: %s", stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatal(err)
		}

		if parsed.Code.LinesAdded != 123 || parsed.Code.LinesDeleted != 45 {
			t.Errorf("expected cached code facts preserved (LinesAdded: 123, LinesDeleted: 45), got: %+v", parsed.Code)
		}
	})
}

func TestCLIGenerate_UnresolvedSessionEnforcement(t *testing.T) {
	t.Parallel()

	t.Run("generate fails with status 1 and logs error when root session cannot be identified", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		emptyDB := filepath.Join(tempDir, "empty.db")
		db, err := sql.Open("sqlite", emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		targetJSON := filepath.Join(tempDir, "should_not_exist.json")
		targetMD := filepath.Join(tempDir, "should_not_exist.md")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "NON_EXISTENT",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", targetJSON,
			"--append", targetMD,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err == nil {
			t.Fatal("expected generate to fail with exit status 1, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: unable to identify root session for ticket 'NON_EXISTENT'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}

		if _, err := os.Stat(targetJSON); !os.IsNotExist(err) {
			t.Fatalf("expected target JSON file not to be created, but it exists: %s", targetJSON)
		}
		if _, err := os.Stat(targetMD); !os.IsNotExist(err) {
			t.Fatalf("expected target markdown file not to be created, but it exists: %s", targetMD)
		}
	})

	t.Run("existing summary file is not modified when generation fails", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		emptyDB := filepath.Join(tempDir, "empty.db")
		db, err := sql.Open("sqlite", emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		summaryMD := filepath.Join(tempDir, "SUMMARY.md")
		initialContent := "# Preserved Summary\n\nOriginal content must not change.\n"
		if err := os.WriteFile(summaryMD, []byte(initialContent), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "NON_EXISTENT_2",
			"--ticket-title", "Test Fixture Ticket",
			"--append", summaryMD,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			t.Fatal("expected command to fail with exit status 1")
		}

		content, err := os.ReadFile(summaryMD)
		if err != nil {
			t.Fatalf("failed reading summary file: %v", err)
		}
		if string(content) != initialContent {
			t.Fatalf("expected summary file to remain unchanged, got:\n%s", string(content))
		}
	})

	t.Run("passing --allow-empty allows execution to proceed when sessions are absent", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "allowed_empty.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", filepath.Join(tempDir, "missing.db"),
			"--ticket-id", "AL-ALLOW-EMPTY",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected --allow-empty to exit with code 0, got err: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("expected output JSON to exist: %v", err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("expected valid JSON: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-ALLOW-EMPTY" {
			t.Errorf("expected AL-ALLOW-EMPTY, got %v", parsed.TicketID)
		}
	})

	t.Run("valid ticket with identifiable session succeeds without --allow-empty", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "valid.db")
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			INSERT INTO session VALUES ('ses_root_valid', NULL, 'build', 'default_model', 'Work on AL-VALID feature', 1000, 2000);
			INSERT INTO message VALUES ('msg_valid', 'ses_root_valid', 1100, 1900, '{"role":"assistant","tokens":{"input":100,"output":50},"cost":0.001}');
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		outJSON := filepath.Join(tempDir, "valid.json")
		summaryMD := filepath.Join(tempDir, "SUMMARY.md")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", dbPath,
			"--ticket-id", "AL-VALID",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", outJSON,
			"--append", summaryMD,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected generate to succeed for valid session: %v, stderr: %s", err, stderr.String())
		}

		if _, err := os.Stat(outJSON); err != nil {
			t.Fatalf("expected JSON output file to exist: %v", err)
		}
		if _, err := os.Stat(summaryMD); err != nil {
			t.Fatalf("expected summary markdown file to exist: %v", err)
		}
	})

	t.Run("generate without ticket id displays UNKNOWN on failure", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", filepath.Join(tempDir, "missing.db"),
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected failure without root session or ticket id")
		}
		expectedErr := "Error: missing ticket ID. Specify --ticket-id explicitly."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}
	})

	t.Run("cached ledger facts with invocations allows generate to succeed without root session in DB", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		emptyDB := filepath.Join(tempDir, "empty.db")
		db, err := sql.Open("sqlite", emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		cachedJSON := filepath.Join(tempDir, "cached_with_inv.json")
		cachedJSONContent := `
{
  "schema_version": 4,
  "migration": {
    "schema_version": 4,
    "status": "completed",
    "steps_completed": [
      "source_backfill"
    ],
    "steps_pending": []
  },
  "ticket_id": "AL-CACHED-INV",
  "sessions": [
    {
      "id": "ses_cached_inv_1",
      "source": "opencode",
      "invocations": [
        {
          "id": "ses_cached_inv_1",
          "user": "engineer",
          "role": "coder",
          "model": "gpt-4o",
          "tokens": {
            "input": 150,
            "output": 75,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "cost_usd": 0.003,
          "parent_id": null,
          "started_at": null,
          "completed_at": null,
          "active_duration_seconds": 0,
          "turns": 0,
          "compactions": 0,
          "user_iterations": 0
        }
      ],
      "tools": [],
      "events": [],
      "anomalies": []
    }
  ]
}
`
		if err := os.WriteFile(cachedJSON, []byte(cachedJSONContent), 0644); err != nil {
			t.Fatal(err)
		}

		summaryMD := filepath.Join(tempDir, "SUMMARY.md")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-CACHED-INV",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", cachedJSON,
			"--append", summaryMD,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected generate to succeed with cached invocations, got err: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(cachedJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}
		if len(parsed.ToDomain().InvocationFacts()) != 1 {
			t.Fatalf("expected 1 invocation from cached facts, got %d", len(parsed.ToDomain().InvocationFacts()))
		}
		if parsed.ToDomain().InvocationFacts()[0].ID != "ses_cached_inv_1" {
			t.Errorf("expected invocation ID ses_cached_inv_1, got %s", parsed.ToDomain().InvocationFacts()[0].ID)
		}
	})

	t.Run("cached ledger facts without invocations fails when DB has no root session", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		emptyDB := filepath.Join(tempDir, "empty.db")
		db, err := sql.Open("sqlite", emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		cachedJSON := filepath.Join(tempDir, "cached_no_inv.json")
		initialJSON := `{
  "schema_version": 4,
  "migration": {"schema_version": 4, "status": "completed", "steps_completed": ["source_backfill"], "steps_pending": []},
  "ticket_id": "AL-CACHED-NO-INV",
  "ticket_title": "Test Fixture Ticket",
  "sessions": []
}`
		if err := os.WriteFile(cachedJSON, []byte(initialJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-CACHED-NO-INV",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", cachedJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err == nil {
			t.Fatal("expected generate to fail with exit status 1, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: unable to identify root session for ticket 'AL-CACHED-NO-INV'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}

		content, err := os.ReadFile(cachedJSON)
		if err != nil {
			t.Fatalf("failed reading cached JSON file: %v", err)
		}
		if string(content) != initialJSON {
			t.Fatalf("expected cached JSON file to remain unchanged, got:\n%s", string(content))
		}
	})

	t.Run("both DB and cached facts have 0 thread IDs and 0 invocations succeeds when --allow-empty is passed", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		emptyDB := filepath.Join(tempDir, "empty.db")
		db, err := sql.Open("sqlite", emptyDB)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
			CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
			CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		cachedJSON := filepath.Join(tempDir, "cached_empty.json")
		initialJSON := `{
  "schema_version": 4,
  "migration": {"schema_version": 4, "status": "completed", "steps_completed": ["source_backfill"], "steps_pending": []},
  "ticket_id": "AL-EMPTY-BOTH",
  "ticket_title": "Test Fixture Ticket",
  "sessions": []
}`
		if err := os.WriteFile(cachedJSON, []byte(initialJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-EMPTY-BOTH",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", cachedJSON,
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected --allow-empty to succeed with code 0, got err: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(cachedJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("expected valid JSON: %v", err)
		}
		if parsed.TicketID == nil || *parsed.TicketID != "AL-EMPTY-BOTH" {
			t.Errorf("expected AL-EMPTY-BOTH, got %v", parsed.TicketID)
		}
		if len(parsed.ToDomain().InvocationFacts()) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(parsed.ToDomain().InvocationFacts()))
		}
	})

	t.Run("database error during workspace path fallback fails cleanly with exit code 1", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		brokenDB := filepath.Join(tempDir, "broken.db")
		db, err := sql.Open("sqlite", brokenDB)
		if err != nil {
			t.Fatal(err)
		}
		// session table exists, but part table is missing.
		// Title search returns sql.ErrNoRows, and workspace path fallback fails when querying part table.
		_, err = db.Exec(`
			CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		`)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
		db.Close()

		targetJSON := filepath.Join(tempDir, "should_not_exist.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", brokenDB,
			"--ticket-id", "AL-DB-ERR",
			"--ticket-title", "Test Fixture Ticket",
			"--json-output", targetJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err == nil {
			t.Fatal("expected failure on database error during fallback")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: unable to identify root session for ticket 'AL-DB-ERR'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}

		if _, err := os.Stat(targetJSON); !os.IsNotExist(err) {
			t.Fatalf("expected target JSON not to exist: %s", targetJSON)
		}
	})
}

func TestCLIBackfillCommand(t *testing.T) {
	t.Parallel()
	wsDir := setupTestWorkspace(t)
	t.Run("backfill requires explicit --source flag", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "backfill", "--workspaces", wsDir)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil || !strings.Contains(stderr.String(), "--source flag is required for backfill") {
			t.Fatalf("expected missing --source to fail explicitly, err=%v stderr=%s", err, stderr.String())
		}
	})

	t.Run("recompute-git requires an explicit code ref", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "backfill", "--source", "opencode", "--workspaces", wsDir, "--recompute-git")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil || !strings.Contains(stderr.String(), "--code-ref is required with --recompute-git") {
			t.Fatalf("expected missing code ref to fail explicitly, err=%v stderr=%s", err, stderr.String())
		}
	})

	t.Run("recompute-git requires an explicit ticket id", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "backfill", "--source", "opencode", "--workspaces", wsDir, "--recompute-git", "--code-ref", "HEAD")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil || !strings.Contains(stderr.String(), "--ticket-id is required when --recompute-git is specified") {
			t.Fatalf("expected missing ticket id to fail explicitly, err=%v stderr=%s", err, stderr.String())
		}
	})

	t.Run("dry-run backfill outputs preview without error", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"backfill",
			"--source", "opencode",
			"--workspaces", wsDir,
			"--dry-run",
			"--migrate-schema",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("backfill dry-run failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "TELEMETRY BACKFILL & SCHEMA MIGRATION [DRY-RUN PREVIEW]") {
			t.Errorf("expected dry-run header, got: %s", out)
		}
		if !strings.Contains(out, "TICKET-1") || !strings.Contains(out, "TICKET-2") {
			t.Errorf("expected ticket results in output, got: %s", out)
		}
		if !strings.Contains(out, "Summary: Scanned: 2") {
			t.Errorf("expected 2 scanned tickets in summary, got: %s", out)
		}
	})

	t.Run("backfill with ticket-id filter executes and writes migration info", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"backfill",
			"--source", "opencode",
			"--workspaces", wsDir,
			"--ticket-id", "TICKET-1",
			"--recompute-git",
			"--code-ref", "HEAD",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("backfill failed: %v, stderr: %s", err, stderr.String())
		}
		var ticketRow []string
		for _, line := range strings.Split(stdout.String(), "\n") {
			fields := strings.Fields(line)
			if len(fields) > 0 && fields[0] == "TICKET-1" {
				ticketRow = fields
				break
			}
		}
		if len(ticketRow) < 4 || ticketRow[3] != "yes" {
			t.Fatalf("expected the requested Git recompute row to report Git yes, got: %s", stdout.String())
		}
		if !strings.Contains(stdout.String(), "Updated: 1") {
			t.Fatalf("expected TICKET-1 recompute to be applied instead of skipped, got: %s", stdout.String())
		}

		data, err := os.ReadFile(filepath.Join(wsDir, "TICKET-1", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading updated file: %v", err)
		}

		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed parsing updated JSON: %v", err)
		}

		if parsed.SchemaVersion == nil || *parsed.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, parsed.SchemaVersion)
		}
		if parsed.Migration != nil {
			t.Errorf("expected successful final backfill to remove temporary migration metadata, got %v", parsed.Migration)
		}
		if strings.Contains(string(data), `"migration"`) {
			t.Errorf("successful ticket-filtered backfill must omit migration metadata: %s", data)
		}
	})
}

func TestCLIVersionCommand(t *testing.T) {
	t.Parallel()

	t.Run("version subcommand outputs formatted string", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version command failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected output to start with 'agentic-lens version ', got: %s", out)
		}
		if !strings.Contains(out, "commit:") || !strings.Contains(out, "built:") {
			t.Errorf("expected commit and built metadata in version output, got: %s", out)
		}
	})

	t.Run("-v flag outputs formatted string", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "-v")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("-v failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected output to start with 'agentic-lens version ', got: %s", out)
		}
	})

	t.Run("--version flag outputs formatted string", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "--version")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("--version failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected output to start with 'agentic-lens version ', got: %s", out)
		}
	})

	t.Run("version --short outputs raw version string", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "--short")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version --short failed: %v, stderr: %s", err, stderr.String())
		}

		out := strings.TrimSpace(stdout.String())
		if !strings.HasPrefix(out, "v") {
			t.Errorf("expected short version to start with 'v', got: %s", out)
		}
		if strings.Contains(out, "commit:") {
			t.Errorf("short version should not contain commit metadata, got: %s", out)
		}
	})

	t.Run("version --json outputs valid JSON payload", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "--json")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version --json failed: %v, stderr: %s", err, stderr.String())
		}

		var payload struct {
			Version   string `json:"version"`
			Commit    string `json:"commit"`
			BuildDate string `json:"build_date"`
		}

		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("version --json failed to produce valid JSON: %v, stdout: %s", err, stdout.String())
		}

		if payload.Version == "" {
			t.Errorf("expected non-empty version in JSON output")
		}
		if payload.Commit == "" {
			t.Errorf("expected non-empty commit in JSON output")
		}
		if payload.BuildDate == "" {
			t.Errorf("expected non-empty build_date in JSON output")
		}

		var domainInfo domain.BuildInfo
		if err := json.Unmarshal(stdout.Bytes(), &domainInfo); err != nil {
			t.Fatalf("failed unmarshaling directly into domain.BuildInfo: %v", err)
		}
		if domainInfo.Version != payload.Version || domainInfo.GitCommit != payload.Commit || domainInfo.BuildDate != payload.BuildDate {
			t.Errorf("mismatch between DTO and domain.BuildInfo: dto=%+v, domain=%+v", payload, domainInfo)
		}
	})

	t.Run("version -short single-dash flag", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "-short")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version -short failed: %v, stderr: %s", err, stderr.String())
		}

		out := strings.TrimSpace(stdout.String())
		if !strings.HasPrefix(out, "v") {
			t.Errorf("expected short version to start with 'v', got: %s", out)
		}
	})

	t.Run("version -json single-dash flag", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "-json")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version -json failed: %v, stderr: %s", err, stderr.String())
		}

		var info domain.BuildInfo
		if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
			t.Fatalf("version -json failed to produce valid JSON: %v", err)
		}
		if info.Version == "" {
			t.Errorf("expected non-empty version")
		}
	})

	t.Run("version with both --json and --short flags prefers JSON", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "--json", "--short")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("version --json --short failed: %v, stderr: %s", err, stderr.String())
		}

		var info domain.BuildInfo
		if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
			t.Fatalf("expected valid JSON output when --json is provided: %v", err)
		}
	})

	t.Run("version with unrecognized flag returns non-zero exit code", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "version", "--bogus-flag-123")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected non-zero exit code for unrecognized flag, got exit code 0")
		}
	})

	t.Run("-version standard Go flag style", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(cliBinaryPath, "-version")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("-version failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected output to start with 'agentic-lens version ', got: %s", out)
		}
	})

	t.Run("root flags: -v after --workspaces outputs version and exits cleanly", func(t *testing.T) {
		t.Parallel()
		wsDir := t.TempDir()
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "-v")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("--workspaces ... -v failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected version output when -v passed after workspaces flag, got: %s", out)
		}
	})

	t.Run("root flags: --version after --workspaces outputs version and exits cleanly", func(t *testing.T) {
		t.Parallel()
		wsDir := t.TempDir()
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--version")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("--workspaces ... --version failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "agentic-lens version ") {
			t.Errorf("expected version output when --version passed after workspaces flag, got: %s", out)
		}
	})
}

func TestCLIMigrateCommand(t *testing.T) {
	wsDir := setupTestWorkspace(t)
	originals := make(map[string][]byte)
	for _, id := range []string{"TICKET-1", "TICKET-2"} {
		path := filepath.Join(wsDir, id, "agentic_metrics.json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var ledger map[string]json.RawMessage
		if err := json.Unmarshal(data, &ledger); err != nil {
			t.Fatal(err)
		}
		version, _ := json.Marshal(3)
		ledger["schema_version"] = version
		data, err = json.Marshal(ledger)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		originals[id] = append([]byte(nil), data...)
	}

	t.Run("dry-run preview outputs table without mutating files", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--dry-run",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("migrate dry-run failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "UNIVERSAL SCHEMA MIGRATION [DRY-RUN PREVIEW]") {
			t.Errorf("expected dry-run header, got: %s", out)
		}
		if !strings.Contains(out, "TICKET-1") || !strings.Contains(out, "TICKET-2") {
			t.Errorf("expected tickets in output, got: %s", out)
		}
		if !strings.Contains(out, "Summary: Scanned: 2 | Migrated: 2 | Up to date: 0") {
			t.Errorf("expected 2 migrated preview in summary, got: %s", out)
		}

		// Verify files are byte-for-byte unchanged by the dry run.
		data, err := os.ReadFile(filepath.Join(wsDir, "TICKET-1", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading file: %v", err)
		}
		if !bytes.Equal(data, originals["TICKET-1"]) {
			t.Errorf("expected file to remain unmutated during dry-run")
		}
	})

	t.Run("migrate with ticket-id filter and backup creates .bak and upgrades schema", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--ticket-id", "TICKET-1",
			"--backup",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("migrate failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "UNIVERSAL SCHEMA MIGRATION [EXECUTION]") {
			t.Errorf("expected execution header, got: %s", out)
		}
		if !strings.Contains(out, "TICKET-1") {
			t.Errorf("expected TICKET-1 in output, got: %s", out)
		}

		// Verify backup file was created
		bakData, err := os.ReadFile(filepath.Join(wsDir, "TICKET-1", "agentic_metrics.json.bak"))
		if err != nil {
			t.Fatalf("expected backup file to exist: %v", err)
		}
		if !bytes.Equal(bakData, originals["TICKET-1"]) {
			t.Errorf("backup file should be the original ledger")
		}

		// Verify migration upgrades schema and omits event rows.
		migratedData, err := os.ReadFile(filepath.Join(wsDir, "TICKET-1", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading migrated file: %v", err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(migratedData, &parsed); err != nil {
			t.Fatalf("failed parsing migrated JSON: %v", err)
		}
		if parsed.SchemaVersion == nil || *parsed.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, parsed.SchemaVersion)
		}
		if parsed.Sessions == nil || len(parsed.Sessions) == 0 || len(parsed.Sessions[0].Events) != 0 {
			t.Errorf("expected session event rows to be omitted")
		}

		// TICKET-2 should remain untouched
		t2Data, err := os.ReadFile(filepath.Join(wsDir, "TICKET-2", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading TICKET-2 file: %v", err)
		}
		if !bytes.Equal(t2Data, originals["TICKET-2"]) {
			t.Errorf("TICKET-2 should not be migrated when filtered by ticket-id")
		}
	})

	t.Run("idempotent migration marks already upgraded ledgers up-to-date", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--ticket-id", "TICKET-1",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("second migrate run failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "up-to-date") {
			t.Errorf("expected status 'up-to-date', got: %s", out)
		}
		if !strings.Contains(out, "Migrated: 0 | Up to date: 1") {
			t.Errorf("expected summary with 1 up to date, got: %s", out)
		}
	})
}

func TestCLIBackfillMigrateSchemaFlag(t *testing.T) {
	wsDir := setupTestWorkspace(t)

	cmd := exec.Command(
		cliBinaryPath,
		"backfill",
		"--source", "opencode",
		"--workspaces", wsDir,
		"--migrate-schema",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("backfill with --migrate-schema failed: %v, stderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "Migrate Schema:              true") {
		t.Errorf("expected Migrate Schema: true in header, got: %s", out)
	}
	if !strings.Contains(out, "Schema Migration:") {
		t.Errorf("expected Schema Migration section in output, got: %s", out)
	}

	// Verify both tickets have schema_version upgraded
	for _, tid := range []string{"TICKET-1", "TICKET-2"} {
		data, err := os.ReadFile(filepath.Join(wsDir, tid, "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading %s: %v", tid, err)
		}
		var parsed adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed parsing %s: %v", tid, err)
		}
		if parsed.SchemaVersion == nil || *parsed.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected %s schema_version %d, got %v", tid, domain.CurrentSchemaVersion, parsed.SchemaVersion)
		}
	}
}

func TestCLIMigrateCommand_Adversarial(t *testing.T) {
	t.Run("fails on non-existent workspaces directory", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", "/non/existent/path/for/sure/workspace",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected migrate command to exit with error for non-existent workspaces")
		}
		if !strings.Contains(stderr.String(), "Migration failed") {
			t.Errorf("expected 'Migration failed' in stderr, got: %s", stderr.String())
		}
	})

	t.Run("empty workspaces directory displays clean message", func(t *testing.T) {
		emptyDir := t.TempDir()
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", emptyDir,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("migrate failed on empty dir: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "No matching ticket workspaces found to migrate") {
			t.Errorf("expected 'No matching ticket workspaces found to migrate', got: %s", out)
		}
	})

	t.Run("filtering for non-existent ticket ID displays clean message", func(t *testing.T) {
		wsDir := setupTestWorkspace(t)
		cmd := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--ticket-id", "NON-EXISTENT-TICKET",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("migrate failed with filter: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "No matching ticket workspaces found to migrate") {
			t.Errorf("expected clean message for unmatching ticket filter, got: %s", out)
		}
	})

	t.Run("update summaries is strictly idempotent across runs", func(t *testing.T) {
		root := t.TempDir()
		wsDir := filepath.Join(root, "workspaces")
		t1Dir := filepath.Join(wsDir, "T-SUM")
		_ = os.MkdirAll(t1Dir, 0755)

		t1JSON := `{
  "ticket_id": "T-SUM",
  "ticket_title": "Summary Test",
  "updated_at": "2026-10-18T10:00:00Z",
  "code": {
    "lines_added": 20,
    "sdlc_iterations": 1
  },
  "sessions": []
}`
		_ = os.WriteFile(filepath.Join(t1Dir, "agentic_metrics.json"), []byte(t1JSON), 0644)
		_ = os.WriteFile(filepath.Join(t1Dir, "SUMMARY.md"), []byte("# T-SUM Summary\nInitial description.\n"), 0644)

		// Pass 1: Migrate and update summary
		cmd1 := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--update-summaries",
		)
		var s1, e1 bytes.Buffer
		cmd1.Stdout = &s1
		cmd1.Stderr = &e1
		if err := cmd1.Run(); err != nil {
			t.Fatalf("pass 1 failed: %v, stderr: %s", err, e1.String())
		}

		sum1, err := os.ReadFile(filepath.Join(t1Dir, "SUMMARY.md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(sum1), "### Workflow Efficiency") {
			t.Errorf("expected SUMMARY.md to contain workflow efficiency table")
		}

		// Pass 2: Re-run
		cmd2 := exec.Command(
			cliBinaryPath,
			"migrate",
			"--workspaces", wsDir,
			"--update-summaries",
		)
		var s2, e2 bytes.Buffer
		cmd2.Stdout = &s2
		cmd2.Stderr = &e2
		if err := cmd2.Run(); err != nil {
			t.Fatalf("pass 2 failed: %v, stderr: %s", err, e2.String())
		}

		sum2, err := os.ReadFile(filepath.Join(t1Dir, "SUMMARY.md"))
		if err != nil {
			t.Fatal(err)
		}
		if string(sum1) != string(sum2) {
			t.Errorf("SUMMARY.md modified during second idempotent run")
		}
	})
}

func TestCLIHelpCommand(t *testing.T) {
	t.Run("root --help outputs structured subcommands and root flags", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "--help")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected --help to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		expectedStrings := []string{
			"AgenticLens Developer Telemetry and Observability CLI",
			"Usage:",
			"agentic-lens <subcommand> [flags]",
			"server",
			"generate",
			"backfill",
			"migrate",
			"version",
			"--workspaces",
			"--ticket-id",
			"--list",
			"--anomalies",
			"--append",
			"--json-output",
			"agentic-lens <subcommand> --help",
		}
		for _, exp := range expectedStrings {
			if !strings.Contains(out, exp) {
				t.Errorf("expected help output to contain %q, but was missing. Output:\n%s", exp, out)
			}
		}
	})

	t.Run("root -h outputs structured subcommands", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "-h")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected -h to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		for _, sub := range []string{"server", "generate", "backfill", "migrate", "version"} {
			if !strings.Contains(out, sub) {
				t.Errorf("expected -h output to contain subcommand %q, but was missing", sub)
			}
		}
	})

	t.Run("root help subcommand outputs structured subcommands", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "help")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected help to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		for _, sub := range []string{"server", "generate", "backfill", "migrate", "version"} {
			if !strings.Contains(out, sub) {
				t.Errorf("expected help output to contain subcommand %q, but was missing", sub)
			}
		}
	})

	t.Run("root -help outputs structured subcommands", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "-help")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected -help to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		for _, sub := range []string{"server", "generate", "backfill", "migrate", "version"} {
			if !strings.Contains(out, sub) {
				t.Errorf("expected -help output to contain subcommand %q, but was missing", sub)
			}
		}
	})

	t.Run("backfill --help lists all backfill flags", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "backfill", "--help")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected backfill --help to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		expectedFlags := []string{
			"--migrate-schema",
			"--recompute-git",
			"--recompute-db",
			"--dry-run",
			"--update-summaries",
			"--force",
			"--workspaces",
			"--ticket-id",
			"--db-path",
			"--code-ref",
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(out, flag) {
				t.Errorf("expected backfill --help output to contain %q, but was missing. Output:\n%s", flag, out)
			}
		}
	})

	t.Run("backfill -h and backfill help list all backfill flags", func(t *testing.T) {
		for _, invocation := range [][]string{{"backfill", "-h"}, {"backfill", "help"}} {
			cmd := exec.Command(cliBinaryPath, invocation...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("expected %v to exit cleanly, got error: %v, stderr: %s", invocation, err, stderr.String())
			}

			out := stdout.String() + stderr.String()
			for _, flag := range []string{"--migrate-schema", "--recompute-git", "--recompute-db", "--dry-run"} {
				if !strings.Contains(out, flag) {
					t.Errorf("expected %v output to contain %q", invocation, flag)
				}
			}
		}
	})

	t.Run("migrate --help lists all migrate flags", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "migrate", "--help")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("expected migrate --help to exit cleanly, got error: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String() + stderr.String()
		expectedFlags := []string{
			"--workspaces",
			"--ticket-id",
			"--target-version",
			"--dry-run",
			"--backup",
			"--update-summaries",
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(out, flag) {
				t.Errorf("expected migrate --help output to contain %q, but was missing. Output:\n%s", flag, out)
			}
		}
	})

	t.Run("migrate -h and migrate help list all migrate flags", func(t *testing.T) {
		for _, invocation := range [][]string{{"migrate", "-h"}, {"migrate", "help"}} {
			cmd := exec.Command(cliBinaryPath, invocation...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("expected %v to exit cleanly, got error: %v, stderr: %s", invocation, err, stderr.String())
			}

			out := stdout.String() + stderr.String()
			for _, flag := range []string{"--target-version", "--dry-run", "--backup", "--update-summaries"} {
				if !strings.Contains(out, flag) {
					t.Errorf("expected %v output to contain %q", invocation, flag)
				}
			}
		}
	})

	t.Run("help <subcommand> delegates to subcommand help", func(t *testing.T) {
		cmdMigrate := exec.Command(cliBinaryPath, "help", "migrate")
		var outMigrate bytes.Buffer
		cmdMigrate.Stdout = &outMigrate
		if err := cmdMigrate.Run(); err != nil {
			t.Fatalf("expected help migrate to exit cleanly: %v", err)
		}
		if !strings.Contains(outMigrate.String(), "--target-version") {
			t.Errorf("expected help migrate to list --target-version")
		}

		cmdBackfill := exec.Command(cliBinaryPath, "help", "backfill")
		var outBackfill bytes.Buffer
		cmdBackfill.Stdout = &outBackfill
		if err := cmdBackfill.Run(); err != nil {
			t.Fatalf("expected help backfill to exit cleanly: %v", err)
		}
		if !strings.Contains(outBackfill.String(), "--migrate-schema") {
			t.Errorf("expected help backfill to list --migrate-schema")
		}
	})
}

func TestGenerate_JSONRawVsMarkdownMappedSeparation(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Setup SQLite database with session agent 'build'
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE session (id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
		CREATE TABLE message (id TEXT PRIMARY KEY, session_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		CREATE TABLE part (id TEXT PRIMARY KEY, session_id TEXT, message_id TEXT, time_created INTEGER, time_updated INTEGER, data TEXT);
		INSERT INTO session VALUES ('ses_root_raw', NULL, 'build', 'claude-3-sonnet', 'Task AL-24 Feature', 1000, 2000);
		INSERT INTO message VALUES ('msg_1', 'ses_root_raw', 1100, 1900, '{"role":"assistant","tokens":{"input":1000,"output":200},"cost":0.05}');
	`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()

	outJSON := filepath.Join(tempDir, "agentic_metrics.json")
	summaryMD := filepath.Join(tempDir, "SUMMARY.md")

	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--ticket-id", "AL-24",
		"--ticket-title", "Test Fixture Ticket",
		"--json-output", outJSON,
		"--append", summaryMD,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
	}

	// 2. Parse generated JSON: raw invocation role must be 'build'
	jsonData, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("reading JSON output failed: %v", err)
	}
	var rawLedger domain.AgenticFactLedger
	if err := json.Unmarshal(jsonData, &rawLedger); err != nil {
		t.Fatalf("unmarshaling JSON failed: %v", err)
	}
	if len(rawLedger.InvocationFacts()) != 1 {
		t.Fatalf("expected 1 invocation in raw facts, got %d", len(rawLedger.InvocationFacts()))
	}
	if rawLedger.InvocationFacts()[0].Role != "build" {
		t.Errorf("expected raw invocation role 'build', got %q", rawLedger.InvocationFacts()[0].Role)
	}

	// 3. Read summary markdown: table must contain orchestrator and NO row for build
	mdData, err := os.ReadFile(summaryMD)
	if err != nil {
		t.Fatalf("reading summary markdown failed: %v", err)
	}
	mdContent := string(mdData)

	if !strings.Contains(mdContent, "| **orchestrator** |") {
		t.Errorf("expected summary markdown to contain '| **orchestrator** |', got:\n%s", mdContent)
	}
	if strings.Contains(mdContent, "| **build** |") {
		t.Errorf("expected summary markdown NOT to contain '| **build** |', got:\n%s", mdContent)
	}
}

func TestCLITicketTypeClassification(t *testing.T) {
	t.Parallel()

	t.Run("explicit ticket-type alias feature normalizes to feat in JSON ledger", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-29",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "feature",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType, got nil")
		}
		if *parsed.TicketType != "feat" {
			t.Errorf("expected normalized TicketType 'feat', got %q", *parsed.TicketType)
		}
	})

	t.Run("explicit invalid ticket-type exits with status 1 and lists valid types", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-29",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "invalid-type",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate to fail with exit status 1 for invalid ticket type, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: invalid ticket type 'invalid-type'. Must be one of: feat, fix, docs, chore, refactor, test, ci, perf, build"
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}

		if _, err := os.Stat(outJSON); !os.IsNotExist(err) {
			t.Fatalf("expected output JSON not to exist, but was created: %s", outJSON)
		}
	})

	t.Run("automatic resolution of PO ticket records docs without flag", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "PO-5",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for PO-5, got nil")
		}
		if *parsed.TicketType != "docs" {
			t.Errorf("expected TicketType 'docs' for PO-5, got %q", *parsed.TicketType)
		}
	})

	t.Run("automatic resolution of GEN ticket records chore without flag", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "GEN-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for GEN-1, got nil")
		}
		if *parsed.TicketType != "chore" {
			t.Errorf("expected TicketType 'chore' for GEN-1, got %q", *parsed.TicketType)
		}
	})

	t.Run("unmatched ticket without strict-type logs warning and records ticket_type null", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "UNMATCHED-9999",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		expectedWarn := "WARNING: Missing ticket type for ticket UNMATCHED-9999. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedWarn) {
			t.Fatalf("expected stderr to contain warning:\n%s\ngot:\n%s", expectedWarn, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType != nil {
			t.Errorf("expected TicketType to be nil, got %q", *parsed.TicketType)
		}

		// Also verify raw JSON contains "ticket_type": null
		if !strings.Contains(string(jsonData), `"ticket_type": null`) {
			t.Errorf("expected raw JSON to contain '\"ticket_type\": null', got:\n%s", string(jsonData))
		}
	})

	t.Run("unmatched ticket with strict-type exits with status 1 and logs error", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "UNMATCHED-9999",
			"--ticket-title", "Test Fixture Ticket",
			"--strict-type",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate with --strict-type on unmatched ticket to fail, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: missing ticket type for ticket 'UNMATCHED-9999'. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected stderr to contain error:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}

		if _, err := os.Stat(outJSON); !os.IsNotExist(err) {
			t.Fatalf("expected output JSON not to exist on strict failure, but found: %s", outJSON)
		}
	})

	t.Run("cached facts with alias feature normalizes to feat when flag omitted", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		// Create existing fact ledger with legacy "feature"
		legacyJSON := `{
  "schema_version": 4,
  "migration": {"schema_version": 4, "status": "completed", "steps_completed": ["source_backfill"], "steps_pending": []},
  "ticket_id": "AL-29",
  "ticket_title": "Test Fixture Ticket",
  "ticket_type": "feature",
  "code": {
    "lines_added": 10,
    "lines_deleted": 2,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": []
}`
		if err := os.WriteFile(outJSON, []byte(legacyJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-29",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType, got nil")
		}
		if *parsed.TicketType != "feat" {
			t.Errorf("expected normalized TicketType 'feat', got %q", *parsed.TicketType)
		}
	})

	t.Run("explicit ticket-type with whitespace and mixed cases normalizes correctly", func(t *testing.T) {
		t.Parallel()
		mixedCases := []struct {
			input    string
			expected domain.ConventionalTicketType
		}{
			{input: "  FeAt  ", expected: domain.TicketTypeFeat},
			{input: "BUGFIX", expected: domain.TicketTypeFix},
			{input: "DoCs", expected: domain.TicketTypeDocs},
			{input: "  cI  ", expected: domain.TicketTypeCI},
			{input: "PeRfOrMaNcE", expected: domain.TicketTypePerf},
			{input: "  bUiLd  ", expected: domain.TicketTypeBuild},
		}

		for _, tc := range mixedCases {
			tc := tc
			t.Run(tc.input, func(t *testing.T) {
				t.Parallel()
				tempDir := t.TempDir()
				outJSON := filepath.Join(tempDir, "facts.json")

				cmd := exec.Command(
					cliBinaryPath,
					"generate",
					"--db-path", setupIsolatedEmptyDB(t),
					"--ticket-id", "AL-29",
					"--ticket-title", "Test Fixture Ticket",
					"--ticket-type", tc.input,
					"--allow-empty",
					"--json-output", outJSON,
				)
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				if err := cmd.Run(); err != nil {
					t.Fatalf("generate failed for input %q: %v, stderr: %s", tc.input, err, stderr.String())
				}

				jsonData, err := os.ReadFile(outJSON)
				if err != nil {
					t.Fatalf("failed reading output JSON: %v", err)
				}

				var parsed domain.AgenticFactLedger
				if err := json.Unmarshal(jsonData, &parsed); err != nil {
					t.Fatalf("failed unmarshaling output JSON: %v", err)
				}

				if parsed.TicketType == nil {
					t.Fatalf("expected non-nil TicketType for input %q, got nil", tc.input)
				}
				if *parsed.TicketType != tc.expected {
					t.Errorf("for input %q, expected %q, got %q", tc.input, tc.expected, *parsed.TicketType)
				}
			})
		}
	})

	t.Run("false prefix POD-1 does NOT match PO- and requires explicit type", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		// Without strict-type: emits warning and writes ticket_type null
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "POD-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		expectedWarn := "WARNING: Missing ticket type for ticket POD-1. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedWarn) {
			t.Fatalf("expected stderr to contain warning:\n%s\ngot:\n%s", expectedWarn, stderr.String())
		}
		if strings.Contains(stderr.String(), "docs") {
			t.Fatalf("stderr should not mention docs for false prefix POD-1, got: %s", stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType != nil {
			t.Fatalf("expected TicketType nil for POD-1, got %q", *parsed.TicketType)
		}

		// With strict-type: terminates with exit status 1
		strictJSON := filepath.Join(tempDir, "strict.json")
		strictCmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "POD-1",
			"--ticket-title", "Test Fixture Ticket",
			"--strict-type",
			"--allow-empty",
			"--json-output", strictJSON,
		)
		var strictOut, strictErr bytes.Buffer
		strictCmd.Stdout = &strictOut
		strictCmd.Stderr = &strictErr

		err = strictCmd.Run()
		if err == nil {
			t.Fatal("expected generate with --strict-type on POD-1 to fail, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: missing ticket type for ticket 'POD-1'. Specify --ticket-type explicitly."
		if !strings.Contains(strictErr.String(), expectedErr) {
			t.Fatalf("expected strict error:\n%s\ngot:\n%s", expectedErr, strictErr.String())
		}
	})

	t.Run("false prefix GENTLE-1 does NOT match GEN- and requires explicit type", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		// Without strict-type: emits warning and writes ticket_type null
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "GENTLE-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		expectedWarn := "WARNING: Missing ticket type for ticket GENTLE-1. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedWarn) {
			t.Fatalf("expected stderr to contain warning:\n%s\ngot:\n%s", expectedWarn, stderr.String())
		}
		if strings.Contains(stderr.String(), "chore") {
			t.Fatalf("stderr should not mention chore for false prefix GENTLE-1, got: %s", stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType != nil {
			t.Fatalf("expected TicketType nil for GENTLE-1, got %q", *parsed.TicketType)
		}

		// With strict-type: terminates with exit status 1
		strictJSON := filepath.Join(tempDir, "strict.json")
		strictCmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "GENTLE-1",
			"--ticket-title", "Test Fixture Ticket",
			"--strict-type",
			"--allow-empty",
			"--json-output", strictJSON,
		)
		var strictOut, strictErr bytes.Buffer
		strictCmd.Stdout = &strictOut
		strictCmd.Stderr = &strictErr

		err = strictCmd.Run()
		if err == nil {
			t.Fatal("expected generate with --strict-type on GENTLE-1 to fail, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: missing ticket type for ticket 'GENTLE-1'. Specify --ticket-type explicitly."
		if !strings.Contains(strictErr.String(), expectedErr) {
			t.Fatalf("expected strict error:\n%s\ngot:\n%s", expectedErr, strictErr.String())
		}
	})

	t.Run("strict-type with missing ticket-id uses UNKNOWN fallback in error message", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--strict-type",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate with --strict-type and missing ticket-id to fail, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: missing ticket ID. Specify --ticket-id explicitly."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected strict error:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}
	})

	t.Run("strict-type with whitespace-only ticket-type fails on unmatched ticket", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "UNMATCHED-9999",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "   \t",
			"--strict-type",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate with whitespace ticket-type and --strict-type to fail, but succeeded")
		}

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		} else {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}

		expectedErr := "Error: missing ticket type for ticket 'UNMATCHED-9999'. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedErr) {
			t.Fatalf("expected strict error:\n%s\ngot:\n%s", expectedErr, stderr.String())
		}
	})

	t.Run("strict-type succeeds when explicit ticket-type is provided", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "UNMATCHED-9999",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "feat",
			"--strict-type",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate with --strict-type and valid --ticket-type failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil || *parsed.TicketType != "feat" {
			t.Errorf("expected TicketType 'feat', got %v", parsed.TicketType)
		}
	})

	t.Run("unrecognized CLI flag exits with non-zero status", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--completely-unknown-adversarial-flag",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate with unrecognized flag to fail, but succeeded")
		}
	})

	t.Run("custom prefix resolves automatically from project config", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"CUSTOM\"\n      ticket_type: \"fix\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "CUSTOM-101",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for CUSTOM-101, got nil")
		}
		if *parsed.TicketType != "fix" {
			t.Errorf("expected TicketType 'fix' for CUSTOM-101, got %q", *parsed.TicketType)
		}
	})

	t.Run("configured prefix overrides default prefix mapping", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"PO\"\n      ticket_type: \"feat\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "PO-88",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for PO-88, got nil")
		}
		if *parsed.TicketType != "feat" {
			t.Errorf("expected overridden TicketType 'feat' for PO-88, got %q", *parsed.TicketType)
		}
	})

	t.Run("explicit CLI ticket-type flag overrides configured prefix", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"CUSTOM\"\n      ticket_type: \"fix\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "CUSTOM-101",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "chore",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for CUSTOM-101, got nil")
		}
		if *parsed.TicketType != "chore" {
			t.Errorf("expected overridden TicketType 'chore' for CUSTOM-101, got %q", *parsed.TicketType)
		}
	})

	t.Run("strict-type succeeds when custom prefix matches in project config", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"HOTFIX\"\n      ticket_type: \"fix\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "HOTFIX-77",
			"--ticket-title", "Test Fixture Ticket",
			"--strict-type",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed with --strict-type: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for HOTFIX-77, got nil")
		}
		if *parsed.TicketType != "fix" {
			t.Errorf("expected TicketType 'fix' for HOTFIX-77, got %q", *parsed.TicketType)
		}
	})

	t.Run("Tier 1 explicit flag overrides Tier 2 git branch prefix", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		// Initialize tempDir as an isolated git repository with a branch fix/AL-30-test
		if err := exec.Command("git", "init", "-b", "main", tempDir).Run(); err != nil {
			t.Fatalf("git init failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run(); err != nil {
			t.Fatalf("git config user.email failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "config", "user.name", "Test Agent").Run(); err != nil {
			t.Fatalf("git config user.name failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "checkout", "-b", "fix/AL-30-test").Run(); err != nil {
			t.Fatalf("git checkout -b failed: %v", err)
		}

		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"AL\"\n      ticket_type: \"docs\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")

		// AL-30 is on branch fix/AL-30-test. Passing --ticket-type chore MUST override the branch prefix "fix".
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-30",
			"--ticket-title", "Test Fixture Ticket",
			"--ticket-type", "chore",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for AL-30, got nil")
		}
		if *parsed.TicketType != "chore" {
			t.Errorf("expected explicit flag 'chore' to override branch 'fix', got %q", *parsed.TicketType)
		}
	})

	t.Run("configured project prefix overrides git branch prefix", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		// Initialize tempDir as an isolated git repository with a commit and branch fix/AL-30-test
		if err := exec.Command("git", "init", "-b", "main", tempDir).Run(); err != nil {
			t.Fatalf("git init failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run(); err != nil {
			t.Fatalf("git config user.email failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "config", "user.name", "Test Agent").Run(); err != nil {
			t.Fatalf("git config user.name failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}
		if err := exec.Command("git", "-C", tempDir, "checkout", "-b", "fix/AL-30-test").Run(); err != nil {
			t.Fatalf("git checkout -b failed: %v", err)
		}

		// Configure AL to map to docs in .agenticlens-config.yaml
		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"AL\"\n      ticket_type: \"docs\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing .agenticlens-config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		// Config prefix "docs" takes precedence over branch prefix "fix".
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-30",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for AL-30, got nil")
		}
		if *parsed.TicketType != "docs" {
			t.Errorf("expected configured prefix 'docs' to override branch prefix 'fix', got %q", *parsed.TicketType)
		}
	})
}

func TestCLI_RootConfigurationDiscovery(t *testing.T) {
	t.Parallel()

	t.Run("generate and scan pick up workspaces and prefixes from root .agenticlens-config.yaml", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		customWsDir := filepath.Join(tempDir, "custom_root_ws")
		ticketWsDir := filepath.Join(customWsDir, "MYROOT-5")
		if err := os.MkdirAll(ticketWsDir, 0755); err != nil {
			t.Fatalf("failed creating ticket dir: %v", err)
		}

		cfgContent := `
project_name: CustomRootTest
ticket:
  prefixes:
    - prefix: "MYROOT"
      ticket_type: "fix"
workspaces:
  dir: "custom_root_ws"
  src_folders:
    - "src"
`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config: %v", err)
		}

		outJSON := filepath.Join(ticketWsDir, "agentic_metrics.json")
		genCmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--ticket-id", "MYROOT-5",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		genCmd.Dir = tempDir
		var genStdout, genStderr bytes.Buffer
		genCmd.Stdout = &genStdout
		genCmd.Stderr = &genStderr

		if err := genCmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, genStderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil || *parsed.TicketType != "fix" {
			t.Errorf("expected TicketType 'fix' from root config, got %v", parsed.TicketType)
		}

		// Verify scan with default workspaces flag resolves custom_root_ws from root config
		scanCmd := exec.Command(
			cliBinaryPath,
			"--list",
		)
		scanCmd.Dir = tempDir
		var scanStdout, scanStderr bytes.Buffer
		scanCmd.Stdout = &scanStdout
		scanCmd.Stderr = &scanStderr

		if err := scanCmd.Run(); err != nil {
			t.Fatalf("scan list failed: %v, stderr: %s", err, scanStderr.String())
		}

		scanOut := scanStdout.String()
		if !strings.Contains(scanOut, "MYROOT-5") {
			t.Errorf("expected MYROOT-5 in scan output, got: %s", scanOut)
		}
	})

	t.Run("root config takes precedence over legacy config folder in CLI", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()

		legacyDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(legacyDir, 0755); err != nil {
			t.Fatalf("failed creating legacy dir: %v", err)
		}

		rootWsDir := filepath.Join(tempDir, "root_ws", "ROOTPRIO-1")
		legacyWsDir := filepath.Join(tempDir, "legacy_ws", "ROOTPRIO-1")
		if err := os.MkdirAll(rootWsDir, 0755); err != nil || os.MkdirAll(legacyWsDir, 0755) != nil {
			t.Fatalf("failed creating ws dirs")
		}

		// Root config maps ROOTPRIO to "feat" and dir to "root_ws"
		rootCfg := `
project_name: RootPriorityTest
ticket:
  prefixes:
    - prefix: "ROOTPRIO"
      ticket_type: "feat"
workspaces:
  dir: "root_ws"
  src_folders:
    - "src"
`
		// Legacy config maps ROOTPRIO to "chore" and dir to "legacy_ws"
		legacyCfg := `
project_name: LegacyPriorityTest
ticket:
  prefixes:
    - prefix: "ROOTPRIO"
      ticket_type: "chore"
workspaces:
  dir: "legacy_ws"
  src_folders:
    - "src"
`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(rootCfg), 0644); err != nil {
			t.Fatalf("failed writing root config: %v", err)
		}
		if err := os.WriteFile(filepath.Join(legacyDir, "config.yaml"), []byte(legacyCfg), 0644); err != nil {
			t.Fatalf("failed writing legacy config: %v", err)
		}

		outJSON := filepath.Join(rootWsDir, "agentic_metrics.json")
		genCmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--ticket-id", "ROOTPRIO-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		genCmd.Dir = tempDir
		var genStdout, genStderr bytes.Buffer
		genCmd.Stdout = &genStdout
		genCmd.Stderr = &genStderr

		if err := genCmd.Run(); err != nil {
			t.Fatalf("generate failed: %v, stderr: %s", err, genStderr.String())
		}

		jsonData, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatalf("failed reading output JSON: %v", err)
		}

		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("failed unmarshaling output JSON: %v", err)
		}

		if parsed.TicketType == nil || *parsed.TicketType != "feat" {
			t.Errorf("expected TicketType 'feat' from root config (overriding legacy chore), got %v", parsed.TicketType)
		}
	})
}

func TestCLIGenerate_TicketMetadataValidation(t *testing.T) {
	t.Parallel()

	t.Run("fails with exit code 1 when ticket-id is missing or whitespace", func(t *testing.T) {
		t.Parallel()

		// Case 1: Flag omitted entirely
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected failure on missing ticket-id")
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
		if !strings.Contains(stderr.String(), "Error: missing ticket ID. Specify --ticket-id explicitly.") {
			t.Fatalf("unexpected stderr: %s", stderr.String())
		}

		// Case 2: Whitespace only
		cmd2 := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "   \t",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
		)
		var stderr2 bytes.Buffer
		cmd2.Stderr = &stderr2
		err2 := cmd2.Run()
		if err2 == nil {
			t.Fatal("expected failure on whitespace ticket-id")
		}
		var exitErr2 *exec.ExitError
		if errors.As(err2, &exitErr2) && exitErr2.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr2.ExitCode())
		}
		if !strings.Contains(stderr2.String(), "Error: missing ticket ID. Specify --ticket-id explicitly.") {
			t.Fatalf("unexpected stderr: %s", stderr2.String())
		}
	})

	t.Run("fails with exit code 1 when ticket-title is missing or whitespace", func(t *testing.T) {
		t.Parallel()

		// Case 1: Flag omitted entirely
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-TITLE-REQ",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err == nil {
			t.Fatal("expected failure on missing ticket-title")
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
		if !strings.Contains(stderr.String(), "Error: missing ticket title for ticket 'AL-TITLE-REQ'. Specify --ticket-title explicitly.") {
			t.Fatalf("unexpected stderr: %s", stderr.String())
		}

		// Case 2: Whitespace only
		cmd2 := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-TITLE-REQ",
			"--ticket-title", "   \n\t",
			"--allow-empty",
		)
		var stderr2 bytes.Buffer
		cmd2.Stderr = &stderr2
		err2 := cmd2.Run()
		if err2 == nil {
			t.Fatal("expected failure on whitespace ticket-title")
		}
		var exitErr2 *exec.ExitError
		if errors.As(err2, &exitErr2) && exitErr2.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr2.ExitCode())
		}
		if !strings.Contains(stderr2.String(), "Error: missing ticket title for ticket 'AL-TITLE-REQ'. Specify --ticket-title explicitly.") {
			t.Fatalf("unexpected stderr: %s", stderr2.String())
		}

		// Case 3: Cached facts has whitespace-only title and flag omitted
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts_ws_title.json")
		wsCachedJSON := `{
  "schema_version": 4,
  "migration": {"schema_version": 4, "status": "completed", "steps_completed": ["source_backfill"], "steps_pending": []},
  "ticket_id": "AL-CACHED-WS-TITLE",
  "ticket_title": "   \t\n",
  "sessions": []
}`
		if err := os.WriteFile(outJSON, []byte(wsCachedJSON), 0644); err != nil {
			t.Fatal(err)
		}
		cmd3 := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-CACHED-WS-TITLE",
			"--json-output", outJSON,
			"--allow-empty",
		)
		var stderr3 bytes.Buffer
		cmd3.Stderr = &stderr3
		err3 := cmd3.Run()
		if err3 == nil {
			t.Fatal("expected failure when cached title is whitespace only")
		}
		var exitErr3 *exec.ExitError
		if errors.As(err3, &exitErr3) && exitErr3.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr3.ExitCode())
		}
		if !strings.Contains(stderr3.String(), "Error: missing ticket title for ticket 'AL-CACHED-WS-TITLE'. Specify --ticket-title explicitly.") {
			t.Fatalf("unexpected stderr: %s", stderr3.String())
		}
	})

	t.Run("cached title fallback succeeds when ticket-title flag is omitted", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")
		cachedJSON := `{
  "schema_version": 4,
  "migration": {"schema_version": 4, "status": "completed", "steps_completed": ["source_backfill"], "steps_pending": []},
  "ticket_id": "AL-CACHED-TITLE",
  "ticket_title": "Cached Fixture Title",
  "sessions": []
}`
		if err := os.WriteFile(outJSON, []byte(cachedJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-CACHED-TITLE",
			"--json-output", outJSON,
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("expected cached title fallback to succeed, got: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.TicketTitle == nil || *parsed.TicketTitle != "Cached Fixture Title" {
			t.Fatalf("expected cached title 'Cached Fixture Title', got %v", parsed.TicketTitle)
		}

		// Also verify explicit --ticket-title flag overrides the cached title
		cmdOverride := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-CACHED-TITLE",
			"--ticket-title", "Overridden Ticket Title",
			"--json-output", outJSON,
			"--allow-empty",
		)
		var stdoutO, stderrO bytes.Buffer
		cmdOverride.Stdout = &stdoutO
		cmdOverride.Stderr = &stderrO
		if err := cmdOverride.Run(); err != nil {
			t.Fatalf("expected explicit title override to succeed, got: %v, stderr: %s", err, stderrO.String())
		}
		dataO, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var parsedO domain.AgenticFactLedger
		if err := json.Unmarshal(dataO, &parsedO); err != nil {
			t.Fatal(err)
		}
		if parsedO.TicketTitle == nil || *parsedO.TicketTitle != "Overridden Ticket Title" {
			t.Fatalf("expected overridden title 'Overridden Ticket Title', got %v", parsedO.TicketTitle)
		}
	})

	t.Run("ticket-type warning issued on stderr for unmatched ticket in non-strict mode", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "UNMATCHED-WARNING",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("expected command to exit 0 in non-strict mode, got: %v, stderr: %s", err, stderr.String())
		}
		expectedWarn := "WARNING: Missing ticket type for ticket UNMATCHED-WARNING. Specify --ticket-type explicitly."
		if !strings.Contains(stderr.String(), expectedWarn) {
			t.Fatalf("expected warning on stderr:\n%s\ngot:\n%s", expectedWarn, stderr.String())
		}
	})

	t.Run("empty invocations warning issued on stderr when invocations count is zero", func(t *testing.T) {
		t.Parallel()
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-EMPTY-INV",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("expected generate with --allow-empty to succeed, got: %v, stderr: %s", err, stderr.String())
		}
		expectedWarn := "WARNING: Zero invocations recorded for ticket AL-EMPTY-INV. This is not expected."
		if !strings.Contains(stderr.String(), expectedWarn) {
			t.Fatalf("expected warning on stderr:\n%s\ngot:\n%s", expectedWarn, stderr.String())
		}
	})

	t.Run("FIX prefix infers ticket_type fix from root .agenticlens-config.yaml", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "facts.json")

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "FIX-1",
			"--ticket-title", "Bug Fix Telemetry Metadata",
			"--allow-empty",
			"--json-output", outJSON,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate failed for FIX-1: %v, stderr: %s", err, stderr.String())
		}

		data, err := os.ReadFile(outJSON)
		if err != nil {
			t.Fatal(err)
		}
		var parsed domain.AgenticFactLedger
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatal(err)
		}
		if parsed.TicketType == nil {
			t.Fatal("expected non-nil TicketType for FIX-1, got nil")
		}
		if *parsed.TicketType != "fix" {
			t.Errorf("expected TicketType 'fix' for FIX-1, got %q", *parsed.TicketType)
		}

		// Also verify case-insensitive prefix inference for lowercase fix-42
		outJSONLower := filepath.Join(tempDir, "facts_lower.json")
		cmdLower := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "fix-42",
			"--ticket-title", "Lowercase Fix Title",
			"--allow-empty",
			"--json-output", outJSONLower,
		)
		var stdoutL, stderrL bytes.Buffer
		cmdLower.Stdout = &stdoutL
		cmdLower.Stderr = &stderrL
		if err := cmdLower.Run(); err != nil {
			t.Fatalf("generate failed for fix-42: %v, stderr: %s", err, stderrL.String())
		}
		dataL, err := os.ReadFile(outJSONLower)
		if err != nil {
			t.Fatal(err)
		}
		var parsedL domain.AgenticFactLedger
		if err := json.Unmarshal(dataL, &parsedL); err != nil {
			t.Fatal(err)
		}
		if parsedL.TicketType == nil || *parsedL.TicketType != "fix" {
			t.Errorf("expected TicketType 'fix' for fix-42, got %v", parsedL.TicketType)
		}
	})
}

func TestCLI_Invariants(t *testing.T) {
	t.Run("fatal on unresolvable base branch", func(t *testing.T) {
		tempDir := t.TempDir()

		// Set up repo with non-standard branch name and no origin remote
		runCmd := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = tempDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
			}
		}
		runCmd("init", "-b", "dev-feature")
		runCmd("config", "user.name", "TestUser")
		runCmd("config", "user.email", "[EMAIL_REDACTED]")
		readmePath := filepath.Join(tempDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		runCmd("add", "README.md")
		runCmd("commit", "-m", "initial commit")

		cfgContent := "project_name: \"TestProject\"\nticket:\n  prefixes:\n    - prefix: \"DEV\"\nworkspaces:\n  src_folders:\n    - \"src\"\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--ticket-id", "DEV-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate command to fail with unresolvable base branch, got exit code 0")
		}

		expectedPrompt := "unable to resolve upstream base branch: configure 'code_metrics.base_branch' in .agenticlens-config.yaml"
		if !strings.Contains(stderr.String(), expectedPrompt) {
			t.Errorf("expected stderr to contain %q, got %q", expectedPrompt, stderr.String())
		}
	})

	t.Run("fatal on empty src_folders", func(t *testing.T) {
		tempDir := t.TempDir()

		runCmd := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = tempDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
			}
		}
		runCmd("init", "-b", "main")
		runCmd("config", "user.name", "TestUser")
		runCmd("config", "user.email", "[EMAIL_REDACTED]")
		readmePath := filepath.Join(tempDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
			t.Fatal(err)
		}
		runCmd("add", "README.md")
		runCmd("commit", "-m", "initial commit")

		cfgContent := "workspaces:\n  src_folders: []\n"
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatal(err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--ticket-id", "DEV-1",
			"--ticket-title", "Test Fixture Ticket",
			"--allow-empty",
			"--json-output", outJSON,
		)
		cmd.Dir = tempDir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected generate command to fail with empty src_folders, got exit code 0")
		}

		expectedPrompt := "workspaces.src_folders is empty or unconfigured: configure 'workspaces.src_folders' in .agenticlens-config.yaml"
		if !strings.Contains(stderr.String(), expectedPrompt) {
			t.Errorf("expected stderr to contain %q, got %q", expectedPrompt, stderr.String())
		}
	})
}
func TestInitCommand(t *testing.T) {
	t.Run("missing required flag --project-name fails with exit code 1", func(t *testing.T) {
		variations := [][]string{
			{"init", "--src-folders", "src", "--ticket-prefixes", "AL"},
			{"init", "--project-name", "", "--src-folders", "src", "--ticket-prefixes", "AL"},
			{"init", "--project-name", "   ", "--src-folders", "src", "--ticket-prefixes", "AL"},
		}

		for _, args := range variations {
			tempDir := t.TempDir()
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = tempDir
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected init to fail without --project-name (args: %v), got nil error", args)
			}
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d for args: %v", exitErr.ExitCode(), args)
			}
			if !strings.Contains(stderr.String(), "Error: missing required flag --project-name") {
				t.Errorf("expected stderr to contain 'Error: missing required flag --project-name', got: %s", stderr.String())
			}
		}
	})

	t.Run("missing required flag --src-folders fails with exit code 1", func(t *testing.T) {
		variations := [][]string{
			{"init", "--project-name", "Acme", "--ticket-prefixes", "AL"},
			{"init", "--project-name", "Acme", "--src-folders", "", "--ticket-prefixes", "AL"},
			{"init", "--project-name", "Acme", "--src-folders", "   ,  ", "--ticket-prefixes", "AL"},
		}

		for _, args := range variations {
			tempDir := t.TempDir()
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = tempDir
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected init to fail without --src-folders (args: %v), got nil error", args)
			}
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d for args: %v", exitErr.ExitCode(), args)
			}
			if !strings.Contains(stderr.String(), "Error: missing required flag --src-folders") {
				t.Errorf("expected stderr to contain 'Error: missing required flag --src-folders', got: %s", stderr.String())
			}
		}
	})

	t.Run("missing required flag --ticket-prefixes fails with exit code 1", func(t *testing.T) {
		variations := [][]string{
			{"init", "--project-name", "Acme", "--src-folders", "src"},
			{"init", "--project-name", "Acme", "--src-folders", "src", "--ticket-prefixes", ""},
			{"init", "--project-name", "Acme", "--src-folders", "src", "--ticket-prefixes", "   "},
		}

		for _, args := range variations {
			tempDir := t.TempDir()
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = tempDir
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err == nil {
				t.Fatalf("expected init to fail without --ticket-prefixes (args: %v), got nil error", args)
			}
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d for args: %v", exitErr.ExitCode(), args)
			}
			if !strings.Contains(stderr.String(), "Error: missing required flag --ticket-prefixes") {
				t.Errorf("expected stderr to contain 'Error: missing required flag --ticket-prefixes', got: %s", stderr.String())
			}
		}
	})

	t.Run("missing all required flags fails with exit code 1", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := exec.Command(cliBinaryPath, "init")
		cmd.Dir = tempDir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected init without any flags to fail, got nil error")
		}
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
		}
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
		if !strings.Contains(stderr.String(), "Error: missing required flag") {
			t.Errorf("expected stderr to contain missing flag error, got: %s", stderr.String())
		}
	})

	t.Run("valid required flags create valid configuration file", func(t *testing.T) {
		tempDir := t.TempDir()

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "AcmeApp",
			"--src-folders", "web,backend",
			"--ticket-prefixes", "ACME",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("agentic-lens init failed: %v, stderr: %s", err, stderr.String())
		}

		expectedPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if _, err := os.Stat(expectedPath); err != nil {
			t.Fatalf("expected config file %s to exist: %v", expectedPath, err)
		}

		loaded, err := adaptersconfig.LoadFromFile(expectedPath)
		if err != nil {
			t.Fatalf("loading generated config failed: %v", err)
		}

		if loaded.ProjectName != "AcmeApp" {
			t.Errorf("expected ProjectName 'AcmeApp', got %q", loaded.ProjectName)
		}
		if loaded.Workspaces.Dir != "workspaces" {
			t.Errorf("expected workspaces.dir 'workspaces', got %q", loaded.Workspaces.Dir)
		}
		if len(loaded.Workspaces.Artifacts) != 0 {
			t.Errorf("expected generated config to have zero workspace artifacts by default, got %d", len(loaded.Workspaces.Artifacts))
		}
		if len(loaded.RoleMappings) != 0 {
			t.Errorf("expected generated config to have zero role mappings, got %v", loaded.RoleMappings)
		}
		if len(loaded.RoleOrder) != 0 {
			t.Errorf("expected generated config to have zero role order, got %v", loaded.RoleOrder)
		}
		rawYAML, err := os.ReadFile(expectedPath)
		if err != nil {
			t.Fatalf("reading generated config: %v", err)
		}
		content := string(rawYAML)
		if !strings.Contains(content, "role_mappings: {}") {
			t.Errorf("expected 'role_mappings: {}' in yaml, got:\n%s", content)
		}
		if !strings.Contains(content, "role_order: []") {
			t.Errorf("expected 'role_order: []' in yaml, got:\n%s", content)
		}
	})

	t.Run("custom flags customize generated configuration", func(t *testing.T) {
		tempDir := t.TempDir()
		customOutput := filepath.Join(tempDir, "custom-config.yaml")

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "AcmePlatform",
			"--workspaces", "team_workspaces",
			"--db-path", "/opt/opencode/data.db",
			"--src-folders", "services,apps,packages",
			"--ticket-prefixes", "ACME,OPS",
			"--output", customOutput,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("agentic-lens init with custom flags failed: %v, stderr: %s", err, stderr.String())
		}

		loaded, err := adaptersconfig.LoadFromFile(customOutput)
		if err != nil {
			t.Fatalf("loading custom config failed: %v", err)
		}

		if loaded.ProjectName != "AcmePlatform" {
			t.Errorf("expected ProjectName 'AcmePlatform', got %q", loaded.ProjectName)
		}
		if loaded.Workspaces.Dir != "team_workspaces" {
			t.Errorf("expected workspaces.dir 'team_workspaces', got %q", loaded.Workspaces.Dir)
		}
		if loaded.Database.Path != "/opt/opencode/data.db" {
			t.Errorf("expected database.path '/opt/opencode/data.db', got %q", loaded.Database.Path)
		}
		if len(loaded.Workspaces.SrcFolders) != 3 || loaded.Workspaces.SrcFolders[0] != "services" {
			t.Errorf("unexpected src_folders: %v", loaded.Workspaces.SrcFolders)
		}
		if len(loaded.Ticket.Prefixes) != 2 || loaded.Ticket.Prefixes[0].Prefix != "ACME" || loaded.Ticket.Prefixes[1].Prefix != "OPS" {
			t.Errorf("expected ACME,OPS peer prefixes, got %v", loaded.Ticket.Prefixes)
		}
		if len(loaded.Workspaces.Artifacts) != 0 {
			t.Errorf("expected generated custom config to have zero workspace artifacts by default, got %d", len(loaded.Workspaces.Artifacts))
		}
		if len(loaded.RoleMappings) != 0 {
			t.Errorf("expected generated custom config to have zero role mappings, got %v", loaded.RoleMappings)
		}
		if len(loaded.RoleOrder) != 0 {
			t.Errorf("expected generated custom config to have zero role order, got %v", loaded.RoleOrder)
		}
	})

	t.Run("comma-separated ticket-prefixes flag creates peer prefixes without injection", func(t *testing.T) {
		tempDir := t.TempDir()
		outConfig := filepath.Join(tempDir, ".agenticlens-config.yaml")

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "PeerTest",
			"--src-folders", "core,cli",
			"--ticket-prefixes", "ALPHA,BETA,GAMMA",
			"--output", outConfig,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("init failed: %v, stderr: %s", err, stderr.String())
		}

		loaded, err := adaptersconfig.LoadFromFile(outConfig)
		if err != nil {
			t.Fatalf("LoadFromFile failed: %v", err)
		}

		if len(loaded.Ticket.Prefixes) != 3 {
			t.Fatalf("expected exactly 3 prefixes, got %d: %v", len(loaded.Ticket.Prefixes), loaded.Ticket.Prefixes)
		}
		expected := []string{"ALPHA", "BETA", "GAMMA"}
		for i, exp := range expected {
			if loaded.Ticket.Prefixes[i].Prefix != exp {
				t.Errorf("prefix[%d]: expected %s, got %s", i, exp, loaded.Ticket.Prefixes[i].Prefix)
			}
		}
	})

	t.Run("empty ticket-prefixes fails with exit code 1", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "PeerTest",
			"--src-folders", "core",
			"--ticket-prefixes", " , , ",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected error for empty ticket-prefixes, but command succeeded")
		}
		if !strings.Contains(stderr.String(), "missing required flag --ticket-prefixes") {
			t.Errorf("expected error message to mention --ticket-prefixes, got: %s", stderr.String())
		}
	})

	t.Run("empty src-folders fails with exit code 1", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "PeerTest",
			"--src-folders", " , , ",
			"--ticket-prefixes", "ALPHA",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected error for empty src-folders, but command succeeded")
		}
		if !strings.Contains(stderr.String(), "missing required flag --src-folders") {
			t.Errorf("expected error message to mention --src-folders, got: %s", stderr.String())
		}
	})

	t.Run("collision fails safely without force", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(target, []byte("project_name: Original\n"), 0644); err != nil {
			t.Fatalf("writing initial config: %v", err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "Original",
			"--src-folders", "src",
			"--ticket-prefixes", "AL",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected init to fail when file exists and --force is omitted")
		}

		if !strings.Contains(stderr.String(), "already exists") {
			t.Errorf("expected stderr to indicate file already exists, got: %s", stderr.String())
		}

		// Ensure file was not changed
		content, _ := os.ReadFile(target)
		if !strings.Contains(string(content), "Original") {
			t.Errorf("expected original content retained, got %s", string(content))
		}
	})

	t.Run("force overwrites existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if err := os.WriteFile(target, []byte("project_name: Original\n"), 0644); err != nil {
			t.Fatalf("writing initial config: %v", err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "Overwritten",
			"--src-folders", "src",
			"--ticket-prefixes", "AL",
			"--force",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("init --force failed: %v, stderr: %s", err, stderr.String())
		}

		loaded, err := adaptersconfig.LoadFromFile(target)
		if err != nil {
			t.Fatalf("loading overwritten config: %v", err)
		}
		if loaded.ProjectName != "Overwritten" {
			t.Errorf("expected project_name 'Overwritten', got %q", loaded.ProjectName)
		}
	})

	t.Run("dry run prints YAML to stdout without creating file", func(t *testing.T) {
		tempDir := t.TempDir()

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--dry-run",
			"--project-name", "DryRunTest",
			"--src-folders", "src",
			"--ticket-prefixes", "DRY",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("init --dry-run failed: %v, stderr: %s", err, stderr.String())
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "project_name: DryRunTest") {
			t.Errorf("expected stdout to contain 'project_name: DryRunTest', got: %s", outStr)
		}

		// Ensure no file was created
		cfgPath := filepath.Join(tempDir, ".agenticlens-config.yaml")
		if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
			t.Errorf("expected %s not to exist after dry-run, but it exists", cfgPath)
		}
	})

	t.Run("spaces in project name and directories", func(t *testing.T) {
		tempDir := t.TempDir()
		projName := "My Special Project"
		customOutput := filepath.Join(tempDir, "nested dir", "app config.yaml")

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", projName,
			"--workspaces", "team workspaces",
			"--src-folders", "my web,my backend",
			"--ticket-prefixes", "SPECIAL",
			"--output", customOutput,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("agentic-lens init with spaces failed: %v, stderr: %s", err, stderr.String())
		}

		loaded, err := adaptersconfig.LoadFromFile(customOutput)
		if err != nil {
			t.Fatalf("loading config with spaces failed: %v", err)
		}
		if loaded.ProjectName != projName {
			t.Errorf("expected ProjectName %q, got %q", projName, loaded.ProjectName)
		}
		if loaded.Workspaces.Dir != "team workspaces" {
			t.Errorf("expected Workspaces.Dir 'team workspaces', got %q", loaded.Workspaces.Dir)
		}
	})

	t.Run("invalid or unknown flag returns error and exits non-zero", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := exec.Command(cliBinaryPath, "init", "--invalid-nonexistent-flag=true")
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected init with invalid flag to fail, got zero exit code")
		}
		if !strings.Contains(stderr.String(), "flag provided but not defined") {
			t.Errorf("expected stderr to contain 'flag provided but not defined', got: %s", stderr.String())
		}
	})

	t.Run("passing deprecated or unknown --ticket-prefix flag fails as unknown flag", func(t *testing.T) {
		tempDir := t.TempDir()
		cmd := exec.Command(cliBinaryPath, "init", "--project-name", "Acme", "--src-folders", "src", "--ticket-prefix", "AL")
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected init with --ticket-prefix to fail as unknown flag, got zero exit code")
		}
		if !strings.Contains(stderr.String(), "flag provided but not defined: -ticket-prefix") {
			t.Errorf("expected stderr to contain 'flag provided but not defined: -ticket-prefix', got: %s", stderr.String())
		}
	})

	t.Run("file collision error without force preserves original contents", func(t *testing.T) {
		tempDir := t.TempDir()
		target := filepath.Join(tempDir, ".agenticlens-config.yaml")
		originalContent := "# Original uncorrupted file\nproject_name: Untouched\n"
		if err := os.WriteFile(target, []byte(originalContent), 0644); err != nil {
			t.Fatalf("writing target file: %v", err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "AttemptedOverwrite",
			"--src-folders", "src",
			"--ticket-prefixes", "AL",
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatal("expected colliding init without --force to fail")
		}
		if !strings.Contains(stderr.String(), "already exists") {
			t.Errorf("expected stderr to indicate file already exists, got: %s", stderr.String())
		}

		readBack, err := os.ReadFile(target)
		if err != nil {
			t.Fatalf("reading target file: %v", err)
		}
		if string(readBack) != originalContent {
			t.Errorf("expected target content to be preserved exactly, got: %s", string(readBack))
		}
	})

	t.Run("permission error handling when output directory is not writable", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Log("running as root; skipping unprivileged permission error assertion")
			return
		}
		tempDir := t.TempDir()
		roDir := filepath.Join(tempDir, "protected")
		if err := os.MkdirAll(roDir, 0555); err != nil {
			t.Fatalf("creating roDir: %v", err)
		}
		defer os.Chmod(roDir, 0755)

		target := filepath.Join(roDir, "config.yaml")
		cmd := exec.Command(
			cliBinaryPath,
			"init",
			"--project-name", "Protected",
			"--src-folders", "src",
			"--ticket-prefixes", "AL",
			"--output", target,
		)
		cmd.Dir = tempDir
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err == nil {
			t.Fatalf("expected init to non-writable directory to fail")
		}
		if !strings.Contains(stderr.String(), "Error writing configuration file") {
			t.Errorf("expected error message in stderr, got: %s", stderr.String())
		}
	})

	t.Run("commands fail with helpful message when configuration is absent", func(t *testing.T) {
		tempDir := t.TempDir()

		// Test root command in dir with no config
		cmdRoot := exec.Command(cliBinaryPath, "--workspaces", tempDir)
		cmdRoot.Dir = tempDir
		var stderrRoot bytes.Buffer
		cmdRoot.Stderr = &stderrRoot
		if err := cmdRoot.Run(); err == nil {
			t.Fatal("expected root command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		}
		if !strings.Contains(stderrRoot.String(), "Project configuration not found. Please run 'agentic-lens init'") {
			t.Errorf("expected prompt to run init, got: %s", stderrRoot.String())
		}

		// Test server command in dir with no config
		cmdServer := exec.Command(cliBinaryPath, "server", "--port", "29199", "--workspaces", tempDir)
		cmdServer.Dir = tempDir
		var stderrServer bytes.Buffer
		cmdServer.Stderr = &stderrServer
		if err := cmdServer.Run(); err == nil {
			t.Fatal("expected server command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		}
		if !strings.Contains(stderrServer.String(), "Project configuration not found. Please run 'agentic-lens init'") {
			t.Errorf("expected prompt to run init, got: %s", stderrServer.String())
		}

		// Test server command in dir with no config and no workspaces flag
		cmdServerNoFlag := exec.Command(cliBinaryPath, "server", "--port", "29198")
		cmdServerNoFlag.Dir = tempDir
		var stderrServerNoFlag bytes.Buffer
		cmdServerNoFlag.Stderr = &stderrServerNoFlag
		if err := cmdServerNoFlag.Run(); err == nil {
			t.Fatal("expected server command without config and without workspaces flag to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		}
		if !strings.Contains(stderrServerNoFlag.String(), "Project configuration not found. Please run 'agentic-lens init'") {
			t.Errorf("expected prompt to run init, got: %s", stderrServerNoFlag.String())
		}

		// Test generate command in dir with no config
		cmdGenerate := exec.Command(cliBinaryPath, "generate", "--ticket-id", "TEST-1", "--allow-empty")
		cmdGenerate.Dir = tempDir
		var stderrGen bytes.Buffer
		cmdGenerate.Stderr = &stderrGen
		if err := cmdGenerate.Run(); err == nil {
			t.Fatal("expected generate command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
			}
		}
		if !strings.Contains(stderrGen.String(), "Project configuration not found. Please run 'agentic-lens init'") {
			t.Errorf("expected prompt to run init, got: %s", stderrGen.String())
		}
	})

	t.Run("commands fail fast without panic when workspaces directory and config are absent", func(t *testing.T) {
		tempDir := t.TempDir()

		// 1. Root command without --workspaces and without config
		cmdRoot := exec.Command(cliBinaryPath, "--list")
		cmdRoot.Dir = tempDir
		var stderrRoot bytes.Buffer
		cmdRoot.Stderr = &stderrRoot
		if err := cmdRoot.Run(); err == nil {
			t.Fatal("expected root list command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got error: %v", err)
			}
		}
		if !strings.Contains(stderrRoot.String(), "workspaces directory not found") && !strings.Contains(stderrRoot.String(), "Project configuration not found") {
			t.Errorf("expected 'workspaces directory not found' or 'Project configuration not found' in stderr, got: %s", stderrRoot.String())
		}
		if strings.Contains(stderrRoot.String(), "panic:") {
			t.Errorf("expected clean exit without panic, got: %s", stderrRoot.String())
		}

		// 2. Server command without --workspaces and without config
		cmdServer := exec.Command(cliBinaryPath, "server", "--port", "29199")
		cmdServer.Dir = tempDir
		var stderrServer bytes.Buffer
		cmdServer.Stderr = &stderrServer
		if err := cmdServer.Run(); err == nil {
			t.Fatal("expected server command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got error: %v", err)
			}
		}
		if !strings.Contains(stderrServer.String(), "workspaces directory not found") && !strings.Contains(stderrServer.String(), "Project configuration not found") {
			t.Errorf("expected 'workspaces directory not found' or 'Project configuration not found' in stderr, got: %s", stderrServer.String())
		}
		if strings.Contains(stderrServer.String(), "panic:") {
			t.Errorf("expected clean exit without panic, got: %s", stderrServer.String())
		}

		// 3. Backfill command without --workspaces and without config
		cmdBackfill := exec.Command(cliBinaryPath, "backfill", "--source", "opencode")
		cmdBackfill.Dir = tempDir
		var stderrBackfill bytes.Buffer
		cmdBackfill.Stderr = &stderrBackfill
		if err := cmdBackfill.Run(); err == nil {
			t.Fatal("expected backfill command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got error: %v", err)
			}
		}
		if !strings.Contains(stderrBackfill.String(), "workspaces directory not found") {
			t.Errorf("expected 'workspaces directory not found' in stderr, got: %s", stderrBackfill.String())
		}
		if strings.Contains(stderrBackfill.String(), "panic:") {
			t.Errorf("expected clean exit without panic, got: %s", stderrBackfill.String())
		}

		// 4. Migrate command without --workspaces and without config
		cmdMigrate := exec.Command(cliBinaryPath, "migrate")
		cmdMigrate.Dir = tempDir
		var stderrMigrate bytes.Buffer
		cmdMigrate.Stderr = &stderrMigrate
		if err := cmdMigrate.Run(); err == nil {
			t.Fatal("expected migrate command without config to fail, got zero exit code")
		} else {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected exit code 1, got error: %v", err)
			}
		}
		if !strings.Contains(stderrMigrate.String(), "workspaces directory not found") {
			t.Errorf("expected 'workspaces directory not found' in stderr, got: %s", stderrMigrate.String())
		}
		if strings.Contains(stderrMigrate.String(), "panic:") {
			t.Errorf("expected clean exit without panic, got: %s", stderrMigrate.String())
		}
	})
}

func TestCLIWithTestDataWorkspaces(t *testing.T) {
	t.Parallel()
	testdataDir := filepath.Join("..", "..", "..", "testdata", "workspaces")
	absTestdataDir, err := filepath.Abs(testdataDir)
	if err != nil {
		t.Fatalf("failed getting abs path for %s: %v", testdataDir, err)
	}
	if _, err := os.Stat(absTestdataDir); err != nil {
		t.Fatalf("testdata/workspaces directory not found at %s: %v", absTestdataDir, err)
	}

	t.Run("list command finds all 5 workspaces and anomalies", func(t *testing.T) {
		cmd := exec.Command(cliBinaryPath, "--workspaces", absTestdataDir, "--list")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI list failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "Found 5 ticket workspaces") {
			t.Errorf("expected 'Found 5 ticket workspaces', got: %s", out)
		}
		expectedTickets := []string{"AL-1", "AL-2", "AL-3", "PO-1", "GEN-1"}
		for _, ticket := range expectedTickets {
			if !strings.Contains(out, ticket) {
				t.Errorf("expected ticket %s in list output, got: %s", ticket, out)
			}
		}
		if !strings.Contains(out, "1 warn") {
			t.Errorf("expected '1 warn' anomaly summary for AL-2 in list output, got: %s", out)
		}
		if !strings.Contains(out, "crit") {
			t.Errorf("expected critical anomaly summary for GEN-1 in list output, got: %s", out)
		}
	})

	t.Run("anomalies command lists identified anomalies in test workspaces", func(t *testing.T) {
		// Test AL-2 anomaly detection
		cmdAL2 := exec.Command(cliBinaryPath, "--workspaces", absTestdataDir, "--ticket-id", "AL-2", "--anomalies")
		var stdoutAL2, stderrAL2 bytes.Buffer
		cmdAL2.Stdout = &stdoutAL2
		cmdAL2.Stderr = &stderrAL2

		if err := cmdAL2.Run(); err != nil {
			t.Fatalf("CLI anomalies for AL-2 failed: %v, stderr: %s", err, stderrAL2.String())
		}

		outAL2 := stdoutAL2.String()
		if !strings.Contains(outAL2, "TOOL_RETRY_STREAK") {
			t.Errorf("expected TOOL_RETRY_STREAK in AL-2 anomalies output, got: %s", outAL2)
		}

		// Test GEN-1 anomaly detection
		cmdGEN1 := exec.Command(cliBinaryPath, "--workspaces", absTestdataDir, "--ticket-id", "GEN-1", "--anomalies")
		var stdoutGEN1, stderrGEN1 bytes.Buffer
		cmdGEN1.Stdout = &stdoutGEN1
		cmdGEN1.Stderr = &stderrGEN1

		if err := cmdGEN1.Run(); err != nil {
			t.Fatalf("CLI anomalies for GEN-1 failed: %v, stderr: %s", err, stderrGEN1.String())
		}

		outGEN1 := stdoutGEN1.String()
		if !strings.Contains(outGEN1, "CONTEXT_SATURATION_SPIKE") {
			t.Errorf("expected CONTEXT_SATURATION_SPIKE in GEN-1 anomalies output, got: %s", outGEN1)
		}

		// Test clean ticket AL-1
		cmdAL1 := exec.Command(cliBinaryPath, "--workspaces", absTestdataDir, "--ticket-id", "AL-1", "--anomalies")
		var stdoutAL1, stderrAL1 bytes.Buffer
		cmdAL1.Stdout = &stdoutAL1
		cmdAL1.Stderr = &stderrAL1

		if err := cmdAL1.Run(); err != nil {
			t.Fatalf("CLI anomalies for AL-1 failed: %v, stderr: %s", err, stderrAL1.String())
		}

		outAL1 := stdoutAL1.String()
		if !strings.Contains(outAL1, "No anomalies detected for ticket AL-1.") {
			t.Errorf("expected 'No anomalies detected for ticket AL-1.', got: %s", outAL1)
		}
	})

	t.Run("server subcommand starts with testdata workspaces and serves test data via HTTP", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed to reserve test port: %v", err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(port), "--workspaces", absTestdataDir)
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand: %v", err)
		}
		defer func() {
			if killErr := cmd.Process.Kill(); killErr != nil {
				// Process might have terminated
			}
			_ = cmd.Wait()
		}()

		baseURL := fmt.Sprintf("http://localhost:%d", port)
		resp := waitForServerReady(t, baseURL+"/api/config", 3*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", resp.StatusCode)
		}

		var cfg struct {
			WorkspacesDir string `json:"workspaces_dir"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
			t.Fatalf("failed decoding /api/config: %v", err)
		}
		if !strings.Contains(cfg.WorkspacesDir, "testdata") {
			t.Errorf("expected workspaces_dir in config to contain testdata, got: %s", cfg.WorkspacesDir)
		}

		wsResp, err := http.Get(baseURL + "/api/workspaces")
		if err != nil {
			t.Fatalf("failed fetching /api/workspaces: %v", err)
		}
		defer wsResp.Body.Close()

		if wsResp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/workspaces, got %d", wsResp.StatusCode)
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(wsResp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding /api/workspaces: %v", err)
		}
		if len(ledgers) != 5 {
			t.Errorf("expected 5 ledgers from testdata/workspaces, got %d", len(ledgers))
		}
	})
}

func TestCLIServerConfigurationResolution(t *testing.T) {
	tempDir := t.TempDir()

	cfgContent := `project_name: ConfigTest
ticket:
  prefixes:
    - prefix: TST
workspaces:
  dir: custom_spaces
  src_folders:
    - src
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing test config: %v", err)
	}

	customDir := filepath.Join(tempDir, "custom_spaces", "TST-1")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatalf("failed creating custom spaces dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(customDir, "agentic_metrics.json"), []byte(sampleFactJSON("TST-1", "Custom Space Feature", 60, 0.50)), 0644); err != nil {
		t.Fatalf("failed writing custom ledger: %v", err)
	}

	overrideDir := filepath.Join(tempDir, "override_spaces", "TST-2")
	if err := os.MkdirAll(overrideDir, 0755); err != nil {
		t.Fatalf("failed creating override spaces dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(overrideDir, "agentic_metrics.json"), []byte(sampleFactJSON("TST-2", "Override Space Feature", 120, 1.50)), 0644); err != nil {
		t.Fatalf("failed writing override ledger: %v", err)
	}

	t.Run("starts server without workspaces flag and reads workspaces from config", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed finding free port: %v", err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(port))
		cmd.Dir = tempDir
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand without workspaces flag: %v", err)
		}
		defer func() {
			if killErr := cmd.Process.Kill(); killErr != nil {
				// Process might have terminated
			}
			_ = cmd.Wait()
		}()

		baseURL := fmt.Sprintf("http://localhost:%d", port)
		resp := waitForServerReady(t, baseURL+"/api/config", 3*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", resp.StatusCode)
		}

		var cfg struct {
			WorkspacesDir string `json:"workspaces_dir"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
			t.Fatalf("failed decoding /api/config: %v", err)
		}
		if !strings.Contains(cfg.WorkspacesDir, "custom_spaces") {
			t.Errorf("expected workspaces_dir to contain custom_spaces, got: %s", cfg.WorkspacesDir)
		}

		wsResp, err := http.Get(baseURL + "/api/workspaces")
		if err != nil {
			t.Fatalf("failed fetching /api/workspaces: %v", err)
		}
		defer wsResp.Body.Close()

		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(wsResp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding /api/workspaces: %v", err)
		}
		if len(ledgers) != 1 || *ledgers[0].TicketID != "TST-1" {
			t.Fatalf("expected 1 ledger TST-1 from custom_spaces, got: %v", ledgers)
		}
	})

	t.Run("starts server with explicit workspaces flag and overrides config workspaces dir", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed finding free port: %v", err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()

		absOverrideDir := filepath.Join(tempDir, "override_spaces")
		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(port), "--workspaces", absOverrideDir)
		cmd.Dir = tempDir
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server subcommand with override workspaces flag: %v", err)
		}
		defer func() {
			if killErr := cmd.Process.Kill(); killErr != nil {
				// Process might have terminated
			}
			_ = cmd.Wait()
		}()

		baseURL := fmt.Sprintf("http://localhost:%d", port)
		resp := waitForServerReady(t, baseURL+"/api/config", 3*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", resp.StatusCode)
		}

		var cfg struct {
			WorkspacesDir string `json:"workspaces_dir"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
			t.Fatalf("failed decoding /api/config: %v", err)
		}
		if !strings.Contains(cfg.WorkspacesDir, "override_spaces") {
			t.Errorf("expected workspaces_dir to contain override_spaces, got: %s", cfg.WorkspacesDir)
		}

		wsResp, err := http.Get(baseURL + "/api/workspaces")
		if err != nil {
			t.Fatalf("failed fetching /api/workspaces: %v", err)
		}
		defer wsResp.Body.Close()

		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(wsResp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding /api/workspaces: %v", err)
		}
		if len(ledgers) != 1 || *ledgers[0].TicketID != "TST-2" {
			t.Fatalf("expected 1 ledger TST-2 from override_spaces, got: %v", ledgers)
		}
	})

	t.Run("fails server startup when workspaces.dir is unconfigured and flag omitted", func(t *testing.T) {
		emptyWsDir := t.TempDir()
		cfgWithoutWs := `project_name: NoWsTest
ticket:
  prefixes:
    - prefix: TST
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(emptyWsDir, ".agenticlens-config.yaml"), []byte(cfgWithoutWs), 0644); err != nil {
			t.Fatalf("failed writing test config: %v", err)
		}

		cmd := exec.Command(cliBinaryPath, "server")
		cmd.Dir = emptyWsDir
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected server command to exit with error when workspaces.dir is unconfigured, got success: %s", string(out))
		}

		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit status 1, got %d", exitErr.ExitCode())
		}

		expectedErrMsg := "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag"
		if !strings.Contains(string(out), expectedErrMsg) {
			t.Errorf("expected output to contain %q, got %q", expectedErrMsg, string(out))
		}
	})

	t.Run("fails root execution when workspaces.dir is unconfigured and flag omitted", func(t *testing.T) {
		emptyWsDir := t.TempDir()
		cfgWithoutWs := `project_name: NoWsTest
ticket:
  prefixes:
    - prefix: TST
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(emptyWsDir, ".agenticlens-config.yaml"), []byte(cfgWithoutWs), 0644); err != nil {
			t.Fatalf("failed writing test config: %v", err)
		}

		cmd := exec.Command(cliBinaryPath)
		cmd.Dir = emptyWsDir
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected root command to exit with error when workspaces.dir is unconfigured, got success: %s", string(out))
		}

		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected exec.ExitError, got %v", err)
		}
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit status 1, got %d", exitErr.ExitCode())
		}

		expectedErrMsg := "Error: workspaces.dir is empty or unconfigured: configure 'workspaces.dir' in .agenticlens-config.yaml or provide --workspaces flag"
		if !strings.Contains(string(out), expectedErrMsg) {
			t.Errorf("expected output to contain %q, got %q", expectedErrMsg, string(out))
		}
	})

	t.Run("starts server when workspaces.dir unconfigured in config but --workspaces flag provided", func(t *testing.T) {
		emptyWsDir := t.TempDir()
		cfgWithoutWs := `project_name: FlagSuppliedTest
ticket:
  prefixes:
    - prefix: TST
workspaces:
  src_folders:
    - src
`
		if err := os.WriteFile(filepath.Join(emptyWsDir, ".agenticlens-config.yaml"), []byte(cfgWithoutWs), 0644); err != nil {
			t.Fatalf("failed writing test config: %v", err)
		}

		flagWsDir := filepath.Join(emptyWsDir, "flag_spaces")
		ticketDir := filepath.Join(flagWsDir, "TST-99")
		if err := os.MkdirAll(ticketDir, 0755); err != nil {
			t.Fatalf("failed creating flag spaces dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(ticketDir, "agentic_metrics.json"), []byte(sampleFactJSON("TST-99", "Flag Space Feature", 45, 0.25)), 0644); err != nil {
			t.Fatalf("failed writing flag ledger: %v", err)
		}

		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("failed finding free port: %v", err)
		}
		port := ln.Addr().(*net.TCPAddr).Port
		_ = ln.Close()

		cmd := exec.Command(cliBinaryPath, "server", "--port", strconv.Itoa(port), "--workspaces", flagWsDir)
		cmd.Dir = emptyWsDir
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed starting server with --workspaces flag: %v", err)
		}
		defer func() {
			if killErr := cmd.Process.Kill(); killErr != nil {
				// Process might have terminated
			}
			_ = cmd.Wait()
		}()

		baseURL := fmt.Sprintf("http://localhost:%d", port)
		resp := waitForServerReady(t, baseURL+"/api/config", 3*time.Second)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", resp.StatusCode)
		}

		var cfg struct {
			WorkspacesDir string `json:"workspaces_dir"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
			t.Fatalf("failed decoding /api/config: %v", err)
		}
		if !strings.Contains(cfg.WorkspacesDir, "flag_spaces") {
			t.Errorf("expected workspaces_dir to contain flag_spaces, got: %s", cfg.WorkspacesDir)
		}
	})
}

func TestCLIGenerate_UnconfiguredPatternWarnings(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "out.json")

	// Create git repository with config missing test_patterns and prod_patterns
	initGitCmd := exec.Command("git", "init")
	initGitCmd.Dir = tempDir
	if out, err := initGitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, string(out))
	}
	setGitEmail := exec.Command("git", "config", "user.email", "test@example.com")
	setGitEmail.Dir = tempDir
	_ = setGitEmail.Run()
	setGitName := exec.Command("git", "config", "user.name", "Test User")
	setGitName.Dir = tempDir
	_ = setGitName.Run()

	cfgContent := `project_name: "Test Patterns Warn"
ticket:
  prefixes:
    - prefix: "AL"
workspaces:
  dir: "workspaces"
  src_folders:
    - "."
database:
  path: "test.db"
code_metrics:
  test_patterns: []
  prod_patterns: []
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}
	gitAdd := exec.Command("git", "add", ".agenticlens-config.yaml")
	gitAdd.Dir = tempDir
	_ = gitAdd.Run()
	gitCommit := exec.Command("git", "commit", "-m", "init")
	gitCommit.Dir = tempDir
	_ = gitCommit.Run()

	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", filepath.Join(tempDir, "test.db"),
		"--ticket-id", "AL-TEST-WARN",
		"--ticket-title", "Test Warn Ticket",
		"--json-output", outJSON,
		"--allow-empty",
	)
	cmd.Dir = tempDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("generate failed: %v, stderr: %s", err, stderr.String())
	}

	errOut := stderr.String()
	if !strings.Contains(errOut, "WARNING: code_metrics.test_patterns is empty or unconfigured") {
		t.Errorf("expected warning for code_metrics.test_patterns in stderr, got: %s", errOut)
	}
	if !strings.Contains(errOut, "WARNING: code_metrics.prod_patterns is empty or unconfigured") {
		t.Errorf("expected warning for code_metrics.prod_patterns in stderr, got: %s", errOut)
	}
}

func TestCLIGenerate_MultiDeviceInvocationsMergeAndSDLC(t *testing.T) {
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "agentic_metrics.json")

	initGitCmd := exec.Command("git", "init")
	initGitCmd.Dir = tempDir
	if out, err := initGitCmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v: %s", err, string(out))
	}
	_ = exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run()

	cfgContent := `project_name: "Test Multi Device"
ticket:
  prefixes:
    - prefix: "AL"
workspaces:
  dir: "workspaces"
  src_folders:
    - "."
database:
  path: "test.db"
code_metrics:
  test_patterns: ["*_test.go"]
  prod_patterns: ["*.go"]
`
	if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
		t.Fatalf("failed writing config: %v", err)
	}
	_ = exec.Command("git", "-C", tempDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tempDir, "commit", "-m", "init").Run()

	schemaVer := domain.CurrentSchemaVersion
	ticketIDStr := "AL-HANDOFF"
	ticketTitleStr := "Cross Device Handoff Ticket"
	cachedLedger := domain.AgenticFactLedger{
		SchemaVersion: &schemaVer,
		TicketID:      &ticketIDStr,
		TicketTitle:   &ticketTitleStr,
		UpdatedAt:     "2026-10-25T10:00:00.000000+00:00", Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:             "ses_device_a",
				Role:           domain.RoleRootAgent,
				Model:          "claude-3-5-sonnet",
				User:           "engineer_a",
				Turns:          6,
				UserIterations: 3,
				Tokens: domain.InvocationTokens{
					Input:  500,
					Output: 200,
				},
			},
		},
			[]domain.ToolFact{
				{SessionID: "session-a", Name: "read", Calls: 5, Errors: 0},
			}, nil), Code: domain.CodeFacts{
			LinesAdded:     10,
			SDLCIterations: 3,
		},
	}
	if err := adapters.SaveFactLedger(outJSON, cachedLedger); err != nil {
		t.Fatalf("failed saving initial cached ledger: %v", err)
	}

	// 2. Set up SQLite DB on "Device B" with local session "ses_device_b" having 2 user messages
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}
	defer db.Close()

	createSchema := `
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
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatalf("failed creating test tables: %v", err)
	}

	now := time.Now().UnixMilli()
	if _, err := db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES ('ses_device_b', NULL, 'root', 'claude-3-5-sonnet', ?, ?)`, now, now); err != nil {
		t.Fatalf("failed inserting session: %v", err)
	}

	// Insert 2 user messages and 2 assistant messages
	uMsg1 := `{"role": "user"}`
	aMsg1 := `{"role": "assistant", "modelID": "claude-3-5-sonnet", "tokens": {"input": 100, "output": 50}}`
	uMsg2 := `{"role": "user"}`
	aMsg2 := `{"role": "assistant", "modelID": "claude-3-5-sonnet", "tokens": {"input": 120, "output": 60}}`

	if _, err := db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES
		('m1', 'ses_device_b', ?, ?, ?),
		('m2', 'ses_device_b', ?, ?, ?),
		('m3', 'ses_device_b', ?, ?, ?),
		('m4', 'ses_device_b', ?, ?, ?)`,
		now, now, uMsg1,
		now+100, now+100, aMsg1,
		now+200, now+200, uMsg2,
		now+300, now+300, aMsg2,
	); err != nil {
		t.Fatalf("failed inserting messages: %v", err)
	}

	// 3. Run generate on Device B
	cmd := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--session-id", "ses_device_b",
		"--ticket-id", "AL-HANDOFF",
		"--ticket-title", "Cross Device Handoff Ticket",
		"--json-output", outJSON,
	)
	cmd.Dir = tempDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate on Device B failed: %v, stderr: %s", err, stderr.String())
	}

	// 4. Verify merged output: both invocations present, SDLC iterations = 3 + 2 = 5
	rawMerged, err := os.ReadFile(outJSON)
	if err != nil {
		t.Fatalf("failed reading merged json: %v", err)
	}

	var mergedDTO adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(rawMerged, &mergedDTO); err != nil {
		t.Fatalf("failed unmarshaling merged json: %v", err)
	}

	if len(mergedDTO.ToDomain().InvocationFacts()) != 2 {
		t.Fatalf("expected 2 merged invocations, got %d", len(mergedDTO.ToDomain().InvocationFacts()))
	}
	var foundA, foundB bool
	for _, inv := range mergedDTO.ToDomain().InvocationFacts() {
		if inv.ID == "ses_device_a" {
			foundA = true
			if inv.UserIterations != 3 {
				t.Errorf("expected invocation A user_iterations 3, got %d", inv.UserIterations)
			}
		}
		if inv.ID == "ses_device_b" {
			foundB = true
			if inv.UserIterations != 2 {
				t.Errorf("expected invocation B user_iterations 2, got %d", inv.UserIterations)
			}
		}
	}
	if !foundA || !foundB {
		t.Errorf("missing merged invocation: foundA=%v, foundB=%v", foundA, foundB)
	}

	if mergedDTO.Code.SDLCIterations != 5 {
		t.Errorf("expected total sdlc_iterations 5 (3+2), got %d", mergedDTO.Code.SDLCIterations)
	}

	// 5. Test idempotency: run generate again and assert no duplicate invocations or inflation
	cmd2 := exec.Command(
		cliBinaryPath,
		"generate",
		"--db-path", dbPath,
		"--session-id", "ses_device_b",
		"--ticket-id", "AL-HANDOFF",
		"--ticket-title", "Cross Device Handoff Ticket",
		"--json-output", outJSON,
	)
	cmd2.Dir = tempDir
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("second generate failed: %v, output: %s", err, string(out))
	}

	rawIdempotent, _ := os.ReadFile(outJSON)
	var idempDTO adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(rawIdempotent, &idempDTO)
	if len(idempDTO.ToDomain().InvocationFacts()) != 2 {
		t.Errorf("idempotency check failed: expected 2 invocations, got %d", len(idempDTO.ToDomain().InvocationFacts()))
	}
	// 6. Runtime ingestion rejects legacy schemas and requires explicit migration and source backfill.
	legacyOutJSON := filepath.Join(tempDir, "legacy.json")
	legacyVer := 3
	legacyLedger := domain.AgenticFactLedger{
		SchemaVersion: &legacyVer,
		TicketID:      &ticketIDStr,
		TicketTitle:   &ticketTitleStr,
		UpdatedAt:     "2026-10-25T10:00:00.000000+00:00", Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:             "ses_legacy",
				Role:           domain.RoleRootAgent,
				Model:          "claude-3-5-sonnet",
				User:           "engineer_a",
				Turns:          6,
				UserIterations: 0,
			},
		}, nil, nil), Code: domain.CodeFacts{
			LinesAdded:     10,
			SDLCIterations: 0,
		},
	}
	if err := adapters.SaveFactLedger(legacyOutJSON, legacyLedger); err != nil {
		t.Fatalf("failed saving legacy ledger: %v", err)
	}

	cmdLegacy := exec.Command(
		cliBinaryPath,
		"generate",
		"--ticket-id", "AL-HANDOFF",
		"--ticket-title", "Cross Device Handoff Ticket",
		"--json-output", legacyOutJSON,
		"--allow-empty",
	)
	cmdLegacy.Dir = tempDir
	var legacyStderr bytes.Buffer
	cmdLegacy.Stderr = &legacyStderr
	if err := cmdLegacy.Run(); err == nil {
		t.Fatal("expected schema-v3 ledger to be rejected")
	}
	if !strings.Contains(legacyStderr.String(), "schema version 3") || !strings.Contains(legacyStderr.String(), "source-backed") {
		t.Errorf("expected actionable schema-v3 rejection, got: %s", legacyStderr.String())
	}
}

func TestCLIGenerate_ThreeDeviceHandoffAccumulation(t *testing.T) {
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "agentic_metrics.json")

	_ = exec.Command("git", "-C", tempDir, "init").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run()

	cfgContent := `project_name: "Test Three Device Handoff"
ticket:
  prefixes: ["AL"]
workspaces:
  dir: "workspaces"
  src_folders: ["."]
database:
  path: "device.db"
code_metrics:
  test_patterns: ["*_test.go"]
  prod_patterns: ["*.go"]
`
	_ = os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644)
	_ = exec.Command("git", "-C", tempDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tempDir, "commit", "-m", "init").Run()

	schemaVer := domain.CurrentSchemaVersion
	ticketIDStr := "AL-THREE-DEVICE"
	ticketTitleStr := "Three Device Handoff"

	// 1. Device A: writes initial ledger with ses_device_a (user_iterations: 4)
	deviceALedger := domain.AgenticFactLedger{
		SchemaVersion: &schemaVer,
		TicketID:      &ticketIDStr,
		TicketTitle:   &ticketTitleStr,
		UpdatedAt:     "2026-10-25T10:00:00.000000+00:00", Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:             "ses_device_a",
				Role:           domain.RoleRootAgent,
				Model:          "claude-3-5-sonnet",
				Turns:          8,
				UserIterations: 4,
				Tokens:         domain.InvocationTokens{Input: 1000, Output: 400},
			},
		}, nil, nil), Code: domain.CodeFacts{
			LinesAdded:     15,
			SDLCIterations: 4,
		},
	}
	if err := adapters.SaveFactLedger(outJSON, deviceALedger); err != nil {
		t.Fatalf("failed saving Device A ledger: %v", err)
	}

	// Helper to create a device sqlite database with root and child sessions
	createDeviceDB := func(dbFile string, rootSessionID string, rootUserTurns int, childSessionID string) {
		db, err := sql.Open("sqlite", dbFile)
		if err != nil {
			t.Fatalf("failed opening sqlite: %v", err)
		}
		defer db.Close()

		createSchema := `
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
		`
		if _, err := db.Exec(createSchema); err != nil {
			t.Fatalf("failed creating schema: %v", err)
		}

		now := time.Now().UnixMilli()
		// Root session
		_, _ = db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES (?, NULL, 'root', 'gpt-4o', ?, ?)`, rootSessionID, now, now)
		// Child session
		if childSessionID != "" {
			_, _ = db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES (?, ?, 'coder', 'gpt-4o', ?, ?)`, childSessionID, rootSessionID, now, now)
			// Child messages (user + assistant)
			_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('c_m1', ?, ?, ?, '{"role": "user"}')`, childSessionID, now, now)
			_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('c_m2', ?, ?, ?, '{"role": "assistant", "modelID": "gpt-4o", "tokens": {"input": 50, "output": 25}}')`, childSessionID, now, now)
		}

		// Root messages
		for i := 0; i < rootUserTurns; i++ {
			uID := fmt.Sprintf("u_%s_%d", rootSessionID, i)
			aID := fmt.Sprintf("a_%s_%d", rootSessionID, i)
			_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES (?, ?, ?, ?, '{"role": "user"}')`, uID, rootSessionID, now+int64(i*100), now+int64(i*100))
			_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES (?, ?, ?, ?, '{"role": "assistant", "modelID": "gpt-4o", "tokens": {"input": 100, "output": 50}}')`, aID, rootSessionID, now+int64(i*100)+50, now+int64(i*100)+50)
		}
	}

	// 2. Device B: root ses_device_b with 3 user turns, and child sub_device_b
	dbB := filepath.Join(tempDir, "device_b.db")
	createDeviceDB(dbB, "ses_device_b", 3, "sub_device_b")

	cmdB := exec.Command(
		cliBinaryPath, "generate",
		"--db-path", dbB,
		"--session-id", "ses_device_b",
		"--ticket-id", "AL-THREE-DEVICE",
		"--ticket-title", "Three Device Handoff",
		"--json-output", outJSON,
	)
	cmdB.Dir = tempDir
	if out, err := cmdB.CombinedOutput(); err != nil {
		t.Fatalf("generate Device B failed: %v, out: %s", err, string(out))
	}

	// Verify Device B state: 3 invocations (ses_device_a, ses_device_b, sub_device_b), SDLC = 4 + 3 = 7
	rawB, _ := os.ReadFile(outJSON)
	var dtoB adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(rawB, &dtoB)
	if len(dtoB.ToDomain().InvocationFacts()) != 3 {
		t.Fatalf("after Device B, expected 3 invocations, got %d", len(dtoB.ToDomain().InvocationFacts()))
	}
	if dtoB.Code.SDLCIterations != 7 {
		t.Fatalf("after Device B, expected SDLCIterations 7 (4+3), got %d", dtoB.Code.SDLCIterations)
	}

	// 3. Device C: root ses_device_c with 5 user turns, and child sub_device_c
	dbC := filepath.Join(tempDir, "device_c.db")
	createDeviceDB(dbC, "ses_device_c", 5, "sub_device_c")

	cmdC := exec.Command(
		cliBinaryPath, "generate",
		"--db-path", dbC,
		"--session-id", "ses_device_c",
		"--ticket-id", "AL-THREE-DEVICE",
		"--ticket-title", "Three Device Handoff",
		"--json-output", outJSON,
	)
	cmdC.Dir = tempDir
	if out, err := cmdC.CombinedOutput(); err != nil {
		t.Fatalf("generate Device C failed: %v, out: %s", err, string(out))
	}

	// Verify Device C state: 5 invocations, SDLC = 4 + 3 + 0 + 5 + 0 = 12
	rawC, _ := os.ReadFile(outJSON)
	var dtoC adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(rawC, &dtoC)
	if len(dtoC.ToDomain().InvocationFacts()) != 5 {
		t.Fatalf("after Device C, expected 5 invocations, got %d", len(dtoC.ToDomain().InvocationFacts()))
	}
	if dtoC.Code.SDLCIterations != 12 {
		t.Errorf("after Device C, expected SDLCIterations 12 (4+3+5), got %d", dtoC.Code.SDLCIterations)
	}

	// Verify each invocation's user_iterations
	expectedIterations := map[string]int64{
		"ses_device_a": 4,
		"ses_device_b": 3,
		"sub_device_b": 0,
		"ses_device_c": 5,
		"sub_device_c": 0,
	}
	for _, inv := range dtoC.ToDomain().InvocationFacts() {
		expected, exists := expectedIterations[inv.ID]
		if !exists {
			t.Errorf("unexpected invocation %s in merged ledger", inv.ID)
		} else if inv.UserIterations != expected {
			t.Errorf("invocation %s expected UserIterations %d, got %d", inv.ID, expected, inv.UserIterations)
		}
	}

	// 4. Test idempotency across repeated Device C invocations
	cmdC2 := exec.Command(
		cliBinaryPath, "generate",
		"--db-path", dbC,
		"--session-id", "ses_device_c",
		"--ticket-id", "AL-THREE-DEVICE",
		"--ticket-title", "Three Device Handoff",
		"--json-output", outJSON,
	)
	cmdC2.Dir = tempDir
	if out, err := cmdC2.CombinedOutput(); err != nil {
		t.Fatalf("repeated generate Device C failed: %v, out: %s", err, string(out))
	}

	rawC2, _ := os.ReadFile(outJSON)
	var dtoC2 adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(rawC2, &dtoC2)
	if len(dtoC2.ToDomain().InvocationFacts()) != 5 {
		t.Errorf("expected strictly 5 invocations after repeat, got %d", len(dtoC2.ToDomain().InvocationFacts()))
	}
	if dtoC2.Code.SDLCIterations != 12 {
		t.Errorf("expected SDLCIterations strictly 12 after repeat, got %d", dtoC2.Code.SDLCIterations)
	}
}

func TestCLIGenerate_InvocationsUpdateInPlaceOnSameDevice(t *testing.T) {
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "agentic_metrics.json")
	dbPath := filepath.Join(tempDir, "same_device.db")

	_ = exec.Command("git", "-C", tempDir, "init").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run()

	cfgContent := `project_name: "Test Update In Place"
ticket:
  prefixes: ["AL"]
workspaces:
  dir: "workspaces"
  src_folders: ["."]
database:
  path: "same_device.db"
code_metrics:
  test_patterns: ["*_test.go"]
  prod_patterns: ["*.go"]
`
	_ = os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644)
	_ = exec.Command("git", "-C", tempDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tempDir, "commit", "-m", "init").Run()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening sqlite: %v", err)
	}
	defer db.Close()

	createSchema := `
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
	`
	if _, err := db.Exec(createSchema); err != nil {
		t.Fatalf("failed creating schema: %v", err)
	}

	now := time.Now().UnixMilli()
	_, _ = db.Exec(`INSERT INTO session (id, parent_id, agent, model, time_created, time_updated) VALUES ('ses_dev', NULL, 'root', 'gpt-4o', ?, ?)`, now, now)

	// First pass: 1 user message, 1 assistant message
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m1', 'ses_dev', ?, ?, '{"role": "user"}')`, now, now)
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m2', 'ses_dev', ?, ?, '{"role": "assistant", "modelID": "gpt-4o", "tokens": {"input": 100, "output": 50}}')`, now+50, now+50)

	cmd1 := exec.Command(
		cliBinaryPath, "generate",
		"--db-path", dbPath,
		"--session-id", "ses_dev",
		"--ticket-id", "AL-UPDATE-IN-PLACE",
		"--ticket-title", "Update In Place Test",
		"--json-output", outJSON,
	)
	cmd1.Dir = tempDir
	if out, err := cmd1.CombinedOutput(); err != nil {
		t.Fatalf("pass 1 failed: %v, out: %s", err, string(out))
	}

	raw1, _ := os.ReadFile(outJSON)
	var dto1 adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(raw1, &dto1)
	if len(dto1.ToDomain().InvocationFacts()) != 1 {
		t.Fatalf("pass 1 expected 1 invocation, got %d", len(dto1.ToDomain().InvocationFacts()))
	}
	if dto1.ToDomain().InvocationFacts()[0].UserIterations != 1 {
		t.Errorf("pass 1 expected UserIterations 1, got %d", dto1.ToDomain().InvocationFacts()[0].UserIterations)
	}
	if dto1.ToDomain().InvocationFacts()[0].Tokens.Input != 100 {
		t.Errorf("pass 1 expected Input tokens 100, got %d", dto1.ToDomain().InvocationFacts()[0].Tokens.Input)
	}

	// Second pass: 2 more user messages, 2 more assistant messages with more tokens, and a compaction
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m3', 'ses_dev', ?, ?, '{"role": "user"}')`, now+100, now+100)
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m4', 'ses_dev', ?, ?, '{"role": "assistant", "modelID": "gpt-4o", "tokens": {"input": 200, "output": 80}}')`, now+150, now+150)
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m5', 'ses_dev', ?, ?, '{"role": "user"}')`, now+200, now+200)
	_, _ = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m6', 'ses_dev', ?, ?, '{"role": "assistant", "modelID": "gpt-4o", "tokens": {"input": 300, "output": 120}}')`, now+250, now+250)
	// Add a compaction part
	compactionJSON := `{"type":"compaction"}`
	_, _ = db.Exec(`INSERT INTO part (id, session_id, message_id, time_created, time_updated, data) VALUES ('p_comp', 'ses_dev', 'm4', ?, ?, ?)`, now+160, now+160, compactionJSON)

	cmd2 := exec.Command(
		cliBinaryPath, "generate",
		"--db-path", dbPath,
		"--session-id", "ses_dev",
		"--ticket-id", "AL-UPDATE-IN-PLACE",
		"--ticket-title", "Update In Place Test",
		"--json-output", outJSON,
	)
	cmd2.Dir = tempDir
	if out, err := cmd2.CombinedOutput(); err != nil {
		t.Fatalf("pass 2 failed: %v, out: %s", err, string(out))
	}

	raw2, _ := os.ReadFile(outJSON)
	var dto2 adapters.AgenticFactLedgerDTO
	_ = json.Unmarshal(raw2, &dto2)
	// Must NOT have duplicated invocation records!
	if len(dto2.ToDomain().InvocationFacts()) != 1 {
		t.Fatalf("pass 2 expected strictly 1 invocation (updated in place), got %d", len(dto2.ToDomain().InvocationFacts()))
	}
	inv := dto2.ToDomain().InvocationFacts()[0]
	if inv.UserIterations != 3 {
		t.Errorf("pass 2 expected UserIterations 3, got %d", inv.UserIterations)
	}
	if inv.Tokens.Input != 600 { // 100 + 200 + 300
		t.Errorf("pass 2 expected Input tokens 600, got %d", inv.Tokens.Input)
	}
	if inv.Compactions != 1 {
		t.Errorf("pass 2 expected Compactions 1, got %d", inv.Compactions)
	}
	if dto2.Code.SDLCIterations != 3 {
		t.Errorf("pass 2 expected total SDLCIterations 3, got %d", dto2.Code.SDLCIterations)
	}
}

func TestCLIGenerate_StrictZeroFallback(t *testing.T) {
	tempDir := t.TempDir()
	outJSON := filepath.Join(tempDir, "agentic_metrics.json")

	_ = exec.Command("git", "-C", tempDir, "init").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.email", "[EMAIL_REDACTED]").Run()
	_ = exec.Command("git", "-C", tempDir, "config", "user.name", "Test User").Run()

	cfgContent := `project_name: "Test Zero Fallback"
ticket:
  prefixes: ["AL"]
workspaces:
  dir: "workspaces"
  src_folders: ["."]
database:
  path: "test.db"
code_metrics:
  test_patterns: ["*_test.go"]
  prod_patterns: ["*.go"]
`
	_ = os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644)
	_ = exec.Command("git", "-C", tempDir, "add", ".").Run()
	_ = exec.Command("git", "-C", tempDir, "commit", "-m", "init").Run()

	ticketIDStr := "AL-ZERO-FALLBACK"
	ticketTitleStr := "Zero Fallback Invariant"

	t.Run("pre-v4 ledgers with zero user_iterations are rejected before accumulation", func(t *testing.T) {
		for _, ver := range []int{1, 2, 3} {
			schemaVer := ver
			ledger := domain.AgenticFactLedger{
				SchemaVersion: &schemaVer,
				TicketID:      &ticketIDStr,
				TicketTitle:   &ticketTitleStr,
				UpdatedAt:     "2026-10-25T10:00:00.000000+00:00", Sessions: sessionFacts([]domain.InvocationFact{
					{
						ID:             "ses_pre_v4",
						Role:           domain.RoleRootAgent,
						Model:          "gpt-4o",
						Turns:          5,
						UserIterations: 0, // unpopulated in pre-v4
					},
				}, nil, nil), Code: domain.CodeFacts{
					LinesAdded:     10,
					SDLCIterations: 7,
				},
			}
			if err := adapters.SaveFactLedger(outJSON, ledger); err != nil {
				t.Fatalf("failed saving v%d ledger: %v", ver, err)
			}

			cmd := exec.Command(
				cliBinaryPath, "generate",
				"--ticket-id", ticketIDStr,
				"--ticket-title", ticketTitleStr,
				"--json-output", outJSON,
				"--allow-empty",
			)
			cmd.Dir = tempDir
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				t.Fatalf("generate v%d unexpectedly accepted a legacy ledger", ver)
			}
			if !strings.Contains(stderr.String(), fmt.Sprintf("schema version %d", ver)) || !strings.Contains(stderr.String(), "source-backed") {
				t.Errorf("v%d: expected actionable schema rejection, got: %s", ver, stderr.String())
			}
			raw, _ := os.ReadFile(outJSON)
			if !bytes.Contains(raw, []byte(fmt.Sprintf(`"schema_version": %d`, ver))) {
				t.Errorf("v%d: rejected ingestion modified the original ledger: %s", ver, raw)
			}
		}
	})

	t.Run("schema-v4 ledger with zero user_iterations does not emit pre-v4 warning", func(t *testing.T) {
		schemaVer := 4
		ledger := domain.AgenticFactLedger{
			SchemaVersion: &schemaVer,
			TicketID:      &ticketIDStr,
			TicketTitle:   &ticketTitleStr,
			UpdatedAt:     "2026-10-25T10:00:00.000000+00:00", Sessions: sessionFacts([]domain.InvocationFact{
				{
					ID:             "ses_v4",
					Role:           domain.RoleRootAgent,
					Model:          "gpt-4o",
					Turns:          5,
					UserIterations: 0,
				},
			}, nil, nil), Code: domain.CodeFacts{
				LinesAdded:     10,
				SDLCIterations: 0,
			},
		}
		if err := adapters.SaveFactLedger(outJSON, ledger); err != nil {
			t.Fatalf("failed saving v4 ledger: %v", err)
		}

		cmd := exec.Command(
			cliBinaryPath, "generate",
			"--ticket-id", ticketIDStr,
			"--ticket-title", ticketTitleStr,
			"--json-output", outJSON,
			"--allow-empty",
		)
		cmd.Dir = tempDir
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("generate v4 failed: %v", err)
		}

		if strings.Contains(stderr.String(), "Existing fact ledger has schema version < 4") {
			t.Errorf("v4: did not expect schema version < 4 warning, got: %s", stderr.String())
		}
	})
}

func TestCLIMigrateReportsDroppedUnscopedToolRows(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspaces", "AL-UNSCOPED")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(workspace, "agentic_metrics.json")
	ledger := `{"schema_version":3,"ticket_id":"AL-UNSCOPED","code":{"sdlc_iterations":0},"tools":[{"name":"bash","calls":3,"errors":1}]}`
	if err := os.WriteFile(ledgerPath, []byte(ledger), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cliBinaryPath, "migrate", "--workspaces", filepath.Dir(workspace))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("migrate failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Removed 1 unscoped legacy tool") || !strings.Contains(string(output), "Re-extract from source logs") {
		t.Fatalf("CLI output must report dropped-row count and recovery guidance: %s", output)
	}
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	var migrated adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &migrated); err != nil {
		t.Fatal(err)
	}
	if migrated.Migration == nil || !strings.Contains(migrated.Migration.Notes, "Removed 1 unscoped legacy tool") {
		t.Fatalf("expected persisted migration note, got %+v", migrated.Migration)
	}
}
