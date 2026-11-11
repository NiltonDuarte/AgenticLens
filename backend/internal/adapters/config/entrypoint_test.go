package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEntrypointLiteralAuthorityAndLocalPaths(t *testing.T) {
	for _, quoted := range []string{"v9.8.7-rc.42+abc1234", `"v9.8.7-rc.42+abc1234"`, "'v9.8.7-rc.42+abc1234'"} {
		t.Run(quoted, func(t *testing.T) {
			root := t.TempDir()
			nested := filepath.Join(root, "deep", "directory")
			if err := os.MkdirAll(nested, 0700); err != nil {
				t.Fatal(err)
			}
			shared := filepath.Join(root, ".agenticlens-config.yml")
			data := "\xef\xbb\xbf# comment\r\nagenticlens: # authority\r\n  version:   " + quoted + " # release\r\n  install_root: /shared/location\r\ntelemetry:\r\n  capture:\r\n    state_dir: state\r\n"
			if err := os.WriteFile(shared, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			install := filepath.Join(t.TempDir(), "local path")
			local := filepath.Join(root, ".agenticlens-config.local.yaml")
			if err := os.WriteFile(local, []byte("agenticlens:\n  install_root: '"+filepath.ToSlash(install)+"'\ntelemetry:\n  capture:\n    enabled: true\n    state_dir: state\n    commit_warning: false\n"), 0600); err != nil {
				t.Fatal(err)
			}
			entry, err := ResolveEntrypoint(nested)
			if err != nil {
				t.Fatal(err)
			}
			if entry.Project != root || entry.ConfigPath != shared || entry.Version != "v9.8.7-rc.42+abc1234" || entry.InstallRoot != install {
				t.Fatalf("entry %+v", entry)
			}
			policy, err := ResolveCapture(nested, "")
			if err != nil || policy.CommitWarning || policy.StateDir != filepath.Join(root, "state") {
				t.Fatalf("policy %+v %v", policy, err)
			}
			if err := os.WriteFile(local, []byte("agenticlens:\n  version: v9.8.7-rc.42+abc1234\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := ResolveEntrypoint(nested); err == nil {
				t.Fatal("local pin presence accepted")
			}
		})
	}
}

func TestEntrypointRejectsAmbiguousUnsafeUnsupportedAndOversized(t *testing.T) {
	for _, data := range []string{
		"agenticlens:\n  version: v1.2.3\n  version: v2.0.0\n",
		"agenticlens:\n  version: v1.2.3\nagenticlens:\n  version: v2.0.0\n",
		"agenticlens:\n  version: latest\n", "agenticlens:\n  version: \u00a0v1.2.3\u00a0\n", "agenticlens:\n  version: ../v1.2.3\n",
		"agenticlens:\n  version: \"v1.2.3+$(touch INJECTED)\"\n",
		"agenticlens:\n  version: v1.2.3/other\n",
		"agenticlens: *pin\n",
		"agenticlens:\n  <<: *pin\n  version: v1.2.3\n", "agenticlens:\n  version: |\n    v1.2.3\n",
		"agenticlens:\n\tversion: v1.2.3\n",
		"agenticlens:\n  version: v1.2.3\n---\nproject_name: x\n",
		"agenticlens:\n  version:v1.2.3\n", "project_name: Missing\n",
		"agenticlens:\n  version: v1.2.3\n  install_root: relative/path\n",
		"agenticlens:\n  version: v1.2.3\n#" + strings.Repeat("x", bootstrapLimit),
	} {
		t.Run(data[:min(len(data), 60)], func(t *testing.T) {
			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := ResolveEntrypoint(root); err == nil {
				t.Fatal("invalid pin accepted")
			}
		})
	}
}

func TestEntrypointStopsAtNearerConfigurationAndGitBoundary(t *testing.T) {
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, ".agenticlens-config.yaml"), []byte("agenticlens:\n  version: v1.2.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(parent, "child")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".agenticlens-config.yml")
	if err := os.WriteFile(path, []byte("agenticlens: [invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveEntrypoint(root); err == nil {
		t.Fatal("invalid nearer configuration bypassed")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: somewhere\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveEntrypoint(root); err == nil {
		t.Fatal("crossed Git boundary")
	}
}

func TestGeneratedPinnedConfigRoundTrips(t *testing.T) {
	data, err := fixtureSharedYAML(t, "John's Project", "custom-work")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	path := filepath.Join(root, ".agenticlens-config.yaml")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AgenticLens.Version != "v9.8.7" || cfg.Workspaces.Dir != "custom-work" || cfg.Telemetry.Capture.StateDir != "" {
		t.Fatalf("configuration %+v", cfg)
	}
	entry, err := ResolveEntrypoint(root)
	if err != nil || entry.Version != cfg.AgenticLens.Version {
		t.Fatalf("entry %+v %v", entry, err)
	}
	t.Setenv("AGENTIC_LENS_VERSION", "v2.0.0")
	if _, err := ResolveEntrypoint(root); err == nil {
		t.Fatal("environment pin override accepted")
	}
}

func TestScaffoldCompletionAcceptsOnlyExactOwnedVersion(t *testing.T) {
	data, err := BootstrapScaffold("v9.8.7")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateBootstrapScaffold(data, "v9.8.7"); err != nil {
		t.Fatal(err)
	}
	for _, changed := range [][]byte{append(append([]byte{}, data...), []byte("project_name: Owner\n")...), []byte(strings.Replace(string(data), "v9.8.7", "v9.8.8", 1)), []byte(strings.Replace(string(data), "incomplete", "edited", 1))} {
		if err := ValidateBootstrapScaffold(changed, "v9.8.7"); err == nil {
			t.Fatal("changed scaffold granted replacement authority")
		}
	}
}

func TestEntrypointAcceptsSupportedYAMLRepresentations(t *testing.T) {
	for _, body := range []string{`agenticlens: {version: v1.2.3}`, "\"agenticlens\":\n  version: v1.2.3\n", "agenticlens:\n  version: !!str v1.2.3\n", "agenticlens:\n  version: &pin v1.2.3\n", "agenticlens:\n    version: v1.2.3\n"} {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.yaml"), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		got, err := ResolveEntrypoint(root)
		if err != nil || got.Version != "v1.2.3" {
			t.Fatalf("supported YAML changed exactpin: %+v %v", got, err)
		}
	}
}
