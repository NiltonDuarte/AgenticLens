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
// If the provided registry is nil, a default registry pre-loaded with StepV1ToV2, StepV2ToV3Sanitized, and StepV3ToV4UserIterations is used.
func NewEngine(registry *Registry) *Engine {
	if registry == nil {
		registry = NewRegistry()
		_ = registry.Register(NewStepV1ToV2())
		_ = registry.Register(NewStepV2ToV3Sanitized())
		_ = registry.Register(NewStepV3ToV4UserIterations())
	}
	return &Engine{
		registry:  registry,
		validator: NewInvariantValidator(),
	}
}

type versionProbe struct {
	SchemaVersion *int `json:"schema_version"`
}

type legacyFactCollections struct {
	Invocations []domain.InvocationFact `json:"invocations"`
	Tools       []domain.ToolFact       `json:"tools"`
	Events      []domain.EventFact      `json:"events"`
}

func decodeLedger(input []byte) (domain.AgenticFactLedger, int, error) {
	version, err := (&Engine{}).DetectVersion(input)
	if err != nil {
		return domain.AgenticFactLedger{}, 0, err
	}
	var ledger domain.AgenticFactLedger
	var rootFields map[string]json.RawMessage
	if err := json.Unmarshal(input, &rootFields); err != nil {
		return domain.AgenticFactLedger{}, 0, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	_, hasSessions := rootFields["sessions"]
	for _, key := range []string{"invocations", "tools", "events"} {
		if _, hasLegacy := rootFields[key]; hasSessions && hasLegacy {
			return domain.AgenticFactLedger{}, 0, fmt.Errorf("mixed ledger shape: canonical sessions cannot coexist with root-level legacy %s; remove ambiguity before migration", key)
		}
	}
	if err := json.Unmarshal(input, &ledger); err != nil {
		return domain.AgenticFactLedger{}, 0, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	var legacy legacyFactCollections
	if err := json.Unmarshal(input, &legacy); err != nil {
		return domain.AgenticFactLedger{}, 0, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
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
		sessions, err := domain.SessionsFromFacts(legacy.Invocations, attributed, legacy.Events)
		if err != nil {
			return domain.AgenticFactLedger{}, 0, fmt.Errorf("normalizing legacy fact collections: %w", err)
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
	if version >= domain.CurrentSchemaVersion && hasSessions {
		normalized, err := domain.PersistAnomalySummaries(ledger)
		if err != nil {
			return domain.AgenticFactLedger{}, 0, fmt.Errorf("sanitizing canonical session events: %w", err)
		}
		ledger = normalized
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
	if targetVersion < 1 {
		return domain.AgenticFactLedger{}, fmt.Errorf("%w: target version %d must be at least 1",
			ErrUnsupportedVersion, targetVersion)
	}

	ledger, sourceVersion, err := decodeLedger(input)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	if err := validateSessions(ledger.Sessions); err != nil {
		return domain.AgenticFactLedger{}, err
	}

	// Idempotency: if already at target version, return ledger unchanged with zero mutations.
	if sourceVersion == targetVersion {
		return ledger, nil
	}

	return e.applyChain(ledger, sourceVersion, targetVersion)
}

// MigrateLedger applies migration steps to an already deserialized domain ledger.
func (e *Engine) MigrateLedger(ledger domain.AgenticFactLedger, targetVersion int) (domain.AgenticFactLedger, error) {
	if targetVersion < 1 {
		return domain.AgenticFactLedger{}, fmt.Errorf("%w: target version %d must be at least 1",
			ErrUnsupportedVersion, targetVersion)
	}

	sourceVersion := 1
	if ledger.SchemaVersion != nil && *ledger.SchemaVersion > 1 {
		sourceVersion = *ledger.SchemaVersion
	}
	if err := validateSessions(ledger.Sessions); err != nil {
		return domain.AgenticFactLedger{}, err
	}

	if sourceVersion == targetVersion {
		return ledger, nil
	}

	return e.applyChain(ledger, sourceVersion, targetVersion)
}

func validateSessions(sessions []domain.SessionFact) error {
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
		for j, inv := range session.Invocations {
			if strings.TrimSpace(inv.ID) == "" {
				return fmt.Errorf("sessions[%d].invocations[%d].id is required", i, j)
			}
			if inv.ParentID != nil && !seen[strings.TrimSpace(*inv.ParentID)] {
				return fmt.Errorf("sessions[%d].invocations[%d].parent_id %q does not identify a session", i, j, *inv.ParentID)
			}
		}
	}
	return nil
}

func (e *Engine) applyChain(ledger domain.AgenticFactLedger, sourceVersion int, targetVersion int) (domain.AgenticFactLedger, error) {
	chain, err := e.registry.PlanChain(sourceVersion, targetVersion)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}

	current := ledger
	for _, step := range chain {
		before := current
		after, err := step.Apply(current)
		if err != nil {
			return domain.AgenticFactLedger{}, fmt.Errorf("applying step %s (%d -> %d): %w",
				step.Description(), step.SourceVersion(), step.TargetVersion(), err)
		}

		if err := e.validator.Validate(before, after, step.TargetVersion()); err != nil {
			return domain.AgenticFactLedger{}, fmt.Errorf("step %s (%d -> %d) failed invariant check: %w",
				step.Description(), step.SourceVersion(), step.TargetVersion(), err)
		}

		current = after
	}

	return current, nil
}
