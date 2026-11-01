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
func captureCommand(args []string, input io.Reader, output io.Writer) error {
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
	state := flags.String("state-dir", capture.DefaultDirectory, "Local ignored capture state directory")
	repo := flags.String("repo", ".", "Repository root for additive local hook installation")
	binary := flags.String("binary", "", "Native release binary for installed hooks (defaults to current executable)")
	sessionID := flags.String("session-id", "", "Exact harness session identity")
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
	absState, err := filepath.Abs(*state)
	if err != nil {
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
			*sourceStore, err = captureSourceStore(*source)
			if err != nil {
				return err
			}
		}
		result, e := capture.Install(ctx, capture.InstallOptions{Repository: *repo, StateDir: absState, Binary: *binary, Source: *source, SourceStore: *sourceStore, DryRun: *dryRun, Remove: command != "install"})
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
	s, err := capture.Open(absState)
	if err != nil {
		return err
	}
	defer func() {
		if e := s.Close(); e != nil {
			fmt.Fprintln(os.Stderr, "capture registry close:", e)
		}
	}()
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
		return json.NewEncoder(output).Encode(struct {
			Registrations []domain.CaptureRegistration `json:"registrations"`
			Installation  []capture.Manifest           `json:"installation"`
			Trust         string                       `json:"trust"`
		}{registrations, manifests, "Trust and current enablement are harness-controlled; installation alone is unverified. Review /hooks for Codex; verify OpenCode plugin loading and the agent response."})
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
		return json.NewEncoder(output).Encode(registration)
	case "request":
		r, e := s.Register(ctx, identity, "", "agent_context_request")
		if e != nil {
			return e
		}
		return json.NewEncoder(output).Encode(r)
	case "bind":
		r, e := s.Bind(ctx, identity, *ticket, *request)
		if e != nil {
			return e
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
func captureSourceStore(source string) (string, error) {
	cfg, err := adaptersconfig.Load("")
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
			configured = cfg.Telemetry.Sources.CodexHome
		}
		return adaptersconfig.ResolveCodexHome("", configured, os.Getenv("CODEX_HOME"), filepath.Join(home, ".codex")), nil
	}
	if source == "opencode" {
		if err == nil {
			return adaptersconfig.ResolveOpenCodePath("", cfg.Telemetry.Sources.OpenCodeDatabasePath, cfg.Database.Path), nil
		}
		return opencode.DefaultDBPath()
	}
	return "", errors.New("install requires --source codex or opencode")
}

func reconcileCaptureGeneration(dir, ticket string) (*capture.ReconcileResult, error) {
	if _, err := os.Stat(filepath.Join(dir, "registry.db")); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	s, err := capture.Open(dir)
	if err != nil {
		return nil, err
	}
	defer s.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	result, err := s.Reconcile(ctx, ticket, git.GetGitUser(), capture.DefaultSessionBudget, capture.DefaultByteBudget)
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
	report := filepath.Join(filepath.Dir(path), "capture_report.json")
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
func validateCaptureSelection(dir, ticket string, source domain.SourceHarness, store string, selected []domain.SessionFact) error {
	if _, err := os.Stat(filepath.Join(dir, "registry.db")); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	s, err := capture.Open(dir)
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
	return nil
}

// saveGenerationLedger keeps the accepted baseline readable if a publisher is
// interrupted during serialization/write. The outer publication lock covers
// this rename and subsequent digest-bound report/summary updates.
func saveGenerationLedger(path string, ledger domain.AgenticFactLedger) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	} else if !errors.Is(err, os.ErrNotExist) {
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
