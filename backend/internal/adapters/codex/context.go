package codex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// ErrSessionContextMismatch distinguishes positively contradictory metadata from
// unavailable or partially flushed source context.
var ErrSessionContextMismatch = errors.New("native session context mismatch")

// SessionMetadataContext contains only identity and scope facts from the native
// metadata record. It does not retain instructions, messages or event payloads.
type SessionMetadataContext struct {
	ID               string
	ParentID         string
	WorkingDirectory string
}

const maxSessionMetadataBytes int64 = 64 * 1024

// ReadSessionContext retains the intrinsic structural record bound for callers
// which do not consume a configured capture byte budget.
func ReadSessionContext(path string) (SessionMetadataContext, error) {
	return readSessionContext(path, maxSessionMetadataBytes, nil)
}

// ReadSessionContextWithinBudget applies the explicit capture allowance before
// any read, while preserving the existing maximum native metadata record size.
func ReadSessionContextWithinBudget(path string, maxSourceBytes int64) (SessionMetadataContext, error) {
	if maxSourceBytes < 1 {
		return SessionMetadataContext{}, errors.New("native session metadata requires a positive explicit source byte budget")
	}
	if maxSourceBytes > maxSessionMetadataBytes {
		maxSourceBytes = maxSessionMetadataBytes
	}
	return readSessionContext(path, maxSourceBytes, nil)
}

// ReadSessionContextWithBudget charges shared source reads including buffered
// bytes and failed inspections.
func ReadSessionContextWithBudget(path string, budget *DiscoveryBudget) (SessionMetadataContext, error) {
	if err := budget.claim(path); err != nil {
		return SessionMetadataContext{}, err
	}
	return readSessionContext(path, maxSessionMetadataBytes, budget)
}

func readSessionContext(path string, limit int64, budget *DiscoveryBudget) (result SessionMetadataContext, err error) {
	file, info, err := openRolloutSource(path)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	var source io.Reader = file
	if budget != nil {
		source = &discoveryBudgetReader{reader: file, budget: budget, remainingSize: info.Size()}
	}
	reader := bufio.NewReader(io.LimitReader(source, limit))
	line, err := reader.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return result, err
	}
	if len(line) == 0 {
		return result, errors.New("session metadata has not flushed")
	}
	if !bytes.HasSuffix(line, []byte{'\n'}) && info.Size() > limit {
		if limit < maxSessionMetadataBytes {
			return result, ErrDiscoveryBudget
		}
		return result, errors.New("native session metadata exceeds its structural record limit")
	}

	var record rolloutRecord
	if err := json.Unmarshal(line, &record); err != nil {
		return result, err
	}
	if record.Type != "session_meta" {
		return result, errors.New("exact native session metadata unavailable")
	}
	metadata, err := decodeSessionMetadata(record.Payload)
	if err != nil {
		return result, err
	}
	return exactSessionContext(metadata.ID, metadata.ParentID, metadata.CWD, conflictingNativeParent(metadata))
}

// exactSessionContext applies the exact-owner metadata rules shared by every
// capture header reader: an identity and working directory are required, and
// both ancestry spellings must agree.
func exactSessionContext(id, parentID, workingDirectory string, conflictingParent bool) (SessionMetadataContext, error) {
	if id == "" || workingDirectory == "" {
		return SessionMetadataContext{}, errors.New("exact native session metadata unavailable")
	}
	if conflictingParent {
		return SessionMetadataContext{}, fmt.Errorf("%w: conflicting native parent metadata", ErrSessionContextMismatch)
	}
	return SessionMetadataContext{ID: id, ParentID: parentID, WorkingDirectory: workingDirectory}, nil
}

// conflictingNativeParent reports a session_meta whose direct parent and
// subagent thread_spawn parent are both present and differ.
func conflictingNativeParent(metadata sessionMetadata) bool {
	var nested struct {
		Subagent struct {
			ThreadSpawn struct {
				ParentID string `json:"parent_thread_id"`
			} `json:"thread_spawn"`
		} `json:"subagent"`
	}
	return json.Unmarshal(metadata.Source, &nested) == nil && nested.Subagent.ThreadSpawn.ParentID != "" && metadata.ParentID != "" && nested.Subagent.ThreadSpawn.ParentID != metadata.ParentID
}

// SessionContext requires exact source identity for ordinary root/parent reads.
// Subagent lifecycle normalization validates ancestry separately without ever
// relabeling the parent callback as a child.
func SessionContext(path, sessionID string) (string, error) {
	metadata, err := ReadSessionContext(path)
	if err != nil {
		return "", err
	}
	if metadata.ID != sessionID {
		return "", fmt.Errorf("%w: callback %s, source %s", ErrSessionContextMismatch, sessionID, metadata.ID)
	}
	return metadata.WorkingDirectory, nil
}
