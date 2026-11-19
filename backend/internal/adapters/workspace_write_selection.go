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
	return location.AcquirePublicationWithProtection(ctx, project, nil)
}

// AcquirePublicationWithProtection also validates the worker's admitted
// private state when generation runs under protected capture.
func (location WorkspaceLocation) AcquirePublicationWithProtection(ctx context.Context, project string, protection *capture.ProtectedPublication) (*capture.PublicationLock, error) {
	if location.reference == nil {
		return nil, fmt.Errorf("workspace publication requires a retained selection")
	}
	return capture.AcquireSelectedPublication(ctx, location.LedgerPath, project, location.directoryInfo, protection)
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

// PrepareWorkspacePublication records an empty destination using the same
// ticket and directory leases that will protect its first publication.
func PrepareWorkspacePublication(ctx context.Context, cfg domain.ProjectConfig, project, ticket, path string, protection *capture.ProtectedPublication) (returnErr error) {
	key := PublicationTicketKey(ticket, cfg.Ticket)
	lease, err := capture.AcquireInitialPublication(ctx, project, key, protection)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lease.Close()) }()

	return prepareWorkspacePublicationWithLease(ctx, cfg, project, ticket, path, protection, lease, false, nil)
}

func prepareWorkspacePublicationWithLease(ctx context.Context, cfg domain.ProjectConfig, project, ticket, path string, protection *capture.ProtectedPublication, lease *capture.InitialPublicationLease, native bool, replacing *capture.InitialPublication) (returnErr error) {
	location, err := initialWorkspaceLocation(cfg, path)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, location.Close()) }()
	owner, err := location.AcquirePublicationWithProtection(ctx, project, protection)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, owner.Close()) }()

	if err := validateEmptyFirstPublication(cfg, ticket, location.LedgerPath); err != nil {
		return err
	}
	if native {
		return lease.PrepareNative(owner, replacing)
	}
	return lease.Prepare(owner)
}

func initialWorkspaceLocation(cfg domain.ProjectConfig, path string) (WorkspaceLocation, error) {
	if filepath.Base(path) != "agentic_metrics.json" {
		return WorkspaceLocation{}, errors.New("first-publication destination must name agentic_metrics.json")
	}
	dir, err := filesystem.PhysicalPath(filepath.Dir(path))
	if err != nil {
		return WorkspaceLocation{}, err
	}
	policy, err := workspaceDirectoryPolicy(cfg, dir)
	if err != nil {
		return WorkspaceLocation{}, err
	}
	location := WorkspaceLocation{Path: dir, LedgerPath: filepath.Join(dir, "agentic_metrics.json"), Artifacts: policy}
	_, err = location.retainPublicationDirectory()
	return location, err
}

func validateEmptyFirstPublication(cfg domain.ProjectConfig, ticket, ledger string) error {
	if err := CheckPublication(ledger); err != nil {
		return err
	}
	if _, err := capture.LstatPublicationFile(ledger); err == nil {
		return errors.New("first-publication destination already contains a ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// A complete ticket lookup is required even when a directory is explicit.
	// Validation of ordinary explicit generation otherwise distinguishes bad
	// individual ledgers from unavailable workspace discovery.
	if _, err := ResolveTicketLocation(cfg, ticket, ""); !workspaceNotFound(err) {
		if err == nil {
			return errors.New("the ticket already has a canonical ledger")
		}
		return err
	}
	return ValidateInitialLedgerDestination(cfg, ticket, ledger)
}

func workspaceNotFound(err error) bool {
	var locationErr *WorkspaceResolutionError
	return errors.As(err, &locationErr) && locationErr.Outcome == "not_found"
}

// SelectGenerationWorkspacePublication resolves current ledger identity first
// for lifecycle writers. A private selection can establish the first output
// only if discovery proves absence or identifies its exact pending recovery.
func SelectGenerationWorkspacePublication(cfg domain.ProjectConfig, ticket, path string, lease *capture.InitialPublicationLease, automatic bool) (WorkspaceLocation, *capture.InitialPublication, error) {
	location, resolveErr := SelectWorkspacePublication(cfg, ticket, path)
	if resolveErr == nil || path != "" || !automatic || lease == nil {
		return location, nil, resolveErr
	}
	selection, err := lease.Read()
	if err != nil {
		return WorkspaceLocation{}, nil, errors.Join(err, location.Close())
	}
	if selection != nil && errors.Is(CheckPublication(selection.Path), ErrPublicationPending) {
		// A private journal can recover its own exact destination even
		// when another interrupted workspace prevents complete discovery.
		if err := location.Close(); err != nil {
			return WorkspaceLocation{}, nil, err
		}
		location, err := initialWorkspaceLocation(cfg, selection.Path)
		return location, selection, err
	}
	if !workspaceNotFound(resolveErr) {
		return location, nil, resolveErr
	}
	if selection == nil {
		return WorkspaceLocation{}, nil, errors.Join(resolveErr, errors.New("run capture prepare-publication to explicitly select the first ledger destination"))
	}
	if selection.ExistingLedgerObserved {
		return WorkspaceLocation{}, nil, errors.New("first-publication selection was retired after observing an existing ledger")
	}
	if selection.CommittedTransaction != "" {
		return WorkspaceLocation{}, nil, errors.New("first-publication selection was consumed; resolve or recover the existing ledger")
	}
	location, err = initialWorkspaceLocation(cfg, selection.Path)
	return location, selection, err
}

// ErrAwaitingPublicationDestination describes the ordinary waiting state for a
// bound ticket that has no current canonical ledger or destination indication.
var ErrAwaitingPublicationDestination = errors.New("awaiting an exact ledger destination from the bound native session")

// SelectNativeWorkspacePublication prefers canonical identity and existing
// private recovery. Only a source-verified, explicitly bound native owner can
// supply a new destination indication, and normal first-publication admission
// gives the resulting selection its authority.
func SelectNativeWorkspacePublication(ctx context.Context, cfg domain.ProjectConfig, project, ticket, path string, lease *capture.InitialPublicationLease, protection *capture.ProtectedPublication, registrations func() ([]domain.CaptureRegistration, error)) (WorkspaceLocation, *capture.InitialPublication, error) {
	location, selection, err := SelectGenerationWorkspacePublication(cfg, ticket, path, lease, true)
	if path != "" || lease == nil {
		return location, selection, err
	}
	if selection != nil && selection.Native && !selection.Closed() {
		if indicatorErr := CheckPublication(selection.Path); indicatorErr != nil {
			return location, selection, err
		}
		if err != nil && !selection.NativeMarkerUnattempted() {
			return location, selection, fmt.Errorf("old native selection is unavailable for recovery: %w", err)
		}
	} else if err == nil || !workspaceNotFound(err) {
		// A canonical ledger, operator selection or marked recovery keeps
		// its existing precedence over native indications.
		return location, selection, err
	}
	rows, err := registrations()
	if err != nil {
		return WorkspaceLocation{}, nil, errors.Join(err, location.Close())
	}
	destination, err := nativeDestinationDirectory(rows, project, ticket)
	if err != nil {
		return WorkspaceLocation{}, nil, errors.Join(err, location.Close())
	}
	ledger := filepath.Join(destination, "agentic_metrics.json")
	if selection != nil {
		if selection.Path == ledger && !selection.NativeMarkerUnattempted() {
			return location, selection, nil
		}
		var err error
		if !selection.NativeMarkerUnattempted() {
			err = retireTentativeNativeLocation(ctx, project, location, protection, selection)
		}
		if err = errors.Join(err, location.Close()); err != nil {
			return WorkspaceLocation{}, nil, err
		}
	}
	if err := prepareWorkspacePublicationWithLease(ctx, cfg, project, ticket, ledger, protection, lease, true, selection); err != nil {
		// A canonical ledger may have appeared through another writer after
		// the original discovery. Never fall back to an open native receipt.
		location, observed, resolveErr := SelectGenerationWorkspacePublication(cfg, ticket, "", lease, true)
		if resolveErr == nil && (observed == nil || errors.Is(CheckPublication(observed.Path), ErrPublicationPending)) {
			return location, observed, nil
		}
		return WorkspaceLocation{}, nil, errors.Join(fmt.Errorf("native destination rejected: %w", err), location.Close())
	}
	return SelectGenerationWorkspacePublication(cfg, ticket, "", lease, true)
}

func retireTentativeNativeLocation(ctx context.Context, project string, location WorkspaceLocation, protection *capture.ProtectedPublication, selection *capture.InitialPublication) (returnErr error) {
	owner, err := location.AcquirePublicationWithProtection(ctx, project, protection)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, owner.Close()) }()
	if owner.DirectoryIdentity() != selection.Directory {
		return errors.New("old native selection no longer matches its directory")
	}
	if err := owner.ValidateLocation(); err != nil {
		return err
	}
	// The public marker is durable before normal replacement can begin.
	// A private staging file without it has no authority over the artifacts.
	if err := CheckPublication(selection.Path); err != nil {
		return err
	}
	if _, err := capture.LstatPublicationFile(selection.Path); err == nil {
		return errors.New("old native selection is occupied by a ledger")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// ValidateCurrentNativePublication confirms the selected directory still has
// every current verified native indication before starting a new transaction.
func ValidateCurrentNativePublication(project, ticket string, selection *capture.InitialPublication, rows []domain.CaptureRegistration) error {
	if selection == nil || !selection.Native {
		return nil
	}
	destination, err := nativeDestinationDirectory(rows, project, ticket)
	if err != nil {
		return err
	}
	if filepath.Join(destination, "agentic_metrics.json") != selection.Path {
		return errors.New("native destination changed before publication; retry current indications")
	}
	return nil
}

func nativeDestinationDirectory(rows []domain.CaptureRegistration, project, ticket string) (string, error) {
	var destination string
	for _, row := range rows {
		if !hasNativePublicationDestination(row, project, ticket) {
			continue
		}
		hint := row.PublicationDestination
		if !filepath.IsAbs(hint) {
			hint = project + string(filepath.Separator) + hint
		}
		dir, err := filesystem.PhysicalPath(hint)
		if err != nil {
			return "", fmt.Errorf("invalid native destination: %w", err)
		}
		info, err := os.Stat(dir)
		if err != nil {
			return "", fmt.Errorf("native destination must be an existing directory: %w", err)
		}
		if !info.IsDir() {
			return "", errors.New("native destination is not a directory")
		}
		if destination != "" && destination != dir {
			return "", errors.New("bound sessions indicated conflicting ledger destinations")
		}
		destination = dir
	}
	if destination == "" {
		return "", ErrAwaitingPublicationDestination
	}
	return destination, nil
}

func HasNativePublicationDestination(rows []domain.CaptureRegistration, project, ticket string) bool {
	for _, row := range rows {
		if hasNativePublicationDestination(row, project, ticket) {
			return true
		}
	}
	return false
}

func hasNativePublicationDestination(row domain.CaptureRegistration, project, ticket string) bool {
	return row.TicketID == ticket && row.ProjectDirectory == project && row.ProjectProof == domain.CaptureProjectVerified &&
		row.Binding == "explicit_agent_response" && row.BindingAcceptedAt != "" && row.PublicationDestination != ""
}

// AdmitExplicitInitialPublication runs after owned baseline recovery. An
// interrupted first replacement may have appeared occupied during selection.
func AdmitExplicitInitialPublication(cfg domain.ProjectConfig, ticket, ledger string, lease *capture.InitialPublicationLease, owner *capture.PublicationLock, protected bool) (*capture.InitialPublication, error) {
	selection, err := lease.Read()
	if err != nil {
		return nil, err
	}
	if selection == nil || selection.Closed() {
		if protected {
			return nil, errors.New("protected first publication requires an operator-prepared destination")
		}
		return nil, ValidateInitialLedgerDestination(cfg, ticket, ledger)
	}
	if selection.Path != ledger {
		return nil, errors.New("a different first-publication destination is already selected for this ticket")
	}
	if owner.DirectoryIdentity() != selection.Directory {
		return nil, errors.New("first-publication directory no longer matches the prepared destination")
	}
	return selection, validateEmptyFirstPublication(cfg, ticket, ledger)
}

// ValidatePreparedWorkspacePublication checks identity and exact recovery
// authority once the destination is held, before any artifact or source read.
func ValidatePreparedWorkspacePublication(cfg domain.ProjectConfig, ticket, ledger string, paths []string, owner *capture.PublicationLock, selection *capture.InitialPublication) error {
	if filepath.Join(owner.OriginalDirectory(), "agentic_metrics.json") != selection.Path || owner.DirectoryIdentity() != selection.Directory {
		return errors.New("first-publication directory no longer matches the prepared destination")
	}
	pending, err := validateInitialPublicationRecovery(ticket, ledger, paths, selection)
	if err != nil {
		return err
	}
	if pending {
		return nil
	}
	if selection.Closed() {
		return errors.New("first-publication selection was already closed")
	}
	return validateEmptyFirstPublication(cfg, ticket, ledger)
}

// ValidateRecoveredInitialWorkspace requires a complete global lookup after
// exact recovery. A committed baseline already read with the exclusive owner
// is excluded from this second scan to avoid reacquiring its directory lock.
func ValidateRecoveredInitialWorkspace(cfg domain.ProjectConfig, ticket, ledger string, hasOwnedBaseline bool) error {
	owned := ""
	if hasOwnedBaseline {
		owned = ledger
	}
	_, _, err := resolveTicketLocation(cfg, ticket, "", false, false, owned)
	if workspaceNotFound(err) {
		return nil
	}
	if err == nil {
		return errors.New("a canonical ledger conflicts with the recovered first-publication destination")
	}
	return err
}
