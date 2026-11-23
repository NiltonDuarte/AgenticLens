package main

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"agenticlens/internal/adapters"
	adaptersconfig "agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/git"
)

const (
	stagedCaptureCheckTimeout = time.Second
	captureAdvisoryPrefix     = "AgenticLens capture advisory:"
	captureHealthUnavailable  = "capture health unavailable; run capture diagnostics to inspect configuration/state and refresh setup if needed."
)

// checkStagedCapture returns an advisory and whether it found staged snapshots.
// Only a confirmed warning takes priority over capture-health inspection.
func checkStagedCapture(ctx context.Context, project, stateDir string) (string, bool) {
	insideGit, err := git.HasWorktreeMarker(project)
	if err != nil {
		return "Staged pending snapshot inspection unavailable. Review Git worktree metadata and run capture check again.", false
	}
	if !insideGit {
		return "", false
	}
	directories, err := adaptersconfig.ResolveStagedPendingDirectories(project)
	if err != nil {
		return "Staged pending snapshot directory unavailable. Set a valid shared agenticlens.pending_dir and run capture check again.", false
	}
	ctx, cancel := context.WithTimeout(ctx, stagedCaptureCheckTimeout)
	defer cancel()
	staged := adapters.InspectStagedPendingSnapshots(ctx, project, directories, stateDir)
	var messages []string
	if staged.AwaitingTicket > 0 {
		messages = append(messages, fmt.Sprintf("%d awaiting a ticket (bind the capture).", staged.AwaitingTicket))
	}
	if staged.AwaitingDestination > 0 {
		messages = append(messages, fmt.Sprintf("%d awaiting a publication destination (run capture diagnostics).", staged.AwaitingDestination))
	}
	warned := len(messages) > 0
	if warned {
		messages = append([]string{"staged pending snapshots."}, messages...)
		messages = append(messages, "Before committing, review or unstage the snapshots. Inspect them with git diff --cached.")
	}
	if staged.Incomplete {
		messages = append(messages, "Staged pending snapshot inspection incomplete. Review git diff --cached and run capture check again.")
	}
	return strings.Join(messages, " "), warned
}

// An installed hook has 512 bytes for the prefix, one line and its newline.
// Preserve the leading health finding and reserve room for the staged diagnostic.
func combineCaptureAdvisory(health, staged string) string {
	health = strings.TrimSpace(strings.TrimPrefix(health, captureAdvisoryPrefix))
	health = strings.Join(strings.Fields(health), " ")
	budget := max(0, 512-len(captureAdvisoryPrefix)-len(staged)-3)
	if len(health) > budget {
		for budget > 0 && !utf8.RuneStart(health[budget]) {
			budget--
		}
		health = strings.TrimSpace(health[:budget])
		if lastSpace := strings.LastIndexByte(health, ' '); lastSpace > 0 {
			health = health[:lastSpace]
		}
	}
	if health == "" {
		return staged
	}
	return health + " " + staged
}
