package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"agenticlens/internal/domain"
)

type HookInput struct {
	SessionID           string `json:"session_id"`
	Event               string `json:"hook_event_name"`
	TranscriptPath      string `json:"transcript_path"`
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
	StopHookActive      bool   `json:"stop_hook_active"`
}
type HookContext struct {
	Event   string `json:"hookEventName"`
	Context string `json:"additionalContext"`
}
type HookOutput struct {
	Context *HookContext `json:"hookSpecificOutput,omitempty"`
	Message string       `json:"systemMessage,omitempty"`
}

func ReadHook(input io.Reader) (HookInput, error) {
	b, err := io.ReadAll(io.LimitReader(input, MaxPayloadBytes+1))
	if err != nil {
		return HookInput{}, err
	}
	if len(b) > MaxPayloadBytes {
		return HookInput{}, errors.New("hook payload exceeds 64 KiB limit")
	}
	var value HookInput
	if err = json.Unmarshal(b, &value); err != nil {
		return value, fmt.Errorf("invalid hook JSON: %w", err)
	}
	if value.SessionID == "" || value.Event == "" {
		return value, errors.New("hook requires session_id and hook_event_name")
	}
	return value, nil
}

// ReadHookContext bounds both volume and elapsed handoff time. Closing a native
// pipe on cancellation releases its blocked read; parsing never commits events.
func ReadHookContext(ctx context.Context, input io.Reader) (HookInput, error) {
	type readResult struct {
		value HookInput
		err   error
	}
	result := make(chan readResult, 1)
	go func() { value, err := ReadHook(input); result <- readResult{value, err} }()
	select {
	case value := <-result:
		return value.value, value.err
	case <-ctx.Done():
		err := ctx.Err()
		if closer, ok := input.(io.Closer); ok {
			err = errors.Join(err, closer.Close())
		}
		return HookInput{}, err
	}
}

func QuoteArgument(value string) string { return "'" + replaceQuotes(value) + "'" }
func replaceQuotes(s string) string {
	result := ""
	for _, r := range s {
		if r == '\'' {
			result += "'\"'\"'"
		} else {
			result += string(r)
		}
	}
	return result
}

func BindingPrompt(r domain.CaptureRegistration, binary, directory string) string {
	return fmt.Sprintf("AgenticLens requires explicit ticket information for this exact %s session %s. Do not infer a ticket from directory, branch, role or timing. If the user has explicitly identified the active ticket, respond now by executing: %s capture bind --state-dir %s --source %s --store %s --session-id %s --request %s --ticket-id <EXPLICIT_TICKET_ID>. If no ticket was explicitly supplied, ask the user for it; leave capture unassigned. Do not invent source, parent, model or usage fields.", r.Identity.Source, r.Identity.SessionID, QuoteArgument(binary), QuoteArgument(directory), r.Identity.Source, QuoteArgument(r.Identity.Store), QuoteArgument(r.Identity.SessionID), QuoteArgument(r.Request))
}

func ReceiveCodex(ctx context.Context, s *Store, input HookInput, sourceStore, binary, directory string) (HookOutput, error) {
	switch input.Event {
	case "SessionStart", "SubagentStart", "SubagentStop", "Stop", "SessionEnd", "UserPromptSubmit":
	default:
		return HookOutput{}, errors.New("unsupported Codex lifecycle event")
	}
	if !filepath.IsAbs(sourceStore) {
		return HookOutput{}, errors.New("Codex store must be absolute")
	}
	canonicalStore, normalizeErr := CanonicalStore(sourceStore)
	if normalizeErr != nil {
		return HookOutput{}, normalizeErr
	}
	sourceStore = canonicalStore
	r, err := s.Register(ctx, domain.CaptureIdentity{Source: domain.SourceCodex, Store: sourceStore, SessionID: input.SessionID}, input.TranscriptPath, input.Event)
	if err != nil {
		return HookOutput{}, err
	}
	// Subagent callbacks carry the parent's session_id. Never attribute that ID to the child.
	if input.Event == "SubagentStart" {
		if err = s.QueueChildStart(ctx, sourceStore, input.SessionID, input.AgentID); err != nil {
			return HookOutput{}, err
		}
		return HookOutput{Context: &HookContext{Event: input.Event, Context: "AgenticLens observed your start. The callback identifies your parent, not your own source identity. Do not bind using the parent's ID. Your identity will be reconciled only from authoritative session metadata; if independently launched, respond to your own SessionStart binding request."}}, nil
	}
	if input.Event == "SubagentStop" && input.AgentTranscriptPath != "" {
		// Child registration is deferred to reconciliation of a proven source path, rather than assuming agent_id equals a harness ID.
		if err = s.QueueChild(ctx, sourceStore, input.AgentTranscriptPath, input.SessionID); err != nil {
			return HookOutput{}, err
		}
	}
	if r.TicketID == "" && (input.Event == "SessionStart" || input.Event == "UserPromptSubmit") {
		return HookOutput{Context: &HookContext{Event: input.Event, Context: BindingPrompt(r, binary, directory)}}, nil
	}
	return HookOutput{}, nil
}
