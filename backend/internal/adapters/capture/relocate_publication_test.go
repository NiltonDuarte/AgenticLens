package capture_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters"
	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// Relocating capture storage preserves public pending artifacts and their
// separate private operational journal. It must not copy transaction state into
// the capture registry or accept a partially replaced ledger as source evidence.
func TestRelocationPreservesActualRecoverablePublicationJournal(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	project, source, base := filepath.Join(root, "project"), filepath.Join(root, "legacy-state"), filepath.Join(root, "home-state")
	for _, path := range []string{project, source} {
		if err := os.MkdirAll(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	var err error
	project, err = capture.ProjectDirectory(project)
	if err != nil {
		t.Fatal(err)
	}
	config, err := fixtureSharedYAML(t, "Publication migration", "records")
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(project, ".agenticlens-config.yaml")
	if err = os.WriteFile(configPath, config, 0600); err != nil {
		t.Fatal(err)
	}
	localPath := filepath.Join(project, ".agenticlens-config.local.yaml")
	if err = os.WriteFile(localPath, []byte("agenticlens:\n  state_dir: publication-state\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: before-update\n"), 0600); err != nil {
		t.Fatal(err)
	}
	update, err := adaptersconfig.PrepareCaptureConfigUpdate(localPath, source)
	if err != nil {
		t.Fatal(err)
	}
	if err = adaptersconfig.WriteCaptureConfigUpdate(localPath, update); err != nil {
		t.Fatal(err)
	}

	schema, ticket, title, ticketType := domain.CurrentSchemaVersion, "T-1", "Accepted publication before interruption", domain.ConventionalTicketType("feat")
	start, end := "2026-01-02T03:04:05Z", "2026-01-02T03:05:05Z"
	owner := domain.SessionFact{ID: "accepted-publication-owner", Source: domain.SourceCodex, User: "migration-fixture",
		Invocations: []domain.InvocationFact{{Role: "coder", Model: "recorded-model", StartedAt: &start, CompletedAt: &end, ActiveDurationSeconds: 60,
			Turns: 2, Tokens: domain.InvocationTokens{Input: 31, Output: 7, CacheRead: 13, PeakContext: 44}, CostEvidence: domain.ObservedCost(0)}},
		Tools: []domain.SessionToolFact{{Name: "exec", Calls: 2, Errors: 1}}, Anomalies: []domain.Anomaly{}}
	accepted := domain.AgenticFactLedger{SchemaVersion: &schema, TicketID: &ticket, TicketTitle: &title, TicketType: &ticketType,
		StartedAt: &start, CompletedAt: &end, UpdatedAt: end, Sessions: []domain.SessionFact{owner}, Anomalies: []domain.Anomaly{}}
	acceptedBody, err := json.Marshal(accepted)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = adapters.DecodeExistingFactLedger(acceptedBody); err != nil {
		t.Fatalf("accepted fixture is not a valid persisted ledger: %v", err)
	}
	output := filepath.Join(project, "records", ticket)
	if err = os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	ledger, summary := filepath.Join(output, "agentic_metrics.json"), filepath.Join(output, "SUMMARY.md")
	acceptedSummary := []byte("Accepted source facts before interrupted publication.\n")
	for path, body := range map[string][]byte{ledger: acceptedBody, summary: acceptedSummary} {
		if err = os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	members := []string{ledger, summary}
	publication, err := capture.AcquirePublication(context.Background(), ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err = adapters.BeginPublication(ledger, members); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ledger, []byte("uncommitted partial replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = publication.Close(); err != nil {
		t.Fatal(err)
	}
	journal := ledger + ".publication-pending"
	privateDirectory := filepath.Join(project, "publication-state", "publication")
	privateBefore := publicationPrivateFiles(t, privateDirectory)
	before := map[string][]byte{}
	identities := map[string]os.FileInfo{}
	for _, path := range []string{ledger, summary, journal} {
		before[path], err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		identities[path], err = os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err = adapters.CheckPublication(ledger); err == nil {
		t.Fatal("fixture did not reach a recoverable interrupted publication")
	}

	// This independently states the supported legacy wire schema rather than
	// copying the importer's table list or generating expected imports from it.
	legacy, err := sql.Open("sqlite", filepath.Join(source, "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer legacy.Close()
	legacy.SetMaxOpenConns(1)
	if _, err = legacy.Exec("PRAGMA journal_mode=WAL; CREATE TABLE capture_meta(version INTEGER NOT NULL); INSERT INTO capture_meta VALUES(1); CREATE TABLE project_owner(singleton INTEGER PRIMARY KEY,directory TEXT NOT NULL); CREATE TABLE registrations(identity TEXT PRIMARY KEY,source TEXT NOT NULL,store TEXT NOT NULL,session_id TEXT NOT NULL,ticket TEXT NOT NULL,record BLOB NOT NULL); CREATE TABLE publication_request(singleton INTEGER PRIMARY KEY CHECK(singleton=1),generation INTEGER NOT NULL,acknowledged INTEGER NOT NULL,requested_at INTEGER NOT NULL); INSERT INTO publication_request VALUES(1,9,4,123); CREATE TABLE capture_diagnostics(sequence INTEGER PRIMARY KEY AUTOINCREMENT,ticket TEXT NOT NULL,phase TEXT NOT NULL,outcome TEXT NOT NULL,record BLOB NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec("INSERT INTO project_owner VALUES(1,?)", project); err != nil {
		t.Fatal(err)
	}
	id := domain.CaptureIdentity{Source: owner.Source, Store: filepath.Join(root, "unavailable-source-home"), SessionID: owner.ID}
	registration := domain.CaptureRegistration{Identity: id, ProjectDirectory: project, ProjectProof: domain.CaptureProjectVerified,
		TicketID: ticket, TicketTitle: title, TicketType: string(ticketType), Binding: "explicit_agent_response", Request: "accepted-binding",
		State: domain.CaptureCurrent, UpdatedAt: end, RefreshedAt: end, Revision: 4096, SourceFingerprint: "4096:1767323105000000000"}
	record, err := json.Marshal(struct {
		Registration domain.CaptureRegistration `json:"registration"`
		Path         string                     `json:"path"`
		Snapshot     []domain.SessionFact       `json:"snapshot,omitempty"`
	}{registration, filepath.Join(id.Store, "sessions", "accepted.jsonl"), []domain.SessionFact{owner}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec("INSERT INTO registrations VALUES(?,?,?,?,?,?)", capture.IdentityKey(id), string(id.Source), id.Store, id.SessionID, ticket, record); err != nil {
		t.Fatal(err)
	}
	diagnostic, err := json.Marshal(domain.CaptureDiagnostic{TicketID: ticket, Phase: "publication", Outcome: "staging", Destination: ledger, Summary: summary, Generation: 9, ObservedAt: end})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = legacy.Exec("INSERT INTO capture_diagnostics(ticket,phase,outcome,record)VALUES(?,?,?,?)", ticket, "publication", "staging", diagnostic); err != nil {
		t.Fatal(err)
	}
	var result capture.RelocationResult
	for attempt := 0; attempt < 2; attempt++ {
		result, err = capture.RelocateState(context.Background(), capture.RelocationOptions{Project: project, FromDirectory: source, ToBase: base})
		if err != nil {
			t.Fatalf("relocation/retry %d: %v", attempt, err)
		}
		for path, body := range before {
			after, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(after, body) {
				t.Fatalf("relocation changed public member/journal %s: %v", path, err)
			}
			info, err := os.Stat(path)
			if err != nil || !os.SameFile(identities[path], info) {
				t.Fatalf("relocation replaced exact public path %s: %v", path, err)
			}
		}
	}
	if err = filepath.WalkDir(result.Destination, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && entry.Name() == filepath.Base(journal) {
			t.Fatalf("relocation copied a publication journal into private state at %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	privateAfter := publicationPrivateFiles(t, privateDirectory)
	if !reflect.DeepEqual(privateAfter, privateBefore) {
		t.Fatal("capture relocation changed private operational publication state")
	}
	current, err := capture.OpenProject(result.Destination, project)
	if err != nil {
		t.Fatal(err)
	}
	defer current.Close()
	rows, err := current.List(context.Background())
	if err != nil || len(rows) != 1 || rows[0].Identity != id || rows[0].Binding != registration.Binding || !reflect.DeepEqual(rows[0].Snapshot, []domain.SessionFact{owner}) {
		t.Fatalf("accepted source evidence changed during pending publication: %+v %v", rows, err)
	}
	pending, err := current.PublicationRequest(context.Background())
	if err != nil || pending.Generation != 9 || pending.Acknowledged != 4 {
		t.Fatalf("interrupted publication was falsely acknowledged: %+v %v", pending, err)
	}
	publication, err = capture.AcquirePublication(context.Background(), ledger)
	if err != nil {
		t.Fatal(err)
	}
	defer publication.Close()
	if recovered, err := adapters.RecoverPublication(ledger, members); err != nil || !recovered {
		t.Fatalf("original public journal no longer recovers: %v %v", recovered, err)
	}
	for path, want := range map[string][]byte{ledger: acceptedBody, summary: acceptedSummary} {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(after, want) {
			t.Fatalf("accepted public artifact not recovered at %s: %v", path, err)
		}
	}
	if err = adapters.CheckPublication(ledger); err != nil {
		t.Fatal(err)
	}
	if recovered, err := adapters.RecoverPublication(ledger, members); err != nil || recovered {
		t.Fatalf("public recovery not idempotent: %v %v", recovered, err)
	}
}

func publicationPrivateFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[rel] = string(data)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return files
}
