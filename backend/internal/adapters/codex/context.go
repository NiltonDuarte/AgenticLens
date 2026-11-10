package codex

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
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

// ReadSessionContext reads only the first bounded native metadata record.
func ReadSessionContext(path string) (SessionMetadataContext, error) {
	var result SessionMetadataContext
	file, err := os.Open(path)
	if err != nil {
		return result, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	found := scanner.Scan()
	if err := errors.Join(scanner.Err(), file.Close()); err != nil {
		return result, err
	}
	if !found {
		return result, errors.New("session metadata has not flushed")
	}
	var record rolloutRecord
	if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
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
