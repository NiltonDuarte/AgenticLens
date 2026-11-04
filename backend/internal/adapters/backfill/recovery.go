package backfill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// sourceRecovery describes this command's bounded recovery, independently of
// whether the preserved ledger already satisfies its migration requirements.
type sourceRecovery struct {
	source      domain.SourceEvents
	sessions    []domain.SessionFact
	complete    bool
	updated     bool
	diagnostics []string
}

func storedTreeRoots(sessions []domain.SessionFact) []domain.SessionFact {
	ids := make(map[string]bool, len(sessions))
	for _, s := range sessions {
		ids[s.ID] = true
	}
	var roots []domain.SessionFact
	for _, s := range sessions {
		if s.ParentID == nil || !ids[*s.ParentID] {
			roots = append(roots, s)
		}
	}
	return roots
}

func (e *BackfillEngine) recoverCodex(stored []domain.SessionFact, home, user string) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	roots := append(storedTreeRoots(stored), stored...)
	covered := make(map[string]bool)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Codex owner tree unavailable or ambiguous")
	}
	for _, root := range roots {
		if covered[root.ID] {
			continue
		}
		var source domain.SourceEvents
		fresh, diagnostics, err := e.readCodexOwners(root.ID, home, &source)
		if len(diagnostics) > 0 {
			out.complete = false
			out.diagnostics = append(out.diagnostics, diagnostics...)
		}

		if err != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex tree %s retained: %v", root.ID, err))
			continue
		}
		accepted := []domain.SessionFact{}
		previous := make(map[string]domain.SessionFact)
		for _, owner := range stored {
			previous[owner.ID] = owner
		}
		for _, owner := range fresh {
			if old, known := previous[owner.ID]; known && ((old.ParentID == nil) != (owner.ParentID == nil) || old.ParentID != nil && owner.ParentID != nil && *old.ParentID != *owner.ParentID) {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex owner %s retained: source parent linkage changed", owner.ID))
				continue
			}
			accepted = append(accepted, owner)
		}
		fresh = accepted
		if len(fresh) == 0 {
			continue
		}
		acceptedIDs := make(map[string]bool, len(fresh))
		for _, owner := range fresh {
			acceptedIDs[owner.ID] = true
		}
		acceptedSource := domain.SourceEvents{}
		for _, id := range source.CompleteSessionIDs {
			if acceptedIDs[id] {
				acceptedSource.CompleteSessionIDs = append(acceptedSource.CompleteSessionIDs, id)
			}
		}
		for _, event := range source.Events {
			if acceptedIDs[event.SessionID] {
				acceptedSource.Events = append(acceptedSource.Events, event)
			}
		}
		out.source = domain.MergeSourceEvents(out.source, acceptedSource)
		resolveCodexSessionUsers(fresh, stored, user)
		replaceRecoveredOwners(&out.sessions, fresh)
		for _, session := range fresh {
			covered[session.ID] = true
		}
		out.updated = true
	}
	for _, old := range stored {
		if !covered[old.ID] {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex stored session %s not recovered", old.ID))
		}
	}
	return out
}

func (e *BackfillEngine) readCodexOwners(root, home string, source ...*domain.SourceEvents) ([]domain.SessionFact, []string, error) {
	diagnostics := []string{}
	paths, err := e.discoverCodexRollouts(codex.DiscoveryOptions{CodexHome: home, SessionID: root})
	if err != nil {
		return nil, diagnostics, fmt.Errorf("discovering source: %w", err)
	}
	if len(paths) == 0 {
		return nil, diagnostics, fmt.Errorf("source rollout unavailable")
	}
	transcripts := make([]codex.Transcript, 0, len(paths))
	for _, p := range paths {
		t, err := e.parseCodexRollout(p)
		if err != nil {
			diagnostics = append(diagnostics, fmt.Sprintf("Codex source owner retained after parse failure: %v", err))
			continue
		}
		if len(t.Warnings) > 0 {
			diagnostics = append(diagnostics, fmt.Sprintf("Codex owner %s is incomplete; previous snapshot retained: %s", t.ThreadID, strings.Join(t.Warnings, "; ")))
			continue
		}
		transcripts = append(transcripts, t)
	}
	if len(transcripts) == 0 {
		return nil, diagnostics, fmt.Errorf("no complete Codex owners available")
	}
	combined, err := e.combineCodex(transcripts)
	if err != nil {
		return nil, diagnostics, fmt.Errorf("grouping source: %w", err)
	}
	fresh, err := combined.SessionFacts()
	if err != nil {
		return nil, diagnostics, fmt.Errorf("extracting source: %w", err)
	}
	if len(source) > 0 {
		events := domain.SourceEvents{Events: combined.Events}
		for _, owner := range fresh {
			events.CompleteSessionIDs = append(events.CompleteSessionIDs, owner.ID)
		}
		*source[0] = domain.MergeSourceEvents(*source[0], events)
	}
	return fresh, diagnostics, nil
}

// openCodeSessionTable selects only known table names. Canonical extraction
// validates the complete schema family before any candidate is published.
func openCodeSessionTable(db *sql.DB) (string, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='session_v2'").Scan(&count); err != nil {
		return "", err
	}
	if count > 0 {
		return "session_v2", nil
	}
	return "session", nil
}

func discoverOpenCodeTreeIDs(db *sql.DB, table, root string) (ids []string, resultErr error) {
	rows, err := db.Query("WITH RECURSIVE tree(id) AS (SELECT id FROM "+table+" WHERE id=? UNION SELECT s.id FROM "+table+" s JOIN tree t ON s.parent_id=t.id) SELECT id FROM tree ORDER BY id", root)
	if err != nil {
		return nil, fmt.Errorf("discovering source descendants: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, rows.Close()) }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning source descendant: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating source descendants: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("source tree unavailable")
	}
	return ids, nil
}

func readOpenCodeTree(db *sql.DB, table, root, user string) ([]domain.SessionFact, error) {
	ids, err := discoverOpenCodeTreeIDs(db, table, root)
	if err != nil {
		return nil, err
	}
	fresh, _, _, err := opencode.ExtractSessionsStrict(db, ids, user, nil)
	if err != nil {
		return nil, err
	}
	found := make(map[string]bool, len(fresh))
	for _, s := range fresh {
		found[s.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			return nil, fmt.Errorf("discovered session %s absent from extraction", id)
		}
	}
	return fresh, nil
}

func recoverOpenCode(db *sql.DB, stored []domain.SessionFact, user, ticket, workspace string, allowFallbacks bool) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	if db == nil {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "OpenCode source unavailable; stored trees retained")
		return out
	}
	table, err := openCodeSessionTable(db)
	if err != nil {
		out.complete = false
		out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode discovery failed: %v", err))
		return out
	}
	if len(stored) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "OpenCode owner history unavailable; no exact selected-source owners")
	}
	covered := make(map[string]bool)
	attempted := make(map[string]bool)
	previous := make(map[string]domain.SessionFact)
	for _, owner := range stored {
		previous[owner.ID] = owner
	}
	for _, root := range stored {
		ids, discoveryErr := discoverOpenCodeTreeIDs(db, table, root.ID)
		if discoveryErr != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s retained: %v", root.ID, discoveryErr))
			continue
		}
		for _, id := range ids {
			if attempted[id] {
				continue
			}
			attempted[id] = true
			var source domain.SourceEvents
			fresh, _, _, readErr := opencode.CaptureSessionSnapshot(context.Background(), db, id, user, math.MaxInt64, &source)
			if readErr != nil || len(fresh) != 1 || fresh[0].ID != id {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s previous snapshot retained after incomplete extraction: %v", id, readErr))
				continue
			}
			if old, known := previous[id]; known && ((old.ParentID == nil) != (fresh[0].ParentID == nil) || old.ParentID != nil && fresh[0].ParentID != nil && *old.ParentID != *fresh[0].ParentID) {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s retained: parent linkage changed", id))
				continue
			}
			out.source = domain.MergeSourceEvents(out.source, source)
			replaceRecoveredOwners(&out.sessions, fresh)
			covered[id] = true
			out.updated = true
		}
	}
	for _, old := range stored {
		if !covered[old.ID] {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode stored session %s not recovered", old.ID))
		}
	}
	// Heuristic discovery may enrich an unknown history only by explicit opt-in.
	// It cannot certify or replace unavailable exact owners or ownerless arrays.
	if !out.updated && allowFallbacks {
		sid, e := opencode.ResolveActiveRootSession(db, "", ticket, "")
		if e != nil || sid == nil || strings.TrimSpace(*sid) == "" {
			if min, max, found := git.GetWorkspaceCommitTimeRange(workspace); found {
				sid, e = opencode.ResolveSessionByWorkspacePath(db, ticket, (min-86400)*1000, (max+7200)*1000)
			}
		}
		if e == nil && sid != nil && strings.TrimSpace(*sid) != "" {
			fresh, e := readOpenCodeTree(db, table, *sid, user)
			if e == nil {
				// Reject overlap so fallback cannot replace unavailable exact stored facts.
				known := make(map[string]bool)
				for _, s := range stored {
					known[s.ID] = true
				}
				overlap := false
				for _, s := range fresh {
					if known[s.ID] {
						overlap = true
					}
				}
				if !overlap {
					out.sessions = append(out.sessions, fresh...)
					out.updated = true
					out.diagnostics = append(out.diagnostics, "heuristic OpenCode tree added; exact historical owner recovery remains incomplete")
				}
			} else {
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode fallback retained baseline: %v", e))
			}
		}
	}
	return out
}
