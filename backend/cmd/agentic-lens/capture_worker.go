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
	"agenticlens/internal/domain"
)

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
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	fence, err := newCaptureWorkerFence(project, state, binary, cfg, protectedHome)
	if err != nil {
		return err
	}
	binary = fence.binary
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
			return nil
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
	var lastPass time.Time
	var nextAttempt time.Time
	var lastFailure string
	successfulPass := false
	backoff := time.Duration(policy.RefreshMS) * time.Millisecond
	lastActivity := time.Now()
	for time.Since(started) < time.Duration(policy.MaxLifetimeMS)*time.Millisecond {
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
		active, activityDetail := captureWorkerActive(project, rowsForActivity, policy.MaxSourceBytes)
		if activityDetail != "" {
			lastFailure = activityDetail
		}
		if scan.Remaining == 0 && time.Now().UnixMilli() >= scan.RetryUntil && !active && time.Since(lastActivity) >= time.Duration(policy.IdleMS)*time.Millisecond && now.UnixMilli()-req.RequestedAt >= int64(policy.IdleMS) && req.Generation == req.Acknowledged {
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
		due := req.Generation != req.Acknowledged || time.Since(lastPass) >= time.Duration(policy.RefreshMS)*time.Millisecond
		if due && !now.Before(nextAttempt) && (time.Since(lastPass) >= time.Duration(policy.RefreshMS)*time.Millisecond || now.UnixMilli()-req.RequestedAt >= int64(policy.DebounceMS)) {
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
			unassigned, unassignedErr := s.Reconcile(unassignedCtx, "", "", policy.MaxSessions, policy.MaxSourceBytes)
			unassignedCancel()
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
			for _, registration := range unassigned.Registrations {
				if registration.State == domain.CaptureFailed || registration.State == domain.CapturePending {
					success = false
				}
			}
			for _, ticket := range ordered {
				enabled, e := automaticCaptureEnabled(project, state)
				if e != nil || !enabled {
					return e
				}
				location, pathErr := adapters.ResolveTicketLocation(cfg, ticket, "")
				if pathErr != nil {
					success = false
					lastFailure = pathErr.Error()
					if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: "failed", Detail: "Resolve an existing ledger or explicitly generate its first destination: " + pathErr.Error()}); e != nil {
						return e
					}
					continue
				}
				dest := location.LedgerPath

				args := []string{"generate", "--ticket-id", ticket, "--json-output", dest, "--capture-dir", state, "--capture-auto", "--max-sessions", strconv.Itoa(policy.MaxSessions), "--max-source-bytes", strconv.FormatInt(policy.MaxSourceBytes, 10)}
				title, kind, metadataErr := domain.CaptureTicketMetadata(rows, ticket)
				if metadataErr != nil {
					success = false
					lastFailure = metadataErr.Error()
					if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: "failed", Destination: dest, Detail: "Publication metadata unavailable or conflicting. Supply an explicit ticket title/type through capture bind and retry reconciliation."}); e != nil {
						return e
					}
					fmt.Fprintln(os.Stderr, "automatic publication pending:", metadataErr)
					continue
				}
				if title != "" {
					args = append(args, "--ticket-title", title)
				}
				if kind != "" {
					args = append(args, "--ticket-type", kind)
				}
				remaining := time.Duration(policy.MaxLifetimeMS)*time.Millisecond - time.Since(started)
				if remaining <= 0 {
					success = false
					break
				}
				if err := s.AttemptPublicationTicket(context.Background(), ticket); err != nil {
					return err
				}
				ctx, cancel := context.WithTimeout(context.Background(), min(remaining, time.Duration(policy.PassTimeoutMS)*time.Millisecond))
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
					lastFailure = err.Error()
					if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: "failed", Destination: dest, Detail: "Native generation failed or timed out. Source acceptance does not prove publication; inspect diagnostics and recover interrupted output with generate."}); e != nil {
						return e
					}
					fmt.Fprintln(os.Stderr, "automatic publication pending:", err)
					if e := capture.WriteWorkerProgress(state, capture.WorkerProgress{BindingScan: scan, Executable: binary, Build: domain.GetBuildInfo(), PID: os.Getpid(), State: "failure", Heartbeat: time.Now().UnixMilli(), Deadline: started.Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond).UnixMilli(), NextAttempt: time.Now().Add(backoff * 2).UnixMilli(), Detail: err.Error()}); e != nil {
						return e
					}
				}
			}
			if success {
				if e := s.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{Phase: "worker", Outcome: "completed", Detail: "Bounded native worker pass completed; ticket publication diagnostics identify accepted artifacts."}); e != nil {
					return e
				}
				successfulPass = true
				lastFailure = ""
				backoff = time.Duration(policy.RefreshMS) * time.Millisecond
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
			if !success {
				backoff = min(backoff*2, time.Duration(policy.RefreshMS)*8*time.Millisecond)
			}
			nextAttempt = time.Now().Add(backoff)
			lastPass = time.Now()
		}
		time.Sleep(time.Duration(policy.RefreshMS) * time.Millisecond)
	}
	current, e := fence.current(project, state)
	if e != nil {
		return e
	}
	if !current {
		return retire()
	}
	// Failed publication cannot create an unbounded retry process chain. Durable
	// pending work restarts on the next supported lifecycle/reconcile integration.
	// An unfinished inbox sweep can hand off if this owner advanced its saved
	// cursor. Publication failures still stop retries once the sweep finishes.
	if (!successfulPass || lastFailure != "") && !(madeInboxProgress && scan.Remaining > 0) && time.Now().UnixMilli() >= scan.RetryUntil {
		return fmt.Errorf("native worker lifetime exhausted; pending generation retained for next integration event: %s", lastFailure)
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
