package capture

import (
	"agenticlens/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestTicketMetadataConflictOrder(t *testing.T) {
	rows := []domain.CaptureRegistration{
		{TicketID: "TEST-1", TicketTitle: "one", TicketType: "feat"},
		{TicketID: "TEST-1", TicketTitle: "two", TicketType: "feat"},
	}
	assertMetadata := func(wantTitle string) {
		t.Helper()
		title, kind, err := domain.CaptureTicketMetadata(rows, "TEST-1")
		if err != nil || title != wantTitle || kind != "feat" {
			t.Fatalf("ticket metadata = (%q, %q, %v); want (%q, feat, nil)", title, kind, err, wantTitle)
		}
	}
	assertMetadata("")
	rows[0], rows[1] = rows[1], rows[0]
	assertMetadata("")
	rows[1] = rows[0]
	assertMetadata("two")

	rows[1].TicketType = "fix"
	for range 2 {
		if _, _, err := domain.CaptureTicketMetadata(rows, "TEST-1"); err == nil {
			t.Fatal("conflicting ticket types accepted")
		}
		rows[0], rows[1] = rows[1], rows[0]
	}
}
func TestSourceProjectDescendantBoundary(t *testing.T) {
	project := t.TempDir()
	child := filepath.Join(project, "src", "pkg")
	if err := os.MkdirAll(child, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ValidateProject(project, child); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "src", ".agenticlens-config.yaml"), []byte("project_name: nested"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateProject(project, child); err == nil {
		t.Fatal("nested project accepted")
	}
	if err := ValidateProject(project, t.TempDir()); err == nil {
		t.Fatal("different project accepted")
	}
}
