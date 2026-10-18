package opencode

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenCodeDB manages a read-only SQLite database connection to OpenCode.
type OpenCodeDB struct {
	db *sql.DB
}

// DefaultDBPath returns the standard location of the OpenCode SQLite database.
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "opencode", "opencode.db"), nil
}

// OpenDB opens a connection to the OpenCode SQLite database in read-only mode.
func OpenDB(dbPath string) (*OpenCodeDB, error) {
	if dbPath == "" {
		defPath, err := DefaultDBPath()
		if err != nil {
			return nil, err
		}
		dbPath = defPath
	}

	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("verifying database path %s: %w", dbPath, err)
	}

	dsn := fmt.Sprintf("file:%s?mode=ro", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database at %s: %w", dbPath, err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pinging sqlite database at %s: %w", dbPath, err)
	}

	return &OpenCodeDB{db: db}, nil
}

// DB returns the underlying *sql.DB handle.
func (c *OpenCodeDB) DB() *sql.DB {
	return c.db
}

// Close closes the database connection.
func (c *OpenCodeDB) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
