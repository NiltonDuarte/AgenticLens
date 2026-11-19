package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
)

// SelectWorkspacePublication retains either an existing ticket's actual home
// or an explicitly named directory that is eligible for first publication.
// The caller must close the selection.
func SelectWorkspacePublication(cfg domain.ProjectConfig, ticket, path string) (WorkspaceLocation, error) {
	if path == "" {
		location, reference, err := resolveTicketReadLocation(cfg, ticket, "")
		if err == nil {
			location.reference = reference
		}
		return location, err
	}

	ledger, err := filesystem.PhysicalPath(path)
	if err != nil {
		return WorkspaceLocation{}, err
	}
	if filepath.Base(ledger) != "agentic_metrics.json" {
		return WorkspaceLocation{}, fmt.Errorf("ledger destination must name the canonical agentic_metrics.json ledger")
	}
	location := WorkspaceLocation{Path: filepath.Dir(ledger), LedgerPath: ledger}
	root, err := location.retainPublicationDirectory()
	if err != nil {
		return WorkspaceLocation{}, err
	}
	location.Artifacts, err = workspaceDirectoryPolicy(cfg, location.Path)
	if err != nil {
		return WorkspaceLocation{}, errors.Join(err, location.Close())
	}
	exists, err := selectedLedgerExists(root, ledger, true)
	if err == nil && !exists {
		err = ValidateInitialLedgerDestination(cfg, ticket, ledger)
		if err == nil {
			_, location.Diagnostics = DiscoverWorkspaceLocations(cfg)
		}
	}
	if err != nil {
		return WorkspaceLocation{}, errors.Join(err, location.Close())
	}
	return location, nil
}

// Close releases the observation. A successfully acquired publication keeps
// its own directory handles and remains valid after this method returns.
func (location *WorkspaceLocation) Close() error {
	if location.reference == nil {
		return nil
	}
	err := location.reference.Close()
	location.reference = nil
	return err
}

// AcquirePublication binds ownership to the selected physical directory
// before any recovery, baseline read or write can use that ownership.
func (location WorkspaceLocation) AcquirePublication(ctx context.Context, project string) (*capture.PublicationLock, error) {
	if location.reference == nil {
		return nil, fmt.Errorf("workspace publication requires a retained selection")
	}
	return capture.AcquireSelectedPublication(ctx, location.LedgerPath, project, location.directoryInfo)
}

// AcquirePublicationRead admits a maintenance preview against the same
// directory without using operational state or taking a write lease.
func (location WorkspaceLocation) AcquirePublicationRead() (*capture.ReadDirectory, error) {
	if location.reference == nil {
		return nil, fmt.Errorf("workspace publication preview requires a retained selection")
	}
	return openResolvedWorkspaceDirectory(location)
}

func (location *WorkspaceLocation) retainPublicationDirectory() (*os.Root, error) {
	root, err := os.OpenRoot(location.Path)
	if err != nil {
		return nil, fmt.Errorf("selected ledger directory unavailable: %w", err)
	}
	opened, statErr := root.Stat(".")
	current, pathErr := os.Lstat(location.Path)
	if statErr != nil || pathErr != nil || !os.SameFile(opened, current) {
		return nil, errors.Join(fmt.Errorf("workspace changed after selection; resolve and retry"), statErr, pathErr, root.Close())
	}
	location.directoryInfo = opened
	location.reference = root
	return root, nil
}

func selectedLedgerExists(root *os.Root, path string, allowMissing bool) (bool, error) {
	info, err := root.Lstat(filepath.Base(path))
	if allowMissing && errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("selected ledger unavailable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("ledger must be a regular file")
	}
	return true, nil
}

func readMaintenanceSelection(location WorkspaceLocation) (body []byte, info os.FileInfo, reference *os.File, err error) {
	root, err := location.retainPublicationDirectory()
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		err = errors.Join(err, location.Close())
		if err != nil && reference != nil {
			err = errors.Join(err, reference.Close())
			reference = nil
		}
	}()
	if _, err = selectedLedgerExists(root, location.LedgerPath, false); err != nil {
		return nil, nil, nil, err
	}
	file, err := root.Open(filepath.Base(location.LedgerPath))
	if err != nil {
		return nil, nil, nil, err
	}
	body, err = io.ReadAll(io.LimitReader(file, publicationLimit+1))
	if err = errors.Join(err, file.Close()); err != nil {
		return nil, nil, nil, err
	}
	if len(body) > publicationLimit {
		return nil, nil, nil, fmt.Errorf("publication member exceeds bounded read")
	}
	reference, err = root.Open(".")
	return body, location.directoryInfo, reference, err
}
