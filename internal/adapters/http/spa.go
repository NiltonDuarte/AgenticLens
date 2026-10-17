package http

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// SPAHandler serves static files from an fs.FS and falls back to index.html for client-side routing.
type SPAHandler struct {
	fileServer http.Handler
	fileSystem fs.FS
}

// NewSPAHandler constructs a new SPAHandler backed by an fs.FS.
func NewSPAHandler(fileSystem fs.FS) *SPAHandler {
	return &SPAHandler{
		fileServer: http.FileServer(http.FS(fileSystem)),
		fileSystem: fileSystem,
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	cleanPath := path.Clean(r.URL.Path)

	// Guard: Never serve SPA fallback for non-existent API routes
	if strings.HasPrefix(cleanPath, "/api/") || cleanPath == "/api" {
		WriteError(w, http.StatusNotFound, "not_found", "API endpoint not found")
		return
	}

	// Normalise root request
	if cleanPath == "/" || cleanPath == "." {
		h.serveIndex(w, r)
		return
	}

	fsPath := strings.TrimPrefix(cleanPath, "/")

	// Attempt opening requested file in the filesystem
	f, err := h.fileSystem.Open(fsPath)
	if err == nil {
		defer f.Close()
		stat, statErr := f.Stat()
		if statErr == nil && !stat.IsDir() {
			// Static file exists and is not directory; delegate to FileServer
			h.fileServer.ServeHTTP(w, r)
			return
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to access asset")
		return
	}

	// Fallback to index.html for client-side routing
	h.serveIndex(w, r)
}

func (h *SPAHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	content, err := fs.ReadFile(h.fileSystem, "index.html")
	if err != nil {
		WriteError(w, http.StatusNotFound, "not_found", "index.html not found")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(string(content)))
}
