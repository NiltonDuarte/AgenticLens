package opencode

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

type schemaFamily string

const (
	schemaFamilyV1 schemaFamily = "v1"
	schemaFamilyV2 schemaFamily = "v2"
)

var requiredColumns = map[schemaFamily]map[string][]string{
	schemaFamilyV1: {"session": {"id", "parent_id", "agent", "model", "title", "time_created", "time_updated"}, "message": {"id", "session_id", "time_created", "time_updated", "data"}, "part": {"id", "session_id", "message_id", "time_created", "time_updated", "data"}},
	schemaFamilyV2: {"session_v2": {"id", "parent_id", "agent", "model", "title", "time_created", "time_updated"}, "session_message": {"id", "session_id", "type", "seq", "time_created", "time_updated", "data"}},
}

func tableColumns(db *sql.DB) (map[string]map[string]bool, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("reading SQLite table inventory: %w", err)
	}
	defer rows.Close()
	out := make(map[string]map[string]bool)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("reading SQLite table name: %w", err)
		}
		out[table] = make(map[string]bool)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating SQLite table inventory: %w", err)
	}
	for table := range out {
		// Table names come from sqlite_master and are quoted as identifiers.
		rows, err := db.Query("PRAGMA table_info(\"" + strings.ReplaceAll(table, "\"", "\"\"") + "\")")
		if err != nil {
			return nil, fmt.Errorf("reading columns for table %s: %w", table, err)
		}
		for rows.Next() {
			var cid, notnull, pk int
			var name, typ string
			var def sql.NullString
			if err := rows.Scan(&cid, &name, &typ, &notnull, &def, &pk); err != nil {
				rows.Close()
				return nil, fmt.Errorf("reading columns for table %s: %w", table, err)
			}
			out[table][name] = true
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("iterating columns for table %s: %w", table, err)
		}
		rows.Close()
	}
	return out, nil
}

func detectSchemaFamily(db *sql.DB) (schemaFamily, error) {
	all, err := tableColumns(db)
	if err != nil {
		return "", err
	}
	complete := map[schemaFamily]bool{schemaFamilyV1: true, schemaFamilyV2: true}
	present := map[schemaFamily]bool{}
	var issues []string
	for family, tables := range requiredColumns {
		for table := range tables {
			if _, ok := all[table]; ok {
				present[family] = true
			}
		}
	}
	for family, tables := range requiredColumns {
		for table, cols := range tables {
			actual, exists := all[table]
			if !exists {
				complete[family] = false
				if present[family] {
					issues = append(issues, string(family)+"."+table+" (table missing)")
				}
				continue
			}
			missing := []string{}
			for _, col := range cols {
				if !actual[col] {
					missing = append(missing, col)
				}
			}
			if len(missing) > 0 {
				complete[family] = false
				sort.Strings(missing)
				issues = append(issues, string(family)+"."+table+"."+strings.Join(missing, ","))
			}
		}
	}
	if (present[schemaFamilyV1] && present[schemaFamilyV2]) || (complete[schemaFamilyV1] && present[schemaFamilyV2]) || (complete[schemaFamilyV2] && present[schemaFamilyV1]) {
		observed := make([]string, 0, len(all))
		for name := range all {
			observed = append(observed, name)
		}
		sort.Strings(observed)
		return "", fmt.Errorf("unsupported mixed OpenCode schema family layout: observed tables [%s], incomplete structures [%s]; v1 requires session/message/part and v2 requires session_v2/session_message. Inspect or select a clean database", strings.Join(observed, ", "), strings.Join(issues, "; "))
	}
	if complete[schemaFamilyV1] {
		return schemaFamilyV1, nil
	}
	if complete[schemaFamilyV2] {
		return schemaFamilyV2, nil
	}
	observed := make([]string, 0, len(all))
	for name := range all {
		observed = append(observed, name)
	}
	sort.Strings(observed)
	return "", fmt.Errorf("unsupported OpenCode database schema family layout: observed tables [%s]; missing or incomplete required structures [%s]. Supported schema families require complete v1 session/message/part or v2 session_v2/session_message layouts; use --db-path to point to a supported database", strings.Join(observed, ", "), strings.Join(issues, "; "))
}
