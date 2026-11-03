//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEntrypointPhysicalProjectHomeCacheAndLiteralSpacePath(t *testing.T) {
	base, home, cache := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	if err := os.Mkdir(filepath.Join(cache, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(cache, "child"), filepath.Join(home, "cache-link")); err != nil {
		t.Fatal(err)
	}
	physical := filepath.Join(base, "physical project")
	if err := os.MkdirAll(filepath.Join(physical, "child"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(filepath.Join(physical, "child"), link); err != nil {
		t.Fatal(err)
	}
	body := []byte("agenticlens:\n  version: v1.2.3\n  install_root: '~/cache-link/..'\n")
	if err := os.WriteFile(filepath.Join(physical, ".agenticlens-config.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	entry, err := ResolveEntrypoint(link + "/..")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Project != physical || entry.InstallRoot != cache {
		t.Fatalf("lexical cleaning changed physical context %+v", entry)
	}
	literal := filepath.Join(base, "literal trailing space ")
	if err := os.Mkdir(literal, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(literal, ".agenticlens-config.yaml"), body, 0600); err != nil {
		t.Fatal(err)
	}
	entry, err = ResolveEntrypoint(literal)
	if err != nil || entry.Project != literal {
		t.Fatalf("literal path whitespace lost %+v %v", entry, err)
	}
}
