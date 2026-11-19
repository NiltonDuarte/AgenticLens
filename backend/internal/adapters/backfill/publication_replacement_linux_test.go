package backfill_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"golang.org/x/sys/unix"
)

func TestMaintenanceBatchRejectsWorkspaceReplacement(t *testing.T) {
	for _, command := range []string{"backfill", "migration"} {
		for _, replacementID := range []string{"AL-2", "AL-4"} {
			t.Run(command+"/"+replacementID, func(t *testing.T) {
				project := migrationTestRoot(t)
				first, selected, last := filepath.Join(project, "a-first"), filepath.Join(project, "b-selected"), filepath.Join(project, "c-last")
				moved := filepath.Join(project, "moved-original")
				firstBody := writeMaintenanceFixtureLedger(t, first, "AL-1")
				original := writeMaintenanceFixtureLedger(t, selected, "AL-2")
				lastBody := writeMaintenanceFixtureLedger(t, last, "AL-3")

				lease, err := capture.AcquirePublication(context.Background(), filepath.Join(first, "agentic_metrics.json"), project)
				if err != nil {
					t.Fatal(err)
				}
				held := true
				release := func() {
					if held {
						held = false
						if err := lease.Close(); err != nil {
							t.Error(err)
						}
					}
				}
				defer release()
				directoryHandles := publicationDirectoryHandles(t, first)
				done := make(chan map[string]error, 1)
				go func() { done <- runMaintenanceBatch(command, project) }()
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				poll := time.NewTicker(2 * time.Millisecond)
				defer poll.Stop()
				// Observe the next ownership descriptor, which uses NOFOLLOW
				// unlike discovery's directory reads. The batch is now waiting
				// on the real lease after freezing its target list.
				for publicationDirectoryHandles(t, first) <= directoryHandles {
					select {
					case result := <-done:
						t.Fatalf("batch finished before the first ownership wait: %v", result)
					case <-ctx.Done():
						t.Fatal("batch never reached the first ownership wait")
					case <-poll.C:
					}
				}
				if err := os.Rename(selected, moved); err != nil {
					t.Fatal(err)
				}
				replacement := writeMaintenanceFixtureLedger(t, selected, replacementID)
				release()
				select {
				case result := <-done:
					if len(result) != 3 {
						t.Fatalf("expected all three selected targets in the report, got %v", result)
					}
					if err := result[filepath.Join(selected, "agentic_metrics.json")]; err == nil || !strings.Contains(err.Error(), "retry") {
						t.Errorf("replacement requires a retry diagnostic, got %v", result)
					}
					for _, valid := range []string{first, last} {
						if err, exists := result[filepath.Join(valid, "agentic_metrics.json")]; !exists || err != nil {
							t.Errorf("valid sibling did not succeed: %s: %v", valid, result)
						}
					}
				case <-ctx.Done():
					t.Fatal("batch did not finish after releasing the first lease")
				}
				for path, before := range map[string][]byte{selected: replacement, moved: original} {
					if after := readMaintenanceFixtureLedger(t, path); !bytes.Equal(after, before) {
						t.Errorf("replacement or moved ledger changed: %s", path)
					}
				}
				for path, before := range map[string][]byte{first: firstBody, last: lastBody} {
					expected := bytes.Replace(before, []byte(`,"generation_policy":{"skills":false,"delegations":false,"anomalies":false}`), nil, 1)
					assertRawCleanupJSONEqual(t, expected, readMaintenanceFixtureLedger(t, path))
				}
			})
		}
	}
}

func runMaintenanceBatch(command, project string) map[string]error {
	result := make(map[string]error)
	if command == "migration" {
		report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{Project: project, TargetVersion: domain.CurrentSchemaVersion})
		if err != nil {
			return map[string]error{"batch": err}
		}
		for _, target := range report.Results {
			result[target.LedgerPath] = target.Error
		}
	} else {
		report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{Project: project})
		if err != nil {
			return map[string]error{"batch": err}
		}
		for _, target := range report.Results {
			result[target.LedgerPath] = target.Error
		}
	}
	return result
}

func writeMaintenanceFixtureLedger(t *testing.T, directory, ticket string) []byte {
	t.Helper()
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	body := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":%q,"sessions":[],"generation_policy":{"skills":false,"delegations":false,"anomalies":false}}`, ticket))
	if err := os.WriteFile(filepath.Join(directory, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	return body
}

func readMaintenanceFixtureLedger(t *testing.T, directory string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(directory, "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func publicationDirectoryHandles(t *testing.T, directory string) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	var count int
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err != nil || target != directory {
			continue
		}
		info, err := os.ReadFile(filepath.Join("/proc/self/fdinfo", entry.Name()))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(info), "\n") {
			flags, found := strings.CutPrefix(line, "flags:")
			if !found {
				continue
			}
			value, err := strconv.ParseUint(strings.TrimSpace(flags), 8, 32)
			if err != nil {
				t.Fatal(err)
			}
			if value&uint64(unix.O_DIRECTORY|unix.O_NOFOLLOW) == uint64(unix.O_DIRECTORY|unix.O_NOFOLLOW) {
				count++
			}
		}
	}
	return count
}
