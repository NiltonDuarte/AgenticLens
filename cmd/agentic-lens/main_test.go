package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
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
