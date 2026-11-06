package opencode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"agenticlens/internal/domain"
)

// CaptureSessionSnapshot reads one exact owner from the selected schema family.
// The caller supplies a dedicated read-only DB. A single connection keeps the
// preflight and extraction in one consistent SQLite read transaction.
func CaptureSessionSnapshot(ctx context.Context, db *sql.DB, id, user string, budget int64, source ...*domain.SourceEvents) ([]domain.SessionFact, int64, string, error) {
	return CaptureSessionSnapshotWithPolicyAt(ctx, db, id, user, budget, nil, time.Now().UTC(), source...)
}

func CaptureSessionSnapshotAt(ctx context.Context, db *sql.DB, id, user string, budget int64, cutoff time.Time, source ...*domain.SourceEvents) ([]domain.SessionFact, int64, string, error) {
	return CaptureSessionSnapshotWithPolicyAt(ctx, db, id, user, budget, nil, cutoff, source...)
}

func CaptureSessionSnapshotWithPolicy(ctx context.Context, db *sql.DB, id, user string, budget int64, policy *domain.GenerationPolicy, source ...*domain.SourceEvents) ([]domain.SessionFact, int64, string, error) {
	return CaptureSessionSnapshotWithPolicyAt(ctx, db, id, user, budget, policy, time.Now().UTC(), source...)
}

// CaptureSessionSnapshotWithPolicyAt keeps optional analysis and duration cutoff request-local.
func CaptureSessionSnapshotWithPolicyAt(ctx context.Context, db *sql.DB, id, user string, budget int64, policy *domain.GenerationPolicy, cutoff time.Time, source ...*domain.SourceEvents) (sessions []domain.SessionFact, bytes int64, updated string, err error) {
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
	var events []domain.EventFact
	var checkpointEvidence []domain.CheckpointInvocationEvidence
	var sink *[]domain.EventFact
	if len(source) > 0 {
		sink = &events
	}
	if family == schemaFamilyV2 {
		sessions, _, _, err = extractSchemaV2SessionsWithCheckpointAt(db, []string{id}, user, nil, true, sink, policy, cutoff, &checkpointEvidence, true)
	} else {
		sessions, _, _, err = extractSchemaV1SessionsWithCheckpointAt(db, []string{id}, user, nil, true, sink, policy, cutoff, &checkpointEvidence, true)
	}
	if err == nil && len(source) > 0 {
		fresh := domain.SourceEvents{Events: events, CheckpointEvidence: checkpointEvidence}
		for _, owner := range sessions {
			fresh.CompleteSessionIDs = append(fresh.CompleteSessionIDs, owner.ID)
		}
		*source[0] = domain.MergeSourceEvents(*source[0], fresh)
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

// CaptureProjectDirectory reads authoritative per-session directory metadata.
// Older layouts without it remain unresolved rather than borrowing caller cwd.
func CaptureProjectDirectory(ctx context.Context, db *sql.DB, id string) (string, error) {
	family, err := captureSchemaFamily(ctx, db, id)
	if err != nil {
		return "", err
	}
	table := "session"
	if family == schemaFamilyV2 {
		table = "session_v2"
	}
	columns, err := tableColumns(db)
	if err != nil {
		return "", err
	}
	if !columns[table]["directory"] {
		return "", errors.New("source schema has no authoritative session directory")
	}
	var directory string
	err = db.QueryRowContext(ctx, `SELECT directory FROM `+table+` WHERE id=?`, id).Scan(&directory)
	return directory, err
}

// CaptureSourceFingerprint probes cumulative source revision without decoding bodies.
func CaptureSourceFingerprint(ctx context.Context, db *sql.DB, id string) (string, error) {
	family, err := captureSchemaFamily(ctx, db, id)
	if err != nil {
		return "", err
	}
	table := "session"
	if family == schemaFamilyV2 {
		table = "session_v2"
	}
	var updated int64
	if err = db.QueryRowContext(ctx, `SELECT time_updated FROM `+table+` WHERE id=?`, id).Scan(&updated); err != nil {
		return "", err
	}
	result := fmt.Sprint(updated)
	tables := []string{"message", "part"}
	if family == schemaFamilyV2 {
		tables = []string{"session_message"}
	}
	for _, table := range tables {
		var count, size, stamp int64
		if err = db.QueryRowContext(ctx, `SELECT count(*),coalesce(sum(length(data)),0),coalesce(max(time_updated),0) FROM `+table+` WHERE session_id=?`, id).Scan(&count, &size, &stamp); err != nil {
			return "", err
		}
		result += fmt.Sprintf(":%d:%d:%d", count, size, stamp)
	}
	return result, nil
}

// CaptureChildFinished reads source-owned native task completion for worker
// liveness. A completed assistant message alone never establishes idle activity.
// The bounded latest-part window may miss older tasks; missing proof stays active.
func CaptureChildFinished(ctx context.Context, db *sql.DB, parent, child string) (bool, error) {
	var finished int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM (SELECT data FROM (
 SELECT data,time_updated FROM part WHERE session_id=? AND length(data)<=65536 ORDER BY time_updated DESC LIMIT 64
 ) WHERE json_valid(data) AND json_extract(data,'$.tool')='task'
 AND json_extract(data,'$.state.metadata.sessionId')=?
 ORDER BY time_updated DESC LIMIT 1) WHERE json_extract(data,'$.state.status')='completed'
 AND EXISTS (SELECT 1 FROM session WHERE id=? AND parent_id=?)`, parent, child, child, parent).Scan(&finished)
	return finished > 0, err
}
