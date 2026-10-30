package codex_test

import (
	"reflect"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

func TestCombinedCodexSessionsPreserveRepeatedRowsAndNestedOwners(t *testing.T) {
	row := domain.InvocationFact{Role: "coder", Model: "shared-model", Tokens: domain.InvocationTokens{Input: 10}, CostUSD: 0.25}
	otherModel := row
	otherModel.Model = "other-model"
	transcripts := []codex.Transcript{
		{ThreadID: "root-thread", SessionID: "root-session", Invocations: []domain.InvocationFact{row, row, otherModel}},
		{ThreadID: "child-thread", SessionID: "child-session", ParentThread: "root-thread", Invocations: []domain.InvocationFact{row, row}},
		{ThreadID: "grandchild-thread", SessionID: "grandchild-session", ParentThread: "child-thread", Invocations: []domain.InvocationFact{row}},
	}
	direct, err := codex.CombineTranscripts(transcripts)
	if err != nil {
		t.Fatal(err)
	}
	subtree, err := codex.CombineTranscripts(transcripts[1:])
	if err != nil {
		t.Fatal(err)
	}
	nested, err := codex.CombineTranscripts([]codex.Transcript{transcripts[0], subtree})
	if err != nil {
		t.Fatal(err)
	}
	for name, combined := range map[string]codex.Transcript{"direct": direct, "nested": nested} {
		t.Run(name, func(t *testing.T) {
			sessions, err := combined.SessionFacts()
			if err != nil {
				t.Fatal(err)
			}
			if len(sessions) != len(transcripts) {
				t.Fatalf("expected three enclosing session owners, got %+v", sessions)
			}
			for i, transcript := range transcripts {
				got := sessions[i]
				if got.ID != transcript.ThreadID || got.Source != domain.SourceCodex || !reflect.DeepEqual(got.Invocations, transcript.Invocations) {
					t.Fatalf("owner %d lost its explicit thread or metric rows: %+v", i, got)
				}
				if transcript.ParentThread == "" {
					if got.ParentID != nil {
						t.Fatalf("root gained a parent: %+v", got)
					}
				} else if got.ParentID == nil || *got.ParentID != transcript.ParentThread {
					t.Fatalf("owner %d lost its parent %q: %+v", i, transcript.ParentThread, got)
				}
			}
			metrics := domain.DeriveMetricsLedger(domain.AgenticFactLedger{Sessions: sessions}, domain.DefaultProjectConfig())
			if metrics.Totals.Invocations != 6 || metrics.Totals.InputTokens != 60 || metrics.Totals.CostUSD != 1.5 {
				t.Fatalf("combined rows were collapsed or counted twice: %+v", metrics.Totals)
			}
			if len(metrics.Threads) != 1 || metrics.Threads[0].SessionID != "root-thread" {
				t.Fatalf("combined hierarchy produced the wrong root: %+v", metrics.Threads)
			}
			var threadRows int64
			for _, role := range metrics.Threads[0].Roles {
				for _, model := range role.Models {
					threadRows += model.Invocations
				}
			}
			if threadRows != 6 {
				t.Fatalf("thread model groups collapsed repeated rows: got %d", threadRows)
			}
		})
	}
}
