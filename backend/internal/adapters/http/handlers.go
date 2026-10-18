package http

import (
	"net/http"
	"strings"
	"time"

	"agenticlens/internal/adapters"
	"agenticlens/internal/domain"
)

// WorkspacesHandler handles GET /api/workspaces
func WorkspacesHandler(workspacesDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		report := adapters.ScanWorkspaces(workspacesDir)

		query := r.URL.Query()
		windowParam := query.Get("window")
		sortParam := query.Get("sort")
		orderParam := query.Get("order")

		ledgers := report.Ledgers
		if windowParam != "" {
			window := domain.ParseTimeWindow(windowParam)
			ledgers = domain.FilterLedgersByTimeWindow(ledgers, window, time.Now())
		}

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

		ledgers := report.Ledgers
		if windowParam != "" {
			window := domain.ParseTimeWindow(windowParam)
			ledgers = domain.FilterLedgersByTimeWindow(ledgers, window, time.Now())
		}

		summary := domain.CalculateFleetSummary(ledgers)
		WriteFleetResponse(w, summary)
	}
}

// VersionHandler handles GET /api/version
func VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := domain.GetBuildInfo()
		WriteVersionResponse(w, info)
	}
}
