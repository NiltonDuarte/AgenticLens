package main_test

import (
	"bytes"
	"compress/gzip"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
)

func TestFIX19AutomaticGenerateRejectsRegisteredOwnerCopies(t *testing.T) {
	for _, variant := range []string{"identical-jsonl", "divergent-gzip", "warning-jsonl"} {
		t.Run(variant, func(t *testing.T) {
			f := newCaptureGenerationFixture(t)
			path := f.write(t, 17)
			f.bind(t, path)
			accepted := f.generate(t)
			assertCaptureRootInput(t, accepted, 17)
			f.write(t, 29)
			body := readCaptureFile(t, path)
			if strings.HasPrefix(variant, "divergent") {
				body = bytes.Replace(body, []byte(`"input_tokens":29`), []byte(`"input_tokens":99`), 1)
			}
			if strings.HasPrefix(variant, "warning") {
				body = append(body, []byte("{malformed\n")...)
			}
			suffix := ".jsonl"
			if strings.HasSuffix(variant, "gzip") {
				suffix += ".gz"
			}
			copyPath := filepath.Join(f.store, "archived_sessions", "copy"+suffix)
			if err := os.MkdirAll(filepath.Dir(copyPath), 0700); err != nil {
				t.Fatal(err)
			}
			if strings.HasSuffix(copyPath, ".gz") {
				file, err := os.Create(copyPath)
				if err != nil {
					t.Fatal(err)
				}
				gz := gzip.NewWriter(file)
				if _, err = gz.Write(body); err != nil {
					t.Fatal(err)
				}
				if err = gz.Close(); err != nil {
					t.Fatal(err)
				}
				if err = file.Close(); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(copyPath, body, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := codex.DiscoverExactRolloutPath(f.store, "root"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
				t.Fatalf("fixture did not reach ambiguity: %v", err)
			}
			for repeat := 0; repeat < 2; repeat++ {
				cmd := exec.Command(cliBinaryPath, f.args(t)...)
				cmd.Dir = f.project
				out, err := cmd.CombinedOutput()
				if !strings.Contains(string(out), "ambiguous") || !strings.Contains(strings.ToLower(string(out)), "incomplete") {
					t.Fatalf("registered owner ambiguity hidden: %v %s", err, out)
				}
				got := f.ledger(t)
				if !reflect.DeepEqual(accepted.Sessions, got.Sessions) {
					t.Fatalf("failed automatic extraction replaced accepted owner facts: got=%+v want=%+v", got.Sessions, accepted.Sessions)
				}
			}
		})
	}
}
