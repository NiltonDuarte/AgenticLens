package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
)

func runCaptureConfigUpdate(output, replacement string, dryRun bool) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("configuration --output must name an existing file")
	}
	path, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	update, err := adaptersconfig.PrepareCaptureConfigUpdate(path, replacement)
	if err != nil {
		return err
	}
	if !update.Changed {
		fmt.Println("Capture configuration already declares telemetry.capture.state_dir; existing file preserved.")
		return nil
	}
	if dryRun {
		fmt.Printf("# Preview: %s\n%s", path, update.Data)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err = adaptersconfig.WriteCaptureConfigUpdateChecked(ctx, path, update, func(change adaptersconfig.CaptureConfigChange) error {
		// Only explicit administration may inspect the historical implicit location
		// when repairing a missing/invalid field. Runtime reads never select it.
		prior := change.PreviousStateDir
		if strings.TrimSpace(prior) == "" {
			prior = ".agenticlens-capture"
		}
		oldBase, err := adaptersconfig.ResolveCaptureStatePath(change.ProjectRoot, prior)
		if err != nil {
			return err
		}
		newBase, err := adaptersconfig.ResolveCaptureStatePath(change.ProjectRoot, change.StateDir)
		if err != nil {
			return err
		}
		if err = capture.CheckStateRelocation(change.ProjectRoot, oldBase, newBase); err != nil {
			return fmt.Errorf("%w; declare the known previous path explicitly with init --update-capture-config --capture-state-dir before relocating; configuration and state preserved", err)
		}
		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("Updated telemetry.capture.state_dir in %s; unrelated settings and explicit local overrides preserved.\n", path)
	return nil
}
