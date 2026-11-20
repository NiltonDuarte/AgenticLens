package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func captureControlAction(command string) bool { return command == "discard" || command == "restore" }

func inspectCaptureControl(ctx context.Context, project, captureID string, output io.Writer) (returnErr error) {
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	store, err := capture.OpenReadOnlyProject(policy.StateDir, project)
	if err != nil {
		return fmt.Errorf("open configured capture state for exact ID lookup: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	status, err := store.CaptureControl(ctx, captureID)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(status)
}

func runCaptureControl(command string, args []string, output io.Writer, protectedHome, nativeProject, nativeSession string) error {
	flags := flag.NewFlagSet("capture "+command, flag.ContinueOnError)
	flags.SetOutput(output)
	captureID := flags.String("capture-id", "", "Exact capture ID from the verified pending capture")
	project, request := nativeProject, ""
	if nativeProject == "" {
		flags.StringVar(&project, "repo", "", "Repository for the capture's local worktree state")
	} else {
		flags.StringVar(&request, "request", "", "Current native capture control request")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || strings.TrimSpace(project) == "" || strings.TrimSpace(*captureID) == "" {
		if nativeProject != "" {
			return errors.New("native OpenCode capture control requires the exact --capture-id")
		}
		return fmt.Errorf("capture %s requires --repo and an exact --capture-id", command)
	}
	if nativeProject != "" && request == "" {
		return errors.New("native OpenCode capture control requires its current request")
	}
	project, err := adaptersconfig.ProjectRoot(project)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	state, coordination, status, err := persistCaptureControl(ctx, project, protectedHome, nativeSession, *captureID, command, request)
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err == nil {
		err = launchCaptureWorker(binary, project, coordination, protectedHome)
	}
	if err != nil {
		return fmt.Errorf("capture control persisted in %s; worker launch failed and work remains queued: %w", state, err)
	}
	return json.NewEncoder(output).Encode(status)
}

func persistCaptureControl(ctx context.Context, project, protectedHome, nativeSession, captureID, command, request string) (state, coordination string, status capture.CaptureControl, returnErr error) {
	admission, err := adaptersconfig.AcquireCaptureAdmission(ctx, project)
	if err != nil {
		return "", "", status, err
	}
	defer func() { returnErr = errors.Join(returnErr, admission.Close()) }()
	policy, err := admission.ResolveCapture("")
	if err != nil {
		return "", "", status, err
	}
	if !policy.Enabled {
		return "", "", status, errors.New("capture control requires enabled automatic capture to process the operation")
	}
	if err := policy.ValidateWorker(); err != nil {
		return "", "", status, err
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return "", "", status, err
	}
	if _, err := adaptersconfig.ResolvePendingDirectory(cfg, project); err != nil {
		return "", "", status, err
	}
	if err := capture.CheckInstalledInboxIsolation(project, cfg.Telemetry.Sources.CodexHome); err != nil {
		return "", "", status, err
	}
	state = policy.StateDir
	home, err := captureWorkerHome(project, "")
	if err != nil {
		return state, "", status, err
	}
	if protectedHome != "" {
		if home == "" {
			return state, "", status, errors.New("protected capture control requires an admitted project")
		}
		same, err := capture.SameStore(home, protectedHome)
		if err != nil || !same {
			return state, "", status, errors.Join(errors.New("protected Codex home differs from project admission"), err)
		}
	}
	if home != "" {
		if err := verifyProtectedWriter(project, home, state); err != nil {
			return state, "", status, err
		}
	}
	location, err := capture.PreviewStateLocation(state, project)
	if err != nil {
		return state, "", status, err
	}
	if location.Provisional {
		return state, "", status, errors.New("configured capture state is unavailable for this worktree")
	}
	coordination = location.CoordinationDirectory
	closeAdmission, err := acquireCaptureControlWorkerAdmission(coordination)
	if err != nil {
		return state, coordination, status, err
	}
	defer func() { returnErr = errors.Join(returnErr, closeAdmission()) }()
	var native domain.CaptureIdentity
	if nativeSession != "" {
		path, err := nativeOpenCodeStore(ctx, project, cfg)
		if err != nil {
			return state, coordination, status, err
		}
		native = domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: nativeSession}
	}
	store, err := capture.OpenProjectContext(ctx, state, project)
	if err != nil {
		return state, coordination, status, err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	if store.CoordinationDirectory() != coordination {
		return state, coordination, status, errors.New("capture worktree coordination changed; retry the operation")
	}
	if nativeSession != "" {
		status, err = store.ApplyOpenCodeCaptureControl(ctx, native, captureID, command, request)
	} else {
		status, err = store.SetCaptureDiscarded(ctx, captureID, command == "discard", nil)
	}
	return state, coordination, status, err
}
