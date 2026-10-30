package main_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestCLIGenerate_SessionRefreshMatchesFreshExtraction(t *testing.T) {
	for _, source := range []domain.SourceHarness{domain.SourceCodex, domain.SourceOpenCode} {
		t.Run(string(source), func(t *testing.T) {
			project := t.TempDir()
			config := "project_name: RefreshFixture\nticket:\n  prefixes: [FIX]\nworkspaces:\n  dir: records\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
			if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			var sourceArgs []string
			var advance func(*testing.T, int)
			var activeCompletedAt string
			if source == domain.SourceCodex {
				home := t.TempDir()
				if err := os.MkdirAll(filepath.Join(home, "sessions"), 0700); err != nil {
					t.Fatal(err)
				}
				rollout := filepath.Join(home, "sessions", "rollout-thread-01.jsonl")
				chunks := []string{
					`{"timestamp":"2026-10-25T09:00:00Z","type":"session_meta","payload":{"id":"thread-01","session_id":"thread-01","timestamp":"2026-10-25T09:00:00Z"}}
{"timestamp":"2026-10-25T09:00:01Z","type":"turn_context","payload":{"model":"gpt-example","turn_id":"turn-01"}}
{"timestamp":"2026-10-25T09:00:02Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-01"}}
{"timestamp":"2026-10-25T09:00:03Z","type":"token_usage_record","payload":{"thread_id":"thread-01","turn_id":"turn-01","usage":{"input_tokens":40,"output_tokens":25}}}
`,
					`{"timestamp":"2026-10-25T09:00:07Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-01"}}
`,
					`{"timestamp":"2026-10-25T09:01:00Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-02"}}
{"timestamp":"2026-10-25T09:01:01Z","type":"token_usage_record","payload":{"thread_id":"thread-01","turn_id":"turn-02","usage":{"input_tokens":200,"output_tokens":50}}}
{"timestamp":"2026-10-25T09:01:02Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-02"}}
`,
				}
				advance = func(t *testing.T, stage int) {
					t.Helper()
					file, err := os.OpenFile(rollout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
					if err != nil {
						t.Fatal(err)
					}
					_, writeErr := file.WriteString(chunks[stage])
					closeErr := file.Close()
					if writeErr != nil || closeErr != nil {
						t.Fatalf("append rollout: write=%v close=%v", writeErr, closeErr)
					}
				}
				sourceArgs = []string{"--codex-home", home}
				activeCompletedAt = "2026-10-25T09:00:03Z"
			} else {
				dbPath := setupIsolatedEmptyDB(t)
				db, err := sql.Open("sqlite", dbPath)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := db.Close(); err != nil {
						t.Error(err)
					}
				})
				start := time.Date(2026, 10, 25, 9, 0, 0, 0, time.UTC).UnixMilli()
				if _, err := db.Exec(`INSERT INTO session (id, agent, model, time_created, time_updated) VALUES ('thread-01', 'root', 'gpt-example', ?, ?)`, start, start); err != nil {
					t.Fatal(err)
				}
				advance = func(t *testing.T, stage int) {
					t.Helper()
					var err error
					switch stage {
					case 0:
						data := fmt.Sprintf(`{"role":"assistant","modelID":"gpt-example","time":{"created":%d},"tokens":{"input":40,"output":25}}`, start+2000)
						_, err = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m1', 'thread-01', ?, ?, ?)`, start+2000, start+2000, data)
					case 1:
						data := fmt.Sprintf(`{"role":"assistant","modelID":"gpt-example","time":{"created":%d,"completed":%d},"tokens":{"input":40,"output":25}}`, start+2000, start+7000)
						_, err = db.Exec(`UPDATE message SET time_updated = ?, data = ? WHERE id = 'm1'`, start+7000, data)
					case 2:
						data := fmt.Sprintf(`{"role":"assistant","modelID":"gpt-example","time":{"created":%d,"completed":%d},"tokens":{"input":200,"output":50}}`, start+60000, start+62000)
						_, err = db.Exec(`INSERT INTO message (id, session_id, time_created, time_updated, data) VALUES ('m2', 'thread-01', ?, ?, ?)`, start+60000, start+62000, data)
					}
					if err != nil {
						t.Fatal(err)
					}
				}
				sourceArgs = []string{"--db-path", dbPath}
				activeCompletedAt = "2026-10-25T09:00:02Z"
			}
			generate := func(t *testing.T, output string) domain.AgenticFactLedger {
				t.Helper()
				args := []string{"generate", "--source", string(source), "--session-id", "thread-01",
					"--ticket-id", "FIX-10", "--ticket-title", "Snapshot refresh regression", "--ticket-type", "fix", "--json-output", output}
				cmd := exec.Command(cliBinaryPath, append(args, sourceArgs...)...)
				cmd.Dir = project
				if result, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("generate failed: %v\n%s", err, result)
				}
				data, err := os.ReadFile(output)
				if err != nil {
					t.Fatal(err)
				}
				var ledger domain.AgenticFactLedger
				if err := json.Unmarshal(data, &ledger); err != nil {
					t.Fatal(err)
				}
				return ledger
			}
			output := filepath.Join(project, "agentic_metrics.json")
			for stage, want := range []struct {
				name          string
				input         int64
				output        int64
				turns         int64
				activeSeconds int64
				completedAt   string
			}{
				{"active", 40, 25, 1, 0, activeCompletedAt},
				{"completed", 40, 25, 1, 5, "2026-10-25T09:00:07Z"},
				{"source_growth", 240, 75, 2, 7, "2026-10-25T09:01:02Z"},
			} {
				t.Run(want.name, func(t *testing.T) {
					advance(t, stage)
					refreshed := generate(t, output)
					fresh := generate(t, filepath.Join(project, fmt.Sprintf("fresh-%d.json", stage)))
					if !reflect.DeepEqual(refreshed.Sessions, fresh.Sessions) {
						t.Errorf("refresh differs from fresh extraction: refreshed=%+v fresh=%+v", refreshed.Sessions, fresh.Sessions)
					}
					if len(refreshed.Sessions) != 1 || refreshed.Sessions[0].ID != "thread-01" || refreshed.Sessions[0].Source != source {
						t.Fatalf("source ownership changed: %+v", refreshed.Sessions)
					}
					rows := refreshed.Sessions[0].Invocations
					if len(rows) != 1 {
						t.Fatalf("expected one authoritative aggregate, got %+v", rows)
					}
					inv := rows[0]
					if inv.Tokens.Input != want.input || inv.Tokens.Output != want.output || inv.Turns != want.turns || inv.ActiveDurationSeconds != want.activeSeconds {
						t.Errorf("incorrect aggregate: %+v", inv)
					}
					if inv.CompletedAt == nil {
						t.Fatal("missing aggregate completion timestamp")
					}
					completed, err := time.Parse(time.RFC3339Nano, *inv.CompletedAt)
					if err != nil || completed.UTC().Format(time.RFC3339) != want.completedAt {
						t.Errorf("expected completion %s, got %s (parse error: %v)", want.completedAt, *inv.CompletedAt, err)
					}
					if repeated := generate(t, output); !reflect.DeepEqual(repeated.Sessions, refreshed.Sessions) {
						t.Errorf("unchanged generation is not idempotent: before=%+v after=%+v", refreshed.Sessions, repeated.Sessions)
					}
				})
			}
		})
	}
}
