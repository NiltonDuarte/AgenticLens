package capture

import (
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"context"
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

func (s *Store) verifyBindingSource(ctx context.Context, r domain.CaptureRegistration, budget *codex.DiscoveryBudget) (string, error) {
	if r.Identity.Source != domain.SourceCodex {
		return r.TranscriptPath, validateOwnerSource(ctx, s.project, r, s.configuration, false, budget)
	}
	if r.TranscriptPath != "" {
		// Codex may archive a captured rollout before its first binding. A
		// missing path may be recovered only inside the original source store.
		if _, err := os.Stat(r.TranscriptPath); r.TicketID != "" || !errors.Is(err, os.ErrNotExist) {
			return r.TranscriptPath, validateOwnerSource(ctx, s.project, r, s.configuration, false, budget)
		}
		if err := codex.ValidateRolloutStore(r.Identity.Store, r.TranscriptPath); err != nil {
			return "", err
		}
	}
	if budget == nil {
		var err error
		budget, err = captureDiscoveryBudget(s.configuration)
		if err != nil {
			return "", err
		}
		budget.Context = ctx
	}
	var path string
	var err error
	relocating := r.TranscriptPath != ""
	if relocating {
		path, err = resolveRegisteredCodexSnapshot(r, "", true, budget)
	} else {
		path, err = codex.DiscoverExactRolloutPathWithBudget(r.Identity.Store, r.Identity.SessionID, budget)
	}
	if err != nil {
		return "", err
	}
	metadata, err := readCodexOwnerMetadata(r.Identity.Store, path, s.configuration, false, budget)
	if err != nil {
		return "", err
	}
	if err := validateCodexOwnerMetadata(s.project, r, metadata); err != nil {
		return "", err
	}
	if relocating && !keepsAcceptedCaptureParent(r, metadata.ParentID) {
		return "", errors.New("recovered Codex source conflicts with the accepted parent identity")
	}
	return path, nil
}

// inspectRegisteredCodexPath allows recovery only when a bound owner's cached
// path is missing. Existing paths retain their metadata-first budget preflight.
func inspectRegisteredCodexPath(r domain.CaptureRegistration) (string, bool, error) {
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
	return registered, pathErr != nil, nil
}

func resolveRegisteredCodexSnapshot(r domain.CaptureRegistration, registered string, relocating bool, budget *codex.DiscoveryBudget) (string, error) {
	path, err := codex.DiscoverExactRolloutPathWithBudget(r.Identity.Store, r.Identity.SessionID, budget)
	if err != nil {
		return "", err
	}
	selected, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	if err := codex.ValidateRolloutStore(r.Identity.Store, selected); err != nil {
		return "", err
	}
	if !relocating {
		if registered != selected {
			return "", fmt.Errorf("exact Codex owner %s does not match its registered source path", r.Identity.SessionID)
		}
		return r.TranscriptPath, nil
	}
	store, err := CanonicalStore(r.Identity.Store)
	if err != nil {
		return "", err
	}
	for _, path := range []string{r.TranscriptPath, selected} {
		physical, err := CanonicalStore(path)
		if err != nil {
			return "", err
		}
		relative, err := filepath.Rel(store, physical)
		if err != nil || !filepath.IsLocal(relative) || relative == "." {
			return "", fmt.Errorf("Codex rollout recovery path escapes its registered store: %s", path)
		}
	}
	return selected, nil
}

func (b *workBudget) selectCodexRollout(r domain.CaptureRegistration, cfg domain.ProjectConfig, registered string, relocating bool) (string, error) {
	if b.bytes < 1 {
		return "", opencode.ErrCaptureBudget
	}
	if b.codexDiscovery == nil {
		budget, err := captureDiscoveryBudget(cfg)
		if err != nil {
			return "", err
		}
		b.codexDiscovery = budget
	}
	discovery := b.codexDiscovery
	before := discovery.SourceBytesRead
	// Discovery has one accounting object for the pass. Other source reads
	// also consume the remaining allowance before the next discovery.
	discovery.MaxSourceBytes = before + b.bytes
	path, err := resolveRegisteredCodexSnapshot(r, registered, relocating, discovery)
	b.charge(discovery.SourceBytesRead - before)
	return path, err
}
