package domain

import (
	"sort"
	"strings"
)

// MigrationEvaluationInput contains all parameters needed by EvaluateMigrationStatus.
type MigrationEvaluationInput struct {
	ExistingMigration  *MigrationInfo
	RecomputeGit       bool
	GitSucceeded       bool
	RecomputeDB        bool
	DBAttempted        bool
	DBSessionFound     bool
	MigratedBy         string
	MigratedAt         string
	HasExistingDBFacts bool
}

// EvaluateMigrationStatus is a pure function that deterministically derives the new MigrationInfo.
// It guards against data loss, records completed and pending steps, and evaluates migration status.
func EvaluateMigrationStatus(input MigrationEvaluationInput) MigrationInfo {
	completedSet := make(map[string]bool)
	pendingSet := make(map[string]bool)

	// Preserve prior steps if existing migration exists
	if input.ExistingMigration != nil {
		for _, s := range input.ExistingMigration.StepsCompleted {
			clean := strings.TrimSpace(s)
			if clean != "" {
				completedSet[clean] = true
			}
		}
		for _, s := range input.ExistingMigration.StepsPending {
			clean := strings.TrimSpace(s)
			if clean != "" {
				pendingSet[clean] = true
			}
		}
	}

	var notes []string

	// Process Git code facts step
	if input.RecomputeGit {
		if input.GitSucceeded {
			completedSet[MigrationStepGitCodeFacts] = true
			delete(pendingSet, MigrationStepGitCodeFacts)
		} else if !completedSet[MigrationStepGitCodeFacts] {
			pendingSet[MigrationStepGitCodeFacts] = true
		}
	} else if input.GitSucceeded {
		// Existing ledger already has valid git facts
		completedSet[MigrationStepGitCodeFacts] = true
		delete(pendingSet, MigrationStepGitCodeFacts)
	}

	// Process OpenCode DB step
	if input.RecomputeDB {
		if input.DBAttempted && input.DBSessionFound {
			completedSet[MigrationStepOpenCodeDB] = true
			delete(pendingSet, MigrationStepOpenCodeDB)
		} else {
			// Session was not found in local database
			if !completedSet[MigrationStepOpenCodeDB] {
				pendingSet[MigrationStepOpenCodeDB] = true
			}
			if input.HasExistingDBFacts {
				notes = append(notes, "Local SQLite session not found; preserved existing invocations and telemetry for subsequent host migration")
			} else {
				notes = append(notes, "Local SQLite session not found; pending database resolution on author host")
			}
		}
	}

	// If no recompute flags were set, but existing migration had completed/pending steps, preserve them.
	// If both steps are already completed, it's completed.
	// Otherwise evaluate status:
	// If there are pending steps, status is partial (or pending if nothing completed yet).
	var stepsCompleted []string
	for step := range completedSet {
		stepsCompleted = append(stepsCompleted, step)
	}
	sort.Strings(stepsCompleted)

	var stepsPending []string
	for step := range pendingSet {
		stepsPending = append(stepsPending, step)
	}
	sort.Strings(stepsPending)

	var status MigrationStatus
	if len(stepsPending) == 0 {
		status = MigrationStatusCompleted
	} else if len(stepsCompleted) > 0 {
		status = MigrationStatusPartial
	} else {
		status = MigrationStatusPending
	}

	author := strings.TrimSpace(input.MigratedBy)
	if author == "" {
		if input.ExistingMigration != nil && input.ExistingMigration.LastMigratedBy != "" {
			author = input.ExistingMigration.LastMigratedBy
		} else {
			author = "unknown"
		}
	}

	ts := strings.TrimSpace(input.MigratedAt)
	if ts == "" {
		if input.ExistingMigration != nil && input.ExistingMigration.LastMigratedAt != "" {
			ts = input.ExistingMigration.LastMigratedAt
		}
	}

	var noteText string
	if len(notes) > 0 {
		noteText = strings.Join(notes, "; ")
	} else if status == MigrationStatusCompleted {
		// When fully completed, clear prior pending/fallback notes
		noteText = ""
	} else if input.ExistingMigration != nil {
		noteText = input.ExistingMigration.Notes
	}

	return MigrationInfo{
		SchemaVersion:  CurrentSchemaVersion,
		Status:         status,
		StepsCompleted: stepsCompleted,
		StepsPending:   stepsPending,
		LastMigratedAt: ts,
		LastMigratedBy: author,
		Notes:          noteText,
	}
}
