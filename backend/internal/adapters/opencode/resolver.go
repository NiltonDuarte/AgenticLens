package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrSessionNotFound = errors.New("opencode root session not found")
var ErrSessionAmbiguous = errors.New("opencode root session selection is ambiguous")

// ResolveActiveRootSession uses only explicit root IDs and literal nonces. The
// ticket ID provides diagnostic context; it never selects a source session.
// When both selectors are supplied, each must independently identify one root
// and both must agree.
func ResolveActiveRootSession(db *sql.DB, nonce, ticketID, explicitSessionID string) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	if strings.TrimSpace(nonce) == "" && strings.TrimSpace(explicitSessionID) == "" {
		return nil, fmt.Errorf("%w: ticket %q cannot select a session; supply --session-id with an existing root or --nonce with a unique literal token", ErrSessionNotFound, strings.TrimSpace(ticketID))
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return nil, err
	}
	table, records := resolutionTables(family)
	var explicitRoot *string
	if strings.TrimSpace(explicitSessionID) != "" {
		id, err := sanitizeSessionID(explicitSessionID)
		if err != nil {
			return nil, fmt.Errorf("invalid --session-id: %w", err)
		}
		rows, err := db.Query(`SELECT id FROM `+table+` WHERE id=? AND parent_id IS NULL LIMIT 2`, id)
		if err != nil {
			return nil, fmt.Errorf("OpenCode schema family %s: validating --session-id: %w", family, err)
		}
		explicitRoot, err = uniqueRoot(rows)
		if err != nil {
			return nil, fmt.Errorf("--session-id %q must identify exactly one existing root: %w", id, err)
		}
	}
	if strings.TrimSpace(nonce) == "" {
		return explicitRoot, nil
	}
	query := sessionTreeCTE(table, records) + `
SELECT DISTINCT tree.root_id FROM tree JOIN records ON records.session_id=tree.id
WHERE instr(records.data,?)>0 ORDER BY tree.root_id LIMIT 2`
	rows, err := db.Query(query, nonce)
	if err != nil {
		return nil, fmt.Errorf("OpenCode schema family %s: resolving literal --nonce: %w", family, err)
	}
	nonceRoot, err := uniqueRoot(rows)
	if err != nil {
		return nil, fmt.Errorf("--nonce must identify exactly one root in the selected OpenCode source: %w", err)
	}
	if explicitRoot != nil && *explicitRoot != *nonceRoot {
		return nil, fmt.Errorf("--session-id selects root %q but --nonce selects root %q; supply agreeing selectors", *explicitRoot, *nonceRoot)
	}
	return nonceRoot, nil
}

// SessionTable validates the selected store's complete supported layout before
// callers perform exact-owner metadata discovery.
func SessionTable(db *sql.DB) (string, error) {
	if db == nil {
		return "", errors.New("database connection is nil")
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return "", err
	}
	table, _ := resolutionTables(family)
	return table, nil
}

func resolutionTables(family schemaFamily) (string, string) {
	if family == schemaFamilyV2 {
		return "session_v2", `SELECT session_id,data FROM session_message`
	}
	return "session", `SELECT session_id,data FROM message UNION ALL SELECT session_id,data FROM part`
}

func sessionTreeCTE(table, records string) string {
	// UNION keeps traversal finite even for malformed duplicate/cyclic links.
	return `WITH RECURSIVE tree(id,root_id) AS (
 SELECT id,id FROM ` + table + ` WHERE parent_id IS NULL
 UNION
 SELECT child.id,tree.root_id FROM ` + table + ` child JOIN tree ON child.parent_id=tree.id
), records AS (` + records + `)
`
}

func uniqueRoot(rows *sql.Rows) (root *string, err error) {
	defer func() { err = errors.Join(err, rows.Close()) }()
	ids := make([]string, 0, 2)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		clean, err := sanitizeSessionID(id)
		if err != nil || clean != id {
			return nil, fmt.Errorf("source root %q has an invalid native session ID", id)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, ErrSessionNotFound
	}
	if len(ids) > 1 {
		return nil, fmt.Errorf("%w: roots %q and %q both match", ErrSessionAmbiguous, ids[0], ids[1])
	}
	return &ids[0], nil
}
