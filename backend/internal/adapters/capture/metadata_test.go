package capture

import (
	"agenticlens/internal/domain"
	"os"
	"path/filepath"
	"testing"
)

func TestTicketMetadataConflictOrder(t *testing.T) {
	for _, field := range []string{"title", "type"} {
		rows := []domain.CaptureRegistration{{TicketID: "TEST-1", TicketTitle: "one", TicketType: "feat"}, {TicketID: "TEST-1", TicketTitle: "one", TicketType: "feat"}}
		if field == "title" {
			rows[1].TicketTitle = "two"
		} else {
			rows[1].TicketType = "fix"
		}
		if _, _, err := domain.CaptureTicketMetadata(rows, "TEST-1"); err == nil {
			t.Fatal("conflict accepted")
		}
		rows[0], rows[1] = rows[1], rows[0]
		if _, _, err := domain.CaptureTicketMetadata(rows, "TEST-1"); err == nil {
			t.Fatal("reverse conflict accepted")
		}
		rows[1] = rows[0]
		if _, _, err := domain.CaptureTicketMetadata(rows, "TEST-1"); err != nil {
			t.Fatal(err)
		}
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
