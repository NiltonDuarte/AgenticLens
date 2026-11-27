package domain

import (
	"fmt"
	"math"
)

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

// EffectiveCost accepts existing in-memory numeric facts as observed, including
// zero. It does not infer wire presence: source extractors and boundary decoders
// explicitly mark missing/null cost before constructing domain facts.
func EffectiveCost(amount float64, evidence CostEvidence) CostEvidence {
	if evidence.State == "" {
		return ObservedCost(amount)
	}
	return evidence
}

func ValidCostAmount(amount float64) bool {
	return !math.IsNaN(amount) && !math.IsInf(amount, 0) && amount >= 0
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

// EffectiveCostSummary supports numeric metrics constructed by older in-memory
// callers. Existing evidence is authoritative; zero amount with zero rows stays
// empty, while a present row with numeric zero is observed. Wire decoding owns
// missing-value semantics independently of this compatibility fallback.
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

// ValidateCostEvidence validates nullable amounts and evidence without I/O.
func ValidateCostEvidence(amount *float64, evidence *CostEvidence, required bool) error {
	if amount != nil && !ValidCostAmount(*amount) {
		return fmt.Errorf("cost_usd must be a finite non-negative number")
	}
	if evidence == nil {
		if required {
			return fmt.Errorf("cost_evidence is required in schema v5")
		}
		return nil
	}
	if evidence.KnownSubtotalUSD != nil && !ValidCostAmount(*evidence.KnownSubtotalUSD) {
		return fmt.Errorf("known_subtotal_usd must be finite and non-negative")
	}
	switch evidence.State {
	case CostObserved:
		if amount == nil || evidence.KnownSubtotalUSD == nil || *amount != *evidence.KnownSubtotalUSD {
			return fmt.Errorf("observed cost requires matching numeric cost_usd and known_subtotal_usd")
		}
	case CostPartial:
		if amount != nil || evidence.KnownSubtotalUSD == nil {
			return fmt.Errorf("partial cost requires null cost_usd and numeric known_subtotal_usd")
		}
	case CostUnavailable:
		if amount != nil || evidence.KnownSubtotalUSD != nil {
			return fmt.Errorf("unavailable cost requires null amounts")
		}
	default:
		return fmt.Errorf("invalid cost_evidence state %q", evidence.State)
	}
	return nil
}

// CostProvenance classifies combined measured and API-equivalent billing provenance.
type CostProvenance string

const (
	CostProvenanceEmpty       CostProvenance = "empty"
	CostProvenanceMeasured    CostProvenance = "measured"
	CostProvenanceEstimated   CostProvenance = "estimated"
	CostProvenanceHybrid      CostProvenance = "hybrid"
	CostProvenancePartial     CostProvenance = "partial"
	CostProvenanceUnavailable CostProvenance = "unavailable"
)

// EquivalentCostSummary separates source-reported measured spend from persisted API-equivalent estimates.
type EquivalentCostSummary struct {
	Provenance             CostProvenance `json:"provenance"`
	MeasuredCostUSD        *float64       `json:"measured_cost_usd"`
	EstimatedCostUSD       *float64       `json:"estimated_cost_usd"`
	TotalEquivalentCostUSD *float64       `json:"total_equivalent_cost_usd"`
	KnownSubtotalUSD       *float64       `json:"known_subtotal_usd"`
	Measured               int64          `json:"measured"`
	Estimated              int64          `json:"estimated"`
	PartialMeasured        int64          `json:"partial_measured"`
	Unavailable            int64          `json:"unavailable"`
	Total                  int64          `json:"total"`
}

func copyCostPtr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	return CostNumber(*value)
}

// SummaryForInvocationEquivalentCost projects one invocation into an EquivalentCostSummary
// using only its persisted measured evidence and persisted cost_estimate.
func SummaryForInvocationEquivalentCost(inv InvocationFact) EquivalentCostSummary {
	evidence := EffectiveCost(inv.CostUSD, inv.CostEvidence)
	result := EquivalentCostSummary{Total: 1}
	switch evidence.State {
	case CostObserved:
		result.Measured = 1
		result.MeasuredCostUSD = CostNumber(inv.CostUSD)
	case CostPartial:
		result.PartialMeasured = 1
		result.MeasuredCostUSD = copyCostPtr(evidence.KnownSubtotalUSD)
	default:
		if inv.CostEstimate != nil {
			result.Estimated = 1
			result.EstimatedCostUSD = CostNumber(inv.CostEstimate.AmountUSD)
		} else {
			result.Unavailable = 1
		}
	}
	return FinalizeEquivalentCostSummary(result)
}

// MergeEquivalentCostSummaries combines two equivalent-cost summaries without
// premature per-invocation rounding.
func MergeEquivalentCostSummaries(a, b EquivalentCostSummary) EquivalentCostSummary {
	result := EquivalentCostSummary{
		Measured:        a.Measured + b.Measured,
		Estimated:       a.Estimated + b.Estimated,
		PartialMeasured: a.PartialMeasured + b.PartialMeasured,
		Unavailable:     a.Unavailable + b.Unavailable,
		Total:           a.Total + b.Total,
	}
	if a.MeasuredCostUSD != nil || b.MeasuredCostUSD != nil {
		sum := 0.0
		if a.MeasuredCostUSD != nil {
			sum += *a.MeasuredCostUSD
		}
		if b.MeasuredCostUSD != nil {
			sum += *b.MeasuredCostUSD
		}
		result.MeasuredCostUSD = CostNumber(sum)
	}
	if a.EstimatedCostUSD != nil || b.EstimatedCostUSD != nil {
		sum := 0.0
		if a.EstimatedCostUSD != nil {
			sum += *a.EstimatedCostUSD
		}
		if b.EstimatedCostUSD != nil {
			sum += *b.EstimatedCostUSD
		}
		result.EstimatedCostUSD = CostNumber(sum)
	}
	return FinalizeEquivalentCostSummary(result)
}

// FinalizeEquivalentCostSummary derives exact provenance and nullable totals from accumulated counts and subtotals.
func FinalizeEquivalentCostSummary(summary EquivalentCostSummary) EquivalentCostSummary {
	if summary.Total == 0 {
		return EquivalentCostSummary{Provenance: CostProvenanceEmpty}
	}
	if summary.Measured == 0 && summary.PartialMeasured == 0 {
		summary.MeasuredCostUSD = nil
	} else {
		summary.MeasuredCostUSD = copyCostPtr(summary.MeasuredCostUSD)
	}
	if summary.Estimated == 0 {
		summary.EstimatedCostUSD = nil
	} else {
		summary.EstimatedCostUSD = copyCostPtr(summary.EstimatedCostUSD)
	}

	if summary.MeasuredCostUSD != nil || summary.EstimatedCostUSD != nil {
		subtotal := 0.0
		if summary.MeasuredCostUSD != nil {
			subtotal += *summary.MeasuredCostUSD
		}
		if summary.EstimatedCostUSD != nil {
			subtotal += *summary.EstimatedCostUSD
		}
		summary.KnownSubtotalUSD = CostNumber(subtotal)
	} else {
		summary.KnownSubtotalUSD = nil
	}

	switch {
	case summary.Measured == summary.Total:
		summary.Provenance = CostProvenanceMeasured
		summary.EstimatedCostUSD = nil
		summary.TotalEquivalentCostUSD = copyCostPtr(summary.MeasuredCostUSD)
	case summary.Estimated == summary.Total:
		summary.Provenance = CostProvenanceEstimated
		summary.MeasuredCostUSD = nil
		summary.TotalEquivalentCostUSD = copyCostPtr(summary.EstimatedCostUSD)
	case summary.Measured+summary.Estimated == summary.Total && summary.Measured > 0 && summary.Estimated > 0:
		summary.Provenance = CostProvenanceHybrid
		summary.TotalEquivalentCostUSD = copyCostPtr(summary.KnownSubtotalUSD)
	case summary.KnownSubtotalUSD != nil:
		summary.Provenance = CostProvenancePartial
		summary.TotalEquivalentCostUSD = nil
	default:
		summary.Provenance = CostProvenanceUnavailable
		summary.MeasuredCostUSD = nil
		summary.EstimatedCostUSD = nil
		summary.TotalEquivalentCostUSD = nil
		summary.KnownSubtotalUSD = nil
	}
	return summary
}

// EffectiveEquivalentCostSummary returns a finalized EquivalentCostSummary,
// falling back to measured-only CostSummary projection for older in-memory callers.
func EffectiveEquivalentCostSummary(equiv EquivalentCostSummary, summary CostSummary, amount float64, count int64) EquivalentCostSummary {
	if equiv.Provenance != "" {
		return FinalizeEquivalentCostSummary(equiv)
	}
	measured := EffectiveCostSummary(summary, amount, count)
	derived := EquivalentCostSummary{
		Measured:        measured.FullyObserved,
		PartialMeasured: measured.Partial,
		Unavailable:     measured.Unavailable,
		Total:           measured.Total,
	}
	if measured.FullyObserved > 0 || measured.Partial > 0 {
		derived.MeasuredCostUSD = copyCostPtr(measured.KnownSubtotalUSD)
	}
	return FinalizeEquivalentCostSummary(derived)
}
