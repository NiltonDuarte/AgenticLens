package backfill

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	for _, metadata := range []bool{false, true} {
		for _, available := range []bool{false, true} {
			t.Run(fmt.Sprintf("metadata=%t/source=%t", metadata, available), func(t *testing.T) {
				root := t.TempDir()
				ticket := "CUSTOM-RECOMPUTE"
				version := domain.CurrentSchemaVersion
				ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: sessionFactsWithInvs([]testInvFact{{SessionID: "inv-existing", Role: "coder", Model: "stale"}}, nil, nil)}
				if metadata {
					ledger.Migration = &domain.MigrationInfo{SchemaVersion: version, Status: domain.MigrationStatusCompleted, StepsCompleted: []string{domain.MigrationStepSourceBackfill}, StepsPending: []string{}}
				}
				workspace := createTestWorkspace(t, root, ticket, ledger)
				dbPath := filepath.Join(root, "source.db")
				if available {
					setupMockOpenCodeDB(t, dbPath, "inv-existing", ticket)
				}
				report, err := NewEngine().Execute(BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "opencode", RecomputeDB: true, DBPath: dbPath})
				if err != nil {
					t.Fatal(err)
				}
				if len(report.Results) != 1 || report.Results[0].Error != nil || report.Results[0].Skipped {
					t.Fatalf("explicit refresh failed/skipped: %+v", report.Results)
				}
				updated, data := readRecoveryLedger(t, workspace)
				if updated.Migration != nil {
					t.Fatalf("optional refresh created migration requirement: %+v", updated.Migration)
				}
				if err := adapters.ValidateRuntimeFactLedger(data); err != nil {
					t.Fatalf("canonical baseline became ineligible: %v", err)
				}
				if available {
					if !report.Results[0].DBUpdated || updated.Sessions[0].Invocations[0].Tokens.Input != 1000 {
						t.Fatalf("full cumulative snapshot not refreshed: %+v", updated.Sessions)
					}
				} else {
					if !report.Results[0].RefreshIncomplete || len(report.Results[0].Diagnostics) == 0 || updated.Sessions[0].Invocations[0].Model != "stale" {
						t.Fatalf("optional failure not preserved/reported: %+v", report.Results)
					}
				}
				routine, err := NewEngine().Execute(BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "opencode", DBPath: dbPath})
				if err != nil || routine.Skipped != 1 {
					t.Fatalf("routine completed skip lost: %+v %v", routine, err)
				}
				_, after := readRecoveryLedger(t, workspace)
				if string(after) != string(data) {
					t.Fatal("routine skip rewrote ledger")
				}
			})
		}
	}
}

func TestBackfillEngine_DryRunPreservesFiles(t *testing.T) {
	tempDir := t.TempDir()
	tid := "AL-101"
	origLedger := domain.AgenticFactLedger{
		TicketID: &tid,
		Code: domain.CodeFacts{
			LinesAdded: 50,
		}, Sessions: sessionFactsWithInvs([]testInvFact{
			{SessionID: "inv-1", Role: "orchestrator", Model: "gpt-4o"},
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
	legacy := `{"schema_version":3,"ticket_id":"AL-SOURCE-BACKFILL","ticket_title":"Source recovery","code":{"sdlc_iterations":0},"sessions":[{"id":"source-session","source":"opencode","user":"unknown","invocations":[],"tools":[]}]}`
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
		WorkspacesDir:  tempDir,
		TicketID:       ticketID,
		RecomputeDB:    true,
		DBPath:         dbPath,
		Source:         "opencode",
		AllowFallbacks: true,
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

func TestBackfillEngine_RecomputeDBPreservesGitAndExactZeroInvocationSteering(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-ZERO-TURNS"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Code:     domain.CodeFacts{SDLCIterations: 4},
		Sessions: sessionFactsWithInvs([]testInvFact{
			{SessionID: "sess-zero", Role: "orchestrator", Model: "mock"},
		}, nil, nil),
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
	if updated.Code.SDLCIterations != 4 {
		t.Fatalf("database-only recompute changed Git facts: %d", updated.Code.SDLCIterations)
	}
	if len(updated.Sessions) != 1 || len(updated.Sessions[0].Invocations) != 1 || updated.Sessions[0].Invocations[0].UserIterations != 0 {
		t.Fatalf("source-confirmed zero invocation steering changed: %+v", updated.Sessions)
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
		}, Sessions: sessionFactsWithInvs([]testInvFact{
			{SessionID: "sess-a1", Role: "orchestrator", Model: "mock"},
		}, nil, nil),
	}
	ledgerB := domain.AgenticFactLedger{
		TicketID: &tidB,
		Code: domain.CodeFacts{
			LinesAdded: 20,
		}, Sessions: sessionFactsWithInvs([]testInvFact{
			{SessionID: "sess-b1", Role: "coder", Model: "mock"},
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
	if len(loadedB.Sessions) != 1 || len(loadedB.Sessions[0].Invocations) != 1 {
		t.Errorf("expected preserved invocation sess-b1 in session snapshot, got %v", loadedB.Sessions)
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

	// Explicit recompute processes both; unavailable A preserves its eligible baseline while B completes.
	if reportB.Skipped != 0 || reportB.Completed != 2 || reportB.RefreshIncomplete != 1 {
		t.Errorf("expected B recovery and eligible A optional failure, got %+v", reportB)
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
	if len(loadedB2.Sessions) != 1 || len(loadedB2.Sessions[0].Invocations) != 1 || loadedB2.Sessions[0].ID != "sess-b1" {
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
					{Role: "orchestrator", Model: "gpt-4o"},
				},
				Tools: []domain.SessionToolFact{
					{Name: "exec", Calls: 3},
				},
			},
			{
				ID:     "opencode-root-1",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "mock-model"},
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
					{Role: "orchestrator", Model: "gpt-4o"},
				},
				Tools: []domain.SessionToolFact{
					{Name: "exec", Calls: 5},
				},
			},
			{
				ID:     "opencode-root-2",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "claude-3-5-sonnet"},
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
		return codex.CombineTranscripts([]codex.Transcript{
			{
				ThreadID:    "root",
				Invocations: []domain.InvocationFact{{Role: "orchestrator", Model: "fresh-model"}},
				Tools:       []domain.ToolFact{{SessionID: "root", Name: "fresh-root", Calls: 2}},
			},
			{
				ThreadID: "child", ParentThread: "root",
				Invocations: []domain.InvocationFact{{Role: "coder", Model: "fresh-model"}},
				Tools:       []domain.ToolFact{{SessionID: "child", Name: "fresh-child", Calls: 3}},
			},
		})
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
	if len(report.Results) != 1 || !report.Results[0].DBUpdated || report.Results[0].DBFound || !report.Results[0].RefreshIncomplete {
		t.Fatalf("independent tree update must not certify unavailable sibling recovery: %+v", report.Results)
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

func setupMockOpenCodeMultiRootDB(t *testing.T, dbPath string, rootIDs []string, ticketID string) {
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
	for i, sid := range rootIDs {
		title := fmt.Sprintf("[%s] Test Session %d", ticketID, i+1)
		_, err = db.Exec(`INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated) VALUES (?, NULL, 'orchestrator', 'mock-model', ?, ?, ?)`,
			sid, title, now-int64(10000*(i+1)), now-int64(1000*(i+1)))
		if err != nil {
			t.Fatalf("failed to insert mock session: %v", err)
		}
		msgID := fmt.Sprintf("m-%d", i+1)
		msgData := fmt.Sprintf(`{"role":"assistant","modelID":"gpt-4o","tokens":{"input":1000,"output":200,"cache":{"read":500},"reasoning":50},"cost":0.05,"time":{"created":%d,"completed":%d}}`,
			now-9000, now-8000)
		_, err = db.Exec(`INSERT INTO message (id, session_id, data, time_created, time_updated) VALUES (?, ?, ?, ?, ?)`, msgID, sid, msgData, now-9000, now-8000)
		if err != nil {
			t.Fatalf("failed to insert mock message: %v", err)
		}
	}
}

func TestBackfillEngine_MultiRootSessionsExtracted(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-MULTI-ROOT"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{
				ID:     "ses-root-1",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "mock-model"},
				},
			},
			{
				ID:     "ses-root-2",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "mock-model"},
				},
			},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	dbPath := filepath.Join(tempDir, "opencode.db")
	setupMockOpenCodeMultiRootDB(t, dbPath, []string{"ses-root-1", "ses-root-2"}, ticketID)

	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir: tempDir,
		TicketID:      ticketID,
		RecomputeDB:   true,
		DBPath:        dbPath,
		Source:        "opencode",
		Force:         true,
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	res := report.Results[0]
	if !res.DBFound {
		t.Errorf("expected DBFound true, got false")
	}
	if !res.DBUpdated {
		t.Errorf("expected DBUpdated true, got false")
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading updated ledger: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshaling updated ledger: %v", err)
	}

	if len(parsed.Sessions) != 2 {
		t.Fatalf("expected 2 sessions preserved, got %d: %+v", len(parsed.Sessions), parsed.Sessions)
	}
	sessionIDs := make(map[string]bool)
	for _, s := range parsed.Sessions {
		sessionIDs[s.ID] = true
	}
	if !sessionIDs["ses-root-1"] || !sessionIDs["ses-root-2"] {
		t.Errorf("expected both ses-root-1 and ses-root-2 in sessions, got %v", sessionIDs)
	}
}

func TestBackfillEngine_StrictSessionTrustRejectsUntrustedTitleFallback(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-REJECT-FALLBACK"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{
				ID:     "ses-orig-123",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "coder", Model: "mock-model"},
				},
			},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	dbPath := filepath.Join(tempDir, "opencode.db")
	// DB has an unrelated review session matching the ticket title, but NOT matching the session ID
	setupMockOpenCodeDB(t, dbPath, "ses-review-456", ticketID)

	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir:  tempDir,
		TicketID:       ticketID,
		RecomputeDB:    true,
		DBPath:         dbPath,
		Source:         "opencode",
		Force:          true,
		AllowFallbacks: false, // Default: no fallbacks
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	res := report.Results[0]
	if res.DBFound {
		t.Errorf("expected DBFound false when session ID is missing without fallbacks, got true")
	}
	if res.DBUpdated {
		t.Errorf("expected DBUpdated false, got true")
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading updated ledger: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshaling updated ledger: %v", err)
	}

	// Must preserve original session ID and NOT hijack review session
	if len(parsed.Sessions) != 1 || parsed.Sessions[0].ID != "ses-orig-123" {
		t.Fatalf("expected original session ses-orig-123 preserved intact, got %+v", parsed.Sessions)
	}
	if parsed.Migration == nil || parsed.Migration.Status != string(domain.MigrationStatusPartial) {
		t.Errorf("expected migration status partial, got %v", parsed.Migration)
	}
}

func TestBackfillEngine_ExplicitFallbackResolvesWhenAllowed(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "AL-EXPLICIT-FALLBACK"
	ledger := domain.AgenticFactLedger{
		TicketID: &ticketID,
		Sessions: []domain.SessionFact{
			{
				ID:     "ses-missing-789",
				Source: domain.SourceOpenCode,
				Invocations: []domain.InvocationFact{
					{Role: "coder", Model: "mock-model"},
				},
			},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	dbPath := filepath.Join(tempDir, "opencode.db")
	setupMockOpenCodeDB(t, dbPath, "ses-review-999", ticketID)

	report, err := NewEngine().Execute(BackfillOptions{
		WorkspacesDir:  tempDir,
		TicketID:       ticketID,
		RecomputeDB:    true,
		DBPath:         dbPath,
		Source:         "opencode",
		Force:          true,
		AllowFallbacks: true, // Explicitly enabled
	})
	if err != nil {
		t.Fatalf("backfill failed: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	res := report.Results[0]
	if res.DBFound || !res.RefreshIncomplete {
		t.Errorf("heuristic enrichment must not certify missing exact history: %+v", res)
	}
	if !res.DBUpdated {
		t.Errorf("expected DBUpdated true with AllowFallbacks=true, got false")
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading updated ledger: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshaling updated ledger: %v", err)
	}

	ids := map[string]bool{}
	for _, session := range parsed.Sessions {
		ids[session.ID] = true
	}
	if len(parsed.Sessions) != 2 || !ids["ses-review-999"] || !ids["ses-missing-789"] {
		t.Fatalf("fallback must enrich without erasing missing exact owner: %+v", parsed.Sessions)
	}
}

func TestBackfillEngine_CodexPreservesChildParentIDAndUser(t *testing.T) {
	tempDir := t.TempDir()
	ticketID := "FIX-10-CODEX-HIERARCHY"
	parentID := "codex-root"
	version := domain.CurrentSchemaVersion
	ledger := domain.AgenticFactLedger{
		SchemaVersion: &version,
		TicketID:      &ticketID,
		Sessions: []domain.SessionFact{
			{
				ID:     "codex-root",
				Source: domain.SourceCodex,
				User:   "dev-lead",
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "gpt-4"},
				},
				Tools: []domain.SessionToolFact{{Name: "bash", Calls: 1}},
			},
			{
				ID:       "codex-child",
				ParentID: &parentID,
				Source:   domain.SourceCodex,
				User:     "dev-subagent",
				Invocations: []domain.InvocationFact{
					{Role: "coder", Model: "gpt-4"},
				},
				Tools: []domain.SessionToolFact{{Name: "edit", Calls: 2}},
			},
		},
	}
	createTestWorkspace(t, tempDir, ticketID, ledger)

	engine := NewEngine()
	engine.discoverCodexRollouts = func(options codex.DiscoveryOptions) ([]string, error) {
		if options.SessionID == "codex-root" {
			return []string{"root-path", "child-path"}, nil
		}
		return nil, nil
	}
	engine.parseCodexRollout = func(path string) (codex.Transcript, error) {
		if path == "root-path" {
			return codex.Transcript{
				ThreadID:  "codex-root",
				SessionID: "codex-root",
				Invocations: []domain.InvocationFact{
					{Role: "orchestrator", Model: "gpt-4o"},
				},
				Tools: []domain.ToolFact{
					{SessionID: "codex-root", Name: "bash", Calls: 3},
				},
			}, nil
		}
		return codex.Transcript{
			ThreadID:     "codex-child",
			SessionID:    "codex-child",
			ParentThread: "codex-root",
			Invocations: []domain.InvocationFact{
				{Role: "coder", Model: "gpt-4o"},
			},
			Tools: []domain.ToolFact{
				{SessionID: "codex-child", Name: "edit", Calls: 5},
			},
		}, nil
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
	if len(report.Results) != 1 || !report.Results[0].DBUpdated || !report.Results[0].DBFound {
		t.Fatalf("expected DBUpdated and DBFound true, got: %+v", report.Results)
	}

	data, err := os.ReadFile(filepath.Join(tempDir, ticketID, "agentic_metrics.json"))
	if err != nil {
		t.Fatalf("reading updated ledger: %v", err)
	}
	var parsed adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshaling updated ledger: %v", err)
	}

	if len(parsed.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(parsed.Sessions))
	}
	byID := make(map[string]adapters.SessionFactDTO)
	for _, s := range parsed.Sessions {
		byID[s.ID] = s
	}

	rootSess, hasRoot := byID["codex-root"]
	if !hasRoot {
		t.Fatalf("missing codex-root in refreshed sessions: %+v", parsed.Sessions)
	}
	if rootSess.ParentID != nil {
		t.Errorf("expected root ParentID to be nil, got %v", *rootSess.ParentID)
	}
	if rootSess.User != "dev-lead" {
		t.Errorf("expected root User='dev-lead', got %q", rootSess.User)
	}

	childSess, hasChild := byID["codex-child"]
	if !hasChild {
		t.Fatalf("missing codex-child in refreshed sessions: %+v", parsed.Sessions)
	}
	if childSess.ParentID == nil || *childSess.ParentID != "codex-root" {
		t.Fatalf("expected child ParentID='codex-root', got %v", childSess.ParentID)
	}
	if childSess.User != "dev-subagent" {
		t.Errorf("expected child User='dev-subagent', got %q", childSess.User)
	}
}

func readRecoveryLedger(t *testing.T, workspace string) (adapters.AgenticFactLedgerDTO, []byte) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(workspace, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	var dto adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		t.Fatal(err)
	}
	return dto, data
}

func TestBackfillEngine_SkillTaskRecoverySnapshots(t *testing.T) {
	for _, tc := range []struct {
		name                                                                                             string
		missingRoot, missingChild, malformed, emptyOwners, emptyObservations, dryRun, blockWrite, cyclic bool
	}{
		{name: "complete"}, {name: "confirmed empty", emptyObservations: true}, {name: "missing root", missingRoot: true},
		{name: "missing child", missingChild: true}, {name: "late malformed part", malformed: true}, {name: "unknown owners", emptyOwners: true},
		{name: "dry run incomplete", missingRoot: true, dryRun: true}, {name: "temporary write failure", blockWrite: true}, {name: "unrooted cycle", cyclic: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			ticket := "CUSTOM-SNAPSHOT"
			v := domain.CurrentSchemaVersion
			oldSkill := domain.SkillFact{Name: "historical", TurnIndex: 1, Success: true}
			oldTask := domain.TaskLinkFact{SubagentType: "tester", Description: "historical"}
			ledger := domain.AgenticFactLedger{SchemaVersion: &v, TicketID: &ticket, Skills: []domain.SkillFact{oldSkill, oldSkill}, TaskLinks: []domain.TaskLinkFact{oldTask, oldTask}, Sessions: []domain.SessionFact{
				{ID: "root-a", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stale-a"}}},
				{ID: "root-b", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "tester", Model: "stale-b"}}},
				{ID: "codex-untouched", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "reviewer", Model: "other-source"}}},
			}}
			if tc.missingChild {
				parent := "root-b"
				ledger.Sessions = append(ledger.Sessions, domain.SessionFact{ID: "child-b", ParentID: &parent, Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-child"}}})
			}
			if tc.emptyOwners {
				ledger.Sessions = ledger.Sessions[2:]
			}
			if tc.cyclic {
				a, b := "cycle-a", "cycle-b"
				ledger.Sessions = append(ledger.Sessions, domain.SessionFact{ID: a, ParentID: &b, Source: domain.SourceOpenCode}, domain.SessionFact{ID: b, ParentID: &a, Source: domain.SourceOpenCode})
			}
			workspace := createTestWorkspace(t, root, ticket, ledger)
			_, before := readRecoveryLedger(t, workspace)
			summaryPath := filepath.Join(workspace, "SUMMARY.md")
			summaryBefore, err := os.ReadFile(summaryPath)
			if err != nil {
				t.Fatal(err)
			}
			dbPath := filepath.Join(root, "source.db")
			roots := []string{"root-a", "root-b"}
			if tc.missingRoot {
				roots = roots[:1]
			}
			setupMockOpenCodeMultiRootDB(t, dbPath, roots, ticket)
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if !tc.emptyObservations {
				for i, kind := range []string{"skill", "skill", "task", "task"} {
					input := `{"name":"fresh"}`
					if kind == "task" {
						input = `{"subagent_type":"coder","description":"fresh"}`
					}
					data := fmt.Sprintf(`{"type":"tool","tool":%q,"state":{"status":"completed","input":%s,"output":"ok"}}`, kind, input)
					if _, err := db.Exec(`INSERT INTO part VALUES(?, 'root-a','m-1',?,1700000000000,1700000000001)`, fmt.Sprintf("fresh-%d", i), data); err != nil {
						t.Fatal(err)
					}
				}
			}
			if tc.malformed {
				if _, err := db.Exec(`INSERT INTO part VALUES('bad','root-b','m-2','{',1800000000000,1800000000001)`); err != nil {
					t.Fatal(err)
				}
			}
			if tc.blockWrite {
				if err := os.Mkdir(filepath.Join(workspace, "agentic_metrics.json.tmp"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			opts := BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "opencode", RecomputeDB: true, DBPath: dbPath, DryRun: tc.dryRun, UpdateSummaries: tc.dryRun}
			report, err := NewEngine().Execute(opts)
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Results) != 1 {
				t.Fatalf("missing ticket outcome: %+v", report)
			}
			res := report.Results[0]
			updated, after := readRecoveryLedger(t, workspace)
			if tc.blockWrite {
				if res.Error == nil || string(before) != string(after) {
					t.Fatalf("failed write published candidate or hid error: %+v", res)
				}
				return
			}
			if res.Error != nil {
				t.Fatalf("unexpected per-ticket failure: %v", res.Error)
			}
			incomplete := tc.missingRoot || tc.missingChild || tc.malformed || tc.emptyOwners || tc.cyclic
			if res.RefreshIncomplete != incomplete || res.CollectionsRetained != incomplete || report.CollectionsRetained != boolInt(incomplete) || report.RefreshIncomplete != boolInt(incomplete) {
				t.Fatalf("command and ticket recovery disagree: %+v / %+v", report, res)
			}
			if incomplete && (!strings.Contains(res.Notes, "previous snapshot") || len(res.Diagnostics) == 0 || len(report.Warnings) == 0) {
				t.Fatalf("retention/mixed age absent from ticket and command diagnostics: %+v", report)
			}
			if tc.dryRun {
				summaryAfter, err := os.ReadFile(summaryPath)
				if err != nil {
					t.Fatal(err)
				}
				if string(before) != string(after) || string(summaryBefore) != string(summaryAfter) {
					t.Fatal("dry run wrote ledger/summary")
				}
				return
			}
			if updated.Migration != nil {
				t.Fatal("optional refresh imposed mandatory migration")
			}
			if err := adapters.ValidateRuntimeFactLedger(after); err != nil {
				t.Fatal(err)
			}
			wantSkill, wantTask := "fresh", "fresh"
			wantRows := 2
			if incomplete {
				wantSkill, wantTask = "historical", "historical"
			}
			if tc.emptyObservations {
				wantRows = 0
			}
			if len(updated.Skills) != wantRows || len(updated.TaskLinks) != wantRows {
				t.Fatalf("collection multiplicity/replacement wrong: %+v %+v", updated.Skills, updated.TaskLinks)
			}
			for _, skill := range updated.Skills {
				if skill.Name != wantSkill {
					t.Fatalf("skill snapshot changed incorrectly: %+v", updated.Skills)
				}
			}
			for _, task := range updated.TaskLinks {
				if task.Description != wantTask {
					t.Fatalf("task snapshot changed incorrectly: %+v", updated.TaskLinks)
				}
			}
			models := map[string]string{}
			for _, session := range updated.Sessions {
				if len(session.Invocations) > 0 {
					models[session.ID] = session.Invocations[0].Model
				}
			}
			if models["codex-untouched"] != "other-source" {
				t.Fatal("other source changed")
			}
			if !tc.emptyOwners && models["root-a"] != "gpt-4o" {
				t.Fatalf("independent complete tree not refreshed: %v", models)
			}
			if tc.missingRoot || tc.missingChild || tc.malformed {
				if models["root-b"] != "stale-b" {
					t.Fatalf("incomplete tree partially replaced: %v", models)
				}
			}
			if tc.missingChild && models["child-b"] != "stored-child" {
				t.Fatalf("unavailable child lost: %v", models)
			}
			if !incomplete && !tc.emptyObservations {
				// A cumulative snapshot grows once, then an unchanged repeat preserves all rows.
				if _, err := db.Exec(`INSERT INTO part VALUES('growth','root-a','m-1','{"type":"tool","tool":"skill","state":{"status":"completed","input":{"name":"growth"},"output":"ok"}}',1800000000000,1800000000001)`); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < 2; i++ {
					again, err := NewEngine().Execute(opts)
					if err != nil || again.Results[0].Error != nil {
						t.Fatalf("repeat failed: %+v %v", again, err)
					}
					snapshot, _ := readRecoveryLedger(t, workspace)
					if len(snapshot.Skills) != 3 || len(snapshot.TaskLinks) != 2 {
						t.Fatalf("cumulative refresh duplicated/dropped observations: %+v %+v", snapshot.Skills, snapshot.TaskLinks)
					}
				}
			}
		})
	}
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func TestBackfillEngine_GlobalRecoveryRequiresEveryKnownTree(t *testing.T) {
	for _, tc := range []struct {
		name, source                                        string
		missingChild, lateParse, groupFailure, parseWarning bool
	}{
		{name: "unavailable Codex sibling", source: "both"}, {name: "unselected OpenCode obligation", source: "codex"},
		{name: "missing stored child", source: "both", missingChild: true}, {name: "late parse failure", source: "both", lateParse: true},
		{name: "group failure", source: "both", groupFailure: true},
		{name: "successful parse with warnings", source: "both", parseWarning: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			ticket := "CUSTOM-GLOBAL"
			v := domain.CurrentSchemaVersion
			parent := "codex-a"
			ledger := domain.AgenticFactLedger{SchemaVersion: &v, TicketID: &ticket, Migration: &domain.MigrationInfo{SchemaVersion: v, Status: domain.MigrationStatusPartial, StepsCompleted: []string{domain.MigrationStepGitCodeFacts}, StepsPending: []string{domain.MigrationStepSourceBackfill, "other_requirement"}}, Skills: []domain.SkillFact{{Name: "keep"}}, Sessions: []domain.SessionFact{
				{ID: "codex-a", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "coder", Model: "old-a"}}},
				{ID: "codex-b", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "coder", Model: "old-b"}}},
				{ID: "opencode-c", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "tester", Model: "old-c"}}},
			}}
			if tc.missingChild {
				ledger.Sessions = append(ledger.Sessions, domain.SessionFact{ID: "missing-child", ParentID: &parent, Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "coder", Model: "old-child"}}})
			}
			workspace := createTestWorkspace(t, root, ticket, ledger)
			engine := NewEngine()
			engine.discoverCodexRollouts = func(opts codex.DiscoveryOptions) ([]string, error) {
				if opts.SessionID == "codex-b" {
					return nil, errors.New("source not on this machine")
				}
				if tc.lateParse {
					return []string{"good", "bad"}, nil
				}
				return []string{"good"}, nil
			}
			engine.parseCodexRollout = func(path string) (codex.Transcript, error) {
				if path == "bad" {
					return codex.Transcript{}, errors.New("malformed late source")
				}
				if tc.parseWarning {
					return codex.Transcript{ThreadID: "codex-a", Warnings: []string{"line 3: malformed JSON record"}}, nil
				}
				return codex.Transcript{ThreadID: "codex-a", Invocations: []domain.InvocationFact{{Role: "coder", Model: "fresh-a"}}}, nil
			}
			if tc.groupFailure {
				engine.combineCodex = func([]codex.Transcript) (codex.Transcript, error) {
					return codex.Transcript{}, errors.New("ambiguous grouping")
				}
			}
			report, err := engine.Execute(BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: tc.source, CodexHome: root, DBPath: filepath.Join(root, "missing.db"), RecomputeDB: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Results) != 1 || report.Results[0].Error != nil {
				t.Fatalf("refresh failed: %+v", report)
			}
			res := report.Results[0]
			if res.DBFound || !res.RefreshIncomplete || !containsMigrationStep(res.PendingSteps, domain.MigrationStepSourceBackfill) || !containsMigrationStep(res.PendingSteps, "other_requirement") {
				t.Fatalf("partial source wrongly discharged global/unrelated requirement: %+v", res)
			}
			updated, data := readRecoveryLedger(t, workspace)
			if err := adapters.ValidateRuntimeFactLedger(data); err == nil {
				t.Fatal("pending global requirement became runtime eligible")
			}
			models := map[string]string{}
			for _, s := range updated.Sessions {
				models[s.ID] = s.Invocations[0].Model
			}
			wantA := "fresh-a"
			if tc.missingChild || tc.lateParse || tc.groupFailure || tc.parseWarning {
				wantA = "old-a"
			}
			if models["codex-a"] != wantA || models["codex-b"] != "old-b" || models["opencode-c"] != "old-c" {
				t.Fatalf("owner tree preservation/update wrong: %v", models)
			}
			if tc.missingChild && models["missing-child"] != "old-child" {
				t.Fatal("missing stored child lost")
			}
			if len(updated.Skills) != 1 || updated.Skills[0].Name != "keep" {
				t.Fatal("Codex success erased ownerless OpenCode snapshot")
			}
			if len(res.Diagnostics) == 0 || len(report.Warnings) == 0 {
				t.Fatal("partial failure not exposed in command and ticket diagnostics")
			}
		})
	}
}

func TestBackfillEngine_SchemaV2RecoversKnownGrandchildren(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		t.Run(fmt.Sprintf("malformed=%t", malformed), func(t *testing.T) {
			root := t.TempDir()
			ticket := "CUSTOM-V2"
			v := domain.CurrentSchemaVersion
			parent, grandparent := "v2-root", "v2-child"
			ledger := domain.AgenticFactLedger{SchemaVersion: &v, TicketID: &ticket, Skills: []domain.SkillFact{{Name: "stored"}}, Sessions: []domain.SessionFact{
				{ID: "v2-root", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "old"}}},
				{ID: "v2-child", ParentID: &parent, Source: domain.SourceOpenCode},
				{ID: "v2-grandchild", ParentID: &grandparent, Source: domain.SourceOpenCode},
			}}
			workspace := createTestWorkspace(t, root, ticket, ledger)
			dbPath := filepath.Join(root, "v2.db")
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if _, err := db.Exec(`CREATE TABLE session_v2(id TEXT PRIMARY KEY,parent_id TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER); CREATE TABLE session_message(id TEXT PRIMARY KEY,session_id TEXT,type TEXT,seq INTEGER,time_created INTEGER,time_updated INTEGER,data TEXT);
  INSERT INTO session_v2 VALUES('v2-root',NULL,'coder','model','root',1700000000000,1700000001000),('v2-child','v2-root','coder','model','child',1700000000000,1700000001000),('v2-grandchild','v2-child','tester','model','grandchild',1700000000000,1700000001000);`); err != nil {
				t.Fatal(err)
			}
			payload := `{"model":{"id":"fresh-v2"},"content":[],"tokens":{"input":17,"output":2},"time":{"created":1700000000000,"completed":1700000001000}}`
			for _, id := range []string{"v2-root", "v2-child", "v2-grandchild"} {
				data := payload
				if malformed && id == "v2-grandchild" {
					data = "{"
				}
				if _, err := db.Exec(`INSERT INTO session_message VALUES(?,?,'assistant',1,1700000000000,1700000001000,?)`, id+"-message", id, data); err != nil {
					t.Fatal(err)
				}
			}
			report, err := NewEngine().Execute(BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "opencode", RecomputeDB: true, DBPath: dbPath})
			if err != nil || report.Results[0].Error != nil {
				t.Fatalf("refresh failed: %+v %v", report, err)
			}
			updated, _ := readRecoveryLedger(t, workspace)
			if len(updated.Sessions) != 3 {
				t.Fatalf("grandchild lost/duplicated: %+v", updated.Sessions)
			}
			if malformed {
				rootPreserved := false
				for _, session := range updated.Sessions {
					if session.ID == "v2-root" {
						rootPreserved = len(session.Invocations) == 1 && session.Invocations[0].Model == "old"
					}
				}
				if !report.Results[0].RefreshIncomplete || !report.Results[0].CollectionsRetained || !rootPreserved || len(updated.Skills) != 1 {
					t.Fatalf("late malformed descendant partially replaced tree: %+v %+v", report.Results, updated)
				}
			} else {
				if report.Results[0].RefreshIncomplete || len(updated.Skills) != 0 {
					t.Fatalf("complete empty recovery did not clear stale producer snapshot: %+v %+v", report.Results, updated.Skills)
				}
				for _, session := range updated.Sessions {
					if len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != 17 {
						t.Fatalf("nested cumulative source snapshot not recovered: %+v", session)
					}
				}
			}
		})
	}
}

func TestBackfillEngine_InvalidCandidatePreservesFiles(t *testing.T) {
	for _, dryRun := range []bool{false, true} {
		t.Run(fmt.Sprintf("dry-run=%t", dryRun), func(t *testing.T) {
			root := t.TempDir()
			ticket := "CUSTOM-INVALID-CANDIDATE"
			v := domain.CurrentSchemaVersion
			ledger := domain.AgenticFactLedger{SchemaVersion: &v, TicketID: &ticket, Sessions: []domain.SessionFact{{ID: "root", Source: domain.SourceCodex, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored"}}}}}
			workspace := createTestWorkspace(t, root, ticket, ledger)
			_, before := readRecoveryLedger(t, workspace)
			summary := filepath.Join(workspace, "SUMMARY.md")
			summaryBefore, err := os.ReadFile(summary)
			if err != nil {
				t.Fatal(err)
			}
			engine := NewEngine()
			engine.discoverCodexRollouts = func(codex.DiscoveryOptions) ([]string, error) { return []string{"source"}, nil }
			engine.parseCodexRollout = func(string) (codex.Transcript, error) {
				return codex.Transcript{ThreadID: "root", Invocations: []domain.InvocationFact{{Role: "coder", Model: "fresh"}}, Tools: []domain.ToolFact{{SessionID: "root", Name: "read", Calls: -1}}}, nil
			}
			report, err := engine.Execute(BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "codex", CodexHome: root, RecomputeDB: true, DryRun: dryRun, UpdateSummaries: true})
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Results) != 1 || report.Results[0].Error == nil || !strings.Contains(report.Results[0].Error.Error(), "validating updated ledger") {
				t.Fatalf("invalid candidate did not reach final validation guard: %+v", report.Results)
			}
			_, after := readRecoveryLedger(t, workspace)
			summaryAfter, err := os.ReadFile(summary)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) || string(summaryBefore) != string(summaryAfter) {
				t.Fatal("final validation failure published candidate/summary")
			}
			if _, err := os.Stat(filepath.Join(workspace, "agentic_metrics.json.tmp")); !os.IsNotExist(err) {
				t.Fatalf("validation failure reached write boundary: %v", err)
			}
		})
	}
}

func TestRecoverOpenCodeRetainsUnavailableOrphanSiblingTree(t *testing.T) {
	for _, intermediary := range []bool{false, true} {
		t.Run(fmt.Sprintf("intermediary=%t", intermediary), func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.db")
			setupMockOpenCodeMultiRootDB(t, path, []string{"source-root", "independent"}, "CUSTOM-ORPHAN")
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			parent := "source-root"
			if intermediary {
				parent = "intermediary"
				if _, err := db.Exec(`INSERT INTO session VALUES('intermediary','source-root','coder','fresh-parent','parent',1700000000000,1700000001000)`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := db.Exec(`INSERT INTO session VALUES('child-a',?,'coder','fresh-child','child',1700000000000,1700000001000)`, parent); err != nil {
				t.Fatal(err)
			}
			stored := []domain.SessionFact{
				{ID: "child-a", ParentID: &parent, Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-a"}}},
				{ID: "child-b", ParentID: &parent, Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-b"}}},
				{ID: "independent", Source: domain.SourceOpenCode, Invocations: []domain.InvocationFact{{Role: "coder", Model: "stored-independent"}}},
			}
			// Exercise the source-recovery boundary directly: runtime file validation
			// independently rejects absent parents, but recovery must also preserve branches.
			recovered := recoverOpenCode(db, stored, "tester", "CUSTOM-ORPHAN", root, false)
			if recovered.complete || !recovered.updated || len(recovered.diagnostics) == 0 {
				t.Fatalf("missing sibling falsely certified or independent refresh lost: %+v", recovered)
			}
			models := map[string]string{}
			for _, s := range recovered.sessions {
				if len(s.Invocations) > 0 {
					models[s.ID] = s.Invocations[0].Model
				}
			}
			if len(recovered.sessions) != 3 || models["child-a"] != "stored-a" || models["child-b"] != "stored-b" || models["independent"] != "gpt-4o" {
				t.Fatalf("partially discoverable orphan tree lost history: %+v", recovered.sessions)
			}
		})
	}
}

func TestBackfillEngine_WarningBearingRolloutsPreserveTree(t *testing.T) {
	for _, tc := range []struct{ name, damaged string }{
		{"malformed middle", "{broken}\n{\"type\":\"turn_context\"}\n"},
		{"malformed final record", "{broken}\n"},
		{"partial final write", "{\"type\":"},
		{"invalid usage", "{\"type\":\"token_usage_record\",\"payload\":{\"usage\":\"invalid\"}}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			v, ticket, parent := domain.CurrentSchemaVersion, "CUSTOM-WARNINGS", "damaged"
			old := func(id string, parentID *string, source domain.SourceHarness) domain.SessionFact {
				return domain.SessionFact{ID: id, ParentID: parentID, Source: source, User: "stored-user", Invocations: []domain.InvocationFact{
					{Role: "coder", Model: "stored", Tokens: domain.InvocationTokens{Input: 1000}},
					{Role: "coder", Model: "stored", Tokens: domain.InvocationTokens{Input: 1000}},
				}}
			}
			ledger := domain.AgenticFactLedger{SchemaVersion: &v, TicketID: &ticket, Sessions: []domain.SessionFact{
				old("damaged", nil, domain.SourceCodex), old("child", &parent, domain.SourceCodex),
				old("healthy", nil, domain.SourceCodex), old("other-harness", nil, domain.SourceOpenCode),
			}}
			workspace := createTestWorkspace(t, root, ticket, ledger)
			baseline, _ := readRecoveryLedger(t, workspace)
			baselineSessions := make(map[string]adapters.SessionFactDTO)
			for _, session := range baseline.Sessions {
				baselineSessions[session.ID] = session
			}
			writeSource := func(id, extra, suffix string) string {
				t.Helper()
				path := filepath.Join(root, id+".jsonl")
				content := fmt.Sprintf("{\"timestamp\":\"2026-01-01T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":%q%s}}\n{\"timestamp\":\"2026-01-01T00:00:01Z\",\"type\":\"token_usage_record\",\"payload\":{\"model\":\"fresh\",\"usage\":{\"input_tokens\":25}}}\n%s", id, extra, suffix)
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				return path
			}
			rootPath := writeSource("damaged", "", "")
			childPath := writeSource("child", `,"source":{"subagent":{"thread_spawn":{"parent_thread_id":"damaged"}}}`, tc.damaged)
			healthyPath := writeSource("healthy", "", "")
			engine := NewEngine()
			engine.discoverCodexRollouts = func(opts codex.DiscoveryOptions) ([]string, error) {
				if opts.SessionID == "damaged" {
					return []string{rootPath, childPath}, nil
				}
				return []string{healthyPath}, nil
			}
			opts := BackfillOptions{WorkspacesDir: root, TicketID: ticket, Source: "codex", CodexHome: root, RecomputeDB: true}
			report, err := engine.Execute(opts)
			if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil {
				t.Fatalf("refresh: %+v %v", report, err)
			}
			if !report.Results[0].RefreshIncomplete || len(report.Results[0].Diagnostics) == 0 || len(report.Warnings) == 0 {
				t.Fatalf("missing partial diagnostics: %+v", report)
			}
			updated, _ := readRecoveryLedger(t, workspace)
			seen := make(map[string]bool)
			for _, s := range updated.Sessions {
				prior, exists := baselineSessions[s.ID]
				if !exists || seen[s.ID] {
					t.Fatalf("unexpected or duplicate session: %s", s.ID)
				}
				seen[s.ID] = true
				if s.ID != "healthy" && !reflect.DeepEqual(s, prior) {
					t.Fatalf("retained session changed: got %+v, want %+v", s, prior)
				}
				if s.ID == "healthy" {
					if len(s.Invocations) != 1 || s.Invocations[0].Tokens.Input != 25 {
						t.Fatalf("healthy sibling not refreshed: %+v", s)
					}
					continue
				}
				if s.User != "stored-user" || len(s.Invocations) != 2 || s.Invocations[0].Tokens.Input != 1000 || s.Invocations[1].Tokens.Input != 1000 || s.Invocations[0].Model != "stored" {
					t.Fatalf("stored facts changed: %+v", s)
				}
				if s.ID == "child" && (s.ParentID == nil || *s.ParentID != parent) {
					t.Fatalf("parent lost: %+v", s)
				}
				if s.ID == "other-harness" && s.Source != string(domain.SourceOpenCode) {
					t.Fatal("source changed")
				}
			}
			if len(updated.Sessions) != 4 {
				t.Fatalf("sessions lost: %+v", updated.Sessions)
			}
			// Completing/repairing the source allows a later retry; a valid final
			// record does not need a newline to be recoverable.
			writeSource("child", `,"source":{"subagent":{"thread_spawn":{"parent_thread_id":"damaged"}}}`, `{"type":"turn_context"}`)
			var recovered []adapters.SessionFactDTO
			for attempt := 0; attempt < 2; attempt++ {
				report, err = engine.Execute(opts)
				if err != nil || report.Results[0].Error != nil || report.Results[0].RefreshIncomplete {
					t.Fatalf("retry failed: %+v %v", report, err)
				}
				updated, _ = readRecoveryLedger(t, workspace)
				if len(updated.Sessions) != len(baselineSessions) {
					t.Fatalf("retry lost sessions: %+v", updated.Sessions)
				}
				if attempt == 1 && !reflect.DeepEqual(updated.Sessions, recovered) {
					t.Fatal("unchanged retry changed recovered facts")
				}
				recovered = updated.Sessions
				for _, s := range updated.Sessions {
					if s.Source == string(domain.SourceCodex) && (len(s.Invocations) != 1 || s.Invocations[0].Tokens.Input != 25) {
						t.Fatalf("retry/idempotence failure: %+v", s)
					}
				}
			}
		})
	}
}
