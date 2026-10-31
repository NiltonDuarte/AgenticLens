package domain

import "math"

// CostState identifies whether a fact reports a complete or partial USD amount.
type CostState string

const (
	CostObserved    CostState = "observed"
	CostPartial     CostState = "partial"
	CostUnavailable CostState = "unavailable"
)

// CostEvidence retains reported contributions when an invocation combines records.
type CostEvidence struct {
	State            CostState `json:"state"`
	KnownSubtotalUSD *float64  `json:"known_subtotal_usd"`
}

// CostSummary counts invocation facts; coverage is not a proportion of spend.
type CostSummaryState string

const (
	CostSummaryEmpty       CostSummaryState = "empty"
	CostSummaryComplete    CostSummaryState = "complete"
	CostSummaryPartial     CostSummaryState = "partial"
	CostSummaryUnavailable CostSummaryState = "unavailable"
)

type CostSummary struct {
	State            CostSummaryState `json:"state"`
	KnownSubtotalUSD *float64         `json:"known_subtotal_usd"`
	FullyObserved    int64            `json:"fully_observed"`
	Partial          int64            `json:"partial"`
	Unavailable      int64            `json:"unavailable"`
	Total            int64            `json:"total"`
}

func CostNumber(amount float64) *float64 { value := amount; return &value }

func ObservedCost(amount float64) CostEvidence {
	return CostEvidence{State: CostObserved, KnownSubtotalUSD: CostNumber(amount)}
}
func UnavailableCost() CostEvidence { return CostEvidence{State: CostUnavailable} }
func PartialCost(amount float64) CostEvidence {
	return CostEvidence{State: CostPartial, KnownSubtotalUSD: CostNumber(amount)}
}

// EffectiveCost accepts existing in-memory numeric facts as observed. Source
// extractors and presence-aware boundary decoders explicitly mark missing cost.
func EffectiveCost(amount float64, evidence CostEvidence) CostEvidence {
	if evidence.State == "" {
		return ObservedCost(amount)
	}
	return evidence
}

func ValidCostAmount(amount float64) bool {
	return !math.IsNaN(amount) && !math.IsInf(amount, 0) && amount >= 0
}

func SummarizeCost(invocations []InvocationFact) CostSummary {
	result := CostSummary{}
	for _, inv := range invocations {
		result = MergeCostSummaries(result, SummaryForCost(inv.CostUSD, inv.CostEvidence))
	}
	return FinalizeCostSummary(result)
}

func SummaryForCost(amount float64, evidence CostEvidence) CostSummary {
	evidence = EffectiveCost(amount, evidence)
	result := CostSummary{Total: 1}
	switch evidence.State {
	case CostObserved:
		result.FullyObserved = 1
		result.KnownSubtotalUSD = CostNumber(amount)
	case CostPartial:
		result.Partial = 1
		result.KnownSubtotalUSD = evidence.KnownSubtotalUSD
	default:
		result.Unavailable = 1
	}
	return FinalizeCostSummary(result)
}

func MergeCostSummaries(a, b CostSummary) CostSummary {
	result := CostSummary{FullyObserved: a.FullyObserved + b.FullyObserved, Partial: a.Partial + b.Partial, Unavailable: a.Unavailable + b.Unavailable, Total: a.Total + b.Total}
	if a.KnownSubtotalUSD != nil || b.KnownSubtotalUSD != nil {
		subtotal := 0.0
		if a.KnownSubtotalUSD != nil {
			subtotal += *a.KnownSubtotalUSD
		}
		if b.KnownSubtotalUSD != nil {
			subtotal += *b.KnownSubtotalUSD
		}
		result.KnownSubtotalUSD = CostNumber(subtotal)
	}
	return FinalizeCostSummary(result)
}

func FinalizeCostSummary(summary CostSummary) CostSummary {
	switch {
	case summary.Total == 0:
		summary.State = CostSummaryEmpty
	case summary.FullyObserved == summary.Total:
		summary.State = CostSummaryComplete
	case summary.KnownSubtotalUSD != nil:
		summary.State = CostSummaryPartial
	default:
		summary.State = CostSummaryUnavailable
	}
	return summary
}

func (summary CostSummary) CompleteAmount() *float64 {
	summary = FinalizeCostSummary(summary)
	if summary.State != CostSummaryComplete {
		return nil
	}
	return summary.KnownSubtotalUSD
}

func (summary CostSummary) Subtotal() float64 {
	if summary.KnownSubtotalUSD == nil {
		return 0
	}
	return *summary.KnownSubtotalUSD
}

// EffectiveCostSummary supports numeric metrics constructed by older callers.
func EffectiveCostSummary(summary CostSummary, amount float64, count int64) CostSummary {
	if summary.State != "" {
		return summary
	}
	if count == 0 && amount == 0 {
		return FinalizeCostSummary(summary)
	}
	if count < 1 {
		count = 1
	}
	return FinalizeCostSummary(CostSummary{FullyObserved: count, Total: count, KnownSubtotalUSD: CostNumber(amount)})
}
