package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

// Internal publisher status for a bounded retry of SQLite registry contention.
const capturePublicationBusyExit = 75
const capturePublicationBlockedExit = 76

func captureWorker(s *capture.Store, project, state string, policy domain.CaptureConfig, protectedHome string) (returnErr error) {
	if err := policy.ValidateWorker(); err != nil {
		return err
	}
	defer func() {
		if returnErr == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		returnErr = errors.Join(returnErr, s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{Phase: "worker", Outcome: "failed", Detail: "Native worker stopped before completing its pass. Accepted facts and pending generations remain recoverable; inspect capture diagnostics and retry capture reconcile."}))
	}()
	lock, err := capture.AcquireWorker(state)
	if err != nil || lock == nil {
		return err
	}
	held := true
	defer func() {
		if held {
			if e := lock.Close(); e != nil {
				fmt.Fprintln(os.Stderr, "worker unlock:", e)
			}
		}
	}()
	// Admission and worker ownership use the same lease. A process queued by
	// ordinary routing before admission rechecks authority before native probes.
	admitted, err := captureWorkerHome(project, protectedHome)
	if err != nil {
		return err
	}
	if admitted != protectedHome {
		if err := lock.Close(); err != nil {
			return err
		}
		held = false
		return launchCaptureWorker("", project, state, admitted)
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return err
	}
	if _, err := adaptersconfig.ResolvePendingDirectory(cfg, project); err != nil {
		return fmt.Errorf("pending capture configuration: %w", err)
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	fence, err := newCaptureWorkerFence(project, state, binary, cfg, protectedHome)
	if err != nil {
		return err
	}
	binary = fence.binary
	var protection *capture.ProtectedPublication
	if protectedHome != "" {
		protection = &fence.publication
	}
	previous, err := capture.WorkerStatus(state)
	if err != nil {
		if !errors.Is(err, capture.ErrInvalidWorkerStatus) {
			return err
		}
		// The progress file is replaceable operational evidence. Restart the
		// inbox sweep from registrations; accepted ledger facts stay intact.
		if err := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "recovered", Detail: "Invalid worker status. Restarting the bounded binding scan from registered sessions."}); err != nil {
			return err
		}
		previous = capture.WorkerProgress{}
	}
	scan := previous.BindingScan
	madeInboxProgress := false
	retire := func() error {
		// Disablement may race an earlier policy read and make the fence false.
		// It stops this owner without a superseded diagnostic or successor.
		enabled, err := automaticCaptureEnabled(project, state)
		if err != nil || !enabled {
			return err
		}
		if err := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "superseded", Detail: "Project release or active installation changed. This native owner retired; only its captured generation may be acknowledged and later work remains pending for the selected native."}); err != nil {
			return err
		}
		if held {
			if err := lock.Close(); err != nil {
				return err
			}
			held = false
		}
		pending, err := s.PublicationRequest(context.Background())
		if err != nil {
			return err
		}
		rows, err := s.List(context.Background())
		if err != nil {
			return err
		}
		active, _ := captureWorkerActive(project, rows, policy.MaxSourceBytes)
		if pending.Generation == pending.Acknowledged && !active && scan.Remaining == 0 && time.Now().UnixMilli() >= scan.RetryUntil {
			children, err := s.CodexChildDiscoveryCursors(context.Background())
			if err != nil || len(children) == 0 {
				return err
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err = launchSelectedCaptureWorker(project, state, protectedHome, func() (string, error) {
			return capture.WorkerSuccessor(ctx, project, state)
		})
		if err != nil {
			// A later disable can also stop successor validation normally. Keep
			// actual policy/state errors and enabled takeover failures visible.
			enabled, policyErr := automaticCaptureEnabled(project, state)
			if policyErr != nil {
				return errors.Join(err, policyErr)
			}
			if !enabled {
				return nil
			}
			diagnosticErr := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "failed", Detail: "Worker release/installation changed, but automatic successor validation failed. Pending work remains in the original namespace. Verify the exact pin, selected native/protocol, owned integration and configured state; retry capture reconcile. Cause: " + err.Error()})
			return errors.Join(fmt.Errorf("native worker retired; pending work retained; selected-native takeover unavailable: %w", err), diagnosticErr)
		}
		return nil
	}
	started := time.Now()
	refresh := time.Duration(policy.RefreshMS) * time.Millisecond
	debounce := time.Duration(policy.DebounceMS) * time.Millisecond
	deadline := started.Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond)
	var lastPass time.Time
	var nextAttempt time.Time
	var wakeAfterBlockedGeneration int64 = -1
	var lastFailure string
	successfulPass := false
	shortRegistryRetryUsed := false
	backoff := refresh
	lastActivity := time.Now()
	initialChildren, err := s.CodexChildDiscoveryCursors(context.Background())
	if err != nil {
		return err
	}
	active := false
	for time.Now().Before(deadline) {
		// Capture before fencing: a request delivered after a release switch
		// cannot be pulled into an older owner's already accepted pass.
		req, err := s.PublicationRequest(context.Background())
		if err != nil {
			return err
		}
		enabled, e := automaticCaptureEnabled(project, state)
		if e != nil || !enabled {
			return e
		}
		current, e := fence.current(project, state)
		if e != nil {
			return e
		}
		if !current {
			return retire()
		}
		now := time.Now()
		remaining := time.Duration(policy.MaxLifetimeMS)*time.Millisecond - time.Since(started)
		if remaining <= 0 {
			break
		}
		proofCtx, proofCancel := context.WithTimeout(context.Background(), min(remaining, time.Duration(policy.PassTimeoutMS)*time.Millisecond))
		proofErr := s.RefreshProjectProofs(proofCtx, min(4, policy.MaxSessions), min(policy.MaxSourceBytes, int64(4*capture.MaxPayloadBytes)))
		proofCancel()
		if proofErr != nil {
			return proofErr
		}
		rowsForActivity, activityErr := s.List(context.Background())
		if activityErr != nil {
			return activityErr
		}
		inboxCtx, inboxCancel := context.WithTimeout(context.Background(), min(time.Until(started.Add(time.Duration(policy.MaxLifetimeMS)*time.Millisecond)), time.Duration(policy.PassTimeoutMS)*time.Millisecond))
		lastRegistration := scan.Last
		accepted, inboxErr := s.ScanCodexBindings(inboxCtx, protectedHome, rowsForActivity, &scan, req.Generation, policy.MaxSessions, policy.MaxSourceBytes)
		inboxCancel()
		madeInboxProgress = madeInboxProgress || lastRegistration != scan.Last
		proofPending := errors.Is(inboxErr, capture.ErrBindingProofPending)
		firstPending := proofPending && scan.RetryUntil == 0
		if firstPending {
			scan.RetryUntil = time.Now().Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond).UnixMilli()
		}
		if err := capture.WriteWorkerProgress(state, capture.WorkerProgress{BindingScan: scan, Executable: binary, Build: domain.GetBuildInfo(), PID: os.Getpid(), State: "running", Heartbeat: now.UnixMilli(), Deadline: started.Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond).UnixMilli(), NextAttempt: nextAttempt.UnixMilli(), Detail: lastFailure + fmt.Sprintf(" generation=%d acknowledged=%d pass_timeout_ms=%d max_sessions=%d max_source_bytes=%d", req.Generation, req.Acknowledged, policy.PassTimeoutMS, policy.MaxSessions, policy.MaxSourceBytes)}); err != nil {
			return err
		}

		if inboxErr != nil && (!proofPending || firstPending) {
			outcome, detail := "failed", inboxErr.Error()
			if proofPending {
				outcome = "pending"
				detail += fmt.Sprintf(". Authenticated response retained. Source verification retries until %s and resumes on the next native callback or reconcile.", time.UnixMilli(scan.RetryUntil).UTC().Format(time.RFC3339))
			}
			if err := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "binding", Outcome: outcome, Detail: detail}); err != nil {
				return err
			}
		}
		if accepted > 0 {
			lastActivity = time.Now()
			req, err = s.PublicationRequest(context.Background())
			if err != nil {
				return err
			}
		}
		var activityDetail string
		active, activityDetail = captureWorkerActive(project, rowsForActivity, policy.MaxSourceBytes)
		if activityDetail != "" {
			lastFailure = activityDetail
		}
		if scan.Remaining == 0 && time.Now().UnixMilli() >= scan.RetryUntil && !active && time.Since(lastActivity) >= time.Duration(policy.IdleMS)*time.Millisecond && now.UnixMilli()-req.RequestedAt >= int64(policy.IdleMS) && req.Generation == req.Acknowledged {
			children, err := s.CodexChildDiscoveryCursors(context.Background())
			if err != nil {
				return err
			}
			if len(children) == 0 {
				// Release before checking again: an event racing shutdown either starts a
				// new owner or is seen here. Never consume a generation without a pass.
				if err = lock.Close(); err != nil {
					return err
				}
				held = false
				after, err := s.PublicationRequest(context.Background())
				if err != nil {
					return err
				}
				if after.Generation != after.Acknowledged {
					enabled, e := automaticCaptureEnabled(project, state)
					if e != nil || !enabled {
						return e
					}
					current, e := fence.current(project, state)
					if e != nil {
						return e
					}
					if !current {
						return retire()
					}
					return launchCaptureWorker(binary, project, state, protectedHome)
				}
				return nil
			}
		}
		now = time.Now()
		periodicDue := now.Sub(lastPass) >= refresh
		retryDue := !nextAttempt.IsZero() && !now.Before(nextAttempt)
		requestDue := captureRequestReady(req, debounce, now)
		newWorkAfterBlock := wakeAfterBlockedGeneration >= 0 && req.Generation > wakeAfterBlockedGeneration && requestDue
		if (!now.Before(nextAttempt) || newWorkAfterBlock) && (retryDue || periodicDue || requestDue) {
			current, e := fence.current(project, state)
			if e != nil {
				return e
			}
			if !current {
				return retire()
			}
			remaining := time.Duration(policy.MaxLifetimeMS)*time.Millisecond - time.Since(started)
			if remaining <= 0 {
				break
			}
			unassignedCtx, unassignedCancel := context.WithTimeout(context.Background(), min(remaining, time.Duration(policy.PassTimeoutMS)*time.Millisecond))
			passConfig, err := adaptersconfig.Load(project)
			if err != nil {
				unassignedCancel()
				return err
			}
			pendingDirectory, err := adaptersconfig.ResolvePendingDirectory(passConfig, project)
			if err != nil {
				unassignedCancel()
				return fmt.Errorf("pending capture configuration: %w", err)
			}
			unassigned, unassignedErr := s.Reconcile(unassignedCtx, "", git.GetGitUserInDirContext(unassignedCtx, project), policy.MaxSessions, policy.MaxSourceBytes, passConfig.Generation.Metrics)
			if unassignedErr == nil {
				writePending := func() error {
					enabled, err := automaticCaptureEnabled(project, state)
					if err != nil || !enabled {
						return err
					}
					entries, err := adapters.SyncPendingSnapshots(unassignedCtx, pendingDirectory, project, unassigned, true, protection, nil)
					if err != nil {
						return err
					}
					return recordPendingSnapshots(unassignedCtx, s, pendingDirectory, "", "Native metrics captured before ticket binding.", unassigned.Cutoff, entries)
				}
				if protection == nil {
					unassignedErr = adaptersconfig.WithOrdinaryCapture(project, writePending)
				} else {
					unassignedErr = writePending()
				}
			}
			unassignedCancel()
			if errors.Is(unassignedErr, adaptersconfig.ErrProtectedCodexProject) {
				return retire()
			}
			if unassignedErr != nil {
				lastFailure = unassignedErr.Error()
			}
			rows, err := s.List(context.Background())
			if err != nil {
				return err
			}
			tickets := map[string]bool{}
			for _, r := range rows {
				if r.TicketID != "" {
					tickets[r.TicketID] = true
				}
			}
			ordered := []string{}
			for ticket := range tickets {
				ordered = append(ordered, ticket)
			}
			sort.Strings(ordered)
			ordered, err = s.PublicationTickets(context.Background(), ordered)
			if err != nil {
				return err
			}
			success := unassignedErr == nil
			busyOnly := unassignedErr == nil || capture.IsRegistryContention(unassignedErr)
			refreshedUnassigned := map[string]bool{}
			for _, owner := range unassigned.RefreshedSessions {
				refreshedUnassigned[owner.ID] = true
			}
			for _, registration := range unassigned.Registrations {
				if registration.State == domain.CaptureFailed || (registration.State == domain.CapturePending && !refreshedUnassigned[registration.Identity.SessionID]) {
					success = false
					busyOnly = false
					lastFailure = registration.Detail
				}
			}
			publicationBlocked := false
			for _, ticket := range ordered {
				enabled, e := automaticCaptureEnabled(project, state)
				if e != nil || !enabled {
					return e
				}
				remaining := time.Duration(policy.MaxLifetimeMS)*time.Millisecond - time.Since(started)
				if remaining <= 0 {
					success = false
					busyOnly = false
					break
				}
				if err := s.AttemptPublicationTicket(context.Background(), ticket); err != nil {
					return err
				}
				ctx, cancel := context.WithTimeout(context.Background(), min(remaining, time.Duration(policy.PassTimeoutMS)*time.Millisecond))
				// The writer resolves any disagreement against its owned ledger.
				title, kind, _ := domain.CaptureTicketMetadata(rows, ticket)
				// The writer resolves and retains its own directory selection
				// rather than receiving a path previously selected by this process.
				args := []string{"generate", "--ticket-id", ticket, "--capture-dir", state, "--capture-auto", "--max-sessions", strconv.Itoa(policy.MaxSessions), "--max-source-bytes", strconv.FormatInt(policy.MaxSourceBytes, 10)}
				if title != "" {
					args = append(args, "--ticket-title", title)
				}
				if kind != "" {
					args = append(args, "--ticket-type", kind)
				}
				publisher := binary
				home, err := captureWorkerHome(project, protectedHome)
				if err != nil {
					cancel()
					return err
				}
				if home != "" {
					if protectedHome == "" {
						cancel()
						return retire()
					}
					p := fence.publication
					args = append([]string{"capture", "publish-protected", p.Project, p.Home, p.State}, args[1:]...)
					publisher, err = adaptersconfig.ResolveProtectedCodexEngine(project, home)
					if err != nil {
						cancel()
						return err
					}
					info, err := os.Stat(publisher)
					if err != nil {
						cancel()
						return err
					}
					if !os.SameFile(info, fence.artifact) {
						cancel()
						return retire()
					}
				}
				generate := func() error {
					cmd := exec.CommandContext(ctx, publisher, args...)
					cmd.Dir = project
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					return cmd.Run()
				}
				if home == "" {
					err = adaptersconfig.WithOrdinaryCapture(project, generate)
				} else {
					err = generate()
				}
				cancel()
				if errors.Is(err, adaptersconfig.ErrProtectedCodexProject) {
					return retire()
				}
				if err != nil {
					success = false
					var exit *exec.ExitError
					exited := errors.As(err, &exit)
					busyOnly = busyOnly && exited && exit.ExitCode() == capturePublicationBusyExit
					lastFailure = err.Error()
					if exited && exit.ExitCode() == capturePublicationBlockedExit {
						publicationBlocked = true
					} else {
						if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: "failed", Detail: "Native publication failed or timed out. Inspect pending snapshots and publication diagnostics."}); e != nil {
							return e
						}
					}
					fmt.Fprintln(os.Stderr, "automatic publication pending:", err)
				}
			}
			if success {
				if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "completed", Detail: "Bounded native worker pass completed; ticket publication diagnostics identify accepted artifacts."}); e != nil {
					return e
				}
				successfulPass = true
				lastFailure = ""
				backoff = refresh
				shortRegistryRetryUsed = false
				nextAttempt = time.Time{}
				wakeAfterBlockedGeneration = -1
				if err = s.AcknowledgePublication(context.Background(), req.Generation); err != nil {
					return err
				}
			}
			afterRows, readErr := s.List(context.Background())
			if readErr != nil {
				return readErr
			}
			before := map[string]string{}
			for _, r := range rows {
				before[capture.IdentityKey(r.Identity)] = r.SourceFingerprint
			}
			for _, r := range afterRows {
				if before[capture.IdentityKey(r.Identity)] != r.SourceFingerprint {
					lastActivity = time.Now()
				}
			}
			lastPass = time.Now()
			if !success {
				wakeAfterBlockedGeneration = -1
				if publicationBlocked {
					wakeAfterBlockedGeneration = req.Generation
				}
				delay := backoff
				if busyOnly && !shortRegistryRetryUsed {
					// The next pass retains the normal rotated ticket order,
					// session/byte allowances and remaining worker lifetime.
					shortRegistryRetryUsed = true
					delay = min(refresh, 250*time.Millisecond)
				} else {
					backoff = min(backoff*2, 8*refresh)
					delay = backoff
				}
				nextAttempt = lastPass.Add(delay)
				if e := capture.WriteWorkerProgress(state, capture.WorkerProgress{BindingScan: scan, Executable: binary, Build: domain.GetBuildInfo(), PID: os.Getpid(), State: "failure", Heartbeat: lastPass.UnixMilli(), Deadline: deadline.UnixMilli(), NextAttempt: nextAttempt.UnixMilli(), Detail: lastFailure}); e != nil {
					return e
				}
			}
		}
		if err := waitForCaptureRequest(s, refresh, debounce, nextAttempt, deadline, wakeAfterBlockedGeneration); err != nil {
			return err
		}
	}
	current, e := fence.current(project, state)
	if e != nil {
		return e
	}
	if !current {
		return retire()
	}
	children, err := s.CodexChildDiscoveryCursors(context.Background())
	if err != nil {
		return err
	}
	childProgress := codexChildDiscoveryAdvanced(initialChildren, children)
	inboxProgress := (madeInboxProgress && scan.Remaining > 0) || time.Now().UnixMilli() < scan.RetryUntil
	// A bounded successor can finish an advancing discovery sweep. A hint
	// whose sweep completed or made no progress cannot create a retry chain.
	if !childProgress && !inboxProgress {
		if !successfulPass || lastFailure != "" {
			return fmt.Errorf("native worker lifetime exhausted; pending generation retained for next integration event: %s", lastFailure)
		}
		pending, err := s.PublicationRequest(context.Background())
		if err != nil {
			return err
		}
		if !active && pending.Generation == pending.Acknowledged {
			return nil
		}
	}
	// A healthy finite owner transfers active work to another bounded native process.
	if err = lock.Close(); err != nil {
		return err
	}
	held = false
	enabled, e := automaticCaptureEnabled(project, state)
	if e != nil || !enabled {
		return e
	}
	current, e = fence.current(project, state)
	if e != nil {
		return e
	}
	if !current {
		return retire()
	}
	return launchSelectedCaptureWorker(project, state, protectedHome, func() (string, error) {
		entry, err := adaptersconfig.ResolveEntrypoint(project)
		if err != nil {
			return "", err
		}
		return entry.NativePath(adaptersconfig.RuntimeCodexHome(), cfg.Telemetry.Sources.CodexHome)
	})
}

func codexChildDiscoveryAdvanced(before, after map[string]string) bool {
	for parent, current := range after {
		if prior, existed := before[parent]; !existed || current > prior {
			return true
		}
	}
	return false
}

func captureRequestReady(request capture.PublicationRequest, debounce time.Duration, now time.Time) bool {
	return request.Generation != request.Acknowledged && !now.Before(time.UnixMilli(request.RequestedAt).Add(debounce))
}

// waitForCaptureRequest checks only the durable worktree request between
// periodic passes. Source proof, activity and extraction stay in the worker
// pass. An unchanged destination block retains backoff while a subsequent
// native callback can wake the worker after the workspace has changed.
func waitForCaptureRequest(store *capture.Store, refresh, debounce time.Duration, nextAttempt, deadline time.Time, wakeAfterBlockedGeneration int64) error {
	now := time.Now()
	wake := now.Add(refresh)
	if deadline.Before(wake) {
		wake = deadline
	}
	if now.Before(nextAttempt) {
		if nextAttempt.Before(wake) {
			wake = nextAttempt
		}
		if wakeAfterBlockedGeneration < 0 {
			time.Sleep(time.Until(wake))
			return nil
		}
	}

	ctx, cancel := context.WithDeadline(context.Background(), wake)
	defer cancel()
	const requestPoll = 250 * time.Millisecond
	ticker := time.NewTicker(requestPoll)
	defer ticker.Stop()
	for ctx.Err() == nil {
		request, err := store.PublicationRequest(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		now := time.Now()
		if captureRequestReady(request, debounce, now) &&
			(!now.Before(nextAttempt) || (wakeAfterBlockedGeneration >= 0 && request.Generation > wakeAfterBlockedGeneration)) {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
	return nil
}

// Disabled policy stops new passes and handoffs. A publication already committing
// may finish within PassTimeoutMS; disable never removes accepted facts.
func automaticCaptureEnabled(project, state string) (bool, error) {
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil || !policy.Enabled {
		return false, err
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return false, err
	}
	if err := capture.CheckInstalledInboxIsolation(project, cfg.Telemetry.Sources.CodexHome); err != nil {
		return false, err
	}
	location, err := capture.PreviewStateLocation(state, project)
	if err != nil {
		return false, err
	}
	if location.Provisional {
		return false, nil
	}
	legacyEnabled, err := capture.AutomaticEnabled(location.CoordinationDirectory)
	if err != nil || legacyEnabled {
		return legacyEnabled, err
	}
	// A migrated user integration remains active independently of its preserved
	// inactive legacy receipt; fresh native adoption needs no project receipt.
	if cfg.Telemetry.Sources.CodexHome != "" {
		installed, err := capture.UserCodexInstalled(cfg.Telemetry.Sources.CodexHome)
		if err != nil || installed {
			return installed, err
		}
	}
	return capture.UserOpenCodeInstalled()
}
