package migration

import (
	"agenticlens/internal/domain"
)

// StepV2ToV3Sanitized upgrades schema version 2 ledgers to schema version 3.
// It prunes routine events, strips error messages from retained events, and updates migration metadata.
type StepV2ToV3Sanitized struct{}

// NewStepV2ToV3Sanitized creates a StepV2ToV3Sanitized instance.
func NewStepV2ToV3Sanitized() *StepV2ToV3Sanitized {
	return &StepV2ToV3Sanitized{}
}

// SourceVersion returns the initial schema version (2).
func (s *StepV2ToV3Sanitized) SourceVersion() int {
	return 2
}

// TargetVersion returns the upgraded schema version (3).
func (s *StepV2ToV3Sanitized) TargetVersion() int {
	return 3
}

// Description returns a human-readable explanation of this migration transition.
func (s *StepV2ToV3Sanitized) Description() string {
	return "Upgrades schema version 2 ledgers to schema version 3 by pruning routine telemetry events and stripping error message strings"
}

// Apply executes the pure migration from schema v2 to schema v3.
func (s *StepV2ToV3Sanitized) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	cloned := ledger
	targetVer := 3
	cloned.SchemaVersion = &targetVer

	origSourceByID := make(map[string]domain.SourceHarness, len(ledger.Sessions))
	for _, orig := range ledger.Sessions {
		origSourceByID[orig.ID] = orig.Source
	}

	events := legacyV3Events(ledger)
	sessions, err := domain.SessionsFromFacts(ledger.InvocationFacts(), ledger.ToolFacts(), events)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	for i := range sessions {
		if src, ok := origSourceByID[sessions[i].ID]; ok && src != "" {
			sessions[i].Source = src
		}
	}
	cloned.Sessions = sessions

	if cloned.Migration == nil {
		cloned.Migration = &domain.MigrationInfo{
			SchemaVersion:  3,
			Status:         domain.MigrationStatusCompleted,
			StepsCompleted: []string{domain.MigrationStepSanitizedEvents},
			StepsPending:   []string{},
			LastMigratedAt: cloned.UpdatedAt,
			LastMigratedBy: "universal-migration-engine",
			Notes:          "Sanitized event telemetry migration from v2 to v3",
		}
	} else {
		mig := *cloned.Migration
		mig.SchemaVersion = 3

		hasStep := false
		for _, step := range mig.StepsCompleted {
			if step == domain.MigrationStepSanitizedEvents {
				hasStep = true
				break
			}
		}
		if !hasStep {
			mig.StepsCompleted = append(mig.StepsCompleted, domain.MigrationStepSanitizedEvents)
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

func legacyV3Events(ledger domain.AgenticFactLedger) []domain.EventFact {
	events := ledger.EventFacts()
	type anomalyTurn struct {
		session string
		turn    int64
	}
	turns := make(map[anomalyTurn]bool)
	wildcardTurns := make(map[int64]bool)
	for _, anomaly := range domain.DetectAnomalies(ledger) {
		if anomaly.Context.Turn == nil {
			continue
		}
		if anomaly.Context.SessionID == nil {
			wildcardTurns[*anomaly.Context.Turn] = true
			continue
		}
		turns[anomalyTurn{session: *anomaly.Context.SessionID, turn: *anomaly.Context.Turn}] = true
	}
	retained := make([]domain.EventFact, 0)
	for _, event := range events {
		if event.Status != "error" && event.Status != "failed" && event.EventType != domain.EventTypeCompaction && !wildcardTurns[event.Turn] && !turns[anomalyTurn{session: event.SessionID, turn: event.Turn}] {
			continue
		}
		event.ErrorMessage, event.PromptContent, event.ReasoningContent = nil, nil, nil
		event.ToolInput, event.ToolOutput = nil, nil
		retained = append(retained, event)
	}
	return domain.SortEvents(retained)
}
