package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ResolveActiveRootSession resolves the active root session identifier using prioritized matching strategies:
// Priority 1: Exact Nonce matching against the part table data.
// Priority 2: Explicit session ID passed directly as parameter.
// Priority 3: Ticket ID matching against root session titles.
// Priority 4: Fallback to the most recent message session parent or ID when neither nonce nor ticket ID is supplied.
func ResolveActiveRootSession(db *sql.DB, nonce string, ticketID string, explicitSessionID string) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	cleanNonce := strings.TrimSpace(nonce)
	cleanTicket := strings.TrimSpace(ticketID)
	cleanExplicit := strings.TrimSpace(explicitSessionID)

	// Priority 1: Exact Nonce Matching
	if cleanNonce != "" {
		query := `
			SELECT DISTINCT s.id
			FROM part p
			JOIN session s ON (p.session_id = s.id OR s.id = (SELECT parent_id FROM session WHERE id = p.session_id))
			WHERE s.parent_id IS NULL AND p.data LIKE ?
			ORDER BY p.time_created DESC LIMIT 1
		`
		var sessionID string
		err := db.QueryRow(query, "%"+cleanNonce+"%").Scan(&sessionID)
		if err == nil {
			return &sessionID, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving session by nonce: %w", err)
		}
	}

	// Priority 2: Explicit Session ID
	if cleanExplicit != "" {
		return &cleanExplicit, nil
	}

	// Priority 3: Ticket ID matching against root session title
	if cleanTicket != "" {
		query := `
			SELECT id
			FROM session
			WHERE parent_id IS NULL AND title LIKE ?
			ORDER BY time_updated DESC LIMIT 1
		`
		var sessionID string
		err := db.QueryRow(query, "%"+cleanTicket+"%").Scan(&sessionID)
		if err == nil {
			return &sessionID, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving session by ticket id: %w", err)
		}
	}

	// Priority 4: Most recent message fallback only if neither nonce nor ticket ID was specified
	if cleanNonce == "" && cleanTicket == "" {
		query := `
			SELECT s.id, s.parent_id
			FROM message m
			JOIN session s ON m.session_id = s.id
			ORDER BY m.time_created DESC LIMIT 1
		`
		var sid string
		var parentID sql.NullString
		err := db.QueryRow(query).Scan(&sid, &parentID)
		if err == nil {
			if parentID.Valid && strings.TrimSpace(parentID.String) != "" {
				res := strings.TrimSpace(parentID.String)
				return &res, nil
			}
			return &sid, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving session by most recent message: %w", err)
		}
	}

	return nil, nil
}

// ResolveSessionByWorkspacePath searches for root sessions whose parts reference the workspace directory,
// constrained by an optional timestamp window (minTimeMs, maxTimeMs) to prevent matching conversations outside the commit range.
func ResolveSessionByWorkspacePath(db *sql.DB, ticketID string, minTimeMs int64, maxTimeMs int64) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	cleanTicket := strings.TrimSpace(ticketID)
	if cleanTicket == "" {
		return nil, nil
	}

	pattern1 := "%workspaces/" + cleanTicket + "/%"
	pattern2 := "%workspaces/" + cleanTicket + "\"%"

	query := `
		SELECT DISTINCT s.id
		FROM part p
		JOIN session s ON (p.session_id = s.id OR s.id = (SELECT parent_id FROM session WHERE id = p.session_id))
		WHERE s.parent_id IS NULL
		  AND (p.data LIKE ? OR p.data LIKE ?)
	`

	var sessionID string
	var err error
	if minTimeMs > 0 && maxTimeMs > 0 {
		query += " AND s.time_created >= ? AND s.time_created <= ? ORDER BY p.time_created ASC LIMIT 1"
		err = db.QueryRow(query, pattern1, pattern2, minTimeMs, maxTimeMs).Scan(&sessionID)
	} else {
		query += " ORDER BY p.time_created ASC LIMIT 1"
		err = db.QueryRow(query, pattern1, pattern2).Scan(&sessionID)
	}
	if err == nil {
		return &sessionID, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return nil, fmt.Errorf("resolving session by workspace path: %w", err)
}
