package backfill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
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
	reads       map[string]OwnerRefreshCoverage
	failures    map[string]error
	failure     error
	// notices report facts, such as replacements and newly added owners,
	// without marking the recovery incomplete.
	notices []string
}

func newSourceRecovery(stored []domain.SessionFact) sourceRecovery {
	return sourceRecovery{
		sessions: append([]domain.SessionFact(nil), stored...),
		complete: len(stored) > 0,
		reads:    make(map[string]OwnerRefreshCoverage),
		failures: make(map[string]error),
	}
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
	out := newSourceRecovery(stored)
	roots := append(storedTreeRoots(stored), stored...)
	previous := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		previous[owner.ID] = owner
	}
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Codex owner tree unavailable or ambiguous")
	}
	for _, root := range roots {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if _, covered := out.reads[root.ID]; covered {
			continue
		}
		var source domain.SourceEvents
		previousDiagnostics := len(out.diagnostics)
		fresh, err := e.readCodexOwners(ctx, root.ID, home, &source, &out)
		if len(out.diagnostics) > previousDiagnostics {
			out.complete = false
		}

		if err != nil {
			out.complete = false
			out.failures[root.ID] = err
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex tree %s retained: %v", root.ID, err))
			continue
		}
		accepted := []domain.SessionFact{}
		for _, owner := range fresh {
			if protected[owner.ID] {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex owner %s conflicts with another recorded harness; history retained", owner.ID))
				continue
			}
			if old, known := previous[owner.ID]; known && ((old.ParentID == nil) != (owner.ParentID == nil) || old.ParentID != nil && owner.ParentID != nil && *old.ParentID != *owner.ParentID) {
				out.complete = false
				out.failures[owner.ID] = errors.New("source parent linkage changed")
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
			out.accept(owner, previous[owner.ID])
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
		resolveSessionUsers(fresh, stored, user)
		replaceRecoveredOwners(&out.sessions, fresh)
		out.updated = true
	}
	for _, old := range stored {
		if _, covered := out.reads[old.ID]; !covered {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Codex stored session %s not recovered", old.ID))
		}
	}
	return out
}

func (e *BackfillEngine) readCodexOwners(ctx context.Context, root, home string, source *domain.SourceEvents, recovery *sourceRecovery) ([]domain.SessionFact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if e.budget == nil || e.budget.bytes < 1 || (e.budget.sessions < 1 && len(e.budget.codexSeen) == 0) {
		return nil, errSourceBudget
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
		if errors.Is(err, codex.ErrDiscoveryBudget) {
			return nil, errSourceBudget
		}
		return nil, fmt.Errorf("discovering source: %w", err)
	}
	if len(paths) == 0 {
		return nil, errOwnerUnavailable
	}
	transcripts := make([]codex.Transcript, 0, len(paths))
	var parseFailure error
	for _, p := range paths {
		t, err := e.budget.parseCodex(ctx, p, e.snapshotCutoff)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, errors.Join(ctxErr, err)
		}
		if err != nil {
			recovery.diagnostics = append(recovery.diagnostics, fmt.Sprintf("Codex source owner retained after parse failure: %v", err))
			parseFailure = err
			if errors.Is(err, errSourceBudget) {
				recovery.failure = err
			}
			continue
		}
		if len(t.Warnings) > 0 {
			parseFailure = fmt.Errorf("incomplete Codex source: %s", strings.Join(t.Warnings, "; "))
			recovery.failures[t.ThreadID] = parseFailure
			recovery.diagnostics = append(recovery.diagnostics, fmt.Sprintf("Codex owner %s is incomplete; previous snapshot retained: %s", t.ThreadID, strings.Join(t.Warnings, "; ")))
			continue
		}
		transcripts = append(transcripts, t)
	}
	if len(transcripts) == 0 {
		return nil, fmt.Errorf("no complete Codex owners available: %w", parseFailure)
	}
	combined, err := codex.CombineTranscriptsWithPolicyContext(ctx, transcripts)
	if err != nil {
		return nil, fmt.Errorf("grouping source: %w", err)
	}
	var policies []domain.GenerationPolicy
	if e.generationPolicy != nil {
		policies = append(policies, *e.generationPolicy)
	}
	fresh, err := combined.SessionFacts(policies...)
	if err := errors.Join(ctx.Err(), err); err != nil {
		return nil, fmt.Errorf("extracting source: %w", err)
	}
	events := domain.SourceEvents{Events: combined.Events}
	for _, owner := range fresh {
		events.CompleteSessionIDs = append(events.CompleteSessionIDs, owner.ID)
	}
	*source = domain.MergeSourceEvents(*source, events)
	return fresh, nil
}

// recoverClaude mirrors recoverCodex: it refreshes exact recorded trees, keeps
// owners whose parent link changed or whose source is unavailable, rejects
// conflicts with other harnesses and adds newly proven children. Each tree is
// read at most once per operation, so an unavailable tree is neither charged
// twice nor reported twice.
func (e *BackfillEngine) recoverClaude(ctx context.Context, stored []domain.SessionFact, home, user string, protected map[string]bool) sourceRecovery {
	out := newSourceRecovery(stored)
	roots := append(storedTreeRoots(stored), stored...)
	if len(roots) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "Claude Code owner tree unavailable or ambiguous")
	}
	previous := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		previous[owner.ID] = owner
	}
	attemptedOwners := make(map[string]bool)
	attemptedTrees := make(map[string]bool)
	for _, root := range roots {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if _, covered := out.reads[root.ID]; covered || attemptedOwners[root.ID] || storedAncestorAttempted(root, previous, attemptedOwners) {
			continue
		}
		attemptedOwners[root.ID] = true
		var source domain.SourceEvents
		read, err := e.readClaudeOwners(ctx, root.ID, home, attemptedTrees, &source)
		if errors.Is(err, errClaudeTreeAttempted) {
			continue
		}
		if len(read.diagnostics) > 0 {
			out.complete = false
			out.diagnostics = append(out.diagnostics, read.diagnostics...)
		}
		if err != nil {
			out.complete = false
			out.failures[root.ID] = err
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code tree %s retained: %v", root.ID, err))
			continue
		}
		fresh := read.sessions
		if _, recordedRoot := previous[read.rootID]; !recordedRoot {
			// Adopting an unrecorded root requires the same project ownership
			// check generate applies to a new root. New owners of a rejected
			// root are skipped; recorded owners in the tree still refresh.
			if err := capture.ValidateProject(e.projectRoot, read.workingDirectory); err != nil {
				out.complete = false
				out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code root %s not added: %v; its unrecorded owners are skipped", read.rootID, err))
				recordedOnly := fresh[:0:0]
				for _, owner := range fresh {
					if _, known := previous[owner.ID]; known {
						recordedOnly = append(recordedOnly, owner)
					}
				}
				fresh = recordedOnly
			}
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
				out.failures[owner.ID] = errors.New("source parent linkage changed")
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
			out.accept(owner, previous[owner.ID])
			if _, known := previous[owner.ID]; !known {
				out.notices = append(out.notices, fmt.Sprintf("Claude Code owner %s added from recorded tree %s", owner.ID, read.rootID))
			}
		}
		out.notices = append(out.notices, claude.ReplacementNotices(stored, accepted)...)
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
		resolveSessionUsers(accepted, stored, user)
		replaceRecoveredOwners(&out.sessions, accepted)
		out.updated = true
	}
	for _, old := range stored {
		if _, covered := out.reads[old.ID]; !covered {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("Claude Code stored session %s not recovered", old.ID))
		}
	}
	return out
}

// storedAncestorAttempted reports whether a recorded ancestor's tree was
// already attempted, so a stored child of an unrecoverable tree is not read or
// reported again.
func storedAncestorAttempted(owner domain.SessionFact, stored map[string]domain.SessionFact, attempted map[string]bool) bool {
	seen := map[string]bool{owner.ID: true}
	for parent := owner.ParentID; parent != nil && !seen[*parent]; {
		if attempted[*parent] {
			return true
		}
		seen[*parent] = true
		next, known := stored[*parent]
		if !known {
			return false
		}
		parent = next.ParentID
	}
	return false
}

var errClaudeTreeAttempted = errors.New("Claude Code tree already attempted in this operation")

type claudeTreeRead struct {
	rootID           string
	workingDirectory string
	sessions         []domain.SessionFact
	diagnostics      []string
}

// readClaudeOwners reads one tree selected by a recorded owner ID, which may be
// a root session ID or a subagent agent ID, charging the shared budget. A tree
// whose root was already attempted is not read again.
func (e *BackfillEngine) readClaudeOwners(ctx context.Context, owner, home string, attemptedTrees map[string]bool, source *domain.SourceEvents) (claudeTreeRead, error) {
	read := claudeTreeRead{diagnostics: []string{}}
	if err := ctx.Err(); err != nil {
		return read, err
	}
	budget, seen, err := e.budget.claudeDiscovery(ctx)
	if err != nil {
		return read, err
	}
	files, err := claude.DiscoverTree(claude.DiscoveryOptions{ClaudeHome: home, SessionID: owner, Budget: budget})
	var tree claude.Tree
	if err == nil {
		read.rootID = files.RootID
		if attemptedTrees[files.RootID] {
			e.budget.chargeClaude(budget, seen)
			return read, errClaudeTreeAttempted
		}
		attemptedTrees[files.RootID] = true
		tree, err = claude.ReadTree(ctx, files, budget)
	}
	e.budget.chargeClaude(budget, seen)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return read, errors.Join(ctxErr, err)
	}
	if errors.Is(err, claude.ErrDiscoveryBudget) {
		return read, errSourceBudget
	}
	if err != nil {
		return read, fmt.Errorf("discovering source: %w", err)
	}
	proven := tree.Prove()
	extraction, err := proven.Facts(e.snapshotCutoff, e.generationPolicy)
	if err != nil {
		return read, fmt.Errorf("extracting source: %w", err)
	}
	read.workingDirectory = extraction.WorkingDirectory
	read.diagnostics = append(read.diagnostics, extraction.Diagnostics...)
	if len(extraction.RootIncomplete) > 0 {
		read.diagnostics = append(read.diagnostics, fmt.Sprintf("Claude Code owner %s is incomplete; previous snapshot retained: %s", files.RootID, strings.Join(extraction.RootIncomplete, "; ")))
	}
	if len(extraction.Sessions) == 0 {
		return read, fmt.Errorf("no complete Claude Code owners available")
	}
	*source = domain.MergeSourceEvents(*source, domain.SourceEvents{Events: extraction.Events, CompleteSessionIDs: extraction.CompleteSessionIDs})
	read.sessions = extraction.Sessions
	return read, nil
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
		return nil, errOwnerUnavailable
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
	if len(fresh) == 0 && bytes == 0 {
		// A missing exact ID incurred only a metadata lookup, no owner snapshot.
		budget.sessions++
		return nil, errOwnerUnavailable
	}
	if len(fresh) != 1 || fresh[0].ID != id {
		return nil, fmt.Errorf("exact source owner %s unavailable", id)
	}
	return fresh, nil
}

func recoverOpenCodeAt(ctx context.Context, db *sql.DB, stored []domain.SessionFact, user, ticket string, workspacePaths []string, allowFallbacks bool, cutoff time.Time, budget *sourceReadBudget, policy *domain.GenerationPolicy, protected map[string]bool) sourceRecovery {
	out := newSourceRecovery(stored)
	if db == nil {
		out.complete = false
		out.failure = errors.New("OpenCode source unavailable")
		out.diagnostics = append(out.diagnostics, "OpenCode source unavailable; stored trees retained")
		return out
	}
	table, err := opencode.SessionTableContext(ctx, db)
	if err != nil {
		out.complete = false
		out.failure = err
		out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode discovery failed: %v", err))
		return out
	}
	if len(stored) == 0 {
		out.complete = false
		out.diagnostics = append(out.diagnostics, "OpenCode owner history unavailable; no exact selected-source owners")
	}
	visited := make(map[string]bool, len(stored))
	previous := make(map[string]domain.SessionFact, len(stored))
	for _, owner := range stored {
		previous[owner.ID] = owner
	}
	readOwner := func(id string) {
		visited[id] = true
		if protected[id] {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s conflicts with another recorded harness; history retained", id))
			return
		}
		var source domain.SourceEvents
		fresh, err := readOpenCodeOwner(ctx, db, id, user, cutoff, budget, policy, &source)
		if err != nil {
			out.complete = false
			out.failures[id] = err
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s previous snapshot retained after incomplete extraction: %v", id, err))
			return
		}
		old, known := previous[id]
		if known && !sameParent(old.ParentID, fresh[0].ParentID) {
			out.complete = false
			out.failures[id] = errors.New("source parent linkage changed")
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s retained: parent linkage changed", id))
			return
		}
		out.source = domain.MergeSourceEvents(out.source, source)
		out.accept(fresh[0], old)
		replaceRecoveredOwners(&out.sessions, fresh)
		out.updated = true
	}

	// Give exact recorded owners the available snapshot budget before looking
	// for unrecorded descendants, regardless of their native sort order.
	for _, owner := range stored {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		readOwner(owner.ID)
	}
	if budget == nil {
		return out
	}
	discovered := make(map[string]bool)
	for _, root := range append(storedTreeRoots(stored), stored...) {
		if ctx.Err() != nil {
			out.complete = false
			return out
		}
		if discovered[root.ID] {
			continue
		}
		discovered[root.ID] = true
		limit := budget.discoveryLimit(len(visited))
		ids, discoveryErr := discoverOpenCodeTreeIDs(ctx, db, table, root.ID, limit)
		if discoveryErr != nil {
			out.complete = false
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("OpenCode owner %s descendant discovery failed: %v", root.ID, discoveryErr))
			continue
		}
		for _, id := range ids {
			if len(ids) < limit {
				discovered[id] = true
			}
			if ctx.Err() != nil {
				out.complete = false
				return out
			}
			if visited[id] {
				continue
			}
			if budget.sessions < 1 || budget.bytes < 1 {
				out.complete = false
				out.failure = errSourceBudget
				break
			}
			readOwner(id)
		}
		if len(ids) == limit {
			out.complete = false
			out.failure = errSourceBudget
		}
	}
	if isSourceBudgetError(out.failure) {
		out.diagnostics = append(out.diagnostics, errSourceBudget.Error())
	}
	for _, old := range stored {
		if _, covered := out.reads[old.ID]; !covered {
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
			out.failure = err
			return out
		}
		sid, err := opencode.DiscoverRootSessionForRecoveryContext(ctx, db, ticket, workspacePaths, 0, 0)
		if err != nil {
			out.diagnostics = append(out.diagnostics, fmt.Sprintf("explicit heuristic recovery unavailable: %v", err))
			out.complete = false
			return out
		}
		if sid != nil {
			ids, err := discoverOpenCodeTreeIDs(ctx, db, table, *sid, budget.discoveryLimit(0))
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
				out.accept(fresh[0], domain.SessionFact{})
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

func sameParent(prior, fresh *string) bool {
	if prior == nil || fresh == nil {
		return prior == fresh
	}
	return *prior == *fresh
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
