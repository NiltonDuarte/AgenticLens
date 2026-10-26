package migration

import (
	"agenticlens/internal/domain"
)

// StepV1ToV2 upgrades unversioned or v1 ledgers to schema version 2.
// It initializes schema_version = 2, establishes baseline migration metadata,
// and ensures a non-nil empty event slice if events were absent.
type StepV1ToV2 struct{}

// NewStepV1ToV2 creates a StepV1ToV2 instance.
func NewStepV1ToV2() *StepV1ToV2 {
	return &StepV1ToV2{}
}

// SourceVersion returns the initial schema version (1).
func (s *StepV1ToV2) SourceVersion() int {
	return 1
}

// TargetVersion returns the upgraded schema version (2).
func (s *StepV1ToV2) TargetVersion() int {
	return 2
}

// Description returns a human-readable explanation of this migration transition.
func (s *StepV1ToV2) Description() string {
	return "Upgrades unversioned or v1 ledgers to schema version 2 with events array and migration metadata"
}

// Apply executes the pure migration from schema v1 to schema v2.
func (s *StepV1ToV2) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	cloned := ledger
	targetVer := 2
	cloned.SchemaVersion = &targetVer
	if cloned.Sessions == nil {
		cloned.Sessions = []domain.SessionFact{}
	}

	for i := range cloned.Sessions {
		if cloned.Sessions[i].Events == nil {
			cloned.Sessions[i].Events = []domain.SessionEventFact{}
		}
	}

	if cloned.Migration == nil {
		cloned.Migration = &domain.MigrationInfo{
			SchemaVersion:  2,
			Status:         domain.MigrationStatusCompleted,
			StepsCompleted: []string{"schema_v2_baseline"},
			StepsPending:   []string{},
			LastMigratedAt: cloned.UpdatedAt,
			LastMigratedBy: "universal-migration-engine",
			Notes:          "Baseline migration from v1 to v2",
		}
	} else {
		mig := *cloned.Migration
		mig.SchemaVersion = 2
		hasBaseline := false
		for _, step := range mig.StepsCompleted {
			if step == "schema_v2_baseline" {
				hasBaseline = true
				break
			}
		}
		if !hasBaseline {
			mig.StepsCompleted = append(mig.StepsCompleted, "schema_v2_baseline")
		}
		if mig.LastMigratedBy == "" {
			mig.LastMigratedBy = "universal-migration-engine"
		}
		if mig.LastMigratedAt == "" {
			mig.LastMigratedAt = cloned.UpdatedAt
		}
		cloned.Migration = &mig
	}

	return cloned, nil
}
