package domain

import (
	"slices"
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
	var stepsCompleted, stepsPending []string

	// Preserve prior steps if existing migration exists
	if input.ExistingMigration != nil {
		for _, s := range input.ExistingMigration.StepsCompleted {
			stepsCompleted = appendMigrationStep(stepsCompleted, completedSet, s)
		}
		for _, s := range input.ExistingMigration.StepsPending {
			stepsPending = appendMigrationStep(stepsPending, pendingSet, s)
		}
	}

	var notes []string

	// Newly recorded steps follow workflow order: Git facts, then source recovery.
	if input.GitSucceeded {
		stepsCompleted = appendMigrationStep(stepsCompleted, completedSet, MigrationStepGitCodeFacts)
		delete(pendingSet, MigrationStepGitCodeFacts)
	} else if input.RecomputeGit && !completedSet[MigrationStepGitCodeFacts] {
		stepsPending = appendMigrationStep(stepsPending, pendingSet, MigrationStepGitCodeFacts)
	}

	// Process OpenCode DB step
	if input.RecomputeDB {
		if input.DBAttempted && input.DBSessionFound {
			stepsCompleted = appendMigrationStep(stepsCompleted, completedSet, MigrationStepOpenCodeDB)
			delete(pendingSet, MigrationStepOpenCodeDB)
		} else {
			// Session was not found in local database
			if !completedSet[MigrationStepOpenCodeDB] {
				stepsPending = appendMigrationStep(stepsPending, pendingSet, MigrationStepOpenCodeDB)
			}
			if input.HasExistingDBFacts {
				notes = append(notes, "Local SQLite session not found; preserved existing invocations and telemetry for subsequent host migration")
			} else {
				notes = append(notes, "Local SQLite session not found; pending database resolution on author host")
			}
		}
	}

	// Remove completed steps without moving the other recorded pending steps.
	stepsPending = slices.DeleteFunc(stepsPending, func(step string) bool { return !pendingSet[step] })
	if len(stepsPending) == 0 {
		stepsPending = nil
	}

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

func appendMigrationStep(steps []string, recorded map[string]bool, step string) []string {
	step = strings.TrimSpace(step)
	if step == "" || recorded[step] {
		return steps
	}
	recorded[step] = true
	return append(steps, step)
}
