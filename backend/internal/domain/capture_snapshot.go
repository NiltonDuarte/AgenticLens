package domain

import "encoding/json"

// CaptureOwnerFactsEqual compares compact source facts independently of user
// attribution, publication-time rate-card estimates, and optional event hydration.
// Equal or unknown checkpoints alone do not establish equality of selected snapshots
// or their provenance.
func CaptureOwnerFactsEqual(a, b SessionFact) (bool, error) {
	// User identity is separate attribution metadata; compare whole compact source
	// facts rather than selecting fields by numerical maximum.
	a.User = ""
	b.User = ""
	if a.Source == "" {
		a.Source = b.Source
	}
	if b.Source == "" {
		b.Source = a.Source
	}
	// Rate-card estimates are publication-time derivations, not raw capture source facts.
	a.Invocations = append([]InvocationFact{}, a.Invocations...)
	for i := range a.Invocations {
		a.Invocations[i].CostEstimate = nil
	}
	b.Invocations = append([]InvocationFact{}, b.Invocations...)
	for i := range b.Invocations {
		b.Invocations[i].CostEstimate = nil
	}
	a.Tools = append([]SessionToolFact{}, a.Tools...)
	b.Tools = append([]SessionToolFact{}, b.Tools...)
	a.Anomalies = append([]Anomaly{}, a.Anomalies...)
	b.Anomalies = append([]Anomaly{}, b.Anomalies...)
	first, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	second, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return string(first) == string(second), nil
}
