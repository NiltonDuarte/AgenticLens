package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

func captureWorker(s *capture.Store, project, state string, policy domain.CaptureConfig) error {
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
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	started := time.Now()
	var lastPass time.Time
	var nextAttempt time.Time
	var lastFailure string
	successfulPass := false
	backoff := time.Duration(policy.RefreshMS) * time.Millisecond
	lastActivity := time.Now()
	for time.Since(started) < time.Duration(policy.MaxLifetimeMS)*time.Millisecond {
		enabled, e := automaticCaptureEnabled(project, state)
		if e != nil || !enabled {
			return e
		}
		req, err := s.PublicationRequest(context.Background())
		if err != nil {
			return err
		}
		now := time.Now()
		if err := capture.WriteWorkerProgress(state, capture.WorkerProgress{PID: os.Getpid(), State: "running", Heartbeat: now.UnixMilli(), Deadline: started.Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond).UnixMilli(), NextAttempt: nextAttempt.UnixMilli(), Detail: lastFailure + fmt.Sprintf(" generation=%d acknowledged=%d pass_timeout_ms=%d max_sessions=%d max_source_bytes=%d", req.Generation, req.Acknowledged, policy.PassTimeoutMS, policy.MaxSessions, policy.MaxSourceBytes)}); err != nil {
			return err
		}
		rowsForActivity, activityErr := s.List(context.Background())
		if activityErr != nil {
			return activityErr
		}
		active := false
		activityBudget := policy.MaxSourceBytes
		for _, r := range rowsForActivity {
			if r.TicketID != "" && r.Activity != "idle" && r.Event != "SessionEnd" && r.Event != "Stop" && r.Event != "session.idle" && r.Event != "session.error" {
				if r.ParentID != "" && r.Identity.Source == domain.SourceOpenCode && r.ProjectDirectory == project && activityBudget >= 4*1024*1024 {
					activityBudget -= 4 * 1024 * 1024
					db, e := opencode.OpenDB(r.Identity.Store)
					if e == nil {
						ctx, cancel := context.WithTimeout(context.Background(), time.Second)
						finished, proofErr := opencode.CaptureChildFinished(ctx, db.DB(), r.ParentID, r.Identity.SessionID)
						cancel()
						closeErr := db.Close()
						if proofErr == nil && closeErr == nil && finished {
							continue
						}
						if proofErr != nil || closeErr != nil {
							lastFailure = fmt.Sprintf("child activity proof unavailable: %v", errors.Join(proofErr, closeErr))
						}
					} else {
						lastFailure = fmt.Sprintf("child activity source unavailable: %v", e)
					}
				}
				active = true
			}
		}
		if !active && time.Since(lastActivity) >= time.Duration(policy.IdleMS)*time.Millisecond && now.UnixMilli()-req.RequestedAt >= int64(policy.IdleMS) && req.Generation == req.Acknowledged {
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
				return capture.LaunchWorker(binary, project, state)
			}
			return nil
		}
		due := req.Generation != req.Acknowledged || time.Since(lastPass) >= time.Duration(policy.RefreshMS)*time.Millisecond
		if due && !now.Before(nextAttempt) && (time.Since(lastPass) >= time.Duration(policy.RefreshMS)*time.Millisecond || now.UnixMilli()-req.RequestedAt >= int64(policy.DebounceMS)) {
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
			success := true
			for _, ticket := range ordered {
				enabled, e := automaticCaptureEnabled(project, state)
				if e != nil || !enabled {
					return e
				}
				dest := filepath.Join(project, cfg.Workspaces.Dir, ticket, "agentic_metrics.json")
				if err = capture.ValidateDestination(project, dest); err != nil {
					return err
				}
				args := []string{"generate", "--ticket-id", ticket, "--json-output", dest, "--capture-dir", state, "--capture-auto"}
				title, kind, metadataErr := domain.CaptureTicketMetadata(rows, ticket)
				if metadataErr != nil {
					success = false
					lastFailure = metadataErr.Error()
					fmt.Fprintln(os.Stderr, "automatic publication pending:", metadataErr)
					continue
				}
				if title != "" {
					args = append(args, "--ticket-title", title)
				}
				if kind != "" {
					args = append(args, "--ticket-type", kind)
				}
				if policy.Summary {
					args = append(args, "--append", filepath.Join(filepath.Dir(dest), "SUMMARY.md"))
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
				cmd := exec.CommandContext(ctx, binary, args...)
				cmd.Dir = project
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				cancel()
				if err != nil {
					success = false
					lastFailure = err.Error()
					fmt.Fprintln(os.Stderr, "automatic publication pending:", err)
					if e := capture.WriteWorkerProgress(state, capture.WorkerProgress{PID: os.Getpid(), State: "failure", Heartbeat: time.Now().UnixMilli(), Deadline: started.Add(time.Duration(policy.MaxLifetimeMS) * time.Millisecond).UnixMilli(), NextAttempt: time.Now().Add(backoff * 2).UnixMilli(), Detail: err.Error()}); e != nil {
						return e
					}
				}
			}
			if success {
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
	// Failed publication cannot create an unbounded retry process chain. Durable
	// pending work restarts on the next supported lifecycle/reconcile integration.
	if !successfulPass || lastFailure != "" {
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
	return capture.LaunchWorker(binary, project, state)
}

func validateGenerationProject(project, state string, sessions []domain.SessionFact, ticket string) error {
	if len(sessions) == 0 {
		return nil
	}
	s, err := capture.OpenProject(state, project)
	if err != nil {
		return err
	}
	defer s.Close()
	rows, err := s.List(context.Background())
	if err != nil {
		return err
	}
	byID := map[string][]domain.CaptureRegistration{}
	for _, r := range rows {
		byID[r.Identity.SessionID] = append(byID[r.Identity.SessionID], r)
	}
	for _, owner := range sessions {
		var candidates []domain.CaptureRegistration
		for _, candidate := range byID[owner.ID] {
			if owner.Source == "" || candidate.Identity.Source == owner.Source {
				candidates = append(candidates, candidate)
			}
		}
		if len(candidates) > 1 {
			return fmt.Errorf("baseline owner %s has ambiguous source/store associations; explicit recovery required", owner.ID)
		}
		if len(candidates) == 0 {
			return fmt.Errorf("baseline owner %s lacks verified project/source association; explicit recovery required", owner.ID)
		}
		r := candidates[0]
		if r.TicketID != ticket {
			return fmt.Errorf("baseline owner %s is bound to ticket %s, refusing %s", owner.ID, r.TicketID, ticket)
		}
		if r.Identity.Source == domain.SourceCodex && r.TranscriptPath != "" {
			if err := codex.ValidateRolloutStore(r.Identity.Store, r.TranscriptPath); err != nil {
				return err
			}
		}

		if owner.Source != "" && r.Identity.Source != owner.Source {
			return errors.New("baseline source conflicts with project registration")
		}
		if r.ProjectDirectory != project {
			if err = capture.ValidateOwner(context.Background(), project, r); err != nil {
				return fmt.Errorf("baseline owner %s project unresolved: %w", owner.ID, err)
			}
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
	return capture.AutomaticEnabled(state)
}
