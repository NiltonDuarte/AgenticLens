package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"agenticlens/internal/domain"
)

// Engine coordinates schema version detection, chain resolution, step execution,
// and invariant validation for telemetry ledgers.
type Engine struct {
	registry  *Registry
	validator *InvariantValidator
}

// NewEngine creates an Engine using the provided registry and standard invariant validator.
// If the provided registry is nil, a default registry pre-loaded with StepV1ToV2, StepV2ToV3Sanitized, StepV3ToV4UserIterations, and StepV4ToV5CostEvidence is used.
func NewEngine(registry *Registry) *Engine {
	if registry == nil {
		registry = NewRegistry()
		for _, step := range []MigrationStep{
			NewStepV1ToV2(),
			NewStepV2ToV3Sanitized(),
			NewStepV3ToV4UserIterations(),
			NewStepV4ToV5CostEvidence(),
			NewStepV5ToV6LeanOccurrences(),
		} {
			if err := registry.Register(step); err != nil {
				// Built-in steps are static; failure indicates a programming error.
				panic(fmt.Errorf("registering built-in migration %d -> %d: %w", step.SourceVersion(), step.TargetVersion(), err))
			}
		}
	}
	return &Engine{
		registry:  registry,
		validator: NewInvariantValidator(),
	}
}

type versionProbe struct {
	SchemaVersion *int `json:"schema_version"`
}

func decodeLedger(input []byte) (legacyLedger, int, error) {
	version, err := (&Engine{}).DetectVersion(input)
	if err != nil {
		return legacyLedger{}, 0, err
	}
	ledger, legacy, hasSessions, err := decodeVersionedLedger(input, version)
	if err != nil {
		return legacyLedger{}, 0, err
	}
	if ledger.Sessions == nil && (len(legacy.Invocations) > 0 || len(legacy.Events) > 0 || len(legacy.Tools) > 0) {
		attributed := make([]domain.ToolFact, 0, len(legacy.Tools))
		dropped := 0
		for _, tool := range legacy.Tools {
			if strings.TrimSpace(tool.SessionID) == "" {
				dropped++
				continue
			}
			attributed = append(attributed, tool)
		}
		owned := make([]legacySessionFact, 0, len(legacy.Invocations))
		for _, row := range legacy.Invocations {
			owned = append(owned, legacySessionFact{
				ID: row.SessionID, ParentID: row.ParentID, User: row.User,
				Invocations: []domain.InvocationFact{row.InvocationFact},
			})
		}
		sessions, err := legacySessionsFromFacts(owned, attributed, legacy.Events)
		if err != nil {
			return legacyLedger{}, 0, fmt.Errorf("normalizing legacy fact collections: %w", err)
		}
		if legacy.Source != "" {
			for i := range sessions {
				sessions[i].Source = legacy.Source
			}
		}
		ledger.Sessions = sessions
		if dropped > 0 {
			note := fmt.Sprintf("Removed %d unscoped legacy tool aggregate row(s) during schema v4 migration; their original session attribution is unavailable. Re-extract from source logs to recover them.", dropped)
			if ledger.Migration == nil {
				ledger.Migration = &domain.MigrationInfo{SchemaVersion: version, Status: domain.MigrationStatusCompleted, StepsCompleted: []string{}, StepsPending: []string{}, Notes: note}
			} else {
				mig := *ledger.Migration
				if mig.Notes != "" {
					mig.Notes += "; "
				}
				mig.Notes += note
				if len(mig.StepsPending) == 0 {
					mig.Status = domain.MigrationStatusCompleted
				} else {
					mig.Status = domain.MigrationStatusPartial
				}
				ledger.Migration = &mig
			}
		}
	}
	if version >= 4 && hasSessions {
		normalized, err := legacyPersistAnomalySummaries(ledger)
		if err != nil {
			return legacyLedger{}, 0, fmt.Errorf("sanitizing canonical session events: %w", err)
		}
		ledger = normalized
	}
	for i := range ledger.Sessions {
		if strings.TrimSpace(string(ledger.Sessions[i].Source)) == "" {
			ledger.Sessions[i].Source = domain.SourceOpenCode
		}
		if strings.TrimSpace(ledger.Sessions[i].User) == "" {
			ledger.Sessions[i].User = "unknown"
		}
	}
	return ledger, version, nil
}

// DetectVersion inspects raw JSON bytes to determine the source schema version.
// Missing, null, or zero/one schema_version values are classified as version 1.
func (e *Engine) DetectVersion(input []byte) (int, error) {
	if len(bytes.TrimSpace(input)) == 0 {
		return 0, fmt.Errorf("%w: empty input payload", ErrMalformedJSON)
	}

	var probe versionProbe
	if err := json.Unmarshal(input, &probe); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}

	if probe.SchemaVersion == nil || *probe.SchemaVersion <= 1 {
		return 1, nil
	}

	return *probe.SchemaVersion, nil
}

// Migrate inspects raw JSON bytes, unmarshals the ledger, plans the sequential upgrade chain,
// and executes transformations while verifying non-destructive domain invariants at each step.
func (e *Engine) Migrate(input []byte, targetVersion int) (domain.AgenticFactLedger, error) {
	if targetVersion < 1 || targetVersion > domain.CurrentSchemaVersion {
		return domain.AgenticFactLedger{}, fmt.Errorf("%w: target version %d is outside the supported range",
			ErrUnsupportedVersion, targetVersion)
	}

	sourceVersion, err := e.DetectVersion(input)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	if sourceVersion > domain.CurrentSchemaVersion {
		return domain.AgenticFactLedger{}, fmt.Errorf("%w: source schema %d is newer than current schema %d", ErrUnsupportedVersion, sourceVersion, domain.CurrentSchemaVersion)
	}
	if sourceVersion == domain.CurrentSchemaVersion {
		if targetVersion != domain.CurrentSchemaVersion {
			return domain.AgenticFactLedger{}, ErrDowngradeNotAllowed
		}
		normalized, err := NormalizeLedgerDurationJSON(input)
		if err != nil {
			return domain.AgenticFactLedger{}, fmt.Errorf("normalizing current duration input: %w", err)
		}
		var current domain.AgenticFactLedger
		decoder := json.NewDecoder(bytes.NewReader(normalized))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&current); err != nil {
			return domain.AgenticFactLedger{}, fmt.Errorf("decoding current schema: %w", err)
		}
		return current, nil
	}
	result, err := e.migrateInput(input, targetVersion)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	return result.current(), nil
}

// MigrateInput exposes the migration-only intermediate model for explicit older
// transitions. Runtime consumers use Migrate and the current canonical output.
func (e *Engine) MigrateInput(input []byte, targetVersion int) (LedgerInput, error) {
	if targetVersion >= domain.CurrentSchemaVersion {
		return legacyLedger{}, fmt.Errorf("%w: use Migrate for the current canonical schema", ErrUnsupportedVersion)
	}
	return e.migrateInput(input, targetVersion)
}
func (e *Engine) migrateInput(input []byte, targetVersion int) (legacyLedger, error) {
	if targetVersion < 1 || targetVersion > domain.CurrentSchemaVersion {
		return legacyLedger{}, ErrUnsupportedVersion
	}
	ledger, sourceVersion, err := decodeLedger(input)
	if err != nil {
		return legacyLedger{}, err
	}
	if err := validateSessions(ledger.Sessions); err != nil {
		return legacyLedger{}, err
	}
	if sourceVersion == targetVersion {
		return ledger, nil
	}
	return e.applyChain(ledger, sourceVersion, targetVersion)
}

func validateSessions(sessions []legacySessionFact) error {
	seen := make(map[string]bool, len(sessions))
	for i, session := range sessions {
		id := strings.TrimSpace(session.ID)
		if id == "" {
			return fmt.Errorf("sessions[%d].id is required", i)
		}
		if seen[id] {
			return fmt.Errorf("sessions[%d].id %q is duplicated", i, id)
		}
		seen[id] = true
	}
	for i, session := range sessions {
		if session.ParentID != nil && !seen[strings.TrimSpace(*session.ParentID)] {
			return fmt.Errorf("sessions[%d].parent_id %q does not identify a session", i, *session.ParentID)
		}
	}
	return nil
}

func (e *Engine) applyChain(ledger legacyLedger, sourceVersion int, targetVersion int) (legacyLedger, error) {
	chain, err := e.registry.PlanChain(sourceVersion, targetVersion)
	if err != nil {
		return legacyLedger{}, err
	}

	current := ledger
	for _, step := range chain {
		before := current
		after, err := step.Apply(current)
		if err != nil {
			return legacyLedger{}, fmt.Errorf("applying step %s (%d -> %d): %w",
				step.Description(), step.SourceVersion(), step.TargetVersion(), err)
		}

		if err := e.validator.Validate(before, after, step.TargetVersion()); err != nil {
			return legacyLedger{}, fmt.Errorf("step %s (%d -> %d) failed invariant check: %w",
				step.Description(), step.SourceVersion(), step.TargetVersion(), err)
		}

		current = after
	}

	if sourceVersion < 4 && targetVersion >= 4 {
		migrationInfo := current.Migration
		if migrationInfo == nil {
			migrationInfo = &domain.MigrationInfo{StepsCompleted: []string{}, StepsPending: []string{}}
		}
		completed := append([]string(nil), migrationInfo.StepsCompleted...)
		pending := append([]string(nil), migrationInfo.StepsPending...)
		found := false
		for _, step := range pending {
			if step == domain.MigrationStepSourceBackfill {
				found = true
				break
			}
		}
		if !found {
			pending = append(pending, domain.MigrationStepSourceBackfill)
		}
		migrationInfo.SchemaVersion = targetVersion
		migrationInfo.Status = domain.MigrationStatusPartial
		migrationInfo.StepsCompleted = completed
		migrationInfo.StepsPending = pending
		if migrationInfo.Notes != "" {
			migrationInfo.Notes += "; "
		}
		migrationInfo.Notes += "Schema migration changed ledger shape only; run source-backed backfill before runtime ingestion."
		current.Migration = migrationInfo
	}

	return current, nil
}
