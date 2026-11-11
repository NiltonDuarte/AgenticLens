package opencode

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// OpenCodeDB manages read-only connections to one explicitly selected source.
type OpenCodeDB struct {
	db *sql.DB
}

// OpenDB applies the explicit lock-wait budget to every pooled connection.
// Resolve symlinks before lexical normalization so symlink/.. traversal retains
// its filesystem meaning. No missing or invalid path selects another store.
func OpenDB(dbPath string, busyTimeoutMs int) (*OpenCodeDB, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, errors.New("OpenCode database path is required: supply --db-path or telemetry.sources.opencode.database_path in the selected local configuration")
	}
	if busyTimeoutMs < 0 {
		return nil, errors.New("database.busy_timeout_ms must be a nonnegative integer")
	}
	resolved, err := filepath.EvalSymlinks(dbPath)
	if err != nil {
		return nil, fmt.Errorf("resolving selected OpenCode database path %q: %w", dbPath, err)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return nil, fmt.Errorf("resolving selected OpenCode database path %q: %w", dbPath, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("verifying selected OpenCode database path %q: %w", dbPath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("selected OpenCode database path %q must identify a regular file", dbPath)
	}
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(resolved)}
	query := url.Values{"mode": {"ro"}, "_pragma": {"query_only(1)", "busy_timeout(" + strconv.Itoa(busyTimeoutMs) + ")"}}
	uri.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, fmt.Errorf("opening selected OpenCode database %q: %w", dbPath, err)
	}
	if err := db.Ping(); err != nil {
		return nil, errors.Join(fmt.Errorf("opening selected OpenCode database %q: %w", dbPath, err), db.Close())
	}
	return &OpenCodeDB{db: db}, nil
}

func (c *OpenCodeDB) DB() *sql.DB {
	return c.db
}

func (c *OpenCodeDB) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
