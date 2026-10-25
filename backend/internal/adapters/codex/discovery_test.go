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
	rootPath := writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-root.jsonl", "root", "", "/repo", "working on TEST-101")
	childPath := writeDiscoveryRollout(t, home, "archived_sessions/rollout-child.jsonl.gz", "child", "root", "/repo", "TEST-101 subagent")
	writeDiscoveryRollout(t, home, "sessions/2026/10/24/rollout-other.jsonl", "other", "", "/other", "working on OTHER-202")

	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "TEST-101", WorkingDir: "/repo"})
	if err != nil {
		t.Fatalf("DiscoverRolloutPaths returned an error: %v", err)
	}
	if !reflect.DeepEqual(paths, []string{rootPath, childPath}) {
		t.Fatalf("unexpected discovered paths: %#v", paths)
	}
}

func TestDiscoverRolloutPathsNonceTakesPriorityOverTicket(t *testing.T) {
	home := t.TempDir()
	newer := writeDiscoveryRollout(t, home, "sessions/2026/10/25/rollout-newer.jsonl", "newer", "", "/repo", "TEST-101 general context")
	older := writeDiscoveryRollout(t, home, "sessions/2026/10/24/rollout-nonce.jsonl", "nonce-thread", "", "/repo", "nonce fixture-nonce-deadbeef")
	if err := os.Chtimes(older, time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "TEST-101", Nonce: "fixture-nonce-deadbeef", WorkingDir: "/repo"})
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
	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "TEST-101", WorkingDir: "/repo"})
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
	_, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, TicketID: "TEST-101", WorkingDir: "/repo"})
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

func TestDiscoveryReadsNestedChildProvenanceBeforeInheritedMetadata(t *testing.T) {
	home := t.TempDir()
	rootPath := writeDiscoveryRollout(t, home, "sessions/root.jsonl", "root", "", "/fixture", "selected-nonce")
	childPath := filepath.Join(home, "sessions", "child.jsonl")
	child := `{"type":"session_meta","payload":{"id":"child","source":{"subagent":{"thread_spawn":{"parent_thread_id":"root","agent_role":"tester"}}}}}
{"type":"session_meta","payload":{"id":"root","source":"cli"}}
`
	if err := os.WriteFile(childPath, []byte(child), 0600); err != nil {
		t.Fatal(err)
	}
	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, Nonce: "selected-nonce"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{rootPath, childPath}) {
		t.Fatalf("nested provenance lost: %v", paths)
	}
	transcript, err := codex.ParseRollout(strings.NewReader(child))
	if err != nil {
		t.Fatal(err)
	}
	if len(transcript.Invocations) != 1 || string(transcript.Invocations[0].Role) != "tester" {
		t.Fatalf("explicit child role lost: %+v", transcript.Invocations)
	}
}
