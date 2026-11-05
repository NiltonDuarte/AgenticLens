package main

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"agenticlens/internal/domain"
)

// writeCheckpointFeedback is best-effort human output, called only after the
// requested generation publication has succeeded. No evidence is persisted.
func writeCheckpointFeedback(writer io.Writer, feedback []domain.CheckpointFeedback) error {
	for _, item := range feedback {
		line := renderCheckpointFeedback(item)
		if line == "" {
			continue
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return fmt.Errorf("writing optional checkpoint feedback: %w", err)
		}
	}
	return nil
}

func checkpointLabel(value string) string {
	if value == "" {
		return "unknown"
	}
	quoted := strconv.QuoteToASCII(value)
	return quoted[1 : len(quoted)-1]
}

func renderCheckpointFeedback(item domain.CheckpointFeedback) string {
	if item.Case == domain.CheckpointUnsupportedContinuation {
		return fmt.Sprintf("AgenticLens checkpoint: automatic next-turn continuation is unsupported or unverified for %s. This command will not end the turn, schedule another turn or resume itself. After normal turn completion, a user or supported harness action must start the next turn before telemetry is refreshed.", checkpointLabel(string(item.CallerSource)))
	}
	prefix := fmt.Sprintf("AgenticLens checkpoint: source=%s; session=%s; invocation=%d; role=%s; model=%s; recorded provisional subtotal=%ds", checkpointLabel(string(item.Source)), checkpointLabel(item.SessionID), item.InvocationRow, checkpointLabel(string(item.Role)), checkpointLabel(item.Model), item.ProvisionalSeconds)
	kind, id, started := checkpointLabel(item.Interval.Kind), checkpointLabel(item.Interval.ID), item.Interval.StartedAt.UTC().Format(time.RFC3339Nano)
	switch item.Case {
	case domain.CheckpointCurrentCaller:
		return fmt.Sprintf("%s. Your verified current %s %s, started at %s, has reached %ds (threshold %ds). At a natural task boundary, complete this turn normally, then refresh telemetry in the next turn. Generating again in this same turn does not close the interval.", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds)
	case domain.CheckpointOtherCaller:
		return fmt.Sprintf("%s. Source reports open %s %s, started at %s, at %ds (threshold %ds). This belongs to another session; verified caller=%s/%s. Let that owner reach a natural completion boundary, then refresh its source. Do not force another agent to complete for telemetry.", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds, checkpointLabel(string(item.CallerSource)), checkpointLabel(item.CallerSessionID))
	case domain.CheckpointUnknownCaller:
		return fmt.Sprintf("%s. Source reports open %s %s, started at %s, at %ds (threshold %ds). The invoking caller is unverified; this interval may belong to another agent. Refresh after the owning session completes naturally; no current-caller rollover is established.", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds)
	case domain.CheckpointUnavailableInterval:
		return fmt.Sprintf("%s (threshold %ds). This subtotal may cover multiple open intervals or retained observations. Current native interval evidence is unavailable or unverified; no single long turn or current-caller rollover is established. Preserve these recorded facts and refresh when the owning source is available.", prefix, item.ThresholdSeconds)
	default:
		return ""
	}
}
