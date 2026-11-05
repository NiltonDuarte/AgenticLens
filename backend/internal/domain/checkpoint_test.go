package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestCheckpointExactAgeAndQualifiedFallback(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC)
	long := CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "native", StartedAt: cutoff.Add(-180 * time.Second)}
	session := SessionFact{ID: "owner", Source: SourceCodex, Invocations: []InvocationFact{{Model: "m", Role: "coder", ActiveDurationSeconds: 180, ProvisionalSeconds: 180}}}
	complete := CheckpointInvocationEvidence{Source: SourceCodex, SessionID: "owner", InvocationRow: 1, Model: "m", Cutoff: cutoff, Complete: true, Intervals: []CheckpointInterval{long}}
	current := CheckpointCaller{Verified: true, Source: SourceCodex, SessionID: "owner", CurrentInterval: &long}
	for _, tc := range []struct {
		name      string
		threshold int64
		mutate    func(*CheckpointInvocationEvidence, *CheckpointCaller)
		want      []CheckpointCase
	}{
		{"inclusive_exact_180000ms", 180, nil, []CheckpointCase{CheckpointCurrentCaller, CheckpointUnsupportedContinuation}},
		{"rounded_179999ms_is_not_missing_proof", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Intervals[0].StartedAt = cutoff.Add(-179999 * time.Millisecond)
			c.CurrentInterval = &e.Intervals[0]
		}, nil},
		{"disabled", 0, nil, nil},
		{"custom_threshold", 181, nil, nil},
		{"custom_inclusive", 90, nil, []CheckpointCase{CheckpointCurrentCaller, CheckpointUnsupportedContinuation}},
		{"unverified_caller_cannot_claim_current", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { c.Verified = false }, []CheckpointCase{CheckpointUnknownCaller}},
		{"different_owner", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { c.SessionID = "caller" }, []CheckpointCase{CheckpointOtherCaller}},
		{"same_id_different_source", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { c.Source = SourceOpenCode }, []CheckpointCase{CheckpointOtherCaller}},
		{"current_native_interval_unmatched", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			c.CurrentInterval = &CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "different", StartedAt: long.StartedAt}
		}, []CheckpointCase{CheckpointUnavailableInterval}},
		{"same_start_in_other_timezone", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			current := *c.CurrentInterval
			current.StartedAt = current.StartedAt.In(time.FixedZone("offset", 3600))
			c.CurrentInterval = &current
		}, []CheckpointCase{CheckpointCurrentCaller, CheckpointUnsupportedContinuation}},
		{"duplicate_native_interval_once", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Intervals = append(e.Intervals, e.Intervals[0])
		}, []CheckpointCase{CheckpointCurrentCaller, CheckpointUnsupportedContinuation}},
		{"conflicting_native_identity_qualified_only", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			other := e.Intervals[0]
			other.StartedAt = other.StartedAt.Add(-time.Second)
			e.Intervals = append(e.Intervals, other)
		}, []CheckpointCase{CheckpointUnavailableInterval}},
		{"historical_known_long_is_qualified", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { e.Intervals[0].SourceOpen = false }, []CheckpointCase{CheckpointUnavailableInterval}},
		{"historical_known_short_remains_suppressed", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Intervals[0].SourceOpen = false
			e.Intervals[0].StartedAt = cutoff.Add(-179999 * time.Millisecond)
		}, nil},
		{"source_evidence_stale", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { e.Cutoff = cutoff.Add(-time.Millisecond) }, []CheckpointCase{CheckpointUnavailableInterval}},
		{"model_attribution_mismatch", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { e.Model = "other" }, []CheckpointCase{CheckpointUnavailableInterval}},
		{"future_interval", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Intervals[0].StartedAt = cutoff.Add(time.Millisecond)
		}, []CheckpointCase{CheckpointUnavailableInterval}},
		{"missing_native_identity", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { e.Intervals[0].ID = "" }, []CheckpointCase{CheckpointUnavailableInterval}},
		{"supported_continuation_adds_no_success_message", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) { c.ContinuationVerified = true }, []CheckpointCase{CheckpointCurrentCaller}},
		{"known_short_intervals_never_f04", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Intervals = []CheckpointInterval{{SourceOpen: true, Kind: "turn", ID: "a", StartedAt: cutoff.Add(-100 * time.Second)}, {SourceOpen: true, Kind: "turn", ID: "b", StartedAt: cutoff.Add(-80 * time.Second)}}
		}, nil},
		{"unknown_caller_known_short_never_f04", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			c.Verified = false
			e.Intervals = []CheckpointInterval{{SourceOpen: true, Kind: "turn", ID: "a", StartedAt: cutoff.Add(-100 * time.Second)}, {SourceOpen: true, Kind: "turn", ID: "b", StartedAt: cutoff.Add(-80 * time.Second)}}
		}, nil},
		{"partial_short_evidence_retains_f04", 180, func(e *CheckpointInvocationEvidence, c *CheckpointCaller) {
			e.Complete = false
			e.Intervals = []CheckpointInterval{{SourceOpen: true, Kind: "turn", ID: "a", StartedAt: cutoff.Add(-100 * time.Second)}}
		}, []CheckpointCase{CheckpointUnavailableInterval}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := complete
			e.Intervals = append([]CheckpointInterval{}, complete.Intervals...)
			c := current
			if tc.mutate != nil {
				tc.mutate(&e, &c)
			}
			got := SelectCheckpointFeedback([]SessionFact{session}, []CheckpointInvocationEvidence{e}, c, cutoff, tc.threshold)
			var cases []CheckpointCase
			for _, item := range got {
				cases = append(cases, item.Case)
				if item.Case != CheckpointUnsupportedContinuation && (item.SessionID != "owner" || item.ProvisionalSeconds != 180 || item.InvocationRow != 1) {
					t.Fatalf("lost owner/subtotal: %+v", item)
				}
				if item.Case == CheckpointCurrentCaller && item.ElapsedSeconds != 180 {
					t.Fatalf("native elapsed: %+v", item)
				}
			}
			if !reflect.DeepEqual(cases, tc.want) {
				t.Fatalf("cases %v want %v: %+v", cases, tc.want, got)
			}
		})
	}
}

func TestCheckpointRowsOrderingMultiplicityAndNoPooling(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC)
	inv := func(model string, seconds int64) InvocationFact {
		return InvocationFact{Model: model, Role: "coder", ActiveDurationSeconds: seconds, ProvisionalSeconds: seconds}
	}
	sessions := []SessionFact{{ID: "z", Source: SourceOpenCode, Invocations: []InvocationFact{inv("m", 180)}}, {ID: "b", Source: SourceCodex, Invocations: []InvocationFact{inv("z", 180), inv("a", 180), inv("a", 180)}}, {ID: "a", Source: SourceCodex, Invocations: []InvocationFact{inv("m", 100), inv("m", 100)}}}
	got := SelectCheckpointFeedback(sessions, nil, CheckpointCaller{}, cutoff, 180)
	var identities []string
	for _, item := range got {
		identities = append(identities, string(item.Source)+"/"+item.SessionID+"/"+item.Model)
		if item.Case != CheckpointUnavailableInterval {
			t.Fatalf("retained owner overclaimed: %+v", item)
		}
	}
	want := []string{"codex/b/a", "codex/b/a", "codex/b/z", "opencode/z/m"}
	if !reflect.DeepEqual(identities, want) || got[0].InvocationRow != 2 || got[1].InvocationRow != 3 {
		t.Fatalf("ordering/multiplicity: %+v", got)
	}
	if sessions[1].Invocations[0].ProvisionalSeconds != 180 {
		t.Fatal("classifier mutated facts")
	}
}

func TestCheckpointConcurrentIntervalsAndContinuationOnce(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC)
	current := CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "current", StartedAt: cutoff.Add(-180 * time.Second)}
	other := CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "other", StartedAt: cutoff.Add(-200 * time.Second)}
	sessions := []SessionFact{{ID: "caller", Source: SourceCodex, Invocations: []InvocationFact{{Model: "a", Role: "coder", ActiveDurationSeconds: 380, ProvisionalSeconds: 380}, {Model: "b", Role: "coder", ActiveDurationSeconds: 180, ProvisionalSeconds: 180}}}, {ID: "child", Source: SourceCodex, Invocations: []InvocationFact{{Model: "m", Role: "tester", ActiveDurationSeconds: 200, ProvisionalSeconds: 200}}}}
	evidence := []CheckpointInvocationEvidence{{Source: SourceCodex, SessionID: "caller", InvocationRow: 1, Model: "a", Cutoff: cutoff, Complete: true, Intervals: []CheckpointInterval{other, current}}, {Source: SourceCodex, SessionID: "caller", InvocationRow: 2, Model: "b", Cutoff: cutoff, Complete: true, Intervals: []CheckpointInterval{current}}, {Source: SourceCodex, SessionID: "child", InvocationRow: 1, Model: "m", Cutoff: cutoff, Complete: true, Intervals: []CheckpointInterval{other}}}
	got := SelectCheckpointFeedback(sessions, evidence, CheckpointCaller{Verified: true, Source: SourceCodex, SessionID: "caller", CurrentInterval: &current}, cutoff, 180)
	counts := map[CheckpointCase]int{}
	for _, item := range got {
		counts[item.Case]++
	}
	if counts[CheckpointCurrentCaller] != 2 || counts[CheckpointOtherCaller] != 1 || counts[CheckpointUnavailableInterval] != 1 || counts[CheckpointUnsupportedContinuation] != 1 || got[len(got)-1].Case != CheckpointUnsupportedContinuation {
		t.Fatalf("concurrent ownership/continuation: %+v", got)
	}
}

func TestCheckpointAmbiguousEvidenceCannotCertifyAnInterval(t *testing.T) {
	cutoff := time.Date(2026, 1, 1, 0, 10, 0, 0, time.UTC)
	sessions := []SessionFact{{ID: "owner", Source: SourceCodex, Invocations: []InvocationFact{{Model: "m", ActiveDurationSeconds: 180, ProvisionalSeconds: 180}}}}
	interval := CheckpointInterval{SourceOpen: true, Kind: "turn", ID: "native", StartedAt: cutoff.Add(-180 * time.Second)}
	e := CheckpointInvocationEvidence{Source: SourceCodex, SessionID: "owner", InvocationRow: 1, Model: "m", Cutoff: cutoff, Complete: true, Intervals: []CheckpointInterval{interval}}
	got := SelectCheckpointFeedback(sessions, []CheckpointInvocationEvidence{e, e}, CheckpointCaller{}, cutoff, 180)
	if len(got) != 1 || got[0].Case != CheckpointUnavailableInterval {
		t.Fatalf("ambiguous owner/row evidence certified: %+v", got)
	}
}
