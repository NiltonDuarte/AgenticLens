package claude

// Synthetic regression fixture for AL-98 defect F1. It reproduces only the
// record keys and envelope markers Claude Code 2.1.331 writes for a local
// slash command after a resumed /compact; it holds no transcript content and
// reads no live store.

import (
	"context"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

const localCommandRoot = "00000000-0000-4000-8000-0000000000f1"

func localCommandTranscript(tail string) string {
	lines := []string{
		`{"type":"user","sessionId":"` + localCommandRoot + `","uuid":"p1","timestamp":"2026-11-24T10:00:00Z","cwd":"/synthetic","promptSource":"sdk","turnOrigin":"sdk","turnPosition":{"promptIndex":0,"turnIndex":1},"message":{"role":"user","content":"synthetic"}}`,
		`{"type":"assistant","sessionId":"` + localCommandRoot + `","requestId":"r1","timestamp":"2026-11-24T10:01:00Z","message":{"id":"m1","model":"model-a","stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":1},"content":[{"type":"text","text":"synthetic"}]}}`,
		`{"type":"system","subtype":"compact_boundary","sessionId":"` + localCommandRoot + `","timestamp":"2026-11-24T10:05:00Z","slug":"s"}`,
		`{"type":"user","sessionId":"` + localCommandRoot + `","timestamp":"2026-11-24T10:05:01Z","isCompactSummary":true,"isVisibleInTranscriptOnly":true,"turnPosition":{"promptIndex":0,"turnIndex":2},"slug":"s","message":{"role":"user","content":"synthetic summary"}}`,
		`{"type":"user","sessionId":"` + localCommandRoot + `","timestamp":"2026-11-24T10:05:02Z","isMeta":true,"rowSource":"harness","slug":"s","message":{"role":"user","content":"<local-command-caveat>synthetic</local-command-caveat>"}}`,
		`{"type":"user","sessionId":"` + localCommandRoot + `","uuid":"c1","timestamp":"2026-11-24T10:05:03Z","rowSource":"local","slug":"s","message":{"role":"user","content":"<command-name>/compact</command-name>"}}`,
	}
	return strings.Join(lines, "\n") + "\n" + tail
}

func localCommandRow(t *testing.T, tail string) domain.InvocationFact {
	t.Helper()
	transcript, err := ParseTranscript(context.Background(), strings.NewReader(localCommandTranscript(tail)), TranscriptExpectation{RootSessionID: localCommandRoot})
	if err != nil {
		t.Fatal(err)
	}
	extraction, err := Tree{RootID: localCommandRoot, SelectedID: localCommandRoot, Root: transcript}.Prove().Facts(time.Date(2026, 11, 24, 11, 0, 0, 0, time.UTC), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(extraction.RootIncomplete) > 0 || len(extraction.Sessions) != 1 || len(extraction.Sessions[0].Invocations) != 1 {
		t.Fatalf("want one complete root row, got incomplete=%v sessions=%d", extraction.RootIncomplete, len(extraction.Sessions))
	}
	return extraction.Sessions[0].Invocations[0]
}

func TestSyntheticLocalSlashCommandIsNeitherTurnNorSteering(t *testing.T) {
	output := `{"type":"user","sessionId":"` + localCommandRoot + `","timestamp":"2026-11-24T10:05:04Z","rowSource":"local","slug":"s","message":{"role":"user","content":"<local-command-stdout>synthetic</local-command-stdout>"}}` + "\n"
	row := localCommandRow(t, output)
	if row.UserIterations != 1 || row.Turns != 1 || row.Compactions != 1 {
		t.Fatalf("local command echo and output must not be steering or turns: iterations=%d turns=%d compactions=%d", row.UserIterations, row.Turns, row.Compactions)
	}
	if row.ActiveDurationSeconds != 60 || row.ProvisionalSeconds != 0 || row.CompletedAt == nil {
		t.Fatalf("a session ending with a local command has no open turn: active=%d provisional=%d completed=%v", row.ActiveDurationSeconds, row.ProvisionalSeconds, row.CompletedAt)
	}

	// A prompt-type slash command opens with the same echo envelope but gets a
	// model reply, so it remains a measured turn.
	reply := `{"type":"assistant","sessionId":"` + localCommandRoot + `","requestId":"r2","timestamp":"2026-11-24T10:05:33Z","message":{"id":"m2","model":"model-a","stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1,"cache_read_input_tokens":1},"content":[{"type":"text","text":"synthetic"}]}}` + "\n"
	row = localCommandRow(t, reply)
	if row.Turns != 2 || row.UserIterations != 2 || row.ActiveDurationSeconds != 90 || row.ProvisionalSeconds != 0 {
		t.Fatalf("an answered slash command stays a turn: turns=%d iterations=%d active=%d provisional=%d", row.Turns, row.UserIterations, row.ActiveDurationSeconds, row.ProvisionalSeconds)
	}
}
