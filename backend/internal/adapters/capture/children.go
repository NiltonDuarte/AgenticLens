package capture

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

func (s *Store) QueueChildStart(ctx context.Context, store, parent, agent string) error {
	return s.queueCodexChildStart(ctx, store, parent, agent, true)
}

func (s *Store) queueCodexChildStart(ctx context.Context, store, parent, agent string, nativeCallback bool) error {
	if agent == "" {
		return errors.New("SubagentStart requires its opaque agent_id")
	}
	canonicalStore, normalizeErr := CanonicalStore(store)
	if normalizeErr != nil {
		return normalizeErr
	}
	store = canonicalStore
	return s.withRegistrationWrite(ctx, func(conn *sql.Conn) error {
		inserted, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO pending_starts VALUES(?,?,?,?)`, s.worktreeID, store, parent, agent)
		if err != nil {
			return err
		}
		count, err := inserted.RowsAffected()
		if err != nil || (count == 0 && !nativeCallback) {
			return err
		}
		// New lifecycle evidence can request another bounded sweep.
		// Rereading the same accepted source hint cannot.
		identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: parent}
		_, err = conn.ExecContext(ctx, `DELETE FROM discovery_cursor WHERE worktree_id=? AND parent=? AND path=''`, s.worktreeID, IdentityKey(identity))
		return err
	})
}

// queueCodexSourceStarts records hints from an accepted native parent read.
// Already proven child edges require no further source discovery.
func (s *Store) queueCodexSourceStarts(ctx context.Context, parent domain.CaptureRegistration, hints []string, known map[string]domain.CaptureRegistration) error {
	for _, hint := range hints {
		id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: parent.Identity.Store, SessionID: hint}
		if existing := known[IdentityKey(id)]; existing.ProjectProof == domain.CaptureProjectVerified && existing.ParentID == parent.Identity.SessionID {
			continue
		}
		if err := s.queueCodexChildStart(ctx, parent.Identity.Store, parent.Identity.SessionID, hint, false); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) reopenCodexChildSweep(ctx context.Context, parent domain.CaptureIdentity) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM discovery_cursor WHERE worktree_id=? AND parent=? AND path=''`, s.worktreeID, IdentityKey(parent))
	return err
}

func (s *Store) reopenTicketCodexChildSweeps(ctx context.Context, ticket string) error {
	_, err := s.db.ExecContext(ctx, `
DELETE FROM discovery_cursor WHERE worktree_id=? AND path='' AND parent IN (
	SELECT identity FROM registrations WHERE worktree_id=? AND source='codex' AND ticket=?
)`, s.worktreeID, s.worktreeID, ticket)
	return err
}

// workClass is a durable priority position, independent of configured budgets.
// Every eligible class eventually receives the full pass allowance first.
type workClass int

const (
	workRefresh workClass = iota
	workStops
	workStarts
	workOpenCode
	workClassCount
)

func (class workClass) cursorKind() string {
	switch class {
	case workStops:
		return "parents-stops"
	case workStarts:
		return "parents-starts"
	case workOpenCode:
		return "parents-opencode"
	default:
		return "parents-refresh"
	}
}

type workBudget struct {
	codexDiscovery    *codex.DiscoveryBudget
	headers           *codex.HeaderCache
	split             *budgetSplit
	deferred          map[string]bool
	controlRevision   int64
	restorePriority   *restorePriorities
	openCodeAttempts  map[string]bool
	refreshCursorHeld bool
	bytes             int64
	sessions          int
	readBytes         int64
	attempts          int
	discoveryPending  bool
}

// budgetSplit reserves part of a pass's byte allowance for header
// identification during the rotating work-class rounds. Identification never
// exceeds its share. Owner extraction is limited to the rest until the split
// is released for the final leftover refresh, which may then use everything
// that remains, including identification's unused share.
type budgetSplit struct {
	identification int64
	extraction     int64
	released       bool
}

// newBudgetSplit reserves a quarter of the pass allowance for identification.
func newBudgetSplit(bytes int64) *budgetSplit {
	identification := bytes / 4
	return &budgetSplit{identification: identification, extraction: bytes - identification}
}

func (b *workBudget) charge(bytes int64) {
	if bytes > b.bytes {
		bytes = b.bytes
	}
	if bytes < 0 {
		bytes = 0
	}
	b.bytes -= bytes
	b.readBytes += bytes
}

// identificationRemaining is what header identification may still read.
func (b *workBudget) identificationRemaining() int64 {
	if b.split == nil {
		return b.bytes
	}
	return max(0, min(b.bytes, b.split.identification))
}

// extractionRemaining is what owner reads, including their own header
// preflight and exact lookup, may still read.
func (b *workBudget) extractionRemaining() int64 {
	if b.split == nil || b.split.released {
		return b.bytes
	}
	return max(0, min(b.bytes, b.split.extraction))
}

// extractionReserved reports that the identification reservation, not the
// pass total, is what limits owner extraction now.
func (b *workBudget) extractionReserved() bool {
	return b.split != nil && !b.split.released && b.extractionRemaining() < b.bytes
}

func (b *workBudget) chargeIdentification(bytes int64) {
	bytes = max(0, min(bytes, b.identificationRemaining()))
	if b.split != nil {
		b.split.identification -= bytes
	}
	b.charge(bytes)
}

func (b *workBudget) chargeExtraction(bytes int64) {
	bytes = max(0, min(bytes, b.extractionRemaining()))
	if b.split != nil && !b.split.released {
		b.split.extraction -= bytes
	}
	b.charge(bytes)
}

// releaseSplit lets the final leftover refresh extract with every remaining
// byte. Identification stays within its own share.
func (b *workBudget) releaseSplit() {
	if b.split != nil {
		b.split.released = true
	}
}

// readHeader reads one rollout header through the pass cache within limit
// bytes and reports the bytes actually read.
func (b *workBudget) readHeader(ctx context.Context, path string, limit int64) (codex.HeaderRead, error) {
	if b.headers == nil {
		b.headers = codex.NewHeaderCache()
	}
	return b.headers.ReadHeader(ctx, path, &codex.DiscoveryBudget{Context: ctx, MaxSourceBytes: max(0, limit)})
}

// readMetadata identifies a child candidate from its first session_meta
// record. It is charged the bytes actually read to the identification share.
// It consumes no owner-extraction session slot and so no source attempt,
// which counts those slots. A cached header is free.
func readMetadata(ctx context.Context, path string, budget *workBudget) (codex.RolloutHeader, error) {
	read, err := budget.readHeader(ctx, path, budget.identificationRemaining())
	budget.chargeIdentification(read.Bytes)
	if errors.Is(err, codex.ErrDiscoveryBudget) {
		return codex.RolloutHeader{}, opencode.ErrCaptureBudget
	}
	if err != nil {
		return codex.RolloutHeader{}, err
	}
	if !read.Header.First {
		return codex.RolloutHeader{}, errors.New("source does not start with session metadata")
	}
	return read.Header, nil
}

// childCreationWindow prunes a parent's child sweep by creation instant. The
// parent's own UUIDv7 decides; otherwise its session_meta timestamp, read
// through the pass cache, decides. Without either the sweep is not pruned.
func childCreationWindow(ctx context.Context, parent domain.CaptureRegistration, budget *workBudget) codex.CreationWindow {
	window := codex.NewCreationWindow(parent.Identity.SessionID, codex.RolloutHeader{})
	if window.Prunes() || parent.TranscriptPath == "" {
		return window
	}
	header, err := readMetadata(ctx, parent.TranscriptPath, budget)
	if err != nil || header.ID != parent.Identity.SessionID {
		return codex.CreationWindow{}
	}
	return codex.NewCreationWindow(parent.Identity.SessionID, header)
}

func (s *Store) inheritChild(ctx context.Context, parent domain.CaptureRegistration, id, path string) (bool, error) {
	identity := domain.CaptureIdentity{Source: parent.Identity.Source, Store: parent.Identity.Store, SessionID: id}
	if err := s.validateOwner(ctx, domain.CaptureRegistration{Identity: identity, TranscriptPath: path}); err != nil {
		return false, nil
	}
	_, err := s.Change(ctx, identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if r.TicketID != "" && (r.Binding == "explicit_agent_response" || r.Binding == "explicit_manual_selection") {
			// Explicit child attribution is authoritative even when its parent belongs
			// to another ticket. Keep that owner's accepted facts and let the parent
			// reconciliation proceed without forcing inheritance or retrying forever.
			return r, nil
		}
		bound, e := domain.CaptureBinding(r.TicketID, "", parent.TicketID)
		if e != nil {
			return r, e
		}
		if r.TicketID == bound && r.ParentID == parent.Identity.SessionID && r.ProjectDirectory == s.project &&
			r.ProjectProof == domain.CaptureProjectVerified && (path == "" || path == r.TranscriptPath) {
			// Repeated discovery of an accepted native edge is idempotent.
			// In particular it cannot undo a source refresh from this pass.
			return r, nil
		}
		if path != "" {
			if r.TranscriptPath != "" && r.TranscriptPath != path {
				if r.Identity.Source == domain.SourceCodex && r.TicketID != "" {
					if _, pathErr := os.Stat(r.TranscriptPath); errors.Is(pathErr, os.ErrNotExist) {
						// The bound owner's refresh proves uniqueness and
						// accepts a relocated path together with fresh facts.
						return r, nil
					}
				}
				return r, errors.New("conflicting child source path")
			}
			r.TranscriptPath = path
		}
		r.ParentID = parent.Identity.SessionID
		r.ProjectDirectory = s.project
		r.ProjectProof = domain.CaptureProjectVerified
		r.TicketID = bound
		if r.Binding == "" {
			r.Binding = "source_parent"
		}
		r.State = domain.CapturePending
		return r, nil
	})
	return err == nil, err
}

// recoverChildren uses the candidate parents for this reconciliation scope.
// Unassigned parents only spend the separate pre-binding source budget.
func (s *Store) recoverChildren(ctx context.Context, ticket string, budget *workBudget, class workClass) (int, error) {
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_children(worktree_id INTEGER NOT NULL,store TEXT,path TEXT,parent TEXT,PRIMARY KEY(worktree_id,store,path)); CREATE TABLE IF NOT EXISTS pending_starts(worktree_id INTEGER NOT NULL,store TEXT,parent TEXT,agent TEXT,PRIMARY KEY(worktree_id,store,parent,agent)); CREATE TABLE IF NOT EXISTS discovery_cursor(worktree_id INTEGER NOT NULL,parent TEXT,path TEXT NOT NULL,PRIMARY KEY(worktree_id,parent)); CREATE TABLE IF NOT EXISTS child_cursor(worktree_id INTEGER NOT NULL,kind TEXT,parent TEXT,key TEXT,more INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(worktree_id,kind,parent))`)
	if err != nil {
		return 0, err
	}
	parents, err := s.ListActive(ctx)
	if err != nil {
		return 0, err
	}
	cursorKind := class.cursorKind()
	if budget.restorePriority != nil {
		cursorKind = "restore-" + cursorKind
	}
	parentCursor, e := s.childCursor(ctx, cursorKind, ticket)
	if e != nil {
		return 0, e
	}

	eligible := []domain.CaptureRegistration{}
	for _, parent := range parents {
		if parent.TicketID != ticket {
			continue
		}
		if budget.restorePriority != nil && budget.restorePriority.discovery[IdentityKey(parent.Identity)] == 0 {
			continue
		}
		if class == workOpenCode {
			if parent.Identity.Source == domain.SourceOpenCode {
				eligible = append(eligible, parent)
			}
			continue
		}
		if parent.Identity.Source != domain.SourceCodex {
			continue
		}
		var count int
		if class == workStarts {
			err = s.db.QueryRowContext(ctx, `
SELECT count(*) FROM pending_starts WHERE worktree_id=? AND store=? AND parent=?
	AND NOT EXISTS (SELECT 1 FROM discovery_cursor WHERE worktree_id=? AND parent=? AND path='')
`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID, s.worktreeID, IdentityKey(parent.Identity)).Scan(&count)
		} else {
			err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM pending_children WHERE worktree_id=? AND store=? AND parent=?`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID).Scan(&count)
		}
		if err != nil {
			return 0, err
		}
		if count > 0 {
			eligible = append(eligible, parent)
		}
	}

	split := 0
	for split < len(eligible) && IdentityKey(eligible[split].Identity) <= parentCursor {
		split++
	}
	eligible = append(append([]domain.CaptureRegistration{}, eligible[split:]...), eligible[:split]...)
	if budget.restorePriority != nil {
		sort.SliceStable(eligible, func(i, j int) bool {
			return budget.restorePriority.discovery[IdentityKey(eligible[i].Identity)] > budget.restorePriority.discovery[IdentityKey(eligible[j].Identity)]
		})
	}
	parentLimit := max(1, budget.sessions)
	known := map[string]domain.CaptureRegistration{}
	for _, r := range parents {
		known[IdentityKey(r.Identity)] = r
	}
	for parentIndex, parent := range eligible {
		if budget.sessions < 1 {
			budget.discoveryPending = true
			break
		}
		if parentIndex >= parentLimit {
			budget.discoveryPending = true
			break
		}
		if err = s.saveChildCursor(ctx, cursorKind, ticket, IdentityKey(parent.Identity), parentIndex+1 < len(eligible)); err != nil {
			return 0, err
		}
		if parent.Identity.Source == domain.SourceOpenCode {
			// A bounded authoritative parent query discovers children without visiting unrelated roots.
			if budget.sessions < 1 || budget.identificationRemaining() < 256 {
				budget.discoveryPending = true
				continue
			}
			if budget.openCodeAttempts != nil {
				budget.openCodeAttempts[IdentityKey(parent.Identity)] = true
			}
			db, e := openOpenCodeSource(s.configuration, parent.Identity.Store)
			if e != nil {
				continue
			}
			after, e := s.childCursor(ctx, "opencode", IdentityKey(parent.Identity))
			if e != nil {
				return 0, errors.Join(e, db.Close())
			}
			children, more, e := opencode.CaptureChildPage(ctx, db.DB(), parent.Identity.SessionID, after, budget.sessions)
			closeErr := db.Close()
			if e != nil || closeErr != nil {
				continue
			}
			last := after
			for _, id := range children {
				if budget.sessions < 1 || budget.identificationRemaining() < 256 {
					more = true
					break
				}
				budget.sessions--
				budget.attempts++
				budget.chargeIdentification(256)
				last = id
				identity := domain.CaptureIdentity{Source: parent.Identity.Source, Store: parent.Identity.Store, SessionID: id}
				if existing, ok := known[IdentityKey(identity)]; ok && existing.TicketID == ticket && existing.ParentID == parent.Identity.SessionID {
					continue
				}
				if _, e = s.inheritChild(ctx, parent, id, ""); e != nil {
					return 0, e
				}
			}
			if !more {
				last = ""
			}
			if more {
				budget.discoveryPending = true
			}
			if e = s.saveOpenCodeChildCursor(ctx, IdentityKey(parent.Identity), last, more, budget.controlRevision); e != nil {
				return 0, e
			}
			continue
		}

		if class == workStops {
			queuedCursor, e := s.childCursor(ctx, "queued", IdentityKey(parent.Identity))
			if e != nil {
				return 0, e
			}
			rows, e := s.db.QueryContext(ctx, `SELECT path FROM pending_children WHERE worktree_id=? AND store=? AND parent=? ORDER BY CASE WHEN path>? THEN 0 ELSE 1 END,path LIMIT ?`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID, queuedCursor, budget.sessions)
			if e != nil {
				return 0, e
			}
			paths := []string{}
			for rows.Next() {
				var path string
				if e = rows.Scan(&path); e != nil {
					return 0, errors.Join(e, rows.Close())
				}
				paths = append(paths, path)
			}
			if e = errors.Join(rows.Err(), rows.Close()); e != nil {
				return 0, e
			}
			previous := queuedCursor
			for _, path := range paths {
				if e = s.saveChildCursor(ctx, "queued", IdentityKey(parent.Identity), path, true); e != nil {
					return 0, e
				}
				header, readErr := readMetadata(ctx, path, budget)
				if errors.Is(readErr, opencode.ErrCaptureBudget) {
					// Identification reached its share: this path comes first next pass.
					budget.discoveryPending = true
					if e = s.saveChildCursor(ctx, "queued", IdentityKey(parent.Identity), previous, true); e != nil {
						return 0, e
					}
					break
				}
				previous = path
				if readErr != nil {
					continue
				}
				if header.ID == "" || header.ParentID != parent.Identity.SessionID {
					continue
				}
				verified, e := s.inheritChild(ctx, parent, header.ID, path)
				if e != nil {
					return 0, e
				}
				if !verified {
					continue
				}
				if e = s.clearExactCodexStart(ctx, parent, header.ID); e != nil {
					return 0, e
				}
				if _, e = s.db.ExecContext(ctx, `DELETE FROM pending_children WHERE worktree_id=? AND store=? AND path=?`, s.worktreeID, parent.Identity.Store, path); e != nil {
					return 0, e
				}
			}
			continue
		}

		var starts int
		if e = s.db.QueryRowContext(ctx, `SELECT count(*) FROM pending_starts WHERE worktree_id=? AND store=? AND parent=?`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID).Scan(&starts); e != nil {
			return 0, e
		}
		if starts == 0 || budget.sessions < 1 || budget.identificationRemaining() < 1 {
			continue
		}
		cursor := ""
		e = s.db.QueryRowContext(ctx, `SELECT path FROM discovery_cursor WHERE worktree_id=? AND parent=?`, s.worktreeID, IdentityKey(parent.Identity)).Scan(&cursor)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return 0, e
		}
		if e == nil && cursor == "" {
			continue
		}
		// A child is created after its parent. Candidates created earlier are
		// neither claimed nor read; names never establish identity.
		window := childCreationWindow(ctx, parent, budget)
		last := ""
		visited := 0
		exhausted := false
		maxEntries := max(16, budget.sessions*16)
		walkErr := walkCodexCaptureRollouts(parent.Identity.Store, func(path string, entry fs.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}

			if walkErr != nil {
				return nil
			}
			if path <= cursor {
				if entry.IsDir() && path < cursor && !strings.HasPrefix(cursor, path+string(filepath.Separator)) {
					return fs.SkipDir
				}
				return nil
			}
			if entry.IsDir() && window.SkipsDirectory(parent.Identity.Store, path) {
				return fs.SkipDir
			}
			if !entry.IsDir() && window.SkipsFile(entry.Name()) {
				return nil
			}
			visited++
			if visited > maxEntries {
				return fs.SkipAll
			}
			if budget.identificationRemaining() < 1 {
				exhausted = true
				return fs.SkipAll
			}
			previous := last
			last = path
			if entry.IsDir() || !strings.HasSuffix(path, ".jsonl") {
				return nil
			}
			header, e := readMetadata(ctx, path, budget)
			if errors.Is(e, opencode.ErrCaptureBudget) {
				// Resume at this candidate once identification has bytes again.
				last = previous
				exhausted = true
				return fs.SkipAll
			}
			if e != nil {
				return nil
			}
			if header.ID == "" || header.ParentID != parent.Identity.SessionID {
				return nil
			}
			verified, e := s.inheritChild(ctx, parent, header.ID, path)
			if e != nil || !verified {
				return e
			}
			return s.clearExactCodexStart(ctx, parent, header.ID)
		})
		if walkErr != nil {
			return 0, walkErr
		}
		// A completed sweep remains closed until new evidence or explicit reconciliation.
		if visited <= maxEntries && !exhausted {
			last = ""
		}
		if exhausted {
			// Identification reached its share; discovery continues, not failed.
			budget.discoveryPending = true
		}
		if _, e = s.db.ExecContext(ctx, `INSERT INTO discovery_cursor VALUES(?,?,?) ON CONFLICT(worktree_id,parent) DO UPDATE SET path=excluded.path`, s.worktreeID, IdentityKey(parent.Identity), last); e != nil {
			return 0, e
		}
	}
	pending, _, err := s.pendingChildWork(ctx, ticket, parents)
	return pending, err
}

// walkCodexCaptureRollouts preserves cursor ordering while excluding package,
// config, and cache trees outside the selected native rollout directories.
func walkCodexCaptureRollouts(store string, visit fs.WalkDirFunc) error {
	for _, tree := range []string{"archived_sessions", "sessions"} {
		root := filepath.Join(store, tree)
		stopped := false
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			err := visit(path, entry, walkErr)
			stopped = errors.Is(err, fs.SkipAll)
			return err
		})
		if err != nil || stopped {
			return err
		}
	}
	return nil
}

// clearExactCodexStart matches a hint only after verifying the child's own
// authoritative session and parent identities.
func (s *Store) clearExactCodexStart(ctx context.Context, parent domain.CaptureRegistration, child string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM pending_starts WHERE worktree_id=? AND store=? AND parent=? AND agent=?`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID, child)
	return err
}

// CodexChildDiscoveryCursors reports unstarted and incomplete sweeps of
// parents. An unresolved hint after a complete sweep remains diagnostic
// evidence without keeping a terminal worker alive.
func (s *Store) CodexChildDiscoveryCursors(ctx context.Context) (map[string]string, error) {
	controls, err := s.captureControls(ctx, nil, "", CaptureControlDiscarded, false)
	if err != nil {
		return nil, err
	}
	var eligible map[string]bool
	if len(controls) != 0 {
		registrations, err := s.List(ctx)
		if err != nil {
			return nil, err
		}
		active, _ := activeCaptureRows(s.project, registrations, controls)
		eligible = make(map[string]bool, len(active))
		for _, row := range active {
			eligible[IdentityKey(row.Identity)] = true
		}
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT r.identity,coalesce(c.path,'')
FROM registrations r
LEFT JOIN discovery_cursor c ON c.worktree_id=r.worktree_id AND c.parent=r.identity
WHERE r.worktree_id=? AND r.source='codex' AND (c.path IS NULL OR c.path<>'')
  AND EXISTS (SELECT 1 FROM pending_starts p WHERE p.worktree_id=r.worktree_id AND p.store=r.store AND p.parent=r.session_id)
`, s.worktreeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cursors := make(map[string]string)
	for rows.Next() {
		var parent, cursor string
		if err := rows.Scan(&parent, &cursor); err != nil {
			return nil, err
		}
		if eligible != nil && !eligible[parent] {
			continue
		}
		cursors[parent] = cursor
	}
	return cursors, rows.Err()
}

func (s *Store) pendingChildWork(ctx context.Context, ticket string, parents []domain.CaptureRegistration) (int, bool, error) {
	pending := 0
	more := false
	for _, parent := range parents {
		if parent.TicketID != ticket {
			continue
		}
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM pending_children WHERE worktree_id=? AND store=? AND parent=?)+(SELECT count(*) FROM pending_starts WHERE worktree_id=? AND store=? AND parent=?)`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID).Scan(&count); err != nil {
			return 0, false, err
		}
		pending += count
		var unfinished int
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM child_cursor WHERE worktree_id=? AND parent=? AND more<>0 AND kind='opencode'`, s.worktreeID, IdentityKey(parent.Identity)).Scan(&unfinished); err != nil {
			return 0, false, err
		}
		more = more || unfinished > 0
	}
	return pending, more || pending > 0, nil
}

// pendingChildLinks returns the (parent, child) links already counted as
// pending child work for the ticket's parents: a queued start whose agent is
// the child, or a queued child path registered to the child.
func (s *Store) pendingChildLinks(ctx context.Context, ticket string, registrations []domain.CaptureRegistration) (map[domain.ChildLink]bool, error) {
	links := map[domain.ChildLink]bool{}
	byPath := make(map[string]string, len(registrations))
	for _, row := range registrations {
		if row.Identity.Source == domain.SourceCodex && row.TranscriptPath != "" {
			byPath[row.Identity.Store+"\x00"+row.TranscriptPath] = row.Identity.SessionID
		}
	}
	for _, parent := range registrations {
		if parent.TicketID != ticket || parent.Identity.Source != domain.SourceCodex {
			continue
		}
		rows, err := s.db.QueryContext(ctx, `SELECT agent,'' FROM pending_starts WHERE worktree_id=? AND store=? AND parent=? UNION ALL SELECT '',path FROM pending_children WHERE worktree_id=? AND store=? AND parent=?`, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID, s.worktreeID, parent.Identity.Store, parent.Identity.SessionID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var agent, path string
			if err := rows.Scan(&agent, &path); err != nil {
				return nil, errors.Join(err, rows.Close())
			}
			child := agent
			if path != "" {
				child = byPath[parent.Identity.Store+"\x00"+path]
			}
			if child != "" {
				links[domain.ChildLink{ParentID: parent.Identity.SessionID, ChildSessionID: child}] = true
			}
		}
		if err := errors.Join(rows.Err(), rows.Close()); err != nil {
			return nil, err
		}
	}
	return links, nil
}

func (s *Store) childCursor(ctx context.Context, kind, parent string) (string, error) {
	var key string
	err := s.db.QueryRowContext(ctx, `SELECT key FROM child_cursor WHERE worktree_id=? AND kind=? AND parent=?`, s.worktreeID, kind, parent).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return key, err
}
func (s *Store) saveChildCursor(ctx context.Context, kind, parent, key string, more bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO child_cursor(worktree_id,kind,parent,key,more)VALUES(?,?,?,?,?) ON CONFLICT(worktree_id,kind,parent) DO UPDATE SET key=excluded.key,more=excluded.more`, s.worktreeID, kind, parent, key, more)
	return err
}

func (s *Store) saveOpenCodeChildCursor(ctx context.Context, parent, key string, more bool, observedControlRevision int64) error {
	// Restore clears the previous scan's cursor inside this same fence. A
	// source query begun earlier cannot put that old continuation back.
	return s.withLifecycleLock(ctx, func() error {
		current, err := s.captureControlRevision(ctx)
		if err != nil {
			return err
		}
		if current != observedControlRevision {
			return ErrCaptureLifecycleChanged
		}
		return s.saveChildCursor(ctx, "opencode", parent, key, more)
	})
}
