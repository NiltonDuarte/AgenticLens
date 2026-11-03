package capture

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"agenticlens/internal/domain"
)

// ProveRegistration checks bounded exact source metadata after the durable
// handoff. Unflushed or conflicting sources remain locally unverified; they
// cannot establish project membership for advisory warnings or publication.
func (s *Store) ProveRegistration(ctx context.Context, row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
	if s.project == "" {
		return row, nil
	}
	proofCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	var proofErr error
	if row.Identity.Source == domain.SourceCodex && row.TranscriptPath == "" {
		proofErr = errors.New("exact transcript path has not been observed")
	} else {
		proofErr = s.validateOwner(proofCtx, row)
	}
	return s.Change(ctx, row.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if proofErr != nil {
			// Accepted project ownership remains a baseline fact when optional
			// metadata becomes unavailable. Latest proof stays explicitly
			// unverified and cannot qualify an unbound advisory or setup health.
			current.ProjectProof = domain.CaptureProjectUnverified
			current.Detail = "Minimal handoff accepted; latest exact source project metadata unavailable or conflicting. Prior accepted association/facts are preserved when known, but current project proof remains unverified. Retry after source flush or inspect configured source paths."
		} else {
			current.ProjectDirectory = s.project
			current.ProjectProof = domain.CaptureProjectVerified
		}
		return current, nil
	})
}

// RefreshProjectProofs retries only observed unbound handoffs, with a separate
// rotating metadata allowance. It cannot consume a ticket's extraction budget
// or discover unrelated sessions. Read-only status/check never calls it.
func (s *Store) RefreshProjectProofs(ctx context.Context, maxSessions int, maxBytes int64) error {
	if s.project == "" || maxSessions < 1 || maxBytes < MaxPayloadBytes {
		return nil
	}
	rows, err := s.List(ctx)
	if err != nil {
		return err
	}
	if _, err = s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS project_proof_cursor(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),identity TEXT NOT NULL,PRIMARY KEY(worktree_id,singleton))`); err != nil {
		return err
	}
	var cursor string
	err = s.db.QueryRowContext(ctx, `SELECT identity FROM project_proof_cursor WHERE worktree_id=? AND singleton=1`, s.worktreeID).Scan(&cursor)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	eligible := []domain.CaptureRegistration{}
	for _, row := range rows {
		if row.TicketID == "" && row.ProjectProof != domain.CaptureProjectVerified {
			eligible = append(eligible, row)
		}
	}
	split := 0
	for split < len(eligible) && IdentityKey(eligible[split].Identity) <= cursor {
		split++
	}
	eligible = append(append([]domain.CaptureRegistration{}, eligible[split:]...), eligible[:split]...)
	for _, row := range eligible {
		if ctx.Err() != nil {
			return nil
		}
		if maxSessions < 1 || maxBytes < MaxPayloadBytes {
			break
		}
		maxSessions--
		maxBytes -= MaxPayloadBytes
		if _, err = s.ProveRegistration(ctx, row); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if _, err = s.db.ExecContext(ctx, `INSERT INTO project_proof_cursor VALUES(?,1,?)ON CONFLICT(worktree_id,singleton)DO UPDATE SET identity=excluded.identity`, s.worktreeID, IdentityKey(row.Identity)); err != nil {
			return err
		}
	}
	return nil
}
