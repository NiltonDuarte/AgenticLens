package capture

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

type CaptureControlState string

const (
	CaptureControlUnknown   CaptureControlState = "unknown"
	CaptureControlActive    CaptureControlState = "active"
	CaptureControlDiscarded CaptureControlState = "discarded"
	CaptureControlRestoring CaptureControlState = "restore_pending"
)

// CaptureControl exposes local disposition without disclosing source paths or credentials.
type CaptureControl struct {
	CaptureID       string                 `json:"capture_id"`
	Root            domain.CaptureIdentity `json:"-"`
	Source          domain.SourceHarness   `json:"source,omitempty"`
	RootSessionID   string                 `json:"root_session_id,omitempty"`
	State           CaptureControlState    `json:"state"`
	Revision        int64                  `json:"revision"`
	CleanupPending  bool                   `json:"cleanup_pending"`
	CoveragePending bool                   `json:"coverage_pending"`
	TicketID        string                 `json:"ticket_id,omitempty"`
	requestedAt     time.Time
}

var ErrCaptureLifecycleChanged = errors.New("capture lifecycle changed; retry with current registry state")

// StableCaptureID is the existing portable pending identity. Registry authority
// still requires resolving a unique verified owner of that ID in the worktree.
func StableCaptureID(source domain.SourceHarness, root string) string {
	digest := sha256.Sum256([]byte(string(source) + "\x00" + root))
	return hex.EncodeToString(digest[:])
}

func validateCaptureID(id string) error {
	value, err := hex.DecodeString(id)
	if err != nil || len(value) != sha256.Size || strings.ToLower(id) != id {
		return errors.New("capture ID must be the full lowercase hash from capture status or its pending snapshot")
	}
	return nil
}

func (s *Store) captureControls(ctx context.Context, conn *sql.Conn, id string, state CaptureControlState, cleanupOnly bool) ([]CaptureControl, error) {
	return s.readCaptureControls(ctx, conn, id, state, cleanupOnly, -1)
}

func (s *Store) readCaptureControls(ctx context.Context, conn *sql.Conn, id string, state CaptureControlState, cleanupOnly bool, limit int) ([]CaptureControl, error) {
	const query = `SELECT capture_id,source,store,session_id,state,revision,cleanup_pending,requested_at FROM capture_controls
WHERE worktree_id=? AND (?='' OR capture_id=?) AND (?='' OR state=?) AND (NOT ? OR cleanup_pending=1) ORDER BY capture_id LIMIT ?`
	var rows *sql.Rows
	var err error
	if conn != nil {
		rows, err = conn.QueryContext(ctx, query, s.worktreeID, id, id, state, state, cleanupOnly, limit)
	} else {
		var exists bool
		err = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='capture_controls')`).Scan(&exists)
		if err != nil || !exists {
			return nil, err
		}
		rows, err = s.db.QueryContext(ctx, query, s.worktreeID, id, id, state, state, cleanupOnly, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var controls []CaptureControl
	for rows.Next() {
		var control CaptureControl
		var at string
		if err := rows.Scan(&control.CaptureID, &control.Root.Source, &control.Root.Store, &control.Root.SessionID, &control.State, &control.Revision, &control.CleanupPending, &at); err != nil {
			return nil, err
		}
		control.Source = control.Root.Source
		control.RootSessionID = control.Root.SessionID
		control.requestedAt, err = time.Parse(time.RFC3339Nano, at)
		if err != nil {
			return nil, fmt.Errorf("reading capture lifecycle time: %w", err)
		}
		switch control.State {
		case CaptureControlActive, CaptureControlDiscarded:
		case CaptureControlRestoring:
			control.CoveragePending = true
		default:
			return nil, errors.New("invalid persisted capture lifecycle state")
		}
		if control.Revision < 1 || StableCaptureID(control.Root.Source, control.Root.SessionID) != control.CaptureID || (control.CleanupPending && control.State != CaptureControlDiscarded) {
			return nil, errors.New("invalid persisted capture lifecycle selection")
		}
		controls = append(controls, control)
	}
	return controls, rows.Err()
}

func (s *Store) CaptureLifecycleRevision(ctx context.Context) (int64, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='capture_lifecycle_revision')`).Scan(&exists)
	if err != nil || !exists {
		return 0, err
	}
	var revision int64
	err = s.db.QueryRowContext(ctx, `SELECT revision FROM capture_lifecycle_revision WHERE worktree_id=?`, s.worktreeID).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return revision, err
}

const captureControlRevisionSQL = `SELECT coalesce(max(revision),0) FROM capture_controls WHERE worktree_id=?`

func (s *Store) captureControlRevision(ctx context.Context) (int64, error) {
	// Membership changes advance the publishing fence, but do not reset a
	// source cursor. Control rows remain durable through completed operations.
	var revision int64
	err := s.db.QueryRowContext(ctx, captureControlRevisionSQL, s.worktreeID).Scan(&revision)
	return revision, err
}

func (s *Store) advanceLifecycleRevision(ctx context.Context, conn *sql.Conn) (int64, error) {
	var revision int64
	err := conn.QueryRowContext(ctx, `INSERT INTO capture_lifecycle_revision VALUES(?,1) ON CONFLICT(worktree_id) DO UPDATE SET revision=revision+1 RETURNING revision`, s.worktreeID).Scan(&revision)
	return revision, err
}

// The short worktree fence serializes dispositions and final publication. A
// publisher takes destination ownership first; registry mutations do no I/O
// against those destinations while they hold this fence.
func (s *Store) withLifecycleLock(ctx context.Context, action func() error) (returnErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.project == "" {
		return action()
	}
	file, err := os.OpenFile(filepath.Join(s.directory, "capture-lifecycle.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("capture lifecycle lock is not a regular file")
	}
	for {
		locked, err := tryPublicationLock(file)
		if err != nil {
			return err
		}
		if locked {
			defer func() { returnErr = errors.Join(returnErr, unlockPublication(file)) }()
			if err := ctx.Err(); err != nil {
				return err
			}
			return action()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (s *Store) withLifecycleRegistrationWrite(ctx context.Context, write func(*sql.Conn) error) error {
	return s.withLifecycleLock(ctx, func() error { return s.withRegistrationWrite(ctx, write) })
}

func (s *Store) withCaptureControlRegistrationWrite(ctx context.Context, observed int64, write func(*sql.Conn) error) error {
	return s.withLifecycleRegistrationWrite(ctx, func(conn *sql.Conn) error {
		var current int64
		if err := conn.QueryRowContext(ctx, captureControlRevisionSQL, s.worktreeID).Scan(&current); err != nil {
			return err
		}
		if current != observed {
			return ErrCaptureLifecycleChanged
		}
		return write(conn)
	})
}

// WithCaptureLifecycle fences a prepared publication against a later control
// operation or membership change. The callback must not take the fence again.
func (s *Store) WithCaptureLifecycle(ctx context.Context, revision int64, action func() error) error {
	return s.withLifecycleLock(ctx, func() error {
		current, err := s.CaptureLifecycleRevision(ctx)
		if err != nil {
			return err
		}
		if current != revision {
			return ErrCaptureLifecycleChanged
		}
		return action()
	})
}

// ProjectDirectory is the accepted association. A failed later proof changes
// ProjectProof but retains the last accepted project and native parent edge.
func acceptedCaptureRegistrations(project string, rows []domain.CaptureRegistration) map[string]domain.CaptureRegistration {
	known := make(map[string]domain.CaptureRegistration, len(rows))
	for _, row := range rows {
		if project != "" && row.ProjectDirectory == project {
			known[IdentityKey(row.Identity)] = row
		}
	}
	return known
}

// Explicit differently bound descendants own their own independent capture.
func independentCapture(row, parent domain.CaptureRegistration) bool {
	return row.TicketID != "" && row.TicketID != parent.TicketID && (row.Binding == "explicit_agent_response" || row.Binding == "explicit_manual_selection")
}

func captureGroupParent(row domain.CaptureRegistration, known map[string]domain.CaptureRegistration) (domain.CaptureRegistration, bool) {
	id := row.Identity
	id.SessionID = row.ParentID
	parent, verified := known[IdentityKey(id)]
	if !verified || independentCapture(row, parent) {
		return domain.CaptureRegistration{}, false
	}
	return parent, true
}

func captureGroups(project string, rows []domain.CaptureRegistration, roots map[string]bool) map[string]string {
	known := acceptedCaptureRegistrations(project, rows)
	groups := make(map[string]string, len(known))
	seen := make(map[string]bool, len(known))
	for key, row := range known {
		if seen[key] {
			continue
		}
		var path []string
		group := ""
		for {
			key = IdentityKey(row.Identity)
			if seen[key] {
				group = groups[key]
				break
			}
			seen[key] = true
			path = append(path, key)
			if roots[key] {
				group = key
				break
			}
			parent, ok := captureGroupParent(row, known)
			if !ok {
				break
			}
			row = parent
		}
		for _, member := range path {
			groups[member] = group
		}
	}
	return groups
}

func activeCaptureRows(project string, rows []domain.CaptureRegistration, controls []CaptureControl) ([]domain.CaptureRegistration, map[string]bool) {
	discardedRoots := map[string]bool{}
	for _, control := range controls {
		if control.State == CaptureControlDiscarded {
			discardedRoots[IdentityKey(control.Root)] = true
		}
	}
	if len(discardedRoots) == 0 {
		return rows, nil
	}
	groups := captureGroups(project, rows, discardedRoots)
	active := make([]domain.CaptureRegistration, 0, len(rows))
	suppressed := make(map[string]bool)
	for _, row := range rows {
		key := IdentityKey(row.Identity)
		if discardedRoots[key] || groups[key] != "" {
			suppressed[row.Identity.SessionID] = true
			continue
		}
		active = append(active, row)
	}
	return active, suppressed
}

func (s *Store) ListActive(ctx context.Context) ([]domain.CaptureRegistration, error) {
	rows, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	controls, err := s.captureControls(ctx, nil, "", CaptureControlDiscarded, false)
	if err != nil {
		return nil, err
	}
	active, _ := activeCaptureRows(s.project, rows, controls)
	return active, nil
}

func (s *Store) resolveCaptureControl(ctx context.Context, conn *sql.Conn, id string) (CaptureControl, error) {
	control := CaptureControl{CaptureID: id, State: CaptureControlUnknown}
	if s.project == "" {
		return control, errors.New("capture control requires a selected worktree")
	}
	controls, err := s.captureControls(ctx, conn, id, "", false)
	if err != nil {
		return control, err
	}
	var rows []domain.CaptureRegistration
	if conn != nil {
		rows, err = readRegistrationRows(conn.QueryContext(ctx, `SELECT record FROM registrations WHERE worktree_id=?`, s.worktreeID))
	} else {
		rows, err = s.List(ctx)
	}
	if err != nil {
		return control, err
	}
	known := acceptedCaptureRegistrations(s.project, rows)
	if len(controls) > 0 {
		control = controls[0]
		control.TicketID = known[IdentityKey(control.Root)].TicketID
	}
	matches := 0
	for _, row := range known {
		if StableCaptureID(row.Identity.Source, row.Identity.SessionID) != id {
			continue
		}
		matches++
		if len(controls) != 0 {
			if IdentityKey(row.Identity) != IdentityKey(control.Root) {
				return control, errors.New("capture ID no longer names the recorded verified source owner")
			}
			continue
		}
		if row.ProjectProof != domain.CaptureProjectVerified {
			continue
		}
		parent, hasParent := captureGroupParent(row, known)
		if hasParent || (row.ParentID != "" && !independentCapture(row, parent)) {
			continue
		}
		control.Root = row.Identity
		control.Source = row.Identity.Source
		control.RootSessionID = row.Identity.SessionID
		control.State = CaptureControlActive
		control.TicketID = row.TicketID
	}
	if matches > 1 {
		return control, errors.New("ambiguous capture ID across verified source stores")
	}
	return control, nil
}

func (s *Store) CaptureControl(ctx context.Context, id string) (CaptureControl, error) {
	if err := validateCaptureID(id); err != nil {
		return CaptureControl{}, err
	}
	return s.resolveCaptureControl(ctx, nil, id)
}

// ConfirmStagedDiscard observes only a durable discard in the selected
// worktree. The staged identity omits the source store, so a competing store
// associated with this worktree leaves the result unconfirmed. Resolve the
// disposition and its bounded registration evidence in one read transaction.
func (s *Store) ConfirmStagedDiscard(ctx context.Context, source domain.SourceHarness, rootSessionID string) (confirmed bool, returnErr error) {
	if s.project == "" {
		return false, errors.New("capture discard inspection requires a selected worktree")
	}
	const controlQuery = `SELECT requested_at,store FROM capture_controls
WHERE worktree_id=? AND capture_id=? AND source=? AND session_id=?
AND state=? AND revision>0 AND trim(store)<>''`
	const registrationQuery = `SELECT store,CASE WHEN length(CAST(record AS BLOB)) <= ? THEN record END
FROM registrations WHERE worktree_id=? AND source=? AND session_id=? LIMIT ?`
	const registrationLimit = 16
	const registrationBytes = 1 << 20
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return false, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			confirmed = false
			returnErr = errors.Join(returnErr, err)
		}
	}()
	id := StableCaptureID(source, rootSessionID)
	var at, store string
	err = tx.QueryRowContext(ctx, controlQuery, s.worktreeID, id, source, rootSessionID, CaptureControlDiscarded).Scan(&at, &store)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := time.Parse(time.RFC3339Nano, at); err != nil {
		return false, fmt.Errorf("reading staged capture discard time: %w", err)
	}
	// Oversized records return NULL without loading their payload.
	rows, err := tx.QueryContext(ctx, registrationQuery, registrationBytes, s.worktreeID, source, rootSessionID, registrationLimit+1)
	if err != nil {
		return false, fmt.Errorf("read staged capture registrations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			confirmed = false
			returnErr = errors.Join(returnErr, err)
		}
	}()
	count := 0
	for rows.Next() {
		count++
		if count > registrationLimit {
			return false, errors.New("staged capture registration inspection exceeded its row limit")
		}
		var rowStore string
		var record []byte
		if err := rows.Scan(&rowStore, &record); err != nil {
			return false, fmt.Errorf("scan staged capture registration: %w", err)
		}
		row, err := decode(record)
		if err != nil {
			return false, fmt.Errorf("decode staged capture registration: %w", err)
		}
		expected := domain.CaptureIdentity{Source: source, Store: rowStore, SessionID: rootSessionID}
		if row.Identity != expected {
			return false, errors.New("staged capture registration identity differs from its indexed identity")
		}
		if rowStore != store && row.ProjectDirectory == s.project {
			return false, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("read staged capture registrations: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) CaptureControlForRoot(ctx context.Context, root domain.CaptureIdentity) (CaptureControl, error) {
	control, err := s.CaptureControl(ctx, StableCaptureID(root.Source, root.SessionID))
	if err != nil {
		return control, err
	}
	if control.State != CaptureControlUnknown && IdentityKey(control.Root) != IdentityKey(root) {
		return CaptureControl{}, errors.New("capture ID is owned by a different native source")
	}
	return control, nil
}

// SetCaptureDiscarded persists the state and publication wake in one transaction.
// A native authenticated request may also compare the exact control revision.
func (s *Store) SetCaptureDiscarded(ctx context.Context, id string, discarded bool, expected *int64) (CaptureControl, error) {
	if err := validateCaptureID(id); err != nil {
		return CaptureControl{}, err
	}
	var control CaptureControl
	err := s.withLifecycleRegistrationWrite(ctx, func(conn *sql.Conn) error {
		var err error
		control, err = s.resolveCaptureControl(ctx, conn, id)
		if err != nil {
			return err
		}
		if control.State == CaptureControlUnknown {
			return errors.New("capture ID does not identify a verified capture root in the selected worktree")
		}
		if expected != nil && control.Revision != *expected {
			return ErrCaptureLifecycleChanged
		}
		if (control.State == CaptureControlDiscarded) == discarded {
			if control.CleanupPending || control.CoveragePending {
				_, err = conn.ExecContext(ctx, requestPublicationSQL, s.worktreeID, time.Now().UnixMilli())
				return err
			}
			return nil
		}
		control.Revision, err = s.advanceLifecycleRevision(ctx, conn)
		if err != nil {
			return err
		}
		control.State = CaptureControlDiscarded
		control.CleanupPending = discarded
		control.CoveragePending = !discarded
		control.requestedAt = time.Now().UTC()
		if !discarded {
			control.State = CaptureControlRestoring
			if err := s.resetRestoreDiscovery(ctx, conn, control); err != nil {
				return err
			}
		}
		_, err = conn.ExecContext(ctx, `INSERT INTO capture_controls VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(worktree_id,capture_id) DO UPDATE SET state=excluded.state,revision=excluded.revision,cleanup_pending=excluded.cleanup_pending,requested_at=excluded.requested_at`, s.worktreeID, id, control.Root.Source, control.Root.Store, control.Root.SessionID, control.State, control.Revision, control.CleanupPending, control.requestedAt.Format(time.RFC3339Nano))
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_reads WHERE worktree_id=? AND capture_id=?`, s.worktreeID, id); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_progress WHERE worktree_id=? AND capture_id=?`, s.worktreeID, id); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_attempts WHERE worktree_id=? AND capture_id=?`, s.worktreeID, id); err != nil {
			return err
		}
		_, err = conn.ExecContext(ctx, requestPublicationSQL, s.worktreeID, control.requestedAt.UnixMilli())
		return err
	})
	return control, err
}

func (s *Store) resetRestoreDiscovery(ctx context.Context, conn *sql.Conn, control CaptureControl) error {
	rows, err := readRegistrationRows(conn.QueryContext(ctx, `SELECT record FROM registrations WHERE worktree_id=?`, s.worktreeID))
	if err != nil {
		return err
	}
	rootKey := IdentityKey(control.Root)
	groups := captureGroups(s.project, rows, map[string]bool{rootKey: true})
	for key, root := range groups {
		if root != rootKey {
			continue
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM discovery_cursor WHERE worktree_id=? AND parent=?`, s.worktreeID, key); err != nil {
			return err
		}
		if _, err := conn.ExecContext(ctx, `DELETE FROM child_cursor WHERE worktree_id=? AND kind='opencode' AND parent=?`, s.worktreeID, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) PendingCaptureDiscards(ctx context.Context) ([]CaptureControl, error) {
	return s.captureControls(ctx, nil, "", CaptureControlDiscarded, true)
}

func (s *Store) ListCaptureControls(ctx context.Context, limit int) ([]CaptureControl, bool, error) {
	if limit <= 0 || limit == int(^uint(0)>>1) {
		return nil, false, errors.New("capture status limit must be positive and bounded")
	}
	controls, err := s.readCaptureControls(ctx, nil, "", "", false, limit+1)
	if err != nil || len(controls) <= limit {
		return controls, false, err
	}
	return controls[:limit], true, nil
}

// WithDiscardedCapture completes cleanup only for the observed still-current
// discard. The caller holds the pending directory lease through the callback.
func (s *Store) WithDiscardedCapture(ctx context.Context, id string, revision int64, action func() error) error {
	return s.withLifecycleLock(ctx, func() error {
		rows, err := s.captureControls(ctx, nil, id, "", false)
		if err != nil {
			return err
		}
		if len(rows) != 1 || rows[0].Revision != revision || rows[0].State != CaptureControlDiscarded {
			return ErrCaptureLifecycleChanged
		}
		if !rows[0].CleanupPending {
			return nil
		}
		if err := action(); err != nil {
			return err
		}
		_, err = s.db.ExecContext(ctx, `UPDATE capture_controls SET cleanup_pending=0 WHERE worktree_id=? AND capture_id=? AND revision=?`, s.worktreeID, id, revision)
		return err
	})
}

func (s *Store) CompleteCaptureRestores(ctx context.Context, controls []CaptureControl) error {
	if len(controls) == 0 {
		return nil
	}
	return s.withRegistrationWrite(ctx, func(conn *sql.Conn) error {
		for _, control := range controls {
			if _, err := conn.ExecContext(ctx, `UPDATE capture_controls SET state=? WHERE worktree_id=? AND capture_id=? AND revision=? AND state=?`, CaptureControlActive, s.worktreeID, control.CaptureID, control.Revision, CaptureControlRestoring); err != nil {
				return err
			}
			if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_reads WHERE worktree_id=? AND capture_id=? AND revision=?`, s.worktreeID, control.CaptureID, control.Revision); err != nil {
				return err
			}
			if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_progress WHERE worktree_id=? AND capture_id=? AND revision=?`, s.worktreeID, control.CaptureID, control.Revision); err != nil {
				return err
			}
			if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_attempts WHERE worktree_id=? AND capture_id=? AND revision=?`, s.worktreeID, control.CaptureID, control.Revision); err != nil {
				return err
			}
		}
		return nil
	})
}
