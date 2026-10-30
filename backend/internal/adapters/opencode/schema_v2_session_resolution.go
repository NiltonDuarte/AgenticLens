package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func resolveSchemaV2RootSession(db *sql.DB, nonce, ticketID, explicit string) (*string, error) {
	nonce, ticketID, explicit = strings.TrimSpace(nonce), strings.TrimSpace(ticketID), strings.TrimSpace(explicit)
	if nonce != "" {
		var id string
		err := db.QueryRow(`SELECT DISTINCT s.id FROM session_v2 s JOIN session_message m ON m.session_id=s.id WHERE s.parent_id IS NULL AND m.data LIKE ? ORDER BY m.time_created DESC LIMIT 1`, "%"+nonce+"%").Scan(&id)
		if err == nil {
			return &id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("resolving OpenCode schema family v2 session by nonce: %w", err)
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
			return nil, fmt.Errorf("resolving OpenCode schema family v2 session by ticket id: %w", err)
		}
		workspaceID, err := resolveSchemaV2WorkspacePath(db, ticketID, 0, 0)
		if err != nil {
			return nil, err
		}
		if workspaceID != nil {
			return workspaceID, nil
		}
	}
	return nil, fmt.Errorf("%w: unable to resolve OpenCode schema family v2 session for ticket %q; specify --session-id or use --nonce", ErrSessionNotFound, ticketID)
}

func resolveSchemaV2WorkspacePath(db *sql.DB, ticket string, minTime, maxTime int64) (*string, error) {
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
		return nil, fmt.Errorf("resolving OpenCode schema family v2 session by workspace path: %w", err)
	}
	return &id, nil
}
