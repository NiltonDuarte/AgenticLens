package backfill

import (
	"bytes"
	"path/filepath"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestBackfillRetainsLedgerTimestampUntilFactsChange(t *testing.T) {
	const firstAuthor, secondAuthor = "First backfill author", "Second backfill author"
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "user.name")
	t.Setenv("GIT_CONFIG_VALUE_0", firstAuthor)
	root := newBackfillProject(t)
	version, ticket, title := domain.CurrentSchemaVersion, "TEST-TIMESTAMPS", "Ledger timestamps"
	completedAt := "2026-10-26T10:00:00Z"
	workspace := createTestWorkspace(t, root, ticket, domain.AgenticFactLedger{
		SchemaVersion: &version,
		TicketID:      &ticket,
		TicketTitle:   &title,
		CompletedAt:   &completedAt,
		Migration: &domain.MigrationInfo{
			SchemaVersion: version,
			Status:        domain.MigrationStatusPartial,
			StepsCompleted: []string{
				domain.MigrationStepGitCodeFacts,
			},
			StepsPending:   []string{"review_historical_coverage"},
			LastMigratedAt: completedAt,
		},
		Sessions: []domain.SessionFact{{
			ID: "root", Source: domain.SourceCodex, User: "fixture-user",
			Invocations: []domain.InvocationFact{{Role: "coder", Model: "original"}},
		}},
	})
	sessionsDir := filepath.Join(root, "sessions")
	writeBackfillRollout(t, sessionsDir, "root", "", "coder", "first-source", map[string]int{"read": 1}, "")

	opts := BackfillOptions{Project: root, TicketID: ticket, Source: "codex", CodexHome: root, RecomputeDB: true}
	refresh := func(at time.Time) {
		t.Helper()
		opts.SnapshotCutoff = at
		report, err := NewEngine().Execute(opts)
		if err != nil || len(report.Results) != 1 || report.Results[0].Error != nil || !report.Results[0].DBUpdated || report.Results[0].RefreshIncomplete {
			t.Fatalf("expected a complete source-backed refresh: %+v, %v", report, err)
		}
	}
	cutoff := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	refresh(cutoff)
	first, baseline := readRecoveryLedger(t, workspace)
	if first.UpdatedAt != cutoff.Format(time.RFC3339Nano) || first.Sessions[0].Invocations[0].Model != "first-source" {
		t.Fatalf("first source refresh did not publish changed facts: updated_at=%q, sessions=%+v", first.UpdatedAt, first.Sessions)
	}
	if first.Migration == nil || len(first.Migration.StepsPending) != 1 || first.Migration.StepsPending[0] != "review_historical_coverage" || first.Migration.LastMigratedAt != cutoff.Format(time.RFC3339Nano) || first.Migration.LastMigratedBy != firstAuthor {
		t.Fatalf("expected first source refresh to retain the pending step and record migration time and author: %+v", first.Migration)
	}
	t.Setenv("GIT_CONFIG_VALUE_0", secondAuthor)
	refresh(cutoff.Add(time.Hour))
	repeated, repeatedBytes := readRecoveryLedger(t, workspace)
	if !bytes.Equal(repeatedBytes, baseline) {
		t.Errorf("an unchanged backfill rewrote the ledger: updated_at %q -> %q", first.UpdatedAt, repeated.UpdatedAt)
	}

	writeBackfillRollout(t, sessionsDir, "root", "", "coder", "second-source", map[string]int{"read": 1}, "")
	changedAt := cutoff.Add(2 * time.Hour)
	refresh(changedAt)
	changed, _ := readRecoveryLedger(t, workspace)
	if changed.UpdatedAt != changedAt.Format(time.RFC3339Nano) || changed.Sessions[0].Invocations[0].Model != "second-source" {
		t.Fatalf("the changed backfill did not publish both the new model and updated_at: updated_at=%q, sessions=%+v", changed.UpdatedAt, changed.Sessions)
	}
	if changed.Migration == nil || changed.Migration.LastMigratedBy != secondAuthor || changed.Migration.LastMigratedAt != changedAt.Format(time.RFC3339Nano) {
		t.Fatalf("the changed backfill did not record the new migration author and time: %+v", changed.Migration)
	}
	if changed.CompletedAt == nil || *changed.CompletedAt != completedAt {
		t.Fatalf("backfill must retain the existing completion timestamp, got %v", changed.CompletedAt)
	}
}
