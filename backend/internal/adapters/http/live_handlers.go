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
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// LiveServiceInterface defines the contract consumed by HTTP live handlers.
type LiveServiceInterface interface {
	CheckAvailability() error
	GetSessionTree(ctx context.Context) (domain.LiveSessionTree, error)
	GetSessionTrace(ctx context.Context, sessionID string) (domain.LiveTraceView, error)
	Subscribe() (<-chan live.LiveEvent, func())
}

func writeLiveSourceError(w http.ResponseWriter, err error) {
	var layoutErr opencode.UnsupportedLayoutError
	switch {
	case errors.As(err, &layoutErr):
		WriteError(w, http.StatusServiceUnavailable, "source_unsupported", err.Error())
	case errors.Is(err, live.ErrSourceUnavailable):
		WriteError(w, http.StatusServiceUnavailable, "source_unavailable", err.Error())
	default:
		WriteError(w, http.StatusInternalServerError, "database_error", err.Error())
	}
}

// LiveSessionsHandler returns active sessions organized into a hierarchical tree.
func LiveSessionsHandler(service LiveServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			writeLiveSourceError(w, live.ErrSourceUnavailable)
			return
		}

		tree, err := service.GetSessionTree(r.Context())
		if err != nil {
			writeLiveSourceError(w, err)
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
			writeLiveSourceError(w, live.ErrSourceUnavailable)
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
			writeLiveSourceError(w, err)
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

		if service == nil {
			writeLiveSourceError(w, live.ErrSourceUnavailable)
			return
		}
		if err := service.CheckAvailability(); err != nil {
			writeLiveSourceError(w, err)
			return
		}

		// Subscribe before the initial read so concurrent updates are queued.
		eventsCh, unsubscribe := service.Subscribe()
		defer unsubscribe()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

		flusher.Flush()

		tree, err := service.GetSessionTree(r.Context())
		var initial live.LiveEvent
		if err == nil {
			initial.Type = live.LiveEventSessionTree
			initial.Data, err = json.Marshal(tree)
		}
		if err != nil {
			initial = live.SourceDiagnosticEvent(err, "")
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", initial.Type, initial.Data); err != nil {
			return
		}
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-eventsCh:
				if !ok {
					return
				}
				if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, event.Data); err != nil {
					return
				}
				flusher.Flush()
			}
		}
	}
}
