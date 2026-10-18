package backfill

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

func createTestWorkspace(t *testing.T, baseDir string, ticketID string, ledger domain.AgenticFactLedger) string {
	t.Helper()
	wsDir := filepath.Join(baseDir, ticketID)
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed to create test workspace dir: %v", err)
	}

	dto := adapters.AgenticFactLedgerDTOFromDomain(ledger)
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal ledger: %v", err)
	}

	metricPath := filepath.Join(wsDir, "agentic_metrics.json")
	if err := os.WriteFile(metricPath, data, 0644); err != nil {
		t.Fatalf("failed to write metric file: %v", err)
	}

	// Also write minimal CONTRACT.md and SUMMARY.md
	contractPath := filepath.Join(wsDir, "CONTRACT.md")
	_ = os.WriteFile(contractPath, []byte("# Contract "+ticketID), 0644)
	summaryPath := filepath.Join(wsDir, "SUMMARY.md")
	_ = os.WriteFile(summaryPath, []byte("# Summary "+ticketID+"\n\n### Workflow Efficiency\nInitial table\n"), 0644)

	return wsDir
}

func setupMockOpenCodeDB(t *testing.T, dbPath string, sessionID string, ticketID string) {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create mock sqlite db: %v", err)
	}
	defer db.Close()

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
		data TEXT,
		time_created INTEGER,
		time_updated INTEGER
	);
	CREATE TABLE part (
		id TEXT PRIMARY KEY,
		session_id TEXT,
		message_id TEXT,
		data TEXT,
		time_created INTEGER,
		time_updated INTEGER
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to initialize mock schema: %v", err)
	}

	now := time.Now().UnixMilli()
	// Insert root session
	title := fmt.Sprintf("[%s] Test Ticket Session", ticketID)
	_, err = db.Exec(`INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES (?, NULL, 'orchestrator', 'mock-model', ?, ?, ?)`,
		sessionID, title, now-10000, now)
	if err != nil {
		t.Fatalf("failed to insert mock session: %v", err)
	}

	// Insert assistant message with tokens
	msgData := fmt.Sprintf(`{"role":"assistant","modelID":"gpt-4o","tokens":{"input":1000,"output":200,"cache":{"read":500},"reasoning":50},"cost":0.01,"time":{"created":%d,"completed":%d}}`,
		now-9000, now-8000)
	_, err = db.Exec(`INSERT INTO message (id, session_id, data, time_created, time_updated) VALUES ('m1', ?, ?, ?, ?)`, sessionID, msgData, now-9000, now-8000)
	if err != nil {
		t.Fatalf("failed to insert mock message: %v", err)
	}

	// Insert tool call part
	partData := `{"type":"tool","tool":"bash","state":{"status":"completed","output":"success"}}`
	_, err = db.Exec(`INSERT INTO part (id, session_id, message_id, data, time_created, time_updated) VALUES ('p1', ?, 'm1', ?, ?, ?)`, sessionID, partData, now-8500, now-8000)
	if err != nil {
		t.Fatalf("failed to insert mock part: %v", err)
	}
}

func TestBackfillEngine_DryRunPreservesFiles(t *testing.T) {
	tempDir := t.TempDir()
	tid := "AL-101"
	origLedger := domain.AgenticFactLedger{
		TicketID: &tid,
		Code: domain.CodeFacts{
			LinesAdded: 50,
		},
		Invocations: []domain.InvocationFact{
			{ID: "inv-1", Role: "orchestrator", Model: "gpt-4o"},
		},
	}
	createTestWorkspace(t, tempDir, tid, origLedger)

	metricPath := filepath.Join(tempDir, tid, "agentic_metrics.json")
	summaryPath := filepath.Join(tempDir, tid, "SUMMARY.md")
	origMetricData, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatalf("failed to read orig metric file: %v", err)
	}
	origSummaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("failed to read orig summary file: %v", err)
	}

	engine := NewEngine()
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir:   tempDir,
		DryRun:          true,
		RecomputeGit:    true,
		UpdateSummaries: true,
	})
	if err != nil {
		t.Fatalf("execute dry-run failed: %v", err)
	}

	if report.Scanned != 1 {
		t.Errorf("expected 1 scanned, got %d", report.Scanned)
	}

	// Verify metric file on disk is strictly byte-for-byte identical (no file modified)
	postMetricData, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatalf("reading post dry-run metric file: %v", err)
	}
	if string(postMetricData) != string(origMetricData) {
		t.Errorf("dry-run violated invariant: metric file was modified")
	}

	// Verify summary file on disk is also strictly byte-for-byte unchanged
	postSummaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("reading post dry-run summary file: %v", err)
	}
	if string(postSummaryData) != string(origSummaryData) {
		t.Errorf("dry-run violated invariant: summary file was modified")
	}

	var loaded adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(postMetricData, &loaded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if loaded.SchemaVersion != nil {
		t.Errorf("expected SchemaVersion nil in dry-run, got %v", *loaded.SchemaVersion)
	}
}

func TestBackfillEngine_MultiMachineLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	tidA := "AL-AUTHORED-ON-A"
	tidB := "AL-AUTHORED-ON-B"

	// Initial unversioned ledgers with existing facts
	ledgerA := domain.AgenticFactLedger{
		TicketID: &tidA,
		Code: domain.CodeFacts{
			LinesAdded: 10,
		},
		Invocations: []domain.InvocationFact{
			{ID: "inv-a1", Role: "orchestrator", Model: "mock"},
		},
	}
	ledgerB := domain.AgenticFactLedger{
		TicketID: &tidB,
		Code: domain.CodeFacts{
			LinesAdded: 20,
		},
		Invocations: []domain.InvocationFact{
			{ID: "inv-b1", Role: "coder", Model: "mock"},
		},
	}

	createTestWorkspace(t, tempDir, tidA, ledgerA)
	createTestWorkspace(t, tempDir, tidB, ledgerB)

	// Machine A has SQLite DB only for tidA
	dbPathA := filepath.Join(tempDir, "machine_a.db")
	setupMockOpenCodeDB(t, dbPathA, "sess-a1", tidA)

	engine := NewEngine()

	// Step 1: Run backfill on Machine A
	reportA, err := engine.Execute(BackfillOptions{
		WorkspacesDir:   tempDir,
		RecomputeGit:    false,
		RecomputeDB:     true,
		DBPath:          dbPathA,
		UpdateSummaries: true,
	})
	if err != nil {
		t.Fatalf("machine A backfill failed: %v", err)
	}

	if reportA.Scanned != 2 {
		t.Errorf("expected 2 scanned, got %d", reportA.Scanned)
	}
	if reportA.Completed != 1 {
		t.Errorf("expected 1 completed (tidA), got %d", reportA.Completed)
	}
	if reportA.Partial != 1 {
		t.Errorf("expected 1 partial (tidB), got %d", reportA.Partial)
	}

	// Verify tidB preserved existing invocations and was marked partial with pending steps ["opencode_db"]
	dataB, err := os.ReadFile(filepath.Join(tempDir, tidB, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading tidB file: %v", err)
	}
	var loadedB adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(dataB, &loadedB); err != nil {
		t.Fatalf("unmarshaling tidB: %v", err)
	}
	if loadedB.SchemaVersion == nil || *loadedB.SchemaVersion != 2 {
		t.Errorf("expected schema_version 2, got %v", loadedB.SchemaVersion)
	}
	if loadedB.Migration == nil || loadedB.Migration.Status != string(domain.MigrationStatusPartial) {
		t.Errorf("expected status partial, got %v", loadedB.Migration)
	}
	if len(loadedB.Migration.StepsPending) != 1 || loadedB.Migration.StepsPending[0] != domain.MigrationStepOpenCodeDB {
		t.Errorf("expected steps_pending to contain %q, got %v", domain.MigrationStepOpenCodeDB, loadedB.Migration.StepsPending)
	}
	if len(loadedB.Invocations) != 1 || loadedB.Invocations[0].ID != "inv-b1" {
		t.Errorf("expected preserved invocation inv-b1, got %v", loadedB.Invocations)
	}

	// Step 2: Developer B on Machine B pulls repo and resumes migration
	// Machine B has SQLite DB for tidB
	dbPathB := filepath.Join(tempDir, "machine_b.db")
	setupMockOpenCodeDB(t, dbPathB, "sess-b1", tidB)

	reportB, err := engine.Execute(BackfillOptions{
		WorkspacesDir:   tempDir,
		RecomputeGit:    false,
		RecomputeDB:     true,
		DBPath:          dbPathB,
		UpdateSummaries: true,
	})
	if err != nil {
		t.Fatalf("machine B backfill failed: %v", err)
	}

	// tidA is already completed, so it is skipped. tidB is updated from partial to completed!
	if reportB.Skipped != 1 {
		t.Errorf("expected 1 skipped (tidA), got %d", reportB.Skipped)
	}
	if reportB.Completed != 1 {
		t.Errorf("expected 1 completed (tidB), got %d", reportB.Completed)
	}

	// Verify tidB is now completed with fresh invocations from machine B db
	dataB2, err := os.ReadFile(filepath.Join(tempDir, tidB, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading tidB2 file: %v", err)
	}
	var loadedB2 adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(dataB2, &loadedB2); err != nil {
		t.Fatalf("unmarshaling tidB2: %v", err)
	}
	if loadedB2.SchemaVersion == nil || *loadedB2.SchemaVersion != 2 {
		t.Errorf("expected schema_version 2, got %v", loadedB2.SchemaVersion)
	}
	if loadedB2.Migration != nil {
		t.Errorf("expected migration object to be omitted on completed status, got %v", loadedB2.Migration)
	}
	if len(loadedB2.Invocations) != 1 || loadedB2.Invocations[0].ID != "sess-b1" {
		t.Errorf("expected refreshed invocation sess-b1 from DB, got %v", loadedB2.Invocations)
	}
}

func TestBackfillEngine_CorruptedFileResilience(t *testing.T) {
	tempDir := t.TempDir()
	tidGood1 := "AL-GOOD-1"
	tidBad1 := "AL-BAD-TRUNCATED"
	tidBad2 := "AL-BAD-EMPTY"
	tidBad3 := "AL-BAD-SCHEMA"
	tidGood2 := "AL-GOOD-2"

	goodLedger1 := domain.AgenticFactLedger{
		TicketID: &tidGood1,
		Code:     domain.CodeFacts{LinesAdded: 5},
	}
	goodLedger2 := domain.AgenticFactLedger{
		TicketID: &tidGood2,
		Code:     domain.CodeFacts{LinesAdded: 15},
	}
	createTestWorkspace(t, tempDir, tidGood1, goodLedger1)
	createTestWorkspace(t, tempDir, tidGood2, goodLedger2)

	// Write various hostile/unparseable json files
	badDir1 := filepath.Join(tempDir, tidBad1)
	_ = os.MkdirAll(badDir1, 0755)
	_ = os.WriteFile(filepath.Join(badDir1, "agentic_metrics.json"), []byte(`{"ticket_id": "AL-BAD-TRUNCATED", "code": {`), 0644)

	badDir2 := filepath.Join(tempDir, tidBad2)
	_ = os.MkdirAll(badDir2, 0755)
	_ = os.WriteFile(filepath.Join(badDir2, "agentic_metrics.json"), []byte(``), 0644)

	badDir3 := filepath.Join(tempDir, tidBad3)
	_ = os.MkdirAll(badDir3, 0755)
	_ = os.WriteFile(filepath.Join(badDir3, "agentic_metrics.json"), []byte(`{"ticket_id": 12345, "code": "invalid_code_type"}`), 0644)

	engine := NewEngine()
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		RecomputeGit:  true,
	})
	if err != nil {
		t.Fatalf("engine execute should not fail fatally on corrupt files: %v", err)
	}

	if report.Scanned != 5 {
		t.Errorf("expected 5 scanned, got %d", report.Scanned)
	}
	if report.Updated != 2 {
		t.Errorf("expected 2 updated, got %d", report.Updated)
	}
	if len(report.Warnings) != 3 {
		t.Errorf("expected 3 warnings for bad files, got %d: %v", len(report.Warnings), report.Warnings)
	}

	// Verify good ledgers were both migrated to version 2 despite corrupt files in batch
	for _, goodTid := range []string{tidGood1, tidGood2} {
		data, err := os.ReadFile(filepath.Join(tempDir, goodTid, "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading migrated good ledger %s: %v", goodTid, err)
		}
		var loaded adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("unmarshaling %s: %v", goodTid, err)
		}
		if loaded.SchemaVersion == nil || *loaded.SchemaVersion != 2 {
			t.Errorf("expected schema_version 2 for %s, got %v", goodTid, loaded.SchemaVersion)
		}
		if loaded.Migration != nil {
			t.Errorf("expected migration object to be omitted on completed status for %s, got %v", goodTid, loaded.Migration)
		}
	}
}
