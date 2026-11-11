package opencode_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"agenticlens/internal/adapters/opencode"
)

func sourceDatabaseFile(t *testing.T, name, marker string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	uri := url.URL{Scheme: "file", Path: filepath.ToSlash(name)}
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE source_marker(value TEXT); INSERT INTO source_marker VALUES(?)`, marker); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestOpenDBAppliesExactBusyPolicyToEveryReadOnlyConnection(t *testing.T) {
	for _, busy := range []int{0, 37} {
		t.Run(strconv.Itoa(busy), func(t *testing.T) {
			name := filepath.Join(t.TempDir(), "source?mode=rw# file.db")
			sourceDatabaseFile(t, name, "selected")
			client, err := opencode.OpenDB(name, busy)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := client.Close(); err != nil {
					t.Error(err)
				}
			})
			client.DB().SetMaxOpenConns(3)
			for i := 0; i < 3; i++ {
				conn, err := client.DB().Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := conn.Close(); err != nil {
						t.Error(err)
					}
				})
				var actualBusy, readOnly int
				if err := conn.QueryRowContext(context.Background(), `PRAGMA busy_timeout`).Scan(&actualBusy); err != nil || actualBusy != busy {
					t.Fatalf("connection %d lock-wait policy: %d, %v; want %d", i, actualBusy, err, busy)
				}
				if err := conn.QueryRowContext(context.Background(), `PRAGMA query_only`).Scan(&readOnly); err != nil || readOnly != 1 {
					t.Fatalf("connection %d read-only policy: %d, %v", i, readOnly, err)
				}
				var marker string
				if err := conn.QueryRowContext(context.Background(), `SELECT value FROM source_marker`).Scan(&marker); err != nil || marker != "selected" {
					t.Fatalf("escaped selected filename: %q, %v", marker, err)
				}
				if _, err := conn.ExecContext(context.Background(), `UPDATE source_marker SET value='changed'`); err == nil {
					t.Fatalf("connection %d allowed source writes", i)
				}
			}
		})
	}
}

func TestOpenDBPreservesPhysicalSymlinkParentTraversal(t *testing.T) {
	root := t.TempDir()
	actual := filepath.Join(root, "actual")
	if err := os.MkdirAll(filepath.Join(actual, "deep"), 0700); err != nil {
		t.Fatal(err)
	}
	sourceDatabaseFile(t, filepath.Join(actual, "source.db"), "physical")
	sourceDatabaseFile(t, filepath.Join(root, "source.db"), "lexical decoy")
	if err := os.Symlink(filepath.Join(actual, "deep"), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	selected := root + string(filepath.Separator) + "link" + string(filepath.Separator) + ".." + string(filepath.Separator) + "source.db"
	client, err := opencode.OpenDB(selected, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	}()
	var marker string
	if err := client.DB().QueryRow(`SELECT value FROM source_marker`).Scan(&marker); err != nil || marker != "physical" {
		t.Fatalf("symlink/.. source selection changed: %q, %v", marker, err)
	}
}

func TestOpenDBNeverSubstitutesMissingOrInvalidSelection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sourceDatabaseFile(t, filepath.Join(home, ".local", "share", "opencode", "opencode.db"), "competing source")
	missing := filepath.Join(t.TempDir(), "missing.db")
	for _, selected := range []string{"", " \t ", missing, t.TempDir()} {
		if client, err := opencode.OpenDB(selected, 0); err == nil || client != nil {
			t.Errorf("selected %q used another location: %v, %v", selected, client, err)
		}
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatalf("missing selected source was created: %v", err)
	}
	name := filepath.Join(t.TempDir(), "source.db")
	sourceDatabaseFile(t, name, "selected")
	if client, err := opencode.OpenDB(name, -1); err == nil || client != nil || !strings.Contains(err.Error(), "database.busy_timeout_ms") {
		t.Fatalf("invalid lock-wait policy silently replaced: %v, %v", client, err)
	}
}
