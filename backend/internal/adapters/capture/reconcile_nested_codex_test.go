package capture_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

func TestReconcileCodexRecoversNestedSpawnWithoutChildCallbacks(t *testing.T) {
	const (
		ticket         = "TEST-100"
		root           = "019c0000-0000-7000-8000-000000000001"
		implementation = "019c0000-0000-7000-8000-000000000002"
		verification   = "019c0000-0000-7000-8000-000000000003"
		nested         = "019c0000-0000-7000-8000-000000000004"
		unrelated      = "019c0000-0000-7000-8000-000000000005"
		passBudget     = 12
	)
	for _, tc := range []struct {
		name, nestedParent, spawnReference string
		wantNested, wantComplete           bool
	}{
		{"source confirms reported child", implementation, nested, true, true},
		{"source parent contradicts reported child", verification, nested, false, false},
		{"opaque result is not a native session ID", implementation, "opaque-agent", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			project := t.TempDir()
			captureProjectConfig(t, project, "Nested native Codex source")
			store, err := capture.OpenProject(t.TempDir(), project)
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			home := t.TempDir()
			active := filepath.Join(home, "sessions", "2026", "01", "01")
			cache := filepath.Join(home, ".tmp", "plugins")
			for _, directory := range []string{active, cache} {
				if err := os.MkdirAll(directory, 0700); err != nil {
					t.Fatal(err)
				}
			}
			write := func(directory, id, parent string, depth, tokens int, spawn string) string {
				t.Helper()
				source := `"cli"`
				if parent != "" {
					source = fmt.Sprintf(`{"subagent":{"thread_spawn":{"parent_thread_id":%q,"depth":%d}}}`, parent, depth)
				}
				body := fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":%q,"cwd":%q,"source":%s}}
{"timestamp":"2026-01-01T00:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}
{"timestamp":"2026-01-01T00:00:02Z","type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}
`, id, project, source, id, tokens)
				if spawn != "" {
					result := fmt.Sprintf(`{"agent_id":%q}`, spawn)
					body += fmt.Sprintf(`{"timestamp":"2026-01-01T00:00:03Z","type":"response_item","payload":{"type":"function_call","call_id":"spawn-nested","name":"spawn_agent","arguments":"{}"}}
{"timestamp":"2026-01-01T00:00:04Z","type":"response_item","payload":{"type":"function_call_output","call_id":"spawn-nested","output":%q}}
`, result)
				}
				path := filepath.Join(directory, "rollout-2026-01-01T00-00-00-"+id+".jsonl")
				if err := os.WriteFile(path, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
				return path
			}
			rootPath := write(active, root, "", 0, 11, "")
			coderPath := write(active, implementation, root, 1, 19, tc.spawnReference)
			verifierPath := write(active, verification, root, 1, 23, "")
			write(active, nested, tc.nestedParent, 2, 31, "")
			write(active, unrelated, "", 0, 47, "")
			write(cache, nested, implementation, 2, 99, "")

			registerBound(t, store, home, rootPath, root, ticket)
			// Codex delivered only the root's direct-child lifecycle evidence.
			// The implementation child's response names the nested source.
			for _, directPath := range []string{coderPath, verifierPath} {
				if err := store.QueueChild(ctx, home, directPath, root); err != nil {
					t.Fatal(err)
				}
			}
			policy := fixtureGenerationPolicy()
			policy.Delegations = false
			reconcile := func() capture.ReconcileResult {
				t.Helper()
				result, err := store.Reconcile(ctx, ticket, "native fixture", passBudget, fixtureByteBudget, policy)
				if err != nil {
					t.Fatal(err)
				}
				if result.SourceAttempts > passBudget || result.SourceBytes > fixtureByteBudget {
					t.Fatalf("nested discovery exceeded bounded source work: attempts=%d bytes=%d", result.SourceAttempts, result.SourceBytes)
				}
				return result
			}
			var latest capture.ReconcileResult
			wantOwners := 3
			if tc.wantNested {
				wantOwners++
			}
			for range 10 {
				latest = reconcile()
				if tc.wantComplete && len(latest.Sessions) == wantOwners && !latest.DiscoveryPending && latest.PendingChildren == 0 && latest.UnresolvedParentLinks == 0 {
					break
				}
				for _, row := range latest.Sessions {
					if row.ID == unrelated || (!tc.wantNested && row.ID == nested) {
						t.Fatalf("source without a matching parent was imported: %+v", row)
					}
				}
			}
			type expectedSession struct {
				parent string
				tokens int64
			}
			expected := map[string]expectedSession{root: {"", 11}, implementation: {root, 19}, verification: {root, 23}}
			if tc.wantNested {
				expected[nested] = expectedSession{implementation, 31}
			}
			if tc.wantComplete && (latest.DiscoveryPending || latest.PendingChildren != 0 || latest.UnresolvedParentLinks != 0) {
				t.Fatalf("capture did not settle: discovery=%t pending=%d links=%d", latest.DiscoveryPending, latest.PendingChildren, latest.UnresolvedParentLinks)
			}
			if len(latest.Sessions) != len(expected) {
				t.Fatalf("expected %d exact owners, got %+v", len(expected), latest.Sessions)
			}
			seen := make(map[string]bool, len(expected))
			for _, session := range latest.Sessions {
				parent := ""
				if session.ParentID != nil {
					parent = *session.ParentID
				}
				want, ok := expected[session.ID]
				if !ok || seen[session.ID] || parent != want.parent || session.Source != domain.SourceCodex ||
					len(session.Invocations) != 1 || session.Invocations[0].Tokens.Input != want.tokens || session.Delegations != nil {
					t.Fatalf("incorrect source owner, usage or parent link: %+v", session)
				}
				seen[session.ID] = true
			}
			if tc.wantNested && !reflect.DeepEqual(latest.Sessions, reconcile().Sessions) {
				t.Fatal("repeat reconciliation changed or duplicated source-confirmed nested facts")
			}
		})
	}
}
