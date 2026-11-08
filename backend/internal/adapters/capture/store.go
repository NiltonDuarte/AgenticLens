// Package capture owns local lifecycle handoff and explicit session registration.
package capture

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
	"modernc.org/sqlite"
)

const RegistryProtocolVersion = 2
const ManifestProtocolVersion = 2

// ProtocolVersion is the compatibility name for the persisted registry protocol.
// Installation manifest readers use ManifestProtocolVersion independently.
const ProtocolVersion = RegistryProtocolVersion

// HandlerProtocolVersion identifies native lifecycle flags independently of the
// persisted registry/manifest schema. Protocol 3 requires named shared storage
// with exact-worktree scope; protocol 2 used one directory per registry.
const HandlerProtocolVersion = 3

// DefaultDirectory names the legacy local base for explicit recovery callers.
// Operational configuration loading never supplies it as a fallback.
const DefaultDirectory = ".agenticlens-capture"
const MaxPayloadBytes = 64 * 1024

type Store struct {
	anomalyThresholds domain.AnomalyThresholdsConfig
	db                *sql.DB
	project           string
	directory         string
	registry          string
	worktreeID        int64
	relocationWorker  *PublicationLock
}
type storedRegistration struct {
	Registration domain.CaptureRegistration `json:"registration"`
	Path         string                     `json:"path"`
	Snapshot     []domain.SessionFact       `json:"snapshot,omitempty"`
}

// Open creates an isolated unassociated registry for adapter callers. Production
// lifecycle commands use OpenProject and require configured project ownership.
func Open(directory string) (*Store, error) { return openRegistry(directory, "", "", false) }

func openRegistry(directory, name, project string, readOnly bool) (*Store, error) {
	return openRegistryMode(directory, name, project, readOnly, "")
}
func openRegistryForRelocation(directory, name, project, migrationID string) (*Store, error) {
	if migrationID == "" {
		return nil, errors.New("relocation requires an explicit recorded migration identity")
	}
	return openRegistryMode(directory, name, project, false, migrationID)
}
func openRegistryMode(directory, name, project string, readOnly bool, migrationID string) (*Store, error) {
	return openRegistryModeContext(context.Background(), directory, name, project, readOnly, migrationID)
}
func openRegistryContext(ctx context.Context, directory, name, project string, readOnly bool) (*Store, error) {
	return openRegistryModeContext(ctx, directory, name, project, readOnly, "")
}

// Concurrent first opens can contend while SQLite switches the new database to
// WAL, before ordinary transaction busy handling applies. Retry only lock errors
// within the constructor's bounded allowance; all other errors remain explicit.
func openRegistryModeContext(ctx context.Context, directory, name, project string, readOnly bool, migrationID string) (*Store, error) {
	retryUntil := time.Now().Add(time.Second)
	for {
		store, err := openRegistryAttemptContext(ctx, directory, name, project, readOnly, migrationID)
		var sqliteErr *sqlite.Error
		if err == nil || readOnly || !errors.As(err, &sqliteErr) || sqliteErr.Code()&255 != 5 && sqliteErr.Code()&255 != 6 || !time.Now().Before(retryUntil) {
			return store, err
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(ctx.Err(), err)
		case <-timer.C:
		}
	}
}

func openRegistryAttemptContext(ctx context.Context, directory, name, project string, readOnly bool, migrationID string) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	allowPending := migrationID != ""
	path, err := filepath.Abs(filepath.Join(directory, "registry.db"))
	if err != nil {
		return nil, err
	}
	uri, err := sqliteFileURI(path)
	if err != nil {
		return nil, err
	}
	if !readOnly {
		if err = ctx.Err(); err != nil {
			return nil, err
		}
		if err = os.MkdirAll(directory, 0700); err != nil {
			return nil, err
		}
	}
	if readOnly {
		info, e := os.Stat(path)
		if e != nil {
			return nil, e
		}
		if !info.Mode().IsRegular() {
			return nil, errors.New("capture registry is not a regular file")
		}
	}
	busyMillis := constructorBusyMillis(ctx, 1000)
	options := fmt.Sprintf("?_pragma=busy_timeout(%d)&_txlock=immediate", busyMillis)
	if readOnly {
		options = fmt.Sprintf("?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(%d)", constructorBusyMillis(ctx, 250))
	}
	db, err := sql.Open("sqlite", uri+options)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	var relocationWorker *PublicationLock
	fail := func(e error) (*Store, error) {
		if relocationWorker != nil {
			e = errors.Join(e, relocationWorker.Close())
			relocationWorker = nil
		}
		return nil, errors.Join(e, db.Close())
	}
	// Observe both schema facts in one SQLite statement. Separate reads can
	// straddle another initializer's commit and mistake its valid new registry
	// for a foreign database with tables but no capture metadata.
	var existing, tables int
	if err = db.QueryRowContext(ctx, "SELECT count(CASE WHEN name='capture_meta' THEN 1 END), count(CASE WHEN type='table' THEN 1 END) FROM sqlite_master").Scan(&existing, &tables); err != nil {
		return fail(err)
	}
	if existing == 0 && tables > 0 {
		return fail(errors.New("unrecognized capture database; explicit recovery required"))
	}
	if existing > 0 {
		var version int
		if err = db.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
			return fail(err)
		}
		if version != RegistryProtocolVersion {
			return fail(fmt.Errorf("unsupported capture registry protocol %d; explicitly relocate legacy state", version))
		}
		if err = validateCoordinateSchema(ctx, db); err != nil {
			return fail(err)
		}
		var owner string
		if err = db.QueryRowContext(ctx, "SELECT name FROM project_identity WHERE singleton=1").Scan(&owner); err != nil {
			return fail(err)
		}
		if owner != name {
			return fail(errors.New("capture namespace belongs to a different configured project name"))
		}
		var scope int64
		e := db.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&scope)
		if e != nil && !errors.Is(e, sql.ErrNoRows) {
			return fail(e)
		}
		if e == nil {
			if e = validateWorktreeOwnershipContext(ctx, db, scope, project); e != nil {
				return fail(e)
			}
			if !allowPending {
				if e = validateScopeActivationContext(ctx, db, scope); e != nil {
					return fail(e)
				}
			} else if e = validateRelocationIdentity(db, scope, migrationID); e != nil {
				return fail(e)
			}
		}
	}
	var id int64
	coord := directory
	if readOnly {
		if err = db.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&id); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = os.ErrNotExist
			}
			return fail(err)
		}
	} else {
		if err = setConstructorBusyTimeout(ctx, db, 1000); err != nil {
			return fail(err)
		}
		if _, err = db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
			return fail(err)
		}

		if err = setConstructorBusyTimeout(ctx, db, 1000); err != nil {
			return fail(err)
		}
		tx, e := db.BeginTx(ctx, nil)
		if e != nil {
			return fail(e)
		}
		rollback := func(e error) (*Store, error) { return fail(errors.Join(e, tx.Rollback())) }
		ownership, e := currentWorktreeOwnership(project)
		if e != nil {
			return rollback(e)
		}
		if _, e = tx.ExecContext(ctx, registrySchema); e != nil {
			return rollback(e)
		}
		if _, e = tx.ExecContext(ctx, "INSERT OR IGNORE INTO project_identity VALUES(1,?)", name); e != nil {
			return rollback(e)
		}
		var owner string
		if e = tx.QueryRowContext(ctx, "SELECT name FROM project_identity WHERE singleton=1").Scan(&owner); e != nil {
			return rollback(e)
		}
		if owner != name {
			return rollback(errors.New("capture namespace belongs to a different configured project name"))
		}
		var previousScope int64
		scopeErr := tx.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&previousScope)
		scopeExisted := scopeErr == nil
		if scopeErr != nil && !errors.Is(scopeErr, sql.ErrNoRows) {
			return rollback(scopeErr)
		}
		if _, e = tx.ExecContext(ctx, "INSERT OR IGNORE INTO worktrees(directory)VALUES(?)", project); e != nil {
			return rollback(e)
		}
		if e = tx.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&id); e != nil {
			return rollback(e)
		}
		if project != "" {
			if scopeExisted {
				var directoryName string
				directoryName, e = coordinateDirectoryNameTx(ctx, tx, id)
				if e == nil {
					coord, e = coordinatePath(directory, directoryName)
				}
			} else {
				coord, e = allocateCoordinate(ctx, tx, directory, id, ownership, randomDirectoryName)
			}
			if e != nil {
				return rollback(e)
			}
		}
		if !scopeExisted {
			if _, e = tx.ExecContext(ctx, "INSERT INTO worktree_ownership(worktree_id,ownership)VALUES(?,?)", id, ownership); e != nil {
				return rollback(e)
			}
		}
		{
			var recorded string
			if e = tx.QueryRowContext(ctx, "SELECT ownership FROM worktree_ownership WHERE worktree_id=?", id).Scan(&recorded); e != nil {
				return rollback(e)
			}
			same, e := sameWorktreeOwnership(recorded, ownership)
			if e != nil {
				return rollback(e)
			}
			if !same {
				return rollback(errors.New("physical worktree or Git ownership changed; explicit recovery required"))
			}
			if recorded != ownership {
				if _, e = tx.ExecContext(ctx, "UPDATE worktree_ownership SET ownership=? WHERE worktree_id=? AND ownership=?", ownership, id, recorded); e != nil {
					return rollback(e)
				}
			}
			if allowPending {
				if e = os.MkdirAll(coord, 0700); e != nil {
					return rollback(e)
				}
				relocationWorker, e = AcquireWorker(coord)
				if e != nil {
					return rollback(e)
				}
				if relocationWorker == nil {
					return rollback(errors.New("destination worktree worker is active; retry relocation after its bounded pass finishes"))
				}
				if e = reserveRelocationScopeTx(tx, id, migrationID); e != nil {
					return rollback(e)
				}
			} else {
				var table int
				if e = tx.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='worktree_relocations'").Scan(&table); e != nil {
					return rollback(e)
				}
				if table > 0 {
					var state string
					e = tx.QueryRowContext(ctx, "SELECT state FROM worktree_relocations WHERE worktree_id=?", id).Scan(&state)
					if e != nil && !errors.Is(e, sql.ErrNoRows) {
						return rollback(e)
					}
					if e == nil && state != "active" {
						return rollback(errors.New("capture worktree relocation incomplete; resume explicit relocation"))
					}
				}
			}
		}
		if e = tx.Commit(); e != nil {
			return fail(e)
		}
	}
	if readOnly && project != "" {
		coord, err = readCoordinatePath(ctx, db, directory, id)
		if err != nil {
			return fail(err)
		}
	}
	if !readOnly {
		if err = ctx.Err(); err != nil {
			return fail(err)
		}
		if err = os.MkdirAll(coord, 0700); err != nil {
			return fail(err)
		}
	}
	store := &Store{db: db, project: project, directory: coord, registry: directory, worktreeID: id, relocationWorker: relocationWorker}
	if !readOnly && !allowPending {
		if err = store.initializeDiagnosticIndexContext(ctx); err != nil {
			return fail(err)
		}
	}
	return store, nil
}
func (s *Store) CoordinationDirectory() string { return s.directory }
func (s *Store) RegistryDirectory() string     { return s.registry }
func (s *Store) Close() error {
	err := s.db.Close()
	if s.relocationWorker != nil {
		err = errors.Join(err, s.relocationWorker.Close())
		s.relocationWorker = nil
	}
	return err
}
func IdentityKey(id domain.CaptureIdentity) string {
	b, err := json.Marshal(id)
	if err != nil {
		panic(err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// encodeWithSnapshot retains an unsupported snapshot in its existing registry
// row while independently valid registration metadata changes. No removed fact
// is decoded, projected, republished or moved into a replacement sidecar.
func encodeWithSnapshot(r domain.CaptureRegistration, opaque json.RawMessage) ([]byte, error) {
	if !r.SnapshotNeedsRefresh {
		return encode(r)
	}
	if len(opaque) == 0 {
		return nil, errors.New("opaque source snapshot missing during metadata update")
	}
	metadata, err := json.Marshal(storedRegistration{Registration: r, Path: r.TranscriptPath})
	if err != nil {
		return nil, err
	}
	result := append([]byte{}, metadata[:len(metadata)-1]...)
	result = append(result, []byte(`,"snapshot":`)...)
	result = append(result, opaque...)
	result = append(result, '}')
	return result, nil
}
func encode(r domain.CaptureRegistration) ([]byte, error) {
	if r.SnapshotNeedsRefresh {
		return nil, errors.New("unsupported snapshot requires native source re-extraction")
	}
	return json.Marshal(storedRegistration{r, r.TranscriptPath, r.Snapshot})
}

type registrationInput struct {
	Registration domain.CaptureRegistration `json:"registration"`
	Path         string                     `json:"path"`
	Snapshot     json.RawMessage            `json:"snapshot"`
}

const snapshotReextractDetail = "Cached snapshot requires native source re-extraction before publishing current lean facts"

func decode(b []byte) (domain.CaptureRegistration, error) {
	var input registrationInput
	if err := json.Unmarshal(b, &input); err != nil {
		return domain.CaptureRegistration{}, err
	}
	r := input.Registration
	r.TranscriptPath = input.Path
	r.Snapshot = nil
	r.SnapshotNeedsRefresh = false
	if len(input.Snapshot) == 0 || bytes.Equal(bytes.TrimSpace(input.Snapshot), []byte("null")) {
		return r, nil
	}
	snapshot, err := decodeCurrentSnapshot(input.Snapshot)
	if err == nil {
		r.Snapshot = snapshot
		return r, nil
	}
	r.SnapshotNeedsRefresh = true
	if r.State == domain.CaptureCurrent {
		r.State = domain.CapturePending
	}
	if !strings.Contains(r.Detail, snapshotReextractDetail) {
		if r.Detail != "" {
			r.Detail += "; "
		}
		r.Detail += snapshotReextractDetail
	}
	return r, nil
}

// Unsupported cache is opaque. Strict current decoding does not convert a rich
// snapshot into empty lean collections or grant source replacement authority.
func decodeCurrentSnapshot(data []byte) ([]domain.SessionFact, error) {
	// Draft duration input is a supported numeric compatibility transition,
	// unlike retired trace/occurrence payloads which remain opaque.
	var rawOwners []json.RawMessage
	if err := json.Unmarshal(data, &rawOwners); err != nil {
		return nil, err
	}
	for i, owner := range rawOwners {
		normalized, err := migration.NormalizeSessionDurationJSON(owner)
		if err != nil {
			return nil, fmt.Errorf("snapshot owner %d: %w", i, err)
		}
		rawOwners[i] = normalized
	}
	normalized, err := json.Marshal(rawOwners)
	if err != nil {
		return nil, err
	}
	data = normalized
	var snapshot []domain.SessionFact
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, err
	}
	var fields []struct {
		Skills      []map[string]json.RawMessage `json:"skills"`
		Delegations []map[string]json.RawMessage `json:"delegations"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for i, owner := range fields {
		for _, fact := range owner.Skills {
			for _, key := range []string{"name", "caller_role", "invoked_at"} {
				if _, ok := fact[key]; !ok {
					return nil, fmt.Errorf("snapshot owner %d skill requires %s", i, key)
				}
			}
		}
		for _, fact := range owner.Delegations {
			for _, key := range []string{"agent_type", "invoked_at"} {
				if _, ok := fact[key]; !ok {
					return nil, fmt.Errorf("snapshot owner %d delegation requires %s", i, key)
				}
			}
		}
	}
	return snapshot, nil
}
func (s *Store) List(ctx context.Context) ([]domain.CaptureRegistration, error) {
	return s.list(ctx, 0)
}

// ListBounded limits inspection output while preserving explicit continuation.
func (s *Store) ListBounded(ctx context.Context, limit int) ([]domain.CaptureRegistration, bool, error) {
	if limit < 1 {
		return nil, false, errors.New("capture inspection requires a positive session limit")
	}
	rows, err := s.list(ctx, limit+1)
	if err != nil {
		return nil, false, err
	}
	more := len(rows) > limit
	if more {
		rows = rows[:limit]
	}
	return rows, more, nil
}
func (s *Store) list(ctx context.Context, limit int) ([]domain.CaptureRegistration, error) {
	var rows *sql.Rows
	var err error
	if limit > 0 {
		rows, err = s.db.QueryContext(ctx, `SELECT record FROM registrations WHERE worktree_id=? ORDER BY identity LIMIT ?`, s.worktreeID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT record FROM registrations WHERE worktree_id=? ORDER BY identity`, s.worktreeID)
	}
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
	rows, err := conn.QueryContext(ctx, `SELECT identity,record FROM registrations WHERE worktree_id=? AND source=? AND session_id=?`, s.worktreeID, id.Source, id.SessionID)
	if err != nil {
		return r, err
	}
	matched := []string{}
	var opaqueSnapshot json.RawMessage
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
			if candidate.SnapshotNeedsRefresh {
				var input registrationInput
				if err := json.Unmarshal(raw, &input); err != nil {
					return r, errors.Join(err, rows.Close())
				}
				opaqueSnapshot = append(json.RawMessage{}, input.Snapshot...)
			}
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
	b, err = encodeWithSnapshot(r, opaqueSnapshot)
	if err != nil {
		return r, err
	}
	for _, oldKey := range matched {
		if oldKey != IdentityKey(id) {
			if _, err = conn.ExecContext(ctx, `DELETE FROM registrations WHERE worktree_id=? AND identity=?`, s.worktreeID, oldKey); err != nil {
				return r, err
			}
		}
	}
	_, err = conn.ExecContext(ctx, `INSERT INTO registrations(worktree_id,identity,source,store,session_id,ticket,record) VALUES(?,?,?,?,?,?,?) ON CONFLICT(worktree_id,identity) DO UPDATE SET ticket=excluded.ticket,record=excluded.record`, s.worktreeID, IdentityKey(id), id.Source, id.Store, id.SessionID, r.TicketID, b)
	if err != nil {
		return r, err
	}
	if _, err = conn.ExecContext(ctx, `COMMIT`); err != nil {
		return r, err
	}
	committed = true
	return r, nil
}

// AcceptedRegistrationError reports a post-handoff failure without pretending
// the durable identity was rolled back. Callers may still schedule recovery.
type AcceptedRegistrationError struct{ Err error }

func (err *AcceptedRegistrationError) Error() string {
	return "registration accepted but installation verification unavailable: " + err.Err.Error()
}
func (err *AcceptedRegistrationError) Unwrap() error { return err.Err }

func (s *Store) Register(ctx context.Context, id domain.CaptureIdentity, path, event string) (domain.CaptureRegistration, error) {
	observedCallback := false
	row, err := s.Change(ctx, id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if path != "" {
			if r.TranscriptPath != "" && r.TranscriptPath != path {
				return r, errors.New("conflicting transcript path for registered identity")
			}
			r.TranscriptPath = path
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if r.FirstObservedAt == "" {
			r.FirstObservedAt = now
		}
		switch event {
		case "SessionStart", "SubagentStart", "SubagentStop", "Stop", "SessionEnd", "UserPromptSubmit", "session.created", "session.updated", "session.status", "session.idle", "session.error":
			r.CallbackObservedAt = now
			observedCallback = true
		}
		r.Event = event
		if r.TicketID != "" {
			r.State = domain.CapturePending
		}
		return r, nil
	})
	if err != nil || !observedCallback {
		return row, err
	}
	// Installation inspection follows the durable minimal handoff. Failure to
	// inspect cannot erase an observed session or fabricate verified installation.
	installation, _, inspectErr := InstallationIdentity(s.directory, string(id.Source))
	if errors.Is(inspectErr, os.ErrNotExist) {
		return row, nil
	}
	if inspectErr != nil {
		return row, &AcceptedRegistrationError{Err: inspectErr}
	}
	updated, err := s.Change(ctx, id, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		current.InstallationID = installation
		return current, nil
	})
	if err != nil {
		return row, &AcceptedRegistrationError{Err: err}
	}
	return updated, nil
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
		if s.project != "" && r.Identity.Source == domain.SourceCodex && r.TranscriptPath == "" {
			path, err := codex.DiscoverExactRolloutPath(r.Identity.Store, r.Identity.SessionID)
			if err != nil {
				return r, fmt.Errorf("binding source identity unavailable; minimal registration retained: %w", err)
			}
			r.TranscriptPath = path
		}
		if err := s.validateOwner(ctx, r); err != nil {
			return r, fmt.Errorf("binding source/project verification failed; minimal registration retained: %w", err)
		}
		r.ProjectDirectory = s.project
		r.ProjectProof = domain.CaptureProjectVerified
		r.TicketID = bound
		r.Binding = "explicit_agent_response"
		r.BindingInstallationID = r.InstallationID
		r.BindingAcceptedAt = time.Now().UTC().Format(time.RFC3339Nano)
		r.State = domain.CapturePending
		return r, nil
	})
}

func constructorBusyMillis(ctx context.Context, maximum int64) int64 {
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline).Milliseconds()
		if remaining < 0 {
			return 0
		}
		if remaining < maximum {
			return remaining
		}
	}
	return maximum
}
func setConstructorBusyTimeout(ctx context.Context, db *sql.DB, maximum int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA busy_timeout=%d", constructorBusyMillis(ctx, maximum)))
	return err
}
