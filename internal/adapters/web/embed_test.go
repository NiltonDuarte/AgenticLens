package web_test

import (
	"io/fs"
	"strings"
	"testing"

	"agenticlens/internal/adapters/web"
)

func TestGetFS(t *testing.T) {
	staticFS, err := web.GetFS()
	if err != nil {
		t.Fatalf("expected nil error from GetFS, got: %v", err)
	}

	content, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		t.Fatalf("expected index.html to be readable, got: %v", err)
	}

	if !strings.Contains(string(content), "AgenticLens") {
		t.Errorf("expected index.html content to contain 'AgenticLens', got: %s", string(content))
	}
}
