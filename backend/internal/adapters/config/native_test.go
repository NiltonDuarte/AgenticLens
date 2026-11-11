package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDevelopmentSelectionPrefersValidHostOutputAndRejectsInvalidPresent(t *testing.T) {
	root := t.TempDir()
	dev := filepath.Join(root, "development")
	cache := filepath.Join(root, "cache")
	if err := os.Mkdir(dev, 0700); err != nil {
		t.Fatal(err)
	}
	shared := filepath.Join(root, ".agenticlens-config.yaml")
	body := "agenticlens:\n  version: v9.8.7\n  binary_dir: development\n  install_root: " + cache + "\n"
	if err := os.WriteFile(shared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := ResolveEntrypoint(root)
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := entry.CachedNativePath()
	if err != nil {
		t.Fatal(err)
	}
	if selected, err := entry.NativePath(); err == nil {
		t.Fatalf("missing configured development output selected %s", selected)
	}
	installed := entry
	installed.BinaryDir = ""
	if selected, err := installed.NativePath(); err != nil || selected != pinned {
		t.Fatalf("explicit installed mode: %s %v", selected, err)
	}
	source := filepath.Join(root, "fixture.go")
	program := "package main\nimport (\"fmt\";\"os\";\"strings\")\nfunc main(){switch strings.Join(os.Args[1:],\" \"){case \"capture protocol\":fmt.Println(\"3\");case \"engine protocol\":fmt.Println(\"1\");case \"version --json\":fmt.Println(`{\"version\":\"v0.0.0-development\",\"commit\":\"actual-dev\",\"build_date\":\"fixture\"}`);default:os.Exit(2)}}\n"
	if err := os.WriteFile(source, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	name := "agentic-lens-engine-" + runtime.GOOS + "-" + runtime.GOARCH
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	output := filepath.Join(dev, name)
	build := exec.Command("go", "build", "-o", output, source)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("native interface fixture build: %s %v", out, err)
	}
	if selected, err := entry.NativePath(); err != nil || selected != output {
		t.Fatalf("valid host development output not preferred: %s %v", selected, err)
	}
	identity, err := ValidateNativeExecutable(output)
	if err != nil || identity.Version != "v0.0.0-development" || identity.GitCommit != "actual-dev" {
		t.Fatalf("development identity relabeled: %+v %v", identity, err)
	}
	if err := os.WriteFile(output, []byte("invalid present output"), 0700); err != nil {
		t.Fatal(err)
	}
	if selected, err := entry.NativePath(); err == nil || selected != "" || !strings.Contains(err.Error(), output) {
		t.Fatalf("invalid output fell back: %s %v", selected, err)
	}
	if err := os.WriteFile(shared, []byte("agenticlens:\n  binary_dir: development\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveEntrypoint(root); err == nil {
		t.Fatal("development-only missing pin accepted")
	}
	if err := os.WriteFile(shared, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".agenticlens-config.local.yaml"), []byte("agenticlens:\n  binary_dir: other\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveEntrypoint(root); err == nil {
		t.Fatal("local development override accepted")
	}
}
