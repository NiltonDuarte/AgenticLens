package backfill_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/domain"
)

func sampleHistoricalJSON(ticketID string) string {
	return `{
		"ticket_id": "` + ticketID + `",
		"ticket_title": "Historical Ticket ` + ticketID + `",
		"ticket_type": "feature",
		"started_at": "2026-10-18T10:00:00Z",
		"completed_at": "2026-10-18T11:00:00Z",
		"updated_at": "2026-10-18T11:00:00Z",
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
				"id": "inv-1",
				"role": "orchestrator",
				"model": "claude-3-sonnet",
				"user": "test-engineer",
				"active_duration_seconds": 120,
				"turns": 6,
				"tokens": {
					"input": 5000,
					"output": 1000
				},
				"cost_usd": 0.25
			}
		],
		"tools": [
			{
				"session_id": "inv-1",
				"name": "bash",
				"calls": 10,
				"errors": 1
			}
		]
	}`
}

func setupMigrationTestWorkspaces(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed to create workspaces dir: %v", err)
	}

	// AL-1: Unversioned v1
	al1Dir := filepath.Join(wsDir, "AL-1")
	_ = os.MkdirAll(al1Dir, 0755)
	if err := os.WriteFile(filepath.Join(al1Dir, "agentic_metrics.json"), []byte(sampleHistoricalJSON("AL-1")), 0644); err != nil {
		t.Fatal(err)
	}

	// AL-2: Explicit v1
	al2Dir := filepath.Join(wsDir, "AL-2")
	_ = os.MkdirAll(al2Dir, 0755)
	al2JSON := `{"schema_version": 1, "ticket_id": "AL-2", "updated_at": "2026-10-18T11:00:00Z", "code": {"lines_added": 50}, "invocations": [], "tools": []}`
	if err := os.WriteFile(filepath.Join(al2Dir, "agentic_metrics.json"), []byte(al2JSON), 0644); err != nil {
		t.Fatal(err)
	}

	// AL-3: Explicit v2 (already up to date)
	al3Dir := filepath.Join(wsDir, "AL-3")
	_ = os.MkdirAll(al3Dir, 0755)
	al3JSON := `{"schema_version": 2, "ticket_id": "AL-3", "updated_at": "2026-10-18T11:00:00Z", "code": {"lines_added": 30}, "invocations": [], "tools": [], "events": []}`
	if err := os.WriteFile(filepath.Join(al3Dir, "agentic_metrics.json"), []byte(al3JSON), 0644); err != nil {
		t.Fatal(err)
	}

	// AL-4: Corrupted JSON
	al4Dir := filepath.Join(wsDir, "AL-4")
	_ = os.MkdirAll(al4Dir, 0755)
	if err := os.WriteFile(filepath.Join(al4Dir, "agentic_metrics.json"), []byte(`{invalid-json!`), 0644); err != nil {
		t.Fatal(err)
	}

	return wsDir
}

func TestMigrationRunner_Batch(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}

	if report.Scanned != 4 {
		t.Errorf("expected 4 scanned workspaces, got %d", report.Scanned)
	}
	if report.Migrated != 3 {
		t.Errorf("expected 3 migrated workspaces, got %d", report.Migrated)
	}
	if report.UpToDate != 0 {
		t.Errorf("expected no up-to-date workspace on initial current migration, got %d", report.UpToDate)
	}
	if len(report.Errors) != 1 {
		t.Errorf("expected 1 non-fatal error for AL-4, got %d", len(report.Errors))
	}

	// Assert AL-1 migrated to schema version 2
	al1Data, err := os.ReadFile(filepath.Join(wsDir, "AL-1", "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(al1Data), `"tokens": {"input":`) {
		t.Fatalf("migration writer expanded compact token record: %s", al1Data)
	}
	var al1DTO adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(al1Data, &al1DTO); err != nil {
		t.Fatalf("failed unmarshaling AL-1: %v", err)
	}
	if al1DTO.SchemaVersion == nil || *al1DTO.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected AL-1 current schema version, got %v", al1DTO.SchemaVersion)
	}
	if al1DTO.Sessions == nil {
		t.Errorf("expected AL-1 sessions array initialized")
	}
	if al1DTO.Code.LinesAdded != 120 {
		t.Errorf("expected AL-1 lines_added 120, got %d", al1DTO.Code.LinesAdded)
	}
	if len(al1DTO.Sessions) != 1 || len(al1DTO.Sessions[0].Invocations) != 1 || al1DTO.Sessions[0].ID != "inv-1" {
		t.Errorf("expected AL-1 invocation preserved")
	}

	// Assert AL-2 migrated
	al2Data, err := os.ReadFile(filepath.Join(wsDir, "AL-2", "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var al2DTO adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(al2Data, &al2DTO); err != nil {
		t.Fatal(err)
	}
	if al2DTO.SchemaVersion == nil || *al2DTO.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected AL-2 current schema version, got %v", al2DTO.SchemaVersion)
	}

	// Test idempotency: re-running produces zero migrations and all 3 valid workspaces are up-to-date
	report2, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report2.Migrated != 0 {
		t.Errorf("expected 0 migrated on second run, got %d", report2.Migrated)
	}
	if report2.UpToDate != 3 {
		t.Errorf("expected 3 up-to-date on second run, got %d", report2.UpToDate)
	}
}

func TestMigrationRunner_DryRun(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	alDir := filepath.Join(wsDir, "AL-DRY")
	_ = os.MkdirAll(alDir, 0755)

	origContent := []byte(sampleHistoricalJSON("AL-DRY"))
	metricPath := filepath.Join(alDir, "agentic_metrics.json")
	if err := os.WriteFile(metricPath, origContent, 0644); err != nil {
		t.Fatal(err)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}

	if report.Migrated != 1 {
		t.Errorf("expected 1 planned migration, got %d", report.Migrated)
	}
	if len(report.Results) != 1 || report.Results[0].Status != "migrated (dry-run)" {
		t.Errorf("expected 'migrated (dry-run)' status, got %+v", report.Results)
	}

	// Assert file on disk was NOT mutated
	afterContent, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterContent) != string(origContent) {
		t.Errorf("dry-run mutated file on disk!")
	}
}

func TestMigrationRunner_Backup(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	alDir := filepath.Join(wsDir, "AL-BAK")
	_ = os.MkdirAll(alDir, 0755)

	origContent := []byte(sampleHistoricalJSON("AL-BAK"))
	metricPath := filepath.Join(alDir, "agentic_metrics.json")
	if err := os.WriteFile(metricPath, origContent, 0644); err != nil {
		t.Fatal(err)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
		Backup:        true,
	})
	if err != nil {
		t.Fatalf("backup run failed: %v", err)
	}

	if report.Migrated != 1 {
		t.Errorf("expected 1 migrated, got %d", report.Migrated)
	}
	if !report.Results[0].BackupCreated {
		t.Errorf("expected BackupCreated to be true")
	}

	// Check .bak file exists and equals origContent
	bakPath := filepath.Join(alDir, "agentic_metrics.json.bak")
	bakData, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("expected .bak file to exist: %v", err)
	}
	if string(bakData) != string(origContent) {
		t.Errorf("backup file content did not match original")
	}

	// Check upgraded file
	upgradedData, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatal(err)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(upgradedData, &dto); err != nil {
		t.Fatal(err)
	}
	if dto.SchemaVersion == nil || *dto.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected upgraded current schema_version, got %v", dto.SchemaVersion)
	}
}

func TestMigrationRunner_CleansCompletedCurrentMetadataWithoutChangingFacts(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "records", "TEST-CLEAN", "iterations", "07")
	if err := os.MkdirAll(workspace, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspace, "agentic_metrics.json")
	original := []byte(`{"schema_version":6,"ticket_id":"TEST-CLEAN","ticket_title":"Portable history","updated_at":"2026-01-01T00:00:00Z","migration":{"schema_version":6,"status":"completed","steps_completed":["schema_v6_lean_occurrences"],"steps_pending":[],"notes":"Schema v6 migration dropped ownerless skills=2 task_links=3"},"sessions":[{"id":"remote-owner","source":"codex","user":"remote-user","invocations":[{"role":"coder","model":"fixture","cost_usd":1.25,"cost_evidence":{"state":"observed","known_subtotal_usd":1.25},"started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":0,"compactions":0,"user_iterations":0,"tokens":{"input":10,"output":0,"cache_read":0,"reasoning":0,"peak_context":0}}],"tools":[],"anomalies":[],"skills":[{"name":"kept","caller_role":null,"invoked_at":null}],"delegations":[{"agent_type":null,"invoked_at":null,"child_session_id":"unavailable-child"}]}]}`)
	if err := adapters.ValidateRuntimeFactLedger(original); err != nil {
		t.Fatalf("current fixture is not runtime-admissible: %v", err)
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	runner := backfill.NewMigrationRunner(nil)
	opts := backfill.MigrationOptions{WorkspacesDir: filepath.Join(root, "records"), TargetVersion: domain.CurrentSchemaVersion, Backup: true, DryRun: true}
	dry, err := runner.Run(opts)
	if err != nil || len(dry.Results) != 1 || dry.Migrated != 1 || dry.Results[0].BackupCreated {
		t.Fatalf("cleanup dry run result: %+v, %v", dry, err)
	}
	if data, err := os.ReadFile(path); err != nil || !bytes.Equal(data, original) {
		t.Fatalf("cleanup dry run changed original: %v", err)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatalf("cleanup dry run created a backup: %v", err)
	}
	opts.DryRun = false
	report, err := runner.Run(opts)
	if err != nil || len(report.Results) != 1 || report.Migrated != 1 || len(report.Errors) != 0 {
		t.Fatalf("cleanup result: %+v, %v", report, err)
	}
	result := report.Results[0]
	if !result.BackupCreated || result.BackupPath != path+".bak" || result.LedgerPath != path || !strings.Contains(result.DiffSummary, "completed steps: [schema_v6_lean_occurrences]") || !strings.Contains(result.DiffSummary, "dropped ownerless skills=2 task_links=3") {
		t.Fatalf("cleanup lost ledger identity, backup or console audit: %+v", result)
	}
	backup, err := os.ReadFile(path + ".bak")
	if err != nil || !bytes.Equal(backup, original) {
		t.Fatalf("cleanup backup did not preserve exact original bytes: %v", err)
	}
	cleaned, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var expected, got map[string]json.RawMessage
	if err := json.Unmarshal(original, &expected); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(cleaned, &got); err != nil {
		t.Fatal(err)
	}
	delete(expected, "migration")
	if _, exists := got["migration"]; exists {
		t.Fatal("completed metadata remains on disk")
	}
	for key, value := range expected {
		var beforeValue, afterValue json.RawMessage
		beforeValue, err = json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		afterValue, err = json.Marshal(got[key])
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(beforeValue, afterValue) {
			t.Fatalf("cleanup changed non-migration field %q: before=%s after=%s", key, beforeValue, afterValue)
		}
	}
	if len(got) != len(expected) {
		t.Fatalf("cleanup changed root member count: before=%d after=%d", len(expected), len(got))
	}
	second, err := runner.Run(opts)
	if err != nil || second.Migrated != 0 || second.UpToDate != 1 || len(second.Errors) != 0 {
		t.Fatalf("repeated cleanup was not up-to-date: %+v, %v", second, err)
	}
	afterSecond, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(afterSecond, cleaned) {
		t.Fatalf("repeated cleanup changed ledger bytes: %v", err)
	}
	backupAfterSecond, err := os.ReadFile(path + ".bak")
	if err != nil || !bytes.Equal(backupAfterSecond, backup) {
		t.Fatalf("repeated cleanup overwrote original rollback backup: %v", err)
	}
}

func TestMigrationRunner_TicketFilter(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TicketID:      "AL-1",
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatal(err)
	}

	if report.Scanned != 1 {
		t.Errorf("expected 1 scanned workspace with filter, got %d", report.Scanned)
	}
	if report.Migrated != 1 || report.Results[0].TicketID != "AL-1" {
		t.Errorf("expected AL-1 migrated, got %+v", report.Results)
	}
}

func TestMigrationRunner_UpdateSummaries(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	alDir := filepath.Join(wsDir, "AL-SUM")
	_ = os.MkdirAll(alDir, 0755)

	metricPath := filepath.Join(alDir, "agentic_metrics.json")
	if err := os.WriteFile(metricPath, []byte(sampleHistoricalJSON("AL-SUM")), 0644); err != nil {
		t.Fatal(err)
	}

	summaryPath := filepath.Join(alDir, "SUMMARY.md")
	initialSummary := "# Summary AL-SUM\n\nSome introductory remarks.\n"
	if err := os.WriteFile(summaryPath, []byte(initialSummary), 0644); err != nil {
		t.Fatal(err)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir:   wsDir,
		TargetVersion:   domain.CurrentSchemaVersion,
		UpdateSummaries: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Migrated != 1 {
		t.Errorf("expected 1 migrated, got %d", report.Migrated)
	}

	sumData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	sumStr := string(sumData)
	if !contains(sumStr, "### Workflow Efficiency") {
		t.Errorf("expected summary to contain metrics table")
	}
}

func contains(s, substr string) bool {
	return domain.FormatTokensK(1000) != "" && len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading file %s for hash: %v", path, err)
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func TestMigrationRunner_DirectoryEdgeCases(t *testing.T) {
	runner := backfill.NewMigrationRunner(nil)

	t.Run("non-existent workspaces directory returns error", func(t *testing.T) {
		_, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: "/non/existent/path/for/sure/migration",
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err == nil {
			t.Fatal("expected error for non-existent directory, got nil")
		}
		if !strings.Contains(err.Error(), "workspaces directory not found") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("empty workspaces directory produces zero results cleanly", func(t *testing.T) {
		emptyDir := t.TempDir()
		report, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: emptyDir,
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Scanned != 0 || report.Migrated != 0 || report.UpToDate != 0 || len(report.Errors) != 0 {
			t.Errorf("expected clean zero report, got %+v", report)
		}
	})

	t.Run("workspaces directory with subdirectories containing no ledgers", func(t *testing.T) {
		root := t.TempDir()
		_ = os.MkdirAll(filepath.Join(root, "TICKET-EMPTY-1"), 0755)
		_ = os.MkdirAll(filepath.Join(root, "TICKET-EMPTY-2"), 0755)
		_ = os.WriteFile(filepath.Join(root, "TICKET-EMPTY-1", "notes.md"), []byte("# Notes"), 0644)

		report, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: root,
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if report.Scanned != 0 || report.Migrated != 0 {
			t.Errorf("expected 0 scanned ledgers, got %d", report.Scanned)
		}
	})
}

func TestMigrationRunner_CorruptedAndDefectiveLedgerResilience(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	_ = os.MkdirAll(wsDir, 0755)

	// 1. Valid v1 ledger
	w1Dir := filepath.Join(wsDir, "AL-VALID-1")
	_ = os.MkdirAll(w1Dir, 0755)
	_ = os.WriteFile(filepath.Join(w1Dir, "agentic_metrics.json"), []byte(sampleHistoricalJSON("AL-VALID-1")), 0644)

	// 2. Zero-byte file
	w2Dir := filepath.Join(wsDir, "AL-ZERO")
	_ = os.MkdirAll(w2Dir, 0755)
	_ = os.WriteFile(filepath.Join(w2Dir, "agentic_metrics.json"), []byte(""), 0644)

	// 3. Syntax error JSON
	w3Dir := filepath.Join(wsDir, "AL-SYNTAX")
	_ = os.MkdirAll(w3Dir, 0755)
	_ = os.WriteFile(filepath.Join(w3Dir, "agentic_metrics.json"), []byte(`{"ticket_id": "AL-SYNTAX", "code": {`), 0644)

	// 4. Corrupted schema_version type
	w4Dir := filepath.Join(wsDir, "AL-CORRUPT-TYPE")
	_ = os.MkdirAll(w4Dir, 0755)
	_ = os.WriteFile(filepath.Join(w4Dir, "agentic_metrics.json"), []byte(`{"schema_version": "corrupted", "ticket_id": "AL-4"}`), 0644)

	// 5. Valid v2 ledger (up to date)
	w5Dir := filepath.Join(wsDir, "AL-UPTODATE")
	_ = os.MkdirAll(w5Dir, 0755)
	v2JSON := `{"schema_version":6,"ticket_id":"AL-UPTODATE","updated_at":"2026-10-18T10:00:00Z","sessions":[],"anomalies":[]}`
	_ = os.WriteFile(filepath.Join(w5Dir, "agentic_metrics.json"), []byte(v2JSON), 0644)

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("unexpected runner execution error: %v", err)
	}

	if report.Scanned != 5 {
		t.Errorf("expected 5 scanned workspaces, got %d", report.Scanned)
	}
	if report.Migrated != 1 {
		t.Errorf("expected 1 migrated workspace, got %d", report.Migrated)
	}
	if report.UpToDate != 1 {
		t.Errorf("expected 1 up-to-date workspace, got %d", report.UpToDate)
	}
	if len(report.Errors) != 3 {
		t.Errorf("expected 3 non-fatal errors for corrupted files, got %d: %v", len(report.Errors), report.Errors)
	}

	// Verify AL-VALID-1 was upgraded to v2
	valData, err := os.ReadFile(filepath.Join(w1Dir, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var valDTO adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(valData, &valDTO); err != nil {
		t.Fatalf("failed unmarshaling upgraded file: %v", err)
	}
	if valDTO.SchemaVersion == nil || *valDTO.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected current schema_version, got %v", valDTO.SchemaVersion)
	}
	if valDTO.Sessions == nil {
		t.Errorf("expected initialized sessions array")
	}

	// Verify zero-byte and syntax error files were not corrupted further or panicked
	zData, _ := os.ReadFile(filepath.Join(w2Dir, "agentic_metrics.json"))
	if len(zData) != 0 {
		t.Errorf("zero-byte file should remain zero-byte on error")
	}
}

func TestMigrationRunner_DryRunImmutabilityStrict(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	_ = os.MkdirAll(wsDir, 0755)

	ticketDirs := []string{"T1", "T2", "T3"}
	filePaths := make([]string, len(ticketDirs))
	beforeHashes := make([]string, len(ticketDirs))

	for i, td := range ticketDirs {
		d := filepath.Join(wsDir, td)
		_ = os.MkdirAll(d, 0755)
		f := filepath.Join(d, "agentic_metrics.json")
		_ = os.WriteFile(f, []byte(sampleHistoricalJSON(td)), 0644)
		filePaths[i] = f
		beforeHashes[i] = fileSHA256(t, f)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if report.Scanned != 3 || report.Migrated != 3 {
		t.Errorf("expected 3 scanned and 3 migrated preview, got scanned %d, migrated %d", report.Scanned, report.Migrated)
	}

	// Assert byte-for-byte SHA256 immutability
	for i, f := range filePaths {
		afterHash := fileSHA256(t, f)
		if beforeHashes[i] != afterHash {
			t.Errorf("dry-run mutated file %s: hash before %s, hash after %s", f, beforeHashes[i], afterHash)
		}

		// Ensure no temp or backup files exist
		if _, err := os.Stat(f + ".tmp"); err == nil {
			t.Errorf("temporary file %s.tmp was left behind after dry-run", f)
		}
		if _, err := os.Stat(f + ".bak"); err == nil {
			t.Errorf("backup file %s.bak was created during dry-run", f)
		}
	}
}

func TestMigrationRunner_IdempotencyStrictAcrossRepeatedRuns(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	_ = os.MkdirAll(wsDir, 0755)

	for _, tid := range []string{"IDEM-1", "IDEM-2"} {
		d := filepath.Join(wsDir, tid)
		_ = os.MkdirAll(d, 0755)
		_ = os.WriteFile(filepath.Join(d, "agentic_metrics.json"), []byte(sampleHistoricalJSON(tid)), 0644)
	}

	runner := backfill.NewMigrationRunner(nil)

	// Pass 1: Migrates both to v2
	r1, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("pass 1 failed: %v", err)
	}
	if r1.Scanned != 2 || r1.Migrated != 2 || r1.UpToDate != 0 {
		t.Errorf("pass 1 unexpected counts: scanned %d, migrated %d, up-to-date %d", r1.Scanned, r1.Migrated, r1.UpToDate)
	}

	// Capture hashes after pass 1
	h1 := fileSHA256(t, filepath.Join(wsDir, "IDEM-1", "agentic_metrics.json"))
	h2 := fileSHA256(t, filepath.Join(wsDir, "IDEM-2", "agentic_metrics.json"))

	// Pass 2: Strictly idempotent (all up-to-date, zero writes)
	r2, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("pass 2 failed: %v", err)
	}
	if r2.Scanned != 2 || r2.Migrated != 0 || r2.UpToDate != 2 || len(r2.Errors) != 0 {
		t.Errorf("pass 2 unexpected counts: scanned %d, migrated %d, up-to-date %d, errors %d",
			r2.Scanned, r2.Migrated, r2.UpToDate, len(r2.Errors))
	}

	// Pass 3: Repeat again
	r3, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("pass 3 failed: %v", err)
	}
	if r3.Migrated != 0 || r3.UpToDate != 2 {
		t.Errorf("pass 3 unexpected counts: migrated %d, up-to-date %d", r3.Migrated, r3.UpToDate)
	}

	// Hashes should remain identical to post-pass-1
	h1After := fileSHA256(t, filepath.Join(wsDir, "IDEM-1", "agentic_metrics.json"))
	h2After := fileSHA256(t, filepath.Join(wsDir, "IDEM-2", "agentic_metrics.json"))
	if h1 != h1After || h2 != h2After {
		t.Errorf("file hashes modified during idempotent re-runs")
	}

	// Ensure no lingering .tmp files
	for _, tid := range []string{"IDEM-1", "IDEM-2"} {
		tmpPath := filepath.Join(wsDir, tid, "agentic_metrics.json.tmp")
		if _, err := os.Stat(tmpPath); err == nil {
			t.Errorf("found lingering temporary file: %s", tmpPath)
		}
	}
}

func TestMigrationRunner_TicketFilterAdversarial(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	t.Run("case insensitive match", func(t *testing.T) {
		report, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: wsDir,
			TicketID:      "al-1", // lowercase
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Scanned != 1 || report.Results[0].TicketID != "AL-1" {
			t.Errorf("expected lowercase al-1 to match AL-1, got %+v", report.Results)
		}
	})

	t.Run("whitespace trimmed match", func(t *testing.T) {
		report, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: wsDir,
			TicketID:      "   AL-2   ",
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Scanned != 1 || report.Results[0].TicketID != "AL-2" {
			t.Errorf("expected trimmed AL-2 to match AL-2, got %+v", report.Results)
		}
	})

	t.Run("non-existent ticket filter yields 0 scanned", func(t *testing.T) {
		report, err := runner.Run(backfill.MigrationOptions{
			WorkspacesDir: wsDir,
			TicketID:      "DOES-NOT-EXIST",
			TargetVersion: domain.CurrentSchemaVersion,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Scanned != 0 || len(report.Results) != 0 {
			t.Errorf("expected 0 scanned, got %d", report.Scanned)
		}
	})
}

func TestMigrationRunner_BatchCurrentPrivacyBounds(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	// Step 1: Dry run to v3 should not modify any files on disk
	reportDry, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
		DryRun:        true,
	})
	if err != nil {
		t.Fatalf("unexpected runner error on dry run: %v", err)
	}
	if reportDry.Migrated != 3 {
		t.Errorf("expected 3 planned migrations in dry-run, got %d", reportDry.Migrated)
	}

	// Verify files remain at original versions after dry-run
	al3DataDry, err := os.ReadFile(filepath.Join(wsDir, "AL-3", "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var al3DTODry struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(al3DataDry, &al3DTODry); err != nil {
		t.Fatal(err)
	}
	if al3DTODry.SchemaVersion == nil || *al3DTODry.SchemaVersion != 2 {
		t.Errorf("expected AL-3 to remain version 2 after dry-run, got %v", al3DTODry.SchemaVersion)
	}

	// Step 2: Live migration to v3
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}

	if report.Scanned != 4 {
		t.Errorf("expected 4 scanned workspaces, got %d", report.Scanned)
	}
	if report.Migrated != 3 {
		t.Errorf("expected 3 migrated workspaces (AL-1, AL-2, AL-3), got %d", report.Migrated)
	}
	if len(report.Errors) != 1 {
		t.Errorf("expected 1 error for corrupt AL-4, got %d", len(report.Errors))
	}

	for _, tid := range []string{"AL-1", "AL-2", "AL-3"} {
		path := filepath.Join(wsDir, tid, "agentic_metrics.json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", tid, err)
		}

		// Verify file size is under 5 KB (5120 bytes)
		if len(data) >= 5120 {
			t.Errorf("workspace %s file size %d exceeds 5 KB limit", tid, len(data))
		}

		// Verify zero "error_message" strings in raw JSON
		if strings.Contains(string(data), `"error_message"`) {
			t.Errorf("workspace %s contains persisted error_message keys: %s", tid, string(data))
		}

		var dto struct {
			SchemaVersion *int                      `json:"schema_version"`
			Sessions      []adapters.SessionFactDTO `json:"sessions"`
		}
		if err := json.Unmarshal(data, &dto); err != nil {
			t.Fatalf("unmarshaling %s: %v", tid, err)
		}

		if dto.SchemaVersion == nil || *dto.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected current schema version for %s, got %v", tid, dto.SchemaVersion)
		}

	}
}

func TestMigrationRunner_BatchCurrentMigrationState(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	// Live batch migration directly to schema version 4
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}

	if report.Scanned != 4 {
		t.Errorf("expected 4 scanned workspaces, got %d", report.Scanned)
	}
	if report.Migrated != 3 {
		t.Errorf("expected 3 migrated workspaces, got %d", report.Migrated)
	}

	for _, tid := range []string{"AL-1", "AL-2", "AL-3"} {
		path := filepath.Join(wsDir, tid, "agentic_metrics.json")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", tid, err)
		}

		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			t.Fatalf("unmarshaling %s: %v", tid, err)
		}

		if dto.SchemaVersion == nil || *dto.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected current schema version for %s, got %v", tid, dto.SchemaVersion)
		}
		if dto.Migration == nil {
			t.Errorf("expected non-nil migration info for %s", tid)
		} else {
			hasV4Step := false
			for _, step := range dto.Migration.StepsCompleted {
				if step == domain.MigrationStepUserIterations {
					hasV4Step = true
					break
				}
			}
			if !hasV4Step {
				t.Errorf("expected %s in steps completed for %s", domain.MigrationStepUserIterations, tid)
			}
			hasSourceBackfillPending := false
			for _, step := range dto.Migration.StepsPending {
				if step == domain.MigrationStepSourceBackfill {
					hasSourceBackfillPending = true
					break
				}
			}
			if !hasSourceBackfillPending || dto.Migration.Status != string(domain.MigrationStatusPartial) {
				t.Errorf("expected shape-only migration to remain partial pending source backfill for %s, got %+v", tid, dto.Migration)
			}
		}

		// If the ledger has code.sdlc_iterations > 0 and invocations, primary root invocation must receive user_iterations
		if dto.Code.SDLCIterations > 0 && len(dto.Sessions) > 0 && len(dto.Sessions[0].Invocations) > 0 {
			if dto.Sessions[0].Invocations[0].UserIterations != dto.Code.SDLCIterations {
				t.Errorf("expected primary root invocation to have UserIterations %d, got %d for %s",
					dto.Code.SDLCIterations, dto.Sessions[0].Invocations[0].UserIterations, tid)
			}
		}
	}
}

func TestMigrationRunner_NormalizesInterimV4AndReportsDroppedUnscopedTools(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspaces", "TEST-SESSION")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspace, "agentic_metrics.json")
	original := []byte(`{"schema_version":4,"ticket_id":"TEST-SESSION","ticket_title":"Session provenance validation","code":{"sdlc_iterations":0},"invocations":[],"tools":[{"name":"bash","calls":4,"errors":1}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: filepath.Dir(workspace),
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}
	if report.Migrated != 1 || report.UpToDate != 0 || len(report.Errors) != 0 {
		t.Fatalf("expected interim v4 normalization, got migrated=%d up_to_date=%d errors=%v", report.Migrated, report.UpToDate, report.Errors)
	}
	if len(report.Results) != 1 || !strings.Contains(report.Results[0].DiffSummary, "Removed 1 unscoped legacy tool") || !strings.Contains(report.Results[0].DiffSummary, "Re-extract") {
		t.Fatalf("expected migration report to surface dropped-row count and recovery guidance, got %+v", report.Results)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var normalized adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &normalized); err != nil {
		t.Fatal(err)
	}
	if len(normalized.Sessions) != 0 {
		t.Fatalf("unscoped tools must not fabricate sessions: %#v", normalized.Sessions)
	}
	if normalized.Migration != nil {
		t.Fatalf("completed migration audit belongs in the report, got persisted %#v", normalized.Migration)
	}
	if strings.Contains(string(data), `"tools"`) || strings.Contains(string(data), `"invocations"`) || strings.Contains(string(data), `"events"`) {
		t.Fatalf("normalized v4 must not serialize root fact arrays: %s", data)
	}
}

func TestMigrationRunner_RejectsMixedCanonicalAndLegacyCollectionsAtomically(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspaces", "TEST-MIXED")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspace, "agentic_metrics.json")
	original := []byte(`{"schema_version":4,"ticket_id":"TEST-MIXED","sessions":[],"invocations":[],"tools":[{"session_id":"s1","name":"bash","calls":1,"errors":0}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	if report.Migrated != 0 || len(report.Errors) != 1 || !strings.Contains(report.Errors[0], "mixed ledger shape") {
		t.Fatalf("expected mixed-shape rejection, got %+v", report)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatal("mixed-shape rejection modified source ledger")
	}
}

func TestMigrationRunner_ValidatesCurrentBeforeUpToDateFastPath(t *testing.T) {
	cases := []struct {
		name          string
		sessions      string
		migration     string
		preserveOnly  bool
		extraMember   string
		expectedError string
	}{
		{
			name:     "missing sessions collection",
			sessions: `null`,
		},
		{
			name:     "missing nested collections",
			sessions: `[{"id":"s1"}]`,
		},
		{
			name:        "unknown current root field with completed metadata",
			sessions:    `[]`,
			migration:   `{"schema_version":6,"status":"completed","steps_completed":[],"steps_pending":[]}`,
			extraMember: `,"retained_extension":{"recorded":"keep"}`,
		},
		{
			name:      "completed metadata cannot bypass missing-session validation",
			sessions:  `null`,
			migration: `{"schema_version":6,"status":"completed","steps_completed":[],"steps_pending":[]}`,
		},
		{
			name:         "completed with pending work",
			preserveOnly: true,
			sessions:     `[]`,
			migration:    `{"schema_version":6,"status":"completed","steps_completed":[],"steps_pending":["source_backfill"]}`,
		},
		{
			name:         "partial migration remains resumable",
			preserveOnly: true,
			sessions:     `[]`,
			migration:    `{"schema_version":6,"status":"partial","steps_completed":["schema_v6_lean_occurrences"],"steps_pending":["source_backfill"]}`,
		},
		{
			name:         "pending migration remains resumable",
			preserveOnly: true,
			sessions:     `[]`,
			migration:    `{"schema_version":6,"status":"pending","steps_completed":[],"steps_pending":["source_backfill"]}`,
		},
		{
			name:          "padded session ID",
			expectedError: `sessions[0].id must not have leading or trailing whitespace`,
			sessions:      `[{"id":" s1 ","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[]}]`,
		},
		{
			name:          "padded session parent ID",
			expectedError: `sessions[1].parent_id must not have leading or trailing whitespace`,
			sessions:      `[{"id":"root","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[]},{"id":"child","source":"opencode","user":"fixture","anomalies":[],"parent_id":" root ","invocations":[],"tools":[]}]`,
		},

		{
			name:          "duplicate session IDs",
			expectedError: `sessions[1].id "s1" is duplicated`,
			sessions:      `[{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[]},{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[]}]`,
		},
		{
			name:          "duplicate tool rows",
			expectedError: `sessions[0].tools[1].name "bash" is duplicated`,
			sessions:      `[{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[{"name":"bash","calls":2,"errors":0},{"name":"bash","calls":1,"errors":0}]}]`,
		},
		{
			name:          "invalid tool counts",
			expectedError: `sessions[0].tools[0] has invalid calls/errors counts`,
			sessions:      `[{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"invocations":[],"tools":[{"name":"bash","calls":1,"errors":2}]}]`,
		},
		{
			name:          "dangling session parent",
			expectedError: `sessions[0].parent_id "missing" does not identify a session`,
			sessions:      `[{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"parent_id":"missing","invocations":[],"tools":[]}]`,
		},
		{
			name:          "retired invocation parent reference",
			expectedError: `unknown field "parent_id"`,
			sessions:      `[{"id":"s1","source":"opencode","user":"fixture","anomalies":[],"invocations":[{"parent_id":"missing","role":"coder","model":"m","started_at":null,"completed_at":null,"active_duration_seconds":0,"turns":0,"tokens":{"input":0,"output":0,"cache_read":0,"reasoning":0,"peak_context":0},"cost_usd":0,"compactions":0,"user_iterations":0}],"tools":[]}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workspace := filepath.Join(t.TempDir(), "workspaces", "TEST-CURRENT")
			if err := os.MkdirAll(workspace, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(workspace, "agentic_metrics.json")
			metadataMember := ""
			if tc.migration != "" {
				metadataMember = `,"migration":` + tc.migration
			}
			original := []byte(`{"schema_version":6,"ticket_id":"TEST-CURRENT","sessions":` + tc.sessions + metadataMember + tc.extraMember + `}`)
			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatal(err)
			}

			report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion, Backup: true})
			if err != nil {
				t.Fatalf("unexpected runner error: %v", err)
			}
			if tc.preserveOnly {
				if report.Migrated != 0 {
					t.Fatalf("unfinished migration metadata was cleaned: %+v", report)
				}
			} else if report.UpToDate != 0 || len(report.Errors) != 1 {
				t.Fatalf("expected current-schema validation error, got %+v", report)
			}
			if tc.expectedError != "" && !strings.Contains(strings.Join(report.Errors, "; "), tc.expectedError) {
				t.Fatalf("named validation path was not reached, want %q: %+v", tc.expectedError, report)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) != string(original) {
				t.Fatal("invalid canonical ledger was modified by validation")
			}
			if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
				t.Fatalf("rejected current ledger created a publication backup: %v", err)
			}
		})
	}
}

func TestMigrationCommandsValidateRawCleanupBeforePublication(t *testing.T) {
	completed := `{"schema_version":6,"status":"completed","steps_completed":[],"steps_pending":[]}`
	for _, tc := range []struct {
		name, metadata, sessions, extraMember, diagnostic string
	}{
		{"completed omitted step arrays", `{"schema_version":6,"status":"completed"}`, `[]`, "", ""},
		{"completed null step arrays", `{"schema_version":6,"status":"completed","steps_completed":null,"steps_pending":null}`, `[]`, "", ""},
		{"unsupported status", `{"schema_version":6,"status":"unrecognized","steps_completed":[],"steps_pending":[],"notes":"Schema v6 migration dropped ownerless skills=2 task_links=3"}`, `[]`, "", "migration"},
		{"absent status", `{"schema_version":6,"steps_completed":[],"steps_pending":[],"notes":"Schema v6 migration dropped ownerless skills=2 task_links=3"}`, `[]`, "", "migration"},
		{"duplicate root member", completed, `[{"id":"remote-owner","source":"codex","user":"fixture","invocations":[],"tools":[],"anomalies":[]}]`, `,"sessions":[]`, `duplicate JSON member "sessions"`},
		{"missing recorded owner tools", completed, `[{"id":"remote-owner","source":"codex","user":"fixture","invocations":[],"anomalies":[]}]`, "", "tools"},
	} {
		for _, command := range []struct {
			name string
			run  func(string) (string, error)
		}{
			{"migrate", func(root string) (string, error) {
				report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: root, TargetVersion: domain.CurrentSchemaVersion, Backup: true})
				if err != nil {
					return "", err
				}
				return strings.Join(report.Errors, "; "), nil
			}},
			{"backfill", func(root string) (string, error) {
				report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{WorkspacesDir: root, Source: "codex"})
				if err != nil {
					return "", err
				}
				return strings.Join(report.Warnings, "; "), nil
			}},
		} {
			t.Run(command.name+"/"+tc.name, func(t *testing.T) {
				root := t.TempDir()
				workspace := filepath.Join(root, "TEST-RAW-ADMISSION")
				if err := os.MkdirAll(workspace, 0700); err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(workspace, "agentic_metrics.json")
				before := []byte(`{"schema_version":6,"ticket_id":"TEST-RAW-ADMISSION","sessions":` + tc.sessions + `,"migration":` + tc.metadata + tc.extraMember + `}`)
				if err := os.WriteFile(path, before, 0600); err != nil {
					t.Fatal(err)
				}
				diagnostic, err := command.run(root)
				if err != nil || tc.diagnostic == "" && diagnostic != "" || !strings.Contains(diagnostic, tc.diagnostic) {
					t.Fatalf("raw admission did not report %q: %s, %v", tc.diagnostic, diagnostic, err)
				}
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if tc.diagnostic == "" {
					var expected, actual map[string]json.RawMessage
					if err := json.Unmarshal(before, &expected); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(after, &actual); err != nil {
						t.Fatal(err)
					}
					delete(expected, "migration")
					if len(expected) != len(actual) {
						t.Fatalf("compatible cleanup did not remove only migration: %s", after)
					}
					for key, raw := range expected {
						if !bytes.Equal(raw, actual[key]) {
							t.Fatalf("compatible cleanup changed %q", key)
						}
					}
					if command.name == "migrate" {
						backup, err := os.ReadFile(path + ".bak")
						if err != nil || !bytes.Equal(before, backup) {
							t.Fatalf("compatible cleanup lost exact rollback bytes: %v", err)
						}
					}
					return
				}
				if !bytes.Equal(before, after) {
					t.Fatalf("rejected raw ledger was rewritten: %s", after)
				}
				for _, suffix := range []string{".bak", ".tmp"} {
					if _, err := os.Stat(path + suffix); !os.IsNotExist(err) {
						t.Fatalf("raw admission created publication artifact %s: %v", suffix, err)
					}
				}
			})
		}
	}
}

func TestMigrationRunner_NormalizesPreIterationV4AnomalySummaries(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspaces", "TEST-CANONICAL")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspace, "agentic_metrics.json")
	original := []byte(`{"schema_version":4,"ticket_id":"TEST-CANONICAL","sessions":[{"id":"s1","invocations":[],"tools":[{"name":"bash","calls":2,"errors":0}],"events":[]}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}
	if report.Migrated != 1 || len(report.Errors) != 0 {
		t.Fatalf("expected pre-iteration v4 ledger to be normalized, got %+v", report)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), `"anomalies": []`) {
		t.Fatalf("normalized v4 ledger is missing canonical anomaly arrays: %s", string(after))
	}
}

func TestMigrationRunner_SanitizesCanonicalV4EventPayloads(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspaces", "TEST-PRIVACY")
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(workspace, "agentic_metrics.json")
	original := []byte(`{"schema_version":4,"ticket_id":"TEST-PRIVACY","sessions":[{"id":"s1","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T10:00:00Z","turn":1,"role":"coder","event_type":"tool_call","status":"error","error_message":"private text"}]}]}`)
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: filepath.Dir(workspace), TargetVersion: domain.CurrentSchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	if report.Migrated != 1 || len(report.Errors) != 0 {
		t.Fatalf("expected v4 privacy normalization, got %+v", report)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "private text") || strings.Contains(string(data), `"error_message"`) {
		t.Fatalf("v4 migration persisted sensitive event payload: %s", data)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatal(err)
	}
	if len(dto.Sessions) != 1 || dto.Sessions[0].Anomalies == nil || dto.Anomalies == nil {
		t.Fatalf("expected isolated error omitted and canonical summary arrays present, got %+v", dto)
	}
}

func TestMigrationRunner_MissingWorkspacesDir(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: "",
		TargetVersion: domain.CurrentSchemaVersion,
	})
	if report != nil {
		t.Errorf("expected nil report, got %+v", report)
	}
	if err == nil {
		t.Fatal("expected error when workspaces dir cannot be resolved, got nil")
	}
	if !errors.Is(err, adapters.ErrWorkspacesDirNotFound) {
		t.Errorf("expected ErrWorkspacesDirNotFound, got %v", err)
	}
}

func TestMigrationRunnerRejectsExistingInvalidStateConfigWithoutMutation(t *testing.T) {
	ws := setupMigrationTestWorkspaces(t)
	root := filepath.Dir(ws)
	invalid := []byte("project_name: Example\nticket:\n  prefixes: [AL]\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  error_rate_min_calls: 5\n  error_rate_threshold: 0.25\n  error_rate_critical_threshold: 0.70\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n")
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), invalid, 0600); err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(ws, "AL-1", "agentic_metrics.json")
	before, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatal(err)
	}
	report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: ws, TargetVersion: domain.CurrentSchemaVersion})
	if err == nil || !strings.Contains(err.Error(), "telemetry.capture.state_dir") || report != nil {
		t.Fatalf("invalid target config fell back to caller config: %+v %v", report, err)
	}
	after, err := os.ReadFile(ledgerPath)
	if err != nil || string(before) != string(after) {
		t.Fatalf("rejected migration changed ledger: %v", err)
	}
}

func TestMigrationRunnerRejectsObsoletePublicationTarget(t *testing.T) {
	ws := setupMigrationTestWorkspaces(t)
	path := filepath.Join(ws, "AL-1", "agentic_metrics.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: ws, TargetVersion: 5, DryRun: true})
	if err == nil || report != nil || !strings.Contains(err.Error(), "current target version 6") {
		t.Fatalf("obsolete target not rejected: %+v %v", report, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("obsolete publication target changed ledger: %v", err)
	}
}

func TestMigrationCommandsCanonicalizeCompletedDraftDuration(t *testing.T) {
	positive := `"duration_evidence":{"state":"provisional","completed_seconds":3,"provisional_seconds":7,"cutoff":"2026-01-01T00:00:10Z","basis":"codex:turn"}`
	zero := `"duration_evidence":{"state":"complete","completed_seconds":4,"provisional_seconds":0,"basis":"codex:turn"},"provisional_seconds":0`
	fixture := func(first, second string) []byte {
		return []byte(`{"schema_version":6,"ticket_id":"TEST-DURATION-CLEANUP","ticket_title":"Raw portable history","code":{"lines_added":19},"migration":{"schema_version":6,"status":"completed","steps_completed":[],"steps_pending":[]},"anomalies":[],"sessions":[{"id":"remote","source":"codex","user":"alice","invocations":[{"role":"coder","model":"fixture","started_at":null,"completed_at":null,"active_duration_seconds":10,` + first + `,"turns":2,"tokens":{"input":31,"output":2,"cache_read":0,"reasoning":0,"peak_context":31},"cost_usd":1.2500,"cost_evidence":{"state":"observed","known_subtotal_usd":1.2500},"compactions":0,"user_iterations":1},{"role":"tester","model":"fixture","started_at":null,"completed_at":null,"active_duration_seconds":4` + second + `,"turns":1,"tokens":{"input":7,"output":0,"cache_read":0,"reasoning":0,"peak_context":7},"cost_usd":null,"cost_evidence":{"state":"unavailable"},"compactions":0,"user_iterations":0}],"tools":[{"name":"bash","calls":2,"errors":0}],"skills":[{"name":"kept","caller_role":null,"invoked_at":null},{"name":"kept","caller_role":null,"invoked_at":null}],"anomalies":[]},{"id":"empty","source":"opencode","user":"bob","invocations":[],"tools":[],"skills":[],"delegations":[],"anomalies":[]},{"id":"unrecorded","source":"codex","user":"charlie","invocations":[],"tools":[]}]}`)
	}
	original := fixture(positive, ","+zero)
	expected := fixture(`"provisional_seconds":7`, "")
	var expectedRoot map[string]json.RawMessage
	if err := json.Unmarshal(expected, &expectedRoot); err != nil {
		t.Fatal(err)
	}
	delete(expectedRoot, "migration")
	expected, err := json.Marshal(expectedRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"migrate", "backfill", "backfill_force"} {
		t.Run(command, func(t *testing.T) {
			root := t.TempDir()
			workspace := filepath.Join(root, "TEST-DURATION-CLEANUP")
			if err := os.MkdirAll(workspace, 0700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(workspace, "agentic_metrics.json")
			if err := adapters.ValidateRuntimeFactLedger(original); err != nil {
				t.Fatalf("invalid isolated fixture: %v", err)
			}
			if err := os.WriteFile(path, original, 0600); err != nil {
				t.Fatal(err)
			}
			if command == "migrate" {
				report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: root, TargetVersion: domain.CurrentSchemaVersion, Backup: true})
				if err != nil || report.Migrated != 1 || len(report.Errors) != 0 {
					t.Fatalf("migration cleanup: %+v %v", report, err)
				}
				backup, err := os.ReadFile(path + ".bak")
				if err != nil || !bytes.Equal(backup, original) {
					t.Fatalf("lost exact rollback bytes: %v", err)
				}
			} else {
				report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{WorkspacesDir: root, Source: "codex", Force: command == "backfill_force"})
				if err != nil || report.Updated != 1 || len(report.Warnings) != 0 {
					t.Fatalf("backfill cleanup: %+v %v", report, err)
				}
			}
			written, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := adapters.ValidateRuntimeFactLedger(written); err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(written, []byte(`"duration_evidence"`)) || bytes.Contains(written, []byte(`"completed_seconds"`)) {
				t.Fatalf("successful cleanup published retired duration metadata: %s", written)
			}
			assertRawCleanupJSONEqual(t, expected, written)
			var dto adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(written, &dto); err != nil {
				t.Fatal(err)
			}
			inv := dto.ToDomain().Sessions[0].Invocations
			if len(inv) != 2 || inv[0].ActiveDurationSeconds != 10 || inv[0].ProvisionalSeconds != 7 || inv[1].ActiveDurationSeconds != 4 || inv[1].ProvisionalSeconds != 0 {
				t.Fatalf("cleanup lost recorded numeric portions: %+v", inv)
			}
		})
	}
}

// Compare raw facts recursively while permitting object member reordering.
// Scalar lexemes, nullable members, array multiplicity and optional presence remain exact.
func assertRawCleanupJSONEqual(t *testing.T, expected, actual []byte) {
	t.Helper()
	expected, actual = bytes.TrimSpace(expected), bytes.TrimSpace(actual)
	if len(expected) == 0 || len(actual) == 0 {
		t.Fatalf("missing JSON: want=%s got=%s", expected, actual)
	}
	switch expected[0] {
	case '{':
		var want, got map[string]json.RawMessage
		if err := json.Unmarshal(expected, &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(actual, &got); err != nil {
			t.Fatal(err)
		}
		if len(want) != len(got) {
			t.Fatalf("raw member count changed: want=%s got=%s", expected, actual)
		}
		for key, value := range want {
			received, exists := got[key]
			if !exists {
				t.Fatalf("raw member %s missing: %s", key, actual)
			}
			assertRawCleanupJSONEqual(t, value, received)
		}
	case '[':
		var want, got []json.RawMessage
		if err := json.Unmarshal(expected, &want); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(actual, &got); err != nil {
			t.Fatal(err)
		}
		if len(want) != len(got) {
			t.Fatalf("raw array multiplicity changed: want=%s got=%s", expected, actual)
		}
		for i := range want {
			assertRawCleanupJSONEqual(t, want[i], got[i])
		}
	default:
		if !bytes.Equal(expected, actual) {
			t.Fatalf("raw scalar changed: want=%s got=%s", expected, actual)
		}
	}
}
