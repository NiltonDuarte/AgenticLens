package main

import (
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
)

// A protected worker keeps its authority across its entire lifetime. Any new
// automatic worker also honors the operator receipt, including OpenCode workers
// whose project has no configured Codex source.
func captureWorkerHome(admitted string) (string, error) {
	if admitted != "" {
		return admitted, nil
	}
	home := adaptersconfig.RuntimeCodexHome()
	installed, err := adaptersconfig.ProtectedCodexInstalled(home)
	if err != nil || !installed {
		return "", err
	}
	return home, nil
}

func checkProtectedCaptureState(project, home, state string) error {
	paths, err := adaptersconfig.BindingStatePaths(project)
	if err != nil {
		return err
	}
	return adaptersconfig.CheckProtectedCodexPaths(project, home, append(paths, state)...)
}

func launchCaptureWorker(binary, project, state, admitted string) error {
	home, err := captureWorkerHome(admitted)
	if err != nil {
		return err
	}
	if home != "" {
		if err := checkProtectedCaptureState(project, home, state); err != nil {
			return err
		}
		return capture.LaunchProtectedWorker(project, state, home)
	}
	return capture.LaunchWorker(binary, project, state)
}
