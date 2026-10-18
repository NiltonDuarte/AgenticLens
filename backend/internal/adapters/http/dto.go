package http

import (
	"encoding/json"
	"net/http"

	"agenticlens/internal/domain"
)

// ErrorResponse represents a structured error response payload.
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

// WorkspaceDetailResponse represents the serialized workspace detail payload.
type WorkspaceDetailResponse struct {
	Ledger    domain.AgenticMetricsLedger `json:"ledger"`
	Artifacts domain.TicketArtifacts      `json:"artifacts"`
	Anomalies []domain.Anomaly            `json:"anomalies"`
}

// BuildInfoDTO represents the serialized build information payload.
type BuildInfoDTO struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

// ResponsePayload defines the permitted concrete types for JSON responses.
type ResponsePayload interface {
	ErrorResponse | []domain.AgenticMetricsLedger | WorkspaceDetailResponse | domain.FleetSummary | []domain.EventFact | BuildInfoDTO
}

// WriteJSON writes a typed JSON response with status code, enforcing concrete response types.
func WriteJSON[T ResponsePayload](w http.ResponseWriter, statusCode int, data T) {
	payload, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		errFallback := []byte(`{"error":"internal_error","message":"failed to encode response","status_code":500}`)
		if _, writeErr := w.Write(errFallback); writeErr != nil {
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if _, err := w.Write(payload); err != nil {
		return
	}
}

// WriteError writes a structured ErrorResponse with status code.
func WriteError(w http.ResponseWriter, statusCode int, errCode string, message string) {
	resp := ErrorResponse{
		Error:      errCode,
		Message:    message,
		StatusCode: statusCode,
	}
	WriteJSON(w, statusCode, resp)
}

// WriteWorkspacesResponse writes a slice of ledgers as JSON.
func WriteWorkspacesResponse(w http.ResponseWriter, ledgers []domain.AgenticMetricsLedger) {
	WriteJSON(w, http.StatusOK, ledgers)
}

// WriteWorkspaceDetailResponse writes a WorkspaceDetailResponse as JSON.
func WriteWorkspaceDetailResponse(w http.ResponseWriter, detail WorkspaceDetailResponse) {
	WriteJSON(w, http.StatusOK, detail)
}

// WriteFleetResponse writes a domain.FleetSummary as JSON.
func WriteFleetResponse(w http.ResponseWriter, summary domain.FleetSummary) {
	WriteJSON(w, http.StatusOK, summary)
}

// WriteVersionResponse writes a domain.BuildInfo as JSON.
func WriteVersionResponse(w http.ResponseWriter, info domain.BuildInfo) {
	dto := BuildInfoDTO{
		Version:   info.Version,
		Commit:    info.GitCommit,
		BuildDate: info.BuildDate,
	}
	WriteJSON(w, http.StatusOK, dto)
}

// WriteEventsResponse writes a slice of EventFact as JSON.
func WriteEventsResponse(w http.ResponseWriter, events []domain.EventFact) {
	if events == nil {
		events = []domain.EventFact{}
	}
	WriteJSON(w, http.StatusOK, events)
}
