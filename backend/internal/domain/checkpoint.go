package domain

import (
	"sort"
	"time"
)

// CheckpointInterval retains native interval timing before grouping. SourceOpen
// separately proves current source openness; known historical timing can still
// suppress individually short intervals. There is no ledger projection.
type CheckpointInterval struct {
	SourceOpen bool
	Kind       string
	ID         string
	StartedAt  time.Time
}

// CheckpointInvocationEvidence is captured before native intervals are grouped.
// Complete means native timing/lifecycle coverage for this row is available.
// Missing or retained-only evidence must not be represented as complete/empty.
type CheckpointInvocationEvidence struct {
	Source        SourceHarness
	SessionID     string
	InvocationRow int
	Model         string
	Cutoff        time.Time
	Complete      bool
	Intervals     []CheckpointInterval
}

// CheckpointCaller is supplied only by an independently verified native caller
// transport. Extraction roots, nonces and capture receipts do not establish it.
type CheckpointCaller struct {
	Verified             bool
	Source               SourceHarness
	SessionID            string
	CurrentInterval      *CheckpointInterval
	ContinuationVerified bool
}

type CheckpointCase string

const (
	CheckpointCurrentCaller           CheckpointCase = "F01"
	CheckpointOtherCaller             CheckpointCase = "F02"
	CheckpointUnknownCaller           CheckpointCase = "F03"
	CheckpointUnavailableInterval     CheckpointCase = "F04"
	CheckpointUnsupportedContinuation CheckpointCase = "F05"
)

// CheckpointFeedback contains only renderer parameters, with owner row identity
// preserved even when two invocations have identical role/model labels.
type CheckpointFeedback struct {
	Case               CheckpointCase
	Source             SourceHarness
	SessionID          string
	InvocationRow      int
	Role               AgentRole
	Model              string
	ProvisionalSeconds int64
	ThresholdSeconds   int64
	Interval           CheckpointInterval
	ElapsedSeconds     int64
	CallerSource       SourceHarness
	CallerSessionID    string
}

// SelectCheckpointFeedback is pure: cutoff and all proof are explicit inputs.
// Exact age is compared before invocation-level rounding can influence it.
func SelectCheckpointFeedback(sessions []SessionFact, evidence []CheckpointInvocationEvidence, caller CheckpointCaller, cutoff time.Time, thresholdSeconds int64) []CheckpointFeedback {
	if thresholdSeconds <= 0 {
		return nil
	}
	type rowKey struct {
		source  SourceHarness
		session string
		row     int
	}
	byRow := make(map[rowKey]CheckpointInvocationEvidence, len(evidence))
	ambiguous := make(map[rowKey]bool)
	for _, item := range evidence {
		key := rowKey{item.Source, item.SessionID, item.InvocationRow}
		if _, exists := byRow[key]; exists {
			ambiguous[key] = true
		}
		byRow[key] = item
	}
	var result []CheckpointFeedback
	needsContinuation := false
	for _, session := range sessions {
		for index, invocation := range session.Invocations {
			base := CheckpointFeedback{Source: session.Source, SessionID: session.ID, InvocationRow: index + 1, Role: invocation.Role, Model: invocation.Model, ProvisionalSeconds: invocation.ProvisionalSeconds, ThresholdSeconds: thresholdSeconds, CallerSource: caller.Source, CallerSessionID: caller.SessionID}
			key := rowKey{session.Source, session.ID, index + 1}
			observed, found := byRow[key]
			found = found && !ambiguous[key] && observed.Model == invocation.Model && observed.Cutoff.Equal(cutoff)
			unresolved := !found || !observed.Complete
			if found {
				type nativeKey struct{ kind, id string }
				unique := make(map[nativeKey]CheckpointInterval)
				conflicts := make(map[nativeKey]bool)
				for _, interval := range observed.Intervals {
					key := nativeKey{interval.Kind, interval.ID}
					if previous, exists := unique[key]; exists && !previous.StartedAt.Equal(interval.StartedAt) {
						conflicts[key] = true
					}
					unique[key] = interval
				}
				seen := make(map[nativeKey]bool)
				for _, interval := range observed.Intervals {
					key := nativeKey{interval.Kind, interval.ID}
					if seen[key] {
						continue
					}
					seen[key] = true
					if conflicts[key] {
						unresolved = true
						continue
					}
					if interval.Kind == "" || interval.ID == "" || interval.StartedAt.IsZero() || interval.StartedAt.After(cutoff) {
						unresolved = true
						continue
					}
					// Dividing milliseconds avoids overflow for a large configured threshold
					// and preserves the inclusive threshold at millisecond precision.
					elapsed := cutoff.Sub(interval.StartedAt).Milliseconds() / 1000
					if elapsed < thresholdSeconds {
						continue
					}
					if !interval.SourceOpen {
						unresolved = true
						continue
					}
					item := base
					item.Interval, item.ElapsedSeconds = interval, elapsed
					switch {
					case !caller.Verified || caller.Source == "" || caller.SessionID == "":
						item.Case = CheckpointUnknownCaller
					case caller.Source != session.Source || caller.SessionID != session.ID:
						item.Case = CheckpointOtherCaller
					case caller.CurrentInterval != nil && caller.CurrentInterval.SourceOpen && caller.CurrentInterval.Kind == interval.Kind && caller.CurrentInterval.ID == interval.ID && caller.CurrentInterval.StartedAt.Equal(interval.StartedAt):
						item.Case = CheckpointCurrentCaller
						needsContinuation = needsContinuation || !caller.ContinuationVerified
					default:
						unresolved = true
						continue
					}
					result = append(result, item)
				}
			}
			if unresolved && invocation.ProvisionalSeconds >= thresholdSeconds {
				base.Case = CheckpointUnavailableInterval
				result = append(result, base)
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		if a.SessionID != b.SessionID {
			return a.SessionID < b.SessionID
		}
		if a.Role != b.Role {
			return a.Role < b.Role
		}
		if a.Model != b.Model {
			return a.Model < b.Model
		}
		if a.InvocationRow != b.InvocationRow {
			return a.InvocationRow < b.InvocationRow
		}
		if !a.Interval.StartedAt.Equal(b.Interval.StartedAt) {
			return a.Interval.StartedAt.Before(b.Interval.StartedAt)
		}
		return a.Interval.ID < b.Interval.ID
	})
	if needsContinuation {
		result = append(result, CheckpointFeedback{Case: CheckpointUnsupportedContinuation, CallerSource: caller.Source})
	}
	return result
}
