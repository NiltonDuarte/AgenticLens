package live

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
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
	revisionMu    sync.Mutex
	revisionConn  *sql.Conn
	revisionEpoch uint64
}

// OpenReadOnlyHook opens a read-only SQLite database connection configured with non-blocking pragmas.
func OpenReadOnlyHook(dbPath string, busyTimeoutMs int) (*SQLiteHook, error) {
	if busyTimeoutMs <= 0 {
		busyTimeoutMs = 5000
	}

	cleanPath := strings.TrimSpace(dbPath)
	if strings.HasPrefix(cleanPath, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			cleanPath = filepath.Join(home, cleanPath[2:])
		}
	}

	isMemory := cleanPath == ":memory:" || strings.Contains(cleanPath, "mode=memory")
	if !isMemory {
		if _, err := os.Stat(cleanPath); err != nil {
			return nil, fmt.Errorf("verifying database path %s: %w", cleanPath, err)
		}
	}

	var dsn string
	if strings.HasPrefix(cleanPath, "file:") {
		dsn = cleanPath
		if !strings.Contains(dsn, "mode=ro") {
			if strings.Contains(dsn, "?") {
				dsn += "&mode=ro"
			} else {
				dsn += "?mode=ro"
			}
		}
		if !strings.Contains(dsn, "_busy_timeout") {
			dsn += fmt.Sprintf("&_busy_timeout=%d", busyTimeoutMs)
		}
	} else if isMemory {
		dsn = cleanPath
	} else {
		dsn = fmt.Sprintf("file:%s?mode=ro&_busy_timeout=%d", cleanPath, busyTimeoutMs)
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening read-only sqlite database at %s: %w", cleanPath, err)
	}

	// Connection pool bounds: reader connections only
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	// Execute read-only, non-blocking pragmas
	pragmas := []string{
		"PRAGMA query_only = ON;",
		fmt.Sprintf("PRAGMA busy_timeout = %d;", busyTimeoutMs),
		"PRAGMA read_uncommitted = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("executing pragma %q: %w", pragma, err)
		}
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging sqlite database at %s: %w", cleanPath, err)
	}

	return &SQLiteHook{
		db:            db,
		dbPath:        cleanPath,
		busyTimeoutMs: busyTimeoutMs,
	}, nil
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
	h.revisionMu.Lock()
	defer h.revisionMu.Unlock()
	if h.db == nil {
		return SourceRevision{}, errors.New("database handle is nil")
	}
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
		closeErr := h.revisionConn.Close()
		h.revisionConn = nil
		return SourceRevision{}, fmt.Errorf("reading source revision: %w", errors.Join(err, closeErr))
	}
	return SourceRevision{Connection: h.revisionEpoch, Version: version}, nil
}

// Close closes the observer before its underlying SQLite connection pool.
func (h *SQLiteHook) Close() error {
	h.revisionMu.Lock()
	defer h.revisionMu.Unlock()
	var observerErr error
	if h.revisionConn != nil {
		observerErr = h.revisionConn.Close()
		h.revisionConn = nil
	}
	if h.db != nil {
		return errors.Join(observerErr, h.db.Close())
	}
	return observerErr
}

// SessionExists checks if a session with the given ID exists in the database.
func (h *SQLiteHook) SessionExists(sessionID string) (bool, error) {
	if h.db == nil {
		return false, errors.New("database handle is nil")
	}

	var count int
	err := h.db.QueryRow("SELECT COUNT(1) FROM session WHERE id = ?", sessionID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking session existence for %s: %w", sessionID, err)
	}
	return count > 0, nil
}

// FetchActiveSessions retrieves active sessions updated within the threshold window,
// along with any direct parents or children.
func (h *SQLiteHook) FetchActiveSessions(activeThresholdSeconds int64, nowMs int64) ([]RawSessionRecord, error) {
	if h.db == nil {
		return nil, errors.New("database handle is nil")
	}

	var rows *sql.Rows
	var err error

	if activeThresholdSeconds > 0 && nowMs > 0 {
		cutoffMs := nowMs - (activeThresholdSeconds * 1000)
		query := `
			SELECT id, parent_id, agent, model, title, time_created, time_updated
			FROM session
			WHERE time_updated >= ? OR time_created >= ?
			   OR id IN (SELECT parent_id FROM session WHERE (time_updated >= ? OR time_created >= ?) AND parent_id IS NOT NULL)
			   OR parent_id IN (SELECT id FROM session WHERE (time_updated >= ? OR time_created >= ?))
			ORDER BY time_created ASC
		`
		rows, err = h.db.Query(query, cutoffMs, cutoffMs, cutoffMs, cutoffMs, cutoffMs, cutoffMs)
	} else {
		query := `
			SELECT id, parent_id, agent, model, title, time_created, time_updated
			FROM session
			ORDER BY time_created ASC
		`
		rows, err = h.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("querying active sessions: %w", err)
	}
	defer rows.Close()

	var records []RawSessionRecord
	for rows.Next() {
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

	return records, nil
}

// FetchSessionTrace extracts messages and parts for the given session ordered by time_created ascending.
func (h *SQLiteHook) FetchSessionTrace(sessionID string) ([]RawMessageRecord, []RawPartRecord, error) {
	if h.db == nil {
		return nil, nil, errors.New("database handle is nil")
	}

	msgTimeCol := "0 AS time_updated"
	if tableHasColumn(h.db, "message", "time_updated") {
		msgTimeCol = "m.time_updated"
	}

	partTimeCol := "0 AS time_updated"
	if tableHasColumn(h.db, "part", "time_updated") {
		partTimeCol = "p.time_updated"
	}

	// 1. Query messages
	qMsgs := fmt.Sprintf(`
		SELECT m.id, m.session_id, m.time_created, %s, m.data
		FROM message m
		WHERE m.session_id = ?
		ORDER BY m.time_created ASC
	`, msgTimeCol)

	msgRows, err := h.db.Query(qMsgs, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("querying messages for session %s: %w", sessionID, err)
	}
	defer msgRows.Close()

	var messages []RawMessageRecord
	for msgRows.Next() {
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

	// 2. Query parts
	qParts := fmt.Sprintf(`
		SELECT p.id, p.session_id, p.message_id, p.time_created, %s, p.data
		FROM part p
		WHERE p.session_id = ?
		ORDER BY p.time_created ASC
	`, partTimeCol)

	partRows, err := h.db.Query(qParts, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("querying parts for session %s: %w", sessionID, err)
	}
	defer partRows.Close()

	var parts []RawPartRecord
	for partRows.Next() {
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

	return messages, parts, nil
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
		rows, err := h.db.Query("SELECT id FROM session WHERE parent_id = ? ORDER BY time_created ASC", pID)
		if err != nil {
			return nil, fmt.Errorf("querying child sessions for parent %s: %w", pID, err)
		}

		for rows.Next() {
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
		rows.Close()
	}

	return childIDs, nil
}

func tableHasColumn(db *sql.DB, tableName, columnName string) bool {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err == nil {
			if strings.EqualFold(name, columnName) {
				return true
			}
		}
	}
	return false
}
