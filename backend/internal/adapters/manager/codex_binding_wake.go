package manager

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

// Match the native engine invocation shown in the binding prompt, including
// absolute paths quoted for the shell. The command is never executed or used
// to resolve the project, executable, session or ticket.
var bindingSubmission = regexp.MustCompile(`^(?:'/[^']*(?:'"'"'[^']*)*'|"/[^"]*"|/[^[:space:];&|<>]+)[[:blank:]]+capture[[:blank:]]+submit-binding[[:blank:]]`)

// The PostToolUse hook runs across all Codex projects. Discard unrelated Bash
// calls before project selection. Forward only the native session context.
func codexBindingWake(input io.Reader, output, diagnostic io.Writer) error {
	body, err := io.ReadAll(io.LimitReader(input, 65537))
	if err != nil {
		return err
	}
	type sessionContext struct {
		Event            string `json:"hook_event_name"`
		SessionID        string `json:"session_id"`
		WorkingDirectory string `json:"cwd"`
		TranscriptPath   string `json:"transcript_path"`
	}
	var call struct {
		sessionContext
		Tool  string `json:"tool_name"`
		Input struct {
			Command string `json:"command"`
		} `json:"tool_input"`
	}
	if len(body) > 65536 || json.Unmarshal(body, &call) != nil || call.Event != "PostToolUse" || call.Tool != "Bash" || !bindingSubmission.MatchString(strings.TrimSpace(call.Input.Command)) {
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	context, err := json.Marshal(call.sessionContext)
	if err != nil {
		return err
	}
	return nativeHook("Codex", []string{"capture", "user-hook"}, bytes.NewReader(context), output, diagnostic)
}
