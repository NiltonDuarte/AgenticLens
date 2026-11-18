package config

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters/filesystem"
)

// An admission scopes execution authority to a checkout. It contains no
// capture registration, ticket identity, repository configuration or metrics.
type codexWorktreeAdmission struct {
	Protocol int    `json:"protocol"`
	Worktree string `json:"worktree"`
	Home     string `json:"codex_home"`
}

func operatorCacheRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home, err = filesystem.PhysicalPath(home)
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agenticlens"), nil
}

func admissionPath(root, project string) string {
	id := sha256.Sum256([]byte(project))
	return filepath.Join(root, "codex-worktree-admissions", fmt.Sprintf("%x.json", id))
}

func admissionForPath(root, project string) (string, error) {
	path := admissionPath(root, project)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 16384 {
		return "", fmt.Errorf("invalid protected Codex worktree admission %s", path)
	}
	data, err := readBootstrap(path)
	if err != nil {
		return "", err
	}
	var r codexWorktreeAdmission
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}
	if r.Protocol != 1 || r.Worktree != project || !filepath.IsAbs(r.Home) {
		return "", fmt.Errorf("invalid protected Codex worktree admission %s", path)
	}
	if err := verifyProtectedNames(project, r.Home, path); err != nil {
		return "", err
	}
	// A completed user uninstall writes a private inactive receipt. A
	// missing/malformed receipt is not evidence that an operator released
	// this boundary; stale worktrees stay protected until explicit repair.
	receipt, err := readCodexCaptureReceipt(r.Home)
	if err != nil {
		return "", fmt.Errorf("admission for %s requires the original Codex installation receipt in %s; restore or explicitly repair it: %w", project, r.Home, err)
	}
	if !receipt.Inbox {
		if receipt.Groups != nil && len(receipt.Groups) == 0 {
			return "", nil
		}
		return "", errors.New("protected worktree receipt no longer establishes its installation; explicitly repair Codex setup")
	}
	return r.Home, nil
}

// Respect actual Git boundaries, including linked worktrees located under
// another checkout. A nested YAML file alone never creates a worktree.
func admissionBoundary(project string) (string, error) {
	for dir := project; ; dir = filepath.Dir(dir) {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			out, err := filesystem.RepositoryGitCommand(ctx, dir, "rev-parse", "--show-toplevel").Output()
			if err != nil {
				return "", fmt.Errorf("cannot verify Git worktree at %s: %w", dir, err)
			}
			root, err := filesystem.PhysicalPath(strings.TrimSuffix(string(out), "\n"))
			if err != nil {
				return "", err
			}
			if root != dir {
				return "", fmt.Errorf("unexpected Git worktree boundary at %s", dir)
			}
			return root, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if filepath.Dir(dir) == dir {
			return "", nil
		}
	}
}

// ProtectedCodexProjectHome follows existing private admissions up the physical
// path. A nested worktree is still writable from an admitted parent; neither
// Git metadata nor a nested project config can detach that subtree.
func ProtectedCodexProjectHome(project string) (string, error) {
	root, err := operatorCacheRoot()
	if err != nil {
		return "", err
	}
	project, err = filesystem.PhysicalPath(project)
	if err != nil {
		return "", err
	}
	for dir := project; ; dir = filepath.Dir(dir) {
		home, err := admissionForPath(root, dir)
		if err != nil || home != "" {
			return home, err
		}
		if filepath.Dir(dir) == dir {
			return "", nil
		}
	}
}

// AdmitProtectedCodexProject runs in the verified engine while the manager
// has verified the engine and the engine holds the namespace worker lease.
func AdmitProtectedCodexProject(project, home string) (returnErr error) {
	return withCaptureRouteLease(true, func() error { return recordCodexWorktree(project, home) })
}

func recordCodexWorktree(project, home string) (returnErr error) {
	project, err := filesystem.PhysicalPath(project)
	if err != nil {
		return err
	}
	home, err = filesystem.PhysicalPath(home)
	if err != nil {
		return err
	}
	root, err := operatorCacheRoot()
	if err != nil {
		return err
	}
	worktree, err := admissionBoundary(project)
	if err != nil {
		return err
	}
	if worktree == "" {
		worktree = project
	}
	path := admissionPath(root, worktree)
	if err := verifyProtectedNames(project, home, path); err != nil {
		return err
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(parent)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return errors.New("protected Codex admission directory must be private")
	}
	active, err := ProtectedCodexProjectHome(project)
	if err != nil {
		return err
	}
	if active != "" {
		if active != home {
			return errors.New("worktree already admitted by another Codex home")
		}
		return nil
	}
	expected, err := json.Marshal(codexWorktreeAdmission{Protocol: 1, Worktree: worktree, Home: home})
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(parent, ".admit-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	if _, err := tmp.Write(expected); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Sync(); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// The exclusive routing lease excludes competing first admissions.
	// Only a verified released record (or an absent path) reaches here.
	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}
	directory, err := os.Open(parent)
	if err != nil {
		return err
	}
	return errors.Join(directory.Sync(), directory.Close())
}

// Never create a repository enrollment or cache just to serve an independent
// OpenCode project. A protected engine necessarily has an existing cache.
// The directory lease spans ordinary engine selection and execution, including
// nested worker starts, and excludes first admission while either is in flight.
func withCaptureRouteLease(exclusive bool, run func() error) (returnErr error) {
	root, err := operatorCacheRoot()
	if err != nil {
		return err
	}
	dir, err := os.OpenFile(root, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_DIRECTORY|syscall.O_NONBLOCK, 0)
	if !exclusive && errors.Is(err, os.ErrNotExist) {
		return run()
	}
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, dir.Close()) }()
	flags := syscall.LOCK_SH | syscall.LOCK_NB
	if exclusive {
		flags = syscall.LOCK_EX | syscall.LOCK_NB
	}
	for {
		err = syscall.Flock(int(dir.Fd()), flags)
		if !errors.Is(err, syscall.EINTR) {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("capture routing in use; retry after the current admission or callback finishes: %w", err)
	}
	return run()
}

// ErrProtectedCodexProject asks an ordinary caller to reselect the protected engine.
var ErrProtectedCodexProject = errors.New("capture worktree requires protected Codex execution")

// WithOrdinaryCapture holds the read-side route guard across selection and execution.
func WithOrdinaryCapture(project string, run func() error) error {
	return withCaptureRouteLease(false, func() error {
		home, err := ProtectedCodexProjectHome(project)
		if err != nil {
			return err
		}
		if home != "" {
			return ErrProtectedCodexProject
		}
		return run()
	})
}

// PrepareProtectedCodexProject serializes the first private admission with
// ordinary callbacks. Once recorded, subsequent callbacks require no lease.
func PrepareProtectedCodexProject(project, home string, admit func() error) error {
	home, err := filesystem.PhysicalPath(home)
	if err != nil {
		return err
	}
	check := func() (bool, error) {
		current, err := ProtectedCodexProjectHome(project)
		if err != nil {
			return false, err
		}
		if current != "" && current != home {
			return false, errors.New("protected worktree belongs to a different Codex home; resolve the operator admission")
		}
		return current != "", nil
	}
	if ok, err := check(); err != nil || ok {
		return err
	}
	admitErr := admit()
	ok, err := check()
	if err != nil {
		return errors.Join(admitErr, err)
	}
	if ok {
		return nil
	}
	if admitErr != nil {
		return admitErr
	}
	return errors.New("verified engine did not record protected worktree admission; install a compatible engine")
}

// ReleaseProtectedCodexHome forgets only this user's completed, uninstalled
// integration. Repeating uninstall retries interrupted private cleanup. Workers
// already admitted remain bound to their original authority and must reverify it.
func ReleaseProtectedCodexHome(home string) error {
	home, err := filesystem.PhysicalPath(home)
	if err != nil {
		return err
	}
	root, err := operatorCacheRoot()
	if err != nil {
		return err
	}
	dir := filepath.Join(root, "codex-worktree-admissions")
	if _, err := os.Lstat(dir); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	return withCaptureRouteLease(true, func() error {
		receipt, err := readCodexCaptureReceipt(home)
		if err != nil {
			return err
		}
		if receipt.Inbox || receipt.Groups == nil || len(receipt.Groups) != 0 {
			return errors.New("Codex user uninstall is not committed; protected admissions preserved")
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.Type().IsRegular() || len(e.Name()) != 69 || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			data, err := readBootstrap(path)
			if err != nil {
				return err
			}
			var r codexWorktreeAdmission
			if err := json.Unmarshal(data, &r); err != nil {
				return err
			}
			if r.Home != home {
				continue
			}
			if r.Protocol != 1 || !filepath.IsAbs(r.Worktree) || path != admissionPath(root, r.Worktree) {
				return fmt.Errorf("invalid protected admission %s; preserved", path)
			}
			if err := verifyProtectedNames(r.Worktree, home, path); err != nil {
				return err
			}
			if err := os.Remove(path); err != nil {
				return err
			}
		}
		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		return errors.Join(f.Sync(), f.Close())
	})
}
