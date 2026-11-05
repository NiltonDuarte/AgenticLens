package domain

import (
	"fmt"
	"math"
)

// MeasuredDuration rounds once at the invocation boundary. The provisional
// portion is included in total, with rounding remainder assigned to it.
func MeasuredDuration(completedMS, provisionalMS int64) (total, provisional int64) {
	completed := int64(math.Round(float64(completedMS) / 1000))
	total = int64(math.Round(float64(completedMS+provisionalMS) / 1000))
	return total, total - completed
}

// ValidateDuration checks recorded numeric duration without inferring whether
// activity completed from a zero or absent provisional contribution.
func ValidateDuration(total, provisional int64) error {
	const maxExactSeconds int64 = 9007199254740991
	if total < 0 || total > maxExactSeconds || provisional < 0 || provisional > total {
		return fmt.Errorf("duration requires nonnegative seconds with provisional_seconds no greater than active_duration_seconds")
	}
	return nil
}

// DurationDescription describes recorded numeric facts without claiming source
// completion or process liveness.
func DurationDescription(provisional int64) string {
	if provisional > 0 {
		return fmt.Sprintf("recorded duration including %ds provisional", provisional)
	}
	return "recorded duration"
}

// DurationInterval separates measured and source-established open activity at
// an explicit cutoff. Missing completion alone does not create an interval.
func DurationInterval(start int64, end *int64, cutoff int64, ongoing bool) (completedMS, provisionalMS int64, provisional bool, err error) {
	if start <= 0 {
		return 0, 0, false, fmt.Errorf("missing or invalid activity start")
	}
	if start > cutoff {
		return 0, 0, false, nil
	}
	if end != nil {
		if *end < start {
			return 0, 0, false, fmt.Errorf("completion precedes activity start")
		}
		if *end <= cutoff {
			return *end - start, 0, false, nil
		}
		return 0, cutoff - start, true, nil
	}
	if ongoing {
		return 0, cutoff - start, true, nil
	}
	return 0, 0, false, nil
}
