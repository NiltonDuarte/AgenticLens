package backfill_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"agenticlens/internal/adapters/backfill"
	"agenticlens/internal/domain"
)

type discoveryProbeOutcome struct {
	PermissionDenied bool
	Error            string
	Scanned          int
	Updated          int
	Tickets          []string
	Diagnostics      []string
}

// The subprocess uses the ordinary APIs against real fixture permissions. A
// root test process drops credentials so chmod denial is exercised rather than
// silently skipping this Linux permission boundary.
func TestWorkspaceDiscoveryPermissionProbe(t *testing.T) {
	if os.Getenv("AL_TEST_DISCOVERY_CHILD") != "1" {
		return
	}
	root := os.Getenv("AL_TEST_DISCOVERY_ROOT")
	blocked := os.Getenv("AL_TEST_DISCOVERY_BLOCKED")
	_, denied := os.ReadDir(blocked)
	outcome := discoveryProbeOutcome{PermissionDenied: os.IsPermission(denied)}
	if os.Getenv("AL_TEST_DISCOVERY_COMMAND") == "migration" {
		report, err := backfill.NewMigrationRunner(nil).Run(backfill.MigrationOptions{WorkspacesDir: filepath.Join(root, "records"), TicketID: os.Getenv("AL_TEST_DISCOVERY_TICKET"), TargetVersion: domain.CurrentSchemaVersion})
		if err != nil {
			outcome.Error = err.Error()
		}
		if report != nil {
			outcome.Scanned, outcome.Updated = report.Scanned, report.Migrated
			outcome.Diagnostics = report.Errors
			for _, result := range report.Results {
				outcome.Tickets = append(outcome.Tickets, result.TicketID)
			}
		}
	} else {
		report, err := backfill.NewEngine().Execute(backfill.BackfillOptions{WorkspacesDir: filepath.Join(root, "records"), TicketID: os.Getenv("AL_TEST_DISCOVERY_TICKET"), Source: "codex"})
		if err != nil {
			outcome.Error = err.Error()
		}
		if report != nil {
			outcome.Scanned, outcome.Updated = report.Scanned, report.Updated
			outcome.Diagnostics = report.Warnings
			for _, result := range report.Results {
				outcome.Tickets = append(outcome.Tickets, result.TicketID)
			}
		}
	}
	body, err := json.Marshal(outcome)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if _, err := os.Stdout.Write(body); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func runDiscoveryPermissionProbe(t *testing.T, root, blocked, command, ticket string) discoveryProbeOutcome {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	var credentials *syscall.Credential
	if os.Geteuid() == 0 {
		// The Go test binary's build directory may be private to root. Copy it
		// into this owned fixture and give the disposable project to the child.
		originalExecutable := executable
		executable = filepath.Join(root, "discovery-probe.test")
		source, err := os.Open(originalExecutable)
		if err != nil {
			t.Fatal(err)
		}
		target, err := os.OpenFile(executable, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			t.Fatal(errors.Join(err, source.Close()))
		}
		_, copyErr := io.Copy(target, source)
		if err := errors.Join(copyErr, source.Close(), target.Close()); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Dir(root), 0755); err != nil {
			t.Fatal(err)
		}
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			return os.Chown(path, 65534, 65534)
		}); err != nil {
			t.Fatal(err)
		}
		credentials = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, "-test.run=^TestWorkspaceDiscoveryPermissionProbe$")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "AL_TEST_DISCOVERY_CHILD=1", "AL_TEST_DISCOVERY_ROOT="+root, "AL_TEST_DISCOVERY_BLOCKED="+blocked, "AL_TEST_DISCOVERY_COMMAND="+command, "AL_TEST_DISCOVERY_TICKET="+ticket)
	if credentials != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{Credential: credentials}
	}
	body, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("real permission probe failed: %v: %s", err, body)
	}
	var outcome discoveryProbeOutcome
	if err := json.Unmarshal(body, &outcome); err != nil {
		t.Fatalf("permission probe output: %v: %s", err, body)
	}
	if !outcome.PermissionDenied {
		t.Fatal("filesystem control did not enforce permission denial; no inaccessible-workspace coverage was established")
	}
	return outcome
}

func TestWorkspaceDiscoveryIsolatesUnreadableChildren(t *testing.T) {
	for _, command := range []string{"migration", "backfill"} {
		for _, ticket := range []string{"GOOD-01", "", "ZZ-INACCESSIBLE"} {
			t.Run(command+"/"+ticket, func(t *testing.T) {
				root := migrationTestRoot(t)
				originals := make(map[string][]byte)
				for _, id := range []string{"GOOD-01", "GOOD-02", "ZZ-INACCESSIBLE"} {
					path := filepath.Join(root, "records", id, "agentic_metrics.json")
					if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
						t.Fatal(err)
					}
					body := []byte(`{"schema_version":6,"ticket_id":"` + id + `","updated_at":"recorded-time","sessions":[],"generation_policy":{"skills":true,"delegations":true,"anomalies":true}}`)
					if err := os.WriteFile(path, body, 0644); err != nil {
						t.Fatal(err)
					}
					originals[id] = body
				}
				blocked := filepath.Join(root, "records", "ZZ-INACCESSIBLE")
				if err := os.Chmod(blocked, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.Chmod(blocked, 0755); err != nil {
						t.Error(err)
					}
				})
				outcome := runDiscoveryPermissionProbe(t, root, blocked, command, ticket)
				want := 2
				if ticket == "GOOD-01" {
					want = 1
				} else if ticket == "ZZ-INACCESSIBLE" {
					want = 0
				}
				if outcome.Error != "" || outcome.Scanned != want || outcome.Updated != want || len(outcome.Tickets) != want {
					t.Fatalf("child failure blocked readable selected/sibling cleanup: %+v", outcome)
				}
				if len(outcome.Diagnostics) != 1 || !strings.Contains(outcome.Diagnostics[0], blocked) || !strings.Contains(outcome.Diagnostics[0], "incomplete") {
					t.Fatalf("uninspected coverage was concealed: %+v", outcome)
				}
				if err := os.Chmod(blocked, 0755); err != nil {
					t.Fatal(err)
				}
				for id, before := range originals {
					body, err := os.ReadFile(filepath.Join(root, "records", id, "agentic_metrics.json"))
					if err != nil {
						t.Fatal(err)
					}
					shouldChange := id != "ZZ-INACCESSIBLE" && (ticket == "" || ticket == id)
					if shouldChange {
						expected := bytes.Replace(before, []byte(`,"generation_policy":{"skills":true,"delegations":true,"anomalies":true}`), nil, 1)
						assertRawCleanupJSONEqual(t, expected, body)
					} else if !bytes.Equal(body, before) {
						t.Fatalf("unselected or inaccessible history changed: %s", id)
					}
				}
			})
		}
	}
}

func TestWorkspaceDiscoveryKeepsRootFailuresFatal(t *testing.T) {
	for _, command := range []string{"migration", "backfill"} {
		t.Run(command, func(t *testing.T) {
			root := migrationTestRoot(t)
			path := filepath.Join(root, "records", "GOOD-01", "agentic_metrics.json")
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			before := []byte(`{"schema_version":6,"ticket_id":"GOOD-01","sessions":[],"generation_policy":{"skills":true,"delegations":true,"anomalies":true}}`)
			if err := os.WriteFile(path, before, 0644); err != nil {
				t.Fatal(err)
			}
			blocked := filepath.Join(root, "records")
			if err := os.Chmod(blocked, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chmod(blocked, 0755); err != nil {
					t.Error(err)
				}
			})
			outcome := runDiscoveryPermissionProbe(t, root, blocked, command, "GOOD-01")
			if outcome.Error == "" || outcome.Updated != 0 || len(outcome.Tickets) != 0 {
				t.Fatalf("root failure was downgraded or another root selected: %+v", outcome)
			}
			if err := os.Chmod(blocked, 0755); err != nil {
				t.Fatal(err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("fatal root failure changed retained ledger: %s, %v", after, err)
			}
		})
	}
}
