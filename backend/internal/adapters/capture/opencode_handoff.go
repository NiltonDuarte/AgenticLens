package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// OpenCodeHandoffParent checks the session in the configured database before
// registry allocation. An unreadable snapshot stays pending for the worker.
// A readable session belonging to a different project cannot enter this scope.
func OpenCodeHandoffParent(ctx context.Context, cfg domain.ProjectConfig, project, store, id string) (parent string, pending error, conflict error) {
	db, err := openOpenCodeSource(cfg, store)
	if err != nil {
		return "", err, nil
	}
	directory, parent, err := opencode.CaptureSessionContext(ctx, db.DB(), id)
	err = errors.Join(err, db.Close())
	if err != nil {
		return "", err, nil
	}
	if err := ValidateProject(project, directory); err != nil {
		return "", nil, fmt.Errorf("OpenCode session %s directory %q and callback project %q: %w", id, directory, project, err)
	}
	return parent, nil, nil
}

// InheritOpenCodeParent answers a child's initial context request immediately
// when the exact database already proves a bound parent. Worker reconciliation
// handles delayed sources or a parent bound later.
func (s *Store) InheritOpenCodeParent(ctx context.Context, child domain.CaptureRegistration, parentID string) (domain.CaptureRegistration, error) {
	if parentID == "" || child.TicketID != "" {
		return child, nil
	}
	parentIdentity := child.Identity
	parentIdentity.SessionID = parentID
	var ticket string
	err := s.db.QueryRowContext(ctx, "SELECT ticket FROM registrations WHERE worktree_id=? AND identity=?", s.worktreeID, IdentityKey(parentIdentity)).Scan(&ticket)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && ticket == "") {
		return child, nil
	}
	if err != nil {
		return child, err
	}
	return s.Change(ctx, child.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if current.TicketID == "" {
			current.TicketID = ticket
			current.ParentID = parentID
			current.Binding = "source_parent"
			current.ProjectDirectory = s.project
			current.ProjectProof = domain.CaptureProjectVerified
			current.State = domain.CapturePending
		}
		return current, nil
	})
}
