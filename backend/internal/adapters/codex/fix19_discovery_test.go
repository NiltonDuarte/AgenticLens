package codex_test

import (
	"agenticlens/internal/adapters/codex"
	"strings"
	"testing"
)

func TestFIX19DiscoveryRejectsCopiesBeforeResolvingRelationships(t *testing.T) {
	for _, archive := range []string{"archived_sessions/copy.jsonl", "archived_sessions/copy.jsonl.gz"} {
		for _, parent := range []string{"root", "other"} {
			t.Run(archive+"/"+parent, func(t *testing.T) {
				home := t.TempDir()
				writeDiscoveryRollout(t, home, "sessions/root.jsonl", "root", "", "/fixture", "root-nonce")
				writeDiscoveryRollout(t, home, "sessions/other.jsonl", "other", "", "/fixture", "other-nonce")
				writeDiscoveryRollout(t, home, "sessions/child.jsonl", "child", "root", "/fixture", "child")
				writeDiscoveryRollout(t, home, archive, "child", parent, "/fixture", "child")
				for _, selector := range []codex.DiscoveryOptions{{CodexHome: home, SessionID: "root"}, {CodexHome: home, SessionID: "child"}, {CodexHome: home, Nonce: "root-nonce"}} {
					for repeat := 0; repeat < 2; repeat++ {
						paths, err := codex.DiscoverRolloutPaths(selector)
						if err == nil || !strings.Contains(err.Error(), "ambiguous") || !strings.Contains(err.Error(), "child") || len(paths) != 0 {
							t.Fatalf("selected family accepted duplicate relationship: paths=%v err=%v", paths, err)
						}
					}
				}
				if _, err := codex.DiscoverExactRolloutPath(home, "child"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
					t.Fatalf("exact discovery lost rejection: %v", err)
				}
				if _, err := codex.DiscoverExactRolloutPath(home, "root"); err != nil {
					t.Fatalf("unambiguous exact owner hidden: %v", err)
				}
				if parent == "root" {
					if paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, SessionID: "other"}); err != nil || len(paths) != 1 {
						t.Fatalf("unrelated family hidden: %v %v", paths, err)
					}
				}
			})
		}
	}
}
