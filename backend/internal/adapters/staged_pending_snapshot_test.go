package adapters

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestStagedPendingSnapshotBudgetsReportPartialCoverage(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		size  int
	}{
		{name: "candidate count", limit: stagedPendingFileLimit},
		{name: "cumulative bytes", limit: 2, size: stagedPendingByteLimit / 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			project := t.TempDir()
			run := func(args ...string) {
				t.Helper()
				cmd := exec.Command("git", append([]string{"-C", project}, args...)...)
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, output)
				}
			}
			run("init", "-qb", "main")
			const pending = "pending"
			if err := os.Mkdir(filepath.Join(project, pending), 0700); err != nil {
				t.Fatal(err)
			}

			var last string
			for n := 0; n <= test.limit; n++ {
				root := fmt.Sprintf("budget-root-%d", n)
				captureID := capture.StableCaptureID(domain.SourceCodex, root)
				body := fmt.Sprintf(`{
					"version":1,"capture_id":%q,"source":"codex","root_session_id":%q,
					"source_cutoff":"2026-01-01T00:00:00Z","state":"awaiting_ticket",
					"sessions":[{"id":%q,"source":"codex","user":"fixture","invocations":[],"tools":[]}]
				}`, captureID, root, root)
				if test.size > 0 {
					body += strings.Repeat(" ", test.size-len(body))
				}
				last = filepath.Join(pending, captureID+".json")
				if err := os.WriteFile(filepath.Join(project, last), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run("add", "--", pending)
			check := func(incomplete bool) {
				t.Helper()
				result := InspectStagedPendingSnapshots(context.Background(), project, []string{pending}, "")
				if result.Incomplete != incomplete || result.AwaitingTicket != test.limit || result.AwaitingDestination != 0 {
					t.Fatalf("expected %d classified snapshots, incomplete=%t; got %+v", test.limit, incomplete, result)
				}
			}
			check(true)
			run("rm", "--cached", "--", last)
			check(false)
		})
	}
}
