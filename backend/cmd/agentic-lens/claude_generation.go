package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/claude"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// claudeGenerationInput carries explicit selections for one manual Claude Code
// extraction. Manual generation consumes no capture budget.
type claudeGenerationInput struct {
	project        string
	claudeHomeFlag string
	configuredHome string
	nonce          string
	sessionID      string
	cutoff         time.Time
	policy         domain.GenerationPolicy
	recorded       []domain.SessionFact
	baseline       []domain.SessionFact
	existingUsers  map[string]string
	gitUser        string
}

type claudeGeneration struct {
	sessions  []domain.SessionFact
	source    domain.SourceEvents
	rootID    string
	startedAt string
	owners    int
}

// extractClaudeGeneration resolves the selected store, discovers and reads the
// selected tree and returns complete owners only. Every failure happens before
// publication, so stored ledgers and summaries stay unchanged. Diagnostics for
// unresolved subagents are written to diagnostics without blocking the root.
func extractClaudeGeneration(input claudeGenerationInput, diagnostics io.Writer) (claudeGeneration, error) {
	home, err := adaptersconfig.ResolveClaudeHome(input.claudeHomeFlag, input.configuredHome)
	if err != nil {
		return claudeGeneration{}, err
	}
	info, err := os.Stat(home)
	if err != nil {
		return claudeGeneration{}, fmt.Errorf("accessing selected Claude Code store %s: %w. Check telemetry.sources.claude.home or --claude-home", home, err)
	}
	if !info.IsDir() {
		return claudeGeneration{}, fmt.Errorf("selected Claude Code store %s must be a directory. Check telemetry.sources.claude.home or --claude-home", home)
	}
	files, err := claude.DiscoverTree(claude.DiscoveryOptions{ClaudeHome: home, Nonce: input.nonce, SessionID: input.sessionID})
	if err != nil {
		return claudeGeneration{}, fmt.Errorf("discovering selected Claude Code source at %s: %w. Check telemetry.sources.claude.home or --claude-home", home, err)
	}
	tree, err := claude.ReadTree(context.Background(), files, nil)
	if err != nil {
		return claudeGeneration{}, fmt.Errorf("reading selected Claude Code source at %s: %w", home, err)
	}
	proven := tree.Prove()
	if !proven.IsProven(files.SelectedID) {
		return claudeGeneration{}, fmt.Errorf("Claude Code agent %s is not a proven subagent of root session %s", files.SelectedID, files.RootID)
	}
	extraction, err := proven.Facts(input.cutoff, &input.policy)
	if err != nil {
		return claudeGeneration{}, err
	}
	for _, diagnostic := range extraction.Diagnostics {
		fmt.Fprintln(diagnostics, diagnostic)
	}
	if len(extraction.RootIncomplete) > 0 {
		return claudeGeneration{}, fmt.Errorf("incomplete Claude Code root session %s; stored snapshots preserved: %s", files.RootID, strings.Join(extraction.RootIncomplete, "; "))
	}
	if err := capture.ValidateSessionMerge(input.recorded, extraction.Sessions); err != nil {
		return claudeGeneration{}, err
	}
	if err := validateClaudeOwnership(input.project, files.RootID, extraction.WorkingDirectory, input.baseline); err != nil {
		return claudeGeneration{}, fmt.Errorf("Source ownership conflict: %w", err)
	}
	resolveCodexSessionUsers(extraction.Sessions, input.existingUsers, input.gitUser)
	return claudeGeneration{
		sessions:  extraction.Sessions,
		source:    domain.SourceEvents{Events: extraction.Events, CheckpointEvidence: extraction.CheckpointEvidence, CompleteSessionIDs: extraction.CompleteSessionIDs},
		rootID:    files.RootID,
		startedAt: extraction.StartedAt,
		owners:    len(extraction.Sessions),
	}, nil
}

// validateClaudeOwnership accepts a root already recorded for this ticket as a
// Claude Code owner. A new root must have recorded its working directory inside
// this project. Proven children inherit the root's check, because isolated
// subagent worktrees can sit inside nested checkouts. No capture registration
// is written.
func validateClaudeOwnership(project, rootID, workingDirectory string, baseline []domain.SessionFact) error {
	for _, owner := range baseline {
		if owner.ID == rootID && owner.Source == domain.SourceClaude {
			return nil
		}
	}
	if err := capture.ValidateProject(project, workingDirectory); err != nil {
		return fmt.Errorf("selected source owner %s: %w", rootID, err)
	}
	return nil
}

func errClaudeHomeBlank(provided map[string]bool, value string) error {
	if provided["claude-home"] && strings.TrimSpace(value) == "" {
		return errors.New("--claude-home must be nonblank when supplied")
	}
	return nil
}
