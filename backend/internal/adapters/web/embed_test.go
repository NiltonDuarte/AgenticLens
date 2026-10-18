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

	htmlStr := string(content)
	if !strings.Contains(htmlStr, "AgenticLens") {
		t.Errorf("expected index.html content to contain 'AgenticLens', got: %s", htmlStr)
	}

	if !strings.Contains(htmlStr, `<div id="root"></div>`) {
		t.Errorf("expected index.html to contain root container div, got: %s", htmlStr)
	}

	if !strings.Contains(htmlStr, "/assets/index-") {
		t.Errorf("expected index.html to reference bundled assets under /assets/, got: %s", htmlStr)
	}

	// Verify assets directory is readable and contains bundled JS and CSS
	entries, err := fs.ReadDir(staticFS, "assets")
	if err != nil {
		t.Fatalf("expected assets directory to be readable in embedded FS: %v", err)
	}

	var hasJS, hasCSS bool
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".js") {
			hasJS = true
			jsContent, readErr := fs.ReadFile(staticFS, "assets/"+entry.Name())
			if readErr != nil {
				t.Errorf("failed to read embedded js asset %s: %v", entry.Name(), readErr)
			}
			if len(jsContent) == 0 {
				t.Errorf("embedded js asset %s is unexpectedly empty", entry.Name())
			}
		}
		if strings.HasSuffix(entry.Name(), ".css") {
			hasCSS = true
			cssContent, readErr := fs.ReadFile(staticFS, "assets/"+entry.Name())
			if readErr != nil {
				t.Errorf("failed to read embedded css asset %s: %v", entry.Name(), readErr)
			}
			if len(cssContent) == 0 {
				t.Errorf("embedded css asset %s is unexpectedly empty", entry.Name())
			}
		}
	}

	if !hasJS {
		t.Errorf("expected at least one .js bundle in embedded assets directory")
	}
	if !hasCSS {
		t.Errorf("expected at least one .css bundle in embedded assets directory")
	}
}
