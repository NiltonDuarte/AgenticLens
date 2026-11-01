// Package capture owns local lifecycle handoff and explicit session registration.
package capture

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"agenticlens/internal/domain"
	_ "modernc.org/sqlite"
)

const ProtocolVersion = 1
const DefaultDirectory = ".agenticlens-capture"
const MaxPayloadBytes = 64 * 1024

type Store struct{ db *sql.DB }
type storedRegistration struct {
	Registration domain.CaptureRegistration `json:"registration"`
	Path         string                     `json:"path"`
	Snapshot     []domain.SessionFact       `json:"snapshot,omitempty"`
	Skills       []domain.SkillFact         `json:"skills,omitempty"`
	TaskLinks    []domain.TaskLinkFact      `json:"task_links,omitempty"`
}

func Open(directory string) (*Store, error) {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, fmt.Errorf("creating local capture directory: %w", err)
	}
	p, err := filepath.Abs(filepath.Join(directory, "registry.db"))
	if err != nil {
		return nil, err
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(p)}
	db, err := sql.Open("sqlite", u.String()+"?_pragma=busy_timeout(1000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS registrations (identity TEXT PRIMARY KEY, source TEXT NOT NULL, store TEXT NOT NULL, session_id TEXT NOT NULL, ticket TEXT NOT NULL, record BLOB NOT NULL); CREATE TABLE IF NOT EXISTS capture_meta (version INTEGER NOT NULL); INSERT INTO capture_meta SELECT 1 WHERE NOT EXISTS(SELECT 1 FROM capture_meta);`); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	var version int
	if err = db.QueryRow(`SELECT version FROM capture_meta`).Scan(&version); err != nil || version != ProtocolVersion {
		return nil, errors.Join(fmt.Errorf("unsupported capture registry protocol %d: %w", version, err), db.Close())
	}
	return &Store{db: db}, nil
}
func (s *Store) Close() error { return s.db.Close() }
func IdentityKey(id domain.CaptureIdentity) string {
	b, err := json.Marshal(id)
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func encode(r domain.CaptureRegistration) ([]byte, error) {
	return json.Marshal(storedRegistration{r, r.TranscriptPath, r.Snapshot, r.Skills, r.TaskLinks})
}
func decode(b []byte) (domain.CaptureRegistration, error) {
	var v storedRegistration
	if err := json.Unmarshal(b, &v); err != nil {
		return domain.CaptureRegistration{}, err
	}
	r := v.Registration
	r.TranscriptPath = v.Path
	r.Snapshot = v.Snapshot
	r.Skills = v.Skills
	r.TaskLinks = v.TaskLinks
	return r, nil
}
func (s *Store) List(ctx context.Context) ([]domain.CaptureRegistration, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT record FROM registrations ORDER BY identity`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []domain.CaptureRegistration{}
	for rows.Next() {
		var b []byte
		if err = rows.Scan(&b); err != nil {
			return nil, err
		}
		r, e := decode(b)
		if e != nil {
			return nil, e
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func (s *Store) Change(ctx context.Context, id domain.CaptureIdentity, change func(domain.CaptureRegistration) (domain.CaptureRegistration, error)) (domain.CaptureRegistration, error) {
	if err := id.Validate(); err != nil {
		return domain.CaptureRegistration{}, err
	}
	var normalizeErr error
	id, normalizeErr = canonicalIdentity(id)
	if normalizeErr != nil {
		return domain.CaptureRegistration{}, normalizeErr
	}
	// A dedicated connection with BEGIN IMMEDIATE serializes read/modify/write across processes.
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return domain.CaptureRegistration{}, err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return domain.CaptureRegistration{}, err
	}
	committed := false
	defer func() {
		if !committed {
			_, rollbackErr := conn.ExecContext(context.Background(), `ROLLBACK`)
			if rollbackErr != nil {
				fmt.Fprintln(os.Stderr, "capture rollback failed:", rollbackErr)
			}
		}
	}()

	var b []byte
	r := domain.CaptureRegistration{Identity: id, State: domain.CaptureUnassigned}
	// Existing rows may predate path normalization. Resolve them inside the same
	// write transaction so an alias cannot create a second binding for one owner.
	rows, err := conn.QueryContext(ctx, `SELECT identity,record FROM registrations WHERE source=? AND session_id=?`, id.Source, id.SessionID)
	if err != nil {
		return r, err
	}
	matched := []string{}
	for rows.Next() {
		var key string
		var raw []byte
		if err = rows.Scan(&key, &raw); err != nil {
			return r, errors.Join(err, rows.Close())
		}
		candidate, e := decode(raw)
		if e != nil {
			return r, errors.Join(e, rows.Close())
		}
		same, e := SameStore(candidate.Identity.Store, id.Store)
		if e != nil {
			return r, errors.Join(e, rows.Close())
		}
		if same {
			matched = append(matched, key)
			r = candidate
		}
	}
	if err = errors.Join(rows.Err(), rows.Close()); err != nil {
		return r, err
	}
	if len(matched) > 1 {
		return r, errors.New("duplicate actual-store session registrations; explicit local repair required before binding or refresh")
	}
	r.Identity = id

	if r.Request == "" {
		var token [16]byte
		if _, err = rand.Read(token[:]); err != nil {
			return r, err
		}
		r.Request = hex.EncodeToString(token[:])
	}
	r, err = change(r)
	if err != nil {
		return r, err
	}
	b, err = encode(r)
	if err != nil {
		return r, err
	}
	for _, oldKey := range matched {
		if oldKey != IdentityKey(id) {
			if _, err = conn.ExecContext(ctx, `DELETE FROM registrations WHERE identity=?`, oldKey); err != nil {
				return r, err
			}
		}
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO registrations(identity,source,store,session_id,ticket,record) VALUES(?,?,?,?,?,?) ON CONFLICT(identity) DO UPDATE SET ticket=excluded.ticket,record=excluded.record`, IdentityKey(id), id.Source, id.Store, id.SessionID, r.TicketID, b)
	if err != nil {
		return r, err
	}
	if _, err = conn.ExecContext(ctx, `COMMIT`); err != nil {
		return r, err
	}
	committed = true
	return r, nil
}
func (s *Store) Register(ctx context.Context, id domain.CaptureIdentity, path, event string) (domain.CaptureRegistration, error) {
	return s.Change(ctx, id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if path != "" {
			if r.TranscriptPath != "" && r.TranscriptPath != path {
				return r, errors.New("conflicting transcript path for registered identity")
			}
			r.TranscriptPath = path
		}
		r.Event = event
		if r.TicketID != "" {
			r.State = domain.CapturePending
		}
		return r, nil
	})
}
func (s *Store) Bind(ctx context.Context, id domain.CaptureIdentity, ticket, request string) (domain.CaptureRegistration, error) {
	return s.Change(ctx, id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if request == "" || request != r.Request {
			return r, errors.New("binding requires the exact request token supplied by this session's hook or capture status")
		}
		bound, err := domain.CaptureBinding(r.TicketID, ticket, "")
		if err != nil {
			return r, err
		}
		if bound == "" {
			return r, errors.New("binding requires explicit ticket ID")
		}
		r.TicketID = bound
		r.Binding = "explicit_agent_response"
		r.State = domain.CapturePending
		return r, nil
	})
}
