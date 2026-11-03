package capture

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const coordinationSchema = `CREATE TABLE IF NOT EXISTS worktree_coordinates(worktree_id INTEGER PRIMARY KEY,token TEXT NOT NULL UNIQUE);`

// Existing unreleased layouts without coordinate mappings are unsupported.
// Never manufacture paths for them or change the database during admission.
func validateCoordinateSchema(ctx context.Context, db *sql.DB) error {
	var present int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='worktree_coordinates'").Scan(&present); err != nil {
		return err
	}
	if present != 1 {
		return errors.New("unsupported capture coordinate layout: worktree mapping table is missing; preserve this state and reset or reinstall capture into a new state directory")
	}
	var missing int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM (SELECT id FROM worktrees w WHERE directory<>'' AND NOT EXISTS(SELECT 1 FROM worktree_coordinates c WHERE c.worktree_id=w.id) LIMIT 1)`).Scan(&missing)
	if err != nil {
		return fmt.Errorf("unsupported capture coordinate layout; preserve this state and reset or reinstall capture into a new state directory: %w", err)
	}
	if missing != 0 {
		return errors.New("unsupported capture coordinate layout: registered worktree mapping is missing; preserve this state and reset or reinstall capture into a new state directory")
	}
	return nil
}

func validCoordinateToken(token string) bool {
	if len(token) == 0 || len(token) > 120 || token == "." || token == ".." {
		return false
	}
	for _, c := range token {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// coordinatePath rejects symlinked coordination directories even for registered
// scopes; a persisted token is authority for a name, not for a redirected target.
func coordinatePath(registry, token string) (string, error) {
	if !validCoordinateToken(token) {
		return "", errors.New("invalid capture worktree coordination token")
	}
	root := filepath.Join(registry, "worktrees")
	for _, path := range []string{root, filepath.Join(root, token)} {
		if err := validateCoordinateDirectory(path); err != nil {
			return "", err
		}
	}
	return filepath.Join(root, token), nil
}

func validateCoordinateDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("capture coordination path is not a physical directory: %s", path)
	}
	return nil
}

func coordinateTokenTx(ctx context.Context, tx *sql.Tx, id int64) (string, error) {
	var version int
	if err := tx.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
		return "", err
	}
	if version != RegistryProtocolVersion {
		return "", fmt.Errorf("unsupported capture registry protocol %d", version)
	}
	var token string
	err := tx.QueryRowContext(ctx, "SELECT token FROM worktree_coordinates WHERE worktree_id=?", id).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("unsupported capture coordinate layout for worktree %d; preserve this state and reset or reinstall capture into a new state directory: %w", id, err)
	}
	return token, nil
}

func readCoordinatePath(ctx context.Context, db *sql.DB, registry string, id int64) (path string, returnErr error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	token, err := coordinateTokenTx(ctx, tx, id)
	if err != nil {
		return "", err
	}
	return coordinatePath(registry, token)
}

func preferredCoordinateToken(ownership string) (string, error) {
	var owner worktreeOwnership
	if err := json.Unmarshal([]byte(ownership), &owner); err != nil {
		return "", err
	}
	if owner.GitDirectory == "" || owner.GitDirectory == owner.Repository.CommonDirectory || filepath.Base(filepath.Dir(owner.GitDirectory)) != "worktrees" {
		return "", nil
	}
	base := filepath.Base(owner.GitDirectory)
	var safe strings.Builder
	for _, c := range base {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' {
			safe.WriteRune(c)
		} else {
			safe.WriteByte('-')
		}
		if safe.Len() >= 48 {
			break
		}
	}
	digest := sha256.Sum256([]byte(owner.GitDirectory + "\x00" + owner.GitIdentity))
	return "wt-" + safe.String() + "-" + hex.EncodeToString(digest[:8]), nil
}

// allocateCoordinate reserves a directory exclusively before publishing its
// mapping. A crash can leave an unclaimed directory, which future opens skip.
func allocateCoordinate(ctx context.Context, tx *sql.Tx, registry string, id int64, ownership string, randomToken func() (string, error)) (string, error) {
	preferred, err := preferredCoordinateToken(ownership)
	if err != nil {
		return "", err
	}
	root := filepath.Join(registry, "worktrees")
	if err = validateCoordinateDirectory(root); err != nil {
		return "", err
	}
	if err = os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	for attempt := 0; attempt < 16; attempt++ {
		token := preferred
		if token == "" || attempt > 0 {
			token, err = randomToken()
			if err != nil {
				return "", fmt.Errorf("generate capture coordination token: %w", err)
			}
		}
		path, err := coordinatePath(registry, token)
		if err != nil { // Reserved non-directory entries are collisions too.
			if validCoordinateToken(token) {
				if _, statErr := os.Lstat(filepath.Join(root, token)); statErr == nil {
					continue
				}
			}
			return "", err
		}
		var used int
		if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM worktree_coordinates WHERE token=?", token).Scan(&used); err != nil {
			return "", err
		}
		if used > 0 {
			continue
		}
		if err = os.Mkdir(path, 0700); errors.Is(err, os.ErrExist) {
			continue
		} else if err != nil {
			return "", err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO worktree_coordinates(worktree_id,token)VALUES(?,?)", id, token); err != nil {
			return "", errors.Join(err, os.Remove(path))
		}
		return path, nil
	}
	return "", errors.New("capture coordination token allocation exhausted after 16 collisions")
}

func randomCoordinateToken() (string, error) {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
