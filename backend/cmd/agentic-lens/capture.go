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
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/adapters/opencode"
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
	if len(args) == 0 {
		return errors.New("usage: capture install|install-check|uninstall-check|status|diagnostics|inspect-state|verify|check|receive|request|bind|reconcile|relocate-state|disable|uninstall|protocol")
	}
	command := args[0]
	if command == "relocate-state" {
		return relocateCaptureCommand(args[1:], output)
	}
	if command == "inspect-state" {
		return inspectCaptureStateCommand(args[1:], output)
	}
	if command == "protocol" {
		_, err := fmt.Fprintln(output, capture.HandlerProtocolVersion)
		return err
	}
	flags := flag.NewFlagSet("capture "+command, flag.ContinueOnError)
	flags.SetOutput(output)
	source := flags.String("source", "", "Harness: codex or opencode")
	sourceStore := flags.String("store", "", "Exact source store (Codex home or OpenCode database)")
	state := flags.String("state-dir", "", "Capture state override (configured telemetry.capture.state_dir is required)")
	repo := flags.String("repo", ".", "Repository root for additive local hook installation")
	binary := flags.String("binary", "", "Selected native release binary (defaults to current executable; workers retain this executable)")
	wrapper := flags.String("wrapper", "", "Owned common version-pinned wrapper for installed integrations and agent binding feedback")
	sessionID := flags.String("session-id", "", "Exact harness session identity")
	ticketTitle := flags.String("ticket-title", "", "Explicit human-readable ticket title for automatic publication")
	ticketType := flags.String("ticket-type", "", "Explicit conventional ticket type")
	ticket := flags.String("ticket-id", "", "Explicit ticket binding or reconciliation target")
	request := flags.String("request", "", "Exact binding request token delivered by hook feedback")
	dryRun := flags.Bool("dry-run", false, "Show owned installation changes without writing")
	byteBudget := flags.Int64("max-source-bytes", capture.DefaultByteBudget, "Source or diagnostic payload bytes per bounded pass")
	sessionBudget := flags.Int("max-sessions", capture.DefaultSessionBudget, "Session work budget per reconciliation or inspection pass")
	beforeDiagnostic := flags.Int64("before-diagnostic", 0, "Read retained diagnostic history before this sequence (status/diagnostics only)")
	timeout := flags.Duration("timeout", 2*time.Second, "Bound handler or registry work duration")
	if err := flags.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
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
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	project, err := adaptersconfig.ProjectRoot(*repo)
	if err != nil {
		return err
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
	// A valid explicit local disable is authoritative before inspecting state
	// or installation health. Diagnostics for other commands remain visible.
	if command == "check" && !policy.CommitWarning {
		return nil
	}
	if command != "check" && policy.DeprecatedDetailedReport {
		fmt.Fprintln(os.Stderr, "telemetry.capture.detailed_report is deprecated and has no effect; remove it and use capture diagnostics. Publication writes ledger and summary only.")
	}
	if _, legacyErr := os.Stat(filepath.Join(policy.StateDir, "registry.db")); legacyErr == nil && filepath.Base(filepath.Dir(policy.StateDir)) != "projects" {
		fmt.Fprintln(os.Stderr, "Legacy unscoped capture registry preserved; source-verified explicit recovery required. New registrations use project-scoped state.")
	}
	provided := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	if !provided["max-sessions"] {
		*sessionBudget = policy.MaxSessions
	}
	if !provided["max-source-bytes"] {
		*byteBudget = policy.MaxSourceBytes
	}
	absState, err := capture.ProjectState(policy.StateDir, project)
	if err != nil {
		return err
	}
	if !inspection && !repair {
		if err := checkCaptureInstallation(project, absState, *state != ""); err != nil {
			return err
		}
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
		result, e := capture.Install(ctx, capture.InstallOptions{Repository: project, StateDir: location.CoordinationDirectory, Binary: *binary, Wrapper: *wrapper, Source: *source, SourceStore: *sourceStore, DryRun: *dryRun, Remove: command != "install"})
		if e != nil {
			return e
		}
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
	if command == "worker" {
		return captureWorker(s, project, s.CoordinationDirectory(), policy)
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
				if e := capture.LaunchWorker(native, project, absState); e != nil {
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
			native, e := os.Executable()
			if e != nil {
				return e
			}
			response, e := capture.ReceiveCodex(ctx, s, value, absoluteStore, native, absState, *wrapper)
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
		identity.SessionID = value.SessionID
		registration, e := s.Register(ctx, identity, "", value.Event)
		if e != nil {
			var accepted *capture.AcceptedRegistrationError
			handoffAccepted = errors.As(e, &accepted)
			return e
		}
		handoffAccepted = true
		registration, e = s.ProveRegistration(ctx, registration)
		if e != nil {
			return e
		}
		registration, e = s.Change(ctx, identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
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
		return json.NewEncoder(output).Encode(registration)
	case "request":
		r, e := s.Register(ctx, identity, "", "agent_context_request")
		if e != nil {
			return e
		}
		handoffAccepted = true
		feedback, e := captureSessionFailureFeedback(ctx, s, r.Identity, policy, absState)
		if e != nil {
			return fmt.Errorf("durable request accepted but failure feedback unavailable; inspect capture diagnostics: %w", e)
		}
		return json.NewEncoder(output).Encode(captureAgentResponse{CaptureRegistration: r, CaptureFeedback: feedback})
	case "bind":
		if *ticketType != "" {
			if _, ok := domain.NormalizeTicketType(*ticketType); !ok {
				return errors.New("invalid ticket type")
			}
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
					normalized, _ := domain.NormalizeTicketType(*ticketType)
					current.TicketType = string(normalized)
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
	if err != nil && !errors.Is(err, adaptersconfig.ErrConfigNotFound) {
		return "", err
	}
	if source == "codex" {
		home, e := os.UserHomeDir()
		if e != nil {
			return "", e
		}
		configured := ""
		if err == nil {
			configured = adaptersconfig.ProjectSourcePath(project, cfg.Telemetry.Sources.CodexHome)
		}
		return adaptersconfig.ResolveCodexHome("", configured, os.Getenv("CODEX_HOME"), filepath.Join(home, ".codex")), nil
	}
	if source == "opencode" {
		if err == nil {
			return adaptersconfig.ResolveOpenCodePath("", adaptersconfig.ProjectSourcePath(project, cfg.Telemetry.Sources.OpenCodeDatabasePath), adaptersconfig.ProjectSourcePath(project, cfg.Database.Path)), nil
		}
		return opencode.DefaultDBPath()
	}
	return "", errors.New("install requires --source codex or opencode")
}

func reconcileCaptureGeneration(project, dir, ticket string) (*capture.ReconcileResult, error) {
	if _, err := os.Stat(filepath.Join(capture.RegistryPathDirectory(dir), "registry.db")); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	s, err := capture.OpenProject(dir, project)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return nil, err
	}
	result, err := s.Reconcile(ctx, ticket, git.GetGitUser(), policy.MaxSessions, policy.MaxSourceBytes)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
func recordGenerationPublication(project, directory, ticket, outcome, destination, summary string, owners []domain.SessionFact, detail string) error {
	s, err := capture.OpenProject(directory, project)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return errors.Join(s.RecordPublication(ctx, ticket, outcome, destination, summary, owners, detail), s.Close())
}
func hasBoundRegistrations(result *capture.ReconcileResult) bool {
	if result == nil {
		return false
	}
	for _, r := range result.Registrations {
		if strings.TrimSpace(r.TicketID) == result.TicketID {
			return true
		}
	}
	return false
}

// mergeCaptureRows preserves legitimate multiplicity while making repeated snapshots idempotent.
// The v5 skill/task rows have no owner ID, so partial refresh cannot safely delete cached rows.
func mergeCaptureRows[T domain.SkillFact | domain.TaskLinkFact](cached, fresh []T) ([]T, error) {
	result := append([]T{}, cached...)
	counts := map[string]int{}
	for _, row := range cached {
		b, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		counts[string(b)]++
	}
	seen := map[string]int{}
	for _, row := range fresh {
		b, err := json.Marshal(row)
		if err != nil {
			return nil, err
		}
		key := string(b)
		seen[key]++
		if seen[key] > counts[key] {
			result = append(result, row)
		}
	}
	return result, nil
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
func validateCaptureSelection(project, dir, ticket string, source domain.SourceHarness, store string, selected []domain.SessionFact) error {
	var err error
	for _, owner := range selected {
		if err = capture.ValidateOwner(context.Background(), project, domain.CaptureRegistration{Identity: domain.CaptureIdentity{Source: source, Store: store, SessionID: owner.ID}}); err != nil {
			return fmt.Errorf("selected source owner %s: %w", owner.ID, err)
		}
	}

	s, err := capture.OpenProject(dir, project)
	if err != nil {
		return err
	}
	defer s.Close()
	absoluteStore, err := capture.CanonicalStore(store)
	if err != nil {
		return err
	}
	owners, err := s.List(context.Background())
	if err != nil {
		return err
	}
	selectedIDs := map[string]bool{}
	for _, session := range selected {
		selectedIDs[session.ID] = true
	}
	for _, owner := range owners {
		if owner.Identity.Source != source || !selectedIDs[owner.Identity.SessionID] || owner.TicketID == "" || owner.TicketID == ticket {
			continue
		}
		sameStore, compareErr := capture.SameStore(owner.Identity.Store, absoluteStore)
		if compareErr != nil {
			return compareErr
		}
		if sameStore {
			return fmt.Errorf("exact session %s is bound to ticket %s, refusing selection for %s", owner.Identity.SessionID, owner.TicketID, ticket)
		}
	}
	for _, owner := range selected {
		id := domain.CaptureIdentity{Source: source, Store: absoluteStore, SessionID: owner.ID}
		_, e := s.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			if r.TicketID != "" && r.TicketID != ticket {
				return r, errors.New("selected owner has conflicting ticket binding")
			}
			r.TicketID = ticket
			r.ProjectDirectory = project
			r.ProjectProof = domain.CaptureProjectVerified
			if r.Binding == "" {
				r.Binding = "explicit_manual_selection"
			}
			if source == domain.SourceCodex && r.TranscriptPath == "" {
				path, err := codex.DiscoverExactRolloutPath(absoluteStore, owner.ID)
				if err != nil {
					return r, err
				}
				r.TranscriptPath = path
			}
			return r, nil
		})
		if e != nil {
			return e
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
	target, err := capture.ProjectDirectory(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".agenticlens-ledger-")
	if err != nil {
		return err
	}
	temp := file.Name()
	if err = file.Close(); err != nil {
		return err
	}
	defer func() {
		if removeErr := os.Remove(temp); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "Generation staging cleanup:", removeErr)
		}
	}()
	if err = adapters.SaveFactLedgerPreservingHistory(temp, ledger, history); err != nil {
		return err
	}
	file, err = os.OpenFile(temp, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(temp, target)
}

// filterCapturePrecedence preserves the canonical source checkpoint when an
// independently lagging registry merely rereads its own older accepted source.
// Equal/unknown checkpoints do not order differing snapshots.
func filterCapturePrecedence(baseline []domain.SessionFact, result *capture.ReconcileResult) error {
	if result == nil {
		return nil
	}
	existing := map[string]domain.SessionFact{}
	for _, owner := range baseline {
		existing[owner.ID] = owner
	}
	accepted := []domain.SessionFact{}
	rejected := map[string]bool{}
	for _, owner := range result.RefreshedSessions {
		previous, ok := existing[owner.ID]
		if !ok {
			accepted = append(accepted, owner)
			continue
		}
		previousTime, previousKnown := domain.CaptureOwnerCheckpoint(previous)
		nextTime, nextKnown := domain.CaptureOwnerCheckpoint(owner)
		equal, e := sameCaptureOwner(previous, owner)
		if e != nil {
			return e
		}
		upgrade, e := domain.CaptureAuxiliaryUpgrade(previous, owner)
		if e != nil {
			return e
		}
		if equal || upgrade || (previousKnown && nextKnown && nextTime.After(previousTime)) {
			accepted = append(accepted, owner)
			continue
		}
		rejected[owner.ID] = true
	}
	result.RefreshedSessions = accepted
	if len(rejected) > 0 {
		// v5 skill/task rows have no session ID. Without per-owner precedence their
		// replacement authority is unknown; keep the accepted canonical baseline.
		result.RefreshedSkills = nil
		result.RefreshedTaskLinks = nil
		for i := range result.Registrations {
			if rejected[result.Registrations[i].Identity.SessionID] {
				result.Registrations[i].State = domain.CapturePending
				result.Registrations[i].Detail = "Canonical owner has a newer or unordered source checkpoint; accepted canonical facts preserved"
			}
		}
		result.Limits = append(result.Limits, "Some successful source rereads did not establish precedence over canonical owner checkpoints; preserved canonical facts and left coverage pending.")
	}
	return nil
}
func sameCaptureOwner(a, b domain.SessionFact) (bool, error) {
	return domain.CaptureOwnerFactsEqual(a, b)
}

// Explicit selection and registered refresh share the same source-evidence
// precedence. A selected path alone does not prove that differing equal/unknown
// checkpoints are newer than accepted canonical source facts.
func rejectOlderExplicitOwners(baseline, fresh []domain.SessionFact) ([]domain.SessionFact, error) {
	previous := map[string]domain.SessionFact{}
	for _, owner := range baseline {
		previous[owner.ID] = owner
	}
	accepted := []domain.SessionFact{}
	for _, owner := range fresh {
		old, ok := previous[owner.ID]
		if ok {
			equal, err := sameCaptureOwner(old, owner)
			if err != nil {
				return nil, err
			}
			oldTime, oldKnown := domain.CaptureOwnerCheckpoint(old)
			nextTime, nextKnown := domain.CaptureOwnerCheckpoint(owner)
			upgrade, err := domain.CaptureAuxiliaryUpgrade(old, owner)
			if err != nil {
				return nil, err
			}
			if !equal && !upgrade && !(oldKnown && nextKnown && nextTime.After(oldTime)) {
				fmt.Fprintf(os.Stderr, "Capture owner %s source checkpoint does not establish newer facts; preserved canonical owner and left refresh unresolved\n", owner.ID)
				continue
			}
		}
		accepted = append(accepted, owner)
	}
	return accepted, nil
}

func checkCaptureInstallation(project, selected string, explicit bool) error {
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	configuredLocation, err := capture.PreviewStateLocation(policy.StateDir, project)
	if err != nil {
		return err
	}
	configured := configuredLocation.CoordinationDirectory
	selectedLocation, err := capture.PreviewStateLocation(selected, project)
	if err != nil {
		return err
	}
	selected = selectedLocation.CoordinationDirectory
	if err = capture.CheckInstalledState(project, configured, explicit); err != nil {
		return err
	}
	if selected != configured {
		return capture.CheckInstalledState(project, selected, explicit)
	}
	return nil
}
