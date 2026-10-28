package backfill

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/codex"
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

func TestBackfillEngine_CleansCompletedMigrationMetadata(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-COMPLETED-METADATA"
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{
		SchemaVersion: &version,
		TicketID:      &ticketID,
		Migration: &domain.MigrationInfo{
			SchemaVersion:  domain.CurrentSchemaVersion,
			Status:         domain.MigrationStatusCompleted,
			StepsCompleted: []string{domain.MigrationStepSourceBackfill},
			StepsPending:   []string{},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	metadataFreeTicket := "AL-COMPLETED-CLEAN"
	createTestWorkspace(t, tempDir, metadataFreeTicket, domain.AgenticFactLedger{
		SchemaVersion: &version,
		TicketID:      &metadataFreeTicket,
	})

	report, err := NewEngine().Execute(BackfillOptions{WorkspacesDir: tempDir})
	if err != nil {
		t.Fatalf("backfill execute failed: %v", err)
	}
	if report.Updated != 1 || report.Skipped != 1 {
		t.Fatalf("expected metadata cleanup and metadata-free skip, got %+v", report)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cleaned adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &cleaned); err != nil {
		t.Fatalf("decode cleaned ledger: %v", err)
	}
	if cleaned.Migration != nil || strings.Contains(string(data), `"migration"`) {
		t.Fatalf("completed migration metadata must be removed: %s", data)
	}
	if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
		t.Fatalf("cleaned schema-v4 ledger should be runtime-ready: %v", err)
	}
}

func TestBackfillEngine_ExplicitRecomputeDBDoesNotSkipCompletedLedger(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-RECOMPUTE-COMPLETED"
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{
		SchemaVersion: &version,
		TicketID:      &ticketID,
		Sessions:      sessionFacts([]domain.InvocationFact{{ID: "inv-existing", Role: "coder", Model: "model"}}, nil, nil),
		Migration: &domain.MigrationInfo{
			SchemaVersion:  domain.CurrentSchemaVersion,
			Status:         domain.MigrationStatusCompleted,
			StepsCompleted: []string{domain.MigrationStepSourceBackfill},
			StepsPending:   []string{},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		RecomputeDB:   true,
		DBPath:        filepath.Join(tempDir, "missing.db"),
		Source:        "opencode",
	})
	if err != nil {
		t.Fatalf("backfill execute failed: %v", err)
	}
	if report.Updated != 1 || report.Skipped != 0 || report.Partial != 1 {
		t.Fatalf("explicit database recompute must process the selected ledger: %+v", report)
	}
	if len(report.Results) != 1 || report.Results[0].Skipped {
		t.Fatalf("selected ledger was silently skipped: %+v", report.Results)
	}
}

func TestBackfillEngine_DryRunPreservesFiles(t *testing.T) {
	tempDir := t.TempDir()
	tid := "AL-101"
	origLedger := domain.AgenticFactLedger{
		TicketID: &tid,
		Code: domain.CodeFacts{
			LinesAdded: 50,
		}, Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-1", Role: "orchestrator", Model: "gpt-4o"},
		}, nil, nil),
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
		TicketID:        tid,
		DryRun:          true,
		RecomputeGit:    true,
		CodeRef:         "HEAD",
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

func TestBackfillEngine_MigrationThenSourceBackfillMakesLedgerReady(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-SOURCE-BACKFILL"
	workspace := filepath.Join(tempDir, ticketID)
	if err := os.MkdirAll(workspace, 0755); err != nil {
		t.Fatal(err)
	}
	metricPath := filepath.Join(workspace, "agentic_metrics.json")
	legacy := `{"schema_version":3,"ticket_id":"AL-SOURCE-BACKFILL","ticket_title":"Source recovery","code":{"sdlc_iterations":0},"sessions":[]}`
	if err := os.WriteFile(metricPath, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := NewMigrationRunner(nil).Run(MigrationOptions{WorkspacesDir: tempDir, TicketID: ticketID, TargetVersion: domain.CurrentSchemaVersion}); err != nil {
		t.Fatalf("explicit schema migration failed: %v", err)
	}
	migrated, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateRuntimeFactLedger(migrated); err == nil {
		t.Fatal("shape-only schema migration must remain ineligible for runtime ingestion")
	}

	dbPath := filepath.Join(tempDir, "opencode.db")
	setupMockOpenCodeDB(t, dbPath, "source-session", ticketID)
	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		RecomputeDB:   true,
		DBPath:        dbPath,
		Source:        "opencode",
	})
	if err != nil {
		t.Fatalf("source-backed backfill failed: %v", err)
	}
	if report.Completed != 1 {
		t.Fatalf("expected source-backed backfill to complete migration, got %+v", report.Results)
	}
	ready, err := os.ReadFile(metricPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := adapters.ValidateRuntimeFactLedger(ready); err != nil {
		t.Fatalf("source-backed result should be runtime-ingestible: %v", err)
	}
}

func TestBackfillEngine_SourceBackfillFailuresRemainPendingAndRuntimeIneligible(t *testing.T) {
	for _, tc := range []struct {
		name       string
		makeSource func(*testing.T, string)
	}{
		{
			name: "no matching source session",
			makeSource: func(t *testing.T, path string) {
				setupMockOpenCodeDB(t, path, "unrelated-session", "OTHER-TICKET")
			},
		},
		{
			name: "source unavailable",
			makeSource: func(t *testing.T, path string) {
				// The missing parent directory makes SQLite unable to open the configured source.
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			ticketID := "AL-SOURCE-FAILURE"
			workspace := filepath.Join(tempDir, ticketID)
			if err := os.MkdirAll(workspace, 0755); err != nil {
				t.Fatal(err)
			}
			metricPath := filepath.Join(workspace, "agentic_metrics.json")
			legacy := `{"schema_version":3,"ticket_id":"AL-SOURCE-FAILURE","code":{"sdlc_iterations":0},"sessions":[]}`
			if err := os.WriteFile(metricPath, []byte(legacy), 0644); err != nil {
				t.Fatal(err)
			}
			if _, err := NewMigrationRunner(nil).Run(MigrationOptions{WorkspacesDir: tempDir, TicketID: ticketID, TargetVersion: domain.CurrentSchemaVersion}); err != nil {
				t.Fatalf("shape-only migration failed: %v", err)
			}

			dbPath := filepath.Join(tempDir, "opencode.db")
			tc.makeSource(t, dbPath)
			report, err := NewEngine().Execute(BackfillOptions{
				WorkspacesDir: tempDir,
				TicketID:      ticketID,
				RecomputeDB:   true,
				DBPath:        dbPath,
				Source:        "opencode",
			})
			if err != nil {
				t.Fatalf("backfill execute failed: %v", err)
			}
			if report.Partial != 1 || report.Completed != 0 {
				t.Fatalf("failed source recovery should report an incomplete migration, got %+v", report)
			}

			persisted, err := os.ReadFile(metricPath)
			if err != nil {
				t.Fatal(err)
			}
			var ledger adapters.AgenticFactLedgerDTO
			if err := json.Unmarshal(persisted, &ledger); err != nil {
				t.Fatalf("decode persisted ledger: %v", err)
			}
			if ledger.Migration == nil || ledger.Migration.Status == string(domain.MigrationStatusCompleted) || !containsMigrationStep(ledger.Migration.StepsPending, domain.MigrationStepSourceBackfill) {
				t.Fatalf("source_backfill must remain persisted as pending after failure: %+v", ledger.Migration)
			}
			if err := adapters.ValidateRuntimeFactLedger(persisted); err == nil {
				t.Fatal("ledger without successful source-backed refresh must remain runtime-ineligible")
			}
		})
	}
}

func TestBackfillEngine_RecomputeDBPreservesExactZeroSteeringTurns(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-ZERO-TURNS"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Code:     domain.CodeFacts{SDLCIterations: 4},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)
	dbPath := filepath.Join(tempDir, "opencode.db")
	setupMockOpenCodeDB(t, dbPath, "sess-zero", ticketID)

	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		RecomputeDB:   true,
		DBPath:        dbPath,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if report.Completed != 1 {
		t.Fatalf("expected one completed workspace, got %d; results=%+v", report.Completed, report.Results)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("read updated ledger: %v", err)
	}
	var updated adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &updated); err != nil {
		t.Fatalf("decode updated ledger: %v", err)
	}
	if updated.Code.SDLCIterations != 0 {
		t.Fatalf("expected exact zero steering turns, got %d", updated.Code.SDLCIterations)
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
		}, Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-a1", Role: "orchestrator", Model: "mock"},
		}, nil, nil),
	}
	ledgerB := domain.AgenticFactLedger{
		TicketID: &tidB,
		Code: domain.CodeFacts{
			LinesAdded: 20,
		}, Sessions: sessionFacts([]domain.InvocationFact{
			{ID: "inv-b1", Role: "coder", Model: "mock"},
		}, nil, nil),
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
		t.Errorf("expected 1 completed (tidA), got %d; results=%+v", reportA.Completed, reportA.Results)
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
	if loadedB.SchemaVersion == nil || *loadedB.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, loadedB.SchemaVersion)
	}
	if loadedB.Migration == nil || loadedB.Migration.Status != string(domain.MigrationStatusPartial) {
		t.Errorf("expected status partial, got %v", loadedB.Migration)
	}
	if err := adapters.ValidateRuntimeFactLedger(dataB); err == nil {
		t.Error("runtime must reject a ledger while its source backfill is pending")
	}
	if len(loadedB.Migration.StepsPending) != 2 || !containsMigrationStep(loadedB.Migration.StepsPending, domain.MigrationStepOpenCodeDB) || !containsMigrationStep(loadedB.Migration.StepsPending, domain.MigrationStepSourceBackfill) {
		t.Errorf("expected steps_pending to contain OpenCode DB and source backfill, got %v", loadedB.Migration.StepsPending)
	}
	if len(loadedB.Sessions) != 1 || len(loadedB.Sessions[0].Invocations) != 1 || loadedB.Sessions[0].Invocations[0].ID != "inv-b1" {
		t.Errorf("expected preserved invocation inv-b1 in session snapshot, got %v", loadedB.Sessions)
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
	if loadedB2.SchemaVersion == nil || *loadedB2.SchemaVersion != domain.CurrentSchemaVersion {
		t.Errorf("expected schema_version %d, got %v", domain.CurrentSchemaVersion, loadedB2.SchemaVersion)
	}
	if loadedB2.Migration != nil {
		t.Errorf("expected final successful backfill to remove migration metadata, got %v", loadedB2.Migration)
	}
	if strings.Contains(string(dataB2), `"migration"`) {
		t.Errorf("final successful backfill must omit migration metadata: %s", dataB2)
	}
	if err := adapters.ValidateRuntimeFactLedger(dataB2); err != nil {
		t.Errorf("final successful backfill should produce a runtime-ready ledger: %v", err)
	}
	if len(loadedB2.Sessions) != 1 || len(loadedB2.Sessions[0].Invocations) != 1 || loadedB2.Sessions[0].Invocations[0].ID != "sess-b1" {
		t.Errorf("expected refreshed invocation sess-b1 in session snapshot, got %v", loadedB2.Sessions)
	}
}

func containsMigrationStep(steps []string, expected string) bool {
	for _, step := range steps {
		if step == expected {
			return true
		}
	}
	return false
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
		Force:         true,
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

	// Good ledgers are shape-migrated despite corrupt siblings, but remain pending
	// until source evidence has actually been extracted.
	for _, goodTid := range []string{tidGood1, tidGood2} {
		data, err := os.ReadFile(filepath.Join(tempDir, goodTid, "agentic_metrics.json"))
		if err != nil {
			t.Fatalf("failed reading migrated good ledger %s: %v", goodTid, err)
		}
		var loaded adapters.AgenticFactLedgerDTO
		if err := json.Unmarshal(data, &loaded); err != nil {
			t.Fatalf("unmarshaling %s: %v", goodTid, err)
		}
		if loaded.SchemaVersion == nil || *loaded.SchemaVersion != domain.CurrentSchemaVersion {
			t.Errorf("expected schema_version %d for %s, got %v", domain.CurrentSchemaVersion, goodTid, loaded.SchemaVersion)
		}
		if loaded.Migration == nil {
			t.Fatalf("expected migration readiness metadata for %s", goodTid)
		}
		if loaded.Migration.Status != string(domain.MigrationStatusPartial) {
			t.Errorf("expected source-backed migration to remain partial for %s, got %s", goodTid, loaded.Migration.Status)
		}
		if !containsMigrationStep(loaded.Migration.StepsPending, domain.MigrationStepSourceBackfill) {
			t.Errorf("expected source_backfill to remain pending for %s, got %v", goodTid, loaded.Migration.StepsPending)
		}
	}
}

func TestBackfillEngine_MissingWorkspacesDir(t *testing.T) {
	tempDir := t.TempDir()
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	engine := NewEngine()
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: "",
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

func TestBackfillEngine_RecomputeGitRequiresTicketID(t *testing.T) {
	tempDir := t.TempDir()
	engine := NewEngine()
	_, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		RecomputeGit:  true,
		CodeRef:       "HEAD",
	})
	if err == nil {
		t.Fatal("expected error when RecomputeGit is true without TicketID, got nil")
	}
	if !strings.Contains(err.Error(), "--ticket-id is required") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestBackfillEngine_SourceValidation(t *testing.T) {
	tempDir := t.TempDir()
	engine := NewEngine()
	_, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		Source:        "invalid-harness",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid source") {
		t.Fatalf("expected invalid source error, got %v", err)
	}
}

func TestBackfillEngine_SourceOpenCodeSkipsCodexSessions(t *testing.T) {
	tempDir := t.TempDir()
	tid := "AL-HYBRID"
	dbPath := filepath.Join(tempDir, "test_opencode.db")
	setupMockOpenCodeDB(t, dbPath, "opencode-root-1", tid)

	origLedger := domain.AgenticFactLedger{
		TicketID: &tid,
		Sessions: []domain.SessionFact{
			{
				ID:     "codex-root-1",
				Source: domain.SourceCodex,
				Invocations: []domain.InvocationFact{
					{ID: "codex-root-1", Role: "orchestrator", Model: "gpt-4o"},
				},
				Tools: []domain.SessionToolFact{
					{Name: "exec", Calls: 3},
				},
			},
			{
				ID:     "opencode-root-1",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{ID: "opencode-root-1", Role: "orchestrator", Model: "mock-model"},
				},
				Tools: []domain.SessionToolFact{},
			},
		},
	}
	createTestWorkspace(t, tempDir, tid, origLedger)

	engine := NewEngine()
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      tid,
		Source:        "opencode",
		RecomputeDB:   true,
		DBPath:        dbPath,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if report.Scanned != 1 || report.Updated != 1 {
		t.Fatalf("expected 1 scanned, 1 updated, got %+v", report)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, tid, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("failed reading result: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed unmarshaling result: %v", err)
	}

	if len(parsed.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(parsed.Sessions))
	}

	var codexSess, opencodeSess *adapters.SessionFactDTO
	for i := range parsed.Sessions {
		if parsed.Sessions[i].ID == "codex-root-1" {
			codexSess = &parsed.Sessions[i]
		} else if parsed.Sessions[i].ID == "opencode-root-1" {
			opencodeSess = &parsed.Sessions[i]
		}
	}

	if codexSess == nil {
		t.Fatalf("codex session was dropped!")
	}
	if codexSess.Source != string(domain.SourceCodex) {
		t.Errorf("expected codex session source %q, got %q", domain.SourceCodex, codexSess.Source)
	}
	if len(codexSess.Tools) != 1 || codexSess.Tools[0].Name != "exec" || codexSess.Tools[0].Calls != 3 {
		t.Errorf("codex session was corrupted: %+v", codexSess)
	}

	if opencodeSess == nil {
		t.Fatalf("opencode session missing!")
	}
	if opencodeSess.Source != string(domain.SourceOpenCode) {
		t.Errorf("expected opencode session source %q, got %q", domain.SourceOpenCode, opencodeSess.Source)
	}
	if len(opencodeSess.Tools) != 1 || opencodeSess.Tools[0].Name != "bash" {
		t.Errorf("expected opencode session to be updated with bash tool, got: %+v", opencodeSess.Tools)
	}
}

func TestBackfillEngine_SourceCodexSkipsOpenCodeSessions(t *testing.T) {
	tempDir := t.TempDir()
	tid := "AL-HYBRID-2"

	origLedger := domain.AgenticFactLedger{
		TicketID: &tid,
		Sessions: []domain.SessionFact{
			{
				ID:     "codex-root-2",
				Source: domain.SourceCodex,
				Invocations: []domain.InvocationFact{
					{ID: "codex-root-2", Role: "orchestrator", Model: "gpt-4o"},
				},
				Tools: []domain.SessionToolFact{
					{Name: "exec", Calls: 5},
				},
			},
			{
				ID:     "opencode-root-2",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{ID: "opencode-root-2", Role: "orchestrator", Model: "claude-3-5-sonnet"},
				},
				Tools: []domain.SessionToolFact{
					{Name: "bash", Calls: 10},
				},
			},
		},
	}
	createTestWorkspace(t, tempDir, tid, origLedger)

	engine := NewEngine()
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      tid,
		Source:        "codex",
		RecomputeDB:   true,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if report.Scanned != 1 {
		t.Fatalf("expected 1 scanned, got %+v", report)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, tid, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("failed reading result: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed unmarshaling result: %v", err)
	}

	var opencodeSess *adapters.SessionFactDTO
	for i := range parsed.Sessions {
		if parsed.Sessions[i].ID == "opencode-root-2" {
			opencodeSess = &parsed.Sessions[i]
		}
	}
	if opencodeSess == nil {
		t.Fatalf("opencode session was dropped during codex backfill!")
	}
	if opencodeSess.Source != string(domain.SourceOpenCode) {
		t.Errorf("expected opencode session source %q, got %q", domain.SourceOpenCode, opencodeSess.Source)
	}
	if len(opencodeSess.Tools) != 1 || opencodeSess.Tools[0].Calls != 10 {
		t.Errorf("opencode session was modified during codex backfill: %+v", opencodeSess)
	}
}

func TestReplaceCodexTreeReplacesRootAndDescendantsTogether(t *testing.T) {
	childParent := "root"
	stored := []domain.SessionFact{
		{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale-root", Calls: 1}}},
		{ID: "child", ParentID: &childParent, Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale-child", Calls: 1}}},
		{ID: "unrelated", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "keep", Calls: 4}}},
	}
	refreshed := []domain.SessionFact{
		{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "fresh-root", Calls: 2}}},
		{ID: "child", ParentID: &childParent, Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "fresh-child", Calls: 3}}},
	}

	replaceCodexTree(&stored, "root", refreshed)

	if len(stored) != 3 {
		t.Fatalf("expected refreshed root and child plus unrelated session, got %d sessions", len(stored))
	}
	byID := make(map[string]domain.SessionFact, len(stored))
	for _, session := range stored {
		if _, duplicate := byID[session.ID]; duplicate {
			t.Fatalf("duplicate session ID %q", session.ID)
		}
		byID[session.ID] = session
	}
	if byID["root"].Tools[0].Name != "fresh-root" || byID["child"].Tools[0].Name != "fresh-child" {
		t.Fatalf("tree was not replaced with the complete refreshed set: %+v", byID)
	}
	if byID["unrelated"].Tools[0].Name != "keep" || byID["unrelated"].Tools[0].Calls != 4 {
		t.Fatalf("unrelated Codex session was modified: %+v", byID["unrelated"])
	}
}

func TestReplaceCodexTreeReplacesSingleSessionWithoutDuplication(t *testing.T) {
	stored := []domain.SessionFact{{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale"}}}}
	refreshed := []domain.SessionFact{{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "fresh"}}}}

	replaceCodexTree(&stored, "root", refreshed)

	if len(stored) != 1 || stored[0].ID != "root" || stored[0].Tools[0].Name != "fresh" {
		t.Fatalf("expected one refreshed root session, got %+v", stored)
	}
}

func TestBackfillEngine_CodexDiscoveryFailurePreservesSessions(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "FIX-8-FAILURE"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "original-root", Calls: 1}}},
			{ID: "unrelated", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "keep", Calls: 2}}},
			{ID: "other-source", Source: domain.SourceOpenCode, Tools: []domain.SessionToolFact{{Name: "unchanged", Calls: 3}}},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)
	engine := NewEngine()
	engine.discoverCodexRollouts = func(codex.DiscoveryOptions) ([]string, error) {
		return nil, errors.New("discovery failed")
	}
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		Source:        "codex",
		CodexHome:     tempDir,
		RecomputeDB:   true,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].DBUpdated || report.Results[0].DBFound {
		t.Fatalf("failed refresh must not report DB update or session found: %+v", report.Results)
	}
	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("failed reading result: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed unmarshaling result: %v", err)
	}
	if len(parsed.Sessions) != 3 {
		t.Fatalf("failed refresh changed session count: %+v", parsed.Sessions)
	}
	wantTools := map[string]string{"root": "original-root", "unrelated": "keep", "other-source": "unchanged"}
	for _, session := range parsed.Sessions {
		want, found := wantTools[session.ID]
		if !found || len(session.Tools) != 1 || session.Tools[0].Name != want {
			t.Errorf("session %q changed after failed refresh: %+v", session.ID, session)
		}
	}
}

func TestBackfillEngine_CodexRefreshesRootTreeOnce(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "FIX-8-TREE"
	childParent := "root"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale-root", Calls: 1}}},
			{ID: "child", ParentID: &childParent, Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "stale-child", Calls: 1}}},
			{ID: "unrelated-codex", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "keep-codex", Calls: 5}}},
			{ID: "opencode", Source: domain.SourceOpenCode, Tools: []domain.SessionToolFact{{Name: "keep", Calls: 7}}},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)
	engine := NewEngine()
	discoveries := 0
	engine.discoverCodexRollouts = func(options codex.DiscoveryOptions) ([]string, error) {
		if options.SessionID == "unrelated-codex" {
			return nil, nil
		}
		discoveries++
		if options.SessionID != "root" {
			t.Fatalf("expected root discovery for child tree, got session %q", options.SessionID)
		}
		return []string{"root-rollout"}, nil
	}
	engine.parseCodexRollout = func(string) (codex.Transcript, error) {
		return codex.Transcript{Invocations: []domain.InvocationFact{
			{ID: "root", Role: "orchestrator", Model: "fresh-model"},
			{ID: "child", ParentID: &childParent, Role: "coder", Model: "fresh-model"},
		}, Tools: []domain.ToolFact{
			{SessionID: "root", Name: "fresh-root", Calls: 2},
			{SessionID: "child", Name: "fresh-child", Calls: 3},
		}}, nil
	}
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		Source:        "codex",
		CodexHome:     tempDir,
		RecomputeDB:   true,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if discoveries != 1 {
		t.Fatalf("expected one rollout tree discovery, got %d", discoveries)
	}
	if len(report.Results) != 1 || !report.Results[0].DBUpdated || !report.Results[0].DBFound {
		t.Fatalf("successful tree refresh was not reported: %+v", report.Results)
	}
	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("failed reading result: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed unmarshaling result: %v", err)
	}
	if len(parsed.Sessions) != 4 {
		t.Fatalf("expected refreshed tree and both unrelated sessions, got %+v", parsed.Sessions)
	}
	want := map[string]string{"root": "fresh-root", "child": "fresh-child", "unrelated-codex": "keep-codex", "opencode": "keep"}
	seen := make(map[string]struct{}, len(want))
	for _, session := range parsed.Sessions {
		tool, found := want[session.ID]
		if !found || len(session.Tools) != 1 || session.Tools[0].Name != tool {
			t.Errorf("unexpected refreshed session %q: %+v", session.ID, session)
		}
		if _, duplicate := seen[session.ID]; duplicate {
			t.Errorf("duplicate session ID %q", session.ID)
		}
		seen[session.ID] = struct{}{}
	}
}

func TestBackfillEngine_CodexParseFailurePreservesTree(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "FIX-8-PARSE-FAILURE"
	childParent := "root"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{ID: "root", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "original-root", Calls: 1}}},
			{ID: "child", ParentID: &childParent, Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "original-child", Calls: 2}}},
			{ID: "unrelated-codex", Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "keep-codex", Calls: 3}}},
			{ID: "opencode", Source: domain.SourceOpenCode, Tools: []domain.SessionToolFact{{Name: "keep-opencode", Calls: 4}}},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)
	engine := NewEngine()
	engine.discoverCodexRollouts = func(codex.DiscoveryOptions) ([]string, error) {
		return []string{"root-rollout"}, nil
	}
	engine.parseCodexRollout = func(string) (codex.Transcript, error) {
		return codex.Transcript{}, errors.New("parse failed")
	}
	report, err := engine.Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		Source:        "codex",
		CodexHome:     tempDir,
		RecomputeDB:   true,
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if len(report.Results) != 1 || report.Results[0].DBUpdated || report.Results[0].DBFound {
		t.Fatalf("failed refresh must not report DB update or session found: %+v", report.Results)
	}
	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("failed reading result: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed unmarshaling result: %v", err)
	}
	want := map[string]string{"root": "original-root", "child": "original-child", "unrelated-codex": "keep-codex", "opencode": "keep-opencode"}
	if len(parsed.Sessions) != len(want) {
		t.Fatalf("parse failure changed session count: %+v", parsed.Sessions)
	}
	for _, session := range parsed.Sessions {
		tool, found := want[session.ID]
		if !found || len(session.Tools) != 1 || session.Tools[0].Name != tool {
			t.Errorf("session %q changed after parse failure: %+v", session.ID, session)
		}
	}
}
