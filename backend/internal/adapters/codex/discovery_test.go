package codex_test

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/codex"
)

func TestDiscoverRolloutPathsMatchesTicketAndIncludesChildThreads(t *testing.T) {
	home := t.TempDir()
	rootPath := writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-root.jsonl", "root", "", "/repo", "working on GEN-07")
	childPath := writeDiscoveryRollout(t, home, "archived_sessions/rollout-child.jsonl.gz", "child", "root", "/repo", "GEN-07 subagent")
	writeDiscoveryRollout(t, home, "sessions/2026/10/24/rollout-other.jsonl", "other", "", "/other", "working on GEN-08")

	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "GEN-07", WorkingDir: "/repo"})
	if err != nil {
		t.Fatalf("DiscoverRolloutPaths returned an error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{rootPath, childPath}) {
		t.Fatalf("unexpected discovered paths: %#v", paths)
	}
}

func TestDiscoverRolloutPathsNonceTakesPriorityOverTicket(t *testing.T) {
	home := t.TempDir()
	newer := writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-newer.jsonl", "newer", "", "/repo", "GEN-07 general context")
	older := writeDiscoveryRollout(t, home, "sessions/2026/10/24/rollout-nonce.jsonl", "nonce-thread", "", "/repo", "nonce GEN07-deadbeef")
	if err := os.Chtimes(older, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "GEN-07", Nonce: "GEN07-deadbeef", WorkingDir: "/repo"})
	if err != nil {
		t.Fatalf("DiscoverRolloutPaths returned an error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{older}) || strings.Contains(strings.Join(paths, " "), newer) {
		t.Fatalf("nonce did not identify its exact thread: %#v", paths)
	}
}

func TestDiscoverRolloutPathsUsesUniqueWorkingDirectoryFallback(t *testing.T) {
	home := t.TempDir()
	path := writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout.jsonl", "only", "", "/repo", "no ticket reference")
	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "GEN-07", WorkingDir: "/repo"})
	if err != nil {
		t.Fatalf("DiscoverRolloutPaths returned an error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{path}) {
		t.Fatalf("unexpected fallback paths: %#v", paths)
	}
}

func TestDiscoverRolloutPathsRejectsAmbiguousWorkingDirectory(t *testing.T) {
	home := t.TempDir()
	writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-one.jsonl", "one", "", "/repo", "no ticket")
	writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-two.jsonl", "two", "", "/repo", "no ticket")
	_, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "GEN-07", WorkingDir: "/repo"})
	if err == nil || !strings.Contains(err.Error(), "--nonce or --session-id") {
		t.Fatalf("expected ambiguous discovery error, got %v", err)
	}
}

func writeDiscoveryRollout(t *testing.T, home, relativePath, id, parentID, cwd, userText string) string {
	t.Helper()
	path := filepath.Join(home, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	metadata, err := json.Marshal(map[string]string{
		"id": id, "session_id": id, "cwd": cwd, "timestamp": "2026-10-25T10:00:00Z", "parent_thread_id": parentID,
	})
	if err != nil {
		t.Fatal(err)
	}
	transcriptLine, err := json.Marshal(map[string]string{"type": "user_message", "message": userText})
	if err != nil {
		t.Fatal(err)
	}
	contents := "{\"type\":\"session_meta\",\"payload\":" + string(metadata) + "}\n" + string(transcriptLine) + "\n"
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(path, ".gz") {
		writer := gzip.NewWriter(file)
		if _, err := writer.Write([]byte(contents)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
	} else if _, err := file.WriteString(contents); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}
