package http

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

func parseDateBoundary(param string, isEnd bool) (*time.Time, error) {
	clean := strings.TrimSpace(param)
	if clean == "" {
		return nil, nil
	}
	if len(clean) == 10 {
		t, err := time.Parse("2006-01-02", clean)
		if err != nil {
			return nil, fmt.Errorf("invalid date format %q: expected YYYY-MM-DD or RFC3339", param)
		}
		if isEnd {
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
			return &endOfDay, nil
		}
		startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		return &startOfDay, nil
	}

	if t, err := time.Parse(time.RFC3339Nano, clean); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	if t, err := time.Parse(time.RFC3339, clean); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05", clean); err == nil {
		utc := t.UTC()
		return &utc, nil
	}
	if t, err := time.Parse("2006-01-02T15:04", clean); err == nil {
		utc := t.UTC()
		if isEnd {
			endOfMin := time.Date(utc.Year(), utc.Month(), utc.Day(), utc.Hour(), utc.Minute(), 59, 999999999, time.UTC)
			return &endOfMin, nil
		}
		return &utc, nil
	}

	return nil, fmt.Errorf("invalid date format %q: expected YYYY-MM-DD or RFC3339", param)
}

func resolveTemporalFilter(windowParam, sinceParam, untilParam string) (*time.Time, *time.Time, error) {
	since, err := parseDateBoundary(sinceParam, false)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing 'since' parameter: %w", err)
	}
	until, err := parseDateBoundary(untilParam, true)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing 'until' parameter: %w", err)
	}

	if since != nil || until != nil {
		return since, until, nil
	}

	if windowParam != "" {
		window := domain.ParseTimeWindow(windowParam)
		if window == domain.TimeWindow1D {
			now := time.Now().UTC()
			refStart := now.Add(-24 * time.Hour)
			refEnd := now
			return &refStart, &refEnd, nil
		}
		if window != domain.TimeWindowAll && window != domain.TimeWindowCustom {
			var days int
			switch window {
			case domain.TimeWindow7D:
				days = 7
			case domain.TimeWindow30D:
				days = 30
			case domain.TimeWindow90D:
				days = 90
			}
			if days > 0 {
				now := time.Now().UTC()
				refEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)
				refStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -days)
				return &refStart, &refEnd, nil
			}
		}
	}

	return nil, nil, nil
}

// WorkspacesHandler handles GET /api/workspaces
func WorkspacesHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := adapters.ScanWorkspaces(workspacesDir)

		query := r.URL.Query()
		windowParam := query.Get("window")
		sinceParam := query.Get("since")
		if sinceParam == "" {
			sinceParam = query.Get("startDate")
		}
		untilParam := query.Get("until")
		if untilParam == "" {
			untilParam = query.Get("endDate")
		}
		sortParam := query.Get("sort")
		orderParam := query.Get("order")

		since, until, err := resolveTemporalFilter(windowParam, sinceParam, untilParam)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		ledgers := domain.FilterLedgersByCompletionDate(report.Ledgers, since, until)

		if sortParam != "" || orderParam != "" {
			field := domain.ParseLedgerSortField(sortParam)
			ascending := strings.ToLower(orderParam) == "asc"
			if orderParam == "" && field == domain.SortFieldTicketID {
				ascending = true
			}
			ledgers = domain.SortLedgers(ledgers, field, ascending)
		}

		if ledgers == nil {
			ledgers = []domain.AgenticMetricsLedger{}
		}

		WriteWorkspacesResponse(w, ledgers)
	}
}

// WorkspaceDetailHandler handles GET /api/workspaces/{ticketId}
func WorkspaceDetailHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID := r.PathValue("ticketId")
		if ticketID == "" || strings.Contains(ticketID, "..") || strings.Contains(ticketID, "/") || strings.Contains(ticketID, "\\") {
			WriteError(w, http.StatusBadRequest, "bad_request", "invalid ticketId parameter")
			return
		}

		targetLedger, err := adapters.LoadTicketWorkspace(ticketID, workspacesDir)
		if err != nil || targetLedger == nil {
			WriteError(w, http.StatusNotFound, "not_found", "Ticket workspace not found")
			return
		}

		artifacts := adapters.ReadTicketArtifacts(ticketID, workspacesDir)

		anomalies := targetLedger.Anomalies
		if anomalies == nil {
			anomalies = []domain.Anomaly{}
		}

		resp := WorkspaceDetailResponse{
			Ledger:    *targetLedger,
			Artifacts: artifacts,
			Anomalies: anomalies,
		}

		WriteWorkspaceDetailResponse(w, resp)
	}
}

// WorkspaceEventsHandler handles GET /api/workspaces/{ticketId}/events
func WorkspaceEventsHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID := r.PathValue("ticketId")
		if ticketID == "" || strings.Contains(ticketID, "..") || strings.Contains(ticketID, "/") || strings.Contains(ticketID, "\\") {
			WriteError(w, http.StatusBadRequest, "bad_request", "invalid ticketId parameter")
			return
		}

		targetLedger, err := adapters.LoadTicketWorkspace(ticketID, workspacesDir)
		if err != nil || targetLedger == nil {
			WriteError(w, http.StatusNotFound, "not_found", "Ticket workspace not found")
			return
		}

		sortedEvents := domain.SortEvents(targetLedger.Facts.Events)
		WriteEventsResponse(w, sortedEvents)
	}
}

// FleetHandler handles GET /api/fleet
func FleetHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := adapters.ScanWorkspaces(workspacesDir)

		query := r.URL.Query()
		windowParam := query.Get("window")
		sinceParam := query.Get("since")
		if sinceParam == "" {
			sinceParam = query.Get("startDate")
		}
		untilParam := query.Get("until")
		if untilParam == "" {
			untilParam = query.Get("endDate")
		}

		since, until, err := resolveTemporalFilter(windowParam, sinceParam, untilParam)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		ledgers := domain.FilterLedgersByCompletionDate(report.Ledgers, since, until)
		summary := domain.CalculateFleetSummary(ledgers)
		WriteFleetResponse(w, summary)
	}
}

// ProgressionHandler handles GET /api/progression
func ProgressionHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := adapters.ScanWorkspaces(workspacesDir)

		query := r.URL.Query()
		windowParam := query.Get("window")
		granularityParam := query.Get("granularity")
		if granularityParam == "" {
			granularityParam = string(domain.GranularityDaily)
		}

		granularity, ok := domain.ParseTimeGranularity(granularityParam)
		if !ok {
			WriteError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("invalid granularity %q: must be hourly, daily, weekly, or cumulative", granularityParam))
			return
		}

		startDateParam := query.Get("startDate")
		if startDateParam == "" {
			startDateParam = query.Get("since")
		}
		endDateParam := query.Get("endDate")
		if endDateParam == "" {
			endDateParam = query.Get("until")
		}

		since, until, err := resolveTemporalFilter(windowParam, startDateParam, endDateParam)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
			return
		}

		buckets := domain.AggregateMetricsOverTime(report.Ledgers, string(granularity), since, until)
		WriteProgressionResponse(w, buckets)
	}
}

// VersionHandler handles GET /api/version
func VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := domain.GetBuildInfo()
		WriteVersionResponse(w, info)
	}
}
