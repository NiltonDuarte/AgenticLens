package main

import (
	"fmt"
	"os"
	"runtime"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// captureWorkerFence retains admitted executable bytes and checks operational
// disablement. A newly preferred build does not retire an active worker.
type captureWorkerFence struct {
	binary   string
	artifact os.FileInfo
	sources  domain.TelemetrySourcePaths
}

func newCaptureWorkerFence(project, state, binary string, cfg domain.ProjectConfig) (captureWorkerFence, error) {
	source := binary
	if runtime.GOOS == "linux" {
		source = "/proc/self/exe"
	}
	physical, err := capture.RetainExecutable(source, state)
	if err != nil {
		return captureWorkerFence{}, err
	}
	identity, err := adaptersconfig.ValidateNativeExecutable(physical)
	if err != nil {
		return captureWorkerFence{}, err
	}
	if identity != domain.GetBuildInfo() {
		return captureWorkerFence{}, fmt.Errorf("retained worker build differs from admitted running build; pending work retained")
	}
	info, err := os.Stat(physical)
	if err != nil {
		return captureWorkerFence{}, err
	}
	return captureWorkerFence{binary: physical, artifact: info, sources: cfg.Telemetry.Sources}, nil
}

func (f captureWorkerFence) current(project, state string) (bool, error) {
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil || !policy.Enabled {
		return false, err
	}

	info, err := os.Stat(f.binary)
	if err != nil {
		return false, fmt.Errorf("worker native artifact unavailable; pending work retained: %w", err)
	}
	if !os.SameFile(info, f.artifact) || info.Size() != f.artifact.Size() || !info.ModTime().Equal(f.artifact.ModTime()) {
		return false, nil
	}
	// Installation and selection identify preferred future invocations, not the
	// admitted owner. Reinstalling B must not redirect A's later children to B.
	// Local source/state changes remain operational boundaries and stop this owner
	// without acknowledging pending work or silently retargeting another checkout.
	configured, err := capture.PreviewStateLocation(policy.StateDir, project)
	if err != nil {
		return false, err
	}
	same, err := capture.SameStore(configured.CoordinationDirectory, state)
	if err != nil {
		return false, err
	}
	if !same {
		return false, fmt.Errorf("configured capture state changed from active %s to %s; pending work retained", state, configured.CoordinationDirectory)
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return false, err
	}
	if cfg.Telemetry.Sources != f.sources {
		return false, fmt.Errorf("configured capture source changed during active worker lifetime; pending work retained in original source/state scope")
	}

	return true, nil
}
