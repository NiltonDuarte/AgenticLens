package capture

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	refreshCursorHeld bool
	bytes             int64
	sessions          int
	readBytes         int64
	attempts          int
	discoveryPending  bool
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

// readMetadata reads only the first bounded session_meta line. It cannot consume
// transcript bodies, and its reservation is charged even on malformed input.
func readMetadata(ctx context.Context, path string, budget *workBudget) (codex.Transcript, error) {
	if budget.sessions < 1 || budget.bytes < 1 {
		return codex.Transcript{}, opencode.ErrCaptureBudget
	}
	budget.sessions--
	budget.attempts++
	initialInfo, err := os.Stat(path)
	if err != nil {
		return codex.Transcript{}, err
	}
	if !initialInfo.Mode().IsRegular() {
		return codex.Transcript{}, errors.New("metadata source is not regular")
	}
	file, err := os.Open(path)
	if err != nil {
		return codex.Transcript{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return codex.Transcript{}, err
	}
	if !info.Mode().IsRegular() {
		return codex.Transcript{}, errors.New("metadata source is not regular")
	}
	reservation := min(info.Size(), int64(MaxPayloadBytes))
	if reservation > budget.bytes {
		return codex.Transcript{}, opencode.ErrCaptureBudget
	}
	budget.charge(reservation)
	if err = ctx.Err(); err != nil {
		return codex.Transcript{}, err
	}
	scanner := bufio.NewScanner(io.LimitReader(file, reservation))
	scanner.Buffer(make([]byte, 4096), MaxPayloadBytes)
	if !scanner.Scan() {
		return codex.Transcript{}, errors.Join(errors.New("source metadata unavailable"), scanner.Err())
	}
	var header struct {
		Type string `json:"type"`
	}
	if err = json.Unmarshal(scanner.Bytes(), &header); err != nil || header.Type != "session_meta" {
		return codex.Transcript{}, errors.New("source does not start with session metadata")
	}
	return codex.ParseRollout(strings.NewReader(scanner.Text() + "\n"))
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
	parents, err := s.List(ctx)
	if err != nil {
		return 0, err
	}
	parentCursor, e := s.childCursor(ctx, class.cursorKind(), ticket)
	if e != nil {
		return 0, e
	}

	eligible := []domain.CaptureRegistration{}
	for _, parent := range parents {
		if parent.TicketID != ticket {
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
		if err = s.saveChildCursor(ctx, class.cursorKind(), ticket, IdentityKey(parent.Identity), parentIndex+1 < len(eligible)); err != nil {
			return 0, err
		}
		if parent.Identity.Source == domain.SourceOpenCode {
			// A bounded authoritative parent query discovers children without visiting unrelated roots.
			if budget.sessions < 1 || budget.bytes < 256 {
				budget.discoveryPending = true
				continue
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
				if budget.sessions < 1 || budget.bytes < 256 {
					more = true
					break
				}
				budget.sessions--
				budget.attempts++
				budget.charge(256)
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
			if e = s.saveChildCursor(ctx, "opencode", IdentityKey(parent.Identity), last, more); e != nil {
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
			for _, path := range paths {
				if e = s.saveChildCursor(ctx, "queued", IdentityKey(parent.Identity), path, true); e != nil {
					return 0, e
				}
				transcript, readErr := readMetadata(ctx, path, budget)
				if readErr != nil {
					continue
				}
				if transcript.ThreadID == "" || transcript.ParentThread != parent.Identity.SessionID {
					continue
				}
				verified, e := s.inheritChild(ctx, parent, transcript.ThreadID, path)
				if e != nil {
					return 0, e
				}
				if !verified {
					continue
				}
				if e = s.clearExactCodexStart(ctx, parent, transcript.ThreadID); e != nil {
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
		if starts == 0 || budget.sessions < 1 || budget.bytes < 1 {
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
		last := ""
		visited := 0
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
			visited++
			if visited > maxEntries || budget.sessions < 1 || budget.bytes < 1 {
				return fs.SkipAll
			}
			last = path
			if entry.IsDir() || !strings.HasSuffix(path, ".jsonl") {
				return nil
			}
			transcript, e := readMetadata(ctx, path, budget)
			if e != nil {
				return nil
			}
			if transcript.ThreadID == "" || transcript.ParentThread != parent.Identity.SessionID {
				return nil
			}
			verified, e := s.inheritChild(ctx, parent, transcript.ThreadID, path)
			if e != nil || !verified {
				return e
			}
			return s.clearExactCodexStart(ctx, parent, transcript.ThreadID)
		})
		if walkErr != nil {
			return 0, walkErr
		}
		// A completed sweep remains closed until new evidence or explicit reconciliation.
		if visited <= maxEntries && budget.sessions > 0 && budget.bytes > 0 {
			last = ""
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
