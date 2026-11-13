package main

import (
	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecoverFailedGenerationPublicationReportsUnconfirmedRestoration(t *testing.T) {
	for _, committed := range []bool{false, true} {
		name := "rollback obstructed"
		if committed {
			name = "journal already committed"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			ledger, summary := filepath.Join(dir, "ledger.json"), filepath.Join(dir, "SUMMARY.md")
			lock, err := capture.AcquirePublication(context.Background(), ledger)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := lock.Close(); err != nil {
					t.Error(err)
				}
			}()
			if err := os.WriteFile(ledger, []byte("accepted ledger"), 0600); err != nil {
				t.Fatal(err)
			}
			paths := []string{ledger, summary}
			if err := adapters.BeginPublication(ledger, paths); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(ledger, []byte("replacement ledger"), 0600); err != nil {
				t.Fatal(err)
			}
			if committed {
				if err := os.WriteFile(summary, []byte("published summary"), 0600); err != nil {
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
