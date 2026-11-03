package capture

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func sharedWorktrees(t *testing.T) (string, string, string) {
	t.Helper()
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.Mkdir(repo, 0700); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	run("init", "-q")
	cfg, err := adaptersconfig.GenerateDefaultYAML("SharedProject", "workspaces", "", []string{"src"}, []domain.TicketPrefix{{Prefix: "TEST"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(repo, ".agenticlens-config.yaml"), cfg, 0600); err != nil {
		t.Fatal(err)
	}
	run("add", ".agenticlens-config.yaml")
	run("-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fixture")
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	run("worktree", "add", "-qb", "a", a)
	run("worktree", "add", "-qb", "b", b)
	return a, b, filepath.Join(root, "state")
}

func TestSharedNamedRegistryScopesConcurrentSameTicketWorktrees(t *testing.T) {
	a, b, base := sharedWorktrees(t)
	one, err := ProjectState(base, a)
	if err != nil {
		t.Fatal(err)
	}
	two, err := ProjectState(base, b)
	if err != nil {
		t.Fatal(err)
	}
	if one != two || filepath.Base(one) != "SharedProject" {
		t.Fatalf("named registry differs: %s %s", one, two)
	}
	first, err := OpenProject(one, a)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := OpenProject(two, b)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if first.worktreeID == second.worktreeID || first.directory == second.directory {
		t.Fatal("worktree coordination overlaps")
	}
	// The same exact minimal identity can be observed before either source exists;
	// durable handoff must not share its challenge or mutable ticket state.
	store := t.TempDir()
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: store, SessionID: "same-observed-identity"}
	ctx := context.Background()
	var group sync.WaitGroup
	errs := make(chan error, 2)
	for _, s := range []*Store{first, second} {
		group.Add(1)
		go func(s *Store) {
			defer group.Done()
			_, e := s.Register(ctx, id, "", "UserPromptSubmit")
			if e == nil {
				_, e = s.Change(ctx, id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
					r.TicketID = "TEST-1"
					return r, nil
				})
			}
			if e == nil {
				e = s.RequestPublication(ctx)
			}
			if e == nil {
				e = s.RecordDiagnostic(ctx, domain.CaptureDiagnostic{TicketID: "TEST-1", Phase: "publication", Outcome: "failed", Detail: s.project})
			}
			errs <- e
		}(s)
	}
	group.Wait()
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	rowsA, err := first.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rowsB, err := second.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(rowsA) != 1 || len(rowsB) != 1 || rowsA[0].Request == rowsB[0].Request {
		t.Fatal("registration/challenge scope mixed")
	}
	if err = first.AcknowledgePublication(ctx, 1); err != nil {
		t.Fatal(err)
	}
	ra, err := first.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := second.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if ra.Acknowledged != 1 || rb.Acknowledged != 0 || rb.Generation != 1 {
		t.Fatalf("request scope mixed: %+v %+v", ra, rb)
	}
	for _, s := range []*Store{first, second} {
		ds, e := s.Diagnostics(ctx)
		if e != nil {
			t.Fatal(e)
		}
		if len(ds) != 1 || ds[0].Detail != s.project {
			t.Fatalf("diagnostics mixed: %+v", ds)
		}
		current, e := s.ReadDiagnostics(ctx, DiagnosticReadOptions{Current: true, Tickets: []string{"TEST-1"}, Limit: 32, MaxBytes: 65536})
		if e != nil {
			t.Fatal(e)
		}
		if !current.MetadataComplete || len(current.Records) != 1 || current.Records[0].Detail != s.project {
			t.Fatalf("heads mixed: %+v", current)
		}
	}
	lockA, err := AcquireWorker(first.directory)
	if err != nil || lockA == nil {
		t.Fatalf("worker A: %v", err)
	}
	defer lockA.Close()
	lockB, err := AcquireWorker(second.directory)
	if err != nil || lockB == nil {
		t.Fatalf("worker B: %v", err)
	}
	defer lockB.Close()
	other, err := AcquireWorker(first.directory)
	if err != nil || other != nil {
		if other != nil {
			other.Close()
		}
		t.Fatalf("same worktree allowed concurrent worker: %v", err)
	}
	// Passing another scope's coordination directory never authorizes that scope.
	if s, e := OpenProject(first.directory, b); e == nil {
		s.Close()
		t.Fatal("foreign coordination path accepted")
	}
}

func TestUnsupportedRegistryOpenDoesNotChangeLegacyDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("CREATE TABLE capture_meta(version INTEGER NOT NULL); INSERT INTO capture_meta VALUES(1); CREATE TABLE legacy_fact(value TEXT); INSERT INTO legacy_fact VALUES('preserved')"); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if s, e := openRegistry(dir, "Project", t.TempDir(), false); e == nil {
		s.Close()
		t.Fatal("protocol1 opened as shared schema2")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("unsupported open changed legacy database")
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, e := os.Stat(path + suffix); !os.IsNotExist(e) {
			t.Fatalf("unsupported open created SQLite coordination file %s: %v", suffix, e)
		}
	}
}

func TestProjectNameNamespacePreservesConfiguredIdentity(t *testing.T) {
	names := []string{"AgenticLens", "AgenticLens Two", "项目", "CON", "agenticlens", "~4142", "Name.", "CON.txt", "COM1.data"}
	seen := map[string]string{}
	for _, name := range names {
		segment, err := ProjectNameSegment(name)
		if err != nil {
			t.Fatal(err)
		}
		if other, exists := seen[segment]; exists {
			t.Fatalf("names %s/%s collide", name, other)
		}
		seen[segment] = name
		if (name == "Name." || name == "CON.txt" || name == "COM1.data") && segment[0] != '~' {
			t.Fatalf("nonportable Windows name remained literal: %q", name)
		}
		if filepath.Base(segment) != segment {
			t.Fatal("name escaped namespace")
		}
	}
	for _, name := range []string{"", ".", "..", "../escape", "name/escape", "name\\escape"} {
		if _, err := ProjectNameSegment(name); err == nil {
			t.Fatalf("unsafe name accepted: %q", name)
		}
	}
}

func TestSharedRegistryRefusesReusedWorktreeButAllowsNormalWrites(t *testing.T) {
	a, _, base := sharedWorktrees(t)
	registry, err := ProjectState(base, a)
	if err != nil {
		t.Fatal(err)
	}
	s, err := OpenProject(registry, a)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(a, "ordinary-agent-output.txt"), []byte("work"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err = OpenProject(registry, a)
	if err != nil {
		t.Fatalf("ordinary file activity invalidated ownership: %v", err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	// Retain the former directory so its inode cannot be immediately reused.
	retired := a + "-retired"
	if err = os.Rename(a, retired); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(a, 0700); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(retired, ".agenticlens-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(a, ".agenticlens-config.yaml"), config, 0600); err != nil {
		t.Fatal(err)
	}
	if reopened, e := OpenProject(registry, a); e == nil {
		reopened.Close()
		t.Fatal("replacement directory acquired retained worktree scope")
	}
}

func TestCaptureSQLiteURIUsesPathAndEscapesFilenameForRealReadOnlyOpen(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "capture #percent% and quote'")
	store, err := Open(directory)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "registry.db")
	uri, err := sqliteFileURI(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "file" || parsed.Host != "" || parsed.Fragment != "" || parsed.RawQuery != "" || !strings.HasPrefix(parsed.Path, "/") {
		t.Fatalf("SQLite filename became URI authority or options: %s", uri)
	}
	if !strings.Contains(uri, "%23") || !strings.Contains(uri, "%25") {
		t.Fatalf("literal filename not escaped: %s", uri)
	}
	reader, err := openReadOnlyDB(path)
	if err != nil {
		t.Fatal(err)
	}
	var version int
	readErr := reader.QueryRow("SELECT version FROM capture_meta").Scan(&version)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || version != ProtocolVersion {
		t.Fatalf("read-only open did not reach exact existing capture file: version=%d read=%v close=%v", version, readErr, closeErr)
	}
}

func TestCoordinationTokensConcurrentFirstOpenAndStableReopen(t *testing.T) {
	a, b, base := sharedWorktrees(t)
	other, _, _ := sharedWorktrees(t)
	type opened struct {
		project, path string
		err           error
	}
	results := make(chan opened, 8)
	start := make(chan struct{})
	projects := []string{a, a, b, b, other, other, a, b}
	gitPrefixes := map[string]string{}
	for _, project := range []string{a, b, other} {
		output, err := exec.Command("git", "-C", project, "rev-parse", "--absolute-git-dir").CombinedOutput()
		if err != nil {
			t.Fatalf("observe fixture administrative directory: %s %v", output, err)
		}
		gitPrefixes[project] = "wt-" + filepath.Base(strings.TrimSpace(string(output))) + "-"
	}
	for _, project := range projects {
		go func(project string) {
			<-start
			s, err := OpenProject(base, project)
			path := ""
			if err == nil {
				path = s.CoordinationDirectory()
				err = s.Close()
			}
			results <- opened{project, path, err}
		}(project)
	}
	close(start)
	byProject := map[string]string{}
	byPath := map[string]string{}
	for range projects {
		got := <-results
		if got.err != nil {
			t.Fatal(got.err)
		}
		if old := byProject[got.project]; old != "" && old != got.path {
			t.Fatalf("scope allocated twice: %s %s", old, got.path)
		}
		if old := byPath[got.path]; old != "" && old != got.project {
			t.Fatal("distinct worktrees share coordinates")
		}
		token, prefix := filepath.Base(got.path), gitPrefixes[got.project]
		if !strings.HasPrefix(token, prefix) {
			t.Fatalf("Git identifier missing: %s", got.path)
		}
		if qualification, err := hex.DecodeString(strings.TrimPrefix(token, prefix)); err != nil || len(qualification) != 8 {
			t.Fatalf("Git identity qualification missing: %s", token)
		}
		byProject[got.project] = got.path
		byPath[got.path] = got.project
	}
	for project, path := range byProject {
		s, err := OpenReadOnlyProject(base, project)
		if err != nil {
			t.Fatal(err)
		}
		if s.CoordinationDirectory() != path {
			t.Fatalf("reopen changed coordinate: %s", s.CoordinationDirectory())
		}
		if err = s.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCoordinationUUIDFallbackPersistsAcrossGitEnrichment(t *testing.T) {
	a, _, base := sharedWorktrees(t)
	cfg, err := os.ReadFile(filepath.Join(a, ".agenticlens-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	if err = os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), cfg, 0600); err != nil {
		t.Fatal(err)
	}
	preview, err := PreviewStateLocation(base, project)
	if err != nil || !preview.Provisional || preview.CoordinationDirectory != "" || preview.WorktreeID != 0 {
		t.Fatalf("invented preview: %+v %v", preview, err)
	}
	if _, err = os.Stat(base); !os.IsNotExist(err) {
		t.Fatalf("preview allocated state: %v", err)
	}
	s, err := OpenProject(base, project)
	if err != nil {
		t.Fatal(err)
	}
	path := s.CoordinationDirectory()
	if raw, e := hex.DecodeString(filepath.Base(path)); e != nil || len(raw) != 8 {
		t.Fatalf("not a short UUID: %s", path)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", "-q", project)
	if out, e := cmd.CombinedOutput(); e != nil {
		t.Fatalf("git init: %s %v", out, e)
	}
	s, err = OpenProject(base, project)
	if err != nil {
		t.Fatal(err)
	}
	if s.CoordinationDirectory() != path {
		t.Fatal("Git enrichment changed persisted token")
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCoordinateAllocationSkipsReservedNamesAndBoundsCollisions(t *testing.T) {
	project := t.TempDir()
	registry := t.TempDir()
	s, err := Open(registry)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	owner, err := currentWorktreeOwnership(project)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := s.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	root := filepath.Join(registry, "worktrees")
	if err = os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "reserved"), []byte("unowned"), 0600); err != nil {
		t.Fatal(err)
	}
	calls := 0
	path, err := allocateCoordinate(context.Background(), tx, registry, 123, owner, func() (string, error) {
		calls++
		if calls == 1 {
			return "reserved", nil
		}
		return "fresh", nil
	})
	if err != nil || filepath.Base(path) != "fresh" || calls != 2 {
		t.Fatalf("reserved collision: %s %d %v", path, calls, err)
	}
	_, err = allocateCoordinate(context.Background(), tx, registry, 124, owner, func() (string, error) { return "fresh", nil })
	if err == nil || !strings.Contains(err.Error(), "16 collisions") {
		t.Fatalf("unbounded collision: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, "reserved"))
	if err != nil || string(body) != "unowned" {
		t.Fatal("reserved state overwritten")
	}
}

func TestUnsupportedCoordinateLayoutFailsWithoutChangingState(t *testing.T) {
	for _, missing := range []string{"table", "mapping"} {
		t.Run(missing, func(t *testing.T) {
			project, _, base := sharedWorktrees(t)
			store, err := OpenProject(base, project)
			if err != nil {
				t.Fatal(err)
			}
			registry, coordinate := store.RegistryDirectory(), store.CoordinationDirectory()
			if err = store.Close(); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(coordinate, "unreleased-local-state")
			if err = os.WriteFile(sentinel, []byte("preserve"), 0600); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(registry, "registry.db")
			db, err := sql.Open("sqlite", path)
			if err != nil {
				t.Fatal(err)
			}
			statement := "DELETE FROM worktree_coordinates"
			if missing == "table" {
				statement = "DROP TABLE worktree_coordinates"
			}
			if _, err = db.Exec(statement); err != nil {
				t.Fatal(err)
			}
			if err = db.Close(); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, readOnly := range []bool{true, false} {
				opened, e := openRegistry(registry, "SharedProject", project, readOnly)
				if opened != nil {
					if closeErr := opened.Close(); closeErr != nil {
						t.Fatal(closeErr)
					}
					t.Fatal("unsupported layout accepted")
				}
				if e == nil || !strings.Contains(e.Error(), "unsupported capture coordinate layout") || !strings.Contains(e.Error(), "reinstall") {
					t.Fatalf("readOnly=%t: missing recovery context: %v", readOnly, e)
				}
			}
			if _, err = PreviewStateLocation(base, project); err == nil || !strings.Contains(err.Error(), "unsupported capture coordinate layout") {
				t.Fatalf("preview manufactured an unallocated scope: %v", err)
			}
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("unsupported database changed: %v", err)
			}
			body, err := os.ReadFile(sentinel)
			if err != nil || string(body) != "preserve" {
				t.Fatal("unsupported local state was overwritten")
			}
		})
	}
}
