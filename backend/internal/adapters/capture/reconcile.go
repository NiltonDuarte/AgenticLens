package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

const DefaultByteBudget int64 = 64 * 1024 * 1024
const DefaultSessionBudget = 256

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
	SourceEvents          domain.SourceEvents          `json:"-"`
	DiagnosticIndex       *DiagnosticIndexProgress     `json:"diagnostic_index,omitempty"`
	StartedAt             string                       `json:"started_at"`
	CutoffSemantics       string                       `json:"cutoff_semantics"`
	SourceBytes           int64                        `json:"source_bytes_reserved"`
	SourceAttempts        int                          `json:"source_attempts"`
	RefreshedSessions     []domain.SessionFact         `json:"-"`
	DiscoveryPending      bool                         `json:"child_discovery_pending"`
	PendingChildren       int                          `json:"pending_children"`
	UnresolvedParentLinks int                          `json:"unresolved_parent_links"`
	Protocol              int                          `json:"protocol"`
	Cutoff                string                       `json:"cutoff"`
	TicketID              string                       `json:"ticket_id"`
	Registrations         []domain.CaptureRegistration `json:"registrations"`
	Sessions              []domain.SessionFact         `json:"-"`
	Limits                []string                     `json:"limits"`
}

func (s *Store) Reconcile(ctx context.Context, ticket, user string, sessionBudget int, byteBudget int64, policies ...domain.GenerationPolicy) (ReconcileResult, error) {
	return s.ReconcileAt(ctx, ticket, user, sessionBudget, byteBudget, time.Now().UTC(), policies...)
}

// ReconcileAt applies one duration cutoff and transient generation policy across source reads.
func (s *Store) ReconcileAt(ctx context.Context, ticket, user string, sessionBudget int, byteBudget int64, cutoff time.Time, policies ...domain.GenerationPolicy) (result ReconcileResult, returnErr error) {
	result = ReconcileResult{Protocol: RegistryProtocolVersion, StartedAt: cutoff.UTC().Format(time.RFC3339Nano), CutoffSemantics: "cutoff is reconciliation completion, not a common as-of source time; refreshed_at is each owner's successful read completion", TicketID: ticket, Limits: []string{"Only registered, explicitly bound or source-proven inherited sessions are selected. Sessions that never delivered a hook cannot be counted as covered.", "Hook installation is not proof of harness trust, agent response or complete source availability."}}
	if sessionBudget < 1 || byteBudget < 1 {
		return result, errors.New("capture budgets must be positive")
	}
	budget := &workBudget{bytes: byteBudget, sessions: sessionBudget}
	refreshed := map[string]bool{}
	defer func() {
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
	}()
	// All work classes rotate priority durably. The first eligible class can
	// use the full session AND byte budget; leftovers are shared in that order.
	// Thus no persistent failure can monopolize either resource across passes.
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS recovery_phase(worktree_id INTEGER NOT NULL,ticket TEXT,phase INTEGER NOT NULL,PRIMARY KEY(worktree_id,ticket))`); err != nil {
		return result, err
	}
	phase := 0
	phaseErr := s.db.QueryRowContext(ctx, `SELECT phase FROM recovery_phase WHERE worktree_id=? AND ticket=?`, s.worktreeID, ticket).Scan(&phase)
	if phaseErr != nil && !errors.Is(phaseErr, sql.ErrNoRows) {
		return result, phaseErr
	}
	phase = phase % int(workClassCount)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO recovery_phase VALUES(?,?,?)ON CONFLICT(worktree_id,ticket) DO UPDATE SET phase=excluded.phase`, s.worktreeID, ticket, (phase+1)%int(workClassCount)); err != nil {
		return result, err
	}
	attempted := map[string]bool{}
	refreshProjection := map[string]domain.CaptureRegistration{}
	for offset := 0; offset < int(workClassCount); offset++ {
		class := workClass((phase + offset) % int(workClassCount))
		if class == workRefresh {
			updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted, policies...)
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
	// immediately. Each owner is attempted at most once per pass.
	if budget.sessions > 0 && budget.bytes > 0 {
		updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted, policies...)
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
	for i, owner := range registrations {
		if projected, ok := refreshProjection[IdentityKey(owner.Identity)]; ok {
			registrations[i] = projected
		}
	}
	pending, discoveryPending, err := s.pendingChildWork(ctx, ticket, registrations)
	if err != nil {
		return result, err
	}
	result.PendingChildren = pending
	result.DiscoveryPending = budget.discoveryPending || discoveryPending

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
				if r.TicketID != parent.TicketID {
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
	for _, r := range registrations {
		knownParents[string(r.Identity.Source)+"\x00"+r.Identity.Store+"\x00"+r.Identity.SessionID] = true
	}
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
	return result, nil
}

// sourceRead never enters registry serialization or diagnostic publication.
type sourceRead struct {
	policy      *domain.GenerationPolicy
	transcripts []codex.Transcript
	events      domain.SourceEvents
}

func extractExact(ctx context.Context, r domain.CaptureRegistration, budget int64, user string, source ...*sourceRead) ([]domain.SessionFact, int64, string, error) {
	return extractExactAt(ctx, r, budget, user, time.Now().UTC(), source...)
}

func extractExactAt(ctx context.Context, r domain.CaptureRegistration, budget int64, user string, cutoff time.Time, source ...*sourceRead) ([]domain.SessionFact, int64, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, "", err
	}
	if budget < 1 {
		return nil, 0, "", opencode.ErrCaptureBudget
	}
	if r.Identity.Source == domain.SourceCodex {
		path := r.TranscriptPath
		if path == "" {
			return nil, 0, "", errors.New("exact Codex transcript path not registered; wait for an authoritative hook")
		}
		if err := codex.ValidateRolloutStore(r.Identity.Store, path); err != nil {
			return nil, 0, "", err
		}
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
		if len(source) > 0 {
			source[0].transcripts = append(source[0].transcripts, transcript)
		}
		var policies []domain.GenerationPolicy
		if len(source) > 0 && source[0].policy != nil {
			policies = append(policies, *source[0].policy)
		}
		sessions, err := transcript.SessionFacts(policies...)
		if err != nil {
			return nil, info.Size(), "", err
		}
		for i := range sessions {
			sessions[i].User = user
		}
		return sessions, info.Size(), transcript.UpdatedAt, nil
	}
	db, err := opencode.OpenDB(r.Identity.Store)
	if err != nil {
		return nil, 0, "", err
	}
	defer db.Close()
	var events domain.SourceEvents
	sessions, bytes, updated, readErr := opencode.CaptureSessionSnapshotWithPolicyAt(ctx, db.DB(), r.Identity.SessionID, user, budget, sourceReadPolicy(source), cutoff, &events)
	if readErr == nil && len(source) > 0 {
		source[0].events = domain.MergeSourceEvents(source[0].events, events)
	}
	return sessions, bytes, updated, readErr
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

// refreshTicketOwners is one scheduler work class; source acceptance and the
// existing owner continuation are unchanged.
func (s *Store) refreshTicketOwners(ctx context.Context, ticket, user string, budget *workBudget, result *ReconcileResult, attempted map[string]bool, policies ...domain.GenerationPolicy) ([]domain.CaptureRegistration, map[string]bool, error) {
	registrations, err := s.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	previousSnapshots := []domain.SessionFact{}
	for _, registration := range registrations {
		if registration.TicketID == ticket {
			previousSnapshots = append(previousSnapshots, registration.Snapshot...)
		}
	}
	if _, err = s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS reconcile_cursor(worktree_id INTEGER NOT NULL,ticket TEXT,identity TEXT NOT NULL,PRIMARY KEY(worktree_id,ticket))`); err != nil {
		return nil, nil, err
	}
	cursor := ""
	cursorErr := s.db.QueryRowContext(ctx, `SELECT identity FROM reconcile_cursor WHERE worktree_id=? AND ticket=?`, s.worktreeID, ticket).Scan(&cursor)
	if cursorErr != nil && !errors.Is(cursorErr, sql.ErrNoRows) {
		return nil, nil, cursorErr
	}
	indexes := []int{}
	for i, r := range registrations {
		if r.TicketID == ticket && !attempted[IdentityKey(r.Identity)] {
			indexes = append(indexes, i)
		}
	}
	// Persisted rotation prevents an early growing owner from starving later roots.
	split := 0
	for split < len(indexes) && IdentityKey(registrations[indexes[split]].Identity) <= cursor {
		split++
	}
	indexes = append(append([]int{}, indexes[split:]...), indexes[:split]...)

	source := sourceRead{policy: captureGenerationPolicy(policies)}
	used := 0
	refreshed := map[string]bool{}
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
		attempted[IdentityKey(r.Identity)] = true
		used++
		budget.sessions--
		budget.attempts++
		var fresh []domain.SessionFact
		var revision int64
		var updatedAt string
		var extractErr error
		if s.project != "" {
			reservation := int64(256)
			if r.Identity.Source == domain.SourceCodex {
				reservation = MaxPayloadBytes
				if info, e := os.Stat(r.TranscriptPath); e == nil {
					reservation = min(reservation, info.Size())
				}
			}
			if budget.bytes < reservation {
				extractErr = opencode.ErrCaptureBudget
			} else {
				budget.charge(reservation)
				extractErr = s.validateOwner(ctx, r)
			}
		}
		if extractErr == nil && r.Identity.Source == domain.SourceCodex {
			extractErr = codex.ValidateRolloutStore(r.Identity.Store, r.TranscriptPath)
		}
		fingerprint, fingerprintErr := sourceFingerprint(ctx, r)
		// Size/count/timestamp fingerprints are operational hints, not proof
		// that a restored complete source has identical recorded facts. Read
		// the exact owner within the existing session/byte budgets.

		if extractErr == nil {
			cutoff, parseErr := time.Parse(time.RFC3339Nano, result.StartedAt)
			if parseErr != nil {
				return nil, nil, parseErr
			}
			fresh, revision, updatedAt, extractErr = extractExactAt(ctx, r, budget.bytes, user, cutoff, &source)
		}
		budget.charge(revision)
		updated, updateErr := s.Change(ctx, r.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			if extractErr != nil {
				current.State = domain.CaptureFailed
				current.Detail = "Source snapshot unavailable or invalid; previous facts preserved. Run capture reconcile after source flush and inspect source configuration."
				if errors.Is(extractErr, opencode.ErrCaptureBudget) {
					current.State = domain.CapturePending
					current.Detail = "Source byte budget reached; increase --max-source-bytes or continue reconciliation; previous facts preserved"
				}
				return current, nil
			}
			if len(fresh) == 0 {
				current.State = domain.CapturePending
				current.Detail = "Source session has not flushed; previous facts preserved"
				return current, nil
			}

			owner := fresh[0]
			current.ParentID = ""
			if owner.ParentID != nil {
				current.ParentID = *owner.ParentID
			}
			// Lean auxiliary facts belong to the exact owner snapshot.
			compact, compactErr := domain.PersistAnomalySummaries(domain.AgenticFactLedger{Sessions: fresh})
			if compactErr != nil {
				return current, compactErr
			}
			current.Snapshot = domain.RetainIncompleteGroupSummaries(previousSnapshots, compact.Sessions)
			current.SnapshotNeedsRefresh = false
			current.ProjectDirectory = s.project
			current.ProjectProof = domain.CaptureProjectVerified
			if fingerprintErr == nil {
				current.SourceFingerprint = fingerprint
			}
			current.Revision = revision
			current.UpdatedAt = updatedAt
			current.State = domain.CaptureCurrent
			current.SourceInstallationID = current.InstallationID
			current.RefreshedAt = time.Now().UTC().Format(time.RFC3339Nano)
			refreshed[IdentityKey(current.Identity)] = true
			current.Detail = "Source snapshot reconciled; checkpoint does not establish ticket completion"
			if current.TicketID == "" {
				current.State = domain.CaptureUnassigned
				current.Detail = "Respond to the session binding request; no ticket inferred"
			}
			return current, nil
		})
		if updateErr != nil {
			return nil, nil, updateErr
		}
		registrations[i] = updated
		if _, err = s.db.ExecContext(ctx, `INSERT INTO reconcile_cursor VALUES(?,?,?) ON CONFLICT(worktree_id,ticket) DO UPDATE SET identity=excluded.identity`, s.worktreeID, ticket, IdentityKey(r.Identity)); err != nil {
			return nil, nil, err
		}
	}
	// Native child metadata is request-local. Resolve only successfully read
	// owners before saving the final lean snapshots; budgets bound these reads.
	if len(source.transcripts) > 0 {
		combined, combineErr := codex.CombineTranscriptsWithPolicy(source.transcripts, policies...)
		if combineErr != nil {
			return nil, nil, combineErr
		}
		source.events = domain.MergeSourceEvents(source.events, domain.SourceEvents{Events: combined.Events, CheckpointEvidence: combined.CheckpointEvidence, CompleteSessionIDs: transcriptSessionIDs(source.transcripts)})
		owners, groupErr := combined.SessionFacts(policies...)
		if groupErr != nil {
			return nil, nil, groupErr
		}
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
			updated, changeErr := s.Change(ctx, registration.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				for j := range current.Snapshot {
					if current.Snapshot[j].ID == owner.ID {
						current.Snapshot[j].Delegations = owner.Delegations
					}
				}
				return current, nil
			})
			if changeErr != nil {
				return nil, nil, changeErr
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
	effectivePolicy := domain.DefaultGenerationPolicy()
	if source.policy != nil {
		effectivePolicy = *source.policy
	}
	summary, summaryErr := domain.PersistSourceAnomalySummariesWithPolicy(domain.AgenticFactLedger{Sessions: all}, source.events, effectivePolicy)
	if summaryErr != nil {
		return nil, nil, summaryErr
	}
	summaries := make(map[string][]domain.Anomaly)
	for _, owner := range summary.Sessions {
		summaries[owner.ID] = owner.Anomalies
	}
	for i, registration := range registrations {
		if !refreshed[IdentityKey(registration.Identity)] {
			continue
		}
		updated, changeErr := s.Change(ctx, registration.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			for j := range current.Snapshot {
				current.Snapshot[j].Anomalies = summaries[current.Snapshot[j].ID]
			}
			return current, nil
		})
		if changeErr != nil {
			return nil, nil, changeErr
		}
		registrations[i] = updated
	}
	result.SourceEvents = domain.MergeSourceEvents(result.SourceEvents, source.events)
	return registrations, refreshed, nil
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

func sourceReadPolicy(sources []*sourceRead) *domain.GenerationPolicy {
	if len(sources) == 0 {
		return nil
	}
	return sources[0].policy
}
