package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_children(store TEXT,path TEXT,parent TEXT,PRIMARY KEY(store,path)); INSERT OR IGNORE INTO pending_children VALUES(?,?,?)`, store, path, parent)
	return err
}

type ReconcileResult struct {
	StartedAt             string                       `json:"started_at"`
	CutoffSemantics       string                       `json:"cutoff_semantics"`
	SourceBytes           int64                        `json:"source_bytes_reserved"`
	SourceAttempts        int                          `json:"source_attempts"`
	RefreshedSessions     []domain.SessionFact         `json:"-"`
	RefreshedSkills       []domain.SkillFact           `json:"-"`
	RefreshedTaskLinks    []domain.TaskLinkFact        `json:"-"`
	DiscoveryPending      bool                         `json:"child_discovery_pending"`
	PendingChildren       int                          `json:"pending_children"`
	UnresolvedParentLinks int                          `json:"unresolved_parent_links"`
	Protocol              int                          `json:"protocol"`
	Cutoff                string                       `json:"cutoff"`
	TicketID              string                       `json:"ticket_id"`
	Registrations         []domain.CaptureRegistration `json:"registrations"`
	Sessions              []domain.SessionFact         `json:"-"`
	Skills                []domain.SkillFact           `json:"-"`
	TaskLinks             []domain.TaskLinkFact        `json:"-"`
	Limits                []string                     `json:"limits"`
}

func (s *Store) Reconcile(ctx context.Context, ticket, user string, sessionBudget int, byteBudget int64) (result ReconcileResult, returnErr error) {
	result = ReconcileResult{Protocol: ProtocolVersion, StartedAt: time.Now().UTC().Format(time.RFC3339Nano), CutoffSemantics: "cutoff is reconciliation completion, not a common as-of source time; refreshed_at is each owner's successful read completion", TicketID: ticket, Limits: []string{"Only registered, explicitly bound or source-proven inherited sessions are selected. Sessions that never delivered a hook cannot be counted as covered.", "Hook installation is not proof of harness trust, agent response or complete source availability."}}
	if sessionBudget < 1 || byteBudget < 1 {
		return result, errors.New("capture budgets must be positive")
	}
	budget := &workBudget{bytes: byteBudget, sessions: sessionBudget}
	defer func() {
		result.Cutoff = time.Now().UTC().Format(time.RFC3339Nano)
		result.SourceBytes = budget.readBytes
		result.SourceAttempts = budget.attempts
	}()
	// All work classes rotate priority durably. The first eligible class can
	// use the full session AND byte budget; leftovers are shared in that order.
	// Thus no persistent failure can monopolize either resource across passes.
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS recovery_phase(ticket TEXT PRIMARY KEY,phase INTEGER NOT NULL)`); err != nil {
		return result, err
	}
	phase := 0
	phaseErr := s.db.QueryRowContext(ctx, `SELECT phase FROM recovery_phase WHERE ticket=?`, ticket).Scan(&phase)
	if phaseErr != nil && !errors.Is(phaseErr, sql.ErrNoRows) {
		return result, phaseErr
	}
	phase = phase % int(workClassCount)
	if _, err := s.db.ExecContext(ctx, `INSERT INTO recovery_phase VALUES(?,?)ON CONFLICT(ticket) DO UPDATE SET phase=excluded.phase`, ticket, (phase+1)%int(workClassCount)); err != nil {
		return result, err
	}
	refreshed := map[string]bool{}
	attempted := map[string]bool{}
	refreshProjection := map[string]domain.CaptureRegistration{}
	for offset := 0; offset < int(workClassCount); offset++ {
		class := workClass((phase + offset) % int(workClassCount))
		if class == workRefresh {
			updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted)
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
		updatedOwners, accepted, err := s.refreshTicketOwners(ctx, ticket, user, budget, &result, attempted)
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
		if r.TicketID == ticket {
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
				result.RefreshedSkills = append(result.RefreshedSkills, r.Skills...)
				result.RefreshedTaskLinks = append(result.RefreshedTaskLinks, r.TaskLinks...)
			}
			result.Skills = append(result.Skills, r.Skills...)
			result.TaskLinks = append(result.TaskLinks, r.TaskLinks...)
		}
	}
	return result, nil
}

func extractExact(ctx context.Context, r domain.CaptureRegistration, budget int64, user string) ([]domain.SessionFact, int64, string, error) {
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
		transcript, err := codex.ParseRollout(io.LimitReader(file, info.Size()))
		if err != nil {
			return nil, info.Size(), "", err
		}
		if len(transcript.Warnings) > 0 {
			return nil, info.Size(), "", errors.New("partial Codex source parse; previous snapshot must be preserved")
		}
		if transcript.ThreadID != r.Identity.SessionID {
			return nil, info.Size(), "", errors.New("hook/source identity mismatch")
		}
		sessions, err := transcript.SessionFacts()
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
	return opencode.CaptureSessionSnapshot(ctx, db.DB(), r.Identity.SessionID, user, budget)
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
	return fmt.Sprintf("\nCapture reconciliation completed: %s (per-owner refreshed_at checkpoints; no common as-of snapshot). Current: %d; pending: %d; failed: %d; unassigned: %d. See [capture_report.json](capture_report.json) for explicit coverage limits. Hook installation alone does not prove complete capture.\n", result.Cutoff, counts[domain.CaptureCurrent], counts[domain.CapturePending], counts[domain.CaptureFailed], counts[domain.CaptureUnassigned])
}

func redactedRegistrations(input []domain.CaptureRegistration) []domain.CaptureRegistration {
	output := append([]domain.CaptureRegistration{}, input...)
	for i := range output {
		output[i].Request = ""
		output[i].Identity.Store = "[local source store]"
		output[i].Detail = strings.TrimSpace(output[i].Detail)
	}
	return output
}
func ReportJSON(result ReconcileResult, digest string) ([]byte, error) {
	owned := []domain.CaptureRegistration{}
	for _, r := range result.Registrations {
		if r.TicketID == result.TicketID {
			owned = append(owned, r)
		}
	}
	result.Registrations = redactedRegistrations(owned)
	return json.MarshalIndent(struct {
		ReconcileResult
		LedgerSHA256 string `json:"ledger_sha256"`
	}{result, digest}, "", "  ")
}

// refreshTicketOwners is one scheduler work class; source acceptance and the
// existing owner continuation are unchanged.
func (s *Store) refreshTicketOwners(ctx context.Context, ticket, user string, budget *workBudget, result *ReconcileResult, attempted map[string]bool) ([]domain.CaptureRegistration, map[string]bool, error) {
	registrations, err := s.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	if _, err = s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS reconcile_cursor(ticket TEXT PRIMARY KEY,identity TEXT NOT NULL)`); err != nil {
		return nil, nil, err
	}
	cursor := ""
	cursorErr := s.db.QueryRowContext(ctx, `SELECT identity FROM reconcile_cursor WHERE ticket=?`, ticket).Scan(&cursor)
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
		fresh, revision, updatedAt, extractErr := extractExact(ctx, r, budget.bytes, user)
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
			if !domain.CaptureSnapshotAccept(current.Revision, revision, current.UpdatedAt, updatedAt) {
				current.State = domain.CapturePending
				current.Detail = "Older or truncated source snapshot rejected; previous facts preserved"
				return current, nil
			}
			owner := fresh[0]
			current.ParentID = ""
			if owner.ParentID != nil {
				current.ParentID = *owner.ParentID
			}
			if r.Identity.Source == domain.SourceOpenCode {
				current.Skills, current.TaskLinks = opencode.ExtractSkillTaskFacts(fresh)
			}
			compact, compactErr := domain.PersistAnomalySummaries(domain.AgenticFactLedger{Sessions: fresh})
			if compactErr != nil {
				return current, compactErr
			}
			current.Snapshot = compact.Sessions
			current.Revision = revision
			current.UpdatedAt = updatedAt
			current.State = domain.CaptureCurrent
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
		if _, err = s.db.ExecContext(ctx, `INSERT INTO reconcile_cursor VALUES(?,?) ON CONFLICT(ticket) DO UPDATE SET identity=excluded.identity`, ticket, IdentityKey(r.Identity)); err != nil {
			return nil, nil, err
		}
	}
	return registrations, refreshed, nil
}
