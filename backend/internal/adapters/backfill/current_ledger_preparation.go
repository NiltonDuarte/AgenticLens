package backfill

import (
	"encoding/json"
	"fmt"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

// currentLedgerPreparation validates the original wire input before any decoding
// can normalize it, and retains a metadata-only candidate for explicit cleanup.
type currentLedgerPreparation struct {
	Ledger           domain.AgenticFactLedger
	WithoutMigration []byte
	Completed        bool
}

func prepareCurrentLedger(content []byte) (currentLedgerPreparation, error) {
	// The formatter owns duplicate-member validation, including escaped keys.
	// Run it on the original bytes before a map can collapse conflicting values.
	if _, err := adapters.FormatLedgerJSON(content); err != nil {
		return currentLedgerPreparation{}, fmt.Errorf("validating original ledger JSON: %w", err)
	}
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

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil {
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
	return currentLedgerPreparation{
		Ledger:           ledger,
		WithoutMigration: withoutMigration,
		Completed:        ledger.Migration != nil && ledger.Migration.Status == domain.MigrationStatusCompleted,
	}, nil
}
