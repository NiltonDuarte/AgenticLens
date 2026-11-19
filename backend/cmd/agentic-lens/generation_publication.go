package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
	if _, _, conflict := domain.CaptureTicketMetadata(rows, ticket); conflict != nil {
		// Canonical history and exact interrupted recovery can proceed.
		// Conflicting types cannot authorize an unmarked first write.
		location, selection, err := adapters.SelectGenerationWorkspacePublication(cfg, ticket, path, lease, true)
		if err != nil || (selection != nil && !errors.Is(adapters.CheckPublication(selection.Path), adapters.ErrPublicationPending)) {
			return adapters.WorkspaceLocation{}, nil, conflict, errors.Join(conflict, err, location.Close())
		}
		return location, selection, conflict, nil
	}
	location, selection, err := adapters.SelectNativeWorkspacePublication(ctx, cfg, project, ticket, path, lease, protection, func() ([]domain.CaptureRegistration, error) {
		return rows, nil
	})
	return location, selection, nil, err
}

func nativeGenerationRegistrations(ctx context.Context, project, registry string) ([]domain.CaptureRegistration, error) {
	store, err := capture.OpenProjectContext(ctx, registry, project)
	if err != nil {
		return nil, err
	}
	rows, readErr := store.List(ctx)
	return rows, errors.Join(readErr, store.Close())
}

func beginAutomaticFirstPublication(project, registry, ticket, ledger string, paths []string, policy domain.CaptureConfig, selection *capture.InitialPublication) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	admissionFailed := false
	err := adapters.BeginInitialPublicationWithAdmission(ledger, paths, selection, func(publishIndicator func() error) error {
		store, err := capture.OpenProjectContext(ctx, registry, project)
		if err != nil {
			admissionFailed = true
			return err
		}
		started := false
		admitErr := store.WithInitialPublicationRegistrations(ctx, ticket, func(rows []domain.CaptureRegistration) error {
			if _, _, err := domain.CaptureTicketMetadata(rows, ticket); err != nil {
				return err
			}
			if err := adapters.ValidateCurrentNativePublication(project, ticket, selection, rows); err != nil {
				return err
			}
			started = true
			return publishIndicator()
		})
		admissionFailed = admitErr != nil && !started
		return errors.Join(admitErr, store.Close())
	})
	return admissionFailed, err
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

func reconcilePendingDestination(project, registry, pending, ticket string, cutoff time.Time, policy domain.CaptureConfig, generationPolicy domain.GenerationPolicy, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (entries []adapters.PendingSnapshotEntry, reconciledAt string, returnErr error) {
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
	enabled, err := automaticCaptureEnabled(project, registry)
	if err != nil || !enabled {
		return nil, result.Cutoff, err
	}
	entries, err = adapters.SyncPendingSnapshots(ctx, pending, project, result, true, protection, canonical)
	return entries, result.Cutoff, err
}

func syncGenerationPendingSnapshots(project, directory string, result capture.ReconcileResult, create bool, policy domain.CaptureConfig, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) ([]adapters.PendingSnapshotEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	return adapters.SyncPendingSnapshots(ctx, directory, project, result, create, protection, canonical)
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
		if len(proposed.Anomalies) == 0 && len(observed.Anomalies) == 0 {
			proposed.Anomalies, observed.Anomalies = nil, nil
		}
		if !reflect.DeepEqual(proposed, observed) {
			return fmt.Errorf("canonical and pending facts disagree for unavailable owner %s; preserving both until source verification is possible", owner.ID)
		}
	}
	return nil
}

func promoteGenerationPendingSnapshots(project, registry, directory, reconciledAt string, entries []adapters.PendingSnapshotEntry, policy domain.CaptureConfig, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) error {
	if len(entries) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(policy.PassTimeoutMS)*time.Millisecond)
	defer cancel()
	if err := adapters.RemovePromotedPendingSnapshots(ctx, directory, project, entries, protection, canonical); err != nil {
		return err
	}
	return recordGenerationPendingSnapshots(project, registry, directory, "promoted", "Pending snapshot promoted after durable canonical publication.", reconciledAt, entries)
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
