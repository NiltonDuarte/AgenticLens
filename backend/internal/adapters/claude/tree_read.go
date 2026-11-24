package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const maxMetaBytes = 1024 * 1024

// SubagentMeta holds the identity fields of a native subagent meta file.
type SubagentMeta struct {
	AgentType string
	ToolUseID string
}

type subagentRead struct {
	files      SubagentFiles
	transcript Transcript
	meta       SubagentMeta
	readErr    error
	metaErr    error
}

// Tree is a read but not yet proven root with its stored subagent files.
type Tree struct {
	RootID      string
	SelectedID  string
	Root        Transcript
	subagents   []subagentRead
	orphanMetas []string
}

// ReadTree reads the selected root and every subagent file beside it. An
// unreadable root, cancellation or budget exhaustion fails the read. An
// unreadable subagent is recorded so it can stay unresolved without blocking
// the rest of the tree.
func ReadTree(ctx context.Context, files TreeFiles, budget *DiscoveryBudget) (Tree, error) {
	if err := ctx.Err(); err != nil {
		return Tree{}, err
	}
	if err := budget.validate(); err != nil {
		return Tree{}, err
	}
	root, err := readTranscriptFile(ctx, files.RootPath, TranscriptExpectation{RootSessionID: files.RootID}, budget)
	if err != nil {
		return Tree{}, fmt.Errorf("reading Claude Code root session %s: %w", files.RootID, err)
	}
	tree := Tree{RootID: files.RootID, SelectedID: files.SelectedID, Root: root, orphanMetas: append([]string(nil), files.OrphanMetas...)}
	for _, subagent := range files.Subagents {
		read := subagentRead{files: subagent}
		read.transcript, read.readErr = readTranscriptFile(ctx, subagent.TranscriptPath, TranscriptExpectation{RootSessionID: files.RootID, AgentID: subagent.AgentID}, budget)
		if fatal := fatalReadError(ctx, read.readErr); fatal != nil {
			return Tree{}, fatal
		}
		if subagent.MetaPath == "" {
			read.metaErr = errors.New("meta file missing")
		} else {
			read.meta, read.metaErr = readMetaFile(ctx, subagent.MetaPath, budget)
			if fatal := fatalReadError(ctx, read.metaErr); fatal != nil {
				return Tree{}, fatal
			}
		}
		tree.subagents = append(tree.subagents, read)
	}
	return tree, ctx.Err()
}

func fatalReadError(ctx context.Context, err error) error {
	if cancelled := ctx.Err(); cancelled != nil {
		return cancelled
	}
	if errors.Is(err, ErrDiscoveryBudget) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

func readTranscriptFile(ctx context.Context, path string, expect TranscriptExpectation, budget *DiscoveryBudget) (transcript Transcript, err error) {
	reader, closeReader, err := openSource(ctx, path, budget, true)
	if err != nil {
		return Transcript{}, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	return ParseTranscript(ctx, reader, expect)
}

func readMetaFile(ctx context.Context, path string, budget *DiscoveryBudget) (meta SubagentMeta, err error) {
	reader, closeReader, err := openSource(ctx, path, budget, false)
	if err != nil {
		return SubagentMeta{}, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	data, err := io.ReadAll(io.LimitReader(reader, maxMetaBytes+1))
	if err != nil {
		return SubagentMeta{}, err
	}
	if len(data) > maxMetaBytes {
		return SubagentMeta{}, errors.New("meta file exceeds its structural size limit")
	}
	var raw struct {
		AgentType string `json:"agentType"`
		ToolUseID string `json:"toolUseId"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return SubagentMeta{}, errors.New("meta file is not valid JSON")
	}
	return SubagentMeta{AgentType: raw.AgentType, ToolUseID: raw.ToolUseID}, nil
}
