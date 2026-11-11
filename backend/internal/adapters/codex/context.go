package codex

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
	return readSessionContext(path, maxSessionMetadataBytes)
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
	return readSessionContext(path, maxSourceBytes)
}

func readSessionContext(path string, limit int64) (result SessionMetadataContext, err error) {
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return result, err
	}
	reader := bufio.NewReader(io.LimitReader(file, limit))
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
	if metadata.ID == "" || metadata.CWD == "" {
		return result, errors.New("exact native session metadata unavailable")
	}
	// Both authoritative ancestry spellings must agree when both are present.
	var nested struct {
		Subagent struct {
			ThreadSpawn struct {
				ParentID string `json:"parent_thread_id"`
			} `json:"thread_spawn"`
		} `json:"subagent"`
	}
	if err := json.Unmarshal(metadata.Source, &nested); err == nil && nested.Subagent.ThreadSpawn.ParentID != "" && metadata.ParentID != "" && nested.Subagent.ThreadSpawn.ParentID != metadata.ParentID {
		return result, fmt.Errorf("%w: conflicting native parent metadata", ErrSessionContextMismatch)
	}
	return SessionMetadataContext{ID: metadata.ID, ParentID: metadata.ParentID, WorkingDirectory: metadata.CWD}, nil
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
