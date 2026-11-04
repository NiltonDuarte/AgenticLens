package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	adaptersgit "agenticlens/internal/adapters/git"

	"agenticlens/internal/domain"
)

// RepositoryProof identifies an actual shared Git directory, not copied objects,
// remote URLs or a configured project label. It is private local provenance.
type RepositoryProof struct {
	CommonDirectory   string
	DirectoryIdentity string
}

// RepositoryProof verifies plumbing against the already validated physical
// ownership association. No source store or transcript is opened.
func (s *Store) RepositoryProof(ctx context.Context) (RepositoryProof, error) {
	if err := validateWorktreeOwnership(s.db, s.worktreeID, s.project); err != nil {
		return RepositoryProof{}, err
	}
	commonPath, err := adaptersgit.ResolveHistoryCommonDirectory(ctx, s.project)
	if err != nil {
		return RepositoryProof{}, fmt.Errorf("history repository proof unavailable: %w", err)
	}
	common, err := filepath.EvalSymlinks(commonPath)
	if err != nil {
		return RepositoryProof{}, err
	}
	identity, err := fileIdentity(common)
	if err != nil {
		return RepositoryProof{}, err
	}
	proof := RepositoryProof{common, identity}
	var recorded string
	if err = s.db.QueryRowContext(ctx, "SELECT ownership FROM worktree_ownership WHERE worktree_id=?", s.worktreeID).Scan(&recorded); err != nil {
		return RepositoryProof{}, err
	}
	var owner worktreeOwnership
	if err = json.Unmarshal([]byte(recorded), &owner); err != nil {
		return RepositoryProof{}, err
	}
	if owner.Repository != proof {
		return RepositoryProof{}, errors.New("history repository proof missing or changed; verify live ownership and explicitly recover provenance")
	}
	return proof, nil
}

func hasExplicitSourceBinding(r domain.CaptureRegistration) bool {
	return r.Binding == "explicit_agent_response" || r.Binding == "explicit_manual_selection" || r.Binding == "source_parent" && r.ParentID != ""
}

// RejectSessionAssociations rejects actual conflicting source/ticket bindings,
// not the existence or location of a receipt for the same portable session.
func (s *Store) RejectSessionAssociations(ctx context.Context, ticket, store string, owners []domain.SessionFact) error {
	for _, owner := range owners {
		rows, err := s.db.QueryContext(ctx, "SELECT store,ticket FROM registrations WHERE session_id=? AND source=? AND ticket<>'' AND ticket<>?", owner.ID, owner.Source, ticket)
		if err != nil {
			return err
		}
		for rows.Next() {
			var previousStore, boundTicket string
			if err = rows.Scan(&previousStore, &boundTicket); err != nil {
				return errors.Join(err, rows.Close())
			}
			same, compareErr := SameStore(previousStore, store)
			if compareErr != nil {
				return errors.Join(compareErr, rows.Close())
			}
			if same {
				return errors.Join(fmt.Errorf("session %s in selected source store is bound to ticket %s, refusing %s", owner.ID, boundTicket, ticket), rows.Close())
			}
		}
		if err = errors.Join(rows.Err(), rows.Close()); err != nil {
			return err
		}
	}
	return nil
}
