package domain

// GenerationPolicy configures optional generate output. It is transient and is
// never persisted as ledger metadata.
type GenerationPolicy struct {
	Skills      bool `json:"skills" yaml:"skills"`
	Delegations bool `json:"delegations" yaml:"delegations"`
	Anomalies   bool `json:"anomalies" yaml:"anomalies"`
}

type GenerationConfig struct {
	Metrics  GenerationPolicy         `json:"metrics" yaml:"metrics"`
	Feedback GenerationFeedbackConfig `json:"feedback" yaml:"feedback"`
}

func DefaultGenerationPolicy() GenerationPolicy {
	return GenerationPolicy{Skills: true, Delegations: true, Anomalies: true}
}

// ProjectGenerationFacts preserves slice presence for enabled categories and
// suppresses disabled categories without changing caller-owned sessions.
func ProjectGenerationFacts(ledger AgenticFactLedger, policy GenerationPolicy) AgenticFactLedger {
	ledger.Sessions = append([]SessionFact{}, ledger.Sessions...)
	if !policy.Anomalies {
		ledger.Anomalies = nil
	}
	for i := range ledger.Sessions {
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

func (ledger AgenticFactLedger) UnrecordedAnomalySessions() int {
	count := 0
	for _, owner := range ledger.Sessions {
		if owner.Anomalies == nil {
			count++
		}
	}
	return count
}

// CompleteAnomalyCoverage requires independently recorded root and owner analysis.
func (ledger AgenticFactLedger) CompleteAnomalyCoverage() bool {
	return ledger.Anomalies != nil && ledger.UnrecordedAnomalySessions() == 0
}

// UnrecordedAnomalyScopes includes the independently observed root/group scope.
func (ledger AgenticFactLedger) UnrecordedAnomalyScopes() int {
	count := ledger.UnrecordedAnomalySessions()
	if ledger.Anomalies == nil {
		count++
	}
	return count
}

// RetainUnrecordedCategories preserves recorded optional evidence when a fresh
// ordinary snapshot did not establish a replacement for that category. A
// recorded empty slice remains authoritative and may remove stale occurrences.
func RetainUnrecordedCategories(previous, fresh []SessionFact) []SessionFact {
	prior := make(map[string]SessionFact, len(previous))
	for _, owner := range previous {
		prior[owner.ID] = owner
	}
	result := append([]SessionFact{}, fresh...)
	for i, owner := range result {
		old, exists := prior[owner.ID]
		if !exists || old.Source != "" && owner.Source != "" && old.Source != owner.Source {
			continue
		}
		if owner.Skills == nil {
			result[i].Skills = old.Skills
		}
		if owner.Delegations == nil {
			result[i].Delegations = old.Delegations
		}
		if owner.Anomalies == nil {
			result[i].Anomalies = old.Anomalies
		}
	}
	return result
}

// GenerationFeedbackConfig controls human diagnostics from explicit generation.
type GenerationFeedbackConfig struct {
	ProvisionalThresholdSeconds int64 `json:"provisional_threshold_seconds" yaml:"provisional_threshold_seconds"`
}

func DefaultGenerationFeedbackConfig() GenerationFeedbackConfig {
	return GenerationFeedbackConfig{ProvisionalThresholdSeconds: 180}
}
