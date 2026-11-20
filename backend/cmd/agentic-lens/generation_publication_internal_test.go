package main

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestCaptureAutomaticPublicationRejectsStaleCanonicalRevision(t *testing.T) {
	ctx := context.Background()
	project, state, native := t.TempDir(), t.TempDir(), t.TempDir()
	writeInternalCLIPublicationPolicy(t, project, ".")
	local := fmt.Sprintf("agenticlens:\n  state_dir: %q\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: %q\n    pass_timeout_ms: 2000\n    max_sessions: 32\n    max_source_bytes: 1048576\n", filepath.Join(project, "publication-state"), state)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	rollout := filepath.Join(native, "sessions", "root.jsonl")
	if err := os.Mkdir(filepath.Dir(rollout), 0700); err != nil {
		t.Fatal(err)
	}
	metadata := fmt.Sprintf(`{"type":"session_meta","payload":{"id":"root","cwd":%q}}`+"\n", project)
	if err := os.WriteFile(rollout, []byte(metadata), 0600); err != nil {
		t.Fatal(err)
	}
	store, err := capture.OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	const ticket = "TEST-100"
	identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: native, SessionID: "root"}
	row, err := store.Register(ctx, identity, rollout, "SessionEnd")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Bind(ctx, identity, ticket, row.Request); err != nil {
		t.Fatal(err)
	}
	observed, err := store.CaptureLifecycleRevision(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetCaptureDiscarded(ctx, capture.StableCaptureID(identity.Source, identity.SessionID), true, nil); err != nil {
		t.Fatal(err)
	}

	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		t.Fatal(err)
	}
	ledger, summary := filepath.Join(project, "agentic_metrics.json"), filepath.Join(project, "SUMMARY.md")
	prior := map[string]string{ledger: `{"ticket_id":"TEST-100"}`, summary: "# Authored notes\n"}
	for path, body := range prior {
		if err := capture.WritePublicationFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ticketLease, err := capture.AcquireInitialPublication(ctx, project, adapters.PublicationTicketKey(ticket, cfg.Ticket), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ticketLease.Close()
	lease, err := capture.AcquirePublication(ctx, ledger, project)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	called := false
	blocked, err := publishAutomaticCaptureArtifacts(project, state, ticket, ledger, cfg.Telemetry.Capture,
		capture.ReconcileResult{LifecycleRevision: observed}, nil, false, nil, func(func([]string) error) error {
			called = true
			return errors.New("obsolete publication callback ran")
		})
	if blocked || called || !errors.Is(err, capture.ErrCaptureLifecycleChanged) {
		t.Fatalf("obsolete publication crossed the lifecycle fence: blocked=%t called=%t err=%v", blocked, called, err)
	}
	for path, expected := range prior {
		if actual, err := os.ReadFile(path); err != nil || string(actual) != expected {
			t.Fatalf("stale publication changed prior bytes for %s: %q, %v", path, actual, err)
		}
	}
	if err := adapters.CheckPublication(ledger); err != nil {
		t.Fatalf("stale publication created transaction state: %v", err)
	}
}

func TestRecoverFailedGenerationPublicationReportsUnconfirmedRestoration(t *testing.T) {
	for _, committed := range []bool{false, true} {
		name := "rollback obstructed"
		if committed {
			name = "journal already committed"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			writeInternalCLIPublicationPolicy(t, dir, ".")
			ledger, summary := filepath.Join(dir, "ledger.json"), filepath.Join(dir, "SUMMARY.md")
			lock, err := capture.AcquirePublication(context.Background(), ledger, dir)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := lock.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := capture.WritePublicationFile(ledger, []byte("accepted ledger"), 0600); err != nil {
				t.Fatal(err)
			}
			paths := []string{ledger, summary}
			if err := adapters.BeginPublication(ledger, paths); err != nil {
				t.Fatal(err)
			}
			if err := capture.WritePublicationFile(ledger, []byte("replacement ledger"), 0600); err != nil {
				t.Fatal(err)
			}
			if committed {
				if err := capture.WritePublicationFile(summary, []byte("published summary"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := adapters.CommitPublication(ledger, paths); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(summary, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(summary, "obstruction"), []byte("preserve"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cause := errors.New("handled publication failure")
			err = recoverFailedGenerationPublication(ledger, paths, cause)
			if !errors.Is(err, cause) || strings.Contains(err.Error(), "restored prior accepted artifact set") {
				t.Fatalf("lost original cause or falsely claimed restoration: %v", err)
			}
			if committed {
				if !strings.Contains(err.Error(), "may already be committed") || !strings.Contains(err.Error(), "could not be confirmed") {
					t.Fatalf("committed-set ambiguity hidden: %v", err)
				}
				body, readErr := os.ReadFile(ledger)
				if readErr != nil || string(body) != "replacement ledger" {
					t.Fatalf("journal-free failure invented restoration: %q %v", body, readErr)
				}
			} else if !strings.Contains(err.Error(), "rollback incomplete") || adapters.CheckPublication(ledger) == nil {
				t.Fatalf("failed rollback hid pending recovery: %v", err)
			}
		})
	}
}
