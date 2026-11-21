package backfill

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

// currentLedgerPreparation validates the original wire input before any decoding
// can normalize it, and retains a duration-canonical candidate for explicit cleanup.
type currentLedgerPreparation struct {
	Ledger                domain.AgenticFactLedger
	WithoutMigration      []byte
	Completed             bool
	FormerMetadataRemoved bool
	CleanupCandidate      []byte
}

// retainStoredFacts restores wire facts that survived the source recovery and
// policy evaluation. The domain projection does not keep stored evidence refs.
func (current currentLedgerPreparation) retainStoredFacts(candidate *adapters.AgenticFactLedgerDTO, ledger domain.AgenticFactLedger, refreshedIDs []string, anomaliesEnabled bool) (adapters.FactLedgerHistory, error) {
	var baseline adapters.AgenticFactLedgerDTO
	if err := json.Unmarshal(current.WithoutMigration, &baseline); err != nil {
		return adapters.FactLedgerHistory{}, fmt.Errorf("decoding stored session evidence: %w", err)
	}
	refreshed := make(map[string]bool, len(refreshedIDs))
	for _, id := range refreshedIDs {
		refreshed[id] = true
	}
	stored := make(map[string]adapters.SessionFactDTO, len(baseline.Sessions))
	for _, owner := range baseline.Sessions {
		stored[owner.ID] = owner
	}
	retained := adapters.FactLedgerHistory{
		SessionIDs:       make(map[string]bool, len(stored)),
		IncompleteGroups: make(map[string]bool),
	}
	for _, owner := range candidate.Sessions {
		if owner.ParentID != nil && refreshed[*owner.ParentID] && !refreshed[owner.ID] {
			retained.IncompleteGroups[*owner.ParentID] = true
		}
	}
	refreshTicketAnomalies := anomaliesEnabled && len(ledger.Sessions) > 0 && !ledger.HasExternalParentReference()
	for i, owner := range candidate.Sessions {
		prior, known := stored[owner.ID]
		if refreshed[owner.ID] {
			if known && retained.IncompleteGroups[owner.ID] {
				restoreGroupEvidence(prior.Anomalies, owner.Anomalies)
			}
			continue
		}
		refreshTicketAnomalies = false
		if known && reflect.DeepEqual(prior.ToDomain(), owner.ToDomain()) {
			candidate.Sessions[i] = prior
			retained.SessionIDs[owner.ID] = true
		}
	}
	if !refreshTicketAnomalies && slices.EqualFunc(baseline.Anomalies, candidate.Anomalies, func(a, b adapters.AnomalyDTO) bool {
		return reflect.DeepEqual(a.ToDomain(), b.ToDomain())
	}) {
		candidate.Anomalies = baseline.Anomalies
		retained.TicketAnomalies = true
	}
	return retained, nil
}

// The source policy keeps an existing rework summary when a recorded child is
// unavailable. Restore the wire references only for that exact surviving fact.
func restoreGroupEvidence(previous, anomalies []adapters.AnomalyDTO) {
	matched := make([]bool, len(previous))
	for i, anomaly := range anomalies {
		if anomaly.RuleID != string(domain.RuleSubagentReworkLoop) || len(anomaly.Evidence) != 0 {
			continue
		}
		for j, old := range previous {
			if !matched[j] && old.RuleID == anomaly.RuleID && reflect.DeepEqual(old.ToDomain(), anomaly.ToDomain()) {
				matched[j] = true
				anomalies[i].Evidence = old.Evidence
				break
			}
		}
	}
}

func prepareCurrentLedger(content []byte) (currentLedgerPreparation, error) {
	// The formatter owns duplicate-member validation, including escaped keys.
	// Run it on the original bytes before a map can collapse conflicting values.
	normalized, err := adapters.NormalizeFormerGenerationMetadata(content)
	if err != nil {
		return currentLedgerPreparation{}, err
	}
	// Use normalized bytes for domain decoding and the raw cleanup candidate;
	// neither path may re-emit former policy/reason fields. Duration receives
	// its own canonical output projection after original input validation.
	formerRemoved := !bytes.Equal(content, normalized)
	content = normalized
	ledger, err := migration.NewEngine(nil).Migrate(content, domain.CurrentSchemaVersion)
	if err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("decoding current ledger: %w", err)
	}
	if info := ledger.Migration; info != nil {
		switch info.Status {
		case domain.MigrationStatusCompleted:
			if len(info.StepsPending) > 0 {
				return currentLedgerPreparation{}, fmt.Errorf("inconsistent completed migration with %d pending steps", len(info.StepsPending))
			}
		case domain.MigrationStatusPending, domain.MigrationStatusPartial:
			if len(info.StepsPending) == 0 {
				return currentLedgerPreparation{}, fmt.Errorf("inconsistent migration status %q without pending steps", info.Status)
			}
		default:
			return currentLedgerPreparation{}, fmt.Errorf("unsupported or missing migration status %q", info.Status)
		}
	}

	// Keep the original non-duration wire facts and optional field presence,
	// while making even metadata-cleanup-only writes use canonical duration.
	canonicalDuration, err := migration.NormalizeLedgerDurationJSON(content)
	if err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("normalizing current cleanup duration: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(canonicalDuration, &fields); err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("reading current ledger members: %w", err)
	}
	delete(fields, "migration")
	withoutMigration, err := json.Marshal(fields)
	if err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("preparing current ledger members: %w", err)
	}
	// Pending metadata is permitted for explicit recovery, but raw facts still
	// need the same required arrays/fields as runtime ingestion. Validate a
	// temporary metadata-free copy without changing the original recovery state.
	if err := adapters.ValidateRuntimeFactLedger(withoutMigration); err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("validating raw current ledger facts: %w", err)
	}
	withoutMigration, err = adapters.FormatLedgerJSON(withoutMigration)
	if err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("formatting current cleanup candidate: %w", err)
	}
	completed := ledger.Migration != nil && ledger.Migration.Status == domain.MigrationStatusCompleted
	candidate := canonicalDuration
	if completed {
		candidate = withoutMigration
	}
	candidate, err = adapters.FormatLedgerJSON(candidate)
	if err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("formatting metadata cleanup candidate: %w", err)
	}
	return currentLedgerPreparation{
		FormerMetadataRemoved: formerRemoved,
		CleanupCandidate:      candidate,
		Ledger:                ledger,
		WithoutMigration:      withoutMigration,
		Completed:             ledger.Migration != nil && ledger.Migration.Status == domain.MigrationStatusCompleted,
	}, nil
}
