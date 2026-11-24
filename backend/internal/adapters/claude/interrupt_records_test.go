package claude

// Synthetic regression fixtures for AL-98 review finding BEH-01. They hold
// only record keys and the harness markers Claude Code 2.1.331 writes for Esc
// interrupts and bash mode; no transcript content and no live store is read.

import (
	"context"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

const interruptRoot = "00000000-0000-4000-8000-0000000000b1"

func interruptRow(t *testing.T, lines ...string) domain.InvocationFact {
	t.Helper()
	body := strings.Join(lines, "\n") + "\n"
	transcript, err := ParseTranscript(context.Background(), strings.NewReader(body), TranscriptExpectation{RootSessionID: interruptRoot})
	if err != nil {
		t.Fatal(err)
	}
	extraction, err := Tree{RootID: interruptRoot, SelectedID: interruptRoot, Root: transcript}.Prove().Facts(time.Date(2026, 11, 24, 11, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(extraction.RootIncomplete) > 0 || len(extraction.Sessions) != 1 || len(extraction.Sessions[0].Invocations) != 1 {
		t.Fatalf("want one complete root row, got incomplete=%v sessions=%d", extraction.RootIncomplete, len(extraction.Sessions))
	}
	return extraction.Sessions[0].Invocations[0]
}

func interruptPrompt(at string) string {
	return `{"type":"user","sessionId":"` + interruptRoot + `","uuid":"p-` + at + `","timestamp":"` + at + `","cwd":"/synthetic","promptSource":"sdk","message":{"role":"user","content":"synthetic"}}`
}

func interruptResponse(at, request, stop, blocks string) string {
	return `{"type":"assistant","sessionId":"` + interruptRoot + `","requestId":"` + request + `","timestamp":"` + at + `","message":{"id":"m-` + request + `","model":"model-a","stop_reason":"` + stop + `","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":1},"content":` + blocks + `}}`
}

func harnessUser(at, content string) string {
	return `{"type":"user","sessionId":"` + interruptRoot + `","timestamp":"` + at + `","message":{"role":"user","content":` + content + `}}`
}

func assertClosedRow(t *testing.T, name string, row domain.InvocationFact, active, turns, steering int64) {
	t.Helper()
	if row.ActiveDurationSeconds != active || row.ProvisionalSeconds != 0 || row.CompletedAt == nil || row.Turns != turns || row.UserIterations != steering {
		t.Fatalf("%s: want active=%d turns=%d steering=%d with no open turn, got active=%d provisional=%d completed=%v turns=%d steering=%d",
			name, active, turns, steering, row.ActiveDurationSeconds, row.ProvisionalSeconds, row.CompletedAt != nil, row.Turns, row.UserIterations)
	}
}

func TestSyntheticInterruptEndsTurnWithoutSteering(t *testing.T) {
	row := interruptRow(t,
		interruptPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:10Z", "r1", "tool_use", `[{"type":"tool_use","id":"t1","name":"Bash","input":{}}]`),
		harnessUser("2026-11-24T10:00:20Z", `[{"type":"text","text":"[Request interrupted by user]"}]`),
	)
	assertClosedRow(t, "interrupt between responses", row, 20, 1, 1)

	row = interruptRow(t,
		interruptPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:05Z", "r1", "tool_use", `[{"type":"tool_use","id":"t1","name":"Bash","input":{}}]`),
		harnessUser("2026-11-24T10:00:30Z", `[{"type":"tool_result","tool_use_id":"t1","is_error":true,"content":"synthetic"},{"type":"text","text":"[Request interrupted by user for tool use]"}]`),
	)
	assertClosedRow(t, "interrupt during tool use", row, 30, 1, 1)

	row = interruptRow(t,
		interruptPrompt("2026-11-24T10:00:00Z"),
		harnessUser("2026-11-24T10:00:07Z", `"[Request interrupted by user]"`),
	)
	assertClosedRow(t, "interrupt before any response", row, 7, 0, 1)
}

func TestSyntheticBashModeIsNeitherTurnNorSteering(t *testing.T) {
	row := interruptRow(t,
		interruptPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:40Z", "r1", "end_turn", `[{"type":"text","text":"synthetic"}]`),
		harnessUser("2026-11-24T10:05:00Z", `"<bash-input>synthetic</bash-input>"`),
		harnessUser("2026-11-24T10:05:01Z", `"<bash-stdout>synthetic</bash-stdout><bash-stderr></bash-stderr>"`),
	)
	assertClosedRow(t, "bash mode after a completed turn", row, 40, 1, 1)
}
