package capture

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestArchiveRefreshRejectsStaleRegistryObservation(t *testing.T) {
	// A source read can have completed before any of these newer writes. Use
	// independent store connections to order the actual registry transactions.
	for _, change := range []struct {
		name   string
		update func(*domain.CaptureRegistration)
	}{
		{"path", func(row *domain.CaptureRegistration) {
			row.TranscriptPath = filepath.Join(row.Identity.Store, "sessions", "newer.jsonl")
		}},
		{"binding", func(row *domain.CaptureRegistration) {
			row.Binding = "explicit_manual_selection"
			row.BindingAcceptedAt = "2026-01-02T00:00:00Z"
		}},
		{"checkpoint", func(row *domain.CaptureRegistration) {
			row.RefreshedAt = "2026-01-02T00:00:00Z"
			row.SourceFingerprint = "newer-source"
			row.Revision = 59
		}},
	} {
		t.Run(change.name, func(t *testing.T) {
			ctx := context.Background()
			cfg, err := adaptersconfig.LoadFromFile("../config/testdata/shared-policy.yaml")
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			open := func() *Store {
				t.Helper()
				s, err := Open(directory, cfg)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := s.Close(); err != nil {
						t.Error(err)
					}
				})
				return s
			}
			reader, writer := open(), open()
			identity := domain.CaptureIdentity{Source: domain.SourceCodex, Store: t.TempDir(), SessionID: "bound-owner"}
			observed, err := reader.Change(ctx, identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				row.TranscriptPath = filepath.Join(row.Identity.Store, "sessions", "owner.jsonl")
				row.TicketID = "TEST-100"
				row.Binding = "explicit_agent_response"
				row.BindingAcceptedAt = "2026-01-01T00:00:00Z"
				row.RefreshedAt = "2026-01-01T00:00:01Z"
				row.SourceFingerprint = "old-source"
				row.Revision = 31
				row.State = domain.CaptureCurrent
				return row, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			newer, err := writer.Change(ctx, observed.Identity, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				change.update(&row)
				return row, nil
			})
			if err != nil {
				t.Fatal(err)
			}
			staleApplyCalled := false
			_, err = reader.saveCaptureRead(ctx, observed, 0, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				staleApplyCalled = true
				row.TranscriptPath = filepath.Join(row.Identity.Store, "archived_sessions", "owner.jsonl")
				row.Revision = 41
				return row, nil
			})
			if staleApplyCalled || !errors.Is(err, errCaptureReadObservationChanged) {
				t.Fatalf("stale archive refresh reached the registry update: called=%v err=%v", staleApplyCalled, err)
			}
			rows, err := reader.List(ctx)
			if err != nil || len(rows) != 1 {
				t.Fatalf("reload current registry state: %d rows, %v", len(rows), err)
			}
			got := rows[0]
			if got.TranscriptPath != newer.TranscriptPath || got.Binding != newer.Binding ||
				got.BindingAcceptedAt != newer.BindingAcceptedAt || got.RefreshedAt != newer.RefreshedAt ||
				got.SourceFingerprint != newer.SourceFingerprint || got.Revision != newer.Revision {
				t.Fatalf("stale archive read overwrote a newer registration: %+v", got)
			}
		})
	}
}
