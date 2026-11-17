package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
)

func cliLocalSettings(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "approved-local.yaml")
	body := fmt.Sprintf("agenticlens:\n  state_dir: %q\n", filepath.Join(filepath.Dir(path), "publication-state")) + "telemetry:\n  default_source: opencode\n  sources:\n    opencode:\n      database_path: source.db\n  capture:\n    enabled: false\n    state_dir: local-state\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n"
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestInitApprovedSettingsPair(t *testing.T) {
	for _, suffix := range []string{"yaml", "yml"} {
		t.Run(suffix, func(t *testing.T) {
			project := t.TempDir()
			shared := filepath.Join(project, ".agenticlens-config."+suffix)
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			settings, localSettings := cliFixtureSettings(t), cliLocalSettings(t)
			args := []string{"init", "--settings", settings, "--local-settings", localSettings, "--workspace-root", "custom records", "--output", shared, "--project-name", "A shared fixture"}
			run := func(extra ...string) ([]byte, error) {
				cmd := exec.Command(cliBinaryPath, append(append([]string{}, args...), extra...)...)
				cmd.Dir = project
				return cmd.CombinedOutput()
			}
			if out, err := run("--dry-run"); err != nil || !bytes.Contains(out, []byte("enabled: false")) {
				t.Fatalf("validated dry run: %v %s", err, out)
			}
			for _, path := range []string{shared, local} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("dry run wrote %s: %v", path, err)
				}
			}
			if out, err := run(); err != nil {
				t.Fatalf("init: %v %s", err, out)
			}
			cfg, err := adaptersconfig.Load(project)
			if err != nil {
				t.Fatal(err)
			}
			if len(cfg.Workspaces.Roots) != 1 {
				t.Fatalf("expected one configured root: %+v", cfg.Workspaces.Roots)
			}
			workspace, err := adaptersconfig.ResolveConfiguredPath(project, cfg.Workspaces.Roots[0].Path)
			if err != nil {
				t.Fatal(err)
			}
			if discovery := cfg.Workspaces.Roots[0].Discovery; discovery.MaxDepth != 0 || discovery.MaxDirectories != 300 {
				t.Fatalf("writer defaults must be explicit root-only/budget300: %+v", discovery)
			}
			if cfg.ProjectName != "A shared fixture" || workspace != filepath.Join(project, "custom records") || cfg.AgenticLens.Version != "v9.8.7" || cfg.Telemetry.Capture.Enabled || cfg.Generation.Summary.Enabled || cfg.Database.BusyTimeoutMs != 0 || cfg.Ticket.CaseInsensitive != true || cfg.RoleOrder == nil {
				t.Fatalf("approved values changed: workspace=%q pin=%q capture=%t summary=%t busy=%d case=%t roles=%v", cfg.Workspaces.Roots[0].Path, cfg.AgenticLens.Version, cfg.Telemetry.Capture.Enabled, cfg.Generation.Summary.Enabled, cfg.Database.BusyTimeoutMs, cfg.Ticket.CaseInsensitive, cfg.RoleOrder)
			}
			publicationState, err := adaptersconfig.ResolvePublicationState(project)
			if err != nil {
				t.Fatal(err)
			}
			if publicationState != filepath.Join(filepath.Dir(localSettings), "publication-state") {
				t.Fatalf("approved operational state changed: %s", publicationState)
			}
			beforeShared, beforeLocal := readCaptureFile(t, shared), readCaptureFile(t, local)
			if out, err := run(); err == nil {
				t.Fatalf("collision accepted: %s", out)
			}
			if !bytes.Equal(beforeShared, readCaptureFile(t, shared)) || !bytes.Equal(beforeLocal, readCaptureFile(t, local)) {
				t.Fatal("collision changed paired outputs")
			}
			if out, err := run("--force"); err != nil {
				t.Fatalf("force: %v %s", err, out)
			}
			if !bytes.Equal(beforeLocal, readCaptureFile(t, local)) {
				t.Fatal("force changed valid local policy")
			}
			for _, path := range []string{publicationState, filepath.Join(project, "local-state"), filepath.Join(project, ".git", "hooks"), filepath.Join(project, ".opencode"), filepath.Join(project, ".codex")} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("configuration installed/activated %s: %v", path, err)
				}
			}
		})
	}
}

func TestInitRejectsInvalidCandidatesBeforeEitherWrite(t *testing.T) {
	for _, scenario := range []string{"missing-workspaces", "blank-workspaces", "missing-settings", "missing-local-choice", "invalid-shared", "invalid-local", "unpaired-output", "yaml-yml-ambiguity"} {
		t.Run(scenario, func(t *testing.T) {
			project := t.TempDir()
			shared := filepath.Join(project, ".agenticlens-config.yaml")
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			settings, localSettings := cliFixtureSettings(t), cliLocalSettings(t)
			args := []string{"init", "--settings", settings, "--local-settings", localSettings, "--workspace-root", "records", "--output", shared}
			switch scenario {
			case "missing-workspaces":
				// A settings file may provide roots without a CLI override; reject absent policy itself.
				body := readCaptureFile(t, settings)
				start := bytes.Index(body, []byte("  roots:\n"))
				end := bytes.Index(body, []byte("  artifacts:"))
				if start < 0 || end < start {
					t.Fatal("fixture workspace policy not found")
				}
				candidate := filepath.Join(t.TempDir(), "missing-roots.yaml")
				body = append(append(append([]byte{}, body[:start]...), []byte("  roots: []\n")...), body[end:]...)
				if err := os.WriteFile(candidate, body, 0600); err != nil {
					t.Fatal(err)
				}
				args[2] = candidate
				args = append(args[:5], args[7:]...)
			case "blank-workspaces":
				args[6] = " "
			case "missing-settings":
				args = append(args[:1], args[3:]...)
			case "missing-local-choice":
				args = append(args[:3], args[5:]...)
			case "invalid-shared":
				path := filepath.Join(t.TempDir(), "invalid.yaml")
				if err := os.WriteFile(path, []byte("project_name: Incomplete\n"), 0600); err != nil {
					t.Fatal(err)
				}
				args[2] = path
			case "invalid-local":
				path := filepath.Join(t.TempDir(), "invalid.yaml")
				if err := os.WriteFile(path, []byte("telemetry:\n  capture:\n    enabled: false\n"), 0600); err != nil {
					t.Fatal(err)
				}
				args[4] = path
			case "unpaired-output":
				args[8] = local
			case "yaml-yml-ambiguity":
				if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yml"), []byte("existing policy"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = project
			if out, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("invalid candidate accepted: %s", out)
			}
			for _, path := range []string{shared, local} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("validation wrote %s: %v", path, err)
				}
			}
			if scenario == "yaml-yml-ambiguity" && string(readCaptureFile(t, filepath.Join(project, ".agenticlens-config.yml"))) != "existing policy" {
				t.Fatal("ambiguity changed existing policy")
			}
		})
	}
}

func TestInitCompleteScaffoldAndPreserveLocal(t *testing.T) {
	for _, scenario := range []string{"complete", "existing-local", "edited", "different-pin", "dry-run"} {
		t.Run(scenario, func(t *testing.T) {
			project := t.TempDir()
			shared := filepath.Join(project, ".agenticlens-config.yaml")
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			scaffold, err := adaptersconfig.BootstrapScaffold("v9.8.7")
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "edited" {
				scaffold = append(scaffold, []byte("# user customization\n")...)
			}
			if err = os.WriteFile(shared, scaffold, 0600); err != nil {
				t.Fatal(err)
			}
			localSettings := cliLocalSettings(t)
			localBody := readCaptureFile(t, localSettings)
			if scenario == "existing-local" {
				if err = os.WriteFile(local, localBody, 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"init", "--complete-scaffold", "--settings", cliFixtureSettings(t), "--local-settings", localSettings, "--output", shared}
			if scenario == "different-pin" {
				args = append(args, "--agenticlens-version", "v9.8.8")
			}
			if scenario == "dry-run" {
				args = append(args, "--dry-run")
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = project
			out, runErr := cmd.CombinedOutput()
			if scenario == "edited" || scenario == "different-pin" {
				if runErr == nil || !bytes.Contains(out, []byte("existing files preserved")) {
					t.Fatalf("modified scaffold accepted: %v %s", runErr, out)
				}
				if !bytes.Equal(scaffold, readCaptureFile(t, shared)) {
					t.Fatal("rejected scaffold changed")
				}
			} else if runErr != nil {
				t.Fatalf("completion failed: %v %s", runErr, out)
			} else if scenario == "dry-run" {
				if !bytes.Equal(scaffold, readCaptureFile(t, shared)) {
					t.Fatal("dry run wrote scaffold")
				}
			} else {
				cfg, err := adaptersconfig.Load(project)
				if err != nil {
					t.Fatal(err)
				}
				if cfg.AgenticLens.Version != "v9.8.7" || strings.Contains(string(readCaptureFile(t, shared)), "bootstrap scaffold") {
					t.Fatal("completion did not preserve exact pin")
				}
			}
			if scenario == "existing-local" && !bytes.Equal(localBody, readCaptureFile(t, local)) {
				t.Fatal("completion changed valid local values/comments")
			}
		})
	}
}

func TestInitForeignOutputKeepsCLIPathsAtInvocationDirectory(t *testing.T) {
	invocation, output := t.TempDir(), t.TempDir()
	for _, name := range []string{"records", "sources", "state"} {
		if err := os.Mkdir(filepath.Join(invocation, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	source := filepath.Join(invocation, "sources", "native.db")
	if err := os.WriteFile(source, []byte("explicit setup path"), 0600); err != nil {
		t.Fatal(err)
	}
	localSettings := filepath.Join(t.TempDir(), "local.yaml")
	policy := "telemetry:\n  capture:\n    pass_timeout_ms: 3000\n    max_sessions: 256\n    max_source_bytes: 67108864\n"
	if err := os.WriteFile(localSettings, []byte(policy), 0600); err != nil {
		t.Fatal(err)
	}
	shared := filepath.Join(output, ".agenticlens-config.yml")
	cmd := exec.Command(cliBinaryPath, "init", "--settings", cliFixtureSettings(t), "--local-settings", localSettings, "--workspace-root", "./records", "--source", "opencode", "--db-path", "./sources/native.db", "--state-dir", "./state", "--publication-state-dir", "./publication", "--capture-enabled", "false", "--output", shared)
	cmd.Dir = invocation
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("foreign output: %v %s", err, out)
	}
	cfg, err := adaptersconfig.Load(output)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := adaptersconfig.ResolveConfiguredPath(output, cfg.Workspaces.Roots[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	state, err := adaptersconfig.ResolveGenerationCapture(output, "")
	if err != nil {
		t.Fatal(err)
	}
	publicationState, err := adaptersconfig.ResolvePublicationState(output)
	if err != nil {
		t.Fatal(err)
	}
	if publicationState != filepath.Join(invocation, "publication") {
		t.Fatalf("publication state changed CLI origin: %s", publicationState)
	}
	if _, err := os.Stat(publicationState); !os.IsNotExist(err) {
		t.Fatalf("init activated publication storage: %v", err)
	}
	if workspace != filepath.Join(invocation, "records") || cfg.Telemetry.Sources.OpenCodeDatabasePath != source || state.StateDir != filepath.Join(invocation, "state") {
		t.Fatalf("CLI paths changed origin after writing foreign output: workspace=%s source=%s state=%s", workspace, cfg.Telemetry.Sources.OpenCodeDatabasePath, state.StateDir)
	}
}

func TestInitLocalTargetFailurePreservesShared(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			project := t.TempDir()
			shared := filepath.Join(project, ".agenticlens-config.yaml")
			local := filepath.Join(project, ".agenticlens-config.local.yaml")
			original := readCaptureFile(t, cliFixtureSettings(t))
			if existing {
				if err := os.WriteFile(shared, original, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(local, 0700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(cliBinaryPath, "init", "--settings", cliFixtureSettings(t), "--local-settings", cliLocalSettings(t), "--workspace-root", "records", "--force")
			cmd.Dir = project
			out, err := cmd.CombinedOutput()
			if err == nil || !bytes.Contains(out, []byte(".agenticlens-config.local.yaml")) || bytes.Contains(out, []byte("Initialized approved")) {
				t.Fatalf("local target failure: %v %s", err, out)
			}
			if existing {
				if !bytes.Equal(original, readCaptureFile(t, shared)) {
					t.Fatal("local failure modified shared configuration")
				}
			} else if _, err := os.Stat(shared); !os.IsNotExist(err) {
				t.Fatalf("local failure created partial shared file: %v", err)
			}
			entries, err := os.ReadDir(project)
			want := 1
			if existing {
				want++
			}
			if err != nil || len(entries) != want {
				t.Fatalf("staged/backup files leaked: %v %v", entries, err)
			}
		})
	}
}

func TestInitExactVersionRequiredBeforeWrites(t *testing.T) {
	for _, pin := range []string{"", "latest", "v1.2.3/../../escape"} {
		t.Run(pin, func(t *testing.T) {
			project := t.TempDir()
			candidate := filepath.Join(t.TempDir(), "shared.yaml")
			body := bytes.Replace(readCaptureFile(t, cliFixtureSettings(t)), []byte("agenticlens:\n  version: v9.8.7\n"), nil, 1)
			if err := os.WriteFile(candidate, body, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"init", "--settings", candidate, "--local-settings", cliLocalSettings(t), "--workspace-root", "records"}
			if pin != "" {
				args = append(args, "--agenticlens-version", pin)
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = project
			if out, err := cmd.CombinedOutput(); err == nil || !bytes.Contains(out, []byte("version")) {
				t.Fatalf("invalid pin accepted: %v %s", err, out)
			}
			entries, err := os.ReadDir(project)
			if err != nil || len(entries) != 0 {
				t.Fatalf("invalid pin wrote files: %v %v", entries, err)
			}
		})
	}
}

func TestCLICommandsMissingConfigurationAreReadOnly(t *testing.T) {
	for _, command := range []string{"root", "server", "generate"} {
		t.Run(command, func(t *testing.T) {
			project := t.TempDir()
			args := []string{"--list"}
			if command == "server" {
				args = []string{"server"}
			}
			if command == "generate" {
				args = []string{"generate", "--source", "codex", "--codex-home", filepath.Join(project, "missing-source"), "--session-id", "root", "--ticket-id", "TEST-1", "--ticket-title", "Missing configuration", "--json-output", filepath.Join(project, "agentic_metrics.json")}
			}
			cmd := exec.Command(cliBinaryPath, args...)
			cmd.Dir = project
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(strings.ToLower(string(out)), "config") {
				t.Fatalf("missing configuration did not fail clearly: %v %s", err, out)
			}
			entries, err := os.ReadDir(project)
			if err != nil || len(entries) != 0 {
				t.Fatalf("missing configuration wrote files: %v %v", entries, err)
			}
		})
	}
}

func TestInitExplicitWorkspaceRootsAndBounds(t *testing.T) {
	for _, scenario := range []struct {
		name   string
		depth  string
		budget string
		valid  bool
	}{
		{name: "two-container-roots", depth: "1", budget: "300", valid: true},
		{name: "negative-depth", depth: "-1", budget: "300"},
		{name: "zero-budget", depth: "0", budget: "0"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			project := t.TempDir()
			cmd := exec.Command(cliBinaryPath, "init", "--settings", cliFixtureSettings(t), "--local-settings", cliLocalSettings(t),
				"--workspace-root", "active", "--workspace-root", "archive", "--discovery-depth", scenario.depth, "--directory-budget", scenario.budget)
			cmd.Dir = project
			out, err := cmd.CombinedOutput()
			if !scenario.valid {
				if err == nil {
					t.Fatalf("invalid discovery bounds accepted: %s", out)
				}
				entries, readErr := os.ReadDir(project)
				if readErr != nil || len(entries) != 0 {
					t.Fatalf("invalid bounds wrote configuration: %v %v", entries, readErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("explicit roots: %v %s", err, out)
			}
			cfg, err := adaptersconfig.Load(project)
			if err != nil {
				t.Fatal(err)
			}
			if len(cfg.Workspaces.Roots) != 2 {
				t.Fatalf("repeated roots lost a configured location: %+v", cfg.Workspaces.Roots)
			}
			for i, name := range []string{"active", "archive"} {
				root := cfg.Workspaces.Roots[i]
				path, err := adaptersconfig.ResolveConfiguredPath(project, root.Path)
				if err != nil || path != filepath.Join(project, name) || root.Discovery.MaxDepth != 1 || root.Discovery.MaxDirectories != 300 {
					t.Fatalf("root settings changed: %+v path=%s err=%v", root, path, err)
				}
			}
		})
	}
}
