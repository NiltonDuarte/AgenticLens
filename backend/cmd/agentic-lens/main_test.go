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
	tanomDir := filepath.Join(wsDir, "TICKET-ANOMALY")
	if err := os.MkdirAll(tanomDir, 0755); err != nil {
		t.Fatalf("failed to create TICKET-ANOMALY dir: %v", err)
	}
	anomalyFactJSON := `{
		"ticket_id": "TICKET-ANOMALY",
		"ticket_title": "Failing Feature",
		"session_id": "ses-anom",
		"updated_at": "2026-10-18T10:00:00Z",
		"tools": [
			{
				"name": "bash",
				"calls": 10,
				"errors": 6
			}
		]
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
		cmd := exec.Command(cliBinaryPath, "--workspaces", wsDir, "--ticket-id", "TICKET-ANOMALY", "--anomalies")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("CLI --anomalies failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "Found 1 anomalies for ticket TICKET-ANOMALY:") {
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

	t.Run("generate subcommand with --anomalies flag on ticket with anomalies", func(t *testing.T) {
		t.Parallel()
		tempDir := t.TempDir()
		outJSON := filepath.Join(tempDir, "anom_gen.json")
		// Seed an existing JSON ledger with high failure rate
		seedJSON := `{
			"ticket_id": "AL-GEN-ANOM",
			"tools": [
				{"name": "bash", "calls": 8, "errors": 6}
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
			"--json-output", outJSON,
			"--anomalies",
			"--allow-empty",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("generate --anomalies failed: %v, stderr: %s", err, stderr.String())
		}

		out := stdout.String()
		if !strings.Contains(out, "Found 1 anomalies for ticket AL-GEN-ANOM:") {
			t.Errorf("expected anomaly count header in generate --anomalies output, got: %s", out)
		}
		if !strings.Contains(out, "CRITICAL") || !strings.Contains(out, "TOOL_HIGH_FAILURE_RATE") {
			t.Errorf("expected CRITICAL TOOL_HIGH_FAILURE_RATE in generate output, got: %s", out)
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
	if parsed.SchemaVersion == nil || *parsed.SchemaVersion != 3 {
		t.Errorf("expected schema_version 3, got %v", parsed.SchemaVersion)
	}
	if len(parsed.Invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(parsed.Invocations))
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

func TestCLIGenerate_V3PrunedEventsAndZeroErrorMessage(t *testing.T) {
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
	if parsed.SchemaVersion == nil || *parsed.SchemaVersion != 3 {
		t.Errorf("expected schema_version 3, got %v", parsed.SchemaVersion)
	}

	// 4. In-memory derived metrics (invocations, tokens, cost) are fully preserved
	if len(parsed.Invocations) != 2 {
		t.Fatalf("expected 2 invocations, got %d", len(parsed.Invocations))
	}

	// 5. Events are pruned: routine tool "read_file" is pruned, but failed tool "execute_cmd" and compaction are preserved
	var hasErrorTool, hasCompaction, hasRoutineTool bool
	for _, ev := range parsed.Events {
		if ev.ErrorMessage != nil {
			t.Errorf("expected event ErrorMessage to be nil, got: %s", *ev.ErrorMessage)
		}
		if ev.EventType == string(domain.EventTypeToolCall) && ev.ToolName != nil && *ev.ToolName == "execute_cmd" {
			hasErrorTool = true
			if ev.Status != "error" {
				t.Errorf("expected execute_cmd status 'error', got %s", ev.Status)
			}
		}
		if ev.EventType == string(domain.EventTypeCompaction) {
			hasCompaction = true
		}
		if ev.EventType == string(domain.EventTypeToolCall) && ev.ToolName != nil && *ev.ToolName == "read_file" {
			hasRoutineTool = true
		}
	}

	if !hasErrorTool {
		t.Errorf("expected failed tool call to be preserved in pruned events")
	}
	if !hasCompaction {
		t.Errorf("expected compaction event to be preserved in pruned events")
	}
	if hasRoutineTool {
		t.Errorf("expected routine successful tool call to be pruned from persisted events")
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
		if len(parsed.Invocations) != 0 {
			t.Errorf("expected 0 invocations for missing DB, got %d", len(parsed.Invocations))
		}
	})

	t.Run("generate with corrupted existing JSON output preserves execution and safely recovers", func(t *testing.T) {
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
			"--json-output", outJSON,
			"--allow-empty",
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
		expectedErr := "Error: unable to identify root session for ticket 'UNKNOWN'. Provide --session-id, --nonce, or ensure workspace files exist in OpenCode database. Telemetry generation aborted."
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
		cachedJSONContent := `{"schema_version":2,"ticket_id":"AL-CACHED-INV","invocations":[{"id":"ses_cached_inv_1","user":"engineer","role":"coder","model":"gpt-4o","tokens":{"input":150,"output":75},"cost_usd":0.003}]}`
		if err := os.WriteFile(cachedJSON, []byte(cachedJSONContent), 0644); err != nil {
			t.Fatal(err)
		}

		summaryMD := filepath.Join(tempDir, "SUMMARY.md")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-CACHED-INV",
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
		if len(parsed.Invocations) != 1 {
			t.Fatalf("expected 1 invocation from cached facts, got %d", len(parsed.Invocations))
		}
		if parsed.Invocations[0].ID != "ses_cached_inv_1" {
			t.Errorf("expected invocation ID ses_cached_inv_1, got %s", parsed.Invocations[0].ID)
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
		initialJSON := `{"schema_version":2,"ticket_id":"AL-CACHED-NO-INV","invocations":[]}`
		if err := os.WriteFile(cachedJSON, []byte(initialJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-CACHED-NO-INV",
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
		initialJSON := `{"schema_version":2,"ticket_id":"AL-EMPTY-BOTH","invocations":[]}`
		if err := os.WriteFile(cachedJSON, []byte(initialJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", emptyDB,
			"--ticket-id", "AL-EMPTY-BOTH",
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
		if len(parsed.Invocations) != 0 {
			t.Errorf("expected 0 invocations, got %d", len(parsed.Invocations))
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

	t.Run("dry-run backfill outputs preview without error", func(t *testing.T) {
		cmd := exec.Command(
			cliBinaryPath,
			"backfill",
			"--workspaces", wsDir,
			"--dry-run",
			"--recompute-git",
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
			"--workspaces", wsDir,
			"--ticket-id", "TICKET-1",
			"--recompute-git",
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			t.Fatalf("backfill failed: %v, stderr: %s", err, stderr.String())
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
			t.Errorf("expected migration object to be omitted on completed status, got %v", parsed.Migration)
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

		// Verify files are unmutated (no schema_version)
		data, err := os.ReadFile(filepath.Join(wsDir, "TICKET-1", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading file: %v", err)
		}
		if strings.Contains(string(data), `"schema_version"`) {
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
		if strings.Contains(string(bakData), `"schema_version"`) {
			t.Errorf("backup file should be the original unversioned ledger")
		}

		// Verify migrated file is upgraded to current schema version with events array
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
		if parsed.Events == nil {
			t.Errorf("expected events array to be initialized")
		}

		// TICKET-2 should remain untouched
		t2Data, err := os.ReadFile(filepath.Join(wsDir, "TICKET-2", "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading TICKET-2 file: %v", err)
		}
		if strings.Contains(string(t2Data), `"schema_version"`) {
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
		"--workspaces", wsDir,
		"--migrate-schema",
		"--recompute-git",
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

func containsEmoji(s string) bool {
	for _, r := range s {
		if (r >= 0x1F600 && r <= 0x1F64F) ||
			(r >= 0x1F300 && r <= 0x1F5FF) ||
			(r >= 0x1F680 && r <= 0x1F6FF) ||
			(r >= 0x2600 && r <= 0x26FF) ||
			(r >= 0x2700 && r <= 0x27BF) ||
			(r >= 0x1F900 && r <= 0x1F9FF) ||
			(r >= 0x1FA70 && r <= 0x1FAFF) {
			return true
		}
	}
	return false
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

	t.Run("output strictly adheres to zero-emoji invariant", func(t *testing.T) {
		wsDir := setupTestWorkspace(t)
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
			t.Fatalf("migrate failed: %v", err)
		}

		if containsEmoji(stdout.String()) {
			t.Errorf("stdout contains forbidden emoji characters: %s", stdout.String())
		}
		if containsEmoji(stderr.String()) {
			t.Errorf("stderr contains forbidden emoji characters: %s", stderr.String())
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
			"code": {"lines_added": 20, "sdlc_iterations": 1},
			"invocations": [],
			"tools": []
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

		if containsEmoji(out) {
			t.Errorf("root --help output contains forbidden emoji characters: %s", out)
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

		if containsEmoji(out) {
			t.Errorf("root -h output contains forbidden emoji characters: %s", out)
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

		if containsEmoji(out) {
			t.Errorf("root help output contains forbidden emoji characters: %s", out)
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

		if containsEmoji(out) {
			t.Errorf("root -help output contains forbidden emoji characters: %s", out)
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
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(out, flag) {
				t.Errorf("expected backfill --help output to contain %q, but was missing. Output:\n%s", flag, out)
			}
		}

		if containsEmoji(out) {
			t.Errorf("backfill --help output contains forbidden emoji characters: %s", out)
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
			if containsEmoji(out) {
				t.Errorf("%v output contains forbidden emoji characters", invocation)
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

		if containsEmoji(out) {
			t.Errorf("migrate --help output contains forbidden emoji characters: %s", out)
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
			if containsEmoji(out) {
				t.Errorf("%v output contains forbidden emoji characters", invocation)
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
	if len(rawLedger.Invocations) != 1 {
		t.Fatalf("expected 1 invocation in raw facts, got %d", len(rawLedger.Invocations))
	}
	if rawLedger.Invocations[0].Role != "build" {
		t.Errorf("expected raw invocation role 'build', got %q", rawLedger.Invocations[0].Role)
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
			"schema_version": 2,
			"ticket_id": "AL-29",
			"ticket_type": "feature",
			"code": {
				"lines_added": 10,
				"lines_deleted": 2,
				"files_changed": 1,
				"sdlc_iterations": 1
			},
			"invocations": [],
			"tools": [],
			"events": []
		}`
		if err := os.WriteFile(outJSON, []byte(legacyJSON), 0644); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-29",
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
			expected string
		}{
			{input: "  FeAt  ", expected: "feat"},
			{input: "BUGFIX", expected: "fix"},
			{input: "DoCs", expected: "docs"},
			{input: "  cI  ", expected: "ci"},
			{input: "PeRfOrMaNcE", expected: "perf"},
			{input: "  bUiLd  ", expected: "build"},
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

		expectedErr := "Error: missing ticket type for ticket 'UNKNOWN'. Specify --ticket-type explicitly."
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
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatalf("failed creating .agenticlens dir: %v", err)
		}
		cfgContent := "ticket:\n  prefixes:\n    - prefix: \"CUSTOM\"\n      ticket_type: \"fix\"\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "CUSTOM-101",
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
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatalf("failed creating .agenticlens dir: %v", err)
		}
		cfgContent := "ticket:\n  prefixes:\n    - prefix: \"PO\"\n      ticket_type: \"feat\"\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "PO-88",
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
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatalf("failed creating .agenticlens dir: %v", err)
		}
		cfgContent := "ticket:\n  prefixes:\n    - prefix: \"CUSTOM\"\n      ticket_type: \"fix\"\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "CUSTOM-101",
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
		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatalf("failed creating .agenticlens dir: %v", err)
		}
		cfgContent := "ticket:\n  prefixes:\n    - prefix: \"HOTFIX\"\n      ticket_type: \"fix\"\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "HOTFIX-77",
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
		if err := exec.Command("git", "init", tempDir).Run(); err != nil {
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

		outJSON := filepath.Join(tempDir, "facts.json")

		// AL-30 is on branch fix/AL-30-test. Passing --ticket-type chore MUST override the branch prefix "fix".
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-30",
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
		if err := exec.Command("git", "init", tempDir).Run(); err != nil {
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

		cfgDir := filepath.Join(tempDir, ".agenticlens")
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			t.Fatalf("failed creating .agenticlens dir: %v", err)
		}
		// Configure AL to map to docs in config.yaml
		cfgContent := "ticket:\n  prefixes:\n    - prefix: \"AL\"\n      ticket_type: \"docs\"\n"
		if err := os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config.yaml: %v", err)
		}

		outJSON := filepath.Join(tempDir, "facts.json")
		// Config prefix "docs" takes precedence over branch prefix "fix".
		cmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--db-path", setupIsolatedEmptyDB(t),
			"--ticket-id", "AL-30",
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
`
		if err := os.WriteFile(filepath.Join(tempDir, ".agenticlens-config.yaml"), []byte(cfgContent), 0644); err != nil {
			t.Fatalf("failed writing config: %v", err)
		}

		outJSON := filepath.Join(ticketWsDir, "agentic_metrics.json")
		genCmd := exec.Command(
			cliBinaryPath,
			"generate",
			"--ticket-id", "MYROOT-5",
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
ticket:
  prefixes:
    - prefix: "ROOTPRIO"
      ticket_type: "feat"
workspaces:
  dir: "root_ws"
`
		// Legacy config maps ROOTPRIO to "chore" and dir to "legacy_ws"
		legacyCfg := `
ticket:
  prefixes:
    - prefix: "ROOTPRIO"
      ticket_type: "chore"
workspaces:
  dir: "legacy_ws"
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
