package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters/filesystem"
)

const maxDiscoveryLineBytes = 8 * 1024 * 1024

// ErrOwnerUnavailable distinguishes a missing exact owner from ambiguous or
// failed source discovery, whose diagnostics must remain visible to callers.
var ErrOwnerUnavailable = errors.New("Codex native session unavailable")

// DiscoveryOptions selects a Codex thread from local rollout files using an
// explicit nonce or thread ID. Ticket text and working-directory heuristics
// are intentionally excluded because they can identify the wrong session.
type DiscoveryOptions struct {
	CodexHome string
	Nonce     string
	SessionID string
	Budget    *DiscoveryBudget
}

type rolloutCandidate struct {
	path          string
	threadID      string
	parentID      string
	nonceMatch    bool
	inspectionErr error
}

// DiscoverRolloutPaths finds the selected root thread and its known child
// threads under the explicitly selected home's sessions and archived_sessions. It reads
// rollout records only to inspect session metadata and compare the requested
// nonce/ticket text; it does not retain or return transcript content.
func DiscoverRolloutPaths(options DiscoveryOptions) ([]string, error) {
	ctx := options.Budget.readContext()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	home := options.CodexHome
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply an explicit source store")
	}
	if err := options.Budget.validate(); err != nil {
		return nil, err
	}
	paths, err := listRolloutPaths(ctx, home, options.Budget)
	if err != nil {
		return nil, err
	}
	candidates := make([]rolloutCandidate, 0, len(paths))
	for _, path := range paths {
		candidate, inspectErr := inspectRollout(ctx, path, options.Nonce, options.Budget)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if inspectErr != nil {
			if errors.Is(inspectErr, ErrDiscoveryBudget) {
				return nil, inspectErr
			}
			// Keep native identity/ancestry already observed before a read failure.
			// Only an incomplete member of the selected family blocks extraction;
			// unrelated or entirely unattributable rollouts remain tolerable.
			if candidate.threadID == "" {
				continue
			}
			candidate.inspectionErr = inspectErr
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no readable Codex rollout JSONL files found under %s", home)
	}

	// Check every relationship asserted by a copy before reducing ancestry to
	// one parent per owner. Otherwise a conflicting copy could route a selected
	// child into a different family and escape the ambiguity check.
	if err := rejectAmbiguousFamily(ctx, candidates, options); err != nil {
		return nil, err
	}

	selectedID := strings.TrimSpace(options.SessionID)
	switch {
	case strings.TrimSpace(options.Nonce) != "":
		matches := matchingRoots(candidates, func(candidate rolloutCandidate) bool { return candidate.nonceMatch })
		if len(matches) == 0 {
			return nil, fmt.Errorf("no Codex rollout contains the supplied nonce")
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("nonce matches multiple Codex root sessions (%s); use --session-id", strings.Join(matches, ", "))
		}
		selectedID = matches[0]
		if explicitID := strings.TrimSpace(options.SessionID); explicitID != "" {
			if !containsThread(candidates, explicitID) {
				return nil, fmt.Errorf("Codex session %q was not found under %s", explicitID, home)
			}
			explicitRoot := rootThreadID(candidates, explicitID)
			if explicitRoot != selectedID {
				return nil, fmt.Errorf("supplied nonce and session ID select different Codex root sessions")
			}
		}
	case selectedID != "":
		found := false
		for _, candidate := range candidates {
			if candidate.threadID == selectedID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Codex session %q was not found under %s", selectedID, home)
		}
		selectedID = rootThreadID(candidates, selectedID)
	default:
		return nil, errors.New("Codex session discovery requires --nonce or --session-id")
	}

	selected := make([]rolloutCandidate, 0)
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if candidate.threadID == selectedID || isDescendant(candidates, candidate.threadID, selectedID) {
			if candidate.inspectionErr != nil {
				return nil, fmt.Errorf("incomplete Codex discovery for selected owner %s at %s: %w", candidate.threadID, candidate.path, candidate.inspectionErr)
			}
			selected = append(selected, candidate)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("Codex root session %q has no readable rollout", selectedID)
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].threadID == selectedID {
			return true
		}
		if selected[j].threadID == selectedID {
			return false
		}
		return selected[i].threadID < selected[j].threadID
	})
	result := make([]string, 0, len(selected))
	for _, candidate := range selected {
		result = append(result, candidate.path)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// rejectAmbiguousFamily follows all observed relationships, without selecting
// an authoritative copy. Ambiguity in an unrelated family does not block reads.
func rejectAmbiguousFamily(ctx context.Context, candidates []rolloutCandidate, options DiscoveryOptions) error {
	members := make(map[string]bool)
	counts := make(map[string]int)
	relations := make(map[string][]string)
	pending := make([]string, 0)
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		counts[candidate.threadID]++
		if candidate.parentID != "" {
			relations[candidate.threadID] = append(relations[candidate.threadID], candidate.parentID)
			relations[candidate.parentID] = append(relations[candidate.parentID], candidate.threadID)
		}
		if candidate.threadID == strings.TrimSpace(options.SessionID) ||
			(strings.TrimSpace(options.Nonce) != "" && candidate.nonceMatch) {
			pending = append(pending, candidate.threadID)
		}
	}
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if members[id] {
			continue
		}
		members[id] = true
		pending = append(pending, relations[id]...)
	}
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		if members[candidate.threadID] && counts[candidate.threadID] > 1 {
			return fmt.Errorf("incomplete Codex discovery: owner %s has ambiguous rollout paths", candidate.threadID)
		}
	}
	return nil
}

func containsThread(candidates []rolloutCandidate, threadID string) bool {
	for _, candidate := range candidates {
		if candidate.threadID == threadID {
			return true
		}
	}
	return false
}

// ParseRolloutPath opens and parses an uncompressed or gzip-compressed Codex rollout.
func ParseRolloutPath(path string) (Transcript, error) {
	return ParseRolloutPathContext(context.Background(), path)
}

func ParseRolloutPathAt(path string, cutoff time.Time) (Transcript, error) {
	return ParseRolloutPathAtContext(context.Background(), path, cutoff)
}

// ParseRolloutPathContext reads and parses one rollout with the caller's context.
// On cancellation it may return owner metadata and observed events for a
// partial trace. A result accompanied by an error is never a snapshot.
func ParseRolloutPathContext(ctx context.Context, path string) (Transcript, error) {
	return ParseRolloutPathAtContext(ctx, path, time.Now().UTC())
}

// ParseRolloutPathAtContext parses a rollout through the selected activity cutoff.
func ParseRolloutPathAtContext(ctx context.Context, path string, cutoff time.Time) (Transcript, error) {
	transcript, _, err := parseRolloutPathContext(ctx, path, cutoff, nil)
	return transcript, err
}

// ParseRolloutPathWithinBudgetContext bounds decompressed bytes and reports the
// amount consumed even when the read is cancelled or otherwise incomplete.
func ParseRolloutPathWithinBudgetContext(ctx context.Context, path string, cutoff time.Time, maxBytes int64) (Transcript, int64, error) {
	if err := ctx.Err(); err != nil {
		return Transcript{}, 0, err
	}
	if maxBytes <= 0 {
		return Transcript{}, 0, ErrDiscoveryBudget
	}
	return parseRolloutPathContext(ctx, path, cutoff, &DiscoveryBudget{Context: ctx, MaxSourceBytes: maxBytes})
}

func parseRolloutPathContext(ctx context.Context, path string, cutoff time.Time, budget *DiscoveryBudget) (transcript Transcript, consumed int64, err error) {
	reader, size, closeReader, err := openRolloutReader(ctx, path)
	if err != nil {
		return Transcript{}, 0, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	var limited *discoveryBudgetReader
	if budget != nil {
		limited = &discoveryBudgetReader{reader: reader, budget: budget, remainingSize: size}
		reader = limited
	}
	transcript, err = ParseRolloutAtContext(ctx, reader, cutoff)
	if budget != nil {
		consumed = budget.SourceBytesRead
		if consumed == budget.MaxSourceBytes && !limited.eof && limited.remainingSize != 0 {
			err = ErrDiscoveryBudget
		}
	}
	if cancelled := ctx.Err(); cancelled != nil {
		err = cancelled
	}
	if err != nil {
		var prefix Transcript
		if budget == nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			prefix = transcript.ownedTracePrefix()
		}
		return prefix, consumed, fmt.Errorf("parsing Codex rollout %s: %w", path, err)
	}
	return transcript, consumed, nil
}

func (transcript Transcript) ownedTracePrefix() Transcript {
	ownerID := transcript.sessionIDOrThread()
	if ownerID == "" {
		return Transcript{}
	}
	prefix := Transcript{
		ThreadID: transcript.ThreadID, SessionID: transcript.SessionID,
		ParentThread: transcript.ParentThread, AgentPath: transcript.AgentPath,
		Warnings: []string{"interrupted Codex rollout"},
	}
	calls := make(map[string]struct{})
	for _, event := range transcript.Events {
		if event.SessionID == ownerID {
			prefix.Events = append(prefix.Events, event)
			if event.CallID != "" {
				calls[event.CallID] = struct{}{}
			}
		}
	}
	for _, delegation := range transcript.Delegations {
		if _, observed := calls[delegation.Provenance.CallID]; observed {
			prefix.Delegations = append(prefix.Delegations, delegation)
		}
	}
	return prefix
}

func listRolloutPaths(ctx context.Context, home string, budget *DiscoveryBudget) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply --codex-home or telemetry.sources.codex.home in the selected local YAML")
	}
	if err := budget.validate(); err != nil {
		return nil, err
	}
	var err error
	home, err = physicalRolloutPath(home)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, directory := range []string{"sessions", "archived_sessions"} {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		root := filepath.Join(home, directory)
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("accessing Codex %s directory: %w", directory, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if strings.HasSuffix(entry.Name(), ".jsonl") || strings.HasSuffix(entry.Name(), ".jsonl.gz") {
				if err := budget.claim(path); err != nil {
					return err
				}
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scanning Codex %s directory: %w", directory, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no rollout files found under %s/sessions or %s/archived_sessions", home, home)
	}
	sort.Strings(paths)
	return paths, nil
}

func inspectRollout(ctx context.Context, path, nonce string, budget *DiscoveryBudget) (candidate rolloutCandidate, err error) {
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if err := budget.claim(path); err != nil {
		return candidate, err
	}
	reader, sourceSize, closeReader, err := openRolloutReader(ctx, path)
	if err != nil {
		return rolloutCandidate{}, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	if budget != nil {
		reader = &discoveryBudgetReader{reader: reader, budget: budget, remainingSize: sourceSize}
	}
	candidate = rolloutCandidate{path: path}
	scanner := bufio.NewScanner(rolloutContextReader{ctx: ctx, reader: reader})
	scanner.Buffer(make([]byte, 64*1024), maxDiscoveryLineBytes)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return candidate, err
		}
		line := scanner.Text()
		if nonce != "" && strings.Contains(line, nonce) {
			candidate.nonceMatch = true
		}
		if candidate.threadID != "" {
			continue
		}
		var record rolloutRecord
		if json.Unmarshal([]byte(line), &record) != nil || record.Type != "session_meta" {
			continue
		}
		metadata, metadataErr := decodeSessionMetadata(record.Payload)
		if metadataErr != nil {
			continue
		}
		candidate.threadID = strings.TrimSpace(metadata.ID)
		candidate.parentID = strings.TrimSpace(metadata.ParentID)
		if nonce == "" {
			break
		}
	}
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if err := scanner.Err(); err != nil {
		return candidate, err
	}
	if candidate.threadID == "" {
		return rolloutCandidate{}, errors.New("rollout has no readable session metadata")
	}
	return candidate, nil
}

func matchingRoots(candidates []rolloutCandidate, match func(rolloutCandidate) bool) []string {
	roots := make(map[string]struct{})
	for _, candidate := range candidates {
		if !match(candidate) {
			continue
		}
		roots[rootThreadID(candidates, candidate.threadID)] = struct{}{}
	}
	result := make([]string, 0, len(roots))
	for rootID := range roots {
		result = append(result, rootID)
	}
	sort.Strings(result)
	return result
}

func rootThreadID(candidates []rolloutCandidate, threadID string) string {
	parents := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		parents[candidate.threadID] = candidate.parentID
	}
	seen := make(map[string]struct{})
	for {
		if _, exists := seen[threadID]; exists {
			return threadID
		}
		seen[threadID] = struct{}{}
		parentID := parents[threadID]
		if parentID == "" || parents[parentID] == "" {
			return parentIDOrSelf(parentID, threadID, parents)
		}
		threadID = parentID
	}
}

func parentIDOrSelf(parentID, threadID string, parents map[string]string) string {
	if parentID != "" {
		if _, exists := parents[parentID]; exists {
			return parentID
		}
	}
	return threadID
}

func isDescendant(candidates []rolloutCandidate, threadID, rootID string) bool {
	seen := make(map[string]struct{})
	current := threadID
	parents := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		parents[candidate.threadID] = candidate.parentID
	}
	for current != "" {
		if current == rootID {
			return true
		}
		if _, exists := seen[current]; exists {
			return false
		}
		seen[current] = struct{}{}
		current = parents[current]
	}
	return false
}

// DiscoverExactRolloutPath selects one owner rather than its complete native
// family. Family discovery remains separate so children validate their own cwd.
func DiscoverExactRolloutPath(home, sessionID string) (string, error) {
	return DiscoverExactRolloutPathContext(context.Background(), home, sessionID)
}

func DiscoverExactRolloutPathWithBudget(home, sessionID string, budget *DiscoveryBudget) (string, error) {
	return discoverExactRolloutPath(budget.readContext(), home, sessionID, budget)
}

// DiscoverExactRolloutPathContext resolves a single unambiguous native owner.
func DiscoverExactRolloutPathContext(ctx context.Context, home, sessionID string) (string, error) {
	return discoverExactRolloutPath(ctx, home, sessionID, nil)
}

func discoverExactRolloutPath(ctx context.Context, home, sessionID string, budget *DiscoveryBudget) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("exact Codex session ID is required")
	}
	paths, err := listRolloutPaths(ctx, home, budget)
	if err != nil {
		return "", err
	}
	selected := ""
	for _, path := range paths {
		candidate, e := inspectRollout(ctx, path, "", budget)
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if e != nil {
			if errors.Is(e, ErrDiscoveryBudget) {
				return "", e
			}
			if candidate.threadID == sessionID {
				return "", fmt.Errorf("incomplete Codex discovery for exact owner %s at %s: %w", sessionID, path, e)
			}
			continue
		}
		if candidate.threadID != sessionID {
			continue
		}
		if selected != "" {
			return "", fmt.Errorf("exact Codex owner %s has ambiguous rollout paths", sessionID)
		}
		selected = path
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if selected == "" {
		return "", fmt.Errorf("exact Codex owner %s unavailable: %w", sessionID, ErrOwnerUnavailable)
	}
	return selected, nil
}

// ValidateRolloutStore proves physical membership in the selected live or archive
// store without rescanning every transcript for each callback. Session identity
// is validated separately from the authoritative metadata by the caller.
func ValidateRolloutStore(home, path string) error {
	var err error
	home, err = physicalRolloutPath(home)
	if err != nil {
		return err
	}

	actual, err := physicalRolloutPath(path)
	if err != nil {
		return err
	}
	for _, name := range []string{"sessions", "archived_sessions"} {
		root, err := physicalRolloutPath(filepath.Join(home, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, actual)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return nil
		}
	}
	return fmt.Errorf("Codex transcript does not belong to selected source store %s", home)
}

func physicalRolloutPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("explicit Codex source path must be nonblank")
	}
	return filesystem.PhysicalPath(path)
}
