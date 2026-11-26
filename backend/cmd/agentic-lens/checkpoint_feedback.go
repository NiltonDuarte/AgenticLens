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
	const refreshOptions = "supported automatic capture can refresh the ledger. Otherwise, run generate outside the completed turn with --source and --session-id for the exact extraction root."
	const refreshGuidance = "After natural completion and source availability, " + refreshOptions
	if item.Case == domain.CheckpointUnsupportedContinuation {
		return fmt.Sprintf("AgenticLens checkpoint: automatic next-turn continuation is unsupported or unverified for %s. Telemetry can be refreshed without starting another turn.", checkpointLabel(string(item.CallerSource)))
	}
	prefix := fmt.Sprintf("AgenticLens checkpoint: source=%s; session=%s; invocation=%d; role=%s; model=%s; recorded provisional subtotal=%ds", checkpointLabel(string(item.Source)), checkpointLabel(item.SessionID), item.InvocationRow, checkpointLabel(string(item.Role)), checkpointLabel(item.Model), item.ProvisionalSeconds)
	kind, id, started := checkpointLabel(item.Interval.Kind), checkpointLabel(item.Interval.ID), item.Interval.StartedAt.UTC().Format(time.RFC3339Nano)
	switch item.Case {
	case domain.CheckpointCurrentCaller:
		return fmt.Sprintf("%s. Your verified current %s %s, started at %s, has reached %ds (threshold %ds). Once your work is complete, finish this turn normally. %s", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds, refreshGuidance)
	case domain.CheckpointOtherCaller:
		return fmt.Sprintf("%s. Source reports open %s %s, started at %s, at %ds (threshold %ds). This belongs to another session; verified caller=%s/%s. %s", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds, checkpointLabel(string(item.CallerSource)), checkpointLabel(item.CallerSessionID), refreshGuidance)
	case domain.CheckpointUnknownCaller:
		return fmt.Sprintf("%s. Source reports open %s %s, started at %s, at %ds (threshold %ds). The invoking caller is unverified; this interval may belong to another agent. %s", prefix, kind, id, started, item.ElapsedSeconds, item.ThresholdSeconds, refreshGuidance)
	case domain.CheckpointUnavailableInterval:
		return fmt.Sprintf("%s (threshold %ds). This subtotal may cover multiple open intervals or retained observations. Current native interval evidence is unavailable or unverified. The subtotal does not identify a single long turn or the current caller. Preserve the recorded facts. Once native completion records are readable, %s", prefix, item.ThresholdSeconds, refreshOptions)
	default:
		return ""
	}
}
