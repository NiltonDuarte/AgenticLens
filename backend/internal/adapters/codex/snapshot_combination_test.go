package codex_test

import (
	"strings"
	"testing"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
)

func TestCombineTranscriptsRejectsRepeatedSnapshotOwners(t *testing.T) {
	row := domain.InvocationFact{Role: "coder", Model: "model", Tokens: domain.InvocationTokens{Input: 10}}
	root := codex.Transcript{ThreadID: "root", Invocations: []domain.InvocationFact{row, row}}
	child := codex.Transcript{ThreadID: "child", ParentThread: "root"}
	nested, err := codex.CombineTranscripts([]codex.Transcript{root, child})
	if err != nil {
		t.Fatal(err)
	}
	warning := root
	warning.Warnings = []string{"partial record"}
	divergent := root
	divergent.Invocations = []domain.InvocationFact{row}
	for name, transcripts := range map[string][]codex.Transcript{
		"identical": {root, root},
		"divergent": {root, divergent},
		"warning":   {root, warning},
		"nested":    {nested, child},
	} {
		t.Run(name, func(t *testing.T) {
			combined, err := codex.CombineTranscripts(transcripts)
			if err == nil || !strings.Contains(err.Error(), "ambiguous") || combined.ToolCalls != 0 || len(combined.Invocations) != 0 {
				t.Fatalf("duplicate owner produced facts: combined=%+v err=%v", combined, err)
			}
		})
	}
	facts, err := nested.SessionFacts()
	if err != nil || len(facts) != 2 || len(facts[0].Invocations) != 2 {
		t.Fatalf("legitimate repeated invocations changed: facts=%+v err=%v", facts, err)
	}
}
