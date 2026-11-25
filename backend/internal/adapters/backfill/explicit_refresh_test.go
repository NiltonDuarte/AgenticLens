package backfill

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestExplicitBackfillOpenCodeTreeReadBudget(t *testing.T) {
	for _, tc := range []struct {
		name, parent, child, remaining string
		discovered                     []string
	}{
		{name: "parent sorts first", parent: "a-parent", child: "z-child"},
		{name: "child sorts first", parent: "z-parent", child: "a-child"},
		{name: "independent owner exceeds budget", parent: "a-parent", child: "z-child", remaining: "zz-independent"},
		{name: "recorded child before newly discovered siblings", parent: "a-parent", child: "z-recorded", discovered: []string{"a-new", "b-new"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project := newBackfillProject(t)
			localPath := filepath.Join(project, ".agenticlens-config.local.yaml")
			local := readRefreshFile(t, localPath)
			limited := bytes.Replace(local, []byte("max_sessions: 256"), []byte("max_sessions: 2"), 1)
			if bytes.Equal(limited, local) {
				t.Fatal("fixture read budget was not set")
			}
			if err := os.WriteFile(localPath, limited, 0600); err != nil {
				t.Fatal(err)
			}

			ticket, version := "TEST-EXACT-OPENCODE-BUDGET", domain.CurrentSchemaVersion
			priorTools := []domain.SessionToolFact{{Name: "historical", Calls: 1}}
			owners := []domain.SessionFact{
				{ID: tc.parent, Source: domain.SourceOpenCode, Tools: priorTools},
				{ID: tc.child, ParentID: &tc.parent, Source: domain.SourceOpenCode, Tools: priorTools},
			}
			if tc.remaining != "" {
				owners = append(owners, domain.SessionFact{ID: tc.remaining, Source: domain.SourceOpenCode, Tools: priorTools})
			}
			slices.SortFunc(owners, func(a, b domain.SessionFact) int { return strings.Compare(a.ID, b.ID) })
			workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, Sessions: owners})
			dbPath := filepath.Join(project, "source.db")
			setupMockOpenCodeDB(t, dbPath, tc.parent, ticket)
			db, err := sql.Open("sqlite", dbPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if _, err := db.Exec(`INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
				VALUES (?, ?, 'tester', 'native-child', 'child', 1700000000000, 1700000001000)`, tc.child, tc.parent); err != nil {
				t.Fatal(err)
			}
			if tc.remaining != "" {
				if _, err := db.Exec(`INSERT INTO session (id, agent, model, title, time_created, time_updated)
					VALUES (?, 'tester', 'native-independent', 'independent', 1700000000000, 1700000001000)`, tc.remaining); err != nil {
					t.Fatal(err)
				}
			}
			for _, id := range tc.discovered {
				if _, err := db.Exec(`INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
					VALUES (?, ?, 'tester', 'native-new-child', 'new child', 1700000000000, 1700000001000)`, id, tc.parent); err != nil {
					t.Fatal(err)
				}
			}

			options := BackfillOptions{Project: project, TicketID: ticket, Source: "opencode", RecomputeDB: true, DBPath: dbPath}
			report, err := NewEngine().Execute(options)
			if err != nil || report == nil || len(report.Results) != 1 || report.Results[0].Error != nil {
				t.Fatalf("refresh failed: %+v, %v", report, err)
			}
			coverage := report.Results[0].Coverage
			wantCounts := RefreshCounts{Recorded: 2, Refreshed: 2}
			wantStatus, wantIncomplete := HarnessComplete, 0
			if tc.remaining != "" {
				wantCounts.Recorded++
				wantCounts.Retained++
			}
			if tc.remaining != "" || len(tc.discovered) > 0 {
				wantStatus, wantIncomplete = HarnessIncomplete, 1
			}
			if report.RefreshIncomplete != wantIncomplete || len(coverage) != 1 || coverage[0].Status != wantStatus || coverage[0].RefreshCounts != wantCounts {
				t.Fatalf("coverage must distinguish exact budget use from a recorded owner that remains to be read: %+v", report)
			}
			if tc.remaining != "" {
				if !slices.ContainsFunc(coverage[0].Owners, func(owner OwnerRefreshCoverage) bool {
					return owner.SessionID == tc.remaining && owner.Status == OwnerRetained && strings.Contains(owner.Reason, "budget")
				}) {
					t.Errorf("expected the independent owner to be retained for budget exhaustion: %+v", coverage)
				}
				accepted, _ := readRecoveryLedger(t, workspace)
				if !slices.ContainsFunc(accepted.ToDomain().Sessions, func(owner domain.SessionFact) bool {
					return owner.ID == tc.remaining && reflect.DeepEqual(owner.Tools, priorTools)
				}) {
					t.Errorf("budget-limited owner's accepted facts changed: %+v", accepted.Sessions)
				}
			}
			assertExplicitWrites(t, report, true, false)
			if len(tc.discovered) > 0 {
				if !coverage[0].BudgetLimited || !slices.ContainsFunc(coverage[0].Owners, func(owner OwnerRefreshCoverage) bool {
					return owner.SessionID == tc.child && owner.Status == OwnerRefreshed
				}) {
					t.Fatalf("new siblings must not displace the recorded child within the budget: %+v", coverage)
				}
				ledgerPath := filepath.Join(workspace, "agentic_metrics.json")
				accepted, stamp := readRefreshFile(t, ledgerPath), markRefreshFile(t, ledgerPath)
				repeat, err := NewEngine().Execute(options)
				if err != nil || repeat == nil || len(repeat.Results) != 1 || repeat.Results[0].Error != nil || len(repeat.Results[0].Coverage) != 1 {
					t.Fatalf("equivalent exact-ticket retry failed: %+v, %v", repeat, err)
				}
				row := repeat.Results[0].Coverage[0]
				if !row.BudgetLimited || row.Status != HarnessIncomplete || row.RefreshCounts != (RefreshCounts{Recorded: 2, VerifiedUnchanged: 2}) {
					t.Errorf("exact-ticket retry with the same cap must keep recorded verification distinct from limited discovery: %+v", row)
				}
				assertExplicitWrites(t, repeat, false, false)
				assertRefreshFile(t, ledgerPath, accepted, stamp)
			}
		})
	}
}

func TestExplicitBackfillPartialRepeatAndIndependentSummary(t *testing.T) {
	project := newBackfillProject(t)
	version, ticket, parent := domain.CurrentSchemaVersion, "TEST-EXPLICIT-REFRESH", "local-parent"
	workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket,
		Migration: &domain.MigrationInfo{
			SchemaVersion: version, Status: domain.MigrationStatusPartial,
			StepsCompleted: []string{domain.MigrationStepGitCodeFacts},
			StepsPending:   []string{domain.MigrationStepSourceBackfill},
		},
		Sessions: []domain.SessionFact{
			{ID: parent, Source: domain.SourceCodex, Tools: []domain.SessionToolFact{{Name: "old", Calls: 1}}},
			{ID: "missing-child", ParentID: &parent, Source: domain.SourceCodex, User: "other-machine", Skills: []domain.SkillFact{{Name: "child-skill"}}, Delegations: []domain.DelegationFact{{Description: "child-task"}}},
			{ID: "other-harness", Source: domain.SourceOpenCode, User: "other-machine", Tools: []domain.SessionToolFact{{Name: "other-tool", Calls: 2}}},
		},
	})
	home := t.TempDir()
	sessions := filepath.Join(home, "sessions")
	if err := os.Mkdir(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	writeBackfillRollout(t, sessions, parent, "", "coder", "model", map[string]int{"fresh": 3}, "")
	writeBackfillRollout(t, sessions, "new-child", parent, "tester", "model", nil, "")
	ledgerPath, summaryPath := filepath.Join(workspace, "agentic_metrics.json"), filepath.Join(workspace, "SUMMARY.md")
	const authoredStart = "# Handwritten heading\n\nReview notes above metrics.\n\n"
	const authoredEnd = "\n## Handoff\nPreserve these notes too.\n"
	stale := []byte(authoredStart + "<!-- agentic-lens:metrics:start -->\nStale metrics\n<!-- agentic-lens:metrics:end -->\n" + authoredEnd)
	if err := os.WriteFile(summaryPath, stale, 0644); err != nil {
		t.Fatal(err)
	}
	initial, _ := readRecoveryLedger(t, workspace)
	before := initial.ToDomain()
	options := BackfillOptions{Project: project, TicketID: ticket, Source: "codex", CodexHome: home, RecomputeDB: true, UpdateSummaries: true,
		SnapshotCutoff: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}
	run := func(ledgerWrite, summaryWrite bool, parentStatus OwnerRefreshStatus) {
		t.Helper()
		report, err := NewEngine().Execute(options)
		if err != nil || report == nil || len(report.Results) != 1 || report.Results[0].Error != nil {
			t.Fatalf("backfill failed: %+v, %v", report, err)
		}
		if report.Updated != 1 || report.RefreshIncomplete != 1 || report.Results[0].NewStatus != string(domain.MigrationStatusPartial) {
			t.Fatalf("expected a processed partial refresh: %+v", report)
		}
		assertExplicitWrites(t, report, ledgerWrite, summaryWrite)
		coverage := report.Results[0].Coverage
		if len(coverage) != 1 || coverage[0].Harness != domain.SourceCodex || coverage[0].Status != HarnessIncomplete {
			t.Fatalf("coverage must include only the incomplete selected harness: %+v", coverage)
		}
		row := coverage[0]
		wantCounts := RefreshCounts{Recorded: 3, VerifiedUnchanged: 2, Retained: 1}
		newChildStatus := OwnerVerifiedUnchanged
		if parentStatus == OwnerRefreshed {
			wantCounts = RefreshCounts{Recorded: 2, Refreshed: 1, Retained: 1, Discovered: 1}
			newChildStatus = OwnerDiscovered
		}
		if row.RefreshCounts != wantCounts || len(row.Owners) != 3 {
			t.Fatalf("coverage does not reconcile with recorded Codex owners: %+v", row)
		}
		for _, owner := range row.Owners {
			switch owner.SessionID {
			case parent:
				if owner.Status != parentStatus || owner.Skills != CollectionEmpty || owner.Tasks != CollectionEmpty {
					t.Errorf("the available parent must report verified empty collections: %+v", owner)
				}
			case "missing-child":
				if owner.Status != OwnerRetained || !strings.Contains(owner.Reason, "not found") || owner.Skills != CollectionRetained || owner.Tasks != CollectionRetained {
					t.Errorf("the unavailable child must report retained collections and a reason: %+v", owner)
				}
			case "new-child":
				if owner.Status != newChildStatus {
					t.Errorf("newly discovered owner must enter the denominator only after publication: %+v", owner)
				}
			default:
				t.Errorf("unexpected recorded owner: %+v", owner)
			}
		}
		options.SnapshotCutoff = options.SnapshotCutoff.Add(time.Hour)
	}
	run(true, true, OwnerRefreshed)
	first, accepted := readRecoveryLedger(t, workspace)
	if first.UpdatedAt == initial.UpdatedAt || first.Migration == nil || len(first.Migration.StepsPending) == 0 {
		t.Fatalf("changed refresh did not publish or lost the pending migration: %+v", first)
	}
	owners := make(map[string]domain.SessionFact)
	for _, owner := range first.ToDomain().Sessions {
		owners[owner.ID] = owner
	}
	if tools := owners[parent].Tools; len(tools) != 1 || tools[0].Name != "fresh" || tools[0].Calls != 3 {
		t.Fatalf("parent's source facts were not published: %+v", tools)
	}
	for _, owner := range before.Sessions[1:] {
		if !reflect.DeepEqual(owner, owners[owner.ID]) {
			t.Errorf("stored facts changed for %s: got %+v, want %+v", owner.ID, owners[owner.ID], owner)
		}
	}
	fresh := readRefreshFile(t, summaryPath)
	if !bytes.Contains(fresh, []byte("Workflow Efficiency")) || !bytes.HasPrefix(fresh, []byte(authoredStart)) || !bytes.HasSuffix(fresh, []byte(authoredEnd)) {
		t.Fatalf("refresh lost managed metrics or authored text:\n%s", fresh)
	}
	ledgerStamp, summaryStamp := markRefreshFile(t, ledgerPath), markRefreshFile(t, summaryPath)
	run(false, false, OwnerVerifiedUnchanged)
	assertRefreshFile(t, ledgerPath, accepted, ledgerStamp)
	assertRefreshFile(t, summaryPath, fresh, summaryStamp)

	if err := os.WriteFile(summaryPath, stale, 0644); err != nil {
		t.Fatal(err)
	}
	options.UpdateSummaries = false
	summaryStamp = markRefreshFile(t, summaryPath)
	run(false, false, OwnerVerifiedUnchanged)
	assertRefreshFile(t, summaryPath, stale, summaryStamp)

	options.UpdateSummaries = true
	run(false, true, OwnerVerifiedUnchanged)
	if got := readRefreshFile(t, summaryPath); !bytes.Equal(got, fresh) {
		t.Fatalf("stale managed metrics were not repaired exactly:\n%s", got)
	}
	assertRefreshFile(t, ledgerPath, accepted, ledgerStamp)

	if err := os.Remove(summaryPath); err != nil {
		t.Fatal(err)
	}
	run(false, true, OwnerVerifiedUnchanged)
	if got := readRefreshFile(t, summaryPath); !strings.Contains(string(got), "Workflow Efficiency") {
		t.Fatalf("missing summary was not recreated with metrics:\n%s", got)
	}
	assertRefreshFile(t, ledgerPath, accepted, ledgerStamp)
}

func TestExplicitBackfillRetainedSkillsCountAsIncomplete(t *testing.T) {
	options, workspace := openCodeRefreshFixture(t)
	options.UpdateSummaries = false
	setRefreshFixturePolicy(t, options.Project, "skills: true", "skills: false")
	report := runOpenCodeRefreshFixture(t, options)
	result := report.Results[0]
	if !result.RefreshIncomplete || report.RefreshIncomplete != 1 || !result.CollectionsRetained || report.CollectionsRetained != 1 || len(result.Coverage) != 1 {
		t.Fatalf("a retained collection must affect harness, ticket, and batch coverage: %+v", report)
	}
	coverage := result.Coverage[0]
	if coverage.Status != HarnessIncomplete || coverage.Refreshed != 1 || coverage.Retained != 0 || len(coverage.Owners) != 1 || coverage.Owners[0].Skills != CollectionRetained {
		t.Fatalf("a refreshed owner still has its unavailable skill collection retained: %+v", coverage)
	}
	accepted, _ := readRecoveryLedger(t, workspace)
	if !reflect.DeepEqual(accepted.ToDomain().SkillFacts(), []domain.SkillFact{{Name: "historical-skill"}}) {
		t.Fatalf("unavailable source collection erased historical skills: %+v", accepted.Sessions)
	}
	assertExplicitWrites(t, report, true, false)
}

func TestBackfillLedgerStagingPreservesConfiguredSummary(t *testing.T) {
	options, workspace := openCodeRefreshFixture(t)
	ledgerPath, ordinarySummary := filepath.Join(workspace, "agentic_metrics.json"), filepath.Join(workspace, "SUMMARY.md")
	prior := readRefreshFile(t, ledgerPath)
	assertExplicitWrites(t, runOpenCodeRefreshFixture(t, options), true, true)
	// Preserve the summary rendered from current source facts, then restore the
	// older ledger so only the ledger needs to be published on the next pass.
	summaryPath := ledgerPath + ".tmp"
	if err := os.Rename(ordinarySummary, summaryPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledgerPath, prior, 0600); err != nil {
		t.Fatal(err)
	}
	setRefreshFixturePolicy(t, options.Project, "filename: SUMMARY.md", "filename: agentic_metrics.json.tmp")
	want, stamp := readRefreshFile(t, summaryPath), markRefreshFile(t, summaryPath)
	assertExplicitWrites(t, runOpenCodeRefreshFixture(t, options), true, false)
	if bytes.Equal(prior, readRefreshFile(t, ledgerPath)) {
		t.Fatal("expected changed source facts to replace the ledger")
	}
	assertRefreshFile(t, summaryPath, want, stamp)
}

func openCodeRefreshFixture(t *testing.T) (BackfillOptions, string) {
	t.Helper()
	project := newBackfillProject(t)
	version, ticket, owner := domain.CurrentSchemaVersion, "TEST-OPENCODE-REFRESH", "local-owner"
	workspace := createTestWorkspace(t, project, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket,
		Sessions: []domain.SessionFact{{ID: owner, Source: domain.SourceOpenCode,
			Skills: []domain.SkillFact{{Name: "historical-skill"}},
			Tools:  []domain.SessionToolFact{{Name: "historical-tool", Calls: 1}}}},
	})
	dbPath := filepath.Join(project, "source.db")
	setupMockOpenCodeDB(t, dbPath, owner, ticket)
	options := BackfillOptions{Project: project, TicketID: ticket, Source: "opencode", DBPath: dbPath, RecomputeDB: true, UpdateSummaries: true,
		SnapshotCutoff: time.Now().Add(time.Second)}
	return options, workspace
}

func runOpenCodeRefreshFixture(t *testing.T, options BackfillOptions) *BackfillReport {
	t.Helper()
	report, err := NewEngine().Execute(options)
	if err != nil || report == nil || len(report.Results) != 1 || report.Results[0].Error != nil {
		t.Fatalf("OpenCode refresh fixture failed: %+v, %v", report, err)
	}
	return report
}

func setRefreshFixturePolicy(t *testing.T, project, old, new string) {
	t.Helper()
	path := filepath.Join(project, ".agenticlens-config.yaml")
	prior := readRefreshFile(t, path)
	if bytes.Count(prior, []byte(old)) != 1 {
		t.Fatalf("expected one fixture setting %q", old)
	}
	if err := os.WriteFile(path, bytes.Replace(prior, []byte(old), []byte(new), 1), 0600); err != nil {
		t.Fatal(err)
	}
}

func assertExplicitWrites(t *testing.T, report *BackfillReport, ledger, summary bool) {
	t.Helper()
	wantLedgers, wantSummaries := 0, 0
	if ledger {
		wantLedgers = 1
	}
	if summary {
		wantSummaries = 1
	}
	result := report.Results[0]
	if result.LedgerWritten != ledger || result.SummaryWritten != summary || report.LedgersWritten != wantLedgers || report.SummariesWritten != wantSummaries {
		t.Errorf("expected ledger/summary writes %t/%t and matching aggregate counts: %+v", ledger, summary, report)
	}
}

func readRefreshFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func markRefreshFile(t *testing.T, path string) time.Time {
	t.Helper()
	stamp := time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func assertRefreshFile(t *testing.T, path string, want []byte, stamp time.Time) {
	t.Helper()
	if got := readRefreshFile(t, path); !bytes.Equal(got, want) {
		t.Errorf("%s: contents changed on equivalent refresh", filepath.Base(path))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(stamp) {
		t.Errorf("%s: unchanged file replaced, mtime %s -> %s", filepath.Base(path), stamp, info.ModTime())
	}
}
