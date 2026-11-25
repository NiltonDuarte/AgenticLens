package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// QueueChild durably stores only the source path and claimed parent for later verification.
func (s *Store) QueueChild(ctx context.Context, store, path, parent string) error {
	if !filepath.IsAbs(path) {
		return errors.New("child transcript path must be absolute")
	}
	canonicalStore, normalizeErr := CanonicalStore(store)
	if normalizeErr != nil {
		return normalizeErr
	}
	store = canonicalStore
	if err := codex.ValidateRolloutStore(store, path); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_children(worktree_id INTEGER NOT NULL,store TEXT,path TEXT,parent TEXT,PRIMARY KEY(worktree_id,store,path)); INSERT OR IGNORE INTO pending_children VALUES(?,?,?,?)`, s.worktreeID, store, path, parent)
	return err
}

type ReconcileResult struct {
	archiveReadGroups     []archiveReadGroup
	unresolvedChildLinks  []domain.ChildLink
	countedChildLinks     map[domain.ChildLink]bool
	reconciliation        *domain.CaptureDiagnostic
	IsolatedSources       []IsolatedSource                    `json:"-"`
	LifecycleRevision     int64                               `json:"-"`
	RestoresReady         []CaptureControl                    `json:"-"`
	ProjectRegistrations  []domain.CaptureRegistration        `json:"-"`
	SourceEvents          domain.SourceEvents                 `json:"-"`
	DiagnosticIndex       *DiagnosticIndexProgress            `json:"diagnostic_index,omitempty"`
	StartedAt             string                              `json:"started_at"`
	CutoffSemantics       string                              `json:"cutoff_semantics"`
	SourceBytes           int64                               `json:"source_bytes_reserved"`
	SourceAttempts        int                                 `json:"source_attempts"`
	RefreshedSessions     []domain.SessionFact                `json:"-"`
	RestoredSessions      []domain.SessionFact                `json:"-"`
	RestoredOwners        map[string][]domain.CaptureIdentity `json:"-"`
	DiscoveryPending      bool                                `json:"child_discovery_pending"`
	PendingChildren       int                                 `json:"pending_children"`
	UnresolvedParentLinks int                                 `json:"unresolved_parent_links"`
	Protocol              int                                 `json:"protocol"`
	Cutoff                string                              `json:"cutoff"`
	TicketID              string                              `json:"ticket_id"`
	Registrations         []domain.CaptureRegistration        `json:"registrations"`
	Sessions              []domain.SessionFact                `json:"-"`
	Limits                []string                            `json:"limits"`
}

func (s *Store) Reconcile(ctx context.Context, ticket, user string, sessionBudget int, byteBudget int64, policies ...domain.GenerationPolicy) (ReconcileResult, error) {
	// Explicit reconciliation can retry a completed unresolved Codex
	// sweep. Automatic generation uses ReconcileAt and needs new evidence.
	if ticket != "" && sessionBudget > 0 && byteBudget > 0 {
		if err := s.reopenTicketCodexChildSweeps(ctx, ticket); err != nil {
			return ReconcileResult{}, err
		}
	}
	return s.ReconcileAt(ctx, ticket, user, sessionBudget, byteBudget, time.Now().UTC(), policies...)
}

// ReconcileAt applies one duration cutoff and transient generation policy across source reads.
func (s *Store) ReconcileAt(ctx context.Context, ticket, user string, sessionBudget int, byteBudget int64, cutoff time.Time, policies ...domain.GenerationPolicy) (result ReconcileResult, returnErr error) {
	result = ReconcileResult{Protocol: RegistryProtocolVersion, StartedAt: cutoff.UTC().Format(time.RFC3339Nano), CutoffSemantics: "cutoff is reconciliation completion, not a common as-of source time; refreshed_at is each owner's successful read completion", TicketID: ticket, Limits: []string{"Only registered, explicitly bound or source-proven inherited sessions are selected. Native descendant discovery can recover a session whose lifecycle hook was absent.", "Hook installation is not proof of harness trust, agent response or complete source availability."}}
	if sessionBudget < 1 || byteBudget < 1 {
		return result, errors.New("capture budgets must be positive")
	}
	var err error
	result.LifecycleRevision, err = s.CaptureLifecycleRevision(ctx)
	if err != nil {
		return result, err
	}
	controlRevision, err := s.captureControlRevision(ctx)
	if err != nil {
		return result, err
	}
	// Header identification keeps a reserved quarter of the byte allowance
	// during the rotating rounds; one pass-local cache serves every header read.
	budget := &workBudget{bytes: byteBudget, sessions: sessionBudget, controlRevision: controlRevision, openCodeAttempts: map[string]bool{}, headers: codex.NewHeaderCache(), split: newBudgetSplit(byteBudget)}
	refreshed := map[string]bool{}
	defer func() {
		result.archiveReadGroups = nil
		result.Cutoff = time.Now().UTC().Format(time.RFC3339Nano)
		result.SourceBytes = budget.readBytes
		result.SourceAttempts = budget.attempts
		diagnostic := domain.CaptureDiagnostic{TicketID: ticket, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &domain.CaptureReconciliationDiagnostic{Attempt: domain.CaptureAttemptCompleted, StartedAt: result.StartedAt, Cutoff: result.Cutoff, SourceBytes: result.SourceBytes, SourceAttempts: result.SourceAttempts, DiscoveryPending: result.DiscoveryPending, PendingChildren: result.PendingChildren, UnresolvedParentLinks: result.UnresolvedParentLinks, Owners: append([]domain.CaptureRegistration(nil), result.Registrations...)}}
		if result.DiscoveryPending || result.PendingChildren > 0 || result.UnresolvedParentLinks > 0 {
			diagnostic.Reconciliation.Attempt = domain.CaptureAttemptPartial
		}
		if returnErr != nil {
			diagnostic.Reconciliation.Attempt = domain.CaptureAttemptFailed
			diagnostic.Outcome = "failed"
			diagnostic.Detail = "Reconciliation interrupted or failed; accepted source facts and pending recovery remain local. Retry capture reconcile."
		}
		for i, row := range diagnostic.Reconciliation.Owners {
			// Only retained lifecycle pending state can be normalized. Later source
			// parent/binding validation remains authoritative after a successful read.
			if refreshed[IdentityKey(row.Identity)] && row.State == domain.CapturePending {
				row.State = domain.CaptureCurrent
				diagnostic.Reconciliation.Owners[i].State = domain.CaptureCurrent
				diagnostic.Reconciliation.Owners[i].Detail = "Source read accepted during this reconciliation pass; prior numeric facts remain separately preserved."
			}
			if returnErr == nil && (row.State != domain.CaptureCurrent || !refreshed[IdentityKey(row.Identity)]) {
				diagnostic.Reconciliation.Attempt = domain.CaptureAttemptPartial
			}
			if row.State == domain.CaptureFailed {
				diagnostic.Outcome = "failed"
				if returnErr == nil {
					diagnostic.Detail = "One or more source snapshots unavailable; prior accepted facts preserved. Inspect source configuration and retry capture reconcile."
				}
			}
		}
		diagnosticCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := s.RecordDiagnostic(diagnosticCtx, diagnostic); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("persisting reconciliation diagnostics: %w", err))
		}
		result.unresolvedChildLinks = nil
		result.reconciliation = &diagnostic
	}()
	// Alternate ordinary and restore-prioritized rounds. The first eligible
	// class can use the full session and byte budget. Each round keeps its own
	// owner rotation so a blocked restore cannot starve ordinary capture.
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS recovery_phase(worktree_id INTEGER NOT NULL,ticket TEXT,phase INTEGER NOT NULL,PRIMARY KEY(worktree_id,ticket))`); err != nil {
		return result, err
	}
	phase := 0
	phaseErr := s.db.QueryRowContext(ctx, `SELECT phase FROM recovery_phase WHERE worktree_id=? AND ticket=?`, s.worktreeID, ticket).Scan(&phase)
	if phaseErr != nil && !errors.Is(phaseErr, sql.ErrNoRows) {
		return result, phaseErr
	}
	phase = phase % RestoreSchedulerRotationPasses
	if _, err := s.db.ExecContext(ctx, `INSERT INTO recovery_phase VALUES(?,?,?)ON CONFLICT(worktree_id,ticket) DO UPDATE SET phase=excluded.phase`, s.worktreeID, ticket, (phase+1)%RestoreSchedulerRotationPasses); err != nil {
		return result, err
	}
	if phase >= int(workClassCount) {
		budget.restorePriority, err = s.restoreWorkPriorities(ctx, ticket)
		if err != nil {
			return result, err
		}
	}
	attempted := map[string]bool{}
	refreshProjection := map[string]domain.CaptureRegistration{}
	archiveProjection := map[string]domain.CaptureRegistration{}
	for offset := 0; offset < int(workClassCount); offset++ {
		class := workClass((phase + offset) % int(workClassCount))
		if class == workRefresh {
			updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted, archiveProjection, policies...)
			if err != nil {
				return result, err
			}
			for key, value := range accepted {
				refreshed[key] = value
			}
			for _, owner := range updatedOwners {
				refreshProjection[IdentityKey(owner.Identity)] = owner
			}
		} else {
			_, err := s.recoverChildren(ctx, ticket, budget, class)
			if err != nil {
				return result, err
			}
		}
	}
	// Discoveries made after the refresh class can use leftover resources
	// immediately, including identification's unused share. Each owner is
	// attempted at most once per pass, except an owner deferred only by the
	// identification reservation.
	budget.releaseSplit()
	if budget.sessions > 0 && budget.bytes > 0 {
		updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted, archiveProjection, policies...)
		if err != nil {
			return result, err
		}
		for key, value := range accepted {
			refreshed[key] = value
		}
		for _, owner := range updatedOwners {
			refreshProjection[IdentityKey(owner.Identity)] = owner
		}
	}
	registrations, err := s.List(ctx)
	if err != nil {
		return result, err
	}
	var currentRows map[string]domain.CaptureRegistration
	if len(result.archiveReadGroups) > 0 {
		currentRows = make(map[string]domain.CaptureRegistration, len(registrations))
		for _, row := range registrations {
			currentRows[IdentityKey(row.Identity)] = row
		}
	}
	staleEvidence := map[string]bool{}
	for i, owner := range registrations {
		key := IdentityKey(owner.Identity)
		if observed, ok := archiveProjection[key]; ok {
			delete(archiveProjection, key)
			if sameCaptureReadObservation(owner, observed) {
				// Keep current facts and any intervening annotation updates.
				// A discarded read has only a request-local pending diagnostic.
				if observed.Detail == errCaptureReadObservationChanged.Error() {
					registrations[i] = pendingChangedObservation(owner)
				}
			} else {
				delete(refreshed, key)
				staleEvidence[owner.Identity.SessionID] = true
				registrations[i] = pendingChangedObservation(owner)
			}
		} else if projected, ok := refreshProjection[key]; ok {
			if sameCaptureReadObservation(owner, projected) {
				registrations[i] = projected
			} else if refreshed[key] {
				delete(refreshed, key)
				staleEvidence[owner.Identity.SessionID] = true
				registrations[i] = pendingChangedObservation(owner)
			}
		}
	}
	for key, missing := range archiveProjection {
		delete(refreshed, key)
		staleEvidence[missing.Identity.SessionID] = true
	}
	if err := result.reprojectArchiveFamilies(registrations, currentRows, refreshed, staleEvidence); err != nil {
		return result, err
	}
	result.SourceEvents = withoutCaptureEvidence(result.SourceEvents, staleEvidence)
	// Fixed-point propagation uses only a parent read from an accepted source snapshot.
	// Each successful pass binds at least one owner; cycles make no progress and remain unassigned.
	for pass := 0; pass < len(registrations); pass++ {
		byID := map[string]domain.CaptureRegistration{}
		for _, r := range registrations {
			byID[string(r.Identity.Source)+"\x00"+r.Identity.Store+"\x00"+r.Identity.SessionID] = r
		}
		changed := false
		for i, r := range registrations {
			if r.ParentID == "" || len(r.Snapshot) == 0 {
				continue
			}
			parent, ok := byID[string(r.Identity.Source)+"\x00"+r.Identity.Store+"\x00"+r.ParentID]
			if !ok || parent.TicketID == "" {
				continue
			}
			if r.TicketID != "" {
				if r.TicketID != parent.TicketID && r.Binding != "explicit_agent_response" && r.Binding != "explicit_manual_selection" {
					r.State = domain.CaptureFailed
					r.Detail = "Conflicting source-proven parent ticket binding"
					registrations[i] = r
				}
				continue
			}
			updated, e := s.Change(ctx, r.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				bound, bindErr := domain.CaptureBinding(current.TicketID, "", parent.TicketID)
				if bindErr != nil {
					return current, bindErr
				}
				current.TicketID = bound
				current.Binding = "source_parent"
				if current.State == domain.CaptureUnassigned {
					current.State = domain.CaptureCurrent
				}
				return current, nil
			})
			if e != nil {
				return result, e
			}
			registrations[i] = updated
			changed = true
		}
		if !changed {
			break
		}
	}

	knownParents := map[string]bool{}
	result.ProjectRegistrations = registrations
	for _, r := range registrations {
		knownParents[string(r.Identity.Source)+"\x00"+r.Identity.Store+"\x00"+r.Identity.SessionID] = true
	}
	controls, err := s.captureControls(ctx, nil, "", CaptureControlDiscarded, false)
	if err != nil {
		return result, err
	}
	registrations, suppressed := activeCaptureRows(s.project, registrations, controls)
	if len(suppressed) > 0 {
		result.SourceEvents = withoutCaptureEvidence(result.SourceEvents, suppressed)
	}
	pending, discoveryPending, err := s.pendingChildWork(ctx, ticket, registrations)
	if err != nil {
		return result, err
	}
	result.countedChildLinks, err = s.pendingChildLinks(ctx, ticket, registrations)
	if err != nil {
		return result, err
	}
	result.PendingChildren = pending
	// Links kept without re-proof are unresolved child evidence, counted once
	// beside pending child work for the same parent.
	result.addUnresolvedChildLinks(result.unresolvedChildLinks)
	result.DiscoveryPending = budget.discoveryPending || discoveryPending
	owners := map[string]string{}
	for _, r := range registrations {
		if r.TicketID == ticket {
			if r.ParentID != "" && !knownParents[string(r.Identity.Source)+"\x00"+r.Identity.Store+"\x00"+r.ParentID] {
				result.UnresolvedParentLinks++
			}
			result.Registrations = append(result.Registrations, r)
		}
		if r.TicketID == ticket && (s.project == "" || r.ProjectDirectory == s.project) {
			for _, session := range r.Snapshot {
				key := IdentityKey(r.Identity)
				if previous, ok := owners[session.ID]; ok && previous != key {
					return result, fmt.Errorf("session identity collision: %s exists in multiple registered source stores; canonical ledger cannot distinguish these owners", session.ID)
				}
				owners[session.ID] = key
			}
			result.Sessions = append(result.Sessions, r.Snapshot...)
			if refreshed[IdentityKey(r.Identity)] {
				result.RefreshedSessions = append(result.RefreshedSessions, r.Snapshot...)
			}
		}
	}
	err = s.captureRestoresReady(ctx, &result, cutoff, registrations, attempted, refreshed, budget.openCodeAttempts)
	if err != nil {
		return result, err
	}
	currentRevision, err := s.CaptureLifecycleRevision(ctx)
	if err != nil {
		return result, err
	}
	if currentRevision != result.LifecycleRevision {
		return result, ErrCaptureLifecycleChanged
	}
	return result, nil
}

// addUnresolvedChildLinks counts accepted delegation links this request did not
// re-prove, retained or refused, as unresolved child evidence
// (PendingChildren). A link already counted as pending child work for the same
// parent, or earlier in this request, is counted once. It returns the number
// of links added.
func (result *ReconcileResult) addUnresolvedChildLinks(links []domain.ChildLink) int {
	if result.countedChildLinks == nil {
		result.countedChildLinks = make(map[domain.ChildLink]bool, len(links))
	}
	added := 0
	for _, link := range links {
		if link.ParentID == "" || link.ChildSessionID == "" || result.countedChildLinks[link] {
			continue
		}
		result.countedChildLinks[link] = true
		added++
	}
	result.PendingChildren += added
	return added
}

// reconciliationDiagnostic returns this request's recorded reconciliation
// diagnostic with its current unresolved child evidence. A completed attempt
// with unresolved child evidence becomes partial. It reports false when the
// request recorded no reconciliation.
func (result ReconcileResult) reconciliationDiagnostic() (domain.CaptureDiagnostic, bool) {
	if result.reconciliation == nil || result.reconciliation.Reconciliation == nil {
		return domain.CaptureDiagnostic{}, false
	}
	diagnostic := *result.reconciliation
	reconciliation := *diagnostic.Reconciliation
	reconciliation.Owners = append([]domain.CaptureRegistration(nil), reconciliation.Owners...)
	reconciliation.PendingChildren = result.PendingChildren
	if reconciliation.Attempt == domain.CaptureAttemptCompleted && result.PendingChildren > 0 {
		reconciliation.Attempt = domain.CaptureAttemptPartial
	}
	diagnostic.Reconciliation = &reconciliation
	return diagnostic, true
}

// ChildLinkRetention reports one application of RetainAcceptedChildLinks.
type ChildLinkRetention struct {
	// Kept is the number of distinct accepted links kept without re-proof.
	Kept int
	// Refused is the number of distinct accepted links, not contradicted, that
	// the rule declined to carry.
	Refused int
	// Cleared is the number of fresh links cleared as contradicted.
	Cleared int
	// Added is the number of links newly counted as unresolved child evidence.
	Added int
}

// RetainAcceptedChildLinks keeps accepted Codex delegation child links on the
// owners a reconciliation refreshed or restored without re-reading their
// children, using accepted facts from another store, such as the committed
// ledger. The request's registry snapshots are child evidence ahead of those
// facts. Kept and refused links count as unresolved child evidence. When that
// count grows, the request's reconciliation is recorded again as partial in
// the project registry, which is opened only then. With no accepted facts the
// rule only clears fresh links that registry or fresh child records
// contradict. Transcript events are unchanged. ctx must bound only this call:
// when it expires during the registry open or write, the error satisfies
// IsRegistryContention.
func RetainAcceptedChildLinks(ctx context.Context, directory, project string, result *ReconcileResult, accepted []domain.SessionFact) (ChildLinkRetention, error) {
	var retention ChildLinkRetention
	if result == nil {
		return retention, nil
	}
	refreshed := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Known: result.Sessions, Fresh: result.RefreshedSessions, Observed: result.RestoredSessions})
	restored := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: accepted, Known: result.Sessions, Fresh: result.RestoredSessions, Observed: result.RefreshedSessions})
	result.RefreshedSessions = refreshed.Sessions
	result.RestoredSessions = restored.Sessions
	kept := map[domain.ChildLink]bool{}
	refused := map[domain.ChildLink]bool{}
	for _, report := range []domain.ChildLinkReport{refreshed.Report, restored.Report} {
		for _, link := range report.RetainedLinks() {
			kept[link] = true
		}
		for _, link := range report.Refused {
			refused[link.Link] = true
		}
		for _, decision := range report.Decisions {
			if decision.Outcome == domain.ChildLinkCleared {
				retention.Cleared++
			}
		}
	}
	for link := range kept {
		delete(refused, link)
	}
	retention.Kept, retention.Refused = len(kept), len(refused)
	retention.Added = result.addUnresolvedChildLinks(append(refreshed.Report.UnresolvedLinks(), restored.Report.UnresolvedLinks()...))
	if retention.Added == 0 {
		return retention, nil
	}
	diagnostic, recorded := result.reconciliationDiagnostic()
	if !recorded {
		return retention, nil
	}
	store, err := OpenProjectContext(ctx, directory, project)
	if err == nil {
		err = errors.Join(store.RecordDiagnostic(ctx, diagnostic), store.Close())
	}
	if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		// The caller bounds only this registry open and diagnostic write. Its
		// deadline expires while another process holds the registry.
		err = fmt.Errorf("%w: %w", errRegistryWaitExpired, err)
	}
	return retention, err
}

// sourceRead never enters registry serialization or diagnostic publication.
type sourceRead struct {
	configuration domain.ProjectConfig
	policy        *domain.GenerationPolicy
	transcripts   []codex.Transcript
	events        domain.SourceEvents
}

func extractExactAt(ctx context.Context, r domain.CaptureRegistration, budget int64, user string, cutoff time.Time, source *sourceRead, metadata *codex.SessionMetadataContext) ([]domain.SessionFact, int64, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, "", err
	}
	if budget < 1 {
		return nil, 0, "", opencode.ErrCaptureBudget
	}
	if r.Identity.Source == domain.SourceCodex {
		path := r.TranscriptPath
		file, err := os.Open(path)
		if err != nil {
			return nil, 0, "", err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return nil, 0, "", err
		}
		if !info.Mode().IsRegular() || info.Size() > budget {
			return nil, 0, "", opencode.ErrCaptureBudget
		}
		transcript, err := codex.ParseRolloutAt(io.LimitReader(file, info.Size()), cutoff)
		if err != nil {
			return nil, info.Size(), "", err
		}
		if len(transcript.Warnings) > 0 {
			return nil, info.Size(), "", errors.New("partial Codex source parse; previous snapshot must be preserved")
		}
		if transcript.ThreadID != r.Identity.SessionID {
			return nil, info.Size(), "", errors.New("hook/source identity mismatch")
		}
		if metadata != nil && (transcript.WorkingDir != metadata.WorkingDirectory || transcript.ParentThread != metadata.ParentID) {
			return nil, info.Size(), "", errors.New("Codex ownership metadata changed during the source read; retry reconciliation")
		}
		var policies []domain.GenerationPolicy
		if source.policy != nil {
			policies = append(policies, *source.policy)
		}
		sessions, err := transcript.SessionFacts(policies...)
		if err != nil {
			return nil, info.Size(), "", err
		}
		for i := range sessions {
			sessions[i].User = user
		}
		source.transcripts = append(source.transcripts, transcript)
		return sessions, info.Size(), transcript.UpdatedAt, nil
	}
	db, err := openOpenCodeSource(source.configuration, r.Identity.Store)
	if err != nil {
		return nil, 0, "", err
	}
	defer db.Close()
	var events domain.SourceEvents
	sessions, bytes, updated, readErr := opencode.CaptureSessionSnapshotWithPolicyAt(ctx, db.DB(), r.Identity.SessionID, user, budget, source.policy, cutoff, &events)
	if readErr == nil {
		source.events = events
	}
	return sessions, bytes, updated, readErr
}

var errCaptureReadObservationChanged = errors.New("capture registration changed during source read; retry reconciliation")

func sameCaptureReadObservation(current, observed domain.CaptureRegistration) bool {
	// Store.Change already resolved the same store and session identity under
	// its write lock. Lifecycle events do not change source authority.
	return current.TranscriptPath == observed.TranscriptPath &&
		current.TicketID == observed.TicketID &&
		current.Binding == observed.Binding &&
		current.Request == observed.Request &&
		current.BindingAcceptedAt == observed.BindingAcceptedAt &&
		current.BindingInstallationID == observed.BindingInstallationID &&
		current.ProjectDirectory == observed.ProjectDirectory &&
		current.ProjectProof == observed.ProjectProof &&
		current.ParentID == observed.ParentID &&
		current.RefreshedAt == observed.RefreshedAt &&
		current.Revision == observed.Revision &&
		current.UpdatedAt == observed.UpdatedAt &&
		current.SourceFingerprint == observed.SourceFingerprint &&
		current.SourceInstallationID == observed.SourceInstallationID &&
		current.SnapshotNeedsRefresh == observed.SnapshotNeedsRefresh
}

func (s *Store) saveCaptureRead(ctx context.Context, observed domain.CaptureRegistration, controlRevision int64, apply func(domain.CaptureRegistration) (domain.CaptureRegistration, error)) (updated domain.CaptureRegistration, returnErr error) {
	if err := observed.Identity.Validate(); err != nil {
		return domain.CaptureRegistration{}, err
	}
	id, err := canonicalIdentity(observed.Identity)
	if err != nil {
		return domain.CaptureRegistration{}, err
	}
	returnErr = s.withCaptureControlRegistrationWrite(ctx, controlRevision, func(conn *sql.Conn) error {
		updated, err = s.changeRegistration(ctx, conn, id, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			if !sameCaptureReadObservation(current, observed) {
				return current, errCaptureReadObservationChanged
			}
			return apply(current)
		})
		return err
	})
	return updated, returnErr
}

func pendingChangedObservation(current domain.CaptureRegistration) domain.CaptureRegistration {
	current.State = domain.CapturePending
	current.Detail = errCaptureReadObservationChanged.Error()
	return current
}

func captureParentID(owner domain.SessionFact) string {
	if owner.ParentID == nil {
		return ""
	}
	return *owner.ParentID
}

func keepsAcceptedCaptureParent(observed domain.CaptureRegistration, parentID string) bool {
	for _, accepted := range observed.Snapshot {
		if accepted.ID == observed.Identity.SessionID && accepted.Source == observed.Identity.Source {
			return captureParentID(accepted) == parentID
		}
	}
	// A source-proven child can have ancestry before its first snapshot.
	return observed.ParentID == "" || observed.ParentID == parentID
}

func withoutCaptureEvidence(source domain.SourceEvents, stale map[string]bool) domain.SourceEvents {
	if len(stale) == 0 {
		return source
	}
	var remaining domain.SourceEvents
	for _, id := range source.CompleteSessionIDs {
		if !stale[id] {
			remaining.CompleteSessionIDs = append(remaining.CompleteSessionIDs, id)
		}
	}
	for _, event := range source.Events {
		if !stale[event.SessionID] {
			remaining.Events = append(remaining.Events, event)
		}
	}
	for _, evidence := range source.CheckpointEvidence {
		if !stale[evidence.SessionID] {
			remaining.CheckpointEvidence = append(remaining.CheckpointEvidence, evidence)
		}
	}
	return remaining
}

func (s *Store) validateReadOwner(ctx context.Context, r domain.CaptureRegistration, budget *workBudget) (*codex.SessionMetadataContext, error) {
	if s.project == "" {
		return nil, nil
	}
	if r.Identity.Source != domain.SourceCodex {
		const reservation = 256
		if budget.extractionRemaining() < reservation {
			return nil, opencode.ErrCaptureBudget
		}
		budget.chargeExtraction(reservation)
		return nil, s.validateOwner(ctx, r)
	}

	info, err := os.Stat(r.TranscriptPath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("Codex source is not a regular file")
	}
	// The owner's own header preflight is part of reading that owner. It is
	// charged the bytes actually read, and a header cached in this pass is free.
	read, err := budget.readHeader(ctx, r.TranscriptPath, min(budget.extractionRemaining(), s.configuration.Telemetry.Capture.MaxSourceBytes))
	budget.chargeExtraction(read.Bytes)
	if errors.Is(err, codex.ErrDiscoveryBudget) {
		return nil, opencode.ErrCaptureBudget
	}
	if err != nil {
		return nil, err
	}
	metadata, err := read.Header.SessionContext()
	if err != nil {
		return nil, err
	}
	return &metadata, validateCodexOwnerMetadata(s.project, r, metadata)
}

func (s *Store) prepareReadOwner(ctx context.Context, r domain.CaptureRegistration, budget *workBudget) (domain.CaptureRegistration, *codex.SessionMetadataContext, bool, error) {
	if r.Identity.Source != domain.SourceCodex {
		metadata, err := s.validateReadOwner(ctx, r, budget)
		return r, metadata, false, err
	}
	registered, relocating, err := inspectRegisteredCodexPath(r)
	if err != nil {
		return r, nil, false, err
	}
	if relocating && r.TicketID == "" && s.project == "" {
		// An unassigned row is recovered only with project proof.
		return r, nil, true, fmt.Errorf("unassigned Codex source relocation requires project proof: %w", codex.ErrOwnerUnavailable)
	}
	var metadata *codex.SessionMetadataContext
	if !relocating {
		metadata, err = s.validateReadOwner(ctx, r, budget)
		if err != nil {
			return r, nil, false, err
		}
	}
	r.TranscriptPath, err = budget.selectCodexRollout(r, s.configuration, registered, relocating)
	if err == nil && relocating {
		metadata, err = s.validateReadOwner(ctx, r, budget)
		if err != nil && metadata != nil {
			// The recovered rollout was read but contradicts the registration.
			err = relocationConflict{err: err}
		}
	}
	return r, metadata, relocating, err
}

// ValidateSessionMerge detects identities the canonical ledger cannot distinguish.
func ValidateSessionMerge(existing, fresh []domain.SessionFact) error {
	byID := map[string]domain.SourceHarness{}
	for _, s := range append(append([]domain.SessionFact{}, existing...), fresh...) {
		if source, ok := byID[s.ID]; ok && source != "" && s.Source != "" && source != s.Source {
			return fmt.Errorf("session ID %s collides across source harnesses", s.ID)
		}
		byID[s.ID] = s.Source
	}
	return nil
}

// CoverageSummary contains no transcript paths or source-store paths.
func CoverageSummary(result ReconcileResult) string {
	counts := map[domain.CaptureState]int{}
	for _, r := range result.Registrations {
		counts[r.State]++
	}
	retained := 0
	for _, r := range result.Registrations {
		if r.State != domain.CaptureCurrent && len(r.Snapshot) > 0 {
			retained++
		}
	}
	summary := fmt.Sprintf("\nCapture observation finalized: %s (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: %d; retained older facts: %d; pending: %d; failed: %d; unassigned within this collection: %d. Unresolved child evidence: %d; parent links: %d; discovery continuation: %t. Source attempts: %d; reserved bytes: %d. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.\n", result.Cutoff, counts[domain.CaptureCurrent], retained, counts[domain.CapturePending], counts[domain.CaptureFailed], counts[domain.CaptureUnassigned], result.PendingChildren, result.UnresolvedParentLinks, result.DiscoveryPending, result.SourceAttempts, result.SourceBytes)
	for index, row := range result.Registrations {
		if index >= 16 {
			summary += "Additional owner freshness details remain in local capture diagnostics.\n"
			break
		}
		refreshed := "unavailable"
		if value, err := time.Parse(time.RFC3339Nano, row.RefreshedAt); err == nil {
			refreshed = value.UTC().Format(time.RFC3339Nano)
		}
		checkpoint := "unknown"
		if value, err := time.Parse(time.RFC3339Nano, row.UpdatedAt); err == nil {
			checkpoint = value.UTC().Format(time.RFC3339Nano)
		}
		summary += fmt.Sprintf("Owner %s/%s: %s; last successful source read %s; accepted source checkpoint %s.\n", row.Identity.Source, row.Identity.SessionID, row.State, refreshed, checkpoint)
	}
	return summary
}

// refreshTicketOwners is one scheduler work class. Restore-prioritized turns
// cover owners that still need exact-revision source evidence.
func (s *Store) refreshTicketOwners(ctx context.Context, ticket, user string, budget *workBudget, result *ReconcileResult, attempted map[string]bool, archiveProjection map[string]domain.CaptureRegistration, policies ...domain.GenerationPolicy) ([]domain.CaptureRegistration, map[string]bool, error) {
	registrations, err := s.ListActive(ctx)
	if err != nil {
		return nil, nil, err
	}
	previousSnapshots := []domain.SessionFact{}
	known := make(map[string]domain.CaptureRegistration, len(registrations))
	for _, registration := range registrations {
		known[IdentityKey(registration.Identity)] = registration
		if registration.TicketID == ticket {
			previousSnapshots = append(previousSnapshots, registration.Snapshot...)
		}
	}
	if _, err = s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS reconcile_cursor(worktree_id INTEGER NOT NULL,ticket TEXT,identity TEXT NOT NULL,PRIMARY KEY(worktree_id,ticket))`); err != nil {
		return nil, nil, err
	}
	const restoreCursorKind = "restore-owner-reads"
	var cursor string
	if budget.restorePriority == nil {
		err = s.db.QueryRowContext(ctx, `SELECT identity FROM reconcile_cursor WHERE worktree_id=? AND ticket=?`, s.worktreeID, ticket).Scan(&cursor)
	} else {
		cursor, err = s.childCursor(ctx, restoreCursorKind, ticket)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}
	indexes := []int{}
	for i, r := range registrations {
		identity := IdentityKey(r.Identity)
		if r.TicketID != ticket || attempted[identity] {
			continue
		}
		if budget.restorePriority != nil {
			switch budget.restorePriority.reads[identity] {
			case 0:
				continue
			case restoreReadCovered:
				if budget.sessions < 1 || budget.bytes < 1 {
					continue
				}
				readAt, err := time.Parse(time.RFC3339Nano, r.RefreshedAt)
				if err == nil && s.restoreReadSourceUnchanged(ctx, r, readAt) {
					continue
				}
				budget.restorePriority.reads[identity] = restoreAttemptPriority(true)
			}
		}
		indexes = append(indexes, i)
	}
	// Persisted rotation prevents an early growing owner from starving later roots.
	split := 0
	for split < len(indexes) && IdentityKey(registrations[indexes[split]].Identity) <= cursor {
		split++
	}
	indexes = append(append([]int{}, indexes[split:]...), indexes[:split]...)
	if budget.restorePriority != nil {
		sort.SliceStable(indexes, func(i, j int) bool {
			left, right := registrations[indexes[i]].Identity, registrations[indexes[j]].Identity
			return budget.restorePriority.reads[IdentityKey(left)] > budget.restorePriority.reads[IdentityKey(right)]
		})
	}
	if len(budget.deferred) > 0 {
		// Owners deferred by the identification reservation come first in the
		// final leftover refresh, ahead of owners discovered during the pass.
		sort.SliceStable(indexes, func(i, j int) bool {
			return budget.deferred[IdentityKey(registrations[indexes[i]].Identity)] && !budget.deferred[IdentityKey(registrations[indexes[j]].Identity)]
		})
	}

	if len(policies) == 0 {
		if err := adaptersconfig.ValidateGenerationPolicy(s.configuration); err != nil {
			return nil, nil, err
		}
		policies = []domain.GenerationPolicy{s.configuration.Generation.Metrics}
	}
	if policies[0].Anomalies {
		if err := adaptersconfig.ValidateAnomalyPolicy(s.configuration); err != nil {
			return nil, nil, err
		}
	}
	source := sourceRead{policy: captureGenerationPolicy(policies), configuration: s.configuration}
	used := 0
	refreshed := map[string]bool{}
	archiveAttempts := map[string]bool{}
	acceptedRelocation := false
	for _, i := range indexes {
		if budget.sessions < 1 {
			for _, remaining := range indexes[used:] {
				registrations[remaining].State = domain.CapturePending
				registrations[remaining].Detail = "Not refreshed within this pass; continue reconciliation"
			}
			result.Limits = append(result.Limits, "Session budget reached; remaining registrations stay pending for the next pass.")
			break
		}
		r := registrations[i]
		priorSourceBytes := budget.readBytes
		attempted[IdentityKey(r.Identity)] = true
		used++
		budget.sessions--
		budget.attempts++
		var fresh []domain.SessionFact
		var revision int64
		var updatedAt string
		sourceOwner, metadata, relocating, extractErr := s.prepareReadOwner(ctx, r, budget)
		var fingerprint string
		var fingerprintErr error
		read := sourceRead{policy: source.policy, configuration: source.configuration}
		if extractErr == nil {
			// A fingerprint is only an operational hint. Always read the
			// selected owner within the remaining reconciliation budget.
			fingerprint, fingerprintErr = sourceFingerprint(ctx, sourceOwner, s.configuration)
			cutoff, parseErr := time.Parse(time.RFC3339Nano, result.StartedAt)
			if parseErr != nil {
				return nil, nil, parseErr
			}
			fresh, revision, updatedAt, extractErr = extractExactAt(ctx, sourceOwner, budget.extractionRemaining(), user, cutoff, &read, metadata)
			if extractErr == nil && relocating && len(fresh) > 0 && !keepsAcceptedCaptureParent(r, captureParentID(fresh[0])) {
				extractErr = relocationConflict{err: errors.New("recovered Codex source conflicts with the accepted parent identity")}
			}
		}
		budget.chargeExtraction(revision)
		key := IdentityKey(r.Identity)
		// A moved unassigned rollout that cannot be recovered on exact proof is
		// isolated: its facts stay, and it neither fails nor stalls the pass.
		var isolation IsolationReason
		isolated := false
		if r.Identity.Source == domain.SourceCodex && r.TicketID == "" && relocating {
			isolation, isolated = unassignedRelocationIsolation(extractErr)
		}
		// An owner refused for budget while the identification reservation
		// limits extraction is retried by the final leftover refresh. The
		// reservation need not be the cause, for example a discovery session
		// limit; the retry is then one bounded extra attempt.
		deferred := (errors.Is(extractErr, opencode.ErrCaptureBudget) || errors.Is(extractErr, codex.ErrDiscoveryBudget)) && budget.extractionReserved()
		updated, updateErr := s.saveCaptureRead(ctx, r, budget.controlRevision, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			if isolated {
				current.State = domain.CapturePending
				current.Detail = fmt.Sprintf("Unassigned Codex source %s after its rollout moved; isolated from this pass and retried on the next integration event or capture reconcile; previous facts preserved: %v", isolation, extractErr)
				return current, nil
			}
			if extractErr != nil {
				current.State = domain.CaptureFailed
				current.Detail = "Source snapshot unavailable or invalid; previous facts preserved: " + extractErr.Error()
				if errors.Is(extractErr, opencode.ErrCaptureBudget) {
					current.State = domain.CapturePending
					current.Detail = "Source byte budget reached; increase --max-source-bytes or continue reconciliation; previous facts preserved"
				} else if errors.Is(extractErr, codex.ErrDiscoveryBudget) {
					current.State = domain.CapturePending
					current.Detail = "Codex discovery session/byte budget reached; increase configured source limits or continue reconciliation; previous facts preserved"
				}
				return current, nil
			}
			if len(fresh) == 0 {
				current.State = domain.CapturePending
				current.Detail = "Source session has not flushed; previous facts preserved"
				return current, nil
			}

			owner := fresh[0]
			current.ParentID = captureParentID(owner)
			// Lean auxiliary facts belong to the exact owner snapshot.
			compact, compactErr := domain.PersistAnomalySummaries(domain.AgenticFactLedger{Sessions: fresh})
			if compactErr != nil {
				return current, compactErr
			}
			if relocating {
				current.TranscriptPath = sourceOwner.TranscriptPath
			}
			// Keep accepted links in this first commit, so a pass that stops
			// before the family annotation never leaves fewer links than
			// before. The annotation applies the same rule to the family read.
			retention := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: previousSnapshots, Fresh: compact.Sessions})
			current.Snapshot = domain.RetainIncompleteGroupSummaries(previousSnapshots, retention.Sessions)
			current.SnapshotNeedsRefresh = false
			current.ProjectDirectory = s.project
			current.ProjectProof = domain.CaptureProjectVerified
			current.SourceFingerprint = fingerprint
			current.Revision = revision
			current.UpdatedAt = updatedAt
			current.State = domain.CaptureCurrent
			current.SourceInstallationID = current.InstallationID
			current.RefreshedAt = time.Now().UTC().Format(time.RFC3339Nano)
			refreshed[key] = true
			current.Detail = "Source snapshot reconciled; checkpoint does not establish ticket completion"
			if current.TicketID == "" {
				current.State = domain.CaptureUnassigned
				current.Detail = "Respond to the session binding request; no ticket inferred"
			}
			return current, nil
		})
		if errors.Is(updateErr, errCaptureReadObservationChanged) {
			updated = pendingChangedObservation(updated)
			budget.refreshCursorHeld = true
		} else if updateErr != nil {
			return nil, nil, updateErr
		}
		if relocating {
			archiveAttempts[key] = true
		}
		if isolated && updateErr == nil {
			result.IsolatedSources = append(result.IsolatedSources, IsolatedSource{Identity: updated.Identity, Reason: isolation, Detail: updated.Detail})
		}
		if deferred && updateErr == nil {
			// The owner was not read; it keeps its session slot and is retried
			// first by the leftover refresh. The rotation cursor follows the
			// existing rule, so an owner too large for any pass cannot pin it.
			delete(attempted, key)
			budget.sessions++
			budget.attempts--
			if budget.deferred == nil {
				budget.deferred = map[string]bool{}
			}
			budget.deferred[key] = true
		}
		known[key] = updated
		if refreshed[key] {
			if updated.Identity.Source == domain.SourceCodex && fingerprintErr == nil && r.SourceFingerprint != fingerprint {
				if err := s.reopenCodexChildSweep(ctx, updated.Identity); err != nil {
					return nil, nil, err
				}
			}
			for _, transcript := range read.transcripts {
				if err := s.queueCodexSourceStarts(ctx, updated, transcript.ChildDiscoveryHints(), known); err != nil {
					return nil, nil, err
				}
			}
			acceptedRelocation = acceptedRelocation || relocating
			source.transcripts = append(source.transcripts, read.transcripts...)
			source.events = domain.MergeSourceEvents(source.events, read.events)
		}
		registrations[i] = updated
		if priorSourceBytes > 0 && (errors.Is(extractErr, opencode.ErrCaptureBudget) || errors.Is(extractErr, codex.ErrDiscoveryBudget)) {
			// Retry the first owner starved by earlier work before that work on
			// the next pass. Keep this continuation across refresh-class calls.
			// A source that fails with the full pass allowance still advances,
			// so a permanently oversized owner cannot pin the rotation.
			budget.refreshCursorHeld = true
		}
		if !budget.refreshCursorHeld {
			if budget.restorePriority == nil {
				_, err = s.db.ExecContext(ctx, `INSERT INTO reconcile_cursor VALUES(?,?,?) ON CONFLICT(worktree_id,ticket) DO UPDATE SET identity=excluded.identity`, s.worktreeID, ticket, key)
			} else {
				err = s.saveChildCursor(ctx, restoreCursorKind, ticket, key, false)
			}
			if err != nil {
				return nil, nil, err
			}
		}
	}
	if len(refreshed) > 0 {
		// A source read may be superseded after its primary snapshot commits.
		// Establish one registry view for accepting all inputs and saving their
		// per-owner and family-derived annotations.
		err = s.withCaptureControlRegistrationWrite(ctx, budget.controlRevision, func(conn *sql.Conn) error {
			current, readErr := readRegistrationRows(conn.QueryContext(ctx, `SELECT record FROM registrations WHERE worktree_id=? ORDER BY identity`, s.worktreeID))
			if readErr != nil {
				return readErr
			}
			var discarded bool
			registrations, discarded = keepCurrentCaptureReads(registrations, current, refreshed, &source)
			if discarded {
				budget.refreshCursorHeld = true
			}
			unresolved, annotationErr := s.finishCaptureAnnotations(ctx, ticket, registrations, previousSnapshots, refreshed, &source, func(ctx context.Context, identity domain.CaptureIdentity, apply func(domain.CaptureRegistration) (domain.CaptureRegistration, error)) (domain.CaptureRegistration, error) {
				if err := identity.Validate(); err != nil {
					return domain.CaptureRegistration{}, err
				}
				id, err := canonicalIdentity(identity)
				if err != nil {
					return domain.CaptureRegistration{}, err
				}
				return s.changeRegistration(ctx, conn, id, apply)
			}, policies)
			if annotationErr == nil {
				result.unresolvedChildLinks = append(result.unresolvedChildLinks, unresolved...)
			}
			return annotationErr
		})
		if err != nil {
			return nil, nil, err
		}
	}
	for _, row := range registrations {
		key := IdentityKey(row.Identity)
		if archiveAttempts[key] {
			archiveProjection[key] = row
		}
	}
	result.SourceEvents = domain.MergeSourceEvents(result.SourceEvents, source.events)
	if acceptedRelocation && len(source.transcripts) > 0 {
		result.recordArchiveReadGroup(registrations, refreshed, source.transcripts, previousSnapshots, *source.policy)
	}
	return registrations, refreshed, nil
}

func keepCurrentCaptureReads(registrations, current []domain.CaptureRegistration, refreshed map[string]bool, source *sourceRead) ([]domain.CaptureRegistration, bool) {
	byKey := make(map[string]domain.CaptureRegistration, len(registrations))
	for _, row := range registrations {
		byKey[IdentityKey(row.Identity)] = row
	}
	stale := map[string]bool{}
	for i, latest := range current {
		key := IdentityKey(latest.Identity)
		observed, exists := byKey[key]
		delete(byKey, key)
		if exists && sameCaptureReadObservation(latest, observed) {
			if !refreshed[key] && observed.State == domain.CapturePending {
				current[i].State = observed.State
				current[i].Detail = observed.Detail
			}
		} else if refreshed[key] {
			delete(refreshed, key)
			stale[latest.Identity.SessionID] = true
			current[i] = pendingChangedObservation(latest)
		}
	}
	for key, missing := range byKey {
		if refreshed[key] {
			delete(refreshed, key)
			stale[missing.Identity.SessionID] = true
		}
	}
	if len(stale) == 0 {
		return current, false
	}
	remaining := source.transcripts[:0]
	for _, transcript := range source.transcripts {
		if !stale[transcript.ThreadID] && !stale[transcript.SessionID] {
			remaining = append(remaining, transcript)
		}
	}
	source.transcripts = remaining
	source.events = withoutCaptureEvidence(source.events, stale)
	return current, true
}

type captureAnnotationWriter func(context.Context, domain.CaptureIdentity, func(domain.CaptureRegistration) (domain.CaptureRegistration, error)) (domain.CaptureRegistration, error)

// finishCaptureAnnotations returns the accepted child links of refreshed
// owners that this read did not re-prove: those it kept and those it refused
// to place.
func (s *Store) finishCaptureAnnotations(ctx context.Context, ticket string, registrations []domain.CaptureRegistration, previous []domain.SessionFact, refreshed map[string]bool, source *sourceRead, write captureAnnotationWriter, policies []domain.GenerationPolicy) ([]domain.ChildLink, error) {
	// Native child metadata is request-local. Resolve only successfully read
	// owners before saving the final lean snapshots; budgets bound these reads.
	var unresolved []domain.ChildLink
	if len(source.transcripts) > 0 {
		seenTranscriptOwners := make(map[string]bool)
		for _, transcript := range source.transcripts {
			id := transcript.ThreadID
			if id == "" {
				id = transcript.SessionID
			}
			if seenTranscriptOwners[id] {
				return nil, fmt.Errorf("session identity collision: %s exists in multiple registered source stores; canonical ledger cannot distinguish these owners", id)
			}
			seenTranscriptOwners[id] = true
		}
		combined, combineErr := codex.CombineTranscriptsWithPolicy(source.transcripts, policies...)
		if combineErr != nil {
			return nil, combineErr
		}
		source.events = domain.MergeSourceEvents(source.events, domain.SourceEvents{Events: combined.Events, CheckpointEvidence: combined.CheckpointEvidence, CompleteSessionIDs: transcriptSessionIDs(source.transcripts)})
		owners, groupErr := combined.SessionFacts(policies...)
		if groupErr != nil {
			return nil, groupErr
		}
		// A bounded pass may refresh a parent without its children. Keep the
		// accepted links it cannot re-prove; transcript events stay as read.
		retention := domain.RetainAcceptedChildLinks(domain.ChildLinkInput{Harness: domain.SourceCodex, Accepted: previous, Fresh: owners})
		owners = retention.Sessions
		unresolved = retention.Report.UnresolvedLinks()
		byID := make(map[string]domain.SessionFact)
		for _, owner := range owners {
			byID[owner.ID] = owner
		}
		for i, registration := range registrations {
			if !refreshed[IdentityKey(registration.Identity)] {
				continue
			}
			owner, ok := byID[registration.Identity.SessionID]
			if !ok {
				continue
			}
			updated, changeErr := write(ctx, registration.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				for j := range current.Snapshot {
					if current.Snapshot[j].ID == owner.ID {
						current.Snapshot[j].Delegations = owner.Delegations
					}
				}
				return current, nil
			})
			if changeErr != nil {
				return nil, changeErr
			}
			registrations[i] = updated
		}
	}
	// Final summaries see all resulting owners and explicit complete source
	// coverage, including retained children missing from this bounded pass.
	all := []domain.SessionFact{}
	for _, registration := range registrations {
		if registration.TicketID == ticket {
			all = append(all, registration.Snapshot...)
		}
	}
	if source.policy == nil {
		return nil, errors.New("explicit generation policy is required for capture summaries")
	}
	effectivePolicy := *source.policy
	summary, summaryErr := domain.PersistSourceAnomalySummariesWithPolicy(domain.AgenticFactLedger{Sessions: all}, source.events, effectivePolicy, s.anomalyThresholds)
	if summaryErr != nil {
		return nil, summaryErr
	}
	summaries := make(map[string][]domain.Anomaly)
	for _, owner := range summary.Sessions {
		summaries[owner.ID] = owner.Anomalies
	}
	for i, registration := range registrations {
		if !refreshed[IdentityKey(registration.Identity)] {
			continue
		}
		updated, changeErr := write(ctx, registration.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			for j := range current.Snapshot {
				current.Snapshot[j].Anomalies = summaries[current.Snapshot[j].ID]
			}
			return current, nil
		})
		if changeErr != nil {
			return nil, changeErr
		}
		registrations[i] = updated
	}
	return unresolved, nil
}

func transcriptSessionIDs(transcripts []codex.Transcript) []string {
	result := make([]string, 0, len(transcripts))
	for _, source := range transcripts {
		id := source.ThreadID
		if id == "" {
			id = source.SessionID
		}
		result = append(result, id)
	}
	return result
}

func captureGenerationPolicy(policies []domain.GenerationPolicy) *domain.GenerationPolicy {
	if len(policies) == 0 {
		return nil
	}
	policy := policies[0]
	return &policy
}
