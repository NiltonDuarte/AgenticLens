package capture

import (
	"errors"
	"fmt"

	"agenticlens/internal/adapters/codex"
	config "agenticlens/internal/adapters/config"
)

// NormalizeCodexSubagentStart handles the native start shape: session_id names
// the parent while transcript_path may name the new child's rollout. Retain the
// child only as a pending source pointer and resolve the parent's own transcript
// independently. Delayed child metadata is not an authoritative contradiction.
func NormalizeCodexSubagentStart(input HookInput, store, project string) (HookInput, error) {
	if input.Event != "SubagentStart" || input.TranscriptPath == "" {
		return input, nil
	}
	if err := codex.ValidateRolloutStore(store, input.TranscriptPath); err != nil {
		return input, err
	}
	metadata, readErr := codex.ReadSessionContext(input.TranscriptPath)
	if errors.Is(readErr, codex.ErrSessionContextMismatch) {
		return input, readErr
	}
	if readErr == nil {
		if metadata.ID == input.SessionID {
			return input, nil
		}
		if metadata.ParentID != input.SessionID {
			return input, fmt.Errorf("%w: callback parent %s, child %s has parent %s", codex.ErrSessionContextMismatch, input.SessionID, metadata.ID, metadata.ParentID)
		}
		childProject, err := config.ProjectRoot(metadata.WorkingDirectory)
		if err != nil {
			return input, err
		}
		same, err := SameStore(project, childProject)
		if err != nil {
			return input, err
		}
		if !same {
			return input, fmt.Errorf("%w: callback project %s differs from child source cwd %s", codex.ErrSessionContextMismatch, project, metadata.WorkingDirectory)
		}
	}
	// The unflushed path might be the child. It must never overwrite a known
	// parent pointer, even if exact parent discovery is temporarily unavailable.
	childPath := input.TranscriptPath
	input.TranscriptPath = ""
	input.AgentTranscriptPath = childPath
	if parentPath, err := codex.DiscoverExactRolloutPath(store, input.SessionID); err == nil {
		input.TranscriptPath = parentPath
	}
	return input, nil
}
