package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func TestCLIBackfillSelectedStoreAndGitPreflight(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	project := filepath.Dir(workspaces)
	enableCLIFixtureSummary(t, project)
	dbPath := setupIsolatedEmptyDB(t)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`INSERT INTO session (id, agent, model, title, time_created, time_updated)
		VALUES ('opencode-historical', 'coder', 'native-model', 'TICKET-1', 1700000000000, 1700000002000)`); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, dbPath, project)
	workspace := filepath.Join(workspaces, "TICKET-1")
	ledgerPath := filepath.Join(workspace, "agentic_metrics.json")
	ticket, version := "TICKET-1", domain.CurrentSchemaVersion
	if err := adapters.SaveFactLedger(ledgerPath, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &ticket,
		Sessions: []domain.SessionFact{
			{ID: "codex-historical", Source: domain.SourceCodex, User: "fixture-user", Skills: []domain.SkillFact{{Name: "historical-skill"}}},
			{ID: "opencode-historical", Source: domain.SourceOpenCode, User: "fixture-user", Delegations: []domain.DelegationFact{{Description: "historical-task"}}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	accepted := readCaptureFile(t, ledgerPath)
	summaryPath := filepath.Join(workspace, "SUMMARY.md")
	summary := []byte("# Authored evidence\nThis summary must survive a preflight failure.\n")
	if err := os.WriteFile(summaryPath, summary, 0600); err != nil {
		t.Fatal(err)
	}
	ledgerStamp, summaryStamp := markCLIBackfillFile(t, ledgerPath), markCLIBackfillFile(t, summaryPath)
	missing := filepath.Join(project, "missing-codex-store")
	const invalidRef = "missing-selected-code-ref"
	for _, tc := range []struct{ name, source, codeRef string }{
		{name: "only missing store selected", source: "codex"},
		{name: "missing store blocks second source and valid Git", source: "both", codeRef: "HEAD"},
		{name: "missing store takes precedence over invalid Git", source: "both", codeRef: invalidRef},
		{name: "invalid Git after successful store preflight", source: "opencode", codeRef: invalidRef},
	} {
		t.Run(tc.name, func(t *testing.T) {
			flags := []string{"backfill", "--source", tc.source, "--recompute-db", "--update-summaries",
				"--ticket-id", ticket, "--db-path", dbPath}
			if tc.source != "opencode" {
				flags = append(flags, "--codex-home", missing)
			}
			if tc.codeRef != "" {
				flags = append(flags, "--recompute-git", "--code-ref", tc.codeRef, "--migrate-schema")
			}
			cmd := workspaceCLICommand(t, workspaces, flags...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			out := strings.ToLower(stdout.String() + stderr.String())
			assertBackfillOutput(t, out, "ticket-1", "ledgers written: 0", "summaries written: 0", "updated: 0", "summary updates")
			if tc.source == "opencode" {
				if err != nil {
					t.Fatalf("failed Git must keep the legacy exit 0 after valid store preflight: %v\n%s", err, out)
				}
				assertBackfillOutput(t, out, invalidRef, "requested git measurement failed before publication", "publication: blocked")
				_, detail, _ := strings.Cut(out, "session opencode-historical: retained")
				detail, _, _ = strings.Cut(detail, "\n")
				assertBackfillOutput(t, detail, "git")
			} else {
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(stderr.String(), "missing-codex-store") {
					t.Fatalf("expected missing-store exit 1 and original diagnostic, got %v\nstdout: %s\nstderr: %s", err, &stdout, &stderr)
				}
				assertBackfillOutput(t, out, "codex refresh: unavailable", "session codex-historical: retained",
					"codex refresh totals | recorded: 1 | refreshed: 0 | verified unchanged: 0 | retained: 1")
			}
			if tc.source != "codex" {
				assertBackfillOutput(t, out, "opencode refresh: blocked", "schema migration", "session opencode-historical: retained",
					"opencode refresh totals | recorded: 1 | refreshed: 0 | verified unchanged: 0 | retained: 1")
			} else if strings.Contains(out, "opencode refresh") || strings.Contains(out, "opencode-historical") {
				t.Errorf("preflight reported an unselected harness:\n%s", out)
			}
			if tc.source == "both" {
				assertBackfillOutput(t, out, "git fact publication")
			}
			assertCLIBackfillFile(t, ledgerPath, accepted, ledgerStamp)
			assertCLIBackfillFile(t, summaryPath, summary, summaryStamp)
		})
	}
}

func TestCLIBackfillRepairsSummaryWithoutRecomputingFacts(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	project := filepath.Dir(workspaces)
	enableCLIFixtureSummary(t, project)
	workspace := filepath.Join(workspaces, "TICKET-1")
	ledgerPath, summaryPath := filepath.Join(workspace, "agentic_metrics.json"), filepath.Join(workspace, "SUMMARY.md")
	version, ticket := domain.CurrentSchemaVersion, "TICKET-1"
	if err := adapters.SaveFactLedger(ledgerPath, domain.AgenticFactLedger{
		SchemaVersion: &version, TicketID: &ticket, TicketTitle: &ticket,
		Sessions: []domain.SessionFact{{ID: "stored-owner", Source: domain.SourceOpenCode, User: "fixture-user"}},
	}); err != nil {
		t.Fatal(err)
	}
	const authored = "# Maintainer notes\n\nThis evidence must remain intact.\n\n"
	stale := []byte(authored + "<!-- agentic-lens:metrics:start -->\nStale metrics\n<!-- agentic-lens:metrics:end -->\n")
	if err := os.WriteFile(summaryPath, stale, 0600); err != nil {
		t.Fatal(err)
	}
	ledger, ledgerStamp := readCaptureFile(t, ledgerPath), markCLIBackfillFile(t, ledgerPath)
	run := func() string {
		t.Helper()
		cmd := workspaceCLICommand(t, workspaces, "backfill", "--source", "opencode", "--ticket-id", ticket,
			"--update-summaries", "--db-path", filepath.Join(project, "unavailable-source.db"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("summary-only backfill failed: %v\n%s", err, out)
		}
		return strings.ToLower(string(out))
	}
	first := run()
	assertBackfillOutput(t, first, "scanned: 1 | updated: 1", "ledgers written: 0 | summaries written: 1")
	updated := readCaptureFile(t, summaryPath)
	if !bytes.HasPrefix(updated, []byte(authored)) || !bytes.Contains(updated, []byte("Workflow Efficiency")) || bytes.Contains(updated, []byte("Stale metrics")) {
		t.Fatalf("stale metrics were not independently repaired around authored text:\n%s", updated)
	}
	assertCLIBackfillFile(t, ledgerPath, ledger, ledgerStamp)
	summaryStamp := markCLIBackfillFile(t, summaryPath)
	assertBackfillOutput(t, run(), "ledgers written: 0 | summaries written: 0")
	assertCLIBackfillFile(t, ledgerPath, ledger, ledgerStamp)
	assertCLIBackfillFile(t, summaryPath, updated, summaryStamp)
}

func TestCLIBackfillOpenCodeBatchBudgetReportsRetry(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	project := filepath.Dir(workspaces)
	localPath := filepath.Join(project, ".agenticlens-config.local.yaml")
	local := readCaptureFile(t, localPath)
	limited := bytes.Replace(local, []byte("max_sessions: 256"), []byte("max_sessions: 1"), 1)
	if bytes.Equal(local, limited) {
		t.Fatal("fixture capture budget was not changed")
	}
	if err := os.WriteFile(localPath, limited, 0600); err != nil {
		t.Fatal(err)
	}
	dbPath := setupIsolatedEmptyDB(t)
	setOpenCodeFixtureProject(t, dbPath, project)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	paths := make([]string, 2)
	prior := make([][]byte, 2)
	for i := range paths {
		ticket, owner := fmt.Sprintf("TICKET-%d", i+1), fmt.Sprintf("budget-owner-%d", i+1)
		if _, err := db.Exec(`INSERT INTO session (id, directory, agent, model, title, time_created, time_updated)
			VALUES (?, ?, 'coder', 'source-model', ?, 1700000000000, 1700000002000)`, owner, project, ticket); err != nil {
			t.Fatal(err)
		}
		version := domain.CurrentSchemaVersion
		paths[i] = filepath.Join(workspaces, ticket, "agentic_metrics.json")
		ledger := domain.AgenticFactLedger{SchemaVersion: &version, TicketID: &ticket, TicketTitle: &ticket,
			Sessions: []domain.SessionFact{{ID: owner, Source: domain.SourceOpenCode, User: "fixture-user", Tools: []domain.SessionToolFact{{Name: "stored-tool", Calls: 1}}}}}
		if err := adapters.SaveFactLedger(paths[i], ledger); err != nil {
			t.Fatal(err)
		}
		prior[i] = readCaptureFile(t, paths[i])
	}
	run := func(args ...string) string {
		t.Helper()
		flags := []string{"backfill", "--source", "opencode", "--recompute-db", "--db-path", dbPath}
		cmd := workspaceCLICommand(t, workspaces, append(flags, args...)...)
		data, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("backfill failed: %v\n%s", err, data)
		}
		return strings.ToLower(string(data))
	}
	batch := run()
	assertBackfillOutput(t, batch, "scanned: 2 | updated: 2", "ledgers written: 1 | summaries written: 0", "incomplete: 1", "session budget-owner-2: retained", "source recovery budget", `--ticket-id "ticket-2"`, "--source opencode", "same configured session and byte limits")
	assertBackfillOutput(t, batch, "opencode refresh: complete | recorded: 1 | refreshed: 1 | verified unchanged: 0 | retained: 0",
		"opencode refresh: incomplete", "recorded: 1 | refreshed: 0 | verified unchanged: 0 | retained: 1",
		"opencode refresh totals | recorded: 2 | refreshed: 1 | verified unchanged: 0 | retained: 1",
		"retained skill/task snapshots: 0")
	_, retainedOwner, _ := strings.Cut(batch, "session budget-owner-2: retained")
	retainedOwner, _, _ = strings.Cut(retainedOwner, "\n")
	assertBackfillOutput(t, retainedOwner, "source recovery budget", "skills: unavailable | tasks: unavailable")
	accepted := readCaptureFile(t, paths[0])
	if bytes.Equal(accepted, prior[0]) || !bytes.Equal(readCaptureFile(t, paths[1]), prior[1]) {
		t.Fatalf("the first ledger should be refreshed and the budget-limited one retained:\n%s", batch)
	}
	retry := run("--ticket-id", "TICKET-2")
	assertBackfillOutput(t, retry, "scanned: 1 | updated: 1", "ledgers written: 1 | summaries written: 0", "incomplete: 0",
		"opencode refresh: complete | recorded: 1 | refreshed: 1 | verified unchanged: 0 | retained: 0")
	if !bytes.Equal(readCaptureFile(t, paths[0]), accepted) || bytes.Equal(readCaptureFile(t, paths[1]), prior[1]) {
		t.Fatalf("retry did not refresh exactly the retained ledger:\n%s", retry)
	}

	// With Git and source recomputation requested together, an equivalent
	// owner must stay verified unchanged while a new Git delta still commits.
	if err := os.Mkdir(filepath.Join(project, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "src", "change.go"), []byte("package fixture\n\nconst value = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	historyGit(t, project, "add", "src/change.go")
	historyGit(t, project, "commit", "-m", "add measured source")
	combined := run("--ticket-id", "TICKET-2", "--recompute-git", "--code-ref", "HEAD")
	assertBackfillOutput(t, combined, "ledgers written: 1 | summaries written: 0",
		"opencode refresh: complete | recorded: 1 | refreshed: 0 | verified unchanged: 1 | retained: 0")
	combinedLedger := readCaptureFile(t, paths[1])
	var result adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(combinedLedger, &result); err != nil {
		t.Fatal(err)
	}
	if result.Code.ProdLinesAdded != 3 || result.Code.FilesChanged != 1 {
		t.Fatalf("combined refresh lost selected Git delta: %+v\n%s", result.Code, combined)
	}
	repeat := run("--ticket-id", "TICKET-2", "--recompute-git", "--code-ref", "HEAD")
	assertBackfillOutput(t, repeat, "scanned: 1 | updated: 1", "ledgers written: 0 | summaries written: 0",
		"opencode refresh: complete | recorded: 1 | refreshed: 0 | verified unchanged: 1 | retained: 0")
	if !bytes.Equal(combinedLedger, readCaptureFile(t, paths[1])) {
		t.Fatalf("equivalent combined refresh rewrote selected ledger:\n%s", repeat)
	}
}

func markCLIBackfillFile(t *testing.T, path string) time.Time {
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

func assertCLIBackfillFile(t *testing.T, path string, contents []byte, stamp time.Time) {
	t.Helper()
	if !bytes.Equal(contents, readCaptureFile(t, path)) {
		t.Errorf("%s: accepted contents changed", filepath.Base(path))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(stamp) {
		t.Errorf("%s: expected modification time %s, got %s", filepath.Base(path), stamp, info.ModTime())
	}
}

func assertBackfillOutput(t *testing.T, output string, expected ...string) {
	t.Helper()
	for _, part := range expected {
		if !strings.Contains(output, part) {
			t.Errorf("backfill report missing %q:\n%s", part, output)
		}
	}
}
