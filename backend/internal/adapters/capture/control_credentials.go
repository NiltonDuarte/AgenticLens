package capture

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"

	"agenticlens/internal/domain"
)

// The native binding exchange can disclose its own request secret. Control
// invitations use a separate, private per-worktree and per-owner signing seed.
func (s *Store) captureControlSeed(ctx context.Context, owner domain.CaptureIdentity, issue bool) (string, error) {
	identity := IdentityKey(owner)
	const read = `SELECT seed FROM capture_control_credentials WHERE worktree_id=? AND identity=?`
	var seed []byte
	err := s.db.QueryRowContext(ctx, read, s.worktreeID, identity).Scan(&seed)
	if errors.Is(err, sql.ErrNoRows) {
		if !issue {
			return "", nil
		}
		var candidate [sha256.Size]byte
		if _, err := rand.Read(candidate[:]); err != nil {
			return "", err
		}
		if _, err := s.db.ExecContext(ctx, `INSERT INTO capture_control_credentials (worktree_id,identity,seed) VALUES(?,?,?) ON CONFLICT(worktree_id,identity) DO NOTHING`, s.worktreeID, identity, candidate[:]); err != nil {
			return "", err
		}
		// Another invitation may already have chosen the durable seed.
		err = s.db.QueryRowContext(ctx, read, s.worktreeID, identity).Scan(&seed)
	}
	if err != nil {
		return "", err
	}
	if len(seed) != sha256.Size {
		return "", errors.New("invalid stored capture control signing seed")
	}
	return string(seed), nil
}
