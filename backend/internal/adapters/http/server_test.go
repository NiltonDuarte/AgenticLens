package http_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

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
