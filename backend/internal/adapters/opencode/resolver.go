package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func resolveV2Root(db *sql.DB, nonce, ticketID, explicit string) (*string, error) {
	nonce, ticketID, explicit = strings.TrimSpace(nonce), strings.TrimSpace(ticketID), strings.TrimSpace(explicit)
	if nonce != "" {
		var id string
		err := db.QueryRow(`SELECT DISTINCT s.id FROM session_v2 s JOIN session_message m ON m.session_id=s.id WHERE s.parent_id IS NULL AND m.data LIKE ? ORDER BY m.time_created DESC LIMIT 1`, "%"+nonce+"%").Scan(&id)
		if err == nil {
			return &id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving v2 session by nonce: %w", err)
		}
	}
	if explicit != "" {
		return &explicit, nil
	}
	if ticketID != "" {
		var id string
		err := db.QueryRow(`SELECT id FROM session_v2 WHERE parent_id IS NULL AND title LIKE ? ORDER BY time_updated DESC LIMIT 1`, "%"+ticketID+"%").Scan(&id)
		if err == nil {
			return &id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving v2 session by ticket id: %w", err)
		}
		workspaceID, err := resolveV2Workspace(db, ticketID, 0, 0)
		if err != nil {
			return nil, err
		}
		if workspaceID != nil {
			return workspaceID, nil
		}
	}
	return nil, fmt.Errorf("%w: unable to resolve OpenCode v2 session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, ticketID)
}

func resolveV2Workspace(db *sql.DB, ticket string, minTime, maxTime int64) (*string, error) {
	ticket = strings.TrimSpace(ticket)
	if ticket == "" {
		return nil, nil
	}
	p1, p2 := "%workspaces/"+ticket+"/%", "%workspaces/"+ticket+`"%`
	query := `SELECT DISTINCT s.id FROM session_v2 s JOIN session_message m ON m.session_id=s.id WHERE s.parent_id IS NULL AND (m.data LIKE ? OR m.data LIKE ?)`
	var id string
	var err error
	if minTime > 0 && maxTime > 0 {
		query += ` AND s.time_created>=? AND s.time_created<=? ORDER BY m.time_created ASC LIMIT 1`
		err = db.QueryRow(query, p1, p2, minTime, maxTime).Scan(&id)
	} else {
		query += ` ORDER BY m.time_created ASC LIMIT 1`
		err = db.QueryRow(query, p1, p2).Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolving v2 session by workspace path: %w", err)
	}
	return &id, nil
}

// ErrSessionNotFound is returned when an OpenCode root session cannot be resolved through nonce, explicit ID, or ticket identifiers.
var ErrSessionNotFound = errors.New("opencode root session not found")

// ResolveActiveRootSession resolves the active root session identifier using prioritized matching strategies:
// Priority 1: Exact Nonce matching against the part table data.
// Priority 2: Explicit session ID passed directly as parameter.
// Priority 3: Ticket ID matching against root session titles. If title matching yields no rows, falls back to workspace path matching via ResolveSessionByWorkspacePath.
// If all strategies fail, returns nil and an error wrapping ErrSessionNotFound.
func ResolveActiveRootSession(db *sql.DB, nonce string, ticketID string, explicitSessionID string) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	family, err := detectSchema(db)
	if err != nil {
		return nil, err
	}
	if family == familyV2 {
		return resolveV2Root(db, nonce, ticketID, explicitSessionID)
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

		fallbackID, fallbackErr := ResolveSessionByWorkspacePath(db, cleanTicket, 0, 0)
		if fallbackErr != nil {
			return nil, fmt.Errorf("resolving session by workspace path fallback: %w", fallbackErr)
		}
		if fallbackID != nil && strings.TrimSpace(*fallbackID) != "" {
			return fallbackID, nil
		}
	}

	return nil, fmt.Errorf("%w: unable to resolve OpenCode session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, cleanTicket)
}

// ResolveSessionByWorkspacePath searches for root sessions whose parts reference the workspace directory,
// constrained by an optional timestamp window (minTimeMs, maxTimeMs) to prevent matching conversations outside the commit range.
func ResolveSessionByWorkspacePath(db *sql.DB, ticketID string, minTimeMs int64, maxTimeMs int64) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	family, err := detectSchema(db)
	if err != nil {
		return nil, err
	}
	if family == familyV2 {
		return resolveV2Workspace(db, ticketID, minTimeMs, maxTimeMs)
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
	err = nil
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
