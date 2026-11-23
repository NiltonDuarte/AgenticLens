package backfill

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

type PublicationStatus string

const (
	PublicationUnchanged PublicationStatus = "unchanged"
	PublicationPublished PublicationStatus = "published"
	PublicationPreview   PublicationStatus = "preview"
	PublicationFailed    PublicationStatus = "failed"
	PublicationBlocked   PublicationStatus = "blocked"
)

// TicketMigrationResult represents the migration result for an individual ticket workspace.
type TicketMigrationResult struct {
	Publication         PublicationStatus        `json:"publication"`
	Coverage            []HarnessRefreshCoverage `json:"coverage,omitempty"`
	LedgerWritten       bool                     `json:"ledger_written"`
	SummaryWritten      bool                     `json:"summary_written"`
	RefreshIncomplete   bool                     `json:"refresh_incomplete"`
	CollectionsRetained bool                     `json:"collections_retained"`
	Diagnostics         []string                 `json:"diagnostics,omitempty"`
	TicketID            string                   `json:"ticket_id"`
	WorkspacePath       string                   `json:"workspace_path"`
	LedgerPath          string                   `json:"ledger_path"`
	PreviousStatus      string                   `json:"previous_status"`
	NewStatus           string                   `json:"new_status"`
	GitUpdated          bool                     `json:"git_updated"`
	DBUpdated           bool                     `json:"db_updated"`
	DBFound             bool                     `json:"db_found"`
	Skipped             bool                     `json:"skipped"`
	PendingSteps        []string                 `json:"pending_steps,omitempty"`
	Notes               string                   `json:"notes,omitempty"`
	Error               error                    `json:"error,omitempty"`
}

// BackfillReport summarizes the aggregate outcome of a backfill run.
type BackfillReport struct {
	Coverage             []HarnessRefreshTotals  `json:"coverage,omitempty"`
	LedgersWritten       int                     `json:"ledgers_written"`
	SummariesWritten     int                     `json:"summaries_written"`
	StorePreflightFailed bool                    `json:"store_preflight_failed,omitempty"`
	BlockedWork          []string                `json:"blocked_work,omitempty"`
	RefreshIncomplete    int                     `json:"refresh_incomplete"`
	CollectionsRetained  int                     `json:"collections_retained"`
	Scanned              int                     `json:"scanned"`
	Updated              int                     `json:"updated"`
	Skipped              int                     `json:"skipped"`
	Partial              int                     `json:"partial"`
	Completed            int                     `json:"completed"`
	Warnings             []string                `json:"warnings"`
	Results              []TicketMigrationResult `json:"results"`
	SchemaMigration      *SchemaMigrationReport  `json:"schema_migration,omitempty"`
}

// Finalize independent report dimensions from the same per-harness evidence.
func (r *TicketMigrationResult) finalizeCoverage() {
	hasCompleteCoverage := false
	hasIncompleteCoverage := false
	for i := range r.Coverage {
		coverage := &r.Coverage[i]
		switch coverage.Status {
		case HarnessComplete:
			hasCompleteCoverage = true
		case HarnessIncomplete, HarnessUnavailable, HarnessBlocked:
			hasIncompleteCoverage = true
		}
		for j := range coverage.Owners {
			owner := &coverage.Owners[j]
			if owner.Skills == CollectionRetained || owner.Tasks == CollectionRetained {
				r.CollectionsRetained = true
			}
			if r.LedgerWritten {
				continue
			}
			switch owner.Status {
			case OwnerRefreshed:
				owner.Status = OwnerProposed
				coverage.Refreshed--
				coverage.Proposed++
			case OwnerDiscovered:
				owner.Status = OwnerProposedDiscovered
				coverage.Discovered--
				coverage.ProposedDiscovered++
			}
		}
	}
	r.RefreshIncomplete = hasIncompleteCoverage || len(r.Coverage) > 0 && !hasCompleteCoverage
}

func (r *BackfillReport) addPublicationAndCoverage(ticket TicketMigrationResult) {
	if ticket.RefreshIncomplete {
		r.RefreshIncomplete++
	}
	if ticket.CollectionsRetained {
		r.CollectionsRetained++
	}
	if ticket.LedgerWritten {
		r.LedgersWritten++
	}
	if ticket.SummaryWritten {
		r.SummariesWritten++
	}
	for _, coverage := range ticket.Coverage {
		var total *HarnessRefreshTotals
		for i := range r.Coverage {
			if r.Coverage[i].Harness == coverage.Harness {
				total = &r.Coverage[i]
				break
			}
		}
		if total == nil {
			r.Coverage = append(r.Coverage, HarnessRefreshTotals{Harness: coverage.Harness})
			total = &r.Coverage[len(r.Coverage)-1]
		}
		total.Recorded += coverage.Recorded
		total.Refreshed += coverage.Refreshed
		total.VerifiedUnchanged += coverage.VerifiedUnchanged
		total.VerifiedUncompared += coverage.VerifiedUncompared
		total.Retained += coverage.Retained
		total.Discovered += coverage.Discovered
		total.Proposed += coverage.Proposed
		total.ProposedDiscovered += coverage.ProposedDiscovered
		if coverage.OwnerInventoryUnknown {
			total.UnknownOwnerInventories++
		}
	}
}

func (r *BackfillReport) blockedBeforePublication(ctx context.Context, targets []adapters.WorkspaceLocation, opts BackfillOptions, cfg domain.ProjectConfig, unavailable domain.SourceHarness, failure error) {
	r.StorePreflightFailed = unavailable != ""
	if opts.RecomputeGit {
		r.BlockedWork = append(r.BlockedWork, "Git fact publication")
	}
	if opts.MigrateSchema {
		r.BlockedWork = append(r.BlockedWork, "schema migration")
	}
	if opts.UpdateSummaries {
		r.BlockedWork = append(r.BlockedWork, "summary updates")
	}
	for _, target := range targets {
		result := TicketMigrationResult{
			TicketID: target.TicketID, WorkspacePath: target.Path, LedgerPath: target.LedgerPath,
			Publication: PublicationBlocked,
		}
		if unavailable == "" {
			result.Error = failure
		}
		if !opts.RecomputeDB {
			r.Results = append(r.Results, result)
			continue
		}
		var openCode, codex []domain.SessionFact
		ledger, inventoryErr := readPreflightLedger(ctx, target, cfg)
		if inventoryErr == nil {
			result.PreviousStatus = migrationStatus(ledger)
			result.NewStatus = result.PreviousStatus
			if ledger.Migration != nil {
				result.PendingSteps = ledger.Migration.StepsPending
			}
			openCode, codex, _ = sessionsByHarness(ledger.Sessions)
		}
		for _, harness := range selectedHarnesses(opts.Source) {
			status := HarnessBlocked
			reason := failure
			if harness == unavailable {
				status = HarnessUnavailable
			} else if unavailable != "" {
				reason = fmt.Errorf("selected %s store failed preflight", unavailable)
			}
			stored := openCode
			if harness == domain.SourceCodex {
				stored = codex
			}
			coverage := recoveredCoverage(harness, stored, sourceRecovery{failure: reason}, nil)
			coverage.Status = status
			if inventoryErr != nil {
				coverage.OwnerInventoryUnknown = true
				coverage.InventoryError = inventoryErr.Error()
			}
			result.Coverage = append(result.Coverage, coverage)
		}
		result.finalizeCoverage()
		r.addPublicationAndCoverage(result)
		r.Results = append(r.Results, result)
	}
}

func readPreflightLedger(ctx context.Context, target adapters.WorkspaceLocation, cfg domain.ProjectConfig) (_ domain.AgenticFactLedger, resultErr error) {
	if err := ctx.Err(); err != nil {
		return domain.AgenticFactLedger{}, err
	}
	owner, err := target.AcquirePublicationRead()
	if err != nil {
		return domain.AgenticFactLedger{}, fmt.Errorf("acquiring ledger read lease: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, owner.Close()) }()
	content, _, err := readMaintenanceLedger(owner, target.LedgerPath, target.TicketID, cfg.Ticket.CaseInsensitive)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	_, _, ledger, err := readBackfillBaseline(content)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	if err := adapters.AgenticFactLedgerDTOFromDomain(ledger).Validate(); err != nil {
		return domain.AgenticFactLedger{}, fmt.Errorf("validating recorded owner inventory: %w", err)
	}
	return ledger, nil
}

// TicketSchemaMigrationResult represents the schema migration result for an individual ticket workspace.
type TicketSchemaMigrationResult struct {
	TicketID      string `json:"ticket_id"`
	WorkspacePath string `json:"workspace_path"`
	LedgerPath    string `json:"ledger_path"`
	SourceVersion int    `json:"source_version"`
	TargetVersion int    `json:"target_version"`
	Status        string `json:"status"`
	BackupCreated bool   `json:"backup_created"`
	BackupPath    string `json:"backup_path,omitempty"`
	DiffSummary   string `json:"diff_summary"`
	Error         error  `json:"error,omitempty"`
}

// SchemaMigrationReport summarizes the aggregate outcome of a schema migration run.
type SchemaMigrationReport struct {
	Scanned  int                           `json:"scanned"`
	Migrated int                           `json:"migrated"`
	UpToDate int                           `json:"up_to_date"`
	Skipped  int                           `json:"skipped"`
	Errors   []string                      `json:"errors"`
	Results  []TicketSchemaMigrationResult `json:"results"`
}

// migrationAuditSummary keeps migration accounting in command reports before
// completed metadata is omitted from the persisted ledger.
func migrationAuditSummary(info *domain.MigrationInfo) string {
	if info == nil {
		return ""
	}
	var parts []string
	if len(info.StepsCompleted) > 0 {
		parts = append(parts, fmt.Sprintf("completed steps: [%s]", strings.Join(info.StepsCompleted, ", ")))
	}
	if notes := strings.TrimSpace(info.Notes); notes != "" {
		parts = append(parts, notes)
	}
	return strings.Join(parts, "; ")
}
