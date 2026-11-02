package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	if err := captureCommand(args, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Capture:", err)
		os.Exit(1)
	}
}
func captureCommand(args []string, input io.Reader, output io.Writer) (returnErr error) {
	if len(args) == 0 {
		return errors.New("usage: capture install|status|receive|request|bind|reconcile|disable|uninstall|protocol")
	}
	command := args[0]
	if command == "protocol" {
		_, err := fmt.Fprintln(output, capture.ProtocolVersion)
		return err
	}
	flags := flag.NewFlagSet("capture "+command, flag.ContinueOnError)
	flags.SetOutput(output)
	source := flags.String("source", "", "Harness: codex or opencode")
	sourceStore := flags.String("store", "", "Exact source store (Codex home or OpenCode database)")
	state := flags.String("state-dir", "", "Capture state base (configuration default when omitted)")
	repo := flags.String("repo", ".", "Repository root for additive local hook installation")
	binary := flags.String("binary", "", "Native release binary for installed hooks (defaults to current executable)")
	sessionID := flags.String("session-id", "", "Exact harness session identity")
	ticketTitle := flags.String("ticket-title", "", "Explicit human-readable ticket title for automatic publication")
	ticketType := flags.String("ticket-type", "", "Explicit conventional ticket type")
	ticket := flags.String("ticket-id", "", "Explicit ticket binding or reconciliation target")
	request := flags.String("request", "", "Exact binding request token delivered by hook feedback")
	dryRun := flags.Bool("dry-run", false, "Show owned installation changes without writing")
	byteBudget := flags.Int64("max-source-bytes", capture.DefaultByteBudget, "Source bytes per reconciliation pass")
	sessionBudget := flags.Int("max-sessions", capture.DefaultSessionBudget, "Session work budget per reconciliation pass")
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
	if *timeout <= 0 || *timeout > 5*time.Minute {
		return errors.New("capture timeout must be positive and at most five minutes")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	project, err := adaptersconfig.ProjectRoot(*repo)
	if err != nil {
		return err
	}
	policy, err := adaptersconfig.ResolveCapture(project, *state)
	if err != nil {
		return err
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
	if err := checkCaptureInstallation(project, absState, *state != ""); err != nil {
		return err
	}
	if *source != "" && *source != "codex" && *source != "opencode" {
		return errors.New("capture source must be codex or opencode")
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
		result, e := capture.Install(ctx, capture.InstallOptions{Repository: project, StateDir: absState, Binary: *binary, Source: *source, SourceStore: *sourceStore, DryRun: *dryRun, Remove: command != "install"})
		if e != nil {
			return e
		}
		return json.NewEncoder(output).Encode(result)
	}
	if command == "status" {
		// Status is read-only when no capture state exists.
		if _, e := os.Stat(filepath.Join(absState, "registry.db")); errors.Is(e, os.ErrNotExist) {
			return json.NewEncoder(output).Encode(struct {
				State  string `json:"state"`
				Detail string `json:"detail"`
			}{"degraded", "No lifecycle registrations. Install hooks, review harness trust, then verify an actual agent binding response."})
		}
	}
	s, err := capture.OpenProject(absState, project)
	if err != nil {
		return err
	}
	defer func() {
		if e := s.Close(); e != nil {
			fmt.Fprintln(os.Stderr, "capture registry close:", e)
		}
	}()
	if command == "worker" {
		return captureWorker(s, project, absState, policy)
	}
	if command == "receive" || command == "bind" || command == "request" || command == "reconcile" {
		defer func() {
			if returnErr == nil {
				enabled, e := automaticCaptureEnabled(project, absState)
				if e != nil {
					returnErr = e
					return
				}
				if !enabled {
					return
				}
				if e := s.RequestPublication(context.Background()); e != nil {
					returnErr = fmt.Errorf("publication request: %w", e)
					return
				}
				if e := capture.LaunchWorker(*binaryOrSelf(binary), project, absState); e != nil {
					returnErr = fmt.Errorf("worker launch: %w", e)
				}
			}
		}()
	}
	if command == "status" {
		registrations, e := s.List(ctx)
		if e != nil {
			return e
		}
		manifests := []capture.Manifest{}
		for _, h := range []string{"codex", "opencode"} {
			m, e := capture.ReadManifest(absState, h)
			if e == nil {
				manifests = append(manifests, m)
			} else if !errors.Is(e, os.ErrNotExist) {
				return e
			}
		}
		progress, e := capture.WorkerStatus(absState)
		if e != nil {
			return e
		}
		publication, e := s.PublicationRequest(ctx)
		if e != nil {
			return e
		}
		return json.NewEncoder(output).Encode(struct {
			Publication   capture.PublicationRequest   `json:"publication"`
			Worker        capture.WorkerProgress       `json:"worker"`
			Registrations []domain.CaptureRegistration `json:"registrations"`
			Installation  []capture.Manifest           `json:"installation"`
			Trust         string                       `json:"trust"`
		}{publication, progress, registrations, manifests, "Trust and current enablement are harness-controlled; installation alone is unverified. Review /hooks for Codex; verify OpenCode plugin loading and the agent response."})
	}
	if command == "reconcile" {
		if *ticket == "" {
			return errors.New("reconcile requires --ticket-id")
		}
		result, e := s.Reconcile(ctx, *ticket, git.GetGitUser(), *sessionBudget, *byteBudget)
		if e != nil {
			return e
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
			response, e := capture.ReceiveCodex(ctx, s, value, absoluteStore, *binaryOrSelf(binary), absState)
			if e != nil {
				return e
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
		return json.NewEncoder(output).Encode(r)
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
func binaryOrSelf(value *string) *string {
	if *value == "" {
		if path, e := os.Executable(); e == nil {
			*value = path
		}
	}
	return value
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
	if _, err := os.Stat(filepath.Join(dir, "registry.db")); errors.Is(err, os.ErrNotExist) {
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
func publishCaptureReport(path string, result capture.ReconcileResult) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	digest := sha256.Sum256(bytes)
	body, err := capture.ReportJSON(result, hex.EncodeToString(digest[:]))
	if err != nil {
		return err
	}
	report, err := capture.ProjectDirectory(filepath.Join(filepath.Dir(path), "capture_report.json"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(report), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(report), ".capture-report-")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "capture report cleanup:", e)
		}
	}()
	if _, err = temp.Write(body); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err = temp.Sync(); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, report)
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
	if err = adapters.SaveFactLedger(temp, ledger); err != nil {
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
		if equal || (previousKnown && nextKnown && nextTime.After(previousTime)) {
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
	// User identity is separate attribution metadata; compare whole compact source
	// facts rather than selecting fields by numerical maximum.
	a.User = ""
	b.User = ""
	a.Events = nil
	b.Events = nil
	if a.Source == "" {
		a.Source = b.Source
	}
	if b.Source == "" {
		b.Source = a.Source
	}
	a.Invocations = append([]domain.InvocationFact{}, a.Invocations...)
	b.Invocations = append([]domain.InvocationFact{}, b.Invocations...)
	a.Tools = append([]domain.SessionToolFact{}, a.Tools...)
	b.Tools = append([]domain.SessionToolFact{}, b.Tools...)
	a.Anomalies = append([]domain.Anomaly{}, a.Anomalies...)
	b.Anomalies = append([]domain.Anomaly{}, b.Anomalies...)
	first, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	second, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return string(first) == string(second), nil
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
			if !equal && !(oldKnown && nextKnown && nextTime.After(oldTime)) {
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
	configured, err := capture.ProjectState(policy.StateDir, project)
	if err != nil {
		return err
	}
	if err = capture.CheckInstalledState(project, configured, explicit); err != nil {
		return err
	}
	if selected != configured {
		return capture.CheckInstalledState(project, selected, explicit)
	}
	return nil
}
