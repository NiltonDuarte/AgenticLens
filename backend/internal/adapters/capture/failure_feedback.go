package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"agenticlens/internal/domain"
)

// SessionFailureFeedback reads at most four relevant terminal outcomes for one
// stored exact owner/ticket. Pending observations cannot hide retained failures.
// It neither refreshes facts nor acknowledges/deletes failures or private detail.

func (s *Store) SessionFailureFeedback(ctx context.Context, identity domain.CaptureIdentity, byteBudget int64) (feedback string, returnErr error) {
	if err := identity.Validate(); err != nil {
		return "", err
	}
	identity, err := canonicalIdentity(identity)
	if err != nil {
		return "", err
	}
	var body []byte
	if err = s.db.QueryRowContext(ctx, `SELECT record FROM registrations WHERE worktree_id=? AND identity=?`, s.worktreeID, IdentityKey(identity)).Scan(&body); errors.Is(err, sql.ErrNoRows) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	registration, err := decode(body)
	if err != nil {
		return "", err
	}
	if registration.TicketID == "" || (s.project != "" && registration.ProjectDirectory != s.project) {
		return "", nil
	}
	_, complete, err := s.diagnosticMetadataComplete(ctx)
	if err != nil {
		return "", err
	}
	incomplete := "AgenticLens capture diagnostic evidence is incomplete. Run capture reconcile for the bound ticket, then capture diagnostics in this project before relying on current capture health."
	if !complete {
		return incomplete, nil
	}
	terminal := func(phase, owner, kind1, kind2 string) (int64, error) {
		var sequence int64
		err := s.db.QueryRowContext(ctx, `SELECT sequence FROM capture_diagnostic_heads WHERE worktree_id=? AND ticket=? AND phase=? AND owner=? AND kind IN(?,?) ORDER BY sequence DESC LIMIT 1`, s.worktreeID, registration.TicketID, phase, owner, kind1, kind2).Scan(&sequence)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return sequence, err
	}
	publication, err := terminal("publication", "", "failed", "committed")
	if err != nil {
		return "", err
	}
	source, err := terminal("reconciliation", IdentityKey(registration.Identity), "failed", "current")
	if err != nil {
		return "", err
	}
	generic, err := terminal("reconciliation", "", "generic_failed", "generic_failed")
	if err != nil {
		return "", err
	}
	if generic > source {
		source = generic
	}
	collection, err := terminal("reconciliation", CollectionAttemptOwner, "failed", "completed")
	if err != nil {
		return "", err
	}
	ambiguous, err := terminal("reconciliation", CollectionAttemptOwner, "latest", "latest")
	if err != nil {
		return "", err
	}
	diagnostics := []domain.CaptureDiagnostic{}
	if byteBudget <= 0 {
		return "", errors.New("capture diagnostic byte budget must be a positive explicit value")
	}
	remaining := byteBudget
	seen := map[int64]bool{}
	for _, sequence := range []int64{publication, source, collection, ambiguous} {
		if sequence == 0 || seen[sequence] {
			continue
		}
		seen[sequence] = true
		var size int64
		if err = s.db.QueryRowContext(ctx, `SELECT length(record) FROM capture_diagnostics WHERE worktree_id=? AND sequence=?`, s.worktreeID, sequence).Scan(&size); errors.Is(err, sql.ErrNoRows) {
			return incomplete, nil
		} else if err != nil {
			return "", err
		}
		if size > remaining {
			return incomplete, nil
		}
		var record []byte
		if err = s.db.QueryRowContext(ctx, `SELECT record FROM capture_diagnostics WHERE worktree_id=? AND sequence=? AND length(record)<=?`, s.worktreeID, sequence, remaining).Scan(&record); errors.Is(err, sql.ErrNoRows) {
			return incomplete, nil
		} else if err != nil {
			return "", err
		}
		var diagnostic domain.CaptureDiagnostic
		if err = json.Unmarshal(record, &diagnostic); err != nil {
			return "", err
		}
		diagnostic.Sequence = sequence
		diagnostics = append(diagnostics, diagnostic)
		remaining -= int64(len(record))
	}
	return domain.CaptureFailureFeedback(registration, diagnostics), nil
}
