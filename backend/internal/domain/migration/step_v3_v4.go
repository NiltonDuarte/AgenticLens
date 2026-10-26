package migration

import (
	"strings"

	"agenticlens/internal/domain"
)

// StepV3ToV4UserIterations upgrades schema version 3 ledgers to schema version 4.
// It populates UserIterations on root invocations from code.sdlc_iterations if unpopulated,
// and updates migration metadata.
type StepV3ToV4UserIterations struct{}

// NewStepV3ToV4UserIterations creates a StepV3ToV4UserIterations instance.
func NewStepV3ToV4UserIterations() *StepV3ToV4UserIterations {
	return &StepV3ToV4UserIterations{}
}

// SourceVersion returns the initial schema version (3).
func (s *StepV3ToV4UserIterations) SourceVersion() int {
	return 3
}

// TargetVersion returns the upgraded schema version (4).
func (s *StepV3ToV4UserIterations) TargetVersion() int {
	return 4
}

// Description returns a human-readable explanation of this migration transition.
func (s *StepV3ToV4UserIterations) Description() string {
	return "Upgrades schema version 3 ledgers to schema version 4 by supporting atomic user_iterations on invocations"
}

// Apply executes the pure migration from schema v3 to schema v4.
func (s *StepV3ToV4UserIterations) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	// Reapply the current persistence allowlist at the schema boundary. Some v3
	// ledgers predate the event payload sanitizer and may still contain error text.
	// Use the error-returning canonicalizer so required evidence without a stable
	// source ID fails migration with a record-level diagnostic instead of being
	// silently dropped by the legacy slice-only pruning helper.
	cloned, err := domain.PersistAnomalySummaries(ledger)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	targetVer := 4
	cloned.SchemaVersion = &targetVer

	for i := range cloned.Sessions {
		if strings.TrimSpace(cloned.Sessions[i].Source) == "" {
			cloned.Sessions[i].Source = domain.SourceOpenCode
		}
	}

	// If root invocations have 0 user_iterations and ledger has Code.SDLCIterations,
	// assign Code.SDLCIterations to the primary root invocation.
	var totalIterations int64
	for _, inv := range cloned.InvocationFacts() {
		totalIterations += inv.UserIterations
	}
	if totalIterations == 0 && cloned.Code.SDLCIterations > 0 {
		for i := range cloned.Sessions {
			if cloned.Sessions[i].ParentID == nil {
				for j := range cloned.Sessions[i].Invocations {
					if cloned.Sessions[i].Invocations[j].ParentID == nil {
						cloned.Sessions[i].Invocations[j].UserIterations = cloned.Code.SDLCIterations
						break
					}
				}
				break
			}
		}
	}

	if cloned.Migration == nil {
		cloned.Migration = &domain.MigrationInfo{
			SchemaVersion:  4,
			Status:         domain.MigrationStatusCompleted,
			StepsCompleted: []string{domain.MigrationStepUserIterations},
			StepsPending:   []string{},
			LastMigratedAt: cloned.UpdatedAt,
			LastMigratedBy: "universal-migration-engine",
			Notes:          "User iterations migration from v3 to v4",
		}
	} else {
		mig := *cloned.Migration
		mig.SchemaVersion = 4

		hasStep := false
		for _, step := range mig.StepsCompleted {
			if step == domain.MigrationStepUserIterations {
				hasStep = true
				break
			}
		}
		if !hasStep {
			mig.StepsCompleted = append(mig.StepsCompleted, domain.MigrationStepUserIterations)
		}

		if len(mig.StepsPending) == 0 {
			mig.Status = domain.MigrationStatusCompleted
		} else {
			mig.Status = domain.MigrationStatusPartial
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
