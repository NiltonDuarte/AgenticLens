package backfill

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
	"modernc.org/sqlite"
)

type discoveryReadContext struct {
	context.Context
	afterRead func()
}

func (ctx discoveryReadContext) Err() error {
	ctx.afterRead()
	return ctx.Context.Err()
}

func TestBackfillCancellationKeepsInterruptedWorkspaceBytesAndEarlierCommit(t *testing.T) {
	project := newBackfillProject(t)
	home := filepath.Join(project, "sessions")
	for _, id := range []string{"prior", "first-owner", "interrupted-owner", "later"} {
		writeBackfillRollout(t, home, id, "", "coder", "model", map[string]int{"fresh-tool": 2}, "")
	}
	newOwner := func(id string, source domain.SourceHarness) domain.SessionFact {
		return domain.SessionFact{
			ID: id, Source: source, User: "fixture-user",
			Tools: []domain.SessionToolFact{{Name: "accepted-tool", Calls: 7}},
		}
	}
	create := func(ticket string, owners ...domain.SessionFact) string {
		schema := domain.CurrentSchemaVersion
		return createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
			SchemaVersion: &schema, TicketID: &ticket, Sessions: owners,
		})
	}
	prior := create("AL-10", newOwner("prior", domain.SourceCodex))
	current := create("AL-20", newOwner("first-owner", domain.SourceCodex), newOwner("interrupted-owner", domain.SourceCodex), newOwner("portable-open-code", domain.SourceOpenCode))
	later := create("AL-30", newOwner("later", domain.SourceCodex))
	_, priorSummary := readBackfillArtifacts(t, prior)
	currentLedger, currentSummary := readBackfillArtifacts(t, current)
	laterLedger, laterSummary := readBackfillArtifacts(t, later)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine := NewEngine()
	var order []string
	var interruptedBytes int64
	engine.discoverCodexRollouts = func(options codex.DiscoveryOptions) ([]string, error) {
		order = append(order, options.SessionID)
		if options.SessionID != "interrupted-owner" {
			return codex.DiscoverRolloutPaths(options)
		}
		if options.Budget == nil || options.Budget.Context == nil {
			return nil, errors.New("source discovery lacks the workspace context")
		}
		// Decorate the real discovery operation to synchronize on bytes
		// already read from a native rollout in the second owner's scan.
		budget := options.Budget
		before := budget.SourceBytesRead
		budget.Context = discoveryReadContext{Context: budget.Context, afterRead: func() {
			if budget.SourceBytesRead > before {
				interruptedBytes = budget.SourceBytesRead - before
				cancel()
			}
		}}
		paths, err := codex.DiscoverRolloutPaths(options)
		if err == nil && interruptedBytes > 0 {
			return nil, fmt.Errorf("native discovery accepted %d paths after cancellation", len(paths))
		}
		return paths, err
	}
	opts := BackfillOptions{
		Project: project, Source: "codex", CodexHome: project, RecomputeDB: true,
		UpdateSummaries: true, SnapshotCutoff: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	report, err := engine.ExecuteContext(ctx, opts)
	if !errors.Is(err, context.Canceled) || report == nil || len(report.Results) != 2 || interruptedBytes == 0 ||
		!slices.Equal(order, []string{"prior", "first-owner", "interrupted-owner"}) {
		t.Fatalf("expected cancellation after the first owner of the second workspace: report=%+v order=%v bytes=%d error=%v", report, order, interruptedBytes, err)
	}
	currentResult := report.Results[1]
	if currentResult.TicketID != "AL-20" || !errors.Is(currentResult.Error, context.Canceled) || !currentResult.RefreshIncomplete {
		t.Fatalf("interrupted workspace was not reported as incomplete: %+v", currentResult)
	}
	afterCurrentLedger, afterCurrentSummary := readBackfillArtifacts(t, current)
	afterLaterLedger, afterLaterSummary := readBackfillArtifacts(t, later)
	if !bytes.Equal(currentLedger, afterCurrentLedger) || !bytes.Equal(currentSummary, afterCurrentSummary) ||
		!bytes.Equal(laterLedger, afterLaterLedger) || !bytes.Equal(laterSummary, afterLaterSummary) {
		t.Fatal("cancellation changed the interrupted or unstarted workspace's persisted artifacts")
	}
	accepted, err := adapters.LoadExistingFactLedger(filepath.Join(prior, "agentic_metrics.json"))
	_, afterPriorSummary := readBackfillArtifacts(t, prior)
	if err != nil || len(accepted.Sessions) != 1 || len(accepted.Sessions[0].Tools) != 1 ||
		accepted.Sessions[0].Tools[0].Name != "fresh-tool" || accepted.Sessions[0].Tools[0].Calls != 2 ||
		bytes.Equal(priorSummary, afterPriorSummary) {
		t.Fatalf("earlier workspace lost its committed artifacts: %+v, %v", accepted, err)
	}
}

func TestBackfillCancellationInSecondHarnessReportsUncommittedFirstRead(t *testing.T) {
	options, workspace := combinedSourceReadFixture(t)
	ledgerPath, summaryPath := filepath.Join(workspace, "agentic_metrics.json"), filepath.Join(workspace, "SUMMARY.md")
	ledger, summary := readBackfillArtifacts(t, workspace)
	ledgerStamp, summaryStamp := markRefreshFile(t, ledgerPath), markRefreshFile(t, summaryPath)
	options.DryRun = true
	control, err := NewEngine().Execute(options)
	if err != nil || control == nil || len(control.Results) != 1 || len(control.Results[0].Coverage) != 2 || control.Results[0].Coverage[0].Proposed != 1 {
		t.Fatalf("control must verify a changed OpenCode owner: %+v, %v", control, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine := NewEngine()
	var codexBytes int64
	cancelOnCodexDiscoveryRead(engine, cancel, &codexBytes)
	options.DryRun = false
	report, err := engine.ExecuteContext(ctx, options)
	if !errors.Is(err, context.Canceled) || codexBytes == 0 || report == nil || len(report.Results) != 1 {
		t.Fatalf("expected cancellation during the actual Codex read after OpenCode: %d bytes, report %+v, error %v", codexBytes, report, err)
	}
	result := report.Results[0]
	if !errors.Is(result.Error, context.Canceled) || result.Publication != PublicationFailed || report.RefreshIncomplete != 1 || len(result.Coverage) != 2 {
		t.Fatalf("workspace did not report interrupted publication: %+v", report)
	}
	openCode, codexCoverage := result.Coverage[0], result.Coverage[1]
	if openCode.Harness != domain.SourceOpenCode || openCode.Refreshed != 0 || openCode.Proposed != 1 || openCode.Retained != 0 ||
		len(openCode.Owners) != 1 || openCode.Owners[0].SessionID != "local-open" || openCode.Owners[0].Status != OwnerProposed {
		t.Errorf("interruption hid the verified but uncommitted OpenCode change: %+v", openCode)
	}
	if codexCoverage.Harness != domain.SourceCodex || codexCoverage.Retained != 1 || len(codexCoverage.Owners) != 1 || codexCoverage.Owners[0].SessionID != "local-codex" ||
		codexCoverage.Owners[0].Status != OwnerRetained || !strings.Contains(codexCoverage.Owners[0].Reason, context.Canceled.Error()) {
		t.Errorf("interrupted Codex owner did not preserve its accepted facts and reason: %+v", codexCoverage)
	}
	assertExplicitWrites(t, report, false, false)
	assertRefreshFile(t, ledgerPath, ledger, ledgerStamp)
	assertRefreshFile(t, summaryPath, summary, summaryStamp)
}

func TestBackfillReportsVerifiedReadsWhenAnomalyPolicyPreventsComparison(t *testing.T) {
	for _, tc := range []struct {
		name        string
		interrupted bool
	}{
		{name: "both sources read"},
		{name: "interrupted during Codex discovery", interrupted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			options, workspace := combinedSourceReadFixture(t)
			setRefreshFixturePolicy(t, options.Project, "  error_rate_min_calls: 7\n", "")
			ledgerPath, summaryPath := filepath.Join(workspace, "agentic_metrics.json"), filepath.Join(workspace, "SUMMARY.md")
			ledger, summary := readBackfillArtifacts(t, workspace)
			ledgerStamp, summaryStamp := markRefreshFile(t, ledgerPath), markRefreshFile(t, summaryPath)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			engine := NewEngine()
			var codexBytes int64
			if tc.interrupted {
				cancelOnCodexDiscoveryRead(engine, cancel, &codexBytes)
			}
			report, err := engine.ExecuteContext(ctx, options)
			if tc.interrupted {
				if !errors.Is(err, context.Canceled) || codexBytes == 0 {
					t.Fatalf("expected native Codex cancellation after bytes read: bytes=%d error=%v", codexBytes, err)
				}
			} else if err != nil {
				t.Fatalf("late policy rejection should be a per-ticket failure: %v", err)
			}
			if report == nil || len(report.Results) != 1 || len(report.Coverage) != 2 {
				t.Fatalf("late policy rejection must retain the two selected harness reports: %+v", report)
			}
			result := report.Results[0]
			if result.Error == nil || !strings.Contains(result.Error.Error(), "error_rate_min_calls") || result.Publication != PublicationFailed || len(result.Coverage) != 2 {
				t.Fatalf("expected anomaly policy rejection after source reads: %+v", result)
			}
			if tc.interrupted && !errors.Is(result.Error, context.Canceled) {
				t.Errorf("policy failure obscured the workspace cancellation: %v", result.Error)
			}
			assertVerifiedUncomparedOwner(t, result.Coverage[0], domain.SourceOpenCode, "local-open")
			if report.Coverage[0].Harness != domain.SourceOpenCode || report.Coverage[0].RefreshCounts != result.Coverage[0].RefreshCounts {
				t.Errorf("batch counts lost OpenCode verification: %+v", report.Coverage)
			}
			if tc.interrupted {
				codexCoverage := result.Coverage[1]
				if codexCoverage.Harness != domain.SourceCodex || codexCoverage.RefreshCounts != (RefreshCounts{Recorded: 1, Retained: 1}) ||
					len(codexCoverage.Owners) != 1 || codexCoverage.Owners[0].SessionID != "local-codex" || codexCoverage.Owners[0].Status != OwnerRetained ||
					!strings.Contains(codexCoverage.Owners[0].Reason, context.Canceled.Error()) {
					t.Errorf("interrupted Codex owner must still report retained facts: %+v", codexCoverage)
				}
			} else {
				assertVerifiedUncomparedOwner(t, result.Coverage[1], domain.SourceCodex, "local-codex")
			}
			assertExplicitWrites(t, report, false, false)
			assertRefreshFile(t, ledgerPath, ledger, ledgerStamp)
			assertRefreshFile(t, summaryPath, summary, summaryStamp)
		})
	}
}

func assertVerifiedUncomparedOwner(t *testing.T, coverage HarnessRefreshCoverage, harness domain.SourceHarness, id string) {
	t.Helper()
	if coverage.Harness != harness || coverage.RefreshCounts != (RefreshCounts{Recorded: 1, VerifiedUncompared: 1}) || len(coverage.Owners) != 1 {
		t.Fatalf("expected one verified owner with an unavailable comparison in %s: %+v", harness, coverage)
	}
	owner := coverage.Owners[0]
	if owner.SessionID != id || owner.Status != OwnerVerifiedUncompared || owner.Skills != CollectionEmpty || owner.Tasks != CollectionEmpty {
		t.Errorf("expected the native read to retain its verified empty collections: %+v", owner)
	}
}

func combinedSourceReadFixture(t *testing.T) (BackfillOptions, string) {
	t.Helper()
	project := newBackfillProject(t)
	ticket, version := "TEST-CANCEL-SECOND-HARNESS", domain.CurrentSchemaVersion
	workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket,
		Sessions: []domain.SessionFact{
			{ID: "local-open", Source: domain.SourceOpenCode, Tools: []domain.SessionToolFact{{Name: "historical", Calls: 1}}},
			{ID: "local-codex", Source: domain.SourceCodex},
		},
	})
	dbPath := filepath.Join(project, "source.db")
	setupMockOpenCodeDB(t, dbPath, "local-open", ticket)
	writeBackfillRollout(t, filepath.Join(project, "sessions"), "local-codex", "", "coder", "native-model", map[string]int{"fresh": 2}, "")
	return BackfillOptions{Project: project, TicketID: ticket, Source: "both", DBPath: dbPath, CodexHome: project,
		RecomputeDB: true, UpdateSummaries: true, SnapshotCutoff: time.Now().Add(time.Second)}, workspace
}

func cancelOnCodexDiscoveryRead(engine *BackfillEngine, cancel context.CancelFunc, bytesRead *int64) {
	engine.discoverCodexRollouts = func(opts codex.DiscoveryOptions) ([]string, error) {
		budget := opts.Budget
		if budget == nil || budget.Context == nil {
			return nil, errors.New("native discovery lacks its source budget or context")
		}
		before := budget.SourceBytesRead
		budget.Context = discoveryReadContext{Context: budget.Context, afterRead: func() {
			*bytesRead = budget.SourceBytesRead - before
			if *bytesRead > 0 {
				cancel()
			}
		}}
		return codex.DiscoverRolloutPaths(opts)
	}
}

func readBackfillArtifacts(t *testing.T, dir string) (ledger, summary []byte) {
	t.Helper()
	ledger, err := os.ReadFile(filepath.Join(dir, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	summary, err = os.ReadFile(filepath.Join(dir, "SUMMARY.md"))
	if err != nil {
		t.Fatal(err)
	}
	return ledger, summary
}

var backfillCancelCollationID atomic.Uint64

func TestBackfillOpenCodeInheritsCallerDeadlineDuringNativeRead(t *testing.T) {
	const fixtureMessages = 4096
	project := newBackfillProject(t)
	path := filepath.Join(project, "opencode.sqlite")
	ticket, version := "AL-66", domain.CurrentSchemaVersion
	workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket,
		Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceOpenCode, User: "tester"}},
	})
	setupMockOpenCodeDB(t, path, "owner", ticket)
	beforeLedger, beforeSummary := readBackfillArtifacts(t, workspace)

	started, proceed := make(chan struct{}), make(chan struct{})
	var release sync.Once
	defer release.Do(func() { close(proceed) })
	var armed, nativeFinished atomic.Bool
	var comparisons atomic.Uint64
	collation := fmt.Sprintf("fixture_backfill_gate_%d", backfillCancelCollationID.Add(1))
	if err := sqlite.RegisterCollationUtf8(collation, func(left, right string) int {
		if armed.Load() {
			switch comparisons.Add(1) {
			case 1:
				close(started)
				<-proceed
			case fixtureMessages:
				nativeFinished.Store(true)
			}
			runtime.Gosched()
		}
		return strings.Compare(left, right)
	}); err != nil {
		t.Fatal(err)
	}
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, setupErr := writer.Exec(fmt.Sprintf(`
		ALTER TABLE message RENAME TO original_message;
		CREATE TABLE message(id TEXT PRIMARY KEY, session_id TEXT COLLATE %[1]s, data TEXT, time_created INTEGER, time_updated INTEGER);
		INSERT INTO message SELECT * FROM original_message;
		DROP TABLE original_message;
		WITH RECURSIVE fixture(n) AS (VALUES(2) UNION ALL SELECT n+1 FROM fixture WHERE n < %[2]d)
		INSERT INTO message SELECT 'unused-'||n, 'different-owner', '{}', n, n FROM fixture;
	`, collation, fixtureMessages))
	if err := errors.Join(setupErr, writer.Close()); err != nil {
		t.Fatal(err)
	}

	guard, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	caller, cancel := context.WithTimeout(guard, 3*time.Second)
	defer cancel()
	opts := BackfillOptions{Project: project, TicketID: ticket, Source: "opencode", DBPath: path, RecomputeDB: true, UpdateSummaries: true}
	type outcome struct {
		report *BackfillReport
		err    error
	}
	done := make(chan outcome, 1)
	armed.Store(true)
	go func() {
		report, err := NewEngine().ExecuteContext(caller, opts)
		done <- outcome{report, err}
	}()
	select {
	case <-started:
	case got := <-done:
		t.Fatalf("backfill returned before its native OpenCode read: %+v", got)
	case <-guard.Done():
		t.Fatal("OpenCode did not enter the native row comparison")
	}
	if err := caller.Err(); err != nil {
		t.Fatalf("the real caller deadline expired before reaching SQLite: %v", err)
	}
	<-caller.Done()
	release.Do(func() { close(proceed) })
	select {
	case got := <-done:
		if nativeFinished.Load() {
			t.Fatal("native OpenCode read ran to completion after the caller deadline")
		}
		if !errors.Is(got.err, context.DeadlineExceeded) || got.report == nil || len(got.report.Results) != 1 {
			t.Fatalf("backfill lost the caller deadline and workspace report: %+v", got)
		}
		result := got.report.Results[0]
		if !errors.Is(result.Error, context.DeadlineExceeded) || !result.RefreshIncomplete || result.DBUpdated {
			t.Fatalf("interrupted workspace reported successful refresh: %+v", result)
		}
	case <-guard.Done():
		t.Fatal("OpenCode backfill held the native read after the caller deadline")
	}
	afterLedger, afterSummary := readBackfillArtifacts(t, workspace)
	if !bytes.Equal(beforeLedger, afterLedger) || !bytes.Equal(beforeSummary, afterSummary) {
		t.Fatal("OpenCode deadline changed the ledger or its summary")
	}
	armed.Store(false)
	retry, err := NewEngine().ExecuteContext(guard, opts)
	if err != nil || retry == nil || len(retry.Results) != 1 || !retry.Results[0].DBUpdated {
		t.Fatalf("fresh OpenCode backfill could not publish: %+v, %v", retry, err)
	}
}

type cancelAfterRecoveryContext struct {
	context.Context
	sourceSelected *bool
	cancel         context.CancelFunc
}

func (ctx cancelAfterRecoveryContext) Err() error {
	err := ctx.Context.Err()
	if pc, _, _, ok := runtime.Caller(1); ok && strings.HasSuffix(runtime.FuncForPC(pc).Name(), "/backfill.(*BackfillEngine).processSingleWorkspace") && *ctx.sourceSelected && err == nil {
		// Observe the post-recovery check as still active, then cancel
		// before serialization reaches the independent publication guard.
		ctx.cancel()
	}
	return err
}

func TestBackfillCancellationAfterSourceSuccessPreservesPublicationBytes(t *testing.T) {
	project := newBackfillProject(t)
	ticket, version := "AL-67", domain.CurrentSchemaVersion
	workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket,
		Sessions: []domain.SessionFact{{ID: "owner", Source: domain.SourceCodex, User: "tester"}},
	})
	writeBackfillRollout(t, filepath.Join(project, "sessions"), "owner", "", "coder", "model", map[string]int{"fresh-tool": 2}, "")
	beforeLedger, beforeSummary := readBackfillArtifacts(t, workspace)
	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	opts := BackfillOptions{Project: project, TicketID: ticket, Source: "codex", CodexHome: project, RecomputeDB: true, UpdateSummaries: true, SnapshotCutoff: cutoff}
	preview := opts
	preview.DryRun = true
	control, err := NewEngine().Execute(preview)
	if err != nil || control == nil || len(control.Results) != 1 || !control.Results[0].DBFound || !control.Results[0].DBUpdated {
		t.Fatalf("fixture did not produce a new candidate from native Codex: %+v, %v", control, err)
	}

	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		t.Fatal(err)
	}
	locations, _, err := discoverWorkspaceLedgers(cfg, ticket, workspace)
	if err != nil || len(locations) != 1 {
		t.Fatalf("fixture target selection failed: locations=%d error=%v", len(locations), err)
	}
	defer locations[0].Close()
	engine := NewEngine()
	engine.snapshotCutoff = cutoff
	engine.budget = &sourceReadBudget{sessions: cfg.Telemetry.Capture.MaxSessions, bytes: cfg.Telemetry.Capture.MaxSourceBytes}
	engine.generationPolicy = &cfg.Generation.Metrics
	var sourceSelected bool
	engine.discoverCodexRollouts = func(options codex.DiscoveryOptions) ([]string, error) {
		paths, err := codex.DiscoverRolloutPaths(options)
		sourceSelected = err == nil && len(paths) == 1
		return paths, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	late := cancelAfterRecoveryContext{Context: ctx, sourceSelected: &sourceSelected, cancel: cancel}
	result := engine.processSingleWorkspace(late, locations[0], opts, nil, "tester", nil, cutoff.Format(time.RFC3339Nano), cfg)
	afterLedger, afterSummary := readBackfillArtifacts(t, workspace)
	if !bytes.Equal(beforeLedger, afterLedger) || !bytes.Equal(beforeSummary, afterSummary) {
		t.Fatal("backfill wrote the prepared artifact set after cancellation")
	}
	if !sourceSelected || !errors.Is(ctx.Err(), context.Canceled) || !errors.Is(result.Error, context.Canceled) || result.DBUpdated {
		t.Fatalf("publication did not stop at the late cancellation checkpoint: %+v", result)
	}
}
