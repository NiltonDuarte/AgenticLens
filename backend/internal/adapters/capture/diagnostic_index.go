package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"agenticlens/internal/domain"
)

type diagnosticIndexState struct {
	cursor, upper, known, attemptKnown int64
	complete                           bool
}

// CollectionAttemptOwner is outside the hexadecimal exact-owner key space.
// Existing head kinds index aggregate failure/completion and ambiguous history.
const CollectionAttemptOwner = ":collection-attempt"

// This semantic watermark separates whole-attempt outcomes from owner reads.
// Imported history without it remains unknown until explicit bounded backfill.
const attemptIndexOwner = ":attempt-index"

func (s *Store) initializeDiagnosticIndex() error {
	return s.initializeDiagnosticIndexContext(context.Background())
}
func (s *Store) initializeDiagnosticIndexContext(ctx context.Context) error {
	if err := setConstructorBusyTimeout(ctx, s.db, 1000); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, diagnosticSchema); err != nil {
		return err
	}
	if err := setConstructorBusyTimeout(ctx, s.db, 1000); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO capture_diagnostic_index(worktree_id,singleton,cursor,upper_sequence,known_max,complete) SELECT ?,1,0,coalesce(max(sequence),0),coalesce(max(sequence),0),CASE WHEN count(*)=0 THEN 1 ELSE 0 END FROM (SELECT sequence FROM capture_diagnostics WHERE worktree_id=? ORDER BY sequence DESC LIMIT 1)`, s.worktreeID, s.worktreeID); err != nil {
		return err
	}
	if err := setConstructorBusyTimeout(ctx, s.db, 1000); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO capture_diagnostic_heads(worktree_id,ticket,phase,owner,kind,sequence) SELECT ?,'','index',?,'latest',0 WHERE NOT EXISTS(SELECT 1 FROM capture_diagnostics WHERE worktree_id=? LIMIT 1)`, s.worktreeID, attemptIndexOwner, s.worktreeID)
	return err
}

// beginDiagnosticWrite reserves the SQLite writer before reading metadata.
// Deferred read-to-write upgrades can return SQLITE_BUSY immediately when a
// concurrent callback commits; a first write uses the configured busy timeout.
func (s *Store) beginDiagnosticWrite(ctx context.Context) (*sql.Tx, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE capture_diagnostic_index SET known_max=known_max WHERE worktree_id=? AND singleton=1`, s.worktreeID); err != nil {
		return nil, errors.Join(err, tx.Rollback())
	}
	return tx, nil
}

func (s *Store) readDiagnosticIndexTx(ctx context.Context, tx *sql.Tx) (state diagnosticIndexState, err error) {
	var complete int
	err = tx.QueryRowContext(ctx, `SELECT cursor,upper_sequence,known_max,complete,coalesce((SELECT sequence FROM capture_diagnostic_heads WHERE worktree_id=? AND ticket='' AND phase='index' AND owner=? AND kind='latest'),-1) FROM capture_diagnostic_index WHERE worktree_id=? AND singleton=1`, s.worktreeID, attemptIndexOwner, s.worktreeID).Scan(&state.cursor, &state.upper, &state.known, &complete, &state.attemptKnown)
	state.complete = complete == 1
	return state, err
}
func (s *Store) detectDiagnosticIndexGap(ctx context.Context, tx *sql.Tx, state *diagnosticIndexState) error {
	var actual int64
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(max(sequence),0) FROM capture_diagnostics WHERE worktree_id=?`, s.worktreeID).Scan(&actual); err != nil {
		return err
	}
	if actual == state.known && actual == state.attemptKnown {
		return nil
	}
	if state.complete {
		state.cursor = min(state.known, state.attemptKnown)
	}
	if state.attemptKnown < 0 || actual < state.known || actual < state.attemptKnown {
		state.cursor = 0
		state.upper = actual
	} else if actual > state.upper {
		state.upper = actual
	}
	state.known = actual
	state.attemptKnown = actual
	state.complete = false
	return nil
}
func (s *Store) writeDiagnosticIndex(ctx context.Context, tx *sql.Tx, state diagnosticIndexState) error {
	complete := 0
	if state.complete {
		complete = 1
	}
	if _, err := tx.ExecContext(ctx, `UPDATE capture_diagnostic_index SET cursor=?,upper_sequence=?,known_max=?,complete=? WHERE worktree_id=? AND singleton=1`, state.cursor, state.upper, state.known, complete, s.worktreeID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO capture_diagnostic_heads(worktree_id,ticket,phase,owner,kind,sequence) VALUES(?,'','index',?,'latest',?) ON CONFLICT(worktree_id,ticket,phase,owner,kind) DO UPDATE SET sequence=excluded.sequence`, s.worktreeID, attemptIndexOwner, state.attemptKnown)
	return err
}
func (s *Store) indexDiagnostic(ctx context.Context, tx *sql.Tx, d domain.CaptureDiagnostic) error {
	head := func(owner, kind string) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO capture_diagnostic_heads(worktree_id,ticket,phase,owner,kind,sequence)VALUES(?,?,?,?,?,?) ON CONFLICT(worktree_id,ticket,phase,owner,kind) DO UPDATE SET sequence=excluded.sequence WHERE sequence<excluded.sequence`, s.worktreeID, d.TicketID, d.Phase, owner, kind, d.Sequence)
		return err
	}
	if err := head("", "latest"); err != nil {
		return err
	}
	switch d.Outcome {
	case "committed", "failed", "accepted", "completed":
		if err := head("", d.Outcome); err != nil {
			return err
		}
	}
	if d.Outcome == "failed" {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO capture_diagnostic_failures VALUES(?,?)`, s.worktreeID, d.Sequence); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM capture_diagnostic_failures WHERE worktree_id=? AND sequence NOT IN(SELECT sequence FROM capture_diagnostic_failures WHERE worktree_id=? ORDER BY sequence DESC LIMIT 32)`, s.worktreeID, s.worktreeID); err != nil {
			return err
		}
	}
	if d.Phase == "reconciliation" {
		switch domain.CaptureAttemptOutcome(d) {
		case domain.CaptureAttemptFailed:
			if err := head(CollectionAttemptOwner, "failed"); err != nil {
				return err
			}
		case domain.CaptureAttemptCompleted:
			if err := head(CollectionAttemptOwner, "completed"); err != nil {
				return err
			}
		case domain.CaptureAttemptUnknown:
			if err := head(CollectionAttemptOwner, "latest"); err != nil {
				return err
			}
		}

		if domain.CaptureAttemptOutcome(d) == domain.CaptureAttemptUnknown && d.Outcome == "failed" && (d.Reconciliation == nil || len(d.Reconciliation.Owners) == 0) {
			if err := head("", "generic_failed"); err != nil {
				return err
			}
		}
		if d.Reconciliation != nil {
			for _, owner := range d.Reconciliation.Owners {
				if err := ctx.Err(); err != nil {
					return err
				}
				if owner.State == domain.CaptureFailed {
					if err := head(IdentityKey(owner.Identity), "failed"); err != nil {
						return err
					}
				}
				if owner.State == domain.CaptureCurrent && owner.RefreshedAt != "" {
					if err := head(IdentityKey(owner.Identity), "current"); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
func (s *Store) pruneIndexedDiagnostics(ctx context.Context, tx *sql.Tx) error {
	// Current/terminal rows remain referenced independently of recent history.
	_, err := tx.ExecContext(ctx, `DELETE FROM capture_diagnostics WHERE worktree_id=? AND sequence NOT IN(SELECT sequence FROM capture_diagnostic_heads WHERE worktree_id=? UNION SELECT sequence FROM capture_diagnostic_failures WHERE worktree_id=?)`, s.worktreeID, s.worktreeID, s.worktreeID)
	return err
}

type DiagnosticIndexProgress struct {
	Complete bool  `json:"complete"`
	Cursor   int64 `json:"cursor"`
	Upper    int64 `json:"upper"`
	Rows     int   `json:"rows"`
	Bytes    int64 `json:"bytes"`
}

// BackfillDiagnosticIndex is explicit bounded mutable maintenance. It never runs
// during Open/Register. Each committed row advances a durable cursor; sequence
// merges cannot let an older backfill replace a newer terminal outcome.
func (s *Store) BackfillDiagnosticIndex(ctx context.Context, maxRows int, maxBytes int64) (progress DiagnosticIndexProgress, returnErr error) {
	if maxRows < 1 || maxBytes < 1 {
		return progress, errors.New("diagnostic index maintenance requires positive row/byte budgets")
	}
	for progress.Rows < maxRows {
		tx, err := s.beginDiagnosticWrite(ctx)
		if err != nil {
			return progress, err
		}
		state, err := s.readDiagnosticIndexTx(ctx, tx)
		if err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		if err = s.detectDiagnosticIndexGap(ctx, tx, &state); err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		progress.Cursor, progress.Upper, progress.Complete = state.cursor, state.upper, state.complete
		if state.complete {
			return progress, tx.Rollback()
		}
		var sequence, size int64
		err = tx.QueryRowContext(ctx, `SELECT sequence,length(record) FROM capture_diagnostics WHERE worktree_id=? AND sequence>? AND sequence<=? ORDER BY sequence LIMIT 1`, s.worktreeID, state.cursor, state.upper).Scan(&sequence, &size)
		if errors.Is(err, sql.ErrNoRows) {
			state.cursor = state.upper
			state.complete = true
			if _, err = tx.ExecContext(ctx, `DELETE FROM capture_diagnostic_heads WHERE worktree_id=? AND NOT(ticket='' AND phase='index' AND owner=':attempt-index' AND kind='latest') AND sequence NOT IN(SELECT sequence FROM capture_diagnostics WHERE worktree_id=?)`, s.worktreeID, s.worktreeID); err != nil {
				return progress, errors.Join(err, tx.Rollback())
			}
			if err = s.writeDiagnosticIndex(ctx, tx, state); err != nil {
				return progress, errors.Join(err, tx.Rollback())
			}
			if err = s.pruneIndexedDiagnostics(ctx, tx); err != nil {
				return progress, errors.Join(err, tx.Rollback())
			}
			if err = tx.Commit(); err != nil {
				return progress, err
			}
			progress.Cursor, progress.Complete = state.cursor, true
			return progress, nil
		}
		if err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		if size > maxBytes-progress.Bytes {
			return progress, tx.Rollback()
		}
		var body []byte
		if err = tx.QueryRowContext(ctx, `SELECT record FROM capture_diagnostics WHERE worktree_id=? AND sequence=? AND length(record)<=?`, s.worktreeID, sequence, maxBytes-progress.Bytes).Scan(&body); err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		var diagnostic domain.CaptureDiagnostic
		if err = json.Unmarshal(body, &diagnostic); err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		diagnostic.Sequence = sequence
		if err = s.indexDiagnostic(ctx, tx, diagnostic); err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		state.cursor = sequence
		if sequence == state.upper {
			state.complete = true
		}
		if err = s.writeDiagnosticIndex(ctx, tx, state); err != nil {
			return progress, errors.Join(err, tx.Rollback())
		}
		if state.complete {
			if err = s.pruneIndexedDiagnostics(ctx, tx); err != nil {
				return progress, errors.Join(err, tx.Rollback())
			}
		}
		if err = tx.Commit(); err != nil {
			return progress, err
		}
		progress.Rows++
		progress.Bytes += size
		progress.Cursor, progress.Upper, progress.Complete = state.cursor, state.upper, state.complete
		if state.complete {
			return progress, nil
		}
	}
	return progress, nil
}
