package capture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// ProjectDirectory resolves physical paths through existing ancestors and dangling
// symlinks before a destination or its parent exists. Source ownership still
// requires independent evidence from an existing source directory.
func ProjectDirectory(path string) (string, error) {
	return filesystem.PhysicalPath(path)
}

// ProjectState isolates external and project-local state by exact physical directory.
// Legacy unscoped registries remain untouched for explicit recovery.
func ProjectState(base, project string) (string, error) {
	project, err := ProjectDirectory(project)
	if err != nil {
		return "", err
	}
	base, err = ProjectDirectory(base)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(project))
	key := hex.EncodeToString(sum[:])
	if filepath.Base(filepath.Dir(base)) == "projects" {
		if filepath.Base(base) != key {
			return "", errors.New("state namespace belongs to another project")
		}
		return base, nil
	}
	return filepath.Join(base, "projects", key), nil
}
func OpenProject(directory, project string) (*Store, error) {
	project, err := ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	s, err := Open(directory)
	if err != nil {
		return nil, err
	}
	if _, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS project_owner (singleton INTEGER PRIMARY KEY CHECK(singleton=1), directory TEXT NOT NULL); INSERT OR IGNORE INTO project_owner VALUES(1,?)`, project); err != nil {
		return nil, errors.Join(err, s.Close())
	}
	var owner string
	if err = s.db.QueryRow(`SELECT directory FROM project_owner WHERE singleton=1`).Scan(&owner); err != nil {
		return nil, errors.Join(err, s.Close())
	}
	if owner != project {
		return nil, errors.Join(errors.New("copied capture state belongs to another project; explicit recovery required"), s.Close())
	}
	s.project = project
	return s, nil
}

// ValidateProject accepts contained source directories unless a nearer project owns them.
// Caller flags and registry location
// cannot grant a session membership in a different project.
func ValidateProject(project, sourceDirectory string) error {
	if sourceDirectory == "" || !filepath.IsAbs(sourceDirectory) {
		return errors.New("source project directory unavailable; ownership unresolved")
	}
	expected, err := ProjectDirectory(project)
	if err != nil {
		return err
	}
	actual, err := filepath.EvalSymlinks(sourceDirectory)
	if err != nil {
		return fmt.Errorf("source project unavailable: %w", err)
	}
	rel, err := filepath.Rel(expected, actual)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return errors.New("source project directory conflicts with configured project")
	}
	for dir := actual; dir != expected; dir = filepath.Dir(dir) {
		for _, marker := range []string{".agenticlens-config.yaml", ".agenticlens-config.yml", ".agenticlens-config.json", ".git"} {
			if _, e := os.Lstat(filepath.Join(dir, marker)); e == nil {
				return fmt.Errorf("source belongs to nearer project %s; explicit project resolution required", dir)
			} else if !errors.Is(e, os.ErrNotExist) {
				return e
			}
		}
	}
	return nil
}
func ValidateOwner(ctx context.Context, project string, r domain.CaptureRegistration) error {
	if r.Identity.Source == domain.SourceCodex {
		path := r.TranscriptPath
		if path == "" {
			var err error
			path, err = codex.DiscoverExactRolloutPath(r.Identity.Store, r.Identity.SessionID)
			if err != nil {
				return err
			}
		}
		if err := codex.ValidateRolloutStore(r.Identity.Store, path); err != nil {
			return err
		}
		b := &workBudget{sessions: 1, bytes: MaxPayloadBytes}
		t, err := readMetadata(ctx, path, b)
		if err != nil {
			return err
		}
		if t.ThreadID != r.Identity.SessionID {
			return errors.New("source identity mismatch")
		}
		return ValidateProject(project, t.WorkingDir)
	}
	db, err := opencode.OpenDB(r.Identity.Store)
	if err != nil {
		return err
	}
	defer db.Close()
	directory, err := opencode.CaptureProjectDirectory(ctx, db.DB(), r.Identity.SessionID)
	if err != nil {
		return err
	}
	return ValidateProject(project, directory)
}
func (s *Store) validateOwner(ctx context.Context, r domain.CaptureRegistration) error {
	if s.project == "" {
		return nil
	}
	return ValidateOwner(ctx, s.project, r)
}

// ValidateDestination forbids publication outside the configured project, including
// nested symlink escapes. It does not create any destination directories.
func ValidateDestination(project, path string) error {
	project, err := ProjectDirectory(project)
	if err != nil {
		return err
	}
	target, err := ProjectDirectory(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(project, target)
	if err != nil {
		return err
	}
	if rel == ".." || (len(rel) > 3 && rel[:3] == ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return errors.New("publication destination escapes project directory")
	}
	return nil
}
