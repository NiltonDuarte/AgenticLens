package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func runCapture(args []string) {
	output := io.Writer(os.Stdout)
	if len(args) > 0 && args[0] == "check" {
		output = os.Stderr
	}
	if err := captureCommand(args, os.Stdin, output); err != nil {
		fmt.Fprintln(os.Stderr, "Capture:", err)
		os.Exit(1)
	}
}
func captureCommand(args []string, input io.Reader, output io.Writer) error {
	var err error
	if len(args) > 0 && args[0] == "receive" {
		var buffered bytes.Buffer
		err = captureCommandAction(args, input, &buffered)
		if err != nil {
			feedback := capture.HookOutput{Message: "AgenticLens capture failed. The session may remain unassigned or pending; run capture diagnostics for this exact project, check configured state/source paths and retry capture reconcile. CLI stderr contains the failure."}
			if encodeErr := json.NewEncoder(output).Encode(feedback); encodeErr != nil {
				err = errors.Join(err, encodeErr)
			}
		} else {
			_, err = output.Write(buffered.Bytes())
		}
	} else {
		err = captureCommandAction(args, input, output)
	}
	if len(args) > 0 && args[0] == "check" && err != nil {
		// This optional health observer must never block its owner's Git operation.
		if _, writeErr := fmt.Fprintln(output, "AgenticLens capture advisory: capture health unavailable; run capture diagnostics to inspect configuration/state and refresh setup if needed."); writeErr != nil {
			fmt.Fprintln(os.Stderr, "capture advisory output unavailable:", writeErr)
		}
		return nil
	}
	return err
}
func captureCommandAction(args []string, input io.Reader, output io.Writer) (returnErr error) {
	// Only the native hook and worker dispatches carry this private selection
	// context. Codex supplies its actual runtime home. Mixed-source hooks and
	// workers may instead carry the home recorded for their admitted project.
	home := ""
	if len(args) >= 3 && args[len(args)-2] == "--protected-codex-home" {
		switch args[0] {
		case "user-hook", "opencode-hook", "worker", "reconcile", "discard", "restore", "admit-protected-worktree":
		default:
			return errors.New("protected Codex selection applies only to native hooks and their worker")
		}
		home = args[len(args)-1]
		if !filepath.IsAbs(home) {
			return errors.New("protected Codex home must be absolute")
		}
		if args[0] == "user-hook" || args[0] == "admit-protected-worktree" {
			same, err := capture.SameStore(adaptersconfig.RuntimeCodexHome(), home)
			if err != nil || !same {
				return errors.New("protected Codex home must match the native runtime home")
			}
		}
		args = args[:len(args)-2]
	}
	return captureAction(args, input, output, home)
}

func captureAction(args []string, input io.Reader, output io.Writer, protectedHome string) error {
	return captureNativeAction(args, input, output, protectedHome, false)
}

func captureNativeAction(args []string, input io.Reader, output io.Writer, protectedHome string, nativeOpenCode bool) (returnErr error) {
	if len(args) == 0 {
		return errors.New("usage: capture install-user|uninstall-user|install|install-check|uninstall-check|status|diagnostics|inspect-state|verify|check|receive|request|bind|submit-binding|discard|restore|prepare-publication|reconcile|relocate-state|disable|uninstall|protocol")
	}
	command := args[0]
	if captureControlAction(command) {
		return runCaptureControl(command, args[1:], output, protectedHome, "", "")
	}
	if command == "publish-protected" {
		if len(args) < 5 {
			return errors.New("protected generation requires project, Codex home, admitted state and generation arguments")
		}
		protection := capture.ProtectedPublication{Project: args[1], Home: args[2], State: args[3]}
		runGeneration(args[4:], &protection)
		return nil
	}
	if command == "submit-binding" {
		return captureSubmitBinding(args[1:], output)
	}
	if command == "prepare-publication" {
		return capturePreparePublication(args[1:], output)
	}
	if command == "opencode-hook" {
		return nativeOpenCodeHook(args[1:], input, output, protectedHome)
	}
	if command == "user-hook" {
		return nativeCodexUserHook(input, output, protectedHome)
	}
	if command == "relocate-state" {
		return relocateCaptureCommand(args[1:], output)
	}
	if command == "inspect-state" {
		return inspectCaptureStateCommand(args[1:], output)
	}
	if command == "protected-worker-protocol" {
		_, err := fmt.Fprintln(output, adaptersconfig.ProtectedWorkerProtocol)
		return err
	}
	if command == "protocol" {
		_, err := fmt.Fprintln(output, capture.HandlerProtocolVersion)
		return err
	}
	flags := flag.NewFlagSet("capture "+command, flag.ContinueOnError)
	flags.SetOutput(output)
	source := flags.String("source", "", "Harness: codex or opencode")
	userOperation := command == "install-user" || command == "uninstall-user"
	userSource := ""
	if userOperation {
		flags.StringVar(&userSource, "user-source", "", "User integration to install or remove: codex or opencode")
	}
	sourceStore := flags.String("store", "", "Exact source store (Codex home or OpenCode database)")
	state := flags.String("state-dir", "", "Capture state override (configured telemetry.capture.state_dir is required)")
	repo := flags.String("repo", ".", "Repository root for additive local hook installation")
	binary := flags.String("binary", "", "Selected native release binary (defaults to current executable; workers retain this executable)")
	wrapper := flags.String("wrapper", "", "Installed native manager for integrations and agent binding feedback")
	entrypoint := flags.String("entrypoint", "", "Installed native PATH command (defaults to the user's agentic-lens command)")
	sessionID := flags.String("session-id", "", "Exact harness session identity")
	ticketTitle := flags.String("ticket-title", "", "Explicit human-readable ticket title for automatic publication")
	ticketType := flags.String("ticket-type", "", "Explicit conventional ticket type")
	ticket := flags.String("ticket-id", "", "Explicit ticket binding or reconciliation target")
	request := flags.String("request", "", "Exact binding request token delivered by hook feedback")
	destination := flags.String("destination-directory", "", "Existing publication directory indicated through the native binding tool")
	destinationOnly := flags.Bool("destination-only", false, "Indicate the publication directory for an already bound session")
	userLevel := flags.Bool("user-level", false, "Native user-level handoff; binding has no project wrapper dependency")
	legacyRepository := flags.String("migrate-legacy-repo", "", "Explicit checkout whose unchanged owned project callbacks should migrate to user capture")
	dryRun := flags.Bool("dry-run", false, "Show owned installation changes without writing")
	byteBudget := flags.Int64("max-source-bytes", 0, "Source or diagnostic payload bytes per bounded pass")
	sessionBudget := flags.Int("max-sessions", 0, "Session work budget per reconciliation or inspection pass")
	var captureID string
	if command == "status" {
		flags.StringVar(&captureID, "capture-id", "", "Read exact local discard or restore status for this capture ID")
	}
	beforeDiagnostic := flags.Int64("before-diagnostic", 0, "Read retained diagnostic history before this sequence (status/diagnostics only)")
	timeout := flags.Duration("timeout", 2*time.Second, "Bound handler/registry work or each required installation phase; optional harness inspection has a separate budget")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	provided := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	if provided["destination-directory"] || provided["destination-only"] {
		if command != "bind" || !nativeOpenCode || !*userLevel || *source != "opencode" {
			return errors.New("destination indication requires the native OpenCode tool or Codex submit-binding")
		}
		if strings.TrimSpace(*destination) == "" {
			return errors.New("destination requires a nonblank directory path")
		}
	}
	if provided["user-source"] {
		if userSource != "opencode" && userSource != "codex" {
			return errors.New("--user-source must select codex or opencode")
		}
		if provided["source"] && *source != userSource {
			return errors.New("--user-source and --source must agree")
		}
		*source = userSource
	}
	for _, choice := range []struct{ flag, value string }{{"store", *sourceStore}, {"state-dir", *state}, {"source", *source}, {"repo", *repo}} {
		if provided[choice.flag] && strings.TrimSpace(choice.value) == "" {
			return fmt.Errorf("--%s must be nonblank when supplied", choice.flag)
		}
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected capture arguments")
	}
	if *beforeDiagnostic < 0 || (*beforeDiagnostic > 0 && command != "status" && command != "diagnostics") {
		return errors.New("--before-diagnostic requires status/diagnostics and a nonnegative sequence")
	}
	if *timeout <= 0 || *timeout > 5*time.Minute {
		return errors.New("capture timeout must be positive and at most five minutes")
	}
	callerCtx := context.Background()
	if command == "install" {
		var stop context.CancelFunc
		callerCtx, stop = signal.NotifyContext(callerCtx, os.Interrupt, syscall.SIGTERM)
		defer stop()
	}
	ctx, cancel := context.WithTimeout(callerCtx, *timeout)
	defer cancel()
	if userOperation {
		entry, err := adaptersconfig.ResolveEntrypoint(*repo)
		if err != nil {
			return err
		}
		options := capture.UserInstallOptions{Repository: entry.Project, Entrypoint: *entrypoint, LegacyRepository: *legacyRepository, DryRun: *dryRun, Remove: command == "uninstall-user"}
		var result capture.InstallResult
		switch *source {
		case "", "codex":
			options.Home = strings.TrimSpace(*sourceStore)
			if options.Home == "" {
				options.Home, err = captureSourceStore("codex", *repo)
			}
			if err == nil {
				result, err = capture.InstallUserCodex(ctx, options)
			}
		case "opencode":
			if provided["state-dir"] || provided["store"] {
				return errors.New("OpenCode user installation reads its config directory from the harness environment; set the source database and capture state in project local settings")
			}
			options.Home, err = capture.OpenCodeConfigHome()
			if err == nil {
				result, err = capture.InstallUserOpenCode(ctx, options)
			}
		default:
			return errors.New("capture source must be codex or opencode")
		}
		if err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(result)
	}
	project, err := adaptersconfig.ProjectRoot(*repo)
	if err != nil {
		return err
	}
	if provided["capture-id"] {
		for option := range provided {
			if option != "repo" && option != "capture-id" && option != "timeout" {
				return fmt.Errorf("--%s cannot be combined with exact capture ID status", option)
			}
		}
		return inspectCaptureControl(ctx, project, captureID, output)
	}
	inspection := command == "status" || command == "diagnostics" || command == "verify" || command == "check"
	gitOnly := command == "install-check" || command == "uninstall-check"
	repair := command == "install" || command == "disable" || command == "uninstall" || gitOnly
	if *dryRun && !repair {
		return errors.New("--dry-run only supports installation changes")
	}
	var admission *adaptersconfig.CaptureAdmission
	releaseAdmission := func() error {
		if admission == nil {
			return nil
		}
		err := admission.Close()
		admission = nil
		return err
	}
	if !inspection && !gitOnly && !*dryRun {
		admission, err = adaptersconfig.AcquireCaptureAdmission(ctx, project)
		if err != nil {
			return err
		}
		defer func() { returnErr = errors.Join(returnErr, releaseAdmission()) }()
	}
	var policy domain.CaptureConfig
	if admission != nil {
		policy, err = admission.ResolveCapture(*state)
	} else {
		policy, err = adaptersconfig.ResolveCapture(project, *state)
	}
	if err != nil {
		return err
	}
	if command == "admit-protected-worktree" {
		if !policy.Enabled || protectedHome == "" || len(provided) != 1 || !provided["repo"] {
			return errors.New("protected worktree admission requires enabled capture, its repository and actual Codex home")
		}
		return admitProtectedCapture(project, protectedHome, policy.StateDir)
	}
	if !policy.Enabled && *userLevel && *source == "opencode" {
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	if !policy.Enabled && (command == "receive" || command == "request" || command == "worker") {
		if command == "receive" {
			_, err := io.WriteString(output, "{}\n")
			return err
		}
		return nil
	}
	// A valid explicit local disable is authoritative before inspecting state
	// or installation health. Diagnostics for other commands remain visible.
	if command == "check" {
		if err := policy.RequireSettings("commit_warning"); err != nil {
			return err
		}
		if !policy.CommitWarning {
			return nil
		}
	}
	if provided["max-sessions"] && *sessionBudget < 1 {
		return errors.New("--max-sessions must be a positive integer")
	}
	if provided["max-source-bytes"] && *byteBudget < 1 {
		return errors.New("--max-source-bytes must be a positive integer")
	}
	usesBudgets := inspection || command == "reconcile" || command == "worker" || command == "receive" || command == "request" || command == "bind"
	if usesBudgets {
		if !provided["max-sessions"] {
			if err := policy.RequireSettings("max_sessions"); err != nil {
				return err
			}
			*sessionBudget = policy.MaxSessions
		}
		if !provided["max-source-bytes"] {
			if err := policy.RequireSettings("max_source_bytes"); err != nil {
				return err
			}
			*byteBudget = policy.MaxSourceBytes
		}
		if *sessionBudget < 1 || *byteBudget < 1 {
			return errors.New("capture session/byte budgets must be positive integers")
		}
		policy.MaxSessions = *sessionBudget
		policy.MaxSourceBytes = *byteBudget
		if policy.Settings != nil {
			policy.Settings["max_sessions"] = true
			policy.Settings["max_source_bytes"] = true
		}
	}
	var openCodeParent string
	var openCodePending error
	if *userLevel && *source == "opencode" {
		if command != "receive" && command != "bind" && command != "request" {
			return errors.New("unsupported native OpenCode operation")
		}
		cfg, loadErr := adaptersconfig.Load(project)
		if loadErr != nil {
			return loadErr
		}
		*sourceStore, err = nativeOpenCodeStore(ctx, project, cfg)
		if err != nil {
			return err
		}
		openCodeParent, openCodePending, err = capture.OpenCodeHandoffParent(ctx, cfg, project, *sourceStore, *sessionID)
		if err != nil {
			return err
		}
	}
	admittedHome := protectedHome
	if admittedHome != "" && (command == "worker" || command == "reconcile" || *source == "opencode") {
		recorded, err := adaptersconfig.ProtectedCodexProjectHome(project)
		if err != nil {
			return err
		}
		if recorded == "" {
			return errors.New("protected capture requires an admitted project")
		}
		same, err := capture.SameStore(recorded, admittedHome)
		if err != nil || !same {
			return errors.New("protected Codex home differs from project admission")
		}
	}
	if command == "worker" || *userLevel || (command == "reconcile" && protectedHome != "") {
		protectedHome, err = captureWorkerHome(project, protectedHome)
		if err != nil {
			return err
		}
		if protectedHome != "" {
			if err := checkProtectedCaptureState(project, protectedHome, policy.StateDir); err != nil {
				return err
			}
		}
	}
	if command == "reconcile" && protectedHome != "" {
		if err := verifyProtectedWriter(project, protectedHome, policy.StateDir); err != nil {
			return err
		}
	}
	absState, err := capture.ProjectState(policy.StateDir, project)
	if err != nil {
		return err
	}

	if *source != "" && *source != "codex" && *source != "opencode" {
		return errors.New("capture source must be codex or opencode")
	}
	if command == "install-check" || command == "uninstall-check" {
		result, e := capture.InstallCheck(ctx, capture.CheckInstallOptions{Repository: project, Wrapper: *wrapper, DryRun: *dryRun, Remove: command == "uninstall-check"})
		if e != nil {
			return e
		}
		return json.NewEncoder(output).Encode(result)
	}
	if command == "install" || command == "disable" || command == "uninstall" {
		if *binary == "" {
			*binary, err = os.Executable()
			if err != nil {
				return err
			}
		}
		if *sourceStore == "" {
			*sourceStore, err = captureSourceStore(*source, project)
			if err != nil {
				return err
			}
		}

		location, e := capture.PreviewStateLocation(absState, project)
		if e != nil {
			return e
		}
		if !*dryRun {
			if e := ctx.Err(); e != nil {
				return e
			}
			store, e := capture.OpenProjectContext(ctx, absState, project)
			if e != nil {
				return e
			}
			location.CoordinationDirectory = store.CoordinationDirectory()
			if e = errors.Join(store.Close(), releaseAdmission()); e != nil {
				return e
			}
		}
		provisional := *dryRun && location.Provisional
		if provisional {
			location.CoordinationDirectory = filepath.Join(location.RegistryDirectory, "worktrees", "allocation-pending")
		}
		installCtx := ctx
		if command == "install" {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("required installation setup: %w", err)
			}
			cancel()
			installCtx = callerCtx
		}
		result, e := capture.Install(installCtx, capture.InstallOptions{Repository: project, StateDir: location.CoordinationDirectory, Binary: *binary, Wrapper: *wrapper, Source: *source, SourceStore: *sourceStore, DryRun: *dryRun, Remove: command != "install", RequiredTimeout: *timeout})
		if e != nil {
			return e
		}
		result.StateAllocationPending = provisional
		return json.NewEncoder(output).Encode(result)
	}
	if command == "status" || command == "diagnostics" || command == "verify" || command == "check" {
		inspectionPolicy := policy
		inspectionPolicy.MaxSessions = *sessionBudget
		inspectionPolicy.MaxSourceBytes = *byteBudget
		return inspectCapture(ctx, command, absState, project, inspectionPolicy, domain.CaptureIdentity{Source: domain.SourceHarness(*source), Store: *sourceStore, SessionID: *sessionID}, *beforeDiagnostic, output)
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if command == "receive" || command == "worker" {
		cfg, err := adaptersconfig.Load(project)
		if err != nil {
			return err
		}
		if err := capture.CheckInstalledInboxIsolation(project, cfg.Telemetry.Sources.CodexHome); err != nil {
			return err
		}
	}
	if command == "worker" && admittedHome == "" && protectedHome != "" {
		if err := releaseAdmission(); err != nil {
			return err
		}
		return launchCaptureWorker("", project, absState, protectedHome)
	}
	s, err := capture.OpenProjectContext(ctx, absState, project)
	if err != nil {
		return err
	}
	if err = releaseAdmission(); err != nil {
		return errors.Join(err, s.Close())
	}
	defer func() {
		if e := s.Close(); e != nil {
			fmt.Fprintln(os.Stderr, "capture registry close:", e)
		}
	}()
	if usesBudgets {
		if err := s.SetSourceBudgets(policy.MaxSessions, policy.MaxSourceBytes); err != nil {
			return err
		}
	}
	if command == "worker" {
		return captureWorker(s, project, s.CoordinationDirectory(), policy, protectedHome)
	}
	absState = s.CoordinationDirectory()
	handoffAccepted := false
	if command == "receive" || command == "bind" || command == "request" || command == "reconcile" {
		defer func() {
			if returnErr == nil || handoffAccepted {
				enabled, e := automaticCaptureEnabled(project, absState)
				if e != nil {
					returnErr = errors.Join(returnErr, e)
					return
				}
				if !enabled {
					return
				}
				if e := s.RequestPublication(context.Background()); e != nil {
					returnErr = errors.Join(returnErr, fmt.Errorf("publication request failed; durable handoff retained; verify capture state permissions: %w", e))
					return
				}
				native, nativeErr := os.Executable()
				if nativeErr != nil {
					returnErr = errors.Join(returnErr, fmt.Errorf("selected native executable unavailable; durable work retained: %w", nativeErr))
					return
				}
				if e := launchCaptureWorker(native, project, absState, protectedHome); e != nil {
					diagnosticErr := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "failed", Detail: "Native worker launch failed. Verify the selected executable and capture state permissions; retry capture reconcile."})
					returnErr = errors.Join(returnErr, fmt.Errorf("worker launch failed; durable registration retained; inspect capture diagnostics: %w", e), diagnosticErr)
				}
			}
		}()
	}
	if command == "reconcile" {
		if *ticket == "" {
			return errors.New("reconcile requires --ticket-id")
		}
		indexProgress, e := s.BackfillDiagnosticIndex(ctx, *sessionBudget, *byteBudget)
		if e != nil {
			return fmt.Errorf("bounded capture diagnostic maintenance failed: %w", e)
		}
		result, e := s.Reconcile(ctx, *ticket, git.GetGitUser(), *sessionBudget, *byteBudget)
		if e != nil {
			return e
		}
		result.DiagnosticIndex = &indexProgress
		if !indexProgress.Complete {
			result.Limits = append(result.Limits, "Diagnostic head metadata maintenance is partial; repeat explicit capture reconcile with sufficient max-sessions/max-source-bytes. Current health remains unverified until indexing completes.")
		}
		return json.NewEncoder(output).Encode(result)
	}
	if *sourceStore == "" {
		return errors.New("receive/request/bind require explicit --store")
	}
	absoluteStore, err := capture.CanonicalStore(*sourceStore)
	if err != nil {
		return err
	}
	identity := domain.CaptureIdentity{Source: domain.SourceHarness(*source), Store: absoluteStore, SessionID: *sessionID}
	switch command {
	case "receive":
		value, e := capture.ReadHookContext(ctx, input)
		if e != nil {
			return e
		}
		if *source == "codex" {
			if *userLevel && value.Event == "PostToolUse" {
				// A successful handoff requests publication and wakes the bounded
				// worker. Only the worker verifies and applies the queued response.
				return json.NewEncoder(output).Encode(capture.HookOutput{})
			}
			native, e := os.Executable()
			if e != nil {
				return e
			}
			var response capture.HookOutput
			if *userLevel {
				response, e = capture.ReceiveCodexUser(ctx, s, value, absoluteStore, native, absState)
			} else {
				response, e = capture.ReceiveCodex(ctx, s, value, absoluteStore, native, absState, *wrapper)
			}
			handoffAccepted = response.Registered
			if e != nil {
				return e
			}
			if value.Event == "SessionStart" || value.Event == "UserPromptSubmit" {
				exact := domain.CaptureIdentity{Source: domain.SourceCodex, Store: absoluteStore, SessionID: value.SessionID}
				feedback, e := captureSessionFailureFeedback(ctx, s, exact, policy, absState)
				if e != nil {
					return fmt.Errorf("durable handoff accepted but failure feedback unavailable; inspect capture diagnostics: %w", e)
				}
				if feedback != "" {
					if response.Context == nil {
						response.Context = &capture.HookContext{Event: value.Event, Context: feedback}
					} else {
						response.Context.Context += "\n" + feedback
					}
					if response.Message != "" {
						response.Message += " "
					}
					response.Message += feedback
				}
			}
			return json.NewEncoder(output).Encode(response)
		}
		switch value.Event {
		case "session.created", "session.updated", "session.status", "session.idle", "session.error":
		default:
			return errors.New("unsupported OpenCode lifecycle event")
		}
		if *userLevel && value.SessionID != identity.SessionID {
			return errors.New("OpenCode callback session differs from verified session")
		}
		identity.SessionID = value.SessionID
		registration, e := s.Register(ctx, identity, "", value.Event)
		if e != nil {
			var accepted *capture.AcceptedRegistrationError
			handoffAccepted = errors.As(e, &accepted)
			return e
		}
		handoffAccepted = true
		if !*userLevel {
			if _, e = s.ProveRegistration(ctx, registration); e != nil {
				return e
			}
		}
		registration, e = s.Change(ctx, identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			if *userLevel {
				if openCodePending != nil {
					r.ProjectProof = domain.CaptureProjectUnverified
				} else {
					var err error
					r, err = capture.RecordOpenCodeHandoff(r, project, openCodeParent)
					if err != nil {
						return r, err
					}
				}
			}
			if value.Event == "session.created" {
				r.Activity = "busy"
			}
			if value.Event == "session.idle" || value.Event == "session.error" {
				r.Activity = "idle"
			}
			if value.Activity == "busy" || value.Activity == "idle" || value.Activity == "retry" {
				r.Activity = value.Activity
			}
			return r, nil
		})
		if e != nil {
			return e
		}
		if openCodePending != nil {
			return fmt.Errorf("registered OpenCode session pending exact source metadata: %w", openCodePending)
		}
		return json.NewEncoder(output).Encode(registration)
	case "request":
		r, e := s.Register(ctx, identity, "", "agent_context_request")
		if e != nil {
			return e
		}
		handoffAccepted = true
		if *userLevel && *source == "opencode" {
			if openCodePending != nil {
				return fmt.Errorf("registered OpenCode session pending exact source metadata: %w", openCodePending)
			}
			r, e = s.InheritOpenCodeParent(ctx, r, openCodeParent)
			if e != nil {
				return e
			}
		}
		feedback, e := captureSessionFailureFeedback(ctx, s, r.Identity, policy, absState)
		if e != nil {
			return fmt.Errorf("durable request accepted but failure feedback unavailable; inspect capture diagnostics: %w", e)
		}
		response := captureAgentResponse{CaptureRegistration: r, CaptureFeedback: feedback}
		if nativeOpenCode {
			control, e := s.CaptureControlInvitation(ctx, r)
			if e != nil {
				return e
			}
			return json.NewEncoder(output).Encode(struct {
				captureAgentResponse
				DestinationRequest string                        `json:"destination_request,omitempty"`
				CaptureControl     *capture.NativeCaptureControl `json:"capture_control,omitempty"`
			}{response, capture.NativePublicationDestinationRequest(r), control})
		}
		return json.NewEncoder(output).Encode(response)
	case "bind":
		if openCodePending != nil {
			return fmt.Errorf("OpenCode source metadata pending; binding unresolved: %w", openCodePending)
		}
		if nativeOpenCode && *destinationOnly {
			r, e := s.IndicateNativeOpenCodeDestination(ctx, identity, capture.BindingTicket{ID: *ticket, Title: *ticketTitle, Type: *ticketType, Destination: *destination}, *request)
			if e != nil {
				return e
			}
			return json.NewEncoder(output).Encode(r)
		}
		var normalizedType domain.ConventionalTicketType
		if *ticketType != "" {
			var ok bool
			normalizedType, ok = domain.NormalizeTicketType(*ticketType)
			if !ok {
				return errors.New("invalid ticket type")
			}
		}
		if nativeOpenCode {
			r, e := s.BindNativeOpenCode(ctx, identity, capture.BindingTicket{ID: *ticket, Title: strings.TrimSpace(*ticketTitle), Type: string(normalizedType), Destination: *destination}, *request)
			if e != nil {
				return e
			}
			return json.NewEncoder(output).Encode(r)
		}
		r, e := s.Bind(ctx, identity, *ticket, *request)
		if e != nil {
			return e
		}
		if strings.TrimSpace(*ticketTitle) != "" || *ticketType != "" {
			r, e = s.Change(ctx, identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				if strings.TrimSpace(*ticketTitle) != "" {
					current.TicketTitle = strings.TrimSpace(*ticketTitle)
				}
				if *ticketType != "" {
					current.TicketType = string(normalizedType)
				}
				return current, nil
			})
			if e != nil {
				return e
			}
		}
		return json.NewEncoder(output).Encode(r)
	default:
		return fmt.Errorf("unknown capture command %q", command)
	}
}
func captureSourceStore(source, project string) (string, error) {
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return "", err
	}
	switch source {
	case "codex":
		return adaptersconfig.ResolveCodexHome("", cfg.Telemetry.Sources.CodexHome)
	case "opencode":
		return adaptersconfig.ResolveOpenCodePath("", cfg.Telemetry.Sources.OpenCodeDatabasePath)
	}
	return "", errors.New("install requires --source codex or opencode")
}

func reconcileCaptureGeneration(project, dir, ticket string, generationPolicies ...domain.GenerationPolicy) (*capture.ReconcileResult, error) {
	return reconcileCaptureGenerationAt(project, dir, ticket, time.Now().UTC(), generationPolicies...)
}

func reconcileCaptureGenerationAt(project, dir, ticket string, cutoff time.Time, generationPolicies ...domain.GenerationPolicy) (*capture.ReconcileResult, error) {
	policy, err := adaptersconfig.ResolveGenerationCapture(project, "")
	if err != nil {
		return nil, err
	}
	return reconcileCaptureGenerationAtWithPolicy(project, dir, ticket, cutoff, policy, generationPolicies...)
}

func reconcileCaptureGenerationAtWithPolicy(project, dir, ticket string, cutoff time.Time, policy domain.CaptureConfig, generationPolicies ...domain.GenerationPolicy) (*capture.ReconcileResult, error) {
	if _, err := os.Stat(filepath.Join(capture.RegistryPathDirectory(dir), "registry.db")); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if err := errors.Join(policy.ValidatePass(), policy.ValidateBudgets()); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	s, err := capture.OpenProjectContext(ctx, dir, project)
	if err != nil {
		return nil, fmt.Errorf("opening the capture registry for reconciliation: %w", err)
	}
	defer s.Close()
	if err := s.SetSourceBudgets(policy.MaxSessions, policy.MaxSourceBytes); err != nil {
		return nil, err
	}
	result, err := s.ReconcileAt(ctx, ticket, git.GetGitUser(), policy.MaxSessions, policy.MaxSourceBytes, cutoff, generationPolicies...)
	if err != nil {
		return nil, fmt.Errorf("reconciling registered owners: %w", err)
	}
	return &result, nil
}

// retainCaptureChildLinks keeps accepted Codex delegation child links on the
// owners an automatic capture pass refreshed or restored without re-reading
// their children. The committed ledger is the accepted baseline, so a machine
// without the earlier capture state keeps them as well. Kept and refused links
// count as unresolved child evidence, and the latest reconciliation is
// recorded as partial. A contended registry ends generation with the busy
// exit code, like reconciliation itself. Transcript events and manual
// generation are unchanged.
func retainCaptureChildLinks(project, directory, ticket string, accepted []domain.SessionFact, result *capture.ReconcileResult) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	retention, err := capture.RetainAcceptedChildLinks(ctx, directory, project, result, accepted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot record retained child links; publication not completed: %v\n", err)
		if capture.IsRegistryContention(err) {
			os.Exit(capturePublicationBusyExit)
		}
		os.Exit(1)
	}
	if retention.Kept == 0 && retention.Refused == 0 && retention.Cleared == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "Capture kept %d accepted delegation child link(s) for ticket %s that this pass did not re-read, could not place %d accepted link(s) whose delegation is ambiguous, has no known invocation time or is held by a sibling, and cleared %d contradicted link(s); unresolved child evidence: %d, so the reconciliation remains partial.\n", retention.Kept, ticket, retention.Refused, retention.Cleared, result.PendingChildren)
}

func recordGenerationPublication(project, directory, ticket, outcome, destination, summary string, owners []domain.SessionFact, detail string, refreshed ...[]domain.SessionFact) error {
	s, err := capture.OpenProject(directory, project)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return errors.Join(s.RecordPublication(ctx, ticket, outcome, destination, summary, owners, detail, refreshed...), s.Close())
}

func fillMissingCaptureSessions(existing, fallback []domain.SessionFact) []domain.SessionFact {
	seen := map[string]bool{}
	for _, owner := range existing {
		seen[owner.ID] = true
	}
	result := append([]domain.SessionFact{}, existing...)
	for _, owner := range fallback {
		if !seen[owner.ID] {
			result = append(result, owner)
			seen[owner.ID] = true
		}
	}
	return result
}
func validateManualSelection(project, ticket string, source domain.SourceHarness, store string, selected []domain.SessionFact, baseline []domain.SessionFact, exactPaths map[string]string, cfg domain.ProjectConfig) error {
	trusted := make(map[string]bool)
	for _, owner := range baseline {
		if owner.Source == source {
			trusted[owner.ID] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for _, owner := range selected {
			if owner.ParentID != nil && trusted[*owner.ParentID] && !trusted[owner.ID] {
				trusted[owner.ID] = true
				changed = true
			}
		}
	}
	for _, owner := range selected {
		registration := domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: source, Store: store, SessionID: owner.ID}, TranscriptPath: exactPaths[owner.ID]}
		if trusted[owner.ID] {
			registration.TicketID = ticket
			registration.Binding = "explicit_manual_selection"
			registration.ProjectDirectory = project
			registration.ProjectProof = domain.CaptureProjectVerified
		}
		if err := capture.ValidateManualOwner(context.Background(), project, registration, cfg); err != nil {
			return fmt.Errorf("selected source owner %s: %w", owner.ID, err)
		}
	}

	return nil
}

// saveGenerationLedger keeps the accepted baseline readable if a publisher is
// interrupted during serialization/write. The outer publication lock covers
// this rename and subsequent digest-bound report/summary updates.
func saveGenerationLedger(path string, ledger domain.AgenticFactLedger) error {
	return saveGenerationLedgerPreservingHistory(path, ledger, nil)
}

func saveGenerationLedgerPreservingHistory(path string, ledger domain.AgenticFactLedger, history []adapters.SessionFactDTO) error {
	target := capture.PublicationPath(path)
	if _, err := capture.StatPublicationFile(filepath.Dir(target)); err != nil {
		return fmt.Errorf("workspace moved or unavailable; resolve and retry: %w", err)
	}
	data, err := adapters.PrepareFactLedgerPublication(ledger, history)
	if err != nil {
		return err
	}
	accepted, err := capture.ReadPublicationFile(target)
	if err == nil {
		equal, compareErr := adapters.LedgerFactsEqual(accepted, data)
		if compareErr != nil {
			return compareErr
		}
		if equal {
			return nil
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading accepted ledger: %w", err)
	}

	file, err := capture.CreatePublicationTemp(filepath.Dir(target), ".agenticlens-ledger-")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer func() {
		if removeErr := capture.RemovePublicationFile(temp); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "Generation staging cleanup:", removeErr)
		}
	}()
	if _, err = file.Write(data); err != nil {
		return errors.Join(err, file.Close())
	}
	if err = file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}
	if err = file.Close(); err != nil {
		return err
	}
	return capture.RenamePublicationFile(temp, target)
}

// filterCapturePrecedence validates identity only. RefreshedSessions contains
// successful complete exact-source reads, including equal/older snapshots.
func filterCapturePrecedence(baseline []domain.SessionFact, result *capture.ReconcileResult) error {
	if result == nil {
		return nil
	}
	return capture.ValidateSessionMerge(baseline, result.RefreshedSessions)
}

// checkCaptureInstallation validates configured and explicitly selected state
// against retained installation ownership before automatic capture operations.
func checkCaptureInstallation(project, selected string, explicit bool) error {
	if _, err := adaptersconfig.SelectConfig(project); err == nil {
		return nil
	}
	selectedLocation, err := capture.PreviewStateLocation(selected, project)
	if err != nil {
		return err
	}
	selected = selectedLocation.CoordinationDirectory
	if selectedLocation.Provisional {
		selected = selectedLocation.RegistryDirectory
	}
	policy, err := adaptersconfig.ResolveGenerationCapture(project, "")
	if err != nil {
		return err
	}
	configuredLocation, err := capture.PreviewStateLocation(policy.StateDir, project)
	if err != nil {
		return err
	}
	configured := configuredLocation.CoordinationDirectory
	if configuredLocation.Provisional {
		configured = configuredLocation.RegistryDirectory
	}
	if err = capture.CheckInstalledState(project, configured, explicit); err != nil {
		return err
	}
	if selected != configured {
		return capture.CheckInstalledState(project, selected, explicit)
	}
	return nil
}
