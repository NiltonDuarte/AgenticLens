package capture

import (
	"context"
	"database/sql"
	"errors"

	"agenticlens/internal/domain"
)

// CaptureRestoreProgress summarizes durable source and discovery evidence for
// the exact revision of one active restore. It is only a scheduling signal.
type CaptureRestoreProgress struct {
	Revision int64
	Owners   map[string]CaptureRestoreOwnerProgress
	root     domain.CaptureIdentity
	ticket   string
	admitted bool
}

type CaptureRestoreOwnerProgress struct {
	Member             bool
	SourceRead         bool
	Discovery          CaptureRestoreDiscoveryState
	Position           string
	readReady          bool
	readAttempted      bool
	discoveryAttempted bool
}

type CaptureRestoreDiscoveryState uint8

const (
	CaptureRestoreNoDiscovery CaptureRestoreDiscoveryState = iota
	CaptureRestoreDiscoveryMissing
	CaptureRestoreDiscoveryAdvancing
	CaptureRestoreDiscoveryComplete
)

func (s *Store) CaptureRestoreProgress(ctx context.Context) (result map[string]CaptureRestoreProgress, returnErr error) {
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { returnErr = errors.Join(returnErr, conn.Close()) }()
	// Use a deferred read transaction. The registry's configured transaction
	// default is immediate for writers, which a progress read does not need.
	if _, err := conn.ExecContext(ctx, `BEGIN`); err != nil {
		return nil, err
	}
	defer func() {
		_, err := conn.ExecContext(context.Background(), `ROLLBACK`)
		returnErr = errors.Join(returnErr, err)
	}()
	result, err = s.readRestoreProgress(ctx, conn)
	if err != nil || len(result) == 0 {
		return result, err
	}
	return result, s.readRestoreProgressMembership(ctx, conn, result)
}

func (s *Store) readRestoreProgress(ctx context.Context, conn *sql.Conn) (map[string]CaptureRestoreProgress, error) {
	// Both kinds of evidence are scoped to the restoring control and its exact
	// revision. The join reads one SQLite snapshot during worker admission.
	const query = `WITH owners AS (
	SELECT capture_id,revision,identity FROM capture_restore_progress WHERE worktree_id=?
	UNION SELECT capture_id,revision,identity FROM capture_restore_attempts WHERE worktree_id=?
)
SELECT c.capture_id,c.revision,c.source,c.store,c.session_id,coalesce(o.identity,''),coalesce(p.position,''),coalesce(p.discovery,0),coalesce(p.complete,0),coalesce(p.source_read,0),
       r.identity IS NOT NULL,coalesce(a.source_read,0),coalesce(a.discovery,0)
FROM capture_controls c
LEFT JOIN owners o ON o.capture_id=c.capture_id AND o.revision=c.revision
LEFT JOIN capture_restore_progress p ON p.worktree_id=c.worktree_id AND p.capture_id=o.capture_id AND p.revision=o.revision AND p.identity=o.identity
LEFT JOIN capture_restore_reads r ON r.worktree_id=c.worktree_id AND r.capture_id=o.capture_id AND r.revision=o.revision AND r.identity=o.identity
LEFT JOIN capture_restore_attempts a ON a.worktree_id=c.worktree_id AND a.capture_id=o.capture_id AND a.revision=o.revision AND a.identity=o.identity
WHERE c.worktree_id=? AND c.state=?`
	rows, err := conn.QueryContext(ctx, query, s.worktreeID, s.worktreeID, s.worktreeID, CaptureControlRestoring)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]CaptureRestoreProgress{}
	for rows.Next() {
		var captureID, identity, position string
		var root domain.CaptureIdentity
		var revision int64
		var discovery, complete bool
		var owner CaptureRestoreOwnerProgress
		if err := rows.Scan(&captureID, &revision, &root.Source, &root.Store, &root.SessionID, &identity, &position, &discovery, &complete, &owner.SourceRead, &owner.readReady, &owner.readAttempted, &owner.discoveryAttempted); err != nil {
			return nil, err
		}
		progress, ok := result[captureID]
		if !ok {
			progress = CaptureRestoreProgress{Revision: revision, root: root, Owners: map[string]CaptureRestoreOwnerProgress{}}
			result[captureID] = progress
		}
		if identity == "" {
			continue
		}
		if discovery {
			owner.Position = position
			switch {
			case complete:
				owner.Discovery = CaptureRestoreDiscoveryComplete
			case position != "":
				owner.Discovery = CaptureRestoreDiscoveryAdvancing
			default:
				owner.Discovery = CaptureRestoreDiscoveryMissing
			}
		}
		progress.Owners[identity] = owner
	}
	return result, rows.Err()
}

func (s *Store) readRestoreProgressMembership(ctx context.Context, conn *sql.Conn, progress map[string]CaptureRestoreProgress) error {
	// Inheritance can invalidate the publishing pass before it records read
	// or cursor receipts. Preserve that real advance in the successor decision
	// without loading any owner's session metrics into the worker poll.
	rows, err := conn.QueryContext(ctx, `SELECT source,store,session_id,ticket,coalesce(json_extract(record,'$.registration.parent_id'),''),coalesce(json_extract(record,'$.registration.binding'),'') FROM registrations WHERE worktree_id=? AND json_extract(record,'$.registration.project_directory')=?`, s.worktreeID, s.project)
	if err != nil {
		return err
	}
	defer rows.Close()
	var registrations []domain.CaptureRegistration
	for rows.Next() {
		row := domain.CaptureRegistration{ProjectDirectory: s.project}
		if err := rows.Scan(&row.Identity.Source, &row.Identity.Store, &row.Identity.SessionID, &row.TicketID, &row.ParentID, &row.Binding); err != nil {
			return err
		}
		registrations = append(registrations, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rootIDs := make(map[string]string, len(progress))
	roots := make(map[string]bool, len(progress))
	for id, entry := range progress {
		root := IdentityKey(entry.root)
		rootIDs[root] = id
		roots[root] = true
	}
	for _, row := range registrations {
		if id, ok := rootIDs[IdentityKey(row.Identity)]; ok {
			entry := progress[id]
			entry.ticket = row.TicketID
			entry.admitted = true
			progress[id] = entry
		}
	}
	for owner, root := range captureGroups(s.project, registrations, roots) {
		if id, ok := rootIDs[root]; ok {
			member := progress[id].Owners[owner]
			member.Member = true
			progress[id].Owners[owner] = member
		}
	}
	return nil
}

// CaptureRestoreProgressAdvanced reports only useful new evidence. OpenCode
// scan replays and unrelated publication work cannot extend an idle worker.
func CaptureRestoreProgressAdvanced(before, after map[string]CaptureRestoreProgress) bool {
	for id, current := range after {
		previous := before[id]
		if previous.Revision != current.Revision {
			continue
		}
		for key, owner := range current.Owners {
			prior := previous.Owners[key]
			if owner.Member && !prior.Member {
				return true
			}
			if owner.SourceRead && !prior.SourceRead {
				return true
			}
			if (owner.readAttempted && !prior.readAttempted) || (owner.discoveryAttempted && !prior.discoveryAttempted) {
				return true
			}
			if owner.Discovery >= CaptureRestoreDiscoveryAdvancing && owner.Discovery > prior.Discovery {
				return true
			}
			if owner.Discovery == CaptureRestoreDiscoveryAdvancing && prior.Discovery == owner.Discovery && owner.Position > prior.Position {
				return true
			}
		}
	}
	return false
}

type restorePriorities struct {
	reads     map[string]int
	discovery map[string]int
}

func (s *Store) restoreWorkPriorities(ctx context.Context, ticket string) (*restorePriorities, error) {
	progress, err := s.CaptureRestoreProgress(ctx)
	if err != nil {
		return nil, err
	}
	var priorities *restorePriorities
	for _, entry := range progress {
		if !entry.admitted || entry.ticket != ticket {
			continue
		}
		if priorities == nil {
			priorities = &restorePriorities{reads: map[string]int{}, discovery: map[string]int{}}
		}
		for key, owner := range entry.Owners {
			if !owner.Member {
				continue
			}
			if !owner.readReady {
				priorities.reads[key] = restoreAttemptPriority(owner.readAttempted)
			}
			if entry.root.Source == domain.SourceCodex || owner.Discovery != CaptureRestoreDiscoveryComplete {
				priorities.discovery[key] = restoreAttemptPriority(owner.discoveryAttempted)
			}
		}
	}
	return priorities, nil
}

func restoreAttemptPriority(attempted bool) int {
	if attempted {
		return 1
	}
	return 2
}
