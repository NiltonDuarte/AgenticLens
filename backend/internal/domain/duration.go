package domain

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type DurationState string

const (
	DurationComplete    DurationState = "complete"
	DurationProvisional DurationState = "provisional"
	DurationPartial     DurationState = "partial"
)

// DurationEvidence describes the observed coverage of one invocation. Absence
// on a legacy invocation means unknown coverage, never a measured zero.
type DurationEvidence struct {
	State              DurationState `json:"state"`
	CompletedSeconds   int64         `json:"completed_seconds"`
	ProvisionalSeconds int64         `json:"provisional_seconds"`
	Cutoff             string        `json:"cutoff,omitempty"`
	Basis              string        `json:"basis"`
	Limitation         string        `json:"limitation,omitempty"`
}

// DurationSummary keeps distinct stored cutoffs rather than inventing a common
// observation time for facts captured on different machines or operations.
type DurationSummary struct {
	ProvisionalSeconds     int64    `json:"provisional_seconds"`
	ProvisionalInvocations int64    `json:"provisional_invocations"`
	UnknownInvocations     int64    `json:"unknown_invocations"`
	Cutoffs                []string `json:"cutoffs"`
}

// MeasuredDuration rounds once at the existing invocation boundary. Assign the
// rounding remainder to provisional time so components equal the numeric total.
func MeasuredDuration(completedMS, provisionalMS int64, provisional bool, cutoff, basis, limitation string) (int64, *DurationEvidence) {
	completed := int64(math.Round(float64(completedMS) / 1000))
	total := int64(math.Round(float64(completedMS+provisionalMS) / 1000))
	state := DurationComplete
	if provisional {
		state = DurationProvisional
	} else {
		cutoff = ""
	}
	if limitation != "" {
		state = DurationPartial
	}
	return total, &DurationEvidence{State: state, CompletedSeconds: completed, ProvisionalSeconds: total - completed, Cutoff: cutoff, Basis: basis, Limitation: limitation}
}

func ValidateDuration(total int64, evidence *DurationEvidence) error {
	if evidence == nil {
		return nil
	}
	if total < 0 || evidence.CompletedSeconds < 0 || evidence.ProvisionalSeconds < 0 || evidence.CompletedSeconds > total || evidence.ProvisionalSeconds != total-evidence.CompletedSeconds || evidence.Basis == "" {
		return fmt.Errorf("duration evidence must have a basis and nonnegative components equal to active duration")
	}
	switch evidence.State {
	case "complete":
		if evidence.Cutoff != "" || evidence.ProvisionalSeconds != 0 || evidence.Limitation != "" {
			return fmt.Errorf("complete duration cannot carry provisional time or limitations")
		}
	case "provisional":
		if evidence.Cutoff == "" || evidence.Limitation != "" {
			return fmt.Errorf("provisional duration requires a cutoff and no unsupported coverage")
		}
	case "partial":
		if evidence.Limitation == "" {
			return fmt.Errorf("partial duration requires a limitation")
		}
	default:
		return fmt.Errorf("unsupported duration coverage %q", evidence.State)
	}
	if evidence.Cutoff != "" {
		if _, err := time.Parse(time.RFC3339Nano, evidence.Cutoff); err != nil {
			return fmt.Errorf("invalid duration cutoff: %w", err)
		}
	}
	if evidence.ProvisionalSeconds > 0 && evidence.Cutoff == "" {
		return fmt.Errorf("provisional seconds require a cutoff")
	}
	return nil
}

func DurationForInvocation(inv InvocationFact) DurationSummary {
	result := DurationSummary{Cutoffs: []string{}}
	e := inv.DurationEvidence
	if e == nil {
		result.UnknownInvocations = 1
		return result
	}
	result.ProvisionalSeconds = e.ProvisionalSeconds
	if e.Cutoff != "" {
		result.ProvisionalInvocations = 1
		result.Cutoffs = append(result.Cutoffs, e.Cutoff)
	}
	if e.State == "partial" {
		result.UnknownInvocations = 1
	}
	return result
}

func MergeDuration(a, b DurationSummary) DurationSummary {
	result := DurationSummary{ProvisionalSeconds: a.ProvisionalSeconds + b.ProvisionalSeconds, ProvisionalInvocations: a.ProvisionalInvocations + b.ProvisionalInvocations, UnknownInvocations: a.UnknownInvocations + b.UnknownInvocations, Cutoffs: []string{}}
	unique := map[string]bool{}
	for _, values := range [][]string{a.Cutoffs, b.Cutoffs} {
		for _, v := range values {
			unique[v] = true
		}
	}
	for v := range unique {
		result.Cutoffs = append(result.Cutoffs, v)
	}
	sort.Strings(result.Cutoffs)
	return result
}

// DurationDescription is shared by CLI summaries; cutoffs are recorded facts.
func DurationDescription(s DurationSummary) string {
	result := "completed measurements"
	if s.UnknownInvocations > 0 {
		result = "recorded duration"
	}
	if s.ProvisionalInvocations > 0 {
		result = fmt.Sprintf("includes %ds provisional across %d invocation(s); snapshot cutoffs: %v", s.ProvisionalSeconds, s.ProvisionalInvocations, s.Cutoffs)
	}
	if s.UnknownInvocations > 0 {
		result += fmt.Sprintf("; coverage unknown/unsupported for %d invocation(s)", s.UnknownInvocations)
	}
	return result
}

// DurationInterval separates completed measurements from a proven unfinished
// interval at cutoff. An absent end alone never establishes current activity.
func DurationInterval(start int64, end *int64, cutoff int64, ongoing bool) (completedMS, provisionalMS int64, provisional bool, limitation string) {
	if start <= 0 {
		return 0, 0, false, "missing or invalid activity start"
	}
	if start > cutoff {
		return 0, 0, false, ""
	}
	if end != nil {
		if *end < start {
			return 0, 0, false, "completion precedes activity start"
		}
		if *end <= cutoff {
			return *end - start, 0, false, ""
		}
		return 0, cutoff - start, true, ""
	}
	if ongoing {
		return 0, cutoff - start, true, ""
	}
	return 0, 0, false, "unfinished activity lacks reliable current execution evidence"
}
