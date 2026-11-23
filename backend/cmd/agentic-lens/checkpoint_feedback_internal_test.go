package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/domain"
)

func TestCheckpointExactFiveMessageCopiesAndEscaping(t *testing.T) {
	base := domain.CheckpointFeedback{Source: domain.SourceCodex, SessionID: "owner", InvocationRow: 2, Role: "coder", Model: "model", ProvisionalSeconds: 200, ThresholdSeconds: 180, Interval: domain.CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "native", StartedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, ElapsedSeconds: 180, CallerSource: domain.SourceCodex, CallerSessionID: "caller"}
	prefix := "AgenticLens checkpoint: source=codex; session=owner; invocation=2; role=coder; model=model; recorded provisional subtotal=200s"
	refresh := " After natural completion and source availability, supported automatic capture can refresh the ledger. Otherwise, run generate outside the completed turn with --source and --session-id for the exact extraction root."
	for _, tc := range []struct {
		kind domain.CheckpointCase
		want string
	}{
		{domain.CheckpointCurrentCaller, prefix + ". Your verified current turn native, started at 2026-01-01T00:00:00Z, has reached 180s (threshold 180s). Once your work is complete, finish this turn normally." + refresh},
		{domain.CheckpointOtherCaller, prefix + ". Source reports open turn native, started at 2026-01-01T00:00:00Z, at 180s (threshold 180s). This belongs to another session; verified caller=codex/caller." + refresh},
		{domain.CheckpointUnknownCaller, prefix + ". Source reports open turn native, started at 2026-01-01T00:00:00Z, at 180s (threshold 180s). The invoking caller is unverified; this interval may belong to another agent." + refresh},
		{domain.CheckpointUnavailableInterval, prefix + " (threshold 180s). This subtotal may cover multiple open intervals or retained observations. Current native interval evidence is unavailable or unverified. The subtotal does not identify a single long turn or the current caller. Preserve the recorded facts. Once native completion records are readable, supported automatic capture can refresh the ledger. Otherwise, run generate outside the completed turn with --source and --session-id for the exact extraction root."},
		{domain.CheckpointUnsupportedContinuation, "AgenticLens checkpoint: automatic next-turn continuation is unsupported or unverified for codex. Telemetry can be refreshed without starting another turn."},
	} {
		t.Run(string(tc.kind), func(t *testing.T) {
			item := base
			item.Case = tc.kind
			var out bytes.Buffer
			if err := writeCheckpointFeedback(&out, []domain.CheckpointFeedback{item}); err != nil {
				t.Fatal(err)
			}
			if out.String() != tc.want+"\n" {
				t.Fatalf("copy changed:\ngot %q\nwant %q", out.String(), tc.want+"\n")
			}
		})
	}
	escaped := base
	escaped.Case = domain.CheckpointOtherCaller
	escaped.SessionID = "owner\nforged\r\x1b[31m"
	escaped.Model = "model\t\"\\"
	escaped.Role = ""
	escaped.Interval.ID = "native\u2028injected"
	escaped.CallerSessionID = "caller\nforged"
	var out bytes.Buffer
	if err := writeCheckpointFeedback(&out, []domain.CheckpointFeedback{escaped}); err != nil {
		t.Fatal(err)
	}
	line := out.String()
	if strings.Count(line, "\n") != 1 || strings.ContainsAny(strings.TrimSuffix(line, "\n"), "\r\t\x1b\u2028") || !strings.Contains(line, `session=owner\nforged\r\x1b[31m`) || !strings.Contains(line, `role=unknown`) || !strings.Contains(line, `model=model\t\"\\`) {
		t.Fatalf("unsafe/nontruthful labels: %q", line)
	}
}

type checkpointFailWriter struct{}

func (checkpointFailWriter) Write([]byte) (int, error) {
	return 0, errors.New("feedback output unavailable")
}
func TestCheckpointWriterReportsOptionalFailure(t *testing.T) {
	err := writeCheckpointFeedback(checkpointFailWriter{}, []domain.CheckpointFeedback{{Case: domain.CheckpointUnavailableInterval}})
	if err == nil || !strings.Contains(err.Error(), "feedback output unavailable") {
		t.Fatalf("optional write error lost: %v", err)
	}
}
