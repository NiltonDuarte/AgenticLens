package capture

import (
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func openOpenCodeSource(cfg domain.ProjectConfig, path string) (*opencode.OpenCodeDB, error) {
	if err := adaptersconfig.ValidateDatabasePolicy(cfg); err != nil {
		return nil, err
	}
	return opencode.OpenDB(path, cfg.Database.BusyTimeoutMs)
}

// SetSourceBudgets applies already-selected explicit CLI limits to this operation;
// it never writes configuration or changes source/session authority.
func (s *Store) SetSourceBudgets(maxSessions int, maxSourceBytes int64) error {
	if maxSessions < 1 || maxSourceBytes < 1 {
		return fmt.Errorf("capture source budgets must be positive integers")
	}
	policy := s.configuration.Telemetry.Capture
	policy.MaxSessions = maxSessions
	policy.MaxSourceBytes = maxSourceBytes
	settings := make(map[string]bool, len(policy.Settings)+2)
	for key, value := range policy.Settings {
		settings[key] = value
	}
	settings["max_sessions"] = true
	settings["max_source_bytes"] = true
	policy.Settings = settings
	s.configuration.Telemetry.Capture = policy
	return nil
}
func captureDiscoveryBudget(cfg domain.ProjectConfig) (*codex.DiscoveryBudget, error) {
	policy := cfg.Telemetry.Capture
	if err := policy.ValidateBudgets(); err != nil {
		return nil, err
	}
	return &codex.DiscoveryBudget{MaxSessions: policy.MaxSessions, MaxSourceBytes: policy.MaxSourceBytes}, nil
}
func discoverExactCaptureRollout(cfg domain.ProjectConfig, store, id string) (string, error) {
	budget, err := captureDiscoveryBudget(cfg)
	if err != nil {
		return "", err
	}
	return codex.DiscoverExactRolloutPathWithBudget(store, id, budget)
}

// resolveRegisteredCodexSnapshot accepts a unique native source. A missing
// path may be repaired only for an existing bound owner in the same store.
func resolveRegisteredCodexSnapshot(r domain.CaptureRegistration, budget *codex.DiscoveryBudget) (path string, relocating bool, err error) {
	if r.TranscriptPath == "" {
		return "", false, errors.New("exact Codex transcript path not registered; wait for an authoritative hook")
	}
	if err := codex.ValidateRolloutStore(r.Identity.Store, r.TranscriptPath); err != nil {
		return "", false, err
	}
	registered, pathErr := filepath.EvalSymlinks(r.TranscriptPath)
	if pathErr != nil && (!errors.Is(pathErr, os.ErrNotExist) || r.TicketID == "") {
		return "", false, pathErr
	}
	relocating = pathErr != nil
	path, err = codex.DiscoverExactRolloutPathWithBudget(r.Identity.Store, r.Identity.SessionID, budget)
	if err != nil {
		return "", relocating, err
	}
	selected, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", relocating, err
	}
	if err := codex.ValidateRolloutStore(r.Identity.Store, selected); err != nil {
		return "", relocating, err
	}
	if !relocating {
		if registered != selected {
			return "", false, fmt.Errorf("exact Codex owner %s does not match its registered source path", r.Identity.SessionID)
		}
		return r.TranscriptPath, false, nil
	}
	store, err := CanonicalStore(r.Identity.Store)
	if err != nil {
		return "", true, err
	}
	for _, path := range []string{r.TranscriptPath, selected} {
		physical, err := CanonicalStore(path)
		if err != nil {
			return "", true, err
		}
		relative, err := filepath.Rel(store, physical)
		if err != nil || !filepath.IsLocal(relative) || relative == "." {
			return "", true, fmt.Errorf("Codex rollout recovery path escapes its registered store: %s", path)
		}
	}
	return selected, true, nil
}

func (b *workBudget) selectCodexRollout(r domain.CaptureRegistration, cfg domain.ProjectConfig) (string, bool, error) {
	if b.bytes < 1 {
		return "", false, opencode.ErrCaptureBudget
	}
	if b.codexDiscovery == nil {
		budget, err := captureDiscoveryBudget(cfg)
		if err != nil {
			return "", false, err
		}
		b.codexDiscovery = budget
	}
	discovery := b.codexDiscovery
	before := discovery.SourceBytesRead
	// Discovery has one accounting object for the pass. Other source reads
	// also consume the remaining allowance before the next discovery.
	discovery.MaxSourceBytes = before + b.bytes
	path, relocating, err := resolveRegisteredCodexSnapshot(r, discovery)
	b.charge(discovery.SourceBytesRead - before)
	return path, relocating, err
}
