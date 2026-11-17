package capture_test

import (
	"agenticlens/internal/adapters/capture"
	"os"
	"path/filepath"
	"testing"
)

// This exercises actual descriptor-relative member lookup, rather than relying
// on Linux's special /proc pathname traversal to establish portability.
func TestPublicationReadRootSurvivesRenameAndRejectsEscapingMembers(t *testing.T) {
	project := t.TempDir()
	home := filepath.Join(project, "selected")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "agentic_metrics.json")
	if err := os.WriteFile(path, []byte("accepted source"), 0600); err != nil {
		t.Fatal(err)
	}
	reader, err := capture.ReadPublicationDirectory(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	archive := filepath.Join(project, "archived")
	if err = os.Rename(home, archive); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, []byte("unrelated replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	body, err := reader.ReadFile("agentic_metrics.json")
	if err != nil || string(body) != "accepted source" {
		t.Fatalf("read followed replacement instead of held home: %q %v", body, err)
	}
	outside := filepath.Join(t.TempDir(), "private")
	if err = os.WriteFile(outside, []byte("outside evidence"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(archive, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, member := range []string{"../selected/agentic_metrics.json", "escape"} {
		if body, err = reader.ReadFile(member); err == nil {
			t.Fatalf("escaping member %q exposed %q", member, body)
		}
	}
	if _, err = os.Stat(filepath.Join(project, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
		t.Fatalf("read initialized private state: %v", err)
	}
}
