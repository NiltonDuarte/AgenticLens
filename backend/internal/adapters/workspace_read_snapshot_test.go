package adapters

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// Resolve first, then replace the pathname. A title observed before resolution
// would not establish which physical directory the final read must protect.
func TestResolvedWorkspaceReadsRejectSameIDDirectoryReplacement(t *testing.T) {
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	privateHome := t.TempDir()
	t.Setenv("HOME", privateHome)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(privateHome, "cache"))
	cfg, err := config.LoadFromFile("config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConfigPath = filepath.Join(project, ".agenticlens-config.yaml")
	cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: ".", Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 1, MaxDirectories: 3}}}
	cfg.Workspaces.Artifacts = []domain.WorkspaceArtifactConfig{{Path: "evidence.md", Title: "Evidence"}}
	selected, replacement, moved := filepath.Join(project, "selected"), filepath.Join(project, "replacement"), filepath.Join(project, "moved")
	originals := map[string][]byte{}
	for _, entry := range []struct{ path, title string }{{selected, "Copy A"}, {replacement, "Copy B"}} {
		if err := os.Mkdir(entry.path, 0700); err != nil {
			t.Fatal(err)
		}
		body := []byte(fmt.Sprintf(`{"schema_version":6,"ticket_id":"Recorded-A","ticket_title":%q,"sessions":[]}`, entry.title))
		originals[entry.title] = body
		for name, contents := range map[string][]byte{"agentic_metrics.json": body, "evidence.md": []byte(entry.title)} {
			if err := os.WriteFile(filepath.Join(entry.path, name), contents, 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	var ambiguity *WorkspaceResolutionError
	if _, err := ResolveTicketLocation(cfg, "Recorded-A", ""); !errors.As(err, &ambiguity) || ambiguity.Outcome != "ambiguous" {
		t.Fatalf("distinct same-ID copies lost explicit selection requirement: %v", err)
	}
	assertCopy := func(location WorkspaceLocation, title string) {
		t.Helper()
		ledger, err := readWorkspaceLedger(location, cfg)
		if err != nil || ledger.TicketTitle == nil || *ledger.TicketTitle != title {
			t.Fatalf("stable selected ledger changed: %+v %v", ledger, err)
		}
		artifacts, err := readResolvedArtifacts("Recorded-A", location, location.Artifacts)
		if err != nil || len(artifacts.Items) != 1 || artifacts.Items[0].Content == nil || *artifacts.Items[0].Content != title {
			t.Fatalf("stable selected artifact changed: %+v %v", artifacts, err)
		}
	}
	resolved, held, err := resolveTicketReadLocation(cfg, "Recorded-A", selected)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	})
	assertCopy(resolved, "Copy A")
	if err := os.Rename(selected, moved); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, selected); err != nil {
		t.Fatal(err)
	}
	ledger, ledgerErr := readWorkspaceLedger(resolved, cfg)
	var ledgerConflict *WorkspaceResolutionError
	if !errors.As(ledgerErr, &ledgerConflict) || ledgerConflict.Outcome != "unavailable" || !strings.Contains(ledgerErr.Error(), "retry") || ledger.TicketID != nil || ledger.TicketTitle != nil {
		t.Fatalf("resolved A silently reopened same-ID B: %+v %v", ledger, ledgerErr)
	}
	artifacts, artifactErr := readResolvedArtifacts("Recorded-A", resolved, resolved.Artifacts)
	var artifactConflict *WorkspaceResolutionError
	if !errors.As(artifactErr, &artifactConflict) || artifactConflict.Outcome != "unavailable" || !strings.Contains(artifactErr.Error(), "retry") || len(artifacts.Items) != 0 {
		t.Fatalf("resolved A exposed replacement artifacts: %+v %v", artifacts, artifactErr)
	}
	// B already occupying the selection when resolution begins is legitimate.
	newlyResolved, newlyHeld, err := resolveTicketReadLocation(cfg, "Recorded-A", selected)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := newlyHeld.Close(); err != nil {
			t.Error(err)
		}
	})
	assertCopy(newlyResolved, "Copy B")
	for path, title := range map[string]string{moved: "Copy A", selected: "Copy B"} {
		if body, err := os.ReadFile(filepath.Join(path, "agentic_metrics.json")); err != nil || !bytes.Equal(body, originals[title]) {
			t.Fatalf("read changed %s ledger: %q %v", title, body, err)
		}
		if body, err := os.ReadFile(filepath.Join(path, "evidence.md")); err != nil || string(body) != title {
			t.Fatalf("read changed %s evidence: %q %v", title, body, err)
		}
	}
	if entries, err := os.ReadDir(privateHome); err != nil || len(entries) != 0 {
		t.Fatalf("portable read initialized private state: %v %v", entries, err)
	}
	if _, err := os.Stat(filepath.Join(project, ".agenticlens-config.local.yaml")); !os.IsNotExist(err) {
		t.Fatalf("read initialized local configuration: %v", err)
	}
}

func TestResolvedReadReferenceAllowsWriterAdmission(t *testing.T) {
	project := t.TempDir()
	writePublicationTestPolicy(t, project)
	path := filepath.Join(project, "agentic_metrics.json")
	body := []byte(`{"schema_version":6,"ticket_id":"Recorded-A","ticket_title":"Copy A","sessions":[]}`)
	if err := os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	_, cfg, err := LoadWorkspaceProject(project)
	if err != nil {
		t.Fatal(err)
	}
	_, held, err := resolveTicketReadLocation(cfg, "Recorded-A", project)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := held.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	owner, err := capture.AcquirePublication(ctx, path, project)
	if err != nil {
		t.Fatalf("transient read reference retained writer-blocking ownership: %v", err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if after, err := os.ReadFile(path); err != nil || !bytes.Equal(after, body) {
		t.Fatalf("ownership admission changed recorded facts: %q %v", after, err)
	}
}
