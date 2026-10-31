package migration

import "agenticlens/internal/domain"

// StepV4ToV5CostEvidence records the availability of invocation USD costs.
// Numeric legacy values remain observed, including zero, without source lookup.
type StepV4ToV5CostEvidence struct{}

func NewStepV4ToV5CostEvidence() *StepV4ToV5CostEvidence { return &StepV4ToV5CostEvidence{} }
func (s *StepV4ToV5CostEvidence) SourceVersion() int     { return 4 }
func (s *StepV4ToV5CostEvidence) TargetVersion() int     { return 5 }
func (s *StepV4ToV5CostEvidence) Description() string {
	return "Adds schema v5 cost availability evidence"
}
func (s *StepV4ToV5CostEvidence) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	cloned, err := domain.PersistAnomalySummaries(ledger)
	if err != nil {
		return domain.AgenticFactLedger{}, err
	}
	version := 5
	cloned.SchemaVersion = &version
	for si := range cloned.Sessions {
		for ii := range cloned.Sessions[si].Invocations {
			inv := &cloned.Sessions[si].Invocations[ii]
			inv.CostEvidence = domain.EffectiveCost(inv.CostUSD, inv.CostEvidence)
		}
	}
	cloned.TaskLinks = append([]domain.TaskLinkFact(nil), ledger.TaskLinks...)
	for i := range cloned.TaskLinks {
		task := &cloned.TaskLinks[i]
		if task.CostEvidence.State == "" {
			if task.CostUSD == 0 {
				task.CostEvidence = domain.UnavailableCost()
			} else {
				task.CostEvidence = domain.ObservedCost(task.CostUSD)
			}
		}
	}
	if cloned.Migration != nil {
		info := *cloned.Migration
		info.SchemaVersion = 5
		info.StepsCompleted = append(append([]string{}, info.StepsCompleted...), "schema_v5_cost_evidence")
		info.StepsPending = append([]string{}, info.StepsPending...)
		cloned.Migration = &info
	}
	return cloned, nil
}
