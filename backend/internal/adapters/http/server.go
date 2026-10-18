package http

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

// Server wraps the HTTP multiplexer, server configuration, and filesystem access.
type Server struct {
	httpServer    *http.Server
	workspacesDir string
	staticFS      fs.FS
}

// NewServer creates a new HTTP Server instance configured with API routes and SPA static serving.
func NewServer(workspacesDir string, staticFS fs.FS) *Server {
	mux := http.NewServeMux()

	// Register API endpoints
	mux.HandleFunc("GET /api/workspaces", WorkspacesHandler(workspacesDir))
	mux.HandleFunc("GET /api/workspaces/{ticketId}", WorkspaceDetailHandler(workspacesDir))
	mux.HandleFunc("GET /api/workspaces/{ticketId}/events", WorkspaceEventsHandler(workspacesDir))
	mux.HandleFunc("GET /api/fleet", FleetHandler(workspacesDir))

	// Register SPA static files and fallback handler on root
	spaHandler := NewSPAHandler(staticFS)
	mux.Handle("/", spaHandler)

	// Wrap entire router with CORS middleware
	handler := CORSMiddleware(mux)

	s := &Server{
		workspacesDir: workspacesDir,
		staticFS:      staticFS,
	}

	s.httpServer = &http.Server{
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Handler returns the root HTTP handler, useful for integration testing with httptest.
func (s *Server) Handler() http.Handler {
	return s.httpServer.Handler
}

// ListenAndServe starts the server on the specified port.
func (s *Server) ListenAndServe(port int) error {
	s.httpServer.Addr = fmt.Sprintf(":%d", port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
