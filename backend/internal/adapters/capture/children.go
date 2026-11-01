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
	if agent == "" {
		return errors.New("SubagentStart requires its opaque agent_id")
	}
	canonicalStore, normalizeErr := CanonicalStore(store)
	if normalizeErr != nil {
		return normalizeErr
	}
	store = canonicalStore
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_starts(store TEXT,parent TEXT,agent TEXT,PRIMARY KEY(store,parent,agent)); INSERT OR IGNORE INTO pending_starts VALUES(?,?,?)`, store, parent, agent)
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
	bytes            int64
	sessions         int
	readBytes        int64
	attempts         int
	discoveryPending bool
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
func (s *Store) inheritChild(ctx context.Context, parent domain.CaptureRegistration, id, path string) error {
	identity := domain.CaptureIdentity{Source: parent.Identity.Source, Store: parent.Identity.Store, SessionID: id}
	_, err := s.Change(ctx, identity, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		bound, e := domain.CaptureBinding(r.TicketID, "", parent.TicketID)
		if e != nil {
			return r, e
		}
		if path != "" {
			if r.TranscriptPath != "" && r.TranscriptPath != path {
				return r, errors.New("conflicting child source path")
			}
			r.TranscriptPath = path
		}
		r.ParentID = parent.Identity.SessionID
		r.TicketID = bound
		if r.Binding == "" {
			r.Binding = "source_parent"
		}
		r.State = domain.CapturePending
		return r, nil
	})
	return err
}

// recoverChildren acts only for proven ticket parents. Unassigned roots are not
// extraction candidates and do not spend this ticket's source budget.
func (s *Store) recoverChildren(ctx context.Context, ticket string, budget *workBudget, class workClass) (int, error) {
	_, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pending_children(store TEXT,path TEXT,parent TEXT,PRIMARY KEY(store,path)); CREATE TABLE IF NOT EXISTS pending_starts(store TEXT,parent TEXT,agent TEXT,PRIMARY KEY(store,parent,agent)); CREATE TABLE IF NOT EXISTS discovery_cursor(parent TEXT PRIMARY KEY,path TEXT NOT NULL); CREATE TABLE IF NOT EXISTS child_cursor(kind TEXT,parent TEXT,key TEXT,more INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(kind,parent))`)
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
		table := "pending_children"
		if class == workStarts {
			table = "pending_starts"
		}
		var count int
		if err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE store=? AND parent=?`, parent.Identity.Store, parent.Identity.SessionID).Scan(&count); err != nil {
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
			db, e := opencode.OpenDB(parent.Identity.Store)
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
				if e = s.inheritChild(ctx, parent, id, ""); e != nil {
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
			rows, e := s.db.QueryContext(ctx, `SELECT path FROM pending_children WHERE store=? AND parent=? ORDER BY CASE WHEN path>? THEN 0 ELSE 1 END,path LIMIT ?`, parent.Identity.Store, parent.Identity.SessionID, queuedCursor, budget.sessions)
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
				if e = s.inheritChild(ctx, parent, transcript.ThreadID, path); e != nil {
					return 0, e
				}
				if _, e = s.db.ExecContext(ctx, `DELETE FROM pending_children WHERE store=? AND path=?`, parent.Identity.Store, path); e != nil {
					return 0, e
				}
			}
			continue
		}

		var starts int
		if e = s.db.QueryRowContext(ctx, `SELECT count(*) FROM pending_starts WHERE store=? AND parent=?`, parent.Identity.Store, parent.Identity.SessionID).Scan(&starts); e != nil {
			return 0, e
		}
		if starts == 0 || budget.sessions < 1 || budget.bytes < 1 {
			continue
		}
		cursor := ""
		e = s.db.QueryRowContext(ctx, `SELECT path FROM discovery_cursor WHERE parent=?`, IdentityKey(parent.Identity)).Scan(&cursor)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return 0, e
		}
		last := ""
		visited := 0
		maxEntries := max(16, budget.sessions*16)
		walkErr := filepath.WalkDir(parent.Identity.Store, func(path string, entry fs.DirEntry, walkErr error) error {
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
			if e = s.inheritChild(ctx, parent, transcript.ThreadID, path); e != nil {
				return e
			}
			// Equality is checked against authoritative metadata, never assumed from agent_id.
			_, e = s.db.ExecContext(ctx, `DELETE FROM pending_starts WHERE store=? AND parent=? AND agent=?`, parent.Identity.Store, parent.Identity.SessionID, transcript.ThreadID)
			return e
		})
		if walkErr != nil {
			return 0, walkErr
		}
		// A completed walk restarts from the beginning next pass to observe late source flushes.
		if visited <= maxEntries && budget.sessions > 0 && budget.bytes > 0 {
			last = ""
		}
		if _, e = s.db.ExecContext(ctx, `INSERT INTO discovery_cursor VALUES(?,?) ON CONFLICT(parent) DO UPDATE SET path=excluded.path`, IdentityKey(parent.Identity), last); e != nil {
			return 0, e
		}
	}
	pending, _, err := s.pendingChildWork(ctx, ticket, parents)
	return pending, err
}

func (s *Store) pendingChildWork(ctx context.Context, ticket string, parents []domain.CaptureRegistration) (int, bool, error) {
	pending := 0
	more := false
	for _, parent := range parents {
		if parent.TicketID != ticket {
			continue
		}
		var count int
		if err := s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM pending_children WHERE store=? AND parent=?)+(SELECT count(*) FROM pending_starts WHERE store=? AND parent=?)`, parent.Identity.Store, parent.Identity.SessionID, parent.Identity.Store, parent.Identity.SessionID).Scan(&count); err != nil {
			return 0, false, err
		}
		pending += count
		var unfinished int
		if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM child_cursor WHERE parent=? AND more<>0 AND kind='opencode'`, IdentityKey(parent.Identity)).Scan(&unfinished); err != nil {
			return 0, false, err
		}
		more = more || unfinished > 0
	}
	return pending, more || pending > 0, nil
}

func (s *Store) childCursor(ctx context.Context, kind, parent string) (string, error) {
	var key string
	err := s.db.QueryRowContext(ctx, `SELECT key FROM child_cursor WHERE kind=? AND parent=?`, kind, parent).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return key, err
}
func (s *Store) saveChildCursor(ctx context.Context, kind, parent, key string, more bool) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO child_cursor(kind,parent,key,more)VALUES(?,?,?,?) ON CONFLICT(kind,parent) DO UPDATE SET key=excluded.key,more=excluded.more`, kind, parent, key, more)
	return err
}
