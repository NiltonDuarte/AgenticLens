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

// RecordOpenCodeHandoff preserves the exact native parent even while its ticket
// is unknown. A conflict cannot turn a previously observed child into a root.
func RecordOpenCodeHandoff(row domain.CaptureRegistration, project, parentID string) (domain.CaptureRegistration, error) {
	if row.Identity.Source != domain.SourceOpenCode || parentID == row.Identity.SessionID {
		return row, errors.New("invalid OpenCode source parent evidence")
	}
	if row.ParentID != "" && row.ParentID != parentID {
		return row, errors.New("OpenCode source parent conflicts with the recorded parent")
	}
	row.ParentID = parentID
	row.ProjectDirectory = project
	row.ProjectProof = domain.CaptureProjectVerified
	return row, nil
}

// InheritOpenCodeParent records a source-proven edge from the initial context
// request and inherits a ticket when the parent already has one.
func (s *Store) InheritOpenCodeParent(ctx context.Context, child domain.CaptureRegistration, parentID string) (domain.CaptureRegistration, error) {
	var ticket string
	if parentID != "" && child.TicketID == "" {
		parent, err := s.openCodeParentTicket(ctx, child.Identity, parentID)
		if err != nil {
			return child, err
		}
		ticket = parent
	}
	if ticket == "" && child.ParentID == parentID && child.ProjectProof == domain.CaptureProjectVerified && child.ProjectDirectory == s.project {
		return child, nil
	}
	return s.Change(ctx, child.Identity, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		current, err := RecordOpenCodeHandoff(current, s.project, parentID)
		if err != nil {
			return current, err
		}
		if current.TicketID == "" && ticket != "" {
			current.TicketID = ticket
			current.Binding = "source_parent"
			current.State = domain.CapturePending
		}
		return current, nil
	})
}

func (s *Store) openCodeParentTicket(ctx context.Context, child domain.CaptureIdentity, parentID string) (string, error) {
	parent := child
	parent.SessionID = parentID
	var ticket string
	err := s.db.QueryRowContext(ctx, "SELECT ticket FROM registrations WHERE worktree_id=? AND identity=?", s.worktreeID, IdentityKey(parent)).Scan(&ticket)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return ticket, err
}

// Native controls require current independent ownership even when a durable
// operation for this identity predates discovery of its native parent.
func (s *Store) openCodeControlRoot(ctx context.Context, row domain.CaptureRegistration) (bool, error) {
	if row.ParentID == "" {
		return true, nil
	}
	if !independentCapture(row, domain.CaptureRegistration{}) {
		return false, nil
	}
	ticket, err := s.openCodeParentTicket(ctx, row.Identity, row.ParentID)
	return err == nil && independentCapture(row, domain.CaptureRegistration{TicketID: ticket}), err
}
