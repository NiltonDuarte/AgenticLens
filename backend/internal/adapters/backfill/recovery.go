package backfill

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// sourceRecovery describes this command's bounded recovery, independently of
// whether the preserved ledger already satisfies its migration requirements.
type sourceRecovery struct {
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

func storedTree(sessions []domain.SessionFact, root string) []domain.SessionFact {
	ids := map[string]bool{root: true}
	for changed := true; changed; {
		changed = false
		for _, s := range sessions {
			if s.ParentID != nil && ids[*s.ParentID] && !ids[s.ID] {
				ids[s.ID] = true
				changed = true
			}
		}
	}
	var tree []domain.SessionFact
	for _, s := range sessions {
		if ids[s.ID] {
			tree = append(tree, s)
		}
	}
	return tree
}

func verifyRecoveredTree(expected, fresh []domain.SessionFact, root string) error {
	ids := make(map[string]domain.SessionFact, len(fresh))
	for _, s := range fresh {
		if _, ok := ids[s.ID]; ok {
			return fmt.Errorf("duplicate recovered session %s", s.ID)
		}
		ids[s.ID] = s
	}
	if _, ok := ids[root]; !ok {
		return fmt.Errorf("requested session %s absent from recovered tree", root)
	}
	for _, session := range fresh {
		current := session.ID
		seen := make(map[string]bool)
		for current != root {
			if seen[current] {
				return fmt.Errorf("cyclic recovered ancestry at %s", current)
			}
			seen[current] = true
			owner, ok := ids[current]
			if !ok || owner.ParentID == nil {
				return fmt.Errorf("session %s does not belong to requested tree %s", session.ID, root)
			}
			current = *owner.ParentID
		}
	}
	for _, old := range expected {
		got, ok := ids[old.ID]
		if !ok {
			return fmt.Errorf("stored child/session %s unavailable", old.ID)
		}
		if (old.ParentID == nil) != (got.ParentID == nil) || (old.ParentID != nil && got.ParentID != nil && *got.ParentID != *old.ParentID) {
			return fmt.Errorf("stored parent linkage changed for %s", old.ID)
		}
	}
	return nil
}

func (e *BackfillEngine) recoverCodex(stored []domain.SessionFact, home, user string) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	roots := storedTreeRoots(stored)
	covered := make(map[string]bool)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Codex owner tree unavailable or ambiguous")
	}
	for _, root := range roots {
		fresh, err := e.readCodexTree(root.ID, home)
		if err == nil {
			// A child-only persisted selection keeps its known external parent link.
			if root.ParentID != nil {
				for i := range fresh {
					if fresh[i].ID == root.ID && fresh[i].ParentID == nil {
						fresh[i].ParentID = root.ParentID
					}
				}
			}
			err = verifyRecoveredTree(storedTree(stored, root.ID), fresh, root.ID)
		}
		if err != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex tree %s retained: %v", root.ID, err))
			continue
		}
		resolveCodexSessionUsers(fresh, stored, user)
		replaceCodexTree(&out.sessions, root.ID, fresh)
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

func (e *BackfillEngine) readCodexTree(root, home string) ([]domain.SessionFact, error) {
	paths, err := e.discoverCodexRollouts(codex.DiscoveryOptions{CodexHome: home, SessionID: root})
	if err != nil {
		return nil, fmt.Errorf("discovering source: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("source rollout unavailable")
	}
	transcripts := make([]codex.Transcript, 0, len(paths))
	for _, p := range paths {
		t, err := e.parseCodexRollout(p)
		if err != nil {
			return nil, fmt.Errorf("parsing source: %w", err)
		}
		if len(t.Warnings) > 0 {
			return nil, fmt.Errorf("source %s is incomplete; retry recovery after source repair or completion: %s", p, strings.Join(t.Warnings, "; "))
		}
		transcripts = append(transcripts, t)
	}
	combined, err := e.combineCodex(transcripts)
	if err != nil {
		return nil, fmt.Errorf("grouping source: %w", err)
	}
	fresh, err := combined.SessionFacts()
	if err != nil {
		return nil, fmt.Errorf("extracting source: %w", err)
	}
	return fresh, nil
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

func exactOpenCodeRoot(db *sql.DB, table, id string) (string, error) {
	seen := make(map[string]bool)
	for {
		if seen[id] {
			return "", fmt.Errorf("cyclic source ancestry for %s", id)
		}
		seen[id] = true
		var parent sql.NullString
		if err := db.QueryRow("SELECT parent_id FROM "+table+" WHERE id=?", id).Scan(&parent); err != nil {
			return "", fmt.Errorf("resolving exact session %s: %w", id, err)
		}
		if !parent.Valid || strings.TrimSpace(parent.String) == "" {
			return id, nil
		}
		id = parent.String
	}
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
	roots := storedTreeRoots(stored)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "OpenCode owner history unavailable or ambiguous; empty owner set is not recovered history")
	}
	recovered := make(map[string]bool)
	covered := make(map[string]bool)
	for _, root := range roots {
		sourceRoot, err := exactOpenCodeRoot(db, table, root.ID)
		if err != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode tree %s retained: %v", root.ID, err))
			continue
		}
		if recovered[sourceRoot] {
			continue
		}
		fresh, err := readOpenCodeTree(db, table, sourceRoot, user)
		// Verify all stored rows the replacement would remove, including orphan
		// siblings attached to source-discovered ancestors absent from the ledger.
		var expected []domain.SessionFact
		removeIDs := replacementSessionIDs(stored, sourceRoot, fresh)
		for _, session := range stored {
			if _, ok := removeIDs[session.ID]; ok {
				expected = append(expected, session)
			}
		}

		if err == nil {
			err = verifyRecoveredTree(expected, fresh, sourceRoot)
		}
		if err != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode tree %s retained: %v", root.ID, err))
			continue
		}
		replaceCodexTree(&out.sessions, sourceRoot, fresh)
		recovered[sourceRoot] = true
		for _, session := range fresh {
			covered[session.ID] = true
		}
		out.updated = true
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
