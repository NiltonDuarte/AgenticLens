package claude

// All fixtures in this file are synthetic. They reuse the record builders of
// adapter_test.go, carry no transcript content and never read ~/.claude or a
// live ticket workspace.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func subagentDir(store syntheticStore, root string) string {
	return filepath.Join(store.home, "projects", syntheticFolder, root, "subagents")
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func childRecords(start string) []record {
	return []record{
		prompt(start, ""),
		assistant(start, "req-"+start, "model-a", "end_turn", usage(1, 1, 1)),
	}
}

// Every unproven, unreadable or incomplete subagent stays out of the ledger,
// is named with its own reason, and never blocks the proven root and child.
func TestSyntheticUnprovenAndUnreadableChildrenStayUnresolved(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "tool_use", usage(1, 1, 1),
			toolUseBlock("call-good", "Agent", record{"subagent_type": "planner"}),
			toolUseBlock("call-contradicted", "Agent", record{"subagent_type": "coder"}),
			toolUseBlock("call-dup", "Agent", record{"subagent_type": "coder"}),
			toolUseBlock("call-removed", "Agent", record{"subagent_type": "coder"}),
			toolUseBlock("call-foreign", "Agent", record{"subagent_type": "coder"}),
			toolUseBlock("call-noagent", "Agent", record{"subagent_type": "coder"})),
		resultRecord("2026-11-24T10:00:02Z", "call-contradicted", false, record{"agentId": "areported"}),
		assistant("2026-11-24T10:00:03Z", "req-2", "model-a", "end_turn", usage(1, 1, 1)),
	})
	// Proven child that also repeats the root's call-dup identity.
	store.subagent(syntheticFolder, syntheticRoot, "agood", "call-good", "planner", append(childRecords("2026-11-24T10:00:04Z"),
		assistant("2026-11-24T10:00:05Z", "req-good-2", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-dup", "Agent", record{"subagent_type": "coder"}), toolUseBlock("call-self", "Agent", record{}),
			toolUseBlock("call-nested", "Agent", record{"subagent_type": "reviewer"}))))
	store.subagent(syntheticFolder, syntheticRoot, "acontradicted", "call-contradicted", "coder", childRecords("2026-11-24T10:00:06Z"))
	store.subagent(syntheticFolder, syntheticRoot, "aduplicate", "call-dup", "coder", childRecords("2026-11-24T10:00:07Z"))
	store.subagent(syntheticFolder, syntheticRoot, "aunproven", "call-missing", "coder", append(childRecords("2026-11-24T10:00:08Z"),
		assistant("2026-11-24T10:00:09Z", "req-unproven-2", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-under-unproven", "Agent", record{"subagent_type": "coder"}))))
	store.subagent(syntheticFolder, syntheticRoot, "anestedunproven", "call-under-unproven", "coder", childRecords("2026-11-24T10:00:10Z"))
	// A self-referencing meta file: the only matching call is in its own file.
	store.subagent(syntheticFolder, syntheticRoot, "aself", "call-selfloop", "coder", append(childRecords("2026-11-24T10:00:11Z"),
		assistant("2026-11-24T10:00:12Z", "req-self-2", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-selfloop", "Agent", record{}))))
	store.subagent(syntheticFolder, syntheticRoot, "aremoved", "call-removed", "coder", childRecords("2026-11-24T10:00:13Z"))
	foreign := childRecords("2026-11-24T10:00:14Z")
	foreign[1]["sessionId"] = "foreign-session"
	store.subagent(syntheticFolder, syntheticRoot, "aforeign", "call-foreign", "coder", foreign)
	noAgent := childRecords("2026-11-24T10:00:15Z")
	noAgent[1]["agentId"] = nil
	store.subagent(syntheticFolder, syntheticRoot, "anoagent", "call-noagent", "coder", noAgent)
	store.subagent(syntheticFolder, syntheticRoot, "abadmeta", "", "", childRecords("2026-11-24T10:00:16Z"))
	writeFile(t, filepath.Join(subagentDir(store, syntheticRoot), "agent-abadmeta.meta.json"), "{not json")
	store.subagent(syntheticFolder, syntheticRoot, "anotool", "", "", childRecords("2026-11-24T10:00:17Z"))
	writeFile(t, filepath.Join(subagentDir(store, syntheticRoot), "agent-anotool.meta.json"), `{"agentType":"coder"}`)
	writeFile(t, filepath.Join(subagentDir(store, syntheticRoot), "agent-aorphan.meta.json"), `{"agentType":"coder","toolUseId":"call-orphan"}`)
	// The nested call's agent was never written, so it has no transcript.

	files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot})
	if err != nil {
		t.Fatal(err)
	}
	// Removed after discovery: the transcript is listed but cannot be read.
	if err := os.Remove(filepath.Join(subagentDir(store, syntheticRoot), "agent-aremoved.jsonl")); err != nil {
		t.Fatal(err)
	}
	tree, err := ReadTree(context.Background(), files, nil)
	if err != nil {
		t.Fatalf("an unreadable subagent must not fail the tree: %v", err)
	}
	extraction, err := tree.Prove().Facts(laterCutoff, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(extraction.RootIncomplete) != 0 {
		t.Fatalf("root must stay complete: %v", extraction.RootIncomplete)
	}
	got := map[string]bool{}
	for _, session := range extraction.Sessions {
		got[session.ID] = true
	}
	if len(got) != 2 || !got[syntheticRoot] || !got["agood"] {
		t.Fatalf("only the root and the proven child may be recorded, got %v", got)
	}
	diagnostics := strings.Join(extraction.Diagnostics, "\n")
	for id, reason := range map[string]string{
		"acontradicted":   "parent result reports a different agentId",
		"aduplicate":      "matches several Agent tool_use records",
		"aunproven":       "matches no Agent tool_use in this tree",
		"anestedunproven": "parent aunproven is not proven",
		"aself":           "names the subagent's own call",
		"aremoved":        "transcript unreadable",
		"aforeign":        "different sessionId",
		"anoagent":        "does not carry the subagent agentId",
		"abadmeta":        "meta file is not valid JSON",
		"anotool":         "meta file has no toolUseId",
		"aorphan":         "meta file has no transcript",
		"areported":       "reported by its parent but no transcript exists",
	} {
		found := false
		for _, line := range extraction.Diagnostics {
			if strings.Contains(line, " "+id+" ") && strings.Contains(line, reason) && strings.Contains(line, "stored facts preserved") {
				found = true
			}
		}
		if !found {
			t.Errorf("missing diagnostic for %s (%q) in:\n%s", id, reason, diagnostics)
		}
	}
	// A proven parent result for the nested call never reported an agentId,
	// so no child ID can be invented for it.
	good := sessionByID(t, extraction.Sessions, "agood")
	for _, delegation := range good.Delegations {
		if delegation.ChildSession != "" {
			t.Fatalf("a delegation without proof or explicit result invented a child: %+v", good.Delegations)
		}
	}
}

func TestSyntheticSelectedAgentResolvesToItsRootAndUnprovenAgentIsReported(t *testing.T) {
	store := newStore(t)
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1)),
	})
	store.subagent(syntheticFolder, syntheticRoot, "aunproven", "call-none", "coder", childRecords("2026-11-24T10:00:02Z"))
	files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: "aunproven"})
	if err != nil {
		t.Fatal(err)
	}
	if files.RootID != syntheticRoot || files.SelectedID != "aunproven" {
		t.Fatalf("agent ID must resolve to its root and keep the selection: %+v", files)
	}
	tree, err := ReadTree(context.Background(), files, nil)
	if err != nil {
		t.Fatal(err)
	}
	proven := tree.Prove()
	if proven.IsProven("aunproven") || !proven.IsProven(syntheticRoot) {
		t.Fatal("an interface-launched or unproven agent must never be proven")
	}
}

func TestSyntheticDiscoverySelectionErrors(t *testing.T) {
	store := newStore(t)
	other := "00000000-0000-4000-8000-000000000002"
	nonce := "synthetic-nonce-91c2"
	withNonce := prompt("2026-11-24T10:00:00Z", "sdk")
	withNonce["message"] = record{"role": "user", "content": "run " + nonce}
	store.root(syntheticFolder, syntheticRoot, []record{withNonce})
	store.root("-other-folder", other, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	// The nonce in non-transcript files never selects a session.
	writeFile(t, filepath.Join(store.home, "projects", "-other-folder", other, "tool-results", "r.jsonl"), "other-nonce-77\n")
	writeFile(t, filepath.Join(store.home, "projects", "-other-folder", "notes.txt"), "other-nonce-77\n")
	writeFile(t, filepath.Join(store.home, "projects", "-other-folder", other, "subagents", "agent-x.meta.json"), `{"note":"other-nonce-77"}`)

	for _, tc := range []struct {
		name    string
		options DiscoveryOptions
		want    string
		is      error
	}{
		{name: "no selector", options: DiscoveryOptions{ClaudeHome: store.home, Nonce: "   "}, want: "requires --nonce or --session-id"},
		{name: "blank store", options: DiscoveryOptions{ClaudeHome: " ", SessionID: syntheticRoot}, want: "--claude-home"},
		{name: "store without projects", options: DiscoveryOptions{ClaudeHome: t.TempDir(), SessionID: syntheticRoot}, want: "projects directory"},
		{name: "nonce and ID disagree", options: DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, SessionID: other}, want: "different Claude Code root sessions"},
		{name: "nonce only in non-transcript files", options: DiscoveryOptions{ClaudeHome: store.home, Nonce: "other-nonce-77"}, want: "no Claude Code transcript contains"},
		{name: "nonce with absent ID", options: DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, SessionID: "absent"}, is: ErrOwnerUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DiscoverTree(tc.options)
			if err == nil {
				t.Fatal("selection must fail")
			}
			if tc.is != nil && !errors.Is(err, tc.is) || tc.want != "" && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q / %v, got %v", tc.want, tc.is, err)
			}
		})
	}
	files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, SessionID: syntheticRoot})
	if err != nil || files.RootID != syntheticRoot {
		t.Fatalf("agreeing nonce and ID must select the root: %+v %v", files, err)
	}
}

func TestSyntheticAgentIDAmbiguityAndMissingRoot(t *testing.T) {
	store := newStore(t)
	other := "00000000-0000-4000-8000-000000000002"
	store.root(syntheticFolder, syntheticRoot, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	store.root(syntheticFolder, other, []record{prompt("2026-11-24T10:00:00Z", "sdk")})
	store.subagent(syntheticFolder, syntheticRoot, "ashared", "c1", "coder", childRecords("2026-11-24T10:00:01Z"))
	store.subagent(syntheticFolder, other, "ashared", "c2", "coder", childRecords("2026-11-24T10:00:01Z"))
	if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: "ashared"}); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("one agent ID under two roots must be ambiguous: %v", err)
	}
	missingRoot := "00000000-0000-4000-8000-000000000003"
	store.subagent(syntheticFolder, missingRoot, "aorphaned", "c3", "coder", childRecords("2026-11-24T10:00:01Z"))
	if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: "aorphaned"}); !errors.Is(err, ErrOwnerUnavailable) {
		t.Fatalf("an agent whose root transcript is gone must be unavailable: %v", err)
	}
}

func TestSyntheticDiscoveryBudgetBoundaries(t *testing.T) {
	store := newStore(t)
	nonce := "synthetic-nonce-budget"
	withNonce := prompt("2026-11-24T10:00:00Z", "sdk")
	withNonce["message"] = record{"role": "user", "content": nonce}
	store.root(syntheticFolder, syntheticRoot, []record{
		prompt("2026-11-24T10:00:00Z", "sdk"),
		assistant("2026-11-24T10:00:01Z", "req-1", "model-a", "end_turn", usage(1, 1, 1),
			toolUseBlock("call-child", "Agent", record{"subagent_type": "coder"})),
	})
	store.subagent(syntheticFolder, syntheticRoot, "achild", "call-child", "coder", []record{withNonce, assistant("2026-11-24T10:00:02Z", "req-c", "model-a", "end_turn", usage(1, 1, 1))})
	size := func(path string) int64 {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		return info.Size()
	}
	rootBytes := size(rootPath(store))
	childBytes := size(filepath.Join(subagentDir(store, syntheticRoot), "agent-achild.jsonl"))
	metaBytes := size(filepath.Join(subagentDir(store, syntheticRoot), "agent-achild.meta.json"))

	t.Run("invalid limits are not a budget exhaustion", func(t *testing.T) {
		_, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot, Budget: &DiscoveryBudget{MaxSessions: 0, MaxSourceBytes: 1}})
		if err == nil || errors.Is(err, ErrDiscoveryBudget) {
			t.Fatalf("invalid limits must be rejected explicitly: %v", err)
		}
	})
	t.Run("byte budget", func(t *testing.T) {
		budget := &DiscoveryBudget{MaxSessions: 10, MaxSourceBytes: rootBytes - 1}
		files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot, Budget: budget})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ReadTree(context.Background(), files, budget); !errors.Is(err, ErrDiscoveryBudget) {
			t.Fatalf("a root larger than the byte budget must stop with the typed error: %v", err)
		}
		if budget.SourceBytesRead != rootBytes-1 {
			t.Fatalf("every byte actually read is charged: want %d, got %d", rootBytes-1, budget.SourceBytesRead)
		}
	})
	t.Run("byte budget during child read", func(t *testing.T) {
		budget := &DiscoveryBudget{MaxSessions: 10, MaxSourceBytes: rootBytes + 1}
		files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, SessionID: syntheticRoot, Budget: budget})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ReadTree(context.Background(), files, budget); !errors.Is(err, ErrDiscoveryBudget) {
			t.Fatalf("exhaustion while reading a child must fail the tree, not leave the child unresolved: %v", err)
		}
	})
	t.Run("session budget during nonce discovery", func(t *testing.T) {
		budget := &DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 20}
		if _, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, Budget: budget}); !errors.Is(err, ErrDiscoveryBudget) {
			t.Fatalf("nonce discovery beyond the session budget must stop: %v", err)
		}
	})
	t.Run("repeated reads use one session slot per file and charge every byte", func(t *testing.T) {
		budget := &DiscoveryBudget{MaxSessions: 2, MaxSourceBytes: 1 << 20}
		files, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, Budget: budget})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ReadTree(context.Background(), files, budget); err != nil {
			t.Fatalf("re-reading discovered files must not need new session slots: %v", err)
		}
		if budget.SessionsRead != 2 {
			t.Fatalf("want 2 sessions, got %d", budget.SessionsRead)
		}
		if want := 2*(rootBytes+childBytes) + metaBytes; budget.SourceBytesRead != want {
			t.Fatalf("want %d bytes charged, got %d", want, budget.SourceBytesRead)
		}
	})
	t.Run("cancellation before discovery", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := DiscoverTree(DiscoveryOptions{ClaudeHome: store.home, Nonce: nonce, Budget: &DiscoveryBudget{Context: cancelled, MaxSessions: 10, MaxSourceBytes: 1 << 20}})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled discovery must report cancellation: %v", err)
		}
	})
	t.Run("cancelled parse", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := ParseTranscript(cancelled, strings.NewReader("{}\n"), TranscriptExpectation{RootSessionID: syntheticRoot}); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled parse must report cancellation: %v", err)
		}
		if _, err := ParseTranscript(context.Background(), nil, TranscriptExpectation{RootSessionID: syntheticRoot}); err == nil {
			t.Fatal("a nil reader must be an error, not an empty transcript")
		}
	})
}
