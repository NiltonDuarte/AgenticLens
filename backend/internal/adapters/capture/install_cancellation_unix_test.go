//go:build linux || darwin

package capture

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallOptionalProbeBoundsInheritedPipes(t *testing.T) {
	for _, source := range []string{"opencode", "codex"} {
		t.Run(source, func(t *testing.T) {
			dir := t.TempDir()
			body := "#!/bin/sh\nsleep 2 &\nwait\n"
			if source == "codex" {
				body = "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'codex-cli 0.166.0'; exit 0; fi\nsleep 2 &\nwait\n"
			}
			if err := os.WriteFile(filepath.Join(dir, source), []byte(body), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			started := time.Now()
			probe(ctx, source)
			if elapsed := time.Since(started); elapsed > time.Second {
				t.Fatalf("100ms optional probe waited for inherited pipes: %s", elapsed)
			}
			if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Fatal("fixture did not reach optional deadline")
			}
		})
	}
}

func TestInstallerFinalWriteCancellationRollsBack(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		t.Run(source, func(t *testing.T) {
			f := newInstallTransactionFixture(t, source)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			final := f.changes[len(f.changes)-1].path
			err := applyInstallTransactionWithWriter(ctx, f.options, f.repo, f.state, f.manifest, f.changes, func(c fileChange) error {
				if err := apply(c); err != nil {
					return err
				}
				if c.path == final {
					cancel()
				}
				return nil
			})
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("last writer canceled required context; transaction returned %v", err)
			}
			assertInstallFiles(t, f.changes, false)
			if _, err := os.Stat(installJournalPath(f.state, source)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("completed rollback retained an unexpected journal: %v", err)
			}
			if err := applyInstallTransaction(context.Background(), f.options, f.repo, f.state, f.manifest, f.changes); err != nil {
				t.Fatalf("uncanceled retry failed: %v", err)
			}
			assertInstallFiles(t, f.changes, true)
		})
	}
}
