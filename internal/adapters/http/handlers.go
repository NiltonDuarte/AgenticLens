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

		report := adapters.ScanWorkspaces(workspacesDir)

		var targetLedger *domain.AgenticMetricsLedger
		for i := range report.Ledgers {
			l := &report.Ledgers[i]
			if l.TicketID != nil && *l.TicketID == ticketID {
				targetLedger = l
				break
			}
		}

		if targetLedger == nil {
			WriteError(w, http.StatusNotFound, "not_found", "Ticket workspace not found")
			return
		}

		artifacts := adapters.ReadTicketArtifacts(ticketID, workspacesDir)

		resp := WorkspaceDetailResponse{
			Ledger:    *targetLedger,
			Artifacts: artifacts,
		}

		WriteWorkspaceDetailResponse(w, resp)
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
