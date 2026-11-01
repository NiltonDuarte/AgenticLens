package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// CaptureIdentity scopes an exact harness session to its source store.
type CaptureIdentity struct {
	Source    SourceHarness `json:"source"`
	Store     string        `json:"store"`
	SessionID string        `json:"session_id"`
}

func (id CaptureIdentity) Validate() error {
	if id.Source != SourceCodex && id.Source != SourceOpenCode {
		return errors.New("capture source must be codex or opencode")
	}
	if strings.TrimSpace(id.Store) == "" || strings.TrimSpace(id.SessionID) == "" {
		return errors.New("capture requires an exact source store and session ID")
	}
	for _, r := range id.SessionID {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return errors.New("invalid capture session ID")
		}
	}
	return nil
}

// CaptureBinding never derives ticket ownership from ambient project metadata.
func CaptureBinding(existing, requested, inherited string) (string, error) {
	if requested == "" {
		requested = inherited
	}
	if existing != "" && requested != "" && existing != requested {
		return "", fmt.Errorf("session already bound to %s; refusing conflicting ticket %s", existing, requested)
	}
	if existing != "" {
		return existing, nil
	}
	if strings.TrimSpace(requested) == "" {
		return "", nil
	}
	for _, r := range requested {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return "", errors.New("ticket ID must contain only letters, digits, hyphens or underscores")
		}
	}
	return requested, nil
}

type CaptureState string

const (
	CapturePending     CaptureState = "pending"
	CaptureUnassigned  CaptureState = "unassigned"
	CaptureFailed      CaptureState = "failed"
	CaptureUnsupported CaptureState = "unsupported"
	CaptureCurrent     CaptureState = "current"
)

// CaptureRegistration contains local correlation metadata, never transcript content.
type CaptureRegistration struct {
	Identity       CaptureIdentity `json:"identity"`
	TicketID       string          `json:"ticket_id"`
	Binding        string          `json:"binding,omitempty"`
	Request        string          `json:"request"`
	TranscriptPath string          `json:"-"`
	ParentID       string          `json:"parent_id,omitempty"`
	Event          string          `json:"event"`
	State          CaptureState    `json:"state"`
	Detail         string          `json:"detail,omitempty"`
	RefreshedAt    string          `json:"refreshed_at,omitempty"`
	UpdatedAt      string          `json:"updated_at"`
	Revision       int64           `json:"revision"`
	Snapshot       []SessionFact   `json:"-"`
	Skills         []SkillFact     `json:"-"`
	TaskLinks      []TaskLinkFact  `json:"-"`
}

// CaptureSnapshotAccept rejects a truncated or older cumulative source snapshot.
func CaptureSnapshotAccept(previousRevision, revision int64, previousTime, nextTime string) bool {
	if revision < previousRevision {
		return false
	}
	if previousTime == "" {
		return true
	}
	previous, err := time.Parse(time.RFC3339Nano, previousTime)
	if err != nil {
		return false
	}
	next, err := time.Parse(time.RFC3339Nano, nextTime)
	if err != nil {
		return false
	}
	return !next.Before(previous)
}

// CaptureOwnerCheckpoint is the latest source-reported invocation completion.
// Ledger publication time, filesystem time and token totals are not substitutes.
func CaptureOwnerCheckpoint(owner SessionFact) (time.Time, bool) {
	var latest time.Time
	known := false
	for _, invocation := range owner.Invocations {
		if invocation.CompletedAt == nil {
			continue
		}
		value, err := time.Parse(time.RFC3339Nano, *invocation.CompletedAt)
		if err != nil {
			return time.Time{}, false
		}
		if !known || value.After(latest) {
			latest = value
			known = true
		}
	}
	return latest, known
}
