package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
	"os"
	"runtime"
	"strings"
)

type HookInput struct {
	WorkingDirectory    string `json:"cwd,omitempty"`
	Activity            string `json:"activity,omitempty"`
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
	Registered bool         `json:"-"`
	Context    *HookContext `json:"hookSpecificOutput,omitempty"`
	Message    string       `json:"systemMessage,omitempty"`
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
	return NativeBindingPrompt(r, binary, directory, runtime.GOOS == "windows")
}

// NativeBindingPrompt renders the executable command for its supported shell.
func NativeBindingPrompt(r domain.CaptureRegistration, binary, directory string, windows bool) string {
	quote := QuoteArgument
	command := quote(binary)
	if windows {
		quote = func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
		command = "& " + quote(binary)
	}
	return fmt.Sprintf("AgenticLens requires explicit ticket information for this exact %s session %s. Do not infer a ticket from directory, branch, role or timing. If the user has explicitly identified the active ticket, respond now by executing: %s capture bind --state-dir %s --source %s --store %s --session-id %s --request %s --ticket-id <EXPLICIT_TICKET_ID> --ticket-title <EXPLICIT_TICKET_TITLE> [--ticket-type <EXPLICIT_TYPE>]. If no ticket was explicitly supplied, ask the user for it; leave capture unassigned. Do not invent source, parent, model or usage fields.", r.Identity.Source, r.Identity.SessionID, command, quote(directory), r.Identity.Source, quote(r.Identity.Store), quote(r.Identity.SessionID), quote(r.Request))
}

func ReceiveCodex(ctx context.Context, s *Store, input HookInput, sourceStore, binary, directory string, wrappers ...string) (HookOutput, error) {
	return receiveCodex(ctx, s, input, sourceStore, binary, directory, false, wrappers...)
}

// ReceiveCodexUser uses the selected native handler for binding feedback without
// depending on a historical per-project wrapper or installation receipt.
func ReceiveCodexUser(ctx context.Context, s *Store, input HookInput, sourceStore, binary, directory string) (HookOutput, error) {
	return receiveCodex(ctx, s, input, sourceStore, binary, directory, true)
}

func receiveCodex(ctx context.Context, s *Store, input HookInput, sourceStore, binary, directory string, userLevel bool, wrappers ...string) (result HookOutput, returnErr error) {
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
	input, normalizeErr = NormalizeCodexSubagentStart(input, sourceStore, s.project)
	if normalizeErr != nil {
		return HookOutput{}, normalizeErr
	}
	// Commit identity and its unverified source pointer before metadata inspection.
	r, err := s.Register(ctx, domain.CaptureIdentity{Source: domain.SourceCodex, Store: sourceStore, SessionID: input.SessionID}, input.TranscriptPath, input.Event)
	if err != nil {
		var accepted *AcceptedRegistrationError
		if errors.As(err, &accepted) {
			return HookOutput{Registered: true}, err
		}
		if input.TranscriptPath != "" {
			err = errors.Join(err, codex.ValidateRolloutStore(sourceStore, input.TranscriptPath))
		}
		return HookOutput{}, fmt.Errorf("durable lifecycle registration failed; verify configured capture state permissions: %w", err)
	}
	defer func() {
		result.Registered = true
		if r.ProjectProof == domain.CaptureProjectUnverified {
			result.Message = "AgenticLens accepted this lifecycle handoff, but exact source project metadata is not yet verified. Run capture diagnostics for this project and retry capture reconcile after source flush; do not infer ticket ownership or complete capture."
		}
	}()
	if input.TranscriptPath != "" {
		if err := codex.ValidateRolloutStore(sourceStore, input.TranscriptPath); err != nil {
			return HookOutput{}, err
		}
	}
	r, err = s.ProveRegistration(ctx, r)
	if err != nil {
		return HookOutput{}, fmt.Errorf("registration accepted but source project proof could not be persisted: %w", err)
	}

	// Subagent callbacks carry the parent's session_id. Never attribute that ID to the child.
	if input.Event == "SubagentStart" {
		if err = s.QueueChildStart(ctx, sourceStore, input.SessionID, input.AgentID); err != nil {
			return HookOutput{}, err
		}
		if input.AgentTranscriptPath != "" {
			if err := s.QueueChild(ctx, sourceStore, input.AgentTranscriptPath, input.SessionID); err != nil {
				return HookOutput{}, err
			}
		}
		return HookOutput{Context: &HookContext{Event: input.Event, Context: "AgenticLens observed your start. The callback identifies your parent, not your own source identity. Do not bind using the parent's ID. Your identity will be reconciled only from authoritative session metadata; if independently launched, respond to your own SessionStart binding request."}}, nil
	}
	if input.Event == "SubagentStop" && input.AgentTranscriptPath != "" {
		// Child registration is deferred to reconciliation of a proven source path, rather than assuming agent_id equals a harness ID.
		if err = s.QueueChild(ctx, sourceStore, input.AgentTranscriptPath, input.SessionID); err != nil {
			return HookOutput{}, err
		}
	}
	if (r.TicketID == "" || r.TicketTitle == "" || (r.InstallationID != "" && r.BindingInstallationID != r.InstallationID)) && (input.Event == "SessionStart" || input.Event == "UserPromptSubmit") {
		prompt := BindingPrompt(r, binary, directory)
		if userLevel {
			// The machine dispatcher already selected this native build. The explicit
			// binding command remains usable even when no PATH launcher was installed.
		} else if len(wrappers) > 0 && wrappers[0] != "" {
			prompt = WrapperBindingPrompt(r, wrappers[0], s.project, directory)
		} else if wrapper, e := installedWrapper(directory, "codex"); e == nil {
			if wrapper == "" {
				prompt = "AgenticLens recorded this session but the installed integration predates the common version-pinned wrapper. Refresh capture setup before binding; inspect capture diagnostics. No ticket has been inferred."
			} else {
				prompt = WrapperBindingPrompt(r, wrapper, s.project, directory)
			}
		} else if !errors.Is(e, os.ErrNotExist) {
			return HookOutput{}, e
		}
		return HookOutput{Context: &HookContext{Event: input.Event, Context: prompt + projectBindingArgument(s)}}, nil
	}
	return HookOutput{}, nil
}

func projectBindingArgument(s *Store) string {
	if s.project == "" {
		return ""
	}
	quote := QuoteArgument
	if runtime.GOOS == "windows" {
		quote = func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	}
	return " Use --repo " + quote(s.project) + " on that binding command to preserve the installed project context."
}

func installedWrapper(directory, source string) (string, error) {
	body, err := os.ReadFile(manifestPath(directory, source))
	if err != nil {
		return "", err
	}
	var value struct {
		Wrapper string `json:"wrapper"`
	}
	err = json.Unmarshal(body, &value)
	return value.Wrapper, err
}

// WrapperBindingPrompt routes a future agent response through the same offline
// project pin. The already-running worker continues to use its native binary.
func WrapperBindingPrompt(r domain.CaptureRegistration, wrapper, project, directory string) string {
	return NativeManagerBindingPrompt(r, wrapper, project, directory, runtime.GOOS == "windows")
}

// NativeManagerBindingPrompt renders the installed native command for the
// selected shell while retaining project selection and literal argument values.
func NativeManagerBindingPrompt(r domain.CaptureRegistration, wrapper, project, directory string, windows bool) string {
	quote := QuoteArgument
	if windows {
		quote = func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	}
	command := quote(wrapper)
	if windows {
		command = "& " + command
	}
	command += " --al-offline --al-project " + quote(project)
	return fmt.Sprintf("AgenticLens requires explicit ticket information for this exact %s session %s. Do not infer a ticket from directory, branch, role or timing. If the user explicitly identified the active ticket, respond by executing: %s capture bind --repo %s --state-dir %s --source %s --store %s --session-id %s --request %s --ticket-id <EXPLICIT_TICKET_ID> --ticket-title <EXPLICIT_TICKET_TITLE> [--ticket-type <EXPLICIT_TYPE>]. Otherwise ask the user for the ticket and leave capture unassigned. Do not invent source, parent, model or usage fields.", r.Identity.Source, r.Identity.SessionID, command, quote(project), quote(directory), r.Identity.Source, quote(r.Identity.Store), quote(r.Identity.SessionID), quote(r.Request))
}
