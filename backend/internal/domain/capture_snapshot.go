package domain

import "encoding/json"

// CaptureOwnerFactsEqual compares compact source facts independently of user
// attribution and optional event hydration. Equal or unknown checkpoints alone
// do not establish equality of selected snapshots or their provenance.
func CaptureOwnerFactsEqual(a, b SessionFact) (bool, error) {
	// User identity is separate attribution metadata; compare whole compact source
	// facts rather than selecting fields by numerical maximum.
	a.User = ""
	b.User = ""
	a.Events = nil
	b.Events = nil
	if a.Source == "" {
		a.Source = b.Source
	}
	if b.Source == "" {
		b.Source = a.Source
	}
	a.Invocations = append([]InvocationFact{}, a.Invocations...)
	b.Invocations = append([]InvocationFact{}, b.Invocations...)
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
