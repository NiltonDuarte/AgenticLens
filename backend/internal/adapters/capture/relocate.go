package capture

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// RelocationOptions describe explicitly authorized maintenance, never a load-time fallback.
type RelocationOptions struct {
	Project, FromDirectory, ToBase string
	DryRun                         bool
}
type RelocationResult struct {
	Project     string `json:"project"`
	Name        string `json:"project_name"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Phase       string `json:"phase"`
	DryRun      bool   `json:"dry_run"`
	WorktreeID  int64  `json:"worktree_id,omitempty"`
}
type relocationJournal struct {
	ID, Project, Name, Destination, Phase string
}
type legacyTable struct{ name, columns, definition string }

// These are the inspected protocol-1 tables. Unknown columns/tables fail closed.
var relocationTables = []legacyTable{
	{"registrations", "identity,source,store,session_id,ticket,record", "identity TEXT PRIMARY KEY,source TEXT NOT NULL,store TEXT NOT NULL,session_id TEXT NOT NULL,ticket TEXT NOT NULL,record BLOB NOT NULL"},
	{"pending_starts", "store,parent,agent", "store TEXT,parent TEXT,agent TEXT,PRIMARY KEY(store,parent,agent)"},
	{"pending_children", "store,path,parent", "store TEXT,path TEXT,parent TEXT,PRIMARY KEY(store,path)"},
	{"discovery_cursor", "parent,path", "parent TEXT PRIMARY KEY,path TEXT NOT NULL"},
	{"child_cursor", "kind,parent,key,more", "kind TEXT,parent TEXT,key TEXT,more INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(kind,parent)"},
	{"project_proof_cursor", "singleton,identity", "singleton INTEGER PRIMARY KEY CHECK(singleton=1),identity TEXT NOT NULL"},
	{"publication_request", "singleton,generation,acknowledged,requested_at", "singleton INTEGER PRIMARY KEY CHECK(singleton=1),generation INTEGER NOT NULL,acknowledged INTEGER NOT NULL,requested_at INTEGER NOT NULL"},
	{"publication_cursor", "singleton,ticket", "singleton INTEGER PRIMARY KEY CHECK(singleton=1),ticket TEXT NOT NULL"},
	{"recovery_phase", "ticket,phase", "ticket TEXT PRIMARY KEY,phase INTEGER NOT NULL"},
	{"reconcile_cursor", "ticket,identity", "ticket TEXT PRIMARY KEY,identity TEXT NOT NULL"},
	{"capture_diagnostics", "sequence,ticket,phase,outcome,record", "sequence INTEGER PRIMARY KEY AUTOINCREMENT,ticket TEXT NOT NULL,phase TEXT NOT NULL,outcome TEXT NOT NULL,record BLOB NOT NULL"},
	{"capture_diagnostic_heads", "ticket,phase,owner,kind,sequence", "ticket TEXT NOT NULL,phase TEXT NOT NULL,owner TEXT NOT NULL,kind TEXT NOT NULL,sequence INTEGER NOT NULL,PRIMARY KEY(ticket,phase,owner,kind)"},
	{"capture_diagnostic_failures", "sequence", "sequence INTEGER PRIMARY KEY"},
	{"capture_diagnostic_index", "singleton,cursor,upper_sequence,known_max,complete", "singleton INTEGER PRIMARY KEY CHECK(singleton=1),cursor INTEGER NOT NULL,upper_sequence INTEGER NOT NULL,known_max INTEGER NOT NULL,complete INTEGER NOT NULL"},
}

// CheckStateRelocation prevents a configuration-only update from abandoning data.
func CheckStateRelocation(project, oldBase, newBase string) error {
	oldPath, err := ProjectDirectory(oldBase)
	if err != nil {
		return err
	}
	newPath, err := ProjectDirectory(newBase)
	if err != nil {
		return err
	}
	if oldPath == newPath {
		return nil
	}
	candidates := []string{filepath.Join(oldPath, "registry.db")}
	project, err = ProjectDirectory(project)
	if err != nil {
		return err
	}
	// Legacy names were physical-path hashes. Inspect only this exact worktree's leaf.
	legacyKey := legacyProjectKey(project)
	candidates = append(candidates, filepath.Join(oldPath, "projects", legacyKey, "registry.db"))
	// Explicit administrative preview: only the configured label is needed.
	// This decoder is never used as an operational configuration fallback.
	cfgPath, findErr := adaptersconfig.FindConfigFile(project)
	if findErr != nil {
		return findErr
	}
	if cfgPath != "" {
		name, nameErr := adaptersconfig.ReadProjectNameForMaintenance(cfgPath)
		if nameErr != nil {
			return nameErr
		}
		segment, nameErr := ProjectNameSegment(name)
		if nameErr != nil {
			return nameErr
		}
		candidates = append(candidates, filepath.Join(oldPath, "projects", segment, "registry.db"))
	}

	for _, path := range candidates {
		if _, e := os.Stat(path); e == nil {
			db, openErr := relocationDB(path, true)
			if openErr != nil {
				return openErr
			}
			var version int
			readErr := db.QueryRow("SELECT version FROM capture_meta").Scan(&version)
			closeErr := db.Close()
			if readErr != nil || closeErr != nil {
				return errors.Join(readErr, closeErr)
			}
			if version == 2 {
				return errors.New("changing a populated protocol-2 shared state location is unsupported; existing configuration and registry are preserved")
			}
			return errors.New("capture state exists at the previous location; use capture relocate-state for supported protocol-1 state before changing telemetry.capture.state_dir")
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
	}
	return nil
}

func legacyProjectKey(project string) string {
	sum := sha256.Sum256([]byte(project))
	return hex.EncodeToString(sum[:])
}

func relocationDB(path string, readOnly bool) (*sql.DB, error) {
	u, err := sqliteFileURI(path)
	if err != nil {
		return nil, err
	}
	options := "?mode=rw&_pragma=busy_timeout(250)"
	if readOnly {
		options = "?mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(250)"
	}
	db, err := sql.Open("sqlite", u+options)
	if err == nil {
		db.SetMaxOpenConns(1)
	}
	return db, err
}

// InspectRetainedState reads an explicitly selected name/worktree even after its
// checkout is gone. It never loads replacement configuration or creates a scope.
type RetainedStateInspection struct {
	Project                   string
	ProjectName               string
	WorktreeID                int64
	Registrations             []domain.CaptureRegistration
	RegistrationsContinuation bool
	Diagnostics               DiagnosticPage
	Publication               PublicationRequest
	Limitation                string
}

func InspectRetainedState(ctx context.Context, registry, name, project string) (result RetainedStateInspection, returnErr error) {
	if _, err := ProjectNameSegment(name); err != nil {
		return result, err
	}
	project, err := ProjectDirectory(project)
	if err != nil {
		return result, err
	}
	registry, err = adaptersconfig.ResolveCaptureStatePath(project, registry)
	if err != nil {
		return result, err
	}
	registry, err = ProjectDirectory(registry)
	if err != nil {
		return result, err
	}
	if err = regularRelocationFile(filepath.Join(registry, "registry.db")); err != nil {
		return result, err
	}
	db, err := relocationDB(filepath.Join(registry, "registry.db"), true)
	if err != nil {
		return result, err
	}
	defer func() { returnErr = errors.Join(returnErr, db.Close()) }()
	var version int
	if err = db.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
		return result, err
	}
	if version != RegistryProtocolVersion {
		return result, errors.New("unsupported retained registry protocol")
	}
	var actualName string
	if err = db.QueryRowContext(ctx, "SELECT name FROM project_identity WHERE singleton=1").Scan(&actualName); err != nil {
		return result, err
	}
	if actualName != name {
		return result, errors.New("retained registry project name differs from explicit selection")
	}
	var id int64
	if err = db.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&id); err != nil {
		return result, err
	}
	store := &Store{db: db, project: project, registry: registry, directory: filepath.Join(registry, "worktrees", fmt.Sprint(id)), worktreeID: id}
	result.Project = project
	result.ProjectName = name
	result.WorktreeID = id
	result.Registrations, result.RegistrationsContinuation, err = store.ListBounded(ctx, 256)
	if err != nil {
		return result, err
	}
	result.Diagnostics, err = store.ReadDiagnostics(ctx, DiagnosticReadOptions{Limit: 256, MaxBytes: 1024 * 1024})
	if err != nil {
		return result, err
	}
	result.Publication, err = store.PublicationRequest(ctx)
	result.Limitation = "Explicit retained observations only. No configuration fallback, current worktree ownership verification, source read, publication, binding or directory recreation was performed."
	return result, err
}

type relocationInstaller func(context.Context, InstallOptions) (InstallResult, error)

// RelocateState always uses the real installer. The private shell boundary lets
// interruption tests exercise real journaled file writes without global hooks.
func RelocateState(ctx context.Context, options RelocationOptions) (RelocationResult, error) {
	return relocateState(ctx, options, Install)
}
func relocateState(ctx context.Context, options RelocationOptions, installer relocationInstaller) (result RelocationResult, returnErr error) {
	project, err := adaptersconfig.ProjectRoot(options.Project)
	if err != nil {
		return result, err
	}
	project, err = ProjectDirectory(project)
	if err != nil {
		return result, err
	}
	cfg, err := adaptersconfig.Load(project)
	if err != nil {
		return result, err
	}
	source, err := adaptersconfig.ResolveCaptureStatePath(project, options.FromDirectory)
	if err != nil {
		return result, err
	}
	source, err = ProjectDirectory(source)
	if err != nil {
		return result, err
	}
	base, err := adaptersconfig.ResolveCaptureStatePath(project, options.ToBase)
	if err != nil {
		return result, err
	}
	destination, err := namedProjectState(base, cfg.ProjectName, project)
	if err != nil {
		return result, err
	}
	if source == destination {
		return result, errors.New("source and destination capture registry are identical")
	}
	if err = regularRelocationFile(filepath.Join(source, "registry.db")); err != nil {
		return result, err
	}
	db, err := relocationDB(filepath.Join(source, "registry.db"), options.DryRun)
	if err != nil {
		return result, err
	}
	defer func() { returnErr = errors.Join(returnErr, db.Close()) }()
	version, err := validateLegacyRegistry(ctx, db, project)
	if err != nil {
		return result, err
	}
	result = RelocationResult{Project: project, Name: cfg.ProjectName, Source: source, Destination: destination, DryRun: options.DryRun, Phase: "preview"}
	journalPath := filepath.Join(source, "relocation.json")
	journal, err := readRelocationJournal(journalPath)
	if err != nil {
		return result, err
	}
	if journal.ID != "" && (journal.Project != project || journal.Name != cfg.ProjectName || journal.Destination != destination) {
		return result, errors.New("existing relocation journal belongs to a different project or destination")
	}
	if version != 1 && (version != -1 || journal.ID == "") {
		return result, errors.New("unsupported capture registry protocol for relocation")
	}
	if version == -1 && journal.Phase == "completed" {
		result.Phase = "completed"
		if store, e := openRegistry(destination, cfg.ProjectName, project, true); e == nil {
			result.WorktreeID = store.worktreeID
			return result, store.Close()
		} else {
			return result, e
		}
	}
	updatePath, update, err := prepareRelocationConfiguration(project, options.ToBase)
	if err != nil {
		return result, err
	}
	manifests, err := relocationManifests(source, project)
	if err != nil {
		return result, err
	}
	if err = previewRelocationDestination(ctx, destination, cfg.ProjectName, project, journal.ID); err != nil {
		return result, err
	}
	if options.DryRun {
		return result, nil
	}
	worker, err := AcquireWorker(source)
	if err != nil {
		return result, err
	}
	if worker == nil {
		return result, errors.New("capture worker is active; retry relocation after its bounded pass finishes")
	}
	defer func() { returnErr = errors.Join(returnErr, worker.Close()) }()
	locks, err := acquireRelocationPublications(ctx, db, project)
	if err != nil {
		return result, err
	}
	defer func() {
		for i := len(locks) - 1; i >= 0; i-- {
			returnErr = errors.Join(returnErr, locks[i].Close())
		}
	}()
	if journal.ID == "" {
		token := make([]byte, 16)
		if _, err = rand.Read(token); err != nil {
			return result, err
		}
		journal = relocationJournal{ID: hex.EncodeToString(token), Project: project, Name: cfg.ProjectName, Destination: destination, Phase: "prepared"}
		if err = writeRelocationJournal(journalPath, journal); err != nil {
			return result, err
		}
	}
	// The new scope and its pending fence commit together, before retiring
	// the source. Existing physical ownership and conflicting data are checked
	// under the same writer reservation; the store holds its worker lease.
	store, err := openRegistryForRelocation(destination, cfg.ProjectName, project, journal.ID)
	if err != nil {
		return result, err
	}
	defer func() { returnErr = errors.Join(returnErr, store.Close()) }()
	result.WorktreeID = store.worktreeID
	if version == 1 {
		if err = backupRelocationRegistry(ctx, db, filepath.Join(source, "relocation-prior.db")); err != nil {
			return result, err
		}
		if err = fenceLegacyRegistry(ctx, db, project); err != nil {
			return result, err
		}
		journal.Phase = "fenced"
		if err = writeRelocationJournal(journalPath, journal); err != nil {
			return result, err
		}
	}
	if err = backupRelocationRegistry(ctx, db, filepath.Join(source, "relocation-fenced.db")); err != nil {
		return result, err
	}

	if err = importRelocation(ctx, store, source, journal.ID); err != nil {
		return result, fmt.Errorf("source retained and fenced; retry relocation: %w", err)
	}
	journal.Phase = "imported"
	if err = writeRelocationJournal(journalPath, journal); err != nil {
		return result, err
	}
	if err = adaptersconfig.WriteCaptureConfigUpdate(updatePath, update); err != nil {
		return result, fmt.Errorf("destination remains pending; configuration activation failed: %w", err)
	}
	for _, m := range manifests {
		target := manifestPath(store.directory, m.Source)
		body, readErr := os.ReadFile(manifestPath(source, m.Source))
		if readErr != nil {
			return result, readErr
		}
		if existing, e := os.ReadFile(target); e == nil {
			var current Manifest
			if e = json.Unmarshal(existing, &current); e != nil {
				return result, e
			}
			if current.Protocol == ManifestProtocolVersion && current.Repository == project && current.StateDir == store.directory {
				body = existing
			}
			if string(existing) != string(body) {
				return result, errors.New("destination installation manifest conflicts; source retained")
			}
		} else if !errors.Is(e, os.ErrNotExist) {
			return result, e
		} else if err = writeRelocationFile(target, body); err != nil {
			return result, err
		}
		{
			if _, err = installer(ctx, InstallOptions{Repository: project, StateDir: store.directory, Source: m.Source, SourceStore: m.SourceStore, Wrapper: m.Wrapper, Remove: !m.Active, AllowLegacyManifest: true, LegacyStateDir: source}); err != nil {
				return result, fmt.Errorf("destination remains pending; owned integration refresh failed: %w", err)
			}
		}
	}
	if err = copyRelocationEvidence(source, store.directory); err != nil {
		return result, err
	}
	if err = activateRelocation(ctx, store, journal.ID); err != nil {
		return result, err
	}
	journal.Phase = "completed"
	if err = writeRelocationJournal(journalPath, journal); err != nil {
		return result, err
	}
	result.Phase = "completed"
	return result, nil
}

func regularRelocationFile(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("relocation requires a regular file: %s", path)
	}
	return nil
}
func readRelocationJournal(path string) (relocationJournal, error) {
	var j relocationJournal
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return j, nil
	}
	if err != nil {
		return j, err
	}
	if len(data) > 64*1024 {
		return j, errors.New("relocation journal exceeds bound")
	}
	err = json.Unmarshal(data, &j)
	return j, err
}
func writeRelocationJournal(path string, j relocationJournal) error {
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return writeRelocationFile(path, append(data, '\n'))
}
func writeRelocationFile(path string, data []byte) (returnErr error) {
	f, err := os.CreateTemp(filepath.Dir(path), ".relocation-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, e)
		}
	}()
	if err = f.Chmod(0600); err != nil {
		return errors.Join(err, f.Close())
	}
	if _, err = f.Write(data); err != nil {
		return errors.Join(err, f.Close())
	}
	if err = f.Sync(); err != nil {
		return errors.Join(err, f.Close())
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	return nil
}
func backupRelocationRegistry(ctx context.Context, db *sql.DB, path string) error {
	if _, err := os.Lstat(path); err == nil {
		if err = regularRelocationFile(path); err != nil {
			return err
		}
		backup, e := relocationDB(path, true)
		if e != nil {
			return e
		}
		var owner, expected, check string
		e = backup.QueryRowContext(ctx, "SELECT directory FROM project_owner WHERE singleton=1").Scan(&owner)
		if e == nil {
			e = backup.QueryRowContext(ctx, "PRAGMA quick_check").Scan(&check)
		}
		if e == nil {
			e = db.QueryRowContext(ctx, "SELECT directory FROM project_owner WHERE singleton=1").Scan(&expected)
		}
		closeErr := backup.Close()
		if e != nil || closeErr != nil {
			return errors.Join(e, closeErr)
		}
		if owner != expected || check != "ok" {
			return errors.New("existing relocation backup is invalid or belongs to another worktree")
		}
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if _, err = db.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return errors.Join(err, os.Remove(path))
	}
	return nil
}

func tableExists(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", name).Scan(&count)
	return count == 1, err
}
func validateLegacyRegistry(ctx context.Context, db *sql.DB, project string) (int, error) {
	var version int
	if err := db.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
		return 0, err
	}
	if version == 2 {
		return 0, errors.New("protocol-2 shared state relocation is unsupported; existing registry is preserved")
	}
	if version != 1 && version != -1 {
		return 0, errors.New("unsupported source protocol; registry preserved")
	}
	var owner string
	if err := db.QueryRowContext(ctx, "SELECT directory FROM project_owner WHERE singleton=1").Scan(&owner); err != nil {
		return 0, fmt.Errorf("legacy physical ownership unavailable: %w", err)
	}
	if owner != project {
		return 0, errors.New("legacy registry belongs to a different physical worktree")
	}
	var unexpected int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='view' OR (type='trigger' AND name NOT LIKE 'relocation_fence_%')").Scan(&unexpected); err != nil {
		return 0, err
	}
	if unexpected != 0 {
		return 0, errors.New("unknown legacy views or triggers; explicit recovery required")
	}
	var triggerCount int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='trigger'").Scan(&triggerCount); err != nil {
		return 0, err
	}
	if version == 1 && triggerCount != 0 || version == -1 && triggerCount != (len(relocationTables)+2)*3 {
		return 0, errors.New("unknown legacy trigger layout; explicit recovery required")
	}
	var metadataRows int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM capture_meta").Scan(&metadataRows); err != nil || metadataRows != 1 {
		return 0, errors.New("legacy protocol metadata is ambiguous")
	}
	allowed := map[string]bool{"capture_meta": true, "project_owner": true, "sqlite_sequence": true}
	for _, table := range relocationTables {
		allowed[table.name] = true
	}
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			break
		}
		if !allowed[name] {
			err = fmt.Errorf("unknown legacy table %s; explicit recovery required", name)
			break
		}
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return 0, err
	}
	for _, table := range relocationTables {
		exists, e := tableExists(ctx, db, table.name)
		if e != nil {
			return 0, e
		}
		if !exists {
			continue
		}
		rows, e := db.QueryContext(ctx, "PRAGMA table_info("+table.name+")")
		if e != nil {
			return 0, e
		}
		columns := []string{}
		for rows.Next() {
			var position, notnull, pk int
			var name, kind string
			var defaultValue sql.NullString
			if e = rows.Scan(&position, &name, &kind, &notnull, &defaultValue, &pk); e != nil {
				break
			}
			columns = append(columns, name)
		}
		e = errors.Join(e, rows.Err(), rows.Close())
		if e != nil {
			return 0, e
		}
		if strings.Join(columns, ",") != table.columns {
			return 0, fmt.Errorf("unknown legacy columns for %s", table.name)
		}
	}
	return version, nil
}
func fenceLegacyRegistry(ctx context.Context, db *sql.DB, project string) (returnErr error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, conn.Close()) }()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("source writer is busy; retry relocation: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_, e := conn.ExecContext(context.Background(), "ROLLBACK")
			returnErr = errors.Join(returnErr, e)
		}
	}()
	var version int
	if err = conn.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
		return err
	}
	if version != 1 {
		return errors.New("source protocol changed before relocation fence")
	}
	var owner string
	if err = conn.QueryRowContext(ctx, "SELECT directory FROM project_owner WHERE singleton=1").Scan(&owner); err != nil {
		return err
	}
	if owner != project {
		return errors.New("source ownership changed before relocation")
	}
	for _, table := range relocationTables {
		if _, err = conn.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS "+table.name+"("+table.definition+")"); err != nil {
			return err
		}
	}
	if _, err = conn.ExecContext(ctx, "UPDATE capture_meta SET version=-1"); err != nil {
		return err
	}
	names := []string{"capture_meta", "project_owner"}
	for _, table := range relocationTables {
		names = append(names, table.name)
	}
	for _, name := range names {
		for _, action := range []string{"INSERT", "UPDATE", "DELETE"} {
			query := "CREATE TRIGGER IF NOT EXISTS relocation_fence_" + name + "_" + action + " BEFORE " + action + " ON " + name + " BEGIN SELECT RAISE(ABORT,'capture state relocated; refresh owned hooks and use the configured destination'); END"
			if _, err = conn.ExecContext(ctx, query); err != nil {
				return err
			}
		}
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}
func prepareRelocationConfiguration(project, replacement string) (string, adaptersconfig.CaptureConfigUpdate, error) {
	path, err := adaptersconfig.FindConfigFile(project)
	if err != nil {
		return "", adaptersconfig.CaptureConfigUpdate{}, err
	}
	local := filepath.Join(project, ".agenticlens-config.local.yaml")
	if _, e := os.Stat(local); e == nil {
		preview, prepareErr := adaptersconfig.PrepareCaptureConfigUpdate(local, "")
		if prepareErr != nil {
			return "", adaptersconfig.CaptureConfigUpdate{}, prepareErr
		}
		if preview.PreviousStateDir != "" {
			path = local
		}
	} else if !errors.Is(e, os.ErrNotExist) {
		return "", adaptersconfig.CaptureConfigUpdate{}, e
	}
	update, err := adaptersconfig.PrepareCaptureConfigUpdate(path, replacement)
	return path, update, err
}
func relocationManifests(source, project string) ([]Manifest, error) {
	result := []Manifest{}
	for _, kind := range []string{"codex", "opencode"} {
		path := manifestPath(source, kind)
		body, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		var m Manifest
		if err = json.Unmarshal(body, &m); err != nil {
			return nil, err
		}
		if m.Protocol != 1 || m.Source != kind || m.Repository != project {
			return nil, errors.New("legacy installation ownership is ambiguous")
		}
		recordedState, stateErr := ProjectDirectory(m.StateDir)
		if m.StateDir == "" || stateErr != nil || recordedState != source {
			return nil, errors.New("legacy installation state directory differs from selected source")
		}
		result = append(result, m)
	}
	return result, nil
}
func previewRelocationDestination(ctx context.Context, path, name, project, id string) (returnErr error) {
	if _, err := os.Stat(filepath.Join(path, "registry.db")); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	db, err := relocationDB(filepath.Join(path, "registry.db"), true)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, db.Close()) }()
	var version int
	if err = db.QueryRowContext(ctx, "SELECT version FROM capture_meta").Scan(&version); err != nil {
		return err
	}
	if version != RegistryProtocolVersion {
		return errors.New("destination registry protocol is incompatible")
	}
	var actual string
	if err = db.QueryRowContext(ctx, "SELECT name FROM project_identity WHERE singleton=1").Scan(&actual); err != nil {
		return err
	}
	if actual != name {
		return errors.New("destination project name conflicts")
	}
	var scope int64
	if err = db.QueryRowContext(ctx, "SELECT id FROM worktrees WHERE directory=?", project).Scan(&scope); errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	}
	if err = validateWorktreeOwnership(db, scope, project); err != nil {
		return err
	}
	exists, err := tableExists(ctx, db, "worktree_relocations")
	if err != nil {
		return err
	}
	if exists {
		var existingID string
		err = db.QueryRowContext(ctx, "SELECT migration_id FROM worktree_relocations WHERE worktree_id=?", scope).Scan(&existingID)
		if err == nil {
			if existingID == id && id != "" {
				return nil
			}
			return errors.New("destination belongs to another relocation")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	for _, table := range relocationDestinationTables() {
		present, tableErr := tableExists(ctx, db, table.name)
		if tableErr != nil {
			return tableErr
		}
		if !present {
			continue
		}
		if table.name == "capture_diagnostic_heads" || table.name == "capture_diagnostic_index" {
			continue
		}
		var count int
		if err = db.QueryRowContext(ctx, "SELECT count(*) FROM "+table.name+" WHERE worktree_id=?", scope).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return errors.New("destination worktree already contains capture state; refusing replacement")
		}
	}
	return nil
}

func acquireRelocationPublications(ctx context.Context, db *sql.DB, project string) ([]*PublicationLock, error) {
	exists, err := tableExists(ctx, db, "capture_diagnostics")
	if err != nil || !exists {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, "SELECT record FROM capture_diagnostics WHERE phase='publication'")
	if err != nil {
		return nil, err
	}
	paths := map[string]bool{}
	for rows.Next() {
		var data []byte
		var d domain.CaptureDiagnostic
		if err = rows.Scan(&data); err != nil {
			break
		}
		if err = json.Unmarshal(data, &d); err != nil {
			break
		}
		if d.Destination != "" {
			if err = ValidateDestination(project, d.Destination); err != nil {
				break
			}
			paths[d.Destination] = true
		}
	}
	err = errors.Join(err, rows.Err(), rows.Close())
	if err != nil {
		return nil, err
	}
	ordered := []string{}
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)
	locks := []*PublicationLock{}
	for _, path := range ordered {
		if _, e := os.Stat(filepath.Dir(path)); errors.Is(e, os.ErrNotExist) {
			continue
		} else if e != nil {
			return nil, e
		}
		lock, e := AcquirePublication(ctx, path)
		if e != nil {
			for _, prior := range locks {
				e = errors.Join(e, prior.Close())
			}
			return nil, e
		}
		locks = append(locks, lock)
	}
	return locks, nil
}

// validateRelocationIdentity bypasses only activation, and only for the
// journal's exact relocation. Physical ownership is checked independently.
func validateRelocationIdentity(db *sql.DB, scope int64, id string) error {
	exists, err := tableExists(context.Background(), db, "worktree_relocations")
	if err != nil || !exists {
		return err
	}
	var recorded, state string
	err = db.QueryRow("SELECT migration_id,state FROM worktree_relocations WHERE worktree_id=?", scope).Scan(&recorded, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if recorded != id || id == "" {
		return errors.New("destination belongs to another relocation")
	}
	if state != "pending" && state != "pending-imported" && state != "active" {
		return errors.New("destination relocation state is invalid")
	}
	return nil
}

// Auxiliary receipts were introduced with schema2. They never belong to a
// schema1 import, but are operational destination writes and must be fenced.
func relocationDestinationTables() []legacyTable {
	tables := make([]legacyTable, 0, len(relocationTables)+1)
	tables = append(tables, relocationTables...)
	return append(tables, legacyTable{name: "capture_auxiliary_receipts"})
}

// reserveRelocationScopeTx is part of scope/ownership allocation's immediate
// transaction. No operationally visible scope exists without its write fence.
func reserveRelocationScopeTx(tx *sql.Tx, scope int64, id string) error {
	if id == "" {
		return errors.New("relocation identity is required")
	}
	var recorded, state string
	err := tx.QueryRow("SELECT migration_id,state FROM worktree_relocations WHERE worktree_id=?", scope).Scan(&recorded, &state)
	if err == nil {
		if recorded != id {
			return errors.New("destination belongs to another relocation")
		}
		if state != "pending" && state != "pending-imported" && state != "active" {
			return errors.New("destination relocation state is invalid")
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	for _, table := range relocationDestinationTables() {
		query := "SELECT count(*) FROM " + table.name + " WHERE worktree_id=?"
		if table.name == "capture_diagnostic_heads" {
			query += " AND NOT(ticket='' AND phase='index' AND owner=':attempt-index' AND kind='latest' AND sequence=0)"
		}
		if table.name == "capture_diagnostic_index" {
			query += " AND (cursor!=0 OR upper_sequence!=0 OR known_max!=0)"
		}
		var count int
		if err = tx.QueryRow(query, scope).Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return fmt.Errorf("destination scope already contains %s; refusing replacement", table.name)
		}
	}
	if _, err = tx.Exec("INSERT INTO worktree_relocations(worktree_id,migration_id,state)VALUES(?,?,'pending')", scope, id); err != nil {
		return err
	}
	for _, table := range relocationDestinationTables() {
		for _, action := range []string{"INSERT", "UPDATE", "DELETE"} {
			row := "NEW"
			if action == "DELETE" {
				row = "OLD"
			}
			query := "CREATE TRIGGER IF NOT EXISTS relocation_pending_" + table.name + "_" + action + " BEFORE " + action + " ON " + table.name + " WHEN EXISTS(SELECT 1 FROM worktree_relocations WHERE worktree_id=" + row + ".worktree_id AND state!='active') BEGIN SELECT RAISE(ABORT,'capture worktree relocation pending; retry explicit relocation'); END"
			if _, err = tx.Exec(query); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRelocationImportTx(ctx context.Context, tx *sql.Tx, store *Store, id string) (string, error) {
	var recordedOwnership string
	if err := tx.QueryRowContext(ctx, "SELECT ownership FROM worktree_ownership WHERE worktree_id=?", store.worktreeID).Scan(&recordedOwnership); err != nil {
		return "", err
	}
	current, err := currentWorktreeOwnership(store.project)
	if err != nil {
		return "", err
	}
	same, err := sameWorktreeOwnership(recordedOwnership, current)
	if err != nil {
		return "", err
	}
	if !same {
		return "", errors.New("physical worktree or Git ownership changed; explicit recovery required")
	}
	var recorded, state string
	if err = tx.QueryRowContext(ctx, "SELECT migration_id,state FROM worktree_relocations WHERE worktree_id=?", store.worktreeID).Scan(&recorded, &state); err != nil {
		return "", err
	}
	if id == "" || recorded != id {
		return "", errors.New("destination relocation identity differs from source journal")
	}
	if state != "pending" && state != "pending-imported" && state != "active" {
		return "", errors.New("destination relocation state is invalid")
	}
	return state, nil
}

// Activation rechecks the same ownership under its writer reservation. A
// checkout replaced while configuration/hooks were refreshed remains pending.
func activateRelocation(ctx context.Context, store *Store, id string) (returnErr error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	done := false
	defer func() {
		if !done {
			returnErr = errors.Join(returnErr, tx.Rollback())
		}
	}()
	state, err := validateRelocationImportTx(ctx, tx, store, id)
	if err != nil {
		return err
	}
	if state != "pending-imported" && state != "active" {
		return errors.New("relocation import is incomplete; activation refused")
	}
	if _, err = tx.ExecContext(ctx, "UPDATE worktree_relocations SET state='active' WHERE worktree_id=? AND migration_id=?", store.worktreeID, id); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	done = true
	return nil
}

func importRelocation(ctx context.Context, store *Store, sourceDir, id string) (returnErr error) {
	// Reservation and the pending fence must already be durable before source
	// retirement. Import never creates a previously operational destination.
	check, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	state, err := validateRelocationImportTx(ctx, check, store, id)
	if err != nil {
		return errors.Join(err, check.Rollback())
	}
	if err = check.Commit(); err != nil {
		return err
	}
	if state == "active" || state == "pending-imported" {
		return nil
	}

	u, err := sqliteFileURI(filepath.Join(sourceDir, "registry.db"))
	if err != nil {
		return err
	}
	if _, err = store.db.ExecContext(ctx, "ATTACH DATABASE ? AS relocation_source", u+"?mode=ro"); err != nil {
		return err
	}
	defer func() {
		_, e := store.db.ExecContext(context.Background(), "DETACH DATABASE relocation_source")
		returnErr = errors.Join(returnErr, e)
	}()
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	done := false
	defer func() {
		if !done {
			returnErr = errors.Join(returnErr, tx.Rollback())
		}
	}()
	state, err = validateRelocationImportTx(ctx, tx, store, id)
	if err != nil {
		return err
	}
	if state == "active" || state == "pending-imported" {
		err = tx.Commit()
		done = err == nil
		return err
	}
	// This active state is private to the import transaction. Other writers
	// remain fenced until final activation commits after configuration/setup.
	if _, err = tx.ExecContext(ctx, "UPDATE worktree_relocations SET state='active' WHERE worktree_id=?", store.worktreeID); err != nil {
		return err
	}
	var offset int64
	if err = tx.QueryRowContext(ctx, "SELECT coalesce(max(sequence),0) FROM capture_diagnostics").Scan(&offset); err != nil {
		return err
	}
	for _, table := range relocationTables {
		expressions := strings.Split(table.columns, ",")
		for index, column := range expressions {
			isSequence := column == "sequence" && (table.name == "capture_diagnostics" || table.name == "capture_diagnostic_heads" || table.name == "capture_diagnostic_failures")
			isIndex := table.name == "capture_diagnostic_index" && (column == "cursor" || column == "upper_sequence" || column == "known_max")
			if isSequence || isIndex {
				expressions[index] = "CASE WHEN " + column + ">0 THEN " + column + "+" + fmt.Sprint(offset) + " ELSE " + column + " END"
			}
		}
		query := "INSERT INTO main." + table.name + "(worktree_id," + table.columns + ") SELECT ?," + strings.Join(expressions, ",") + " FROM relocation_source." + table.name
		// New empty scope initialization creates only derived empty metadata.
		if table.name == "capture_diagnostic_index" || table.name == "capture_diagnostic_heads" {
			if _, err = tx.ExecContext(ctx, "DELETE FROM main."+table.name+" WHERE worktree_id=?", store.worktreeID); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, query, store.worktreeID); err != nil {
			return fmt.Errorf("importing %s: %w", table.name, err)
		}
	}
	if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO capture_diagnostic_index(worktree_id,singleton,cursor,upper_sequence,known_max,complete) SELECT ?,1,0,coalesce(max(sequence),0),coalesce(max(sequence),0),CASE WHEN count(*)=0 THEN 1 ELSE 0 END FROM capture_diagnostics WHERE worktree_id=?", store.worktreeID, store.worktreeID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE worktree_relocations SET state='pending-imported' WHERE worktree_id=?", store.worktreeID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	done = true
	return nil
}
func copyRelocationEvidence(source, destination string) error {
	for _, name := range []string{"worker-status.json", "worker.log"} {
		path := filepath.Join(source, name)
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		if err := regularRelocationFile(path); err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, "prior-"+name)
		if existing, e := os.ReadFile(target); e == nil {
			if string(existing) != string(body) {
				return errors.New("retained prior coordination evidence conflicts")
			}
			continue
		} else if !errors.Is(e, os.ErrNotExist) {
			return e
		}
		if err = writeRelocationFile(target, body); err != nil {
			return err
		}
	}
	return nil
}
