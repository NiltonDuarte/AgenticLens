package capture

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type worktreeOwnership struct {
	DirectoryIdentity string
	GitDirectory      string
	GitIdentity       string
	Repository        RepositoryProof `json:",omitempty"`
}

// currentWorktreeOwnership records observed filesystem/Git ownership independently
// of the configured project label. It is not a replacement for source validation.
func currentWorktreeOwnership(project string) (string, error) {
	if project == "" {
		return "", nil
	}
	info, err := os.Stat(project)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("capture worktree is not a directory")
	}
	identity, err := fileIdentity(project)
	if err != nil {
		return "", err
	}
	proof := worktreeOwnership{DirectoryIdentity: identity}
	marker := filepath.Join(project, ".git")
	gitInfo, err := os.Stat(marker)
	if err == nil {
		gitDir := marker
		if !gitInfo.IsDir() {
			if gitInfo.Size() > 4096 {
				return "", errors.New("Git worktree marker exceeds ownership bound")
			}
			body, e := os.ReadFile(marker)
			if e != nil {
				return "", e
			}
			line := strings.TrimSpace(string(body))
			if !strings.HasPrefix(line, "gitdir:") {
				return "", errors.New("Git worktree marker unavailable")
			}
			gitDir = strings.TrimSpace(strings.TrimPrefix(line, "gitdir:"))
			if !filepath.IsAbs(gitDir) {
				gitDir = filepath.Join(project, gitDir)
			}
		}
		proof.GitDirectory, err = filepath.EvalSymlinks(gitDir)
		if err != nil {
			return "", err
		}
		proof.GitIdentity, err = fileIdentity(proof.GitDirectory)
		if err != nil {
			return "", err
		}
		proof.Repository, err = observedRepositoryProof(proof.GitDirectory)
		if err != nil {
			return "", err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	body, err := json.Marshal(proof)
	return string(body), err
}

func validateWorktreeOwnership(db *sql.DB, id int64, project string) error {
	return validateWorktreeOwnershipContext(context.Background(), db, id, project)
}
func validateWorktreeOwnershipContext(ctx context.Context, db *sql.DB, id int64, project string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	current, err := currentWorktreeOwnership(project)
	if err != nil {
		return fmt.Errorf("worktree ownership unavailable; retained capture state requires explicit recovery: %w", err)
	}
	var recorded string
	err = db.QueryRowContext(ctx, "SELECT ownership FROM worktree_ownership WHERE worktree_id=?", id).Scan(&recorded)
	if err != nil {
		return fmt.Errorf("worktree ownership association unavailable: %w", err)
	}
	same, e := sameWorktreeOwnership(recorded, current)
	if e != nil {
		return e
	}
	if !same {
		return errors.New("physical worktree or Git ownership changed; retained capture state requires explicit recovery")
	}
	return nil
}

// Adding Git to the same observed filesystem directory enriches its ownership;
// it does not transfer a scope to a replacement directory or another Git owner.
func sameWorktreeOwnership(recorded, current string) (bool, error) {
	if recorded == current {
		return true, nil
	}
	var before, after worktreeOwnership
	if err := json.Unmarshal([]byte(recorded), &before); err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(current), &after); err != nil {
		return false, err
	}
	if before.DirectoryIdentity == "" || before.DirectoryIdentity != after.DirectoryIdentity {
		return false, nil
	}
	if before.GitDirectory == "" && before.GitIdentity == "" && after.GitDirectory != "" && after.GitIdentity != "" {
		return true, nil
	}
	return before.GitDirectory == after.GitDirectory && before.GitIdentity == after.GitIdentity && (before.Repository == after.Repository || before.Repository == (RepositoryProof{})), nil
}

// The administrative commondir is observed without requiring a Git executable
// for ordinary capture. Historical authorization independently verifies it with
// Git plumbing before treating it as repository authority.
func observedRepositoryProof(gitDirectory string) (RepositoryProof, error) {
	common := gitDirectory
	marker := filepath.Join(gitDirectory, "commondir")
	info, err := os.Stat(marker)
	if err == nil {
		if !info.Mode().IsRegular() || info.Size() > 4096 {
			return RepositoryProof{}, errors.New("Git common-directory marker exceeds ownership bound")
		}
		body, err := os.ReadFile(marker)
		if err != nil {
			return RepositoryProof{}, err
		}
		common = strings.TrimSpace(string(body))
		if common == "" {
			return RepositoryProof{}, errors.New("empty Git common directory")
		}
		if !filepath.IsAbs(common) {
			common = filepath.Join(gitDirectory, common)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return RepositoryProof{}, err
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return RepositoryProof{}, err
	}
	identity, err := fileIdentity(common)
	if err != nil {
		return RepositoryProof{}, err
	}
	return RepositoryProof{CommonDirectory: common, DirectoryIdentity: identity}, nil
}
