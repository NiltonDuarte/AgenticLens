package adapters

import (
	"agenticlens/internal/domain"
)

func factCostAmount(amount float64, evidence domain.CostEvidence) *float64 {
	if evidence.State == domain.CostObserved {
		return domain.CostNumber(amount)
	}
	return nil
}

func costFactEvidence(amount *float64, evidence *domain.CostEvidence) domain.CostEvidence {
	if evidence != nil {
		return *evidence
	}
	if amount != nil {
		return domain.ObservedCost(*amount)
	}
	return domain.UnavailableCost()
}

func costFactAccumulator(amount *float64, evidence domain.CostEvidence) float64 {
	if evidence.State == domain.CostPartial && evidence.KnownSubtotalUSD != nil {
		return *evidence.KnownSubtotalUSD
	}
	if amount != nil {
		return *amount
	}
	return 0
}

func costEvidencePointer(evidence domain.CostEvidence) *domain.CostEvidence { return &evidence }
