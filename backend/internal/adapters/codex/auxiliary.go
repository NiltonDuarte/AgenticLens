package codex

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"

	"agenticlens/internal/domain"
)

type auxiliaryCall struct {
	Name      string
	Input     string
	EventID   string
	Timestamp string
	Turn      int64
}

// Only explicit native calls are interpreted; text announcing a skill or child
// metadata alone does not prove an invocation.
func (transcript *Transcript) beginAuxiliary(callID, name, input, eventID, timestamp string) {
	if transcript.auxiliaryCalls == nil {
		transcript.auxiliaryCalls = map[string]auxiliaryCall{}
	}
	if _, exists := transcript.auxiliaryCalls[callID]; exists {
		return
	}
	transcript.auxiliaryCalls[callID] = auxiliaryCall{Name: name, Input: input, EventID: eventID, Timestamp: timestamp, Turn: transcript.Turns}
}

func (transcript *Transcript) finishAuxiliary(callID, output string) {
	call, ok := transcript.auxiliaryCalls[callID]
	if !ok {
		return
	}
	delete(transcript.auxiliaryCalls, callID)
	provenance := domain.FactProvenance{SourceEventID: call.EventID, SourceInvocationID: callID}
	var args struct {
		Name         string `json:"name"`
		SubagentType string `json:"subagent_type"`
		AgentType    string `json:"agent_type"`
		Description  string `json:"description"`
		TaskName     string `json:"task_name"`
	}
	// Native arguments are structured; an exec wrapper is inspected only for
	// positively recognized read commands, never evaluated as JavaScript.
	argsValid := json.Unmarshal([]byte(call.Input), &args) == nil
	name := strings.TrimPrefix(strings.TrimPrefix(call.Name, "functions."), "collaboration.")
	switch name {
	case "spawn_agent", "delegate_task", "mcp__t3_code__delegate_task":
		if !argsValid {
			return
		}
		kind := args.SubagentType
		if kind == "" {
			kind = args.AgentType
		}
		description := args.Description
		if description == "" {
			description = args.TaskName
		}
		link := domain.TaskLinkFact{FactProvenance: provenance, SubagentType: kind, Description: description, ParentTurn: int(call.Turn), Timestamp: call.Timestamp, CostEvidence: domain.UnavailableCost()}
		var result struct {
			SessionID      string `json:"session_id"`
			ChildSessionID string `json:"child_session_id"`
			AgentID        string `json:"agent_id"`
			TaskName       string `json:"task_name"`
			ChildThreadID  string `json:"childThreadId"`
		}
		if json.Unmarshal([]byte(output), &result) == nil {
			link.ChildSession = result.ChildSessionID
			if link.ChildSession == "" {
				link.ChildSession = result.SessionID
			}
			if link.ChildSession == "" {
				link.ChildSession = result.AgentID
			}
			if link.ChildSession == "" {
				link.ChildSession = result.ChildThreadID
			}
			link.ChildReference = result.TaskName
		}
		transcript.TaskLinks = append(transcript.TaskLinks, link)
	case "skill":
		if !argsValid || args.Name == "" {
			return
		}
		transcript.Skills = append(transcript.Skills, domain.SkillFact{FactProvenance: provenance, Name: args.Name, TurnIndex: int(call.Turn), Timestamp: call.Timestamp, CallerRole: string(transcript.role), Evidence: "skill_tool", Outcome: "unavailable"})
	case "exec", "exec_command", "mcp__functions__exec_command":
		if !completedReadOutput(output) {
			return
		}
		for _, skill := range readSkillNames(call.Input, name == "exec") {
			transcript.Skills = append(transcript.Skills, domain.SkillFact{FactProvenance: provenance, Name: skill, TurnIndex: int(call.Turn), Timestamp: call.Timestamp, CallerRole: string(transcript.role), Evidence: "skill_file_read", Outcome: "unavailable"})
		}
	}
}

// completedReadOutput accepts exactly one identifiable command result. A wrapper
// containing multiple results cannot establish which command each status belongs to.
func completedReadOutput(output string) bool {
	succeeded, known := readExitStatus(output)
	return known && succeeded
}
func readExitStatus(output string) (bool, bool) {
	var value struct {
		ExitCode *int `json:"exit_code"`
	}
	if json.Unmarshal([]byte(output), &value) == nil && value.ExitCode != nil {
		return *value.ExitCode == 0, true
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal([]byte(output), &blocks) == nil {
		count, success := 0, false
		for _, block := range blocks {
			if block.Type == "input_text" {
				if passed, known := readExitStatus(block.Text); known {
					count++
					success = passed
				}
			}
		}
		if count > 1 {
			return false, true
		}
		return success, count == 1
	}
	if strings.HasPrefix(output, "Exit code: ") {
		line, _, _ := strings.Cut(output, "\n")
		return line == "Exit code: 0", true
	}
	if strings.HasPrefix(output, "Chunk ID:") || strings.HasPrefix(output, "Wall time:") {
		headers, _, ok := strings.Cut(output, "Final output:")
		if ok && strings.Contains(headers, "\nProcess exited with code ") {
			return strings.Contains(headers, "\nProcess exited with code 0\n"), true
		}
	}
	return false, false
}

// A deliberately bounded wrapper grammar proves one unconditional awaited call.
// Optional catalog printing is pure and cannot execute another shell command.
var singleReadWrapper = regexp.MustCompile(`(?s)^\s*text\(\s*await\s+tools\.exec_command\(\s*(\{.*\})\s*\)\s*\)\s*;?\s*(?:text\(ALL_TOOLS\.filter\(x=>/[a-zA-Z0-9_|]+/\.test\(x\.name\)\)\);?\s*)?$`)
var unquotedCommandKey = regexp.MustCompile(`^\{\s*cmd\s*:`)

func readSkillNames(input string, wrapper bool) []string {
	var command struct {
		Cmd     string `json:"cmd"`
		Command string `json:"command"`
		Shell   string `json:"shell"`
	}
	if wrapper {
		match := singleReadWrapper.FindStringSubmatch(input)
		if len(match) != 2 {
			return nil
		}
		literal := unquotedCommandKey.ReplaceAllString(match[1], `{"cmd":`)
		if json.Unmarshal([]byte(literal), &command) != nil {
			return nil
		}
	} else if json.Unmarshal([]byte(input), &command) != nil {
		return nil
	}
	if command.Shell != "" {
		return nil
	}
	cmd := command.Cmd
	if cmd == "" {
		cmd = command.Command
	}
	commands := []string{cmd}
	var names []string
	for _, cmd := range commands {
		// Shell substitutions, pipelines and redirections make output attribution
		// ambiguous. Support literal cat reads in successful && chains only.
		if strings.ContainsAny(cmd, "`$|;><\n#") {
			continue
		}

		segments := strings.Split(cmd, "&&")
		safe := true
		for _, segment := range segments {
			words := strings.Fields(strings.TrimSpace(segment))
			if len(words) == 0 {
				safe = false
				break
			}
			if words[0] == "cat" {
				continue
			}
			switch strings.Join(words, " ") {
			case "pwd", "git status --short", "git branch --show-current", "git rev-parse HEAD":
			default:
				safe = false
			}
		}
		if !safe {
			continue
		}
		for _, segment := range segments {
			words := strings.Fields(strings.TrimSpace(segment))
			if len(words) < 2 || words[0] != "cat" {
				continue
			}

			// Any option or shell syntax makes this cat invocation unsupported.
			unsupported := false
			for _, word := range words[1:] {
				if strings.HasPrefix(strings.Trim(word, "\"'"), "-") || strings.ContainsAny(word, "&(){}*?!\\") {
					unsupported = true
				}
			}
			if unsupported {
				continue
			}
			for _, word := range words[1:] {
				path := strings.Trim(word, "\"'")
				if strings.HasPrefix(path, "-") || filepath.Base(path) != "SKILL.md" || !strings.Contains(filepath.ToSlash(path), "/skills/") {
					continue
				}
				names = append(names, filepath.Base(filepath.Dir(path)))
			}
		}
	}
	return names
}
