//go:build !windows

package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func managerBuild(t *testing.T, directory, name, source, version, arch string) string {
	t.Helper()
	module, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, name)
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-ldflags", "-X agenticlens/internal/domain.Version="+version, "-o", path, source)
	command.Dir = module
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS="+runtime.GOOS, "GOARCH="+arch)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build %s: %v\n%s", source, err, output)
	}
	return path
}

func managerInvoke(t *testing.T, binary, cwd string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = cwd
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("manager did not terminate: %v\n%s", ctx.Err(), output)
	}
	return string(output), err
}

func TestNativeArtifactAdmissionAndStandaloneCompanion(t *testing.T) {
	root := t.TempDir()
	const version = "v9.4.2"
	engine := managerBuild(t, root, "engine", "./cmd/agentic-lens", version, runtime.GOARCH)
	manager := managerBuild(t, root, "manager", "./cmd/agentic-lens-manager", version, runtime.GOARCH)
	const nextVersion = "v9.4.3"
	updatedManager := managerBuild(t, root, "updated-manager", "./cmd/agentic-lens-manager", nextVersion, runtime.GOARCH)
	foreignArch := "arm64"
	if runtime.GOARCH == foreignArch {
		foreignArch = "amd64"
	}
	foreign := managerBuild(t, root, "foreign-engine", "./cmd/agentic-lens", version, foreignArch)
	foreignManager := managerBuild(t, root, "foreign-manager", "./cmd/agentic-lens-manager", version, foreignArch)
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	outside := filepath.Join(root, "unconfigured")
	if err := os.Mkdir(outside, 0700); err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(home, ".agenticlens")
	output, err := managerInvoke(t, manager, outside, "version", "--short")
	if err == nil || !strings.Contains(output, version) {
		t.Fatalf("missing companion did not diagnose its exact release: %q %v", output, err)
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("ordinary missing-engine command initialized cache: %v", err)
	}
	output, err = managerInvoke(t, manager, outside, "manager", "--help")
	if err != nil || !strings.Contains(output, "self-update") {
		t.Fatalf("maintenance help needs companion engine: %q %v", output, err)
	}
	output, err = managerInvoke(t, manager, outside, "--manager-version")
	if err != nil || strings.TrimSpace(output) != version {
		t.Fatalf("manager identity depends on engine: %q %v", output, err)
	}
	engineBytes := managerRead(t, engine)
	managerBytes := managerRead(t, manager)
	install := []string{"install", version, "--local-bin", engine, "--sha256", digest(engineBytes)}
	output, err = managerInvoke(t, manager, outside, install...)
	if err != nil {
		t.Fatalf("install companion: %v\n%s", err, output)
	}
	leaf := filepath.Join(cache, "bin", version, runtime.GOOS+"-"+runtime.GOARCH, artifactName("agentic-lens-engine"))
	admitted, err := os.Stat(leaf)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(leaf, 0600); err != nil {
		t.Fatal(err)
	}
	output, err = managerInvoke(t, manager, outside, install...)
	if err != nil {
		t.Fatalf("idempotent permission repair: %v\n%s", err, output)
	}
	repaired, err := os.Stat(leaf)
	if err != nil || !os.SameFile(admitted, repaired) || repaired.Mode().Perm()&0100 == 0 {
		t.Fatalf("idempotent repair retired identity or permissions: %v", err)
	}
	output, err = managerInvoke(t, manager, outside, "version", "--short")
	if err != nil || strings.TrimSpace(output) != version {
		t.Fatalf("installed companion not selected outside project: %q %v", output, err)
	}
	for _, alias := range []string{"-v", "--version", "-version"} {
		t.Run("standalone "+alias, func(t *testing.T) {
			expected, err := managerInvoke(t, engine, outside, alias)
			if err != nil || !strings.Contains(expected, version) {
				t.Fatalf("raw engine alias control: %q %v", expected, err)
			}
			actual, err := managerInvoke(t, manager, outside, alias)
			if err != nil || actual != expected {
				t.Fatalf("standalone engine version alias changed: got=%q want=%q error=%v", actual, expected, err)
			}
		})
	}
	roleProject := filepath.Join(root, "configured-project")
	if err := os.Mkdir(roleProject, 0700); err != nil {
		t.Fatal(err)
	}
	managerWrite(t, filepath.Join(roleProject, ".agenticlens-config.yaml"), []byte(fmt.Sprintf("agenticlens:\n  version: %s\n  install_root: %q\n", version, cache)), 0600)
	output, err = managerInvoke(t, manager, roleProject, "capture", "protocol")
	if err != nil || strings.TrimSpace(output) != "3" {
		t.Fatalf("wrong-role control cannot forward a real engine interface: %q %v", output, err)
	}
	t.Run("reject real candidates before destination writes", func(t *testing.T) {
		for _, test := range []struct {
			name, source, wanted, checksum, fragment string
		}{
			{"engine checksum", engine, version, strings.Repeat("0", 64), "SHA256"},
			{"engine identity", engine, "v9.4.3", digest(engineBytes), "identity"},
			{"manager as engine", manager, version, digest(managerBytes), "native engine interface"},
			{"foreign host", foreign, version, digest(managerRead(t, foreign)), "engine rejected"},
			{"unsafe version", engine, "../latest", digest(engineBytes), "version"},
		} {
			t.Run(test.name, func(t *testing.T) {
				destination := filepath.Join(root, strings.ReplaceAll(test.name, " ", "-"), "cache")
				caller := outside
				if test.name == "manager as engine" {
					caller = roleProject
				}
				output, err := managerInvoke(t, manager, caller, "install", test.wanted, "--local-bin", test.source, "--sha256", test.checksum, "--install-root", destination)
				if err == nil || !strings.Contains(output, test.fragment) {
					t.Fatalf("invalid native admitted or wrong rejection: %q %v", output, err)
				}
				if _, err := os.Stat(destination); !os.IsNotExist(err) {
					t.Fatalf("refused artifact initialized cache: %v", err)
				}
			})
		}
	})
	t.Run("foreign local manager is not a cached engine", func(t *testing.T) {
		destination := filepath.Join(root, "foreign-local-cache")
		output, err := managerInvoke(t, manager, roleProject, "install", version, "--platform", runtime.GOOS+"-"+foreignArch, "--local-bin", foreignManager, "--sha256", digest(managerRead(t, foreignManager)), "--install-root", destination)
		if err == nil || !strings.Contains(output, "foreign-platform local") {
			t.Fatalf("foreign local role bypass: %q %v", output, err)
		}
		if _, err := os.Stat(destination); !os.IsNotExist(err) {
			t.Fatalf("refused foreign local artifact initialized cache: %v", err)
		}
	})
	t.Run("ordinary execution rejects mislabeled cached identity", func(t *testing.T) {
		project := filepath.Join(root, "mislabeled-cache-project")
		if err := os.Mkdir(project, 0700); err != nil {
			t.Fatal(err)
		}
		shared := filepath.Join(project, ".agenticlens-config.yaml")
		policy := []byte(fmt.Sprintf("agenticlens:\n  version: %s\n  install_root: %q\n", nextVersion, cache))
		managerWrite(t, shared, policy, 0600)
		mislabeled := filepath.Join(cache, "bin", nextVersion, runtime.GOOS+"-"+runtime.GOARCH, artifactName("agentic-lens-engine"))
		if err := os.MkdirAll(filepath.Dir(mislabeled), 0700); err != nil {
			t.Fatal(err)
		}
		managerWrite(t, mislabeled, engineBytes, 0700)
		output, err := managerInvoke(t, manager, project, "version", "--short")
		if err == nil || !strings.Contains(output, "cached engine identity "+version+" differs from exact pin "+nextVersion) {
			t.Fatalf("ordinary command accepted mislabeled cache: %q %v", output, err)
		}
		if !bytes.Equal(managerRead(t, shared), policy) || !bytes.Equal(managerRead(t, mislabeled), engineBytes) {
			t.Fatal("runtime identity rejection rewrote the pin or cache")
		}
	})
	t.Run("explicit cache root resolves symlink before parent traversal", func(t *testing.T) {
		physical := filepath.Join(root, "physical-root")
		physicalLeaf := filepath.Join(physical, "leaf")
		if err := os.MkdirAll(physicalLeaf, 0700); err != nil {
			t.Fatal(err)
		}
		alias := filepath.Join(root, "logical-root")
		if err := os.Symlink(physicalLeaf, alias); err != nil {
			t.Fatal(err)
		}
		// Keep the parent component literal: filepath.Join would erase the regression input.
		input := alias + "/../native-cache"
		expectedRoot := filepath.Join(physical, "native-cache")
		incorrectRoot := filepath.Join(root, "native-cache")
		output, err := managerInvoke(t, manager, roleProject, "install", version, "--local-bin", engine, "--sha256", digest(engineBytes), "--install-root", input)
		if err != nil {
			t.Fatalf("install through physical root: %q %v", output, err)
		}
		if _, err := os.Stat(incorrectRoot); !os.IsNotExist(err) {
			t.Fatalf("installation collapsed parent traversal before following symlink: %v", err)
		}
		installed := filepath.Join(expectedRoot, "bin", version, runtime.GOOS+"-"+runtime.GOARCH, artifactName("agentic-lens-engine"))
		if !bytes.Equal(managerRead(t, installed), engineBytes) {
			t.Fatal("physical cache received different engine bytes")
		}
		output, err = managerInvoke(t, manager, roleProject, "versions", "--install-root", input)
		if err != nil {
			t.Fatalf("list through physical root: %q %v", output, err)
		}
		var inventory struct {
			InstallRoot string   `json:"install_root"`
			Engines     []string `json:"engines"`
		}
		if err := json.Unmarshal([]byte(output), &inventory); err != nil {
			t.Fatal(err)
		}
		if inventory.InstallRoot != expectedRoot || len(inventory.Engines) != 1 || inventory.Engines[0] != version+"/"+runtime.GOOS+"-"+runtime.GOARCH {
			t.Fatalf("versions inspected a different root: %+v", inventory)
		}
	})
	t.Run("explicit cache repair", func(t *testing.T) {
		managerWrite(t, leaf, []byte("partial source rebuild"), 0700)
		output, err := managerInvoke(t, manager, outside, install...)
		if err == nil || !strings.Contains(output, "--force") || string(managerRead(t, leaf)) != "partial source rebuild" {
			t.Fatalf("implicit repair replaced immutable leaf: %q %v", output, err)
		}
		output, err = managerInvoke(t, manager, outside, append(append([]string{}, install...), "--force")...)
		if err != nil || !bytes.Equal(managerRead(t, leaf), engineBytes) {
			t.Fatalf("explicit repair failed: %q %v", output, err)
		}
	})
	t.Run("self-update replaces the invoked custom installation", func(t *testing.T) {
		target := filepath.Join(root, "custom native bin", "agentic-lens")
		output, err := managerInvoke(t, manager, roleProject, "self-update", "--version", version, "--local-bin", manager, "--sha256", digest(managerBytes), "--dest", target)
		if err != nil {
			t.Fatalf("install custom manager: %q %v", output, err)
		}
		alias := filepath.Join(root, "manager-alias")
		if err := os.Symlink(target, alias); err != nil {
			t.Fatal(err)
		}
		shared := filepath.Join(roleProject, ".agenticlens-config.yaml")
		beforePolicy := managerRead(t, shared)
		output, err = managerInvoke(t, alias, roleProject, "self-update", "--version", nextVersion, "--local-bin", updatedManager, "--sha256", digest(managerRead(t, updatedManager)))
		if err != nil {
			t.Fatalf("default self-update of custom manager: %q %v", output, err)
		}
		identity, err := Validate(target)
		if err != nil || identity.Version != nextVersion || !bytes.Equal(managerRead(t, target), managerRead(t, updatedManager)) {
			t.Fatalf("invoked custom manager was left on old build: %+v %v", identity, err)
		}
		if !bytes.Equal(beforePolicy, managerRead(t, shared)) || !bytes.Equal(engineBytes, managerRead(t, leaf)) {
			t.Fatal("manager update changed project pin or cached engine")
		}
		defaultTarget := filepath.Join(home, ".local", "bin", "agentic-lens")
		for _, path := range []string{defaultTarget, defaultTarget + ".installation.json", defaultTarget + ".sha256", defaultTarget + ".pending.json"} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("self-update created another manager installation at %s: %v", path, err)
			}
		}
	})
	t.Run("manager admission preserves its installation", func(t *testing.T) {
		target := filepath.Join(root, "user-bin", "agentic-lens")
		output, err := managerInvoke(t, manager, outside, "self-update", "--version", version, "--local-bin", manager, "--sha256", digest(managerBytes), "--dest", target)
		if err != nil {
			t.Fatalf("install native manager: %q %v", output, err)
		}
		before := managerSnapshot(t, target)
		for _, test := range []struct{ name, source, wanted, checksum string }{
			{"checksum", manager, version, strings.Repeat("0", 64)},
			{"identity", manager, "v9.4.3", digest(managerBytes)},
			{"engine role", engine, version, digest(engineBytes)},
		} {
			t.Run(test.name, func(t *testing.T) {
				output, err := managerInvoke(t, manager, outside, "self-update", "--version", test.wanted, "--local-bin", test.source, "--sha256", test.checksum, "--dest", target)
				if err == nil {
					t.Fatalf("invalid manager admitted: %q", output)
				}
				for path, expected := range before {
					if managerSnapshot(t, target)[path] != expected {
						t.Fatalf("rejected manager changed %s", path)
					}
				}
			})
		}
	})
	t.Run("malformed project cannot use companion", func(t *testing.T) {
		managerWrite(t, filepath.Join(outside, ".agenticlens-config.yaml"), []byte("agenticlens: [invalid]\n"), 0600)
		output, err := managerInvoke(t, manager, outside, "version", "--short")
		if err == nil || strings.TrimSpace(output) == version {
			t.Fatalf("malformed selected project fell back to companion: %q %v", output, err)
		}
	})
}

func TestManagerDisabledHookValidatesWithoutSourceOrCacheIO(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(project, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	state := filepath.Join(root, "uninitialized-state")
	source := filepath.Join(root, "unavailable-source")
	shared, err := os.ReadFile("../config/testdata/shared-policy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	shared = bytes.Replace(shared, []byte("version: v9.8.7"), []byte("version: v9.4.2"), 1)
	for _, suffix := range []string{"yaml", "yml"} {
		sharedPath := filepath.Join(project, ".agenticlens-config."+suffix)
		managerWrite(t, sharedPath, shared, 0600)
		local := filepath.Join(project, ".agenticlens-config.local.yaml")
		configuration := fmt.Sprintf("telemetry:\n  capture:\n    enabled: false\n    state_dir: %q\n  default_source: codex\n  sources:\n    codex:\n      home: %q\n", state, source)
		managerWrite(t, local, []byte(configuration), 0600)
		payload := []byte(fmt.Sprintf(`{"cwd":%q,"session_id":"session-disabled","hook_event_name":"SessionStart"}`, project))
		var output bytes.Buffer
		if err := Run([]string{"codex-hook"}, bytes.NewReader(payload), &output, &output); err != nil || strings.TrimSpace(output.String()) != "{}" {
			t.Fatalf("disabled hook performed engine selection: %q %v", output.String(), err)
		}
		managerWrite(t, local, []byte(strings.Replace(configuration, "enabled: false", "enabled: wrong", 1)), 0600)
		output.Reset()
		if err := Run([]string{"codex-hook"}, bytes.NewReader(payload), &output, &output); err == nil {
			t.Fatal("invalid disabled policy bypassed validation")
		}
		if err := os.Remove(sharedPath); err != nil {
			t.Fatal(err)
		}
	}
	payload := []byte(fmt.Sprintf(`{"cwd":%q,"session_id":"outside","hook_event_name":"SessionStart"}`, root))
	var output bytes.Buffer
	if err := Run([]string{"codex-hook"}, bytes.NewReader(payload), &output, &output); err != nil || strings.TrimSpace(output.String()) != "{}" {
		t.Fatalf("no-project callback must remain no-op: %q %v", output.String(), err)
	}
	for _, path := range []string{state, source, filepath.Join(home, ".agenticlens")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("disabled/invalid hook initialized %s: %v", path, err)
		}
	}
}

func TestManagerExecutableExcludesEngineDependencyClosure(t *testing.T) {
	module, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "list", "-deps", "./cmd/agentic-lens-manager")
	command.Dir = module
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("resolve actual manager dependency closure: %v\n%s", err, output)
	}
	for _, forbidden := range []string{
		"agenticlens/internal/adapters/capture", "agenticlens/internal/adapters/web",
		"agenticlens/cmd/agentic-lens", "modernc.org/sqlite",
	} {
		for _, imported := range strings.Fields(string(output)) {
			if imported == forbidden || strings.HasPrefix(imported, forbidden+"/") {
				t.Fatalf("manager executable links engine implementation %s", imported)
			}
		}
	}
}
