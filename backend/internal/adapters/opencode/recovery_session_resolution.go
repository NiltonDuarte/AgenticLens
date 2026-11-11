package opencode

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// DiscoverRootSessionForRecovery is an explicitly requested heuristic, separate
// from direct generation. Callers supply evidence derived from their configured
// ticket workspace; no conventional root is invented. Ambiguous evidence never
// selects the newest root or falls through to another strategy.
func DiscoverRootSessionForRecovery(db *sql.DB, ticketID string, workspacePaths []string, minTimeMs, maxTimeMs int64) (*string, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}
	if strings.TrimSpace(ticketID) == "" {
		return nil, errors.New("explicit recovery requires a nonblank ticket ID")
	}
	if len(workspacePaths) == 0 {
		return nil, errors.New("explicit recovery requires configured ticket-workspace paths")
	}
	if minTimeMs < 0 || maxTimeMs < 0 || (minTimeMs == 0) != (maxTimeMs == 0) || minTimeMs > maxTimeMs {
		return nil, errors.New("explicit recovery time bounds must both be zero or an ordered positive range")
	}
	patterns := make([]string, 0, len(workspacePaths)*2)
	for i, workspace := range workspacePaths {
		if strings.TrimSpace(workspace) == "" {
			return nil, fmt.Errorf("explicit recovery workspace path %d must be nonblank", i)
		}
		workspace = strings.TrimRight(workspace, "/")
		if workspace == "" {
			return nil, fmt.Errorf("explicit recovery workspace path %d must identify the ticket workspace", i)
		}
		encoded, err := json.Marshal(workspace)
		if err != nil {
			return nil, fmt.Errorf("encoding configured recovery workspace: %w", err)
		}
		for _, literal := range []string{workspace, string(encoded[1 : len(encoded)-1])} {
			patterns = append(patterns, literal+"/", literal+`"`, literal+`\"`)
		}
	}
	encodedPatterns, err := json.Marshal(patterns)
	if err != nil {
		return nil, fmt.Errorf("encoding recovery workspace evidence: %w", err)
	}
	family, err := detectSchemaFamily(db)
	if err != nil {
		return nil, err
	}
	table, records := resolutionTables(family)
	query := sessionTreeCTE(table, records) + `
SELECT DISTINCT s.id FROM ` + table + ` s WHERE s.parent_id IS NULL
AND (?=0 OR s.time_created>=?) AND (?=0 OR s.time_created<=?)
AND (instr(s.title,?)>0 OR EXISTS (
 SELECT 1 FROM tree JOIN records ON records.session_id=tree.id
 WHERE tree.root_id=s.id AND EXISTS (
  SELECT 1 FROM json_each(?) evidence WHERE instr(records.data,evidence.value)>0
 ))) ORDER BY s.id LIMIT 2`
	rows, err := db.Query(query, minTimeMs, minTimeMs, maxTimeMs, maxTimeMs, ticketID, string(encodedPatterns))
	if err != nil {
		return nil, fmt.Errorf("OpenCode schema family %s: explicit recovery discovery: %w", family, err)
	}
	root, err := uniqueRoot(rows)
	if errors.Is(err, ErrSessionNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("explicit OpenCode recovery evidence must select one root: %w", err)
	}
	return root, nil
}
