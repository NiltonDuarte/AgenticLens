package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"agenticlens/internal/adapters/live"
	"agenticlens/internal/domain"
)

// LiveServiceInterface defines the contract consumed by HTTP live handlers.
type LiveServiceInterface interface {
	GetSessionTree(ctx context.Context) (domain.LiveSessionTree, error)
	GetSessionTrace(ctx context.Context, sessionID string) (domain.LiveTraceView, error)
	Subscribe() (<-chan live.LiveEvent, func())
}

// LiveSessionsHandler returns active sessions organized into a hierarchical tree.
func LiveSessionsHandler(service LiveServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			WriteLiveSessionTreeResponse(w, domain.LiveSessionTree{
				RootSessions:        []*domain.LiveSessionNode{},
				LastRefreshedAt:     "",
				TotalActiveSessions: 0,
			})
			return
		}

		tree, err := service.GetSessionTree(r.Context())
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}

		WriteLiveSessionTreeResponse(w, tree)
	}
}

// LiveTraceHandler returns execution messages, reasoning, and tool errors for a target session.
func LiveTraceHandler(service LiveServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		cleanID := strings.TrimSpace(sessionID)
		lowerID := strings.ToLower(cleanID)

		if cleanID == "" || strings.Contains(cleanID, "..") || strings.Contains(cleanID, "/") || strings.Contains(cleanID, "\\") ||
			strings.Contains(lowerID, "%2e") || strings.Contains(lowerID, "%2f") || strings.Contains(lowerID, "%5c") {
			WriteError(w, http.StatusBadRequest, "invalid_session_id", "session id contains invalid path traversal characters")
			return
		}

		if service == nil {
			WriteError(w, http.StatusNotFound, "session_not_found", fmt.Sprintf("session %s not found", cleanID))
			return
		}

		trace, err := service.GetSessionTrace(r.Context(), cleanID)
		if err != nil {
			if errors.Is(err, live.ErrInvalidSessionID) {
				WriteError(w, http.StatusBadRequest, "invalid_session_id", err.Error())
				return
			}
			if errors.Is(err, live.ErrSessionNotFound) {
				WriteError(w, http.StatusNotFound, "session_not_found", fmt.Sprintf("session %s not found", cleanID))
				return
			}
			WriteError(w, http.StatusInternalServerError, "database_error", err.Error())
			return
		}

		WriteLiveTraceResponse(w, trace)
	}
}

// LiveStreamHandler establishes a Server-Sent Events (SSE) connection streaming real-time live events.
func LiveStreamHandler(service LiveServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

		flusher.Flush()

		if service == nil {
			<-r.Context().Done()
			return
		}

		// Push initial session_tree event upon connection
		tree, err := service.GetSessionTree(r.Context())
		if err == nil {
			if treeData, err := json.Marshal(tree); err == nil {
				fmt.Fprintf(w, "event: session_tree\ndata: %s\n\n", treeData)
				flusher.Flush()
			}
		}

		// Subscribe to real-time broadcaster
		eventsCh, unsubscribe := service.Subscribe()
		defer unsubscribe()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-eventsCh:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, event.Data)
				flusher.Flush()
			}
		}
	}
}
