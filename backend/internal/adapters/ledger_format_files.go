package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LedgerFormatMode string

const (
	LedgerFormatWrite  LedgerFormatMode = "write"
	LedgerFormatDryRun LedgerFormatMode = "dry-run"
	LedgerFormatCheck  LedgerFormatMode = "check"
)

type LedgerFormatOptions struct {
	Workspaces string
	All        bool
	TicketIDs  []string
	Mode       LedgerFormatMode
}
type LedgerFormatResult struct {
	Path    string
	Changed bool
	Written bool
	Error   error
}
type formatTarget struct {
	path   string
	data   []byte
	mode   fs.FileMode
	lock   *capture.PublicationLock
	result int
}

func withinFormatRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func formatLedgerPaths(opts LedgerFormatOptions) ([]string, string, error) {
	if opts.All == (len(opts.TicketIDs) > 0) {
		return nil, "", errors.New("exactly one ledger selection is required")
	}
	root, err := ResolveWorkspacesDir(opts.Workspaces)
	if err != nil {
		return nil, "", err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, "", fmt.Errorf("workspace root: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, "", err
	}
	if !info.IsDir() {
		return nil, "", errors.New("workspace root must be a directory")
	}
	paths := []string{}
	if opts.All {
		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && entry.Name() == "agentic_metrics.json" {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, "", fmt.Errorf("discovering ledgers: %w", err)
		}
	} else {
		seen := map[string]bool{}
		for _, id := range opts.TicketIDs {
			if id == "" || id == "." || id == ".." || strings.TrimSpace(id) != id || strings.ContainsAny(id, "/\\\x00") || filepath.IsAbs(id) {
				return nil, "", fmt.Errorf("unsafe or empty ticket ID %q", id)
			}
			if !seen[id] {
				paths = append(paths, filepath.Join(root, id, "agentic_metrics.json"))
				seen[id] = true
			}
		}
	}
	sort.Strings(paths)
	return paths, root, nil
}

// FormatLedgerFiles preflights every selected target under publication ownership
// before replacing any ledger. Results retain partial writes on later I/O errors.
func FormatLedgerFiles(ctx context.Context, opts LedgerFormatOptions) (results []LedgerFormatResult, err error) {
	if opts.Mode != "" && opts.Mode != LedgerFormatWrite && opts.Mode != LedgerFormatDryRun && opts.Mode != LedgerFormatCheck {
		return nil, errors.New("invalid format mode")
	}
	paths, root, err := formatLedgerPaths(opts)
	if err != nil {
		return nil, err
	}
	targets := []formatTarget{}
	seen := map[string]bool{}
	infos := []os.FileInfo{}
	defer func() {
		for _, t := range targets {
			if e := t.lock.Close(); e != nil {
				err = errors.Join(err, fmt.Errorf("publication unlock %s: %w", t.path, e))
			}
		}
	}()
	// Resolve and sort physical paths before taking multiple locks, so overlapping
	// selections use one global order and cannot deadlock each other.
	physical := map[string]string{}
	for _, path := range paths {
		canonical, e := filepath.EvalSymlinks(path)
		if e == nil && !withinFormatRoot(root, canonical) {
			e = errors.New("ledger physical path escapes selected workspace root")
		}
		if e == nil && filepath.Base(canonical) != "agentic_metrics.json" {
			e = errors.New("ledger alias must resolve to agentic_metrics.json")
		}
		if e != nil {
			results = append(results, LedgerFormatResult{Path: path, Error: e})
			err = errors.Join(err, e)
			continue
		}
		if seen[canonical] {
			continue
		}
		seen[canonical] = true
		physical[canonical] = path
	}
	canonicalPaths := make([]string, 0, len(physical))
	for path := range physical {
		canonicalPaths = append(canonicalPaths, path)
	}
	sort.Strings(canonicalPaths)
	for _, path := range canonicalPaths {
		r := LedgerFormatResult{Path: physical[path]}
		index := len(results)
		results = append(results, r)
		lock, e := capture.AcquirePublication(ctx, path)
		if e != nil {
			results[index].Error = e
			err = errors.Join(err, e)
			continue
		}
		t := formatTarget{path: path, lock: lock, result: index}
		targets = append(targets, t)
		target := &targets[len(targets)-1]
		// Revalidate aliases after ownership, before reading any bytes.
		real, e := filepath.EvalSymlinks(physical[path])
		if e == nil && real != path {
			e = errors.New("ledger alias changed while acquiring ownership")
		}
		var info os.FileInfo
		if e == nil {
			info, e = os.Stat(path)
		}
		if e == nil && (!info.Mode().IsRegular() || info.Size() > publicationLimit) {
			e = errors.New("ledger must be a bounded regular file")
		}
		if e == nil {
			for _, prior := range infos {
				if os.SameFile(prior, info) {
					e = errors.New("hard-linked ledger aliases cannot be safely replaced; use symbolic aliases")
				}
			}
		}
		if e == nil {
			infos = append(infos, info)
		}
		var data []byte
		if e == nil {
			data, e = readOwnedLedger(path)
		}
		if e == nil {
			target.data, e = FormatLedgerJSON(data)
		}
		if e != nil {
			results[index].Error = e
			err = errors.Join(err, fmt.Errorf("%s: %w", r.Path, e))
			continue
		}
		target.mode = info.Mode()
		results[index].Changed = !bytes.Equal(data, target.data)
	}
	if err != nil {
		return results, fmt.Errorf("format preflight failed; no ledgers replaced: %w", err)
	}
	if opts.Mode == LedgerFormatCheck || opts.Mode == LedgerFormatDryRun {
		return results, nil
	}
	for _, target := range targets {
		if !results[target.result].Changed {
			continue
		}
		if e := replaceFormattedLedger(target.path, target.data, target.mode); e != nil {
			results[target.result].Error = e
			err = errors.Join(err, fmt.Errorf("%s: %w", target.path, e))
			continue
		}
		results[target.result].Written = true
	}
	return results, err
}
func replaceFormattedLedger(path string, data []byte, mode fs.FileMode) error {
	// One-file atomic replacement needs no multi-artifact journal. Reject pending
	// sets above; an interrupted stage leaves the original ledger accepted.
	f, err := os.CreateTemp(filepath.Dir(path), ".ledger-format-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "format staging cleanup:", e)
		}
	}()
	if err = f.Chmod(mode); err == nil {
		_, err = f.Write(data)
	}
	if e := errors.Join(err, f.Sync(), f.Close()); e != nil {
		return e
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
