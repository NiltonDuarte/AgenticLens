package codex_test

import (
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
)

func TestBoundedDiscoveryNeverReturnsAnIncompleteFamily(t *testing.T) {
	home := t.TempDir()
	writeDiscoveryRollout(t, home, "sessions/a.jsonl", "root", "", "/fixture", "selected-token")
	writeDiscoveryRollout(t, home, "sessions/b.jsonl", "child", "root", "/fixture", "child prompt")
	tooSmall := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 100000}
	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, Nonce: "selected-token", Budget: tooSmall})
	if !errors.Is(err, codex.ErrDiscoveryBudget) || len(paths) != 0 || tooSmall.SessionsRead != 1 {
		t.Fatalf("partial family accepted after session cap: %v, %+v, %v", paths, tooSmall, err)
	}
	budget := &codex.DiscoveryBudget{MaxSessions: 2, MaxSourceBytes: 100000}
	paths, err = codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, SessionID: "root", Budget: budget})
	if err != nil || len(paths) != 2 || budget.SessionsRead != 2 || budget.SourceBytesRead <= 0 {
		t.Fatalf("complete bounded family: %v, %+v, %v", paths, budget, err)
	}
	firstBytes := budget.SourceBytesRead
	paths, err = codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, SessionID: "root", Budget: budget})
	if err != nil || len(paths) != 2 || budget.SessionsRead != 2 || budget.SourceBytesRead <= firstBytes {
		t.Fatalf("seen paths became free source reads or new session attempts: %v, %+v, %v", paths, budget, err)
	}
}

func TestBoundedDiscoveryChargesDecompressedBytesAndRejectsPrefixMatches(t *testing.T) {
	home := t.TempDir()
	name := filepath.Join(home, "sessions", "owner.jsonl.gz")
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(file)
	body := `{"type":"session_meta","payload":{"id":"root"}}` + "\n" + `{"type":"event_msg","payload":{"type":"user_message","message":"selected-token ` + strings.Repeat("x", 100000) + `"}}` + "\n"
	if _, err := compressed.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	budget := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 512}
	paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, Nonce: "selected-token", Budget: budget})
	if len(paths) != 0 || !errors.Is(err, codex.ErrDiscoveryBudget) || budget.SourceBytesRead != 512 {
		t.Fatalf("compressed prefix was treated as complete source evidence: %v, %+v, %v", paths, budget, err)
	}
	full := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: int64(len(body))}
	paths, err = codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: home, Nonce: "selected-token", Budget: full})
	if err != nil || len(paths) != 1 || full.SourceBytesRead != int64(len(body)) {
		t.Fatalf("exact decoded-byte allowance rejected complete input: %v, %+v, %v", paths, full, err)
	}
}

func TestBoundedExactMetadataSelectionReportsExhaustion(t *testing.T) {
	home := t.TempDir()
	name := writeDiscoveryRollout(t, home, "sessions/owner.jsonl", "owner", "", "/fixture", "prompt")
	budget := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 10}
	if path, err := codex.DiscoverExactRolloutPathWithBudget(home, "owner", budget); path != "" || !errors.Is(err, codex.ErrDiscoveryBudget) || budget.SourceBytesRead != 10 {
		t.Fatalf("exact metadata lookup bypassed byte limit: %s, %+v, %v", path, budget, err)
	}
	info, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	budget = &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: info.Size()}
	if path, err := codex.DiscoverExactRolloutPathWithBudget(home, "owner", budget); err != nil || path != name || budget.SourceBytesRead != info.Size() {
		t.Fatalf("complete exact metadata lookup: %s, %+v, %v", path, budget, err)
	}
	for _, invalid := range []*codex.DiscoveryBudget{
		{MaxSessions: 0, MaxSourceBytes: 10},
		{MaxSessions: 1, MaxSourceBytes: 0},
		{MaxSessions: 1, MaxSourceBytes: 100, SeenPaths: map[string]bool{name: true}},
	} {
		if path, err := codex.DiscoverExactRolloutPathWithBudget(home, "owner", invalid); err == nil || path != "" {
			t.Fatalf("invalid explicit policy became unlimited: %q, %+v, %v", path, invalid, err)
		}
	}
}

func TestCodexSelectionNeverDiscoversAnotherHome(t *testing.T) {
	home := t.TempDir()
	writeDiscoveryRollout(t, home, "sessions/owner.jsonl", "owner", "", "/fixture", "token")
	t.Setenv("CODEX_HOME", home)
	if paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{SessionID: "owner"}); err == nil || len(paths) != 0 {
		t.Fatalf("missing selected home used environment: %v, %v", paths, err)
	}
	if path, err := codex.DiscoverExactRolloutPath("", "owner"); err == nil || path != "" {
		t.Fatalf("exact selection used another home: %q, %v", path, err)
	}
	if err := codex.ValidateRolloutStore("", filepath.Join(home, "sessions", "owner.jsonl")); err == nil {
		t.Fatal("empty selected home became a store authority")
	}
	if paths, err := codex.DiscoverRolloutPaths(codex.DiscoveryOptions{CodexHome: filepath.Join(t.TempDir(), "missing"), SessionID: "owner"}); err == nil || len(paths) != 0 {
		t.Fatalf("unavailable selected home used environment: %v, %v", paths, err)
	}
}

func TestKnownNativeMetadataRespectsExplicitByteAllowance(t *testing.T) {
	home := t.TempDir()
	name := writeDiscoveryRollout(t, home, "sessions/owner.jsonl", "owner", "", "/fixture", "prompt")
	body, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	lineBytes := int64(strings.IndexByte(string(body), '\n') + 1)
	if lineBytes < 1 {
		t.Fatal("fixture metadata record lacks a line boundary")
	}
	if metadata, err := codex.ReadSessionContextWithinBudget(name, 1); err == nil || metadata.ID != "" {
		t.Fatalf("tiny selected allowance read native identity: %+v, %v", metadata, err)
	}
	if metadata, err := codex.ReadSessionContextWithinBudget(name, lineBytes); err != nil || metadata.ID != "owner" || metadata.WorkingDirectory != "/fixture" {
		t.Fatalf("complete metadata record rejected at exact byte boundary: %+v, %v", metadata, err)
	}
	if metadata, err := codex.ReadSessionContextWithinBudget(name, 0); err == nil || metadata.ID != "" {
		t.Fatalf("zero became an implicit metadata allowance: %+v, %v", metadata, err)
	}
	prefix := `{"type":"session_meta","payload":{"id":"owner","cwd":"/fixture"}}`
	if err := os.WriteFile(name, []byte(prefix+strings.Repeat(" ", 100)), 0600); err != nil {
		t.Fatal(err)
	}
	if metadata, err := codex.ReadSessionContextWithinBudget(name, int64(len(prefix))); !errors.Is(err, codex.ErrDiscoveryBudget) || metadata.ID != "" {
		t.Fatalf("JSON prefix without a proven record boundary was accepted: %+v, %v", metadata, err)
	}
}
