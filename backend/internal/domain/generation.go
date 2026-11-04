package domain

// GenerationPolicy records which optional categories a generated ledger includes.
// A missing policy on historical ledgers preserves the enabled defaults.
type GenerationPolicy struct {
	Skills      bool `json:"skills" yaml:"skills"`
	Delegations bool `json:"delegations" yaml:"delegations"`
	Anomalies   bool `json:"anomalies" yaml:"anomalies"`
}

// GenerationConfig configures optional output for generate only.
type GenerationConfig struct {
	Metrics GenerationPolicy `json:"metrics" yaml:"metrics"`
}

func DefaultGenerationPolicy() GenerationPolicy {
	return GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
}

func (ledger AgenticFactLedger) EffectiveGenerationPolicy() GenerationPolicy {
	if ledger.GenerationPolicy == nil {
		return DefaultGenerationPolicy()
	}
	return *ledger.GenerationPolicy
}

// ProjectGenerationFacts returns an independent owner slice, suppressing optional
// categories without changing ordinary execution facts or the source snapshot.
func ProjectGenerationFacts(ledger AgenticFactLedger, policy GenerationPolicy) AgenticFactLedger {
	ledger.GenerationPolicy = &policy
	ledger.Sessions = append([]SessionFact{}, ledger.Sessions...)
	if !policy.Anomalies {
		ledger.Anomalies = nil
	}
	for i := range ledger.Sessions {
		if policy.Anomalies && ledger.Sessions[i].Anomalies == nil {
			ledger.Sessions[i].AnomalyAvailability = AnomalyUnavailable
		}
		if !policy.Skills {
			ledger.Sessions[i].Skills = nil
		}
		if !policy.Delegations {
			ledger.Sessions[i].Delegations = nil
		}
		if !policy.Anomalies {
			ledger.Sessions[i].Anomalies = nil
		}
	}
	return ledger
}

// UnavailableAnomalySessions counts owners whose omitted analysis cannot be
// recovered without a complete source-backed refresh. Other owners stay usable.
func (ledger AgenticFactLedger) UnavailableAnomalySessions() int {
	count := 0
	for _, owner := range ledger.Sessions {
		if owner.AnomalyAvailability == AnomalyUnavailable {
			count++
		}
	}
	return count
}
