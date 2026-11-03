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

// CaptureAuxiliaryUpgrade permits only a source-backed expansion into previously
// unsupported empty auxiliary fields. It does not order changed metric snapshots.
func CaptureAuxiliaryUpgrade(before, after SessionFact) (bool, error) {
	if before.AuxiliaryExtractionVersion != 0 || after.AuxiliaryExtractionVersion != CurrentAuxiliaryExtractionVersion || len(before.Skills) != 0 || len(before.TaskLinks) != 0 || before.ID != after.ID || before.User != after.User || before.Source == "" || before.Source != after.Source {
		return false, nil
	}
	// Optional source hydration may add events, but accepted persisted event facts
	// must never be replaced by differing evidence under an extraction upgrade.
	if len(before.Events) > 0 {
		a, err := json.Marshal(before.Events)
		if err != nil {
			return false, err
		}
		b, err := json.Marshal(after.Events)
		if err != nil {
			return false, err
		}
		if string(a) != string(b) {
			return false, nil
		}
	}
	after.Skills, after.TaskLinks = before.Skills, before.TaskLinks
	after.AuxiliaryExtractionVersion = before.AuxiliaryExtractionVersion
	return CaptureOwnerFactsEqual(before, after)
}
