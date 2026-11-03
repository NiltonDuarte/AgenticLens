package capture

import (
	adaptersconfig "agenticlens/internal/adapters/config"
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

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

// StateLocation separates shared persistence from exact-worktree coordination.
type StateLocation struct {
	RegistryDirectory     string
	CoordinationDirectory string
	WorktreeID            int64
	// Provisional means no scope has been allocated; CoordinationDirectory is empty.
	Provisional bool
}

func ProjectNameSegment(name string) (string, error) {
	if name == "" || !utf8.ValidString(name) || len(name) > 120 || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
		return "", errors.New("project_name must be a nonempty portable namespace name of at most 120 UTF-8 bytes")
	}
	literal := true
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
			literal = false
		}
	}
	stem := strings.ToUpper(strings.Split(name, ".")[0])
	if strings.HasSuffix(name, ".") || stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" || len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9' {
		literal = false
	}
	if literal {
		return name, nil
	}
	return "~" + hex.EncodeToString([]byte(name)), nil
}

// ProjectState requires the configured project label; physical worktree identity
// never substitutes for it. An explicit coordination path must already belong to
// this exact worktree before it can resolve the shared registry.
func ProjectState(base, project string) (string, error) {
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return "", err
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return "", err
	}
	return namedProjectState(base, cfg.ProjectName, project)
}
func namedProjectState(base, name, project string) (string, error) {
	return namedProjectStateContext(context.Background(), base, name, project)
}
func namedProjectStateContext(ctx context.Context, base, name, project string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	segment, err := ProjectNameSegment(name)
	if err != nil {
		return "", err
	}
	lexical, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	lexical = filepath.Clean(lexical)
	physical, err := ProjectDirectory(base)
	if err != nil {
		return "", err
	}
	// Preserve lexical namespace context until the full target has been checked.
	// Otherwise an escaping namespace symlink could be reinterpreted as a base.
	if isCoordinationPath(lexical) {
		return resolveCoordinationNamespace(ctx, lexical, physical, segment, name, project)
	}
	if filepath.Base(filepath.Dir(lexical)) == "projects" {
		if filepath.Base(lexical) != segment {
			return "", errors.New("state namespace differs from configured project_name")
		}
		return resolveContainedNamespace(project, filepath.Dir(filepath.Dir(lexical)), lexical, segment)
	}
	// An explicitly selected base/coordinate alias may itself resolve to an
	// already-completed namespace; its exact physical scope remains validated.
	if isCoordinationPath(physical) {
		return resolveCoordinationNamespace(ctx, lexical, physical, segment, name, project)
	}
	if filepath.Base(filepath.Dir(physical)) == "projects" {
		return resolveContainedNamespace(project, filepath.Dir(filepath.Dir(physical)), physical, segment)
	}
	if _, e := os.Stat(filepath.Join(physical, "registry.db")); e == nil {
		return "", errors.New("state base contains an unscoped registry; explicitly relocate legacy state before operational access")
	} else if !errors.Is(e, os.ErrNotExist) {
		return "", e
	}
	return resolveContainedNamespace(project, physical, filepath.Join(physical, "projects", segment), segment)
}

func resolveContainedNamespace(project, base, namespace, segment string) (string, error) {
	base, err := ProjectDirectory(base)
	if err != nil {
		return "", err
	}
	namespace, err = ProjectDirectory(namespace)
	if err != nil {
		return "", err
	}
	_, contained, err := relativeContainedPath(base, namespace)
	if err != nil {
		return "", err
	}
	if !contained {
		return "", errors.New("capture project namespace escapes the selected state base; configure the intended base explicitly")
	}
	if filepath.Base(filepath.Dir(namespace)) != "projects" || filepath.Base(namespace) != segment {
		return "", errors.New("capture namespace alias must preserve projects/<configured-project-name>; configure the intended base explicitly")
	}
	return validateStatePlacement(project, namespace)
}

func resolveCoordinationNamespace(ctx context.Context, lexical, physical, segment, name, project string) (string, error) {
	if !isCoordinationPath(physical) {
		return "", errors.New("coordination alias does not resolve to a recorded worktree directory")
	}
	registry := filepath.Dir(filepath.Dir(physical))
	if isCoordinationPath(lexical) {
		logicalRegistry := filepath.Dir(filepath.Dir(lexical))
		if filepath.Base(filepath.Dir(logicalRegistry)) == "projects" {
			if filepath.Base(logicalRegistry) != segment {
				return "", errors.New("state namespace differs from configured project_name")
			}
			expected, err := resolveContainedNamespace(project, filepath.Dir(filepath.Dir(logicalRegistry)), logicalRegistry, segment)
			if err != nil {
				return "", err
			}
			if expected != registry {
				return "", errors.New("coordination alias differs from its validated project namespace")
			}
		}
	}
	registry, err := resolveContainedNamespace(project, filepath.Dir(filepath.Dir(registry)), registry, segment)
	if err != nil {
		return "", err
	}
	store, err := openRegistryContext(ctx, registry, name, project, true)
	if err != nil {
		return "", err
	}
	coord := store.CoordinationDirectory()
	err = store.Close()
	if err != nil {
		return "", err
	}
	if coord != physical {
		return "", errors.New("coordination path belongs to another worktree")
	}
	return registry, nil
}

func OpenProject(directory, project string) (*Store, error) {
	return OpenProjectContext(context.Background(), directory, project)
}

// OpenProjectContext shares the caller's admission/handoff deadline through the
// same registry initializer. An expired caller cannot allocate a new scope.
func OpenProjectContext(ctx context.Context, directory, project string) (*Store, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return nil, err
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	directory, err = namedProjectStateContext(ctx, directory, cfg.ProjectName, project)
	if err != nil {
		return nil, err
	}
	return openRegistryContext(ctx, directory, cfg.ProjectName, project, false)
}
func ExistingLocation(directory, project string) (StateLocation, error) {
	s, err := OpenReadOnlyProject(directory, project)
	if err != nil {
		return StateLocation{}, err
	}
	location := StateLocation{RegistryDirectory: s.RegistryDirectory(), CoordinationDirectory: s.CoordinationDirectory(), WorktreeID: s.worktreeID}
	return location, s.Close()
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

// PreviewStateLocation performs no allocation. Unallocated scopes have no
// coordination path until OpenProject commits their unique directory name.
func PreviewStateLocation(directory, project string) (StateLocation, error) {
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return StateLocation{}, err
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return StateLocation{}, err
	}
	directory, err = namedProjectState(directory, cfg.ProjectName, project)
	if err != nil {
		return StateLocation{}, err
	}
	location := StateLocation{RegistryDirectory: directory, Provisional: true}
	s, err := openRegistry(directory, cfg.ProjectName, project, true)
	if err == nil {
		location.Provisional = false
		location.CoordinationDirectory = s.directory
		location.WorktreeID = s.worktreeID
		return location, s.Close()
	}
	if !errors.Is(err, os.ErrNotExist) {
		return location, err
	}

	return location, nil
}

func openReadOnlyDB(path string) (*sql.DB, error) {
	uri, err := sqliteFileURI(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", uri+"?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(250)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return db, nil
}
func logicalRegistryDirectory(state string) string {
	if isCoordinationPath(state) {
		return filepath.Dir(filepath.Dir(state))
	}
	return state
}

// RegistryPathDirectory separates a validated coordination path from its logical
// registry for filesystem presence checks. OpenProject still validates metadata.
func RegistryPathDirectory(state string) string { return logicalRegistryDirectory(state) }

func validateScopeActivation(db *sql.DB, id int64) error {
	return validateScopeActivationContext(context.Background(), db, id)
}
func validateScopeActivationContext(ctx context.Context, db *sql.DB, id int64) error {
	var exists int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='worktree_relocations'").Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	var state string
	err := db.QueryRowContext(ctx, "SELECT state FROM worktree_relocations WHERE worktree_id=?", id).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if state != "active" {
		return errors.New("capture worktree relocation is incomplete; resume explicit relocation before operational access")
	}
	return nil
}

func isCoordinationPath(state string) bool {
	registry := filepath.Dir(filepath.Dir(state))
	if filepath.Base(filepath.Dir(state)) != "worktrees" || filepath.Base(filepath.Dir(registry)) != "projects" {
		return false
	}
	return validDirectoryName(filepath.Base(state))
}

func validateStatePlacement(project, registry string) (string, error) {
	for _, file := range []string{".codex/hooks.json", ".codex/agenticlens-installation.json", ".opencode/plugins/agenticlens-capture.ts", ".opencode/agenticlens-installation.json", ".agenticlens-config.yaml", ".agenticlens-config.yml", ".agenticlens-config.json", ".agenticlens-config.local.yaml"} {
		target := filepath.Join(project, filepath.FromSlash(file))
		_, contained, err := relativeContainedPath(target, registry)
		if err != nil {
			return "", err
		}
		if contained {
			return "", errors.New("capture state cannot occupy an integration/configuration file or its descendants")
		}
	}
	return registry, nil
}
