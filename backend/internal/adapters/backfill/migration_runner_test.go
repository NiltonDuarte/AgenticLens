package backfill_test

import (
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
		TargetVersion: 2,
	})
	if err != nil {
		t.Fatalf("unexpected runner error: %v", err)
	}

	if report.Scanned != 4 {
		t.Errorf("expected 4 scanned workspaces, got %d", report.Scanned)
	}
	if report.Migrated != 2 {
		t.Errorf("expected 2 migrated workspaces, got %d", report.Migrated)
	}
	if report.UpToDate != 1 {
		t.Errorf("expected 1 up-to-date workspace, got %d", report.UpToDate)
	}
	if len(report.Errors) != 1 {
		t.Errorf("expected 1 non-fatal error for AL-4, got %d", len(report.Errors))
	}

	// Assert AL-1 migrated to schema version 2
	al1Data, err := os.ReadFile(filepath.Join(wsDir, "AL-1", "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var al1DTO adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(al1Data, &al1DTO); err != nil {
		t.Fatalf("failed unmarshaling AL-1: %v", err)
	}
	if al1DTO.SchemaVersion == nil || *al1DTO.SchemaVersion != 2 {
		t.Errorf("expected AL-1 schema version 2, got %v", al1DTO.SchemaVersion)
	}
	if al1DTO.Events == nil {
		t.Errorf("expected AL-1 events array initialized")
	}
	if al1DTO.Code.LinesAdded != 120 {
		t.Errorf("expected AL-1 lines_added 120, got %d", al1DTO.Code.LinesAdded)
	}
	if len(al1DTO.Invocations) != 1 || al1DTO.Invocations[0].ID != "inv-1" {
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
	if al2DTO.SchemaVersion == nil || *al2DTO.SchemaVersion != 2 {
		t.Errorf("expected AL-2 schema version 2, got %v", al2DTO.SchemaVersion)
	}

	// Test idempotency: re-running produces zero migrations and all 3 valid workspaces are up-to-date
	report2, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: 2,
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
		TargetVersion: 2,
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
		TargetVersion: 2,
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
	if dto.SchemaVersion == nil || *dto.SchemaVersion != 2 {
		t.Errorf("expected upgraded schema_version 2, got %v", dto.SchemaVersion)
	}
}

func TestMigrationRunner_TicketFilter(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TicketID:      "AL-1",
		TargetVersion: 2,
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
		TargetVersion:   2,
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
			TargetVersion: 2,
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
			TargetVersion: 2,
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
			TargetVersion: 2,
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
	v2JSON := `{"schema_version": 2, "ticket_id": "AL-UPTODATE", "updated_at": "2026-10-18T10:00:00Z", "events": []}`
	_ = os.WriteFile(filepath.Join(w5Dir, "agentic_metrics.json"), []byte(v2JSON), 0644)

	runner := backfill.NewMigrationRunner(nil)
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: 2,
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
	if valDTO.SchemaVersion == nil || *valDTO.SchemaVersion != 2 {
		t.Errorf("expected schema_version 2, got %v", valDTO.SchemaVersion)
	}
	if valDTO.Events == nil {
		t.Errorf("expected initialized events array")
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
		TargetVersion: 2,
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
		TargetVersion: 2,
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
		TargetVersion: 2,
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
		TargetVersion: 2,
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
			TargetVersion: 2,
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
			TargetVersion: 2,
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
			TargetVersion: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Scanned != 0 || len(report.Results) != 0 {
			t.Errorf("expected 0 scanned, got %d", report.Scanned)
		}
	})
}

func TestMigrationRunner_BatchV3(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	// Step 1: Dry run to v3 should not modify any files on disk
	reportDry, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: 3,
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
	var al3DTODry adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(al3DataDry, &al3DTODry); err != nil {
		t.Fatal(err)
	}
	if al3DTODry.SchemaVersion == nil || *al3DTODry.SchemaVersion != 2 {
		t.Errorf("expected AL-3 to remain version 2 after dry-run, got %v", al3DTODry.SchemaVersion)
	}

	// Step 2: Live migration to v3
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: 3,
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

		var dto adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			t.Fatalf("unmarshaling %s: %v", tid, err)
		}

		if dto.SchemaVersion == nil || *dto.SchemaVersion != 3 {
			t.Errorf("expected schema version 3 for %s, got %v", tid, dto.SchemaVersion)
		}

		// Verify all events have nil ErrorMessage
		for i, ev := range dto.Events {
			if ev.ErrorMessage != nil {
				t.Errorf("workspace %s event[%d] has non-nil ErrorMessage: %s", tid, i, *ev.ErrorMessage)
			}
		}
	}
}

func TestMigrationRunner_BatchV4(t *testing.T) {
	wsDir := setupMigrationTestWorkspaces(t)
	runner := backfill.NewMigrationRunner(nil)

	// Live batch migration directly to schema version 4
	report, err := runner.Run(backfill.MigrationOptions{
		WorkspacesDir: wsDir,
		TargetVersion: 4,
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

		if dto.SchemaVersion == nil || *dto.SchemaVersion != 4 {
			t.Errorf("expected schema version 4 for %s, got %v", tid, dto.SchemaVersion)
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
		}

		// If the ledger has code.sdlc_iterations > 0 and invocations, primary root invocation must receive user_iterations
		if dto.Code.SDLCIterations > 0 && len(dto.Invocations) > 0 {
			if dto.Invocations[0].UserIterations != dto.Code.SDLCIterations {
				t.Errorf("expected primary root invocation to have UserIterations %d, got %d for %s",
					dto.Code.SDLCIterations, dto.Invocations[0].UserIterations, tid)
			}
		}
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
		TargetVersion: 2,
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
