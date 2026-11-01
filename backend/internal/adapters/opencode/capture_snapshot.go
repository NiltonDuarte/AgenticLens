package opencode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agenticlens/internal/domain"
)

// CaptureSessionSnapshot reads one exact owner from the selected schema family.
// The caller supplies a dedicated read-only DB. A single connection keeps the
// preflight and extraction in one consistent SQLite read transaction.
func CaptureSessionSnapshot(ctx context.Context, db *sql.DB, id, user string, budget int64) (sessions []domain.SessionFact, bytes int64, updated string, err error) {
	if _, err = sanitizeSessionID(id); err != nil {
		return
	}
	db.SetMaxOpenConns(1)
	if _, err = db.ExecContext(ctx, `BEGIN`); err != nil {
		return
	}
	defer func() {
		_, rollbackErr := db.ExecContext(context.Background(), `ROLLBACK`)
		err = errors.Join(err, rollbackErr)
	}()
	family, e := captureSchemaFamily(ctx, db, id)
	if e != nil {
		err = e
		return
	}
	table := "session"
	if family == schemaFamilyV2 {
		table = "session_v2"
	}
	var timestamp int64
	err = db.QueryRowContext(ctx, `SELECT time_updated FROM `+table+` WHERE id=?`, id).Scan(&timestamp)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return
	}
	if err != nil {
		return
	}
	updated = formatEpochMilli(timestamp)
	queries := []string{`SELECT coalesce(sum(length(data)+length(id)+64),0) FROM message WHERE session_id=?`, `SELECT coalesce(sum(length(data)+length(id)+64),0) FROM part WHERE session_id=?`}
	if family == schemaFamilyV2 {
		queries = []string{`SELECT coalesce(sum(length(data)+length(id)+64),0) FROM session_message WHERE session_id=?`}
	}
	bytes = 256
	for _, query := range queries {
		var size int64
		if err = db.QueryRowContext(ctx, query, id).Scan(&size); err != nil {
			return
		}
		bytes += size
	}
	if bytes > budget {
		err = fmt.Errorf("%w: session source snapshot requires %d bytes", ErrCaptureBudget, bytes)
		return
	}
	if family == schemaFamilyV2 {
		sessions, _, _, err = extractSchemaV2Sessions(db, []string{id}, user, nil, true)
	} else {
		sessions, _, _, err = extractSchemaV1Sessions(db, []string{id}, user, nil, true, true)
	}
	return
}

// ErrCaptureBudget marks pending work rather than failed or zero source data.
var ErrCaptureBudget = errors.New("capture source byte budget reached")

// captureSchemaFamily selects only by exact authoritative session membership.
// Mixed stores are safe only when one complete family owns the requested ID.
func captureSchemaFamily(ctx context.Context, db *sql.DB, id string) (schemaFamily, error) {
	columns, err := tableColumns(db)
	if err != nil {
		return "", err
	}
	matches := []schemaFamily{}
	for _, family := range []schemaFamily{schemaFamilyV1, schemaFamilyV2} {
		complete := true
		for table, names := range requiredColumns[family] {
			for _, name := range names {
				if !columns[table][name] {
					complete = false
				}
			}
		}
		if !complete {
			continue
		}
		table := "session"
		if family == schemaFamilyV2 {
			table = "session_v2"
		}
		var count int
		if err = db.QueryRowContext(ctx, `SELECT count(*) FROM `+table+` WHERE id=?`, id).Scan(&count); err != nil {
			return "", err
		}
		if count == 1 {
			matches = append(matches, family)
		}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		return "", errors.New("exact source session ID exists in multiple OpenCode families; refusing ambiguous extraction")
	}
	// No session has flushed. Validate a single family layout to distinguish pending
	// from unsupported, without selecting by title or directory.
	return detectSchemaFamily(db)
}

// CaptureChildIDs returns only source-proven native children of the exact parent.
func CaptureChildIDs(ctx context.Context, db *sql.DB, parent string, limit int) ([]string, error) {
	ids, _, err := CaptureChildPage(ctx, db, parent, "", limit)
	return ids, err
}

// CaptureChildPage advances by source ID before LIMIT, so known prefixes cannot
// permanently hide later native children under a bounded discovery allowance.
func CaptureChildPage(ctx context.Context, db *sql.DB, parent, after string, limit int) ([]string, bool, error) {
	if limit < 1 {
		return nil, false, errors.New("child discovery limit must be positive")
	}
	family, err := captureSchemaFamily(ctx, db, parent)
	if err != nil {
		return nil, false, err
	}
	table := "session"
	if family == schemaFamilyV2 {
		table = "session_v2"
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM `+table+` WHERE parent_id=? AND id>? ORDER BY id LIMIT ?`, parent, after, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, false, err
		}
		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, false, err
	}
	more := len(ids) > limit
	if more {
		ids = ids[:limit]
	}
	return ids, more, nil
}
