package backfill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"agenticlens/internal/adapters/claude"
	"agenticlens/internal/adapters/codex"
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

func (e *BackfillEngine) recoverCodex(ctx context.Context, stored []domain.SessionFact, home, user string, protected map[string]bool) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	roots := append(storedTreeRoots(stored), stored...)
	covered := make(map[string]bool)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Codex owner tree unavailable or ambiguous")
	}
	for _, root := range roots {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if covered[root.ID] {
			continue
		}
		var source domain.SourceEvents
		fresh, diagnostics, err := e.readCodexOwners(ctx, root.ID, home, &source)
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
			if protected[owner.ID] {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex owner %s conflicts with another recorded harness; history retained", owner.ID))
				continue
			}
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

func (e *BackfillEngine) readCodexOwners(ctx context.Context, root, home string, source ...*domain.SourceEvents) ([]domain.SessionFact, []string, error) {
	diagnostics := []string{}
	if err := ctx.Err(); err != nil {
		return nil, diagnostics, err
	}
	if e.budget == nil || e.budget.bytes < 1 || (e.budget.sessions < 1 && len(e.budget.codexSeen) == 0) {
		return nil, diagnostics, errSourceBudget
	}
	if e.budget.codexSeen == nil {
		e.budget.codexSeen = make(map[string]bool)
	}
	seen := len(e.budget.codexSeen)
	budget := &codex.DiscoveryBudget{Context: ctx, MaxSessions: e.budget.sessions + seen, MaxSourceBytes: e.budget.bytes, SessionsRead: seen, SeenPaths: e.budget.codexSeen}
	paths, err := e.discoverCodexRollouts(codex.DiscoveryOptions{CodexHome: home, SessionID: root, Budget: budget})
	e.budget.sessions -= budget.SessionsRead - seen
	e.budget.bytes -= budget.SourceBytesRead
	if err != nil {
		return nil, diagnostics, fmt.Errorf("discovering source: %w", err)
	}
	if len(paths) == 0 {
		return nil, diagnostics, fmt.Errorf("source rollout unavailable")
	}
	transcripts := make([]codex.Transcript, 0, len(paths))
	for _, p := range paths {
		t, err := e.budget.parseCodex(ctx, p, e.snapshotCutoff)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, diagnostics, errors.Join(ctxErr, err)
		}
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
	combined, err := codex.CombineTranscriptsWithPolicyContext(ctx, transcripts)
	if err != nil {
		return nil, diagnostics, fmt.Errorf("grouping source: %w", err)
	}
	var policies []domain.GenerationPolicy
	if e.generationPolicy != nil {
		policies = append(policies, *e.generationPolicy)
	}
	fresh, err := combined.SessionFacts(policies...)
	if err := errors.Join(ctx.Err(), err); err != nil {
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

// recoverClaude mirrors recoverCodex: it refreshes exact recorded trees, keeps
// owners whose parent link changed or whose source is unavailable, rejects
// conflicts with other harnesses and adds newly proven children.
func (e *BackfillEngine) recoverClaude(ctx context.Context, stored []domain.SessionFact, home, user string, protected map[string]bool) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	roots := append(storedTreeRoots(stored), stored...)
	covered := make(map[string]bool)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Claude Code owner tree unavailable or ambiguous")
	}
	previous := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		previous[owner.ID] = owner
	}
	for _, root := range roots {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if covered[root.ID] {
			continue
		}
		var source domain.SourceEvents
		fresh, diagnostics, err := e.readClaudeOwners(ctx, root.ID, home, &source)
		if len(diagnostics) > 0 {
			out.complete = false
			out.diagnostics = append(out.diagnostics, diagnostics...)
		}
		if err != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code tree %s retained: %v", root.ID, err))
			continue
		}
		accepted := []domain.SessionFact{}
		for _, owner := range fresh {
			if protected[owner.ID] {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code owner %s conflicts with another recorded harness; history retained", owner.ID))
				continue
			}
			if old, known := previous[owner.ID]; known && ((old.ParentID == nil) != (owner.ParentID == nil) || old.ParentID != nil && owner.ParentID != nil && *old.ParentID != *owner.ParentID) {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code owner %s retained: source parent linkage changed", owner.ID))
				continue
			}
			accepted = append(accepted, owner)
		}
		if len(accepted) == 0 {
			continue
		}
		acceptedIDs := make(map[string]bool, len(accepted))
		for _, owner := range accepted {
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
		resolveCodexSessionUsers(accepted, stored, user)
		replaceRecoveredOwners(&out.sessions, accepted)
		for _, session := range accepted {
			covered[session.ID] = true
		}
		out.updated = true
	}
	for _, old := range stored {
		if !covered[old.ID] {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code stored session %s not recovered", old.ID))
		}
	}
	return out
}

// readClaudeOwners reads one tree selected by a recorded owner ID, which may be
// a root session ID or a subagent agent ID, charging the shared budget.
func (e *BackfillEngine) readClaudeOwners(ctx context.Context, owner, home string, source *domain.SourceEvents) ([]domain.SessionFact, []string, error) {
	diagnostics := []string{}
	if err := ctx.Err(); err != nil {
		return nil, diagnostics, err
	}
	budget, seen, err := e.budget.claudeDiscovery(ctx)
	if err != nil {
		return nil, diagnostics, err
	}
	files, err := claude.DiscoverTree(claude.DiscoveryOptions{ClaudeHome: home, SessionID: owner, Budget: budget})
	var tree claude.Tree
	if err == nil {
		tree, err = claude.ReadTree(ctx, files, budget)
	}
	e.budget.chargeClaude(budget, seen)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, diagnostics, errors.Join(ctxErr, err)
	}
	if errors.Is(err, claude.ErrDiscoveryBudget) {
		return nil, diagnostics, errSourceBudget
	}
	if err != nil {
		return nil, diagnostics, fmt.Errorf("discovering source: %w", err)
	}
	proven := tree.Prove()
	extraction, err := proven.Facts(e.snapshotCutoff, e.generationPolicy)
	if err != nil {
		return nil, diagnostics, fmt.Errorf("extracting source: %w", err)
	}
	diagnostics = append(diagnostics, extraction.Diagnostics...)
	if len(extraction.RootIncomplete) > 0 {
		diagnostics = append(diagnostics, fmt.Sprintf("Claude Code owner %s is incomplete; previous snapshot retained: %s", files.RootID, strings.Join(extraction.RootIncomplete, "; ")))
	}
	if len(extraction.Sessions) == 0 {
		return nil, diagnostics, fmt.Errorf("no complete Claude Code owners available")
	}
	*source = domain.MergeSourceEvents(*source, domain.SourceEvents{Events: extraction.Events, CompleteSessionIDs: extraction.CompleteSessionIDs})
	return extraction.Sessions, diagnostics, nil
}

func discoverOpenCodeTreeIDs(ctx context.Context, db *sql.DB, table, root string, limit int) (ids []string, resultErr error) {
	// Sorting the recursive output by id drains the tree under QueryContext
	// before the driver returns the first row.
	rows, err := db.QueryContext(ctx, "WITH RECURSIVE tree(id) AS (SELECT id FROM "+table+" WHERE id=? UNION SELECT s.id FROM "+table+" s JOIN tree t ON s.parent_id=t.id) SELECT id FROM tree ORDER BY id LIMIT ?", root, limit)
	if err != nil {
		return nil, fmt.Errorf("discovering source descendants: %w", err)
	}
	defer func() { resultErr = errors.Join(resultErr, rows.Close()) }()
	for ctx.Err() == nil && rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning source descendant: %w", err)
		}
		ids = append(ids, id)
	}
	if err := errors.Join(ctx.Err(), rows.Err()); err != nil {
		return nil, fmt.Errorf("iterating source descendants: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("source tree unavailable")
	}
	return ids, nil
}

func readOpenCodeOwner(ctx context.Context, db *sql.DB, id, user string, cutoff time.Time, budget *sourceReadBudget, policy *domain.GenerationPolicy, source *domain.SourceEvents) ([]domain.SessionFact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := budget.beginOwner(); err != nil {
		return nil, err
	}
	fresh, bytes, _, err := opencode.CaptureSessionSnapshotWithPolicyAt(ctx, db, id, user, budget.bytes, policy, cutoff, source)
	if err != nil {
		return nil, err
	}
	budget.bytes -= bytes
	if len(fresh) != 1 || fresh[0].ID != id {
		return nil, fmt.Errorf("exact source owner %s unavailable", id)
	}
	return fresh, nil
}

func recoverOpenCodeAt(ctx context.Context, db *sql.DB, stored []domain.SessionFact, user, ticket string, workspacePaths []string, allowFallbacks bool, cutoff time.Time, budget *sourceReadBudget, policy *domain.GenerationPolicy, protected map[string]bool) sourceRecovery {
	out := sourceRecovery{sessions: append([]domain.SessionFact(nil), stored...), complete: len(stored) > 0}
	if db == nil {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "OpenCode source unavailable; stored trees retained")
		return out
	}
	table, err := opencode.SessionTableContext(ctx, db)
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
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if budget == nil || budget.sessions < 1 || budget.bytes < 1 {
			out.complete = false
			out.diagnostics = append(out.diagnostics, errSourceBudget.Error())
			break
		}
		ids, discoveryErr := discoverOpenCodeTreeIDs(ctx, db, table, root.ID, budget.discoveryLimit())
		if discoveryErr != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s retained: %v", root.ID, discoveryErr))
			continue
		}
		for _, id := range ids {
			if ctx.Err() != nil {
				out.complete = false
				return out
			}
			if attempted[id] {
				continue
			}
			attempted[id] = true
			var source domain.SourceEvents
			fresh, readErr := readOpenCodeOwner(ctx, db, id, user, cutoff, budget, policy, &source)
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
	if ctx.Err() != nil {
		out.complete = false
		return out
	}
	// Heuristic discovery may enrich an unknown history only by explicit opt-in.
	// It cannot certify or replace unavailable exact owners or ownerless arrays.
	if allowFallbacks && !out.updated {

		if err := validateRecoveryDiscoveryBudget(ctx, db, table, budget); err != nil {
			out.diagnostics = append(out.diagnostics, err.Error())
			out.complete = false
			return out
		}
		sid, err := opencode.DiscoverRootSessionForRecoveryContext(ctx, db, ticket, workspacePaths, 0, 0)
		if err != nil {
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("explicit heuristic recovery unavailable: %v", err))
			out.complete = false
			return out
		}
		if sid != nil {
			ids, err := discoverOpenCodeTreeIDs(ctx, db, table, *sid, budget.discoveryLimit())
			if err != nil {
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("heuristic source tree unavailable: %v", err))
				return out
			}
			for _, id := range ids {
				if _, known := previous[id]; known || protected[id] {
					out.diagnostics = append(out.diagnostics, "heuristic source tree overlaps a known owner; recorded history retained")
					return out
				}
			}
			for _, id := range ids {
				var source domain.SourceEvents
				fresh, err := readOpenCodeOwner(ctx, db, id, user, cutoff, budget, policy, &source)
				if err != nil {
					out.diagnostics = append(out.diagnostics, fmt.Sprintf("heuristic source owner %s remains incomplete: %v", id, err))
					out.complete = false
					break
				}
				out.sessions = append(out.sessions, fresh...)
				out.source = domain.MergeSourceEvents(out.source, source)
				out.updated = true
			}
			if out.updated {
				out.diagnostics = append(out.diagnostics, "heuristic OpenCode owners added by explicit opt-in; exact historical owner recovery remains incomplete")
			} else {
				out.diagnostics = append(out.diagnostics, "no complete heuristic OpenCode owner available within configured recovery limits")
			}
			out.complete = false
		}
	}

	return out
}

func validateRecoveryDiscoveryBudget(ctx context.Context, db *sql.DB, table string, budget *sourceReadBudget) error {
	if budget == nil || budget.sessions < 1 || budget.bytes < 1 {
		return errSourceBudget
	}
	var sessions int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&sessions); err != nil {
		return err
	}
	if sessions > budget.sessions {
		return errSourceBudget
	}
	tables := []string{"message", "part"}
	if table == "session_v2" {
		tables = []string{"session_message"}
	}
	var bytes int64
	for _, records := range tables {
		var size int64
		if err := db.QueryRowContext(ctx, "SELECT coalesce(sum(length(CAST(data AS BLOB))),0) FROM "+records).Scan(&size); err != nil {
			return err
		}
		if size > budget.bytes-bytes {
			return errSourceBudget
		}
		bytes += size
	}
	budget.bytes -= bytes
	return nil
}
