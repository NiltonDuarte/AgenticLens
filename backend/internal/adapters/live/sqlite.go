package live

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"agenticlens/internal/adapters/opencode"
)

// RawSessionRecord represents an unmarshaled session row from SQLite.
type RawSessionRecord struct {
	ID          string
	ParentID    *string
	Agent       string
	Model       string
	Title       string
	TimeCreated int64
	TimeUpdated int64
}

// RawMessageRecord represents an unmarshaled message row from SQLite.
type RawMessageRecord struct {
	ID          string
	SessionID   string
	TimeCreated int64
	TimeUpdated int64
	Data        string
}

// RawPartRecord represents an unmarshaled part row from SQLite.
type RawPartRecord struct {
	ID          string
	SessionID   string
	MessageID   string
	TimeCreated int64
	TimeUpdated int64
	Data        string
}

// SQLiteHook provides a non-locking, read-only SQLite hook for querying live agent telemetry.
type SQLiteHook struct {
	db            *sql.DB
	dbPath        string
	busyTimeoutMs int
	revisionMu    readLock
	revisionConn  *sql.Conn
	revisionEpoch uint64
}

// OpenReadOnlyHook opens a read-only SQLite database connection configured with non-blocking pragmas.
func OpenReadOnlyHook(dbPath string, busyTimeoutMs int) (*SQLiteHook, error) {
	client, err := opencode.OpenDB(dbPath, busyTimeoutMs)
	if err != nil {
		return nil, err
	}
	db := client.DB()
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	return &SQLiteHook{db: db, dbPath: dbPath, busyTimeoutMs: busyTimeoutMs, revisionMu: make(readLock, 1)}, nil
}

// DB returns the underlying database handle for test introspection.
func (h *SQLiteHook) DB() *sql.DB {
	return h.db
}

// SourceRevision identifies changes committed through other SQLite connections.
// data_version is comparable only on the same connection, so retain one observer.
type SourceRevision struct {
	Connection uint64
	Version    int64
}

func (h *SQLiteHook) SourceRevision(ctx context.Context) (SourceRevision, error) {
	if h.db == nil {
		return SourceRevision{}, errors.New("database handle is nil")
	}
	if err := h.revisionMu.LockContext(ctx); err != nil {
		return SourceRevision{}, err
	}
	defer h.revisionMu.Unlock()
	if h.revisionConn == nil {
		conn, err := h.db.Conn(ctx)
		if err != nil {
			return SourceRevision{}, fmt.Errorf("opening source revision observer: %w", err)
		}
		h.revisionConn = conn
		h.revisionEpoch++
	}
	var version int64
	if err := h.revisionConn.QueryRowContext(ctx, "PRAGMA data_version").Scan(&version); err != nil {
		if cancelled := ctx.Err(); cancelled != nil {
			return SourceRevision{}, cancelled
		}
		closeErr := h.revisionConn.Close()
		h.revisionConn = nil
		return SourceRevision{}, fmt.Errorf("reading source revision: %w", errors.Join(err, closeErr))
	}
	return SourceRevision{Connection: h.revisionEpoch, Version: version}, ctx.Err()
}

// Close closes the observer before its underlying SQLite connection pool.
func (h *SQLiteHook) Close() error {
	if h.db == nil {
		return nil
	}
	h.revisionMu.Lock()
	defer h.revisionMu.Unlock()
	var observerErr error
	if h.revisionConn != nil {
		observerErr = h.revisionConn.Close()
		h.revisionConn = nil
	}
	return errors.Join(observerErr, h.db.Close())
}

// SessionExists checks if a session with the given ID exists in the database.
func (h *SQLiteHook) SessionExists(sessionID string) (bool, error) {
	return h.sessionExists(context.Background(), sessionID)
}

func (h *SQLiteHook) sessionExists(ctx context.Context, sessionID string) (bool, error) {
	if h.db == nil {
		return false, errors.New("database handle is nil")
	}

	var count int
	err := h.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM session WHERE id = ?", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking session existence for %s: %w", sessionID, err)
	}
	return count > 0, ctx.Err()
}

// FetchActiveSessions retrieves active sessions updated within the threshold window,
// along with any direct parents or children.
func (h *SQLiteHook) FetchActiveSessions(activeThresholdSeconds int64, nowMs int64) ([]RawSessionRecord, error) {
	return h.fetchActiveSessions(context.Background(), activeThresholdSeconds, nowMs)
}

func (h *SQLiteHook) fetchActiveSessions(ctx context.Context, activeThresholdSeconds int64, nowMs int64) ([]RawSessionRecord, error) {
	if activeThresholdSeconds <= 0 || nowMs <= 0 {
		return nil, errors.New("active session lookup requires a positive explicit activity window and current timestamp")
	}
	if h.db == nil {
		return nil, errors.New("database handle is nil")
	}

	cutoffMs := nowMs - (activeThresholdSeconds * 1000)
	query := `
		WITH active_sessions AS MATERIALIZED (
			SELECT id, parent_id, agent, model, title, time_created, time_updated
			FROM session
			WHERE time_updated >= ? OR time_created >= ?
			   OR id IN (SELECT parent_id FROM session WHERE (time_updated >= ? OR time_created >= ?) AND parent_id IS NOT NULL)
			   OR parent_id IN (SELECT id FROM session WHERE (time_updated >= ? OR time_created >= ?))
		)
		SELECT id, parent_id, agent, model, title, time_created, time_updated
		FROM active_sessions
		ORDER BY time_created ASC
	`
	rows, err := h.db.QueryContext(ctx, query, cutoffMs, cutoffMs, cutoffMs, cutoffMs, cutoffMs, cutoffMs)
	if err != nil {
		return nil, fmt.Errorf("querying active sessions: %w", err)
	}
	defer rows.Close()

	var records []RawSessionRecord
	for ctx.Err() == nil && rows.Next() {
		var id string
		var parentID, agent, model, title sql.NullString
		var timeCreated, timeUpdated sql.NullInt64

		if err := rows.Scan(&id, &parentID, &agent, &model, &title, &timeCreated, &timeUpdated); err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}

		var parentPtr *string
		if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
			val := strings.TrimSpace(parentID.String)
			parentPtr = &val
		}

		records = append(records, RawSessionRecord{
			ID:          id,
			ParentID:    parentPtr,
			Agent:       agent.String,
			Model:       model.String,
			Title:       title.String,
			TimeCreated: timeCreated.Int64,
			TimeUpdated: timeUpdated.Int64,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating active session rows: %w", err)
	}

	return records, ctx.Err()
}

// FetchSessionTrace extracts messages and parts for the given session ordered by time_created ascending.
func (h *SQLiteHook) FetchSessionTrace(sessionID string) ([]RawMessageRecord, []RawPartRecord, error) {
	return h.fetchSessionTrace(context.Background(), sessionID)
}

func (h *SQLiteHook) fetchSessionTrace(ctx context.Context, sessionID string) ([]RawMessageRecord, []RawPartRecord, error) {
	if h.db == nil {
		return nil, nil, errors.New("database handle is nil")
	}

	msgTimeCol := "0 AS time_updated"
	hasMsgTime, err := tableHasColumn(ctx, h.db, "message", "time_updated")
	if err != nil {
		return nil, nil, fmt.Errorf("inspecting message schema: %w", err)
	}
	if hasMsgTime {
		msgTimeCol = "m.time_updated"
	}

	partTimeCol := "0 AS time_updated"
	hasPartTime, err := tableHasColumn(ctx, h.db, "part", "time_updated")
	if err != nil {
		return nil, nil, fmt.Errorf("inspecting part schema: %w", err)
	}
	if hasPartTime {
		partTimeCol = "p.time_updated"
	}

	// 1. Query messages
	// The driver interrupts the first SQLite step of QueryContext, while
	// later Rows.Next calls cannot interrupt a native step already running.
	// Materialize source work in that first step and retain the output order.
	qMsgs := fmt.Sprintf(`
		WITH source_messages AS MATERIALIZED (
			SELECT m.id, m.session_id, m.time_created, %s, m.data
			FROM message m
			WHERE m.session_id = ?
		)
		SELECT id, session_id, time_created, time_updated, data
		FROM source_messages
		ORDER BY time_created ASC
	`, msgTimeCol)

	msgRows, err := h.db.QueryContext(ctx, qMsgs, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("querying messages for session %s: %w", sessionID, err)
	}
	defer msgRows.Close()

	var messages []RawMessageRecord
	for ctx.Err() == nil && msgRows.Next() {
		var id, sID string
		var tCreated, tUpdated sql.NullInt64
		var data sql.NullString

		if err := msgRows.Scan(&id, &sID, &tCreated, &tUpdated, &data); err != nil {
			return nil, nil, fmt.Errorf("scanning message row for session %s: %w", sessionID, err)
		}

		messages = append(messages, RawMessageRecord{
			ID:          id,
			SessionID:   sID,
			TimeCreated: tCreated.Int64,
			TimeUpdated: tUpdated.Int64,
			Data:        data.String,
		})
	}
	if err := msgRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating message rows for session %s: %w", sessionID, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	// 2. Query parts
	qParts := fmt.Sprintf(`
		WITH source_parts AS MATERIALIZED (
			SELECT p.id, p.session_id, p.message_id, p.time_created, %s, p.data
			FROM part p
			WHERE p.session_id = ?
		)
		SELECT id, session_id, message_id, time_created, time_updated, data
		FROM source_parts
		ORDER BY time_created ASC
	`, partTimeCol)

	partRows, err := h.db.QueryContext(ctx, qParts, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("querying parts for session %s: %w", sessionID, err)
	}
	defer partRows.Close()

	var parts []RawPartRecord
	for ctx.Err() == nil && partRows.Next() {
		var id, sID string
		var mID sql.NullString
		var tCreated, tUpdated sql.NullInt64
		var data sql.NullString

		if err := partRows.Scan(&id, &sID, &mID, &tCreated, &tUpdated, &data); err != nil {
			return nil, nil, fmt.Errorf("scanning part row for session %s: %w", sessionID, err)
		}

		parts = append(parts, RawPartRecord{
			ID:          id,
			SessionID:   sID,
			MessageID:   mID.String,
			TimeCreated: tCreated.Int64,
			TimeUpdated: tUpdated.Int64,
			Data:        data.String,
		})
	}
	if err := partRows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterating part rows for session %s: %w", sessionID, err)
	}

	return messages, parts, ctx.Err()
}

// FetchSessionTraceMulti extracts chronological messages and parts across multiple session IDs.
// Resulting messages and parts are stably ordered by TimeCreated ascending.
func (h *SQLiteHook) FetchSessionTraceMulti(sessionIDs []string) ([]RawMessageRecord, []RawPartRecord, error) {
	if h.db == nil {
		return nil, nil, errors.New("database handle is nil")
	}

	cleanIDs := make([]string, 0, len(sessionIDs))
	seen := make(map[string]bool)
	for _, id := range sessionIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			cleanIDs = append(cleanIDs, trimmed)
		}
	}

	if len(cleanIDs) == 0 {
		return []RawMessageRecord{}, []RawPartRecord{}, nil
	}

	var allMessages []RawMessageRecord
	var allParts []RawPartRecord

	for _, sID := range cleanIDs {
		msgs, parts, err := h.FetchSessionTrace(sID)
		if err != nil {
			return nil, nil, fmt.Errorf("fetching trace for session %s: %w", sID, err)
		}
		allMessages = append(allMessages, msgs...)
		allParts = append(allParts, parts...)
	}

	sort.SliceStable(allMessages, func(i, j int) bool {
		return allMessages[i].TimeCreated < allMessages[j].TimeCreated
	})

	sort.SliceStable(allParts, func(i, j int) bool {
		return allParts[i].TimeCreated < allParts[j].TimeCreated
	})

	return allMessages, allParts, nil
}

// ResolveChildSessionIDs queries child sub-session IDs where parent_id matches any of parentSessionIDs.
func (h *SQLiteHook) ResolveChildSessionIDs(parentSessionIDs []string) ([]string, error) {
	return h.resolveChildSessionIDs(context.Background(), parentSessionIDs)
}

func (h *SQLiteHook) resolveChildSessionIDs(ctx context.Context, parentSessionIDs []string) ([]string, error) {
	if h.db == nil {
		return nil, errors.New("database handle is nil")
	}

	cleanIDs := make([]string, 0, len(parentSessionIDs))
	seen := make(map[string]bool)
	for _, id := range parentSessionIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			cleanIDs = append(cleanIDs, trimmed)
		}
	}

	if len(cleanIDs) == 0 {
		return []string{}, nil
	}

	var childIDs []string
	childSeen := make(map[string]bool)

	for _, pID := range cleanIDs {
		rows, err := h.db.QueryContext(ctx, `
			WITH child_sessions AS MATERIALIZED (
				SELECT id, time_created FROM session WHERE parent_id = ?
			)
			SELECT id FROM child_sessions ORDER BY time_created ASC
		`, pID)
		if err != nil {
			return nil, fmt.Errorf("querying child sessions for parent %s: %w", pID, err)
		}

		for ctx.Err() == nil && rows.Next() {
			var cid string
			if err := rows.Scan(&cid); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scanning child session row: %w", err)
			}
			trimmedCid := strings.TrimSpace(cid)
			if trimmedCid != "" && !childSeen[trimmedCid] {
				childSeen[trimmedCid] = true
				childIDs = append(childIDs, trimmedCid)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterating child session rows: %w", err)
		}
		if err := rows.Close(); err != nil {
			return nil, fmt.Errorf("closing child session rows: %w", err)
		}
	}

	return childIDs, ctx.Err()
}

func tableHasColumn(ctx context.Context, db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for ctx.Err() == nil && rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
			if strings.EqualFold(name, columnName) {
				return true, ctx.Err()
			}
		}
	}
	return false, errors.Join(ctx.Err(), rows.Err())
}
