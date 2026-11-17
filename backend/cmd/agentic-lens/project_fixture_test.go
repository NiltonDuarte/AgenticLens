package main_test

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
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
				db, err := sql.Open("sqlite", path)
				if err != nil {
					t.Fatal(err)
				}
				var exists int
				queryErr := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='session'`).Scan(&exists)
				db.Close()
				// Invalid-source tests must reach the command with their bytes untouched.
				if queryErr == nil && exists == 1 {
					setOpenCodeFixtureProject(t, path, project)
				}
			}
		}
	}
}

// prepareGenerateFixture owns explicit policy and native source ownership for
// isolated command tests. Existing custom configurations remain untouched.
func prepareGenerateFixture(t *testing.T, cmd *exec.Cmd) {
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
	}
	path := filepath.Join(cmd.Dir, ".agenticlens-config.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeCLISharedPolicy(t, cmd.Dir)
		cfg, err := adaptersconfig.Load(cmd.Dir)
		if err != nil {
			t.Fatal(err)
		}
		directory := cmd.Dir
		for i, arg := range cmd.Args {
			if arg == "--json-output" && i+1 < len(cmd.Args) {
				directory = filepath.Dir(cmd.Args[i+1])
				if !filepath.IsAbs(directory) {
					directory = filepath.Join(cmd.Dir, directory)
				}
				break
			}
		}
		if err := os.MkdirAll(directory, 0700); err != nil {
			t.Fatal(err)
		}
		cfg.Workspaces.Roots = []domain.WorkspaceRootConfig{{Path: directory, Discovery: domain.WorkspaceDiscoveryConfig{MaxDepth: 0, MaxDirectories: 300}}}
		body, err := adaptersconfig.GenerateConfigYAML(cfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, body, 0600); err != nil {
			t.Fatal(err)
		}
	}
	local := filepath.Join(cmd.Dir, ".agenticlens-config.local.yaml")
	if _, err := os.Stat(local); os.IsNotExist(err) {
		if err := writeCaptureFixtureLocal(t, captureGenerationFixture{project: cmd.Dir, state: filepath.Join(cmd.Dir, ".agenticlens-capture")}, []byte("telemetry:\n  default_source: opencode\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n    refresh_ms: 100\n    debounce_ms: 0\n    idle_ms: 300\n    max_lifetime_ms: 5000\n    commit_warning: true\n    enabled: false\n")); err != nil {
			t.Fatal(err)
		}
	}
	localBody, localErr := os.ReadFile(local)
	if localErr != nil {
		t.Fatal(localErr)
	}
	if !strings.Contains(string(localBody), "agenticlens:\n") {
		localBody = append(localBody, []byte(fmt.Sprintf("\nagenticlens:\n  state_dir: %q\n", filepath.Join(cmd.Dir, ".publication-state")))...)
		if err := os.WriteFile(local, localBody, 0600); err != nil {
			t.Fatal(err)
		}
	}
	initCLIGitFixture(t, cmd.Dir)
	for i, arg := range cmd.Args {
		if arg == "--codex-home" && i+1 < len(cmd.Args) && filepath.Base(cmd.Args[i+1]) == "empty-codex-source" {
			body := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":\"fixture-empty-root\",\"cwd\":%q}}\n", cmd.Dir)
			if err := os.WriteFile(filepath.Join(cmd.Args[i+1], "sessions", "root.jsonl"), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	setOpenCodeFixtureProjectIfSelected(t, cmd, cmd.Dir)
}
