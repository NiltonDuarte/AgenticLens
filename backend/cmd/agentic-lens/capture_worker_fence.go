package main

import (
	"errors"
	"fmt"
	"os"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// captureWorkerFence retains the executable and installations that admitted the
// owner. A pass captures its generation before checking this fence; later events
// cannot be acknowledged by a process admitted under an older selection.
type captureWorkerFence struct {
	binary        string
	artifact      os.FileInfo
	pin           string
	installations map[string]string
}

func newCaptureWorkerFence(project, state, binary string, cfg domain.ProjectConfig) (captureWorkerFence, error) {
	physical, err := capture.ProjectDirectory(binary)
	if err != nil {
		return captureWorkerFence{}, err
	}
	info, err := os.Stat(physical)
	if err != nil {
		return captureWorkerFence{}, err
	}
	fence := captureWorkerFence{binary: physical, artifact: info, pin: cfg.AgenticLens.Version}
	if fence.pin != "" {
		entry, err := adaptersconfig.ResolveEntrypoint(project)
		if err != nil {
			return fence, err
		}
		selected, err := entry.NativePath()
		if err != nil {
			return fence, err
		}
		selected, err = capture.ProjectDirectory(selected)
		if err != nil {
			return fence, err
		}
		if entry.Version != domain.GetBuildInfo().Version || selected != physical {
			return fence, errors.New("native worker does not match the exact current project release; pending work retained for the selected native")
		}
	}
	fence.installations, err = activeWorkerInstallations(state)
	return fence, err
}

func activeWorkerInstallations(state string) (map[string]string, error) {
	identities := map[string]string{}
	for _, source := range []string{"codex", "opencode"} {
		identity, _, err := capture.InstallationIdentity(state, source)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if identity != "" {
			identities[source] = identity
		}
	}
	return identities, nil
}

func (f captureWorkerFence) current(project, state string) (bool, error) {
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return false, err
	}
	if cfg.AgenticLens.Version != f.pin {
		return false, nil
	}
	if f.pin != "" {
		entry, err := adaptersconfig.ResolveEntrypoint(project)
		if err != nil {
			return false, err
		}
		selected, err := entry.NativePath()
		if err != nil {
			return false, err
		}
		selected, err = capture.ProjectDirectory(selected)
		if err != nil {
			return false, err
		}
		if selected != f.binary || entry.Version != domain.GetBuildInfo().Version {
			return false, nil
		}
	}
	info, err := os.Stat(f.binary)
	if err != nil {
		return false, fmt.Errorf("worker native artifact unavailable; pending work retained: %w", err)
	}
	if !os.SameFile(info, f.artifact) || info.Size() != f.artifact.Size() || !info.ModTime().Equal(f.artifact.ModTime()) {
		return false, nil
	}
	current, err := activeWorkerInstallations(state)
	if err != nil {
		return false, err
	}
	// Disabling one source preserves a surviving unchanged integration. An added
	// or replaced active integration requires a fresh owner, even at the same pin.
	for source, identity := range current {
		if f.installations[source] != identity {
			return false, nil
		}
	}
	return true, nil
}
