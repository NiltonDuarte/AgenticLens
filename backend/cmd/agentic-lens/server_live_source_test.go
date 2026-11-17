package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
)

func TestServerReportsSelectedLiveSourceAcrossRoutes(t *testing.T) {
	const aux = "CREATE TABLE session_message(id TEXT PRIMARY KEY, session_id TEXT, type TEXT, seq INTEGER, time_created INTEGER, time_updated INTEGER, data TEXT);"
	const v2 = aux + `
CREATE TABLE session_v2(id TEXT PRIMARY KEY, parent_id TEXT, agent TEXT, model TEXT, title TEXT, time_created INTEGER, time_updated INTEGER);
INSERT INTO session_v2 VALUES ('ses-orch',NULL,'orchestrator','model','Historical root',10,20), ('ses-coder','ses-orch','coder','model','Historical child',12,20);
INSERT INTO session_message VALUES ('msg','ses-orch','user',1,10,20,'{"time":{"created":10},"text":"historical v2 evidence"}');`

	cases := []struct {
		name     string
		setupSQL string
		code     string
		reason   string
	}{
		{"active-v1", `INSERT INTO session VALUES('active-owner', NULL, 'coder', 'model', 'Active fixture', unixepoch()*1000, unixepoch()*1000);
INSERT INTO message VALUES('active-message','active-owner',1,2,'{"role":"assistant","content":"current v1 evidence","tokens":{"input":220000,"output":5000}}');`, "", ""},
		{"legacy-v1-and-empty-auxiliary", "ALTER TABLE message DROP COLUMN time_updated; ALTER TABLE part DROP COLUMN time_updated;" + aux, "", ""},
		{"v2", v2, "source_unsupported", "v2"},
		{"mixed", v2, "source_unsupported", "mixed"},
		{"unknown", "CREATE TABLE unrelated(id TEXT);", "source_unsupported", "unknown"},
		{"incomplete", "CREATE TABLE session(id TEXT);", "source_unsupported", "malformed"},
		{"unset", "", "source_unavailable", "required"},
		{"missing", "", "source_unavailable", "no such file"},
		{"directory", "", "source_unavailable", "regular file"},
		{"corrupt", "", "source_unavailable", "not a database"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			records := setupTestWorkspace(t)
			configureExplicitPolicyWorkspace(t, records)
			project := filepath.Dir(records)
			source := filepath.Join(project, "source.db")
			switch tc.name {
			case "active-v1", "legacy-v1-and-empty-auxiliary", "mixed":
				source = setupIsolatedEmptyDB(t)
			case "directory":
				source = t.TempDir()
			case "corrupt":
				if err := os.WriteFile(source, []byte("invalid SQLite bytes"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if tc.setupSQL != "" {
				db, err := sql.Open("sqlite", source)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = db.Exec(tc.setupSQL); err != nil {
					t.Fatal(err)
				}
				if err = db.Close(); err != nil {
					t.Fatal(err)
				}
			}
			// A valid v1 path was written by setupTestWorkspace. Select the
			// alternative with --db-path, except for the configured-only v2 case.
			flags := []string{"--db-path", source}
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			switch tc.name {
			case "unset":
				if err := os.Remove(local); err != nil {
					t.Fatal(err)
				}
				flags = nil
			case "v2":
				body := fmt.Sprintf("telemetry:\n  sources:\n    opencode:\n      database_path: %q\n", source)
				if err := os.WriteFile(local, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
				flags = nil
			}

			ledger := filepath.Join(records, "TICKET-1", "agentic_metrics.json")
			if tc.name == "unset" {
				// Ship just the existing ledger to the project without source/private state.
				origin := setupTestWorkspace(t)
				portable := readCaptureFile(t, filepath.Join(origin, "TICKET-2", "agentic_metrics.json"))
				ledger = filepath.Join(records, "copied", "agentic_metrics.json")
				if err := os.Rename(filepath.Join(records, "TICKET-2"), filepath.Join(records, "copied")); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(ledger, portable, 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := readCaptureFile(t, ledger)
			address, log := startFixtureDashboard(t, project, flags...)
			client := &http.Client{Timeout: 3 * time.Second}
			startup := readCaptureFile(t, log)
			if active := bytes.Contains(startup, []byte("Live telemetry active")); active != (tc.reason == "") {
				t.Fatalf("unexpected source activation: %s", startup)
			}
			if tc.reason == "" {
				resp, err := client.Get(address + "/api/live/sessions")
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				var tree domain.LiveSessionTree
				if err = json.NewDecoder(resp.Body).Decode(&tree); err != nil || resp.StatusCode != 200 || tree.RootSessions == nil {
					t.Fatalf("failed v1 tree: status=%d tree=%+v err=%v", resp.StatusCode, tree, err)
				}
				if tc.name == "active-v1" {
					if tree.TotalActiveSessions != 1 || len(tree.RootSessions) != 1 || tree.RootSessions[0].ID != "active-owner" {
						t.Fatalf("lost active v1 owner: %+v", tree)
					}
					checkRESTTreeTraceFailure(t, client, address, source, tree)
					traceResp, err := client.Get(address + "/api/live/sessions/active-owner/trace")
					if err != nil {
						t.Fatal(err)
					}
					defer traceResp.Body.Close()
					var trace domain.LiveTraceView
					if err := json.NewDecoder(traceResp.Body).Decode(&trace); err != nil || traceResp.StatusCode != 200 || len(trace.Messages) != 1 || trace.Messages[0].Content != "current v1 evidence" {
						t.Fatalf("failed active v1 trace: %+v (%v)", trace, err)
					}
				} else if tree.TotalActiveSessions != 0 || len(tree.RootSessions) != 0 {
					t.Fatalf("idle v1 source should yield measured empty: %+v", tree)
				}
				return
			}

			var first adaptershttp.ErrorResponse
			for i, route := range []string{"sessions", "sessions/ses-orch/trace", "stream"} {
				resp, err := client.Get(address + "/api/live/" + route)
				if err != nil {
					t.Fatal(err)
				}
				var body adaptershttp.ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&body)
				resp.Body.Close()
				if err != nil || resp.StatusCode < 400 || body.StatusCode != resp.StatusCode || !strings.Contains(strings.ToLower(body.Message), tc.reason) {
					t.Fatalf("%s: status=%d body=%+v err=%v; startup=%s", route, resp.StatusCode, body, err, readCaptureFile(t, log))
				}
				if body.Error != tc.code || resp.StatusCode != http.StatusServiceUnavailable {
					t.Fatalf("%s: expected 503/%s, got %+v", route, tc.code, body)
				}
				if i == 0 {
					first = body
					if !bytes.Contains(readCaptureFile(t, log), []byte(body.Message)) {
						t.Fatalf("startup and REST disagree: %s / %+v", readCaptureFile(t, log), body)
					}
				} else if body != first {
					t.Fatalf("%s: %+v, sessions: %+v", route, body, first)
				}
			}

			ticket := "TICKET-1"
			if tc.name == "unset" {
				ticket = "TICKET-2"
				resp, err := client.Get(address + "/api/workspaces/" + ticket)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				var detail adaptershttp.WorkspaceDetailResponse
				if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil || resp.StatusCode != 200 {
					t.Fatalf("portable detail: %d, %v", resp.StatusCode, err)
				}
				sessions := detail.Ledger.Facts.Sessions
				if len(sessions) != 2 || sessions[0].ID != "ses-orch" || sessions[0].User != "test-engineer" || len(sessions[0].Tools) != 1 || sessions[0].Tools[0].Calls != 10 || sessions[1].ParentID == nil || *sessions[1].ParentID != sessions[0].ID {
					t.Fatalf("copied ledger lost recorded facts: %+v", sessions)
				}
			}
			resp, err := client.Get(address + "/api/workspaces/" + ticket + "/events")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			var history domain.WorkspaceEventsResponse
			if err = json.NewDecoder(resp.Body).Decode(&history); err != nil || resp.StatusCode != 200 {
				t.Fatalf("historical events: %v (HTTP %d)", err, resp.StatusCode)
			}
			if tc.name == "v2" {
				if history.Availability != domain.TraceComplete || !slices.ContainsFunc(history.Events, func(event domain.EventFact) bool {
					return event.SessionID == "ses-orch" && event.PromptContent != nil && *event.PromptContent == "historical v2 evidence"
				}) {
					t.Fatalf("clean v2 exact-owner hydration: %+v", history)
				}
			} else if history.Availability != domain.TraceUnavailable || len(history.Events) != 0 {
				t.Fatalf("unexpected source events: %+v", history)
			}
			if tc.name == "mixed" && !strings.Contains(strings.Join(history.Coverage.Diagnostics, " "), "mixed") {
				t.Fatalf("missing mixed-layout exact-owner diagnosis: %+v", history)
			}
			if !bytes.Equal(before, readCaptureFile(t, ledger)) {
				t.Fatal("GET changed recorded ledger")
			}
		})
	}
}

// The REST fallback reads only the tree when no session is selected. A trace
// lookup failure during that read must not replace observed alerts with zero.
func checkRESTTreeTraceFailure(t *testing.T, client *http.Client, address, source string, before domain.LiveSessionTree) {
	t.Helper()
	if len(before.ActiveAnomalies) == 0 || len(before.RootSessions[0].Anomalies) == 0 {
		t.Fatal("fixture must expose an observed per-session anomaly")
	}
	writer, err := sql.Open("sqlite", source)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	if _, err := writer.Exec("ALTER TABLE message RENAME TO unreadable_message"); err != nil {
		t.Fatal(err)
	}
	resp, err := client.Get(address + "/api/live/sessions")
	if err != nil {
		t.Fatal(err)
	}
	var failure adaptershttp.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&failure)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusInternalServerError || failure.Error != "database_error" || !strings.Contains(failure.Message, "message") || !strings.Contains(failure.Message, "active-owner") {
		t.Fatalf("failed tree enrichment returned HTTP %d: %+v (%v)", resp.StatusCode, failure, err)
	}
	if _, err := writer.Exec("ALTER TABLE unreadable_message RENAME TO message"); err != nil {
		t.Fatal(err)
	}
	recovered, err := client.Get(address + "/api/live/sessions")
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Body.Close()
	var tree domain.LiveSessionTree
	err = json.NewDecoder(recovered.Body).Decode(&tree)
	if err != nil || recovered.StatusCode != http.StatusOK || len(tree.RootSessions) != 1 || !reflect.DeepEqual(tree.ActiveAnomalies, before.ActiveAnomalies) || !reflect.DeepEqual(tree.RootSessions[0].Anomalies, before.RootSessions[0].Anomalies) {
		t.Fatalf("unchanged source lost measured anomalies: %+v (%v)", tree, err)
	}
}
