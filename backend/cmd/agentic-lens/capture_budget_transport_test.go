package main_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestNativeWorkerPublicationKeepsExplicitBudgetOverrides(t *testing.T) {
	for _, mode := range []string{"missing-local-budgets", "different-local-budgets"} {
		t.Run(mode, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 37)
			f.bind(t, path)
			writeCaptureSuccessfulPublicationPolicy(t, f.project, f.state)
			localPath := filepath.Join(f.project, ".agenticlens-config.local.yaml")
			local := string(readCaptureFile(t, localPath))
			local = strings.ReplaceAll(local, "    max_sessions: 256\n", "")
			local = strings.ReplaceAll(local, "    max_source_bytes: 67108864\n", "")
			if mode == "different-local-budgets" {
				local = strings.Replace(local, "  capture:\n", "  capture:\n    max_sessions: 1\n    max_source_bytes: 1\n", 1)
			}
			if err := os.WriteFile(localPath, []byte(local), 0600); err != nil {
				t.Fatal(err)
			}
			registry, err := capture.OpenProject(f.state, f.project)
			if err != nil {
				t.Fatal(err)
			}
			defer registry.Close()
			id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: f.store, SessionID: "root"}
			if _, err := registry.Change(context.Background(), id, func(row domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				row.Activity = "idle"
				row.TicketTitle = "Explicit budget publication"
				row.TicketType = "feat"
				return row, nil
			}); err != nil {
				t.Fatal(err)
			}
			if err := registry.RequestPublication(context.Background()); err != nil {
				t.Fatal(err)
			}
			output, err := runCaptureNative(t, f.project, "", "capture", "worker", "--repo", f.project, "--state-dir", f.state, "--max-sessions", "16", "--max-source-bytes", "1048576")
			if err != nil {
				t.Fatalf("effective budgets lost across native child: %v %s", err, output)
			}
			ledgerPath := filepath.Join(f.project, "records", "TEST-100", "agentic_metrics.json")
			var ledger domain.AgenticFactLedger
			if err := json.Unmarshal(readCaptureFile(t, ledgerPath), &ledger); err != nil {
				t.Fatal(err)
			}
			assertCaptureRootInput(t, ledger, 37)
			req, err := registry.PublicationRequest(context.Background())
			if err != nil || req.Generation != req.Acknowledged {
				t.Fatalf("successful worker generation not acknowledged: %+v %v", req, err)
			}
			if got := string(readCaptureFile(t, localPath)); got != local {
				t.Fatal("worker changed local configuration")
			}
		})
	}
}

func TestGenerateRejectsInvalidExplicitCaptureBudgets(t *testing.T) {
	for _, flag := range []string{"--max-sessions", "--max-source-bytes"} {
		for _, value := range []string{"0", "-1"} {
			t.Run(flag+value, func(t *testing.T) {
				f := newCaptureGenerationFixture(t)
				f.write(t, 37)
				args := f.args(t, flag, value)
				output, err := runCaptureNative(t, f.project, "", args...)
				if err == nil || !strings.Contains(string(output), flag+" must be a positive integer") {
					t.Fatalf("invalid explicit budget fell through to configured value: %v %s", err, output)
				}
				if _, err := os.Stat(f.output); !os.IsNotExist(err) {
					t.Fatalf("invalid explicit budget wrote output: %v", err)
				}
			})
		}
	}
}
