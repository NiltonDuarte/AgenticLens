package main_test

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// setOpenCodeFixtureProject writes the supported native session directory column;
// a command flag alone must never count as source ownership evidence.
func setOpenCodeFixtureProject(t *testing.T, path, project string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM pragma_table_info('session') WHERE name='directory'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		if _, err = db.Exec(`ALTER TABLE session ADD COLUMN directory TEXT`); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(`UPDATE session SET directory=?`, project); err != nil {
		t.Fatal(err)
	}
}
func setOpenCodeFixtureProjectIfSelected(t *testing.T, cmd *exec.Cmd, project string) {
	t.Helper()
	for i, arg := range cmd.Args {
		if arg == "--db-path" && i+1 < len(cmd.Args) {
			path := cmd.Args[i+1]
			if _, err := os.Stat(path); err == nil {
				setOpenCodeFixtureProject(t, path, project)
			}
		}
	}
}

// prepareLegacyGenerateFixture gives older CLI tests their own configured project.
// Tests with explicit project/configuration retain that configuration unchanged.
func prepareLegacyGenerateFixture(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if cmd.Dir == "" {
		for i, arg := range cmd.Args {
			if (arg == "--json-output" || arg == "--append") && i+1 < len(cmd.Args) {
				cmd.Dir = filepath.Dir(cmd.Args[i+1])
				break
			}
		}
		if cmd.Dir == "" {
			cmd.Dir = t.TempDir()
		}
		if err := os.MkdirAll(cmd.Dir, 0700); err != nil {
			t.Fatal(err)
		}
		config := filepath.Join(cmd.Dir, ".agenticlens-config.yaml")
		if _, err := os.Stat(config); os.IsNotExist(err) {
			data := "project_name: CLIFixture\ntelemetry:\n  capture:\n    state_dir: .agenticlens-capture\nworkspaces:\n  dir: records\nticket:\n  prefixes:\n    - prefix: AL\n    - prefix: GEN\n      ticket_type: chore\n    - prefix: PO\n      ticket_type: docs\n    - prefix: FIX\n      ticket_type: fix\nrole_mappings:\n  build: orchestrator\ncode_metrics:\n  source_roots: [src]\nanomaly_thresholds:\n  context_ceiling_warning_tokens: 160000\n  context_ceiling_critical_tokens: 200000\n"
			if err := os.WriteFile(config, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	setOpenCodeFixtureProjectIfSelected(t, cmd, cmd.Dir)
	for i, arg := range cmd.Args {
		if arg == "--json-output" && i+1 < len(cmd.Args) {
			body, err := os.ReadFile(cmd.Args[i+1])
			if err != nil {
				continue
			}
			var ledger domain.AgenticFactLedger
			if json.Unmarshal(body, &ledger) == nil && len(ledger.Sessions) > 0 {
				scoped, err := capture.ProjectState(filepath.Join(cmd.Dir, ".agenticlens-capture"), cmd.Dir)
				if err != nil {
					t.Fatal(err)
				}
				registry, err := capture.OpenProject(scoped, cmd.Dir)
				if err != nil {
					t.Fatal(err)
				}
				rows, err := registry.List(context.Background())
				registry.Close()
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) == 0 {
					acceptFixtureOwners(t, cmd.Dir, scoped, *ledger.TicketID, ledger.Sessions)
				}
			}
		}
	}
}

// acceptFixtureOwners creates native ownership evidence for previously accepted
// snapshots, then removes the enrichment source. This retains the original
// unavailable-source preservation oracle without trusting a copied ledger.
func acceptFixtureOwners(t *testing.T, project, state, ticket string, sessions []domain.SessionFact) {
	t.Helper()
	scoped, err := capture.ProjectState(state, project)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := capture.OpenProject(scoped, project)
	if err != nil {
		t.Fatal(err)
	}
	defer registry.Close()
	dbPath := setupIsolatedEmptyDB(t)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, owner := range sessions {
		if owner.Source != domain.SourceOpenCode {
			continue
		}
		if _, err = db.Exec(`INSERT INTO session(id) VALUES(?)`, owner.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	setOpenCodeFixtureProject(t, dbPath, project)
	for _, owner := range sessions {
		if owner.Source != domain.SourceOpenCode {
			continue
		}
		id := domain.CaptureIdentity{Source: owner.Source, Store: dbPath, SessionID: owner.ID}
		r, err := registry.Register(context.Background(), id, "", "fixture accepted ownership")
		if err != nil {
			t.Fatal(err)
		}
		if err = capture.ValidateOwner(context.Background(), project, r); err != nil {
			t.Fatal(err)
		}
		if _, err = registry.Bind(context.Background(), id, ticket, r.Request); err != nil {
			t.Fatal(err)
		}
		if _, err = registry.Change(context.Background(), id, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
			r.ProjectDirectory = project
			return r, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.Remove(dbPath); err != nil {
		t.Fatal(err)
	}
}
