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

	sourceVersion, err := e.DetectVersion(input)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}

	var ledger domain.AgenticFactLedger
	if err := json.Unmarshal(input, &ledger); err != nil {
		return domain.AgenticFactLedger{}, fmt.Errorf("%w: %v", ErrMalformedJSON, err)
	}
	if sourceVersion >= 4 {
		if err := validateToolSessionIDs(ledger.Tools); err != nil {
			return domain.AgenticFactLedger{}, err
		}
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
	if sourceVersion >= 4 {
		if err := validateToolSessionIDs(ledger.Tools); err != nil {
			return domain.AgenticFactLedger{}, err
		}
	}

	if sourceVersion == targetVersion {
		return ledger, nil
	}

	return e.applyChain(ledger, sourceVersion, targetVersion)
}

func validateToolSessionIDs(tools []domain.ToolFact) error {
	for i, tool := range tools {
		if strings.TrimSpace(tool.SessionID) == "" {
			return fmt.Errorf("tools[%d] is missing required session_id; regenerate or re-extract this ledger from source logs", i)
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
