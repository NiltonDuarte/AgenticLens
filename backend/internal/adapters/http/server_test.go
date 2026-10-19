package http_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
	"agenticlens/internal/domain"
)

func sampleTestLedgerJSON(ticketID, title string, duration int64, cost float64, invocations int64) string {
	return `{
		"ticket_id": "` + ticketID + `",
		"ticket_title": "` + title + `",
		"ticket_type": "feature",
		"started_at": "2026-10-17T10:00:00Z",
		"completed_at": "2026-10-17T11:00:00Z",
		"updated_at": "2026-10-17T11:00:00Z",
		"code": {
			"lines_added": 120,
			"lines_deleted": 20,
			"test_lines_added": 60,
			"test_lines_deleted": 0,
			"prod_lines_added": 60,
			"prod_lines_deleted": 20,
			"files_changed": 4,
			"sdlc_iterations": 2
		},
		"invocations": [
			{
				"id": "ses-orch",
				"role": "orchestrator",
				"model": "claude-3-sonnet",
				"user": "tester",
				"active_duration_seconds": ` + string(rune(duration)) + `,
				"turns": 6,
				"tokens": {
					"input": 5000,
					"output": 1000,
					"cache_read": 10000,
					"reasoning": 300,
					"peak_context": 80000
				},
				"cost_usd": ` + "0.5" + `,
				"compactions": 0
			}
		],
		"tools": [
			{
				"name": "bash",
				"calls": 5,
				"errors": 0
			}
		]
	}`
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
		"code": {"lines_added": 50, "lines_deleted": 10, "test_lines_added": 20, "test_lines_deleted": 0, "prod_lines_added": 30, "prod_lines_deleted": 10, "files_changed": 2, "sdlc_iterations": 1},
		"invocations": [
			{
				"id": "ses-1",
				"role": "orchestrator",
				"model": "gpt-4",
				"user": "dev",
				"active_duration_seconds": 60,
				"turns": 2,
				"tokens": {"input": 1000, "output": 200, "cache_read": 500, "reasoning": 0, "peak_context": 50000},
				"cost_usd": 0.10,
				"compactions": 0
			}
		],
		"tools": []
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
		"code": {"lines_added": 100, "lines_deleted": 20, "test_lines_added": 50, "test_lines_deleted": 0, "prod_lines_added": 50, "prod_lines_deleted": 20, "files_changed": 3, "sdlc_iterations": 1},
		"invocations": [
			{
				"id": "ses-2",
				"role": "orchestrator",
				"model": "gpt-4",
				"user": "dev",
				"active_duration_seconds": 120,
				"turns": 4,
				"tokens": {"input": 2000, "output": 400, "cache_read": 1000, "reasoning": 100, "peak_context": 60000},
				"cost_usd": 0.50,
				"compactions": 0
			}
		],
		"tools": []
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

	server := adaptershttp.NewServer(wsDir, mockFS)
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
		server := adaptershttp.NewServer(wsDir, mockFS)
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
		srv := adaptershttp.NewServer(emptyDir, mockFS)
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
			"invocations": [
				{
					"id": "ses-0",
					"role": "orchestrator",
					"active_duration_seconds": 0,
					"turns": 0,
					"tokens": {"input": 0, "output": 0, "cache_read": 0, "reasoning": 0, "peak_context": 0},
					"cost_usd": 0.0
				}
			]
		}`
		if err := os.WriteFile(filepath.Join(ticketDir, "agentic_metrics.json"), []byte(zeroJSON), 0644); err != nil {
			t.Fatal(err)
		}

		srv := adaptershttp.NewServer(zeroDir, mockFS)
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
		events := []domain.EventFact{
			{
				SessionID: "ses-1",
				Timestamp: "2026-10-18T10:00:00Z",
				Turn:      1,
				Role:      "coder",
				EventType: domain.EventTypeToolCall,
				Status:    "success",
			},
		}
		adaptershttp.WriteEventsResponse(rec, events)
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		var parsed []domain.EventFact
		if err := json.NewDecoder(rec.Body).Decode(&parsed); err != nil {
			t.Fatalf("failed decoding events response: %v", err)
		}
		if len(parsed) != 1 || parsed[0].SessionID != "ses-1" {
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
		"code": {"lines_added": 10, "lines_deleted": 0, "test_lines_added": 5, "test_lines_deleted": 0, "prod_lines_added": 5, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
		"invocations": [
			{
				"id": "ses-1",
				"role": "orchestrator",
				"model": "gpt-4",
				"user": "dev",
				"active_duration_seconds": 60,
				"turns": 5,
				"tokens": {"input": 1000, "output": 200, "cache_read": 500, "reasoning": 0, "peak_context": 50000},
				"cost_usd": 0.10,
				"compactions": 0
			}
		],
		"tools": [],
		"events": [
			{
				"session_id": "ses-1",
				"timestamp": "2026-10-18T10:00:00Z",
				"turn": 1,
				"role": "coder",
				"event_type": "tool_call",
				"tool_name": "bash",
				"status": "error"
			},
			{
				"session_id": "ses-1",
				"timestamp": "2026-10-18T10:00:05Z",
				"turn": 1,
				"role": "coder",
				"event_type": "tool_call",
				"tool_name": "bash",
				"status": "error"
			},
			{
				"session_id": "ses-1",
				"timestamp": "2026-10-18T10:00:10Z",
				"turn": 1,
				"role": "coder",
				"event_type": "tool_call",
				"tool_name": "bash",
				"status": "error"
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
		"code": {"lines_added": 5, "lines_deleted": 0, "test_lines_added": 0, "test_lines_deleted": 0, "prod_lines_added": 5, "prod_lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
		"invocations": [],
		"tools": []
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
		"events": [
			{"session_id": "ses-late", "timestamp": "2026-10-18T12:00:00Z", "turn": 2, "role": "coder", "event_type": "tool_call", "status": "success"},
			{"session_id": "ses-early-t2", "timestamp": "2026-10-18T10:00:00Z", "turn": 2, "role": "coder", "event_type": "tool_call", "status": "success"},
			{"session_id": "ses-early-t1", "timestamp": "2026-10-18T10:00:00Z", "turn": 1, "role": "orchestrator", "event_type": "turn_start", "status": "success"}
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
	server := adaptershttp.NewServer(wsDir, mockFS)
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

		var events []domain.EventFact
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			t.Fatalf("failed decoding events: %v", err)
		}
		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}
		if events[0].Timestamp != "2026-10-18T10:00:00Z" || events[1].Timestamp != "2026-10-18T10:00:05Z" || events[2].Timestamp != "2026-10-18T10:00:10Z" {
			t.Errorf("unexpected events order or content: %+v", events)
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

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed reading body: %v", err)
		}
		if strings.TrimSpace(string(body)) != "[]" {
			t.Fatalf("expected empty JSON array [], got: %s", string(body))
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

		var events []domain.EventFact
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			t.Fatalf("failed decoding events: %v", err)
		}
		if len(events) != 3 {
			t.Fatalf("expected 3 events, got %d", len(events))
		}
		if events[0].SessionID != "ses-early-t1" || events[1].SessionID != "ses-early-t2" || events[2].SessionID != "ses-late" {
			t.Errorf("expected sorted events [ses-early-t1, ses-early-t2, ses-late], got [%s, %s, %s]",
				events[0].SessionID, events[1].SessionID, events[2].SessionID)
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
		"events": [
			{"session_id": "ses-2", "timestamp": "2026-10-18T10:00:10Z", "turn": 2},
			{"session_id": "ses-1", "timestamp": "2026-10-18T10:00:05Z", "turn": 1}
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

		body := strings.TrimSpace(rec.Body.String())
		if body != "[]" {
			t.Errorf("expected empty array '[]', got %q", body)
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

		var events []domain.EventFact
		if err := json.Unmarshal(rec.Body.Bytes(), &events); err != nil {
			t.Fatalf("failed unmarshaling events: %v", err)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if events[0].SessionID != "ses-1" || events[1].SessionID != "ses-2" {
			t.Errorf("expected events stably sorted: %+v", events)
		}
	})
}

func TestProgressionHandler(t *testing.T) {
	wsDir := setupFixtureWorkspaces(t)
	server := adaptershttp.NewServer(wsDir, fstest.MapFS{})

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
			"code": {"lines_added": 10, "lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
			"invocations": [
				{
					"id": "inv-rec-1",
					"role": "coder",
					"model": "gpt-4",
					"user": "dev",
					"started_at": "%s",
					"completed_at": "%s",
					"tokens": {"input": 500, "output": 100, "cache_read": 200, "reasoning": 0, "peak_context": 2000},
					"cost_usd": 0.05
				}
			],
			"tools": []
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
			"code": {"lines_added": 10, "lines_deleted": 0, "files_changed": 1, "sdlc_iterations": 1},
			"invocations": [
				{
					"id": "inv-old-1",
					"role": "coder",
					"model": "gpt-4",
					"user": "dev",
					"started_at": "%s",
					"completed_at": "%s",
					"tokens": {"input": 500, "output": 100, "cache_read": 200, "reasoning": 0, "peak_context": 2000},
					"cost_usd": 0.05
				}
			],
			"tools": []
		}`, oldTS, oldTS, oldTS, oldTS, oldTS)
		_ = os.WriteFile(filepath.Join(tOldDir, "agentic_metrics.json"), []byte(oldJSON), 0644)

		s := adaptershttp.NewServer(wsDir, fstest.MapFS{})

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
		s := adaptershttp.NewServer(emptyWs, fstest.MapFS{})
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
