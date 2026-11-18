package main

import (
	"errors"
	"os"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
)

// A protected worker keeps its authority across its entire lifetime. Any new
// automatic worker honors the existing operator admission for this worktree.
func captureWorkerHome(project, admitted string) (string, error) {
	if admitted != "" {
		return admitted, nil
	}
	return adaptersconfig.ProtectedCodexProjectHome(project)
}

func checkProtectedCaptureState(project, home, state string) error {
	paths, err := adaptersconfig.BindingStatePaths(project)
	if err != nil {
		return err
	}
	return adaptersconfig.CheckProtectedCodexPaths(project, home, append(paths, state)...)
}

func launchCaptureWorker(binary, project, state, admitted string) error {
	return launchSelectedCaptureWorker(project, state, admitted, func() (string, error) { return binary, nil })
}

func launchSelectedCaptureWorker(project, state, admitted string, ordinary func() (string, error)) error {
	home, err := captureWorkerHome(project, admitted)
	if err != nil {
		return err
	}
	if home == "" {
		err := adaptersconfig.WithOrdinaryCapture(project, func() error {
			binary, err := ordinary()
			if err != nil {
				return err
			}
			return capture.LaunchWorker(binary, project, state)
		})
		if !errors.Is(err, adaptersconfig.ErrProtectedCodexProject) {
			return err
		}
		home, err = adaptersconfig.ProtectedCodexProjectHome(project)
		if err != nil {
			return err
		}
		if home == "" {
			return errors.New("protected worker admission changed during launch; retry capture reconcile")
		}
	}
	if err := checkProtectedCaptureState(project, home, state); err != nil {
		return err
	}
	return capture.LaunchProtectedWorker(project, state, home)
}

// First admission holds configuration admission (at the caller) and checks
// existing work before writing the private marker under exclusive routing.
func admitProtectedCapture(project, home, state string) (result error) {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	if _, _, err := runningProtectedEngine(project, home, binary); err != nil {
		return err
	}
	if err := checkProtectedCaptureState(project, home, state); err != nil {
		return err
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return err
	}
	if cfg.Telemetry.Sources.CodexHome == "" {
		return errors.New("protected admission requires the configured Codex source")
	}
	same, err := capture.SameStore(cfg.Telemetry.Sources.CodexHome, home)
	if err != nil || !same {
		return errors.New("configured Codex source differs from the native runtime home")
	}
	store, err := capture.OpenReadOnlyProject(state, project)
	if errors.Is(err, os.ErrNotExist) {
		return adaptersconfig.AdmitProtectedCodexProject(project, home)
	}
	if err != nil {
		return err
	}
	defer func() { result = errors.Join(result, store.Close()) }()
	if err := checkProtectedCaptureState(project, home, store.CoordinationDirectory()); err != nil {
		return err
	}
	lease, err := capture.AcquireWorker(store.CoordinationDirectory())
	if err != nil {
		return err
	}
	if lease == nil {
		return errors.New("an ordinary capture worker is active; retry Codex admission after it drains")
	}
	defer func() { result = errors.Join(result, lease.Close()) }()
	// The new owner must inherit every pending generation. Lease acquisition,
	// rather than checkpoint age or publication success, proves prior release.
	return adaptersconfig.AdmitProtectedCodexProject(project, home)
}
