package capture

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agenticlens/internal/domain"
)

// A restore may spend several bounded passes proving the whole group. Record
// accepted reads for its exact revision so a later pass cannot mistake cached
// pre-discard facts, or a stale pass after a concurrent transition, for proof.
func (s *Store) captureRestoresReady(ctx context.Context, result ReconcileResult, cutoff time.Time, rows []domain.CaptureRegistration, attempted, refreshed, openCodeAttempts map[string]bool) (ready []CaptureControl, returnErr error) {
	controls, err := s.captureControls(ctx, nil, "", CaptureControlRestoring, false)
	if err != nil || len(controls) == 0 {
		return nil, err
	}
	roots := make(map[string]bool, len(controls))
	for _, control := range controls {
		roots[IdentityKey(control.Root)] = true
	}
	owners := make(map[string][]domain.CaptureRegistration, len(controls))
	groups := captureGroups(s.project, rows, roots)
	known := acceptedCaptureRegistrations(s.project, result.ProjectRegistrations)
	activeRootTickets := make(map[string]string, len(controls))
	for _, row := range rows {
		identity := IdentityKey(row.Identity)
		if roots[identity] {
			activeRootTickets[identity] = row.TicketID
		}
		if key := groups[identity]; key != "" {
			owners[key] = append(owners[key], row)
		}
	}
	returnErr = s.WithCaptureLifecycle(ctx, result.LifecycleRevision, func() error {
		return s.withRegistrationWrite(ctx, func(conn *sql.Conn) error {
			for _, control := range controls {
				if cutoff.Before(control.requestedAt) {
					continue
				}
				rootKey := IdentityKey(control.Root)
				ticket, present := activeRootTickets[rootKey]
				group := owners[rootKey]
				if !present || len(group) == 0 || ticket != result.TicketID {
					continue
				}
				if err := s.recordRestoreAttempts(ctx, conn, control, group, attempted, openCodeAttempts); err != nil {
					return err
				}
				covered, err := s.recordRestoreReads(ctx, conn, control, group, attempted, refreshed)
				if err != nil {
					return err
				}
				pending, err := s.restoreGroupDiscoveryPending(ctx, conn, control, ticket, group, known)
				if err != nil {
					return err
				}
				if covered && !pending {
					ready = append(ready, control)
				}
			}
			return nil
		})
	})
	return ready, returnErr
}

func (s *Store) recordRestoreAttempts(ctx context.Context, conn *sql.Conn, control CaptureControl, owners []domain.CaptureRegistration, attempted, openCodeAttempts map[string]bool) error {
	for _, row := range owners {
		key := IdentityKey(row.Identity)
		if attempted[key] || openCodeAttempts[key] {
			// Count a first attempt once so later owners remain reachable even
			// when an unavailable sibling has consumed its turn in the rotation.
			if _, err := conn.ExecContext(ctx, `INSERT INTO capture_restore_attempts VALUES(?,?,?,?,?,?) ON CONFLICT(worktree_id,capture_id,revision,identity) DO UPDATE SET source_read=max(source_read,excluded.source_read),discovery=max(discovery,excluded.discovery)`, s.worktreeID, control.CaptureID, control.Revision, key, attempted[key], openCodeAttempts[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

// A ticket can contain independent captures. Only this group's actual native
// ancestry and discovery work can leave its restore incomplete.
func (s *Store) restoreGroupDiscoveryPending(ctx context.Context, conn *sql.Conn, control CaptureControl, ticket string, owners []domain.CaptureRegistration, known map[string]domain.CaptureRegistration) (bool, error) {
	groupPending := false
	for _, row := range owners {
		if row.TicketID != ticket {
			groupPending = true
			continue
		}
		if row.ParentID != "" {
			parent := row.Identity
			parent.SessionID = row.ParentID
			upstream, present := known[IdentityKey(parent)]
			// The admitted independently bound root owns its own source
			// history even if its upstream root has no local registration.
			independentRoot := row.Identity == control.Root && independentCapture(row, upstream)
			if !present && !independentRoot {
				groupPending = true
			}
		}
		var pending bool
		var err error
		if row.Identity.Source == domain.SourceOpenCode {
			// Restore clears each owner's cursor. An owner needs its own
			// completed scan for this revision, which may take several passes.
			pending, err = s.recordRestoreOpenCodeScan(ctx, conn, control, row.Identity)
		} else {
			err = conn.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pending_children WHERE worktree_id=? AND store=? AND parent=?) OR EXISTS(SELECT 1 FROM pending_starts WHERE worktree_id=? AND store=? AND parent=?)`, s.worktreeID, row.Identity.Store, row.Identity.SessionID, s.worktreeID, row.Identity.Store, row.Identity.SessionID).Scan(&pending)
		}
		if err != nil {
			return false, err
		}
		groupPending = groupPending || pending
	}
	return groupPending, nil
}

func (s *Store) recordRestoreOpenCodeScan(ctx context.Context, conn *sql.Conn, control CaptureControl, owner domain.CaptureIdentity) (bool, error) {
	key := IdentityKey(owner)
	var cursor string
	var more bool
	err := conn.QueryRowContext(ctx, `SELECT key,more FROM child_cursor WHERE worktree_id=? AND kind='opencode' AND parent=?`, s.worktreeID, key).Scan(&cursor, &more)
	if errors.Is(err, sql.ErrNoRows) {
		more = true
	} else if err != nil {
		return false, err
	}
	// Ordinary polling can restart a completed OpenCode scan. Retain the
	// completion for this restore and require new receipts for any newly
	// discovered owners. A replay cannot count as progress indefinitely.
	var complete bool
	err = conn.QueryRowContext(ctx, `INSERT INTO capture_restore_progress VALUES(?,?,?,?,1,?,?,0) ON CONFLICT(worktree_id,capture_id,revision,identity) DO UPDATE SET discovery=1,position=max(position,excluded.position),complete=max(complete,excluded.complete) RETURNING complete`, s.worktreeID, control.CaptureID, control.Revision, key, cursor, !more).Scan(&complete)
	return !complete, err
}

func (s *Store) recordRestoreReads(ctx context.Context, conn *sql.Conn, control CaptureControl, owners []domain.CaptureRegistration, attempted, refreshed map[string]bool) (bool, error) {
	all := true
	for _, row := range owners {
		key := IdentityKey(row.Identity)
		verified := row.ProjectProof == domain.CaptureProjectVerified
		if verified && refreshed[key] && len(row.Snapshot) != 0 && (row.State == domain.CaptureCurrent || row.State == domain.CaptureUnassigned) {
			if _, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO capture_restore_reads VALUES(?,?,?,?)`, s.worktreeID, control.CaptureID, control.Revision, key); err != nil {
				return false, err
			}
			// Readiness can expire on a later failed attempt. Preserve the
			// progress milestone so rereading this same owner cannot extend
			// an idle worker over and over while another source is missing.
			if _, err := conn.ExecContext(ctx, `INSERT INTO capture_restore_progress VALUES(?,?,?,?,0,'',0,1) ON CONFLICT(worktree_id,capture_id,revision,identity) DO UPDATE SET source_read=1`, s.worktreeID, control.CaptureID, control.Revision, key); err != nil {
				return false, err
			}
			continue
		}
		if attempted[key] || !verified || row.State == domain.CaptureFailed || len(row.Snapshot) == 0 {
			if _, err := conn.ExecContext(ctx, `DELETE FROM capture_restore_reads WHERE worktree_id=? AND capture_id=? AND revision=? AND identity=?`, s.worktreeID, control.CaptureID, control.Revision, key); err != nil {
				return false, err
			}
			all = false
			continue
		}
		var read bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM capture_restore_reads WHERE worktree_id=? AND capture_id=? AND revision=? AND identity=?)`, s.worktreeID, control.CaptureID, control.Revision, key).Scan(&read); err != nil {
			return false, err
		}
		all = all && read
	}
	return all, nil
}
