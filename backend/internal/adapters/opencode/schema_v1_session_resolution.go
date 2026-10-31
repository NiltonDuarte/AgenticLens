package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func resolveSchemaV1RootSession(db *sql.DB, nonce, ticketID, explicitSessionID string) (*string, error) {
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
			return nil, fmt.Errorf("OpenCode schema family v1: resolving session by nonce: %w", err)
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
			return nil, fmt.Errorf("OpenCode schema family v1: resolving session by ticket id: %w", err)
		}

		fallbackID, fallbackErr := resolveSchemaV1WorkspacePath(db, cleanTicket, 0, 0)
		if fallbackErr != nil {
			return nil, fmt.Errorf("OpenCode schema family v1: resolving session by workspace path fallback: %w", fallbackErr)
		}
		if fallbackID != nil && strings.TrimSpace(*fallbackID) != "" {
			return fallbackID, nil
		}
	}

	return nil, fmt.Errorf("OpenCode schema family v1: %w: unable to resolve OpenCode session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, cleanTicket)
}

func resolveSchemaV1WorkspacePath(db *sql.DB, ticketID string, minTimeMs, maxTimeMs int64) (*string, error) {
	var err error
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
	return nil, fmt.Errorf("OpenCode schema family v1: resolving session by workspace path: %w", err)
}
