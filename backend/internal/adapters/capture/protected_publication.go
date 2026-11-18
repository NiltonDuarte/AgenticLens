package capture

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
)

// ProtectedPublication pins the worktree, operator home and physical state
// selected by a protected worker for its generation child.
type ProtectedPublication struct {
	Project, Home, State string
}

func AdmitProtectedPublication(project, home string) (ProtectedPublication, error) {
	var p ProtectedPublication
	var err error
	p.Project, err = filesystem.PhysicalPath(project)
	if err != nil {
		return p, err
	}
	p.Home, err = config.ProtectedCodexProjectHome(p.Project)
	if err != nil {
		return p, err
	}
	actual, err := filesystem.PhysicalPath(home)
	if err != nil || p.Home == "" || actual != p.Home {
		return p, errors.New("protected publication requires the admitted Codex home")
	}
	state, err := config.ResolvePublicationState(p.Project)
	if err != nil {
		return p, err
	}
	if err := config.CheckProtectedCodexPaths(p.Project, p.Home, state); err != nil {
		return p, err
	}
	p.State, err = filesystem.PhysicalPath(state)
	return p, err
}

func (p ProtectedPublication) Validate(project string) error {
	current, err := AdmitProtectedPublication(project, p.Home)
	if err != nil {
		return err
	}
	if current != p {
		return errors.New("protected publication project or state changed since worker admission; pending work retained")
	}
	return nil
}

// Acquire uses this verified physical state for both the lock and its journal.
func (p ProtectedPublication) Acquire(ctx context.Context, ledger, project string) (*PublicationLock, error) {
	return acquirePublication(ctx, ledger, project, &p)
}

func (p *ProtectedPublication) checkPrivatePath(path string) error {
	if p == nil {
		return nil
	}
	if err := config.CheckProtectedCodexPaths(p.Project, p.Home, path); err != nil {
		return err
	}
	physical, err := filesystem.PhysicalPath(path)
	if err != nil {
		return err
	}
	// All internally derived descendants of the admitted state must stay at
	// their exact location, including an existing lock or recovery journal.
	if physical != path || !filepath.IsAbs(path) {
		return fmt.Errorf("protected publication state path has been redirected: %s", path)
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok || !info.Mode().IsRegular() || stat.Nlink != 1 {
			return fmt.Errorf("protected publication state requires a single-link regular file: %s", path)
		}
	}
	return nil
}

// CheckPrivatePath checks a concrete journal path before creating or recovering it.
func (lock *PublicationLock) CheckPrivatePath(path string) error {
	return lock.protection.checkPrivatePath(path)
}
