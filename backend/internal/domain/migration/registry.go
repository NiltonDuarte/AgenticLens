package migration

import (
	"fmt"
	"sync"
)

// Registry manages registered MigrationStep implementations and plans sequential migration chains.
type Registry struct {
	mu    sync.RWMutex
	steps map[int]MigrationStep
}

// NewRegistry initializes an empty thread-safe migration step registry.
func NewRegistry() *Registry {
	return &Registry{
		steps: make(map[int]MigrationStep),
	}
}

// Register registers a discrete version migration step in the registry.
// Each step must advance the schema version by exactly one increment (TargetVersion == SourceVersion + 1).
func (r *Registry) Register(step MigrationStep) error {
	if step == nil {
		return fmt.Errorf("%w: nil migration step provided", ErrUnsupportedVersion)
	}

	if step.TargetVersion() != step.SourceVersion()+1 {
		return fmt.Errorf("%w: step target version %d must be exactly source version %d + 1",
			ErrUnsupportedVersion, step.TargetVersion(), step.SourceVersion())
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.steps[step.SourceVersion()]; exists {
		return fmt.Errorf("%w: source version %d already registered", ErrDuplicateStep, step.SourceVersion())
	}

	r.steps[step.SourceVersion()] = step
	return nil
}

// GetStep returns the registered step starting at the given source version, if present.
func (r *Registry) GetStep(sourceVersion int) (MigrationStep, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	step, ok := r.steps[sourceVersion]
	return step, ok
}

// PlanChain resolves a sequential, contiguous chain of migration steps from sourceVersion to targetVersion.
func (r *Registry) PlanChain(sourceVersion int, targetVersion int) ([]MigrationStep, error) {
	if sourceVersion == targetVersion {
		return []MigrationStep{}, nil
	}

	if sourceVersion > targetVersion {
		return nil, fmt.Errorf("%w: cannot migrate from version %d down to %d",
			ErrDowngradeNotAllowed, sourceVersion, targetVersion)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	chain := make([]MigrationStep, 0, targetVersion-sourceVersion)
	for v := sourceVersion; v < targetVersion; v++ {
		step, exists := r.steps[v]
		if !exists {
			return nil, fmt.Errorf("%w: no migration step found for transition from version %d to %d",
				ErrMissingStep, v, v+1)
		}
		chain = append(chain, step)
	}

	return chain, nil
}
