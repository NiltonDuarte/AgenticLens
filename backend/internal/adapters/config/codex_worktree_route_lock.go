package config

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"agenticlens/internal/adapters/filesystem"
)

// A shared lock at every physical ancestor lets a worktree promotion exclude
// all ordinary callbacks beneath it. Only the promoted directory is locked
// exclusively, so sibling worktrees can admit or execute concurrently.
func withCaptureRouteLease(project string, exclusive bool, run func() error) (result error) {
	root, err := operatorCacheRoot()
	if err != nil {
		return err
	}
	locks := filepath.Join(root, "codex-worktree-route-locks")
	if _, err := os.Lstat(locks); !exclusive && errors.Is(err, os.ErrNotExist) {
		// Independent OpenCode installations need no coordination files. Older
		// Codex integrations can predate this lock directory; initialize it
		// for their next callback when a protected receipt already exists.
		receipt, err := readCodexCaptureReceipt(RuntimeCodexHome())
		if errors.Is(err, os.ErrNotExist) || (err == nil && !receipt.Inbox) {
			return run()
		}
		if err != nil {
			return err
		}
		if err := PrepareCodexRouteLocks(project, RuntimeCodexHome()); err != nil {
			return err
		}
	}
	if err := prepareCodexRouteLocks(locks); err != nil {
		return err
	}
	project, err = filesystem.PhysicalPath(project)
	if err != nil {
		return err
	}
	for dir := project; ; dir = filepath.Dir(dir) {
		key := sha256.Sum256([]byte(dir))
		path := filepath.Join(locks, fmt.Sprintf("%x.lock", key))
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0600)
		if err != nil {
			return err
		}
		defer func() { result = errors.Join(result, file.Close()) }()
		if err := lockCaptureRoute(file, exclusive && dir == project); err != nil {
			return err
		}
		if filepath.Dir(dir) == dir {
			break
		}
	}
	return run()
}

// Prepare the private coordination directory before exposing fresh Codex hooks.
func PrepareCodexRouteLocks(project, home string) error {
	root, err := operatorCacheRoot()
	if err != nil {
		return err
	}
	locks := filepath.Join(root, "codex-worktree-route-locks")
	if err := CheckProtectedCodexPaths(project, home, locks); err != nil {
		return err
	}
	return prepareCodexRouteLocks(locks)
}

func prepareCodexRouteLocks(locks string) error {
	physical, err := filesystem.PhysicalPath(locks)
	if err != nil {
		return err
	}
	if physical != locks {
		return errors.New("protected Codex route lock directory contains an alias")
	}
	if err := os.MkdirAll(locks, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(locks)
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || !ok || int(stat.Uid) != os.Geteuid() || info.Mode().Perm()&0077 != 0 {
		return errors.New("protected Codex route lock directory must be private and operator-owned")
	}
	return nil
}

func withRouteLock(file *os.File, exclusive bool, run func() error) (result error) {
	defer func() { result = errors.Join(result, file.Close()) }()
	if err := lockCaptureRoute(file, exclusive); err != nil {
		return err
	}
	return run()
}

func lockCaptureRoute(file *os.File, exclusive bool) error {
	flags := syscall.LOCK_SH | syscall.LOCK_NB
	if exclusive {
		flags = syscall.LOCK_EX | syscall.LOCK_NB
	}
	for {
		err := syscall.Flock(int(file.Fd()), flags)
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if err != nil {
			return fmt.Errorf("capture routing in use; retry after the current admission or callback finishes: %w", err)
		}
		return nil
	}
}
