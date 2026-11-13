package http

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

// Server wraps the HTTP multiplexer, server configuration, and filesystem access.
type Server struct {
	httpServer    *http.Server
	workspacesDir string
	staticFS      fs.FS
	liveService   LiveServiceInterface
	projectConfig domain.ProjectConfig
}

// NewServerWithConfig creates a new HTTP Server instance configured with custom ProjectConfig, API routes, and SPA static serving.
func NewServerWithConfig(workspacesDir string, staticFS fs.FS, cfg domain.ProjectConfig, liveService ...LiveServiceInterface) *Server {
	cleanWsDir := strings.TrimSpace(workspacesDir)
	if cfg.ConfigPath != "" {
		cleanWsDir = filepath.Dir(cfg.ConfigPath)
	}

	var ls LiveServiceInterface
	if len(liveService) > 0 {
		ls = liveService[0]
	}

	var hydrator WorkspaceEventHydrator
	if h, ok := ls.(WorkspaceEventHydrator); ok {
		hydrator = h
	}

	mux := http.NewServeMux()

	// Register API endpoints
	mux.HandleFunc("GET /api/version", VersionHandler())
	mux.HandleFunc("GET /api/config", ConfigHandler(cfg, cleanWsDir))
	mux.HandleFunc("GET /api/workspaces", WorkspacesHandler(cleanWsDir))
	mux.HandleFunc("GET /api/workspaces/{ticketId}", WorkspaceDetailHandler(cleanWsDir))
	mux.HandleFunc("GET /api/workspaces/{ticketId}/events", WorkspaceEventsHandler(cleanWsDir, hydrator))
	mux.HandleFunc("GET /api/fleet", FleetHandler(cleanWsDir))
	mux.HandleFunc("GET /api/comparison", ComparisonHandler(cleanWsDir))
	mux.HandleFunc("GET /api/progression", ProgressionHandler(cleanWsDir))

	// Register Live API endpoints
	mux.HandleFunc("GET /api/live/sessions", LiveSessionsHandler(ls))
	mux.HandleFunc("GET /api/live/sessions/{id}/trace", LiveTraceHandler(ls))
	mux.HandleFunc("GET /api/live/stream", LiveStreamHandler(ls))

	// Register SPA static files and fallback handler on root
	spaHandler := NewSPAHandler(staticFS)
	mux.Handle("/", spaHandler)

	// Wrap entire router with CORS middleware
	handler := CORSMiddleware(mux)

	s := &Server{
		workspacesDir: cleanWsDir,
		staticFS:      staticFS,
		liveService:   ls,
		projectConfig: cfg,
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

// Bind validates and binds exactly one loopback port before it is advertised.
func (s *Server) Bind(port int) (net.Listener, error) {
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}
	address := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	s.httpServer.Addr = listener.Addr().String()
	return listener, nil
}
func (s *Server) Serve(listener net.Listener) error { return s.httpServer.Serve(listener) }
func (s *Server) ListenAndServe(port int) error {
	listener, err := s.Bind(port)
	if err != nil {
		return err
	}
	return s.Serve(listener)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
