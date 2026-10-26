package http_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"agenticlens/internal/adapters/config"
	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"

	_ "modernc.org/sqlite"
)

func mustNewServer(t *testing.T, workspacesDir string, staticFS fs.FS, liveService ...adaptershttp.LiveServiceInterface) *adaptershttp.Server {
	t.Helper()
	server, err := adaptershttp.NewServer(workspacesDir, staticFS, liveService...)
	if err != nil {
		t.Fatalf("mustNewServer failed: %v", err)
	}
	return server
}

func sampleTestLedgerJSON(ticketID, title string, duration int64, cost float64, invocations int64) string {
	return fmt.Sprintf(`{"schema_version":4,"ticket_id":%q,"ticket_title":%q,"ticket_type":"feature","started_at":"2026-10-17T10:00:00Z","completed_at":"2026-10-17T11:00:00Z","updated_at":"2026-10-17T11:00:00Z","code":{"lines_added":120,"lines_deleted":20,"test_lines_added":60,"test_lines_deleted":0,"prod_lines_added":60,"prod_lines_deleted":20,"files_changed":4,"sdlc_iterations":2},"sessions":[{"id":"ses-orch","invocations":[{"id":"ses-orch","role":"orchestrator","model":"claude-3-sonnet","user":"tester","active_duration_seconds":%d,"turns":%d,"tokens":{"input":5000,"output":1000,"cache_read":10000,"reasoning":300,"peak_context":80000},"cost_usd":%f,"compactions":0}],"tools":[{"name":"bash","calls":5,"errors":0}],"events":[]}]}`, ticketID, title, duration, invocations, cost)
}

func setupFixtureWorkspaces(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed creating workspaces dir: %v", err)
	}

	t1Dir := filepath.Join(wsDir, "AL-1")
	if err := os.MkdirAll(t1Dir, 0755); err != nil {
		t.Fatalf("failed creating AL-1 dir: %v", err)
	}
	t1JSON := `{
  "ticket_id": "AL-1",
  "ticket_title": "First Ticket",
  "ticket_type": "feature",
  "started_at": "2026-10-15T10:00:00Z",
  "completed_at": "2026-10-15T11:00:00Z",
  "updated_at": "2026-10-15T11:00:00Z",
  "code": {
    "lines_added": 50,
    "lines_deleted": 10,
    "test_lines_added": 20,
    "test_lines_deleted": 0,
    "prod_lines_added": 30,
    "prod_lines_deleted": 10,
    "files_changed": 2,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "ses-1",
      "invocations": [
        {
          "id": "ses-1",
          "role": "orchestrator",
          "model": "gpt-4",
          "user": "dev",
          "active_duration_seconds": 60,
          "turns": 2,
          "tokens": {
            "input": 1000,
            "output": 200,
            "cache_read": 500,
            "reasoning": 0,
            "peak_context": 50000
          },
          "cost_usd": 0.1,
          "compactions": 0
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(t1Dir, "agentic_metrics.json"), []byte(t1JSON), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(t1Dir, "CONTRACT.md"), []byte("# AL-1 Contract Content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(t1Dir, "SUMMARY.md"), []byte("# AL-1 Summary Content"), 0644); err != nil {
		t.Fatal(err)
	}

	t2Dir := filepath.Join(wsDir, "AL-2")
	if err := os.MkdirAll(t2Dir, 0755); err != nil {
		t.Fatalf("failed creating AL-2 dir: %v", err)
	}
	t2JSON := `{
  "ticket_id": "AL-2",
  "ticket_title": "Second Ticket",
  "ticket_type": "feature",
  "started_at": "2026-10-17T10:00:00Z",
  "completed_at": "2026-10-17T11:00:00Z",
  "updated_at": "2026-10-17T11:00:00Z",
  "code": {
    "lines_added": 100,
    "lines_deleted": 20,
    "test_lines_added": 50,
    "test_lines_deleted": 0,
    "prod_lines_added": 50,
    "prod_lines_deleted": 20,
    "files_changed": 3,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "ses-2",
      "invocations": [
        {
          "id": "ses-2",
          "role": "orchestrator",
          "model": "gpt-4",
          "user": "dev",
          "active_duration_seconds": 120,
          "turns": 4,
          "tokens": {
            "input": 2000,
            "output": 400,
            "cache_read": 1000,
            "reasoning": 100,
            "peak_context": 60000
          },
          "cost_usd": 0.5,
          "compactions": 0
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(t2Dir, "agentic_metrics.json"), []byte(t2JSON), 0644); err != nil {
		t.Fatal(err)
	}

	return wsDir
}

func setupTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	wsDir := setupFixtureWorkspaces(t)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><head><title>AgenticLens</title></head><body>Root</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('loaded');"),
		},
	}

	server := mustNewServer(t, wsDir, mockFS)
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(func() {
		ts.Close()
	})

	return ts, wsDir
}

func TestGetWorkspaces(t *testing.T) {
	ts, _ := setupTestServer(t)

	t.Run("list all workspaces", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces")
		if err != nil {
			t.Fatalf("failed GET /api/workspaces: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if len(ledgers) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(ledgers))
		}
	})

	t.Run("sort workspaces by cost descending", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces?sort=cost&order=desc")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if len(ledgers) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(ledgers))
		}

		if *ledgers[0].TicketID != "AL-2" || *ledgers[1].TicketID != "AL-1" {
			t.Errorf("expected AL-2 then AL-1, got %s then %s", *ledgers[0].TicketID, *ledgers[1].TicketID)
		}
	})

	t.Run("sort workspaces by ticket_id ascending and descending", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}
		for _, tid := range []string{"AL-10", "AL-2", "AL-1"} {
			dir := filepath.Join(wsDir, tid)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			data := `{"ticket_id":"` + tid + `","updated_at":"2026-10-18T10:00:00Z"}`
			if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), []byte(data), 0644); err != nil {
				t.Fatal(err)
			}
		}

		mockFS := fstest.MapFS{
			"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
		}
		server := mustNewServer(t, wsDir, mockFS)
		srv := httptest.NewServer(server.Handler())
		defer srv.Close()

		// Ascending
		respAsc, err := http.Get(srv.URL + "/api/workspaces?sort=ticket_id&order=asc")
		if err != nil {
			t.Fatalf("failed GET asc: %v", err)
		}
		defer respAsc.Body.Close()
		if respAsc.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", respAsc.StatusCode)
		}
		var ledgersAsc []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respAsc.Body).Decode(&ledgersAsc); err != nil {
			t.Fatalf("failed decode asc: %v", err)
		}
		if len(ledgersAsc) != 3 {
			t.Fatalf("expected 3 ledgers, got %d", len(ledgersAsc))
		}
		if *ledgersAsc[0].TicketID != "AL-1" || *ledgersAsc[1].TicketID != "AL-2" || *ledgersAsc[2].TicketID != "AL-10" {
			t.Errorf("expected AL-1, AL-2, AL-10; got %s, %s, %s",
				*ledgersAsc[0].TicketID, *ledgersAsc[1].TicketID, *ledgersAsc[2].TicketID)
		}

		// Descending
		respDesc, err := http.Get(srv.URL + "/api/workspaces?sort=ticket_id&order=desc")
		if err != nil {
			t.Fatalf("failed GET desc: %v", err)
		}
		defer respDesc.Body.Close()
		if respDesc.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", respDesc.StatusCode)
		}
		var ledgersDesc []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respDesc.Body).Decode(&ledgersDesc); err != nil {
			t.Fatalf("failed decode desc: %v", err)
		}
		if len(ledgersDesc) != 3 {
			t.Fatalf("expected 3 ledgers, got %d", len(ledgersDesc))
		}
		if *ledgersDesc[0].TicketID != "AL-10" || *ledgersDesc[1].TicketID != "AL-2" || *ledgersDesc[2].TicketID != "AL-1" {
			t.Errorf("expected AL-10, AL-2, AL-1; got %s, %s, %s",
				*ledgersDesc[0].TicketID, *ledgersDesc[1].TicketID, *ledgersDesc[2].TicketID)
		}

		// Default sort (no params): naturally ordered ascending by default ScanWorkspaces
		respDefault, err := http.Get(srv.URL + "/api/workspaces")
		if err != nil {
			t.Fatalf("failed GET default: %v", err)
		}
		defer respDefault.Body.Close()
		var ledgersDefault []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respDefault.Body).Decode(&ledgersDefault); err != nil {
			t.Fatalf("failed decode default: %v", err)
		}
		if len(ledgersDefault) != 3 || *ledgersDefault[0].TicketID != "AL-1" || *ledgersDefault[1].TicketID != "AL-2" || *ledgersDefault[2].TicketID != "AL-10" {
			t.Errorf("expected default natural order AL-1, AL-2, AL-10; got %v", ledgersDefault)
		}

		// sort=ticket_id without order param: defaults to ascending
		respNoOrder, err := http.Get(srv.URL + "/api/workspaces?sort=ticket_id")
		if err != nil {
			t.Fatalf("failed GET sort=ticket_id: %v", err)
		}
		defer respNoOrder.Body.Close()
		var ledgersNoOrder []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respNoOrder.Body).Decode(&ledgersNoOrder); err != nil {
			t.Fatalf("failed decode sort=ticket_id: %v", err)
		}
		if len(ledgersNoOrder) != 3 || *ledgersNoOrder[0].TicketID != "AL-1" || *ledgersNoOrder[1].TicketID != "AL-2" || *ledgersNoOrder[2].TicketID != "AL-10" {
			t.Errorf("expected sort=ticket_id without order to be asc AL-1, AL-2, AL-10; got %v", ledgersNoOrder)
		}

		// sort with unknown parameter: defaults safely to updated_at without panicking
		respUnknown, err := http.Get(srv.URL + "/api/workspaces?sort=nonexistent_field")
		if err != nil {
			t.Fatalf("failed GET unknown sort: %v", err)
		}
		defer respUnknown.Body.Close()
		if respUnknown.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for unknown sort field, got %d", respUnknown.StatusCode)
		}
	})

	t.Run("sort workspaces by completed_at", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces?sort=completed_at&order=desc")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decode: %v", err)
		}
		if len(ledgers) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(ledgers))
		}
	})

	t.Run("temporal boundary filtering with valid dates", func(t *testing.T) {
		// Both tickets AL-1 (2026-10-15) and AL-2 (2026-10-17) fall within 2026-10-15 to 2026-10-17
		resp, err := http.Get(ts.URL + "/api/workspaces?since=2026-10-15&until=2026-10-17")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decode: %v", err)
		}
		if len(ledgers) != 2 {
			t.Fatalf("expected 2 ledgers between 2026-10-15 and 2026-10-17, got %d", len(ledgers))
		}

		// Single day 2026-10-17 matches AL-2 only
		respSingle, err := http.Get(ts.URL + "/api/workspaces?since=2026-10-17&until=2026-10-17")
		if err != nil {
			t.Fatalf("failed GET single day: %v", err)
		}
		defer respSingle.Body.Close()
		var ledgersSingle []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respSingle.Body).Decode(&ledgersSingle); err != nil {
			t.Fatalf("failed decode single day: %v", err)
		}
		if len(ledgersSingle) != 1 || *ledgersSingle[0].TicketID != "AL-2" {
			t.Fatalf("expected 1 ledger (AL-2) on 2026-10-17, got %+v", ledgersSingle)
		}

		// Filter beyond all tickets
		respFuture, err := http.Get(ts.URL + "/api/workspaces?since=2027-01-01")
		if err != nil {
			t.Fatalf("failed GET future: %v", err)
		}
		defer respFuture.Body.Close()
		var emptyLedgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respFuture.Body).Decode(&emptyLedgers); err != nil {
			t.Fatalf("failed decode future: %v", err)
		}
		if len(emptyLedgers) != 0 {
			t.Fatalf("expected 0 ledgers, got %d", len(emptyLedgers))
		}
	})

	t.Run("invalid since or until returns 400 Bad Request", func(t *testing.T) {
		respSince, err := http.Get(ts.URL + "/api/workspaces?since=not-a-date")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer respSince.Body.Close()
		if respSince.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for bad since, got %d", respSince.StatusCode)
		}

		respUntil, err := http.Get(ts.URL + "/api/workspaces?until=not-a-date")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer respUntil.Body.Close()
		if respUntil.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for bad until, got %d", respUntil.StatusCode)
		}

		respFleetSince, err := http.Get(ts.URL + "/api/fleet?since=invalid-date")
		if err != nil {
			t.Fatalf("failed GET fleet bad since: %v", err)
		}
		defer respFleetSince.Body.Close()
		if respFleetSince.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for fleet bad since, got %d", respFleetSince.StatusCode)
		}

		respFleetUntil, err := http.Get(ts.URL + "/api/fleet?until=invalid-date")
		if err != nil {
			t.Fatalf("failed GET fleet bad until: %v", err)
		}
		defer respFleetUntil.Body.Close()
		if respFleetUntil.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for fleet bad until, got %d", respFleetUntil.StatusCode)
		}
	})

	t.Run("inverted date ranges return empty results without panicking", func(t *testing.T) {
		respWS, err := http.Get(ts.URL + "/api/workspaces?since=2026-10-25&until=2026-10-10")
		if err != nil {
			t.Fatalf("failed GET workspaces inverted: %v", err)
		}
		defer respWS.Body.Close()
		if respWS.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK for inverted date range, got %d", respWS.StatusCode)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respWS.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decode: %v", err)
		}
		if len(ledgers) != 0 {
			t.Errorf("expected 0 ledgers for inverted date range, got %d", len(ledgers))
		}

		respFleet, err := http.Get(ts.URL + "/api/fleet?since=2026-10-25&until=2026-10-10")
		if err != nil {
			t.Fatalf("failed GET fleet inverted: %v", err)
		}
		defer respFleet.Body.Close()
		if respFleet.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK for fleet inverted, got %d", respFleet.StatusCode)
		}
		var fleet domain.FleetSummary
		if err := json.NewDecoder(respFleet.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decode fleet: %v", err)
		}
		if fleet.TotalTickets != 0 || fleet.TotalCostUSD != 0.0 {
			t.Errorf("expected zeroed fleet summary for inverted date range, got %+v", fleet)
		}
	})

	t.Run("timezone offsets and window presets in query parameters", func(t *testing.T) {
		// Timezone offset query parameters: AL-2 completed at 2026-10-17T11:00:00Z
		// since 2026-10-17T19:00:00+09:00 == 2026-10-17T10:00:00Z (before completion)
		// until 2026-10-17T08:00:00-04:00 == 2026-10-17T12:00:00Z (after completion)
		respTZ, err := http.Get(ts.URL + "/api/workspaces?since=2026-10-17T19:00:00%2B09:00&until=2026-10-17T08:00:00-04:00")
		if err != nil {
			t.Fatalf("failed GET with timezone offsets: %v", err)
		}
		defer respTZ.Body.Close()
		if respTZ.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", respTZ.StatusCode)
		}
		var ledgersTZ []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respTZ.Body).Decode(&ledgersTZ); err != nil {
			t.Fatalf("failed decode: %v", err)
		}
		if len(ledgersTZ) != 1 || *ledgersTZ[0].TicketID != "AL-2" {
			t.Fatalf("expected exactly AL-2 matching timezone window, got %+v", ledgersTZ)
		}

		// Window presets on /api/fleet
		for _, w := range []string{"7d", "30d", "90d", "all"} {
			respWin, err := http.Get(ts.URL + "/api/fleet?window=" + w)
			if err != nil {
				t.Fatalf("failed GET fleet window %s: %v", w, err)
			}
			respWin.Body.Close()
			if respWin.StatusCode != http.StatusOK {
				t.Errorf("expected 200 for window %s, got %d", w, respWin.StatusCode)
			}
		}
	})
}

func TestGetWorkspaceDetail(t *testing.T) {
	ts, _ := setupTestServer(t)

	t.Run("existing ticket returns detail with artifacts", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-1")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			t.Fatalf("failed decoding detail: %v", err)
		}

		if detail.Ledger.TicketID == nil || *detail.Ledger.TicketID != "AL-1" {
			t.Errorf("expected ticket ID AL-1")
		}
		if detail.Artifacts.ContractContent == nil || *detail.Artifacts.ContractContent != "# AL-1 Contract Content" {
			t.Errorf("expected contract content matching fixture")
		}
		if detail.Artifacts.SummaryContent == nil || *detail.Artifacts.SummaryContent != "# AL-1 Summary Content" {
			t.Errorf("expected summary content matching fixture")
		}
	})

	t.Run("dynamic custom workspace artifacts configuration in detail endpoint", func(t *testing.T) {
		tempDir := t.TempDir()
		wsDir := filepath.Join(tempDir, "workspaces")
		tDir := filepath.Join(wsDir, "AL-CUSTOM")
		if err := os.MkdirAll(tDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tDir, "agentic_metrics.json"), []byte(sampleTestLedgerJSON("AL-CUSTOM", "Custom Ticket", 100, 1.0, 2)), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tDir, "DESIGN.md"), []byte("# Architectural Design Doc"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tDir, "AUDIT.md"), []byte("# Security Audit Log"), 0644); err != nil {
			t.Fatal(err)
		}

		customCfg := domain.ProjectConfig{
			ProjectName: "CustomProject",
			Workspaces: domain.WorkspacesConfig{
				Dir: wsDir,
				Artifacts: []domain.WorkspaceArtifactConfig{
					{
						Path:        "DESIGN.md",
						Title:       "Design Specification",
						Description: "High-level architecture design doc",
					},
					{
						Path:        "AUDIT.md",
						Title:       "Security Audit",
						Description: "Threat model and audit review",
					},
					{
						Path:        "MISSING.md",
						Title:       "Missing Report",
						Description: "Report not yet created",
					},
				},
			},
		}

		customSrv := adaptershttp.NewServerWithConfig(wsDir, fstest.MapFS{}, customCfg)
		customTS := httptest.NewServer(customSrv.Handler())
		defer customTS.Close()

		resp, err := http.Get(customTS.URL + "/api/workspaces/AL-CUSTOM")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			t.Fatalf("failed decoding detail: %v", err)
		}

		if len(detail.Artifacts.Items) != 3 {
			t.Fatalf("expected 3 artifact items, got %d", len(detail.Artifacts.Items))
		}

		// Item 0: DESIGN.md
		if detail.Artifacts.Items[0].Path != "DESIGN.md" || detail.Artifacts.Items[0].Title != "Design Specification" {
			t.Errorf("item 0 mismatch: %+v", detail.Artifacts.Items[0])
		}
		if detail.Artifacts.Items[0].Content == nil || *detail.Artifacts.Items[0].Content != "# Architectural Design Doc" {
			t.Errorf("expected DESIGN.md content, got %v", detail.Artifacts.Items[0].Content)
		}

		// Item 1: AUDIT.md
		if detail.Artifacts.Items[1].Path != "AUDIT.md" || detail.Artifacts.Items[1].Title != "Security Audit" {
			t.Errorf("item 1 mismatch: %+v", detail.Artifacts.Items[1])
		}
		if detail.Artifacts.Items[1].Content == nil || *detail.Artifacts.Items[1].Content != "# Security Audit Log" {
			t.Errorf("expected AUDIT.md content, got %v", detail.Artifacts.Items[1].Content)
		}

		// Item 2: MISSING.md
		if detail.Artifacts.Items[2].Path != "MISSING.md" || detail.Artifacts.Items[2].Title != "Missing Report" {
			t.Errorf("item 2 mismatch: %+v", detail.Artifacts.Items[2])
		}
		if detail.Artifacts.Items[2].Content != nil {
			t.Errorf("expected nil content for missing file, got %v", detail.Artifacts.Items[2].Content)
		}
	})

	t.Run("non-existent ticket returns 404 error response", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/DOES-NOT-EXIST")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}

		var errResp adaptershttp.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed decoding error response: %v", err)
		}

		if errResp.Error != "not_found" {
			t.Errorf("expected error not_found, got %s", errResp.Error)
		}
		if errResp.StatusCode != 404 {
			t.Errorf("expected status_code 404, got %d", errResp.StatusCode)
		}
	})
}

func TestGetFleet(t *testing.T) {
	ts, _ := setupTestServer(t)

	t.Run("default fleet summary without params", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/fleet")
		if err != nil {
			t.Fatalf("failed GET /api/fleet: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var fleet domain.FleetSummary
		if err := json.NewDecoder(resp.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}

		if fleet.TotalTickets != 2 {
			t.Errorf("expected 2 total tickets, got %d", fleet.TotalTickets)
		}
		if fleet.TotalCostUSD != 0.60 {
			t.Errorf("expected total cost 0.60, got %f", fleet.TotalCostUSD)
		}
	})

	t.Run("window preset parameter accepts 7d, 30d, 90d, all", func(t *testing.T) {
		for _, w := range []string{"7d", "30d", "90d", "all"} {
			resp, err := http.Get(ts.URL + "/api/fleet?window=" + w)
			if err != nil {
				t.Fatalf("failed GET /api/fleet?window=%s: %v", w, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 for window %s, got %d", w, resp.StatusCode)
			}
		}
	})

	t.Run("temporal since and until filtering", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/fleet?since=2026-10-15&until=2026-10-17")
		if err != nil {
			t.Fatalf("failed GET /api/fleet: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		var fleet domain.FleetSummary
		if err := json.NewDecoder(resp.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}
		if fleet.TotalTickets != 2 {
			t.Errorf("expected 2 tickets, got %d", fleet.TotalTickets)
		}

		// Single day 2026-10-17 matches AL-2 only
		respSingle, err := http.Get(ts.URL + "/api/fleet?since=2026-10-17&until=2026-10-17")
		if err != nil {
			t.Fatalf("failed GET single day: %v", err)
		}
		defer respSingle.Body.Close()
		var fleetSingle domain.FleetSummary
		if err := json.NewDecoder(respSingle.Body).Decode(&fleetSingle); err != nil {
			t.Fatalf("failed decoding single fleet: %v", err)
		}
		if fleetSingle.TotalTickets != 1 {
			t.Errorf("expected 1 ticket on 2026-10-17, got %d", fleetSingle.TotalTickets)
		}

		// Future since filter returns 0 tickets
		respFuture, err := http.Get(ts.URL + "/api/fleet?since=2027-01-01")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer respFuture.Body.Close()
		var emptyFleet domain.FleetSummary
		if err := json.NewDecoder(respFuture.Body).Decode(&emptyFleet); err != nil {
			t.Fatalf("failed decoding empty fleet: %v", err)
		}
		if emptyFleet.TotalTickets != 0 {
			t.Errorf("expected 0 tickets for future filter, got %d", emptyFleet.TotalTickets)
		}
	})

	t.Run("invalid since or until returns 400 Bad Request", func(t *testing.T) {
		respSince, err := http.Get(ts.URL + "/api/fleet?since=invalid-date")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer respSince.Body.Close()
		if respSince.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", respSince.StatusCode)
		}

		respUntil, err := http.Get(ts.URL + "/api/fleet?until=invalid-date")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer respUntil.Body.Close()
		if respUntil.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", respUntil.StatusCode)
		}
	})
}

func TestCORSPreflight(t *testing.T) {
	ts, _ := setupTestServer(t)

	req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/workspaces", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204 No Content for preflight, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *")
	}
	if resp.Header.Get("Access-Control-Allow-Methods") != "GET, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, OPTIONS")
	}
}

func TestStaticAndSPAFallback(t *testing.T) {
	ts, _ := setupTestServer(t)

	t.Run("serves root index.html", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("expected text/html charset=utf-8, got %s", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("serves existing static asset", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/assets/app.js")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("spa fallback for non-existent client-side route", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/tickets/AL-1")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 SPA fallback, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
			t.Errorf("expected text/html charset=utf-8, got %s", resp.Header.Get("Content-Type"))
		}
	})

	t.Run("api routes do not fallback to html on 404", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/unknown-endpoint")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}

		var errResp adaptershttp.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.Fatalf("expected valid JSON ErrorResponse for missing API endpoint: %v", err)
		}
		if errResp.Error != "not_found" {
			t.Errorf("expected error not_found, got %s", errResp.Error)
		}
	})
}

func TestAdversarialHTTPEdgeCases(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><body>AgenticLens</body></html>"),
		},
	}

	t.Run("empty workspaces directory handling", func(t *testing.T) {
		emptyDir := t.TempDir()
		srv := mustNewServer(t, emptyDir, mockFS)
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		// /api/workspaces returns empty JSON array []
		resp, err := http.Get(ts.URL + "/api/workspaces")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatalf("failed decoding ledgers: %v", err)
		}
		if len(ledgers) != 0 {
			t.Errorf("expected 0 ledgers, got %d", len(ledgers))
		}

		// /api/fleet returns zeroed fleet summary
		respFleet, err := http.Get(ts.URL + "/api/fleet")
		if err != nil {
			t.Fatal(err)
		}
		defer respFleet.Body.Close()
		if respFleet.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", respFleet.StatusCode)
		}
		var fleet domain.FleetSummary
		if err := json.NewDecoder(respFleet.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}
		if fleet.TotalTickets != 0 || fleet.TotalCostUSD != 0.0 {
			t.Errorf("expected zeroed fleet summary, got %+v", fleet)
		}

		// /api/workspaces/ANY returns 404
		respDetail, err := http.Get(ts.URL + "/api/workspaces/ANY")
		if err != nil {
			t.Fatal(err)
		}
		defer respDetail.Body.Close()
		if respDetail.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", respDetail.StatusCode)
		}
	})

	t.Run("missing artifacts for existing ticket returns null content gracefully", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		resp, err := http.Get(ts.URL + "/api/workspaces/AL-2")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			t.Fatalf("failed decoding: %v", err)
		}

		if detail.Artifacts.ContractContent != nil {
			t.Errorf("expected nil ContractContent for AL-2")
		}
		if detail.Artifacts.SummaryContent != nil {
			t.Errorf("expected nil SummaryContent for AL-2")
		}
	})

	t.Run("zero tokens, zero cost, and zero duration ticket API handling", func(t *testing.T) {
		zeroDir := t.TempDir()
		ticketDir := filepath.Join(zeroDir, "AL-ZERO")
		if err := os.MkdirAll(ticketDir, 0755); err != nil {
			t.Fatal(err)
		}
		zeroJSON := `{
  "ticket_id": "AL-ZERO",
  "ticket_title": "Zero Ticket",
  "updated_at": "2026-10-17T12:00:00Z",
  "sessions": [
    {
      "id": "ses-0",
      "invocations": [
        {
          "id": "ses-0",
          "role": "orchestrator",
          "active_duration_seconds": 0,
          "turns": 0,
          "tokens": {
            "input": 0,
            "output": 0,
            "cache_read": 0,
            "reasoning": 0,
            "peak_context": 0
          },
          "cost_usd": 0.0
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`
		if err := os.WriteFile(filepath.Join(ticketDir, "agentic_metrics.json"), []byte(zeroJSON), 0644); err != nil {
			t.Fatal(err)
		}

		srv := mustNewServer(t, zeroDir, mockFS)
		ts := httptest.NewServer(srv.Handler())
		defer ts.Close()

		resp, err := http.Get(ts.URL + "/api/workspaces/AL-ZERO")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		respFleet, err := http.Get(ts.URL + "/api/fleet")
		if err != nil {
			t.Fatal(err)
		}
		defer respFleet.Body.Close()
		if respFleet.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 from fleet, got %d", respFleet.StatusCode)
		}
		var fleet domain.FleetSummary
		if err := json.NewDecoder(respFleet.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}
		if fleet.TotalCostUSD != 0.0 || fleet.TotalTickets != 1 {
			t.Errorf("unexpected fleet values for zero ticket: %+v", fleet)
		}
	})

	t.Run("malformed and edge query parameters", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Malformed query parameters should fall back to defaults and return 200
		urls := []string{
			ts.URL + "/api/workspaces?window=invalid_window&sort=bogus_sort&order=junk",
			ts.URL + "/api/workspaces?sort=cost&order=ASC",
			ts.URL + "/api/workspaces?window=30d",
			ts.URL + "/api/fleet?window=malformed_window",
		}

		for _, u := range urls {
			resp, err := http.Get(u)
			if err != nil {
				t.Fatalf("request failed for %s: %v", u, err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", u, resp.StatusCode)
			}
		}
	})

	t.Run("path traversal attempts via HTTP and handler", func(t *testing.T) {
		ts, wsDir := setupTestServer(t)

		// 1. Encoded traversal via live server
		encodedPaths := []string{
			"/api/workspaces/%2e%2e",
			"/api/workspaces/%2e%2e%2f",
			"/api/workspaces/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		}
		for _, p := range encodedPaths {
			resp, err := http.Get(ts.URL + p)
			if err != nil {
				t.Fatalf("request failed for %s: %v", p, err)
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Errorf("expected non-200 for traversal attempt %s, got %d", p, resp.StatusCode)
			}
		}

		// 2. Direct handler invocation with hostile path values
		detailHandler := adaptershttp.WorkspaceDetailHandler(wsDir)
		eventsHandler := adaptershttp.WorkspaceEventsHandler(wsDir)
		hostileIDs := []string{
			"../../etc/passwd",
			"..",
			"../AL-1",
			"AL-1/../../",
			`..\..\windows\win.ini`,
		}
		for _, hid := range hostileIDs {
			// Detail handler
			reqDetail := httptest.NewRequest(http.MethodGet, "/api/workspaces/"+hid, nil)
			reqDetail.SetPathValue("ticketId", hid)
			recDetail := httptest.NewRecorder()
			detailHandler.ServeHTTP(recDetail, reqDetail)

			if recDetail.Code != http.StatusBadRequest {
				t.Errorf("WorkspaceDetailHandler: expected 400 Bad Request for hostile ticketId %q, got %d", hid, recDetail.Code)
			}

			// Events handler
			reqEvents := httptest.NewRequest(http.MethodGet, "/api/workspaces/"+hid+"/events", nil)
			reqEvents.SetPathValue("ticketId", hid)
			recEvents := httptest.NewRecorder()
			eventsHandler.ServeHTTP(recEvents, reqEvents)

			if recEvents.Code != http.StatusBadRequest {
				t.Errorf("WorkspaceEventsHandler: expected 400 Bad Request for hostile ticketId %q, got %d", hid, recEvents.Code)
			}
		}
	})

	t.Run("spa method not allowed and api routing guards", func(t *testing.T) {
		ts, _ := setupTestServer(t)

		// Method not allowed on SPA routes
		postReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/", nil)
		postResp, err := http.DefaultClient.Do(postReq)
		if err != nil {
			t.Fatal(err)
		}
		postResp.Body.Close()
		if postResp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected 405 Method Not Allowed for POST /, got %d", postResp.StatusCode)
		}

		// API routes must return 404 JSON, never SPA HTML
		apiPaths := []string{"/api", "/api/", "/api/unknown", "/api/workspaces/not-found/sub"}
		for _, ap := range apiPaths {
			resp, err := http.Get(ts.URL + ap)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("expected 404 for %s, got %d", ap, resp.StatusCode)
			}
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected application/json for 404, got %s", contentType)
			}
		}
	})
}

func TestResponseHelpersAndSerialization(t *testing.T) {
	t.Run("WriteWorkspacesResponse writes valid JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		adaptershttp.WriteWorkspacesResponse(rec, []domain.AgenticMetricsLedger{})
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("unexpected content type: %s", rec.Header().Get("Content-Type"))
		}
		if strings.TrimSpace(rec.Body.String()) != "[]" {
			t.Errorf("expected empty array, got %s", rec.Body.String())
		}
	})

	t.Run("WriteWorkspaceDetailResponse writes valid JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		ticketID := "T-100"
		detail := adaptershttp.WorkspaceDetailResponse{
			Ledger: domain.AgenticMetricsLedger{
				TicketID: &ticketID,
			},
		}
		adaptershttp.WriteWorkspaceDetailResponse(rec, detail)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		var parsed adaptershttp.WorkspaceDetailResponse
		if err := json.NewDecoder(rec.Body).Decode(&parsed); err != nil {
			t.Fatalf("failed decoding: %v", err)
		}
		if parsed.Ledger.TicketID == nil || *parsed.Ledger.TicketID != ticketID {
			t.Errorf("unexpected ticket ID: %v", parsed.Ledger.TicketID)
		}
	})

	t.Run("WriteFleetResponse writes valid JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		summary := domain.FleetSummary{
			TotalTickets: 5,
			TotalCostUSD: 1.25,
		}
		adaptershttp.WriteFleetResponse(rec, summary)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		var parsed domain.FleetSummary
		if err := json.NewDecoder(rec.Body).Decode(&parsed); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}
		if parsed.TotalTickets != 5 || parsed.TotalCostUSD != 1.25 {
			t.Errorf("unexpected fleet summary: %+v", parsed)
		}
	})

	t.Run("WriteError writes structured error response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		adaptershttp.WriteError(rec, http.StatusBadRequest, "invalid_input", "invalid query")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
		var errResp adaptershttp.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed decoding error response: %v", err)
		}
		if errResp.Error != "invalid_input" || errResp.Message != "invalid query" || errResp.StatusCode != 400 {
			t.Errorf("unexpected error response: %+v", errResp)
		}
	})

	t.Run("WriteVersionResponse writes valid JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		info := domain.BuildInfo{
			Version:   "v0.1.0-dev",
			GitCommit: "abcdef1",
			BuildDate: "2026-10-18T10:00:00Z",
		}
		adaptershttp.WriteVersionResponse(rec, info)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("unexpected content type: %s", rec.Header().Get("Content-Type"))
		}
		var parsed adaptershttp.BuildInfoDTO
		if err := json.NewDecoder(rec.Body).Decode(&parsed); err != nil {
			t.Fatalf("failed decoding version response: %v", err)
		}
		if parsed.Version != "v0.1.0-dev" || parsed.Commit != "abcdef1" || parsed.BuildDate != "2026-10-18T10:00:00Z" {
			t.Errorf("unexpected version DTO values: %+v", parsed)
		}
	})

	t.Run("WriteEventsResponse writes valid JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		events := domain.WorkspaceEventsResponse{
			TicketID:   "TKT-1",
			IsHydrated: false,
			Events: []domain.EventFact{
				{
					SessionID: "ses-1",
					Timestamp: "2026-10-18T10:00:00Z",
					Turn:      1,
					Role:      "coder",
					EventType: domain.EventTypeToolCall,
					Status:    "success",
				},
			},
			DataSource: domain.DataSourceCommittedLedger,
		}
		adaptershttp.WriteEventsResponse(rec, events)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		var parsed domain.WorkspaceEventsResponse
		if err := json.NewDecoder(rec.Body).Decode(&parsed); err != nil {
			t.Fatalf("failed decoding events response: %v", err)
		}
		if len(parsed.Events) != 1 || parsed.Events[0].SessionID != "ses-1" {
			t.Errorf("unexpected events parsed: %+v", parsed)
		}
	})
}

func TestGetVersion(t *testing.T) {
	ts, _ := setupTestServer(t)

	t.Run("successful GET returns 200 with BuildInfoDTO schema", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/version")
		if err != nil {
			t.Fatalf("failed GET /api/version: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		contentType := resp.Header.Get("Content-Type")
		if contentType != "application/json; charset=utf-8" {
			t.Errorf("expected application/json; charset=utf-8, got %q", contentType)
		}

		var dto adaptershttp.BuildInfoDTO
		if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
			t.Fatalf("failed decoding version response: %v", err)
		}

		if dto.Version == "" {
			t.Errorf("expected non-empty version in response")
		}
		if dto.Commit == "" {
			t.Errorf("expected non-empty commit in response")
		}
		if dto.BuildDate == "" {
			t.Errorf("expected non-empty build_date in response")
		}
	})

	t.Run("unsupported HTTP methods return non-200 status", func(t *testing.T) {
		disallowedMethods := []string{
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
		}

		for _, method := range disallowedMethods {
			req, err := http.NewRequest(method, ts.URL+"/api/version", nil)
			if err != nil {
				t.Fatalf("failed creating request: %v", err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed executing %s /api/version: %v", method, err)
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				t.Errorf("expected non-200 for %s /api/version, got %d", method, resp.StatusCode)
			}
		}
	})

	t.Run("subpath under version endpoint returns 404 JSON, not SPA fallback", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/version/unrecognized")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 for subpath /api/version/unrecognized, got %d", resp.StatusCode)
		}
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected application/json for 404, got %s", contentType)
		}
	})

	t.Run("query parameters are safely ignored and 200 OK returned", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/version?param=junk&another=test&flag=true")
		if err != nil {
			t.Fatalf("failed GET with query params: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK with query parameters, got %d", resp.StatusCode)
		}

		var dto adaptershttp.BuildInfoDTO
		if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
			t.Fatalf("failed decoding version response: %v", err)
		}
		if dto.Version == "" {
			t.Errorf("expected non-empty version")
		}
	})

	t.Run("CORS preflight on version endpoint returns 204 No Content", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodOptions, ts.URL+"/api/version", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", "GET")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed OPTIONS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 No Content for CORS preflight, got %d", resp.StatusCode)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("expected Access-Control-Allow-Origin: *")
		}
	})

	t.Run("empty domain variables fall back to defaults in HTTP response", func(t *testing.T) {
		origVersion := domain.Version
		origCommit := domain.GitCommit
		origDate := domain.BuildDate
		t.Cleanup(func() {
			domain.Version = origVersion
			domain.GitCommit = origCommit
			domain.BuildDate = origDate
		})

		domain.Version = ""
		domain.GitCommit = ""
		domain.BuildDate = ""

		resp, err := http.Get(ts.URL + "/api/version")
		if err != nil {
			t.Fatalf("failed GET: %v", err)
		}
		defer resp.Body.Close()

		var dto adaptershttp.BuildInfoDTO
		if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
			t.Fatalf("failed decoding: %v", err)
		}

		if dto.Version != domain.DefaultVersion {
			t.Errorf("expected default version %q, got %q", domain.DefaultVersion, dto.Version)
		}
		if dto.Commit != domain.DefaultGitCommit {
			t.Errorf("expected default commit %q, got %q", domain.DefaultGitCommit, dto.Commit)
		}
		if dto.BuildDate != domain.DefaultBuildDate {
			t.Errorf("expected default build date %q, got %q", domain.DefaultBuildDate, dto.BuildDate)
		}
	})
}

func TestWorkspaceEventsAndAnomaliesHTTP(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed creating workspaces dir: %v", err)
	}

	// Workspace 1: with events and anomalies
	w1Dir := filepath.Join(wsDir, "AL-EVENT-1")
	if err := os.MkdirAll(w1Dir, 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}
	w1JSON := `{
  "ticket_id": "AL-EVENT-1",
  "ticket_title": "Event Ticket",
  "ticket_type": "feature",
  "started_at": "2026-10-18T10:00:00Z",
  "completed_at": "2026-10-18T11:00:00Z",
  "updated_at": "2026-10-18T11:00:00Z",
  "code": {
    "lines_added": 10,
    "lines_deleted": 0,
    "test_lines_added": 5,
    "test_lines_deleted": 0,
    "prod_lines_added": 5,
    "prod_lines_deleted": 0,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "ses-1",
      "invocations": [
        {
          "id": "ses-1",
          "role": "orchestrator",
          "model": "gpt-4",
          "user": "dev",
          "active_duration_seconds": 60,
          "turns": 5,
          "tokens": {
            "input": 1000,
            "output": 200,
            "cache_read": 500,
            "reasoning": 0,
            "peak_context": 50000
          },
          "cost_usd": 0.1,
          "compactions": 0
        }
      ],
      "tools": [],
      "events": [
        {
          "timestamp": "2026-10-18T10:00:00Z",
          "turn": 1,
          "role": "coder",
          "event_type": "tool_call",
          "tool_name": "bash",
          "status": "error"
        },
        {
          "timestamp": "2026-10-18T10:00:05Z",
          "turn": 1,
          "role": "coder",
          "event_type": "tool_call",
          "tool_name": "bash",
          "status": "error"
        },
        {
          "timestamp": "2026-10-18T10:00:10Z",
          "turn": 1,
          "role": "coder",
          "event_type": "tool_call",
          "tool_name": "bash",
          "status": "error"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(w1Dir, "agentic_metrics.json"), []byte(w1JSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Workspace 2: legacy v1 with no events and no anomalies
	w2Dir := filepath.Join(wsDir, "AL-LEGACY")
	if err := os.MkdirAll(w2Dir, 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}
	w2JSON := `{
  "ticket_id": "AL-LEGACY",
  "ticket_title": "Legacy Ticket",
  "ticket_type": "chore",
  "started_at": "2026-10-17T10:00:00Z",
  "completed_at": "2026-10-17T10:30:00Z",
  "updated_at": "2026-10-17T10:30:00Z",
  "code": {
    "lines_added": 5,
    "lines_deleted": 0,
    "test_lines_added": 0,
    "test_lines_deleted": 0,
    "prod_lines_added": 5,
    "prod_lines_deleted": 0,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": []
}`
	if err := os.WriteFile(filepath.Join(w2Dir, "agentic_metrics.json"), []byte(w2JSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Workspace 3: with unsorted events
	w3Dir := filepath.Join(wsDir, "AL-UNSORTED")
	if err := os.MkdirAll(w3Dir, 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}
	w3JSON := `{
		"ticket_id": "AL-UNSORTED",
		"ticket_title": "Unsorted Events Ticket",
		"ticket_type": "feature",
		"started_at": "2026-10-18T10:00:00Z",
		"sessions": [
			{"id":"ses-late","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T12:00:00Z","turn":2,"role":"coder","event_type":"tool_call","status":"success"}]},
			{"id":"ses-early-t2","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T10:00:00Z","turn":2,"role":"coder","event_type":"tool_call","status":"success"}]},
			{"id":"ses-early-t1","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T10:00:00Z","turn":1,"role":"orchestrator","event_type":"turn_message","status":"success"}]}
		]
	}`
	if err := os.WriteFile(filepath.Join(w3Dir, "agentic_metrics.json"), []byte(w3JSON), 0644); err != nil {
		t.Fatal(err)
	}

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<!DOCTYPE html><html><head><title>AgenticLens</title></head><body>Root</body></html>"),
		},
	}
	server := mustNewServer(t, wsDir, mockFS)
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(func() { ts.Close() })

	t.Run("GET /api/workspaces/AL-EVENT-1/events returns events array", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-EVENT-1/events")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var eventsResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventsResp); err != nil {
			t.Fatalf("failed decoding events response: %v", err)
		}
		if eventsResp.TicketID != "AL-EVENT-1" {
			t.Errorf("expected TicketID AL-EVENT-1, got %s", eventsResp.TicketID)
		}
		if eventsResp.IsHydrated {
			t.Errorf("expected IsHydrated == false without live hook, got true")
		}
		if eventsResp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %s, got %s", domain.DataSourceCommittedLedger, eventsResp.DataSource)
		}
		if len(eventsResp.Events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(eventsResp.Events))
		}
		if eventsResp.Events[0].Timestamp != "2026-10-18T10:00:00Z" || eventsResp.Events[1].Timestamp != "2026-10-18T10:00:05Z" || eventsResp.Events[2].Timestamp != "2026-10-18T10:00:10Z" {
			t.Errorf("unexpected events order or content: %+v", eventsResp.Events)
		}
	})

	t.Run("GET /api/workspaces/AL-LEGACY/events returns empty array", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-LEGACY/events")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var eventsResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventsResp); err != nil {
			t.Fatalf("failed decoding events response: %v", err)
		}
		if eventsResp.TicketID != "AL-LEGACY" {
			t.Errorf("expected TicketID AL-LEGACY, got %s", eventsResp.TicketID)
		}
		if eventsResp.IsHydrated {
			t.Errorf("expected IsHydrated == false, got true")
		}
		if len(eventsResp.Events) != 0 {
			t.Fatalf("expected 0 events, got %d", len(eventsResp.Events))
		}
	})

	t.Run("GET /api/workspaces/NON-EXISTENT/events returns 404", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/NON-EXISTENT/events")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}

		var errResp adaptershttp.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed decoding error response: %v", err)
		}
		if errResp.Error != "not_found" {
			t.Errorf("expected not_found, got %s", errResp.Error)
		}
		if errResp.Message != "Ticket workspace not found" {
			t.Errorf("expected 'Ticket workspace not found', got %q", errResp.Message)
		}
	})

	t.Run("GET /api/workspaces/AL-UNSORTED/events returns events sorted by timestamp and turn", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-UNSORTED/events")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var eventsResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventsResp); err != nil {
			t.Fatalf("failed decoding events: %v", err)
		}
		if len(eventsResp.Events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(eventsResp.Events))
		}
		if eventsResp.Events[0].SessionID != "ses-early-t1" || eventsResp.Events[1].SessionID != "ses-early-t2" || eventsResp.Events[2].SessionID != "ses-late" {
			t.Errorf("expected sorted events [ses-early-t1, ses-early-t2, ses-late], got [%s, %s, %s]",
				eventsResp.Events[0].SessionID, eventsResp.Events[1].SessionID, eventsResp.Events[2].SessionID)
		}
	})

	t.Run("GET /api/workspaces/AL-EVENT-1 includes non-empty anomalies", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-EVENT-1")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
			t.Fatalf("failed decoding workspace detail: %v", err)
		}
		if len(detail.Anomalies) == 0 {
			t.Fatalf("expected detected anomalies in AL-EVENT-1, got empty slice")
		}
		if detail.Anomalies[0].RuleID != domain.RuleToolRetryStreak {
			t.Errorf("expected TOOL_RETRY_STREAK rule, got %s", detail.Anomalies[0].RuleID)
		}
	})

	t.Run("GET /api/workspaces/AL-LEGACY includes empty array for anomalies (never null)", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-LEGACY")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed reading body: %v", err)
		}

		var rawMap map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawMap); err != nil {
			t.Fatalf("failed parsing raw JSON: %v", err)
		}

		anomaliesJSON, exists := rawMap["anomalies"]
		if !exists {
			t.Fatal("anomalies key must exist in workspace detail JSON")
		}
		if string(anomaliesJSON) != "[]" {
			t.Fatalf("expected anomalies to be [], got %s", string(anomaliesJSON))
		}
	})

	t.Run("GET /api/fleet includes anomaly_summary", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/fleet")
		if err != nil {
			t.Fatalf("HTTP get failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var fleet domain.FleetSummary
		if err := json.NewDecoder(resp.Body).Decode(&fleet); err != nil {
			t.Fatalf("failed decoding fleet: %v", err)
		}
		if fleet.AnomalySummary.TotalAnomalies == 0 {
			t.Errorf("expected fleet anomaly summary to reflect AL-EVENT-1 anomaly, got 0")
		}
		if fleet.AnomalySummary.TicketsWithAnomalies != 1 {
			t.Errorf("expected 1 ticket with anomalies, got %d", fleet.AnomalySummary.TicketsWithAnomalies)
		}
	})

	t.Run("path traversal on /events is rejected", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/..%2F..%2Fetc%2Fpasswd/events")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 400 or 404 for traversal, got %d", resp.StatusCode)
		}
	})
}

func TestWorkspaceEventsHandler_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	wsDir := filepath.Join(tempDir, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Setup 1: Workspace with zero-byte metrics file
	zeroByteDir := filepath.Join(wsDir, "AL-ZERO-BYTE")
	if err := os.MkdirAll(zeroByteDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(zeroByteDir, "agentic_metrics.json"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup 2: Workspace with corrupted/truncated JSON
	corruptedDir := filepath.Join(wsDir, "AL-CORRUPTED")
	if err := os.MkdirAll(corruptedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(corruptedDir, "agentic_metrics.json"), []byte(`{"ticket_id": "AL-CORRUPTED", "events": [{"turn":`), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup 3: Workspace with empty events field
	emptyEventsDir := filepath.Join(wsDir, "AL-NO-EVENTS")
	if err := os.MkdirAll(emptyEventsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(emptyEventsDir, "agentic_metrics.json"), []byte(`{"ticket_id": "AL-NO-EVENTS", "updated_at": "2026-10-18T10:00:00Z"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup 4: Valid ticket workspace alongside corrupted sibling
	validDir := filepath.Join(wsDir, "AL-TARGETED-VALID")
	if err := os.MkdirAll(validDir, 0755); err != nil {
		t.Fatal(err)
	}
	validJSON := `{
		"ticket_id": "AL-TARGETED-VALID",
		"updated_at": "2026-10-18T10:00:00Z",
		"sessions": [
			{"id":"ses-2","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T10:00:10Z","turn":2,"role":"orchestrator","event_type":"tool_call","status":"completed"}]},
			{"id":"ses-1","invocations":[],"tools":[],"events":[{"timestamp":"2026-10-18T10:00:05Z","turn":1,"role":"orchestrator","event_type":"tool_call","status":"completed"}]}
		]
	}`
	if err := os.WriteFile(filepath.Join(validDir, "agentic_metrics.json"), []byte(validJSON), 0644); err != nil {
		t.Fatal(err)
	}

	handler := adaptershttp.WorkspaceEventsHandler(wsDir)

	t.Run("empty ticketId parameter returns 400 Bad Request", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces//events", nil)
		req.SetPathValue("ticketId", "")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for empty ticketId, got %d", rec.Code)
		}
	})

	t.Run("path traversal attempts in ticketId return 400 Bad Request", func(t *testing.T) {
		t.Parallel()
		hostileIDs := []string{
			"..",
			"../AL-1",
			"..\\AL-1",
			"sub/ticket",
			"ticket/../other",
		}
		for _, hid := range hostileIDs {
			req := httptest.NewRequest(http.MethodGet, "/api/workspaces/"+hid+"/events", nil)
			req.SetPathValue("ticketId", hid)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 Bad Request for hostile ticketId %q, got %d", hid, rec.Code)
			}
		}
	})

	t.Run("zero-byte metrics file returns 404 without panicking", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-ZERO-BYTE/events", nil)
		req.SetPathValue("ticketId", "AL-ZERO-BYTE")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found for zero-byte metrics file, got %d", rec.Code)
		}
	})

	t.Run("corrupted JSON metrics file returns 404 without panicking", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-CORRUPTED/events", nil)
		req.SetPathValue("ticketId", "AL-CORRUPTED")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found for corrupted JSON metrics file, got %d", rec.Code)
		}
	})

	t.Run("workspace with missing events array returns 200 with empty array JSON", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-NO-EVENTS/events", nil)
		req.SetPathValue("ticketId", "AL-NO-EVENTS")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("expected application/json content type, got %s", contentType)
		}

		var eventsResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(rec.Body).Decode(&eventsResp); err != nil {
			t.Fatalf("failed decoding events response: %v", err)
		}
		if len(eventsResp.Events) != 0 {
			t.Errorf("expected 0 events, got %d", len(eventsResp.Events))
		}
		if eventsResp.IsHydrated {
			t.Errorf("expected IsHydrated == false, got true")
		}
	})

	t.Run("corrupted sibling workspace does not interfere with targeted ticket load", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-TARGETED-VALID/events", nil)
		req.SetPathValue("ticketId", "AL-TARGETED-VALID")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for valid ticket despite corrupted sibling, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var eventsResp domain.WorkspaceEventsResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &eventsResp); err != nil {
			t.Fatalf("failed unmarshaling events response: %v", err)
		}
		if len(eventsResp.Events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(eventsResp.Events))
		}
		if eventsResp.Events[0].SessionID != "ses-1" || eventsResp.Events[1].SessionID != "ses-2" {
			t.Errorf("expected events stably sorted: %+v", eventsResp.Events)
		}
	})

	t.Run("hydrator success returns hydrated events response with SQLite data source", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-TARGETED-VALID/events", nil)
		req.SetPathValue("ticketId", "AL-TARGETED-VALID")
		rec := httptest.NewRecorder()

		prompt := "test prompt"
		mockH := &mockHydrator{
			resp: domain.WorkspaceEventsResponse{
				TicketID:   "AL-TARGETED-VALID",
				IsHydrated: true,
				Events: []domain.EventFact{
					{
						SessionID:     "ses-1",
						Timestamp:     "2026-10-18T10:00:00Z",
						Turn:          1,
						Role:          "coder",
						EventType:     domain.EventTypeTurnMessage,
						Status:        "success",
						PromptContent: &prompt,
					},
				},
				DataSource: domain.DataSourceLocalSQLite,
			},
		}

		h := adaptershttp.WorkspaceEventsHandler(wsDir, mockH)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}
		if !resp.IsHydrated {
			t.Errorf("expected IsHydrated == true")
		}
		if resp.DataSource != domain.DataSourceLocalSQLite {
			t.Errorf("expected DataSource == %s, got %s", domain.DataSourceLocalSQLite, resp.DataSource)
		}
		if len(resp.Events) != 1 || resp.Events[0].PromptContent == nil || *resp.Events[0].PromptContent != "test prompt" {
			t.Errorf("unexpected hydrated events: %+v", resp.Events)
		}
	})

	t.Run("hydrator failure gracefully falls back to committed ledger", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-TARGETED-VALID/events", nil)
		req.SetPathValue("ticketId", "AL-TARGETED-VALID")
		rec := httptest.NewRecorder()

		mockH := &mockHydrator{
			err: fmt.Errorf("sqlite database locked"),
		}

		h := adaptershttp.WorkspaceEventsHandler(wsDir, mockH)
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		var resp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed decoding response: %v", err)
		}
		if resp.IsHydrated {
			t.Errorf("expected IsHydrated == false on hydrator error")
		}
		if resp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %s, got %s", domain.DataSourceCommittedLedger, resp.DataSource)
		}
		if len(resp.Events) != 2 {
			t.Errorf("expected 2 fallback events from ledger, got %d", len(resp.Events))
		}
	})
}

type mockHydrator struct {
	resp domain.WorkspaceEventsResponse
	err  error
}

func (m *mockHydrator) HydrateWorkspaceEvents(ctx context.Context, ticketID string, ledger domain.AgenticMetricsLedger) (domain.WorkspaceEventsResponse, error) {
	if m.err != nil {
		return domain.WorkspaceEventsResponse{}, m.err
	}
	return m.resp, nil
}

func TestWorkspaceEvents_EndToEndLiveServiceHydrationAndPathTraversal(t *testing.T) {
	// Setup temporary SQLite database
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "test_events.db")

	rwDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed opening rw db: %v", err)
	}

	if _, err := rwDB.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		rwDB.Close()
		t.Fatalf("failed setting WAL mode: %v", err)
	}

	_, err = rwDB.Exec(`
		CREATE TABLE session (
			id TEXT PRIMARY KEY,
			parent_id TEXT,
			agent TEXT,
			model TEXT,
			title TEXT,
			time_created INTEGER,
			time_updated INTEGER
		);
		CREATE TABLE message (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
		CREATE TABLE part (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			message_id TEXT,
			time_created INTEGER,
			time_updated INTEGER,
			data TEXT
		);
	`)
	if err != nil {
		rwDB.Close()
		t.Fatalf("failed creating schema: %v", err)
	}

	now := time.Now().UnixMilli()

	// Seed session ses_al33
	_, err = rwDB.Exec(`
		INSERT INTO session (id, parent_id, agent, model, title, time_created, time_updated)
		VALUES ('ses_al33', NULL, 'orchestrator', 'gemini-3.8-flash', 'Feature AL-33 End-to-End', ?, ?);
	`, now, now)
	if err != nil {
		rwDB.Close()
		t.Fatalf("failed seeding session: %v", err)
	}

	// Seed user prompt and assistant response with reasoning and tokens
	_, err = rwDB.Exec(`
		INSERT INTO message (id, session_id, time_created, time_updated, data)
		VALUES
			('m_user_1', 'ses_al33', ?, ?, '{"role":"user","content":"Execute full dual-plane test"}'),
			('m_asst_1', 'ses_al33', ?, ?, '{"role":"assistant","content":"Testing dual-plane","reasoning":"Verifying invariants","tokens":{"input":2000,"output":300,"cache":{"read":1000},"reasoning":80}}');
	`, now+100, now+100, now+200, now+350)
	if err != nil {
		rwDB.Close()
		t.Fatalf("failed seeding messages: %v", err)
	}

	// Seed tool call part
	_, err = rwDB.Exec(`
		INSERT INTO part (id, session_id, message_id, time_created, time_updated, data)
		VALUES
			('p_tool_1', 'ses_al33', 'm_asst_1', ?, ?, '{"type":"tool","tool":"test_runner","state":{"status":"completed","input":{"suite":"backend"},"output":"all tests passed"}}');
	`, now+250, now+300)
	if err != nil {
		rwDB.Close()
		t.Fatalf("failed seeding part: %v", err)
	}
	rwDB.Close()

	// Setup workspaces directory
	wsDir := t.TempDir()

	// 1. Ticket AL-33 (mapped to ses_al33)
	al33Dir := filepath.Join(wsDir, "AL-33")
	if err := os.MkdirAll(al33Dir, 0755); err != nil {
		t.Fatal(err)
	}
	al33JSON := `{
  "ticket_id": "AL-33",
  "ticket_title": "End-to-End Dual Plane",
  "ticket_type": "feature",
  "started_at": "2026-10-19T10:00:00Z",
  "completed_at": "2026-10-19T11:00:00Z",
  "updated_at": "2026-10-19T11:00:00Z",
  "sessions": [
    {
      "id": "ses_al33",
      "invocations": [
        {
          "id": "ses_al33",
          "role": "orchestrator",
          "model": "gemini-3.8-flash"
        }
      ],
      "tools": [],
      "events": [
        {
          "timestamp": "2026-10-19T10:00:00Z",
          "turn": 1,
          "role": "orchestrator",
          "event_type": "turn_message",
          "status": "completed"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(al33Dir, "agentic_metrics.json"), []byte(al33JSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 2. Ticket AL-OFFLINE (mapped to a session not present in SQLite)
	alOfflineDir := filepath.Join(wsDir, "AL-OFFLINE")
	if err := os.MkdirAll(alOfflineDir, 0755); err != nil {
		t.Fatal(err)
	}
	offlineJSON := `{
  "ticket_id": "AL-OFFLINE",
  "ticket_title": "Offline Ticket",
  "ticket_type": "fix",
  "started_at": "2026-10-19T09:00:00Z",
  "completed_at": "2026-10-19T09:30:00Z",
  "updated_at": "2026-10-19T09:30:00Z",
  "sessions": [
    {
      "id": "ses_missing_99",
      "invocations": [
        {
          "id": "ses_missing_99",
          "role": "coder",
          "model": "gemini-3.8-flash"
        }
      ],
      "tools": [],
      "events": [
        {
          "timestamp": "2026-10-19T09:00:00Z",
          "turn": 1,
          "role": "coder",
          "event_type": "compaction",
          "status": "completed"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(filepath.Join(alOfflineDir, "agentic_metrics.json"), []byte(offlineJSON), 0644); err != nil {
		t.Fatal(err)
	}

	hook, err := live.OpenReadOnlyHook(dbPath, 5000)
	if err != nil {
		t.Fatalf("OpenReadOnlyHook failed: %v", err)
	}
	defer hook.Close()

	liveSvc := live.NewLiveService(hook, domain.ProjectConfig{})
	defer liveSvc.Close()

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<!DOCTYPE html><html><body>Root</body></html>")},
	}

	server := mustNewServer(t, wsDir, mockFS, liveSvc)
	ts := httptest.NewServer(server.Handler())
	defer ts.Close()

	// Snapshot directory state before any HTTP calls
	beforeSnap, err := os.ReadDir(wsDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("GET /api/workspaces/AL-33/events returns hydrated response from SQLite", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-33/events")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var eventResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if eventResp.TicketID != "AL-33" {
			t.Errorf("expected TicketID AL-33, got %s", eventResp.TicketID)
		}
		if !eventResp.IsHydrated {
			t.Errorf("expected IsHydrated == true")
		}
		if eventResp.DataSource != domain.DataSourceLocalSQLite {
			t.Errorf("expected DataSource == %s, got %s", domain.DataSourceLocalSQLite, eventResp.DataSource)
		}
		if len(eventResp.Events) < 2 {
			t.Fatalf("expected at least 2 hydrated events, got %d", len(eventResp.Events))
		}

		// Verify rich diagnostic context
		var foundTurnMsg, foundToolCall bool
		for _, ev := range eventResp.Events {
			if ev.EventType == domain.EventTypeTurnMessage {
				foundTurnMsg = true
				if ev.PromptContent == nil || *ev.PromptContent != "Execute full dual-plane test" {
					t.Errorf("unexpected prompt content: %v", ev.PromptContent)
				}
				if ev.ReasoningContent == nil || *ev.ReasoningContent != "Verifying invariants" {
					t.Errorf("unexpected reasoning content: %v", ev.ReasoningContent)
				}
				if ev.ContextTokens != 3000 { // 2000 input + 1000 cache
					t.Errorf("expected 3000 tokens, got %d", ev.ContextTokens)
				}
			}
			if ev.EventType == domain.EventTypeToolCall {
				foundToolCall = true
				if ev.ToolName == nil || *ev.ToolName != "test_runner" {
					t.Errorf("unexpected tool name: %v", ev.ToolName)
				}
				if ev.ToolInput == nil || !strings.Contains(*ev.ToolInput, "backend") {
					t.Errorf("unexpected tool input: %v", ev.ToolInput)
				}
				if ev.ToolOutput == nil || *ev.ToolOutput != "all tests passed" {
					t.Errorf("unexpected tool output: %v", ev.ToolOutput)
				}
			}
		}

		if !foundTurnMsg {
			t.Errorf("missing turn_message event")
		}
		if !foundToolCall {
			t.Errorf("missing tool_call event")
		}
	})

	t.Run("GET /api/workspaces/AL-OFFLINE/events gracefully falls back to committed ledger", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/workspaces/AL-OFFLINE/events")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var eventResp domain.WorkspaceEventsResponse
		if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
			t.Fatalf("failed decoding json: %v", err)
		}

		if eventResp.TicketID != "AL-OFFLINE" {
			t.Errorf("expected TicketID AL-OFFLINE, got %s", eventResp.TicketID)
		}
		if eventResp.IsHydrated {
			t.Errorf("expected IsHydrated == false for missing session")
		}
		if eventResp.DataSource != domain.DataSourceCommittedLedger {
			t.Errorf("expected DataSource == %s, got %s", domain.DataSourceCommittedLedger, eventResp.DataSource)
		}
		if len(eventResp.Events) != 1 {
			t.Fatalf("expected 1 fallback event, got %d", len(eventResp.Events))
		}
		if eventResp.Events[0].EventType != domain.EventTypeCompaction {
			t.Errorf("expected compaction event, got %s", eventResp.Events[0].EventType)
		}
	})

	t.Run("GET /api/workspaces/{ticketId}/events rejects adversarial path traversal", func(t *testing.T) {
		hostilePaths := []string{
			"/api/workspaces/..%2F..%2Fetc%2Fpasswd/events",
			"/api/workspaces/%2e%2e%2f%2e%2e%2fetc/events",
			"/api/workspaces/..%5c..%5cwindows%5cwin.ini/events",
			"/api/workspaces/../events",
			"/api/workspaces/..../events",
		}

		for _, p := range hostilePaths {
			req, err := http.NewRequest(http.MethodGet, ts.URL+p, nil)
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
				t.Errorf("path %s returned unexpected status %d, expected 400 or 404", p, resp.StatusCode)
			}
		}
	})

	// Assert ZERO disk writes in workspaces directory during all hydration and fallback requests
	afterSnap, err := os.ReadDir(wsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(beforeSnap) != len(afterSnap) {
		t.Errorf("zero disk persistence violation: workspace directory count changed from %d to %d", len(beforeSnap), len(afterSnap))
	}
}

func setupMultiFilterWorkspaces(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	wsDir := filepath.Join(root, "workspaces")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed creating workspaces dir: %v", err)
	}

	tickets := []struct {
		id          string
		ticketType  string
		completedAt string
		cost        string
	}{
		{"AL-1", "feature", "2026-10-15T11:00:00Z", "1.00"},
		{"AL-2", "bug", "2026-10-16T11:00:00Z", "2.00"},
		{"PO-1", "docs", "2026-10-17T11:00:00Z", "3.00"},
		{"GEN-1", "chore", "2026-10-18T11:00:00Z", "4.00"},
	}

	for _, tc := range tickets {
		dir := filepath.Join(wsDir, tc.id)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed creating dir for %s: %v", tc.id, err)
		}
		rawJSON := `{
			"ticket_id": "` + tc.id + `",
			"ticket_title": "Title ` + tc.id + `",
			"ticket_type": "` + tc.ticketType + `",
			"started_at": "2026-10-10T10:00:00Z",
			"completed_at": "` + tc.completedAt + `",
			"updated_at": "` + tc.completedAt + `",
			"code": {"lines_added": 10, "lines_deleted": 0, "test_lines_added": 5, "test_lines_deleted": 0, "prod_lines_added": 5, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
			"sessions": [{
			"id": "ses-` + tc.id + `",
			"invocations": [
				{
					"id": "ses-` + tc.id + `",
					"role": "orchestrator",
					"model": "gpt-4",
					"user": "dev",
					"active_duration_seconds": 60,
					"turns": 1,
					"tokens": {"input": 100, "output": 50, "cache_read": 0, "reasoning": 0, "peak_context": 5000},
					"cost_usd": ` + tc.cost + `,
					"compactions": 0
				}
			],
			"tools": [],
			"events": []
			}]
		}`
		if err := os.WriteFile(filepath.Join(dir, "agentic_metrics.json"), []byte(rawJSON), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return wsDir
}

func TestMultiFilterEndpoints(t *testing.T) {
	t.Parallel()

	wsDir := setupMultiFilterWorkspaces(t)
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>OK</html>")},
	}
	server := mustNewServer(t, wsDir, mockFS)
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(func() { ts.Close() })

	t.Run("GET /api/workspaces with repeated and comma-separated ticket_type and prefix filters", func(t *testing.T) {
		t.Parallel()

		// 1. Single ticket_type alias feat -> AL-1
		resp, err := http.Get(ts.URL + "/api/workspaces?ticket_type=feat")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()
		if len(ledgers) != 1 || *ledgers[0].TicketID != "AL-1" {
			t.Fatalf("expected AL-1, got %+v", ledgers)
		}

		// 2. Comma-separated ticket_type: feat,bug -> AL-1, AL-2
		resp2, err := http.Get(ts.URL + "/api/workspaces?ticket_type=feat,bug")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers2 []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp2.Body).Decode(&ledgers2); err != nil {
			t.Fatal(err)
		}
		_ = resp2.Body.Close()
		if len(ledgers2) != 2 {
			t.Fatalf("expected 2 ledgers for feat,bug, got %d", len(ledgers2))
		}

		// 3. Repeated ticket_type params: ticket_type=feat&ticket_type=docs -> AL-1, PO-1
		resp3, err := http.Get(ts.URL + "/api/workspaces?ticket_type=feat&ticket_type=docs")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers3 []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp3.Body).Decode(&ledgers3); err != nil {
			t.Fatal(err)
		}
		_ = resp3.Body.Close()
		if len(ledgers3) != 2 {
			t.Fatalf("expected 2 ledgers, got %d", len(ledgers3))
		}

		// 4. Prefix filter: prefix=AL -> AL-1, AL-2
		resp4, err := http.Get(ts.URL + "/api/workspaces?prefix=al")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers4 []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp4.Body).Decode(&ledgers4); err != nil {
			t.Fatal(err)
		}
		_ = resp4.Body.Close()
		if len(ledgers4) != 2 {
			t.Fatalf("expected 2 ledgers for prefix=al, got %d", len(ledgers4))
		}

		// 5. Joint prefix and ticket_type: prefix=AL&ticket_type=bug -> AL-2
		resp5, err := http.Get(ts.URL + "/api/workspaces?prefix=AL&ticket_type=bug")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers5 []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp5.Body).Decode(&ledgers5); err != nil {
			t.Fatal(err)
		}
		_ = resp5.Body.Close()
		if len(ledgers5) != 1 || *ledgers5[0].TicketID != "AL-2" {
			t.Fatalf("expected AL-2, got %+v", ledgers5)
		}

		// 6. 'all' resets filter -> all 4 ledgers
		resp6, err := http.Get(ts.URL + "/api/workspaces?prefix=all&ticket_type=all")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers6 []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp6.Body).Decode(&ledgers6); err != nil {
			t.Fatal(err)
		}
		_ = resp6.Body.Close()
		if len(ledgers6) != 4 {
			t.Fatalf("expected 4 ledgers, got %d", len(ledgers6))
		}
	})

	t.Run("GET /api/fleet with multi-filters", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get(ts.URL + "/api/fleet?ticket_type=chore")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}
		var summary domain.FleetSummary
		if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
			t.Fatal(err)
		}
		if summary.TotalTickets != 1 {
			t.Errorf("expected 1 ticket for chore filter, got %d", summary.TotalTickets)
		}
		if summary.TotalCostUSD != 4.00 {
			t.Errorf("expected 4.00 cost, got %f", summary.TotalCostUSD)
		}
	})

	t.Run("GET /api/comparison endpoint", func(t *testing.T) {
		t.Parallel()

		// 1. Success with ticket_type filter
		resp, err := http.Get(ts.URL + "/api/comparison?ticket_type=bug")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var rows []domain.TicketComparisonRow
		if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 row for bug filter, got %d", len(rows))
		}
		if rows[0].TicketID != "AL-2" {
			t.Errorf("expected ticket ID AL-2, got %s", rows[0].TicketID)
		}
		if rows[0].TicketType == nil || *rows[0].TicketType != domain.TicketTypeFix {
			t.Errorf("expected ticket_type fix, got %v", rows[0].TicketType)
		}

		// 2. Bad request on invalid temporal filter
		respBad, err := http.Get(ts.URL + "/api/comparison?since=not-a-date")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = respBad.Body.Close() }()
		if respBad.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", respBad.StatusCode)
		}
	})

	t.Run("GET /api/progression endpoint", func(t *testing.T) {
		t.Parallel()

		// 1. Success - chronologically sorted buckets
		resp, err := http.Get(ts.URL + "/api/progression?prefix=AL,PO")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var buckets []domain.TimeBucketMetric
		if err := json.NewDecoder(resp.Body).Decode(&buckets); err != nil {
			t.Fatal(err)
		}
		if len(buckets) != 3 {
			t.Fatalf("expected 3 progression buckets for AL,PO, got %d", len(buckets))
		}
		// Chronological check: 2026-10-15 -> 2026-10-16 -> 2026-10-17
		if buckets[0].BucketKey != "2026-10-15" || buckets[1].BucketKey != "2026-10-16" || buckets[2].BucketKey != "2026-10-17" {
			t.Errorf("expected chronological order 2026-10-15, 2026-10-16, 2026-10-17, got %s, %s, %s",
				buckets[0].BucketKey, buckets[1].BucketKey, buckets[2].BucketKey)
		}

		// 2. Bad request on invalid temporal filter
		respBad, err := http.Get(ts.URL + "/api/progression?since=not-a-date")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = respBad.Body.Close() }()
		if respBad.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request, got %d", respBad.StatusCode)
		}
	})
}

func TestMultiFilterEndpoints_AdversarialEdgeCases(t *testing.T) {
	t.Parallel()

	wsDir := setupMultiFilterWorkspaces(t)
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>OK</html>")},
	}
	server := mustNewServer(t, wsDir, mockFS)
	ts := httptest.NewServer(server.Handler())
	t.Cleanup(func() { ts.Close() })

	t.Run("GET /api/fleet equivalence between comma-separated and repeated parameters", func(t *testing.T) {
		t.Parallel()

		// Comma-delimited
		respComma, err := http.Get(ts.URL + "/api/fleet?ticket_type=feat,bug")
		if err != nil {
			t.Fatal(err)
		}
		var summaryComma domain.FleetSummary
		if err := json.NewDecoder(respComma.Body).Decode(&summaryComma); err != nil {
			t.Fatal(err)
		}
		_ = respComma.Body.Close()

		// Repeated parameters
		respRepeated, err := http.Get(ts.URL + "/api/fleet?ticket_type=feat&ticket_type=bug")
		if err != nil {
			t.Fatal(err)
		}
		var summaryRepeated domain.FleetSummary
		if err := json.NewDecoder(respRepeated.Body).Decode(&summaryRepeated); err != nil {
			t.Fatal(err)
		}
		_ = respRepeated.Body.Close()

		if summaryComma.TotalTickets != 2 || summaryRepeated.TotalTickets != 2 {
			t.Fatalf("expected 2 tickets each, got comma=%d repeated=%d", summaryComma.TotalTickets, summaryRepeated.TotalTickets)
		}
		if summaryComma.TotalCostUSD != 3.00 || summaryRepeated.TotalCostUSD != 3.00 {
			t.Errorf("expected cost 3.00 each, got comma=%f repeated=%f", summaryComma.TotalCostUSD, summaryRepeated.TotalCostUSD)
		}

		// Prefix repeated vs comma on fleet
		respPrefixComma, err := http.Get(ts.URL + "/api/fleet?prefix=AL,PO")
		if err != nil {
			t.Fatal(err)
		}
		var prefixComma domain.FleetSummary
		if err := json.NewDecoder(respPrefixComma.Body).Decode(&prefixComma); err != nil {
			t.Fatal(err)
		}
		_ = respPrefixComma.Body.Close()

		respPrefixRepeated, err := http.Get(ts.URL + "/api/fleet?prefix=AL&prefix=PO")
		if err != nil {
			t.Fatal(err)
		}
		var prefixRepeated domain.FleetSummary
		if err := json.NewDecoder(respPrefixRepeated.Body).Decode(&prefixRepeated); err != nil {
			t.Fatal(err)
		}
		_ = respPrefixRepeated.Body.Close()

		if prefixComma.TotalTickets != 3 || prefixRepeated.TotalTickets != 3 {
			t.Fatalf("expected 3 tickets for prefix AL,PO, got comma=%d repeated=%d", prefixComma.TotalTickets, prefixRepeated.TotalTickets)
		}
		if prefixComma.TotalCostUSD != 6.00 || prefixRepeated.TotalCostUSD != 6.00 {
			t.Errorf("expected 6.00 cost, got comma=%f repeated=%f", prefixComma.TotalCostUSD, prefixRepeated.TotalCostUSD)
		}

		// Mixed comma and repeated on fleet
		respMixed, err := http.Get(ts.URL + "/api/fleet?ticket_type=feat,bug&ticket_type=docs")
		if err != nil {
			t.Fatal(err)
		}
		var summaryMixed domain.FleetSummary
		if err := json.NewDecoder(respMixed.Body).Decode(&summaryMixed); err != nil {
			t.Fatal(err)
		}
		_ = respMixed.Body.Close()
		if summaryMixed.TotalTickets != 3 || summaryMixed.TotalCostUSD != 6.00 {
			t.Errorf("expected 3 tickets and 6.00 cost for mixed, got tickets=%d cost=%f", summaryMixed.TotalTickets, summaryMixed.TotalCostUSD)
		}
	})

	t.Run("GET /api/workspaces trailing commas and redundant separators", func(t *testing.T) {
		t.Parallel()

		resp, err := http.Get(ts.URL + "/api/workspaces?ticket_type=feat,,bug,")
		if err != nil {
			t.Fatal(err)
		}
		var ledgers []domain.AgenticMetricsLedger
		if err := json.NewDecoder(resp.Body).Decode(&ledgers); err != nil {
			t.Fatal(err)
		}
		_ = resp.Body.Close()

		if len(ledgers) != 2 {
			t.Fatalf("expected 2 ledgers for feat,,bug,, got %d", len(ledgers))
		}

		// Combined repeated prefixes with comma-separated token
		respMixedPrefix, err := http.Get(ts.URL + "/api/workspaces?prefix=AL,PO&prefix=GEN")
		if err != nil {
			t.Fatal(err)
		}
		var ledgersMixed []domain.AgenticMetricsLedger
		if err := json.NewDecoder(respMixedPrefix.Body).Decode(&ledgersMixed); err != nil {
			t.Fatal(err)
		}
		_ = respMixedPrefix.Body.Close()

		if len(ledgersMixed) != 4 {
			t.Fatalf("expected 4 ledgers for prefix=AL,PO&prefix=GEN, got %d", len(ledgersMixed))
		}
	})

	t.Run("GET /api/comparison comma-separated and repeated equivalence", func(t *testing.T) {
		t.Parallel()

		// Comma-separated
		respComma, err := http.Get(ts.URL + "/api/comparison?ticket_type=feat,bug")
		if err != nil {
			t.Fatal(err)
		}
		var rowsComma []domain.TicketComparisonRow
		if err := json.NewDecoder(respComma.Body).Decode(&rowsComma); err != nil {
			t.Fatal(err)
		}
		_ = respComma.Body.Close()

		// Repeated
		respRepeated, err := http.Get(ts.URL + "/api/comparison?ticket_type=feat&ticket_type=bug")
		if err != nil {
			t.Fatal(err)
		}
		var rowsRepeated []domain.TicketComparisonRow
		if err := json.NewDecoder(respRepeated.Body).Decode(&rowsRepeated); err != nil {
			t.Fatal(err)
		}
		_ = respRepeated.Body.Close()

		if len(rowsComma) != 2 || len(rowsRepeated) != 2 {
			t.Fatalf("expected 2 comparison rows, got comma=%d repeated=%d", len(rowsComma), len(rowsRepeated))
		}

		// Prefix repeated on comparison
		respPrefixRepeated, err := http.Get(ts.URL + "/api/comparison?prefix=AL&prefix=GEN")
		if err != nil {
			t.Fatal(err)
		}
		var rowsPrefix []domain.TicketComparisonRow
		if err := json.NewDecoder(respPrefixRepeated.Body).Decode(&rowsPrefix); err != nil {
			t.Fatal(err)
		}
		_ = respPrefixRepeated.Body.Close()

		if len(rowsPrefix) != 3 {
			t.Fatalf("expected 3 comparison rows for prefix AL and GEN, got %d", len(rowsPrefix))
		}

		// Non-matching filter returns 200 OK with empty slice
		respEmpty, err := http.Get(ts.URL + "/api/comparison?ticket_type=nonexistent_category")
		if err != nil {
			t.Fatal(err)
		}
		if respEmpty.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK for disjoint filter, got %d", respEmpty.StatusCode)
		}
		var rowsEmpty []domain.TicketComparisonRow
		if err := json.NewDecoder(respEmpty.Body).Decode(&rowsEmpty); err != nil {
			t.Fatal(err)
		}
		_ = respEmpty.Body.Close()
		if len(rowsEmpty) != 0 {
			t.Errorf("expected empty comparison rows, got %d", len(rowsEmpty))
		}
	})

	t.Run("GET /api/progression repeated parameters and disjoint queries", func(t *testing.T) {
		t.Parallel()

		// Repeated ticket_type
		respRepeated, err := http.Get(ts.URL + "/api/progression?ticket_type=feat&ticket_type=bug")
		if err != nil {
			t.Fatal(err)
		}
		var bucketsRepeated []domain.TimeBucketMetric
		if err := json.NewDecoder(respRepeated.Body).Decode(&bucketsRepeated); err != nil {
			t.Fatal(err)
		}
		_ = respRepeated.Body.Close()

		if len(bucketsRepeated) != 2 {
			t.Fatalf("expected 2 buckets for repeated ticket_type, got %d", len(bucketsRepeated))
		}

		// Repeated prefix
		respPrefixRepeated, err := http.Get(ts.URL + "/api/progression?prefix=AL&prefix=GEN")
		if err != nil {
			t.Fatal(err)
		}
		var bucketsPrefix []domain.TimeBucketMetric
		if err := json.NewDecoder(respPrefixRepeated.Body).Decode(&bucketsPrefix); err != nil {
			t.Fatal(err)
		}
		_ = respPrefixRepeated.Body.Close()

		if len(bucketsPrefix) != 3 {
			t.Fatalf("expected 3 buckets for repeated prefix, got %d", len(bucketsPrefix))
		}

		// Disjoint query returns empty slice
		respEmpty, err := http.Get(ts.URL + "/api/progression?prefix=NONEXISTENT")
		if err != nil {
			t.Fatal(err)
		}
		if respEmpty.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", respEmpty.StatusCode)
		}
		var bucketsEmpty []domain.TimeBucketMetric
		if err := json.NewDecoder(respEmpty.Body).Decode(&bucketsEmpty); err != nil {
			t.Fatal(err)
		}
		_ = respEmpty.Body.Close()
		if len(bucketsEmpty) != 0 {
			t.Errorf("expected 0 buckets for nonexistent prefix, got %d", len(bucketsEmpty))
		}
	})

	t.Run("path traversal attempts in filter parameters do not panic", func(t *testing.T) {
		t.Parallel()

		traversalQueries := []string{
			"/api/fleet?prefix=../../etc/passwd",
			"/api/workspaces?prefix=../../etc/passwd",
			"/api/comparison?prefix=../../etc/passwd",
			"/api/progression?prefix=../../etc/passwd",
			"/api/workspaces?ticket_type=../../etc/passwd",
		}

		for _, u := range traversalQueries {
			resp, err := http.Get(ts.URL + u)
			if err != nil {
				t.Fatalf("failed query %s: %v", u, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 OK for %s, got %d", u, resp.StatusCode)
			}
			_ = resp.Body.Close()
		}
	})
}

func TestProgressionHandler(t *testing.T) {
	wsDir := setupFixtureWorkspaces(t)
	server := mustNewServer(t, wsDir, fstest.MapFS{})

	t.Run("default progression returns 200 with daily buckets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling progression response: %v", err)
		}

		if len(buckets) == 0 {
			t.Errorf("expected at least one progression bucket")
		}
		// In daily granularity, bucket key format is YYYY-MM-DD
		for _, b := range buckets {
			if len(b.BucketKey) != 10 {
				t.Errorf("expected daily bucket key format YYYY-MM-DD, got %s", b.BucketKey)
			}
		}
	})

	t.Run("hourly granularity produces hourly bucket keys", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?granularity=hourly", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling progression response: %v", err)
		}

		if len(buckets) == 0 {
			t.Errorf("expected at least one hourly progression bucket")
		}
		for _, b := range buckets {
			if !strings.HasSuffix(b.BucketKey, ":00:00Z") {
				t.Errorf("expected hourly bucket key format YYYY-MM-DDTHH:00:00Z, got %s", b.BucketKey)
			}
		}
	})

	t.Run("custom datetime range with YYYY-MM-DDTHH:mm format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?startDate=2026-10-15T09:00&endDate=2026-10-15T12:00&granularity=hourly", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling progression response: %v", err)
		}

		for _, b := range buckets {
			if !strings.HasPrefix(b.BucketKey, "2026-10-15T") {
				t.Errorf("expected bucket in range, got %s", b.BucketKey)
			}
		}
	})

	t.Run("invalid datetime returns 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?startDate=invalid-date", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("invalid granularity returns 400 Bad Request", func(t *testing.T) {
		invalidGranularities := []string{
			"unsupported",
			"bogus",
			"minutely",
			"secondly",
			"yearly",
			"cumulative",
		}
		for _, g := range invalidGranularities {
			req := httptest.NewRequest(http.MethodGet, "/api/progression?granularity="+g, nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 Bad Request for granularity=%s, got %d (body: %s)", g, rec.Code, rec.Body.String())
			}
			var errResp adaptershttp.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
				t.Errorf("expected JSON error response: %v", err)
			}
			if errResp.Error != "bad_request" {
				t.Errorf("expected error code 'bad_request', got %s", errResp.Error)
			}
		}
	})

	t.Run("granularity=cumulative returns 400 Bad Request with proper error structure", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?granularity=cumulative", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d (body: %s)", rec.Code, rec.Body.String())
		}
		var errResp adaptershttp.ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
			t.Fatalf("failed unmarshaling error response: %v", err)
		}
		if errResp.Error != "bad_request" {
			t.Errorf("expected error code 'bad_request', got %q", errResp.Error)
		}
		expectedMsg := `invalid granularity "cumulative": must be hourly, daily, or weekly`
		if errResp.Message != expectedMsg {
			t.Errorf("expected error message %q, got %q", expectedMsg, errResp.Message)
		}
	})

	t.Run("invalid until parameter returns 400 Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?until=not-a-date", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("window=1d filters last 24 hours dynamically", func(t *testing.T) {
		// Create temporary workspace with one recent ticket and one old ticket
		root := t.TempDir()
		wsDir := filepath.Join(root, "workspaces")
		if err := os.MkdirAll(wsDir, 0755); err != nil {
			t.Fatal(err)
		}

		now := time.Now().UTC()
		recentTS := now.Add(-1 * time.Hour).Format(time.RFC3339)
		oldTS := now.Add(-48 * time.Hour).Format(time.RFC3339)

		tRecentDir := filepath.Join(wsDir, "RECENT-1")
		_ = os.MkdirAll(tRecentDir, 0755)
		recentJSON := fmt.Sprintf(`{
  "ticket_id": "RECENT-1",
  "ticket_title": "Recent Ticket",
  "ticket_type": "feature",
  "started_at": "%s",
  "completed_at": "%s",
  "updated_at": "%s",
  "code": {
    "lines_added": 10,
    "lines_deleted": 0,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "inv-rec-1",
      "invocations": [
        {
          "id": "inv-rec-1",
          "role": "coder",
          "model": "gpt-4",
          "user": "dev",
          "started_at": "%s",
          "completed_at": "%s",
          "tokens": {
            "input": 500,
            "output": 100,
            "cache_read": 200,
            "reasoning": 0,
            "peak_context": 2000
          },
          "cost_usd": 0.05
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`, recentTS, recentTS, recentTS, recentTS, recentTS)
		_ = os.WriteFile(filepath.Join(tRecentDir, "agentic_metrics.json"), []byte(recentJSON), 0644)

		tOldDir := filepath.Join(wsDir, "OLD-1")
		_ = os.MkdirAll(tOldDir, 0755)
		oldJSON := fmt.Sprintf(`{
  "ticket_id": "OLD-1",
  "ticket_title": "Old Ticket",
  "ticket_type": "feature",
  "started_at": "%s",
  "completed_at": "%s",
  "updated_at": "%s",
  "code": {
    "lines_added": 10,
    "lines_deleted": 0,
    "files_changed": 1,
    "sdlc_iterations": 1
  },
  "sessions": [
    {
      "id": "inv-old-1",
      "invocations": [
        {
          "id": "inv-old-1",
          "role": "coder",
          "model": "gpt-4",
          "user": "dev",
          "started_at": "%s",
          "completed_at": "%s",
          "tokens": {
            "input": 500,
            "output": 100,
            "cache_read": 200,
            "reasoning": 0,
            "peak_context": 2000
          },
          "cost_usd": 0.05
        }
      ],
      "tools": [],
      "events": []
    }
  ]
}`, oldTS, oldTS, oldTS, oldTS, oldTS)
		_ = os.WriteFile(filepath.Join(tOldDir, "agentic_metrics.json"), []byte(oldJSON), 0644)

		s := mustNewServer(t, wsDir, fstest.MapFS{})

		req := httptest.NewRequest(http.MethodGet, "/api/progression?window=1d&granularity=hourly", nil)
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling progression response: %v", err)
		}

		if len(buckets) != 1 {
			t.Fatalf("expected exactly 1 bucket within last 24h, got %d", len(buckets))
		}
		if buckets[0].Invocations != 1 {
			t.Errorf("expected 1 invocation, got %d", buckets[0].Invocations)
		}
		if buckets[0].CostUSD != 0.05 {
			t.Errorf("expected 0.05 cost, got %f", buckets[0].CostUSD)
		}
	})

	t.Run("inverted date range returns 200 OK with empty slice", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/progression?startDate=2026-10-15T18:00&endDate=2026-10-15T09:00", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling progression response: %v", err)
		}
		if len(buckets) != 0 {
			t.Fatalf("expected 0 buckets for inverted dates, got %d", len(buckets))
		}
	})

	t.Run("empty workspace returns 200 OK with empty slice", func(t *testing.T) {
		emptyWs := t.TempDir()
		s := mustNewServer(t, emptyWs, fstest.MapFS{})
		req := httptest.NewRequest(http.MethodGet, "/api/progression", nil)
		rec := httptest.NewRecorder()
		s.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var buckets []domain.TimeBucketMetric
		if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
			t.Fatalf("failed unmarshaling empty progression response: %v", err)
		}
		if len(buckets) != 0 {
			t.Errorf("expected 0 buckets for empty workspaces, got %d", len(buckets))
		}
	})

	t.Run("fleet and workspaces handlers support window=1d", func(t *testing.T) {
		for _, path := range []string{"/api/fleet?window=1d", "/api/workspaces?window=1d"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("expected 200 OK for %s, got %d", path, rec.Code)
			}
		}
	})
}

func TestProgressionHandler_AL34_Adversarial(t *testing.T) {
	tempDir := t.TempDir()

	// Seed workspace with authentic multi-turn records
	wsDir := filepath.Join(tempDir, "AL-TEST-34")
	if err := os.MkdirAll(wsDir, 0755); err != nil {
		t.Fatalf("failed creating test workspace: %v", err)
	}

	rawJSON := `{
  "ticket_id": "AL-TEST-34",
  "ticket_type": "feature",
  "started_at": "2026-10-15T09:00:00Z",
  "completed_at": "2026-10-15T18:00:00Z",
  "updated_at": "2026-10-15T18:00:00Z",
  "code": {
    "lines_added": 250,
    "lines_deleted": 75,
    "files_changed": 5
  },
  "sessions": [{
    "id": "inv-1",
    "invocations": [
      {
        "id": "inv-1",
        "started_at": "2026-10-15T09:00:00Z",
        "completed_at": "2026-10-15T18:00:00Z",
        "tokens": {
          "input": 1500,
          "output": 350,
          "cache_read": 1200,
          "reasoning": 80
        },
        "cost_usd": 0.05
      }
    ],
    "tools": [],
    "events": []
  }]
}`
	if err := os.WriteFile(filepath.Join(wsDir, "agentic_metrics.json"), []byte(rawJSON), 0644); err != nil {
		t.Fatalf("failed writing test metrics: %v", err)
	}

	server := mustNewServer(t, tempDir, fstest.MapFS{})

	t.Run("hostile granularity inputs rejecting cumulative and arbitrary strings with HTTP 400", func(t *testing.T) {
		hostileGranularities := []string{
			"cumulative",
			"CUMULATIVE",
			"Cumulative",
			"cUmUlAtIvE",
			"  cumulative  ",
			"%20%20cumulative%20%20",
			"yearly",
			"minutely",
			"secondly",
			"monthly",
			"decade",
			"all",
			"none",
			"../../etc/passwd",
			"%00nullbyte",
			"<script>alert(1)</script>",
			"hourly;DROP TABLE;",
		}

		for _, g := range hostileGranularities {
			req := httptest.NewRequest(http.MethodGet, "/api/progression?granularity="+url.QueryEscape(g), nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("granularity %q: expected 400 Bad Request, got %d (body: %s)", g, rec.Code, rec.Body.String())
				continue
			}

			var errResp adaptershttp.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
				t.Errorf("granularity %q: failed unmarshaling error response: %v", g, err)
				continue
			}
			if errResp.Error != "bad_request" {
				t.Errorf("granularity %q: expected error code 'bad_request', got %q", g, errResp.Error)
			}
			if !strings.Contains(errResp.Message, "must be hourly, daily, or weekly") {
				t.Errorf("granularity %q: expected message to state must be hourly, daily, or weekly, got %q", g, errResp.Message)
			}
		}
	})

	t.Run("valid granularities return 200 OK and CumulativeNetLinesAdded is correctly calculated", func(t *testing.T) {
		validGranularities := []string{"hourly", "daily", "weekly", ""}

		for _, g := range validGranularities {
			url := "/api/progression"
			if g != "" {
				url += "?granularity=" + g
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200 OK for %s, got %d (body: %s)", url, rec.Code, rec.Body.String())
			}

			var buckets []domain.TimeBucketMetric
			if err := json.Unmarshal(rec.Body.Bytes(), &buckets); err != nil {
				t.Fatalf("failed unmarshaling progression response for %s: %v", url, err)
			}

			if len(buckets) != 1 {
				t.Fatalf("expected 1 bucket for %s, got %d", url, len(buckets))
			}

			b := buckets[0]
			if b.LinesAdded != 250 {
				t.Errorf("[%s] expected linesAdded 250, got %d", url, b.LinesAdded)
			}
			if b.LinesDeleted != 75 {
				t.Errorf("[%s] expected linesDeleted 75, got %d", url, b.LinesDeleted)
			}
			if b.NetLinesAdded != 175 {
				t.Errorf("[%s] expected NetLinesAdded 175, got %d", url, b.NetLinesAdded)
			}
			if b.CumulativeNetLinesAdded != 175 {
				t.Errorf("[%s] expected CumulativeNetLinesAdded 175, got %d", url, b.CumulativeNetLinesAdded)
			}
			if b.CumulativeNetLinesAdded != b.CumulativeLinesAdded-b.CumulativeLinesDeleted {
				t.Errorf("[%s] CumulativeNetLinesAdded mismatch: %d != %d - %d",
					url, b.CumulativeNetLinesAdded, b.CumulativeLinesAdded, b.CumulativeLinesDeleted)
			}
		}
	})
}

func TestConfigHandler_DefaultConfig(t *testing.T) {
	wsDir := setupFixtureWorkspaces(t)
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Mock SPA</body></html>")},
	}
	server := mustNewServer(t, wsDir, staticFS)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var dto adaptershttp.ProjectConfigDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("failed unmarshaling config response: %v", err)
	}

	if dto.ProjectName != "AgenticLens" {
		t.Errorf("expected default project_name 'AgenticLens', got %q", dto.ProjectName)
	}

	if dto.WorkspacesDir != wsDir {
		t.Errorf("expected workspaces_dir %q, got %q", wsDir, dto.WorkspacesDir)
	}

	expectedBuildInfo := domain.GetBuildInfo()
	if dto.Version != expectedBuildInfo.Version {
		t.Errorf("expected version %q, got %q", expectedBuildInfo.Version, dto.Version)
	}

	if len(dto.Prefixes) < 5 {
		t.Fatalf("expected at least 5 configured prefixes, got %d", len(dto.Prefixes))
	}
	prefixMap := make(map[string]domain.TicketPrefix)
	for _, p := range dto.Prefixes {
		prefixMap[p.Prefix] = p
	}
	for _, expectedPrefix := range []string{"AL", "AL2", "GEN", "PO", "FIX"} {
		if _, ok := prefixMap[expectedPrefix]; !ok {
			t.Errorf("expected configured prefix %s to be present", expectedPrefix)
		}
	}
	if gen, ok := prefixMap["GEN"]; !ok || gen.TicketType != "chore" {
		t.Errorf("expected prefix GEN with chore ticket type, got %+v", gen)
	}
	if po, ok := prefixMap["PO"]; !ok || po.TicketType != "docs" {
		t.Errorf("expected prefix PO with docs ticket type, got %+v", po)
	}
	if fix, ok := prefixMap["FIX"]; !ok || fix.TicketType != "fix" {
		t.Errorf("expected prefix FIX with fix ticket type, got %+v", fix)
	}

	if len(dto.RoleOrder) != 5 {
		t.Fatalf("expected 5 canonical roles, got %d", len(dto.RoleOrder))
	}
	expectedRoles := []string{"orchestrator", "planner", "coder", "tester", "reviewer"}
	for i, r := range expectedRoles {
		if dto.RoleOrder[i] != r {
			t.Errorf("expected role at index %d to be %q, got %q", i, r, dto.RoleOrder[i])
		}
	}
}

func TestConfigHandler_CustomConfig(t *testing.T) {
	wsDir := setupFixtureWorkspaces(t)
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Mock SPA</body></html>")},
	}

	customCfg := domain.ProjectConfig{
		ProjectName: "CustomObservabilityRepo",
		Ticket: domain.TicketConfig{
			Prefixes: []domain.TicketPrefix{
				{
					Prefix:      "FEAT",
					Description: "Custom feature tickets",
					TicketType:  "feat",
				},
				{
					Prefix:      "BUG",
					Description: "Bug fixes",
					TicketType:  "fix",
				},
			},
		},
		Workspaces: domain.WorkspacesConfig{
			Dir: "custom_workspaces",
		},
		RoleOrder: []string{"planner", "coder", "reviewer"},
	}

	server := adaptershttp.NewServerWithConfig(wsDir, staticFS, customCfg)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var dto adaptershttp.ProjectConfigDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("failed unmarshaling config response: %v", err)
	}

	if dto.ProjectName != "CustomObservabilityRepo" {
		t.Errorf("expected project_name 'CustomObservabilityRepo', got %q", dto.ProjectName)
	}

	if dto.WorkspacesDir != wsDir {
		t.Errorf("expected workspaces_dir %q, got %q", wsDir, dto.WorkspacesDir)
	}

	if len(dto.Prefixes) != 2 {
		t.Fatalf("expected 2 custom prefixes, got %d", len(dto.Prefixes))
	}
	if dto.Prefixes[0].Prefix != "FEAT" || dto.Prefixes[0].TicketType != "feat" {
		t.Errorf("expected FEAT prefix, got %+v", dto.Prefixes[0])
	}
	if dto.Prefixes[1].Prefix != "BUG" || dto.Prefixes[1].TicketType != "fix" {
		t.Errorf("expected BUG prefix, got %+v", dto.Prefixes[1])
	}

	if len(dto.RoleOrder) != 3 {
		t.Fatalf("expected 3 custom roles, got %d", len(dto.RoleOrder))
	}
	if dto.RoleOrder[0] != "planner" || dto.RoleOrder[1] != "coder" || dto.RoleOrder[2] != "reviewer" {
		t.Errorf("unexpected role order: %+v", dto.RoleOrder)
	}
}

func TestConfigHandler_EmptyConfigPreserved(t *testing.T) {
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Mock SPA</body></html>")},
	}

	emptyCfg := domain.ProjectConfig{
		ProjectName: "",
		Ticket: domain.TicketConfig{
			Prefixes: nil,
		},
		Workspaces: domain.WorkspacesConfig{
			Dir: "",
		},
		RoleOrder: nil,
	}

	server := adaptershttp.NewServerWithConfig("", staticFS, emptyCfg)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var dto adaptershttp.ProjectConfigDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("failed unmarshaling config response: %v", err)
	}

	if dto.ProjectName != "" {
		t.Errorf("expected empty project_name, got %q", dto.ProjectName)
	}

	if dto.WorkspacesDir != "" {
		t.Errorf("expected empty workspaces_dir, got %q", dto.WorkspacesDir)
	}

	if len(dto.Prefixes) != 0 || dto.Prefixes == nil {
		t.Errorf("expected non-nil empty slice for prefixes, got len %d (nil: %v)", len(dto.Prefixes), dto.Prefixes == nil)
	}

	if len(dto.RoleOrder) != 0 || dto.RoleOrder == nil {
		t.Errorf("expected non-nil empty slice for role order, got len %d (nil: %v)", len(dto.RoleOrder), dto.RoleOrder == nil)
	}

	rawJSON := rec.Body.String()
	if !strings.Contains(rawJSON, `"prefixes":[]`) {
		t.Errorf("expected raw JSON to contain '\"prefixes\":[]', got %s", rawJSON)
	}
	if !strings.Contains(rawJSON, `"role_order":[]`) {
		t.Errorf("expected raw JSON to contain '\"role_order\":[]', got %s", rawJSON)
	}
	if strings.Contains(rawJSON, `"prefixes":null`) {
		t.Errorf("raw JSON must not contain '\"prefixes\":null', got %s", rawJSON)
	}
	if strings.Contains(rawJSON, `"role_order":null`) {
		t.Errorf("raw JSON must not contain '\"role_order\":null', got %s", rawJSON)
	}
}

func TestConfigHandler_DefaultProjectConfig(t *testing.T) {
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Mock SPA</body></html>")},
	}
	server := adaptershttp.NewServerWithConfig("workspaces", staticFS, domain.DefaultProjectConfig())

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var dto adaptershttp.ProjectConfigDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
		t.Fatalf("failed unmarshaling /api/config: %v", err)
	}

	if len(dto.RoleOrder) != 0 || dto.RoleOrder == nil {
		t.Errorf("expected empty non-nil role_order, got len %d (nil: %v)", len(dto.RoleOrder), dto.RoleOrder == nil)
	}

	rawJSON := rec.Body.String()
	if !strings.Contains(rawJSON, `"role_order":[]`) {
		t.Errorf("expected raw JSON to contain '\"role_order\":[]', got %s", rawJSON)
	}
	if strings.Contains(rawJSON, `"role_order":null`) {
		t.Errorf("raw JSON must not contain '\"role_order\":null', got %s", rawJSON)
	}
}

func TestConfigHandler_AdversarialEdgeCases(t *testing.T) {
	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>Mock SPA</body></html>")},
	}

	t.Run("whitespace-only fields trim cleanly without panics", func(t *testing.T) {
		hostileCfg := domain.ProjectConfig{
			ProjectName: "   \t\r\n   ",
			Ticket: domain.TicketConfig{
				Prefixes: nil,
			},
			Workspaces: domain.WorkspacesConfig{
				Dir: "   \n\t   ",
			},
			RoleOrder: nil,
		}

		server := adaptershttp.NewServerWithConfig("", staticFS, hostileCfg)

		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var dto adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
			t.Fatalf("failed unmarshaling config response: %v", err)
		}

		if dto.ProjectName != "" {
			t.Errorf("expected trimmed whitespace project_name to be empty string, got %q", dto.ProjectName)
		}
		if dto.WorkspacesDir != "" {
			t.Errorf("expected trimmed whitespace workspaces_dir to be empty string, got %q", dto.WorkspacesDir)
		}

		rawJSON := rec.Body.String()
		if !strings.Contains(rawJSON, `"prefixes":[]`) {
			t.Errorf("expected JSON output to contain '\"prefixes\":[]', got: %s", rawJSON)
		}
		if !strings.Contains(rawJSON, `"role_order":[]`) {
			t.Errorf("expected JSON output to contain '\"role_order\":[]', got: %s", rawJSON)
		}
		if strings.Contains(rawJSON, `"prefixes":null`) {
			t.Errorf("JSON output must never contain 'prefixes:null', got: %s", rawJSON)
		}
		if strings.Contains(rawJSON, `"role_order":null`) {
			t.Errorf("JSON output must never contain 'role_order:null', got: %s", rawJSON)
		}
	})

	t.Run("non-GET HTTP methods return method not allowed or not found", func(t *testing.T) {
		server := mustNewServer(t, "workspaces", staticFS)

		disallowedMethods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
		for _, method := range disallowedMethods {
			req := httptest.NewRequest(method, "/api/config", nil)
			rec := httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
				t.Errorf("expected status 405 Method Not Allowed or 404 Not Found for %s /api/config, got %d", method, rec.Code)
			}
		}
	})
}

func TestServerWithTestDataWorkspaces(t *testing.T) {
	testdataDir := filepath.Join("..", "..", "..", "..", "testdata", "workspaces")
	if _, err := os.Stat(testdataDir); err != nil {
		t.Fatalf("testdata/workspaces directory not found at %s: %v", testdataDir, err)
	}

	server := mustNewServer(t, testdataDir, fstest.MapFS{})
	handler := server.Handler()

	t.Run("discovers all 5 test workspaces via /api/workspaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.Unmarshal(rec.Body.Bytes(), &ledgers); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces: %v", err)
		}

		if len(ledgers) != 5 {
			t.Fatalf("expected 5 ledgers from testdata/workspaces, got %d", len(ledgers))
		}

		ticketIDs := make(map[string]bool)
		for _, l := range ledgers {
			if l.TicketID != nil {
				ticketIDs[*l.TicketID] = true
			}
		}

		expectedTickets := []string{"AL-1", "AL-2", "AL-3", "PO-1", "GEN-1"}
		for _, id := range expectedTickets {
			if !ticketIDs[id] {
				t.Errorf("expected workspace %s to be discovered, but was missing", id)
			}
		}
	})

	t.Run("aggregates fleet summary from testdata/workspaces via /api/fleet", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/fleet", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var summary domain.FleetSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &summary); err != nil {
			t.Fatalf("failed unmarshaling /api/fleet: %v", err)
		}

		if summary.TotalTickets != 5 {
			t.Errorf("expected 5 total tickets in fleet summary, got %d", summary.TotalTickets)
		}
	})

	t.Run("detects TOOL_RETRY_STREAK anomaly on AL-2 workspace detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-2", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces/AL-2: %v", err)
		}

		found := false
		for _, a := range detail.Anomalies {
			if a.RuleID == "TOOL_RETRY_STREAK" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected TOOL_RETRY_STREAK anomaly in AL-2 detail, found anomalies: %+v", detail.Anomalies)
		}
	})

	t.Run("detects CONTEXT_SATURATION_SPIKE anomaly on GEN-1 workspace detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/GEN-1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces/GEN-1: %v", err)
		}

		found := false
		for _, a := range detail.Anomalies {
			if a.RuleID == "CONTEXT_SATURATION_SPIKE" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected CONTEXT_SATURATION_SPIKE anomaly in GEN-1 detail, found anomalies: %+v", detail.Anomalies)
		}
	})

	t.Run("serves markdown artifacts for AL-1 workspace", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var detail adaptershttp.WorkspaceDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces/AL-1: %v", err)
		}

		if detail.Artifacts.ContractContent == nil || !strings.Contains(*detail.Artifacts.ContractContent, "AL-1") {
			t.Errorf("expected contract content to contain 'AL-1'")
		}
		if detail.Artifacts.SummaryContent == nil || !strings.Contains(*detail.Artifacts.SummaryContent, "AL-1") {
			t.Errorf("expected summary content to contain 'AL-1'")
		}
	})

	t.Run("reports testdata workspaces dir via /api/config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var cfg adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
			t.Fatalf("failed unmarshaling /api/config: %v", err)
		}

		if !strings.Contains(cfg.WorkspacesDir, "testdata/workspaces") && !strings.Contains(cfg.WorkspacesDir, filepath.Join("testdata", "workspaces")) {
			t.Errorf("expected workspaces_dir to contain testdata/workspaces, got: %s", cfg.WorkspacesDir)
		}
	})

	t.Run("returns 404 for tickets not present in testdata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-10", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404 Not Found for AL-10 in testdata workspaces, got %d", rec.Code)
		}
	})
}

func TestServerWithProductionWorkspaces_DoesNotTouchTestdataWorkspaces(t *testing.T) {
	parityWorkspacesDir := filepath.Join("..", "..", "..", "..", "testdata", "parity_workspaces")
	if _, currentFile, _, ok := runtime.Caller(0); ok {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "testdata", "parity_workspaces"))
		if fi, err := os.Stat(candidate); err == nil && fi.IsDir() {
			parityWorkspacesDir = candidate
		}
	}
	if _, err := os.Stat(parityWorkspacesDir); err != nil {
		t.Fatalf("parity workspaces directory not found at %s: %v", parityWorkspacesDir, err)
	}

	server := mustNewServer(t, parityWorkspacesDir, fstest.MapFS{})
	handler := server.Handler()

	t.Run("reports parity workspaces dir and does not reference testdata/workspaces in /api/config", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var cfg adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
			t.Fatalf("failed unmarshaling /api/config: %v", err)
		}

		if strings.HasSuffix(cfg.WorkspacesDir, "testdata/workspaces") || strings.HasSuffix(cfg.WorkspacesDir, filepath.Join("testdata", "workspaces")) {
			t.Errorf("expected workspaces_dir to not be testdata/workspaces, got: %s", cfg.WorkspacesDir)
		}
		if !strings.Contains(cfg.WorkspacesDir, "parity_workspaces") {
			t.Errorf("expected workspaces_dir to contain parity_workspaces, got: %s", cfg.WorkspacesDir)
		}
	})

	t.Run("serves parity workspaces exclusively without testdata contamination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", rec.Code)
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.Unmarshal(rec.Body.Bytes(), &ledgers); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces: %v", err)
		}

		if len(ledgers) <= 5 {
			t.Fatalf("expected > 5 ledgers from parity workspaces, got %d", len(ledgers))
		}

		var al1Ledger *domain.AgenticMetricsLedger
		for i := range ledgers {
			if ledgers[i].TicketID != nil && *ledgers[i].TicketID == "AL-1" {
				al1Ledger = &ledgers[i]
				break
			}
		}

		if al1Ledger == nil {
			t.Fatal("expected AL-1 in parity workspaces")
		}

		testdataTitle := "Core Go domain models and aggregations engine"
		if al1Ledger.TicketTitle != nil && *al1Ledger.TicketTitle == testdataTitle {
			t.Errorf("expected parity AL-1 title, but got testdata title: %s", *al1Ledger.TicketTitle)
		}

		foundAL10 := false
		for _, l := range ledgers {
			if l.TicketID != nil && *l.TicketID == "AL-10" {
				foundAL10 = true
				break
			}
		}
		if !foundAL10 {
			t.Errorf("expected AL-10 to be discovered in parity workspaces")
		}
	})

	t.Run("serves AL-10 workspace detail successfully from parity", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces/AL-10", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200 OK for AL-10 in parity workspaces, got %d", rec.Code)
		}
	})
}

func TestServerWorkspaceDirConfigSynchronization(t *testing.T) {
	tempDir := t.TempDir()

	cfgSpaceDir := filepath.Join(tempDir, "config_spaces")
	if err := os.MkdirAll(filepath.Join(cfgSpaceDir, "CFG-1"), 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}
	sampleCFG := sampleTestLedgerJSON("CFG-1", "Config Feature", 10, 0.10, 1)
	if err := os.WriteFile(filepath.Join(cfgSpaceDir, "CFG-1", "agentic_metrics.json"), []byte(sampleCFG), 0644); err != nil {
		t.Fatalf("failed writing ledger: %v", err)
	}

	overrideSpaceDir := filepath.Join(tempDir, "override_spaces")
	if err := os.MkdirAll(filepath.Join(overrideSpaceDir, "OVR-1"), 0755); err != nil {
		t.Fatalf("failed creating dir: %v", err)
	}
	sampleOVR := sampleTestLedgerJSON("OVR-1", "Override Feature", 20, 0.20, 1)
	if err := os.WriteFile(filepath.Join(overrideSpaceDir, "OVR-1", "agentic_metrics.json"), []byte(sampleOVR), 0644); err != nil {
		t.Fatalf("failed writing ledger: %v", err)
	}

	t.Run("resolves workspaces dir from ProjectConfig when workspacesDir argument is empty", func(t *testing.T) {
		cfg := domain.ProjectConfig{
			ProjectName: "AutoResolveTest",
			Workspaces: domain.WorkspacesConfig{
				Dir: cfgSpaceDir,
			},
		}

		server := adaptershttp.NewServerWithConfig("", fstest.MapFS{}, cfg)
		handler := server.Handler()

		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", rec.Code)
		}

		var dto adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
			t.Fatalf("failed unmarshaling /api/config: %v", err)
		}

		if dto.WorkspacesDir != cfgSpaceDir {
			t.Errorf("expected workspaces_dir %q, got %q", cfgSpaceDir, dto.WorkspacesDir)
		}

		wsReq := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		wsRec := httptest.NewRecorder()
		handler.ServeHTTP(wsRec, wsReq)

		if wsRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/workspaces, got %d", wsRec.Code)
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.Unmarshal(wsRec.Body.Bytes(), &ledgers); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces: %v", err)
		}

		if len(ledgers) != 1 || *ledgers[0].TicketID != "CFG-1" {
			t.Errorf("expected 1 ledger CFG-1, got %v", ledgers)
		}
	})

	t.Run("prioritizes explicit workspacesDir argument over ProjectConfig", func(t *testing.T) {
		cfg := domain.ProjectConfig{
			ProjectName: "OverrideTest",
			Workspaces: domain.WorkspacesConfig{
				Dir: cfgSpaceDir,
			},
		}

		server := adaptershttp.NewServerWithConfig(overrideSpaceDir, fstest.MapFS{}, cfg)
		handler := server.Handler()

		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", rec.Code)
		}

		var dto adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
			t.Fatalf("failed unmarshaling /api/config: %v", err)
		}

		if dto.WorkspacesDir != overrideSpaceDir {
			t.Errorf("expected workspaces_dir %q, got %q", overrideSpaceDir, dto.WorkspacesDir)
		}

		wsReq := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		wsRec := httptest.NewRecorder()
		handler.ServeHTTP(wsRec, wsReq)

		if wsRec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/workspaces, got %d", wsRec.Code)
		}

		var ledgers []domain.AgenticMetricsLedger
		if err := json.Unmarshal(wsRec.Body.Bytes(), &ledgers); err != nil {
			t.Fatalf("failed unmarshaling /api/workspaces: %v", err)
		}

		if len(ledgers) != 1 || *ledgers[0].TicketID != "OVR-1" {
			t.Errorf("expected 1 ledger OVR-1, got %v", ledgers)
		}
	})

	t.Run("NewServer constructor synchronizes explicit workspacesDir to config", func(t *testing.T) {
		server := mustNewServer(t, overrideSpaceDir, fstest.MapFS{})
		handler := server.Handler()

		req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK from /api/config, got %d", rec.Code)
		}

		var dto adaptershttp.ProjectConfigDTO
		if err := json.Unmarshal(rec.Body.Bytes(), &dto); err != nil {
			t.Fatalf("failed unmarshaling /api/config: %v", err)
		}

		if dto.WorkspacesDir != overrideSpaceDir {
			t.Errorf("expected workspaces_dir %q, got %q", overrideSpaceDir, dto.WorkspacesDir)
		}
	})
}

func TestNewServer_MissingConfigFailsFast(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	wsDir := filepath.Join(tempDir, "workspaces")
	srv, err := adaptershttp.NewServer(wsDir, fstest.MapFS{})
	if srv != nil {
		t.Fatalf("expected nil server when config is missing, got: %+v", srv)
	}
	if err == nil {
		t.Fatal("expected non-nil error when config is missing, got nil")
	}
	if !errors.Is(err, config.ErrConfigNotFound) {
		t.Fatalf("expected error wrapping ErrConfigNotFound, got: %v", err)
	}
}
