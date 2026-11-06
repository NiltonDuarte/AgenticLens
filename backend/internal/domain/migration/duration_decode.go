package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

// draftDurationEvidence is accepted only at input compatibility boundaries.
// New canonical facts carry the numeric provisional portion directly.
type draftDurationEvidence struct {
	State              string `json:"state"`
	CompletedSeconds   int64  `json:"completed_seconds"`
	ProvisionalSeconds int64  `json:"provisional_seconds"`
	Cutoff             string `json:"cutoff,omitempty"`
	Basis              string `json:"basis"`
	Limitation         string `json:"limitation,omitempty"`
}

// DecodeProvisionalSeconds accepts original, draft and numeric wire inputs.
// Raw member presence distinguishes an explicit zero from an absent member.
func DecodeProvisionalSeconds(total int64, flat, draft json.RawMessage) (int64, error) {
	var provisional int64
	if len(flat) > 0 {
		if bytes.Equal(bytes.TrimSpace(flat), []byte("null")) {
			return 0, fmt.Errorf("provisional_seconds must be an integer, not null")
		}
		if err := json.Unmarshal(flat, &provisional); err != nil {
			return 0, fmt.Errorf("decoding provisional_seconds: %w", err)
		}
	}
	if len(draft) > 0 {
		if bytes.Equal(bytes.TrimSpace(draft), []byte("null")) {
			return 0, fmt.Errorf("duration_evidence must be an object, not null")
		}
		if _, err := durationObjectFields(draft); err != nil {
			return 0, fmt.Errorf("draft duration_evidence: %w", err)
		}
		var old draftDurationEvidence
		decoder := json.NewDecoder(bytes.NewReader(draft))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&old); err != nil {
			return 0, fmt.Errorf("decoding draft duration_evidence: %w", err)
		}
		var extra json.RawMessage
		if err := decoder.Decode(&extra); err != io.EOF {
			return 0, fmt.Errorf("draft duration_evidence must contain one JSON value")
		}
		if old.CompletedSeconds < 0 || old.ProvisionalSeconds < 0 || old.CompletedSeconds > total || old.ProvisionalSeconds != total-old.CompletedSeconds || strings.TrimSpace(old.Basis) == "" {
			return 0, fmt.Errorf("draft duration components require a basis and nonnegative components equal to active duration")
		}
		switch old.State {
		case "complete":
			if old.Cutoff != "" || old.ProvisionalSeconds != 0 || old.Limitation != "" {
				return 0, fmt.Errorf("complete draft duration cannot carry provisional time or limitations")
			}
		case "provisional":
			if old.Cutoff == "" || old.Limitation != "" {
				return 0, fmt.Errorf("provisional draft duration requires a cutoff and no unsupported coverage")
			}
		case "partial":
			if strings.TrimSpace(old.Limitation) == "" {
				return 0, fmt.Errorf("partial draft duration requires a limitation")
			}
		default:
			return 0, fmt.Errorf("unsupported draft duration coverage %q", old.State)
		}
		if old.Cutoff != "" {
			if _, err := time.Parse(time.RFC3339Nano, old.Cutoff); err != nil {
				return 0, fmt.Errorf("invalid draft duration cutoff: %w", err)
			}
		}
		if old.ProvisionalSeconds > 0 && old.Cutoff == "" {
			return 0, fmt.Errorf("draft provisional seconds require a cutoff")
		}
		if len(flat) > 0 && provisional != old.ProvisionalSeconds {
			return 0, fmt.Errorf("conflicting provisional_seconds and draft duration_evidence")
		}
		provisional = old.ProvisionalSeconds
	}
	if err := domain.ValidateDuration(total, provisional); err != nil {
		return 0, err
	}
	return provisional, nil
}

// NormalizeInvocationDurationJSON rewrites only the duration representation of
// one input invocation. Unknown non-duration fields stay for the owning strict
// reader to validate; caller-owned bytes and other facts remain unchanged.
func NormalizeInvocationDurationJSON(data []byte) ([]byte, error) {
	fields, err := durationObjectFields(data)
	if err != nil {
		return nil, err
	}
	total, err := decodeActiveDurationSeconds(fields["active_duration_seconds"])
	if err != nil {
		return nil, err
	}
	provisional, err := DecodeProvisionalSeconds(total, fields["provisional_seconds"], fields["duration_evidence"])
	if err != nil {
		return nil, err
	}
	_, hasDraft := fields["duration_evidence"]
	_, hasFlat := fields["provisional_seconds"]
	if !hasDraft && (!hasFlat || provisional > 0) {
		return data, nil
	}
	delete(fields, "duration_evidence")
	delete(fields, "provisional_seconds")
	if provisional > 0 {
		value, err := json.Marshal(provisional)
		if err != nil {
			return nil, err
		}
		fields["provisional_seconds"] = value
	}
	return json.Marshal(fields)
}

// durationObjectFields rejects ambiguous duplicate members before compatibility
// normalization can replace their original spelling or representation.
func durationObjectFields(data []byte) (map[string]json.RawMessage, error) {
	return readDurationObjectFields(data, false)
}

// Historical decoding retains its pre-existing tolerance for unrelated member
// duplication. Duration members alone must remain unambiguous before projection.
func readDurationObjectFields(data []byte, allowUnrelatedDuplicates bool) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if token != json.Delim('{') {
		return nil, fmt.Errorf("duration input must be an object")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("duration input member must be a string")
		}
		// Raw presence checks and Go's case-folded struct matching must agree.
		// Reject only aliases of duration members, rather than changing the
		// owning schema's handling of unrelated JSON fields.
		durationMember := false
		for _, canonical := range []string{"active_duration_seconds", "provisional_seconds", "duration_evidence"} {
			if name == canonical {
				durationMember = true
			}
			if name != canonical && strings.EqualFold(name, canonical) {
				return nil, fmt.Errorf("noncanonical duration member %q; use %s", name, canonical)
			}
		}
		if _, exists := fields[name]; exists && (!allowUnrelatedDuplicates || durationMember) {
			return nil, fmt.Errorf("duplicate duration input member %q", name)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("duration input must contain one JSON object")
	}
	return fields, nil
}

func decodeActiveDurationSeconds(raw json.RawMessage) (int64, error) {
	var total int64
	if len(raw) > 0 {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return 0, fmt.Errorf("active_duration_seconds must be an integer, not null")
		}
		if err := json.Unmarshal(raw, &total); err != nil {
			return 0, fmt.Errorf("decoding active_duration_seconds: %w", err)
		}
	}
	return total, nil
}

// NormalizeSessionDurationJSON preserves the owner's other wire members while
// validating and canonicalizing every duration fact at an input boundary.
func NormalizeSessionDurationJSON(data []byte) ([]byte, error) {
	owner, err := durationObjectFields(data)
	if err != nil {
		return nil, err
	}
	if raw, exists := owner["invocations"]; exists {
		var invocations []json.RawMessage
		if err := json.Unmarshal(raw, &invocations); err != nil {
			return nil, err
		}
		changed := false
		for i, inv := range invocations {
			normalized, err := NormalizeInvocationDurationJSON(inv)
			if err != nil {
				return nil, fmt.Errorf("invocation %d: %w", i, err)
			}
			changed = changed || !bytes.Equal(normalized, inv)
			invocations[i] = normalized
		}
		if !changed {
			return data, nil
		}
		normalized, err := json.Marshal(invocations)
		if err != nil {
			return nil, err
		}
		owner["invocations"] = normalized
		return json.Marshal(owner)
	}
	return data, nil
}

// NormalizeLedgerDurationJSON is the current-schema migration input expansion.
// It does not rewrite any file or infer source completion from numeric values.
func NormalizeLedgerDurationJSON(data []byte) ([]byte, error) {
	fields, err := durationObjectFields(data)
	if err != nil {
		return nil, err
	}
	if raw, exists := fields["sessions"]; exists {
		var sessions []json.RawMessage
		if err := json.Unmarshal(raw, &sessions); err != nil {
			return nil, err
		}
		changed := false
		for i, owner := range sessions {
			normalized, err := NormalizeSessionDurationJSON(owner)
			if err != nil {
				return nil, fmt.Errorf("session %d: %w", i, err)
			}
			changed = changed || !bytes.Equal(normalized, owner)
			sessions[i] = normalized
		}
		if !changed {
			return data, nil
		}
		normalized, err := json.Marshal(sessions)
		if err != nil {
			return nil, err
		}
		fields["sessions"] = normalized
		return json.Marshal(fields)
	}
	return data, nil
}

func validateHistoricalDurationInvocation(data []byte) error {
	fields, err := readDurationObjectFields(data, true)
	if err != nil {
		return err
	}
	total, err := decodeActiveDurationSeconds(fields["active_duration_seconds"])
	if err != nil {
		return err
	}
	_, err = DecodeProvisionalSeconds(total, fields["provisional_seconds"], fields["duration_evidence"])
	return err
}
