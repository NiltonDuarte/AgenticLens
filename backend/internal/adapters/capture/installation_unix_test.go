//go:build !windows

package capture_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

func TestMain(m *testing.M) {
	if os.Getenv("AL_CAPTURE_INSTALL_FIXTURE") == "1" {
		if len(os.Args) == 3 && os.Args[1] == "engine" && os.Args[2] == "protocol" {
			fmt.Println("1")
			os.Exit(0)
		}
		if len(os.Args) == 3 && os.Args[1] == "capture" && os.Args[2] == "protocol" {
			fmt.Println(capture.HandlerProtocolVersion)
			os.Exit(0)
		}
		if len(os.Args) == 3 && os.Args[1] == "version" && os.Args[2] == "--short" {
			fmt.Println("v9.8.7")
			os.Exit(0)
		}
		if len(os.Args) == 3 && os.Args[1] == "version" && os.Args[2] == "--json" {
			fmt.Println(`{"version":"v9.8.7","commit":"adapter-fixture","build_date":"fixture"}`)
			os.Exit(0)
		}
	}
	code := m.Run()
	if installationManagerRoot != "" {
		os.RemoveAll(installationManagerRoot)
	}
	os.Exit(code)
}

func TestInstalledHealthOwnedConfigurationReadOnly(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		t.Run(source, func(t *testing.T) {
			project, state := t.TempDir(), t.TempDir()
			binary, wrapper := installFixture(t, project)
			storePath := t.TempDir()
			if source == "opencode" {
				storePath = filepath.Join(storePath, "source.db")
			}
			installed, err := capture.Install(context.Background(), capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: source, SourceStore: storePath})
			if err != nil {
				t.Fatal(err)
			}
			manifest := installed.Manifest
			anchor := filepath.Join(project, "."+source, "agenticlens-installation.json")
			original := checkRead(t, manifest.ConfigPath)
			anchorBody := checkRead(t, anchor)
			// Preserve real stored history through both healthy and stale checks.
			store, err := capture.OpenProject(state, project)
			if err != nil {
				t.Fatal(err)
			}
			id := domain.CaptureIdentity{Source: domain.SourceHarness(source), Store: storePath, SessionID: "observed-root"}
			if _, err := store.Register(context.Background(), id, "", "SessionStart"); err != nil {
				t.Fatal(err)
			}
			if err := store.RecordDiagnostic(context.Background(), domain.CaptureDiagnostic{TicketID: "TEST-100", Phase: "publication", Outcome: "committed"}); err != nil {
				t.Fatal(err)
			}
			if err := store.Close(); err != nil {
				t.Fatal(err)
			}
			type healthCase struct {
				name   string
				change func(t *testing.T)
				stale  bool
			}
			cases := []healthCase{
				{"unchanged", func(t *testing.T) {}, false},
				{"missing anchor retains exact local ownership", func(t *testing.T) {
					if err := os.Remove(anchor); err != nil {
						t.Fatal(err)
					}
				}, false},
				{"missing owned integration", func(t *testing.T) {
					if err := os.Remove(manifest.ConfigPath); err != nil {
						t.Fatal(err)
					}
				}, true},
				{"missing integration and anchor", func(t *testing.T) {
					for _, path := range []string{anchor, manifest.ConfigPath} {
						if err := os.Remove(path); err != nil {
							t.Fatal(err)
						}
					}
				}, true},
				{"manifest points to another config", func(t *testing.T) {
					changed := manifest
					changed.ConfigPath = filepath.Join(t.TempDir(), "foreign integration")
					body, err := json.Marshal(changed)
					if err != nil {
						t.Fatal(err)
					}
					checkWrite(t, anchor, string(body), 0600)
				}, true},
				{"identical integration outside project", func(t *testing.T) {
					outside := filepath.Join(t.TempDir(), "foreign config")
					checkWrite(t, outside, string(original), 0600)
					if err := os.Remove(manifest.ConfigPath); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(outside, manifest.ConfigPath); err != nil {
						t.Fatal(err)
					}
				}, true},
				{"oversized integration unavailable", func(t *testing.T) {
					checkWrite(t, manifest.ConfigPath, strings.Repeat(" ", 1024*1024+1), 0600)
				}, true},
				{"nonregular owned integration", func(t *testing.T) {
					if err := os.Remove(manifest.ConfigPath); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(manifest.ConfigPath, 0700); err != nil {
						t.Fatal(err)
					}
				}, true},
			}
			if source == "codex" {
				codexChange := func(t *testing.T, kind string) {
					t.Helper()
					var config map[string]json.RawMessage
					var hooks map[string][]json.RawMessage
					if err := json.Unmarshal(original, &config); err != nil {
						t.Fatal(err)
					}
					if err := json.Unmarshal(config["hooks"], &hooks); err != nil {
						t.Fatal(err)
					}
					event := manifest.Groups[0].Event
					switch kind {
					case "unrelated":
						config["userSetting"] = json.RawMessage(`true`)
						hooks[event] = append(hooks[event], json.RawMessage(`{"hooks":[{"type":"command","command":"user-owned sentinel"}]}`))
					case "reorder":
						for event, groups := range hooks {
							for i, group := range groups {
								var reordered map[string][]map[string]json.RawMessage
								if err := json.Unmarshal(group, &reordered); err != nil {
									t.Fatal(err)
								}
								body, err := json.Marshal(reordered)
								if err != nil {
									t.Fatal(err)
								}
								hooks[event][i] = body
							}
						}
					case "missing":
						delete(hooks, event)
					case "edited":
						var group map[string][]map[string]json.RawMessage
						if err := json.Unmarshal(hooks[event][0], &group); err != nil {
							t.Fatal(err)
						}
						group["hooks"][0]["timeout"] = json.RawMessage(`9`)
						body, err := json.Marshal(group)
						if err != nil {
							t.Fatal(err)
						}
						hooks[event][0] = body
					case "duplicate":
						hooks[event] = append(hooks[event], hooks[event][0])
					}
					hookBody, err := json.Marshal(hooks)
					if err != nil {
						t.Fatal(err)
					}
					config["hooks"] = hookBody
					body, err := json.MarshalIndent(config, "", "    ")
					if err != nil {
						t.Fatal(err)
					}
					checkWrite(t, manifest.ConfigPath, string(body), 0600)
				}
				for _, kind := range []string{"unrelated", "reorder", "missing", "edited", "duplicate"} {
					cases = append(cases, healthCase{kind + " codex group", func(t *testing.T) { codexChange(t, kind) }, kind != "unrelated" && kind != "reorder"})
				}
				cases = append(cases, healthCase{"invalid codex config", func(t *testing.T) {
					checkWrite(t, manifest.ConfigPath, "{invalid", 0600)
				}, true})
			} else {
				cases = append(cases, healthCase{"edited plugin", func(t *testing.T) {
					checkWrite(t, manifest.ConfigPath, string(original)+"\n// user edit\n", 0600)
				}, true}, healthCase{"unrelated plugin", func(t *testing.T) {
					checkWrite(t, filepath.Join(filepath.Dir(manifest.ConfigPath), "user-plugin.ts"), "export default {}\n", 0600)
				}, false})
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					if err := os.RemoveAll(manifest.ConfigPath); err != nil {
						t.Fatal(err)
					}
					checkWrite(t, manifest.ConfigPath, string(original), 0600)
					checkWrite(t, anchor, string(anchorBody), 0600)
					tc.change(t)
					before := installationFixtureTree(t, project, state)
					err := capture.CheckInstalledState(project, state, false)
					if (err != nil) != tc.stale {
						t.Fatalf("installation health stale=%t: %v", tc.stale, err)
					}
					issues, err := capture.InstallationIssues(project, state)
					if err != nil || (len(issues) > 0) != tc.stale {
						t.Fatalf("structured installation issues: %v %v", issues, err)
					}
					if after := installationFixtureTree(t, project, state); !reflect.DeepEqual(before, after) {
						t.Fatal("health inspection modified configuration, manifest or stored exchange history")
					}
				})
			}
		})
	}
}

func installationFixtureTree(t *testing.T, roots ...string) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			body := ""
			if !entry.IsDir() {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				body = string(data)
			}
			files[path] = info.Mode().String() + body
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return files
}

func TestInstallationIssuesSeparatesNamespaceConflictsFromDrift(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary, wrapper := installFixture(t, project)
	installed, err := capture.Install(context.Background(), capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "codex", SourceStore: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	anchor := filepath.Join(project, ".codex", "agenticlens-installation.json")
	local := filepath.Join(state, "install-codex.json")
	anchorBody, localBody := checkRead(t, anchor), checkRead(t, local)
	configBody := checkRead(t, installed.Manifest.ConfigPath)
	for _, tc := range []struct {
		name  string
		local bool
		fatal bool
		edit  func(*capture.Manifest)
	}{
		{"foreign discovery project", false, true, func(m *capture.Manifest) { m.Repository = t.TempDir() }},
		{"foreign selected local project", true, true, func(m *capture.Manifest) { m.Repository = t.TempDir() }},
		{"discovery source conflict", false, true, func(m *capture.Manifest) { m.Source = "opencode" }},
		{"selected local source conflict", true, true, func(m *capture.Manifest) { m.Source = "opencode" }},
		{"stale template accessible", false, false, func(m *capture.Manifest) { m.TemplateVersion = "previous" }},
		{"stale protocol accessible", false, false, func(m *capture.Manifest) { m.Protocol = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkWrite(t, anchor, string(anchorBody), 0600)
			checkWrite(t, local, string(localBody), 0600)
			checkWrite(t, installed.Manifest.ConfigPath, string(configBody), 0600)
			manifest := installed.Manifest
			tc.edit(&manifest)
			path := anchor
			if tc.local {
				path = local
				// No marker remains to select a different namespace. The requested
				// local state still cannot import a foreign project's manifest.
				for _, path := range []string{anchor, installed.Manifest.ConfigPath} {
					if err := os.Remove(path); err != nil {
						t.Fatal(err)
					}
				}
			}
			body, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			checkWrite(t, path, string(body), 0600)
			before := installationFixtureTree(t, project, state)
			issues, err := capture.InstallationIssues(project, state)
			if (err != nil) != tc.fatal || (!tc.fatal && len(issues) != 1) || (tc.fatal && len(issues) != 0) {
				t.Fatalf("namespace/drift distinction: issues=%v err=%v", issues, err)
			}
			if after := installationFixtureTree(t, project, state); !reflect.DeepEqual(before, after) {
				t.Fatal("namespace inspection modified owned state")
			}
		})
	}
}

// Real shell selection and a native ELF test helper guard installer provenance.
func installFixture(t *testing.T, project string) (string, string) {
	t.Helper()
	t.Setenv("AL_CAPTURE_INSTALL_FIXTURE", "1")
	t.Setenv("GORACE", "atexit_sleep_ms=0")
	harnessDir := t.TempDir()
	for source, body := range map[string]string{
		"opencode": "#!/bin/sh\n[ \"$*\" = \"--version\" ] || exit 2\nprintf '1.19.5\\n'\n",
		"codex":    "#!/bin/sh\ncase \"$*\" in\n  --version) printf 'codex-cli 0.166.0\\n';;\n  'features list') printf 'hooks enabled\\n';;\n  *) exit 2;;\nesac\n",
	} {
		if err := os.WriteFile(filepath.Join(harnessDir, source), []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", harnessDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	cache := t.TempDir()
	captureProjectConfig(t, project, "InstallationFixture")
	originalConfig, err := os.ReadFile(filepath.Join(project, ".agenticlens-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	config := strings.Replace(string(originalConfig), "  version: v9.8.7\n", "  version: v9.8.7\n  install_root: '"+cache+"'\n", 1)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := adaptersconfig.ResolveEntrypoint(project)
	if err != nil {
		t.Fatal(err)
	}
	binary, err := entry.NativePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(binary), 0700); err != nil {
		t.Fatal(err)
	}
	testBinary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(testBinary, binary); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(installationManagerExecutable(t))
	if err != nil {
		t.Fatal(err)
	}
	wrapper := filepath.Join(t.TempDir(), "common wrapper")
	if err := os.WriteFile(wrapper, body, 0700); err != nil {
		t.Fatal(err)
	}
	return binary, wrapper
}

func TestInstalledStateDriftAliasRemovalAndDryRun(t *testing.T) {
	project, state := t.TempDir(), filepath.Join(t.TempDir(), "state with ' quote")
	binary, wrapper := installFixture(t, project)
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "codex", SourceStore: t.TempDir()}
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(project, ".codex", "agenticlens-installation.json")); err != nil {
		t.Fatal(err)
	}
	states, err := capture.InstalledStates(project)
	if err != nil || len(states) != 1 || states[0] != state {
		t.Fatalf("discovery %v %v", states, err)
	}
	if err := capture.CheckInstalledState(project, t.TempDir(), false); err == nil {
		t.Fatal("state drift hidden")
	}
	if err := capture.CheckInstalledState(project, t.TempDir(), true); err != nil {
		t.Fatal("intentional override rejected", err)
	}
	if issues, err := capture.InstallationIssues(project, t.TempDir()); err != nil || len(issues) != 1 || !strings.Contains(issues[0], "state") {
		t.Fatalf("state disagreement missing from structured inspection: %v %v", issues, err)
	}
	alias := filepath.Join(t.TempDir(), "project")
	if err := os.Symlink(project, alias); err != nil {
		t.Fatal(err)
	}
	options.Repository = alias
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal("alias reinstall", err)
	}
	options.Remove, options.DryRun = true, true
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if enabled, err := capture.AutomaticEnabled(state); err != nil || !enabled {
		t.Fatal("dry-run deactivated", err)
	}
	options.DryRun = false
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal("alias removal", err)
	}
	if enabled, err := capture.AutomaticEnabled(state); err != nil || enabled {
		t.Fatal("removal remained active", err)
	}
	options.Repository = t.TempDir()
	if _, err := capture.Install(context.Background(), options); err == nil {
		t.Fatal("foreign project removal accepted")
	}
}

func TestLegacyOpenCodeInstallationDiscovery(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary, wrapper := installFixture(t, project)
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "opencode", SourceStore: filepath.Join(t.TempDir(), "source.db")}
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(project, ".opencode", "agenticlens-installation.json")); err != nil {
		t.Fatal(err)
	}
	states, err := capture.InstalledStates(project)
	if err != nil || len(states) != 1 || states[0] != state {
		t.Fatalf("legacy plugin discovery %v %v", states, err)
	}
}

func TestInstallationManifestStableAndPinSwitchRequiresRefresh(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary, wrapper := installFixture(t, project)
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "opencode", SourceStore: filepath.Join(t.TempDir(), "source.db")}
	first, err := capture.Install(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(state, "install-opencode.json")
	before, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if first.Manifest.Wrapper != wrapper || first.Manifest.Version != "v9.8.7" || first.Manifest.SourceStore == "" || first.Manifest.TemplateVersion == "" || first.Manifest.WrapperHash == "" {
		t.Fatalf("missing provenance %+v", first.Manifest)
	}
	if _, err := capture.Install(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("idempotent reinstall changed manifest identity")
	}
	if err := capture.CheckInstalledState(project, state, false); err != nil {
		t.Fatal("physical native alias failed health", err)
	}
	path := filepath.Join(project, ".agenticlens-config.yaml")
	shared, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(string(shared), "v9.8.7", "v9.8.8", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := capture.CheckInstalledState(project, state, false); err == nil {
		t.Fatal("pin switch retained false installation health")
	}
	if err := os.WriteFile(path, shared, 0600); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(wrapper)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wrapper, append(body, []byte("\n# local edit\n")...), 0700); err != nil {
		t.Fatal(err)
	}
	if err := capture.CheckInstalledState(project, state, false); err == nil {
		t.Fatal("wrapper drift hidden")
	}
}

func TestInstallationManifestRecordsChangedOptionalProbeOnly(t *testing.T) {
	project, state := t.TempDir(), t.TempDir()
	binary, wrapper := installFixture(t, project)
	options := capture.InstallOptions{Repository: project, StateDir: state, Binary: binary, Wrapper: wrapper, Source: "opencode", SourceStore: filepath.Join(t.TempDir(), "source.db")}
	first, err := capture.Install(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	if first.Manifest.HarnessVersion != "1.19.5" {
		t.Fatalf("controlled successful probe unavailable: %+v", first.Manifest)
	}
	harness, err := exec.LookPath("opencode")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(harness, []byte("#!/bin/sh\nexit 42\n"), 0700); err != nil {
		t.Fatal(err)
	}
	second, err := capture.Install(context.Background(), options)
	if err != nil {
		t.Fatalf("optional probe failure rejected the valid native installation: %v", err)
	}
	if second.Manifest.HarnessVersion != "unavailable" || !strings.Contains(second.Manifest.Capability, "Harness unavailable") {
		t.Fatalf("controlled probe failure not recorded: %+v", second.Manifest)
	}
	beforeIdentity, afterIdentity := first.Manifest, second.Manifest
	beforeIdentity.HarnessVersion, afterIdentity.HarnessVersion = "", ""
	beforeIdentity.Capability, afterIdentity.Capability = "", ""
	if !reflect.DeepEqual(beforeIdentity, afterIdentity) {
		t.Fatalf("optional probe changed required identity: before=%+v after=%+v", beforeIdentity, afterIdentity)
	}
	t.Logf("controlled manifest delta: HarnessVersion %q -> %q; Capability %q -> %q", first.Manifest.HarnessVersion, second.Manifest.HarnessVersion, first.Manifest.Capability, second.Manifest.Capability)
}

func TestCompatibleGlobalWrapperUpdatePreservesOtherPinnedProject(t *testing.T) {
	projectA, stateA := t.TempDir(), t.TempDir()
	nativeA, wrapper := installFixture(t, projectA)
	opts := capture.InstallOptions{Repository: projectA, StateDir: stateA, Binary: nativeA, Wrapper: wrapper, Source: "opencode", SourceStore: filepath.Join(t.TempDir(), "source-a.db")}
	if _, err := capture.Install(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(stateA, "install-opencode.json")
	manifestA, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	pluginPath := filepath.Join(projectA, ".opencode", "plugins", "agenticlens-capture.ts")
	pluginA, err := os.ReadFile(pluginPath)
	if err != nil {
		t.Fatal(err)
	}
	nativeBefore, err := os.ReadFile(nativeA)
	if err != nil {
		t.Fatal(err)
	}
	entryA, err := adaptersconfig.ResolveEntrypoint(projectA)
	if err != nil {
		t.Fatal(err)
	}
	projectB, stateB := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(projectB, ".agenticlens-config.yaml"), []byte("agenticlens:\n  version: v9.8.8\n  install_root: '"+entryA.InstallRoot+"'\n"), 0600); err != nil {
		t.Fatal(err)
	}
	entryB, err := adaptersconfig.ResolveEntrypoint(projectB)
	if err != nil {
		t.Fatal(err)
	}
	nativeB, err := entryB.NativePath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(nativeB), 0700); err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-ldflags", "-X main.version=v9.8.8", "-o", nativeB, filepath.Join("..", "..", "..", "..", "scripts", "tests", "fixtures", "entrypoint_probe.go"))
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build real second-version native: %s %v", out, err)
	}
	body, err := os.ReadFile(wrapper)
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, []byte("\n# Compatible installer-owned common entrypoint update\n")...)
	if err := os.WriteFile(wrapper, body, 0700); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(body)
	if err := os.WriteFile(wrapper+".sha256", []byte(hex.EncodeToString(sum[:])+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	opts.Repository, opts.StateDir, opts.Binary = projectB, stateB, nativeB
	opts.SourceStore = filepath.Join(t.TempDir(), "source-b.db")
	if _, err := capture.Install(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if err := capture.CheckInstalledState(projectA, stateA, false); err != nil {
		t.Fatal("B's compatible shared wrapper update invalidated A", err)
	}
	for path, before := range map[string][]byte{manifestPath: manifestA, pluginPath: pluginA, nativeA: nativeBefore} {
		after, err := os.ReadFile(path)
		if err != nil || string(after) != string(before) {
			t.Fatalf("B installation changed A artifact %s: %v", path, err)
		}
	}
	for project, want := range map[string]string{projectA: "v9.8.7", projectB: "v9.8.8"} {
		out, err := exec.Command(wrapper, "--al-offline", "--al-project", project, "version", "--short").CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) != want {
			t.Fatalf("project dispatch %s: %s %v", project, out, err)
		}
	}
}

var installationManagerOnce sync.Once
var installationManagerRoot, installationManagerPath string
var installationManagerErr error

func installationManagerExecutable(t *testing.T) string {
	t.Helper()
	installationManagerOnce.Do(func() {
		installationManagerRoot, installationManagerErr = os.MkdirTemp("", "agenticlens-adapter-manager-")
		if installationManagerErr != nil {
			return
		}
		installationManagerPath = filepath.Join(installationManagerRoot, "agentic-lens")
		command := exec.Command("go", "build", "-ldflags", "-X agenticlens/internal/domain.Version=v9.8.7", "-o", installationManagerPath, "./cmd/agentic-lens-manager")
		command.Dir = filepath.Join("..", "..", "..")
		if output, err := command.CombinedOutput(); err != nil {
			installationManagerErr = fmt.Errorf("build manager: %w: %s", err, output)
		}
	})
	if installationManagerErr != nil {
		t.Fatal(installationManagerErr)
	}
	return installationManagerPath
}
