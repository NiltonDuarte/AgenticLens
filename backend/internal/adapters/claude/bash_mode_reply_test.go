package claude

// Synthetic regression fixture for AL-98 native defect F2. It mirrors the
// record keys Claude Code 2.1.331 writes for an interactive "!" command: a
// <bash-input> record without promptSource, then a <bash-stdout> record that
// carries turnPosition and turnOrigin "human" but no promptSource, then a model
// reply. It holds no transcript content and reads no live store. Helpers come
// from interrupt_records_test.go.

import "testing"

func typedPrompt(at string) string {
	return `{"type":"user","sessionId":"` + interruptRoot + `","uuid":"p-` + at + `","timestamp":"` + at + `","cwd":"/synthetic","promptSource":"typed","turnOrigin":"human","turnPosition":{"promptIndex":0,"turnIndex":1},"message":{"role":"user","content":"synthetic"}}`
}

func bashOutput(at string) string {
	return `{"type":"user","sessionId":"` + interruptRoot + `","timestamp":"` + at + `","turnOrigin":"human","turnPosition":{"promptIndex":0,"turnIndex":2},"message":{"role":"user","content":"<bash-stdout>synthetic</bash-stdout><bash-stderr></bash-stderr>"}}`
}

func TestSyntheticAnsweredBashModeIsANonSteeringTurn(t *testing.T) {
	row := interruptRow(t,
		typedPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:10Z", "r1", "end_turn", `[{"type":"text","text":"synthetic"}]`),
		harnessUser("2026-11-24T10:01:00Z", `"<bash-input>synthetic</bash-input>"`),
		bashOutput("2026-11-24T10:01:00.300Z"),
		interruptResponse("2026-11-24T10:01:01.500Z", "r2", "end_turn", `[{"type":"thinking","thinking":""}]`),
		interruptResponse("2026-11-24T10:01:02Z", "r2", "end_turn", `[{"type":"text","text":"synthetic"}]`),
	)
	// 10 s for the typed turn plus 2 s from the bash input to the reply's last
	// terminal record; the bash exchange is a turn but not steering.
	assertClosedRow(t, "answered bash command", row, 12, 2, 1)
}

func TestSyntheticUnansweredBashModeBeforeNextPromptIsNoTurn(t *testing.T) {
	row := interruptRow(t,
		typedPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:10Z", "r1", "end_turn", `[{"type":"text","text":"synthetic"}]`),
		harnessUser("2026-11-24T10:01:00Z", `"<bash-input>synthetic</bash-input>"`),
		bashOutput("2026-11-24T10:01:01Z"),
		typedPrompt("2026-11-24T10:02:00Z"),
		interruptResponse("2026-11-24T10:02:05Z", "r2", "end_turn", `[{"type":"text","text":"synthetic"}]`),
	)
	assertClosedRow(t, "unanswered bash command", row, 15, 2, 2)

	row = interruptRow(t,
		typedPrompt("2026-11-24T10:00:00Z"),
		interruptResponse("2026-11-24T10:00:10Z", "r1", "end_turn", `[{"type":"text","text":"synthetic"}]`),
		harnessUser("2026-11-24T10:01:00Z", `"<bash-input>synthetic</bash-input>"`),
		bashOutput("2026-11-24T10:01:01Z"),
	)
	assertClosedRow(t, "bash command ending the transcript", row, 10, 1, 1)
}
