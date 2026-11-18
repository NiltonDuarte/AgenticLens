package capture

import (
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"fmt"
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

// validateRegisteredCodexSnapshot checks uniqueness without replacing the hook's
// registered source path. A registration is a binding, not authority to choose
// between multiple native snapshots for the same owner.
func validateRegisteredCodexSnapshot(r domain.CaptureRegistration, budget *codex.DiscoveryBudget) error {
	if err := codex.ValidateRolloutStore(r.Identity.Store, r.TranscriptPath); err != nil {
		return err
	}
	path, err := codex.DiscoverExactRolloutPathWithBudget(r.Identity.Store, r.Identity.SessionID, budget)
	if err != nil {
		return err
	}
	registered, err := filepath.EvalSymlinks(r.TranscriptPath)
	if err != nil {
		return err
	}
	selected, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	if registered != selected {
		return fmt.Errorf("exact Codex owner %s does not match its registered source path", r.Identity.SessionID)
	}
	return nil
}
