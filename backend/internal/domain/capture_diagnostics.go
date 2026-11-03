package domain

import "time"

// CaptureDiagnostic is bounded local operational evidence, never metric facts.
// A source acceptance does not establish successful artifact publication.
type CaptureDiagnostic struct {
	Sequence       int64                            `json:"sequence"`
	TicketID       string                           `json:"ticket_id,omitempty"`
	Phase          string                           `json:"phase"`
	Outcome        string                           `json:"outcome"`
	ObservedAt     string                           `json:"observed_at"`
	Detail         string                           `json:"detail,omitempty"`
	Generation     int64                            `json:"generation,omitempty"`
	Destination    string                           `json:"destination,omitempty"`
	Summary        string                           `json:"summary,omitempty"`
	Owners         []CapturePublishedOwner          `json:"owners,omitempty"`
	Reconciliation *CaptureReconciliationDiagnostic `json:"reconciliation,omitempty"`
}

type CapturePublishedOwner struct {
	Association       string        `json:"association,omitempty"`
	Store             string        `json:"store,omitempty"`
	InstallationID    string        `json:"installation_id,omitempty"`
	Source            SourceHarness `json:"source"`
	SessionID         string        `json:"session_id"`
	Checkpoint        string        `json:"checkpoint,omitempty"`
	SourceFingerprint string        `json:"source_fingerprint,omitempty"`
}

type CaptureReconciliationAttempt string

const (
	CaptureAttemptFailed    CaptureReconciliationAttempt = "failed"
	CaptureAttemptCompleted CaptureReconciliationAttempt = "completed"
	CaptureAttemptPartial   CaptureReconciliationAttempt = "partial"
	CaptureAttemptUnknown   CaptureReconciliationAttempt = "unknown"
)

// CaptureAttemptOutcome separates aggregate validation from individual reads.
// Older failed records do not identify whether a whole attempt or one owner
// failed. Successful legacy observations cannot acknowledge a typed fatal error.
func CaptureAttemptOutcome(d CaptureDiagnostic) CaptureReconciliationAttempt {
	if d.Phase != "reconciliation" {
		return ""
	}
	if d.Reconciliation != nil {
		switch d.Reconciliation.Attempt {
		case CaptureAttemptFailed, CaptureAttemptCompleted, CaptureAttemptPartial:
			return d.Reconciliation.Attempt
		case "":
		default:
			return CaptureAttemptUnknown
		}
	}
	return CaptureAttemptUnknown
}

type CaptureReconciliationDiagnostic struct {
	Attempt               CaptureReconciliationAttempt `json:"attempt_outcome,omitempty"`
	StartedAt             string                       `json:"started_at"`
	Cutoff                string                       `json:"cutoff"`
	SourceBytes           int64                        `json:"source_bytes_reserved"`
	SourceAttempts        int                          `json:"source_attempts"`
	DiscoveryPending      bool                         `json:"child_discovery_pending"`
	PendingChildren       int                          `json:"pending_children"`
	UnresolvedParentLinks int                          `json:"unresolved_parent_links"`
	Owners                []CaptureRegistration        `json:"owners"`
}

// CaptureHealth describes evidence of the live exchange independently of installation.
type CaptureHealth struct {
	State               string `json:"state"`
	CallbackObserved    bool   `json:"callback_observed"`
	BindingAccepted     bool   `json:"binding_accepted"`
	SourceAccepted      bool   `json:"source_accepted"`
	PublicationAccepted bool   `json:"publication_accepted"`
	Detail              string `json:"detail"`
}

func CaptureVerification(rows []CaptureRegistration, diagnostics []CaptureDiagnostic) CaptureHealth {
	h := CaptureHealth{State: "verification_pending", Detail: "Installation and protocol checks do not prove live callback delivery. Review normal harness trust and verify an actual callback, explicit binding response, source read and publication."}
	latest := map[string]CaptureDiagnostic{}
	publications := map[string]CaptureDiagnostic{}
	for _, d := range diagnostics {
		key := d.TicketID + "\x00" + d.Phase
		if old, ok := latest[key]; !ok || d.Sequence > old.Sequence {
			latest[key] = d
		}
		if d.Phase == "publication" && d.Outcome == "committed" {
			if old, ok := publications[d.TicketID]; !ok || d.Sequence > old.Sequence {
				publications[d.TicketID] = d
			}
		}
	}
	for _, d := range latest {
		if (d.Outcome == "failed" && d.Phase != "reconciliation") || d.Outcome == "staging" || d.Outcome == "preparing" {
			h.State = "degraded"
		}
	}
	verified := false
	attemptUnknown := false
	for _, row := range rows {
		callback := row.CallbackObservedAt != ""
		binding := row.Binding == "explicit_agent_response" && row.InstallationID != "" && row.BindingInstallationID == row.InstallationID
		source := row.RefreshedAt != "" && len(row.Snapshot) > 0 && row.SourceInstallationID == row.InstallationID
		published := false
		for _, owner := range publications[row.TicketID].Owners {
			if owner.Source == row.Identity.Source && owner.SessionID == row.Identity.SessionID && owner.Store == row.Identity.Store && owner.InstallationID == row.InstallationID {
				published = true
			}
		}
		h.CallbackObserved = h.CallbackObserved || callback
		h.BindingAccepted = h.BindingAccepted || binding
		h.SourceAccepted = h.SourceAccepted || source
		h.PublicationAccepted = h.PublicationAccepted || published
		verified = verified || (callback && binding && source && published)
		failures := CaptureCurrentFailures(row, diagnostics)
		attemptUnknown = attemptUnknown || failures.AttemptUnknown
		if row.State == CaptureFailed || failures.Publication || failures.Source || failures.Collection {
			h.State = "degraded"
		}
	}
	if attemptUnknown && h.State != "degraded" {
		h.Detail = "Historical reconciliation records do not identify the whole-attempt outcome. Run a complete capture reconcile before certifying current collection health."
	}
	if h.State != "degraded" && verified && !attemptUnknown {
		h.State = "verified"
		h.Detail = "One recorded session completed callback, explicit binding, accepted source read and committed publication. This verifies recorded sessions only; harness trust and unobserved sessions remain outside this evidence."
	}
	return h
}

// CaptureAdvisoryEvidence contains only persisted, exact-project observations.
type CaptureAdvisoryEvidence struct {
	Registrations []CaptureRegistration
	Diagnostics   []CaptureDiagnostic
	Publications  []CapturePublicationEvidence
	Caller        *CaptureIdentity
}

type CapturePublicationEvidence struct {
	TicketID    string
	Available   bool
	Sessions    []SessionFact
	PublishedAt string
}

// CaptureAdvisoryIssues does not treat pending work or missing caller identity as a defect.
func CaptureAdvisoryIssues(e CaptureAdvisoryEvidence) []string {
	issues := []string{}
	add := func(value string) {
		for _, old := range issues {
			if old == value {
				return
			}
		}
		issues = append(issues, value)
	}
	for _, row := range e.Registrations {
		if row.TicketID == "" {
			if row.ProjectProof != CaptureProjectVerified {
				continue
			}
			add("Observed sessions remain unbound; respond to their exact capture binding requests.")
			continue
		}
		failures := CaptureCurrentFailures(row, e.Diagnostics)
		if failures.Publication || failures.Collection {
			add("Capture has an unresolved operational or publication failure; inspect capture diagnostics and recover before relying on the latest metrics.")
		}
		if failures.AttemptUnknown {
			add("Historical whole-collection reconciliation outcome is unknown; run a complete capture reconcile before relying on current capture health.")
		}
		if row.State == CaptureFailed || row.ProjectProof == CaptureProjectUnverified || failures.Source {
			add("A source refresh failed or exact source metadata verification is unavailable; accepted facts were preserved. Inspect capture diagnostics and retry capture reconcile.")
		}
		var publication *CapturePublicationEvidence
		for i := range e.Publications {
			if e.Publications[i].TicketID == row.TicketID {
				publication = &e.Publications[i]
				break
			}
		}
		if publication == nil || !publication.Available {
			add("An observed bound collection has no readable metrics publication; run capture reconcile and generate.")
			continue
		}
		var published *SessionFact
		for i := range publication.Sessions {
			if publication.Sessions[i].ID == row.Identity.SessionID && (publication.Sessions[i].Source == row.Identity.Source || publication.Sessions[i].Source == "") {
				published = &publication.Sessions[i]
				break
			}
		}
		if published == nil {
			add("An observed bound contributor is absent from its metrics publication; reconcile and publish that collection.")
			continue
		}
		// Only source-backed checkpoints establish omitted data. Generic active or
		// pending markers, callback times and missing caller identity cannot do so.
		for _, accepted := range row.Snapshot {
			if accepted.ID != row.Identity.SessionID {
				continue
			}
			before, oldKnown := CaptureOwnerCheckpoint(*published)
			after, newKnown := CaptureOwnerCheckpoint(accepted)
			if !oldKnown || !newKnown || !after.After(before) {
				continue
			}
			readAt, readKnown := time.Parse(time.RFC3339Nano, row.RefreshedAt)
			publishedAt, publicationKnown := time.Parse(time.RFC3339Nano, publication.PublishedAt)
			postSnapshotCaller := e.Caller != nil && *e.Caller == row.Identity && row.State != CaptureFailed && readKnown == nil && publicationKnown == nil && readAt.After(publishedAt) && after.After(publishedAt)
			if !postSnapshotCaller {
				add("Accepted source data is newer than its metrics publication; reconcile and publish available facts.")
			}
		}
	}
	latest := map[string]CaptureDiagnostic{}
	for _, d := range e.Diagnostics {
		key := d.TicketID + "\x00" + d.Phase
		if previous, ok := latest[key]; !ok || d.Sequence > previous.Sequence {
			latest[key] = d
		}
	}
	for _, d := range latest {
		if (d.Outcome == "failed" && d.Phase != "reconciliation") || d.Outcome == "staging" || d.Outcome == "preparing" {
			add("Capture has an unresolved operational or publication failure; inspect capture diagnostics and recover before relying on the latest metrics.")
		}
		if d.Reconciliation != nil && (d.Reconciliation.PendingChildren > 0 || d.Reconciliation.UnresolvedParentLinks > 0 || d.Reconciliation.DiscoveryPending) {
			add("Recorded child evidence remains unresolved; continue capture reconcile.")
		}
	}
	return issues
}

// CaptureTerminalFailures is unresolved terminal evidence for an exact owner.
// Pending retries preserve it; only a relevant committed publication or successful
// exact source read supersedes the corresponding failure.
type CaptureTerminalFailures struct {
	Publication    bool
	Source         bool
	Collection     bool
	AttemptUnknown bool
}

func CaptureCurrentFailures(row CaptureRegistration, diagnostics []CaptureDiagnostic) CaptureTerminalFailures {
	if row.TicketID == "" {
		return CaptureTerminalFailures{}
	}
	var publication CaptureDiagnostic
	var source CaptureDiagnostic
	sourceFailed := false
	var attemptFailed, attemptCompleted, attemptUnknown int64
	for _, diagnostic := range diagnostics {
		if diagnostic.TicketID != row.TicketID {
			continue
		}
		if diagnostic.Phase == "publication" && (diagnostic.Outcome == "failed" || diagnostic.Outcome == "committed") && diagnostic.Sequence > publication.Sequence {
			publication = diagnostic
		}
		switch CaptureAttemptOutcome(diagnostic) {
		case CaptureAttemptFailed:
			if diagnostic.Sequence > attemptFailed {
				attemptFailed = diagnostic.Sequence
			}
		case CaptureAttemptCompleted:
			if diagnostic.Sequence > attemptCompleted {
				attemptCompleted = diagnostic.Sequence
			}
		case CaptureAttemptUnknown:
			if diagnostic.Sequence > attemptUnknown {
				attemptUnknown = diagnostic.Sequence
			}
		}
		if diagnostic.Phase != "reconciliation" || diagnostic.Sequence <= source.Sequence {
			continue
		}
		known, failed := false, false
		if diagnostic.Reconciliation != nil {
			for _, owner := range diagnostic.Reconciliation.Owners {
				if owner.Identity != row.Identity {
					continue
				}
				if owner.State == CaptureFailed {
					known = true
					failed = true
				}
				if owner.State == CaptureCurrent && owner.RefreshedAt != "" {
					known = true
				}
			}
		}
		if CaptureAttemptOutcome(diagnostic) == CaptureAttemptUnknown && diagnostic.Outcome == "failed" && (diagnostic.Reconciliation == nil || len(diagnostic.Reconciliation.Owners) == 0) {
			known = true
			failed = true
		}
		if known {
			source = diagnostic
			sourceFailed = failed
		}
	}
	if sourceFailed {
		acceptedAt, acceptedErr := time.Parse(time.RFC3339Nano, row.RefreshedAt)
		failedAt, failedErr := time.Parse(time.RFC3339Nano, source.ObservedAt)
		if acceptedErr == nil && failedErr == nil && acceptedAt.After(failedAt) {
			sourceFailed = false
		}
	}
	return CaptureTerminalFailures{Publication: publication.Outcome == "failed", Source: sourceFailed, Collection: attemptFailed > attemptCompleted, AttemptUnknown: attemptUnknown > attemptCompleted}
}

// CaptureFailureFeedback contains fixed copy only, never private diagnostic details.
func CaptureFailureFeedback(row CaptureRegistration, diagnostics []CaptureDiagnostic) string {
	failures := CaptureCurrentFailures(row, diagnostics)
	publicationFailed, sourceFailed := failures.Publication, failures.Source

	if !publicationFailed && !sourceFailed && !failures.Collection {
		if failures.AttemptUnknown {
			return "AgenticLens capture whole-collection reconciliation evidence is incomplete. Run a complete capture reconcile for this session's bound ticket before relying on current capture health."
		}
		return ""
	}
	message := "AgenticLens capture is degraded for this session's bound collection: "
	if publicationFailed {
		message += "a confirmed publication failed and has no newer committed replacement. "
	}
	if failures.Collection {
		message += "a confirmed whole-collection reconciliation failed and has no newer completed collection pass. "
	}
	if sourceFailed {
		message += "a confirmed source reconciliation failed and has no newer successful read. "
	}
	return message + "Previously accepted facts and artifacts remain the baseline; latest freshness or replacement is unconfirmed. Run capture diagnostics, then retry capture reconcile and generate using this exact project and ticket. Do not claim complete capture."
}
