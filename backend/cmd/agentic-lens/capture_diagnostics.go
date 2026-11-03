package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

type captureInspection struct {
	DiagnosticMetadataComplete     bool                         `json:"diagnostic_metadata_complete"`
	DiagnosticsAvailable           bool                         `json:"diagnostics_available"`
	DiagnosticsContinuation        bool                         `json:"diagnostics_continuation"`
	CurrentDiagnosticsContinuation bool                         `json:"current_diagnostics_continuation"`
	DiagnosticsBefore              int64                        `json:"diagnostics_before,omitempty"`
	DiagnosticBytes                int64                        `json:"diagnostic_bytes_read"`
	InstallationIssues             []string                     `json:"installation_issues"`
	Continuation                   bool                         `json:"registrations_continuation"`
	Publication                    capture.PublicationRequest   `json:"publication"`
	Worker                         capture.WorkerProgress       `json:"worker"`
	Registrations                  []domain.CaptureRegistration `json:"registrations"`
	Installation                   []capture.Manifest           `json:"installation"`
	Diagnostics                    []domain.CaptureDiagnostic   `json:"diagnostics"`
	Verification                   domain.CaptureHealth         `json:"verification"`
	Limits                         []string                     `json:"limits"`
}

func inspectCapture(ctx context.Context, command, directory, project string, policy domain.CaptureConfig, caller domain.CaptureIdentity, beforeDiagnostic int64, output io.Writer) (returnErr error) {
	view := captureInspection{Registrations: []domain.CaptureRegistration{}, Installation: []capture.Manifest{}, Diagnostics: []domain.CaptureDiagnostic{}, Limits: []string{"Recorded observations only; no source discovery, extraction, binding or worker launch performed. Harness trust and unobserved sessions are unknown."}}
	if policy.DeprecatedDetailedReport {
		view.Limits = append(view.Limits, "telemetry.capture.detailed_report is deprecated and has no effect. Remove it; publication uses ledger and summary only, with detailed diagnostics kept locally.")
	}
	location, err := capture.PreviewStateLocation(directory, project)
	if err != nil {
		return err
	}
	directory = location.CoordinationDirectory
	issues, err := capture.InstallationIssues(project, directory)
	if err != nil {
		return err
	}
	view.InstallationIssues = issues
	configuredPolicy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	configuredLocation, err := capture.PreviewStateLocation(configuredPolicy.StateDir, project)
	if err != nil {
		return err
	}
	configured := configuredLocation.CoordinationDirectory
	if configured != directory {
		configuredIssues, err := capture.InstallationIssues(project, configured)
		if err != nil {
			return err
		}
		for _, issue := range configuredIssues {
			view.InstallationIssues = append(view.InstallationIssues, "Configured capture state drift: "+issue)
		}
		if len(configuredIssues) == 0 {
			view.Limits = append(view.Limits, "Explicit capture state differs from the configured default; no owned installation drift was established. The selected project namespace remains isolated.")
		}
	}
	store, err := capture.OpenReadOnlyProject(directory, project)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read-only capture inspection failed: %w", err)
	}
	if store != nil {
		defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
		view.Registrations, view.Continuation, err = store.ListBounded(ctx, policy.MaxSessions)
		if err != nil {
			return err
		}
		if policy.MaxSessions > (int(^uint(0)>>1)-8)/10 {
			return errors.New("capture diagnostic session budget is too large")
		}
		tickets := []string{}
		seenTickets := map[string]bool{}
		owners := []capture.DiagnosticOwner{}
		for _, row := range view.Registrations {
			if row.TicketID != "" {
				owners = append(owners, capture.DiagnosticOwner{TicketID: row.TicketID, Identity: row.Identity})
			}
			if row.TicketID != "" && !seenTickets[row.TicketID] {
				tickets = append(tickets, row.TicketID)
				seenTickets[row.TicketID] = true
			}
		}
		current, readErr := store.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Current: true, Tickets: tickets, ExactOwners: owners, Limit: policy.MaxSessions*10 + 8, MaxBytes: policy.MaxSourceBytes})
		if readErr != nil {
			return readErr
		}
		view.DiagnosticsAvailable = current.Available
		view.DiagnosticMetadataComplete = current.MetadataComplete
		if current.Available && !current.MetadataComplete {
			view.Limits = append(view.Limits, "Diagnostic head metadata is incomplete or predates indexed inspection. Run explicit capture reconcile --ticket-id for bounded mutable maintenance; repeat with sufficient row/byte budgets. Read-only inspection did not backfill history and cannot certify current health.")
		}
		if !current.Available {
			view.Limits = append(view.Limits, "Registry predates durable diagnostics; publication/failure history is unavailable, not measured empty. Read-only inspection did not initialize missing diagnostic metadata.")
		}
		view.Diagnostics = current.Records
		view.CurrentDiagnosticsContinuation = current.More
		view.DiagnosticsContinuation = current.More
		view.DiagnosticBytes = current.Bytes
		if command == "status" || command == "diagnostics" {
			remaining := policy.MaxSourceBytes - current.Bytes
			if remaining > 0 {
				history, readErr := store.ReadDiagnostics(ctx, capture.DiagnosticReadOptions{Before: beforeDiagnostic, Limit: policy.MaxSessions, MaxBytes: remaining})
				if readErr != nil {
					return readErr
				}
				view.DiagnosticsContinuation = view.DiagnosticsContinuation || history.More
				view.DiagnosticsBefore = history.NextBefore
				view.DiagnosticBytes += history.Bytes
				sequences := map[int64]bool{}
				for _, item := range view.Diagnostics {
					sequences[item.Sequence] = true
				}
				for _, item := range history.Records {
					if !sequences[item.Sequence] {
						view.Diagnostics = append(view.Diagnostics, item)
						sequences[item.Sequence] = true
					}
				}
			} else {
				view.DiagnosticsContinuation = true
			}
		}
		sort.Slice(view.Diagnostics, func(i, j int) bool { return view.Diagnostics[i].Sequence < view.Diagnostics[j].Sequence })
		view.Publication, err = store.PublicationRequest(ctx)
		if err != nil {
			return err
		}
	}
	view.Worker, err = capture.WorkerStatus(directory)
	if err != nil {
		return err
	}
	for _, source := range []string{"codex", "opencode"} {
		manifest, e := capture.ReadManifest(directory, source)
		if e == nil {
			view.Installation = append(view.Installation, manifest)
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
	}
	if view.DiagnosticsContinuation {
		view.Limits = append(view.Limits, "Diagnostic inspection is partial. Continue status/diagnostics with --before-diagnostic diagnostics_before or increase --max-sessions/--max-source-bytes; current terminal facts remain stored.")
	}
	if view.Continuation {
		view.Limits = append(view.Limits, "Registration inspection reached max_sessions; output is partial. Increase the explicit capture session budget for a complete local inspection.")
	}
	if caller.Store != "" {
		caller.Store, err = capture.CanonicalStore(caller.Store)
		if err != nil {
			return err
		}
	}
	verifiedRows := []domain.CaptureRegistration{}
	for _, row := range view.Registrations {
		if caller.Source != "" && caller.Source != row.Identity.Source {
			continue
		}
		if caller.SessionID != "" && caller.SessionID != row.Identity.SessionID {
			continue
		}
		if caller.Store != "" {
			same, e := capture.SameStore(caller.Store, row.Identity.Store)
			if e != nil {
				return e
			}
			if !same {
				continue
			}
		}
		if row.ProjectProof != domain.CaptureProjectVerified || row.ProjectDirectory != project {
			continue
		}
		installation, configuredStore, e := capture.InstallationIdentity(directory, string(row.Identity.Source))
		if errors.Is(e, os.ErrNotExist) {
			continue
		}
		if e != nil {
			return e
		}
		if installation == "" || row.InstallationID != installation {
			continue
		}
		if configuredStore != "" {
			same, e := capture.SameStore(configuredStore, row.Identity.Store)
			if e != nil {
				return e
			}
			if !same {
				continue
			}
		}
		verifiedRows = append(verifiedRows, row)
	}
	for _, diagnostic := range view.Diagnostics {
		if diagnostic.Phase != "publication" || diagnostic.Destination == "" {
			continue
		}
		legacy := filepath.Join(filepath.Dir(diagnostic.Destination), "capture_report.json")
		if e := capture.ValidateDestination(project, legacy); e != nil {
			continue
		}
		if _, e := os.Lstat(legacy); e == nil {
			view.Limits = append(view.Limits, "An obsolete capture_report.json remains beside a recorded publication. Remove it manually after interrupted-publication recovery; use capture diagnostics for current evidence.")
			break
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
	}
	view.Verification = domain.CaptureVerification(verifiedRows, view.Diagnostics)
	if view.CurrentDiagnosticsContinuation {
		if view.Verification.State != "degraded" {
			view.Verification.State = "verification_pending"
		}
		view.Verification.Detail = "Current diagnostic evidence exceeds the inspection budget; stage flags describe retrieved evidence only. Increase --max-sessions/--max-source-bytes and inspect diagnostics continuation before claiming runtime verification."
	}
	if len(view.InstallationIssues) > 0 {
		view.Verification.State = "degraded"
		view.Verification.Detail = "Recorded exchange flags are historical evidence. Current owned installation is missing, edited or stale; inspect capture status installation_issues and explicitly refresh or remove owned capture setup."
	}
	if command == "verify" {
		return json.NewEncoder(output).Encode(view.Verification)
	}
	if command != "check" {
		return json.NewEncoder(output).Encode(view)
	}
	scopedRows := []domain.CaptureRegistration{}
	for _, row := range view.Registrations {
		if row.ProjectDirectory == "" || row.ProjectDirectory == project {
			scopedRows = append(scopedRows, row)
		}
	}
	evidence := domain.CaptureAdvisoryEvidence{Registrations: scopedRows, Diagnostics: view.Diagnostics}
	if caller.Source != "" && caller.Store != "" && caller.SessionID != "" {
		if err = caller.Validate(); err != nil {
			return err
		}
		caller.Store, err = capture.CanonicalStore(caller.Store)
		if err != nil {
			return err
		}
		evidence.Caller = &caller
	}
	tickets := map[string]bool{}
	for _, row := range view.Registrations {
		if row.TicketID != "" {
			tickets[row.TicketID] = true
		}
	}
	if len(tickets) > 0 {
		cfg, e := adaptersconfig.Load(project)
		if e != nil {
			return e
		}
		ordered := []string{}
		for ticket := range tickets {
			ordered = append(ordered, ticket)
		}
		sort.Strings(ordered)
		for _, ticket := range ordered {
			destination := filepath.Join(project, cfg.Workspaces.Dir, ticket, "agentic_metrics.json")
			if e = capture.ValidateDestination(project, destination); e != nil {
				return e
			}
			publication := domain.CapturePublicationEvidence{TicketID: ticket}
			// A persisted explicit destination is authoritative for this collection.
			for _, d := range view.Diagnostics {
				if d.TicketID == ticket && d.Phase == "publication" && d.Outcome == "committed" {
					destination = d.Destination
					publication.PublishedAt = d.ObservedAt
				}
			}
			if e = capture.ValidateDestination(project, destination); e != nil {
				return e
			}
			ledger, readErr := adapters.LoadObservedFactLedger(destination)
			if readErr == nil && ledger != nil && ledger.TicketID != nil && *ledger.TicketID == ticket {
				publication.Available = true
				publication.Sessions = ledger.Sessions
			}
			evidence.Publications = append(evidence.Publications, publication)
		}
	}
	issues = domain.CaptureAdvisoryIssues(evidence)
	if len(view.InstallationIssues) > 0 {
		issues = append(issues, "Owned capture installation is missing, edited or stale; inspect capture status and explicitly refresh or remove owned capture setup.")
	}
	if view.CurrentDiagnosticsContinuation {
		issues = append(issues, "Capture health reached its diagnostic row/byte budget; increase --max-sessions/--max-source-bytes and inspect capture diagnostics for remaining current outcomes.")
	}
	if view.Continuation {
		issues = append(issues, "Capture health inspection reached the configured session limit; increase telemetry.capture.max_sessions and inspect local diagnostics for remaining observations.")
	}
	if len(issues) > 0 {
		message := strings.Join(issues, " ")
		if len(message) > 1800 {
			message = message[:1800] + " See capture diagnostics."
		}
		_, err = fmt.Fprintln(output, "AgenticLens capture advisory:", message)
	}
	return err
}
