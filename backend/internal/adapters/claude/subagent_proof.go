package claude

import (
	"fmt"
	"sort"
	"strings"
)

// ProvenChild is a subagent whose parent is proven by native source records.
type ProvenChild struct {
	Transcript Transcript
	ParentID   string
	// Role is the native agentType, recorded verbatim.
	Role string
}

// ProvenTree contains the root and only those subagents whose parentage the
// source proves. Diagnostics name every subagent left unresolved.
type ProvenTree struct {
	RootID      string
	SelectedID  string
	Root        Transcript
	Children    []ProvenChild
	Diagnostics []string
	// childByToolUse maps a parent's Agent tool_use ID to its proven child.
	childByToolUse map[string]string
}

type agentCall struct {
	ownerID  string
	resultID string
}

// Prove links a subagent only when its meta toolUseId matches exactly one
// Agent tool_use in a proven transcript of the same tree, and no parent result
// contradicts its agent ID. Nested children are proven through proven parents.
// An unproven or unreadable subagent never receives an inferred parent.
func (tree Tree) Prove() ProvenTree {
	proven := ProvenTree{RootID: tree.RootID, SelectedID: tree.SelectedID, Root: tree.Root, childByToolUse: make(map[string]string)}
	calls := make(map[string][]agentCall)
	collect := func(transcript Transcript) {
		for _, use := range transcript.toolUses {
			if use.name != "Agent" {
				continue
			}
			calls[use.id] = append(calls[use.id], agentCall{ownerID: transcript.OwnerID, resultID: transcript.results[use.id].agentID})
		}
	}
	collect(tree.Root)
	readable := make(map[string]bool)
	for _, subagent := range tree.subagents {
		if subagent.readErr == nil {
			collect(subagent.transcript)
			readable[subagent.files.AgentID] = true
		}
	}
	accepted := map[string]bool{tree.RootID: true}
	resolved := make(map[string]bool)
	reasons := make(map[string]string)
	for changed := true; changed; {
		changed = false
		for _, subagent := range tree.subagents {
			id := subagent.files.AgentID
			if resolved[id] {
				continue
			}
			parent, reason := subagentParent(subagent, calls)
			if reason != "" {
				reasons[id] = reason
				continue
			}
			if !accepted[parent.ownerID] {
				reasons[id] = fmt.Sprintf("parent %s is not proven", parent.ownerID)
				continue
			}
			resolved[id] = true
			accepted[id] = true
			delete(reasons, id)
			changed = true
			proven.childByToolUse[subagent.meta.ToolUseID] = id
			proven.Children = append(proven.Children, ProvenChild{Transcript: subagent.transcript, ParentID: parent.ownerID, Role: subagent.meta.AgentType})
		}
	}
	sort.Slice(proven.Children, func(i, j int) bool {
		return proven.Children[i].Transcript.OwnerID < proven.Children[j].Transcript.OwnerID
	})
	ids := make([]string, 0, len(reasons))
	for id := range reasons {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		proven.Diagnostics = append(proven.Diagnostics, fmt.Sprintf("Claude Code subagent %s unresolved: %s; stored facts preserved", id, reasons[id]))
	}
	for _, id := range tree.orphanMetas {
		proven.Diagnostics = append(proven.Diagnostics, fmt.Sprintf("Claude Code subagent %s unavailable: meta file has no transcript; stored facts preserved", id))
	}
	// Explicit child IDs reported by parent results without any transcript.
	reported := make(map[string]bool)
	for _, recorded := range calls {
		for _, call := range recorded {
			if call.resultID != "" && !readable[call.resultID] && !reported[call.resultID] && !containsString(tree.orphanMetas, call.resultID) && !hasSubagentFile(tree, call.resultID) {
				reported[call.resultID] = true
			}
		}
	}
	explicit := make([]string, 0, len(reported))
	for id := range reported {
		explicit = append(explicit, id)
	}
	sort.Strings(explicit)
	for _, id := range explicit {
		proven.Diagnostics = append(proven.Diagnostics, fmt.Sprintf("Claude Code subagent %s unavailable: reported by its parent but no transcript exists; stored facts preserved", id))
	}
	return proven
}

func subagentParent(subagent subagentRead, calls map[string][]agentCall) (agentCall, string) {
	id := subagent.files.AgentID
	if subagent.readErr != nil {
		return agentCall{}, "transcript unreadable: " + subagent.readErr.Error()
	}
	if subagent.metaErr != nil {
		return agentCall{}, subagent.metaErr.Error()
	}
	toolUseID := strings.TrimSpace(subagent.meta.ToolUseID)
	if toolUseID == "" {
		return agentCall{}, "meta file has no toolUseId"
	}
	matches := calls[toolUseID]
	if len(matches) == 0 {
		return agentCall{}, "meta toolUseId matches no Agent tool_use in this tree"
	}
	if len(matches) > 1 {
		return agentCall{}, "meta toolUseId matches several Agent tool_use records"
	}
	parent := matches[0]
	if parent.ownerID == id {
		return agentCall{}, "meta toolUseId names the subagent's own call"
	}
	if parent.resultID != "" && parent.resultID != id {
		return agentCall{}, "parent result reports a different agentId"
	}
	return parent, ""
}

func hasSubagentFile(tree Tree, id string) bool {
	for _, subagent := range tree.subagents {
		if subagent.files.AgentID == id {
			return true
		}
	}
	return false
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// IsProven reports whether the ID is the root or a proven child.
func (tree ProvenTree) IsProven(id string) bool {
	if id == tree.RootID {
		return true
	}
	for _, child := range tree.Children {
		if child.Transcript.OwnerID == id {
			return true
		}
	}
	return false
}
