package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

func selectNativeGenerationWorkspace(project, registry, ticket, path string, cfg domain.ProjectConfig, policy domain.CaptureConfig, lease *capture.InitialPublicationLease, protection *capture.ProtectedPublication) (adapters.WorkspaceLocation, *capture.InitialPublication, error, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	rows, err := nativeGenerationRegistrations(ctx, project, registry)
	if err != nil {
		return adapters.WorkspaceLocation{}, nil, nil, err
	}
	if conflict := firstGenerationMetadataConflict(rows, ticket); conflict != nil {
		// Canonical history and exact interrupted recovery can proceed.
		// Conflicting types cannot authorize an unmarked first write.
		location, selection, err := adapters.SelectGenerationWorkspacePublication(cfg, ticket, path, lease, true)
		if err != nil || (selection != nil && !errors.Is(adapters.CheckPublication(selection.Path), adapters.ErrPublicationPending)) {
			return adapters.WorkspaceLocation{}, nil, conflict, errors.Join(conflict, err, location.Close())
		}
		if selection == nil && path == "" {
			// The canonical ledger already has its own title. A title
			// disagreement does not prevent applying an agreed type.
			_, _, conflict = domain.CaptureTicketMetadata(rows, ticket)
		}
		return location, selection, conflict, nil
	}
	location, selection, err := adapters.SelectNativeWorkspacePublication(ctx, cfg, project, ticket, path, lease, protection, func() ([]domain.CaptureRegistration, error) {
		return rows, nil
	})
	return location, selection, nil, err
}

func firstGenerationMetadataConflict(rows []domain.CaptureRegistration, ticket string) error {
	title, _, err := domain.CaptureTicketMetadata(rows, ticket)
	if err != nil || title != "" {
		return err
	}
	for _, row := range rows {
		if row.TicketID == ticket && strings.TrimSpace(row.TicketTitle) != "" {
			return fmt.Errorf("ticket %s has conflicting explicit titles; supply an agreed title before first publication", ticket)
		}
	}
	return nil
}

func nativeGenerationRegistrations(ctx context.Context, project, registry string) ([]domain.CaptureRegistration, error) {
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return nil, err
	}
	rows, readErr := store.ListActive(ctx)
	return rows, errors.Join(readErr, store.Close())
}

func beginAutomaticFirstPublication(ctx context.Context, store *capture.Store, project, ticket, ledger string, paths []string, selection *capture.InitialPublication) (bool, error) {
	admissionFailed := false
	err := adapters.BeginInitialPublicationWithAdmission(ledger, paths, selection, func(publishIndicator func() error) error {
		started := false
		admitErr := store.WithInitialPublicationRegistrations(ctx, ticket, func(rows []domain.CaptureRegistration) error {
			if err := firstGenerationMetadataConflict(rows, ticket); err != nil {
				return err
			}
			if err := adapters.ValidateCurrentNativePublication(project, ticket, selection, rows); err != nil {
				return err
			}
			started = true
			return publishIndicator()
		})
		admissionFailed = admitErr != nil && !started
		return admitErr
	})
	return admissionFailed, err
}

func publishGenerationArtifacts(ledger, summary, markdown string, facts domain.AgenticFactLedger, history []adapters.SessionFactDTO, begin func([]string) error) error {
	paths := authorizedPublicationPaths(ledger, summary)
	if ledger != "" {
		if err := begin(paths); err != nil {
			return fmt.Errorf("cannot stage recoverable publication: %w", err)
		}
		if err := saveGenerationLedgerPreservingHistory(ledger, facts, history); err != nil {
			return recoverFailedGenerationPublication(ledger, paths, fmt.Errorf("writing JSON to %s: %w", ledger, err))
		}
	}
	if summary != "" {
		path, err := capture.ProjectDirectory(summary)
		if err != nil {
			return recoverFailedGenerationPublication(ledger, paths, fmt.Errorf("resolving summary file %s: %w", summary, err))
		}
		if err := adapters.UpdateSummaryFile(path, markdown); err != nil {
			return recoverFailedGenerationPublication(ledger, paths, fmt.Errorf("updating summary file %s: %w", summary, err))
		}
	}
	if ledger != "" {
		if err := adapters.CommitPublication(ledger, paths); err != nil {
			return recoverFailedGenerationPublication(ledger, paths, fmt.Errorf("committing artifact publication: %w", err))
		}
	}
	return nil
}

func publishAutomaticCaptureArtifacts(project, registry, ticket, ledger string, policy domain.CaptureConfig, result capture.ReconcileResult, selection *capture.InitialPublication, first bool, facts []domain.SessionFact, publish func(func([]string) error) error) (blocked bool, returnErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return false, err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	closeAdmission, err := capture.AcquireWorkerControlAdmission(store.CoordinationDirectory())
	if err != nil {
		return false, err
	}
	defer func() { returnErr = errors.Join(returnErr, closeAdmission()) }()
	// The caller already retains the ticket and destination leases. Hold the
	// worktree lifecycle revision through marker creation, replacement and commit.
	err = store.WithCaptureLifecycle(ctx, result.LifecycleRevision, func() error {
		begin := func(paths []string) error {
			if !first {
				return adapters.BeginInitialPublication(ledger, paths, selection)
			}
			var err error
			blocked, err = beginAutomaticFirstPublication(ctx, store, project, ticket, ledger, paths, selection)
			return err
		}
		if err := publish(begin); err != nil {
			return err
		}
		if len(result.RestoresReady) == 0 {
			return nil
		}
		materialized := make(map[string]bool, len(facts))
		for _, fact := range facts {
			materialized[capture.StableCaptureID(fact.Source, fact.ID)] = true
		}
		if err := completeMaterializedCaptureRestores(ctx, store, result.RestoresReady, func(control capture.CaptureControl) bool {
			return materialized[control.CaptureID]
		}); err != nil {
			return fmt.Errorf("artifacts committed; restore coverage acknowledgement pending: %w", err)
		}
		return nil
	})
	return blocked, err
}

// recoverMarkedCapturePublication resolves only existing recovery authority.
// Discard never admits a destination hint or starts another publication.
func recoverMarkedCapturePublication(ctx context.Context, project, ticket string, cfg domain.ProjectConfig, protection *capture.ProtectedPublication) (returnErr error) {
	lease, err := capture.AcquireInitialPublication(ctx, project, adapters.PublicationTicketKey(ticket, cfg.Ticket), protection)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lease.Close()) }()
	prepared, err := lease.Read()
	if err != nil {
		return err
	}
	var location adapters.WorkspaceLocation
	var recoverySelection *capture.InitialPublication
	pathsFor := func(ledger string) ([]string, error) {
		summary, err := resolveGenerationSummaryPath(project, cfg, "", false, ledger, ticket)
		if err != nil {
			return nil, err
		}
		paths := authorizedPublicationPaths(ledger, summary)
		for _, path := range paths {
			if err := adapters.ValidateWorkspaceArtifactDestination(cfg, ledger, path); err != nil {
				return nil, err
			}
		}
		return paths, nil
	}
	marked := false
	if prepared != nil {
		markerErr := adapters.CheckPublication(prepared.Path)
		marked = errors.Is(markerErr, adapters.ErrPublicationPending)
		if markerErr != nil && !marked {
			return fmt.Errorf("reading retained publication marker for %s: %w", ticket, markerErr)
		}
	}
	if marked {
		location, recoverySelection, err = adapters.SelectGenerationWorkspacePublication(cfg, ticket, "", lease, true)
	} else {
		location, err = adapters.SelectWorkspacePublicationRecovery(ctx, cfg, project, ticket, pathsFor, protection)
	}
	defer func() { returnErr = errors.Join(returnErr, location.Close()) }()
	if err != nil {
		var resolution *adapters.WorkspaceResolutionError
		if errors.As(err, &resolution) && resolution.Outcome == "not_found" {
			return nil
		}
		return fmt.Errorf("resolving existing publication recovery for %s: %w", ticket, err)
	}
	if marked && location.LedgerPath != prepared.Path {
		return fmt.Errorf("retained transaction for %s conflicts with the current canonical location", ticket)
	}
	if err := adapters.CheckPublication(location.LedgerPath); !errors.Is(err, adapters.ErrPublicationPending) {
		return err
	}
	paths, err := pathsFor(location.LedgerPath)
	if err != nil {
		return err
	}
	owner, err := location.AcquirePublicationWithProtection(ctx, project, protection)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, owner.Close()) }()
	if recoverySelection != nil {
		if err := adapters.ValidatePreparedWorkspacePublication(cfg, ticket, location.LedgerPath, paths, owner, recoverySelection); err != nil {
			return err
		}
	} else if err := adapters.ValidateOrdinaryPublicationRecovery(cfg, ticket, location.LedgerPath, paths); err != nil {
		return err
	}
	_, err = adapters.RecoverPublication(location.LedgerPath, paths)
	if err != nil {
		return fmt.Errorf("recovering interrupted publication for %s before discard cleanup: %w", ticket, err)
	}
	return nil
}

func reportGenerationPublicationBlock(project, registry, ticket, ledger, summary string, cause error) {
	fmt.Fprintln(os.Stderr, "Automatic publication pending:", cause)
	if capture.IsRegistryContention(cause) {
		os.Exit(capturePublicationBusyExit)
	}
	if err := recordGenerationPublication(project, registry, ticket, "failed", ledger, summary, nil, cause.Error()); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot persist publication block:", err)
		os.Exit(1)
	}
	os.Exit(capturePublicationBlockedExit)
}

func retainBlockedFirstPublication(project, registry, pending, ticket, ledger, summary string, cutoff time.Time, expectedRevision int64, policy domain.CaptureConfig, generationPolicy domain.GenerationPolicy, protection *capture.ProtectedPublication, owner *capture.PublicationLock, cause error) {
	entries, reconciledAt, pendingErr := reconcilePendingDestination(project, registry, pending, ticket, cutoff, expectedRevision, policy, generationPolicy, protection, owner)
	if pendingErr == nil {
		pendingErr = recordGenerationPendingSnapshots(project, registry, pending, "", cause.Error(), reconciledAt, entries)
	}
	reportGenerationPublicationBlock(project, registry, ticket, ledger, summary, errors.Join(cause, pendingErr))
}

func recordPendingSnapshots(ctx context.Context, store *capture.Store, directory, outcome, detail, reconciledAt string, entries []adapters.PendingSnapshotEntry) error {
	for _, entry := range entries {
		snapshot := entry.Snapshot
		state := outcome
		if state == "" {
			state = snapshot.State
		}
		root := domain.CapturePublishedOwner{Source: snapshot.Source, SessionID: snapshot.RootSessionID}
		for _, owner := range snapshot.Sessions {
			if owner.ID != root.SessionID {
				continue
			}
			if checkpoint, known := domain.CaptureOwnerCheckpoint(owner.ToDomain()); known {
				root.Checkpoint = checkpoint.Format(time.RFC3339Nano)
			}
			break
		}
		record := domain.CaptureDiagnostic{
			TicketID: snapshot.TicketID, Phase: "pending_snapshot", Outcome: state, Detail: detail,
			Destination: filepath.Join(directory, snapshot.CaptureID+".json"),
			Owners:      []domain.CapturePublishedOwner{root},
			Reconciliation: &domain.CaptureReconciliationDiagnostic{
				StartedAt: snapshot.SourceCutoff, Cutoff: reconciledAt,
			},
		}
		if err := store.RecordDiagnostic(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func recordGenerationPendingSnapshots(project, registry, directory, outcome, detail, reconciledAt string, entries []adapters.PendingSnapshotEntry) error {
	if len(entries) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return err
	}
	return errors.Join(recordPendingSnapshots(ctx, store, directory, outcome, detail, reconciledAt, entries), store.Close())
}

func checkExpectedCaptureLifecycle(expected, observed int64) error {
	if expected >= 0 && observed != expected {
		return capture.ErrCaptureLifecycleChanged
	}
	return nil
}

func reconcilePendingDestination(project, registry, pending, ticket string, cutoff time.Time, expectedRevision int64, policy domain.CaptureConfig, generationPolicy domain.GenerationPolicy, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (entries []adapters.PendingSnapshotEntry, reconciledAt string, returnErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return nil, "", err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	if err := store.SetSourceBudgets(policy.MaxSessions, policy.MaxSourceBytes); err != nil {
		return nil, "", err
	}
	result, err := store.ReconcileAt(ctx, ticket, git.GetGitUserInDirContext(ctx, project), policy.MaxSessions, policy.MaxSourceBytes, cutoff, generationPolicy)
	if err != nil {
		return nil, "", err
	}
	if err := checkExpectedCaptureLifecycle(expectedRevision, result.LifecycleRevision); err != nil {
		return nil, result.Cutoff, err
	}
	enabled, err := automaticCaptureEnabled(project, registry)
	if err != nil || !enabled {
		return nil, result.Cutoff, err
	}
	entries, err = syncPendingCapture(ctx, store, pending, project, result, true, protection, canonical)
	return entries, result.Cutoff, err
}

func syncGenerationPendingSnapshots(project, registry, directory string, result capture.ReconcileResult, create bool, policy domain.CaptureConfig, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (entries []adapters.PendingSnapshotEntry, returnErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return nil, err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	return syncPendingCapture(ctx, store, directory, project, result, create, protection, canonical)
}

func syncPendingCapture(ctx context.Context, store *capture.Store, directory, project string, result capture.ReconcileResult, create bool, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) ([]adapters.PendingSnapshotEntry, error) {
	entries, err := adapters.SyncPendingSnapshots(ctx, store, directory, project, result, create, protection, canonical)
	if err != nil || len(entries) == 0 || len(result.RestoresReady) == 0 {
		return entries, err
	}
	materialized := make(map[string]bool, len(entries))
	for _, entry := range entries {
		materialized[entry.Snapshot.CaptureID] = true
	}
	return entries, store.WithCaptureLifecycle(ctx, result.LifecycleRevision, func() error {
		return completeMaterializedCaptureRestores(ctx, store, result.RestoresReady, func(control capture.CaptureControl) bool {
			return materialized[control.CaptureID]
		})
	})
}

func completeMaterializedCaptureRestores(ctx context.Context, store *capture.Store, ready []capture.CaptureControl, materialized func(capture.CaptureControl) bool) error {
	complete := make([]capture.CaptureControl, 0, len(ready))
	for _, control := range ready {
		if materialized(control) {
			complete = append(complete, control)
		}
	}
	if len(complete) == 0 {
		return nil
	}
	return store.CompleteCaptureRestores(ctx, complete)
}

func verifyPendingPromotionCoverage(published, refreshed, pending []domain.SessionFact) error {
	current := make(map[string]domain.SessionFact, len(published))
	for _, owner := range published {
		current[owner.ID] = owner
	}
	fresh := make(map[string]bool, len(refreshed))
	for _, owner := range refreshed {
		fresh[owner.ID] = true
	}
	for _, owner := range pending {
		accepted, ok := current[owner.ID]
		if !ok || accepted.Source != owner.Source {
			return fmt.Errorf("canonical publication does not cover pending owner %s", owner.ID)
		}
		if fresh[owner.ID] {
			continue
		}
		proposed, observed := adapters.SessionFactDTOFromDomain(accepted), adapters.SessionFactDTOFromDomain(owner)
		// Canonical attribution can differ, but unavailable owner evidence
		// (including portable anomaly summaries) must be covered.
		proposed.User, observed.User = "", ""
		if !reflect.DeepEqual(proposed, observed) {
			return fmt.Errorf("canonical and pending facts disagree for unavailable owner %s; preserving both until source verification is possible", owner.ID)
		}
	}
	return nil
}

func promoteGenerationPendingSnapshots(project, registry, directory, reconciledAt string, revision int64, entries []adapters.PendingSnapshotEntry, policy domain.CaptureConfig, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (returnErr error) {
	if len(entries) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	if err := adapters.RemovePromotedPendingSnapshots(ctx, store, revision, directory, project, entries, protection, canonical); err != nil {
		if errors.Is(err, capture.ErrCaptureLifecycleChanged) {
			return nil // A later lifecycle request owns the pending snapshot.
		}
		return err
	}
	return recordPendingSnapshots(ctx, store, directory, "promoted", "Pending snapshot promoted after durable canonical publication.", reconciledAt, entries)
}

// recoverFailedGenerationPublication recovers the artifact set after a handled
// replacement failure. The caller still owns the publication lock.
// Process interruptions retain the same journal for recovery on the next run.
func recoverFailedGenerationPublication(ledger string, paths []string, cause error) error {
	if ledger == "" {
		return cause
	}
	recovered, err := adapters.RecoverPublication(ledger, paths)
	if err != nil {
		return errors.Join(cause, fmt.Errorf("publication rollback incomplete or finalization failed; inspect the recovery journal and retry publication: %w", err))
	}
	if !recovered {
		// Commit can remove the journal before its final directory sync fails.
		// Without the prior bytes, restoration cannot be claimed or fabricated.
		return errors.Join(cause, errors.New("no prior-good publication journal remains; replacements may already be committed and restoration could not be confirmed"))
	}
	return fmt.Errorf("%w; recovered interrupted artifact publication", cause)
}
