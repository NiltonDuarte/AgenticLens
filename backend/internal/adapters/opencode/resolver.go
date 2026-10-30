package opencode

import (
	"database/sql"
	"errors"
)

// ErrSessionNotFound is returned when an OpenCode root session cannot be resolved through nonce, explicit ID, or ticket identifiers.
var ErrSessionNotFound = errors.New("opencode root session not found")

// ResolveActiveRootSession resolves the active root session identifier using prioritized matching strategies:
// Priority 1: Exact Nonce matching against stored message/tool records.
// Priority 2: Explicit session ID passed directly as parameter.
// Priority 3: Ticket ID matching against root session titles. If title matching yields no rows, falls back to workspace path matching via ResolveSessionByWorkspacePath.
// If all strategies fail, returns nil and an error wrapping ErrSessionNotFound.
func ResolveActiveRootSession(db *sql.DB, nonce string, ticketID string, explicitSessionID string) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return nil, err
	}
	if family == schemaFamilyV2 {
		return resolveSchemaV2RootSession(db, nonce, ticketID, explicitSessionID)
	}
	return resolveSchemaV1RootSession(db, nonce, ticketID, explicitSessionID)
}

// ResolveSessionByWorkspacePath searches for root sessions whose stored message/tool records reference the workspace directory,
// constrained by an optional timestamp window (minTimeMs, maxTimeMs) to prevent matching conversations outside the commit range.
func ResolveSessionByWorkspacePath(db *sql.DB, ticketID string, minTimeMs int64, maxTimeMs int64) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return nil, err
	}
	if family == schemaFamilyV2 {
		return resolveSchemaV2WorkspacePath(db, ticketID, minTimeMs, maxTimeMs)
	}
	return resolveSchemaV1WorkspacePath(db, ticketID, minTimeMs, maxTimeMs)
}
